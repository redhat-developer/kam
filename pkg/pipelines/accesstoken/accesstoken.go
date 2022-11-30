package accesstoken

import (
	"fmt"
	"net/url"
	"os"
	"strings"

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
	envVarName := GetEnvVarName(hostName)
	accessToken := os.Getenv(envVarName)
	if accessToken == "" {
		accessToken, err = keyring.Get(KeyringServiceName, hostName)
		if err != nil {
			return "", err
		}
	}
	return accessToken, nil
}

// HostFromURL extracts the hostname from the url passed
func HostFromURL(s string) (string, error) {
	p, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	return strings.ToLower(p.Host), nil
}

// SetSecret sets the secret in the keyring
func SetAccessToken(repoURL, accessToken string) error {
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

// GetEnvVarName contains the logic for the naming convention of the environment variable that contains the accesstoken
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
