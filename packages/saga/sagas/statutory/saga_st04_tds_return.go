// Package statutory provides saga handlers for statutory compliance workflows
package statutory

import (
	"errors"
	"fmt"

	"p9e.in/samavaya/packages/saga"
)

// TDSReturnFilingSaga implements SAGA-ST04: TDS Return Filing (Tax Deduction Return) workflow
// Business Flow: ExtractTDSTransactions → ClassifyBySection → ValidateDeducteeInfo →
//               CalculateTDSDeduction → VerifyTDSDeposited → MatchTDSDepositedVsDeducted →
//               GenerateTDSSchedules → CreateTDSReturn → SubmitToTDSClearingHouse → ArchiveReturn
// TDS Compliance: Tax Deducted at Source returns filed quarterly/annually
// Standard TDS Sections: 194C (contractors), 194D (insurance), 194H (brokerage), etc.
// Critical Steps: 4 (CalculateTDSDeduction), 7 (GenerateTDSSchedules), 9 (SubmitToTDSClearingHouse)
type TDSReturnFilingSaga struct {
	steps []*saga.StepDefinition
}

// NewTDSReturnFilingSaga creates a new TDS Return Filing saga handler
func NewTDSReturnFilingSaga() saga.SagaHandler {
	return &TDSReturnFilingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Extract TDS Transactions (Quarterly/Annual)
			{
				StepNumber:    1,
				ServiceName:   "tds",
				HandlerMethod: "ExtractTDSTransactions",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"tan":               "$.input.tan",
					"filingPeriod":      "$.input.filing_period",
					"periodType":        "$.input.period_type",
					"periodStartDate":   "$.input.period_start_date",
					"periodEndDate":     "$.input.period_end_date",
				},
				TimeoutSeconds: 35,
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
			// Step 2: Classify TDS Transactions by Section (194C, 194D, 194H, etc.)
			{
				StepNumber:    2,
				ServiceName:   "tds",
				HandlerMethod: "ClassifyTDSBySection",
				InputMapping: map[string]string{
					"tdsTransactions":    "$.steps.1.result.tds_transactions",
					"sectionMaster":      "$.input.section_master",
					"classificationRules": "$.input.classification_rules",
				},
				TimeoutSeconds:    30,
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
			// Step 3: Validate Deductee Information (PAN, Name, Address)
			{
				StepNumber:    3,
				ServiceName:   "tds",
				HandlerMethod: "ValidateDeducteeInfo",
				InputMapping: map[string]string{
					"classifiedTransactions": "$.steps.2.result.classified_transactions",
					"panRegistry":            "$.input.pan_registry",
					"vendorMaster":           "$.input.vendor_master",
				},
				TimeoutSeconds:    30,
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
			// Step 4: Calculate TDS Deduction Amount - CRITICAL
			{
				StepNumber:    4,
				ServiceName:   "tds",
				HandlerMethod: "CalculateTDSDeduction",
				InputMapping: map[string]string{
					"validatedTransactions": "$.steps.3.result.validated_transactions",
					"tdsRateMaster":         "$.input.tds_rate_master",
					"tdsThresholdRules":     "$.input.tds_threshold_rules",
					"deductionExemptions":   "$.input.deduction_exemptions",
				},
				TimeoutSeconds:    35,
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
			// Step 5: Verify TDS Deposited with Government
			{
				StepNumber:    5,
				ServiceName:   "banking",
				HandlerMethod: "VerifyTDSDeposited",
				InputMapping: map[string]string{
					"tan":                    "$.input.tan",
					"calculatedTDS":          "$.steps.4.result.calculated_tds",
					"bankStatements":         "$.input.bank_statements",
					"depositsVerificationList": "$.input.deposits_verification_list",
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
			// Step 6: Match TDS Deposited vs TDS Deducted
			{
				StepNumber:    6,
				ServiceName:   "tds",
				HandlerMethod: "MatchTDSDepositedVsDeducted",
				InputMapping: map[string]string{
					"calculatedTDS":   "$.steps.4.result.calculated_tds",
					"depositedTDS":    "$.steps.5.result.deposited_tds",
					"reconciliationRules": "$.input.reconciliation_rules",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Generate TDS Return Schedules - CRITICAL
			{
				StepNumber:    7,
				ServiceName:   "tds",
				HandlerMethod: "GenerateTDSReturnSchedules",
				InputMapping: map[string]string{
					"tan":                    "$.input.tan",
					"filingPeriod":           "$.input.filing_period",
					"classifiedTransactions": "$.steps.2.result.classified_transactions",
					"calculatedTDS":          "$.steps.4.result.calculated_tds",
					"matchingResults":        "$.steps.6.result.matching_results",
				},
				TimeoutSeconds:    35,
				IsCritical:        true,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Create Quarterly/Annual TDS Return
			{
				StepNumber:    8,
				ServiceName:   "tds",
				HandlerMethod: "CreateTDSReturn",
				InputMapping: map[string]string{
					"tan":             "$.input.tan",
					"filingPeriod":    "$.input.filing_period",
					"periodType":      "$.input.period_type",
					"schedules":       "$.steps.7.result.tds_schedules",
					"matchingSummary": "$.steps.6.result.matching_summary",
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
			// Step 9: Submit to TDS Clearing House - CRITICAL
			{
				StepNumber:    9,
				ServiceName:   "compliance-postings",
				HandlerMethod: "SubmitToTDSClearingHouse",
				InputMapping: map[string]string{
					"tan":           "$.input.tan",
					"tdsReturn":     "$.steps.8.result.tds_return",
					"filingPeriod":  "$.input.filing_period",
					"dscCertificate": "$.input.dsc_certificate",
				},
				TimeoutSeconds:    45,
				IsCritical:        true,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        5,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Archive TDS Return for Compliance
			{
				StepNumber:    10,
				ServiceName:   "compliance-postings",
				HandlerMethod: "ArchiveTDSReturn",
				InputMapping: map[string]string{
					"tan":               "$.input.tan",
					"filingPeriod":      "$.input.filing_period",
					"tdsReturn":         "$.steps.8.result.tds_return",
					"submissionResponse": "$.steps.9.result.submission_response",
					"schedules":         "$.steps.7.result.tds_schedules",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
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

			// Step 102: Clear Section Classification (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "tds",
				HandlerMethod: "ClearSectionClassification",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 103: Clear Deductee Validation (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "tds",
				HandlerMethod: "ClearDeducteeValidation",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 104: Clear TDS Calculations (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "tds",
				HandlerMethod: "ClearTDSCalculations",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 105: Clear Deposit Verification (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "banking",
				HandlerMethod: "ClearDepositVerification",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: Clear Matching Results (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "tds",
				HandlerMethod: "ClearMatchingResults",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: Delete TDS Return Schedules (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "tds",
				HandlerMethod: "DeleteTDSReturnSchedules",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 108: Delete TDS Return (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "tds",
				HandlerMethod: "DeleteTDSReturn",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 109: Rollback TDS Return Status on Submission Failure (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "compliance-postings",
				HandlerMethod: "RollbackTDSReturnStatus",
				InputMapping: map[string]string{
					"tan":        "$.input.tan",
					"filingPeriod": "$.input.filing_period",
					"reason":     "TDS return submission to clearing house failed",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *TDSReturnFilingSaga) SagaType() string {
	return "SAGA-ST04"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *TDSReturnFilingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *TDSReturnFilingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters for TDS return filing
func (s *TDSReturnFilingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type for TDS return filing")
	}

	// Validate TAN (Tax Account Number, 10 characters)
	if inputMap["tan"] == nil {
		return errors.New("tan is required for TDS return filing")
	}
	tan, ok := inputMap["tan"].(string)
	if !ok || len(tan) != 10 {
		return errors.New("tan must be a 10-character string")
	}

	// Validate filing period format (YYYY-MM or Q1/Q2/Q3/Q4-YYYY for quarterly)
	if inputMap["filing_period"] == nil {
		return errors.New("filing_period is required")
	}
	filingPeriod, ok := inputMap["filing_period"].(string)
	if !ok || len(filingPeriod) == 0 {
		return errors.New("filing_period is required and cannot be empty")
	}

	// Validate period type (QUARTERLY or ANNUAL)
	if inputMap["period_type"] == nil {
		return errors.New("period_type is required (QUARTERLY or ANNUAL)")
	}
	periodType, ok := inputMap["period_type"].(string)
	if !ok || (periodType != "QUARTERLY" && periodType != "ANNUAL") {
		return fmt.Errorf("period_type must be QUARTERLY or ANNUAL, got: %s", periodType)
	}

	// Validate period dates
	if inputMap["period_start_date"] == nil {
		return errors.New("period_start_date is required")
	}
	if inputMap["period_end_date"] == nil {
		return errors.New("period_end_date is required")
	}

	// Validate section master (defines TDS sections like 194C, 194D, etc.)
	if inputMap["section_master"] == nil {
		return errors.New("section_master is required for section classification")
	}

	// Validate classification rules
	if inputMap["classification_rules"] == nil {
		return errors.New("classification_rules are required")
	}

	// Validate PAN registry for deductee validation
	if inputMap["pan_registry"] == nil {
		return errors.New("pan_registry is required for deductee PAN validation")
	}

	// Validate vendor master
	if inputMap["vendor_master"] == nil {
		return errors.New("vendor_master is required for deductee information validation")
	}

	// Validate TDS rate master
	if inputMap["tds_rate_master"] == nil {
		return errors.New("tds_rate_master is required for TDS deduction calculation")
	}

	// Validate TDS threshold rules
	if inputMap["tds_threshold_rules"] == nil {
		return errors.New("tds_threshold_rules are required")
	}

	// Validate deduction exemptions
	if inputMap["deduction_exemptions"] == nil {
		return errors.New("deduction_exemptions are required")
	}

	// Validate bank statements for deposit verification
	if inputMap["bank_statements"] == nil {
		return errors.New("bank_statements are required for TDS deposit verification")
	}

	// Validate deposits verification list
	if inputMap["deposits_verification_list"] == nil {
		return errors.New("deposits_verification_list is required for verifying TDS deposits")
	}

	// Validate reconciliation rules
	if inputMap["reconciliation_rules"] == nil {
		return errors.New("reconciliation_rules are required")
	}

	// Validate DSC certificate
	if inputMap["dsc_certificate"] == nil {
		return errors.New("dsc_certificate is required for TDS return submission")
	}

	return nil
}
