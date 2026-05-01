// Package workflow provides saga handlers for workflow management
package workflow

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ParallelConsolidationSaga implements SAGA-WF03: Parallel Consolidation & Multi-Branch Processing
// Business Flow: InitiateParallelProcessing → CreateProcessingTasks →
//               ExecuteProcessingBranch1 → ExecuteProcessingBranch2 → ExecuteProcessingBranch3 →
//               ExecuteConsolidationBranch1 → ExecuteConsolidationBranch2 → ExecuteConsolidationBranch3 →
//               ReconcileAllBranchResults → PostConsolidatedResult
// Steps: 10 forward + 9 compensation = 19 total
// Timeout: 480 seconds, Critical steps: 1,9,10
type ParallelConsolidationSaga struct {
	steps []*saga.StepDefinition
}

// NewParallelConsolidationSaga creates a new Parallel Consolidation saga handler
func NewParallelConsolidationSaga() saga.SagaHandler {
	return &ParallelConsolidationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Parallel Processing for Multiple Branches (CRITICAL)
			{
				StepNumber:    1,
				ServiceName:   "workflow",
				HandlerMethod: "InitiateParallelProcessing",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"processID":      "$.input.process_id",
					"branchCount":    "$.input.branch_count",
					"branchList":     "$.input.branch_list",
					"initiationDate": "$.input.initiation_date",
					"processType":    "$.input.process_type",
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
				CompensationSteps: []int32{},
			},
			// Step 2: Create Processing Tasks for Each Branch
			{
				StepNumber:    2,
				ServiceName:   "workflow",
				HandlerMethod: "CreateProcessingTasks",
				InputMapping: map[string]string{
					"processID":        "$.steps.1.result.process_id",
					"branchList":       "$.input.branch_list",
					"branchCount":      "$.input.branch_count",
					"taskTemplate":     "$.input.task_template",
					"executionContext": "$.steps.1.result.execution_context",
				},
				TimeoutSeconds: 40,
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
			// Step 3: Execute Processing Step in Branch 1 (Parallel)
			{
				StepNumber:    3,
				ServiceName:   "branch",
				HandlerMethod: "ExecuteProcessingStepBranch1",
				InputMapping: map[string]string{
					"processID":   "$.steps.1.result.process_id",
					"branchID":    "$.input.branch_list.0.branch_id",
					"branchName":  "$.input.branch_list.0.branch_name",
					"taskID":      "$.steps.2.result.branch_1_task_id",
					"processData": "$.steps.2.result.branch_1_process_data",
				},
				TimeoutSeconds: 80,
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
			// Step 4: Execute Processing Step in Branch 2 (Parallel)
			{
				StepNumber:    4,
				ServiceName:   "branch",
				HandlerMethod: "ExecuteProcessingStepBranch2",
				InputMapping: map[string]string{
					"processID":   "$.steps.1.result.process_id",
					"branchID":    "$.input.branch_list.1.branch_id",
					"branchName":  "$.input.branch_list.1.branch_name",
					"taskID":      "$.steps.2.result.branch_2_task_id",
					"processData": "$.steps.2.result.branch_2_process_data",
				},
				TimeoutSeconds: 80,
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
			// Step 5: Execute Processing Step in Branch 3 (Parallel)
			{
				StepNumber:    5,
				ServiceName:   "branch",
				HandlerMethod: "ExecuteProcessingStepBranch3",
				InputMapping: map[string]string{
					"processID":   "$.steps.1.result.process_id",
					"branchID":    "$.input.branch_list.2.branch_id",
					"branchName":  "$.input.branch_list.2.branch_name",
					"taskID":      "$.steps.2.result.branch_3_task_id",
					"processData": "$.steps.2.result.branch_3_process_data",
				},
				TimeoutSeconds: 80,
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
			// Step 6: Execute Consolidation Step in Branch 1 (Parallel)
			{
				StepNumber:    6,
				ServiceName:   "consolidation",
				HandlerMethod: "ExecuteConsolidationStepBranch1",
				InputMapping: map[string]string{
					"processID":      "$.steps.1.result.process_id",
					"branchID":       "$.input.branch_list.0.branch_id",
					"processingResult": "$.steps.3.result.processing_result",
					"taskID":         "$.steps.2.result.branch_1_consolidation_task_id",
					"consolidationRules": "$.input.consolidation_rules",
				},
				TimeoutSeconds: 70,
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
			// Step 7: Execute Consolidation Step in Branch 2 (Parallel)
			{
				StepNumber:    7,
				ServiceName:   "consolidation",
				HandlerMethod: "ExecuteConsolidationStepBranch2",
				InputMapping: map[string]string{
					"processID":      "$.steps.1.result.process_id",
					"branchID":       "$.input.branch_list.1.branch_id",
					"processingResult": "$.steps.4.result.processing_result",
					"taskID":         "$.steps.2.result.branch_2_consolidation_task_id",
					"consolidationRules": "$.input.consolidation_rules",
				},
				TimeoutSeconds: 70,
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
			// Step 8: Execute Consolidation Step in Branch 3 (Parallel)
			{
				StepNumber:    8,
				ServiceName:   "consolidation",
				HandlerMethod: "ExecuteConsolidationStepBranch3",
				InputMapping: map[string]string{
					"processID":      "$.steps.1.result.process_id",
					"branchID":       "$.input.branch_list.2.branch_id",
					"processingResult": "$.steps.5.result.processing_result",
					"taskID":         "$.steps.2.result.branch_3_consolidation_task_id",
					"consolidationRules": "$.input.consolidation_rules",
				},
				TimeoutSeconds: 70,
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
			// Step 9: Reconcile All Branch Results (CRITICAL)
			{
				StepNumber:    9,
				ServiceName:   "reconciliation",
				HandlerMethod: "ReconcileAllBranchResults",
				InputMapping: map[string]string{
					"processID":                "$.steps.1.result.process_id",
					"branch1Result":            "$.steps.6.result.consolidation_result",
					"branch2Result":            "$.steps.7.result.consolidation_result",
					"branch3Result":            "$.steps.8.result.consolidation_result",
					"reconciliationRules":      "$.input.reconciliation_rules",
					"reconciliationThresholds": "$.input.reconciliation_thresholds",
				},
				TimeoutSeconds: 60,
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
			// Step 10: Post Consolidated Result (CRITICAL)
			{
				StepNumber:    10,
				ServiceName:   "workflow",
				HandlerMethod: "PostConsolidatedResult",
				InputMapping: map[string]string{
					"processID":            "$.steps.1.result.process_id",
					"reconciliationStatus": "$.steps.9.result.reconciliation_status",
					"consolidatedData":     "$.steps.9.result.consolidated_data",
					"postingDate":          "$.steps.9.result.reconciliation_timestamp",
					"postingRules":         "$.input.posting_rules",
				},
				TimeoutSeconds: 50,
				IsCritical:     true,
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

			// Step 110: Revert Processing Tasks Creation (compensates step 2)
			{
				StepNumber:    110,
				ServiceName:   "workflow",
				HandlerMethod: "RevertProcessingTasksCreation",
				InputMapping: map[string]string{
					"processID": "$.steps.1.result.process_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
			// Step 111: Rollback Branch 1 Processing Update (compensates step 3)
			{
				StepNumber:    111,
				ServiceName:   "branch",
				HandlerMethod: "RollbackProcessingUpdateBranch1",
				InputMapping: map[string]string{
					"processID": "$.steps.1.result.process_id",
					"branchID":  "$.input.branch_list.0.branch_id",
					"taskID":    "$.steps.2.result.branch_1_task_id",
				},
				TimeoutSeconds: 80,
				IsCritical:     false,
			},
			// Step 112: Rollback Branch 2 Processing Update (compensates step 4)
			{
				StepNumber:    112,
				ServiceName:   "branch",
				HandlerMethod: "RollbackProcessingUpdateBranch2",
				InputMapping: map[string]string{
					"processID": "$.steps.1.result.process_id",
					"branchID":  "$.input.branch_list.1.branch_id",
					"taskID":    "$.steps.2.result.branch_2_task_id",
				},
				TimeoutSeconds: 80,
				IsCritical:     false,
			},
			// Step 113: Rollback Branch 3 Processing Update (compensates step 5)
			{
				StepNumber:    113,
				ServiceName:   "branch",
				HandlerMethod: "RollbackProcessingUpdateBranch3",
				InputMapping: map[string]string{
					"processID": "$.steps.1.result.process_id",
					"branchID":  "$.input.branch_list.2.branch_id",
					"taskID":    "$.steps.2.result.branch_3_task_id",
				},
				TimeoutSeconds: 80,
				IsCritical:     false,
			},
			// Step 114: Rollback Consolidation Update Branch 1 (compensates step 6)
			{
				StepNumber:    114,
				ServiceName:   "consolidation",
				HandlerMethod: "RollbackConsolidationUpdateBranch1",
				InputMapping: map[string]string{
					"processID": "$.steps.1.result.process_id",
					"branchID":  "$.input.branch_list.0.branch_id",
					"taskID":    "$.steps.2.result.branch_1_consolidation_task_id",
				},
				TimeoutSeconds: 70,
				IsCritical:     false,
			},
			// Step 115: Rollback Consolidation Update Branch 2 (compensates step 7)
			{
				StepNumber:    115,
				ServiceName:   "consolidation",
				HandlerMethod: "RollbackConsolidationUpdateBranch2",
				InputMapping: map[string]string{
					"processID": "$.steps.1.result.process_id",
					"branchID":  "$.input.branch_list.1.branch_id",
					"taskID":    "$.steps.2.result.branch_2_consolidation_task_id",
				},
				TimeoutSeconds: 70,
				IsCritical:     false,
			},
			// Step 116: Rollback Consolidation Update Branch 3 (compensates step 8)
			{
				StepNumber:    116,
				ServiceName:   "consolidation",
				HandlerMethod: "RollbackConsolidationUpdateBranch3",
				InputMapping: map[string]string{
					"processID": "$.steps.1.result.process_id",
					"branchID":  "$.input.branch_list.2.branch_id",
					"taskID":    "$.steps.2.result.branch_3_consolidation_task_id",
				},
				TimeoutSeconds: 70,
				IsCritical:     false,
			},
			// Step 117: Revert Reconciliation (compensates step 9)
			{
				StepNumber:    117,
				ServiceName:   "reconciliation",
				HandlerMethod: "RevertReconciliation",
				InputMapping: map[string]string{
					"processID": "$.steps.1.result.process_id",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 118: Revert Consolidated Result Posting (compensates step 10)
			{
				StepNumber:    118,
				ServiceName:   "workflow",
				HandlerMethod: "RevertConsolidatedResultPosting",
				InputMapping: map[string]string{
					"processID": "$.steps.1.result.process_id",
				},
				TimeoutSeconds: 50,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ParallelConsolidationSaga) SagaType() string {
	return "SAGA-WF03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ParallelConsolidationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ParallelConsolidationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ParallelConsolidationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	inputData, ok := inputMap["input"].(map[string]interface{})
	if !ok {
		return errors.New("missing input object")
	}

	// Required fields
	requiredFields := []string{"process_id", "branch_count", "branch_list", "initiation_date"}
	for _, field := range requiredFields {
		if inputData[field] == nil {
			return errors.New("missing required field: " + field)
		}
	}

	// Validate process_id
	processID, ok := inputData["process_id"].(string)
	if !ok || processID == "" {
		return errors.New("process_id must be a non-empty string")
	}

	// Validate branch_count
	branchCount, ok := inputData["branch_count"].(float64)
	if !ok || branchCount <= 0 || branchCount > 100 {
		return errors.New("branch_count must be a positive number between 1 and 100")
	}

	// Validate branch_list is an array
	branchList, ok := inputData["branch_list"].([]interface{})
	if !ok || len(branchList) == 0 {
		return errors.New("branch_list must be a non-empty array")
	}

	// Validate branch_list count matches branch_count
	if int(branchCount) != len(branchList) {
		return errors.New("branch_count must match the number of branches in branch_list")
	}

	// Validate process_type if present
	if processType, ok := inputData["process_type"].(string); ok && processType == "" {
		return errors.New("process_type must be a non-empty string if provided")
	}

	// Validate consolidation_rules if present
	if rules, ok := inputData["consolidation_rules"].(string); ok && rules == "" {
		return errors.New("consolidation_rules must be a non-empty string if provided")
	}

	// Validate reconciliation_rules if present
	if rules, ok := inputData["reconciliation_rules"].(string); ok && rules == "" {
		return errors.New("reconciliation_rules must be a non-empty string if provided")
	}

	return nil
}
