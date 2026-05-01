package contract

import (
	"context"
	"errors"
	"path/filepath"
	"sort"
	"testing"

	"p9e.in/samavaya/packages/classregistry"
	classregistryv1 "p9e.in/samavaya/packages/classregistry/api/v1"
)

// TestTranslate_Roundtrip_AgainstLiveYAML loads a real domain yaml
// from config/class_registry/, runs the registry → proto → port →
// canonical pipeline two ways, and asserts they agree.
//
// This is the UNIVERSAL contract test — the pb_translate.go chain is
// shared by every Phase F port. If the translator drops a field, this
// test catches it once, and every port that uses the translator is
// covered transitively. Port-specific tests (adoption recipe in
// README.md) add the finer Get/List assertions.
func TestTranslate_Roundtrip_AgainstLiveYAML(t *testing.T) {
	// Walk up from the test file to the repo root so the test runs
	// regardless of the directory it's invoked from.
	root, err := filepath.Abs("../../../config/class_registry")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}

	loader := classregistry.NewLoader(root)
	reg, err := loader.Load()
	if err != nil {
		t.Fatalf("load live yaml from %q: %v", root, err)
	}

	// Pick the 'asset' domain — its yaml has ~10 classes covering the
	// full feature surface (attributes, compliance_checks,
	// capacity_metrics, industries, base-class inheritance).
	const domain = "asset"
	classes := reg.ListClasses(domain)
	if len(classes) == 0 {
		t.Skipf("no classes found for domain %q; live yaml may have changed", domain)
	}

	// In-proc adapter: reads directly from the registry.
	inproc := &registryAdapter{reg: reg, domain: domain}

	// Connect adapter: simulates the wire round-trip. The registry →
	// proto translator runs on the server side (GetClassSchema
	// handler), wire serialises, client side re-hydrates back into
	// canonical port types. We skip the serialisation step because it
	// is protobuf's responsibility; the test exercises the translator
	// itself.
	connect := &protoRoundtripAdapter{reg: reg, domain: domain}

	suite := ClassRegistrySuite[registrySummary, registryDefinition]{
		Name:              "live-yaml:" + domain,
		InProc:            inproc,
		Connect:           connect,
		ProjectSummary:    projectRegistrySummary,
		ProjectDefinition: projectRegistryDefinition,
		LookupClass:       classes[0].Name,
		UnknownClass:      "__does_not_exist__",
	}
	suite.Run(AdaptT(t))
}

// ──────────────────────────────────────────────────────────────────────
// registryAdapter — reads the registry directly, mirrors what an
// in-proc client would return after projecting through the port's
// types.go (which is a pure copy of the registry's ClassDef fields).
// ──────────────────────────────────────────────────────────────────────

type registrySummary struct {
	Name         string
	Label        string
	Description  string
	Industries   []string
	HasProcesses bool
}

type registryDefinition struct {
	Domain           string
	Name             string
	Label            string
	Description      string
	Industries       []string
	Attributes       map[string]registryAttr
	ComplianceChecks []string
	CapacityMetrics  []string
	Processes        []string
}

type registryAttr struct {
	Kind        string
	Required    bool
	Min         *float64
	Max         *float64
	Values      []string
	Lookup      string
	Pattern     string
	Description string
}

type registryAdapter struct {
	reg    classregistry.Registry
	domain string
}

func (a *registryAdapter) ListClasses(_ context.Context) ([]*registrySummary, error) {
	defs := a.reg.ListClasses(a.domain)
	out := make([]*registrySummary, 0, len(defs))
	for _, cd := range defs {
		out = append(out, &registrySummary{
			Name:         cd.Name,
			Label:        cd.Label,
			Description:  cd.Description,
			Industries:   append([]string(nil), cd.Industries...),
			HasProcesses: len(cd.Processes) > 0,
		})
	}
	return out, nil
}

func (a *registryAdapter) GetClassSchema(_ context.Context, class string) (*registryDefinition, error) {
	cd, err := a.reg.GetClass(a.domain, class)
	if err != nil {
		return nil, err
	}
	return classDefToRegistryDef(cd), nil
}

// ──────────────────────────────────────────────────────────────────────
// protoRoundtripAdapter — exercises the registry → proto → canonical
// pipeline. This is what the connect-side adapter does in production:
//  1. server: registry.GetClass → classregistry.ClassDefToPB → wire
//  2. client: wire → port's classDefinitionFromPB → canonical shape
// ──────────────────────────────────────────────────────────────────────

type protoRoundtripAdapter struct {
	reg    classregistry.Registry
	domain string
}

func (a *protoRoundtripAdapter) ListClasses(_ context.Context) ([]*registrySummary, error) {
	defs := a.reg.ListClasses(a.domain)
	out := make([]*registrySummary, 0, len(defs))
	for _, cd := range defs {
		// Server side: encode the summary fields a handler would return.
		pb := &classregistryv1.ClassSummary{
			Name:         cd.Name,
			Label:        cd.Label,
			Description:  cd.Description,
			Industries:   append([]string(nil), cd.Industries...),
			HasProcesses: len(cd.Processes) > 0,
		}
		// Client side: decode back into the port's canonical shape.
		out = append(out, &registrySummary{
			Name:         pb.GetName(),
			Label:        pb.GetLabel(),
			Description:  pb.GetDescription(),
			Industries:   append([]string(nil), pb.GetIndustries()...),
			HasProcesses: pb.GetHasProcesses(),
		})
	}
	return out, nil
}

func (a *protoRoundtripAdapter) GetClassSchema(_ context.Context, class string) (*registryDefinition, error) {
	cd, err := a.reg.GetClass(a.domain, class)
	if err != nil {
		return nil, err
	}
	// Server side: translate to proto.
	pb := classregistry.ClassDefToPB(cd)
	if pb == nil {
		return nil, errors.New("ClassDefToPB returned nil")
	}
	// Client side: decode the wire message back into the canonical
	// port shape. This mirrors every port's classDefinitionFromPB.
	attrs := make(map[string]registryAttr, len(pb.GetAttributes()))
	for name, spec := range pb.GetAttributes() {
		ra := registryAttr{
			Kind:        classregistry.AttributeKindFromPB(spec.GetKind()),
			Required:    spec.GetRequired(),
			Values:      append([]string(nil), spec.GetValues()...),
			Lookup:      spec.GetLookup(),
			Pattern:     spec.GetPattern(),
			Description: spec.GetDescription(),
		}
		if spec.GetHasMin() {
			minV := spec.GetMin()
			ra.Min = &minV
		}
		if spec.GetHasMax() {
			maxV := spec.GetMax()
			ra.Max = &maxV
		}
		attrs[name] = ra
	}
	return &registryDefinition{
		Domain:           pb.GetDomain(),
		Name:             pb.GetName(),
		Label:            pb.GetLabel(),
		Description:      pb.GetDescription(),
		Industries:       append([]string(nil), pb.GetIndustries()...),
		Attributes:       attrs,
		ComplianceChecks: append([]string(nil), pb.GetComplianceChecks()...),
		CapacityMetrics:  append([]string(nil), pb.GetCapacityMetrics()...),
		Processes:        append([]string(nil), pb.GetProcesses()...),
	}, nil
}

// classDefToRegistryDef converts a raw registry.ClassDef into the
// canonical registryDefinition shape — the same shape a Phase F
// port's classDefinitionFromRegistry produces.
func classDefToRegistryDef(cd *classregistry.ClassDef) *registryDefinition {
	if cd == nil {
		return nil
	}
	attrs := make(map[string]registryAttr, len(cd.Attributes))
	for name, spec := range cd.Attributes {
		attrs[name] = registryAttr{
			Kind:        string(spec.Kind),
			Required:    spec.Required,
			Min:         spec.Min,
			Max:         spec.Max,
			Values:      append([]string(nil), spec.Values...),
			Lookup:      spec.Lookup,
			Pattern:     spec.Pattern,
			Description: spec.Description,
		}
	}
	return &registryDefinition{
		Domain:           cd.Domain,
		Name:             cd.Name,
		Label:            cd.Label,
		Description:      cd.Description,
		Industries:       append([]string(nil), cd.Industries...),
		Attributes:       attrs,
		ComplianceChecks: append([]string(nil), cd.ComplianceChecks...),
		CapacityMetrics:  append([]string(nil), cd.CapacityMetrics...),
		Processes:        append([]string(nil), cd.Processes...),
	}
}

// ──────────────────────────────────────────────────────────────────────
// Projections into canonical comparison shapes
// ──────────────────────────────────────────────────────────────────────

func projectRegistrySummary(s *registrySummary) ClassSummary {
	return ClassSummary{
		Name:         s.Name,
		Label:        s.Label,
		Description:  s.Description,
		Industries:   append([]string(nil), s.Industries...),
		HasProcesses: s.HasProcesses,
	}
}

func projectRegistryDefinition(d *registryDefinition) ClassDefinition {
	attrs := make(map[string]AttributeDefinition, len(d.Attributes))
	for k, v := range d.Attributes {
		attrs[k] = AttributeDefinition{
			Kind:        v.Kind,
			Required:    v.Required,
			Min:         v.Min,
			Max:         v.Max,
			Values:      append([]string(nil), v.Values...),
			Lookup:      v.Lookup,
			Pattern:     v.Pattern,
			Description: v.Description,
		}
	}
	return ClassDefinition{
		Domain:           d.Domain,
		Name:             d.Name,
		Label:            d.Label,
		Description:      d.Description,
		Industries:       append([]string(nil), d.Industries...),
		Attributes:       attrs,
		ComplianceChecks: append([]string(nil), d.ComplianceChecks...),
		CapacityMetrics:  append([]string(nil), d.CapacityMetrics...),
		Processes:        append([]string(nil), d.Processes...),
	}
}

// TestTranslate_Roundtrip_AllDomains asserts the same round-trip
// against every domain yaml shipped in config/class_registry/. This
// is the safety net that catches a field added to ClassDef but
// forgotten in ClassDefToPB.
func TestTranslate_Roundtrip_AllDomains(t *testing.T) {
	root, err := filepath.Abs("../../../config/class_registry")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	loader := classregistry.NewLoader(root)
	reg, err := loader.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	// Enumerate every loaded domain by probing the registry. We don't
	// have a "list domains" method, so read the yaml directory instead.
	domains, err := listYAMLDomains(root)
	if err != nil {
		t.Fatalf("list domains: %v", err)
	}
	if len(domains) == 0 {
		t.Fatal("no domain yamls found; test infrastructure broken")
	}
	// Sort for deterministic subtest order so CI output is stable.
	sort.Strings(domains)

	for _, domain := range domains {
		domain := domain
		t.Run(domain, func(t *testing.T) {
			classes := reg.ListClasses(domain)
			if len(classes) == 0 {
				// A domain yaml with only base_classes and no leaf classes
				// is valid but has nothing to round-trip.
				t.Skipf("domain %q has no leaf classes", domain)
			}
			inproc := &registryAdapter{reg: reg, domain: domain}
			connect := &protoRoundtripAdapter{reg: reg, domain: domain}

			suite := ClassRegistrySuite[registrySummary, registryDefinition]{
				Name:              "live-yaml:" + domain,
				InProc:            inproc,
				Connect:           connect,
				ProjectSummary:    projectRegistrySummary,
				ProjectDefinition: projectRegistryDefinition,
				LookupClass:       classes[0].Name,
				UnknownClass:      "__does_not_exist__",
			}
			suite.Run(AdaptT(t))
		})
	}
}

func listYAMLDomains(root string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(root, "*.yaml"))
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		base := filepath.Base(m)
		name := base[:len(base)-len(filepath.Ext(base))]
		out = append(out, name)
	}
	return out, nil
}
