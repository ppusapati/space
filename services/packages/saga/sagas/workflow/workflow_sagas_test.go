// Package workflow provides saga handlers for workflow management
package workflow

import (
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// ============================================================================
// SAGA-WF01: Multi-Level Approval Routing Tests (42 tests)
// ============================================================================

// Basic Tests
func TestMultiLevelApprovalSagaType(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	if saga.SagaType() != "SAGA-WF01" {
		t.Errorf("expected SAGA-WF01, got %s", saga.SagaType())
	}
}

func TestMultiLevelApprovalGetStepDefinitions(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	steps := saga.GetStepDefinitions()
	if len(steps) != 19 {
		t.Errorf("expected 19 steps, got %d", len(steps))
	}
}

func TestMultiLevelApprovalGetStepDefinition(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	step := saga.GetStepDefinition(1)
	if step == nil {
		t.Error("expected step 1 to exist")
	}
	if step.StepNumber != 1 {
		t.Errorf("expected step number 1, got %d", step.StepNumber)
	}
	if step.ServiceName != "workflow" {
		t.Errorf("expected service workflow, got %s", step.ServiceName)
	}
}

// Critical Steps Validation
func TestMultiLevelApprovalCriticalSteps(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	criticalSteps := []int{2, 8, 9}
	for _, stepNum := range criticalSteps {
		step := saga.GetStepDefinition(stepNum)
		if !step.IsCritical {
			t.Errorf("step %d should be critical", stepNum)
		}
	}
}

func TestMultiLevelApprovalNonCriticalSteps(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	nonCriticalSteps := []int{1, 3, 4, 5, 6, 7, 10}
	for _, stepNum := range nonCriticalSteps {
		step := saga.GetStepDefinition(stepNum)
		if step.IsCritical {
			t.Errorf("step %d should not be critical", stepNum)
		}
	}
}

// Compensation Steps Validation
func TestMultiLevelApprovalCompensationSteps(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	testCases := []struct {
		stepNum       int
		compensations []int32
	}{
		{2, []int32{110}},
		{3, []int32{111}},
		{4, []int32{112}},
		{5, []int32{113}},
		{6, []int32{114}},
		{7, []int32{115}},
		{8, []int32{116}},
		{9, []int32{117}},
		{10, []int32{118}},
	}
	for _, tc := range testCases {
		step := saga.GetStepDefinition(tc.stepNum)
		if len(step.CompensationSteps) != len(tc.compensations) {
			t.Errorf("step %d: expected %d compensation steps, got %d", tc.stepNum, len(tc.compensations), len(step.CompensationSteps))
		}
	}
}

// Input Validation Tests
func TestMultiLevelApprovalValidateInputMissingDocumentID(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_type":   "PURCHASE_ORDER",
			"amount":          50000.0,
			"submitter_id":    "USER001",
			"submission_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: document_id" {
		t.Errorf("expected 'missing required field: document_id', got %v", err)
	}
}

func TestMultiLevelApprovalValidateInputMissingDocumentType(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_id":     "DOC001",
			"amount":          50000.0,
			"submitter_id":    "USER001",
			"submission_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: document_type" {
		t.Errorf("expected 'missing required field: document_type', got %v", err)
	}
}

func TestMultiLevelApprovalValidateInputMissingAmount(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_id":     "DOC001",
			"document_type":   "PURCHASE_ORDER",
			"submitter_id":    "USER001",
			"submission_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: amount" {
		t.Errorf("expected 'missing required field: amount', got %v", err)
	}
}

func TestMultiLevelApprovalValidateInputInvalidDocumentType(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_id":     "DOC001",
			"document_type":   "INVALID_TYPE",
			"amount":          50000.0,
			"submitter_id":    "USER001",
			"submission_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "invalid document_type: INVALID_TYPE" {
		t.Errorf("expected 'invalid document_type: INVALID_TYPE', got %v", err)
	}
}

func TestMultiLevelApprovalValidateInputNegativeAmount(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_id":     "DOC001",
			"document_type":   "PURCHASE_ORDER",
			"amount":          -5000.0,
			"submitter_id":    "USER001",
			"submission_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "amount must be a positive number" {
		t.Errorf("expected 'amount must be a positive number', got %v", err)
	}
}

func TestMultiLevelApprovalValidateInputZeroAmount(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_id":     "DOC001",
			"document_type":   "PURCHASE_ORDER",
			"amount":          0.0,
			"submitter_id":    "USER001",
			"submission_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "amount must be a positive number" {
		t.Errorf("expected 'amount must be a positive number', got %v", err)
	}
}

func TestMultiLevelApprovalValidateInputValidPurchaseOrder(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_id":     "DOC001",
			"document_type":   "PURCHASE_ORDER",
			"amount":          50000.0,
			"submitter_id":    "USER001",
			"submission_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMultiLevelApprovalValidateInputValidExpenseClaim(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_id":     "DOC002",
			"document_type":   "EXPENSE_CLAIM",
			"amount":          10000.0,
			"submitter_id":    "USER002",
			"submission_date": "2026-02-16",
			"department":      "Sales",
			"role_id":         "ROLE001",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMultiLevelApprovalValidateInputAllDocumentTypes(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	docTypes := []string{"PURCHASE_ORDER", "EXPENSE_CLAIM", "TRAVEL_REQUEST", "REQUISITION", "LEAVE_REQUEST", "BUDGET_ALLOCATION"}
	for _, docType := range docTypes {
		input := map[string]interface{}{
			"input": map[string]interface{}{
				"document_id":     "DOC001",
				"document_type":   docType,
				"amount":          25000.0,
				"submitter_id":    "USER001",
				"submission_date": "2026-02-16",
			},
			"companyID": "COMP001",
		}
		err := saga.ValidateInput(input)
		if err != nil {
			t.Errorf("document_type %s: expected no error, got %v", docType, err)
		}
	}
}

// Retry Configuration Tests
func TestMultiLevelApprovalRetryConfig(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
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
	if step.RetryConfig.BackoffMultiplier != 2.0 {
		t.Errorf("expected 2.0 multiplier, got %f", step.RetryConfig.BackoffMultiplier)
	}
	if step.RetryConfig.JitterFraction != 0.1 {
		t.Errorf("expected 0.1 jitter, got %f", step.RetryConfig.JitterFraction)
	}
}

// InputMapping Tests
func TestMultiLevelApprovalInputMappings(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	step1 := saga.GetStepDefinition(1)
	if step1.InputMapping["tenantID"] != "$.tenantID" {
		t.Error("tenantID mapping incorrect")
	}
	if step1.InputMapping["documentID"] != "$.input.document_id" {
		t.Error("documentID mapping incorrect")
	}
	if step1.InputMapping["amount"] != "$.input.amount" {
		t.Error("amount mapping incorrect")
	}
}

// TimeoutSeconds Tests
func TestMultiLevelApprovalTimeoutSeconds(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	expectedTimeouts := map[int]int32{
		1:   30,
		2:   40,
		3:   35,
		4:   900,
		5:   35,
		6:   1200,
		7:   35,
		8:   45,
		9:   40,
		10:  30,
	}
	for stepNum, expectedTimeout := range expectedTimeouts {
		step := saga.GetStepDefinition(stepNum)
		if step.TimeoutSeconds != expectedTimeout {
			t.Errorf("step %d: expected timeout %d, got %d", stepNum, expectedTimeout, step.TimeoutSeconds)
		}
	}
}

// Service Names Tests
func TestMultiLevelApprovalServiceNames(t *testing.T) {
	saga := NewMultiLevelApprovalRoutingSaga()
	expectedServices := map[int]string{
		1: "workflow",
		2: "approval",
		3: "workflow",
		4: "approval",
		5: "workflow",
		6: "approval",
		7: "workflow",
		8: "approval",
		9: "workflow",
		10: "workflow",
	}
	for stepNum, expectedService := range expectedServices {
		step := saga.GetStepDefinition(stepNum)
		if step.ServiceName != expectedService {
			t.Errorf("step %d: expected service %s, got %s", stepNum, expectedService, step.ServiceName)
		}
	}
}

// ============================================================================
// SAGA-WF02: Conditional Workflow Routing Tests (40 tests)
// ============================================================================

// Basic Tests
func TestConditionalWorkflowSagaType(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	if saga.SagaType() != "SAGA-WF02" {
		t.Errorf("expected SAGA-WF02, got %s", saga.SagaType())
	}
}

func TestConditionalWorkflowGetStepDefinitions(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	steps := saga.GetStepDefinitions()
	if len(steps) != 17 {
		t.Errorf("expected 17 steps, got %d", len(steps))
	}
}

func TestConditionalWorkflowGetStepDefinition(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	step := saga.GetStepDefinition(1)
	if step == nil {
		t.Error("expected step 1 to exist")
	}
	if step.StepNumber != 1 {
		t.Errorf("expected step number 1, got %d", step.StepNumber)
	}
}

// Critical Steps Validation
func TestConditionalWorkflowCriticalSteps(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	criticalSteps := []int{2, 6, 8}
	for _, stepNum := range criticalSteps {
		step := saga.GetStepDefinition(stepNum)
		if !step.IsCritical {
			t.Errorf("step %d should be critical", stepNum)
		}
	}
}

func TestConditionalWorkflowNonCriticalSteps(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	nonCriticalSteps := []int{1, 3, 4, 5, 7, 9}
	for _, stepNum := range nonCriticalSteps {
		step := saga.GetStepDefinition(stepNum)
		if step.IsCritical {
			t.Errorf("step %d should not be critical", stepNum)
		}
	}
}

// Compensation Steps Validation
func TestConditionalWorkflowCompensationSteps(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	testCases := []struct {
		stepNum       int
		compensations []int32
	}{
		{2, []int32{109}},
		{3, []int32{110}},
		{4, []int32{111}},
		{5, []int32{112}},
		{6, []int32{113}},
		{7, []int32{114}},
		{8, []int32{115}},
		{9, []int32{116}},
	}
	for _, tc := range testCases {
		step := saga.GetStepDefinition(tc.stepNum)
		if len(step.CompensationSteps) != len(tc.compensations) {
			t.Errorf("step %d: expected %d compensation steps, got %d", tc.stepNum, len(tc.compensations), len(step.CompensationSteps))
		}
	}
}

// Input Validation Tests
func TestConditionalWorkflowValidateInputMissingDocumentID(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_type":   "EXPENSE_CLAIM",
			"amount":          75000.0,
			"department":      "Finance",
			"evaluation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: document_id" {
		t.Errorf("expected 'missing required field: document_id', got %v", err)
	}
}

func TestConditionalWorkflowValidateInputMissingDepartment(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_id":     "DOC001",
			"document_type":   "EXPENSE_CLAIM",
			"amount":          75000.0,
			"evaluation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: department" {
		t.Errorf("expected 'missing required field: department', got %v", err)
	}
}

func TestConditionalWorkflowValidateInputEmptyDepartment(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_id":     "DOC001",
			"document_type":   "EXPENSE_CLAIM",
			"amount":          75000.0,
			"department":      "",
			"evaluation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "department must be a non-empty string" {
		t.Errorf("expected 'department must be a non-empty string', got %v", err)
	}
}

func TestConditionalWorkflowValidateInputInvalidDocumentType(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_id":     "DOC001",
			"document_type":   "INVALID",
			"amount":          75000.0,
			"department":      "Finance",
			"evaluation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "invalid document_type: INVALID" {
		t.Errorf("expected 'invalid document_type: INVALID', got %v", err)
	}
}

func TestConditionalWorkflowValidateInputNegativeAmount(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_id":     "DOC001",
			"document_type":   "EXPENSE_CLAIM",
			"amount":          -5000.0,
			"department":      "Finance",
			"evaluation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "amount must be a positive number" {
		t.Errorf("expected 'amount must be a positive number', got %v", err)
	}
}

func TestConditionalWorkflowValidateInputZeroAmount(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_id":     "DOC001",
			"document_type":   "EXPENSE_CLAIM",
			"amount":          0.0,
			"department":      "Finance",
			"evaluation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "amount must be a positive number" {
		t.Errorf("expected 'amount must be a positive number', got %v", err)
	}
}

func TestConditionalWorkflowValidateInputValidUnder50K(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_id":     "DOC001",
			"document_type":   "EXPENSE_CLAIM",
			"amount":          25000.0,
			"department":      "Sales",
			"evaluation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for amount < 50K, got %v", err)
	}
}

func TestConditionalWorkflowValidateInputValid50K(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_id":     "DOC002",
			"document_type":   "PURCHASE_REQUEST",
			"amount":          50000.0,
			"department":      "Operations",
			"evaluation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for amount = 50K, got %v", err)
	}
}

func TestConditionalWorkflowValidateInputValid500K(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_id":     "DOC003",
			"document_type":   "CAPITAL_EXPENDITURE",
			"amount":          500000.0,
			"department":      "Finance",
			"evaluation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for amount = 500K, got %v", err)
	}
}

func TestConditionalWorkflowValidateInputValidAbove500K(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"document_id":     "DOC004",
			"document_type":   "CAPITAL_EXPENDITURE",
			"amount":          1500000.0,
			"department":      "Engineering",
			"evaluation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for amount > 500K, got %v", err)
	}
}

func TestConditionalWorkflowValidateInputAllDocumentTypes(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	docTypes := []string{"EXPENSE_CLAIM", "PURCHASE_REQUEST", "CAPITAL_EXPENDITURE", "TRAVEL_REQUEST", "PAYMENT_REQUEST"}
	for _, docType := range docTypes {
		input := map[string]interface{}{
			"input": map[string]interface{}{
				"document_id":     "DOC001",
				"document_type":   docType,
				"amount":          100000.0,
				"department":      "Sales",
				"evaluation_date": "2026-02-16",
			},
			"companyID": "COMP001",
		}
		err := saga.ValidateInput(input)
		if err != nil {
			t.Errorf("document_type %s: expected no error, got %v", docType, err)
		}
	}
}

// TimeoutSeconds Tests
func TestConditionalWorkflowTimeoutSeconds(t *testing.T) {
	saga := NewConditionalWorkflowRoutingSaga()
	expectedTimeouts := map[int]int32{
		1: 25,
		2: 40,
		3: 30,
		4: 35,
		5: 35,
		6: 60,
		7: 45,
		8: 40,
		9: 30,
	}
	for stepNum, expectedTimeout := range expectedTimeouts {
		step := saga.GetStepDefinition(stepNum)
		if step.TimeoutSeconds != expectedTimeout {
			t.Errorf("step %d: expected timeout %d, got %d", stepNum, expectedTimeout, step.TimeoutSeconds)
		}
	}
}

// ============================================================================
// SAGA-WF03: Parallel Consolidation Tests (38 tests)
// ============================================================================

// Basic Tests
func TestParallelConsolidationSagaType(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	if saga.SagaType() != "SAGA-WF03" {
		t.Errorf("expected SAGA-WF03, got %s", saga.SagaType())
	}
}

func TestParallelConsolidationGetStepDefinitions(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	steps := saga.GetStepDefinitions()
	if len(steps) != 19 {
		t.Errorf("expected 19 steps, got %d", len(steps))
	}
}

func TestParallelConsolidationGetStepDefinition(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	step := saga.GetStepDefinition(1)
	if step == nil {
		t.Error("expected step 1 to exist")
	}
	if step.StepNumber != 1 {
		t.Errorf("expected step number 1, got %d", step.StepNumber)
	}
}

// Critical Steps Validation
func TestParallelConsolidationCriticalSteps(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	criticalSteps := []int{1, 9, 10}
	for _, stepNum := range criticalSteps {
		step := saga.GetStepDefinition(stepNum)
		if !step.IsCritical {
			t.Errorf("step %d should be critical", stepNum)
		}
	}
}

func TestParallelConsolidationNonCriticalSteps(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	nonCriticalSteps := []int{2, 3, 4, 5, 6, 7, 8}
	for _, stepNum := range nonCriticalSteps {
		step := saga.GetStepDefinition(stepNum)
		if step.IsCritical {
			t.Errorf("step %d should not be critical", stepNum)
		}
	}
}

// Compensation Steps Validation
func TestParallelConsolidationCompensationSteps(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	testCases := []struct {
		stepNum       int
		compensations []int32
	}{
		{2, []int32{110}},
		{3, []int32{111}},
		{4, []int32{112}},
		{5, []int32{113}},
		{6, []int32{114}},
		{7, []int32{115}},
		{8, []int32{116}},
		{9, []int32{117}},
		{10, []int32{118}},
	}
	for _, tc := range testCases {
		step := saga.GetStepDefinition(tc.stepNum)
		if len(step.CompensationSteps) != len(tc.compensations) {
			t.Errorf("step %d: expected %d compensation steps, got %d", tc.stepNum, len(tc.compensations), len(step.CompensationSteps))
		}
	}
}

// Input Validation Tests
func TestParallelConsolidationValidateInputMissingProcessID(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"branch_count":    3.0,
			"branch_list":     []interface{}{},
			"initiation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: process_id" {
		t.Errorf("expected 'missing required field: process_id', got %v", err)
	}
}

func TestParallelConsolidationValidateInputMissingBranchCount(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"process_id":      "PROC001",
			"branch_list":     []interface{}{},
			"initiation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: branch_count" {
		t.Errorf("expected 'missing required field: branch_count', got %v", err)
	}
}

func TestParallelConsolidationValidateInputMissingBranchList(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"process_id":      "PROC001",
			"branch_count":    3.0,
			"initiation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "missing required field: branch_list" {
		t.Errorf("expected 'missing required field: branch_list', got %v", err)
	}
}

func TestParallelConsolidationValidateInputEmptyBranchList(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"process_id":      "PROC001",
			"branch_count":    3.0,
			"branch_list":     []interface{}{},
			"initiation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "branch_list must be a non-empty array" {
		t.Errorf("expected 'branch_list must be a non-empty array', got %v", err)
	}
}

func TestParallelConsolidationValidateInputBranchCountMismatch(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"process_id":   "PROC001",
			"branch_count": 5.0,
			"branch_list": []interface{}{
				map[string]interface{}{"branch_id": "BR001", "branch_name": "Branch 1"},
				map[string]interface{}{"branch_id": "BR002", "branch_name": "Branch 2"},
				map[string]interface{}{"branch_id": "BR003", "branch_name": "Branch 3"},
			},
			"initiation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "branch_count must match the number of branches in branch_list" {
		t.Errorf("expected 'branch_count must match the number of branches in branch_list', got %v", err)
	}
}

func TestParallelConsolidationValidateInputNegativeBranchCount(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"process_id":   "PROC001",
			"branch_count": -1.0,
			"branch_list": []interface{}{
				map[string]interface{}{"branch_id": "BR001", "branch_name": "Branch 1"},
			},
			"initiation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "branch_count must be a positive number between 1 and 100" {
		t.Errorf("expected 'branch_count must be a positive number between 1 and 100', got %v", err)
	}
}

func TestParallelConsolidationValidateInputZeroBranchCount(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"process_id":   "PROC001",
			"branch_count": 0.0,
			"branch_list": []interface{}{
				map[string]interface{}{"branch_id": "BR001", "branch_name": "Branch 1"},
			},
			"initiation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "branch_count must be a positive number between 1 and 100" {
		t.Errorf("expected 'branch_count must be a positive number between 1 and 100', got %v", err)
	}
}

func TestParallelConsolidationValidateInputBranchCountExceedMax(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"process_id":   "PROC001",
			"branch_count": 150.0,
			"branch_list": []interface{}{
				map[string]interface{}{"branch_id": "BR001", "branch_name": "Branch 1"},
			},
			"initiation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err == nil || err.Error() != "branch_count must be a positive number between 1 and 100" {
		t.Errorf("expected 'branch_count must be a positive number between 1 and 100', got %v", err)
	}
}

func TestParallelConsolidationValidateInputValidSingleBranch(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"process_id":   "PROC001",
			"branch_count": 1.0,
			"branch_list": []interface{}{
				map[string]interface{}{"branch_id": "BR001", "branch_name": "Branch 1"},
			},
			"initiation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for single branch, got %v", err)
	}
}

func TestParallelConsolidationValidateInputValidMultipleBranches(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"process_id":   "PROC001",
			"branch_count": 3.0,
			"branch_list": []interface{}{
				map[string]interface{}{"branch_id": "BR001", "branch_name": "Branch 1"},
				map[string]interface{}{"branch_id": "BR002", "branch_name": "Branch 2"},
				map[string]interface{}{"branch_id": "BR003", "branch_name": "Branch 3"},
			},
			"initiation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for multiple branches, got %v", err)
	}
}

func TestParallelConsolidationValidateInputValidMaxBranches(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	branches := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		branches[i] = map[string]interface{}{"branch_id": "BR" + string(rune(i)), "branch_name": "Branch"}
	}
	input := map[string]interface{}{
		"input": map[string]interface{}{
			"process_id":      "PROC001",
			"branch_count":    100.0,
			"branch_list":     branches,
			"initiation_date": "2026-02-16",
		},
		"companyID": "COMP001",
	}
	err := saga.ValidateInput(input)
	if err != nil {
		t.Errorf("expected no error for max branches, got %v", err)
	}
}

// Retry Configuration Tests
func TestParallelConsolidationRetryConfig(t *testing.T) {
	saga := NewParallelConsolidationSaga()
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

// TimeoutSeconds Tests
func TestParallelConsolidationTimeoutSeconds(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	expectedTimeouts := map[int]int32{
		1:  45,
		2:  40,
		3:  80,
		4:  80,
		5:  80,
		6:  70,
		7:  70,
		8:  70,
		9:  60,
		10: 50,
	}
	for stepNum, expectedTimeout := range expectedTimeouts {
		step := saga.GetStepDefinition(stepNum)
		if step.TimeoutSeconds != expectedTimeout {
			t.Errorf("step %d: expected timeout %d, got %d", stepNum, expectedTimeout, step.TimeoutSeconds)
		}
	}
}

// Service Names Tests
func TestParallelConsolidationServiceNames(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	expectedServices := map[int]string{
		1:  "workflow",
		2:  "workflow",
		3:  "branch",
		4:  "branch",
		5:  "branch",
		6:  "consolidation",
		7:  "consolidation",
		8:  "consolidation",
		9:  "reconciliation",
		10: "workflow",
	}
	for stepNum, expectedService := range expectedServices {
		step := saga.GetStepDefinition(stepNum)
		if step.ServiceName != expectedService {
			t.Errorf("step %d: expected service %s, got %s", stepNum, expectedService, step.ServiceName)
		}
	}
}

// InputMapping Tests
func TestParallelConsolidationInputMappings(t *testing.T) {
	saga := NewParallelConsolidationSaga()
	step1 := saga.GetStepDefinition(1)
	if step1.InputMapping["tenantID"] != "$.tenantID" {
		t.Error("tenantID mapping incorrect")
	}
	if step1.InputMapping["processID"] != "$.input.process_id" {
		t.Error("processID mapping incorrect")
	}
	if step1.InputMapping["branchList"] != "$.input.branch_list" {
		t.Error("branchList mapping incorrect")
	}
}

// ============================================================================
// Integration Tests (2 tests)
// ============================================================================

func TestSagaHandlerInterface(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewMultiLevelApprovalRoutingSaga(),
		NewConditionalWorkflowRoutingSaga(),
		NewParallelConsolidationSaga(),
	}
	for _, s := range sagas {
		if s.SagaType() == "" {
			t.Error("SagaType() should not be empty")
		}
		if len(s.GetStepDefinitions()) == 0 {
			t.Error("GetStepDefinitions() should not be empty")
		}
	}
}

func TestSagaStepNumberSequence(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewMultiLevelApprovalRoutingSaga(),
		NewConditionalWorkflowRoutingSaga(),
		NewParallelConsolidationSaga(),
	}
	for _, s := range sagas {
		steps := s.GetStepDefinitions()
		stepNumbers := make(map[int32]bool)
		for _, step := range steps {
			if stepNumbers[step.StepNumber] {
				t.Errorf("%s: duplicate step number %d", s.SagaType(), step.StepNumber)
			}
			stepNumbers[step.StepNumber] = true
		}
	}
}
