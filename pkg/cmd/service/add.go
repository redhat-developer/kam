package service

import (
	"fmt"

	"github.com/openshift/odo/pkg/log"

	"github.com/redhat-developer/kam/pkg/cmd/genericclioptions"

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
	addExample = ktemplates.Examples(`	
	# Add a Service to an environment in GitOps
	# Example: kam service add --env-name new-env --app-name app-bus --service-name bus --git-repo-url http://github.com/<your organization>/bus.git --pipelines-folder <path to GitOps file>
	
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

	log.Successf("Created Service %s successfully at environment %s.\n", o.ServiceName, o.EnvName)
	log.Info(" WARNING: Generated secrets are not encrypted. Deploying the GitOps configuration without encrypting secrets is insecure and is not recommended.\n For more information on secret management see: https://github.com/redhat-developer/kam/tree/master/docs/journey/day1#secrets\n")
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

	cmd.Flags().StringVar(&o.GitRepoURL, "git-repo-url", "", "Service repository URL e.g. https://github.com/organisation/repository - only needed when you need to rebuild the source image for the environment")
	cmd.Flags().StringVar(&o.WebhookSecret, "webhook-secret", "", "Source Git repository webhook secret (if not provided, it will be auto-generated)")
	cmd.Flags().StringVar(&o.AppName, "app-name", "", "Name of the application where the service will be added")
	cmd.Flags().StringVar(&o.ServiceName, "service-name", "", "Name of the service to be added")
	cmd.Flags().StringVar(&o.EnvName, "env-name", "", "Name of the environment where the service will be added")
	cmd.Flags().StringVar(&o.ImageRepo, "image-repo", "", "Image registry of the form <registry>/<username>/<image name> or <project>/<app> which is used to push newly built images")
	cmd.Flags().StringVar(&o.PipelinesFolderPath, "pipelines-folder", ".", "Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml")

	// required flags
	_ = cmd.MarkFlagRequired("service-name")
	_ = cmd.MarkFlagRequired("app-name")
	_ = cmd.MarkFlagRequired("env-name")
	return cmd
}
