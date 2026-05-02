// network.go — GroundNetworkProvider interface.
//
// → REQ-FUNC-GS-HW-003
// → design.md §4.4
//
// A GroundNetworkProvider is the abstraction over a *network of
// dishes* — a single Provider may control one antenna (own-dish) or
// hundreds (AWS Ground Station, KSAT). The scheduling pipeline talks
// to providers to *reserve a contact window* with a satellite; once
// reserved, the actual signal-flow is handled by the underlying
// HardwareDriver + AntennaController for own-dish, or by the
// provider's managed dataflow for AWS GS / KSAT.

package hardware

import (
	"context"
	"errors"
	"time"
)

// ContactRequest is a request to reserve a single contact window. The
// scheduler passes the satellite identity + the desired pass envelope
// + minimum acceptable elevation; the provider decides which
// antenna(s) to allocate and replies with a Contact handle.
type ContactRequest struct {
	// SatelliteID is the spacecraft this contact is for. The string
	// format is provider-specific (NORAD ID, internal mission ID,
	// AWS GS satellite ARN). The scheduler maps SpacecraftProfile
	// → provider-specific ID via per-provider config.
	SatelliteID string

	// Window is the requested contact window in wall-clock UTC.
	// Provider returns ErrNoCapacity when no antenna in its network
	// can serve the entire window.
	Window TimeWindow

	// MinElevationDeg is the minimum elevation the antenna must be
	// able to track during the window. Lower values relax the
	// allocation constraint but degrade link budget.
	MinElevationDeg float64

	// FrequenciesHz lists the centre frequencies the contact will
	// use. Required for managed providers (AWS GS allocates
	// hardware frontend per frequency); ignored by own-dish.
	FrequenciesHz []uint64

	// Bands lists the bands the contact will use. Provider returns
	// ErrInvalidConfig when the requested bands cannot be served by
	// any antenna in its network.
	Bands []Band
}

// TimeWindow is a closed-open wall-clock interval [Start, End).
type TimeWindow struct {
	Start time.Time
	End   time.Time
}

// Contact is the handle returned by AllocateContact. The scheduler
// uses ContactID to release the contact later via ReleaseContact and
// to correlate downstream telemetry with the originating reservation.
type Contact struct {
	// ContactID uniquely identifies the contact within the
	// provider's namespace. Stable for the lifetime of the
	// reservation.
	ContactID string

	// AllocatedAntennaID names the antenna serving this contact
	// inside the provider. Diagnostic only — the scheduler does not
	// route through this field.
	AllocatedAntennaID string

	// Window is the actual reserved window (may be a subset of the
	// request when the provider truncates to honour adjacency rules).
	Window TimeWindow

	// State is the current lifecycle state of the contact.
	State ContactState

	// Cost is an opaque provider-specific cost tag (USD cents,
	// internal credits). Empty when the provider is unmetered
	// (own-dish).
	Cost string
}

// ContactState is the lifecycle of a Contact reservation.
type ContactState string

const (
	// ContactReserved is the post-AllocateContact state. The
	// provider has committed to the window but has not yet started.
	ContactReserved ContactState = "reserved"

	// ContactScheduled means the provider has all preflight checks
	// done and will execute at Window.Start.
	ContactScheduled ContactState = "scheduled"

	// ContactExecuting means Window.Start has passed and the
	// antenna is now tracking the satellite.
	ContactExecuting ContactState = "executing"

	// ContactCompleted means Window.End has passed and the
	// antenna has been released.
	ContactCompleted ContactState = "completed"

	// ContactCancelled is the post-ReleaseContact state when the
	// caller released the contact before Window.Start.
	ContactCancelled ContactState = "cancelled"

	// ContactFailed is set when the provider could not execute the
	// contact (hardware fault, weather abort).
	ContactFailed ContactState = "failed"
)

// GroundNetworkProvider is the interface every ground-network
// adapter implements. Three operations cover the lifecycle:
// reserve, release, list. A fourth (Capabilities) lets the
// scheduling pipeline understand what the provider can serve before
// it tries to allocate.
//
// All methods are safe for concurrent calls — the provider
// internally serialises its access to the underlying provider API.
type GroundNetworkProvider interface {
	// AllocateContact reserves a contact window. Returns
	// ErrNoCapacity when no antenna can serve the request and
	// ErrInvalidConfig when the request is malformed (window in the
	// past, end before start, no bands listed).
	//
	// On success the returned Contact is in ContactReserved state.
	AllocateContact(ctx context.Context, req ContactRequest) (Contact, error)

	// ReleaseContact frees a previously-reserved contact. Safe to
	// call on contacts in any state; no-op for contacts already in
	// ContactCompleted or ContactCancelled. Returns
	// ErrUnknownContact when the contactID is not known to this
	// provider.
	ReleaseContact(ctx context.Context, contactID string) error

	// ListContacts returns every contact the provider currently
	// tracks for this account, optionally filtered by state. Pass
	// states=nil to receive every contact. Sorted by Window.Start
	// ascending.
	ListContacts(ctx context.Context, states []ContactState) ([]Contact, error)

	// NetworkCapabilities reports the provider's static envelope
	// (supported bands, frequencies, antenna count). Intended for
	// scheduler pre-flight checks — the scheduler can avoid issuing
	// AllocateContact requests doomed to ErrInvalidConfig.
	NetworkCapabilities() NetworkCapabilities

	// Close releases any open connections / sessions held against
	// the provider's API. Safe to call multiple times.
	Close() error
}

// NetworkCapabilities describes a single provider's static envelope.
type NetworkCapabilities struct {
	// ProviderName identifies the adapter (e.g. "own-dish",
	// "aws-gs", "ksat", "ssc"). Diagnostic only.
	ProviderName string

	// AntennaCount is the number of physical dishes the provider
	// can serve from. own-dish typically reports 1; AWS GS reports
	// the count of dishes in the configured AWS region.
	AntennaCount int

	// SupportedBands lists every band any antenna in the provider's
	// network can serve. Per-antenna details may be more
	// restrictive — the provider takes care of routing to a
	// suitable antenna at AllocateContact time.
	SupportedBands []Band

	// MinFreqHz / MaxFreqHz bound the union of the network's
	// tunable range.
	MinFreqHz uint64
	MaxFreqHz uint64

	// Metered reports whether the provider charges per contact.
	// Used by the scheduler to log cost projections.
	Metered bool

	// MinAdvanceNoticeBeforeStart is the minimum lead time between
	// AllocateContact and Window.Start the provider requires. AWS
	// GS requires ~15 minutes; own-dish accepts 0.
	MinAdvanceNoticeBeforeStart time.Duration
}

// GroundNetworkProviderFactory builds a GroundNetworkProvider from
// adapter-specific config. One factory per provider name lives in
// the registry.
type GroundNetworkProviderFactory func(ctx context.Context, config any) (GroundNetworkProvider, error)

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrNoCapacity is returned when no antenna in the provider's
// network can serve the requested contact window.
var ErrNoCapacity = errors.New("hardware: no capacity for contact")

// ErrUnknownContact is returned by ReleaseContact when the supplied
// contactID is not known to the provider.
var ErrUnknownContact = errors.New("hardware: unknown contact")
