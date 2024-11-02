package list

import (
	"github.com/operator-framework/kubectl-operator/internal/cmd/internal/log"
	experimentalaction "github.com/operator-framework/kubectl-operator/internal/pkg/experimental/action"
	"github.com/operator-framework/kubectl-operator/pkg/action"
	"github.com/spf13/cobra"
)

func NewListPackagesCmd(cfg *action.Configuration) *cobra.Command {
	i := experimentalaction.NewPackageList(cfg)
	i.Logf = log.Printf

	cmd := &cobra.Command{
		Use:   "packages",
		Short: "List packages",
		Run: func(cmd *cobra.Command, args []string) {
			pkgs, err := i.Run(cmd.Context())
			if err != nil {
				log.Fatalf("failed to list packages: %v", err)
			}
			if len(pkgs) == 0 {
				log.Print("no packages found")
			}

			for _, pkg := range pkgs {
				i.Logf("- %s", pkg.Name)
			}
		},
	}

	return cmd
}
