package webhook

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redhat-developer/kam/pkg/pipelines/config"
	"github.com/zalando/go-keyring"
)

func TestBuildURL(t *testing.T) {
	testcases := []struct {
		host   string
		hasTLS bool
		want   string
	}{
		{
			host:   "test.example.com",
			hasTLS: false,
			want:   "http://test.example.com",
		},
		{
			host:   "test.example.com",
			hasTLS: true,
			want:   "https://test.example.com",
		},
	}

	for i, tt := range testcases {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			got := buildURL(tt.host, tt.hasTLS)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("result mismatch got\n%s", diff)
			}
		})
	}
}
func TestGetGitRepoURL(t *testing.T) {
	testcases := []struct {
		manifest    *config.Manifest
		isCICD      bool
		serviceName *QualifiedServiceName
		want        string
	}{
		{
			manifest: &config.Manifest{
				GitOpsURL: "https://github.com/foo/bar.git",
			},
			isCICD: true,
			want:   "https://github.com/foo/bar.git",
		},
		{
			manifest: &config.Manifest{},
			want:     "",
		},
		{
			manifest: &config.Manifest{
				GitOpsURL: "https://github.com/foo/bar.git",
				Environments: []*config.Environment{
					{

						Name: "notmyenv",
						Apps: []*config.Application{
							{
								Name: "notmyapp",
								Services: []*config.Service{
									{
										Name:      "notmyservice",
										SourceURL: "https://not/mine",
									},
								},
							},
						},
					},
					{
						Name: "myenv",
						Apps: []*config.Application{
							{
								Name: "notmyapp",
								Services: []*config.Service{
									{
										Name:      "notmyservice",
										SourceURL: "https://not/mine",
									},
								},
							},
							{
								Name: "myapp",
								Services: []*config.Service{
									{
										Name:      "notmyservice",
										SourceURL: "https://not/mine",
									},

									{
										Name:      "notmyserviceagain",
										SourceURL: "https://not/mine",
									},
								},
							},
							{
								Name: "notmyapp2",
								Services: []*config.Service{

									{
										Name:      "notmyserviceagain",
										SourceURL: "https://not/mine",
									},
								},
							},
						},
					},
				},
			},
			isCICD:      false,
			serviceName: &QualifiedServiceName{EnvironmentName: "myenv", ServiceName: "notmyservice"},
			want:        "https://not/mine",
		},
	}

	for i, tt := range testcases {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			got := getRepoURL(tt.manifest, tt.isCICD, tt.serviceName)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("result mismatch got\n%s", diff)
			}
		})
	}
}

func TestGetAccessToken(t *testing.T) {
	keyring.MockInit()
	optionTests := []struct {
		name             string
		gitRepo          string
		envVarPresent    bool
		tokenRingPresent bool
		hostName         string
		expectedToken    string
	}{
		{"Github Token not present in keyring/env-var", "https://github.com/example/test.git", false, false, "github.com", ""},
		{"Github Token present in env-var", "https://github.com/example/test.git", true, false, "github.com", "abc123"},
		{"Github Token present in keyring", "https://github.com/example/test.git", false, true, "github.com", "xyz123"},
		{"Github Token defaults to keyring although env-var present", "https://github.com/example/test.git", true, true, "github.com", "xyz123"},
		{"Gitlab Token not present in keyring/env-var", "https://gitlab.com/example/test.git", false, false, "gitlab.com", ""},
		{"Gitlab Token present in env-var", "https://gitlab.com/example/test.git", true, false, "gitlab.com", "abc123"},
		{"Gitlab Token present in keyring", "https://gitlab.com/example/test.git", false, true, "gitlab.com", "xyz123"},
		{"Gitlab Token defaults to keyring although env-var present", "https://gitlab.com/example/test.git", true, true, "gitlab.com", "xyz123"},
	}

	for i, tt := range optionTests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			defer os.Unsetenv("TESTGITTOKEN")
			if tt.envVarPresent {
				err := os.Setenv("TESTGITTOKEN", "abc123")
				if err != nil {
					t.Errorf("Error in setting the environment variable")
				}
			}
			if tt.tokenRingPresent {
				err := keyring.Set(KeyringServiceName, tt.hostName, "xyz123")
				if err != nil {
					t.Error(err)
				}
			}
			token, _ := getAccessToken(tt.gitRepo)

			if token != tt.expectedToken {
				t.Errorf("%v : getAcessToken returned %v, expected %v", tt.name, token, tt.expectedToken)
			}
		})
	}
}
