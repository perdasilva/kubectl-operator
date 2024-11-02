package list

import (
	"fmt"
	"github.com/fatih/color"
	catalogdv1alpha1 "github.com/operator-framework/catalogd/api/core/v1alpha1"
	"github.com/operator-framework/kubectl-operator/internal/cmd/internal/log"
	experimentalaction "github.com/operator-framework/kubectl-operator/internal/pkg/experimental/action"
	"github.com/operator-framework/kubectl-operator/internal/pkg/experimental/util"
	"github.com/operator-framework/kubectl-operator/pkg/action"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
)

func NewListClusterCatalogsCmd(cfg *action.Configuration) *cobra.Command {
	i := experimentalaction.NewClusterCatalogList(cfg)
	i.Logf = log.Printf

	cmd := &cobra.Command{
		Use:   "catalogs",
		Short: "List cluster catalogs",
		Run: func(cmd *cobra.Command, args []string) {
			catalogs, err := i.Run(cmd.Context())
			if err != nil {
				log.Fatalf("failed to list catalogs: %v", err)
			}
			if len(catalogs) == 0 {
				log.Print("no catalogs found")
			}

			table, err := tabularize(catalogs)
			if err != nil {
				log.Fatalf("failed to render catalogs: %v", err)
			}
			table.Render()
		},
	}

	return cmd
}

func tabularize(catalogs []catalogdv1alpha1.ClusterCatalog) (*util.AsciiTable, error) {
	header := []string{"Name", "Image", "Priority", "Enabled", "Serving", "Progression", "Last Unpacked"}
	var rowStyle util.RowStylelizer = func(row util.Row) {
		// Style "Image"
		if row[1] == "" {
			row[1] = util.StyleUnknown("Unknown")
		} else {
			row[1] = util.StyleCustom(row[1], color.HiCyanString)
		}

		// Style "Enabled"
		switch row[3] {
		case catalogdv1alpha1.AvailabilityEnabled:
			row[3] = util.StyleSuccess(row[3])
		case catalogdv1alpha1.AvailabilityDisabled:
			row[3] = util.StyleBlocked(row[3])
		default:
			row[3] = util.StyleUnknown("Unknown")
		}

		// Style "Serving"
		switch row[4] {
		case catalogdv1alpha1.ReasonAvailable:
			row[4] = util.StyleSuccess(row[4])
		case catalogdv1alpha1.ReasonUnavailable:
			row[4] = util.StyleError(row[4])
		case catalogdv1alpha1.ReasonDisabled:
			row[4] = util.StyleBlocked(row[4])
		default:
			row[4] = util.StyleUnknown("Unknown")
		}

		// Style "Progression"
		switch row[5] {
		case catalogdv1alpha1.ReasonSucceeded:
			row[5] = util.StyleSuccess(row[5])
		case catalogdv1alpha1.ReasonRetrying:
			row[5] = util.StyleWarning(row[5])
		case catalogdv1alpha1.ReasonBlocked:
			row[5] = util.StyleBlocked(row[5])
		default:
			row[4] = util.StyleUnknown("Unknown")
		}
	}

	table := util.NewAsciiTable(header, util.WithRowStylizer(rowStyle))
	for _, catalog := range catalogs {
		row := util.Row{}
		row.Append(catalog.Name)
		image := ""
		if catalog.Spec.Source.Image != nil {
			image = catalog.Spec.Source.Image.Ref
		}
		row.Append(image)
		row.Append(fmt.Sprintf("%d", catalog.Spec.Priority))
		row.Append(catalog.Spec.Availability)
		serving := ""
		if condition := meta.FindStatusCondition(catalog.Status.Conditions, catalogdv1alpha1.TypeServing); condition != nil {
			serving = condition.Reason
		}
		row.Append(serving)
		progressing := ""
		if condition := meta.FindStatusCondition(catalog.Status.Conditions, catalogdv1alpha1.TypeProgressing); condition != nil {
			progressing = condition.Reason
		}
		row.Append(progressing)
		row.Append(catalog.Status.LastUnpacked.String())

		if err := table.Append(row); err != nil {
			return nil, err
		}
	}
	return table, nil
}
