// Package warranty provides saga handlers for warranty and service module workflows
package warranty

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// SpareParts Saga implements SAGA-W03: Spare Parts workflow
// Business Flow: SubmitPartsRequisition → CheckStockAvailability → InitiateProcurement →
// SelectVendor → GeneratePurchaseOrder → ReceiveParts → PerformQCInspection →
// UpdateInventory → AllocateCost → PostGL → UpdateStockAvailability
// Timeout: 300 seconds, Critical steps: 1,2,5,8,10
type SparePartsSaga struct {
	steps []*saga.StepDefinition
}

// NewSparePartsSaga creates a new Spare Parts saga handler (SAGA-W03)
func NewSparePartsSaga() saga.SagaHandler {
	return &SparePartsSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Submit Parts Requisition
			{
				StepNumber:    1,
				ServiceName:   "spare-parts",
				HandlerMethod: "SubmitPartsRequisition",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"requisitionID":  "$.input.requisition_id",
					"partsCode":      "$.input.parts_code",
					"partName":       "$.input.part_name",
					"quantity":       "$.input.quantity",
					"requisitionDate": "$.input.requisition_date",
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
			// Step 2: Check Stock Availability
			{
				StepNumber:    2,
				ServiceName:   "inventory",
				HandlerMethod: "CheckStockAvailability",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"partsCode":      "$.input.parts_code",
					"quantity":       "$.input.quantity",
					"requisition":    "$.steps.1.result.requisition",
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
			// Step 3: Initiate Procurement
			{
				StepNumber:    3,
				ServiceName:   "procurement",
				HandlerMethod: "InitiateProcurement",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"requisitionID":  "$.input.requisition_id",
					"partsCode":      "$.input.parts_code",
					"quantity":       "$.input.quantity",
					"stockStatus":    "$.steps.2.result.stock_status",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{103},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Select Vendor
			{
				StepNumber:    4,
				ServiceName:   "vendor",
				HandlerMethod: "SelectVendor",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"partsCode":      "$.input.parts_code",
					"quantity":       "$.input.quantity",
					"procurement":    "$.steps.3.result.procurement_initiation",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Generate Purchase Order
			{
				StepNumber:    5,
				ServiceName:   "purchase-order",
				HandlerMethod: "GeneratePurchaseOrder",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"requisitionID":  "$.input.requisition_id",
					"vendorID":       "$.steps.4.result.vendor_id",
					"partsCode":      "$.input.parts_code",
					"quantity":       "$.input.quantity",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Receive Parts
			{
				StepNumber:    6,
				ServiceName:   "procurement",
				HandlerMethod: "ReceiveParts",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"purchaseOrderID": "$.steps.5.result.purchase_order_id",
					"quantity":       "$.input.quantity",
					"receivedDate":   "$.input.received_date",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Perform QC Inspection
			{
				StepNumber:    7,
				ServiceName:   "quality",
				HandlerMethod: "PerformQCInspection",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"purchaseOrderID": "$.steps.5.result.purchase_order_id",
					"partsReceived":  "$.steps.6.result.parts_received",
					"inspectionDate": "$.input.inspection_date",
				},
				TimeoutSeconds: 60,
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
			// Step 8: Update Inventory
			{
				StepNumber:    8,
				ServiceName:   "inventory",
				HandlerMethod: "UpdateInventory",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"partsCode":      "$.input.parts_code",
					"quantity":       "$.input.quantity",
					"qcInspection":   "$.steps.7.result.qc_inspection",
					"updateDate":     "$.input.received_date",
				},
				TimeoutSeconds: 45,
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
			// Step 9: Allocate Cost
			{
				StepNumber:    9,
				ServiceName:   "procurement",
				HandlerMethod: "AllocateCost",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"purchaseOrderID": "$.steps.5.result.purchase_order_id",
					"costCenter":     "$.input.cost_center",
					"partsReceived":  "$.steps.6.result.parts_received",
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
			// Step 10: Post GL (General Ledger)
			{
				StepNumber:    10,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostPartsJournal",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"purchaseOrderID": "$.steps.5.result.purchase_order_id",
					"costAllocation": "$.steps.9.result.cost_allocation",
					"journalDate":    "$.input.received_date",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{110},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 11: Update Stock Availability
			{
				StepNumber:    11,
				ServiceName:   "spare-parts",
				HandlerMethod: "UpdateStockAvailability",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"requisitionID":  "$.input.requisition_id",
					"partsCode":      "$.input.parts_code",
					"inventoryUpdate": "$.steps.8.result.inventory_update",
					"glPosting":      "$.steps.10.result.journal_entries",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
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

			// Step 103: CancelProcurement (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "procurement",
				HandlerMethod: "CancelProcurement",
				InputMapping: map[string]string{
					"requisitionID": "$.input.requisition_id",
					"procurementInitiation": "$.steps.3.result.procurement_initiation",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 104: RejectVendorSelection (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "vendor",
				HandlerMethod: "RejectVendorSelection",
				InputMapping: map[string]string{
					"requisitionID": "$.input.requisition_id",
					"vendorID": "$.steps.4.result.vendor_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: CancelPurchaseOrder (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "purchase-order",
				HandlerMethod: "CancelPurchaseOrder",
				InputMapping: map[string]string{
					"requisitionID": "$.input.requisition_id",
					"purchaseOrderID": "$.steps.5.result.purchase_order_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 106: RejectPartsReceipt (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "procurement",
				HandlerMethod: "RejectPartsReceipt",
				InputMapping: map[string]string{
					"requisitionID": "$.input.requisition_id",
					"partsReceived": "$.steps.6.result.parts_received",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 107: RejectQCInspection (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "quality",
				HandlerMethod: "RejectQCInspection",
				InputMapping: map[string]string{
					"requisitionID": "$.input.requisition_id",
					"qcInspection": "$.steps.7.result.qc_inspection",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 108: ReverseInventoryUpdate (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "inventory",
				HandlerMethod: "ReverseInventoryUpdate",
				InputMapping: map[string]string{
					"requisitionID": "$.input.requisition_id",
					"partsCode": "$.input.parts_code",
					"inventoryUpdate": "$.steps.8.result.inventory_update",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 109: ReverseCostAllocation (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "procurement",
				HandlerMethod: "ReverseCostAllocation",
				InputMapping: map[string]string{
					"requisitionID": "$.input.requisition_id",
					"costAllocation": "$.steps.9.result.cost_allocation",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 110: ReversePartsJournal (compensates step 10)
			{
				StepNumber:    110,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReversePartsJournal",
				InputMapping: map[string]string{
					"requisitionID": "$.input.requisition_id",
					"journalEntryID": "$.steps.10.result.journal_entry_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *SparePartsSaga) SagaType() string {
	return "SAGA-W03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *SparePartsSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *SparePartsSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *SparePartsSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["requisition_id"] == nil {
		return errors.New("requisition_id is required")
	}

	requisitionID, ok := inputMap["requisition_id"].(string)
	if !ok || requisitionID == "" {
		return errors.New("requisition_id must be a non-empty string")
	}

	if inputMap["parts_code"] == nil {
		return errors.New("parts_code is required")
	}

	partsCode, ok := inputMap["parts_code"].(string)
	if !ok || partsCode == "" {
		return errors.New("parts_code must be a non-empty string")
	}

	if inputMap["quantity"] == nil {
		return errors.New("quantity is required")
	}

	return nil
}
