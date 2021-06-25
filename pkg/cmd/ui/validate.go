package ui

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/redhat-developer/kam/pkg/cmd/utility"
	"github.com/redhat-developer/kam/pkg/pipelines/git"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/AlecAivazis/survey.v1/terminal"
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

func makeAccessTokenCheck(serviceRepo string) survey.Validator {
	return func(input interface{}) error {
		return ValidateAccessToken(input, serviceRepo)
	}
}

func makeURLValidatorCheck() survey.Validator {
	return func(input interface{}) error {
		return validateURL(input)
	}
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

// ValidateAccessToken validates if the access token is correct for a particular service repo
func ValidateAccessToken(input interface{}, serviceRepo string) error {
	if s, ok := input.(string); ok {
		repo, err := git.NewRepository(serviceRepo, s)
		if err != nil {
			return fmt.Errorf("%w. %s", err, "Check that the --private-repo-driver option is provided.")
		}
		parsedURL, err := url.Parse(serviceRepo)
		if err != nil {
			return fmt.Errorf("failed to parse the provided URL %q: %w", serviceRepo, err)
		}
		repoName, err := git.GetRepoName(parsedURL)
		if err != nil {
			return fmt.Errorf("failed to get the repository name from %q: %w", serviceRepo, err)
		}
		_, _, err = repo.Client.Repositories.Find(context.Background(), repoName)
		if err != nil {
			return fmt.Errorf("The token passed is incorrect for repository %s", repoName)
		}
		return nil
	}
	return nil
}

func validateURL(input interface{}) error {
	if u, ok := input.(string); ok {
		p, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("invalid URL, err: %v", err)
		}
		if p.Host == "" {
			return fmt.Errorf("could not identify host from %q", u)
		}
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
	if err == nil {
		return
	}
	if err == terminal.InterruptErr {
		os.Exit(1)
	}
	klog.V(4).Infof("Encountered an error processing prompt: %v", err)
}
