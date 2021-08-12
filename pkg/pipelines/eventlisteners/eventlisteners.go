package eventlisteners

import (
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redhat-developer/kam/pkg/pipelines/meta"
	"github.com/redhat-developer/kam/pkg/pipelines/scm"
)

// Filters for interceptors
const (
	// GitOpsWebhookSecret is the name of the generated secret for hooks from
	// the GitOps repository.
	GitOpsWebhookSecret = "gitops-webhook-secret" //nolint:gosec

	// WebhookSecretKey is the name of the generated secret for hooks from the
	// bootstrapped application.
	WebhookSecretKey = "webhook-secret-key"
)

var (
	eventListenerTypeMeta = meta.TypeMeta("EventListener", "triggers.tekton.dev/v1alpha1")
)

// Generate will create the required eventlisteners.
func Generate(repo scm.Repository, ns, saName, secretName string) (triggersv1.EventListener, error) {
	pushTrigger, err := repo.CreatePushTrigger("ci-dryrun-from-push", secretName, ns, "ci-dryrun-from-push-template", []string{"github-push-binding"})
	if err != nil {
		return triggersv1.EventListener{}, err
	}
	return triggersv1.EventListener{
		TypeMeta:   eventListenerTypeMeta,
		ObjectMeta: createListenerObjectMeta("cicd-event-listener", ns),
		Spec: triggersv1.EventListenerSpec{
			ServiceAccountName: saName,
			Triggers: []triggersv1.EventListenerTrigger{
				pushTrigger,
			},
		},
	}, nil
}

// CreateELFromTriggers creates an EventListener from a supplied set of
// trigger, with the provided namespace and name.
func CreateELFromTriggers(cicdNS, saName string, triggers []triggersv1.EventListenerTrigger) *triggersv1.EventListener {
	return &triggersv1.EventListener{
		TypeMeta: eventListenerTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cicd-event-listener",
			Namespace: cicdNS,
		},
		Spec: triggersv1.EventListenerSpec{
			ServiceAccountName: saName,
			Triggers:           triggers,
		},
	}
}

func createListenerObjectMeta(name, ns string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: ns,
	}
}
