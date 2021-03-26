package kamsuite

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v10"
)

// FeatureContext defines godog.Suite steps for the test suite.
func FeatureContext(s *godog.Suite) {

	// KAM related steps
	s.Step(`^directory "([^"]*)" should exist$`,
		DirectoryShouldExist)

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

	s.BeforeScenario(func(this *messages.Pickle) {
		fmt.Println("Before scenario")
	})

	s.AfterScenario(func(*messages.Pickle, error) {
		// Checking it for local test
		_, ci := os.LookupEnv("CI")
		if !ci {
			deleteGithubRepoStep1 := []string{"alias", "set", "repo-delete", `api -X DELETE "repos/$1"`}
			deleteGithubRepoStep2 := []string{"repo-delete", strings.Split(strings.Split(os.Getenv("GITOPS_REPO_URL"), ".com/")[1], ".")[0]}
			deleteGitlabRepoStep := []string{"repo", "delete", strings.Split(strings.Split(os.Getenv("GITOPS_REPO_URL"), ".com/")[1], ".")[0], "-y"}
			ok, _ := executeGithubRepoDeleteCommad(deleteGithubRepoStep1)
			if !ok {
				os.Exit(1)
			}
			ok, errMessage := executeGithubRepoDeleteCommad(deleteGithubRepoStep2)
			if !ok {
				fmt.Println(errMessage)
			}
			ok, errMessage = executeGitlabRepoDeleteCommad(deleteGitlabRepoStep)
			if !ok {
				fmt.Println(errMessage)
			}
		}
	})

}

func envVariableCheck() bool {
	envVars := []string{"SERVICE_REPO_URL", "GITOPS_REPO_URL", "IMAGE_REPO", "DOCKERCONFIGJSON_PATH", "GIT_ACCESS_TOKEN"}
	val, ok := os.LookupEnv("CI")
	if !ok {
		for _, envVar := range envVars {
			_, ok := os.LookupEnv(envVar)
			if !ok {
				fmt.Printf("%s is not set\n", envVar)
				return false
			}
		}
		if strings.Contains(os.Getenv("GITOPS_REPO_URL"), "github") {
			os.Setenv("GITHUB_TOKEN", os.Getenv("GIT_ACCESS_TOKEN"))
		} else {
			os.Setenv("GITLAB_TOKEN", os.Getenv("GIT_ACCESS_TOKEN"))
		}
	} else {
		if val == "prow" {
			fmt.Printf("Running e2e test in OpenShift CI\n")
			os.Setenv("SERVICE_REPO_URL", "https://github.com/kam-bot/taxi")
			os.Setenv("GITOPS_REPO_URL", "https://github.com/kam-bot/taxi-"+os.Getenv("PRNO"))
			os.Setenv("IMAGE_REPO", "quay.io/kam-bot/taxi")
			os.Setenv("DOCKERCONFIGJSON_PATH", os.Getenv("KAM_QUAY_DOCKER_CONF_SECRET_FILE"))
			os.Setenv("GIT_ACCESS_TOKEN", os.Getenv("GITHUB_TOKEN"))
		} else {
			fmt.Printf("You cannot run e2e test locally against OpenShift CI\n")
			return false
		}
		return true
	}
	return true
}

func executeGithubRepoDeleteCommad(arg []string) (bool, string) {
	var stderr bytes.Buffer
	cmd := exec.Command("gh", arg...)
	fmt.Println("github command is : ", cmd.Args)
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return false, stderr.String()
	}
	return true, stderr.String()
}

func executeGitlabRepoDeleteCommad(arg []string) (bool, string) {
	var stderr bytes.Buffer
	cmd := exec.Command("glab", arg...)
	fmt.Println("gitlab command is : ", cmd.Args)
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return false, stderr.String()
	}
	return true, stderr.String()
}
