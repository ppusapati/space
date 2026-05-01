// Package healthcare provides saga handlers for healthcare workflows
package healthcare

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// PatientRegistrationSaga implements SAGA-HC01: Patient Registration & Medical Records workflow
// Business Flow: CreatePatientRecord → ValidatePatientInfo → InitializePatientFile → RegisterInsurance → CreatePatientAccount → LinkMedicalHistory → SetupCompliance → VerifyDocumentation → CompleteRegistration
// Steps: 9 forward + 8 compensation = 17 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,6,9
type PatientRegistrationSaga struct {
	steps []*saga.StepDefinition
}

// NewPatientRegistrationSaga creates a new Patient Registration saga handler
func NewPatientRegistrationSaga() saga.SagaHandler {
	return &PatientRegistrationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Patient Record
			{
				StepNumber:    1,
				ServiceName:   "patient-management",
				HandlerMethod: "CreatePatientRecord",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"companyID":  "$.companyID",
					"branchID":   "$.branchID",
					"patientID":  "$.input.patient_id",
					"firstName":  "$.input.first_name",
					"lastName":   "$.input.last_name",
					"dob":        "$.input.dob",
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
			// Step 2: Validate Patient Information
			{
				StepNumber:    2,
				ServiceName:   "patient-management",
				HandlerMethod: "ValidatePatientInfo",
				InputMapping: map[string]string{
					"patientID":    "$.steps.1.result.patient_id",
					"firstName":    "$.input.first_name",
					"lastName":     "$.input.last_name",
					"dob":          "$.input.dob",
					"validateRules": "true",
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
			// Step 3: Initialize Patient File
			{
				StepNumber:    3,
				ServiceName:   "medical-records",
				HandlerMethod: "InitializePatientFile",
				InputMapping: map[string]string{
					"patientID":          "$.steps.1.result.patient_id",
					"firstName":          "$.input.first_name",
					"lastName":           "$.input.last_name",
					"dob":                "$.input.dob",
					"validationResult":   "$.steps.2.result.validation_result",
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
			// Step 4: Register Insurance
			{
				StepNumber:    4,
				ServiceName:   "insurance",
				HandlerMethod: "RegisterInsurance",
				InputMapping: map[string]string{
					"patientID":     "$.steps.1.result.patient_id",
					"firstName":     "$.input.first_name",
					"lastName":      "$.input.last_name",
					"medicalFile":   "$.steps.3.result.medical_file",
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
			// Step 5: Create Patient Account
			{
				StepNumber:    5,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "CreatePatientAccount",
				InputMapping: map[string]string{
					"patientID":        "$.steps.1.result.patient_id",
					"firstName":        "$.input.first_name",
					"lastName":         "$.input.last_name",
					"insuranceData":    "$.steps.4.result.insurance_data",
				},
				TimeoutSeconds:    25,
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
			// Step 6: Link Medical History
			{
				StepNumber:    6,
				ServiceName:   "medical-records",
				HandlerMethod: "LinkMedicalHistory",
				InputMapping: map[string]string{
					"patientID":    "$.steps.1.result.patient_id",
					"medicalFile": "$.steps.3.result.medical_file",
					"dob":         "$.input.dob",
				},
				TimeoutSeconds:    25,
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
			// Step 7: Setup Compliance
			{
				StepNumber:    7,
				ServiceName:   "compliance",
				HandlerMethod: "SetupCompliance",
				InputMapping: map[string]string{
					"patientID":       "$.steps.1.result.patient_id",
					"firstName":       "$.input.first_name",
					"lastName":        "$.input.last_name",
					"complianceLevel": "Standard",
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
			// Step 8: Verify Documentation
			{
				StepNumber:    8,
				ServiceName:   "medical-records",
				HandlerMethod: "VerifyDocumentation",
				InputMapping: map[string]string{
					"patientID":        "$.steps.1.result.patient_id",
					"medicalFile":      "$.steps.3.result.medical_file",
					"complianceStatus": "$.steps.7.result.compliance_status",
				},
				TimeoutSeconds:    20,
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
			// Step 9: Complete Registration
			{
				StepNumber:    9,
				ServiceName:   "patient-management",
				HandlerMethod: "CompleteRegistration",
				InputMapping: map[string]string{
					"patientID":              "$.steps.1.result.patient_id",
					"medicalFile":           "$.steps.3.result.medical_file",
					"accountData":           "$.steps.5.result.account_data",
					"complianceStatus":      "$.steps.7.result.compliance_status",
					"documentationVerified": "$.steps.8.result.verification_status",
					"registrationStatus":    "Completed",
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

			// Step 101: Revert Patient Validation (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "patient-management",
				HandlerMethod: "RevertPatientValidation",
				InputMapping: map[string]string{
					"patientID": "$.steps.1.result.patient_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Revert Patient File Initialization (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "medical-records",
				HandlerMethod: "RevertPatientFileInitialization",
				InputMapping: map[string]string{
					"patientID": "$.steps.1.result.patient_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Unregister Insurance (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "insurance",
				HandlerMethod: "UnregisterInsurance",
				InputMapping: map[string]string{
					"patientID": "$.steps.1.result.patient_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Revert Patient Account Creation (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "RevertPatientAccountCreation",
				InputMapping: map[string]string{
					"patientID": "$.steps.1.result.patient_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Unlink Medical History (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "medical-records",
				HandlerMethod: "UnlinkMedicalHistory",
				InputMapping: map[string]string{
					"patientID": "$.steps.1.result.patient_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Revert Compliance Setup (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "compliance",
				HandlerMethod: "RevertComplianceSetup",
				InputMapping: map[string]string{
					"patientID": "$.steps.1.result.patient_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 107: Revert Documentation Verification (compensates step 8)
			{
				StepNumber:    107,
				ServiceName:   "medical-records",
				HandlerMethod: "RevertDocumentationVerification",
				InputMapping: map[string]string{
					"patientID": "$.steps.1.result.patient_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *PatientRegistrationSaga) SagaType() string {
	return "SAGA-HC01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *PatientRegistrationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *PatientRegistrationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *PatientRegistrationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["patient_id"] == nil {
		return errors.New("patient_id is required")
	}

	if inputMap["first_name"] == nil {
		return errors.New("first_name is required")
	}

	if inputMap["last_name"] == nil {
		return errors.New("last_name is required")
	}

	if inputMap["dob"] == nil {
		return errors.New("dob is required (format: YYYY-MM-DD)")
	}

	return nil
}
