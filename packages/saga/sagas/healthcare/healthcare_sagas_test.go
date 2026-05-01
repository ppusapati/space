// Package healthcare provides comprehensive unit tests for healthcare saga handlers
package healthcare

import (
	"strings"
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// Helper function to check if error message contains substring
func contains(str, substring string) bool {
	return strings.Contains(str, substring)
}

// ========== PATIENT REGISTRATION SAGA TESTS (SAGA-HC01) ==========

// TestPatientRegistrationSagaType verifies Patient Registration saga returns correct type
func TestPatientRegistrationSagaType(t *testing.T) {
	s := NewPatientRegistrationSaga()
	if s.SagaType() != "SAGA-HC01" {
		t.Errorf("expected SAGA-HC01, got %s", s.SagaType())
	}
}

// TestPatientRegistrationStepCount verifies step count
func TestPatientRegistrationStepCount(t *testing.T) {
	s := NewPatientRegistrationSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 17 // 9 forward + 8 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestPatientRegistrationImplementsInterface verifies saga implements SagaHandler
func TestPatientRegistrationImplementsInterface(t *testing.T) {
	s := NewPatientRegistrationSaga()
	var _ saga.SagaHandler = s
}

// TestPatientRegistrationValidation verifies input validation
func TestPatientRegistrationValidation(t *testing.T) {
	s := NewPatientRegistrationSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid patient registration input",
			map[string]interface{}{
				"patient_id":       "pat-001",
				"first_name":       "John",
				"last_name":        "Doe",
				"date_of_birth":    "1990-01-15",
				"email":            "john@example.com",
				"phone":            "+91-9876543210",
				"gender":           "M",
				"blood_group":      "O+",
			},
			false,
			"",
		},
		{
			"missing patient_id",
			map[string]interface{}{
				"first_name":    "John",
				"last_name":     "Doe",
				"date_of_birth": "1990-01-15",
				"email":         "john@example.com",
			},
			true,
			"patient_id is required",
		},
		{
			"missing first_name",
			map[string]interface{}{
				"patient_id":    "pat-001",
				"last_name":     "Doe",
				"date_of_birth": "1990-01-15",
				"email":         "john@example.com",
			},
			true,
			"first_name is required",
		},
		{
			"invalid date_of_birth",
			map[string]interface{}{
				"patient_id":    "pat-001",
				"first_name":    "John",
				"last_name":     "Doe",
				"date_of_birth": "invalid-date",
				"email":         "john@example.com",
			},
			true,
			"date_of_birth must be valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.hasErr && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

// TestPatientProfileCreationInput verifies profile creation step
func TestPatientProfileCreationInput(t *testing.T) {
	s := NewPatientRegistrationSaga()
	stepDef := s.GetStepDefinition(1)

	if stepDef == nil {
		t.Error("step 1 (Patient Profile Creation) not found")
	}
}

// TestMedicalHistoryInput verifies medical history step
func TestMedicalHistoryInput(t *testing.T) {
	s := NewPatientRegistrationSaga()
	stepDef := s.GetStepDefinition(3)

	if stepDef == nil {
		t.Error("step 3 (Medical History) not found")
	}
}

// TestInsuranceDetailsInput verifies insurance details step
func TestInsuranceDetailsInput(t *testing.T) {
	s := NewPatientRegistrationSaga()
	stepDef := s.GetStepDefinition(5)

	if stepDef == nil {
		t.Error("step 5 (Insurance Details) not found")
	}
}

// TestPatientRegistrationCompensation verifies compensation steps
func TestPatientRegistrationCompensation(t *testing.T) {
	s := NewPatientRegistrationSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount != 8 {
		t.Errorf("expected 8 compensation steps, got %d", compensationCount)
	}
}

// ========== SERVICE DELIVERY SAGA TESTS (SAGA-HC02) ==========

// TestServiceDeliverySagaType verifies Service Delivery saga returns correct type
func TestServiceDeliverySagaType(t *testing.T) {
	s := NewServiceDeliverySaga()
	if s.SagaType() != "SAGA-HC02" {
		t.Errorf("expected SAGA-HC02, got %s", s.SagaType())
	}
}

// TestServiceDeliveryStepCount verifies step count (LARGEST: 23 steps)
func TestServiceDeliveryStepCount(t *testing.T) {
	s := NewServiceDeliverySaga()
	steps := s.GetStepDefinitions()
	expectedCount := 23 // 13 forward + 10 compensation (largest saga)
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestServiceDeliveryImplementsInterface verifies saga implements SagaHandler
func TestServiceDeliveryImplementsInterface(t *testing.T) {
	s := NewServiceDeliverySaga()
	var _ saga.SagaHandler = s
}

// TestServiceDeliveryValidation verifies input validation
func TestServiceDeliveryValidation(t *testing.T) {
	s := NewServiceDeliverySaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid service delivery input",
			map[string]interface{}{
				"service_id":        "svc-001",
				"patient_id":        "pat-001",
				"appointment_date":  "2026-02-20",
				"appointment_time":  "10:00",
				"service_type":      "CONSULTATION",
				"provider_id":       "prov-001",
				"department":        "CARDIOLOGY",
				"estimated_amount":  5000.00,
			},
			false,
			"",
		},
		{
			"missing service_id",
			map[string]interface{}{
				"patient_id":       "pat-001",
				"appointment_date": "2026-02-20",
				"appointment_time": "10:00",
				"service_type":     "CONSULTATION",
			},
			true,
			"service_id is required",
		},
		{
			"invalid appointment_date",
			map[string]interface{}{
				"service_id":       "svc-001",
				"patient_id":       "pat-001",
				"appointment_date": "invalid-date",
				"appointment_time": "10:00",
				"service_type":     "CONSULTATION",
			},
			true,
			"appointment_date must be valid",
		},
		{
			"negative estimated_amount",
			map[string]interface{}{
				"service_id":       "svc-001",
				"patient_id":       "pat-001",
				"appointment_date": "2026-02-20",
				"appointment_time": "10:00",
				"service_type":     "CONSULTATION",
				"estimated_amount": -1000.00,
			},
			true,
			"estimated_amount must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.hasErr && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

// TestAppointmentSchedulingInput verifies appointment scheduling step
func TestAppointmentSchedulingInput(t *testing.T) {
	s := NewServiceDeliverySaga()
	stepDef := s.GetStepDefinition(2)

	if stepDef == nil {
		t.Error("step 2 (Appointment Scheduling) not found")
	}
}

// TestServiceExecutionInput verifies service execution step
func TestServiceExecutionInput(t *testing.T) {
	s := NewServiceDeliverySaga()
	stepDef := s.GetStepDefinition(5)

	if stepDef == nil {
		t.Error("step 5 (Service Execution) not found")
	}
}

// TestDiagnosisInput verifies diagnosis step
func TestDiagnosisInput(t *testing.T) {
	s := NewServiceDeliverySaga()
	stepDef := s.GetStepDefinition(7)

	if stepDef == nil {
		t.Error("step 7 (Diagnosis) not found")
	}
}

// TestTreatmentInput verifies treatment step
func TestTreatmentInput(t *testing.T) {
	s := NewServiceDeliverySaga()
	stepDef := s.GetStepDefinition(8)

	if stepDef == nil {
		t.Error("step 8 (Treatment) not found")
	}
}

// TestServiceDeliveryBillingInput verifies billing step
func TestServiceDeliveryBillingInput(t *testing.T) {
	s := NewServiceDeliverySaga()
	stepDef := s.GetStepDefinition(10)

	if stepDef == nil {
		t.Error("step 10 (Billing) not found")
	}
}

// TestServiceDeliveryCompensation verifies compensation steps
func TestServiceDeliveryCompensation(t *testing.T) {
	s := NewServiceDeliverySaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount != 10 {
		t.Errorf("expected 10 compensation steps, got %d", compensationCount)
	}
}

// ========== CLAIMS PROCESSING SAGA TESTS (SAGA-HC03) ==========

// TestClaimsProcessingSagaType verifies Claims Processing saga returns correct type
func TestClaimsProcessingSagaType(t *testing.T) {
	s := NewClaimsProcessingSaga()
	if s.SagaType() != "SAGA-HC03" {
		t.Errorf("expected SAGA-HC03, got %s", s.SagaType())
	}
}

// TestClaimsProcessingStepCount verifies step count
func TestClaimsProcessingStepCount(t *testing.T) {
	s := NewClaimsProcessingSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 19 // 11 forward + 8 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestClaimsProcessingImplementsInterface verifies saga implements SagaHandler
func TestClaimsProcessingImplementsInterface(t *testing.T) {
	s := NewClaimsProcessingSaga()
	var _ saga.SagaHandler = s
}

// TestClaimsProcessingValidation verifies input validation
func TestClaimsProcessingValidation(t *testing.T) {
	s := NewClaimsProcessingSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid claims processing input",
			map[string]interface{}{
				"claim_id":          "claim-001",
				"patient_id":        "pat-001",
				"service_id":        "svc-001",
				"claim_amount":      5000.00,
				"insurance_id":      "ins-001",
				"submission_date":   "2026-02-16",
				"service_date":      "2026-02-10",
			},
			false,
			"",
		},
		{
			"missing claim_id",
			map[string]interface{}{
				"patient_id":      "pat-001",
				"service_id":      "svc-001",
				"claim_amount":    5000.00,
				"insurance_id":    "ins-001",
				"submission_date": "2026-02-16",
			},
			true,
			"claim_id is required",
		},
		{
			"invalid claim_amount",
			map[string]interface{}{
				"claim_id":        "claim-001",
				"patient_id":      "pat-001",
				"service_id":      "svc-001",
				"claim_amount":    -5000.00,
				"insurance_id":    "ins-001",
				"submission_date": "2026-02-16",
			},
			true,
			"claim_amount must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.hasErr && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

// TestClaimSubmissionInput verifies claim submission step
func TestClaimSubmissionInput(t *testing.T) {
	s := NewClaimsProcessingSaga()
	stepDef := s.GetStepDefinition(2)

	if stepDef == nil {
		t.Error("step 2 (Claim Submission) not found")
	}
}

// TestEligibilityVerificationInput verifies eligibility verification step
func TestEligibilityVerificationInput(t *testing.T) {
	s := NewClaimsProcessingSaga()
	stepDef := s.GetStepDefinition(3)

	if stepDef == nil {
		t.Error("step 3 (Eligibility Verification) not found")
	}
}

// TestClaimProcessingInput verifies claim processing step
func TestClaimProcessingInput(t *testing.T) {
	s := NewClaimsProcessingSaga()
	stepDef := s.GetStepDefinition(5)

	if stepDef == nil {
		t.Error("step 5 (Claim Processing) not found")
	}
}

// TestClaimPaymentInput verifies payment step
func TestClaimPaymentInput(t *testing.T) {
	s := NewClaimsProcessingSaga()
	stepDef := s.GetStepDefinition(8)

	if stepDef == nil {
		t.Error("step 8 (Payment) not found")
	}
}

// TestClaimsProcessingCompensation verifies compensation steps
func TestClaimsProcessingCompensation(t *testing.T) {
	s := NewClaimsProcessingSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount != 8 {
		t.Errorf("expected 8 compensation steps, got %d", compensationCount)
	}
}

// ========== MEDICAL SUPPLY SAGA TESTS (SAGA-HC04) ==========

// TestMedicalSupplySagaType verifies Medical Supply saga returns correct type
func TestMedicalSupplySagaType(t *testing.T) {
	s := NewMedicalSupplySaga()
	if s.SagaType() != "SAGA-HC04" {
		t.Errorf("expected SAGA-HC04, got %s", s.SagaType())
	}
}

// TestMedicalSupplyStepCount verifies step count
func TestMedicalSupplyStepCount(t *testing.T) {
	s := NewMedicalSupplySaga()
	steps := s.GetStepDefinitions()
	expectedCount := 19 // 11 forward + 8 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestMedicalSupplyImplementsInterface verifies saga implements SagaHandler
func TestMedicalSupplyImplementsInterface(t *testing.T) {
	s := NewMedicalSupplySaga()
	var _ saga.SagaHandler = s
}

// TestMedicalSupplyValidation verifies input validation
func TestMedicalSupplyValidation(t *testing.T) {
	s := NewMedicalSupplySaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid medical supply input",
			map[string]interface{}{
				"requisition_id": "req-001",
				"department":     "EMERGENCY",
				"supplies":       []interface{}{map[string]interface{}{"item": "Bandage", "qty": 100}},
				"urgency":        "HIGH",
				"required_date":  "2026-02-17",
			},
			false,
			"",
		},
		{
			"missing requisition_id",
			map[string]interface{}{
				"department":    "EMERGENCY",
				"supplies":      []interface{}{map[string]interface{}{"item": "Bandage", "qty": 100}},
				"urgency":       "HIGH",
				"required_date": "2026-02-17",
			},
			true,
			"requisition_id is required",
		},
		{
			"empty supplies",
			map[string]interface{}{
				"requisition_id": "req-001",
				"department":     "EMERGENCY",
				"supplies":       []interface{}{},
				"urgency":        "HIGH",
				"required_date":  "2026-02-17",
			},
			true,
			"supplies cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.hasErr && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

// TestSupplyRequisitionInput verifies supply requisition step
func TestSupplyRequisitionInput(t *testing.T) {
	s := NewMedicalSupplySaga()
	stepDef := s.GetStepDefinition(2)

	if stepDef == nil {
		t.Error("step 2 (Supply Requisition) not found")
	}
}

// TestSupplyProcurementInput verifies procurement step
func TestSupplyProcurementInput(t *testing.T) {
	s := NewMedicalSupplySaga()
	stepDef := s.GetStepDefinition(4)

	if stepDef == nil {
		t.Error("step 4 (Procurement) not found")
	}
}

// TestSupplyInventoryUpdateInput verifies inventory update step
func TestSupplyInventoryUpdateInput(t *testing.T) {
	s := NewMedicalSupplySaga()
	stepDef := s.GetStepDefinition(6)

	if stepDef == nil {
		t.Error("step 6 (Inventory Update) not found")
	}
}

// TestMedicalSupplyCompensation verifies compensation steps
func TestMedicalSupplyCompensation(t *testing.T) {
	s := NewMedicalSupplySaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount != 8 {
		t.Errorf("expected 8 compensation steps, got %d", compensationCount)
	}
}

// ========== COMPLIANCE SAGA TESTS (SAGA-HC05) ==========

// TestComplianceSagaType verifies Compliance saga returns correct type
func TestComplianceSagaType(t *testing.T) {
	s := NewComplianceSaga()
	if s.SagaType() != "SAGA-HC05" {
		t.Errorf("expected SAGA-HC05, got %s", s.SagaType())
	}
}

// TestComplianceStepCount verifies step count
func TestComplianceStepCount(t *testing.T) {
	s := NewComplianceSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 15 // 8 forward + 7 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestComplianceImplementsInterface verifies saga implements SagaHandler
func TestComplianceImplementsInterface(t *testing.T) {
	s := NewComplianceSaga()
	var _ saga.SagaHandler = s
}

// TestComplianceValidation verifies input validation
func TestComplianceValidation(t *testing.T) {
	s := NewComplianceSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid compliance input",
			map[string]interface{}{
				"audit_id":      "audit-001",
				"audit_type":    "HIPAA",
				"audit_date":    "2026-02-16",
				"department":    "RECORDS",
				"scope":         "Patient Data Protection",
			},
			false,
			"",
		},
		{
			"missing audit_id",
			map[string]interface{}{
				"audit_type": "HIPAA",
				"audit_date": "2026-02-16",
				"department": "RECORDS",
			},
			true,
			"audit_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestAuditInitiationInput verifies audit initiation step
func TestAuditInitiationInput(t *testing.T) {
	s := NewComplianceSaga()
	stepDef := s.GetStepDefinition(1)

	if stepDef == nil {
		t.Error("step 1 (Audit Initiation) not found")
	}
}

// TestRegulatoryCheckInput verifies regulatory check step
func TestRegulatoryCheckInput(t *testing.T) {
	s := NewComplianceSaga()
	stepDef := s.GetStepDefinition(3)

	if stepDef == nil {
		t.Error("step 3 (Regulatory Check) not found")
	}
}

// TestComplianceCompensation verifies compensation steps
func TestComplianceCompensation(t *testing.T) {
	s := NewComplianceSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount != 7 {
		t.Errorf("expected 7 compensation steps, got %d", compensationCount)
	}
}

// ========== PROVIDER NETWORK SAGA TESTS (SAGA-HC06) ==========

// TestProviderNetworkSagaType verifies Provider Network saga returns correct type
func TestProviderNetworkSagaType(t *testing.T) {
	s := NewProviderNetworkSaga()
	if s.SagaType() != "SAGA-HC06" {
		t.Errorf("expected SAGA-HC06, got %s", s.SagaType())
	}
}

// TestProviderNetworkStepCount verifies step count
func TestProviderNetworkStepCount(t *testing.T) {
	s := NewProviderNetworkSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 13 // 7 forward + 6 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestProviderNetworkImplementsInterface verifies saga implements SagaHandler
func TestProviderNetworkImplementsInterface(t *testing.T) {
	s := NewProviderNetworkSaga()
	var _ saga.SagaHandler = s
}

// TestProviderNetworkValidation verifies input validation
func TestProviderNetworkValidation(t *testing.T) {
	s := NewProviderNetworkSaga()

	tests := []struct {
		name    string
		input   interface{}
		hasErr  bool
		errMsg  string
	}{
		{
			"valid provider network input",
			map[string]interface{}{
				"provider_id":     "prov-001",
				"provider_name":   "Dr. Smith",
				"specialization":  "CARDIOLOGY",
				"registration_no": "REG123456",
				"onboard_date":    "2026-02-16",
			},
			false,
			"",
		},
		{
			"missing provider_id",
			map[string]interface{}{
				"provider_name":   "Dr. Smith",
				"specialization":  "CARDIOLOGY",
				"registration_no": "REG123456",
				"onboard_date":    "2026-02-16",
			},
			true,
			"provider_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if tt.hasErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.hasErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestProviderOnboardingInput verifies provider onboarding step
func TestProviderOnboardingInput(t *testing.T) {
	s := NewProviderNetworkSaga()
	stepDef := s.GetStepDefinition(2)

	if stepDef == nil {
		t.Error("step 2 (Provider Onboarding) not found")
	}
}

// TestNetworkConfigInput verifies network configuration step
func TestNetworkConfigInput(t *testing.T) {
	s := NewProviderNetworkSaga()
	stepDef := s.GetStepDefinition(4)

	if stepDef == nil {
		t.Error("step 4 (Network Configuration) not found")
	}
}

// TestProviderNetworkCompensation verifies compensation steps
func TestProviderNetworkCompensation(t *testing.T) {
	s := NewProviderNetworkSaga()
	steps := s.GetStepDefinitions()

	compensationCount := 0
	for _, step := range steps {
		if step.StepNumber > 100 {
			compensationCount++
		}
	}

	if compensationCount != 6 {
		t.Errorf("expected 6 compensation steps, got %d", compensationCount)
	}
}

// ========== INTEGRATION TESTS ==========

// TestHealthcareSagasInterface verifies all healthcare sagas implement SagaHandler
func TestHealthcareSagasInterface(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewPatientRegistrationSaga(),
		NewServiceDeliverySaga(),
		NewClaimsProcessingSaga(),
		NewMedicalSupplySaga(),
		NewComplianceSaga(),
		NewProviderNetworkSaga(),
	}

	for _, s := range sagas {
		if s == nil {
			t.Error("saga is nil")
		}
		if s.SagaType() == "" {
			t.Error("saga type is empty")
		}
		if len(s.GetStepDefinitions()) == 0 {
			t.Errorf("saga %s has no steps", s.SagaType())
		}
	}
}

// TestHealthcareSagaTypes verifies all saga types return correct identifiers
func TestHealthcareSagaTypes(t *testing.T) {
	sagas := []struct {
		name         string
		saga         saga.SagaHandler
		expectedType string
	}{
		{"Patient Registration", NewPatientRegistrationSaga(), "SAGA-HC01"},
		{"Service Delivery", NewServiceDeliverySaga(), "SAGA-HC02"},
		{"Claims Processing", NewClaimsProcessingSaga(), "SAGA-HC03"},
		{"Medical Supply", NewMedicalSupplySaga(), "SAGA-HC04"},
		{"Compliance", NewComplianceSaga(), "SAGA-HC05"},
		{"Provider Network", NewProviderNetworkSaga(), "SAGA-HC06"},
	}

	for _, tt := range sagas {
		t.Run(tt.name, func(t *testing.T) {
			if tt.saga.SagaType() != tt.expectedType {
				t.Errorf("expected %s, got %s", tt.expectedType, tt.saga.SagaType())
			}
		})
	}
}

// TestHealthcareSagasGetStepDefinitions verifies step retrieval
func TestHealthcareSagasGetStepDefinitions(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewPatientRegistrationSaga(),
		NewServiceDeliverySaga(),
		NewClaimsProcessingSaga(),
		NewMedicalSupplySaga(),
		NewComplianceSaga(),
		NewProviderNetworkSaga(),
	}

	for _, s := range sagas {
		steps := s.GetStepDefinitions()
		if len(steps) == 0 {
			t.Errorf("saga %s has no steps", s.SagaType())
		}

		// Verify first step exists
		firstStep := s.GetStepDefinition(1)
		if firstStep == nil {
			t.Errorf("saga %s: step 1 not found", s.SagaType())
		}
	}
}

// TestHealthcareSagasInvalidStepLookup verifies invalid step lookup returns nil
func TestHealthcareSagasInvalidStepLookup(t *testing.T) {
	s := NewPatientRegistrationSaga()
	invalidStep := s.GetStepDefinition(999)
	if invalidStep != nil {
		t.Error("invalid step should return nil")
	}
}

// TestHealthcareSagasNilInput verifies nil input handling
func TestHealthcareSagasNilInput(t *testing.T) {
	s := NewPatientRegistrationSaga()
	err := s.ValidateInput(nil)
	if err == nil {
		t.Error("expected error for nil input")
	}
}

// TestHealthcareSagasEmptyMapInput verifies empty map input handling
func TestHealthcareSagasEmptyMapInput(t *testing.T) {
	s := NewServiceDeliverySaga()
	err := s.ValidateInput(map[string]interface{}{})
	if err == nil {
		t.Error("expected error for empty map input")
	}
}

// TestHealthcareSagasStringInput verifies string input rejection
func TestHealthcareSagasStringInput(t *testing.T) {
	s := NewClaimsProcessingSaga()
	err := s.ValidateInput("invalid string")
	if err == nil {
		t.Error("expected error for string input")
	}
}

// TestHealthcareSagasIntInput verifies integer input rejection
func TestHealthcareSagasIntInput(t *testing.T) {
	s := NewMedicalSupplySaga()
	err := s.ValidateInput(12345)
	if err == nil {
		t.Error("expected error for integer input")
	}
}

// TestHealthcareSagasCriticalStepMarking verifies critical steps are marked
func TestHealthcareSagasCriticalStepMarking(t *testing.T) {
	s := NewPatientRegistrationSaga()
	steps := s.GetStepDefinitions()

	hasCritical := false
	for _, step := range steps {
		if step.IsCritical {
			hasCritical = true
			break
		}
	}

	if !hasCritical {
		t.Error("no critical steps found in patient registration saga")
	}
}

// TestHealthcareSagasRetryConfiguration verifies retry config exists
func TestHealthcareSagasRetryConfiguration(t *testing.T) {
	s := NewServiceDeliverySaga()
	stepDef := s.GetStepDefinition(1)

	if stepDef.RetryConfig == nil {
		t.Error("retry configuration missing in step 1")
	}
}

// TestHealthcareSagasTimeoutConfiguration verifies timeout config exists
func TestHealthcareSagasTimeoutConfiguration(t *testing.T) {
	s := NewClaimsProcessingSaga()
	stepDef := s.GetStepDefinition(1)

	if stepDef.TimeoutSeconds == 0 {
		t.Error("timeout configuration missing in step 1")
	}
}

// TestHealthcareSagasLargestSagaServiceDelivery verifies Service Delivery is largest
func TestHealthcareSagasLargestSagaServiceDelivery(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewPatientRegistrationSaga(),
		NewServiceDeliverySaga(),
		NewClaimsProcessingSaga(),
		NewMedicalSupplySaga(),
		NewComplianceSaga(),
		NewProviderNetworkSaga(),
	}

	maxSteps := 0
	largestSagaType := ""
	for _, s := range sagas {
		steps := len(s.GetStepDefinitions())
		if steps > maxSteps {
			maxSteps = steps
			largestSagaType = s.SagaType()
		}
	}

	if largestSagaType != "SAGA-HC02" {
		t.Errorf("expected SAGA-HC02 (Service Delivery) to be largest, but %s has %d steps", largestSagaType, maxSteps)
	}

	if maxSteps != 23 {
		t.Errorf("expected Service Delivery saga to have 23 steps, got %d", maxSteps)
	}
}

// TestHealthcareSagasCompensationStepChaining verifies compensation chains exist
func TestHealthcareSagasCompensationStepChaining(t *testing.T) {
	s := NewServiceDeliverySaga()
	steps := s.GetStepDefinitions()

	// Verify forward steps have compensation mappings
	forwardStepsWithCompensation := 0
	for _, step := range steps {
		if step.StepNumber <= 100 && len(step.CompensationSteps) > 0 {
			forwardStepsWithCompensation++
		}
	}

	if forwardStepsWithCompensation == 0 {
		t.Error("no forward steps have compensation steps mapped")
	}
}
