package pipelines

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/chetan-rns/gitops-cli/pkg/cli/pipelines/ui"
	"github.com/chetan-rns/gitops-cli/pkg/cli/pipelines/utility"
	"github.com/chetan-rns/gitops-cli/pkg/genericclioptions"
	"github.com/chetan-rns/gitops-cli/pkg/pipelines"
	"github.com/chetan-rns/gitops-cli/pkg/pipelines/ioutils"
	"github.com/chetan-rns/gitops-cli/pkg/pipelines/namespaces"
	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/openshift/odo/pkg/log"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	// BootstrapRecommendedCommandName the recommended command name
	BootstrapRecommendedCommandName = "bootstrap"

	sealedSecretsName   = "sealed-secrets-controller"
	sealedSecretsNS     = "kube-system"
	argoCDNS            = "argocd"
	argoCDOperatorName  = "argocd-operator"
	argoCDServerName    = "argocd-server"
	pipelinesOperatorNS = "openshift-operators"
)

type drivers []string

var supportedDrivers = drivers{
	"github",
	"gitlab",
}

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

	bootstrapLongDesc  = ktemplates.LongDesc(`Bootstrap GitOps CI/CD Manifest`)
	bootstrapShortDesc = `Bootstrap pipelines with a starter configuration`
)

// BootstrapParameters encapsulates the parameters for the odo pipelines init command.
type BootstrapParameters struct {
	*pipelines.BootstrapOptions
	// generic context options common to all commands
	*genericclioptions.Context
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

	clientSet, err := namespaces.GetClientSet()
	if err != nil {
		return err
	}

	if io.PrivateRepoDriver != "" {
		host, err := hostFromURL(io.GitOpsRepoURL)
		if err != nil {
			return err
		}
		identifier := factory.NewDriverIdentifier(factory.Mapping(host, io.PrivateRepoDriver))
		factory.DefaultIdentifier = identifier
	}

	// ask for sealed secrets only when default is absent
	flagset := cmd.Flags()
	if flagset.NFlag() == 0 {
		err := checkBootstrapDependencies(io, clientSet, log.NewStatus(os.Stdout))
		if err != nil {
			return err
		}
		err = initiateInteractiveMode(io)
		if err != nil {
			return err
		}
	} else {
		err := nonInteractiveMode(io, clientSet)
		if err != nil {
			return err
		}
	}
	return nil
}

// nonInteractiveMode gets triggered if a flag is passed, checks for mandatory flags.
func nonInteractiveMode(io *BootstrapParameters, clientSet *kubernetes.Clientset) error {
	mandatoryFlags := map[string]string{io.ServiceRepoURL: "service-repo-url", io.GitOpsRepoURL: "gitops-repo-url", io.ImageRepo: "image-repo"}
	for key, value := range mandatoryFlags {
		if key == "" {
			return fmt.Errorf("The mandatory flag %q has not been set", value)
		}
	}
	err := checkBootstrapDependencies(io, clientSet, log.NewStatus(os.Stdout))
	if err != nil {
		return err
	}
	return nil
}

// initiateInteractiveMode starts the interactive mode impplementation if no flags are passed.
func initiateInteractiveMode(io *BootstrapParameters) error {
	// ask for sealed secrets only when default is absent
	if io.SealedSecretsService == (types.NamespacedName{}) {
		io.SealedSecretsService.Name = ui.EnterSealedSecretService(&io.SealedSecretsService)

	}
	io.GitOpsRepoURL = ui.EnterGitRepo()
	io.GitOpsRepoURL = utility.AddGitSuffixIfNecessary(io.GitOpsRepoURL)
	if !isKnownDriver(io.GitOpsRepoURL) {
		io.PrivateRepoDriver = ui.SelectPrivateRepoDriver()
		host, err := hostFromURL(io.GitOpsRepoURL)
		if err != nil {
			return fmt.Errorf("failed to parse the gitops url: %w", err)
		}
		identifier := factory.NewDriverIdentifier(factory.Mapping(host, io.PrivateRepoDriver))
		factory.DefaultIdentifier = identifier
	}
	option := ui.SelectOptionImageRepository()
	if option == "Openshift Internal repository" {
		io.InternalRegistryHostname = ui.EnterInternalRegistry()
		io.ImageRepo = ui.EnterImageRepoInternalRegistry()
	} else {
		io.DockerConfigJSONFilename = ui.EnterDockercfg()
		io.ImageRepo = ui.EnterImageRepoExternalRepository()
	}
	io.GitOpsWebhookSecret = ui.EnterGitWebhookSecret()
	io.ServiceRepoURL = ui.EnterServiceRepoURL()
	io.ServiceWebhookSecret = ui.EnterServiceWebhookSecret()
	commitStatusTrackerCheck := ui.SelectOptionCommitStatusTracker()
	if commitStatusTrackerCheck == "yes" {
		io.StatusTrackerAccessToken = ui.EnterStatusTrackerAccessToken(io.ServiceRepoURL)
	}
	io.Prefix = ui.EnterPrefix()
	io.OutputPath = ui.EnterOutputPath()
	io.Overwrite = true
	return nil
}

func checkBootstrapDependencies(io *BootstrapParameters, kubeClient kubernetes.Interface, spinner status) error {
	var errs []error
	client := utility.NewClient(kubeClient)
	log.Progressf("\nChecking dependencies\n")

	spinner.Start("Checking if Sealed Secrets is installed with the default configuration", false)
	err := client.CheckIfSealedSecretsExists(types.NamespacedName{Namespace: sealedSecretsNS, Name: sealedSecretsName})
	setSpinnerStatus(spinner, "Please install Sealed Secrets from https://github.com/bitnami-labs/sealed-secrets/releases", err)
	if err == nil {
		io.SealedSecretsService.Name = sealedSecretsName
		io.SealedSecretsService.Namespace = sealedSecretsNS
	} else if !errors.IsNotFound(err) {
		return clusterErr(err.Error())
	}

	spinner.Start("Checking if ArgoCD Operator is installed with the default configuration", false)
	err = client.CheckIfArgoCDExists(argoCDNS)
	setSpinnerStatus(spinner, "Please install ArgoCD operator from OperatorHub", err)
	if err != nil {
		if !errors.IsNotFound(err) {
			return clusterErr(err.Error())
		}
		errs = append(errs, err)
	}

	spinner.Start("Checking if OpenShift Pipelines Operator is installed with the default configuration", false)
	err = client.CheckIfPipelinesExists(pipelinesOperatorNS)
	setSpinnerStatus(spinner, "Please install OpenShift Pipelines operator from OperatorHub", err)
	if err != nil {
		if !errors.IsNotFound(err) {
			return clusterErr(err.Error())
		}
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("Failed to satisfy the required dependencies")
	}
	return nil
}

func setSpinnerStatus(spinner status, warningMsg string, err error) {
	if err != nil {
		if errors.IsNotFound(err) {
			spinner.WarningStatus(warningMsg)
		}
		spinner.End(false)
		return
	}
	spinner.End(true)
}

// Validate validates the parameters of the BootstrapParameters.
func (io *BootstrapParameters) Validate() error {
	gr, err := url.Parse(io.GitOpsRepoURL)
	if err != nil {
		return fmt.Errorf("failed to parse url %s: %w", io.GitOpsRepoURL, err)
	}

	// TODO: this won't work with GitLab as the repo can have more path elements.
	if len(utility.RemoveEmptyStrings(strings.Split(gr.Path, "/"))) != 2 {
		return fmt.Errorf("repo must be org/repo: %s", strings.Trim(gr.Path, ".git"))
	}

	if io.PrivateRepoDriver != "" {
		if !supportedDrivers.supported(io.PrivateRepoDriver) {
			return fmt.Errorf("invalid driver type: %q", io.PrivateRepoDriver)
		}
	}

	io.Prefix = utility.MaybeCompletePrefix(io.Prefix)
	io.GitOpsRepoURL = utility.AddGitSuffixIfNecessary(io.GitOpsRepoURL)
	io.ServiceRepoURL = utility.AddGitSuffixIfNecessary(io.ServiceRepoURL)

	return nil
}

// Run runs the project Bootstrap command.
func (io *BootstrapParameters) Run() error {
	err := pipelines.Bootstrap(io.BootstrapOptions, ioutils.NewFilesystem())
	if err != nil {
		return err
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
	bootstrapCmd.Flags().StringVar(&o.SealedSecretsService.Namespace, "sealed-secrets-ns", "kube-system", "Namespace in which the Sealed Secrets operator is installed, automatically generated secrets are encrypted with this operator")
	bootstrapCmd.Flags().StringVar(&o.SealedSecretsService.Name, "sealed-secrets-svc", "sealed-secrets-controller", "Name of the Sealed Secrets Services that encrypts secrets")
	bootstrapCmd.Flags().StringVar(&o.StatusTrackerAccessToken, "status-tracker-access-token", "", "Used to authenticate requests to push commit-statuses to your Git hosting service")
	bootstrapCmd.Flags().BoolVar(&o.Overwrite, "overwrite", false, "Overwrites previously existing GitOps configuration (if any)")
	bootstrapCmd.Flags().StringVar(&o.ServiceRepoURL, "service-repo-url", "", "Provide the URL for your Service repository e.g. https://github.com/organisation/service.git")
	bootstrapCmd.Flags().StringVar(&o.ServiceWebhookSecret, "service-webhook-secret", "", "Provide a secret that we can use to authenticate incoming hooks from your Git hosting service for the Service repository. (if not provided, it will be auto-generated)")
	bootstrapCmd.Flags().StringVar(&o.PrivateRepoDriver, "private-repo-driver", "", "If your Git repositories are on a custom domain, please indicate which driver to use github or gitlab")
	return bootstrapCmd
}

func nextSteps() {
	log.Success("Bootstrapped OpenShift resources sucessfully.\n",
		"Next Steps:\n",
		"Please refer to https://github.com/rhd-gitops-example/docs/ to get started.",
	)
}

func clusterErr(errMsg string) error {
	return fmt.Errorf("Couldn't connect to cluster: %s", errMsg)
}

func isKnownDriver(repoURL string) bool {
	host, err := hostFromURL(repoURL)
	if err != nil {
		return false
	}
	_, err = factory.DefaultIdentifier.Identify(host)
	if err == nil {
		return true
	}
	return false
}

func hostFromURL(s string) (string, error) {
	p, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	return strings.ToLower(p.Host), nil
}
