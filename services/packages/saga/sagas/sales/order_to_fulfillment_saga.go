// Package sales provides saga handlers for sales module workflows
package sales

import (
	"errors"
	"p9e.in/samavaya/packages/saga"
)

// OrderToFulfillmentSaga implements the Order-to-Fulfillment workflow
// Business flow: Mark for Fulfillment → Allocate Stock → Create Pick List → Confirm Picking →
// Create Package → Generate E-Way Bill → Create Shipment → Post Goods Issue → Send Tracking
type OrderToFulfillmentSaga struct {
	steps []*saga.StepDefinition
}

// NewOrderToFulfillmentSaga creates a new Order-to-Fulfillment saga handler
func NewOrderToFulfillmentSaga() saga.SagaHandler {
	return &OrderToFulfillmentSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Mark Order for Fulfillment
			{
				StepNumber:    1,
				ServiceName:   "sales-order",
				HandlerMethod: "MarkForFulfillment",
				InputMapping: map[string]string{
					"orderID":  "$.input.order_id",
					"tenantID": "$.tenantID",
				},
				TimeoutSeconds: 20,
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

			// Step 2: Allocate Stock (inventory-core)
			{
				StepNumber:    2,
				ServiceName:   "inventory-core",
				HandlerMethod: "AllocateStock",
				InputMapping: map[string]string{
					"orderID":  "$.input.order_id",
					"tenantID": "$.tenantID",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{102},
			},

			// Step 3: Create Pick List (wms)
			{
				StepNumber:    3,
				ServiceName:   "wms",
				HandlerMethod: "CreatePickList",
				InputMapping: map[string]string{
					"orderID":  "$.input.order_id",
					"tenantID": "$.tenantID",
				},
				TimeoutSeconds: 20,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{103},
			},

			// Step 4: Confirm Picking (wms)
			{
				StepNumber:    4,
				ServiceName:   "wms",
				HandlerMethod: "ConfirmPicking",
				InputMapping: map[string]string{
					"pickListID": "$.steps.3.result.pick_list_id",
					"tenantID":   "$.tenantID",
				},
				TimeoutSeconds: 20,
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

			// Step 5: Create Package (fulfillment)
			{
				StepNumber:    5,
				ServiceName:   "fulfillment",
				HandlerMethod: "CreatePackage",
				InputMapping: map[string]string{
					"pickListID": "$.steps.3.result.pick_list_id",
					"orderID":    "$.input.order_id",
					"tenantID":   "$.tenantID",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{105},
			},

			// Step 6: Generate E-Way Bill (e-way-bill)
			{
				StepNumber:    6,
				ServiceName:   "e-way-bill",
				HandlerMethod: "GenerateEWB",
				InputMapping: map[string]string{
					"shipmentID": "$.steps.5.result.shipment_id",
					"tenantID":   "$.tenantID",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        5,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      120000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{106},
			},

			// Step 7: Create Shipment (shipping)
			{
				StepNumber:    7,
				ServiceName:   "shipping",
				HandlerMethod: "CreateShipment",
				InputMapping: map[string]string{
					"shipmentID": "$.steps.5.result.shipment_id",
					"ewbNumber":  "$.steps.6.result.ewb_number",
					"tenantID":   "$.tenantID",
				},
				TimeoutSeconds: 40,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        5,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      60000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{107},
			},

			// Step 8: Post Goods Issue (inventory-core)
			{
				StepNumber:    8,
				ServiceName:   "inventory-core",
				HandlerMethod: "PostGoodsIssue",
				InputMapping: map[string]string{
					"shipmentID": "$.steps.5.result.shipment_id",
					"tenantID":   "$.tenantID",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{108},
			},

			// Step 9: Send Tracking Info (notification)
			{
				StepNumber:    9,
				ServiceName:   "notification",
				HandlerMethod: "SendTrackingInfo",
				InputMapping: map[string]string{
					"shipmentID": "$.steps.5.result.shipment_id",
					"trackingNo": "$.steps.7.result.tracking_number",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},

			// Compensation steps
			{
				StepNumber:    102,
				ServiceName:   "inventory-core",
				HandlerMethod: "ReleaseAllocation",
				InputMapping: map[string]string{
					"orderID": "$.input.order_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},

			{
				StepNumber:    103,
				ServiceName:   "wms",
				HandlerMethod: "CancelPickList",
				InputMapping: map[string]string{
					"pickListID": "$.steps.3.result.pick_list_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},

			{
				StepNumber:    105,
				ServiceName:   "fulfillment",
				HandlerMethod: "CancelPackage",
				InputMapping: map[string]string{
					"shipmentID": "$.steps.5.result.shipment_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},

			{
				StepNumber:    106,
				ServiceName:   "e-way-bill",
				HandlerMethod: "MarkPendingCancellation",
				InputMapping: map[string]string{
					"ewbNumber": "$.steps.6.result.ewb_number",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},

			{
				StepNumber:    107,
				ServiceName:   "shipping",
				HandlerMethod: "CancelShipment",
				InputMapping: map[string]string{
					"shipmentID":  "$.steps.5.result.shipment_id",
					"trackingNo":  "$.steps.7.result.tracking_number",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      60000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},

			{
				StepNumber:    108,
				ServiceName:   "inventory-core",
				HandlerMethod: "ReverseGoodsIssue",
				InputMapping: map[string]string{
					"shipmentID": "$.steps.5.result.shipment_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *OrderToFulfillmentSaga) SagaType() string {
	return "SAGA-S03"
}

// GetStepDefinitions returns all steps
func (s *OrderToFulfillmentSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns definition for a specific step
func (s *OrderToFulfillmentSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the input for saga execution
func (s *OrderToFulfillmentSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type: expected map[string]interface{}")
	}

	if inputMap["order_id"] == nil || inputMap["order_id"] == "" {
		return errors.New("order_id is required for Order-to-Fulfillment saga")
	}

	return nil
}
