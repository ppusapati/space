// Package hr provides saga handlers for HR & Payroll module workflows
package hr

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// PFESIRemittanceSaga implements SAGA-SR02: PF/ESI Monthly Remittance workflow
// Business Flow: Extract employee attendance records → Calculate PF contribution →
// Check ESI applicability → Calculate ESI contribution → Deduct PF from salary →
// Calculate employer liability → Create ECR file → Post GL entries →
// Generate payment challan → Mark as submitted (10 forward steps + 10 compensation = 20 total)
// Critical steps: 3, 5, 8 (3 critical)
// Non-critical steps: 1, 2, 4, 6, 7, 9, 10 (7 non-critical)
// Timeout: 300s aggregate (mix of 20-45s per step)
// Statutory compliance: PF Act, ESI Act (remittance by 15th of next month)
type PFESIRemittanceSaga struct {
	steps []*saga.StepDefinition
}

// NewPFESIRemittanceSaga creates a new PF/ESI Remittance saga handler
func NewPFESIRemittanceSaga() saga.SagaHandler {
	return &PFESIRemittanceSaga{
		steps: []*saga.StepDefinition{
			// ===== FORWARD STEPS (1-10) =====

			// Step 1: Extract Employee Attendance Records - NON-CRITICAL
			{
				StepNumber:    1,
				ServiceName:   "attendance",
				HandlerMethod: "ExtractEmployeeAttendanceRecords",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"pfESIRemittanceID": "$.input.pf_esi_remittance_id",
					"remittanceMonth":   "$.input.remittance_month",
					"remittanceYear":    "$.input.remittance_year",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},
			// Step 2: Calculate PF Contribution - NON-CRITICAL
			{
				StepNumber:    2,
				ServiceName:   "payroll",
				HandlerMethod: "CalculatePFContribution",
				InputMapping: map[string]string{
					"tenantID":                   "$.tenantID",
					"companyID":                  "$.companyID",
					"branchID":                   "$.branchID",
					"pfESIRemittanceID":          "$.input.pf_esi_remittance_id",
					"remittanceMonth":            "$.input.remittance_month",
					"attendanceRecordsExtracted": "$.steps.1.result.attendance_records_extracted",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{101},
			},
			// Step 3: Check ESI Applicability - CRITICAL
			{
				StepNumber:    3,
				ServiceName:   "payroll",
				HandlerMethod: "CheckESIApplicability",
				InputMapping: map[string]string{
					"tenantID":                 "$.tenantID",
					"companyID":                "$.companyID",
					"branchID":                 "$.branchID",
					"pfESIRemittanceID":        "$.input.pf_esi_remittance_id",
					"remittanceMonth":          "$.input.remittance_month",
					"pfContributionCalculated": "$.steps.2.result.pf_contribution_calculated",
				},
				TimeoutSeconds: 40,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{102},
			},
			// Step 4: Calculate ESI Contribution - NON-CRITICAL
			{
				StepNumber:    4,
				ServiceName:   "payroll",
				HandlerMethod: "CalculateESIContribution",
				InputMapping: map[string]string{
					"tenantID":                "$.tenantID",
					"companyID":               "$.companyID",
					"branchID":                "$.branchID",
					"pfESIRemittanceID":       "$.input.pf_esi_remittance_id",
					"remittanceMonth":         "$.input.remittance_month",
					"esiApplicabilityChecked": "$.steps.3.result.esi_applicability_checked",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{103},
			},
			// Step 5: Deduct PF from Employee Salary - CRITICAL
			{
				StepNumber:    5,
				ServiceName:   "payroll",
				HandlerMethod: "DeductPFFromSalary",
				InputMapping: map[string]string{
					"tenantID":                  "$.tenantID",
					"companyID":                 "$.companyID",
					"branchID":                  "$.branchID",
					"pfESIRemittanceID":         "$.input.pf_esi_remittance_id",
					"remittanceMonth":           "$.input.remittance_month",
					"esiContributionCalculated": "$.steps.4.result.esi_contribution_calculated",
				},
				TimeoutSeconds: 40,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{104},
			},
			// Step 6: Calculate Employer Liability - NON-CRITICAL
			{
				StepNumber:    6,
				ServiceName:   "payroll",
				HandlerMethod: "CalculateEmployerLiability",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"pfESIRemittanceID":    "$.input.pf_esi_remittance_id",
					"remittanceMonth":      "$.input.remittance_month",
					"pfDeductionProcessed": "$.steps.5.result.pf_deduction_processed",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{105},
			},
			// Step 7: Create ECR (Electronic Challan Reconciliation) File - NON-CRITICAL
			{
				StepNumber:    7,
				ServiceName:   "payroll",
				HandlerMethod: "CreateECRFile",
				InputMapping: map[string]string{
					"tenantID":                    "$.tenantID",
					"companyID":                   "$.companyID",
					"branchID":                    "$.branchID",
					"pfESIRemittanceID":           "$.input.pf_esi_remittance_id",
					"remittanceMonth":             "$.input.remittance_month",
					"employerLiabilityCalculated": "$.steps.6.result.employer_liability_calculated",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{106},
			},
			// Step 8: Post GL - PF/ESI Liability DR, Bank/Payable CR - CRITICAL
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostPFESILiabilityEntries",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"pfESIRemittanceID": "$.input.pf_esi_remittance_id",
					"remittanceMonth":   "$.input.remittance_month",
					"ecrFileCreated":    "$.steps.7.result.ecr_file_created",
				},
				TimeoutSeconds: 40,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{107},
			},
			// Step 9: Generate Payment Challan (NEFT/RTGS) - NON-CRITICAL
			{
				StepNumber:    9,
				ServiceName:   "banking",
				HandlerMethod: "GeneratePaymentChallan",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"pfESIRemittanceID": "$.input.pf_esi_remittance_id",
					"remittanceMonth":   "$.input.remittance_month",
					"glEntriesPosted":   "$.steps.8.result.gl_entries_posted",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{108},
			},
			// Step 10: Mark as Submitted to Authority - NON-CRITICAL
			{
				StepNumber:    10,
				ServiceName:   "payroll",
				HandlerMethod: "MarkRemittanceSubmitted",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"pfESIRemittanceID": "$.input.pf_esi_remittance_id",
					"remittanceMonth":   "$.input.remittance_month",
					"challanGenerated":  "$.steps.9.result.challan_generated",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{109},
			},

			// ===== COMPENSATION STEPS (101-110) =====

			// Step 101: Undo Calculate PF Contribution (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "payroll",
				HandlerMethod: "UndoCalculatePFContribution",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"pfESIRemittanceID": "$.input.pf_esi_remittance_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Undo Check ESI Applicability (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "payroll",
				HandlerMethod: "UndoCheckESIApplicability",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"pfESIRemittanceID": "$.input.pf_esi_remittance_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Undo Calculate ESI Contribution (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "payroll",
				HandlerMethod: "UndoCalculateESIContribution",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"pfESIRemittanceID": "$.input.pf_esi_remittance_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Reverse PF Deduction from Salary (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "payroll",
				HandlerMethod: "UndoDeductPFFromSalary",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"pfESIRemittanceID": "$.input.pf_esi_remittance_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: Undo Calculate Employer Liability (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "payroll",
				HandlerMethod: "UndoCalculateEmployerLiability",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"pfESIRemittanceID": "$.input.pf_esi_remittance_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: Undo Create ECR File (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "payroll",
				HandlerMethod: "UndoCreateECRFile",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"pfESIRemittanceID": "$.input.pf_esi_remittance_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Reverse GL Posting (compensates step 8)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "UndoPostPFESILiabilityEntries",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"pfESIRemittanceID": "$.input.pf_esi_remittance_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: Undo Generate Payment Challan (compensates step 9)
			{
				StepNumber:    108,
				ServiceName:   "banking",
				HandlerMethod: "UndoGeneratePaymentChallan",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"pfESIRemittanceID": "$.input.pf_esi_remittance_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 109: Undo Mark as Submitted (compensates step 10)
			{
				StepNumber:    109,
				ServiceName:   "payroll",
				HandlerMethod: "UndoMarkRemittanceSubmitted",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"pfESIRemittanceID": "$.input.pf_esi_remittance_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *PFESIRemittanceSaga) SagaType() string {
	return "SAGA-SR02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *PFESIRemittanceSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *PFESIRemittanceSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: pf_esi_remittance_id, remittance_month, remittance_year, company_id
func (s *PFESIRemittanceSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	// Extract the nested 'input' object
	innerInput, ok := inputMap["input"].(map[string]interface{})
	if !ok {
		return errors.New("missing 'input' field in saga input")
	}

	// Validate pf_esi_remittance_id
	if innerInput["pf_esi_remittance_id"] == nil {
		return errors.New("missing required field: pf_esi_remittance_id")
	}
	pfESIRemittanceID, ok := innerInput["pf_esi_remittance_id"].(string)
	if !ok || pfESIRemittanceID == "" {
		return errors.New("pf_esi_remittance_id must be a non-empty string")
	}

	// Validate remittance_month
	if innerInput["remittance_month"] == nil {
		return errors.New("missing required field: remittance_month")
	}
	remittanceMonth, ok := innerInput["remittance_month"].(string)
	if !ok || remittanceMonth == "" {
		return errors.New("remittance_month must be a non-empty string")
	}

	// Validate remittance_year
	if innerInput["remittance_year"] == nil {
		return errors.New("missing required field: remittance_year")
	}
	remittanceYear, ok := innerInput["remittance_year"].(string)
	if !ok || remittanceYear == "" {
		return errors.New("remittance_year must be a non-empty string")
	}

	// Validate company_id
	if innerInput["company_id"] == nil {
		return errors.New("missing required field: company_id")
	}
	companyID, ok := innerInput["company_id"].(string)
	if !ok || companyID == "" {
		return errors.New("company_id must be a non-empty string")
	}

	return nil
}
