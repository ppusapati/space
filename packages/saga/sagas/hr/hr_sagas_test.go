// Package hr provides tests for HR & Payroll module saga workflows
package hr

import (
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// ===== SAGA-H02: Employee Onboarding Saga Tests =====

func TestEmployeeOnboardingSagaType(t *testing.T) {
	s := NewEmployeeOnboardingSaga()
	if s.SagaType() != "SAGA-H02" {
		t.Errorf("expected SAGA-H02, got %s", s.SagaType())
	}
}

func TestEmployeeOnboardingSagaStepCount(t *testing.T) {
	s := NewEmployeeOnboardingSaga()
	steps := s.GetStepDefinitions()
	// Should have 10 forward + 8 compensation = 18 total
	if len(steps) != 18 {
		t.Errorf("expected 18 steps, got %d", len(steps))
	}
}

func TestEmployeeOnboardingSagaInputValidation(t *testing.T) {
	s := NewEmployeeOnboardingSaga()
	tests := []struct {
		name      string
		input     map[string]interface{}
		wantError bool
	}{
		{
			"valid input",
			map[string]interface{}{
				"employee_id": "EMP-001",
				"first_name":  "John",
				"last_name":   "Doe",
				"email":       "john@example.com",
				"designation": "Engineer",
				"department":  "Engineering",
			},
			false,
		},
		{
			"missing employee_id",
			map[string]interface{}{
				"first_name":  "John",
				"last_name":   "Doe",
				"email":       "john@example.com",
				"designation": "Engineer",
				"department":  "Engineering",
			},
			true,
		},
		{
			"missing first_name",
			map[string]interface{}{
				"employee_id": "EMP-001",
				"last_name":   "Doe",
				"email":       "john@example.com",
				"designation": "Engineer",
				"department":  "Engineering",
			},
			true,
		},
		{
			"missing last_name",
			map[string]interface{}{
				"employee_id": "EMP-001",
				"first_name":  "John",
				"email":       "john@example.com",
				"designation": "Engineer",
				"department":  "Engineering",
			},
			true,
		},
		{
			"missing email",
			map[string]interface{}{
				"employee_id": "EMP-001",
				"first_name":  "John",
				"last_name":   "Doe",
				"designation": "Engineer",
				"department":  "Engineering",
			},
			true,
		},
		{
			"missing designation",
			map[string]interface{}{
				"employee_id": "EMP-001",
				"first_name":  "John",
				"last_name":   "Doe",
				"email":       "john@example.com",
				"department":  "Engineering",
			},
			true,
		},
		{
			"missing department",
			map[string]interface{}{
				"employee_id": "EMP-001",
				"first_name":  "John",
				"last_name":   "Doe",
				"email":       "john@example.com",
				"designation": "Engineer",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateInput() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestEmployeeOnboardingSagaGetStepDefinition(t *testing.T) {
	s := NewEmployeeOnboardingSaga()
	tests := []struct {
		stepNum    int
		wantExists bool
	}{
		{1, true},
		{5, true},
		{10, true},
		{101, true},
		{108, true},
		{0, false},
		{11, false},
		{109, false},
	}

	for _, tt := range tests {
		t.Run("step_"+string(rune(tt.stepNum)), func(t *testing.T) {
			step := s.GetStepDefinition(tt.stepNum)
			if (step != nil) != tt.wantExists {
				t.Errorf("GetStepDefinition(%d) exists = %v, want %v", tt.stepNum, step != nil, tt.wantExists)
			}
		})
	}
}

func TestEmployeeOnboardingSagaCompensationSteps(t *testing.T) {
	s := NewEmployeeOnboardingSaga()

	// Step 1 should have no compensation
	step1 := s.GetStepDefinition(1)
	if step1 != nil && len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have no compensation steps, got %d", len(step1.CompensationSteps))
	}

	// Step 10 should have no compensation
	step10 := s.GetStepDefinition(10)
	if step10 != nil && len(step10.CompensationSteps) != 0 {
		t.Errorf("step 10 should have no compensation steps, got %d", len(step10.CompensationSteps))
	}

	// Step 2 should have compensation
	step2 := s.GetStepDefinition(2)
	if step2 != nil && len(step2.CompensationSteps) == 0 {
		t.Errorf("step 2 should have compensation steps")
	}
}

func TestEmployeeOnboardingSagaCriticalSteps(t *testing.T) {
	s := NewEmployeeOnboardingSaga()
	expectedCritical := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 10: true}

	for stepNum, shouldBeCritical := range expectedCritical {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.IsCritical != shouldBeCritical {
			t.Errorf("step %d critical status mismatch: expected %v, got %v", stepNum, shouldBeCritical, step.IsCritical)
		}
	}
}

func TestEmployeeOnboardingSagaServiceNames(t *testing.T) {
	s := NewEmployeeOnboardingSaga()
	expectedServices := map[int]string{
		1: "employee",
		2: "user",
		3: "access",
		4: "salary-structure",
		5: "payroll",
	}

	for stepNum, expectedService := range expectedServices {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.ServiceName != expectedService {
			t.Errorf("step %d service mismatch: expected %s, got %s", stepNum, expectedService, step.ServiceName)
		}
	}
}

func TestEmployeeOnboardingSagaRetryConfig(t *testing.T) {
	s := NewEmployeeOnboardingSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	if step1.RetryConfig == nil {
		t.Fatal("step 1 retry config is nil")
	}

	if step1.RetryConfig.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", step1.RetryConfig.MaxRetries)
	}

	if step1.RetryConfig.InitialBackoffMs != 1000 {
		t.Errorf("expected InitialBackoffMs=1000, got %d", step1.RetryConfig.InitialBackoffMs)
	}
}

// ===== SAGA-H04: Expense Reimbursement Saga Tests =====

func TestExpenseReimbursementSagaType(t *testing.T) {
	s := NewExpenseReimbursementSaga()
	if s.SagaType() != "SAGA-H04" {
		t.Errorf("expected SAGA-H04, got %s", s.SagaType())
	}
}

func TestExpenseReimbursementSagaStepCount(t *testing.T) {
	s := NewExpenseReimbursementSaga()
	steps := s.GetStepDefinitions()
	// Should have 10 forward + 8 compensation = 18 total
	if len(steps) != 18 {
		t.Errorf("expected 18 steps, got %d", len(steps))
	}
}

func TestExpenseReimbursementSagaInputValidation(t *testing.T) {
	s := NewExpenseReimbursementSaga()
	tests := []struct {
		name      string
		input     map[string]interface{}
		wantError bool
	}{
		{
			"valid input",
			map[string]interface{}{
				"expense_id":       "EXP-001",
				"employee_id":      "EMP-001",
				"amount":           5000.00,
				"cost_center_id":   "CC-001",
				"submission_date":  "2026-02-14",
			},
			false,
		},
		{
			"missing expense_id",
			map[string]interface{}{
				"employee_id":     "EMP-001",
				"amount":          5000.00,
				"cost_center_id":  "CC-001",
				"submission_date": "2026-02-14",
			},
			true,
		},
		{
			"missing employee_id",
			map[string]interface{}{
				"expense_id":      "EXP-001",
				"amount":          5000.00,
				"cost_center_id":  "CC-001",
				"submission_date": "2026-02-14",
			},
			true,
		},
		{
			"missing amount",
			map[string]interface{}{
				"expense_id":      "EXP-001",
				"employee_id":     "EMP-001",
				"cost_center_id":  "CC-001",
				"submission_date": "2026-02-14",
			},
			true,
		},
		{
			"missing cost_center_id",
			map[string]interface{}{
				"expense_id":      "EXP-001",
				"employee_id":     "EMP-001",
				"amount":          5000.00,
				"submission_date": "2026-02-14",
			},
			true,
		},
		{
			"missing submission_date",
			map[string]interface{}{
				"expense_id":     "EXP-001",
				"employee_id":    "EMP-001",
				"amount":         5000.00,
				"cost_center_id": "CC-001",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateInput() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestExpenseReimbursementSagaGetStepDefinition(t *testing.T) {
	s := NewExpenseReimbursementSaga()
	step := s.GetStepDefinition(1)
	if step == nil {
		t.Errorf("step 1 not found")
	}
	if step.ServiceName != "expense" {
		t.Errorf("step 1 service mismatch: expected 'expense', got %s", step.ServiceName)
	}
}

func TestExpenseReimbursementSagaCriticalSteps(t *testing.T) {
	s := NewExpenseReimbursementSaga()
	expectedCritical := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 10: true}

	for stepNum, shouldBeCritical := range expectedCritical {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.IsCritical != shouldBeCritical {
			t.Errorf("step %d critical status mismatch: expected %v, got %v", stepNum, shouldBeCritical, step.IsCritical)
		}
	}
}

func TestExpenseReimbursementSagaCompensationSteps(t *testing.T) {
	s := NewExpenseReimbursementSaga()

	step1 := s.GetStepDefinition(1)
	if step1 != nil && len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have no compensation steps, got %d", len(step1.CompensationSteps))
	}

	step10 := s.GetStepDefinition(10)
	if step10 != nil && len(step10.CompensationSteps) != 0 {
		t.Errorf("step 10 should have no compensation steps, got %d", len(step10.CompensationSteps))
	}
}

// ===== SAGA-H06: Appraisal & Salary Revision Saga Tests =====

func TestAppraisalSalaryRevisionSagaType(t *testing.T) {
	s := NewAppraisalSalaryRevisionSaga()
	if s.SagaType() != "SAGA-H06" {
		t.Errorf("expected SAGA-H06, got %s", s.SagaType())
	}
}

func TestAppraisalSalaryRevisionSagaStepCount(t *testing.T) {
	s := NewAppraisalSalaryRevisionSaga()
	steps := s.GetStepDefinitions()
	// Should have 9 forward + 7 compensation = 16 total
	if len(steps) != 16 {
		t.Errorf("expected 16 steps, got %d", len(steps))
	}
}

func TestAppraisalSalaryRevisionSagaInputValidation(t *testing.T) {
	s := NewAppraisalSalaryRevisionSaga()
	tests := []struct {
		name      string
		input     map[string]interface{}
		wantError bool
	}{
		{
			"valid input",
			map[string]interface{}{
				"appraisal_id":      "APR-001",
				"employee_id":       "EMP-001",
				"performance_rating": "Excellent",
				"new_salary":        150000.00,
			},
			false,
		},
		{
			"missing appraisal_id",
			map[string]interface{}{
				"employee_id":        "EMP-001",
				"performance_rating": "Excellent",
				"new_salary":         150000.00,
			},
			true,
		},
		{
			"missing employee_id",
			map[string]interface{}{
				"appraisal_id":       "APR-001",
				"performance_rating": "Excellent",
				"new_salary":         150000.00,
			},
			true,
		},
		{
			"missing performance_rating",
			map[string]interface{}{
				"appraisal_id": "APR-001",
				"employee_id":  "EMP-001",
				"new_salary":   150000.00,
			},
			true,
		},
		{
			"missing new_salary",
			map[string]interface{}{
				"appraisal_id":       "APR-001",
				"employee_id":        "EMP-001",
				"performance_rating": "Excellent",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateInput() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAppraisalSalaryRevisionSagaCriticalSteps(t *testing.T) {
	s := NewAppraisalSalaryRevisionSaga()
	expectedCritical := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 9: true}

	for stepNum, shouldBeCritical := range expectedCritical {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.IsCritical != shouldBeCritical {
			t.Errorf("step %d critical status mismatch: expected %v, got %v", stepNum, shouldBeCritical, step.IsCritical)
		}
	}
}

func TestAppraisalSalaryRevisionSagaGetStepDefinition(t *testing.T) {
	s := NewAppraisalSalaryRevisionSaga()
	step := s.GetStepDefinition(1)
	if step == nil {
		t.Errorf("step 1 not found")
	}
	if step.ServiceName != "appraisal" {
		t.Errorf("step 1 service mismatch: expected 'appraisal', got %s", step.ServiceName)
	}
}

// ===== SAGA-H05: Leave Application Saga Tests =====

func TestLeaveApplicationSagaType(t *testing.T) {
	s := NewLeaveApplicationSaga()
	if s.SagaType() != "SAGA-H05" {
		t.Errorf("expected SAGA-H05, got %s", s.SagaType())
	}
}

func TestLeaveApplicationSagaStepCount(t *testing.T) {
	s := NewLeaveApplicationSaga()
	steps := s.GetStepDefinitions()
	// Should have 8 forward + 4 compensation = 12 total
	if len(steps) != 12 {
		t.Errorf("expected 12 steps, got %d", len(steps))
	}
}

func TestLeaveApplicationSagaInputValidation(t *testing.T) {
	s := NewLeaveApplicationSaga()
	tests := []struct {
		name      string
		input     map[string]interface{}
		wantError bool
	}{
		{
			"valid input",
			map[string]interface{}{
				"employee_id": "EMP-001",
				"leave_type":  "Annual",
				"start_date":  "2026-03-01",
				"end_date":    "2026-03-10",
				"reason":      "Vacation",
				"manager_id":  "MGR-001",
			},
			false,
		},
		{
			"missing employee_id",
			map[string]interface{}{
				"leave_type": "Annual",
				"start_date": "2026-03-01",
				"end_date":   "2026-03-10",
				"reason":     "Vacation",
				"manager_id": "MGR-001",
			},
			true,
		},
		{
			"missing leave_type",
			map[string]interface{}{
				"employee_id": "EMP-001",
				"start_date":  "2026-03-01",
				"end_date":    "2026-03-10",
				"reason":      "Vacation",
				"manager_id":  "MGR-001",
			},
			true,
		},
		{
			"missing start_date",
			map[string]interface{}{
				"employee_id": "EMP-001",
				"leave_type":  "Annual",
				"end_date":    "2026-03-10",
				"reason":      "Vacation",
				"manager_id":  "MGR-001",
			},
			true,
		},
		{
			"missing end_date",
			map[string]interface{}{
				"employee_id": "EMP-001",
				"leave_type":  "Annual",
				"start_date":  "2026-03-01",
				"reason":      "Vacation",
				"manager_id":  "MGR-001",
			},
			true,
		},
		{
			"missing reason",
			map[string]interface{}{
				"employee_id": "EMP-001",
				"leave_type":  "Annual",
				"start_date":  "2026-03-01",
				"end_date":    "2026-03-10",
				"manager_id":  "MGR-001",
			},
			true,
		},
		{
			"missing manager_id",
			map[string]interface{}{
				"employee_id": "EMP-001",
				"leave_type":  "Annual",
				"start_date":  "2026-03-01",
				"end_date":    "2026-03-10",
				"reason":      "Vacation",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateInput() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestLeaveApplicationSagaCriticalSteps(t *testing.T) {
	s := NewLeaveApplicationSaga()
	expectedCritical := map[int]bool{1: true, 2: true, 4: true, 8: true}

	for stepNum, shouldBeCritical := range expectedCritical {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.IsCritical != shouldBeCritical {
			t.Errorf("step %d critical status mismatch: expected %v, got %v", stepNum, shouldBeCritical, step.IsCritical)
		}
	}
}

func TestLeaveApplicationSagaGetStepDefinition(t *testing.T) {
	s := NewLeaveApplicationSaga()
	step := s.GetStepDefinition(1)
	if step == nil {
		t.Errorf("step 1 not found")
	}
	if step.ServiceName != "leave" {
		t.Errorf("step 1 service mismatch: expected 'leave', got %s", step.ServiceName)
	}
}

// ===== Cross-Saga Tests =====

func TestAllHRSagasImplementInterface(t *testing.T) {
	var _ saga.SagaHandler = NewEmployeeOnboardingSaga()
	var _ saga.SagaHandler = NewExpenseReimbursementSaga()
	var _ saga.SagaHandler = NewAppraisalSalaryRevisionSaga()
	var _ saga.SagaHandler = NewLeaveApplicationSaga()
	// Phase 4C sagas: TODO
	// var _ saga.SagaHandler = NewPayrollProcessingSaga()
	// var _ saga.SagaHandler = NewEmployeeExitSaga()
}

func TestUniqueHRSagaTypes(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewEmployeeOnboardingSaga(),
		NewExpenseReimbursementSaga(),
		NewAppraisalSalaryRevisionSaga(),
		NewLeaveApplicationSaga(),
		// Phase 4C: TODO add H01, H03
	}

	sagaTypes := make(map[string]bool)
	for _, handler := range sagas {
		sagaType := handler.SagaType()
		if sagaTypes[sagaType] {
			t.Errorf("duplicate saga type: %s", sagaType)
		}
		sagaTypes[sagaType] = true
	}

	expectedTypes := map[string]bool{
		"SAGA-H02": true,
		"SAGA-H04": true,
		"SAGA-H05": true,
		"SAGA-H06": true,
	}

	for expectedType := range expectedTypes {
		if !sagaTypes[expectedType] {
			t.Errorf("missing saga type: %s", expectedType)
		}
	}
}

func TestHRSagasHaveSteps(t *testing.T) {
	sagas := map[string]saga.SagaHandler{
		"SAGA-H02": NewEmployeeOnboardingSaga(),
		"SAGA-H04": NewExpenseReimbursementSaga(),
		"SAGA-H05": NewLeaveApplicationSaga(),
		"SAGA-H06": NewAppraisalSalaryRevisionSaga(),
	}

	for sagaType, handler := range sagas {
		steps := handler.GetStepDefinitions()
		if len(steps) == 0 {
			t.Errorf("%s has no steps", sagaType)
		}

		step1 := handler.GetStepDefinition(1)
		if step1 == nil {
			t.Errorf("%s missing step 1", sagaType)
		}
	}
}

func TestHRSagasHaveServiceNames(t *testing.T) {
	sagas := map[string]saga.SagaHandler{
		"SAGA-H02": NewEmployeeOnboardingSaga(),
		"SAGA-H04": NewExpenseReimbursementSaga(),
		"SAGA-H05": NewLeaveApplicationSaga(),
		"SAGA-H06": NewAppraisalSalaryRevisionSaga(),
	}

	for sagaType, handler := range sagas {
		steps := handler.GetStepDefinitions()
		for _, step := range steps {
			if step.ServiceName == "" {
				t.Errorf("%s step %d has empty service name", sagaType, step.StepNumber)
			}
		}
	}
}

func TestHRSagasHaveHandlerMethods(t *testing.T) {
	sagas := map[string]saga.SagaHandler{
		"SAGA-H02": NewEmployeeOnboardingSaga(),
		"SAGA-H04": NewExpenseReimbursementSaga(),
		"SAGA-H05": NewLeaveApplicationSaga(),
		"SAGA-H06": NewAppraisalSalaryRevisionSaga(),
	}

	for sagaType, handler := range sagas {
		steps := handler.GetStepDefinitions()
		for _, step := range steps {
			if step.HandlerMethod == "" {
				t.Errorf("%s step %d has empty handler method", sagaType, step.StepNumber)
			}
		}
	}
}

func TestHRSagasTimeoutValues(t *testing.T) {
	sagas := map[string]saga.SagaHandler{
		"SAGA-H02": NewEmployeeOnboardingSaga(),
		"SAGA-H04": NewExpenseReimbursementSaga(),
		"SAGA-H05": NewLeaveApplicationSaga(),
		"SAGA-H06": NewAppraisalSalaryRevisionSaga(),
	}

	for sagaType, handler := range sagas {
		steps := handler.GetStepDefinitions()
		for _, step := range steps {
			if step.TimeoutSeconds <= 0 {
				t.Errorf("%s step %d has invalid timeout: %d", sagaType, step.StepNumber, step.TimeoutSeconds)
			}
			if step.TimeoutSeconds > 120 {
				t.Logf("%s step %d timeout unusually high: %d seconds", sagaType, step.StepNumber, step.TimeoutSeconds)
			}
		}
	}
}

func TestEmployeeOnboardingSagaInputValidationInvalidType(t *testing.T) {
	s := NewEmployeeOnboardingSaga()
	err := s.ValidateInput("not a map")
	if err == nil {
		t.Errorf("expected error for invalid input type, got nil")
	}
}

func TestExpenseReimbursementSagaInputValidationInvalidType(t *testing.T) {
	s := NewExpenseReimbursementSaga()
	err := s.ValidateInput([]string{"array"})
	if err == nil {
		t.Errorf("expected error for invalid input type, got nil")
	}
}

func TestAppraisalSalaryRevisionSagaInputValidationInvalidType(t *testing.T) {
	s := NewAppraisalSalaryRevisionSaga()
	err := s.ValidateInput(123)
	if err == nil {
		t.Errorf("expected error for invalid input type, got nil")
	}
}

func TestLeaveApplicationSagaInputValidationInvalidType(t *testing.T) {
	s := NewLeaveApplicationSaga()
	err := s.ValidateInput(nil)
	if err == nil {
		t.Errorf("expected error for invalid input type, got nil")
	}
}
