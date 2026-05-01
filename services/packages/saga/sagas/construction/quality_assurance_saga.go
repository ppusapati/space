// Package construction provides saga handlers for construction module workflows
package construction

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// QualityAssuranceSaga implements SAGA-C05: Quality Assurance & Inspection Management
// Business Flow: ScheduleQualityInspection → PrepareInspectionChecklist → ConductFieldInspection → DocumentInspectionFindings → AnalyzeQualityMetrics → GenerateInspectionReport → ApproveInspectionReport → ArchiveInspectionData → NotifyStakeholders
// Timeout: 120 seconds, Critical steps: 1,2,3,4,6,9
type QualityAssuranceSaga struct {
	steps []*saga.StepDefinition
}

// NewQualityAssuranceSaga creates a new Quality Assurance & Inspection Management saga handler
func NewQualityAssuranceSaga() saga.SagaHandler {
	return &QualityAssuranceSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Schedule Quality Inspection
			{
				StepNumber:    1,
				ServiceName:   "quality-inspection",
				HandlerMethod: "ScheduleQualityInspection",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"inspectionID":          "$.input.inspection_id",
					"inspectionType":        "$.input.inspection_type",
					"locationDescription":   "$.input.location_description",
					"scheduledDate":         "$.input.scheduled_date",
				},
				TimeoutSeconds: 30,
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
			// Step 2: Prepare Inspection Checklist
			{
				StepNumber:    2,
				ServiceName:   "quality-inspection",
				HandlerMethod: "PrepareInspectionChecklist",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"inspectionID":        "$.input.inspection_id",
					"inspectionType":      "$.input.inspection_type",
					"locationDescription": "$.input.location_description",
					"schedule":            "$.steps.1.result.inspection_schedule",
				},
				TimeoutSeconds:    25,
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
			// Step 3: Conduct Field Inspection
			{
				StepNumber:    3,
				ServiceName:   "construction-site",
				HandlerMethod: "ConductFieldInspection",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"inspectionID":        "$.input.inspection_id",
					"inspectionType":      "$.input.inspection_type",
					"locationDescription": "$.input.location_description",
					"checklist":           "$.steps.2.result.inspection_checklist",
				},
				TimeoutSeconds:    45,
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
			// Step 4: Document Inspection Findings
			{
				StepNumber:    4,
				ServiceName:   "quality-inspection",
				HandlerMethod: "DocumentInspectionFindings",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"inspectionID":        "$.input.inspection_id",
					"locationDescription": "$.input.location_description",
					"fieldData":           "$.steps.3.result.field_data",
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
			// Step 5: Analyze Quality Metrics
			{
				StepNumber:    5,
				ServiceName:   "project-planning",
				HandlerMethod: "AnalyzeQualityMetrics",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"inspectionID":   "$.input.inspection_id",
					"findings":       "$.steps.4.result.inspection_findings",
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
			// Step 6: Generate Inspection Report
			{
				StepNumber:    6,
				ServiceName:   "quality-inspection",
				HandlerMethod: "GenerateInspectionReport",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"inspectionID":     "$.input.inspection_id",
					"inspectionType":   "$.input.inspection_type",
					"findings":         "$.steps.4.result.inspection_findings",
					"qualityMetrics":   "$.steps.5.result.quality_metrics",
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
			// Step 7: Approve Inspection Report
			{
				StepNumber:    7,
				ServiceName:   "approval",
				HandlerMethod: "ApproveInspectionReport",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"inspectionID":        "$.input.inspection_id",
					"inspectionType":      "$.input.inspection_type",
					"inspectionReport":    "$.steps.6.result.inspection_report",
				},
				TimeoutSeconds:    30,
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
			// Step 8: Archive Inspection Data
			{
				StepNumber:    8,
				ServiceName:   "project-planning",
				HandlerMethod: "ArchiveInspectionData",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"inspectionID":     "$.input.inspection_id",
					"inspectionReport": "$.steps.6.result.inspection_report",
					"approval":         "$.steps.7.result.approval_status",
				},
				TimeoutSeconds:    25,
				IsCritical:        false,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Notify Stakeholders
			{
				StepNumber:    9,
				ServiceName:   "quality-inspection",
				HandlerMethod: "NotifyStakeholders",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"inspectionID":     "$.input.inspection_id",
					"inspectionType":   "$.input.inspection_type",
					"locationDescription": "$.input.location_description",
					"inspectionReport": "$.steps.6.result.inspection_report",
					"approval":         "$.steps.7.result.approval_status",
				},
				TimeoutSeconds: 30,
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

			// Step 102: CancelInspectionChecklist (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "quality-inspection",
				HandlerMethod: "CancelInspectionChecklist",
				InputMapping: map[string]string{
					"inspectionID":        "$.input.inspection_id",
					"inspectionChecklist": "$.steps.2.result.inspection_checklist",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: CancelFieldInspection (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "construction-site",
				HandlerMethod: "CancelFieldInspection",
				InputMapping: map[string]string{
					"inspectionID":  "$.input.inspection_id",
					"fieldData":     "$.steps.3.result.field_data",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 104: ClearInspectionFindings (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "quality-inspection",
				HandlerMethod: "ClearInspectionFindings",
				InputMapping: map[string]string{
					"inspectionID":        "$.input.inspection_id",
					"inspectionFindings": "$.steps.4.result.inspection_findings",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: ReverseQualityAnalysis (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "project-planning",
				HandlerMethod: "ReverseQualityAnalysis",
				InputMapping: map[string]string{
					"inspectionID":   "$.input.inspection_id",
					"qualityMetrics": "$.steps.5.result.quality_metrics",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: DeleteInspectionReport (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "quality-inspection",
				HandlerMethod: "DeleteInspectionReport",
				InputMapping: map[string]string{
					"inspectionID":     "$.input.inspection_id",
					"inspectionReport": "$.steps.6.result.inspection_report",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: RevokeReportApproval (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "approval",
				HandlerMethod: "RevokeReportApproval",
				InputMapping: map[string]string{
					"inspectionID":    "$.input.inspection_id",
					"approvalStatus":  "$.steps.7.result.approval_status",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: UnarchiveInspectionData (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "project-planning",
				HandlerMethod: "UnarchiveInspectionData",
				InputMapping: map[string]string{
					"inspectionID":     "$.input.inspection_id",
					"archivedData":     "$.steps.8.result.archived_data",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *QualityAssuranceSaga) SagaType() string {
	return "SAGA-C05"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *QualityAssuranceSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *QualityAssuranceSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *QualityAssuranceSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["inspection_id"] == nil {
		return errors.New("inspection_id is required")
	}

	inspectionID, ok := inputMap["inspection_id"].(string)
	if !ok || inspectionID == "" {
		return errors.New("inspection_id must be a non-empty string")
	}

	if inputMap["inspection_type"] == nil {
		return errors.New("inspection_type is required")
	}

	inspectionType, ok := inputMap["inspection_type"].(string)
	if !ok || inspectionType == "" {
		return errors.New("inspection_type must be a non-empty string")
	}

	if inputMap["location_description"] == nil {
		return errors.New("location_description is required")
	}

	locationDescription, ok := inputMap["location_description"].(string)
	if !ok || locationDescription == "" {
		return errors.New("location_description must be a non-empty string")
	}

	if inputMap["scheduled_date"] == nil {
		return errors.New("scheduled_date is required")
	}

	scheduledDate, ok := inputMap["scheduled_date"].(string)
	if !ok || scheduledDate == "" {
		return errors.New("scheduled_date must be a non-empty string")
	}

	return nil
}
