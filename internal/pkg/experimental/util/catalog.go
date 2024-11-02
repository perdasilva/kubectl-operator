package util

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/blang/semver/v4"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/operator-framework/operator-registry/pkg/registry"
	"io/fs"
	"os"
	"path"
	"regexp"
	"sort"
)

type LocalCatalog struct {
	Name     string
	Contents fs.FS
}

type PackageSearchPredicate func(packageName string) bool

func PackageNameMatches(regexp *regexp.Regexp) PackageSearchPredicate {
	return func(packageName string) bool {
		return regexp.MatchString(packageName)
	}
}

type LocalPackage struct {
	Parent   LocalCatalog
	Name     string
	Contents fs.FS
}

type LocalBundle struct {
	declcfg.Bundle
	Version string
	Parent  LocalPackage
}

type ByVersion []LocalBundle

func (a ByVersion) Len() int {
	return len(a)
}
func (a ByVersion) Less(i, j int) bool {
	v1 := semver.MustParse(a[i].Version)
	v2 := semver.MustParse(a[j].Version)
	return v1.GT(v2)
}
func (a ByVersion) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type BundleSearchPredicate func(bundle declcfg.Bundle) bool

func (p LocalPackage) ListBundles(ctx context.Context, predicates ...BundleSearchPredicate) ([]LocalBundle, error) {
	var localBundles []LocalBundle
	if err := declcfg.WalkMetasFS(ctx, p.Contents, func(path string, meta *declcfg.Meta, err error) error {
		if err != nil {
			return err
		}
		if meta.Schema != declcfg.SchemaBundle {
			return nil
		}
		var bundle declcfg.Bundle
		if err := json.Unmarshal(meta.Blob, &bundle); err != nil {
			return err
		}
		keep := true
		for _, predicate := range predicates {
			if !predicate(bundle) {
				keep = false
				break
			}
		}
		if keep {
			var version string
			for _, prop := range bundle.Properties {
				if prop.Type == registry.PackageType {
					var pkgProp registry.PackageProperty
					if err := json.Unmarshal(prop.Value, &pkgProp); err != nil {
						return err
					}
					version = pkgProp.Version
					break
				}
			}
			localBundles = append(localBundles, LocalBundle{
				Bundle:  bundle,
				Parent:  p,
				Version: version,
			})
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("error listing bundles: %w", err)
	}
	sort.Sort(ByVersion(localBundles))
	return localBundles, nil
}

func (l LocalCatalog) ListPackages(ctx context.Context, predicates ...PackageSearchPredicate) ([]LocalPackage, error) {
	entries, err := filterDirectory(ctx, l.Contents, func(entry os.DirEntry) bool {
		for _, predicate := range predicates {
			if !predicate(entry.Name()) {
				return false
			}
		}
		return true
	})

	if err != nil {
		return nil, err
	}

	var localPackages []LocalPackage
	for entry := range entries {
		packageFBC, err := fs.Sub(l.Contents, entry.Name())
		if err != nil {
			return nil, err
		}
		localPackages = append(localPackages, LocalPackage{
			Name:     entry.Name(),
			Parent:   l,
			Contents: packageFBC,
		})
	}

	return localPackages, nil
}

type CatalogSearchPredicate func(catalogName string) bool

func CatalogNameMatches(regexp regexp.Regexp) CatalogSearchPredicate {
	return func(catalogName string) bool {
		return regexp.MatchString(catalogName)
	}
}

func GetBundlesForPackage(ctx context.Context, cacheDir string, packageName string) ([]LocalBundle, error) {
	catalogs, err := ListLocalCatalogs(ctx, cacheDir)
	if err != nil {
		return nil, err
	}
	var pkg *LocalPackage
	for _, catalog := range catalogs {
		if stat, err := fs.Stat(catalog.Contents, packageName); err != nil {
			return nil, err
		} else if !stat.IsDir() {
			return nil, fmt.Errorf("package %q is not found", packageName)
		}
		// TODO: taking the first match here
		contents, err := fs.Sub(catalog.Contents, packageName)
		if err != nil {
			return nil, err
		}
		pkg = &LocalPackage{
			Parent:   catalog,
			Name:     packageName,
			Contents: contents,
		}
		break
	}
	if pkg == nil {
		return nil, fmt.Errorf("package %q not found", packageName)
	}
	bundles, err := pkg.ListBundles(ctx)

	if err != nil {
		return nil, err
	}
	return bundles, nil
}

func ListLocalCatalogs(ctx context.Context, cacheDir string, predicates ...CatalogSearchPredicate) ([]LocalCatalog, error) {
	rootDir := os.DirFS(cacheDir)
	entries, err := filterDirectory(ctx, rootDir, func(entry os.DirEntry) bool {
		for _, predicate := range predicates {
			if !predicate(entry.Name()) {
				return false
			}
		}
		return true
	})

	if err != nil {
		return nil, err
	}

	var localCatalogs []LocalCatalog
	for entry := range entries {
		subDirs, err := fs.ReadDir(rootDir, entry.Name())
		if err != nil {
			return nil, err
		}
		// TODO: what happens when we have multiple catalog versions for the same catalog
		rt := subDirs[0].Name()
		localCatalogs = append(localCatalogs, LocalCatalog{
			Name:     entry.Name(),
			Contents: os.DirFS(path.Join(cacheDir, entry.Name(), rt, "configs")),
		})
	}

	return localCatalogs, nil
}

func filterDirectory(ctx context.Context, rootDir fs.FS, filter func(entry os.DirEntry) bool) (<-chan os.DirEntry, error) {
	dirEntries, err := fs.ReadDir(rootDir, ".")
	if err != nil {
		return nil, err
	}

	out := make(chan os.DirEntry)
	go func() {
		defer close(out)
		for _, entry := range dirEntries {
			select {
			case <-ctx.Done():
				return
			default:
				if !entry.IsDir() {
					continue
				}
				if filter(entry) {
					out <- entry
				}
			}
		}
	}()
	return out, nil
}
