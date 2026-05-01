// Package healthcare provides saga handlers for healthcare workflows
package healthcare

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ComplianceSaga implements SAGA-HC05: Healthcare Compliance & Regulatory Reporting workflow
// Business Flow: InitiateComplianceCheck → ValidateCheckScope → ReviewPatientRecords → EvaluateCompliance → IdentifyViolations → PrepareComplianceReport → SubmitRegulatoryReport → TrackComplianceAction → UpdateComplianceStatus → CompleteComplianceCheck
// Steps: 9 forward + 8 compensation = 17 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,6,9
type ComplianceSaga struct {
	steps []*saga.StepDefinition
}

// NewComplianceSaga creates a new Healthcare Compliance saga handler
func NewComplianceSaga() saga.SagaHandler {
	return &ComplianceSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Compliance Check
			{
				StepNumber:    1,
				ServiceName:   "compliance",
				HandlerMethod: "InitiateComplianceCheck",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"complianceCheckID":  "$.input.compliance_check_id",
					"checkType":          "$.input.check_type",
					"periodStart":        "$.input.period_start",
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
			// Step 2: Validate Check Scope
			{
				StepNumber:    2,
				ServiceName:   "compliance",
				HandlerMethod: "ValidateCheckScope",
				InputMapping: map[string]string{
					"complianceCheckID": "$.steps.1.result.compliance_check_id",
					"checkType":         "$.input.check_type",
					"periodStart":       "$.input.period_start",
					"validateRules":     "true",
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
			// Step 3: Review Patient Records
			{
				StepNumber:    3,
				ServiceName:   "patient-management",
				HandlerMethod: "ReviewPatientRecords",
				InputMapping: map[string]string{
					"complianceCheckID": "$.steps.1.result.compliance_check_id",
					"checkType":         "$.input.check_type",
					"periodStart":       "$.input.period_start",
					"scopeData":         "$.steps.2.result.scope_data",
				},
				TimeoutSeconds:    45,
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
			// Step 4: Evaluate Compliance
			{
				StepNumber:    4,
				ServiceName:   "compliance",
				HandlerMethod: "EvaluateCompliance",
				InputMapping: map[string]string{
					"complianceCheckID": "$.steps.1.result.compliance_check_id",
					"checkType":         "$.input.check_type",
					"patientReview":     "$.steps.3.result.patient_review",
					"scopeData":         "$.steps.2.result.scope_data",
				},
				TimeoutSeconds:    40,
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
			// Step 5: Identify Violations
			{
				StepNumber:    5,
				ServiceName:   "compliance",
				HandlerMethod: "IdentifyViolations",
				InputMapping: map[string]string{
					"complianceCheckID":  "$.steps.1.result.compliance_check_id",
					"complianceResult":   "$.steps.4.result.compliance_result",
					"patientReview":      "$.steps.3.result.patient_review",
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
			// Step 6: Prepare Compliance Report
			{
				StepNumber:    6,
				ServiceName:   "regulatory-reporting",
				HandlerMethod: "PrepareComplianceReport",
				InputMapping: map[string]string{
					"complianceCheckID": "$.steps.1.result.compliance_check_id",
					"checkType":         "$.input.check_type",
					"complianceResult":  "$.steps.4.result.compliance_result",
					"violations":        "$.steps.5.result.violations",
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
			// Step 7: Submit Regulatory Report
			{
				StepNumber:    7,
				ServiceName:   "regulatory-reporting",
				HandlerMethod: "SubmitRegulatoryReport",
				InputMapping: map[string]string{
					"complianceCheckID": "$.steps.1.result.compliance_check_id",
					"checkType":         "$.input.check_type",
					"complianceReport":  "$.steps.6.result.compliance_report",
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
			// Step 8: Track Compliance Action
			{
				StepNumber:    8,
				ServiceName:   "audit",
				HandlerMethod: "TrackComplianceAction",
				InputMapping: map[string]string{
					"complianceCheckID": "$.steps.1.result.compliance_check_id",
					"checkType":         "$.input.check_type",
					"violations":        "$.steps.5.result.violations",
					"complianceReport":  "$.steps.6.result.compliance_report",
				},
				TimeoutSeconds:    25,
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
			// Step 9: Complete Compliance Check
			{
				StepNumber:    9,
				ServiceName:   "compliance",
				HandlerMethod: "CompleteComplianceCheck",
				InputMapping: map[string]string{
					"complianceCheckID":  "$.steps.1.result.compliance_check_id",
					"checkType":          "$.input.check_type",
					"complianceResult":   "$.steps.4.result.compliance_result",
					"complianceReport":   "$.steps.6.result.compliance_report",
					"completionStatus":   "Completed",
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

			// Step 101: Revert Patient Records Review (compensates step 3)
			{
				StepNumber:    101,
				ServiceName:   "patient-management",
				HandlerMethod: "RevertPatientRecordsReview",
				InputMapping: map[string]string{
					"complianceCheckID": "$.steps.1.result.compliance_check_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 102: Revert Compliance Evaluation (compensates step 4)
			{
				StepNumber:    102,
				ServiceName:   "compliance",
				HandlerMethod: "RevertComplianceEvaluation",
				InputMapping: map[string]string{
					"complianceCheckID": "$.steps.1.result.compliance_check_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
			// Step 103: Clear Violation Identification (compensates step 5)
			{
				StepNumber:    103,
				ServiceName:   "compliance",
				HandlerMethod: "ClearViolationIdentification",
				InputMapping: map[string]string{
					"complianceCheckID": "$.steps.1.result.compliance_check_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Withdraw Compliance Report (compensates step 6)
			{
				StepNumber:    104,
				ServiceName:   "regulatory-reporting",
				HandlerMethod: "WithdrawComplianceReport",
				InputMapping: map[string]string{
					"complianceCheckID": "$.steps.1.result.compliance_check_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 105: Retract Regulatory Report (compensates step 7)
			{
				StepNumber:    105,
				ServiceName:   "regulatory-reporting",
				HandlerMethod: "RetractRegulatoryReport",
				InputMapping: map[string]string{
					"complianceCheckID": "$.steps.1.result.compliance_check_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: Revert Compliance Tracking (compensates step 8)
			{
				StepNumber:    106,
				ServiceName:   "audit",
				HandlerMethod: "RevertComplianceTracking",
				InputMapping: map[string]string{
					"complianceCheckID": "$.steps.1.result.compliance_check_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ComplianceSaga) SagaType() string {
	return "SAGA-HC05"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ComplianceSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ComplianceSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ComplianceSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["compliance_check_id"] == nil {
		return errors.New("compliance_check_id is required")
	}

	if inputMap["check_type"] == nil {
		return errors.New("check_type is required")
	}

	if inputMap["period_start"] == nil {
		return errors.New("period_start is required (format: YYYY-MM-DD)")
	}

	return nil
}
