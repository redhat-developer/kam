package scm

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redhat-developer/kam/pkg/pipelines/triggers"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreatePushBindingForGithub(t *testing.T) {
	repo, err := NewRepository("http://github.com/org/test")
	assertNoError(t, err)
	want := triggersv1.TriggerBinding{
		TypeMeta: triggers.TriggerBindingTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name:      "github-push-binding",
			Namespace: "testns",
		},
		Spec: triggersv1.TriggerBindingSpec{
			Params: []triggersv1.Param{
				{
					Name:  "gitrepositoryurl",
					Value: "$(body.repository.clone_url)",
				},
				{
					Name:  "fullname",
					Value: "$(body.repository.full_name)",
				},
				{
					Name:  triggers.GitRef,
					Value: "$(body.ref)",
				},
				{
					Name:  triggers.GitCommitID,
					Value: "$(body.head_commit.id)",
				},
				{
					Name:  triggers.GitCommitDate,
					Value: "$(body.head_commit.timestamp)",
				},
				{
					Name:  triggers.GitCommitMessage,
					Value: "$(body.head_commit.message)",
				},
				{
					Name:  triggers.GitCommitAuthor,
					Value: "$(body.head_commit.author.name)",
				},
			},
		},
	}
	got, name := repo.CreatePushBinding("testns")
	if name != "github-push-binding" {
		t.Fatalf("CreatePushBinding() returned a wrong binding: want %v got %v", "github-push-binding", name)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("CreatePushBinding() failed:\n%s", diff)
	}
}

func TestCreateCDTriggersForGithub(t *testing.T) {
	repo, err := NewRepository("http://github.com/org/test")
	assertNoError(t, err)
	want := triggersv1.EventListenerTrigger{
		Name: "test",
		Bindings: []*triggersv1.EventListenerBinding{
			{Name: "test-binding"},
		},
		Template: &triggersv1.EventListenerTemplate{Name: "test-template"},
		Interceptors: []*triggersv1.EventInterceptor{
			{
				GitHub: &triggersv1.GitHubInterceptor{
					SecretRef: &triggersv1.SecretRef{SecretKey: "webhook-secret-key", SecretName: "secret", Namespace: "ns"},
				},
			},
			{
				CEL: &triggersv1.CELInterceptor{
					Filter:   fmt.Sprintf(githubPushEventFilters, "org/test"),
					Overlays: branchRefOverlay,
				},
			},
		},
	}
	got := repo.CreatePushTrigger("test", "secret", "ns", "test-template", []string{"test-binding"})
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("CreateCDTrigger() failed:\n%s", diff)
	}
}

func TestNewGitHubRepository(t *testing.T) {
	tests := []struct {
		url      string
		repoPath string
		errMsg   string
	}{
		{
			"http://github.org",
			"",
			"unable to identify driver from hostname: github.org",
		},
		{
			"http://github.com/",
			"",
			"invalid repository URL http://github.com/: path is empty",
		},
		{
			"http://github.com/foo/bar",
			"foo/bar",
			"",
		},
		{
			"https://githuB.com/foo/bar.git",
			"foo/bar",
			"",
		},
		{
			"https://githuB.com/foo/bar/test.git",
			"",
			"invalid repository path for github: /foo/bar/test.git",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(rt *testing.T) {
			repo, err := NewRepository(tt.url)
			if err != nil {
				if diff := cmp.Diff(tt.errMsg, err.Error()); diff != "" {
					rt.Fatalf("repo path errMsg mismatch: \n%s", diff)
				}
			}
			if repo != nil {
				if diff := cmp.Diff(tt.repoPath, repo.(*repository).path); diff != "" {
					rt.Fatalf("repo path mismatch: got\n%s", diff)
				}
			}
		})
	}
}
