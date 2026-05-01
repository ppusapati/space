// Package construction provides saga handlers for construction module workflows
package construction

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ProjectClosureSaga implements SAGA-C07: Project Closure & Final Settlement
// Business Flow: ValidateProjectCompletion → PerformFinalInspection → ReconcileProjectAccounts → ProcessFinalBilling → SettleOutstandingClaims → ArchiveProjectData → GenerateClosureReport → PostClosureJournals → ReleaseBonds → FinalizeProjectClosure
// Timeout: 120 seconds, Critical steps: 1,2,3,5,8,10
type ProjectClosureSaga struct {
	steps []*saga.StepDefinition
}

// NewProjectClosureSaga creates a new Project Closure & Final Settlement saga handler
func NewProjectClosureSaga() saga.SagaHandler {
	return &ProjectClosureSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Validate Project Completion
			{
				StepNumber:    1,
				ServiceName:   "construction-project",
				HandlerMethod: "ValidateProjectCompletion",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"projectID":     "$.input.project_id",
					"closureDate":   "$.input.closure_date",
					"finalStatus":   "$.input.final_status",
					"settlementType": "$.input.settlement_type",
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
			// Step 2: Perform Final Inspection
			{
				StepNumber:    2,
				ServiceName:   "quality-inspection",
				HandlerMethod: "PerformFinalInspection",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"projectID":      "$.input.project_id",
					"closureDate":    "$.input.closure_date",
					"completionData": "$.steps.1.result.completion_data",
				},
				TimeoutSeconds:    45,
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
			// Step 3: Reconcile Project Accounts
			{
				StepNumber:    3,
				ServiceName:   "project-costing",
				HandlerMethod: "ReconcileProjectAccounts",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"projectID":      "$.input.project_id",
					"closureDate":    "$.input.closure_date",
					"finalInspection": "$.steps.2.result.final_inspection_result",
				},
				TimeoutSeconds:    45,
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
			// Step 4: Process Final Billing
			{
				StepNumber:    4,
				ServiceName:   "construction-billing",
				HandlerMethod: "ProcessFinalBilling",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"projectID":             "$.input.project_id",
					"closureDate":           "$.input.closure_date",
					"projectReconciliation": "$.steps.3.result.reconciliation_data",
				},
				TimeoutSeconds:    45,
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
			// Step 5: Settle Outstanding Claims
			{
				StepNumber:    5,
				ServiceName:   "settlement",
				HandlerMethod: "SettleOutstandingClaims",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"projectID":         "$.input.project_id",
					"settlementType":    "$.input.settlement_type",
					"finalBillingData":  "$.steps.4.result.final_billing_data",
				},
				TimeoutSeconds:    60,
				IsCritical:        true,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Archive Project Data
			{
				StepNumber:    6,
				ServiceName:   "construction-project",
				HandlerMethod: "ArchiveProjectData",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"projectID":     "$.input.project_id",
					"closureDate":   "$.input.closure_date",
					"settlement":    "$.steps.5.result.settlement_summary",
				},
				TimeoutSeconds:    30,
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
			// Step 7: Generate Closure Report
			{
				StepNumber:    7,
				ServiceName:   "construction-project",
				HandlerMethod: "GenerateClosureReport",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"projectID":        "$.input.project_id",
					"closureDate":      "$.input.closure_date",
					"finalInspection":  "$.steps.2.result.final_inspection_result",
					"reconciliation":   "$.steps.3.result.reconciliation_data",
					"settlement":       "$.steps.5.result.settlement_summary",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Post Closure Journals
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostClosureJournals",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"projectID":        "$.input.project_id",
					"closureDate":      "$.input.closure_date",
					"reconciliation":   "$.steps.3.result.reconciliation_data",
					"finalBillingData": "$.steps.4.result.final_billing_data",
				},
				TimeoutSeconds:    45,
				IsCritical:        true,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Release Bonds
			{
				StepNumber:    9,
				ServiceName:   "settlement",
				HandlerMethod: "ReleaseBonds",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"projectID":      "$.input.project_id",
					"closureDate":    "$.input.closure_date",
					"settlement":     "$.steps.5.result.settlement_summary",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Finalize Project Closure
			{
				StepNumber:    10,
				ServiceName:   "construction-project",
				HandlerMethod: "FinalizeProjectClosure",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"projectID":       "$.input.project_id",
					"closureDate":     "$.input.closure_date",
					"finalStatus":     "$.input.final_status",
					"closureReport":   "$.steps.7.result.closure_report",
					"journalPosting":  "$.steps.8.result.journal_entries",
					"bondRelease":     "$.steps.9.result.bond_release_status",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
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

			// Step 102: RevertFinalInspection (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "quality-inspection",
				HandlerMethod: "RevertFinalInspection",
				InputMapping: map[string]string{
					"projectID":                  "$.input.project_id",
					"finalInspectionResult": "$.steps.2.result.final_inspection_result",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 103: RevertReconciliation (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "project-costing",
				HandlerMethod: "RevertReconciliation",
				InputMapping: map[string]string{
					"projectID":                "$.input.project_id",
					"reconciliationData": "$.steps.3.result.reconciliation_data",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 104: ReverseFinalBilling (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "construction-billing",
				HandlerMethod: "ReverseFinalBilling",
				InputMapping: map[string]string{
					"projectID":          "$.input.project_id",
					"finalBillingData":   "$.steps.4.result.final_billing_data",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 105: ReverseSettlement (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "settlement",
				HandlerMethod: "ReverseSettlement",
				InputMapping: map[string]string{
					"projectID":           "$.input.project_id",
					"settlementSummary": "$.steps.5.result.settlement_summary",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 106: UnarchiveProjectData (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "construction-project",
				HandlerMethod: "UnarchiveProjectData",
				InputMapping: map[string]string{
					"projectID":     "$.input.project_id",
					"archivedData": "$.steps.6.result.archived_data",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: DeleteClosureReport (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "construction-project",
				HandlerMethod: "DeleteClosureReport",
				InputMapping: map[string]string{
					"projectID":      "$.input.project_id",
					"closureReport": "$.steps.7.result.closure_report",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: ReverseClosureJournals (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseClosureJournals",
				InputMapping: map[string]string{
					"projectID":     "$.input.project_id",
					"journalEntries": "$.steps.8.result.journal_entries",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 109: RetainBonds (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "settlement",
				HandlerMethod: "RetainBonds",
				InputMapping: map[string]string{
					"projectID":           "$.input.project_id",
					"bondReleaseStatus": "$.steps.9.result.bond_release_status",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ProjectClosureSaga) SagaType() string {
	return "SAGA-C07"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ProjectClosureSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ProjectClosureSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ProjectClosureSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["project_id"] == nil {
		return errors.New("project_id is required")
	}

	projectID, ok := inputMap["project_id"].(string)
	if !ok || projectID == "" {
		return errors.New("project_id must be a non-empty string")
	}

	if inputMap["closure_date"] == nil {
		return errors.New("closure_date is required")
	}

	closureDate, ok := inputMap["closure_date"].(string)
	if !ok || closureDate == "" {
		return errors.New("closure_date must be a non-empty string")
	}

	if inputMap["final_status"] == nil {
		return errors.New("final_status is required")
	}

	finalStatus, ok := inputMap["final_status"].(string)
	if !ok || finalStatus == "" {
		return errors.New("final_status must be a non-empty string")
	}

	if inputMap["settlement_type"] == nil {
		return errors.New("settlement_type is required")
	}

	settlementType, ok := inputMap["settlement_type"].(string)
	if !ok || settlementType == "" {
		return errors.New("settlement_type must be a non-empty string")
	}

	return nil
}
