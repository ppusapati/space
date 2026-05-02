package hardware_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"p9e.in/chetana/packages/hardware"
	"p9e.in/chetana/packages/hardware/fake"
)

// ============================================================================
// HardwareDriver conformance — exercises every interface method against
// the in-memory fake and asserts the documented error contract.
// ============================================================================

func newDriver(t *testing.T) hardware.HardwareDriver {
	t.Helper()
	d, err := fake.NewHardwareDriver(context.Background(), nil)
	if err != nil {
		t.Fatalf("NewHardwareDriver: %v", err)
	}
	return d
}

// goodTune returns a TuneRequest accepted by the fake's default
// capabilities. Reused across the driver tests.
func goodTune() hardware.TuneRequest {
	return hardware.TuneRequest{
		CenterHz:     437_500_000,
		SampleRateHz: 1_000_000,
		Band:         hardware.BandUHF,
		Modulation:   hardware.ModBPSK,
	}
}

func TestDriver_Tune_HappyPath(t *testing.T) {
	d := newDriver(t)
	defer d.Close()
	if err := d.Tune(context.Background(), goodTune()); err != nil {
		t.Fatalf("Tune: %v", err)
	}
}

func TestDriver_Tune_RejectsOutOfRangeFrequency(t *testing.T) {
	d := newDriver(t)
	defer d.Close()
	bad := goodTune()
	bad.CenterHz = 1 // below MinFreqHz
	err := d.Tune(context.Background(), bad)
	if !errors.Is(err, hardware.ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig; got %v", err)
	}
}

func TestDriver_Tune_RejectsUnsupportedSampleRate(t *testing.T) {
	d := newDriver(t)
	defer d.Close()
	bad := goodTune()
	bad.SampleRateHz = 12_345_678
	err := d.Tune(context.Background(), bad)
	if !errors.Is(err, hardware.ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig; got %v", err)
	}
}

func TestDriver_Tune_RejectsUnsupportedBand(t *testing.T) {
	d := newDriver(t)
	defer d.Close()
	bad := goodTune()
	bad.Band = "Ka" // not in the fake's defaults
	err := d.Tune(context.Background(), bad)
	if !errors.Is(err, hardware.ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig; got %v", err)
	}
}

func TestDriver_SetGain_RequiresPriorTune(t *testing.T) {
	d := newDriver(t)
	defer d.Close()
	if err := d.SetGain(context.Background(), 30); !errors.Is(err, hardware.ErrNotTuned) {
		t.Fatalf("expected ErrNotTuned; got %v", err)
	}
}

func TestDriver_SetGain_RejectsOutOfRange(t *testing.T) {
	d := newDriver(t)
	defer d.Close()
	if err := d.Tune(context.Background(), goodTune()); err != nil {
		t.Fatalf("Tune: %v", err)
	}
	if err := d.SetGain(context.Background(), 999); !errors.Is(err, hardware.ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig; got %v", err)
	}
}

func TestDriver_RxIQ_DeliversFrames(t *testing.T) {
	d := newDriver(t)
	defer d.Close()
	if err := d.Tune(context.Background(), goodTune()); err != nil {
		t.Fatalf("Tune: %v", err)
	}
	out := make(chan []hardware.IQSample, 4)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	doneCh := make(chan error, 1)
	go func() { doneCh <- d.RxIQ(ctx, 1024, out) }()

	// Drain a frame; the test passes when at least one frame arrives
	// before ctx fires.
	select {
	case frame := <-out:
		if len(frame) != 1024 {
			t.Errorf("got frame size %d, want 1024", len(frame))
		}
	case <-time.After(800 * time.Millisecond):
		t.Fatal("no RX frame within 800 ms")
	}
	cancel()
	<-doneCh
}

func TestDriver_RxIQ_RejectsZeroChunk(t *testing.T) {
	d := newDriver(t)
	defer d.Close()
	if err := d.Tune(context.Background(), goodTune()); err != nil {
		t.Fatalf("Tune: %v", err)
	}
	out := make(chan []hardware.IQSample, 1)
	err := d.RxIQ(context.Background(), 0, out)
	if !errors.Is(err, hardware.ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig; got %v", err)
	}
}

func TestDriver_RxIQ_RejectsConcurrentRX(t *testing.T) {
	d := newDriver(t)
	defer d.Close()
	if err := d.Tune(context.Background(), goodTune()); err != nil {
		t.Fatalf("Tune: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out1 := make(chan []hardware.IQSample, 4)
	go d.RxIQ(ctx, 256, out1)
	// Wait until the RX loop is committed.
	time.Sleep(20 * time.Millisecond)

	out2 := make(chan []hardware.IQSample, 4)
	err := d.RxIQ(ctx, 256, out2)
	if !errors.Is(err, hardware.ErrBusy) {
		t.Fatalf("expected ErrBusy; got %v", err)
	}
}

func TestDriver_TxIQ_RequiresTune(t *testing.T) {
	d := newDriver(t)
	defer d.Close()
	err := d.TxIQ(context.Background(), []hardware.IQSample{{I: 1}})
	if !errors.Is(err, hardware.ErrNotTuned) {
		t.Fatalf("expected ErrNotTuned; got %v", err)
	}
}

func TestDriver_TxIQ_RejectsEmptyBurst(t *testing.T) {
	d := newDriver(t)
	defer d.Close()
	if err := d.Tune(context.Background(), goodTune()); err != nil {
		t.Fatalf("Tune: %v", err)
	}
	if err := d.TxIQ(context.Background(), nil); !errors.Is(err, hardware.ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig; got %v", err)
	}
}

func TestDriver_TxIQ_RejectsOversizedBurst(t *testing.T) {
	d := newDriver(t)
	defer d.Close()
	tune := goodTune()
	if err := d.Tune(context.Background(), tune); err != nil {
		t.Fatalf("Tune: %v", err)
	}
	// Default fake bounds TX burst at 1 second of samples.
	burst := make([]hardware.IQSample, int(tune.SampleRateHz)+1)
	if err := d.TxIQ(context.Background(), burst); !errors.Is(err, hardware.ErrBufferOverflow) {
		t.Fatalf("expected ErrBufferOverflow; got %v", err)
	}
}

func TestDriver_TxStream_AbortedOnInputClose(t *testing.T) {
	d := newDriver(t)
	defer d.Close()
	if err := d.Tune(context.Background(), goodTune()); err != nil {
		t.Fatalf("Tune: %v", err)
	}
	in := make(chan []hardware.IQSample)
	close(in) // abort immediately
	err := d.TxStream(context.Background(), in)
	if !errors.Is(err, hardware.ErrTransmissionAborted) {
		t.Fatalf("expected ErrTransmissionAborted; got %v", err)
	}
}

func TestDriver_Close_Idempotent(t *testing.T) {
	d := newDriver(t)
	if err := d.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := d.Close(); !errors.Is(err, hardware.ErrAlreadyClosed) {
		t.Fatalf("second Close: expected ErrAlreadyClosed; got %v", err)
	}
}

func TestDriver_AfterClose_AllOpsFail(t *testing.T) {
	d := newDriver(t)
	d.Close()
	if err := d.Tune(context.Background(), goodTune()); !errors.Is(err, hardware.ErrAlreadyClosed) {
		t.Errorf("Tune: expected ErrAlreadyClosed; got %v", err)
	}
	if err := d.SetGain(context.Background(), 30); !errors.Is(err, hardware.ErrAlreadyClosed) {
		t.Errorf("SetGain: expected ErrAlreadyClosed; got %v", err)
	}
}

// ============================================================================
// AntennaController conformance
// ============================================================================

func newAntenna(t *testing.T) hardware.AntennaController {
	t.Helper()
	a, err := fake.NewAntennaController(context.Background(), nil)
	if err != nil {
		t.Fatalf("NewAntennaController: %v", err)
	}
	return a
}

func TestAntenna_SetAzEl_HappyPath(t *testing.T) {
	a := newAntenna(t)
	defer a.Close()
	target := hardware.AzEl{AzimuthDeg: 180, ElevationDeg: 45}
	if err := a.SetAzEl(context.Background(), target); err != nil {
		t.Fatalf("SetAzEl: %v", err)
	}
	got, err := a.GetAzEl(context.Background())
	if err != nil {
		t.Fatalf("GetAzEl: %v", err)
	}
	if got != target {
		t.Errorf("got %v; want %v", got, target)
	}
}

func TestAntenna_SetAzEl_RejectsOutsideEnvelope(t *testing.T) {
	a := newAntenna(t)
	defer a.Close()
	cases := []hardware.AzEl{
		{AzimuthDeg: -5, ElevationDeg: 30},
		{AzimuthDeg: 500, ElevationDeg: 30},
		{AzimuthDeg: 100, ElevationDeg: -10},
		{AzimuthDeg: 100, ElevationDeg: 200},
	}
	for _, target := range cases {
		err := a.SetAzEl(context.Background(), target)
		if !errors.Is(err, hardware.ErrInvalidPointing) {
			t.Errorf("target %v: expected ErrInvalidPointing; got %v", target, err)
		}
	}
}

func TestAntenna_SetTrack_WalksTrajectory(t *testing.T) {
	a := newAntenna(t)
	defer a.Close()
	now := time.Now()
	traj := []hardware.TrackPoint{
		{When: now.Add(20 * time.Millisecond), AzEl: hardware.AzEl{AzimuthDeg: 10, ElevationDeg: 20}},
		{When: now.Add(60 * time.Millisecond), AzEl: hardware.AzEl{AzimuthDeg: 20, ElevationDeg: 30}},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := a.SetTrack(ctx, traj); err != nil {
		t.Fatalf("SetTrack: %v", err)
	}
	got, err := a.GetAzEl(context.Background())
	if err != nil {
		t.Fatalf("GetAzEl: %v", err)
	}
	if got != traj[len(traj)-1].AzEl {
		t.Errorf("final position %v; want %v", got, traj[len(traj)-1].AzEl)
	}
}

func TestAntenna_SetTrack_RejectsEmpty(t *testing.T) {
	a := newAntenna(t)
	defer a.Close()
	if err := a.SetTrack(context.Background(), nil); !errors.Is(err, hardware.ErrInvalidTrack) {
		t.Fatalf("expected ErrInvalidTrack; got %v", err)
	}
}

func TestAntenna_SetTrack_RejectsNonMonotonicTime(t *testing.T) {
	a := newAntenna(t)
	defer a.Close()
	now := time.Now()
	bad := []hardware.TrackPoint{
		{When: now.Add(50 * time.Millisecond), AzEl: hardware.AzEl{AzimuthDeg: 10, ElevationDeg: 20}},
		{When: now.Add(40 * time.Millisecond), AzEl: hardware.AzEl{AzimuthDeg: 20, ElevationDeg: 30}},
	}
	if err := a.SetTrack(context.Background(), bad); !errors.Is(err, hardware.ErrInvalidTrack) {
		t.Fatalf("expected ErrInvalidTrack; got %v", err)
	}
}

func TestAntenna_Park_DrivesToParkPosition(t *testing.T) {
	a := newAntenna(t)
	defer a.Close()
	if err := a.Park(context.Background()); err != nil {
		t.Fatalf("Park: %v", err)
	}
	got, _ := a.GetAzEl(context.Background())
	want := hardware.AzEl{AzimuthDeg: 0, ElevationDeg: 90}
	if got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}

func TestAntenna_Stow_DrivesToStowPosition(t *testing.T) {
	a := newAntenna(t)
	defer a.Close()
	if err := a.Stow(context.Background()); err != nil {
		t.Fatalf("Stow: %v", err)
	}
	got, _ := a.GetAzEl(context.Background())
	want := hardware.AzEl{AzimuthDeg: 0, ElevationDeg: 0}
	if got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}

func TestAntenna_Close_Idempotent(t *testing.T) {
	a := newAntenna(t)
	if err := a.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := a.Close(); !errors.Is(err, hardware.ErrAlreadyClosed) {
		t.Fatalf("second Close: expected ErrAlreadyClosed; got %v", err)
	}
}

// ============================================================================
// GroundNetworkProvider conformance
// ============================================================================

func newProvider(t *testing.T) hardware.GroundNetworkProvider {
	t.Helper()
	p, err := fake.NewGroundNetworkProvider(context.Background(), &fake.GroundNetworkProviderConfig{
		Capabilities: hardware.NetworkCapabilities{
			ProviderName:                "fake",
			AntennaCount:                2,
			SupportedBands:              []hardware.Band{hardware.BandUHF, hardware.BandS},
			MinFreqHz:                   100_000_000,
			MaxFreqHz:                   3_000_000_000,
			MinAdvanceNoticeBeforeStart: 0,
		},
	})
	if err != nil {
		t.Fatalf("NewGroundNetworkProvider: %v", err)
	}
	return p
}

func futureWindow() hardware.TimeWindow {
	now := time.Now()
	return hardware.TimeWindow{
		Start: now.Add(time.Minute),
		End:   now.Add(11 * time.Minute),
	}
}

func TestProvider_AllocateContact_HappyPath(t *testing.T) {
	p := newProvider(t)
	defer p.Close()
	c, err := p.AllocateContact(context.Background(), hardware.ContactRequest{
		SatelliteID:     "ISS",
		Window:          futureWindow(),
		MinElevationDeg: 10,
		Bands:           []hardware.Band{hardware.BandUHF},
	})
	if err != nil {
		t.Fatalf("AllocateContact: %v", err)
	}
	if c.ContactID == "" {
		t.Error("empty ContactID")
	}
	if c.State != hardware.ContactReserved {
		t.Errorf("state=%q; want reserved", c.State)
	}
}

func TestProvider_AllocateContact_ValidationErrors(t *testing.T) {
	p := newProvider(t)
	defer p.Close()
	cases := map[string]hardware.ContactRequest{
		"empty SatelliteID": {Window: futureWindow(), Bands: []hardware.Band{hardware.BandUHF}},
		"end before start": {SatelliteID: "x", Window: hardware.TimeWindow{
			Start: time.Now().Add(2 * time.Hour), End: time.Now().Add(time.Hour),
		}, Bands: []hardware.Band{hardware.BandUHF}},
		"no bands": {SatelliteID: "x", Window: futureWindow()},
		"unsupported band": {SatelliteID: "x", Window: futureWindow(),
			Bands: []hardware.Band{hardware.BandX}},
		"bad min elev": {SatelliteID: "x", Window: futureWindow(),
			Bands: []hardware.Band{hardware.BandUHF}, MinElevationDeg: 999},
	}
	for name, req := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := p.AllocateContact(context.Background(), req)
			if !errors.Is(err, hardware.ErrInvalidConfig) {
				t.Errorf("%q: expected ErrInvalidConfig; got %v", name, err)
			}
		})
	}
}

func TestProvider_AllocateContact_NoCapacity(t *testing.T) {
	p := newProvider(t) // AntennaCount=2
	defer p.Close()
	for i := 0; i < 2; i++ {
		_, err := p.AllocateContact(context.Background(), hardware.ContactRequest{
			SatelliteID: "ISS", Window: futureWindow(),
			Bands: []hardware.Band{hardware.BandUHF},
		})
		if err != nil {
			t.Fatalf("alloc %d: %v", i, err)
		}
	}
	_, err := p.AllocateContact(context.Background(), hardware.ContactRequest{
		SatelliteID: "ISS", Window: futureWindow(),
		Bands: []hardware.Band{hardware.BandUHF},
	})
	if !errors.Is(err, hardware.ErrNoCapacity) {
		t.Fatalf("expected ErrNoCapacity; got %v", err)
	}
}

func TestProvider_ReleaseContact_UnknownID(t *testing.T) {
	p := newProvider(t)
	defer p.Close()
	if err := p.ReleaseContact(context.Background(), "nonexistent"); !errors.Is(err, hardware.ErrUnknownContact) {
		t.Fatalf("expected ErrUnknownContact; got %v", err)
	}
}

func TestProvider_ReleaseContact_TransitionsToCancelled(t *testing.T) {
	p := newProvider(t)
	defer p.Close()
	c, _ := p.AllocateContact(context.Background(), hardware.ContactRequest{
		SatelliteID: "ISS", Window: futureWindow(),
		Bands: []hardware.Band{hardware.BandUHF},
	})
	if err := p.ReleaseContact(context.Background(), c.ContactID); err != nil {
		t.Fatalf("Release: %v", err)
	}
	contacts, _ := p.ListContacts(context.Background(), nil)
	if len(contacts) != 1 || contacts[0].State != hardware.ContactCancelled {
		t.Errorf("expected one cancelled contact; got %+v", contacts)
	}
}

func TestProvider_ListContacts_FilterByState(t *testing.T) {
	p := newProvider(t)
	defer p.Close()
	c1, _ := p.AllocateContact(context.Background(), hardware.ContactRequest{
		SatelliteID: "ISS", Window: futureWindow(), Bands: []hardware.Band{hardware.BandUHF},
	})
	_, _ = p.AllocateContact(context.Background(), hardware.ContactRequest{
		SatelliteID: "TIANGONG", Window: futureWindow(), Bands: []hardware.Band{hardware.BandUHF},
	})
	p.ReleaseContact(context.Background(), c1.ContactID)

	all, _ := p.ListContacts(context.Background(), nil)
	if len(all) != 2 {
		t.Errorf("all: got %d, want 2", len(all))
	}
	reserved, _ := p.ListContacts(context.Background(), []hardware.ContactState{hardware.ContactReserved})
	if len(reserved) != 1 {
		t.Errorf("reserved-only: got %d, want 1", len(reserved))
	}
	cancelled, _ := p.ListContacts(context.Background(), []hardware.ContactState{hardware.ContactCancelled})
	if len(cancelled) != 1 {
		t.Errorf("cancelled-only: got %d, want 1", len(cancelled))
	}
}

func TestProvider_NetworkCapabilities(t *testing.T) {
	p := newProvider(t)
	defer p.Close()
	caps := p.NetworkCapabilities()
	if caps.AntennaCount != 2 {
		t.Errorf("AntennaCount = %d; want 2", caps.AntennaCount)
	}
}

// ============================================================================
// Registry tests
// ============================================================================

func TestRegistry_RegisterAndLookup(t *testing.T) {
	reg := hardware.NewRegistry()
	if err := fake.Register(reg); err != nil {
		t.Fatalf("fake.Register: %v", err)
	}
	d, err := reg.NewHardwareDriver(context.Background(), fake.AdapterName, nil)
	if err != nil {
		t.Fatalf("NewHardwareDriver: %v", err)
	}
	d.Close()
}

func TestRegistry_RejectsDuplicateName(t *testing.T) {
	reg := hardware.NewRegistry()
	if err := reg.RegisterHardwareDriver("a", fake.NewHardwareDriver); err != nil {
		t.Fatalf("first register: %v", err)
	}
	err := reg.RegisterHardwareDriver("a", fake.NewHardwareDriver)
	if !errors.Is(err, hardware.ErrDuplicateAdapter) {
		t.Errorf("expected ErrDuplicateAdapter; got %v", err)
	}
}

func TestRegistry_RejectsEmptyName(t *testing.T) {
	reg := hardware.NewRegistry()
	if err := reg.RegisterHardwareDriver("", fake.NewHardwareDriver); !errors.Is(err, hardware.ErrInvalidAdapterName) {
		t.Errorf("expected ErrInvalidAdapterName; got %v", err)
	}
}

func TestRegistry_RejectsNilFactory(t *testing.T) {
	reg := hardware.NewRegistry()
	if err := reg.RegisterHardwareDriver("a", nil); err == nil {
		t.Error("expected error for nil factory")
	}
}

func TestRegistry_UnknownAdapter(t *testing.T) {
	reg := hardware.NewRegistry()
	_, err := reg.NewHardwareDriver(context.Background(), "nope", nil)
	if !errors.Is(err, hardware.ErrUnknownAdapter) {
		t.Errorf("driver: expected ErrUnknownAdapter; got %v", err)
	}
	_, err = reg.NewAntennaController(context.Background(), "nope", nil)
	if !errors.Is(err, hardware.ErrUnknownAdapter) {
		t.Errorf("antenna: expected ErrUnknownAdapter; got %v", err)
	}
	_, err = reg.NewGroundNetworkProvider(context.Background(), "nope", nil)
	if !errors.Is(err, hardware.ErrUnknownAdapter) {
		t.Errorf("provider: expected ErrUnknownAdapter; got %v", err)
	}
}

func TestRegistry_NamesSorted(t *testing.T) {
	reg := hardware.NewRegistry()
	for _, n := range []string{"zeta", "alpha", "mu"} {
		_ = reg.RegisterHardwareDriver(n, fake.NewHardwareDriver)
	}
	got := reg.HardwareDriverNames()
	want := []string{"alpha", "mu", "zeta"}
	for i, n := range want {
		if got[i] != n {
			t.Errorf("names[%d]=%q; want %q", i, got[i], n)
		}
	}
}
