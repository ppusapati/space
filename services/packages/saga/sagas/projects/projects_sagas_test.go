// Package projects provides tests for projects module saga workflows
package projects

import (
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// ===== SAGA-PR01: Project Billing Saga Tests =====

func TestProjectBillingSagaType(t *testing.T) {
	s := NewProjectBillingSaga()
	if s.SagaType() != "SAGA-PR01" {
		t.Errorf("expected SAGA-PR01, got %s", s.SagaType())
	}
}

func TestProjectBillingSagaStepCount(t *testing.T) {
	s := NewProjectBillingSaga()
	steps := s.GetStepDefinitions()
	// Should have 7 forward + 5 compensation = 12 total
	if len(steps) != 12 {
		t.Errorf("expected 12 steps, got %d", len(steps))
	}
}

func TestProjectBillingSagaInputValidation(t *testing.T) {
	s := NewProjectBillingSaga()
	tests := []struct {
		name      string
		input     map[string]interface{}
		wantError bool
	}{
		{
			"valid input",
			map[string]interface{}{
				"timesheet_id":  "TS-001",
				"project_id":    "PRJ-001",
				"billing_period": "2026-02",
			},
			false,
		},
		{
			"missing timesheet_id",
			map[string]interface{}{
				"project_id":    "PRJ-001",
				"billing_period": "2026-02",
			},
			true,
		},
		{
			"missing project_id",
			map[string]interface{}{
				"timesheet_id":  "TS-001",
				"billing_period": "2026-02",
			},
			true,
		},
		{
			"missing billing_period",
			map[string]interface{}{
				"timesheet_id": "TS-001",
				"project_id":   "PRJ-001",
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

func TestProjectBillingSagaGetStepDefinition(t *testing.T) {
	s := NewProjectBillingSaga()
	tests := []struct {
		stepNum    int
		wantExists bool
	}{
		{1, true},
		{4, true},
		{7, true},
		{101, true},
		{105, true},
		{0, false},
		{8, false},
		{106, false},
	}

	for _, tt := range tests {
		step := s.GetStepDefinition(tt.stepNum)
		if (step != nil) != tt.wantExists {
			t.Errorf("GetStepDefinition(%d) exists = %v, want %v", tt.stepNum, step != nil, tt.wantExists)
		}
	}
}

func TestProjectBillingSagaCriticalSteps(t *testing.T) {
	s := NewProjectBillingSaga()
	expectedCritical := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 7: true}

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

func TestProjectBillingSagaServiceNames(t *testing.T) {
	s := NewProjectBillingSaga()
	expectedServices := map[int]string{
		1: "timesheet",
		2: "project-costing",
		3: "sales-invoice",
		4: "general-ledger",
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

func TestProjectBillingSagaCompensationSteps(t *testing.T) {
	s := NewProjectBillingSaga()

	step1 := s.GetStepDefinition(1)
	if step1 != nil && len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have no compensation steps, got %d", len(step1.CompensationSteps))
	}

	step7 := s.GetStepDefinition(7)
	if step7 != nil && len(step7.CompensationSteps) != 0 {
		t.Errorf("step 7 should have no compensation steps, got %d", len(step7.CompensationSteps))
	}
}

// ===== SAGA-PR02: Progress Billing Saga Tests =====

func TestProgressBillingSagaType(t *testing.T) {
	s := NewProgressBillingSaga()
	if s.SagaType() != "SAGA-PR02" {
		t.Errorf("expected SAGA-PR02, got %s", s.SagaType())
	}
}

func TestProgressBillingSagaStepCount(t *testing.T) {
	s := NewProgressBillingSaga()
	steps := s.GetStepDefinitions()
	// Should have 8 forward + 6 compensation = 14 total
	if len(steps) != 14 {
		t.Errorf("expected 14 steps, got %d", len(steps))
	}
}

func TestProgressBillingSagaInputValidation(t *testing.T) {
	s := NewProgressBillingSaga()
	tests := []struct {
		name      string
		input     map[string]interface{}
		wantError bool
	}{
		{
			"valid input",
			map[string]interface{}{
				"boq_id":              "BOQ-001",
				"billing_cycle":       "Cycle 1",
				"progress_percentage": 25,
				"measured_date":       "2026-02-14",
			},
			false,
		},
		{
			"missing boq_id",
			map[string]interface{}{
				"billing_cycle":       "Cycle 1",
				"progress_percentage": 25,
				"measured_date":       "2026-02-14",
			},
			true,
		},
		{
			"missing billing_cycle",
			map[string]interface{}{
				"boq_id":              "BOQ-001",
				"progress_percentage": 25,
				"measured_date":       "2026-02-14",
			},
			true,
		},
		{
			"missing progress_percentage",
			map[string]interface{}{
				"boq_id":        "BOQ-001",
				"billing_cycle": "Cycle 1",
				"measured_date": "2026-02-14",
			},
			true,
		},
		{
			"missing measured_date",
			map[string]interface{}{
				"boq_id":              "BOQ-001",
				"billing_cycle":       "Cycle 1",
				"progress_percentage": 25,
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

func TestProgressBillingSagaCriticalSteps(t *testing.T) {
	s := NewProgressBillingSaga()
	expectedCritical := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 8: true}

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

func TestProgressBillingSagaGetStepDefinition(t *testing.T) {
	s := NewProgressBillingSaga()
	step := s.GetStepDefinition(1)
	if step == nil {
		t.Errorf("step 1 not found")
	}
	if step.ServiceName != "boq" {
		t.Errorf("step 1 service mismatch: expected 'boq', got %s", step.ServiceName)
	}
}

func TestProgressBillingSagaCompensationSteps(t *testing.T) {
	s := NewProgressBillingSaga()

	step1 := s.GetStepDefinition(1)
	if step1 != nil && len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have no compensation steps, got %d", len(step1.CompensationSteps))
	}

	step8 := s.GetStepDefinition(8)
	if step8 != nil && len(step8.CompensationSteps) != 0 {
		t.Errorf("step 8 should have no compensation steps, got %d", len(step8.CompensationSteps))
	}
}

// ===== SAGA-PR03: Subcontractor Payment Saga Tests =====

func TestSubcontractorPaymentSagaType(t *testing.T) {
	s := NewSubcontractorPaymentSaga()
	if s.SagaType() != "SAGA-PR03" {
		t.Errorf("expected SAGA-PR03, got %s", s.SagaType())
	}
}

func TestSubcontractorPaymentSagaStepCount(t *testing.T) {
	s := NewSubcontractorPaymentSaga()
	steps := s.GetStepDefinitions()
	// Should have 9 forward + 8 compensation = 17 total
	if len(steps) != 17 {
		t.Errorf("expected 17 steps, got %d", len(steps))
	}
}

func TestSubcontractorPaymentSagaInputValidation(t *testing.T) {
	s := NewSubcontractorPaymentSaga()
	tests := []struct {
		name      string
		input     map[string]interface{}
		wantError bool
	}{
		{
			"valid input",
			map[string]interface{}{
				"invoice_id":     "INV-001",
				"contractor_id":  "CON-001",
				"invoice_amount": 50000.00,
				"deduction_type": "Retention",
				"project_id":     "PRJ-001",
			},
			false,
		},
		{
			"missing invoice_id",
			map[string]interface{}{
				"contractor_id":  "CON-001",
				"invoice_amount": 50000.00,
				"deduction_type": "Retention",
				"project_id":     "PRJ-001",
			},
			true,
		},
		{
			"missing contractor_id",
			map[string]interface{}{
				"invoice_id":     "INV-001",
				"invoice_amount": 50000.00,
				"deduction_type": "Retention",
				"project_id":     "PRJ-001",
			},
			true,
		},
		{
			"missing invoice_amount",
			map[string]interface{}{
				"invoice_id":     "INV-001",
				"contractor_id":  "CON-001",
				"deduction_type": "Retention",
				"project_id":     "PRJ-001",
			},
			true,
		},
		{
			"missing deduction_type",
			map[string]interface{}{
				"invoice_id":     "INV-001",
				"contractor_id":  "CON-001",
				"invoice_amount": 50000.00,
				"project_id":     "PRJ-001",
			},
			true,
		},
		{
			"missing project_id",
			map[string]interface{}{
				"invoice_id":     "INV-001",
				"contractor_id":  "CON-001",
				"invoice_amount": 50000.00,
				"deduction_type": "Retention",
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

func TestSubcontractorPaymentSagaCriticalSteps(t *testing.T) {
	s := NewSubcontractorPaymentSaga()
	expectedCritical := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 9: true}

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

func TestSubcontractorPaymentSagaGetStepDefinition(t *testing.T) {
	s := NewSubcontractorPaymentSaga()
	step := s.GetStepDefinition(1)
	if step == nil {
		t.Errorf("step 1 not found")
	}
	if step.ServiceName != "accounts-payable" {
		t.Errorf("step 1 service mismatch: expected 'accounts-payable', got %s", step.ServiceName)
	}
}

func TestSubcontractorPaymentSagaCompensationSteps(t *testing.T) {
	s := NewSubcontractorPaymentSaga()

	step1 := s.GetStepDefinition(1)
	if step1 != nil && len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have no compensation steps, got %d", len(step1.CompensationSteps))
	}

	step9 := s.GetStepDefinition(9)
	if step9 != nil && len(step9.CompensationSteps) != 0 {
		t.Errorf("step 9 should have no compensation steps, got %d", len(step9.CompensationSteps))
	}
}

// ===== SAGA-PR04: Project Close Saga Tests =====

func TestProjectCloseSagaType(t *testing.T) {
	s := NewProjectCloseSaga()
	if s.SagaType() != "SAGA-PR04" {
		t.Errorf("expected SAGA-PR04, got %s", s.SagaType())
	}
}

func TestProjectCloseSagaStepCount(t *testing.T) {
	s := NewProjectCloseSaga()
	steps := s.GetStepDefinitions()
	// Should have 8 forward + 6 compensation = 14 total
	if len(steps) != 14 {
		t.Errorf("expected 14 steps, got %d", len(steps))
	}
}

func TestProjectCloseSagaInputValidation(t *testing.T) {
	s := NewProjectCloseSaga()
	tests := []struct {
		name      string
		input     map[string]interface{}
		wantError bool
	}{
		{
			"valid input",
			map[string]interface{}{
				"project_id":   "PRJ-001",
				"closing_date": "2026-02-28",
				"final_amount": 500000.00,
			},
			false,
		},
		{
			"missing project_id",
			map[string]interface{}{
				"closing_date": "2026-02-28",
				"final_amount": 500000.00,
			},
			true,
		},
		{
			"missing closing_date",
			map[string]interface{}{
				"project_id":   "PRJ-001",
				"final_amount": 500000.00,
			},
			true,
		},
		{
			"missing final_amount",
			map[string]interface{}{
				"project_id":   "PRJ-001",
				"closing_date": "2026-02-28",
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

func TestProjectCloseSagaCriticalSteps(t *testing.T) {
	s := NewProjectCloseSaga()
	expectedCritical := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 8: true}

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

func TestProjectCloseSagaGetStepDefinition(t *testing.T) {
	s := NewProjectCloseSaga()
	step := s.GetStepDefinition(1)
	if step == nil {
		t.Errorf("step 1 not found")
	}
	if step.ServiceName != "project" {
		t.Errorf("step 1 service mismatch: expected 'project', got %s", step.ServiceName)
	}
}

func TestProjectCloseSagaCompensationSteps(t *testing.T) {
	s := NewProjectCloseSaga()

	step1 := s.GetStepDefinition(1)
	if step1 != nil && len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have no compensation steps, got %d", len(step1.CompensationSteps))
	}

	step8 := s.GetStepDefinition(8)
	if step8 != nil && len(step8.CompensationSteps) != 0 {
		t.Errorf("step 8 should have no compensation steps, got %d", len(step8.CompensationSteps))
	}
}

// ===== Cross-Saga Tests =====

func TestAllProjectsSagasImplementInterface(t *testing.T) {
	var _ saga.SagaHandler = NewProjectBillingSaga()
	var _ saga.SagaHandler = NewProgressBillingSaga()
	var _ saga.SagaHandler = NewSubcontractorPaymentSaga()
	var _ saga.SagaHandler = NewProjectCloseSaga()
}

func TestUniqueProjectsSagaTypes(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewProjectBillingSaga(),
		NewProgressBillingSaga(),
		NewSubcontractorPaymentSaga(),
		NewProjectCloseSaga(),
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
		"SAGA-PR01": true,
		"SAGA-PR02": true,
		"SAGA-PR03": true,
		"SAGA-PR04": true,
	}

	for expectedType := range expectedTypes {
		if !sagaTypes[expectedType] {
			t.Errorf("missing saga type: %s", expectedType)
		}
	}
}

func TestProjectsSagasHaveSteps(t *testing.T) {
	sagas := map[string]saga.SagaHandler{
		"SAGA-PR01": NewProjectBillingSaga(),
		"SAGA-PR02": NewProgressBillingSaga(),
		"SAGA-PR03": NewSubcontractorPaymentSaga(),
		"SAGA-PR04": NewProjectCloseSaga(),
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

func TestProjectsSagasHaveServiceNames(t *testing.T) {
	sagas := map[string]saga.SagaHandler{
		"SAGA-PR01": NewProjectBillingSaga(),
		"SAGA-PR02": NewProgressBillingSaga(),
		"SAGA-PR03": NewSubcontractorPaymentSaga(),
		"SAGA-PR04": NewProjectCloseSaga(),
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

func TestProjectsSagasHaveHandlerMethods(t *testing.T) {
	sagas := map[string]saga.SagaHandler{
		"SAGA-PR01": NewProjectBillingSaga(),
		"SAGA-PR02": NewProgressBillingSaga(),
		"SAGA-PR03": NewSubcontractorPaymentSaga(),
		"SAGA-PR04": NewProjectCloseSaga(),
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

func TestProjectsSagasTimeoutValues(t *testing.T) {
	sagas := map[string]saga.SagaHandler{
		"SAGA-PR01": NewProjectBillingSaga(),
		"SAGA-PR02": NewProgressBillingSaga(),
		"SAGA-PR03": NewSubcontractorPaymentSaga(),
		"SAGA-PR04": NewProjectCloseSaga(),
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

func TestProjectBillingSagaInputValidationInvalidType(t *testing.T) {
	s := NewProjectBillingSaga()
	err := s.ValidateInput("not a map")
	if err == nil {
		t.Errorf("expected error for invalid input type, got nil")
	}
}

func TestProgressBillingSagaInputValidationInvalidType(t *testing.T) {
	s := NewProgressBillingSaga()
	err := s.ValidateInput(42)
	if err == nil {
		t.Errorf("expected error for invalid input type, got nil")
	}
}

func TestSubcontractorPaymentSagaInputValidationInvalidType(t *testing.T) {
	s := NewSubcontractorPaymentSaga()
	err := s.ValidateInput([]int{1, 2, 3})
	if err == nil {
		t.Errorf("expected error for invalid input type, got nil")
	}
}

func TestProjectCloseSagaInputValidationInvalidType(t *testing.T) {
	s := NewProjectCloseSaga()
	err := s.ValidateInput(nil)
	if err == nil {
		t.Errorf("expected error for invalid input type, got nil")
	}
}

func TestProjectsSagasRetryConfig(t *testing.T) {
	sagas := map[string]saga.SagaHandler{
		"SAGA-PR01": NewProjectBillingSaga(),
		"SAGA-PR02": NewProgressBillingSaga(),
		"SAGA-PR03": NewSubcontractorPaymentSaga(),
		"SAGA-PR04": NewProjectCloseSaga(),
	}

	for sagaType, handler := range sagas {
		step1 := handler.GetStepDefinition(1)
		if step1 == nil {
			t.Errorf("%s step 1 not found", sagaType)
			continue
		}

		if step1.RetryConfig == nil {
			t.Errorf("%s step 1 retry config is nil", sagaType)
			continue
		}

		if step1.RetryConfig.MaxRetries != 3 {
			t.Errorf("%s step 1 MaxRetries mismatch: expected 3, got %d", sagaType, step1.RetryConfig.MaxRetries)
		}

		if step1.RetryConfig.InitialBackoffMs != 1000 {
			t.Errorf("%s step 1 InitialBackoffMs mismatch: expected 1000, got %d", sagaType, step1.RetryConfig.InitialBackoffMs)
		}
	}
}

func TestProjectsSagasInputMappingPresence(t *testing.T) {
	sagas := map[string]saga.SagaHandler{
		"SAGA-PR01": NewProjectBillingSaga(),
		"SAGA-PR02": NewProgressBillingSaga(),
		"SAGA-PR03": NewSubcontractorPaymentSaga(),
		"SAGA-PR04": NewProjectCloseSaga(),
	}

	for sagaType, handler := range sagas {
		steps := handler.GetStepDefinitions()
		for _, step := range steps {
			if len(step.InputMapping) == 0 {
				t.Errorf("%s step %d has no input mapping", sagaType, step.StepNumber)
			}
		}
	}
}

func TestProjectsSagasStepNumberSequence(t *testing.T) {
	sagas := map[string]saga.SagaHandler{
		"SAGA-PR01": NewProjectBillingSaga(),
		"SAGA-PR02": NewProgressBillingSaga(),
		"SAGA-PR03": NewSubcontractorPaymentSaga(),
		"SAGA-PR04": NewProjectCloseSaga(),
	}

	for sagaType, handler := range sagas {
		steps := handler.GetStepDefinitions()
		for i, step := range steps {
			if step.StepNumber != int32(i+1) {
				t.Logf("Warning: %s step at index %d has step number %d (expected sequential)", sagaType, i, step.StepNumber)
			}
		}
	}
}
