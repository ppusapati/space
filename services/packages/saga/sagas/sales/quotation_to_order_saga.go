// Package sales provides saga handlers for sales module workflows
package sales

import (
	"errors"
	"p9e.in/samavaya/packages/saga"
)

// QuotationToOrderSaga implements the Quotation-to-Order conversion workflow
// Business flow: Accept Quotation → Lock Pricing → Create Order → Update Opportunity → Notify Sales Team
type QuotationToOrderSaga struct {
	steps []*saga.StepDefinition
}

// NewQuotationToOrderSaga creates a new Quotation-to-Order saga handler
func NewQuotationToOrderSaga() saga.SagaHandler {
	return &QuotationToOrderSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Accept Quotation (crm service)
			{
				StepNumber:    1,
				ServiceName:   "crm",
				HandlerMethod: "AcceptQuotation",
				InputMapping: map[string]string{
					"quotationID": "$.input.quotation_id",
					"tenantID":    "$.tenantID",
					"companyID":   "$.companyID",
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

			// Step 2: Lock Pricing (pricing service)
			{
				StepNumber:    2,
				ServiceName:   "pricing",
				HandlerMethod: "LockPricing",
				InputMapping: map[string]string{
					"quotationID": "$.input.quotation_id",
					"tenantID":    "$.tenantID",
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
				CompensationSteps: []int32{102},
			},

			// Step 3: Create Order from Quotation (sales-order service)
			{
				StepNumber:    3,
				ServiceName:   "sales-order",
				HandlerMethod: "CreateFromQuotation",
				InputMapping: map[string]string{
					"quotationID": "$.input.quotation_id",
					"tenantID":    "$.tenantID",
					"companyID":   "$.companyID",
					"branchID":    "$.branchID",
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
				CompensationSteps: []int32{103},
			},

			// Step 4: Update Opportunity to Won (crm service)
			{
				StepNumber:    4,
				ServiceName:   "crm",
				HandlerMethod: "UpdateOpportunity",
				InputMapping: map[string]string{
					"opportunityID": "$.input.opportunity_id",
					"orderID":       "$.steps.3.result.order_id",
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
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
				CompensationSteps: []int32{104},
			},

			// Step 5: Notify Sales Team (notification service)
			{
				StepNumber:    5,
				ServiceName:   "notification",
				HandlerMethod: "NotifySalesTeam",
				InputMapping: map[string]string{
					"orderID":     "$.steps.3.result.order_id",
					"quotationID": "$.input.quotation_id",
				},
				TimeoutSeconds: 15,
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

			// Compensation Step 102: Release Pricing Lock
			{
				StepNumber:    102,
				ServiceName:   "pricing",
				HandlerMethod: "ReleasePricingLock",
				InputMapping: map[string]string{
					"quotationID": "$.input.quotation_id",
				},
				TimeoutSeconds: 20,
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

			// Compensation Step 103: Cancel Order
			{
				StepNumber:    103,
				ServiceName:   "sales-order",
				HandlerMethod: "CancelOrder",
				InputMapping: map[string]string{
					"orderID": "$.steps.3.result.order_id",
					"reason":  "Saga compensation - subsequent steps failed",
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

			// Compensation Step 104: Revert Opportunity Stage
			{
				StepNumber:    104,
				ServiceName:   "crm",
				HandlerMethod: "RevertOpportunityStage",
				InputMapping: map[string]string{
					"opportunityID": "$.input.opportunity_id",
					"tenantID":      "$.tenantID",
				},
				TimeoutSeconds: 20,
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

			// Compensation Step 101: Revert Quotation
			{
				StepNumber:    101,
				ServiceName:   "crm",
				HandlerMethod: "RevertQuotation",
				InputMapping: map[string]string{
					"quotationID": "$.input.quotation_id",
					"tenantID":    "$.tenantID",
				},
				TimeoutSeconds: 20,
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
func (s *QuotationToOrderSaga) SagaType() string {
	return "SAGA-S02"
}

// GetStepDefinitions returns all steps
func (s *QuotationToOrderSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns definition for a specific step
func (s *QuotationToOrderSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the input for saga execution
func (s *QuotationToOrderSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type: expected map[string]interface{}")
	}

	if inputMap["quotation_id"] == nil || inputMap["quotation_id"] == "" {
		return errors.New("quotation_id is required")
	}

	if inputMap["opportunity_id"] == nil || inputMap["opportunity_id"] == "" {
		return errors.New("opportunity_id is required")
	}

	return nil
}
