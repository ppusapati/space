// Package finance provides saga handlers for finance module workflows
package finance

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// CostCenterAllocationSaga implements SAGA-F08: Cost Center Allocation workflow
// Business Flow: IdentifyOverheadCosts → DetermineAllocationBasis → CalculateAllocations → PostAllocationJournals → UpdateCostCenterBalances → GenerateDepartmentalPL → CompleteAllocation
// Management Accounting: Overhead allocation to cost centers for departmental P&L
type CostCenterAllocationSaga struct {
	steps []*saga.StepDefinition
}

// NewCostCenterAllocationSaga creates a new Cost Center Allocation saga handler
func NewCostCenterAllocationSaga() saga.SagaHandler {
	return &CostCenterAllocationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Identify Overhead Costs to Allocate
			{
				StepNumber:    1,
				ServiceName:   "cost-center",
				HandlerMethod: "IdentifyOverheadCosts",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"allocationPeriod": "$.input.allocation_period",
					"costPoolID":       "$.input.cost_pool_id",
					"overheadAccounts": "$.input.overhead_accounts",
					"allocationMethod": "$.input.allocation_method",
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
			// Step 2: Determine Allocation Basis
			{
				StepNumber:    2,
				ServiceName:   "cost-center",
				HandlerMethod: "DetermineAllocationBasis",
				InputMapping: map[string]string{
					"allocationID":     "$.steps.1.result.allocation_id",
					"allocationMethod": "$.input.allocation_method",
					"targetCostCenters": "$.input.target_cost_centers",
					"allocationDrivers": "$.input.allocation_drivers",
					"allocationPeriod": "$.input.allocation_period",
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
			// Step 3: Calculate Cost Center Allocations
			{
				StepNumber:    3,
				ServiceName:   "cost-center",
				HandlerMethod: "CalculateCostCenterAllocations",
				InputMapping: map[string]string{
					"allocationID":     "$.steps.1.result.allocation_id",
					"totalOverheadCost": "$.steps.1.result.total_overhead_cost",
					"allocationBasis":  "$.steps.2.result.allocation_basis",
					"targetCostCenters": "$.input.target_cost_centers",
					"allocationMethod": "$.input.allocation_method",
				},
				TimeoutSeconds:    35,
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
			// Step 4: Post Allocation Journal Entries
			{
				StepNumber:    4,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostAllocationJournals",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"allocationID":     "$.steps.1.result.allocation_id",
					"allocations":      "$.steps.3.result.allocations",
					"allocationDate":   "$.input.allocation_date",
					"costPoolID":       "$.input.cost_pool_id",
				},
				TimeoutSeconds:    40,
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
			// Step 5: Update Cost Center Balances
			{
				StepNumber:    5,
				ServiceName:   "cost-center",
				HandlerMethod: "UpdateCostCenterBalances",
				InputMapping: map[string]string{
					"allocationID":     "$.steps.1.result.allocation_id",
					"allocations":      "$.steps.3.result.allocations",
					"allocationPeriod": "$.input.allocation_period",
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
			// Step 6: Generate Departmental P&L Reports
			{
				StepNumber:    6,
				ServiceName:   "financial-reports",
				HandlerMethod: "GenerateDepartmentalPL",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"allocationID":     "$.steps.1.result.allocation_id",
					"allocations":      "$.steps.3.result.allocations",
					"targetCostCenters": "$.input.target_cost_centers",
					"reportPeriod":     "$.input.allocation_period",
				},
				TimeoutSeconds:    35,
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
			// Step 7: Complete Allocation Process
			{
				StepNumber:    7,
				ServiceName:   "cost-center",
				HandlerMethod: "CompleteAllocation",
				InputMapping: map[string]string{
					"allocationID":     "$.steps.1.result.allocation_id",
					"totalOverheadCost": "$.steps.1.result.total_overhead_cost",
					"allocations":      "$.steps.3.result.allocations",
					"completionDate":   "$.input.allocation_date",
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

			// Step 101: Cancel Overhead Identification (compensates step 1)
			{
				StepNumber:    101,
				ServiceName:   "cost-center",
				HandlerMethod: "CancelOverheadIdentification",
				InputMapping: map[string]string{
					"allocationID": "$.steps.1.result.allocation_id",
					"reason":       "Saga compensation - cost center allocation failed",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Clear Allocation Basis (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "cost-center",
				HandlerMethod: "ClearAllocationBasis",
				InputMapping: map[string]string{
					"allocationID": "$.steps.1.result.allocation_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Delete Allocation Calculations (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "cost-center",
				HandlerMethod: "DeleteAllocationCalculations",
				InputMapping: map[string]string{
					"allocationID": "$.steps.1.result.allocation_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 104: Reverse Allocation Journals (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseAllocationJournals",
				InputMapping: map[string]string{
					"allocationID": "$.steps.1.result.allocation_id",
					"journalDate":  "$.input.allocation_date",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
			// Step 105: Revert Cost Center Balances (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "cost-center",
				HandlerMethod: "RevertCostCenterBalances",
				InputMapping: map[string]string{
					"allocationID": "$.steps.1.result.allocation_id",
					"allocations":  "$.steps.3.result.allocations",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: Delete Departmental P&L Reports (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "financial-reports",
				HandlerMethod: "DeleteDepartmentalPLReports",
				InputMapping: map[string]string{
					"allocationID": "$.steps.1.result.allocation_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *CostCenterAllocationSaga) SagaType() string {
	return "SAGA-F08"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *CostCenterAllocationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *CostCenterAllocationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *CostCenterAllocationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["allocation_period"] == nil {
		return errors.New("allocation_period is required")
	}

	if inputMap["cost_pool_id"] == nil {
		return errors.New("cost_pool_id is required")
	}

	if inputMap["overhead_accounts"] == nil {
		return errors.New("overhead_accounts are required")
	}

	overheadAccounts, ok := inputMap["overhead_accounts"].([]interface{})
	if !ok || len(overheadAccounts) == 0 {
		return errors.New("overhead_accounts must be a non-empty list")
	}

	if inputMap["allocation_method"] == nil {
		return errors.New("allocation_method is required (e.g., DIRECT, STEP_DOWN, RECIPROCAL)")
	}

	if inputMap["target_cost_centers"] == nil {
		return errors.New("target_cost_centers are required")
	}

	targetCostCenters, ok := inputMap["target_cost_centers"].([]interface{})
	if !ok || len(targetCostCenters) == 0 {
		return errors.New("target_cost_centers must be a non-empty list")
	}

	if inputMap["allocation_drivers"] == nil {
		return errors.New("allocation_drivers are required")
	}

	allocationDrivers, ok := inputMap["allocation_drivers"].(map[string]interface{})
	if !ok || len(allocationDrivers) == 0 {
		return errors.New("allocation_drivers must be a non-empty map")
	}

	if inputMap["allocation_date"] == nil {
		return errors.New("allocation_date is required")
	}

	// Validate allocation method
	allocationMethod, ok := inputMap["allocation_method"].(string)
	if ok {
		validMethods := map[string]bool{
			"DIRECT":     true,
			"STEP_DOWN":  true,
			"RECIPROCAL": true,
			"ACTIVITY_BASED": true,
		}
		if !validMethods[allocationMethod] {
			return errors.New("invalid allocation_method: must be DIRECT, STEP_DOWN, RECIPROCAL, or ACTIVITY_BASED")
		}
	}

	return nil
}
