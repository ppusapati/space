// Package supplychain provides saga handlers for supply chain module workflows
package supplychain

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ThirdPartyLogisticsCoordinationSaga implements SAGA-SC03: 3PL Coordination & Outsourced Logistics
// Business Flow: CreateThirdPartyOrder → SelectLogisticsProvider → AssignPickupSchedule → RequestShipmentQuote → ValidateThirdPartyCapacity → CreateLogisticsContract → ManageCarrierTracking → ReconcileFreightCharges → PostThirdPartyJournals → CompleteThirdPartyOrder
// Timeout: 180 seconds, Critical steps: 1,2,3,4,6,8,10
type ThirdPartyLogisticsCoordinationSaga struct {
	steps []*saga.StepDefinition
}

// NewThirdPartyLogisticsCoordinationSaga creates a new 3PL Coordination & Outsourced Logistics saga handler
func NewThirdPartyLogisticsCoordinationSaga() saga.SagaHandler {
	return &ThirdPartyLogisticsCoordinationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Third-Party Order
			{
				StepNumber:    1,
				ServiceName:   "logistics",
				HandlerMethod: "CreateThirdPartyOrder",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"thirdPartyOrderID": "$.input.3pl_order_id",
					"providerID":      "$.input.provider_id",
					"shipmentDate":    "$.input.shipment_date",
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
			// Step 2: Select Logistics Provider
			{
				StepNumber:    2,
				ServiceName:   "third-party-logistics",
				HandlerMethod: "SelectLogisticsProvider",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"providerID":          "$.input.provider_id",
					"shipmentDate":        "$.input.shipment_date",
					"thirdPartyOrderData": "$.steps.1.result.third_party_order_data",
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
			// Step 3: Assign Pickup Schedule
			{
				StepNumber:    3,
				ServiceName:   "warehouse",
				HandlerMethod: "AssignPickupSchedule",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"shipmentDate":      "$.input.shipment_date",
					"providerSelection": "$.steps.2.result.provider_selection",
					"thirdPartyOrderData": "$.steps.1.result.third_party_order_data",
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
			// Step 4: Request Shipment Quote
			{
				StepNumber:    4,
				ServiceName:   "shipment",
				HandlerMethod: "RequestShipmentQuote",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"thirdPartyOrderID": "$.input.3pl_order_id",
					"providerID":        "$.input.provider_id",
					"pickupSchedule":    "$.steps.3.result.pickup_schedule",
					"thirdPartyOrderData": "$.steps.1.result.third_party_order_data",
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
			// Step 5: Validate Third-Party Capacity
			{
				StepNumber:    5,
				ServiceName:   "third-party-logistics",
				HandlerMethod: "ValidateThirdPartyCapacity",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"providerID":      "$.input.provider_id",
					"shipmentDate":    "$.input.shipment_date",
					"shipmentQuote":   "$.steps.4.result.shipment_quote",
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
			// Step 6: Create Logistics Contract
			{
				StepNumber:    6,
				ServiceName:   "logistics",
				HandlerMethod: "CreateLogisticsContract",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"thirdPartyOrderID":     "$.input.3pl_order_id",
					"providerID":            "$.input.provider_id",
					"shipmentQuote":         "$.steps.4.result.shipment_quote",
					"capacityValidation":    "$.steps.5.result.capacity_validation",
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
			// Step 7: Manage Carrier Tracking
			{
				StepNumber:    7,
				ServiceName:   "shipment",
				HandlerMethod: "ManageCarrierTracking",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"thirdPartyOrderID": "$.input.3pl_order_id",
					"logisticsContract": "$.steps.6.result.logistics_contract",
					"pickupSchedule":    "$.steps.3.result.pickup_schedule",
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
			// Step 8: Reconcile Freight Charges
			{
				StepNumber:    8,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ReconcileFreightCharges",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"thirdPartyOrderID":  "$.input.3pl_order_id",
					"providerID":         "$.input.provider_id",
					"shipmentQuote":      "$.steps.4.result.shipment_quote",
					"logisticsContract":  "$.steps.6.result.logistics_contract",
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
			// Step 9: Post Third-Party Journals
			{
				StepNumber:    9,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostThirdPartyJournals",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"thirdPartyOrderID":  "$.input.3pl_order_id",
					"shipmentDate":       "$.input.shipment_date",
					"freightCharges":     "$.steps.8.result.freight_charges",
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
			// Step 10: Complete Third-Party Order
			{
				StepNumber:    10,
				ServiceName:   "logistics",
				HandlerMethod: "CompleteThirdPartyOrder",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"thirdPartyOrderID":  "$.input.3pl_order_id",
					"logisticsContract":  "$.steps.6.result.logistics_contract",
					"trackingData":       "$.steps.7.result.tracking_data",
					"journalEntries":     "$.steps.9.result.journal_entries",
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

			// Step 102: ReverseProviderSelection (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "third-party-logistics",
				HandlerMethod: "ReverseProviderSelection",
				InputMapping: map[string]string{
					"thirdPartyOrderID": "$.input.3pl_order_id",
					"providerSelection": "$.steps.2.result.provider_selection",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 103: CancelPickupSchedule (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "warehouse",
				HandlerMethod: "CancelPickupSchedule",
				InputMapping: map[string]string{
					"pickupSchedule": "$.steps.3.result.pickup_schedule",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 104: CancelShipmentQuote (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "shipment",
				HandlerMethod: "CancelShipmentQuote",
				InputMapping: map[string]string{
					"thirdPartyOrderID": "$.input.3pl_order_id",
					"shipmentQuote":     "$.steps.4.result.shipment_quote",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 105: ReverseCapacityValidation (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "third-party-logistics",
				HandlerMethod: "ReverseCapacityValidation",
				InputMapping: map[string]string{
					"providerID":         "$.input.provider_id",
					"capacityValidation": "$.steps.5.result.capacity_validation",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 106: VoidLogisticsContract (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "logistics",
				HandlerMethod: "VoidLogisticsContract",
				InputMapping: map[string]string{
					"thirdPartyOrderID":  "$.input.3pl_order_id",
					"logisticsContract":  "$.steps.6.result.logistics_contract",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 107: StopCarrierTracking (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "shipment",
				HandlerMethod: "StopCarrierTracking",
				InputMapping: map[string]string{
					"thirdPartyOrderID": "$.input.3pl_order_id",
					"trackingData":      "$.steps.7.result.tracking_data",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 108: ReverseFreightReconciliation (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ReverseFreightReconciliation",
				InputMapping: map[string]string{
					"thirdPartyOrderID": "$.input.3pl_order_id",
					"freightCharges":    "$.steps.8.result.freight_charges",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 109: ReverseThirdPartyJournals (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseThirdPartyJournals",
				InputMapping: map[string]string{
					"thirdPartyOrderID": "$.input.3pl_order_id",
					"journalEntries":    "$.steps.9.result.journal_entries",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 110: CancelThirdPartyOrderCompletion (compensates step 10)
			{
				StepNumber:    110,
				ServiceName:   "logistics",
				HandlerMethod: "CancelThirdPartyOrderCompletion",
				InputMapping: map[string]string{
					"thirdPartyOrderID": "$.input.3pl_order_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ThirdPartyLogisticsCoordinationSaga) SagaType() string {
	return "SAGA-SC03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ThirdPartyLogisticsCoordinationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ThirdPartyLogisticsCoordinationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ThirdPartyLogisticsCoordinationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["3pl_order_id"] == nil {
		return errors.New("3pl_order_id is required")
	}

	thirdPartyOrderID, ok := inputMap["3pl_order_id"].(string)
	if !ok || thirdPartyOrderID == "" {
		return errors.New("3pl_order_id must be a non-empty string")
	}

	if inputMap["provider_id"] == nil {
		return errors.New("provider_id is required")
	}

	providerID, ok := inputMap["provider_id"].(string)
	if !ok || providerID == "" {
		return errors.New("provider_id must be a non-empty string")
	}

	if inputMap["shipment_date"] == nil {
		return errors.New("shipment_date is required")
	}

	shipmentDate, ok := inputMap["shipment_date"].(string)
	if !ok || shipmentDate == "" {
		return errors.New("shipment_date must be a non-empty string")
	}

	return nil
}
