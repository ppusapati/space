// Package banking provides saga handlers for banking module workflows
package banking

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ChequeManagementSaga implements SAGA-B04: Cheque Management & Processing workflow
// Business Flow: ValidateCheque → IssueCheque → RecordCheque → ProcessCheque → UpdateChequeStatus → PostChequeEntry → ClearCheque → ArchiveCheque → CloseTransaction
// Timeout: 120 seconds, Critical steps: 1,2,3,4,6,9
type ChequeManagementSaga struct {
	steps []*saga.StepDefinition
}

// NewChequeManagementSaga creates a new Cheque Management & Processing saga handler
func NewChequeManagementSaga() saga.SagaHandler {
	return &ChequeManagementSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Validate Cheque
			{
				StepNumber:    1,
				ServiceName:   "banking",
				HandlerMethod: "ValidateCheque",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"chequeID":      "$.input.cheque_id",
					"chequeAmount":  "$.input.cheque_amount",
					"payee":         "$.input.payee",
					"chequeDate":    "$.input.cheque_date",
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
			// Step 2: Issue Cheque
			{
				StepNumber:    2,
				ServiceName:   "cheque-management",
				HandlerMethod: "IssueCheque",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"chequeID":        "$.input.cheque_id",
					"chequeAmount":    "$.input.cheque_amount",
					"payee":           "$.input.payee",
					"chequeDate":      "$.input.cheque_date",
					"validationResult": "$.steps.1.result.validation_result",
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
			// Step 3: Record Cheque
			{
				StepNumber:    3,
				ServiceName:   "banking",
				HandlerMethod: "RecordCheque",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"chequeID":        "$.input.cheque_id",
					"chequeAmount":    "$.input.cheque_amount",
					"payee":           "$.input.payee",
					"chequeDate":      "$.input.cheque_date",
					"issueDetails":    "$.steps.2.result.issue_details",
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
			// Step 4: Process Cheque
			{
				StepNumber:    4,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ProcessCheque",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"chequeID":        "$.input.cheque_id",
					"chequeAmount":    "$.input.cheque_amount",
					"payee":           "$.input.payee",
					"chequeRecord":    "$.steps.3.result.cheque_record",
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
			// Step 5: Update Cheque Status
			{
				StepNumber:    5,
				ServiceName:   "banking",
				HandlerMethod: "UpdateChequeStatus",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"chequeID":        "$.input.cheque_id",
					"chequeStatus":    "PROCESSED",
					"processDetails":  "$.steps.4.result.process_details",
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
			// Step 6: Post Cheque Entry
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostChequeEntry",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"chequeID":        "$.input.cheque_id",
					"chequeAmount":    "$.input.cheque_amount",
					"payee":           "$.input.payee",
					"statusUpdate":    "$.steps.5.result.status_update",
					"journalDate":     "$.input.cheque_date",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Clear Cheque
			{
				StepNumber:    7,
				ServiceName:   "cheque-management",
				HandlerMethod: "ClearCheque",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"chequeID":        "$.input.cheque_id",
					"journalEntryID":  "$.steps.6.result.journal_entry_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Archive Cheque
			{
				StepNumber:    8,
				ServiceName:   "banking",
				HandlerMethod: "ArchiveCheque",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"chequeID":        "$.input.cheque_id",
					"chequeRecord":    "$.steps.3.result.cheque_record",
					"clearanceDetails": "$.steps.7.result.clearance_details",
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
			// Step 9: Close Transaction
			{
				StepNumber:    9,
				ServiceName:   "accounting",
				HandlerMethod: "CloseTransaction",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"chequeID":        "$.input.cheque_id",
					"chequeAmount":    "$.input.cheque_amount",
					"archiveDetails":  "$.steps.8.result.archive_details",
					"closureDate":     "$.input.cheque_date",
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

			// Step 102: RevokeChequIssuance (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "cheque-management",
				HandlerMethod: "RevokeChequIssuance",
				InputMapping: map[string]string{
					"chequeID":      "$.input.cheque_id",
					"issueDetails":  "$.steps.2.result.issue_details",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: ReverseChequeRecord (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "banking",
				HandlerMethod: "ReverseChequeRecord",
				InputMapping: map[string]string{
					"chequeID":      "$.input.cheque_id",
					"chequeRecord":  "$.steps.3.result.cheque_record",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 104: ReverseChequeProcessing (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ReverseChequeProcessing",
				InputMapping: map[string]string{
					"chequeID":       "$.input.cheque_id",
					"processDetails": "$.steps.4.result.process_details",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: RevertChequeStatus (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "banking",
				HandlerMethod: "RevertChequeStatus",
				InputMapping: map[string]string{
					"chequeID":       "$.input.cheque_id",
					"previousStatus": "ISSUED",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: ReverseChequeEntry (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseChequeEntry",
				InputMapping: map[string]string{
					"chequeID":       "$.input.cheque_id",
					"journalEntryID": "$.steps.6.result.journal_entry_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: ReverseChequeClearing (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "cheque-management",
				HandlerMethod: "ReverseChequeClearing",
				InputMapping: map[string]string{
					"chequeID":         "$.input.cheque_id",
					"clearanceDetails": "$.steps.7.result.clearance_details",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 108: ReverseArchival (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "banking",
				HandlerMethod: "ReverseArchival",
				InputMapping: map[string]string{
					"chequeID":       "$.input.cheque_id",
					"archiveDetails": "$.steps.8.result.archive_details",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ChequeManagementSaga) SagaType() string {
	return "SAGA-B04"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ChequeManagementSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ChequeManagementSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ChequeManagementSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["cheque_id"] == nil {
		return errors.New("cheque_id is required")
	}

	chequeID, ok := inputMap["cheque_id"].(string)
	if !ok || chequeID == "" {
		return errors.New("cheque_id must be a non-empty string")
	}

	if inputMap["cheque_amount"] == nil {
		return errors.New("cheque_amount is required")
	}

	chequeAmount, ok := inputMap["cheque_amount"].(string)
	if !ok || chequeAmount == "" {
		return errors.New("cheque_amount must be a non-empty string")
	}

	if inputMap["payee"] == nil {
		return errors.New("payee is required")
	}

	payee, ok := inputMap["payee"].(string)
	if !ok || payee == "" {
		return errors.New("payee must be a non-empty string")
	}

	if inputMap["cheque_date"] == nil {
		return errors.New("cheque_date is required")
	}

	chequeDate, ok := inputMap["cheque_date"].(string)
	if !ok || chequeDate == "" {
		return errors.New("cheque_date must be a non-empty string")
	}

	return nil
}
