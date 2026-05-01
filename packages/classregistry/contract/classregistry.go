package contract

import (
	"context"
	"fmt"
	"sort"
	"testing"
)

// TB is the narrow subset of testing.TB the harness needs. Using an
// interface (rather than *testing.T directly) lets the harness's own
// regression tests substitute a capturing stub so they can assert
// whether a scenario SHOULD have produced a failure, without pulling
// the whole testing framework into a test-within-a-test.
//
// *testing.T satisfies this interface natively via AdaptT; production
// callers pass their t wrapped.
type TB interface {
	Helper()
	Run(name string, f func(TB)) bool
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Log(args ...any)
	Failed() bool
}

// tbAdapter wraps a *testing.T into the TB interface by translating
// the nested callback from t.Run(string, func(*testing.T)) into
// Run(string, func(TB)).
type tbAdapter struct {
	*testing.T
}

// AdaptT lifts a *testing.T into a TB for harness use.
//
// Call site:
//
//	suite.Run(contract.AdaptT(t))
func AdaptT(t *testing.T) TB {
	t.Helper()
	return &tbAdapter{T: t}
}

func (a *tbAdapter) Run(name string, f func(TB)) bool {
	return a.T.Run(name, func(sub *testing.T) {
		f(&tbAdapter{T: sub})
	})
}

// ClassRegistryAdapter is the subset of any Phase F port interface that
// the classregistry-contract harness exercises. Both the in-process
// and ConnectRPC adapters of a port satisfy this via their own
// ListClasses / GetClassSchema methods — the harness uses small
// projections (Project* below) to normalise the concrete per-port
// types into the canonical comparison shapes.
//
// The harness takes a pair of adapters (one of each shape) plus the
// projection functions. The projections are always one-liners — each
// port's types already carry exactly these fields; the projection just
// copies them into the canonical type.
type ClassRegistryAdapter[S any, D any] interface {
	ListClasses(ctx context.Context) ([]*S, error)
	GetClassSchema(ctx context.Context, class string) (*D, error)
}

// ClassRegistrySuite wires two adapters and their projections. Both
// adapters must satisfy the same ClassRegistryAdapter shape with the
// same port-specific S (summary) and D (definition) types. The harness
// calls both adapters through the same scenarios and deep-compares the
// projected results.
type ClassRegistrySuite[S any, D any] struct {
	// Name identifies the port in test output (e.g. "business/asset/vehicle").
	Name string

	// InProc is the in-process adapter built from the internal service.
	InProc ClassRegistryAdapter[S, D]

	// Connect is the ConnectRPC adapter built from a fake/stub connect
	// client that is wired to read from the SAME underlying fixtures as
	// InProc. Test set-up is responsible for ensuring the two adapters
	// see the same data; the harness asserts they PROCESS it the same.
	Connect ClassRegistryAdapter[S, D]

	// ProjectSummary converts the port's ClassSummary type into the
	// canonical shape for comparison. Typical implementation:
	//   func(s *client.ClassSummary) ClassSummary {
	//     return ClassSummary{
	//       Name: s.Name, Label: s.Label, …,
	//     }
	//   }
	ProjectSummary func(*S) ClassSummary

	// ProjectDefinition converts the port's ClassDefinition type into
	// the canonical shape for comparison.
	ProjectDefinition func(*D) ClassDefinition

	// LookupClass is the class name to pass to GetClassSchema during
	// the "known-class" assertion. Every port's yaml has at least one
	// real class; the caller supplies one that is known to exist in
	// both adapter fixtures.
	LookupClass string

	// UnknownClass is a class name that MUST NOT exist in the registry.
	// Both adapters should return an error. The harness asserts that
	// neither adapter succeeds; it does NOT require the errors to have
	// identical text because transport-layer errors legitimately
	// differ (in-proc returns a typed CLASSREGISTRY_CLASS_NOT_FOUND;
	// Connect may wrap with a connect.CodeNotFound envelope). The
	// harness DOES require both to have failed — silent divergence
	// between "registered" and "unregistered" paths is the bug we
	// catch here.
	UnknownClass string
}

// Run executes the contract suite. It is a test driver, not an
// assertion library — it calls t.Errorf on mismatches and continues so
// a single broken port reveals every mismatched method in one test
// run, not one at a time.
func (s ClassRegistrySuite[S, D]) Run(t TB) {
	t.Helper()

	if s.InProc == nil {
		t.Fatalf("%s: InProc adapter is nil", s.Name)
	}
	if s.Connect == nil {
		t.Fatalf("%s: Connect adapter is nil", s.Name)
	}
	if s.ProjectSummary == nil || s.ProjectDefinition == nil {
		t.Fatalf("%s: projection functions must be provided", s.Name)
	}
	if s.LookupClass == "" {
		t.Fatalf("%s: LookupClass must identify an existing class in the fixtures", s.Name)
	}

	t.Run("ListClasses_equivalence", func(t TB) {
		s.runListClasses(t)
	})
	t.Run("GetClassSchema_known_equivalence", func(t TB) {
		s.runGetClassSchemaKnown(t)
	})
	if s.UnknownClass != "" {
		t.Run("GetClassSchema_unknown_both_error", func(t TB) {
			s.runGetClassSchemaUnknown(t)
		})
	}
}

func (s ClassRegistrySuite[S, D]) runListClasses(t TB) {
	ctx := context.Background()

	inprocRaw, inprocErr := s.InProc.ListClasses(ctx)
	connectRaw, connectErr := s.Connect.ListClasses(ctx)

	if inprocErr != nil {
		t.Errorf("InProc.ListClasses returned error: %v", inprocErr)
	}
	if connectErr != nil {
		t.Errorf("Connect.ListClasses returned error: %v", connectErr)
	}
	if inprocErr != nil || connectErr != nil {
		return
	}

	inproc := projectAndSort(inprocRaw, s.ProjectSummary)
	connect := projectAndSort(connectRaw, s.ProjectSummary)

	if len(inproc) != len(connect) {
		t.Errorf("ListClasses count mismatch: inproc=%d connect=%d", len(inproc), len(connect))
		// Continue — a length mismatch is informative but we still want
		// to see WHICH classes diverged.
	}

	seen := make(map[string]bool)
	for _, ip := range inproc {
		seen[ip.Name] = true
	}
	for _, cn := range connect {
		if !seen[cn.Name] {
			t.Errorf("ListClasses: class %q present in connect but absent from inproc", cn.Name)
		}
	}
	for _, ip := range inproc {
		found := false
		for _, cn := range connect {
			if cn.Name == ip.Name {
				found = true
				if diff := diffSummary(ip, cn); diff != "" {
					t.Errorf("ListClasses[%s] field mismatch: %s", ip.Name, diff)
				}
				break
			}
		}
		if !found {
			t.Errorf("ListClasses: class %q present in inproc but absent from connect", ip.Name)
		}
	}
}

func (s ClassRegistrySuite[S, D]) runGetClassSchemaKnown(t TB) {
	ctx := context.Background()

	inprocRaw, inprocErr := s.InProc.GetClassSchema(ctx, s.LookupClass)
	connectRaw, connectErr := s.Connect.GetClassSchema(ctx, s.LookupClass)

	if inprocErr != nil {
		t.Fatalf("InProc.GetClassSchema(%q): %v", s.LookupClass, inprocErr)
	}
	if connectErr != nil {
		t.Fatalf("Connect.GetClassSchema(%q): %v", s.LookupClass, connectErr)
	}
	if inprocRaw == nil {
		t.Fatalf("InProc.GetClassSchema(%q) returned nil", s.LookupClass)
	}
	if connectRaw == nil {
		t.Fatalf("Connect.GetClassSchema(%q) returned nil", s.LookupClass)
	}

	inproc := s.ProjectDefinition(inprocRaw)
	connect := s.ProjectDefinition(connectRaw)

	if diff := diffDefinition(inproc, connect); diff != "" {
		t.Errorf("GetClassSchema(%q) mismatch:\n%s", s.LookupClass, diff)
	}
}

func (s ClassRegistrySuite[S, D]) runGetClassSchemaUnknown(t TB) {
	ctx := context.Background()

	_, inprocErr := s.InProc.GetClassSchema(ctx, s.UnknownClass)
	_, connectErr := s.Connect.GetClassSchema(ctx, s.UnknownClass)

	if inprocErr == nil {
		t.Errorf("InProc.GetClassSchema(%q): expected error for unknown class, got nil", s.UnknownClass)
	}
	if connectErr == nil {
		t.Errorf("Connect.GetClassSchema(%q): expected error for unknown class, got nil", s.UnknownClass)
	}
	// We deliberately do NOT compare error messages — transport-layer
	// wrapping legitimately differs. The contract is: both paths fail.
}

// projectAndSort maps the port-specific slice through the projection
// then sorts by Name so comparison is deterministic.
func projectAndSort[S any](in []*S, project func(*S) ClassSummary) []ClassSummary {
	if in == nil {
		return nil
	}
	out := make([]ClassSummary, 0, len(in))
	for _, s := range in {
		out = append(out, project(s))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// diffSummary returns "" when the two summaries are equivalent.
// Non-empty returns are human-readable field diffs used by t.Errorf.
func diffSummary(a, b ClassSummary) string {
	var diffs []string
	if a.Name != b.Name {
		diffs = append(diffs, fmt.Sprintf("  Name: inproc=%q connect=%q", a.Name, b.Name))
	}
	if a.Label != b.Label {
		diffs = append(diffs, fmt.Sprintf("  Label: inproc=%q connect=%q", a.Label, b.Label))
	}
	if a.Description != b.Description {
		diffs = append(diffs, fmt.Sprintf("  Description: inproc=%q connect=%q", a.Description, b.Description))
	}
	if !stringSliceEqual(a.Industries, b.Industries) {
		diffs = append(diffs, fmt.Sprintf("  Industries: inproc=%v connect=%v", a.Industries, b.Industries))
	}
	if a.HasProcesses != b.HasProcesses {
		diffs = append(diffs, fmt.Sprintf("  HasProcesses: inproc=%v connect=%v", a.HasProcesses, b.HasProcesses))
	}
	if len(diffs) == 0 {
		return ""
	}
	return "\n" + joinLines(diffs)
}

// diffDefinition returns "" when the two definitions are equivalent.
func diffDefinition(a, b ClassDefinition) string {
	var diffs []string
	if a.Domain != b.Domain {
		diffs = append(diffs, fmt.Sprintf("  Domain: inproc=%q connect=%q", a.Domain, b.Domain))
	}
	if a.Name != b.Name {
		diffs = append(diffs, fmt.Sprintf("  Name: inproc=%q connect=%q", a.Name, b.Name))
	}
	if a.Label != b.Label {
		diffs = append(diffs, fmt.Sprintf("  Label: inproc=%q connect=%q", a.Label, b.Label))
	}
	if a.Description != b.Description {
		diffs = append(diffs, fmt.Sprintf("  Description: inproc=%q connect=%q", a.Description, b.Description))
	}
	if !stringSliceEqual(a.Industries, b.Industries) {
		diffs = append(diffs, fmt.Sprintf("  Industries: inproc=%v connect=%v", a.Industries, b.Industries))
	}
	if !stringSliceEqual(a.ComplianceChecks, b.ComplianceChecks) {
		diffs = append(diffs, fmt.Sprintf("  ComplianceChecks: inproc=%v connect=%v", a.ComplianceChecks, b.ComplianceChecks))
	}
	if !stringSliceEqual(a.CapacityMetrics, b.CapacityMetrics) {
		diffs = append(diffs, fmt.Sprintf("  CapacityMetrics: inproc=%v connect=%v", a.CapacityMetrics, b.CapacityMetrics))
	}
	if !stringSliceEqual(a.Processes, b.Processes) {
		diffs = append(diffs, fmt.Sprintf("  Processes: inproc=%v connect=%v", a.Processes, b.Processes))
	}
	if attrDiffs := diffAttributes(a.Attributes, b.Attributes); attrDiffs != "" {
		diffs = append(diffs, attrDiffs)
	}
	if len(diffs) == 0 {
		return ""
	}
	return joinLines(diffs)
}

func diffAttributes(a, b map[string]AttributeDefinition) string {
	var diffs []string
	allKeys := make(map[string]struct{}, len(a)+len(b))
	for k := range a {
		allKeys[k] = struct{}{}
	}
	for k := range b {
		allKeys[k] = struct{}{}
	}
	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		av, aok := a[k]
		bv, bok := b[k]
		if !aok {
			diffs = append(diffs, fmt.Sprintf("  Attribute %q: present only in connect", k))
			continue
		}
		if !bok {
			diffs = append(diffs, fmt.Sprintf("  Attribute %q: present only in inproc", k))
			continue
		}
		if attr := diffAttribute(k, av, bv); attr != "" {
			diffs = append(diffs, attr)
		}
	}
	return joinLines(diffs)
}

func diffAttribute(key string, a, b AttributeDefinition) string {
	var diffs []string
	if a.Kind != b.Kind {
		diffs = append(diffs, fmt.Sprintf("    Kind: inproc=%q connect=%q", a.Kind, b.Kind))
	}
	if a.Required != b.Required {
		diffs = append(diffs, fmt.Sprintf("    Required: inproc=%v connect=%v", a.Required, b.Required))
	}
	if !floatPtrEqual(a.Min, b.Min) {
		diffs = append(diffs, fmt.Sprintf("    Min: inproc=%v connect=%v", derefFloat(a.Min), derefFloat(b.Min)))
	}
	if !floatPtrEqual(a.Max, b.Max) {
		diffs = append(diffs, fmt.Sprintf("    Max: inproc=%v connect=%v", derefFloat(a.Max), derefFloat(b.Max)))
	}
	if !stringSliceEqual(a.Values, b.Values) {
		diffs = append(diffs, fmt.Sprintf("    Values: inproc=%v connect=%v", a.Values, b.Values))
	}
	if a.Lookup != b.Lookup {
		diffs = append(diffs, fmt.Sprintf("    Lookup: inproc=%q connect=%q", a.Lookup, b.Lookup))
	}
	if a.Pattern != b.Pattern {
		diffs = append(diffs, fmt.Sprintf("    Pattern: inproc=%q connect=%q", a.Pattern, b.Pattern))
	}
	if a.Description != b.Description {
		diffs = append(diffs, fmt.Sprintf("    Description: inproc=%q connect=%q", a.Description, b.Description))
	}
	if len(diffs) == 0 {
		return ""
	}
	return fmt.Sprintf("  Attribute %q diverges:\n%s", key, joinLines(diffs))
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	ac := append([]string(nil), a...)
	bc := append([]string(nil), b...)
	sort.Strings(ac)
	sort.Strings(bc)
	for i := range ac {
		if ac[i] != bc[i] {
			return false
		}
	}
	return true
}

func floatPtrEqual(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func derefFloat(f *float64) any {
	if f == nil {
		return "nil"
	}
	return *f
}

func joinLines(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += "\n"
		}
		out += s
	}
	return out
}
