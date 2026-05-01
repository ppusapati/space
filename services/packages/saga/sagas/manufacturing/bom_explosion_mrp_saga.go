// Package manufacturing provides saga handlers for manufacturing module workflows
package manufacturing

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// BOMExplosionMRPSaga implements SAGA-M03: BOM Explosion & MRP workflow
// Business Flow: Receive sales order → Explode BOM → Plan material requirements → Schedule procurement → Update inventory plan → Create manufacturing orders → Confirm schedule → Complete MRP run
type BOMExplosionMRPSaga struct {
	steps []*saga.StepDefinition
}

// NewBOMExplosionMRPSaga creates a new BOM Explosion & MRP saga handler
func NewBOMExplosionMRPSaga() saga.SagaHandler {
	return &BOMExplosionMRPSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Receive Sales Order
			{
				StepNumber:    1,
				ServiceName:   "sales-order",
				HandlerMethod: "GetSalesOrderDetails",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"salesOrderID":  "$.input.sales_order_id",
				},
				TimeoutSeconds: 15,
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
			// Step 2: Explode BOM
			{
				StepNumber:    2,
				ServiceName:   "bom",
				HandlerMethod: "ExploadBillOfMaterial",
				InputMapping: map[string]string{
					"salesOrderID": "$.steps.1.result.sales_order_id",
					"orderLineIDs": "$.steps.1.result.order_line_ids",
					"productID":    "$.input.product_id",
					"bomID":        "$.input.bom_id",
					"quantity":     "$.steps.1.result.quantity",
				},
				TimeoutSeconds:    20,
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
			// Step 3: Plan Material Requirements
			{
				StepNumber:    3,
				ServiceName:   "production-planning",
				HandlerMethod: "PlanRequirements",
				InputMapping: map[string]string{
					"bomLines":      "$.steps.2.result.bom_lines",
					"quantity":      "$.steps.1.result.quantity",
					"requiredDate":  "$.steps.1.result.required_date",
					"salesOrderID":  "$.steps.1.result.sales_order_id",
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
			// Step 4: Schedule Procurement
			{
				StepNumber:    4,
				ServiceName:   "procurement",
				HandlerMethod: "ScheduleRequisitions",
				InputMapping: map[string]string{
					"bomLines":          "$.steps.2.result.bom_lines",
					"materialRequirements": "$.steps.3.result.material_requirements",
					"requiredDate":      "$.steps.1.result.required_date",
					"salesOrderID":      "$.steps.1.result.sales_order_id",
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
			// Step 5: Update Inventory Plan
			{
				StepNumber:    5,
				ServiceName:   "inventory-core",
				HandlerMethod: "UpdateForecastDemand",
				InputMapping: map[string]string{
					"bomLines":           "$.steps.2.result.bom_lines",
					"materialRequirements": "$.steps.3.result.material_requirements",
					"requiredDate":       "$.steps.1.result.required_date",
					"salesOrderID":       "$.steps.1.result.sales_order_id",
				},
				TimeoutSeconds:    15,
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
			// Step 6: Create Manufacturing Orders
			{
				StepNumber:    6,
				ServiceName:   "production-order",
				HandlerMethod: "CreateOrders",
				InputMapping: map[string]string{
					"bomLines":           "$.steps.2.result.bom_lines",
					"materialRequirements": "$.steps.3.result.material_requirements",
					"requiredDate":       "$.steps.1.result.required_date",
					"salesOrderID":       "$.steps.1.result.sales_order_id",
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
			// Step 7: Confirm Schedule
			{
				StepNumber:    7,
				ServiceName:   "work-center",
				HandlerMethod: "ConfirmSchedule",
				InputMapping: map[string]string{
					"orderIDs":     "$.steps.6.result.order_ids",
					"requiredDate": "$.steps.1.result.required_date",
				},
				TimeoutSeconds:    20,
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
			// Step 8: Complete MRP Run
			{
				StepNumber:    8,
				ServiceName:   "production-planning",
				HandlerMethod: "CompleteMRPRun",
				InputMapping: map[string]string{
					"orderIDs":     "$.steps.6.result.order_ids",
					"requiredDate": "$.steps.1.result.required_date",
					"salesOrderID": "$.steps.1.result.sales_order_id",
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

			// Step 101: Revert BOM Explosion (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "bom",
				HandlerMethod: "RevertBillOfMaterial",
				InputMapping: map[string]string{
					"bomLines":     "$.steps.2.result.bom_lines",
					"salesOrderID": "$.steps.1.result.sales_order_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 102: Revert Material Planning (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "production-planning",
				HandlerMethod: "RevertRequirements",
				InputMapping: map[string]string{
					"materialRequirements": "$.steps.3.result.material_requirements",
					"salesOrderID":         "$.steps.1.result.sales_order_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 103: Revert Procurement Scheduling (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "procurement",
				HandlerMethod: "RevertRequisitions",
				InputMapping: map[string]string{
					"requisitionIDs": "$.steps.4.result.requisition_ids",
					"salesOrderID":   "$.steps.1.result.sales_order_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 104: Revert Inventory Plan (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "inventory-core",
				HandlerMethod: "RevertForecastDemand",
				InputMapping: map[string]string{
					"materialRequirements": "$.steps.3.result.material_requirements",
					"salesOrderID":         "$.steps.1.result.sales_order_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 105: Revert Manufacturing Orders (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "production-order",
				HandlerMethod: "RevertOrders",
				InputMapping: map[string]string{
					"orderIDs":     "$.steps.6.result.order_ids",
					"salesOrderID": "$.steps.1.result.sales_order_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: Revert Schedule Confirmation (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "work-center",
				HandlerMethod: "RevertScheduleConfirmation",
				InputMapping: map[string]string{
					"orderIDs":     "$.steps.6.result.order_ids",
					"salesOrderID": "$.steps.1.result.sales_order_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *BOMExplosionMRPSaga) SagaType() string {
	return "SAGA-M03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *BOMExplosionMRPSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *BOMExplosionMRPSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *BOMExplosionMRPSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["sales_order_id"] == nil {
		return errors.New("sales_order_id is required")
	}

	if inputMap["bom_id"] == nil {
		return errors.New("bom_id is required")
	}

	if inputMap["product_id"] == nil {
		return errors.New("product_id is required")
	}

	return nil
}
