// Package banking provides saga handlers for banking module workflows
package banking

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// WireTransferSaga implements SAGA-B01: Wire Transfer & Payment Authorization workflow
// Business Flow: ValidatePayment → AuthorizePayment → CheckAccountBalance → DeductFunds → NotifyBeneficiary → PostPaymentJournal → UpdatePaymentStatus → CompletePayment → LogTransaction → ArchiveTransaction
// Steps: 10 forward + 9 compensation = 19 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,7,10
type WireTransferSaga struct {
	steps []*saga.StepDefinition
}

// NewWireTransferSaga creates a new Wire Transfer saga handler
func NewWireTransferSaga() saga.SagaHandler {
	return &WireTransferSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Validate Payment
			{
				StepNumber:    1,
				ServiceName:   "banking",
				HandlerMethod: "ValidatePayment",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"paymentID":            "$.input.payment_id",
					"paymentAmount":        "$.input.payment_amount",
					"beneficiaryAccount":   "$.input.beneficiary_account",
					"paymentMethod":        "$.input.payment_method",
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
			// Step 2: Authorize Payment
			{
				StepNumber:    2,
				ServiceName:   "payment-gateway",
				HandlerMethod: "AuthorizePayment",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"paymentID":            "$.input.payment_id",
					"paymentAmount":        "$.input.payment_amount",
					"paymentMethod":        "$.input.payment_method",
					"beneficiaryAccount":   "$.input.beneficiary_account",
					"validationResult":     "$.steps.1.result.validation_result",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				CompensationSteps: []int32{102},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Check Account Balance
			{
				StepNumber:    3,
				ServiceName:   "banking",
				HandlerMethod: "CheckAccountBalance",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"paymentID":            "$.input.payment_id",
					"paymentAmount":        "$.input.payment_amount",
					"authorizationToken":   "$.steps.2.result.authorization_token",
				},
				TimeoutSeconds: 25,
				IsCritical:     true,
				CompensationSteps: []int32{103},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Deduct Funds
			{
				StepNumber:    4,
				ServiceName:   "banking",
				HandlerMethod: "DeductFunds",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"paymentID":            "$.input.payment_id",
					"paymentAmount":        "$.input.payment_amount",
					"accountBalance":       "$.steps.3.result.account_balance",
					"authorizationToken":   "$.steps.2.result.authorization_token",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Notify Beneficiary
			{
				StepNumber:    5,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "NotifyBeneficiary",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"paymentID":            "$.input.payment_id",
					"paymentAmount":        "$.input.payment_amount",
					"beneficiaryAccount":   "$.input.beneficiary_account",
					"deductionConfirmation": "$.steps.4.result.deduction_confirmation",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Post Payment Journal
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostPaymentJournal",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"paymentID":            "$.input.payment_id",
					"paymentAmount":        "$.input.payment_amount",
					"beneficiaryAccount":   "$.input.beneficiary_account",
					"deductionConfirmation": "$.steps.4.result.deduction_confirmation",
					"journalDate":          "$.input.journal_date",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Update Payment Status
			{
				StepNumber:    7,
				ServiceName:   "banking",
				HandlerMethod: "UpdatePaymentStatus",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"paymentID":            "$.input.payment_id",
					"paymentStatus":        "TRANSFERRED",
					"journalEntryID":       "$.steps.6.result.journal_entry_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     true,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Complete Payment
			{
				StepNumber:    8,
				ServiceName:   "accounts-payable",
				HandlerMethod: "CompletePayment",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"paymentID":            "$.input.payment_id",
					"paymentAmount":        "$.input.payment_amount",
					"statusUpdate":         "$.steps.7.result.status_update",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Log Transaction
			{
				StepNumber:    9,
				ServiceName:   "approval",
				HandlerMethod: "LogTransaction",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"paymentID":            "$.input.payment_id",
					"paymentAmount":        "$.input.payment_amount",
					"paymentStatus":        "$.steps.7.result.payment_status",
					"completionDetails":    "$.steps.8.result.completion_details",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Archive Transaction
			{
				StepNumber:    10,
				ServiceName:   "banking",
				HandlerMethod: "ArchiveTransaction",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"paymentID":            "$.input.payment_id",
					"transactionLog":       "$.steps.9.result.transaction_log",
					"archiveDate":          "$.input.archive_date",
				},
				TimeoutSeconds: 25,
				IsCritical:     true,
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

			// Step 102: Revoke Payment Authorization (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "payment-gateway",
				HandlerMethod: "RevokePaymentAuthorization",
				InputMapping: map[string]string{
					"paymentID":            "$.input.payment_id",
					"authorizationToken":   "$.steps.2.result.authorization_token",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: RestoreAccountBalance (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "banking",
				HandlerMethod: "RestoreAccountBalance",
				InputMapping: map[string]string{
					"paymentID":            "$.input.payment_id",
					"balanceCheckResult":   "$.steps.3.result.balance_check_result",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 104: ReverseDeduction (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "banking",
				HandlerMethod: "ReverseDeduction",
				InputMapping: map[string]string{
					"paymentID":            "$.input.payment_id",
					"paymentAmount":        "$.input.payment_amount",
					"deductionConfirmation": "$.steps.4.result.deduction_confirmation",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: CancelBeneficiaryNotification (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "CancelBeneficiaryNotification",
				InputMapping: map[string]string{
					"paymentID":            "$.input.payment_id",
					"notificationDetails":  "$.steps.5.result.notification_details",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: ReversePaymentJournal (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReversePaymentJournal",
				InputMapping: map[string]string{
					"paymentID":            "$.input.payment_id",
					"journalEntryID":       "$.steps.6.result.journal_entry_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: RevertPaymentStatus (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "banking",
				HandlerMethod: "RevertPaymentStatus",
				InputMapping: map[string]string{
					"paymentID":            "$.input.payment_id",
					"previousStatus":       "PENDING",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 108: ReversePaymentCompletion (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ReversePaymentCompletion",
				InputMapping: map[string]string{
					"paymentID":            "$.input.payment_id",
					"completionDetails":    "$.steps.8.result.completion_details",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 109: ClearTransactionLog (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "approval",
				HandlerMethod: "ClearTransactionLog",
				InputMapping: map[string]string{
					"paymentID":            "$.input.payment_id",
					"transactionLog":       "$.steps.9.result.transaction_log",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *WireTransferSaga) SagaType() string {
	return "SAGA-B01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *WireTransferSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *WireTransferSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *WireTransferSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["payment_id"] == nil {
		return errors.New("payment_id is required")
	}

	paymentID, ok := inputMap["payment_id"].(string)
	if !ok || paymentID == "" {
		return errors.New("payment_id must be a non-empty string")
	}

	if inputMap["payment_amount"] == nil {
		return errors.New("payment_amount is required")
	}

	paymentAmount, ok := inputMap["payment_amount"].(string)
	if !ok || paymentAmount == "" {
		return errors.New("payment_amount must be a non-empty string")
	}

	if inputMap["beneficiary_account"] == nil {
		return errors.New("beneficiary_account is required")
	}

	beneficiaryAccount, ok := inputMap["beneficiary_account"].(string)
	if !ok || beneficiaryAccount == "" {
		return errors.New("beneficiary_account must be a non-empty string")
	}

	if inputMap["payment_method"] == nil {
		return errors.New("payment_method is required")
	}

	paymentMethod, ok := inputMap["payment_method"].(string)
	if !ok || paymentMethod == "" {
		return errors.New("payment_method must be a non-empty string")
	}

	validMethods := map[string]bool{"NEFT": true, "RTGS": true, "IMPS": true}
	if !validMethods[paymentMethod] {
		return errors.New("payment_method must be one of: NEFT, RTGS, IMPS")
	}

	return nil
}
