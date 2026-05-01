// Package healthcare provides saga handlers for healthcare workflows
package healthcare

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ClaimsProcessingSaga implements SAGA-HC03: Insurance Claims Processing workflow
// Business Flow: ValidateClaim → VerifyPatientEligibility → RetrieveServiceRecord → CalculateClaimAmount → ApplyClaimRules → SubmitClaimToInsurer → TrackClaimStatus → ProcessClaimPayment → UpdatePatientAccount → ApplyClaimJournal → GenerateClaimReport → CompleteClaimProcessing
// Steps: 11 forward + 8 compensation = 19 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,7,10
type ClaimsProcessingSaga struct {
	steps []*saga.StepDefinition
}

// NewClaimsProcessingSaga creates a new Insurance Claims Processing saga handler
func NewClaimsProcessingSaga() saga.SagaHandler {
	return &ClaimsProcessingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Validate Claim
			{
				StepNumber:    1,
				ServiceName:   "claims-processing",
				HandlerMethod: "ValidateClaim",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"companyID":   "$.companyID",
					"branchID":    "$.branchID",
					"claimID":     "$.input.claim_id",
					"patientID":   "$.input.patient_id",
					"serviceDate": "$.input.service_date",
					"claimAmount": "$.input.claim_amount",
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
			// Step 2: Verify Patient Eligibility
			{
				StepNumber:    2,
				ServiceName:   "insurance",
				HandlerMethod: "VerifyPatientEligibility",
				InputMapping: map[string]string{
					"claimID":     "$.steps.1.result.claim_id",
					"patientID":   "$.input.patient_id",
					"serviceDate": "$.input.service_date",
					"verifyActive": "true",
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
			// Step 3: Retrieve Service Record
			{
				StepNumber:    3,
				ServiceName:   "medical-records",
				HandlerMethod: "RetrieveServiceRecord",
				InputMapping: map[string]string{
					"claimID":     "$.steps.1.result.claim_id",
					"patientID":   "$.input.patient_id",
					"serviceDate": "$.input.service_date",
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
			// Step 4: Calculate Claim Amount
			{
				StepNumber:    4,
				ServiceName:   "billing",
				HandlerMethod: "CalculateClaimAmount",
				InputMapping: map[string]string{
					"claimID":         "$.steps.1.result.claim_id",
					"patientID":       "$.input.patient_id",
					"claimAmount":     "$.input.claim_amount",
					"serviceRecord":   "$.steps.3.result.service_record",
					"eligibilityData": "$.steps.2.result.eligibility_data",
				},
				TimeoutSeconds:    30,
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
			// Step 5: Apply Claim Rules
			{
				StepNumber:    5,
				ServiceName:   "billing",
				HandlerMethod: "ApplyClaimRules",
				InputMapping: map[string]string{
					"claimID":         "$.steps.1.result.claim_id",
					"patientID":       "$.input.patient_id",
					"calculatedAmount": "$.steps.4.result.calculated_amount",
					"serviceRecord":   "$.steps.3.result.service_record",
				},
				TimeoutSeconds:    25,
				IsCritical:        false,
				CompensationSteps: []int32{103},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Submit Claim to Insurer
			{
				StepNumber:    6,
				ServiceName:   "insurance",
				HandlerMethod: "SubmitClaimToInsurer",
				InputMapping: map[string]string{
					"claimID":        "$.steps.1.result.claim_id",
					"patientID":      "$.input.patient_id",
					"claimAmount":    "$.steps.4.result.calculated_amount",
					"appliedRules":   "$.steps.5.result.applied_rules",
					"serviceRecord":  "$.steps.3.result.service_record",
				},
				TimeoutSeconds:    35,
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
			// Step 7: Track Claim Status
			{
				StepNumber:    7,
				ServiceName:   "claims-processing",
				HandlerMethod: "TrackClaimStatus",
				InputMapping: map[string]string{
					"claimID":        "$.steps.1.result.claim_id",
					"patientID":      "$.input.patient_id",
					"submissionData": "$.steps.6.result.submission_data",
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
			// Step 8: Process Claim Payment
			{
				StepNumber:    8,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ProcessClaimPayment",
				InputMapping: map[string]string{
					"claimID":       "$.steps.1.result.claim_id",
					"patientID":     "$.input.patient_id",
					"claimAmount":   "$.steps.4.result.calculated_amount",
					"claimStatus":   "$.steps.7.result.claim_status",
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
			// Step 9: Update Patient Account
			{
				StepNumber:    9,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "UpdatePatientAccountWithClaim",
				InputMapping: map[string]string{
					"claimID":        "$.steps.1.result.claim_id",
					"patientID":      "$.input.patient_id",
					"claimAmount":    "$.steps.4.result.calculated_amount",
					"paymentData":    "$.steps.8.result.payment_data",
				},
				TimeoutSeconds:    25,
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
			// Step 10: Apply Claim Journal
			{
				StepNumber:    10,
				ServiceName:   "general-ledger",
				HandlerMethod: "ApplyClaimJournal",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"claimID":       "$.steps.1.result.claim_id",
					"claimAmount":   "$.steps.4.result.calculated_amount",
					"journalDate":   "$.input.service_date",
					"paymentData":   "$.steps.8.result.payment_data",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 11: Complete Claim Processing
			{
				StepNumber:    11,
				ServiceName:   "claims-processing",
				HandlerMethod: "CompleteClaimProcessing",
				InputMapping: map[string]string{
					"claimID":         "$.steps.1.result.claim_id",
					"patientID":       "$.input.patient_id",
					"claimAmount":     "$.steps.4.result.calculated_amount",
					"journalEntries":  "$.steps.10.result.journal_entries",
					"completionStatus": "Completed",
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

			// Step 101: Revert Service Record Retrieval (compensates step 3)
			{
				StepNumber:    101,
				ServiceName:   "medical-records",
				HandlerMethod: "RevertServiceRecordRetrieval",
				InputMapping: map[string]string{
					"claimID": "$.steps.1.result.claim_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Revert Claim Amount Calculation (compensates step 4)
			{
				StepNumber:    102,
				ServiceName:   "billing",
				HandlerMethod: "RevertClaimAmountCalculation",
				InputMapping: map[string]string{
					"claimID": "$.steps.1.result.claim_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Clear Claim Rules Application (compensates step 5)
			{
				StepNumber:    103,
				ServiceName:   "billing",
				HandlerMethod: "ClearClaimRulesApplication",
				InputMapping: map[string]string{
					"claimID": "$.steps.1.result.claim_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 104: Withdraw Claim Submission (compensates step 6)
			{
				StepNumber:    104,
				ServiceName:   "insurance",
				HandlerMethod: "WithdrawClaimSubmission",
				InputMapping: map[string]string{
					"claimID": "$.steps.1.result.claim_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 105: Revert Claim Status Tracking (compensates step 7)
			{
				StepNumber:    105,
				ServiceName:   "claims-processing",
				HandlerMethod: "RevertClaimStatusTracking",
				InputMapping: map[string]string{
					"claimID": "$.steps.1.result.claim_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Reverse Claim Payment Processing (compensates step 8)
			{
				StepNumber:    106,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ReverseClaimPaymentProcessing",
				InputMapping: map[string]string{
					"claimID": "$.steps.1.result.claim_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Revert Patient Account Update (compensates step 9)
			{
				StepNumber:    107,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "RevertPatientAccountUpdateForClaim",
				InputMapping: map[string]string{
					"claimID": "$.steps.1.result.claim_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 108: Reverse Claim Journal (compensates step 10)
			{
				StepNumber:    108,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseClaimJournal",
				InputMapping: map[string]string{
					"claimID": "$.steps.1.result.claim_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ClaimsProcessingSaga) SagaType() string {
	return "SAGA-HC03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ClaimsProcessingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ClaimsProcessingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ClaimsProcessingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["claim_id"] == nil {
		return errors.New("claim_id is required")
	}

	if inputMap["patient_id"] == nil {
		return errors.New("patient_id is required")
	}

	if inputMap["service_date"] == nil {
		return errors.New("service_date is required (format: YYYY-MM-DD)")
	}

	if inputMap["claim_amount"] == nil {
		return errors.New("claim_amount is required")
	}

	return nil
}
