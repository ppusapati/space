// Package statutory provides saga handlers for statutory compliance workflows
package statutory

import (
	"errors"
	"fmt"

	"p9e.in/samavaya/packages/saga"
)

// GSTR2ITCSaga implements SAGA-ST02: GSTR-2 ITC Claim (Input Tax Claim) workflow
// Business Flow: ExtractPurchaseInvoices → FilterEligibleSupplies → ValidateVendorDetails →
//               CalculateEligibleITC → CheckITCReversalRules → CalculateNetITC →
//               VerifyOutputTaxRule → GenerateGSTR2Schedules → UpdateITCLedger → ArchiveGSTR2
// GSTR-2 Compliance: GST Return for inward supplies/ITC claim, filed monthly by last day of month
// Critical Steps: 4 (CalculateEligibleITC), 6 (CalculateNetITC), 9 (UpdateITCLedger)
type GSTR2ITCSaga struct {
	steps []*saga.StepDefinition
}

// NewGSTR2ITCSaga creates a new GSTR-2 ITC Claim saga handler
func NewGSTR2ITCSaga() saga.SagaHandler {
	return &GSTR2ITCSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Extract Purchase Invoices (Period-based)
			{
				StepNumber:    1,
				ServiceName:   "gst",
				HandlerMethod: "ExtractPurchaseInvoices",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"gstin":           "$.input.gstin",
					"filingPeriod":    "$.input.filing_period",
					"periodStartDate": "$.input.period_start_date",
					"periodEndDate":   "$.input.period_end_date",
				},
				TimeoutSeconds: 35,
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
			// Step 2: Filter Eligible Supplies (Taxable, Registered Vendor)
			{
				StepNumber:    2,
				ServiceName:   "gst",
				HandlerMethod: "FilterEligibleSupplies",
				InputMapping: map[string]string{
					"purchaseInvoices":    "$.steps.1.result.purchase_invoices",
					"eligibilityRules":    "$.input.eligibility_rules",
					"exemptSupplyList":    "$.input.exempt_supply_list",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{102},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Validate Vendor PAN/GSTIN
			{
				StepNumber:    3,
				ServiceName:   "gst",
				HandlerMethod: "ValidateVendorDetails",
				InputMapping: map[string]string{
					"eligibleInvoices": "$.steps.2.result.eligible_invoices",
					"vendorMaster":     "$.input.vendor_master",
					"gstinRegistry":    "$.input.gstin_registry",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{103},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Calculate Eligible ITC (Actual GST Paid) - CRITICAL
			{
				StepNumber:    4,
				ServiceName:   "gst",
				HandlerMethod: "CalculateEligibleITC",
				InputMapping: map[string]string{
					"validatedInvoices":    "$.steps.3.result.validated_invoices",
					"itcCalculationRules":  "$.input.itc_calculation_rules",
					"invoiceClassification": "$.input.invoice_classification",
				},
				TimeoutSeconds:    35,
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
			// Step 5: Check ITC Reversal Rules (Blocked Credits)
			{
				StepNumber:    5,
				ServiceName:   "gst",
				HandlerMethod: "CheckITCReversalRules",
				InputMapping: map[string]string{
					"calculatedITC":      "$.steps.4.result.calculated_itc",
					"reversalRules":      "$.input.reversal_rules",
					"personalUseList":    "$.input.personal_use_list",
					"exemptSupplyList":   "$.input.exempt_supply_list",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Calculate Net ITC Available - CRITICAL
			{
				StepNumber:    6,
				ServiceName:   "gst",
				HandlerMethod: "CalculateNetITC",
				InputMapping: map[string]string{
					"totalCalculatedITC": "$.steps.4.result.calculated_itc",
					"reversalAmount":     "$.steps.5.result.reversal_amount",
					"previousCreditAdjustments": "$.input.previous_credit_adjustments",
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
			// Step 7: Verify Output Tax ≥ Input Tax (Availability Rule)
			{
				StepNumber:    7,
				ServiceName:   "gst",
				HandlerMethod: "VerifyOutputTaxAvailabilityRule",
				InputMapping: map[string]string{
					"netITC":           "$.steps.6.result.net_itc",
					"outputTaxLiability": "$.input.output_tax_liability",
					"previousPeriodITC": "$.input.previous_period_itc",
				},
				TimeoutSeconds:    25,
				IsCritical:        true,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Generate GSTR-2 Schedules
			{
				StepNumber:    8,
				ServiceName:   "gst",
				HandlerMethod: "GenerateGSTR2Schedules",
				InputMapping: map[string]string{
					"gstin":              "$.input.gstin",
					"filingPeriod":       "$.input.filing_period",
					"availableITC":       "$.steps.7.result.available_itc",
					"eligibleInvoices":   "$.steps.2.result.eligible_invoices",
					"reversalDetails":    "$.steps.5.result.reversal_details",
				},
				TimeoutSeconds:    30,
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
			// Step 9: Update ITC Ledger - CRITICAL
			{
				StepNumber:    9,
				ServiceName:   "gst",
				HandlerMethod: "UpdateITCLedger",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"gstin":           "$.input.gstin",
					"filingPeriod":    "$.input.filing_period",
					"availableITC":    "$.steps.7.result.available_itc",
					"gstr2Schedules":  "$.steps.8.result.gstr2_schedules",
				},
				TimeoutSeconds:    35,
				IsCritical:        true,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Archive GSTR-2 for Audit
			{
				StepNumber:    10,
				ServiceName:   "compliance-postings",
				HandlerMethod: "ArchiveGSTR2",
				InputMapping: map[string]string{
					"gstin":             "$.input.gstin",
					"filingPeriod":      "$.input.filing_period",
					"gstr2Schedules":    "$.steps.8.result.gstr2_schedules",
					"availableITC":      "$.steps.7.result.available_itc",
					"reversalDetails":   "$.steps.5.result.reversal_details",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
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

			// Step 102: Clear Eligibility Filtering (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "gst",
				HandlerMethod: "ClearEligibilityFiltering",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 103: Clear Vendor Validation (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "gst",
				HandlerMethod: "ClearVendorValidation",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 104: Clear ITC Calculations (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "gst",
				HandlerMethod: "ClearITCCalculations",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 105: Clear Reversal Checks (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "gst",
				HandlerMethod: "ClearReversalChecks",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: Clear Net ITC Calculation (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "gst",
				HandlerMethod: "ClearNetITCCalculation",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: Reverse Availability Rule Verification (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "gst",
				HandlerMethod: "ReverseAvailabilityRuleVerification",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 108: Delete GSTR-2 Schedules (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "gst",
				HandlerMethod: "DeleteGSTR2Schedules",
				InputMapping: map[string]string{
					"filingPeriod": "$.input.filing_period",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 109: Reverse ITC Ledger Entries (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "gst",
				HandlerMethod: "ReverseITCLedgerEntries",
				InputMapping: map[string]string{
					"gstin":        "$.input.gstin",
					"filingPeriod": "$.input.filing_period",
					"reason":       "GSTR-2 ITC claim reversal",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *GSTR2ITCSaga) SagaType() string {
	return "SAGA-ST02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *GSTR2ITCSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *GSTR2ITCSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters for GSTR-2 ITC claim
func (s *GSTR2ITCSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type for GSTR-2 ITC claim")
	}

	// Validate GSTIN format (15 characters, alphanumeric)
	if inputMap["gstin"] == nil {
		return errors.New("gstin is required for GSTR-2 ITC claim")
	}
	gstin, ok := inputMap["gstin"].(string)
	if !ok || len(gstin) != 15 {
		return errors.New("gstin must be a 15-character string")
	}

	// Validate filing period
	if inputMap["filing_period"] == nil {
		return errors.New("filing_period is required (format: YYYY-MM)")
	}
	filingPeriod, ok := inputMap["filing_period"].(string)
	if !ok || len(filingPeriod) != 7 {
		return fmt.Errorf("filing_period must be in YYYY-MM format, got: %s", filingPeriod)
	}

	// Validate period dates
	if inputMap["period_start_date"] == nil {
		return errors.New("period_start_date is required")
	}
	if inputMap["period_end_date"] == nil {
		return errors.New("period_end_date is required")
	}

	// Validate eligibility rules
	if inputMap["eligibility_rules"] == nil {
		return errors.New("eligibility_rules are required")
	}

	// Validate exempt supply list
	if inputMap["exempt_supply_list"] == nil {
		return errors.New("exempt_supply_list is required for filtering")
	}

	// Validate vendor master
	if inputMap["vendor_master"] == nil {
		return errors.New("vendor_master is required for vendor validation")
	}

	// Validate GSTIN registry
	if inputMap["gstin_registry"] == nil {
		return errors.New("gstin_registry is required for vendor GSTIN validation")
	}

	// Validate ITC calculation rules
	if inputMap["itc_calculation_rules"] == nil {
		return errors.New("itc_calculation_rules are required")
	}

	// Validate reversal rules (personal use, exempt supplies)
	if inputMap["reversal_rules"] == nil {
		return errors.New("reversal_rules are required for ITC compliance")
	}

	// Validate personal use list
	if inputMap["personal_use_list"] == nil {
		return errors.New("personal_use_list is required for reversal checking")
	}

	// Validate output tax liability
	if inputMap["output_tax_liability"] == nil {
		return errors.New("output_tax_liability is required for availability rule verification")
	}

	// Validate previous period ITC (for matching against GSTR-1 data)
	if inputMap["previous_period_itc"] == nil {
		return errors.New("previous_period_itc is required for continuity")
	}

	return nil
}
