// Package warranty provides saga handlers for warranty and service module workflows
package warranty

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// WarrantyClaimSaga implements SAGA-W01: Warranty Claim workflow
// Business Flow: SubmitClaim → ValidateEligibility → DetermineCoverage → ApproveClaimRequest →
// EstimateCost → ResolveDefect → SettleClaim → GenerateInvoice → PostGL → NotifyCustomer
// Timeout: 180 seconds, Critical steps: 1,2,3,4,8,9
type WarrantyClaimSaga struct {
	steps []*saga.StepDefinition
}

// NewWarrantyClaimSaga creates a new Warranty Claim saga handler (SAGA-W01)
func NewWarrantyClaimSaga() saga.SagaHandler {
	return &WarrantyClaimSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Submit Claim
			{
				StepNumber:    1,
				ServiceName:   "warranty",
				HandlerMethod: "SubmitClaim",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"claimID":      "$.input.claim_id",
					"productID":    "$.input.product_id",
					"customerID":   "$.input.customer_id",
					"issueDescription": "$.input.issue_description",
					"claimDate":    "$.input.claim_date",
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
			// Step 2: Validate Eligibility
			{
				StepNumber:    2,
				ServiceName:   "warranty",
				HandlerMethod: "ValidateEligibility",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"claimID":      "$.input.claim_id",
					"productID":    "$.input.product_id",
					"customerID":   "$.input.customer_id",
					"claimSubmission": "$.steps.1.result.claim_submission",
				},
				TimeoutSeconds: 45,
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
			// Step 3: Determine Coverage
			{
				StepNumber:    3,
				ServiceName:   "warranty",
				HandlerMethod: "DetermineCoverage",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"claimID":      "$.input.claim_id",
					"productID":    "$.input.product_id",
					"eligibilityCheck": "$.steps.2.result.eligibility_check",
					"warrantyType": "$.input.warranty_type",
				},
				TimeoutSeconds: 45,
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
			// Step 4: Approve Claim Request
			{
				StepNumber:    4,
				ServiceName:   "approval",
				HandlerMethod: "ApproveClaimRequest",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"claimID":      "$.input.claim_id",
					"coverageDetermination": "$.steps.3.result.coverage_determination",
					"approvalLevel": "$.input.approval_level",
				},
				TimeoutSeconds: 60,
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
			// Step 5: Estimate Cost
			{
				StepNumber:    5,
				ServiceName:   "warranty",
				HandlerMethod: "EstimateCost",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"claimID":      "$.input.claim_id",
					"productID":    "$.input.product_id",
					"claimApproval": "$.steps.4.result.claim_approval",
					"issueDescription": "$.input.issue_description",
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
			// Step 6: Resolve Defect
			{
				StepNumber:    6,
				ServiceName:   "asset",
				HandlerMethod: "ResolveDefect",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"claimID":      "$.input.claim_id",
					"productID":    "$.input.product_id",
					"costEstimate": "$.steps.5.result.cost_estimate",
					"defectType":   "$.input.defect_type",
				},
				TimeoutSeconds: 120,
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
			// Step 7: Settle Claim
			{
				StepNumber:    7,
				ServiceName:   "insurance",
				HandlerMethod: "SettleClaim",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"claimID":      "$.input.claim_id",
					"customerID":   "$.input.customer_id",
					"defectResolution": "$.steps.6.result.defect_resolution",
					"costEstimate": "$.steps.5.result.cost_estimate",
				},
				TimeoutSeconds: 60,
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
			// Step 8: Generate Invoice
			{
				StepNumber:    8,
				ServiceName:   "sales-invoice",
				HandlerMethod: "GenerateInvoice",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"claimID":      "$.input.claim_id",
					"customerID":   "$.input.customer_id",
					"productID":    "$.input.product_id",
					"claimSettlement": "$.steps.7.result.claim_settlement",
					"invoiceDate":  "$.input.claim_date",
				},
				TimeoutSeconds: 45,
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
			// Step 9: Post GL (General Ledger)
			{
				StepNumber:    9,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostClaimJournal",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"claimID":      "$.input.claim_id",
					"invoiceID":    "$.steps.8.result.invoice_id",
					"costEstimate": "$.steps.5.result.cost_estimate",
					"journalDate":  "$.input.claim_date",
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
			// Step 10: Notify Customer
			{
				StepNumber:    10,
				ServiceName:   "notification",
				HandlerMethod: "NotifyCustomer",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"claimID":      "$.input.claim_id",
					"customerID":   "$.input.customer_id",
					"glPosting":    "$.steps.9.result.journal_entries",
					"claimStatus":  "RESOLVED",
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

			// Step 103: RevertCoverageDetermination (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "warranty",
				HandlerMethod: "RevertCoverageDetermination",
				InputMapping: map[string]string{
					"claimID": "$.input.claim_id",
					"coverageDetermination": "$.steps.3.result.coverage_determination",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 104: RejectClaimApproval (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "approval",
				HandlerMethod: "RejectClaimApproval",
				InputMapping: map[string]string{
					"claimID": "$.input.claim_id",
					"claimApproval": "$.steps.4.result.claim_approval",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 105: CancelCostEstimate (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "warranty",
				HandlerMethod: "CancelCostEstimate",
				InputMapping: map[string]string{
					"claimID": "$.input.claim_id",
					"costEstimate": "$.steps.5.result.cost_estimate",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: ReverseDefectResolution (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "asset",
				HandlerMethod: "ReverseDefectResolution",
				InputMapping: map[string]string{
					"claimID": "$.input.claim_id",
					"productID": "$.input.product_id",
					"defectResolution": "$.steps.6.result.defect_resolution",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 107: ReverseClaimSettlement (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "insurance",
				HandlerMethod: "ReverseClaimSettlement",
				InputMapping: map[string]string{
					"claimID": "$.input.claim_id",
					"claimSettlement": "$.steps.7.result.claim_settlement",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 108: ReverseInvoice (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "sales-invoice",
				HandlerMethod: "ReverseInvoice",
				InputMapping: map[string]string{
					"claimID": "$.input.claim_id",
					"invoiceID": "$.steps.8.result.invoice_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 109: ReverseGLPosting (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseClaimJournal",
				InputMapping: map[string]string{
					"claimID": "$.input.claim_id",
					"journalEntryID": "$.steps.9.result.journal_entry_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *WarrantyClaimSaga) SagaType() string {
	return "SAGA-W01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *WarrantyClaimSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *WarrantyClaimSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *WarrantyClaimSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["claim_id"] == nil {
		return errors.New("claim_id is required")
	}

	claimID, ok := inputMap["claim_id"].(string)
	if !ok || claimID == "" {
		return errors.New("claim_id must be a non-empty string")
	}

	if inputMap["product_id"] == nil {
		return errors.New("product_id is required")
	}

	productID, ok := inputMap["product_id"].(string)
	if !ok || productID == "" {
		return errors.New("product_id must be a non-empty string")
	}

	if inputMap["customer_id"] == nil {
		return errors.New("customer_id is required")
	}

	customerID, ok := inputMap["customer_id"].(string)
	if !ok || customerID == "" {
		return errors.New("customer_id must be a non-empty string")
	}

	if inputMap["issue_description"] == nil {
		return errors.New("issue_description is required")
	}

	issueDesc, ok := inputMap["issue_description"].(string)
	if !ok || issueDesc == "" {
		return errors.New("issue_description must be a non-empty string")
	}

	if inputMap["claim_date"] == nil {
		return errors.New("claim_date is required")
	}

	return nil
}
