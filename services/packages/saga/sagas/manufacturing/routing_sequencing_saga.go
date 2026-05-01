// Package manufacturing provides saga handlers for manufacturing module workflows
package manufacturing

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// RoutingSequencingSaga implements SAGA-M05: Routing & Operation Sequencing workflow
// Business Flow: Get routing → Sequence operations → Assign work centers → Generate job cards → Schedule jobs
type RoutingSequencingSaga struct {
	steps []*saga.StepDefinition
}

// NewRoutingSequencingSaga creates a new Routing & Operation Sequencing saga handler
func NewRoutingSequencingSaga() saga.SagaHandler {
	return &RoutingSequencingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Get Routing
			{
				StepNumber:    1,
				ServiceName:   "production-order",
				HandlerMethod: "GetRouting",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"orderID":       "$.input.order_id",
					"productID":     "$.input.product_id",
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
			// Step 2: Sequence Operations
			{
				StepNumber:    2,
				ServiceName:   "routing",
				HandlerMethod: "SequenceOperations",
				InputMapping: map[string]string{
					"routingID":  "$.steps.1.result.routing_id",
					"productID":  "$.input.product_id",
					"orderID":    "$.input.order_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{102},
			},
			// Step 3: Assign Work Centers
			{
				StepNumber:    3,
				ServiceName:   "work-center",
				HandlerMethod: "AssignWorkCenters",
				InputMapping: map[string]string{
					"routingID":   "$.steps.1.result.routing_id",
					"orderID":     "$.input.order_id",
					"operationID": "$.steps.2.result.operation_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{103},
			},
			// Step 4: Generate Job Cards
			{
				StepNumber:    4,
				ServiceName:   "job-card",
				HandlerMethod: "GenerateJobCards",
				InputMapping: map[string]string{
					"routingID":     "$.steps.1.result.routing_id",
					"orderID":       "$.input.order_id",
					"operationID":   "$.steps.2.result.operation_id",
					"workCenterID":  "$.steps.3.result.work_center_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{104},
			},
			// Step 5: Schedule Job Cards
			{
				StepNumber:    5,
				ServiceName:   "shop-floor",
				HandlerMethod: "ScheduleJobCards",
				InputMapping: map[string]string{
					"jobCardID":     "$.steps.4.result.job_card_id",
					"workCenterID":  "$.steps.3.result.work_center_id",
					"startDate":     "$.input.start_date",
				},
				TimeoutSeconds:    15,
				IsCritical:        false,
				CompensationSteps: []int32{105},
			},
			// Step 6: Update Production Order
			{
				StepNumber:    6,
				ServiceName:   "production-order",
				HandlerMethod: "UpdateOperations",
				InputMapping: map[string]string{
					"orderID":     "$.input.order_id",
					"jobCardID":   "$.steps.4.result.job_card_id",
					"routingID":   "$.steps.1.result.routing_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{},
			},
			// ===== COMPENSATION STEPS =====

			// Step 102: Revert Operation Sequencing (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "routing",
				HandlerMethod: "RevertSequencing",
				InputMapping: map[string]string{
					"routingID": "$.steps.1.result.routing_id",
					"orderID":   "$.input.order_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 103: Revert Work Center Assignment (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "work-center",
				HandlerMethod: "RevertAssignment",
				InputMapping: map[string]string{
					"routingID": "$.steps.1.result.routing_id",
					"orderID":   "$.input.order_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 104: Delete Job Cards (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "job-card",
				HandlerMethod: "DeleteJobCards",
				InputMapping: map[string]string{
					"jobCardID": "$.steps.4.result.job_card_id",
					"orderID":   "$.input.order_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 105: Cancel Scheduling (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "shop-floor",
				HandlerMethod: "CancelScheduling",
				InputMapping: map[string]string{
					"jobCardID": "$.steps.4.result.job_card_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *RoutingSequencingSaga) SagaType() string {
	return "SAGA-M05"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *RoutingSequencingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *RoutingSequencingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *RoutingSequencingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["order_id"] == nil {
		return errors.New("order_id is required")
	}

	if inputMap["product_id"] == nil {
		return errors.New("product_id is required")
	}

	if inputMap["start_date"] == nil {
		return errors.New("start_date is required")
	}

	return nil
}
