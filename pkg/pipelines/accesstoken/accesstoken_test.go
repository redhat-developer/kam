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
		{"Github Token not present in keyring/env-var", "https://githubtest.com/example/test.git", false, false, "github.com", ""},
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

func TestKeyRingFlagSet(t *testing.T) {
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
		{"overwrite github access token in keyring", "https://github.com/example/service.git", "registry/username/repo", "xyz123", "github.com", "xyz123"},
		{"set the gitlab access Token in keyring", "https://gitlab.com/example/service.git", "registry/username/repo", "test123", "gitlab.com", "test123"},
		{"overwrite gitlab access token in keyring", "https://gitlab.com/example/service.git", "registry/username/repo", "test345", "gitlab.com", "test345"},
	}

	for _, tt := range optionTests {
		err := SetSecret(tt.gitRepo, tt.gitToken)
		if err != nil {
			t.Errorf("checkGitAccessToken() mode failed with error: %v", err)
		}
		gitopsToken, _ := keyring.Get(KeyringServiceName, tt.expectedKey)
		if tt.expectedToken != gitopsToken {
			t.Errorf("TestKeyRingFlagSet() Failed since expected token %v did not match %v", tt.expectedToken, gitopsToken)
		}
	}
}

func TestKeyRingFlagNotSet(t *testing.T) {
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
		{"with token in keyring present(gitlab)", "https://gitlab.com/example/service.git", "xyz123", true, false, "xyz123"},
		{"with token in environment variable(github)", "https://gitlab.com/example/service.git", "xyz123", false, true, "xyz123"},
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
			secret, err := CheckGitAccessToken("", tt.serviceRepo)
			if err != nil {
				t.Errorf("checkGitAccessToken() mode failed with error: %v", err)
			}
			if tt.expectedToken != secret {
				t.Fatalf("TestKeyRingFlagNotSet() Failed since expected token %v did not match %v", tt.expectedToken, secret)
			}
		})
	}
}

func TestAccessToken(t *testing.T) {
	keyring.MockInit()
	cmdTests := []struct {
		desc      string
		testToken string
		testURL   string
		wantErr   string
	}{
		{"Access Token is incorrect",
			"test123",
			"https://github.com/user/repo.git",
			"Please enter a valid access token: The token passed is incorrect for repository user/repo",
		},
		{"Unable to retrieve token from keyring/env-var",
			"",
			"https://github.com/user/repo.git",
			"unable to retrieve the access token from the keyring/env-var: kindly pass the --git-host-access-token",
		},
	}

	for _, tt := range cmdTests {
		t.Run(tt.desc, func(t *testing.T) {
			_, err := CheckGitAccessToken(tt.testToken, tt.testURL)
			if err.Error() != tt.wantErr {
				t.Errorf("got %s, want %s", err, tt.wantErr)
			}
		})
	}
}
