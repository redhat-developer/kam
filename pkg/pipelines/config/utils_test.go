package config

import (
	"testing"

	"github.com/chetan-rns/gitops-cli/pkg/pipelines/ioutils"
	"github.com/chetan-rns/gitops-cli/pkg/pipelines/yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/jenkins-x/go-scm/scm/factory"
)

func TestLoadManifestUpdatesDrivers(t *testing.T) {
	d, err := factory.DefaultIdentifier.Identify("example.com")
	if err == nil {
		t.Fatalf("successfully identified an unknown host as %q", d)
	}

	fs := ioutils.NewMemoryFilesystem()
	c := &Manifest{
		Config: &Config{
			Git: &GitConfig{
				Drivers: map[string]string{
					"example.com": "github",
				},
			},
		},
	}
	_, err = yaml.WriteResources(fs, "/manifest", map[string]interface{}{
		"pipelines.yaml": c,
	})
	if err != nil {
		t.Fatal(err)
	}

	m, err := LoadManifest(fs, "/manifest")
	if err != nil {
		t.Fatal("Failed to load manifest")
	}
	if diff := cmp.Diff(c, m); diff != "" {
		t.Fatalf("diff in loading manifest:\n%s", diff)
	}

	d, err = factory.DefaultIdentifier.Identify("example.com")
	if err != nil {
		t.Fatal("failed to identify driver after loading from manifest")
	}
	if d != "github" {
		t.Fatalf("incorrectly identified driver, got %q, want %q", d, "github")
	}
}
