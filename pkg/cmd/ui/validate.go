package ui

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/odo/pkg/log"
	"github.com/redhat-developer/kam/pkg/cmd/utility"
	"github.com/redhat-developer/kam/pkg/pipelines/accesstoken"
	"github.com/redhat-developer/kam/pkg/pipelines/git"
	"github.com/redhat-developer/kam/pkg/pipelines/ioutils"
	"github.com/redhat-developer/kam/pkg/pipelines/namespaces"
	"github.com/zalando/go-keyring"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/AlecAivazis/survey.v1/terminal"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/klog"
)

const minSecretLen = 16

func makePrefixValidator() survey.Validator {
	return func(input interface{}) error {
		return validatePrefix(input)
	}
}

func makeSecretValidator() survey.Validator {
	return func(input interface{}) error {
		return validateSecretLength(input)
	}
}

func makeOverWriteValidator(path string) survey.Validator {
	return func(input interface{}) error {
		return validateOverwriteOption(input, path)
	}
}

func makeSealedSecretNamespaceValidator(sealedSecretService *types.NamespacedName) survey.Validator {
	return func(input interface{}) error {
		return validateSealedSecretNamespace(input, sealedSecretService)
	}
}

func makeSealedSecretServiceValidator(sealedSecretService *types.NamespacedName) survey.Validator {
	return func(input interface{}) error {
		return validateSealedSecretService(input, sealedSecretService)
	}
}

// SetAccessToken validates the access  token and the sets/retrieves the value from the keyring based on the situation.
func SetAccessToken(io *RepoParams) error {
	if io.GitHostAccessToken != "" {
		err := ValidateAccessToken(io)
		if err != nil {
			return err
		}
	}
	if io.KeyringServiceRequired {
		err := accesstoken.SetAccessToken(io.RepoInfo.RepoURL, io.GitHostAccessToken)
		if err != nil {
			return err
		}
	}
	if io.GitHostAccessToken == "" {
		secret, err := accesstoken.GetAccessToken(io.RepoInfo.RepoURL)
		if err != nil {
			return fmt.Errorf("unable to use access-token from keyring/env-var: %v, please pass a valid token to --git-host-access-token", err)
		}
		io.GitHostAccessToken = secret
	}
	return nil
}

// ValidatePrefix checks the length of the prefix with the env crosses 63 chars or not
func validatePrefix(input interface{}) error {
	if s, ok := input.(string); ok {
		prefix := utility.MaybeCompletePrefix(s)
		s = prefix + "stage"
		if len(s) < 64 {
			err := ValidateName(s)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("The prefix %s, must be less than 58 characters", prefix)
		}
		return nil
	}
	return nil
}

// ValidateName will do validation of application & component names according to DNS (RFC 1123) rules
// Criteria for valid name in kubernetes: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/architecture/identifiers.md
func ValidateName(name string) error {

	errorList := validation.IsDNS1123Label(name)

	if len(errorList) != 0 {
		return fmt.Errorf("%s is not a valid name:  %s", name, strings.Join(errorList, " "))
	}

	return nil
}

func validateSecretLength(input interface{}) error {
	if s, ok := input.(string); ok {
		err := checkSecretLength(s)
		if err {
			return fmt.Errorf("The length of the secret must be at least %d characters", minSecretLen)
		}
		return nil
	}
	return nil
}

// validateOverwriteOption(  validates the URL
func validateOverwriteOption(input interface{}, path string) error {
	if s, ok := input.(string); ok {
		if s == "no" {
			exists, _ := ioutils.IsExisting(ioutils.NewFilesystem(), filepath.Join(path, "pipelines.yaml"))
			if exists {
				EnterOutputPath()
			}
		}
		return nil
	}
	return nil

}

// ValidateAccessToken validates if the access token is correct for a particular service repo
func ValidateAccessToken(io *RepoParams) error {
	repo, err := git.NewRepository(io.RepoInfo.RepoURL, io.GitHostAccessToken)
	if err != nil {
		return err
	}
	parsedURL, err := url.Parse(io.RepoInfo.RepoURL)
	if err != nil {
		return fmt.Errorf("failed to parse the provided URL %q: %w", io.RepoInfo.RepoURL, err)
	}
	repoName, err := git.GetRepoName(parsedURL)
	if err != nil {
		return fmt.Errorf("failed to get the repository name from %q: %w", io.RepoInfo.RepoURL, err)
	}
	_, res, err := repo.Client.Repositories.Find(context.Background(), repoName)
	if err != nil && res.Status == 401 {
		return apierrors.NewForbidden(schema.GroupResource{}, "", fmt.Errorf("Invalid access token, unable to authenticate client for repo: %s", io.RepoInfo.RepoURL))
	}
	if err != nil && res.Status == 404 {
		log.Warningf("Note: The git repo %v cannot be found, usually occurs when the repository does not exist", io.RepoInfo.RepoURL)
		io.RepoInfo.GitRepoValid = false
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}

// validateSealedSecretService validates to see if the sealed secret service is present in the correct namespace.
func validateSealedSecretNamespace(input interface{}, sealedSecretService *types.NamespacedName) error {
	if s, ok := input.(string); ok {
		clientSet, err := namespaces.GetClientSet()
		if err != nil {
			return err
		}
		exists, _ := namespaces.Exists(clientSet, s)
		if !exists {
			return fmt.Errorf("The namespace %s is not found on the cluster", s)
		}
		sealedSecretService.Namespace = s
	}
	return nil
}

// validateSealedSecretService validates to see if the sealed secret service is present in the correct namespace.
func validateSealedSecretService(input interface{}, sealedSecretService *types.NamespacedName) error {
	if s, ok := input.(string); ok {
		client, err := utility.NewClient()
		if err != nil {
			return err
		}
		sealedSecretService.Name = s
		return client.CheckIfSealedSecretsExists(*sealedSecretService)
	}
	return nil
}

func validateRepoTokenCreds(repoParams *RepoParams) error {
	secret, err := accesstoken.GetAccessToken(repoParams.RepoInfo.RepoURL)
	if err != nil && err != keyring.ErrNotFound {
		return err
	}
	if secret == "" {
		token, err := EnterGitHostAccessToken(repoParams.RepoInfo.RepoURL)
		repoParams.GitHostAccessToken = token
		repoParams.KeyringServiceRequired = UseKeyringRingSvc()
		err = SetAccessToken(repoParams)
		if err != nil {
			return err
		}
	} else {
		err := ValidateAccessToken(repoParams)
		if err != nil {
			return err
		}
		repoParams.GitHostAccessToken = secret
	}
	return nil
}
func checkSecretLength(secret string) bool {
	if secret != "" {
		if len(secret) < minSecretLen {
			return true
		}
	}
	return false
}

// handleError handles UI-related errors, in particular useful to gracefully handle ctrl-c interrupts gracefully
func handleError(err error) {
	if err != nil {
		if err == terminal.InterruptErr {
			os.Exit(1)
		} else {
			klog.V(4).Infof("Encountered an error processing prompt: %v", err)
		}
	}
}
