package calchelpers

import (
	"fmt"
	"sort"
	"strings"

	"p9e.in/chetana/packages/classregistry"
)

// DispatcherInfo is the narrow interface test helpers need from any
// ClassDispatcher[Req, Resp] instance — just the calculation name
// and the list of supported classes. Keeps the helpers generic-free
// at the call site.
type DispatcherInfo interface {
	CalculationName() string
	SupportedClasses() []string
}

// ConformanceMismatch is one diagnostic produced by ConformanceCheck.
// Kind is "missing_handler" (the class registry says a class uses
// the calculation but no handler is registered) or "stale_handler"
// (a handler is registered for a class that doesn't opt into the
// calculation).
type ConformanceMismatch struct {
	Kind   string // "missing_handler" | "stale_handler"
	Domain string
	Class  string
	Detail string
}

func (m ConformanceMismatch) String() string {
	return fmt.Sprintf("%s  %s/%s — %s", m.Kind, m.Domain, m.Class, m.Detail)
}

// ConformanceCheck cross-checks a calculation's ClassDispatcher
// against the class registry. For every (domain, class) pair whose
// registry entry lists the calculation in `processes:`, the
// dispatcher must have a handler. Conversely, every class the
// dispatcher registers handlers for must opt into the calculation
// in at least one domain.
//
// Returns nil slice when everything matches.
//
// The intended use is a single _test.go file per calculation service
// (e.g. business/manufacturing/oee/conformance_test.go) that loads
// the registry, constructs the service, and calls this function.
// The test fails when the two sides drift — either because a new
// class declared `processes: [oee_calculation]` without the service
// adding a handler, or because a handler was added for a class that
// doesn't actually use the calculation.
//
// Scope: `domains` is the list of domains the calculation is
// relevant to. A single calculation rarely spans more than one
// domain (OEE is a work-center concept; progress_billing is a
// project concept), but some cross-domain calculations exist (cost
// rollup spans bom + inventory). Tests pass every relevant domain.
func ConformanceCheck(
	reg classregistry.Registry,
	disp DispatcherInfo,
	domains []string,
) []ConformanceMismatch {
	calcName := disp.CalculationName()
	registered := toSet(disp.SupportedClasses())

	// Forward: every class in the registry that lists this calculation
	// in `processes:` must have a handler.
	expected := map[string]string{} // class → domain (first domain that declared it)
	for _, domain := range domains {
		for _, cd := range reg.ListClasses(domain) {
			if containsString(cd.Processes, calcName) {
				if _, seen := expected[cd.Name]; !seen {
					expected[cd.Name] = domain
				}
			}
		}
	}

	var mismatches []ConformanceMismatch

	for class, domain := range expected {
		if _, ok := registered[class]; !ok {
			mismatches = append(mismatches, ConformanceMismatch{
				Kind:   "missing_handler",
				Domain: domain,
				Class:  class,
				Detail: fmt.Sprintf("class registry declares processes:[%s] but the service has no handler registered — add disp.Register(%q, fn)", calcName, class),
			})
		}
	}

	// Reverse: every registered handler must point at a class that
	// opts into the calculation in at least one of the given domains.
	for class := range registered {
		if _, ok := expected[class]; !ok {
			mismatches = append(mismatches, ConformanceMismatch{
				Kind:   "stale_handler",
				Domain: strings.Join(domains, ","),
				Class:  class,
				Detail: fmt.Sprintf("handler registered but no class across %v declares processes:[%s] — remove the Register call or add the class entry", domains, calcName),
			})
		}
	}

	sort.Slice(mismatches, func(i, j int) bool {
		if mismatches[i].Kind != mismatches[j].Kind {
			return mismatches[i].Kind < mismatches[j].Kind
		}
		if mismatches[i].Domain != mismatches[j].Domain {
			return mismatches[i].Domain < mismatches[j].Domain
		}
		return mismatches[i].Class < mismatches[j].Class
	})

	return mismatches
}

func toSet(in []string) map[string]struct{} {
	out := make(map[string]struct{}, len(in))
	for _, s := range in {
		out[s] = struct{}{}
	}
	return out
}

func containsString(in []string, needle string) bool {
	for _, s := range in {
		if s == needle {
			return true
		}
	}
	return false
}
