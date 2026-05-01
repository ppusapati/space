// Package healthcare provides saga handlers for healthcare workflows
package healthcare

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ServiceDeliverySaga implements SAGA-HC02: Healthcare Service Delivery & Billing workflow
// Business Flow: ValidateAppointment → RetrievePatientRecord → ScheduleService → InitializeServiceRecord → ExecuteService → RecordServiceDetails → CalculateServiceCharge → GenerateBillingDocument → ProcessInsuranceClaim → UpdatePatientAccount → ApplyServiceJournal → LinkComplianceRecord → CompleteBillingProcess
// Steps: 13 forward + 8 compensation = 21 total (largest healthcare saga)
// Timeout: 120 seconds, Critical steps: 1,2,3,4,6,8,11
type ServiceDeliverySaga struct {
	steps []*saga.StepDefinition
}

// NewServiceDeliverySaga creates a new Healthcare Service Delivery saga handler
func NewServiceDeliverySaga() saga.SagaHandler {
	return &ServiceDeliverySaga{
		steps: []*saga.StepDefinition{
			// Step 1: Validate Appointment
			{
				StepNumber:    1,
				ServiceName:   "service-delivery",
				HandlerMethod: "ValidateAppointment",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"appointmentID": "$.input.appointment_id",
					"patientID":     "$.input.patient_id",
					"serviceDate":   "$.input.service_date",
					"providerID":    "$.input.provider_id",
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
			// Step 2: Retrieve Patient Record
			{
				StepNumber:    2,
				ServiceName:   "medical-records",
				HandlerMethod: "RetrievePatientRecord",
				InputMapping: map[string]string{
					"patientID":      "$.input.patient_id",
					"appointmentID":  "$.steps.1.result.appointment_id",
					"retrieveActive": "true",
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
			// Step 3: Schedule Service
			{
				StepNumber:    3,
				ServiceName:   "service-delivery",
				HandlerMethod: "ScheduleService",
				InputMapping: map[string]string{
					"appointmentID":  "$.steps.1.result.appointment_id",
					"patientID":      "$.input.patient_id",
					"serviceDate":    "$.input.service_date",
					"providerID":     "$.input.provider_id",
					"patientRecord":  "$.steps.2.result.patient_record",
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
			// Step 4: Initialize Service Record
			{
				StepNumber:    4,
				ServiceName:   "medical-records",
				HandlerMethod: "InitializeServiceRecord",
				InputMapping: map[string]string{
					"appointmentID":   "$.steps.1.result.appointment_id",
					"patientID":       "$.input.patient_id",
					"serviceDate":     "$.input.service_date",
					"providerID":      "$.input.provider_id",
					"serviceSchedule": "$.steps.3.result.service_schedule",
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
			// Step 5: Execute Service
			{
				StepNumber:    5,
				ServiceName:   "service-delivery",
				HandlerMethod: "ExecuteService",
				InputMapping: map[string]string{
					"appointmentID":   "$.steps.1.result.appointment_id",
					"patientID":       "$.input.patient_id",
					"providerID":      "$.input.provider_id",
					"serviceDate":     "$.input.service_date",
					"serviceRecord":   "$.steps.4.result.service_record",
				},
				TimeoutSeconds:    45,
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
			// Step 6: Record Service Details
			{
				StepNumber:    6,
				ServiceName:   "medical-records",
				HandlerMethod: "RecordServiceDetails",
				InputMapping: map[string]string{
					"appointmentID": "$.steps.1.result.appointment_id",
					"patientID":     "$.input.patient_id",
					"serviceData":   "$.steps.5.result.service_data",
					"serviceRecord": "$.steps.4.result.service_record",
				},
				TimeoutSeconds:    25,
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
			// Step 7: Calculate Service Charge
			{
				StepNumber:    7,
				ServiceName:   "billing",
				HandlerMethod: "CalculateServiceCharge",
				InputMapping: map[string]string{
					"appointmentID": "$.steps.1.result.appointment_id",
					"patientID":     "$.input.patient_id",
					"serviceDate":   "$.input.service_date",
					"serviceData":   "$.steps.5.result.service_data",
					"providerID":    "$.input.provider_id",
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
			// Step 8: Generate Billing Document
			{
				StepNumber:    8,
				ServiceName:   "billing",
				HandlerMethod: "GenerateBillingDocument",
				InputMapping: map[string]string{
					"appointmentID":   "$.steps.1.result.appointment_id",
					"patientID":       "$.input.patient_id",
					"serviceDate":     "$.input.service_date",
					"serviceCharge":   "$.steps.7.result.service_charge",
					"billingDetails":  "$.steps.7.result.billing_details",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Process Insurance Claim
			{
				StepNumber:    9,
				ServiceName:   "insurance",
				HandlerMethod: "ProcessInsuranceClaim",
				InputMapping: map[string]string{
					"appointmentID":     "$.steps.1.result.appointment_id",
					"patientID":         "$.input.patient_id",
					"serviceCharge":     "$.steps.7.result.service_charge",
					"billingDocument":   "$.steps.8.result.billing_document",
				},
				TimeoutSeconds:    35,
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
			// Step 10: Update Patient Account
			{
				StepNumber:    10,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "UpdatePatientAccount",
				InputMapping: map[string]string{
					"appointmentID":   "$.steps.1.result.appointment_id",
					"patientID":       "$.input.patient_id",
					"serviceCharge":   "$.steps.7.result.service_charge",
					"insuranceClaim":  "$.steps.9.result.insurance_claim",
					"billingDocument": "$.steps.8.result.billing_document",
				},
				TimeoutSeconds:    25,
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
			// Step 11: Apply Service Journal
			{
				StepNumber:    11,
				ServiceName:   "general-ledger",
				HandlerMethod: "ApplyServiceJournal",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"appointmentID":   "$.steps.1.result.appointment_id",
					"serviceCharge":   "$.steps.7.result.service_charge",
					"journalDate":     "$.input.service_date",
					"billingDocument": "$.steps.8.result.billing_document",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 12: Link Compliance Record
			{
				StepNumber:    12,
				ServiceName:   "compliance",
				HandlerMethod: "LinkComplianceRecord",
				InputMapping: map[string]string{
					"appointmentID":     "$.steps.1.result.appointment_id",
					"patientID":         "$.input.patient_id",
					"serviceDate":       "$.input.service_date",
					"billingDocument":   "$.steps.8.result.billing_document",
				},
				TimeoutSeconds:    20,
				IsCritical:        false,
				CompensationSteps: []int32{110},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 13: Complete Billing Process
			{
				StepNumber:    13,
				ServiceName:   "billing",
				HandlerMethod: "CompleteBillingProcess",
				InputMapping: map[string]string{
					"appointmentID":     "$.steps.1.result.appointment_id",
					"patientID":         "$.input.patient_id",
					"billingDocument":   "$.steps.8.result.billing_document",
					"insuranceClaim":    "$.steps.9.result.insurance_claim",
					"journalEntries":    "$.steps.11.result.journal_entries",
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

			// Step 101: Cancel Service Schedule (compensates step 3)
			{
				StepNumber:    101,
				ServiceName:   "service-delivery",
				HandlerMethod: "CancelServiceSchedule",
				InputMapping: map[string]string{
					"appointmentID": "$.steps.1.result.appointment_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Revert Service Record Initialization (compensates step 4)
			{
				StepNumber:    102,
				ServiceName:   "medical-records",
				HandlerMethod: "RevertServiceRecordInitialization",
				InputMapping: map[string]string{
					"appointmentID": "$.steps.1.result.appointment_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Cancel Service Execution (compensates step 5)
			{
				StepNumber:    103,
				ServiceName:   "service-delivery",
				HandlerMethod: "CancelServiceExecution",
				InputMapping: map[string]string{
					"appointmentID": "$.steps.1.result.appointment_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 104: Revert Service Details Recording (compensates step 6)
			{
				StepNumber:    104,
				ServiceName:   "medical-records",
				HandlerMethod: "RevertServiceDetailsRecording",
				InputMapping: map[string]string{
					"appointmentID": "$.steps.1.result.appointment_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Clear Service Charge Calculation (compensates step 7)
			{
				StepNumber:    105,
				ServiceName:   "billing",
				HandlerMethod: "ClearServiceChargeCalculation",
				InputMapping: map[string]string{
					"appointmentID": "$.steps.1.result.appointment_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: Cancel Billing Document (compensates step 8)
			{
				StepNumber:    106,
				ServiceName:   "billing",
				HandlerMethod: "CancelBillingDocument",
				InputMapping: map[string]string{
					"appointmentID": "$.steps.1.result.appointment_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Revert Insurance Claim Processing (compensates step 9)
			{
				StepNumber:    107,
				ServiceName:   "insurance",
				HandlerMethod: "RevertInsuranceClaimProcessing",
				InputMapping: map[string]string{
					"appointmentID": "$.steps.1.result.appointment_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 108: Revert Patient Account Update (compensates step 10)
			{
				StepNumber:    108,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "RevertPatientAccountUpdate",
				InputMapping: map[string]string{
					"appointmentID": "$.steps.1.result.appointment_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 109: Reverse Service Journal (compensates step 11)
			{
				StepNumber:    109,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseServiceJournal",
				InputMapping: map[string]string{
					"appointmentID": "$.steps.1.result.appointment_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 110: Unlink Compliance Record (compensates step 12)
			{
				StepNumber:    110,
				ServiceName:   "compliance",
				HandlerMethod: "UnlinkComplianceRecord",
				InputMapping: map[string]string{
					"appointmentID": "$.steps.1.result.appointment_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ServiceDeliverySaga) SagaType() string {
	return "SAGA-HC02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ServiceDeliverySaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ServiceDeliverySaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ServiceDeliverySaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["appointment_id"] == nil {
		return errors.New("appointment_id is required")
	}

	if inputMap["patient_id"] == nil {
		return errors.New("patient_id is required")
	}

	if inputMap["service_date"] == nil {
		return errors.New("service_date is required (format: YYYY-MM-DD)")
	}

	if inputMap["provider_id"] == nil {
		return errors.New("provider_id is required")
	}

	return nil
}
