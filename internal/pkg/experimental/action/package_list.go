package action

import (
	"context"
	"github.com/operator-framework/kubectl-operator/internal/pkg/experimental/util"
	"github.com/operator-framework/kubectl-operator/pkg/action"
)

type PackageList struct {
	config  *action.Configuration
	Package string
	Logf    func(string, ...interface{})
}

func NewPackageList(cfg *action.Configuration) *PackageList {
	return &PackageList{
		config: cfg,
		Logf:   func(string, ...interface{}) {},
	}
}

func (l *PackageList) Run(ctx context.Context) ([]util.LocalPackage, error) {
	catalogs, err := util.ListLocalCatalogs(ctx, l.config.CacheDir)
	if err != nil {
		return nil, err
	}
	var packageList []util.LocalPackage
	for _, catalog := range catalogs {
		pkgs, err := catalog.ListPackages(ctx)
		if err != nil {
			return nil, err
		}
		packageList = append(packageList, pkgs...)
	}
	return packageList, nil
}
