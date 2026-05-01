// Package healthcare provides saga handlers for healthcare workflows
package healthcare

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// QualityAssuranceSaga implements SAGA-HC07: Healthcare Quality & Patient Safety workflow
// Business Flow: InitiateQualityReview → DefineReviewScope → GatherQualityData → AnalyzeQualityMetrics → IdentifyImprovementAreas → PrepareQualityReport → ImplementCorrectiveActions → TrackQualityImprovement → CompleteQualityReview
// Steps: 9 forward + 8 compensation = 17 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,6,9
type QualityAssuranceSaga struct {
	steps []*saga.StepDefinition
}

// NewQualityAssuranceSaga creates a new Healthcare Quality & Patient Safety saga handler
func NewQualityAssuranceSaga() saga.SagaHandler {
	return &QualityAssuranceSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Quality Review
			{
				StepNumber:    1,
				ServiceName:   "quality-assurance",
				HandlerMethod: "InitiateQualityReview",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"qualityReviewID":  "$.input.quality_review_id",
					"reviewPeriod":     "$.input.review_period",
					"reviewType":       "$.input.review_type",
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
			// Step 2: Define Review Scope
			{
				StepNumber:    2,
				ServiceName:   "quality-assurance",
				HandlerMethod: "DefineReviewScope",
				InputMapping: map[string]string{
					"qualityReviewID": "$.steps.1.result.quality_review_id",
					"reviewPeriod":    "$.input.review_period",
					"reviewType":      "$.input.review_type",
					"validateRules":   "true",
				},
				TimeoutSeconds:    30,
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
			// Step 3: Gather Quality Data
			{
				StepNumber:    3,
				ServiceName:   "patient-management",
				HandlerMethod: "GatherQualityData",
				InputMapping: map[string]string{
					"qualityReviewID": "$.steps.1.result.quality_review_id",
					"reviewPeriod":    "$.input.review_period",
					"reviewType":      "$.input.review_type",
					"scopeData":       "$.steps.2.result.scope_data",
				},
				TimeoutSeconds:    45,
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
			// Step 4: Analyze Quality Metrics
			{
				StepNumber:    4,
				ServiceName:   "quality-assurance",
				HandlerMethod: "AnalyzeQualityMetrics",
				InputMapping: map[string]string{
					"qualityReviewID": "$.steps.1.result.quality_review_id",
					"reviewType":      "$.input.review_type",
					"qualityData":     "$.steps.3.result.quality_data",
					"scopeData":       "$.steps.2.result.scope_data",
				},
				TimeoutSeconds:    40,
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
			// Step 5: Identify Improvement Areas
			{
				StepNumber:    5,
				ServiceName:   "quality-assurance",
				HandlerMethod: "IdentifyImprovementAreas",
				InputMapping: map[string]string{
					"qualityReviewID":   "$.steps.1.result.quality_review_id",
					"qualityAnalysis":   "$.steps.4.result.quality_analysis",
					"qualityData":       "$.steps.3.result.quality_data",
				},
				TimeoutSeconds:    30,
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
			// Step 6: Prepare Quality Report
			{
				StepNumber:    6,
				ServiceName:   "compliance",
				HandlerMethod: "PrepareQualityReport",
				InputMapping: map[string]string{
					"qualityReviewID":  "$.steps.1.result.quality_review_id",
					"reviewType":       "$.input.review_type",
					"qualityAnalysis":  "$.steps.4.result.quality_analysis",
					"improvementAreas": "$.steps.5.result.improvement_areas",
				},
				TimeoutSeconds:    35,
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
			// Step 7: Implement Corrective Actions
			{
				StepNumber:    7,
				ServiceName:   "quality-assurance",
				HandlerMethod: "ImplementCorrectiveActions",
				InputMapping: map[string]string{
					"qualityReviewID":  "$.steps.1.result.quality_review_id",
					"improvementAreas": "$.steps.5.result.improvement_areas",
					"qualityReport":    "$.steps.6.result.quality_report",
				},
				TimeoutSeconds:    40,
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
			// Step 8: Track Quality Improvement
			{
				StepNumber:    8,
				ServiceName:   "audit",
				HandlerMethod: "TrackQualityImprovement",
				InputMapping: map[string]string{
					"qualityReviewID":     "$.steps.1.result.quality_review_id",
					"reviewType":          "$.input.review_type",
					"qualityReport":       "$.steps.6.result.quality_report",
					"correctiveActions":   "$.steps.7.result.corrective_actions",
				},
				TimeoutSeconds:    25,
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
			// Step 9: Complete Quality Review
			{
				StepNumber:    9,
				ServiceName:   "quality-assurance",
				HandlerMethod: "CompleteQualityReview",
				InputMapping: map[string]string{
					"qualityReviewID":   "$.steps.1.result.quality_review_id",
					"reviewType":        "$.input.review_type",
					"qualityReport":     "$.steps.6.result.quality_report",
					"improvementAreas":  "$.steps.5.result.improvement_areas",
					"completionStatus":  "Completed",
				},
				TimeoutSeconds:    15,
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

			// Step 101: Revert Quality Data Collection (compensates step 3)
			{
				StepNumber:    101,
				ServiceName:   "patient-management",
				HandlerMethod: "RevertQualityDataCollection",
				InputMapping: map[string]string{
					"qualityReviewID": "$.steps.1.result.quality_review_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 102: Revert Quality Metrics Analysis (compensates step 4)
			{
				StepNumber:    102,
				ServiceName:   "quality-assurance",
				HandlerMethod: "RevertQualityMetricsAnalysis",
				InputMapping: map[string]string{
					"qualityReviewID": "$.steps.1.result.quality_review_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
			// Step 103: Clear Improvement Areas Identification (compensates step 5)
			{
				StepNumber:    103,
				ServiceName:   "quality-assurance",
				HandlerMethod: "ClearImprovementAreasIdentification",
				InputMapping: map[string]string{
					"qualityReviewID": "$.steps.1.result.quality_review_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Withdraw Quality Report (compensates step 6)
			{
				StepNumber:    104,
				ServiceName:   "compliance",
				HandlerMethod: "WithdrawQualityReport",
				InputMapping: map[string]string{
					"qualityReviewID": "$.steps.1.result.quality_review_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 105: Revert Corrective Actions Implementation (compensates step 7)
			{
				StepNumber:    105,
				ServiceName:   "quality-assurance",
				HandlerMethod: "RevertCorrectiveActionsImplementation",
				InputMapping: map[string]string{
					"qualityReviewID": "$.steps.1.result.quality_review_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
			// Step 106: Revert Quality Improvement Tracking (compensates step 8)
			{
				StepNumber:    106,
				ServiceName:   "audit",
				HandlerMethod: "RevertQualityImprovementTracking",
				InputMapping: map[string]string{
					"qualityReviewID": "$.steps.1.result.quality_review_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *QualityAssuranceSaga) SagaType() string {
	return "SAGA-HC07"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *QualityAssuranceSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *QualityAssuranceSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *QualityAssuranceSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["quality_review_id"] == nil {
		return errors.New("quality_review_id is required")
	}

	if inputMap["review_period"] == nil {
		return errors.New("review_period is required (format: YYYY-MM to YYYY-MM)")
	}

	if inputMap["review_type"] == nil {
		return errors.New("review_type is required")
	}

	return nil
}
