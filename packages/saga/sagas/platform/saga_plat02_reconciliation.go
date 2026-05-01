// Package platform provides saga handlers for platform module workflows
package platform

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// CrossModuleReconciliationSaga implements SAGA-PLAT02: Cross-Module Reconciliation & Validation workflow
// Business Flow: ExtractGLControlTotals → ExtractARSubledger → ExtractAPSubledger → ExtractInventoryValuation →
// CompareGLvsAR → CompareGLvsAP → CompareGLvsInventory → ReconcileVariances → CreateReconciliationReport → PostReconciliationJE
// Purpose: Reconcile data across modules (GL to subledgers), validate consistency
type CrossModuleReconciliationSaga struct {
	steps []*saga.StepDefinition
}

// NewCrossModuleReconciliationSaga creates a new Cross-Module Reconciliation saga handler
func NewCrossModuleReconciliationSaga() saga.SagaHandler {
	return &CrossModuleReconciliationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Extract GL Control Totals by Account
			{
				StepNumber:    1,
				ServiceName:   "general-ledger",
				HandlerMethod: "ExtractGLControlTotals",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"periodStart":    "$.input.period_start",
					"periodEnd":      "$.input.period_end",
					"accountTypes":   "$.input.account_types",
					"includeDetail":  "$.input.include_detail_level",
				},
				TimeoutSeconds: 90,
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
			// Step 2: Extract AR Subledger Totals
			{
				StepNumber:    2,
				ServiceName:   "sales-invoice",
				HandlerMethod: "ExtractARSubledgerTotals",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"periodStart":   "$.input.period_start",
					"periodEnd":     "$.input.period_end",
					"arAccountCodes": "$.input.ar_account_codes",
					"customerFilter": "$.input.customer_filter",
				},
				TimeoutSeconds: 75,
				IsCritical:     true,
				CompensationSteps: []int32{110},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Extract AP Subledger Totals
			{
				StepNumber:    3,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "ExtractAPSubledgerTotals",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"periodStart":    "$.input.period_start",
					"periodEnd":      "$.input.period_end",
					"apAccountCodes": "$.input.ap_account_codes",
					"vendorFilter":   "$.input.vendor_filter",
				},
				TimeoutSeconds: 75,
				IsCritical:     true,
				CompensationSteps: []int32{111},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Extract Inventory Valuation from Subledger
			{
				StepNumber:    4,
				ServiceName:   "inventory-core",
				HandlerMethod: "ExtractInventoryValuation",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"periodStart":         "$.input.period_start",
					"periodEnd":           "$.input.period_end",
					"inventoryAccountCodes": "$.input.inventory_account_codes",
					"valuationMethod":     "$.input.valuation_method",
				},
				TimeoutSeconds: 90,
				IsCritical:     true,
				CompensationSteps: []int32{112},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Compare GL vs AR (identify variance)
			{
				StepNumber:    5,
				ServiceName:   "reconciliation",
				HandlerMethod: "CompareGLvsAR",
				InputMapping: map[string]string{
					"glTotals":        "$.steps.1.result.gl_totals",
					"arSubledger":     "$.steps.2.result.ar_subledger",
					"toleranceAmount": "$.input.tolerance_amount",
					"tolerancePercent": "$.input.tolerance_percent",
				},
				TimeoutSeconds: 60,
				IsCritical:     true,
				CompensationSteps: []int32{113},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Compare GL vs AP (identify variance)
			{
				StepNumber:    6,
				ServiceName:   "reconciliation",
				HandlerMethod: "CompareGLvsAP",
				InputMapping: map[string]string{
					"glTotals":        "$.steps.1.result.gl_totals",
					"apSubledger":     "$.steps.3.result.ap_subledger",
					"toleranceAmount": "$.input.tolerance_amount",
					"tolerancePercent": "$.input.tolerance_percent",
				},
				TimeoutSeconds: 60,
				IsCritical:     true,
				CompensationSteps: []int32{114},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Compare GL vs Inventory (identify variance)
			{
				StepNumber:    7,
				ServiceName:   "reconciliation",
				HandlerMethod: "CompareGLvsInventory",
				InputMapping: map[string]string{
					"glTotals":        "$.steps.1.result.gl_totals",
					"inventoryValuation": "$.steps.4.result.inventory_valuation",
					"toleranceAmount": "$.input.tolerance_amount",
					"tolerancePercent": "$.input.tolerance_percent",
				},
				TimeoutSeconds: 75,
				IsCritical:     true,
				CompensationSteps: []int32{115},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Reconcile Variances (reclassifications, timing differences)
			{
				StepNumber:    8,
				ServiceName:   "reconciliation",
				HandlerMethod: "ReconcileVariances",
				InputMapping: map[string]string{
					"arVariance":      "$.steps.5.result.variance",
					"apVariance":      "$.steps.6.result.variance",
					"inventoryVariance": "$.steps.7.result.variance",
					"reconciliationMethod": "$.input.reconciliation_method",
					"treatmentRules":  "$.input.variance_treatment_rules",
				},
				TimeoutSeconds: 90,
				IsCritical:     false,
				CompensationSteps: []int32{116},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Create Reconciliation Report
			{
				StepNumber:    9,
				ServiceName:   "reconciliation",
				HandlerMethod: "CreateReconciliationReport",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"glTotals":        "$.steps.1.result.gl_totals",
					"arVariance":      "$.steps.5.result.variance",
					"apVariance":      "$.steps.6.result.variance",
					"inventoryVariance": "$.steps.7.result.variance",
					"reconciliationStatus": "$.steps.8.result.reconciliation_status",
					"periodStart":     "$.input.period_start",
					"periodEnd":       "$.input.period_end",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
				CompensationSteps: []int32{117},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Post Reconciliation Journal Entry (if needed)
			{
				StepNumber:    10,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostReconciliationJournal",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"arVariance":      "$.steps.5.result.variance",
					"apVariance":      "$.steps.6.result.variance",
					"inventoryVariance": "$.steps.7.result.variance",
					"journalDate":     "$.input.journal_date",
					"postIfVariance": "$.input.post_reconciliation_entries",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{118},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// ===== COMPENSATION STEPS =====

			// Step 110: Revert AR Extraction (compensates step 2)
			{
				StepNumber:    110,
				ServiceName:   "sales-invoice",
				HandlerMethod: "RevertARSubledgerExtraction",
				InputMapping: map[string]string{
					"tenantID":  "$.tenantID",
					"companyID": "$.companyID",
					"branchID":  "$.branchID",
					"reason":    "Saga compensation - reconciliation failed",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 111: Revert AP Extraction (compensates step 3)
			{
				StepNumber:    111,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "RevertAPSubledgerExtraction",
				InputMapping: map[string]string{
					"tenantID":  "$.tenantID",
					"companyID": "$.companyID",
					"branchID":  "$.branchID",
					"reason":    "Saga compensation - reconciliation failed",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 112: Revert Inventory Extraction (compensates step 4)
			{
				StepNumber:    112,
				ServiceName:   "inventory-core",
				HandlerMethod: "RevertInventoryValuationExtraction",
				InputMapping: map[string]string{
					"tenantID":  "$.tenantID",
					"companyID": "$.companyID",
					"branchID":  "$.branchID",
					"reason":    "Saga compensation - reconciliation failed",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 113: Clear GL vs AR Comparison (compensates step 5)
			{
				StepNumber:    113,
				ServiceName:   "reconciliation",
				HandlerMethod: "ClearGLvsARComparison",
				InputMapping: map[string]string{
					"tenantID":  "$.tenantID",
					"companyID": "$.companyID",
					"branchID":  "$.branchID",
					"reason":    "Saga compensation - reconciliation aborted",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 114: Clear GL vs AP Comparison (compensates step 6)
			{
				StepNumber:    114,
				ServiceName:   "reconciliation",
				HandlerMethod: "ClearGLvsAPComparison",
				InputMapping: map[string]string{
					"tenantID":  "$.tenantID",
					"companyID": "$.companyID",
					"branchID":  "$.branchID",
					"reason":    "Saga compensation - reconciliation aborted",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 115: Clear GL vs Inventory Comparison (compensates step 7)
			{
				StepNumber:    115,
				ServiceName:   "reconciliation",
				HandlerMethod: "ClearGLvsInventoryComparison",
				InputMapping: map[string]string{
					"tenantID":  "$.tenantID",
					"companyID": "$.companyID",
					"branchID":  "$.branchID",
					"reason":    "Saga compensation - reconciliation aborted",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 116: Revert Variance Reconciliation (compensates step 8)
			{
				StepNumber:    116,
				ServiceName:   "reconciliation",
				HandlerMethod: "RevertVarianceReconciliation",
				InputMapping: map[string]string{
					"tenantID":  "$.tenantID",
					"companyID": "$.companyID",
					"branchID":  "$.branchID",
					"reason":    "Saga compensation - reconciliation failed",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 117: Delete Reconciliation Report (compensates step 9)
			{
				StepNumber:    117,
				ServiceName:   "reconciliation",
				HandlerMethod: "DeleteReconciliationReport",
				InputMapping: map[string]string{
					"tenantID":  "$.tenantID",
					"companyID": "$.companyID",
					"branchID":  "$.branchID",
					"reason":    "Saga compensation - reconciliation aborted",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 118: Reverse Reconciliation Journal Entry (compensates step 10)
			{
				StepNumber:    118,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseReconciliationJournal",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"companyID":   "$.companyID",
					"branchID":    "$.branchID",
					"journalDate": "$.input.journal_date",
					"reason":      "Saga compensation - reconciliation cancelled",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *CrossModuleReconciliationSaga) SagaType() string {
	return "SAGA-PLAT02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *CrossModuleReconciliationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *CrossModuleReconciliationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *CrossModuleReconciliationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["period_start"] == nil {
		return errors.New("period_start is required (YYYY-MM-DD format)")
	}

	if inputMap["period_end"] == nil {
		return errors.New("period_end is required (YYYY-MM-DD format)")
	}

	if inputMap["account_types"] == nil {
		return errors.New("account_types is required")
	}

	accountTypes, ok := inputMap["account_types"].([]interface{})
	if !ok || len(accountTypes) == 0 {
		return errors.New("account_types must be a non-empty list")
	}

	if inputMap["ar_account_codes"] == nil {
		return errors.New("ar_account_codes is required")
	}

	arAccounts, ok := inputMap["ar_account_codes"].([]interface{})
	if !ok || len(arAccounts) == 0 {
		return errors.New("ar_account_codes must be a non-empty list")
	}

	if inputMap["ap_account_codes"] == nil {
		return errors.New("ap_account_codes is required")
	}

	apAccounts, ok := inputMap["ap_account_codes"].([]interface{})
	if !ok || len(apAccounts) == 0 {
		return errors.New("ap_account_codes must be a non-empty list")
	}

	if inputMap["inventory_account_codes"] == nil {
		return errors.New("inventory_account_codes is required")
	}

	invAccounts, ok := inputMap["inventory_account_codes"].([]interface{})
	if !ok || len(invAccounts) == 0 {
		return errors.New("inventory_account_codes must be a non-empty list")
	}

	if inputMap["tolerance_amount"] == nil {
		return errors.New("tolerance_amount is required (numeric value)")
	}

	toleranceAmount, ok := inputMap["tolerance_amount"].(float64)
	if !ok || toleranceAmount < 0 {
		return errors.New("tolerance_amount must be a non-negative number")
	}

	if inputMap["tolerance_percent"] == nil {
		return errors.New("tolerance_percent is required (0-100)")
	}

	tolerancePercent, ok := inputMap["tolerance_percent"].(float64)
	if !ok || tolerancePercent < 0 || tolerancePercent > 100 {
		return errors.New("tolerance_percent must be a number between 0 and 100")
	}

	if inputMap["journal_date"] == nil {
		return errors.New("journal_date is required (YYYY-MM-DD format)")
	}

	if inputMap["post_reconciliation_entries"] == nil {
		return errors.New("post_reconciliation_entries is required (boolean)")
	}

	if inputMap["reconciliation_method"] == nil {
		return errors.New("reconciliation_method is required (e.g., THREE_WAY, FOUR_WAY)")
	}

	return nil
}
