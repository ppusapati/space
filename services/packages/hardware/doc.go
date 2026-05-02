// Package hardware defines the abstraction interfaces every chetana
// ground-station service uses to talk to physical hardware:
//
//   • HardwareDriver        — software-defined radios (USRP, RTL-SDR,
//                              custom).
//   • AntennaController     — rotators (Hamlib rotctld, GS-232,
//                              custom).
//   • GroundNetworkProvider — networks of dishes (own-dish, AWS
//                              Ground Station, KSAT/SSC).
//
// → REQ-FUNC-GS-HW-001 (HardwareDriver)
// → REQ-FUNC-GS-HW-002 (AntennaController)
// → REQ-FUNC-GS-HW-003 (GroundNetworkProvider)
// → design.md §4.4
//
// The interfaces live in chetana-platform so they can be referenced
// by neutral code (mocks, tests, schedulers). Concrete adapters that
// touch ITAR-controlled defense articles (USRP firmware, custom
// modulators, ITAR-controlled antenna geometries) live in
// chetana-defense per the two-repo posture documented in
// plan/design.md §2.
//
// The three interface contracts are deliberately narrow — they only
// expose the operations the scheduling + telemetry pipelines need.
// Vendor-specific tuning (e.g. USRP timing knobs) lives behind
// adapter-private configuration loaded at registry initialisation.
//
// # Selecting an adapter at runtime
//
// Services use the Registry to look up an adapter by name from
// configuration:
//
//	reg := hardware.NewRegistry()
//	hardware.RegisterUHD(reg)            // chetana-defense package
//	hardware.RegisterRTL(reg)            // chetana-defense package
//	hardware.RegisterFake(reg)           // this package, for tests
//
//	driver, err := reg.NewHardwareDriver(ctx, cfg.SDR.Name, cfg.SDR.Config)
//	if err != nil { return err }
//	defer driver.Close()
//
// Test code uses the in-memory fake from package hardware/fake to
// exercise the full state machine without touching real hardware.
package hardware
