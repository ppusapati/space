// Package supplychain provides saga handlers for supply chain module workflows
package supplychain

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ReverseLogisticsReturnsManagementSaga implements SAGA-SC09: Reverse Logistics & Returns Management
// Business Flow: InitiateReturnProcess → CreateReturnShipment → PickReturnItems → PackReturnShipment → GenerateReturnLabel → UpdateInventoryForReturn → ReconcileReturnCosts → ProcessReturnCredit → PostReturnJournals → CompleteReturnProcess
// Timeout: 180 seconds, Critical steps: 1,2,3,4,6,8,10
type ReverseLogisticsReturnsManagementSaga struct {
	steps []*saga.StepDefinition
}

// NewReverseLogisticsReturnsManagementSaga creates a new Reverse Logistics & Returns Management saga handler
func NewReverseLogisticsReturnsManagementSaga() saga.SagaHandler {
	return &ReverseLogisticsReturnsManagementSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Return Process
			{
				StepNumber:    1,
				ServiceName:   "reverse-logistics",
				HandlerMethod: "InitiateReturnProcess",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"reverseID":        "$.input.reverse_id",
					"returnShipmentID": "$.input.return_shipment_id",
					"originDate":       "$.input.origin_date",
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
			// Step 2: Create Return Shipment
			{
				StepNumber:    2,
				ServiceName:   "returns",
				HandlerMethod: "CreateReturnShipment",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"reverseID":        "$.input.reverse_id",
					"returnShipmentID": "$.input.return_shipment_id",
					"returnProcessData": "$.steps.1.result.return_process_data",
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
			// Step 3: Pick Return Items
			{
				StepNumber:    3,
				ServiceName:   "warehouse",
				HandlerMethod: "PickReturnItems",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"reverseID":           "$.input.reverse_id",
					"returnShipmentData":  "$.steps.2.result.return_shipment_data",
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
			// Step 4: Pack Return Shipment
			{
				StepNumber:    4,
				ServiceName:   "warehouse",
				HandlerMethod: "PackReturnShipment",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"reverseID":         "$.input.reverse_id",
					"pickedReturnItems": "$.steps.3.result.picked_return_items",
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
			// Step 5: Generate Return Label
			{
				StepNumber:    5,
				ServiceName:   "shipment",
				HandlerMethod: "GenerateReturnLabel",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"reverseID":        "$.input.reverse_id",
					"packedShipment":   "$.steps.4.result.packed_shipment",
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
			// Step 6: Update Inventory for Return
			{
				StepNumber:    6,
				ServiceName:   "inventory",
				HandlerMethod: "UpdateInventoryForReturn",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"reverseID":         "$.input.reverse_id",
					"pickedReturnItems": "$.steps.3.result.picked_return_items",
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
			// Step 7: Reconcile Return Costs
			{
				StepNumber:    7,
				ServiceName:   "returns",
				HandlerMethod: "ReconcileReturnCosts",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"reverseID":       "$.input.reverse_id",
					"returnLabel":     "$.steps.5.result.return_label",
					"packedShipment":  "$.steps.4.result.packed_shipment",
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
			// Step 8: Process Return Credit
			{
				StepNumber:    8,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ProcessReturnCredit",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"reverseID":     "$.input.reverse_id",
					"returnCosts":   "$.steps.7.result.return_costs",
					"originDate":    "$.input.origin_date",
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
			// Step 9: Post Return Journals
			{
				StepNumber:    9,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostReturnJournals",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"reverseID":      "$.input.reverse_id",
					"originDate":     "$.input.origin_date",
					"returnCredit":   "$.steps.8.result.return_credit",
					"returnCosts":    "$.steps.7.result.return_costs",
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
			// Step 10: Complete Return Process
			{
				StepNumber:    10,
				ServiceName:   "reverse-logistics",
				HandlerMethod: "CompleteReturnProcess",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"reverseID":        "$.input.reverse_id",
					"returnShipmentID": "$.input.return_shipment_id",
					"returnCredit":     "$.steps.8.result.return_credit",
					"journalEntries":   "$.steps.9.result.journal_entries",
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

			// Step 102: CancelReturnShipment (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "returns",
				HandlerMethod: "CancelReturnShipment",
				InputMapping: map[string]string{
					"reverseID":           "$.input.reverse_id",
					"returnShipmentData":  "$.steps.2.result.return_shipment_data",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 103: ReversePickedItems (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "warehouse",
				HandlerMethod: "ReversePickedItems",
				InputMapping: map[string]string{
					"reverseID":          "$.input.reverse_id",
					"pickedReturnItems":  "$.steps.3.result.picked_return_items",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 104: UnpackReturnShipment (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "warehouse",
				HandlerMethod: "UnpackReturnShipment",
				InputMapping: map[string]string{
					"reverseID":       "$.input.reverse_id",
					"packedShipment":  "$.steps.4.result.packed_shipment",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 105: VoidReturnLabel (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "shipment",
				HandlerMethod: "VoidReturnLabel",
				InputMapping: map[string]string{
					"reverseID":    "$.input.reverse_id",
					"returnLabel":  "$.steps.5.result.return_label",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 106: ReverseInventoryReturnUpdate (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "inventory",
				HandlerMethod: "ReverseInventoryReturnUpdate",
				InputMapping: map[string]string{
					"reverseID": "$.input.reverse_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 107: ReverseReturnCostReconciliation (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "returns",
				HandlerMethod: "ReverseReturnCostReconciliation",
				InputMapping: map[string]string{
					"reverseID":    "$.input.reverse_id",
					"returnCosts":  "$.steps.7.result.return_costs",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 108: ReverseReturnCredit (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ReverseReturnCredit",
				InputMapping: map[string]string{
					"reverseID":    "$.input.reverse_id",
					"returnCredit": "$.steps.8.result.return_credit",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 109: ReverseReturnJournals (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseReturnJournals",
				InputMapping: map[string]string{
					"reverseID":      "$.input.reverse_id",
					"journalEntries": "$.steps.9.result.journal_entries",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 110: CancelReturnProcessCompletion (compensates step 10)
			{
				StepNumber:    110,
				ServiceName:   "reverse-logistics",
				HandlerMethod: "CancelReturnProcessCompletion",
				InputMapping: map[string]string{
					"reverseID":        "$.input.reverse_id",
					"returnShipmentID": "$.input.return_shipment_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ReverseLogisticsReturnsManagementSaga) SagaType() string {
	return "SAGA-SC09"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ReverseLogisticsReturnsManagementSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ReverseLogisticsReturnsManagementSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ReverseLogisticsReturnsManagementSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["reverse_id"] == nil {
		return errors.New("reverse_id is required")
	}

	reverseID, ok := inputMap["reverse_id"].(string)
	if !ok || reverseID == "" {
		return errors.New("reverse_id must be a non-empty string")
	}

	if inputMap["return_shipment_id"] == nil {
		return errors.New("return_shipment_id is required")
	}

	returnShipmentID, ok := inputMap["return_shipment_id"].(string)
	if !ok || returnShipmentID == "" {
		return errors.New("return_shipment_id must be a non-empty string")
	}

	if inputMap["origin_date"] == nil {
		return errors.New("origin_date is required")
	}

	originDate, ok := inputMap["origin_date"].(string)
	if !ok || originDate == "" {
		return errors.New("origin_date must be a non-empty string")
	}

	return nil
}
