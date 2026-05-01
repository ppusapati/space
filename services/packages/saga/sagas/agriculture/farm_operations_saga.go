// Package agriculture provides saga handlers for agricultural workflows
package agriculture

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// FarmOperationsSaga implements SAGA-A02: Farm Operations & Activity Tracking workflow
// Business Flow: LogActivityRecord → ValidateActivity → UpdateCropMonitoring → AllocateLaborResources → ProcessLaborCost → UpdateInventoryUsage → CalculateOperationCost → ApplyOperationJournal → UpdateCostCenter → CompleteFarmActivity
// Steps: 10 forward + 9 compensation = 19 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,7,10
type FarmOperationsSaga struct {
	steps []*saga.StepDefinition
}

// NewFarmOperationsSaga creates a new Farm Operations saga handler
func NewFarmOperationsSaga() saga.SagaHandler {
	return &FarmOperationsSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Log Activity Record
			{
				StepNumber:    1,
				ServiceName:   "agriculture",
				HandlerMethod: "LogActivityRecord",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"activityLogID":  "$.input.activity_log_id",
					"farmID":         "$.input.farm_id",
					"activityType":   "$.input.activity_type",
					"activityDate":   "$.input.activity_date",
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
			// Step 2: Validate Activity Details
			{
				StepNumber:    2,
				ServiceName:   "crop-monitoring",
				HandlerMethod: "ValidateActivity",
				InputMapping: map[string]string{
					"activityLogID": "$.steps.1.result.activity_log_id",
					"farmID":        "$.input.farm_id",
					"activityType":  "$.input.activity_type",
					"activityDate":  "$.input.activity_date",
					"validateRules": "true",
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
			// Step 3: Update Crop Monitoring Data
			{
				StepNumber:    3,
				ServiceName:   "crop-monitoring",
				HandlerMethod: "UpdateCropMonitoring",
				InputMapping: map[string]string{
					"activityLogID":       "$.steps.1.result.activity_log_id",
					"farmID":              "$.input.farm_id",
					"activityType":        "$.input.activity_type",
					"validationResult":    "$.steps.2.result.validation_result",
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
			// Step 4: Allocate Labor Resources
			{
				StepNumber:    4,
				ServiceName:   "labor-management",
				HandlerMethod: "AllocateLaborResources",
				InputMapping: map[string]string{
					"activityLogID":    "$.steps.1.result.activity_log_id",
					"farmID":           "$.input.farm_id",
					"activityType":     "$.input.activity_type",
					"monitoringData":   "$.steps.3.result.monitoring_data",
				},
				TimeoutSeconds:    30,
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
			// Step 5: Process Labor Cost
			{
				StepNumber:    5,
				ServiceName:   "labor-management",
				HandlerMethod: "ProcessLaborCost",
				InputMapping: map[string]string{
					"activityLogID":   "$.steps.1.result.activity_log_id",
					"laborAllocation": "$.steps.4.result.labor_allocation",
					"activityDate":    "$.input.activity_date",
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
			// Step 6: Update Inventory Usage
			{
				StepNumber:    6,
				ServiceName:   "inventory",
				HandlerMethod: "UpdateInventoryUsage",
				InputMapping: map[string]string{
					"activityLogID": "$.steps.1.result.activity_log_id",
					"activityType":  "$.input.activity_type",
					"farmID":        "$.input.farm_id",
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
			// Step 7: Calculate Operation Cost
			{
				StepNumber:    7,
				ServiceName:   "cost-center",
				HandlerMethod: "CalculateOperationCost",
				InputMapping: map[string]string{
					"activityLogID":   "$.steps.1.result.activity_log_id",
					"laborCost":       "$.steps.5.result.labor_cost",
					"inventoryCost":   "$.steps.6.result.inventory_cost",
					"activityDate":    "$.input.activity_date",
				},
				TimeoutSeconds:    25,
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
			// Step 8: Apply Operation Journal Entries
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "ApplyFarmOperationJournal",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"activityLogID":   "$.steps.1.result.activity_log_id",
					"operationCost":   "$.steps.7.result.operation_cost",
					"journalDate":     "$.input.activity_date",
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
			// Step 9: Update Cost Center Records
			{
				StepNumber:    9,
				ServiceName:   "cost-center",
				HandlerMethod: "UpdateCostCenterRecords",
				InputMapping: map[string]string{
					"activityLogID": "$.steps.1.result.activity_log_id",
					"operationCost": "$.steps.7.result.operation_cost",
					"farmID":        "$.input.farm_id",
				},
				TimeoutSeconds:    20,
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
			// Step 10: Complete Farm Activity
			{
				StepNumber:    10,
				ServiceName:   "agriculture",
				HandlerMethod: "CompleteFarmActivity",
				InputMapping: map[string]string{
					"activityLogID":      "$.steps.1.result.activity_log_id",
					"journalEntries":     "$.steps.8.result.journal_entries",
					"operationCost":      "$.steps.7.result.operation_cost",
					"completionStatus":   "Completed",
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

			// Step 101: Revert Activity Validation (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "crop-monitoring",
				HandlerMethod: "RevertActivityValidation",
				InputMapping: map[string]string{
					"activityLogID": "$.steps.1.result.activity_log_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Revert Crop Monitoring Update (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "crop-monitoring",
				HandlerMethod: "RevertCropMonitoringUpdate",
				InputMapping: map[string]string{
					"activityLogID": "$.steps.1.result.activity_log_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Deallocate Labor Resources (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "labor-management",
				HandlerMethod: "DeallocateLaborResources",
				InputMapping: map[string]string{
					"activityLogID": "$.steps.1.result.activity_log_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Revert Labor Cost Processing (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "labor-management",
				HandlerMethod: "RevertLaborCostProcessing",
				InputMapping: map[string]string{
					"activityLogID": "$.steps.1.result.activity_log_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Revert Inventory Usage Update (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "inventory",
				HandlerMethod: "RevertInventoryUsageUpdate",
				InputMapping: map[string]string{
					"activityLogID": "$.steps.1.result.activity_log_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Clear Operation Cost Calculation (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "cost-center",
				HandlerMethod: "ClearOperationCostCalculation",
				InputMapping: map[string]string{
					"activityLogID": "$.steps.1.result.activity_log_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 107: Reverse Operation Journal Entries (compensates step 8)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseFarmOperationJournal",
				InputMapping: map[string]string{
					"activityLogID": "$.steps.1.result.activity_log_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: Revert Cost Center Update (compensates step 9)
			{
				StepNumber:    108,
				ServiceName:   "cost-center",
				HandlerMethod: "RevertCostCenterUpdate",
				InputMapping: map[string]string{
					"activityLogID": "$.steps.1.result.activity_log_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *FarmOperationsSaga) SagaType() string {
	return "SAGA-A02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *FarmOperationsSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *FarmOperationsSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *FarmOperationsSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["activity_log_id"] == nil {
		return errors.New("activity_log_id is required")
	}

	if inputMap["farm_id"] == nil {
		return errors.New("farm_id is required")
	}

	if inputMap["activity_type"] == nil {
		return errors.New("activity_type is required")
	}

	if inputMap["activity_date"] == nil {
		return errors.New("activity_date is required (format: YYYY-MM-DD)")
	}

	return nil
}
