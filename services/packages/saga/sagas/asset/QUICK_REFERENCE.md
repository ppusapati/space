# Asset Sagas Quick Reference

## Saga Types & Purpose

| Saga ID | Name | Steps | Critical | Timeout | Purpose |
|---------|------|-------|----------|---------|---------|
| A01 | Asset Acquisition | 10 | 3,5,7,8 | 300s | IAS 16 capitalization, GL posting, depreciation schedule creation |
| A02 | Asset Depreciation | 8 | 3,6,7 | 180s | Monthly accrual, GL posting, NBV update |
| A03 | Asset Disposal | 9 | 1,2,4,6,8 | 240s | Gain/loss calculation, GL posting, asset retirement |
| A04 | Asset Revaluation | 7 | 1,2,3,5 | 180s | Fair value determination, revaluation reserve, accumulated depr reset |

## Key Concepts

### Asset Acquisition (A01)
- **When:** Purchase of depreciable asset >1 year useful life
- **Key Step:** Step 7 (GL posting) - must succeed
- **GL Structure:** FA Dr / Creditor Cr
- **Failure Path:** 9/10 compensation if step 8 fails (schedule creation)

### Asset Depreciation (A02)
- **When:** Monthly (scheduler trigger)
- **Batch:** Steps 1-4 process all ACTIVE assets
- **GL Structure:** Depr Exp Dr / Accum Depr Cr
- **Key Constraint:** Cannot depreciate below salvage value (Step 3)

### Asset Disposal (A03)
- **When:** Sale, scrap, or donate
- **GL Variance:** Depends on gain vs. loss
- **Gain:** Cash Dr / Asset Cr / Accum Depr Dr / Gain Cr
- **Loss:** Cash Dr / Loss Dr / Asset Cr / Accum Depr Dr
- **Key Step:** Step 2 (NBV retrieval) determines calculation

### Asset Revaluation (A04)
- **When:** Annual or event-based
- **Reserve:** Equity account, not P&L
- **Depreciation Reset:** Option per IAS 16
- **Tracking:** Multiple revaluation history maintained

## Input Templates

### A01: Asset Acquisition
```json
{
  "po_id": "PO001",
  "po_line_id": "POL001",
  "part_number": "PART123",
  "quantity": 1,
  "asset_tag": "ASSET001",
  "asset_category": "Building",
  "asset_location": "Warehouse A",
  "description": "Industrial Building",
  "serial_number": "SN12345",
  "manufacturer_name": "MFG Corp",
  "po_amount": 500000.0,
  "freight_cost": 5000.0,
  "installation_cost": 10000.0,
  "customs_duty": 2500.0,
  "useful_life_years": 40.0,
  "depreciation_method": "SLM",
  "capitalization_threshold": 50000.0,
  "received_date": "2026-02-16",
  "start_depreciation_date": "2026-03-01",
  "fixed_asset_account": "1200",
  "creditor_account": "2100",
  "created_by": "USER123"
}
```

### A02: Asset Depreciation
```json
{
  "depreciation_date": "2026-02-28",
  "period_id": "PERIOD202602",
  "depreciation_expense_account": "6100",
  "accumulated_depreciation_account": "1500",
  "journal_description": "Monthly Depreciation - February 2026"
}
```

### A03: Asset Disposal (SALE)
```json
{
  "asset_id": "ASSET001",
  "disposal_type": "SALE",
  "disposal_reason": "Obsolete Equipment",
  "disposal_date": "2026-02-16",
  "sale_price": 50000.0,
  "buyer_name": "ABC Trading Ltd",
  "buyer_type": "CORPORATE",
  "payment_terms": "NET30",
  "approver_id": "APPROVER001",
  "fixed_asset_account": "1200",
  "accumulated_depreciation_account": "1500",
  "gain_on_sale_account": "7500",
  "loss_on_sale_account": "7600",
  "cash_account": "1000",
  "accounts_receivable_account": "1100"
}
```

### A04: Asset Revaluation
```json
{
  "asset_id": "ASSET001",
  "revaluation_date": "2026-02-16",
  "valuation_method": "APPRAISAL",
  "appraisal_amount": 150000.0,
  "prior_book_value": 100000.0,
  "accumulated_depreciation": 25000.0,
  "trigger_type": "ANNUAL",
  "revaluation_reason": "Annual Fair Value Assessment",
  "appraiser_name": "Valuation Expert Inc",
  "appraisal_certificate": "VAL-2026-001",
  "fixed_asset_account": "1200",
  "accumulated_depreciation_account": "1500",
  "revaluation_reserve_account": "3150",
  "gain_on_revaluation_account": "7500",
  "loss_on_revaluation_account": "7600",
  "depreciation_reset_option": "ELIMINATE"
}
```

## Service Mapping

### All Sagas Use These Services

| Service | Methods | Purpose |
|---------|---------|---------|
| `asset` | CreateAssetMasterRecord, UpdateAssetStatus, UpdateAssetStatusToRetired, UpdateAssetNBV, ArchiveDisposalRecord, ArchiveRevaluationRecord, ExtractActiveAssets, TriggerRevaluation, InitiateAssetDisposal | Asset master operations |
| `fixed-assets` | ValidateCapitalizationCriteria, CalculateCostBasis, CalculateDepreciableBasis, RetrieveCurrentNBV, CalculateGainLoss, ApplyDepreciationCap, CalculateRevaluationAmount, DetermineRevaluationTreatment, DetermineFairValue | Financial calculations |
| `depreciation` | DetermineDepreciationMethod, CalculateMonthlyDepreciation, CalculateAccumulatedDepreciationImpact, CreateDepreciationJournalEntry, CreateDepreciationSchedule, ArchiveDepreciationEntry, RemoveFromDepreciationSchedule, ResetAccumulatedDepreciation, CreateDisposalJournalEntry | Depreciation operations |
| `general-ledger` | PostAssetAcquisitionGL, PostDepreciationToGL, PostAssetDisposalGL, PostAssetRevaluationGL | GL posting |
| `purchase-order` | ReceiveGoodsFromPO | PO integration |
| `accounts-receivable` | DetermineSaleProceeds | AR for disposals |
| `notification` | SendAssetCreationNotification | Event notifications |
| `approval` | (implicit) | Disposal approval |

## Error Scenarios & Recovery

### A01: Asset Acquisition
| Failure | Step | Compensation |
|---------|------|--------------|
| Validation fails | 2 | None (pre-GL) |
| Asset creation fails | 3 | Revert PO (101) |
| GL posting fails | 7 | Reverse GL (106) + delete asset |
| Schedule creation fails | 8 | Delete schedule (107) |

### A02: Asset Depreciation
| Failure | Step | Recovery |
|---------|------|----------|
| Asset extraction fails | 1 | Retry with backoff |
| Depreciation cap exceeds | 3 | Log warning, use capped amount |
| GL posting fails | 6 | Reverse GL entries (205) |

### A03: Asset Disposal
| Failure | Step | Recovery |
|---------|------|----------|
| Retrieval of NBV fails | 2 | Mark asset as locked, alert admin |
| GL posting fails | 6 | Reverse entries (304), unlock asset |
| Status update fails | 7 | Manual intervention required |

### A04: Asset Revaluation
| Failure | Step | Recovery |
|---------|------|----------|
| Fair value determination fails | 2 | Reject, manual appraisal required |
| GL posting fails | 5 | Reverse GL (403), restore prior values |

## Validation Rules Quick Ref

### Disposal Types
```
SALE   → requires: sale_price, buyer_name
SCRAP  → optional: scrap_value
DONATE → optional: donation_org
```

### Depreciation Methods
```
SLM   → (Cost - Salvage) / Years / 12
WDV   → (Cost - Accum) * Rate / 100 / 12
UNITS → (Cost - Salvage) / Units * Units_Produced
```

### Valuation Methods
```
MARKET    → Market comparable sales
APPRAISAL → Professional appraisal
INCOME    → Discounted cash flows
COST      → Cost basis adjusted for inflation
```

## Common GL Accounts

| Account | Purpose | A01 | A02 | A03 | A04 |
|---------|---------|-----|-----|-----|-----|
| 1000 | Cash | ✓ | | ✓ | |
| 1100 | Accounts Receivable | ✓ | | ✓ | |
| 1200 | Fixed Assets | ✓ | | ✓ | ✓ |
| 1500 | Accumulated Depreciation | ✓ | ✓ | ✓ | ✓ |
| 2100 | Creditors | ✓ | | | |
| 3150 | Revaluation Reserve | | | | ✓ |
| 6100 | Depreciation Expense | | ✓ | | |
| 7500 | Gain on Disposal | | | ✓ | |
| 7600 | Loss on Disposal | | | ✓ | |

## Compensation Step Mapping

| Saga | Forward Steps | Compensation |
|------|--------------|--------------|
| A01 | 1-10 | 101-108 |
| A02 | 1-8 | 201-207 |
| A03 | 1-9 | 301-307 |
| A04 | 1-7 | 401-405 |

## Testing Commands

```bash
# Run all asset saga tests
go test ./packages/saga/sagas/asset/... -v

# Run specific saga tests
go test ./packages/saga/sagas/asset/... -run TestAssetAcquisition -v

# Run with coverage
go test ./packages/saga/sagas/asset/... -cover

# Run integration tests only
go test ./packages/saga/sagas/asset/... -run TestAll -v
```

## Deployment Checklist

- [ ] Update go.mod with saga package imports
- [ ] Verify all service method implementations
- [ ] Configure GL account mappings in input
- [ ] Test with actual PO data (A01)
- [ ] Verify monthly scheduler trigger (A02)
- [ ] Test gain vs. loss scenarios (A03)
- [ ] Validate appraisal workflow (A04)
- [ ] Run full integration test
- [ ] Deploy to staging
- [ ] User acceptance testing
- [ ] Production deployment

## Monitoring Points

### Critical Metrics
- A01: Step 7 GL posting success rate
- A02: Monthly processing duration (all assets)
- A03: Gain/loss calculation accuracy
- A04: Revaluation reserve consistency

### Alert Thresholds
- Retry failures > 5%
- GL posting errors > 1%
- Timeout occurrences > 0.1%
- Compensation activations > 1%

## Quick Troubleshooting

| Issue | Cause | Solution |
|-------|-------|----------|
| Asset not created | Step 3 failed | Check asset_tag uniqueness |
| GL posting fails | Account invalid | Verify GL account numbers |
| Depreciation skipped | Asset not ACTIVE | Check asset status in DB |
| Gain/loss = 0 | NBV = proceeds | Verify sale_price calculation |
| Revaluation fails | Appraisal missing | Provide appraisal_certificate |

## Documentation Files

- `IMPLEMENTATION_SUMMARY.md` - Complete technical documentation
- `QUICK_REFERENCE.md` - This file
- Individual saga files contain inline comments
- Test file demonstrates all validation scenarios
