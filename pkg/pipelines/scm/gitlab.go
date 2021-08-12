package scm

import (
	"net/url"
	"strings"

	"github.com/redhat-developer/kam/pkg/pipelines/triggers"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

const (
	gitlabPushEventFilters = "header.match('X-Gitlab-Event','Push Hook') && body.project.path_with_namespace == '%s'"
	gitlabType             = "gitlab"
)

type gitlabSpec struct {
	pushBinding string
}

func init() {
	gits[gitlabType] = newGitLab
}

func newGitLab(rawURL string) (Repository, error) {
	path, err := processRawURL(rawURL, proccessGitLabPath)
	if err != nil {
		return nil, err
	}
	return &repository{url: rawURL, path: path, spec: &gitlabSpec{pushBinding: "gitlab-push-binding"}}, nil
}

func proccessGitLabPath(parsedURL *url.URL) (string, error) {
	components, err := splitRepositoryPath(parsedURL)
	if err != nil {
		return "", err
	}
	if len(components) < 2 {
		return "", invalidRepoPathError(gitlabType, parsedURL.Path)
	}
	path := strings.Join(components, "/")
	return path, nil
}

func (r *gitlabSpec) pushBindingName() string {
	return r.pushBinding
}

func (r *gitlabSpec) pushBindingParams() []triggersv1.Param {
	return []triggersv1.Param{
		createBindingParam("gitrepositoryurl", "$(body.project.git_http_url)"),
		createBindingParam("fullname", "$(body.project.path_with_namespace)"),
		createBindingParam(triggers.GitRef, "$(extensions.ref)"),
		createBindingParam(triggers.GitCommitID, "$(body.after)"),
		createBindingParam(triggers.GitCommitDate, "$(body.commits[-1:].timestamp)"),
		createBindingParam(triggers.GitCommitMessage, "$(body.commits[-1:].message)"),
		createBindingParam(triggers.GitCommitAuthor, "$(body.commits[-1:].author.name)"),
	}
}

func (r *gitlabSpec) pushEventFilters() string {
	return gitlabPushEventFilters
}

func (r *gitlabSpec) eventInterceptor(secretNamespace, secretName string) (*triggersv1.EventInterceptor, error) {
	raw, err := secretParam(secretName, webhookSecretKey)
	if err != nil {
		return nil, err
	}
	return eventInterceptorWithSecret(gitlabType, raw), nil
}
