package contract

import (
	"context"
	"fmt"
	"sort"
)

// EntityAdapter is the subset of a Phase F port interface covering
// the two entity-level methods the harness exercises: Get and List.
// Both adapter shapes satisfy this via their own concrete entity type
// E and filter type F.
//
// Filter type F is generic because different ports support different
// filter shapes (class + status, class + type, class + warehouse_id,
// etc.). Only Class filter behaviour is exercised by this harness, so
// callers construct F values that set Class and leave other fields
// zero.
type EntityAdapter[E any, F any] interface {
	Get(ctx context.Context, id string) (*E, error)
	List(ctx context.Context, filter F) ([]*E, int32, error)
}

// EntitySuite wires two adapters and their projection into a
// canonical id+class shape. The harness runs:
//
//   - Get(id) through both adapters for every seeded id, compares.
//   - List with no filter through both adapters, compares ids present.
//   - List with a class filter through both adapters, compares the
//     filtered set.
//
// ProjectIdentity is a small function `func(*E) (id, class string)`.
// We only compare id+class equivalence between adapters here because:
//
//   - Other fields vary by port and the per-port test is free to add
//     extra assertions on top of this harness.
//   - id+class is the minimum surface every Phase F port has and
//     captures the most common divergence cause (class filter applied
//     differently on inproc vs connect).
//
// For deeper per-field parity, a port may supply a richer projection
// by calling the harness's DiffFn hook (below).
type EntitySuite[E any, F any] struct {
	Name string

	InProc  EntityAdapter[E, F]
	Connect EntityAdapter[E, F]

	// SeededIDs is the ordered list of entity ids both adapters return.
	// The harness calls Get on each and compares; it also asserts that
	// List (no filter) returns AT LEAST these ids in both adapters.
	SeededIDs []string

	// ProjectIdentity extracts (id, class) from a port entity.
	ProjectIdentity func(*E) (id string, class string)

	// MakeClassFilter returns a filter value F configured with the
	// provided class name. Callers construct their port-specific filter
	// (e.g. `client.ListFilter{Class: class, Limit: 1000}`) here.
	MakeClassFilter func(class string) F

	// MakeEmptyFilter returns a zero filter used for the unconstrained
	// list comparison.
	MakeEmptyFilter func() F

	// ClassFilterValue is a class name that appears on at least one of
	// the seeded entities. The harness asserts both adapters return
	// the same ids when filtered by this class.
	ClassFilterValue string

	// DiffFn is OPTIONAL. When set, the harness calls it with matched
	// (inproc, connect) entity pairs from Get calls and reports any
	// non-empty return as a diff. Tests that want full-field parity
	// past id+class wire this to a port-specific diff helper.
	DiffFn func(inproc, connect *E) string
}

// Run executes the entity contract suite.
func (s EntitySuite[E, F]) Run(t TB) {
	t.Helper()

	if s.InProc == nil {
		t.Fatalf("%s: InProc adapter is nil", s.Name)
	}
	if s.Connect == nil {
		t.Fatalf("%s: Connect adapter is nil", s.Name)
	}
	if s.ProjectIdentity == nil {
		t.Fatalf("%s: ProjectIdentity must be provided", s.Name)
	}
	if s.MakeEmptyFilter == nil {
		t.Fatalf("%s: MakeEmptyFilter must be provided", s.Name)
	}
	if len(s.SeededIDs) == 0 {
		t.Fatalf("%s: SeededIDs is empty — supply at least one seeded id", s.Name)
	}

	t.Run("Get_equivalence", func(t TB) {
		s.runGet(t)
	})
	t.Run("List_empty_filter_equivalence", func(t TB) {
		s.runListEmpty(t)
	})
	if s.MakeClassFilter != nil && s.ClassFilterValue != "" {
		t.Run("List_class_filter_equivalence", func(t TB) {
			s.runListClassFilter(t)
		})
	}
}

func (s EntitySuite[E, F]) runGet(t TB) {
	ctx := context.Background()
	for _, id := range s.SeededIDs {
		t.Run("id="+id, func(t TB) {
			inprocE, inprocErr := s.InProc.Get(ctx, id)
			connectE, connectErr := s.Connect.Get(ctx, id)

			if inprocErr != nil {
				t.Errorf("InProc.Get(%q): %v", id, inprocErr)
			}
			if connectErr != nil {
				t.Errorf("Connect.Get(%q): %v", id, connectErr)
			}
			if inprocErr != nil || connectErr != nil {
				return
			}
			if inprocE == nil {
				t.Fatalf("InProc.Get(%q) returned nil", id)
			}
			if connectE == nil {
				t.Fatalf("Connect.Get(%q) returned nil", id)
			}

			ipID, ipClass := s.ProjectIdentity(inprocE)
			cnID, cnClass := s.ProjectIdentity(connectE)
			if ipID != cnID {
				t.Errorf("Get(%q) id mismatch: inproc=%q connect=%q", id, ipID, cnID)
			}
			if ipClass != cnClass {
				t.Errorf("Get(%q) class mismatch: inproc=%q connect=%q", id, ipClass, cnClass)
			}
			if s.DiffFn != nil {
				if d := s.DiffFn(inprocE, connectE); d != "" {
					t.Errorf("Get(%q) field-diff:\n%s", id, d)
				}
			}
		})
	}
}

func (s EntitySuite[E, F]) runListEmpty(t TB) {
	ctx := context.Background()
	inprocRaw, _, inprocErr := s.InProc.List(ctx, s.MakeEmptyFilter())
	connectRaw, _, connectErr := s.Connect.List(ctx, s.MakeEmptyFilter())

	if inprocErr != nil {
		t.Errorf("InProc.List: %v", inprocErr)
	}
	if connectErr != nil {
		t.Errorf("Connect.List: %v", connectErr)
	}
	if inprocErr != nil || connectErr != nil {
		return
	}

	inprocIDs := s.idSet(inprocRaw)
	connectIDs := s.idSet(connectRaw)

	// Assert both contain every seeded id.
	for _, id := range s.SeededIDs {
		if _, ok := inprocIDs[id]; !ok {
			t.Errorf("List(empty): inproc missing seeded id %q", id)
		}
		if _, ok := connectIDs[id]; !ok {
			t.Errorf("List(empty): connect missing seeded id %q", id)
		}
	}

	// Assert symmetric difference is empty.
	for id := range inprocIDs {
		if _, ok := connectIDs[id]; !ok {
			t.Errorf("List(empty): inproc has %q but connect does not", id)
		}
	}
	for id := range connectIDs {
		if _, ok := inprocIDs[id]; !ok {
			t.Errorf("List(empty): connect has %q but inproc does not", id)
		}
	}
}

func (s EntitySuite[E, F]) runListClassFilter(t TB) {
	ctx := context.Background()
	inprocRaw, _, inprocErr := s.InProc.List(ctx, s.MakeClassFilter(s.ClassFilterValue))
	connectRaw, _, connectErr := s.Connect.List(ctx, s.MakeClassFilter(s.ClassFilterValue))

	if inprocErr != nil {
		t.Errorf("InProc.List(class=%q): %v", s.ClassFilterValue, inprocErr)
	}
	if connectErr != nil {
		t.Errorf("Connect.List(class=%q): %v", s.ClassFilterValue, connectErr)
	}
	if inprocErr != nil || connectErr != nil {
		return
	}

	inprocIDs := s.idSet(inprocRaw)
	connectIDs := s.idSet(connectRaw)

	// Symmetric check.
	ipOnly := diffIDSet(inprocIDs, connectIDs)
	cnOnly := diffIDSet(connectIDs, inprocIDs)
	if len(ipOnly) > 0 {
		t.Errorf("List(class=%q): inproc-only ids: %v", s.ClassFilterValue, ipOnly)
	}
	if len(cnOnly) > 0 {
		t.Errorf("List(class=%q): connect-only ids: %v", s.ClassFilterValue, cnOnly)
	}

	// Every returned entity must carry the filter class — otherwise the
	// filter isn't being applied the same way on both paths.
	for _, e := range inprocRaw {
		_, class := s.ProjectIdentity(e)
		if class != s.ClassFilterValue {
			t.Errorf("InProc.List(class=%q): returned entity has class=%q", s.ClassFilterValue, class)
		}
	}
	for _, e := range connectRaw {
		_, class := s.ProjectIdentity(e)
		if class != s.ClassFilterValue {
			t.Errorf("Connect.List(class=%q): returned entity has class=%q", s.ClassFilterValue, class)
		}
	}
}

func (s EntitySuite[E, F]) idSet(es []*E) map[string]struct{} {
	m := make(map[string]struct{}, len(es))
	for _, e := range es {
		if e == nil {
			continue
		}
		id, _ := s.ProjectIdentity(e)
		m[id] = struct{}{}
	}
	return m
}

func diffIDSet(a, b map[string]struct{}) []string {
	var out []string
	for id := range a {
		if _, ok := b[id]; !ok {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// String returns a multi-line summary of the suite configuration,
// useful for debugging — call t.Log(suite.String()) from a parent
// test that wires many suites.
func (s ClassRegistrySuite[S, D]) String() string {
	return fmt.Sprintf("classregistry-contract[%s] lookup=%q unknown=%q",
		s.Name, s.LookupClass, s.UnknownClass)
}

func (s EntitySuite[E, F]) String() string {
	return fmt.Sprintf("entity-contract[%s] seeded=%d class_filter=%q",
		s.Name, len(s.SeededIDs), s.ClassFilterValue)
}
