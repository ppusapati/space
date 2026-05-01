// Package purchase provides saga handlers for purchase module workflows
package purchase

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// PurchaseReturnSaga implements SAGA-P02: Purchase Return workflow
// Business Flow: Create Return → QC Inspection → Reverse Stock → Create Debit Note → Adjust AP → Reverse GL → Complete Return
type PurchaseReturnSaga struct {
	steps []*saga.StepDefinition
}

// NewPurchaseReturnSaga creates a new Purchase Return saga handler
func NewPurchaseReturnSaga() saga.SagaHandler {
	return &PurchaseReturnSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Purchase Return
			{
				StepNumber:    1,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "CreatePurchaseReturn",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"invoiceID":     "$.input.invoice_id",
					"vendorID":      "$.input.vendor_id",
					"returnReason":  "$.input.return_reason",
					"returnAmount":  "$.input.return_amount",
				},
				TimeoutSeconds: 20,
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
			// Step 2: Inspect Returned Goods
			{
				StepNumber:    2,
				ServiceName:   "qc",
				HandlerMethod: "InspectReturnedGoods",
				InputMapping: map[string]string{
					"returnID":      "$.steps.1.result.return_id",
					"inspectionType": "RETURN_INSPECTION",
					"quantity":      "$.input.quantity",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{},
			},
			// Step 3: Reverse Stock Receipt
			{
				StepNumber:    3,
				ServiceName:   "inventory-core",
				HandlerMethod: "ReverseReceipt",
				InputMapping: map[string]string{
					"invoiceID": "$.input.invoice_id",
					"returnID":  "$.steps.1.result.return_id",
					"quantity":  "$.input.quantity",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{103},
			},
			// Step 4: Create Debit Note
			{
				StepNumber:    4,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "CreateDebitNote",
				InputMapping: map[string]string{
					"returnID":     "$.steps.1.result.return_id",
					"invoiceID":    "$.input.invoice_id",
					"vendorID":     "$.input.vendor_id",
					"debitAmount":  "$.input.return_amount",
					"noteReason":   "Return - " + "$.input.return_reason",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{104},
			},
			// Step 5: Adjust AP for Return
			{
				StepNumber:    5,
				ServiceName:   "accounts-payable",
				HandlerMethod: "AdjustAPForReturn",
				InputMapping: map[string]string{
					"invoiceID":    "$.input.invoice_id",
					"returnID":     "$.steps.1.result.return_id",
					"adjustAmount": "$.input.return_amount",
					"vendorID":     "$.input.vendor_id",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{105},
			},
			// Step 6: Post Return Journal
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostReturnJournal",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"returnID":      "$.steps.1.result.return_id",
					"returnAmount":  "$.input.return_amount",
					"invoiceID":     "$.input.invoice_id",
					"journalDate":   "$.input.journal_date",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{106},
			},
			// Step 7: Complete Return
			{
				StepNumber:    7,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "CompleteReturn",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{107},
			},
			// ===== COMPENSATION STEPS =====

			// Step 101: Cancel Return (compensates step 1)
			{
				StepNumber:    101,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "CancelReturn",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 103: Restore Stock (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "inventory-core",
				HandlerMethod: "RestoreStock",
				InputMapping: map[string]string{
					"invoiceID": "$.input.invoice_id",
					"returnID":  "$.steps.1.result.return_id",
					"quantity":  "$.input.quantity",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 104: Cancel Debit Note (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "CancelDebitNote",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 105: Revert AP Adjustment (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "accounts-payable",
				HandlerMethod: "RevertAPAdjustment",
				InputMapping: map[string]string{
					"invoiceID": "$.input.invoice_id",
					"returnID":  "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: Reverse Return Journal (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseReturnJournal",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 107: Revert Return Completion (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "RevertReturnCompletion",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *PurchaseReturnSaga) SagaType() string {
	return "SAGA-P02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *PurchaseReturnSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *PurchaseReturnSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *PurchaseReturnSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["invoice_id"] == nil {
		return errors.New("invoice_id is required")
	}

	if inputMap["vendor_id"] == nil {
		return errors.New("vendor_id is required")
	}

	if inputMap["return_reason"] == nil {
		return errors.New("return_reason is required")
	}

	if inputMap["return_amount"] == nil {
		return errors.New("return_amount is required")
	}

	if inputMap["quantity"] == nil {
		return errors.New("quantity is required")
	}

	return nil
}
