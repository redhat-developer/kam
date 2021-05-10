package scm

import (
	"encoding/json"
	"net/url"
	"strings"

	res "github.com/redhat-developer/kam/pkg/pipelines/resources"
	"github.com/redhat-developer/kam/pkg/pipelines/tasks"
	"github.com/redhat-developer/kam/pkg/pipelines/triggers"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
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

func (r *githubSpec) eventInterceptor(secretNamespace, secretName string) *triggersv1.EventInterceptor {
	return &triggersv1.EventInterceptor{
		GitHub: &triggersv1.GitHubInterceptor{
			SecretRef: &triggersv1.SecretRef{
				SecretName: secretName,
				SecretKey:  webhookSecretKey,
			},
		},
	}
}

func (r *githubSpec) addCommitStatusTask(path, ns string, output res.Resources) {
	output[path] = tasks.CreateCommitStatusTask(ns)
}

// for a given pipeline:
// 1. add an initial task that sets the status as pending
// 2. add the finally task that sets the final status
func (r *githubSpec) addFinallyTaskToPipeline(pipeline *pipelinev1.Pipeline) {
	pipeline.Spec.Tasks = prependTask(pipeline.Spec.Tasks, createPendingCommitStatusTask("set-commit-status"))
	task := "$(tasks.build-image.status)"
	if pipeline.Name == "ci-dryrun-from-push-pipeline" {
		task = "$(tasks.apply-source.status)"
	}
	pipeline.Spec.Finally = []pipelinev1.PipelineTask{createFinallyCommitStatusTask("set-commit-status", task)}
}

// add the missing params in the trigger template and pipelinerun
func (r *githubSpec) addFinallyTaskParams(template *triggersv1.TriggerTemplate) error {
	// add trigger template params
	reqTemplateParam := createTemplateParamSpec("fullname", "The repository name for this PullRequest.")
	addTemplateParamIfMissing(template, reqTemplateParam)

	// add pipelinerun params
	reqPipelineRunParms := []pipelinev1.Param{
		createPipelineBindingParam("REPO", "$(tt.params.fullname)"),
		createPipelineBindingParam("GIT_REPO", "$(tt.params.gitrepositoryurl)"),
		createPipelineBindingParam("COMMIT_SHA", "$(tt.params."+triggers.GitCommitID+")"),
	}
	prByte := template.Spec.ResourceTemplates[0].Raw
	pr := &pipelinev1.PipelineRun{}
	err := json.Unmarshal(prByte, pr)
	if err != nil {
		return err
	}

	for _, param := range reqPipelineRunParms {
		addPRParamIfMissing(pr, param)
	}

	prByte, err = json.Marshal(pr)
	if err != nil {
		return err
	}
	template.Spec.ResourceTemplates[0].Raw = prByte
	return nil
}
