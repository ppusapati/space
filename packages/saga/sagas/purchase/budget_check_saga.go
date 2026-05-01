// Package purchase provides saga handlers for purchase module workflows
package purchase

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// BudgetCheckSaga implements SAGA-P04: Budget Check workflow
// Business Flow: Check Budget → Lock Budget → Consume Budget → Approve Requisition → Record Consumption → Update Utilization → Release Lock
type BudgetCheckSaga struct {
	steps []*saga.StepDefinition
}

// NewBudgetCheckSaga creates a new Budget Check saga handler
func NewBudgetCheckSaga() saga.SagaHandler {
	return &BudgetCheckSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Check Available Budget
			{
				StepNumber:    1,
				ServiceName:   "budget",
				HandlerMethod: "CheckAvailableBudget",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"budgetCodeID":   "$.input.budget_code_id",
					"requestedAmount": "$.input.requested_amount",
					"fiscalYear":     "$.input.fiscal_year",
				},
				TimeoutSeconds: 15,
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
			// Step 2: Lock Budget Amount
			{
				StepNumber:    2,
				ServiceName:   "budget",
				HandlerMethod: "LockBudgetAmount",
				InputMapping: map[string]string{
					"budgetCodeID":   "$.input.budget_code_id",
					"lockedAmount":   "$.input.requested_amount",
					"lockReason":     "$.input.lock_reason",
					"requisitionID":  "$.input.requisition_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{102},
			},
			// Step 3: Consume Budget
			{
				StepNumber:    3,
				ServiceName:   "budget",
				HandlerMethod: "ConsumeBudget",
				InputMapping: map[string]string{
					"budgetCodeID":   "$.input.budget_code_id",
					"consumedAmount": "$.input.requested_amount",
					"lockID":         "$.steps.2.result.lock_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{103},
			},
			// Step 4: Approve Requisition
			{
				StepNumber:    4,
				ServiceName:   "procurement",
				HandlerMethod: "ApproveRequisition",
				InputMapping: map[string]string{
					"requisitionID": "$.input.requisition_id",
					"approverID":    "$.input.approver_id",
					"approvalNote":  "$.input.approval_note",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{104},
			},
			// Step 5: Record Consumption
			{
				StepNumber:    5,
				ServiceName:   "budget",
				HandlerMethod: "RecordConsumption",
				InputMapping: map[string]string{
					"budgetCodeID":   "$.input.budget_code_id",
					"consumedAmount": "$.input.requested_amount",
					"requisitionID":  "$.input.requisition_id",
					"consumptionDate": "$.input.consumption_date",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{105},
			},
			// Step 6: Update Utilization
			{
				StepNumber:    6,
				ServiceName:   "budget",
				HandlerMethod: "UpdateUtilization",
				InputMapping: map[string]string{
					"budgetCodeID":     "$.input.budget_code_id",
					"consumptionID":    "$.steps.5.result.consumption_id",
					"recalculateTotal": "true",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{106},
			},
			// Step 7: Release Lock (non-critical)
			{
				StepNumber:    7,
				ServiceName:   "budget",
				HandlerMethod: "ReleaseLock",
				InputMapping: map[string]string{
					"lockID": "$.steps.2.result.lock_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        false, // Non-critical: if fails, doesn't trigger compensation
				CompensationSteps: []int32{},
			},
			// ===== COMPENSATION STEPS =====

			// Step 102: Release Budget Lock (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "budget",
				HandlerMethod: "ReleaseBudgetLock",
				InputMapping: map[string]string{
					"lockID": "$.steps.2.result.lock_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 103: Restore Budget (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "budget",
				HandlerMethod: "RestoreBudget",
				InputMapping: map[string]string{
					"budgetCodeID":   "$.input.budget_code_id",
					"restoredAmount": "$.input.requested_amount",
					"consumptionID":  "$.steps.5.result.consumption_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 104: Revert Requisition Approval (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "procurement",
				HandlerMethod: "RevertRequisitionApproval",
				InputMapping: map[string]string{
					"requisitionID": "$.input.requisition_id",
					"revertReason":  "Budget check compensation",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 105: Delete Consumption Record (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "budget",
				HandlerMethod: "DeleteConsumptionRecord",
				InputMapping: map[string]string{
					"consumptionID": "$.steps.5.result.consumption_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 106: Revert Utilization (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "budget",
				HandlerMethod: "RevertUtilization",
				InputMapping: map[string]string{
					"budgetCodeID":  "$.input.budget_code_id",
					"consumptionID": "$.steps.5.result.consumption_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *BudgetCheckSaga) SagaType() string {
	return "SAGA-P04"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *BudgetCheckSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *BudgetCheckSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *BudgetCheckSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["budget_code_id"] == nil {
		return errors.New("budget_code_id is required")
	}

	if inputMap["requested_amount"] == nil {
		return errors.New("requested_amount is required")
	}

	if inputMap["requisition_id"] == nil {
		return errors.New("requisition_id is required")
	}

	if inputMap["approver_id"] == nil {
		return errors.New("approver_id is required")
	}

	if inputMap["fiscal_year"] == nil {
		return errors.New("fiscal_year is required")
	}

	return nil
}
