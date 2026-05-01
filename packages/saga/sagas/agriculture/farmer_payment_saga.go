// Package agriculture provides saga handlers for agricultural workflows
package agriculture

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// FarmerPaymentSaga implements SAGA-A05: Farmer Payment & Advance Management workflow
// Business Flow: InitiatePayment → ValidatePaymentDetails → ValidateFarmerAccount → AuthorizePayment → ProcessBankTransfer → UpdateFarmerLedger → PostPaymentJournal → RecordPaymentApproval → NotifyPaymentCompletion → ConfirmPayment
// Steps: 9 forward + 8 compensation = 17 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,6,9
type FarmerPaymentSaga struct {
	steps []*saga.StepDefinition
}

// NewFarmerPaymentSaga creates a new Farmer Payment saga handler
func NewFarmerPaymentSaga() saga.SagaHandler {
	return &FarmerPaymentSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Payment
			{
				StepNumber:    1,
				ServiceName:   "agriculture",
				HandlerMethod: "InitiatePayment",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"paymentID":      "$.input.payment_id",
					"farmerID":       "$.input.farmer_id",
					"paymentAmount":  "$.input.payment_amount",
					"paymentDate":    "$.input.payment_date",
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
			// Step 2: Validate Payment Details
			{
				StepNumber:    2,
				ServiceName:   "farmer-management",
				HandlerMethod: "ValidatePaymentDetails",
				InputMapping: map[string]string{
					"paymentID":      "$.steps.1.result.payment_id",
					"farmerID":       "$.input.farmer_id",
					"paymentAmount":  "$.input.payment_amount",
					"validateAmount": "true",
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
			// Step 3: Validate Farmer Account
			{
				StepNumber:    3,
				ServiceName:   "farmer-management",
				HandlerMethod: "ValidateFarmerAccount",
				InputMapping: map[string]string{
					"paymentID":          "$.steps.1.result.payment_id",
					"farmerID":           "$.input.farmer_id",
					"paymentValidation":  "$.steps.2.result.payment_validation",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{102},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Authorize Payment
			{
				StepNumber:    4,
				ServiceName:   "approval",
				HandlerMethod: "AuthorizePayment",
				InputMapping: map[string]string{
					"paymentID":          "$.steps.1.result.payment_id",
					"paymentAmount":      "$.input.payment_amount",
					"farmerValidation":   "$.steps.3.result.farmer_validation",
				},
				TimeoutSeconds:    30,
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
			// Step 5: Process Bank Transfer
			{
				StepNumber:    5,
				ServiceName:   "banking",
				HandlerMethod: "ProcessBankTransfer",
				InputMapping: map[string]string{
					"paymentID":         "$.steps.1.result.payment_id",
					"paymentAmount":     "$.input.payment_amount",
					"authorizationToken": "$.steps.4.result.authorization_token",
					"farmerBankDetails": "$.steps.3.result.farmer_bank_details",
				},
				TimeoutSeconds:    40,
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
			// Step 6: Update Farmer Ledger
			{
				StepNumber:    6,
				ServiceName:   "farmer-management",
				HandlerMethod: "UpdateFarmerLedger",
				InputMapping: map[string]string{
					"paymentID":      "$.steps.1.result.payment_id",
					"farmerID":       "$.input.farmer_id",
					"paymentAmount":  "$.input.payment_amount",
					"transferResult": "$.steps.5.result.transfer_result",
				},
				TimeoutSeconds:    25,
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
			// Step 7: Post Payment Journal Entries
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "ApplyFarmerPaymentJournal",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"paymentID":      "$.steps.1.result.payment_id",
					"paymentAmount":  "$.input.payment_amount",
					"journalDate":    "$.input.payment_date",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Record Payment Approval
			{
				StepNumber:    8,
				ServiceName:   "approval",
				HandlerMethod: "RecordPaymentApprovalRecord",
				InputMapping: map[string]string{
					"paymentID":   "$.steps.1.result.payment_id",
					"approvalID":  "$.steps.4.result.approval_id",
					"farmerID":    "$.input.farmer_id",
				},
				TimeoutSeconds:    20,
				IsCritical:        false,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Confirm Payment
			{
				StepNumber:    9,
				ServiceName:   "agriculture",
				HandlerMethod: "ConfirmPayment",
				InputMapping: map[string]string{
					"paymentID":        "$.steps.1.result.payment_id",
					"journalEntries":   "$.steps.7.result.journal_entries",
					"completionStatus": "Confirmed",
				},
				TimeoutSeconds:    15,
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

			// Step 101: Revert Payment Details Validation (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "farmer-management",
				HandlerMethod: "RevertPaymentDetailsValidation",
				InputMapping: map[string]string{
					"paymentID": "$.steps.1.result.payment_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Revert Farmer Account Validation (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "farmer-management",
				HandlerMethod: "RevertFarmerAccountValidation",
				InputMapping: map[string]string{
					"paymentID": "$.steps.1.result.payment_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: Revoke Payment Authorization (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "approval",
				HandlerMethod: "RevokePaymentAuthorization",
				InputMapping: map[string]string{
					"paymentID": "$.steps.1.result.payment_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Reverse Bank Transfer (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "banking",
				HandlerMethod: "ReverseBankTransfer",
				InputMapping: map[string]string{
					"paymentID": "$.steps.1.result.payment_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
			// Step 105: Revert Farmer Ledger Update (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "farmer-management",
				HandlerMethod: "RevertFarmerLedgerUpdate",
				InputMapping: map[string]string{
					"paymentID": "$.steps.1.result.payment_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Reverse Payment Journal (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseFarmerPaymentJournal",
				InputMapping: map[string]string{
					"paymentID": "$.steps.1.result.payment_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Clear Payment Approval Record (compensates step 8)
			{
				StepNumber:    107,
				ServiceName:   "approval",
				HandlerMethod: "ClearPaymentApprovalRecord",
				InputMapping: map[string]string{
					"paymentID": "$.steps.1.result.payment_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *FarmerPaymentSaga) SagaType() string {
	return "SAGA-A05"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *FarmerPaymentSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *FarmerPaymentSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *FarmerPaymentSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["payment_id"] == nil {
		return errors.New("payment_id is required")
	}

	if inputMap["farmer_id"] == nil {
		return errors.New("farmer_id is required")
	}

	if inputMap["payment_amount"] == nil {
		return errors.New("payment_amount is required")
	}

	if inputMap["payment_date"] == nil {
		return errors.New("payment_date is required (format: YYYY-MM-DD)")
	}

	return nil
}
