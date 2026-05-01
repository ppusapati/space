// Package inventory provides saga handlers for inventory module workflows
package inventory

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// InterWarehouseTransferSaga implements SAGA-I01: Inter-Warehouse Transfer workflow
// Business Flow: Create Transfer → Issue from Source → Create In-Transit → Post Source GL → Ship → Receive at Dest → Post Dest GL → Update Status → Complete
// Special: In-transit inventory is tracked separately (not counted in either warehouse)
type InterWarehouseTransferSaga struct {
	steps []*saga.StepDefinition
}

// NewInterWarehouseTransferSaga creates a new Inter-Warehouse Transfer saga handler
func NewInterWarehouseTransferSaga() saga.SagaHandler {
	return &InterWarehouseTransferSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Transfer Order
			{
				StepNumber:    1,
				ServiceName:   "stock-transfer",
				HandlerMethod: "CreateTransferOrder",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"sourceWarehouseID": "$.input.source_warehouse_id",
					"destWarehouseID": "$.input.dest_warehouse_id",
					"items":           "$.input.items",
					"transferReason":  "$.input.transfer_reason",
				},
				TimeoutSeconds: 25,
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
			// Step 2: Issue Stock from Source Warehouse
			{
				StepNumber:    2,
				ServiceName:   "inventory-core",
				HandlerMethod: "IssueStockFromSource",
				InputMapping: map[string]string{
					"transferID":      "$.steps.1.result.transfer_id",
					"sourceWarehouseID": "$.input.source_warehouse_id",
					"items":           "$.input.items",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{102},
			},
			// Step 3: Create In-Transit Record
			{
				StepNumber:    3,
				ServiceName:   "stock-transfer",
				HandlerMethod: "CreateInTransitRecord",
				InputMapping: map[string]string{
					"transferID":      "$.steps.1.result.transfer_id",
					"sourceWarehouseID": "$.input.source_warehouse_id",
					"destWarehouseID": "$.input.dest_warehouse_id",
					"items":           "$.input.items",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{103},
			},
			// Step 4: Post Source Warehouse GL
			{
				StepNumber:    4,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostSourceWarehouseGL",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"transferID":      "$.steps.1.result.transfer_id",
					"items":           "$.input.items",
					"sourceWarehouseID": "$.input.source_warehouse_id",
					"journalDate":     "$.input.journal_date",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{104},
			},
			// Step 5: Create Internal Shipment
			{
				StepNumber:    5,
				ServiceName:   "shipping",
				HandlerMethod: "CreateInternalShipment",
				InputMapping: map[string]string{
					"transferID":      "$.steps.1.result.transfer_id",
					"sourceWarehouseID": "$.input.source_warehouse_id",
					"destWarehouseID": "$.input.dest_warehouse_id",
					"items":           "$.input.items",
					"expectedDeliveryDate": "$.input.expected_delivery_date",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{105},
			},
			// Step 6: Receive Stock at Destination Warehouse
			{
				StepNumber:    6,
				ServiceName:   "inventory-core",
				HandlerMethod: "ReceiveStockAtDestination",
				InputMapping: map[string]string{
					"transferID":      "$.steps.1.result.transfer_id",
					"destWarehouseID": "$.input.dest_warehouse_id",
					"items":           "$.input.items",
					"receivedItems":   "$.input.received_items",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{106},
			},
			// Step 7: Post Destination Warehouse GL
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostDestinationWarehouseGL",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"transferID":      "$.steps.1.result.transfer_id",
					"items":           "$.input.items",
					"destWarehouseID": "$.input.dest_warehouse_id",
					"journalDate":     "$.input.journal_date",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{107},
			},
			// Step 8: Update Transfer Status
			{
				StepNumber:    8,
				ServiceName:   "stock-transfer",
				HandlerMethod: "UpdateTransferStatus",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
					"newStatus":  "COMPLETED",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{108},
			},
			// Step 9: Complete Transfer (finalization)
			{
				StepNumber:    9,
				ServiceName:   "stock-transfer",
				HandlerMethod: "CompleteTransfer",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{},
			},
			// ===== COMPENSATION STEPS =====

			// Step 101: Cancel Transfer Order (compensates step 1)
			{
				StepNumber:    101,
				ServiceName:   "stock-transfer",
				HandlerMethod: "CancelTransferOrder",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 102: Restore Source Stock (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "inventory-core",
				HandlerMethod: "RestoreSourceStock",
				InputMapping: map[string]string{
					"transferID":      "$.steps.1.result.transfer_id",
					"sourceWarehouseID": "$.input.source_warehouse_id",
					"items":           "$.input.items",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: Delete In-Transit Record (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "stock-transfer",
				HandlerMethod: "DeleteInTransitRecord",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 104: Reverse Source GL (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseSourceGL",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Cancel Shipment (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "shipping",
				HandlerMethod: "CancelShipment",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Reverse Destination Receipt (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "inventory-core",
				HandlerMethod: "ReverseDestinationReceipt",
				InputMapping: map[string]string{
					"transferID":      "$.steps.1.result.transfer_id",
					"destWarehouseID": "$.input.dest_warehouse_id",
					"items":           "$.input.items",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 107: Reverse Destination GL (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseDestinationGL",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 108: Revert Transfer Status (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "stock-transfer",
				HandlerMethod: "RevertTransferStatus",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *InterWarehouseTransferSaga) SagaType() string {
	return "SAGA-I01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *InterWarehouseTransferSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *InterWarehouseTransferSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *InterWarehouseTransferSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["source_warehouse_id"] == nil {
		return errors.New("source_warehouse_id is required")
	}

	if inputMap["dest_warehouse_id"] == nil {
		return errors.New("dest_warehouse_id is required")
	}

	if inputMap["items"] == nil {
		return errors.New("items are required")
	}

	items, ok := inputMap["items"].([]interface{})
	if !ok || len(items) == 0 {
		return errors.New("items must be a non-empty list")
	}

	if inputMap["transfer_reason"] == nil {
		return errors.New("transfer_reason is required")
	}

	return nil
}
