// Package projects provides saga handlers for projects module workflows
package projects

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ProgressBillingSaga implements SAGA-PR02: Progress Billing (Construction)
// Business Flow: Get BOQ → Measure work progress → Calculate invoice amount → Deduct retention → Create invoice → Post revenue → Send invoice → Complete billing
type ProgressBillingSaga struct {
	steps []*saga.StepDefinition
}

// NewProgressBillingSaga creates a new Progress Billing saga handler
func NewProgressBillingSaga() saga.SagaHandler {
	return &ProgressBillingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Get BOQ Certificate
			{
				StepNumber:    1,
				ServiceName:   "boq",
				HandlerMethod: "GetBOQCertificate",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"boqID":          "$.input.boq_id",
					"billingCycle":   "$.input.billing_cycle",
					"measuredDate":   "$.input.measured_date",
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
			// Step 2: Measure Work Progress
			{
				StepNumber:    2,
				ServiceName:   "progress-billing",
				HandlerMethod: "MeasureWorkProgress",
				InputMapping: map[string]string{
					"boqID":              "$.input.boq_id",
					"progressPercentage": "$.input.progress_percentage",
					"measuredDate":       "$.input.measured_date",
					"billingCycle":       "$.input.billing_cycle",
					"boqCertificate":     "$.steps.1.result.boq_certificate",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{101},
			},
			// Step 3: Calculate Billing Amount
			{
				StepNumber:    3,
				ServiceName:   "boq",
				HandlerMethod: "CalculateBillingAmount",
				InputMapping: map[string]string{
					"boqID":           "$.input.boq_id",
					"progress":        "$.steps.2.result.measured_progress",
					"progressPercentage": "$.input.progress_percentage",
					"contractAmount":  "$.steps.1.result.contract_amount",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{102},
			},
			// Step 4: Deduct Retention Money
			{
				StepNumber:    4,
				ServiceName:   "progress-billing",
				HandlerMethod: "CalculateRetention",
				InputMapping: map[string]string{
					"boqID":          "$.input.boq_id",
					"billingAmount":  "$.steps.3.result.billing_amount",
					"retentionRate":  "$.steps.1.result.retention_percentage",
					"billingCycle":   "$.input.billing_cycle",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{103},
			},
			// Step 5: Create Progress Invoice
			{
				StepNumber:    5,
				ServiceName:   "sales-invoice",
				HandlerMethod: "CreateProgressInvoice",
				InputMapping: map[string]string{
					"boqID":                "$.input.boq_id",
					"billingCycle":         "$.input.billing_cycle",
					"netBillingAmount":     "$.steps.4.result.net_amount",
					"retentionAmount":      "$.steps.4.result.retention_amount",
					"invoiceDate":          "$.input.measured_date",
					"progressPercentage":   "$.input.progress_percentage",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{104},
			},
			// Step 6: Post Progress Revenue
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostProgressRevenue",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"invoiceID":       "$.steps.5.result.invoice_id",
					"boqID":           "$.input.boq_id",
					"revenueAmount":   "$.steps.4.result.net_amount",
					"retentionAmount": "$.steps.4.result.retention_amount",
					"journalDate":     "$.input.measured_date",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{105},
			},
			// Step 7: Send Progress Invoice
			{
				StepNumber:    7,
				ServiceName:   "notification",
				HandlerMethod: "SendInvoice",
				InputMapping: map[string]string{
					"invoiceID":     "$.steps.5.result.invoice_id",
					"customerID":    "$.steps.1.result.customer_id",
					"customerEmail": "$.steps.1.result.customer_email",
					"invoiceNumber": "$.steps.5.result.invoice_number",
					"billingCycle":  "$.input.billing_cycle",
				},
				TimeoutSeconds:    10,
				IsCritical:        false,
				CompensationSteps: []int32{106},
			},
			// Step 8: Complete Billing Cycle
			{
				StepNumber:    8,
				ServiceName:   "progress-billing",
				HandlerMethod: "CompleteBillingCycle",
				InputMapping: map[string]string{
					"boqID":           "$.input.boq_id",
					"billingCycle":    "$.input.billing_cycle",
					"invoiceID":       "$.steps.5.result.invoice_id",
					"netAmount":       "$.steps.4.result.net_amount",
					"progressData":    "$.steps.2.result.measured_progress",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{},
			},
			// ===== COMPENSATION STEPS =====

			// Step 101: Revert Work Progress Measurement (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "progress-billing",
				HandlerMethod: "RevertWorkProgressMeasurement",
				InputMapping: map[string]string{
					"boqID":       "$.input.boq_id",
					"billingCycle": "$.input.billing_cycle",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 102: Reverse Billing Amount Calculation (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "boq",
				HandlerMethod: "ReverseBillingAmountCalculation",
				InputMapping: map[string]string{
					"boqID": "$.input.boq_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 103: Reverse Retention Deduction (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "progress-billing",
				HandlerMethod: "ReverseRetentionDeduction",
				InputMapping: map[string]string{
					"boqID":        "$.input.boq_id",
					"billingCycle": "$.input.billing_cycle",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 104: Cancel Progress Invoice (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "sales-invoice",
				HandlerMethod: "CancelProgressInvoice",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.5.result.invoice_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 105: Reverse Progress Revenue Entry (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseProgressRevenue",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.5.result.invoice_id",
					"boqID":     "$.input.boq_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 106: Mark Progress Invoice Not Sent (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "notification",
				HandlerMethod: "MarkInvoiceNotSent",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.5.result.invoice_id",
				},
				TimeoutSeconds: 10,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ProgressBillingSaga) SagaType() string {
	return "SAGA-PR02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ProgressBillingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ProgressBillingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ProgressBillingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["boq_id"] == nil {
		return errors.New("boq_id is required")
	}

	if inputMap["billing_cycle"] == nil {
		return errors.New("billing_cycle is required")
	}

	if inputMap["progress_percentage"] == nil {
		return errors.New("progress_percentage is required")
	}

	if inputMap["measured_date"] == nil {
		return errors.New("measured_date is required")
	}

	return nil
}
