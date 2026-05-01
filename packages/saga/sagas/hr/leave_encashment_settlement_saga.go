// Package hr provides saga handlers for HR & Payroll module workflows
package hr

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// LeaveEncashmentSettlementSaga implements SAGA-SR04: Leave Encashment & Settlement (Full & Final) workflow
// Business Flow: Identify employee exit → Extract leave balance → Calculate leave encashment →
// Calculate gratuity (4.81% for 5+ years) → Calculate other dues → Create full & final settlement entry →
// Post GL: Settlement Expense DR, Payable CR → Generate settlement report (8 forward steps + 8 compensation = 16 total)
// Critical steps: 2, 5, 7 (3 critical)
// Non-critical steps: 1, 3, 4, 6, 8 (5 non-critical)
// Timeout: 240s aggregate (mix of 20-45s per step)
// Statutory compliance: Gratuity Act (4.81% for 5+ years service, capped at 10 lakhs)
type LeaveEncashmentSettlementSaga struct {
	steps []*saga.StepDefinition
}

// NewLeaveEncashmentSettlementSaga creates a new Leave Encashment & Settlement saga handler
func NewLeaveEncashmentSettlementSaga() saga.SagaHandler {
	return &LeaveEncashmentSettlementSaga{
		steps: []*saga.StepDefinition{
			// ===== FORWARD STEPS (1-8) =====

			// Step 1: Identify Employee Exit - NON-CRITICAL
			{
				StepNumber:    1,
				ServiceName:   "employee",
				HandlerMethod: "IdentifyEmployeeExit",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"settlementID": "$.input.settlement_id",
					"employeeID":   "$.input.employee_id",
					"exitType":     "$.input.exit_type",
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
			// Step 2: Extract Leave Balance - CRITICAL
			{
				StepNumber:    2,
				ServiceName:   "leave",
				HandlerMethod: "ExtractLeaveBalance",
				InputMapping: map[string]string{
					"tenantID":               "$.tenantID",
					"companyID":              "$.companyID",
					"branchID":               "$.branchID",
					"settlementID":           "$.input.settlement_id",
					"employeeID":             "$.input.employee_id",
					"employeeExitIdentified": "$.steps.1.result.employee_exit_identified",
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
				CompensationSteps: []int32{101},
			},
			// Step 3: Calculate Leave Encashment - NON-CRITICAL
			{
				StepNumber:    3,
				ServiceName:   "payroll",
				HandlerMethod: "CalculateLeaveEncashment",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"settlementID":          "$.input.settlement_id",
					"employeeID":            "$.input.employee_id",
					"leaveBalanceExtracted": "$.steps.2.result.leave_balance_extracted",
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
				CompensationSteps: []int32{102},
			},
			// Step 4: Calculate Gratuity (4.81% for 5+ years) - NON-CRITICAL
			{
				StepNumber:    4,
				ServiceName:   "payroll",
				HandlerMethod: "CalculateGratuity",
				InputMapping: map[string]string{
					"tenantID":                  "$.tenantID",
					"companyID":                 "$.companyID",
					"branchID":                  "$.branchID",
					"settlementID":              "$.input.settlement_id",
					"employeeID":                "$.input.employee_id",
					"leaveEncashmentCalculated": "$.steps.3.result.leave_encashment_calculated",
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
			// Step 5: Calculate Other Dues - CRITICAL
			{
				StepNumber:    5,
				ServiceName:   "payroll",
				HandlerMethod: "CalculateOtherDues",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"settlementID":       "$.input.settlement_id",
					"employeeID":         "$.input.employee_id",
					"gratuityCalculated": "$.steps.4.result.gratuity_calculated",
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
			// Step 6: Create Full & Final Settlement Entry (Payable) - NON-CRITICAL
			{
				StepNumber:    6,
				ServiceName:   "accounts-payable",
				HandlerMethod: "CreateSettlementPayable",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"settlementID":        "$.input.settlement_id",
					"employeeID":          "$.input.employee_id",
					"otherDuesCalculated": "$.steps.5.result.other_dues_calculated",
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
			// Step 7: Post GL - Settlement Expense DR, Payable CR - CRITICAL
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostSettlementGLEntries",
				InputMapping: map[string]string{
					"tenantID":                 "$.tenantID",
					"companyID":                "$.companyID",
					"branchID":                 "$.branchID",
					"settlementID":             "$.input.settlement_id",
					"employeeID":               "$.input.employee_id",
					"settlementPayableCreated": "$.steps.6.result.settlement_payable_created",
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
				CompensationSteps: []int32{106},
			},
			// Step 8: Generate Settlement Report (For Manager Approval) - NON-CRITICAL
			{
				StepNumber:    8,
				ServiceName:   "approval",
				HandlerMethod: "GenerateSettlementReport",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"settlementID":    "$.input.settlement_id",
					"employeeID":      "$.input.employee_id",
					"glEntriesPosted": "$.steps.7.result.gl_entries_posted",
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

			// ===== COMPENSATION STEPS (101-107) =====

			// Step 101: Revert Extract Leave Balance (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "leave",
				HandlerMethod: "UndoExtractLeaveBalance",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"settlementID": "$.input.settlement_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Revert Calculate Leave Encashment (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "payroll",
				HandlerMethod: "UndoCalculateLeaveEncashment",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"settlementID": "$.input.settlement_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Revert Calculate Gratuity (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "payroll",
				HandlerMethod: "UndoCalculateGratuity",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"settlementID": "$.input.settlement_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Revert Calculate Other Dues (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "payroll",
				HandlerMethod: "UndoCalculateOtherDues",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"settlementID": "$.input.settlement_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: Revert Create Settlement Payable (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "accounts-payable",
				HandlerMethod: "UndoCreateSettlementPayable",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"settlementID": "$.input.settlement_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: Reverse GL Posting (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "UndoPostSettlementGLEntries",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"settlementID": "$.input.settlement_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Undo Generate Settlement Report (compensates step 8)
			{
				StepNumber:    107,
				ServiceName:   "approval",
				HandlerMethod: "UndoGenerateSettlementReport",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"settlementID": "$.input.settlement_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *LeaveEncashmentSettlementSaga) SagaType() string {
	return "SAGA-SR04"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *LeaveEncashmentSettlementSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *LeaveEncashmentSettlementSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: settlement_id, employee_id, exit_type, company_id
func (s *LeaveEncashmentSettlementSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	// Extract the nested 'input' object
	innerInput, ok := inputMap["input"].(map[string]interface{})
	if !ok {
		return errors.New("missing 'input' field in saga input")
	}

	// Validate settlement_id
	if innerInput["settlement_id"] == nil {
		return errors.New("missing required field: settlement_id")
	}
	settlementID, ok := innerInput["settlement_id"].(string)
	if !ok || settlementID == "" {
		return errors.New("settlement_id must be a non-empty string")
	}

	// Validate employee_id
	if innerInput["employee_id"] == nil {
		return errors.New("missing required field: employee_id")
	}
	employeeID, ok := innerInput["employee_id"].(string)
	if !ok || employeeID == "" {
		return errors.New("employee_id must be a non-empty string")
	}

	// Validate exit_type
	if innerInput["exit_type"] == nil {
		return errors.New("missing required field: exit_type")
	}
	exitType, ok := innerInput["exit_type"].(string)
	if !ok || exitType == "" {
		return errors.New("exit_type must be a non-empty string")
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
