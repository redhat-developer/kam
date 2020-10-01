package scm

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

func TestCreateListenerBinding(t *testing.T) {
	validListenerBinding := triggersv1.EventListenerBinding{
		Name: "sample",
	}
	listenerBinding := createListenerBinding("sample")
	if diff := cmp.Diff(validListenerBinding, *listenerBinding); diff != "" {
		t.Fatalf("createListenerBinding() failed:\n%s", diff)
	}
}

func TestCreateListenerTemplate(t *testing.T) {
	validListenerTemplate := &triggersv1.EventListenerTemplate{
		Name: "sample",
	}
	listenerTemplate := createListenerTemplate("sample")
	if diff := cmp.Diff(validListenerTemplate, listenerTemplate); diff != "" {
		t.Fatalf("createListenerTemplate() failed:\n%s", diff)
	}
}

func TestCreateEventInterceptor(t *testing.T) {
	validEventInterceptor := triggersv1.EventInterceptor{
		CEL: &triggersv1.CELInterceptor{
			Filter:   "sampleFilter sample",
			Overlays: branchRefOverlay,
		},
	}
	eventInterceptor := createEventInterceptor("sampleFilter %s", "sample")
	if diff := cmp.Diff(validEventInterceptor, *eventInterceptor); diff != "" {
		t.Fatalf("createEventInterceptor() failed:\n%s", diff)
	}
}

func TestHostnameFromURL(t *testing.T) {
	hostTests := []struct {
		repoURL  string
		wantHost string
		wantErr  string
	}{
		{"https://github.com/example/example.git", "github.com", ""},
		{"https://example.com/example/example.git", "example.com", ""},
		{"https:/%/", "", "parse \"https:/%/\": invalid URL escape \"%/\""},
		{"https://GITHUB.COM/test/test.git", "github.com", ""},
	}

	for _, tt := range hostTests {
		h, err := HostnameFromURL(tt.repoURL)
		if tt.wantErr == "" && err != nil {
			t.Errorf("got an error %q", err)
			continue
		}
		if tt.wantErr != "" && err != nil && tt.wantErr != err.Error() {
			t.Errorf("error failed: got %q, want %q", err, tt.wantErr)
		}
		if h != tt.wantHost {
			t.Errorf("HostnameFromURL(%q) got host %q, want %q", tt.repoURL, h, tt.wantHost)
		}
	}
}
