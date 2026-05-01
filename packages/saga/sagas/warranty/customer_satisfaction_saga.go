// Package warranty provides saga handlers for warranty and service module workflows
package warranty

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// CustomerSatisfactionSaga implements SAGA-W05: Customer Satisfaction workflow
// Business Flow: CompleteService → RequestFeedback → CollectResponse → AnalyzeFeedback →
// ScoreSatisfaction → AssessRemediation → ExecuteRework → CollectFollowUpFeedback →
// VerifySatisfaction → PostGLAdjustmentForRework → ProcessIncentive → CompleteProcess
// Timeout: 240 seconds, Critical steps: 1,3,4,8,9,11
type CustomerSatisfactionSaga struct {
	steps []*saga.StepDefinition
}

// NewCustomerSatisfactionSaga creates a new Customer Satisfaction saga handler (SAGA-W05)
func NewCustomerSatisfactionSaga() saga.SagaHandler {
	return &CustomerSatisfactionSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Complete Service
			{
				StepNumber:    1,
				ServiceName:   "service-delivery",
				HandlerMethod: "CompleteService",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"customerID":    "$.input.customer_id",
					"completionDate": "$.input.completion_date",
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
			// Step 2: Request Feedback
			{
				StepNumber:    2,
				ServiceName:   "feedback",
				HandlerMethod: "RequestFeedback",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"customerID":    "$.input.customer_id",
					"serviceCompletion": "$.steps.1.result.service_completion",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{102},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Collect Response
			{
				StepNumber:    3,
				ServiceName:   "feedback",
				HandlerMethod: "CollectResponse",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"feedbackRequest": "$.steps.2.result.feedback_request",
					"responseDeadline": "$.input.response_deadline",
				},
				TimeoutSeconds: 60,
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
			// Step 4: Analyze Feedback
			{
				StepNumber:    4,
				ServiceName:   "quality",
				HandlerMethod: "AnalyzeFeedback",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"feedbackResponse": "$.steps.3.result.feedback_response",
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
			// Step 5: Score Satisfaction
			{
				StepNumber:    5,
				ServiceName:   "satisfaction",
				HandlerMethod: "ScoreSatisfaction",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"feedbackAnalysis": "$.steps.4.result.feedback_analysis",
				},
				TimeoutSeconds: 30,
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
			// Step 6: Assess Remediation
			{
				StepNumber:    6,
				ServiceName:   "satisfaction",
				HandlerMethod: "AssessRemediation",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"satisfactionScore": "$.steps.5.result.satisfaction_score",
					"remediationThreshold": "$.input.remediation_threshold",
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
			// Step 7: Execute Rework
			{
				StepNumber:    7,
				ServiceName:   "service-delivery",
				HandlerMethod: "ExecuteRework",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"remediationAssessment": "$.steps.6.result.remediation_assessment",
					"reworkDate":    "$.input.rework_date",
				},
				TimeoutSeconds: 120,
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
			// Step 8: Collect Follow-Up Feedback
			{
				StepNumber:    8,
				ServiceName:   "feedback",
				HandlerMethod: "CollectFollowUpFeedback",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"customerID":    "$.input.customer_id",
					"reworkExecution": "$.steps.7.result.rework_execution",
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
			// Step 9: Verify Satisfaction
			{
				StepNumber:    9,
				ServiceName:   "satisfaction",
				HandlerMethod: "VerifySatisfaction",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"followUpFeedback": "$.steps.8.result.follow_up_feedback",
					"satisfactionThreshold": "$.input.satisfaction_threshold",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Post GL Adjustment For Rework
			{
				StepNumber:    10,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostReworkAdjustment",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"reworkExecution": "$.steps.7.result.rework_execution",
					"adjustmentAmount": "$.input.rework_adjustment_amount",
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
			// Step 11: Process Incentive
			{
				StepNumber:    11,
				ServiceName:   "approval",
				HandlerMethod: "ProcessIncentive",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"satisfactionVerification": "$.steps.9.result.satisfaction_verification",
					"incentiveType": "$.input.incentive_type",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{111},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 12: Complete Process
			{
				StepNumber:    12,
				ServiceName:   "satisfaction",
				HandlerMethod: "CompleteProcess",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"incentiveProcessing": "$.steps.11.result.incentive_processing",
					"completionDate": "$.input.completion_date",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
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

			// Step 102: CancelFeedbackRequest (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "feedback",
				HandlerMethod: "CancelFeedbackRequest",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"feedbackRequest": "$.steps.2.result.feedback_request",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 103: CancelResponseCollection (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "feedback",
				HandlerMethod: "CancelResponseCollection",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"feedbackResponse": "$.steps.3.result.feedback_response",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: ClearAnalysis (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "quality",
				HandlerMethod: "ClearAnalysis",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"feedbackAnalysis": "$.steps.4.result.feedback_analysis",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 105: ClearSatisfactionScore (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "satisfaction",
				HandlerMethod: "ClearSatisfactionScore",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"satisfactionScore": "$.steps.5.result.satisfaction_score",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: CancelRemediationAssessment (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "satisfaction",
				HandlerMethod: "CancelRemediationAssessment",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"remediationAssessment": "$.steps.6.result.remediation_assessment",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: ReverseReworkExecution (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "service-delivery",
				HandlerMethod: "ReverseReworkExecution",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"reworkExecution": "$.steps.7.result.rework_execution",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 108: CancelFollowUpFeedback (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "feedback",
				HandlerMethod: "CancelFollowUpFeedback",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"followUpFeedback": "$.steps.8.result.follow_up_feedback",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 109: RevertSatisfactionVerification (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "satisfaction",
				HandlerMethod: "RevertSatisfactionVerification",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"satisfactionVerification": "$.steps.9.result.satisfaction_verification",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 110: ReverseReworkAdjustment (compensates step 10)
			{
				StepNumber:    110,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseReworkAdjustment",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"journalEntryID": "$.steps.10.result.journal_entry_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 111: ReverseIncentiveProcessing (compensates step 11)
			{
				StepNumber:    111,
				ServiceName:   "approval",
				HandlerMethod: "ReverseIncentiveProcessing",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"incentiveProcessing": "$.steps.11.result.incentive_processing",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *CustomerSatisfactionSaga) SagaType() string {
	return "SAGA-W05"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *CustomerSatisfactionSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *CustomerSatisfactionSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *CustomerSatisfactionSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["service_request_id"] == nil {
		return errors.New("service_request_id is required")
	}

	serviceRequestID, ok := inputMap["service_request_id"].(string)
	if !ok || serviceRequestID == "" {
		return errors.New("service_request_id must be a non-empty string")
	}

	if inputMap["customer_id"] == nil {
		return errors.New("customer_id is required")
	}

	customerID, ok := inputMap["customer_id"].(string)
	if !ok || customerID == "" {
		return errors.New("customer_id must be a non-empty string")
	}

	if inputMap["completion_date"] == nil {
		return errors.New("completion_date is required")
	}

	return nil
}
