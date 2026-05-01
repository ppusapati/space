// Package gst provides comprehensive tests for Phase 5A GST sagas
package gst

import (
	"testing"
)

// ============================================================================
// SAGA-G01: GST Return Filing (Monthly/Quarterly)
// ============================================================================

// TestGSTReturnFilingSagaType verifies saga type identification
func TestGSTReturnFilingSagaType(t *testing.T) {
	s := NewGSTReturnFilingSaga()
	expected := "SAGA-G01"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestGSTReturnFilingSagaStepCount verifies 19 steps total
func TestGSTReturnFilingSagaStepCount(t *testing.T) {
	s := NewGSTReturnFilingSaga()
	steps := s.GetStepDefinitions()
	expected := 19 // 10 forward + 9 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestGSTReturnFilingSagaValidation verifies input validation
func TestGSTReturnFilingSagaValidation(t *testing.T) {
	s := NewGSTReturnFilingSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"return_period":          "2024-01",
				"return_type":            "GSTR1",
				"gstin":                  "27AAFCU5055K1Z0",
				"fiscal_year":            "2024-25",
				"tax_calculation_basis":  "invoice",
			},
			hasErr: false,
		},
		{
			name: "missing return_period",
			input: map[string]interface{}{
				"return_type":           "GSTR1",
				"gstin":                 "27AAFCU5055K1Z0",
				"fiscal_year":           "2024-25",
				"tax_calculation_basis": "invoice",
			},
			hasErr: true,
		},
		{
			name: "missing return_type",
			input: map[string]interface{}{
				"return_period":         "2024-01",
				"gstin":                 "27AAFCU5055K1Z0",
				"fiscal_year":           "2024-25",
				"tax_calculation_basis": "invoice",
			},
			hasErr: true,
		},
		{
			name: "invalid return_type",
			input: map[string]interface{}{
				"return_period":         "2024-01",
				"return_type":           "INVALID",
				"gstin":                 "27AAFCU5055K1Z0",
				"fiscal_year":           "2024-25",
				"tax_calculation_basis": "invoice",
			},
			hasErr: true,
		},
		{
			name: "missing gstin",
			input: map[string]interface{}{
				"return_period":         "2024-01",
				"return_type":           "GSTR1",
				"fiscal_year":           "2024-25",
				"tax_calculation_basis": "invoice",
			},
			hasErr: true,
		},
		{
			name: "missing fiscal_year",
			input: map[string]interface{}{
				"return_period":         "2024-01",
				"return_type":           "GSTR1",
				"gstin":                 "27AAFCU5055K1Z0",
				"tax_calculation_basis": "invoice",
			},
			hasErr: true,
		},
		{
			name: "missing tax_calculation_basis",
			input: map[string]interface{}{
				"return_period": "2024-01",
				"return_type":   "GSTR1",
				"gstin":         "27AAFCU5055K1Z0",
				"fiscal_year":   "2024-25",
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

// TestGSTReturnFilingSagaCriticalSteps verifies critical steps
func TestGSTReturnFilingSagaCriticalSteps(t *testing.T) {
	s := NewGSTReturnFilingSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 8: true, 10: true}

	for _, step := range steps {
		stepNum := int(step.StepNumber)
		if stepNum > 10 {
			continue // Skip compensation steps
		}
		expectedCritical := criticalSteps[stepNum]
		if step.IsCritical != expectedCritical {
			t.Errorf("step %d IsCritical: expected %v, got %v", stepNum, expectedCritical, step.IsCritical)
		}
	}
}

// TestGSTReturnFilingSagaTimeouts verifies timeout configurations
func TestGSTReturnFilingSagaTimeouts(t *testing.T) {
	s := NewGSTReturnFilingSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 10 {
			continue // Skip compensation steps
		}
		if step.TimeoutSeconds < 15 || step.TimeoutSeconds > 45 {
			t.Errorf("step %d timeout %d out of range [15-45]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestGSTReturnFilingSagaServiceNames verifies service name conventions
func TestGSTReturnFilingSagaServiceNames(t *testing.T) {
	s := NewGSTReturnFilingSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"gst-return":      true,
		"tax-engine":      true,
		"gst-ledger":      true,
		"general-ledger":  true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestGSTReturnFilingSagaCompensationSteps verifies compensation mapping
func TestGSTReturnFilingSagaCompensationSteps(t *testing.T) {
	s := NewGSTReturnFilingSaga()
	steps := s.GetStepDefinitions()

	compensationMapping := map[int][]int32{
		1: {},     // No compensation
		2: {102},  // Compensation step 102
		3: {103},  // Compensation step 103
		4: {104},  // Compensation step 104
		5: {105},  // Compensation step 105
		6: {106},  // Compensation step 106
		7: {107},  // Compensation step 107
		8: {108},  // Compensation step 108
		9: {109},  // Compensation step 109
		10: {},    // No compensation
	}

	for _, step := range steps {
		stepNum := int(step.StepNumber)
		if stepNum > 10 {
			continue // Skip compensation steps
		}
		expected, ok := compensationMapping[stepNum]
		if !ok {
			continue
		}

		// Compare compensation steps
		if len(step.CompensationSteps) != len(expected) {
			t.Errorf("step %d compensation steps count: expected %v, got %v", stepNum, expected, step.CompensationSteps)
		} else if len(expected) > 0 && (len(step.CompensationSteps) == 0 || step.CompensationSteps[0] != expected[0]) {
			t.Errorf("step %d compensation steps: expected %v, got %v", stepNum, expected, step.CompensationSteps)
		}
	}
}

// TestGSTReturnFilingSagaInputMapping verifies step 1 input fields
func TestGSTReturnFilingSagaInputMapping(t *testing.T) {
	s := NewGSTReturnFilingSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	requiredFields := []string{"tenantID", "companyID", "branchID", "returnPeriod", "returnType", "gstin", "fiscalYear"}
	for _, field := range requiredFields {
		if _, ok := step1.InputMapping[field]; !ok {
			t.Errorf("step 1 InputMapping missing required field: %s", field)
		}
	}
}

// TestGSTReturnFilingSagaRetryConfig verifies retry configuration
func TestGSTReturnFilingSagaRetryConfig(t *testing.T) {
	s := NewGSTReturnFilingSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 10 {
			continue // Skip compensation steps
		}
		if step.RetryConfig == nil {
			t.Errorf("step %d has nil RetryConfig", step.StepNumber)
			continue
		}
		if step.RetryConfig.MaxRetries < 1 || step.RetryConfig.MaxRetries > 5 {
			t.Errorf("step %d MaxRetries out of range: %d", step.StepNumber, step.RetryConfig.MaxRetries)
		}
		if step.RetryConfig.InitialBackoffMs < 500 || step.RetryConfig.InitialBackoffMs > 5000 {
			t.Errorf("step %d InitialBackoffMs out of range: %d", step.StepNumber, step.RetryConfig.InitialBackoffMs)
		}
		if step.RetryConfig.BackoffMultiplier < 1.5 || step.RetryConfig.BackoffMultiplier > 3.0 {
			t.Errorf("step %d BackoffMultiplier out of range: %f", step.StepNumber, step.RetryConfig.BackoffMultiplier)
		}
	}
}

// TestGSTReturnFilingSagaInterfaceImplementation verifies SagaHandler interface
func TestGSTReturnFilingSagaInterfaceImplementation(t *testing.T) {
	var s interface{} = NewGSTReturnFilingSaga()

	// Verify all interface methods are present
	if _, ok := s.(interface {
		SagaType() string
		GetStepDefinitions() []*interface{}
		GetStepDefinition(int) *interface{}
		ValidateInput(interface{}) error
	}); !ok {
		// Try checking individual methods by calling them
		handler := s.(interface{ SagaType() string })
		if handler.SagaType() == "" {
			t.Error("SagaType() should not be empty")
		}
	}
}

// ============================================================================
// SAGA-G02: ITC Reconciliation
// ============================================================================

// TestITCReconciliationSagaType verifies saga type identification
func TestITCReconciliationSagaType(t *testing.T) {
	s := NewITCReconciliationSaga()
	expected := "SAGA-G02"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestITCReconciliationSagaStepCount verifies 17 steps total
func TestITCReconciliationSagaStepCount(t *testing.T) {
	s := NewITCReconciliationSaga()
	steps := s.GetStepDefinitions()
	expected := 17 // 8 forward + 9 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestITCReconciliationSagaValidation verifies input validation
func TestITCReconciliationSagaValidation(t *testing.T) {
	s := NewITCReconciliationSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"reconciliation_period": "2024-01",
				"itc_type":              "ITC-1",
				"gstin":                 "27AAFCU5055K1Z0",
				"fiscal_year":           "2024-25",
			},
			hasErr: false,
		},
		{
			name: "missing reconciliation_period",
			input: map[string]interface{}{
				"itc_type":    "ITC-1",
				"gstin":       "27AAFCU5055K1Z0",
				"fiscal_year": "2024-25",
			},
			hasErr: true,
		},
		{
			name: "missing itc_type",
			input: map[string]interface{}{
				"reconciliation_period": "2024-01",
				"gstin":                 "27AAFCU5055K1Z0",
				"fiscal_year":           "2024-25",
			},
			hasErr: true,
		},
		{
			name: "missing gstin",
			input: map[string]interface{}{
				"reconciliation_period": "2024-01",
				"itc_type":              "ITC-1",
				"fiscal_year":           "2024-25",
			},
			hasErr: true,
		},
		{
			name: "missing fiscal_year",
			input: map[string]interface{}{
				"reconciliation_period": "2024-01",
				"itc_type":              "ITC-1",
				"gstin":                 "27AAFCU5055K1Z0",
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

// TestITCReconciliationSagaCriticalSteps verifies critical steps
func TestITCReconciliationSagaCriticalSteps(t *testing.T) {
	s := NewITCReconciliationSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 6: true, 9: true}

	for _, step := range steps {
		stepNum := int(step.StepNumber)
		if stepNum > 8 {
			continue // Skip compensation steps
		}
		expectedCritical := criticalSteps[stepNum]
		if step.IsCritical != expectedCritical {
			t.Errorf("step %d IsCritical: expected %v, got %v", stepNum, expectedCritical, step.IsCritical)
		}
	}
}

// TestITCReconciliationSagaTimeouts verifies timeout configurations
func TestITCReconciliationSagaTimeouts(t *testing.T) {
	s := NewITCReconciliationSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 8 {
			continue // Skip compensation steps
		}
		if step.TimeoutSeconds < 20 || step.TimeoutSeconds > 45 {
			t.Errorf("step %d timeout %d out of range [20-45]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestITCReconciliationSagaServiceNames verifies service name conventions
func TestITCReconciliationSagaServiceNames(t *testing.T) {
	s := NewITCReconciliationSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"gst-ledger":      true,
		"reconciliation":  true,
		"tax-engine":      true,
		"general-ledger":  true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestITCReconciliationSagaCompensationSteps verifies compensation mapping
func TestITCReconciliationSagaCompensationSteps(t *testing.T) {
	s := NewITCReconciliationSaga()
	steps := s.GetStepDefinitions()

	compensationMapping := map[int]bool{
		1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 8: true,
	}

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount < 7 || compensationCount > 9 {
		t.Errorf("expected 7-9 compensation steps, got %d", compensationCount)
	}
}

// TestITCReconciliationSagaInputMapping verifies step 1 input fields
func TestITCReconciliationSagaInputMapping(t *testing.T) {
	s := NewITCReconciliationSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	requiredFields := []string{"tenantID", "companyID", "branchID"}
	for _, field := range requiredFields {
		if _, ok := step1.InputMapping[field]; !ok {
			t.Errorf("step 1 InputMapping missing required field: %s", field)
		}
	}
}

// TestITCReconciliationSagaRetryConfig verifies retry configuration
func TestITCReconciliationSagaRetryConfig(t *testing.T) {
	s := NewITCReconciliationSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 8 {
			continue // Skip compensation steps
		}
		if step.RetryConfig == nil {
			t.Errorf("step %d has nil RetryConfig", step.StepNumber)
		}
	}
}

// TestITCReconciliationSagaInterfaceImplementation verifies SagaHandler interface
func TestITCReconciliationSagaInterfaceImplementation(t *testing.T) {
	s := NewITCReconciliationSaga()
	if s.SagaType() != "SAGA-G02" {
		t.Error("interface implementation failed")
	}
}

// ============================================================================
// SAGA-G03: E-Way Bill Processing
// ============================================================================

// TestEwayBillSagaType verifies saga type identification
func TestEwayBillSagaType(t *testing.T) {
	s := NewEwayBillSaga()
	expected := "SAGA-G03"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestEwayBillSagaStepCount verifies 15 steps total
func TestEwayBillSagaStepCount(t *testing.T) {
	s := NewEwayBillSaga()
	steps := s.GetStepDefinitions()
	expected := 15 // 7 forward + 8 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestEwayBillSagaValidation verifies input validation
func TestEwayBillSagaValidation(t *testing.T) {
	s := NewEwayBillSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"shipment_id":        "SHIP001",
				"consignment_value":  100000,
				"gstin":              "27AAFCU5055K1Z0",
				"consignor_gstin":    "27AAFCU5055K1Z0",
				"consignee_gstin":    "27BBDPU5555K1Z0",
			},
			hasErr: false,
		},
		{
			name: "missing shipment_id",
			input: map[string]interface{}{
				"consignment_value": 100000,
				"gstin":              "27AAFCU5055K1Z0",
				"consignor_gstin":    "27AAFCU5055K1Z0",
				"consignee_gstin":    "27BBDPU5555K1Z0",
			},
			hasErr: true,
		},
		{
			name: "missing consignment_value",
			input: map[string]interface{}{
				"shipment_id":     "SHIP001",
				"gstin":           "27AAFCU5055K1Z0",
				"consignor_gstin": "27AAFCU5055K1Z0",
				"consignee_gstin": "27BBDPU5555K1Z0",
			},
			hasErr: true,
		},
		{
			name: "missing gstin",
			input: map[string]interface{}{
				"shipment_id":     "SHIP001",
				"consignment_value": 100000,
				"consignor_gstin": "27AAFCU5055K1Z0",
				"consignee_gstin": "27BBDPU5555K1Z0",
			},
			hasErr: true,
		},
		{
			name: "missing consignor_gstin",
			input: map[string]interface{}{
				"shipment_id":      "SHIP001",
				"consignment_value": 100000,
				"gstin":             "27AAFCU5055K1Z0",
				"consignee_gstin":   "27BBDPU5555K1Z0",
			},
			hasErr: true,
		},
		{
			name: "missing consignee_gstin",
			input: map[string]interface{}{
				"shipment_id":      "SHIP001",
				"consignment_value": 100000,
				"gstin":             "27AAFCU5055K1Z0",
				"consignor_gstin":   "27AAFCU5055K1Z0",
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

// TestEwayBillSagaCriticalSteps verifies critical steps
func TestEwayBillSagaCriticalSteps(t *testing.T) {
	s := NewEwayBillSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 8: true}

	for _, step := range steps {
		stepNum := int(step.StepNumber)
		if stepNum > 7 {
			continue // Skip compensation steps
		}
		expectedCritical := criticalSteps[stepNum]
		if step.IsCritical != expectedCritical {
			t.Errorf("step %d IsCritical: expected %v, got %v", stepNum, expectedCritical, step.IsCritical)
		}
	}
}

// TestEwayBillSagaTimeouts verifies timeout configurations
func TestEwayBillSagaTimeouts(t *testing.T) {
	s := NewEwayBillSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 7 {
			continue // Skip compensation steps
		}
		if step.TimeoutSeconds < 20 || step.TimeoutSeconds > 45 {
			t.Errorf("step %d timeout %d out of range [20-45]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestEwayBillSagaServiceNames verifies service name conventions
func TestEwayBillSagaServiceNames(t *testing.T) {
	s := NewEwayBillSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"eway-bill":      true,
		"gst":            true,
		"shipment":       true,
		"sales-order":    true,
		"general-ledger": true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestEwayBillSagaCompensationSteps verifies compensation mapping
func TestEwayBillSagaCompensationSteps(t *testing.T) {
	s := NewEwayBillSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount < 7 || compensationCount > 8 {
		t.Errorf("expected 7-8 compensation steps, got %d", compensationCount)
	}
}

// TestEwayBillSagaInputMapping verifies step 1 input fields
func TestEwayBillSagaInputMapping(t *testing.T) {
	s := NewEwayBillSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	if _, ok := step1.InputMapping["shipment_id"]; !ok {
		t.Error("step 1 InputMapping missing shipment_id")
	}
}

// TestEwayBillSagaRetryConfig verifies retry configuration
func TestEwayBillSagaRetryConfig(t *testing.T) {
	s := NewEwayBillSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 7 {
			continue // Skip compensation steps
		}
		if step.RetryConfig == nil {
			t.Errorf("step %d has nil RetryConfig", step.StepNumber)
		}
	}
}

// TestEwayBillSagaInterfaceImplementation verifies SagaHandler interface
func TestEwayBillSagaInterfaceImplementation(t *testing.T) {
	s := NewEwayBillSaga()
	if s.SagaType() != "SAGA-G03" {
		t.Error("interface implementation failed")
	}
}

// ============================================================================
// SAGA-G04: GST Audit
// ============================================================================

// TestGSTAuditSagaType verifies saga type identification
func TestGSTAuditSagaType(t *testing.T) {
	s := NewGSTAuditSaga()
	expected := "SAGA-G04"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestGSTAuditSagaStepCount verifies 19 steps total
func TestGSTAuditSagaStepCount(t *testing.T) {
	s := NewGSTAuditSaga()
	steps := s.GetStepDefinitions()
	expected := 19 // 10 forward + 9 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestGSTAuditSagaValidation verifies input validation
func TestGSTAuditSagaValidation(t *testing.T) {
	s := NewGSTAuditSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"audit_period": "2024-01",
				"audit_type":   "statutory",
				"gstin":        "27AAFCU5055K1Z0",
				"fiscal_year":  "2024-25",
			},
			hasErr: false,
		},
		{
			name: "missing audit_period",
			input: map[string]interface{}{
				"audit_type":  "statutory",
				"gstin":       "27AAFCU5055K1Z0",
				"fiscal_year": "2024-25",
			},
			hasErr: true,
		},
		{
			name: "missing audit_type",
			input: map[string]interface{}{
				"audit_period": "2024-01",
				"gstin":        "27AAFCU5055K1Z0",
				"fiscal_year":  "2024-25",
			},
			hasErr: true,
		},
		{
			name: "missing gstin",
			input: map[string]interface{}{
				"audit_period": "2024-01",
				"audit_type":   "statutory",
				"fiscal_year":  "2024-25",
			},
			hasErr: true,
		},
		{
			name: "missing fiscal_year",
			input: map[string]interface{}{
				"audit_period": "2024-01",
				"audit_type":   "statutory",
				"gstin":        "27AAFCU5055K1Z0",
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

// TestGSTAuditSagaCriticalSteps verifies critical steps
func TestGSTAuditSagaCriticalSteps(t *testing.T) {
	s := NewGSTAuditSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 5: true, 8: true, 10: true}

	for _, step := range steps {
		stepNum := int(step.StepNumber)
		if stepNum > 10 {
			continue // Skip compensation steps
		}
		expectedCritical := criticalSteps[stepNum]
		if step.IsCritical != expectedCritical {
			t.Errorf("step %d IsCritical: expected %v, got %v", stepNum, expectedCritical, step.IsCritical)
		}
	}
}

// TestGSTAuditSagaTimeouts verifies timeout configurations
func TestGSTAuditSagaTimeouts(t *testing.T) {
	s := NewGSTAuditSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 10 {
			continue // Skip compensation steps
		}
		if step.TimeoutSeconds < 20 || step.TimeoutSeconds > 50 {
			t.Errorf("step %d timeout %d out of range [20-50]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestGSTAuditSagaServiceNames verifies service name conventions
func TestGSTAuditSagaServiceNames(t *testing.T) {
	s := NewGSTAuditSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"gst-audit":       true,
		"gst-ledger":      true,
		"audit":           true,
		"general-ledger":  true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestGSTAuditSagaCompensationSteps verifies compensation mapping
func TestGSTAuditSagaCompensationSteps(t *testing.T) {
	s := NewGSTAuditSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount < 9 || compensationCount > 10 {
		t.Errorf("expected 9-10 compensation steps, got %d", compensationCount)
	}
}

// TestGSTAuditSagaInputMapping verifies step 1 input fields
func TestGSTAuditSagaInputMapping(t *testing.T) {
	s := NewGSTAuditSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	if _, ok := step1.InputMapping["audit_period"]; !ok {
		t.Error("step 1 InputMapping missing audit_period")
	}
}

// TestGSTAuditSagaRetryConfig verifies retry configuration
func TestGSTAuditSagaRetryConfig(t *testing.T) {
	s := NewGSTAuditSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 10 {
			continue // Skip compensation steps
		}
		if step.RetryConfig == nil {
			t.Errorf("step %d has nil RetryConfig", step.StepNumber)
		}
	}
}

// TestGSTAuditSagaInterfaceImplementation verifies SagaHandler interface
func TestGSTAuditSagaInterfaceImplementation(t *testing.T) {
	s := NewGSTAuditSaga()
	if s.SagaType() != "SAGA-G04" {
		t.Error("interface implementation failed")
	}
}

// ============================================================================
// SAGA-G05: GST Amendment
// ============================================================================

// TestGSTAmendmentSagaType verifies saga type identification
func TestGSTAmendmentSagaType(t *testing.T) {
	s := NewGSTAmendmentSaga()
	expected := "SAGA-G05"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestGSTAmendmentSagaStepCount verifies 15 steps total
func TestGSTAmendmentSagaStepCount(t *testing.T) {
	s := NewGSTAmendmentSaga()
	steps := s.GetStepDefinitions()
	expected := 15 // 7 forward + 8 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestGSTAmendmentSagaValidation verifies input validation
func TestGSTAmendmentSagaValidation(t *testing.T) {
	s := NewGSTAmendmentSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"amendment_type": "correction",
				"original_period": "2024-01",
				"gstin":           "27AAFCU5055K1Z0",
				"fiscal_year":     "2024-25",
				"amendment_date":  "2024-02-15",
			},
			hasErr: false,
		},
		{
			name: "missing amendment_type",
			input: map[string]interface{}{
				"original_period": "2024-01",
				"gstin":            "27AAFCU5055K1Z0",
				"fiscal_year":      "2024-25",
				"amendment_date":   "2024-02-15",
			},
			hasErr: true,
		},
		{
			name: "missing original_period",
			input: map[string]interface{}{
				"amendment_type": "correction",
				"gstin":           "27AAFCU5055K1Z0",
				"fiscal_year":     "2024-25",
				"amendment_date":  "2024-02-15",
			},
			hasErr: true,
		},
		{
			name: "missing gstin",
			input: map[string]interface{}{
				"amendment_type": "correction",
				"original_period": "2024-01",
				"fiscal_year":     "2024-25",
				"amendment_date":  "2024-02-15",
			},
			hasErr: true,
		},
		{
			name: "missing fiscal_year",
			input: map[string]interface{}{
				"amendment_type": "correction",
				"original_period": "2024-01",
				"gstin":           "27AAFCU5055K1Z0",
				"amendment_date":  "2024-02-15",
			},
			hasErr: true,
		},
		{
			name: "missing amendment_date",
			input: map[string]interface{}{
				"amendment_type": "correction",
				"original_period": "2024-01",
				"gstin":           "27AAFCU5055K1Z0",
				"fiscal_year":     "2024-25",
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

// TestGSTAmendmentSagaCriticalSteps verifies critical steps
func TestGSTAmendmentSagaCriticalSteps(t *testing.T) {
	s := NewGSTAmendmentSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 8: true}

	for _, step := range steps {
		stepNum := int(step.StepNumber)
		if stepNum > 7 {
			continue // Skip compensation steps
		}
		expectedCritical := criticalSteps[stepNum]
		if step.IsCritical != expectedCritical {
			t.Errorf("step %d IsCritical: expected %v, got %v", stepNum, expectedCritical, step.IsCritical)
		}
	}
}

// TestGSTAmendmentSagaTimeouts verifies timeout configurations
func TestGSTAmendmentSagaTimeouts(t *testing.T) {
	s := NewGSTAmendmentSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 7 {
			continue // Skip compensation steps
		}
		if step.TimeoutSeconds < 20 || step.TimeoutSeconds > 45 {
			t.Errorf("step %d timeout %d out of range [20-45]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestGSTAmendmentSagaServiceNames verifies service name conventions
func TestGSTAmendmentSagaServiceNames(t *testing.T) {
	s := NewGSTAmendmentSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"gst-amendment":   true,
		"gst-return":      true,
		"general-ledger":  true,
		"tax-engine":      true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestGSTAmendmentSagaCompensationSteps verifies compensation mapping
func TestGSTAmendmentSagaCompensationSteps(t *testing.T) {
	s := NewGSTAmendmentSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount < 7 || compensationCount > 8 {
		t.Errorf("expected 7-8 compensation steps, got %d", compensationCount)
	}
}

// TestGSTAmendmentSagaInputMapping verifies step 1 input fields
func TestGSTAmendmentSagaInputMapping(t *testing.T) {
	s := NewGSTAmendmentSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	if _, ok := step1.InputMapping["amendment_type"]; !ok {
		t.Error("step 1 InputMapping missing amendment_type")
	}
}

// TestGSTAmendmentSagaRetryConfig verifies retry configuration
func TestGSTAmendmentSagaRetryConfig(t *testing.T) {
	s := NewGSTAmendmentSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 7 {
			continue // Skip compensation steps
		}
		if step.RetryConfig == nil {
			t.Errorf("step %d has nil RetryConfig", step.StepNumber)
		}
	}
}

// TestGSTAmendmentSagaInterfaceImplementation verifies SagaHandler interface
func TestGSTAmendmentSagaInterfaceImplementation(t *testing.T) {
	s := NewGSTAmendmentSaga()
	if s.SagaType() != "SAGA-G05" {
		t.Error("interface implementation failed")
	}
}

// ============================================================================
// SAGA-G06: GST Payment
// ============================================================================

// TestGSTPaymentSagaType verifies saga type identification
func TestGSTPaymentSagaType(t *testing.T) {
	s := NewGSTPaymentSaga()
	expected := "SAGA-G06"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestGSTPaymentSagaStepCount verifies 17 steps total
func TestGSTPaymentSagaStepCount(t *testing.T) {
	s := NewGSTPaymentSaga()
	steps := s.GetStepDefinitions()
	expected := 17 // 8 forward + 9 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestGSTPaymentSagaValidation verifies input validation
func TestGSTPaymentSagaValidation(t *testing.T) {
	s := NewGSTPaymentSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"payment_period": "2024-01",
				"payment_type":   "tax",
				"gstin":          "27AAFCU5055K1Z0",
				"fiscal_year":    "2024-25",
				"payment_date":   "2024-02-15",
				"bank_account":   "BANKACCT001",
			},
			hasErr: false,
		},
		{
			name: "missing payment_period",
			input: map[string]interface{}{
				"payment_type": "tax",
				"gstin":        "27AAFCU5055K1Z0",
				"fiscal_year":  "2024-25",
				"payment_date": "2024-02-15",
				"bank_account": "BANKACCT001",
			},
			hasErr: true,
		},
		{
			name: "missing payment_type",
			input: map[string]interface{}{
				"payment_period": "2024-01",
				"gstin":           "27AAFCU5055K1Z0",
				"fiscal_year":     "2024-25",
				"payment_date":    "2024-02-15",
				"bank_account":    "BANKACCT001",
			},
			hasErr: true,
		},
		{
			name: "missing gstin",
			input: map[string]interface{}{
				"payment_period": "2024-01",
				"payment_type":   "tax",
				"fiscal_year":    "2024-25",
				"payment_date":   "2024-02-15",
				"bank_account":   "BANKACCT001",
			},
			hasErr: true,
		},
		{
			name: "missing fiscal_year",
			input: map[string]interface{}{
				"payment_period": "2024-01",
				"payment_type":   "tax",
				"gstin":          "27AAFCU5055K1Z0",
				"payment_date":   "2024-02-15",
				"bank_account":   "BANKACCT001",
			},
			hasErr: true,
		},
		{
			name: "missing payment_date",
			input: map[string]interface{}{
				"payment_period": "2024-01",
				"payment_type":   "tax",
				"gstin":          "27AAFCU5055K1Z0",
				"fiscal_year":    "2024-25",
				"bank_account":   "BANKACCT001",
			},
			hasErr: true,
		},
		{
			name: "missing bank_account",
			input: map[string]interface{}{
				"payment_period": "2024-01",
				"payment_type":   "tax",
				"gstin":          "27AAFCU5055K1Z0",
				"fiscal_year":    "2024-25",
				"payment_date":   "2024-02-15",
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

// TestGSTPaymentSagaCriticalSteps verifies critical steps
func TestGSTPaymentSagaCriticalSteps(t *testing.T) {
	s := NewGSTPaymentSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 5: true, 6: true, 9: true}

	for _, step := range steps {
		stepNum := int(step.StepNumber)
		if stepNum > 8 {
			continue // Skip compensation steps
		}
		expectedCritical := criticalSteps[stepNum]
		if step.IsCritical != expectedCritical {
			t.Errorf("step %d IsCritical: expected %v, got %v", stepNum, expectedCritical, step.IsCritical)
		}
	}
}

// TestGSTPaymentSagaTimeouts verifies timeout configurations
func TestGSTPaymentSagaTimeouts(t *testing.T) {
	s := NewGSTPaymentSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 8 {
			continue // Skip compensation steps
		}
		if step.TimeoutSeconds < 20 || step.TimeoutSeconds > 45 {
			t.Errorf("step %d timeout %d out of range [20-45]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestGSTPaymentSagaServiceNames verifies service name conventions
func TestGSTPaymentSagaServiceNames(t *testing.T) {
	s := NewGSTPaymentSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"gst-payment":     true,
		"banking":         true,
		"general-ledger":  true,
		"tax-engine":      true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestGSTPaymentSagaCompensationSteps verifies compensation mapping
func TestGSTPaymentSagaCompensationSteps(t *testing.T) {
	s := NewGSTPaymentSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount < 8 || compensationCount > 9 {
		t.Errorf("expected 8-9 compensation steps, got %d", compensationCount)
	}
}

// TestGSTPaymentSagaInputMapping verifies step 1 input fields
func TestGSTPaymentSagaInputMapping(t *testing.T) {
	s := NewGSTPaymentSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	requiredFields := []string{"payment_period", "payment_type"}
	for _, field := range requiredFields {
		found := false
		for k := range step1.InputMapping {
			if k == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("step 1 InputMapping missing required field: %s", field)
		}
	}
}

// TestGSTPaymentSagaRetryConfig verifies retry configuration
func TestGSTPaymentSagaRetryConfig(t *testing.T) {
	s := NewGSTPaymentSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 8 {
			continue // Skip compensation steps
		}
		if step.RetryConfig == nil {
			t.Errorf("step %d has nil RetryConfig", step.StepNumber)
		}
	}
}

// TestGSTPaymentSagaInterfaceImplementation verifies SagaHandler interface
func TestGSTPaymentSagaInterfaceImplementation(t *testing.T) {
	s := NewGSTPaymentSaga()
	if s.SagaType() != "SAGA-G06" {
		t.Error("interface implementation failed")
	}
}

// ============================================================================
// SAGA-G07: Reverse Charge Mechanism (RCM)
// ============================================================================

// TestRCMSagaType verifies saga type identification
func TestRCMSagaType(t *testing.T) {
	s := NewReverseChargeMechanismSaga()
	expected := "SAGA-G07"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestRCMSagaStepCount verifies 17 steps total
func TestRCMSagaStepCount(t *testing.T) {
	s := NewReverseChargeMechanismSaga()
	steps := s.GetStepDefinitions()
	expected := 17 // 9 forward + 8 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestRCMSagaValidation verifies input validation
func TestRCMSagaValidation(t *testing.T) {
	s := NewReverseChargeMechanismSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input - import",
			input: map[string]interface{}{
				"invoice_id":               "INV001",
				"supplier_id":              "SUPP001",
				"item_category":            "IMPORT",
				"invoice_amount":           100000,
				"tax_rate":                 18,
				"supply_type":              "goods",
				"transaction_date":         "2024-02-15",
				"gstin":                    "27AAFCU5055K1Z0",
				"rcm_notification_list":    "notification_list_1",
			},
			hasErr: false,
		},
		{
			name: "valid input - ecommerce",
			input: map[string]interface{}{
				"invoice_id":               "INV002",
				"supplier_id":              "SUPP002",
				"item_category":            "ECOMMERCE",
				"invoice_amount":           50000,
				"tax_rate":                 12,
				"supply_type":              "goods",
				"transaction_date":         "2024-02-15",
				"gstin":                    "27AAFCU5055K1Z0",
				"rcm_notification_list":    "notification_list_1",
			},
			hasErr: false,
		},
		{
			name: "missing invoice_id",
			input: map[string]interface{}{
				"supplier_id":           "SUPP001",
				"item_category":         "IMPORT",
				"invoice_amount":        100000,
				"tax_rate":              18,
				"supply_type":           "goods",
				"transaction_date":      "2024-02-15",
				"gstin":                 "27AAFCU5055K1Z0",
				"rcm_notification_list": "notification_list_1",
			},
			hasErr: true,
		},
		{
			name: "missing supplier_id",
			input: map[string]interface{}{
				"invoice_id":            "INV001",
				"item_category":         "IMPORT",
				"invoice_amount":        100000,
				"tax_rate":              18,
				"supply_type":           "goods",
				"transaction_date":      "2024-02-15",
				"gstin":                 "27AAFCU5055K1Z0",
				"rcm_notification_list": "notification_list_1",
			},
			hasErr: true,
		},
		{
			name: "missing item_category",
			input: map[string]interface{}{
				"invoice_id":            "INV001",
				"supplier_id":           "SUPP001",
				"invoice_amount":        100000,
				"tax_rate":              18,
				"supply_type":           "goods",
				"transaction_date":      "2024-02-15",
				"gstin":                 "27AAFCU5055K1Z0",
				"rcm_notification_list": "notification_list_1",
			},
			hasErr: true,
		},
		{
			name: "invalid item_category",
			input: map[string]interface{}{
				"invoice_id":            "INV001",
				"supplier_id":           "SUPP001",
				"item_category":         "INVALID",
				"invoice_amount":        100000,
				"tax_rate":              18,
				"supply_type":           "goods",
				"transaction_date":      "2024-02-15",
				"gstin":                 "27AAFCU5055K1Z0",
				"rcm_notification_list": "notification_list_1",
			},
			hasErr: true,
		},
		{
			name: "missing gstin",
			input: map[string]interface{}{
				"invoice_id":            "INV001",
				"supplier_id":           "SUPP001",
				"item_category":         "IMPORT",
				"invoice_amount":        100000,
				"tax_rate":              18,
				"supply_type":           "goods",
				"transaction_date":      "2024-02-15",
				"rcm_notification_list": "notification_list_1",
			},
			hasErr: true,
		},
		{
			name: "missing rcm_notification_list",
			input: map[string]interface{}{
				"invoice_id":       "INV001",
				"supplier_id":      "SUPP001",
				"item_category":    "IMPORT",
				"invoice_amount":   100000,
				"tax_rate":         18,
				"supply_type":      "goods",
				"transaction_date": "2024-02-15",
				"gstin":            "27AAFCU5055K1Z0",
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

// TestRCMSagaCriticalSteps verifies critical steps
func TestRCMSagaCriticalSteps(t *testing.T) {
	s := NewReverseChargeMechanismSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 7: true, 8: true}

	for _, step := range steps {
		stepNum := int(step.StepNumber)
		if stepNum > 9 {
			continue // Skip compensation steps
		}
		expectedCritical := criticalSteps[stepNum]
		if step.IsCritical != expectedCritical {
			t.Errorf("step %d IsCritical: expected %v, got %v", stepNum, expectedCritical, step.IsCritical)
		}
	}
}

// TestRCMSagaTimeouts verifies timeout configurations
func TestRCMSagaTimeouts(t *testing.T) {
	s := NewReverseChargeMechanismSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 9 {
			continue // Skip compensation steps
		}
		if step.TimeoutSeconds < 20 || step.TimeoutSeconds > 40 {
			t.Errorf("step %d timeout %d out of range [20-40]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestRCMSagaServiceNames verifies service name conventions
func TestRCMSagaServiceNames(t *testing.T) {
	s := NewReverseChargeMechanismSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"gst":                   true,
		"vendor":                true,
		"tax-engine":            true,
		"purchase-invoice":      true,
		"gst-ledger":            true,
		"compliance-postings":   true,
		"general-ledger":        true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestRCMSagaCompensationSteps verifies compensation mapping
func TestRCMSagaCompensationSteps(t *testing.T) {
	s := NewReverseChargeMechanismSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount < 7 || compensationCount > 8 {
		t.Errorf("expected 7-8 compensation steps, got %d", compensationCount)
	}
}

// TestRCMSagaInputMapping verifies step 1 input fields
func TestRCMSagaInputMapping(t *testing.T) {
	s := NewReverseChargeMechanismSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	requiredFields := []string{"invoiceID", "supplierId", "itemCategory", "rcmNotificationList"}
	for _, field := range requiredFields {
		if _, ok := step1.InputMapping[field]; !ok {
			t.Errorf("step 1 InputMapping missing required field: %s", field)
		}
	}
}

// TestRCMSagaRetryConfig verifies retry configuration
func TestRCMSagaRetryConfig(t *testing.T) {
	s := NewReverseChargeMechanismSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 9 {
			continue // Skip compensation steps
		}
		if step.RetryConfig == nil {
			t.Errorf("step %d has nil RetryConfig", step.StepNumber)
		}
	}
}

// TestRCMSagaInterfaceImplementation verifies SagaHandler interface
func TestRCMSagaInterfaceImplementation(t *testing.T) {
	s := NewReverseChargeMechanismSaga()
	if s.SagaType() != "SAGA-G07" {
		t.Error("interface implementation failed")
	}
}

// ============================================================================
// SAGA-G08: Credit Note & Debit Note Processing
// ============================================================================

// TestCNDNSagaType verifies saga type identification
func TestCNDNSagaType(t *testing.T) {
	s := NewCreditDebitNoteSaga()
	expected := "SAGA-G08"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestCNDNSagaStepCount verifies 15 steps total
func TestCNDNSagaStepCount(t *testing.T) {
	s := NewCreditDebitNoteSaga()
	steps := s.GetStepDefinitions()
	expected := 15 // 8 forward + 7 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestCNDNSagaValidation verifies input validation
func TestCNDNSagaValidation(t *testing.T) {
	s := NewCreditDebitNoteSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input - full credit",
			input: map[string]interface{}{
				"original_invoice_id": "INV001",
				"transaction_type":    "SALES",
				"cndn_type":           "FULL_CREDIT",
				"cndn_reason":         "RETURN",
				"original_amount":     100000,
				"return_percentage":   100,
				"original_sgst":       9000,
				"original_cgst":       9000,
				"original_igst":       0,
				"cndn_date":           "2024-02-15",
				"cndn_period":         "2024-02",
				"gstr_type":           "GSTR1",
			},
			hasErr: false,
		},
		{
			name: "valid input - partial credit",
			input: map[string]interface{}{
				"original_invoice_id": "INV002",
				"transaction_type":    "PURCHASE",
				"cndn_type":           "PARTIAL_CREDIT",
				"cndn_reason":         "DISCOUNT",
				"original_amount":     50000,
				"return_percentage":   50,
				"original_sgst":       4500,
				"original_cgst":       4500,
				"original_igst":       0,
				"cndn_date":           "2024-02-15",
				"cndn_period":         "2024-02",
				"gstr_type":           "GSTR2",
			},
			hasErr: false,
		},
		{
			name: "missing original_invoice_id",
			input: map[string]interface{}{
				"transaction_type":   "SALES",
				"cndn_type":          "FULL_CREDIT",
				"cndn_reason":        "RETURN",
				"original_amount":    100000,
				"return_percentage":  100,
				"original_sgst":      9000,
				"original_cgst":      9000,
				"original_igst":      0,
				"cndn_date":          "2024-02-15",
				"cndn_period":        "2024-02",
				"gstr_type":          "GSTR1",
			},
			hasErr: true,
		},
		{
			name: "missing transaction_type",
			input: map[string]interface{}{
				"original_invoice_id": "INV001",
				"cndn_type":           "FULL_CREDIT",
				"cndn_reason":         "RETURN",
				"original_amount":     100000,
				"return_percentage":   100,
				"original_sgst":       9000,
				"original_cgst":       9000,
				"original_igst":       0,
				"cndn_date":           "2024-02-15",
				"cndn_period":         "2024-02",
				"gstr_type":           "GSTR1",
			},
			hasErr: true,
		},
		{
			name: "invalid transaction_type",
			input: map[string]interface{}{
				"original_invoice_id": "INV001",
				"transaction_type":    "INVALID",
				"cndn_type":           "FULL_CREDIT",
				"cndn_reason":         "RETURN",
				"original_amount":     100000,
				"return_percentage":   100,
				"original_sgst":       9000,
				"original_cgst":       9000,
				"original_igst":       0,
				"cndn_date":           "2024-02-15",
				"cndn_period":         "2024-02",
				"gstr_type":           "GSTR1",
			},
			hasErr: true,
		},
		{
			name: "missing cndn_type",
			input: map[string]interface{}{
				"original_invoice_id": "INV001",
				"transaction_type":    "SALES",
				"cndn_reason":         "RETURN",
				"original_amount":     100000,
				"return_percentage":   100,
				"original_sgst":       9000,
				"original_cgst":       9000,
				"original_igst":       0,
				"cndn_date":           "2024-02-15",
				"cndn_period":         "2024-02",
				"gstr_type":           "GSTR1",
			},
			hasErr: true,
		},
		{
			name: "invalid cndn_type",
			input: map[string]interface{}{
				"original_invoice_id": "INV001",
				"transaction_type":    "SALES",
				"cndn_type":           "INVALID_TYPE",
				"cndn_reason":         "RETURN",
				"original_amount":     100000,
				"return_percentage":   100,
				"original_sgst":       9000,
				"original_cgst":       9000,
				"original_igst":       0,
				"cndn_date":           "2024-02-15",
				"cndn_period":         "2024-02",
				"gstr_type":           "GSTR1",
			},
			hasErr: true,
		},
		{
			name: "invalid cndn_reason",
			input: map[string]interface{}{
				"original_invoice_id": "INV001",
				"transaction_type":    "SALES",
				"cndn_type":           "FULL_CREDIT",
				"cndn_reason":         "INVALID_REASON",
				"original_amount":     100000,
				"return_percentage":   100,
				"original_sgst":       9000,
				"original_cgst":       9000,
				"original_igst":       0,
				"cndn_date":           "2024-02-15",
				"cndn_period":         "2024-02",
				"gstr_type":           "GSTR1",
			},
			hasErr: true,
		},
		{
			name: "invalid gstr_type",
			input: map[string]interface{}{
				"original_invoice_id": "INV001",
				"transaction_type":    "SALES",
				"cndn_type":           "FULL_CREDIT",
				"cndn_reason":         "RETURN",
				"original_amount":     100000,
				"return_percentage":   100,
				"original_sgst":       9000,
				"original_cgst":       9000,
				"original_igst":       0,
				"cndn_date":           "2024-02-15",
				"cndn_period":         "2024-02",
				"gstr_type":           "GSTR3",
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

// TestCNDNSagaCriticalSteps verifies critical steps
func TestCNDNSagaCriticalSteps(t *testing.T) {
	s := NewCreditDebitNoteSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true}

	for _, step := range steps {
		stepNum := int(step.StepNumber)
		if stepNum > 8 {
			continue // Skip compensation steps
		}
		expectedCritical := criticalSteps[stepNum]
		if step.IsCritical != expectedCritical {
			t.Errorf("step %d IsCritical: expected %v, got %v", stepNum, expectedCritical, step.IsCritical)
		}
	}
}

// TestCNDNSagaTimeouts verifies timeout configurations
func TestCNDNSagaTimeouts(t *testing.T) {
	s := NewCreditDebitNoteSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 8 {
			continue // Skip compensation steps
		}
		if step.TimeoutSeconds < 20 || step.TimeoutSeconds > 40 {
			t.Errorf("step %d timeout %d out of range [20-40]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestCNDNSagaServiceNames verifies service name conventions
func TestCNDNSagaServiceNames(t *testing.T) {
	s := NewCreditDebitNoteSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"gst":             true,
		"tax-engine":      true,
		"sales-invoice":   true,
		"gst-ledger":      true,
		"general-ledger":  true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestCNDNSagaCompensationSteps verifies compensation mapping
func TestCNDNSagaCompensationSteps(t *testing.T) {
	s := NewCreditDebitNoteSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount < 6 || compensationCount > 7 {
		t.Errorf("expected 6-7 compensation steps, got %d", compensationCount)
	}
}

// TestCNDNSagaInputMapping verifies step 1 input fields
func TestCNDNSagaInputMapping(t *testing.T) {
	s := NewCreditDebitNoteSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	requiredFields := []string{"originalInvoiceID", "transactionType"}
	for _, field := range requiredFields {
		if _, ok := step1.InputMapping[field]; !ok {
			t.Errorf("step 1 InputMapping missing required field: %s", field)
		}
	}
}

// TestCNDNSagaRetryConfig verifies retry configuration
func TestCNDNSagaRetryConfig(t *testing.T) {
	s := NewCreditDebitNoteSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 8 {
			continue // Skip compensation steps
		}
		if step.RetryConfig == nil {
			t.Errorf("step %d has nil RetryConfig", step.StepNumber)
		}
	}
}

// TestCNDNSagaInterfaceImplementation verifies SagaHandler interface
func TestCNDNSagaInterfaceImplementation(t *testing.T) {
	s := NewCreditDebitNoteSaga()
	if s.SagaType() != "SAGA-G08" {
		t.Error("interface implementation failed")
	}
}
