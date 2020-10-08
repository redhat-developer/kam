package pipelines

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
)

var defaultClientFactory = factory.FromRepoURL

const defaultRepoDescription = "Bootstrapped GitOps Repository"

// BootstrapRepository creates a new empty Git repository in the upstream git
// hosting service from the GitOpsRepoURL.
func BootstrapRepository(o *BootstrapOptions) error {
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

	client, err := defaultClientFactory(u.String())
	if err != nil {
		return fmt.Errorf("failed to create a client to access %q: %w", o.GitOpsRepoURL, err)
	}
	ctx := context.Background()
	// If we're creating the repository in a personal user's account, it's a
	// different API call that's made, clearing the org triggers go-scm to use
	// the "create repo in personal account" endpoint.
	currentUser, _, err := client.Users.Find(ctx)
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
		return fmt.Errorf("failed to create repository %q in namespace %q: %w", repoName, org, err)
	}
	return err
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
