package service

import (
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"os"

	"github.com/openshift/odo/pkg/log"

	"github.com/redhat-developer/kam/pkg/cmd/genericclioptions"
	"github.com/redhat-developer/kam/pkg/pipelines/secrets"

	"github.com/redhat-developer/kam/pkg/cmd/utility"
	"github.com/redhat-developer/kam/pkg/pipelines"
	"github.com/redhat-developer/kam/pkg/pipelines/ioutils"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	addRecommendedCommandName = "add"
)

var (
	addExample = ktemplates.Examples(`	Add a Service to an environment in GitOps 
	%[1]s`)

	addLongDesc  = ktemplates.LongDesc(`Add a Service to an environment in GitOps`)
	addShortDesc = `Add a new service`
)

var (
	defaultSealedSecretsServiceName = types.NamespacedName{Namespace: secrets.SealedSecretsNS, Name: secrets.SealedSecretsController}
)

type status interface {
	WarningStatus(status string)
	Start(status string, debug bool)
	End(status bool)
}

// AddServiceOptions encapsulates the parameters for service add command
type AddServiceOptions struct {
	*pipelines.AddServiceOptions
}

// Complete is called when the command is completed
func (o *AddServiceOptions) Complete(name string, cmd *cobra.Command, args []string) error {
	o.GitRepoURL = utility.AddGitSuffixIfNecessary(o.GitRepoURL)
	return nil
}

// Validate validates the parameters of the EnvParameters.
func (o *AddServiceOptions) Validate() error {
	return nil
}

// Run runs the project bootstrap command.
func (o *AddServiceOptions) Run() error {
	client, err := utility.NewClient()
	if err := checkServiceDependencies(o, client, log.NewStatus(os.Stdout)); err != nil {
		return err
	}

	err = pipelines.AddService(o.AddServiceOptions, ioutils.NewFilesystem())

	if err != nil {
		return err
	}
	log.Successf("Created Service %s successfully at environment %s.", o.ServiceName, o.EnvName)
	return nil
}

func checkServiceDependencies(io *AddServiceOptions, client *utility.Client, spinner status) error {
	spinner.Start("Checking if Sealed Secrets is installed with default configuration", false)
	if err := checkAndSetSealedSecretsConfig(io, client, defaultSealedSecretsServiceName); err != nil {

		warnIfNotFound(spinner, "The Sealed Secrets Operator was not detected", err)
		io.Insecure = true
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to check for Sealed Secrets Operator: %w", err)
		}
		utility.DisplayUnsealedSecretsWarning()
		log.Progressf("  WARNING: Unencrypted secrets will be created in a secrets folder that is a sibling to the pipelines folder")
		log.Progressf("           Deploying this GitOps configuration without encrypting secrets is insecure and is not recommended")
	}
	return nil
}

// check and remember the given Sealed Secrets configuration if is available otherwise return the error
func checkAndSetSealedSecretsConfig(io *AddServiceOptions, client *utility.Client, sealedConfig types.NamespacedName) error {
	if err := client.CheckIfSealedSecretsExists(sealedConfig); err != nil {
		return err
	} else {
		io.SealedSecretsService = sealedConfig
	}
	return nil
}

func warnIfNotFound(spinner status, warningMsg string, err error) {
	if apierrors.IsNotFound(err) {
		spinner.WarningStatus(warningMsg)
	}
	spinner.End(false)
}

func newCmdAdd(name, fullName string) *cobra.Command {
	o := &AddServiceOptions{AddServiceOptions: &pipelines.AddServiceOptions{}}

	cmd := &cobra.Command{
		Use:     name,
		Short:   addShortDesc,
		Long:    addLongDesc,
		Example: fmt.Sprintf(addExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	cmd.Flags().StringVar(&o.GitRepoURL, "git-repo-url", "", "GitOps repository e.g. https://github.com/organisation/repository - only needed when you need to rebuild the source image for the environment")
	cmd.Flags().StringVar(&o.WebhookSecret, "webhook-secret", "", "Source Git repository webhook secret (if not provided, it will be auto-generated)")
	cmd.Flags().StringVar(&o.AppName, "app-name", "", "Name of the application where the service will be added")
	cmd.Flags().StringVar(&o.ServiceName, "service-name", "", "Name of the service to be added")
	cmd.Flags().StringVar(&o.EnvName, "env-name", "", "Name of the environment where the service will be added")
	cmd.Flags().StringVar(&o.ImageRepo, "image-repo", "", "Image registry of the form <registry>/<username>/<image name> or <project>/<app> which is used to push newly built images")
	cmd.Flags().StringVar(&o.PipelinesFolderPath, "pipelines-folder", ".", "Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml")

	cmd.Flags().StringVar(&o.SealedSecretsService.Namespace, "sealed-secrets-ns", secrets.SealedSecretsNS, "Namespace in which the Sealed Secrets operator is installed, automatically generated secrets are encrypted with this operator")
	cmd.Flags().StringVar(&o.SealedSecretsService.Name, "sealed-secrets-svc", secrets.SealedSecretsController, "Name of the Sealed Secrets services that encrypts secrets")
	cmd.Flags().BoolVar(&o.Insecure, "insecure", false, "Set to true to use unencrypted secrets instead of sealed secrets.")

	// required flags
	_ = cmd.MarkFlagRequired("service-name")
	_ = cmd.MarkFlagRequired("app-name")
	_ = cmd.MarkFlagRequired("env-name")
	return cmd
}
