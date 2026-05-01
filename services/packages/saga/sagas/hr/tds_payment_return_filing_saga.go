// Package hr provides saga handlers for HR & Payroll module workflows
package hr

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// TDSPaymentReturnFilingSaga implements SAGA-SR03: TDS Payment & Return Filing workflow
// Business Flow: Extract TDS deductions → Classify by section → Validate deductee PAN/name/address →
// Calculate total TDS deducted → Verify TDS deposited with bank → Match TDS deposited vs deducted →
// Identify discrepancies → Create TDS return schedules → Generate TDS return form →
// Submit to TDS Clearing House → Archive return (11 forward steps + 11 compensation = 22 total)
// Critical steps: 4, 6, 10 (3 critical)
// Non-critical steps: 1, 2, 3, 5, 7, 8, 9, 11 (8 non-critical)
// Timeout: 360s aggregate (mix of 20-45s per step)
// Statutory compliance: Income Tax Sections 203-204 (TDS collection, quarterly/annual return)
type TDSPaymentReturnFilingSaga struct {
	steps []*saga.StepDefinition
}

// NewTDSPaymentReturnFilingSaga creates a new TDS Payment & Return Filing saga handler
func NewTDSPaymentReturnFilingSaga() saga.SagaHandler {
	return &TDSPaymentReturnFilingSaga{
		steps: []*saga.StepDefinition{
			// ===== FORWARD STEPS (1-11) =====

			// Step 1: Extract TDS Deductions - NON-CRITICAL
			{
				StepNumber:    1,
				ServiceName:   "tds",
				HandlerMethod: "ExtractTDSDeductions",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"tdsReturnFilingID": "$.input.tds_return_filing_id",
					"filingPeriod":      "$.input.filing_period",
					"assessmentYear":    "$.input.assessment_year",
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
			// Step 2: Classify by Section - NON-CRITICAL
			{
				StepNumber:    2,
				ServiceName:   "tds",
				HandlerMethod: "ClassifyTDSBySection",
				InputMapping: map[string]string{
					"tenantID":               "$.tenantID",
					"companyID":              "$.companyID",
					"branchID":               "$.branchID",
					"tdsReturnFilingID":      "$.input.tds_return_filing_id",
					"filingPeriod":           "$.input.filing_period",
					"tdsDeductionsExtracted": "$.steps.1.result.tds_deductions_extracted",
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
				CompensationSteps: []int32{101},
			},
			// Step 3: Validate Deductee PAN/Name/Address - NON-CRITICAL
			{
				StepNumber:    3,
				ServiceName:   "employee",
				HandlerMethod: "ValidateDeducteeDetails",
				InputMapping: map[string]string{
					"tenantID":                  "$.tenantID",
					"companyID":                 "$.companyID",
					"branchID":                  "$.branchID",
					"tdsReturnFilingID":         "$.input.tds_return_filing_id",
					"filingPeriod":              "$.input.filing_period",
					"tdsClassificationComplete": "$.steps.2.result.tds_classification_complete",
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
				CompensationSteps: []int32{102},
			},
			// Step 4: Calculate Total TDS Deducted - CRITICAL
			{
				StepNumber:    4,
				ServiceName:   "tds",
				HandlerMethod: "CalculateTotalTDSDeducted",
				InputMapping: map[string]string{
					"tenantID":                 "$.tenantID",
					"companyID":                "$.companyID",
					"branchID":                 "$.branchID",
					"tdsReturnFilingID":        "$.input.tds_return_filing_id",
					"filingPeriod":             "$.input.filing_period",
					"deducteeDetailsValidated": "$.steps.3.result.deductee_details_validated",
				},
				TimeoutSeconds: 40,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{103},
			},
			// Step 5: Verify TDS Deposited with Bank - NON-CRITICAL
			{
				StepNumber:    5,
				ServiceName:   "banking",
				HandlerMethod: "VerifyTDSDeposited",
				InputMapping: map[string]string{
					"tenantID":                   "$.tenantID",
					"companyID":                  "$.companyID",
					"branchID":                   "$.branchID",
					"tdsReturnFilingID":          "$.input.tds_return_filing_id",
					"filingPeriod":               "$.input.filing_period",
					"totalTDSDeductedCalculated": "$.steps.4.result.total_tds_deducted_calculated",
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
				CompensationSteps: []int32{104},
			},
			// Step 6: Match TDS Deposited vs Deducted - CRITICAL
			{
				StepNumber:    6,
				ServiceName:   "tds",
				HandlerMethod: "MatchTDSDepositedVsDeducted",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"tdsReturnFilingID":  "$.input.tds_return_filing_id",
					"filingPeriod":       "$.input.filing_period",
					"tdsDepositVerified": "$.steps.5.result.tds_deposit_verified",
				},
				TimeoutSeconds: 40,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{105},
			},
			// Step 7: Identify Discrepancies - NON-CRITICAL
			{
				StepNumber:    7,
				ServiceName:   "tds",
				HandlerMethod: "IdentifyDiscrepancies",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"tdsReturnFilingID": "$.input.tds_return_filing_id",
					"filingPeriod":      "$.input.filing_period",
					"tdsMatched":        "$.steps.6.result.tds_matched",
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
				CompensationSteps: []int32{106},
			},
			// Step 8: Create TDS Return Schedules - NON-CRITICAL
			{
				StepNumber:    8,
				ServiceName:   "tds",
				HandlerMethod: "CreateTDSReturnSchedules",
				InputMapping: map[string]string{
					"tenantID":                "$.tenantID",
					"companyID":               "$.companyID",
					"branchID":                "$.branchID",
					"tdsReturnFilingID":       "$.input.tds_return_filing_id",
					"filingPeriod":            "$.input.filing_period",
					"discrepanciesIdentified": "$.steps.7.result.discrepancies_identified",
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
				CompensationSteps: []int32{107},
			},
			// Step 9: Generate TDS Return Form - NON-CRITICAL
			{
				StepNumber:    9,
				ServiceName:   "tds",
				HandlerMethod: "GenerateTDSReturnForm",
				InputMapping: map[string]string{
					"tenantID":                  "$.tenantID",
					"companyID":                 "$.companyID",
					"branchID":                  "$.branchID",
					"tdsReturnFilingID":         "$.input.tds_return_filing_id",
					"filingPeriod":              "$.input.filing_period",
					"tdsReturnSchedulesCreated": "$.steps.8.result.tds_return_schedules_created",
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
				CompensationSteps: []int32{108},
			},
			// Step 10: Submit to TDS Clearing House - CRITICAL
			{
				StepNumber:    10,
				ServiceName:   "compliance-postings",
				HandlerMethod: "SubmitTDSReturnToClearingHouse",
				InputMapping: map[string]string{
					"tenantID":               "$.tenantID",
					"companyID":              "$.companyID",
					"branchID":               "$.branchID",
					"tdsReturnFilingID":      "$.input.tds_return_filing_id",
					"filingPeriod":           "$.input.filing_period",
					"tdsReturnFormGenerated": "$.steps.9.result.tds_return_form_generated",
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
				CompensationSteps: []int32{109},
			},
			// Step 11: Archive Return for Compliance Record - NON-CRITICAL
			{
				StepNumber:    11,
				ServiceName:   "tds",
				HandlerMethod: "ArchiveReturnForCompliance",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"tdsReturnFilingID":  "$.input.tds_return_filing_id",
					"filingPeriod":       "$.input.filing_period",
					"tdsReturnSubmitted": "$.steps.10.result.tds_return_submitted",
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
				CompensationSteps: []int32{110},
			},

			// ===== COMPENSATION STEPS (101-111) =====

			// Step 101: Undo Classify by Section (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "tds",
				HandlerMethod: "UndoClassifyTDSBySection",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"tdsReturnFilingID": "$.input.tds_return_filing_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Undo Validate Deductee Details (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "employee",
				HandlerMethod: "UndoValidateDeducteeDetails",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"tdsReturnFilingID": "$.input.tds_return_filing_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Undo Calculate Total TDS Deducted (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "tds",
				HandlerMethod: "UndoCalculateTotalTDSDeducted",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"tdsReturnFilingID": "$.input.tds_return_filing_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Undo Verify TDS Deposited (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "banking",
				HandlerMethod: "UndoVerifyTDSDeposited",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"tdsReturnFilingID": "$.input.tds_return_filing_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: Undo Match TDS Deposited vs Deducted (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "tds",
				HandlerMethod: "UndoMatchTDSDepositedVsDeducted",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"tdsReturnFilingID": "$.input.tds_return_filing_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: Undo Identify Discrepancies (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "tds",
				HandlerMethod: "UndoIdentifyDiscrepancies",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"tdsReturnFilingID": "$.input.tds_return_filing_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Undo Create TDS Return Schedules (compensates step 8)
			{
				StepNumber:    107,
				ServiceName:   "tds",
				HandlerMethod: "UndoCreateTDSReturnSchedules",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"tdsReturnFilingID": "$.input.tds_return_filing_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: Undo Generate TDS Return Form (compensates step 9)
			{
				StepNumber:    108,
				ServiceName:   "tds",
				HandlerMethod: "UndoGenerateTDSReturnForm",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"tdsReturnFilingID": "$.input.tds_return_filing_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 109: Rollback TDS Return Submission (compensates step 10)
			{
				StepNumber:    109,
				ServiceName:   "compliance-postings",
				HandlerMethod: "UndoSubmitTDSReturnToClearingHouse",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"tdsReturnFilingID": "$.input.tds_return_filing_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 110: Undo Archive Return (compensates step 11)
			{
				StepNumber:    110,
				ServiceName:   "tds",
				HandlerMethod: "UndoArchiveReturnForCompliance",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"tdsReturnFilingID": "$.input.tds_return_filing_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *TDSPaymentReturnFilingSaga) SagaType() string {
	return "SAGA-SR03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *TDSPaymentReturnFilingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *TDSPaymentReturnFilingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: tds_return_filing_id, filing_period, assessment_year, company_id
func (s *TDSPaymentReturnFilingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	// Extract the nested 'input' object
	innerInput, ok := inputMap["input"].(map[string]interface{})
	if !ok {
		return errors.New("missing 'input' field in saga input")
	}

	// Validate tds_return_filing_id
	if innerInput["tds_return_filing_id"] == nil {
		return errors.New("missing required field: tds_return_filing_id")
	}
	tdsReturnFilingID, ok := innerInput["tds_return_filing_id"].(string)
	if !ok || tdsReturnFilingID == "" {
		return errors.New("tds_return_filing_id must be a non-empty string")
	}

	// Validate filing_period
	if innerInput["filing_period"] == nil {
		return errors.New("missing required field: filing_period")
	}
	filingPeriod, ok := innerInput["filing_period"].(string)
	if !ok || filingPeriod == "" {
		return errors.New("filing_period must be a non-empty string")
	}

	// Validate assessment_year
	if innerInput["assessment_year"] == nil {
		return errors.New("missing required field: assessment_year")
	}
	assessmentYear, ok := innerInput["assessment_year"].(string)
	if !ok || assessmentYear == "" {
		return errors.New("assessment_year must be a non-empty string")
	}

	// Validate company_id
	if innerInput["company_id"] == nil {
		return errors.New("missing required field: company_id")
	}
	companyID, ok := innerInput["company_id"].(string)
	if !ok || companyID == "" {
		return errors.New("company_id must be a non-empty string")
	}

	return nil
}
