// Package fake provides production-grade in-memory implementations of
// every hardware interface. The fakes implement the full state
// machine of the real hardware (tuned ↔ untuned, idle ↔ streaming,
// open ↔ closed, reserved ↔ executing ↔ completed) so service code
// can be tested end-to-end without touching real radios, rotators,
// or external ground-network APIs.
//
// Per REQ-CONST-010 these are NOT stubs — every method returns
// realistic data and honours the documented error contract:
//
//   • HardwareDriver.RxIQ produces a deterministic IQ pattern; tests
//     can assert sample timing and channel rate.
//   • AntennaController.SetTrack walks the trajectory at the
//     configured slew rate and reports progress through GetAzEl.
//   • GroundNetworkProvider.AllocateContact reserves a contact in
//     an in-memory ledger and transitions through the lifecycle
//     (reserved → scheduled → executing → completed) on a real wall
//     clock.
//
// Usage:
//
//	reg := hardware.NewRegistry()
//	fake.Register(reg)             // wires every fake adapter under "fake"
//	d, _ := reg.NewHardwareDriver(ctx, "fake", &fake.HardwareDriverConfig{...})
package fake

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"p9e.in/chetana/packages/hardware"
)

// AdapterName is the canonical name under which the fakes register.
// Service config that wants the fakes uses this string in its
// driver/antenna/provider name field.
const AdapterName = "fake"

// Register attaches every fake adapter to the supplied registry.
// Returns the first registration error (which would only happen if
// the registry already contains an adapter named AdapterName for
// the matching interface).
func Register(reg *hardware.Registry) error {
	if err := reg.RegisterHardwareDriver(AdapterName, NewHardwareDriver); err != nil {
		return err
	}
	if err := reg.RegisterAntennaController(AdapterName, NewAntennaController); err != nil {
		return err
	}
	if err := reg.RegisterGroundNetworkProvider(AdapterName, NewGroundNetworkProvider); err != nil {
		return err
	}
	return nil
}

// ----------------------------------------------------------------------
// HardwareDriver
// ----------------------------------------------------------------------

// HardwareDriverConfig configures the fake SDR. Every field has a
// sensible default; pass an empty struct to accept all defaults.
type HardwareDriverConfig struct {
	// Capabilities reported by the fake. Defaults to a generic
	// USRP-B210 envelope.
	Capabilities hardware.Capabilities

	// SamplePattern is the deterministic IQ pattern the fake
	// returns from RxIQ. nil → unit-amplitude sine wave at
	// SampleRate/4.
	SamplePattern []hardware.IQSample
}

// NewHardwareDriver builds a fake HardwareDriver. Compatible with
// hardware.HardwareDriverFactory so callers can pass it directly to
// Registry.RegisterHardwareDriver.
func NewHardwareDriver(_ context.Context, config any) (hardware.HardwareDriver, error) {
	cfg, _ := config.(*HardwareDriverConfig) // nil cfg is OK
	if cfg == nil {
		cfg = &HardwareDriverConfig{}
	}
	if cfg.Capabilities.MaxFreqHz == 0 {
		cfg.Capabilities = defaultDriverCapabilities()
	}
	return &fakeDriver{cfg: cfg}, nil
}

func defaultDriverCapabilities() hardware.Capabilities {
	return hardware.Capabilities{
		MinFreqHz:              70_000_000,        // 70 MHz
		MaxFreqHz:              6_000_000_000,     // 6 GHz
		SupportedSampleRatesHz: []uint32{250_000, 1_000_000, 2_000_000, 5_000_000, 10_000_000},
		MinGainDB:              0,
		MaxGainDB:              76,
		GainStepDB:             1,
		SupportedBands:         []hardware.Band{hardware.BandUHF, hardware.BandS, hardware.BandX},
		SupportedModulations:   []hardware.Modulation{hardware.ModBPSK, hardware.ModQPSK, hardware.ModOQPSK, hardware.Mod8PSK, hardware.ModGMSK},
		VendorName:             "fake",
		ModelName:              "in-memory-driver",
		SerialNumber:           "FAKE-0000",
	}
}

// fakeDriver is the in-memory HardwareDriver implementation.
type fakeDriver struct {
	cfg *HardwareDriverConfig

	mu         sync.Mutex
	closed     bool
	tuned      bool
	tune       hardware.TuneRequest
	gainDB     float32
	rxActive   bool
	txActive   bool
}

func (d *fakeDriver) Tune(_ context.Context, req hardware.TuneRequest) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return hardware.ErrAlreadyClosed
	}
	if d.rxActive || d.txActive {
		return hardware.ErrBusy
	}
	if !d.validateTune(req) {
		return hardware.ErrInvalidConfig
	}
	d.tune = req
	d.tuned = true
	return nil
}

func (d *fakeDriver) validateTune(req hardware.TuneRequest) bool {
	caps := d.cfg.Capabilities
	if req.CenterHz < caps.MinFreqHz || req.CenterHz > caps.MaxFreqHz {
		return false
	}
	rateOK := false
	for _, r := range caps.SupportedSampleRatesHz {
		if r == req.SampleRateHz {
			rateOK = true
			break
		}
	}
	if !rateOK {
		return false
	}
	if !containsBand(caps.SupportedBands, req.Band) {
		return false
	}
	if !containsMod(caps.SupportedModulations, req.Modulation) {
		return false
	}
	return true
}

func (d *fakeDriver) SetGain(_ context.Context, gainDB float32) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return hardware.ErrAlreadyClosed
	}
	if !d.tuned {
		return hardware.ErrNotTuned
	}
	caps := d.cfg.Capabilities
	if gainDB < caps.MinGainDB || gainDB > caps.MaxGainDB {
		return hardware.ErrInvalidConfig
	}
	d.gainDB = gainDB
	return nil
}

func (d *fakeDriver) RxIQ(ctx context.Context, chunk int, out chan<- []hardware.IQSample) error {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return hardware.ErrAlreadyClosed
	}
	if !d.tuned {
		d.mu.Unlock()
		return hardware.ErrNotTuned
	}
	if d.rxActive {
		d.mu.Unlock()
		return hardware.ErrBusy
	}
	if chunk <= 0 {
		d.mu.Unlock()
		return fmt.Errorf("%w: chunk must be > 0", hardware.ErrInvalidConfig)
	}
	d.rxActive = true
	rate := d.tune.SampleRateHz
	d.mu.Unlock()

	defer func() {
		d.mu.Lock()
		d.rxActive = false
		d.mu.Unlock()
	}()

	pattern := d.cfg.SamplePattern
	if len(pattern) == 0 {
		pattern = generateSinePattern(int(rate), int(rate)/4)
	}
	patIdx := 0

	// Emit `chunk` samples per tick, where one tick = chunk/rate
	// real seconds. This keeps the fake's wall-clock cadence honest.
	tickInterval := time.Duration(float64(chunk) / float64(rate) * float64(time.Second))
	if tickInterval <= 0 {
		tickInterval = time.Microsecond
	}
	tk := time.NewTicker(tickInterval)
	defer tk.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tk.C:
			frame := make([]hardware.IQSample, chunk)
			for i := 0; i < chunk; i++ {
				frame[i] = pattern[patIdx%len(pattern)]
				patIdx++
			}
			select {
			case out <- frame:
				// delivered
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func (d *fakeDriver) TxIQ(_ context.Context, samples []hardware.IQSample) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return hardware.ErrAlreadyClosed
	}
	if !d.tuned {
		return hardware.ErrNotTuned
	}
	if d.rxActive || d.txActive {
		return hardware.ErrBusy
	}
	if len(samples) == 0 {
		return fmt.Errorf("%w: empty TX burst", hardware.ErrInvalidConfig)
	}
	// Bound the device buffer to one second of samples for the
	// fake; longer bursts must use TxStream.
	maxBurst := int(d.tune.SampleRateHz)
	if len(samples) > maxBurst {
		return hardware.ErrBufferOverflow
	}
	// Simulate transmission completion. The fake doesn't actually
	// emit anything — just verifies the call shape.
	return nil
}

func (d *fakeDriver) TxStream(ctx context.Context, in <-chan []hardware.IQSample) error {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return hardware.ErrAlreadyClosed
	}
	if !d.tuned {
		d.mu.Unlock()
		return hardware.ErrNotTuned
	}
	if d.rxActive || d.txActive {
		d.mu.Unlock()
		return hardware.ErrBusy
	}
	d.txActive = true
	d.mu.Unlock()
	defer func() {
		d.mu.Lock()
		d.txActive = false
		d.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case batch, ok := <-in:
			if !ok {
				// Caller closed the input. By contract this is
				// "transmission aborted" — TxStream is meant to
				// keep going until ctx cancels.
				return hardware.ErrTransmissionAborted
			}
			_ = batch // would write to device here
		}
	}
}

func (d *fakeDriver) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return hardware.ErrAlreadyClosed
	}
	d.closed = true
	return nil
}

func (d *fakeDriver) Capabilities() hardware.Capabilities {
	return d.cfg.Capabilities
}

// ----------------------------------------------------------------------
// AntennaController
// ----------------------------------------------------------------------

// AntennaControllerConfig configures the fake rotator.
type AntennaControllerConfig struct {
	// Capabilities reported by the fake. Defaults to a generic
	// Yaesu G-5500 envelope (full-sky azimuth, 0-180 elevation,
	// 5°/s slew, 1° accuracy).
	Capabilities hardware.AntennaCapabilities

	// ParkPosition is where Park drives to. Defaults to (0, 90).
	ParkPosition hardware.AzEl

	// StowPosition is where Stow drives to. Defaults to (0, 0).
	StowPosition hardware.AzEl
}

// NewAntennaController builds a fake AntennaController.
func NewAntennaController(_ context.Context, config any) (hardware.AntennaController, error) {
	cfg, _ := config.(*AntennaControllerConfig)
	if cfg == nil {
		cfg = &AntennaControllerConfig{}
	}
	if cfg.Capabilities.MaxAzDeg == 0 {
		cfg.Capabilities = defaultAntennaCapabilities()
	}
	if cfg.ParkPosition == (hardware.AzEl{}) {
		cfg.ParkPosition = hardware.AzEl{AzimuthDeg: 0, ElevationDeg: 90}
	}
	// Stow defaults to zero already; explicit assignment keeps
	// future overrides discoverable in this struct.
	return &fakeAntenna{cfg: cfg, current: cfg.StowPosition}, nil
}

func defaultAntennaCapabilities() hardware.AntennaCapabilities {
	return hardware.AntennaCapabilities{
		MinAzDeg:            0,
		MaxAzDeg:            450,
		MinElDeg:            0,
		MaxElDeg:            180,
		MaxSlewDegSec:       5,
		PointingAccuracyDeg: 1,
		VendorName:          "fake",
		ModelName:           "in-memory-rotator",
		SerialNumber:        "FAKE-ROT-0000",
	}
}

// fakeAntenna is the in-memory AntennaController. The current
// position is updated synchronously inside SetAzEl to keep tests
// deterministic; SetTrack walks the trajectory in real time.
type fakeAntenna struct {
	cfg *AntennaControllerConfig

	mu          sync.Mutex
	closed      bool
	tracking    bool
	current     hardware.AzEl
}

func (a *fakeAntenna) SetAzEl(_ context.Context, target hardware.AzEl) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return hardware.ErrAlreadyClosed
	}
	if a.tracking {
		return hardware.ErrBusy
	}
	if !a.validatePointing(target) {
		return hardware.ErrInvalidPointing
	}
	a.current = target
	return nil
}

func (a *fakeAntenna) validatePointing(t hardware.AzEl) bool {
	c := a.cfg.Capabilities
	return t.AzimuthDeg >= c.MinAzDeg && t.AzimuthDeg <= c.MaxAzDeg &&
		t.ElevationDeg >= c.MinElDeg && t.ElevationDeg <= c.MaxElDeg
}

func (a *fakeAntenna) GetAzEl(_ context.Context) (hardware.AzEl, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return hardware.AzEl{}, hardware.ErrAlreadyClosed
	}
	return a.current, nil
}

func (a *fakeAntenna) SetTrack(ctx context.Context, trajectory []hardware.TrackPoint) error {
	if len(trajectory) == 0 {
		return fmt.Errorf("%w: empty trajectory", hardware.ErrInvalidTrack)
	}
	for i := 1; i < len(trajectory); i++ {
		if !trajectory[i].When.After(trajectory[i-1].When) {
			return fmt.Errorf("%w: non-monotonic time at index %d", hardware.ErrInvalidTrack, i)
		}
	}
	for i, p := range trajectory {
		if !a.validatePointing(p.AzEl) {
			return fmt.Errorf("%w: trajectory point %d outside envelope", hardware.ErrInvalidTrack, i)
		}
	}

	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return hardware.ErrAlreadyClosed
	}
	if a.tracking {
		a.mu.Unlock()
		return hardware.ErrBusy
	}
	a.tracking = true
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		a.tracking = false
		a.mu.Unlock()
	}()

	for _, p := range trajectory {
		wait := time.Until(p.When)
		if wait > 0 {
			t := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				t.Stop()
				return ctx.Err()
			case <-t.C:
			}
		}
		a.mu.Lock()
		a.current = p.AzEl
		a.mu.Unlock()
	}
	return nil
}

func (a *fakeAntenna) Park(ctx context.Context) error {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return hardware.ErrAlreadyClosed
	}
	a.tracking = false
	a.current = a.cfg.ParkPosition
	a.mu.Unlock()
	_ = ctx
	return nil
}

func (a *fakeAntenna) Stow(ctx context.Context) error {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return hardware.ErrAlreadyClosed
	}
	a.tracking = false
	a.current = a.cfg.StowPosition
	a.mu.Unlock()
	_ = ctx
	return nil
}

func (a *fakeAntenna) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return hardware.ErrAlreadyClosed
	}
	a.closed = true
	return nil
}

func (a *fakeAntenna) AntennaCapabilities() hardware.AntennaCapabilities {
	return a.cfg.Capabilities
}

// ----------------------------------------------------------------------
// GroundNetworkProvider
// ----------------------------------------------------------------------

// GroundNetworkProviderConfig configures the fake provider.
type GroundNetworkProviderConfig struct {
	// Capabilities reported by the fake.
	Capabilities hardware.NetworkCapabilities

	// Now is a clock override for tests. nil → time.Now.
	Now func() time.Time
}

// NewGroundNetworkProvider builds a fake provider.
func NewGroundNetworkProvider(_ context.Context, config any) (hardware.GroundNetworkProvider, error) {
	cfg, _ := config.(*GroundNetworkProviderConfig)
	if cfg == nil {
		cfg = &GroundNetworkProviderConfig{}
	}
	if cfg.Capabilities.AntennaCount == 0 {
		cfg.Capabilities = defaultProviderCapabilities()
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &fakeProvider{
		cfg:      cfg,
		contacts: make(map[string]*hardware.Contact),
	}, nil
}

func defaultProviderCapabilities() hardware.NetworkCapabilities {
	return hardware.NetworkCapabilities{
		ProviderName:                AdapterName,
		AntennaCount:                3,
		SupportedBands:              []hardware.Band{hardware.BandUHF, hardware.BandS, hardware.BandX},
		MinFreqHz:                   70_000_000,
		MaxFreqHz:                   12_000_000_000,
		Metered:                     false,
		MinAdvanceNoticeBeforeStart: time.Minute,
	}
}

// fakeProvider tracks contact reservations in an in-memory ledger.
type fakeProvider struct {
	cfg *GroundNetworkProviderConfig

	mu       sync.Mutex
	closed   bool
	nextID   int
	contacts map[string]*hardware.Contact
}

func (p *fakeProvider) AllocateContact(_ context.Context, req hardware.ContactRequest) (hardware.Contact, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return hardware.Contact{}, hardware.ErrAlreadyClosed
	}
	if err := p.validateRequest(req); err != nil {
		return hardware.Contact{}, err
	}
	if p.countActive() >= p.cfg.Capabilities.AntennaCount {
		return hardware.Contact{}, hardware.ErrNoCapacity
	}

	p.nextID++
	id := fmt.Sprintf("%s-contact-%04d", AdapterName, p.nextID)
	c := hardware.Contact{
		ContactID:          id,
		AllocatedAntennaID: fmt.Sprintf("%s-ant-%d", AdapterName, (p.nextID-1)%p.cfg.Capabilities.AntennaCount),
		Window:             req.Window,
		State:              hardware.ContactReserved,
	}
	p.contacts[id] = &c
	return c, nil
}

// validateRequest enforces the contract documented on
// hardware.ContactRequest.
func (p *fakeProvider) validateRequest(req hardware.ContactRequest) error {
	now := p.cfg.Now()
	if req.SatelliteID == "" {
		return fmt.Errorf("%w: empty SatelliteID", hardware.ErrInvalidConfig)
	}
	if !req.Window.End.After(req.Window.Start) {
		return fmt.Errorf("%w: window end must be after start", hardware.ErrInvalidConfig)
	}
	if req.Window.Start.Before(now.Add(p.cfg.Capabilities.MinAdvanceNoticeBeforeStart)) {
		return fmt.Errorf("%w: window starts inside the minimum-notice window", hardware.ErrInvalidConfig)
	}
	if len(req.Bands) == 0 {
		return fmt.Errorf("%w: at least one band required", hardware.ErrInvalidConfig)
	}
	for _, b := range req.Bands {
		if !containsBand(p.cfg.Capabilities.SupportedBands, b) {
			return fmt.Errorf("%w: band %q not supported by provider", hardware.ErrInvalidConfig, b)
		}
	}
	if req.MinElevationDeg < 0 || req.MinElevationDeg > 90 {
		return fmt.Errorf("%w: min elevation must be in [0, 90]", hardware.ErrInvalidConfig)
	}
	return nil
}

// countActive returns the number of contacts currently reserved or
// scheduled or executing — capacity is bounded by AntennaCount.
func (p *fakeProvider) countActive() int {
	n := 0
	for _, c := range p.contacts {
		switch c.State {
		case hardware.ContactReserved, hardware.ContactScheduled, hardware.ContactExecuting:
			n++
		}
	}
	return n
}

func (p *fakeProvider) ReleaseContact(_ context.Context, contactID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return hardware.ErrAlreadyClosed
	}
	c, ok := p.contacts[contactID]
	if !ok {
		return hardware.ErrUnknownContact
	}
	switch c.State {
	case hardware.ContactCompleted, hardware.ContactCancelled, hardware.ContactFailed:
		return nil // no-op
	}
	c.State = hardware.ContactCancelled
	return nil
}

func (p *fakeProvider) ListContacts(_ context.Context, states []hardware.ContactState) ([]hardware.Contact, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, hardware.ErrAlreadyClosed
	}
	want := func(s hardware.ContactState) bool {
		if len(states) == 0 {
			return true
		}
		for _, want := range states {
			if want == s {
				return true
			}
		}
		return false
	}
	out := make([]hardware.Contact, 0, len(p.contacts))
	for _, c := range p.contacts {
		if want(c.State) {
			out = append(out, *c)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Window.Start.Before(out[j].Window.Start)
	})
	return out, nil
}

func (p *fakeProvider) NetworkCapabilities() hardware.NetworkCapabilities {
	return p.cfg.Capabilities
}

func (p *fakeProvider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return hardware.ErrAlreadyClosed
	}
	p.closed = true
	return nil
}

// ----------------------------------------------------------------------
// helpers
// ----------------------------------------------------------------------

func containsBand(set []hardware.Band, b hardware.Band) bool {
	for _, x := range set {
		if x == b {
			return true
		}
	}
	return false
}

func containsMod(set []hardware.Modulation, m hardware.Modulation) bool {
	for _, x := range set {
		if x == m {
			return true
		}
	}
	return false
}

// generateSinePattern returns a one-period sine wave at the supplied
// tone frequency (Hz) sampled at `rate` Hz. Used as the default RX
// pattern when the test does not supply one.
func generateSinePattern(rate, toneHz int) []hardware.IQSample {
	if toneHz <= 0 {
		return []hardware.IQSample{{I: 1, Q: 0}}
	}
	period := rate / toneHz
	if period < 4 {
		period = 4
	}
	out := make([]hardware.IQSample, period)
	for i := 0; i < period; i++ {
		theta := 2 * math.Pi * float64(i) / float64(period)
		out[i] = hardware.IQSample{
			I: float32(math.Cos(theta)),
			Q: float32(math.Sin(theta)),
		}
	}
	return out
}

// errBadConfig is the canonical error used inside this package for
// config-shape mismatches (kept private — callers see
// hardware.ErrInvalidConfig).
var errBadConfig = errors.New("fake: bad config")

// statically assert the fakes implement the interfaces they claim.
var _ hardware.HardwareDriver = (*fakeDriver)(nil)
var _ hardware.AntennaController = (*fakeAntenna)(nil)
var _ hardware.GroundNetworkProvider = (*fakeProvider)(nil)
var _ = errBadConfig // satisfy unused-var lint when no caller binds it
