// Package workflow provides saga handlers for workflow management
package workflow

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// MultiLevelApprovalRoutingSaga implements SAGA-WF01: Multi-Level Approval Routing workflow
// Business Flow: SubmitDocument → DetermineApprovalHierarchy → RouteToLevel1 → MonitorLevel1 →
//               RouteToLevel2 → MonitorLevel2 → RouteToLevel3 → CollectApprovals →
//               UpdateDocumentStatus → TriggerNextStep
// Steps: 10 forward + 9 compensation = 19 total
// Timeout: 600 seconds, Critical steps: 2,8,9
type MultiLevelApprovalRoutingSaga struct {
	steps []*saga.StepDefinition
}

// NewMultiLevelApprovalRoutingSaga creates a new Multi-Level Approval Routing saga handler
func NewMultiLevelApprovalRoutingSaga() saga.SagaHandler {
	return &MultiLevelApprovalRoutingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Submit Document for Approval
			{
				StepNumber:    1,
				ServiceName:   "workflow",
				HandlerMethod: "SubmitDocumentForApproval",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"documentID":        "$.input.document_id",
					"documentType":      "$.input.document_type",
					"amount":            "$.input.amount",
					"submitterID":       "$.input.submitter_id",
					"submissionDate":    "$.input.submission_date",
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
			// Step 2: Determine Approval Hierarchy (CRITICAL)
			{
				StepNumber:    2,
				ServiceName:   "approval",
				HandlerMethod: "DetermineApprovalHierarchy",
				InputMapping: map[string]string{
					"documentID":   "$.steps.1.result.document_id",
					"documentType": "$.input.document_type",
					"amount":       "$.input.amount",
					"department":   "$.input.department",
					"roleID":       "$.input.role_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     true,
				CompensationSteps: []int32{110},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Route to Level-1 Approver
			{
				StepNumber:    3,
				ServiceName:   "workflow",
				HandlerMethod: "RouteToLevel1Approver",
				InputMapping: map[string]string{
					"documentID":          "$.steps.1.result.document_id",
					"approvalLevel":       "1",
					"approverID":          "$.steps.2.result.level_1_approver_id",
					"hierarchyDetails":    "$.steps.2.result.hierarchy_details",
					"routingDate":         "$.steps.1.result.submission_date",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
				CompensationSteps: []int32{111},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Monitor Level-1 Approval (with timeout escalation)
			{
				StepNumber:    4,
				ServiceName:   "approval",
				HandlerMethod: "MonitorLevel1Approval",
				InputMapping: map[string]string{
					"documentID":      "$.steps.1.result.document_id",
					"approvalLevel":   "1",
					"approverID":      "$.steps.2.result.level_1_approver_id",
					"escalationRules": "$.steps.2.result.escalation_rules",
					"timeoutMinutes":  "15",
				},
				TimeoutSeconds: 900,
				IsCritical:     false,
				CompensationSteps: []int32{112},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Route to Level-2 Approver (if required)
			{
				StepNumber:    5,
				ServiceName:   "workflow",
				HandlerMethod: "RouteToLevel2Approver",
				InputMapping: map[string]string{
					"documentID":          "$.steps.1.result.document_id",
					"approvalLevel":       "2",
					"approverID":          "$.steps.2.result.level_2_approver_id",
					"level1ApprovalResult": "$.steps.4.result.approval_status",
					"hierarchyDetails":    "$.steps.2.result.hierarchy_details",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
				CompensationSteps: []int32{113},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Monitor Level-2 Approval (with timeout escalation)
			{
				StepNumber:    6,
				ServiceName:   "approval",
				HandlerMethod: "MonitorLevel2Approval",
				InputMapping: map[string]string{
					"documentID":      "$.steps.1.result.document_id",
					"approvalLevel":   "2",
					"approverID":      "$.steps.2.result.level_2_approver_id",
					"escalationRules": "$.steps.2.result.escalation_rules",
					"timeoutMinutes":  "20",
				},
				TimeoutSeconds: 1200,
				IsCritical:     false,
				CompensationSteps: []int32{114},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Route to Level-3 Approver (if required)
			{
				StepNumber:    7,
				ServiceName:   "workflow",
				HandlerMethod: "RouteToLevel3Approver",
				InputMapping: map[string]string{
					"documentID":          "$.steps.1.result.document_id",
					"approvalLevel":       "3",
					"approverID":          "$.steps.2.result.level_3_approver_id",
					"level2ApprovalResult": "$.steps.6.result.approval_status",
					"hierarchyDetails":    "$.steps.2.result.hierarchy_details",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
				CompensationSteps: []int32{115},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Collect All Approvals (CRITICAL)
			{
				StepNumber:    8,
				ServiceName:   "approval",
				HandlerMethod: "CollectAllApprovals",
				InputMapping: map[string]string{
					"documentID":           "$.steps.1.result.document_id",
					"level1Approval":       "$.steps.4.result.approval_status",
					"level2Approval":       "$.steps.6.result.approval_status",
					"level3Approval":       "$.steps.7.result.approval_status",
					"hierarchyRequirements": "$.steps.2.result.hierarchy_details",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{116},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Update Document Status to APPROVED (CRITICAL)
			{
				StepNumber:    9,
				ServiceName:   "workflow",
				HandlerMethod: "UpdateDocumentStatusToApproved",
				InputMapping: map[string]string{
					"documentID":          "$.steps.1.result.document_id",
					"allApprovalsStatus":  "$.steps.8.result.all_approvals_status",
					"approvalSummary":     "$.steps.8.result.approval_summary",
					"approvalCompletedAt": "$.steps.8.result.completion_timestamp",
				},
				TimeoutSeconds: 40,
				IsCritical:     true,
				CompensationSteps: []int32{117},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Trigger Next Workflow Step
			{
				StepNumber:    10,
				ServiceName:   "workflow",
				HandlerMethod: "TriggerNextWorkflowStep",
				InputMapping: map[string]string{
					"documentID":    "$.steps.1.result.document_id",
					"documentType":  "$.input.document_type",
					"currentStatus": "APPROVED",
					"workflowID":    "$.steps.1.result.workflow_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{118},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// ===== COMPENSATION STEPS =====

			// Step 110: Revert Approval Hierarchy Determination (compensates step 2)
			{
				StepNumber:    110,
				ServiceName:   "approval",
				HandlerMethod: "RevertApprovalHierarchyDetermination",
				InputMapping: map[string]string{
					"documentID": "$.steps.1.result.document_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
			// Step 111: Revert Level-1 Routing (compensates step 3)
			{
				StepNumber:    111,
				ServiceName:   "workflow",
				HandlerMethod: "RevertLevel1Routing",
				InputMapping: map[string]string{
					"documentID": "$.steps.1.result.document_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 112: Revert Level-1 Approval Monitoring (compensates step 4)
			{
				StepNumber:    112,
				ServiceName:   "approval",
				HandlerMethod: "RevertLevel1ApprovalMonitoring",
				InputMapping: map[string]string{
					"documentID": "$.steps.1.result.document_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 113: Revert Level-2 Routing (compensates step 5)
			{
				StepNumber:    113,
				ServiceName:   "workflow",
				HandlerMethod: "RevertLevel2Routing",
				InputMapping: map[string]string{
					"documentID": "$.steps.1.result.document_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 114: Revert Level-2 Approval Monitoring (compensates step 6)
			{
				StepNumber:    114,
				ServiceName:   "approval",
				HandlerMethod: "RevertLevel2ApprovalMonitoring",
				InputMapping: map[string]string{
					"documentID": "$.steps.1.result.document_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 115: Revert Level-3 Routing (compensates step 7)
			{
				StepNumber:    115,
				ServiceName:   "workflow",
				HandlerMethod: "RevertLevel3Routing",
				InputMapping: map[string]string{
					"documentID": "$.steps.1.result.document_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 116: Revert Approvals Collection (compensates step 8)
			{
				StepNumber:    116,
				ServiceName:   "approval",
				HandlerMethod: "RevertApprovalsCollection",
				InputMapping: map[string]string{
					"documentID": "$.steps.1.result.document_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 117: Revert Document Status Update (compensates step 9)
			{
				StepNumber:    117,
				ServiceName:   "workflow",
				HandlerMethod: "RevertDocumentStatusUpdate",
				InputMapping: map[string]string{
					"documentID": "$.steps.1.result.document_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
			// Step 118: Revert Next Step Trigger (compensates step 10)
			{
				StepNumber:    118,
				ServiceName:   "workflow",
				HandlerMethod: "RevertNextStepTrigger",
				InputMapping: map[string]string{
					"documentID": "$.steps.1.result.document_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *MultiLevelApprovalRoutingSaga) SagaType() string {
	return "SAGA-WF01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *MultiLevelApprovalRoutingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *MultiLevelApprovalRoutingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *MultiLevelApprovalRoutingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	inputData, ok := inputMap["input"].(map[string]interface{})
	if !ok {
		return errors.New("missing input object")
	}

	// Required fields
	requiredFields := []string{"document_id", "document_type", "amount", "submitter_id", "submission_date"}
	for _, field := range requiredFields {
		if inputData[field] == nil {
			return errors.New("missing required field: " + field)
		}
	}

	// Validate document type
	docType, ok := inputData["document_type"].(string)
	if !ok || docType == "" {
		return errors.New("document_type must be a non-empty string")
	}

	validDocTypes := map[string]bool{
		"PURCHASE_ORDER": true,
		"EXPENSE_CLAIM":  true,
		"TRAVEL_REQUEST": true,
		"REQUISITION":    true,
		"LEAVE_REQUEST":  true,
		"BUDGET_ALLOCATION": true,
	}
	if !validDocTypes[docType] {
		return errors.New("invalid document_type: " + docType)
	}

	// Validate amount
	amount, ok := inputData["amount"].(float64)
	if !ok || amount < 0 {
		return errors.New("amount must be a positive number")
	}

	// Department and role are optional but validate if present
	if dept, ok := inputData["department"].(string); ok && dept == "" {
		return errors.New("department must be a non-empty string if provided")
	}

	if roleID, ok := inputData["role_id"].(string); ok && roleID == "" {
		return errors.New("role_id must be a non-empty string if provided")
	}

	return nil
}
