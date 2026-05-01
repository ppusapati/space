package calchelpers

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"p9e.in/samavaya/packages/errors"
)

// ClassDispatcher is the class-aware branching primitive for Layer 3
// calculation services. A calculation registers a handler function
// per class it supports; at runtime the service calls Dispatch with
// the entity's class string and the dispatcher picks the matching
// handler.
//
// The pattern works for any calculation shape — the generic type
// parameters Req and Resp let each service keep typed contracts.
// There is no string-key result map; every handler has the same
// typed signature and the dispatcher never handles types at runtime.
//
// Typical usage (sketch — actual calculation lives in its own
// package):
//
//	type OEEInput struct { WorkCenterID string; Start, End time.Time }
//	type OEEResult struct { Availability, Performance, Quality float64 }
//
//	type oeeService struct {
//	    disp *calchelpers.ClassDispatcher[OEEInput, OEEResult]
//	    wc   workcenterclient.Client
//	    reg  classregistry.Registry
//	}
//
//	func NewOEEService(wc workcenterclient.Client, reg classregistry.Registry) *oeeService {
//	    s := &oeeService{wc: wc, reg: reg}
//	    s.disp = calchelpers.NewClassDispatcher[OEEInput, OEEResult]("oee_calculation")
//	    s.disp.Register("mfg_discrete", s.calcDiscrete)
//	    s.disp.Register("solar_mfg_line", s.calcDiscrete)
//	    s.disp.Register("pharma_mfg_line", s.calcBatch)
//	    s.disp.Register("mfg_batch",       s.calcBatch)
//	    return s
//	}
//
//	func (s *oeeService) Calculate(ctx context.Context, in OEEInput) (*OEEResult, error) {
//	    wc, err := s.wc.Get(ctx, in.WorkCenterID)
//	    if err != nil {
//	        return nil, err
//	    }
//	    return s.disp.Dispatch(ctx, wc.Class, in)
//	}
//
// # Error codes
//
// Two typed errors, both errors.BadRequest (the caller is asking for
// something invalid, not an internal problem):
//
//   - CLASS_UNSUPPORTED — the entity's class doesn't appear in the
//     dispatcher's registry. The message names the calculation and
//     the class so authors know exactly what to add.
//   - CLASS_REGISTERED_BUT_NIL — a handler was registered for the
//     class but the function reference is nil. This would be a
//     programming mistake in the calculation service's init path;
//     the dispatcher refuses to invoke nil.
//
// # Thread safety
//
// Registration happens at service construction time; Dispatch is
// called from request handlers. The mutex covers both so late-
// registration patterns (e.g. class-registry hot reload in a future
// F.x task) stay safe even if rare.
type ClassDispatcher[Req any, Resp any] struct {
	// CalculationName is the user-facing name of the calculation
	// (e.g. "oee_calculation"). Appears in error messages so the
	// author knows which service complained.
	calculationName string

	mu       sync.RWMutex
	handlers map[string]Handler[Req, Resp]
}

// Handler is the function shape every class variant implements.
// ctx carries the request deadline and RLS scope; in is the typed
// calculation input; the return pair is the typed result plus any
// error. Handlers should return errors.* typed errors so the
// composition root can translate to wire-appropriate codes.
type Handler[Req any, Resp any] func(ctx context.Context, in Req) (*Resp, error)

// NewClassDispatcher constructs a dispatcher for the named
// calculation. The name is used only for diagnostic messages; the
// dispatcher doesn't cross-check against the class registry itself
// (that's the Conformance helper's job).
func NewClassDispatcher[Req any, Resp any](calculationName string) *ClassDispatcher[Req, Resp] {
	return &ClassDispatcher[Req, Resp]{
		calculationName: calculationName,
		handlers:        make(map[string]Handler[Req, Resp]),
	}
}

// Register binds a handler to a class name. Calling Register twice
// for the same class overwrites the previous binding — services may
// do this intentionally to replace a variant at test time.
func (d *ClassDispatcher[Req, Resp]) Register(class string, h Handler[Req, Resp]) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[class] = h
}

// Dispatch looks up the handler for the given class and invokes it.
// Returns a typed errors.BadRequest when the class has no handler
// or when the registered handler is nil.
func (d *ClassDispatcher[Req, Resp]) Dispatch(ctx context.Context, class string, in Req) (*Resp, error) {
	d.mu.RLock()
	h, ok := d.handlers[class]
	d.mu.RUnlock()
	if !ok {
		return nil, errors.BadRequest(
			"CLASS_UNSUPPORTED",
			fmt.Sprintf("calculation %q has no handler for class %q (supported: %v)",
				d.calculationName, class, d.SupportedClasses()),
		)
	}
	if h == nil {
		return nil, errors.InternalServer(
			"CLASS_REGISTERED_BUT_NIL",
			fmt.Sprintf("calculation %q registered a nil handler for class %q — programming error in service init",
				d.calculationName, class),
		)
	}
	return h(ctx, in)
}

// SupportedClasses returns the sorted list of classes the
// dispatcher has handlers for. Used in error messages and by the
// Conformance helper.
func (d *ClassDispatcher[Req, Resp]) SupportedClasses() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]string, 0, len(d.handlers))
	for class := range d.handlers {
		out = append(out, class)
	}
	sort.Strings(out)
	return out
}

// CalculationName returns the dispatcher's calculation name —
// convenient for services that want to include it in their own
// diagnostic output.
func (d *ClassDispatcher[Req, Resp]) CalculationName() string {
	return d.calculationName
}
