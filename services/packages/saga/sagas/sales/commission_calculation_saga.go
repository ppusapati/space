// Package sales provides saga handlers for sales module workflows
package sales

import (
	"errors"
	"p9e.in/samavaya/packages/saga"
)

// CommissionCalculationSaga implements the Commission Calculation workflow
// Business flow: Mark Invoice Paid → Calculate Commission → Accrue Commission → Post GL
type CommissionCalculationSaga struct {
	steps []*saga.StepDefinition
}

// NewCommissionCalculationSaga creates a new Commission Calculation saga handler
func NewCommissionCalculationSaga() saga.SagaHandler {
	return &CommissionCalculationSaga{
		steps: []*saga.StepDefinition{
			{StepNumber: 1, ServiceName: "accounts-receivable", HandlerMethod: "MarkInvoicePaid", InputMapping: map[string]string{"invoiceID": "$.input.invoice_id", "paymentAmount": "$.input.payment_amount", "tenantID": "$.tenantID"}, TimeoutSeconds: 20, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 2, ServiceName: "commission", HandlerMethod: "CalculateCommission", InputMapping: map[string]string{"invoiceID": "$.input.invoice_id", "amount": "$.input.payment_amount", "tenantID": "$.tenantID", "companyID": "$.companyID"}, TimeoutSeconds: 30, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{102}},
			{StepNumber: 3, ServiceName: "payroll", HandlerMethod: "AccrueCommission", InputMapping: map[string]string{"commissionID": "$.steps.2.result.commission_id", "amount": "$.steps.2.result.commission_amount", "tenantID": "$.tenantID"}, TimeoutSeconds: 30, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{103}},
			{StepNumber: 4, ServiceName: "general-ledger", HandlerMethod: "PostCommission", InputMapping: map[string]string{"commissionID": "$.steps.2.result.commission_id", "amount": "$.steps.2.result.commission_amount", "tenantID": "$.tenantID", "companyID": "$.companyID"}, TimeoutSeconds: 30, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{104}},
			{StepNumber: 5, ServiceName: "commission", HandlerMethod: "NotifySalesPerson", InputMapping: map[string]string{"commissionID": "$.steps.2.result.commission_id", "salesPersonID": "$.steps.2.result.sales_person_id"}, TimeoutSeconds: 15, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 102, ServiceName: "commission", HandlerMethod: "DeleteCommission", InputMapping: map[string]string{"commissionID": "$.steps.2.result.commission_id"}, TimeoutSeconds: 20, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 103, ServiceName: "payroll", HandlerMethod: "ReverseAccrual", InputMapping: map[string]string{"commissionID": "$.steps.2.result.commission_id"}, TimeoutSeconds: 30, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 104, ServiceName: "general-ledger", HandlerMethod: "ReverseCommissionPosting", InputMapping: map[string]string{"commissionID": "$.steps.2.result.commission_id"}, TimeoutSeconds: 30, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
		},
	}
}

func (s *CommissionCalculationSaga) SagaType() string { return "SAGA-S05" }
func (s *CommissionCalculationSaga) GetStepDefinitions() []*saga.StepDefinition { return s.steps }
func (s *CommissionCalculationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

func (s *CommissionCalculationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}
	if inputMap["invoice_id"] == nil {
		return errors.New("invoice_id is required")
	}
	if inputMap["payment_amount"] == nil {
		return errors.New("payment_amount is required")
	}
	return nil
}
