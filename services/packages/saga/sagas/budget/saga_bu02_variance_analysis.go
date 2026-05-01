// Package budget provides saga handlers for budget management workflows
package budget

import (
	"errors"
	"fmt"

	"p9e.in/samavaya/packages/saga"
)

// VarianceAnalysisSaga implements SAGA-BU02: Variance Analysis & Budget Review
// Business Flow: 8 steps for variance calculation, analysis, and investigation
// Variance Management: Extract actuals, calculate variance, threshold checking, alerts
// Critical Steps: 3, 5, 7
// Timeout: 90 seconds
type VarianceAnalysisSaga struct {
	steps []*saga.StepDefinition
}

// NewVarianceAnalysisSaga creates a new Variance Analysis & Budget Review saga handler
func NewVarianceAnalysisSaga() saga.SagaHandler {
	return &VarianceAnalysisSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Extract Actual Spends from GL - Aggregate GL transactions
			{
				StepNumber:    1,
				ServiceName:   "general-ledger",
				HandlerMethod: "ExtractActualSpends",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"budgetID":      "$.input.budget_id",
					"startDate":     "$.input.start_date",
					"endDate":       "$.input.end_date",
					"costCenters":   "$.input.cost_centers",
					"department":    "$.input.department",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{108},
			},
			// Step 2: Get Budgeted Amounts - Retrieve budget allocations
			{
				StepNumber:    2,
				ServiceName:   "budget",
				HandlerMethod: "GetBudgetedAmounts",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"companyID":   "$.companyID",
					"branchID":    "$.branchID",
					"budgetID":    "$.input.budget_id",
					"startDate":   "$.input.start_date",
					"endDate":     "$.input.end_date",
					"costCenters": "$.input.cost_centers",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{109},
			},
			// Step 3: Calculate Variance - Actual - Budget (Unfavorable if positive) - CRITICAL
			{
				StepNumber:    3,
				ServiceName:   "budget",
				HandlerMethod: "CalculateVariance",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"budgetID":        "$.input.budget_id",
					"actualSpends":    "$.steps.1.result.actual_spends",
					"budgetedAmounts": "$.steps.2.result.budgeted_amounts",
					"costCenters":     "$.steps.2.result.cost_centers",
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
				CompensationSteps: []int32{110},
			},
			// Step 4: Calculate Variance Percentage - Variance / Budget * 100
			{
				StepNumber:    4,
				ServiceName:   "budget",
				HandlerMethod: "CalculateVariancePercentage",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"budgetID":       "$.input.budget_id",
					"varianceAmount": "$.steps.3.result.variance_amount",
					"budgetedAmounts": "$.steps.2.result.budgeted_amounts",
					"costCenters":    "$.steps.3.result.cost_centers",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{111},
			},
			// Step 5: Check if Exceeds Threshold (10%) - Determine if alert needed - CRITICAL
			{
				StepNumber:    5,
				ServiceName:   "budget",
				HandlerMethod: "CheckVarianceThreshold",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"budgetID":              "$.input.budget_id",
					"variancePercentages":   "$.steps.4.result.variance_percentages",
					"thresholdPercentage":   "$.input.threshold_percentage",
					"costCenters":           "$.steps.4.result.cost_centers",
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
				CompensationSteps: []int32{112},
			},
			// Step 6: Route Alert if Threshold Exceeded - Send notifications to stakeholders
			{
				StepNumber:    6,
				ServiceName:   "notification",
				HandlerMethod: "RouteVarianceAlert",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"budgetID":             "$.input.budget_id",
					"exceedsThreshold":     "$.steps.5.result.exceeds_threshold",
					"variancePercentages":  "$.steps.4.result.variance_percentages",
					"costCentersExceeding": "$.steps.5.result.cost_centers_exceeding",
					"department":           "$.input.department",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        2,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      10000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},
			// Step 7: Create Investigation Ticket - Create investigation task for variance - CRITICAL
			{
				StepNumber:    7,
				ServiceName:   "budget",
				HandlerMethod: "CreateInvestigationTicket",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"budgetID":             "$.input.budget_id",
					"varianceAmount":       "$.steps.3.result.variance_amount",
					"variancePercentages":  "$.steps.4.result.variance_percentages",
					"costCentersExceeding": "$.steps.5.result.cost_centers_exceeding",
					"assignedTo":           "$.input.assigned_to",
					"dueDate":              "$.input.due_date",
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
				CompensationSteps: []int32{113},
			},
			// Step 8: Archive Variance Record - Store variance data for reporting
			{
				StepNumber:    8,
				ServiceName:   "budget",
				HandlerMethod: "ArchiveVarianceRecord",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"budgetID":             "$.input.budget_id",
					"varianceAmount":       "$.steps.3.result.variance_amount",
					"variancePercentages":  "$.steps.4.result.variance_percentages",
					"actualSpends":         "$.steps.1.result.actual_spends",
					"budgetedAmounts":      "$.steps.2.result.budgeted_amounts",
					"investigationID":      "$.steps.7.result.investigation_id",
					"archiveDate":          "$.input.archive_date",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{114},
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *VarianceAnalysisSaga) SagaType() string {
	return "SAGA-BU02"
}

// GetStepDefinitions returns all step definitions for this saga
func (s *VarianceAnalysisSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *VarianceAnalysisSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: budget_id, start_date, end_date, cost_centers, threshold_percentage
func (s *VarianceAnalysisSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	// Extract the nested 'input' object
	innerInput, ok := inputMap["input"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing 'input' field in saga input")
	}

	// Validate budget_id
	if innerInput["budget_id"] == nil {
		return errors.New("missing required field: budget_id")
	}
	budgetID, ok := innerInput["budget_id"].(string)
	if !ok || budgetID == "" {
		return errors.New("budget_id must be a non-empty string")
	}

	// Validate start_date
	if innerInput["start_date"] == nil {
		return errors.New("missing required field: start_date")
	}
	startDate, ok := innerInput["start_date"].(string)
	if !ok || startDate == "" {
		return errors.New("start_date must be a non-empty string")
	}

	// Validate end_date
	if innerInput["end_date"] == nil {
		return errors.New("missing required field: end_date")
	}
	endDate, ok := innerInput["end_date"].(string)
	if !ok || endDate == "" {
		return errors.New("end_date must be a non-empty string")
	}

	// Validate cost_centers
	if innerInput["cost_centers"] == nil {
		return errors.New("missing required field: cost_centers")
	}
	costCenters, ok := innerInput["cost_centers"].([]interface{})
	if !ok || len(costCenters) == 0 {
		return errors.New("cost_centers must be a non-empty array")
	}

	// Validate threshold_percentage (default 10%)
	if innerInput["threshold_percentage"] != nil {
		thresholdPct, ok := innerInput["threshold_percentage"].(float64)
		if !ok || thresholdPct <= 0 || thresholdPct > 100 {
			return errors.New("threshold_percentage must be a number between 0 and 100")
		}
	}

	// Validate company_id (from context)
	if inputMap["companyID"] == nil {
		return errors.New("missing companyID in saga context")
	}

	return nil
}
