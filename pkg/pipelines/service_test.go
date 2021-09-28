package pipelines

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/redhat-developer/kam/pkg/pipelines/argocd"
	"github.com/redhat-developer/kam/pkg/pipelines/config"
	"github.com/redhat-developer/kam/pkg/pipelines/eventlisteners"
	"github.com/redhat-developer/kam/pkg/pipelines/ioutils"
	"github.com/redhat-developer/kam/pkg/pipelines/meta"
	res "github.com/redhat-developer/kam/pkg/pipelines/resources"
	"github.com/redhat-developer/kam/pkg/pipelines/secrets"
	"github.com/redhat-developer/kam/pkg/pipelines/triggers"
	"github.com/spf13/afero"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func TestServiceResourcesWithCICD(t *testing.T) {
	fakeFs := ioutils.NewMemoryFilesystem()
	m := buildManifest(true, false)
	hookSecret, err := secrets.CreateUnsealedSecret(
		meta.NamespacedName(
			"cicd", "webhook-secret-test-dev-test"),
		"123",
		eventlisteners.WebhookSecretKey)
	assertNoError(t, err)

	wantOther := res.Resources{
		"secrets/webhook-secret-test-dev-test.yaml": hookSecret,
	}
	want := res.Resources{
		"environments/test-dev/apps/test-app/base/kustomization.yaml": &res.Kustomization{
			Bases: []string{"../services/test-svc", "../services/test"}},
		"environments/test-dev/apps/test-app/kustomization.yaml": &res.Kustomization{
			Bases:        []string{"overlays"},
			CommonLabels: map[string]string{"app.openshift.io/vcs-source": "org/test"},
		},
		"environments/test-dev/apps/test-app/overlays/kustomization.yaml": &res.Kustomization{
			Bases: []string{"../base"}},
		"pipelines.yaml": &config.Manifest{
			Config: &config.Config{
				Pipelines: &config.PipelinesConfig{
					Name: "cicd",
				},
			},
			GitOpsURL: "http://github.com/org/test",
			Environments: []*config.Environment{
				{
					Name: "test-dev",
					Apps: []*config.Application{
						{
							Name: "test-app",
							Services: []*config.Service{
								{
									Name:      "test-svc",
									SourceURL: "https://github.com/myproject/test-svc",
									Webhook: &config.Webhook{
										Secret: &config.Secret{
											Name:      "webhook-secret-test-dev-test-svc",
											Namespace: "cicd",
										},
									},
								},
								{
									Name:      "test",
									SourceURL: "http://github.com/org/test",
									Webhook: &config.Webhook{
										Secret: &config.Secret{
											Name:      "webhook-secret-test-dev-test",
											Namespace: "cicd",
										},
									},
									Pipelines: &config.Pipelines{
										Integration: &config.TemplateBinding{Bindings: []string{"test-dev-test-app-test-binding", "github-push-binding"}},
									},
								},
							},
						},
					},
					Pipelines: &config.Pipelines{
						Integration: &config.TemplateBinding{Template: "app-ci-template", Bindings: []string{"github-push-binding"}},
					},
				},
			},
		},
	}

	got, otherResources, err := serviceResources(m, fakeFs, &AddServiceOptions{
		AppName:             "test-app",
		EnvName:             "test-dev",
		GitRepoURL:          "http://github.com/org/test",
		PipelinesFolderPath: pipelinesFile,
		WebhookSecret:       "123",
		ServiceName:         "test",
	})
	assertNoError(t, err)
	if diff := cmp.Diff(got, want, cmpopts.IgnoreMapEntries(func(k string, v interface{}) bool {
		_, ok := want[k]
		return !ok
	})); diff != "" {
		t.Fatalf("serviceResources() failed: %v", diff)
	}
	if diff := cmp.Diff(1, len(otherResources)); diff != "" {
		t.Fatalf("other resources contain one entry:\n%s", diff)
	}
	if diff := cmp.Diff(otherResources, wantOther); diff != "" {
		t.Fatalf("serviceResources() failed to create secrets properly:\n%s", diff)
	}
}

func TestServiceResourcesWithArgoCD(t *testing.T) {
	fakeFs := ioutils.NewMemoryFilesystem()
	m := buildManifest(false, true)

	want := res.Resources{
		"environments/test-dev/apps/test-app/base/kustomization.yaml": &res.Kustomization{
			Bases: []string{
				"../services/test-svc",
				"../services/test",
			},
		},
		"environments/test-dev/apps/test-app/kustomization.yaml": &res.Kustomization{
			Bases:        []string{"overlays"},
			CommonLabels: map[string]string{"app.openshift.io/vcs-source": "org/test"},
		},
		"environments/test-dev/apps/test-app/overlays/kustomization.yaml": &res.Kustomization{
			Bases: []string{"../base"},
		},
		"pipelines.yaml": &config.Manifest{
			Config: &config.Config{
				ArgoCD: &config.ArgoCDConfig{
					Namespace: argocd.ArgoCDNamespace,
				},
			},
			GitOpsURL: "http://github.com/org/test",
			Environments: []*config.Environment{
				{
					Name: "test-dev",
					Apps: []*config.Application{
						{
							Name: "test-app",
							Services: []*config.Service{
								{
									Name:      "test-svc",
									SourceURL: "https://github.com/myproject/test-svc",
								},
								{
									Name:      "test",
									SourceURL: "http://github.com/org/test",
								},
							},
						},
					},
				},
			},
		},
	}

	got, otherResources, err := serviceResources(m, fakeFs, &AddServiceOptions{
		AppName:             "test-app",
		EnvName:             "test-dev",
		GitRepoURL:          "http://github.com/org/test",
		PipelinesFolderPath: pipelinesFile,
		WebhookSecret:       "123",
		ServiceName:         "test",
	})
	assertNoError(t, err)
	if diff := cmp.Diff(got, want, cmpopts.IgnoreMapEntries(func(k string, v interface{}) bool {
		_, ok := want[k]
		return !ok
	})); diff != "" {
		t.Fatalf("serviceResources() failed: %v", diff)
	}
	if diff := cmp.Diff(0, len(otherResources)); diff != "" {
		t.Fatalf("other resources is not empty:\n%s", diff)
	}
}

func TestServiceResourcesWithoutArgoCD(t *testing.T) {
	fakeFs := ioutils.NewMemoryFilesystem()
	m := buildManifest(false, false)
	want := res.Resources{
		"environments/test-dev/apps/test-app/base/kustomization.yaml": &res.Kustomization{

			Bases: []string{"../services/test-svc", "../services/test"}},
		"environments/test-dev/apps/test-app/kustomization.yaml": &res.Kustomization{
			Bases:        []string{"overlays"},
			CommonLabels: map[string]string{"app.openshift.io/vcs-source": "org/test"},
		},
		"environments/test-dev/apps/test-app/overlays/kustomization.yaml": &res.Kustomization{
			Bases: []string{"../base"}},
		"environments/test-dev/env/base/kustomization.yaml": &res.Kustomization{
			Resources: []string{"test-dev-environment.yaml"},
			Bases:     []string{"../../apps/test-app/overlays"},
		},
		"pipelines.yaml": &config.Manifest{
			GitOpsURL: "http://github.com/org/test",
			Environments: []*config.Environment{
				{
					Name: "test-dev",
					Apps: []*config.Application{
						{
							Name: "test-app",
							Services: []*config.Service{
								{
									Name:      "test-svc",
									SourceURL: "https://github.com/myproject/test-svc",
								},
								{
									Name:      "test",
									SourceURL: "http://github.com/org/test",
								},
							},
						},
					},
				},
			},
		},
	}

	got, otherResources, err := serviceResources(m, fakeFs, &AddServiceOptions{
		AppName:             "test-app",
		EnvName:             "test-dev",
		GitRepoURL:          "http://github.com/org/test",
		PipelinesFolderPath: pipelinesFile,
		WebhookSecret:       "123",
		ServiceName:         "test",
	})
	assertNoError(t, err)
	if diff := cmp.Diff(want, got, cmpopts.IgnoreMapEntries(func(k string, v interface{}) bool {
		_, ok := want[k]
		return !ok
	})); diff != "" {
		t.Fatalf("serviceResources() failed: %v", diff)
	}
	if diff := cmp.Diff(0, len(otherResources)); diff != "" {
		t.Fatalf("other resources is not empty:\n%s", diff)
	}
}

func TestAddServiceWithoutApp(t *testing.T) {
	fakeFs := ioutils.NewMemoryFilesystem()
	m := buildManifest(false, false)
	want := res.Resources{
		"environments/test-dev/apps/new-app/base/kustomization.yaml":     &res.Kustomization{Bases: []string{"../services/test"}},
		"environments/test-dev/apps/new-app/overlays/kustomization.yaml": &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/apps/new-app/kustomization.yaml": &res.Kustomization{
			Bases:        []string{"overlays"},
			CommonLabels: map[string]string{"app.openshift.io/vcs-source": "org/test"},
		},
		"environments/test-dev/apps/new-app/services/test/base/kustomization.yaml":          &res.Kustomization{Bases: []string{"./config"}},
		"environments/test-dev/apps/new-app/services/test/kustomization.yaml":               &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/apps/new-app/services/test/overlays/kustomization.yaml":      &res.Kustomization{Bases: []string{"../base"}},
		"environments/cicd/base/pipelines/03-secrets/webhook-secret-test-dev-test-svc.yaml": nil,
		"pipelines.yaml": &config.Manifest{
			GitOpsURL: "http://github.com/org/test",
			Environments: []*config.Environment{
				{
					Name: "test-dev",
					Apps: []*config.Application{
						{
							Name: "test-app",
							Services: []*config.Service{
								{
									Name:      "test-svc",
									SourceURL: "https://github.com/myproject/test-svc",
								},
							},
						},
						{
							Name: "new-app",
							Services: []*config.Service{
								{Name: "test", SourceURL: "http://github.com/org/test"},
							},
						},
					},
				},
			},
		},
	}

	got, otherResources, err := serviceResources(m, fakeFs, &AddServiceOptions{
		AppName:             "new-app",
		EnvName:             "test-dev",
		GitRepoURL:          "http://github.com/org/test",
		PipelinesFolderPath: pipelinesFile,
		WebhookSecret:       "123",
		ServiceName:         "test",
	})
	assertNoError(t, err)
	for i := range want {
		if diff := cmp.Diff(want[i], got[i]); diff != "" {
			t.Fatalf("serviceResources() failed: %v", diff)
		}
	}
	if diff := cmp.Diff(0, len(otherResources)); diff != "" {
		t.Fatalf("other resources is not empty:\n%s", diff)
	}
}

func TestAddServiceFilePaths(t *testing.T) {
	fakeFs := ioutils.NewMemoryFilesystem()
	outputPath := afero.GetTempDir(fakeFs, "test")
	pipelinesPath := filepath.Join(outputPath, pipelinesFile) // Don't call filepath.ToSlash
	m := buildManifest(true, true)
	b, err := yaml.Marshal(m)
	assertNoError(t, err)
	err = afero.WriteFile(fakeFs, pipelinesPath, b, 0644)
	assertNoError(t, err)
	wantedPaths := []string{
		"environments/test-dev/apps/new-app/base/kustomization.yaml",
		"environments/test-dev/apps/new-app/overlays/kustomization.yaml",
		"environments/test-dev/apps/new-app/kustomization.yaml",
		"environments/test-dev/apps/new-app/services/test/base/kustomization.yaml",
		"environments/test-dev/apps/new-app/services/test/overlays/kustomization.yaml",
		"environments/test-dev/apps/new-app/services/test/kustomization.yaml",
		"config/cicd/base/kustomization.yaml",
		"pipelines.yaml",
		"config/argocd/test-dev-test-app-app.yaml",
		"config/argocd/test-dev-new-app-app.yaml",
	}
	err = AddService(&AddServiceOptions{
		AppName:             "new-app",
		EnvName:             "test-dev",
		GitRepoURL:          "http://github.com/org/test",
		PipelinesFolderPath: outputPath,
		WebhookSecret:       "123",
		ServiceName:         "test",
	}, fakeFs)
	assertNoError(t, err)
	for _, path := range wantedPaths {
		t.Run(fmt.Sprintf("checking path %s already exists", path), func(rt *testing.T) {
			// The inmemory version of Afero doesn't return errors
			exists, _ := fakeFs.Exists(filepath.Join(outputPath, path)) // Don't call filepath.ToSlash
			if !exists {
				t.Fatalf("The file is not present at : %v", path)
			}
		})
	}
}

func TestAddServiceFolderPaths(t *testing.T) {
	fakeFs := ioutils.NewMemoryFilesystem()
	outputPath := afero.GetTempDir(fakeFs, "test")
	pipelinesPath := filepath.Join(outputPath, pipelinesFile) // Don't call filepath.ToSlash
	m := buildManifest(true, true)
	b, err := yaml.Marshal(m)
	assertNoError(t, err)
	err = afero.WriteFile(fakeFs, pipelinesPath, b, 0644)
	assertNoError(t, err)
	wantedPaths := []string{
		"environments/test-dev/apps/new-app/services/test/base/config",
	}
	err = AddService(&AddServiceOptions{
		AppName:             "new-app",
		EnvName:             "test-dev",
		GitRepoURL:          "http://github.com/org/test",
		PipelinesFolderPath: outputPath,
		WebhookSecret:       "123",
		ServiceName:         "test",
	}, fakeFs)
	assertNoError(t, err)
	for _, path := range wantedPaths {
		t.Run(fmt.Sprintf("checking path %s already exists", path), func(rt *testing.T) {
			// The inmemory version of Afero doesn't return errors
			exists, _ := fakeFs.DirExists(filepath.Join(outputPath, path)) // Don't call filepath.ToSlash
			if !exists {
				t.Fatalf("The directory does not exist at path : %v", path)
			}
		})
	}
}

func TestServiceWithArgoCD(t *testing.T) {
	fakeFs := ioutils.NewMemoryFilesystem()
	m := buildManifest(true, true)
	want := res.Resources{
		"pipelines.yaml": &config.Manifest{
			Config: &config.Config{
				Pipelines: &config.PipelinesConfig{
					Name: "cicd",
				},
				ArgoCD: &config.ArgoCDConfig{
					Namespace: argocd.ArgoCDNamespace,
				},
			},
			GitOpsURL: "http://github.com/org/test",
			Environments: []*config.Environment{
				{
					Name: "test-dev",
					Apps: []*config.Application{
						{
							Name: "test-app",
							Services: []*config.Service{
								{
									Name:      "test-svc",
									SourceURL: "https://github.com/myproject/test-svc",
									Webhook: &config.Webhook{
										Secret: &config.Secret{
											Name:      "webhook-secret-test-dev-test-svc",
											Namespace: "cicd",
										},
									},
								},
								{
									Name:      "test",
									SourceURL: "http://github.com/org/test",
									Webhook: &config.Webhook{
										Secret: &config.Secret{
											Name:      "webhook-secret-test-dev-test",
											Namespace: "cicd",
										},
									},
									Pipelines: &config.Pipelines{
										Integration: &config.TemplateBinding{Bindings: []string{"test-dev-test-app-test-binding", "github-push-binding"}},
									},
								},
							},
						},
					},
					Pipelines: &config.Pipelines{
						Integration: &config.TemplateBinding{Template: "app-ci-template", Bindings: []string{"github-push-binding"}},
					},
				},
			},
		},
	}
	argo, err := argocd.Build(argocd.ArgoCDNamespace, "http://github.com/org/test", m)
	assertNoError(t, err)
	want = res.Merge(argo, want)
	got, otherResources, err := serviceResources(m, fakeFs, &AddServiceOptions{
		AppName:             "test-app",
		EnvName:             "test-dev",
		GitRepoURL:          "http://github.com/org/test",
		PipelinesFolderPath: pipelinesFile,
		WebhookSecret:       "123",
		ServiceName:         "test",
	})
	assertNoError(t, err)
	if diff := cmp.Diff(got, want, cmpopts.IgnoreMapEntries(func(k string, v interface{}) bool {
		_, ok := want[k]
		return !ok
	})); diff != "" {
		t.Fatalf("serviceResources() failed: %v", diff)
	}
	if diff := cmp.Diff(1, len(otherResources)); diff != "" {
		t.Fatalf("other resources is not empty:\n%s", diff)
	}
}

func TestAddServiceWithImageWithNoPipelines(t *testing.T) {
	fakeFs := ioutils.NewMemoryFilesystem()
	outputPath := afero.GetTempDir(fakeFs, "test")
	pipelinesPath := filepath.Join(outputPath, pipelinesFile) // Don't call filepath.ToSlash
	m := buildManifest(true, true)
	m.Environments = append(m.Environments, &config.Environment{
		Name: "staging",
	})
	b, err := yaml.Marshal(m)
	assertNoError(t, err)
	err = afero.WriteFile(fakeFs, pipelinesPath, b, 0644)
	assertNoError(t, err)
	wantedPaths := []string{
		"environments/staging/apps/new-app/services/test/base/config",
	}
	err = AddService(&AddServiceOptions{
		AppName:             "new-app",
		EnvName:             "staging",
		GitRepoURL:          "http://github.com/org/test",
		PipelinesFolderPath: outputPath,
		ImageRepo:           "testing/testing",
		WebhookSecret:       "123",
		ServiceName:         "test",
	}, fakeFs)
	assertNoError(t, err)
	for _, path := range wantedPaths {
		t.Run(fmt.Sprintf("checking path %s already exists", path), func(rt *testing.T) {
			// The inmemory version of Afero doesn't return errors
			exists, _ := fakeFs.DirExists(filepath.Join(outputPath, path)) // Don't call filepath.ToSlash
			if !exists {
				t.Fatalf("The directory does not exist at path : %v", path)
			}
		})
	}
}

func TestAddServiceWithoutImage(t *testing.T) {
	fakeFs := ioutils.NewMemoryFilesystem()
	outputPath := afero.GetTempDir(fakeFs, "test")
	pipelinesPath := filepath.ToSlash(filepath.Join(outputPath, pipelinesFile))
	m := buildManifest(true, true)
	m.Environments = append(m.Environments, &config.Environment{
		Name: "staging",
	})
	b, err := yaml.Marshal(m)
	assertNoError(t, err)
	err = afero.WriteFile(fakeFs, pipelinesPath, b, 0644)
	assertNoError(t, err)

	err = AddService(&AddServiceOptions{
		AppName:             "new-app",
		EnvName:             "staging",
		GitRepoURL:          "http://github.com/org/test",
		PipelinesFolderPath: outputPath,
		WebhookSecret:       "123",
		ServiceName:         "test",
	}, fakeFs)
	assertNoError(t, err)

	files := res.Resources{
		"config/cicd/base/05-bindings/staging-new-app-test-binding.yaml": triggers.CreateImageRepoBinding("cicd", "staging-new-app-test-binding", "image-registry.openshift-image-registry.svc:5000/cicd/test", "false"),
	}

	for path, resource := range files {
		t.Run(fmt.Sprintf("checking path %s already exists", path), func(rt *testing.T) {
			filePath := filepath.Join(outputPath, path) // Don't call filepath.ToSlash
			exists, err := fakeFs.Exists(filePath)
			assertNoError(t, err)
			if !exists {
				t.Fatalf("The file does not exist at path : %s", filepath.Join(outputPath, path)) // Don't call filepath.ToSlash
			}
			got, err := fakeFs.ReadFile(filePath)
			assertNoError(t, err)

			want, err := yaml.Marshal(resource)
			assertNoError(t, err)

			if diff := cmp.Diff(want, got); diff != "" {
				t.Fatalf("AddServices failed: %s", diff)
			}
		})
	}
}

func TestAddServiceWithoutGitRepoUrl(t *testing.T) {
	fakeFs := ioutils.NewMemoryFilesystem()
	m := buildManifest(false, false)
	want := res.Resources{
		"environments/test-dev/apps/new-app/base/kustomization.yaml":     &res.Kustomization{Bases: []string{"../services/test"}},
		"environments/test-dev/apps/new-app/overlays/kustomization.yaml": &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/apps/new-app/kustomization.yaml": &res.Kustomization{
			Bases:        []string{"overlays"},
			CommonLabels: map[string]string{"app.openshift.io/vcs-source": "org/test"},
		},
		"environments/test-dev/apps/new-app/services/test/base/kustomization.yaml":          &res.Kustomization{Bases: []string{"./config"}},
		"environments/test-dev/apps/new-app/services/test/kustomization.yaml":               &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/apps/new-app/services/test/overlays/kustomization.yaml":      &res.Kustomization{Bases: []string{"../base"}},
		"environments/cicd/base/pipelines/03-secrets/webhook-secret-test-dev-test-svc.yaml": nil,
		"pipelines.yaml": &config.Manifest{
			GitOpsURL: "http://github.com/org/test",
			Environments: []*config.Environment{
				{
					Name: "test-dev",
					Apps: []*config.Application{
						{
							Name: "test-app",
							Services: []*config.Service{
								{
									Name:      "test-svc",
									SourceURL: "https://github.com/myproject/test-svc",
								},
							},
						},
						{
							Name: "new-app",
							Services: []*config.Service{
								{Name: "test"},
							},
						},
					},
				},
			},
		},
	}

	got, otherResources, err := serviceResources(m, fakeFs, &AddServiceOptions{
		AppName:             "new-app",
		EnvName:             "test-dev",
		PipelinesFolderPath: pipelinesFile,
		WebhookSecret:       "123",
		ServiceName:         "test",
	})
	assertNoError(t, err)
	for i := range want {
		if diff := cmp.Diff(want[i], got[i]); diff != "" {
			t.Fatalf("serviceResources() failed: %v", diff)
		}
	}
	if diff := cmp.Diff(0, len(otherResources)); diff != "" {
		t.Fatalf("other resources is not empty:\n%s", diff)
	}
}

func buildManifest(withPipelines, withArgoCD bool) *config.Manifest {
	m := config.Manifest{
		GitOpsURL: "http://github.com/org/test",
	}
	m.Environments = environment(withPipelines)
	if withArgoCD {
		m.Config = &config.Config{
			ArgoCD: &config.ArgoCDConfig{
				Namespace: argocd.ArgoCDNamespace,
			},
		}
	}

	if withPipelines {
		if m.Config == nil {
			m.Config = &config.Config{}
		}
		m.Config.Pipelines = &config.PipelinesConfig{
			Name: "cicd",
		}
	}
	return &m
}

func environment(withPipelinesConfig bool) []*config.Environment {
	env := []*config.Environment{
		{
			Name: "test-dev",
			Apps: []*config.Application{
				{
					Name: "test-app",
					Services: []*config.Service{
						{
							Name:      "test-svc",
							SourceURL: "https://github.com/myproject/test-svc",
						},
					},
				},
			},
		},
	}

	if withPipelinesConfig {
		env[0].Apps[0].Services[0].Webhook = &config.Webhook{
			Secret: &config.Secret{
				Name:      "webhook-secret-test-dev-test-svc",
				Namespace: "cicd",
			},
		}
	}

	return env
}

func TestCreateSvcImageBinding(t *testing.T) {
	cfg := &config.PipelinesConfig{
		Name: "cicd",
	}
	env := &config.Environment{
		Name: "new-env",
	}
	bindingName, bindingFilename, resources := createSvcImageBinding(cfg, env, "newapp", "new-svc", "quay.io/user/app", false)
	if diff := cmp.Diff(bindingName, "new-env-newapp-new-svc-binding"); diff != "" {
		t.Errorf("bindingName failed: %v", diff)
	}
	if diff := cmp.Diff(bindingFilename, "05-bindings/new-env-newapp-new-svc-binding.yaml"); diff != "" {
		t.Errorf("bindingFilename failed: %v", diff)
	}

	triggerBinding := triggersv1.TriggerBinding{
		TypeMeta:   v1.TypeMeta{Kind: "TriggerBinding", APIVersion: "triggers.tekton.dev/v1alpha1"},
		ObjectMeta: v1.ObjectMeta{Name: "new-env-newapp-new-svc-binding", Namespace: "cicd"},
		Spec: triggersv1.TriggerBindingSpec{
			Params: []triggersv1.Param{
				{
					Name:  "imageRepo",
					Value: "quay.io/user/app",
				},
				{
					Name:  "tlsVerify",
					Value: "false",
				},
			},
		},
	}

	wantResources := res.Resources{"config/cicd/base/05-bindings/new-env-newapp-new-svc-binding.yaml": triggerBinding}
	if diff := cmp.Diff(resources, wantResources); diff != "" {
		t.Errorf("resources failed: %v", diff)
	}
}
