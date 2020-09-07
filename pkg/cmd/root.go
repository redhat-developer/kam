package cmd

import (
	"log"

	"github.com/chetan-rns/gitops-cli/pkg/cmd/pipelines"
	"github.com/chetan-rns/gitops-cli/pkg/cmd/util"
	"github.com/spf13/cobra"
)

var (
	gitopsLong = "CLI tool to scaffold your GitOps repository"
)

func makeRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gitops-cli",
		Short: "gitops-cli",
		Long:  gitopsLong,
	}
	// Add all subcommands to base command
	rootCmd.AddCommand(
		pipelines.NewCmdPipelines(pipelines.RecommendedCommandName, util.GetFullName("gitops-cli", pipelines.RecommendedCommandName)))

	return rootCmd
}

// Execute is the main entry point into this component.
func Execute() {
	if err := makeRootCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}
