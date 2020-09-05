package pipelines

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chetan-rns/gitops-cli/pkg/cmd/pipelines/environment"
	"github.com/chetan-rns/gitops-cli/pkg/cmd/pipelines/service"
	"github.com/chetan-rns/gitops-cli/pkg/cmd/pipelines/webhook"
)

// RecommendedCommandName is the recommended pipelines command name.
const RecommendedCommandName = "pipelines"

// NewCmdPipelines implements the component odo command
func NewCmdPipelines(name, fullName string) *cobra.Command {
	var pipelinesCmd = &cobra.Command{
		Use:   name,
		Short: "Pipeline operations",
		Example: fmt.Sprintf("%s\n%s\n\n  See sub-commands individually for more examples",
			fullName, BootstrapRecommendedCommandName),
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	bootstrapCmd := NewCmdBootstrap(BootstrapRecommendedCommandName, GetFullName(fullName, BootstrapRecommendedCommandName))
	envCmd := environment.NewCmdEnv(environment.EnvRecommendedCommandName, GetFullName(fullName, environment.EnvRecommendedCommandName))
	serviceCmd := service.NewCmd(service.RecommendedCommandName, GetFullName(fullName, service.RecommendedCommandName))

	webhookCmd := webhook.NewCmdWebhook(webhook.RecommendedCommandName, GetFullName(fullName, webhook.RecommendedCommandName))

	pipelinesCmd.AddCommand(bootstrapCmd)
	pipelinesCmd.AddCommand(envCmd)
	pipelinesCmd.AddCommand(serviceCmd)
	pipelinesCmd.AddCommand(webhookCmd)

	buildCmd := NewCmdBuild(BuildRecommendedCommandName, GetFullName(fullName, BuildRecommendedCommandName))
	pipelinesCmd.AddCommand(buildCmd)

	pipelinesCmd.Annotations = map[string]string{"command": "main"}
	// pipelinesCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return pipelinesCmd
}

// GetFullName generates a command's full name based on its parent's full name and its own name
func GetFullName(parentName, name string) string {
	return parentName + " " + name
}
