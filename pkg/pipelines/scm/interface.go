package scm

import (
	res "github.com/redhat-developer/kam/pkg/pipelines/resources"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

// Repository interface exposes generic functions that will be
// implemented by repositories (Github,Gitlab,Bitbucket,etc)
type Repository interface {
	// Get Push TriggerBinding name for this repository provider
	PushBindingName() string

	// Create a TriggerBinding for Push Request hooks
	CreatePushBinding(namespace string) (triggersv1.TriggerBinding, string)

	// Create an eventlistener trigger for Push event
	CreatePushTrigger(name, secretName, secretNs, template string, bindings []string) triggersv1.EventListenerTrigger

	// Git Repository URL
	URL() string

	AddCommitStatusTask(path, ns string, output res.Resources)

	AddFinallyTaskToPipeline(*pipelinev1.Pipeline)

	AddFinallyTaskParams(*triggersv1.TriggerTemplate) error
}
