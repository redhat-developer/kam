package pipelines

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
)

// check if the file exists or not
// TODO: This should be renamed for clarity.
func CheckFileExists(fs afero.Fs, dockerConfigJSONFilename string) (string, error) {
	if dockerConfigJSONFilename == "" {
		return "", errors.New("failed to generate path to file: must provide --dockerconfigjson parameter")
	}
	authJSONPath, err := homedir.Expand(dockerConfigJSONFilename)
	if err != nil {
		return "", fmt.Errorf("failed to generate path to file: %v", err)
	}
	_, err = fs.Stat(authJSONPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist at the file-path passed to access the dockercfgjson, kindly enter a valid file path")
	}
	if err != nil {
		return "", fmt.Errorf("failed to read Docker config %q : %s", authJSONPath, err)
	}
	return authJSONPath, nil
}

func repoURL(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("failed to parse %q: %w", u, err)
	}
	parsed.Path = ""
	parsed.User = nil
	return parsed.String(), nil
}
