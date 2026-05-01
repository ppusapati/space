// Package contract provides black-box equivalence tests that run the
// same scenarios through both adapter shapes of an ADR-0003 port
// (in-process and ConnectRPC) and assert that the two adapters return
// identical results.
//
// # Why this exists
//
// ADR-0003 ports are exposed through two adapters:
//
//   - The in-process adapter (inproc.go) — wraps the internal service
//     directly, zero serialization overhead.
//   - The ConnectRPC adapter (connect.go) — wraps a generated client,
//     crosses a wire boundary.
//
// Both adapters must satisfy the same Client interface and must be
// observably indistinguishable from the perspective of cross-service
// callers. If they diverge — a proto field not translated, an enum
// value missing from a lookup table, a filter applied inconsistently —
// the in-process path (typically monolith) silently drifts from the
// split-deployment path (production).
//
// This package is the harness that prevents that drift.
//
// # What it provides
//
// Two shapes of contract tests:
//
//  1. [ClassRegistryContract] — runs [ListClasses] and [GetClassSchema]
//     through both adapters of any Phase F classregistry port and
//     asserts pairwise equivalence. Every Phase F port (~64 services)
//     has this surface, so one helper covers most of the codebase.
//
//  2. [EntityContract] — a per-port helper that runs [Get] and [List]
//     through both adapters with the same seed data and compares the
//     returned entities field-by-field. Callers supply a small
//     generator (Go generics) because entity types differ per port.
//
// Tests that use the harness are placed in the port's own package
// (`<svc>/client/contract_test.go`) so they fail at build-time alongside
// any adapter changes. The harness itself is stateless — each test
// constructs a fresh adapter pair.
//
// # Layering
//
// This package depends on:
//
//   - packages/classregistry — the Registry interface the in-proc path uses.
//   - packages/classregistry/api/v1 — the proto types the connect path uses.
//
// It does NOT depend on any specific service. Ports wire it into their
// own test files, supplying the service-specific adapters.
//
// # Non-goals
//
// The harness does not:
//
//   - Validate business logic beyond what both adapters see identically.
//   - Exercise integration against a running database.
//   - Test the ConnectRPC transport itself (retries, codecs, compression).
//     Those belong in connectrpc integration tests.
//
// It only verifies: "given the same input, the two adapter shapes
// produce the same output."
package contract
