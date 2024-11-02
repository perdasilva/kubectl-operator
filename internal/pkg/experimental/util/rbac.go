package util

import (
	"encoding/json"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/api/rbac/v1"
	"sigs.k8s.io/yaml"
)

type BundleInstallRBACRules struct {
	clusterScopeRules   []v1.PolicyRule
	namespaceScopeRules []v1.PolicyRule
}

func (br *BundleInstallRBACRules) AddClusterScopedRule(rule v1.PolicyRule) {
	br.clusterScopeRules = append(br.clusterScopeRules, rule)
}

func (br *BundleInstallRBACRules) AddNamespaceScopedRule(rule v1.PolicyRule) {
	br.namespaceScopeRules = append(br.namespaceScopeRules, rule)
}

func flattenRule(rule v1.PolicyRule) <-chan v1.PolicyRule {
	rules := make(chan v1.PolicyRule)
	go func() {
		for _, apiGroup := range rule.APIGroups {
			for _, resource := range rule.Resources {
				c := v1.PolicyRule{
					APIGroups: []string{apiGroup},
					Resources: []string{resource},
					Verbs:     make([]string, len(rule.Verbs)),
				}
				copy(c.Verbs, rule.Verbs)
				rules <- c
			}
		}
	}()
	return rules
}

//func GenerateRBAC(bundleFS fs.FS) {
//	var resources []KubernetesResource
//	var csv KubernetesResource
//	var crds []KubernetesResource
//
//	if err := fs.WalkDir(bundleFS, ".", func(path string, d fs.DirEntry, err error) error {
//		if err != nil {
//			return err
//		}
//		if !d.IsDir() && (strings.HasSuffix(d.Name(), ".yaml") || strings.HasSuffix(d.Name(), ".yml") || strings.HasSuffix(d.Name(), ".json")) {
//			data, err := os.ReadFile(path)
//			if err != nil {
//				return fmt.Errorf("error reading file %q: %w", path, err)
//			}
//
//			if isK8s, kind, apiVersion := IsKubernetesResource(data); isK8s {
//				res := KubernetesResource{
//					FileName:   path,
//					Kind:       kind,
//					APIVersion: apiVersion,
//					Data:       data,
//				}
//
//				if res.Kind == "ClusterServiceVersion" {
//					csv = res
//				} else if res.Kind == "CustomResourceDefinition" {
//					crds = append(crds, res)
//				} else {
//					resources = append(resources, res)
//				}
//			}
//		}
//		return nil
//	}); err != nil {
//	}
//}

//func extractCRDRules(crds []KubernetesResource) ([]v1.PolicyRule, error) {
//
//}

func extractCsvRules(res KubernetesResource) ([]v1.PolicyRule, []v1.PolicyRule, error) {
	var clusterRules []v1.PolicyRule
	var namespaceRules []v1.PolicyRule
	var csv v1alpha1.ClusterServiceVersion
	if err := yaml.Unmarshal(res.Data, &csv); err != nil {
		return nil, nil, err
	}
	var depNames []string
	for _, deps := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
		depNames = append(depNames, deps.Name)
	}

	// add deployment rules
	namespaceRules = append(namespaceRules, v1.PolicyRule{
		APIGroups: []string{"apps"},
		Resources: []string{"deployments"},
		Verbs:     []string{"create", "list", "watch"},
	}, v1.PolicyRule{
		APIGroups:     []string{"apps"},
		Resources:     []string{"deployments"},
		Verbs:         []string{"get", "update", "patch", "delete"},
		ResourceNames: depNames,
	})

	var saNames []string
	for _, clusterPerms := range csv.Spec.InstallStrategy.StrategySpec.ClusterPermissions {
		saNames = append(saNames, clusterPerms.ServiceAccountName)
		clusterRules = append(clusterRules, clusterPerms.Rules...)
	}

	for _, perms := range csv.Spec.InstallStrategy.StrategySpec.Permissions {
		saNames = append(saNames, perms.ServiceAccountName)
		namespaceRules = append(namespaceRules, perms.Rules...)
	}

	// add service account rules
	namespaceRules = append(namespaceRules, v1.PolicyRule{
		APIGroups: []string{""},
		Resources: []string{"serviceaccounts"},
		Verbs:     []string{"create", "list", "watch"},
	}, v1.PolicyRule{
		APIGroups:     []string{""},
		Resources:     []string{"serviceaccounts"},
		Verbs:         []string{"get", "update", "patch", "delete"},
		ResourceNames: saNames,
	})

	return namespaceRules, clusterRules, nil
}

type KubernetesResource struct {
	FileName   string
	Kind       string
	APIVersion string
	Data       []byte
}

// IsKubernetesResource checks if a file is a Kubernetes manifest and returns the kind and apiVersion
func IsKubernetesResource(data []byte) (bool, string, string) {
	var manifest map[string]interface{}

	// Try unmarshaling as JSON
	if json.Unmarshal(data, &manifest) != nil {
		// If JSON fails, try YAML
		if yaml.Unmarshal(data, &manifest) != nil {
			return false, "", ""
		}
	}

	// Check for Kubernetes-specific fields
	apiVersion, apiOk := manifest["apiVersion"].(string)
	kind, kindOk := manifest["kind"].(string)
	if apiOk && kindOk {
		return true, kind, apiVersion
	}
	return false, "", ""
}
