package environment

import (
	"fmt"

	"github.com/openshift/odo/pkg/log"
	"github.com/redhat-developer/kam/pkg/cmd/genericclioptions"
	"github.com/redhat-developer/kam/pkg/pipelines"
	"github.com/redhat-developer/kam/pkg/pipelines/ioutils"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	// AddEnvRecommendedCommandName the recommended command name
	AddEnvRecommendedCommandName = "add"
)

var (
	addEnvExample = ktemplates.Examples(`
	# Add a new environment to GitOps
	# Example: kam environment add --env-name new-env --pipelines-folder <path to GitOps folder>
	
	%[1]s 
	`)

	addEnvLongDesc  = ktemplates.LongDesc(`Add a new environment to the GitOps repository`)
	addEnvShortDesc = `Add a new environment`
)

// AddEnvParameters encapsulates the parameters for the kam pipelines init command.
type AddEnvParameters struct {
	envName         string
	pipelinesFolder string
	cluster         string
}

// NewAddEnvParameters bootstraps a AddEnvParameters instance.
func NewAddEnvParameters() *AddEnvParameters {
	return &AddEnvParameters{}
}

// Complete completes AddEnvParameters after they've been created.
//
// If the prefix provided doesn't have a "-" then one is added, this makes the
// generated environment names nicer to read.
func (eo *AddEnvParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	return nil
}

// Validate validates the parameters of the EnvParameters.
func (eo *AddEnvParameters) Validate() error {
	return nil
}

// Run runs the project bootstrap command.
func (eo *AddEnvParameters) Run() error {
	options := pipelines.EnvParameters{
		EnvName:             eo.envName,
		PipelinesFolderPath: eo.pipelinesFolder,
		Cluster:             eo.cluster,
	}
	err := pipelines.AddEnv(&options, ioutils.NewFilesystem())
	if err != nil {
		return err
	}
	log.Successf("Created Environment %s successfully.", eo.envName)
	return nil
}

// NewCmdAddEnv creates the project add environment command.
func NewCmdAddEnv(name, fullName string) *cobra.Command {
	o := NewAddEnvParameters()

	addEnvCmd := &cobra.Command{
		Use:     name,
		Short:   addEnvShortDesc,
		Long:    addEnvLongDesc,
		Example: fmt.Sprintf(addEnvExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	addEnvCmd.Flags().StringVar(&o.envName, "env-name", "", "Name of the environment/namespace")
	_ = addEnvCmd.MarkFlagRequired("env-name")
	addEnvCmd.Flags().StringVar(&o.pipelinesFolder, "pipelines-folder", ".", "Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml")
	addEnvCmd.Flags().StringVar(&o.cluster, "cluster", "", "Deployment cluster e.g. https://kubernetes.local.svc")
	return addEnvCmd
}
