// Package supplychain provides saga handlers for supply chain module workflows
package supplychain

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// OrderFulfillmentOutboundLogisticsSaga implements SAGA-SC04: Order Fulfillment & Outbound Logistics
// Business Flow: InitiateFulfillmentProcess → PickOrderItems → PackShipment → GenerateShippingLabel → UpdateInventoryForShipment → CreateShipmentRecord → SelectShippingCarrier → ArrangePickup → TrackShipmentStatus → PostFulfillmentJournals → CompleteFulfillment
// Timeout: 180 seconds, Critical steps: 1,2,3,4,6,8,11
type OrderFulfillmentOutboundLogisticsSaga struct {
	steps []*saga.StepDefinition
}

// NewOrderFulfillmentOutboundLogisticsSaga creates a new Order Fulfillment & Outbound Logistics saga handler
func NewOrderFulfillmentOutboundLogisticsSaga() saga.SagaHandler {
	return &OrderFulfillmentOutboundLogisticsSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Fulfillment Process
			{
				StepNumber:    1,
				ServiceName:   "order-fulfillment",
				HandlerMethod: "InitiateFulfillmentProcess",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"orderID":        "$.input.order_id",
					"fulfillmentID":  "$.input.fulfillment_id",
					"customerID":     "$.input.customer_id",
					"shippingDate":   "$.input.shipping_date",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},
			// Step 2: Pick Order Items
			{
				StepNumber:    2,
				ServiceName:   "warehouse",
				HandlerMethod: "PickOrderItems",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"orderID":               "$.input.order_id",
					"fulfillmentID":         "$.input.fulfillment_id",
					"fulfillmentProcessData": "$.steps.1.result.fulfillment_process_data",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{102},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Pack Shipment
			{
				StepNumber:    3,
				ServiceName:   "warehouse",
				HandlerMethod: "PackShipment",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"fulfillmentID":   "$.input.fulfillment_id",
					"orderID":         "$.input.order_id",
					"pickedItems":     "$.steps.2.result.picked_items",
					"fulfillmentProcessData": "$.steps.1.result.fulfillment_process_data",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{103},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Generate Shipping Label
			{
				StepNumber:    4,
				ServiceName:   "shipment",
				HandlerMethod: "GenerateShippingLabel",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"fulfillmentID":   "$.input.fulfillment_id",
					"orderID":         "$.input.order_id",
					"customerID":      "$.input.customer_id",
					"packedShipment":  "$.steps.3.result.packed_shipment",
				},
				TimeoutSeconds: 60,
				IsCritical:     true,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Update Inventory for Shipment
			{
				StepNumber:    5,
				ServiceName:   "inventory",
				HandlerMethod: "UpdateInventoryForShipment",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"orderID":       "$.input.order_id",
					"fulfillmentID": "$.input.fulfillment_id",
					"pickedItems":   "$.steps.2.result.picked_items",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Create Shipment Record
			{
				StepNumber:    6,
				ServiceName:   "shipment",
				HandlerMethod: "CreateShipmentRecord",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"fulfillmentID":  "$.input.fulfillment_id",
					"orderID":        "$.input.order_id",
					"shippingLabel":  "$.steps.4.result.shipping_label",
					"packedShipment": "$.steps.3.result.packed_shipment",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Select Shipping Carrier
			{
				StepNumber:    7,
				ServiceName:   "logistics",
				HandlerMethod: "SelectShippingCarrier",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"fulfillmentID":  "$.input.fulfillment_id",
					"shippingDate":   "$.input.shipping_date",
					"shipmentRecord": "$.steps.6.result.shipment_record",
					"shippingLabel":  "$.steps.4.result.shipping_label",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Arrange Pickup
			{
				StepNumber:    8,
				ServiceName:   "logistics",
				HandlerMethod: "ArrangePickup",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"fulfillmentID":   "$.input.fulfillment_id",
					"shippingDate":    "$.input.shipping_date",
					"shipmentRecord":  "$.steps.6.result.shipment_record",
					"carrierSelection": "$.steps.7.result.carrier_selection",
				},
				TimeoutSeconds: 60,
				IsCritical:     true,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Track Shipment Status
			{
				StepNumber:    9,
				ServiceName:   "shipment",
				HandlerMethod: "TrackShipmentStatus",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"fulfillmentID":  "$.input.fulfillment_id",
					"shipmentRecord": "$.steps.6.result.shipment_record",
					"pickupArrangement": "$.steps.8.result.pickup_arrangement",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Post Fulfillment Journals
			{
				StepNumber:    10,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostFulfillmentJournals",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"fulfillmentID":   "$.input.fulfillment_id",
					"orderID":         "$.input.order_id",
					"shippingDate":    "$.input.shipping_date",
					"shipmentRecord":  "$.steps.6.result.shipment_record",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{110},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 11: Complete Fulfillment
			{
				StepNumber:    11,
				ServiceName:   "order-fulfillment",
				HandlerMethod: "CompleteFulfillment",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"fulfillmentID":   "$.input.fulfillment_id",
					"orderID":         "$.input.order_id",
					"customerID":      "$.input.customer_id",
					"shipmentRecord":  "$.steps.6.result.shipment_record",
					"trackingStatus":  "$.steps.9.result.tracking_status",
					"journalEntries":  "$.steps.10.result.journal_entries",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				CompensationSteps: []int32{},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// ===== COMPENSATION STEPS =====

			// Step 102: CancelPickedItems (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "warehouse",
				HandlerMethod: "CancelPickedItems",
				InputMapping: map[string]string{
					"fulfillmentID": "$.input.fulfillment_id",
					"pickedItems":   "$.steps.2.result.picked_items",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 103: UnpackShipment (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "warehouse",
				HandlerMethod: "UnpackShipment",
				InputMapping: map[string]string{
					"fulfillmentID":  "$.input.fulfillment_id",
					"packedShipment": "$.steps.3.result.packed_shipment",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 104: VoidShippingLabel (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "shipment",
				HandlerMethod: "VoidShippingLabel",
				InputMapping: map[string]string{
					"fulfillmentID":  "$.input.fulfillment_id",
					"shippingLabel":  "$.steps.4.result.shipping_label",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 105: ReverseInventoryShipmentUpdate (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "inventory",
				HandlerMethod: "ReverseInventoryShipmentUpdate",
				InputMapping: map[string]string{
					"fulfillmentID": "$.input.fulfillment_id",
					"orderID":       "$.input.order_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 106: VoidShipmentRecord (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "shipment",
				HandlerMethod: "VoidShipmentRecord",
				InputMapping: map[string]string{
					"fulfillmentID":  "$.input.fulfillment_id",
					"shipmentRecord": "$.steps.6.result.shipment_record",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 107: ReverseCarrierSelection (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "logistics",
				HandlerMethod: "ReverseCarrierSelection",
				InputMapping: map[string]string{
					"fulfillmentID":   "$.input.fulfillment_id",
					"carrierSelection": "$.steps.7.result.carrier_selection",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 108: CancelPickupArrangement (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "logistics",
				HandlerMethod: "CancelPickupArrangement",
				InputMapping: map[string]string{
					"fulfillmentID":      "$.input.fulfillment_id",
					"pickupArrangement":  "$.steps.8.result.pickup_arrangement",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 109: StopShipmentTracking (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "shipment",
				HandlerMethod: "StopShipmentTracking",
				InputMapping: map[string]string{
					"fulfillmentID":  "$.input.fulfillment_id",
					"trackingStatus": "$.steps.9.result.tracking_status",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 110: ReverseFulfillmentJournals (compensates step 10)
			{
				StepNumber:    110,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseFulfillmentJournals",
				InputMapping: map[string]string{
					"fulfillmentID":  "$.input.fulfillment_id",
					"journalEntries": "$.steps.10.result.journal_entries",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 111: CancelFulfillmentCompletion (compensates step 11)
			{
				StepNumber:    111,
				ServiceName:   "order-fulfillment",
				HandlerMethod: "CancelFulfillmentCompletion",
				InputMapping: map[string]string{
					"fulfillmentID": "$.input.fulfillment_id",
					"orderID":       "$.input.order_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *OrderFulfillmentOutboundLogisticsSaga) SagaType() string {
	return "SAGA-SC04"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *OrderFulfillmentOutboundLogisticsSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *OrderFulfillmentOutboundLogisticsSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *OrderFulfillmentOutboundLogisticsSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["order_id"] == nil {
		return errors.New("order_id is required")
	}

	orderID, ok := inputMap["order_id"].(string)
	if !ok || orderID == "" {
		return errors.New("order_id must be a non-empty string")
	}

	if inputMap["fulfillment_id"] == nil {
		return errors.New("fulfillment_id is required")
	}

	fulfillmentID, ok := inputMap["fulfillment_id"].(string)
	if !ok || fulfillmentID == "" {
		return errors.New("fulfillment_id must be a non-empty string")
	}

	if inputMap["customer_id"] == nil {
		return errors.New("customer_id is required")
	}

	customerID, ok := inputMap["customer_id"].(string)
	if !ok || customerID == "" {
		return errors.New("customer_id must be a non-empty string")
	}

	if inputMap["shipping_date"] == nil {
		return errors.New("shipping_date is required")
	}

	shippingDate, ok := inputMap["shipping_date"].(string)
	if !ok || shippingDate == "" {
		return errors.New("shipping_date must be a non-empty string")
	}

	return nil
}
