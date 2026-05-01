// Package gst provides saga handlers for GST compliance workflows
package gst

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// GSTAmendmentSaga implements SAGA-G05: GST Amendment & Correction workflow
// Business Flow: ValidateAmendment → CalculateTaxImpact → PostAmendmentJournal → UpdateGSTR2 → CorrectionNotice → UpdateCompliance → CompleteAmendment
// GST Compliance: Amendment and correction of GST filings per GST Rules
type GSTAmendmentSaga struct {
	steps []*saga.StepDefinition
}

// NewGSTAmendmentSaga creates a new GST Amendment saga handler
func NewGSTAmendmentSaga() saga.SagaHandler {
	return &GSTAmendmentSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Validate Amendment Request
			{
				StepNumber:    1,
				ServiceName:   "gst-amendment",
				HandlerMethod: "ValidateAmendmentRequest",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"amendmentType":  "$.input.amendment_type",
					"originalPeriod": "$.input.original_period",
					"gstin":          "$.input.gstin",
					"fiscalYear":     "$.input.fiscal_year",
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
			// Step 2: Calculate Tax Impact of Amendment
			{
				StepNumber:    2,
				ServiceName:   "tax-engine",
				HandlerMethod: "CalculateAmendmentTaxImpact",
				InputMapping: map[string]string{
					"amendmentID":    "$.steps.1.result.amendment_id",
					"amendmentType":  "$.input.amendment_type",
					"originalPeriod": "$.input.original_period",
					"amendmentData":  "$.input.amendment_data",
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
			// Step 3: Post Amendment Journal Entry
			{
				StepNumber:    3,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostAmendmentJournalEntry",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"amendmentID":   "$.steps.1.result.amendment_id",
					"taxImpact":     "$.steps.2.result.tax_impact",
					"journalDate":   "$.input.amendment_date",
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
			// Step 4: Update GSTR-2 with Amendment
			{
				StepNumber:    4,
				ServiceName:   "gst-return",
				HandlerMethod: "UpdateGSTR2WithAmendment",
				InputMapping: map[string]string{
					"amendmentID":    "$.steps.1.result.amendment_id",
					"originalPeriod": "$.input.original_period",
					"taxImpact":      "$.steps.2.result.tax_impact",
					"gstin":          "$.input.gstin",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Update Accounts Payable with Corrections
			{
				StepNumber:    5,
				ServiceName:   "accounts-payable",
				HandlerMethod: "UpdateAPWithAmendmentCorrections",
				InputMapping: map[string]string{
					"amendmentID":    "$.steps.1.result.amendment_id",
					"originalPeriod": "$.input.original_period",
					"taxImpact":      "$.steps.2.result.tax_impact",
					"amendmentType":  "$.input.amendment_type",
				},
				TimeoutSeconds:    25,
				IsCritical:        false,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Generate Correction Notice
			{
				StepNumber:    6,
				ServiceName:   "gst-amendment",
				HandlerMethod: "GenerateCorrectionNotice",
				InputMapping: map[string]string{
					"amendmentID":    "$.steps.1.result.amendment_id",
					"amendmentType":  "$.input.amendment_type",
					"originalPeriod": "$.input.original_period",
					"taxImpact":      "$.steps.2.result.tax_impact",
				},
				TimeoutSeconds:    20,
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
			// Step 7: Update Compliance Records
			{
				StepNumber:    7,
				ServiceName:   "gst-amendment",
				HandlerMethod: "UpdateComplianceRecords",
				InputMapping: map[string]string{
					"amendmentID":        "$.steps.1.result.amendment_id",
					"correctionNotice":   "$.steps.6.result.correction_notice",
					"originalPeriod":     "$.input.original_period",
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
			// Step 8: Complete Amendment Process
			{
				StepNumber:    8,
				ServiceName:   "gst-amendment",
				HandlerMethod: "CompleteAmendmentProcess",
				InputMapping: map[string]string{
					"amendmentID":      "$.steps.1.result.amendment_id",
					"complianceStatus": "AMENDED",
					"completionDate":   "$.input.amendment_date",
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

			// Step 101: Cancel Amendment Validation (compensates step 1)
			{
				StepNumber:    101,
				ServiceName:   "gst-amendment",
				HandlerMethod: "CancelAmendmentValidation",
				InputMapping: map[string]string{
					"amendmentID": "$.steps.1.result.amendment_id",
					"reason":      "Saga compensation - GST amendment failed",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 102: Clear Tax Impact Calculation (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "tax-engine",
				HandlerMethod: "ClearAmendmentTaxImpactCalculation",
				InputMapping: map[string]string{
					"amendmentID": "$.steps.1.result.amendment_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: Reverse Amendment Journal (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseAmendmentJournal",
				InputMapping: map[string]string{
					"amendmentID": "$.steps.1.result.amendment_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Revert GSTR-2 Amendment (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "gst-return",
				HandlerMethod: "RevertGSTR2Amendment",
				InputMapping: map[string]string{
					"amendmentID":    "$.steps.1.result.amendment_id",
					"originalPeriod": "$.input.original_period",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Revert AP Corrections (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "accounts-payable",
				HandlerMethod: "RevertAPAmendmentCorrections",
				InputMapping: map[string]string{
					"amendmentID": "$.steps.1.result.amendment_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Delete Correction Notice (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "gst-amendment",
				HandlerMethod: "DeleteCorrectionNotice",
				InputMapping: map[string]string{
					"amendmentID": "$.steps.1.result.amendment_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: Clear Compliance Records Update (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "gst-amendment",
				HandlerMethod: "ClearComplianceRecordsUpdate",
				InputMapping: map[string]string{
					"amendmentID": "$.steps.1.result.amendment_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *GSTAmendmentSaga) SagaType() string {
	return "SAGA-G05"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *GSTAmendmentSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *GSTAmendmentSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *GSTAmendmentSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["amendment_type"] == nil {
		return errors.New("amendment_type is required (INVOICE_CORRECTION, TAX_CORRECTION, EXEMPTION_CHANGE, etc.)")
	}

	amendmentType, ok := inputMap["amendment_type"].(string)
	if !ok {
		return errors.New("amendment_type must be a string")
	}

	validTypes := map[string]bool{
		"INVOICE_CORRECTION": true,
		"TAX_CORRECTION":     true,
		"EXEMPTION_CHANGE":   true,
		"RATE_CHANGE":        true,
		"RETURN_CORRECTION":  true,
	}

	if !validTypes[amendmentType] {
		return errors.New("amendment_type must be INVOICE_CORRECTION, TAX_CORRECTION, EXEMPTION_CHANGE, RATE_CHANGE, or RETURN_CORRECTION")
	}

	if inputMap["original_period"] == nil {
		return errors.New("original_period is required (format: YYYY-MM)")
	}

	if inputMap["gstin"] == nil {
		return errors.New("gstin is required")
	}

	if inputMap["fiscal_year"] == nil {
		return errors.New("fiscal_year is required")
	}

	if inputMap["amendment_date"] == nil {
		return errors.New("amendment_date is required")
	}

	if inputMap["amendment_data"] == nil {
		return errors.New("amendment_data is required")
	}

	return nil
}
