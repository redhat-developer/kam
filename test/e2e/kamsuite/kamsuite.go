package kamsuite

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v10"
)

var (
	gitopsrepodir string
	originaldir   string
)

// FeatureContext defines godog.Suite steps for the test suite.
func FeatureContext(s *godog.Suite) {

	// KAM related steps
	s.Step(`^create gitops temporary directory$`,
		GitopsDir)
	s.Step(`^go to the gitops temporary directory$`,
		GoToGitopsDirPath)

	s.BeforeSuite(func() {
		fmt.Println("Before suite")
		if !envVariableCheck() {
			os.Exit(1)
		}
	})

	s.AfterSuite(func() {
		fmt.Println("After suite")
		deleteStep1 := "alias set delete 'api -X DELETE \"repos/$1\"'"
		deleteStep2 := "alias repo-delete kam-bot/" + os.Getenv("GITOPS_REPO_URL")
		if !executeGhCommad(deleteStep1) {
			os.Exit(1)
		}
		if !executeGhCommad(deleteStep2) {
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

func executeGhCommad(arg string) bool {
	ghExecPath, err := exec.LookPath("gh")
	if err != nil {
		fmt.Println("Error is ", err)
		return false
	}
	cmdDeleteRepo := &exec.Cmd{
		Path:   ghExecPath,
		Args:   []string{ghExecPath, arg},
		Stderr: os.Stderr,
	}
	if cmdDeleteRepo.Stderr != nil {
		fmt.Println("Error is ", cmdDeleteRepo.Stderr)
		return false
	}
	return true
}

// GitopsDir creates a temporary gitops dir
func GitopsDir() (string, error) {
	var err error
	gitopsrepodir, err = ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	return gitopsrepodir, nil
}

// WorkingDirPath gets the working dir
func WorkingDirPath() (string, error) {
	var err error
	originaldir, err = os.Getwd()
	if err != nil {
		return "", err
	}
	return originaldir, nil
}

// GoToGitopsDirPath change the working dir
func GoToGitopsDirPath() error {
	err := os.Chdir(gitopsrepodir)
	if err != nil {
		return err
	}
	return nil
}

// GoToKamDirPath change the working dir
func GoToKamDirPath() error {
	err := os.Chdir(originaldir)
	if err != nil {
		return err
	}
	return nil
}
