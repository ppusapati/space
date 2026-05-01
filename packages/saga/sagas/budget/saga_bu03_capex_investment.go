// Package budget provides saga handlers for budget management workflows
package budget

import (
	"errors"
	"fmt"

	"p9e.in/samavaya/packages/saga"
)

// CapExInvestmentSaga implements SAGA-BU03: CapEx Proposal & Investment Approval
// Business Flow: 11 steps for capital expenditure from proposal to asset creation and GL posting
// Capital Investment: Technical/financial eval, ROI/NPV calcs, multi-level approval, procurement
// Critical Steps: 5, 8, 11
// Timeout: 300 seconds
type CapExInvestmentSaga struct {
	steps []*saga.StepDefinition
}

// NewCapExInvestmentSaga creates a new CapEx Proposal & Investment Approval saga handler
func NewCapExInvestmentSaga() saga.SagaHandler {
	return &CapExInvestmentSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Submit CapEx Proposal - Initial proposal submission
			{
				StepNumber:    1,
				ServiceName:   "capex-proposal",
				HandlerMethod: "SubmitCapExProposal",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"title":         "$.input.title",
					"department":    "$.input.department",
					"description":   "$.input.description",
					"proposedAmount": "$.input.proposed_amount",
					"capexCategory": "$.input.capex_category",
					"justification": "$.input.justification",
					"submittedBy":   "$.input.submitted_by",
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
				CompensationSteps: []int32{110},
			},
			// Step 2: Technical Evaluation - Review technical feasibility
			{
				StepNumber:    2,
				ServiceName:   "capex-proposal",
				HandlerMethod: "TechnicalEvaluation",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"proposalID":     "$.steps.1.result.proposal_id",
					"capexCategory":  "$.input.capex_category",
					"specifications": "$.input.specifications",
					"technicalTeam": "$.input.technical_team",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        2,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{111},
			},
			// Step 3: Financial Evaluation - Calculate ROI, IRR, NPV
			{
				StepNumber:    3,
				ServiceName:   "budget",
				HandlerMethod: "FinancialEvaluation",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"proposalID":      "$.steps.1.result.proposal_id",
					"investmentAmount": "$.input.proposed_amount",
					"projectedBenefits": "$.input.projected_benefits",
					"projectLife":     "$.input.project_life_years",
					"discountRate":    "$.input.discount_rate",
					"salvageValue":    "$.input.salvage_value",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{112},
			},
			// Step 4: Cost Estimation - Detailed cost breakdown and contingency
			{
				StepNumber:    4,
				ServiceName:   "capex-proposal",
				HandlerMethod: "CostEstimation",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"proposalID":       "$.steps.1.result.proposal_id",
					"baseAmount":       "$.input.proposed_amount",
					"contingencyRate":  "$.input.contingency_rate",
					"installationCost": "$.input.installation_cost",
					"freightCost":      "$.input.freight_cost",
					"licensesCost":     "$.input.licenses_cost",
				},
				TimeoutSeconds: 30,
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
			// Step 5: Reserve Budget from CapEx Pool - Allocate budget - CRITICAL
			{
				StepNumber:    5,
				ServiceName:   "budget",
				HandlerMethod: "ReserveCapExBudget",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"proposalID":     "$.steps.1.result.proposal_id",
					"estimatedCost":  "$.steps.4.result.total_estimated_cost",
					"capexCategory":  "$.input.capex_category",
					"reserveYear":    "$.input.capex_year",
					"allocationName": "$.steps.1.result.proposal_id",
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
				CompensationSteps: []int32{114},
			},
			// Step 6: Submit Department Approval - Department level approval
			{
				StepNumber:    6,
				ServiceName:   "approval-workflow",
				HandlerMethod: "SubmitCapExDepartmentApproval",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"proposalID":      "$.steps.1.result.proposal_id",
					"department":      "$.input.department",
					"approverRole":    "DEPT_HEAD",
					"estimatedCost":   "$.steps.4.result.total_estimated_cost",
					"businessCase":    "$.steps.3.result.financial_summary",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        2,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{115},
			},
			// Step 7: Submit Finance Approval - Finance director approval
			{
				StepNumber:    7,
				ServiceName:   "approval-workflow",
				HandlerMethod: "SubmitCapExFinanceApproval",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"proposalID":      "$.steps.1.result.proposal_id",
					"approverRole":    "FINANCE_DIRECTOR",
					"estimatedCost":   "$.steps.4.result.total_estimated_cost",
					"roi":             "$.steps.3.result.roi",
					"irr":             "$.steps.3.result.irr",
					"npv":             "$.steps.3.result.npv",
					"deptApproved":    "$.steps.6.result.approved",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        2,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{116},
			},
			// Step 8: Board Approval if Above Threshold (>5M) - Board level approval - CRITICAL
			{
				StepNumber:    8,
				ServiceName:   "approval-workflow",
				HandlerMethod: "SubmitCapExBoardApproval",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"proposalID":         "$.steps.1.result.proposal_id",
					"estimatedCost":      "$.steps.4.result.total_estimated_cost",
					"approvalThreshold":  "$.input.board_approval_threshold",
					"requiresBoardApproval": "$.steps.4.result.requires_board_approval",
					"financeApproved":    "$.steps.7.result.approved",
				},
				TimeoutSeconds: 40,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        2,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      60000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{117},
			},
			// Step 9: Create Procurement Work Order - Generate PO for equipment/services
			{
				StepNumber:    9,
				ServiceName:   "procurement",
				HandlerMethod: "CreateCapExWorkOrder",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"proposalID":       "$.steps.1.result.proposal_id",
					"estimatedCost":    "$.steps.4.result.total_estimated_cost",
					"lineItems":        "$.input.line_items",
					"vendorList":       "$.input.preferred_vendors",
					"deliverySchedule": "$.input.delivery_schedule",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{118},
			},
			// Step 10: Create Asset Record - Create asset master for tracking
			{
				StepNumber:    10,
				ServiceName:   "asset",
				HandlerMethod: "CreateCapExAssetRecord",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"proposalID":        "$.steps.1.result.proposal_id",
					"assetDescription":  "$.input.title",
					"assetCategory":     "$.input.capex_category",
					"location":          "$.input.asset_location",
					"estimatedCost":     "$.steps.4.result.total_estimated_cost",
					"usefulLifeYears":   "$.input.useful_life_years",
					"depreciationMethod": "$.input.depreciation_method",
					"workOrderID":       "$.steps.9.result.work_order_id",
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
				CompensationSteps: []int32{119},
			},
			// Step 11: GL Posting for Capitalization - Post asset and corresponding GL entries - CRITICAL
			{
				StepNumber:    11,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostCapExCapitalization",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"proposalID":        "$.steps.1.result.proposal_id",
					"assetID":           "$.steps.10.result.asset_id",
					"capitalizedAmount": "$.steps.4.result.total_estimated_cost",
					"fixedAssetAccount": "$.input.fixed_asset_account",
					"capitalReserveAccount": "$.input.capital_reserve_account",
					"budgetAllocation":  "$.steps.5.result.allocation_id",
					"postingDate":       "$.input.posting_date",
				},
				TimeoutSeconds: 35,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{120},
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *CapExInvestmentSaga) SagaType() string {
	return "SAGA-BU03"
}

// GetStepDefinitions returns all step definitions for this saga
func (s *CapExInvestmentSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *CapExInvestmentSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: title, department, proposed_amount, capex_category, projected_benefits
func (s *CapExInvestmentSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	// Extract the nested 'input' object
	innerInput, ok := inputMap["input"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing 'input' field in saga input")
	}

	// Validate title
	if innerInput["title"] == nil {
		return errors.New("missing required field: title")
	}
	title, ok := innerInput["title"].(string)
	if !ok || title == "" {
		return errors.New("title must be a non-empty string")
	}

	// Validate department
	if innerInput["department"] == nil {
		return errors.New("missing required field: department")
	}
	department, ok := innerInput["department"].(string)
	if !ok || department == "" {
		return errors.New("department must be a non-empty string")
	}

	// Validate proposed_amount
	if innerInput["proposed_amount"] == nil {
		return errors.New("missing required field: proposed_amount")
	}
	proposedAmount, ok := innerInput["proposed_amount"].(float64)
	if !ok || proposedAmount <= 0 {
		return errors.New("proposed_amount must be a positive number")
	}

	// Validate capex_category
	if innerInput["capex_category"] == nil {
		return errors.New("missing required field: capex_category")
	}
	capexCategory, ok := innerInput["capex_category"].(string)
	if !ok || capexCategory == "" {
		return errors.New("capex_category must be a non-empty string")
	}

	// Validate projected_benefits
	if innerInput["projected_benefits"] == nil {
		return errors.New("missing required field: projected_benefits")
	}
	projectedBenefits, ok := innerInput["projected_benefits"].(float64)
	if !ok || projectedBenefits < 0 {
		return errors.New("projected_benefits must be a non-negative number")
	}

	// Validate project_life_years
	if innerInput["project_life_years"] != nil {
		projectLife, ok := innerInput["project_life_years"].(float64)
		if !ok || projectLife <= 0 {
			return errors.New("project_life_years must be a positive number")
		}
	}

	// Validate discount_rate (optional, should be 0-100 if present)
	if innerInput["discount_rate"] != nil {
		discountRate, ok := innerInput["discount_rate"].(float64)
		if !ok || discountRate < 0 || discountRate > 100 {
			return errors.New("discount_rate must be between 0 and 100")
		}
	}

	// Validate line_items
	if innerInput["line_items"] == nil {
		return errors.New("missing required field: line_items")
	}
	lineItems, ok := innerInput["line_items"].([]interface{})
	if !ok || len(lineItems) == 0 {
		return errors.New("line_items must be a non-empty array")
	}

	// Validate company_id (from context)
	if inputMap["companyID"] == nil {
		return errors.New("missing companyID in saga context")
	}

	return nil
}
