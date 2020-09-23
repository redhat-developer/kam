package config

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redhat-developer/gitops-cli/pkg/pipelines/ioutils"
	"github.com/redhat-developer/gitops-cli/pkg/pipelines/yaml"
)

func TestParse(t *testing.T) {
	parseTests := []struct {
		filename string
		want     *Manifest
	}{
		{"testdata/example1.yaml", &Manifest{
			Config: &Config{
				Pipelines: &PipelinesConfig{
					Name: "test-pipelines",
				},
				ArgoCD: &ArgoCDConfig{
					Namespace: "test-argocd",
				},
				Git: &GitConfig{
					Drivers: map[string]string{
						"test.example.com": "github",
					},
				},
			},
			Environments: []*Environment{
				{
					Name: "development",
					Pipelines: &Pipelines{
						Integration: &TemplateBinding{
							Template: "dev-ci-template",
							Bindings: []string{"dev-ci-binding"},
						},
					},
					Apps: []*Application{
						{
							Name: "my-app-1",
							Services: []*Service{
								{
									Name:      "service-http",
									SourceURL: "https://github.com/myproject/myservice.git",
								},
							},
						},
						{
							Name: "my-app-2",
							Services: []*Service{
								{Name: "service-redis"},
							},
						},
					},
				},
				{
					Name: "staging",
					Apps: []*Application{
						{Name: "my-app-1",
							ConfigRepo: &Repository{
								URL:            "https://github.com/testing/testing",
								TargetRevision: "master",
								Path:           "config",
							},
						},
					},
				},
				{
					Name: "production",
					Apps: []*Application{
						{
							Name: "my-app-1",
							Services: []*Service{
								{Name: "service-http"},
								{Name: "service-metrics"},
							},
						},
					},
				},
			},
		},
		},

		{"testdata/example2.yaml", &Manifest{
			Environments: []*Environment{
				{
					Name: "development",
					Apps: []*Application{
						{
							Name: "my-app-1",
							Services: []*Service{
								{
									Name:      "app-1-service-http",
									SourceURL: "https://github.com/myproject/myservice.git",
								},
								{Name: "app-1-service-metrics"},
							},
						},
					},
				},
				{
					Name: "tst-cicd",
				},
			},
		},
		},
		{"testdata/example-with-cluster.yaml", &Manifest{
			Environments: []*Environment{
				{
					Name:    "development",
					Cluster: "testing.cluster",
					Apps: []*Application{
						{Name: "my-app-1",
							Services: []*Service{
								{Name: "service-http",
									SourceURL: "https://github.com/myproject/myservice.git"},
							}},
					},
				},
			},
		},
		},
	}

	for _, tt := range parseTests {
		t.Run(fmt.Sprintf("parsing %s", tt.filename), func(rt *testing.T) {
			fs := ioutils.NewFilesystem()
			f, err := fs.Open(tt.filename)
			if err != nil {
				rt.Fatalf("failed to open %v: %s", tt.filename, err)
			}
			defer f.Close()

			got, err := Parse(f)
			if err != nil {
				rt.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				rt.Errorf("Parse(%s) failed diff\n%s", tt.filename, diff)
			}
		})
	}
}

func TestParsePipelinesFolder(t *testing.T) {

	want := &Manifest{
		Environments: []*Environment{
			{
				Name:    "development",
				Cluster: "testing.cluster",
				Apps: []*Application{
					{Name: "my-app-1",
						Services: []*Service{
							{Name: "service-http",
								SourceURL: "https://github.com/myproject/myservice.git"},
						}},
				},
			},
		},
	}

	fakeFs := ioutils.NewMemoryFilesystem()
	err := yaml.MarshalItemToFile(fakeFs, "gitops/pipelines.yaml", want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParsePipelinesFolder(fakeFs, "gitops")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("ParsePipelinesFolder() failed: %s", diff)
	}
}
