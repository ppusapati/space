# Phase 4: Advanced Distributed Sagas Implementation

**Status:** In Development
**Last Updated:** February 15, 2026
**Target Completion:** Q1 2026

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Phase 4A: Foundation Sagas](#phase-4a-foundation-sagas)
3. [Phase 4B: Core Operation Sagas](#phase-4b-core-operation-sagas)
4. [Phase 4C: Critical System Sagas](#phase-4c-critical-system-sagas)
5. [Module Integration Guide](#module-integration-guide)
6. [Saga Engine Architecture](#saga-engine-architecture)
7. [Service Registry](#service-registry)
8. [Implementation Patterns](#implementation-patterns)
9. [Testing Strategy](#testing-strategy)
10. [FX Module Integration](#fx-module-integration)
11. [Step-by-Step Implementation Guide](#step-by-step-implementation-guide)
12. [Deployment & Operations](#deployment--operations)
13. [Glossary & References](#glossary--references)

---

## Executive Summary

Phase 4 extends the samavaya ERP saga engine with **24 advanced distributed sagas** across **4 enterprise modules** (Finance, Manufacturing, HR, Projects). These sagas coordinate complex multi-service workflows with eventual consistency, automatic compensation, and fault tolerance.

### Phase 4 Overview

| Metric | Value |
|--------|-------|
| **Total Sagas** | 24 |
| **Phase 4A Sagas** | 6 (Foundation) |
| **Phase 4B Sagas** | 10 (Core Operations) |
| **Phase 4C Sagas** | 8 (Critical Systems) |
| **Total Steps** | 200+ forward steps |
| **Compensation Steps** | 180+ compensation steps |
| **Files Created** | 32 (sagas + tests + FX modules) |
| **Lines of Code** | 8,500+ |
| **Service Integrations** | 35+ services |
| **Test Coverage** | 190+ unit tests |
| **Expected Coverage** | 90%+ |

### Architecture Highlights

- **Distributed Transaction Coordination:** Multi-service workflows with ACID-like guarantees
- **Eventual Consistency Model:** Compensation-based rollback instead of 2-phase commit
- **Compensation-Based Rollback:** Reverse steps with idempotency to ensure consistency
- **Circuit Breaker Protection:** Prevent cascading failures across services
- **Exponential Backoff Retry:** Intelligent retry with jitter (1s, 2s, 4s, 8s, ...)
- **Kafka-Based Event Publishing:** Asynchronous audit trail and event notification
- **Idempotent Step Execution:** Safe replay of steps without side effects

### Key Design Decisions

1. **Compensation vs. 2-Phase Commit**
   - Sagas use compensation (forward with optional backward) instead of 2PC
   - Reason: Better availability, no distributed locks, tolerates service failures
   - Trade-off: Eventual consistency vs. immediate ACID guarantees

2. **Service-Based Architecture**
   - Sagas coordinate independent services via ConnectRPC
   - Reason: Loose coupling, independent scaling, technology flexibility
   - Trade-off: Network latency, eventual consistency, complexity

3. **Kafka for Event Publishing**
   - All saga events published asynchronously to Kafka
   - Reason: Decoupled event consumers, audit trail, ordering guarantees
   - Trade-off: Additional infrastructure dependency

4. **Circuit Breaker Pattern**
   - Auto-failure on repeated service failures
   - Reason: Prevent cascading failures, fast-fail instead of retry storms
   - Trade-off: Temporary unavailability vs. resource exhaustion

5. **100+ Compensation Numbering**
   - Compensation steps numbered 101+, 102+, 103+, etc.
   - Reason: Clear separation, easy identification, predictable ordering
   - Trade-off: Limits forward steps to 100 (sufficient for complexity)

### Timeline & Phases

```
Phase 4A (Feb 15-20):
  ├─ F05: Revenue Recognition
  ├─ F06: Asset Capitalization
  ├─ F07: GST Credit Reversal
  ├─ F08: Cost Center Allocation
  ├─ H05: Leave Application
  └─ M05: Job Card Consumption

Phase 4B (Feb 20-28):
  ├─ M03: BOM Explosion & MRP (3 steps)
  ├─ M04: Production Order (4 steps)
  ├─ M06: Quality Rework (3 steps)
  ├─ H02: Employee Onboarding (4 steps)
  ├─ H04: Appraisal & Salary Revision (4 steps)
  ├─ H06: Expense Reimbursement (3 steps)
  ├─ PR01: Project Billing (4 steps)
  ├─ PR02: Progress Billing (5 steps)
  ├─ PR03: Subcontractor Payment (4 steps)
  └─ PR04: Project Close (4 steps)

Phase 4C (Feb 28-Mar 15):
  ├─ F01: Month-End Close (12 steps, no compensation)
  ├─ F02: Bank Reconciliation (11 steps)
  ├─ F03: Multi-Currency Revaluation (8 steps)
  ├─ F04: Intercompany Transaction (7 steps)
  ├─ H01: Payroll Processing (10 steps)
  ├─ H03: Employee Exit (8 steps)
  ├─ M01: Production Order Release (6 steps)
  └─ M02: Subcontracting (8 steps)
```

---

## Phase 4A: Foundation Sagas

Foundation sagas provide essential workflows for establishing baseline operational processes. These sagas serve as building blocks for more complex workflows in Phases 4B and 4C.

### Overview

Phase 4A implements **6 foundational sagas** across Finance, HR, and Manufacturing:

| Saga ID | Module | Name | Steps | Compensation | Status |
|---------|--------|------|-------|--------------|--------|
| **F05** | Finance | Revenue Recognition | 6 | 5 | In Progress |
| **F06** | Finance | Asset Capitalization | 5 | 4 | In Progress |
| **F07** | Finance | GST Credit Reversal | 5 | 4 | In Progress |
| **F08** | Finance | Cost Center Allocation | 6 | 5 | In Progress |
| **H05** | HR | Leave Application | 4 | 3 | In Progress |
| **M05** | Manufacturing | Job Card Consumption | 5 | 4 | In Progress |

### SAGA-F05: Revenue Recognition

**Business Purpose:** Recognize revenue from customer invoices according to accounting policies and compliance requirements.

**Workflow:**
1. Retrieve customer invoice and contract terms
2. Calculate revenue recognition amount based on policy
3. Create revenue recognition journal entries
4. Post to general ledger
5. Update accounts receivable subsidiary ledger
6. Record in revenue recognition register

**Steps:** 6 forward + 5 compensation = 11 total

**Critical Steps:** 1, 2, 3, 4 (accounting compliance)

**Key Services:**
- Accounts Receivable (AR)
- General Ledger (GL)
- Journal Engine
- Compliance Rules Engine

**Timeout Configuration:**
- Step 1-2: 15 seconds (lookup)
- Step 3-5: 20 seconds (posting)
- Step 6: 10 seconds (update)

**Input Requirements:**
```
{
  "tenantID": "tenant-123",
  "companyID": "company-456",
  "branchID": "branch-789",
  "input": {
    "invoice_id": "INV-2026-001",
    "revenue_date": "2026-02-15",
    "policy_id": "POL-REV-ASC606"
  }
}
```

**Example Workflow:**
1. User posts invoice in AR
2. System triggers SAGA-F05 execution
3. Saga retrieves invoice details and contract terms
4. Calculates revenue per ASC 606 standard (step 2 - Critical)
5. Creates GL journal entries (step 3 - Critical)
6. Posts to subsidiary ledger (step 4 - Critical)
7. Updates AR aging on completion
8. If step 4 fails: executes compensation steps in reverse (steps 104, 103, 102, 101)

---

### SAGA-F06: Asset Capitalization

**Business Purpose:** Capitalize eligible capital assets, set up depreciation schedules, and manage asset tags.

**Workflow:**
1. Validate asset acquisition details and cost
2. Check capitalization policy eligibility
3. Create fixed asset record with cost basis
4. Setup depreciation schedule
5. Create asset-to-GL mapping

**Steps:** 5 forward + 4 compensation = 9 total

**Critical Steps:** 1, 2, 3, 4

**Key Services:**
- Fixed Assets
- General Ledger
- Depreciation Engine
- Asset Management

**Timeout Configuration:**
- Step 1: 12 seconds (validation)
- Step 2: 15 seconds (policy check)
- Step 3-4: 20 seconds (record creation)
- Step 5: 10 seconds (mapping)

**Input Requirements:**
```
{
  "tenantID": "tenant-123",
  "companyID": "company-456",
  "input": {
    "asset_id": "AST-2026-001",
    "asset_cost": 500000.00,
    "asset_date": "2026-02-15",
    "depreciation_method": "STRAIGHT_LINE",
    "useful_life_years": 5
  }
}
```

---

### SAGA-F07: GST Credit Reversal

**Business Purpose:** Reverse GST credit on blocked items/transactions per regulatory compliance.

**Workflow:**
1. Identify transactions with blocked GST credit
2. Calculate blocked credit amount
3. Create reversing journal entry
4. Post to GST clearing account
5. Update GSTR-1 adjustment register

**Steps:** 5 forward + 4 compensation = 9 total

**Critical Steps:** 1, 3, 4

**Key Services:**
- GST Engine
- General Ledger
- Tax Compliance
- Journal Engine

**Timeout Configuration:**
- Step 1-2: 15 seconds
- Step 3-4: 20 seconds
- Step 5: 10 seconds

---

### SAGA-F08: Cost Center Allocation

**Business Purpose:** Allocate indirect costs to cost centers based on allocation bases.

**Workflow:**
1. Retrieve indirect cost transactions
2. Calculate allocation percentages
3. Create allocation journal entries
4. Post to cost center GL accounts
5. Update cost center balances
6. Create allocation audit log

**Steps:** 6 forward + 5 compensation = 11 total

**Critical Steps:** 2, 3, 4, 5

**Key Services:**
- Cost Center Service
- General Ledger
- Journal Engine
- Allocation Engine

---

### SAGA-H05: Leave Application

**Business Purpose:** Process employee leave applications with approval workflow and balance updates.

**Workflow:**
1. Validate employee eligibility and leave balance
2. Submit leave request
3. Route to approver (manager)
4. Update leave balance on approval

**Steps:** 4 forward + 3 compensation = 7 total

**Critical Steps:** 1, 4

**Key Services:**
- HR / Employee Service
- Leave Management
- Approval Workflow
- Notification

**Timeout Configuration:**
- Step 1-2: 10 seconds (validation)
- Step 3: 5 seconds (routing)
- Step 4: 15 seconds (update)

**Input Requirements:**
```
{
  "tenantID": "tenant-123",
  "companyID": "company-456",
  "input": {
    "employee_id": "EMP-001",
    "leave_type": "ANNUAL",
    "from_date": "2026-03-01",
    "to_date": "2026-03-05",
    "days": 5
  }
}
```

---

### SAGA-M05: Job Card Consumption

**Business Purpose:** Record material consumption from job cards during production.

**Workflow:**
1. Retrieve job card and production details
2. Record material consumption
3. Update inventory GL accounts
4. Update work-in-process inventory
5. Create consumption audit trail

**Steps:** 5 forward + 4 compensation = 9 total

**Critical Steps:** 2, 3, 4

**Key Services:**
- Job Card Service
- Inventory Management
- Production Planning
- General Ledger

---

## Phase 4B: Core Operation Sagas

Core operation sagas implement standard business processes that drive daily operations across manufacturing, HR, and projects.

### Overview

Phase 4B implements **10 sagas** with 8-10 steps each across three modules:

| Module | Saga ID | Name | Steps | Status |
|--------|---------|------|-------|--------|
| **Manufacturing** | M03 | BOM Explosion & MRP | 8+6 | In Progress |
| | M04 | Production Order Release | 6+5 | In Progress |
| | M06 | Quality Rework | 5+4 | In Progress |
| **HR** | H02 | Employee Onboarding | 10+8 | In Progress |
| | H04 | Appraisal & Salary Revision | 7+6 | In Progress |
| | H06 | Expense Reimbursement | 6+5 | In Progress |
| **Projects** | PR01 | Project Billing | 7+6 | In Progress |
| | PR02 | Progress Billing | 8+7 | In Progress |
| | PR03 | Subcontractor Payment | 6+5 | In Progress |
| | PR04 | Project Close | 7+6 | In Progress |

### Manufacturing: SAGA-M03 - BOM Explosion & MRP

**Business Purpose:** Explode bill of materials, plan material requirements, and schedule procurement for sales orders.

**Saga Flow:**

```
Sales Order
     ↓
[Step 1] Get Sales Order Details (critical) ────┐
     ↓                                           │
[Step 2] Explode BOM (critical) ◄────────────────┤
     ↓                                           │
[Step 3] Plan Material Requirements (critical) ──┤
     ↓                                           │
[Step 4] Schedule Procurement (critical) ────────┤
     ↓                                           │
[Step 5] Update Inventory Plan (critical) ───────┤
     ↓                                           │
[Step 6] Create Manufacturing Orders            │ Compensation
     ↓                                           │ Chain on
[Step 7] Confirm Schedule (non-critical)        │ Failure
     ↓                                           │
[Step 8] Complete MRP Run (critical) ────────────┘
     ↓
Complete with orders scheduled
```

**Step Details:**

| Step | Service | Method | Critical | Timeout | Input | Compensation |
|------|---------|--------|----------|---------|-------|--------------|
| 1 | sales-order | GetSalesOrderDetails | Yes | 15s | sales_order_id | None |
| 2 | bom | ExploadBillOfMaterial | Yes | 20s | product_id, bom_id | Step 101 |
| 3 | production-planning | PlanRequirements | Yes | 20s | bom_lines, quantity | Step 102 |
| 4 | procurement | ScheduleRequisitions | Yes | 20s | material_requirements | Step 103 |
| 5 | inventory-core | UpdateForecastDemand | Yes | 15s | forecast_data | Step 104 |
| 6 | production-order | CreateOrders | No | 20s | order_specs | Step 105 |
| 7 | work-center | ConfirmSchedule | No | 20s | order_ids | Step 106 |
| 8 | production-planning | CompleteMRPRun | Yes | 15s | completion_data | None |

**Compensation Steps:**

- **Step 101:** Revert BOM Explosion - Clears exploded BOM data
- **Step 102:** Revert Material Planning - Removes planned requirements
- **Step 103:** Revert Procurement Scheduling - Cancels requisitions
- **Step 104:** Revert Inventory Plan - Removes forecast adjustments
- **Step 105:** Revert Manufacturing Orders - Cancels created orders
- **Step 106:** Revert Schedule Confirmation - Releases work center capacity

**Example Workflow:**

```
Input:
  sales_order_id: "SO-2026-001"
  product_id: "PROD-123"
  bom_id: "BOM-123"
  quantity: 100

Execution:
1. Retrieve SO details (order qty: 100, required date: 2026-03-15)
2. Explode BOM (components: Mat1: 50 units, Mat2: 30 units, Mat3: 20 units)
3. Plan requirements (with safety stock, lead times)
4. Create requisitions (PO proposals to suppliers)
5. Update inventory forecast (adjust available qty)
6. Create manufacturing orders (production order MO-001)
7. Schedule work centers (assign to CNC, Assembly, QC)
8. Mark MRP run complete

Result:
  mro_id: "MO-2026-001"
  requisition_ids: ["REQ-001", "REQ-002", "REQ-003"]
  scheduled_date: "2026-03-10"
```

---

### Manufacturing: SAGA-M04 - Production Order Release

**Business Purpose:** Release production orders to shop floor with BOM, routing, and capacity confirmation.

**Steps:** 6 forward + 5 compensation = 11 total

**Critical Steps:** 1, 2, 3, 5

**Key Services:**
- Production Order Service
- Routing Engine
- Work Center Management
- Shop Floor Control
- Material Management

**Workflow:**
1. Validate production order and BOM
2. Confirm material availability
3. Release to shop floor
4. Create job cards
5. Reserve material in inventory
6. Notify work centers

---

### Manufacturing: SAGA-M06 - Quality Rework

**Business Purpose:** Handle quality inspection failures with rework orders and scrap management.

**Steps:** 5 forward + 4 compensation = 9 total

**Critical Steps:** 1, 3, 4

**Key Services:**
- Quality Control Service
- Production Order Service
- Inventory Management
- Scrap Management

---

### HR: SAGA-H02 - Employee Onboarding

**Business Purpose:** Complete employee onboarding with account creation, benefits setup, and orientation.

**Steps:** 10 forward + 8 compensation = 18 total

**Critical Steps:** 1, 2, 3, 4, 5

**Workflow:**
1. Create employee record
2. Assign employee ID and email
3. Setup payroll account
4. Enroll in benefits
5. Schedule orientation
6. Create asset assignment (laptop, badge)
7. Setup system access (permissions)
8. Send welcome notification
9. Create onboarding checklist
10. Record onboarding completion

**Key Services:**
- HR / Employee Master
- Payroll
- Benefits Management
- Notification Service
- Access Control
- Asset Management

**Example Workflow:**

```
New Employee: Priya Sharma

Step 1: Create Employee Record
  → emp_id: EMP-2026-001, status: ACTIVE

Step 2: Assign Credentials
  → email: priya.sharma@company.com
  → emp_code: EMP001

Step 3: Setup Payroll
  → salary_structure_id: SS-ENGINEER-2026
  → bank_account: validated

Step 4: Enroll Benefits
  → health_insurance: HI-2026
  → life_insurance: LI-2026

Step 5: Schedule Orientation
  → date: 2026-03-01
  → trainer: HR_ADMIN_001

Step 6-10: Asset provisioning, access setup, notifications
```

---

### HR: SAGA-H04 - Appraisal & Salary Revision

**Business Purpose:** Execute annual appraisal cycle with performance ratings and salary adjustments.

**Steps:** 7 forward + 6 compensation = 13 total

**Critical Steps:** 1, 3, 5

**Key Services:**
- Appraisal Management
- Salary Structure
- Payroll
- HR Master
- Notification

---

### HR: SAGA-H06 - Expense Reimbursement

**Business Purpose:** Process employee expense reimbursements with approval and payment.

**Steps:** 6 forward + 5 compensation = 11 total

**Critical Steps:** 1, 3, 5

**Key Services:**
- Expense Management
- Approval Workflow
- Accounts Payable
- Payment Processing
- Bank Integration

---

### Projects: SAGA-PR01 - Project Billing

**Business Purpose:** Generate and process project billing based on milestones and deliverables.

**Steps:** 7 forward + 6 compensation = 13 total

**Critical Steps:** 2, 4, 6

**Workflow:**
1. Retrieve project details and billing terms
2. Calculate billable amounts (milestone/time-based)
3. Create project invoice
4. Post to accounts receivable
5. Send to customer
6. Record in project costing
7. Update billing status

**Key Services:**
- Project Management
- Project Costing
- Accounts Receivable
- Invoice Generation
- Customer Portal

---

### Projects: SAGA-PR02 - Progress Billing

**Business Purpose:** Generate progress-based invoices during project execution.

**Steps:** 8 forward + 7 compensation = 15 total

**Critical Steps:** 2, 4, 5, 7

**Key Services:**
- Project Management
- Progress Tracking
- Accounts Receivable
- Invoice Generation
- Project Analytics

---

### Projects: SAGA-PR03 - Subcontractor Payment

**Business Purpose:** Manage subcontractor billing and payment processing.

**Steps:** 6 forward + 5 compensation = 11 total

**Critical Steps:** 1, 3, 5

**Key Services:**
- Subcontractor Management
- Accounts Payable
- Invoice Processing
- Payment Processing

---

### Projects: SAGA-PR04 - Project Close

**Business Purpose:** Close completed projects with final billing, settlement, and archival.

**Steps:** 7 forward + 6 compensation = 13 total

**Critical Steps:** 2, 4, 5, 6

**Key Services:**
- Project Management
- Project Costing
- Financial Close
- AR / AP Settlement
- Archive Service

---

## Phase 4C: Critical System Sagas

Critical system sagas implement essential financial and HR processes with strict compliance requirements. These sagas have the highest availability requirements and most stringent testing.

### Overview

Phase 4C implements **8 critical sagas** with 10+ steps and rigorous compensation:

| Module | Saga ID | Name | Steps | Type | Status |
|--------|---------|------|-------|------|--------|
| **Finance** | F01 | Month-End Close | 12 | No Compensation | In Progress |
| | F02 | Bank Reconciliation | 11 | With Compensation | In Progress |
| | F03 | Multi-Currency Revaluation | 8 | With Compensation | In Progress |
| | F04 | Intercompany Transaction | 7 | With Compensation | In Progress |
| **HR** | H01 | Payroll Processing | 10 | With Compensation | In Progress |
| | H03 | Employee Exit | 8 | With Compensation | In Progress |
| **Manufacturing** | M01 | Production Order Release | 6 | With Compensation | In Progress |
| | M02 | Subcontracting | 8 | With Compensation | In Progress |

### Finance: SAGA-F01 - Month-End Close

**SPECIAL CASE: No Compensation Steps**

**Business Purpose:** Perform comprehensive monthly financial close with accruals, consolidations, and reporting.

**Saga Flow:**

```
Month-End Trigger
     ↓
[Step 1] Lock accounting period (critical)
     ↓
[Step 2] Validate GL balances (critical)
     ↓
[Step 3] Process accruals (critical)
     ↓
[Step 4] Reconcile intercompany (critical)
     ↓
[Step 5] Calculate provisions (critical)
     ↓
[Step 6] Process tax adjustments (critical)
     ↓
[Step 7] Consolidate financials (critical)
     ↓
[Step 8] Generate trial balance (critical)
     ↓
[Step 9] Create standard journal entries (critical)
     ↓
[Step 10] Post closing entries (critical)
     ↓
[Step 11] Generate financial statements (critical)
     ↓
[Step 12] Unlock period for reporting
     ↓
Close Complete
```

**Why No Compensation:**
- Month-end close is atomic - either fully succeeds or fully fails
- Partial compensation would leave accounting records in inconsistent state
- Retries and exceptions escalate to financial controller
- Once step succeeds, moving backward is more risky than continuing forward
- All steps are critical; failure of any step stops entire process

**Step Details:**

| Step | Service | Method | Critical | Timeout | Purpose |
|------|---------|--------|----------|---------|---------|
| 1 | general-ledger | LockAccountingPeriod | Yes | 10s | Prevent further posting |
| 2 | general-ledger | ValidateGLBalances | Yes | 30s | Verify GL integrity |
| 3 | general-ledger | ProcessAccruals | Yes | 45s | Accrue expenses/revenues |
| 4 | general-ledger | ReconcileIntercompany | Yes | 60s | Match IC transactions |
| 5 | tax-engine | CalculateProvisions | Yes | 45s | Tax/legal provisions |
| 6 | tax-engine | ProcessTaxAdjustments | Yes | 45s | Monthly tax adjustments |
| 7 | general-ledger | ConsolidateFinancials | Yes | 120s | Company consolidation |
| 8 | general-ledger | GenerateTrialBalance | Yes | 30s | TB verification |
| 9 | journal | CreateStandardJournalEntries | Yes | 30s | Closing entries |
| 10 | general-ledger | PostClosingEntries | Yes | 60s | Post to GL |
| 11 | financial-reports | GenerateFinancialStatements | Yes | 120s | P&L, Balance Sheet, CF |
| 12 | general-ledger | UnlockPeriodForReporting | Yes | 10s | Allow period viewing |

**Input Requirements:**

```go
{
  "tenantID": "tenant-123",
  "companyID": "company-456",
  "branchID": "branch-789",
  "input": {
    "fiscal_month": 2,           // February
    "fiscal_year": 2026,
    "close_date": "2026-02-28",
    "initiated_by": "controller@company.com"
  }
}
```

**Example Workflow:**

```
SAGA-F01 Month-End Close Execution

Input: Month = Feb 2026, Close Date = 2026-02-28

1. Lock Feb 2026 accounting period
   → No more postings allowed after this point

2. Validate GL balances
   → Check GL accounts have valid balances
   → Verify no orphaned transactions

3. Process accruals (e.g., Accrued Expenses)
   → Accrue utility bills not yet invoiced
   → Accrue salary for days worked
   → Result: 15 accrual entries created

4. Reconcile IC transactions
   → Match SI-2026-001 with PO-2026-001
   → Verify pricing consistency
   → Result: 8 IC transactions matched

5. Calculate provisions
   → Tax provision: $50,000
   → Legal provision: $10,000
   → Warranty provision: $5,000

6. Process tax adjustments
   → Timing differences
   → Rate changes
   → Result: $15,000 additional tax

7. Consolidate financials
   → Eliminate IC transactions
   → Consolidate subsidiary results
   → Result: Consolidated GL ready

8. Generate trial balance
   → Assets: $10M
   → Liabilities: $5M
   → Equity: $5M
   → Verification: BALANCED ✓

9. Create closing entries
   → Close revenue accounts
   → Close expense accounts
   → Create closing journal entries

10. Post closing entries
    → Update retained earnings
    → Zero out temporary accounts

11. Generate financial statements
    → P&L: Net Income = $500k
    → Balance Sheet: Total Assets = $10M
    → Cash Flow Statement

12. Unlock period for reporting
    → Month closed successfully
    → Ready for external reporting

Completion Time: ~8-10 minutes
Error Handling: If any step fails, escalate to controller (no automatic retry)
```

**Special Handling:**

- If any step fails, entire saga fails (no partial closes)
- Orchestrator notifies CFO/Controller immediately on failure
- No automatic retry; manual intervention required
- Monthly close can only be attempted once per period
- Locks prevent concurrent close attempts

**Metrics & Monitoring:**

```
Success Rate Target: 99.99%
Max Duration: 15 minutes
Expected Duration: 8 minutes
Timeout Escalation: 5 minutes after timeout
Monitoring: Real-time dashboard showing close progress
```

---

### Finance: SAGA-F02 - Bank Reconciliation

**Business Purpose:** Reconcile bank statements with cash GL account balances.

**Steps:** 11 forward + 10 compensation = 21 total

**Critical Steps:** 1, 3, 5, 7, 9, 10

**Workflow:**
1. Upload and validate bank statement
2. Retrieve outstanding checks and deposits
3. Match cleared transactions to GL
4. Identify discrepancies
5. Create reconciliation adjustments
6. Process NSF checks (if any)
7. Post bank reconciliation entries
8. Update cash balance
9. Create reconciliation report
10. Archive bank statement
11. Notify accounting

**Key Services:**
- Bank Integration Service
- General Ledger
- Cash Management
- Reconciliation Engine
- Notification Service

**Example Reconciliation:**

```
Bank Statement: ABC Bank, Acct 1234567
Statement Period: Feb 1-28, 2026
Bank Balance: $150,000

GL Cash Balance: $135,000

Differences:
  Outstanding Checks: ($15,000)
  Deposits in Transit: $30,000
  Bank Charges: ($100)
  NSF Check: ($500)
  Interest: $600

Reconciliation:
  GL Balance:          $135,000
  + Deposits in Transit:  $30,000
  - Outstanding Checks:  ($15,000)
  - Bank Charges:         ($100)
  - NSF Check:            ($500)
  + Interest:             $600
  = Reconciled:        $150,000 ✓

Compensation Steps (if adjustment fails):
  - Reverse reconciliation entries
  - Reset cash balance
  - Restore statement for reprocessing
```

---

### Finance: SAGA-F03 - Multi-Currency Revaluation

**Business Purpose:** Revalue foreign currency transactions at period end exchange rates.

**Steps:** 8 forward + 7 compensation = 15 total

**Critical Steps:** 2, 4, 6

**Workflow:**
1. Retrieve exchange rates for period end
2. Identify foreign currency transactions
3. Calculate revaluation adjustments
4. Create revaluation journal entries
5. Post to GL
6. Calculate FX gains/losses
7. Update subsidiary ledgers
8. Generate revaluation report

**Key Services:**
- Currency/Exchange Rate Service
- General Ledger
- AR / AP Services
- Journal Engine
- Tax Engine

**Example:**

```
Date: Feb 28, 2026
Currency: USD

Outstanding AR: INR 1,000,000 (Invoice Rate: 1 USD = 82 INR)
Current Rate: 1 USD = 84 INR

Revaluation Adjustment:
  Invoice USD Amount: $12,195 (1,000,000 / 82)
  Current USD Amount: $11,905 (1,000,000 / 84)
  FX Loss: $290

GL Entry:
  Dr. FX Loss Account      $290
  Cr. AR - Currency Gain      $290
```

---

### Finance: SAGA-F04 - Intercompany Transaction

**Business Purpose:** Process and settle intercompany transactions between company units.

**Steps:** 7 forward + 6 compensation = 13 total

**Critical Steps:** 2, 4, 6

**Workflow:**
1. Retrieve IC transaction details
2. Validate IC transaction pricing
3. Create IC sale/purchase invoices
4. Post to GL in both companies
5. Create IC receivable/payable
6. Generate IC reconciliation
7. Settle IC account

**Key Services:**
- Intercompany Service
- General Ledger
- AR / AP Services
- Consolidation Engine
- Compliance Checker

---

### HR: SAGA-H01 - Payroll Processing

**Business Purpose:** Calculate and process monthly payroll for all employees.

**Steps:** 10 forward + 9 compensation = 19 total

**Critical Steps:** 1, 3, 5, 7, 9

**Workflow:**
1. Lock payroll period (prevent timesheet changes)
2. Retrieve employee data and salary structures
3. Calculate gross salary
4. Calculate deductions (tax, insurance, etc.)
5. Calculate overtime and bonuses
6. Process garnishments and adjustments
7. Generate payroll register
8. Create salary payment batches
9. Post to GL
10. Send salary slips to employees

**Key Services:**
- HR / Employee Master
- Payroll Engine
- Tax Calculator
- GL Service
- Bank Integration
- Notification Service
- Tax Filing

**Compensation Workflow:**

If Step 9 (GL posting) fails:
- Reverse salary batches (Step 109)
- Reverse GL entries (Step 108)
- Revert tax calculations (Step 107)
- Restore employee balances (Step 106)
- Restore payroll locks (Step 105)

**Example Payroll Run:**

```
Payroll Period: Feb 2026
Employees: 100
Total Payroll: $250,000

Step Execution:
1. Lock Feb 2026 payroll (prevent further changes)
   ✓ Locked for 50 employees

2. Retrieve employee data
   ✓ Basic: $150,000
   ✓ DA/HRA: $30,000
   ✓ Allowances: $15,000

3. Calculate gross salary
   ✓ Total Gross: $195,000

4. Calculate deductions
   ✓ Income Tax: $25,000
   ✓ PF: $15,000
   ✓ Insurance: $5,000
   ✓ Total Deductions: $45,000

5. Add overtime/bonuses
   ✓ Overtime: $5,000
   ✓ Performance Bonus: $10,000

6. Process special adjustments
   ✓ Loan EMI: $500
   ✓ Advance Recovery: $1,000

7. Generate payroll register
   ✓ Register created with all calculations

8. Create payment batches
   ✓ Batch 1: Direct Deposit (70 employees)
   ✓ Batch 2: Check Payment (20 employees)
   ✓ Batch 3: Cash (10 employees)

9. Post to GL
   ✓ Salary Expense: $195,000
   ✓ Payroll Tax Payable: $25,000
   ✓ Bank/Cash: $165,000
   ✓ PF Payable: $15,000

10. Send salary slips
    ✓ Email sent to 85 employees
    ✓ Portal access: 100 employees

Completion: 15 minutes
Status: COMPLETED SUCCESSFULLY
```

---

### HR: SAGA-H03 - Employee Exit

**Business Purpose:** Handle complete employee exit with settlement, benefits closure, and access revocation.

**Steps:** 8 forward + 7 compensation = 15 total

**Critical Steps:** 2, 4, 6, 7

**Workflow:**
1. Validate exit data and documentation
2. Calculate final settlement
3. Process outstanding reimbursements
4. Close benefits and insurance
5. Revoke system access
6. Return company assets
7. Create exit letter and documents
8. Archive employee record

**Key Services:**
- HR / Employee Master
- Payroll Engine
- Benefits Management
- Access Control
- Asset Management
- Document Management
- Compliance Service

---

### Manufacturing: SAGA-M01 - Production Order Release

**Business Purpose:** Release production orders to shop floor with complete documentation.

**Steps:** 6 forward + 5 compensation = 11 total

**Critical Steps:** 1, 3, 5

**Key Services:**
- Production Order Service
- Material Management
- Work Center Management
- Job Card Creation
- Quality Control

---

### Manufacturing: SAGA-M02 - Subcontracting

**Business Purpose:** Manage complete subcontracting workflow from order to receipt and payment.

**Steps:** 8 forward + 7 compensation = 15 total

**Critical Steps:** 1, 3, 5, 7

**Key Services:**
- Subcontracting Service
- Production Planning
- Procurement
- Quality Inspection
- AP / Payment

---

## Module Integration Guide

### Finance Module

The Finance module integrates 8 sagas across accounting, tax, and reporting domains.

**Module Structure:**

```go
packages/saga/sagas/finance/
├── fx.go
├── saga_f05_revenue_recognition.go
├── saga_f06_asset_capitalization.go
├── saga_f07_gst_credit_reversal.go
├── saga_f08_cost_center_allocation.go
├── month_end_close_saga.go           // F01: Special case
├── bank_reconciliation_saga.go        // F02
├── multi_currency_revaluation_saga.go // F03
├── intercompany_transaction_saga.go   // F04
├── finance_sagas_test.go
├── finance_critical_sagas_test.go
└── README.md
```

**FX Module Pattern:**

```go
// packages/saga/sagas/finance/fx.go
package finance

import "go.uber.org/fx"

// FinanceSagasModule provides all finance saga handlers
var FinanceSagasModule = fx.Module(
    "finance_sagas",
    fx.Provide(
        NewRevenueRecognitionSaga,
        NewAssetCapitalizationSaga,
        NewGSTCreditReversalSaga,
        NewCostCenterAllocationSaga,
        NewMonthEndCloseSaga,
        NewBankReconciliationSaga,
        NewMultiCurrencyRevaluationSaga,
        NewIntercompanyTransactionSaga,
    ),
)

// FinanceSagasRegistrationModule registers all finance sagas with the registry
var FinanceSagasRegistrationModule = fx.Module(
    "finance_sagas_registration",
    fx.Invoke(registerFinanceSagas),
)

// registerFinanceSagas registers all finance saga handlers with the registry
func registerFinanceSagas(
    registry *orchestrator.SagaRegistry,
    f05 *saga.SagaHandler,
    f06 *saga.SagaHandler,
    // ... more sagas
) error {
    if err := registry.RegisterHandler("SAGA-F05", f05); err != nil {
        return err
    }
    // ... register remaining sagas
    return nil
}
```

**Service Endpoints (Finance Module):**

| Service | Port | URL |
|---------|------|-----|
| general-ledger | 8100 | http://localhost:8100 |
| accounts-receivable | 8103 | http://localhost:8103 |
| accounts-payable | 8104 | http://localhost:8104 |
| journal | 8083 | http://localhost:8083 |
| transaction | 8084 | http://localhost:8084 |
| billing | 8105 | http://localhost:8105 |
| reconciliation | 8107 | http://localhost:8107 |
| cost-center | 8109 | http://localhost:8109 |
| financial-reports | 8111 | http://localhost:8111 |
| financial-close | 8093 | http://localhost:8093 |
| compliance-postings | 8094 | http://localhost:8094 |
| tax-engine | 8091 | http://localhost:8091 |
| cash-management | 8088 | http://localhost:8088 |

**Adding New Finance Saga:**

1. Create saga file: `saga_fXX_name.go`
2. Implement `SagaHandler` interface with steps and compensation
3. Add constructor to `fx.go` Provide list
4. Register in `registerFinanceSagas()` function
5. Add tests in `finance_sagas_test.go`
6. Update service registry if new services involved

---

### Manufacturing Module

The Manufacturing module integrates 6 sagas across production planning, execution, and quality.

**Module Structure:**

```
packages/saga/sagas/manufacturing/
├── fx.go
├── bom_explosion_mrp_saga.go
├── production_order_saga.go
├── quality_rework_saga.go
├── job_card_consumption_saga.go
├── routing_sequencing_saga.go
├── subcontracting_saga.go
├── manufacturing_sagas_test.go
├── manufacturing_critical_sagas_test.go
└── README.md
```

**Service Endpoints (Manufacturing Module):**

| Service | Port | URL |
|---------|------|-----|
| bom | 8190 | http://localhost:8190 |
| production-order | 8191 | http://localhost:8191 |
| production-planning | 8192 | http://localhost:8192 |
| shop-floor | 8193 | http://localhost:8193 |
| quality-production | 8194 | http://localhost:8194 |
| subcontracting | 8195 | http://localhost:8195 |
| work-center | 8196 | http://localhost:8196 |
| routing | 8197 | http://localhost:8197 |
| job-card | 8198 | http://localhost:8198 |

---

### HR Module

The HR module integrates 6 sagas across recruitment, onboarding, leave, payroll, and exit.

**Module Structure:**

```
packages/saga/sagas/hr/
├── fx.go
├── employee_onboarding_saga.go
├── leave_application_saga.go
├── appraisal_salary_revision_saga.go
├── expense_reimbursement_saga.go
├── payroll_processing_saga.go
├── employee_exit_saga.go
├── hr_sagas_test.go
├── hr_critical_sagas_test.go
└── README.md
```

**Service Endpoints (HR Module):**

| Service | Port | URL |
|---------|------|-----|
| employee | 8113 | http://localhost:8113 |
| leave | 8114 | http://localhost:8114 |
| attendance | 8115 | http://localhost:8115 |
| salary-structure | 8117 | http://localhost:8117 |
| recruitment | 8118 | http://localhost:8118 |
| appraisal | 8173 | http://localhost:8173 |
| expense | 8174 | http://localhost:8174 |
| exit | 8175 | http://localhost:8175 |
| payroll | 8116 | http://localhost:8116 |

---

### Projects Module

The Projects module integrates 4 sagas across project execution, billing, and close.

**Module Structure:**

```
packages/saga/sagas/projects/
├── fx.go
├── project_billing_saga.go
├── progress_billing_saga.go
├── subcontractor_payment_saga.go
├── project_close_saga.go
├── projects_sagas_test.go
└── README.md
```

**Service Endpoints (Projects Module):**

| Service | Port | URL |
|---------|------|-----|
| project | 8160 | http://localhost:8160 |
| task | 8161 | http://localhost:8161 |
| timesheet | 8162 | http://localhost:8162 |
| project-costing | 8163 | http://localhost:8163 |
| boq | 8164 | http://localhost:8164 |
| sub-contractor | 8165 | http://localhost:8165 |
| progress-billing | 8166 | http://localhost:8166 |

---

## Saga Engine Architecture

The saga engine provides production-ready distributed transaction orchestration with fault tolerance, compensation, and observability.

### Core Components

#### 1. SagaOrchestratorImpl

**Purpose:** Coordinates saga execution from start to finish

**Responsibilities:**
- Validate saga type and input
- Create saga execution record
- Execute steps sequentially
- Handle timeouts and retries
- Trigger compensation on failure
- Publish events
- Persist state

**Thread Safety:** RWMutex-protected per saga type

**Key Methods:**
- `ExecuteSaga(ctx, sagaType, input)` - Start new saga
- `ResumeSaga(ctx, sagaID)` - Resume from last step
- `GetExecution(ctx, sagaID)` - Retrieve current state
- `GetExecutionTimeline(ctx, sagaID)` - Get all step history

#### 2. SagaRegistry

**Purpose:** Manages registration and lookup of saga handlers

**Responsibilities:**
- Register handlers for saga types
- Retrieve handler by saga type
- Prevent duplicate registrations
- Support dynamic registration

**Thread Safety:** RWMutex-protected for concurrent access

**Key Methods:**
- `RegisterHandler(sagaType, handler)` - Register handler
- `GetHandler(sagaType)` - Retrieve handler
- `HasHandler(sagaType)` - Check existence
- `GetAllHandlers()` - Get all registered

#### 3. SagaStepExecutor

**Purpose:** Executes individual saga steps via RPC

**Responsibilities:**
- Invoke service handlers
- Handle idempotency
- Apply input mappings (JSONPath)
- Deserialize responses
- Manage timeouts
- Log step execution

**Key Methods:**
- `ExecuteStep(ctx, sagaID, stepNum, stepDef)` - Execute step
- `GetStepStatus(ctx, sagaID, stepNum)` - Get step status

#### 4. SagaTimeoutHandler

**Purpose:** Manages timeouts, retries, and circuit breaker

**Responsibilities:**
- Set up step timeouts
- Calculate exponential backoff
- Detect timeouts
- Manage retry attempts
- Implement circuit breaker pattern

**Retry Configuration:**
```go
type RetryConfiguration struct {
    MaxRetries        int32   // 3
    InitialBackoffMs  int32   // 1000
    MaxBackoffMs      int32   // 30000
    BackoffMultiplier float64 // 2.0
    JitterFraction    float64 // 0.1
}
```

**Exponential Backoff Calculation:**
```
backoff = min(
    initialBackoff * (multiplier ^ attempt),
    maxBackoff
) * (1 + jitter)

Sequence: 1s, 2s, 4s, 8s, 16s, 30s, 30s, ...
With jitter: 0.9-1.1s, 1.8-2.2s, 3.6-4.4s, ...
```

#### 5. SagaEventPublisher

**Purpose:** Publishes saga events to Kafka asynchronously

**Responsibilities:**
- Publish step lifecycle events
- Publish compensation events
- Publish saga completion events
- Support event subscribers

**Event Types:**
- `SAGA.STEP.STARTED`
- `SAGA.STEP.COMPLETED`
- `SAGA.STEP.FAILED`
- `SAGA.STEP.RETRYING`
- `SAGA.COMPENSATION.STARTED`
- `SAGA.COMPENSATION.COMPLETED`
- `SAGA.COMPLETED`
- `SAGA.FAILED`

#### 6. SagaRepository

**Purpose:** Persists saga state to PostgreSQL

**Responsibilities:**
- Create saga execution records
- Update saga status
- Retrieve saga history
- Support saga recovery

#### 7. SagaCompensationEngine

**Purpose:** Executes compensation steps on failure

**Responsibilities:**
- Identify compensation steps
- Execute in reverse order
- Log compensation actions
- Maintain audit trail

### Saga Execution Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Client calls ExecuteSaga("SAGA-S01", input)                  │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. Orchestrator Validation                                      │
│  ├─ Check saga type registered                                  │
│  ├─ Validate input against schema                               │
│  └─ Create SagaExecution record with status=RUNNING             │
└────────────────────┬────────────────────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
        ▼                         ▼
   ┌─────────────────┐    ┌──────────────────┐
   │ For each step   │    │ Publish Event    │
   │ 1 to N:         │    │ STEP.STARTED     │
   └────────┬────────┘    └──────────────────┘
            │
            ▼
   ┌─────────────────────────────────────────┐
   │ 3. Setup Step Timeout                   │
   │    Set deadline based on stepTimeout    │
   └────────┬────────────────────────────────┘
            │
            ▼
   ┌─────────────────────────────────────────┐
   │ 4. Execute Step (with retry loop)       │
   │                                         │
   │ for attempt = 0 to maxRetries:          │
   │   ├─ Invoke service handler             │
   │   │  (via RpcConnector)                 │
   │   │                                     │
   │   ├─ If success: BREAK                  │
   │   │                                     │
   │   ├─ If retryable error:                │
   │   │  ├─ Calculate backoff               │
   │   │  ├─ Sleep for backoff duration      │
   │   │  └─ Publish STEP.RETRYING event     │
   │   │                                     │
   │   └─ If non-retryable error:            │
   │      └─ GOTO Compensation (Step 6)      │
   └────────┬────────────────────────────────┘
            │
            ▼
   ┌─────────────────────────────────────────┐
   │ 5. Log Step Result & Update State       │
   │  ├─ Record result in ExecutionLog       │
   │  ├─ Update SagaExecution with result    │
   │  ├─ Publish STEP.COMPLETED event        │
   │  └─ Cancel step timeout                 │
   └────────┬────────────────────────────────┘
            │
        ┌───┴──────────────────────────┐
        │                              │
   ┌────▼─────────────────┐  ┌────────▼─────────────────┐
   │ More steps?          │  │ All steps completed?     │
   │ YES: loop to Step 1  │  │ YES: GOTO Success (7)    │
   └──────────────────────┘  └──────────────────────────┘
                                     │
                                ┌────┴──────────────┐
                                │                   │
        ┌───────────────────────▼──────────────────▼──────────────┐
        │ 6. Compensation: ON FAILURE                              │
        │                                                          │
        │  ├─ Set SagaExecution status = COMPENSATING             │
        │  ├─ Publish COMPENSATION.STARTED event                  │
        │  │                                                      │
        │  ├─ For each compensation step (in REVERSE order):      │
        │  │   ├─ Look up compensation step definition            │
        │  │   ├─ Execute compensation handler                    │
        │  │   ├─ Log compensation action                         │
        │  │   └─ Continue even if compensation fails             │
        │  │       (best effort - don't compound failure)         │
        │  │                                                      │
        │  ├─ Set SagaExecution status = COMPENSATED              │
        │  ├─ Publish COMPENSATION.COMPLETED event                │
        │  └─ Publish SAGA.FAILED event                           │
        │                                                          │
        └───────────────────┬──────────────────────────────────────┘
                            │
        ┌───────────────────▼──────────────────────────────────────┐
        │ 7. Success: ON COMPLETION                                │
        │                                                          │
        │  ├─ Set SagaExecution status = COMPLETED                │
        │  ├─ Set CompletedAt timestamp                           │
        │  ├─ Publish SAGA.COMPLETED event                        │
        │  └─ Return SagaExecution with results                   │
        │                                                          │
        └───────────────────┬──────────────────────────────────────┘
                            │
                            ▼
                    ┌─────────────────┐
                    │ Saga Complete   │
                    │ Return to Client│
                    └─────────────────┘
```

### Compensation Logic

```
Scenario: SAGA-M03 fails at Step 5

Forward execution:
  Step 1: ✓ Get sales order
  Step 2: ✓ Explode BOM
  Step 3: ✓ Plan material requirements
  Step 4: ✓ Schedule procurement
  Step 5: ✗ Update inventory plan (FAILS)

Compensation execution (reverse order):
  Step 104 (compensates 5): ✓ Revert inventory plan
  Step 103 (compensates 4): ✓ Revert procurement
  Step 102 (compensates 3): ✓ Revert material planning
  Step 101 (compensates 2): ✓ Revert BOM explosion
  (No compensation for Step 1 - read-only)

Final State:
  - All forward steps rolled back
  - System back to initial state
  - SagaExecution marked as COMPENSATED
  - Error details recorded for investigation
```

### Circuit Breaker Strategy

```go
// Circuit breaker states for service failures
type CircuitBreakerStatus int

const (
    CLOSED     = 0  // Service OK, requests flowing
    OPEN       = 1  // Service failing, requests blocked
    HALF_OPEN = 2  // Testing if service recovered
)

// Thresholds
CircuitBreakerThreshold = 5        // Failures before opening
CircuitBreakerResetMs = 60000      // 60 second window before retry

// Behavior:
// 1. Request succeeds -> state: CLOSED
// 2. 5 failures in a row -> state: OPEN (fail fast)
// 3. After 60s -> state: HALF_OPEN (test probe)
// 4. Probe succeeds -> state: CLOSED
// 5. Probe fails -> state: OPEN (try again in 60s)
```

---

## Service Registry

Complete list of saga-enabled services with endpoint URLs.

### Finance Services

| Service | Port | Role | Status |
|---------|------|------|--------|
| general-ledger | 8100 | Master GL posting | Ready |
| accounts-receivable | 8103 | Customer invoicing | Ready |
| accounts-payable | 8104 | Vendor invoicing | Ready |
| journal | 8083 | Journal management | Ready |
| transaction | 8084 | Transaction tracking | Ready |
| billing | 8105 | Billing operations | Ready |
| reconciliation | 8107 | Account reconciliation | Ready |
| cost-center | 8109 | Cost allocation | Ready |
| financial-reports | 8111 | Report generation | Ready |
| financial-close | 8093 | Period close | Ready |
| compliance-postings | 8094 | Compliance rules | Ready |
| tax-engine | 8091 | Tax calculation | Ready |
| cash-management | 8088 | Cash forecasting | Ready |
| depreciation | 8169 | Asset depreciation | Ready |

### Manufacturing Services

| Service | Port | Role | Status |
|---------|------|------|--------|
| bom | 8190 | BOM management | Ready |
| production-order | 8191 | Production orders | Ready |
| production-planning | 8192 | MRP planning | Ready |
| shop-floor | 8193 | Shop floor control | Ready |
| quality-production | 8194 | QC inspection | Ready |
| subcontracting | 8195 | Subcontractor orders | Ready |
| work-center | 8196 | Work center scheduling | Ready |
| routing | 8197 | Routing definition | Ready |
| job-card | 8198 | Job cards | Ready |

### HR Services

| Service | Port | Role | Status |
|---------|------|------|--------|
| employee | 8113 | Employee master | Ready |
| leave | 8114 | Leave management | Ready |
| attendance | 8115 | Attendance tracking | Ready |
| salary-structure | 8117 | Salary configuration | Ready |
| recruitment | 8118 | Recruitment management | Ready |
| appraisal | 8173 | Performance appraisal | Ready |
| expense | 8174 | Expense management | Ready |
| exit | 8175 | Employee exit | Ready |
| payroll | 8116 | Payroll processing | Ready |

### Projects Services

| Service | Port | Role | Status |
|---------|------|------|--------|
| project | 8160 | Project management | Ready |
| task | 8161 | Task management | Ready |
| timesheet | 8162 | Timesheet tracking | Ready |
| project-costing | 8163 | Project costing | Ready |
| boq | 8164 | Bill of quantities | Ready |
| sub-contractor | 8165 | Subcontractor mgmt | Ready |
| progress-billing | 8166 | Progress billing | Ready |

### Support Services

| Service | Port | Role | Status |
|---------|------|------|--------|
| audit | 7007 | Audit logging | Ready |
| notification | 7005 | Notifications | Ready |
| approval | 6008 | Approval workflow | Ready |
| user | 6003 | User management | Ready |
| access | 6002 | Access control | Ready |
| asset | 8167 | Asset management | Ready |

### Inventory & Sales Services

| Service | Port | Role | Status |
|---------|------|------|--------|
| inventory-core | 8179 | Inventory management | Ready |
| wms | 8180 | Warehouse mgmt | Ready |
| stock-transfer | 8181 | Stock transfers | Ready |
| qc | 8182 | Quality control | Ready |
| lot-serial | 8183 | Lot/serial tracking | Ready |
| cycle-count | 8184 | Cycle counting | Ready |
| barcode | 8185 | Barcode mgmt | Ready |
| planning | 8186 | Demand planning | Ready |
| fulfillment | 8187 | Order fulfillment | Ready |
| shipping | 8188 | Shipping mgmt | Ready |
| sales-order | 8119 | Sales orders | Ready |
| sales-invoice | 8120 | Sales invoices | Ready |
| crm | 8121 | CRM management | Ready |
| territory | 8122 | Territory mgmt | Ready |
| commission | 8123 | Commission calc | Ready |
| pricing | 8124 | Pricing engine | Ready |
| dealer | 8125 | Dealer management | Ready |
| sales-analytics | 8126 | Sales analytics | Ready |
| route-planning | 8127 | Route planning | Ready |
| field-sales | 8128 | Field sales mgmt | Ready |

### Adding a New Service

1. **Port Allocation:** Choose port in module range
2. **Service Registry:** Add entry to `sagas/registry.go`
3. **Endpoint Registration:** Update `GetServiceEndpoint()` function
4. **Saga References:** Update step definitions if service used
5. **Test Registration:** Update test service registry

**Example:**

```go
// In packages/saga/sagas/registry.go

var ServiceRegistry = map[string]string{
    // ... existing services ...

    // New service
    "my-new-service": "http://localhost:8199",
}

// Test it
func TestNewServiceRegistry(t *testing.T) {
    endpoint := GetServiceEndpoint("my-new-service")
    assert.Equal(t, "http://localhost:8199", endpoint)
}
```

---

## Implementation Patterns

### Standard Saga Structure

All saga handlers follow this pattern:

```go
package mymodule

import "p9e.in/samavaya/packages/saga"

// MySaga implements SAGA-MOD-## business workflow
// Business Flow: Step 1 -> Step 2 -> ... -> Step N
type MySaga struct {
    steps []*saga.StepDefinition
}

// NewMySaga creates new saga handler
func NewMySaga() saga.SagaHandler {
    return &MySaga{
        steps: []*saga.StepDefinition{
            // Step definitions here
        },
    }
}

// SagaType returns saga identifier
func (s *MySaga) SagaType() string {
    return "SAGA-MOD-##"
}

// GetStepDefinitions returns all steps
func (s *MySaga) GetStepDefinitions() []*saga.StepDefinition {
    return s.steps
}

// GetStepDefinition returns specific step
func (s *MySaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
    for _, step := range s.steps {
        if step.StepNumber == int32(stepNum) {
            return step
        }
    }
    return nil
}

// ValidateInput validates saga input
func (s *MySaga) ValidateInput(input interface{}) error {
    inputMap, ok := input.(map[string]interface{})
    if !ok {
        return errors.New("invalid input type")
    }

    // Validate required fields
    if inputMap["required_field"] == nil {
        return errors.New("required_field is required")
    }

    return nil
}
```

### Step Definition Template

```go
// Forward Step (numbered 1-99)
{
    StepNumber:    1,
    ServiceName:   "service-name",
    HandlerMethod: "MethodName",
    InputMapping: map[string]string{
        "inputField": "$.input.field_name",
        "prevResult": "$.steps.1.result.field_name",
    },
    TimeoutSeconds: 15,
    IsCritical:     true,
    CompensationSteps: []int32{101},  // Compensation step numbers
    RetryConfig: &saga.RetryConfiguration{
        MaxRetries:        3,
        InitialBackoffMs:  1000,
        MaxBackoffMs:      30000,
        BackoffMultiplier: 2.0,
        JitterFraction:    0.1,
    },
}

// Compensation Step (numbered 101+)
{
    StepNumber:    101,
    ServiceName:   "service-name",
    HandlerMethod: "RevertMethodName",
    InputMapping: map[string]string{
        "itemID": "$.steps.1.result.item_id",
    },
    TimeoutSeconds: 15,
    IsCritical:     false,  // Compensation failures don't fail saga
    CompensationSteps: []int32{},  // No compensation for compensation
    RetryConfig: &saga.RetryConfiguration{
        MaxRetries:        2,  // Fewer retries for compensation
        InitialBackoffMs:  500,
        MaxBackoffMs:      10000,
        BackoffMultiplier: 2.0,
        JitterFraction:    0.1,
    },
}
```

### Input Validation Template

```go
func (s *MySaga) ValidateInput(input interface{}) error {
    inputMap, ok := input.(map[string]interface{})
    if !ok {
        return fmt.Errorf("invalid input type: expected map, got %T", input)
    }

    // Required field validation
    requiredFields := []string{"field1", "field2", "field3"}
    for _, field := range requiredFields {
        if inputMap[field] == nil {
            return fmt.Errorf("%s is required", field)
        }
    }

    // Type validation
    if id, ok := inputMap["id"].(string); !ok || id == "" {
        return errors.New("id must be non-empty string")
    }

    // Value range validation
    if amount, ok := inputMap["amount"].(float64); !ok || amount <= 0 {
        return errors.New("amount must be positive number")
    }

    // Date validation
    if dateStr, ok := inputMap["date"].(string); !ok {
        if _, err := time.Parse("2006-01-02", dateStr); err != nil {
            return fmt.Errorf("invalid date format: %v", err)
        }
    }

    return nil
}
```

### Critical vs Non-Critical Steps

```go
// Critical Step: Saga fails if this step fails
{
    StepNumber: 1,
    IsCritical: true,  // Must succeed
}

// Non-Critical Step: Saga continues even if step fails
{
    StepNumber: 6,
    IsCritical: false,  // Can fail without saga failing
}

// Decision Logic:
//   - Critical: Essential for transaction correctness
//     Examples: GL posting, inventory deduction, payment
//   - Non-Critical: Nice-to-have, doesn't affect core flow
//     Examples: notification, audit logging, report generation
```

### JSONPath Input Mapping

Input mapping uses JSONPath expressions to extract and transform data:

```go
InputMapping: map[string]string{
    // From saga input
    "tenantID": "$.tenantID",
    "companyID": "$.companyID",

    // From nested input
    "customerID": "$.input.customer_id",
    "invoiceID": "$.input.invoice_id",

    // From previous step result
    "orderID": "$.steps.1.result.order_id",
    "orderLines": "$.steps.1.result.order_lines",

    // From multiple step results
    "totalAmount": "$.steps.5.result.total_amount",
    "taxAmount": "$.steps.6.result.tax_amount",
}
```

**Common JSONPath Patterns:**

| Pattern | Description | Example |
|---------|-------------|---------|
| `$.field` | Root field | `$.tenantID` → "tenant-123" |
| `$.input.field` | Nested input | `$.input.customer_id` → "CUST-001" |
| `$.steps.N.result.field` | Step N result | `$.steps.2.result.qty` → 100 |
| `$.steps.N.error` | Step N error | Error message if step failed |

### Timeout Configuration

```go
// Short timeout (fast operations)
TimeoutSeconds: 10  // Lookup, read-only operations

// Medium timeout (standard operations)
TimeoutSeconds: 20  // Create records, post to GL

// Long timeout (complex operations)
TimeoutSeconds: 60  // Consolidation, report generation

// Guidelines:
// 1-5s: Micro-service calls, simple lookups
// 10-20s: Database operations, standard posting
// 30-60s: Complex calculations, consolidations
// 120s: Report generation, consolidations
```

### Retry Strategy

```go
// Default retry configuration (suitable for most steps)
RetryConfig: &saga.RetryConfiguration{
    MaxRetries:        3,              // Retry up to 3 times
    InitialBackoffMs:  1000,           // Start with 1 second
    MaxBackoffMs:      30000,          // Cap at 30 seconds
    BackoffMultiplier: 2.0,            // Double each time
    JitterFraction:    0.1,            // 10% randomness
}

// Sequence: 1s, 2s, 4s, 8s, 16s, 30s, 30s, ...
// With jitter: 0.9-1.1s, 1.8-2.2s, 3.6-4.4s, ...

// Aggressive retry (for transient failures)
RetryConfig: &saga.RetryConfiguration{
    MaxRetries:        5,
    InitialBackoffMs:  500,
    MaxBackoffMs:      20000,
    BackoffMultiplier: 2.0,
    JitterFraction:    0.1,
}

// Conservative retry (for critical operations)
RetryConfig: &saga.RetryConfiguration{
    MaxRetries:        2,
    InitialBackoffMs:  2000,
    MaxBackoffMs:      10000,
    BackoffMultiplier: 1.5,
    JitterFraction:    0.05,
}
```

---

## Testing Strategy

### Test File Organization

```go
// packages/saga/sagas/mymodule/my_saga_test.go
package mymodule

import (
    "testing"
    "p9e.in/samavaya/packages/saga"
)

// Test categories
func TestMySaga_Type(t *testing.T) { }           // Metadata
func TestMySaga_Steps(t *testing.T) { }          // Step definition
func TestMySaga_StepCount(t *testing.T) { }      // Step validation
func TestMySaga_InputValidation(t *testing.T) { } // Input checks
func TestMySaga_CriticalSteps(t *testing.T) { }  // Critical path
func TestMySaga_Compensation(t *testing.T) { }   // Compensation setup
func TestMySaga_Timeouts(t *testing.T) { }       // Timeout config
func TestMySaga_RetryConfig(t *testing.T) { }    // Retry strategy

// Integration tests
func TestMySagaExecution_HappyPath(t *testing.T) { }      // Success flow
func TestMySagaExecution_FailureFlow(t *testing.T) { }    // Failure flow
func TestMySagaExecution_Compensation(t *testing.T) { }   // Rollback

// packages/saga/sagas/mymodule/my_saga_critical_test.go
func TestMySaga_CriticalPath(t *testing.T) { }        // Critical steps
func TestMySaga_CriticalFailureHandling(t *testing.T) { } // Critical failures
```

### Test Patterns

**Metadata Test:**

```go
func TestMyServiceSaga_Type(t *testing.T) {
    saga := NewMyServiceSaga()

    assert.Equal(t, "SAGA-MOD-##", saga.SagaType())
}
```

**Step Definition Test:**

```go
func TestMyServiceSaga_StepDefinitions(t *testing.T) {
    saga := NewMyServiceSaga()
    steps := saga.GetStepDefinitions()

    // Validate step count
    assert.Equal(t, 9, len(steps))  // 6 forward + 3 compensation

    // Validate specific steps
    assert.Equal(t, int32(1), steps[0].StepNumber)
    assert.Equal(t, "service-name", steps[0].ServiceName)
    assert.True(t, steps[0].IsCritical)

    // Validate compensation steps
    compensationStep := saga.GetStepDefinition(101)
    assert.NotNil(t, compensationStep)
    assert.Equal(t, int32(101), compensationStep.StepNumber)
}
```

**Input Validation Test:**

```go
func TestMyServiceSaga_InputValidation(t *testing.T) {
    saga := NewMyServiceSaga()

    tests := []struct {
        name    string
        input   interface{}
        wantErr bool
    }{
        {
            name: "valid input",
            input: map[string]interface{}{
                "field1": "value1",
                "field2": 100,
            },
            wantErr: false,
        },
        {
            name:    "invalid type",
            input:   "not a map",
            wantErr: true,
        },
        {
            name: "missing required field",
            input: map[string]interface{}{
                "field1": "value1",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := saga.ValidateInput(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

**Critical Steps Test:**

```go
func TestMyServiceSaga_CriticalSteps(t *testing.T) {
    saga := NewMyServiceSaga()
    steps := saga.GetStepDefinitions()

    // Only forward steps (1-99) can be critical
    for _, step := range steps {
        if step.StepNumber >= 100 {  // Compensation step
            assert.False(t, step.IsCritical,
                "Compensation steps should not be critical")
        }
    }

    // Validate critical steps
    criticalSteps := []int32{1, 2, 3, 4}  // Expected critical
    for _, stepNum := range criticalSteps {
        step := saga.GetStepDefinition(int(stepNum))
        assert.NotNil(t, step)
        assert.True(t, step.IsCritical)
    }
}
```

### Coverage Requirements

| Category | Target | Comments |
|----------|--------|----------|
| Input Validation | 100% | All paths exercised |
| Happy Path | 100% | Success flow complete |
| Error Paths | 95%+ | All error conditions |
| Compensation | 90%+ | Rollback scenarios |
| Timeouts | 85%+ | Timeout handling |
| Retries | 85%+ | Retry logic |
| **Overall** | **90%+** | Production ready |

---

## FX Module Integration

### Module Registration Pattern

All saga modules follow FX integration pattern:

```go
// packages/saga/sagas/mymodule/fx.go
package mymodule

import (
    "go.uber.org/fx"
    "p9e.in/samavaya/packages/saga"
    "p9e.in/samavaya/packages/saga/orchestrator"
)

// MyModuleSagasModule provides all saga handlers for the module
var MyModuleSagasModule = fx.Module(
    "mymodule_sagas",

    // Provide all saga handlers
    fx.Provide(
        NewMySaga1,
        NewMySaga2,
        NewMySaga3,
    ),
)

// MyModuleSagasRegistrationModule registers all sagas with registry
var MyModuleSagasRegistrationModule = fx.Module(
    "mymodule_sagas_registration",

    // Invoke registration function
    fx.Invoke(registerMyModuleSagas),
)

// registerMyModuleSagas registers all sagas with the registry
func registerMyModuleSagas(
    registry *orchestrator.SagaRegistry,
    saga1 saga.SagaHandler,
    saga2 saga.SagaHandler,
    saga3 saga.SagaHandler,
) error {
    if err := registry.RegisterHandler("SAGA-MOD-01", saga1); err != nil {
        return err
    }
    if err := registry.RegisterHandler("SAGA-MOD-02", saga2); err != nil {
        return err
    }
    if err := registry.RegisterHandler("SAGA-MOD-03", saga3); err != nil {
        return err
    }
    return nil
}
```

### Main FX Module Integration

In `packages/saga/fx.go`, import and use module:

```go
import (
    "p9e.in/samavaya/packages/saga/sagas/mymodule"
)

var SagaEngineModule = fx.Module(
    "saga_engine",

    // ... other components ...

    // MyModule Saga Handlers (Phase 4)
    mymodule.MyModuleSagasModule,
    mymodule.MyModuleSagasRegistrationModule,

    // ... rest of module ...
)
```

### Dependency Injection Pattern

Sagas use FX for dependency injection:

```go
// Constructor with FX pattern
func NewMyServiceSaga(
    logger p9log.Logger,
    config *saga.DefaultConfig,
) saga.SagaHandler {
    return &MyServiceSaga{
        steps: defineSteps(),
        // Initialize with dependencies if needed
    }
}

// Note: Most sagas are stateless and just define steps
// Dependencies injected at executor level for step execution
```

---

## Step-by-Step Implementation Guide

### For Developers Implementing New Sagas

#### Phase 1: Plan the Saga

1. **Identify Business Flow**
   - Document process steps
   - Identify critical vs optional steps
   - Map to services

2. **Define Step Sequence**
   - Number each forward step (1-99)
   - Identify compensation steps (101-199)
   - Determine dependencies

3. **Estimate Timeouts**
   - Fast steps: 10-15s
   - Standard: 20-30s
   - Complex: 60s+

4. **Plan Compensation**
   - For each critical step, define rollback
   - Use numbered mapping (step 2 → 102, etc.)

#### Phase 2: Implement the Saga

**File: `packages/saga/sagas/mymodule/my_saga.go`**

```go
package mymodule

import (
    "errors"
    "p9e.in/samavaya/packages/saga"
)

// MySaga implements SAGA-MOD-## workflow
type MySaga struct {
    steps []*saga.StepDefinition
}

// NewMySaga creates saga
func NewMySaga() saga.SagaHandler {
    return &MySaga{
        steps: []*saga.StepDefinition{
            // Step 1: ...
            // Step 2: ...
            // ... (continue)
            // Step 101: Compensation for Step 1
            // ... (etc)
        },
    }
}

// SagaType returns SAGA-MOD-##
func (s *MySaga) SagaType() string {
    return "SAGA-MOD-##"
}

// GetStepDefinitions returns all steps
func (s *MySaga) GetStepDefinitions() []*saga.StepDefinition {
    return s.steps
}

// GetStepDefinition returns step by number
func (s *MySaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
    for _, step := range s.steps {
        if step.StepNumber == int32(stepNum) {
            return step
        }
    }
    return nil
}

// ValidateInput validates saga input
func (s *MySaga) ValidateInput(input interface{}) error {
    inputMap, ok := input.(map[string]interface{})
    if !ok {
        return errors.New("invalid input type")
    }

    // Validate required fields
    if inputMap["field1"] == nil {
        return errors.New("field1 is required")
    }

    return nil
}
```

#### Phase 3: Add to FX Module

**Update: `packages/saga/sagas/mymodule/fx.go`**

```go
// Add to MyModuleSagasModule
var MyModuleSagasModule = fx.Module(
    "mymodule_sagas",
    fx.Provide(
        // ... existing ...
        NewMySaga,  // ADD THIS
    ),
)

// Update registration function
func registerMyModuleSagas(
    registry *orchestrator.SagaRegistry,
    // ... existing ...
    mySaga saga.SagaHandler,  // ADD THIS
) error {
    // ... existing ...
    if err := registry.RegisterHandler("SAGA-MOD-##", mySaga); err != nil {
        return err
    }
    return nil
}
```

#### Phase 4: Write Tests

**File: `packages/saga/sagas/mymodule/my_saga_test.go`**

```go
package mymodule

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestMySaga_Type(t *testing.T) {
    saga := NewMySaga()
    assert.Equal(t, "SAGA-MOD-##", saga.SagaType())
}

func TestMySaga_Steps(t *testing.T) {
    saga := NewMySaga()
    steps := saga.GetStepDefinitions()

    // Validate counts
    assert.Equal(t, 8, len(steps))  // 5 forward + 3 compensation

    // Validate each step
    step1 := saga.GetStepDefinition(1)
    assert.NotNil(t, step1)
    assert.Equal(t, int32(1), step1.StepNumber)
    assert.Equal(t, "service-name", step1.ServiceName)
}

func TestMySaga_InputValidation(t *testing.T) {
    saga := NewMySaga()

    tests := []struct {
        name    string
        input   interface{}
        wantErr bool
    }{
        {
            name: "valid input",
            input: map[string]interface{}{
                "field1": "value",
            },
            wantErr: false,
        },
        {
            name:    "missing field",
            input:   map[string]interface{}{},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := saga.ValidateInput(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

#### Phase 5: Verify Service Registry

1. Check all services used in saga are in `registry.go`
2. Verify port allocations don't conflict
3. Add new services if needed

#### Phase 6: Document

Add entry to module README.md with:
- Saga ID and name
- Business purpose
- Step count
- Key services
- Example input

### Common Pitfalls

**Mistake 1: Wrong step numbering**
```go
// WRONG: Skipping numbers
steps = []int{1, 2, 5, 6}  // Gap at 3-4

// RIGHT: Contiguous numbering
steps = []int{1, 2, 3, 4}
```

**Mistake 2: Compensation doesn't map correctly**
```go
// WRONG: Step 5 compensates to Step 105, not 5XX
{StepNumber: 5, CompensationSteps: []int32{5}}  // Self-reference!

// RIGHT: Use 100+ numbering
{StepNumber: 5, CompensationSteps: []int32{105}}
```

**Mistake 3: Missing input validation**
```go
// WRONG: No validation
func (s *MySaga) ValidateInput(input interface{}) error {
    return nil  // Always passes
}

// RIGHT: Comprehensive validation
func (s *MySaga) ValidateInput(input interface{}) error {
    inputMap, ok := input.(map[string]interface{})
    if !ok {
        return errors.New("invalid type")
    }
    if inputMap["required"] == nil {
        return errors.New("required field missing")
    }
    return nil
}
```

**Mistake 4: Incorrect JSONPath**
```go
// WRONG: Invalid JSONPath
InputMapping: map[string]string{
    "field": "steps[1].result.field",  // Square brackets
}

// RIGHT: Use dot notation
InputMapping: map[string]string{
    "field": "$.steps.1.result.field",
}
```

**Mistake 5: Non-critical step that should be critical**
```go
// WRONG: GL posting non-critical
{
    StepNumber: 5,
    IsCritical: false,
    // GL posting failed but saga continues!
}

// RIGHT: GL posting is critical
{
    StepNumber: 5,
    IsCritical: true,
    // Saga fails if GL posting fails
}
```

### Best Practices

1. **Step Naming:** Use clear, verb-noun pairs
   - "Create Customer Invoice"
   - "Post to General Ledger"
   - "Revert Inventory Allocation"

2. **Timeout Selection:**
   - Fast: 10-15 seconds (lookups, validation)
   - Standard: 20-30 seconds (creates, updates)
   - Slow: 60+ seconds (consolidation, reports)

3. **Critical Step Selection:**
   - Any step that updates financial records
   - Any step that commits inventory
   - Any step that posts to GL

4. **Compensation Planning:**
   - Every critical step needs compensation
   - Compensation should be atomic
   - Compensation failures don't fail saga (best effort)

5. **Input Validation:**
   - Validate all required fields
   - Check field types
   - Validate value ranges for business fields

6. **Testing:**
   - Test each step individually in unit tests
   - Test happy path in integration tests
   - Test failure and compensation paths
   - Aim for 90%+ coverage

---

## Deployment & Operations

### Integration with Orchestrator

```go
// In your service's main.go
import "p9e.in/samavaya/packages/saga"

func main() {
    var app *fx.App

    // Create FX app with saga engine
    app = fx.New(
        // ... other modules ...
        saga.SagaEngineModule,  // Provides orchestrator

        // Your service handler
        fx.Invoke(setupSagaHandler),
    )

    if err := app.Start(context.Background()); err != nil {
        log.Fatal(err)
    }

    defer app.Stop(context.Background())
}

// Your saga handler
type OrderHandler struct {
    orchestrator saga.SagaOrchestrator
}

func setupSagaHandler(
    handler *OrderHandler,
    orchestrator saga.SagaOrchestrator,
) {
    handler.orchestrator = orchestrator
}

// Start saga from HTTP handler
func (h *OrderHandler) ProcessOrder(ctx context.Context, req *OrderRequest) error {
    execution, err := h.orchestrator.ExecuteSaga(
        ctx,
        "SAGA-S01",
        &saga.SagaExecutionInput{
            TenantID:  req.TenantID,
            CompanyID: req.CompanyID,
            BranchID:  req.BranchID,
            Input: map[string]interface{}{
                "customer_id": req.CustomerID,
                "order_items": req.Items,
            },
        },
    )

    if err != nil {
        log.Error("Saga failed", err)
        return err
    }

    log.Infof("Saga completed: %s", execution.SagaID)
    return nil
}
```

### Event Publishing to Kafka

```go
// Events automatically published to Kafka
// Topic: saga-events (configurable)
// Partition: Based on sagaID for ordering

// Event examples:
{
    "event_type": "SAGA.STEP.STARTED",
    "saga_id": "SAGA-123",
    "saga_type": "SAGA-S01",
    "step_number": 1,
    "timestamp": "2026-02-15T10:30:00Z",
}

{
    "event_type": "SAGA.STEP.COMPLETED",
    "saga_id": "SAGA-123",
    "saga_type": "SAGA-S01",
    "step_number": 1,
    "result": {"customer_id": "CUST-001", ...},
    "duration_ms": 250,
    "timestamp": "2026-02-15T10:30:01Z",
}

{
    "event_type": "SAGA.COMPENSATION.STARTED",
    "saga_id": "SAGA-123",
    "saga_type": "SAGA-S01",
    "failed_step": 5,
    "error": "Service timeout",
    "timestamp": "2026-02-15T10:30:05Z",
}

{
    "event_type": "SAGA.SAGA.FAILED",
    "saga_id": "SAGA-123",
    "saga_type": "SAGA-S01",
    "error": "Step 5 failed after 3 retries",
    "compensation_status": "COMPLETED",
    "timestamp": "2026-02-15T10:30:10Z",
}
```

### Monitoring & Observability

**Prometheus Metrics:**

```
# Saga execution metrics
saga_execution_total{saga_type="SAGA-S01",status="success"} 1000
saga_execution_total{saga_type="SAGA-S01",status="failed"} 10
saga_execution_duration_seconds{saga_type="SAGA-S01",quantile="p50"} 2.5
saga_execution_duration_seconds{saga_type="SAGA-S01",quantile="p99"} 8.2

# Step execution metrics
saga_step_execution_total{saga_type="SAGA-S01",step_number="1"} 1005
saga_step_execution_total{saga_type="SAGA-S01",step_number="1",status="failed"} 5
saga_step_duration_seconds{saga_type="SAGA-S01",step_number="1",quantile="p50"} 0.25

# Retry metrics
saga_step_retries_total{saga_type="SAGA-S01",step_number="1"} 12
saga_step_retry_success_total{saga_type="SAGA-S01",step_number="1"} 8

# Compensation metrics
saga_compensation_total{saga_type="SAGA-S01"} 8
saga_compensation_success_total{saga_type="SAGA-S01"} 7
```

**Dashboard Queries:**

```sql
-- Saga success rate (last 24h)
SELECT saga_type,
       COUNT(*) as total,
       SUM(CASE WHEN status='success' THEN 1 ELSE 0 END) as succeeded,
       ROUND(100.0 * SUM(CASE WHEN status='success' THEN 1 ELSE 0 END) / COUNT(*), 2) as success_rate
FROM saga_executions
WHERE executed_at >= NOW() - INTERVAL 24 HOUR
GROUP BY saga_type;

-- Slow sagas (p95 execution time)
SELECT saga_type, PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_seconds) as p95_duration
FROM saga_executions
WHERE executed_at >= NOW() - INTERVAL 24 HOUR
GROUP BY saga_type
ORDER BY p95_duration DESC;

-- Failed steps causing sagas to fail
SELECT saga_type, step_number, COUNT(*) as failure_count
FROM saga_execution_logs
WHERE status = 'failed' AND is_critical = true
GROUP BY saga_type, step_number
ORDER BY failure_count DESC
LIMIT 20;
```

### Error Handling & Recovery

**Automatic Recovery:**
1. Timeout handler retries with exponential backoff
2. Circuit breaker prevents cascading failures
3. Compensation reverses partial state

**Manual Recovery:**

```go
// Retrieve failed saga
execution, err := orchestrator.GetExecution(ctx, "SAGA-123")
if err != nil {
    log.Error("Failed to retrieve saga", err)
}

// Check failure details
log.Infof("Saga: %s", execution.SagaID)
log.Infof("Status: %s", execution.Status)
log.Infof("Failed Step: %d", execution.CurrentStepNumber)
log.Infof("Error: %s", execution.ErrorMessage)

// Retrieve step history
timeline, err := orchestrator.GetExecutionTimeline(ctx, "SAGA-123")
for _, step := range timeline {
    log.Infof("Step %d: %s (%s)", step.StepNumber, step.Status, step.ErrorMessage)
}

// Manual retry
resumedExecution, err := orchestrator.ResumeSaga(ctx, "SAGA-123")
if err != nil {
    log.Error("Resume failed", err)
}
```

### Performance Tuning

**Database Optimization:**

```sql
-- Index for saga lookups
CREATE INDEX saga_executions_saga_id ON saga_executions(saga_id);
CREATE INDEX saga_executions_status ON saga_executions(status);
CREATE INDEX saga_executions_created_at ON saga_executions(created_at);

-- Index for execution logs
CREATE INDEX saga_exec_logs_saga_id ON saga_execution_logs(saga_id);
CREATE INDEX saga_exec_logs_step_num ON saga_execution_logs(step_number);
```

**Connection Pool Tuning:**

```go
// Optimize for saga workload
dbConfig := &database.Config{
    MaxOpenConnections:    100,    // Up from 25
    MaxIdleConnections:    20,     // Up from 5
    ConnectionMaxLifetime: 5 * time.Minute,
    ConnectionMaxIdleTime: 2 * time.Minute,
}
```

**Kafka Optimization:**

```go
// Batch saga events
kafkaConfig := &kafka.Config{
    BatchSize: 100,           // Accumulate 100 events
    BatchBytes: 1024 * 1024,  // Or 1MB of data
    Timeout: 5 * time.Second, // Or 5 seconds max wait
    Compression: "snappy",    // Compress batches
}
```

### Troubleshooting Guide

**Issue: Saga stuck in RUNNING state**

```
Possible causes:
1. Service unreachable - Check service endpoints in registry
2. Timeout too short - Increase TimeoutSeconds for steps
3. Deadlock in step execution - Check RPC connector logs
4. Database connection issue - Verify DB connectivity

Resolution:
1. Check service availability: curl http://localhost:8119/health
2. Review step timeout in saga definition
3. Check RPC connector logs for blocked calls
4. Monitor database connection pool
5. If stuck >30 min: Investigate or mark saga as FAILED manually
```

**Issue: High retry rate for specific step**

```
Possible causes:
1. Service too slow for timeout - Increase timeout
2. Downstream service failing - Check downstream service
3. Network issues - Check network connectivity
4. Legitimate transient failures - Accept and continue

Resolution:
1. Check step logs for actual duration
2. Test downstream service independently
3. Review network metrics
4. Consider increasing MaxRetries if transient
```

**Issue: Compensation steps failing**

```
Note: Compensation failures don't fail saga (best effort)

Possible causes:
1. State already changed by concurrent operation
2. Downstream service unavailable
3. Incorrect compensation logic

Resolution:
1. Review saga for race conditions
2. Check downstream service availability
3. Verify compensation step definitions
4. Manual cleanup may be required
```

---

## Glossary & References

### Key Terms

**Saga:** A distributed transaction pattern that coordinates multiple services with eventual consistency and compensation-based rollback.

**Saga Execution:** A single run of a saga, with unique ID (ULID), status tracking, and complete audit trail.

**Step Definition:** Configuration for a single saga step, including service, method, timeout, retry, and compensation mapping.

**Forward Step:** Main business logic step (numbered 1-99), executed when saga runs successfully.

**Compensation Step:** Rollback step (numbered 100+), executed when saga fails to reverse forward step effects.

**Critical Step:** Forward step that must succeed; saga fails if critical step fails. Opposite: Non-critical steps fail silently.

**InputMapping:** JSONPath expressions mapping saga input and previous step results to current step inputs.

**Idempotency:** Property ensuring step can be safely retried without changing result or side effects.

**Circuit Breaker:** Fault tolerance pattern that fast-fails when service fails repeatedly (avoid retry storms).

**Exponential Backoff:** Retry strategy with increasing delays: 1s, 2s, 4s, 8s, ... (with jitter).

**Eventual Consistency:** Model where all replicas eventually converge to same state (vs. immediate ACID).

**RPC Connector:** HTTP client for invoking saga steps via ConnectRPC protocol.

**Kafka Event Publishing:** Asynchronous event stream of saga lifecycle for audit trail and integration.

### Related Documentation

- **Saga Orchestrator:** `packages/saga/orchestrator/README.md`
- **Compensation Engine:** `packages/saga/compensation/README.md` (in development)
- **RPC Connector:** `packages/saga/connector/README.md` (in development)
- **Event Publisher:** `packages/saga/events/README.md` (in development)
- **Saga Models:** `packages/saga/models/README.md`
- **Saga Config:** `packages/saga/config.go`

### Code Examples

**Starting a Saga:**

```go
execution, err := orch.ExecuteSaga(ctx, "SAGA-S01", &saga.SagaExecutionInput{
    TenantID:  "tenant-123",
    CompanyID: "company-456",
    BranchID:  "branch-789",
    Input: map[string]interface{}{
        "customer_id": "CUST-001",
        "order_items": items,
    },
})
if err != nil {
    log.Error("Saga failed", err)
}
```

**Checking Saga Status:**

```go
execution, err := orch.GetExecution(ctx, "SAGA-123")
switch execution.Status {
case "RUNNING":
    log.Infof("Saga running, current step: %d", execution.CurrentStepNumber)
case "COMPLETED":
    log.Infof("Saga completed successfully")
case "FAILED":
    log.Infof("Saga failed: %s", execution.ErrorMessage)
case "COMPENSATED":
    log.Infof("Saga rolled back: %s", execution.ErrorMessage)
}
```

**Handling Saga Events:**

```go
// Kafka consumer subscribing to saga events
messages, err := consumer.Messages()
for msg := range messages {
    var event map[string]interface{}
    json.Unmarshal(msg.Value, &event)

    switch event["event_type"] {
    case "SAGA.STEP.COMPLETED":
        log.Infof("Step %d completed", event["step_number"])
    case "SAGA.COMPENSATION.STARTED":
        log.Warnf("Saga compensating due to failed step %d", event["failed_step"])
    case "SAGA.SAGA.COMPLETED":
        log.Infof("Saga completed successfully")
    }
}
```

### References

1. **Saga Pattern:** Chris Richardson's Microservices Patterns
2. **Eventual Consistency:** Werner Vogels' AWS article
3. **Circuit Breaker:** Release It! by Michael Nygard
4. **ConnectRPC:** https://connectrpc.com/
5. **Kafka:** https://kafka.apache.org/documentation/

---

## Appendix: Phase 4 Implementation Checklist

### Phase 4A (Feb 15-20)

- [ ] SAGA-F05: Revenue Recognition (6 steps)
- [ ] SAGA-F06: Asset Capitalization (5 steps)
- [ ] SAGA-F07: GST Credit Reversal (5 steps)
- [ ] SAGA-F08: Cost Center Allocation (6 steps)
- [ ] SAGA-H05: Leave Application (4 steps)
- [ ] SAGA-M05: Job Card Consumption (5 steps)
- [ ] Finance FX module setup
- [ ] HR FX module setup
- [ ] Manufacturing FX module setup
- [ ] Unit tests (90%+ coverage)
- [ ] Integration tests

### Phase 4B (Feb 20-28)

- [ ] SAGA-M03: BOM Explosion & MRP (8 steps)
- [ ] SAGA-M04: Production Order (6 steps)
- [ ] SAGA-M06: Quality Rework (5 steps)
- [ ] SAGA-H02: Employee Onboarding (10 steps)
- [ ] SAGA-H04: Appraisal & Salary Revision (7 steps)
- [ ] SAGA-H06: Expense Reimbursement (6 steps)
- [ ] SAGA-PR01: Project Billing (7 steps)
- [ ] SAGA-PR02: Progress Billing (8 steps)
- [ ] SAGA-PR03: Subcontractor Payment (6 steps)
- [ ] SAGA-PR04: Project Close (7 steps)
- [ ] Projects FX module setup
- [ ] Unit tests
- [ ] Integration tests
- [ ] Load tests (1,000 concurrent sagas)

### Phase 4C (Feb 28-Mar 15)

- [ ] SAGA-F01: Month-End Close (12 steps, no compensation)
- [ ] SAGA-F02: Bank Reconciliation (11 steps)
- [ ] SAGA-F03: Multi-Currency Revaluation (8 steps)
- [ ] SAGA-F04: Intercompany Transaction (7 steps)
- [ ] SAGA-H01: Payroll Processing (10 steps)
- [ ] SAGA-H03: Employee Exit (8 steps)
- [ ] SAGA-M01: Production Order Release (6 steps)
- [ ] SAGA-M02: Subcontracting (8 steps)
- [ ] Critical test suites
- [ ] Stress tests (10,000+ concurrent sagas)
- [ ] Chaos engineering tests
- [ ] Production deployment

---

**Document Version:** 1.0
**Last Reviewed:** February 15, 2026
**Next Review:** March 1, 2026
**Owner:** Architecture Team
