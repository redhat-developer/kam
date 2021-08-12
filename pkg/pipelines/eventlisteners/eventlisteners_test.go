package eventlisteners

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redhat-developer/kam/pkg/pipelines/scm"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGenerateEventListener(t *testing.T) {
	repo, err := scm.NewRepository("http://github.com/org/test")
	if err != nil {
		t.Fatal(err)
	}
	trigger, err := repo.CreatePushTrigger("ci-dryrun-from-push", "test", "testing", "ci-dryrun-from-push-template", []string{"github-push-binding"})
	if err != nil {
		t.Fatal(err)
	}
	validEventListener := triggersv1.EventListener{
		TypeMeta: eventListenerTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cicd-event-listener",
			Namespace: "testing",
		},
		Spec: triggersv1.EventListenerSpec{
			ServiceAccountName: "pipeline",
			Triggers: []triggersv1.EventListenerTrigger{
				trigger,
			},
		},
	}
	eventListener, err := Generate(repo, "testing", "pipeline", "test")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(validEventListener, eventListener); diff != "" {
		t.Fatalf("Generate() failed:\n%s", diff)
	}
}

func TestCreateListenerObjectMeta(t *testing.T) {
	validObjectMeta := metav1.ObjectMeta{
		Name:      "sample",
		Namespace: "testing",
	}
	objectMeta := createListenerObjectMeta("sample", "testing")
	if diff := cmp.Diff(validObjectMeta, objectMeta); diff != "" {
		t.Fatalf("createListenerObjectMeta() failed:\n%s", diff)
	}
}
