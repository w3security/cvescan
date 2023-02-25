package cyclonedx

import (
	"bytes"
	"errors"
	"sort"
	"strconv"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/samber/lo"
	"golang.org/x/exp/maps"
	"golang.org/x/xerrors"

	ftypes "github.com/w3security/cvescan/pkg/fanal/types"
	"github.com/w3security/cvescan/pkg/log"
	"github.com/w3security/cvescan/pkg/purl"
	"github.com/w3security/cvescan/pkg/types"
)

var (
	ErrPURLEmpty = errors.New("purl empty error")
)

type CycloneDX struct {
	*types.SBOM

	dependencies map[string][]string
	components   map[string]cdx.Component
}

func (c *CycloneDX) UnmarshalJSON(b []byte) error {
	log.Logger.Debug("Unmarshaling CycloneDX JSON...")
	if c.SBOM == nil {
		c.SBOM = &types.SBOM{}
	}
	bom := cdx.NewBOM()
	decoder := cdx.NewBOMDecoder(bytes.NewReader(b), cdx.BOMFileFormatJSON)
	if err := decoder.Decode(bom); err != nil {
		return xerrors.Errorf("CycloneDX decode error: %w", err)
	}

	if !isTrivySBOM(bom) {
		log.Logger.Warnf("Third-party SBOM may lead to inaccurate vulnerability detection")
		log.Logger.Warnf("Recommend using cvescan to generate SBOMs")
	}

	if err := c.parseSBOM(bom); err != nil {
		return xerrors.Errorf("failed to parse sbom: %w", err)
	}

	sort.Slice(c.Applications, func(i, j int) bool {
		if c.Applications[i].Type != c.Applications[j].Type {
			return c.Applications[i].Type < c.Applications[j].Type
		}
		return c.Applications[i].FilePath < c.Applications[j].FilePath
	})

	var metadata ftypes.Metadata
	if bom.Metadata != nil {
		metadata.Timestamp = bom.Metadata.Timestamp
		if bom.Metadata.Component != nil {
			metadata.Component = toTrivyCdxComponent(lo.FromPtr(bom.Metadata.Component))
		}
	}

	var components []ftypes.Component
	for _, component := range lo.FromPtr(bom.Components) {
		components = append(components, toTrivyCdxComponent(component))
	}

	// Keep the original SBOM
	c.CycloneDX = &ftypes.CycloneDX{
		BOMFormat:    bom.BOMFormat,
		SpecVersion:  ftypes.SpecVersion(bom.SpecVersion),
		SerialNumber: bom.SerialNumber,
		Version:      bom.Version,
		Metadata:     metadata,
		Components:   components,
	}
	return nil
}

func (c *CycloneDX) parseSBOM(bom *cdx.BOM) error {
	c.dependencies = dependencyMap(bom.Dependencies)
	c.components = componentMap(bom.Metadata, bom.Components)
	var seen = make(map[string]struct{})
	for bomRef := range c.dependencies {
		component := c.components[bomRef]
		switch component.Type {
		case cdx.ComponentTypeOS: // OS info and OS packages
			seen[component.BOMRef] = struct{}{}
			c.OS = toOS(component)
			pkgInfo, err := c.parseOSPkgs(component, seen)
			if err != nil {
				return xerrors.Errorf("failed to parse os packages: %w", err)
			}
			c.Packages = append(c.Packages, pkgInfo)
		case cdx.ComponentTypeApplication: // It would be a lock file in a CycloneDX report generated by cvescan
			if lookupProperty(component.Properties, PropertyType) == "" {
				continue
			}
			app, err := c.parseLangPkgs(component, seen)
			if err != nil {
				return xerrors.Errorf("failed to parse language packages: %w", err)
			}
			c.Applications = append(c.Applications, *app)
		case cdx.ComponentTypeLibrary:
			// It is an individual package not associated with any lock files and should be processed later.
			// e.g. .gemspec, .egg and .wheel
			continue
		}
	}

	var libComponents []cdx.Component
	for ref, component := range c.components {
		if _, ok := seen[ref]; ok {
			continue
		}
		if component.Type == cdx.ComponentTypeLibrary {
			libComponents = append(libComponents, component)
		}

		// For third-party SBOMs.
		// If there are no operating-system dependent libraries, make them implicitly dependent.
		if component.Type == cdx.ComponentTypeOS {
			if lo.IsNotEmpty(c.OS) {
				return xerrors.New("multiple OSes are not supported")
			}
			c.OS = toOS(component)
		}
	}

	pkgInfos, aggregatedApps, err := aggregatePkgs(libComponents)
	if err != nil {
		return xerrors.Errorf("failed to aggregate packages: %w", err)
	}

	// For third party SBOMs.
	// If a package that depends on the operating-system did not exist,
	// but an os package is found during aggregate, it is used.
	if len(c.Packages) == 0 && len(pkgInfos) != 0 {
		if !c.OS.Detected() {
			log.Logger.Warnf("Ignore the OS package as no OS information is found.")
		} else {
			c.Packages = pkgInfos
		}
	}
	c.Applications = append(c.Applications, aggregatedApps...)

	return nil
}

func (c *CycloneDX) parseOSPkgs(component cdx.Component, seen map[string]struct{}) (ftypes.PackageInfo, error) {
	components := c.walkDependencies(component.BOMRef)
	pkgs, err := parsePkgs(components, seen)
	if err != nil {
		return ftypes.PackageInfo{}, xerrors.Errorf("failed to parse os package: %w", err)
	}

	return ftypes.PackageInfo{
		Packages: pkgs,
	}, nil
}

func (c *CycloneDX) parseLangPkgs(component cdx.Component, seen map[string]struct{}) (*ftypes.Application, error) {
	components := c.walkDependencies(component.BOMRef)
	components = lo.UniqBy(components, func(c cdx.Component) string {
		return c.BOMRef
	})

	app := toApplication(component)
	pkgs, err := parsePkgs(components, seen)
	if err != nil {
		return nil, xerrors.Errorf("failed to parse language-specific packages: %w", err)
	}
	app.Libraries = pkgs

	return app, nil
}

func parsePkgs(components []cdx.Component, seen map[string]struct{}) ([]ftypes.Package, error) {
	var pkgs []ftypes.Package
	for _, com := range components {
		seen[com.BOMRef] = struct{}{}
		_, _, pkg, err := toPackage(com)
		if err != nil {
			if errors.Is(err, ErrPURLEmpty) {
				continue
			}
			return nil, xerrors.Errorf("failed to parse language package: %w", err)
		}
		pkgs = append(pkgs, *pkg)
	}
	return pkgs, nil
}

// walkDependencies takes all nested dependencies of the root component.
func (c *CycloneDX) walkDependencies(rootRef string) []cdx.Component {
	// e.g. Library A, B, C, D and E will be returned as dependencies of Application 1.
	// type: Application 1
	//   - type: Library A
	//     - type: Library B
	//   - type: Application 2
	//     - type: Library C
	//     - type: Application 3
	//       - type: Library D
	//       - type: Library E
	var components []cdx.Component
	for _, dep := range c.dependencies[rootRef] {
		component, ok := c.components[dep]
		if !ok {
			continue
		}

		// Take only 'Libraries'
		if component.Type == cdx.ComponentTypeLibrary {
			components = append(components, component)
		}

		components = append(components, c.walkDependencies(dep)...)
	}
	return components
}

func componentMap(metadata *cdx.Metadata, components *[]cdx.Component) map[string]cdx.Component {
	cmap := make(map[string]cdx.Component)

	for _, component := range lo.FromPtr(components) {
		cmap[component.BOMRef] = component
	}
	if metadata != nil && metadata.Component != nil {
		cmap[metadata.Component.BOMRef] = *metadata.Component
	}
	return cmap
}

func dependencyMap(deps *[]cdx.Dependency) map[string][]string {
	depMap := make(map[string][]string)

	for _, dep := range lo.FromPtr(deps) {
		if _, ok := depMap[dep.Ref]; ok {
			continue
		}
		var refs []string
		if dep.Dependencies != nil {
			refs = append(refs, *dep.Dependencies...)
		}

		depMap[dep.Ref] = refs
	}
	return depMap
}

func aggregatePkgs(libs []cdx.Component) ([]ftypes.PackageInfo, []ftypes.Application, error) {
	osPkgMap := map[string]ftypes.Packages{}
	langPkgMap := map[string]ftypes.Packages{}
	for _, lib := range libs {
		isOSPkg, pkgType, pkg, err := toPackage(lib)
		if err != nil {
			if errors.Is(err, ErrPURLEmpty) {
				continue
			}
			return nil, nil, xerrors.Errorf("failed to parse the component: %w", err)
		}

		if isOSPkg {
			osPkgMap[pkgType] = append(osPkgMap[pkgType], *pkg)
		} else {
			langPkgMap[pkgType] = append(langPkgMap[pkgType], *pkg)
		}
	}

	if len(osPkgMap) > 1 {
		return nil, nil, xerrors.Errorf("multiple types of OS packages in SBOM are not supported (%q)",
			maps.Keys(osPkgMap))
	}

	var osPkgs ftypes.PackageInfo
	for _, pkgs := range osPkgMap {
		// Just take the first element
		sort.Sort(pkgs)
		osPkgs = ftypes.PackageInfo{Packages: pkgs}
		break
	}

	var apps []ftypes.Application
	for pkgType, pkgs := range langPkgMap {
		sort.Sort(pkgs)
		apps = append(apps, ftypes.Application{
			Type:      pkgType,
			Libraries: pkgs,
		})
	}
	return []ftypes.PackageInfo{osPkgs}, apps, nil
}

func toOS(component cdx.Component) ftypes.OS {
	return ftypes.OS{
		Family: component.Name,
		Name:   component.Version,
	}
}

func toApplication(component cdx.Component) *ftypes.Application {
	return &ftypes.Application{
		Type:     lookupProperty(component.Properties, PropertyType),
		FilePath: component.Name,
	}
}

func toPackage(component cdx.Component) (bool, string, *ftypes.Package, error) {
	if component.PackageURL == "" {
		log.Logger.Warnf("Skip the component (BOM-Ref: %s) as the PURL is empty", component.BOMRef)
		return false, "", nil, ErrPURLEmpty
	}
	p, err := purl.FromString(component.PackageURL)
	if err != nil {
		return false, "", nil, xerrors.Errorf("failed to parse purl: %w", err)
	}

	pkg := p.Package()
	pkg.Ref = component.BOMRef

	for _, license := range lo.FromPtr(component.Licenses) {
		pkg.Licenses = append(pkg.Licenses, license.Expression)
	}

	for _, prop := range lo.FromPtr(component.Properties) {
		if strings.HasPrefix(prop.Name, Namespace) {
			switch strings.TrimPrefix(prop.Name, Namespace) {
			case PropertyPkgID:
				pkg.ID = prop.Value
			case PropertySrcName:
				pkg.SrcName = prop.Value
			case PropertySrcVersion:
				pkg.SrcVersion = prop.Value
			case PropertySrcRelease:
				pkg.SrcRelease = prop.Value
			case PropertySrcEpoch:
				pkg.SrcEpoch, err = strconv.Atoi(prop.Value)
				if err != nil {
					return false, "", nil, xerrors.Errorf("failed to parse source epoch: %w", err)
				}
			case PropertyModularitylabel:
				pkg.Modularitylabel = prop.Value
			case PropertyLayerDiffID:
				pkg.Layer.DiffID = prop.Value
			}
		}
	}

	isOSPkg := p.IsOSPkg()
	if isOSPkg {
		// Fill source package information for components in third-party SBOMs .
		if pkg.SrcName == "" {
			pkg.SrcName = pkg.Name
		}
		if pkg.SrcVersion == "" {
			pkg.SrcVersion = pkg.Version
		}
		if pkg.SrcRelease == "" {
			pkg.SrcRelease = pkg.Release
		}
		if pkg.SrcEpoch == 0 {
			pkg.SrcEpoch = pkg.Epoch
		}
	}

	return isOSPkg, p.PackageType(), pkg, nil
}

func toTrivyCdxComponent(component cdx.Component) ftypes.Component {
	return ftypes.Component{
		BOMRef:     component.BOMRef,
		MIMEType:   component.MIMEType,
		Type:       ftypes.ComponentType(component.Type),
		Name:       component.Name,
		Version:    component.Version,
		PackageURL: component.PackageURL,
	}
}

func lookupProperty(properties *[]cdx.Property, key string) string {
	for _, p := range lo.FromPtr(properties) {
		if p.Name == Namespace+key {
			return p.Value
		}
	}
	return ""
}

func isTrivySBOM(c *cdx.BOM) bool {
	if c == nil || c.Metadata == nil || c.Metadata.Tools == nil {
		return false
	}

	for _, tool := range *c.Metadata.Tools {
		if tool.Vendor == ToolVendor && tool.Name == ToolName {
			return true
		}
	}
	return false
}
