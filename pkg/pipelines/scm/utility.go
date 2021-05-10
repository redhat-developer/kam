package scm

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/jenkins-x/go-scm/scm/factory"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

var (
	branchRefOverlay = []triggersv1.CELOverlay{
		{Key: "ref", Expression: "body.ref.split('/')[2]"},
	}
)

func invalidRepoPathError(gitType, path string) error {
	return fmt.Errorf("invalid repository path for %s: %s", gitType, path)
}

func unsupportedGitTypeError(gitType string) error {
	return fmt.Errorf("unsupported Git repository type: %s", gitType)
}

func invalidRepoURLError(repoURL, reason string) error {
	return fmt.Errorf("invalid repository URL %s: %s", repoURL, reason)
}

func createEventInterceptor(filter, repoName string) *triggersv1.EventInterceptor {
	return &triggersv1.EventInterceptor{
		CEL: &triggersv1.CELInterceptor{
			Filter:   fmt.Sprintf(filter, repoName),
			Overlays: branchRefOverlay,
		},
	}
}

func createListenerTemplate(name string) *triggersv1.EventListenerTemplate {
	return &triggersv1.EventListenerTemplate{
		Name: name,
	}
}

func createListenerBinding(name string) *triggersv1.EventListenerBinding {
	return &triggersv1.EventListenerBinding{
		Ref: name,
	}
}

func createBindings(names []string) []*triggersv1.EventListenerBinding {
	bindings := make([]*triggersv1.EventListenerBinding, len(names))
	for i, name := range names {
		bindings[i] = createListenerBinding(name)
	}
	return bindings
}

func createBindingParam(name, value string) triggersv1.Param {
	return triggersv1.Param{
		Name:  name,
		Value: value,
	}
}

func processRawURL(rawURL string, processPath func(*url.URL) (string, error)) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	path, err := processPath(parsedURL)
	if err != nil {
		return "", err
	}
	return path, nil
}

func splitRepositoryPath(parsedURL *url.URL) ([]string, error) {
	var components []string
	for _, s := range strings.Split(parsedURL.Path, "/") {
		if s != "" {
			components = append(components, s)
		}
	}
	if len(components) < 1 {
		return nil, invalidRepoURLError(parsedURL.String(), "path is empty")
	}
	components[len(components)-1] = strings.TrimSuffix(components[len(components)-1], ".git")
	return components, nil
}

// GetDriverName gets the driver to be used for this repo url, using the go-scm
// default identifier.
func GetDriverName(rawURL string) (string, error) {
	host, err := HostnameFromURL(rawURL)
	if err != nil {
		return "", err
	}
	return factory.DefaultIdentifier.Identify(host)
}

// HostnameFromURL returns the host from a URL.
func HostnameFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return strings.ToLower(u.Host), nil
}

func createPendingCommitStatusTask(name string) pipelinev1.PipelineTask {
	return pipelinev1.PipelineTask{
		Name:    name,
		TaskRef: createTaskRef(name, pipelinev1.NamespacedTaskKind),
		Params: []pipelinev1.Param{
			createTaskParam("REPO", "$(params.REPO)"),
			createTaskParam("COMMIT_SHA", "$(params.COMMIT_SHA)"),
			createTaskParam("DESCRIPTION", "The build has started"),
			createTaskParam("STATE", "pending"),
		},
	}
}

func createFinallyCommitStatusTask(name, task string) pipelinev1.PipelineTask {
	return pipelinev1.PipelineTask{
		Name:    "final-commit-status",
		TaskRef: createTaskRef(name, pipelinev1.NamespacedTaskKind),
		Params: []pipelinev1.Param{
			createTaskParam("REPO", "$(params.REPO)"),
			createTaskParam("COMMIT_SHA", "$(params.COMMIT_SHA)"),
			createTaskParam("DESCRIPTION", "The build is completed"),
			createTaskParam("STATE", task),
		},
	}
}

func createTaskRef(name string, kind pipelinev1.TaskKind) *pipelinev1.TaskRef {
	return &pipelinev1.TaskRef{
		Name: name,
		Kind: kind,
	}
}

func createTaskParam(name, value string) pipelinev1.Param {
	return pipelinev1.Param{
		Name: name,

		Value: pipelinev1.ArrayOrString{
			Type:      pipelinev1.ParamTypeString,
			StringVal: value,
		},
	}
}

func prependTask(tasks []pipelinev1.PipelineTask, task pipelinev1.PipelineTask) []pipelinev1.PipelineTask {
	temp := []pipelinev1.PipelineTask{task}
	return append(temp, tasks...)
}

func addTemplateParamIfMissing(template *triggersv1.TriggerTemplate, newParam triggersv1.ParamSpec) {
	found := false
	for _, param := range template.Spec.Params {
		if param == newParam {
			found = true
		}
	}
	if !found {
		template.Spec.Params = append(template.Spec.Params, newParam)
	}
}

func createTemplateParamSpec(name, description string) triggersv1.ParamSpec {
	return triggersv1.ParamSpec{
		Name:        name,
		Description: description,
	}
}

func createPipelineBindingParam(name, value string) pipelinev1.Param {
	return pipelinev1.Param{
		Name: name,
		Value: pipelinev1.ArrayOrString{
			StringVal: value,
			Type:      pipelinev1.ParamTypeString,
		},
	}
}

func addPRParamIfMissing(pr *pipelinev1.PipelineRun, newParam pipelinev1.Param) {
	found := false
	for _, param := range pr.Spec.Params {
		if reflect.DeepEqual(param, newParam) {
			found = true
		}
	}
	if !found {
		pr.Spec.Params = append(pr.Spec.Params, newParam)
	}
}
