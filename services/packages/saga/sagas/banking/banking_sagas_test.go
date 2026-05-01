// Package banking provides comprehensive tests for Phase 5A Banking sagas
package banking

import (
	"testing"
)

// ============================================================================
// SAGA-B01: Wire Transfer & Payment Authorization
// ============================================================================

// TestWireTransferSagaType verifies saga type identification
func TestWireTransferSagaType(t *testing.T) {
	s := NewWireTransferSaga()
	expected := "SAGA-B01"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestWireTransferSagaStepCount verifies 18 steps total
func TestWireTransferSagaStepCount(t *testing.T) {
	s := NewWireTransferSaga()
	steps := s.GetStepDefinitions()
	expected := 18 // 10 forward + 8 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestWireTransferSagaValidation verifies input validation
func TestWireTransferSagaValidation(t *testing.T) {
	s := NewWireTransferSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"payment_id":          "PAY001",
				"payment_amount":      100000,
				"beneficiary_account": "ACC123",
				"payment_method":      "wire",
			},
			hasErr: false,
		},
		{
			name: "missing payment_id",
			input: map[string]interface{}{
				"payment_amount":      100000,
				"beneficiary_account": "ACC123",
				"payment_method":      "wire",
			},
			hasErr: true,
		},
		{
			name: "missing payment_amount",
			input: map[string]interface{}{
				"payment_id":          "PAY001",
				"beneficiary_account": "ACC123",
				"payment_method":      "wire",
			},
			hasErr: true,
		},
		{
			name: "missing beneficiary_account",
			input: map[string]interface{}{
				"payment_id":     "PAY001",
				"payment_amount": 100000,
				"payment_method": "wire",
			},
			hasErr: true,
		},
		{
			name: "missing payment_method",
			input: map[string]interface{}{
				"payment_id":          "PAY001",
				"payment_amount":      100000,
				"beneficiary_account": "ACC123",
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

// TestWireTransferSagaCriticalSteps verifies critical steps
func TestWireTransferSagaCriticalSteps(t *testing.T) {
	s := NewWireTransferSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 7: true, 10: true}

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

// TestWireTransferSagaTimeouts verifies timeout configurations
func TestWireTransferSagaTimeouts(t *testing.T) {
	s := NewWireTransferSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 10 {
			continue // Skip compensation steps
		}
		if step.TimeoutSeconds < 20 || step.TimeoutSeconds > 45 {
			t.Errorf("step %d timeout %d out of range [20-45]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestWireTransferSagaServiceNames verifies service name conventions
func TestWireTransferSagaServiceNames(t *testing.T) {
	s := NewWireTransferSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"banking":           true,
		"payment-gateway":   true,
		"general-ledger":    true,
		"approval":          true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestWireTransferSagaCompensationSteps verifies compensation mapping
func TestWireTransferSagaCompensationSteps(t *testing.T) {
	s := NewWireTransferSaga()
	steps := s.GetStepDefinitions()

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

// TestWireTransferSagaInputMapping verifies step 1 input fields
func TestWireTransferSagaInputMapping(t *testing.T) {
	s := NewWireTransferSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	requiredFields := []string{"tenantID", "companyID", "branchID", "paymentID", "paymentAmount"}
	for _, field := range requiredFields {
		if _, ok := step1.InputMapping[field]; !ok {
			t.Errorf("step 1 InputMapping missing required field: %s", field)
		}
	}
}

// TestWireTransferSagaRetryConfig verifies retry configuration
func TestWireTransferSagaRetryConfig(t *testing.T) {
	s := NewWireTransferSaga()
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
	}
}

// TestWireTransferSagaInterfaceImplementation verifies SagaHandler interface
func TestWireTransferSagaInterfaceImplementation(t *testing.T) {
	s := NewWireTransferSaga()
	if s.SagaType() != "SAGA-B01" {
		t.Error("interface implementation failed")
	}
}

// ============================================================================
// SAGA-B02: Bank Reconciliation
// ============================================================================

// TestBankReconciliationSagaType verifies saga type identification
func TestBankReconciliationSagaType(t *testing.T) {
	s := NewBankReconciliationSaga()
	expected := "SAGA-B02"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestBankReconciliationSagaStepCount verifies 20 steps total
func TestBankReconciliationSagaStepCount(t *testing.T) {
	s := NewBankReconciliationSaga()
	steps := s.GetStepDefinitions()
	expected := 20 // 11 forward + 9 compensation (longest banking saga)
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestBankReconciliationSagaValidation verifies input validation
func TestBankReconciliationSagaValidation(t *testing.T) {
	s := NewBankReconciliationSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"reconciliation_date": "2024-02-15",
				"bank_account_id":     "ACC001",
				"statement_period":    "2024-02",
			},
			hasErr: false,
		},
		{
			name: "missing reconciliation_date",
			input: map[string]interface{}{
				"bank_account_id":  "ACC001",
				"statement_period": "2024-02",
			},
			hasErr: true,
		},
		{
			name: "missing bank_account_id",
			input: map[string]interface{}{
				"reconciliation_date": "2024-02-15",
				"statement_period":    "2024-02",
			},
			hasErr: true,
		},
		{
			name: "missing statement_period",
			input: map[string]interface{}{
				"reconciliation_date": "2024-02-15",
				"bank_account_id":     "ACC001",
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

// TestBankReconciliationSagaCriticalSteps verifies critical steps
func TestBankReconciliationSagaCriticalSteps(t *testing.T) {
	s := NewBankReconciliationSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 8: true, 11: true}

	for _, step := range steps {
		stepNum := int(step.StepNumber)
		if stepNum > 11 {
			continue // Skip compensation steps
		}
		expectedCritical := criticalSteps[stepNum]
		if step.IsCritical != expectedCritical {
			t.Errorf("step %d IsCritical: expected %v, got %v", stepNum, expectedCritical, step.IsCritical)
		}
	}
}

// TestBankReconciliationSagaTimeouts verifies timeout configurations
func TestBankReconciliationSagaTimeouts(t *testing.T) {
	s := NewBankReconciliationSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 11 {
			continue // Skip compensation steps
		}
		if step.TimeoutSeconds < 25 || step.TimeoutSeconds > 50 {
			t.Errorf("step %d timeout %d out of range [25-50]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestBankReconciliationSagaServiceNames verifies service name conventions
func TestBankReconciliationSagaServiceNames(t *testing.T) {
	s := NewBankReconciliationSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"banking":            true,
		"reconciliation":     true,
		"cash-management":    true,
		"general-ledger":     true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestBankReconciliationSagaCompensationSteps verifies compensation mapping
func TestBankReconciliationSagaCompensationSteps(t *testing.T) {
	s := NewBankReconciliationSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount < 8 || compensationCount > 10 {
		t.Errorf("expected 8-10 compensation steps, got %d", compensationCount)
	}
}

// TestBankReconciliationSagaInputMapping verifies step 1 input fields
func TestBankReconciliationSagaInputMapping(t *testing.T) {
	s := NewBankReconciliationSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	if _, ok := step1.InputMapping["bank_account_id"]; !ok {
		t.Error("step 1 InputMapping missing bank_account_id")
	}
}

// TestBankReconciliationSagaRetryConfig verifies retry configuration
func TestBankReconciliationSagaRetryConfig(t *testing.T) {
	s := NewBankReconciliationSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 11 {
			continue // Skip compensation steps
		}
		if step.RetryConfig == nil {
			t.Errorf("step %d has nil RetryConfig", step.StepNumber)
		}
	}
}

// TestBankReconciliationSagaInterfaceImplementation verifies SagaHandler interface
func TestBankReconciliationSagaInterfaceImplementation(t *testing.T) {
	s := NewBankReconciliationSaga()
	if s.SagaType() != "SAGA-B02" {
		t.Error("interface implementation failed")
	}
}

// ============================================================================
// SAGA-B03: Cash Positioning
// ============================================================================

// TestCashPositioningSagaType verifies saga type identification
func TestCashPositioningSagaType(t *testing.T) {
	s := NewCashPositioningSaga()
	expected := "SAGA-B03"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestCashPositioningSagaStepCount verifies 14 steps total
func TestCashPositioningSagaStepCount(t *testing.T) {
	s := NewCashPositioningSaga()
	steps := s.GetStepDefinitions()
	expected := 14 // 7 forward + 7 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestCashPositioningSagaValidation verifies input validation
func TestCashPositioningSagaValidation(t *testing.T) {
	s := NewCashPositioningSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"forecast_period":    "2024-02",
				"consolidation_level": "company",
				"currency":            "INR",
			},
			hasErr: false,
		},
		{
			name: "missing forecast_period",
			input: map[string]interface{}{
				"consolidation_level": "company",
				"currency":             "INR",
			},
			hasErr: true,
		},
		{
			name: "missing consolidation_level",
			input: map[string]interface{}{
				"forecast_period": "2024-02",
				"currency":         "INR",
			},
			hasErr: true,
		},
		{
			name: "missing currency",
			input: map[string]interface{}{
				"forecast_period":     "2024-02",
				"consolidation_level": "company",
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

// TestCashPositioningSagaCriticalSteps verifies critical steps
func TestCashPositioningSagaCriticalSteps(t *testing.T) {
	s := NewCashPositioningSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 5: true, 8: true}

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

// TestCashPositioningSagaTimeouts verifies timeout configurations
func TestCashPositioningSagaTimeouts(t *testing.T) {
	s := NewCashPositioningSaga()
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

// TestCashPositioningSagaServiceNames verifies service name conventions
func TestCashPositioningSagaServiceNames(t *testing.T) {
	s := NewCashPositioningSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"cash-management": true,
		"banking":         true,
		"general-ledger":  true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestCashPositioningSagaCompensationSteps verifies compensation mapping
func TestCashPositioningSagaCompensationSteps(t *testing.T) {
	s := NewCashPositioningSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount < 6 || compensationCount > 8 {
		t.Errorf("expected 6-8 compensation steps, got %d", compensationCount)
	}
}

// TestCashPositioningSagaInputMapping verifies step 1 input fields
func TestCashPositioningSagaInputMapping(t *testing.T) {
	s := NewCashPositioningSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	if _, ok := step1.InputMapping["forecast_period"]; !ok {
		t.Error("step 1 InputMapping missing forecast_period")
	}
}

// TestCashPositioningSagaRetryConfig verifies retry configuration
func TestCashPositioningSagaRetryConfig(t *testing.T) {
	s := NewCashPositioningSaga()
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

// TestCashPositioningSagaInterfaceImplementation verifies SagaHandler interface
func TestCashPositioningSagaInterfaceImplementation(t *testing.T) {
	s := NewCashPositioningSaga()
	if s.SagaType() != "SAGA-B03" {
		t.Error("interface implementation failed")
	}
}

// ============================================================================
// SAGA-B04: Cheque Management
// ============================================================================

// TestChequeManagementSagaType verifies saga type identification
func TestChequeManagementSagaType(t *testing.T) {
	s := NewChequeManagementSaga()
	expected := "SAGA-B04"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestChequeManagementSagaStepCount verifies 16 steps total
func TestChequeManagementSagaStepCount(t *testing.T) {
	s := NewChequeManagementSaga()
	steps := s.GetStepDefinitions()
	expected := 16 // 8 forward + 8 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestChequeManagementSagaValidation verifies input validation
func TestChequeManagementSagaValidation(t *testing.T) {
	s := NewChequeManagementSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"cheque_id":     "CHQ001",
				"cheque_amount": 50000,
				"payee":         "Vendor XYZ",
				"cheque_date":   "2024-02-28",
			},
			hasErr: false,
		},
		{
			name: "missing cheque_id",
			input: map[string]interface{}{
				"cheque_amount": 50000,
				"payee":         "Vendor XYZ",
				"cheque_date":   "2024-02-28",
			},
			hasErr: true,
		},
		{
			name: "missing cheque_amount",
			input: map[string]interface{}{
				"cheque_id":   "CHQ001",
				"payee":       "Vendor XYZ",
				"cheque_date": "2024-02-28",
			},
			hasErr: true,
		},
		{
			name: "missing payee",
			input: map[string]interface{}{
				"cheque_id":     "CHQ001",
				"cheque_amount": 50000,
				"cheque_date":   "2024-02-28",
			},
			hasErr: true,
		},
		{
			name: "missing cheque_date",
			input: map[string]interface{}{
				"cheque_id":     "CHQ001",
				"cheque_amount": 50000,
				"payee":         "Vendor XYZ",
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

// TestChequeManagementSagaCriticalSteps verifies critical steps
func TestChequeManagementSagaCriticalSteps(t *testing.T) {
	s := NewChequeManagementSaga()
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

// TestChequeManagementSagaTimeouts verifies timeout configurations
func TestChequeManagementSagaTimeouts(t *testing.T) {
	s := NewChequeManagementSaga()
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

// TestChequeManagementSagaServiceNames verifies service name conventions
func TestChequeManagementSagaServiceNames(t *testing.T) {
	s := NewChequeManagementSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"banking":           true,
		"cheque-management": true,
		"accounting":        true,
		"general-ledger":    true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestChequeManagementSagaCompensationSteps verifies compensation mapping
func TestChequeManagementSagaCompensationSteps(t *testing.T) {
	s := NewChequeManagementSaga()
	steps := s.GetStepDefinitions()

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

// TestChequeManagementSagaInputMapping verifies step 1 input fields
func TestChequeManagementSagaInputMapping(t *testing.T) {
	s := NewChequeManagementSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	if _, ok := step1.InputMapping["cheque_id"]; !ok {
		t.Error("step 1 InputMapping missing cheque_id")
	}
}

// TestChequeManagementSagaRetryConfig verifies retry configuration
func TestChequeManagementSagaRetryConfig(t *testing.T) {
	s := NewChequeManagementSaga()
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

// TestChequeManagementSagaInterfaceImplementation verifies SagaHandler interface
func TestChequeManagementSagaInterfaceImplementation(t *testing.T) {
	s := NewChequeManagementSaga()
	if s.SagaType() != "SAGA-B04" {
		t.Error("interface implementation failed")
	}
}

// ============================================================================
// SAGA-B05: Payment Gateway Processing
// ============================================================================

// TestPaymentGatewaySagaType verifies saga type identification
func TestPaymentGatewaySagaType(t *testing.T) {
	s := NewPaymentGatewaySaga()
	expected := "SAGA-B05"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestPaymentGatewaySagaStepCount verifies 18 steps total
func TestPaymentGatewaySagaStepCount(t *testing.T) {
	s := NewPaymentGatewaySaga()
	steps := s.GetStepDefinitions()
	expected := 18 // 9 forward + 9 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestPaymentGatewaySagaValidation verifies input validation
func TestPaymentGatewaySagaValidation(t *testing.T) {
	s := NewPaymentGatewaySaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"transaction_id":  "TXN001",
				"transaction_amount": 100000,
				"gateway_type":    "credit-card",
				"customer_id":     "CUST001",
			},
			hasErr: false,
		},
		{
			name: "missing transaction_id",
			input: map[string]interface{}{
				"transaction_amount": 100000,
				"gateway_type":       "credit-card",
				"customer_id":        "CUST001",
			},
			hasErr: true,
		},
		{
			name: "missing transaction_amount",
			input: map[string]interface{}{
				"transaction_id": "TXN001",
				"gateway_type":   "credit-card",
				"customer_id":    "CUST001",
			},
			hasErr: true,
		},
		{
			name: "missing gateway_type",
			input: map[string]interface{}{
				"transaction_id":     "TXN001",
				"transaction_amount": 100000,
				"customer_id":        "CUST001",
			},
			hasErr: true,
		},
		{
			name: "missing customer_id",
			input: map[string]interface{}{
				"transaction_id":     "TXN001",
				"transaction_amount": 100000,
				"gateway_type":       "credit-card",
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

// TestPaymentGatewaySagaCriticalSteps verifies critical steps
func TestPaymentGatewaySagaCriticalSteps(t *testing.T) {
	s := NewPaymentGatewaySaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 7: true, 10: true}

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

// TestPaymentGatewaySagaTimeouts verifies timeout configurations
func TestPaymentGatewaySagaTimeouts(t *testing.T) {
	s := NewPaymentGatewaySaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 9 {
			continue // Skip compensation steps
		}
		if step.TimeoutSeconds < 20 || step.TimeoutSeconds > 45 {
			t.Errorf("step %d timeout %d out of range [20-45]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestPaymentGatewaySagaServiceNames verifies service name conventions
func TestPaymentGatewaySagaServiceNames(t *testing.T) {
	s := NewPaymentGatewaySaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"payment-gateway":    true,
		"fraud-detection":    true,
		"settlement":         true,
		"general-ledger":     true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestPaymentGatewaySagaCompensationSteps verifies compensation mapping
func TestPaymentGatewaySagaCompensationSteps(t *testing.T) {
	s := NewPaymentGatewaySaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount < 8 || compensationCount > 10 {
		t.Errorf("expected 8-10 compensation steps, got %d", compensationCount)
	}
}

// TestPaymentGatewaySagaInputMapping verifies step 1 input fields
func TestPaymentGatewaySagaInputMapping(t *testing.T) {
	s := NewPaymentGatewaySaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	if _, ok := step1.InputMapping["transaction_id"]; !ok {
		t.Error("step 1 InputMapping missing transaction_id")
	}
}

// TestPaymentGatewaySagaRetryConfig verifies retry configuration
func TestPaymentGatewaySagaRetryConfig(t *testing.T) {
	s := NewPaymentGatewaySaga()
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

// TestPaymentGatewaySagaInterfaceImplementation verifies SagaHandler interface
func TestPaymentGatewaySagaInterfaceImplementation(t *testing.T) {
	s := NewPaymentGatewaySaga()
	if s.SagaType() != "SAGA-B05" {
		t.Error("interface implementation failed")
	}
}

// ============================================================================
// SAGA-B06: Compliance Monitoring
// ============================================================================

// TestComplianceMonitoringSagaType verifies saga type identification
func TestComplianceMonitoringSagaType(t *testing.T) {
	s := NewComplianceMonitoringSaga()
	expected := "SAGA-B06"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestComplianceMonitoringSagaStepCount verifies 16 steps total
func TestComplianceMonitoringSagaStepCount(t *testing.T) {
	s := NewComplianceMonitoringSaga()
	steps := s.GetStepDefinitions()
	expected := 16 // 8 forward + 8 compensation
	if len(steps) != expected {
		t.Errorf("expected %d steps, got %d", expected, len(steps))
	}
}

// TestComplianceMonitoringSagaValidation verifies input validation
func TestComplianceMonitoringSagaValidation(t *testing.T) {
	s := NewComplianceMonitoringSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"monitoring_period": "2024-02",
				"aml_threshold":      1000000,
				"ctr_threshold":      10000000,
			},
			hasErr: false,
		},
		{
			name: "missing monitoring_period",
			input: map[string]interface{}{
				"aml_threshold": 1000000,
				"ctr_threshold": 10000000,
			},
			hasErr: true,
		},
		{
			name: "missing aml_threshold",
			input: map[string]interface{}{
				"monitoring_period": "2024-02",
				"ctr_threshold":      10000000,
			},
			hasErr: true,
		},
		{
			name: "missing ctr_threshold",
			input: map[string]interface{}{
				"monitoring_period": "2024-02",
				"aml_threshold":      1000000,
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

// TestComplianceMonitoringSagaCriticalSteps verifies critical steps
func TestComplianceMonitoringSagaCriticalSteps(t *testing.T) {
	s := NewComplianceMonitoringSaga()
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

// TestComplianceMonitoringSagaTimeouts verifies timeout configurations
func TestComplianceMonitoringSagaTimeouts(t *testing.T) {
	s := NewComplianceMonitoringSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber > 8 {
			continue // Skip compensation steps
		}
		if step.TimeoutSeconds < 25 || step.TimeoutSeconds > 50 {
			t.Errorf("step %d timeout %d out of range [25-50]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestComplianceMonitoringSagaServiceNames verifies service name conventions
func TestComplianceMonitoringSagaServiceNames(t *testing.T) {
	s := NewComplianceMonitoringSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"banking":                true,
		"compliance":             true,
		"fraud-detection":        true,
		"regulatory-reporting":   true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestComplianceMonitoringSagaCompensationSteps verifies compensation mapping
func TestComplianceMonitoringSagaCompensationSteps(t *testing.T) {
	s := NewComplianceMonitoringSaga()
	steps := s.GetStepDefinitions()

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

// TestComplianceMonitoringSagaInputMapping verifies step 1 input fields
func TestComplianceMonitoringSagaInputMapping(t *testing.T) {
	s := NewComplianceMonitoringSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	if _, ok := step1.InputMapping["monitoring_period"]; !ok {
		t.Error("step 1 InputMapping missing monitoring_period")
	}
}

// TestComplianceMonitoringSagaRetryConfig verifies retry configuration
func TestComplianceMonitoringSagaRetryConfig(t *testing.T) {
	s := NewComplianceMonitoringSaga()
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

// TestComplianceMonitoringSagaInterfaceImplementation verifies SagaHandler interface
func TestComplianceMonitoringSagaInterfaceImplementation(t *testing.T) {
	s := NewComplianceMonitoringSaga()
	if s.SagaType() != "SAGA-B06" {
		t.Error("interface implementation failed")
	}
}
