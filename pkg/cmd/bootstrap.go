package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/zalando/go-keyring"

	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/openshift/odo/pkg/log"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/kam/pkg/cmd/genericclioptions"
	"github.com/redhat-developer/kam/pkg/cmd/ui"
	"github.com/redhat-developer/kam/pkg/cmd/utility"
	"github.com/redhat-developer/kam/pkg/pipelines"
	"github.com/redhat-developer/kam/pkg/pipelines/accesstoken"
	"github.com/redhat-developer/kam/pkg/pipelines/ioutils"
	"github.com/redhat-developer/kam/pkg/pipelines/secrets"
	"github.com/redhat-developer/kam/pkg/pipelines/statustracker"
)

const (
	// BootstrapRecommendedCommandName the recommended command name
	BootstrapRecommendedCommandName = "bootstrap"

	argoCDNS              = "argocd"
	pipelinesOperatorNS   = "openshift-operators"
	gitopsRepoURLFlag     = "gitops-repo-url"
	serviceRepoURLFlag    = "service-repo-url"
	imageRepoFlag         = "image-repo"
	argoCdOperatorName    = "ArgoCD Operator"
	pipelinesOperatorName = "OpenShift Pipelines Operator"
)

type drivers []string

var (
	supportedDrivers = drivers{
		"github",
		"gitlab",
	}
	defaultSealedSecretsServiceName = types.NamespacedName{Namespace: secrets.SealedSecretsNS, Name: secrets.SealedSecretsController}
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
    %[1]s 
    `)

	bootstrapLongDesc  = ktemplates.LongDesc(`Bootstrap GitOps CI/CD Manifests`)
	bootstrapShortDesc = `Bootstrap GitOps CI/CD with a starter configuration`
)

// BootstrapParameters encapsulates the parameters for the kam pipelines init command.
type BootstrapParameters struct {
	*pipelines.BootstrapOptions
	PushToGit bool // records whether or not the repository should be pushed to git.
}

type status interface {
	WarningStatus(status string)
	Start(status string, debug bool)
	End(status bool)
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

	if cmd.Flags().NFlag() == 0 {
		return initiateInteractiveMode(io, client)
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
	mandatoryFlags := map[string]string{serviceRepoURLFlag: io.ServiceRepoURL, gitopsRepoURLFlag: io.GitOpsRepoURL, imageRepoFlag: io.ImageRepo}
	if err := checkMandatoryFlags(mandatoryFlags); err != nil {
		return err
	}
	err := setAccessTokenNonInteractive(io)
	if err != nil {
		return err
	}
	return nil
}

func setAccessTokenNonInteractive(io *BootstrapParameters) error {
	if io.GitHostAccessToken != "" {
		err := ui.ValidateAccessToken(io.GitHostAccessToken, io.ServiceRepoURL)
		if err != nil {
			return fmt.Errorf("Please enter a valid access token for --save-token-keyring: %v", err)
		}
	}
	if io.SaveTokenKeyRing != "" {
		err := ui.ValidateAccessToken(io.SaveTokenKeyRing, io.ServiceRepoURL)
		if err != nil {
			return fmt.Errorf("Please enter a valid access token for --save-token-keyring: %v", err)
		}
		err = accesstoken.SetAccessToken(io.ServiceRepoURL, io.SaveTokenKeyRing)
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

func checkMandatoryFlags(flags map[string]string) error {
	missingFlags := []string{}
	mandatoryFlags := []string{serviceRepoURLFlag, gitopsRepoURLFlag, imageRepoFlag}
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
func initiateInteractiveMode(io *BootstrapParameters, client *utility.Client) error {
	log.Progressf("\nStarting interactive prompt\n")
	// ask for sealed secrets only when default is absent
	if client.CheckIfSealedSecretsExists(defaultSealedSecretsServiceName) != nil {
		io.SealedSecretsService.Namespace = ui.EnterSealedSecretService(&io.SealedSecretsService)
	}
	io.GitOpsRepoURL = utility.AddGitSuffixIfNecessary(ui.EnterGitRepo())
	if !isKnownDriver(io.GitOpsRepoURL) {
		io.PrivateRepoDriver = ui.SelectPrivateRepoDriver()
		host, err := accesstoken.HostFromURL(io.GitOpsRepoURL)
		if err != nil {
			return fmt.Errorf("failed to parse the gitops url: %w", err)
		}
		identifier := factory.NewDriverIdentifier(factory.Mapping(host, io.PrivateRepoDriver))
		factory.DefaultIdentifier = identifier
	}
	if ui.UseInternalRegistry() {
		io.InternalRegistryHostname = ui.EnterInternalRegistry()
		io.ImageRepo = ui.EnterImageRepoInternalRegistry()
	} else {
		io.ImageRepo = ui.EnterImageRepoExternalRepository()
		io.DockerConfigJSONFilename = ui.EnterDockercfg()
	}
	io.GitOpsWebhookSecret = ui.EnterGitWebhookSecret(io.GitOpsRepoURL)
	io.ServiceRepoURL = utility.AddGitSuffixIfNecessary(ui.EnterServiceRepoURL())
	io.ServiceWebhookSecret = ui.EnterGitWebhookSecret(io.ServiceRepoURL)
	secret, err := accesstoken.GetAccessToken(io.ServiceRepoURL)
	if err != nil && err != keyring.ErrNotFound {
		return err
	}
	if secret == "" {
		io.GitHostAccessToken = ui.EnterGitHostAccessToken(io.ServiceRepoURL)
		keyringOption := ui.UseKeyringRingSvc()
		if keyringOption {
			err := accesstoken.SetAccessToken(io.ServiceRepoURL, io.GitHostAccessToken)
			if err != nil {
				return err
			}
		}
	} else {
		io.GitHostAccessToken = secret
	}
	io.CommitStatusTracker = ui.SetupCommitStatusTracker()
	io.PushToGit = ui.SelectOptionPushToGit()
	io.Prefix = ui.EnterPrefix()
	io.OutputPath = ui.EnterOutputPath()
	io.Overwrite = true
	return nil
}

func checkBootstrapDependencies(io *BootstrapParameters, client *utility.Client, spinner status) error {
	missingDeps := []string{}
	log.Progressf("\nChecking dependencies\n")

	spinner.Start("Checking if Sealed Secrets is installed with the default configuration", false)
	if err := client.CheckIfSealedSecretsExists(defaultSealedSecretsServiceName); err != nil {
		warnIfNotFound(spinner, "Please install Sealed Secrets Operator from OperatorHub", err)
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to check for Sealed Secrets Operator: %w", err)
		}
		// We do not add Sealed Secret Operator to missingDeps since we this dependency can be resolved
		// by optional flags or interactive user inputs.
	} else {
		io.SealedSecretsService = defaultSealedSecretsServiceName
	}

	spinner.Start("Checking if ArgoCD Operator is installed with the default configuration", false)
	if err := client.CheckIfArgoCDExists(argoCDNS); err != nil {
		warnIfNotFound(spinner, "Please install ArgoCD Operator from OperatorHub", err)
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to check for ArgoCD Operator: %w", err)
		}
		missingDeps = append(missingDeps, argoCdOperatorName)
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

func warnIfNotFound(spinner status, warningMsg string, err error) {
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

	// TODO: this may not work with GitLab as the repo can have more path elements.
	if len(utility.RemoveEmptyStrings(strings.Split(gr.Path, "/"))) != 2 {
		return fmt.Errorf("repo must be org/repo: %s", strings.Trim(gr.Path, ".git"))
	}

	if io.PrivateRepoDriver != "" {
		if !supportedDrivers.supported(io.PrivateRepoDriver) {
			return fmt.Errorf("invalid driver type: %q", io.PrivateRepoDriver)
		}
	}
	if io.CommitStatusTracker && io.GitHostAccessToken == "" {
		return errors.New("--git-host-access-token is required if commit-status-tracker is enabled")
	}
	io.Prefix = utility.MaybeCompletePrefix(io.Prefix)
	return nil
}

// Run runs the project Bootstrap command.
func (io *BootstrapParameters) Run() error {
	log.Progressf("\nCompleting Bootstrap process\n")
	err := pipelines.Bootstrap(io.BootstrapOptions, ioutils.NewFilesystem())
	if err != nil {
		return err
	}
	if io.PushToGit {
		err = pipelines.BootstrapRepository(io.BootstrapOptions, factory.FromRepoURL, pipelines.NewCmdExecutor())
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
	bootstrapCmd.Flags().StringVar(&o.OutputPath, "output", ".", "Path to write GitOps resources")
	bootstrapCmd.Flags().StringVarP(&o.Prefix, "prefix", "p", "", "Add a prefix to the environment names(Dev, stage,prod,cicd etc.) to distinguish and identify individual environments")
	bootstrapCmd.Flags().StringVar(&o.DockerConfigJSONFilename, "dockercfgjson", "~/.docker/config.json", "Filepath to config.json which authenticates the image push to the desired image registry ")
	bootstrapCmd.Flags().StringVar(&o.InternalRegistryHostname, "image-repo-internal-registry-hostname", "image-registry.openshift-image-registry.svc:5000", "Host-name for internal image registry e.g. docker-registry.default.svc.cluster.local:5000, used if you are pushing your images to the internal image registry")
	bootstrapCmd.Flags().StringVar(&o.ImageRepo, "image-repo", "", "Image repository of the form <registry>/<username>/<repository> or <project>/<app> which is used to push newly built images")
	bootstrapCmd.Flags().StringVar(&o.SealedSecretsService.Namespace, "sealed-secrets-ns", secrets.SealedSecretsNS, "Namespace in which the Sealed Secrets operator is installed, automatically generated secrets are encrypted with this operator")
	bootstrapCmd.Flags().StringVar(&o.SealedSecretsService.Name, "sealed-secrets-svc", secrets.SealedSecretsController, "Name of the Sealed Secrets Services that encrypts secrets")
	bootstrapCmd.Flags().StringVar(&o.GitHostAccessToken, statustracker.CommitStatusTrackerSecret, "", "Used to authenticate repository clones, and commit-status notifications (if enabled). Access token is encrypted and stored on local file system by keyring, will be updated/reused.")
	bootstrapCmd.Flags().BoolVar(&o.Overwrite, "overwrite", false, "Overwrites previously existing GitOps configuration (if any)")
	bootstrapCmd.Flags().StringVar(&o.ServiceRepoURL, "service-repo-url", "", "Provide the URL for your Service repository e.g. https://github.com/organisation/service.git")
	bootstrapCmd.Flags().StringVar(&o.ServiceWebhookSecret, "service-webhook-secret", "", "Provide a secret that we can use to authenticate incoming hooks from your Git hosting service for the Service repository. (if not provided, it will be auto-generated)")
	bootstrapCmd.Flags().StringVar(&o.SaveTokenKeyRing, "save-token-keyring", "", "Explicitely pass this flag to update the git-host-access-token in the keyring on your local file system")
	bootstrapCmd.Flags().StringVar(&o.PrivateRepoDriver, "private-repo-driver", "", "If your Git repositories are on a custom domain, please indicate which driver to use github or gitlab")
	bootstrapCmd.Flags().BoolVar(&o.CommitStatusTracker, "commit-status-tracker", true, "Enable or disable the commit-status-tracker which reports the success/failure of your pipelineruns to GitHub/GitLab")
	bootstrapCmd.Flags().BoolVar(&o.PushToGit, "push-to-git", false, "If true, automatically creates and populates the gitops-repo-url with the generated resources")
	return bootstrapCmd
}

func nextSteps() {
	log.Success("Bootstrapped OpenShift resources successfully\n\n",
		"Next Steps:\n",
		"Please refer to https://github.com/redhat-developer/kam/tree/master/docs to get started.",
	)
}

func isKnownDriver(repoURL string) bool {
	host, err := accesstoken.HostFromURL(repoURL)
	if err != nil {
		return false
	}
	_, err = factory.DefaultIdentifier.Identify(host)
	return err == nil
}
