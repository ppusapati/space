// Package finance provides tests for critical finance saga handlers
package finance

import (
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// ============================================================================
// BankReconciliationSaga (SAGA-F02) Tests
// ============================================================================

// TestBankReconciliationSagaType verifies saga type identification
func TestBankReconciliationSagaType(t *testing.T) {
	s := NewBankReconciliationSaga()
	expected := "SAGA-F02"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestBankReconciliationSagaStepCount verifies forward and compensation steps
func TestBankReconciliationSagaStepCount(t *testing.T) {
	s := NewBankReconciliationSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 21 // 11 forward + 10 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestBankReconciliationSagaValidation verifies input validation
func TestBankReconciliationSagaValidation(t *testing.T) {
	s := NewBankReconciliationSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"reconciliation_id": "REC001",
				"bank_account_id":   "ACC001",
				"start_date":        "2024-01-01",
				"end_date":          "2024-01-31",
			},
			hasErr: false,
		},
		{
			name: "missing reconciliation_id",
			input: map[string]interface{}{
				"bank_account_id": "ACC001",
				"start_date":      "2024-01-01",
				"end_date":        "2024-01-31",
			},
			hasErr: true,
			errMsg: "reconciliation_id is required",
		},
		{
			name: "missing bank_account_id",
			input: map[string]interface{}{
				"reconciliation_id": "REC001",
				"start_date":        "2024-01-01",
				"end_date":          "2024-01-31",
			},
			hasErr: true,
			errMsg: "bank_account_id is required",
		},
		{
			name: "missing start_date",
			input: map[string]interface{}{
				"reconciliation_id": "REC001",
				"bank_account_id":   "ACC001",
				"end_date":          "2024-01-31",
			},
			hasErr: true,
			errMsg: "start_date is required",
		},
		{
			name: "missing end_date",
			input: map[string]interface{}{
				"reconciliation_id": "REC001",
				"bank_account_id":   "ACC001",
				"start_date":        "2024-01-01",
			},
			hasErr: true,
			errMsg: "end_date is required",
		},
		{
			name: "invalid reconciliation_id type",
			input: map[string]interface{}{
				"reconciliation_id": 123,
				"bank_account_id":   "ACC001",
				"start_date":        "2024-01-01",
				"end_date":          "2024-01-31",
			},
			hasErr: true,
		},
		{
			name: "empty reconciliation_id",
			input: map[string]interface{}{
				"reconciliation_id": "",
				"bank_account_id":   "ACC001",
				"start_date":        "2024-01-01",
				"end_date":          "2024-01-31",
			},
			hasErr: true,
		},
		{
			name:   "invalid input type",
			input:  "invalid",
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

// TestBankReconciliationSagaCriticalSteps verifies critical step markers
func TestBankReconciliationSagaCriticalSteps(t *testing.T) {
	s := NewBankReconciliationSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 8: true, 11: true}

	for _, step := range steps {
		stepNum := int(step.StepNumber)
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

	forwardSteps := steps[:11] // First 11 are forward steps
	for _, step := range forwardSteps {
		if step.TimeoutSeconds < 30 || step.TimeoutSeconds > 45 {
			t.Errorf("step %d timeout %d out of range [30-45]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestBankReconciliationSagaServiceNames verifies service name conventions
func TestBankReconciliationSagaServiceNames(t *testing.T) {
	s := NewBankReconciliationSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"banking":              true,
		"reconciliation":       true,
		"accounts-payable":     true,
		"accounts-receivable":  true,
		"general-ledger":       true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestBankReconciliationSagaCompensationSteps verifies compensation mappings
func TestBankReconciliationSagaCompensationSteps(t *testing.T) {
	s := NewBankReconciliationSaga()
	steps := s.GetStepDefinitions()

	// Steps 2-11 should have compensation steps in range 102-111
	for stepNum := 2; stepNum <= 11; stepNum++ {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}

		// Skip step 1 and 11 which have empty compensation
		if stepNum == 1 || stepNum == 11 {
			continue
		}

		hasCompensation := len(step.CompensationSteps) > 0
		if !hasCompensation {
			t.Errorf("step %d should have compensation steps", stepNum)
		}

		for _, compStep := range step.CompensationSteps {
			if compStep < 102 || compStep > 111 {
				t.Errorf("step %d compensation %d out of range [102-111]", stepNum, compStep)
			}
		}
	}
}

// TestBankReconciliationSagaInputMapping verifies step 1 input mapping
func TestBankReconciliationSagaInputMapping(t *testing.T) {
	s := NewBankReconciliationSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	requiredFields := []string{
		"tenantID", "companyID", "branchID", "reconciliationID",
		"bankAccountID", "startDate", "endDate",
	}

	for _, field := range requiredFields {
		if _, ok := step1.InputMapping[field]; !ok {
			t.Errorf("step 1 InputMapping missing required field: %s", field)
		}
	}
}

// TestBankReconciliationSagaRetryConfig verifies standard retry configuration
func TestBankReconciliationSagaRetryConfig(t *testing.T) {
	s := NewBankReconciliationSaga()
	steps := s.GetStepDefinitions()

	forwardSteps := steps[:11] // Forward steps only
	for _, step := range forwardSteps {
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

// TestBankReconciliationSagaFirstAndLastSteps verifies terminal steps
func TestBankReconciliationSagaFirstAndLastSteps(t *testing.T) {
	s := NewBankReconciliationSaga()

	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}
	if len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have empty CompensationSteps, got %d", len(step1.CompensationSteps))
	}

	step11 := s.GetStepDefinition(11)
	if step11 == nil {
		t.Fatal("step 11 not found")
	}
	if len(step11.CompensationSteps) != 0 {
		t.Errorf("step 11 should have empty CompensationSteps, got %d", len(step11.CompensationSteps))
	}
}

// TestBankReconciliationSagaImplementsInterface verifies interface implementation
func TestBankReconciliationSagaImplementsInterface(t *testing.T) {
	var _ saga.SagaHandler = (*BankReconciliationSaga)(nil)
	s := NewBankReconciliationSaga()

	if s.SagaType() == "" {
		t.Error("SagaType() should not return empty string")
	}
	if len(s.GetStepDefinitions()) == 0 {
		t.Error("GetStepDefinitions() should not return empty slice")
	}
}

// TestBankReconciliationSagaGetStepByNumber verifies step retrieval
func TestBankReconciliationSagaGetStepByNumber(t *testing.T) {
	s := NewBankReconciliationSaga()

	tests := []struct {
		stepNum   int
		shouldExist bool
	}{
		{1, true},
		{5, true},
		{11, true},
		{102, true},
		{110, true},
		{999, false},
		{0, false},
		{-1, false},
	}

	for _, tc := range tests {
		step := s.GetStepDefinition(tc.stepNum)
		if tc.shouldExist && step == nil {
			t.Errorf("step %d should exist but was nil", tc.stepNum)
		}
		if !tc.shouldExist && step != nil {
			t.Errorf("step %d should not exist but was found", tc.stepNum)
		}
	}
}

// ============================================================================
// MultiCurrencyRevaluationSaga (SAGA-F03) Tests
// ============================================================================

// TestMultiCurrencyRevaluationSagaType verifies saga type identification
func TestMultiCurrencyRevaluationSagaType(t *testing.T) {
	s := NewMultiCurrencyRevaluationSaga()
	expected := "SAGA-F03"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestMultiCurrencyRevaluationSagaStepCount verifies forward and compensation steps
func TestMultiCurrencyRevaluationSagaStepCount(t *testing.T) {
	s := NewMultiCurrencyRevaluationSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 19 // 10 forward + 9 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestMultiCurrencyRevaluationSagaValidation verifies input validation
func TestMultiCurrencyRevaluationSagaValidation(t *testing.T) {
	s := NewMultiCurrencyRevaluationSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"revaluation_id":   "REV001",
				"revaluation_date": "2024-01-31",
				"currency_list":    []interface{}{"USD", "EUR", "GBP"},
			},
			hasErr: false,
		},
		{
			name: "missing revaluation_id",
			input: map[string]interface{}{
				"revaluation_date": "2024-01-31",
				"currency_list":    []interface{}{"USD", "EUR"},
			},
			hasErr: true,
			errMsg: "revaluation_id is required",
		},
		{
			name: "missing revaluation_date",
			input: map[string]interface{}{
				"revaluation_id": "REV001",
				"currency_list":  []interface{}{"USD", "EUR"},
			},
			hasErr: true,
			errMsg: "revaluation_date is required",
		},
		{
			name: "missing currency_list",
			input: map[string]interface{}{
				"revaluation_id":   "REV001",
				"revaluation_date": "2024-01-31",
			},
			hasErr: true,
			errMsg: "currency_list is required",
		},
		{
			name: "empty revaluation_id",
			input: map[string]interface{}{
				"revaluation_id":   "",
				"revaluation_date": "2024-01-31",
				"currency_list":    []interface{}{"USD"},
			},
			hasErr: true,
		},
		{
			name: "empty currency_list",
			input: map[string]interface{}{
				"revaluation_id":   "REV001",
				"revaluation_date": "2024-01-31",
				"currency_list":    []interface{}{},
			},
			hasErr: true,
			errMsg: "currency_list must be a non-empty array",
		},
		{
			name: "invalid currency_list type",
			input: map[string]interface{}{
				"revaluation_id":   "REV001",
				"revaluation_date": "2024-01-31",
				"currency_list":    "USD,EUR",
			},
			hasErr: true,
		},
		{
			name:   "invalid input type",
			input:  "invalid",
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

// TestMultiCurrencyRevaluationSagaCriticalSteps verifies critical step markers
func TestMultiCurrencyRevaluationSagaCriticalSteps(t *testing.T) {
	s := NewMultiCurrencyRevaluationSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 10: true}

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

// TestMultiCurrencyRevaluationSagaTimeouts verifies timeout configurations
func TestMultiCurrencyRevaluationSagaTimeouts(t *testing.T) {
	s := NewMultiCurrencyRevaluationSaga()
	steps := s.GetStepDefinitions()

	forwardSteps := steps[:10] // First 10 are forward steps
	for _, step := range forwardSteps {
		if step.TimeoutSeconds < 30 || step.TimeoutSeconds > 60 {
			t.Errorf("step %d timeout %d out of range [30-60]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestMultiCurrencyRevaluationSagaServiceNames verifies service name conventions
func TestMultiCurrencyRevaluationSagaServiceNames(t *testing.T) {
	s := NewMultiCurrencyRevaluationSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"currency":             true,
		"accounts-receivable":  true,
		"accounts-payable":     true,
		"banking":              true,
		"cost-center":          true,
		"general-ledger":       true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestMultiCurrencyRevaluationSagaCompensationSteps verifies compensation mappings
func TestMultiCurrencyRevaluationSagaCompensationSteps(t *testing.T) {
	s := NewMultiCurrencyRevaluationSaga()
	steps := s.GetStepDefinitions()

	// Steps 2-9 should have compensation steps
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

// TestMultiCurrencyRevaluationSagaInputMapping verifies step 1 input mapping
func TestMultiCurrencyRevaluationSagaInputMapping(t *testing.T) {
	s := NewMultiCurrencyRevaluationSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	requiredFields := []string{
		"tenantID", "companyID", "branchID", "revaluationID",
		"revaluationDate", "currencyList",
	}

	for _, field := range requiredFields {
		if _, ok := step1.InputMapping[field]; !ok {
			t.Errorf("step 1 InputMapping missing required field: %s", field)
		}
	}
}

// TestMultiCurrencyRevaluationSagaRetryConfig verifies retry configuration including special config
func TestMultiCurrencyRevaluationSagaRetryConfig(t *testing.T) {
	s := NewMultiCurrencyRevaluationSaga()

	// Test standard retry config for steps 1-9
	for i := 1; i <= 9; i++ {
		step := s.GetStepDefinition(i)
		if step.RetryConfig == nil {
			t.Errorf("step %d missing RetryConfig", step.StepNumber)
			continue
		}

		if step.RetryConfig.MaxRetries != 3 {
			t.Errorf("step %d MaxRetries: expected 3, got %d", step.StepNumber, step.RetryConfig.MaxRetries)
		}
	}

	// Test special retry config for step 10
	step10 := s.GetStepDefinition(10)
	if step10.RetryConfig == nil {
		t.Fatal("step 10 missing RetryConfig")
	}
	if step10.RetryConfig.MaxRetries != 5 {
		t.Errorf("step 10 MaxRetries: expected 5, got %d", step10.RetryConfig.MaxRetries)
	}
}

// TestMultiCurrencyRevaluationSagaFirstAndLastSteps verifies terminal steps
func TestMultiCurrencyRevaluationSagaFirstAndLastSteps(t *testing.T) {
	s := NewMultiCurrencyRevaluationSaga()

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

// TestMultiCurrencyRevaluationSagaImplementsInterface verifies interface implementation
func TestMultiCurrencyRevaluationSagaImplementsInterface(t *testing.T) {
	var _ saga.SagaHandler = (*MultiCurrencyRevaluationSaga)(nil)
	s := NewMultiCurrencyRevaluationSaga()

	if s.SagaType() == "" {
		t.Error("SagaType() should not return empty string")
	}
	if len(s.GetStepDefinitions()) == 0 {
		t.Error("GetStepDefinitions() should not return empty slice")
	}
}

// TestMultiCurrencyRevaluationSagaGetStepByNumber verifies step retrieval
func TestMultiCurrencyRevaluationSagaGetStepByNumber(t *testing.T) {
	s := NewMultiCurrencyRevaluationSaga()

	tests := []struct {
		stepNum     int
		shouldExist bool
	}{
		{1, true},
		{5, true},
		{10, true},
		{102, true},
		{109, true},
		{999, false},
		{0, false},
		{-1, false},
	}

	for _, tc := range tests {
		step := s.GetStepDefinition(tc.stepNum)
		if tc.shouldExist && step == nil {
			t.Errorf("step %d should exist but was nil", tc.stepNum)
		}
		if !tc.shouldExist && step != nil {
			t.Errorf("step %d should not exist but was found", tc.stepNum)
		}
	}
}

// ============================================================================
// IntercompanyTransactionSaga (SAGA-F04) Tests
// ============================================================================

// TestIntercompanyTransactionSagaType verifies saga type identification
func TestIntercompanyTransactionSagaType(t *testing.T) {
	s := NewIntercompanyTransactionSaga()
	expected := "SAGA-F04"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestIntercompanyTransactionSagaStepCount verifies forward and compensation steps
func TestIntercompanyTransactionSagaStepCount(t *testing.T) {
	s := NewIntercompanyTransactionSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 15 // 8 forward + 7 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestIntercompanyTransactionSagaValidation verifies input validation
func TestIntercompanyTransactionSagaValidation(t *testing.T) {
	s := NewIntercompanyTransactionSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"invoice_id":        "INV001",
				"from_company_id":   "COMP001",
				"to_company_id":     "COMP002",
				"amount":            10000.00,
			},
			hasErr: false,
		},
		{
			name: "missing invoice_id",
			input: map[string]interface{}{
				"from_company_id": "COMP001",
				"to_company_id":   "COMP002",
				"amount":          10000.00,
			},
			hasErr: true,
			errMsg: "invoice_id is required",
		},
		{
			name: "missing from_company_id",
			input: map[string]interface{}{
				"invoice_id":      "INV001",
				"to_company_id":   "COMP002",
				"amount":          10000.00,
			},
			hasErr: true,
			errMsg: "from_company_id is required",
		},
		{
			name: "missing to_company_id",
			input: map[string]interface{}{
				"invoice_id":      "INV001",
				"from_company_id": "COMP001",
				"amount":          10000.00,
			},
			hasErr: true,
			errMsg: "to_company_id is required",
		},
		{
			name: "missing amount",
			input: map[string]interface{}{
				"invoice_id":      "INV001",
				"from_company_id": "COMP001",
				"to_company_id":   "COMP002",
			},
			hasErr: true,
			errMsg: "amount is required",
		},
		{
			name: "negative amount",
			input: map[string]interface{}{
				"invoice_id":      "INV001",
				"from_company_id": "COMP001",
				"to_company_id":   "COMP002",
				"amount":          -1000.00,
			},
			hasErr: true,
		},
		{
			name: "same company ids",
			input: map[string]interface{}{
				"invoice_id":      "INV001",
				"from_company_id": "COMP001",
				"to_company_id":   "COMP001",
				"amount":          10000.00,
			},
			hasErr: true,
			errMsg: "from_company_id and to_company_id must be different",
		},
		{
			name: "invalid amount type",
			input: map[string]interface{}{
				"invoice_id":      "INV001",
				"from_company_id": "COMP001",
				"to_company_id":   "COMP002",
				"amount":          "10000",
			},
			hasErr: true,
		},
		{
			name: "empty invoice_id",
			input: map[string]interface{}{
				"invoice_id":      "",
				"from_company_id": "COMP001",
				"to_company_id":   "COMP002",
				"amount":          10000.00,
			},
			hasErr: true,
		},
		{
			name:   "invalid input type",
			input:  "invalid",
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

// TestIntercompanyTransactionSagaCriticalSteps verifies critical step markers
func TestIntercompanyTransactionSagaCriticalSteps(t *testing.T) {
	s := NewIntercompanyTransactionSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 6: true, 8: true}

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

// TestIntercompanyTransactionSagaTimeouts verifies timeout configurations
func TestIntercompanyTransactionSagaTimeouts(t *testing.T) {
	s := NewIntercompanyTransactionSaga()
	steps := s.GetStepDefinitions()

	forwardSteps := steps[:8] // First 8 are forward steps
	for _, step := range forwardSteps {
		if step.TimeoutSeconds < 30 || step.TimeoutSeconds > 45 {
			t.Errorf("step %d timeout %d out of range [30-45]", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestIntercompanyTransactionSagaServiceNames verifies service name conventions
func TestIntercompanyTransactionSagaServiceNames(t *testing.T) {
	s := NewIntercompanyTransactionSaga()
	steps := s.GetStepDefinitions()

	expectedServices := map[string]bool{
		"accounts-payable":    true,
		"accounts-receivable": true,
		"general-ledger":      true,
		"cost-center":         true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestIntercompanyTransactionSagaCompensationSteps verifies compensation mappings
func TestIntercompanyTransactionSagaCompensationSteps(t *testing.T) {
	s := NewIntercompanyTransactionSaga()
	steps := s.GetStepDefinitions()

	// Steps 2-7 should have compensation steps
	for stepNum := 2; stepNum <= 7; stepNum++ {
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
			if compStep < 102 || compStep > 107 {
				t.Errorf("step %d compensation %d out of range [102-107]", stepNum, compStep)
			}
		}
	}
}

// TestIntercompanyTransactionSagaInputMapping verifies step 1 input mapping
func TestIntercompanyTransactionSagaInputMapping(t *testing.T) {
	s := NewIntercompanyTransactionSaga()
	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}

	requiredFields := []string{
		"tenantID", "companyID", "branchID", "invoiceID",
		"fromCompanyID", "toCompanyID", "amount",
	}

	for _, field := range requiredFields {
		if _, ok := step1.InputMapping[field]; !ok {
			t.Errorf("step 1 InputMapping missing required field: %s", field)
		}
	}
}

// TestIntercompanyTransactionSagaRetryConfig verifies standard retry configuration
func TestIntercompanyTransactionSagaRetryConfig(t *testing.T) {
	s := NewIntercompanyTransactionSaga()

	forwardSteps := s.GetStepDefinitions()[:8] // Forward steps only
	for _, step := range forwardSteps {
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

// TestIntercompanyTransactionSagaFirstAndLastSteps verifies terminal steps
func TestIntercompanyTransactionSagaFirstAndLastSteps(t *testing.T) {
	s := NewIntercompanyTransactionSaga()

	step1 := s.GetStepDefinition(1)
	if step1 == nil {
		t.Fatal("step 1 not found")
	}
	if len(step1.CompensationSteps) != 0 {
		t.Errorf("step 1 should have empty CompensationSteps, got %d", len(step1.CompensationSteps))
	}

	step8 := s.GetStepDefinition(8)
	if step8 == nil {
		t.Fatal("step 8 not found")
	}
	if len(step8.CompensationSteps) != 0 {
		t.Errorf("step 8 should have empty CompensationSteps, got %d", len(step8.CompensationSteps))
	}
}

// TestIntercompanyTransactionSagaImplementsInterface verifies interface implementation
func TestIntercompanyTransactionSagaImplementsInterface(t *testing.T) {
	var _ saga.SagaHandler = (*IntercompanyTransactionSaga)(nil)
	s := NewIntercompanyTransactionSaga()

	if s.SagaType() == "" {
		t.Error("SagaType() should not return empty string")
	}
	if len(s.GetStepDefinitions()) == 0 {
		t.Error("GetStepDefinitions() should not return empty slice")
	}
}

// TestIntercompanyTransactionSagaGetStepByNumber verifies step retrieval
func TestIntercompanyTransactionSagaGetStepByNumber(t *testing.T) {
	s := NewIntercompanyTransactionSaga()

	tests := []struct {
		stepNum     int
		shouldExist bool
	}{
		{1, true},
		{4, true},
		{8, true},
		{102, true},
		{107, true},
		{999, false},
		{0, false},
		{-1, false},
	}

	for _, tc := range tests {
		step := s.GetStepDefinition(tc.stepNum)
		if tc.shouldExist && step == nil {
			t.Errorf("step %d should exist but was nil", tc.stepNum)
		}
		if !tc.shouldExist && step != nil {
			t.Errorf("step %d should not exist but was found", tc.stepNum)
		}
	}
}
