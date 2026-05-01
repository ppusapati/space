// Package manufacturing provides saga handlers for manufacturing module workflows
package manufacturing

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// QualityReworkSaga implements SAGA-M06: Quality Rework workflow
// Business Flow: Identify defects → Create rework order → Quarantine materials → Plan rework → Execute rework → Verify quality → Update costs → Complete rework
type QualityReworkSaga struct {
	steps []*saga.StepDefinition
}

// NewQualityReworkSaga creates a new Quality Rework saga handler
func NewQualityReworkSaga() saga.SagaHandler {
	return &QualityReworkSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Identify Defects
			{
				StepNumber:    1,
				ServiceName:   "quality-production",
				HandlerMethod: "IdentifyDefect",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"qualityCheckID": "$.input.quality_check_id",
					"productID":      "$.input.product_id",
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
			// Step 2: Create Rework Order
			{
				StepNumber:    2,
				ServiceName:   "production-order",
				HandlerMethod: "CreateReworkOrder",
				InputMapping: map[string]string{
					"qualityCheckID": "$.steps.1.result.quality_check_id",
					"productID":      "$.input.product_id",
					"defectType":     "$.input.defect_type",
					"defectQuantity": "$.steps.1.result.defect_quantity",
				},
				TimeoutSeconds:    15,
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
			// Step 3: Quarantine Materials
			{
				StepNumber:    3,
				ServiceName:   "inventory-core",
				HandlerMethod: "CreateQuarantineLocation",
				InputMapping: map[string]string{
					"qualityCheckID": "$.steps.1.result.quality_check_id",
					"productID":      "$.input.product_id",
					"defectQuantity": "$.steps.1.result.defect_quantity",
					"warehouseID":    "$.steps.1.result.warehouse_id",
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
			// Step 4: Plan Rework
			{
				StepNumber:    4,
				ServiceName:   "job-card",
				HandlerMethod: "CreateReworkJobCard",
				InputMapping: map[string]string{
					"reworkOrderID":   "$.steps.2.result.rework_order_id",
					"qualityCheckID":  "$.steps.1.result.quality_check_id",
					"productID":       "$.input.product_id",
					"defectType":      "$.input.defect_type",
					"defectQuantity":  "$.steps.1.result.defect_quantity",
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
			// Step 5: Execute Rework
			{
				StepNumber:    5,
				ServiceName:   "shop-floor",
				HandlerMethod: "ExecuteRework",
				InputMapping: map[string]string{
					"reworkJobCardID": "$.steps.4.result.rework_job_card_id",
					"reworkOrderID":   "$.steps.2.result.rework_order_id",
					"defectQuantity":  "$.steps.1.result.defect_quantity",
					"defectType":      "$.input.defect_type",
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
			// Step 6: Verify Quality
			{
				StepNumber:    6,
				ServiceName:   "quality-production",
				HandlerMethod: "VerifyReworkQuality",
				InputMapping: map[string]string{
					"reworkJobCardID": "$.steps.4.result.rework_job_card_id",
					"reworkOrderID":   "$.steps.2.result.rework_order_id",
					"productID":       "$.input.product_id",
					"defectType":      "$.input.defect_type",
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
			// Step 7: Update Costs
			{
				StepNumber:    7,
				ServiceName:   "cost-center",
				HandlerMethod: "UpdateReworkCosts",
				InputMapping: map[string]string{
					"reworkOrderID":   "$.steps.2.result.rework_order_id",
					"defectQuantity":  "$.steps.1.result.defect_quantity",
					"defectType":      "$.input.defect_type",
					"reworkJobCardID": "$.steps.4.result.rework_job_card_id",
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
			// Step 8: Complete Rework
			{
				StepNumber:    8,
				ServiceName:   "production-order",
				HandlerMethod: "CompleteReworkOrder",
				InputMapping: map[string]string{
					"reworkOrderID":   "$.steps.2.result.rework_order_id",
					"reworkJobCardID": "$.steps.4.result.rework_job_card_id",
					"productID":       "$.input.product_id",
					"defectQuantity":  "$.steps.1.result.defect_quantity",
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

			// Step 101: Revert Rework Order Creation (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "production-order",
				HandlerMethod: "RevertReworkOrderCreation",
				InputMapping: map[string]string{
					"reworkOrderID": "$.steps.2.result.rework_order_id",
					"productID":     "$.input.product_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 102: Revert Quarantine (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "inventory-core",
				HandlerMethod: "RevertQuarantineLocation",
				InputMapping: map[string]string{
					"quarantineLocationID": "$.steps.3.result.quarantine_location_id",
					"productID":            "$.input.product_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 103: Revert Rework Planning (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "job-card",
				HandlerMethod: "RevertReworkJobCard",
				InputMapping: map[string]string{
					"reworkJobCardID": "$.steps.4.result.rework_job_card_id",
					"reworkOrderID":   "$.steps.2.result.rework_order_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 104: Revert Rework Execution (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "shop-floor",
				HandlerMethod: "RevertReworkExecution",
				InputMapping: map[string]string{
					"reworkJobCardID": "$.steps.4.result.rework_job_card_id",
					"reworkOrderID":   "$.steps.2.result.rework_order_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 105: Revert Quality Verification (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "quality-production",
				HandlerMethod: "RevertQualityVerification",
				InputMapping: map[string]string{
					"reworkJobCardID": "$.steps.4.result.rework_job_card_id",
					"reworkOrderID":   "$.steps.2.result.rework_order_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: Revert Rework Costs (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "cost-center",
				HandlerMethod: "RevertReworkCosts",
				InputMapping: map[string]string{
					"reworkOrderID": "$.steps.2.result.rework_order_id",
					"productID":     "$.input.product_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *QualityReworkSaga) SagaType() string {
	return "SAGA-M06"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *QualityReworkSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *QualityReworkSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *QualityReworkSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["quality_check_id"] == nil {
		return errors.New("quality_check_id is required")
	}

	if inputMap["defect_type"] == nil {
		return errors.New("defect_type is required")
	}

	if inputMap["product_id"] == nil {
		return errors.New("product_id is required")
	}

	return nil
}
