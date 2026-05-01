// Package finance provides saga handlers for finance module workflows
package finance

import (
	"errors"
	"fmt"

	"p9e.in/samavaya/packages/saga"
)

// MonthEndCloseSaga implements SAGA-F01: Month-End Financial Close workflow
// CRITICAL: This is a SPECIAL saga with NO compensation steps (CompensationSteps: [] for ALL steps)
// This is a one-way, irreversible critical process - once started, it cannot be undone
// Business Flow: 12 sequential steps that lock, validate, accrue, reconcile, allocate, post, and finalize
// Timeout: 540s aggregate (12 steps with 30-60s each)
type MonthEndCloseSaga struct {
	steps []*saga.StepDefinition
}

// NewMonthEndCloseSaga creates a new Month-End Financial Close saga handler
func NewMonthEndCloseSaga() saga.SagaHandler {
	return &MonthEndCloseSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Lock Period - CRITICAL
			{
				StepNumber:    1,
				ServiceName:   "general-ledger",
				HandlerMethod: "LockPeriod",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"periodID":      "$.input.period_id",
					"closeDate":     "$.input.close_date",
					"closingMonth":  "$.input.closing_month",
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
			// Step 2: Validation Step (validates GL completeness) - CRITICAL
			{
				StepNumber:    2,
				ServiceName:   "general-ledger",
				HandlerMethod: "ValidationStep",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"periodID":      "$.input.period_id",
					"lockAcquired":  "$.steps.1.result.lock_acquired",
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
			// Step 3: Accrue Receivables (AR) - CRITICAL
			{
				StepNumber:    3,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "AccrueReceivables",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"periodID":           "$.input.period_id",
					"closeDate":          "$.input.close_date",
					"validationStatus":   "$.steps.2.result.validation_status",
				},
				TimeoutSeconds: 45,
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
			// Step 4: Accrue Payables (AP) - CRITICAL
			{
				StepNumber:    4,
				ServiceName:   "accounts-payable",
				HandlerMethod: "AccruePayables",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"periodID":           "$.input.period_id",
					"closeDate":          "$.input.close_date",
					"arAccrualAmount":    "$.steps.3.result.accrual_amount",
				},
				TimeoutSeconds: 45,
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
			// Step 5: Reconcile All Bank Accounts (Final) - CRITICAL
			{
				StepNumber:    5,
				ServiceName:   "banking",
				HandlerMethod: "ReconcileAllBankAccounts",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"periodID":           "$.input.period_id",
					"closeDate":          "$.input.close_date",
					"apAccrualAmount":    "$.steps.4.result.accrual_amount",
				},
				TimeoutSeconds: 60,
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
			// Step 6: Adjust Inventory Values - CRITICAL
			{
				StepNumber:    6,
				ServiceName:   "inventory",
				HandlerMethod: "AdjustInventoryValues",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"periodID":              "$.input.period_id",
					"closeDate":             "$.input.close_date",
					"bankReconciliationID":  "$.steps.5.result.reconciliation_id",
				},
				TimeoutSeconds: 45,
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
			// Step 7: Allocate Overhead Costs - CRITICAL
			{
				StepNumber:    7,
				ServiceName:   "cost-center",
				HandlerMethod: "AllocateOverheadCosts",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"periodID":             "$.input.period_id",
					"closingMonth":         "$.input.closing_month",
					"inventoryAdjustments": "$.steps.6.result.adjustment_summary",
				},
				TimeoutSeconds: 45,
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
			// Step 8: Post Journal Entries - CRITICAL
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostJournalEntries",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"periodID":            "$.input.period_id",
					"closeDate":           "$.input.close_date",
					"overheadAllocations": "$.steps.7.result.allocation_entries",
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
			// Step 9: Calculate Taxes - NON-CRITICAL
			{
				StepNumber:    9,
				ServiceName:   "tax-engine",
				HandlerMethod: "CalculateTaxes",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"periodID":          "$.input.period_id",
					"closingMonth":      "$.input.closing_month",
					"journalEntries":    "$.steps.8.result.posted_entries",
				},
				TimeoutSeconds: 45,
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
			// Step 10: Generate Financial Statements - NON-CRITICAL
			{
				StepNumber:    10,
				ServiceName:   "financial-close",
				HandlerMethod: "GenerateFinancialStatements",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"periodID":        "$.input.period_id",
					"closingMonth":    "$.input.closing_month",
					"taxCalculations": "$.steps.9.result.tax_summary",
				},
				TimeoutSeconds: 60,
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
			// Step 11: Log Closing Process (Audit) - CRITICAL
			{
				StepNumber:    11,
				ServiceName:   "audit",
				HandlerMethod: "LogClosingProcess",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"periodID":           "$.input.period_id",
					"closeDate":          "$.input.close_date",
					"financialStatements": "$.steps.10.result.statement_ids",
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
			// Step 12: Finalize Closing (Mark period closed) - CRITICAL
			{
				StepNumber:    12,
				ServiceName:   "general-ledger",
				HandlerMethod: "FinalizeClosing",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"periodID":        "$.input.period_id",
					"closeDate":       "$.input.close_date",
					"closingMonth":    "$.input.closing_month",
					"auditTrailID":    "$.steps.11.result.audit_trail_id",
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
		},
	}
}

// SagaType returns the saga type identifier
func (s *MonthEndCloseSaga) SagaType() string {
	return "SAGA-F01"
}

// GetStepDefinitions returns all step definitions (forward steps ONLY, NO compensation)
func (s *MonthEndCloseSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *MonthEndCloseSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: period_id, close_date, closing_month, company_id
func (s *MonthEndCloseSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	// Extract the nested 'input' object
	innerInput, ok := inputMap["input"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing 'input' field in saga input")
	}

	// Validate period_id
	if innerInput["period_id"] == nil {
		return errors.New("missing required field: period_id")
	}
	periodID, ok := innerInput["period_id"].(string)
	if !ok || periodID == "" {
		return errors.New("period_id must be a non-empty string")
	}

	// Validate close_date
	if innerInput["close_date"] == nil {
		return errors.New("missing required field: close_date")
	}
	closeDate, ok := innerInput["close_date"].(string)
	if !ok || closeDate == "" {
		return errors.New("close_date must be a non-empty string")
	}

	// Validate closing_month
	if innerInput["closing_month"] == nil {
		return errors.New("missing required field: closing_month")
	}
	closingMonth, ok := innerInput["closing_month"].(string)
	if !ok || closingMonth == "" {
		return errors.New("closing_month must be a non-empty string")
	}

	// Validate company_id
	if innerInput["company_id"] == nil {
		return errors.New("missing required field: company_id")
	}
	companyID, ok := innerInput["company_id"].(string)
	if !ok || companyID == "" {
		return errors.New("company_id must be a non-empty string")
	}

	return nil
}
