// Package test holds end-to-end tests that exercise the hardware
// abstraction the way real services use it: register the fake
// adapters in a fresh registry, then drive a complete pass workflow
// (allocate contact → tune driver → track antenna → release) against
// the resulting handles.
//
// → REQ-FUNC-GS-HW-001 / -002 / -003
// → TASK-P0-HW-001 verification: integration coverage.
package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"p9e.in/chetana/packages/hardware"
	"p9e.in/chetana/packages/hardware/fake"
)

// TestEndToEnd_RegistryDrivenPass simulates a single ground-station
// pass:
//
//   1. Build a registry, register fakes for all three interfaces.
//   2. Look up each adapter by name (the way a real service would
//      from configuration).
//   3. Allocate a contact via the network provider.
//   4. Tune the driver, set gain, kick off RX.
//   5. Walk a short tracking trajectory on the antenna.
//   6. Release the contact, close every handle.
//
// A failure at any step indicates a contract regression in the
// abstraction layer.
func TestEndToEnd_RegistryDrivenPass(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	reg := hardware.NewRegistry()
	if err := fake.Register(reg); err != nil {
		t.Fatalf("fake.Register: %v", err)
	}

	// 1. Look up adapters from "configuration" (just constants here).
	driver, err := reg.NewHardwareDriver(ctx, fake.AdapterName, nil)
	if err != nil {
		t.Fatalf("NewHardwareDriver: %v", err)
	}
	defer driver.Close()

	antenna, err := reg.NewAntennaController(ctx, fake.AdapterName, nil)
	if err != nil {
		t.Fatalf("NewAntennaController: %v", err)
	}
	defer antenna.Close()

	provider, err := reg.NewGroundNetworkProvider(ctx, fake.AdapterName, &fake.GroundNetworkProviderConfig{
		Capabilities: hardware.NetworkCapabilities{
			ProviderName:                "fake",
			AntennaCount:                1,
			SupportedBands:              []hardware.Band{hardware.BandUHF},
			MinFreqHz:                   400_000_000,
			MaxFreqHz:                   500_000_000,
			MinAdvanceNoticeBeforeStart: 0,
		},
	})
	if err != nil {
		t.Fatalf("NewGroundNetworkProvider: %v", err)
	}
	defer provider.Close()

	// 2. Allocate the contact.
	now := time.Now()
	c, err := provider.AllocateContact(ctx, hardware.ContactRequest{
		SatelliteID: "ISS",
		Window: hardware.TimeWindow{
			Start: now.Add(50 * time.Millisecond),
			End:   now.Add(time.Second),
		},
		MinElevationDeg: 10,
		Bands:           []hardware.Band{hardware.BandUHF},
	})
	if err != nil {
		t.Fatalf("AllocateContact: %v", err)
	}
	if c.State != hardware.ContactReserved {
		t.Errorf("contact state = %q; want reserved", c.State)
	}

	// 3. Tune the driver, set gain.
	if err := driver.Tune(ctx, hardware.TuneRequest{
		CenterHz:     437_500_000,
		SampleRateHz: 1_000_000,
		Band:         hardware.BandUHF,
		Modulation:   hardware.ModBPSK,
	}); err != nil {
		t.Fatalf("Tune: %v", err)
	}
	if err := driver.SetGain(ctx, 30); err != nil {
		t.Fatalf("SetGain: %v", err)
	}

	// 4. Start RX in background; verify at least one frame.
	rxCtx, rxCancel := context.WithCancel(ctx)
	defer rxCancel()
	frames := make(chan []hardware.IQSample, 4)
	rxDone := make(chan error, 1)
	go func() { rxDone <- driver.RxIQ(rxCtx, 256, frames) }()

	select {
	case f := <-frames:
		if len(f) != 256 {
			t.Errorf("frame size = %d; want 256", len(f))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("no RX frame within 500 ms")
	}

	// 5. Walk a small tracking trajectory.
	traj := []hardware.TrackPoint{
		{When: time.Now().Add(20 * time.Millisecond), AzEl: hardware.AzEl{AzimuthDeg: 100, ElevationDeg: 30}},
		{When: time.Now().Add(60 * time.Millisecond), AzEl: hardware.AzEl{AzimuthDeg: 130, ElevationDeg: 45}},
	}
	if err := antenna.SetTrack(ctx, traj); err != nil {
		t.Fatalf("SetTrack: %v", err)
	}
	got, _ := antenna.GetAzEl(ctx)
	if got != traj[len(traj)-1].AzEl {
		t.Errorf("antenna at %v; want %v", got, traj[len(traj)-1].AzEl)
	}

	rxCancel()
	if err := <-rxDone; err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("RxIQ exit: %v", err)
	}

	// 6. Release the contact; assert state.
	if err := provider.ReleaseContact(ctx, c.ContactID); err != nil {
		t.Fatalf("ReleaseContact: %v", err)
	}
	listed, _ := provider.ListContacts(ctx, nil)
	if len(listed) != 1 || listed[0].State != hardware.ContactCancelled {
		t.Errorf("post-release listing: %+v", listed)
	}
}

// TestEndToEnd_AntennaParkAndStow exercises the maintenance ops in
// a single test — operators rely on Park/Stow always succeeding, even
// after a tracking abort.
func TestEndToEnd_AntennaParkAndStow(t *testing.T) {
	ctx := context.Background()
	reg := hardware.NewRegistry()
	if err := fake.Register(reg); err != nil {
		t.Fatalf("fake.Register: %v", err)
	}
	a, err := reg.NewAntennaController(ctx, fake.AdapterName, nil)
	if err != nil {
		t.Fatalf("NewAntennaController: %v", err)
	}
	defer a.Close()

	// Move somewhere arbitrary.
	if err := a.SetAzEl(ctx, hardware.AzEl{AzimuthDeg: 200, ElevationDeg: 60}); err != nil {
		t.Fatalf("SetAzEl: %v", err)
	}
	// Park MUST drive to (0, 90).
	if err := a.Park(ctx); err != nil {
		t.Fatalf("Park: %v", err)
	}
	got, _ := a.GetAzEl(ctx)
	if got != (hardware.AzEl{AzimuthDeg: 0, ElevationDeg: 90}) {
		t.Errorf("park position %v; want (0, 90)", got)
	}
	// Stow MUST drive to (0, 0).
	if err := a.Stow(ctx); err != nil {
		t.Fatalf("Stow: %v", err)
	}
	got, _ = a.GetAzEl(ctx)
	if got != (hardware.AzEl{AzimuthDeg: 0, ElevationDeg: 0}) {
		t.Errorf("stow position %v; want (0, 0)", got)
	}
}
