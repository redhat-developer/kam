package cmd

import (
	"log"

	"github.com/rhd-gitops-example/gitops-cli/pkg/cmd/environment"
	"github.com/rhd-gitops-example/gitops-cli/pkg/cmd/service"
	"github.com/rhd-gitops-example/gitops-cli/pkg/cmd/utility"
	"github.com/rhd-gitops-example/gitops-cli/pkg/cmd/version"
	"github.com/rhd-gitops-example/gitops-cli/pkg/cmd/webhook"
	"github.com/spf13/cobra"
)

var (
	gitopsLong = "CLI tool to scaffold your GitOps repository"
	fullName   = "gitops"
)

func makeRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gitops",
		Short: "gitops",
		Long:  gitopsLong,
	}

	// Add all subcommands to base command
	rootCmd.AddCommand(
		NewCmdBootstrap(BootstrapRecommendedCommandName, utility.GetFullName(fullName, BootstrapRecommendedCommandName)),
		environment.NewCmdEnv(environment.EnvRecommendedCommandName, utility.GetFullName(fullName, environment.EnvRecommendedCommandName)),
		service.NewCmd(service.RecommendedCommandName, utility.GetFullName(fullName, service.RecommendedCommandName)),
		version.NewCmd(version.RecommendedCommandName, utility.GetFullName(fullName, version.RecommendedCommandName)),
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
