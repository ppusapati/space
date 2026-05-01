// Package gst provides saga handlers for GST compliance workflows
package gst

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// GSTPaymentSaga implements SAGA-G06: GST Payment & Tax Settlement workflow
// Business Flow: CalculatePayableAmount → GenerateChallans → ProcessPayment → UpdateGSTLedger → PostPaymentJournal → GenerateReceipt → CompletePayment
// GST Compliance: GST payment and settlement including CGST, SGST, IGST
type GSTPaymentSaga struct {
	steps []*saga.StepDefinition
}

// NewGSTPaymentSaga creates a new GST Payment saga handler
func NewGSTPaymentSaga() saga.SagaHandler {
	return &GSTPaymentSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Calculate Payable GST Amount
			{
				StepNumber:    1,
				ServiceName:   "tax-engine",
				HandlerMethod: "CalculatePayableGSTAmount",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"paymentPeriod":  "$.input.payment_period",
					"paymentType":    "$.input.payment_type",
					"gstin":          "$.input.gstin",
					"fiscalYear":     "$.input.fiscal_year",
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
			// Step 2: Generate Challan for Payment
			{
				StepNumber:    2,
				ServiceName:   "gst-payment",
				HandlerMethod: "GenerateChallanForPayment",
				InputMapping: map[string]string{
					"paymentID":      "$.steps.1.result.payment_id",
					"payableAmount":  "$.steps.1.result.payable_amount",
					"paymentType":    "$.input.payment_type",
					"paymentPeriod":  "$.input.payment_period",
					"gstin":          "$.input.gstin",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{102},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Get Payment Approval
			{
				StepNumber:    3,
				ServiceName:   "approval",
				HandlerMethod: "ApproveGSTPayment",
				InputMapping: map[string]string{
					"paymentID":     "$.steps.1.result.payment_id",
					"payableAmount": "$.steps.1.result.payable_amount",
					"approvalType":  "GST_PAYMENT",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{103},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Process Bank Payment
			{
				StepNumber:    4,
				ServiceName:   "banking",
				HandlerMethod: "ProcessGSTBankPayment",
				InputMapping: map[string]string{
					"paymentID":      "$.steps.1.result.payment_id",
					"payableAmount":  "$.steps.1.result.payable_amount",
					"challanNumber":  "$.steps.2.result.challan_number",
					"bankAccount":    "$.input.bank_account",
					"paymentMethod":  "$.input.payment_method",
				},
				TimeoutSeconds:    40,
				IsCritical:        true,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Update GST Ledger with Payment
			{
				StepNumber:    5,
				ServiceName:   "gst-ledger",
				HandlerMethod: "UpdateGSTLedgerWithPayment",
				InputMapping: map[string]string{
					"paymentID":       "$.steps.1.result.payment_id",
					"payableAmount":   "$.steps.1.result.payable_amount",
					"paymentReference": "$.steps.4.result.payment_reference",
					"paymentType":     "$.input.payment_type",
					"paymentPeriod":   "$.input.payment_period",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Post Payment to General Ledger
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostGSTPaymentJournal",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"paymentID":       "$.steps.1.result.payment_id",
					"payableAmount":   "$.steps.1.result.payable_amount",
					"paymentReference": "$.steps.4.result.payment_reference",
					"journalDate":     "$.input.payment_date",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Generate Payment Receipt
			{
				StepNumber:    7,
				ServiceName:   "gst-payment",
				HandlerMethod: "GeneratePaymentReceipt",
				InputMapping: map[string]string{
					"paymentID":        "$.steps.1.result.payment_id",
					"payableAmount":    "$.steps.1.result.payable_amount",
					"paymentReference": "$.steps.4.result.payment_reference",
					"challanNumber":    "$.steps.2.result.challan_number",
				},
				TimeoutSeconds:    20,
				IsCritical:        false,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Update Accounts Payable
			{
				StepNumber:    8,
				ServiceName:   "accounts-payable",
				HandlerMethod: "UpdateAPWithGSTPayment",
				InputMapping: map[string]string{
					"paymentID":       "$.steps.1.result.payment_id",
					"payableAmount":   "$.steps.1.result.payable_amount",
					"paymentReference": "$.steps.4.result.payment_reference",
				},
				TimeoutSeconds:    20,
				IsCritical:        false,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Complete GST Payment
			{
				StepNumber:    9,
				ServiceName:   "gst-payment",
				HandlerMethod: "CompleteGSTPayment",
				InputMapping: map[string]string{
					"paymentID":        "$.steps.1.result.payment_id",
					"paymentReference": "$.steps.4.result.payment_reference",
					"completionDate":   "$.input.payment_date",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// ===== COMPENSATION STEPS =====

			// Step 101: Cancel Payment Calculation (compensates step 1)
			{
				StepNumber:    101,
				ServiceName:   "tax-engine",
				HandlerMethod: "CancelPayableGSTCalculation",
				InputMapping: map[string]string{
					"paymentID": "$.steps.1.result.payment_id",
					"reason":    "Saga compensation - GST payment failed",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 102: Void Generated Challan (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "gst-payment",
				HandlerMethod: "VoidGeneratedChallan",
				InputMapping: map[string]string{
					"paymentID":     "$.steps.1.result.payment_id",
					"challanNumber": "$.steps.2.result.challan_number",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: Reject Payment Approval (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "approval",
				HandlerMethod: "RejectGSTPaymentApproval",
				InputMapping: map[string]string{
					"paymentID": "$.steps.1.result.payment_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 104: Reverse Bank Payment (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "banking",
				HandlerMethod: "ReverseBankGSTPayment",
				InputMapping: map[string]string{
					"paymentID":        "$.steps.1.result.payment_id",
					"paymentReference": "$.steps.4.result.payment_reference",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
			// Step 105: Revert GST Ledger Update (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "gst-ledger",
				HandlerMethod: "RevertGSTLedgerPaymentUpdate",
				InputMapping: map[string]string{
					"paymentID": "$.steps.1.result.payment_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Reverse Payment Journal (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseGSTPaymentJournal",
				InputMapping: map[string]string{
					"paymentID": "$.steps.1.result.payment_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Delete Payment Receipt (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "gst-payment",
				HandlerMethod: "DeletePaymentReceipt",
				InputMapping: map[string]string{
					"paymentID": "$.steps.1.result.payment_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 108: Revert AP Update (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "accounts-payable",
				HandlerMethod: "RevertAPGSTPaymentUpdate",
				InputMapping: map[string]string{
					"paymentID": "$.steps.1.result.payment_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *GSTPaymentSaga) SagaType() string {
	return "SAGA-G06"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *GSTPaymentSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *GSTPaymentSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *GSTPaymentSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["payment_period"] == nil {
		return errors.New("payment_period is required (format: YYYY-MM)")
	}

	if inputMap["payment_type"] == nil {
		return errors.New("payment_type is required (CGST, SGST, IGST, COMBINED)")
	}

	paymentType, ok := inputMap["payment_type"].(string)
	if !ok {
		return errors.New("payment_type must be a string")
	}

	validTypes := map[string]bool{
		"CGST":     true,
		"SGST":     true,
		"IGST":     true,
		"COMBINED": true,
	}

	if !validTypes[paymentType] {
		return errors.New("payment_type must be CGST, SGST, IGST, or COMBINED")
	}

	if inputMap["gstin"] == nil {
		return errors.New("gstin is required")
	}

	if inputMap["fiscal_year"] == nil {
		return errors.New("fiscal_year is required")
	}

	if inputMap["payment_date"] == nil {
		return errors.New("payment_date is required")
	}

	if inputMap["bank_account"] == nil {
		return errors.New("bank_account is required")
	}

	if inputMap["payment_method"] == nil {
		return errors.New("payment_method is required (NEFT, RTGS, CHEQUE, etc.)")
	}

	return nil
}
