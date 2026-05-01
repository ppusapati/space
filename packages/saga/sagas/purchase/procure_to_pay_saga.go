// Package purchase provides saga handlers for purchase module workflows
package purchase

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ProcureToPaySaga implements SAGA-P01: Procure-to-Pay workflow
// Business Flow: Create PO → Approve PO → Send to Vendor → Receive Goods → Update Stock → Create Invoice → 3-Way Match → Approve Invoice → Post AP → Post GL → Process Payment → Close PO
type ProcureToPaySaga struct {
	steps []*saga.StepDefinition
}

// NewProcureToPaySaga creates a new Procure-to-Pay saga handler
func NewProcureToPaySaga() saga.SagaHandler {
	return &ProcureToPaySaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Purchase Order
			{
				StepNumber:    1,
				ServiceName:   "purchase-order",
				HandlerMethod: "CreatePurchaseOrder",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"vendorID":      "$.input.vendor_id",
					"items":         "$.input.items",
					"deliveryDate":  "$.input.delivery_date",
					"paymentTerms":  "$.input.payment_terms",
				},
				TimeoutSeconds: 25,
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
			// Step 2: Approve Purchase Order
			{
				StepNumber:    2,
				ServiceName:   "purchase-order",
				HandlerMethod: "ApprovePurchaseOrder",
				InputMapping: map[string]string{
					"poID":        "$.steps.1.result.po_id",
					"approverID":  "$.input.approver_id",
					"approvalNote": "$.input.approval_note",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{102},
			},
			// Step 3: Send PO to Vendor
			{
				StepNumber:    3,
				ServiceName:   "purchase-order",
				HandlerMethod: "SendPOToVendor",
				InputMapping: map[string]string{
					"poID":          "$.steps.1.result.po_id",
					"vendorID":      "$.input.vendor_id",
					"sendMethod":    "$.input.send_method",
					"vendorEmail":   "$.input.vendor_email",
				},
				TimeoutSeconds:    20,
				IsCritical:        false, // Non-critical: email failures don't block
				CompensationSteps: []int32{},
			},
			// Step 4: Create Goods Receipt
			{
				StepNumber:    4,
				ServiceName:   "purchase-order",
				HandlerMethod: "CreateGoodsReceipt",
				InputMapping: map[string]string{
					"poID":          "$.steps.1.result.po_id",
					"receiptDate":   "$.input.receipt_date",
					"receivedItems": "$.input.received_items",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{104},
			},
			// Step 5: Receive Stock
			{
				StepNumber:    5,
				ServiceName:   "inventory-core",
				HandlerMethod: "ReceiveStock",
				InputMapping: map[string]string{
					"poID":          "$.steps.1.result.po_id",
					"grnID":         "$.steps.4.result.grn_id",
					"receivedItems": "$.input.received_items",
					"warehouseID":   "$.input.warehouse_id",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{105},
			},
			// Step 6: Create Invoice
			{
				StepNumber:    6,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "CreateInvoice",
				InputMapping: map[string]string{
					"poID":          "$.steps.1.result.po_id",
					"vendorID":      "$.input.vendor_id",
					"invoiceNumber": "$.input.invoice_number",
					"invoiceDate":   "$.input.invoice_date",
					"invoiceAmount": "$.input.invoice_amount",
					"grnID":         "$.steps.4.result.grn_id",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{106},
			},
			// Step 7: Perform Three-Way Match
			{
				StepNumber:    7,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "PerformThreeWayMatch",
				InputMapping: map[string]string{
					"poID":      "$.steps.1.result.po_id",
					"grnID":     "$.steps.4.result.grn_id",
					"invoiceID": "$.steps.6.result.invoice_id",
					"matchTolerance": "$.input.match_tolerance",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{107},
			},
			// Step 8: Approve Invoice
			{
				StepNumber:    8,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "ApproveInvoice",
				InputMapping: map[string]string{
					"invoiceID":     "$.steps.6.result.invoice_id",
					"approverID":    "$.input.approver_id",
					"approvalNotes": "$.input.approval_notes",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{108},
			},
			// Step 9: Create AP Entry
			{
				StepNumber:    9,
				ServiceName:   "accounts-payable",
				HandlerMethod: "CreateAPEntry",
				InputMapping: map[string]string{
					"invoiceID":     "$.steps.6.result.invoice_id",
					"vendorID":      "$.input.vendor_id",
					"invoiceAmount": "$.input.invoice_amount",
					"dueDate":       "$.input.due_date",
					"poID":          "$.steps.1.result.po_id",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{109},
			},
			// Step 10: Post Purchase Journal
			{
				StepNumber:    10,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostPurchaseJournal",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"invoiceID":     "$.steps.6.result.invoice_id",
					"invoiceAmount": "$.input.invoice_amount",
					"vendorID":      "$.input.vendor_id",
					"journalDate":   "$.input.journal_date",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{110},
			},
			// Step 11: Process Vendor Payment
			{
				StepNumber:    11,
				ServiceName:   "banking",
				HandlerMethod: "ProcessVendorPayment",
				InputMapping: map[string]string{
					"invoiceID":     "$.steps.6.result.invoice_id",
					"vendorID":      "$.input.vendor_id",
					"paymentAmount": "$.input.invoice_amount",
					"paymentMethod": "$.input.payment_method",
					"bankAccountID": "$.input.bank_account_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        5,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      120000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{111},
			},
			// Step 12: Close Purchase Order
			{
				StepNumber:    12,
				ServiceName:   "purchase-order",
				HandlerMethod: "ClosePurchaseOrder",
				InputMapping: map[string]string{
					"poID": "$.steps.1.result.po_id",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{112},
			},
			// ===== COMPENSATION STEPS =====

			// Step 101: Cancel PO (compensates step 1)
			{
				StepNumber:    101,
				ServiceName:   "purchase-order",
				HandlerMethod: "CancelPurchaseOrder",
				InputMapping: map[string]string{
					"poID": "$.steps.1.result.po_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 102: Revert PO Approval (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "purchase-order",
				HandlerMethod: "RevertApproval",
				InputMapping: map[string]string{
					"poID": "$.steps.1.result.po_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 104: Cancel Goods Receipt (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "purchase-order",
				HandlerMethod: "CancelGoodsReceipt",
				InputMapping: map[string]string{
					"grnID": "$.steps.4.result.grn_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Reverse Stock Receipt (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "inventory-core",
				HandlerMethod: "ReverseStockReceipt",
				InputMapping: map[string]string{
					"grnID":         "$.steps.4.result.grn_id",
					"receivedItems": "$.input.received_items",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Cancel Invoice (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "CancelInvoice",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.6.result.invoice_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 107: Clear Matching Data (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "ClearMatchingData",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.6.result.invoice_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 108: Revert Invoice Approval (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "RevertApproval",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.6.result.invoice_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 109: Reverse AP Entry (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ReverseAPEntry",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.6.result.invoice_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 110: Reverse Purchase Journal (compensates step 10)
			{
				StepNumber:    110,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReversePurchaseJournal",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.6.result.invoice_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 111: Void Payment (compensates step 11) - PARTIAL (may need manual intervention)
			{
				StepNumber:    111,
				ServiceName:   "banking",
				HandlerMethod: "VoidPayment",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.6.result.invoice_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      60000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 112: Revert PO Closure (compensates step 12)
			{
				StepNumber:    112,
				ServiceName:   "purchase-order",
				HandlerMethod: "RevertClosure",
				InputMapping: map[string]string{
					"poID": "$.steps.1.result.po_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ProcureToPaySaga) SagaType() string {
	return "SAGA-P01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ProcureToPaySaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ProcureToPaySaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ProcureToPaySaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["vendor_id"] == nil {
		return errors.New("vendor_id is required")
	}

	if inputMap["items"] == nil {
		return errors.New("items are required")
	}

	items, ok := inputMap["items"].([]interface{})
	if !ok || len(items) == 0 {
		return errors.New("items must be a non-empty list")
	}

	if inputMap["delivery_date"] == nil {
		return errors.New("delivery_date is required")
	}

	if inputMap["invoice_amount"] == nil {
		return errors.New("invoice_amount is required")
	}

	return nil
}
