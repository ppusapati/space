// antenna.go — AntennaController interface.
//
// → REQ-FUNC-GS-HW-002
// → design.md §4.4
//
// An AntennaController commands a single rotator (Yaesu G-5500 via
// Hamlib rotctld, GS-232 protocol over RS-232/TCP, or a custom
// adapter). Coordinates are azimuth + elevation in degrees per the
// IEEE convention (azimuth 0=North, 90=East; elevation 0=horizon,
// 90=zenith).

package hardware

import (
	"context"
	"errors"
	"time"
)

// AzEl is an absolute pointing target in degrees.
type AzEl struct {
	AzimuthDeg   float64 // 0..360, 0=North, increasing clockwise
	ElevationDeg float64 // -90..+90, 0=horizon, 90=zenith
}

// TrackPoint is a single point in a tracking trajectory. The
// SetTrack method consumes a sequence of these to chase a moving
// target (a satellite during a pass).
type TrackPoint struct {
	When time.Time // wall-clock time at which the rotator should be at AzEl
	AzEl AzEl
}

// AntennaController is the interface every rotator adapter
// implements. Six methods cover the operational lifecycle: tune,
// query position, track a trajectory, park, stow, close.
//
// All methods are safe to call from a single goroutine. SetTrack
// installs a background loop on the controller; concurrent calls to
// SetAzEl while a track is active produce ErrBusy. SetTrack can be
// cancelled by cancelling its context or by calling Park / Stow,
// which preempt any active track.
type AntennaController interface {
	// SetAzEl commands an absolute pointing. Returns when the
	// rotator reports it has reached the target (within the
	// adapter-defined slew tolerance) or when ctx is cancelled.
	// Returns ErrInvalidPointing when the target is outside the
	// rotator's mechanical envelope (e.g. negative elevation on a
	// rotator that has no below-horizon range).
	SetAzEl(ctx context.Context, target AzEl) error

	// GetAzEl reports the rotator's current pointing. Position is
	// read fresh from the device each call; cached values are not
	// returned. Side effects: none.
	GetAzEl(ctx context.Context) (AzEl, error)

	// SetTrack hands the controller a pre-computed trajectory and
	// returns when the trajectory completes (the last point's When
	// time has passed) or ctx is cancelled. Each TrackPoint's When
	// time MUST be strictly increasing; ErrInvalidTrack is returned
	// otherwise.
	//
	// The controller adjusts pointing continuously between points;
	// the rate is adapter-dependent (typically 1 Hz update for
	// Hamlib, faster for direct serial). Calling SetAzEl while a
	// track is active returns ErrBusy.
	SetTrack(ctx context.Context, trajectory []TrackPoint) error

	// Park drives the rotator to the adapter's configured park
	// position (usually az=0, el=90 — pointing straight up so wind
	// and rain shed off the dish). Preempts any active track.
	// Returns when the park position is reached.
	Park(ctx context.Context) error

	// Stow drives the rotator to the adapter's configured stow
	// position (usually az=0, el=0 — horizontal, locked) for
	// maintenance / high-wind events. Preempts any active track.
	// Returns when the stow position is reached.
	Stow(ctx context.Context) error

	// Close releases the controller handle. After Close the
	// controller MUST NOT issue any further movement commands; a
	// subsequent call to any method returns ErrAlreadyClosed.
	Close() error

	// AntennaCapabilities reports the rotator envelope + speed.
	AntennaCapabilities() AntennaCapabilities
}

// AntennaCapabilities describes one rotator's mechanical envelope.
type AntennaCapabilities struct {
	// MinAzDeg / MaxAzDeg bound the azimuth range. Most rotators
	// allow 0..450 to permit overshoot through North; specifying
	// values outside the documented range yields ErrInvalidPointing.
	MinAzDeg float64
	MaxAzDeg float64

	// MinElDeg / MaxElDeg bound the elevation range. Below-horizon
	// rotators have MinElDeg < 0.
	MinElDeg float64
	MaxElDeg float64

	// MaxSlewDegSec is the rotator's maximum slew rate. SetTrack
	// trajectories that demand faster motion are tracked at the max
	// rate (with a corresponding pointing error reported via
	// telemetry).
	MaxSlewDegSec float64

	// PointingAccuracyDeg is the rotator's claimed pointing
	// accuracy. SetAzEl returns when the rotator is within this
	// many degrees of the target.
	PointingAccuracyDeg float64

	// VendorName / ModelName / SerialNumber are diagnostic strings.
	VendorName   string
	ModelName    string
	SerialNumber string
}

// AntennaControllerFactory builds an AntennaController from
// adapter-specific config. One factory per adapter name lives in
// the registry.
type AntennaControllerFactory func(ctx context.Context, config any) (AntennaController, error)

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrInvalidPointing is returned when a target azimuth/elevation lies
// outside the rotator's mechanical envelope.
var ErrInvalidPointing = errors.New("hardware: invalid pointing")

// ErrInvalidTrack is returned when SetTrack is called with a
// malformed trajectory (empty, non-monotonic, or containing an
// invalid pointing).
var ErrInvalidTrack = errors.New("hardware: invalid track")
