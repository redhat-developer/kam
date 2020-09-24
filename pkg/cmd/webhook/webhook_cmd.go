package webhook

import (
	"fmt"

	"github.com/spf13/cobra"

	backend "github.com/redhat-developer/kam/pkg/pipelines/webhook"
)

type options struct {
	accessToken         string
	envName             string
	isCICD              bool
	pipelinesFolderPath string
	serviceName         string
}

// Complete completes createOptions after they've been created
func (o *options) Complete(name string, cmd *cobra.Command, args []string) (err error) {

	return nil

}

// Validate validates the createOptions based on completed values
func (o *options) Validate() (err error) {

	if o.isCICD {
		if o.serviceName != "" || o.envName != "" {
			return fmt.Errorf("Only one of 'cicd' or 'env-name/service-name' can be specified")
		}
	} else {
		if o.serviceName == "" || o.envName == "" {
			return fmt.Errorf("One of 'cicd' or 'env-name/service-name' must be specified")
		}
	}

	return nil
}

func (o *options) setFlags(command *cobra.Command) {

	// pipeline option
	command.Flags().StringVar(&o.pipelinesFolderPath, "pipelines-folder", ".", "Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml")

	// access-token option
	command.Flags().StringVar(&o.accessToken, "access-token", "", "Access token to be used to create Git repository webhook")
	_ = command.MarkFlagRequired("access-token")

	// cicd option
	command.Flags().BoolVar(&o.isCICD, "cicd", false, "Provide this flag if the target Git repository is a CI/CD configuration repository")

	// service option
	command.Flags().StringVar(&o.serviceName, "service-name", "", "Provide service name if the target Git repository is a service's source repository.")
	command.Flags().StringVar(&o.envName, "env-name", "", "Provide environment name if the target Git repository is a service's source repository.")

}

func (o *options) getAppServiceNames() *backend.QualifiedServiceName {

	return &backend.QualifiedServiceName{
		EnvironmentName: o.envName,
		ServiceName:     o.serviceName,
	}
}
