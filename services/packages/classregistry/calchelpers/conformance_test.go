package calchelpers

import (
	"testing"

	"p9e.in/chetana/packages/classregistry"
)

// fakeDisp implements DispatcherInfo without depending on
// ClassDispatcher's generic parameters — keeps tests simple.
type fakeDisp struct {
	name    string
	classes []string
}

func (f *fakeDisp) CalculationName() string    { return f.name }
func (f *fakeDisp) SupportedClasses() []string { return f.classes }

// ───────────────────────────────────────────────────────────────────────────
// Clean — registry + dispatcher match perfectly
// ───────────────────────────────────────────────────────────────────────────

func TestConformance_Clean(t *testing.T) {
	reg := mustLoad(t, `
domain: workcenter
classes:
  mfg_discrete:
    attributes: {}
    processes: [oee_calculation]
  solar_mfg_line:
    attributes: {}
    processes: [oee_calculation]
  water_treatment:
    attributes: {}
`)
	disp := &fakeDisp{
		name:    "oee_calculation",
		classes: []string{"mfg_discrete", "solar_mfg_line"},
	}
	mismatches := ConformanceCheck(reg, disp, []string{"workcenter"})
	if len(mismatches) != 0 {
		t.Errorf("unexpected mismatches: %+v", mismatches)
	}
}

// ───────────────────────────────────────────────────────────────────────────
// Missing handler — registry declares a class but service didn't register
// ───────────────────────────────────────────────────────────────────────────

func TestConformance_MissingHandler(t *testing.T) {
	reg := mustLoad(t, `
domain: workcenter
classes:
  mfg_discrete:
    attributes: {}
    processes: [oee_calculation]
  pharma_mfg_line:
    attributes: {}
    processes: [oee_calculation]
`)
	disp := &fakeDisp{
		name:    "oee_calculation",
		classes: []string{"mfg_discrete"}, // forgot pharma
	}
	mismatches := ConformanceCheck(reg, disp, []string{"workcenter"})
	if len(mismatches) != 1 {
		t.Fatalf("got %d mismatches, want 1: %+v", len(mismatches), mismatches)
	}
	m := mismatches[0]
	if m.Kind != "missing_handler" {
		t.Errorf("kind: %q", m.Kind)
	}
	if m.Class != "pharma_mfg_line" {
		t.Errorf("class: %q", m.Class)
	}
}

// ───────────────────────────────────────────────────────────────────────────
// Stale handler — service registered a class that doesn't opt in
// ───────────────────────────────────────────────────────────────────────────

func TestConformance_StaleHandler(t *testing.T) {
	reg := mustLoad(t, `
domain: workcenter
classes:
  mfg_discrete:
    attributes: {}
    processes: [oee_calculation]
`)
	disp := &fakeDisp{
		name:    "oee_calculation",
		classes: []string{"mfg_discrete", "removed_class"},
	}
	mismatches := ConformanceCheck(reg, disp, []string{"workcenter"})
	if len(mismatches) != 1 {
		t.Fatalf("got %d mismatches, want 1: %+v", len(mismatches), mismatches)
	}
	m := mismatches[0]
	if m.Kind != "stale_handler" {
		t.Errorf("kind: %q", m.Kind)
	}
	if m.Class != "removed_class" {
		t.Errorf("class: %q", m.Class)
	}
}

// ───────────────────────────────────────────────────────────────────────────
// Both directions at once — output is stable + sorted
// ───────────────────────────────────────────────────────────────────────────

func TestConformance_BothDirections(t *testing.T) {
	reg := mustLoad(t, `
domain: workcenter
classes:
  a:
    attributes: {}
    processes: [oee_calculation]
  b:
    attributes: {}
    processes: [oee_calculation]
`)
	disp := &fakeDisp{
		name:    "oee_calculation",
		classes: []string{"a", "c"}, // b missing, c stale
	}
	mismatches := ConformanceCheck(reg, disp, []string{"workcenter"})
	if len(mismatches) != 2 {
		t.Fatalf("got %d, want 2: %+v", len(mismatches), mismatches)
	}
	// Sorted: missing_handler before stale_handler alphabetically.
	if mismatches[0].Kind != "missing_handler" || mismatches[0].Class != "b" {
		t.Errorf("unexpected 0: %+v", mismatches[0])
	}
	if mismatches[1].Kind != "stale_handler" || mismatches[1].Class != "c" {
		t.Errorf("unexpected 1: %+v", mismatches[1])
	}
}

// ───────────────────────────────────────────────────────────────────────────
// Cross-domain — calculation relevant to multiple domains
// ───────────────────────────────────────────────────────────────────────────

func TestConformance_MultipleDomains(t *testing.T) {
	reg, err := classregistry.LoadBytes(map[string][]byte{
		"bom": []byte(`domain: bom
classes:
  eng_bom:
    attributes: {}
    processes: [cost_rollup]
`),
		"inventory": []byte(`domain: inventory
classes:
  finished_goods:
    attributes: {}
    processes: [cost_rollup]
`),
	})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	disp := &fakeDisp{
		name:    "cost_rollup",
		classes: []string{"eng_bom", "finished_goods"},
	}
	mismatches := ConformanceCheck(reg, disp, []string{"bom", "inventory"})
	if len(mismatches) != 0 {
		t.Errorf("unexpected: %+v", mismatches)
	}
}

// ───────────────────────────────────────────────────────────────────────────
// helpers
// ───────────────────────────────────────────────────────────────────────────

func mustLoad(t *testing.T, yaml string) classregistry.Registry {
	t.Helper()
	reg, err := classregistry.LoadBytes(map[string][]byte{"workcenter": []byte(yaml)})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return reg
}
