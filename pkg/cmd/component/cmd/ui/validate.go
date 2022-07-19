package ui

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/AlecAivazis/survey.v1/terminal"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/klog"
)

const minSecretLen = 16

func makeURLValidatorCheck() survey.Validator {
	return func(input interface{}) error {
		return validateURL(input)
	}
}

//This function validates the Target Port
func ValidateTargetPort(number int) error {
	if number < 1025 || number > 65536 {
		return fmt.Errorf("%v is not valid", number)
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

func handleError(err error) {
	if err == nil {
		return
	}
	if err == terminal.InterruptErr {
		os.Exit(1)
	}
	klog.V(4).Infof("Encountered an error processing prompt: %v", err)
}

func makeSecretValidator() survey.Validator {
	return func(input interface{}) error {
		return validateSecretLength(input)
	}
}

func validateSecretLength(input interface{}) error {
	if s, ok := input.(string); ok {
		err := checkSecretLength(s)
		if err {
			return fmt.Errorf("the length of the secret must be at least %d characters", minSecretLen)
		}
		return nil
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
