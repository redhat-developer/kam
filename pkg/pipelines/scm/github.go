package scm

import (
	"net/url"
	"strings"

	"github.com/redhat-developer/kam/pkg/pipelines/triggers"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

const (
	githubPushEventFilters = "(header.match('X-GitHub-Event', 'push') && body.repository.full_name == '%s')"
	githubType             = "github"
)

type githubSpec struct {
	pushBinding string
}

func init() {
	gits[githubType] = newGitHub
}

func newGitHub(rawURL string) (Repository, error) {
	path, err := processRawURL(rawURL, proccessGitHubPath)
	if err != nil {
		return nil, err
	}
	return &repository{url: rawURL, path: path, spec: &githubSpec{pushBinding: "github-push-binding"}}, nil
}

func proccessGitHubPath(parsedURL *url.URL) (string, error) {
	components, err := splitRepositoryPath(parsedURL)
	if err != nil {
		return "", err
	}

	if len(components) != 2 {
		return "", invalidRepoPathError(githubType, parsedURL.Path)
	}
	path := strings.Join(components, "/")
	return path, nil
}

func (r *githubSpec) pushBindingName() string {
	return r.pushBinding
}

func (r *githubSpec) pushBindingParams() []triggersv1.Param {
	return []triggersv1.Param{
		createBindingParam("gitrepositoryurl", "$(body.repository.clone_url)"),
		createBindingParam("fullname", "$(body.repository.full_name)"),
		createBindingParam(triggers.GitRef, "$(extensions.ref)"),
		createBindingParam(triggers.GitCommitID, "$(body.head_commit.id)"),
		createBindingParam(triggers.GitCommitDate, "$(body.head_commit.timestamp)"),
		createBindingParam(triggers.GitCommitMessage, "$(body.head_commit.message)"),
		createBindingParam(triggers.GitCommitAuthor, "$(body.head_commit.author.name)"),
	}
}

func (r *githubSpec) pushEventFilters() string {
	return githubPushEventFilters
}

func (r *githubSpec) eventInterceptor(secretNamespace, secretName string) (*triggersv1.EventInterceptor, error) {
	raw, err := secretParam(secretName, webhookSecretKey)
	if err != nil {
		return nil, err
	}
	return eventInterceptorWithSecret(githubType, raw), nil
}
