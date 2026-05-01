// Package hr provides saga handlers for HR & Payroll module workflows
package hr

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// EmployeeOnboardingSaga implements SAGA-H02: Employee Onboarding workflow
// Business Flow: Create employee record → Create user account → Grant system access → Create salary structure →
// Setup payroll → Assign asset → Send notification → Complete onboarding
type EmployeeOnboardingSaga struct {
	steps []*saga.StepDefinition
}

// NewEmployeeOnboardingSaga creates a new Employee Onboarding saga handler
func NewEmployeeOnboardingSaga() saga.SagaHandler {
	return &EmployeeOnboardingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Employee Record
			{
				StepNumber:    1,
				ServiceName:   "employee",
				HandlerMethod: "CreateEmployee",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"employeeID":    "$.input.employee_id",
					"firstName":     "$.input.first_name",
					"lastName":      "$.input.last_name",
					"email":         "$.input.email",
					"designation":   "$.input.designation",
					"department":    "$.input.department",
					"joinDate":      "$.input.join_date",
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
			// Step 2: Create User Account
			{
				StepNumber:    2,
				ServiceName:   "user",
				HandlerMethod: "CreateUser",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"employeeID":   "$.steps.1.result.employee_id",
					"email":        "$.input.email",
					"firstName":    "$.input.first_name",
					"lastName":     "$.input.last_name",
					"username":     "$.input.email",
					"initialRole":  "EMPLOYEE",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{101},
			},
			// Step 3: Grant System Access
			{
				StepNumber:    3,
				ServiceName:   "access",
				HandlerMethod: "GrantAccessRole",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"userID":        "$.steps.2.result.user_id",
					"employeeID":    "$.steps.1.result.employee_id",
					"accessLevel":   "EMPLOYEE_STANDARD",
					"effectiveFrom": "$.input.join_date",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{102},
			},
			// Step 4: Create Salary Structure
			{
				StepNumber:    4,
				ServiceName:   "salary-structure",
				HandlerMethod: "CreateStructure",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"employeeID":    "$.steps.1.result.employee_id",
					"salaryTemplate": "$.input.salary_template",
					"baseSalary":    "$.input.base_salary",
					"effectiveFrom": "$.input.join_date",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{103},
			},
			// Step 5: Setup Payroll Employee
			{
				StepNumber:    5,
				ServiceName:   "payroll",
				HandlerMethod: "SetupPayrollEmployee",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"employeeID":      "$.steps.1.result.employee_id",
					"salaryStructure": "$.steps.4.result.structure_id",
					"bankAccount":     "$.input.bank_account",
					"effectiveFrom":   "$.input.join_date",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{104},
			},
			// Step 6: Assign Asset
			{
				StepNumber:    6,
				ServiceName:   "asset",
				HandlerMethod: "AssignToEmployee",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"employeeID":   "$.steps.1.result.employee_id",
					"assetType":    "LAPTOP",
					"assignmentID": "$.input.asset_assignment_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        false,
				CompensationSteps: []int32{105},
			},
			// Step 7: Create Payslip Template
			{
				StepNumber:    7,
				ServiceName:   "payroll",
				HandlerMethod: "CreatePayslipTemplate",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"employeeID":         "$.steps.1.result.employee_id",
					"templateName":       "Standard",
					"salaryStructureID":  "$.steps.4.result.structure_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        false,
				CompensationSteps: []int32{106},
			},
			// Step 8: Send Welcome Email
			{
				StepNumber:    8,
				ServiceName:   "notification",
				HandlerMethod: "SendNotification",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"employeeID":    "$.steps.1.result.employee_id",
					"notificationType": "WELCOME_EMAIL",
					"recipient":     "$.input.email",
					"firstName":     "$.input.first_name",
				},
				TimeoutSeconds:    10,
				IsCritical:        false,
				CompensationSteps: []int32{107},
			},
			// Step 9: Create Employee Folder
			{
				StepNumber:    9,
				ServiceName:   "asset",
				HandlerMethod: "CreateEmployeeFolder",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"employeeID":  "$.steps.1.result.employee_id",
					"firstName":   "$.input.first_name",
					"lastName":    "$.input.last_name",
				},
				TimeoutSeconds:    15,
				IsCritical:        false,
				CompensationSteps: []int32{108},
			},
			// Step 10: Complete Onboarding
			{
				StepNumber:    10,
				ServiceName:   "employee",
				HandlerMethod: "CompleteOnboarding",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"employeeID":  "$.steps.1.result.employee_id",
					"status":      "ACTIVE",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{},
			},
			// ===== COMPENSATION STEPS =====

			// Step 101: Revert User Account Creation (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "user",
				HandlerMethod: "DeactivateUser",
				InputMapping: map[string]string{
					"tenantID": "$.tenantID",
					"userID":   "$.steps.2.result.user_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 102: Revoke System Access (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "access",
				HandlerMethod: "RevokeAccessRole",
				InputMapping: map[string]string{
					"tenantID": "$.tenantID",
					"userID":   "$.steps.2.result.user_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 103: Delete Salary Structure (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "salary-structure",
				HandlerMethod: "DeleteStructure",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"structureID":   "$.steps.4.result.structure_id",
					"employeeID":    "$.steps.1.result.employee_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 104: Revoke Payroll Setup (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "payroll",
				HandlerMethod: "RevokePayrollSetup",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.steps.1.result.employee_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 105: Revoke Asset Assignment (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "asset",
				HandlerMethod: "RevokeAssetAssignment",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"employeeID":    "$.steps.1.result.employee_id",
					"assignmentID":  "$.input.asset_assignment_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 106: Delete Payslip Template (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "payroll",
				HandlerMethod: "DeletePayslipTemplate",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"employeeID":   "$.steps.1.result.employee_id",
					"templateID":   "$.steps.7.result.template_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 107: Revoke Notification (compensates step 8)
			{
				StepNumber:    107,
				ServiceName:   "notification",
				HandlerMethod: "RevokeNotification",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"notificationID":   "$.steps.8.result.notification_id",
				},
				TimeoutSeconds: 10,
				IsCritical:     false,
			},
			// Step 108: Delete Employee Folder (compensates step 9)
			{
				StepNumber:    108,
				ServiceName:   "asset",
				HandlerMethod: "DeleteEmployeeFolder",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.steps.1.result.employee_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *EmployeeOnboardingSaga) SagaType() string {
	return "SAGA-H02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *EmployeeOnboardingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *EmployeeOnboardingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *EmployeeOnboardingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["employee_id"] == nil {
		return errors.New("employee_id is required")
	}

	if inputMap["first_name"] == nil {
		return errors.New("first_name is required")
	}

	if inputMap["last_name"] == nil {
		return errors.New("last_name is required")
	}

	if inputMap["email"] == nil {
		return errors.New("email is required")
	}

	if inputMap["designation"] == nil {
		return errors.New("designation is required")
	}

	if inputMap["department"] == nil {
		return errors.New("department is required")
	}

	return nil
}
