// Package hr provides tests for HR Statutory module saga workflows
package hr

import (
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// ===== SAGA-SR01: Form 16 Generation Tests =====

func TestForm16GenerationSagaType(t *testing.T) {
	s := NewForm16GenerationSaga()
	if s.SagaType() != "SAGA-SR01" {
		t.Errorf("expected SAGA-SR01, got %s", s.SagaType())
	}
}

func TestForm16GenerationSagaStepCount(t *testing.T) {
	s := NewForm16GenerationSaga()
	steps := s.GetStepDefinitions()
	// Should have 9 forward + 8 compensation = 17 total
	if len(steps) != 17 {
		t.Errorf("expected 17 steps, got %d", len(steps))
	}
}

func TestForm16GenerationSagaGetStepDefinition(t *testing.T) {
	s := NewForm16GenerationSaga()
	tests := []struct {
		name     string
		stepNum  int
		expected string
	}{
		{"step 1", 1, "ExtractEmployeeSalaryRecords"},
		{"step 3", 3, "SumTDSDeductions"},
		{"step 6", 6, "CalculateNetTaxableIncome"},
		{"step 9", 9, "IssueForm16ToEmployee"},
		{"step 101", 101, "UndoCalculateTotalGrossSalary"},
		{"step 108", 108, "UndoIssueForm16ToEmployee"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := s.GetStepDefinition(tc.stepNum)
			if step == nil {
				t.Errorf("step %d not found", tc.stepNum)
				return
			}
			if step.HandlerMethod != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, step.HandlerMethod)
			}
		})
	}
}

func TestForm16GenerationSagaCriticalSteps(t *testing.T) {
	s := NewForm16GenerationSaga()
	steps := s.GetStepDefinitions()
	criticalSteps := []int{3, 6, 9}

	for _, stepNum := range criticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("critical step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be marked critical", stepNum)
		}
	}

	// Verify non-critical steps
	nonCriticalSteps := []int{1, 2, 4, 5, 7, 8}
	for _, stepNum := range nonCriticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("non-critical step %d not found", stepNum)
			continue
		}
		if step.IsCritical {
			t.Errorf("step %d should not be marked critical", stepNum)
		}
	}
}

func TestForm16GenerationSagaInputValidation(t *testing.T) {
	s := NewForm16GenerationSaga()
	tests := []struct {
		name      string
		input     map[string]interface{}
		wantError bool
	}{
		{
			"valid input",
			map[string]interface{}{
				"input": map[string]interface{}{
					"form16_generation_id": "F16-2024-001",
					"assessment_year":      "2023-24",
					"employee_id":          "EMP-001",
					"company_id":           "COM-001",
				},
			},
			false,
		},
		{
			"missing form16_generation_id",
			map[string]interface{}{
				"input": map[string]interface{}{
					"assessment_year": "2023-24",
					"employee_id":     "EMP-001",
					"company_id":      "COM-001",
				},
			},
			true,
		},
		{
			"missing assessment_year",
			map[string]interface{}{
				"input": map[string]interface{}{
					"form16_generation_id": "F16-2024-001",
					"employee_id":          "EMP-001",
					"company_id":           "COM-001",
				},
			},
			true,
		},
		{
			"missing employee_id",
			map[string]interface{}{
				"input": map[string]interface{}{
					"form16_generation_id": "F16-2024-001",
					"assessment_year":      "2023-24",
					"company_id":           "COM-001",
				},
			},
			true,
		},
		{
			"missing company_id",
			map[string]interface{}{
				"input": map[string]interface{}{
					"form16_generation_id": "F16-2024-001",
					"assessment_year":      "2023-24",
					"employee_id":          "EMP-001",
				},
			},
			true,
		},
		{
			"empty form16_generation_id",
			map[string]interface{}{
				"input": map[string]interface{}{
					"form16_generation_id": "",
					"assessment_year":      "2023-24",
					"employee_id":          "EMP-001",
					"company_id":           "COM-001",
				},
			},
			true,
		},
		{
			"invalid input type",
			"not a map",
			true,
		},
		{
			"missing input field",
			map[string]interface{}{
				"form16_generation_id": "F16-2024-001",
			},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := s.ValidateInput(tc.input)
			if (err != nil) != tc.wantError {
				t.Errorf("ValidateInput() error = %v, wantError %v", err, tc.wantError)
			}
		})
	}
}

func TestForm16GenerationSagaServiceNames(t *testing.T) {
	s := NewForm16GenerationSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[int]string{
		1: "payroll",
		2: "salary-structure",
		3: "tds",
		4: "payroll",
		5: "payroll",
		6: "tds",
		7: "payroll",
		8: "tds",
		9: "notification",
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

func TestForm16GenerationSagaCompensationSteps(t *testing.T) {
	s := NewForm16GenerationSaga()

	tests := []struct {
		forwardStep       int
		expectedCompSteps []int32
	}{
		{2, []int32{101}},
		{3, []int32{102}},
		{4, []int32{103}},
		{5, []int32{104}},
		{6, []int32{105}},
		{7, []int32{106}},
		{8, []int32{107}},
		{9, []int32{108}},
	}

	for _, tc := range tests {
		step := s.GetStepDefinition(tc.forwardStep)
		if step == nil {
			t.Errorf("step %d not found", tc.forwardStep)
			continue
		}
		if len(step.CompensationSteps) != len(tc.expectedCompSteps) {
			t.Errorf("step %d: expected %d compensation steps, got %d", tc.forwardStep, len(tc.expectedCompSteps), len(step.CompensationSteps))
			continue
		}
		for i, expected := range tc.expectedCompSteps {
			if i >= len(step.CompensationSteps) || step.CompensationSteps[i] != expected {
				t.Errorf("step %d: expected compensation step %d, got %d", tc.forwardStep, expected, step.CompensationSteps[i])
			}
		}
	}
}

// ===== SAGA-SR02: PF/ESI Remittance Tests =====

func TestPFESIRemittanceSagaType(t *testing.T) {
	s := NewPFESIRemittanceSaga()
	if s.SagaType() != "SAGA-SR02" {
		t.Errorf("expected SAGA-SR02, got %s", s.SagaType())
	}
}

func TestPFESIRemittanceSagaStepCount(t *testing.T) {
	s := NewPFESIRemittanceSaga()
	steps := s.GetStepDefinitions()
	// Should have 10 forward + 9 compensation = 19 total
	if len(steps) != 19 {
		t.Errorf("expected 19 steps, got %d", len(steps))
	}
}

func TestPFESIRemittanceSagaGetStepDefinition(t *testing.T) {
	s := NewPFESIRemittanceSaga()
	tests := []struct {
		name     string
		stepNum  int
		expected string
	}{
		{"step 1", 1, "ExtractEmployeeAttendanceRecords"},
		{"step 3", 3, "CheckESIApplicability"},
		{"step 5", 5, "DeductPFFromSalary"},
		{"step 8", 8, "PostPFESILiabilityEntries"},
		{"step 10", 10, "MarkRemittanceSubmitted"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := s.GetStepDefinition(tc.stepNum)
			if step == nil {
				t.Errorf("step %d not found", tc.stepNum)
				return
			}
			if step.HandlerMethod != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, step.HandlerMethod)
			}
		})
	}
}

func TestPFESIRemittanceSagaCriticalSteps(t *testing.T) {
	s := NewPFESIRemittanceSaga()
	criticalSteps := []int{3, 5, 8}

	for _, stepNum := range criticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("critical step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be marked critical", stepNum)
		}
	}
}

func TestPFESIRemittanceSagaInputValidation(t *testing.T) {
	s := NewPFESIRemittanceSaga()
	tests := []struct {
		name      string
		input     map[string]interface{}
		wantError bool
	}{
		{
			"valid input",
			map[string]interface{}{
				"input": map[string]interface{}{
					"pf_esi_remittance_id": "PER-2024-001",
					"remittance_month":     "January",
					"remittance_year":      "2024",
					"company_id":           "COM-001",
				},
			},
			false,
		},
		{
			"missing pf_esi_remittance_id",
			map[string]interface{}{
				"input": map[string]interface{}{
					"remittance_month": "January",
					"remittance_year":  "2024",
					"company_id":       "COM-001",
				},
			},
			true,
		},
		{
			"missing remittance_month",
			map[string]interface{}{
				"input": map[string]interface{}{
					"pf_esi_remittance_id": "PER-2024-001",
					"remittance_year":      "2024",
					"company_id":           "COM-001",
				},
			},
			true,
		},
		{
			"missing remittance_year",
			map[string]interface{}{
				"input": map[string]interface{}{
					"pf_esi_remittance_id": "PER-2024-001",
					"remittance_month":     "January",
					"company_id":           "COM-001",
				},
			},
			true,
		},
		{
			"missing company_id",
			map[string]interface{}{
				"input": map[string]interface{}{
					"pf_esi_remittance_id": "PER-2024-001",
					"remittance_month":     "January",
					"remittance_year":      "2024",
				},
			},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := s.ValidateInput(tc.input)
			if (err != nil) != tc.wantError {
				t.Errorf("ValidateInput() error = %v, wantError %v", err, tc.wantError)
			}
		})
	}
}

// ===== SAGA-SR03: TDS Payment & Return Filing Tests =====

func TestTDSPaymentReturnFilingSagaType(t *testing.T) {
	s := NewTDSPaymentReturnFilingSaga()
	if s.SagaType() != "SAGA-SR03" {
		t.Errorf("expected SAGA-SR03, got %s", s.SagaType())
	}
}

func TestTDSPaymentReturnFilingSagaStepCount(t *testing.T) {
	s := NewTDSPaymentReturnFilingSaga()
	steps := s.GetStepDefinitions()
	// Should have 11 forward + 10 compensation = 21 total
	if len(steps) != 21 {
		t.Errorf("expected 21 steps, got %d", len(steps))
	}
}

func TestTDSPaymentReturnFilingSagaGetStepDefinition(t *testing.T) {
	s := NewTDSPaymentReturnFilingSaga()
	tests := []struct {
		name     string
		stepNum  int
		expected string
	}{
		{"step 1", 1, "ExtractTDSDeductions"},
		{"step 4", 4, "CalculateTotalTDSDeducted"},
		{"step 6", 6, "MatchTDSDepositedVsDeducted"},
		{"step 10", 10, "SubmitTDSReturnToClearingHouse"},
		{"step 11", 11, "ArchiveReturnForCompliance"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := s.GetStepDefinition(tc.stepNum)
			if step == nil {
				t.Errorf("step %d not found", tc.stepNum)
				return
			}
			if step.HandlerMethod != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, step.HandlerMethod)
			}
		})
	}
}

func TestTDSPaymentReturnFilingSagaCriticalSteps(t *testing.T) {
	s := NewTDSPaymentReturnFilingSaga()
	criticalSteps := []int{4, 6, 10}

	for _, stepNum := range criticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("critical step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be marked critical", stepNum)
		}
	}
}

func TestTDSPaymentReturnFilingSagaInputValidation(t *testing.T) {
	s := NewTDSPaymentReturnFilingSaga()
	tests := []struct {
		name      string
		input     map[string]interface{}
		wantError bool
	}{
		{
			"valid input",
			map[string]interface{}{
				"input": map[string]interface{}{
					"tds_return_filing_id": "TRF-2024-Q1",
					"filing_period":        "Q1",
					"assessment_year":      "2023-24",
					"company_id":           "COM-001",
				},
			},
			false,
		},
		{
			"missing tds_return_filing_id",
			map[string]interface{}{
				"input": map[string]interface{}{
					"filing_period":   "Q1",
					"assessment_year": "2023-24",
					"company_id":      "COM-001",
				},
			},
			true,
		},
		{
			"missing filing_period",
			map[string]interface{}{
				"input": map[string]interface{}{
					"tds_return_filing_id": "TRF-2024-Q1",
					"assessment_year":      "2023-24",
					"company_id":           "COM-001",
				},
			},
			true,
		},
		{
			"missing assessment_year",
			map[string]interface{}{
				"input": map[string]interface{}{
					"tds_return_filing_id": "TRF-2024-Q1",
					"filing_period":        "Q1",
					"company_id":           "COM-001",
				},
			},
			true,
		},
		{
			"missing company_id",
			map[string]interface{}{
				"input": map[string]interface{}{
					"tds_return_filing_id": "TRF-2024-Q1",
					"filing_period":        "Q1",
					"assessment_year":      "2023-24",
				},
			},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := s.ValidateInput(tc.input)
			if (err != nil) != tc.wantError {
				t.Errorf("ValidateInput() error = %v, wantError %v", err, tc.wantError)
			}
		})
	}
}

// ===== SAGA-SR04: Leave Encashment & Settlement Tests =====

func TestLeaveEncashmentSettlementSagaType(t *testing.T) {
	s := NewLeaveEncashmentSettlementSaga()
	if s.SagaType() != "SAGA-SR04" {
		t.Errorf("expected SAGA-SR04, got %s", s.SagaType())
	}
}

func TestLeaveEncashmentSettlementSagaStepCount(t *testing.T) {
	s := NewLeaveEncashmentSettlementSaga()
	steps := s.GetStepDefinitions()
	// Should have 8 forward + 7 compensation = 15 total
	if len(steps) != 15 {
		t.Errorf("expected 15 steps, got %d", len(steps))
	}
}

func TestLeaveEncashmentSettlementSagaGetStepDefinition(t *testing.T) {
	s := NewLeaveEncashmentSettlementSaga()
	tests := []struct {
		name     string
		stepNum  int
		expected string
	}{
		{"step 1", 1, "IdentifyEmployeeExit"},
		{"step 2", 2, "ExtractLeaveBalance"},
		{"step 5", 5, "CalculateOtherDues"},
		{"step 7", 7, "PostSettlementGLEntries"},
		{"step 8", 8, "GenerateSettlementReport"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := s.GetStepDefinition(tc.stepNum)
			if step == nil {
				t.Errorf("step %d not found", tc.stepNum)
				return
			}
			if step.HandlerMethod != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, step.HandlerMethod)
			}
		})
	}
}

func TestLeaveEncashmentSettlementSagaCriticalSteps(t *testing.T) {
	s := NewLeaveEncashmentSettlementSaga()
	criticalSteps := []int{2, 5, 7}

	for _, stepNum := range criticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("critical step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be marked critical", stepNum)
		}
	}
}

func TestLeaveEncashmentSettlementSagaInputValidation(t *testing.T) {
	s := NewLeaveEncashmentSettlementSaga()
	tests := []struct {
		name      string
		input     map[string]interface{}
		wantError bool
	}{
		{
			"valid input",
			map[string]interface{}{
				"input": map[string]interface{}{
					"settlement_id": "SETTLE-2024-001",
					"employee_id":   "EMP-001",
					"exit_type":     "resignation",
					"company_id":    "COM-001",
				},
			},
			false,
		},
		{
			"missing settlement_id",
			map[string]interface{}{
				"input": map[string]interface{}{
					"employee_id": "EMP-001",
					"exit_type":   "resignation",
					"company_id":  "COM-001",
				},
			},
			true,
		},
		{
			"missing employee_id",
			map[string]interface{}{
				"input": map[string]interface{}{
					"settlement_id": "SETTLE-2024-001",
					"exit_type":     "resignation",
					"company_id":    "COM-001",
				},
			},
			true,
		},
		{
			"missing exit_type",
			map[string]interface{}{
				"input": map[string]interface{}{
					"settlement_id": "SETTLE-2024-001",
					"employee_id":   "EMP-001",
					"company_id":    "COM-001",
				},
			},
			true,
		},
		{
			"missing company_id",
			map[string]interface{}{
				"input": map[string]interface{}{
					"settlement_id": "SETTLE-2024-001",
					"employee_id":   "EMP-001",
					"exit_type":     "resignation",
				},
			},
			true,
		},
		{
			"valid input - termination",
			map[string]interface{}{
				"input": map[string]interface{}{
					"settlement_id": "SETTLE-2024-002",
					"employee_id":   "EMP-002",
					"exit_type":     "termination",
					"company_id":    "COM-001",
				},
			},
			false,
		},
		{
			"valid input - retirement",
			map[string]interface{}{
				"input": map[string]interface{}{
					"settlement_id": "SETTLE-2024-003",
					"employee_id":   "EMP-003",
					"exit_type":     "retirement",
					"company_id":    "COM-001",
				},
			},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := s.ValidateInput(tc.input)
			if (err != nil) != tc.wantError {
				t.Errorf("ValidateInput() error = %v, wantError %v", err, tc.wantError)
			}
		})
	}
}

// ===== Cross-Saga Tests =====

func TestAllStatutorySagasImplementInterface(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewForm16GenerationSaga(),
		NewPFESIRemittanceSaga(),
		NewTDSPaymentReturnFilingSaga(),
		NewLeaveEncashmentSettlementSaga(),
	}

	for _, s := range sagas {
		if s.SagaType() == "" {
			t.Error("SagaType() returned empty string")
		}
		if len(s.GetStepDefinitions()) == 0 {
			t.Errorf("%s has no step definitions", s.SagaType())
		}
		// Test that all steps have positive step numbers
		for _, step := range s.GetStepDefinitions() {
			if step.StepNumber <= 0 {
				t.Errorf("%s step has non-positive number: %d", s.SagaType(), step.StepNumber)
			}
			// Test that retry config is present for forward steps (1-99)
			if step.StepNumber < 100 && step.RetryConfig == nil {
				t.Errorf("%s forward step %d missing retry config", s.SagaType(), step.StepNumber)
			}
		}
	}
}

func TestStatutorySagasRetryConfiguration(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewForm16GenerationSaga(),
		NewPFESIRemittanceSaga(),
		NewTDSPaymentReturnFilingSaga(),
		NewLeaveEncashmentSettlementSaga(),
	}

	for _, s := range sagas {
		for _, step := range s.GetStepDefinitions() {
			// Only forward steps should have retry config
			if step.StepNumber < 100 {
				if step.RetryConfig == nil {
					t.Errorf("%s step %d missing retry config", s.SagaType(), step.StepNumber)
					continue
				}
				if step.RetryConfig.MaxRetries != 3 {
					t.Errorf("%s step %d: expected 3 retries, got %d", s.SagaType(), step.StepNumber, step.RetryConfig.MaxRetries)
				}
				if step.RetryConfig.InitialBackoffMs != 1000 {
					t.Errorf("%s step %d: expected 1000ms initial backoff, got %d", s.SagaType(), step.StepNumber, step.RetryConfig.InitialBackoffMs)
				}
				if step.RetryConfig.BackoffMultiplier != 2.0 {
					t.Errorf("%s step %d: expected 2.0 backoff multiplier, got %f", s.SagaType(), step.StepNumber, step.RetryConfig.BackoffMultiplier)
				}
			}
		}
	}
}

func TestStatutorySagasServiceNameFormatting(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewForm16GenerationSaga(),
		NewPFESIRemittanceSaga(),
		NewTDSPaymentReturnFilingSaga(),
		NewLeaveEncashmentSettlementSaga(),
	}

	hyphenatedServices := map[string]bool{
		"general-ledger":      true,
		"accounts-payable":    true,
		"salary-structure":    true,
		"compliance-postings": true,
	}

	nonHyphenatedServices := map[string]bool{
		"payroll":      true,
		"tds":          true,
		"leave":        true,
		"employee":     true,
		"banking":      true,
		"attendance":   true,
		"approval":     true,
		"notification": true,
	}

	for _, s := range sagas {
		for _, step := range s.GetStepDefinitions() {
			if step.StepNumber < 100 { // Only forward steps
				// Check that service names are properly formatted
				if hyphenatedServices[step.ServiceName] {
					if step.ServiceName != "general-ledger" &&
						step.ServiceName != "accounts-payable" &&
						step.ServiceName != "salary-structure" &&
						step.ServiceName != "compliance-postings" {
						t.Errorf("%s step %d has unexpected service: %s", s.SagaType(), step.StepNumber, step.ServiceName)
					}
				} else if !nonHyphenatedServices[step.ServiceName] {
					t.Errorf("%s step %d has unknown service: %s", s.SagaType(), step.StepNumber, step.ServiceName)
				}
			}
		}
	}
}

func TestStatutorySagasTimeoutValues(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewForm16GenerationSaga(),
		NewPFESIRemittanceSaga(),
		NewTDSPaymentReturnFilingSaga(),
		NewLeaveEncashmentSettlementSaga(),
	}

	for _, s := range sagas {
		for _, step := range s.GetStepDefinitions() {
			if step.StepNumber < 100 { // Only forward steps
				if step.TimeoutSeconds <= 0 {
					t.Errorf("%s step %d has non-positive timeout: %d", s.SagaType(), step.StepNumber, step.TimeoutSeconds)
				}
				if step.TimeoutSeconds > 60 {
					t.Errorf("%s step %d has unusually high timeout: %d", s.SagaType(), step.StepNumber, step.TimeoutSeconds)
				}
			}
		}
	}
}

func TestStatutorySagasInputMappingCompleteness(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewForm16GenerationSaga(),
		NewPFESIRemittanceSaga(),
		NewTDSPaymentReturnFilingSaga(),
		NewLeaveEncashmentSettlementSaga(),
	}

	for _, s := range sagas {
		for _, step := range s.GetStepDefinitions() {
			if step.StepNumber < 100 { // Only forward steps
				if len(step.InputMapping) == 0 {
					t.Errorf("%s step %d has no input mapping", s.SagaType(), step.StepNumber)
				}
				// Check that tenantID and companyID are always mapped
				if step.InputMapping["tenantID"] == "" {
					t.Errorf("%s step %d missing tenantID mapping", s.SagaType(), step.StepNumber)
				}
				if step.InputMapping["companyID"] == "" {
					t.Errorf("%s step %d missing companyID mapping", s.SagaType(), step.StepNumber)
				}
			}
		}
	}
}

func TestForm16GenerationSagaServices(t *testing.T) {
	s := NewForm16GenerationSaga()
	expectedServices := map[int]string{
		1: "payroll",
		2: "salary-structure",
		3: "tds",
		4: "payroll",
		5: "payroll",
		6: "tds",
		7: "payroll",
		8: "tds",
		9: "notification",
	}

	for stepNum, expectedService := range expectedServices {
		step := s.GetStepDefinition(stepNum)
		if step.ServiceName != expectedService {
			t.Errorf("form16 step %d: expected %s, got %s", stepNum, expectedService, step.ServiceName)
		}
	}
}

func TestPFESIRemittanceSagaServices(t *testing.T) {
	s := NewPFESIRemittanceSaga()
	expectedServices := map[int]string{
		1:  "attendance",
		2:  "payroll",
		3:  "payroll",
		4:  "payroll",
		5:  "payroll",
		6:  "payroll",
		7:  "payroll",
		8:  "general-ledger",
		9:  "banking",
		10: "payroll",
	}

	for stepNum, expectedService := range expectedServices {
		step := s.GetStepDefinition(stepNum)
		if step.ServiceName != expectedService {
			t.Errorf("pf_esi step %d: expected %s, got %s", stepNum, expectedService, step.ServiceName)
		}
	}
}

func TestTDSPaymentReturnFilingSagaServices(t *testing.T) {
	s := NewTDSPaymentReturnFilingSaga()
	expectedServices := map[int]string{
		1:  "tds",
		2:  "tds",
		3:  "employee",
		4:  "tds",
		5:  "banking",
		6:  "tds",
		7:  "tds",
		8:  "tds",
		9:  "tds",
		10: "compliance-postings",
		11: "tds",
	}

	for stepNum, expectedService := range expectedServices {
		step := s.GetStepDefinition(stepNum)
		if step.ServiceName != expectedService {
			t.Errorf("tds_filing step %d: expected %s, got %s", stepNum, expectedService, step.ServiceName)
		}
	}
}

func TestLeaveEncashmentSettlementSagaServices(t *testing.T) {
	s := NewLeaveEncashmentSettlementSaga()
	expectedServices := map[int]string{
		1: "employee",
		2: "leave",
		3: "payroll",
		4: "payroll",
		5: "payroll",
		6: "accounts-payable",
		7: "general-ledger",
		8: "approval",
	}

	for stepNum, expectedService := range expectedServices {
		step := s.GetStepDefinition(stepNum)
		if step.ServiceName != expectedService {
			t.Errorf("leave_settlement step %d: expected %s, got %s", stepNum, expectedService, step.ServiceName)
		}
	}
}
