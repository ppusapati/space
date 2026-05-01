// Package sales provides saga handlers for sales module workflows
package sales

import (
	"errors"
	"p9e.in/samavaya/packages/saga"
)

// EInvoiceGenerationSaga implements the E-Invoice Generation workflow
// Business flow: Validate → Generate JSON → Call GSTN API → Update Invoice →
// Record in GST Ledger → Audit Log → Send Notification
type EInvoiceGenerationSaga struct {
	steps []*saga.StepDefinition
}

// NewEInvoiceGenerationSaga creates a new E-Invoice Generation saga handler
func NewEInvoiceGenerationSaga() saga.SagaHandler {
	return &EInvoiceGenerationSaga{
		steps: []*saga.StepDefinition{
			{StepNumber: 1, ServiceName: "sales-invoice", HandlerMethod: "ValidateForEInvoice", InputMapping: map[string]string{"invoiceID": "$.input.invoice_id", "tenantID": "$.tenantID"}, TimeoutSeconds: 20, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 2, ServiceName: "e-invoice", HandlerMethod: "GenerateJSON", InputMapping: map[string]string{"invoiceID": "$.input.invoice_id", "tenantID": "$.tenantID"}, TimeoutSeconds: 30, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{102}},
			{StepNumber: 3, ServiceName: "e-invoice", HandlerMethod: "CallGSTNAPI", InputMapping: map[string]string{"invoiceID": "$.input.invoice_id", "jsonData": "$.steps.2.result.json_data", "tenantID": "$.tenantID"}, TimeoutSeconds: 120, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 10, InitialBackoffMs: 2000, MaxBackoffMs: 120000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{103}},
			{StepNumber: 4, ServiceName: "sales-invoice", HandlerMethod: "UpdateWithIRN", InputMapping: map[string]string{"invoiceID": "$.input.invoice_id", "irn": "$.steps.3.result.irn", "qrCode": "$.steps.3.result.qr_code", "tenantID": "$.tenantID"}, TimeoutSeconds: 30, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{104}},
			{StepNumber: 5, ServiceName: "gst", HandlerMethod: "RecordEInvoice", InputMapping: map[string]string{"invoiceID": "$.input.invoice_id", "irn": "$.steps.3.result.irn", "tenantID": "$.tenantID", "companyID": "$.companyID"}, TimeoutSeconds: 30, IsCritical: true, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{105}},
			{StepNumber: 6, ServiceName: "audit", HandlerMethod: "LogEInvoice", InputMapping: map[string]string{"invoiceID": "$.input.invoice_id", "irn": "$.steps.3.result.irn", "tenantID": "$.tenantID"}, TimeoutSeconds: 20, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 7, ServiceName: "notification", HandlerMethod: "SendEInvoice", InputMapping: map[string]string{"invoiceID": "$.input.invoice_id", "irn": "$.steps.3.result.irn", "qrCode": "$.steps.3.result.qr_code"}, TimeoutSeconds: 15, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 102, ServiceName: "e-invoice", HandlerMethod: "DeleteJSON", InputMapping: map[string]string{"invoiceID": "$.input.invoice_id"}, TimeoutSeconds: 20, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 103, ServiceName: "e-invoice", HandlerMethod: "MarkGSTNFailed", InputMapping: map[string]string{"invoiceID": "$.input.invoice_id"}, TimeoutSeconds: 20, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 104, ServiceName: "sales-invoice", HandlerMethod: "RemoveIRN", InputMapping: map[string]string{"invoiceID": "$.input.invoice_id"}, TimeoutSeconds: 30, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
			{StepNumber: 105, ServiceName: "gst", HandlerMethod: "RemoveEInvoiceRecord", InputMapping: map[string]string{"invoiceID": "$.input.invoice_id"}, TimeoutSeconds: 30, IsCritical: false, RetryConfig: &saga.RetryConfiguration{MaxRetries: 3, InitialBackoffMs: 1000, MaxBackoffMs: 30000, BackoffMultiplier: 2.0, JitterFraction: 0.1}, CompensationSteps: []int32{}},
		},
	}
}

func (s *EInvoiceGenerationSaga) SagaType() string { return "SAGA-S06" }
func (s *EInvoiceGenerationSaga) GetStepDefinitions() []*saga.StepDefinition { return s.steps }
func (s *EInvoiceGenerationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

func (s *EInvoiceGenerationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}
	if inputMap["invoice_id"] == nil {
		return errors.New("invoice_id is required")
	}
	return nil
}
