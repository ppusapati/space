// Package budget provides saga handlers for budget management workflows
package budget

import (
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// ============================================================================
// SAGA-BU01: Budget Approval & Control Tests (14 tests)
// ============================================================================

func TestBudgetApprovalControlSagaType(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	if sagaHandler.SagaType() != "SAGA-BU01" {
		t.Errorf("expected SAGA-BU01, got %s", sagaHandler.SagaType())
	}
}

func TestBudgetApprovalControlGetStepDefinitions(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	steps := sagaHandler.GetStepDefinitions()
	if len(steps) != 9 {
		t.Errorf("expected 9 steps, got %d", len(steps))
	}
}

func TestBudgetApprovalControlGetStepDefinition(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	step := sagaHandler.GetStepDefinition(1)
	if step == nil {
		t.Error("expected step 1 to exist")
	}
	if step.StepNumber != 1 {
		t.Errorf("expected step number 1, got %d", step.StepNumber)
	}
	if step.ServiceName != "budget" {
		t.Errorf("expected service budget, got %s", step.ServiceName)
	}
}

func TestBudgetApprovalControlCriticalSteps(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	criticalSteps := []int{3, 4, 6, 8}
	for _, stepNum := range criticalSteps {
		step := sagaHandler.GetStepDefinition(stepNum)
		if !step.IsCritical {
			t.Errorf("step %d should be critical", stepNum)
		}
	}
}

func TestBudgetApprovalControlCompensationSteps(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	step1 := sagaHandler.GetStepDefinition(1)
	if len(step1.CompensationSteps) != 1 || step1.CompensationSteps[0] != 110 {
		t.Error("step 1 should have compensation step 110")
	}
	step6 := sagaHandler.GetStepDefinition(6)
	if len(step6.CompensationSteps) != 1 || step6.CompensationSteps[0] != 114 {
		t.Error("step 6 should have compensation step 114")
	}
}

func TestBudgetApprovalControlValidateInputMissingDepartment(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"budget_year":   2026.0,
			"total_amount":  500000.0,
			"currency":      "USD",
			"cost_centers":  []interface{}{"CC001", "CC002"},
			"allocations":   []interface{}{250000.0, 250000.0},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: department" {
		t.Errorf("expected 'missing required field: department', got %v", err)
	}
}

func TestBudgetApprovalControlValidateInputMissingBudgetYear(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"department":   "Operations",
			"total_amount": 500000.0,
			"currency":     "USD",
			"cost_centers": []interface{}{"CC001", "CC002"},
			"allocations":  []interface{}{250000.0, 250000.0},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: budget_year" {
		t.Errorf("expected 'missing required field: budget_year', got %v", err)
	}
}

func TestBudgetApprovalControlValidateInputMissingTotalAmount(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"department":   "Operations",
			"budget_year":  2026.0,
			"currency":     "USD",
			"cost_centers": []interface{}{"CC001", "CC002"},
			"allocations":  []interface{}{250000.0, 250000.0},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: total_amount" {
		t.Errorf("expected 'missing required field: total_amount', got %v", err)
	}
}

func TestBudgetApprovalControlValidateInputMissingCurrency(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"department":   "Operations",
			"budget_year":  2026.0,
			"total_amount": 500000.0,
			"cost_centers": []interface{}{"CC001", "CC002"},
			"allocations":  []interface{}{250000.0, 250000.0},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: currency" {
		t.Errorf("expected 'missing required field: currency', got %v", err)
	}
}

func TestBudgetApprovalControlValidateInputMissingCostCenters(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"department":   "Operations",
			"budget_year":  2026.0,
			"total_amount": 500000.0,
			"currency":     "USD",
			"allocations":  []interface{}{250000.0, 250000.0},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: cost_centers" {
		t.Errorf("expected 'missing required field: cost_centers', got %v", err)
	}
}

func TestBudgetApprovalControlValidateInputMismatchedAllocations(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"department":   "Operations",
			"budget_year":  2026.0,
			"total_amount": 500000.0,
			"currency":     "USD",
			"cost_centers": []interface{}{"CC001", "CC002"},
			"allocations":  []interface{}{250000.0},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "allocations must have same length as cost_centers" {
		t.Errorf("expected 'allocations must have same length as cost_centers', got %v", err)
	}
}

func TestBudgetApprovalControlValidateInputNegativeAmount(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"department":   "Operations",
			"budget_year":  2026.0,
			"total_amount": -500000.0,
			"currency":     "USD",
			"cost_centers": []interface{}{"CC001", "CC002"},
			"allocations":  []interface{}{250000.0, 250000.0},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "total_amount must be a positive number" {
		t.Errorf("expected 'total_amount must be a positive number', got %v", err)
	}
}

func TestBudgetApprovalControlValidateInputMissingCompanyID(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"department":   "Operations",
			"budget_year":  2026.0,
			"total_amount": 500000.0,
			"currency":     "USD",
			"cost_centers": []interface{}{"CC001", "CC002"},
			"allocations":  []interface{}{250000.0, 250000.0},
		},
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "missing companyID in saga context" {
		t.Errorf("expected 'missing companyID in saga context', got %v", err)
	}
}

func TestBudgetApprovalControlValidateInputValid(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"department":   "Operations",
			"budget_year":  2026.0,
			"total_amount": 500000.0,
			"currency":     "USD",
			"cost_centers": []interface{}{"CC001", "CC002"},
			"allocations":  []interface{}{250000.0, 250000.0},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for valid input, got %v", err)
	}
}

func TestBudgetApprovalControlTimeouts(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	step1 := sagaHandler.GetStepDefinition(1)
	if step1.TimeoutSeconds != 15 {
		t.Errorf("step 1 expected timeout 15s, got %d", step1.TimeoutSeconds)
	}
	step8 := sagaHandler.GetStepDefinition(8)
	if step8.TimeoutSeconds != 20 {
		t.Errorf("step 8 expected timeout 20s, got %d", step8.TimeoutSeconds)
	}
}

func TestBudgetApprovalControlRetryConfig(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	step := sagaHandler.GetStepDefinition(1)
	if step.RetryConfig == nil {
		t.Error("expected retry configuration")
	}
	if step.RetryConfig.MaxRetries != 3 {
		t.Errorf("expected 3 retries, got %d", step.RetryConfig.MaxRetries)
	}
	if step.RetryConfig.BackoffMultiplier != 2.0 {
		t.Errorf("expected 2.0 multiplier, got %f", step.RetryConfig.BackoffMultiplier)
	}
}

// ============================================================================
// SAGA-BU02: Variance Analysis & Budget Review Tests (12 tests)
// ============================================================================

func TestVarianceAnalysisSagaType(t *testing.T) {
	sagaHandler := NewVarianceAnalysisSaga()
	if sagaHandler.SagaType() != "SAGA-BU02" {
		t.Errorf("expected SAGA-BU02, got %s", sagaHandler.SagaType())
	}
}

func TestVarianceAnalysisGetStepDefinitions(t *testing.T) {
	sagaHandler := NewVarianceAnalysisSaga()
	steps := sagaHandler.GetStepDefinitions()
	if len(steps) != 8 {
		t.Errorf("expected 8 steps, got %d", len(steps))
	}
}

func TestVarianceAnalysisGetStepDefinition(t *testing.T) {
	sagaHandler := NewVarianceAnalysisSaga()
	step := sagaHandler.GetStepDefinition(1)
	if step == nil {
		t.Error("expected step 1 to exist")
	}
	if step.StepNumber != 1 {
		t.Errorf("expected step number 1, got %d", step.StepNumber)
	}
	if step.ServiceName != "general-ledger" {
		t.Errorf("expected service general-ledger, got %s", step.ServiceName)
	}
}

func TestVarianceAnalysisCriticalSteps(t *testing.T) {
	sagaHandler := NewVarianceAnalysisSaga()
	criticalSteps := []int{3, 5, 7}
	for _, stepNum := range criticalSteps {
		step := sagaHandler.GetStepDefinition(stepNum)
		if !step.IsCritical {
			t.Errorf("step %d should be critical", stepNum)
		}
	}
}

func TestVarianceAnalysisCompensationSteps(t *testing.T) {
	sagaHandler := NewVarianceAnalysisSaga()
	step3 := sagaHandler.GetStepDefinition(3)
	if len(step3.CompensationSteps) != 1 || step3.CompensationSteps[0] != 110 {
		t.Error("step 3 should have compensation step 110")
	}
	step7 := sagaHandler.GetStepDefinition(7)
	if len(step7.CompensationSteps) != 1 || step7.CompensationSteps[0] != 113 {
		t.Error("step 7 should have compensation step 113")
	}
}

func TestVarianceAnalysisValidateInputMissingBudgetID(t *testing.T) {
	sagaHandler := NewVarianceAnalysisSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"start_date":          "2026-01-01",
			"end_date":            "2026-01-31",
			"cost_centers":        []interface{}{"CC001", "CC002"},
			"threshold_percentage": 10.0,
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: budget_id" {
		t.Errorf("expected 'missing required field: budget_id', got %v", err)
	}
}

func TestVarianceAnalysisValidateInputMissingStartDate(t *testing.T) {
	sagaHandler := NewVarianceAnalysisSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"budget_id":            "BUD001",
			"end_date":             "2026-01-31",
			"cost_centers":         []interface{}{"CC001", "CC002"},
			"threshold_percentage": 10.0,
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: start_date" {
		t.Errorf("expected 'missing required field: start_date', got %v", err)
	}
}

func TestVarianceAnalysisValidateInputMissingEndDate(t *testing.T) {
	sagaHandler := NewVarianceAnalysisSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"budget_id":            "BUD001",
			"start_date":           "2026-01-01",
			"cost_centers":         []interface{}{"CC001", "CC002"},
			"threshold_percentage": 10.0,
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: end_date" {
		t.Errorf("expected 'missing required field: end_date', got %v", err)
	}
}

func TestVarianceAnalysisValidateInputMissingCostCenters(t *testing.T) {
	sagaHandler := NewVarianceAnalysisSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"budget_id":            "BUD001",
			"start_date":           "2026-01-01",
			"end_date":             "2026-01-31",
			"threshold_percentage": 10.0,
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: cost_centers" {
		t.Errorf("expected 'missing required field: cost_centers', got %v", err)
	}
}

func TestVarianceAnalysisValidateInputInvalidThresholdPercentage(t *testing.T) {
	sagaHandler := NewVarianceAnalysisSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"budget_id":            "BUD001",
			"start_date":           "2026-01-01",
			"end_date":             "2026-01-31",
			"cost_centers":         []interface{}{"CC001", "CC002"},
			"threshold_percentage": 150.0,
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "threshold_percentage must be a number between 0 and 100" {
		t.Errorf("expected 'threshold_percentage must be a number between 0 and 100', got %v", err)
	}
}

func TestVarianceAnalysisValidateInputValid(t *testing.T) {
	sagaHandler := NewVarianceAnalysisSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"budget_id":            "BUD001",
			"start_date":           "2026-01-01",
			"end_date":             "2026-01-31",
			"cost_centers":         []interface{}{"CC001", "CC002"},
			"threshold_percentage": 10.0,
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for valid input, got %v", err)
	}
}

func TestVarianceAnalysisTimeouts(t *testing.T) {
	sagaHandler := NewVarianceAnalysisSaga()
	step1 := sagaHandler.GetStepDefinition(1)
	if step1.TimeoutSeconds != 20 {
		t.Errorf("step 1 expected timeout 20s, got %d", step1.TimeoutSeconds)
	}
}

func TestVarianceAnalysisRetryConfig(t *testing.T) {
	sagaHandler := NewVarianceAnalysisSaga()
	step := sagaHandler.GetStepDefinition(1)
	if step.RetryConfig == nil {
		t.Error("expected retry configuration")
	}
	if step.RetryConfig.MaxRetries != 3 {
		t.Errorf("expected 3 retries, got %d", step.RetryConfig.MaxRetries)
	}
}

// ============================================================================
// SAGA-BU03: CapEx Proposal & Investment Approval Tests (15 tests)
// ============================================================================

func TestCapExInvestmentSagaType(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	if sagaHandler.SagaType() != "SAGA-BU03" {
		t.Errorf("expected SAGA-BU03, got %s", sagaHandler.SagaType())
	}
}

func TestCapExInvestmentGetStepDefinitions(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	steps := sagaHandler.GetStepDefinitions()
	if len(steps) != 11 {
		t.Errorf("expected 11 steps, got %d", len(steps))
	}
}

func TestCapExInvestmentGetStepDefinition(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	step := sagaHandler.GetStepDefinition(1)
	if step == nil {
		t.Error("expected step 1 to exist")
	}
	if step.StepNumber != 1 {
		t.Errorf("expected step number 1, got %d", step.StepNumber)
	}
	if step.ServiceName != "capex-proposal" {
		t.Errorf("expected service capex-proposal, got %s", step.ServiceName)
	}
}

func TestCapExInvestmentCriticalSteps(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	criticalSteps := []int{5, 8, 11}
	for _, stepNum := range criticalSteps {
		step := sagaHandler.GetStepDefinition(stepNum)
		if !step.IsCritical {
			t.Errorf("step %d should be critical", stepNum)
		}
	}
}

func TestCapExInvestmentCompensationSteps(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	step1 := sagaHandler.GetStepDefinition(1)
	if len(step1.CompensationSteps) != 1 || step1.CompensationSteps[0] != 110 {
		t.Error("step 1 should have compensation step 110")
	}
	step11 := sagaHandler.GetStepDefinition(11)
	if len(step11.CompensationSteps) != 1 || step11.CompensationSteps[0] != 120 {
		t.Error("step 11 should have compensation step 120")
	}
}

func TestCapExInvestmentValidateInputMissingTitle(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"department":         "Operations",
			"proposed_amount":    1000000.0,
			"capex_category":     "Equipment",
			"projected_benefits": 200000.0,
			"line_items":         []interface{}{"Item1", "Item2"},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: title" {
		t.Errorf("expected 'missing required field: title', got %v", err)
	}
}

func TestCapExInvestmentValidateInputMissingDepartment(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"title":               "New Equipment Purchase",
			"proposed_amount":     1000000.0,
			"capex_category":      "Equipment",
			"projected_benefits":  200000.0,
			"line_items":          []interface{}{"Item1", "Item2"},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: department" {
		t.Errorf("expected 'missing required field: department', got %v", err)
	}
}

func TestCapExInvestmentValidateInputMissingProposedAmount(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"title":               "New Equipment Purchase",
			"department":          "Operations",
			"capex_category":      "Equipment",
			"projected_benefits":  200000.0,
			"line_items":          []interface{}{"Item1", "Item2"},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: proposed_amount" {
		t.Errorf("expected 'missing required field: proposed_amount', got %v", err)
	}
}

func TestCapExInvestmentValidateInputMissingCapexCategory(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"title":               "New Equipment Purchase",
			"department":          "Operations",
			"proposed_amount":     1000000.0,
			"projected_benefits":  200000.0,
			"line_items":          []interface{}{"Item1", "Item2"},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: capex_category" {
		t.Errorf("expected 'missing required field: capex_category', got %v", err)
	}
}

func TestCapExInvestmentValidateInputMissingProjectedBenefits(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"title":              "New Equipment Purchase",
			"department":         "Operations",
			"proposed_amount":    1000000.0,
			"capex_category":     "Equipment",
			"line_items":         []interface{}{"Item1", "Item2"},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: projected_benefits" {
		t.Errorf("expected 'missing required field: projected_benefits', got %v", err)
	}
}

func TestCapExInvestmentValidateInputMissingLineItems(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"title":               "New Equipment Purchase",
			"department":          "Operations",
			"proposed_amount":     1000000.0,
			"capex_category":      "Equipment",
			"projected_benefits":  200000.0,
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: line_items" {
		t.Errorf("expected 'missing required field: line_items', got %v", err)
	}
}

func TestCapExInvestmentValidateInputNegativeProposedAmount(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"title":               "New Equipment Purchase",
			"department":          "Operations",
			"proposed_amount":     -1000000.0,
			"capex_category":      "Equipment",
			"projected_benefits":  200000.0,
			"line_items":          []interface{}{"Item1", "Item2"},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "proposed_amount must be a positive number" {
		t.Errorf("expected 'proposed_amount must be a positive number', got %v", err)
	}
}

func TestCapExInvestmentValidateInputNegativeProjectedBenefits(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"title":               "New Equipment Purchase",
			"department":          "Operations",
			"proposed_amount":     1000000.0,
			"capex_category":      "Equipment",
			"projected_benefits":  -200000.0,
			"line_items":          []interface{}{"Item1", "Item2"},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "projected_benefits must be a non-negative number" {
		t.Errorf("expected 'projected_benefits must be a non-negative number', got %v", err)
	}
}

func TestCapExInvestmentValidateInputInvalidDiscountRate(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"title":               "New Equipment Purchase",
			"department":          "Operations",
			"proposed_amount":     1000000.0,
			"capex_category":      "Equipment",
			"projected_benefits":  200000.0,
			"discount_rate":       150.0,
			"line_items":          []interface{}{"Item1", "Item2"},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err == nil || err.Error() != "discount_rate must be between 0 and 100" {
		t.Errorf("expected 'discount_rate must be between 0 and 100', got %v", err)
	}
}

func TestCapExInvestmentValidateInputValid(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"title":               "New Equipment Purchase",
			"department":          "Operations",
			"proposed_amount":     1000000.0,
			"capex_category":      "Equipment",
			"projected_benefits":  200000.0,
			"discount_rate":       10.0,
			"project_life_years":  5.0,
			"line_items":          []interface{}{"Item1", "Item2"},
		},
		"companyID": "COMP001",
	}
	err := sagaHandler.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for valid input, got %v", err)
	}
}

func TestCapExInvestmentTimeouts(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	step1 := sagaHandler.GetStepDefinition(1)
	if step1.TimeoutSeconds != 20 {
		t.Errorf("step 1 expected timeout 20s, got %d", step1.TimeoutSeconds)
	}
	step11 := sagaHandler.GetStepDefinition(11)
	if step11.TimeoutSeconds != 35 {
		t.Errorf("step 11 expected timeout 35s, got %d", step11.TimeoutSeconds)
	}
}

func TestCapExInvestmentRetryConfig(t *testing.T) {
	sagaHandler := NewCapExInvestmentSaga()
	step := sagaHandler.GetStepDefinition(1)
	if step.RetryConfig == nil {
		t.Error("expected retry configuration")
	}
	if step.RetryConfig.MaxRetries != 3 {
		t.Errorf("expected 3 retries, got %d", step.RetryConfig.MaxRetries)
	}
	step8 := sagaHandler.GetStepDefinition(8)
	if step8.RetryConfig.MaxRetries != 2 {
		t.Errorf("step 8 expected 2 retries, got %d", step8.RetryConfig.MaxRetries)
	}
}

// ============================================================================
// Cross-Saga Integration Tests (6 tests)
// ============================================================================

func TestBudgetSagasAreRegistered(t *testing.T) {
	handlers := []saga.SagaHandler{
		NewBudgetApprovalControlSaga(),
		NewVarianceAnalysisSaga(),
		NewCapExInvestmentSaga(),
	}
	sagaTypes := map[string]bool{
		"SAGA-BU01": false,
		"SAGA-BU02": false,
		"SAGA-BU03": false,
	}
	for _, handler := range handlers {
		if _, exists := sagaTypes[handler.SagaType()]; !exists {
			t.Errorf("unexpected saga type: %s", handler.SagaType())
		}
		sagaTypes[handler.SagaType()] = true
	}
	for sagaType, found := range sagaTypes {
		if !found {
			t.Errorf("saga type %s not found", sagaType)
		}
	}
}

func TestAllBudgetSagasHaveValidInputMapping(t *testing.T) {
	handlers := []saga.SagaHandler{
		NewBudgetApprovalControlSaga(),
		NewVarianceAnalysisSaga(),
		NewCapExInvestmentSaga(),
	}
	for _, handler := range handlers {
		steps := handler.GetStepDefinitions()
		for _, step := range steps {
			if step.InputMapping == nil {
				t.Errorf("%s step %d has no input mapping", handler.SagaType(), step.StepNumber)
			}
			if len(step.InputMapping) == 0 {
				t.Errorf("%s step %d has empty input mapping", handler.SagaType(), step.StepNumber)
			}
		}
	}
}

func TestAllBudgetSagasHaveValidRetryConfig(t *testing.T) {
	handlers := []saga.SagaHandler{
		NewBudgetApprovalControlSaga(),
		NewVarianceAnalysisSaga(),
		NewCapExInvestmentSaga(),
	}
	for _, handler := range handlers {
		steps := handler.GetStepDefinitions()
		for _, step := range steps {
			if step.RetryConfig == nil {
				t.Errorf("%s step %d has no retry configuration", handler.SagaType(), step.StepNumber)
			}
			if step.RetryConfig.MaxRetries == 0 {
				t.Errorf("%s step %d has zero retries", handler.SagaType(), step.StepNumber)
			}
			if step.RetryConfig.BackoffMultiplier <= 1.0 {
				t.Errorf("%s step %d has invalid backoff multiplier", handler.SagaType(), step.StepNumber)
			}
		}
	}
}

func TestBudgetSagaStepServiceNames(t *testing.T) {
	bu01 := NewBudgetApprovalControlSaga()
	expectedServices := map[int]string{
		1: "budget",
		2: "general-ledger",
		3: "approval-workflow",
		4: "approval-workflow",
		5: "cost-center",
		6: "general-ledger",
		7: "budget",
		8: "budget",
		9: "notification",
	}
	for stepNum, expectedService := range expectedServices {
		step := bu01.GetStepDefinition(stepNum)
		if step.ServiceName != expectedService {
			t.Errorf("BU01 step %d expected service %s, got %s", stepNum, expectedService, step.ServiceName)
		}
	}
}

func TestBudgetSagaNonExistentStep(t *testing.T) {
	sagaHandler := NewBudgetApprovalControlSaga()
	step := sagaHandler.GetStepDefinition(999)
	if step != nil {
		t.Error("expected nil for non-existent step")
	}
}

func TestBudgetSagasContextMapping(t *testing.T) {
	handlers := []saga.SagaHandler{
		NewBudgetApprovalControlSaga(),
		NewVarianceAnalysisSaga(),
		NewCapExInvestmentSaga(),
	}
	contextFields := []string{"tenantID", "companyID", "branchID"}
	for _, handler := range handlers {
		steps := handler.GetStepDefinitions()
		for _, step := range steps {
			for _, field := range contextFields {
				if _, exists := step.InputMapping[field]; !exists {
					t.Errorf("%s step %d missing context field %s", handler.SagaType(), step.StepNumber, field)
				}
			}
		}
	}
}
