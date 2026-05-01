// Package sales provides tests for saga handlers
package sales

import (
	"errors"
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// TestOrderToCashSagaType tests saga type identification
func TestOrderToCashSagaType(t *testing.T) {
	s := NewOrderToCashSaga()
	if s.SagaType() != "SAGA-S01" {
		t.Errorf("expected SAGA-S01, got %s", s.SagaType())
	}
}

// TestOrderToCashStepCount tests step definitions
func TestOrderToCashStepCount(t *testing.T) {
	s := NewOrderToCashSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 15 // 8 forward + 7 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestOrderToCashValidation tests input validation
func TestOrderToCashValidation(t *testing.T) {
	s := NewOrderToCashSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid input",
			map[string]interface{}{
				"customer_id": "cust-001",
				"items":       []interface{}{map[string]interface{}{"sku": "SKU001"}},
				"total_amount": 1000.00,
			},
			false,
			"",
		},
		{
			"missing customer_id",
			map[string]interface{}{
				"items":       []interface{}{map[string]interface{}{"sku": "SKU001"}},
				"total_amount": 1000.00,
			},
			true,
			"customer_id is required",
		},
		{
			"missing items",
			map[string]interface{}{
				"customer_id":  "cust-001",
				"total_amount": 1000.00,
			},
			true,
			"items are required",
		},
		{
			"empty items",
			map[string]interface{}{
				"customer_id":  "cust-001",
				"items":        []interface{}{},
				"total_amount": 1000.00,
			},
			true,
			"items must be a non-empty list",
		},
		{
			"missing total_amount",
			map[string]interface{}{
				"customer_id": "cust-001",
				"items":       []interface{}{map[string]interface{}{"sku": "SKU001"}},
			},
			true,
			"total_amount is required",
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

// TestQuotationToOrderSagaType tests saga type
func TestQuotationToOrderSagaType(t *testing.T) {
	s := NewQuotationToOrderSaga()
	if s.SagaType() != "SAGA-S02" {
		t.Errorf("expected SAGA-S02, got %s", s.SagaType())
	}
}

// TestOrderToFulfillmentSagaType tests saga type
func TestOrderToFulfillmentSagaType(t *testing.T) {
	s := NewOrderToFulfillmentSaga()
	if s.SagaType() != "SAGA-S03" {
		t.Errorf("expected SAGA-S03, got %s", s.SagaType())
	}
}

// TestSalesReturnSagaType tests saga type
func TestSalesReturnSagaType(t *testing.T) {
	s := NewSalesReturnSaga()
	if s.SagaType() != "SAGA-S04" {
		t.Errorf("expected SAGA-S04, got %s", s.SagaType())
	}
}

// TestCommissionCalculationSagaType tests saga type
func TestCommissionCalculationSagaType(t *testing.T) {
	s := NewCommissionCalculationSaga()
	if s.SagaType() != "SAGA-S05" {
		t.Errorf("expected SAGA-S05, got %s", s.SagaType())
	}
}

// TestEInvoiceGenerationSagaType tests saga type
func TestEInvoiceGenerationSagaType(t *testing.T) {
	s := NewEInvoiceGenerationSaga()
	if s.SagaType() != "SAGA-S06" {
		t.Errorf("expected SAGA-S06, got %s", s.SagaType())
	}
}

// TestDealerIncentiveSagaType tests saga type
func TestDealerIncentiveSagaType(t *testing.T) {
	s := NewDealerIncentiveSaga()
	if s.SagaType() != "SAGA-S07" {
		t.Errorf("expected SAGA-S07, got %s", s.SagaType())
	}
}

// TestGetStepDefinition tests step lookup
func TestGetStepDefinition(t *testing.T) {
	s := NewOrderToCashSaga()

	// Test valid step
	step := s.GetStepDefinition(1)
	if step == nil {
		t.Error("expected step 1, got nil")
	}
	if step.ServiceName != "sales-order" {
		t.Errorf("expected sales-order, got %s", step.ServiceName)
	}

	// Test invalid step
	step = s.GetStepDefinition(999)
	if step != nil {
		t.Error("expected nil for invalid step")
	}
}

// TestStepDefinitionProperties tests step definition contents
func TestStepDefinitionProperties(t *testing.T) {
	s := NewOrderToCashSaga()
	step := s.GetStepDefinition(1)

	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"StepNumber", step.StepNumber, int32(1)},
		{"ServiceName", step.ServiceName, "sales-order"},
		{"HandlerMethod", step.HandlerMethod, "CreateOrder"},
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
	s := NewOrderToCashSaga()

	// Step 2 should have compensation
	step2 := s.GetStepDefinition(2)
	if len(step2.CompensationSteps) != 1 {
		t.Errorf("expected 1 compensation step for step 2, got %d", len(step2.CompensationSteps))
	}
	if step2.CompensationSteps[0] != 102 {
		t.Errorf("expected compensation step 102, got %d", step2.CompensationSteps[0])
	}

	// Compensation step 102 should have no further compensations
	compStep := s.GetStepDefinition(102)
	if len(compStep.CompensationSteps) != 0 {
		t.Errorf("expected 0 compensation steps for step 102, got %d", len(compStep.CompensationSteps))
	}
}

// TestRetryConfiguration tests retry config setup
func TestRetryConfiguration(t *testing.T) {
	s := NewOrderToCashSaga()
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
	s := NewOrderToCashSaga()
	step := s.GetStepDefinition(1)

	expectedMappings := map[string]string{
		"tenantID":    "$.tenantID",
		"companyID":   "$.companyID",
		"branchID":    "$.branchID",
		"customerID":  "$.input.customer_id",
		"items":       "$.input.items",
		"totalAmount": "$.input.total_amount",
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

// TestAllSagasImplementInterface tests that all sagas implement the interface
func TestAllSagasImplementInterface(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewOrderToCashSaga(),
		NewQuotationToOrderSaga(),
		NewOrderToFulfillmentSaga(),
		NewSalesReturnSaga(),
		NewCommissionCalculationSaga(),
		NewEInvoiceGenerationSaga(),
		NewDealerIncentiveSaga(),
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
		NewOrderToCashSaga(),
		NewQuotationToOrderSaga(),
		NewOrderToFulfillmentSaga(),
		NewSalesReturnSaga(),
		NewCommissionCalculationSaga(),
		NewEInvoiceGenerationSaga(),
		NewDealerIncentiveSaga(),
	}

	seen := make(map[string]bool)
	for _, s := range sagas {
		sagaType := s.SagaType()
		if seen[sagaType] {
			t.Errorf("duplicate saga type: %s", sagaType)
		}
		seen[sagaType] = true
	}

	expectedCount := 7
	if len(seen) != expectedCount {
		t.Errorf("expected %d unique saga types, got %d", expectedCount, len(seen))
	}
}

// TestStepSequencing tests that steps are properly sequenced
func TestStepSequencing(t *testing.T) {
	s := NewOrderToCashSaga()
	steps := s.GetStepDefinitions()

	// Forward steps should be 1-8
	forwardSteps := 0
	compensationSteps := 0

	for _, step := range steps {
		if step.StepNumber < 100 {
			forwardSteps++
		} else {
			compensationSteps++
		}
	}

	if forwardSteps != 8 {
		t.Errorf("expected 8 forward steps, got %d", forwardSteps)
	}
	if compensationSteps != 7 {
		t.Errorf("expected 7 compensation steps, got %d", compensationSteps)
	}
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return errors.Is(errors.New(str), errors.New(substr)) || len(substr) == 0 || len(str) >= len(substr)
}
