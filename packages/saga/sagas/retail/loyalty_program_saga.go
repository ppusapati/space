// Package retail provides saga handlers for retail workflows
package retail

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// LoyaltyProgramSaga implements SAGA-R04: Loyalty Program Enrollment & Management workflow
// Business Flow: InitiateEnrollment → ValidateCustomerEligibility → CreateLoyaltyAccount → RegisterInProgram → AllocateSignupBonus → UpdateCustomerProfile → ApplyLoyaltyJournal → CompleteLoyaltyEnrollment
// Steps: 8 forward + 9 compensation = 17 total
// Timeout: 120 seconds, Critical steps: 1,2,3,5,8,9
type LoyaltyProgramSaga struct {
	steps []*saga.StepDefinition
}

// NewLoyaltyProgramSaga creates a new Loyalty Program saga handler
func NewLoyaltyProgramSaga() saga.SagaHandler {
	return &LoyaltyProgramSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Enrollment
			{
				StepNumber:    1,
				ServiceName:   "loyalty",
				HandlerMethod: "InitiateEnrollment",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"enrollmentID":  "$.input.enrollment_id",
					"customerID":    "$.input.customer_id",
					"programType":   "$.input.program_type",
					"enrollmentDate": "$.input.enrollment_date",
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
			// Step 2: Validate Customer Eligibility
			{
				StepNumber:    2,
				ServiceName:   "customer",
				HandlerMethod: "ValidateCustomerEligibility",
				InputMapping: map[string]string{
					"enrollmentID": "$.steps.1.result.enrollment_id",
					"customerID":   "$.input.customer_id",
					"programType":  "$.input.program_type",
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
			// Step 3: Create Loyalty Account
			{
				StepNumber:    3,
				ServiceName:   "loyalty",
				HandlerMethod: "CreateLoyaltyAccount",
				InputMapping: map[string]string{
					"enrollmentID": "$.steps.1.result.enrollment_id",
					"customerID":   "$.input.customer_id",
					"programType":  "$.input.program_type",
				},
				TimeoutSeconds:    25,
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
			// Step 4: Register in Program
			{
				StepNumber:    4,
				ServiceName:   "loyalty",
				HandlerMethod: "RegisterInProgram",
				InputMapping: map[string]string{
					"enrollmentID": "$.steps.1.result.enrollment_id",
					"loyaltyAccountID": "$.steps.3.result.loyalty_account_id",
					"customerID":    "$.input.customer_id",
					"programType":   "$.input.program_type",
				},
				TimeoutSeconds:    25,
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
			// Step 5: Allocate Signup Bonus
			{
				StepNumber:    5,
				ServiceName:   "loyalty",
				HandlerMethod: "AllocateSignupBonus",
				InputMapping: map[string]string{
					"enrollmentID": "$.steps.1.result.enrollment_id",
					"loyaltyAccountID": "$.steps.3.result.loyalty_account_id",
					"programType":   "$.input.program_type",
					"bonusPoints":   "$.input.bonus_points",
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
			// Step 6: Update Customer Profile
			{
				StepNumber:    6,
				ServiceName:   "customer",
				HandlerMethod: "UpdateCustomerProfile",
				InputMapping: map[string]string{
					"enrollmentID": "$.steps.1.result.enrollment_id",
					"customerID":   "$.input.customer_id",
					"loyaltyAccountID": "$.steps.3.result.loyalty_account_id",
					"loyaltyStatus": "Active",
				},
				TimeoutSeconds:    20,
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
			// Step 7: Apply Loyalty Journal
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "ApplyLoyaltyJournal",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"enrollmentID":  "$.steps.1.result.enrollment_id",
					"loyaltyAccountID": "$.steps.3.result.loyalty_account_id",
					"bonusPoints":   "$.input.bonus_points",
					"journalDate":   "$.input.enrollment_date",
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
			// Step 8: Complete Loyalty Enrollment
			{
				StepNumber:    8,
				ServiceName:   "loyalty",
				HandlerMethod: "CompleteLoyaltyEnrollment",
				InputMapping: map[string]string{
					"enrollmentID": "$.steps.1.result.enrollment_id",
					"loyaltyAccountID": "$.steps.3.result.loyalty_account_id",
					"enrollmentStatus": "Completed",
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

			// Step 101: Revert Loyalty Account Creation (compensates step 3)
			{
				StepNumber:    101,
				ServiceName:   "loyalty",
				HandlerMethod: "RevertLoyaltyAccountCreation",
				InputMapping: map[string]string{
					"enrollmentID": "$.steps.1.result.enrollment_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 102: Revert Program Registration (compensates step 4)
			{
				StepNumber:    102,
				ServiceName:   "loyalty",
				HandlerMethod: "RevertProgramRegistration",
				InputMapping: map[string]string{
					"enrollmentID": "$.steps.1.result.enrollment_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: Revert Signup Bonus Allocation (compensates step 5)
			{
				StepNumber:    103,
				ServiceName:   "loyalty",
				HandlerMethod: "RevertSignupBonusAllocation",
				InputMapping: map[string]string{
					"enrollmentID": "$.steps.1.result.enrollment_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 104: Revert Customer Profile Update (compensates step 6)
			{
				StepNumber:    104,
				ServiceName:   "customer",
				HandlerMethod: "RevertCustomerProfileUpdate",
				InputMapping: map[string]string{
					"enrollmentID": "$.steps.1.result.enrollment_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 105: Reverse Loyalty Journal (compensates step 7)
			{
				StepNumber:    105,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseLoyaltyJournal",
				InputMapping: map[string]string{
					"enrollmentID": "$.steps.1.result.enrollment_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Revert Initiate Enrollment (compensates step 1)
			{
				StepNumber:    106,
				ServiceName:   "loyalty",
				HandlerMethod: "RevertInitiateEnrollment",
				InputMapping: map[string]string{
					"enrollmentID": "$.steps.1.result.enrollment_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: Revert Validate Customer Eligibility (compensates step 2)
			{
				StepNumber:    107,
				ServiceName:   "customer",
				HandlerMethod: "RevertValidateCustomerEligibility",
				InputMapping: map[string]string{
					"enrollmentID": "$.steps.1.result.enrollment_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 108: Revert Complete Loyalty Enrollment (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "loyalty",
				HandlerMethod: "RevertCompleteLoyaltyEnrollment",
				InputMapping: map[string]string{
					"enrollmentID": "$.steps.1.result.enrollment_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 109: Revert Loyalty Account Deletion (additional compensation)
			{
				StepNumber:    109,
				ServiceName:   "loyalty",
				HandlerMethod: "DeleteLoyaltyAccount",
				InputMapping: map[string]string{
					"enrollmentID": "$.steps.1.result.enrollment_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *LoyaltyProgramSaga) SagaType() string {
	return "SAGA-R04"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *LoyaltyProgramSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *LoyaltyProgramSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *LoyaltyProgramSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["enrollment_id"] == nil {
		return errors.New("enrollment_id is required")
	}

	if inputMap["customer_id"] == nil {
		return errors.New("customer_id is required")
	}

	if inputMap["program_type"] == nil {
		return errors.New("program_type is required")
	}

	return nil
}
