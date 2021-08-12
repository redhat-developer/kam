package scm

import (
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
	CreatePushTrigger(name, secretName, secretNs, template string, bindings []string) (triggersv1.EventListenerTrigger, error)

	// Git Repository URL
	URL() string
}
