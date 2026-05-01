# Inventory Module Sagas

This package provides saga handler implementations for the inventory module workflows. Sagas orchestrate distributed transactions across multiple microservices, with automatic compensation on failure.

## Implemented Sagas

### SAGA-I01: Inter-Warehouse Transfer
**Business Flow:** Create Transfer → Issue from Source → Create In-Transit → Post Source GL → Ship → Receive at Dest → Post Dest GL → Update Status → Complete Transfer

**Forward Steps (9):**
1. stock-transfer.CreateTransferOrder
2. inventory-core.IssueStockFromSource - Reduce source warehouse
3. stock-transfer.CreateInTransitRecord - Track in-transit inventory separately
4. general-ledger.PostSourceWarehouseGL - Dr: In-Transit, Cr: Inventory
5. shipping.CreateInternalShipment - Logistics tracking
6. inventory-core.ReceiveStockAtDestination - Increase dest warehouse
7. general-ledger.PostDestinationWarehouseGL - Dr: Inventory, Cr: In-Transit
8. stock-transfer.UpdateTransferStatus - Mark as COMPLETED
9. stock-transfer.CompleteTransfer - Finalization

**Compensation Steps (8):**
- 101-108: Complete reversal including GL reversals at both locations

**Special Considerations:**
- In-transit inventory tracked separately (not counted in either warehouse)
- GL entries must balance: Source Dr:In-Transit Cr:Inv = Dest Dr:Inv Cr:In-Transit
- Partial shipments not in Phase 3 scope
- Cross-company transfers require additional GL entries (future)

**Timeout:** 120 seconds

### SAGA-I02: Cycle Count & Stock Adjustment
**Business Flow:** Create Count → Freeze Stock → Execute Count → Calculate Variance → Approve Adjustment → Update Stock → Post GL → Audit Trail → Complete

**Forward Steps (9):**
1. cycle-count.CreateCycleCount
2. inventory-core.FreezeStockMovement - Lock stock for accuracy
3. cycle-count.ExecuteCount - Record physical count
4. cycle-count.CalculateVariance - System vs. Physical delta
5. cycle-count.ApproveAdjustment - Manager approval for variance
6. inventory-core.AdjustStock - Update stock levels (+ or -)
7. general-ledger.PostAdjustmentJournal - GL impact (Dr/Cr: Inventory Variance)
8. cycle-count.RecordAuditTrail - Compliance logging
9. cycle-count.CompleteCount - Finalize, unfreeze stock

**Compensation Steps (8):**
- 101-108: Complete reversal including stock unfreezing

**Timeout:** 90 seconds

### SAGA-I03: Quality Rejection
**Business Flow:** Create Inspection → Fail Inspection → Create Rejection → Adjust Stock → Update Lot Status → Post GL

**Forward Steps (6):**
1. qc.CreateInspection
2. qc.FailInspection
3. qc.CreateRejectionRecord
4. inventory-core.AdjustRejectedStock - Move to rejected stock location
5. lot-serial.UpdateLotStatus - Mark REJECTED
6. general-ledger.PostRejectionJournal - Dr: Inventory Loss, Cr: Inventory

**Compensation Steps (5):**
- 101-105: Complete reversal including lot status reset

**Timeout:** 60 seconds

### SAGA-I04: Lot/Serial Tracking
**Business Flow:** Generate Lot → Assign Serials → Update Lot Master → Record Genealogy → Update Expiry → Activate Lot → Audit Log

**Forward Steps (7):**
1. lot-serial.GenerateLotNumber
2. lot-serial.AssignSerialNumbers
3. lot-serial.UpdateLotMaster
4. lot-serial.RecordGenealogy - Parent-child relationships
5. lot-serial.UpdateExpiryTracking - For perishable items
6. lot-serial.ActivateLot
7. audit.LogLotCreation - Compliance audit (non-critical)

**Compensation Steps (6):**
- 101-106: Complete rollback

**Timeout:** 60 seconds

### SAGA-I05: Demand Planning & MRP
**Business Flow:** Calculate Forecast → Run MRP → Generate Requisitions → Allocate Stock → Update Parameters → Check Safety Stock → Notify Procurement

**Forward Steps (7):**
1. planning.CalculateForecast - Demand forecasting
2. planning.RunMRP - Material requirements planning
3. procurement.GenerateRequisitions - Auto-generate purchase requisitions
4. inventory-core.AllocateAvailableStock - Allocate existing stock
5. planning.UpdatePlanningParameters - Update lead times, EOQ
6. planning.CheckSafetyStock - Safety stock validation (non-critical)
7. notification.NotifyProcurement - Alert procurement team (non-critical)

**Compensation Steps (6):**
- 102-106: Complete reversal

**Timeout:** 90 seconds

## Usage

### Registering Saga Handlers

The saga handlers are automatically registered via FX modules:

```go
// In your main application
fx.New(
    // ... other modules
    saga.SagaEngineModule,           // Core saga engine
    connector.ConnectorModule,        // RPC connector
    inventory.InventorySagasModule,  // Inventory saga handlers
    inventory.InventorySagasRegistrationModule, // Register with orchestrator
)
```

### Executing a Saga

```go
execution, err := orchestrator.ExecuteSaga(ctx, "SAGA-I01", &saga.SagaExecutionInput{
    TenantID:  "tenant-123",
    CompanyID: "company-456",
    BranchID:  "branch-789",
    Input: map[string]interface{}{
        "source_warehouse_id": "wh-source-001",
        "dest_warehouse_id":   "wh-dest-001",
        "items": []interface{}{
            map[string]interface{}{
                "product_id": "prod-001",
                "quantity": 100,
            },
        },
        "transfer_reason": "Stock rebalancing",
        "expected_delivery_date": "2026-04-14",
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
| SAGA-I01 | source_warehouse_id, dest_warehouse_id, items[], transfer_reason | expected_delivery_date |
| SAGA-I02 | warehouse_id, count_type, approver_id, count_details | variance_threshold |
| SAGA-I03 | receipt_id, product_id, quantity, failure_reason, lot_id | - |
| SAGA-I04 | product_id, quantity, manufacturing_date, serial_start, serial_end | parent_lot_id, expiry_date |
| SAGA-I05 | planning_horizon, forecast_method, historical_periods | allocation_method, eoq_updates |

## Compensation Logic

### Pattern 1: Stock Movement Reversal (SAGA-I01, I02)
When stock is moved and later steps fail:
- Reverse stock transactions (ReverseStockReceipt, ReverseDestinationReceipt)
- Clear in-transit records (DeleteInTransitRecord) for SAGA-I01
- Reverse GL impact (ReverseSourceGL, ReverseDestinationGL)
- Unfreeze stock (UnfreezeStock) for SAGA-I02

### Pattern 2: In-Transit State Management (SAGA-I01)
- In-transit inventory tracked separately (not in either warehouse)
- GL entries must balance both warehouses
- On failure, in-transit records are deleted
- Stock is restored to source warehouse

### Pattern 3: Quality Adjustment Rollback (SAGA-I03)
When QC fails:
- Inspection result is reverted
- Rejection record is deleted
- Stock is moved back to normal location
- Lot status is reset

### Pattern 4: Lot Cleanup (SAGA-I04)
When lot creation fails:
- Lot is deleted completely
- Serial numbers are released
- Lot master is deleted
- Genealogy relationships are cleared
- Expiry tracking is cleared

### Pattern 5: Planning Reversal (SAGA-I05)
When MRP execution fails:
- MRP run is reverted
- Generated requisitions are deleted
- Stock allocations are released
- Planning parameters are reverted

## Compensation Examples

### SAGA-I01 Failure at Destination Receive
```
Forward execution:
Step 1: CreateTransferOrder ✓
Step 2: IssueStockFromSource ✓
Step 3: CreateInTransitRecord ✓
Step 4: PostSourceWarehouseGL ✓
Step 5: CreateInternalShipment ✓
Step 6: ReceiveStockAtDestination ✗ (FAILS - partial receipt)

Compensation triggered:
Step 107: ReverseDestinationGL ✓
Step 106: ReverseDestinationReceipt ✓
Step 105: CancelShipment ✓
Step 104: ReverseSourceGL ✓
Step 103: DeleteInTransitRecord ✓
Step 102: RestoreSourceStock ✓
Step 101: CancelTransferOrder ✓

Result: Transfer cancelled, stock back in source warehouse, no in-transit records
```

### SAGA-I02 Cycle Count Failure
```
Forward execution:
Step 1: CreateCycleCount ✓
Step 2: FreezeStockMovement ✓
Step 3: ExecuteCount ✓
Step 4: CalculateVariance ✓
Step 5: ApproveAdjustment ✗ (FAILS - variance too high)

Compensation triggered:
Step 105: RevertApproval ✓
Step 104: ClearVarianceData ✓
Step 102: UnfreezeStock ✓
Step 101: CancelCycleCount ✓

Result: Stock unfrozen, count cancelled, no GL entries posted
```

## Retry Strategy

| Step Type | Max Retries | Initial Backoff | Max Backoff | Use Case |
|-----------|-------------|-----------------|-------------|----------|
| Internal (inventory-core, cycle-count) | 3 | 1s | 30s | Service dependency |
| Cross-module (stock-transfer, qc) | 3 | 1s | 30s | Coordination |
| External (shipping, notification) | 3 | 1s | 30s | Non-critical |

Exponential backoff formula: `backoff = initial * (2^retry_count) + jitter`

## Service Registry

Service endpoints mapped in `packages/saga/sagas/registry.go`:

```
inventory-core:  http://localhost:8179
wms:             http://localhost:8180
stock-transfer:  http://localhost:8181
qc:              http://localhost:8182
lot-serial:      http://localhost:8183
cycle-count:     http://localhost:8184
barcode:         http://localhost:8185
planning:        http://localhost:8186
shipping:        http://localhost:8188
general-ledger:  http://localhost:8100
notification:    http://localhost:7005
audit:           http://localhost:7007
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

Unit tests are provided in `inventory_sagas_test.go`:

```bash
go test ./packages/saga/sagas/inventory/... -v
```

Tests cover:
- Saga type identification
- Step definition counts
- Input validation
- In-transit state handling
- Variance calculation
- Lot/serial tracking
- Compensation step references

## Compliance & Regulations

### Inter-Warehouse Transfer (SAGA-I01)
- Stock must be issued before receipt
- In-transit inventory not counted in either warehouse
- GL entries must balance both locations
- Audit trail required for stock movement

### Cycle Count (SAGA-I02)
- Stock must be frozen during count
- Variance threshold must be approved
- GL adjustment required for variance
- Audit trail mandatory

### Quality Rejection (SAGA-I03)
- QC inspection required before rejection
- Rejected stock moved to specific location
- Lot status updated to REJECTED
- GL entry posted for inventory loss

### Lot/Serial Tracking (SAGA-I04)
- Lot number auto-generated per product
- Serial numbers assigned within lot range
- Genealogy tracks parent-child relationships
- Expiry tracking for perishable items

### Demand Planning (SAGA-I05)
- Forecast method configurable
- MRP considers safety stock
- Requisitions auto-generated from MRP
- Safety stock validation before completion

## Troubleshooting

### Saga not executing
- Verify handler is registered with correct saga type (SAGA-I01-I05)
- Check input validation passes
- Ensure warehouse/product IDs exist

### Steps failing
- Verify service endpoint is registered in ServiceRegistry
- Check RPC connector can reach service
- Validate warehouse stock levels (for SAGA-I01, I02)

### Compensation not happening
- Confirm CompensationSteps are defined
- Check critical steps fail correctly
- Verify stock unfreezing for SAGA-I02

### In-Transit Issues (SAGA-I01)
- Verify in-transit inventory not counted in warehouse totals
- Check GL entries balance at both locations
- Confirm in-transit records deleted on compensation

### Quality Rejection Issues (SAGA-I03)
- Verify QC service returns correct inspection status
- Check rejected stock location exists
- Confirm lot status updated correctly

## Advanced Algorithms

### SAGA-I04: Lot/Serial Tracking Algorithms

#### Batch ID Generation
- Format: `LOT-YYYYMM-XXXXX`
- YYYYMM = Current year-month (e.g., 202602 for Feb 2026)
- XXXXX = 5-digit sequential counter unique within month
- Auto-increment logic: Counter resets on month boundary
- Example: LOT-202602-00001, LOT-202602-00002, etc.

#### Genealogy Tree Management
- Parent-child relationship tracking
- Supports multi-level genealogy (grandparent → parent → child)
- Used for:
  - Production BOM tracking (raw materials → finished goods)
  - Sub-lot creation (bulk lot → smaller lots)
  - Traceability chains
- Query: "Find all ancestors of lot X" (recursive parent walk)
- Query: "Find all descendants of lot X" (recursive child walk)

#### Traceability Chain
```
Lot → Stock Location → Warehouse → Sales Invoice → Customer
```
- Enables forward trace: Find all customers who received product from specific lot
- Critical for pharma/food recalls: Identify affected customers
- Typical recall scenario:
  1. Quality issue detected in LOT-202602-00001
  2. Query: Which customers received this lot?
  3. Generate recall notification list

#### Expiry Management
- Expiry date = Manufacturing Date + Shelf Life
- For perishable goods:
  - Flag items 30 days before expiry (warning)
  - Block sales 1 day before expiry
  - Auto-move to disposal queue on expiry date
- Shelf life by product type:
  - Pharma: 24-60 months typical
  - Food: 3-12 months typical
  - Electronics: N/A (no expiry)

#### Recall Management
- When recall triggered:
  1. Mark lot as RECALLED
  2. Find all customers (via sales invoices)
  3. Generate recall notification
  4. Track customer acknowledgment
  5. Option to accept returns or run corrective action

### SAGA-I05: Demand Planning & Auto-Reorder Algorithms

#### Exponential Smoothing Forecast
```
Formula: F_t = α * D_(t-1) + (1-α) * F_(t-1)
Where:
  α = 0.3 (weighting factor for recent data)
  D = Actual demand
  F = Forecast
```
- Weights recent months more heavily (70% of weight on last period)
- Smooths seasonal variations
- Calculated monthly/weekly based on planning frequency
- Input: Historical sales data (minimum 12-24 periods recommended)

#### Safety Stock Calculation
```
Formula: SS = Z * σ * √L
Where:
  Z = Z-score (1.65 for 95% service level, 2.33 for 99%)
  σ = Standard deviation of demand
  L = Lead time in days

Service Levels:
  90% (Z=1.28) → Lower safety stock, higher stockout risk
  95% (Z=1.65) → Balanced (recommended default)
  99% (Z=2.33) → Higher safety stock, lower stockout risk
```
- Protects against demand variability and lead time uncertainty
- Higher service level = higher inventory cost but lower stockout cost
- Trade-off: Increased carrying cost vs. lost sales

#### Reorder Point (ROP)
```
Formula: ROP = (d̄ * LT) + SS
Where:
  d̄ = Average daily consumption
  LT = Lead time in days
  SS = Safety stock (from above)

Example:
  Daily consumption = 50 units
  Lead time = 10 days
  Safety stock = 300 units
  ROP = (50 * 10) + 300 = 800 units
```
- Trigger point: When inventory falls below ROP, auto-generate PO
- Ensures fresh supply arrives before stockout

#### Economic Order Quantity (EOQ)
```
Formula: EOQ = √(2 * D * S / H)
Where:
  D = Annual demand (units)
  S = Ordering cost per order ($)
  H = Holding/carrying cost per unit per year ($)

Example:
  Annual demand = 10,000 units
  Order cost = $50 per order
  Holding cost = $2 per unit per year
  EOQ = √(2 * 10000 * 50 / 2) = √500,000 = 707 units
```
- Optimal order quantity minimizing total inventory cost
- Ordering cost + Holding cost are balanced
- Order frequency = Annual demand / EOQ
- In above example: 10,000 / 707 ≈ 14 orders per year

#### MRP Logic
- Backward scheduling from demand to procurement:
  1. Start with forecasted demand for month
  2. Subtract current inventory
  3. Add safety stock requirement
  4. Calculate gross requirement
  5. Apply supplier lead times
  6. Generate PO for arrival before demand date
- Prevents both stockout and excess inventory
- Re-runs typically weekly/monthly

#### Lead Time Buffering
- Supplier lead time: Time from PO to goods receipt
- Must order: LT days before demand date
- If demand in week 4, and LT=10 days → Order in week 2
- Considers supplier reliability:
  - Reliable suppliers: Use nominal lead time
  - Unreliable suppliers: Add buffer days

## Future Enhancements

- Partial shipment support for SAGA-I01
- Cross-company transfer GL adjustments
- Advanced variance analysis for SAGA-I02
- Lot genealogy visualization
- Real-time demand planning dashboard
- Performance metrics collection
- Saga state replay for debugging
- ML-based forecast method selection
- Dynamic safety stock recalculation
- Supplier performance scoring
