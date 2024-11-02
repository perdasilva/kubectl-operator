package list

import (
	"github.com/operator-framework/kubectl-operator/internal/cmd/internal/log"
	experimentalaction "github.com/operator-framework/kubectl-operator/internal/pkg/experimental/action"
	"github.com/operator-framework/kubectl-operator/internal/pkg/experimental/util"
	"github.com/operator-framework/kubectl-operator/pkg/action"
	"github.com/operator-framework/operator-controller/api/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
)

func NewListClusterExtensionsCmd(cfg *action.Configuration) *cobra.Command {
	i := experimentalaction.NewClusterExtensionList(cfg)
	i.Logf = log.Printf

	cmd := &cobra.Command{
		Use:   "extensions",
		Short: "List cluster extensions",
		Run: func(cmd *cobra.Command, args []string) {
			clusterExtensions, err := i.Run(cmd.Context())
			if err != nil {
				log.Fatalf("failed to list cluster extensions: %v", err)
			}
			if len(clusterExtensions) == 0 {
				log.Print("no cluster extensions found")
			}

			table, err := tabularizeExtensions(clusterExtensions)
			if err != nil {
				log.Fatalf("failed to render cluster extensions: %v", err)
			}
			table.Render()
		},
	}

	return cmd
}

func tabularizeExtensions(extensions []v1alpha1.ClusterExtension) (*util.AsciiTable, error) {
	header := []string{"Name", "Package", "Version", "Upgrade Policy", "Namespace", "Service Account", "Installed Bundle", "Installed", "Deprecated", "Progression"}
	var rowStyle util.RowStylelizer = func(row util.Row) {
		for idx, value := range row {
			if value == "" {
				row[idx] = util.StyleUnknown("Unknown")
			}
		}
		// Style "Installed Condition"
		switch row[7] {
		case v1alpha1.ReasonSucceeded:
			row[7] = util.StyleSuccess("Installed")
		case v1alpha1.ReasonBlocked:
			row[7] = util.StyleBlocked(row[7])
		case v1alpha1.ReasonFailed:
			row[7] = util.StyleError(row[7])
		case v1alpha1.ReasonRetrying:
			row[7] = util.StyleReconciling(row[7])
		default:
			row[7] = util.StyleUnknown("Unknown")
		}

		// Style "Deprecated"
		switch row[8] {
		case "False":
			row[8] = util.StyleSuccess("No")
		case "True":
			row[8] = util.StyleError("Yes")
		default:
			row[8] = util.StyleUnknown("Unknown")
		}

		// Style "Progression"
		switch row[9] {
		case v1alpha1.ReasonSucceeded:
			row[9] = util.StyleSuccess("Concluded")
		case v1alpha1.ReasonBlocked:
			row[9] = util.StyleBlocked(row[9])
		case v1alpha1.ReasonFailed:
			row[9] = util.StyleError(row[9])
		case v1alpha1.ReasonRetrying:
			row[9] = util.StyleReconciling(row[9])
		default:
			row[9] = util.StyleUnknown("Unknown")
		}
	}

	table := util.NewAsciiTable(
		header,
		util.WithRowStylizer(rowStyle),
	)
	for _, extension := range extensions {
		row := util.Row{}

		row.Append(extension.Name)

		if extension.Spec.Source.SourceType == v1alpha1.SourceTypeCatalog {
			row.Append(extension.Spec.Source.Catalog.PackageName)
			row.Append(extension.Spec.Source.Catalog.Version)
			row.Append(string(extension.Spec.Source.Catalog.UpgradeConstraintPolicy))
		} else {
			row.Append("", "", "")
		}

		row.Append(extension.Spec.Install.Namespace)
		row.Append(extension.Spec.Install.ServiceAccount.Name)

		if extension.Status.Install != nil {
			row.Append(extension.Status.Install.Bundle.Name)
		} else {
			row.Append("")
		}

		if condition := meta.FindStatusCondition(extension.Status.Conditions, v1alpha1.TypeInstalled); condition != nil {
			row.Append(condition.Reason)
		} else {
			row.Append("")
		}

		if condition := meta.FindStatusCondition(extension.Status.Conditions, v1alpha1.TypeDeprecated); condition != nil {
			row.Append(string(condition.Status))
		} else {
			row.Append("")
		}

		if condition := meta.FindStatusCondition(extension.Status.Conditions, v1alpha1.TypeProgressing); condition != nil {
			row.Append(condition.Reason)
		} else {
			row.Append("")
		}

		if err := table.Append(row); err != nil {
			return nil, err
		}
	}
	return table, nil
}
