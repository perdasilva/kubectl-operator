package action

import (
	"context"
	catalogdv1alpha1 "github.com/operator-framework/catalogd/api/core/v1alpha1"
	"github.com/operator-framework/kubectl-operator/pkg/action"
)

type ClusterCatalogList struct {
	config *action.Configuration

	Package string

	Logf func(string, ...interface{})
}

func NewClusterCatalogList(cfg *action.Configuration) *ClusterCatalogList {
	return &ClusterCatalogList{
		config: cfg,
		Logf:   func(string, ...interface{}) {},
	}
}

func (l *ClusterCatalogList) Run(ctx context.Context) ([]catalogdv1alpha1.ClusterCatalog, error) {
	catalogList := catalogdv1alpha1.ClusterCatalogList{}
	err := l.config.Client.List(ctx, &catalogList)
	if err != nil {
		return nil, err
	}
	return catalogList.Items, nil
}
