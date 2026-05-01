// Package inventory provides saga handlers for inventory module workflows
package inventory

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// CycleCountSaga implements SAGA-I02: Cycle Count & Stock Adjustment workflow
// Business Flow: Create Count → Freeze Stock → Execute Count → Calculate Variance → Approve Adjustment → Update Stock → Post GL → Audit Trail → Complete
type CycleCountSaga struct {
	steps []*saga.StepDefinition
}

// NewCycleCountSaga creates a new Cycle Count saga handler
func NewCycleCountSaga() saga.SagaHandler {
	return &CycleCountSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Cycle Count
			{
				StepNumber:    1,
				ServiceName:   "cycle-count",
				HandlerMethod: "CreateCycleCount",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"warehouseID":     "$.input.warehouse_id",
					"countType":       "$.input.count_type",
					"countScheduleID": "$.input.count_schedule_id",
				},
				TimeoutSeconds: 20,
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
			// Step 2: Freeze Stock Movement
			{
				StepNumber:    2,
				ServiceName:   "inventory-core",
				HandlerMethod: "FreezeStockMovement",
				InputMapping: map[string]string{
					"warehouseID": "$.input.warehouse_id",
					"countID":     "$.steps.1.result.count_id",
					"freezeReason": "Cycle count in progress",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{102},
			},
			// Step 3: Execute Count
			{
				StepNumber:    3,
				ServiceName:   "cycle-count",
				HandlerMethod: "ExecuteCount",
				InputMapping: map[string]string{
					"countID":      "$.steps.1.result.count_id",
					"warehouseID":  "$.input.warehouse_id",
					"countDetails": "$.input.count_details",
				},
				TimeoutSeconds:    45,
				IsCritical:        true,
				CompensationSteps: []int32{},
			},
			// Step 4: Calculate Variance
			{
				StepNumber:    4,
				ServiceName:   "cycle-count",
				HandlerMethod: "CalculateVariance",
				InputMapping: map[string]string{
					"countID":        "$.steps.1.result.count_id",
					"warehouseID":    "$.input.warehouse_id",
					"varianceThreshold": "$.input.variance_threshold",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{104},
			},
			// Step 5: Approve Adjustment
			{
				StepNumber:    5,
				ServiceName:   "cycle-count",
				HandlerMethod: "ApproveAdjustment",
				InputMapping: map[string]string{
					"countID":     "$.steps.1.result.count_id",
					"approverID":  "$.input.approver_id",
					"approvalNote": "$.input.approval_note",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{105},
			},
			// Step 6: Adjust Stock
			{
				StepNumber:    6,
				ServiceName:   "inventory-core",
				HandlerMethod: "AdjustStock",
				InputMapping: map[string]string{
					"countID":     "$.steps.1.result.count_id",
					"warehouseID": "$.input.warehouse_id",
					"adjustmentDetails": "$.steps.4.result.variance_details",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{106},
			},
			// Step 7: Post Adjustment Journal
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostAdjustmentJournal",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"countID":       "$.steps.1.result.count_id",
					"adjustmentDetails": "$.steps.4.result.variance_details",
					"journalDate":   "$.input.journal_date",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{107},
			},
			// Step 8: Record Audit Trail
			{
				StepNumber:    8,
				ServiceName:   "cycle-count",
				HandlerMethod: "RecordAuditTrail",
				InputMapping: map[string]string{
					"countID":    "$.steps.1.result.count_id",
					"userID":     "$.input.user_id",
					"action":     "CYCLE_COUNT_COMPLETED",
					"details":    "$.steps.4.result.variance_details",
				},
				TimeoutSeconds:    15,
				IsCritical:        false, // Non-critical audit
				CompensationSteps: []int32{108},
			},
			// Step 9: Complete Count
			{
				StepNumber:    9,
				ServiceName:   "cycle-count",
				HandlerMethod: "CompleteCount",
				InputMapping: map[string]string{
					"countID": "$.steps.1.result.count_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{},
			},
			// ===== COMPENSATION STEPS =====

			// Step 101: Cancel Cycle Count (compensates step 1)
			{
				StepNumber:    101,
				ServiceName:   "cycle-count",
				HandlerMethod: "CancelCycleCount",
				InputMapping: map[string]string{
					"countID": "$.steps.1.result.count_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 102: Unfreeze Stock (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "inventory-core",
				HandlerMethod: "UnfreezeStock",
				InputMapping: map[string]string{
					"warehouseID": "$.input.warehouse_id",
					"countID":     "$.steps.1.result.count_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 104: Clear Variance Data (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "cycle-count",
				HandlerMethod: "ClearVarianceData",
				InputMapping: map[string]string{
					"countID": "$.steps.1.result.count_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Revert Approval (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "cycle-count",
				HandlerMethod: "RevertApproval",
				InputMapping: map[string]string{
					"countID": "$.steps.1.result.count_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Reverse Adjustment (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "inventory-core",
				HandlerMethod: "ReverseAdjustment",
				InputMapping: map[string]string{
					"countID":     "$.steps.1.result.count_id",
					"warehouseID": "$.input.warehouse_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 107: Reverse Adjustment Journal (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseAdjustmentJournal",
				InputMapping: map[string]string{
					"countID": "$.steps.1.result.count_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 108: Delete Audit Log (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "cycle-count",
				HandlerMethod: "DeleteAuditLog",
				InputMapping: map[string]string{
					"countID": "$.steps.1.result.count_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *CycleCountSaga) SagaType() string {
	return "SAGA-I02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *CycleCountSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *CycleCountSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *CycleCountSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["warehouse_id"] == nil {
		return errors.New("warehouse_id is required")
	}

	if inputMap["count_type"] == nil {
		return errors.New("count_type is required")
	}

	if inputMap["approver_id"] == nil {
		return errors.New("approver_id is required")
	}

	if inputMap["count_details"] == nil {
		return errors.New("count_details are required")
	}

	return nil
}
