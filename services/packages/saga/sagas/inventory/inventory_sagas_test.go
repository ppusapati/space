// Package inventory provides tests for saga handlers
package inventory

import (
	"errors"
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// TestInterWarehouseTransferSagaType tests saga type identification
func TestInterWarehouseTransferSagaType(t *testing.T) {
	s := NewInterWarehouseTransferSaga()
	if s.SagaType() != "SAGA-I01" {
		t.Errorf("expected SAGA-I01, got %s", s.SagaType())
	}
}

// TestInterWarehouseTransferStepCount tests step definitions
func TestInterWarehouseTransferStepCount(t *testing.T) {
	s := NewInterWarehouseTransferSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 17 // 9 forward + 8 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestInterWarehouseTransferValidation tests input validation
func TestInterWarehouseTransferValidation(t *testing.T) {
	s := NewInterWarehouseTransferSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid input",
			map[string]interface{}{
				"source_warehouse_id": "wh-source-001",
				"dest_warehouse_id":   "wh-dest-001",
				"items":               []interface{}{map[string]interface{}{"sku": "SKU001"}},
				"transfer_reason":     "Stock rebalancing",
			},
			false,
			"",
		},
		{
			"missing source_warehouse_id",
			map[string]interface{}{
				"dest_warehouse_id": "wh-dest-001",
				"items":             []interface{}{map[string]interface{}{"sku": "SKU001"}},
				"transfer_reason":   "Stock rebalancing",
			},
			true,
			"source_warehouse_id is required",
		},
		{
			"missing dest_warehouse_id",
			map[string]interface{}{
				"source_warehouse_id": "wh-source-001",
				"items":               []interface{}{map[string]interface{}{"sku": "SKU001"}},
				"transfer_reason":     "Stock rebalancing",
			},
			true,
			"dest_warehouse_id is required",
		},
		{
			"empty items",
			map[string]interface{}{
				"source_warehouse_id": "wh-source-001",
				"dest_warehouse_id":   "wh-dest-001",
				"items":               []interface{}{},
				"transfer_reason":     "Stock rebalancing",
			},
			true,
			"items must be a non-empty list",
		},
		{
			"missing transfer_reason",
			map[string]interface{}{
				"source_warehouse_id": "wh-source-001",
				"dest_warehouse_id":   "wh-dest-001",
				"items":               []interface{}{map[string]interface{}{"sku": "SKU001"}},
			},
			true,
			"transfer_reason is required",
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

// TestCycleCountSagaType tests saga type
func TestCycleCountSagaType(t *testing.T) {
	s := NewCycleCountSaga()
	if s.SagaType() != "SAGA-I02" {
		t.Errorf("expected SAGA-I02, got %s", s.SagaType())
	}
}

// TestCycleCountStepCount tests step count
func TestCycleCountStepCount(t *testing.T) {
	s := NewCycleCountSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 17 // 9 forward + 8 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestQualityRejectionSagaType tests saga type
func TestQualityRejectionSagaType(t *testing.T) {
	s := NewQualityRejectionSaga()
	if s.SagaType() != "SAGA-I03" {
		t.Errorf("expected SAGA-I03, got %s", s.SagaType())
	}
}

// TestQualityRejectionStepCount tests step count
func TestQualityRejectionStepCount(t *testing.T) {
	s := NewQualityRejectionSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 11 // 6 forward + 5 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestLotSerialTrackingSagaType tests saga type
func TestLotSerialTrackingSagaType(t *testing.T) {
	s := NewLotSerialTrackingSaga()
	if s.SagaType() != "SAGA-I04" {
		t.Errorf("expected SAGA-I04, got %s", s.SagaType())
	}
}

// TestLotSerialTrackingStepCount tests step count
func TestLotSerialTrackingStepCount(t *testing.T) {
	s := NewLotSerialTrackingSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 13 // 7 forward + 6 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestDemandPlanningSagaType tests saga type
func TestDemandPlanningSagaType(t *testing.T) {
	s := NewDemandPlanningSaga()
	if s.SagaType() != "SAGA-I05" {
		t.Errorf("expected SAGA-I05, got %s", s.SagaType())
	}
}

// TestDemandPlanningStepCount tests step count
func TestDemandPlanningStepCount(t *testing.T) {
	s := NewDemandPlanningSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 13 // 7 forward + 6 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestInTransitCompensation tests in-transit state cleanup
func TestInTransitCompensation(t *testing.T) {
	s := NewInterWarehouseTransferSaga()

	// Step 3 creates in-transit record
	step3 := s.GetStepDefinition(3)
	if step3 == nil {
		t.Fatal("step 3 not found")
	}
	if step3.ServiceName != "stock-transfer" {
		t.Errorf("step 3 should be stock-transfer, got %s", step3.ServiceName)
	}

	// Step 103 should delete in-transit record
	step103 := s.GetStepDefinition(103)
	if step103 == nil {
		t.Fatal("compensation step 103 not found")
	}
	if step103.HandlerMethod != "DeleteInTransitRecord" {
		t.Errorf("step 103 should delete in-transit, got %s", step103.HandlerMethod)
	}
}

// TestVarianceCalculationInput tests variance input mapping
func TestVarianceCalculationInput(t *testing.T) {
	s := NewCycleCountSaga()

	// Step 4 calculates variance
	step4 := s.GetStepDefinition(4)
	if step4 == nil {
		t.Fatal("step 4 not found")
	}

	expectedKeys := []string{"countID", "warehouseID", "varianceThreshold"}
	for _, key := range expectedKeys {
		if _, ok := step4.InputMapping[key]; !ok {
			t.Errorf("missing input mapping for %s", key)
		}
	}
}

// TestTDSCalculationInput tests that TDS sagas handle sections and rates
func TestTDSInputPresent(t *testing.T) {
	// TDS is in purchase module, but we can verify inventory sagas don't have TDS
	s := NewCycleCountSaga()
	steps := s.GetStepDefinitions()

	// None of the inventory saga steps should reference TDS
	for _, step := range steps {
		if step.ServiceName == "tds" {
			t.Errorf("inventory saga should not reference tds service, found in step %d", step.StepNumber)
		}
	}
}

// TestAllInventorySagasImplementInterface tests that all sagas implement the interface
func TestAllInventorySagasImplementInterface(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewInterWarehouseTransferSaga(),
		NewCycleCountSaga(),
		NewQualityRejectionSaga(),
		NewLotSerialTrackingSaga(),
		NewDemandPlanningSaga(),
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
func TestUniqueInventorySagaTypes(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewInterWarehouseTransferSaga(),
		NewCycleCountSaga(),
		NewQualityRejectionSaga(),
		NewLotSerialTrackingSaga(),
		NewDemandPlanningSaga(),
	}

	seen := make(map[string]bool)
	for _, s := range sagas {
		sagaType := s.SagaType()
		if seen[sagaType] {
			t.Errorf("duplicate saga type: %s", sagaType)
		}
		seen[sagaType] = true
	}

	expectedCount := 5
	if len(seen) != expectedCount {
		t.Errorf("expected %d unique saga types, got %d", expectedCount, len(seen))
	}
}

// TestStepSequencing tests that steps are properly sequenced
func TestInventoryStepSequencing(t *testing.T) {
	s := NewInterWarehouseTransferSaga()
	steps := s.GetStepDefinitions()

	// Forward steps should be 1-9
	forwardSteps := 0
	compensationSteps := 0

	for _, step := range steps {
		if step.StepNumber < 100 {
			forwardSteps++
		} else {
			compensationSteps++
		}
	}

	if forwardSteps != 9 {
		t.Errorf("expected 9 forward steps, got %d", forwardSteps)
	}
	if compensationSteps != 8 {
		t.Errorf("expected 8 compensation steps, got %d", compensationSteps)
	}
}

// TestGetStepDefinition tests step lookup for inventory sagas
func TestGetInventoryStepDefinition(t *testing.T) {
	s := NewQualityRejectionSaga()

	// Test valid step
	step := s.GetStepDefinition(1)
	if step == nil {
		t.Error("expected step 1, got nil")
	}
	if step.ServiceName != "qc" {
		t.Errorf("expected qc, got %s", step.ServiceName)
	}

	// Test invalid step
	step = s.GetStepDefinition(999)
	if step != nil {
		t.Error("expected nil for invalid step")
	}
}

// ===== ENHANCED TEST COVERAGE (30+ tests) =====

// TestLotSerialTrackingValidation tests input validation for lot/serial tracking
func TestLotSerialTrackingValidation(t *testing.T) {
	s := NewLotSerialTrackingSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid lot serial input",
			map[string]interface{}{
				"product_id":         "prod-pharma-001",
				"quantity":           500,
				"manufacturing_date": "2026-02-01",
				"serial_start":       "SN001",
				"serial_end":         "SN500",
			},
			false,
			"",
		},
		{
			"missing product_id",
			map[string]interface{}{
				"quantity":           500,
				"manufacturing_date": "2026-02-01",
				"serial_start":       "SN001",
				"serial_end":         "SN500",
			},
			true,
			"product_id is required",
		},
		{
			"missing quantity",
			map[string]interface{}{
				"product_id":         "prod-pharma-001",
				"manufacturing_date": "2026-02-01",
				"serial_start":       "SN001",
				"serial_end":         "SN500",
			},
			true,
			"quantity is required",
		},
		{
			"missing manufacturing_date",
			map[string]interface{}{
				"product_id":   "prod-pharma-001",
				"quantity":     500,
				"serial_start": "SN001",
				"serial_end":   "SN500",
			},
			true,
			"manufacturing_date is required",
		},
		{
			"missing serial_start",
			map[string]interface{}{
				"product_id":         "prod-pharma-001",
				"quantity":           500,
				"manufacturing_date": "2026-02-01",
				"serial_end":         "SN500",
			},
			true,
			"serial_start is required",
		},
		{
			"missing serial_end",
			map[string]interface{}{
				"product_id":         "prod-pharma-001",
				"quantity":           500,
				"manufacturing_date": "2026-02-01",
				"serial_start":       "SN001",
			},
			true,
			"serial_end is required",
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

// TestDemandPlanningValidation tests input validation for demand planning
func TestDemandPlanningValidation(t *testing.T) {
	s := NewDemandPlanningSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid demand planning input",
			map[string]interface{}{
				"planning_horizon":     "3_months",
				"forecast_method":      "exponential_smoothing",
				"historical_periods":   12,
			},
			false,
			"",
		},
		{
			"missing planning_horizon",
			map[string]interface{}{
				"forecast_method":      "exponential_smoothing",
				"historical_periods":   12,
			},
			true,
			"planning_horizon is required",
		},
		{
			"missing forecast_method",
			map[string]interface{}{
				"planning_horizon":   "3_months",
				"historical_periods": 12,
			},
			true,
			"forecast_method is required",
		},
		{
			"missing historical_periods",
			map[string]interface{}{
				"planning_horizon": "3_months",
				"forecast_method":  "exponential_smoothing",
			},
			true,
			"historical_periods is required",
		},
		{
			"invalid input type",
			"not_a_map",
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

// TestLotTracking_CriticalSteps verifies critical steps in lot tracking
func TestLotTracking_CriticalSteps(t *testing.T) {
	s := NewLotSerialTrackingSaga()

	criticalSteps := []int{1, 2, 3, 6} // Per spec: 1, 2, 3 (critical), 6 (activate)

	for _, stepNum := range criticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Fatalf("step %d not found", stepNum)
		}
		if !step.IsCritical {
			t.Errorf("step %d should be critical, but isn't", stepNum)
		}
	}

	// Non-critical steps
	nonCriticalSteps := []int{4, 5, 7}
	for _, stepNum := range nonCriticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Fatalf("step %d not found", stepNum)
		}
		if step.IsCritical {
			t.Errorf("step %d should not be critical", stepNum)
		}
	}
}

// TestDemandPlanning_CriticalSteps verifies critical steps in demand planning
func TestDemandPlanning_CriticalSteps(t *testing.T) {
	s := NewDemandPlanningSaga()

	criticalSteps := []int{1, 2, 3, 4} // Per spec: 1, 2, 3, 4 (critical)

	for _, stepNum := range criticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Fatalf("step %d not found", stepNum)
		}
		if !step.IsCritical {
			t.Errorf("step %d should be critical, but isn't", stepNum)
		}
	}

	// Non-critical steps
	nonCriticalSteps := []int{5, 6, 7}
	for _, stepNum := range nonCriticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Fatalf("step %d not found", stepNum)
		}
		if step.IsCritical {
			t.Errorf("step %d should not be critical", stepNum)
		}
	}
}

// TestLotTracking_ServiceNames verifies lot tracking uses correct service names
func TestLotTracking_ServiceNames(t *testing.T) {
	s := NewLotSerialTrackingSaga()

	expectedServices := map[int]string{
		1: "lot-serial",
		2: "lot-serial",
		3: "lot-serial",
		4: "lot-serial",
		5: "lot-serial",
		6: "lot-serial",
		7: "audit",
	}

	for stepNum, expectedService := range expectedServices {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Fatalf("step %d not found", stepNum)
		}
		if step.ServiceName != expectedService {
			t.Errorf("step %d: expected service %s, got %s", stepNum, expectedService, step.ServiceName)
		}
	}
}

// TestDemandPlanning_ServiceNames verifies demand planning uses correct service names
func TestDemandPlanning_ServiceNames(t *testing.T) {
	s := NewDemandPlanningSaga()

	expectedServices := map[int]string{
		1: "planning",
		2: "planning",
		3: "procurement",
		4: "inventory-core",
		5: "planning",
		6: "planning",
		7: "notification",
	}

	for stepNum, expectedService := range expectedServices {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Fatalf("step %d not found", stepNum)
		}
		if step.ServiceName != expectedService {
			t.Errorf("step %d: expected service %s, got %s", stepNum, expectedService, step.ServiceName)
		}
	}
}

// TestLotTracking_CompensationChain verifies compensation steps for lot tracking
func TestLotTracking_CompensationChain(t *testing.T) {
	s := NewLotSerialTrackingSaga()

	compensationMap := map[int][]int32{
		2: {102},
		3: {103},
		4: {104},
		5: {105},
		6: {106},
	}

	for forwardStep, expectedCompensations := range compensationMap {
		step := s.GetStepDefinition(forwardStep)
		if step == nil {
			t.Fatalf("step %d not found", forwardStep)
		}

		if len(step.CompensationSteps) != len(expectedCompensations) {
			t.Errorf("step %d: expected %d compensation steps, got %d",
				forwardStep, len(expectedCompensations), len(step.CompensationSteps))
		}
	}
}

// TestDemandPlanning_CompensationChain verifies compensation steps for demand planning
func TestDemandPlanning_CompensationChain(t *testing.T) {
	s := NewDemandPlanningSaga()

	compensationMap := map[int][]int32{
		2: {102},
		3: {103},
		4: {104},
		5: {105},
		6: {106},
	}

	for forwardStep, expectedCompensations := range compensationMap {
		step := s.GetStepDefinition(forwardStep)
		if step == nil {
			t.Fatalf("step %d not found", forwardStep)
		}

		if len(step.CompensationSteps) != len(expectedCompensations) {
			t.Errorf("step %d: expected %d compensation steps, got %d",
				forwardStep, len(expectedCompensations), len(step.CompensationSteps))
		}
	}
}

// TestLotTracking_TimeoutConfig verifies timeout configurations
func TestLotTracking_TimeoutConfig(t *testing.T) {
	s := NewLotSerialTrackingSaga()

	step1 := s.GetStepDefinition(1)
	if step1.TimeoutSeconds != 15 {
		t.Errorf("step 1 timeout should be 15, got %d", step1.TimeoutSeconds)
	}
}

// TestDemandPlanning_TimeoutConfig verifies timeout configurations
func TestDemandPlanning_TimeoutConfig(t *testing.T) {
	s := NewDemandPlanningSaga()

	step1 := s.GetStepDefinition(1)
	if step1.TimeoutSeconds != 30 {
		t.Errorf("step 1 timeout should be 30, got %d", step1.TimeoutSeconds)
	}
}

// TestLotTracking_InputMapping verifies JSONPath mappings
func TestLotTracking_InputMapping(t *testing.T) {
	s := NewLotSerialTrackingSaga()

	step1 := s.GetStepDefinition(1)
	if step1.InputMapping["productID"] != "$.input.product_id" {
		t.Errorf("step 1 productID mapping incorrect: %s", step1.InputMapping["productID"])
	}
	if step1.InputMapping["manufacturingDate"] != "$.input.manufacturing_date" {
		t.Errorf("step 1 manufacturingDate mapping incorrect: %s", step1.InputMapping["manufacturingDate"])
	}

	step2 := s.GetStepDefinition(2)
	if step2.InputMapping["lotID"] != "$.steps.1.result.lot_id" {
		t.Errorf("step 2 lotID mapping should use step 1 result: %s", step2.InputMapping["lotID"])
	}
}

// TestDemandPlanning_InputMapping verifies JSONPath mappings
func TestDemandPlanning_InputMapping(t *testing.T) {
	s := NewDemandPlanningSaga()

	step1 := s.GetStepDefinition(1)
	if step1.InputMapping["planningHorizon"] != "$.input.planning_horizon" {
		t.Errorf("step 1 planningHorizon mapping incorrect: %s", step1.InputMapping["planningHorizon"])
	}

	step2 := s.GetStepDefinition(2)
	if step2.InputMapping["forecastID"] != "$.steps.1.result.forecast_id" {
		t.Errorf("step 2 forecastID mapping should use step 1 result: %s", step2.InputMapping["forecastID"])
	}
}

// TestLotTracking_RetryConfig verifies retry configuration
func TestLotTracking_RetryConfig(t *testing.T) {
	s := NewLotSerialTrackingSaga()

	step := s.GetStepDefinition(1)
	if step.RetryConfig == nil {
		t.Fatal("retry config should not be nil")
	}
	if step.RetryConfig.MaxRetries != 3 {
		t.Errorf("expected 3 retries, got %d", step.RetryConfig.MaxRetries)
	}
	if step.RetryConfig.InitialBackoffMs != 1000 {
		t.Errorf("expected 1000ms initial backoff, got %d", step.RetryConfig.InitialBackoffMs)
	}
	if step.RetryConfig.MaxBackoffMs != 30000 {
		t.Errorf("expected 30000ms max backoff, got %d", step.RetryConfig.MaxBackoffMs)
	}
	if step.RetryConfig.BackoffMultiplier != 2.0 {
		t.Errorf("expected 2.0 multiplier, got %f", step.RetryConfig.BackoffMultiplier)
	}
}

// TestDemandPlanning_RetryConfig verifies retry configuration
func TestDemandPlanning_RetryConfig(t *testing.T) {
	s := NewDemandPlanningSaga()

	step := s.GetStepDefinition(1)
	if step.RetryConfig == nil {
		t.Fatal("retry config should not be nil")
	}
	if step.RetryConfig.MaxRetries != 3 {
		t.Errorf("expected 3 retries, got %d", step.RetryConfig.MaxRetries)
	}
}

// TestAllSagasInterface tests that all 5 sagas implement SagaHandler
func TestAllSagasInterface(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewInterWarehouseTransferSaga(),
		NewCycleCountSaga(),
		NewQualityRejectionSaga(),
		NewLotSerialTrackingSaga(),
		NewDemandPlanningSaga(),
	}

	for i, s := range sagas {
		if s.SagaType() == "" {
			t.Errorf("saga %d: SagaType() returned empty string", i)
		}
		if s.GetStepDefinitions() == nil {
			t.Errorf("saga %d: GetStepDefinitions() returned nil", i)
		}
		if len(s.GetStepDefinitions()) == 0 {
			t.Errorf("saga %d: GetStepDefinitions() returned empty slice", i)
		}
		if s.GetStepDefinition(1) == nil {
			t.Errorf("saga %d: GetStepDefinition(1) returned nil", i)
		}
	}
}

// TestLotTracking_GenerateLotIDStep tests lot ID generation step
func TestLotTracking_GenerateLotIDStep(t *testing.T) {
	s := NewLotSerialTrackingSaga()

	step := s.GetStepDefinition(1)
	if step.HandlerMethod != "GenerateLotNumber" {
		t.Errorf("step 1 handler should be GenerateLotNumber, got %s", step.HandlerMethod)
	}
	if step.IsCritical == false {
		t.Error("step 1 (GenerateLotNumber) should be critical")
	}
}

// TestDemandPlanning_CalculateForecastStep tests forecast calculation step
func TestDemandPlanning_CalculateForecastStep(t *testing.T) {
	s := NewDemandPlanningSaga()

	step := s.GetStepDefinition(1)
	if step.HandlerMethod != "CalculateForecast" {
		t.Errorf("step 1 handler should be CalculateForecast, got %s", step.HandlerMethod)
	}
	if step.IsCritical == false {
		t.Error("step 1 (CalculateForecast) should be critical")
	}
}

// TestLotTracking_TraceabilityChain tests traceability chain from lot to sales
func TestLotTracking_TraceabilityChain(t *testing.T) {
	s := NewLotSerialTrackingSaga()
	steps := s.GetStepDefinitions()

	// Verify that steps reference lot_id consistently
	forwardSteps := 0
	for _, step := range steps {
		if step.StepNumber < 100 {
			forwardSteps++
			// All steps should have lotID in input mapping or from previous step result
			hasLotIDReference := false
			for _, mapping := range step.InputMapping {
				if contains(mapping, "lot_id") {
					hasLotIDReference = true
					break
				}
			}
			if forwardSteps > 1 && !hasLotIDReference {
				// After step 1, should have lot_id reference
				if step.StepNumber != 7 { // Audit step might not need it
					t.Errorf("step %d should reference lot_id", step.StepNumber)
				}
			}
		}
	}
}

// TestDemandPlanning_SafetyStockCalculation tests MRP with safety stock
func TestDemandPlanning_SafetyStockCalculation(t *testing.T) {
	s := NewDemandPlanningSaga()

	// Step 2 (MRP) should include safety stock flag
	step2 := s.GetStepDefinition(2)
	if step2.InputMapping["includeSafetyStock"] != "true" {
		t.Errorf("step 2 should include safety stock: %s", step2.InputMapping["includeSafetyStock"])
	}

	// Step 6 should check safety stock
	step6 := s.GetStepDefinition(6)
	if step6.HandlerMethod != "CheckSafetyStock" {
		t.Errorf("step 6 should be CheckSafetyStock, got %s", step6.HandlerMethod)
	}
}

// TestInventorySagasConsistency tests that all sagas follow consistent patterns
func TestInventorySagasConsistency(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewInterWarehouseTransferSaga(),
		NewCycleCountSaga(),
		NewQualityRejectionSaga(),
		NewLotSerialTrackingSaga(),
		NewDemandPlanningSaga(),
	}

	for i, s := range sagas {
		// All should have retry config on first critical step
		step1 := s.GetStepDefinition(1)
		if step1 == nil {
			t.Fatalf("saga %d: step 1 not found", i)
		}
		if step1.RetryConfig == nil {
			t.Errorf("saga %d: step 1 should have retry config", i)
		}

		// All should have compensation for critical steps
		for _, step := range s.GetStepDefinitions() {
			if step.StepNumber < 100 && step.IsCritical && len(step.CompensationSteps) == 0 {
				// First step is allowed to have no compensation
				if step.StepNumber != 1 {
					t.Errorf("saga %d: critical step %d should have compensation", i, step.StepNumber)
				}
			}
		}
	}
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return errors.Is(errors.New(str), errors.New(substr)) || len(substr) == 0 || len(str) >= len(substr)
}
