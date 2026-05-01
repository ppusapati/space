// Package healthcare provides saga handlers for healthcare workflows
package healthcare

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ProviderNetworkSaga implements SAGA-HC06: Healthcare Provider Network Management workflow
// Business Flow: CreateProviderProfile → ValidateProviderCredentials → LinkToNetwork → SetupContractTerms → InitializePaymentTerms → ApplyNetworkJournal → UpdateProviderRecords → CompleteProviderNetwork
// Steps: 8 forward + 7 compensation = 15 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,6,8
type ProviderNetworkSaga struct {
	steps []*saga.StepDefinition
}

// NewProviderNetworkSaga creates a new Provider Network saga handler
func NewProviderNetworkSaga() saga.SagaHandler {
	return &ProviderNetworkSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Provider Profile
			{
				StepNumber:    1,
				ServiceName:   "provider-network",
				HandlerMethod: "CreateProviderProfile",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"providerID":    "$.input.provider_id",
					"providerName":  "$.input.provider_name",
					"networkID":     "$.input.network_id",
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
			// Step 2: Validate Provider Credentials
			{
				StepNumber:    2,
				ServiceName:   "provider-network",
				HandlerMethod: "ValidateProviderCredentials",
				InputMapping: map[string]string{
					"providerID":      "$.steps.1.result.provider_id",
					"providerName":    "$.input.provider_name",
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
			// Step 3: Link to Network
			{
				StepNumber:    3,
				ServiceName:   "insurance",
				HandlerMethod: "LinkProviderToNetwork",
				InputMapping: map[string]string{
					"providerID":         "$.steps.1.result.provider_id",
					"providerName":       "$.input.provider_name",
					"networkID":          "$.input.network_id",
					"credentialData":     "$.steps.2.result.credential_data",
				},
				TimeoutSeconds:    30,
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
			// Step 4: Setup Contract Terms
			{
				StepNumber:    4,
				ServiceName:   "contract-management",
				HandlerMethod: "SetupContractTerms",
				InputMapping: map[string]string{
					"providerID":       "$.steps.1.result.provider_id",
					"providerName":     "$.input.provider_name",
					"networkID":        "$.input.network_id",
					"networkLinkData":  "$.steps.3.result.network_link_data",
				},
				TimeoutSeconds:    35,
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
			// Step 5: Initialize Payment Terms
			{
				StepNumber:    5,
				ServiceName:   "accounts-payable",
				HandlerMethod: "InitializePaymentTerms",
				InputMapping: map[string]string{
					"providerID":      "$.steps.1.result.provider_id",
					"providerName":    "$.input.provider_name",
					"contractTerms":   "$.steps.4.result.contract_terms",
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
			// Step 6: Apply Network Journal
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "ApplyNetworkJournal",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"providerID":    "$.steps.1.result.provider_id",
					"contractTerms": "$.steps.4.result.contract_terms",
					"journalDate":   "$.input.network_id",
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
			// Step 7: Update Provider Records
			{
				StepNumber:    7,
				ServiceName:   "provider-network",
				HandlerMethod: "UpdateProviderRecords",
				InputMapping: map[string]string{
					"providerID":        "$.steps.1.result.provider_id",
					"providerName":      "$.input.provider_name",
					"networkID":         "$.input.network_id",
					"contractTerms":     "$.steps.4.result.contract_terms",
					"paymentTerms":      "$.steps.5.result.payment_terms",
					"journalEntries":    "$.steps.6.result.journal_entries",
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
			// Step 8: Complete Provider Network
			{
				StepNumber:    8,
				ServiceName:   "provider-network",
				HandlerMethod: "CompleteProviderNetwork",
				InputMapping: map[string]string{
					"providerID":        "$.steps.1.result.provider_id",
					"providerName":      "$.input.provider_name",
					"networkID":         "$.input.network_id",
					"contractTerms":     "$.steps.4.result.contract_terms",
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

			// Step 101: Unlink Provider from Network (compensates step 3)
			{
				StepNumber:    101,
				ServiceName:   "insurance",
				HandlerMethod: "UnlinkProviderFromNetwork",
				InputMapping: map[string]string{
					"providerID": "$.steps.1.result.provider_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Revert Contract Terms Setup (compensates step 4)
			{
				StepNumber:    102,
				ServiceName:   "contract-management",
				HandlerMethod: "RevertContractTermsSetup",
				InputMapping: map[string]string{
					"providerID": "$.steps.1.result.provider_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 103: Revert Payment Terms Initialization (compensates step 5)
			{
				StepNumber:    103,
				ServiceName:   "accounts-payable",
				HandlerMethod: "RevertPaymentTermsInitialization",
				InputMapping: map[string]string{
					"providerID": "$.steps.1.result.provider_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Reverse Network Journal (compensates step 6)
			{
				StepNumber:    104,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseNetworkJournal",
				InputMapping: map[string]string{
					"providerID": "$.steps.1.result.provider_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: Revert Provider Records Update (compensates step 7)
			{
				StepNumber:    105,
				ServiceName:   "provider-network",
				HandlerMethod: "RevertProviderRecordsUpdate",
				InputMapping: map[string]string{
					"providerID": "$.steps.1.result.provider_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ProviderNetworkSaga) SagaType() string {
	return "SAGA-HC06"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ProviderNetworkSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ProviderNetworkSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ProviderNetworkSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["provider_id"] == nil {
		return errors.New("provider_id is required")
	}

	if inputMap["provider_name"] == nil {
		return errors.New("provider_name is required")
	}

	if inputMap["network_id"] == nil {
		return errors.New("network_id is required")
	}

	return nil
}
