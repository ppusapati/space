// Package manufacturing provides saga handlers for manufacturing module workflows
package manufacturing

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ProductionOrderSaga implements SAGA-M01: Production Order Execution workflow
// Business Flow: Issue production order → Allocate resources → Reserve materials → Create job cards →
// Schedule production → Allocate capacity → Create routings → Start production → Update job progress →
// Complete production order
//
// Compensation: If any critical step fails, automatically reverses previous steps
// in reverse order to maintain data consistency and release locked resources
type ProductionOrderSaga struct {
	steps []*saga.StepDefinition
}

// NewProductionOrderSaga creates a new Production Order Execution saga handler
func NewProductionOrderSaga() saga.SagaHandler {
	return &ProductionOrderSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Issue Production Order (production-order service)
			// Issues the production order to start the production process
			{
				StepNumber:    1,
				ServiceName:   "production-order",
				HandlerMethod: "IssueProductionOrder",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"productionOrderID": "$.input.production_order_id",
					"productID":      "$.input.product_id",
					"startDate":      "$.input.start_date",
					"scheduledDate":  "$.input.scheduled_date",
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
			// Step 2: Allocate Resources (production-planning service)
			// Allocates necessary resources for production
			{
				StepNumber:    2,
				ServiceName:   "production-planning",
				HandlerMethod: "AllocateResources",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"productionOrderID":     "$.steps.1.result.production_order_id",
					"productID":             "$.input.product_id",
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
			// Step 3: Reserve Materials (shop-floor service)
			// Reserves materials on the shop floor for production
			{
				StepNumber:    3,
				ServiceName:   "shop-floor",
				HandlerMethod: "ReserveMaterials",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
					"productID":         "$.input.product_id",
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
				CompensationSteps: []int32{102},
			},
			// Step 4: Create Job Cards (job-card service)
			// Creates job cards for the production order
			{
				StepNumber:    4,
				ServiceName:   "job-card",
				HandlerMethod: "CreateJobCards",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
					"productID":         "$.input.product_id",
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
				CompensationSteps: []int32{103},
			},
			// Step 5: Schedule Production (production-planning service)
			// Schedules the production timeline and milestones
			{
				StepNumber:    5,
				ServiceName:   "production-planning",
				HandlerMethod: "ScheduleProduction",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
					"scheduledDate":     "$.input.scheduled_date",
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
			// Step 6: Allocate Capacity (work-center service)
			// Allocates capacity at work centers for the production order
			{
				StepNumber:    6,
				ServiceName:   "work-center",
				HandlerMethod: "AllocateCapacity",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
					"productID":         "$.input.product_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{105},
			},
			// Step 7: Create Routings (routing service)
			// Creates routing sequences for production operations
			{
				StepNumber:    7,
				ServiceName:   "routing",
				HandlerMethod: "CreateRoutings",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
					"productID":         "$.input.product_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{106},
			},
			// Step 8: Start Production (shop-floor service)
			// Starts the actual production execution on shop floor
			{
				StepNumber:    8,
				ServiceName:   "shop-floor",
				HandlerMethod: "StartProduction",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
					"productID":         "$.input.product_id",
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
			// Step 9: Update Job Progress (job-card service)
			// Updates job card progress during production
			{
				StepNumber:    9,
				ServiceName:   "job-card",
				HandlerMethod: "UpdateJobProgress",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{108},
			},
			// Step 10: Complete Production Order (production-order service)
			// Completes the production order after successful execution
			{
				StepNumber:    10,
				ServiceName:   "production-order",
				HandlerMethod: "CompleteProductionOrder",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"productionOrderID": "$.steps.1.result.production_order_id",
					"productID":         "$.input.product_id",
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

			// Compensation Step 101: Deallocate Resources (compensates step 2)
			// Deallocates resources assigned to the production order
			{
				StepNumber:    101,
				ServiceName:   "production-planning",
				HandlerMethod: "DeallocateResources",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
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
			// Compensation Step 102: Release Reserved Materials (compensates step 3)
			// Releases reserved materials back to shop floor inventory
			{
				StepNumber:    102,
				ServiceName:   "shop-floor",
				HandlerMethod: "ReleaseReservedMaterials",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
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
			// Compensation Step 103: Delete Job Cards (compensates step 4)
			// Deletes created job cards for the production order
			{
				StepNumber:    103,
				ServiceName:   "job-card",
				HandlerMethod: "DeleteJobCards",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
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
			// Compensation Step 104: Unschedule Production (compensates step 5)
			// Removes production schedule for the order
			{
				StepNumber:    104,
				ServiceName:   "production-planning",
				HandlerMethod: "UnscheduleProduction",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
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
			// Compensation Step 105: Deallocate Capacity (compensates step 6)
			// Deallocates work center capacity
			{
				StepNumber:    105,
				ServiceName:   "work-center",
				HandlerMethod: "DeallocateCapacity",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
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
			// Compensation Step 106: Delete Routings (compensates step 7)
			// Deletes created routing sequences
			{
				StepNumber:    106,
				ServiceName:   "routing",
				HandlerMethod: "DeleteRoutings",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
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
			// Compensation Step 107: Stop Production (compensates step 8)
			// Stops the production execution on shop floor
			{
				StepNumber:    107,
				ServiceName:   "shop-floor",
				HandlerMethod: "StopProduction",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
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
			// Compensation Step 108: Revert Job Progress (compensates step 9)
			// Reverts job card progress updates
			{
				StepNumber:    108,
				ServiceName:   "job-card",
				HandlerMethod: "RevertJobProgress",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
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
			// Compensation Step 109: Cancel Production Order Completion (compensates step 10)
			// Cancels the production order completion status
			{
				StepNumber:    109,
				ServiceName:   "production-order",
				HandlerMethod: "CancelCompletion",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
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
func (s *ProductionOrderSaga) SagaType() string {
	return "SAGA-M01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ProductionOrderSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ProductionOrderSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: production_order_id, product_id, start_date, scheduled_date
func (s *ProductionOrderSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type: expected map[string]interface{}")
	}

	// Validate production_order_id
	if inputMap["production_order_id"] == nil || inputMap["production_order_id"] == "" {
		return errors.New("production_order_id is required for Production Order Execution saga")
	}

	// Validate product_id
	if inputMap["product_id"] == nil || inputMap["product_id"] == "" {
		return errors.New("product_id is required for Production Order Execution saga")
	}

	// Validate start_date
	if inputMap["start_date"] == nil || inputMap["start_date"] == "" {
		return errors.New("start_date is required for Production Order Execution saga")
	}

	// Validate scheduled_date
	if inputMap["scheduled_date"] == nil || inputMap["scheduled_date"] == "" {
		return errors.New("scheduled_date is required for Production Order Execution saga")
	}

	return nil
}
