// Package supplychain provides saga handlers for supply chain module workflows
package supplychain

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// InboundLogisticsReceivingSaga implements SAGA-SC01: Inbound Logistics & Receiving
// Business Flow: ReceiveShipment → CreateReceiptRecord → ValidateShipmentDocuments → InspectReceivedGoods → RecordGoodsReceipt → MatchInvoiceDetails → UpdateInventoryStock → ReconcileReceiptCosts → PostReceiptJournals → UpdateSupplierMetrics → FinalizeInboundReceipt
// Timeout: 180 seconds, Critical steps: 1,2,3,4,6,8,11
type InboundLogisticsReceivingSaga struct {
	steps []*saga.StepDefinition
}

// NewInboundLogisticsReceivingSaga creates a new Inbound Logistics & Receiving saga handler
func NewInboundLogisticsReceivingSaga() saga.SagaHandler {
	return &InboundLogisticsReceivingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Receive Shipment
			{
				StepNumber:    1,
				ServiceName:   "logistics",
				HandlerMethod: "ReceiveShipment",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"shipmentID":    "$.input.shipment_id",
					"poNumber":      "$.input.po_number",
					"supplierID":    "$.input.supplier_id",
					"deliveryDate":  "$.input.delivery_date",
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
			// Step 2: Create Receipt Record
			{
				StepNumber:    2,
				ServiceName:   "warehouse",
				HandlerMethod: "CreateReceiptRecord",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"shipmentID":    "$.input.shipment_id",
					"poNumber":      "$.input.po_number",
					"supplierID":    "$.input.supplier_id",
					"shipmentData":  "$.steps.1.result.shipment_data",
					"warehouseID":   "$.steps.1.result.warehouse_id",
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
			// Step 3: Validate Shipment Documents
			{
				StepNumber:    3,
				ServiceName:   "procurement",
				HandlerMethod: "ValidateShipmentDocuments",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"poNumber":      "$.input.po_number",
					"supplierID":    "$.input.supplier_id",
					"shipmentData":  "$.steps.1.result.shipment_data",
					"receiptID":     "$.steps.2.result.receipt_id",
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
			// Step 4: Inspect Received Goods
			{
				StepNumber:    4,
				ServiceName:   "quality-inspection",
				HandlerMethod: "InspectReceivedGoods",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"shipmentID":        "$.input.shipment_id",
					"receiptID":         "$.steps.2.result.receipt_id",
					"shipmentData":      "$.steps.1.result.shipment_data",
					"documentValidation": "$.steps.3.result.validation_status",
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
			// Step 5: Record Goods Receipt
			{
				StepNumber:    5,
				ServiceName:   "warehouse",
				HandlerMethod: "RecordGoodsReceipt",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"receiptID":      "$.steps.2.result.receipt_id",
					"shipmentID":     "$.input.shipment_id",
					"inspectionData": "$.steps.4.result.inspection_data",
					"warehouseID":    "$.steps.1.result.warehouse_id",
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
			// Step 6: Match Invoice Details
			{
				StepNumber:    6,
				ServiceName:   "procurement",
				HandlerMethod: "MatchInvoiceDetails",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"poNumber":        "$.input.po_number",
					"supplierID":      "$.input.supplier_id",
					"receiptID":       "$.steps.2.result.receipt_id",
					"shipmentData":    "$.steps.1.result.shipment_data",
					"receiptDetails":  "$.steps.5.result.receipt_details",
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
			// Step 7: Update Inventory Stock
			{
				StepNumber:    7,
				ServiceName:   "inventory",
				HandlerMethod: "UpdateInventoryStock",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"warehouseID":    "$.steps.1.result.warehouse_id",
					"receiptID":      "$.steps.2.result.receipt_id",
					"receiptDetails": "$.steps.5.result.receipt_details",
					"matchingResult": "$.steps.6.result.matching_result",
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
			// Step 8: Reconcile Receipt Costs
			{
				StepNumber:    8,
				ServiceName:   "procurement",
				HandlerMethod: "ReconcileReceiptCosts",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"poNumber":       "$.input.po_number",
					"supplierID":     "$.input.supplier_id",
					"receiptID":      "$.steps.2.result.receipt_id",
					"matchingResult": "$.steps.6.result.matching_result",
					"inventoryUpdate": "$.steps.7.result.inventory_update",
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
			// Step 9: Post Receipt Journals
			{
				StepNumber:    9,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostReceiptJournals",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"receiptID":        "$.steps.2.result.receipt_id",
					"deliveryDate":     "$.input.delivery_date",
					"costReconciliation": "$.steps.8.result.cost_reconciliation",
					"receiptDetails":   "$.steps.5.result.receipt_details",
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
			// Step 10: Update Supplier Metrics
			{
				StepNumber:    10,
				ServiceName:   "procurement",
				HandlerMethod: "UpdateSupplierMetrics",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"supplierID":      "$.input.supplier_id",
					"receiptID":       "$.steps.2.result.receipt_id",
					"inspectionData":  "$.steps.4.result.inspection_data",
					"costReconciliation": "$.steps.8.result.cost_reconciliation",
				},
				TimeoutSeconds: 30,
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
			// Step 11: Finalize Inbound Receipt
			{
				StepNumber:    11,
				ServiceName:   "warehouse",
				HandlerMethod: "FinalizeInboundReceipt",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"receiptID":        "$.steps.2.result.receipt_id",
					"shipmentID":       "$.input.shipment_id",
					"receiptDetails":   "$.steps.5.result.receipt_details",
					"journalPosting":   "$.steps.9.result.journal_entries",
					"supplierMetrics":  "$.steps.10.result.metrics_update",
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

			// Step 102: RejectReceiptRecord (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "warehouse",
				HandlerMethod: "RejectReceiptRecord",
				InputMapping: map[string]string{
					"receiptID":   "$.steps.2.result.receipt_id",
					"shipmentID":  "$.input.shipment_id",
					"receiptData": "$.steps.2.result.receipt_data",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 103: InvalidateDocumentValidation (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "procurement",
				HandlerMethod: "InvalidateDocumentValidation",
				InputMapping: map[string]string{
					"poNumber":   "$.input.po_number",
					"receiptID":  "$.steps.2.result.receipt_id",
					"validation": "$.steps.3.result.validation_data",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 104: ReverseQualityInspection (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "quality-inspection",
				HandlerMethod: "ReverseQualityInspection",
				InputMapping: map[string]string{
					"shipmentID":      "$.input.shipment_id",
					"receiptID":       "$.steps.2.result.receipt_id",
					"inspectionData":  "$.steps.4.result.inspection_data",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 105: ReverseGoodsReceipt (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "warehouse",
				HandlerMethod: "ReverseGoodsReceipt",
				InputMapping: map[string]string{
					"receiptID":      "$.steps.2.result.receipt_id",
					"receiptDetails": "$.steps.5.result.receipt_details",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 106: ReverseInvoiceMatching (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "procurement",
				HandlerMethod: "ReverseInvoiceMatching",
				InputMapping: map[string]string{
					"receiptID":      "$.steps.2.result.receipt_id",
					"matchingResult": "$.steps.6.result.matching_result",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 107: ReverseInventoryUpdate (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "inventory",
				HandlerMethod: "ReverseInventoryUpdate",
				InputMapping: map[string]string{
					"warehouseID":    "$.steps.1.result.warehouse_id",
					"receiptID":      "$.steps.2.result.receipt_id",
					"inventoryUpdate": "$.steps.7.result.inventory_update",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 108: ReverseCostReconciliation (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "procurement",
				HandlerMethod: "ReverseCostReconciliation",
				InputMapping: map[string]string{
					"receiptID":        "$.steps.2.result.receipt_id",
					"costReconciliation": "$.steps.8.result.cost_reconciliation",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 109: ReverseReceiptJournals (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseReceiptJournals",
				InputMapping: map[string]string{
					"receiptID":      "$.steps.2.result.receipt_id",
					"journalEntries": "$.steps.9.result.journal_entries",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 110: ReverseSupplierMetrics (compensates step 10)
			{
				StepNumber:    110,
				ServiceName:   "procurement",
				HandlerMethod: "ReverseSupplierMetrics",
				InputMapping: map[string]string{
					"supplierID":  "$.input.supplier_id",
					"receiptID":   "$.steps.2.result.receipt_id",
					"metricsUpdate": "$.steps.10.result.metrics_update",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 111: CancelInboundReceiptFinalization (compensates step 11)
			{
				StepNumber:    111,
				ServiceName:   "warehouse",
				HandlerMethod: "CancelInboundReceiptFinalization",
				InputMapping: map[string]string{
					"receiptID":        "$.steps.2.result.receipt_id",
					"shipmentID":       "$.input.shipment_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *InboundLogisticsReceivingSaga) SagaType() string {
	return "SAGA-SC01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *InboundLogisticsReceivingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *InboundLogisticsReceivingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *InboundLogisticsReceivingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["shipment_id"] == nil {
		return errors.New("shipment_id is required")
	}

	shipmentID, ok := inputMap["shipment_id"].(string)
	if !ok || shipmentID == "" {
		return errors.New("shipment_id must be a non-empty string")
	}

	if inputMap["po_number"] == nil {
		return errors.New("po_number is required")
	}

	poNumber, ok := inputMap["po_number"].(string)
	if !ok || poNumber == "" {
		return errors.New("po_number must be a non-empty string")
	}

	if inputMap["supplier_id"] == nil {
		return errors.New("supplier_id is required")
	}

	supplierID, ok := inputMap["supplier_id"].(string)
	if !ok || supplierID == "" {
		return errors.New("supplier_id must be a non-empty string")
	}

	if inputMap["delivery_date"] == nil {
		return errors.New("delivery_date is required")
	}

	deliveryDate, ok := inputMap["delivery_date"].(string)
	if !ok || deliveryDate == "" {
		return errors.New("delivery_date must be a non-empty string")
	}

	return nil
}
