// Package retail provides saga handlers for retail workflows
package retail

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// StockTransferSaga implements SAGA-R06: Stock Transfer Between Locations workflow
// Business Flow: InitiateTransfer → ValidateLocationInventory → ReserveInventorySource → CreateShipment → UpdateInTransitInventory → ReceiveAtDestination → UpdateDestinationInventory → GenerateTransferJournal → CompleteTransfer
// Steps: 9 forward + 10 compensation = 19 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,5,8,10
type StockTransferSaga struct {
	steps []*saga.StepDefinition
}

// NewStockTransferSaga creates a new Stock Transfer saga handler
func NewStockTransferSaga() saga.SagaHandler {
	return &StockTransferSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Transfer
			{
				StepNumber:    1,
				ServiceName:   "inventory",
				HandlerMethod: "InitiateTransfer",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"transferID":     "$.input.transfer_id",
					"fromLocation":   "$.input.from_location",
					"toLocation":     "$.input.to_location",
					"transferDate":   "$.input.transfer_date",
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
			// Step 2: Validate Location Inventory
			{
				StepNumber:    2,
				ServiceName:   "inventory",
				HandlerMethod: "ValidateLocationInventory",
				InputMapping: map[string]string{
					"transferID":   "$.steps.1.result.transfer_id",
					"fromLocation": "$.input.from_location",
					"quantity":     "$.input.quantity",
					"itemDetails":  "$.input.item_details",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Reserve Inventory at Source
			{
				StepNumber:    3,
				ServiceName:   "inventory",
				HandlerMethod: "ReserveInventorySource",
				InputMapping: map[string]string{
					"transferID":   "$.steps.1.result.transfer_id",
					"fromLocation": "$.input.from_location",
					"itemDetails":  "$.input.item_details",
					"quantity":     "$.input.quantity",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{101},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Create Shipment
			{
				StepNumber:    4,
				ServiceName:   "logistics",
				HandlerMethod: "CreateShipment",
				InputMapping: map[string]string{
					"transferID":   "$.steps.1.result.transfer_id",
					"fromLocation": "$.input.from_location",
					"toLocation":   "$.input.to_location",
					"itemDetails":  "$.input.item_details",
					"quantity":     "$.input.quantity",
				},
				TimeoutSeconds:    35,
				IsCritical:        true,
				CompensationSteps: []int32{102},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Update In-Transit Inventory
			{
				StepNumber:    5,
				ServiceName:   "inventory",
				HandlerMethod: "UpdateInTransitInventory",
				InputMapping: map[string]string{
					"transferID":   "$.steps.1.result.transfer_id",
					"shipmentID":   "$.steps.4.result.shipment_id",
					"fromLocation": "$.input.from_location",
					"quantity":     "$.input.quantity",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{103},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Receive at Destination
			{
				StepNumber:    6,
				ServiceName:   "warehouse",
				HandlerMethod: "ReceiveAtDestination",
				InputMapping: map[string]string{
					"transferID":   "$.steps.1.result.transfer_id",
					"shipmentID":   "$.steps.4.result.shipment_id",
					"toLocation":   "$.input.to_location",
					"itemDetails":  "$.input.item_details",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Update Destination Inventory
			{
				StepNumber:    7,
				ServiceName:   "inventory",
				HandlerMethod: "UpdateDestinationInventory",
				InputMapping: map[string]string{
					"transferID":     "$.steps.1.result.transfer_id",
					"toLocation":     "$.input.to_location",
					"receivedItems":  "$.steps.6.result.received_items",
					"quantity":       "$.input.quantity",
				},
				TimeoutSeconds:    25,
				IsCritical:        false,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Generate Transfer Journal
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "GenerateTransferJournal",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"transferID":   "$.steps.1.result.transfer_id",
					"fromLocation": "$.input.from_location",
					"toLocation":   "$.input.to_location",
					"quantity":     "$.input.quantity",
					"journalDate":  "$.input.transfer_date",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Complete Transfer
			{
				StepNumber:    9,
				ServiceName:   "inventory",
				HandlerMethod: "CompleteTransfer",
				InputMapping: map[string]string{
					"transferID":      "$.steps.1.result.transfer_id",
					"journalEntries":  "$.steps.8.result.journal_entries",
					"transferStatus":  "Completed",
				},
				TimeoutSeconds:    20,
				IsCritical:        false,
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

			// Step 101: Revert Source Inventory Reservation (compensates step 3)
			{
				StepNumber:    101,
				ServiceName:   "inventory",
				HandlerMethod: "RevertSourceInventoryReservation",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Revert Shipment Creation (compensates step 4)
			{
				StepNumber:    102,
				ServiceName:   "logistics",
				HandlerMethod: "RevertShipmentCreation",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 103: Revert In-Transit Inventory Update (compensates step 5)
			{
				StepNumber:    103,
				ServiceName:   "inventory",
				HandlerMethod: "RevertInTransitInventoryUpdate",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 104: Revert Destination Reception (compensates step 6)
			{
				StepNumber:    104,
				ServiceName:   "warehouse",
				HandlerMethod: "RevertDestinationReception",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: Revert Destination Inventory Update (compensates step 7)
			{
				StepNumber:    105,
				ServiceName:   "inventory",
				HandlerMethod: "RevertDestinationInventoryUpdate",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Reverse Transfer Journal (compensates step 8)
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseTransferJournal",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Revert Initiate Transfer (compensates step 1)
			{
				StepNumber:    107,
				ServiceName:   "inventory",
				HandlerMethod: "RevertInitiateTransfer",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 108: Revert Validate Location Inventory (compensates step 2)
			{
				StepNumber:    108,
				ServiceName:   "inventory",
				HandlerMethod: "RevertValidateLocationInventory",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 109: Revert Complete Transfer (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "inventory",
				HandlerMethod: "RevertCompleteTransfer",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 110: Revert Update Destination Inventory (additional compensation)
			{
				StepNumber:    110,
				ServiceName:   "inventory",
				HandlerMethod: "RevertUpdateDestinationInventory",
				InputMapping: map[string]string{
					"transferID": "$.steps.1.result.transfer_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *StockTransferSaga) SagaType() string {
	return "SAGA-R06"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *StockTransferSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *StockTransferSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *StockTransferSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["transfer_id"] == nil {
		return errors.New("transfer_id is required")
	}

	if inputMap["from_location"] == nil {
		return errors.New("from_location is required")
	}

	if inputMap["to_location"] == nil {
		return errors.New("to_location is required")
	}

	if inputMap["quantity"] == nil {
		return errors.New("quantity is required")
	}

	return nil
}
