package ui

import (
	"fmt"
	"net/url"
	"path/filepath"

	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/redhat-developer/kam/pkg/pipelines/ioutils"
	"k8s.io/apimachinery/pkg/types"
)

// EnterGitRepo allows the user to specify the git repository in a prompt.
func EnterGitRepo() string {
	var gitOpsURL string
	prompt := &survey.Input{
		Message: "Provide the URL for your GitOps repository",
		Help:    "The GitOps repository stores your GitOps configuration files, including your Openshift Pipelines resources for driving automated deployments and builds.  Please enter a valid git repository e.g. https://github.com/example/myorg.git",
	}
	err := survey.AskOne(prompt, &gitOpsURL, survey.Required)
	handleError(err)

	p, err := url.Parse(gitOpsURL)
	handleError(err)
	if p.Host == "" {
		handleError(fmt.Errorf("could not identify host from %q", gitOpsURL))
	}
	return gitOpsURL
}

// EnterInternalRegistry allows the user to specify the internal registry in a UI prompt.
func EnterInternalRegistry() string {
	var internalRegistry string
	prompt := &survey.Input{
		Message: "Host-name for internal image registry to be used if you are pushing your images to the internal image registry",
		Default: "image-registry.openshift-image-registry.svc:5000",
	}

	err := survey.AskOne(prompt, &internalRegistry, nil)
	handleError(err)
	return internalRegistry
}

// EnterImageRepoInternalRegistry allows the user to specify the internal image
// registry in a UI prompt.
func EnterImageRepoInternalRegistry() string {
	var imageRepo string
	prompt := &survey.Input{
		Message: "Image registry of the form <project>/<app> which is used to push newly built images.",
		Help:    "By default images are built from source, whenever there is a push to the repository for your service source code and this image will be pushed to the image registry specified in this parameter, if the value is of the form <registry>/<username>/<image name>, then it assumed that it is an upstream image registry e.g. Quay, if it's of the form <project>/<app> the internal registry present on the current cluster will be used as the image registry.",
	}

	err := survey.AskOne(prompt, &imageRepo, survey.Required)
	handleError(err)
	return imageRepo
}

// EnterDockercfg allows the user to specify the path to the docker config json
// file for external image registry authentication in a UI prompt.
func EnterDockercfg() string {
	var dockerCfg string
	prompt := &survey.Input{
		Message: "Path to config.json which authenticates image pushes to the desired image registry",
		Help:    "The secret present in the file path generates a secure secret that authenticates the push of the image built when the app-ci pipeline is run. The image along with the necessary labels will be present on the upstream image registry of choice.",
		Default: "~/.docker/config.json",
	}

	err := survey.AskOne(prompt, &dockerCfg, nil)
	handleError(err)
	return dockerCfg
}

// EnterImageRepoExternalRepository allows the user to specify the type of image
// registry they wish to use in a UI prompt.
func EnterImageRepoExternalRepository() string {
	var imageRepoExt string
	prompt := &survey.Input{
		Message: "Image registry of the form <registry>/<username>/<image name> which is used to push newly built images.",
		Help:    "By default images are built from source whenever there is a push to the repository for your service source code and this image will be pushed to the image registry specified in this parameter, if the value is of the form <registry>/<username>/<image name>, then it assumed that it is an upstream image registry e.g. Quay, if its of the form <project>/<app> the internal registry present on the current cluster will be used as the image registry.",
	}

	err := survey.AskOne(prompt, &imageRepoExt, survey.Required)
	handleError(err)
	return imageRepoExt
}

// EnterOutputPath allows the user to specify the path where the gitops configuration must reside locally in a UI prompt.
func EnterOutputPath() string {
	var outputPath string
	prompt := &survey.Input{
		Message: "Provide a path to write GitOps resources?",
		Help:    "This is the path where the GitOps repository configuration is stored locally before you push it to the repository GitopsRepoURL",
		Default: ".",
	}

	err := survey.AskOne(prompt, &outputPath, nil)
	exists, filePathError := ioutils.IsExisting(ioutils.NewFilesystem(), filepath.Join(outputPath, "pipelines.yaml"))
	if exists {
		SelectOptionOverwrite(outputPath)
	}
	if !exists {
		err = filePathError
	}
	handleError(err)
	return outputPath
}

// EnterGitWebhookSecret allows the user to specify the webhook secret string
// they wish to authenticate push/pull to GitOps repo in a UI prompt.
func EnterGitWebhookSecret() string {
	var gitWebhookSecret string
	prompt := &survey.Input{
		Message: "Provide a secret (minimum 16 characters) that we can use to authenticate incoming hooks from your Git hosting service for the Service repository. (if not provided, it will be auto-generated)",
		Help:    "You can provide a string that is used as a shared secret to authenticate the origin of hook notifications from your git host.",
	}

	err := survey.AskOne(prompt, &gitWebhookSecret, makeSecretValidator())
	handleError(err)
	return gitWebhookSecret
}

// enterSealedSecretService , if the secret isnt installed using the operator it is necessary to manually add the sealed-secrets-controller name through this UI prompt.
func enterSealedSecretService() string {
	var sealedSecret string
	prompt := &survey.Input{
		Message: "Name of the Sealed Secrets Service that encrypts secrets",
		Help:    "If you have a custom installation of the Sealed Secrets operator, we need to know where to communicate with it to seal your secrets.",
	}
	err := survey.AskOne(prompt, &sealedSecret, survey.Required)
	handleError(err)
	return sealedSecret
}

// EnterSealedSecretService , prompts the UI to ask for the sealed-secrets-namespaces
func EnterSealedSecretService(sealedSecretService *types.NamespacedName) string {
	var sealedNs string
	prompt := &survey.Input{
		Message: "Provide a namespace in which the Sealed Secrets operator is installed, automatically generated secrets are encrypted with this operator?",
		Help:    "If you have a custom installation of the Sealed Secrets operator, we need to know how to communicate with it to seal your secrets",
	}

	err := survey.AskOne(prompt, &sealedNs, makeSealedSecretsService(sealedSecretService))
	handleError(err)
	return sealedNs
}

// EnterGitHostAccessToken , it becomes necessary to add the personal access
// token to access upstream git hosts.
func EnterGitHostAccessToken(serviceRepo string) string {
	var accessToken string
	prompt := &survey.Password{
		Message: fmt.Sprintf("Please provide a token used to authenticate requests to %q", serviceRepo),
		Help:    "commit-status-tracker reports the completion status of OpenShift pipeline runs to your Git hosting status on success or failure, this token will be encrypted as a secret in your cluster.\nIf you are using Github, please see here for how to generate a token https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token\nIf you are using GitLab, please see here for how to generate a token https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html",
	}
	err := survey.AskOne(prompt, &accessToken, makeAccessTokenCheck(serviceRepo))
	handleError(err)
	return accessToken
}

// EnterPrefix , if we desire to add the prefix to differentiate between namespaces, then this is the way forward.
func EnterPrefix() string {
	var prefix string
	prompt := &survey.Input{
		Message: "Add a prefix to the environment names(dev, stage, cicd etc.) to distinguish and identify individual environments?",
		Help:    "The prefix helps differentiate between the different namespaces on the cluster, the default namespace cicd will appear as test-cicd if the prefix passed is test.",
	}
	err := survey.AskOne(prompt, &prefix, makePrefixValidator())
	handleError(err)
	return prefix
}

// EnterServiceRepoURL , allows users to differentiate between the bootstrap and init options, addition of the service repo url will allow users to bootstrap an environment through the UI prompt.
func EnterServiceRepoURL() string {
	var serviceRepo string
	prompt := &survey.Input{
		Message: "Provide the URL for your Service repository e.g. https://github.com/organisation/service.git",
		Help:    "The repository name where the source code of your service is situated, this will configure a very basic CI for this repository using OpenShift pipelines.",
	}
	err := survey.AskOne(prompt, &serviceRepo, survey.Required)
	handleError(err)

	p, err := url.Parse(serviceRepo)
	handleError(err)
	if p.Host == "" {
		handleError(fmt.Errorf("could not identify host from %q", serviceRepo))
	}
	return serviceRepo
}

// EnterServiceWebhookSecret allows the user to specify the webhook secret string they wish to authenticate push/pull to service repo in a UI prompt.
func EnterServiceWebhookSecret() string {
	var serviceWebhookSecret string
	prompt := &survey.Input{
		Message: "Provide a secret (minimum 16 characters) that we can use to authenticate incoming hooks from your Git hosting service for the Service repository. (if not provided, it will be auto-generated)",
		Help:    "You can provide a string that is used as a shared secret to authenticate the origin of hook notifications from your git host.",
	}
	err := survey.AskOne(prompt, &serviceWebhookSecret, makeSecretValidator())

	handleError(err)
	return serviceWebhookSecret
}

// SelectOptionImageRepository , allows users an option between the Internal image registry and the external image registry through the UI prompt.
func SelectOptionImageRepository() string {
	var optionImageRegistry string
	prompt := &survey.Select{
		Message: "Select type of image registry",
		Options: []string{"Openshift Internal registry", "External Registry"},
		Default: "Openshift Internal registry",
	}

	err := survey.AskOne(prompt, &optionImageRegistry, survey.Required)
	handleError(err)
	return optionImageRegistry
}

// SelectOptionOverwrite allows users the option to overwrite the current gitops configuration locally through the UI prompt.
func SelectOptionOverwrite(path string) string {
	var overwrite string
	prompt := &survey.Select{
		Message: "Do you want to overwrite your output path?",
		Options: []string{"yes", "no"},
		Default: "no",
	}
	err := survey.AskOne(prompt, &overwrite, makeOverWriteValidator(path))
	handleError(err)
	return overwrite
}

// SelectOptionCommitStatusTracker allows users the option to select if they
// want to incorporate the feature of the commit status tracker through the UI prompt.
func SelectOptionCommitStatusTracker() string {
	var optionCommitStatusTracker string
	prompt := &survey.Select{
		Message: "Do you want to enable commit-status-tracker?",
		Help:    "commit-status-tracker reports the completion status of OpenShift pipeline runs to your git host on success or failure",
		Options: []string{"yes", "no"},
	}
	err := survey.AskOne(prompt, &optionCommitStatusTracker, survey.Required)
	handleError(err)
	return optionCommitStatusTracker
}

// SelectPrivateRepoDriver lets users choose the driver for their git hosting
// service.
func SelectPrivateRepoDriver() string {
	var driver string
	prompt := &survey.Select{
		Message: "Please select which driver to use for your Git host",
		Options: []string{"github", "gitlab"},
	}

	err := survey.AskOne(prompt, &driver, survey.Required)
	handleError(err)
	return driver
}

// IsPrivateRepo lets the user choose between a private / public repository.
func IsPrivateRepo() bool {
	var response string
	prompt := &survey.Select{
		Message: "Is this repository a private repository?",
		Options: []string{"yes", "no"},
	}

	err := survey.AskOne(prompt, &response, survey.Required)
	handleError(err)
	return response == "yes"
}

// SelectOptionPushToGit allows users the option to select if they
// want to incorporate the feature of the commit status tracker through the UI prompt.
func SelectOptionPushToGit() bool {
	var optionPushToGit string
	prompt := &survey.Select{
		Message: "Do you want to create and push the resources to your gitops repository?",
		Help:    "This will create a private repository, commit and push the generated resources and requires an auth token with the correct privileges",
		Options: []string{"yes", "no"},
	}
	err := survey.AskOne(prompt, &optionPushToGit, survey.Required)
	handleError(err)
	return optionPushToGit == "yes"
}
