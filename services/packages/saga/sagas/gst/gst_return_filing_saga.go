// Package gst provides saga handlers for GST compliance workflows
package gst

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// GSTReturnFilingSaga implements SAGA-G01: GST Return Filing (Monthly/Quarterly) workflow
// Business Flow: InitiateReturn → ValidateInvoices → CalculateTaxLiability → ReconcileGSTR2 → FinalizeReturn → PostLedgerEntries → FileReturn → CompleteReturn
// GST Compliance: Filing GSTR-1, GSTR-2, GSTR-3B returns as per prescribed timelines
type GSTReturnFilingSaga struct {
	steps []*saga.StepDefinition
}

// NewGSTReturnFilingSaga creates a new GST Return Filing saga handler
func NewGSTReturnFilingSaga() saga.SagaHandler {
	return &GSTReturnFilingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate GST Return
			{
				StepNumber:    1,
				ServiceName:   "gst-return",
				HandlerMethod: "InitiateGSTReturn",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"returnPeriod":  "$.input.return_period",
					"returnType":    "$.input.return_type",
					"gstin":         "$.input.gstin",
					"fiscalYear":    "$.input.fiscal_year",
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
			// Step 2: Validate Invoices for Return
			{
				StepNumber:    2,
				ServiceName:   "gst-return",
				HandlerMethod: "ValidateInvoicesForReturn",
				InputMapping: map[string]string{
					"returnID":     "$.steps.1.result.return_id",
					"returnPeriod": "$.input.return_period",
					"returnType":   "$.input.return_type",
					"validateGST":  "true",
				},
				TimeoutSeconds:    30,
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
			// Step 3: Calculate Tax Liability
			{
				StepNumber:    3,
				ServiceName:   "tax-engine",
				HandlerMethod: "CalculateGSTLiability",
				InputMapping: map[string]string{
					"returnID":       "$.steps.1.result.return_id",
					"invoiceData":    "$.steps.2.result.invoice_data",
					"returnPeriod":   "$.input.return_period",
					"taxCalculationBasis": "$.input.tax_calculation_basis",
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
			// Step 4: Reconcile with GSTR-2 Data
			{
				StepNumber:    4,
				ServiceName:   "gst-ledger",
				HandlerMethod: "ReconcileGSTR2Data",
				InputMapping: map[string]string{
					"returnID":      "$.steps.1.result.return_id",
					"taxLiability":  "$.steps.3.result.tax_liability",
					"returnPeriod":  "$.input.return_period",
					"gstin":         "$.input.gstin",
				},
				TimeoutSeconds:    25,
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
			// Step 5: Calculate ITC Utilization
			{
				StepNumber:    5,
				ServiceName:   "gst-ledger",
				HandlerMethod: "CalculateITCUtilization",
				InputMapping: map[string]string{
					"returnID":       "$.steps.1.result.return_id",
					"availableITC":   "$.steps.4.result.available_itc",
					"taxLiability":   "$.steps.3.result.tax_liability",
					"returnPeriod":   "$.input.return_period",
				},
				TimeoutSeconds:    25,
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
			// Step 6: Finalize Return Data
			{
				StepNumber:    6,
				ServiceName:   "gst-return",
				HandlerMethod: "FinalizeReturnData",
				InputMapping: map[string]string{
					"returnID":       "$.steps.1.result.return_id",
					"taxLiability":   "$.steps.3.result.tax_liability",
					"itcUtilization": "$.steps.5.result.itc_utilization",
					"reconciliationStatus": "$.steps.4.result.reconciliation_status",
				},
				TimeoutSeconds:    20,
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
			// Step 7: Post Journal Entries
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostGSTReturnJournal",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"returnID":       "$.steps.1.result.return_id",
					"taxLiability":   "$.steps.3.result.tax_liability",
					"itcUtilization": "$.steps.5.result.itc_utilization",
					"journalDate":    "$.input.return_period",
				},
				TimeoutSeconds:    30,
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
			// Step 8: File Return with Portal
			{
				StepNumber:    8,
				ServiceName:   "gst-return",
				HandlerMethod: "FileReturnWithPortal",
				InputMapping: map[string]string{
					"returnID":  "$.steps.1.result.return_id",
					"returnData": "$.steps.6.result.return_data",
					"gstin":     "$.input.gstin",
					"returnType": "$.input.return_type",
				},
				TimeoutSeconds:    40,
				IsCritical:        true,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Generate Return Confirmation
			{
				StepNumber:    9,
				ServiceName:   "gst-return",
				HandlerMethod: "GenerateReturnConfirmation",
				InputMapping: map[string]string{
					"returnID":        "$.steps.1.result.return_id",
					"filingStatus":    "$.steps.8.result.filing_status",
					"referenceNumber": "$.steps.8.result.reference_number",
				},
				TimeoutSeconds:    15,
				IsCritical:        false,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Complete Return Filing
			{
				StepNumber:    10,
				ServiceName:   "gst-return",
				HandlerMethod: "CompleteReturnFiling",
				InputMapping: map[string]string{
					"returnID":          "$.steps.1.result.return_id",
					"filingConfirmation": "$.steps.9.result.filing_confirmation",
					"completionDate":    "$.input.return_period",
				},
				TimeoutSeconds:    15,
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

			// Step 101: Cancel Return Initiation (compensates step 1)
			{
				StepNumber:    101,
				ServiceName:   "gst-return",
				HandlerMethod: "CancelReturnInitiation",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
					"reason":   "Saga compensation - GST return filing failed",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 102: Clear Invoice Validation (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "gst-return",
				HandlerMethod: "ClearInvoiceValidation",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: Clear Tax Calculation (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "tax-engine",
				HandlerMethod: "ClearGSTLiabilityCalculation",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 104: Revert GSTR-2 Reconciliation (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "gst-ledger",
				HandlerMethod: "RevertGSTR2Reconciliation",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Revert ITC Calculation (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "gst-ledger",
				HandlerMethod: "RevertITCUtilizationCalculation",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Revert Return Finalization (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "gst-return",
				HandlerMethod: "RevertReturnFinalization",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: Reverse Journal Entries (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseGSTReturnJournal",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: Withdraw Filed Return (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "gst-return",
				HandlerMethod: "WithdrawFiledReturn",
				InputMapping: map[string]string{
					"returnID":        "$.steps.1.result.return_id",
					"referenceNumber": "$.steps.8.result.reference_number",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
			// Step 109: Delete Return Confirmation (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "gst-return",
				HandlerMethod: "DeleteReturnConfirmation",
				InputMapping: map[string]string{
					"returnID": "$.steps.1.result.return_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *GSTReturnFilingSaga) SagaType() string {
	return "SAGA-G01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *GSTReturnFilingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *GSTReturnFilingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *GSTReturnFilingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["return_period"] == nil {
		return errors.New("return_period is required (format: YYYY-MM)")
	}

	if inputMap["return_type"] == nil {
		return errors.New("return_type is required (GSTR1, GSTR2, GSTR3B)")
	}

	returnType, ok := inputMap["return_type"].(string)
	if !ok {
		return errors.New("return_type must be a string")
	}

	validTypes := map[string]bool{
		"GSTR1":  true,
		"GSTR2":  true,
		"GSTR3B": true,
	}

	if !validTypes[returnType] {
		return errors.New("return_type must be GSTR1, GSTR2, or GSTR3B")
	}

	if inputMap["gstin"] == nil {
		return errors.New("gstin is required")
	}

	if inputMap["fiscal_year"] == nil {
		return errors.New("fiscal_year is required")
	}

	if inputMap["tax_calculation_basis"] == nil {
		return errors.New("tax_calculation_basis is required")
	}

	return nil
}
