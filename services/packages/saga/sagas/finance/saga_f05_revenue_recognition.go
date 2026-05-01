// Package finance provides saga handlers for finance module workflows
package finance

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// RevenueRecognitionSaga implements SAGA-F05: Revenue Recognition (IndAS 115) workflow
// Business Flow: CreateServiceContract → IdentifyPerformanceObligations → ScheduleRecognition → PostDeferredRevenue → RecognizeMonthlyRevenue → PostRevenueRecognition → UpdateContract → CompleteContract
// IndAS 115 Compliance: Five-step model for revenue recognition from contracts with customers
type RevenueRecognitionSaga struct {
	steps []*saga.StepDefinition
}

// NewRevenueRecognitionSaga creates a new Revenue Recognition saga handler
func NewRevenueRecognitionSaga() saga.SagaHandler {
	return &RevenueRecognitionSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Service Contract
			{
				StepNumber:    1,
				ServiceName:   "billing",
				HandlerMethod: "CreateServiceContract",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"customerID":      "$.input.customer_id",
					"contractAmount":  "$.input.contract_amount",
					"startDate":       "$.input.start_date",
					"endDate":         "$.input.end_date",
					"contractTerms":   "$.input.contract_terms",
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
			// Step 2: Identify Performance Obligations (IndAS 115 Step 2)
			{
				StepNumber:    2,
				ServiceName:   "billing",
				HandlerMethod: "IdentifyPerformanceObligations",
				InputMapping: map[string]string{
					"contractID":      "$.steps.1.result.contract_id",
					"contractAmount":  "$.input.contract_amount",
					"deliverables":    "$.input.deliverables",
					"standaloneSellingPrices": "$.input.standalone_selling_prices",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{102},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Schedule Revenue Recognition (IndAS 115 Step 5)
			{
				StepNumber:    3,
				ServiceName:   "billing",
				HandlerMethod: "ScheduleRevenueRecognition",
				InputMapping: map[string]string{
					"contractID":      "$.steps.1.result.contract_id",
					"performanceObligations": "$.steps.2.result.performance_obligations",
					"recognitionPattern": "$.input.recognition_pattern",
					"startDate":       "$.input.start_date",
					"endDate":         "$.input.end_date",
				},
				TimeoutSeconds:    25,
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
			// Step 4: Post Deferred Revenue Entry
			{
				StepNumber:    4,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostDeferredRevenueEntry",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"contractID":      "$.steps.1.result.contract_id",
					"contractAmount":  "$.input.contract_amount",
					"deferredRevenueAccount": "$.input.deferred_revenue_account",
					"journalDate":     "$.input.journal_date",
				},
				TimeoutSeconds:    30,
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
			// Step 5: Create Monthly Revenue Recognition Records
			{
				StepNumber:    5,
				ServiceName:   "billing",
				HandlerMethod: "CreateMonthlyRevenueRecords",
				InputMapping: map[string]string{
					"contractID":      "$.steps.1.result.contract_id",
					"revenueSchedule": "$.steps.3.result.revenue_schedule",
					"recognitionBasis": "$.input.recognition_basis",
				},
				TimeoutSeconds:    25,
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
			// Step 6: Post Revenue Recognition Journal
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostRevenueRecognitionJournal",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"contractID":      "$.steps.1.result.contract_id",
					"revenueSchedule": "$.steps.3.result.revenue_schedule",
					"revenueAccount":  "$.input.revenue_account",
					"deferredRevenueAccount": "$.input.deferred_revenue_account",
					"journalDate":     "$.input.journal_date",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Update Contract Status
			{
				StepNumber:    7,
				ServiceName:   "billing",
				HandlerMethod: "UpdateContractStatus",
				InputMapping: map[string]string{
					"contractID": "$.steps.1.result.contract_id",
					"newStatus":  "RECOGNITION_ACTIVE",
				},
				TimeoutSeconds:    20,
				IsCritical:        false,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Complete Revenue Recognition Setup
			{
				StepNumber:    8,
				ServiceName:   "billing",
				HandlerMethod: "CompleteRevenueRecognitionSetup",
				InputMapping: map[string]string{
					"contractID":      "$.steps.1.result.contract_id",
					"revenueSchedule": "$.steps.3.result.revenue_schedule",
					"setupDate":       "$.input.journal_date",
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
			// ===== COMPENSATION STEPS =====

			// Step 101: Cancel Contract (compensates step 1)
			{
				StepNumber:    101,
				ServiceName:   "billing",
				HandlerMethod: "CancelServiceContract",
				InputMapping: map[string]string{
					"contractID": "$.steps.1.result.contract_id",
					"reason":     "Saga compensation - revenue recognition failed",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 102: Clear Performance Obligations (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "billing",
				HandlerMethod: "ClearPerformanceObligations",
				InputMapping: map[string]string{
					"contractID": "$.steps.1.result.contract_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: Delete Revenue Schedule (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "billing",
				HandlerMethod: "DeleteRevenueSchedule",
				InputMapping: map[string]string{
					"contractID": "$.steps.1.result.contract_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 104: Reverse Deferred Revenue Entry (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseDeferredRevenueEntry",
				InputMapping: map[string]string{
					"contractID": "$.steps.1.result.contract_id",
					"journalDate": "$.input.journal_date",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: Delete Monthly Revenue Records (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "billing",
				HandlerMethod: "DeleteMonthlyRevenueRecords",
				InputMapping: map[string]string{
					"contractID": "$.steps.1.result.contract_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Reverse Revenue Recognition Journal (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseRevenueRecognitionJournal",
				InputMapping: map[string]string{
					"contractID": "$.steps.1.result.contract_id",
					"journalDate": "$.input.journal_date",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Revert Contract Status (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "billing",
				HandlerMethod: "RevertContractStatus",
				InputMapping: map[string]string{
					"contractID": "$.steps.1.result.contract_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *RevenueRecognitionSaga) SagaType() string {
	return "SAGA-F05"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *RevenueRecognitionSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *RevenueRecognitionSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *RevenueRecognitionSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["customer_id"] == nil {
		return errors.New("customer_id is required")
	}

	if inputMap["contract_amount"] == nil {
		return errors.New("contract_amount is required")
	}

	amount, ok := inputMap["contract_amount"].(float64)
	if !ok || amount <= 0 {
		return errors.New("contract_amount must be a positive number")
	}

	if inputMap["start_date"] == nil {
		return errors.New("start_date is required")
	}

	if inputMap["end_date"] == nil {
		return errors.New("end_date is required")
	}

	if inputMap["deliverables"] == nil {
		return errors.New("deliverables are required for performance obligation identification")
	}

	deliverables, ok := inputMap["deliverables"].([]interface{})
	if !ok || len(deliverables) == 0 {
		return errors.New("deliverables must be a non-empty list")
	}

	if inputMap["revenue_account"] == nil {
		return errors.New("revenue_account is required")
	}

	if inputMap["deferred_revenue_account"] == nil {
		return errors.New("deferred_revenue_account is required")
	}

	if inputMap["recognition_pattern"] == nil {
		return errors.New("recognition_pattern is required (e.g., STRAIGHT_LINE, MILESTONE_BASED)")
	}

	return nil
}
