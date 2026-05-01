// Package workflow provides saga handlers for workflow management
package workflow

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ConditionalWorkflowRoutingSaga implements SAGA-WF02: Conditional Workflow Routing
// Business Flow: EvaluateDocumentProperties → DetermineApprovalRequirement →
//               RouteLessThan50K → RouteBetween50KAnd500K → RouteGreaterThan500K →
//               ExecuteApprovalPath → CollectConditionalApprovals →
//               UpdateDocumentStatusByCondition → TriggerPostApprovalActions
// Steps: 9 forward + 8 compensation = 17 total
// Timeout: 300 seconds, Critical steps: 2,6,8
type ConditionalWorkflowRoutingSaga struct {
	steps []*saga.StepDefinition
}

// NewConditionalWorkflowRoutingSaga creates a new Conditional Workflow Routing saga handler
func NewConditionalWorkflowRoutingSaga() saga.SagaHandler {
	return &ConditionalWorkflowRoutingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Evaluate Document Properties
			{
				StepNumber:    1,
				ServiceName:   "workflow",
				HandlerMethod: "EvaluateDocumentProperties",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"documentID":     "$.input.document_id",
					"documentType":   "$.input.document_type",
					"amount":         "$.input.amount",
					"department":     "$.input.department",
					"evaluationDate": "$.input.evaluation_date",
				},
				TimeoutSeconds: 25,
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
			// Step 2: Determine Approval Requirement (CRITICAL)
			{
				StepNumber:    2,
				ServiceName:   "rule-engine",
				HandlerMethod: "DetermineApprovalRequirement",
				InputMapping: map[string]string{
					"documentID":    "$.steps.1.result.document_id",
					"amount":        "$.input.amount",
					"department":    "$.input.department",
					"documentType":  "$.input.document_type",
					"ruleSetID":     "$.input.rule_set_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     true,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Route for <50K (self-approval allowed)
			{
				StepNumber:    3,
				ServiceName:   "approval",
				HandlerMethod: "RouteLessThan50K",
				InputMapping: map[string]string{
					"documentID":         "$.steps.1.result.document_id",
					"amount":             "$.input.amount",
					"approvalThreshold":  "50000",
					"conditionMet":       "$.steps.2.result.route_under_50k",
					"approverID":         "$.steps.2.result.self_approval_user_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{110},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Route for 50K-500K (department manager approval)
			{
				StepNumber:    4,
				ServiceName:   "department",
				HandlerMethod: "RouteBetween50KAnd500K",
				InputMapping: map[string]string{
					"documentID":           "$.steps.1.result.document_id",
					"amount":               "$.input.amount",
					"lowerThreshold":       "50000",
					"upperThreshold":       "500000",
					"conditionMet":         "$.steps.2.result.route_50k_to_500k",
					"departmentManagerID":  "$.steps.2.result.department_manager_id",
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
			// Step 5: Route for >500K (finance & director approval)
			{
				StepNumber:    5,
				ServiceName:   "approval",
				HandlerMethod: "RouteGreaterThan500K",
				InputMapping: map[string]string{
					"documentID":       "$.steps.1.result.document_id",
					"amount":           "$.input.amount",
					"upperThreshold":   "500000",
					"conditionMet":     "$.steps.2.result.route_above_500k",
					"financeApprover":  "$.steps.2.result.finance_approver_id",
					"directorApprover": "$.steps.2.result.director_id",
				},
				TimeoutSeconds: 35,
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
			// Step 6: Execute Determined Approval Path (CRITICAL)
			{
				StepNumber:    6,
				ServiceName:   "workflow",
				HandlerMethod: "ExecuteApprovalPath",
				InputMapping: map[string]string{
					"documentID":        "$.steps.1.result.document_id",
					"selectedRoute":     "$.steps.2.result.approval_route",
					"route3Status":      "$.steps.3.result.routing_status",
					"route4Status":      "$.steps.4.result.routing_status",
					"route5Status":      "$.steps.5.result.routing_status",
					"executionStrategy": "$.steps.2.result.execution_strategy",
				},
				TimeoutSeconds: 60,
				IsCritical:     true,
				CompensationSteps: []int32{113},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Collect Conditional Approvals
			{
				StepNumber:    7,
				ServiceName:   "approval",
				HandlerMethod: "CollectConditionalApprovals",
				InputMapping: map[string]string{
					"documentID":       "$.steps.1.result.document_id",
					"approvalRoute":    "$.steps.2.result.approval_route",
					"pathExecutionResult": "$.steps.6.result.execution_status",
					"collectionTimeout": "$.steps.2.result.collection_timeout",
				},
				TimeoutSeconds: 45,
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
			// Step 8: Update Document Status Based on Condition Met (CRITICAL)
			{
				StepNumber:    8,
				ServiceName:   "workflow",
				HandlerMethod: "UpdateDocumentStatusByCondition",
				InputMapping: map[string]string{
					"documentID":             "$.steps.1.result.document_id",
					"conditionalApprovals":   "$.steps.7.result.approvals_collected",
					"approvalRoute":          "$.steps.2.result.approval_route",
					"documentStatus":         "$.steps.7.result.document_status",
					"completionTimestamp":    "$.steps.7.result.completion_timestamp",
				},
				TimeoutSeconds: 40,
				IsCritical:     true,
				CompensationSteps: []int32{115},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Trigger Post-Approval Actions
			{
				StepNumber:    9,
				ServiceName:   "notification",
				HandlerMethod: "TriggerPostApprovalActions",
				InputMapping: map[string]string{
					"documentID":       "$.steps.1.result.document_id",
					"documentType":     "$.input.document_type",
					"approvalRoute":    "$.steps.2.result.approval_route",
					"finalStatus":      "$.steps.8.result.document_status",
					"notificationList": "$.steps.2.result.post_approval_notifications",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{116},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// ===== COMPENSATION STEPS =====

			// Step 109: Revert Approval Requirement Determination (compensates step 2)
			{
				StepNumber:    109,
				ServiceName:   "rule-engine",
				HandlerMethod: "RevertApprovalRequirement",
				InputMapping: map[string]string{
					"documentID": "$.steps.1.result.document_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
			// Step 110: Revert <50K Routing (compensates step 3)
			{
				StepNumber:    110,
				ServiceName:   "approval",
				HandlerMethod: "RevertLessThan50KRouting",
				InputMapping: map[string]string{
					"documentID": "$.steps.1.result.document_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 111: Revert 50K-500K Routing (compensates step 4)
			{
				StepNumber:    111,
				ServiceName:   "department",
				HandlerMethod: "RevertBetween50KAnd500KRouting",
				InputMapping: map[string]string{
					"documentID": "$.steps.1.result.document_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 112: Revert >500K Routing (compensates step 5)
			{
				StepNumber:    112,
				ServiceName:   "approval",
				HandlerMethod: "RevertGreaterThan500KRouting",
				InputMapping: map[string]string{
					"documentID": "$.steps.1.result.document_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 113: Revert Approval Path Execution (compensates step 6)
			{
				StepNumber:    113,
				ServiceName:   "workflow",
				HandlerMethod: "RevertApprovalPathExecution",
				InputMapping: map[string]string{
					"documentID": "$.steps.1.result.document_id",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 114: Revert Conditional Approvals Collection (compensates step 7)
			{
				StepNumber:    114,
				ServiceName:   "approval",
				HandlerMethod: "RevertConditionalApprovalsCollection",
				InputMapping: map[string]string{
					"documentID": "$.steps.1.result.document_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 115: Revert Document Status Update (compensates step 8)
			{
				StepNumber:    115,
				ServiceName:   "workflow",
				HandlerMethod: "RevertDocumentStatusUpdate",
				InputMapping: map[string]string{
					"documentID": "$.steps.1.result.document_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
			// Step 116: Revert Post-Approval Actions (compensates step 9)
			{
				StepNumber:    116,
				ServiceName:   "notification",
				HandlerMethod: "RevertPostApprovalActions",
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
func (s *ConditionalWorkflowRoutingSaga) SagaType() string {
	return "SAGA-WF02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ConditionalWorkflowRoutingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ConditionalWorkflowRoutingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ConditionalWorkflowRoutingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	inputData, ok := inputMap["input"].(map[string]interface{})
	if !ok {
		return errors.New("missing input object")
	}

	// Required fields
	requiredFields := []string{"document_id", "document_type", "amount", "department", "evaluation_date"}
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
		"EXPENSE_CLAIM":       true,
		"PURCHASE_REQUEST":    true,
		"CAPITAL_EXPENDITURE": true,
		"TRAVEL_REQUEST":      true,
		"PAYMENT_REQUEST":     true,
	}
	if !validDocTypes[docType] {
		return errors.New("invalid document_type: " + docType)
	}

	// Validate amount
	amount, ok := inputData["amount"].(float64)
	if !ok || amount < 0 {
		return errors.New("amount must be a positive number")
	}

	// Validate department
	department, ok := inputData["department"].(string)
	if !ok || department == "" {
		return errors.New("department must be a non-empty string")
	}

	// Optional rule_set_id validation
	if ruleSetID, ok := inputData["rule_set_id"].(string); ok && ruleSetID == "" {
		return errors.New("rule_set_id must be a non-empty string if provided")
	}

	return nil
}
