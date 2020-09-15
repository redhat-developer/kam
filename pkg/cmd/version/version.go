package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended command name.
const RecommendedCommandName = "version"

var Version string

// NewCmd creates a new command
func NewCmd(name, fullName string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: "Print the version information",
		Long:  "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("gitops version %s\n", Version)
		},
	}
}
