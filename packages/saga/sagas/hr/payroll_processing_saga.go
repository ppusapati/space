// Package hr provides saga handlers for HR & Payroll module workflows
package hr

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// PayrollProcessingSaga implements SAGA-H01: Payroll Processing workflow
// Business Flow: Lock payroll period → Fetch employee data → Calculate base salary →
// Calculate attendance → Calculate deductions → Calculate tax and contributions → Apply adjustments →
// Calculate net payable → Preview GL entries → Authorize payments → Accrue payables → Post GL entries →
// Finalize payroll run (13 forward steps + 12 compensation = 25 total)
// Critical steps: 1,2,3,4,5,6,10,12,13 (9 critical)
// Non-critical steps: 7,8,9,11 (4 non-critical)
// Timeout: 180s aggregate (mix of 30-45s per step)
type PayrollProcessingSaga struct {
	steps []*saga.StepDefinition
}

// NewPayrollProcessingSaga creates a new Payroll Processing saga handler
func NewPayrollProcessingSaga() saga.SagaHandler {
	return &PayrollProcessingSaga{
		steps: []*saga.StepDefinition{
			// ===== FORWARD STEPS (1-13) =====

			// Step 1: Lock Payroll Period - CRITICAL
			{
				StepNumber:    1,
				ServiceName:   "payroll",
				HandlerMethod: "LockPayrollPeriod",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"payrollRunID":  "$.input.payroll_run_id",
					"payrollPeriod": "$.input.payroll_period",
					"payrollDate":   "$.input.payroll_date",
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
			// Step 2: Fetch Employee Data - CRITICAL
			{
				StepNumber:    2,
				ServiceName:   "payroll",
				HandlerMethod: "FetchEmployeeData",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"payrollRunID": "$.input.payroll_run_id",
					"lockAcquired": "$.steps.1.result.lock_acquired",
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
				CompensationSteps: []int32{102},
			},
			// Step 3: Calculate Base Salary - CRITICAL
			{
				StepNumber:    3,
				ServiceName:   "salary-structure",
				HandlerMethod: "CalculateBaseSalary",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"payrollRunID":        "$.input.payroll_run_id",
					"payrollPeriod":       "$.input.payroll_period",
					"employeeDataFetched": "$.steps.2.result.employees_fetched",
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
				CompensationSteps: []int32{103},
			},
			// Step 4: Calculate Attendance - CRITICAL
			{
				StepNumber:    4,
				ServiceName:   "attendance",
				HandlerMethod: "CalculateAttendance",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"payrollRunID":     "$.input.payroll_run_id",
					"payrollPeriod":    "$.input.payroll_period",
					"salaryCalculated": "$.steps.3.result.salary_calculated",
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
				CompensationSteps: []int32{104},
			},
			// Step 5: Calculate Deductions - CRITICAL
			{
				StepNumber:    5,
				ServiceName:   "salary-structure",
				HandlerMethod: "CalculateDeductions",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"payrollRunID":         "$.input.payroll_run_id",
					"payrollPeriod":        "$.input.payroll_period",
					"attendanceCalculated": "$.steps.4.result.attendance_calculated",
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
				CompensationSteps: []int32{105},
			},
			// Step 6: Calculate Tax and Contributions - CRITICAL
			{
				StepNumber:    6,
				ServiceName:   "salary-structure",
				HandlerMethod: "CalculateTaxAndContributions",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"payrollRunID":         "$.input.payroll_run_id",
					"payrollPeriod":        "$.input.payroll_period",
					"deductionsCalculated": "$.steps.5.result.deductions_calculated",
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
				CompensationSteps: []int32{106},
			},
			// Step 7: Apply Salary Adjustments - NON-CRITICAL
			{
				StepNumber:    7,
				ServiceName:   "payroll",
				HandlerMethod: "ApplySalaryAdjustments",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"payrollRunID":  "$.input.payroll_run_id",
					"payrollPeriod": "$.input.payroll_period",
					"taxCalculated": "$.steps.6.result.tax_calculated",
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
			// Step 8: Calculate Net Payable - NON-CRITICAL
			{
				StepNumber:    8,
				ServiceName:   "salary-structure",
				HandlerMethod: "CalculateNetPayable",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"payrollRunID":       "$.input.payroll_run_id",
					"payrollPeriod":      "$.input.payroll_period",
					"adjustmentsApplied": "$.steps.7.result.adjustments_applied",
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
			// Step 9: Preview Payroll Entries in General Ledger - NON-CRITICAL
			{
				StepNumber:    9,
				ServiceName:   "general-ledger",
				HandlerMethod: "PreviewPayrollEntries",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"payrollRunID":         "$.input.payroll_run_id",
					"payrollPeriod":        "$.input.payroll_period",
					"netPayableCalculated": "$.steps.8.result.net_payable_calculated",
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
			// Step 10: Authorize Payment Instructions - CRITICAL
			{
				StepNumber:    10,
				ServiceName:   "banking",
				HandlerMethod: "AuthorizePaymentInstructions",
				InputMapping: map[string]string{
					"tenantID":                "$.tenantID",
					"companyID":               "$.companyID",
					"branchID":                "$.branchID",
					"payrollRunID":            "$.input.payroll_run_id",
					"payrollDate":             "$.input.payroll_date",
					"payrollEntriesPreviewed": "$.steps.9.result.entries_previewed",
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
				CompensationSteps: []int32{110},
			},
			// Step 11: Accrue Employee Payables - NON-CRITICAL
			{
				StepNumber:    11,
				ServiceName:   "accounts-payable",
				HandlerMethod: "AccrueEmployeePayables",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"payrollRunID":      "$.input.payroll_run_id",
					"payrollPeriod":     "$.input.payroll_period",
					"paymentAuthorized": "$.steps.10.result.payment_authorized",
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
				CompensationSteps: []int32{111},
			},
			// Step 12: Post Payroll Entries to General Ledger - CRITICAL
			{
				StepNumber:    12,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostPayrollEntries",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"payrollRunID":    "$.input.payroll_run_id",
					"payrollPeriod":   "$.input.payroll_period",
					"payablesAccrued": "$.steps.11.result.payables_accrued",
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
				CompensationSteps: []int32{112},
			},
			// Step 13: Finalize Payroll Run - CRITICAL
			{
				StepNumber:    13,
				ServiceName:   "payroll",
				HandlerMethod: "FinalizePayrollRun",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"payrollRunID":  "$.input.payroll_run_id",
					"entriesPosted": "$.steps.12.result.entries_posted",
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

			// ===== COMPENSATION STEPS (102-113) =====

			// Step 102: Undo Fetch Employee Data (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "payroll",
				HandlerMethod: "UndoFetchEmployeeData",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"payrollRunID": "$.input.payroll_run_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Undo Calculate Base Salary (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "salary-structure",
				HandlerMethod: "UndoCalculateBaseSalary",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"payrollRunID": "$.input.payroll_run_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Undo Calculate Attendance (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "attendance",
				HandlerMethod: "UndoCalculateAttendance",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"payrollRunID": "$.input.payroll_run_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: Undo Calculate Deductions (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "salary-structure",
				HandlerMethod: "UndoCalculateDeductions",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"payrollRunID": "$.input.payroll_run_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: Undo Calculate Tax and Contributions (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "salary-structure",
				HandlerMethod: "UndoCalculateTaxAndContributions",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"payrollRunID": "$.input.payroll_run_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Undo Apply Salary Adjustments (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "payroll",
				HandlerMethod: "UndoApplySalaryAdjustments",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"payrollRunID": "$.input.payroll_run_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: Undo Calculate Net Payable (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "salary-structure",
				HandlerMethod: "UndoCalculateNetPayable",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"payrollRunID": "$.input.payroll_run_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 109: Undo Preview Payroll Entries (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "general-ledger",
				HandlerMethod: "UndoPreviewPayrollEntries",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"payrollRunID": "$.input.payroll_run_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 110: Undo Authorize Payment Instructions (compensates step 10)
			{
				StepNumber:    110,
				ServiceName:   "banking",
				HandlerMethod: "UndoAuthorizePaymentInstructions",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"payrollRunID": "$.input.payroll_run_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 111: Undo Accrue Employee Payables (compensates step 11)
			{
				StepNumber:    111,
				ServiceName:   "accounts-payable",
				HandlerMethod: "UndoAccrueEmployeePayables",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"payrollRunID": "$.input.payroll_run_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 112: Undo Post Payroll Entries (compensates step 12)
			{
				StepNumber:    112,
				ServiceName:   "general-ledger",
				HandlerMethod: "UndoPostPayrollEntries",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"payrollRunID": "$.input.payroll_run_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *PayrollProcessingSaga) SagaType() string {
	return "SAGA-H01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *PayrollProcessingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *PayrollProcessingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: payroll_run_id, payroll_period, payroll_date, company_id
func (s *PayrollProcessingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	// Extract the nested 'input' object
	innerInput, ok := inputMap["input"].(map[string]interface{})
	if !ok {
		return errors.New("missing 'input' field in saga input")
	}

	// Validate payroll_run_id
	if innerInput["payroll_run_id"] == nil {
		return errors.New("missing required field: payroll_run_id")
	}
	payrollRunID, ok := innerInput["payroll_run_id"].(string)
	if !ok || payrollRunID == "" {
		return errors.New("payroll_run_id must be a non-empty string")
	}

	// Validate payroll_period
	if innerInput["payroll_period"] == nil {
		return errors.New("missing required field: payroll_period")
	}
	payrollPeriod, ok := innerInput["payroll_period"].(string)
	if !ok || payrollPeriod == "" {
		return errors.New("payroll_period must be a non-empty string")
	}

	// Validate payroll_date
	if innerInput["payroll_date"] == nil {
		return errors.New("missing required field: payroll_date")
	}
	payrollDate, ok := innerInput["payroll_date"].(string)
	if !ok || payrollDate == "" {
		return errors.New("payroll_date must be a non-empty string")
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
