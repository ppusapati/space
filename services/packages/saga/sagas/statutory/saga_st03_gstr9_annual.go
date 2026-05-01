// Package statutory provides saga handlers for statutory compliance workflows
package statutory

import (
	"errors"
	"fmt"

	"p9e.in/samavaya/packages/saga"
)

// GSTR9AnnualSaga implements SAGA-ST03: GSTR-9 Annual Return (Annual GST Reconciliation) workflow
// Business Flow: ExtractGSTR1Data → ExtractGSTR2Data → ReconcileTaxableSupplies →
//               ReconcileITCClaimed → IdentifyDiscrepancies → CalculateAnnualTaxLiability →
//               CalculateTotalITC → CalculateNetTaxPayable → AdjustForPayments →
//               CreateReconciliationReport → SubmitGSTR9 → ArchiveAnnualReturn
// GSTR-9 Compliance: Annual GST reconciliation return, filed within prescribed period
// Complexity: VERY HIGH (reconciliation across 12 months of GSTR-1 and GSTR-2 data)
// Critical Steps: 5 (IdentifyDiscrepancies), 8 (CalculateNetTaxPayable), 11 (SubmitGSTR9)
type GSTR9AnnualSaga struct {
	steps []*saga.StepDefinition
}

// NewGSTR9AnnualSaga creates a new GSTR-9 Annual Return saga handler
func NewGSTR9AnnualSaga() saga.SagaHandler {
	return &GSTR9AnnualSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Extract 12 Months of GSTR-1 Data
			{
				StepNumber:    1,
				ServiceName:   "gst",
				HandlerMethod: "ExtractGSTR1AnnualData",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"gstin":          "$.input.gstin",
					"financialYear":  "$.input.financial_year",
					"startDate":      "$.input.fy_start_date",
					"endDate":        "$.input.fy_end_date",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},
			// Step 2: Extract 12 Months of GSTR-2 Data
			{
				StepNumber:    2,
				ServiceName:   "gst",
				HandlerMethod: "ExtractGSTR2AnnualData",
				InputMapping: map[string]string{
					"gstin":          "$.input.gstin",
					"financialYear":  "$.input.financial_year",
					"startDate":      "$.input.fy_start_date",
					"endDate":        "$.input.fy_end_date",
				},
				TimeoutSeconds:    45,
				IsCritical:        true,
				CompensationSteps: []int32{102},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Reconcile Taxable Supplies (GSTR-1 vs Actual)
			{
				StepNumber:    3,
				ServiceName:   "reconciliation",
				HandlerMethod: "ReconcileTaxableSupplies",
				InputMapping: map[string]string{
					"gstr1AnnualData": "$.steps.1.result.gstr1_annual_data",
					"actualSalesData": "$.input.actual_sales_data",
					"reconciliationRules": "$.input.reconciliation_rules",
				},
				TimeoutSeconds:    40,
				IsCritical:        true,
				CompensationSteps: []int32{103},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Reconcile ITC Claimed (GSTR-2 vs Actual)
			{
				StepNumber:    4,
				ServiceName:   "reconciliation",
				HandlerMethod: "ReconcileITCClaimed",
				InputMapping: map[string]string{
					"gstr2AnnualData": "$.steps.2.result.gstr2_annual_data",
					"actualPurchaseData": "$.input.actual_purchase_data",
					"reconciliationRules": "$.input.reconciliation_rules",
				},
				TimeoutSeconds:    40,
				IsCritical:        true,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Identify Discrepancies (Mismatches, Missing Invoices) - CRITICAL
			{
				StepNumber:    5,
				ServiceName:   "reconciliation",
				HandlerMethod: "IdentifyDiscrepancies",
				InputMapping: map[string]string{
					"taxableSuppliesReconciliation": "$.steps.3.result.reconciliation_data",
					"itcReconciliation":             "$.steps.4.result.itc_reconciliation_data",
					"discrepancyThresholds":         "$.input.discrepancy_thresholds",
				},
				TimeoutSeconds:    40,
				IsCritical:        true,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Calculate Annual Tax Liability
			{
				StepNumber:    6,
				ServiceName:   "gst",
				HandlerMethod: "CalculateAnnualTaxLiability",
				InputMapping: map[string]string{
					"gstr1AnnualData":    "$.steps.1.result.gstr1_annual_data",
					"adjustedSupplies":   "$.steps.3.result.adjusted_supplies",
					"taxCalculationRules": "$.input.tax_calculation_rules",
				},
				TimeoutSeconds:    35,
				IsCritical:        true,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Calculate Total ITC Available
			{
				StepNumber:    7,
				ServiceName:   "gst",
				HandlerMethod: "CalculateTotalITCAvailable",
				InputMapping: map[string]string{
					"gstr2AnnualData":      "$.steps.2.result.gstr2_annual_data",
					"adjustedITC":          "$.steps.4.result.adjusted_itc",
					"itcReversalRecords":   "$.input.itc_reversal_records",
				},
				TimeoutSeconds:    35,
				IsCritical:        true,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Calculate Net Tax Payable/Receivable - CRITICAL
			{
				StepNumber:    8,
				ServiceName:   "gst",
				HandlerMethod: "CalculateNetTaxPayable",
				InputMapping: map[string]string{
					"annualTaxLiability": "$.steps.6.result.annual_tax_liability",
					"totalITCAvailable":  "$.steps.7.result.total_itc_available",
					"taxPaymentsRecords": "$.input.tax_payments_records",
				},
				TimeoutSeconds:    40,
				IsCritical:        true,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Adjust for Payment of Tax (Challans, Returns)
			{
				StepNumber:    9,
				ServiceName:   "general-ledger",
				HandlerMethod: "AdjustForTaxPayments",
				InputMapping: map[string]string{
					"netTaxPayable":      "$.steps.8.result.net_tax_payable",
					"challanPayments":    "$.input.challan_payments",
					"refundClaims":       "$.input.refund_claims",
					"adjustmentRecords":  "$.input.adjustment_records",
				},
				TimeoutSeconds:    35,
				IsCritical:        false,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Create Reconciliation Report
			{
				StepNumber:    10,
				ServiceName:   "reconciliation",
				HandlerMethod: "CreateAnnualReconciliationReport",
				InputMapping: map[string]string{
					"gstin":                   "$.input.gstin",
					"financialYear":           "$.input.financial_year",
					"discrepancies":           "$.steps.5.result.discrepancies",
					"reconciliationSummary":   "$.steps.3.result.reconciliation_summary",
					"itcReconciliationSummary": "$.steps.4.result.itc_reconciliation_summary",
					"finalNetTaxPayable":      "$.steps.9.result.final_net_tax_payable",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{110},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 11: Submit GSTR-9 to Authority - CRITICAL
			{
				StepNumber:    11,
				ServiceName:   "compliance-postings",
				HandlerMethod: "SubmitGSTR9",
				InputMapping: map[string]string{
					"gstin":                 "$.input.gstin",
					"financialYear":         "$.input.financial_year",
					"reconciliationReport":  "$.steps.10.result.reconciliation_report",
					"discrepancies":         "$.steps.5.result.discrepancies",
					"dscCertificate":        "$.input.dsc_certificate",
				},
				TimeoutSeconds:    50,
				IsCritical:        true,
				CompensationSteps: []int32{111},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        5,
					InitialBackoffMs:  3000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 12: Archive Annual Return
			{
				StepNumber:    12,
				ServiceName:   "compliance-postings",
				HandlerMethod: "ArchiveGSTR9",
				InputMapping: map[string]string{
					"gstin":                "$.input.gstin",
					"financialYear":        "$.input.financial_year",
					"reconciliationReport": "$.steps.10.result.reconciliation_report",
					"submissionResponse":   "$.steps.11.result.submission_response",
					"discrepancies":        "$.steps.5.result.discrepancies",
				},
				TimeoutSeconds:    35,
				IsCritical:        false,
				CompensationSteps: []int32{},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  2000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// ===== COMPENSATION STEPS =====

			// Step 102: Clear GSTR-2 Data Extraction (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "gst",
				HandlerMethod: "ClearGSTR2DataExtraction",
				InputMapping: map[string]string{
					"financialYear": "$.input.financial_year",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Clear Taxable Supplies Reconciliation (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "reconciliation",
				HandlerMethod: "ClearTaxableSuppliesReconciliation",
				InputMapping: map[string]string{
					"financialYear": "$.input.financial_year",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Clear ITC Reconciliation (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "reconciliation",
				HandlerMethod: "ClearITCReconciliation",
				InputMapping: map[string]string{
					"financialYear": "$.input.financial_year",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: Clear Discrepancy Identification (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "reconciliation",
				HandlerMethod: "ClearDiscrepancyIdentification",
				InputMapping: map[string]string{
					"financialYear": "$.input.financial_year",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: Clear Annual Tax Liability (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "gst",
				HandlerMethod: "ClearAnnualTaxLiability",
				InputMapping: map[string]string{
					"financialYear": "$.input.financial_year",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Clear Total ITC Calculation (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "gst",
				HandlerMethod: "ClearTotalITCCalculation",
				InputMapping: map[string]string{
					"financialYear": "$.input.financial_year",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: Clear Net Tax Payable Calculation (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "gst",
				HandlerMethod: "ClearNetTaxPayableCalculation",
				InputMapping: map[string]string{
					"financialYear": "$.input.financial_year",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 109: Reverse Tax Payment Adjustments (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseTaxPaymentAdjustments",
				InputMapping: map[string]string{
					"financialYear": "$.input.financial_year",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 110: Delete Reconciliation Report (compensates step 10)
			{
				StepNumber:    110,
				ServiceName:   "reconciliation",
				HandlerMethod: "DeleteReconciliationReport",
				InputMapping: map[string]string{
					"gstin":          "$.input.gstin",
					"financialYear":  "$.input.financial_year",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 111: Revert GSTR-9 Status on Submission Failure (compensates step 11)
			{
				StepNumber:    111,
				ServiceName:   "compliance-postings",
				HandlerMethod: "RollbackGSTR9Status",
				InputMapping: map[string]string{
					"gstin":         "$.input.gstin",
					"financialYear": "$.input.financial_year",
					"reason":        "GSTR-9 submission failed",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *GSTR9AnnualSaga) SagaType() string {
	return "SAGA-ST03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *GSTR9AnnualSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *GSTR9AnnualSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters for GSTR-9 annual return
func (s *GSTR9AnnualSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type for GSTR-9 annual return")
	}

	// Validate GSTIN format (15 characters, alphanumeric)
	if inputMap["gstin"] == nil {
		return errors.New("gstin is required for GSTR-9 annual return")
	}
	gstin, ok := inputMap["gstin"].(string)
	if !ok || len(gstin) != 15 {
		return errors.New("gstin must be a 15-character string")
	}

	// Validate financial year format (YYYY-YY)
	if inputMap["financial_year"] == nil {
		return errors.New("financial_year is required (format: YYYY-YY)")
	}
	financialYear, ok := inputMap["financial_year"].(string)
	if !ok || len(financialYear) != 7 {
		return fmt.Errorf("financial_year must be in YYYY-YY format, got: %s", financialYear)
	}

	// Validate FY start and end dates
	if inputMap["fy_start_date"] == nil {
		return errors.New("fy_start_date is required")
	}
	if inputMap["fy_end_date"] == nil {
		return errors.New("fy_end_date is required")
	}

	// Validate actual sales data for reconciliation
	if inputMap["actual_sales_data"] == nil {
		return errors.New("actual_sales_data is required for GSTR-1 reconciliation")
	}

	// Validate actual purchase data for reconciliation
	if inputMap["actual_purchase_data"] == nil {
		return errors.New("actual_purchase_data is required for GSTR-2 reconciliation")
	}

	// Validate reconciliation rules
	if inputMap["reconciliation_rules"] == nil {
		return errors.New("reconciliation_rules are required")
	}

	// Validate discrepancy thresholds
	if inputMap["discrepancy_thresholds"] == nil {
		return errors.New("discrepancy_thresholds are required for identifying significant discrepancies")
	}

	// Validate tax calculation rules
	if inputMap["tax_calculation_rules"] == nil {
		return errors.New("tax_calculation_rules are required")
	}

	// Validate ITC reversal records
	if inputMap["itc_reversal_records"] == nil {
		return errors.New("itc_reversal_records are required")
	}

	// Validate tax payments records
	if inputMap["tax_payments_records"] == nil {
		return errors.New("tax_payments_records are required for payment adjustments")
	}

	// Validate challan payments
	if inputMap["challan_payments"] == nil {
		return errors.New("challan_payments are required")
	}

	// Validate refund claims
	if inputMap["refund_claims"] == nil {
		return errors.New("refund_claims are required")
	}

	// Validate adjustment records
	if inputMap["adjustment_records"] == nil {
		return errors.New("adjustment_records are required")
	}

	// Validate DSC certificate
	if inputMap["dsc_certificate"] == nil {
		return errors.New("dsc_certificate is required for GSTR-9 submission")
	}

	return nil
}
