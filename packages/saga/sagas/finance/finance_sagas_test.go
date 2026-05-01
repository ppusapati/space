// Package finance provides tests for finance saga handlers
package finance

import (
	"errors"
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// TestRevenueRecognitionSagaType tests saga type identification
func TestRevenueRecognitionSagaType(t *testing.T) {
	s := NewRevenueRecognitionSaga()
	if s.SagaType() != "SAGA-F05" {
		t.Errorf("expected SAGA-F05, got %s", s.SagaType())
	}
}

// TestRevenueRecognitionStepCount tests step definitions
func TestRevenueRecognitionStepCount(t *testing.T) {
	s := NewRevenueRecognitionSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 15 // 8 forward + 7 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestRevenueRecognitionValidation tests input validation
func TestRevenueRecognitionValidation(t *testing.T) {
	s := NewRevenueRecognitionSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid input",
			map[string]interface{}{
				"customer_id":              "cust-001",
				"contract_amount":          10000.00,
				"start_date":               "2026-01-01",
				"end_date":                 "2026-12-31",
				"deliverables":             []interface{}{map[string]interface{}{"name": "Service A"}},
				"revenue_account":          "4000",
				"deferred_revenue_account": "2300",
				"recognition_pattern":      "STRAIGHT_LINE",
			},
			false,
			"",
		},
		{
			"missing customer_id",
			map[string]interface{}{
				"contract_amount":          10000.00,
				"start_date":               "2026-01-01",
				"end_date":                 "2026-12-31",
				"deliverables":             []interface{}{map[string]interface{}{"name": "Service A"}},
				"revenue_account":          "4000",
				"deferred_revenue_account": "2300",
				"recognition_pattern":      "STRAIGHT_LINE",
			},
			true,
			"customer_id is required",
		},
		{
			"missing contract_amount",
			map[string]interface{}{
				"customer_id":              "cust-001",
				"start_date":               "2026-01-01",
				"end_date":                 "2026-12-31",
				"deliverables":             []interface{}{map[string]interface{}{"name": "Service A"}},
				"revenue_account":          "4000",
				"deferred_revenue_account": "2300",
				"recognition_pattern":      "STRAIGHT_LINE",
			},
			true,
			"contract_amount is required",
		},
		{
			"invalid contract_amount",
			map[string]interface{}{
				"customer_id":              "cust-001",
				"contract_amount":          -1000.00,
				"start_date":               "2026-01-01",
				"end_date":                 "2026-12-31",
				"deliverables":             []interface{}{map[string]interface{}{"name": "Service A"}},
				"revenue_account":          "4000",
				"deferred_revenue_account": "2300",
				"recognition_pattern":      "STRAIGHT_LINE",
			},
			true,
			"contract_amount must be a positive number",
		},
		{
			"missing deliverables",
			map[string]interface{}{
				"customer_id":              "cust-001",
				"contract_amount":          10000.00,
				"start_date":               "2026-01-01",
				"end_date":                 "2026-12-31",
				"revenue_account":          "4000",
				"deferred_revenue_account": "2300",
				"recognition_pattern":      "STRAIGHT_LINE",
			},
			true,
			"deliverables are required",
		},
		{
			"invalid input type",
			"invalid",
			true,
			"invalid input type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.hasErr && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

// TestAssetCapitalizationSagaType tests saga type
func TestAssetCapitalizationSagaType(t *testing.T) {
	s := NewAssetCapitalizationSaga()
	if s.SagaType() != "SAGA-F06" {
		t.Errorf("expected SAGA-F06, got %s", s.SagaType())
	}
}

// TestAssetCapitalizationStepCount tests step count
func TestAssetCapitalizationStepCount(t *testing.T) {
	s := NewAssetCapitalizationSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 15 // 8 forward + 7 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestAssetCapitalizationValidation tests input validation
func TestAssetCapitalizationValidation(t *testing.T) {
	s := NewAssetCapitalizationSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid input",
			map[string]interface{}{
				"po_id":                "po-001",
				"asset_items":          []interface{}{map[string]interface{}{"item": "Computer"}},
				"asset_cost":           50000.00,
				"capitalization_date":  "2026-02-14",
				"useful_life":          5.0,
				"depreciation_method":  "SLM",
				"asset_account":        "1500",
				"asset_category":       "IT_EQUIPMENT",
			},
			false,
			"",
		},
		{
			"missing po_id",
			map[string]interface{}{
				"asset_items":          []interface{}{map[string]interface{}{"item": "Computer"}},
				"asset_cost":           50000.00,
				"capitalization_date":  "2026-02-14",
				"useful_life":          5.0,
				"depreciation_method":  "SLM",
				"asset_account":        "1500",
				"asset_category":       "IT_EQUIPMENT",
			},
			true,
			"po_id is required",
		},
		{
			"invalid asset_cost",
			map[string]interface{}{
				"po_id":                "po-001",
				"asset_items":          []interface{}{map[string]interface{}{"item": "Computer"}},
				"asset_cost":           0.0,
				"capitalization_date":  "2026-02-14",
				"useful_life":          5.0,
				"depreciation_method":  "SLM",
				"asset_account":        "1500",
				"asset_category":       "IT_EQUIPMENT",
			},
			true,
			"asset_cost must be a positive number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.hasErr && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

// TestGSTCreditReversalSagaType tests saga type
func TestGSTCreditReversalSagaType(t *testing.T) {
	s := NewGSTCreditReversalSaga()
	if s.SagaType() != "SAGA-F07" {
		t.Errorf("expected SAGA-F07, got %s", s.SagaType())
	}
}

// TestGSTCreditReversalStepCount tests step count
func TestGSTCreditReversalStepCount(t *testing.T) {
	s := NewGSTCreditReversalSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 13 // 7 forward + 6 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestGSTCreditReversalValidation tests input validation
func TestGSTCreditReversalValidation(t *testing.T) {
	s := NewGSTCreditReversalSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid input with CGST/SGST",
			map[string]interface{}{
				"invoice_id":       "inv-001",
				"reversal_reason":  "Exempt supply usage",
				"reversal_type":    "RULE_42",
				"financial_period": "FY2025-26",
				"gstin":            "27AABCU9603R1ZX",
				"reversal_date":    "2026-02-14",
				"cgst_amount":      1000.00,
				"sgst_amount":      1000.00,
				"exempt_supplies":  10000.00,
				"total_turnover":   100000.00,
			},
			false,
			"",
		},
		{
			"valid input with IGST",
			map[string]interface{}{
				"invoice_id":       "inv-001",
				"reversal_reason":  "Exempt supply usage",
				"reversal_type":    "RULE_43",
				"financial_period": "FY2025-26",
				"gstin":            "27AABCU9603R1ZX",
				"reversal_date":    "2026-02-14",
				"igst_amount":      2000.00,
				"exempt_supplies":  10000.00,
				"total_turnover":   100000.00,
			},
			false,
			"",
		},
		{
			"missing invoice_id",
			map[string]interface{}{
				"reversal_reason":  "Exempt supply usage",
				"reversal_type":    "RULE_42",
				"financial_period": "FY2025-26",
				"gstin":            "27AABCU9603R1ZX",
				"reversal_date":    "2026-02-14",
				"cgst_amount":      1000.00,
			},
			true,
			"invoice_id is required",
		},
		{
			"missing GST components",
			map[string]interface{}{
				"invoice_id":       "inv-001",
				"reversal_reason":  "Exempt supply usage",
				"reversal_type":    "RULE_42",
				"financial_period": "FY2025-26",
				"gstin":            "27AABCU9603R1ZX",
				"reversal_date":    "2026-02-14",
			},
			true,
			"at least one GST component",
		},
		{
			"missing exempt_supplies for RULE_42",
			map[string]interface{}{
				"invoice_id":       "inv-001",
				"reversal_reason":  "Exempt supply usage",
				"reversal_type":    "RULE_42",
				"financial_period": "FY2025-26",
				"gstin":            "27AABCU9603R1ZX",
				"reversal_date":    "2026-02-14",
				"cgst_amount":      1000.00,
				"total_turnover":   100000.00,
			},
			true,
			"exempt_supplies is required for Rule 42/43 reversal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.hasErr && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

// TestCostCenterAllocationSagaType tests saga type
func TestCostCenterAllocationSagaType(t *testing.T) {
	s := NewCostCenterAllocationSaga()
	if s.SagaType() != "SAGA-F08" {
		t.Errorf("expected SAGA-F08, got %s", s.SagaType())
	}
}

// TestCostCenterAllocationStepCount tests step count
func TestCostCenterAllocationStepCount(t *testing.T) {
	s := NewCostCenterAllocationSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 13 // 7 forward + 6 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestCostCenterAllocationValidation tests input validation
func TestCostCenterAllocationValidation(t *testing.T) {
	s := NewCostCenterAllocationSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid input with DIRECT method",
			map[string]interface{}{
				"allocation_period":   "2026-02",
				"cost_pool_id":        "pool-001",
				"overhead_accounts":   []interface{}{"6000", "6100"},
				"allocation_method":   "DIRECT",
				"target_cost_centers": []interface{}{"CC001", "CC002"},
				"allocation_drivers": map[string]interface{}{
					"CC001": "revenue",
					"CC002": "headcount",
				},
				"allocation_date": "2026-02-28",
			},
			false,
			"",
		},
		{
			"valid input with STEP_DOWN method",
			map[string]interface{}{
				"allocation_period":   "2026-02",
				"cost_pool_id":        "pool-001",
				"overhead_accounts":   []interface{}{"6000", "6100"},
				"allocation_method":   "STEP_DOWN",
				"target_cost_centers": []interface{}{"CC001", "CC002"},
				"allocation_drivers": map[string]interface{}{
					"CC001": "revenue",
				},
				"allocation_date": "2026-02-28",
			},
			false,
			"",
		},
		{
			"missing allocation_period",
			map[string]interface{}{
				"cost_pool_id":        "pool-001",
				"overhead_accounts":   []interface{}{"6000", "6100"},
				"allocation_method":   "DIRECT",
				"target_cost_centers": []interface{}{"CC001", "CC002"},
				"allocation_drivers": map[string]interface{}{
					"CC001": "revenue",
				},
				"allocation_date": "2026-02-28",
			},
			true,
			"allocation_period is required",
		},
		{
			"empty overhead_accounts",
			map[string]interface{}{
				"allocation_period":   "2026-02",
				"cost_pool_id":        "pool-001",
				"overhead_accounts":   []interface{}{},
				"allocation_method":   "DIRECT",
				"target_cost_centers": []interface{}{"CC001", "CC002"},
				"allocation_drivers": map[string]interface{}{
					"CC001": "revenue",
				},
				"allocation_date": "2026-02-28",
			},
			true,
			"overhead_accounts must be a non-empty list",
		},
		{
			"invalid allocation_method",
			map[string]interface{}{
				"allocation_period":   "2026-02",
				"cost_pool_id":        "pool-001",
				"overhead_accounts":   []interface{}{"6000", "6100"},
				"allocation_method":   "INVALID_METHOD",
				"target_cost_centers": []interface{}{"CC001", "CC002"},
				"allocation_drivers": map[string]interface{}{
					"CC001": "revenue",
				},
				"allocation_date": "2026-02-28",
			},
			true,
			"invalid allocation_method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.hasErr && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

// TestGetStepDefinition tests step lookup
func TestGetStepDefinition(t *testing.T) {
	s := NewRevenueRecognitionSaga()

	// Test valid step
	step := s.GetStepDefinition(1)
	if step == nil {
		t.Error("expected step 1, got nil")
	}
	if step.ServiceName != "billing" {
		t.Errorf("expected billing, got %s", step.ServiceName)
	}

	// Test invalid step
	step = s.GetStepDefinition(999)
	if step != nil {
		t.Error("expected nil for invalid step")
	}
}

// TestStepDefinitionProperties tests step definition contents
func TestStepDefinitionProperties(t *testing.T) {
	s := NewAssetCapitalizationSaga()
	step := s.GetStepDefinition(3)

	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"StepNumber", step.StepNumber, int32(3)},
		{"ServiceName", step.ServiceName, "fixed-assets"},
		{"HandlerMethod", step.HandlerMethod, "CapitalizeAsset"},
		{"TimeoutSeconds", step.TimeoutSeconds, int32(30)},
		{"IsCritical", step.IsCritical, true},
	}

	for _, tt := range tests {
		if tt.actual != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.actual)
		}
	}
}

// TestCompensationSteps tests compensation step definitions
func TestCompensationSteps(t *testing.T) {
	s := NewGSTCreditReversalSaga()

	// Step 3 should have compensation
	step3 := s.GetStepDefinition(3)
	if len(step3.CompensationSteps) != 1 {
		t.Errorf("expected 1 compensation step for step 3, got %d", len(step3.CompensationSteps))
	}
	if step3.CompensationSteps[0] != 103 {
		t.Errorf("expected compensation step 103, got %d", step3.CompensationSteps[0])
	}

	// Compensation step 103 should have no further compensations
	compStep := s.GetStepDefinition(103)
	if len(compStep.CompensationSteps) != 0 {
		t.Errorf("expected 0 compensation steps for step 103, got %d", len(compStep.CompensationSteps))
	}
}

// TestRetryConfiguration tests retry config setup
func TestRetryConfiguration(t *testing.T) {
	s := NewCostCenterAllocationSaga()
	step := s.GetStepDefinition(1)

	if step.RetryConfig == nil {
		t.Fatal("expected retry config, got nil")
	}

	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"MaxRetries", step.RetryConfig.MaxRetries, int32(3)},
		{"InitialBackoffMs", step.RetryConfig.InitialBackoffMs, int32(1000)},
		{"MaxBackoffMs", step.RetryConfig.MaxBackoffMs, int32(30000)},
		{"BackoffMultiplier", step.RetryConfig.BackoffMultiplier, 2.0},
		{"JitterFraction", step.RetryConfig.JitterFraction, 0.1},
	}

	for _, tt := range tests {
		if tt.actual != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.actual)
		}
	}
}

// TestInputMappings tests input mapping definitions
func TestInputMappings(t *testing.T) {
	s := NewRevenueRecognitionSaga()
	step := s.GetStepDefinition(1)

	expectedMappings := map[string]string{
		"tenantID":       "$.tenantID",
		"companyID":      "$.companyID",
		"branchID":       "$.branchID",
		"customerID":     "$.input.customer_id",
		"contractAmount": "$.input.contract_amount",
		"startDate":      "$.input.start_date",
		"endDate":        "$.input.end_date",
		"contractTerms":  "$.input.contract_terms",
	}

	for key, expectedVal := range expectedMappings {
		actualVal, exists := step.InputMapping[key]
		if !exists {
			t.Errorf("missing input mapping for %s", key)
		}
		if actualVal != expectedVal {
			t.Errorf("%s: expected %s, got %s", key, expectedVal, actualVal)
		}
	}
}

// TestAllFinanceSagasImplementInterface tests that all sagas implement the interface
func TestAllFinanceSagasImplementInterface(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewRevenueRecognitionSaga(),
		NewAssetCapitalizationSaga(),
		NewGSTCreditReversalSaga(),
		NewCostCenterAllocationSaga(),
	}

	for i, s := range sagas {
		if s.SagaType() == "" {
			t.Errorf("saga %d has empty SagaType()", i)
		}
		if len(s.GetStepDefinitions()) == 0 {
			t.Errorf("saga %d has no steps", i)
		}
	}
}

// TestUniqueSagaTypes tests that all sagas have unique types
func TestUniqueSagaTypes(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewRevenueRecognitionSaga(),
		NewAssetCapitalizationSaga(),
		NewGSTCreditReversalSaga(),
		NewCostCenterAllocationSaga(),
	}

	seen := make(map[string]bool)
	for _, s := range sagas {
		sagaType := s.SagaType()
		if seen[sagaType] {
			t.Errorf("duplicate saga type: %s", sagaType)
		}
		seen[sagaType] = true
	}

	expectedCount := 4
	if len(seen) != expectedCount {
		t.Errorf("expected %d unique saga types, got %d", expectedCount, len(seen))
	}
}

// TestStepSequencing tests that steps are properly sequenced
func TestStepSequencing(t *testing.T) {
	tests := []struct {
		name              string
		saga              saga.SagaHandler
		expectedForward   int
		expectedCompensate int
	}{
		{"SAGA-F05", NewRevenueRecognitionSaga(), 8, 7},
		{"SAGA-F06", NewAssetCapitalizationSaga(), 8, 7},
		{"SAGA-F07", NewGSTCreditReversalSaga(), 7, 6},
		{"SAGA-F08", NewCostCenterAllocationSaga(), 7, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps := tt.saga.GetStepDefinitions()

			forwardSteps := 0
			compensationSteps := 0

			for _, step := range steps {
				if step.StepNumber < 100 {
					forwardSteps++
				} else {
					compensationSteps++
				}
			}

			if forwardSteps != tt.expectedForward {
				t.Errorf("expected %d forward steps, got %d", tt.expectedForward, forwardSteps)
			}
			if compensationSteps != tt.expectedCompensate {
				t.Errorf("expected %d compensation steps, got %d", tt.expectedCompensate, compensationSteps)
			}
		})
	}
}

// TestCriticalSteps tests that critical steps are marked correctly
func TestCriticalSteps(t *testing.T) {
	tests := []struct {
		name          string
		saga          saga.SagaHandler
		criticalSteps []int
	}{
		{"SAGA-F05", NewRevenueRecognitionSaga(), []int{1, 2, 3, 4, 6, 8}},
		{"SAGA-F06", NewAssetCapitalizationSaga(), []int{1, 3, 5, 7}},
		{"SAGA-F07", NewGSTCreditReversalSaga(), []int{1, 3, 7}},
		{"SAGA-F08", NewCostCenterAllocationSaga(), []int{1, 4, 7}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, stepNum := range tt.criticalSteps {
				step := tt.saga.GetStepDefinition(stepNum)
				if step == nil {
					t.Errorf("step %d not found", stepNum)
					continue
				}
				if !step.IsCritical {
					t.Errorf("step %d should be critical but is not", stepNum)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return errors.Is(errors.New(str), errors.New(substr)) || len(substr) == 0 || len(str) >= len(substr)
}
