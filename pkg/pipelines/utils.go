package pipelines

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
)

const defaultRepoDescription = "Bootstrapped GitOps Repository"

type clientFactory = func(string) (*scm.Client, error)

type executor interface {
	execute(baseDir, command string, args ...string) ([]byte, error)
}

// BootstrapRepository creates a new empty Git repository in the upstream git
// hosting service from the GitOpsRepoURL.
func BootstrapRepository(o *BootstrapOptions, f clientFactory, e executor) error {
	if o.GitHostAccessToken == "" {
		return nil
	}

	u, err := url.Parse(o.GitOpsRepoURL)
	if err != nil {
		return fmt.Errorf("failed to parse GitOps repo URL %q: %w", o.GitOpsRepoURL, err)
	}
	parts := strings.Split(u.Path, "/")
	org := parts[1]
	repoName := strings.TrimSuffix(strings.Join(parts[2:], "/"), ".git")
	u.User = url.UserPassword("", o.GitHostAccessToken)

	client, err := f(u.String())
	if err != nil {
		return fmt.Errorf("failed to create a client to access %q: %w", o.GitOpsRepoURL, err)
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
	created, _, err := client.Repositories.Create(context.Background(), ri)
	if err != nil {
		return fmt.Errorf("failed to create repository %q in namespace %q: %w", repoName, org, err)
	}
	if err := pushRepository(o, created.CloneSSH, e); err != nil {
		return fmt.Errorf("failed to push bootstrapped resources: %s", err)
	}
	return err
}

func pushRepository(o *BootstrapOptions, remote string, e executor) error {
	log.Printf("DEBUGGING: the home dir = %s", os.Getenv("HOME"))
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("DEBUGGING: could not get wd: %s", err)
	}
	log.Printf("DEBUGGING: the cwd = %s", wd)

	u, err := user.Current()
	if err != nil {
		log.Printf("DEBUGGING: could not get current user: %s", err)
	}
	log.Printf("DEBUGGING: the current user config = %+v\n", u)

	if out, err := e.execute(o.OutputPath, "git", "init", "."); err != nil {
		return fmt.Errorf("failed to initialize git repository in %q %q: %s", o.OutputPath, string(out), err)
	}
	if out, err := e.execute(o.OutputPath, "git", "add", "pipelines.yaml", "config", "environments"); err != nil {
		return fmt.Errorf("failed to add pipelines.yaml to repository in %q %q: %s", o.OutputPath, string(out), err)
	}
	if out, err := e.execute(o.OutputPath, "git", "commit", "-m", "Bootstrapped commit"); err != nil {
		return fmt.Errorf("failed to commit files to repository in %q %q: %s", o.OutputPath, string(out), err)
	}
	if out, err := e.execute(o.OutputPath, "git", "branch", "-m", "main"); err != nil {
		return fmt.Errorf("failed to switch to branch 'main' in repository in %q %q: %s", o.OutputPath, string(out), err)
	}
	if out, err := e.execute(o.OutputPath, "git", "remote", "add", "origin", remote); err != nil {
		return fmt.Errorf("failed add remote 'origin' %q to repository in %q %q: %s", remote, o.OutputPath, string(out), err)
	}
	if out, err := e.execute(o.OutputPath, "git", "push", "-u", "origin", "main"); err != nil {
		return fmt.Errorf("failed push remote to repository %q %q: %s", remote, string(out), err)
	}
	return nil
}

func repoURL(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("failed to parse %q: %w", u, err)
	}
	parsed.Path = ""
	parsed.User = nil
	return parsed.String(), nil
}

// NewCmdExecutor creates and returns an executor implementation that uses
// exec.Command to execute the commands.
func NewCmdExecutor() cmdExecutor {
	return cmdExecutor{}
}

type cmdExecutor struct {
}

func (e cmdExecutor) execute(baseDir, command string, args ...string) ([]byte, error) {
	c := exec.Command(command, args...)
	c.Dir = baseDir
	return c.CombinedOutput()
}
