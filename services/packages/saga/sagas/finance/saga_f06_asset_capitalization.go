// Package finance provides saga handlers for finance module workflows
package finance

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// AssetCapitalizationSaga implements SAGA-F06: Asset Capitalization workflow
// Business Flow: ReceiveAsset → CreateAssetMaster → CapitalizeAsset → CreateDepreciationSchedule → PostAssetCapitalization → StartDepreciation → CompleteCapitalization → AssignToLocation
// IndAS 16 Compliance: Property, Plant & Equipment recognition and measurement
type AssetCapitalizationSaga struct {
	steps []*saga.StepDefinition
}

// NewAssetCapitalizationSaga creates a new Asset Capitalization saga handler
func NewAssetCapitalizationSaga() saga.SagaHandler {
	return &AssetCapitalizationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Receive Asset from Purchase Order
			{
				StepNumber:    1,
				ServiceName:   "purchase-order",
				HandlerMethod: "ReceiveAssetFromPO",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"poID":          "$.input.po_id",
					"assetItems":    "$.input.asset_items",
					"receiptDate":   "$.input.receipt_date",
					"supplierID":    "$.input.supplier_id",
				},
				TimeoutSeconds: 25,
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
			// Step 2: Create Asset Master Record
			{
				StepNumber:    2,
				ServiceName:   "asset",
				HandlerMethod: "CreateAssetMaster",
				InputMapping: map[string]string{
					"poID":          "$.input.po_id",
					"grnID":         "$.steps.1.result.grn_id",
					"assetItems":    "$.input.asset_items",
					"assetCategory": "$.input.asset_category",
					"assetClass":    "$.input.asset_class",
					"assetTags":     "$.input.asset_tags",
				},
				TimeoutSeconds:    25,
				IsCritical:        false,
				CompensationSteps: []int32{102},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Capitalize Asset (IndAS 16 Recognition)
			{
				StepNumber:    3,
				ServiceName:   "fixed-assets",
				HandlerMethod: "CapitalizeAsset",
				InputMapping: map[string]string{
					"assetID":           "$.steps.2.result.asset_id",
					"assetCost":         "$.input.asset_cost",
					"capitalizationDate": "$.input.capitalization_date",
					"usefulLife":        "$.input.useful_life",
					"residualValue":     "$.input.residual_value",
					"depreciationMethod": "$.input.depreciation_method",
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
			// Step 4: Create Depreciation Schedule
			{
				StepNumber:    4,
				ServiceName:   "depreciation",
				HandlerMethod: "CreateDepreciationSchedule",
				InputMapping: map[string]string{
					"assetID":           "$.steps.2.result.asset_id",
					"assetCost":         "$.input.asset_cost",
					"usefulLife":        "$.input.useful_life",
					"residualValue":     "$.input.residual_value",
					"depreciationMethod": "$.input.depreciation_method",
					"startDate":         "$.input.capitalization_date",
				},
				TimeoutSeconds:    25,
				IsCritical:        false,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Post Asset Capitalization Journal Entry
			{
				StepNumber:    5,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostAssetCapitalizationJournal",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"assetID":         "$.steps.2.result.asset_id",
					"assetCost":       "$.input.asset_cost",
					"assetAccount":    "$.input.asset_account",
					"capitalizationDate": "$.input.capitalization_date",
					"poID":            "$.input.po_id",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Start Depreciation Process
			{
				StepNumber:    6,
				ServiceName:   "depreciation",
				HandlerMethod: "StartDepreciation",
				InputMapping: map[string]string{
					"assetID":     "$.steps.2.result.asset_id",
					"scheduleID":  "$.steps.4.result.schedule_id",
					"startDate":   "$.input.capitalization_date",
				},
				TimeoutSeconds:    20,
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
			// Step 7: Complete Asset Capitalization
			{
				StepNumber:    7,
				ServiceName:   "fixed-assets",
				HandlerMethod: "CompleteAssetCapitalization",
				InputMapping: map[string]string{
					"assetID":     "$.steps.2.result.asset_id",
					"poID":        "$.input.po_id",
					"grnID":       "$.steps.1.result.grn_id",
					"scheduleID":  "$.steps.4.result.schedule_id",
				},
				TimeoutSeconds:    20,
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
			// Step 8: Assign Asset to Location/Department
			{
				StepNumber:    8,
				ServiceName:   "asset",
				HandlerMethod: "AssignAssetToLocation",
				InputMapping: map[string]string{
					"assetID":      "$.steps.2.result.asset_id",
					"locationID":   "$.input.location_id",
					"departmentID": "$.input.department_id",
					"custodianID":  "$.input.custodian_id",
					"assignmentDate": "$.input.capitalization_date",
				},
				TimeoutSeconds:    20,
				IsCritical:        false,
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

			// Step 101: Cancel Asset Receipt (compensates step 1)
			{
				StepNumber:    101,
				ServiceName:   "purchase-order",
				HandlerMethod: "CancelAssetReceipt",
				InputMapping: map[string]string{
					"grnID": "$.steps.1.result.grn_id",
					"reason": "Saga compensation - asset capitalization failed",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 102: Delete Asset Master Record (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "asset",
				HandlerMethod: "DeleteAssetMaster",
				InputMapping: map[string]string{
					"assetID": "$.steps.2.result.asset_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: Reverse Asset Capitalization (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "fixed-assets",
				HandlerMethod: "ReverseAssetCapitalization",
				InputMapping: map[string]string{
					"assetID": "$.steps.2.result.asset_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Delete Depreciation Schedule (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "depreciation",
				HandlerMethod: "DeleteDepreciationSchedule",
				InputMapping: map[string]string{
					"assetID":    "$.steps.2.result.asset_id",
					"scheduleID": "$.steps.4.result.schedule_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Reverse Capitalization Journal (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseAssetCapitalizationJournal",
				InputMapping: map[string]string{
					"assetID": "$.steps.2.result.asset_id",
					"journalDate": "$.input.capitalization_date",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: Stop Depreciation Process (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "depreciation",
				HandlerMethod: "StopDepreciation",
				InputMapping: map[string]string{
					"assetID":    "$.steps.2.result.asset_id",
					"scheduleID": "$.steps.4.result.schedule_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: Revert Capitalization Completion (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "fixed-assets",
				HandlerMethod: "RevertCapitalizationCompletion",
				InputMapping: map[string]string{
					"assetID": "$.steps.2.result.asset_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *AssetCapitalizationSaga) SagaType() string {
	return "SAGA-F06"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *AssetCapitalizationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *AssetCapitalizationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *AssetCapitalizationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["po_id"] == nil {
		return errors.New("po_id is required")
	}

	if inputMap["asset_items"] == nil {
		return errors.New("asset_items are required")
	}

	assetItems, ok := inputMap["asset_items"].([]interface{})
	if !ok || len(assetItems) == 0 {
		return errors.New("asset_items must be a non-empty list")
	}

	if inputMap["asset_cost"] == nil {
		return errors.New("asset_cost is required")
	}

	cost, ok := inputMap["asset_cost"].(float64)
	if !ok || cost <= 0 {
		return errors.New("asset_cost must be a positive number")
	}

	if inputMap["capitalization_date"] == nil {
		return errors.New("capitalization_date is required")
	}

	if inputMap["useful_life"] == nil {
		return errors.New("useful_life is required")
	}

	usefulLife, ok := inputMap["useful_life"].(float64)
	if !ok || usefulLife <= 0 {
		return errors.New("useful_life must be a positive number (in years)")
	}

	if inputMap["depreciation_method"] == nil {
		return errors.New("depreciation_method is required (e.g., SLM, WDV)")
	}

	if inputMap["asset_account"] == nil {
		return errors.New("asset_account is required for GL posting")
	}

	if inputMap["asset_category"] == nil {
		return errors.New("asset_category is required")
	}

	return nil
}
