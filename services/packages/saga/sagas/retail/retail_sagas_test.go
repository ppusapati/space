// Package retail provides comprehensive unit tests for retail saga handlers
package retail

import (
	"errors"
	"strings"
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// Helper function to check if error message contains substring
func contains(str, substring string) bool {
	return strings.Contains(str, substring)
}

// ========== POS TRANSACTION SAGA TESTS (SAGA-R01) ==========

// TestPOSTransactionSagaType verifies POS Transaction saga returns correct type
func TestPOSTransactionSagaType(t *testing.T) {
	s := NewPOSTransactionSaga()
	if s.SagaType() != "SAGA-R01" {
		t.Errorf("expected SAGA-R01, got %s", s.SagaType())
	}
}

// TestPOSTransactionStepCount verifies step count matches specification
func TestPOSTransactionStepCount(t *testing.T) {
	s := NewPOSTransactionSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 23 // 12 forward + 11 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestPOSTransactionImplementsInterface verifies saga implements SagaHandler
func TestPOSTransactionImplementsInterface(t *testing.T) {
	s := NewPOSTransactionSaga()
	var _ saga.SagaHandler = s
}

// TestPOSTransactionValidation verifies input validation
func TestPOSTransactionValidation(t *testing.T) {
	s := NewPOSTransactionSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			"valid POS transaction input",
			map[string]interface{}{
				"transaction_id":    "txn-001",
				"terminal_id":       "term-001",
				"operator_id":       "op-001",
				"transaction_time":  "2026-02-16T10:00:00Z",
				"customer_id":       "cust-001",
				"item_details":      []interface{}{map[string]interface{}{"sku": "SKU001", "qty": 2}},
				"total_amount":      1000.00,
				"payment_method":    "CARD",
				"payment_reference": "REF123",
			},
			false,
			"",
		},
		{
			"missing transaction_id",
			map[string]interface{}{
				"terminal_id":      "term-001",
				"operator_id":      "op-001",
				"transaction_time": "2026-02-16T10:00:00Z",
				"customer_id":      "cust-001",
				"item_details":     []interface{}{map[string]interface{}{"sku": "SKU001"}},
				"total_amount":     1000.00,
				"payment_method":   "CARD",
			},
			true,
			"transaction_id is required",
		},
		{
			"missing terminal_id",
			map[string]interface{}{
				"transaction_id":   "txn-001",
				"operator_id":      "op-001",
				"transaction_time": "2026-02-16T10:00:00Z",
				"customer_id":      "cust-001",
				"item_details":     []interface{}{map[string]interface{}{"sku": "SKU001"}},
				"total_amount":     1000.00,
				"payment_method":   "CARD",
			},
			true,
			"terminal_id is required",
		},
		{
			"invalid total_amount",
			map[string]interface{}{
				"transaction_id":   "txn-001",
				"terminal_id":      "term-001",
				"operator_id":      "op-001",
				"transaction_time": "2026-02-16T10:00:00Z",
				"customer_id":      "cust-001",
				"item_details":     []interface{}{map[string]interface{}{"sku": "SKU001"}},
				"total_amount":     -100.00,
				"payment_method":   "CARD",
			},
			true,
			"total_amount must be a positive number",
		},
		{
			"missing item_details",
			map[string]interface{}{
				"transaction_id":   "txn-001",
				"terminal_id":      "term-001",
				"operator_id":      "op-001",
				"transaction_time": "2026-02-16T10:00:00Z",
				"customer_id":      "cust-001",
				"total_amount":     1000.00,
				"payment_method":   "CARD",
			},
			true,
			"item_details are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.hasErr && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

// TestPOSTransactionCriticalSteps verifies critical steps are marked correctly
func TestPOSTransactionCriticalSteps(t *testing.T) {
	s := NewPOSTransactionSaga()
	steps := s.GetStepDefinitions()

	criticalSteps := []int{1, 2, 3, 4, 7, 8, 12}
	for _, step := range criticalSteps {
		stepDef := s.GetStepDefinition(step)
		if !stepDef.IsCritical {
			t.Errorf("step %d should be marked as critical", step)
		}
	}
}

// TestPOSTransactionCompensation verifies compensation steps exist
func TestPOSTransactionCompensation(t *testing.T) {
	s := NewPOSTransactionSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount != 11 {
		t.Errorf("expected 11 compensation steps, got %d", compensationCount)
	}
}

// ========== INVENTORY SYNC SAGA TESTS (SAGA-R02) ==========

// TestInventorySyncSagaType verifies Inventory Sync saga returns correct type
func TestInventorySyncSagaType(t *testing.T) {
	s := NewInventorySyncSaga()
	if s.SagaType() != "SAGA-R02" {
		t.Errorf("expected SAGA-R02, got %s", s.SagaType())
	}
}

// TestInventorySyncStepCount verifies step count
func TestInventorySyncStepCount(t *testing.T) {
	s := NewInventorySyncSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 21 // 11 forward + 10 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestInventorySyncImplementsInterface verifies saga implements SagaHandler
func TestInventorySyncImplementsInterface(t *testing.T) {
	s := NewInventorySyncSaga()
	var _ saga.SagaHandler = s
}

// TestInventorySyncValidation verifies input validation
func TestInventorySyncValidation(t *testing.T) {
	s := NewInventorySyncSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			"valid inventory sync input",
			map[string]interface{}{
				"sync_run_id":  "sync-001",
				"sync_type":    "FULL",
				"channel_list": []interface{}{"POS", "ECOMMERCE", "WAREHOUSE"},
				"sync_time":    "2026-02-16T10:00:00Z",
			},
			false,
			"",
		},
		{
			"missing sync_run_id",
			map[string]interface{}{
				"sync_type":    "FULL",
				"channel_list": []interface{}{"POS", "ECOMMERCE"},
				"sync_time":    "2026-02-16T10:00:00Z",
			},
			true,
			"sync_run_id is required",
		},
		{
			"invalid sync_type",
			map[string]interface{}{
				"sync_run_id":  "sync-001",
				"sync_type":    "INVALID",
				"channel_list": []interface{}{"POS"},
				"sync_time":    "2026-02-16T10:00:00Z",
			},
			true,
			"sync_type must be FULL or INCREMENTAL",
		},
		{
			"empty channel_list",
			map[string]interface{}{
				"sync_run_id":  "sync-001",
				"sync_type":    "FULL",
				"channel_list": []interface{}{},
				"sync_time":    "2026-02-16T10:00:00Z",
			},
			true,
			"channel_list cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.hasErr && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

// TestInventorySyncMultiChannelInput verifies multi-channel sync input handling
func TestInventorySyncMultiChannelInput(t *testing.T) {
	s := NewInventorySyncSaga()

	validInput := map[string]interface{}{
		"sync_run_id":  "sync-001",
		"sync_type":    "FULL",
		"channel_list": []interface{}{"POS", "ECOMMERCE", "WAREHOUSE"},
		"sync_time":    "2026-02-16T10:00:00Z",
	}

	err := s.ValidateInput(validInput)
	if err != nil {
		t.Errorf("expected valid multi-channel input to pass, got error: %v", err)
	}
}

// TestInventorySyncConflictResolution verifies conflict resolution steps
func TestInventorySyncConflictResolution(t *testing.T) {
	s := NewInventorySyncSaga()
	stepDef := s.GetStepDefinition(7) // ReconcileInventory step

	if stepDef == nil {
		t.Error("step 7 (ReconcileInventory) not found")
	}
	if stepDef.HandlerMethod != "ReconcileInventory" {
		t.Errorf("step 7 handler should be ReconcileInventory, got %s", stepDef.HandlerMethod)
	}
}

// TestInventorySyncCompensation verifies compensation steps
func TestInventorySyncCompensation(t *testing.T) {
	s := NewInventorySyncSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount != 10 {
		t.Errorf("expected 10 compensation steps, got %d", compensationCount)
	}
}

// ========== RETURN/REFUND SAGA TESTS (SAGA-R03) ==========

// TestReturnRefundSagaType verifies Return/Refund saga returns correct type
func TestReturnRefundSagaType(t *testing.T) {
	s := NewReturnRefundSaga()
	if s.SagaType() != "SAGA-R03" {
		t.Errorf("expected SAGA-R03, got %s", s.SagaType())
	}
}

// TestReturnRefundStepCount verifies step count
func TestReturnRefundStepCount(t *testing.T) {
	s := NewReturnRefundSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 19 // 10 forward + 9 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestReturnRefundImplementsInterface verifies saga implements SagaHandler
func TestReturnRefundImplementsInterface(t *testing.T) {
	s := NewReturnRefundSaga()
	var _ saga.SagaHandler = s
}

// TestReturnRefundValidation verifies input validation
func TestReturnRefundValidation(t *testing.T) {
	s := NewReturnRefundSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			"valid return/refund input",
			map[string]interface{}{
				"return_id":       "ret-001",
				"original_txn_id": "txn-001",
				"customer_id":     "cust-001",
				"return_reason":   "DAMAGED",
				"return_items":    []interface{}{map[string]interface{}{"sku": "SKU001", "qty": 1}},
				"return_amount":   500.00,
				"original_amount": 1000.00,
			},
			false,
			"",
		},
		{
			"missing return_id",
			map[string]interface{}{
				"original_txn_id": "txn-001",
				"customer_id":     "cust-001",
				"return_reason":   "DAMAGED",
				"return_items":    []interface{}{map[string]interface{}{"sku": "SKU001"}},
				"return_amount":   500.00,
			},
			true,
			"return_id is required",
		},
		{
			"invalid return_amount",
			map[string]interface{}{
				"return_id":       "ret-001",
				"original_txn_id": "txn-001",
				"customer_id":     "cust-001",
				"return_reason":   "DAMAGED",
				"return_items":    []interface{}{map[string]interface{}{"sku": "SKU001"}},
				"return_amount":   -500.00,
				"original_amount": 1000.00,
			},
			true,
			"return_amount must be a positive number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.hasErr && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

// TestRefundCalculationInput verifies refund calculation step
func TestRefundCalculationInput(t *testing.T) {
	s := NewReturnRefundSaga()
	stepDef := s.GetStepDefinition(5) // Refund Calculation

	if stepDef == nil {
		t.Error("refund calculation step not found")
	}
}

// TestReturnInspectionInput verifies return inspection step
func TestReturnInspectionInput(t *testing.T) {
	s := NewReturnRefundSaga()
	stepDef := s.GetStepDefinition(3) // Return Inspection

	if stepDef == nil {
		t.Error("return inspection step not found")
	}
}

// TestRefundProcessingInput verifies refund processing step
func TestRefundProcessingInput(t *testing.T) {
	s := NewReturnRefundSaga()
	stepDef := s.GetStepDefinition(8) // Refund Processing

	if stepDef == nil {
		t.Error("refund processing step not found")
	}
}

// TestReturnRefundCompensation verifies compensation steps
func TestReturnRefundCompensation(t *testing.T) {
	s := NewReturnRefundSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount != 9 {
		t.Errorf("expected 9 compensation steps, got %d", compensationCount)
	}
}

// ========== LOYALTY PROGRAM SAGA TESTS (SAGA-R04) ==========

// TestLoyaltyProgramSagaType verifies Loyalty Program saga returns correct type
func TestLoyaltyProgramSagaType(t *testing.T) {
	s := NewLoyaltyProgramSaga()
	if s.SagaType() != "SAGA-R04" {
		t.Errorf("expected SAGA-R04, got %s", s.SagaType())
	}
}

// TestLoyaltyProgramStepCount verifies step count
func TestLoyaltyProgramStepCount(t *testing.T) {
	s := NewLoyaltyProgramSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 17 // 9 forward + 8 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestLoyaltyProgramImplementsInterface verifies saga implements SagaHandler
func TestLoyaltyProgramImplementsInterface(t *testing.T) {
	s := NewLoyaltyProgramSaga()
	var _ saga.SagaHandler = s
}

// TestLoyaltyProgramValidation verifies input validation
func TestLoyaltyProgramValidation(t *testing.T) {
	s := NewLoyaltyProgramSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			"valid loyalty program input",
			map[string]interface{}{
				"loyalty_id":         "loy-001",
				"customer_id":        "cust-001",
				"program_type":       "POINTS",
				"transaction_amount": 1000.00,
			},
			false,
			"",
		},
		{
			"missing loyalty_id",
			map[string]interface{}{
				"customer_id":        "cust-001",
				"program_type":       "POINTS",
				"transaction_amount": 1000.00,
			},
			true,
			"loyalty_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestLoyaltyEnrollmentInput verifies enrollment step
func TestLoyaltyEnrollmentInput(t *testing.T) {
	s := NewLoyaltyProgramSaga()
	stepDef := s.GetStepDefinition(1)

	if stepDef == nil {
		t.Error("step 1 not found")
	}
}

// TestPointsCalculationInput verifies points calculation step
func TestPointsCalculationInput(t *testing.T) {
	s := NewLoyaltyProgramSaga()
	stepDef := s.GetStepDefinition(4)

	if stepDef == nil {
		t.Error("step 4 not found")
	}
}

// TestTierUpgradeInput verifies tier upgrade step
func TestTierUpgradeInput(t *testing.T) {
	s := NewLoyaltyProgramSaga()
	stepDef := s.GetStepDefinition(6)

	if stepDef == nil {
		t.Error("step 6 not found")
	}
}

// TestRewardRedemptionInput verifies reward redemption step
func TestRewardRedemptionInput(t *testing.T) {
	s := NewLoyaltyProgramSaga()
	stepDef := s.GetStepDefinition(8)

	if stepDef == nil {
		t.Error("step 8 not found")
	}
}

// TestLoyaltyProgramCompensation verifies compensation steps
func TestLoyaltyProgramCompensation(t *testing.T) {
	s := NewLoyaltyProgramSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount != 8 {
		t.Errorf("expected 8 compensation steps, got %d", compensationCount)
	}
}

// ========== PROMOTION SAGA TESTS (SAGA-R05) ==========

// TestPromotionSagaType verifies Promotion saga returns correct type
func TestPromotionSagaType(t *testing.T) {
	s := NewPromotionSaga()
	if s.SagaType() != "SAGA-R05" {
		t.Errorf("expected SAGA-R05, got %s", s.SagaType())
	}
}

// TestPromotionStepCount verifies step count
func TestPromotionStepCount(t *testing.T) {
	s := NewPromotionSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 19 // 10 forward + 9 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestPromotionImplementsInterface verifies saga implements SagaHandler
func TestPromotionImplementsInterface(t *testing.T) {
	s := NewPromotionSaga()
	var _ saga.SagaHandler = s
}

// TestPromotionValidation verifies input validation
func TestPromotionValidation(t *testing.T) {
	s := NewPromotionSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			"valid promotion input",
			map[string]interface{}{
				"promo_id":       "promo-001",
				"promo_name":     "Spring Sale",
				"promo_type":     "PERCENTAGE",
				"discount_value": 15.0,
				"start_date":     "2026-02-16",
				"end_date":       "2026-03-31",
				"target_skus":    []interface{}{"SKU001", "SKU002"},
				"max_uses":       1000,
			},
			false,
			"",
		},
		{
			"missing promo_id",
			map[string]interface{}{
				"promo_name":     "Spring Sale",
				"promo_type":     "PERCENTAGE",
				"discount_value": 15.0,
				"start_date":     "2026-02-16",
				"end_date":       "2026-03-31",
			},
			true,
			"promo_id is required",
		},
		{
			"invalid discount_value",
			map[string]interface{}{
				"promo_id":       "promo-001",
				"promo_name":     "Spring Sale",
				"promo_type":     "PERCENTAGE",
				"discount_value": -10.0,
				"start_date":     "2026-02-16",
				"end_date":       "2026-03-31",
			},
			true,
			"discount_value must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestPromotionRuleValidation verifies promotion rule validation step
func TestPromotionRuleValidation(t *testing.T) {
	s := NewPromotionSaga()
	stepDef := s.GetStepDefinition(2)

	if stepDef == nil {
		t.Error("step 2 not found")
	}
}

// TestPromotionPricingCalculation verifies pricing calculation step
func TestPromotionPricingCalculation(t *testing.T) {
	s := NewPromotionSaga()
	stepDef := s.GetStepDefinition(4)

	if stepDef == nil {
		t.Error("step 4 not found")
	}
}

// TestPromotionApprovalInput verifies approval step
func TestPromotionApprovalInput(t *testing.T) {
	s := NewPromotionSaga()
	stepDef := s.GetStepDefinition(6)

	if stepDef == nil {
		t.Error("step 6 not found")
	}
}

// TestPromotionInventoryAllocationInput verifies inventory allocation step
func TestPromotionInventoryAllocationInput(t *testing.T) {
	s := NewPromotionSaga()
	stepDef := s.GetStepDefinition(7)

	if stepDef == nil {
		t.Error("step 7 not found")
	}
}

// TestPromotionCompensation verifies compensation steps
func TestPromotionCompensation(t *testing.T) {
	s := NewPromotionSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount != 9 {
		t.Errorf("expected 9 compensation steps, got %d", compensationCount)
	}
}

// ========== STOCK TRANSFER SAGA TESTS (SAGA-R06) ==========

// TestStockTransferSagaType verifies Stock Transfer saga returns correct type
func TestStockTransferSagaType(t *testing.T) {
	s := NewStockTransferSaga()
	if s.SagaType() != "SAGA-R06" {
		t.Errorf("expected SAGA-R06, got %s", s.SagaType())
	}
}

// TestStockTransferStepCount verifies step count
func TestStockTransferStepCount(t *testing.T) {
	s := NewStockTransferSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 19 // 10 forward + 9 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestStockTransferImplementsInterface verifies saga implements SagaHandler
func TestStockTransferImplementsInterface(t *testing.T) {
	s := NewStockTransferSaga()
	var _ saga.SagaHandler = s
}

// TestStockTransferValidation verifies input validation
func TestStockTransferValidation(t *testing.T) {
	s := NewStockTransferSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			"valid stock transfer input",
			map[string]interface{}{
				"transfer_id":     "xfer-001",
				"source_location": "LOC001",
				"dest_location":   "LOC002",
				"transfer_items":  []interface{}{map[string]interface{}{"sku": "SKU001", "qty": 50}},
				"transfer_qty":    50,
			},
			false,
			"",
		},
		{
			"missing transfer_id",
			map[string]interface{}{
				"source_location": "LOC001",
				"dest_location":   "LOC002",
				"transfer_items":  []interface{}{map[string]interface{}{"sku": "SKU001", "qty": 50}},
				"transfer_qty":    50,
			},
			true,
			"transfer_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestSourceLocationInput verifies source location step
func TestSourceLocationInput(t *testing.T) {
	s := NewStockTransferSaga()
	stepDef := s.GetStepDefinition(2)

	if stepDef == nil {
		t.Error("step 2 not found")
	}
}

// TestDestinationLocationInput verifies destination location step
func TestDestinationLocationInput(t *testing.T) {
	s := NewStockTransferSaga()
	stepDef := s.GetStepDefinition(3)

	if stepDef == nil {
		t.Error("step 3 not found")
	}
}

// TestQuantityValidationTransfer verifies quantity validation
func TestQuantityValidationTransfer(t *testing.T) {
	s := NewStockTransferSaga()
	stepDef := s.GetStepDefinition(4)

	if stepDef == nil {
		t.Error("step 4 not found")
	}
}

// TestInventoryDeductionInput verifies inventory deduction step
func TestInventoryDeductionInput(t *testing.T) {
	s := NewStockTransferSaga()
	stepDef := s.GetStepDefinition(5)

	if stepDef == nil {
		t.Error("step 5 not found")
	}
}

// TestStockTransferCompensation verifies compensation steps
func TestStockTransferCompensation(t *testing.T) {
	s := NewStockTransferSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount != 9 {
		t.Errorf("expected 9 compensation steps, got %d", compensationCount)
	}
}

// ========== CUSTOMER ACCOUNT SAGA TESTS (SAGA-R07) ==========

// TestCustomerAccountSagaType verifies Customer Account saga returns correct type
func TestCustomerAccountSagaType(t *testing.T) {
	s := NewCustomerAccountSaga()
	if s.SagaType() != "SAGA-R07" {
		t.Errorf("expected SAGA-R07, got %s", s.SagaType())
	}
}

// TestCustomerAccountStepCount verifies step count
func TestCustomerAccountStepCount(t *testing.T) {
	s := NewCustomerAccountSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 15 // 8 forward + 7 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestCustomerAccountImplementsInterface verifies saga implements SagaHandler
func TestCustomerAccountImplementsInterface(t *testing.T) {
	s := NewCustomerAccountSaga()
	var _ saga.SagaHandler = s
}

// TestCustomerAccountValidation verifies input validation
func TestCustomerAccountValidation(t *testing.T) {
	s := NewCustomerAccountSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			"valid customer account input",
			map[string]interface{}{
				"customer_id": "cust-001",
				"first_name":  "John",
				"last_name":   "Doe",
				"email":       "john@example.com",
				"phone":       "+91-9876543210",
			},
			false,
			"",
		},
		{
			"missing customer_id",
			map[string]interface{}{
				"first_name": "John",
				"last_name":  "Doe",
				"email":      "john@example.com",
			},
			true,
			"customer_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestCustomerProfileCreationInput verifies profile creation step
func TestCustomerProfileCreationInput(t *testing.T) {
	s := NewCustomerAccountSaga()
	stepDef := s.GetStepDefinition(1)

	if stepDef == nil {
		t.Error("step 1 not found")
	}
}

// TestAddressSetupInput verifies address setup step
func TestAddressSetupInput(t *testing.T) {
	s := NewCustomerAccountSaga()
	stepDef := s.GetStepDefinition(3)

	if stepDef == nil {
		t.Error("step 3 not found")
	}
}

// TestCustomerAccountCompensation verifies compensation steps
func TestCustomerAccountCompensation(t *testing.T) {
	s := NewCustomerAccountSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount != 7 {
		t.Errorf("expected 7 compensation steps, got %d", compensationCount)
	}
}

// ========== INTEGRATION TESTS ==========

// TestRetailSagasInterface verifies all retail sagas implement SagaHandler
func TestRetailSagasInterface(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewPOSTransactionSaga(),
		NewInventorySyncSaga(),
		NewReturnRefundSaga(),
		NewLoyaltyProgramSaga(),
		NewPromotionSaga(),
		NewStockTransferSaga(),
		NewCustomerAccountSaga(),
	}

	for _, s := range sagas {
		if s == nil {
			t.Error("saga is nil")
		}
		if s.SagaType() == "" {
			t.Error("saga type is empty")
		}
		if len(s.GetStepDefinitions()) == 0 {
			t.Errorf("saga %s has no steps", s.SagaType())
		}
	}
}

// TestRetailSagaTypes verifies all saga types return correct identifiers
func TestRetailSagaTypes(t *testing.T) {
	sagas := []struct {
		name         string
		saga         saga.SagaHandler
		expectedType string
	}{
		{"POS Transaction", NewPOSTransactionSaga(), "SAGA-R01"},
		{"Inventory Sync", NewInventorySyncSaga(), "SAGA-R02"},
		{"Return/Refund", NewReturnRefundSaga(), "SAGA-R03"},
		{"Loyalty Program", NewLoyaltyProgramSaga(), "SAGA-R04"},
		{"Promotion", NewPromotionSaga(), "SAGA-R05"},
		{"Stock Transfer", NewStockTransferSaga(), "SAGA-R06"},
		{"Customer Account", NewCustomerAccountSaga(), "SAGA-R07"},
	}

	for _, tt := range sagas {
		t.Run(tt.name, func(t *testing.T) {
			if tt.saga.SagaType() != tt.expectedType {
				t.Errorf("expected %s, got %s", tt.expectedType, tt.saga.SagaType())
			}
		})
	}
}

// TestRetailSagasGetStepDefinitions verifies step retrieval
func TestRetailSagasGetStepDefinitions(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewPOSTransactionSaga(),
		NewInventorySyncSaga(),
		NewReturnRefundSaga(),
		NewLoyaltyProgramSaga(),
		NewPromotionSaga(),
		NewStockTransferSaga(),
		NewCustomerAccountSaga(),
	}

	for _, s := range sagas {
		steps := s.GetStepDefinitions()
		if len(steps) == 0 {
			t.Errorf("saga %s has no steps", s.SagaType())
		}

		// Verify first step exists
		firstStep := s.GetStepDefinition(1)
		if firstStep == nil {
			t.Errorf("saga %s: step 1 not found", s.SagaType())
		}
	}
}

// TestRetailSagasInvalidStepLookup verifies invalid step lookup returns nil
func TestRetailSagasInvalidStepLookup(t *testing.T) {
	s := NewPOSTransactionSaga()
	invalidStep := s.GetStepDefinition(999)
	if invalidStep != nil {
		t.Error("invalid step should return nil")
	}
}

// TestRetailSagasNilInput verifies nil input handling
func TestRetailSagasNilInput(t *testing.T) {
	s := NewPOSTransactionSaga()
	err := s.ValidateInput(nil)
	if err == nil {
		t.Error("expected error for nil input")
	}
}

// TestRetailSagasEmptyMapInput verifies empty map input handling
func TestRetailSagasEmptyMapInput(t *testing.T) {
	s := NewPOSTransactionSaga()
	err := s.ValidateInput(map[string]interface{}{})
	if err == nil {
		t.Error("expected error for empty map input")
	}
}

// TestRetailSagasStringInput verifies string input rejection
func TestRetailSagasStringInput(t *testing.T) {
	s := NewInventorySyncSaga()
	err := s.ValidateInput("invalid string")
	if err == nil {
		t.Error("expected error for string input")
	}
}

// TestRetailSagasIntInput verifies integer input rejection
func TestRetailSagasIntInput(t *testing.T) {
	s := NewReturnRefundSaga()
	err := s.ValidateInput(12345)
	if err == nil {
		t.Error("expected error for integer input")
	}
}

// TestRetailSagasCriticalStepMarking verifies critical steps are marked
func TestRetailSagasCriticalStepMarking(t *testing.T) {
	s := NewPOSTransactionSaga()
	steps := s.GetStepDefinitions()

	hasCritical := false
	for _, step := range steps {
		if step.IsCritical {
			hasCritical = true
			break
		}
	}

	if !hasCritical {
		t.Error("no critical steps found in POS transaction saga")
	}
}

// TestRetailSagasRetryConfiguration verifies retry config exists
func TestRetailSagasRetryConfiguration(t *testing.T) {
	s := NewPOSTransactionSaga()
	stepDef := s.GetStepDefinition(1)

	if stepDef.RetryConfig == nil {
		t.Error("retry configuration missing in step 1")
	}
}
