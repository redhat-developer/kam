package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"

	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/openshift/odo/pkg/log"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/kam/pkg/cmd/genericclioptions"
	"github.com/redhat-developer/kam/pkg/cmd/ui"
	"github.com/redhat-developer/kam/pkg/cmd/utility"
	"github.com/redhat-developer/kam/pkg/pipelines"
	"github.com/redhat-developer/kam/pkg/pipelines/accesstoken"
	"github.com/redhat-developer/kam/pkg/pipelines/argocd"
	"github.com/redhat-developer/kam/pkg/pipelines/imagerepo"
	"github.com/redhat-developer/kam/pkg/pipelines/ioutils"
)

const (
	// BootstrapRecommendedCommandName the recommended command name
	BootstrapRecommendedCommandName = "bootstrap"

	pipelinesOperatorNS    = "openshift-operators"
	gitopsRepoURLFlag      = "gitops-repo-url"
	serviceRepoURLFlag     = "service-repo-url"
	gitHostAccessTokenFlag = "git-host-access-token"
	imageRepoFlag          = "image-repo"
	gitopsOperatorName     = "OpenShift GitOps Operator"
	pipelinesOperatorName  = "OpenShift Pipelines Operator"
)

type drivers []string

var (
	supportedDrivers = drivers{
		"github",
		"gitlab",
	}
)

func (d drivers) supported(s string) bool {
	for _, v := range d {
		if s == v {
			return true
		}
	}
	return false
}

var (
	bootstrapExample = ktemplates.Examples(`
    # Bootstrap OpenShift pipelines.
		kam bootstrap --service-repo-url https://github.com/<your organization>/taxi.git --gitops-repo-url https://github.com/<your organization>/gitops.git --image-repo quay.io/<username>/<image-repo> --dockercfgjson ~/Downloads/<username>-robot-auth.json --git-host-access-token <your git access token> --output <path to write GitOps resources> --push-to-git=true
		
    %[1]s 
    `)

	bootstrapLongDesc  = ktemplates.LongDesc(`Bootstrap GitOps CI/CD Manifests`)
	bootstrapShortDesc = `Bootstrap GitOps CI/CD with a starter configuration`
)

// BootstrapParameters encapsulates the parameters for the kam pipelines init command.
type BootstrapParameters struct {
	*pipelines.BootstrapOptions
	Interactive bool
}

// NewBootstrapParameters bootsraps a Bootstrap Parameters instance.
func NewBootstrapParameters() *BootstrapParameters {
	return &BootstrapParameters{
		BootstrapOptions: &pipelines.BootstrapOptions{},
	}
}

// Complete completes BootstrapParameters after they've been created.
// If the prefix provided doesn't have a "-" then one is added, this makes the
// generated environment names nicer to read.
func (io *BootstrapParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	client, err := utility.NewClient()
	if err != nil {
		return err
	}

	if io.PrivateRepoDriver != "" {
		host, err := accesstoken.HostFromURL(io.GitOpsRepoURL)
		if err != nil {
			return err
		}
		identifier := factory.NewDriverIdentifier(factory.Mapping(host, io.PrivateRepoDriver))
		factory.DefaultIdentifier = identifier
	}
	if err := checkBootstrapDependencies(io, client, log.NewStatus(os.Stdout)); err != nil {
		return err
	}

	if cmd.Flags().NFlag() == 0 || io.Interactive {
		return initiateInteractiveMode(io, client, cmd)
	}

	addGitURLSuffixIfNecessary(io)
	return nonInteractiveMode(io, client)
}

func addGitURLSuffixIfNecessary(io *BootstrapParameters) {
	io.GitOpsRepoURL = utility.AddGitSuffixIfNecessary(io.GitOpsRepoURL)
	io.ServiceRepoURL = utility.AddGitSuffixIfNecessary(io.ServiceRepoURL)
}

// nonInteractiveMode gets triggered if a flag is passed, checks for mandatory flags.
func nonInteractiveMode(io *BootstrapParameters, client *utility.Client) error {
	mandatoryFlags := map[string]string{serviceRepoURLFlag: io.ServiceRepoURL, gitopsRepoURLFlag: io.GitOpsRepoURL, gitHostAccessTokenFlag: io.GitHostAccessToken}
	if err := checkMandatoryFlags(mandatoryFlags); err != nil {
		return err
	}
	err := setAccessToken(io)
	if err != nil {
		return err
	}
	return nil
}

func checkMandatoryFlags(flags map[string]string) error {
	missingFlags := []string{}
	mandatoryFlags := []string{serviceRepoURLFlag, gitopsRepoURLFlag, gitHostAccessTokenFlag}
	for _, flag := range mandatoryFlags {
		if flags[flag] == "" {
			missingFlags = append(missingFlags, fmt.Sprintf("%q", flag))
		}
	}
	if len(missingFlags) > 0 {
		return missingFlagErr(missingFlags)
	}
	return nil
}

func missingFlagErr(flags []string) error {
	return fmt.Errorf("required flag(s) %s not set", strings.Join(flags, ", "))
}

// initiateInteractiveMode starts the interactive mode impplementation if no flags are passed.
func initiateInteractiveMode(io *BootstrapParameters, client *utility.Client, cmd *cobra.Command) error {
	log.Progressf("\nStarting interactive prompt\n")
	// Prompt if user wants to use all default values and only be prompted with required or other necessary questions
	promptForAll := !ui.UseDefaultValues()
	if io.GitOpsRepoURL == "" {
		io.GitOpsRepoURL = ui.EnterGitRepo()
	}
	io.GitOpsRepoURL = utility.AddGitSuffixIfNecessary(io.GitOpsRepoURL)
	if !isKnownDriver(io.GitOpsRepoURL) {
		io.PrivateRepoDriver = ui.SelectPrivateRepoDriver()
		host, err := accesstoken.HostFromURL(io.GitOpsRepoURL)
		if err != nil {
			return fmt.Errorf("failed to parse the gitops url: %w", err)
		}
		identifier := factory.NewDriverIdentifier(factory.Mapping(host, io.PrivateRepoDriver))
		factory.DefaultIdentifier = identifier
	}
	if io.ImageRepo != "" {
		isInternalRegistry, _, err := imagerepo.ValidateImageRepo(io.ImageRepo)
		if err != nil {
			return err
		}
		if !isInternalRegistry {
			if !cmd.Flag("dockercfgjson").Changed && promptForAll {
				log.Progressf("The supplied image repository has been detected as an external repository.")
				io.DockerConfigJSONFilename = ui.EnterDockercfg()
			}
		}
	} else if promptForAll {
		if ui.UseInternalRegistry() {
			io.ImageRepo = ui.EnterImageRepoInternalRegistry()
		} else {
			io.ImageRepo = ui.EnterImageRepoExternalRepository()
			io.DockerConfigJSONFilename = ui.EnterDockercfg()
		}
	}
	if promptForAll {
		io.GitOpsWebhookSecret = ui.EnterGitWebhookSecret(io.GitOpsRepoURL)
	}
	if io.ServiceRepoURL == "" {
		io.ServiceRepoURL = ui.EnterServiceRepoURL()
	}
	io.ServiceRepoURL = utility.AddGitSuffixIfNecessary(io.ServiceRepoURL)
	if promptForAll {
		io.ServiceWebhookSecret = ui.EnterGitWebhookSecret(io.ServiceRepoURL)
	}
	secret, err := accesstoken.GetAccessToken(io.ServiceRepoURL)
	if err != nil && err != keyring.ErrNotFound {
		return err
	}
	if secret == "" { // We must prompt for the token
		if io.GitHostAccessToken == "" {
			io.GitHostAccessToken = ui.EnterGitHostAccessToken(io.ServiceRepoURL)
		}
		if !cmd.Flag("save-token-keyring").Changed {
			io.SaveTokenKeyRing = ui.UseKeyringRingSvc()
		}
		setAccessToken(io)
	} else {
		io.GitHostAccessToken = secret
	}
	if !cmd.Flag("push-to-git").Changed && promptForAll {
		io.PushToGit = ui.SelectOptionPushToGit()
	}
	if io.Prefix == "" && promptForAll {
		io.Prefix = ui.EnterPrefix()
	}
	outputPathOverridden := cmd.Flag("output").Changed
	if !outputPathOverridden {
		// Override the default path to be ./{gitops repo name}
		repoName, err := repoFromURL(io.GitOpsRepoURL)
		if err != nil {
			repoName = "gitops"
		}
		io.OutputPath = filepath.Join(".", repoName)
	}
	appFs := ioutils.NewFilesystem()
	io.OutputPath, io.Overwrite = ui.VerifyOutputPath(appFs, io.OutputPath, io.Overwrite, outputPathOverridden, promptForAll)
	if !io.Overwrite {
		if ui.PathExists(appFs, filepath.Join(io.OutputPath, "..", "secrets")) {
			return fmt.Errorf("the secrets folder located as a sibling of the output folder %s already exists. Delete or rename the secrets folder and try again", io.OutputPath)
		}
		if io.PushToGit && ui.PathExists(appFs, filepath.Join(io.OutputPath, ".git")) {
			return fmt.Errorf("the .git folder in output path %s already exists. Delete or rename the .git folder and try again", io.OutputPath)
		}
	}
	return nil
}

func repoFromURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	parts := strings.Split(u.Path, "/")
	return strings.TrimSuffix(parts[len(parts)-1], ".git"), nil
}

func setAccessToken(io *BootstrapParameters) error {
	if io.GitHostAccessToken != "" {
		err := ui.ValidateAccessToken(io.GitHostAccessToken, io.ServiceRepoURL)
		if err != nil {
			return fmt.Errorf("Access token validation failed: %v", err)
		}
	}
	if io.SaveTokenKeyRing {
		err := accesstoken.SetAccessToken(io.ServiceRepoURL, io.GitHostAccessToken)
		if err != nil {
			return err
		}
	}
	if io.GitHostAccessToken == "" {
		secret, err := accesstoken.GetAccessToken(io.ServiceRepoURL)
		if err != nil {
			return fmt.Errorf("unable to use access-token from keyring/env-var: %v, please pass a valid token to --git-host-access-token", err)
		}
		io.GitHostAccessToken = secret
	}
	return nil
}

func checkBootstrapDependencies(io *BootstrapParameters, client *utility.Client, spinner utility.Status) error {
	missingDeps := []string{}
	log.Progressf("\nChecking dependencies\n")

	spinner.Start("Checking if Argo CD is installed with the default configuration", false)
	if err := client.CheckIfArgoCDExists(argocd.ArgoCDNamespace); err != nil {
		warnIfNotFound(spinner, "Please install OpenShift GitOps Operator from OperatorHub", err)
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to check for OpenShift GitOps Operator: %w", err)
		}
		missingDeps = append(missingDeps, gitopsOperatorName)
	}

	spinner.Start("Checking if OpenShift Pipelines Operator is installed with the default configuration", false)
	if err := client.CheckIfPipelinesExists(pipelinesOperatorNS); err != nil {
		warnIfNotFound(spinner, "Please install OpenShift Pipelines Operator from OperatorHub", err)
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to check for OpenShift Pipelines Operator: %w", err)
		}
		missingDeps = append(missingDeps, pipelinesOperatorName)
	}
	spinner.End(true)
	if len(missingDeps) > 0 {
		return fmt.Errorf("failed to satisfy the required dependencies: %s", strings.Join(missingDeps, ", "))
	}
	return nil
}

func warnIfNotFound(spinner utility.Status, warningMsg string, err error) {
	if apierrors.IsNotFound(err) {
		spinner.WarningStatus(warningMsg)
	}
	spinner.End(false)
}

// Validate validates the parameters of the BootstrapParameters.
func (io *BootstrapParameters) Validate() error {
	gr, err := url.Parse(io.GitOpsRepoURL)
	if err != nil {
		return fmt.Errorf("failed to parse url %s: %w", io.GitOpsRepoURL, err)
	}

	// GitLab as the repo can have more path elements. Check that there are at least 2.
	if len(utility.RemoveEmptyStrings(strings.Split(gr.Path, "/"))) <= 2 {
		return fmt.Errorf("repo must be org/repo: %s", strings.Trim(gr.Path, ".git"))
	}

	if io.PrivateRepoDriver != "" {
		if !supportedDrivers.supported(io.PrivateRepoDriver) {
			return fmt.Errorf("invalid driver type: %q", io.PrivateRepoDriver)
		}
	}
	if io.SaveTokenKeyRing && io.GitHostAccessToken == "" {
		return errors.New("--git-host-access-token is required if --save-token-keyring is enabled")
	}
	io.Prefix = utility.MaybeCompletePrefix(io.Prefix)
	return nil
}

// Run runs the project Bootstrap command.
func (io *BootstrapParameters) Run() error {
	log.Progressf("\nCompleting Bootstrap process\n")
	appFs := ioutils.NewFilesystem()
	err := pipelines.Bootstrap(io.BootstrapOptions, appFs)
	if err != nil {
		return err
	}
	if io.PushToGit {
		err = pipelines.BootstrapRepository(io.BootstrapOptions, factory.FromRepoURL, pipelines.NewCmdExecutor(), appFs)
		if err != nil {
			return fmt.Errorf("failed to create the gitops repository: %q: %w", io.GitOpsRepoURL, err)
		}
		log.Successf("Created repository")
	}
	nextSteps()
	return nil
}

// NewCmdBootstrap creates the project init command.
func NewCmdBootstrap(name, fullName string) *cobra.Command {
	o := NewBootstrapParameters()

	bootstrapCmd := &cobra.Command{
		Use:     name,
		Short:   bootstrapShortDesc,
		Long:    bootstrapLongDesc,
		Example: fmt.Sprintf(bootstrapExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	bootstrapCmd.Flags().StringVar(&o.GitOpsRepoURL, "gitops-repo-url", "", "Provide the URL for your GitOps repository e.g. https://github.com/organisation/repository.git")
	bootstrapCmd.Flags().StringVar(&o.GitOpsWebhookSecret, "gitops-webhook-secret", "", "Provide a secret that we can use to authenticate incoming hooks from your Git hosting service for the GitOps repository. (if not provided, it will be auto-generated)")
	bootstrapCmd.Flags().StringVar(&o.OutputPath, "output", "./gitops", "Path to write GitOps resources")
	bootstrapCmd.Flags().StringVarP(&o.Prefix, "prefix", "p", "", "Add a prefix to the environment names(Dev, stage,prod,cicd etc.) to distinguish and identify individual environments")
	bootstrapCmd.Flags().StringVar(&o.DockerConfigJSONFilename, "dockercfgjson", "~/.docker/config.json", "Filepath to config.json which authenticates the image push to the desired image registry ")
	bootstrapCmd.Flags().StringVar(&o.ImageRepo, "image-repo", "", "Image repository of the form <registry>/<username>/<repository> or <project>/<app> which is used to push newly built images")
	bootstrapCmd.Flags().StringVar(&o.GitHostAccessToken, "git-host-access-token", "", "Used to authenticate repository clones. Access token is encrypted and stored on local file system by keyring, will be updated/reused.")
	bootstrapCmd.Flags().BoolVar(&o.Overwrite, "overwrite", false, "Overwrites previously existing GitOps configuration (if any) on the local filesystem")
	bootstrapCmd.Flags().StringVar(&o.ServiceRepoURL, "service-repo-url", "", "Provide the URL for your Service repository e.g. https://github.com/organisation/service.git")
	bootstrapCmd.Flags().StringVar(&o.ServiceWebhookSecret, "service-webhook-secret", "", "Provide a secret that we can use to authenticate incoming hooks from your Git hosting service for the Service repository. (if not provided, it will be auto-generated)")
	bootstrapCmd.Flags().BoolVar(&o.SaveTokenKeyRing, "save-token-keyring", false, "Explicitly pass this flag to update the git-host-access-token in the keyring on your local machine")
	bootstrapCmd.Flags().StringVar(&o.PrivateRepoDriver, "private-repo-driver", "", "If your Git repositories are on a custom domain, please indicate which driver to use github or gitlab")
	bootstrapCmd.Flags().BoolVar(&o.PushToGit, "push-to-git", false, "If true, automatically creates and populates the gitops-repo-url with the generated resources")
	bootstrapCmd.Flags().BoolVar(&o.Interactive, "interactive", false, "If true, enable prompting for most options if not already specified on the command line")
	return bootstrapCmd
}

func nextSteps() {
	log.Success("Bootstrapped OpenShift resources successfully\n\n",
		"Next Steps:\n",
		"Please refer to https://github.com/redhat-developer/kam/tree/master/docs to get started.\n",
	)
	log.Info(" WARNING: Generated secrets are not encrypted. Deploying the GitOps configuration without encrypting secrets is insecure and is not recommended.\n For more information on secret management see: https://github.com/redhat-developer/kam/tree/master/docs/journey/day1#secrets\n")
}

func isKnownDriver(repoURL string) bool {
	host, err := accesstoken.HostFromURL(repoURL)
	if err != nil {
		return false
	}
	_, err = factory.DefaultIdentifier.Identify(host)
	return err == nil
}
