package olmv1

import (
	"github.com/operator-framework/kubectl-operator/internal/cmd/internal/olmv1/list"
	"github.com/operator-framework/kubectl-operator/pkg/action"
	"github.com/spf13/cobra"
)

func NewOlmv1ListCommand(cfg *action.Configuration) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List OLMv1 resources.",
		Long:  "List OLMv1 resources.",
	}

	cmd.AddCommand(
		list.NewListClusterCatalogsCmd(cfg),
		list.NewListClusterExtensionsCmd(cfg),
		list.NewListPackagesCmd(cfg),
		list.NewListBundlesCmd(cfg),
	)

	return cmd
}
