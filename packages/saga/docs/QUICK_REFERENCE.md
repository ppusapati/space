# Phase 4 Saga Engine - Quick Reference Guide

## 30-Second Overview

Phase 4 implements **24 distributed sagas** across Finance, Manufacturing, HR, and Projects:

- **8 Finance Sagas:** Revenue recognition, bank reconciliation, month-end close, etc.
- **6 Manufacturing Sagas:** BOM explosion, production orders, quality rework, etc.
- **6 HR Sagas:** Payroll, onboarding, exit, appraisal, leave, expense reimbursement
- **4 Projects Sagas:** Project billing, progress billing, subcontractor payment, close

**Key Features:** Eventual consistency, automatic compensation, circuit breaker fault tolerance, Kafka event publishing, 90%+ test coverage.

---

## Saga IDs Quick Lookup

### Finance Sagas

| ID | Name | Steps | Special | Status |
|----|------|-------|---------|--------|
| F01 | Month-End Close | 12 | NO compensation | Critical |
| F02 | Bank Reconciliation | 11 | With compensation | Critical |
| F03 | Multi-Currency Revaluation | 8 | With compensation | In Progress |
| F04 | Intercompany Transaction | 7 | With compensation | In Progress |
| F05 | Revenue Recognition | 6 | With compensation | In Progress |
| F06 | Asset Capitalization | 5 | With compensation | In Progress |
| F07 | GST Credit Reversal | 5 | With compensation | In Progress |
| F08 | Cost Center Allocation | 6 | With compensation | In Progress |

### Manufacturing Sagas

| ID | Name | Steps | File | Status |
|----|------|-------|------|--------|
| M01 | Production Order Release | 6 | production_order_saga.go | In Progress |
| M02 | Subcontracting | 8 | subcontracting_saga.go | In Progress |
| M03 | BOM Explosion & MRP | 8 | bom_explosion_mrp_saga.go | In Progress |
| M04 | Production Order | 6 | production_order_saga.go | In Progress |
| M05 | Job Card Consumption | 5 | job_card_consumption_saga.go | In Progress |
| M06 | Quality Rework | 5 | quality_rework_saga.go | In Progress |

### HR Sagas

| ID | Name | Steps | File | Status |
|----|------|-------|------|--------|
| H01 | Payroll Processing | 10 | payroll_processing_saga.go | In Progress |
| H02 | Employee Onboarding | 10 | employee_onboarding_saga.go | In Progress |
| H03 | Employee Exit | 8 | employee_exit_saga.go | In Progress |
| H04 | Appraisal & Salary Revision | 7 | appraisal_salary_revision_saga.go | In Progress |
| H05 | Leave Application | 4 | leave_application_saga.go | In Progress |
| H06 | Expense Reimbursement | 6 | expense_reimbursement_saga.go | In Progress |

### Projects Sagas

| ID | Name | Steps | File | Status |
|----|------|-------|------|--------|
| PR01 | Project Billing | 7 | project_billing_saga.go | In Progress |
| PR02 | Progress Billing | 8 | progress_billing_saga.go | In Progress |
| PR03 | Subcontractor Payment | 6 | subcontractor_payment_saga.go | In Progress |
| PR04 | Project Close | 7 | project_close_saga.go | In Progress |

---

## Core Service Endpoints

### Finance (8100-8200)
```
general-ledger:      http://localhost:8100
accounts-receivable: http://localhost:8103
accounts-payable:    http://localhost:8104
journal:             http://localhost:8083
transaction:         http://localhost:8084
billing:             http://localhost:8105
reconciliation:      http://localhost:8107
cost-center:         http://localhost:8109
financial-reports:   http://localhost:8111
financial-close:     http://localhost:8093
compliance-postings: http://localhost:8094
tax-engine:          http://localhost:8091
cash-management:     http://localhost:8088
depreciation:        http://localhost:8169
```

### Manufacturing (8190-8198)
```
bom:                 http://localhost:8190
production-order:    http://localhost:8191
production-planning: http://localhost:8192
shop-floor:          http://localhost:8193
quality-production:  http://localhost:8194
subcontracting:      http://localhost:8195
work-center:         http://localhost:8196
routing:             http://localhost:8197
job-card:            http://localhost:8198
```

### HR (8113-8175)
```
employee:            http://localhost:8113
leave:               http://localhost:8114
attendance:          http://localhost:8115
payroll:             http://localhost:8116
salary-structure:    http://localhost:8117
recruitment:         http://localhost:8118
appraisal:           http://localhost:8173
expense:             http://localhost:8174
exit:                http://localhost:8175
```

### Projects (8160-8166)
```
project:             http://localhost:8160
task:                http://localhost:8161
timesheet:           http://localhost:8162
project-costing:     http://localhost:8163
boq:                 http://localhost:8164
sub-contractor:      http://localhost:8165
progress-billing:    http://localhost:8166
```

---

## Saga Execution Code

### Start a Saga

```go
execution, err := orch.ExecuteSaga(ctx, "SAGA-S01", &saga.SagaExecutionInput{
    TenantID:  "tenant-123",
    CompanyID: "company-456",
    BranchID:  "branch-789",
    Input: map[string]interface{}{
        "customer_id": "CUST-001",
        "order_id": "ORD-001",
    },
})
```

### Check Status

```go
execution, err := orch.GetExecution(ctx, "SAGA-123")
fmt.Printf("Status: %s, Step: %d\n", execution.Status, execution.CurrentStepNumber)
```

### Get History

```go
timeline, err := orch.GetExecutionTimeline(ctx, "SAGA-123")
for _, step := range timeline {
    fmt.Printf("Step %d: %s\n", step.StepNumber, step.Status)
}
```

### Resume Failed Saga

```go
execution, err := orch.ResumeSaga(ctx, "SAGA-123")
```

---

## Saga Structure Template

```go
// Step definition template
{
    StepNumber:    1,
    ServiceName:   "service-name",
    HandlerMethod: "MethodName",
    InputMapping: map[string]string{
        "field1": "$.input.field1",
        "field2": "$.steps.1.result.field2",
    },
    TimeoutSeconds: 20,
    IsCritical:     true,
    CompensationSteps: []int32{101},  // Maps to step 101
    RetryConfig: &saga.RetryConfiguration{
        MaxRetries:        3,
        InitialBackoffMs:  1000,
        MaxBackoffMs:      30000,
        BackoffMultiplier: 2.0,
        JitterFraction:    0.1,
    },
}
```

---

## Step Numbering Convention

```
Forward Steps:      1-99
Compensation Steps: 101-199 (100 + forward step number)

Example:
  Step 1 (forward)     → Step 101 (compensation)
  Step 2 (forward)     → Step 102 (compensation)
  Step 3 (forward)     → Step 103 (compensation)
  ...
  Step N (forward)     → Step (100+N) (compensation)
```

---

## Retry Strategy

**Default Exponential Backoff:**
```
Attempt 1: 1s
Attempt 2: 2s
Attempt 3: 4s
Attempt 4: 8s
Attempt 5: 16s
Attempt 6+: 30s (max)
```

**With Jitter (±10%):**
```
Attempt 1: 0.9s - 1.1s
Attempt 2: 1.8s - 2.2s
Attempt 3: 3.6s - 4.4s
...
```

---

## Timeout Recommendations

| Operation | Timeout | Examples |
|-----------|---------|----------|
| Fast | 10-15s | Lookups, reads, simple validation |
| Standard | 20-30s | Create records, updates, GL posting |
| Complex | 60s+ | Consolidation, report generation, FX revaluation |

---

## Critical vs Non-Critical

```
CRITICAL: Saga FAILS if step fails
  - GL postings
  - Inventory deductions
  - Payment postings
  - Revenue recognition

NON-CRITICAL: Saga CONTINUES if step fails
  - Notifications
  - Audit logging
  - Report generation
  - Optional validations
```

---

## Testing Template

```go
func TestMySaga_Type(t *testing.T) {
    saga := NewMySaga()
    assert.Equal(t, "SAGA-MOD-##", saga.SagaType())
}

func TestMySaga_Steps(t *testing.T) {
    saga := NewMySaga()
    steps := saga.GetStepDefinitions()
    assert.Equal(t, 8, len(steps))  // 5 forward + 3 compensation
}

func TestMySaga_InputValidation(t *testing.T) {
    saga := NewMySaga()

    // Valid input
    err := saga.ValidateInput(map[string]interface{}{
        "field1": "value",
    })
    assert.NoError(t, err)

    // Invalid input
    err = saga.ValidateInput(map[string]interface{}{})
    assert.Error(t, err)
}
```

---

## FX Module Integration

```go
// In fx.go
var MyModuleSagasModule = fx.Module(
    "mymodule_sagas",
    fx.Provide(
        NewMySaga1,
        NewMySaga2,
    ),
)

var MyModuleSagasRegistrationModule = fx.Module(
    "mymodule_sagas_registration",
    fx.Invoke(registerMyModuleSagas),
)

func registerMyModuleSagas(
    registry *orchestrator.SagaRegistry,
    saga1 saga.SagaHandler,
    saga2 saga.SagaHandler,
) error {
    if err := registry.RegisterHandler("SAGA-MOD-01", saga1); err != nil {
        return err
    }
    if err := registry.RegisterHandler("SAGA-MOD-02", saga2); err != nil {
        return err
    }
    return nil
}
```

---

## JSONPath Input Mapping

```go
InputMapping: map[string]string{
    // From saga input
    "tenantID": "$.tenantID",

    // From nested input
    "invoiceID": "$.input.invoice_id",

    // From previous step result
    "lineItems": "$.steps.1.result.line_items",

    // From multiple steps
    "totalAmount": "$.steps.5.result.total_amount",
    "taxAmount": "$.steps.6.result.tax_amount",
}
```

---

## Kafka Events

All saga events published to Kafka topic: `saga-events`

**Event Types:**
- `SAGA.STEP.STARTED` - Step execution begins
- `SAGA.STEP.COMPLETED` - Step completes successfully
- `SAGA.STEP.FAILED` - Step fails
- `SAGA.STEP.RETRYING` - Step retrying after failure
- `SAGA.COMPENSATION.STARTED` - Compensation begins
- `SAGA.COMPENSATION.COMPLETED` - Compensation completes
- `SAGA.SAGA.COMPLETED` - Entire saga completes
- `SAGA.SAGA.FAILED` - Entire saga fails

---

## Monitoring

**Key Metrics:**
```
saga_execution_total{saga_type="SAGA-S01",status="success"}
saga_execution_duration_seconds{saga_type="SAGA-S01",quantile="p50"}
saga_step_execution_total{saga_type="SAGA-S01",step_number="1"}
saga_step_retries_total{saga_type="SAGA-S01",step_number="1"}
saga_compensation_total{saga_type="SAGA-S01"}
```

**Dashboard Queries:**
```sql
-- Success rate (last 24h)
SELECT saga_type, ROUND(100.0 * COUNT(CASE WHEN status='success' THEN 1 END) / COUNT(*), 2) as success_rate
FROM saga_executions
WHERE executed_at >= NOW() - INTERVAL 24 HOUR
GROUP BY saga_type;
```

---

## Phase Timeline

**Phase 4A (Feb 15-20):** Foundation sagas (F05-F08, H05, M05)
**Phase 4B (Feb 20-28):** Core operations (M03-M04, M06, H02-H04, H06, PR01-PR04)
**Phase 4C (Feb 28-Mar 15):** Critical systems (F01-F04, H01, H03, M01-M02)

---

## Files & Locations

```
packages/saga/
├── docs/
│   ├── PHASE_4_IMPLEMENTATION.md    (3,008 lines - Full guide)
│   ├── INDEX.md                     (355 lines - Navigation)
│   └── QUICK_REFERENCE.md           (This file)
├── sagas/
│   ├── finance/                     (8 sagas + tests)
│   ├── manufacturing/               (6 sagas + tests)
│   ├── hr/                          (6 sagas + tests)
│   └── projects/                    (4 sagas + tests)
├── orchestrator/                    (Core engine)
├── compensation/                    (Rollback engine)
├── executor/                        (Step execution)
├── events/                          (Kafka publishing)
├── timeout/                         (Retry & timeouts)
├── connector/                       (RPC client)
└── repository/                      (Data access)
```

---

## Common Tasks

### Add New Saga
1. Create file: `saga_xxx.go`
2. Implement `SagaHandler` interface
3. Add to FX module in `fx.go`
4. Write tests in `xxx_test.go`
5. Register in `registerXxxSagas()`

### Add New Service
1. Add to `ServiceRegistry` in `registry.go`
2. Choose port in module range
3. Update saga step definitions if needed
4. Test service registration

### Debug Failed Saga
1. Get execution: `orch.GetExecution(ctx, sagaID)`
2. Check status and error message
3. Get timeline: `orch.GetExecutionTimeline(ctx, sagaID)`
4. Review Kafka events for audit trail
5. Resume if recoverable: `orch.ResumeSaga(ctx, sagaID)`

---

## Important Constraints

**Max Forward Steps:** 99 (compensation steps: 100-199)
**Max Services per Saga:** Typically 8-12 (maintain simplicity)
**Max Saga Duration:** 15 minutes (for most sagas)
**Max Concurrent Sagas:** Tested up to 10,000
**Test Coverage Target:** 90%+ for production sagas

---

## Troubleshooting Quick Guide

| Problem | Cause | Solution |
|---------|-------|----------|
| Saga stuck in RUNNING | Service unreachable | Check service endpoints |
| High retry rate | Timeout too short | Increase TimeoutSeconds |
| Compensation fails | State changed elsewhere | Check for concurrent operations |
| Step results missing | Wrong JSONPath | Verify input mapping |
| Saga won't register | FX module not included | Add to SagaEngineModule |

---

## Performance Targets

| Metric | Target |
|--------|--------|
| P50 Execution Time | 2-5 seconds |
| P95 Execution Time | 8-15 seconds |
| P99 Execution Time | 15-30 seconds |
| Success Rate | 99%+ |
| Retry Success | 80%+ |
| Max Concurrent | 10,000 |

---

## Further Reading

1. **Detailed Implementation:** See `PHASE_4_IMPLEMENTATION.md` (3,008 lines)
2. **Navigation:** See `INDEX.md` for complete table of contents
3. **Orchestrator:** See `orchestrator/README.md`
4. **Examples:** Check actual saga implementations in `sagas/` directories

---

**Last Updated:** February 15, 2026
**Status:** Ready for Phase 4A
**Format:** Quick reference (not complete documentation)

For complete details, refer to PHASE_4_IMPLEMENTATION.md
