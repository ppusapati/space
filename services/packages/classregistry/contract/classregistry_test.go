package contract

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// These tests verify the harness itself behaves correctly. They use
// a capturing stub (recordingTB) in place of *testing.T so the outer
// test can observe whether the harness would have reported a failure
// under each scenario.

// recordingTB implements the TB interface and records whether any
// Errorf or Fatalf call occurred. Fatalf panics via panic(fatalMark)
// to abort the calling suite, matching *testing.T.FailNow semantics;
// the panic is recovered in Run so the outer test flow is preserved.
type recordingTB struct {
	name    string
	failed  bool
	logs    []string
	subRuns []*recordingTB
}

type fatalMark string

func (r *recordingTB) Helper() {}

func (r *recordingTB) Run(name string, f func(TB)) bool {
	sub := &recordingTB{name: name}
	r.subRuns = append(r.subRuns, sub)
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				if _, ok := rec.(fatalMark); ok {
					// Fatalf-induced panic: suite aborted here. Still
					// counts as failure.
					return
				}
				panic(rec) // re-throw unexpected panics
			}
		}()
		f(sub)
	}()
	if sub.failed {
		r.failed = true
	}
	return !sub.failed
}

func (r *recordingTB) Errorf(format string, args ...any) {
	r.failed = true
	r.logs = append(r.logs, fmt.Sprintf(format, args...))
}

func (r *recordingTB) Fatalf(format string, args ...any) {
	r.failed = true
	r.logs = append(r.logs, fmt.Sprintf(format, args...))
	panic(fatalMark("fatal"))
}

func (r *recordingTB) Log(args ...any) {
	r.logs = append(r.logs, fmt.Sprint(args...))
}

func (r *recordingTB) Failed() bool { return r.failed }

// runCapturing invokes the suite body inside a panic-recover so that
// Fatalf() at the top level of the suite (before any sub-run opens)
// still returns cleanly to the caller. Sub-run Fatalfs are recovered
// inside recordingTB.Run.
func runCapturing(run func(TB)) *recordingTB {
	rec := &recordingTB{}
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(fatalMark); !ok {
					panic(r)
				}
			}
		}()
		run(rec)
	}()
	return rec
}

// assertSuitePasses drives a suite and fails the outer test if the
// harness reports any failure.
func assertSuitePasses(t *testing.T, run func(TB)) {
	t.Helper()
	rec := runCapturing(run)
	if rec.failed {
		t.Errorf("suite should pass; harness reported failure:\n  %s",
			strings.Join(rec.logs, "\n  "))
	}
}

// assertSuiteFails drives a suite and fails the outer test if the
// harness does NOT report any failure.
func assertSuiteFails(t *testing.T, run func(TB)) {
	t.Helper()
	rec := runCapturing(run)
	if !rec.failed {
		t.Errorf("suite should fail; harness reported no failure")
	}
}

// ──────────────────────────────────────────────────────────────────────
// Test stubs
// ──────────────────────────────────────────────────────────────────────

type stubSummary struct {
	name, label, desc string
	industries        []string
	hasProc           bool
}

type stubDefinition struct {
	domain, name, label string
	attrs               map[string]stubAttr
}

type stubAttr struct {
	kind     string
	required bool
	values   []string
}

type stubAdapter struct {
	classes   []*stubSummary
	schemas   map[string]*stubDefinition
	listErr   error
	lookupErr error
}

func (a *stubAdapter) ListClasses(_ context.Context) ([]*stubSummary, error) {
	return a.classes, a.listErr
}

func (a *stubAdapter) GetClassSchema(_ context.Context, class string) (*stubDefinition, error) {
	if a.lookupErr != nil {
		return nil, a.lookupErr
	}
	d, ok := a.schemas[class]
	if !ok {
		return nil, errors.New("CLASSREGISTRY_CLASS_NOT_FOUND")
	}
	return d, nil
}

func projectSummary(s *stubSummary) ClassSummary {
	return ClassSummary{
		Name: s.name, Label: s.label, Description: s.desc,
		Industries: s.industries, HasProcesses: s.hasProc,
	}
}

func projectDefinition(d *stubDefinition) ClassDefinition {
	attrs := make(map[string]AttributeDefinition, len(d.attrs))
	for k, v := range d.attrs {
		attrs[k] = AttributeDefinition{
			Kind: v.kind, Required: v.required, Values: append([]string(nil), v.values...),
		}
	}
	return ClassDefinition{
		Domain: d.domain, Name: d.name, Label: d.label, Attributes: attrs,
	}
}

// ──────────────────────────────────────────────────────────────────────
// ClassRegistrySuite regression tests
// ──────────────────────────────────────────────────────────────────────

func TestSuite_PassesWhenAdaptersAgree(t *testing.T) {
	classes := []*stubSummary{
		{name: "a", label: "A", industries: []string{"x", "y"}, hasProc: false},
		{name: "b", label: "B", industries: []string{"z"}, hasProc: true},
	}
	schemas := map[string]*stubDefinition{
		"a": {domain: "test", name: "a", label: "A", attrs: map[string]stubAttr{
			"k1": {kind: "string", required: true},
		}},
	}
	inproc := &stubAdapter{classes: classes, schemas: schemas}
	connect := &stubAdapter{classes: classes, schemas: schemas}

	suite := ClassRegistrySuite[stubSummary, stubDefinition]{
		Name:              "test-port",
		InProc:            inproc,
		Connect:           connect,
		ProjectSummary:    projectSummary,
		ProjectDefinition: projectDefinition,
		LookupClass:       "a",
		UnknownClass:      "does_not_exist",
	}
	assertSuitePasses(t, suite.Run)
}

func TestSuite_CatchesSummaryDivergence(t *testing.T) {
	aClasses := []*stubSummary{{name: "c1", label: "C1", industries: []string{"agri"}}}
	bClasses := []*stubSummary{{name: "c1", label: "C1", industries: []string{"solar"}}}
	def := &stubDefinition{domain: "test", name: "c1", label: "C1"}
	schemas := map[string]*stubDefinition{"c1": def}

	inproc := &stubAdapter{classes: aClasses, schemas: schemas}
	connect := &stubAdapter{classes: bClasses, schemas: schemas}

	suite := ClassRegistrySuite[stubSummary, stubDefinition]{
		Name:              "test-port",
		InProc:            inproc,
		Connect:           connect,
		ProjectSummary:    projectSummary,
		ProjectDefinition: projectDefinition,
		LookupClass:       "c1",
	}
	assertSuiteFails(t, suite.Run)
}

func TestSuite_CatchesMissingClass(t *testing.T) {
	aClasses := []*stubSummary{
		{name: "x", label: "X"},
		{name: "y", label: "Y"},
	}
	bClasses := []*stubSummary{
		{name: "x", label: "X"},
	}
	def := &stubDefinition{domain: "test", name: "x", label: "X"}
	schemas := map[string]*stubDefinition{"x": def}

	inproc := &stubAdapter{classes: aClasses, schemas: schemas}
	connect := &stubAdapter{classes: bClasses, schemas: schemas}

	suite := ClassRegistrySuite[stubSummary, stubDefinition]{
		Name:              "test-port",
		InProc:            inproc,
		Connect:           connect,
		ProjectSummary:    projectSummary,
		ProjectDefinition: projectDefinition,
		LookupClass:       "x",
	}
	assertSuiteFails(t, suite.Run)
}

func TestSuite_CatchesSchemaAttributeDivergence(t *testing.T) {
	classes := []*stubSummary{{name: "p", label: "P"}}
	aSchemas := map[string]*stubDefinition{
		"p": {domain: "t", name: "p", label: "P", attrs: map[string]stubAttr{
			"f1": {kind: "string", required: true},
		}},
	}
	bSchemas := map[string]*stubDefinition{
		"p": {domain: "t", name: "p", label: "P", attrs: map[string]stubAttr{
			"f1": {kind: "int", required: true},
		}},
	}
	inproc := &stubAdapter{classes: classes, schemas: aSchemas}
	connect := &stubAdapter{classes: classes, schemas: bSchemas}

	suite := ClassRegistrySuite[stubSummary, stubDefinition]{
		Name:              "test-port",
		InProc:            inproc,
		Connect:           connect,
		ProjectSummary:    projectSummary,
		ProjectDefinition: projectDefinition,
		LookupClass:       "p",
	}
	assertSuiteFails(t, suite.Run)
}

func TestSuite_CatchesOneSidedSuccess(t *testing.T) {
	def := &stubDefinition{domain: "t", name: "k", label: "K"}
	classes := []*stubSummary{{name: "k", label: "K"}}

	inproc := &stubAdapter{
		classes: classes,
		schemas: map[string]*stubDefinition{"k": def, "unknown": def},
	}
	connect := &stubAdapter{
		classes: classes,
		schemas: map[string]*stubDefinition{"k": def},
	}

	suite := ClassRegistrySuite[stubSummary, stubDefinition]{
		Name:              "test-port",
		InProc:            inproc,
		Connect:           connect,
		ProjectSummary:    projectSummary,
		ProjectDefinition: projectDefinition,
		LookupClass:       "k",
		UnknownClass:      "unknown",
	}
	assertSuiteFails(t, suite.Run)
}

func TestSuite_AllowsTransportErrorWrapping(t *testing.T) {
	def := &stubDefinition{domain: "t", name: "k", label: "K"}
	classes := []*stubSummary{{name: "k", label: "K"}}
	schemas := map[string]*stubDefinition{"k": def}

	inproc := &stubAdapter{classes: classes, schemas: schemas}
	connect := &stubAdapter{classes: classes, schemas: schemas}

	suite := ClassRegistrySuite[stubSummary, stubDefinition]{
		Name:              "test-port",
		InProc:            inproc,
		Connect:           connect,
		ProjectSummary:    projectSummary,
		ProjectDefinition: projectDefinition,
		LookupClass:       "k",
		UnknownClass:      "gone",
	}
	assertSuitePasses(t, suite.Run)
}

func TestSuite_FatalsOnNilAdapter(t *testing.T) {
	suite := ClassRegistrySuite[stubSummary, stubDefinition]{
		Name:              "test-port",
		InProc:            nil,
		Connect:           &stubAdapter{},
		ProjectSummary:    projectSummary,
		ProjectDefinition: projectDefinition,
		LookupClass:       "x",
	}
	assertSuiteFails(t, suite.Run)
}

// ──────────────────────────────────────────────────────────────────────
// Diff helper tests
// ──────────────────────────────────────────────────────────────────────

func TestDiffSummary_FormatsHumanReadable(t *testing.T) {
	a := ClassSummary{Name: "x", Label: "X", Industries: []string{"a"}}
	b := ClassSummary{Name: "x", Label: "Y", Industries: []string{"b"}}
	diff := diffSummary(a, b)
	if diff == "" {
		t.Fatal("expected non-empty diff")
	}
	if !strings.Contains(diff, "Label") || !strings.Contains(diff, "Industries") {
		t.Errorf("diff should mention the divergent fields: %s", diff)
	}
}

func TestDiffSummary_EmptyWhenEqual(t *testing.T) {
	a := ClassSummary{Name: "x", Industries: []string{"a", "b"}}
	b := ClassSummary{Name: "x", Industries: []string{"b", "a"}} // unsorted — treated as equal
	if d := diffSummary(a, b); d != "" {
		t.Errorf("expected empty diff for equal summaries; got %s", d)
	}
}

func TestDiffAttributes_CatchesOneSidedKey(t *testing.T) {
	a := map[string]AttributeDefinition{"f1": {Kind: "string"}}
	b := map[string]AttributeDefinition{"f1": {Kind: "string"}, "f2": {Kind: "int"}}
	diff := diffAttributes(a, b)
	if diff == "" || !strings.Contains(diff, "f2") {
		t.Errorf("expected diff to flag f2 as present only in connect: %s", diff)
	}
}

func TestFloatPtrEqual(t *testing.T) {
	f1 := 3.14
	f2 := 3.14
	f3 := 2.71

	cases := []struct {
		name string
		a, b *float64
		want bool
	}{
		{"both nil", nil, nil, true},
		{"a nil", nil, &f1, false},
		{"b nil", &f1, nil, false},
		{"equal values", &f1, &f2, true},
		{"different values", &f1, &f3, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := floatPtrEqual(c.a, c.b)
			if got != c.want {
				t.Errorf("floatPtrEqual(%v, %v) = %v, want %v",
					derefFloat(c.a), derefFloat(c.b), got, c.want)
			}
		})
	}
}

func TestStringSliceEqual_OrderIndependent(t *testing.T) {
	if !stringSliceEqual([]string{"a", "b", "c"}, []string{"c", "b", "a"}) {
		t.Errorf("stringSliceEqual should be order-independent")
	}
	if stringSliceEqual([]string{"a", "b"}, []string{"a", "b", "c"}) {
		t.Errorf("different length slices are not equal")
	}
	if stringSliceEqual(nil, []string{"a"}) {
		t.Errorf("nil and non-empty are not equal")
	}
	if !stringSliceEqual(nil, nil) {
		t.Errorf("nil and nil are equal")
	}
}

// ──────────────────────────────────────────────────────────────────────
// EntitySuite regression tests
// ──────────────────────────────────────────────────────────────────────

type stubEntity struct {
	id, class string
}

type stubEntityAdapter struct {
	byID map[string]*stubEntity
}

func (a *stubEntityAdapter) Get(_ context.Context, id string) (*stubEntity, error) {
	return a.byID[id], nil
}

func (a *stubEntityAdapter) List(_ context.Context, _ any) ([]*stubEntity, int32, error) {
	out := make([]*stubEntity, 0, len(a.byID))
	for _, v := range a.byID {
		out = append(out, v)
	}
	return out, int32(len(out)), nil
}

func TestEntitySuite_PassesWhenAdaptersAgree(t *testing.T) {
	data := map[string]*stubEntity{
		"1": {id: "1", class: "x"},
		"2": {id: "2", class: "y"},
	}
	inproc := &stubEntityAdapter{byID: data}
	connect := &stubEntityAdapter{byID: data}

	suite := EntitySuite[stubEntity, any]{
		Name:            "test-entity",
		InProc:          inproc,
		Connect:         connect,
		SeededIDs:       []string{"1", "2"},
		ProjectIdentity: func(v *stubEntity) (string, string) { return v.id, v.class },
		MakeEmptyFilter: func() any { return nil },
	}
	assertSuitePasses(t, suite.Run)
}

func TestEntitySuite_Get_CatchesClassDivergence(t *testing.T) {
	inproc := &stubEntityAdapter{
		byID: map[string]*stubEntity{"1": {id: "1", class: "x"}},
	}
	connect := &stubEntityAdapter{
		byID: map[string]*stubEntity{"1": {id: "1", class: "y"}},
	}

	suite := EntitySuite[stubEntity, any]{
		Name:            "test-entity",
		InProc:          inproc,
		Connect:         connect,
		SeededIDs:       []string{"1"},
		ProjectIdentity: func(v *stubEntity) (string, string) { return v.id, v.class },
		MakeEmptyFilter: func() any { return nil },
	}
	assertSuiteFails(t, suite.Run)
}

func TestEntitySuite_List_CatchesMissingRow(t *testing.T) {
	inproc := &stubEntityAdapter{
		byID: map[string]*stubEntity{
			"1": {id: "1", class: "x"},
			"2": {id: "2", class: "x"},
		},
	}
	connect := &stubEntityAdapter{
		byID: map[string]*stubEntity{
			"1": {id: "1", class: "x"},
		},
	}

	suite := EntitySuite[stubEntity, any]{
		Name:            "test-entity",
		InProc:          inproc,
		Connect:         connect,
		SeededIDs:       []string{"1"},
		ProjectIdentity: func(v *stubEntity) (string, string) { return v.id, v.class },
		MakeEmptyFilter: func() any { return nil },
	}
	assertSuiteFails(t, suite.Run)
}
