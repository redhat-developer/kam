package scm

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

func TestCreateListenerBinding(t *testing.T) {
	validListenerBinding := triggersv1.EventListenerBinding{
		Ref: "sample",
	}
	listenerBinding := createListenerBinding("sample")
	if diff := cmp.Diff(validListenerBinding, *listenerBinding); diff != "" {
		t.Fatalf("createListenerBinding() failed:\n%s", diff)
	}
}

func TestCreateListenerTemplate(t *testing.T) {
	name := "sample"
	validListenerTemplate := &triggersv1.EventListenerTemplate{
		Ref: &name,
	}
	listenerTemplate := createListenerTemplate(&name)
	if diff := cmp.Diff(validListenerTemplate, listenerTemplate); diff != "" {
		t.Fatalf("createListenerTemplate() failed:\n%s", diff)
	}
}

func TestCreateEventInterceptor(t *testing.T) {
	filter := "sampleFilter %s"
	repo := "sample"
	rawFilter, rawOverlays, err := celParams(filter, repo)
	assertNoError(t, err)
	validEventInterceptor := triggersv1.EventInterceptor{
		Ref: triggersv1.InterceptorRef{
			Name: "cel",
		},
		Params: []triggersv1.InterceptorParams{
			{
				Name: "filter",
				Value: v1.JSON{
					Raw: rawFilter,
				},
			},
			{
				Name: "overlays",
				Value: v1.JSON{
					Raw: rawOverlays,
				},
			},
		},
	}
	eventInterceptor, err := createEventInterceptor("sampleFilter %s", "sample")
	assertNoError(t, err)
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
