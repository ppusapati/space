// Package gst provides saga handlers for GST compliance workflows
package gst

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// EwayBillSaga implements SAGA-G03: E-way Bill Generation & Management workflow
// Business Flow: ValidateShipment → CalculateConsignmentValue → GenerateEwayBill → UpdateShipmentStatus → PostLogisticsEntry → CompleteEwayBill
// GST Compliance: E-way Bill generation as per GST Rules for goods movement above threshold
type EwayBillSaga struct {
	steps []*saga.StepDefinition
}

// NewEwayBillSaga creates a new E-way Bill Generation saga handler
func NewEwayBillSaga() saga.SagaHandler {
	return &EwayBillSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Validate Shipment for E-way Bill
			{
				StepNumber:    1,
				ServiceName:   "eway-bill",
				HandlerMethod: "ValidateShipmentForEwayBill",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"shipmentID":         "$.input.shipment_id",
					"consignmentValue":   "$.input.consignment_value",
					"gstin":              "$.input.gstin",
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
			// Step 2: Calculate Consignment Details
			{
				StepNumber:    2,
				ServiceName:   "eway-bill",
				HandlerMethod: "CalculateConsignmentDetails",
				InputMapping: map[string]string{
					"ewayBillID":         "$.steps.1.result.eway_bill_id",
					"shipmentID":         "$.input.shipment_id",
					"consignmentValue":   "$.input.consignment_value",
					"goodsDescription":   "$.input.goods_description",
				},
				TimeoutSeconds:    20,
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
			// Step 3: Validate Consignor/Consignee Details
			{
				StepNumber:    3,
				ServiceName:   "gst",
				HandlerMethod: "ValidateConsignorConsignee",
				InputMapping: map[string]string{
					"ewayBillID":       "$.steps.1.result.eway_bill_id",
					"consignorGSTIN":   "$.input.consignor_gstin",
					"consigneeGSTIN":   "$.input.consignee_gstin",
				},
				TimeoutSeconds:    20,
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
			// Step 4: Generate E-way Bill with Portal
			{
				StepNumber:    4,
				ServiceName:   "eway-bill",
				HandlerMethod: "GenerateEwayBillPortal",
				InputMapping: map[string]string{
					"ewayBillID":        "$.steps.1.result.eway_bill_id",
					"consignmentDetails": "$.steps.2.result.consignment_details",
					"gstin":             "$.input.gstin",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Update Shipment Status
			{
				StepNumber:    5,
				ServiceName:   "shipment",
				HandlerMethod: "UpdateShipmentWithEwayBill",
				InputMapping: map[string]string{
					"shipmentID":       "$.input.shipment_id",
					"ewayBillID":       "$.steps.1.result.eway_bill_id",
					"ewayBillNumber":   "$.steps.4.result.eway_bill_number",
					"validityDate":     "$.steps.4.result.validity_date",
				},
				TimeoutSeconds:    20,
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
			// Step 6: Post Logistics Entry
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostEwayBillLogisticsEntry",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"ewayBillID":       "$.steps.1.result.eway_bill_id",
					"consignmentValue": "$.input.consignment_value",
					"journalDate":      "$.input.shipment_date",
				},
				TimeoutSeconds:    25,
				IsCritical:        false,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Generate E-way Bill Report
			{
				StepNumber:    7,
				ServiceName:   "eway-bill",
				HandlerMethod: "GenerateEwayBillReport",
				InputMapping: map[string]string{
					"ewayBillID":   "$.steps.1.result.eway_bill_id",
					"ewayBillNumber": "$.steps.4.result.eway_bill_number",
				},
				TimeoutSeconds:    15,
				IsCritical:        false,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Complete E-way Bill Generation
			{
				StepNumber:    8,
				ServiceName:   "eway-bill",
				HandlerMethod: "CompleteEwayBillGeneration",
				InputMapping: map[string]string{
					"ewayBillID":     "$.steps.1.result.eway_bill_id",
					"ewayBillNumber": "$.steps.4.result.eway_bill_number",
					"completionDate": "$.input.shipment_date",
				},
				TimeoutSeconds:    15,
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
			// ===== COMPENSATION STEPS =====

			// Step 101: Cancel E-way Bill Validation (compensates step 1)
			{
				StepNumber:    101,
				ServiceName:   "eway-bill",
				HandlerMethod: "CancelEwayBillValidation",
				InputMapping: map[string]string{
					"ewayBillID": "$.steps.1.result.eway_bill_id",
					"reason":     "Saga compensation - E-way bill generation failed",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 102: Clear Consignment Calculation (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "eway-bill",
				HandlerMethod: "ClearConsignmentCalculation",
				InputMapping: map[string]string{
					"ewayBillID": "$.steps.1.result.eway_bill_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 103: Revert Consignor/Consignee Validation (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "gst",
				HandlerMethod: "RevertConsignorConsigneeValidation",
				InputMapping: map[string]string{
					"ewayBillID": "$.steps.1.result.eway_bill_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 104: Cancel Portal E-way Bill (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "eway-bill",
				HandlerMethod: "CancelPortalEwayBill",
				InputMapping: map[string]string{
					"ewayBillID":     "$.steps.1.result.eway_bill_id",
					"ewayBillNumber": "$.steps.4.result.eway_bill_number",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: Revert Shipment Update (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "shipment",
				HandlerMethod: "RevertShipmentEwayBillUpdate",
				InputMapping: map[string]string{
					"shipmentID": "$.input.shipment_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: Reverse Logistics Entry (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseEwayBillLogisticsEntry",
				InputMapping: map[string]string{
					"ewayBillID": "$.steps.1.result.eway_bill_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 107: Delete E-way Bill Report (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "eway-bill",
				HandlerMethod: "DeleteEwayBillReport",
				InputMapping: map[string]string{
					"ewayBillID": "$.steps.1.result.eway_bill_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *EwayBillSaga) SagaType() string {
	return "SAGA-G03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *EwayBillSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *EwayBillSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *EwayBillSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["shipment_id"] == nil {
		return errors.New("shipment_id is required")
	}

	if inputMap["consignment_value"] == nil {
		return errors.New("consignment_value is required")
	}

	value, ok := inputMap["consignment_value"].(float64)
	if !ok || value <= 0 {
		return errors.New("consignment_value must be a positive number")
	}

	if inputMap["gstin"] == nil {
		return errors.New("gstin is required")
	}

	if inputMap["consignor_gstin"] == nil {
		return errors.New("consignor_gstin is required")
	}

	if inputMap["consignee_gstin"] == nil {
		return errors.New("consignee_gstin is required")
	}

	if inputMap["goods_description"] == nil {
		return errors.New("goods_description is required")
	}

	if inputMap["shipment_date"] == nil {
		return errors.New("shipment_date is required")
	}

	return nil
}
