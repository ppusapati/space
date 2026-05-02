# `packages/classregistry/contract` — ADR-0003 adapter contract harness

This package implements [Roadmap D.3](../../../docs/BI_CONSOLIDATION_ROADMAP.md#phase-d--follow-up-tasks-flagged-by-adrs): run the same black-box test suite against both the in-process and ConnectRPC adapters of an ADR-0003 port and catch divergence between the two paths.

## Why

Every Phase F port exposes the same `Client` interface twice:

- `NewInProcClient(svc)` — wraps the internal service directly. Monolith uses this; zero wire overhead.
- `NewConnectClient(conn)` — wraps the generated ConnectRPC client. Split deployments use this; crosses the wire.

Callers are supposed to be unable to tell them apart. But translation bugs creep in silently — a proto field added without an adapter update, an enum missing from a lookup table, a filter applied one way in `inproc.go` and another way in `connect.go`. The harness turns that silent drift into a hard CI failure.

## Two suites

### `ClassRegistrySuite` — the uniform classregistry surface

Every Phase F port has a uniform pair:

```go
ListClasses(ctx) ([]*ClassSummary, error)
GetClassSchema(ctx, class) (*ClassDefinition, error)
```

Both adapters forward to the same `classregistry.Registry`. Divergence is possible in the `pb_translate.go` round-trip (proto→port and registry→proto). `ClassRegistrySuite` runs:

- `ListClasses_equivalence` — both adapters return the same class set.
- `GetClassSchema_known_equivalence` — same resolved definition for a known class.
- `GetClassSchema_unknown_both_error` — both fail for an unknown class (error TEXT may differ; only presence is compared).

### `EntitySuite` — port-specific `Get` / `List` parity

Every Phase F port also has:

```go
Get(ctx, id) (*Entity, error)
List(ctx, filter) ([]*Entity, int32, error)
```

but `Entity` and `filter` are different per port. `EntitySuite` is generic over both:

- `Get_equivalence` for each seeded id — `ProjectIdentity(entity) → (id, class)` is compared pairwise.
- `List_empty_filter_equivalence` — symmetric difference on the id sets must be empty.
- `List_class_filter_equivalence` — when the caller supplies a `ClassFilterValue`, both adapters are expected to return the same filtered id set AND every returned entity must carry that class on both paths.
- An optional `DiffFn` hook lets callers add per-field comparison beyond `id+class`.

## Adoption recipe (per port, ~30 lines)

Place this in `<port>/client/contract_test.go`:

```go
package client

import (
    "testing"

    "p9e.in/chetana/packages/classregistry/contract"
    // plus your port's service + registry wiring
)

func TestContract_MyPort(t *testing.T) {
    // 1. Build a test classregistry.Registry seeded with the domain's
    //    yaml — loader_test.go patterns in packages/classregistry/
    //    show how.
    reg := buildTestRegistry(t)

    // 2. Build an in-proc service instance wired to the registry.
    svc := buildTestService(t, reg)
    inproc := NewInProcClient(svc)

    // 3. Build a fake Connect client that reads from the SAME registry
    //    so the two paths share a single source of truth. Mirror the
    //    fakeDatasetConnectClient pattern in
    //    core/bi/dataset/client/connect_test.go.
    fakeConn := buildFakeConnectClient(t, reg)
    connect := NewConnectClient(fakeConn)

    // 4. Wire the classregistry suite.
    crSuite := contract.ClassRegistrySuite[ClassSummary, ClassDefinition]{
        Name:              "business/<domain>/<port>",
        InProc:            inproc,
        Connect:           connect,
        ProjectSummary:    projectClassSummary,
        ProjectDefinition: projectClassDefinition,
        LookupClass:       "<known-class-in-yaml>",
        UnknownClass:      "nonexistent_class",
    }
    crSuite.Run(contract.AdaptT(t))

    // 5. Wire the entity suite.
    entSuite := contract.EntitySuite[Entity, ListFilter]{
        Name:             "business/<domain>/<port>",
        InProc:           inproc,
        Connect:          connect,
        SeededIDs:        []string{"seed-1", "seed-2"},
        ProjectIdentity:  func(e *Entity) (string, string) { return e.ID, e.Class },
        MakeEmptyFilter:  func() ListFilter { return ListFilter{Limit: 1000} },
        MakeClassFilter:  func(class string) ListFilter { return ListFilter{Class: class, Limit: 1000} },
        ClassFilterValue: "<known-class-in-yaml>",
    }
    entSuite.Run(contract.AdaptT(t))
}

func projectClassSummary(s *ClassSummary) contract.ClassSummary {
    return contract.ClassSummary{
        Name:         s.Name,
        Label:        s.Label,
        Description:  s.Description,
        Industries:   s.Industries,
        HasProcesses: s.HasProcesses,
    }
}

func projectClassDefinition(d *ClassDefinition) contract.ClassDefinition {
    attrs := make(map[string]contract.AttributeDefinition, len(d.Attributes))
    for k, v := range d.Attributes {
        attrs[k] = contract.AttributeDefinition{
            Kind:        v.Kind,
            Required:    v.Required,
            Min:         v.Min,
            Max:         v.Max,
            Values:      v.Values,
            Lookup:      v.Lookup,
            Pattern:     v.Pattern,
            Description: v.Description,
        }
    }
    return contract.ClassDefinition{
        Domain:           d.Domain,
        Name:             d.Name,
        Label:            d.Label,
        Description:      d.Description,
        Industries:       d.Industries,
        Attributes:       attrs,
        ComplianceChecks: d.ComplianceChecks,
        CapacityMetrics:  d.CapacityMetrics,
        Processes:        d.Processes,
    }
}
```

The projection functions are one-liners per port because every port's `ClassSummary`/`ClassDefinition` carries exactly these fields under exactly these names.

## What the harness does NOT test

- **Business logic** — if a port's `Create` method has a bug, that bug is present on both adapters and the harness passes. Per-service unit tests cover correctness; this covers *adapter-shape parity*.
- **Transport concerns** — retries, codecs, compression, server timeouts. Those belong in ConnectRPC integration tests.
- **Error text parity** — we only require both adapters to fail on the same conditions, not fail with identical messages. Transport wrapping (`connect.CodeNotFound` envelope on the wire side, typed `errors.NotFound` on in-proc) is legitimate.

## Why TB, not *testing.T

The harness accepts a narrow `TB` interface instead of `*testing.T`. This lets the harness's own regression tests substitute a capturing stub that records failures without propagating them. `contract.AdaptT(t)` lifts a `*testing.T` into `TB` for production callers; the lift is a one-liner at every call site.

## Files

| File | Purpose |
|---|---|
| `doc.go` | Package overview |
| `types.go` | Canonical `ClassSummary` / `ClassDefinition` / `AttributeDefinition` comparison shapes |
| `classregistry.go` | `ClassRegistrySuite` — `ListClasses` + `GetClassSchema` equivalence |
| `entity.go` | `EntitySuite` — `Get` + `List` equivalence |
| `classregistry_test.go` | Harness regression tests using the recordingTB stub |
