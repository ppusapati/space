// Package statutory provides saga handlers for statutory compliance workflows
package statutory

import (
	"errors"
	"fmt"

	"p9e.in/samavaya/packages/saga"
)

// GSTR1FilingSaga implements SAGA-ST01: GSTR-1 Filing (Sales Tax Return) workflow
// Business Flow: ExtractSalesInvoices → ClassifyByType → ValidateHSNSAC → CalculateTaxByRate →
//               GenerateSchedules → ApplyAmendments → CalculateTaxLiability → GenerateXML →
//               SubmitToGSTN → UpdateComplianceStatus → ArchiveGSTR1
// GSTR-1 Compliance: GST Return for outward supplies, filed monthly by 20th of next month
// Critical Steps: 5 (GenerateSchedules), 8 (GenerateXML), 9 (SubmitToGSTN)
type GSTR1FilingSaga struct {
	steps []*saga.StepDefinition
}

// NewGSTR1FilingSaga creates a new GSTR-1 Filing saga handler
func NewGSTR1FilingSaga() saga.SagaHandler {
	return &GSTR1FilingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Extract Sales Invoices (Period-based)
			{
				StepNumber:    1,
				ServiceName:   "gst",
				HandlerMethod: "ExtractSalesInvoices",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"gstin":           "$.input.gstin",
					"filingPeriod":    "$.input.filing_period",
					"periodStartDate": "$.input.period_start_date",
					"periodEndDate":   "$.input.period_end_date",
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
			// Step 2: Classify Invoices by Transaction Type (B2B, B2C, export, nil-rated)
			{
				StepNumber:    2,
				ServiceName:   "gst",
				HandlerMethod: "ClassifyInvoicesByType",
				InputMapping: map[string]string{
					"salesInvoices":     "$.steps.1.result.sales_invoices",
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
			// Step 3: Validate HSN/SAC Codes and Description
			{
				StepNumber:    3,
				ServiceName:   "gst",
				HandlerMethod: "ValidateHSNSACCodes",
				InputMapping: map[string]string{
					"classifiedInvoices": "$.steps.2.result.classified_invoices",
					"hsnMaster":          "$.input.hsn_master",
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
			// Step 4: Calculate SGST/CGST/IGST by Tax Rate
			{
				StepNumber:    4,
				ServiceName:   "gst",
				HandlerMethod: "CalculateTaxByRate",
				InputMapping: map[string]string{
					"validatedInvoices": "$.steps.3.result.validated_invoices",
					"taxRateMaster":     "$.input.tax_rate_master",
				},
				TimeoutSeconds:    30,
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
			// Step 5: Generate GSTR-1 Schedules (B2B, B2C, export, etc.) - CRITICAL
			{
				StepNumber:    5,
				ServiceName:   "gst",
				HandlerMethod: "GenerateGSTR1Schedules",
				InputMapping: map[string]string{
					"taxCalculatedInvoices": "$.steps.4.result.tax_calculated_invoices",
					"gstin":                 "$.input.gstin",
					"filingPeriod":          "$.input.filing_period",
				},
				TimeoutSeconds:    35,
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
			// Step 6: Apply Amendments and Credit Notes
			{
				StepNumber:    6,
				ServiceName:   "gst",
				HandlerMethod: "ApplyAmendmentsAndCredits",
				InputMapping: map[string]string{
					"gstr1Schedules":    "$.steps.5.result.gstr1_schedules",
					"creditNotes":       "$.input.credit_notes",
					"amendments":        "$.input.amendments",
					"filingPeriod":      "$.input.filing_period",
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
			// Step 7: Calculate Total Tax Liability
			{
				StepNumber:    7,
				ServiceName:   "gst",
				HandlerMethod: "CalculateTotalTaxLiability",
				InputMapping: map[string]string{
					"adjustedSchedules": "$.steps.6.result.adjusted_schedules",
					"previousReturns":   "$.input.previous_returns",
				},
				TimeoutSeconds:    25,
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
			// Step 8: Generate GSTR-1 XML Format - CRITICAL
			{
				StepNumber:    8,
				ServiceName:   "gst",
				HandlerMethod: "GenerateGSTR1XML",
				InputMapping: map[string]string{
					"gstin":              "$.input.gstin",
					"returnPeriod":       "$.input.filing_period",
					"schedules":          "$.steps.6.result.adjusted_schedules",
					"taxLiability":       "$.steps.7.result.tax_liability",
					"declarantDetails":   "$.input.declarant_details",
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
			// Step 9: Submit to GSTN (API Call) - CRITICAL (Higher timeout for external API)
			{
				StepNumber:    9,
				ServiceName:   "compliance-postings",
				HandlerMethod: "SubmitToGSTN",
				InputMapping: map[string]string{
					"gstrXML":       "$.steps.8.result.gstr_xml",
					"gstin":         "$.input.gstin",
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
			// Step 10: Update Compliance Status
			{
				StepNumber:    10,
				ServiceName:   "gst",
				HandlerMethod: "UpdateComplianceStatus",
				InputMapping: map[string]string{
					"gstin":              "$.input.gstin",
					"filingPeriod":       "$.input.filing_period",
					"submissionResponse": "$.steps.9.result.submission_response",
					"status":             "FILED",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{110},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 11: Archive GSTR-1 for Audit
			{
				StepNumber:    11,
				ServiceName:   "compliance-postings",
				HandlerMethod: "ArchiveGSTR1",
				InputMapping: map[string]string{
					"gstin":              "$.input.gstin",
					"filingPeriod":       "$.input.filing_period",
					"gstrXML":            "$.steps.8.result.gstr_xml",
					"submissionResponse": "$.steps.9.result.submission_response",
					"schedules":          "$.steps.6.result.adjusted_schedules",
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

			// Step 102: Clear Invoice Classification (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "gst",
				HandlerMethod: "ClearInvoiceClassification",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 103: Clear Validation Records (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "gst",
				HandlerMethod: "ClearValidationRecords",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 104: Clear Tax Calculations (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "gst",
				HandlerMethod: "ClearTaxCalculations",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 105: Delete GSTR-1 Schedules (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "gst",
				HandlerMethod: "DeleteGSTR1Schedules",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Clear Amendments (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "gst",
				HandlerMethod: "ClearAmendments",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: Clear Tax Liability Calculation (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "gst",
				HandlerMethod: "ClearTaxLiabilityCalculation",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 108: Delete GSTR-1 XML (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "gst",
				HandlerMethod: "DeleteGSTR1XML",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 109: Rollback GSTR-1 Status on Submission Failure (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "compliance-postings",
				HandlerMethod: "RollbackGSTR1Status",
				InputMapping: map[string]string{
					"gstin":        "$.input.gstin",
					"filingPeriod": "$.input.filing_period",
					"reason":       "GSTN submission failed",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 110: Revert Compliance Status (compensates step 10)
			{
				StepNumber:    110,
				ServiceName:   "gst",
				HandlerMethod: "RevertComplianceStatus",
				InputMapping: map[string]string{
					"gstin":        "$.input.gstin",
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *GSTR1FilingSaga) SagaType() string {
	return "SAGA-ST01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *GSTR1FilingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *GSTR1FilingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters for GSTR-1 filing
func (s *GSTR1FilingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type for GSTR-1 filing")
	}

	// Validate GSTIN format (15 characters, alphanumeric)
	if inputMap["gstin"] == nil {
		return errors.New("gstin is required for GSTR-1 filing")
	}
	gstin, ok := inputMap["gstin"].(string)
	if !ok || len(gstin) != 15 {
		return errors.New("gstin must be a 15-character string")
	}

	// Validate filing period
	if inputMap["filing_period"] == nil {
		return errors.New("filing_period is required (format: YYYY-MM)")
	}
	filingPeriod, ok := inputMap["filing_period"].(string)
	if !ok || len(filingPeriod) != 7 {
		return fmt.Errorf("filing_period must be in YYYY-MM format, got: %s", filingPeriod)
	}

	// Validate period dates
	if inputMap["period_start_date"] == nil {
		return errors.New("period_start_date is required")
	}
	if inputMap["period_end_date"] == nil {
		return errors.New("period_end_date is required")
	}

	// Validate classification rules
	if inputMap["classification_rules"] == nil {
		return errors.New("classification_rules are required (B2B, B2C, export, nil-rated)")
	}

	// Validate HSN/SAC master
	if inputMap["hsn_master"] == nil {
		return errors.New("hsn_master is required for HSN/SAC validation")
	}

	// Validate tax rate master
	if inputMap["tax_rate_master"] == nil {
		return errors.New("tax_rate_master is required for tax calculations")
	}

	// Validate declarant details
	if inputMap["declarant_details"] == nil {
		return errors.New("declarant_details are required (name, designation, etc.)")
	}

	// Validate DSC certificate (Digital Signature Certificate)
	if inputMap["dsc_certificate"] == nil {
		return errors.New("dsc_certificate is required for GSTN submission")
	}

	return nil
}
