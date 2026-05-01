// Package projects provides saga handlers for projects module workflows
package projects

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ProjectBillingSaga implements SAGA-PR01: Project Billing (T&M)
// Business Flow: Get timesheet → Calculate billable hours → Create invoice → Post revenue → Send invoice → Record transaction → Complete billing
type ProjectBillingSaga struct {
	steps []*saga.StepDefinition
}

// NewProjectBillingSaga creates a new Project Billing saga handler
func NewProjectBillingSaga() saga.SagaHandler {
	return &ProjectBillingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Get Approved Timesheet
			{
				StepNumber:    1,
				ServiceName:   "timesheet",
				HandlerMethod: "GetApprovedTimesheet",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"timesheetID":    "$.input.timesheet_id",
					"projectID":      "$.input.project_id",
					"billingPeriod":  "$.input.billing_period",
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
			// Step 2: Calculate Billable Hours
			{
				StepNumber:    2,
				ServiceName:   "project-costing",
				HandlerMethod: "CalculateBillableAmount",
				InputMapping: map[string]string{
					"timesheetID":    "$.steps.1.result.timesheet_id",
					"projectID":      "$.input.project_id",
					"approvedHours":  "$.steps.1.result.approved_hours",
					"billingRate":    "$.steps.1.result.billing_rate",
					"billingPeriod":  "$.input.billing_period",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{101},
			},
			// Step 3: Create Project Invoice
			{
				StepNumber:    3,
				ServiceName:   "sales-invoice",
				HandlerMethod: "CreateInvoice",
				InputMapping: map[string]string{
					"projectID":       "$.input.project_id",
					"timesheetID":     "$.steps.1.result.timesheet_id",
					"billableAmount":  "$.steps.2.result.billable_amount",
					"invoiceDate":     "$.input.billing_period",
					"customerID":      "$.steps.1.result.customer_id",
					"invoiceType":     "T&M",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{102},
			},
			// Step 4: Post Revenue Recognition
			{
				StepNumber:    4,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostRevenueEntry",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"invoiceID":       "$.steps.3.result.invoice_id",
					"projectID":       "$.input.project_id",
					"revenueAmount":   "$.steps.2.result.billable_amount",
					"journalDate":     "$.input.billing_period",
					"revenueType":     "Services",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{103},
			},
			// Step 5: Send Invoice
			{
				StepNumber:    5,
				ServiceName:   "notification",
				HandlerMethod: "SendInvoice",
				InputMapping: map[string]string{
					"invoiceID":    "$.steps.3.result.invoice_id",
					"customerID":   "$.steps.1.result.customer_id",
					"customerEmail": "$.steps.1.result.customer_email",
					"invoiceNumber": "$.steps.3.result.invoice_number",
				},
				TimeoutSeconds:    10,
				IsCritical:        false,
				CompensationSteps: []int32{104},
			},
			// Step 6: Record Invoice Transaction
			{
				StepNumber:    6,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "RecordInvoiceTransaction",
				InputMapping: map[string]string{
					"invoiceID":      "$.steps.3.result.invoice_id",
					"customerID":     "$.steps.1.result.customer_id",
					"transactionAmount": "$.steps.2.result.billable_amount",
					"transactionDate":   "$.input.billing_period",
					"projectID":      "$.input.project_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        false,
				CompensationSteps: []int32{105},
			},
			// Step 7: Complete Billing Run
			{
				StepNumber:    7,
				ServiceName:   "project-costing",
				HandlerMethod: "CompleteBillingRun",
				InputMapping: map[string]string{
					"timesheetID":    "$.steps.1.result.timesheet_id",
					"projectID":      "$.input.project_id",
					"invoiceID":      "$.steps.3.result.invoice_id",
					"billingPeriod":  "$.input.billing_period",
					"billableAmount": "$.steps.2.result.billable_amount",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{},
			},
			// ===== COMPENSATION STEPS =====

			// Step 101: Reverse Billable Amount Calculation (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "project-costing",
				HandlerMethod: "ReverseBillableCalculation",
				InputMapping: map[string]string{
					"timesheetID": "$.steps.1.result.timesheet_id",
					"projectID":   "$.input.project_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 102: Cancel Invoice (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "sales-invoice",
				HandlerMethod: "CancelInvoice",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.3.result.invoice_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 103: Reverse Revenue Entry (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseRevenueEntry",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.3.result.invoice_id",
					"projectID": "$.input.project_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 104: Mark Invoice Not Sent (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "notification",
				HandlerMethod: "MarkInvoiceNotSent",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.3.result.invoice_id",
				},
				TimeoutSeconds: 10,
				IsCritical:     false,
			},
			// Step 105: Reverse Transaction Record (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ReverseTransactionRecord",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.3.result.invoice_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ProjectBillingSaga) SagaType() string {
	return "SAGA-PR01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ProjectBillingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ProjectBillingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ProjectBillingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["timesheet_id"] == nil {
		return errors.New("timesheet_id is required")
	}

	if inputMap["project_id"] == nil {
		return errors.New("project_id is required")
	}

	if inputMap["billing_period"] == nil {
		return errors.New("billing_period is required")
	}

	return nil
}
