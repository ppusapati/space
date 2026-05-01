// Package projects provides saga handlers for projects module workflows
package projects

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// SubcontractorPaymentSaga implements SAGA-PR03: Sub-Contractor Payment
// Business Flow: Get contractor invoice → Validate invoice → Calculate TDS → Deduct retention → Create payment voucher → Post GL entry → Process banking → Notify contractor → Complete payment
type SubcontractorPaymentSaga struct {
	steps []*saga.StepDefinition
}

// NewSubcontractorPaymentSaga creates a new Sub-Contractor Payment saga handler
func NewSubcontractorPaymentSaga() saga.SagaHandler {
	return &SubcontractorPaymentSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Get Contractor Invoice
			{
				StepNumber:    1,
				ServiceName:   "sub-contractor",
				HandlerMethod: "GetContractorInvoice",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"invoiceID":       "$.input.invoice_id",
					"contractorID":    "$.input.contractor_id",
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
			// Step 2: Validate Invoice
			{
				StepNumber:    2,
				ServiceName:   "sub-contractor",
				HandlerMethod: "ValidateInvoice",
				InputMapping: map[string]string{
					"invoiceID":       "$.input.invoice_id",
					"contractorID":    "$.input.contractor_id",
					"invoiceAmount":   "$.input.invoice_amount",
					"contractDetails": "$.steps.1.result.contract_details",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{101},
			},
			// Step 3: Calculate TDS Deduction
			{
				StepNumber:    3,
				ServiceName:   "tds",
				HandlerMethod: "CalculateTDSOnContractor",
				InputMapping: map[string]string{
					"invoiceID":       "$.input.invoice_id",
					"contractorID":    "$.input.contractor_id",
					"invoiceAmount":   "$.input.invoice_amount",
					"deductionType":   "$.input.deduction_type",
					"tdsApplicable":   "$.steps.1.result.tds_applicable",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{102},
			},
			// Step 4: Deduct Retention Money
			{
				StepNumber:    4,
				ServiceName:   "sub-contractor",
				HandlerMethod: "CalculateRetentionDeduction",
				InputMapping: map[string]string{
					"invoiceID":       "$.input.invoice_id",
					"contractorID":    "$.input.contractor_id",
					"invoiceAmount":   "$.input.invoice_amount",
					"retentionRate":   "$.steps.1.result.retention_percentage",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{103},
			},
			// Step 5: Calculate Payment Amount
			{
				StepNumber:    5,
				ServiceName:   "sub-contractor",
				HandlerMethod: "CalculateNetPayment",
				InputMapping: map[string]string{
					"invoiceID":       "$.input.invoice_id",
					"invoiceAmount":   "$.input.invoice_amount",
					"tdsAmount":       "$.steps.3.result.tds_amount",
					"retentionAmount": "$.steps.4.result.retention_amount",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{104},
			},
			// Step 6: Create Payment Voucher
			{
				StepNumber:    6,
				ServiceName:   "accounts-payable",
				HandlerMethod: "CreateVendorPayment",
				InputMapping: map[string]string{
					"invoiceID":       "$.input.invoice_id",
					"contractorID":    "$.input.contractor_id",
					"paymentAmount":   "$.steps.5.result.net_payment_amount",
					"paymentDate":     "$.steps.1.result.invoice_date",
					"tdsAmount":       "$.steps.3.result.tds_amount",
					"retentionAmount": "$.steps.4.result.retention_amount",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{105},
			},
			// Step 7: Post Payment GL Entry
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostPaymentEntry",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"voucherID":      "$.steps.6.result.voucher_id",
					"invoiceID":      "$.input.invoice_id",
					"paymentAmount":  "$.steps.5.result.net_payment_amount",
					"tdsAmount":      "$.steps.3.result.tds_amount",
					"journalDate":    "$.steps.1.result.invoice_date",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{106},
			},
			// Step 8: Process Banking Payment
			{
				StepNumber:    8,
				ServiceName:   "banking",
				HandlerMethod: "InitiateVendorPayment",
				InputMapping: map[string]string{
					"voucherID":      "$.steps.6.result.voucher_id",
					"contractorID":   "$.input.contractor_id",
					"paymentAmount":  "$.steps.5.result.net_payment_amount",
					"paymentMethod":  "$.steps.1.result.payment_method",
					"bankAccountID":  "$.steps.1.result.bank_account_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        5,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      120000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{107},
			},
			// Step 9: Complete Payment
			{
				StepNumber:    9,
				ServiceName:   "sub-contractor",
				HandlerMethod: "CompletePayment",
				InputMapping: map[string]string{
					"invoiceID":      "$.input.invoice_id",
					"contractorID":   "$.input.contractor_id",
					"voucherID":      "$.steps.6.result.voucher_id",
					"paymentAmount":  "$.steps.5.result.net_payment_amount",
					"paymentStatus":  "Completed",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{},
			},
			// ===== COMPENSATION STEPS =====

			// Step 101: Revert Invoice Validation (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "sub-contractor",
				HandlerMethod: "RevertInvoiceValidation",
				InputMapping: map[string]string{
					"invoiceID": "$.input.invoice_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 102: Reverse TDS Calculation (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "tds",
				HandlerMethod: "ReverseTDSCalculation",
				InputMapping: map[string]string{
					"invoiceID": "$.input.invoice_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 103: Reverse Retention Deduction (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "sub-contractor",
				HandlerMethod: "ReverseRetentionDeduction",
				InputMapping: map[string]string{
					"invoiceID": "$.input.invoice_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 104: Reverse Payment Amount Calculation (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "sub-contractor",
				HandlerMethod: "ReversePaymentCalculation",
				InputMapping: map[string]string{
					"invoiceID": "$.input.invoice_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 105: Cancel Payment Voucher (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "accounts-payable",
				HandlerMethod: "CancelVendorPayment",
				InputMapping: map[string]string{
					"voucherID": "$.steps.6.result.voucher_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 106: Reverse Payment GL Entry (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReversePaymentEntry",
				InputMapping: map[string]string{
					"voucherID": "$.steps.6.result.voucher_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 107: Void Banking Payment (compensates step 8)
			{
				StepNumber:    107,
				ServiceName:   "banking",
				HandlerMethod: "VoidVendorPayment",
				InputMapping: map[string]string{
					"voucherID": "$.steps.6.result.voucher_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      60000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *SubcontractorPaymentSaga) SagaType() string {
	return "SAGA-PR03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *SubcontractorPaymentSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *SubcontractorPaymentSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *SubcontractorPaymentSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["invoice_id"] == nil {
		return errors.New("invoice_id is required")
	}

	if inputMap["contractor_id"] == nil {
		return errors.New("contractor_id is required")
	}

	if inputMap["invoice_amount"] == nil {
		return errors.New("invoice_amount is required")
	}

	if inputMap["deduction_type"] == nil {
		return errors.New("deduction_type is required")
	}

	return nil
}
