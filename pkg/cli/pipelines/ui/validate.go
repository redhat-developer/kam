package ui

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/chetan-rns/gitops-cli/pkg/cli/pipelines/utility"
	"github.com/chetan-rns/gitops-cli/pkg/cli/util/validation"
	"github.com/chetan-rns/gitops-cli/pkg/pipelines/git"
	"github.com/chetan-rns/gitops-cli/pkg/pipelines/ioutils"
	"github.com/chetan-rns/gitops-cli/pkg/pipelines/secrets"
	"gopkg.in/AlecAivazis/survey.v1"
	"k8s.io/apimachinery/pkg/types"
)

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

func makeSealedSecretsService(sealedSecretService *types.NamespacedName) survey.Validator {
	return func(input interface{}) error {
		return validateSealedSecretService(input, sealedSecretService)
	}
}

func makeAccessTokenCheck(serviceRepo string) survey.Validator {
	return func(input interface{}) error {
		return validateAccessToken(input, serviceRepo)
	}
}

// ValidatePrefix checks the length of the prefix with the env crosses 63 chars or not
func validatePrefix(input interface{}) error {
	if s, ok := input.(string); ok {
		prefix := utility.MaybeCompletePrefix(s)
		s = prefix + "stage"
		if len(s) < 64 {
			err := validation.ValidateName(s)
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

func validateSecretLength(input interface{}) error {
	if s, ok := input.(string); ok {
		err := CheckSecretLength(s)
		if err {
			return fmt.Errorf("The secret length should 16 or more ")
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

// validateAccessToken validates if the access token is correct for a particular service repo
func validateAccessToken(input interface{}, serviceRepo string) error {
	if s, ok := input.(string); ok {
		repo, _ := git.NewRepository(serviceRepo, s)
		parsedURL, err := url.Parse(serviceRepo)
		repoName, err := git.GetRepoName(parsedURL)
		_, _, err = repo.Client.Repositories.Find(context.Background(), repoName)
		if err != nil {
			return fmt.Errorf("The token passed is incorrect for repository %s", repoName)
		}
		return nil
	}
	return nil
}

// validateSealedSecretService validates to see if the sealed secret service is present in the correct namespace.
func validateSealedSecretService(input interface{}, sealedSecretService *types.NamespacedName) error {
	if s, ok := input.(string); ok {
		sealedSecretService.Name = s
		sealedSecretService.Namespace = EnterSealedSecretNamespace()
		_, err := secrets.GetClusterPublicKey(*sealedSecretService)
		if err != nil {
			if compareError(err, sealedSecretService.Name) {
				return fmt.Errorf("The given service %q is not installed in the right namespace %q", sealedSecretService.Name, sealedSecretService.Namespace)
			}
			return errors.New("sealed secrets could not be configured sucessfully")
		}
		return nil
	}
	return nil
}

func compareError(err error, sealedSecretService string) bool {
	createdError := fmt.Errorf("cannot fetch certificate: services \"%s\" not found", sealedSecretService)
	return err.Error() == createdError.Error()
}

// check if the length of secret is less than 16 chars
func CheckSecretLength(secret string) bool {
	if secret != "" {
		if len(secret) < 16 {
			return true
		}
	}
	return false
}
