// Package warranty provides saga handlers for warranty and service module workflows
package warranty

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ExtendedWarrantySaga implements SAGA-W06: Extended Warranty Plans workflow
// Business Flow: SubscribePlan → VerifyCoverage → LinkProduct → ActivatePlan →
// SetupBillingSchedule → GenerateInvoice → ProcessPayment → PostGL → TrackCoverage →
// SetupClaimLimit → NotifyActivation → ManagePlan
// Timeout: 180 seconds, Critical steps: 1,3,5,7,8,10
type ExtendedWarrantySaga struct {
	steps []*saga.StepDefinition
}

// NewExtendedWarrantySaga creates a new Extended Warranty Plans saga handler (SAGA-W06)
func NewExtendedWarrantySaga() saga.SagaHandler {
	return &ExtendedWarrantySaga{
		steps: []*saga.StepDefinition{
			// Step 1: Subscribe Plan
			{
				StepNumber:    1,
				ServiceName:   "warranty-plan",
				HandlerMethod: "SubscribePlan",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"planID":        "$.input.plan_id",
					"customerID":    "$.input.customer_id",
					"planName":      "$.input.plan_name",
					"subscriptionDate": "$.input.subscription_date",
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
			// Step 2: Verify Coverage
			{
				StepNumber:    2,
				ServiceName:   "warranty-plan",
				HandlerMethod: "VerifyCoverage",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"planID":        "$.input.plan_id",
					"customerID":    "$.input.customer_id",
					"planSubscription": "$.steps.1.result.plan_subscription",
				},
				TimeoutSeconds: 45,
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
			// Step 3: Link Product
			{
				StepNumber:    3,
				ServiceName:   "sales-order",
				HandlerMethod: "LinkProduct",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"planID":        "$.input.plan_id",
					"productID":     "$.input.product_id",
					"coverageVerification": "$.steps.2.result.coverage_verification",
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
			// Step 4: Activate Plan
			{
				StepNumber:    4,
				ServiceName:   "warranty-plan",
				HandlerMethod: "ActivatePlan",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"planID":        "$.input.plan_id",
					"productLinking": "$.steps.3.result.product_linking",
					"activationDate": "$.input.activation_date",
				},
				TimeoutSeconds: 30,
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
			// Step 5: Setup Billing Schedule
			{
				StepNumber:    5,
				ServiceName:   "billing",
				HandlerMethod: "SetupBillingSchedule",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"planID":        "$.input.plan_id",
					"customerID":    "$.input.customer_id",
					"planActivation": "$.steps.4.result.plan_activation",
					"billingFrequency": "$.input.billing_frequency",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Generate Invoice
			{
				StepNumber:    6,
				ServiceName:   "sales-invoice",
				HandlerMethod: "GenerateWarrantyInvoice",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"planID":        "$.input.plan_id",
					"customerID":    "$.input.customer_id",
					"billingSchedule": "$.steps.5.result.billing_schedule",
					"invoiceDate":   "$.input.subscription_date",
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
			// Step 7: Process Payment
			{
				StepNumber:    7,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ProcessWarrantyPayment",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"planID":        "$.input.plan_id",
					"customerID":    "$.input.customer_id",
					"invoiceID":     "$.steps.6.result.invoice_id",
					"paymentMethod": "$.input.payment_method",
				},
				TimeoutSeconds: 60,
				IsCritical:     true,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Post GL (General Ledger)
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostWarrantyJournal",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"planID":        "$.input.plan_id",
					"invoiceID":     "$.steps.6.result.invoice_id",
					"paymentProcessing": "$.steps.7.result.payment_processing",
					"journalDate":   "$.input.subscription_date",
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
			// Step 9: Track Coverage
			{
				StepNumber:    9,
				ServiceName:   "warranty-plan",
				HandlerMethod: "TrackCoverage",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"planID":        "$.input.plan_id",
					"productID":     "$.input.product_id",
					"glPosting":     "$.steps.8.result.journal_entries",
					"trackingStartDate": "$.input.activation_date",
				},
				TimeoutSeconds: 30,
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
			// Step 10: Setup Claim Limit
			{
				StepNumber:    10,
				ServiceName:   "warranty-plan",
				HandlerMethod: "SetupClaimLimit",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"planID":        "$.input.plan_id",
					"coverageTracking": "$.steps.9.result.coverage_tracking",
					"claimLimit":    "$.input.claim_limit",
					"limitType":     "$.input.limit_type",
				},
				TimeoutSeconds: 30,
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
			// Step 11: Notify Activation
			{
				StepNumber:    11,
				ServiceName:   "notification",
				HandlerMethod: "NotifyPlanActivation",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"planID":        "$.input.plan_id",
					"customerID":    "$.input.customer_id",
					"claimLimitSetup": "$.steps.10.result.claim_limit_setup",
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
			// Step 12: Manage Plan
			{
				StepNumber:    12,
				ServiceName:   "warranty-plan",
				HandlerMethod: "ManagePlan",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"planID":        "$.input.plan_id",
					"customerID":    "$.input.customer_id",
					"notificationRecord": "$.steps.11.result.notification_record",
					"managementDate": "$.input.subscription_date",
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

			// Step 102: CancelCoverageVerification (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "warranty-plan",
				HandlerMethod: "CancelCoverageVerification",
				InputMapping: map[string]string{
					"planID": "$.input.plan_id",
					"coverageVerification": "$.steps.2.result.coverage_verification",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: UnlinkProduct (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "sales-order",
				HandlerMethod: "UnlinkProduct",
				InputMapping: map[string]string{
					"planID": "$.input.plan_id",
					"productID": "$.input.product_id",
					"productLinking": "$.steps.3.result.product_linking",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: DeactivatePlan (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "warranty-plan",
				HandlerMethod: "DeactivatePlan",
				InputMapping: map[string]string{
					"planID": "$.input.plan_id",
					"planActivation": "$.steps.4.result.plan_activation",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: CancelBillingSchedule (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "billing",
				HandlerMethod: "CancelBillingSchedule",
				InputMapping: map[string]string{
					"planID": "$.input.plan_id",
					"billingSchedule": "$.steps.5.result.billing_schedule",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: CancelWarrantyInvoice (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "sales-invoice",
				HandlerMethod: "CancelWarrantyInvoice",
				InputMapping: map[string]string{
					"planID": "$.input.plan_id",
					"invoiceID": "$.steps.6.result.invoice_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 107: ReversePaymentProcessing (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ReverseWarrantyPayment",
				InputMapping: map[string]string{
					"planID": "$.input.plan_id",
					"paymentProcessing": "$.steps.7.result.payment_processing",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 108: ReverseWarrantyJournal (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseWarrantyJournal",
				InputMapping: map[string]string{
					"planID": "$.input.plan_id",
					"journalEntryID": "$.steps.8.result.journal_entry_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 109: CancelCoverageTracking (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "warranty-plan",
				HandlerMethod: "CancelCoverageTracking",
				InputMapping: map[string]string{
					"planID": "$.input.plan_id",
					"coverageTracking": "$.steps.9.result.coverage_tracking",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 110: CancelClaimLimitSetup (compensates step 10)
			{
				StepNumber:    110,
				ServiceName:   "warranty-plan",
				HandlerMethod: "CancelClaimLimitSetup",
				InputMapping: map[string]string{
					"planID": "$.input.plan_id",
					"claimLimitSetup": "$.steps.10.result.claim_limit_setup",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ExtendedWarrantySaga) SagaType() string {
	return "SAGA-W06"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ExtendedWarrantySaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ExtendedWarrantySaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ExtendedWarrantySaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["plan_id"] == nil {
		return errors.New("plan_id is required")
	}

	planID, ok := inputMap["plan_id"].(string)
	if !ok || planID == "" {
		return errors.New("plan_id must be a non-empty string")
	}

	if inputMap["customer_id"] == nil {
		return errors.New("customer_id is required")
	}

	customerID, ok := inputMap["customer_id"].(string)
	if !ok || customerID == "" {
		return errors.New("customer_id must be a non-empty string")
	}

	if inputMap["product_id"] == nil {
		return errors.New("product_id is required")
	}

	productID, ok := inputMap["product_id"].(string)
	if !ok || productID == "" {
		return errors.New("product_id must be a non-empty string")
	}

	if inputMap["subscription_date"] == nil {
		return errors.New("subscription_date is required")
	}

	return nil
}
