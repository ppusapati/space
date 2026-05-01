// Package construction provides saga handlers for construction module workflows
package construction

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ConstructionProjectInitiationSaga implements SAGA-C01: Construction Project Initiation
// Business Flow: ValidateProjectScope → AllocateProjectBudget → SetupProjectSchedule → CreateProjectStructure → InitializeProjectAccounts → SetupApprovalHierarchy → ConfigureQualityFramework → EstablishCommunicationChannels → AuthorizeProjectStart → InitializeProjectTracker
// Timeout: 120 seconds, Critical steps: 1,2,3,4,8,10
type ConstructionProjectInitiationSaga struct {
	steps []*saga.StepDefinition
}

// NewConstructionProjectInitiationSaga creates a new Construction Project Initiation saga handler
func NewConstructionProjectInitiationSaga() saga.SagaHandler {
	return &ConstructionProjectInitiationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Validate Project Scope
			{
				StepNumber:    1,
				ServiceName:   "construction-project",
				HandlerMethod: "ValidateProjectScope",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"projectID":       "$.input.project_id",
					"projectName":     "$.input.project_name",
					"contractValue":   "$.input.contract_value",
					"startDate":       "$.input.start_date",
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
			// Step 2: Allocate Project Budget
			{
				StepNumber:    2,
				ServiceName:   "procurement",
				HandlerMethod: "AllocateProjectBudget",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"projectID":     "$.input.project_id",
					"contractValue": "$.input.contract_value",
					"scope":         "$.steps.1.result.scope_details",
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
			// Step 3: Setup Project Schedule
			{
				StepNumber:    3,
				ServiceName:   "project-planning",
				HandlerMethod: "SetupProjectSchedule",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"projectID":     "$.input.project_id",
					"projectName":   "$.input.project_name",
					"startDate":     "$.input.start_date",
					"scope":         "$.steps.1.result.scope_details",
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
			// Step 4: Create Project Structure
			{
				StepNumber:    4,
				ServiceName:   "construction-project",
				HandlerMethod: "CreateProjectStructure",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"projectID":     "$.input.project_id",
					"projectName":   "$.input.project_name",
					"scope":         "$.steps.1.result.scope_details",
					"budget":        "$.steps.2.result.budget_allocation",
					"schedule":      "$.steps.3.result.project_schedule",
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
			// Step 5: Initialize Project Accounts
			{
				StepNumber:    5,
				ServiceName:   "general-ledger",
				HandlerMethod: "InitializeProjectAccounts",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"projectID":     "$.input.project_id",
					"projectName":   "$.input.project_name",
					"contractValue": "$.input.contract_value",
					"structure":     "$.steps.4.result.project_structure",
				},
				TimeoutSeconds:    30,
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
			// Step 6: Setup Approval Hierarchy
			{
				StepNumber:    6,
				ServiceName:   "approval",
				HandlerMethod: "SetupApprovalHierarchy",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"companyID":   "$.companyID",
					"branchID":    "$.branchID",
					"projectID":   "$.input.project_id",
					"structure":   "$.steps.4.result.project_structure",
				},
				TimeoutSeconds:    20,
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
			// Step 7: Configure Quality Framework
			{
				StepNumber:    7,
				ServiceName:   "project-planning",
				HandlerMethod: "ConfigureQualityFramework",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"companyID":   "$.companyID",
					"branchID":    "$.branchID",
					"projectID":   "$.input.project_id",
					"scope":       "$.steps.1.result.scope_details",
				},
				TimeoutSeconds:    25,
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
			// Step 8: Establish Communication Channels
			{
				StepNumber:    8,
				ServiceName:   "construction-project",
				HandlerMethod: "EstablishCommunicationChannels",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"companyID":   "$.companyID",
					"branchID":    "$.branchID",
					"projectID":   "$.input.project_id",
					"structure":   "$.steps.4.result.project_structure",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Authorize Project Start
			{
				StepNumber:    9,
				ServiceName:   "approval",
				HandlerMethod: "AuthorizeProjectStart",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"projectID":    "$.input.project_id",
					"approvals":    "$.steps.6.result.approval_setup",
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
			// Step 10: Initialize Project Tracker
			{
				StepNumber:    10,
				ServiceName:   "construction-project",
				HandlerMethod: "InitializeProjectTracker",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"projectID":      "$.input.project_id",
					"projectName":    "$.input.project_name",
					"structure":      "$.steps.4.result.project_structure",
					"schedule":       "$.steps.3.result.project_schedule",
					"authorization": "$.steps.9.result.authorization_status",
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

			// Step 102: ReleaseProjectBudget (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "procurement",
				HandlerMethod: "ReleaseProjectBudget",
				InputMapping: map[string]string{
					"projectID":       "$.input.project_id",
					"budgetAllocation": "$.steps.2.result.budget_allocation",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: CancelProjectSchedule (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "project-planning",
				HandlerMethod: "CancelProjectSchedule",
				InputMapping: map[string]string{
					"projectID":     "$.input.project_id",
					"projectSchedule": "$.steps.3.result.project_schedule",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: DeleteProjectStructure (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "construction-project",
				HandlerMethod: "DeleteProjectStructure",
				InputMapping: map[string]string{
					"projectID":        "$.input.project_id",
					"projectStructure": "$.steps.4.result.project_structure",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: ReverseProjectAccounts (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseProjectAccounts",
				InputMapping: map[string]string{
					"projectID":        "$.input.project_id",
					"projectAccounts": "$.steps.5.result.project_accounts",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: RemoveApprovalHierarchy (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "approval",
				HandlerMethod: "RemoveApprovalHierarchy",
				InputMapping: map[string]string{
					"projectID":      "$.input.project_id",
					"approvalSetup": "$.steps.6.result.approval_setup",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: RemoveQualityFramework (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "project-planning",
				HandlerMethod: "RemoveQualityFramework",
				InputMapping: map[string]string{
					"projectID":        "$.input.project_id",
					"qualityFramework": "$.steps.7.result.quality_framework",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 108: DisconnectCommunicationChannels (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "construction-project",
				HandlerMethod: "DisconnectCommunicationChannels",
				InputMapping: map[string]string{
					"projectID":      "$.input.project_id",
					"communicationChannels": "$.steps.8.result.communication_channels",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 109: RevokeProjectStart (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "approval",
				HandlerMethod: "RevokeProjectStart",
				InputMapping: map[string]string{
					"projectID":              "$.input.project_id",
					"authorizationStatus": "$.steps.9.result.authorization_status",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ConstructionProjectInitiationSaga) SagaType() string {
	return "SAGA-C01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ConstructionProjectInitiationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ConstructionProjectInitiationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ConstructionProjectInitiationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["project_id"] == nil {
		return errors.New("project_id is required")
	}

	projectID, ok := inputMap["project_id"].(string)
	if !ok || projectID == "" {
		return errors.New("project_id must be a non-empty string")
	}

	if inputMap["project_name"] == nil {
		return errors.New("project_name is required")
	}

	projectName, ok := inputMap["project_name"].(string)
	if !ok || projectName == "" {
		return errors.New("project_name must be a non-empty string")
	}

	if inputMap["contract_value"] == nil {
		return errors.New("contract_value is required")
	}

	contractValue, ok := inputMap["contract_value"].(string)
	if !ok || contractValue == "" {
		return errors.New("contract_value must be a non-empty string")
	}

	if inputMap["start_date"] == nil {
		return errors.New("start_date is required")
	}

	startDate, ok := inputMap["start_date"].(string)
	if !ok || startDate == "" {
		return errors.New("start_date must be a non-empty string")
	}

	return nil
}
