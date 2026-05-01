// Package inventory provides saga handlers for inventory module workflows
package inventory

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// DemandPlanningSaga implements SAGA-I05: Demand Planning & MRP workflow
// Business Flow: Calculate Forecast → Run MRP → Generate Requisitions → Allocate Stock → Update Parameters → Check Safety Stock → Notify Procurement
type DemandPlanningSaga struct {
	steps []*saga.StepDefinition
}

// NewDemandPlanningSaga creates a new Demand Planning saga handler
func NewDemandPlanningSaga() saga.SagaHandler {
	return &DemandPlanningSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Calculate Forecast
			{
				StepNumber:    1,
				ServiceName:   "planning",
				HandlerMethod: "CalculateForecast",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"planningHorizon": "$.input.planning_horizon",
					"forecastMethod":  "$.input.forecast_method",
					"historicalPeriods": "$.input.historical_periods",
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
			// Step 2: Run MRP
			{
				StepNumber:    2,
				ServiceName:   "planning",
				HandlerMethod: "RunMRP",
				InputMapping: map[string]string{
					"forecastID":      "$.steps.1.result.forecast_id",
					"planningHorizon": "$.input.planning_horizon",
					"includeSafetyStock": "true",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{102},
			},
			// Step 3: Generate Requisitions
			{
				StepNumber:    3,
				ServiceName:   "procurement",
				HandlerMethod: "GenerateRequisitions",
				InputMapping: map[string]string{
					"mrpRunID":  "$.steps.2.result.mrp_run_id",
					"forecastID": "$.steps.1.result.forecast_id",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{103},
			},
			// Step 4: Allocate Available Stock
			{
				StepNumber:    4,
				ServiceName:   "inventory-core",
				HandlerMethod: "AllocateAvailableStock",
				InputMapping: map[string]string{
					"mrpRunID":        "$.steps.2.result.mrp_run_id",
					"allocationMethod": "$.input.allocation_method",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{104},
			},
			// Step 5: Update Planning Parameters
			{
				StepNumber:    5,
				ServiceName:   "planning",
				HandlerMethod: "UpdatePlanningParameters",
				InputMapping: map[string]string{
					"mrpRunID":        "$.steps.2.result.mrp_run_id",
					"eoqUpdates":      "$.input.eoq_updates",
					"leadTimeUpdates": "$.input.lead_time_updates",
				},
				TimeoutSeconds:    20,
				IsCritical:        false, // Non-critical
				CompensationSteps: []int32{105},
			},
			// Step 6: Check Safety Stock
			{
				StepNumber:    6,
				ServiceName:   "planning",
				HandlerMethod: "CheckSafetyStock",
				InputMapping: map[string]string{
					"mrpRunID":      "$.steps.2.result.mrp_run_id",
					"forecastID":     "$.steps.1.result.forecast_id",
				},
				TimeoutSeconds:    25,
				IsCritical:        false, // Non-critical
				CompensationSteps: []int32{106},
			},
			// Step 7: Notify Procurement (non-critical)
			{
				StepNumber:    7,
				ServiceName:   "notification",
				HandlerMethod: "NotifyProcurement",
				InputMapping: map[string]string{
					"mrpRunID":     "$.steps.2.result.mrp_run_id",
					"requisitionCount": "$.steps.3.result.requisition_count",
					"notificationType": "MRP_COMPLETION",
				},
				TimeoutSeconds:    15,
				IsCritical:        false, // Non-critical
				CompensationSteps: []int32{},
			},
			// ===== COMPENSATION STEPS =====

			// Step 102: Revert MRP (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "planning",
				HandlerMethod: "RevertMRP",
				InputMapping: map[string]string{
					"mrpRunID": "$.steps.2.result.mrp_run_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Delete Generated Requisitions (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "procurement",
				HandlerMethod: "DeleteGeneratedRequisitions",
				InputMapping: map[string]string{
					"mrpRunID": "$.steps.2.result.mrp_run_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 104: Release Allocations (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "inventory-core",
				HandlerMethod: "ReleaseAllocations",
				InputMapping: map[string]string{
					"mrpRunID": "$.steps.2.result.mrp_run_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Revert Parameters (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "planning",
				HandlerMethod: "RevertParameters",
				InputMapping: map[string]string{
					"mrpRunID": "$.steps.2.result.mrp_run_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: Clear Safety Stock Flags (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "planning",
				HandlerMethod: "ClearSafetyStockFlags",
				InputMapping: map[string]string{
					"mrpRunID": "$.steps.2.result.mrp_run_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *DemandPlanningSaga) SagaType() string {
	return "SAGA-I05"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *DemandPlanningSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *DemandPlanningSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *DemandPlanningSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["planning_horizon"] == nil {
		return errors.New("planning_horizon is required")
	}

	if inputMap["forecast_method"] == nil {
		return errors.New("forecast_method is required")
	}

	if inputMap["historical_periods"] == nil {
		return errors.New("historical_periods is required")
	}

	return nil
}
