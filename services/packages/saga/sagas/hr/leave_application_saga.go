// Package hr provides saga handlers for HR & Payroll module workflows
package hr

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// LeaveApplicationSaga implements SAGA-H05: Leave Application & Approval workflow
// Business Flow: Apply for leave → Check balance → Get approval → Grant leave → Update calendar → Adjust payroll
type LeaveApplicationSaga struct {
	steps []*saga.StepDefinition
}

// NewLeaveApplicationSaga creates a new Leave Application saga handler
func NewLeaveApplicationSaga() saga.SagaHandler {
	return &LeaveApplicationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Submit Leave Application
			{
				StepNumber:    1,
				ServiceName:   "leave",
				HandlerMethod: "SubmitApplication",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"employeeID":    "$.input.employee_id",
					"leaveType":     "$.input.leave_type",
					"startDate":     "$.input.start_date",
					"endDate":       "$.input.end_date",
					"reason":        "$.input.reason",
				},
				TimeoutSeconds: 15,
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
			// Step 2: Check Leave Balance
			{
				StepNumber:    2,
				ServiceName:   "leave",
				HandlerMethod: "CheckBalance",
				InputMapping: map[string]string{
					"applicationID": "$.steps.1.result.application_id",
					"employeeID":    "$.input.employee_id",
					"leaveType":     "$.input.leave_type",
					"daysRequested": "$.input.days_requested",
				},
				TimeoutSeconds:    10,
				IsCritical:        true,
				CompensationSteps: []int32{},
			},
			// Step 3: Request Leave Approval
			{
				StepNumber:    3,
				ServiceName:   "approval",
				HandlerMethod: "RequestLeaveApproval",
				InputMapping: map[string]string{
					"applicationID": "$.steps.1.result.application_id",
					"managerID":     "$.input.manager_id",
					"employeeID":    "$.input.employee_id",
					"reason":        "$.input.reason",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{105},
			},
			// Step 4: Grant Leave (if approved)
			{
				StepNumber:    4,
				ServiceName:   "leave",
				HandlerMethod: "GrantLeave",
				InputMapping: map[string]string{
					"applicationID": "$.steps.1.result.application_id",
					"employeeID":    "$.input.employee_id",
					"daysRequested": "$.input.days_requested",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{104},
			},
			// Step 5: Block Calendar
			{
				StepNumber:    5,
				ServiceName:   "attendance",
				HandlerMethod: "BlockCalendar",
				InputMapping: map[string]string{
					"employeeID": "$.input.employee_id",
					"startDate":  "$.input.start_date",
					"endDate":    "$.input.end_date",
				},
				TimeoutSeconds:    15,
				IsCritical:        false,
				CompensationSteps: []int32{105},
			},
			// Step 6: Adjust Payroll for LOP (Loss of Pay)
			{
				StepNumber:    6,
				ServiceName:   "payroll",
				HandlerMethod: "AdjustPayrollForLOP",
				InputMapping: map[string]string{
					"applicationID": "$.steps.1.result.application_id",
					"employeeID":    "$.input.employee_id",
					"lopDays":       "$.input.lop_days",
				},
				TimeoutSeconds:    15,
				IsCritical:        false,
				CompensationSteps: []int32{106},
			},
			// Step 7: Notify Employee
			{
				StepNumber:    7,
				ServiceName:   "notification",
				HandlerMethod: "NotifyEmployee",
				InputMapping: map[string]string{
					"applicationID": "$.steps.1.result.application_id",
					"employeeID":    "$.input.employee_id",
					"status":        "APPROVED",
				},
				TimeoutSeconds:    10,
				IsCritical:        false,
				CompensationSteps: []int32{},
			},
			// Step 8: Complete Application
			{
				StepNumber:    8,
				ServiceName:   "leave",
				HandlerMethod: "CompleteApplication",
				InputMapping: map[string]string{
					"applicationID": "$.steps.1.result.application_id",
				},
				TimeoutSeconds:    10,
				IsCritical:        true,
				CompensationSteps: []int32{},
			},
			// ===== COMPENSATION STEPS =====

			// Step 101: Auto-reject (compensates step 3)
			{
				StepNumber:    101,
				ServiceName:   "leave",
				HandlerMethod: "RejectApplication",
				InputMapping: map[string]string{
					"applicationID": "$.steps.1.result.application_id",
					"reason":        "Insufficient balance",
				},
				TimeoutSeconds: 10,
				IsCritical:     false,
			},
			// Step 104: Revert Leave Grant (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "leave",
				HandlerMethod: "RevertLeaveGrant",
				InputMapping: map[string]string{
					"applicationID": "$.steps.1.result.application_id",
					"employeeID":    "$.input.employee_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 105: Revert Calendar Block (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "attendance",
				HandlerMethod: "UnblockCalendar",
				InputMapping: map[string]string{
					"employeeID": "$.input.employee_id",
					"startDate":  "$.input.start_date",
					"endDate":    "$.input.end_date",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 106: Revert Payroll Adjustment (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "payroll",
				HandlerMethod: "RevertPayrollAdjustment",
				InputMapping: map[string]string{
					"applicationID": "$.steps.1.result.application_id",
					"employeeID":    "$.input.employee_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *LeaveApplicationSaga) SagaType() string {
	return "SAGA-H05"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *LeaveApplicationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *LeaveApplicationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *LeaveApplicationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["employee_id"] == nil {
		return errors.New("employee_id is required")
	}

	if inputMap["leave_type"] == nil {
		return errors.New("leave_type is required")
	}

	if inputMap["start_date"] == nil {
		return errors.New("start_date is required")
	}

	if inputMap["end_date"] == nil {
		return errors.New("end_date is required")
	}

	if inputMap["reason"] == nil {
		return errors.New("reason is required")
	}

	if inputMap["manager_id"] == nil {
		return errors.New("manager_id is required")
	}

	return nil
}
