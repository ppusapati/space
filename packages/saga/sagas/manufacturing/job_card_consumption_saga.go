// Package manufacturing provides saga handlers for manufacturing module workflows
package manufacturing

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// JobCardConsumptionSaga implements SAGA-M04: Job Card & Material Consumption workflow
// Business Flow: Get production order → Issue materials → Create job card → Start production → Backflush consumption → Record operations → Update costing → Complete job card
type JobCardConsumptionSaga struct {
	steps []*saga.StepDefinition
}

// NewJobCardConsumptionSaga creates a new Job Card & Material Consumption saga handler
func NewJobCardConsumptionSaga() saga.SagaHandler {
	return &JobCardConsumptionSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Get Production Order
			{
				StepNumber:    1,
				ServiceName:   "production-order",
				HandlerMethod: "GetOrder",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"orderID":        "$.input.order_id",
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
			// Step 2: Issue Materials
			{
				StepNumber:    2,
				ServiceName:   "inventory-core",
				HandlerMethod: "IssueForProduction",
				InputMapping: map[string]string{
					"orderID":     "$.steps.1.result.order_id",
					"bomLines":    "$.steps.1.result.bom_lines",
					"quantity":    "$.steps.1.result.quantity",
					"warehouseID": "$.steps.1.result.warehouse_id",
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
			// Step 3: Create Job Card
			{
				StepNumber:    3,
				ServiceName:   "job-card",
				HandlerMethod: "CreateJobCard",
				InputMapping: map[string]string{
					"orderID":       "$.steps.1.result.order_id",
					"productID":     "$.steps.1.result.product_id",
					"quantity":      "$.steps.1.result.quantity",
					"issuedMaterials": "$.steps.2.result.issued_materials",
				},
				TimeoutSeconds:    15,
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
			// Step 4: Start Production
			{
				StepNumber:    4,
				ServiceName:   "shop-floor",
				HandlerMethod: "StartJobCard",
				InputMapping: map[string]string{
					"jobCardID":  "$.steps.3.result.job_card_id",
					"orderID":    "$.steps.1.result.order_id",
					"employeeID": "$.input.employee_id",
					"startTime":  "$.input.start_time",
				},
				TimeoutSeconds:    15,
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
			// Step 5: Backflush Consumption
			{
				StepNumber:    5,
				ServiceName:   "inventory-core",
				HandlerMethod: "BackflushConsumption",
				InputMapping: map[string]string{
					"jobCardID":       "$.steps.3.result.job_card_id",
					"orderID":         "$.steps.1.result.order_id",
					"productionLines": "$.steps.1.result.production_lines",
					"completedQuantity": "$.input.completed_quantity",
				},
				TimeoutSeconds:    20,
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
			// Step 6: Record Operations
			{
				StepNumber:    6,
				ServiceName:   "shop-floor",
				HandlerMethod: "RecordOperationCompletion",
				InputMapping: map[string]string{
					"jobCardID":        "$.steps.3.result.job_card_id",
					"orderID":          "$.steps.1.result.order_id",
					"completedQuantity": "$.input.completed_quantity",
					"endTime":          "$.input.end_time",
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
			// Step 7: Update Costing
			{
				StepNumber:    7,
				ServiceName:   "cost-center",
				HandlerMethod: "UpdateJobCost",
				InputMapping: map[string]string{
					"jobCardID":        "$.steps.3.result.job_card_id",
					"orderID":          "$.steps.1.result.order_id",
					"completedQuantity": "$.input.completed_quantity",
					"employeeID":       "$.input.employee_id",
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
			// Step 8: Complete Job Card
			{
				StepNumber:    8,
				ServiceName:   "job-card",
				HandlerMethod: "CompleteJobCard",
				InputMapping: map[string]string{
					"jobCardID":        "$.steps.3.result.job_card_id",
					"orderID":          "$.steps.1.result.order_id",
					"completedQuantity": "$.input.completed_quantity",
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

			// Step 101: Revert Material Issuance (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "inventory-core",
				HandlerMethod: "RevertMaterialIssuance",
				InputMapping: map[string]string{
					"issuanceID":   "$.steps.2.result.issuance_id",
					"orderID":      "$.steps.1.result.order_id",
					"issuedMaterials": "$.steps.2.result.issued_materials",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 102: Revert Job Card Creation (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "job-card",
				HandlerMethod: "RevertJobCardCreation",
				InputMapping: map[string]string{
					"jobCardID": "$.steps.3.result.job_card_id",
					"orderID":   "$.steps.1.result.order_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 103: Revert Production Start (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "shop-floor",
				HandlerMethod: "RevertJobCardStart",
				InputMapping: map[string]string{
					"jobCardID": "$.steps.3.result.job_card_id",
					"orderID":   "$.steps.1.result.order_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 104: Revert Backflush Consumption (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "inventory-core",
				HandlerMethod: "RevertBackflushConsumption",
				InputMapping: map[string]string{
					"jobCardID": "$.steps.3.result.job_card_id",
					"orderID":   "$.steps.1.result.order_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 105: Revert Operation Recording (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "shop-floor",
				HandlerMethod: "RevertOperationCompletion",
				InputMapping: map[string]string{
					"jobCardID": "$.steps.3.result.job_card_id",
					"orderID":   "$.steps.1.result.order_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: Revert Costing Update (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "cost-center",
				HandlerMethod: "RevertJobCost",
				InputMapping: map[string]string{
					"jobCardID": "$.steps.3.result.job_card_id",
					"orderID":   "$.steps.1.result.order_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *JobCardConsumptionSaga) SagaType() string {
	return "SAGA-M04"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *JobCardConsumptionSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *JobCardConsumptionSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *JobCardConsumptionSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["order_id"] == nil {
		return errors.New("order_id is required")
	}

	if inputMap["job_card_id"] == nil {
		return errors.New("job_card_id is required")
	}

	if inputMap["employee_id"] == nil {
		return errors.New("employee_id is required")
	}

	return nil
}
