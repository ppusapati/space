// Package warranty provides saga handlers for warranty and service module workflows
package warranty

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// SLAManagementSaga implements SAGA-W04: SLA Management workflow
// Business Flow: DetectSLABreach → TriggerEscalation → NotifyManagement → InitiateResponse →
// CommitResolution → CommunicateWithCustomer → ExecuteResolution → CheckSLACompliance →
// ProcessRemediation → PostGLAdjustment → CompleteResolution
// Timeout: 120 seconds, Critical steps: 1,2,6,8,10
type SLAManagementSaga struct {
	steps []*saga.StepDefinition
}

// NewSLAManagementSaga creates a new SLA Management saga handler (SAGA-W04)
func NewSLAManagementSaga() saga.SagaHandler {
	return &SLAManagementSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Detect SLA Breach
			{
				StepNumber:    1,
				ServiceName:   "sla",
				HandlerMethod: "DetectSLABreach",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"slaID":         "$.input.sla_id",
					"detectionTime": "$.input.detection_time",
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
			// Step 2: Trigger Escalation
			{
				StepNumber:    2,
				ServiceName:   "escalation",
				HandlerMethod: "TriggerEscalation",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"slaID":         "$.input.sla_id",
					"breachDetails": "$.steps.1.result.breach_details",
					"escalationLevel": "$.input.escalation_level",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				CompensationSteps: []int32{102},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Notify Management
			{
				StepNumber:    3,
				ServiceName:   "notification",
				HandlerMethod: "NotifyManagement",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"escalationTrigger": "$.steps.2.result.escalation_trigger",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{103},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Initiate Response
			{
				StepNumber:    4,
				ServiceName:   "service-agreement",
				HandlerMethod: "InitiateResponse",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"slaID":         "$.input.sla_id",
					"breachDetails": "$.steps.1.result.breach_details",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Commit Resolution
			{
				StepNumber:    5,
				ServiceName:   "approval",
				HandlerMethod: "CommitResolution",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"responseInitiation": "$.steps.4.result.response_initiation",
					"resolutionDeadline": "$.input.resolution_deadline",
				},
				TimeoutSeconds: 45,
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
			// Step 6: Communicate With Customer
			{
				StepNumber:    6,
				ServiceName:   "notification",
				HandlerMethod: "CommunicateWithCustomer",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"customerID":    "$.input.customer_id",
					"resolutionCommitment": "$.steps.5.result.resolution_commitment",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Execute Resolution
			{
				StepNumber:    7,
				ServiceName:   "service-delivery",
				HandlerMethod: "ExecuteResolution",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"resolutionCommitment": "$.steps.5.result.resolution_commitment",
					"customerCommunication": "$.steps.6.result.communication_record",
				},
				TimeoutSeconds: 90,
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
			// Step 8: Check SLA Compliance
			{
				StepNumber:    8,
				ServiceName:   "sla",
				HandlerMethod: "CheckSLACompliance",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"slaID":         "$.input.sla_id",
					"resolutionExecution": "$.steps.7.result.resolution_execution",
				},
				TimeoutSeconds: 30,
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
			// Step 9: Process Remediation
			{
				StepNumber:    9,
				ServiceName:   "approval",
				HandlerMethod: "ProcessRemediation",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"complianceCheck": "$.steps.8.result.compliance_check",
					"remediationPolicy": "$.input.remediation_policy",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Post GL Adjustment
			{
				StepNumber:    10,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostSLAAdjustment",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"remediationApproval": "$.steps.9.result.remediation_approval",
					"adjustmentAmount": "$.input.adjustment_amount",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{110},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 11: Complete Resolution
			{
				StepNumber:    11,
				ServiceName:   "sla",
				HandlerMethod: "CompleteResolution",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"slaID":         "$.input.sla_id",
					"glAdjustment":  "$.steps.10.result.journal_entries",
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

			// Step 102: CancelEscalation (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "escalation",
				HandlerMethod: "CancelEscalation",
				InputMapping: map[string]string{
					"slaID": "$.input.sla_id",
					"escalationTrigger": "$.steps.2.result.escalation_trigger",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: CancelManagementNotification (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "notification",
				HandlerMethod: "CancelManagementNotification",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"notificationID": "$.steps.3.result.notification_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 104: CancelResponseInitiation (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "service-agreement",
				HandlerMethod: "CancelResponseInitiation",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"responseInitiation": "$.steps.4.result.response_initiation",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: WithdrawResolutionCommitment (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "approval",
				HandlerMethod: "WithdrawResolutionCommitment",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"resolutionCommitment": "$.steps.5.result.resolution_commitment",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: CancelCustomerCommunication (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "notification",
				HandlerMethod: "CancelCustomerCommunication",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"communicationRecord": "$.steps.6.result.communication_record",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: ReverseResolutionExecution (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "service-delivery",
				HandlerMethod: "ReverseResolutionExecution",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"resolutionExecution": "$.steps.7.result.resolution_execution",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 108: RevertComplianceCheck (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "sla",
				HandlerMethod: "RevertComplianceCheck",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"complianceCheck": "$.steps.8.result.compliance_check",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 109: CancelRemediationApproval (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "approval",
				HandlerMethod: "CancelRemediationApproval",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"remediationApproval": "$.steps.9.result.remediation_approval",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 110: ReverseSLAAdjustment (compensates step 10)
			{
				StepNumber:    110,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseSLAAdjustment",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"journalEntryID": "$.steps.10.result.journal_entry_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *SLAManagementSaga) SagaType() string {
	return "SAGA-W04"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *SLAManagementSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *SLAManagementSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *SLAManagementSaga) ValidateInput(input interface{}) error {
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

	if inputMap["sla_id"] == nil {
		return errors.New("sla_id is required")
	}

	slaID, ok := inputMap["sla_id"].(string)
	if !ok || slaID == "" {
		return errors.New("sla_id must be a non-empty string")
	}

	if inputMap["detection_time"] == nil {
		return errors.New("detection_time is required")
	}

	return nil
}
