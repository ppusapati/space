// Package finance provides saga handlers for finance module workflows
package finance

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// MultiCurrencyRevaluationSaga implements SAGA-F03: Multi-Currency Revaluation workflow
// Business Flow: FetchExchangeRates → RevaluateAR → RevaluateAP → RevaluateBankBalances → UpdateIntraGroupAmounts → CreateGainLossEntries → CreateUnrealizedGainEntries → PostRevaluationEntries → StoreExchangeRateSnapshot → ScheduleAutoReversal
// Timeout: 120 seconds, Critical steps: 1,2,3,4,5,10
type MultiCurrencyRevaluationSaga struct {
	steps []*saga.StepDefinition
}

// NewMultiCurrencyRevaluationSaga creates a new Multi-Currency Revaluation saga handler
func NewMultiCurrencyRevaluationSaga() saga.SagaHandler {
	return &MultiCurrencyRevaluationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Fetch Exchange Rates
			{
				StepNumber:    1,
				ServiceName:   "currency",
				HandlerMethod: "FetchExchangeRates",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"revaluationID":     "$.input.revaluation_id",
					"revaluationDate":   "$.input.revaluation_date",
					"currencyList":      "$.input.currency_list",
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
			// Step 2: Revaluate Accounts Receivable
			{
				StepNumber:    2,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "RevaluateAR",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"revaluationID":     "$.input.revaluation_id",
					"revaluationDate":   "$.input.revaluation_date",
					"exchangeRates":     "$.steps.1.result.exchange_rates",
					"currencyList":      "$.input.currency_list",
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
			// Step 3: Revaluate Accounts Payable
			{
				StepNumber:    3,
				ServiceName:   "accounts-payable",
				HandlerMethod: "RevaluateAP",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"revaluationID":     "$.input.revaluation_id",
					"revaluationDate":   "$.input.revaluation_date",
					"exchangeRates":     "$.steps.1.result.exchange_rates",
					"currencyList":      "$.input.currency_list",
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
			// Step 4: Revaluate Bank Balances
			{
				StepNumber:    4,
				ServiceName:   "banking",
				HandlerMethod: "RevaluateBankBalances",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"revaluationID":     "$.input.revaluation_id",
					"revaluationDate":   "$.input.revaluation_date",
					"exchangeRates":     "$.steps.1.result.exchange_rates",
					"currencyList":      "$.input.currency_list",
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
			// Step 5: Update Intra-Group Amounts
			{
				StepNumber:    5,
				ServiceName:   "cost-center",
				HandlerMethod: "UpdateIntraGroupAmounts",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"revaluationID":     "$.input.revaluation_id",
					"revaluationDate":   "$.input.revaluation_date",
					"exchangeRates":     "$.steps.1.result.exchange_rates",
					"arRevaluation":     "$.steps.2.result.revaluation_data",
					"apRevaluation":     "$.steps.3.result.revaluation_data",
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
				CompensationSteps: []int32{105},
			},
			// Step 6: Create Gain/Loss Entries
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "CreateGainLossEntries",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"revaluationID":     "$.input.revaluation_id",
					"revaluationDate":   "$.input.revaluation_date",
					"arRevaluation":     "$.steps.2.result.revaluation_data",
					"apRevaluation":     "$.steps.3.result.revaluation_data",
					"bankRevaluation":   "$.steps.4.result.revaluation_data",
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
			// Step 7: Create Unrealized Gain Entries
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "CreateUnrealizedGainEntries",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"revaluationID":     "$.input.revaluation_id",
					"revaluationDate":   "$.input.revaluation_date",
					"exchangeRates":     "$.steps.1.result.exchange_rates",
					"gainLossEntries":   "$.steps.6.result.gain_loss_entries",
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
			// Step 8: Post Revaluation Entries to GL
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostRevaluationEntries",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"revaluationID":         "$.input.revaluation_id",
					"revaluationDate":       "$.input.revaluation_date",
					"gainLossEntries":       "$.steps.6.result.gain_loss_entries",
					"unrealizedGainEntries": "$.steps.7.result.unrealized_gain_entries",
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
				CompensationSteps: []int32{108},
			},
			// Step 9: Store Exchange Rate Snapshot
			{
				StepNumber:    9,
				ServiceName:   "currency",
				HandlerMethod: "StoreExchangeRateSnapshot",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"revaluationID":     "$.input.revaluation_id",
					"revaluationDate":   "$.input.revaluation_date",
					"exchangeRates":     "$.steps.1.result.exchange_rates",
					"currencyList":      "$.input.currency_list",
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
			// Step 10: Schedule Auto-Reversal
			{
				StepNumber:    10,
				ServiceName:   "currency",
				HandlerMethod: "ScheduleAutoReversal",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"revaluationID":     "$.input.revaluation_id",
					"revaluationDate":   "$.input.revaluation_date",
					"journalEntries":    "$.steps.8.result.journal_entries",
					"reverralScheduleDate": "$.input.revaluation_date",
				},
				TimeoutSeconds: 60,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        5,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},
			// ===== COMPENSATION STEPS =====

			// Step 102: ReverseARRevaluation (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ReverseARRevaluation",
				InputMapping: map[string]string{
					"revaluationID":   "$.input.revaluation_id",
					"revaluationData": "$.steps.2.result.revaluation_data",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: ReverseAPRevaluation (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ReverseAPRevaluation",
				InputMapping: map[string]string{
					"revaluationID":   "$.input.revaluation_id",
					"revaluationData": "$.steps.3.result.revaluation_data",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: ReverseBankBalanceRevaluation (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "banking",
				HandlerMethod: "ReverseBankBalanceRevaluation",
				InputMapping: map[string]string{
					"revaluationID":   "$.input.revaluation_id",
					"revaluationData": "$.steps.4.result.revaluation_data",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: ReverseIntraGroupAmountUpdate (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "cost-center",
				HandlerMethod: "ReverseIntraGroupAmountUpdate",
				InputMapping: map[string]string{
					"revaluationID": "$.input.revaluation_id",
					"updatedAmounts": "$.steps.5.result.updated_amounts",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: ReverseGainLossEntries (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseGainLossEntries",
				InputMapping: map[string]string{
					"revaluationID":  "$.input.revaluation_id",
					"gainLossEntries": "$.steps.6.result.gain_loss_entries",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: ReverseUnrealizedGainEntries (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseUnrealizedGainEntries",
				InputMapping: map[string]string{
					"revaluationID":       "$.input.revaluation_id",
					"unrealizedGainEntries": "$.steps.7.result.unrealized_gain_entries",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: ReverseRevaluationEntries (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseRevaluationEntries",
				InputMapping: map[string]string{
					"revaluationID":   "$.input.revaluation_id",
					"journalEntries":  "$.steps.8.result.journal_entries",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 109: ClearExchangeRateSnapshot (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "currency",
				HandlerMethod: "ClearExchangeRateSnapshot",
				InputMapping: map[string]string{
					"revaluationID": "$.input.revaluation_id",
					"snapshotData":  "$.steps.9.result.snapshot_data",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *MultiCurrencyRevaluationSaga) SagaType() string {
	return "SAGA-F03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *MultiCurrencyRevaluationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *MultiCurrencyRevaluationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *MultiCurrencyRevaluationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["revaluation_id"] == nil {
		return errors.New("revaluation_id is required")
	}

	revaluationID, ok := inputMap["revaluation_id"].(string)
	if !ok || revaluationID == "" {
		return errors.New("revaluation_id must be a non-empty string")
	}

	if inputMap["revaluation_date"] == nil {
		return errors.New("revaluation_date is required")
	}

	revaluationDate, ok := inputMap["revaluation_date"].(string)
	if !ok || revaluationDate == "" {
		return errors.New("revaluation_date must be a non-empty string")
	}

	if inputMap["currency_list"] == nil {
		return errors.New("currency_list is required")
	}

	currencyList, ok := inputMap["currency_list"].([]interface{})
	if !ok || len(currencyList) == 0 {
		return errors.New("currency_list must be a non-empty array")
	}

	return nil
}
