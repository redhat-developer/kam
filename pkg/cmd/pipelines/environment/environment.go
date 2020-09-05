package environment

import (
	"fmt"

	"github.com/spf13/cobra"
)

// EnvRecommendedCommandName is the recommended environment command name.
const EnvRecommendedCommandName = "environment"

// NewCmdEnv create a new environment command
func NewCmdEnv(name, fullName string) *cobra.Command {

	addEnvCmd := NewCmdAddEnv(AddEnvRecommendedCommandName, GetFullName(fullName, AddEnvRecommendedCommandName))

	var envCmd = &cobra.Command{
		Use:   name,
		Short: "Manage an environment in GitOps",
		Example: fmt.Sprintf("%s\n%s\n\n  See sub-commands individually for more examples",
			fullName, AddEnvRecommendedCommandName),
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	envCmd.Flags().AddFlagSet(addEnvCmd.Flags())
	envCmd.AddCommand(addEnvCmd)

	envCmd.Annotations = map[string]string{"command": "main"}
	// envCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return envCmd
}

// GetFullName generates a command's full name based on its parent's full name and its own name
func GetFullName(parentName, name string) string {
	return parentName + " " + name
}
