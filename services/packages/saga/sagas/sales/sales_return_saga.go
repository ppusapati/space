// Package sales provides saga handlers for sales module workflows
package sales

import (
	"errors"
	"p9e.in/samavaya/packages/saga"
)

// SalesReturnSaga implements the Sales Return workflow
// Business flow: Create Return → Inspect Goods → Receive Return → Create Credit Note →
// Adjust AR → Post GL → Process Refund → Complete Return
type SalesReturnSaga struct {
	steps []*saga.StepDefinition
}

// NewSalesReturnSaga creates a new Sales Return saga handler
func NewSalesReturnSaga() saga.SagaHandler {
	return &SalesReturnSaga{
		steps: []*saga.StepDefinition{
			{StepNumber: 1, ServiceName: "returns", HandlerMethod: "CreateReturn", InputMapping: map[string]string{"invoiceID": "$.input.invoice_id", "tenantID": "$.tenantID", "companyID": "$.companyID"}, TimeoutSeconds: 20, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 2, ServiceName: "qc", HandlerMethod: "InspectReturnedGoods", InputMapping: map[string]string{"returnID": "$.steps.1.result.return_id", "tenantID": "$.tenantID"}, TimeoutSeconds: 30, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 3, ServiceName: "inventory-core", HandlerMethod: "ReceiveReturn", InputMapping: map[string]string{"returnID": "$.steps.1.result.return_id", "tenantID": "$.tenantID"}, TimeoutSeconds: 30, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{103}},
			{StepNumber: 4, ServiceName: "sales-invoice", HandlerMethod: "CreateCreditNote", InputMapping: map[string]string{"returnID": "$.steps.1.result.return_id", "invoiceID": "$.input.invoice_id", "tenantID": "$.tenantID", "companyID": "$.companyID"}, TimeoutSeconds: 30, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{104}},
			{StepNumber: 5, ServiceName: "accounts-receivable", HandlerMethod: "AdjustAR", InputMapping: map[string]string{"creditNoteID": "$.steps.4.result.credit_note_id", "tenantID": "$.tenantID", "companyID": "$.companyID"}, TimeoutSeconds: 30, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{105}},
			{StepNumber: 6, ServiceName: "general-ledger", HandlerMethod: "PostCreditNote", InputMapping: map[string]string{"creditNoteID": "$.steps.4.result.credit_note_id", "tenantID": "$.tenantID", "companyID": "$.companyID"}, TimeoutSeconds: 30, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{106}},
			{StepNumber: 7, ServiceName: "banking", HandlerMethod: "ProcessRefund", InputMapping: map[string]string{"creditNoteID": "$.steps.4.result.credit_note_id", "amount": "$.input.return_amount", "tenantID": "$.tenantID"}, TimeoutSeconds: 60, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{107}},
			{StepNumber: 8, ServiceName: "returns", HandlerMethod: "CompleteReturn", InputMapping: map[string]string{"returnID": "$.steps.1.result.return_id", "tenantID": "$.tenantID"}, TimeoutSeconds: 20, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 103, ServiceName: "inventory-core", HandlerMethod: "RemoveReturnStock", InputMapping: map[string]string{"returnID": "$.steps.1.result.return_id"}, TimeoutSeconds: 30, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 104, ServiceName: "sales-invoice", HandlerMethod: "DeleteCreditNote", InputMapping: map[string]string{"creditNoteID": "$.steps.4.result.credit_note_id"}, TimeoutSeconds: 30, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 105, ServiceName: "accounts-receivable", HandlerMethod: "ReverseARAdjustment", InputMapping: map[string]string{"creditNoteID": "$.steps.4.result.credit_note_id", "tenantID": "$.tenantID"}, TimeoutSeconds: 30, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 106, ServiceName: "general-ledger", HandlerMethod: "ReverseGLPosting", InputMapping: map[string]string{"creditNoteID": "$.steps.4.result.credit_note_id", "tenantID": "$.tenantID"}, TimeoutSeconds: 30, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 107, ServiceName: "banking", HandlerMethod: "ReverseRefund", InputMapping: map[string]string{"creditNoteID": "$.steps.4.result.credit_note_id", "tenantID": "$.tenantID"}, TimeoutSeconds: 60, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 101, ServiceName: "returns", HandlerMethod: "CancelReturn", InputMapping: map[string]string{"returnID": "$.steps.1.result.return_id", "reason": "Saga compensation"}, TimeoutSeconds: 20, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
		},
	}
}

func (s *SalesReturnSaga) SagaType() string { return "SAGA-S04" }
func (s *SalesReturnSaga) GetStepDefinitions() []*saga.StepDefinition { return s.steps }
func (s *SalesReturnSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

func (s *SalesReturnSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}
	if inputMap["invoice_id"] == nil {
		return errors.New("invoice_id is required")
	}
	if inputMap["return_amount"] == nil {
		return errors.New("return_amount is required")
	}
	return nil
}
