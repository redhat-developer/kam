package argocd

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/config"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/meta"

	// This is a hack because ArgoCD doesn't support a compatible (code-wise)
	// version of k8s in common with odo.
	argov1 "github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/argocd/operator/v1alpha1"
	argoappv1 "github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/argocd/v1alpha1"
	res "github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/resources"
)

const testRepoURL = "https://github.com/rhd-example-gitops/example"

var (
	testApp = &config.Application{
		Name: "http-api",
	}
	configRepoApp = &config.Application{
		Name: "prod-api",
		ConfigRepo: &config.Repository{
			URL:            "https://github.com/rhd-example-gitops/other-repo",
			Path:           "deploys",
			TargetRevision: "master",
		},
	}

	testEnv = &config.Environment{
		Name: "test-dev",
		Apps: []*config.Application{
			testApp,
		},
	}
)

func TestBuildCreatesArgoCD(t *testing.T) {
	m := &config.Manifest{
		Environments: []*config.Environment{
			testEnv,
		},
		Config: &config.Config{
			ArgoCD: &config.ArgoCDConfig{Namespace: "argocd"},
		},
	}

	files, err := Build(ArgoCDNamespace, testRepoURL, m)
	if err != nil {
		t.Fatal(err)
	}

	want := res.Resources{
		"config/argocd/test-dev-http-api-app.yaml": &argoappv1.Application{
			TypeMeta:   applicationTypeMeta,
			ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ArgoCDNamespace, "test-dev-http-api")),
			Spec: argoappv1.ApplicationSpec{
				Source: makeSource(testEnv, testEnv.Apps[0], testRepoURL),
				Destination: argoappv1.ApplicationDestination{
					Server:    defaultServer,
					Namespace: "test-dev",
				},
				Project:    defaultProject,
				SyncPolicy: syncPolicy,
			},
		},
		"config/argocd/argo-app.yaml":      fakeArgoApplication(),
		"config/argocd/argocd.yaml":        fakeArgoCDResource(t, ArgoCDNamespace),
		"config/argocd/kustomization.yaml": &res.Kustomization{Resources: []string{"argo-app.yaml", "argocd.yaml", "test-dev-http-api-app.yaml"}},
	}

	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestBuildCreatesArgoCDWithMultipleApps(t *testing.T) {
	prodEnv := &config.Environment{
		Name: "test-production",
		Apps: []*config.Application{
			testApp,
		},
	}
	m := &config.Manifest{
		Environments: []*config.Environment{
			prodEnv,
			testEnv,
		},
		Config: &config.Config{
			ArgoCD: &config.ArgoCDConfig{Namespace: "argocd"},
		},
	}

	files, err := Build(ArgoCDNamespace, testRepoURL, m)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 5 {
		t.Fatalf("got %d files, want 3\n", len(files))
	}
	want := &res.Kustomization{Resources: []string{"argo-app.yaml", "argocd.yaml", "test-dev-http-api-app.yaml", "test-production-http-api-app.yaml"}}
	if diff := cmp.Diff(want, files["config/argocd/kustomization.yaml"]); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestBuildWithNoRepoURL(t *testing.T) {
	m := &config.Manifest{
		Environments: []*config.Environment{
			testEnv,
		},
		Config: &config.Config{
			ArgoCD: &config.ArgoCDConfig{Namespace: "argocd"},
		},
	}

	files, err := Build(ArgoCDNamespace, "", m)
	if err != nil {
		t.Fatal(err)
	}
	want := res.Resources{}
	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestBuildWithNoArgoCDConfig(t *testing.T) {
	m := &config.Manifest{
		Environments: []*config.Environment{
			testEnv,
		},
	}

	files, err := Build(ArgoCDNamespace, testRepoURL, m)
	if err != nil {
		t.Fatal(err)
	}
	want := res.Resources{}
	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestBuildWithRepoConfig(t *testing.T) {
	prodEnv := &config.Environment{
		Name: "test-production",
		Apps: []*config.Application{
			configRepoApp,
		},
	}

	m := &config.Manifest{
		Environments: []*config.Environment{
			prodEnv,
		},
		Config: &config.Config{
			ArgoCD: &config.ArgoCDConfig{Namespace: "argocd"},
		},
	}

	files, err := Build(ArgoCDNamespace, testRepoURL, m)
	if err != nil {
		t.Fatal(err)
	}

	want := res.Resources{
		"config/argocd/test-production-prod-api-app.yaml": &argoappv1.Application{
			TypeMeta:   applicationTypeMeta,
			ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ArgoCDNamespace, "test-production-prod-api")),
			Spec: argoappv1.ApplicationSpec{
				Source: makeSource(prodEnv, prodEnv.Apps[0], testRepoURL),
				Destination: argoappv1.ApplicationDestination{
					Server:    defaultServer,
					Namespace: "test-production",
				},
				Project:    defaultProject,
				SyncPolicy: syncPolicy,
			},
		},
		"config/argocd/argo-app.yaml":      fakeArgoApplication(),
		"config/argocd/argocd.yaml":        fakeArgoCDResource(t, ArgoCDNamespace),
		"config/argocd/kustomization.yaml": &res.Kustomization{Resources: []string{"argo-app.yaml", "argocd.yaml", "test-production-prod-api-app.yaml"}},
	}

	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestBuildAddsClusterToApp(t *testing.T) {
	testEnv = &config.Environment{
		Name:    "test-dev",
		Cluster: "not.real.cluster",
		Apps: []*config.Application{
			testApp,
		},
	}

	m := &config.Manifest{
		Config: &config.Config{
			ArgoCD: &config.ArgoCDConfig{Namespace: "argocd"},
		},
		Environments: []*config.Environment{
			testEnv,
		},
	}

	files, err := Build(ArgoCDNamespace, testRepoURL, m)
	if err != nil {
		t.Fatal(err)
	}

	want := res.Resources{
		"config/argocd/test-dev-http-api-app.yaml": &argoappv1.Application{
			TypeMeta:   applicationTypeMeta,
			ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ArgoCDNamespace, "test-dev-http-api")),
			Spec: argoappv1.ApplicationSpec{
				Source: makeSource(testEnv, testEnv.Apps[0], testRepoURL),
				Destination: argoappv1.ApplicationDestination{
					Server:    "not.real.cluster",
					Namespace: "test-dev",
				},
				Project:    defaultProject,
				SyncPolicy: syncPolicy,
			},
		},
		"config/argocd/argo-app.yaml":      fakeArgoApplication(),
		"config/argocd/argocd.yaml":        fakeArgoCDResource(t, ArgoCDNamespace),
		"config/argocd/kustomization.yaml": &res.Kustomization{Resources: []string{"argo-app.yaml", "argocd.yaml", "test-dev-http-api-app.yaml"}},
	}

	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestIgnoreDifferences(t *testing.T) {
	want := &argoappv1.Application{
		TypeMeta:   applicationTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ArgoCDNamespace, "argo-app")),
		Spec: argoappv1.ApplicationSpec{
			Source:      argoappv1.ApplicationSource{Path: "config/argocd"},
			Destination: argoappv1.ApplicationDestination{Server: "https://kubernetes.default.svc", Namespace: "argocd"},
			Project:     "default",
		},
	}
	got := ignoreDifferences(want)
	want.Spec.IgnoreDifferences = ignoreDifferencesFields
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("ignoreDifferences() failed: %s", diff)
	}
}

func fakeArgoApplication() *argoappv1.Application {
	return &argoappv1.Application{
		TypeMeta:   applicationTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ArgoCDNamespace, "argo-app")),
		Spec: argoappv1.ApplicationSpec{
			Source:            argoappv1.ApplicationSource{Path: "config/argocd"},
			Destination:       argoappv1.ApplicationDestination{Server: "https://kubernetes.default.svc", Namespace: "argocd"},
			Project:           "default",
			SyncPolicy:        &argoappv1.SyncPolicy{Automated: &argoappv1.SyncPolicyAutomated{Prune: true, SelfHeal: true}},
			IgnoreDifferences: ignoreDifferencesFields,
		},
	}
}

func fakeArgoCDResource(t *testing.T, ns string) *argov1.ArgoCD {
	res, err := argoCDResource(ns)
	if err != nil {
		t.Fatal(err)
	}
	return res
}
