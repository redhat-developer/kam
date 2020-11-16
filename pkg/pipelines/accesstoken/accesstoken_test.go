package accesstoken

import (
	"fmt"
	"os"
	"testing"

	"github.com/zalando/go-keyring"
)

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
			hostName, err := HostFromURL(tt.gitRepo)
			if err != nil {
				t.Error("Failed to get host name from access token")
			}
			envVar := GetEnvVarName(hostName)
			defer os.Unsetenv(envVar)
			if tt.envVarPresent {
				err := os.Setenv(envVar, "abc123")
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
			token, _ := GetAccessToken(tt.gitRepo)

			if token != tt.expectedToken {
				t.Errorf("%v : GetAcessToken returned %v, expected %v", tt.name, token, tt.expectedToken)
			}
		})
	}
}
