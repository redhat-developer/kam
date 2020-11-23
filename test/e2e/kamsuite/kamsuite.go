package kamsuite

import (
	"fmt"
	"os"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v10"
)

// FeatureContext defines godog.Suite steps for the test suite.
func FeatureContext(s *godog.Suite) {

	// KAM related steps

	s.BeforeSuite(func() {
		fmt.Println("Before suite")
		if !envVariableCheck() {
			os.Exit(1)
		}
	})

	s.AfterSuite(func() {
		fmt.Println("After suite")
	})

	s.BeforeFeature(func(this *messages.GherkinDocument) {
		fmt.Println("Before feature")
	})

	s.AfterFeature(func(this *messages.GherkinDocument) {
		fmt.Println("After feature")
	})
}

func envVariableCheck() bool {
	envVars := []string{"SERVICE_REPO_URL", "GITOPS_REPO_URL", "IMAGE_REPO", "DOCKERCONFIGJSON_PATH", "GITHUB_TOKEN"}
	val, ok := os.LookupEnv("CI")
	if !ok {
		for _, envVar := range envVars {
			_, ok := os.LookupEnv(envVar)
			if !ok {
				fmt.Printf("%s is not set\n", envVar)
				return false
			}
		}
	} else {
		if val == "prow" {
			fmt.Printf("Running e2e test in OpenShift CI\n")
			os.Setenv("SERVICE_REPO_URL", "https://github.com/rhd-gitops-example/taxi")
			os.Setenv("GITOPS_REPO_URL", "https://github.com/kam-bot/taxi-"+os.Getenv("PRNO"))
			os.Setenv("IMAGE_REPO", "quay.io/kam-bot/taxi")
			os.Setenv("DOCKERCONFIGJSON_PATH", os.Getenv("KAM_QUAY_DOCKER_CONF_SECRET"))
			os.Getenv("GITHUB_TOKEN")
		} else {
			fmt.Printf("You cannot run e2e test locally against OpenShift CI\n")
			return false
		}
		return true
	}
	return true
}
