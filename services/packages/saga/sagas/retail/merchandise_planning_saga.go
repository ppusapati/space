// Package retail provides saga handlers for retail workflows
package retail

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// MerchandisePlanningallocation implements SAGA-R09: Merchandise Planning & Allocation workflow
// Business Flow: InitiatePlan → ValidatePlanningPeriod → CalculateDemandForecasting → BudgetAllocation → CreateProcurementPlan → AllocateToLocations → UpdateInventoryTargets → GeneratePlanningJournal → ConfirmMerchandisePlan
// Steps: 9 forward + 10 compensation = 19 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,7,10
type MerchandisePlanningallocation struct {
	steps []*saga.StepDefinition
}

// NewMerchandisePlanningallocation creates a new Merchandise Planning saga handler
func NewMerchandisePlanningallocation() saga.SagaHandler {
	return &MerchandisePlanningallocation{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Plan
			{
				StepNumber:    1,
				ServiceName:   "merchandise-planning",
				HandlerMethod: "InitiatePlan",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"planID":          "$.input.plan_id",
					"planningPeriod":  "$.input.planning_period",
					"budgetAmount":    "$.input.budget_amount",
					"planStartDate":   "$.input.plan_start_date",
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
			// Step 2: Validate Planning Period
			{
				StepNumber:    2,
				ServiceName:   "merchandise-planning",
				HandlerMethod: "ValidatePlanningPeriod",
				InputMapping: map[string]string{
					"planID":         "$.steps.1.result.plan_id",
					"planningPeriod": "$.input.planning_period",
					"planStartDate":  "$.input.plan_start_date",
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
			// Step 3: Calculate Demand Forecasting
			{
				StepNumber:    3,
				ServiceName:   "merchandise-planning",
				HandlerMethod: "CalculateDemandForecasting",
				InputMapping: map[string]string{
					"planID":          "$.steps.1.result.plan_id",
					"planningPeriod":  "$.input.planning_period",
					"historicalData":  "$.input.historical_data",
				},
				TimeoutSeconds:    30,
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
			// Step 4: Budget Allocation
			{
				StepNumber:    4,
				ServiceName:   "merchandise-planning",
				HandlerMethod: "BudgetAllocation",
				InputMapping: map[string]string{
					"planID":           "$.steps.1.result.plan_id",
					"budgetAmount":     "$.input.budget_amount",
					"demandForecast":   "$.steps.3.result.demand_forecast",
				},
				TimeoutSeconds:    25,
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
			// Step 5: Create Procurement Plan
			{
				StepNumber:    5,
				ServiceName:   "procurement",
				HandlerMethod: "CreateProcurementPlan",
				InputMapping: map[string]string{
					"planID":         "$.steps.1.result.plan_id",
					"budgetAllocation": "$.steps.4.result.budget_allocation",
					"demandForecast": "$.steps.3.result.demand_forecast",
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
			// Step 6: Allocate to Locations
			{
				StepNumber:    6,
				ServiceName:   "inventory",
				HandlerMethod: "AllocateToLocations",
				InputMapping: map[string]string{
					"planID":         "$.steps.1.result.plan_id",
					"budgetAllocation": "$.steps.4.result.budget_allocation",
					"locationList":   "$.input.location_list",
				},
				TimeoutSeconds:    25,
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
			// Step 7: Update Inventory Targets
			{
				StepNumber:    7,
				ServiceName:   "inventory",
				HandlerMethod: "UpdateInventoryTargets",
				InputMapping: map[string]string{
					"planID":            "$.steps.1.result.plan_id",
					"locationAllocation": "$.steps.6.result.location_allocation",
				},
				TimeoutSeconds:    25,
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
			// Step 8: Generate Planning Journal
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "GeneratePlanningJournal",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"planID":           "$.steps.1.result.plan_id",
					"budgetAmount":     "$.input.budget_amount",
					"budgetAllocation": "$.steps.4.result.budget_allocation",
					"journalDate":      "$.input.plan_start_date",
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
			// Step 9: Confirm Merchandise Plan
			{
				StepNumber:    9,
				ServiceName:   "merchandise-planning",
				HandlerMethod: "ConfirmMerchandisePlan",
				InputMapping: map[string]string{
					"planID":          "$.steps.1.result.plan_id",
					"journalEntries":  "$.steps.8.result.journal_entries",
					"planStatus":      "Confirmed",
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

			// Step 101: Revert Budget Allocation (compensates step 4)
			{
				StepNumber:    101,
				ServiceName:   "merchandise-planning",
				HandlerMethod: "RevertBudgetAllocation",
				InputMapping: map[string]string{
					"planID": "$.steps.1.result.plan_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 102: Revert Procurement Plan Creation (compensates step 5)
			{
				StepNumber:    102,
				ServiceName:   "procurement",
				HandlerMethod: "RevertProcurementPlanCreation",
				InputMapping: map[string]string{
					"planID": "$.steps.1.result.plan_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Revert Location Allocation (compensates step 6)
			{
				StepNumber:    103,
				ServiceName:   "inventory",
				HandlerMethod: "RevertLocationAllocation",
				InputMapping: map[string]string{
					"planID": "$.steps.1.result.plan_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 104: Revert Inventory Target Updates (compensates step 7)
			{
				StepNumber:    104,
				ServiceName:   "inventory",
				HandlerMethod: "RevertInventoryTargetUpdates",
				InputMapping: map[string]string{
					"planID": "$.steps.1.result.plan_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Reverse Planning Journal (compensates step 8)
			{
				StepNumber:    105,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReversePlanningJournal",
				InputMapping: map[string]string{
					"planID": "$.steps.1.result.plan_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: Revert Initiate Plan (compensates step 1)
			{
				StepNumber:    106,
				ServiceName:   "merchandise-planning",
				HandlerMethod: "RevertInitiatePlan",
				InputMapping: map[string]string{
					"planID": "$.steps.1.result.plan_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: Revert Validate Planning Period (compensates step 2)
			{
				StepNumber:    107,
				ServiceName:   "merchandise-planning",
				HandlerMethod: "RevertValidatePlanningPeriod",
				InputMapping: map[string]string{
					"planID": "$.steps.1.result.plan_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 108: Revert Calculate Demand Forecasting (compensates step 3)
			{
				StepNumber:    108,
				ServiceName:   "merchandise-planning",
				HandlerMethod: "RevertCalculateDemandForecasting",
				InputMapping: map[string]string{
					"planID": "$.steps.1.result.plan_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 109: Revert Confirm Merchandise Plan (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "merchandise-planning",
				HandlerMethod: "RevertConfirmMerchandisePlan",
				InputMapping: map[string]string{
					"planID": "$.steps.1.result.plan_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 110: Revert Procurement Plan Impact (additional compensation)
			{
				StepNumber:    110,
				ServiceName:   "procurement",
				HandlerMethod: "RevertProcurementPlanImpact",
				InputMapping: map[string]string{
					"planID": "$.steps.1.result.plan_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *MerchandisePlanningallocation) SagaType() string {
	return "SAGA-R09"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *MerchandisePlanningallocation) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *MerchandisePlanningallocation) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *MerchandisePlanningallocation) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["plan_id"] == nil {
		return errors.New("plan_id is required")
	}

	if inputMap["planning_period"] == nil {
		return errors.New("planning_period is required")
	}

	if inputMap["budget_amount"] == nil {
		return errors.New("budget_amount is required")
	}

	return nil
}
