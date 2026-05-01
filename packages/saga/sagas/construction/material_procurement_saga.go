// Package construction provides saga handlers for construction module workflows
package construction

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// MaterialProcurementSaga implements SAGA-C03: Material Procurement & Site Delivery (longest construction saga)
// Business Flow: CreateMaterialRequisition → ValidateMaterialSpecifications → PublishTenderNotice → EvaluateQuotations → SelectVendor → CreatePurchaseOrder → ScheduleDelivery → TrackShipment → ReceiveMaterialAtSite → InspectDeliveredMaterial → UpdateSiteInventory → PostProcurementEntry
// Timeout: 180 seconds, Critical steps: 1,2,3,4,7,8,11
// Note: 11 forward + 10 compensation = 21 total steps (longest construction saga)
type MaterialProcurementSaga struct {
	steps []*saga.StepDefinition
}

// NewMaterialProcurementSaga creates a new Material Procurement & Site Delivery saga handler
func NewMaterialProcurementSaga() saga.SagaHandler {
	return &MaterialProcurementSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Material Requisition
			{
				StepNumber:    1,
				ServiceName:   "procurement",
				HandlerMethod: "CreateMaterialRequisition",
				InputMapping: map[string]string{
					"tenantID":                "$.tenantID",
					"companyID":               "$.companyID",
					"branchID":                "$.branchID",
					"materialRequisitionID":   "$.input.material_requisition_id",
					"projectID":               "$.input.project_id",
					"deliveryLocation":       "$.input.delivery_location",
					"expectedDeliveryDate":    "$.input.expected_delivery_date",
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
			// Step 2: Validate Material Specifications
			{
				StepNumber:    2,
				ServiceName:   "construction-site",
				HandlerMethod: "ValidateMaterialSpecifications",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"requisition":           "$.steps.1.result.requisition_details",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{112},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Publish Tender Notice
			{
				StepNumber:    3,
				ServiceName:   "procurement",
				HandlerMethod: "PublishTenderNotice",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"specifications":       "$.steps.2.result.validated_specs",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{113},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Evaluate Quotations
			{
				StepNumber:    4,
				ServiceName:   "procurement",
				HandlerMethod: "EvaluateQuotations",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"tenderNotice":          "$.steps.3.result.tender_notice",
				},
				TimeoutSeconds:    45,
				IsCritical:        true,
				CompensationSteps: []int32{114},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Select Vendor
			{
				StepNumber:    5,
				ServiceName:   "procurement",
				HandlerMethod: "SelectVendor",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"evaluatedQuotations":   "$.steps.4.result.evaluated_quotations",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{115},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Create Purchase Order
			{
				StepNumber:    6,
				ServiceName:   "accounts-payable",
				HandlerMethod: "CreatePurchaseOrder",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"vendorSelection":       "$.steps.5.result.vendor_selection",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{116},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Schedule Delivery
			{
				StepNumber:    7,
				ServiceName:   "construction-site",
				HandlerMethod: "ScheduleDelivery",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"purchaseOrder":         "$.steps.6.result.purchase_order",
					"expectedDeliveryDate":  "$.input.expected_delivery_date",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{117},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Track Shipment
			{
				StepNumber:    8,
				ServiceName:   "inventory",
				HandlerMethod: "TrackShipment",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"deliverySchedule":      "$.steps.7.result.delivery_schedule",
				},
				TimeoutSeconds:    45,
				IsCritical:        true,
				CompensationSteps: []int32{118},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Receive Material At Site
			{
				StepNumber:    9,
				ServiceName:   "construction-site",
				HandlerMethod: "ReceiveMaterialAtSite",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"deliveryLocation":      "$.input.delivery_location",
					"shipmentTracking":      "$.steps.8.result.shipment_tracking",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{119},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Inspect Delivered Material
			{
				StepNumber:    10,
				ServiceName:   "quality-inspection",
				HandlerMethod: "InspectDeliveredMaterial",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"receivedMaterial":      "$.steps.9.result.received_material",
					"specifications":       "$.steps.2.result.validated_specs",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{120},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 11: Update Site Inventory
			{
				StepNumber:    11,
				ServiceName:   "construction-site",
				HandlerMethod: "UpdateSiteInventory",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"deliveryLocation":      "$.input.delivery_location",
					"inspectionResult":      "$.steps.10.result.inspection_result",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{121},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 12: Post Procurement Entry (Note: This is step 12, compensation starts at 112)
			{
				StepNumber:    12,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostProcurementEntry",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"purchaseOrder":         "$.steps.6.result.purchase_order",
					"siteInventory":         "$.steps.11.result.inventory_update",
					"journalDate":           "$.input.expected_delivery_date",
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

			// Step 112: InvalidateMaterialSpecifications (compensates step 2)
			{
				StepNumber:    112,
				ServiceName:   "construction-site",
				HandlerMethod: "InvalidateMaterialSpecifications",
				InputMapping: map[string]string{
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"specifications":       "$.steps.2.result.validated_specs",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 113: WithdrawTenderNotice (compensates step 3)
			{
				StepNumber:    113,
				ServiceName:   "procurement",
				HandlerMethod: "WithdrawTenderNotice",
				InputMapping: map[string]string{
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"tenderNotice":          "$.steps.3.result.tender_notice",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 114: RejectQuotations (compensates step 4)
			{
				StepNumber:    114,
				ServiceName:   "procurement",
				HandlerMethod: "RejectQuotations",
				InputMapping: map[string]string{
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"quotations":            "$.steps.4.result.evaluated_quotations",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 115: CancelVendorSelection (compensates step 5)
			{
				StepNumber:    115,
				ServiceName:   "procurement",
				HandlerMethod: "CancelVendorSelection",
				InputMapping: map[string]string{
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"vendorSelection":       "$.steps.5.result.vendor_selection",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 116: CancelPurchaseOrder (compensates step 6)
			{
				StepNumber:    116,
				ServiceName:   "accounts-payable",
				HandlerMethod: "CancelPurchaseOrder",
				InputMapping: map[string]string{
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"purchaseOrderID":       "$.steps.6.result.purchase_order_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 117: CancelDeliverySchedule (compensates step 7)
			{
				StepNumber:    117,
				ServiceName:   "construction-site",
				HandlerMethod: "CancelDeliverySchedule",
				InputMapping: map[string]string{
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"deliverySchedule":      "$.steps.7.result.delivery_schedule",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 118: StopShipmentTracking (compensates step 8)
			{
				StepNumber:    118,
				ServiceName:   "inventory",
				HandlerMethod: "StopShipmentTracking",
				InputMapping: map[string]string{
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"shipmentTracking":      "$.steps.8.result.shipment_tracking",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 119: RejectedMaterialReturn (compensates step 9)
			{
				StepNumber:    119,
				ServiceName:   "construction-site",
				HandlerMethod: "RejectedMaterialReturn",
				InputMapping: map[string]string{
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"receivedMaterial":      "$.steps.9.result.received_material",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 120: CancelMaterialInspection (compensates step 10)
			{
				StepNumber:    120,
				ServiceName:   "quality-inspection",
				HandlerMethod: "CancelMaterialInspection",
				InputMapping: map[string]string{
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"inspectionResult":      "$.steps.10.result.inspection_result",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 121: ReverseInventoryUpdate (compensates step 11)
			{
				StepNumber:    121,
				ServiceName:   "construction-site",
				HandlerMethod: "ReverseInventoryUpdate",
				InputMapping: map[string]string{
					"materialRequisitionID": "$.input.material_requisition_id",
					"projectID":             "$.input.project_id",
					"inventoryUpdate":       "$.steps.11.result.inventory_update",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *MaterialProcurementSaga) SagaType() string {
	return "SAGA-C03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *MaterialProcurementSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *MaterialProcurementSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *MaterialProcurementSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["material_requisition_id"] == nil {
		return errors.New("material_requisition_id is required")
	}

	materialRequisitionID, ok := inputMap["material_requisition_id"].(string)
	if !ok || materialRequisitionID == "" {
		return errors.New("material_requisition_id must be a non-empty string")
	}

	if inputMap["project_id"] == nil {
		return errors.New("project_id is required")
	}

	projectID, ok := inputMap["project_id"].(string)
	if !ok || projectID == "" {
		return errors.New("project_id must be a non-empty string")
	}

	if inputMap["delivery_location"] == nil {
		return errors.New("delivery_location is required")
	}

	deliveryLocation, ok := inputMap["delivery_location"].(string)
	if !ok || deliveryLocation == "" {
		return errors.New("delivery_location must be a non-empty string")
	}

	if inputMap["expected_delivery_date"] == nil {
		return errors.New("expected_delivery_date is required")
	}

	expectedDeliveryDate, ok := inputMap["expected_delivery_date"].(string)
	if !ok || expectedDeliveryDate == "" {
		return errors.New("expected_delivery_date must be a non-empty string")
	}

	return nil
}
