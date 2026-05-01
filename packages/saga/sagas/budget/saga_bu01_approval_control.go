// Package budget provides saga handlers for budget management workflows
package budget

import (
	"errors"
	"fmt"

	"p9e.in/samavaya/packages/saga"
)

// BudgetApprovalControlSaga implements SAGA-BU01: Budget Approval & Control
// Business Flow: 9 steps for complete budget creation from proposal to activation
// Budget Control: Multi-level approval (department, finance), GL posting, spending limits
// Critical Steps: 3, 4, 6, 8
// Timeout: 120 seconds
type BudgetApprovalControlSaga struct {
	steps []*saga.StepDefinition
}

// NewBudgetApprovalControlSaga creates a new Budget Approval & Control saga handler
func NewBudgetApprovalControlSaga() saga.SagaHandler {
	return &BudgetApprovalControlSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Budget Proposal - Initialize with department and cost centers
			{
				StepNumber:    1,
				ServiceName:   "budget",
				HandlerMethod: "CreateBudgetProposal",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"department":    "$.input.department",
					"budgetYear":    "$.input.budget_year",
					"totalAmount":   "$.input.total_amount",
					"currency":      "$.input.currency",
					"description":   "$.input.description",
					"createdBy":     "$.input.created_by",
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
				CompensationSteps: []int32{110},
			},
			// Step 2: Validate Against Historical Actuals - Compare with prior year data
			{
				StepNumber:    2,
				ServiceName:   "general-ledger",
				HandlerMethod: "ValidateHistoricalActuals",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"department":    "$.input.department",
					"budgetYear":    "$.input.budget_year",
					"budgetAmount":  "$.input.total_amount",
					"priorYears":    "$.input.prior_years",
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
				CompensationSteps: []int32{},
			},
			// Step 3: Submit to Department Manager - Department level approval - CRITICAL
			{
				StepNumber:    3,
				ServiceName:   "approval-workflow",
				HandlerMethod: "SubmitDepartmentApproval",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"budgetID":     "$.steps.1.result.budget_id",
					"department":   "$.input.department",
					"approverRole": "DEPT_MANAGER",
					"totalAmount":  "$.input.total_amount",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        2,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{111},
			},
			// Step 4: Submit to Finance Manager - Finance level approval - CRITICAL
			{
				StepNumber:    4,
				ServiceName:   "approval-workflow",
				HandlerMethod: "SubmitFinanceApproval",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"budgetID":     "$.steps.1.result.budget_id",
					"approverRole": "FINANCE_MANAGER",
					"deptApproval": "$.steps.3.result.approved",
					"totalAmount":  "$.input.total_amount",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        2,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{112},
			},
			// Step 5: Allocate to Cost Centers - Distribute budget across cost centers
			{
				StepNumber:    5,
				ServiceName:   "cost-center",
				HandlerMethod: "AllocateBudgetToCostCenters",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"budgetID":       "$.steps.1.result.budget_id",
					"costCenters":    "$.input.cost_centers",
					"allocations":    "$.input.allocations",
					"financeApproval": "$.steps.4.result.approved",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{113},
			},
			// Step 6: Post Budget to GL - Create budget GL entries - CRITICAL
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostBudgetEntries",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"budgetID":          "$.steps.1.result.budget_id",
					"budgetAmount":      "$.input.total_amount",
					"costCenterAllocations": "$.steps.5.result.allocations",
					"postingDate":       "$.input.posting_date",
					"budgetAccount":     "$.input.budget_account",
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
				CompensationSteps: []int32{114},
			},
			// Step 7: Configure Spending Limits - Set threshold rules and alerts
			{
				StepNumber:    7,
				ServiceName:   "budget",
				HandlerMethod: "ConfigureSpendingLimits",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"budgetID":         "$.steps.1.result.budget_id",
					"monthlyLimit":     "$.input.monthly_limit",
					"quarterlyLimit":   "$.input.quarterly_limit",
					"alertThreshold":   "$.input.alert_threshold",
					"approvalRequired": "$.input.approval_required_above",
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
				CompensationSteps: []int32{115},
			},
			// Step 8: Activate Budget - Set status to ACTIVE - CRITICAL
			{
				StepNumber:    8,
				ServiceName:   "budget",
				HandlerMethod: "ActivateBudget",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"budgetID":          "$.steps.1.result.budget_id",
					"glPostingComplete": "$.steps.6.result.gl_posting_complete",
					"activationDate":    "$.input.activation_date",
					"activatedBy":       "$.input.activated_by",
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
				CompensationSteps: []int32{116},
			},
			// Step 9: Send Activation Notification - Notify stakeholders
			{
				StepNumber:    9,
				ServiceName:   "notification",
				HandlerMethod: "SendBudgetActivationNotification",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"budgetID":     "$.steps.1.result.budget_id",
					"department":   "$.input.department",
					"totalAmount":  "$.input.total_amount",
					"budgetYear":   "$.input.budget_year",
					"activatedBy":  "$.input.activated_by",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        2,
					InitialBackoffMs:  500,
					MaxBackoffMs:      5000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *BudgetApprovalControlSaga) SagaType() string {
	return "SAGA-BU01"
}

// GetStepDefinitions returns all step definitions for this saga
func (s *BudgetApprovalControlSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *BudgetApprovalControlSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: department, budget_year, total_amount, currency, cost_centers
func (s *BudgetApprovalControlSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	// Extract the nested 'input' object
	innerInput, ok := inputMap["input"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing 'input' field in saga input")
	}

	// Validate department
	if innerInput["department"] == nil {
		return errors.New("missing required field: department")
	}
	department, ok := innerInput["department"].(string)
	if !ok || department == "" {
		return errors.New("department must be a non-empty string")
	}

	// Validate budget_year
	if innerInput["budget_year"] == nil {
		return errors.New("missing required field: budget_year")
	}
	budgetYear, ok := innerInput["budget_year"].(float64)
	if !ok || budgetYear <= 0 {
		return errors.New("budget_year must be a positive number")
	}

	// Validate total_amount
	if innerInput["total_amount"] == nil {
		return errors.New("missing required field: total_amount")
	}
	totalAmount, ok := innerInput["total_amount"].(float64)
	if !ok || totalAmount <= 0 {
		return errors.New("total_amount must be a positive number")
	}

	// Validate currency
	if innerInput["currency"] == nil {
		return errors.New("missing required field: currency")
	}
	currency, ok := innerInput["currency"].(string)
	if !ok || currency == "" {
		return errors.New("currency must be a non-empty string")
	}

	// Validate cost_centers
	if innerInput["cost_centers"] == nil {
		return errors.New("missing required field: cost_centers")
	}
	costCenters, ok := innerInput["cost_centers"].([]interface{})
	if !ok || len(costCenters) == 0 {
		return errors.New("cost_centers must be a non-empty array")
	}

	// Validate allocations
	if innerInput["allocations"] == nil {
		return errors.New("missing required field: allocations")
	}
	allocations, ok := innerInput["allocations"].([]interface{})
	if !ok || len(allocations) != len(costCenters) {
		return errors.New("allocations must have same length as cost_centers")
	}

	// Validate company_id (from context)
	if inputMap["companyID"] == nil {
		return errors.New("missing companyID in saga context")
	}

	return nil
}
