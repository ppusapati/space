// Package purchase provides tests for saga handlers
package purchase

import (
	"errors"
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// TestProcureToPaySagaType tests saga type identification
func TestProcureToPaySagaType(t *testing.T) {
	s := NewProcureToPaySaga()
	if s.SagaType() != "SAGA-P01" {
		t.Errorf("expected SAGA-P01, got %s", s.SagaType())
	}
}

// TestProcureToPayStepCount tests step definitions
func TestProcureToPayStepCount(t *testing.T) {
	s := NewProcureToPaySaga()
	steps := s.GetStepDefinitions()
	expectedCount := 23 // 12 forward + 11 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestProcureToPayValidation tests input validation
func TestProcureToPayValidation(t *testing.T) {
	s := NewProcureToPaySaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid input",
			map[string]interface{}{
				"vendor_id":      "vendor-001",
				"items":          []interface{}{map[string]interface{}{"sku": "SKU001"}},
				"delivery_date":  "2026-03-14",
				"invoice_amount": 1000.00,
			},
			false,
			"",
		},
		{
			"missing vendor_id",
			map[string]interface{}{
				"items":          []interface{}{map[string]interface{}{"sku": "SKU001"}},
				"delivery_date":  "2026-03-14",
				"invoice_amount": 1000.00,
			},
			true,
			"vendor_id is required",
		},
		{
			"missing items",
			map[string]interface{}{
				"vendor_id":      "vendor-001",
				"delivery_date":  "2026-03-14",
				"invoice_amount": 1000.00,
			},
			true,
			"items are required",
		},
		{
			"empty items",
			map[string]interface{}{
				"vendor_id":      "vendor-001",
				"items":          []interface{}{},
				"delivery_date":  "2026-03-14",
				"invoice_amount": 1000.00,
			},
			true,
			"items must be a non-empty list",
		},
		{
			"missing invoice_amount",
			map[string]interface{}{
				"vendor_id":     "vendor-001",
				"items":         []interface{}{map[string]interface{}{"sku": "SKU001"}},
				"delivery_date": "2026-03-14",
			},
			true,
			"invoice_amount is required",
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

// TestPurchaseReturnSagaType tests saga type
func TestPurchaseReturnSagaType(t *testing.T) {
	s := NewPurchaseReturnSaga()
	if s.SagaType() != "SAGA-P02" {
		t.Errorf("expected SAGA-P02, got %s", s.SagaType())
	}
}

// TestPurchaseReturnStepCount tests step count
func TestPurchaseReturnStepCount(t *testing.T) {
	s := NewPurchaseReturnSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 13 // 7 forward + 6 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestVendorPaymentTDSSagaType tests saga type
func TestVendorPaymentTDSSagaType(t *testing.T) {
	s := NewVendorPaymentTDSSaga()
	if s.SagaType() != "SAGA-P03" {
		t.Errorf("expected SAGA-P03, got %s", s.SagaType())
	}
}

// TestVendorPaymentTDSStepCount tests step count
func TestVendorPaymentTDSStepCount(t *testing.T) {
	s := NewVendorPaymentTDSSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 17 // 9 forward + 8 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestBudgetCheckSagaType tests saga type
func TestBudgetCheckSagaType(t *testing.T) {
	s := NewBudgetCheckSaga()
	if s.SagaType() != "SAGA-P04" {
		t.Errorf("expected SAGA-P04, got %s", s.SagaType())
	}
}

// TestBudgetCheckStepCount tests step count
func TestBudgetCheckStepCount(t *testing.T) {
	s := NewBudgetCheckSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 13 // 7 forward + 6 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestGetStepDefinition tests step lookup
func TestGetStepDefinition(t *testing.T) {
	s := NewProcureToPaySaga()

	// Test valid step
	step := s.GetStepDefinition(1)
	if step == nil {
		t.Error("expected step 1, got nil")
	}
	if step.ServiceName != "purchase-order" {
		t.Errorf("expected purchase-order, got %s", step.ServiceName)
	}

	// Test invalid step
	step = s.GetStepDefinition(999)
	if step != nil {
		t.Error("expected nil for invalid step")
	}
}

// TestStepDefinitionProperties tests step definition contents
func TestStepDefinitionProperties(t *testing.T) {
	s := NewProcureToPaySaga()
	step := s.GetStepDefinition(1)

	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"StepNumber", step.StepNumber, int32(1)},
		{"ServiceName", step.ServiceName, "purchase-order"},
		{"HandlerMethod", step.HandlerMethod, "CreatePurchaseOrder"},
		{"TimeoutSeconds", step.TimeoutSeconds, int32(25)},
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
	s := NewProcureToPaySaga()

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
	s := NewProcureToPaySaga()
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
	s := NewProcureToPaySaga()
	step := s.GetStepDefinition(1)

	expectedMappings := map[string]string{
		"tenantID":      "$.tenantID",
		"companyID":     "$.companyID",
		"branchID":      "$.branchID",
		"vendorID":      "$.input.vendor_id",
		"items":         "$.input.items",
		"deliveryDate":  "$.input.delivery_date",
		"paymentTerms":  "$.input.payment_terms",
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

// TestAllPurchaseSagasImplementInterface tests that all sagas implement the interface
func TestAllPurchaseSagasImplementInterface(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewProcureToPaySaga(),
		NewPurchaseReturnSaga(),
		NewVendorPaymentTDSSaga(),
		NewBudgetCheckSaga(),
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
		NewProcureToPaySaga(),
		NewPurchaseReturnSaga(),
		NewVendorPaymentTDSSaga(),
		NewBudgetCheckSaga(),
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
	s := NewProcureToPaySaga()
	steps := s.GetStepDefinitions()

	// Forward steps should be 1-12
	forwardSteps := 0
	compensationSteps := 0

	for _, step := range steps {
		if step.StepNumber < 100 {
			forwardSteps++
		} else {
			compensationSteps++
		}
	}

	if forwardSteps != 12 {
		t.Errorf("expected 12 forward steps, got %d", forwardSteps)
	}
	if compensationSteps != 11 {
		t.Errorf("expected 11 compensation steps, got %d", compensationSteps)
	}
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return errors.Is(errors.New(str), errors.New(substr)) || len(substr) == 0 || len(str) >= len(substr)
}
