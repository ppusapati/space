// Package finance provides saga handlers for finance module workflows
package finance

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// BankReconciliationSaga implements SAGA-F02: Bank Reconciliation workflow
// Business Flow: RetrieveBankStatement → MatchTransactions → IdentifyDiscrepancies → ClassifyDifferences → ReconcileVendorPayments → ReconcileCustomerReceipts → ReviewGLEntries → ResolveOutstandingItems → UpdateBankReconciliationStatus → PostReconciliationEntries → FinalizeReconciliation
// Timeout: 180 seconds, Critical steps: 1,2,3,4,8,11
type BankReconciliationSaga struct {
	steps []*saga.StepDefinition
}

// NewBankReconciliationSaga creates a new Bank Reconciliation saga handler
func NewBankReconciliationSaga() saga.SagaHandler {
	return &BankReconciliationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Retrieve Bank Statement
			{
				StepNumber:    1,
				ServiceName:   "banking",
				HandlerMethod: "RetrieveBankStatement",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationID":   "$.input.reconciliation_id",
					"bankAccountID":      "$.input.bank_account_id",
					"startDate":          "$.input.start_date",
					"endDate":            "$.input.end_date",
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
			// Step 2: Match Transactions
			{
				StepNumber:    2,
				ServiceName:   "reconciliation",
				HandlerMethod: "MatchTransactions",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationID":   "$.input.reconciliation_id",
					"bankAccountID":      "$.input.bank_account_id",
					"bankStatement":      "$.steps.1.result.bank_statement",
					"startDate":          "$.input.start_date",
					"endDate":            "$.input.end_date",
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
				CompensationSteps: []int32{102},
			},
			// Step 3: Identify Discrepancies
			{
				StepNumber:    3,
				ServiceName:   "banking",
				HandlerMethod: "IdentifyDiscrepancies",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationID":   "$.input.reconciliation_id",
					"matchedTransactions": "$.steps.2.result.matched_transactions",
					"unmatchedBankItems": "$.steps.2.result.unmatched_bank_items",
					"unmatchedGLItems":   "$.steps.2.result.unmatched_gl_items",
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
				CompensationSteps: []int32{103},
			},
			// Step 4: Classify Differences
			{
				StepNumber:    4,
				ServiceName:   "reconciliation",
				HandlerMethod: "ClassifyDifferences",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationID":   "$.input.reconciliation_id",
					"discrepancies":      "$.steps.3.result.discrepancies",
					"outstandingItems":   "$.steps.3.result.outstanding_items",
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
				CompensationSteps: []int32{104},
			},
			// Step 5: Reconcile Vendor Payments
			{
				StepNumber:    5,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ReconcileVendorPayments",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationID":   "$.input.reconciliation_id",
					"bankAccountID":      "$.input.bank_account_id",
					"classifiedItems":    "$.steps.4.result.classified_items",
					"startDate":          "$.input.start_date",
					"endDate":            "$.input.end_date",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{105},
			},
			// Step 6: Reconcile Customer Receipts
			{
				StepNumber:    6,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ReconcileCustomerReceipts",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationID":   "$.input.reconciliation_id",
					"bankAccountID":      "$.input.bank_account_id",
					"classifiedItems":    "$.steps.4.result.classified_items",
					"startDate":          "$.input.start_date",
					"endDate":            "$.input.end_date",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{106},
			},
			// Step 7: Review GL Entries
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReviewGLEntries",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationID":   "$.input.reconciliation_id",
					"bankAccountID":      "$.input.bank_account_id",
					"classifiedItems":    "$.steps.4.result.classified_items",
					"startDate":          "$.input.start_date",
					"endDate":            "$.input.end_date",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{107},
			},
			// Step 8: Resolve Outstanding Items
			{
				StepNumber:    8,
				ServiceName:   "banking",
				HandlerMethod: "ResolveOutstandingItems",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationID":   "$.input.reconciliation_id",
					"bankAccountID":      "$.input.bank_account_id",
					"classifiedItems":    "$.steps.4.result.classified_items",
					"vendorReconciliation": "$.steps.5.result.reconciliation_summary",
					"customerReconciliation": "$.steps.6.result.reconciliation_summary",
					"glReview":           "$.steps.7.result.review_summary",
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
				CompensationSteps: []int32{108},
			},
			// Step 9: Update Bank Reconciliation Status
			{
				StepNumber:    9,
				ServiceName:   "banking",
				HandlerMethod: "UpdateBankReconciliationStatus",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationID":   "$.input.reconciliation_id",
					"bankAccountID":      "$.input.bank_account_id",
					"resolvedItems":      "$.steps.8.result.resolved_items",
					"reconciliationStatus": "RESOLVED",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{109},
			},
			// Step 10: Post Reconciliation Entries to GL
			{
				StepNumber:    10,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostReconciliationEntries",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationID":   "$.input.reconciliation_id",
					"bankAccountID":      "$.input.bank_account_id",
					"resolvedItems":      "$.steps.8.result.resolved_items",
					"journalDate":        "$.input.end_date",
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
				CompensationSteps: []int32{110},
			},
			// Step 11: Finalize Reconciliation
			{
				StepNumber:    11,
				ServiceName:   "reconciliation",
				HandlerMethod: "FinalizeReconciliation",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationID":   "$.input.reconciliation_id",
					"bankAccountID":      "$.input.bank_account_id",
					"resolvedItems":      "$.steps.8.result.resolved_items",
					"glPosting":          "$.steps.10.result.journal_entries",
					"reconciliationDate": "$.input.end_date",
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
			// ===== COMPENSATION STEPS =====

			// Step 102: Unmatch Transactions (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "reconciliation",
				HandlerMethod: "UnmatchTransactions",
				InputMapping: map[string]string{
					"reconciliationID": "$.input.reconciliation_id",
					"matchedData":      "$.steps.2.result.matched_data",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Clear Discrepancies (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "banking",
				HandlerMethod: "ClearDiscrepancies",
				InputMapping: map[string]string{
					"reconciliationID": "$.input.reconciliation_id",
					"discrepancies":    "$.steps.3.result.discrepancies",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: UnclassifyDifferences (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "reconciliation",
				HandlerMethod: "UnclassifyDifferences",
				InputMapping: map[string]string{
					"reconciliationID":  "$.input.reconciliation_id",
					"classifiedItems":   "$.steps.4.result.classified_items",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: ReverseVendorPaymentReconciliation (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ReverseVendorPaymentReconciliation",
				InputMapping: map[string]string{
					"reconciliationID":    "$.input.reconciliation_id",
					"reconciliationSummary": "$.steps.5.result.reconciliation_summary",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: ReverseCustomerReceiptReconciliation (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ReverseCustomerReceiptReconciliation",
				InputMapping: map[string]string{
					"reconciliationID":    "$.input.reconciliation_id",
					"reconciliationSummary": "$.steps.6.result.reconciliation_summary",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: ReverseGLReview (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseGLReview",
				InputMapping: map[string]string{
					"reconciliationID": "$.input.reconciliation_id",
					"reviewSummary":    "$.steps.7.result.review_summary",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: ReverseOutstandingItemResolution (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "banking",
				HandlerMethod: "ReverseOutstandingItemResolution",
				InputMapping: map[string]string{
					"reconciliationID": "$.input.reconciliation_id",
					"resolvedItems":    "$.steps.8.result.resolved_items",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 109: ReverseReconciliationStatusUpdate (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "banking",
				HandlerMethod: "ReverseReconciliationStatusUpdate",
				InputMapping: map[string]string{
					"reconciliationID": "$.input.reconciliation_id",
					"previousStatus":   "PENDING",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 110: ReverseReconciliationEntries (compensates step 10)
			{
				StepNumber:    110,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseReconciliationEntries",
				InputMapping: map[string]string{
					"reconciliationID": "$.input.reconciliation_id",
					"journalEntries":   "$.steps.10.result.journal_entries",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *BankReconciliationSaga) SagaType() string {
	return "SAGA-F02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *BankReconciliationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *BankReconciliationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *BankReconciliationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["reconciliation_id"] == nil {
		return errors.New("reconciliation_id is required")
	}

	reconciliationID, ok := inputMap["reconciliation_id"].(string)
	if !ok || reconciliationID == "" {
		return errors.New("reconciliation_id must be a non-empty string")
	}

	if inputMap["bank_account_id"] == nil {
		return errors.New("bank_account_id is required")
	}

	bankAccountID, ok := inputMap["bank_account_id"].(string)
	if !ok || bankAccountID == "" {
		return errors.New("bank_account_id must be a non-empty string")
	}

	if inputMap["start_date"] == nil {
		return errors.New("start_date is required")
	}

	startDate, ok := inputMap["start_date"].(string)
	if !ok || startDate == "" {
		return errors.New("start_date must be a non-empty string")
	}

	if inputMap["end_date"] == nil {
		return errors.New("end_date is required")
	}

	endDate, ok := inputMap["end_date"].(string)
	if !ok || endDate == "" {
		return errors.New("end_date must be a non-empty string")
	}

	return nil
}
