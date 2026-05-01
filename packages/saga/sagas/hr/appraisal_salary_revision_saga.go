// Package hr provides saga handlers for HR & Payroll module workflows
package hr

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// AppraisalSalaryRevisionSaga implements SAGA-H06: Appraisal & Salary Revision workflow
// Business Flow: Complete appraisal → Calculate salary increase → Create salary structure →
// Get approval → Update payroll → Generate offer letter → Effective from date → Notify HR → Complete revision
type AppraisalSalaryRevisionSaga struct {
	steps []*saga.StepDefinition
}

// NewAppraisalSalaryRevisionSaga creates a new Appraisal & Salary Revision saga handler
func NewAppraisalSalaryRevisionSaga() saga.SagaHandler {
	return &AppraisalSalaryRevisionSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Complete Appraisal
			{
				StepNumber:    1,
				ServiceName:   "appraisal",
				HandlerMethod: "CompleteAppraisal",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"appraisalID":        "$.input.appraisal_id",
					"employeeID":         "$.input.employee_id",
					"performanceRating":  "$.input.performance_rating",
					"evaluatorID":        "$.input.evaluator_id",
					"completionDate":     "$.input.completion_date",
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
			// Step 2: Calculate Salary Increase
			{
				StepNumber:    2,
				ServiceName:   "appraisal",
				HandlerMethod: "CalculateSalaryIncrease",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"appraisalID":       "$.steps.1.result.appraisal_id",
					"employeeID":        "$.input.employee_id",
					"performanceRating": "$.input.performance_rating",
					"currentSalary":     "$.input.current_salary",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{101},
			},
			// Step 3: Create New Salary Structure
			{
				StepNumber:    3,
				ServiceName:   "salary-structure",
				HandlerMethod: "CreateSalaryStructure",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"employeeID":        "$.input.employee_id",
					"appraisalID":       "$.steps.1.result.appraisal_id",
					"newSalary":         "$.steps.2.result.new_salary",
					"salaryTemplate":    "$.input.salary_template",
					"baseSalary":        "$.steps.2.result.new_salary",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{102},
			},
			// Step 4: Request Approval for Revision
			{
				StepNumber:    4,
				ServiceName:   "approval",
				HandlerMethod: "RequestApproval",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"appraisalID":       "$.steps.1.result.appraisal_id",
					"employeeID":        "$.input.employee_id",
					"approverID":        "$.input.approver_id",
					"newSalary":         "$.steps.2.result.new_salary",
					"performanceRating": "$.input.performance_rating",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{103},
			},
			// Step 5: Update Payroll Structure
			{
				StepNumber:    5,
				ServiceName:   "payroll",
				HandlerMethod: "UpdatePayrollStructure",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"employeeID":         "$.input.employee_id",
					"salaryStructureID":  "$.steps.3.result.structure_id",
					"newSalary":          "$.steps.2.result.new_salary",
					"effectiveFromDate":  "$.input.effective_from_date",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{104},
			},
			// Step 6: Generate Offer Letter
			{
				StepNumber:    6,
				ServiceName:   "notification",
				HandlerMethod: "GenerateOfferLetter",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"employeeID":       "$.input.employee_id",
					"appraisalID":      "$.steps.1.result.appraisal_id",
					"newSalary":        "$.steps.2.result.new_salary",
					"effectiveFromDate": "$.input.effective_from_date",
				},
				TimeoutSeconds:    15,
				IsCritical:        false,
				CompensationSteps: []int32{105},
			},
			// Step 7: Set Effective From Date
			{
				StepNumber:    7,
				ServiceName:   "appraisal",
				HandlerMethod: "SetEffectiveFromDate",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"appraisalID":       "$.steps.1.result.appraisal_id",
					"employeeID":        "$.input.employee_id",
					"effectiveFromDate": "$.input.effective_from_date",
				},
				TimeoutSeconds:    15,
				IsCritical:        false,
				CompensationSteps: []int32{106},
			},
			// Step 8: Notify HR & Employee
			{
				StepNumber:    8,
				ServiceName:   "notification",
				HandlerMethod: "SendNotification",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"employeeID":        "$.input.employee_id",
					"notificationType":  "SALARY_REVISION_APPROVED",
					"newSalary":         "$.steps.2.result.new_salary",
					"effectiveFromDate": "$.input.effective_from_date",
				},
				TimeoutSeconds:    10,
				IsCritical:        false,
				CompensationSteps: []int32{107},
			},
			// Step 9: Complete Revision
			{
				StepNumber:    9,
				ServiceName:   "appraisal",
				HandlerMethod: "CompleteRevision",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"appraisalID":      "$.steps.1.result.appraisal_id",
					"employeeID":       "$.input.employee_id",
					"revisionStatus":   "COMPLETED",
					"effectiveFromDate": "$.input.effective_from_date",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{},
			},
			// ===== COMPENSATION STEPS =====

			// Step 101: Revert Salary Calculation (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "appraisal",
				HandlerMethod: "RevertSalaryCalculation",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"appraisalID": "$.steps.1.result.appraisal_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 102: Delete Salary Structure (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "salary-structure",
				HandlerMethod: "DeleteSalaryStructure",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"structureID":       "$.steps.3.result.structure_id",
					"employeeID":        "$.input.employee_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 103: Reject Approval Request (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "approval",
				HandlerMethod: "RejectApprovalRequest",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"approvalID":  "$.steps.4.result.approval_id",
					"reason":      "Cascade rejection due to downstream failure",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 104: Revert Payroll Structure Update (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "payroll",
				HandlerMethod: "RevertPayrollStructure",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"employeeID": "$.input.employee_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 105: Revoke Offer Letter (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "notification",
				HandlerMethod: "RevokeOfferLetter",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"employeeID":    "$.input.employee_id",
					"letterID":      "$.steps.6.result.letter_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 106: Revert Effective Date (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "appraisal",
				HandlerMethod: "RevertEffectiveFromDate",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"appraisalID": "$.steps.1.result.appraisal_id",
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
		},
	}
}

// SagaType returns the saga type identifier
func (s *AppraisalSalaryRevisionSaga) SagaType() string {
	return "SAGA-H06"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *AppraisalSalaryRevisionSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *AppraisalSalaryRevisionSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *AppraisalSalaryRevisionSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["appraisal_id"] == nil {
		return errors.New("appraisal_id is required")
	}

	if inputMap["employee_id"] == nil {
		return errors.New("employee_id is required")
	}

	if inputMap["performance_rating"] == nil {
		return errors.New("performance_rating is required")
	}

	if inputMap["new_salary"] == nil {
		return errors.New("new_salary is required")
	}

	return nil
}
