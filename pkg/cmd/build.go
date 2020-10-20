package cmd

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
	// BuildRecommendedCommandName the recommended command name
	BuildRecommendedCommandName = "build"
)

var (
	buildExample = ktemplates.Examples(`
	# Build files from pipelines
	%[1]s 
	`)

	buildLongDesc  = ktemplates.LongDesc(`Build GitOps pipelines files, generating the ArgoCD applications and OpenShift Pipelines EventListener`)
	buildShortDesc = `Build pipelines files`
)

// BuildParameters encapsulates the parameters for the kam pipelines build command.
type BuildParameters struct {
	pipelinesFolderPath string
	output              string // path to add Gitops resources
}

// NewBuildParameters bootstraps a BuildParameters instance.
func NewBuildParameters() *BuildParameters {
	return &BuildParameters{}
}

// Complete completes BuildParameters after they've been created.
func (io *BuildParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	return nil
}

// Validate validates the parameters of the BuildParameters.
func (io *BuildParameters) Validate() error {
	return nil
}

// Run runs the project bootstrap command.
func (io *BuildParameters) Run() error {
	options := pipelines.BuildParameters{
		PipelinesFolderPath: io.pipelinesFolderPath,
		OutputPath:          io.output,
	}
	err := pipelines.BuildResources(&options, ioutils.NewFilesystem())
	if err != nil {
		return err
	}
	log.Success("Built successfully.")
	return nil
}

// NewCmdBuild creates the pipelines build command.
func NewCmdBuild(name, fullName string) *cobra.Command {
	o := NewBuildParameters()
	buildCmd := &cobra.Command{
		Use:     name,
		Short:   buildShortDesc,
		Long:    buildLongDesc,
		Example: fmt.Sprintf(buildExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	buildCmd.Flags().StringVar(&o.output, "output", ".", "Folder path to add GitOps resources")
	buildCmd.Flags().StringVar(&o.pipelinesFolderPath, "pipelines-folder", ".", "Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml")
	return buildCmd
}
