// Package asset provides saga handlers for asset management workflows
package asset

import (
	"errors"
	"fmt"

	"p9e.in/samavaya/packages/saga"
)

// AssetDisposalSaga implements SAGA-A03: Asset Disposal (Gain/Loss Calculation)
// Business Flow: 9 steps for asset disposal with gain/loss calculation and GL posting
// Accounting Standard: IAS 16 (Derecognition of PP&E), Section 45 (Capital Gains Tax)
// Critical Steps: 2, 4, 6, 8
// Timeout: 240 seconds
// Supports: Sale, Scrap, Donation disposals
type AssetDisposalSaga struct {
	steps []*saga.StepDefinition
}

// NewAssetDisposalSaga creates a new Asset Disposal saga handler
func NewAssetDisposalSaga() saga.SagaHandler {
	return &AssetDisposalSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Asset Disposal - Sale, Scrap, or Donate
			{
				StepNumber:    1,
				ServiceName:   "asset",
				HandlerMethod: "InitiateAssetDisposal",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"assetID":        "$.input.asset_id",
					"disposalType":   "$.input.disposal_type",
					"disposalReason": "$.input.disposal_reason",
					"disposalDate":   "$.input.disposal_date",
					"approverID":     "$.input.approver_id",
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
				CompensationSteps: []int32{301},
			},
			// Step 2: Retrieve Current NBV - Cost - Accumulated Depreciation - CRITICAL
			{
				StepNumber:    2,
				ServiceName:   "fixed-assets",
				HandlerMethod: "RetrieveCurrentNBV",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"assetID":       "$.input.asset_id",
					"disposalDate":  "$.input.disposal_date",
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
			// Step 3: Determine Sale Proceeds - Cash or credit/AR
			{
				StepNumber:    3,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "DetermineSaleProceeds",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"assetID":            "$.input.asset_id",
					"disposalType":       "$.input.disposal_type",
					"salePrice":          "$.input.sale_price",
					"buyerName":          "$.input.buyer_name",
					"buyerType":          "$.input.buyer_type",
					"paymentTerms":       "$.input.payment_terms",
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
				CompensationSteps: []int32{302},
			},
			// Step 4: Calculate Gain/Loss - Proceeds - NBV - CRITICAL
			{
				StepNumber:    4,
				ServiceName:   "fixed-assets",
				HandlerMethod: "CalculateGainLoss",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"assetID":           "$.input.asset_id",
					"currentNBV":        "$.steps.2.result.current_nbv",
					"saleProceeds":      "$.steps.3.result.sale_proceeds",
					"disposalType":      "$.input.disposal_type",
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
			// Step 5: Create Disposal Journal Entry - GL posting logic varies by gain/loss
			{
				StepNumber:    5,
				ServiceName:   "depreciation",
				HandlerMethod: "CreateDisposalJournalEntry",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"assetID":               "$.input.asset_id",
					"costBasis":             "$.steps.2.result.cost_basis",
					"accumulatedDepreciation": "$.steps.2.result.accumulated_depreciation",
					"gainLoss":              "$.steps.4.result.gain_loss",
					"gainLossType":          "$.steps.4.result.gain_loss_type",
					"disposalDate":          "$.input.disposal_date",
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
				CompensationSteps: []int32{303},
			},
			// Step 6: Post GL Entries - Complex logic based on gain/loss - CRITICAL
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostAssetDisposalGL",
				InputMapping: map[string]string{
					"tenantID":                "$.tenantID",
					"companyID":               "$.companyID",
					"branchID":                "$.branchID",
					"assetID":                 "$.input.asset_id",
					"costBasis":               "$.steps.2.result.cost_basis",
					"accumulatedDepreciation": "$.steps.2.result.accumulated_depreciation",
					"saleProceeds":            "$.steps.3.result.sale_proceeds",
					"gainLoss":                "$.steps.4.result.gain_loss",
					"gainLossType":            "$.steps.4.result.gain_loss_type",
					"journalEntries":          "$.steps.5.result.journal_entries",
					"disposalDate":            "$.input.disposal_date",
					"fixedAssetAccount":       "$.input.fixed_asset_account",
					"accDepreciationAccount":  "$.input.accumulated_depreciation_account",
					"gainOnSaleAccount":       "$.input.gain_on_sale_account",
					"lossOnSaleAccount":       "$.input.loss_on_sale_account",
					"cashAccount":             "$.input.cash_account",
					"arAccount":               "$.input.accounts_receivable_account",
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
				CompensationSteps: []int32{304},
			},
			// Step 7: Update Asset Status - Mark as RETIRED
			{
				StepNumber:    7,
				ServiceName:   "asset",
				HandlerMethod: "UpdateAssetStatusToRetired",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"assetID":           "$.input.asset_id",
					"status":            "RETIRED",
					"disposalDate":      "$.input.disposal_date",
					"disposalType":      "$.input.disposal_type",
					"disposalReason":    "$.input.disposal_reason",
					"glPostingID":       "$.steps.6.result.posting_id",
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
				CompensationSteps: []int32{305},
			},
			// Step 8: Remove from Depreciation Schedule - CRITICAL
			{
				StepNumber:    8,
				ServiceName:   "depreciation",
				HandlerMethod: "RemoveFromDepreciationSchedule",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"assetID":       "$.input.asset_id",
					"disposalDate":  "$.input.disposal_date",
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
				CompensationSteps: []int32{306},
			},
			// Step 9: Archive Disposal Record
			{
				StepNumber:    9,
				ServiceName:   "asset",
				HandlerMethod: "ArchiveDisposalRecord",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"assetID":               "$.input.asset_id",
					"disposalDate":          "$.input.disposal_date",
					"glPostingID":           "$.steps.6.result.posting_id",
					"disposalType":          "$.input.disposal_type",
					"gainLoss":              "$.steps.4.result.gain_loss",
					"saleProceeds":          "$.steps.3.result.sale_proceeds",
					"disposalRefNumber":     "$.input.disposal_ref_number",
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
				CompensationSteps: []int32{307},
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *AssetDisposalSaga) SagaType() string {
	return "SAGA-A03"
}

// GetStepDefinitions returns all step definitions for this saga
func (s *AssetDisposalSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *AssetDisposalSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: asset_id, disposal_type, disposal_date
// Optional: sale_price (required for SALE type)
func (s *AssetDisposalSaga) ValidateInput(input interface{}) error {
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

	// Validate disposal_type
	if innerInput["disposal_type"] == nil {
		return errors.New("missing required field: disposal_type")
	}
	disposalType, ok := innerInput["disposal_type"].(string)
	if !ok || disposalType == "" {
		return errors.New("disposal_type must be a non-empty string (SALE, SCRAP, DONATE)")
	}

	// Validate disposal_type values
	validTypes := map[string]bool{"SALE": true, "SCRAP": true, "DONATE": true}
	if !validTypes[disposalType] {
		return fmt.Errorf("invalid disposal_type: %s (must be SALE, SCRAP, or DONATE)", disposalType)
	}

	// Validate disposal_date
	if innerInput["disposal_date"] == nil {
		return errors.New("missing required field: disposal_date")
	}
	disposalDate, ok := innerInput["disposal_date"].(string)
	if !ok || disposalDate == "" {
		return errors.New("disposal_date must be a non-empty string (YYYY-MM-DD format)")
	}

	// For SALE type, validate sale_price
	if disposalType == "SALE" {
		if innerInput["sale_price"] == nil {
			return errors.New("sale_price is required for SALE disposal type")
		}
		salePrice, ok := innerInput["sale_price"].(float64)
		if !ok || salePrice < 0 {
			return errors.New("sale_price must be a non-negative number")
		}

		// Validate buyer information for SALE
		if innerInput["buyer_name"] == nil {
			return errors.New("buyer_name is required for SALE disposal type")
		}
		buyerName, ok := innerInput["buyer_name"].(string)
		if !ok || buyerName == "" {
			return errors.New("buyer_name must be a non-empty string")
		}
	}

	// Validate company_id (from context)
	if inputMap["companyID"] == nil {
		return errors.New("missing companyID in saga context")
	}

	return nil
}
