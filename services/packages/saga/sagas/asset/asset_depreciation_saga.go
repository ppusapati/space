// Package asset provides saga handlers for asset management workflows
package asset

import (
	"errors"
	"fmt"

	"p9e.in/samavaya/packages/saga"
)

// AssetDepreciationSaga implements SAGA-A02: Asset Depreciation (Monthly Accrual)
// Business Flow: 8 steps for monthly depreciation calculation and GL posting
// Accounting Standard: IAS 16 (Depreciation measurement)
// Critical Steps: 3, 6, 7
// Timeout: 180 seconds
// Execution: Monthly via scheduler, processes all ACTIVE assets
type AssetDepreciationSaga struct {
	steps []*saga.StepDefinition
}

// NewAssetDepreciationSaga creates a new Asset Depreciation saga handler
func NewAssetDepreciationSaga() saga.SagaHandler {
	return &AssetDepreciationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Extract Active Assets - Filter by status ACTIVE, not retired
			{
				StepNumber:    1,
				ServiceName:   "asset",
				HandlerMethod: "ExtractActiveAssets",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"depreciationDate": "$.input.depreciation_date",
					"periodID":         "$.input.period_id",
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
			// Step 2: Calculate Monthly Depreciation - Using depreciation schedule
			{
				StepNumber:    2,
				ServiceName:   "depreciation",
				HandlerMethod: "CalculateMonthlyDepreciation",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"assets":             "$.steps.1.result.active_assets",
					"depreciationDate":   "$.input.depreciation_date",
					"periodID":           "$.input.period_id",
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
				CompensationSteps: []int32{201},
			},
			// Step 3: Apply Depreciation Cap - Not below salvage value - CRITICAL
			{
				StepNumber:    3,
				ServiceName:   "fixed-assets",
				HandlerMethod: "ApplyDepreciationCap",
				InputMapping: map[string]string{
					"tenantID":                "$.tenantID",
					"companyID":               "$.companyID",
					"branchID":                "$.branchID",
					"assets":                  "$.steps.1.result.active_assets",
					"monthlyDepreciations":    "$.steps.2.result.monthly_depreciation_entries",
					"depreciationDate":        "$.input.depreciation_date",
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
				CompensationSteps: []int32{202},
			},
			// Step 4: Calculate Accumulated Depreciation Impact
			{
				StepNumber:    4,
				ServiceName:   "depreciation",
				HandlerMethod: "CalculateAccumulatedDepreciationImpact",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"assets":               "$.steps.1.result.active_assets",
					"cappedDepreciations":  "$.steps.3.result.capped_depreciation_entries",
					"depreciationDate":     "$.input.depreciation_date",
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
				CompensationSteps: []int32{203},
			},
			// Step 5: Create Depreciation Journal Entry - Depreciation Expense DR, Accumulated Depr CR
			{
				StepNumber:    5,
				ServiceName:   "depreciation",
				HandlerMethod: "CreateDepreciationJournalEntry",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"assets":               "$.steps.1.result.active_assets",
					"accumulatedDeprImpact": "$.steps.4.result.accumulated_depreciation_impact",
					"depreciationDate":     "$.input.depreciation_date",
					"periodID":             "$.input.period_id",
					"journalDescription":   "$.input.journal_description",
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
				CompensationSteps: []int32{204},
			},
			// Step 6: Post to General Ledger - CRITICAL
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostDepreciationToGL",
				InputMapping: map[string]string{
					"tenantID":                     "$.tenantID",
					"companyID":                    "$.companyID",
					"branchID":                     "$.branchID",
					"depreciationJournalEntries":   "$.steps.5.result.journal_entries",
					"depreciationDate":             "$.input.depreciation_date",
					"periodID":                     "$.input.period_id",
					"depreciationExpenseAccount":   "$.input.depreciation_expense_account",
					"accumulatedDepreciationAccount": "$.input.accumulated_depreciation_account",
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
				CompensationSteps: []int32{205},
			},
			// Step 7: Update Asset NBV - Cost - Accumulated Depreciation - CRITICAL
			{
				StepNumber:    7,
				ServiceName:   "asset",
				HandlerMethod: "UpdateAssetNBV",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"assets":               "$.steps.1.result.active_assets",
					"accumulatedDeprImpact": "$.steps.4.result.accumulated_depreciation_impact",
					"depreciationDate":     "$.input.depreciation_date",
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
				CompensationSteps: []int32{206},
			},
			// Step 8: Archive Depreciation Entry
			{
				StepNumber:    8,
				ServiceName:   "depreciation",
				HandlerMethod: "ArchiveDepreciationEntry",
				InputMapping: map[string]string{
					"tenantID":                     "$.tenantID",
					"companyID":                    "$.companyID",
					"branchID":                     "$.branchID",
					"depreciationJournalEntries":   "$.steps.5.result.journal_entries",
					"glPostingID":                  "$.steps.6.result.posting_id",
					"depreciationDate":             "$.input.depreciation_date",
					"periodID":                     "$.input.period_id",
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
				CompensationSteps: []int32{207},
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *AssetDepreciationSaga) SagaType() string {
	return "SAGA-A02"
}

// GetStepDefinitions returns all step definitions for this saga
func (s *AssetDepreciationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *AssetDepreciationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: depreciation_date, period_id
func (s *AssetDepreciationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	// Extract the nested 'input' object
	innerInput, ok := inputMap["input"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing 'input' field in saga input")
	}

	// Validate depreciation_date
	if innerInput["depreciation_date"] == nil {
		return errors.New("missing required field: depreciation_date")
	}
	depreciationDate, ok := innerInput["depreciation_date"].(string)
	if !ok || depreciationDate == "" {
		return errors.New("depreciation_date must be a non-empty string (YYYY-MM-DD format)")
	}

	// Validate period_id
	if innerInput["period_id"] == nil {
		return errors.New("missing required field: period_id")
	}
	periodID, ok := innerInput["period_id"].(string)
	if !ok || periodID == "" {
		return errors.New("period_id must be a non-empty string")
	}

	// Validate company_id (from context)
	if inputMap["companyID"] == nil {
		return errors.New("missing companyID in saga context")
	}

	// Validate optional GL account mappings if provided
	if innerInput["depreciation_expense_account"] != nil {
		depExpAccount, ok := innerInput["depreciation_expense_account"].(string)
		if !ok || depExpAccount == "" {
			return errors.New("depreciation_expense_account must be a non-empty string if provided")
		}
	}

	if innerInput["accumulated_depreciation_account"] != nil {
		accDepAccount, ok := innerInput["accumulated_depreciation_account"].(string)
		if !ok || accDepAccount == "" {
			return errors.New("accumulated_depreciation_account must be a non-empty string if provided")
		}
	}

	return nil
}
