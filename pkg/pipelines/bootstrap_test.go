package pipelines

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/redhat-developer/kam/pkg/pipelines/argocd"
	"github.com/redhat-developer/kam/pkg/pipelines/config"
	"github.com/redhat-developer/kam/pkg/pipelines/deployment"
	"github.com/redhat-developer/kam/pkg/pipelines/eventlisteners"
	"github.com/redhat-developer/kam/pkg/pipelines/ioutils"
	"github.com/redhat-developer/kam/pkg/pipelines/meta"
	res "github.com/redhat-developer/kam/pkg/pipelines/resources"
	"github.com/redhat-developer/kam/pkg/pipelines/roles"
	"github.com/redhat-developer/kam/pkg/pipelines/routes"
	"github.com/redhat-developer/kam/pkg/pipelines/scm"
	"github.com/redhat-developer/kam/pkg/pipelines/secrets"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	testSvcRepo    = "https://github.com/my-org/http-api.git"
	testGitOpsRepo = "https://github.com/my-org/gitops.git"
)

var testpipelineConfig = &config.PipelinesConfig{Name: "tst-cicd"}
var testArgoCDConfig = &config.ArgoCDConfig{Namespace: "tst-argocd"}
var Config = &config.Config{ArgoCD: testArgoCDConfig, Pipelines: testpipelineConfig}

func TestBootstrapManifest(t *testing.T) {

	params := &BootstrapOptions{
		Prefix:               "tst-",
		GitOpsRepoURL:        testGitOpsRepo,
		ImageRepo:            "image/repo",
		GitOpsWebhookSecret:  "123",
		GitHostAccessToken:   "test-token",
		ServiceRepoURL:       testSvcRepo,
		ServiceWebhookSecret: "456",
	}
	r, otherResources, err := bootstrapResources(params, ioutils.NewMemoryFilesystem())
	fatalIfError(t, err)

	otherResourcesNotEmpty := 4
	if diff := cmp.Diff(otherResourcesNotEmpty, len(otherResources)); diff != "" {
		t.Fatalf("other resources is empty:\n%s", diff)
	}

	hookSecret, err := secrets.CreateUnsealedSecret(
		meta.NamespacedName("tst-cicd", "webhook-secret-tst-dev-http-api"), "456", eventlisteners.WebhookSecretKey)
	if err != nil {
		t.Fatal(err)
	}
	svc := createBootstrapService("app-http-api", "tst-dev", "http-api")
	route, err := routes.NewFromService(svc)
	if err != nil {
		t.Fatal(err)
	}
	wantOther := res.Resources{
		"secrets/webhook-secret-tst-dev-http-api.yaml": hookSecret,
	}
	if diff := cmp.Diff(wantOther, otherResources, cmpopts.IgnoreMapEntries(func(k string, v interface{}) bool {
		_, ok := wantOther[k]
		return !ok
	})); diff != "" {
		t.Fatalf("bootstrapped resources:\n%s", diff)
	}

	want := res.Resources{
		"environments/tst-dev/apps/app-http-api/services/http-api/base/config/100-deployment.yaml": deployment.Create(
			"app-http-api", "tst-dev", "http-api", bootstrapImage,
			deployment.ContainerPort(8080)),
		"environments/tst-dev/apps/app-http-api/services/http-api/base/config/200-service.yaml": svc,
		"environments/tst-dev/apps/app-http-api/services/http-api/base/config/300-route.yaml":   route,
		"environments/tst-dev/apps/app-http-api/services/http-api/base/config/kustomization.yaml": &res.Kustomization{
			Resources: []string{"100-deployment.yaml", "200-service.yaml", "300-route.yaml"}},
		pipelinesFile: &config.Manifest{
			Version:   version,
			GitOpsURL: "https://github.com/my-org/gitops.git",
			Environments: []*config.Environment{
				{
					Pipelines: &config.Pipelines{
						Integration: &config.TemplateBinding{
							Template: "app-ci-template",
							Bindings: []string{"github-push-binding"},
						},
					},
					Name: "tst-dev",

					Apps: []*config.Application{
						{
							Name: "app-http-api",
							Services: []*config.Service{
								{
									Name:      "http-api",
									SourceURL: testSvcRepo,
									Webhook: &config.Webhook{
										Secret: &config.Secret{
											Name:      "webhook-secret-tst-dev-http-api",
											Namespace: "tst-cicd",
										},
									},
									Pipelines: &config.Pipelines{
										Integration: &config.TemplateBinding{Bindings: []string{"tst-dev-app-http-api-http-api-binding", "github-push-binding"}},
									},
								},
							},
						},
					},
				},
				{Name: "tst-stage"},
			},
			Config: &config.Config{
				Pipelines: &config.PipelinesConfig{Name: "tst-cicd"},
				ArgoCD:    &config.ArgoCDConfig{Namespace: argocd.ArgoCDNamespace},
			},
		},
	}

	if diff := cmp.Diff(want, r, cmpopts.IgnoreMapEntries(func(k string, v interface{}) bool {
		_, ok := want[k]
		return !ok
	})); diff != "" {
		t.Fatalf("bootstrapped resources:\n%s", diff)
	}
	// No secrets in output
	wantResources := []string{
		"01-namespaces/cicd-environment.yaml",
		"01-namespaces/image-environment.yaml",
		"02-rolebindings/internal-registry-image-binding.yaml",
		"02-rolebindings/pipeline-service-account.yaml",
		"02-rolebindings/pipeline-service-role.yaml",
		"02-rolebindings/pipeline-service-rolebinding.yaml",
		"03-tasks/deploy-from-source-task.yaml",
		"03-tasks/set-commit-status-task.yaml",
		"04-pipelines/app-ci-pipeline.yaml",
		"04-pipelines/ci-dryrun-from-push-pipeline.yaml",
		"05-bindings/github-push-binding.yaml",
		"05-bindings/tst-dev-app-http-api-http-api-binding.yaml",
		"06-templates/app-ci-build-from-push-template.yaml",
		"06-templates/ci-dryrun-from-push-template.yaml",
		"07-eventlisteners/cicd-event-listener.yaml",
		"08-routes/gitops-webhook-event-listener.yaml",
	}
	k := r["config/tst-cicd/base/kustomization.yaml"].(res.Kustomization)
	if diff := cmp.Diff(wantResources, k.Resources); diff != "" {
		t.Fatalf("base kustomization failed:\n%s\n", diff)
	}
}

func TestBootstrapCreatesRepository(t *testing.T) {
	params := &BootstrapOptions{
		Prefix:               "tst-",
		GitOpsRepoURL:        testGitOpsRepo,
		ImageRepo:            "image/repo",
		GitOpsWebhookSecret:  "123",
		GitHostAccessToken:   "test-token",
		ServiceRepoURL:       testSvcRepo,
		ServiceWebhookSecret: "456",
	}
	err := Bootstrap(params, ioutils.NewMemoryFilesystem())
	fatalIfError(t, err)
}

func TestOrgRepoFromURL(t *testing.T) {
	want := "my-org/gitops"
	got, err := orgRepoFromURL(testGitOpsRepo)
	fatalIfError(t, err)
	if got != want {
		t.Fatalf("orgRepFromURL(%s) got %s, want %s", testGitOpsRepo, got, want)
	}
}

func TestApplicationFromRepo(t *testing.T) {
	want := &config.Application{
		Name: "app-http-api",
		Services: []*config.Service{
			{

				Name: "http-api",
			},
		},
	}
	svc := &config.Service{
		Name: "http-api",
	}

	got, err := applicationFromRepo(testSvcRepo, svc)
	fatalIfError(t, err)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("bootstrapped resources:\n%s", diff)
	}
}

func TestOverwriteFlag(t *testing.T) {
	fakeFs := ioutils.NewMemoryFilesystem()
	params := &BootstrapOptions{
		Prefix:               "tst-",
		GitOpsRepoURL:        testGitOpsRepo,
		ImageRepo:            "image/repo",
		GitOpsWebhookSecret:  "123",
		GitHostAccessToken:   "test-token",
		ServiceRepoURL:       testSvcRepo,
		ServiceWebhookSecret: "456",
	}
	err := Bootstrap(params, fakeFs)
	fatalIfError(t, err)

	got := Bootstrap(params, fakeFs)
	want := "pipelines.yaml in output path already exists. If you want to replace your existing files, please rerun with --overwrite"
	if diff := cmp.Diff(want, got.Error()); diff != "" {
		t.Fatalf("overwrite failed:\n%s", diff)
	}
}

func TestOverwriteFlagExistingGitDirectory(t *testing.T) {
	fakeFs := ioutils.NewMemoryFilesystem()
	params := &BootstrapOptions{
		Prefix:               "tst-",
		GitOpsRepoURL:        testGitOpsRepo,
		ImageRepo:            "image/repo",
		GitOpsWebhookSecret:  "123",
		GitHostAccessToken:   "test-token",
		ServiceRepoURL:       testSvcRepo,
		ServiceWebhookSecret: "456",
		OutputPath:           "/tmp",
		PushToGit:            true,
	}
	err := fakeFs.MkdirAll(filepath.Join(params.OutputPath, ".git"), 0755)
	assertNoError(t, err)

	got := Bootstrap(params, fakeFs)
	want := ".git in output path already exists. If you want to replace your existing files, please rerun with --overwrite"
	if diff := cmp.Diff(want, got.Error()); diff != "" {
		t.Fatalf("overwrite failed:\n%s", diff)
	}

	params.Overwrite = true
	err = Bootstrap(params, fakeFs)
	fatalIfError(t, err)
}

func TestCreateManifest(t *testing.T) {
	repoURL := "https://github.com/foo/bar.git"
	want := &config.Manifest{
		GitOpsURL: repoURL,
		Config:    Config,
		Version:   version,
	}
	got := createManifest(repoURL, Config)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("pipelines didn't match: %s\n", diff)
	}
}

func TestInitialFiles(t *testing.T) {
	prefix := "tst-"
	gitOpsURL := "https://github.com/foo/test-repo"
	gitOpsWebhook := "123"
	o := BootstrapOptions{Prefix: prefix, GitOpsWebhookSecret: gitOpsWebhook, DockerConfigJSONFilename: ""}

	fakeFs := ioutils.NewMemoryFilesystem()
	repo, err := scm.NewRepository(gitOpsURL)
	assertNoError(t, err)
	got, _, err := createInitialFiles(fakeFs, repo, &o)
	assertNoError(t, err)

	want := res.Resources{
		pipelinesFile: createManifest(gitOpsURL, &config.Config{Pipelines: testpipelineConfig}),
	}
	resources, _, err := createCICDResources(fakeFs, repo, testpipelineConfig, &o)
	if err != nil {
		t.Fatalf("CreatePipelineResources() failed due to :%s\n", err)
	}
	files := getResourceFiles(resources)

	want = res.Merge(addPrefixToResources("config/tst-cicd/base", resources), want)
	want = res.Merge(addPrefixToResources("config/tst-cicd", getCICDKustomization(files)), want)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("outputs didn't match: %s\n", diff)
	}
}

func TestGetCICDKustomization(t *testing.T) {
	want := res.Resources{
		"overlays/kustomization.yaml": res.Kustomization{
			Bases: []string{"../base"},
		},
		"base/kustomization.yaml": res.Kustomization{
			Resources: []string{"resource1", "resource2"},
		},
	}
	got := getCICDKustomization([]string{"resource1", "resource2"})
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("getCICDKustomization was not correct: %s\n", diff)
	}
}

func TestGenerateSecrets(t *testing.T) {
	ns := "test-ns"
	outputs := res.Resources{}
	otherOutputs := res.Resources{}
	sa := roles.CreateServiceAccount(meta.NamespacedName("test-ns", "test-sa"))
	o := &BootstrapOptions{
		GitHostAccessToken: "abc123",
		ServiceRepoURL:     "https://gl.example.com/my-org/my-project.git",
	}

	err := generateSecrets(outputs, otherOutputs, sa, ns, o)
	fatalIfError(t, err)

	wantSA := &corev1.ServiceAccount{
		TypeMeta: meta.TypeMeta("ServiceAccount", "v1"),
		ObjectMeta: meta.ObjectMeta(
			types.NamespacedName{Name: "test-sa", Namespace: "test-ns"},
		),
		Secrets: []corev1.ObjectReference{{Name: authTokenSecretName}, {Name: basicAuthTokenName}},
	}
	if diff := cmp.Diff(wantSA, outputs[serviceAccountPath]); diff != "" {
		t.Fatalf("generatedSecrets failed to update the ServiceAccount:\n%s", diff)
	}
}
func TestAddPrefixToResources(t *testing.T) {
	files := map[string]interface{}{
		"base/kustomization.yaml": map[string]interface{}{
			"resources": []string{},
		},
		"overlays/kustomization.yaml": map[string]interface{}{
			"bases": []string{"../base"},
		},
	}

	want := map[string]interface{}{
		"test-prefix/base/kustomization.yaml": map[string]interface{}{
			"resources": []string{},
		},
		"test-prefix/overlays/kustomization.yaml": map[string]interface{}{
			"bases": []string{"../base"},
		},
	}
	if diff := cmp.Diff(want, addPrefixToResources("test-prefix", files)); diff != "" {
		t.Fatalf("addPrefixToResources failed, diff %s\n", diff)
	}
}

func fatalIfError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
