// Package sales provides saga handlers for sales module workflows
package sales

import (
	"errors"
	"p9e.in/samavaya/packages/saga"
)

// DealerIncentiveSaga implements the Dealer Performance & Incentive workflow
// Business flow: Calculate Dealer Sales → Calculate Incentive → Request Approval →
// Approve Incentive → Create AP Entry → Post GL → Notify Dealer
type DealerIncentiveSaga struct {
	steps []*saga.StepDefinition
}

// NewDealerIncentiveSaga creates a new Dealer Incentive saga handler
func NewDealerIncentiveSaga() saga.SagaHandler {
	return &DealerIncentiveSaga{
		steps: []*saga.StepDefinition{
			{StepNumber: 1, ServiceName: "sales-analytics", HandlerMethod: "CalculateDealerSales", InputMapping: map[string]string{"dealerID": "$.input.dealer_id", "month": "$.input.month", "tenantID": "$.tenantID", "companyID": "$.companyID"}, TimeoutSeconds: 60, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 2, ServiceName: "dealer", HandlerMethod: "CalculateIncentive", InputMapping: map[string]string{"dealerID": "$.input.dealer_id", "salesAmount": "$.steps.1.result.total_sales", "month": "$.input.month", "tenantID": "$.tenantID", "companyID": "$.companyID"}, TimeoutSeconds: 30, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{102}},
			{StepNumber: 3, ServiceName: "approval", HandlerMethod: "RequestIncentiveApproval", InputMapping: map[string]string{"dealerID": "$.input.dealer_id", "incentiveID": "$.steps.2.result.incentive_id", "amount": "$.steps.2.result.incentive_amount", "tenantID": "$.tenantID"}, TimeoutSeconds: 30, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{103}},
			{StepNumber: 4, ServiceName: "approval", HandlerMethod: "ApproveIncentive", InputMapping: map[string]string{"approvalID": "$.steps.3.result.approval_id", "tenantID": "$.tenantID"}, TimeoutSeconds: 30, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{104}},
			{StepNumber: 5, ServiceName: "accounts-payable", HandlerMethod: "CreateIncentivePayable", InputMapping: map[string]string{"dealerID": "$.input.dealer_id", "incentiveID": "$.steps.2.result.incentive_id", "amount": "$.steps.2.result.incentive_amount", "tenantID": "$.tenantID", "companyID": "$.companyID"}, TimeoutSeconds: 30, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{105}},
			{StepNumber: 6, ServiceName: "general-ledger", HandlerMethod: "PostIncentive", InputMapping: map[string]string{"incentiveID": "$.steps.2.result.incentive_id", "amount": "$.steps.2.result.incentive_amount", "tenantID": "$.tenantID", "companyID": "$.companyID"}, TimeoutSeconds: 30, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{106}},
			{StepNumber: 7, ServiceName: "dealer", HandlerMethod: "NotifyDealer", InputMapping: map[string]string{"dealerID": "$.input.dealer_id", "incentiveAmount": "$.steps.2.result.incentive_amount", "month": "$.input.month"}, TimeoutSeconds: 15, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 102, ServiceName: "dealer", HandlerMethod: "DeleteIncentive", InputMapping: map[string]string{"incentiveID": "$.steps.2.result.incentive_id"}, TimeoutSeconds: 20, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 103, ServiceName: "approval", HandlerMethod: "CancelApprovalRequest", InputMapping: map[string]string{"approvalID": "$.steps.3.result.approval_id"}, TimeoutSeconds: 20, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 104, ServiceName: "approval", HandlerMethod: "RejectApproval", InputMapping: map[string]string{"approvalID": "$.steps.3.result.approval_id", "reason": "Saga compensation"}, TimeoutSeconds: 20, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 105, ServiceName: "accounts-payable", HandlerMethod: "DeleteAPEntry", InputMapping: map[string]string{"incentiveID": "$.steps.2.result.incentive_id"}, TimeoutSeconds: 30, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 106, ServiceName: "general-ledger", HandlerMethod: "ReverseIncentivePosting", InputMapping: map[string]string{"incentiveID": "$.steps.2.result.incentive_id"}, TimeoutSeconds: 30, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
		},
	}
}

func (s *DealerIncentiveSaga) SagaType() string { return "SAGA-S07" }
func (s *DealerIncentiveSaga) GetStepDefinitions() []*saga.StepDefinition { return s.steps }
func (s *DealerIncentiveSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

func (s *DealerIncentiveSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}
	if inputMap["dealer_id"] == nil {
		return errors.New("dealer_id is required")
	}
	if inputMap["month"] == nil {
		return errors.New("month is required")
	}
	return nil
}
