package yaml

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redhat-developer/kam/pkg/pipelines/namespaces"
	res "github.com/redhat-developer/kam/pkg/pipelines/resources"
	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

func TestWriteResources(t *testing.T) {
	fs := afero.NewOsFs()
	homeEnv := "HOME"
	originalHome := os.Getenv(homeEnv)
	defer os.Setenv(homeEnv, originalHome)
	path, cleanup := makeTempDir(t)
	defer cleanup()
	os.Setenv("HOME", path)
	sampleYAML := namespaces.Create("test", "https://github.com/org/test")
	r := res.Resources{
		"test/myfile.yaml": sampleYAML,
	}

	_, err := WriteResources(fs, "~/manifest", r)
	assertNoError(t, err)

	want, err := yaml.Marshal(sampleYAML)
	assertNoError(t, err)
	got, err := ioutil.ReadFile(filepath.Join(path, "manifest/test/myfile.yaml"))
	assertNoError(t, err)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("files not written to correct location: %s", diff)
	}
}

func makeTempDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := ioutil.TempDir(os.TempDir(), "manifest")
	assertNoError(t, err)
	return dir, func() {
		err := os.RemoveAll(dir)
		assertNoError(t, err)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
