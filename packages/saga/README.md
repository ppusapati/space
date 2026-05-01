# packages/saga — Saga Orchestration Engine

## Status (2026-04-19): BUILDS CLEAN

The saga subtree lives under `packages/saga` as a regular subtree of the
`packages/` module. `go build ./packages/saga/...` succeeds across every
subpackage — root, engine, events, executor, connector, orchestrator,
compensation, timeout, repository, models, and all 19 domain
`sagas/<domain>` packages.

## What B.8 (roadmap 2026-04-19) landed

A single pass through the saga subtree resolved every issue that kept it
from building:

### Module boundary

1. **`go.mod` + `go.sum` removed.** No more nested module; saga joins the
   parent `packages/` module.
2. **`go.work`** has a note explaining saga's inclusion via the parent
   module — no separate use-directive.

### Cycle break

3. **Root ↔ compensation import cycle broken (Option B).** The FX
   composition that used to live in `saga/fx.go` (imported every
   subpackage) moved into a new subpackage `saga/engine/fx.go`. Root
   `saga` now holds only contracts: `interfaces.go`, `config.go`,
   `errors.go`, `registry.go`.

### Contract alignment (the big work)

4. **`saga/interfaces.go` imports `saga/models`** and re-exports every
   model type via `type Foo = models.Foo` aliasing. `saga.StepDefinition`
   and `models.StepDefinition` now refer to the same underlying type —
   the 182 internal imports that mixed both forms all resolve.
5. **`SagaEventPublisher` interface aligned to `EventPublisherImpl`'s
   actual method signatures** — took `*SagaExecution` instead of
   `(sagaID, stepNum)` because the publisher needs the execution context
   to build the event payload. Added `PublishSagaStarted`.
6. **`SagaRepository` extended** — added `GetExecution`,
   `CreateExecution`, `UpdateExecution` alongside the existing `GetByID`,
   `GetBySagaID`. Call sites use whichever name reads cleaner.
7. **`SagaExecutionLogRepository` extended** — added
   `GetExecutionLog`, `CreateExecutionLog`, `UpdateExecutionLog`.
8. **`SagaCompensationEngine` interface aligned to
   `CompensationEngineImpl`** — takes richer arguments
   (`*SagaExecution`, `[]*StepDefinition`) so the orchestrator doesn't
   re-fetch state when starting compensation.
9. **`NewSagaError`** made variadic to accept both the 4-arg saga-level
   and 6-arg step-level shapes used around the codebase.
10. **`saga/registry.go`** added — process-wide
    `GlobalSagaRegistry` singleton (`Register`, `Get`, `MustGet`,
    `Types`) referenced by all 19 `sagas/<domain>/fx.go` init hooks.

### Model + type shape fixes

11. **`models.SagaExecutionInput`** — added `TenantID`, `CompanyID`,
    `BranchID`, `TimeoutSeconds` so batch-processor calls can target a
    tenant without a request context.
12. **`models.StepExecution`** — added `ErrorMessage`, `ExecutedAt` (as
    `time.Time`, not `*time.Time` — call sites assign `time.Now()`
    directly), and `ExecutionTime` (as `int64`, mirrors
    `ExecutionTimeMs`).
13. **`models.StepResult.Result`** changed from `interface{}` to
    `[]byte`. Callers hand pre-marshalled JSON; StepExecution.Result
    can be populated without a type assertion.
14. **`models.SagaEvent.Data`** changed from `[]byte` to
    `map[string]interface{}` so call sites assign the typed payload
    directly; JSON encoding happens at the Kafka boundary.
15. **`config.go` syntax fix** — struct-literal defaults (`int32 = 60`)
    aren't valid Go; moved defaults into `NewDefaultConfig()`.

### Enum renames (callers were using nonexistent names)

16. **`models.SagaEventType*` → `models.SagaEvent*`** (8 call sites in
    `events/event_publisher.go`).
17. **`models.StepStatusSucceeded` → `models.StepStatusSuccess`** (across
    `executor`, `orchestrator`, `compensation`, and their test files).
18. **`models.StepStatus` → `models.StepExecutionStatus`** (in
    `compensation/compensation_log_repository.go`).

### Stdlib + package path fixes

19. **`connector/rpc_connector.go`** — `http.NewReadCloser` (doesn't
    exist) → local `NewReadCloser` helper returning `io.ReadCloser`
    (which does).
20. **`mesh.go`** — `math.Rand` → `math/rand/v2`.
21. **`executor/rpc_connector.go`** — unused `saga` import dropped.
22. **`loadbalancer/algorithms/round_robin.go`** — stray mid-file
    `import "time"` moved to the top import block.
23. **`mesh/api/api.go`** — stray mid-file `import "context"` moved.

### Package-name typos

24. **`sagas/critical_sagas_test.go`** — `package saga` → `package sagas`.
25. **`sagas/supply-chain/supply_chain_sagas_test.go`** —
    `package supply_chain` → `package supplychain`.
26. **`sagas/land_acquisition_saga_example.go`** — old `kosha/saga`
    import path rewritten; file gated under `//go:build saga_examples`
    because its StepDefinition field references don't match the
    canonical shape.

### Domain constructor-name mismatches

`fx.go` files referenced constructor names that didn't exist. Renamed to
match actual symbols:

| fx.go referenced                | actual constructor                                       |
|---------------------------------|-----------------------------------------------------------|
| NewProjectInitiationSaga        | NewConstructionProjectInitiationSaga (construction)       |
| NewProductionOrderExecutionSaga | NewProductionOrderSaga (manufacturing)                    |
| NewBankReconciliationSaga       | NewBankReconciliationMultiSaga (banking)                  |
| NewSparepartsSaga               | NewSparePartsSaga (warranty)                              |
| NewMerchandisePlaningSaga       | NewMerchandisePlanningallocation (retail)                 |
| NewInboundLogisticsSaga         | NewInboundLogisticsReceivingSaga (supply-chain)           |
| NewWarehouseOpsSaga             | NewWarehouseOperationsManagementSaga (supply-chain)       |
| NewThreePLCoordinationSaga      | NewThirdPartyLogisticsCoordinationSaga (supply-chain)     |
| NewOrderFulfillmentSaga         | NewOrderFulfillmentOutboundLogisticsSaga (supply-chain)   |
| NewDistributionCenterSaga       | NewDistributionCenterOperationsSaga (supply-chain)        |
| NewRouteOptimizationSaga        | NewRouteOptimizationSchedulingSaga (supply-chain)         |
| NewSupplyChainVisibilitySaga    | NewSupplyChainVisibilityTrackingSaga (supply-chain)       |
| NewSupplierPerformanceSaga      | NewSupplierPerformanceCollaborationSaga (supply-chain)    |
| NewReverseLogisticsSaga         | NewReverseLogisticsReturnsManagementSaga (supply-chain)   |

### Banking syntax fix

27. **`sagas/banking/wire_transfer_saga.go`** — missing comma after a
    StepDefinition literal inside a slice; the parser mistook the next
    step for a new top-level declaration.

### Timeout circuit-breaker signature

28. **`timeout/circuit_breaker.go`** — `Reset()` returned `error`; the
    interface says no return. Aligned. Test updated to match.

## Running saga tests

```
go build ./packages/saga/...
go vet   ./packages/saga/...
go test  ./packages/saga/...
```

All three should succeed. The `saga_examples` build tag remains for
`land_acquisition_saga_example.go` until its StepDefinition field
references are brought current with the canonical model.

## Non-goals

This README is a precise record of what B.8 shipped. Saga tests that
exercise real Kafka / Postgres stay gated on the external infrastructure
Gates (G1 Postgres, G2 running monolith) — the build-time contract is
now clean regardless.
