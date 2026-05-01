// Package hr provides saga handlers for HR & Payroll module workflows
package hr

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// Form16GenerationSaga implements SAGA-SR01: Form 16 Generation (Annual Tax Certificate) workflow
// Business Flow: Extract employee salary records → Calculate total gross salary →
// Sum all TDS deductions → Calculate deductions u/s 80C → Calculate deductions u/s 80D →
// Calculate net taxable income → Generate Form 16 Part A → Generate Form 16 Part B →
// Issue Form 16 to employee (9 forward steps + 8 compensation = 17 total)
// Critical steps: 3, 6, 9 (3 critical)
// Non-critical steps: 1, 2, 4, 5, 7, 8 (6 non-critical)
// Timeout: 240s aggregate (mix of 20-45s per step)
// Statutory compliance: Income Tax Act Section 203 (Form 16 issuance by June 30)
type Form16GenerationSaga struct {
	steps []*saga.StepDefinition
}

// NewForm16GenerationSaga creates a new Form 16 Generation saga handler
func NewForm16GenerationSaga() saga.SagaHandler {
	return &Form16GenerationSaga{
		steps: []*saga.StepDefinition{
			// ===== FORWARD STEPS (1-9) =====

			// Step 1: Extract Employee Salary Records - NON-CRITICAL
			{
				StepNumber:    1,
				ServiceName:   "payroll",
				HandlerMethod: "ExtractEmployeeSalaryRecords",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"form16GenerationID": "$.input.form16_generation_id",
					"assessmentYear":     "$.input.assessment_year",
					"employeeID":         "$.input.employee_id",
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
			// Step 2: Calculate Total Gross Salary - NON-CRITICAL
			{
				StepNumber:    2,
				ServiceName:   "salary-structure",
				HandlerMethod: "CalculateTotalGrossSalary",
				InputMapping: map[string]string{
					"tenantID":               "$.tenantID",
					"companyID":              "$.companyID",
					"branchID":               "$.branchID",
					"form16GenerationID":     "$.input.form16_generation_id",
					"assessmentYear":         "$.input.assessment_year",
					"salaryRecordsExtracted": "$.steps.1.result.salary_records_extracted",
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
			// Step 3: Sum All TDS Deductions - CRITICAL
			{
				StepNumber:    3,
				ServiceName:   "tds",
				HandlerMethod: "SumTDSDeductions",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"form16GenerationID":    "$.input.form16_generation_id",
					"assessmentYear":        "$.input.assessment_year",
					"grossSalaryCalculated": "$.steps.2.result.gross_salary_calculated",
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
			// Step 4: Calculate Deductions u/s 80C - NON-CRITICAL
			{
				StepNumber:    4,
				ServiceName:   "payroll",
				HandlerMethod: "CalculateDeductionsSection80C",
				InputMapping: map[string]string{
					"tenantID":                "$.tenantID",
					"companyID":               "$.companyID",
					"branchID":                "$.branchID",
					"form16GenerationID":      "$.input.form16_generation_id",
					"assessmentYear":          "$.input.assessment_year",
					"tdsDeductionsCalculated": "$.steps.3.result.tds_deductions_calculated",
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
			// Step 5: Calculate Deductions u/s 80D - NON-CRITICAL
			{
				StepNumber:    5,
				ServiceName:   "payroll",
				HandlerMethod: "CalculateDeductionsSection80D",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"form16GenerationID":   "$.input.form16_generation_id",
					"assessmentYear":       "$.input.assessment_year",
					"section80CCalculated": "$.steps.4.result.section_80c_calculated",
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
				CompensationSteps: []int32{104},
			},
			// Step 6: Calculate Net Taxable Income - CRITICAL
			{
				StepNumber:    6,
				ServiceName:   "tds",
				HandlerMethod: "CalculateNetTaxableIncome",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"form16GenerationID":   "$.input.form16_generation_id",
					"assessmentYear":       "$.input.assessment_year",
					"section80DCalculated": "$.steps.5.result.section_80d_calculated",
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
				CompensationSteps: []int32{105},
			},
			// Step 7: Generate Form 16 Part A - NON-CRITICAL
			{
				StepNumber:    7,
				ServiceName:   "payroll",
				HandlerMethod: "GenerateForm16PartA",
				InputMapping: map[string]string{
					"tenantID":                   "$.tenantID",
					"companyID":                  "$.companyID",
					"branchID":                   "$.branchID",
					"form16GenerationID":         "$.input.form16_generation_id",
					"assessmentYear":             "$.input.assessment_year",
					"netTaxableIncomeCalculated": "$.steps.6.result.net_taxable_income_calculated",
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
			// Step 8: Generate Form 16 Part B - NON-CRITICAL
			{
				StepNumber:    8,
				ServiceName:   "tds",
				HandlerMethod: "GenerateForm16PartB",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"form16GenerationID":   "$.input.form16_generation_id",
					"assessmentYear":       "$.input.assessment_year",
					"form16PartAGenerated": "$.steps.7.result.form16_part_a_generated",
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
				CompensationSteps: []int32{107},
			},
			// Step 9: Issue Form 16 to Employee (Notification) - CRITICAL
			{
				StepNumber:    9,
				ServiceName:   "notification",
				HandlerMethod: "IssueForm16ToEmployee",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"form16GenerationID":   "$.input.form16_generation_id",
					"assessmentYear":       "$.input.assessment_year",
					"form16PartBGenerated": "$.steps.8.result.form16_part_b_generated",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{108},
			},

			// ===== COMPENSATION STEPS (101-108) =====

			// Step 101: Revert Calculate Total Gross Salary (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "salary-structure",
				HandlerMethod: "UndoCalculateTotalGrossSalary",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"form16GenerationID": "$.input.form16_generation_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Revert Sum TDS Deductions (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "tds",
				HandlerMethod: "UndoSumTDSDeductions",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"form16GenerationID": "$.input.form16_generation_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Revert Calculate Deductions u/s 80C (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "payroll",
				HandlerMethod: "UndoCalculateDeductionsSection80C",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"form16GenerationID": "$.input.form16_generation_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Revert Calculate Deductions u/s 80D (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "payroll",
				HandlerMethod: "UndoCalculateDeductionsSection80D",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"form16GenerationID": "$.input.form16_generation_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: Revert Calculate Net Taxable Income (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "tds",
				HandlerMethod: "UndoCalculateNetTaxableIncome",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"form16GenerationID": "$.input.form16_generation_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: Revert Generate Form 16 Part A (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "payroll",
				HandlerMethod: "UndoGenerateForm16PartA",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"form16GenerationID": "$.input.form16_generation_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Revert Generate Form 16 Part B (compensates step 8)
			{
				StepNumber:    107,
				ServiceName:   "tds",
				HandlerMethod: "UndoGenerateForm16PartB",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"form16GenerationID": "$.input.form16_generation_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: Revert Issue Form 16 to Employee (compensates step 9)
			{
				StepNumber:    108,
				ServiceName:   "notification",
				HandlerMethod: "UndoIssueForm16ToEmployee",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"form16GenerationID": "$.input.form16_generation_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *Form16GenerationSaga) SagaType() string {
	return "SAGA-SR01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *Form16GenerationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *Form16GenerationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: form16_generation_id, assessment_year, company_id, employee_id
func (s *Form16GenerationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	// Extract the nested 'input' object
	innerInput, ok := inputMap["input"].(map[string]interface{})
	if !ok {
		return errors.New("missing 'input' field in saga input")
	}

	// Validate form16_generation_id
	if innerInput["form16_generation_id"] == nil {
		return errors.New("missing required field: form16_generation_id")
	}
	form16GenerationID, ok := innerInput["form16_generation_id"].(string)
	if !ok || form16GenerationID == "" {
		return errors.New("form16_generation_id must be a non-empty string")
	}

	// Validate assessment_year
	if innerInput["assessment_year"] == nil {
		return errors.New("missing required field: assessment_year")
	}
	assessmentYear, ok := innerInput["assessment_year"].(string)
	if !ok || assessmentYear == "" {
		return errors.New("assessment_year must be a non-empty string")
	}

	// Validate employee_id
	if innerInput["employee_id"] == nil {
		return errors.New("missing required field: employee_id")
	}
	employeeID, ok := innerInput["employee_id"].(string)
	if !ok || employeeID == "" {
		return errors.New("employee_id must be a non-empty string")
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
