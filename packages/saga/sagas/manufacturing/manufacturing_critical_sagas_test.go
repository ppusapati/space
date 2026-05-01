// Package manufacturing provides saga handlers for manufacturing module workflows
package manufacturing

import (
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// ===== PRODUCTION ORDER SAGA (M01) TESTS =====

// TestProductionOrderSagaType tests saga type identification
func TestProductionOrderSagaType(t *testing.T) {
	s := NewProductionOrderSaga()
	if s.SagaType() != "SAGA-M01" {
		t.Errorf("expected SAGA-M01, got %s", s.SagaType())
	}
}

// TestProductionOrderSagaStepCount tests total step count
func TestProductionOrderSagaStepCount(t *testing.T) {
	s := NewProductionOrderSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 19 // 10 forward + 9 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestProductionOrderSagaValidation tests input validation
func TestProductionOrderSagaValidation(t *testing.T) {
	s := NewProductionOrderSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			"valid input",
			map[string]interface{}{
				"order_id":       "PO-001",
				"product_id":     "PROD-001",
				"quantity":       100.0,
				"planned_date":   "2025-03-01",
				"cost_center_id": "CC-001",
			},
			false,
			"",
		},
		{
			"missing order_id",
			map[string]interface{}{
				"product_id":     "PROD-001",
				"quantity":       100.0,
				"planned_date":   "2025-03-01",
				"cost_center_id": "CC-001",
			},
			true,
			"order_id is required",
		},
		{
			"missing product_id",
			map[string]interface{}{
				"order_id":       "PO-001",
				"quantity":       100.0,
				"planned_date":   "2025-03-01",
				"cost_center_id": "CC-001",
			},
			true,
			"product_id is required",
		},
		{
			"missing quantity",
			map[string]interface{}{
				"order_id":       "PO-001",
				"product_id":     "PROD-001",
				"planned_date":   "2025-03-01",
				"cost_center_id": "CC-001",
			},
			true,
			"quantity is required",
		},
		{
			"missing planned_date",
			map[string]interface{}{
				"order_id":       "PO-001",
				"product_id":     "PROD-001",
				"quantity":       100.0,
				"cost_center_id": "CC-001",
			},
			true,
			"planned_date is required",
		},
		{
			"missing cost_center_id",
			map[string]interface{}{
				"order_id":     "PO-001",
				"product_id":   "PROD-001",
				"quantity":     100.0,
				"planned_date": "2025-03-01",
			},
			true,
			"cost_center_id is required",
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
		})
	}
}

// TestProductionOrderSagaCriticalSteps tests critical step identification
func TestProductionOrderSagaCriticalSteps(t *testing.T) {
	s := NewProductionOrderSaga()
	criticalSteps := []int{1, 2, 3, 4, 5, 10}

	for _, stepNum := range criticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be critical but is not", stepNum)
		}
	}

	// Verify non-critical steps
	nonCriticalSteps := []int{6, 7, 8, 9}
	for _, stepNum := range nonCriticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.IsCritical {
			t.Errorf("step %d should not be critical but is", stepNum)
		}
	}
}

// TestProductionOrderSagaTimeouts tests timeout configuration
func TestProductionOrderSagaTimeouts(t *testing.T) {
	s := NewProductionOrderSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber < 100 { // Only forward steps
			if step.TimeoutSeconds < 25 || step.TimeoutSeconds > 60 {
				t.Errorf("step %d timeout %d is out of expected range (25-60)", step.StepNumber, step.TimeoutSeconds)
			}
		}
	}
}

// TestProductionOrderSagaServiceNames tests service name format
func TestProductionOrderSagaServiceNames(t *testing.T) {
	s := NewProductionOrderSaga()
	expectedServices := map[int]string{
		1:  "production-order",
		2:  "inventory-core",
		3:  "shop-floor",
		4:  "routing",
		5:  "cost-center",
		6:  "shop-floor",
		7:  "quality-production",
		8:  "inventory-core",
		9:  "cost-center",
		10: "general-ledger",
	}

	for stepNum, expectedService := range expectedServices {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.ServiceName != expectedService {
			t.Errorf("step %d: expected service %s, got %s", stepNum, expectedService, step.ServiceName)
		}
	}
}

// TestProductionOrderSagaCompensationSteps tests compensation step configuration
func TestProductionOrderSagaCompensationSteps(t *testing.T) {
	s := NewProductionOrderSaga()

	// Verify steps 2-10 have compensation steps
	for stepNum := 2; stepNum <= 10; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if len(step.CompensationSteps) == 0 {
			t.Errorf("step %d should have compensation steps but has none", stepNum)
		}
	}

	// Verify step 1 has no compensation
	step1 := s.GetStepDefinition(1)
	if len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have no compensation steps but has %d", len(step1.CompensationSteps))
	}
}

// TestProductionOrderSagaFirstAndLastSteps tests first and last step properties
func TestProductionOrderSagaFirstAndLastSteps(t *testing.T) {
	s := NewProductionOrderSaga()

	// Step 1 should have no compensation
	step1 := s.GetStepDefinition(1)
	if len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have no compensation, but has %d", len(step1.CompensationSteps))
	}

	// Step 10 should have compensation
	step10 := s.GetStepDefinition(10)
	if len(step10.CompensationSteps) == 0 {
		t.Errorf("step 10 should have compensation steps")
	}
	if step10.CompensationSteps[0] != 110 {
		t.Errorf("step 10 compensation should be 110, got %d", step10.CompensationSteps[0])
	}
}

// TestProductionOrderSagaImplementsInterface tests interface implementation
func TestProductionOrderSagaImplementsInterface(t *testing.T) {
	var _ saga.SagaHandler = NewProductionOrderSaga()
}

// TestProductionOrderSagaGetStepByNumber tests step lookup by number
func TestProductionOrderSagaGetStepByNumber(t *testing.T) {
	s := NewProductionOrderSaga()

	// Test valid step
	step := s.GetStepDefinition(1)
	if step == nil {
		t.Error("expected step 1, got nil")
	}

	// Test invalid step
	step = s.GetStepDefinition(999)
	if step != nil {
		t.Error("expected nil for invalid step 999")
	}

	// Test compensation step
	compStep := s.GetStepDefinition(102)
	if compStep == nil {
		t.Error("expected compensation step 102, got nil")
	}
	if compStep.ServiceName != "inventory-core" {
		t.Errorf("expected compensation step 102 to be inventory-core, got %s", compStep.ServiceName)
	}
}

// TestProductionOrderSagaInputMapping tests input mapping configuration
func TestProductionOrderSagaInputMapping(t *testing.T) {
	s := NewProductionOrderSaga()
	step := s.GetStepDefinition(1)

	expectedMappings := map[string]string{
		"tenantID":     "$.tenantID",
		"companyID":    "$.companyID",
		"branchID":     "$.branchID",
		"orderID":      "$.input.order_id",
		"productID":    "$.input.product_id",
		"quantity":     "$.input.quantity",
		"plannedDate":  "$.input.planned_date",
		"costCenterID": "$.input.cost_center_id",
	}

	for key, expectedVal := range expectedMappings {
		actualVal, exists := step.InputMapping[key]
		if !exists {
			t.Errorf("missing input mapping for %s", key)
			continue
		}
		if actualVal != expectedVal {
			t.Errorf("%s: expected %s, got %s", key, expectedVal, actualVal)
		}
	}
}

// ===== SUBCONTRACTING SAGA (M02) TESTS =====

// TestSubcontractingSagaType tests saga type identification
func TestSubcontractingSagaType(t *testing.T) {
	s := NewSubcontractingSaga()
	if s.SagaType() != "SAGA-M02" {
		t.Errorf("expected SAGA-M02, got %s", s.SagaType())
	}
}

// TestSubcontractingSagaStepCount tests total step count
func TestSubcontractingSagaStepCount(t *testing.T) {
	s := NewSubcontractingSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 19 // 10 forward + 9 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestSubcontractingSagaValidation tests input validation
func TestSubcontractingSagaValidation(t *testing.T) {
	s := NewSubcontractingSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			"valid input",
			map[string]interface{}{
				"order_id":         "SCON-001",
				"subcontractor_id": "SUB-001",
				"item_id":          "ITEM-001",
				"quantity":         50.0,
				"unit_cost":        1000.0,
			},
			false,
			"",
		},
		{
			"missing order_id",
			map[string]interface{}{
				"subcontractor_id": "SUB-001",
				"item_id":          "ITEM-001",
				"quantity":         50.0,
				"unit_cost":        1000.0,
			},
			true,
			"order_id is required",
		},
		{
			"missing subcontractor_id",
			map[string]interface{}{
				"order_id":  "SCON-001",
				"item_id":   "ITEM-001",
				"quantity":  50.0,
				"unit_cost": 1000.0,
			},
			true,
			"subcontractor_id is required",
		},
		{
			"missing item_id",
			map[string]interface{}{
				"order_id":         "SCON-001",
				"subcontractor_id": "SUB-001",
				"quantity":         50.0,
				"unit_cost":        1000.0,
			},
			true,
			"item_id is required",
		},
		{
			"missing quantity",
			map[string]interface{}{
				"order_id":         "SCON-001",
				"subcontractor_id": "SUB-001",
				"item_id":          "ITEM-001",
				"unit_cost":        1000.0,
			},
			true,
			"quantity is required",
		},
		{
			"missing unit_cost",
			map[string]interface{}{
				"order_id":         "SCON-001",
				"subcontractor_id": "SUB-001",
				"item_id":          "ITEM-001",
				"quantity":         50.0,
			},
			true,
			"unit_cost is required",
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
		})
	}
}

// TestSubcontractingSagaCriticalSteps tests critical step identification
func TestSubcontractingSagaCriticalSteps(t *testing.T) {
	s := NewSubcontractingSaga()
	criticalSteps := []int{1, 2, 3, 4, 5, 8, 10}

	for _, stepNum := range criticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be critical but is not", stepNum)
		}
	}

	// Verify non-critical steps
	nonCriticalSteps := []int{6, 7, 9}
	for _, stepNum := range nonCriticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.IsCritical {
			t.Errorf("step %d should not be critical but is", stepNum)
		}
	}
}

// TestSubcontractingSagaTimeouts tests timeout configuration
func TestSubcontractingSagaTimeouts(t *testing.T) {
	s := NewSubcontractingSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber < 100 { // Only forward steps
			if step.TimeoutSeconds < 30 || step.TimeoutSeconds > 45 {
				t.Errorf("step %d timeout %d is out of expected range (30-45)", step.StepNumber, step.TimeoutSeconds)
			}
		}
	}
}

// TestSubcontractingSagaServiceNames tests service name format
func TestSubcontractingSagaServiceNames(t *testing.T) {
	s := NewSubcontractingSaga()
	expectedServices := map[int]string{
		1:  "subcontracting",
		2:  "purchase-order",
		3:  "inventory-core",
		4:  "quality-production",
		5:  "accounts-payable",
		6:  "subcontracting",
		7:  "quality-production",
		8:  "inventory-core",
		9:  "subcontracting",
		10: "general-ledger",
	}

	for stepNum, expectedService := range expectedServices {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.ServiceName != expectedService {
			t.Errorf("step %d: expected service %s, got %s", stepNum, expectedService, step.ServiceName)
		}
	}
}

// TestSubcontractingSagaCompensationSteps tests compensation step configuration
func TestSubcontractingSagaCompensationSteps(t *testing.T) {
	s := NewSubcontractingSaga()

	// Verify steps 2-10 have compensation steps
	for stepNum := 2; stepNum <= 10; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if len(step.CompensationSteps) == 0 {
			t.Errorf("step %d should have compensation steps but has none", stepNum)
		}
	}

	// Verify step 1 has no compensation
	step1 := s.GetStepDefinition(1)
	if len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have no compensation steps but has %d", len(step1.CompensationSteps))
	}
}

// TestSubcontractingSagaSpecialRetry tests special retry configuration for step 6
func TestSubcontractingSagaSpecialRetry(t *testing.T) {
	s := NewSubcontractingSaga()
	step6 := s.GetStepDefinition(6)

	if step6 == nil {
		t.Fatal("step 6 not found")
	}
	if step6.RetryConfig == nil {
		t.Fatal("step 6 retry config is nil")
	}

	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"MaxRetries", step6.RetryConfig.MaxRetries, int32(5)},
		{"InitialBackoffMs", step6.RetryConfig.InitialBackoffMs, int32(2000)},
		{"MaxBackoffMs", step6.RetryConfig.MaxBackoffMs, int32(60000)},
	}

	for _, tt := range tests {
		if tt.actual != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.actual)
		}
	}
}

// TestSubcontractingSagaFirstAndLastSteps tests first and last step properties
func TestSubcontractingSagaFirstAndLastSteps(t *testing.T) {
	s := NewSubcontractingSaga()

	// Step 1 should have no compensation
	step1 := s.GetStepDefinition(1)
	if len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have no compensation, but has %d", len(step1.CompensationSteps))
	}

	// Step 10 should have compensation
	step10 := s.GetStepDefinition(10)
	if len(step10.CompensationSteps) == 0 {
		t.Errorf("step 10 should have compensation steps")
	}
	if step10.CompensationSteps[0] != 110 {
		t.Errorf("step 10 compensation should be 110, got %d", step10.CompensationSteps[0])
	}
}

// TestSubcontractingSagaImplementsInterface tests interface implementation
func TestSubcontractingSagaImplementsInterface(t *testing.T) {
	var _ saga.SagaHandler = NewSubcontractingSaga()
}

// TestSubcontractingSagaGetStepByNumber tests step lookup by number
func TestSubcontractingSagaGetStepByNumber(t *testing.T) {
	s := NewSubcontractingSaga()

	// Test valid step
	step := s.GetStepDefinition(1)
	if step == nil {
		t.Error("expected step 1, got nil")
	}

	// Test invalid step
	step = s.GetStepDefinition(999)
	if step != nil {
		t.Error("expected nil for invalid step 999")
	}

	// Test compensation step
	compStep := s.GetStepDefinition(102)
	if compStep == nil {
		t.Error("expected compensation step 102, got nil")
	}
	if compStep.ServiceName != "purchase-order" {
		t.Errorf("expected compensation step 102 to be purchase-order, got %s", compStep.ServiceName)
	}
}

// TestSubcontractingSagaInputMapping tests input mapping configuration
func TestSubcontractingSagaInputMapping(t *testing.T) {
	s := NewSubcontractingSaga()
	step := s.GetStepDefinition(1)

	expectedMappings := map[string]string{
		"tenantID":        "$.tenantID",
		"companyID":       "$.companyID",
		"branchID":        "$.branchID",
		"orderID":         "$.input.order_id",
		"subcontractorID": "$.input.subcontractor_id",
		"itemID":          "$.input.item_id",
		"quantity":        "$.input.quantity",
		"unitCost":        "$.input.unit_cost",
	}

	for key, expectedVal := range expectedMappings {
		actualVal, exists := step.InputMapping[key]
		if !exists {
			t.Errorf("missing input mapping for %s", key)
			continue
		}
		if actualVal != expectedVal {
			t.Errorf("%s: expected %s, got %s", key, expectedVal, actualVal)
		}
	}
}

// ===== CROSS-SAGA TESTS =====

// TestAllManufacturingSagasImplementInterface tests that all sagas implement the interface
func TestAllManufacturingSagasImplementInterface(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewProductionOrderSaga(),
		NewSubcontractingSaga(),
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
		NewProductionOrderSaga(),
		NewSubcontractingSaga(),
	}

	seen := make(map[string]bool)
	for _, s := range sagas {
		sagaType := s.SagaType()
		if seen[sagaType] {
			t.Errorf("duplicate saga type: %s", sagaType)
		}
		seen[sagaType] = true
	}

	expectedCount := 2
	if len(seen) != expectedCount {
		t.Errorf("expected %d unique saga types, got %d", expectedCount, len(seen))
	}
}

// TestStepSequencing tests that steps are properly sequenced
func TestStepSequencing(t *testing.T) {
	tests := []struct {
		name            string
		saga            saga.SagaHandler
		expectedForward int
		expectedComp    int
	}{
		{
			"ProductionOrderSaga",
			NewProductionOrderSaga(),
			10,
			9,
		},
		{
			"SubcontractingSaga",
			NewSubcontractingSaga(),
			10,
			9,
		},
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
			if compensationSteps != tt.expectedComp {
				t.Errorf("expected %d compensation steps, got %d", tt.expectedComp, compensationSteps)
			}
		})
	}
}
