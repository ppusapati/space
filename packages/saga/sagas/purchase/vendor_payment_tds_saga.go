// Package purchase provides saga handlers for purchase module workflows
package purchase

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// VendorPaymentTDSSaga implements SAGA-P03: Vendor Payment with TDS workflow
// Business Flow: Validate Payment → Calculate TDS → Deduct TDS → Record Withholding → Pay Vendor → Update AP → Post GL → Record Challan → Update TDS Return
type VendorPaymentTDSSaga struct {
	steps []*saga.StepDefinition
}

// NewVendorPaymentTDSSaga creates a new Vendor Payment with TDS saga handler
func NewVendorPaymentTDSSaga() saga.SagaHandler {
	return &VendorPaymentTDSSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Validate Payment
			{
				StepNumber:    1,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ValidatePayment",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"vendorID":      "$.input.vendor_id",
					"invoiceID":     "$.input.invoice_id",
					"paymentAmount": "$.input.payment_amount",
				},
				TimeoutSeconds: 20,
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
			// Step 2: Calculate TDS
			{
				StepNumber:    2,
				ServiceName:   "tds",
				HandlerMethod: "CalculateTDS",
				InputMapping: map[string]string{
					"vendorID":       "$.input.vendor_id",
					"paymentAmount":  "$.input.payment_amount",
					"tdsSection":     "$.input.tds_section",
					"tdsRate":        "$.input.tds_rate",
					"fiscalYear":     "$.input.fiscal_year",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{102},
			},
			// Step 3: Deduct TDS
			{
				StepNumber:    3,
				ServiceName:   "tds",
				HandlerMethod: "DeductTDS",
				InputMapping: map[string]string{
					"paymentAmount":  "$.input.payment_amount",
					"tdsAmount":      "$.steps.2.result.tds_amount",
					"vendorID":       "$.input.vendor_id",
					"invoiceID":      "$.input.invoice_id",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{103},
			},
			// Step 4: Record Withholding
			{
				StepNumber:    4,
				ServiceName:   "tds",
				HandlerMethod: "RecordWithholding",
				InputMapping: map[string]string{
					"vendorID":      "$.input.vendor_id",
					"witholdAmount": "$.steps.2.result.tds_amount",
					"tdsSection":    "$.input.tds_section",
					"invoiceID":     "$.input.invoice_id",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{104},
			},
			// Step 5: Process Net Payment
			{
				StepNumber:    5,
				ServiceName:   "banking",
				HandlerMethod: "ProcessNetPayment",
				InputMapping: map[string]string{
					"vendorID":       "$.input.vendor_id",
					"netAmount":      "$.steps.3.result.net_amount",
					"paymentMethod":  "$.input.payment_method",
					"bankAccountID":  "$.input.bank_account_id",
					"invoiceID":      "$.input.invoice_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        5,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      120000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{105},
			},
			// Step 6: Update AP for Payment
			{
				StepNumber:    6,
				ServiceName:   "accounts-payable",
				HandlerMethod: "UpdateAPForPayment",
				InputMapping: map[string]string{
					"invoiceID":     "$.input.invoice_id",
					"vendorID":      "$.input.vendor_id",
					"paymentAmount": "$.steps.3.result.net_amount",
					"witholdingAmount": "$.steps.2.result.tds_amount",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{106},
			},
			// Step 7: Post TDS Journal
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostTDSJournal",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"invoiceID":      "$.input.invoice_id",
					"vendorID":       "$.input.vendor_id",
					"netAmount":      "$.steps.3.result.net_amount",
					"tdsAmount":      "$.steps.2.result.tds_amount",
					"journalDate":    "$.input.journal_date",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{107},
			},
			// Step 8: Record Challan
			{
				StepNumber:    8,
				ServiceName:   "tds",
				HandlerMethod: "RecordChallan",
				InputMapping: map[string]string{
					"vendorID":       "$.input.vendor_id",
					"witholdAmount":  "$.steps.2.result.tds_amount",
					"tdsSection":     "$.input.tds_section",
					"challanDate":    "$.input.challan_date",
					"fiscalYear":     "$.input.fiscal_year",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{108},
			},
			// Step 9: Update TDS Return
			{
				StepNumber:    9,
				ServiceName:   "tds",
				HandlerMethod: "UpdateTDSReturn",
				InputMapping: map[string]string{
					"vendorID":      "$.input.vendor_id",
					"witholdAmount": "$.steps.2.result.tds_amount",
					"tdsSection":    "$.input.tds_section",
					"fiscalYear":    "$.input.fiscal_year",
				},
				TimeoutSeconds:    20,
				IsCritical:        false, // Non-critical: update can happen later
				CompensationSteps: []int32{109},
			},
			// ===== COMPENSATION STEPS =====

			// Step 102: Reverse TDS Calculation (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "tds",
				HandlerMethod: "ReverseTDSCalculation",
				InputMapping: map[string]string{
					"vendorID":     "$.input.vendor_id",
					"invoiceID":    "$.input.invoice_id",
					"tdsAmount":    "$.steps.2.result.tds_amount",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 103: Restore TDS Deduction (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "tds",
				HandlerMethod: "RestoreTDSDeduction",
				InputMapping: map[string]string{
					"vendorID":    "$.input.vendor_id",
					"invoiceID":   "$.input.invoice_id",
					"tdsAmount":   "$.steps.2.result.tds_amount",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 104: Cancel Withholding (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "tds",
				HandlerMethod: "CancelWithholding",
				InputMapping: map[string]string{
					"vendorID":    "$.input.vendor_id",
					"invoiceID":   "$.input.invoice_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 105: Void Payment (compensates step 5) - PARTIAL
			{
				StepNumber:    105,
				ServiceName:   "banking",
				HandlerMethod: "VoidPayment",
				InputMapping: map[string]string{
					"invoiceID": "$.input.invoice_id",
					"vendorID":  "$.input.vendor_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      60000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 106: Restore AP Entry (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "accounts-payable",
				HandlerMethod: "RestoreAPEntry",
				InputMapping: map[string]string{
					"invoiceID": "$.input.invoice_id",
					"vendorID":  "$.input.vendor_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: Reverse TDS Journal (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseTDSJournal",
				InputMapping: map[string]string{
					"invoiceID": "$.input.invoice_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 108: Remove Challan Mapping (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "tds",
				HandlerMethod: "RemoveChallanMapping",
				InputMapping: map[string]string{
					"vendorID":   "$.input.vendor_id",
					"invoiceID":  "$.input.invoice_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 109: Revert TDS Return (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "tds",
				HandlerMethod: "RevertTDSReturn",
				InputMapping: map[string]string{
					"vendorID":   "$.input.vendor_id",
					"invoiceID":  "$.input.invoice_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *VendorPaymentTDSSaga) SagaType() string {
	return "SAGA-P03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *VendorPaymentTDSSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *VendorPaymentTDSSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *VendorPaymentTDSSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["vendor_id"] == nil {
		return errors.New("vendor_id is required")
	}

	if inputMap["invoice_id"] == nil {
		return errors.New("invoice_id is required")
	}

	if inputMap["payment_amount"] == nil {
		return errors.New("payment_amount is required")
	}

	if inputMap["tds_section"] == nil {
		return errors.New("tds_section is required")
	}

	if inputMap["tds_rate"] == nil {
		return errors.New("tds_rate is required")
	}

	return nil
}
