// Package agriculture provides saga handlers for agricultural workflows
package agriculture

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// CropPlanningSaga implements SAGA-A01: Crop Planning & Resource Allocation workflow
// Business Flow: InitiateCropPlan → ValidateFarmArea → AllocateResources → ProcureSeedsFertilizer → ProcessInventoryUpdate → CalculateBudgetRequirement → ReserveBudget → ApplyLedgerEntries → ConfirmPlanning
// Steps: 9 forward + 8 compensation = 17 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,6,9
type CropPlanningSaga struct {
	steps []*saga.StepDefinition
}

// NewCropPlanningSaga creates a new Crop Planning saga handler
func NewCropPlanningSaga() saga.SagaHandler {
	return &CropPlanningSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Crop Plan
			{
				StepNumber:    1,
				ServiceName:   "agriculture",
				HandlerMethod: "InitiateCropPlan",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"cropPlanID":     "$.input.crop_plan_id",
					"cropType":       "$.input.crop_type",
					"farmArea":       "$.input.farm_area",
					"plantingSeason": "$.input.planting_season",
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
			// Step 2: Validate Farm Area and Soil Conditions
			{
				StepNumber:    2,
				ServiceName:   "crop-planning",
				HandlerMethod: "ValidateFarmArea",
				InputMapping: map[string]string{
					"cropPlanID":     "$.steps.1.result.crop_plan_id",
					"farmArea":       "$.input.farm_area",
					"cropType":       "$.input.crop_type",
					"plantingSeason": "$.input.planting_season",
					"validateSoil":   "true",
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
			// Step 3: Allocate Farm Resources
			{
				StepNumber:    3,
				ServiceName:   "crop-planning",
				HandlerMethod: "AllocateResources",
				InputMapping: map[string]string{
					"cropPlanID":       "$.steps.1.result.crop_plan_id",
					"cropType":         "$.input.crop_type",
					"farmArea":         "$.input.farm_area",
					"validationResult": "$.steps.2.result.validation_result",
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
			// Step 4: Procure Seeds and Fertilizer
			{
				StepNumber:    4,
				ServiceName:   "procurement",
				HandlerMethod: "ProcureSeedsFertilizer",
				InputMapping: map[string]string{
					"cropPlanID":        "$.steps.1.result.crop_plan_id",
					"cropType":          "$.input.crop_type",
					"allocationData":    "$.steps.3.result.allocation_data",
					"farmArea":          "$.input.farm_area",
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
			// Step 5: Process Inventory Update
			{
				StepNumber:    5,
				ServiceName:   "inventory",
				HandlerMethod: "ProcessInventoryUpdate",
				InputMapping: map[string]string{
					"cropPlanID":          "$.steps.1.result.crop_plan_id",
					"procurementResult":   "$.steps.4.result.procurement_result",
					"cropType":            "$.input.crop_type",
				},
				TimeoutSeconds:    25,
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
			// Step 6: Calculate Budget Requirement
			{
				StepNumber:    6,
				ServiceName:   "cost-center",
				HandlerMethod: "CalculateBudgetRequirement",
				InputMapping: map[string]string{
					"cropPlanID":         "$.steps.1.result.crop_plan_id",
					"allocationData":     "$.steps.3.result.allocation_data",
					"procurementCost":    "$.steps.4.result.procurement_cost",
					"farmArea":           "$.input.farm_area",
					"plantingSeason":     "$.input.planting_season",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Reserve Budget Allocation
			{
				StepNumber:    7,
				ServiceName:   "cost-center",
				HandlerMethod: "ReserveBudgetAllocation",
				InputMapping: map[string]string{
					"cropPlanID":       "$.steps.1.result.crop_plan_id",
					"budgetAmount":     "$.steps.6.result.budget_amount",
					"plantingSeason":   "$.input.planting_season",
				},
				TimeoutSeconds:    20,
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
			// Step 8: Apply Ledger Entries
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "ApplyCropPlanningJournal",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"cropPlanID":       "$.steps.1.result.crop_plan_id",
					"budgetAmount":     "$.steps.6.result.budget_amount",
					"procurementCost":  "$.steps.4.result.procurement_cost",
					"journalDate":      "$.input.planting_season",
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
			// Step 9: Confirm Crop Planning
			{
				StepNumber:    9,
				ServiceName:   "agriculture",
				HandlerMethod: "ConfirmCropPlanning",
				InputMapping: map[string]string{
					"cropPlanID":         "$.steps.1.result.crop_plan_id",
					"budgetReservation":  "$.steps.7.result.budget_reservation",
					"journalEntries":     "$.steps.8.result.journal_entries",
					"completionStatus":   "Confirmed",
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

			// Step 101: Revert Farm Area Validation (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "crop-planning",
				HandlerMethod: "RevertFarmAreaValidation",
				InputMapping: map[string]string{
					"cropPlanID": "$.steps.1.result.crop_plan_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 102: Deallocate Farm Resources (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "crop-planning",
				HandlerMethod: "DeallocateResources",
				InputMapping: map[string]string{
					"cropPlanID": "$.steps.1.result.crop_plan_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Cancel Procurement (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "procurement",
				HandlerMethod: "CancelSeedsFertilizerProcurement",
				InputMapping: map[string]string{
					"cropPlanID": "$.steps.1.result.crop_plan_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 104: Revert Inventory Update (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "inventory",
				HandlerMethod: "RevertInventoryUpdate",
				InputMapping: map[string]string{
					"cropPlanID": "$.steps.1.result.crop_plan_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Clear Budget Calculation (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "cost-center",
				HandlerMethod: "ClearBudgetCalculation",
				InputMapping: map[string]string{
					"cropPlanID": "$.steps.1.result.crop_plan_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Release Budget Reservation (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "cost-center",
				HandlerMethod: "ReleaseBudgetReservation",
				InputMapping: map[string]string{
					"cropPlanID": "$.steps.1.result.crop_plan_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: Reverse Journal Entries (compensates step 8)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseCropPlanningJournal",
				InputMapping: map[string]string{
					"cropPlanID": "$.steps.1.result.crop_plan_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *CropPlanningSaga) SagaType() string {
	return "SAGA-A01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *CropPlanningSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *CropPlanningSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *CropPlanningSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["crop_plan_id"] == nil {
		return errors.New("crop_plan_id is required")
	}

	if inputMap["crop_type"] == nil {
		return errors.New("crop_type is required")
	}

	if inputMap["farm_area"] == nil {
		return errors.New("farm_area is required")
	}

	if inputMap["planting_season"] == nil {
		return errors.New("planting_season is required (format: YYYY-MM)")
	}

	return nil
}
