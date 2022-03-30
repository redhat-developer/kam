package webhook

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/openshift/odo/pkg/log"
	"github.com/spf13/cobra"

	"github.com/redhat-developer/kam/pkg/cmd/genericclioptions"
	backend "github.com/redhat-developer/kam/pkg/pipelines/webhook"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const createRecommendedCommandName = "create"

var (
	createExample = ktemplates.Examples(`	
	# Create a new Git repository webhook.
	# Example: kam webhook create --git-host-access-token <git host access token> --env-name dev --service-name taxi
	
	%[1]s`)
)

type createOptions struct {
	options
}

// Run contains the logic for the kam command
func (o *createOptions) Run() error {
	id, err := backend.Create(o.accessToken, o.pipelinesFolderPath, o.getAppServiceNames(), o.isCICD)

	if err != nil {
		return fmt.Errorf("unable to create webhook: %v", err)
	}

	if id != "" {
		if log.IsJSON() {
			outputSuccess(id)
		} else {
			w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
			fmt.Fprintln(w, "CREATED ID")
			fmt.Fprintln(w, "==========")
			fmt.Fprintln(w, id)
			w.Flush()
		}
	}

	return nil
}

func newCmdCreate(name, fullName string) *cobra.Command {
	o := &createOptions{}
	command := &cobra.Command{
		Use:     name,
		Short:   "Create a new webhook.",
		Long:    "Create a new Git repository webhook that triggers CI/CD pipeline runs.",
		Example: fmt.Sprintf(createExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	o.setFlags(command)
	return command
}

// outputSuccess outputs a "successful" machine-readable output format in json
func outputSuccess(machineOutput interface{}) {
	printableOutput, err := marshalJSONIndented(machineOutput)

	// If we error out... there's no way to output it (since we disable logging when using -o json)
	if err != nil {
		fmt.Fprintf(log.GetStderr(), "unable to unmarshal JSON: %s\n", err.Error())
	} else {
		fmt.Fprintf(log.GetStdout(), "%s\n", string(printableOutput))
	}
}

// marshalJSONIndented returns indented json representation of obj
func marshalJSONIndented(obj interface{}) ([]byte, error) {
	return json.MarshalIndent(obj, "", "	")
}
