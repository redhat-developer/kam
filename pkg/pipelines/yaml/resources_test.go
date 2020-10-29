package yaml

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redhat-developer/kam/pkg/pipelines/namespaces"
	res "github.com/redhat-developer/kam/pkg/pipelines/resources"
	"github.com/redhat-developer/kam/test"
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
	os.Setenv(homeEnv, path)
	sampleYAML := namespaces.Create("test", "https://github.com/org/test")
	r := res.Resources{
		"test/myfile.yaml": sampleYAML,
	}

	tests := []struct {
		name   string
		path   string
		errMsg string
	}{
		{"Path with ~", "~/manifest", ""},
		{"Path without ~", filepath.Join(path, "manifest/gitops"), ""},
		{"Path without permission", "/", "failed to MkDirAll for /test/myfile.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := WriteResources(fs, tt.path, r)
			if !test.ErrorMatch(t, tt.errMsg, err) {
				t.Fatalf("error mismatch: got %v, want %v", err, tt.errMsg)
			}
			if tt.path[0] == '~' {
				tt.path = filepath.Join(path, strings.Split(tt.path, "~")[1])
			}
			if err == nil {
				assertResourceExists(t, filepath.Join(tt.path, "test/myfile.yaml"), sampleYAML)
			}
		})
	}
}

func makeTempDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := ioutil.TempDir(os.TempDir(), "manifest")
	test.AssertNoError(t, err)
	return dir, func() {
		err := os.RemoveAll(dir)
		test.AssertNoError(t, err)
	}
}

func assertResourceExists(t *testing.T, path string, resource interface{}) {
	t.Helper()
	want, err := yaml.Marshal(resource)
	test.AssertNoError(t, err)
	got, err := ioutil.ReadFile(path)
	test.AssertNoError(t, err)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("files not written to correct location: %s", diff)
	}
}
