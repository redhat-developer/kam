package meta

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLabels(t *testing.T) {
	v := AddLabels(map[string]string{
		"app":     "my-app",
		"service": "my-service",
	})
	om := &metav1.ObjectMeta{}
	v(om)

	want := &metav1.ObjectMeta{
		Labels: map[string]string{
			"app":     "my-app",
			"service": "my-service",
		},
	}

	if diff := cmp.Diff(want, om); diff != "" {
		t.Fatalf("failed to add labels:\n%s", diff)
	}
}
