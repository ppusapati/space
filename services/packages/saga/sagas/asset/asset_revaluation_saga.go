// Package asset provides saga handlers for asset management workflows
package asset

import (
	"errors"
	"fmt"

	"p9e.in/samavaya/packages/saga"
)

// AssetRevaluationSaga implements SAGA-A04: Asset Revaluation (IAS 16 Fair Value)
// Business Flow: 7 steps for fair value revaluation and GL posting
// Accounting Standard: IAS 16 (Revaluation measurement), IndAS 16
// Critical Steps: 2, 3, 5
// Timeout: 180 seconds
// Triggers: Annual revaluation, event-based triggers (market change, appraisal update)
type AssetRevaluationSaga struct {
	steps []*saga.StepDefinition
}

// NewAssetRevaluationSaga creates a new Asset Revaluation saga handler
func NewAssetRevaluationSaga() saga.SagaHandler {
	return &AssetRevaluationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Trigger Revaluation - Annual or event-based
			{
				StepNumber:    1,
				ServiceName:   "asset",
				HandlerMethod: "TriggerRevaluation",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"assetID":           "$.input.asset_id",
					"triggerType":       "$.input.trigger_type",
					"revaluationDate":   "$.input.revaluation_date",
					"revaluationReason": "$.input.revaluation_reason",
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
				CompensationSteps: []int32{401},
			},
			// Step 2: Determine Fair Value - Market assessment or appraisal - CRITICAL
			{
				StepNumber:    2,
				ServiceName:   "fixed-assets",
				HandlerMethod: "DetermineFairValue",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"assetID":               "$.input.asset_id",
					"revaluationDate":       "$.input.revaluation_date",
					"valuationMethod":       "$.input.valuation_method",
					"appraisalAmount":       "$.input.appraisal_amount",
					"marketComparables":     "$.input.market_comparables",
					"appraisalCertificate":  "$.input.appraisal_certificate",
					"appraiserName":         "$.input.appraiser_name",
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
			// Step 3: Calculate Revaluation Increase/Decrease - CRITICAL
			{
				StepNumber:    3,
				ServiceName:   "fixed-assets",
				HandlerMethod: "CalculateRevaluationAmount",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"assetID":          "$.input.asset_id",
					"priorBookValue":   "$.input.prior_book_value",
					"fairValue":        "$.steps.2.result.fair_value",
					"revaluationDate":  "$.input.revaluation_date",
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
			// Step 4: Determine Treatment - Revaluation Reserve (Equity) adjustment
			{
				StepNumber:    4,
				ServiceName:   "fixed-assets",
				HandlerMethod: "DetermineRevaluationTreatment",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"assetID":            "$.input.asset_id",
					"revaluationAmount":  "$.steps.3.result.revaluation_amount",
					"revaluationType":    "$.steps.3.result.revaluation_type",
					"previousRevaluations": "$.input.previous_revaluations",
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
				CompensationSteps: []int32{402},
			},
			// Step 5: Post GL Entries - Asset updated to fair value, Revaluation Reserve - CRITICAL
			{
				StepNumber:    5,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostAssetRevaluationGL",
				InputMapping: map[string]string{
					"tenantID":                  "$.tenantID",
					"companyID":                 "$.companyID",
					"branchID":                  "$.branchID",
					"assetID":                   "$.input.asset_id",
					"fairValue":                 "$.steps.2.result.fair_value",
					"revaluationAmount":         "$.steps.3.result.revaluation_amount",
					"revaluationType":           "$.steps.3.result.revaluation_type",
					"revaluationTreatment":      "$.steps.4.result.treatment",
					"revaluationDate":           "$.input.revaluation_date",
					"fixedAssetAccount":         "$.input.fixed_asset_account",
					"accumulatedDepreciationAccount": "$.input.accumulated_depreciation_account",
					"revaluationReserveAccount": "$.input.revaluation_reserve_account",
					"gainOnRevaluationAccount":  "$.input.gain_on_revaluation_account",
					"lossOnRevaluationAccount":  "$.input.loss_on_revaluation_account",
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
				CompensationSteps: []int32{403},
			},
			// Step 6: Reset Accumulated Depreciation - Per IAS 16 revaluation option
			{
				StepNumber:    6,
				ServiceName:   "depreciation",
				HandlerMethod: "ResetAccumulatedDepreciation",
				InputMapping: map[string]string{
					"tenantID":               "$.tenantID",
					"companyID":              "$.companyID",
					"branchID":               "$.branchID",
					"assetID":                "$.input.asset_id",
					"fairValue":              "$.steps.2.result.fair_value",
					"accumulatedDepreciation": "$.input.accumulated_depreciation",
					"revaluationDate":        "$.input.revaluation_date",
					"depreciationResetOption": "$.input.depreciation_reset_option",
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
				CompensationSteps: []int32{404},
			},
			// Step 7: Archive Revaluation Record
			{
				StepNumber:    7,
				ServiceName:   "asset",
				HandlerMethod: "ArchiveRevaluationRecord",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"assetID":              "$.input.asset_id",
					"fairValue":            "$.steps.2.result.fair_value",
					"revaluationAmount":    "$.steps.3.result.revaluation_amount",
					"revaluationType":      "$.steps.3.result.revaluation_type",
					"revaluationDate":      "$.input.revaluation_date",
					"glPostingID":          "$.steps.5.result.posting_id",
					"appraisalCertificate": "$.input.appraisal_certificate",
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
				CompensationSteps: []int32{405},
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *AssetRevaluationSaga) SagaType() string {
	return "SAGA-A04"
}

// GetStepDefinitions returns all step definitions for this saga
func (s *AssetRevaluationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *AssetRevaluationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: asset_id, revaluation_date, valuation_method, appraisal_amount
func (s *AssetRevaluationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	// Extract the nested 'input' object
	innerInput, ok := inputMap["input"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing 'input' field in saga input")
	}

	// Validate asset_id
	if innerInput["asset_id"] == nil {
		return errors.New("missing required field: asset_id")
	}
	assetID, ok := innerInput["asset_id"].(string)
	if !ok || assetID == "" {
		return errors.New("asset_id must be a non-empty string")
	}

	// Validate revaluation_date
	if innerInput["revaluation_date"] == nil {
		return errors.New("missing required field: revaluation_date")
	}
	revaluationDate, ok := innerInput["revaluation_date"].(string)
	if !ok || revaluationDate == "" {
		return errors.New("revaluation_date must be a non-empty string (YYYY-MM-DD format)")
	}

	// Validate valuation_method
	if innerInput["valuation_method"] == nil {
		return errors.New("missing required field: valuation_method")
	}
	valuationMethod, ok := innerInput["valuation_method"].(string)
	if !ok || valuationMethod == "" {
		return errors.New("valuation_method must be a non-empty string")
	}

	// Validate valuation method values
	validMethods := map[string]bool{"MARKET": true, "APPRAISAL": true, "INCOME": true, "COST": true}
	if !validMethods[valuationMethod] {
		return fmt.Errorf("invalid valuation_method: %s (must be MARKET, APPRAISAL, INCOME, or COST)", valuationMethod)
	}

	// Validate appraisal_amount
	if innerInput["appraisal_amount"] == nil {
		return errors.New("missing required field: appraisal_amount")
	}
	appraisalAmount, ok := innerInput["appraisal_amount"].(float64)
	if !ok || appraisalAmount <= 0 {
		return errors.New("appraisal_amount must be a positive number")
	}

	// Validate trigger_type if provided
	if innerInput["trigger_type"] != nil {
		triggerType, ok := innerInput["trigger_type"].(string)
		if !ok || triggerType == "" {
			return errors.New("trigger_type must be a non-empty string if provided")
		}
		// Validate trigger type values
		validTriggers := map[string]bool{"ANNUAL": true, "EVENT_BASED": true, "MARKET_CHANGE": true}
		if !validTriggers[triggerType] {
			return fmt.Errorf("invalid trigger_type: %s (must be ANNUAL, EVENT_BASED, or MARKET_CHANGE)", triggerType)
		}
	}

	// Validate prior_book_value
	if innerInput["prior_book_value"] == nil {
		return errors.New("missing required field: prior_book_value")
	}
	priorBookValue, ok := innerInput["prior_book_value"].(float64)
	if !ok || priorBookValue < 0 {
		return errors.New("prior_book_value must be a non-negative number")
	}

	// Validate company_id (from context)
	if inputMap["companyID"] == nil {
		return errors.New("missing companyID in saga context")
	}

	return nil
}
