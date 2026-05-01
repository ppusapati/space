# Sales Module Sagas

This package provides saga handler implementations for the sales module workflows. Sagas orchestrate distributed transactions across multiple microservices, with automatic compensation on failure.

## Implemented Sagas

### SAGA-S01: Order-to-Cash
**Business Flow:** Customer places order → Stock reserved → Order confirmed → Invoice generated → E-Invoice created → AR entry posted → GL updated → Customer notified

**Forward Steps (8):**
1. sales-order.CreateOrder - Create order (DRAFT)
2. inventory-core.ReserveStock - Reserve inventory
3. sales-order.ConfirmOrder - Confirm order
4. sales-invoice.CreateInvoice - Generate invoice
5. e-invoice.GenerateIRN - GSTN IRN generation (5 retries, external API)
6. accounts-receivable.CreateAREntry - Create AR entry
7. general-ledger.PostJournal - Post GL entries
8. notification.SendInvoice - Send to customer

**Compensation Steps (7):**
- 101: CancelOrder (Step 1)
- 102: ReleaseReservation (Step 2)
- 103: RevertConfirmation (Step 3)
- 104: CancelInvoice (Step 4)
- 105: MarkPendingIRN (Step 5 - partial)
- 106: ReverseAREntry (Step 6)
- 107: ReverseJournal (Step 7)

**Timeout:** 60 seconds
**Idempotency:** Transaction ID-based

### SAGA-S02: Quotation-to-Order Conversion
**Business Flow:** Accept quotation → Lock pricing → Create order → Update opportunity → Notify sales team

**Forward Steps (5):**
1. crm.AcceptQuotation
2. pricing.LockPricing
3. sales-order.CreateFromQuotation
4. crm.UpdateOpportunity
5. notification.NotifySalesTeam (non-critical)

**Compensation Steps (4):**
- 101: RevertQuotation
- 102: ReleasePricingLock
- 103: CancelOrder
- 104: RevertOpportunityStage

**Timeout:** 30 seconds

### SAGA-S03: Order-to-Fulfillment
**Business Flow:** Mark for fulfillment → Allocate stock → Pick → Confirm picking → Package → Generate EWB → Create shipment → Post goods issue → Send tracking

**Forward Steps (9):**
1. sales-order.MarkForFulfillment
2. inventory-core.AllocateStock
3. wms.CreatePickList
4. wms.ConfirmPicking
5. fulfillment.CreatePackage
6. e-way-bill.GenerateEWB (5 retries, external API, non-critical)
7. shipping.CreateShipment (5 retries, external API)
8. inventory-core.PostGoodsIssue
9. notification.SendTrackingInfo (non-critical)

**Compensation Steps (6):**
- 102: ReleaseAllocation
- 103: CancelPickList
- 105: CancelPackage
- 106: MarkPendingCancellation (EWB partial)
- 107: CancelShipment
- 108: ReverseGoodsIssue

**Timeout:** 120 seconds

### SAGA-S04: Sales Return
**Business Flow:** Create return → Inspect goods → Receive return → Create credit note → Adjust AR → Post GL → Process refund → Complete return

**Forward Steps (8):**
1. returns.CreateReturn
2. qc.InspectReturnedGoods
3. inventory-core.ReceiveReturn
4. sales-invoice.CreateCreditNote
5. accounts-receivable.AdjustAR
6. general-ledger.PostCreditNote
7. banking.ProcessRefund (external API)
8. returns.CompleteReturn (non-critical)

**Timeout:** 90 seconds

### SAGA-S05: Commission Calculation
**Business Flow:** Mark invoice paid → Calculate commission → Accrue commission → Post GL → Notify sales person

**Forward Steps (5):**
1. accounts-receivable.MarkInvoicePaid
2. commission.CalculateCommission
3. payroll.AccrueCommission
4. general-ledger.PostCommission
5. commission.NotifySalesPerson (non-critical)

**Compensation Steps (3):**
- 102: DeleteCommission
- 103: ReverseAccrual
- 104: ReverseCommissionPosting

**Timeout:** 30 seconds

### SAGA-S06: E-Invoice Generation
**Business Flow:** Validate → Generate JSON → Call GSTN API → Update invoice → Record in GST ledger → Audit log → Send notification

**Forward Steps (7):**
1. sales-invoice.ValidateForEInvoice
2. e-invoice.GenerateJSON
3. e-invoice.CallGSTNAPI (10 retries, external API)
4. sales-invoice.UpdateWithIRN
5. gst.RecordEInvoice
6. audit.LogEInvoice (non-critical)
7. notification.SendEInvoice (non-critical)

**Critical:** IRN generation is mandatory for B2B invoices > ₹5L

**Timeout:** 180 seconds (GSTN can be slow)

### SAGA-S07: Dealer Performance & Incentive
**Business Flow:** Calculate dealer sales → Calculate incentive → Request approval → Approve → Create AP entry → Post GL → Notify dealer

**Forward Steps (7):**
1. sales-analytics.CalculateDealerSales
2. dealer.CalculateIncentive
3. approval.RequestIncentiveApproval
4. approval.ApproveIncentive
5. accounts-payable.CreateIncentivePayable
6. general-ledger.PostIncentive
7. dealer.NotifyDealer (non-critical)

**Timeout:** 120 seconds

## Usage

### Registering Saga Handlers

The saga handlers are automatically registered via FX modules:

```go
// In your main application
fx.New(
    // ... other modules
    saga.SagaEngineModule,           // Core saga engine
    connector.ConnectorModule,        // RPC connector
    sales.SalesSagasModule,           // Sales saga handlers
    sales.SalesSagasRegistrationModule, // Register with orchestrator
)
```

### Executing a Saga

```go
execution, err := orchestrator.ExecuteSaga(ctx, "SAGA-S01", &saga.SagaExecutionInput{
    TenantID:  "tenant-123",
    CompanyID: "company-456",
    BranchID:  "branch-789",
    Input: map[string]interface{}{
        "customer_id": "cust-001",
        "items": []interface{}{
            map[string]interface{}{
                "product_id": "prod-001",
                "quantity": 10,
                "price": 100.00,
            },
        },
        "total_amount": 1000.00,
        "due_date": "2026-03-14",
    },
    Metadata: map[string]string{
        "order_ref": "ORD-2026-001",
    },
})

if err != nil {
    log.Printf("Saga execution failed: %v", err)
}

log.Printf("Saga %s: status=%s, current_step=%d/%d",
    execution.ID, execution.Status, execution.CurrentStep, execution.TotalSteps)
```

### Input Requirements

Each saga requires specific input fields:

| Saga | Required Fields | Optional Fields |
|------|-----------------|-----------------|
| SAGA-S01 | customer_id, items[], total_amount | due_date |
| SAGA-S02 | quotation_id, opportunity_id | - |
| SAGA-S03 | order_id | - |
| SAGA-S04 | invoice_id, return_amount | - |
| SAGA-S05 | invoice_id, payment_amount | - |
| SAGA-S06 | invoice_id | - |
| SAGA-S07 | dealer_id, month | - |

## Compensation Logic

Compensation executes automatically when a critical step fails:

1. **Reverse Execution:** Steps are compensated in reverse order (last successful first)
2. **Idempotency:** Compensation steps check if already executed to prevent double-execution
3. **Non-Critical Failures:** Non-critical step failures are logged but don't trigger compensation
4. **Partial Compensation:** Some operations (GSTN IRN, EWB) can only be partially reversed
5. **Manual Intervention:** Failed compensations require manual resolution

### Compensation Example (SAGA-S01)

```
Forward execution:
Step 1: CreateOrder ✓
Step 2: ReserveStock ✓
Step 3: ConfirmOrder ✓
Step 4: CreateInvoice ✗ (FAILS)

Compensation triggered:
Step 103: RevertConfirmation ✓
Step 102: ReleaseReservation ✓
Step 101: CancelOrder ✓

Result: Order cancelled, stock released, order reverted to DRAFT
```

## Retry Strategy

Different steps have different retry configurations:

| Step Type | Max Retries | Initial Backoff | Max Backoff | Use Case |
|-----------|-------------|-----------------|-------------|----------|
| Internal (sales-order, inventory) | 3 | 1s | 30s | Service dependency |
| External API (GSTN, Carrier, Banking) | 5-10 | 2s | 120s | Unreliable network |
| Non-critical | 3 | 1s | 30s | Logging, notification |

Exponential backoff formula: `backoff = initial * (2^retry_count) + jitter`

## Service Registry

The service registry maps service names to endpoints. Default configuration:

```go
"sales-order":        "http://localhost:8119",
"sales-invoice":      "http://localhost:8120",
"crm":                "http://localhost:8121",
"territory":          "http://localhost:8122",
"commission":         "http://localhost:8123",
"pricing":            "http://localhost:8124",
"dealer":             "http://localhost:8125",
"sales-analytics":    "http://localhost:8126",
"route-planning":     "http://localhost:8127",
"field-sales":        "http://localhost:8128",
"inventory-core":     "http://localhost:8179",
"wms":                "http://localhost:8180",
"qc":                 "http://localhost:8182",
"fulfillment":        "http://localhost:8187",
"shipping":           "http://localhost:8188",
"general-ledger":     "http://localhost:8100",
"accounts-receivable":"http://localhost:8103",
"accounts-payable":   "http://localhost:8104",
"payroll":            "http://localhost:8116",
"gst":                "http://localhost:8155",
"e-invoice":          "http://localhost:8156",
"e-way-bill":         "http://localhost:8157",
"banking":            "http://localhost:8159",
```

## Monitoring & Observability

### Events Published

Each saga publishes events throughout execution:

- `StepStarted` - Step execution begins
- `StepCompleted` - Step completed successfully
- `StepFailed` - Step failed
- `StepRetrying` - Step retry in progress
- `SagaCompleted` - Entire saga completed
- `SagaFailed` - Saga failed (compensation started)
- `CompensationStarted` - Compensation process begins
- `CompensationCompleted` - Compensation completed

### Status Tracking

Monitor saga execution with `orchestrator.GetExecution()`:

```go
execution, _ := orchestrator.GetExecution(ctx, sagaID)
log.Printf("Status: %s", execution.Status)       // RUNNING, COMPLETED, FAILED, COMPENSATING, COMPENSATED
log.Printf("Progress: %d/%d", execution.CurrentStep, execution.TotalSteps)
log.Printf("Error: %s", execution.ErrorMessage)
```

### Timeline

Retrieve detailed execution timeline with `orchestrator.GetExecutionTimeline()`:

```go
timeline, _ := orchestrator.GetExecutionTimeline(ctx, sagaID)
for _, step := range timeline {
    log.Printf("Step %d: %s at %v (took %dms, retries: %d)",
        step.StepNumber, step.Status, step.CompletedAt,
        step.ExecutionTimeMs, step.RetryCount)
}
```

## Testing

Unit tests are provided in `sales_sagas_test.go`:

```bash
go test ./packages/saga/sagas/sales/... -v
```

Tests cover:
- Saga type identification
- Step definition counts and properties
- Input validation
- Compensation step references
- Retry configuration
- Input mapping definitions
- Unique saga types
- Step sequencing

## Compliance & Regulations

### GST/E-Invoice (SAGA-S01, SAGA-S06)
- Mandatory for B2B invoices > ₹5L
- IRN must be generated within 24 hours
- Cannot modify invoice after IRN generation
- Cancellation allowed within 24 hours only

### E-Way Bill (SAGA-S03)
- Mandatory for goods > ₹50K
- Validity based on distance
- Stock movement must match EWB

### Returns & Credit Notes (SAGA-S04)
- Credit note must reference original invoice
- GST reversal must be reported in GSTR-1
- Return within warranty period

### Commission (SAGA-S05, SAGA-S07)
- Calculated on invoice payment
- Accrued in payroll system
- Posted to GL as expense

## Troubleshooting

### Saga not executing
- Verify handler is registered with correct saga type
- Check input validation passes
- Ensure repository is configured

### Steps failing
- Verify service endpoint is registered in ServiceRegistry
- Check RPC connector can reach service
- Validate step definition matches service interface

### Compensation not happening
- Confirm CompensationSteps are defined
- Check critical steps fail correctly
- Verify compensation steps reference valid step numbers

### Manual Intervention Required
- Check saga_compensation_log for failure details
- GSTN/Carrier/Banking failures often require manual follow-up
- Create tasks in task management system for ops team

## Future Enhancements

- Real-time saga monitoring dashboard
- Saga state replay for debugging
- Advanced compensation strategies
- Integration with event streaming (Kafka)
- Performance metrics collection
- Sagastate visualization
