// Package manufacturing provides saga handlers for manufacturing module workflows
package manufacturing

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// SubcontractingCostTrackingSaga implements SAGA-M10: Subcontracting Cost Tracking workflow
// Business Flow: Create subcontracting work order → Extract unit cost from PO →
// Calculate total subcontracting cost → Allocate cost to production order →
// Match subcontract invoice with order → Calculate variance → Post cost to GL →
// Update vendor records → Track payment status → Archive subcontracting records
//
// Compensation: If any critical step fails, automatically reverses allocations and
// cancels work order to maintain accurate cost tracking and vendor records
type SubcontractingCostTrackingSaga struct {
	steps []*saga.StepDefinition
}

// NewSubcontractingCostTrackingSaga creates a new Subcontracting Cost Tracking saga handler
func NewSubcontractingCostTrackingSaga() saga.SagaHandler {
	return &SubcontractingCostTrackingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Subcontracting Work Order (subcontracting service)
			// Creates subcontracting work order for tracking
			{
				StepNumber:    1,
				ServiceName:   "subcontracting",
				HandlerMethod: "CreateSubcontractingWorkOrder",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"productionOrderID": "$.input.production_order_id",
					"vendorID":         "$.input.vendor_id",
					"productID":        "$.input.product_id",
					"workOrderQuantity": "$.input.work_order_quantity",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},
			// Step 2: Extract Unit Cost from Purchase Order (purchase-invoice service)
			// Extracts unit cost from the PO for subcontracting
			{
				StepNumber:    2,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "ExtractUnitCostFromPO",
				InputMapping: map[string]string{
					"vendorID":     "$.input.vendor_id",
					"productID":    "$.input.product_id",
					"poID":         "$.input.po_id",
					"costingPeriod": "$.input.costing_period",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{110},
			},
			// Step 3: Calculate Total Subcontracting Cost (cost-center service)
			// Calculates total cost by multiplying unit cost and quantity
			{
				StepNumber:    3,
				ServiceName:   "cost-center",
				HandlerMethod: "CalculateTotalSubcontractingCost",
				InputMapping: map[string]string{
					"unitCost":           "$.steps.2.result.unit_cost",
					"workOrderQuantity":  "$.steps.1.result.work_order_quantity",
					"vendorID":           "$.input.vendor_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 4: Allocate Cost to Production Order (production-order service)
			// Allocates total subcontracting cost to the production order
			{
				StepNumber:    4,
				ServiceName:   "production-order",
				HandlerMethod: "AllocateCostToProductionOrder",
				InputMapping: map[string]string{
					"productionOrderID":   "$.input.production_order_id",
					"totalSubcontractCost": "$.steps.3.result.total_subcontracting_cost",
					"workOrderID":        "$.steps.1.result.work_order_id",
					"allocationMethod":   "$.input.allocation_method",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{111},
			},
			// Step 5: Match Subcontract Invoice with Order (purchase-invoice service)
			// Matches received invoice with subcontracting work order
			{
				StepNumber:    5,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "MatchSubcontractInvoiceWithOrder",
				InputMapping: map[string]string{
					"workOrderID":       "$.steps.1.result.work_order_id",
					"invoiceID":         "$.input.invoice_id",
					"invoiceAmount":     "$.input.invoice_amount",
					"vendorID":          "$.input.vendor_id",
				},
				TimeoutSeconds: 60,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{112},
			},
			// Step 6: Calculate Variance (cost-center service)
			// Calculates variance between estimated and invoiced cost
			{
				StepNumber:    6,
				ServiceName:   "cost-center",
				HandlerMethod: "CalculateInvoiceVariance",
				InputMapping: map[string]string{
					"estimatedCost":     "$.steps.3.result.total_subcontracting_cost",
					"invoiceAmount":     "$.input.invoice_amount",
					"workOrderID":      "$.steps.1.result.work_order_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 7: Post Cost to GL (general-ledger service)
			// Posts subcontracting cost to WIP GL account
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostSubcontractingCostToGL",
				InputMapping: map[string]string{
					"totalSubcontractCost": "$.steps.3.result.total_subcontracting_cost",
					"wipAccount":         "$.input.wip_account",
					"costCenterID":       "$.input.cost_center_id",
					"postingDate":        "$.input.posting_date",
					"workOrderID":       "$.steps.1.result.work_order_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{113},
			},
			// Step 8: Update Vendor Records (vendor service)
			// Updates vendor performance and cost tracking records
			{
				StepNumber:    8,
				ServiceName:   "vendor",
				HandlerMethod: "UpdateVendorRecords",
				InputMapping: map[string]string{
					"vendorID":             "$.input.vendor_id",
					"invoiceAmount":        "$.input.invoice_amount",
					"variance":             "$.steps.6.result.variance",
					"invoiceID":           "$.input.invoice_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 9: Track Payment Status (purchase-invoice service)
			// Tracks payment status of subcontracting invoice
			{
				StepNumber:    9,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "TrackPaymentStatus",
				InputMapping: map[string]string{
					"invoiceID":      "$.input.invoice_id",
					"workOrderID":   "$.steps.1.result.work_order_id",
					"paymentTerms":  "$.input.payment_terms",
					"dueDate":       "$.input.due_date",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{114},
			},
			// Step 10: Archive Subcontracting Records (subcontracting service)
			// Archives subcontracting records for historical tracking
			{
				StepNumber:    10,
				ServiceName:   "subcontracting",
				HandlerMethod: "ArchiveSubcontractingRecords",
				InputMapping: map[string]string{
					"workOrderID":         "$.steps.1.result.work_order_id",
					"invoiceID":           "$.input.invoice_id",
					"totalSubcontractCost": "$.steps.3.result.total_subcontracting_cost",
					"archiveDate":         "$.input.archive_date",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},

			// ===== COMPENSATION STEPS =====

			// Compensation Step 110: Restore PO Cost Data (compensates step 2)
			// Restores PO cost data if extraction failed
			{
				StepNumber:    110,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "RestorePOCostData",
				InputMapping: map[string]string{
					"poID":         "$.input.po_id",
					"vendorID":     "$.input.vendor_id",
					"backupID":    "$.steps.2.result.backup_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},
			// Compensation Step 111: Reverse Cost Allocation (compensates step 4)
			// Reverses cost allocation to production order
			{
				StepNumber:    111,
				ServiceName:   "production-order",
				HandlerMethod: "ReverseCostAllocation",
				InputMapping: map[string]string{
					"productionOrderID":    "$.input.production_order_id",
					"totalSubcontractCost": "$.steps.3.result.total_subcontracting_cost",
					"workOrderID":         "$.steps.1.result.work_order_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},
			// Compensation Step 112: Undo Invoice Match (compensates step 5)
			// Undoes the invoice matching with work order
			{
				StepNumber:    112,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "UndoInvoiceMatch",
				InputMapping: map[string]string{
					"workOrderID": "$.steps.1.result.work_order_id",
					"invoiceID":   "$.input.invoice_id",
					"matchID":    "$.steps.5.result.match_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},
			// Compensation Step 113: Reverse GL Posting (compensates step 7)
			// Reverses GL posting made for subcontracting cost
			{
				StepNumber:    113,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseSubcontractingCostGLPosting",
				InputMapping: map[string]string{
					"totalSubcontractCost": "$.steps.3.result.total_subcontracting_cost",
					"wipAccount":         "$.input.wip_account",
					"reversalDate":       "$.input.reversal_date",
					"workOrderID":       "$.steps.1.result.work_order_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},
			// Compensation Step 114: Cancel Payment Tracking (compensates step 9)
			// Cancels payment tracking for subcontracting invoice
			{
				StepNumber:    114,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "CancelPaymentTracking",
				InputMapping: map[string]string{
					"invoiceID":    "$.input.invoice_id",
					"workOrderID": "$.steps.1.result.work_order_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *SubcontractingCostTrackingSaga) SagaType() string {
	return "SAGA-M10"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *SubcontractingCostTrackingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *SubcontractingCostTrackingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: production_order_id, vendor_id, product_id, po_id
func (s *SubcontractingCostTrackingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type: expected map[string]interface{}")
	}

	// Validate production_order_id
	if inputMap["production_order_id"] == nil || inputMap["production_order_id"] == "" {
		return errors.New("production_order_id is required for Subcontracting Cost Tracking saga")
	}

	// Validate vendor_id
	if inputMap["vendor_id"] == nil || inputMap["vendor_id"] == "" {
		return errors.New("vendor_id is required for Subcontracting Cost Tracking saga")
	}

	// Validate product_id
	if inputMap["product_id"] == nil || inputMap["product_id"] == "" {
		return errors.New("product_id is required for Subcontracting Cost Tracking saga")
	}

	// Validate po_id
	if inputMap["po_id"] == nil || inputMap["po_id"] == "" {
		return errors.New("po_id is required for Subcontracting Cost Tracking saga")
	}

	// Validate work_order_quantity
	if inputMap["work_order_quantity"] == nil {
		return errors.New("work_order_quantity is required for Subcontracting Cost Tracking saga")
	}

	return nil
}
