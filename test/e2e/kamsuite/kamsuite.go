package kamsuite

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v10"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/redhat-developer/kam/pkg/pipelines/yaml"
	"k8s.io/apimachinery/pkg/util/wait"

	deployment "github.com/redhat-developer/kam/pkg/pipelines/deployment"
	"github.com/redhat-developer/kam/pkg/pipelines/ioutils"
	res "github.com/redhat-developer/kam/pkg/pipelines/resources"
)

// FeatureContext defines godog.Suite steps for the test suite.
func FeatureContext(s *godog.Suite) {

	// Check if directory exists
	s.Step(`^directory "([^"]*)" should exist$`,
		DirectoryShouldExist)

	// For creating repository with specified name
	s.Step(`^"([^"]*)" repository is created$`,
		createRepository)

	// For checking if the repository exists
	s.Step(`^"([^"]*)" repository exists$`,
		checkRepository)

	s.Step(`^wait for application "([^"]*)" to be in "([^"]*)" state$`,
		waitSync)

	// Adding sample kubernetes resource
	s.Step(`^add kubernetes resource to the service in new environment$`,
		addResource)

	s.Step(`^create and merge pull request after all the checks pass$`,
		createPR)

	s.BeforeSuite(func() {
		fmt.Println("Before suite")
		if !envVariableCheck() {
			os.Exit(1)
		}
		err := loginToArgoAPIServer()
		if err != nil {
			log.Fatal(err)
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

		// Clearing working directory before each scenario
		f, err := os.Open("./")
		if err == nil {
			err := os.RemoveAll("bootstrapresources")
			if err != nil {
				log.Fatal(err)
			}
			err2 := os.RemoveAll("secrets")
			if err2 != nil {
				log.Fatal(err2)
			}
			fmt.Println("Cleared working directory")
		}
		defer f.Close()
	})

	s.AfterScenario(func(*messages.Pickle, error) {
		fmt.Println("After scenario")
		err := deleteRepository(os.Getenv("GITOPS_REPO_URL"))
		if err != nil {
			fmt.Println(os.Getenv("GITOPS_REPO_URL"), ":", err)
		}
	})
}

type VCS struct {
	GITHUB GIT `json:"github"`
	GITLAB GIT `json:"gitlab"`
}

type GIT struct {
	SERVICE_REPO_URL      string `json:"service_repo_url"`
	GITOPS_REPO_URL       string `json:"gitops_repo_url"`
	IMAGE_REPO            string `json:"image_repo"`
	BUS_REPO_URL          string `json:"bus_repo_url"`
	DOCKERCONFIGJSON_PATH string `json:"dockerconfigjson_path"`
	GIT_ACCESS_TOKEN      string `json:"git_access_token"`
}

func envVariableCheck() bool {

	var stdout bytes.Buffer
	var vcs VCS

	cmd := exec.Command("pwd")
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

	jsonFile, err := os.Open(strings.Split(stdout.String(), "/test")[0] + "/variables.json")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Configuration file not set")
		return false
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &vcs)

	val, ok := os.LookupEnv("CI")
	if !ok {

		if os.Getenv("VCS") == "github" {
			os.Setenv("SERVICE_REPO_URL", vcs.GITHUB.SERVICE_REPO_URL)
			os.Setenv("GITOPS_REPO_URL", vcs.GITHUB.GITOPS_REPO_URL)
			os.Setenv("IMAGE_REPO", vcs.GITHUB.IMAGE_REPO)
			os.Setenv("BUS_REPO_URL", vcs.GITHUB.BUS_REPO_URL)
			os.Setenv("DOCKERCONFIGJSON_PATH", vcs.GITHUB.DOCKERCONFIGJSON_PATH)
			os.Setenv("GIT_ACCESS_TOKEN", vcs.GITHUB.GIT_ACCESS_TOKEN)
		} else if os.Getenv("VCS") == "gitlab" {
			os.Setenv("SERVICE_REPO_URL", vcs.GITLAB.SERVICE_REPO_URL)
			os.Setenv("GITOPS_REPO_URL", vcs.GITLAB.GITOPS_REPO_URL)
			os.Setenv("IMAGE_REPO", vcs.GITLAB.IMAGE_REPO)
			os.Setenv("BUS_REPO_URL", vcs.GITLAB.BUS_REPO_URL)
			os.Setenv("DOCKERCONFIGJSON_PATH", vcs.GITLAB.DOCKERCONFIGJSON_PATH)
			os.Setenv("GIT_ACCESS_TOKEN", vcs.GITLAB.GIT_ACCESS_TOKEN)
		} else {
			fmt.Println("VCS not set")
			os.Exit(1)
		}

		envVars := []string{"SERVICE_REPO_URL", "GITOPS_REPO_URL", "IMAGE_REPO", "DOCKERCONFIGJSON_PATH", "GIT_ACCESS_TOKEN", "BUS_REPO_URL"}
		for _, envVar := range envVars {
			value := os.Getenv(envVar)
			if len(value) == 0 {
				fmt.Printf("%s is not set\n", envVar)
				return false
			}
		}

	} else {
		if val == "prow" {
			fmt.Printf("Running e2e test in OpenShift CI\n")
			majorVersion, err := openhiftServerVersion()
			if err != nil {
				fmt.Printf("OpenShift API server version not found\n")
				return false
			}
			if os.Getenv("VCS") == "github" {
				os.Setenv("SERVICE_REPO_URL", vcs.GITHUB.SERVICE_REPO_URL)
				os.Setenv("GITOPS_REPO_URL", vcs.GITHUB.GITOPS_REPO_URL+os.Getenv("PRNO")+majorVersion)
				os.Setenv("IMAGE_REPO", vcs.GITHUB.IMAGE_REPO)
				os.Setenv("BUS_REPO_URL", vcs.GITHUB.BUS_REPO_URL)
				os.Setenv("DOCKERCONFIGJSON_PATH", os.Getenv("KAM_QUAY_DOCKER_CONF_SECRET_FILE"))
				os.Setenv("GIT_ACCESS_TOKEN", os.Getenv("GITHUB_TOKEN"))
			} else if os.Getenv("VCS") == "gitlab" {
				os.Setenv("SERVICE_REPO_URL", vcs.GITLAB.SERVICE_REPO_URL)
				os.Setenv("GITOPS_REPO_URL", vcs.GITLAB.GITOPS_REPO_URL+os.Getenv("PRNO")+majorVersion)
				os.Setenv("IMAGE_REPO", vcs.GITLAB.IMAGE_REPO)
				os.Setenv("BUS_REPO_URL", vcs.GITLAB.BUS_REPO_URL)
				os.Setenv("DOCKERCONFIGJSON_PATH", os.Getenv("KAM_QUAY_DOCKER_CONF_SECRET_FILE"))
				os.Setenv("GIT_ACCESS_TOKEN", os.Getenv("GITLAB_TOKEN"))
			} else {
				fmt.Println("VCS not set")
				os.Exit(1)
			}

		} else {
			fmt.Printf("You cannot run e2e test locally against OpenShift CI\n")
			return false
		}
	}
	return true
}

func waitSync(app string, state string) error {
	err := wait.Poll(time.Second*1, time.Minute*20, func() (bool, error) {
		return argoAppStatusMatch(state, app)
	})
	if err != nil {
		return fmt.Errorf("error is : %v", err)
	}
	return nil
}

func parse(repo string) (string, *scm.Client, error) {
	parsed, err := url.Parse(repo)
	if err != nil {
		return "", nil, err
	}
	name := strings.Split(repo, "/")
	parsed.User = url.UserPassword("", os.Getenv("GIT_ACCESS_TOKEN"))
	client, err := factory.FromRepoURL(parsed.String())
	if err != nil {
		return "", nil, err
	}
	return name[4], client, err
}

func checkRepository(repo string) error {
	if repo == "taxi" {
		_, _, err := parse(os.Getenv("SERVICE_REPO_URL"))
		if err != nil {
			return err
		}
	} else if repo == "bus" {
		_, _, err := parse(os.Getenv("BUS_REPO_URL"))
		if err != nil {
			return err
		}
	}
	return nil
}

func createRepository(repo string) error {

	var repository *scm.Repository

	name, client, err := parse(os.Getenv("GITOPS_REPO_URL"))
	if err != nil {
		return err
	}

	ri := &scm.RepositoryInput{
		Private:     false,
		Description: "",
		Namespace:   "",
		Name:        name,
	}

	repository, _, err = client.Repositories.Create(context.Background(), ri)

	if err != nil {
		return err
	}
	fmt.Printf("Created repositry: %v\n", repository.Name)
	return nil
}

func deleteRepository(repo string) error {
	if os.Getenv("VCS") == "github" {
		name, client, err := parse(repo)
		if err != nil {
			return err
		}

		_, err = client.Repositories.Delete(context.Background(), strings.Split(repo, ".com/")[1])
		if err != nil {
			return err
		}
		fmt.Printf("Deleted repositry: %v\n", name)

		return nil
	} else if os.Getenv("VCS") == "gitlab" {
		var stderr bytes.Buffer

		glabPath, err := executableBinaryPath("glab")
		if err != nil {
			return err
		}

		cmd := exec.Command(glabPath, "auth", "login", "-t", os.Getenv("GIT_ACCESS_TOKEN"))
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return err
		}

		cmd = exec.Command(glabPath, "repo", "delete", strings.Split(os.Getenv("GITOPS_REPO_URL"), ".com/")[1], "-y")
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return err
		}
		fmt.Printf("Deleted repositry: %v\n", os.Getenv("GITOPS_REPO_URL"))

	}
	return nil
}

func createPR() error {
	_, client, err := parse(os.Getenv("GITOPS_REPO_URL"))
	if err != nil {
		return err
	}

	pr := &scm.PullRequestInput{
		Title: "Add new service",
		Head:  "addNewService",
		Base:  "main",
		Body:  "Add new service",
	}

	var response *scm.PullRequest
	response, _, err = client.PullRequests.Create(context.Background(), strings.Split(os.Getenv("GITOPS_REPO_URL"), ".com/")[1], pr)
	if err != nil {
		return err
	}
	fmt.Println("PR has been created")
	err = waitPass(response.Number)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

func mergePR(number int) error {
	_, client, err := parse(os.Getenv("GITOPS_REPO_URL"))
	if err != nil {
		return err
	}

	merge := &scm.PullRequestMergeOptions{
		MergeMethod:        "merge",
		DeleteSourceBranch: true,
	}

	_, err = client.PullRequests.Merge(context.Background(), strings.Split(os.Getenv("GITOPS_REPO_URL"), ".com/")[1], number, merge)
	if err != nil {
		return err
	}
	fmt.Println("PR has been merged!")
	return nil
}

func waitPass(number int) error {

	var stderr bytes.Buffer
	var stdout bytes.Buffer

	ocPath, err := executableBinaryPath("oc")
	if err != nil {
		return err
	}

	err = wait.Poll(time.Second*1, time.Minute*30, func() (bool, error) {

		cmd := exec.Command(ocPath, "get", "pipelinerun", "-n", "cicd")
		cmd.Stderr = &stderr
		cmd.Stdout = &stdout
		err = cmd.Run()

		if err != nil {
			return false, err
		}

		if stdout.String() == "" {
			if strings.Contains(stderr.String(), "No resources found in cicd namespace.") {
				return false, nil
			}
		}
		return true, nil
	})

	if err != nil {
		return err
	}

	fmt.Println("Waiting for checks to pass")
	err = wait.Poll(time.Second*1, time.Minute*20, func() (bool, error) {

		output, err := exec.Command(ocPath, "get", "pipelinerun", "-n", "cicd", "--sort-by=.status.startTime", "-o", "jsonpath='{.items[-1].status.conditions[0].status}'").Output()
		if err != nil {
			fmt.Println("Error:", err)
			return false, err
		}

		if string(output) == "'Unknown'" {
			fmt.Print(".")
			return false, nil
		} else if string(output) == "'True'" {
			fmt.Println("\nPR is ready to be merged, all checks passed!")
			return true, nil
		} else {
			fmt.Println("\nOne or more checks failed!")
			return false, err
		}
	})

	if err != nil {
		return err
	} else {
		err := mergePR(number)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func loginToArgoAPIServer() error {
	var stderr bytes.Buffer
	argocdPath, err := executableBinaryPath("argocd")
	if err != nil {
		return err
	}

	argocdServer, err := argocdAPIServer()
	if err != nil {
		return err
	}

	argocdPassword, err := argocdAPIServerPassword()
	if err != nil {
		return err
	}

	cmd := exec.Command(argocdPath, "login", "--username", "admin", "--password", argocdPassword, argocdServer, "--grpc-web", "--insecure")
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	fmt.Println("Logged in to ArgoCD API server successfully")
	return nil
}

func executableBinaryPath(executable string) (string, error) {
	path, err := exec.LookPath(executable)
	if err != nil {
		return "", err
	}
	return path, nil
}

func argocdAPIServer() (string, error) {
	var stdout bytes.Buffer

	ocPath, err := executableBinaryPath("oc")
	if err != nil {
		return "", err
	}

	deployments := []string{"openshift-gitops-server", "openshift-gitops-repo-server",
		"openshift-gitops-redis", "openshift-gitops-applicationset-controller", "kam", "cluster"}

	for index := range deployments {
		err = wait.Poll(time.Second*1, time.Minute*10, func() (bool, error) {
			return waitForDeploymentsUpAndRunning("openshift-gitops", deployments[index])
		})
	}

	if err != nil {
		return "", err
	}

	cmd := exec.Command(ocPath, "get", "routes", "-n", "openshift-gitops",
		"-o", "jsonpath='{.items[?(@.metadata.name==\"openshift-gitops-server\")].spec.host}'")
	cmd.Stdout = &stdout
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return strings.Trim(stdout.String(), "'"), nil
}

func argocdAPIServerPassword() (string, error) {
	var stdout, stderr bytes.Buffer
	ocPath, err := executableBinaryPath("oc")
	if err != nil {
		return "", err
	}

	cmd := exec.Command(ocPath, "get", "secret", "openshift-gitops-cluster", "-n", "openshift-gitops", "-o", "jsonpath='{.data.admin\\.password}'")

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(strings.Trim(stdout.String(), "'"))
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func waitForDeploymentsUpAndRunning(namespace string, deploymentName string) (bool, error) {
	var stderr, stdout bytes.Buffer
	ocPath, err := executableBinaryPath("oc")
	if err != nil {
		return false, err
	}
	cmd := exec.Command(ocPath, "rollout", "status", "deployment", deploymentName, "-n", namespace)

	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err = cmd.Run()
	if err == nil {
		if strings.Contains(stdout.String(), "successfully rolled out") {
			return true, err
		}
		return false, err
	}
	return false, err
}

func argoAppStatusMatch(matchString string, appName string) (bool, error) {
	var stdout bytes.Buffer
	argocdPath, err := executableBinaryPath("argocd")
	if err != nil {
		return false, err
	}

	appList := []string{"app", "list"}
	cmd := exec.Command(argocdPath, appList...)
	cmd.Stdout = &stdout
	if err = cmd.Run(); err != nil {
		return false, err
	}

	re, _ := regexp.Compile(appName + ".+")
	appDetailsString := re.FindString(stdout.String())
	if appDetailsString == " " {
		return false, nil
	}
	if strings.Contains(appDetailsString, matchString) {
		return true, nil
	}
	return false, nil
}

func openhiftServerVersion() (string, error) {
	var stdout bytes.Buffer
	ocPath, err := executableBinaryPath("oc")
	if err != nil {
		return "", err
	}
	cmd := exec.Command(ocPath, "version")
	cmd.Stdout = &stdout
	if err = cmd.Run(); err != nil {
		return "", err
	}

	re := regexp.MustCompile(`Server\s+Version:\s+(\d\.\d\d?)`)
	return strings.Replace(strings.Trim(re.FindStringSubmatch(stdout.String())[1], "\""), ".", "", -1), nil
}

func addResource() error {
	appFs := ioutils.NewFilesystem()
	const (
		path           = "bootstrapresources/environments/new-env/apps/app-bus/services/bus/base/config"
		partOf         = "app-bus"
		env            = "new-env"
		name           = "bus"
		bootstrapImage = "nginxinc/nginx-unprivileged:latest"
	)
	resources := res.Resources{}
	resources[filepath.Join(path, "deployment.yaml")] = deployment.Create(partOf, env, name, bootstrapImage, deployment.ContainerPort(8080))
	resources[filepath.Join(path, "kustomization.yaml")] = &res.Kustomization{
		Resources: []string{
			"deployment.yaml",
		}}
	_, err := yaml.WriteResources(appFs, "", resources)
	return err
}
