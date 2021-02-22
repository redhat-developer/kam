package resources

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_AddResource(t *testing.T) {
	k := Kustomization{}
	k.AddResources("testing.yaml", "testing2.yaml")

	if diff := cmp.Diff([]string{"testing.yaml", "testing2.yaml"}, k.Resources); diff != "" {
		t.Fatalf("failed to add resources:\n%s", diff)
	}
}

func Test_AddResource_with_duplicates(t *testing.T) {
	k := Kustomization{}
	k.AddResources("testing.yaml", "testing2.yaml")
	k.AddResources("testing.yaml")

	if diff := cmp.Diff([]string{"testing.yaml", "testing2.yaml"}, k.Resources); diff != "" {
		t.Fatalf("failed to add resources:\n%s", diff)
	}
}

func Test_AddResource_sorts_elements(t *testing.T) {
	k := Kustomization{}
	k.AddResources("service.yaml", "deployment.yaml", "namespace.yaml")

	if diff := cmp.Diff([]string{"deployment.yaml", "namespace.yaml", "service.yaml"}, k.Resources); diff != "" {
		t.Fatalf("failed to sort resources:\n%s", diff)
	}
}
