// Package agriculture provides saga handlers for agricultural workflows
package agriculture

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ComplianceCertificationSaga implements SAGA-A07: Agricultural Compliance & Certification workflow
// Business Flow: InitiateCertification → ValidateFarmCompliance → PerformAudit → AssessQualityStandards → GenerateCertificationReport → UpdateComplianceStatus → PostCertificationJournal → ArchiveCertificationRecord
// Steps: 8 forward + 7 compensation = 15 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,7,8
type ComplianceCertificationSaga struct {
	steps []*saga.StepDefinition
}

// NewComplianceCertificationSaga creates a new Agricultural Compliance & Certification saga handler
func NewComplianceCertificationSaga() saga.SagaHandler {
	return &ComplianceCertificationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Certification
			{
				StepNumber:    1,
				ServiceName:   "agriculture",
				HandlerMethod: "InitiateCertification",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"certificationID":   "$.input.certification_id",
					"farmID":            "$.input.farm_id",
					"certificationType": "$.input.certification_type",
					"auditDate":         "$.input.audit_date",
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
			// Step 2: Validate Farm Compliance
			{
				StepNumber:    2,
				ServiceName:   "compliance",
				HandlerMethod: "ValidateFarmCompliance",
				InputMapping: map[string]string{
					"certificationID":   "$.steps.1.result.certification_id",
					"farmID":            "$.input.farm_id",
					"certificationType": "$.input.certification_type",
					"validateRules":     "true",
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
			// Step 3: Perform Audit
			{
				StepNumber:    3,
				ServiceName:   "audit",
				HandlerMethod: "PerformFarmAudit",
				InputMapping: map[string]string{
					"certificationID":   "$.steps.1.result.certification_id",
					"farmID":            "$.input.farm_id",
					"complianceStatus":  "$.steps.2.result.compliance_status",
					"auditDate":         "$.input.audit_date",
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
			// Step 4: Assess Quality Standards
			{
				StepNumber:    4,
				ServiceName:   "quality-inspection",
				HandlerMethod: "AssessQualityStandards",
				InputMapping: map[string]string{
					"certificationID":   "$.steps.1.result.certification_id",
					"farmID":            "$.input.farm_id",
					"certificationType": "$.input.certification_type",
					"auditResult":       "$.steps.3.result.audit_result",
				},
				TimeoutSeconds:    35,
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
			// Step 5: Generate Certification Report
			{
				StepNumber:    5,
				ServiceName:   "certification",
				HandlerMethod: "GenerateCertificationReport",
				InputMapping: map[string]string{
					"certificationID":   "$.steps.1.result.certification_id",
					"auditResult":       "$.steps.3.result.audit_result",
					"qualityAssessment": "$.steps.4.result.quality_assessment",
				},
				TimeoutSeconds:    30,
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
			// Step 6: Update Compliance Status
			{
				StepNumber:    6,
				ServiceName:   "compliance",
				HandlerMethod: "UpdateComplianceStatus",
				InputMapping: map[string]string{
					"certificationID":      "$.steps.1.result.certification_id",
					"farmID":               "$.input.farm_id",
					"certificationReport":  "$.steps.5.result.certification_report",
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
			// Step 7: Post Certification Journal Entries
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "ApplyCertificationJournal",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"certificationID":   "$.steps.1.result.certification_id",
					"certificationType": "$.input.certification_type",
					"journalDate":       "$.input.audit_date",
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
			// Step 8: Archive Certification Record
			{
				StepNumber:    8,
				ServiceName:   "regulatory-reporting",
				HandlerMethod: "ArchiveCertificationRecord",
				InputMapping: map[string]string{
					"certificationID":      "$.steps.1.result.certification_id",
					"certificationReport":  "$.steps.5.result.certification_report",
					"completionStatus":     "Completed",
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

			// Step 101: Revert Farm Compliance Validation (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "compliance",
				HandlerMethod: "RevertFarmComplianceValidation",
				InputMapping: map[string]string{
					"certificationID": "$.steps.1.result.certification_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Revert Audit (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "audit",
				HandlerMethod: "RevertFarmAudit",
				InputMapping: map[string]string{
					"certificationID": "$.steps.1.result.certification_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
			// Step 103: Revert Quality Standards Assessment (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "quality-inspection",
				HandlerMethod: "RevertQualityStandardsAssessment",
				InputMapping: map[string]string{
					"certificationID": "$.steps.1.result.certification_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 104: Delete Certification Report (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "certification",
				HandlerMethod: "DeleteCertificationReport",
				InputMapping: map[string]string{
					"certificationID": "$.steps.1.result.certification_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: Revert Compliance Status Update (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "compliance",
				HandlerMethod: "RevertComplianceStatusUpdate",
				InputMapping: map[string]string{
					"certificationID": "$.steps.1.result.certification_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Reverse Certification Journal (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseCertificationJournal",
				InputMapping: map[string]string{
					"certificationID": "$.steps.1.result.certification_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Cancel Certification Initiation (compensates step 1)
			{
				StepNumber:    107,
				ServiceName:   "agriculture",
				HandlerMethod: "CancelCertification",
				InputMapping: map[string]string{
					"certificationID": "$.steps.1.result.certification_id",
					"reason":          "Saga compensation - Certification process failed",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ComplianceCertificationSaga) SagaType() string {
	return "SAGA-A07"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ComplianceCertificationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ComplianceCertificationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ComplianceCertificationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["certification_id"] == nil {
		return errors.New("certification_id is required")
	}

	if inputMap["farm_id"] == nil {
		return errors.New("farm_id is required")
	}

	if inputMap["certification_type"] == nil {
		return errors.New("certification_type is required")
	}

	if inputMap["audit_date"] == nil {
		return errors.New("audit_date is required (format: YYYY-MM-DD)")
	}

	return nil
}
