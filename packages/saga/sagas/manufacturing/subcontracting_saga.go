// Package manufacturing provides saga handlers for manufacturing module workflows
package manufacturing

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// SubcontractingSaga implements SAGA-M02: Subcontracting workflow
// Business Flow: Create subcontract order → Raise subcontract PO → Issue materials to subcontractor →
// Send materials for process → Receive partially completed goods → Inspect subcontracted goods →
// Update quality status → Receive completed goods → Record subcontractor invoice → Complete subcontracting order
//
// Compensation: If any critical step fails, automatically reverses previous steps
// in reverse order to maintain data consistency and recover materials/funds
type SubcontractingSaga struct {
	steps []*saga.StepDefinition
}

// NewSubcontractingSaga creates a new Subcontracting saga handler
func NewSubcontractingSaga() saga.SagaHandler {
	return &SubcontractingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Subcontract Order (subcontracting service)
			// Creates a subcontracting order for outsourced manufacturing
			{
				StepNumber:    1,
				ServiceName:   "subcontracting",
				HandlerMethod: "CreateSubcontractOrder",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"poID":            "$.input.po_id",
					"subcontractorID": "$.input.subcontractor_id",
					"productID":       "$.input.product_id",
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
				CompensationSteps: []int32{},
			},
			// Step 2: Raise Subcontract PO (procurement service)
			// Raises a purchase order for the subcontracting work
			{
				StepNumber:    2,
				ServiceName:   "procurement",
				HandlerMethod: "RaiseSubcontractPO",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
					"subcontractorID":    "$.input.subcontractor_id",
					"productID":          "$.input.product_id",
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
				CompensationSteps: []int32{101},
			},
			// Step 3: Issue Materials to Subcontractor (inventory-core service)
			// Issues raw materials/components to subcontractor for processing
			{
				StepNumber:    3,
				ServiceName:   "inventory-core",
				HandlerMethod: "IssueMaterialsToSubcontractor",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
					"materialIssueDateDate": "$.input.material_issue_date",
					"productID":          "$.input.product_id",
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
				CompensationSteps: []int32{102},
			},
			// Step 4: Send Materials for Process (subcontracting service)
			// Sends materials to subcontractor for processing
			{
				StepNumber:    4,
				ServiceName:   "subcontracting",
				HandlerMethod: "SendMaterialsForProcess",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
					"subcontractorID":    "$.input.subcontractor_id",
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
				CompensationSteps: []int32{103},
			},
			// Step 5: Receive Partially Completed Goods (subcontracting service)
			// Receives partially completed goods from subcontractor
			{
				StepNumber:    5,
				ServiceName:   "subcontracting",
				HandlerMethod: "ReceivePartiallyCompletedGoods",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
					"productID":          "$.input.product_id",
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
				CompensationSteps: []int32{104},
			},
			// Step 6: Inspect Subcontracted Goods (quality-production service)
			// Performs quality inspection on subcontracted items
			{
				StepNumber:    6,
				ServiceName:   "quality-production",
				HandlerMethod: "InspectSubcontractedGoods",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
					"productID":          "$.input.product_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{105},
			},
			// Step 7: Update Quality Status (subcontracting service)
			// Updates quality status after inspection
			{
				StepNumber:    7,
				ServiceName:   "subcontracting",
				HandlerMethod: "UpdateQualityStatus",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{106},
			},
			// Step 8: Receive Completed Goods (inventory-core service)
			// Receives and warehouses completed goods from subcontractor
			{
				StepNumber:    8,
				ServiceName:   "inventory-core",
				HandlerMethod: "ReceiveCompletedGoods",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
					"productID":          "$.input.product_id",
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
				CompensationSteps: []int32{107},
			},
			// Step 9: Record Subcontractor Invoice (accounts-payable service)
			// Records subcontractor invoice in accounts payable
			{
				StepNumber:    9,
				ServiceName:   "accounts-payable",
				HandlerMethod: "RecordSubcontractorInvoice",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
					"subcontractorID":    "$.input.subcontractor_id",
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
			// Step 10: Complete Subcontracting Order (subcontracting service)
			// Completes the subcontracting order after successful processing
			{
				StepNumber:    10,
				ServiceName:   "subcontracting",
				HandlerMethod: "CompleteSubcontractingOrder",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
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
				CompensationSteps: []int32{109},
			},

			// ===== COMPENSATION STEPS =====

			// Compensation Step 101: Cancel Subcontract PO (compensates step 2)
			// Cancels the subcontract PO raised with procurement
			{
				StepNumber:    101,
				ServiceName:   "procurement",
				HandlerMethod: "CancelSubcontractPO",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
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
			// Compensation Step 102: Return Issued Materials (compensates step 3)
			// Returns issued materials back to inventory
			{
				StepNumber:    102,
				ServiceName:   "inventory-core",
				HandlerMethod: "ReturnIssuedMaterials",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
					"productID":          "$.input.product_id",
				},
				TimeoutSeconds: 45,
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
			// Compensation Step 103: Revoke Material Shipment (compensates step 4)
			// Revokes/cancels the material shipment to subcontractor
			{
				StepNumber:    103,
				ServiceName:   "subcontracting",
				HandlerMethod: "RevokeMaterialShipment",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
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
			// Compensation Step 104: Revoke Partial Goods Receipt (compensates step 5)
			// Revokes the receipt of partially completed goods
			{
				StepNumber:    104,
				ServiceName:   "subcontracting",
				HandlerMethod: "RevokePartialGoodsReceipt",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
				},
				TimeoutSeconds: 45,
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
			// Compensation Step 105: Revert Quality Inspection (compensates step 6)
			// Reverts the quality inspection data
			{
				StepNumber:    105,
				ServiceName:   "quality-production",
				HandlerMethod: "RevertQualityInspection",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
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
			// Compensation Step 106: Revert Quality Status Update (compensates step 7)
			// Reverts quality status updates
			{
				StepNumber:    106,
				ServiceName:   "subcontracting",
				HandlerMethod: "RevertQualityStatus",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
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
			// Compensation Step 107: Return Completed Goods (compensates step 8)
			// Returns received completed goods back to subcontractor
			{
				StepNumber:    107,
				ServiceName:   "inventory-core",
				HandlerMethod: "ReturnCompletedGoods",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
					"productID":          "$.input.product_id",
				},
				TimeoutSeconds: 45,
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
			// Compensation Step 108: Reverse Subcontractor Invoice (compensates step 9)
			// Reverses the subcontractor invoice in accounts payable
			{
				StepNumber:    108,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ReverseSubcontractorInvoice",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
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
			// Compensation Step 109: Cancel Subcontracting Order Completion (compensates step 10)
			// Cancels the completion of the subcontracting order
			{
				StepNumber:    109,
				ServiceName:   "subcontracting",
				HandlerMethod: "CancelCompletion",
				InputMapping: map[string]string{
					"subcontractOrderID": "$.steps.1.result.subcontract_order_id",
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
func (s *SubcontractingSaga) SagaType() string {
	return "SAGA-M02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *SubcontractingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *SubcontractingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: po_id, subcontractor_id, material_issue_date, product_id
func (s *SubcontractingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type: expected map[string]interface{}")
	}

	// Validate po_id
	if inputMap["po_id"] == nil || inputMap["po_id"] == "" {
		return errors.New("po_id is required for Subcontracting saga")
	}

	// Validate subcontractor_id
	if inputMap["subcontractor_id"] == nil || inputMap["subcontractor_id"] == "" {
		return errors.New("subcontractor_id is required for Subcontracting saga")
	}

	// Validate material_issue_date
	if inputMap["material_issue_date"] == nil || inputMap["material_issue_date"] == "" {
		return errors.New("material_issue_date is required for Subcontracting saga")
	}

	// Validate product_id
	if inputMap["product_id"] == nil || inputMap["product_id"] == "" {
		return errors.New("product_id is required for Subcontracting saga")
	}

	return nil
}
