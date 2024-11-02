package action

import (
	"context"
	"github.com/operator-framework/kubectl-operator/internal/pkg/experimental/util"
	"github.com/operator-framework/kubectl-operator/pkg/action"
)

type BundleList struct {
	config  *action.Configuration
	Package string
	Logf    func(string, ...interface{})
}

func NewBundleList(cfg *action.Configuration) *BundleList {
	return &BundleList{
		config: cfg,
		Logf:   func(string, ...interface{}) {},
	}
}

func (l *BundleList) Run(ctx context.Context) ([]util.LocalBundle, error) {
	return util.GetBundlesForPackage(ctx, l.config.CacheDir, l.Package)
}
