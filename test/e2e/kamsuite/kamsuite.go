package kamsuite

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

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
		val, ok := os.LookupEnv("CI")
		if ok && val == "prow" {
			err := os.MkdirAll(os.Getenv("HOME")+"/.ssh", 0755)
			if err != nil {
				fmt.Println(err.Error())
			}

			f, err := os.OpenFile(filepath.Join(os.Getenv("HOME"), ".ssh", "config"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			if _, err = f.Write([]byte("Host github.com\n\tStrictHostKeyChecking no\n")); err != nil {
				f.Close() // ignore error; Write error takes precedence
				log.Fatal(err)
			}

			content, err := ioutil.ReadFile(filepath.Join(os.Getenv("HOME"), ".ssh", "config"))

			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(string(content))

			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}
	})

	s.AfterSuite(func() {
		fmt.Println("After suite")
		deleteGhRepoStep1 := "alias set delete 'api -X DELETE \"repos/$1\"'"
		deleteGhRepoStep2 := "alias repo-delete kam-bot/" + os.Getenv("GITOPS_REPO_URL")
		if !executeGhCommad(deleteGhRepoStep1) || !executeGhCommad(deleteGhRepoStep2) {
			os.Exit(1)
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

func executeGhCommad(arg string) bool {
	cmd := exec.Command("gh", arg)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}
