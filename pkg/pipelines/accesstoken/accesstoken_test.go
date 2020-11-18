package accesstoken

import (
	"fmt"
	"os"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestSetAccessToken(t *testing.T) {
	keyring.MockInit()
	optionTests := []struct {
		name          string
		gitRepo       string
		imagerepo     string
		gitToken      string
		expectedKey   string
		expectedToken string
	}{
		{"set the github access token in keyring", "https://github.com/example/service.git", "registry/username/repo", "abc123", "github.com", "abc123"},
		{"overwrite github access token in keyring with same secret", "https://github.com/example/service.git", "registry/username/repo", "xyz123", "github.com", "xyz123"},
		{"overwrite github access token in keyring with diffrent secret", "https://github.com/example/service.git", "registry/username/repo", "abc123", "github.com", "abc123"},
		{"set the gitlab access Token in keyring", "https://gitlab.com/example/service.git", "registry/username/repo", "test123", "gitlab.com", "test123"},
		{"overwrite gitlab access token in keyring with same secret", "https://gitlab.com/example/service.git", "registry/username/repo", "test345", "gitlab.com", "test345"},
		{"overwrite gitlab access token in keyring with same secret", "https://gitlab.com/example/service.git", "registry/username/repo", "abc123", "gitlab.com", "abc123"},
	}

	for _, tt := range optionTests {
		err := SetAccessToken(tt.gitRepo, tt.gitToken)
		if err != nil {
			t.Errorf("checkGitAccessToken() mode failed with error: %v", err)
		}
		gitopsToken, _ := keyring.Get(KeyringServiceName, tt.expectedKey)
		if tt.expectedToken != gitopsToken {
			t.Errorf("TestKeyRingFlagSet() Failed since expected token %v did not match %v", tt.expectedToken, gitopsToken)
		}
	}
}

func TestGetAccessToken(t *testing.T) {
	keyring.MockInit()
	optionTests := []struct {
		name          string
		serviceRepo   string
		gitToken      string
		envVarPresent bool
		keyRing       bool
		expectedToken string
	}{
		{"with token in keyring present(github)", "https://github.com/example/service.git", "abc123", true, false, "abc123"},
		{"with token in environment variable(github)", "https://github.com/example/service.git", "xyz123", false, true, "xyz123"},
		{"with token in environment variable(github) and keyring", "https://github.com/example/service.git", "xyz123", true, true, "xyz123"},
		{"with token in keyring present(gitlab)", "https://gitlab.com/example/service.git", "xyz123", true, false, "xyz123"},
		{"with token in environment variable(gitlab)", "https://gitlab.com/example/service.git", "xyz123", false, true, "xyz123"},
		{"with token in environment variable(gitlab) and keyring", "https://gitlab.com/example/service.git", "xyz123", false, true, "xyz123"},
	}

	for i, tt := range optionTests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			hostName, err := HostFromURL(tt.serviceRepo)
			if err != nil {
				t.Error("Failed to get host name from access token")
			}
			envVar := GetEnvVarName(hostName)
			defer os.Unsetenv(envVar)
			if tt.envVarPresent {
				err := os.Setenv(envVar, tt.expectedToken)
				if err != nil {
					t.Errorf("Error in setting the environment variable")
				}
			}
			if tt.keyRing {
				err := keyring.Set("kam", hostName, tt.expectedToken)
				if err != nil {
					t.Error(err)
				}
			}
			secret, err := GetAccessToken(tt.serviceRepo)
			if err != nil {
				t.Errorf("checkGitAccessToken() mode failed with error: %v", err)
			}
			if tt.expectedToken != secret {
				t.Fatalf("TestKeyRingFlagNotSet() Failed since expected token %v did not match %v", tt.expectedToken, secret)
			}
		})
	}
}
