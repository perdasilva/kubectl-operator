package action

import (
	"context"
	containertypes "github.com/containers/image/v5/types"
	"github.com/go-logr/logr"
	catalogdv1alpha1 "github.com/operator-framework/catalogd/api/core/v1alpha1"
	"github.com/operator-framework/kubectl-operator/internal/pkg/experimental/catalogd/source"
	"github.com/operator-framework/kubectl-operator/pkg/action"
	"k8s.io/apimachinery/pkg/types"
)

type ClusterCatalogCache struct {
	config  *action.Configuration
	Catalog string
	Logf    func(string, ...interface{})
}

func NewClusterCatalogCache(cfg *action.Configuration) *ClusterCatalogCache {
	return &ClusterCatalogCache{
		config: cfg,
		Logf:   func(string, ...interface{}) {},
	}
}

func (l *ClusterCatalogCache) Run(ctx context.Context) (*source.Result, error) {
	catalog := catalogdv1alpha1.ClusterCatalog{}
	err := l.config.Client.Get(ctx, types.NamespacedName{Namespace: "", Name: l.Catalog}, &catalog)
	if err != nil {
		return nil, err
	}

	l.Logf("Unpacking image...")
	ir := source.ContainersImageRegistry{
		BaseCachePath: l.config.CacheDir,
		SourceContextFunc: func(logger logr.Logger) (*containertypes.SystemContext, error) {
			return &containertypes.SystemContext{}, nil
		},
	}

	result, err := ir.Unpack(ctx, &catalog)
	if err != nil {
		return nil, err
	}

	return result, nil
}
