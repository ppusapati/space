// registry.go — adapter registry.
//
// → REQ-FUNC-GS-HW-001 / -002 / -003
// → design.md §4.4
//
// Service entrypoints register every adapter they want available at
// boot, then the scheduler / RF service looks adapters up by name
// from configuration. The registry is the single coupling point
// between the abstract interfaces in this package and the concrete
// adapters in chetana-defense (UHD, RTL, Hamlib, GS-232, AWS GS) +
// the in-memory fake in package hardware/fake (tests).

package hardware

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Registry holds the per-process map of adapter-name → factory for
// each of the three interfaces. Service entrypoints construct one
// Registry, register every adapter they ship, then pass the registry
// down to the gs-rf / gs-scheduler services.
//
// Concurrency: every method is goroutine-safe.
type Registry struct {
	mu sync.RWMutex

	drivers   map[string]HardwareDriverFactory
	antennas  map[string]AntennaControllerFactory
	providers map[string]GroundNetworkProviderFactory
}

// NewRegistry returns a fresh empty Registry. Service entrypoints
// typically wire it up like:
//
//	reg := hardware.NewRegistry()
//	uhd.Register(reg)        // chetana-defense package
//	hamlib.Register(reg)     // chetana-defense package
//	awsgs.Register(reg)      // chetana-defense package
//	fake.Register(reg)       // services/packages/hardware/fake (tests)
func NewRegistry() *Registry {
	return &Registry{
		drivers:   make(map[string]HardwareDriverFactory),
		antennas:  make(map[string]AntennaControllerFactory),
		providers: make(map[string]GroundNetworkProviderFactory),
	}
}

// ----------------------------------------------------------------------
// Registration
// ----------------------------------------------------------------------

// RegisterHardwareDriver attaches a HardwareDriver factory under the
// supplied name. Returns ErrDuplicateAdapter when a factory with
// the same name was already registered (the same name across the
// three interface types is allowed; only same-interface duplicates
// collide).
func (r *Registry) RegisterHardwareDriver(name string, factory HardwareDriverFactory) error {
	if name == "" {
		return ErrInvalidAdapterName
	}
	if factory == nil {
		return fmt.Errorf("hardware: nil HardwareDriverFactory for %q", name)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.drivers[name]; ok {
		return fmt.Errorf("hardware: %w: HardwareDriver %q", ErrDuplicateAdapter, name)
	}
	r.drivers[name] = factory
	return nil
}

// RegisterAntennaController attaches an AntennaController factory
// under the supplied name. Same duplicate semantics as
// RegisterHardwareDriver.
func (r *Registry) RegisterAntennaController(name string, factory AntennaControllerFactory) error {
	if name == "" {
		return ErrInvalidAdapterName
	}
	if factory == nil {
		return fmt.Errorf("hardware: nil AntennaControllerFactory for %q", name)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.antennas[name]; ok {
		return fmt.Errorf("hardware: %w: AntennaController %q", ErrDuplicateAdapter, name)
	}
	r.antennas[name] = factory
	return nil
}

// RegisterGroundNetworkProvider attaches a GroundNetworkProvider
// factory under the supplied name.
func (r *Registry) RegisterGroundNetworkProvider(name string, factory GroundNetworkProviderFactory) error {
	if name == "" {
		return ErrInvalidAdapterName
	}
	if factory == nil {
		return fmt.Errorf("hardware: nil GroundNetworkProviderFactory for %q", name)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.providers[name]; ok {
		return fmt.Errorf("hardware: %w: GroundNetworkProvider %q", ErrDuplicateAdapter, name)
	}
	r.providers[name] = factory
	return nil
}

// ----------------------------------------------------------------------
// Lookup
// ----------------------------------------------------------------------

// NewHardwareDriver looks up the named driver factory and invokes it
// with the supplied config. Returns ErrUnknownAdapter when the name
// is not registered.
func (r *Registry) NewHardwareDriver(ctx context.Context, name string, config any) (HardwareDriver, error) {
	r.mu.RLock()
	f, ok := r.drivers[name]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("hardware: %w: HardwareDriver %q", ErrUnknownAdapter, name)
	}
	return f(ctx, config)
}

// NewAntennaController looks up the named antenna factory and
// invokes it with the supplied config. Returns ErrUnknownAdapter
// when the name is not registered.
func (r *Registry) NewAntennaController(ctx context.Context, name string, config any) (AntennaController, error) {
	r.mu.RLock()
	f, ok := r.antennas[name]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("hardware: %w: AntennaController %q", ErrUnknownAdapter, name)
	}
	return f(ctx, config)
}

// NewGroundNetworkProvider looks up the named provider factory and
// invokes it with the supplied config. Returns ErrUnknownAdapter
// when the name is not registered.
func (r *Registry) NewGroundNetworkProvider(ctx context.Context, name string, config any) (GroundNetworkProvider, error) {
	r.mu.RLock()
	f, ok := r.providers[name]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("hardware: %w: GroundNetworkProvider %q", ErrUnknownAdapter, name)
	}
	return f(ctx, config)
}

// ----------------------------------------------------------------------
// Introspection
// ----------------------------------------------------------------------

// HardwareDriverNames returns the sorted list of registered driver
// adapter names. Used by configuration validators to surface a
// helpful error when a service config references an unknown adapter.
func (r *Registry) HardwareDriverNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.drivers))
	for k := range r.drivers {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// AntennaControllerNames returns the sorted list of registered
// antenna adapter names.
func (r *Registry) AntennaControllerNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.antennas))
	for k := range r.antennas {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// GroundNetworkProviderNames returns the sorted list of registered
// provider adapter names.
func (r *Registry) GroundNetworkProviderNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.providers))
	for k := range r.providers {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrInvalidAdapterName is returned by Register* when called with an
// empty name.
var ErrInvalidAdapterName = errors.New("hardware: empty adapter name")

// ErrDuplicateAdapter is returned by Register* when a second factory
// is registered against the same name within the same interface type.
var ErrDuplicateAdapter = errors.New("hardware: duplicate adapter")

// ErrUnknownAdapter is returned by New* when the requested adapter
// name has not been registered.
var ErrUnknownAdapter = errors.New("hardware: unknown adapter")
