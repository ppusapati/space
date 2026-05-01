// Package retail provides saga handlers for retail workflows
package retail

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ReturnRefundSaga implements SAGA-R03: Return and Refund Processing workflow
// Business Flow: InitiateReturn → ValidateReturnEligibility → InspectReturnedItems → UpdateInventoryReceived → GenerateDebitNote → ProcessRefund → UpdateAccountsReceivable → ApplyReturnJournal → RecordReturnCompletion
// Steps: 9 forward + 10 compensation = 19 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,6,8,9
type ReturnRefundSaga struct {
	steps []*saga.StepDefinition
}

// NewReturnRefundSaga creates a new Return and Refund Processing saga handler
func NewReturnRefundSaga() saga.SagaHandler {
	return &ReturnRefundSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Return
			{
				StepNumber:    1,
				ServiceName:   "returns",
				HandlerMethod: "InitiateReturn",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"returnID":        "$.input.return_id",
					"originalOrderID": "$.input.original_order_id",
					"customerID":      "$.input.customer_id",
					"returnReason":    "$.input.return_reason",
					"returnDate":      "$.input.return_date",
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
			// Step 2: Validate Return Eligibility
			{
				StepNumber:    2,
				ServiceName:   "sales",
				HandlerMethod: "ValidateReturnEligibility",
				InputMapping: map[string]string{
					"returnID":        "$.steps.1.result.return_id",
					"originalOrderID": "$.input.original_order_id",
					"customerID":      "$.input.customer_id",
					"returnReason":    "$.input.return_reason",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Inspect Returned Items
			{
				StepNumber:    3,
				ServiceName:   "returns",
				HandlerMethod: "InspectReturnedItems",
				InputMapping: map[string]string{
					"returnID":        "$.steps.1.result.return_id",
					"itemDetails":     "$.input.item_details",
					"inspectionNotes": "$.input.inspection_notes",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Update Inventory - Received
			{
				StepNumber:    4,
				ServiceName:   "inventory",
				HandlerMethod: "UpdateInventoryReceived",
				InputMapping: map[string]string{
					"returnID":    "$.steps.1.result.return_id",
					"itemDetails": "$.steps.3.result.inspected_items",
					"warehouseID": "$.input.warehouse_id",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{101},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Generate Debit Note
			{
				StepNumber:    5,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "GenerateDebitNote",
				InputMapping: map[string]string{
					"returnID":      "$.steps.1.result.return_id",
					"customerID":    "$.input.customer_id",
					"refundAmount":  "$.input.refund_amount",
					"itemDetails":   "$.steps.3.result.inspected_items",
				},
				TimeoutSeconds:    25,
				IsCritical:        false,
				CompensationSteps: []int32{102},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Process Refund
			{
				StepNumber:    6,
				ServiceName:   "payment-gateway",
				HandlerMethod: "ProcessRefund",
				InputMapping: map[string]string{
					"returnID":       "$.steps.1.result.return_id",
					"refundAmount":   "$.input.refund_amount",
					"paymentMethod":  "$.input.payment_method",
					"originalOrderID": "$.input.original_order_id",
				},
				TimeoutSeconds:    35,
				IsCritical:        true,
				CompensationSteps: []int32{103},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Update Accounts Receivable
			{
				StepNumber:    7,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "UpdateAccountsReceivable",
				InputMapping: map[string]string{
					"returnID":   "$.steps.1.result.return_id",
					"customerID": "$.input.customer_id",
					"debitNote":  "$.steps.5.result.debit_note_id",
					"refundAmount": "$.input.refund_amount",
				},
				TimeoutSeconds:    25,
				IsCritical:        false,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Apply Return Journal
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "ApplyReturnJournal",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"returnID":       "$.steps.1.result.return_id",
					"refundAmount":   "$.input.refund_amount",
					"debitNoteID":    "$.steps.5.result.debit_note_id",
					"journalDate":    "$.input.return_date",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Record Return Completion
			{
				StepNumber:    9,
				ServiceName:   "returns",
				HandlerMethod: "RecordReturnCompletion",
				InputMapping: map[string]string{
					"returnID":        "$.steps.1.result.return_id",
					"journalEntries":  "$.steps.8.result.journal_entries",
					"refundStatus":    "$.steps.6.result.refund_status",
					"completionStatus": "Completed",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// ===== COMPENSATION STEPS =====

			// Step 101: Revert Inventory Receipt (compensates step 4)
			{
				StepNumber:    101,
				ServiceName:   "inventory",
				HandlerMethod: "RevertInventoryReceived",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Revert Debit Note (compensates step 5)
			{
				StepNumber:    102,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "RevertDebitNote",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: Reverse Refund (compensates step 6)
			{
				StepNumber:    103,
				ServiceName:   "payment-gateway",
				HandlerMethod: "ReverseRefund",
				InputMapping: map[string]string{
					"returnID":      "$.steps.1.result.return_id",
					"refundAmount":  "$.input.refund_amount",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 104: Revert Accounts Receivable Update (compensates step 7)
			{
				StepNumber:    104,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "RevertAccountsReceivableUpdate",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Reverse Return Journal (compensates step 8)
			{
				StepNumber:    105,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseReturnJournal",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: Revert Initiate Return (compensates step 1)
			{
				StepNumber:    106,
				ServiceName:   "returns",
				HandlerMethod: "RevertInitiateReturn",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: Revert Validate Return Eligibility (compensates step 2)
			{
				StepNumber:    107,
				ServiceName:   "sales",
				HandlerMethod: "RevertValidateReturnEligibility",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 108: Revert Inspect Returned Items (compensates step 3)
			{
				StepNumber:    108,
				ServiceName:   "returns",
				HandlerMethod: "RevertInspectReturnedItems",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 109: Revert Generate Debit Note (compensates step 5)
			{
				StepNumber:    109,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "RevertGenerateDebitNote",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 110: Revert Refund Record Completion (compensates step 9)
			{
				StepNumber:    110,
				ServiceName:   "returns",
				HandlerMethod: "RevertRecordReturnCompletion",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ReturnRefundSaga) SagaType() string {
	return "SAGA-R03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ReturnRefundSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ReturnRefundSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ReturnRefundSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["return_id"] == nil {
		return errors.New("return_id is required")
	}

	if inputMap["original_order_id"] == nil {
		return errors.New("original_order_id is required")
	}

	if inputMap["refund_amount"] == nil {
		return errors.New("refund_amount is required")
	}

	return nil
}
