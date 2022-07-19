//
// Copyright 2021-2022 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitops

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"

	gitopsv1alpha1 "github.com/redhat-developer/gitops-generator/api/v1alpha1"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
)

const defaultRepoDescription = "Bootstrapped GitOps Repository based on Components"

type Executor interface {
	Execute(baseDir, command string, args ...string) ([]byte, error)
	GenerateParentKustomize(fs afero.Afero, gitOpsFolder string, commonStoragePVC *corev1.PersistentVolumeClaim) error
}

// GenerateCloneAndPush takes in the following args and generates the gitops resources for a given component
// 1. outputPath: Where to output the gitops resources to
// 2. remote: A string of the form https://$token@github.com/<org>/<repo>. Corresponds to the component's gitops repository
// 2. component: A component struct corresponding to a single Component in an Application in AS
// 4. The executor to use to execute the git commands (either gitops.executor or gitops.mockExecutor)
// 5. The filesystem object used to create (either ioutils.NewFilesystem() or ioutils.NewMemoryFilesystem())
// 6. The branch to push to
// 7. The path within the repository to generate the resources in
// 8. The gitops config containing the build bundle;
// Adapted from https://github.com/redhat-developer/kam/blob/master/pkg/pipelines/utils.go#L79
func GenerateCloneAndPush(outputPath string, remote string, component gitopsv1alpha1.Component, e Executor, appFs afero.Afero, branch string, context string, commonStorage *corev1.PersistentVolumeClaim) error {
	componentName := component.Name
	if out, err := e.Execute(outputPath, "git", "clone", remote, componentName); err != nil {
		return fmt.Errorf("failed to clone git repository in %q %q: %s", outputPath, string(out), err)
	}

	repoPath := filepath.Join(outputPath, componentName)

	// Checkout the specified branch
	if _, err := e.Execute(repoPath, "git", "switch", branch); err != nil {
		if out, err := e.Execute(repoPath, "git", "checkout", "-b", branch); err != nil {
			return fmt.Errorf("failed to checkout branch %q in %q %q: %s", branch, repoPath, string(out), err)
		}
	}

	if out, err := e.Execute(repoPath, "rm", "-rf", filepath.Join("components", componentName)); err != nil {
		return fmt.Errorf("failed to delete %q folder in repository in %q %q: %s", filepath.Join("components", componentName), repoPath, string(out), err)
	}

	// Generate the gitops resources and update the parent kustomize yaml file
	gitopsFolder := filepath.Join(repoPath, context)
	componentPath := filepath.Join(gitopsFolder, "components", componentName, "base")
	if err := Generate(appFs, gitopsFolder, componentPath, component, commonStorage); err != nil {
		return fmt.Errorf("failed to generate the gitops resources in %q for component %q: %s", componentPath, componentName, err)
	}

	if out, err := e.Execute(repoPath, "git", "add", "."); err != nil {
		return fmt.Errorf("failed to add files for component %q to repository in %q %q: %s", componentName, repoPath, string(out), err)
	}

	// See if any files changed, and if so, commit and push them up to the repository
	if out, err := e.Execute(repoPath, "git", "--no-pager", "diff", "--cached"); err != nil {
		return fmt.Errorf("failed to check git diff in repository %q %q: %s", repoPath, string(out), err)
	} else if string(out) != "" {
		// Commit the changes and push
		if out, err := e.Execute(repoPath, "git", "commit", "-m", "Generate GitOps resources"); err != nil {
			return fmt.Errorf("failed to commit files to repository in %q %q: %s", repoPath, string(out), err)
		}
		if out, err := e.Execute(repoPath, "git", "push", "origin", branch); err != nil {
			return fmt.Errorf("failed push remote to repository %q %q: %s", remote, string(out), err)
		}
	}

	return nil
}

func GenerateAndPush(outputPath string, remote string, component gitopsv1alpha1.Component, e Executor, appFs afero.Afero, branch string, doPush bool, createdBy string, commonStorage *corev1.PersistentVolumeClaim) error {
	CreatedBy = createdBy

	componentName := component.Spec.ComponentName
	repoPath := filepath.Join(outputPath, component.Spec.Application)

	// Generate the gitops resources and update the parent kustomize yaml file
	gitopsFolder := repoPath

	gitHostAccessToken := component.Spec.Secret
	gitOpsRepoURL := component.Spec.Source.GitSource.URL

	componentPath := filepath.Join(gitopsFolder, "components", componentName, "base")
	if err := Generate(appFs, gitopsFolder, componentPath, component, commonStorage); err != nil {
		return fmt.Errorf("failed to generate the gitops resources in %q for component %q: %s", componentPath, componentName, err)
	}

	// Commit the changes and push
	if doPush {
		u, err := url.Parse(gitOpsRepoURL)
		if err != nil {
			return fmt.Errorf("failed to parse GitOps repo URL %q: %w", gitOpsRepoURL, err)
		}
		parts := strings.Split(u.Path, "/")
		org := parts[1]
		repoName := strings.TrimSuffix(strings.Join(parts[2:], "/"), ".git")
		u.User = url.UserPassword("", gitHostAccessToken)

		client, err := factory.FromRepoURL(u.String())
		if err != nil {
			return fmt.Errorf("failed to create a client to access %q: %w", gitOpsRepoURL, err)
		}
		ctx := context.Background()
		// If we're creating the repository in a personal user's account, it's a
		// different API call that's made, clearing the org triggers go-scm to use
		// the "create repo in personal account" endpoint.
		currentUser, _, err := client.Users.Find(ctx)
		if err != nil {
			return fmt.Errorf("failed to get the user with their auth token: %w", err)
		}
		if currentUser.Login == org {
			org = ""
		}

		ri := &scm.RepositoryInput{
			Private:     true,
			Description: defaultRepoDescription,
			Namespace:   org,
			Name:        repoName,
		}
		_, _, err = client.Repositories.Create(context.Background(), ri)
		if err != nil {
			repo := fmt.Sprintf("%s/%s", org, repoName)
			if org == "" {
				repo = fmt.Sprintf("%s/%s", currentUser.Login, repoName)
			}
			if _, resp, err := client.Repositories.Find(context.Background(), repo); err == nil && resp.Status == 200 {
				return fmt.Errorf("failed to create repository, repo already exists")
			}
			return fmt.Errorf("failed to create repository %q in namespace %q: %w", repoName, org, err)
		}

		if out, err := e.Execute(repoPath, "git", "init", "."); err != nil {
			return fmt.Errorf("failed to initialize git repository in %q %q: %s", repoPath, string(out), err)
		}

		if out, err := e.Execute(repoPath, "git", "add", "."); err != nil {
			return fmt.Errorf("failed to add components to repository in %q %q: %s", repoPath, string(out), err)
		}
		if out, err := e.Execute(repoPath, "git", "commit", "-m", "Generate GitOps resources"); err != nil {
			return fmt.Errorf("failed to commit files to repository in %q %q: %s", repoPath, string(out), err)
		}
		if out, err := e.Execute(repoPath, "git", "branch", "-m", branch); err != nil {
			return fmt.Errorf("failed to switch to branch %q in repository in %q %q: %s", branch, repoPath, string(out), err)
		}
		if out, err := e.Execute(repoPath, "git", "remote", "add", "origin", remote); err != nil {
			return fmt.Errorf("failed to add files for component %q, to remote 'origin' %q to repository in %q %q: %s", componentName, remote, repoPath, string(out), err)
		}
		if out, err := e.Execute(repoPath, "git", "push", "-u", "origin", branch); err != nil {
			return fmt.Errorf("failed push remote to repository %q %q: %s", remote, string(out), err)
		}
	}

	return nil
}

// RemoveAndPush takes in the following args and updates the gitops resources by removing the given component
// 1. outputPath: Where to output the gitops resources to
// 2. remote: A string of the form https://$token@github.com/<org>/<repo>. Corresponds to the component's gitops repository
// 2. component: The component name corresponding to a single Component in an Application in AS. eg. component.Name
// 4. The executor to use to execute the git commands (either gitops.executor or gitops.mockExecutor)
// 5. The filesystem object used to create (either ioutils.NewFilesystem() or ioutils.NewMemoryFilesystem())
// 6. The branch to push to
// 7. The path within the repository to generate the resources in
func RemoveAndPush(outputPath string, remote string, componentName string, e Executor, appFs afero.Afero, branch string, context string) error {
	if out, err := e.Execute(outputPath, "git", "clone", remote, componentName); err != nil {
		return fmt.Errorf("failed to clone git repository in %q %q: %s", outputPath, string(out), err)
	}

	repoPath := filepath.Join(outputPath, componentName)

	// Checkout the specified branch
	if _, err := e.Execute(repoPath, "git", "switch", branch); err != nil {
		if out, err := e.Execute(repoPath, "git", "checkout", "-b", branch); err != nil {
			return fmt.Errorf("failed to checkout branch %q in %q %q: %s", branch, repoPath, string(out), err)
		}
	}

	// Generate the gitops resources and update the parent kustomize yaml file
	gitopsFolder := filepath.Join(repoPath, context)
	componentPath := filepath.Join(gitopsFolder, "components", componentName)
	if out, err := e.Execute(repoPath, "rm", "-rf", componentPath); err != nil {
		return fmt.Errorf("failed to delete %q folder in repository in %q %q: %s", componentPath, repoPath, string(out), err)
	}
	if err := e.GenerateParentKustomize(appFs, gitopsFolder, nil); err != nil {
		return fmt.Errorf("failed to re-generate the gitops resources in %q for component %q: %s", componentPath, componentName, err)
	}

	if out, err := e.Execute(repoPath, "git", "add", "."); err != nil {
		return fmt.Errorf("failed to add files for component %q to repository in %q %q: %s", componentName, repoPath, string(out), err)
	}

	// See if any files changed, and if so, commit and push them up to the repository
	if out, err := e.Execute(repoPath, "git", "--no-pager", "diff", "--cached"); err != nil {
		return fmt.Errorf("failed to check git diff in repository %q %q: %s", repoPath, string(out), err)
	} else if string(out) != "" {
		// Commit the changes and push
		if out, err := e.Execute(repoPath, "git", "commit", "-m", fmt.Sprintf("Removed component %s", componentName)); err != nil {
			return fmt.Errorf("failed to commit files to repository in %q %q: %s", repoPath, string(out), err)
		}
		if out, err := e.Execute(repoPath, "git", "push", "origin", branch); err != nil {
			return fmt.Errorf("failed push remote to repository %q %q: %s", remote, string(out), err)
		}
	}

	return nil
}

// NewCmdExecutor creates and returns an executor implementation that uses
// exec.Command to execute the commands.
func NewCmdExecutor() CmdExecutor {
	return CmdExecutor{}
}

type CmdExecutor struct {
}

func (e CmdExecutor) Execute(baseDir, command string, args ...string) ([]byte, error) {
	c := exec.Command(command, args...)
	c.Dir = baseDir
	output, err := c.CombinedOutput()
	return output, err
}

func (e CmdExecutor) GenerateParentKustomize(fs afero.Afero, gitOpsFolder string, commonStoragePVC *corev1.PersistentVolumeClaim) error {
	return GenerateParentKustomize(fs, gitOpsFolder, commonStoragePVC)
}
