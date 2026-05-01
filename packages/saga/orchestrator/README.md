# Saga Orchestrator Package

The orchestrator package provides the core saga execution engine - it coordinates the execution of distributed transaction steps with timeout, retry, and compensation handling.

## Components

### 1. SagaOrchestratorImpl

**File:** `orchestrator.go`

The main orchestration engine that manages saga execution from start to finish.

**Usage:**

```go
import "p9e.in/samavaya/packages/saga"
import "p9e.in/samavaya/packages/saga/orchestrator"

// Create orchestrator (typically done via FX)
orch := orchestrator.NewSagaOrchestratorImpl(
    registry,
    executor,
    timeoutHandler,
    eventPublisher,
    repository,
    execLogRepository,
    config,
)

// Start a saga
execution, err := orch.ExecuteSaga(ctx, "SAGA-S01", &saga.SagaExecutionInput{
    TenantID:  "tenant-123",
    CompanyID: "company-456",
    BranchID:  "branch-789",
})

if err != nil {
    // Handle error
}

// Resume a failed saga
execution, err := orch.ResumeSaga(ctx, "SAGA-ID-789")

// Get current saga state
execution, err := orch.GetExecution(ctx, "SAGA-ID-789")

// Get execution timeline
timeline, err := orch.GetExecutionTimeline(ctx, "SAGA-ID-789")
```

**Key Features:**

- Sequential step execution
- Exponential backoff retry (1s, 2s, 4s, 8s, ...)
- Step-level and saga-level timeouts
- Automatic compensation on failure
- Event publishing for audit trail
- State persistence after each step
- Thread-safe concurrent execution

### 2. SagaRegistry

**File:** `saga_registry.go`

Manages registration and lookup of saga handlers.

**Usage:**

```go
registry := orchestrator.NewSagaRegistry()

// Register a handler
handler := &OrderSagaHandler{ /* ... */ }
err := registry.RegisterHandler("SAGA-S01", handler)

// Get a handler
handler, err := registry.GetHandler("SAGA-S01")

// Check if handler exists
exists := registry.HasHandler("SAGA-S01")

// Get all handlers
allHandlers := registry.GetAllHandlers()

// Count handlers
count := registry.GetHandlerCount()
```

**Features:**

- Duplicate handler detection
- Thread-safe concurrent access
- Dynamic handler registration
- Testing utilities (RemoveHandler, ClearHandlers)

### 3. ExecutionPlanner

**File:** `execution_planner.go`

Plans saga execution sequence and validates dependencies.

**Usage:**

```go
stepDefs := handler.GetStepDefinitions()
planner := orchestrator.NewExecutionPlanner(stepDefs)

// Plan execution sequence
plan, err := planner.PlanExecution()

// Check if step can execute
canExecute, err := planner.CanExecuteStep(2, executionState)

// Get critical path (steps that must succeed)
criticalSteps := planner.GetCriticalPath()

// Get optional steps
optionalSteps := planner.GetOptionalSteps()

// Estimate execution time
estimatedSeconds := planner.EstimateExecutionTime()
```

**Features:**

- Sequential execution planning
- Dependency validation
- Critical path analysis
- Execution time estimation
- Circular dependency detection

## Execution Flow

```
1. Client calls ExecuteSaga("SAGA-S01", input)

2. Orchestrator validates:
   - Saga type is registered
   - Input is valid
   - Step definitions exist

3. Creates SagaExecution record with:
   - Unique ID (ULID)
   - Tenant/Company/Branch IDs
   - Saga status: RUNNING
   - Current step: 1

4. For each step:
   a. Setup timeout (step-level)
   b. Execute step (with retry loop)
      - Attempt step execution
      - On failure, backoff and retry
      - Exponential backoff: 1s, 2s, 4s, 8s, ...
   c. Log step result
   d. Update execution state
   e. Cancel step timeout

5. On critical step failure:
   a. Mark saga as COMPENSATING
   b. Reverse all successful steps
   c. Mark saga as COMPENSATED

6. On success:
   a. Mark saga as COMPLETED
   b. Set CompletedAt timestamp

7. Publish appropriate events:
   - SAGA.STEP.STARTED
   - SAGA.STEP.COMPLETED or SAGA.STEP.FAILED
   - SAGA.SAGA.COMPLETED or SAGA.SAGA.FAILED
   - SAGA.COMPENSATION.STARTED/COMPLETED (on failure)

8. Return SagaExecution with final state
```

## Configuration

The orchestrator uses `saga.DefaultConfig` for default values:

```go
config := &saga.DefaultConfig{
    DefaultTimeoutSeconds:    60,      // Step timeout
    DefaultMaxRetries:        3,       // Retry attempts
    DefaultInitialBackoff:    time.Second,
    DefaultMaxBackoff:        30 * time.Second,
    BackoffMultiplier:        2.0,
    JitterFraction:           0.1,
    CircuitBreakerThreshold:  5,
    CircuitBreakerResetMs:    60000,
}
```

## Error Handling

### Retryable Errors

Errors that will be retried with exponential backoff:
- Temporary network errors
- Timeout errors (from timeout handler)
- Service unavailable errors
- Errors configured in `RetryableErrors` list

### Non-Retryable Errors

Errors that fail immediately:
- Validation errors
- Authorization errors
- Not found errors
- Errors configured in `NonRetryableErrors` list

### Critical vs Non-Critical Steps

```go
// Critical step (must succeed)
step := &saga.StepDefinition{
    StepNumber: 1,
    IsCritical: true,  // Saga fails if this step fails
}

// Non-critical step (saga continues if this step fails)
step := &saga.StepDefinition{
    StepNumber: 2,
    IsCritical: false,  // Saga continues even if this fails
}
```

## Thread Safety

The orchestrator is thread-safe for concurrent saga execution:

```go
// These can run concurrently (different sagas)
go orch.ExecuteSaga(ctx1, "SAGA-S01", input1)  // SAGA-123
go orch.ExecuteSaga(ctx2, "SAGA-S01", input2)  // SAGA-456

// Reading is concurrent-safe
go orch.GetExecution(ctx3, "SAGA-123")
go orch.GetExecution(ctx4, "SAGA-456")

// Only one write per saga
// (ExecuteSaga uses exclusive lock per saga type,
//  not per saga instance, for simplicity)
```

## Testing

The package includes comprehensive unit tests:

```bash
go test ./packages/saga/orchestrator -v
go test ./packages/saga/orchestrator -cover
```

Test coverage includes:
- Happy path saga execution
- Error scenarios
- Timeout handling
- Retry logic
- Handler registration
- Registry operations
- Execution planning
- Critical path analysis

## FX Integration

The orchestrator is integrated into the FX dependency injection framework:

```go
import "p9e.in/samavaya/packages/saga/orchestrator"

var Module = fx.Module(
    "order",
    orchestrator.SagaOrchestratorModule,  // Provides SagaOrchestrator
    // ... other modules
)

// In your handler
type Handler struct {
    orchestrator saga.SagaOrchestrator `name:""`
}
```

## Next Steps

The orchestrator requires these components (being implemented in Phase 0 Days 4-5):

1. **SagaStepExecutor** - Execute individual steps via RPC
2. **SagaTimeoutHandler** - Manage timeouts and retries
3. **SagaEventPublisher** - Publish saga events to Kafka
4. **SagaRepository** - Persist saga state to PostgreSQL
5. **SagaExecutionLogRepository** - Log step executions

## Dependencies

- `p9e.in/samavaya/packages/saga` - Interfaces and models
- `go.uber.org/fx` - Dependency injection (for module)

## Package Structure

```
packages/saga/orchestrator/
├── orchestrator.go           # Main orchestrator implementation
├── saga_registry.go          # Handler registry
├── execution_planner.go      # Execution planning
├── orchestrator_test.go      # Unit tests
├── fx.go                     # FX module
└── README.md                 # This file
```

---

**Status:** ✅ Complete (Phase 0 Days 2-3)
**Next:** Phase 0 Days 4-5 - Executor & RPC
