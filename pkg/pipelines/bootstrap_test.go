package pipelines

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/redhat-developer/kam/pkg/pipelines/config"
	"github.com/redhat-developer/kam/pkg/pipelines/deployment"
	"github.com/redhat-developer/kam/pkg/pipelines/eventlisteners"
	"github.com/redhat-developer/kam/pkg/pipelines/ioutils"
	"github.com/redhat-developer/kam/pkg/pipelines/meta"
	res "github.com/redhat-developer/kam/pkg/pipelines/resources"
	"github.com/redhat-developer/kam/pkg/pipelines/roles"
	"github.com/redhat-developer/kam/pkg/pipelines/scm"
	"github.com/redhat-developer/kam/pkg/pipelines/secrets"
	"github.com/redhat-developer/kam/pkg/pipelines/statustracker"
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
	defer func(f secrets.PublicKeyFunc) {
		secrets.DefaultPublicKeyFunc = f
	}(secrets.DefaultPublicKeyFunc)

	secrets.DefaultPublicKeyFunc = makeTestKey(t)

	params := &BootstrapOptions{
		Prefix:               "tst-",
		GitOpsRepoURL:        testGitOpsRepo,
		ImageRepo:            "image/repo",
		GitOpsWebhookSecret:  "123",
		GitHostAccessToken:   "test-token",
		ServiceRepoURL:       testSvcRepo,
		ServiceWebhookSecret: "456",
		CommitStatusTracker:  true,
	}
	r, err := bootstrapResources(params, ioutils.NewMemoryFilesystem())
	fatalIfError(t, err)

	hookSecret, err := secrets.CreateSealedSecret(
		meta.NamespacedName("tst-cicd", "webhook-secret-tst-dev-http-api"),
		meta.NamespacedName("test-ns", "service"), "456", eventlisteners.WebhookSecretKey)
	if err != nil {
		t.Fatal(err)
	}
	want := res.Resources{
		"config/tst-cicd/base/03-secrets/webhook-secret-tst-dev-http-api.yaml": hookSecret,
		"environments/tst-dev/apps/app-http-api/services/http-api/base/config/100-deployment.yaml": deployment.Create(
			"app-http-api", "tst-dev", "http-api", bootstrapImage,
			deployment.ContainerPort(8080)),
		"environments/tst-dev/apps/app-http-api/services/http-api/base/config/200-service.yaml": createBootstrapService(
			"app-http-api", "tst-dev", "http-api"),
		"environments/tst-dev/apps/app-http-api/services/http-api/base/config/kustomization.yaml": &res.Kustomization{
			Resources: []string{"100-deployment.yaml", "200-service.yaml"}},
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
				ArgoCD:    &config.ArgoCDConfig{Namespace: "argocd"},
			},
		},
	}

	if diff := cmp.Diff(want, r, cmpopts.IgnoreMapEntries(func(k string, v interface{}) bool {
		_, ok := want[k]
		return !ok
	})); diff != "" {
		t.Fatalf("bootstrapped resources:\n%s", diff)
	}

	wantResources := []string{
		"01-namespaces/cicd-environment.yaml",
		"01-namespaces/image.yaml",
		"02-rolebindings/commit-status-tracker-role.yaml",
		"02-rolebindings/commit-status-tracker-rolebinding.yaml",
		"02-rolebindings/commit-status-tracker-service-account.yaml",
		"02-rolebindings/internal-registry-image-binding.yaml",
		"02-rolebindings/pipeline-service-account.yaml",
		"02-rolebindings/pipeline-service-role.yaml",
		"02-rolebindings/pipeline-service-rolebinding.yaml",
		"03-secrets/git-host-access-token.yaml",
		"03-secrets/git-host-basic-auth-token.yaml",
		"03-secrets/gitops-webhook-secret.yaml",
		"03-secrets/webhook-secret-tst-dev-http-api.yaml",
		"04-tasks/deploy-from-source-task.yaml",
		"05-pipelines/app-ci-pipeline.yaml",
		"05-pipelines/ci-dryrun-from-push-pipeline.yaml",
		"06-bindings/github-push-binding.yaml",
		"06-bindings/tst-dev-app-http-api-http-api-binding.yaml",
		"07-templates/app-ci-build-from-push-template.yaml",
		"07-templates/ci-dryrun-from-push-template.yaml",
		"08-eventlisteners/cicd-event-listener.yaml",
		"09-routes/gitops-webhook-event-listener.yaml",
		"10-commit-status-tracker/operator.yaml",
	}
	k := r["config/tst-cicd/base/kustomization.yaml"].(res.Kustomization)
	if diff := cmp.Diff(wantResources, k.Resources); diff != "" {
		t.Fatalf("base kustomization failed:\n%s\n", diff)
	}
}

func TestBootstrapCreatesRepository(t *testing.T) {
	defer func(f secrets.PublicKeyFunc) {
		secrets.DefaultPublicKeyFunc = f
	}(secrets.DefaultPublicKeyFunc)

	secrets.DefaultPublicKeyFunc = makeTestKey(t)
	fakeGitData := stubOutGitClientFactory(t, "test-token")

	params := &BootstrapOptions{
		Prefix:               "tst-",
		GitOpsRepoURL:        testGitOpsRepo,
		ImageRepo:            "image/repo",
		GitOpsWebhookSecret:  "123",
		GitHostAccessToken:   "test-token",
		ServiceRepoURL:       testSvcRepo,
		ServiceWebhookSecret: "456",
		CommitStatusTracker:  true,
	}
	err := Bootstrap(params, ioutils.NewMemoryFilesystem())
	fatalIfError(t, err)

	assertRepositoryCreated(t, fakeGitData, "my-org", "gitops")
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
	defer func(f secrets.PublicKeyFunc) {
		secrets.DefaultPublicKeyFunc = f
	}(secrets.DefaultPublicKeyFunc)
	_ = stubOutGitClientFactory(t, "test-token")

	secrets.DefaultPublicKeyFunc = makeTestKey(t)
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
	o := BootstrapOptions{Prefix: prefix, GitOpsWebhookSecret: gitOpsWebhook, DockerConfigJSONFilename: "", SealedSecretsService: meta.NamespacedName("", "")}

	defer stubDefaultPublicKeyFunc(t)()
	fakeFs := ioutils.NewMemoryFilesystem()
	repo, err := scm.NewRepository(gitOpsURL)
	assertNoError(t, err)
	got, err := createInitialFiles(fakeFs, repo, &o)
	assertNoError(t, err)

	want := res.Resources{
		pipelinesFile: createManifest(gitOpsURL, &config.Config{Pipelines: testpipelineConfig}),
	}
	resources, err := createCICDResources(fakeFs, repo, testpipelineConfig, &o)
	if err != nil {
		t.Fatalf("CreatePipelineResources() failed due to :%s\n", err)
	}
	files := getResourceFiles(resources)

	want = res.Merge(addPrefixToResources("config/tst-cicd/base", resources), want)
	want = res.Merge(addPrefixToResources("config/tst-cicd", getCICDKustomization(files)), want)

	if diff := cmp.Diff(want, got, cmpopts.IgnoreMapEntries(ignoreSecrets)); diff != "" {
		t.Fatalf("outputs didn't match: %s\n", diff)
	}
}

func ignoreSecrets(k string, v interface{}) bool {
	return k == "config/tst-cicd/base/03-secrets/gitops-webhook-secret.yaml"
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

func TestGenerateSecrets(t *testing.T) {
	defer stubDefaultPublicKeyFunc(t)()
	ns := "test-ns"
	outputs := res.Resources{}
	sa := roles.CreateServiceAccount(meta.NamespacedName("test-ns", "test-sa"))
	o := &BootstrapOptions{
		SealedSecretsService: meta.NamespacedName("sealed-secrets", "secrets"),
		GitHostAccessToken:   "abc123",
		ServiceRepoURL:       "https://gl.example.com/my-org/my-project.git",
		CommitStatusTracker:  true,
	}

	err := generateSecrets(outputs, sa, ns, o)
	fatalIfError(t, err)

	wantSA := &corev1.ServiceAccount{
		TypeMeta: meta.TypeMeta("ServiceAccount", "v1"),
		ObjectMeta: meta.ObjectMeta(
			types.NamespacedName{Name: "test-sa", Namespace: "test-ns"},
		),
		Secrets: []corev1.ObjectReference{{Name: statustracker.CommitStatusTrackerSecret}, {Name: "git-host-basic-auth-token"}},
	}
	if diff := cmp.Diff(wantSA, outputs[serviceAccountPath]); diff != "" {
		t.Fatalf("generatedSecrets failed to update the ServiceAccount:\n%s", diff)
	}

	wantBasicAuthSecret := &ssv1alpha1.SealedSecret{
		TypeMeta: meta.TypeMeta("SealedSecret", "bitnami.com/v1alpha1"),
		ObjectMeta: meta.ObjectMeta(
			types.NamespacedName{Name: "git-host-basic-auth-token", Namespace: "test-ns"},
		),
		Spec: ssv1alpha1.SealedSecretSpec{
			Template: ssv1alpha1.SecretTemplateSpec{
				ObjectMeta: meta.ObjectMeta(
					types.NamespacedName{Name: "git-host-basic-auth-token", Namespace: "test-ns"},
					meta.AddAnnotations(
						map[string]string{
							"tekton.dev/git-0": "https://gl.example.com",
						}),
				),
				Type: corev1.SecretTypeBasicAuth,
			},
		},
	}
	if diff := cmp.Diff(wantBasicAuthSecret, outputs[basicAuthTokenPath],
		cmpopts.IgnoreFields(ssv1alpha1.SealedSecret{}, "Spec.EncryptedData", "ObjectMeta.Annotations")); diff != "" {
		t.Fatalf("generatedSecrets failed to create basic auth token secret:\n%s", diff)
	}
	wantAuthSecret := &ssv1alpha1.SealedSecret{
		TypeMeta: meta.TypeMeta("SealedSecret", "bitnami.com/v1alpha1"),
		ObjectMeta: meta.ObjectMeta(
			types.NamespacedName{Name: statustracker.CommitStatusTrackerSecret, Namespace: "test-ns"},
		),
		Spec: ssv1alpha1.SealedSecretSpec{
			Template: ssv1alpha1.SecretTemplateSpec{
				ObjectMeta: meta.ObjectMeta(
					types.NamespacedName{Name: statustracker.CommitStatusTrackerSecret, Namespace: "test-ns"},
				),
				Type: corev1.SecretTypeOpaque,
			},
		},
	}
	if diff := cmp.Diff(wantAuthSecret, outputs[authTokenPath],
		cmpopts.IgnoreFields(ssv1alpha1.SealedSecret{}, "Spec.EncryptedData", "ObjectMeta.Annotations")); diff != "" {
		t.Fatalf("generatedSecrets failed to create auth token secret:\n%s", diff)
	}
}

func TestGenerateSecretsWithNoCommitStatusTracker(t *testing.T) {
	defer stubDefaultPublicKeyFunc(t)()
	ns := "test-ns"
	outputs := res.Resources{}
	sa := roles.CreateServiceAccount(meta.NamespacedName("test-ns", "test-sa"))
	o := &BootstrapOptions{
		SealedSecretsService: meta.NamespacedName("sealed-secrets", "secrets"),
		GitHostAccessToken:   "abc123",
		ServiceRepoURL:       "https://gl.example.com/my-org/my-project.git",
		CommitStatusTracker:  false,
	}

	err := generateSecrets(outputs, sa, ns, o)
	fatalIfError(t, err)

	wantSA := &corev1.ServiceAccount{
		TypeMeta: meta.TypeMeta("ServiceAccount", "v1"),
		ObjectMeta: meta.ObjectMeta(
			types.NamespacedName{Name: "test-sa", Namespace: "test-ns"},
		),
		Secrets: []corev1.ObjectReference{{Name: "git-host-basic-auth-token"}},
	}
	if diff := cmp.Diff(wantSA, outputs[serviceAccountPath]); diff != "" {
		t.Fatalf("generatedSecrets failed to update the ServiceAccount:\n%s", diff)
	}

	wantBasicAuthSecret := &ssv1alpha1.SealedSecret{
		TypeMeta: meta.TypeMeta("SealedSecret", "bitnami.com/v1alpha1"),
		ObjectMeta: meta.ObjectMeta(
			types.NamespacedName{Name: "git-host-basic-auth-token", Namespace: "test-ns"},
		),
		Spec: ssv1alpha1.SealedSecretSpec{
			Template: ssv1alpha1.SecretTemplateSpec{
				ObjectMeta: meta.ObjectMeta(
					types.NamespacedName{Name: "git-host-basic-auth-token", Namespace: "test-ns"},
					meta.AddAnnotations(
						map[string]string{
							"tekton.dev/git-0": "https://gl.example.com",
						}),
				),
				Type: corev1.SecretTypeBasicAuth,
			},
		},
	}
	if diff := cmp.Diff(wantBasicAuthSecret, outputs[basicAuthTokenPath],
		cmpopts.IgnoreFields(ssv1alpha1.SealedSecret{}, "Spec.EncryptedData", "ObjectMeta.Annotations")); diff != "" {
		t.Fatalf("generatedSecrets failed to create basic auth token secret:\n%s", diff)
	}
	if outputs[authTokenPath] != nil {
		t.Fatalf("auth token secret for commit status tracker generated: %#v", outputs[authTokenPath])
	}
}

func fatalIfError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func makeTestKey(t *testing.T) func(service types.NamespacedName) (*rsa.PublicKey, error) {
	return func(service types.NamespacedName) (*rsa.PublicKey, error) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("failed to generate a private RSA key: %s", err)
		}
		return &key.PublicKey, nil
	}
}
