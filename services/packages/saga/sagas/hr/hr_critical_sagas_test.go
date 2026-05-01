// Package hr provides saga handlers for HR & Payroll module workflows
package hr

import (
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// ===== PAYROLL PROCESSING SAGA (H01) TESTS =====

// TestPayrollProcessingSagaType tests saga type identification
func TestPayrollProcessingSagaType(t *testing.T) {
	s := NewPayrollProcessingSaga()
	if s.SagaType() != "SAGA-H01" {
		t.Errorf("expected SAGA-H01, got %s", s.SagaType())
	}
}

// TestPayrollProcessingSagaStepCount tests total step count
func TestPayrollProcessingSagaStepCount(t *testing.T) {
	s := NewPayrollProcessingSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 25 // 13 forward + 12 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestPayrollProcessingSagaValidation tests input validation
func TestPayrollProcessingSagaValidation(t *testing.T) {
	s := NewPayrollProcessingSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			"valid input",
			map[string]interface{}{
				"payroll_period_id": "PP-2025-03",
				"processing_date":   "2025-03-31",
				"employee_list":     []interface{}{map[string]interface{}{"employee_id": "EMP-001"}},
				"deduction_config":  map[string]interface{}{"advance": 5000.0},
			},
			false,
			"",
		},
		{
			"missing payroll_period_id",
			map[string]interface{}{
				"processing_date":  "2025-03-31",
				"employee_list":    []interface{}{map[string]interface{}{"employee_id": "EMP-001"}},
				"deduction_config": map[string]interface{}{"advance": 5000.0},
			},
			true,
			"payroll_period_id is required",
		},
		{
			"missing processing_date",
			map[string]interface{}{
				"payroll_period_id": "PP-2025-03",
				"employee_list":     []interface{}{map[string]interface{}{"employee_id": "EMP-001"}},
				"deduction_config":  map[string]interface{}{"advance": 5000.0},
			},
			true,
			"processing_date is required",
		},
		{
			"missing employee_list",
			map[string]interface{}{
				"payroll_period_id": "PP-2025-03",
				"processing_date":   "2025-03-31",
				"deduction_config":  map[string]interface{}{"advance": 5000.0},
			},
			true,
			"employee_list is required",
		},
		{
			"empty employee_list",
			map[string]interface{}{
				"payroll_period_id": "PP-2025-03",
				"processing_date":   "2025-03-31",
				"employee_list":     []interface{}{},
				"deduction_config":  map[string]interface{}{"advance": 5000.0},
			},
			true,
			"employee_list must be a non-empty array",
		},
		{
			"missing deduction_config",
			map[string]interface{}{
				"payroll_period_id": "PP-2025-03",
				"processing_date":   "2025-03-31",
				"employee_list":     []interface{}{map[string]interface{}{"employee_id": "EMP-001"}},
			},
			true,
			"deduction_config is required",
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

// TestPayrollProcessingSagaCriticalSteps tests critical step identification
func TestPayrollProcessingSagaCriticalSteps(t *testing.T) {
	s := NewPayrollProcessingSaga()
	criticalSteps := []int{1, 2, 3, 4, 5, 8, 13}

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
	nonCriticalSteps := []int{6, 7, 9, 10, 11, 12}
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

// TestPayrollProcessingSagaTimeouts tests timeout configuration
func TestPayrollProcessingSagaTimeouts(t *testing.T) {
	s := NewPayrollProcessingSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber < 100 { // Only forward steps
			if step.TimeoutSeconds < 30 || step.TimeoutSeconds > 90 {
				t.Errorf("step %d timeout %d is out of expected range (30-90)", step.StepNumber, step.TimeoutSeconds)
			}
		}
	}
}

// TestPayrollProcessingSagaServiceNames tests service name format with hyphens
func TestPayrollProcessingSagaServiceNames(t *testing.T) {
	s := NewPayrollProcessingSaga()
	expectedServices := map[int]string{
		1:  "attendance",
		2:  "leave",
		3:  "salary-structure",
		4:  "tds",
		5:  "payroll",
		6:  "payroll",
		7:  "payroll",
		8:  "banking",
		9:  "general-ledger",
		10: "tds",
		11: "employee",
		12: "notification",
		13: "payroll",
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

// TestPayrollProcessingSagaCompensationSteps tests compensation step configuration
func TestPayrollProcessingSagaCompensationSteps(t *testing.T) {
	s := NewPayrollProcessingSaga()

	// Verify steps 2-12 have compensation steps
	for stepNum := 2; stepNum <= 12; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if len(step.CompensationSteps) == 0 {
			t.Errorf("step %d should have compensation steps but has none", stepNum)
		}
	}

	// Verify step 1 and 13 have no compensation
	step1 := s.GetStepDefinition(1)
	if len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have no compensation steps but has %d", len(step1.CompensationSteps))
	}

	step13 := s.GetStepDefinition(13)
	if len(step13.CompensationSteps) != 0 {
		t.Errorf("step 13 should have no compensation steps but has %d", len(step13.CompensationSteps))
	}
}

// TestPayrollProcessingSagaBankingRetry tests special retry configuration for step 8
func TestPayrollProcessingSagaBankingRetry(t *testing.T) {
	s := NewPayrollProcessingSaga()
	step8 := s.GetStepDefinition(8)

	if step8 == nil {
		t.Fatal("step 8 not found")
	}
	if step8.RetryConfig == nil {
		t.Fatal("step 8 retry config is nil")
	}

	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"MaxRetries", step8.RetryConfig.MaxRetries, int32(5)},
		{"InitialBackoffMs", step8.RetryConfig.InitialBackoffMs, int32(2000)},
		{"MaxBackoffMs", step8.RetryConfig.MaxBackoffMs, int32(120000)},
		{"TimeoutSeconds", step8.TimeoutSeconds, int32(90)},
		{"IsCritical", step8.IsCritical, true},
	}

	for _, tt := range tests {
		if tt.actual != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.actual)
		}
	}
}

// TestPayrollProcessingSagaFirstAndLastSteps tests first and last step properties
func TestPayrollProcessingSagaFirstAndLastSteps(t *testing.T) {
	s := NewPayrollProcessingSaga()

	// Step 1 should have no compensation
	step1 := s.GetStepDefinition(1)
	if len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have no compensation, but has %d", len(step1.CompensationSteps))
	}

	// Step 13 (last) should have no compensation
	step13 := s.GetStepDefinition(13)
	if len(step13.CompensationSteps) != 0 {
		t.Errorf("step 13 should have no compensation, but has %d", len(step13.CompensationSteps))
	}
}

// TestPayrollProcessingSagaImplementsInterface tests interface implementation
func TestPayrollProcessingSagaImplementsInterface(t *testing.T) {
	var _ saga.SagaHandler = NewPayrollProcessingSaga()
}

// TestPayrollProcessingSagaGetStepByNumber tests step lookup by number
func TestPayrollProcessingSagaGetStepByNumber(t *testing.T) {
	s := NewPayrollProcessingSaga()

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
	if compStep.ServiceName != "leave" {
		t.Errorf("expected compensation step 102 to be leave, got %s", compStep.ServiceName)
	}
}

// TestPayrollProcessingSagaInputMapping tests input mapping configuration
func TestPayrollProcessingSagaInputMapping(t *testing.T) {
	s := NewPayrollProcessingSaga()
	step := s.GetStepDefinition(1)

	expectedMappings := map[string]string{
		"tenantID":        "$.tenantID",
		"companyID":       "$.companyID",
		"branchID":        "$.branchID",
		"payrollPeriodID": "$.input.payroll_period_id",
		"processingDate":  "$.input.processing_date",
		"employeeList":    "$.input.employee_list",
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

// ===== EMPLOYEE EXIT SAGA (H03) TESTS =====

// TestEmployeeExitSagaType tests saga type identification
func TestEmployeeExitSagaType(t *testing.T) {
	s := NewEmployeeExitSaga()
	if s.SagaType() != "SAGA-H03" {
		t.Errorf("expected SAGA-H03, got %s", s.SagaType())
	}
}

// TestEmployeeExitSagaStepCount tests total step count
func TestEmployeeExitSagaStepCount(t *testing.T) {
	s := NewEmployeeExitSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 19 // 10 forward + 9 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestEmployeeExitSagaValidation tests input validation
func TestEmployeeExitSagaValidation(t *testing.T) {
	s := NewEmployeeExitSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			"valid input",
			map[string]interface{}{
				"employee_id": "EMP-001",
				"exit_date":   "2025-03-31",
				"exit_reason": "Resignation",
				"fnf_amount":  150000.0,
			},
			false,
			"",
		},
		{
			"missing employee_id",
			map[string]interface{}{
				"exit_date":   "2025-03-31",
				"exit_reason": "Resignation",
				"fnf_amount":  150000.0,
			},
			true,
			"employee_id is required",
		},
		{
			"missing exit_date",
			map[string]interface{}{
				"employee_id": "EMP-001",
				"exit_reason": "Resignation",
				"fnf_amount":  150000.0,
			},
			true,
			"exit_date is required",
		},
		{
			"missing exit_reason",
			map[string]interface{}{
				"employee_id": "EMP-001",
				"exit_date":   "2025-03-31",
				"fnf_amount":  150000.0,
			},
			true,
			"exit_reason is required",
		},
		{
			"missing fnf_amount",
			map[string]interface{}{
				"employee_id": "EMP-001",
				"exit_date":   "2025-03-31",
				"exit_reason": "Resignation",
			},
			true,
			"fnf_amount is required",
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

// TestEmployeeExitSagaCriticalSteps tests critical step identification
func TestEmployeeExitSagaCriticalSteps(t *testing.T) {
	s := NewEmployeeExitSaga()
	criticalSteps := []int{1, 2, 3, 4, 8, 10}

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
	nonCriticalSteps := []int{5, 6, 7, 9}
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

// TestEmployeeExitSagaTimeouts tests timeout configuration
func TestEmployeeExitSagaTimeouts(t *testing.T) {
	s := NewEmployeeExitSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber < 100 { // Only forward steps
			if step.TimeoutSeconds < 30 || step.TimeoutSeconds > 45 {
				t.Errorf("step %d timeout %d is out of expected range (30-45)", step.StepNumber, step.TimeoutSeconds)
			}
		}
	}
}

// TestEmployeeExitSagaServiceNames tests service name format with hyphens
func TestEmployeeExitSagaServiceNames(t *testing.T) {
	s := NewEmployeeExitSaga()
	expectedServices := map[int]string{
		1:  "employee",
		2:  "leave",
		3:  "payroll",
		4:  "expense",
		5:  "asset",
		6:  "user",
		7:  "access",
		8:  "accounts-payable",
		9:  "general-ledger",
		10: "notification",
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

// TestEmployeeExitSagaCompensationSteps tests compensation step configuration
func TestEmployeeExitSagaCompensationSteps(t *testing.T) {
	s := NewEmployeeExitSaga()

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

// TestEmployeeExitSagaFirstAndLastSteps tests first and last step properties
func TestEmployeeExitSagaFirstAndLastSteps(t *testing.T) {
	s := NewEmployeeExitSaga()

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
	if step10.CompensationSteps[0] != 109 {
		t.Errorf("step 10 compensation should be 109, got %d", step10.CompensationSteps[0])
	}
}

// TestEmployeeExitSagaImplementsInterface tests interface implementation
func TestEmployeeExitSagaImplementsInterface(t *testing.T) {
	var _ saga.SagaHandler = NewEmployeeExitSaga()
}

// TestEmployeeExitSagaGetStepByNumber tests step lookup by number
func TestEmployeeExitSagaGetStepByNumber(t *testing.T) {
	s := NewEmployeeExitSaga()

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
	compStep := s.GetStepDefinition(101)
	if compStep == nil {
		t.Error("expected compensation step 101, got nil")
	}
	if compStep.ServiceName != "leave" {
		t.Errorf("expected compensation step 101 to be leave, got %s", compStep.ServiceName)
	}
}

// TestEmployeeExitSagaInputMapping tests input mapping configuration
func TestEmployeeExitSagaInputMapping(t *testing.T) {
	s := NewEmployeeExitSaga()
	step := s.GetStepDefinition(1)

	expectedMappings := map[string]string{
		"tenantID":   "$.tenantID",
		"companyID":  "$.companyID",
		"branchID":   "$.branchID",
		"employeeID": "$.input.employee_id",
		"exitDate":   "$.input.exit_date",
		"exitReason": "$.input.exit_reason",
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

// ===== CROSS-SAGA HR TESTS =====

// TestAllHRSagasImplementInterface tests that all sagas implement the interface
func TestAllHRSagasImplementInterface(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewPayrollProcessingSaga(),
		NewEmployeeExitSaga(),
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

// TestUniqueSagaTypesHR tests that all sagas have unique types
func TestUniqueSagaTypesHR(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewPayrollProcessingSaga(),
		NewEmployeeExitSaga(),
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

// TestHRStepSequencing tests that steps are properly sequenced
func TestHRStepSequencing(t *testing.T) {
	tests := []struct {
		name            string
		saga            saga.SagaHandler
		expectedForward int
		expectedComp    int
	}{
		{
			"PayrollProcessingSaga",
			NewPayrollProcessingSaga(),
			13,
			12,
		},
		{
			"EmployeeExitSaga",
			NewEmployeeExitSaga(),
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
