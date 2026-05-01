// Package asset provides saga handlers for asset management workflows
package asset

import (
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// ============================================================================
// SAGA-A01: Asset Acquisition Tests (14 tests)
// ============================================================================

func TestAssetAcquisitionSagaType(t *testing.T) {
	saga := NewAssetAcquisitionSaga()
	if saga.SagaType() != "SAGA-A01" {
		t.Errorf("expected SAGA-A01, got %s", saga.SagaType())
	}
}

func TestAssetAcquisitionGetStepDefinitions(t *testing.T) {
	saga := NewAssetAcquisitionSaga()
	steps := saga.GetStepDefinitions()
	if len(steps) != 10 {
		t.Errorf("expected 10 steps, got %d", len(steps))
	}
}

func TestAssetAcquisitionGetStepDefinition(t *testing.T) {
	saga := NewAssetAcquisitionSaga()
	step := saga.GetStepDefinition(1)
	if step == nil {
		t.Error("expected step 1 to exist")
	}
	if step.StepNumber != 1 {
		t.Errorf("expected step number 1, got %d", step.StepNumber)
	}
	if step.ServiceName != "purchase-order" {
		t.Errorf("expected service purchase-order, got %s", step.ServiceName)
	}
}

func TestAssetAcquisitionCriticalSteps(t *testing.T) {
	saga := NewAssetAcquisitionSaga()
	criticalSteps := []int{3, 5, 7, 8}
	for _, stepNum := range criticalSteps {
		step := saga.GetStepDefinition(stepNum)
		if !step.IsCritical {
			t.Errorf("step %d should be critical", stepNum)
		}
	}
}

func TestAssetAcquisitionCompensationSteps(t *testing.T) {
	saga := NewAssetAcquisitionSaga()
	step3 := saga.GetStepDefinition(3)
	if len(step3.CompensationSteps) != 1 || step3.CompensationSteps[0] != 102 {
		t.Error("step 3 should have compensation step 102")
	}
}

func TestAssetAcquisitionValidateInputMissingPOID(t *testing.T) {
	saga := NewAssetAcquisitionSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"po_line_id":       "POL001",
			"asset_tag":        "ASSET001",
			"asset_category":   "Building",
			"po_amount":        100000.0,
			"useful_life_years": 10.0,
			"asset_location":   "Warehouse",
			"received_date":    "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: po_id" {
		t.Errorf("expected 'missing required field: po_id', got %v", err)
	}
}

func TestAssetAcquisitionValidateInputMissingAssetTag(t *testing.T) {
	saga := NewAssetAcquisitionSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"po_id":             "PO001",
			"po_line_id":        "POL001",
			"asset_category":    "Building",
			"po_amount":         100000.0,
			"useful_life_years": 10.0,
			"asset_location":    "Warehouse",
			"received_date":     "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: asset_tag" {
		t.Errorf("expected 'missing required field: asset_tag', got %v", err)
	}
}

func TestAssetAcquisitionValidateInputInvalidPoAmount(t *testing.T) {
	saga := NewAssetAcquisitionSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"po_id":             "PO001",
			"po_line_id":        "POL001",
			"asset_tag":         "ASSET001",
			"asset_category":    "Building",
			"po_amount":         0.0,
			"useful_life_years": 10.0,
			"asset_location":    "Warehouse",
			"received_date":     "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "po_amount must be a positive number" {
		t.Errorf("expected 'po_amount must be a positive number', got %v", err)
	}
}

func TestAssetAcquisitionValidateInputInvalidUsefulLife(t *testing.T) {
	saga := NewAssetAcquisitionSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"po_id":             "PO001",
			"po_line_id":        "POL001",
			"asset_tag":         "ASSET001",
			"asset_category":    "Building",
			"po_amount":         100000.0,
			"useful_life_years": -5.0,
			"asset_location":    "Warehouse",
			"received_date":     "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "useful_life_years must be a positive number" {
		t.Errorf("expected 'useful_life_years must be a positive number', got %v", err)
	}
}

func TestAssetAcquisitionValidateInputValid(t *testing.T) {
	saga := NewAssetAcquisitionSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"po_id":             "PO001",
			"po_line_id":        "POL001",
			"asset_tag":         "ASSET001",
			"asset_category":    "Building",
			"po_amount":         100000.0,
			"useful_life_years": 10.0,
			"asset_location":    "Warehouse",
			"received_date":     "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAssetAcquisitionRetryConfig(t *testing.T) {
	saga := NewAssetAcquisitionSaga()
	step := saga.GetStepDefinition(1)
	if step.RetryConfig.MaxRetries != 3 {
		t.Errorf("expected 3 retries, got %d", step.RetryConfig.MaxRetries)
	}
	if step.RetryConfig.InitialBackoffMs != 1000 {
		t.Errorf("expected 1000ms initial backoff, got %d", step.RetryConfig.InitialBackoffMs)
	}
	if step.RetryConfig.MaxBackoffMs != 30000 {
		t.Errorf("expected 30000ms max backoff, got %d", step.RetryConfig.MaxBackoffMs)
	}
}

func TestAssetAcquisitionInputMappings(t *testing.T) {
	saga := NewAssetAcquisitionSaga()
	step1 := saga.GetStepDefinition(1)
	if step1.InputMapping["tenantID"] != "$.tenantID" {
		t.Error("tenantID mapping incorrect")
	}
	if step1.InputMapping["poID"] != "$.input.po_id" {
		t.Error("poID mapping incorrect")
	}
}

// ============================================================================
// SAGA-A02: Asset Depreciation Tests (13 tests)
// ============================================================================

func TestAssetDepreciationSagaType(t *testing.T) {
	saga := NewAssetDepreciationSaga()
	if saga.SagaType() != "SAGA-A02" {
		t.Errorf("expected SAGA-A02, got %s", saga.SagaType())
	}
}

func TestAssetDepreciationGetStepDefinitions(t *testing.T) {
	saga := NewAssetDepreciationSaga()
	steps := saga.GetStepDefinitions()
	if len(steps) != 8 {
		t.Errorf("expected 8 steps, got %d", len(steps))
	}
}

func TestAssetDepreciationCriticalSteps(t *testing.T) {
	saga := NewAssetDepreciationSaga()
	criticalSteps := []int{3, 6, 7}
	for _, stepNum := range criticalSteps {
		step := saga.GetStepDefinition(stepNum)
		if !step.IsCritical {
			t.Errorf("step %d should be critical", stepNum)
		}
	}
}

func TestAssetDepreciationNonCriticalSteps(t *testing.T) {
	saga := NewAssetDepreciationSaga()
	nonCriticalSteps := []int{4, 5, 8}
	for _, stepNum := range nonCriticalSteps {
		step := saga.GetStepDefinition(stepNum)
		if step.IsCritical {
			t.Errorf("step %d should not be critical", stepNum)
		}
	}
}

func TestAssetDepreciationValidateInputMissingDepreciationDate(t *testing.T) {
	saga := NewAssetDepreciationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"period_id": "PERIOD202602",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: depreciation_date" {
		t.Errorf("expected 'missing required field: depreciation_date', got %v", err)
	}
}

func TestAssetDepreciationValidateInputMissingPeriodID(t *testing.T) {
	saga := NewAssetDepreciationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"depreciation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: period_id" {
		t.Errorf("expected 'missing required field: period_id', got %v", err)
	}
}

func TestAssetDepreciationValidateInputValid(t *testing.T) {
	saga := NewAssetDepreciationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"depreciation_date":            "2026-02-16",
			"period_id":                    "PERIOD202602",
			"depreciation_expense_account": "6100",
			"accumulated_depreciation_account": "1500",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAssetDepreciationServiceNames(t *testing.T) {
	saga := NewAssetDepreciationSaga()
	expectedServices := map[int]string{
		1: "asset",
		2: "depreciation",
		3: "fixed-assets",
		4: "depreciation",
		5: "depreciation",
		6: "general-ledger",
		7: "asset",
		8: "depreciation",
	}
	for stepNum, expectedService := range expectedServices {
		step := saga.GetStepDefinition(stepNum)
		if step.ServiceName != expectedService {
			t.Errorf("step %d: expected service %s, got %s", stepNum, expectedService, step.ServiceName)
		}
	}
}

func TestAssetDepreciationTimeouts(t *testing.T) {
	saga := NewAssetDepreciationSaga()
	step1 := saga.GetStepDefinition(1)
	if step1.TimeoutSeconds != 45 {
		t.Errorf("step 1: expected 45s timeout, got %d", step1.TimeoutSeconds)
	}
	step6 := saga.GetStepDefinition(6)
	if step6.TimeoutSeconds != 30 {
		t.Errorf("step 6: expected 30s timeout, got %d", step6.TimeoutSeconds)
	}
}

func TestAssetDepreciationGetStepDefinitionNotFound(t *testing.T) {
	saga := NewAssetDepreciationSaga()
	step := saga.GetStepDefinition(99)
	if step != nil {
		t.Errorf("expected nil for non-existent step, got %v", step)
	}
}

// ============================================================================
// SAGA-A03: Asset Disposal Tests (14 tests)
// ============================================================================

func TestAssetDisposalSagaType(t *testing.T) {
	saga := NewAssetDisposalSaga()
	if saga.SagaType() != "SAGA-A03" {
		t.Errorf("expected SAGA-A03, got %s", saga.SagaType())
	}
}

func TestAssetDisposalGetStepDefinitions(t *testing.T) {
	saga := NewAssetDisposalSaga()
	steps := saga.GetStepDefinitions()
	if len(steps) != 9 {
		t.Errorf("expected 9 steps, got %d", len(steps))
	}
}

func TestAssetDisposalCriticalSteps(t *testing.T) {
	saga := NewAssetDisposalSaga()
	criticalSteps := []int{1, 2, 4, 6, 8}
	for _, stepNum := range criticalSteps {
		step := saga.GetStepDefinition(stepNum)
		if !step.IsCritical {
			t.Errorf("step %d should be critical", stepNum)
		}
	}
}

func TestAssetDisposalValidateInputMissingAssetID(t *testing.T) {
	saga := NewAssetDisposalSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"disposal_type": "SALE",
			"disposal_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: asset_id" {
		t.Errorf("expected 'missing required field: asset_id', got %v", err)
	}
}

func TestAssetDisposalValidateInputMissingDisposalType(t *testing.T) {
	saga := NewAssetDisposalSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"asset_id":       "ASSET001",
			"disposal_date":  "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: disposal_type" {
		t.Errorf("expected 'missing required field: disposal_type', got %v", err)
	}
}

func TestAssetDisposalValidateInputInvalidDisposalType(t *testing.T) {
	saga := NewAssetDisposalSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"asset_id":      "ASSET001",
			"disposal_type": "INVALID",
			"disposal_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil {
		t.Error("expected error for invalid disposal_type")
	}
}

func TestAssetDisposalValidateInputSaleWithoutPrice(t *testing.T) {
	saga := NewAssetDisposalSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"asset_id":       "ASSET001",
			"disposal_type":  "SALE",
			"disposal_date":  "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "sale_price is required for SALE disposal type" {
		t.Errorf("expected 'sale_price is required', got %v", err)
	}
}

func TestAssetDisposalValidateInputSaleWithoutBuyer(t *testing.T) {
	saga := NewAssetDisposalSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"asset_id":       "ASSET001",
			"disposal_type":  "SALE",
			"disposal_date":  "2026-02-16",
			"sale_price":     50000.0,
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "buyer_name is required for SALE disposal type" {
		t.Errorf("expected 'buyer_name is required', got %v", err)
	}
}

func TestAssetDisposalValidateInputSaleValid(t *testing.T) {
	saga := NewAssetDisposalSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"asset_id":       "ASSET001",
			"disposal_type":  "SALE",
			"disposal_date":  "2026-02-16",
			"sale_price":     50000.0,
			"buyer_name":     "ABC Company",
			"buyer_type":     "CORPORATE",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for valid SALE input, got %v", err)
	}
}

func TestAssetDisposalValidateInputScrapValid(t *testing.T) {
	saga := NewAssetDisposalSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"asset_id":       "ASSET001",
			"disposal_type":  "SCRAP",
			"disposal_date":  "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for valid SCRAP input, got %v", err)
	}
}

func TestAssetDisposalValidateInputDonateValid(t *testing.T) {
	saga := NewAssetDisposalSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"asset_id":       "ASSET001",
			"disposal_type":  "DONATE",
			"disposal_date":  "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for valid DONATE input, got %v", err)
	}
}

// ============================================================================
// SAGA-A04: Asset Revaluation Tests (12 tests)
// ============================================================================

func TestAssetRevaluationSagaType(t *testing.T) {
	saga := NewAssetRevaluationSaga()
	if saga.SagaType() != "SAGA-A04" {
		t.Errorf("expected SAGA-A04, got %s", saga.SagaType())
	}
}

func TestAssetRevaluationGetStepDefinitions(t *testing.T) {
	saga := NewAssetRevaluationSaga()
	steps := saga.GetStepDefinitions()
	if len(steps) != 7 {
		t.Errorf("expected 7 steps, got %d", len(steps))
	}
}

func TestAssetRevaluationCriticalSteps(t *testing.T) {
	saga := NewAssetRevaluationSaga()
	criticalSteps := []int{1, 2, 3, 5}
	for _, stepNum := range criticalSteps {
		step := saga.GetStepDefinition(stepNum)
		if !step.IsCritical {
			t.Errorf("step %d should be critical", stepNum)
		}
	}
}

func TestAssetRevaluationValidateInputMissingAssetID(t *testing.T) {
	saga := NewAssetRevaluationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"revaluation_date": "2026-02-16",
			"valuation_method": "APPRAISAL",
			"appraisal_amount":  150000.0,
			"prior_book_value": 100000.0,
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: asset_id" {
		t.Errorf("expected 'missing required field: asset_id', got %v", err)
	}
}

func TestAssetRevaluationValidateInputMissingRevaluationDate(t *testing.T) {
	saga := NewAssetRevaluationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"asset_id":         "ASSET001",
			"valuation_method": "APPRAISAL",
			"appraisal_amount":  150000.0,
			"prior_book_value": 100000.0,
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: revaluation_date" {
		t.Errorf("expected 'missing required field: revaluation_date', got %v", err)
	}
}

func TestAssetRevaluationValidateInputInvalidValuationMethod(t *testing.T) {
	saga := NewAssetRevaluationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"asset_id":         "ASSET001",
			"revaluation_date": "2026-02-16",
			"valuation_method": "INVALID_METHOD",
			"appraisal_amount":  150000.0,
			"prior_book_value": 100000.0,
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil {
		t.Error("expected error for invalid valuation_method")
	}
}

func TestAssetRevaluationValidateInputValidMethods(t *testing.T) {
	saga := NewAssetRevaluationSaga()
	validMethods := []string{"MARKET", "APPRAISAL", "INCOME", "COST"}
	for _, method := range validMethods {
		input := map[string]interface{}{
			"input": map[string]interface{}{
				"asset_id":         "ASSET001",
				"revaluation_date": "2026-02-16",
				"valuation_method": method,
				"appraisal_amount":  150000.0,
				"prior_book_value": 100000.0,
			},
			"companyID": "COMP001",
		}
		err := saga.ValidateInput(input)
		if err != nil {
			t.Errorf("expected no error for method %s, got %v", method, err)
		}
	}
}

func TestAssetRevaluationValidateInputNegativeAppraisalAmount(t *testing.T) {
	saga := NewAssetRevaluationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"asset_id":         "ASSET001",
			"revaluation_date": "2026-02-16",
			"valuation_method": "APPRAISAL",
			"appraisal_amount":  -50000.0,
			"prior_book_value": 100000.0,
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "appraisal_amount must be a positive number" {
		t.Errorf("expected 'appraisal_amount must be a positive number', got %v", err)
	}
}

func TestAssetRevaluationValidateInputInvalidTriggerType(t *testing.T) {
	saga := NewAssetRevaluationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"asset_id":         "ASSET001",
			"revaluation_date": "2026-02-16",
			"valuation_method": "APPRAISAL",
			"appraisal_amount":  150000.0,
			"prior_book_value": 100000.0,
			"trigger_type":     "INVALID_TRIGGER",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil {
		t.Error("expected error for invalid trigger_type")
	}
}

func TestAssetRevaluationValidateInputValid(t *testing.T) {
	saga := NewAssetRevaluationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"asset_id":         "ASSET001",
			"revaluation_date": "2026-02-16",
			"valuation_method": "APPRAISAL",
			"appraisal_amount":  150000.0,
			"prior_book_value": 100000.0,
			"trigger_type":     "ANNUAL",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// ============================================================================
// Integration Tests (5 tests)
// ============================================================================

func TestAllSagasImplementSagaHandler(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewAssetAcquisitionSaga(),
		NewAssetDepreciationSaga(),
		NewAssetDisposalSaga(),
		NewAssetRevaluationSaga(),
	}
	for _, s := range sagas {
		if s.SagaType() == "" {
			t.Error("saga type should not be empty")
		}
		if len(s.GetStepDefinitions()) == 0 {
			t.Errorf("%s should have at least one step", s.SagaType())
		}
	}
}

func TestAllSagasHaveValidServiceNames(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewAssetAcquisitionSaga(),
		NewAssetDepreciationSaga(),
		NewAssetDisposalSaga(),
		NewAssetRevaluationSaga(),
	}
	validServices := map[string]bool{
		"asset":                      true,
		"purchase-order":             true,
		"fixed-assets":               true,
		"depreciation":               true,
		"general-ledger":             true,
		"notification":               true,
		"accounts-receivable":        true,
		"approval":                   true,
	}
	for _, saga := range sagas {
		for _, step := range saga.GetStepDefinitions() {
			if !validServices[step.ServiceName] {
				t.Errorf("%s: invalid service name %s", saga.SagaType(), step.ServiceName)
			}
		}
	}
}

func TestAllSagasHaveRetryConfigs(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewAssetAcquisitionSaga(),
		NewAssetDepreciationSaga(),
		NewAssetDisposalSaga(),
		NewAssetRevaluationSaga(),
	}
	for _, saga := range sagas {
		for i, step := range saga.GetStepDefinitions() {
			if step.RetryConfig == nil {
				t.Errorf("%s step %d: retry config should not be nil", saga.SagaType(), i+1)
			}
			if step.RetryConfig.MaxRetries <= 0 {
				t.Errorf("%s step %d: max retries should be positive", saga.SagaType(), i+1)
			}
		}
	}
}

func TestAllSagasHaveTimeouts(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewAssetAcquisitionSaga(),
		NewAssetDepreciationSaga(),
		NewAssetDisposalSaga(),
		NewAssetRevaluationSaga(),
	}
	for _, saga := range sagas {
		for i, step := range saga.GetStepDefinitions() {
			if step.TimeoutSeconds <= 0 {
				t.Errorf("%s step %d: timeout should be positive", saga.SagaType(), i+1)
			}
		}
	}
}

func TestProvideAssetSagaHandlers(t *testing.T) {
	handlers := ProvideAssetSagaHandlers()
	if len(handlers) != 4 {
		t.Errorf("expected 4 saga handlers, got %d", len(handlers))
	}
	sagaTypes := make(map[string]bool)
	for _, handler := range handlers {
		sagaTypes[handler.SagaType()] = true
	}
	expectedTypes := map[string]bool{"SAGA-A01": true, "SAGA-A02": true, "SAGA-A03": true, "SAGA-A04": true}
	for expectedType := range expectedTypes {
		if !sagaTypes[expectedType] {
			t.Errorf("expected saga type %s not found", expectedType)
		}
	}
}
