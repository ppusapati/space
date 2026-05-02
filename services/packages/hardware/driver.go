// driver.go — HardwareDriver interface.
//
// → REQ-FUNC-GS-HW-001
// → design.md §4.4
//
// A HardwareDriver is the software-facing handle on a single SDR
// device (USRP B210, RTL-SDR, custom). It is a stateful object — a
// driver carries per-channel tuning + gain, an open RX stream, and an
// open TX stream. Callers MUST hold one driver per physical device
// and MUST call Close before discarding it.

package hardware

import (
	"context"
	"errors"
	"time"
)

// Band names the RF band the driver is currently tuned to. The
// canonical values match the Band enum in
// services/packages/proto/space/satellite/v1/profile.proto so the
// scheduling code can map a SpacecraftProfile band directly onto the
// SDR configuration.
type Band string

const (
	// BandUHF covers the amateur + commercial UHF spectrum used by
	// most CubeSats (~400 MHz).
	BandUHF Band = "UHF"
	// BandS covers ~2 GHz S-band downlink (mission/payload TM).
	BandS Band = "S"
	// BandX covers ~8 GHz X-band downlink (high-rate payload).
	BandX Band = "X"
)

// Modulation is the demodulation scheme applied to incoming IQ data.
// Aligns with the Modulation enum in profile.proto.
type Modulation string

// Canonical Modulation values; mirror profile.proto and
// services/packages/profile.Modulation.
const (
	ModBPSK  Modulation = "BPSK"
	ModQPSK  Modulation = "QPSK"
	ModOQPSK Modulation = "OQPSK"
	Mod8PSK  Modulation = "8PSK"
	ModGMSK  Modulation = "GMSK"
)

// TuneRequest fully describes a tuning request. All fields are
// required; the driver returns ErrInvalidConfig when any field is
// outside the device's capabilities (call Capabilities() first to
// learn the device limits).
type TuneRequest struct {
	// CenterHz is the centre frequency in Hertz. Must lie within the
	// device's tunable range.
	CenterHz uint64

	// SampleRateHz is the IQ sample rate in samples-per-second per
	// channel. Must lie within the device's supported set.
	SampleRateHz uint32

	// Band is the band label the request is for. Used by the driver
	// to engage the correct front-end filter / LNA / amplifier path.
	Band Band

	// Modulation is the demodulation scheme the receive pipeline will
	// apply. The driver itself does not demodulate — this hint lets
	// hardware that has front-end demod assist (e.g. some USRPs)
	// pre-configure the path.
	Modulation Modulation
}

// IQSample is a single complex baseband sample. Float32 because that
// is what the open-source SDR ecosystem standardises on (UHD,
// SoapySDR, GNU Radio); converters to int16 / int8 are adapter-side
// concerns.
type IQSample struct {
	I float32
	Q float32
}

// HardwareDriver is the interface every SDR adapter implements. The
// six methods are the minimum set the chetana scheduling +
// TM/TC pipelines require. Vendor-specific knobs are exposed via
// adapter-private configuration loaded at registry initialisation,
// not via this interface.
//
// All methods are safe to call from a single goroutine. The Rx and
// Tx streaming methods install long-lived background work (a goroutine
// pulling from the device into the supplied channel); concurrent
// calls to Tune / SetGain / Close while a stream is active produce
// well-defined errors (ErrBusy) rather than races.
type HardwareDriver interface {
	// Tune programs the device's centre frequency, sample rate, and
	// band-specific front-end. Returns ErrInvalidConfig when the
	// request exceeds the device capabilities and ErrBusy when an
	// active RX/TX stream prevents reconfiguration.
	//
	// Idempotent — calling Tune with the same TuneRequest twice is a
	// no-op.
	Tune(ctx context.Context, req TuneRequest) error

	// SetGain programs the receive gain in dB. The acceptable range
	// is reported by Capabilities; out-of-range values yield
	// ErrInvalidConfig. Side effect: changes the receive sensitivity
	// for any active RX stream within ~10 ms.
	SetGain(ctx context.Context, gainDB float32) error

	// RxIQ opens a continuous RX stream. Samples are delivered to
	// `out` in the order received from the device; backpressure is
	// the consumer's responsibility (an unbuffered channel will drop
	// samples). The driver writes one logical "frame" of `chunk`
	// samples per channel write to keep allocator pressure bounded.
	//
	// Returns when ctx is cancelled. RxIQ MUST be called after Tune.
	// Calling RxIQ while another RX stream is active returns ErrBusy.
	RxIQ(ctx context.Context, chunk int, out chan<- []IQSample) error

	// TxIQ programs the device for a one-shot TX burst. `samples`
	// must be at least `chunk` long; the driver transmits them
	// at-real-time and returns when the device acknowledges the
	// final sample. Long bursts that exceed the device buffer return
	// ErrBufferOverflow; use TxStream for continuous TX.
	//
	// Side effect: switches the front-end into TX mode for the
	// duration of the burst.
	TxIQ(ctx context.Context, samples []IQSample) error

	// TxStream opens a continuous TX stream. Samples written to `in`
	// are emitted at the configured sample rate; the driver returns
	// when ctx is cancelled OR when `in` is closed. Closing `in`
	// before all samples are transmitted returns
	// ErrTransmissionAborted.
	TxStream(ctx context.Context, in <-chan []IQSample) error

	// Close releases the device handle, stops any active streams,
	// and waits for them to drain. Safe to call multiple times; the
	// second and subsequent calls return ErrAlreadyClosed.
	Close() error

	// Capabilities reports the device's supported tuning + gain
	// envelope. The same values are valid for the lifetime of the
	// handle; callers may cache the result.
	Capabilities() Capabilities
}

// Capabilities describes a single SDR device's limits. Returned from
// HardwareDriver.Capabilities; immutable for the lifetime of the
// driver handle.
type Capabilities struct {
	// MinFreqHz / MaxFreqHz bound the tunable range.
	MinFreqHz uint64
	MaxFreqHz uint64

	// SupportedSampleRatesHz lists every sample rate the device can
	// be configured to. Tune() with an off-list rate returns
	// ErrInvalidConfig.
	SupportedSampleRatesHz []uint32

	// MinGainDB / MaxGainDB / GainStepDB describe the receive AGC
	// envelope.
	MinGainDB  float32
	MaxGainDB  float32
	GainStepDB float32

	// SupportedBands lists which bands the device front-end can
	// engage.
	SupportedBands []Band

	// SupportedModulations lists modulations the device's optional
	// front-end demod assist can pre-configure for. Adapters with no
	// front-end demod assist return all values from this package's
	// canonical Modulation list.
	SupportedModulations []Modulation

	// VendorName / ModelName / SerialNumber are diagnostic strings.
	VendorName   string
	ModelName    string
	SerialNumber string
}

// HardwareDriverFactory builds a HardwareDriver from an
// adapter-specific config blob and the surrounding context. The
// registry calls factories at lookup time; one factory per adapter
// name lives in the registry.
type HardwareDriverFactory func(ctx context.Context, config any) (HardwareDriver, error)

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrInvalidConfig is returned when a Tune / SetGain / TxIQ / TxStream
// argument exceeds the device's capabilities or violates an
// invariant (e.g. zero sample rate).
var ErrInvalidConfig = errors.New("hardware: invalid config")

// ErrBusy is returned when an operation cannot proceed because the
// driver is in a state that forbids it (Tune called while RX is
// streaming; second concurrent RxIQ on the same handle).
var ErrBusy = errors.New("hardware: device busy")

// ErrBufferOverflow is returned when a TxIQ request is larger than
// the device buffer can hold; callers should split the burst or use
// TxStream.
var ErrBufferOverflow = errors.New("hardware: buffer overflow")

// ErrTransmissionAborted is returned when TxStream's input channel
// closes before all queued samples have been emitted.
var ErrTransmissionAborted = errors.New("hardware: transmission aborted")

// ErrAlreadyClosed is returned by Close on the second + subsequent
// call against the same handle.
var ErrAlreadyClosed = errors.New("hardware: driver already closed")

// ErrNotTuned is returned when an operation requires a prior Tune
// (RxIQ / TxIQ / TxStream / SetGain on a fresh handle).
var ErrNotTuned = errors.New("hardware: device not tuned")

// rxBackpressureWindow caps the maximum time the driver waits for the
// consumer to read from the RX channel before counting a drop. It is
// exposed in this file (not test-only) because the in-memory fake
// honours the same value, keeping conformance tests stable.
const rxBackpressureWindow = 200 * time.Millisecond
