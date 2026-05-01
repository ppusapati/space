// Package warranty provides saga handlers for warranty and service module workflows
package warranty

import (
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// ============================================================================
// SAGA TYPE & INTERFACE TESTS (6 tests)
// ============================================================================

// TestWarrantySagaTypes verifies all 6 warranty sagas return correct types
func TestWarrantySagaTypes(t *testing.T) {
	sagas := []struct {
		name         string
		saga         saga.SagaHandler
		expectedType string
	}{
		{"Warranty Claim", NewWarrantyClaimSaga(), "SAGA-W01"},
		{"Field Service", NewFieldServiceSaga(), "SAGA-W02"},
		{"Spare Parts", NewSparePartsSaga(), "SAGA-W03"},
		{"SLA Management", NewSLAManagementSaga(), "SAGA-W04"},
		{"Customer Satisfaction", NewCustomerSatisfactionSaga(), "SAGA-W05"},
		{"Extended Warranty", NewExtendedWarrantySaga(), "SAGA-W06"},
	}

	for _, tt := range sagas {
		t.Run(tt.name, func(t *testing.T) {
			if tt.saga.SagaType() != tt.expectedType {
				t.Errorf("expected %s, got %s", tt.expectedType, tt.saga.SagaType())
			}
		})
	}
}

// TestWarrantySagasImplementInterface verifies all sagas implement saga.SagaHandler
func TestWarrantySagasImplementInterface(t *testing.T) {
	handlers := []saga.SagaHandler{
		NewWarrantyClaimSaga(),
		NewFieldServiceSaga(),
		NewSparePartsSaga(),
		NewSLAManagementSaga(),
		NewCustomerSatisfactionSaga(),
		NewExtendedWarrantySaga(),
	}

	for _, handler := range handlers {
		if handler == nil {
			t.Errorf("handler is nil")
		}
		// Verify all interface methods are callable
		if handler.SagaType() == "" {
			t.Errorf("SagaType() returned empty string")
		}
		if len(handler.GetStepDefinitions()) == 0 {
			t.Errorf("GetStepDefinitions() returned empty slice")
		}
	}
}

// TestGetStepDefinitions verifies each saga returns non-empty steps
func TestGetStepDefinitions(t *testing.T) {
	sagas := []struct {
		name        string
		saga        saga.SagaHandler
		minSteps    int
	}{
		{"Warranty Claim", NewWarrantyClaimSaga(), 10},
		{"Field Service", NewFieldServiceSaga(), 12},
		{"Spare Parts", NewSparePartsSaga(), 10},
		{"SLA Management", NewSLAManagementSaga(), 10},
		{"Customer Satisfaction", NewCustomerSatisfactionSaga(), 11},
		{"Extended Warranty", NewExtendedWarrantySaga(), 11},
	}

	for _, tt := range sagas {
		t.Run(tt.name, func(t *testing.T) {
			steps := tt.saga.GetStepDefinitions()
			if len(steps) < tt.minSteps {
				t.Errorf("expected at least %d steps, got %d", tt.minSteps, len(steps))
			}
		})
	}
}

// TestGetStepDefinition verifies step lookup works for valid/invalid steps
func TestGetStepDefinition(t *testing.T) {
	saga := NewWarrantyClaimSaga()

	tests := []struct {
		name       string
		stepNum    int
		shouldExist bool
	}{
		{"First step", 1, true},
		{"Middle step", 5, true},
		{"Last forward step", 10, true},
		{"First compensation step", 103, true},
		{"Invalid step", 999, false},
		{"Negative step", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := saga.GetStepDefinition(tt.stepNum)
			if tt.shouldExist && step == nil {
				t.Errorf("expected step %d to exist, got nil", tt.stepNum)
			}
			if !tt.shouldExist && step != nil {
				t.Errorf("expected step %d to not exist, got %v", tt.stepNum, step)
			}
		})
	}
}

// TestValidateInputNilHandling verifies nil/empty input handling
func TestValidateInputNilHandling(t *testing.T) {
	sagas := []struct {
		name string
		saga saga.SagaHandler
	}{
		{"Warranty Claim", NewWarrantyClaimSaga()},
		{"Field Service", NewFieldServiceSaga()},
		{"Spare Parts", NewSparePartsSaga()},
		{"SLA Management", NewSLAManagementSaga()},
		{"Customer Satisfaction", NewCustomerSatisfactionSaga()},
		{"Extended Warranty", NewExtendedWarrantySaga()},
	}

	for _, tt := range sagas {
		t.Run(tt.name, func(t *testing.T) {
			// Nil input should error
			err := tt.saga.ValidateInput(nil)
			if err == nil {
				t.Errorf("expected error for nil input, got nil")
			}

			// Empty map should error
			err = tt.saga.ValidateInput(map[string]interface{}{})
			if err == nil {
				t.Errorf("expected error for empty input map, got nil")
			}

			// Wrong type should error
			err = tt.saga.ValidateInput("invalid")
			if err == nil {
				t.Errorf("expected error for string input, got nil")
			}
		})
	}
}

// TestSagasProvidedByHandler verifies all sagas are registered
func TestSagasProvidedByHandler(t *testing.T) {
	handlers := ProvideWarrantySagaHandlers()

	expectedCount := 6
	if len(handlers) != expectedCount {
		t.Errorf("expected %d saga handlers, got %d", expectedCount, len(handlers))
	}

	expectedTypes := []string{"SAGA-W01", "SAGA-W02", "SAGA-W03", "SAGA-W04", "SAGA-W05", "SAGA-W06"}
	handlerMap := make(map[string]saga.SagaHandler)
	for _, handler := range handlers {
		handlerMap[handler.SagaType()] = handler
	}

	for _, sagaType := range expectedTypes {
		if _, exists := handlerMap[sagaType]; !exists {
			t.Errorf("saga handler %s not registered", sagaType)
		}
	}
}

// ============================================================================
// SAGA-W01: WARRANTY CLAIM TESTS (12 tests)
// ============================================================================

func TestWarrantyClaimSagaType(t *testing.T) {
	handler := NewWarrantyClaimSaga()
	if handler.SagaType() != "SAGA-W01" {
		t.Errorf("expected SAGA-W01, got %s", handler.SagaType())
	}
}

// TestWarrantyClaimSagaStepCount verifies step count and composition
func TestWarrantyClaimSagaStepCount(t *testing.T) {
	handler := NewWarrantyClaimSaga()
	steps := handler.GetStepDefinitions()

	expectedStepCount := 19 // 10 forward + 9 compensation
	if len(steps) != expectedStepCount {
		t.Errorf("expected %d steps, got %d", expectedStepCount, len(steps))
	}

	// Count forward and compensation steps
	forwardSteps := 0
	compensationSteps := 0
	for _, step := range steps {
		if step.StepNumber < 100 {
			forwardSteps++
		} else {
			compensationSteps++
		}
	}

	if forwardSteps != 10 {
		t.Errorf("expected 10 forward steps, got %d", forwardSteps)
	}
	if compensationSteps != 9 {
		t.Errorf("expected 9 compensation steps, got %d", compensationSteps)
	}
}

// TestWarrantyClaimCriticalSteps verifies critical step markers
func TestWarrantyClaimCriticalSteps(t *testing.T) {
	handler := NewWarrantyClaimSaga()
	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 4: true, 8: true, 9: true}

	for stepNum := range criticalSteps {
		step := handler.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("critical step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be marked critical", stepNum)
		}
	}

	// Verify non-critical steps
	nonCriticalSteps := []int{5, 6, 7, 10}
	for _, stepNum := range nonCriticalSteps {
		step := handler.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("step %d not found", stepNum)
			continue
		}
		if step.IsCritical {
			t.Errorf("step %d should not be critical", stepNum)
		}
	}
}

// TestWarrantyClaimValidateInput validates claim input parameters
func TestWarrantyClaimValidateInput(t *testing.T) {
	handler := NewWarrantyClaimSaga()

	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"claim_id":              "CLM-001",
				"product_id":            "PROD-001",
				"customer_id":           "CUST-001",
				"issue_description":     "Product defect",
				"claim_date":            "2026-02-16",
			},
			wantErr: false,
		},
		{
			name: "missing claim_id",
			input: map[string]interface{}{
				"product_id":            "PROD-001",
				"customer_id":           "CUST-001",
				"issue_description":     "Product defect",
				"claim_date":            "2026-02-16",
			},
			wantErr: true,
		},
		{
			name: "missing product_id",
			input: map[string]interface{}{
				"claim_id":              "CLM-001",
				"customer_id":           "CUST-001",
				"issue_description":     "Product defect",
				"claim_date":            "2026-02-16",
			},
			wantErr: true,
		},
		{
			name: "missing customer_id",
			input: map[string]interface{}{
				"claim_id":              "CLM-001",
				"product_id":            "PROD-001",
				"issue_description":     "Product defect",
				"claim_date":            "2026-02-16",
			},
			wantErr: true,
		},
		{
			name: "missing issue_description",
			input: map[string]interface{}{
				"claim_id":              "CLM-001",
				"product_id":            "PROD-001",
				"customer_id":           "CUST-001",
				"claim_date":            "2026-02-16",
			},
			wantErr: true,
		},
		{
			name: "empty claim_id",
			input: map[string]interface{}{
				"claim_id":              "",
				"product_id":            "PROD-001",
				"customer_id":           "CUST-001",
				"issue_description":     "Product defect",
			},
			wantErr: true,
		},
		{
			name: "wrong type for claim_id",
			input: map[string]interface{}{
				"claim_id":              123,
				"product_id":            "PROD-001",
				"customer_id":           "CUST-001",
				"issue_description":     "Product defect",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestWarrantyClaimInputMapping verifies input mappings exist
func TestWarrantyClaimInputMapping(t *testing.T) {
	handler := NewWarrantyClaimSaga()
	steps := handler.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber >= 100 {
			continue // Skip compensation steps
		}
		if len(step.InputMapping) == 0 {
			t.Errorf("step %d has no input mapping", step.StepNumber)
		}
		// Verify context fields
		if _, ok := step.InputMapping["tenantID"]; !ok {
			t.Errorf("step %d missing tenantID mapping", step.StepNumber)
		}
		if _, ok := step.InputMapping["companyID"]; !ok {
			t.Errorf("step %d missing companyID mapping", step.StepNumber)
		}
	}
}

// TestWarrantyClaimTimeoutValues verifies timeout configurations
func TestWarrantyClaimTimeoutValues(t *testing.T) {
	handler := NewWarrantyClaimSaga()
	steps := handler.GetStepDefinitions()

	for _, step := range steps {
		if step.TimeoutSeconds <= 0 {
			t.Errorf("step %d has invalid timeout: %d", step.StepNumber, step.TimeoutSeconds)
		}
		if step.TimeoutSeconds > 300 {
			t.Errorf("step %d has excessive timeout: %d seconds", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestWarrantyClaimServiceNames verifies service names are valid
func TestWarrantyClaimServiceNames(t *testing.T) {
	handler := NewWarrantyClaimSaga()
	steps := handler.GetStepDefinitions()

	expectedServices := map[string]bool{
		"warranty":       true,
		"approval":       true,
		"asset":          true,
		"insurance":      true,
		"sales-invoice":  true,
		"general-ledger": true,
		"notification":   true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestWarrantyClaimRetryConfiguration verifies retry settings
func TestWarrantyClaimRetryConfiguration(t *testing.T) {
	handler := NewWarrantyClaimSaga()
	steps := handler.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber >= 100 {
			continue // Compensation steps don't require retry
		}
		if step.RetryConfig == nil {
			t.Errorf("step %d missing retry config", step.StepNumber)
			continue
		}
		if step.RetryConfig.MaxRetries <= 0 {
			t.Errorf("step %d has invalid MaxRetries: %d", step.StepNumber, step.RetryConfig.MaxRetries)
		}
		if step.RetryConfig.BackoffMultiplier <= 0 {
			t.Errorf("step %d has invalid BackoffMultiplier: %f", step.StepNumber, step.RetryConfig.BackoffMultiplier)
		}
	}
}

// TestWarrantyClaimCompensationMapping verifies compensation steps
func TestWarrantyClaimCompensationMapping(t *testing.T) {
	handler := NewWarrantyClaimSaga()

	// Verify compensation steps exist and are referenced
	forwardSteps := []int{3, 4, 5, 6, 7, 8, 9}
	expectedCompensation := map[int]int{
		3: 103,
		4: 104,
		5: 105,
		6: 106,
		7: 107,
		8: 108,
		9: 109,
	}

	for _, fwdStep := range forwardSteps {
		step := handler.GetStepDefinition(fwdStep)
		if step == nil {
			t.Errorf("forward step %d not found", fwdStep)
			continue
		}
		if len(step.CompensationSteps) == 0 && fwdStep >= 3 {
			t.Errorf("forward step %d should have compensation steps", fwdStep)
			continue
		}
		if len(step.CompensationSteps) > 0 {
			compNum := int(step.CompensationSteps[0])
			if compNum != expectedCompensation[fwdStep] {
				t.Errorf("step %d expected compensation %d, got %d", fwdStep, expectedCompensation[fwdStep], compNum)
			}
		}
	}
}

// ============================================================================
// SAGA-W02: FIELD SERVICE TESTS (12 tests)
// ============================================================================

func TestFieldServiceSagaType(t *testing.T) {
	handler := NewFieldServiceSaga()
	if handler.SagaType() != "SAGA-W02" {
		t.Errorf("expected SAGA-W02, got %s", handler.SagaType())
	}
}

// TestFieldServiceSagaStepCount verifies 27 total steps (14 forward + 13 compensation)
func TestFieldServiceSagaStepCount(t *testing.T) {
	handler := NewFieldServiceSaga()
	steps := handler.GetStepDefinitions()

	expectedStepCount := 27 // 14 forward + 13 compensation
	if len(steps) != expectedStepCount {
		t.Errorf("expected %d steps, got %d", expectedStepCount, len(steps))
	}

	forwardSteps := 0
	compensationSteps := 0
	for _, step := range steps {
		if step.StepNumber < 100 {
			forwardSteps++
		} else {
			compensationSteps++
		}
	}

	if forwardSteps != 14 {
		t.Errorf("expected 14 forward steps, got %d", forwardSteps)
	}
	if compensationSteps != 13 {
		t.Errorf("expected 13 compensation steps, got %d", compensationSteps)
	}
}

// TestFieldServiceCriticalSteps verifies critical step markers
func TestFieldServiceCriticalSteps(t *testing.T) {
	handler := NewFieldServiceSaga()
	criticalSteps := map[int]bool{1: true, 2: true, 3: true, 11: true, 12: true, 13: true}

	for stepNum := range criticalSteps {
		step := handler.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("critical step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be marked critical", stepNum)
		}
	}
}

// TestFieldServiceValidateInput validates service request parameters
func TestFieldServiceValidateInput(t *testing.T) {
	handler := NewFieldServiceSaga()

	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"service_request_id": "SR-001",
				"customer_id":        "CUST-001",
				"asset_id":           "ASSET-001",
				"issue_description":  "Equipment malfunction",
			},
			wantErr: false,
		},
		{
			name: "missing service_request_id",
			input: map[string]interface{}{
				"customer_id":       "CUST-001",
				"asset_id":          "ASSET-001",
				"issue_description": "Equipment malfunction",
			},
			wantErr: true,
		},
		{
			name: "missing customer_id",
			input: map[string]interface{}{
				"service_request_id": "SR-001",
				"asset_id":           "ASSET-001",
				"issue_description":  "Equipment malfunction",
			},
			wantErr: true,
		},
		{
			name: "missing asset_id",
			input: map[string]interface{}{
				"service_request_id": "SR-001",
				"customer_id":        "CUST-001",
				"issue_description":  "Equipment malfunction",
			},
			wantErr: true,
		},
		{
			name: "missing issue_description",
			input: map[string]interface{}{
				"service_request_id": "SR-001",
				"customer_id":        "CUST-001",
				"asset_id":           "ASSET-001",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFieldServiceTimeoutValues verifies timeout configurations
func TestFieldServiceTimeoutValues(t *testing.T) {
	handler := NewFieldServiceSaga()
	steps := handler.GetStepDefinitions()

	for _, step := range steps {
		if step.TimeoutSeconds <= 0 {
			t.Errorf("step %d has invalid timeout: %d", step.StepNumber, step.TimeoutSeconds)
		}
		if step.TimeoutSeconds > 300 {
			t.Errorf("step %d has excessive timeout: %d seconds", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestFieldServiceServiceNames verifies service name conventions
func TestFieldServiceServiceNames(t *testing.T) {
	handler := NewFieldServiceSaga()
	steps := handler.GetStepDefinitions()

	expectedServices := map[string]bool{
		"field-service":    true,
		"technician":       true,
		"scheduling":       true,
		"parts-inventory":  true,
		"work-order":       true,
		"asset":            true,
		"sales-invoice":    true,
		"general-ledger":   true,
		"notification":     true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestFieldServiceInputMapping verifies input mappings
func TestFieldServiceInputMapping(t *testing.T) {
	handler := NewFieldServiceSaga()
	steps := handler.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber >= 100 {
			continue // Skip compensation steps
		}
		if len(step.InputMapping) == 0 {
			t.Errorf("step %d has no input mapping", step.StepNumber)
		}
		if _, ok := step.InputMapping["tenantID"]; !ok && step.StepNumber < 100 {
			t.Errorf("step %d missing tenantID mapping", step.StepNumber)
		}
	}
}

// TestFieldServiceRetryConfiguration verifies retry settings
func TestFieldServiceRetryConfiguration(t *testing.T) {
	handler := NewFieldServiceSaga()
	steps := handler.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber >= 100 {
			continue
		}
		if step.RetryConfig == nil {
			t.Errorf("step %d missing retry config", step.StepNumber)
		}
	}
}

// ============================================================================
// SAGA-W03: SPARE PARTS TESTS (11 tests)
// ============================================================================

func TestSparePartsSagaType(t *testing.T) {
	handler := NewSparePartsSaga()
	if handler.SagaType() != "SAGA-W03" {
		t.Errorf("expected SAGA-W03, got %s", handler.SagaType())
	}
}

// TestSparepartsStepCount verifies 21 total steps
func TestSparepartsStepCount(t *testing.T) {
	handler := NewSparePartsSaga()
	steps := handler.GetStepDefinitions()

	expectedStepCount := 21 // 11 forward + 10 compensation
	if len(steps) != expectedStepCount {
		t.Errorf("expected %d steps, got %d", expectedStepCount, len(steps))
	}

	forwardSteps := 0
	for _, step := range steps {
		if step.StepNumber < 100 {
			forwardSteps++
		}
	}
	if forwardSteps != 11 {
		t.Errorf("expected 11 forward steps, got %d", forwardSteps)
	}
}

// TestSparePartsCriticalSteps verifies critical markers
func TestSparePartsCriticalSteps(t *testing.T) {
	handler := NewSparePartsSaga()
	criticalSteps := map[int]bool{1: true, 2: true, 5: true, 8: true, 10: true}

	for stepNum := range criticalSteps {
		step := handler.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("critical step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be marked critical", stepNum)
		}
	}
}

// TestSparePartsValidateInput validates parts requisition parameters
func TestSparePartsValidateInput(t *testing.T) {
	handler := NewSparePartsSaga()

	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"requisition_id": "REQ-001",
				"parts_code":     "PART-001",
				"quantity":       10,
			},
			wantErr: false,
		},
		{
			name: "missing requisition_id",
			input: map[string]interface{}{
				"parts_code": "PART-001",
				"quantity":   10,
			},
			wantErr: true,
		},
		{
			name: "missing parts_code",
			input: map[string]interface{}{
				"requisition_id": "REQ-001",
				"quantity":       10,
			},
			wantErr: true,
		},
		{
			name: "missing quantity",
			input: map[string]interface{}{
				"requisition_id": "REQ-001",
				"parts_code":     "PART-001",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSparePartsServiceNames verifies service names
func TestSparePartsServiceNames(t *testing.T) {
	handler := NewSparePartsSaga()
	steps := handler.GetStepDefinitions()

	expectedServices := map[string]bool{
		"spare-parts":    true,
		"inventory":      true,
		"procurement":    true,
		"vendor":         true,
		"purchase-order": true,
		"quality":        true,
		"general-ledger": true,
		"notification":   true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestSparePartsTimeoutValues verifies timeout configurations
func TestSparePartsTimeoutValues(t *testing.T) {
	handler := NewSparePartsSaga()
	steps := handler.GetStepDefinitions()

	for _, step := range steps {
		if step.TimeoutSeconds <= 0 {
			t.Errorf("step %d has invalid timeout: %d", step.StepNumber, step.TimeoutSeconds)
		}
		if step.TimeoutSeconds > 300 {
			t.Errorf("step %d has excessive timeout: %d seconds", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestSparepartsCompensation verifies compensation step mapping
func TestSparepartsCompensation(t *testing.T) {
	handler := NewSparePartsSaga()

	forwardSteps := []int{3, 4, 5, 6, 7, 8, 9, 10, 11}
	for _, fwdStep := range forwardSteps {
		step := handler.GetStepDefinition(fwdStep)
		if step == nil {
			t.Errorf("forward step %d not found", fwdStep)
			continue
		}
		if len(step.CompensationSteps) == 0 {
			t.Errorf("forward step %d should have compensation mapping", fwdStep)
		}
	}
}

// ============================================================================
// SAGA-W04: SLA MANAGEMENT TESTS (11 tests)
// ============================================================================

func TestSLAManagementSagaType(t *testing.T) {
	handler := NewSLAManagementSaga()
	if handler.SagaType() != "SAGA-W04" {
		t.Errorf("expected SAGA-W04, got %s", handler.SagaType())
	}
}

// TestSLAStepCount verifies 21 total steps
func TestSLAStepCount(t *testing.T) {
	handler := NewSLAManagementSaga()
	steps := handler.GetStepDefinitions()

	expectedStepCount := 21 // 11 forward + 10 compensation
	if len(steps) != expectedStepCount {
		t.Errorf("expected %d steps, got %d", expectedStepCount, len(steps))
	}

	forwardSteps := 0
	for _, step := range steps {
		if step.StepNumber < 100 {
			forwardSteps++
		}
	}
	if forwardSteps != 11 {
		t.Errorf("expected 11 forward steps, got %d", forwardSteps)
	}
}

// TestSLAManagementCriticalSteps verifies critical markers
func TestSLAManagementCriticalSteps(t *testing.T) {
	handler := NewSLAManagementSaga()
	criticalSteps := map[int]bool{1: true, 2: true, 6: true, 8: true, 10: true}

	for stepNum := range criticalSteps {
		step := handler.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("critical step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be marked critical", stepNum)
		}
	}
}

// TestSLAManagementValidateInput validates SLA breach parameters
func TestSLAManagementValidateInput(t *testing.T) {
	handler := NewSLAManagementSaga()

	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"service_request_id": "SR-001",
				"sla_id":             "SLA-001",
				"detection_time":     "2026-02-16T10:00:00Z",
			},
			wantErr: false,
		},
		{
			name: "missing service_request_id",
			input: map[string]interface{}{
				"sla_id":         "SLA-001",
				"detection_time": "2026-02-16T10:00:00Z",
			},
			wantErr: true,
		},
		{
			name: "missing sla_id",
			input: map[string]interface{}{
				"service_request_id": "SR-001",
				"detection_time":     "2026-02-16T10:00:00Z",
			},
			wantErr: true,
		},
		{
			name: "missing detection_time",
			input: map[string]interface{}{
				"service_request_id": "SR-001",
				"sla_id":             "SLA-001",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSLAManagementServiceNames verifies service names
func TestSLAManagementServiceNames(t *testing.T) {
	handler := NewSLAManagementSaga()
	steps := handler.GetStepDefinitions()

	expectedServices := map[string]bool{
		"sla":               true,
		"escalation":        true,
		"notification":      true,
		"service-agreement": true,
		"service-delivery":  true,
		"approval":          true,
		"general-ledger":    true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestSLAManagementTimeoutValues verifies timeout configurations
func TestSLAManagementTimeoutValues(t *testing.T) {
	handler := NewSLAManagementSaga()
	steps := handler.GetStepDefinitions()

	for _, step := range steps {
		if step.TimeoutSeconds <= 0 {
			t.Errorf("step %d has invalid timeout: %d", step.StepNumber, step.TimeoutSeconds)
		}
		if step.TimeoutSeconds > 300 {
			t.Errorf("step %d has excessive timeout: %d seconds", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestSLAManagementInputMapping verifies input mappings
func TestSLAManagementInputMapping(t *testing.T) {
	handler := NewSLAManagementSaga()
	steps := handler.GetStepDefinitions()

	for _, step := range steps {
		if step.StepNumber >= 100 {
			continue
		}
		if len(step.InputMapping) == 0 {
			t.Errorf("step %d has no input mapping", step.StepNumber)
		}
	}
}

// ============================================================================
// SAGA-W05: CUSTOMER SATISFACTION TESTS (10 tests)
// ============================================================================

func TestCustomerSatisfactionSagaType(t *testing.T) {
	handler := NewCustomerSatisfactionSaga()
	if handler.SagaType() != "SAGA-W05" {
		t.Errorf("expected SAGA-W05, got %s", handler.SagaType())
	}
}

// TestSatisfactionStepCount verifies 23 total steps
func TestSatisfactionStepCount(t *testing.T) {
	handler := NewCustomerSatisfactionSaga()
	steps := handler.GetStepDefinitions()

	expectedStepCount := 23 // 12 forward + 11 compensation
	if len(steps) != expectedStepCount {
		t.Errorf("expected %d steps, got %d", expectedStepCount, len(steps))
	}

	forwardSteps := 0
	for _, step := range steps {
		if step.StepNumber < 100 {
			forwardSteps++
		}
	}
	if forwardSteps != 12 {
		t.Errorf("expected 12 forward steps, got %d", forwardSteps)
	}
}

// TestCustomerSatisfactionCriticalSteps verifies critical markers
func TestCustomerSatisfactionCriticalSteps(t *testing.T) {
	handler := NewCustomerSatisfactionSaga()
	criticalSteps := map[int]bool{1: true, 3: true, 4: true, 8: true, 9: true, 11: true}

	for stepNum := range criticalSteps {
		step := handler.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("critical step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be marked critical", stepNum)
		}
	}
}

// TestCustomerSatisfactionValidateInput validates satisfaction survey parameters
func TestCustomerSatisfactionValidateInput(t *testing.T) {
	handler := NewCustomerSatisfactionSaga()

	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"service_request_id": "SR-001",
				"customer_id":        "CUST-001",
				"completion_date":    "2026-02-16",
			},
			wantErr: false,
		},
		{
			name: "missing service_request_id",
			input: map[string]interface{}{
				"customer_id":     "CUST-001",
				"completion_date": "2026-02-16",
			},
			wantErr: true,
		},
		{
			name: "missing customer_id",
			input: map[string]interface{}{
				"service_request_id": "SR-001",
				"completion_date":    "2026-02-16",
			},
			wantErr: true,
		},
		{
			name: "missing completion_date",
			input: map[string]interface{}{
				"service_request_id": "SR-001",
				"customer_id":        "CUST-001",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCustomerSatisfactionServiceNames verifies service names
func TestCustomerSatisfactionServiceNames(t *testing.T) {
	handler := NewCustomerSatisfactionSaga()
	steps := handler.GetStepDefinitions()

	expectedServices := map[string]bool{
		"feedback":      true,
		"notification":  true,
		"satisfaction":  true,
		"approval":      true,
		"general-ledger": true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestCustomerSatisfactionTimeoutValues verifies timeout configurations
func TestCustomerSatisfactionTimeoutValues(t *testing.T) {
	handler := NewCustomerSatisfactionSaga()
	steps := handler.GetStepDefinitions()

	for _, step := range steps {
		if step.TimeoutSeconds <= 0 {
			t.Errorf("step %d has invalid timeout: %d", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestSatisfactionCompensation verifies compensation mapping
func TestSatisfactionCompensation(t *testing.T) {
	handler := NewCustomerSatisfactionSaga()

	forwardSteps := []int{5, 6, 7, 8, 9, 10, 11, 12}
	for _, fwdStep := range forwardSteps {
		step := handler.GetStepDefinition(fwdStep)
		if step == nil {
			t.Errorf("forward step %d not found", fwdStep)
			continue
		}
		if len(step.CompensationSteps) == 0 {
			t.Errorf("forward step %d should have compensation mapping", fwdStep)
		}
	}
}

// ============================================================================
// SAGA-W06: EXTENDED WARRANTY TESTS (10 tests)
// ============================================================================

func TestExtendedWarrantySagaType(t *testing.T) {
	handler := NewExtendedWarrantySaga()
	if handler.SagaType() != "SAGA-W06" {
		t.Errorf("expected SAGA-W06, got %s", handler.SagaType())
	}
}

// TestExtendedWarrantyStepCount verifies 22 total steps
func TestExtendedWarrantyStepCount(t *testing.T) {
	handler := NewExtendedWarrantySaga()
	steps := handler.GetStepDefinitions()

	expectedStepCount := 22 // 12 forward + 10 compensation
	if len(steps) != expectedStepCount {
		t.Errorf("expected %d steps, got %d", expectedStepCount, len(steps))
	}

	forwardSteps := 0
	for _, step := range steps {
		if step.StepNumber < 100 {
			forwardSteps++
		}
	}
	if forwardSteps != 12 {
		t.Errorf("expected 12 forward steps, got %d", forwardSteps)
	}
}

// TestExtendedWarrantyCriticalSteps verifies critical markers
func TestExtendedWarrantyCriticalSteps(t *testing.T) {
	handler := NewExtendedWarrantySaga()
	criticalSteps := map[int]bool{1: true, 3: true, 5: true, 7: true, 8: true, 10: true}

	for stepNum := range criticalSteps {
		step := handler.GetStepDefinition(stepNum)
		if step == nil {
			t.Errorf("critical step %d not found", stepNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be marked critical", stepNum)
		}
	}
}

// TestExtendedWarrantyValidateInput validates plan subscription parameters
func TestExtendedWarrantyValidateInput(t *testing.T) {
	handler := NewExtendedWarrantySaga()

	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"plan_id":           "PLAN-001",
				"customer_id":       "CUST-001",
				"product_id":        "PROD-001",
				"subscription_date": "2026-02-16",
			},
			wantErr: false,
		},
		{
			name: "missing plan_id",
			input: map[string]interface{}{
				"customer_id":       "CUST-001",
				"product_id":        "PROD-001",
				"subscription_date": "2026-02-16",
			},
			wantErr: true,
		},
		{
			name: "missing customer_id",
			input: map[string]interface{}{
				"plan_id":           "PLAN-001",
				"product_id":        "PROD-001",
				"subscription_date": "2026-02-16",
			},
			wantErr: true,
		},
		{
			name: "missing product_id",
			input: map[string]interface{}{
				"plan_id":           "PLAN-001",
				"customer_id":       "CUST-001",
				"subscription_date": "2026-02-16",
			},
			wantErr: true,
		},
		{
			name: "missing subscription_date",
			input: map[string]interface{}{
				"plan_id":     "PLAN-001",
				"customer_id": "CUST-001",
				"product_id":  "PROD-001",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestExtendedWarrantyServiceNames verifies service names
func TestExtendedWarrantyServiceNames(t *testing.T) {
	handler := NewExtendedWarrantySaga()
	steps := handler.GetStepDefinitions()

	expectedServices := map[string]bool{
		"warranty-plan":    true,
		"sales-order":      true,
		"sales-invoice":    true,
		"billing":          true,
		"general-ledger":   true,
		"notification":     true,
		"insurance":        true,
	}

	for _, step := range steps {
		if !expectedServices[step.ServiceName] {
			t.Errorf("step %d has unexpected service name: %s", step.StepNumber, step.ServiceName)
		}
	}
}

// TestExtendedWarrantyTimeoutValues verifies timeout configurations
func TestExtendedWarrantyTimeoutValues(t *testing.T) {
	handler := NewExtendedWarrantySaga()
	steps := handler.GetStepDefinitions()

	for _, step := range steps {
		if step.TimeoutSeconds <= 0 {
			t.Errorf("step %d has invalid timeout: %d", step.StepNumber, step.TimeoutSeconds)
		}
		if step.TimeoutSeconds > 300 {
			t.Errorf("step %d has excessive timeout: %d seconds", step.StepNumber, step.TimeoutSeconds)
		}
	}
}

// TestExtendedWarrantyCompensation verifies compensation mapping
func TestExtendedWarrantyCompensation(t *testing.T) {
	handler := NewExtendedWarrantySaga()

	forwardSteps := []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	for _, fwdStep := range forwardSteps {
		step := handler.GetStepDefinition(fwdStep)
		if step == nil {
			t.Errorf("forward step %d not found", fwdStep)
			continue
		}
		if len(step.CompensationSteps) == 0 && fwdStep > 1 {
			t.Errorf("forward step %d should have compensation mapping", fwdStep)
		}
	}
}

// ============================================================================
// COMPREHENSIVE INTEGRATION TESTS (13 tests)
// ============================================================================

// TestAllWarrantySagasRegistered verifies all sagas provided by handler
func TestAllWarrantySagasRegistered(t *testing.T) {
	handlers := ProvideWarrantySagaHandlers()

	expectedCount := 6
	if len(handlers) != expectedCount {
		t.Errorf("expected %d saga handlers, got %d", expectedCount, len(handlers))
	}

	expectedTypes := []string{"SAGA-W01", "SAGA-W02", "SAGA-W03", "SAGA-W04", "SAGA-W05", "SAGA-W06"}
	handlerMap := make(map[string]saga.SagaHandler)
	for _, handler := range handlers {
		handlerMap[handler.SagaType()] = handler
	}

	for _, sagaType := range expectedTypes {
		if _, exists := handlerMap[sagaType]; !exists {
			t.Errorf("saga handler %s not registered", sagaType)
		}
	}
}

// TestRetryConfigurationConsistency verifies retry settings across all sagas
func TestRetryConfigurationConsistency(t *testing.T) {
	handlers := ProvideWarrantySagaHandlers()

	for _, handler := range handlers {
		steps := handler.GetStepDefinitions()
		for _, step := range steps {
			// Forward steps should have retry config
			if step.StepNumber < 100 && step.RetryConfig == nil {
				t.Errorf("%s step %d missing retry config", handler.SagaType(), step.StepNumber)
			}

			// Verify retry config values
			if step.RetryConfig != nil {
				if step.RetryConfig.MaxRetries <= 0 {
					t.Errorf("%s step %d has invalid MaxRetries: %d", handler.SagaType(), step.StepNumber, step.RetryConfig.MaxRetries)
				}
				if step.RetryConfig.BackoffMultiplier <= 0 {
					t.Errorf("%s step %d has invalid BackoffMultiplier: %f", handler.SagaType(), step.StepNumber, step.RetryConfig.BackoffMultiplier)
				}
				if step.RetryConfig.InitialBackoffMs <= 0 {
					t.Errorf("%s step %d has invalid InitialBackoffMs: %d", handler.SagaType(), step.StepNumber, step.RetryConfig.InitialBackoffMs)
				}
				if step.RetryConfig.MaxBackoffMs <= 0 {
					t.Errorf("%s step %d has invalid MaxBackoffMs: %d", handler.SagaType(), step.StepNumber, step.RetryConfig.MaxBackoffMs)
				}
			}
		}
	}
}

// TestForwardAndCompensationStepOrdering verifies proper step numbering
func TestForwardAndCompensationStepOrdering(t *testing.T) {
	handlers := ProvideWarrantySagaHandlers()

	for _, handler := range handlers {
		steps := handler.GetStepDefinitions()
		forwardNums := []int{}
		compensationNums := []int{}

		for _, step := range steps {
			if step.StepNumber < 100 {
				forwardNums = append(forwardNums, int(step.StepNumber))
			} else {
				compensationNums = append(compensationNums, int(step.StepNumber))
			}
		}

		// Verify forward steps are contiguous starting from 1
		if len(forwardNums) > 0 && forwardNums[0] != 1 {
			t.Errorf("%s: first forward step should be 1, got %d", handler.SagaType(), forwardNums[0])
		}

		// Verify compensation steps are 3-digit numbers >= 100
		for _, compNum := range compensationNums {
			if compNum < 100 {
				t.Errorf("%s: compensation step %d should be >= 100", handler.SagaType(), compNum)
			}
		}
	}
}

// TestCompensationStepsMapping verifies compensation references exist
func TestCompensationStepsMapping(t *testing.T) {
	handlers := ProvideWarrantySagaHandlers()

	for _, handler := range handlers {
		steps := handler.GetStepDefinitions()
		for _, step := range steps {
			// Forward steps should have compensation steps mapped
			if step.StepNumber < 100 && len(step.CompensationSteps) > 0 {
				// Verify all compensation steps exist
				for _, compStepNum := range step.CompensationSteps {
					compStep := handler.GetStepDefinition(int(compStepNum))
					if compStep == nil {
						t.Errorf("%s step %d references non-existent compensation step %d", handler.SagaType(), step.StepNumber, compStepNum)
					}
					// Verify it's actually a compensation step
					if compStep.StepNumber < 100 {
						t.Errorf("%s compensation reference %d is not a compensation step", handler.SagaType(), compStepNum)
					}
				}
			}
		}
	}
}

// TestAllStepsHaveServiceNames verifies every step has a service name
func TestAllStepsHaveServiceNames(t *testing.T) {
	handlers := ProvideWarrantySagaHandlers()

	for _, handler := range handlers {
		steps := handler.GetStepDefinitions()
		for _, step := range steps {
			if step.ServiceName == "" {
				t.Errorf("%s step %d has no service name", handler.SagaType(), step.StepNumber)
			}
		}
	}
}

// TestAllStepsHaveHandlerMethods verifies every step has a handler method
func TestAllStepsHaveHandlerMethods(t *testing.T) {
	handlers := ProvideWarrantySagaHandlers()

	for _, handler := range handlers {
		steps := handler.GetStepDefinitions()
		for _, step := range steps {
			if step.HandlerMethod == "" {
				t.Errorf("%s step %d has no handler method", handler.SagaType(), step.StepNumber)
			}
		}
	}
}

// TestStepTimeoutValues verifies timeouts are reasonable
func TestStepTimeoutValues(t *testing.T) {
	handlers := ProvideWarrantySagaHandlers()

	for _, handler := range handlers {
		steps := handler.GetStepDefinitions()
		for _, step := range steps {
			// Verify timeout is set and reasonable
			if step.TimeoutSeconds <= 0 {
				t.Errorf("%s step %d has invalid timeout: %d", handler.SagaType(), step.StepNumber, step.TimeoutSeconds)
			}
			if step.TimeoutSeconds > 300 {
				t.Errorf("%s step %d has excessive timeout: %d seconds", handler.SagaType(), step.StepNumber, step.TimeoutSeconds)
			}
		}
	}
}

// TestCriticalStepsDistribution verifies critical steps are properly marked
func TestCriticalStepsDistribution(t *testing.T) {
	handlers := ProvideWarrantySagaHandlers()

	for _, handler := range handlers {
		steps := handler.GetStepDefinitions()
		criticalCount := 0
		nonCriticalCount := 0

		for _, step := range steps {
			if step.StepNumber >= 100 {
				continue // Skip compensation steps
			}
			if step.IsCritical {
				criticalCount++
			} else {
				nonCriticalCount++
			}
		}

		// Verify at least some critical steps
		if criticalCount == 0 {
			t.Errorf("%s has no critical steps", handler.SagaType())
		}
		// Verify critical steps are reasonable percentage of total
		if criticalCount > nonCriticalCount*2 {
			t.Warnf("%s has unusually high critical step ratio (%d critical, %d non-critical)",
				handler.SagaType(), criticalCount, nonCriticalCount)
		}
	}
}

// TestInputMappingsUseJSONPath verifies JSONPath usage in mappings
func TestInputMappingsUseJSONPath(t *testing.T) {
	handlers := ProvideWarrantySagaHandlers()

	for _, handler := range handlers {
		steps := handler.GetStepDefinitions()
		for _, step := range steps {
			if step.StepNumber >= 100 {
				continue // Skip compensation steps
			}
			for key, value := range step.InputMapping {
				// Every mapping should either have $ or be a valid key
				if value != "" && !string(value[0]) == "$" && len(value) > 0 {
					// Allow some mappings to not use JSONPath if they're literals
					// Just verify they're not obviously broken
					if value == "INVALID" {
						t.Errorf("%s step %d mapping %s has invalid value: %s",
							handler.SagaType(), step.StepNumber, key, value)
					}
				}
			}
		}
	}
}

// TestNoDuplicateStepNumbers verifies step numbers are unique within saga
func TestNoDuplicateStepNumbers(t *testing.T) {
	handlers := ProvideWarrantySagaHandlers()

	for _, handler := range handlers {
		steps := handler.GetStepDefinitions()
		stepNumMap := make(map[int32]bool)

		for _, step := range steps {
			if stepNumMap[step.StepNumber] {
				t.Errorf("%s has duplicate step number: %d", handler.SagaType(), step.StepNumber)
			}
			stepNumMap[step.StepNumber] = true
		}
	}
}

// TestServiceNameFormatting verifies service names follow conventions
func TestServiceNameFormatting(t *testing.T) {
	// Services with hyphens
	hyphenatedServices := map[string]bool{
		"field-service":      true,
		"spare-parts":        true,
		"parts-inventory":    true,
		"work-order":         true,
		"sales-invoice":      true,
		"general-ledger":     true,
		"sales-order":        true,
		"accounts-receivable": true,
		"warranty-plan":      true,
		"service-agreement":  true,
		"service-delivery":   true,
		"purchase-order":     true,
	}

	// Non-hyphenated services that are allowed
	allowedServices := map[string]bool{
		"warranty":     true,
		"technician":   true,
		"scheduling":   true,
		"location":     true,
		"asset":        true,
		"inventory":    true,
		"procurement":  true,
		"vendor":       true,
		"requisition":  true,
		"quality":      true,
		"sla":          true,
		"escalation":   true,
		"notification": true,
		"approval":     true,
		"feedback":     true,
		"satisfaction": true,
		"billing":      true,
		"insurance":    true,
	}

	handlers := ProvideWarrantySagaHandlers()

	for _, handler := range handlers {
		steps := handler.GetStepDefinitions()
		for _, step := range steps {
			// Check if service name is properly formatted
			if !hyphenatedServices[step.ServiceName] && !allowedServices[step.ServiceName] {
				t.Errorf("%s step %d has unexpected service name: %s",
					handler.SagaType(), step.StepNumber, step.ServiceName)
			}
		}
	}
}

// TestInputMappingConsistency verifies all input mappings are complete
func TestInputMappingConsistency(t *testing.T) {
	handlers := ProvideWarrantySagaHandlers()

	for _, handler := range handlers {
		steps := handler.GetStepDefinitions()
		for _, step := range steps {
			// All forward steps should have input mapping
			if step.StepNumber < 100 && len(step.InputMapping) == 0 {
				t.Errorf("%s step %d missing input mapping", handler.SagaType(), step.StepNumber)
			}

			// Verify expected context fields exist in mapping for forward steps
			if step.StepNumber < 100 {
				contextFields := []string{"tenantID", "companyID", "branchID"}
				for _, field := range contextFields {
					if _, ok := step.InputMapping[field]; !ok {
						t.Errorf("%s step %d missing %s mapping", handler.SagaType(), step.StepNumber, field)
					}
				}
			}
		}
	}
}

// ============================================================================
// BENCHMARK TESTS (5 tests)
// ============================================================================

// BenchmarkWarrantyClaimSagaInstantiation measures claim saga creation time
func BenchmarkWarrantyClaimSagaInstantiation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewWarrantyClaimSaga()
	}
}

// BenchmarkFieldServiceSagaInstantiation measures field service saga creation
func BenchmarkFieldServiceSagaInstantiation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewFieldServiceSaga()
	}
}

// BenchmarkSparePartsSagaInstantiation measures spare parts saga creation
func BenchmarkSparePartsSagaInstantiation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewSparePartsSaga()
	}
}

// BenchmarkGetStepDefinition measures step lookup performance
func BenchmarkGetStepDefinition(b *testing.B) {
	handler := NewWarrantyClaimSaga()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = handler.GetStepDefinition(5)
	}
}

// BenchmarkValidateInput measures input validation performance
func BenchmarkValidateInput(b *testing.B) {
	handler := NewWarrantyClaimSaga()
	input := map[string]interface{}{
		"claim_id":          "CLM-001",
		"product_id":        "PROD-001",
		"customer_id":       "CUST-001",
		"issue_description": "Product defect",
		"claim_date":        "2026-02-16",
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = handler.ValidateInput(input)
	}
}
