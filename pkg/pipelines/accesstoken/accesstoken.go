package accesstoken

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/redhat-developer/kam/pkg/cmd/ui"
	"github.com/zalando/go-keyring"
)

// KeyringServiceName refers to service name used to set the accesstoken in the keyring
const KeyringServiceName = "kam"

// GetAccessToken returns the token from either that is stored in the keyring or the environment variable in this order.
func GetAccessToken(gitRepoURL string) (string, error) {
	hostName, err := HostFromURL(gitRepoURL)
	if err != nil {
		return "", err
	}
	accessToken, err := keyring.Get(KeyringServiceName, hostName)
	if err != nil && err != keyring.ErrNotFound {
		return "", err
	}
	if err != nil && err == keyring.ErrNotFound {
		envVarName := GetEnvVarName(hostName)
		accessToken = os.Getenv(envVarName)
		if accessToken == "" {
			return "", nil
		}
	}
	return accessToken, nil
}

//CheckGitAccessToken sets the access-token in the keyring if access-token is present, else it tries to retrieve it from the keyring/env-var
func CheckGitAccessToken(accessToken, repoURL string) (string, error) {
	if accessToken != "" {
		err := ui.ValidateAccessToken(accessToken, repoURL)
		if err != nil {
			return "", fmt.Errorf("Please enter a valid access token: %v", err)
		}
		err = SetSecret(repoURL, accessToken)
		if err == nil {
			return accessToken, nil
		}
		return "", err
	}
	secret, err := GetAccessToken(repoURL)
	if err != nil {
		return "", err
	}
	if secret == "" && err == nil {
		return "", errors.New("unable to retrieve the access token from the keyring/env-var: kindly pass the --git-host-access-token")
	}
	return secret, nil
}

// HostFromURL extracts the hostname from the url passed
func HostFromURL(s string) (string, error) {
	p, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	return strings.ToLower(p.Host), nil
}

//SetSecret sets the secret in the keyring
func SetSecret(repoURL, accessToken string) error {
	hostName, err := HostFromURL(repoURL)
	if err != nil {
		return err
	}
	secret, err := getSecret(hostName)
	if err != nil {
		return err
	}
	if accessToken != secret {
		err := keyring.Set(KeyringServiceName, hostName, accessToken)
		if err != nil {
			return fmt.Errorf("unable to set access token for repo %q using keyring: %w", repoURL, err)
		}
	}
	return nil
}

//GetEnvVarName contains the logic for the naming convention of the environment variable that contains the accesstoken
func GetEnvVarName(hostName string) string {
	FmtHostName := strings.ReplaceAll(hostName, ".", "_")
	envVarName := strings.ToUpper(FmtHostName) + "_TOKEN"
	return envVarName
}

func getSecret(hostName string) (string, error) {
	secret, err := keyring.Get(KeyringServiceName, hostName)
	if err != nil && err != keyring.ErrNotFound {
		return "", err
	}
	return secret, nil
}
