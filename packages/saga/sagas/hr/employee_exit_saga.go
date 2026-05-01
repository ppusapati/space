// Package hr provides saga handlers for HR & Payroll module workflows
package hr

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// EmployeeExitSaga implements SAGA-H03: Employee Exit & Full and Final Settlement workflow
// Business Flow: Initiate exit → Offboard employee → Calculate final settlement → Process final payment →
// Deactivate assets → Revoke access → Post exit costs → Update employee status → Log exit process →
// Send exit notification
//
// Compensation: If any critical step fails, automatically reverses previous steps
// in reverse order to restore employee data and resources
type EmployeeExitSaga struct {
	steps []*saga.StepDefinition
}

// NewEmployeeExitSaga creates a new Employee Exit saga handler
func NewEmployeeExitSaga() saga.SagaHandler {
	return &EmployeeExitSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Employee Exit (employee service)
			// Initiates the employee exit process with exit date and reason
			{
				StepNumber:    1,
				ServiceName:   "employee",
				HandlerMethod: "InitiateExit",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"companyID":  "$.companyID",
					"branchID":   "$.branchID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
					"reason":     "$.input.reason",
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
			// Step 2: Offboard Employee (employee service)
			// Performs employee offboarding including final document collection
			{
				StepNumber:    2,
				ServiceName:   "employee",
				HandlerMethod: "OffboardEmployee",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
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
				CompensationSteps: []int32{101},
			},
			// Step 3: Calculate Final Settlement (salary-structure service)
			// Calculates final settlement amount including all dues and deductions
			{
				StepNumber:    3,
				ServiceName:   "salary-structure",
				HandlerMethod: "CalculateFinalSettlement",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"employeeID":            "$.input.employee_id",
					"exitDate":              "$.input.exit_date",
					"fullAndFinalDate":      "$.input.full_and_final_date",
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
			// Step 4: Process Final Payment (payroll service)
			// Processes final payroll payment including settlement amount
			{
				StepNumber:    4,
				ServiceName:   "payroll",
				HandlerMethod: "ProcessFinalPayment",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"employeeID":      "$.input.employee_id",
					"exitDate":        "$.input.exit_date",
					"settlementAmount": "$.steps.3.result.settlement_amount",
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
			// Step 5: Deactivate Assets (asset service)
			// Deactivates all assets assigned to the employee
			{
				StepNumber:    5,
				ServiceName:   "asset",
				HandlerMethod: "DeactivateAssets",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
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
			// Step 6: Revoke Access (access service)
			// Revokes all system access and permissions for the employee
			{
				StepNumber:    6,
				ServiceName:   "access",
				HandlerMethod: "RevokeAccess",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{105},
			},
			// Step 7: Post Exit Costs (general-ledger service)
			// Posts exit-related costs to general ledger
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostExitCosts",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"companyID":  "$.companyID",
					"branchID":   "$.branchID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{106},
			},
			// Step 8: Update Employee Status (employee service)
			// Updates employee status to inactive/exit
			{
				StepNumber:    8,
				ServiceName:   "employee",
				HandlerMethod: "UpdateEmployeeStatus",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
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
				CompensationSteps: []int32{107},
			},
			// Step 9: Log Exit Process (exit service)
			// Logs the complete exit process for audit and compliance
			{
				StepNumber:    9,
				ServiceName:   "exit",
				HandlerMethod: "LogExitProcess",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
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
				CompensationSteps: []int32{108},
			},
			// Step 10: Send Exit Notification (notification service)
			// Sends exit notification to relevant stakeholders
			{
				StepNumber:    10,
				ServiceName:   "notification",
				HandlerMethod: "SendExitNotification",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
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
				CompensationSteps: []int32{109},
			},
			// ===== COMPENSATION STEPS =====

			// Compensation Step 101: Revert Offboarding (compensates step 2)
			// Reverts employee offboarding status
			{
				StepNumber:    101,
				ServiceName:   "employee",
				HandlerMethod: "RevertOffboarding",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
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
			// Compensation Step 102: Revert Final Settlement Calculation (compensates step 3)
			// Reverts calculated final settlement
			{
				StepNumber:    102,
				ServiceName:   "salary-structure",
				HandlerMethod: "RevertFinalSettlement",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
				},
				TimeoutSeconds: 45,
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
			// Compensation Step 103: Reverse Final Payment (compensates step 4)
			// Reverses the final payment if possible
			{
				StepNumber:    103,
				ServiceName:   "payroll",
				HandlerMethod: "ReverseFinalPayment",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
				},
				TimeoutSeconds: 45,
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
			// Compensation Step 104: Reactivate Assets (compensates step 5)
			// Reactivates deactivated assets
			{
				StepNumber:    104,
				ServiceName:   "asset",
				HandlerMethod: "ReactivateAssets",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
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
			// Compensation Step 105: Restore Access (compensates step 6)
			// Restores access and permissions for the employee
			{
				StepNumber:    105,
				ServiceName:   "access",
				HandlerMethod: "RestoreAccess",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
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
			// Compensation Step 106: Reverse Exit Costs (compensates step 7)
			// Reverses exit cost GL entries
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseExitCosts",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
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
			// Compensation Step 107: Restore Employee Status (compensates step 8)
			// Restores employee to active status
			{
				StepNumber:    107,
				ServiceName:   "employee",
				HandlerMethod: "RestoreEmployeeStatus",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
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
			// Compensation Step 108: Revert Exit Log (compensates step 9)
			// Reverts/removes exit process log entry
			{
				StepNumber:    108,
				ServiceName:   "exit",
				HandlerMethod: "RevertExitLog",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
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
			// Compensation Step 109: Revoke Exit Notification (compensates step 10)
			// Revokes/cancels exit notification
			{
				StepNumber:    109,
				ServiceName:   "notification",
				HandlerMethod: "RevokeExitNotification",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.input.employee_id",
					"exitDate":   "$.input.exit_date",
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
		},
	}
}

// SagaType returns the saga type identifier
func (s *EmployeeExitSaga) SagaType() string {
	return "SAGA-H03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *EmployeeExitSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *EmployeeExitSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: employee_id, exit_date, full_and_final_date, reason
func (s *EmployeeExitSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type: expected map[string]interface{}")
	}

	// Validate employee_id
	if inputMap["employee_id"] == nil || inputMap["employee_id"] == "" {
		return errors.New("employee_id is required for Employee Exit saga")
	}

	// Validate exit_date
	if inputMap["exit_date"] == nil || inputMap["exit_date"] == "" {
		return errors.New("exit_date is required for Employee Exit saga")
	}

	// Validate full_and_final_date
	if inputMap["full_and_final_date"] == nil || inputMap["full_and_final_date"] == "" {
		return errors.New("full_and_final_date is required for Employee Exit saga")
	}

	// Validate reason
	if inputMap["reason"] == nil || inputMap["reason"] == "" {
		return errors.New("reason is required for Employee Exit saga")
	}

	return nil
}
