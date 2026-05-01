// Package construction provides saga handlers for construction module workflows
package construction

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ProgressBillingSaga implements SAGA-C02: Progress Billing (Construction)
// Business Flow: MeasureProgressOnSite → CalculateProgressBillingAmount → ValidateQualityChecks → PrepareProgressBill → ApproveProgressBill → CreateBillingInvoice → PostRevenueEntry → RecordProjectCost → UpdateContractProgress → FinalizeProgressBillingCycle
// Timeout: 120 seconds, Critical steps: 1,2,3,4,7,10
type ProgressBillingSaga struct {
	steps []*saga.StepDefinition
}

// NewProgressBillingSaga creates a new Progress Billing saga handler
func NewProgressBillingSaga() saga.SagaHandler {
	return &ProgressBillingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Measure Progress On Site
			{
				StepNumber:    1,
				ServiceName:   "quality-inspection",
				HandlerMethod: "MeasureProgressOnSite",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"billingCycleID":      "$.input.billing_cycle_id",
					"projectID":           "$.input.project_id",
					"progressPercentage":  "$.input.progress_percentage",
					"measuredDate":        "$.input.measured_date",
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
			// Step 2: Calculate Progress Billing Amount
			{
				StepNumber:    2,
				ServiceName:   "construction-billing",
				HandlerMethod: "CalculateProgressBillingAmount",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"billingCycleID":     "$.input.billing_cycle_id",
					"projectID":          "$.input.project_id",
					"progressPercentage": "$.input.progress_percentage",
					"progressData":       "$.steps.1.result.progress_data",
				},
				TimeoutSeconds:    30,
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
			// Step 3: Validate Quality Checks
			{
				StepNumber:    3,
				ServiceName:   "quality-inspection",
				HandlerMethod: "ValidateQualityChecks",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"projectID":          "$.input.project_id",
					"billingCycleID":     "$.input.billing_cycle_id",
					"progressPercentage": "$.input.progress_percentage",
					"progressData":       "$.steps.1.result.progress_data",
				},
				TimeoutSeconds:    30,
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
			// Step 4: Prepare Progress Bill
			{
				StepNumber:    4,
				ServiceName:   "construction-billing",
				HandlerMethod: "PrepareProgressBill",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"billingCycleID":     "$.input.billing_cycle_id",
					"projectID":          "$.input.project_id",
					"billingAmount":      "$.steps.2.result.billing_amount",
					"qualityCheck":       "$.steps.3.result.quality_status",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Approve Progress Bill
			{
				StepNumber:    5,
				ServiceName:   "construction-billing",
				HandlerMethod: "ApproveProgressBill",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"billingCycleID": "$.input.billing_cycle_id",
					"projectID":       "$.input.project_id",
					"bill":            "$.steps.4.result.bill_details",
				},
				TimeoutSeconds:    30,
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
			// Step 6: Create Billing Invoice
			{
				StepNumber:    6,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "CreateBillingInvoice",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"billingCycleID": "$.input.billing_cycle_id",
					"projectID":       "$.input.project_id",
					"bill":            "$.steps.4.result.bill_details",
					"approval":        "$.steps.5.result.approval_status",
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
			// Step 7: Post Revenue Entry
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostRevenueEntry",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"billingCycleID": "$.input.billing_cycle_id",
					"projectID":       "$.input.project_id",
					"invoiceID":       "$.steps.6.result.invoice_id",
					"billingAmount":   "$.steps.2.result.billing_amount",
					"journalDate":     "$.input.measured_date",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Record Project Cost
			{
				StepNumber:    8,
				ServiceName:   "project-costing",
				HandlerMethod: "RecordProjectCost",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"projectID":       "$.input.project_id",
					"billingCycleID": "$.input.billing_cycle_id",
					"progressPercentage": "$.input.progress_percentage",
					"revenueEntry":    "$.steps.7.result.journal_entry",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Update Contract Progress
			{
				StepNumber:    9,
				ServiceName:   "construction-billing",
				HandlerMethod: "UpdateContractProgress",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"projectID":          "$.input.project_id",
					"billingCycleID":     "$.input.billing_cycle_id",
					"progressPercentage": "$.input.progress_percentage",
					"costData":           "$.steps.8.result.cost_data",
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
			// Step 10: Finalize Progress Billing Cycle
			{
				StepNumber:    10,
				ServiceName:   "construction-billing",
				HandlerMethod: "FinalizeProgressBillingCycle",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"billingCycleID":     "$.input.billing_cycle_id",
					"projectID":          "$.input.project_id",
					"progressPercentage": "$.input.progress_percentage",
					"invoiceID":          "$.steps.6.result.invoice_id",
					"contractProgress":   "$.steps.9.result.contract_progress",
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

			// Step 102: ReverseProgressCalculation (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "construction-billing",
				HandlerMethod: "ReverseProgressCalculation",
				InputMapping: map[string]string{
					"billingCycleID": "$.input.billing_cycle_id",
					"projectID":      "$.input.project_id",
					"billingAmount":  "$.steps.2.result.billing_amount",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: ReverseSiteProgressMeasurement (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "quality-inspection",
				HandlerMethod: "ReverseSiteProgressMeasurement",
				InputMapping: map[string]string{
					"projectID":       "$.input.project_id",
					"billingCycleID": "$.input.billing_cycle_id",
					"progressData":    "$.steps.1.result.progress_data",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: CancelProgressBill (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "construction-billing",
				HandlerMethod: "CancelProgressBill",
				InputMapping: map[string]string{
					"billingCycleID": "$.input.billing_cycle_id",
					"projectID":      "$.input.project_id",
					"bill":           "$.steps.4.result.bill_details",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: RejectProgressBill (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "construction-billing",
				HandlerMethod: "RejectProgressBill",
				InputMapping: map[string]string{
					"billingCycleID": "$.input.billing_cycle_id",
					"projectID":      "$.input.project_id",
					"approval":       "$.steps.5.result.approval_status",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: CancelBillingInvoice (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "CancelBillingInvoice",
				InputMapping: map[string]string{
					"billingCycleID": "$.input.billing_cycle_id",
					"projectID":      "$.input.project_id",
					"invoiceID":      "$.steps.6.result.invoice_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: ReverseRevenueEntry (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseRevenueEntry",
				InputMapping: map[string]string{
					"projectID":      "$.input.project_id",
					"journalEntry":   "$.steps.7.result.journal_entry",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: ReverseProjectCost (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "project-costing",
				HandlerMethod: "ReverseProjectCost",
				InputMapping: map[string]string{
					"projectID":      "$.input.project_id",
					"billingCycleID": "$.input.billing_cycle_id",
					"costData":       "$.steps.8.result.cost_data",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 109: RevertContractProgress (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "construction-billing",
				HandlerMethod: "RevertContractProgress",
				InputMapping: map[string]string{
					"projectID":           "$.input.project_id",
					"billingCycleID":     "$.input.billing_cycle_id",
					"contractProgress":   "$.steps.9.result.contract_progress",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ProgressBillingSaga) SagaType() string {
	return "SAGA-C02"
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

	if inputMap["billing_cycle_id"] == nil {
		return errors.New("billing_cycle_id is required")
	}

	billingCycleID, ok := inputMap["billing_cycle_id"].(string)
	if !ok || billingCycleID == "" {
		return errors.New("billing_cycle_id must be a non-empty string")
	}

	if inputMap["project_id"] == nil {
		return errors.New("project_id is required")
	}

	projectID, ok := inputMap["project_id"].(string)
	if !ok || projectID == "" {
		return errors.New("project_id must be a non-empty string")
	}

	if inputMap["progress_percentage"] == nil {
		return errors.New("progress_percentage is required")
	}

	progressPercentage, ok := inputMap["progress_percentage"].(string)
	if !ok || progressPercentage == "" {
		return errors.New("progress_percentage must be a non-empty string")
	}

	if inputMap["measured_date"] == nil {
		return errors.New("measured_date is required")
	}

	measuredDate, ok := inputMap["measured_date"].(string)
	if !ok || measuredDate == "" {
		return errors.New("measured_date must be a non-empty string")
	}

	return nil
}
