// Package retail provides saga handlers for retail workflows
package retail

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// PromotionSaga implements SAGA-R05: Promotion & Discount Execution workflow
// Business Flow: InitiatePromotion → ValidatePromotionRules → LoadPromotionConfig → ApplyPromotionToPricing → UpdatePromotionInventory → PublishPromotionChannel → GeneratePromotionJournal → UpdateRevenueAccounts → RecordPromotionExecution
// Steps: 9 forward + 10 compensation = 19 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,7,10
type PromotionSaga struct {
	steps []*saga.StepDefinition
}

// NewPromotionSaga creates a new Promotion saga handler
func NewPromotionSaga() saga.SagaHandler {
	return &PromotionSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Promotion
			{
				StepNumber:    1,
				ServiceName:   "promotion",
				HandlerMethod: "InitiatePromotion",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"promotionID":    "$.input.promotion_id",
					"promotionCode":  "$.input.promotion_code",
					"promotionType":  "$.input.promotion_type",
					"startDate":      "$.input.start_date",
					"endDate":        "$.input.end_date",
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
			// Step 2: Validate Promotion Rules
			{
				StepNumber:    2,
				ServiceName:   "promotion",
				HandlerMethod: "ValidatePromotionRules",
				InputMapping: map[string]string{
					"promotionID":   "$.steps.1.result.promotion_id",
					"promotionType": "$.input.promotion_type",
					"discountPercent": "$.input.discount_percent",
					"minAmount":     "$.input.min_purchase_amount",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Load Promotion Config
			{
				StepNumber:    3,
				ServiceName:   "promotion",
				HandlerMethod: "LoadPromotionConfig",
				InputMapping: map[string]string{
					"promotionID":   "$.steps.1.result.promotion_id",
					"promotionCode": "$.input.promotion_code",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Apply Promotion to Pricing
			{
				StepNumber:    4,
				ServiceName:   "pricing",
				HandlerMethod: "ApplyPromotionToPricing",
				InputMapping: map[string]string{
					"promotionID":     "$.steps.1.result.promotion_id",
					"promotionCode":   "$.input.promotion_code",
					"discountPercent": "$.input.discount_percent",
					"applicableItems": "$.input.applicable_items",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{101},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Update Promotion Inventory
			{
				StepNumber:    5,
				ServiceName:   "inventory",
				HandlerMethod: "UpdatePromotionInventory",
				InputMapping: map[string]string{
					"promotionID":     "$.steps.1.result.promotion_id",
					"applicableItems": "$.input.applicable_items",
					"availableQuantity": "$.input.available_quantity",
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
			// Step 6: Publish Promotion to Channel
			{
				StepNumber:    6,
				ServiceName:   "pos",
				HandlerMethod: "PublishPromotionChannel",
				InputMapping: map[string]string{
					"promotionID":   "$.steps.1.result.promotion_id",
					"promotionCode": "$.input.promotion_code",
					"discountInfo":  "$.steps.4.result.discount_info",
				},
				TimeoutSeconds:    20,
				IsCritical:        false,
				CompensationSteps: []int32{103},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Generate Promotion Journal
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "GeneratePromotionJournal",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"promotionID":     "$.steps.1.result.promotion_id",
					"promotionType":   "$.input.promotion_type",
					"discountPercent": "$.input.discount_percent",
					"journalDate":     "$.input.start_date",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Update Revenue Accounts
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "UpdateRevenueAccounts",
				InputMapping: map[string]string{
					"promotionID":      "$.steps.1.result.promotion_id",
					"journalEntries":   "$.steps.7.result.journal_entries",
					"discountAmount":   "$.input.expected_discount_amount",
				},
				TimeoutSeconds:    20,
				IsCritical:        false,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Record Promotion Execution
			{
				StepNumber:    9,
				ServiceName:   "promotion",
				HandlerMethod: "RecordPromotionExecution",
				InputMapping: map[string]string{
					"promotionID":     "$.steps.1.result.promotion_id",
					"journalEntries":  "$.steps.7.result.journal_entries",
					"executionStatus": "Active",
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

			// Step 101: Revert Promotion Pricing Application (compensates step 4)
			{
				StepNumber:    101,
				ServiceName:   "pricing",
				HandlerMethod: "RevertPromotionPricingApplication",
				InputMapping: map[string]string{
					"promotionID": "$.steps.1.result.promotion_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 102: Revert Promotion Inventory Update (compensates step 5)
			{
				StepNumber:    102,
				ServiceName:   "inventory",
				HandlerMethod: "RevertPromotionInventoryUpdate",
				InputMapping: map[string]string{
					"promotionID": "$.steps.1.result.promotion_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: Revert Promotion Channel Publication (compensates step 6)
			{
				StepNumber:    103,
				ServiceName:   "pos",
				HandlerMethod: "RevertPromotionChannelPublication",
				InputMapping: map[string]string{
					"promotionID": "$.steps.1.result.promotion_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 104: Reverse Promotion Journal (compensates step 7)
			{
				StepNumber:    104,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReversePromotionJournal",
				InputMapping: map[string]string{
					"promotionID": "$.steps.1.result.promotion_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Revert Revenue Account Update (compensates step 8)
			{
				StepNumber:    105,
				ServiceName:   "general-ledger",
				HandlerMethod: "RevertRevenueAccountUpdate",
				InputMapping: map[string]string{
					"promotionID": "$.steps.1.result.promotion_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: Revert Initiate Promotion (compensates step 1)
			{
				StepNumber:    106,
				ServiceName:   "promotion",
				HandlerMethod: "RevertInitiatePromotion",
				InputMapping: map[string]string{
					"promotionID": "$.steps.1.result.promotion_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: Revert Validate Promotion Rules (compensates step 2)
			{
				StepNumber:    107,
				ServiceName:   "promotion",
				HandlerMethod: "RevertValidatePromotionRules",
				InputMapping: map[string]string{
					"promotionID": "$.steps.1.result.promotion_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 108: Revert Load Promotion Config (compensates step 3)
			{
				StepNumber:    108,
				ServiceName:   "promotion",
				HandlerMethod: "RevertLoadPromotionConfig",
				InputMapping: map[string]string{
					"promotionID": "$.steps.1.result.promotion_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 109: Revert Record Promotion Execution (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "promotion",
				HandlerMethod: "RevertRecordPromotionExecution",
				InputMapping: map[string]string{
					"promotionID": "$.steps.1.result.promotion_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 110: Revert Update Inventory Metrics (additional compensation)
			{
				StepNumber:    110,
				ServiceName:   "inventory",
				HandlerMethod: "RevertPromotionInventoryMetrics",
				InputMapping: map[string]string{
					"promotionID": "$.steps.1.result.promotion_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *PromotionSaga) SagaType() string {
	return "SAGA-R05"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *PromotionSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *PromotionSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *PromotionSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["promotion_id"] == nil {
		return errors.New("promotion_id is required")
	}

	if inputMap["promotion_code"] == nil {
		return errors.New("promotion_code is required")
	}

	if inputMap["start_date"] == nil {
		return errors.New("start_date is required")
	}

	return nil
}
