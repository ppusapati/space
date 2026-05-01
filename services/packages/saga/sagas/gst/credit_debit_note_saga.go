// Package gst provides saga handlers for GST compliance workflows
package gst

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// CreditDebitNoteSaga implements SAGA-G08: Credit Note & Debit Note Processing workflow
// Business Flow: IdentifyTransactionType → DetermineCNDNType → CalculateCNDNAmount → CalculateGSTImpact → CreateCNDNRecord → UpdateGSTR → PostGLEntries → ArchiveCNDNRecord
// GST Compliance: CN/DN for returns, discounts, tax corrections with GSTR-1/2 impact tracking
type CreditDebitNoteSaga struct {
	steps []*saga.StepDefinition
}

// NewCreditDebitNoteSaga creates a new Credit Note & Debit Note saga handler
func NewCreditDebitNoteSaga() saga.SagaHandler {
	return &CreditDebitNoteSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Identify transaction requiring CN/DN (return, discount, correction)
			{
				StepNumber:    1,
				ServiceName:   "gst",
				HandlerMethod: "IdentifyCNDNTransaction",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"originalInvoiceID":   "$.input.original_invoice_id",
					"transactionType":     "$.input.transaction_type",
					"cndn_reason":         "$.input.cndn_reason",
				},
				TimeoutSeconds: 25,
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
			// Step 2: Determine CN/DN type (full, partial, tax correction)
			{
				StepNumber:    2,
				ServiceName:   "gst",
				HandlerMethod: "DetermineCNDNType",
				InputMapping: map[string]string{
					"originalInvoiceID": "$.input.original_invoice_id",
					"cnDnType":          "$.input.cndn_type",
					"returnPercentage":  "$.input.return_percentage",
					"originalAmount":    "$.input.original_amount",
				},
				TimeoutSeconds: 20,
				IsCritical:     true,
				CompensationSteps: []int32{},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Calculate CN/DN amount (original invoice amount * percentage)
			{
				StepNumber:    3,
				ServiceName:   "tax-engine",
				HandlerMethod: "CalculateCNDNAmount",
				InputMapping: map[string]string{
					"originalInvoiceID": "$.input.original_invoice_id",
					"originalAmount":    "$.input.original_amount",
					"returnPercentage":  "$.input.return_percentage",
					"cnDnType":          "$.input.cndn_type",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				CompensationSteps: []int32{103},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Calculate GST impact (SGST, CGST, IGST adjustment)
			{
				StepNumber:    4,
				ServiceName:   "tax-engine",
				HandlerMethod: "CalculateCNDNGSTImpact",
				InputMapping: map[string]string{
					"originalInvoiceID": "$.input.original_invoice_id",
					"cnDnAmount":        "$.steps.3.result.cndn_amount",
					"originalSGST":      "$.input.original_sgst",
					"originalCGST":      "$.input.original_cgst",
					"originalIGST":      "$.input.original_igst",
					"cnDnType":          "$.input.cndn_type",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Create CN/DN in sales/purchase ledger
			{
				StepNumber:    5,
				ServiceName:   "sales-invoice",
				HandlerMethod: "CreateCNDNLedgerEntry",
				InputMapping: map[string]string{
					"originalInvoiceID": "$.input.original_invoice_id",
					"cnDnAmount":        "$.steps.3.result.cndn_amount",
					"cnDnType":          "$.input.cndn_type",
					"cnDnDate":          "$.input.cndn_date",
					"reason":            "$.input.cndn_reason",
					"sgstAmount":        "$.steps.4.result.cndn_sgst",
					"cgstAmount":        "$.steps.4.result.cndn_cgst",
					"igstAmount":        "$.steps.4.result.cndn_igst",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Update GSTR-1/2 (credits/debits impact next month return)
			{
				StepNumber:    6,
				ServiceName:   "gst-ledger",
				HandlerMethod: "UpdateGSTRWithCNDN",
				InputMapping: map[string]string{
					"originalInvoiceID": "$.input.original_invoice_id",
					"cnDnAmount":        "$.steps.3.result.cndn_amount",
					"gstrType":          "$.input.gstr_type",
					"cnDnPeriod":        "$.input.cndn_period",
					"sgstAmount":        "$.steps.4.result.cndn_sgst",
					"cgstAmount":        "$.steps.4.result.cndn_cgst",
					"igstAmount":        "$.steps.4.result.cndn_igst",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Post GL entries (Sales/Purchase Adjustment)
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostCNDNGLEntries",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"originalInvoiceID": "$.input.original_invoice_id",
					"cnDnAmount":        "$.steps.3.result.cndn_amount",
					"journalDate":       "$.input.cndn_date",
					"cnDnType":          "$.input.cndn_type",
					"sgstAmount":        "$.steps.4.result.cndn_sgst",
					"cgstAmount":        "$.steps.4.result.cndn_cgst",
					"igstAmount":        "$.steps.4.result.cndn_igst",
				},
				TimeoutSeconds: 35,
				IsCritical:     true,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Archive CN/DN record for compliance
			{
				StepNumber:    8,
				ServiceName:   "gst",
				HandlerMethod: "ArchiveCNDNRecord",
				InputMapping: map[string]string{
					"originalInvoiceID": "$.input.original_invoice_id",
					"cnDnAmount":        "$.steps.3.result.cndn_amount",
					"archiveDate":       "$.input.cndn_date",
					"archiveReason":     "CNDN processing complete",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
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

			// Step 103: Revert CN/DN Amount Calculation (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "tax-engine",
				HandlerMethod: "RevertCNDNAmountCalculation",
				InputMapping: map[string]string{
					"originalInvoiceID": "$.input.original_invoice_id",
					"reason":            "Saga compensation - CNDN processing failed",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Revert GST Impact Calculation (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "tax-engine",
				HandlerMethod: "RevertCNDNGSTImpact",
				InputMapping: map[string]string{
					"originalInvoiceID": "$.input.original_invoice_id",
					"cnDnAmount":        "$.steps.3.result.cndn_amount",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: Reverse CN/DN Ledger Entry (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "sales-invoice",
				HandlerMethod: "ReverseCNDNLedgerEntry",
				InputMapping: map[string]string{
					"originalInvoiceID": "$.input.original_invoice_id",
					"cnDnAmount":        "$.steps.3.result.cndn_amount",
					"cnDnType":          "$.input.cndn_type",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: Revert GSTR Update (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "gst-ledger",
				HandlerMethod: "RevertGSTRCNDNUpdate",
				InputMapping: map[string]string{
					"originalInvoiceID": "$.input.original_invoice_id",
					"cnDnAmount":        "$.steps.3.result.cndn_amount",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Reverse GL Entries (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseCNDNGLEntries",
				InputMapping: map[string]string{
					"originalInvoiceID": "$.input.original_invoice_id",
					"cnDnAmount":        "$.steps.3.result.cndn_amount",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *CreditDebitNoteSaga) SagaType() string {
	return "SAGA-G08"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *CreditDebitNoteSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *CreditDebitNoteSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *CreditDebitNoteSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["original_invoice_id"] == nil {
		return errors.New("original_invoice_id is required")
	}

	if inputMap["transaction_type"] == nil {
		return errors.New("transaction_type is required (SALES, PURCHASE)")
	}

	transactionType, ok := inputMap["transaction_type"].(string)
	if !ok {
		return errors.New("transaction_type must be a string")
	}

	validTransactionTypes := map[string]bool{
		"SALES":    true,
		"PURCHASE": true,
	}

	if !validTransactionTypes[transactionType] {
		return errors.New("transaction_type must be SALES or PURCHASE")
	}

	if inputMap["cndn_type"] == nil {
		return errors.New("cndn_type is required (FULL_CREDIT, PARTIAL_CREDIT, TAX_CORRECTION, DISCOUNT)")
	}

	cndnType, ok := inputMap["cndn_type"].(string)
	if !ok {
		return errors.New("cndn_type must be a string")
	}

	validCNDNTypes := map[string]bool{
		"FULL_CREDIT":     true,
		"PARTIAL_CREDIT":  true,
		"TAX_CORRECTION":  true,
		"DISCOUNT":        true,
	}

	if !validCNDNTypes[cndnType] {
		return errors.New("cndn_type must be FULL_CREDIT, PARTIAL_CREDIT, TAX_CORRECTION, or DISCOUNT")
	}

	if inputMap["cndn_reason"] == nil {
		return errors.New("cndn_reason is required (RETURN, DISCOUNT, ERROR, OTHER)")
	}

	cnDnReason, ok := inputMap["cndn_reason"].(string)
	if !ok {
		return errors.New("cndn_reason must be a string")
	}

	validReasons := map[string]bool{
		"RETURN": true,
		"DISCOUNT": true,
		"ERROR":   true,
		"OTHER":   true,
	}

	if !validReasons[cnDnReason] {
		return errors.New("cndn_reason must be RETURN, DISCOUNT, ERROR, or OTHER")
	}

	if inputMap["original_amount"] == nil {
		return errors.New("original_amount is required")
	}

	if inputMap["return_percentage"] == nil {
		return errors.New("return_percentage is required (0-100)")
	}

	if inputMap["original_sgst"] == nil {
		return errors.New("original_sgst is required")
	}

	if inputMap["original_cgst"] == nil {
		return errors.New("original_cgst is required")
	}

	if inputMap["original_igst"] == nil {
		return errors.New("original_igst is required")
	}

	if inputMap["cndn_date"] == nil {
		return errors.New("cndn_date is required")
	}

	if inputMap["cndn_period"] == nil {
		return errors.New("cndn_period is required (format: YYYY-MM)")
	}

	if inputMap["gstr_type"] == nil {
		return errors.New("gstr_type is required (GSTR1 for sales, GSTR2 for purchase)")
	}

	gstrType, ok := inputMap["gstr_type"].(string)
	if !ok {
		return errors.New("gstr_type must be a string")
	}

	validGSTRTypes := map[string]bool{
		"GSTR1": true,
		"GSTR2": true,
	}

	if !validGSTRTypes[gstrType] {
		return errors.New("gstr_type must be GSTR1 or GSTR2")
	}

	return nil
}
