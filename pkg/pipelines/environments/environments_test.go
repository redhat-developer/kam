package environments

import (
	"os"
	"runtime"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redhat-developer/kam/pkg/pipelines/config"
	"github.com/redhat-developer/kam/pkg/pipelines/ioutils"
	"github.com/redhat-developer/kam/pkg/pipelines/namespaces"
	res "github.com/redhat-developer/kam/pkg/pipelines/resources"
	"github.com/spf13/afero"
)

const testGitOpsRepoURL = "https://github.com/example/example.git"

func TestBuildEnvironmentFilesWithAppsToEnvironment(t *testing.T) {
	var appFs = ioutils.NewMemoryFilesystem()
	m := buildManifestWithCICD()

	files, err := Build(appFs, m, "pipelines", AppsToEnvironments)
	if err != nil {
		t.Fatal(err)
	}
	want := res.Resources{
		"environments/test-dev/apps/my-app-1/base/kustomization.yaml": &res.Kustomization{
			Bases: []string{
				"../services/service-http",
				"../services/service-metrics",
			},
		},
		"environments/test-dev/apps/my-app-1/kustomization.yaml": &res.Kustomization{
			Bases: []string{"overlays"},
			CommonLabels: map[string]string{
				vcsSourceLabel: "example/example",
			}},
		"environments/test-dev/apps/my-app-1/overlays/kustomization.yaml":                          &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/env/base/test-dev-environment.yaml":                                 namespaces.Create("test-dev", testGitOpsRepoURL),
		"environments/test-dev/env/base/test-dev-rolebinding.yaml":                                 createRoleBinding(m.Environments[0], "cicd", "pipelines"),
		"environments/test-dev/env/base/kustomization.yaml":                                        &res.Kustomization{Resources: []string{"test-dev-environment.yaml", "test-dev-rolebinding.yaml"}},
		"environments/test-dev/env/overlays/kustomization.yaml":                                    &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/apps/my-app-1/services/service-http/kustomization.yaml":             &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/apps/my-app-1/services/service-http/base/kustomization.yaml":        &res.Kustomization{Bases: []string{"./config"}},
		"environments/test-dev/apps/my-app-1/services/service-http/overlays/kustomization.yaml":    &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/apps/my-app-1/services/service-metrics/kustomization.yaml":          &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/apps/my-app-1/services/service-metrics/base/kustomization.yaml":     &res.Kustomization{Bases: []string{"./config"}},
		"environments/test-dev/apps/my-app-1/services/service-metrics/overlays/kustomization.yaml": &res.Kustomization{Bases: []string{"../base"}},
	}

	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestBuildEnvironmentFilesWithEnvironmentsToApps(t *testing.T) {
	var appFs = ioutils.NewMemoryFilesystem()
	m := buildManifestWithCICD()

	files, err := Build(appFs, m, "pipelines", EnvironmentsToApps)
	if err != nil {
		t.Fatal(err)
	}
	want := res.Resources{
		"environments/test-dev/apps/my-app-1/base/kustomization.yaml": &res.Kustomization{
			Bases: []string{
				"../services/service-http",
				"../services/service-metrics",
			},
		},
		"environments/test-dev/apps/my-app-1/kustomization.yaml": &res.Kustomization{
			Bases: []string{"overlays"},
			CommonLabels: map[string]string{
				vcsSourceLabel: "example/example",
			},
		},
		"environments/test-dev/apps/my-app-1/overlays/kustomization.yaml": &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/env/base/test-dev-environment.yaml":        namespaces.Create("test-dev", testGitOpsRepoURL),
		"environments/test-dev/env/base/test-dev-rolebinding.yaml":        createRoleBinding(m.Environments[0], "cicd", "pipelines"),
		"environments/test-dev/env/base/kustomization.yaml": &res.Kustomization{
			Resources: []string{"test-dev-environment.yaml", "test-dev-rolebinding.yaml"},
			Bases:     []string{"../../apps/my-app-1/overlays"},
		},
		"environments/test-dev/env/overlays/kustomization.yaml":                                    &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/apps/my-app-1/services/service-http/kustomization.yaml":             &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/apps/my-app-1/services/service-http/base/kustomization.yaml":        &res.Kustomization{Bases: []string{"./config"}},
		"environments/test-dev/apps/my-app-1/services/service-http/overlays/kustomization.yaml":    &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/apps/my-app-1/services/service-metrics/kustomization.yaml":          &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/apps/my-app-1/services/service-metrics/base/kustomization.yaml":     &res.Kustomization{Bases: []string{"./config"}},
		"environments/test-dev/apps/my-app-1/services/service-metrics/overlays/kustomization.yaml": &res.Kustomization{Bases: []string{"../base"}},
	}

	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestBuildEnvironmentsDoesNotOutputCIorArgo(t *testing.T) {
	var appFs = ioutils.NewMemoryFilesystem()
	m := &config.Manifest{
		Config: &config.Config{
			Pipelines: &config.PipelinesConfig{
				Name: "cicd",
			},
			ArgoCD: &config.ArgoCDConfig{
				Namespace: "argocd",
			},
		},
	}

	files, err := Build(appFs, m, "pipelines", AppsToEnvironments)
	if err != nil {
		t.Fatal(err)
	}

	want := res.Resources{}
	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func mustWriteFile(t *testing.T, fs afero.Fs, path string, data []byte, perm os.FileMode) {
	t.Helper()
	err := afero.WriteFile(fs, path, data, perm)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildEnvironmentsAddsKustomizedFiles(t *testing.T) {
	var appFs = ioutils.NewMemoryFilesystem()
	err := appFs.MkdirAll("environments/test-dev/base", 0755)
	if err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, appFs, "environments/test-dev/base/volume.yaml", []byte(`this is a file`), 0644)
	mustWriteFile(t, appFs, "environments/test-dev/base/test-dev-environment.yaml", []byte(`this is a file`), 0644)
	mustWriteFile(t, appFs, "environments/test-dev/base/routes/01-route.yaml", []byte(`this is a file`), 0644)

	m := &config.Manifest{
		Config: &config.Config{
			Pipelines: &config.PipelinesConfig{
				Name: "cicd",
			},
		},
		Environments: []*config.Environment{
			{Name: "test-dev"},
		},
	}

	resources, err := Build(appFs, m, "pipelines", EnvironmentsToApps)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"environments/test-dev/env/base/kustomization.yaml",
		"environments/test-dev/env/base/test-dev-environment.yaml",
		"environments/test-dev/env/overlays/kustomization.yaml",
	}
	sort.Strings(want)

	if diff := cmp.Diff(want, filesFromResources(resources)); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestBuildEnvironmentFilesWithNoCICDEnv(t *testing.T) {
	var appFs = ioutils.NewMemoryFilesystem()
	m := buildManifest()

	files, err := Build(appFs, m, "pipelines", AppsToEnvironments)
	if err != nil {
		t.Fatal(err)
	}

	want := res.Resources{
		"environments/test-dev/apps/my-app-1/base/kustomization.yaml": &res.Kustomization{
			Bases: []string{
				"../services/service-http",
				"../services/service-metrics",
			},
		},
		"environments/test-dev/apps/my-app-1/kustomization.yaml": &res.Kustomization{
			Bases: []string{"overlays"},
			CommonLabels: map[string]string{
				vcsSourceLabel: "example/example",
			},
		},
		"environments/test-dev/apps/my-app-1/overlays/kustomization.yaml":                          &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/env/base/test-dev-environment.yaml":                                 namespaces.Create("test-dev", testGitOpsRepoURL),
		"environments/test-dev/env/base/kustomization.yaml":                                        &res.Kustomization{Resources: []string{"test-dev-environment.yaml"}},
		"environments/test-dev/env/overlays/kustomization.yaml":                                    &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/apps/my-app-1/services/service-http/kustomization.yaml":             &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/apps/my-app-1/services/service-http/base/kustomization.yaml":        &res.Kustomization{Bases: []string{"./config"}},
		"environments/test-dev/apps/my-app-1/services/service-http/overlays/kustomization.yaml":    &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/apps/my-app-1/services/service-metrics/kustomization.yaml":          &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/apps/my-app-1/services/service-metrics/base/kustomization.yaml":     &res.Kustomization{Bases: []string{"./config"}},
		"environments/test-dev/apps/my-app-1/services/service-metrics/overlays/kustomization.yaml": &res.Kustomization{Bases: []string{"../base"}},
	}

	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestListFiles(t *testing.T) {
	var appFs = ioutils.NewMemoryFilesystem()
	var envPath = "environments/test-dev"
	var basePath = "environments/test-dev/base"
	err := appFs.MkdirAll(basePath, 0755)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		envPath = "environments\\test-dev"
		basePath = "environments\\test-dev\\base"
		mustWriteFile(t, appFs, "environments\\test-dev\\base\\01-namespaces\\cicd-environment.yaml", []byte(`this is a file`), 0644)
		mustWriteFile(t, appFs, "environments\\test-dev\\base\\03-tasks\\deploy-from-source-task.yaml", []byte(`this is a file`), 0644)
		mustWriteFile(t, appFs, "environments\\test-dev\\base\\04-pipelines\\app-ci-pipeline.yaml", []byte(`this is a file`), 0644)
		mustWriteFile(t, appFs, "environments\\test-dev\\base\\08-routes\\gitops-webhook-event-listener.yaml", []byte(`this is a file`), 0644)
	} else {
		mustWriteFile(t, appFs, "environments/test-dev/base/01-namespaces/cicd-environment.yaml", []byte(`this is a file`), 0644)
		mustWriteFile(t, appFs, "environments/test-dev/base/03-tasks/deploy-from-source-task.yaml", []byte(`this is a file`), 0644)
		mustWriteFile(t, appFs, "environments/test-dev/base/04-pipelines/app-ci-pipeline.yaml", []byte(`this is a file`), 0644)
		mustWriteFile(t, appFs, "environments/test-dev/base/08-routes/gitops-webhook-event-listener.yaml", []byte(`this is a file`), 0644)
	}
	exists, err := appFs.DirExists(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("Folder does not exist")
	}
	// This will strip the prefix base from the list of files
	listFiles, err := ListFiles(appFs, basePath)
	want := StringSet{
		"01-namespaces/cicd-environment.yaml":          true,
		"03-tasks/deploy-from-source-task.yaml":        true,
		"04-pipelines/app-ci-pipeline.yaml":            true,
		"08-routes/gitops-webhook-event-listener.yaml": true,
	}
	if diff := cmp.Diff(want, listFiles); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func filesFromResources(r res.Resources) []string {
	names := []string{}
	for k := range r {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func buildManifestWithCICD() *config.Manifest {
	m := buildManifest()
	m.Config = &config.Config{
		Pipelines: &config.PipelinesConfig{
			Name: "cicd",
		},
	}
	return m
}

func buildManifest() *config.Manifest {
	return &config.Manifest{
		GitOpsURL:    testGitOpsRepoURL,
		Environments: createEnvironment(),
	}
}

func createEnvironment() []*config.Environment {
	return []*config.Environment{
		{
			Name: "test-dev",
			Apps: []*config.Application{
				{
					Name: "my-app-1",
					Services: []*config.Service{
						{
							Name:      "service-http",
							SourceURL: "https://github.com/myproject/myservice.git",
						},
						{
							Name: "service-metrics",
						},
					},
				},
			},
		},
	}
}
