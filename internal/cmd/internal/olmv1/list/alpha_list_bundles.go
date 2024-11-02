package list

import (
	"github.com/operator-framework/kubectl-operator/internal/cmd/internal/log"
	experimentalaction "github.com/operator-framework/kubectl-operator/internal/pkg/experimental/action"
	"github.com/operator-framework/kubectl-operator/pkg/action"
	"github.com/spf13/cobra"
)

func NewListBundlesCmd(cfg *action.Configuration) *cobra.Command {
	i := experimentalaction.NewBundleList(cfg)
	i.Logf = log.Printf

	cmd := &cobra.Command{
		Use:   "bundles",
		Short: "List bundles",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			i.Package = args[0]
			bundles, err := i.Run(cmd.Context())
			if err != nil {
				log.Fatalf("failed to list bundles: %v", err)
			}
			if len(bundles) == 0 {
				log.Print("no bundles found")
			}

			for _, bundle := range bundles {
				i.Logf("- %s", bundle.Name)
			}
		},
	}

	return cmd
}
