package olmv1

import (
	"github.com/operator-framework/kubectl-operator/internal/cmd/internal/log"
	experimentalaction "github.com/operator-framework/kubectl-operator/internal/pkg/experimental/action"
	"github.com/operator-framework/kubectl-operator/pkg/action"
	"github.com/spf13/cobra"
)

func NewCacheCatalogCommand(cfg *action.Configuration) *cobra.Command {
	i := experimentalaction.NewClusterCatalogCache(cfg)
	i.Logf = log.Printf

	cmd := &cobra.Command{
		Use:   "cache-catalog <catalog>",
		Short: "Cache a ClusterCatalog locally",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			i.Catalog = args[0]
			result, err := i.Run(cmd.Context())
			if err != nil {
				log.Fatalf("failed to cache catalog: %v", err)
			}
			log.Printf(result.Message)
		},
	}

	return cmd
}
