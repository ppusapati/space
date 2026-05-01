// Package construction provides saga handlers for construction module workflows
package construction

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// SiteExpensesSaga implements SAGA-C06: Construction Site Expenses & Cost Control
// Business Flow: CaptureExpenseTransaction → CategorizeExpense → ValidateExpenseAmount → UpdateCostCenter → CheckBudgetCompliance → ApproveExpense → PostExpenseEntry → RecordExpenseForPayment → UpdateProjectCostAnalysis → FinalizeExpenseRecording
// Timeout: 120 seconds, Critical steps: 1,2,3,4,6,9
type SiteExpensesSaga struct {
	steps []*saga.StepDefinition
}

// NewSiteExpensesSaga creates a new Construction Site Expenses & Cost Control saga handler
func NewSiteExpensesSaga() saga.SagaHandler {
	return &SiteExpensesSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Capture Expense Transaction
			{
				StepNumber:    1,
				ServiceName:   "construction-site",
				HandlerMethod: "CaptureExpenseTransaction",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"expenseID":      "$.input.expense_id",
					"projectID":      "$.input.project_id",
					"expenseAmount":  "$.input.expense_amount",
					"expenseDate":    "$.input.expense_date",
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
			// Step 2: Categorize Expense
			{
				StepNumber:    2,
				ServiceName:   "cost-center",
				HandlerMethod: "CategorizeExpense",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"expenseID":         "$.input.expense_id",
					"projectID":         "$.input.project_id",
					"expenseAmount":     "$.input.expense_amount",
					"expenseTransaction": "$.steps.1.result.expense_transaction",
				},
				TimeoutSeconds:    25,
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
			// Step 3: Validate Expense Amount
			{
				StepNumber:    3,
				ServiceName:   "construction-site",
				HandlerMethod: "ValidateExpenseAmount",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"expenseID":             "$.input.expense_id",
					"projectID":             "$.input.project_id",
					"expenseAmount":         "$.input.expense_amount",
					"expenseCategory":       "$.steps.2.result.expense_category",
					"expenseTransaction":    "$.steps.1.result.expense_transaction",
				},
				TimeoutSeconds:    25,
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
			// Step 4: Update Cost Center
			{
				StepNumber:    4,
				ServiceName:   "cost-center",
				HandlerMethod: "UpdateCostCenter",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"projectID":           "$.input.project_id",
					"expenseID":           "$.input.expense_id",
					"expenseAmount":       "$.input.expense_amount",
					"expenseCategory":     "$.steps.2.result.expense_category",
					"validatedExpense":    "$.steps.3.result.validated_expense",
				},
				TimeoutSeconds:    30,
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
			// Step 5: Check Budget Compliance
			{
				StepNumber:    5,
				ServiceName:   "budget-control",
				HandlerMethod: "CheckBudgetCompliance",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"projectID":         "$.input.project_id",
					"expenseID":         "$.input.expense_id",
					"expenseAmount":     "$.input.expense_amount",
					"costCenterUpdate":  "$.steps.4.result.cost_center_update",
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
			// Step 6: Approve Expense
			{
				StepNumber:    6,
				ServiceName:   "approval",
				HandlerMethod: "ApproveExpense",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"expenseID":            "$.input.expense_id",
					"projectID":            "$.input.project_id",
					"expenseAmount":        "$.input.expense_amount",
					"budgetCompliance":     "$.steps.5.result.compliance_status",
				},
				TimeoutSeconds:    30,
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
			// Step 7: Post Expense Entry
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostExpenseEntry",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"expenseID":           "$.input.expense_id",
					"projectID":           "$.input.project_id",
					"expenseAmount":       "$.input.expense_amount",
					"expenseDate":         "$.input.expense_date",
					"expenseCategory":     "$.steps.2.result.expense_category",
					"approvalStatus":      "$.steps.6.result.approval_status",
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
			// Step 8: Record Expense For Payment
			{
				StepNumber:    8,
				ServiceName:   "construction-site",
				HandlerMethod: "RecordExpenseForPayment",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"expenseID":      "$.input.expense_id",
					"projectID":      "$.input.project_id",
					"expenseAmount":  "$.input.expense_amount",
					"journalEntry":   "$.steps.7.result.journal_entry",
				},
				TimeoutSeconds:    30,
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
			// Step 9: Update Project Cost Analysis
			{
				StepNumber:    9,
				ServiceName:   "project-costing",
				HandlerMethod: "UpdateProjectCostAnalysis",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"projectID":          "$.input.project_id",
					"expenseID":          "$.input.expense_id",
					"expenseAmount":      "$.input.expense_amount",
					"expenseCategory":    "$.steps.2.result.expense_category",
					"costCenterUpdate":   "$.steps.4.result.cost_center_update",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Finalize Expense Recording
			{
				StepNumber:    10,
				ServiceName:   "construction-site",
				HandlerMethod: "FinalizeExpenseRecording",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"expenseID":          "$.input.expense_id",
					"projectID":          "$.input.project_id",
					"expenseAmount":      "$.input.expense_amount",
					"expenseTransaction": "$.steps.1.result.expense_transaction",
					"costAnalysis":       "$.steps.9.result.cost_analysis",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
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

			// Step 102: ReverseCategoryAssignment (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "cost-center",
				HandlerMethod: "ReverseCategoryAssignment",
				InputMapping: map[string]string{
					"expenseID":       "$.input.expense_id",
					"expenseCategory": "$.steps.2.result.expense_category",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: ReverseExpenseValidation (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "construction-site",
				HandlerMethod: "ReverseExpenseValidation",
				InputMapping: map[string]string{
					"expenseID":         "$.input.expense_id",
					"validatedExpense": "$.steps.3.result.validated_expense",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 104: ReverseCostCenterUpdate (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "cost-center",
				HandlerMethod: "ReverseCostCenterUpdate",
				InputMapping: map[string]string{
					"projectID":           "$.input.project_id",
					"expenseID":           "$.input.expense_id",
					"costCenterUpdate":    "$.steps.4.result.cost_center_update",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: ReverseComplianceCheck (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "budget-control",
				HandlerMethod: "ReverseComplianceCheck",
				InputMapping: map[string]string{
					"projectID":        "$.input.project_id",
					"expenseID":        "$.input.expense_id",
					"complianceStatus": "$.steps.5.result.compliance_status",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: RejectExpenseApproval (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "approval",
				HandlerMethod: "RejectExpenseApproval",
				InputMapping: map[string]string{
					"expenseID":       "$.input.expense_id",
					"approvalStatus": "$.steps.6.result.approval_status",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: ReverseExpensePosting (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseExpensePosting",
				InputMapping: map[string]string{
					"expenseID":     "$.input.expense_id",
					"journalEntry": "$.steps.7.result.journal_entry",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: CancelPaymentRecord (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "construction-site",
				HandlerMethod: "CancelPaymentRecord",
				InputMapping: map[string]string{
					"expenseID":     "$.input.expense_id",
					"paymentRecord": "$.steps.8.result.payment_record",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 109: ReverseCostAnalysisUpdate (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "project-costing",
				HandlerMethod: "ReverseCostAnalysisUpdate",
				InputMapping: map[string]string{
					"projectID":      "$.input.project_id",
					"expenseID":      "$.input.expense_id",
					"costAnalysis":   "$.steps.9.result.cost_analysis",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *SiteExpensesSaga) SagaType() string {
	return "SAGA-C06"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *SiteExpensesSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *SiteExpensesSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *SiteExpensesSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["expense_id"] == nil {
		return errors.New("expense_id is required")
	}

	expenseID, ok := inputMap["expense_id"].(string)
	if !ok || expenseID == "" {
		return errors.New("expense_id must be a non-empty string")
	}

	if inputMap["project_id"] == nil {
		return errors.New("project_id is required")
	}

	projectID, ok := inputMap["project_id"].(string)
	if !ok || projectID == "" {
		return errors.New("project_id must be a non-empty string")
	}

	if inputMap["expense_amount"] == nil {
		return errors.New("expense_amount is required")
	}

	expenseAmount, ok := inputMap["expense_amount"].(string)
	if !ok || expenseAmount == "" {
		return errors.New("expense_amount must be a non-empty string")
	}

	if inputMap["expense_date"] == nil {
		return errors.New("expense_date is required")
	}

	expenseDate, ok := inputMap["expense_date"].(string)
	if !ok || expenseDate == "" {
		return errors.New("expense_date must be a non-empty string")
	}

	return nil
}
