package kamsuite

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

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

		ghLoginCommand := []string{"auth", "login", "--with-token"}
		if !executeGhLoginCommand(ghLoginCommand) {
			os.Exit(1)
		}
		if err := executeGhCreateRepo(os.Getenv("GITOPS_REPO_URL")); err != nil {
			log.Printf("failed to create repo: %s", err)
			os.Exit(1)
		}

	})

	s.AfterSuite(func() {
		fmt.Println("After suite")
		deleteGhRepoStep1 := []string{"alias", "set", "repo-delete", `api -X DELETE "repos/$1"`}
		deleteGhRepoStep2 := []string{"repo-delete", strings.Split(strings.Split(os.Getenv("GITOPS_REPO_URL"), "github.com/")[1], ".")[0]}
		ok, _ := executeGhRepoDeleteCommand(deleteGhRepoStep1)
		if !ok {
			os.Exit(1)
		}
		ok, errMessage := executeGhRepoDeleteCommand(deleteGhRepoStep2)
		if !ok {
			fmt.Println(errMessage)
		}
	})

	s.BeforeFeature(func(this *messages.GherkinDocument) {
		fmt.Println("Before feature")
	})

	s.AfterFeature(func(this *messages.GherkinDocument) {
		fmt.Println("After feature")
	})
}

func envVariableCheck() bool {
	envVars := []string{"SERVICE_REPO_URL", "GITOPS_REPO_URL", "IMAGE_REPO", "DOCKERCONFIGJSON_PATH", "GITHUB_TOKEN", "KAM_GITHUB_TOKEN_FILE"}
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
			os.Setenv("SERVICE_REPO_URL", "https://github.com/kam-bot/taxi")
			os.Setenv("GITOPS_REPO_URL", "https://github.com/kam-bot/taxi-"+os.Getenv("PRNO"))
			os.Setenv("IMAGE_REPO", "quay.io/kam-bot/taxi")
			os.Setenv("DOCKERCONFIGJSON_PATH", os.Getenv("KAM_QUAY_DOCKER_CONF_SECRET_FILE"))
		} else {
			fmt.Printf("You cannot run e2e test locally against OpenShift CI\n")
			return false
		}
		return true
	}
	return true
}

func executeGhLoginCommand(arg []string) bool {
	var stderr bytes.Buffer
	f, err := os.Open(os.Getenv("KAM_GITHUB_TOKEN_FILE"))
	if err != nil {
		fmt.Println("Error is : ", err)
		return false
	}
	cmd := exec.Command("gh", arg...)
	cmd.Stdin = bufio.NewReader(f)
	fmt.Println("gh command is : ", cmd.Args)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return false
	}
	return true
}

func executeGhRepoDeleteCommand(arg []string) (bool, string) {
	var stderr bytes.Buffer
	cmd := exec.Command("gh", arg...)
	fmt.Println("gh command is : ", cmd.Args)
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return false, stderr.String()
	}
	return true, stderr.String()
}

func executeGhCreateRepo(repo string) error {
	var stderr bytes.Buffer
	cmd := exec.Command("gh", "repo", "create", strings.TrimSuffix(repo, ".git"), "--private", "--confirm")
	fmt.Println("gh command is : ", cmd.Args)
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("gh failed creation %q: %s", repo, err)
	}
	return nil
}
