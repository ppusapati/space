// Tests for Phase 4C critical system sagas. Package name matches the
// sagas/ directory's production files (sagas — plural). B.8 fix 2026-04-19.
package sagas

import (
	"testing"

	"p9e.in/samavaya/packages/saga/sagas/finance"
	"p9e.in/samavaya/packages/saga/sagas/hr"
	"p9e.in/samavaya/packages/saga/sagas/manufacturing"
)

// ============================================================================
// SAGA-F01: Month-End Close (Special - NO Compensation)
// ============================================================================

// TestMonthEndCloseSagaType verifies saga type identification
func TestMonthEndCloseSagaType(t *testing.T) {
	s := finance.NewMonthEndCloseSaga()
	expected := "SAGA-F01"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestMonthEndCloseSagaStepCount verifies 12 forward steps with no compensation
func TestMonthEndCloseSagaStepCount(t *testing.T) {
	s := finance.NewMonthEndCloseSaga()
	steps := s.GetStepDefinitions()
	expected := 12 // 12 forward steps, NO compensation steps
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestMonthEndCloseSagaValidation verifies input validation
func TestMonthEndCloseSagaValidation(t *testing.T) {
	s := finance.NewMonthEndCloseSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"period_id":      "P001",
				"close_date":     "2024-01-31",
				"closing_month":  "202401",
				"company_id":     "COM001",
			},
			hasErr: false,
		},
		{
			name: "missing period_id",
			input: map[string]interface{}{
				"close_date":    "2024-01-31",
				"closing_month": "202401",
				"company_id":    "COM001",
			},
			hasErr: true,
		},
		{
			name: "missing close_date",
			input: map[string]interface{}{
				"period_id":     "P001",
				"closing_month": "202401",
				"company_id":    "COM001",
			},
			hasErr: true,
		},
		{
			name: "missing closing_month",
			input: map[string]interface{}{
				"period_id":  "P001",
				"close_date": "2024-01-31",
				"company_id": "COM001",
			},
			hasErr: true,
		},
		{
			name: "missing company_id",
			input: map[string]interface{}{
				"period_id":     "P001",
				"close_date":    "2024-01-31",
				"closing_month": "202401",
			},
			hasErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := s.ValidateInput(tc.input)
			if tc.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.hasErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// TestMonthEndCloseSagaCriticalSteps verifies critical steps
func TestMonthEndCloseSagaCriticalSteps(t *testing.T) {
	s := finance.NewMonthEndCloseSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 8: true, 11: true, 12: true}

	for _, step := range steps {
		stepNum := int(step.StepNumber)
		expectedCritical := criticalSteps[stepNum]
		if step.IsCritical != expectedCritical {
			t.Errorf("step %d IsCritical: expected %v, got %v", stepNum, expectedCritical, step.IsCritical)
		}
	}
}

// TestMonthEndCloseSagaNonCriticalSteps verifies non-critical steps
func TestMonthEndCloseSagaNonCriticalSteps(t *testing.T) {
	s := finance.NewMonthEndCloseSaga()
	steps := s.GetStepDefinitions()

	nonCriticalSteps := map[int]bool{9: true, 10: true}

	for _, step := range steps {
		stepNum := int(step.StepNumber)
		isExpectedNonCritical := nonCriticalSteps[stepNum]
		if isExpectedNonCritical && step.IsCritical {
			t.Errorf("step %d should be non-critical, got critical", stepNum)
		}
	}
}

// TestMonthEndCloseSagaTimeouts verifies timeout configurations
func TestMonthEndCloseSagaTimeouts(t *testing.T) {
	s := finance.NewMonthEndCloseSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.TimeoutSeconds < 30 || step.TimeoutSeconds > 60 {
			t.Errorf("step %d timeout %d out of range [30-60]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestMonthEndCloseSagaNoCompensation verifies NO compensation (special saga)
func TestMonthEndCloseSagaNoCompensation(t *testing.T) {
	s := finance.NewMonthEndCloseSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if len(step.CompensationSteps) != 0 {
			t.Errorf("step %d should have empty CompensationSteps, got %v", step.StepNumber, step.CompensationSteps)
		}
	}
}

// TestMonthEndCloseSagaServiceNames verifies service name conventions
func TestMonthEndCloseSagaServiceNames(t *testing.T) {
	s := finance.NewMonthEndCloseSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"general-ledger":       true,
		"accounts-receivable":  true,
		"accounts-payable":     true,
		"banking":              true,
		"inventory":            true,
		"cost-center":          true,
		"financial-close":      true,
		"audit":                true,
		"tax-engine":           true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestMonthEndCloseSagaInputMapping verifies step input mappings use correct JSONPath
func TestMonthEndCloseSagaInputMapping(t *testing.T) {
	s := finance.NewMonthEndCloseSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	requiredFields := []string{"tenantID", "companyID", "branchID", "periodID", "closeDate", "closingMonth"}
	for _, field := range requiredFields {
		if _, ok := step1.InputMapping[field]; !ok {
			t.Errorf("step 1 InputMapping missing required field: %s", field)
		}
	}

	// Verify inter-step dependencies use $.steps.N.result.* pattern
	step3 := s.GetStepDefinition(3)
	if step3 == nil {
		t.Fatal("step 3 not found")
	}
	if val, ok := step3.InputMapping["validationStatus"]; !ok || val != "$.steps.2.result.validation_status" {
		t.Errorf("step 3 should depend on step 2 result")
	}
}

// TestMonthEndCloseSagaSequential verifies inter-step dependencies
func TestMonthEndCloseSagaSequential(t *testing.T) {
	s := finance.NewMonthEndCloseSaga()

	// Step 3 depends on step 2
	step3 := s.GetStepDefinition(3)
	if step3 == nil {
		t.Fatal("step 3 not found")
	}
	hasDependency := false
	for _, v := range step3.InputMapping {
		if v == "$.steps.2.result.validation_status" {
			hasDependency = true
			break
		}
	}
	if !hasDependency {
		t.Error("step 3 should depend on step 2")
	}

	// Step 4 depends on step 3
	step4 := s.GetStepDefinition(4)
	if step4 == nil {
		t.Fatal("step 4 not found")
	}
	hasDependency = false
	for _, v := range step4.InputMapping {
		if v == "$.steps.3.result.accrual_amount" {
			hasDependency = true
			break
		}
	}
	if !hasDependency {
		t.Error("step 4 should depend on step 3")
	}
}

// ============================================================================
// SAGA-H01: Payroll Processing
// ============================================================================

// TestPayrollProcessingSagaType verifies saga type identification
func TestPayrollProcessingSagaType(t *testing.T) {
	s := hr.NewPayrollProcessingSaga()
	expected := "SAGA-H01"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestPayrollProcessingSagaStepCount verifies 25 steps (13 forward + 12 compensation)
func TestPayrollProcessingSagaStepCount(t *testing.T) {
	s := hr.NewPayrollProcessingSaga()
	steps := s.GetStepDefinitions()
	expected := 25 // 13 forward + 12 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestPayrollProcessingSagaValidation verifies input validation
func TestPayrollProcessingSagaValidation(t *testing.T) {
	s := hr.NewPayrollProcessingSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"payroll_run_id": "PR001",
				"payroll_period": "2024-01",
				"payroll_date":   "2024-01-31",
				"company_id":     "COM001",
			},
			hasErr: false,
		},
		{
			name: "missing payroll_run_id",
			input: map[string]interface{}{
				"payroll_period": "2024-01",
				"payroll_date":   "2024-01-31",
				"company_id":     "COM001",
			},
			hasErr: true,
		},
		{
			name: "missing payroll_period",
			input: map[string]interface{}{
				"payroll_run_id": "PR001",
				"payroll_date":   "2024-01-31",
				"company_id":     "COM001",
			},
			hasErr: true,
		},
		{
			name: "missing payroll_date",
			input: map[string]interface{}{
				"payroll_run_id": "PR001",
				"payroll_period": "2024-01",
				"company_id":     "COM001",
			},
			hasErr: true,
		},
		{
			name: "missing company_id",
			input: map[string]interface{}{
				"payroll_run_id": "PR001",
				"payroll_period": "2024-01",
				"payroll_date":   "2024-01-31",
			},
			hasErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := s.ValidateInput(tc.input)
			if tc.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.hasErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// TestPayrollProcessingSagaCriticalSteps verifies critical steps
func TestPayrollProcessingSagaCriticalSteps(t *testing.T) {
	s := hr.NewPayrollProcessingSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 10: true, 12: true, 13: true}

	for _, step := range steps {
		stepNum := int(step.StepNumber)
		if stepNum > 13 {
			continue // Skip compensation steps
		}
		expectedCritical := criticalSteps[stepNum]
		if step.IsCritical != expectedCritical {
			t.Errorf("step %d IsCritical: expected %v, got %v", stepNum, expectedCritical, step.IsCritical)
		}
	}
}

// TestPayrollProcessingSagaNonCriticalSteps verifies non-critical steps
func TestPayrollProcessingSagaNonCriticalSteps(t *testing.T) {
	s := hr.NewPayrollProcessingSaga()

	nonCriticalSteps := map[int]bool{7: true, 8: true, 9: true, 11: true}

	for stepNum := range nonCriticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.IsCritical {
			t.Errorf("step %d should be non-critical, got critical", stepNum)
		}
	}
}

// TestPayrollProcessingSagaTimeouts verifies timeout configurations
func TestPayrollProcessingSagaTimeouts(t *testing.T) {
	s := hr.NewPayrollProcessingSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 13 {
			continue // Skip compensation steps
		}
		if step.TimeoutSeconds < 30 || step.TimeoutSeconds > 45 {
			t.Errorf("step %d timeout %d out of range [30-45]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestPayrollProcessingSagaCompensationSteps verifies compensation mappings
func TestPayrollProcessingSagaCompensationSteps(t *testing.T) {
	s := hr.NewPayrollProcessingSaga()

	// Steps 2-12 should have compensation steps in range 102-112
	for stepNum := 2; stepNum <= 12; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}

		hasCompensation := len(step.CompensationSteps) > 0
		if !hasCompensation {
			t.Errorf("step %d should have compensation steps", stepNum)
		}

		for _, compStep := range step.CompensationSteps {
			if compStep < 102 || compStep > 112 {
				t.Errorf("step %d compensation %d out of range [102-112]", stepNum, compStep)
			}
		}
	}
}

// TestPayrollProcessingSagaServiceNames verifies service name conventions
func TestPayrollProcessingSagaServiceNames(t *testing.T) {
	s := hr.NewPayrollProcessingSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"payroll":            true,
		"salary-structure":   true,
		"attendance":         true,
		"general-ledger":     true,
		"banking":            true,
		"accounts-payable":   true,
	}

	for _, step := range steps {
		if step.StepNumber > 13 {
			continue // Skip compensation steps
		}
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestPayrollProcessingSagaInputMapping verifies field mappings in step 1
func TestPayrollProcessingSagaInputMapping(t *testing.T) {
	s := hr.NewPayrollProcessingSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	requiredFields := []string{"tenantID", "companyID", "branchID", "payrollRunID", "payrollPeriod", "payrollDate"}
	for _, field := range requiredFields {
		if _, ok := step1.InputMapping[field]; !ok {
			t.Errorf("step 1 InputMapping missing required field: %s", field)
		}
	}
}

// TestPayrollProcessingSagaFirstLastSteps verifies terminal steps
func TestPayrollProcessingSagaFirstLastSteps(t *testing.T) {
	s := hr.NewPayrollProcessingSaga()

	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}
	if len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have empty CompensationSteps, got %d", len(step1.CompensationSteps))
	}

	step13 := s.GetStepDefinition(13)
	if step13 == nil {
		t.Fatal("step 13 not found")
	}
	if len(step13.CompensationSteps) != 0 {
		t.Errorf("step 13 should have empty CompensationSteps, got %d", len(step13.CompensationSteps))
	}
}

// TestPayrollProcessingSagaRetryConfig verifies standard retry configuration
func TestPayrollProcessingSagaRetryConfig(t *testing.T) {
	s := hr.NewPayrollProcessingSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 13 {
			continue // Skip compensation steps
		}
		if step.RetryConfig == nil {
			t.Errorf("step %d missing RetryConfig", step.StepNumber)
			continue
		}

		if step.RetryConfig.MaxRetries != 3 {
			t.Errorf("step %d MaxRetries: expected 3, got %d", step.StepNumber, step.RetryConfig.MaxRetries)
		}
		if step.RetryConfig.InitialBackoffMs != 1000 {
			t.Errorf("step %d InitialBackoffMs: expected 1000, got %d", step.StepNumber, step.RetryConfig.InitialBackoffMs)
		}
		if step.RetryConfig.MaxBackoffMs != 30000 {
			t.Errorf("step %d MaxBackoffMs: expected 30000, got %d", step.StepNumber, step.RetryConfig.MaxBackoffMs)
		}
	}
}

// TestPayrollProcessingSagaInterfaceImplementation verifies interface implementation
func TestPayrollProcessingSagaInterfaceImplementation(t *testing.T) {
	var _ SagaHandler = (*hr.PayrollProcessingSaga)(nil)
	s := hr.NewPayrollProcessingSaga()

	if s.SagaType() == "" {
		t.Error("SagaType() should not return empty string")
	}
	if len(s.GetStepDefinitions()) == 0 {
		t.Error("GetStepDefinitions() should not return empty slice")
	}
}

// ============================================================================
// SAGA-H03: Employee Exit
// ============================================================================

// TestEmployeeExitSagaType verifies saga type identification
func TestEmployeeExitSagaType(t *testing.T) {
	s := hr.NewEmployeeExitSaga()
	expected := "SAGA-H03"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestEmployeeExitSagaStepCount verifies 19 steps (10 forward + 9 compensation)
func TestEmployeeExitSagaStepCount(t *testing.T) {
	s := hr.NewEmployeeExitSaga()
	steps := s.GetStepDefinitions()
	expected := 19 // 10 forward + 9 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestEmployeeExitSagaValidation verifies input validation
func TestEmployeeExitSagaValidation(t *testing.T) {
	s := hr.NewEmployeeExitSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"employee_id":          "EMP001",
				"exit_date":            "2024-02-15",
				"full_and_final_date":  "2024-02-28",
				"reason":               "voluntary",
			},
			hasErr: false,
		},
		{
			name: "missing employee_id",
			input: map[string]interface{}{
				"exit_date":            "2024-02-15",
				"full_and_final_date":  "2024-02-28",
				"reason":               "voluntary",
			},
			hasErr: true,
		},
		{
			name: "missing exit_date",
			input: map[string]interface{}{
				"employee_id":         "EMP001",
				"full_and_final_date": "2024-02-28",
				"reason":              "voluntary",
			},
			hasErr: true,
		},
		{
			name: "missing full_and_final_date",
			input: map[string]interface{}{
				"employee_id": "EMP001",
				"exit_date":   "2024-02-15",
				"reason":      "voluntary",
			},
			hasErr: true,
		},
		{
			name: "missing reason",
			input: map[string]interface{}{
				"employee_id":         "EMP001",
				"exit_date":           "2024-02-15",
				"full_and_final_date": "2024-02-28",
			},
			hasErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := s.ValidateInput(tc.input)
			if tc.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.hasErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// TestEmployeeExitSagaCriticalSteps verifies critical steps
func TestEmployeeExitSagaCriticalSteps(t *testing.T) {
	s := hr.NewEmployeeExitSaga()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 8: true, 9: true, 10: true}

	for stepNum := range criticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be critical, got non-critical", stepNum)
		}
	}
}

// TestEmployeeExitSagaNonCriticalSteps verifies non-critical steps
func TestEmployeeExitSagaNonCriticalSteps(t *testing.T) {
	s := hr.NewEmployeeExitSaga()

	nonCriticalSteps := map[int]bool{6: true, 7: true}

	for stepNum := range nonCriticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.IsCritical {
			t.Errorf("step %d should be non-critical, got critical", stepNum)
		}
	}
}

// TestEmployeeExitSagaTimeouts verifies timeout configurations
func TestEmployeeExitSagaTimeouts(t *testing.T) {
	s := hr.NewEmployeeExitSaga()

	for stepNum := 1; stepNum <= 10; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.TimeoutSeconds < 30 || step.TimeoutSeconds > 45 {
			t.Errorf("step %d timeout %d out of range [30-45]", stepNum, step.TimeoutSeconds)
		}
	}
}

// TestEmployeeExitSagaCompensationSteps verifies compensation mappings
func TestEmployeeExitSagaCompensationSteps(t *testing.T) {
	s := hr.NewEmployeeExitSaga()

	// Steps 2-9 should have compensation steps in range 102-109
	for stepNum := 2; stepNum <= 9; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}

		hasCompensation := len(step.CompensationSteps) > 0
		if !hasCompensation {
			t.Errorf("step %d should have compensation steps", stepNum)
		}

		for _, compStep := range step.CompensationSteps {
			if compStep < 102 || compStep > 109 {
				t.Errorf("step %d compensation %d out of range [102-109]", stepNum, compStep)
			}
		}
	}
}

// TestEmployeeExitSagaServiceNames verifies service name conventions
func TestEmployeeExitSagaServiceNames(t *testing.T) {
	s := hr.NewEmployeeExitSaga()

	expectedServices := map[string]bool{
		"employee":          true,
		"salary-structure":  true,
		"payroll":           true,
		"asset":             true,
		"access":            true,
		"general-ledger":    true,
		"exit":              true,
		"notification":      true,
	}

	for stepNum := 1; stepNum <= 10; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", stepNum, step.ServiceName)
		}
	}
}

// TestEmployeeExitSagaInputMapping verifies field mappings
func TestEmployeeExitSagaInputMapping(t *testing.T) {
	s := hr.NewEmployeeExitSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	requiredFields := []string{"tenantID", "companyID", "employeeID", "exitDate", "reason"}
	for _, field := range requiredFields {
		if _, ok := step1.InputMapping[field]; !ok {
			t.Errorf("step 1 InputMapping missing required field: %s", field)
		}
	}
}

// TestEmployeeExitSagaFirstLastSteps verifies terminal steps
func TestEmployeeExitSagaFirstLastSteps(t *testing.T) {
	s := hr.NewEmployeeExitSaga()

	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}
	if len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have empty CompensationSteps, got %d", len(step1.CompensationSteps))
	}

	step10 := s.GetStepDefinition(10)
	if step10 == nil {
		t.Fatal("step 10 not found")
	}
	if len(step10.CompensationSteps) != 0 {
		t.Errorf("step 10 should have empty CompensationSteps, got %d", len(step10.CompensationSteps))
	}
}

// TestEmployeeExitSagaRetryConfig verifies standard retry configuration
func TestEmployeeExitSagaRetryConfig(t *testing.T) {
	s := hr.NewEmployeeExitSaga()

	for stepNum := 1; stepNum <= 10; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.RetryConfig == nil {
			t.Errorf("step %d missing RetryConfig", stepNum)
			continue
		}

		if step.RetryConfig.MaxRetries != 3 {
			t.Errorf("step %d MaxRetries: expected 3, got %d", stepNum, step.RetryConfig.MaxRetries)
		}
		if step.RetryConfig.InitialBackoffMs != 1000 {
			t.Errorf("step %d InitialBackoffMs: expected 1000, got %d", stepNum, step.RetryConfig.InitialBackoffMs)
		}
		if step.RetryConfig.MaxBackoffMs != 30000 {
			t.Errorf("step %d MaxBackoffMs: expected 30000, got %d", stepNum, step.RetryConfig.MaxBackoffMs)
		}
	}
}

// TestEmployeeExitSagaInterfaceImplementation verifies interface implementation
func TestEmployeeExitSagaInterfaceImplementation(t *testing.T) {
	var _ SagaHandler = (*hr.EmployeeExitSaga)(nil)
	s := hr.NewEmployeeExitSaga()

	if s.SagaType() == "" {
		t.Error("SagaType() should not return empty string")
	}
	if len(s.GetStepDefinitions()) == 0 {
		t.Error("GetStepDefinitions() should not return empty slice")
	}
}

// ============================================================================
// SAGA-M01: Production Order Execution
// ============================================================================

// TestProductionOrderExecutionSagaType verifies saga type identification
func TestProductionOrderExecutionSagaType(t *testing.T) {
	s := manufacturing.NewProductionOrderSaga()
	expected := "SAGA-M01"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestProductionOrderExecutionSagaStepCount verifies 19 steps (10 forward + 9 compensation)
func TestProductionOrderExecutionSagaStepCount(t *testing.T) {
	s := manufacturing.NewProductionOrderSaga()
	steps := s.GetStepDefinitions()
	expected := 19 // 10 forward + 9 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestProductionOrderExecutionSagaValidation verifies input validation
func TestProductionOrderExecutionSagaValidation(t *testing.T) {
	s := manufacturing.NewProductionOrderSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"production_order_id": "PO001",
				"start_date":          "2024-02-15",
				"scheduled_date":      "2024-02-28",
				"product_id":          "PROD001",
			},
			hasErr: false,
		},
		{
			name: "missing production_order_id",
			input: map[string]interface{}{
				"start_date":     "2024-02-15",
				"scheduled_date": "2024-02-28",
				"product_id":     "PROD001",
			},
			hasErr: true,
		},
		{
			name: "missing start_date",
			input: map[string]interface{}{
				"production_order_id": "PO001",
				"scheduled_date":      "2024-02-28",
				"product_id":          "PROD001",
			},
			hasErr: true,
		},
		{
			name: "missing scheduled_date",
			input: map[string]interface{}{
				"production_order_id": "PO001",
				"start_date":          "2024-02-15",
				"product_id":          "PROD001",
			},
			hasErr: true,
		},
		{
			name: "missing product_id",
			input: map[string]interface{}{
				"production_order_id": "PO001",
				"start_date":          "2024-02-15",
				"scheduled_date":      "2024-02-28",
			},
			hasErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := s.ValidateInput(tc.input)
			if tc.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.hasErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// TestProductionOrderExecutionSagaCriticalSteps verifies critical steps
func TestProductionOrderExecutionSagaCriticalSteps(t *testing.T) {
	s := manufacturing.NewProductionOrderSaga()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 8: true, 10: true}

	for stepNum := range criticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be critical, got non-critical", stepNum)
		}
	}
}

// TestProductionOrderExecutionSagaNonCriticalSteps verifies non-critical steps
func TestProductionOrderExecutionSagaNonCriticalSteps(t *testing.T) {
	s := manufacturing.NewProductionOrderSaga()

	nonCriticalSteps := map[int]bool{6: true, 7: true, 9: true}

	for stepNum := range nonCriticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.IsCritical {
			t.Errorf("step %d should be non-critical, got critical", stepNum)
		}
	}
}

// TestProductionOrderExecutionSagaTimeouts verifies timeout configurations
func TestProductionOrderExecutionSagaTimeouts(t *testing.T) {
	s := manufacturing.NewProductionOrderSaga()

	for stepNum := 1; stepNum <= 10; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.TimeoutSeconds < 30 || step.TimeoutSeconds > 45 {
			t.Errorf("step %d timeout %d out of range [30-45]", stepNum, step.TimeoutSeconds)
		}
	}
}

// TestProductionOrderExecutionSagaCompensationSteps verifies compensation mappings
func TestProductionOrderExecutionSagaCompensationSteps(t *testing.T) {
	s := manufacturing.NewProductionOrderSaga()

	// Steps 2-9 should have compensation steps in range 102-109
	for stepNum := 2; stepNum <= 9; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}

		hasCompensation := len(step.CompensationSteps) > 0
		if !hasCompensation {
			t.Errorf("step %d should have compensation steps", stepNum)
		}

		for _, compStep := range step.CompensationSteps {
			if compStep < 102 || compStep > 109 {
				t.Errorf("step %d compensation %d out of range [102-109]", stepNum, compStep)
			}
		}
	}
}

// TestProductionOrderExecutionSagaServiceNames verifies service name conventions
func TestProductionOrderExecutionSagaServiceNames(t *testing.T) {
	s := manufacturing.NewProductionOrderSaga()

	expectedServices := map[string]bool{
		"production-order":     true,
		"production-planning":  true,
		"shop-floor":           true,
		"job-card":             true,
		"work-center":          true,
		"routing":              true,
	}

	for stepNum := 1; stepNum <= 10; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", stepNum, step.ServiceName)
		}
	}
}

// TestProductionOrderExecutionSagaInputMapping verifies field mappings
func TestProductionOrderExecutionSagaInputMapping(t *testing.T) {
	s := manufacturing.NewProductionOrderSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	requiredFields := []string{"tenantID", "companyID", "branchID", "productionOrderID", "productID", "startDate", "scheduledDate"}
	for _, field := range requiredFields {
		if _, ok := step1.InputMapping[field]; !ok {
			t.Errorf("step 1 InputMapping missing required field: %s", field)
		}
	}
}

// TestProductionOrderExecutionSagaFirstLastSteps verifies terminal steps
func TestProductionOrderExecutionSagaFirstLastSteps(t *testing.T) {
	s := manufacturing.NewProductionOrderSaga()

	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}
	if len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have empty CompensationSteps, got %d", len(step1.CompensationSteps))
	}

	step10 := s.GetStepDefinition(10)
	if step10 == nil {
		t.Fatal("step 10 not found")
	}
	if len(step10.CompensationSteps) != 0 {
		t.Errorf("step 10 should have empty CompensationSteps, got %d", len(step10.CompensationSteps))
	}
}

// TestProductionOrderExecutionSagaRetryConfig verifies standard retry configuration
func TestProductionOrderExecutionSagaRetryConfig(t *testing.T) {
	s := manufacturing.NewProductionOrderSaga()

	for stepNum := 1; stepNum <= 10; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.RetryConfig == nil {
			t.Errorf("step %d missing RetryConfig", stepNum)
			continue
		}

		if step.RetryConfig.MaxRetries != 3 {
			t.Errorf("step %d MaxRetries: expected 3, got %d", stepNum, step.RetryConfig.MaxRetries)
		}
		if step.RetryConfig.InitialBackoffMs != 1000 {
			t.Errorf("step %d InitialBackoffMs: expected 1000, got %d", stepNum, step.RetryConfig.InitialBackoffMs)
		}
		if step.RetryConfig.MaxBackoffMs != 30000 {
			t.Errorf("step %d MaxBackoffMs: expected 30000, got %d", stepNum, step.RetryConfig.MaxBackoffMs)
		}
	}
}

// TestProductionOrderExecutionSagaInterfaceImplementation verifies interface implementation
func TestProductionOrderExecutionSagaInterfaceImplementation(t *testing.T) {
	var _ SagaHandler = (*manufacturing.ProductionOrderSaga)(nil)
	s := manufacturing.NewProductionOrderSaga()

	if s.SagaType() == "" {
		t.Error("SagaType() should not return empty string")
	}
	if len(s.GetStepDefinitions()) == 0 {
		t.Error("GetStepDefinitions() should not return empty slice")
	}
}

// ============================================================================
// SAGA-M02: Subcontracting
// ============================================================================

// TestSubcontractingSagaType verifies saga type identification
func TestSubcontractingSagaType(t *testing.T) {
	s := manufacturing.NewSubcontractingSaga()
	expected := "SAGA-M02"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestSubcontractingSagaStepCount verifies 19 steps (10 forward + 9 compensation)
func TestSubcontractingSagaStepCount(t *testing.T) {
	s := manufacturing.NewSubcontractingSaga()
	steps := s.GetStepDefinitions()
	expected := 19 // 10 forward + 9 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestSubcontractingSagaValidation verifies input validation
func TestSubcontractingSagaValidation(t *testing.T) {
	s := manufacturing.NewSubcontractingSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"po_id":                 "PO001",
				"subcontractor_id":      "SUB001",
				"material_issue_date":   "2024-02-15",
				"product_id":            "PROD001",
			},
			hasErr: false,
		},
		{
			name: "missing po_id",
			input: map[string]interface{}{
				"subcontractor_id":    "SUB001",
				"material_issue_date": "2024-02-15",
				"product_id":          "PROD001",
			},
			hasErr: true,
		},
		{
			name: "missing subcontractor_id",
			input: map[string]interface{}{
				"po_id":              "PO001",
				"material_issue_date": "2024-02-15",
				"product_id":         "PROD001",
			},
			hasErr: true,
		},
		{
			name: "missing material_issue_date",
			input: map[string]interface{}{
				"po_id":            "PO001",
				"subcontractor_id": "SUB001",
				"product_id":       "PROD001",
			},
			hasErr: true,
		},
		{
			name: "missing product_id",
			input: map[string]interface{}{
				"po_id":              "PO001",
				"subcontractor_id":   "SUB001",
				"material_issue_date": "2024-02-15",
			},
			hasErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := s.ValidateInput(tc.input)
			if tc.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.hasErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// TestSubcontractingSagaCriticalSteps verifies critical steps
func TestSubcontractingSagaCriticalSteps(t *testing.T) {
	s := manufacturing.NewSubcontractingSaga()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 8: true, 9: true, 10: true}

	for stepNum := range criticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be critical, got non-critical", stepNum)
		}
	}
}

// TestSubcontractingSagaNonCriticalSteps verifies non-critical steps
func TestSubcontractingSagaNonCriticalSteps(t *testing.T) {
	s := manufacturing.NewSubcontractingSaga()

	nonCriticalSteps := map[int]bool{6: true, 7: true}

	for stepNum := range nonCriticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.IsCritical {
			t.Errorf("step %d should be non-critical, got critical", stepNum)
		}
	}
}

// TestSubcontractingSagaTimeouts verifies timeout configurations
func TestSubcontractingSagaTimeouts(t *testing.T) {
	s := manufacturing.NewSubcontractingSaga()

	for stepNum := 1; stepNum <= 10; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.TimeoutSeconds < 30 || step.TimeoutSeconds > 45 {
			t.Errorf("step %d timeout %d out of range [30-45]", stepNum, step.TimeoutSeconds)
		}
	}
}

// TestSubcontractingSagaCompensationSteps verifies compensation mappings
func TestSubcontractingSagaCompensationSteps(t *testing.T) {
	s := manufacturing.NewSubcontractingSaga()

	// Steps 2-9 should have compensation steps in range 102-109
	for stepNum := 2; stepNum <= 9; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}

		hasCompensation := len(step.CompensationSteps) > 0
		if !hasCompensation {
			t.Errorf("step %d should have compensation steps", stepNum)
		}

		for _, compStep := range step.CompensationSteps {
			if compStep < 102 || compStep > 109 {
				t.Errorf("step %d compensation %d out of range [102-109]", stepNum, compStep)
			}
		}
	}
}

// TestSubcontractingSagaServiceNames verifies service name conventions
func TestSubcontractingSagaServiceNames(t *testing.T) {
	s := manufacturing.NewSubcontractingSaga()

	expectedServices := map[string]bool{
		"subcontracting":        true,
		"procurement":           true,
		"inventory-core":        true,
		"quality-production":    true,
		"accounts-payable":      true,
	}

	for stepNum := 1; stepNum <= 10; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", stepNum, step.ServiceName)
		}
	}
}

// TestSubcontractingSagaInputMapping verifies field mappings
func TestSubcontractingSagaInputMapping(t *testing.T) {
	s := manufacturing.NewSubcontractingSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	requiredFields := []string{"tenantID", "companyID", "poID", "subcontractorID", "materialIssueDate", "productID"}
	for _, field := range requiredFields {
		if _, ok := step1.InputMapping[field]; !ok {
			t.Errorf("step 1 InputMapping missing required field: %s", field)
		}
	}
}

// TestSubcontractingSagaFirstLastSteps verifies terminal steps
func TestSubcontractingSagaFirstLastSteps(t *testing.T) {
	s := manufacturing.NewSubcontractingSaga()

	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}
	if len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have empty CompensationSteps, got %d", len(step1.CompensationSteps))
	}

	step10 := s.GetStepDefinition(10)
	if step10 == nil {
		t.Fatal("step 10 not found")
	}
	if len(step10.CompensationSteps) != 0 {
		t.Errorf("step 10 should have empty CompensationSteps, got %d", len(step10.CompensationSteps))
	}
}

// TestSubcontractingSagaRetryConfig verifies standard retry configuration
func TestSubcontractingSagaRetryConfig(t *testing.T) {
	s := manufacturing.NewSubcontractingSaga()

	for stepNum := 1; stepNum <= 10; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.RetryConfig == nil {
			t.Errorf("step %d missing RetryConfig", stepNum)
			continue
		}

		if step.RetryConfig.MaxRetries != 3 {
			t.Errorf("step %d MaxRetries: expected 3, got %d", stepNum, step.RetryConfig.MaxRetries)
		}
		if step.RetryConfig.InitialBackoffMs != 1000 {
			t.Errorf("step %d InitialBackoffMs: expected 1000, got %d", stepNum, step.RetryConfig.InitialBackoffMs)
		}
		if step.RetryConfig.MaxBackoffMs != 30000 {
			t.Errorf("step %d MaxBackoffMs: expected 30000, got %d", stepNum, step.RetryConfig.MaxBackoffMs)
		}
	}
}

// TestSubcontractingSagaInterfaceImplementation verifies interface implementation
func TestSubcontractingSagaInterfaceImplementation(t *testing.T) {
	var _ SagaHandler = (*manufacturing.SubcontractingSaga)(nil)
	s := manufacturing.NewSubcontractingSaga()

	if s.SagaType() == "" {
		t.Error("SagaType() should not return empty string")
	}
	if len(s.GetStepDefinitions()) == 0 {
		t.Error("GetStepDefinitions() should not return empty slice")
	}
}
