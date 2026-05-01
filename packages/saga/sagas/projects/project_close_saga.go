// Package projects provides saga handlers for projects module workflows
package projects

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ProjectCloseSaga implements SAGA-PR04: Project Close & Profitability
// Business Flow: Finalize project → Calculate profitability → Close accounts → Generate final invoice → Post final GL → Get final approval → Archive project → Complete close
type ProjectCloseSaga struct {
	steps []*saga.StepDefinition
}

// NewProjectCloseSaga creates a new Project Close saga handler
func NewProjectCloseSaga() saga.SagaHandler {
	return &ProjectCloseSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Finalize Project Work
			{
				StepNumber:    1,
				ServiceName:   "project",
				HandlerMethod: "FinalizeProject",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"projectID":    "$.input.project_id",
					"closingDate":  "$.input.closing_date",
					"finalizationNotes": "$.input.finalization_notes",
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
			// Step 2: Calculate Profitability
			{
				StepNumber:    2,
				ServiceName:   "project-costing",
				HandlerMethod: "CalculateProjectProfitability",
				InputMapping: map[string]string{
					"projectID":     "$.input.project_id",
					"closingDate":   "$.input.closing_date",
					"finalAmount":   "$.input.final_amount",
					"includeOverhead": "true",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{101},
			},
			// Step 3: Close Project Accounts
			{
				StepNumber:    3,
				ServiceName:   "general-ledger",
				HandlerMethod: "CloseProjectAccounts",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"companyID":   "$.companyID",
					"branchID":    "$.branchID",
					"projectID":   "$.input.project_id",
					"closingDate": "$.input.closing_date",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{102},
			},
			// Step 4: Generate Final Invoice
			{
				StepNumber:    4,
				ServiceName:   "sales-invoice",
				HandlerMethod: "CreateFinalInvoice",
				InputMapping: map[string]string{
					"projectID":        "$.input.project_id",
					"finalAmount":      "$.input.final_amount",
					"invoiceDate":      "$.input.closing_date",
					"profitabilityData": "$.steps.2.result.profitability_summary",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{103},
			},
			// Step 5: Post Final GL Entry
			{
				StepNumber:    5,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostFinalProjectEntry",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"projectID":      "$.input.project_id",
					"finalInvoiceID": "$.steps.4.result.final_invoice_id",
					"finalAmount":    "$.input.final_amount",
					"journalDate":    "$.input.closing_date",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{104},
			},
			// Step 6: Get Final Approval
			{
				StepNumber:    6,
				ServiceName:   "approval",
				HandlerMethod: "RequestApproval",
				InputMapping: map[string]string{
					"projectID":    "$.input.project_id",
					"approverType": "ProjectManager",
					"approverID":   "$.input.approver_id",
					"approvalData": "$.steps.2.result.profitability_summary",
				},
				TimeoutSeconds:    20,
				IsCritical:        false,
				CompensationSteps: []int32{105},
			},
			// Step 7: Archive Project Data
			{
				StepNumber:    7,
				ServiceName:   "project",
				HandlerMethod: "ArchiveProject",
				InputMapping: map[string]string{
					"projectID":     "$.input.project_id",
					"archiveDate":   "$.input.closing_date",
					"retentionDays": "2555",
				},
				TimeoutSeconds:    15,
				IsCritical:        false,
				CompensationSteps: []int32{106},
			},
			// Step 8: Complete Project Close
			{
				StepNumber:    8,
				ServiceName:   "project",
				HandlerMethod: "CompleteProjectClose",
				InputMapping: map[string]string{
					"projectID":      "$.input.project_id",
					"finalInvoiceID": "$.steps.4.result.final_invoice_id",
					"closingDate":    "$.input.closing_date",
					"profitability":  "$.steps.2.result.profit_percentage",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{},
			},
			// ===== COMPENSATION STEPS =====

			// Step 101: Revert Profitability Calculation (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "project-costing",
				HandlerMethod: "RevertProfitabilityCalculation",
				InputMapping: map[string]string{
					"projectID": "$.input.project_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 102: Reopen Project Accounts (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReopenProjectAccounts",
				InputMapping: map[string]string{
					"projectID": "$.input.project_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 103: Cancel Final Invoice (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "sales-invoice",
				HandlerMethod: "CancelFinalInvoice",
				InputMapping: map[string]string{
					"finalInvoiceID": "$.steps.4.result.final_invoice_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 104: Reverse Final GL Entry (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseFinalProjectEntry",
				InputMapping: map[string]string{
					"finalInvoiceID": "$.steps.4.result.final_invoice_id",
					"projectID":      "$.input.project_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 105: Revoke Final Approval (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "approval",
				HandlerMethod: "RevokeApproval",
				InputMapping: map[string]string{
					"projectID": "$.input.project_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: Restore Archived Project (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "project",
				HandlerMethod: "RestoreArchivedProject",
				InputMapping: map[string]string{
					"projectID": "$.input.project_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ProjectCloseSaga) SagaType() string {
	return "SAGA-PR04"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ProjectCloseSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ProjectCloseSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ProjectCloseSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["project_id"] == nil {
		return errors.New("project_id is required")
	}

	if inputMap["closing_date"] == nil {
		return errors.New("closing_date is required")
	}

	if inputMap["final_amount"] == nil {
		return errors.New("final_amount is required")
	}

	return nil
}
