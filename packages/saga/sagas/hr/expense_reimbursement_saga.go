// Package hr provides saga handlers for HR & Payroll module workflows
package hr

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ExpenseReimbursementSaga implements SAGA-H04: Expense Reimbursement workflow
// Business Flow: Submit expense → Request approval → Validate receipts → Calculate reimbursement →
// Update cost center → Create payment → Process banking → Post to GL → Notify employee → Complete reimbursement
type ExpenseReimbursementSaga struct {
	steps []*saga.StepDefinition
}

// NewExpenseReimbursementSaga creates a new Expense Reimbursement saga handler
func NewExpenseReimbursementSaga() saga.SagaHandler {
	return &ExpenseReimbursementSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Submit Expense Report
			{
				StepNumber:    1,
				ServiceName:   "expense",
				HandlerMethod: "SubmitExpenseReport",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"expenseID":       "$.input.expense_id",
					"employeeID":      "$.input.employee_id",
					"amount":          "$.input.amount",
					"costCenterID":    "$.input.cost_center_id",
					"submissionDate":  "$.input.submission_date",
					"category":        "$.input.expense_category",
					"description":     "$.input.description",
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
			// Step 2: Request Approval
			{
				StepNumber:    2,
				ServiceName:   "approval",
				HandlerMethod: "RequestApproval",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"expenseID":   "$.steps.1.result.expense_id",
					"employeeID":  "$.input.employee_id",
					"amount":      "$.input.amount",
					"approverID":  "$.input.approver_id",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{101},
			},
			// Step 3: Validate Receipts
			{
				StepNumber:    3,
				ServiceName:   "expense",
				HandlerMethod: "ValidateReceipts",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"expenseID":  "$.steps.1.result.expense_id",
					"receiptIDs": "$.input.receipt_ids",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{102},
			},
			// Step 4: Calculate Reimbursement
			{
				StepNumber:    4,
				ServiceName:   "expense",
				HandlerMethod: "CalculateReimbursement",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"expenseID":   "$.steps.1.result.expense_id",
					"amount":      "$.input.amount",
					"employeeID":  "$.input.employee_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{103},
			},
			// Step 5: Update Cost Center
			{
				StepNumber:    5,
				ServiceName:   "cost-center",
				HandlerMethod: "UpdateCostCenterExpense",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"costCenterID":  "$.input.cost_center_id",
					"expenseID":     "$.steps.1.result.expense_id",
					"amount":        "$.steps.4.result.reimbursement_amount",
					"category":      "$.input.expense_category",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{104},
			},
			// Step 6: Create Payment Voucher
			{
				StepNumber:    6,
				ServiceName:   "accounts-payable",
				HandlerMethod: "CreatePaymentVoucher",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"expenseID":       "$.steps.1.result.expense_id",
					"employeeID":      "$.input.employee_id",
					"amount":          "$.steps.4.result.reimbursement_amount",
					"voucherType":     "EXPENSE_REIMBURSEMENT",
					"costCenterID":    "$.input.cost_center_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        false,
				CompensationSteps: []int32{105},
			},
			// Step 7: Process Banking Payment
			{
				StepNumber:    7,
				ServiceName:   "banking",
				HandlerMethod: "InitiatePayment",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"voucherID":     "$.steps.6.result.voucher_id",
					"employeeID":    "$.input.employee_id",
					"amount":        "$.steps.4.result.reimbursement_amount",
					"paymentMethod": "BANK_TRANSFER",
				},
				TimeoutSeconds:    20,
				IsCritical:        false,
				CompensationSteps: []int32{106},
			},
			// Step 8: Post to General Ledger
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostExpenseTransaction",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"expenseID":     "$.steps.1.result.expense_id",
					"voucherID":     "$.steps.6.result.voucher_id",
					"amount":        "$.steps.4.result.reimbursement_amount",
					"costCenterID":  "$.input.cost_center_id",
					"accountCode":   "$.input.expense_account_code",
				},
				TimeoutSeconds:    15,
				IsCritical:        false,
				CompensationSteps: []int32{107},
			},
			// Step 9: Notify Employee
			{
				StepNumber:    9,
				ServiceName:   "notification",
				HandlerMethod: "SendNotification",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"employeeID":    "$.input.employee_id",
					"notificationType": "REIMBURSEMENT_PROCESSED",
					"amount":        "$.steps.4.result.reimbursement_amount",
					"expenseID":     "$.steps.1.result.expense_id",
				},
				TimeoutSeconds:    10,
				IsCritical:        false,
				CompensationSteps: []int32{108},
			},
			// Step 10: Complete Reimbursement
			{
				StepNumber:    10,
				ServiceName:   "expense",
				HandlerMethod: "CompleteReimbursement",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"expenseID":   "$.steps.1.result.expense_id",
					"status":      "COMPLETED",
					"paymentDate": "$.current_timestamp",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{},
			},
			// ===== COMPENSATION STEPS =====

			// Step 101: Reject Approval Request (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "approval",
				HandlerMethod: "RejectApprovalRequest",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"approvalID":  "$.steps.2.result.approval_id",
					"reason":      "Cascade rejection due to expense validation failure",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 102: Invalidate Receipt Validation (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "expense",
				HandlerMethod: "InvalidateReceipts",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"expenseID":  "$.steps.1.result.expense_id",
					"receiptIDs": "$.input.receipt_ids",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 103: Revert Reimbursement Calculation (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "expense",
				HandlerMethod: "RevertReimbursementCalculation",
				InputMapping: map[string]string{
					"tenantID":   "$.tenantID",
					"expenseID":  "$.steps.1.result.expense_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 104: Revert Cost Center Update (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "cost-center",
				HandlerMethod: "RevertCostCenterExpense",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"costCenterID":  "$.input.cost_center_id",
					"expenseID":     "$.steps.1.result.expense_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 105: Void Payment Voucher (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "accounts-payable",
				HandlerMethod: "VoidPaymentVoucher",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"voucherID":   "$.steps.6.result.voucher_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 106: Reverse Banking Payment (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "banking",
				HandlerMethod: "ReversePayment",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"voucherID":   "$.steps.6.result.voucher_id",
					"paymentID":   "$.steps.7.result.payment_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: Reverse GL Transaction (compensates step 8)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseTransaction",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"transactionID":  "$.steps.8.result.transaction_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 108: Revoke Notification (compensates step 9)
			{
				StepNumber:    108,
				ServiceName:   "notification",
				HandlerMethod: "RevokeNotification",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"notificationID":   "$.steps.9.result.notification_id",
				},
				TimeoutSeconds: 10,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ExpenseReimbursementSaga) SagaType() string {
	return "SAGA-H04"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ExpenseReimbursementSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ExpenseReimbursementSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ExpenseReimbursementSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["expense_id"] == nil {
		return errors.New("expense_id is required")
	}

	if inputMap["employee_id"] == nil {
		return errors.New("employee_id is required")
	}

	if inputMap["amount"] == nil {
		return errors.New("amount is required")
	}

	if inputMap["cost_center_id"] == nil {
		return errors.New("cost_center_id is required")
	}

	if inputMap["submission_date"] == nil {
		return errors.New("submission_date is required")
	}

	return nil
}
