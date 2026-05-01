// Package banking provides saga handlers for banking module workflows
package banking

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// BankReconciliationMultiSaga implements SAGA-B02: Bank Reconciliation - Multi-Bank workflow
// Business Flow: RetrieveBankStatements → MatchBankTransactions → IdentifyDiscrepancies → ClassifyBankDifferences → ReconcileAPTransactions → ReconcileARTransactions → ReviewBankGLEntries → ResolveBankItems → UpdateBankStatus → PostBankJournals → FinalizeBankReconciliation
// Timeout: 180 seconds, Critical steps: 1,2,3,4,8,11
type BankReconciliationMultiSaga struct {
	steps []*saga.StepDefinition
}

// NewBankReconciliationMultiSaga creates a new Bank Reconciliation - Multi-Bank saga handler
func NewBankReconciliationMultiSaga() saga.SagaHandler {
	return &BankReconciliationMultiSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Retrieve Bank Statements
			{
				StepNumber:    1,
				ServiceName:   "banking",
				HandlerMethod: "RetrieveBankStatements",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationDate": "$.input.reconciliation_date",
					"bankAccountID":      "$.input.bank_account_id",
					"statementPeriod":    "$.input.statement_period",
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
			// Step 2: Match Bank Transactions
			{
				StepNumber:    2,
				ServiceName:   "reconciliation",
				HandlerMethod: "MatchBankTransactions",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationDate": "$.input.reconciliation_date",
					"bankAccountID":      "$.input.bank_account_id",
					"bankStatements":     "$.steps.1.result.bank_statements",
					"statementPeriod":    "$.input.statement_period",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{102},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Identify Discrepancies
			{
				StepNumber:    3,
				ServiceName:   "banking",
				HandlerMethod: "IdentifyDiscrepancies",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"reconciliationDate":  "$.input.reconciliation_date",
					"bankAccountID":       "$.input.bank_account_id",
					"matchedTransactions": "$.steps.2.result.matched_transactions",
					"unmatchedBankItems":  "$.steps.2.result.unmatched_bank_items",
					"unmatchedGLItems":    "$.steps.2.result.unmatched_gl_items",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{103},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Classify Bank Differences
			{
				StepNumber:    4,
				ServiceName:   "cash-management",
				HandlerMethod: "ClassifyBankDifferences",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationDate": "$.input.reconciliation_date",
					"bankAccountID":      "$.input.bank_account_id",
					"discrepancies":      "$.steps.3.result.discrepancies",
					"outstandingItems":   "$.steps.3.result.outstanding_items",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Reconcile AP Transactions
			{
				StepNumber:    5,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ReconcileAPTransactions",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationDate": "$.input.reconciliation_date",
					"bankAccountID":      "$.input.bank_account_id",
					"classifiedItems":    "$.steps.4.result.classified_items",
					"statementPeriod":    "$.input.statement_period",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Reconcile AR Transactions
			{
				StepNumber:    6,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ReconcileARTransactions",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationDate": "$.input.reconciliation_date",
					"bankAccountID":      "$.input.bank_account_id",
					"classifiedItems":    "$.steps.4.result.classified_items",
					"statementPeriod":    "$.input.statement_period",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Review Bank GL Entries
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReviewBankGLEntries",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationDate": "$.input.reconciliation_date",
					"bankAccountID":      "$.input.bank_account_id",
					"classifiedItems":    "$.steps.4.result.classified_items",
					"statementPeriod":    "$.input.statement_period",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Resolve Bank Items
			{
				StepNumber:    8,
				ServiceName:   "banking",
				HandlerMethod: "ResolveBankItems",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"reconciliationDate":  "$.input.reconciliation_date",
					"bankAccountID":       "$.input.bank_account_id",
					"classifiedItems":     "$.steps.4.result.classified_items",
					"apReconciliation":    "$.steps.5.result.reconciliation_summary",
					"arReconciliation":    "$.steps.6.result.reconciliation_summary",
					"glReview":            "$.steps.7.result.review_summary",
				},
				TimeoutSeconds: 60,
				IsCritical:     true,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Update Bank Status
			{
				StepNumber:    9,
				ServiceName:   "banking",
				HandlerMethod: "UpdateBankStatus",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationDate": "$.input.reconciliation_date",
					"bankAccountID":      "$.input.bank_account_id",
					"resolvedItems":      "$.steps.8.result.resolved_items",
					"reconciliationStatus": "RESOLVED",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Post Bank Journals
			{
				StepNumber:    10,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostBankJournals",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationDate": "$.input.reconciliation_date",
					"bankAccountID":      "$.input.bank_account_id",
					"resolvedItems":      "$.steps.8.result.resolved_items",
					"journalDate":        "$.input.reconciliation_date",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{110},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 11: Finalize Bank Reconciliation
			{
				StepNumber:    11,
				ServiceName:   "cash-management",
				HandlerMethod: "FinalizeBankReconciliation",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationDate": "$.input.reconciliation_date",
					"bankAccountID":      "$.input.bank_account_id",
					"resolvedItems":      "$.steps.8.result.resolved_items",
					"glPosting":          "$.steps.10.result.journal_entries",
					"statementPeriod":    "$.input.statement_period",
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

			// Step 102: UnmatchBankTransactions (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "reconciliation",
				HandlerMethod: "UnmatchBankTransactions",
				InputMapping: map[string]string{
					"reconciliationDate": "$.input.reconciliation_date",
					"bankAccountID":      "$.input.bank_account_id",
					"matchedData":        "$.steps.2.result.matched_data",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 103: ClearDiscrepancies (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "banking",
				HandlerMethod: "ClearDiscrepancies",
				InputMapping: map[string]string{
					"reconciliationDate": "$.input.reconciliation_date",
					"bankAccountID":      "$.input.bank_account_id",
					"discrepancies":      "$.steps.3.result.discrepancies",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 104: UnclassifyBankDifferences (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "cash-management",
				HandlerMethod: "UnclassifyBankDifferences",
				InputMapping: map[string]string{
					"reconciliationDate": "$.input.reconciliation_date",
					"bankAccountID":      "$.input.bank_account_id",
					"classifiedItems":    "$.steps.4.result.classified_items",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 105: ReverseAPReconciliation (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ReverseAPReconciliation",
				InputMapping: map[string]string{
					"reconciliationDate":    "$.input.reconciliation_date",
					"bankAccountID":         "$.input.bank_account_id",
					"reconciliationSummary": "$.steps.5.result.reconciliation_summary",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 106: ReverseARReconciliation (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ReverseARReconciliation",
				InputMapping: map[string]string{
					"reconciliationDate":    "$.input.reconciliation_date",
					"bankAccountID":         "$.input.bank_account_id",
					"reconciliationSummary": "$.steps.6.result.reconciliation_summary",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 107: ReverseBankGLReview (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseBankGLReview",
				InputMapping: map[string]string{
					"reconciliationDate": "$.input.reconciliation_date",
					"bankAccountID":      "$.input.bank_account_id",
					"reviewSummary":      "$.steps.7.result.review_summary",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 108: ReverseBankItemResolution (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "banking",
				HandlerMethod: "ReverseBankItemResolution",
				InputMapping: map[string]string{
					"reconciliationDate": "$.input.reconciliation_date",
					"bankAccountID":      "$.input.bank_account_id",
					"resolvedItems":      "$.steps.8.result.resolved_items",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 109: ReverseBankStatusUpdate (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "banking",
				HandlerMethod: "ReverseBankStatusUpdate",
				InputMapping: map[string]string{
					"reconciliationDate": "$.input.reconciliation_date",
					"bankAccountID":      "$.input.bank_account_id",
					"previousStatus":     "PENDING",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 110: ReverseBankJournals (compensates step 10)
			{
				StepNumber:    110,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseBankJournals",
				InputMapping: map[string]string{
					"reconciliationDate": "$.input.reconciliation_date",
					"bankAccountID":      "$.input.bank_account_id",
					"journalEntries":     "$.steps.10.result.journal_entries",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *BankReconciliationMultiSaga) SagaType() string {
	return "SAGA-B02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *BankReconciliationMultiSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *BankReconciliationMultiSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *BankReconciliationMultiSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["reconciliation_date"] == nil {
		return errors.New("reconciliation_date is required")
	}

	reconciliationDate, ok := inputMap["reconciliation_date"].(string)
	if !ok || reconciliationDate == "" {
		return errors.New("reconciliation_date must be a non-empty string")
	}

	if inputMap["bank_account_id"] == nil {
		return errors.New("bank_account_id is required")
	}

	bankAccountID, ok := inputMap["bank_account_id"].(string)
	if !ok || bankAccountID == "" {
		return errors.New("bank_account_id must be a non-empty string")
	}

	if inputMap["statement_period"] == nil {
		return errors.New("statement_period is required")
	}

	statementPeriod, ok := inputMap["statement_period"].(string)
	if !ok || statementPeriod == "" {
		return errors.New("statement_period must be a non-empty string")
	}

	return nil
}
