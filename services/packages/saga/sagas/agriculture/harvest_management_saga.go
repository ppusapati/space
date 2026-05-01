// Package agriculture provides saga handlers for agricultural workflows
package agriculture

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// HarvestManagementSaga implements SAGA-A03: Harvest & Post-Harvest Management workflow
// Business Flow: InitiateHarvest → ValidateHarvestReadiness → ScheduleHarvestActivities → ConductQualityInspection → ProcessHarvestYield → UpdateInventoryWithHarvest → ProcessPostHarvestHandling → CalculateHarvestCost → StorageAllocation → ApplyHarvestJournal → CompleteHarvestOperation
// Steps: 11 forward + 10 compensation = 21 total (longest agriculture saga)
// Timeout: 180 seconds, Critical steps: 1,2,3,4,6,8,11
type HarvestManagementSaga struct {
	steps []*saga.StepDefinition
}

// NewHarvestManagementSaga creates a new Harvest Management saga handler
func NewHarvestManagementSaga() saga.SagaHandler {
	return &HarvestManagementSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Harvest
			{
				StepNumber:    1,
				ServiceName:   "agriculture",
				HandlerMethod: "InitiateHarvest",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"harvestID":    "$.input.harvest_id",
					"farmID":       "$.input.farm_id",
					"cropType":     "$.input.crop_type",
					"harvestDate":  "$.input.harvest_date",
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
			// Step 2: Validate Harvest Readiness
			{
				StepNumber:    2,
				ServiceName:   "harvest-management",
				HandlerMethod: "ValidateHarvestReadiness",
				InputMapping: map[string]string{
					"harvestID":      "$.steps.1.result.harvest_id",
					"farmID":         "$.input.farm_id",
					"cropType":       "$.input.crop_type",
					"validateCrop":   "true",
					"harvestDate":    "$.input.harvest_date",
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
			// Step 3: Schedule Harvest Activities
			{
				StepNumber:    3,
				ServiceName:   "harvest-management",
				HandlerMethod: "ScheduleHarvestActivities",
				InputMapping: map[string]string{
					"harvestID":          "$.steps.1.result.harvest_id",
					"cropType":           "$.input.crop_type",
					"readinessData":      "$.steps.2.result.readiness_data",
					"harvestDate":        "$.input.harvest_date",
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
			// Step 4: Conduct Quality Inspection
			{
				StepNumber:    4,
				ServiceName:   "quality-inspection",
				HandlerMethod: "ConductHarvestInspection",
				InputMapping: map[string]string{
					"harvestID":      "$.steps.1.result.harvest_id",
					"cropType":       "$.input.crop_type",
					"farmID":         "$.input.farm_id",
					"harvestSchedule": "$.steps.3.result.harvest_schedule",
				},
				TimeoutSeconds:    40,
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
			// Step 5: Process Harvest Yield
			{
				StepNumber:    5,
				ServiceName:   "harvest-management",
				HandlerMethod: "ProcessHarvestYield",
				InputMapping: map[string]string{
					"harvestID":         "$.steps.1.result.harvest_id",
					"cropType":          "$.input.crop_type",
					"inspectionResult":  "$.steps.4.result.inspection_result",
					"harvestDate":       "$.input.harvest_date",
				},
				TimeoutSeconds:    35,
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
			// Step 6: Update Inventory with Harvest Yield
			{
				StepNumber:    6,
				ServiceName:   "inventory",
				HandlerMethod: "UpdateInventoryWithHarvest",
				InputMapping: map[string]string{
					"harvestID":      "$.steps.1.result.harvest_id",
					"cropType":       "$.input.crop_type",
					"harvestYield":   "$.steps.5.result.harvest_yield",
					"farmID":         "$.input.farm_id",
				},
				TimeoutSeconds:    30,
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
			// Step 7: Process Post-Harvest Handling
			{
				StepNumber:    7,
				ServiceName:   "harvest-management",
				HandlerMethod: "ProcessPostHarvestHandling",
				InputMapping: map[string]string{
					"harvestID":      "$.steps.1.result.harvest_id",
					"harvestYield":   "$.steps.5.result.harvest_yield",
					"cropType":       "$.input.crop_type",
				},
				TimeoutSeconds:    40,
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
			// Step 8: Calculate Harvest Cost
			{
				StepNumber:    8,
				ServiceName:   "cost-center",
				HandlerMethod: "CalculateHarvestCost",
				InputMapping: map[string]string{
					"harvestID":      "$.steps.1.result.harvest_id",
					"harvestYield":   "$.steps.5.result.harvest_yield",
					"costData":       "$.steps.7.result.cost_data",
					"harvestDate":    "$.input.harvest_date",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Storage Allocation and Management
			{
				StepNumber:    9,
				ServiceName:   "storage",
				HandlerMethod: "AllocateStorageForHarvest",
				InputMapping: map[string]string{
					"harvestID":      "$.steps.1.result.harvest_id",
					"harvestYield":   "$.steps.5.result.harvest_yield",
					"cropType":       "$.input.crop_type",
				},
				TimeoutSeconds:    35,
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
			// Step 10: Apply Harvest Journal Entries
			{
				StepNumber:    10,
				ServiceName:   "general-ledger",
				HandlerMethod: "ApplyHarvestJournal",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"harvestID":       "$.steps.1.result.harvest_id",
					"harvestCost":     "$.steps.8.result.harvest_cost",
					"harvestYield":    "$.steps.5.result.harvest_yield",
					"journalDate":     "$.input.harvest_date",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 11: Complete Harvest Operation
			{
				StepNumber:    11,
				ServiceName:   "agriculture",
				HandlerMethod: "CompleteHarvestOperation",
				InputMapping: map[string]string{
					"harvestID":       "$.steps.1.result.harvest_id",
					"journalEntries":  "$.steps.10.result.journal_entries",
					"storageData":     "$.steps.9.result.storage_data",
					"completionStatus": "Completed",
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

			// Step 101: Revert Harvest Readiness Validation (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "harvest-management",
				HandlerMethod: "RevertHarvestReadinessValidation",
				InputMapping: map[string]string{
					"harvestID": "$.steps.1.result.harvest_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Cancel Harvest Schedule (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "harvest-management",
				HandlerMethod: "CancelHarvestSchedule",
				InputMapping: map[string]string{
					"harvestID": "$.steps.1.result.harvest_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Revert Quality Inspection (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "quality-inspection",
				HandlerMethod: "RevertHarvestInspection",
				InputMapping: map[string]string{
					"harvestID": "$.steps.1.result.harvest_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
			// Step 104: Clear Harvest Yield Processing (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "harvest-management",
				HandlerMethod: "ClearHarvestYieldProcessing",
				InputMapping: map[string]string{
					"harvestID": "$.steps.1.result.harvest_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 105: Revert Inventory Harvest Update (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "inventory",
				HandlerMethod: "RevertHarvestInventoryUpdate",
				InputMapping: map[string]string{
					"harvestID": "$.steps.1.result.harvest_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: Revert Post-Harvest Processing (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "harvest-management",
				HandlerMethod: "RevertPostHarvestProcessing",
				InputMapping: map[string]string{
					"harvestID": "$.steps.1.result.harvest_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
			// Step 107: Clear Harvest Cost Calculation (compensates step 8)
			{
				StepNumber:    107,
				ServiceName:   "cost-center",
				HandlerMethod: "ClearHarvestCostCalculation",
				InputMapping: map[string]string{
					"harvestID": "$.steps.1.result.harvest_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: Release Storage Allocation (compensates step 9)
			{
				StepNumber:    108,
				ServiceName:   "storage",
				HandlerMethod: "ReleaseStorageAllocation",
				InputMapping: map[string]string{
					"harvestID": "$.steps.1.result.harvest_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 109: Reverse Harvest Journal Entries (compensates step 10)
			{
				StepNumber:    109,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseHarvestJournal",
				InputMapping: map[string]string{
					"harvestID": "$.steps.1.result.harvest_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 110: Cancel Harvest Initiation (compensates step 1)
			{
				StepNumber:    110,
				ServiceName:   "agriculture",
				HandlerMethod: "CancelHarvestInitiation",
				InputMapping: map[string]string{
					"harvestID": "$.steps.1.result.harvest_id",
					"reason":    "Saga compensation - Harvest operation failed",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *HarvestManagementSaga) SagaType() string {
	return "SAGA-A03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *HarvestManagementSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *HarvestManagementSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *HarvestManagementSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["harvest_id"] == nil {
		return errors.New("harvest_id is required")
	}

	if inputMap["farm_id"] == nil {
		return errors.New("farm_id is required")
	}

	if inputMap["crop_type"] == nil {
		return errors.New("crop_type is required")
	}

	if inputMap["harvest_date"] == nil {
		return errors.New("harvest_date is required (format: YYYY-MM-DD)")
	}

	return nil
}
