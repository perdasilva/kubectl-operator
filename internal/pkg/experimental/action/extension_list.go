package action

import (
	"context"
	"github.com/operator-framework/kubectl-operator/pkg/action"
	"github.com/operator-framework/operator-controller/api/v1alpha1"
)

type ClusterExtensionList struct {
	config *action.Configuration

	Package string

	Logf func(string, ...interface{})
}

func NewClusterExtensionList(cfg *action.Configuration) *ClusterExtensionList {
	return &ClusterExtensionList{
		config: cfg,
		Logf:   func(string, ...interface{}) {},
	}
}

func (l *ClusterExtensionList) Run(ctx context.Context) ([]v1alpha1.ClusterExtension, error) {
	extensionList := v1alpha1.ClusterExtensionList{}
	err := l.config.Client.List(ctx, &extensionList)
	if err != nil {
		return nil, err
	}
	return extensionList.Items, nil
}
