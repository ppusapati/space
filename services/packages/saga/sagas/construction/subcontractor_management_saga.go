// Package construction provides saga handlers for construction module workflows
package construction

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// SubcontractorManagementSaga implements SAGA-C04: Subcontractor Management & Payment
// Business Flow: ValidateSubcontractorQualifications → EstablishSubcontractorAgreement → AllocateSubcontractingScope → MonitorSubcontractorWork → ConductWorkInspection → PreparePaymentClaim → ApprovePaymentClaim → ProcessSubcontractorPayment → UpdateContractMetrics → FinalizeSubcontractorCycle
// Timeout: 120 seconds, Critical steps: 1,2,3,5,7,10
type SubcontractorManagementSaga struct {
	steps []*saga.StepDefinition
}

// NewSubcontractorManagementSaga creates a new Subcontractor Management & Payment saga handler
func NewSubcontractorManagementSaga() saga.SagaHandler {
	return &SubcontractorManagementSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Validate Subcontractor Qualifications
			{
				StepNumber:    1,
				ServiceName:   "subcontracting",
				HandlerMethod: "ValidateSubcontractorQualifications",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"workDescription":   "$.input.work_description",
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
			// Step 2: Establish Subcontractor Agreement
			{
				StepNumber:    2,
				ServiceName:   "subcontracting",
				HandlerMethod: "EstablishSubcontractorAgreement",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"contractAmount":    "$.input.contract_amount",
					"workDescription":   "$.input.work_description",
					"qualifications":    "$.steps.1.result.qualifications_validated",
				},
				TimeoutSeconds:    30,
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
			// Step 3: Allocate Subcontracting Scope
			{
				StepNumber:    3,
				ServiceName:   "construction-site",
				HandlerMethod: "AllocateSubcontractingScope",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"workDescription":   "$.input.work_description",
					"agreement":         "$.steps.2.result.agreement_details",
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
			// Step 4: Monitor Subcontractor Work
			{
				StepNumber:    4,
				ServiceName:   "construction-site",
				HandlerMethod: "MonitorSubcontractorWork",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"scope":             "$.steps.3.result.allocated_scope",
				},
				TimeoutSeconds:    30,
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
			// Step 5: Conduct Work Inspection
			{
				StepNumber:    5,
				ServiceName:   "quality-inspection",
				HandlerMethod: "ConductWorkInspection",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"workMonitoring":    "$.steps.4.result.work_monitoring",
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
			// Step 6: Prepare Payment Claim
			{
				StepNumber:    6,
				ServiceName:   "subcontracting",
				HandlerMethod: "PreparePaymentClaim",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"contractAmount":    "$.input.contract_amount",
					"workInspection":    "$.steps.5.result.inspection_result",
				},
				TimeoutSeconds:    30,
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
			// Step 7: Approve Payment Claim
			{
				StepNumber:    7,
				ServiceName:   "approval",
				HandlerMethod: "ApprovePaymentClaim",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"paymentClaim":      "$.steps.6.result.payment_claim",
				},
				TimeoutSeconds:    30,
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
			// Step 8: Process Subcontractor Payment
			{
				StepNumber:    8,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ProcessSubcontractorPayment",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"paymentClaim":      "$.steps.6.result.payment_claim",
					"approval":          "$.steps.7.result.approval_status",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Update Contract Metrics
			{
				StepNumber:    9,
				ServiceName:   "general-ledger",
				HandlerMethod: "UpdateContractMetrics",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"contractAmount":    "$.input.contract_amount",
					"paymentRecord":     "$.steps.8.result.payment_record",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Finalize Subcontractor Cycle
			{
				StepNumber:    10,
				ServiceName:   "subcontracting",
				HandlerMethod: "FinalizeSubcontractorCycle",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"agreement":         "$.steps.2.result.agreement_details",
					"paymentRecord":     "$.steps.8.result.payment_record",
					"metrics":           "$.steps.9.result.contract_metrics",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
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

			// Step 102: TerminateSubcontractorAgreement (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "subcontracting",
				HandlerMethod: "TerminateSubcontractorAgreement",
				InputMapping: map[string]string{
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"agreementDetails": "$.steps.2.result.agreement_details",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: DeallocateSubcontractingScope (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "construction-site",
				HandlerMethod: "DeallocateSubcontractingScope",
				InputMapping: map[string]string{
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"allocatedScope":    "$.steps.3.result.allocated_scope",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: StopWorkMonitoring (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "construction-site",
				HandlerMethod: "StopWorkMonitoring",
				InputMapping: map[string]string{
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"workMonitoring":    "$.steps.4.result.work_monitoring",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: CancelWorkInspection (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "quality-inspection",
				HandlerMethod: "CancelWorkInspection",
				InputMapping: map[string]string{
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"inspectionResult": "$.steps.5.result.inspection_result",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: RejectPaymentClaim (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "subcontracting",
				HandlerMethod: "RejectPaymentClaim",
				InputMapping: map[string]string{
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"paymentClaim":      "$.steps.6.result.payment_claim",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: RevokePaymentApproval (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "approval",
				HandlerMethod: "RevokePaymentApproval",
				InputMapping: map[string]string{
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"approvalStatus":    "$.steps.7.result.approval_status",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: ReverseSubcontractorPayment (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ReverseSubcontractorPayment",
				InputMapping: map[string]string{
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"paymentRecord":     "$.steps.8.result.payment_record",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 109: ReverseContractMetrics (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseContractMetrics",
				InputMapping: map[string]string{
					"subcontractID":     "$.input.subcontract_id",
					"subcontractorID":   "$.input.subcontractor_id",
					"contractMetrics":   "$.steps.9.result.contract_metrics",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *SubcontractorManagementSaga) SagaType() string {
	return "SAGA-C04"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *SubcontractorManagementSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *SubcontractorManagementSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *SubcontractorManagementSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["subcontract_id"] == nil {
		return errors.New("subcontract_id is required")
	}

	subcontractID, ok := inputMap["subcontract_id"].(string)
	if !ok || subcontractID == "" {
		return errors.New("subcontract_id must be a non-empty string")
	}

	if inputMap["subcontractor_id"] == nil {
		return errors.New("subcontractor_id is required")
	}

	subcontractorID, ok := inputMap["subcontractor_id"].(string)
	if !ok || subcontractorID == "" {
		return errors.New("subcontractor_id must be a non-empty string")
	}

	if inputMap["contract_amount"] == nil {
		return errors.New("contract_amount is required")
	}

	contractAmount, ok := inputMap["contract_amount"].(string)
	if !ok || contractAmount == "" {
		return errors.New("contract_amount must be a non-empty string")
	}

	if inputMap["work_description"] == nil {
		return errors.New("work_description is required")
	}

	workDescription, ok := inputMap["work_description"].(string)
	if !ok || workDescription == "" {
		return errors.New("work_description must be a non-empty string")
	}

	return nil
}
