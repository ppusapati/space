// Tests for the supplychain saga handlers. Package name matches the rest
// of the supply-chain/ directory (supplychain — no underscore) so Go treats
// them as one package. B.8 fix 2026-04-19.
package supplychain

import (
	"strings"
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// Helper function to check if error message contains substring
func contains(str, substring string) bool {
	return strings.Contains(str, substring)
}

// ========== INBOUND LOGISTICS SAGA TESTS (SAGA-SC01) ==========

// TestInboundLogisticsSagaType verifies Inbound Logistics saga returns correct type
func TestInboundLogisticsSagaType(t *testing.T) {
	s := NewInboundLogisticsSaga()
	if s.SagaType() != "SAGA-SC01" {
		t.Errorf("expected SAGA-SC01, got %s", s.SagaType())
	}
}

// TestInboundLogisticsStepCount verifies step count
func TestInboundLogisticsStepCount(t *testing.T) {
	s := NewInboundLogisticsSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 21 // 12 forward + 9 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestInboundLogisticsImplementsInterface verifies saga implements SagaHandler
func TestInboundLogisticsImplementsInterface(t *testing.T) {
	s := NewInboundLogisticsSaga()
	var _ saga.SagaHandler = s
}

// TestInboundLogisticsValidation verifies input validation
func TestInboundLogisticsValidation(t *testing.T) {
	s := NewInboundLogisticsSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid inbound logistics input",
			map[string]interface{}{
				"inbound_id":      "inb-001",
				"po_id":           "po-001",
				"shipment_id":     "ship-001",
				"warehouse_id":    "wh-001",
				"expected_qty":    100,
				"received_qty":    100,
				"receiving_date":  "2026-02-16",
			},
			false,
			"",
		},
		{
			"missing inbound_id",
			map[string]interface{}{
				"po_id":          "po-001",
				"shipment_id":    "ship-001",
				"warehouse_id":   "wh-001",
				"expected_qty":   100,
				"received_qty":   100,
				"receiving_date": "2026-02-16",
			},
			true,
			"inbound_id is required",
		},
		{
			"qty mismatch",
			map[string]interface{}{
				"inbound_id":     "inb-001",
				"po_id":          "po-001",
				"shipment_id":    "ship-001",
				"warehouse_id":   "wh-001",
				"expected_qty":   100,
				"received_qty":   95,
				"receiving_date": "2026-02-16",
			},
			false,
			"",
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

// TestShipmentReceiptInput verifies shipment receipt step
func TestShipmentReceiptInput(t *testing.T) {
	s := NewInboundLogisticsSaga()
	stepDef := s.GetStepDefinition(2)

	if stepDef == nil {
		t.Error("step 2 (Shipment Receipt) not found")
	}
}

// TestQualityInspectionInput verifies quality inspection step
func TestQualityInspectionInput(t *testing.T) {
	s := NewInboundLogisticsSaga()
	stepDef := s.GetStepDefinition(4)

	if stepDef == nil {
		t.Error("step 4 (Quality Inspection) not found")
	}
}

// TestPutAwayInput verifies put away step
func TestPutAwayInput(t *testing.T) {
	s := NewInboundLogisticsSaga()
	stepDef := s.GetStepDefinition(6)

	if stepDef == nil {
		t.Error("step 6 (Put Away) not found")
	}
}

// TestReceiptGLPostingInput verifies GL posting step
func TestReceiptGLPostingInput(t *testing.T) {
	s := NewInboundLogisticsSaga()
	stepDef := s.GetStepDefinition(10)

	if stepDef == nil {
		t.Error("step 10 (GL Posting) not found")
	}
}

// TestInboundLogisticsCompensation verifies compensation steps
func TestInboundLogisticsCompensation(t *testing.T) {
	s := NewInboundLogisticsSaga()
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

// ========== WAREHOUSE OPERATIONS SAGA TESTS (SAGA-SC02) ==========

// TestWarehouseOpsSagaType verifies Warehouse Operations saga returns correct type
func TestWarehouseOpsSagaType(t *testing.T) {
	s := NewWarehouseOpsSaga()
	if s.SagaType() != "SAGA-SC02" {
		t.Errorf("expected SAGA-SC02, got %s", s.SagaType())
	}
}

// TestWarehouseOpsStepCount verifies step count
func TestWarehouseOpsStepCount(t *testing.T) {
	s := NewWarehouseOpsSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 19 // 11 forward + 8 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestWarehouseOpsImplementsInterface verifies saga implements SagaHandler
func TestWarehouseOpsImplementsInterface(t *testing.T) {
	s := NewWarehouseOpsSaga()
	var _ saga.SagaHandler = s
}

// TestWarehouseOpsValidation verifies input validation
func TestWarehouseOpsValidation(t *testing.T) {
	s := NewWarehouseOpsSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid warehouse ops input",
			map[string]interface{}{
				"order_id":      "ord-001",
				"warehouse_id":  "wh-001",
				"pick_items":    []interface{}{map[string]interface{}{"sku": "SKU001", "qty": 5}},
				"order_date":    "2026-02-16",
			},
			false,
			"",
		},
		{
			"missing order_id",
			map[string]interface{}{
				"warehouse_id": "wh-001",
				"pick_items":   []interface{}{map[string]interface{}{"sku": "SKU001", "qty": 5}},
				"order_date":   "2026-02-16",
			},
			true,
			"order_id is required",
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

// TestPickingInput verifies picking step
func TestPickingInput(t *testing.T) {
	s := NewWarehouseOpsSaga()
	stepDef := s.GetStepDefinition(3)

	if stepDef == nil {
		t.Error("step 3 (Picking) not found")
	}
}

// TestPackingInput verifies packing step
func TestPackingInput(t *testing.T) {
	s := NewWarehouseOpsSaga()
	stepDef := s.GetStepDefinition(5)

	if stepDef == nil {
		t.Error("step 5 (Packing) not found")
	}
}

// TestShippingInput verifies shipping step
func TestShippingInput(t *testing.T) {
	s := NewWarehouseOpsSaga()
	stepDef := s.GetStepDefinition(7)

	if stepDef == nil {
		t.Error("step 7 (Shipping) not found")
	}
}

// TestWarehouseOpsCompensation verifies compensation steps
func TestWarehouseOpsCompensation(t *testing.T) {
	s := NewWarehouseOpsSaga()
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

// ========== 3PL COORDINATION SAGA TESTS (SAGA-SC03) ==========

// TestThirdPartyLogisticsCoordinationSagaType verifies 3PL saga returns correct type
func TestThirdPartyLogisticsCoordinationSagaType(t *testing.T) {
	s := NewThirdPartyLogisticsCoordinationSaga()
	if s.SagaType() != "SAGA-SC03" {
		t.Errorf("expected SAGA-SC03, got %s", s.SagaType())
	}
}

// TestThirdPartyLogisticsCoordinationStepCount verifies step count
func TestThirdPartyLogisticsCoordinationStepCount(t *testing.T) {
	s := NewThirdPartyLogisticsCoordinationSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 19 // 11 forward + 8 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestThirdPartyLogisticsCoordinationImplementsInterface verifies saga implements SagaHandler
func TestThirdPartyLogisticsCoordinationImplementsInterface(t *testing.T) {
	s := NewThirdPartyLogisticsCoordinationSaga()
	var _ saga.SagaHandler = s
}

// TestThirdPartyLogisticsCoordinationValidation verifies input validation
func TestThirdPartyLogisticsCoordinationValidation(t *testing.T) {
	s := NewThirdPartyLogisticsCoordinationSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid 3PL input",
			map[string]interface{}{
				"shipment_id":   "ship-001",
				"origin":        "LOC001",
				"destination":   "LOC002",
				"carrier_id":    "carr-001",
				"shipment_date": "2026-02-16",
				"weight_kg":     50,
			},
			false,
			"",
		},
		{
			"missing shipment_id",
			map[string]interface{}{
				"origin":        "LOC001",
				"destination":   "LOC002",
				"carrier_id":    "carr-001",
				"shipment_date": "2026-02-16",
			},
			true,
			"shipment_id is required",
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

// TestCarrierSelectionInput verifies carrier selection step
func TestCarrierSelectionInput(t *testing.T) {
	s := NewThirdPartyLogisticsCoordinationSaga()
	stepDef := s.GetStepDefinition(3)

	if stepDef == nil {
		t.Error("step 3 (Carrier Selection) not found")
	}
}

// TestShipmentTrackingInput verifies shipment tracking step
func TestShipmentTrackingInput(t *testing.T) {
	s := NewThirdPartyLogisticsCoordinationSaga()
	stepDef := s.GetStepDefinition(6)

	if stepDef == nil {
		t.Error("step 6 (Shipment Tracking) not found")
	}
}

// TestCarrierPaymentInput verifies carrier payment step
func TestCarrierPaymentInput(t *testing.T) {
	s := NewThirdPartyLogisticsCoordinationSaga()
	stepDef := s.GetStepDefinition(8)

	if stepDef == nil {
		t.Error("step 8 (Carrier Payment) not found")
	}
}

// TestThirdPartyLogisticsCoordinationCompensation verifies compensation steps
func TestThirdPartyLogisticsCoordinationCompensation(t *testing.T) {
	s := NewThirdPartyLogisticsCoordinationSaga()
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

// ========== ORDER FULFILLMENT SAGA TESTS (SAGA-SC04) ==========

// TestOrderFulfillmentSagaType verifies Order Fulfillment saga returns correct type
func TestOrderFulfillmentSagaType(t *testing.T) {
	s := NewOrderFulfillmentSaga()
	if s.SagaType() != "SAGA-SC04" {
		t.Errorf("expected SAGA-SC04, got %s", s.SagaType())
	}
}

// TestOrderFulfillmentStepCount verifies step count
func TestOrderFulfillmentStepCount(t *testing.T) {
	s := NewOrderFulfillmentSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 21 // 12 forward + 9 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestOrderFulfillmentImplementsInterface verifies saga implements SagaHandler
func TestOrderFulfillmentImplementsInterface(t *testing.T) {
	s := NewOrderFulfillmentSaga()
	var _ saga.SagaHandler = s
}

// TestOrderFulfillmentValidation verifies input validation
func TestOrderFulfillmentValidation(t *testing.T) {
	s := NewOrderFulfillmentSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid order fulfillment input",
			map[string]interface{}{
				"order_id":      "ord-001",
				"customer_id":   "cust-001",
				"fulfillment_items": []interface{}{map[string]interface{}{"sku": "SKU001", "qty": 2}},
				"delivery_date": "2026-02-20",
				"warehouse_id":  "wh-001",
			},
			false,
			"",
		},
		{
			"missing order_id",
			map[string]interface{}{
				"customer_id": "cust-001",
				"fulfillment_items": []interface{}{map[string]interface{}{"sku": "SKU001", "qty": 2}},
				"delivery_date": "2026-02-20",
			},
			true,
			"order_id is required",
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

// TestOrderVerificationInput verifies order verification step
func TestOrderVerificationInput(t *testing.T) {
	s := NewOrderFulfillmentSaga()
	stepDef := s.GetStepDefinition(2)

	if stepDef == nil {
		t.Error("step 2 (Order Verification) not found")
	}
}

// TestInventoryReservationInput verifies inventory reservation step
func TestInventoryReservationInput(t *testing.T) {
	s := NewOrderFulfillmentSaga()
	stepDef := s.GetStepDefinition(4)

	if stepDef == nil {
		t.Error("step 4 (Inventory Reservation) not found")
	}
}

// TestWarehousePickingInput verifies warehouse picking step
func TestWarehousePickingInput(t *testing.T) {
	s := NewOrderFulfillmentSaga()
	stepDef := s.GetStepDefinition(6)

	if stepDef == nil {
		t.Error("step 6 (Warehouse Picking) not found")
	}
}

// TestInvoiceGenerationInput verifies invoice generation step
func TestInvoiceGenerationInput(t *testing.T) {
	s := NewOrderFulfillmentSaga()
	stepDef := s.GetStepDefinition(8)

	if stepDef == nil {
		t.Error("step 8 (Invoice Generation) not found")
	}
}

// TestOrderFulfillmentCompensation verifies compensation steps
func TestOrderFulfillmentCompensation(t *testing.T) {
	s := NewOrderFulfillmentSaga()
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

// ========== DISTRIBUTION CENTER SAGA TESTS (SAGA-SC05) ==========

// TestDistributionCenterSagaType verifies Distribution Center saga returns correct type
func TestDistributionCenterSagaType(t *testing.T) {
	s := NewDistributionCenterSaga()
	if s.SagaType() != "SAGA-SC05" {
		t.Errorf("expected SAGA-SC05, got %s", s.SagaType())
	}
}

// TestDistributionCenterStepCount verifies step count
func TestDistributionCenterStepCount(t *testing.T) {
	s := NewDistributionCenterSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 17 // 10 forward + 7 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestDistributionCenterImplementsInterface verifies saga implements SagaHandler
func TestDistributionCenterImplementsInterface(t *testing.T) {
	s := NewDistributionCenterSaga()
	var _ saga.SagaHandler = s
}

// TestDistributionCenterValidation verifies input validation
func TestDistributionCenterValidation(t *testing.T) {
	s := NewDistributionCenterSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid distribution center input",
			map[string]interface{}{
				"dc_id":      "dc-001",
				"inbound_items": []interface{}{map[string]interface{}{"sku": "SKU001", "qty": 100}},
				"process_date": "2026-02-16",
			},
			false,
			"",
		},
		{
			"missing dc_id",
			map[string]interface{}{
				"inbound_items": []interface{}{map[string]interface{}{"sku": "SKU001", "qty": 100}},
				"process_date": "2026-02-16",
			},
			true,
			"dc_id is required",
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

// TestInboundReceivingInput verifies inbound receiving step
func TestInboundReceivingInput(t *testing.T) {
	s := NewDistributionCenterSaga()
	stepDef := s.GetStepDefinition(2)

	if stepDef == nil {
		t.Error("step 2 (Inbound Receiving) not found")
	}
}

// TestCrossDockingInput verifies cross-docking step
func TestCrossDockingInput(t *testing.T) {
	s := NewDistributionCenterSaga()
	stepDef := s.GetStepDefinition(4)

	if stepDef == nil {
		t.Error("step 4 (Cross Docking) not found")
	}
}

// TestOutboundPrepInput verifies outbound prep step
func TestOutboundPrepInput(t *testing.T) {
	s := NewDistributionCenterSaga()
	stepDef := s.GetStepDefinition(7)

	if stepDef == nil {
		t.Error("step 7 (Outbound Prep) not found")
	}
}

// TestDistributionCenterCompensation verifies compensation steps
func TestDistributionCenterCompensation(t *testing.T) {
	s := NewDistributionCenterSaga()
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

// ========== ROUTE OPTIMIZATION SAGA TESTS (SAGA-SC06) ==========

// TestRouteOptimizationSagaType verifies Route Optimization saga returns correct type
func TestRouteOptimizationSagaType(t *testing.T) {
	s := NewRouteOptimizationSaga()
	if s.SagaType() != "SAGA-SC06" {
		t.Errorf("expected SAGA-SC06, got %s", s.SagaType())
	}
}

// TestRouteOptimizationStepCount verifies step count
func TestRouteOptimizationStepCount(t *testing.T) {
	s := NewRouteOptimizationSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 17 // 9 forward + 8 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestRouteOptimizationImplementsInterface verifies saga implements SagaHandler
func TestRouteOptimizationImplementsInterface(t *testing.T) {
	s := NewRouteOptimizationSaga()
	var _ saga.SagaHandler = s
}

// TestRouteOptimizationValidation verifies input validation
func TestRouteOptimizationValidation(t *testing.T) {
	s := NewRouteOptimizationSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid route optimization input",
			map[string]interface{}{
				"route_id":     "route-001",
				"stops":        []interface{}{map[string]interface{}{"location": "LOC001"}, map[string]interface{}{"location": "LOC002"}},
				"optimization_date": "2026-02-16",
			},
			false,
			"",
		},
		{
			"missing route_id",
			map[string]interface{}{
				"stops":        []interface{}{map[string]interface{}{"location": "LOC001"}},
				"optimization_date": "2026-02-16",
			},
			true,
			"route_id is required",
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

// TestRouteCalculationInput verifies route calculation step
func TestRouteCalculationInput(t *testing.T) {
	s := NewRouteOptimizationSaga()
	stepDef := s.GetStepDefinition(3)

	if stepDef == nil {
		t.Error("step 3 (Route Calculation) not found")
	}
}

// TestVehicleAssignmentInput verifies vehicle assignment step
func TestVehicleAssignmentInput(t *testing.T) {
	s := NewRouteOptimizationSaga()
	stepDef := s.GetStepDefinition(5)

	if stepDef == nil {
		t.Error("step 5 (Vehicle Assignment) not found")
	}
}

// TestDriverAllocationInput verifies driver allocation step
func TestDriverAllocationInput(t *testing.T) {
	s := NewRouteOptimizationSaga()
	stepDef := s.GetStepDefinition(6)

	if stepDef == nil {
		t.Error("step 6 (Driver Allocation) not found")
	}
}

// TestRouteOptimizationCompensation verifies compensation steps
func TestRouteOptimizationCompensation(t *testing.T) {
	s := NewRouteOptimizationSaga()
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

// ========== SUPPLY CHAIN VISIBILITY SAGA TESTS (SAGA-SC07) ==========

// TestSupplyChainVisibilitySagaType verifies Supply Chain Visibility saga returns correct type
func TestSupplyChainVisibilitySagaType(t *testing.T) {
	s := NewSupplyChainVisibilitySaga()
	if s.SagaType() != "SAGA-SC07" {
		t.Errorf("expected SAGA-SC07, got %s", s.SagaType())
	}
}

// TestSupplyChainVisibilityStepCount verifies step count
func TestSupplyChainVisibilityStepCount(t *testing.T) {
	s := NewSupplyChainVisibilitySaga()
	steps := s.GetStepDefinitions()
	expectedCount := 15 // 8 forward + 7 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestSupplyChainVisibilityImplementsInterface verifies saga implements SagaHandler
func TestSupplyChainVisibilityImplementsInterface(t *testing.T) {
	s := NewSupplyChainVisibilitySaga()
	var _ saga.SagaHandler = s
}

// TestSupplyChainVisibilityValidation verifies input validation
func TestSupplyChainVisibilityValidation(t *testing.T) {
	s := NewSupplyChainVisibilitySaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid visibility input",
			map[string]interface{}{
				"tracking_id":  "track-001",
				"shipment_id":  "ship-001",
				"event_type":   "DEPARTURE",
				"timestamp":    "2026-02-16T10:00:00Z",
			},
			false,
			"",
		},
		{
			"missing tracking_id",
			map[string]interface{}{
				"shipment_id": "ship-001",
				"event_type":  "DEPARTURE",
				"timestamp":   "2026-02-16T10:00:00Z",
			},
			true,
			"tracking_id is required",
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

// TestLocationTrackingInput verifies location tracking step
func TestLocationTrackingInput(t *testing.T) {
	s := NewSupplyChainVisibilitySaga()
	stepDef := s.GetStepDefinition(2)

	if stepDef == nil {
		t.Error("step 2 (Location Tracking) not found")
	}
}

// TestStatusUpdateInput verifies status update step
func TestStatusUpdateInput(t *testing.T) {
	s := NewSupplyChainVisibilitySaga()
	stepDef := s.GetStepDefinition(4)

	if stepDef == nil {
		t.Error("step 4 (Status Update) not found")
	}
}

// TestCustomerNotificationInput verifies customer notification step
func TestCustomerNotificationInput(t *testing.T) {
	s := NewSupplyChainVisibilitySaga()
	stepDef := s.GetStepDefinition(6)

	if stepDef == nil {
		t.Error("step 6 (Customer Notification) not found")
	}
}

// TestSupplyChainVisibilityCompensation verifies compensation steps
func TestSupplyChainVisibilityCompensation(t *testing.T) {
	s := NewSupplyChainVisibilitySaga()
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

// ========== SUPPLIER PERFORMANCE SAGA TESTS (SAGA-SC08) ==========

// TestSupplierPerformanceSagaType verifies Supplier Performance saga returns correct type
func TestSupplierPerformanceSagaType(t *testing.T) {
	s := NewSupplierPerformanceSaga()
	if s.SagaType() != "SAGA-SC08" {
		t.Errorf("expected SAGA-SC08, got %s", s.SagaType())
	}
}

// TestSupplierPerformanceStepCount verifies step count
func TestSupplierPerformanceStepCount(t *testing.T) {
	s := NewSupplierPerformanceSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 17 // 9 forward + 8 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestSupplierPerformanceImplementsInterface verifies saga implements SagaHandler
func TestSupplierPerformanceImplementsInterface(t *testing.T) {
	s := NewSupplierPerformanceSaga()
	var _ saga.SagaHandler = s
}

// TestSupplierPerformanceValidation verifies input validation
func TestSupplierPerformanceValidation(t *testing.T) {
	s := NewSupplierPerformanceSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid supplier performance input",
			map[string]interface{}{
				"performance_id": "perf-001",
				"supplier_id":    "supp-001",
				"period":         "Q1-2026",
				"metrics":        map[string]interface{}{"on_time": 95.5, "quality": 98.2},
			},
			false,
			"",
		},
		{
			"missing performance_id",
			map[string]interface{}{
				"supplier_id": "supp-001",
				"period":      "Q1-2026",
				"metrics":     map[string]interface{}{"on_time": 95.5},
			},
			true,
			"performance_id is required",
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

// TestPerformanceCalculationInput verifies performance calculation step
func TestPerformanceCalculationInput(t *testing.T) {
	s := NewSupplierPerformanceSaga()
	stepDef := s.GetStepDefinition(2)

	if stepDef == nil {
		t.Error("step 2 (Performance Calculation) not found")
	}
}

// TestSupplierScoringInput verifies supplier scoring step
func TestSupplierScoringInput(t *testing.T) {
	s := NewSupplierPerformanceSaga()
	stepDef := s.GetStepDefinition(4)

	if stepDef == nil {
		t.Error("step 4 (Supplier Scoring) not found")
	}
}

// TestSupplierRankingInput verifies supplier ranking step
func TestSupplierRankingInput(t *testing.T) {
	s := NewSupplierPerformanceSaga()
	stepDef := s.GetStepDefinition(6)

	if stepDef == nil {
		t.Error("step 6 (Supplier Ranking) not found")
	}
}

// TestSupplierPerformanceCompensation verifies compensation steps
func TestSupplierPerformanceCompensation(t *testing.T) {
	s := NewSupplierPerformanceSaga()
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

// ========== INTEGRATION TESTS ==========

// TestSupplyChainSagasInterface verifies all supply chain sagas implement SagaHandler
func TestSupplyChainSagasInterface(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewInboundLogisticsSaga(),
		NewWarehouseOpsSaga(),
		NewThirdPartyLogisticsCoordinationSaga(),
		NewOrderFulfillmentSaga(),
		NewDistributionCenterSaga(),
		NewRouteOptimizationSaga(),
		NewSupplyChainVisibilitySaga(),
		NewSupplierPerformanceSaga(),
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

// TestSupplyChainSagaTypes verifies all saga types return correct identifiers
func TestSupplyChainSagaTypes(t *testing.T) {
	sagas := []struct {
		name         string
		saga         saga.SagaHandler
		expectedType string
	}{
		{"Inbound Logistics", NewInboundLogisticsSaga(), "SAGA-SC01"},
		{"Warehouse Operations", NewWarehouseOpsSaga(), "SAGA-SC02"},
		{"3PL Coordination", NewThirdPartyLogisticsCoordinationSaga(), "SAGA-SC03"},
		{"Order Fulfillment", NewOrderFulfillmentSaga(), "SAGA-SC04"},
		{"Distribution Center", NewDistributionCenterSaga(), "SAGA-SC05"},
		{"Route Optimization", NewRouteOptimizationSaga(), "SAGA-SC06"},
		{"Supply Chain Visibility", NewSupplyChainVisibilitySaga(), "SAGA-SC07"},
		{"Supplier Performance", NewSupplierPerformanceSaga(), "SAGA-SC08"},
	}

	for _, tt := range sagas {
		t.Run(tt.name, func(t *testing.T) {
			if tt.saga.SagaType() != tt.expectedType {
				t.Errorf("expected %s, got %s", tt.expectedType, tt.saga.SagaType())
			}
		})
	}
}

// TestSupplyChainSagasGetStepDefinitions verifies step retrieval
func TestSupplyChainSagasGetStepDefinitions(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewInboundLogisticsSaga(),
		NewWarehouseOpsSaga(),
		NewThirdPartyLogisticsCoordinationSaga(),
		NewOrderFulfillmentSaga(),
		NewDistributionCenterSaga(),
		NewRouteOptimizationSaga(),
		NewSupplyChainVisibilitySaga(),
		NewSupplierPerformanceSaga(),
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

// TestSupplyChainSagasInvalidStepLookup verifies invalid step lookup returns nil
func TestSupplyChainSagasInvalidStepLookup(t *testing.T) {
	s := NewInboundLogisticsSaga()
	invalidStep := s.GetStepDefinition(999)
	if invalidStep != nil {
		t.Error("invalid step should return nil")
	}
}

// TestSupplyChainSagasNilInput verifies nil input handling
func TestSupplyChainSagasNilInput(t *testing.T) {
	s := NewInboundLogisticsSaga()
	err := s.ValidateInput(nil)
	if err == nil {
		t.Error("expected error for nil input")
	}
}

// TestSupplyChainSagasEmptyMapInput verifies empty map input handling
func TestSupplyChainSagasEmptyMapInput(t *testing.T) {
	s := NewWarehouseOpsSaga()
	err := s.ValidateInput(map[string]interface{}{})
	if err == nil {
		t.Error("expected error for empty map input")
	}
}

// TestSupplyChainSagasStringInput verifies string input rejection
func TestSupplyChainSagasStringInput(t *testing.T) {
	s := NewOrderFulfillmentSaga()
	err := s.ValidateInput("invalid string")
	if err == nil {
		t.Error("expected error for string input")
	}
}

// TestSupplyChainSagasIntInput verifies integer input rejection
func TestSupplyChainSagasIntInput(t *testing.T) {
	s := NewDistributionCenterSaga()
	err := s.ValidateInput(12345)
	if err == nil {
		t.Error("expected error for integer input")
	}
}

// TestSupplyChainSagasCriticalStepMarking verifies critical steps are marked
func TestSupplyChainSagasCriticalStepMarking(t *testing.T) {
	s := NewInboundLogisticsSaga()
	steps := s.GetStepDefinitions()

	hasCritical := false
	for _, step := range steps {
		if step.IsCritical {
			hasCritical = true
			break
		}
	}

	if !hasCritical {
		t.Error("no critical steps found in inbound logistics saga")
	}
}

// TestSupplyChainSagasRetryConfiguration verifies retry config exists
func TestSupplyChainSagasRetryConfiguration(t *testing.T) {
	s := NewOrderFulfillmentSaga()
	stepDef := s.GetStepDefinition(1)

	if stepDef.RetryConfig == nil {
		t.Error("retry configuration missing in step 1")
	}
}

// TestSupplyChainSagasTimeoutConfiguration verifies timeout config exists
func TestSupplyChainSagasTimeoutConfiguration(t *testing.T) {
	s := NewWarehouseOpsSaga()
	stepDef := s.GetStepDefinition(1)

	if stepDef.TimeoutSeconds == 0 {
		t.Error("timeout configuration missing in step 1")
	}
}
