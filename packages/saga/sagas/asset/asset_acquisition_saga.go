// Package asset provides saga handlers for asset management workflows
package asset

import (
	"errors"
	"fmt"

	"p9e.in/samavaya/packages/saga"
)

// AssetAcquisitionSaga implements SAGA-A01: Asset Acquisition (IAS 16 Capitalization)
// Business Flow: 10 steps for complete asset acquisition from PO receipt to depreciation schedule creation
// Accounting Standard: IAS 16 (Recognition and Measurement of PP&E)
// Critical Steps: 3, 5, 7, 8
// Timeout: 300 seconds
type AssetAcquisitionSaga struct {
	steps []*saga.StepDefinition
}

// NewAssetAcquisitionSaga creates a new Asset Acquisition saga handler
func NewAssetAcquisitionSaga() saga.SagaHandler {
	return &AssetAcquisitionSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Receive Goods from PO - Validate part numbers and quantities
			{
				StepNumber:    1,
				ServiceName:   "purchase-order",
				HandlerMethod: "ReceiveGoodsFromPO",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"poID":          "$.input.po_id",
					"poLineID":      "$.input.po_line_id",
					"partNumber":    "$.input.part_number",
					"quantity":      "$.input.quantity",
					"receivedDate":  "$.input.received_date",
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
				CompensationSteps: []int32{101},
			},
			// Step 2: Validate Capitalization Criteria - IAS 16 requirements
			{
				StepNumber:    2,
				ServiceName:   "fixed-assets",
				HandlerMethod: "ValidateCapitalizationCriteria",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"poAmount":           "$.input.po_amount",
					"poID":               "$.input.po_id",
					"capitalizationThreshold": "$.input.capitalization_threshold",
					"usefulLifeYears":    "$.input.useful_life_years",
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
			// Step 3: Create Asset Master Record - CRITICAL
			{
				StepNumber:    3,
				ServiceName:   "asset",
				HandlerMethod: "CreateAssetMasterRecord",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"assetTag":         "$.input.asset_tag",
					"assetCategory":    "$.input.asset_category",
					"assetLocation":    "$.input.asset_location",
					"description":      "$.input.description",
					"serialNumber":     "$.input.serial_number",
					"manufacturerName": "$.input.manufacturer_name",
					"validationStatus": "$.steps.2.result.validation_status",
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
			// Step 4: Calculate Cost Basis - PO amount + freight + installation
			{
				StepNumber:    4,
				ServiceName:   "fixed-assets",
				HandlerMethod: "CalculateCostBasis",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"assetID":           "$.steps.3.result.asset_id",
					"poAmount":          "$.input.po_amount",
					"freightCost":       "$.input.freight_cost",
					"installationCost":  "$.input.installation_cost",
					"customsDuty":       "$.input.customs_duty",
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
				CompensationSteps: []int32{103},
			},
			// Step 5: Calculate Depreciable Basis - Cost + capitalized interest - CRITICAL
			{
				StepNumber:    5,
				ServiceName:   "depreciation",
				HandlerMethod: "CalculateDepreciableBasis",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"assetID":               "$.steps.3.result.asset_id",
					"costBasis":             "$.steps.4.result.cost_basis",
					"capitalizedInterest":   "$.input.capitalized_interest",
					"salvageValue":          "$.input.salvage_value",
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
			// Step 6: Determine Depreciation Method - SLM, WDV, or Units
			{
				StepNumber:    6,
				ServiceName:   "depreciation",
				HandlerMethod: "DetermineDepreciationMethod",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"assetID":            "$.steps.3.result.asset_id",
					"assetCategory":      "$.input.asset_category",
					"usefulLifeYears":    "$.input.useful_life_years",
					"depreciationMethod": "$.input.depreciation_method",
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
			// Step 7: Post GL Entries - Fixed Asset DR, Creditor CR - CRITICAL
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostAssetAcquisitionGL",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"assetID":              "$.steps.3.result.asset_id",
					"costBasis":            "$.steps.4.result.cost_basis",
					"creditorAccount":      "$.input.creditor_account",
					"fixedAssetAccount":    "$.input.fixed_asset_account",
					"transactionDate":      "$.input.transaction_date",
					"poID":                 "$.input.po_id",
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
				CompensationSteps: []int32{106},
			},
			// Step 8: Create Depreciation Schedule - Monthly accrual starting next month - CRITICAL
			{
				StepNumber:    8,
				ServiceName:   "depreciation",
				HandlerMethod: "CreateDepreciationSchedule",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"assetID":               "$.steps.3.result.asset_id",
					"depreciableBasis":      "$.steps.5.result.depreciable_basis",
					"usefulLifeYears":       "$.input.useful_life_years",
					"depreciationMethod":    "$.steps.6.result.depreciation_method",
					"startDate":             "$.input.start_depreciation_date",
					"acquisitionDate":       "$.input.received_date",
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
				CompensationSteps: []int32{107},
			},
			// Step 9: Update Asset Registry - Status to ACTIVE
			{
				StepNumber:    9,
				ServiceName:   "asset",
				HandlerMethod: "UpdateAssetStatus",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"companyID":  "$.companyID",
					"branchID":   "$.branchID",
					"assetID":    "$.steps.3.result.asset_id",
					"status":     "ACTIVE",
					"statusDate": "$.input.received_date",
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
				CompensationSteps: []int32{108},
			},
			// Step 10: Send Asset Creation Notification
			{
				StepNumber:    10,
				ServiceName:   "notification",
				HandlerMethod: "SendAssetCreationNotification",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"companyID":   "$.companyID",
					"branchID":    "$.branchID",
					"assetID":     "$.steps.3.result.asset_id",
					"assetTag":    "$.input.asset_tag",
					"assetName":   "$.input.description",
					"costBasis":   "$.steps.4.result.cost_basis",
					"createdBy":   "$.input.created_by",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        2,
					InitialBackoffMs:  500,
					MaxBackoffMs:      5000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *AssetAcquisitionSaga) SagaType() string {
	return "SAGA-A01"
}

// GetStepDefinitions returns all step definitions for this saga
func (s *AssetAcquisitionSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *AssetAcquisitionSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: po_id, po_line_id, asset_tag, asset_category, po_amount, useful_life_years, asset_location
func (s *AssetAcquisitionSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	// Extract the nested 'input' object
	innerInput, ok := inputMap["input"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing 'input' field in saga input")
	}

	// Validate po_id
	if innerInput["po_id"] == nil {
		return errors.New("missing required field: po_id")
	}
	poID, ok := innerInput["po_id"].(string)
	if !ok || poID == "" {
		return errors.New("po_id must be a non-empty string")
	}

	// Validate po_line_id
	if innerInput["po_line_id"] == nil {
		return errors.New("missing required field: po_line_id")
	}
	poLineID, ok := innerInput["po_line_id"].(string)
	if !ok || poLineID == "" {
		return errors.New("po_line_id must be a non-empty string")
	}

	// Validate asset_tag
	if innerInput["asset_tag"] == nil {
		return errors.New("missing required field: asset_tag")
	}
	assetTag, ok := innerInput["asset_tag"].(string)
	if !ok || assetTag == "" {
		return errors.New("asset_tag must be a non-empty string")
	}

	// Validate asset_category
	if innerInput["asset_category"] == nil {
		return errors.New("missing required field: asset_category")
	}
	assetCategory, ok := innerInput["asset_category"].(string)
	if !ok || assetCategory == "" {
		return errors.New("asset_category must be a non-empty string")
	}

	// Validate po_amount
	if innerInput["po_amount"] == nil {
		return errors.New("missing required field: po_amount")
	}
	poAmount, ok := innerInput["po_amount"].(float64)
	if !ok || poAmount <= 0 {
		return errors.New("po_amount must be a positive number")
	}

	// Validate useful_life_years
	if innerInput["useful_life_years"] == nil {
		return errors.New("missing required field: useful_life_years")
	}
	usefulLife, ok := innerInput["useful_life_years"].(float64)
	if !ok || usefulLife <= 0 {
		return errors.New("useful_life_years must be a positive number")
	}

	// Validate asset_location
	if innerInput["asset_location"] == nil {
		return errors.New("missing required field: asset_location")
	}
	assetLocation, ok := innerInput["asset_location"].(string)
	if !ok || assetLocation == "" {
		return errors.New("asset_location must be a non-empty string")
	}

	// Validate received_date
	if innerInput["received_date"] == nil {
		return errors.New("missing required field: received_date")
	}
	receivedDate, ok := innerInput["received_date"].(string)
	if !ok || receivedDate == "" {
		return errors.New("received_date must be a non-empty string")
	}

	// Validate company_id (from context, not input)
	if inputMap["companyID"] == nil {
		return errors.New("missing tenantID in saga context")
	}

	return nil
}
