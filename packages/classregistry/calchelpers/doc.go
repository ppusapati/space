// Package calchelpers is the shared library for Layer 3 calculation
// services under business/<module>/<calculation>/.
//
// Phase F's three-layer pattern (see
// docs/GENERIC_INDUSTRY_IMPLEMENTATION.md §4) puts calculations in
// their own typed services — one per calculation, not per vertical.
// Each service reads the entity's class through the class registry
// and branches internally on class when the algorithm varies. This
// package provides the primitives every calculation needs so the
// 30+ Layer 3 services in the roadmap don't each reinvent them.
//
// # What lives here
//
//   - ClassDispatcher — the class-aware branching primitive. A
//     calculation registers a variant per class; at runtime it
//     resolves the variant by the entity's class and invokes it.
//     Unknown classes produce a typed error instead of silently
//     defaulting to a variant that shouldn't apply.
//
//   - Conformance helpers — test utilities asserting that a
//     calculation service registers every class declared with its
//     calculation in the class registry's `processes:` list, so
//     drift is caught at test time rather than runtime.
//
// # What does NOT live here
//
//   - The calculations themselves (OEE, BOM explosion, depreciation,
//     …). Those are separate packages under business/<module>/<calc>/.
//   - Per-service proto types or repository code. Each calculation
//     owns its own.
//   - The class registry (which is the parent package) or the
//     validator (also parent).
package calchelpers
