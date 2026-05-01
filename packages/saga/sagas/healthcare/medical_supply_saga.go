// Package healthcare provides saga handlers for healthcare workflows
package healthcare

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// MedicalSupplySaga implements SAGA-HC04: Medical Supply Chain & Inventory workflow
// Business Flow: CreateSupplyOrder → ValidateOrderItems → ReserveInventory → ProcessProcurement → ScheduleDelivery → ExecuteQualityInspection → UpdateInventoryRecords → ProcessPayment → ApplySupplyJournal → UpdateCostCenter → CompleteSupplyOrder
// Steps: 11 forward + 8 compensation = 19 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,6,8,10
type MedicalSupplySaga struct {
	steps []*saga.StepDefinition
}

// NewMedicalSupplySaga creates a new Medical Supply Chain saga handler
func NewMedicalSupplySaga() saga.SagaHandler {
	return &MedicalSupplySaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Supply Order
			{
				StepNumber:    1,
				ServiceName:   "medical-supply-chain",
				HandlerMethod: "CreateSupplyOrder",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"supplyOrderID":  "$.input.supply_order_id",
					"itemCode":       "$.input.item_code",
					"quantity":       "$.input.quantity",
					"deliveryDate":   "$.input.delivery_date",
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
			// Step 2: Validate Order Items
			{
				StepNumber:    2,
				ServiceName:   "medical-supply-chain",
				HandlerMethod: "ValidateOrderItems",
				InputMapping: map[string]string{
					"supplyOrderID": "$.steps.1.result.supply_order_id",
					"itemCode":      "$.input.item_code",
					"quantity":      "$.input.quantity",
					"validateRules": "true",
				},
				TimeoutSeconds:    30,
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
			// Step 3: Reserve Inventory
			{
				StepNumber:    3,
				ServiceName:   "inventory",
				HandlerMethod: "ReserveInventory",
				InputMapping: map[string]string{
					"supplyOrderID": "$.steps.1.result.supply_order_id",
					"itemCode":      "$.input.item_code",
					"quantity":      "$.input.quantity",
					"orderDetails":  "$.steps.1.result.order_details",
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
			// Step 4: Process Procurement
			{
				StepNumber:    4,
				ServiceName:   "procurement",
				HandlerMethod: "ProcessProcurement",
				InputMapping: map[string]string{
					"supplyOrderID":   "$.steps.1.result.supply_order_id",
					"itemCode":        "$.input.item_code",
					"quantity":        "$.input.quantity",
					"inventoryReserve": "$.steps.3.result.inventory_reserve",
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
			// Step 5: Schedule Delivery
			{
				StepNumber:    5,
				ServiceName:   "warehouse",
				HandlerMethod: "ScheduleDelivery",
				InputMapping: map[string]string{
					"supplyOrderID":      "$.steps.1.result.supply_order_id",
					"itemCode":           "$.input.item_code",
					"quantity":           "$.input.quantity",
					"deliveryDate":       "$.input.delivery_date",
					"procurementData":    "$.steps.4.result.procurement_data",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{103},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Execute Quality Inspection
			{
				StepNumber:    6,
				ServiceName:   "quality-inspection",
				HandlerMethod: "ExecuteQualityInspection",
				InputMapping: map[string]string{
					"supplyOrderID":   "$.steps.1.result.supply_order_id",
					"itemCode":        "$.input.item_code",
					"quantity":        "$.input.quantity",
					"deliveryDate":    "$.input.delivery_date",
					"deliverySchedule": "$.steps.5.result.delivery_schedule",
				},
				TimeoutSeconds:    45,
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
			// Step 7: Update Inventory Records
			{
				StepNumber:    7,
				ServiceName:   "inventory",
				HandlerMethod: "UpdateInventoryRecords",
				InputMapping: map[string]string{
					"supplyOrderID":    "$.steps.1.result.supply_order_id",
					"itemCode":         "$.input.item_code",
					"quantity":         "$.input.quantity",
					"inspectionResult": "$.steps.6.result.inspection_result",
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
			// Step 8: Process Payment
			{
				StepNumber:    8,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ProcessSupplyPayment",
				InputMapping: map[string]string{
					"supplyOrderID":   "$.steps.1.result.supply_order_id",
					"itemCode":        "$.input.item_code",
					"quantity":        "$.input.quantity",
					"procurementData": "$.steps.4.result.procurement_data",
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
			// Step 9: Apply Supply Journal
			{
				StepNumber:    9,
				ServiceName:   "general-ledger",
				HandlerMethod: "ApplySupplyJournal",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"supplyOrderID": "$.steps.1.result.supply_order_id",
					"paymentData":   "$.steps.8.result.payment_data",
					"journalDate":   "$.input.delivery_date",
				},
				TimeoutSeconds:    30,
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
			// Step 10: Update Cost Center
			{
				StepNumber:    10,
				ServiceName:   "cost-center",
				HandlerMethod: "UpdateCostCenterForSupply",
				InputMapping: map[string]string{
					"supplyOrderID":   "$.steps.1.result.supply_order_id",
					"itemCode":        "$.input.item_code",
					"quantity":        "$.input.quantity",
					"paymentData":     "$.steps.8.result.payment_data",
					"procurementData": "$.steps.4.result.procurement_data",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 11: Complete Supply Order
			{
				StepNumber:    11,
				ServiceName:   "medical-supply-chain",
				HandlerMethod: "CompleteSupplyOrder",
				InputMapping: map[string]string{
					"supplyOrderID":     "$.steps.1.result.supply_order_id",
					"itemCode":          "$.input.item_code",
					"quantity":          "$.input.quantity",
					"paymentData":       "$.steps.8.result.payment_data",
					"inventoryUpdated":  "$.steps.7.result.inventory_updated",
					"completionStatus":  "Completed",
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

			// Step 101: Release Inventory Reserve (compensates step 3)
			{
				StepNumber:    101,
				ServiceName:   "inventory",
				HandlerMethod: "ReleaseInventoryReserve",
				InputMapping: map[string]string{
					"supplyOrderID": "$.steps.1.result.supply_order_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Revert Procurement Processing (compensates step 4)
			{
				StepNumber:    102,
				ServiceName:   "procurement",
				HandlerMethod: "RevertProcurementProcessing",
				InputMapping: map[string]string{
					"supplyOrderID": "$.steps.1.result.supply_order_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 103: Cancel Delivery Schedule (compensates step 5)
			{
				StepNumber:    103,
				ServiceName:   "warehouse",
				HandlerMethod: "CancelDeliverySchedule",
				InputMapping: map[string]string{
					"supplyOrderID": "$.steps.1.result.supply_order_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Revert Quality Inspection (compensates step 6)
			{
				StepNumber:    104,
				ServiceName:   "quality-inspection",
				HandlerMethod: "RevertQualityInspection",
				InputMapping: map[string]string{
					"supplyOrderID": "$.steps.1.result.supply_order_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 105: Revert Inventory Update (compensates step 7)
			{
				StepNumber:    105,
				ServiceName:   "inventory",
				HandlerMethod: "RevertInventoryUpdate",
				InputMapping: map[string]string{
					"supplyOrderID": "$.steps.1.result.supply_order_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Reverse Supply Payment (compensates step 8)
			{
				StepNumber:    106,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ReverseSupplyPayment",
				InputMapping: map[string]string{
					"supplyOrderID": "$.steps.1.result.supply_order_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Reverse Supply Journal (compensates step 9)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseSupplyJournal",
				InputMapping: map[string]string{
					"supplyOrderID": "$.steps.1.result.supply_order_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: Revert Cost Center Update (compensates step 10)
			{
				StepNumber:    108,
				ServiceName:   "cost-center",
				HandlerMethod: "RevertCostCenterUpdateForSupply",
				InputMapping: map[string]string{
					"supplyOrderID": "$.steps.1.result.supply_order_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *MedicalSupplySaga) SagaType() string {
	return "SAGA-HC04"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *MedicalSupplySaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *MedicalSupplySaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *MedicalSupplySaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["supply_order_id"] == nil {
		return errors.New("supply_order_id is required")
	}

	if inputMap["item_code"] == nil {
		return errors.New("item_code is required")
	}

	if inputMap["quantity"] == nil {
		return errors.New("quantity is required")
	}

	if inputMap["delivery_date"] == nil {
		return errors.New("delivery_date is required (format: YYYY-MM-DD)")
	}

	return nil
}
