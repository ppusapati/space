# Purchase Module Sagas

This package provides saga handler implementations for the purchase module workflows. Sagas orchestrate distributed transactions across multiple microservices, with automatic compensation on failure.

## Implemented Sagas

### SAGA-P01: Procure-to-Pay
**Business Flow:** Create PO → Approve PO → Send to Vendor → Receive Goods (GRN) → Update Stock → Create Invoice → 3-Way Match → Approve Invoice → Post AP → Post GL → Process Payment → Close PO

**Forward Steps (12):**
1. purchase-order.CreatePurchaseOrder - Create PO (DRAFT)
2. purchase-order.ApprovePurchaseOrder - Approve PO
3. purchase-order.SendPOToVendor - Email/EDI to vendor
4. purchase-order.CreateGoodsReceipt - GRN creation
5. inventory-core.ReceiveStock - Update stock levels
6. purchase-invoice.CreateInvoice - Vendor invoice entry
7. purchase-invoice.PerformThreeWayMatch - PO ↔ GRN ↔ Invoice matching
8. purchase-invoice.ApproveInvoice - Invoice approval
9. accounts-payable.CreateAPEntry - AP posting
10. general-ledger.PostPurchaseJournal - GL entries (Dr: Expense/Asset, Cr: AP)
11. banking.ProcessVendorPayment - Payment execution
12. purchase-order.ClosePurchaseOrder - Mark PO as CLOSED

**Compensation Steps (11):** Automatic reversal of all forward steps in reverse order
- 101-112: Comprehensive compensation including payment voiding (PARTIAL)

**Timeout:** 180 seconds (longest saga - external payment processing)
**Idempotency:** Invoice ID-based

### SAGA-P02: Purchase Return
**Business Flow:** Create Return → QC Inspection → Reverse Stock → Create Debit Note → Adjust AP → Post GL → Complete Return

**Forward Steps (7):**
1. purchase-invoice.CreatePurchaseReturn
2. qc.InspectReturnedGoods
3. inventory-core.ReverseReceipt
4. purchase-invoice.CreateDebitNote
5. accounts-payable.AdjustAPForReturn
6. general-ledger.PostReturnJournal
7. purchase-invoice.CompleteReturn

**Compensation Steps (6):**
- 101-107: Comprehensive reversal including QC inspection

**Timeout:** 90 seconds

### SAGA-P03: Vendor Payment with TDS
**Business Flow:** Validate Payment → Calculate TDS → Deduct TDS → Record Withholding → Pay Vendor → Update AP → Post GL → Record Challan → Update TDS Return

**Forward Steps (9):**
1. accounts-payable.ValidatePayment
2. tds.CalculateTDS - Based on section (194C, 194J, etc.), rate, threshold
3. tds.DeductTDS - Gross - TDS = Net
4. tds.RecordWithholding
5. banking.ProcessNetPayment - Pay net amount (external API)
6. accounts-payable.UpdateAPForPayment
7. general-ledger.PostTDSJournal - Dr: AP, Cr: Bank + TDS Payable
8. tds.RecordChallan - Link to quarterly filing
9. tds.UpdateTDSReturn

**Compensation Steps (8):**
- 102-109: TDS reversal, payment void (PARTIAL), GL reversal, etc.

**Special Considerations:**
- TDS rates vary by section (194C, 194J, etc.)
- Threshold checks (₹30K for 194C)
- Form 16A certificate generation (non-critical)

**Timeout:** 120 seconds
**Retry Strategy:** Step 5 (Banking): 5 retries, Step 8 (Challan): 3 retries, Others: 3 retries

### SAGA-P04: Budget Check & Consumption
**Business Flow:** Check Budget → Lock Budget → Consume Budget → Approve Requisition → Record Consumption → Update Utilization → Release Lock

**Forward Steps (7):**
1. budget.CheckAvailableBudget
2. budget.LockBudgetAmount - Soft reservation
3. budget.ConsumeBudget
4. procurement.ApproveRequisition
5. budget.RecordConsumption
6. budget.UpdateUtilization
7. budget.ReleaseLock (non-critical)

**Compensation Steps (6):**
- 102-106: Complete rollback

**Timeout:** 60 seconds

## Usage

### Registering Saga Handlers

The saga handlers are automatically registered via FX modules:

```go
// In your main application
fx.New(
    // ... other modules
    saga.SagaEngineModule,           // Core saga engine
    connector.ConnectorModule,        // RPC connector
    purchase.PurchaseSagasModule,    // Purchase saga handlers
    purchase.PurchaseSagasRegistrationModule, // Register with orchestrator
)
```

### Executing a Saga

```go
execution, err := orchestrator.ExecuteSaga(ctx, "SAGA-P01", &saga.SagaExecutionInput{
    TenantID:  "tenant-123",
    CompanyID: "company-456",
    BranchID:  "branch-789",
    Input: map[string]interface{}{
        "vendor_id":      "vendor-001",
        "items": []interface{}{
            map[string]interface{}{
                "product_id": "prod-001",
                "quantity": 10,
                "unit_price": 100.00,
            },
        },
        "delivery_date": "2026-04-14",
        "payment_terms": "Net 30",
        "invoice_amount": 1000.00,
    },
})

if err != nil {
    log.Printf("Saga execution failed: %v", err)
}

log.Printf("Saga %s: status=%s, current_step=%d/%d",
    execution.ID, execution.Status, execution.CurrentStep, execution.TotalSteps)
```

### Input Requirements

| Saga | Required Fields | Optional Fields |
|------|-----------------|-----------------|
| SAGA-P01 | vendor_id, items[], delivery_date, invoice_amount | payment_terms, approval_note |
| SAGA-P02 | invoice_id, vendor_id, return_reason, return_amount, quantity | - |
| SAGA-P03 | vendor_id, invoice_id, payment_amount, tds_section, tds_rate | journal_date |
| SAGA-P04 | budget_code_id, requested_amount, requisition_id, approver_id, fiscal_year | approval_note |

## Compensation Logic

### Pattern 1: Purchase Order Lifecycle (SAGA-P01)
When a step fails:
- Previous purchase order approvals are reverted
- Goods receipts are cancelled
- Received stock is reversed to source warehouse
- Invoices are cancelled with matching data cleared
- GL entries are reversed with opposing Dr/Cr
- Payments are voided (may require manual intervention if banking fails)

### Pattern 2: TDS Withholding Reversal (SAGA-P03)
When TDS payment fails:
- TDS calculation is reversed
- Deducted amount is restored
- Withholding record is cancelled
- Challan mapping is removed
- GL entries are reversed
- TDS return is reverted

### Pattern 3: Budget Consumption Rollback (SAGA-P04)
When requisition approval fails:
- Budget lock is released
- Consumed budget is restored
- Consumption audit record is deleted
- Utilization percentage is recalculated

## Compensation Examples

### SAGA-P01 Failure Scenario
```
Forward execution:
Step 1: CreatePurchaseOrder ✓
Step 2: ApprovePurchaseOrder ✓
Step 3: SendPOToVendor ✓ (non-critical)
Step 4: CreateGoodsReceipt ✓
Step 5: ReceiveStock ✓
Step 6: CreateInvoice ✓
Step 7: PerformThreeWayMatch ✓
Step 8: ApproveInvoice ✗ (FAILS)

Compensation triggered:
Step 107: ClearMatchingData ✓
Step 106: CancelInvoice ✓
Step 105: ReverseStockReceipt ✓
Step 104: CancelGoodsReceipt ✓
Step 102: RevertApproval ✓
Step 101: CancelPurchaseOrder ✓

Result: Purchase order cancelled, stock released, matching data cleared
```

## Retry Strategy

| Step Type | Max Retries | Initial Backoff | Max Backoff | Use Case |
|-----------|-------------|-----------------|-------------|----------|
| Internal (purchase-order, inventory) | 3 | 1s | 30s | Service dependency |
| External API (Banking) | 5 | 2s | 120s | Unreliable network |
| Non-critical (SendPO, ReleaseLock) | 3 | 1s | 30s | Optional steps |

Exponential backoff formula: `backoff = initial * (2^retry_count) + jitter`

## Service Registry

Service endpoints mapped in `packages/saga/sagas/registry.go`:

```
purchase-order:   http://localhost:8142
purchase-invoice: http://localhost:8143
procurement:      http://localhost:8141
accounts-payable: http://localhost:8104
general-ledger:   http://localhost:8100
banking:          http://localhost:8159
tds:              http://localhost:8158
budget:           http://localhost:8193
```

## Monitoring & Observability

### Events Published

- `StepStarted` - Step execution begins
- `StepCompleted` - Step completed successfully
- `StepFailed` - Step failed
- `StepRetrying` - Step retry in progress
- `SagaCompleted` - Entire saga completed
- `SagaFailed` - Saga failed (compensation started)
- `CompensationStarted` - Compensation process begins
- `CompensationCompleted` - Compensation completed

### Status Tracking

```go
execution, _ := orchestrator.GetExecution(ctx, sagaID)
log.Printf("Status: %s", execution.Status)       // RUNNING, COMPLETED, FAILED, COMPENSATING, COMPENSATED
log.Printf("Progress: %d/%d", execution.CurrentStep, execution.TotalSteps)
log.Printf("Error: %s", execution.ErrorMessage)
```

## Testing

Unit tests are provided in `purchase_sagas_test.go`:

```bash
go test ./packages/saga/sagas/purchase/... -v
```

Tests cover:
- Saga type identification
- Step definition counts
- Input validation
- Compensation step references
- Retry configuration
- Input mapping definitions
- Unique saga types

## Compliance & Regulations

### Purchase Order (SAGA-P01)
- PO must be approved before GRN
- GRN must match PO before invoicing
- 3-way match required before payment

### Purchase Return (SAGA-P02)
- Return reason must be documented
- QC inspection required for defective goods
- Debit note must reference original invoice

### Vendor Payment with TDS (SAGA-P03)
- TDS rates by section (194C, 194J, etc.)
- Threshold validation (e.g., ₹30K for 194C)
- Challan filing required monthly/quarterly
- Form 16A generation for vendors

### Budget Management (SAGA-P04)
- Soft lock prevents over-allocation
- Consumption recorded for audit trail
- Utilization percentage tracking

## Troubleshooting

### Saga not executing
- Verify handler is registered with correct saga type (SAGA-P01, P02, etc.)
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
- Banking payment failures (VoidPayment partial)
- TDS challan filing issues
- Budget override scenarios

## Future Enhancements

- Real-time saga monitoring dashboard
- Saga state replay for debugging
- Advanced compensation strategies
- Integration with event streaming (Kafka)
- Performance metrics collection
- Saga state visualization
