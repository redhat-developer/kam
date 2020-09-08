package service

import (
	"fmt"

	"github.com/openshift/odo/pkg/log"
	"github.com/rhd-gitops-example/gitops-cli/pkg/cmd/genericclioptions"
	"github.com/rhd-gitops-example/gitops-cli/pkg/cmd/utility"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/ioutils"
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
	err := pipelines.AddService(o.AddServiceOptions, ioutils.NewFilesystem())

	if err != nil {
		return err
	}
	log.Successf("Created Service %s sucessfully at environment %s.", o.ServiceName, o.EnvName)
	return nil
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

	cmd.Flags().StringVar(&o.GitRepoURL, "git-repo-url", "", "GitOps repository e.g. https://github.com/organisation/repository")
	cmd.Flags().StringVar(&o.WebhookSecret, "webhook-secret", "", "Source Git repository webhook secret (if not provided, it will be auto-generated)")
	cmd.Flags().StringVar(&o.AppName, "app-name", "", "Name of the application where the service will be added")
	cmd.Flags().StringVar(&o.ServiceName, "service-name", "", "Name of the service to be added")
	cmd.Flags().StringVar(&o.EnvName, "env-name", "", "Name of the environment where the service will be added")
	cmd.Flags().StringVar(&o.ImageRepo, "image-repo", "", "Image repository of the form <registry>/<username>/<repository> or <project>/<app> which is used to push newly built images")
	cmd.Flags().StringVar(&o.InternalRegistryHostname, "image-repo-internal-registry-hostname", "image-registry.openshift-image-registry.svc:5000", "Host-name for internal image registry e.g. docker-registry.default.svc.cluster.local:5000, used if you are pushing your images to the internal image registry")
	cmd.Flags().StringVar(&o.PipelinesFolderPath, "pipelines-folder", ".", "Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml")

	cmd.Flags().StringVar(&o.SealedSecretsService.Namespace, "sealed-secrets-ns", "kube-system", "Namespace in which the Sealed Secrets operator is installed, automatically generated secrets are encrypted with this operator")
	cmd.Flags().StringVar(&o.SealedSecretsService.Name, "sealed-secrets-svc", "sealed-secrets-controller", "Name of the Sealed Secrets services that encrypts secrets")

	// required flags
	_ = cmd.MarkFlagRequired("service-name")
	_ = cmd.MarkFlagRequired("app-name")
	_ = cmd.MarkFlagRequired("env-name")
	return cmd
}
