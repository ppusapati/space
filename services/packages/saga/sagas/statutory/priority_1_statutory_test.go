// Package statutory provides saga handlers for statutory compliance workflows
package statutory

import (
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// ===== GSTR-1 FILING SAGA (ST01) TESTS =====

func TestGSTR1FilingSagaType(t *testing.T) {
	s := NewGSTR1FilingSaga().(*GSTR1FilingSaga)
	expected := "SAGA-ST01"
	if s.SagaType() != expected {
		t.Errorf("expected saga type %s, got %s", expected, s.SagaType())
	}
}

func TestGSTR1FilingSagaStepDefinitions(t *testing.T) {
	s := NewGSTR1FilingSaga().(*GSTR1FilingSaga)
	steps := s.GetStepDefinitions()

	// Should have 11 forward steps + 9 compensation steps = 20 total
	expectedCount := 20
	if len(steps) != expectedCount {
		t.Errorf("expected %d step definitions, got %d", expectedCount, len(steps))
	}
}

func TestGSTR1FilingSagaCriticalSteps(t *testing.T) {
	s := NewGSTR1FilingSaga().(*GSTR1FilingSaga)
	criticalSteps := []int{1, 4, 5, 7, 8, 9, 10}

	for _, stepNum := range criticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be marked as critical", stepNum)
		}
	}
}

func TestGSTR1FilingSagaValidateInputMissingGSTIN(t *testing.T) {
	s := NewGSTR1FilingSaga().(*GSTR1FilingSaga)
	input := map[string]interface{}{
		"filing_period": "2024-02",
	}
	err := s.ValidateInput(input)
	if err == nil {
		t.Error("expected error for missing gstin")
	}
}

func TestGSTR1FilingSagaValidateInputInvalidGSTIN(t *testing.T) {
	s := NewGSTR1FilingSaga().(*GSTR1FilingSaga)
	input := map[string]interface{}{
		"gstin":          "123",
		"filing_period":  "2024-02",
	}
	err := s.ValidateInput(input)
	if err == nil {
		t.Error("expected error for invalid gstin length")
	}
}

func TestGSTR1FilingSagaValidateInputMissingFilingPeriod(t *testing.T) {
	s := NewGSTR1FilingSaga().(*GSTR1FilingSaga)
	input := map[string]interface{}{
		"gstin": "29AABCT0959A2Z5",
	}
	err := s.ValidateInput(input)
	if err == nil {
		t.Error("expected error for missing filing_period")
	}
}

func TestGSTR1FilingSagaValidateInputInvalidFilingPeriod(t *testing.T) {
	s := NewGSTR1FilingSaga().(*GSTR1FilingSaga)
	input := map[string]interface{}{
		"gstin":          "29AABCT0959A2Z5",
		"filing_period":  "02-2024",
	}
	err := s.ValidateInput(input)
	if err == nil {
		t.Error("expected error for invalid filing_period format")
	}
}

func TestGSTR1FilingSagaValidateInputMissingClassificationRules(t *testing.T) {
	s := NewGSTR1FilingSaga().(*GSTR1FilingSaga)
	input := map[string]interface{}{
		"gstin":               "29AABCT0959A2Z5",
		"filing_period":       "2024-02",
		"period_start_date":   "2024-02-01",
		"period_end_date":     "2024-02-29",
		"hsn_master":          map[string]interface{}{},
		"tax_rate_master":     map[string]interface{}{},
	}
	err := s.ValidateInput(input)
	if err == nil {
		t.Error("expected error for missing classification_rules")
	}
}

func TestGSTR1FilingSagaValidateInputMissingDSCCertificate(t *testing.T) {
	s := NewGSTR1FilingSaga().(*GSTR1FilingSaga)
	input := map[string]interface{}{
		"gstin":                "29AABCT0959A2Z5",
		"filing_period":        "2024-02",
		"period_start_date":    "2024-02-01",
		"period_end_date":      "2024-02-29",
		"classification_rules": map[string]interface{}{},
		"hsn_master":           map[string]interface{}{},
		"tax_rate_master":      map[string]interface{}{},
		"declarant_details":    map[string]interface{}{},
	}
	err := s.ValidateInput(input)
	if err == nil {
		t.Error("expected error for missing dsc_certificate")
	}
}

func TestGSTR1FilingSagaValidateInputValid(t *testing.T) {
	s := NewGSTR1FilingSaga().(*GSTR1FilingSaga)
	input := map[string]interface{}{
		"gstin":                "29AABCT0959A2Z5",
		"filing_period":        "2024-02",
		"period_start_date":    "2024-02-01",
		"period_end_date":      "2024-02-29",
		"classification_rules": map[string]interface{}{},
		"hsn_master":           map[string]interface{}{},
		"tax_rate_master":      map[string]interface{}{},
		"declarant_details":    map[string]interface{}{},
		"dsc_certificate":      "cert_data",
	}
	err := s.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for valid input, got: %v", err)
	}
}

func TestGSTR1FilingSagaRetryConfiguration(t *testing.T) {
	s := NewGSTR1FilingSaga().(*GSTR1FilingSaga)

	// Check step 9 (SubmitToGSTN) which should have higher retry config
	step := s.GetStepDefinition(9)
	if step.RetryConfig.MaxRetries != 5 {
		t.Errorf("step 9 should have 5 max retries for GSTN API, got %d", step.RetryConfig.MaxRetries)
	}
	if step.TimeoutSeconds != 45 {
		t.Errorf("step 9 should have 45s timeout for GSTN API, got %d", step.TimeoutSeconds)
	}
}

// ===== GSTR-2 ITC SAGA (ST02) TESTS =====

func TestGSTR2ITCSagaType(t *testing.T) {
	s := NewGSTR2ITCSaga().(*GSTR2ITCSaga)
	expected := "SAGA-ST02"
	if s.SagaType() != expected {
		t.Errorf("expected saga type %s, got %s", expected, s.SagaType())
	}
}

func TestGSTR2ITCSagaStepDefinitions(t *testing.T) {
	s := NewGSTR2ITCSaga().(*GSTR2ITCSaga)
	steps := s.GetStepDefinitions()

	// Should have 10 forward steps + 9 compensation steps = 19 total
	expectedCount := 19
	if len(steps) != expectedCount {
		t.Errorf("expected %d step definitions, got %d", expectedCount, len(steps))
	}
}

func TestGSTR2ITCSagaCriticalSteps(t *testing.T) {
	s := NewGSTR2ITCSaga().(*GSTR2ITCSaga)
	criticalSteps := []int{1, 4, 6, 7, 9}

	for _, stepNum := range criticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be marked as critical", stepNum)
		}
	}
}

func TestGSTR2ITCSagaValidateInputMissingGSTIN(t *testing.T) {
	s := NewGSTR2ITCSaga().(*GSTR2ITCSaga)
	input := map[string]interface{}{
		"filing_period": "2024-02",
	}
	err := s.ValidateInput(input)
	if err == nil {
		t.Error("expected error for missing gstin")
	}
}

func TestGSTR2ITCSagaValidateInputInvalidGSTIN(t *testing.T) {
	s := NewGSTR2ITCSaga().(*GSTR2ITCSaga)
	input := map[string]interface{}{
		"gstin": "INVALID",
	}
	err := s.ValidateInput(input)
	if err == nil {
		t.Error("expected error for invalid gstin length")
	}
}

func TestGSTR2ITCSagaValidateInputMissingEligibilityRules(t *testing.T) {
	s := NewGSTR2ITCSaga().(*GSTR2ITCSaga)
	input := map[string]interface{}{
		"gstin":               "29AABCT0959A2Z5",
		"filing_period":       "2024-02",
		"period_start_date":   "2024-02-01",
		"period_end_date":     "2024-02-29",
	}
	err := s.ValidateInput(input)
	if err == nil {
		t.Error("expected error for missing eligibility_rules")
	}
}

func TestGSTR2ITCSagaValidateInputValid(t *testing.T) {
	s := NewGSTR2ITCSaga().(*GSTR2ITCSaga)
	input := map[string]interface{}{
		"gstin":                    "29AABCT0959A2Z5",
		"filing_period":            "2024-02",
		"period_start_date":        "2024-02-01",
		"period_end_date":          "2024-02-29",
		"eligibility_rules":        map[string]interface{}{},
		"exempt_supply_list":       map[string]interface{}{},
		"vendor_master":            map[string]interface{}{},
		"gstin_registry":           map[string]interface{}{},
		"itc_calculation_rules":    map[string]interface{}{},
		"reversal_rules":           map[string]interface{}{},
		"personal_use_list":        map[string]interface{}{},
		"output_tax_liability":     100000.0,
		"previous_period_itc":      map[string]interface{}{},
		"previous_credit_adjustments": map[string]interface{}{},
	}
	err := s.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for valid input, got: %v", err)
	}
}

// ===== GSTR-9 ANNUAL SAGA (ST03) TESTS =====

func TestGSTR9AnnualSagaType(t *testing.T) {
	s := NewGSTR9AnnualSaga().(*GSTR9AnnualSaga)
	expected := "SAGA-ST03"
	if s.SagaType() != expected {
		t.Errorf("expected saga type %s, got %s", expected, s.SagaType())
	}
}

func TestGSTR9AnnualSagaStepDefinitions(t *testing.T) {
	s := NewGSTR9AnnualSaga().(*GSTR9AnnualSaga)
	steps := s.GetStepDefinitions()

	// Should have 12 forward steps + 11 compensation steps = 23 total
	expectedCount := 23
	if len(steps) != expectedCount {
		t.Errorf("expected %d step definitions, got %d", expectedCount, len(steps))
	}
}

func TestGSTR9AnnualSagaCriticalSteps(t *testing.T) {
	s := NewGSTR9AnnualSaga().(*GSTR9AnnualSaga)
	criticalSteps := []int{1, 2, 3, 4, 5, 6, 7, 8, 10, 11}

	for _, stepNum := range criticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be marked as critical", stepNum)
		}
	}
}

func TestGSTR9AnnualSagaValidateInputMissingGSTIN(t *testing.T) {
	s := NewGSTR9AnnualSaga().(*GSTR9AnnualSaga)
	input := map[string]interface{}{
		"financial_year": "2023-24",
	}
	err := s.ValidateInput(input)
	if err == nil {
		t.Error("expected error for missing gstin")
	}
}

func TestGSTR9AnnualSagaValidateInputMissingFinancialYear(t *testing.T) {
	s := NewGSTR9AnnualSaga().(*GSTR9AnnualSaga)
	input := map[string]interface{}{
		"gstin": "29AABCT0959A2Z5",
	}
	err := s.ValidateInput(input)
	if err == nil {
		t.Error("expected error for missing financial_year")
	}
}

func TestGSTR9AnnualSagaValidateInputInvalidFinancialYearFormat(t *testing.T) {
	s := NewGSTR9AnnualSaga().(*GSTR9AnnualSaga)
	input := map[string]interface{}{
		"gstin":          "29AABCT0959A2Z5",
		"financial_year": "2023/24",
	}
	err := s.ValidateInput(input)
	if err == nil {
		t.Error("expected error for invalid financial_year format")
	}
}

func TestGSTR9AnnualSagaValidateInputValid(t *testing.T) {
	s := NewGSTR9AnnualSaga().(*GSTR9AnnualSaga)
	input := map[string]interface{}{
		"gstin":                     "29AABCT0959A2Z5",
		"financial_year":            "2023-24",
		"fy_start_date":             "2023-04-01",
		"fy_end_date":               "2024-03-31",
		"actual_sales_data":         map[string]interface{}{},
		"actual_purchase_data":      map[string]interface{}{},
		"reconciliation_rules":      map[string]interface{}{},
		"discrepancy_thresholds":    map[string]interface{}{},
		"tax_calculation_rules":     map[string]interface{}{},
		"itc_reversal_records":      map[string]interface{}{},
		"tax_payments_records":      map[string]interface{}{},
		"challan_payments":          map[string]interface{}{},
		"refund_claims":             map[string]interface{}{},
		"adjustment_records":        map[string]interface{}{},
		"dsc_certificate":           "cert_data",
		"previous_returns":          map[string]interface{}{},
	}
	err := s.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for valid input, got: %v", err)
	}
}

func TestGSTR9AnnualSagaHigherTimeouts(t *testing.T) {
	s := NewGSTR9AnnualSaga().(*GSTR9AnnualSaga)

	// Step 1 should have higher timeout for extracting annual data
	step := s.GetStepDefinition(1)
	if step.TimeoutSeconds != 45 {
		t.Errorf("step 1 should have 45s timeout for annual data extraction, got %d", step.TimeoutSeconds)
	}

	// Step 11 (SubmitGSTR9) should have 50s timeout
	step = s.GetStepDefinition(11)
	if step.TimeoutSeconds != 50 {
		t.Errorf("step 11 should have 50s timeout for GSTR-9 submission, got %d", step.TimeoutSeconds)
	}
}

// ===== TDS RETURN FILING SAGA (ST04) TESTS =====

func TestTDSReturnFilingSagaType(t *testing.T) {
	s := NewTDSReturnFilingSaga().(*TDSReturnFilingSaga)
	expected := "SAGA-ST04"
	if s.SagaType() != expected {
		t.Errorf("expected saga type %s, got %s", expected, s.SagaType())
	}
}

func TestTDSReturnFilingSagaStepDefinitions(t *testing.T) {
	s := NewTDSReturnFilingSaga().(*TDSReturnFilingSaga)
	steps := s.GetStepDefinitions()

	// Should have 10 forward steps + 9 compensation steps = 19 total
	expectedCount := 19
	if len(steps) != expectedCount {
		t.Errorf("expected %d step definitions, got %d", expectedCount, len(steps))
	}
}

func TestTDSReturnFilingSagaCriticalSteps(t *testing.T) {
	s := NewTDSReturnFilingSaga().(*TDSReturnFilingSaga)
	criticalSteps := []int{1, 4, 5, 6, 7, 8, 9}

	for _, stepNum := range criticalSteps {
		step := s.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be marked as critical", stepNum)
		}
	}
}

func TestTDSReturnFilingSagaValidateInputMissingTAN(t *testing.T) {
	s := NewTDSReturnFilingSaga().(*TDSReturnFilingSaga)
	input := map[string]interface{}{
		"filing_period": "2024-Q1",
	}
	err := s.ValidateInput(input)
	if err == nil {
		t.Error("expected error for missing tan")
	}
}

func TestTDSReturnFilingSagaValidateInputInvalidTAN(t *testing.T) {
	s := NewTDSReturnFilingSaga().(*TDSReturnFilingSaga)
	input := map[string]interface{}{
		"tan": "12345",
	}
	err := s.ValidateInput(input)
	if err == nil {
		t.Error("expected error for invalid tan length")
	}
}

func TestTDSReturnFilingSagaValidateInputInvalidPeriodType(t *testing.T) {
	s := NewTDSReturnFilingSaga().(*TDSReturnFilingSaga)
	input := map[string]interface{}{
		"tan":         "1234567890",
		"filing_period": "2024-Q1",
		"period_type": "MONTHLY",
	}
	err := s.ValidateInput(input)
	if err == nil {
		t.Error("expected error for invalid period_type")
	}
}

func TestTDSReturnFilingSagaValidateInputValidQuarterly(t *testing.T) {
	s := NewTDSReturnFilingSaga().(*TDSReturnFilingSaga)
	input := map[string]interface{}{
		"tan":                      "1234567890",
		"filing_period":            "2024-Q1",
		"period_type":              "QUARTERLY",
		"period_start_date":        "2024-01-01",
		"period_end_date":          "2024-03-31",
		"section_master":           map[string]interface{}{},
		"classification_rules":     map[string]interface{}{},
		"pan_registry":             map[string]interface{}{},
		"vendor_master":            map[string]interface{}{},
		"tds_rate_master":          map[string]interface{}{},
		"tds_threshold_rules":      map[string]interface{}{},
		"deduction_exemptions":     map[string]interface{}{},
		"bank_statements":          map[string]interface{}{},
		"deposits_verification_list": map[string]interface{}{},
		"reconciliation_rules":     map[string]interface{}{},
		"dsc_certificate":          "cert_data",
	}
	err := s.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for valid quarterly input, got: %v", err)
	}
}

func TestTDSReturnFilingSagaValidateInputValidAnnual(t *testing.T) {
	s := NewTDSReturnFilingSaga().(*TDSReturnFilingSaga)
	input := map[string]interface{}{
		"tan":                      "1234567890",
		"filing_period":            "2024-ANNUAL",
		"period_type":              "ANNUAL",
		"period_start_date":        "2023-04-01",
		"period_end_date":          "2024-03-31",
		"section_master":           map[string]interface{}{},
		"classification_rules":     map[string]interface{}{},
		"pan_registry":             map[string]interface{}{},
		"vendor_master":            map[string]interface{}{},
		"tds_rate_master":          map[string]interface{}{},
		"tds_threshold_rules":      map[string]interface{}{},
		"deduction_exemptions":     map[string]interface{}{},
		"bank_statements":          map[string]interface{}{},
		"deposits_verification_list": map[string]interface{}{},
		"reconciliation_rules":     map[string]interface{}{},
		"dsc_certificate":          "cert_data",
	}
	err := s.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for valid annual input, got: %v", err)
	}
}

// ===== CROSS-SAGA VALIDATION TESTS =====

func TestAllStatutorySagasImplementSagaHandler(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewGSTR1FilingSaga(),
		NewGSTR2ITCSaga(),
		NewGSTR9AnnualSaga(),
		NewTDSReturnFilingSaga(),
	}

	for _, s := range sagas {
		if s == nil {
			t.Error("saga handler is nil")
			continue
		}

		if s.SagaType() == "" {
			t.Error("saga type is empty")
		}

		steps := s.GetStepDefinitions()
		if len(steps) == 0 {
			t.Errorf("saga %s has no step definitions", s.SagaType())
		}

		if s.GetStepDefinition(1) == nil {
			t.Errorf("saga %s does not have step 1", s.SagaType())
		}
	}
}

func TestStatutorySagasHaveCompensationSteps(t *testing.T) {
	tests := []struct {
		name string
		saga saga.SagaHandler
	}{
		{"GSTR-1", NewGSTR1FilingSaga()},
		{"GSTR-2", NewGSTR2ITCSaga()},
		{"GSTR-9", NewGSTR9AnnualSaga()},
		{"TDS", NewTDSReturnFilingSaga()},
	}

	for _, test := range tests {
		steps := test.saga.GetStepDefinitions()

		// Find compensation steps (step numbers >= 100)
		hasCompensation := false
		for _, step := range steps {
			if step.StepNumber >= 100 {
				hasCompensation = true
				break
			}
		}

		if !hasCompensation {
			t.Errorf("%s saga should have compensation steps", test.name)
		}
	}
}

func TestStatutorySagasServiceNames(t *testing.T) {
	tests := []struct {
		sagaName string
		saga     saga.SagaHandler
		expectedServices map[string]bool
	}{
		{
			"GSTR-1",
			NewGSTR1FilingSaga(),
			map[string]bool{
				"gst": true,
				"general-ledger": true,
				"compliance-postings": true,
			},
		},
		{
			"GSTR-2",
			NewGSTR2ITCSaga(),
			map[string]bool{
				"gst": true,
				"general-ledger": true,
				"compliance-postings": true,
			},
		},
		{
			"GSTR-9",
			NewGSTR9AnnualSaga(),
			map[string]bool{
				"gst": true,
				"reconciliation": true,
				"general-ledger": true,
				"compliance-postings": true,
			},
		},
		{
			"TDS",
			NewTDSReturnFilingSaga(),
			map[string]bool{
				"tds": true,
				"banking": true,
				"compliance-postings": true,
			},
		},
	}

	for _, test := range tests {
		steps := test.saga.GetStepDefinitions()
		foundServices := make(map[string]bool)

		for _, step := range steps {
			if test.expectedServices[step.ServiceName] {
				foundServices[step.ServiceName] = true
			}
		}

		for service := range test.expectedServices {
			if !foundServices[service] {
				t.Errorf("%s saga missing service %s", test.sagaName, service)
			}
		}
	}
}

func TestStatutorySagasRetryConfigurations(t *testing.T) {
	tests := []struct {
		sagaName string
		saga     saga.SagaHandler
	}{
		{"GSTR-1", NewGSTR1FilingSaga()},
		{"GSTR-2", NewGSTR2ITCSaga()},
		{"GSTR-9", NewGSTR9AnnualSaga()},
		{"TDS", NewTDSReturnFilingSaga()},
	}

	for _, test := range tests {
		steps := test.saga.GetStepDefinitions()

		for _, step := range steps {
			if step.RetryConfig == nil {
				t.Errorf("%s saga step %d has no retry config", test.sagaName, step.StepNumber)
				continue
			}

			if step.RetryConfig.MaxRetries < 1 {
				t.Errorf("%s saga step %d has invalid max retries", test.sagaName, step.StepNumber)
			}

			if step.RetryConfig.MaxBackoffMs < step.RetryConfig.InitialBackoffMs {
				t.Errorf("%s saga step %d has invalid backoff configuration", test.sagaName, step.StepNumber)
			}
		}
	}
}

func TestStatutorySagasInputValidation(t *testing.T) {
	tests := []struct {
		name string
		saga saga.SagaHandler
		validInput map[string]interface{}
		invalidInputs []map[string]interface{}
	}{
		{
			"GSTR-1",
			NewGSTR1FilingSaga(),
			map[string]interface{}{
				"gstin": "29AABCT0959A2Z5",
				"filing_period": "2024-02",
				"period_start_date": "2024-02-01",
				"period_end_date": "2024-02-29",
				"classification_rules": map[string]interface{}{},
				"hsn_master": map[string]interface{}{},
				"tax_rate_master": map[string]interface{}{},
				"declarant_details": map[string]interface{}{},
				"dsc_certificate": "cert",
			},
			[]map[string]interface{}{
				{},
				{"gstin": "short"},
				{"gstin": "29AABCT0959A2Z5", "filing_period": "invalid"},
			},
		},
	}

	for _, test := range tests {
		// Test valid input
		if err := test.saga.ValidateInput(test.validInput); err != nil {
			t.Errorf("%s saga validation failed for valid input: %v", test.name, err)
		}

		// Test invalid inputs
		for i, invalidInput := range test.invalidInputs {
			if err := test.saga.ValidateInput(invalidInput); err == nil {
				t.Errorf("%s saga validation should fail for invalid input %d", test.name, i)
			}
		}
	}
}

// ===== FX MODULE TESTS =====

func TestStatutorySagasModuleProviders(t *testing.T) {
	sagas := ProvideStatutorySagaHandlers()

	expectedCount := 4
	if len(sagas) != expectedCount {
		t.Errorf("expected %d sagas from provider, got %d", expectedCount, len(sagas))
	}

	expectedTypes := []string{"SAGA-ST01", "SAGA-ST02", "SAGA-ST03", "SAGA-ST04"}
	for i, expectedType := range expectedTypes {
		if i >= len(sagas) {
			t.Errorf("missing saga at index %d", i)
			continue
		}
		if sagas[i].SagaType() != expectedType {
			t.Errorf("expected saga type %s at index %d, got %s", expectedType, i, sagas[i].SagaType())
		}
	}
}

// ===== TAX COMPLIANCE SCENARIO TESTS =====

func TestGSTR1FilingSagaTaxScenarios(t *testing.T) {
	s := NewGSTR1FilingSaga().(*GSTR1FilingSaga)

	// Scenario 1: B2B transaction
	input := map[string]interface{}{
		"gstin": "29AABCT0959A2Z5",
		"filing_period": "2024-02",
		"period_start_date": "2024-02-01",
		"period_end_date": "2024-02-29",
		"classification_rules": map[string]interface{}{"b2b": true},
		"hsn_master": map[string]interface{}{},
		"tax_rate_master": map[string]interface{}{},
		"declarant_details": map[string]interface{}{},
		"dsc_certificate": "cert",
	}

	if err := s.ValidateInput(input); err != nil {
		t.Errorf("failed validation for B2B scenario: %v", err)
	}
}

func TestGSTR2ITCSagaTaxScenarios(t *testing.T) {
	s := NewGSTR2ITCSaga().(*GSTR2ITCSaga)

	// Scenario 1: ITC claim with eligible supplies
	input := map[string]interface{}{
		"gstin": "29AABCT0959A2Z5",
		"filing_period": "2024-02",
		"period_start_date": "2024-02-01",
		"period_end_date": "2024-02-29",
		"eligibility_rules": map[string]interface{}{},
		"exempt_supply_list": map[string]interface{}{},
		"vendor_master": map[string]interface{}{},
		"gstin_registry": map[string]interface{}{},
		"itc_calculation_rules": map[string]interface{}{},
		"reversal_rules": map[string]interface{}{},
		"personal_use_list": map[string]interface{}{},
		"output_tax_liability": 100000.0,
		"previous_period_itc": map[string]interface{}{},
		"previous_credit_adjustments": map[string]interface{}{},
	}

	if err := s.ValidateInput(input); err != nil {
		t.Errorf("failed validation for ITC claim scenario: %v", err)
	}
}

func TestTDSReturnFilingSagaTaxScenarios(t *testing.T) {
	s := NewTDSReturnFilingSaga().(*TDSReturnFilingSaga)

	// Scenario 1: TDS deduction for contractor (194C)
	input := map[string]interface{}{
		"tan": "1234567890",
		"filing_period": "2024-Q1",
		"period_type": "QUARTERLY",
		"period_start_date": "2024-01-01",
		"period_end_date": "2024-03-31",
		"section_master": map[string]interface{}{"194C": "Contractors"},
		"classification_rules": map[string]interface{}{},
		"pan_registry": map[string]interface{}{},
		"vendor_master": map[string]interface{}{},
		"tds_rate_master": map[string]interface{}{},
		"tds_threshold_rules": map[string]interface{}{},
		"deduction_exemptions": map[string]interface{}{},
		"bank_statements": map[string]interface{}{},
		"deposits_verification_list": map[string]interface{}{},
		"reconciliation_rules": map[string]interface{}{},
		"dsc_certificate": "cert",
	}

	if err := s.ValidateInput(input); err != nil {
		t.Errorf("failed validation for TDS contractor scenario: %v", err)
	}
}
