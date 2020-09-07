package cmd

import (
	"log"

	"github.com/chetan-rns/gitops-cli/pkg/cmd/environment"
	"github.com/chetan-rns/gitops-cli/pkg/cmd/service"
	"github.com/chetan-rns/gitops-cli/pkg/cmd/utility"
	"github.com/chetan-rns/gitops-cli/pkg/cmd/webhook"
	"github.com/spf13/cobra"
)

var (
	gitopsLong = "CLI tool to scaffold your GitOps repository"
	fullName   = "gitops-cli"
)

func makeRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gitops-cli",
		Short: "gitops-cli",
		Long:  gitopsLong,
	}

	// Add all subcommands to base command
	rootCmd.AddCommand(
		NewCmdBootstrap(BootstrapRecommendedCommandName, utility.GetFullName(fullName, BootstrapRecommendedCommandName)),
		environment.NewCmdEnv(environment.EnvRecommendedCommandName, utility.GetFullName(fullName, environment.EnvRecommendedCommandName)),
		service.NewCmd(service.RecommendedCommandName, utility.GetFullName(fullName, service.RecommendedCommandName)),
		webhook.NewCmdWebhook(webhook.RecommendedCommandName, utility.GetFullName(fullName, webhook.RecommendedCommandName)),
		NewCmdBuild(BuildRecommendedCommandName, utility.GetFullName(fullName, BuildRecommendedCommandName)),
	)

	return rootCmd
}

// Execute is the main entry point into this component.
func Execute() {
	if err := makeRootCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}
