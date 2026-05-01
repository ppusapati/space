// Package retail provides saga handlers for retail workflows
package retail

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// CustomerAccountSaga implements SAGA-R07: Customer Account Management workflow
// Business Flow: InitiateAccountSetup → ValidateCustomerData → CreateCustomerProfile → RegisterContactInfo → InitializeAccountLedger → UpdatePOSProfile → ApplyAccountJournal → CompleteAccountSetup
// Steps: 8 forward + 7 compensation = 15 total
// Timeout: 120 seconds, Critical steps: 1,2,3,5,7,8
type CustomerAccountSaga struct {
	steps []*saga.StepDefinition
}

// NewCustomerAccountSaga creates a new Customer Account Management saga handler
func NewCustomerAccountSaga() saga.SagaHandler {
	return &CustomerAccountSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Account Setup
			{
				StepNumber:    1,
				ServiceName:   "customer",
				HandlerMethod: "InitiateAccountSetup",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"companyID":   "$.companyID",
					"branchID":    "$.branchID",
					"customerID":  "$.input.customer_id",
					"email":       "$.input.email",
					"phone":       "$.input.phone",
					"setupDate":   "$.input.setup_date",
				},
				TimeoutSeconds: 20,
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
			// Step 2: Validate Customer Data
			{
				StepNumber:    2,
				ServiceName:   "customer",
				HandlerMethod: "ValidateCustomerData",
				InputMapping: map[string]string{
					"customerID": "$.steps.1.result.customer_id",
					"email":      "$.input.email",
					"phone":      "$.input.phone",
					"validateDuplicates": "true",
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
			// Step 3: Create Customer Profile
			{
				StepNumber:    3,
				ServiceName:   "customer",
				HandlerMethod: "CreateCustomerProfile",
				InputMapping: map[string]string{
					"customerID": "$.steps.1.result.customer_id",
					"email":      "$.input.email",
					"phone":      "$.input.phone",
					"address":    "$.input.address",
					"city":       "$.input.city",
				},
				TimeoutSeconds:    20,
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
			// Step 4: Register Contact Information
			{
				StepNumber:    4,
				ServiceName:   "customer",
				HandlerMethod: "RegisterContactInfo",
				InputMapping: map[string]string{
					"customerID": "$.steps.1.result.customer_id",
					"email":      "$.input.email",
					"phone":      "$.input.phone",
					"altPhone":   "$.input.alternate_phone",
				},
				TimeoutSeconds:    15,
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
			// Step 5: Initialize Account Ledger
			{
				StepNumber:    5,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "InitializeAccountLedger",
				InputMapping: map[string]string{
					"customerID":  "$.steps.1.result.customer_id",
					"creditLimit": "$.input.credit_limit",
					"paymentTerms": "$.input.payment_terms",
				},
				TimeoutSeconds:    20,
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
			// Step 6: Update POS Profile
			{
				StepNumber:    6,
				ServiceName:   "pos",
				HandlerMethod: "UpdatePOSProfile",
				InputMapping: map[string]string{
					"customerID": "$.steps.1.result.customer_id",
					"email":      "$.input.email",
					"phone":      "$.input.phone",
				},
				TimeoutSeconds:    15,
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
			// Step 7: Apply Account Journal
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "ApplyAccountJournal",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"customerID":    "$.steps.1.result.customer_id",
					"creditLimit":   "$.input.credit_limit",
					"journalDate":   "$.input.setup_date",
				},
				TimeoutSeconds:    20,
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
			// Step 8: Complete Account Setup
			{
				StepNumber:    8,
				ServiceName:   "customer",
				HandlerMethod: "CompleteAccountSetup",
				InputMapping: map[string]string{
					"customerID":       "$.steps.1.result.customer_id",
					"journalEntries":   "$.steps.7.result.journal_entries",
					"setupStatus":      "Completed",
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

			// Step 101: Revert Customer Profile Creation (compensates step 3)
			{
				StepNumber:    101,
				ServiceName:   "customer",
				HandlerMethod: "RevertCustomerProfileCreation",
				InputMapping: map[string]string{
					"customerID": "$.steps.1.result.customer_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 102: Revert Contact Information Registration (compensates step 4)
			{
				StepNumber:    102,
				ServiceName:   "customer",
				HandlerMethod: "RevertContactInfoRegistration",
				InputMapping: map[string]string{
					"customerID": "$.steps.1.result.customer_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 103: Revert Account Ledger Initialization (compensates step 5)
			{
				StepNumber:    103,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "RevertAccountLedgerInitialization",
				InputMapping: map[string]string{
					"customerID": "$.steps.1.result.customer_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 104: Revert POS Profile Update (compensates step 6)
			{
				StepNumber:    104,
				ServiceName:   "pos",
				HandlerMethod: "RevertPOSProfileUpdate",
				InputMapping: map[string]string{
					"customerID": "$.steps.1.result.customer_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 105: Reverse Account Journal (compensates step 7)
			{
				StepNumber:    105,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseAccountJournal",
				InputMapping: map[string]string{
					"customerID": "$.steps.1.result.customer_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: Revert Initiate Account Setup (compensates step 1)
			{
				StepNumber:    106,
				ServiceName:   "customer",
				HandlerMethod: "RevertInitiateAccountSetup",
				InputMapping: map[string]string{
					"customerID": "$.steps.1.result.customer_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 107: Revert Complete Account Setup (compensates step 8)
			{
				StepNumber:    107,
				ServiceName:   "customer",
				HandlerMethod: "RevertCompleteAccountSetup",
				InputMapping: map[string]string{
					"customerID": "$.steps.1.result.customer_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *CustomerAccountSaga) SagaType() string {
	return "SAGA-R07"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *CustomerAccountSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *CustomerAccountSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *CustomerAccountSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["customer_id"] == nil {
		return errors.New("customer_id is required")
	}

	if inputMap["email"] == nil {
		return errors.New("email is required")
	}

	if inputMap["phone"] == nil {
		return errors.New("phone is required")
	}

	return nil
}
