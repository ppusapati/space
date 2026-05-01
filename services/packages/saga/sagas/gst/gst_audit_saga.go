// Package gst provides saga handlers for GST compliance workflows
package gst

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// GSTAuditSaga implements SAGA-G04: GST Audit & Compliance Check workflow
// Business Flow: InitiateAudit → ValidateGSTR1 → ValidateGSTR2 → AnalyzeBilledItems → CheckNegativeITC → GenerateAuditReport → PostAuditFindings → CompleteAudit
// GST Compliance: Comprehensive GST compliance audit and validation
type GSTAuditSaga struct {
	steps []*saga.StepDefinition
}

// NewGSTAuditSaga creates a new GST Audit & Compliance Check saga handler
func NewGSTAuditSaga() saga.SagaHandler {
	return &GSTAuditSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate GST Audit
			{
				StepNumber:    1,
				ServiceName:   "gst-audit",
				HandlerMethod: "InitiateGSTAudit",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"auditPeriod":   "$.input.audit_period",
					"auditType":     "$.input.audit_type",
					"gstin":         "$.input.gstin",
					"fiscalYear":    "$.input.fiscal_year",
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
			// Step 2: Validate GSTR-1 Returns
			{
				StepNumber:    2,
				ServiceName:   "gst-ledger",
				HandlerMethod: "ValidateGSTR1Returns",
				InputMapping: map[string]string{
					"auditID":     "$.steps.1.result.audit_id",
					"auditPeriod": "$.input.audit_period",
					"gstin":       "$.input.gstin",
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
			// Step 3: Validate GSTR-2 Returns
			{
				StepNumber:    3,
				ServiceName:   "gst-ledger",
				HandlerMethod: "ValidateGSTR2Returns",
				InputMapping: map[string]string{
					"auditID":     "$.steps.1.result.audit_id",
					"auditPeriod": "$.input.audit_period",
					"gstin":       "$.input.gstin",
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
			// Step 4: Analyze Billed Items
			{
				StepNumber:    4,
				ServiceName:   "general-ledger",
				HandlerMethod: "AnalyzeGSTBilledItems",
				InputMapping: map[string]string{
					"auditID":          "$.steps.1.result.audit_id",
					"gstr1Validation":  "$.steps.2.result.validation_result",
					"auditPeriod":      "$.input.audit_period",
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
			// Step 5: Check Negative ITC & Compliance Rules
			{
				StepNumber:    5,
				ServiceName:   "accounts-payable",
				HandlerMethod: "CheckNegativeITCCompliance",
				InputMapping: map[string]string{
					"auditID":           "$.steps.1.result.audit_id",
					"gstr2Validation":   "$.steps.3.result.validation_result",
					"billedItemsAnalysis": "$.steps.4.result.analysis_result",
					"auditType":         "$.input.audit_type",
				},
				TimeoutSeconds:    30,
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
			// Step 6: Verify Accounts Receivable Compliance
			{
				StepNumber:    6,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "VerifyGSTARCompliance",
				InputMapping: map[string]string{
					"auditID":    "$.steps.1.result.audit_id",
					"auditPeriod": "$.input.audit_period",
					"gstin":      "$.input.gstin",
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
			// Step 7: Generate Audit Report
			{
				StepNumber:    7,
				ServiceName:   "gst-audit",
				HandlerMethod: "GenerateAuditReport",
				InputMapping: map[string]string{
					"auditID":              "$.steps.1.result.audit_id",
					"gstr1Validation":      "$.steps.2.result.validation_result",
					"gstr2Validation":      "$.steps.3.result.validation_result",
					"billedItemsAnalysis":  "$.steps.4.result.analysis_result",
					"negativeITCCheck":     "$.steps.5.result.compliance_check_result",
					"arCompliance":         "$.steps.6.result.compliance_result",
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
			// Step 8: Post Audit Findings to Ledger
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostAuditFindingsJournal",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"auditID":        "$.steps.1.result.audit_id",
					"auditReport":    "$.steps.7.result.audit_report",
					"journalDate":    "$.input.audit_period",
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
			// Step 9: Generate Compliance Certificate
			{
				StepNumber:    9,
				ServiceName:   "audit",
				HandlerMethod: "GenerateGSTComplianceCertificate",
				InputMapping: map[string]string{
					"auditID":     "$.steps.1.result.audit_id",
					"auditReport": "$.steps.7.result.audit_report",
					"auditPeriod": "$.input.audit_period",
					"gstin":       "$.input.gstin",
				},
				TimeoutSeconds:    20,
				IsCritical:        false,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Complete GST Audit
			{
				StepNumber:    10,
				ServiceName:   "gst-audit",
				HandlerMethod: "CompleteGSTAudit",
				InputMapping: map[string]string{
					"auditID":               "$.steps.1.result.audit_id",
					"auditReport":           "$.steps.7.result.audit_report",
					"complianceCertificate": "$.steps.9.result.compliance_certificate",
					"completionDate":        "$.input.audit_period",
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

			// Step 101: Cancel Audit Initiation (compensates step 1)
			{
				StepNumber:    101,
				ServiceName:   "gst-audit",
				HandlerMethod: "CancelAuditInitiation",
				InputMapping: map[string]string{
					"auditID": "$.steps.1.result.audit_id",
					"reason":  "Saga compensation - GST audit failed",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 102: Clear GSTR-1 Validation (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "gst-ledger",
				HandlerMethod: "ClearGSTR1Validation",
				InputMapping: map[string]string{
					"auditID": "$.steps.1.result.audit_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: Clear GSTR-2 Validation (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "gst-ledger",
				HandlerMethod: "ClearGSTR2Validation",
				InputMapping: map[string]string{
					"auditID": "$.steps.1.result.audit_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 104: Clear Billed Items Analysis (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "general-ledger",
				HandlerMethod: "ClearBilledItemsAnalysis",
				InputMapping: map[string]string{
					"auditID": "$.steps.1.result.audit_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: Clear ITC Compliance Check (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ClearITCComplianceCheck",
				InputMapping: map[string]string{
					"auditID": "$.steps.1.result.audit_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Clear AR Compliance Verification (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ClearARComplianceVerification",
				InputMapping: map[string]string{
					"auditID": "$.steps.1.result.audit_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 107: Delete Audit Report (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "gst-audit",
				HandlerMethod: "DeleteAuditReport",
				InputMapping: map[string]string{
					"auditID": "$.steps.1.result.audit_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 108: Reverse Audit Journal (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseAuditFindingsJournal",
				InputMapping: map[string]string{
					"auditID": "$.steps.1.result.audit_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 109: Delete Compliance Certificate (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "audit",
				HandlerMethod: "DeleteComplianceCertificate",
				InputMapping: map[string]string{
					"auditID": "$.steps.1.result.audit_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *GSTAuditSaga) SagaType() string {
	return "SAGA-G04"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *GSTAuditSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *GSTAuditSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *GSTAuditSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["audit_period"] == nil {
		return errors.New("audit_period is required (format: YYYY-MM)")
	}

	if inputMap["audit_type"] == nil {
		return errors.New("audit_type is required (ROUTINE, SPECIAL, SYSTEM, etc.)")
	}

	auditType, ok := inputMap["audit_type"].(string)
	if !ok {
		return errors.New("audit_type must be a string")
	}

	validTypes := map[string]bool{
		"ROUTINE":      true,
		"SPECIAL":      true,
		"SYSTEM":       true,
		"TRANSACTION":  true,
		"COMPLIANCE":   true,
	}

	if !validTypes[auditType] {
		return errors.New("audit_type must be ROUTINE, SPECIAL, SYSTEM, TRANSACTION, or COMPLIANCE")
	}

	if inputMap["gstin"] == nil {
		return errors.New("gstin is required")
	}

	if inputMap["fiscal_year"] == nil {
		return errors.New("fiscal_year is required")
	}

	return nil
}
