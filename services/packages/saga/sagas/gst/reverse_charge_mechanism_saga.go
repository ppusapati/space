// Package gst provides saga handlers for GST compliance workflows
package gst

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ReverseChargeMechanismSaga implements SAGA-G07: Reverse Charge Mechanism (RCM) workflow
// Business Flow: CheckRCMEligibility → VerifyVendorRegistration → CalculateRCMLiability → CreateRCMJournal → PostRCMLedger → UpdateComplianceCalendar → DetermineITCEligibility → PostGLEntries → ArchiveRCMRecord
// GST Compliance: RCM applies when recipient is liable for GST on notified supplies (imports, e-commerce, specific services)
type ReverseChargeMechanismSaga struct {
	steps []*saga.StepDefinition
}

// NewReverseChargeMechanismSaga creates a new Reverse Charge Mechanism saga handler
func NewReverseChargeMechanismSaga() saga.SagaHandler {
	return &ReverseChargeMechanismSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Check if supply falls under RCM list (notified goods/services)
			{
				StepNumber:    1,
				ServiceName:   "gst",
				HandlerMethod: "CheckRCMEligibility",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"invoiceID":           "$.input.invoice_id",
					"supplierId":          "$.input.supplier_id",
					"itemCategory":        "$.input.item_category",
					"rcmNotificationList": "$.input.rcm_notification_list",
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
			// Step 2: Verify vendor registration (if unregistered vendor)
			{
				StepNumber:    2,
				ServiceName:   "vendor",
				HandlerMethod: "VerifyVendorRegistration",
				InputMapping: map[string]string{
					"supplierId":      "$.input.supplier_id",
					"gstin":           "$.input.input.vendor_gstin",
					"registrationTyp": "$.input.vendor_registration_type",
				},
				TimeoutSeconds: 25,
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
			// Step 3: Calculate RCM liability (GST amount on purchase)
			{
				StepNumber:    3,
				ServiceName:   "tax-engine",
				HandlerMethod: "CalculateRCMLiability",
				InputMapping: map[string]string{
					"invoiceID":         "$.input.invoice_id",
					"invoiceAmount":     "$.input.invoice_amount",
					"taxRate":           "$.input.tax_rate",
					"supplyType":        "$.input.supply_type",
					"sgstAmount":        "$.input.sgst_amount",
					"cgstAmount":        "$.input.cgst_amount",
					"igstAmount":        "$.input.igst_amount",
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
			// Step 4: Create RCM journal entry (Vendor Payable DR, RCM Payable CR)
			{
				StepNumber:    4,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "CreateRCMJournalEntry",
				InputMapping: map[string]string{
					"invoiceID":      "$.input.invoice_id",
					"supplierId":     "$.input.supplier_id",
					"rcmAmount":      "$.steps.3.result.total_rcm_amount",
					"sgstAmount":     "$.steps.3.result.rcm_sgst",
					"cgstAmount":     "$.steps.3.result.rcm_cgst",
					"igstAmount":     "$.steps.3.result.rcm_igst",
					"journalDate":    "$.input.transaction_date",
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
			// Step 5: Post to GST RCM ledger
			{
				StepNumber:    5,
				ServiceName:   "gst-ledger",
				HandlerMethod: "PostRCMLedgerEntry",
				InputMapping: map[string]string{
					"invoiceID":      "$.input.invoice_id",
					"rcmAmount":      "$.steps.3.result.total_rcm_amount",
					"supplierGSTIN":  "$.input.supplier_gstin",
					"ledgerPeriod":   "$.input.transaction_period",
					"journalEntry":   "$.steps.4.result.journal_entry_id",
					"supplyType":     "$.input.supply_type",
				},
				TimeoutSeconds: 25,
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
			// Step 6: Update compliance calendar with RCM posting
			{
				StepNumber:    6,
				ServiceName:   "compliance-postings",
				HandlerMethod: "UpdateComplianceCalendarRCM",
				InputMapping: map[string]string{
					"invoiceID":     "$.input.invoice_id",
					"rcmAmount":     "$.steps.3.result.total_rcm_amount",
					"postingPeriod": "$.input.transaction_period",
					"gstinNumber":   "$.input.gstin",
					"complianceType": "RCM_POSTING",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Determine if ITC available on RCM (depends on supply type)
			{
				StepNumber:    7,
				ServiceName:   "gst",
				HandlerMethod: "DetermineITCEligibility",
				InputMapping: map[string]string{
					"invoiceID":     "$.input.invoice_id",
					"supplyType":    "$.input.supply_type",
					"rcmAmount":     "$.steps.3.result.total_rcm_amount",
					"invoiceAmount": "$.input.invoice_amount",
					"supplyCategory": "$.input.supply_category",
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
			// Step 8: Post GL entries (Supplier Payable adjustment)
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostRCMGLEntries",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"invoiceID":       "$.input.invoice_id",
					"rcmAmount":       "$.steps.3.result.total_rcm_amount",
					"journalDate":     "$.input.transaction_date",
					"itcEligible":     "$.steps.7.result.itc_eligible",
					"supplierId":      "$.input.supplier_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     true,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Archive RCM record for returns
			{
				StepNumber:    9,
				ServiceName:   "gst",
				HandlerMethod: "ArchiveRCMRecord",
				InputMapping: map[string]string{
					"invoiceID":    "$.input.invoice_id",
					"rcmAmount":    "$.steps.3.result.total_rcm_amount",
					"archiveDate":  "$.input.transaction_date",
					"archiveReason": "RCM processing complete",
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

			// Step 103: Revert RCM Calculation (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "tax-engine",
				HandlerMethod: "RevertRCMCalculation",
				InputMapping: map[string]string{
					"invoiceID": "$.input.invoice_id",
					"reason":    "Saga compensation - RCM processing failed",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Reverse RCM Journal Entry (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "purchase-invoice",
				HandlerMethod: "ReverseRCMJournalEntry",
				InputMapping: map[string]string{
					"invoiceID":      "$.input.invoice_id",
					"journalEntryID": "$.steps.4.result.journal_entry_id",
					"rcmAmount":      "$.steps.3.result.total_rcm_amount",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: Revert GST RCM Ledger Entry (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "gst-ledger",
				HandlerMethod: "RevertRCMLedgerEntry",
				InputMapping: map[string]string{
					"invoiceID": "$.input.invoice_id",
					"rcmAmount": "$.steps.3.result.total_rcm_amount",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Remove Compliance Calendar RCM Entry (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "compliance-postings",
				HandlerMethod: "RemoveComplianceCalendarRCM",
				InputMapping: map[string]string{
					"invoiceID": "$.input.invoice_id",
					"rcmAmount": "$.steps.3.result.total_rcm_amount",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 108: Reverse RCM GL Entries (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseRCMGLEntries",
				InputMapping: map[string]string{
					"invoiceID": "$.input.invoice_id",
					"rcmAmount": "$.steps.3.result.total_rcm_amount",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ReverseChargeMechanismSaga) SagaType() string {
	return "SAGA-G07"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ReverseChargeMechanismSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ReverseChargeMechanismSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ReverseChargeMechanismSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["invoice_id"] == nil {
		return errors.New("invoice_id is required")
	}

	if inputMap["supplier_id"] == nil {
		return errors.New("supplier_id is required")
	}

	if inputMap["item_category"] == nil {
		return errors.New("item_category is required (IMPORT, ECOMMERCE, WORKS_CONTRACT, SERVICE, OTHER)")
	}

	itemCategory, ok := inputMap["item_category"].(string)
	if !ok {
		return errors.New("item_category must be a string")
	}

	validCategories := map[string]bool{
		"IMPORT":         true,
		"ECOMMERCE":      true,
		"WORKS_CONTRACT": true,
		"SERVICE":        true,
		"OTHER":          true,
	}

	if !validCategories[itemCategory] {
		return errors.New("item_category must be IMPORT, ECOMMERCE, WORKS_CONTRACT, SERVICE, or OTHER")
	}

	if inputMap["invoice_amount"] == nil {
		return errors.New("invoice_amount is required")
	}

	if inputMap["tax_rate"] == nil {
		return errors.New("tax_rate is required (5, 12, 18, 28)")
	}

	if inputMap["supply_type"] == nil {
		return errors.New("supply_type is required")
	}

	if inputMap["transaction_date"] == nil {
		return errors.New("transaction_date is required")
	}

	if inputMap["gstin"] == nil {
		return errors.New("gstin is required")
	}

	if inputMap["rcm_notification_list"] == nil {
		return errors.New("rcm_notification_list is required")
	}

	return nil
}
