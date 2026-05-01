// Package gst provides saga handlers for GST compliance workflows
package gst

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ITCReconciliationSaga implements SAGA-G02: ITC (Input Tax Credit) Reconciliation workflow
// Business Flow: InitiateReconciliation → FetchGSTR2Data → CalculateAvailableITC → MatchSupplierInvoices → ReconcileBilledItems → PostAdjustments → CompleteReconciliation
// GST Compliance: Reconcile ITC between GSTR-2 and books of accounts
type ITCReconciliationSaga struct {
	steps []*saga.StepDefinition
}

// NewITCReconciliationSaga creates a new ITC Reconciliation saga handler
func NewITCReconciliationSaga() saga.SagaHandler {
	return &ITCReconciliationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate ITC Reconciliation
			{
				StepNumber:    1,
				ServiceName:   "gst-ledger",
				HandlerMethod: "InitiateITCReconciliation",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"reconciliationPeriod": "$.input.reconciliation_period",
					"itcType":              "$.input.itc_type",
					"gstin":                "$.input.gstin",
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
			// Step 2: Fetch GSTR-2 Data
			{
				StepNumber:    2,
				ServiceName:   "gst-return",
				HandlerMethod: "FetchGSTR2Data",
				InputMapping: map[string]string{
					"reconciliationID":   "$.steps.1.result.reconciliation_id",
					"reconciliationPeriod": "$.input.reconciliation_period",
					"gstin":              "$.input.gstin",
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
			// Step 3: Calculate Available ITC
			{
				StepNumber:    3,
				ServiceName:   "gst-ledger",
				HandlerMethod: "CalculateAvailableITC",
				InputMapping: map[string]string{
					"reconciliationID":  "$.steps.1.result.reconciliation_id",
					"gstr2Data":         "$.steps.2.result.gstr2_data",
					"itcType":           "$.input.itc_type",
					"reconciliationPeriod": "$.input.reconciliation_period",
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
			// Step 4: Match Supplier Invoices
			{
				StepNumber:    4,
				ServiceName:   "accounts-payable",
				HandlerMethod: "MatchSupplierInvoicesWithGST",
				InputMapping: map[string]string{
					"reconciliationID":   "$.steps.1.result.reconciliation_id",
					"gstr2Data":          "$.steps.2.result.gstr2_data",
					"reconciliationPeriod": "$.input.reconciliation_period",
				},
				TimeoutSeconds:    30,
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
			// Step 5: Identify Discrepancies
			{
				StepNumber:    5,
				ServiceName:   "reconciliation",
				HandlerMethod: "IdentifyITCDiscrepancies",
				InputMapping: map[string]string{
					"reconciliationID": "$.steps.1.result.reconciliation_id",
					"matchedInvoices":  "$.steps.4.result.matched_invoices",
					"availableITC":     "$.steps.3.result.available_itc",
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
			// Step 6: Post Reconciliation Adjustments
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostITCReconciliationAdjustments",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"reconciliationID":   "$.steps.1.result.reconciliation_id",
					"discrepancies":      "$.steps.5.result.discrepancies",
					"journalDate":        "$.input.reconciliation_period",
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
			// Step 7: Generate Reconciliation Report
			{
				StepNumber:    7,
				ServiceName:   "gst-ledger",
				HandlerMethod: "GenerateReconciliationReport",
				InputMapping: map[string]string{
					"reconciliationID":   "$.steps.1.result.reconciliation_id",
					"discrepancies":      "$.steps.5.result.discrepancies",
					"adjustments":        "$.steps.6.result.adjustments",
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
			// Step 8: Finalize ITC Reconciliation
			{
				StepNumber:    8,
				ServiceName:   "gst-ledger",
				HandlerMethod: "FinalizeITCReconciliation",
				InputMapping: map[string]string{
					"reconciliationID":   "$.steps.1.result.reconciliation_id",
					"reconciliationReport": "$.steps.7.result.reconciliation_report",
					"completionDate":     "$.input.reconciliation_period",
				},
				TimeoutSeconds:    20,
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
			// Step 9: Complete ITC Reconciliation
			{
				StepNumber:    9,
				ServiceName:   "gst-ledger",
				HandlerMethod: "CompleteITCReconciliation",
				InputMapping: map[string]string{
					"reconciliationID": "$.steps.1.result.reconciliation_id",
					"finalStatus":      "COMPLETED",
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

			// Step 101: Cancel Reconciliation Initiation (compensates step 1)
			{
				StepNumber:    101,
				ServiceName:   "gst-ledger",
				HandlerMethod: "CancelReconciliationInitiation",
				InputMapping: map[string]string{
					"reconciliationID": "$.steps.1.result.reconciliation_id",
					"reason":           "Saga compensation - ITC reconciliation failed",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 102: Clear GSTR-2 Data Fetch (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "gst-return",
				HandlerMethod: "ClearGSTR2DataFetch",
				InputMapping: map[string]string{
					"reconciliationID": "$.steps.1.result.reconciliation_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: Clear ITC Calculation (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "gst-ledger",
				HandlerMethod: "ClearAvailableITCCalculation",
				InputMapping: map[string]string{
					"reconciliationID": "$.steps.1.result.reconciliation_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 104: Revert Invoice Matching (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "accounts-payable",
				HandlerMethod: "RevertSupplierInvoiceMatching",
				InputMapping: map[string]string{
					"reconciliationID": "$.steps.1.result.reconciliation_id",
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
					"reconciliationID": "$.steps.1.result.reconciliation_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: Reverse Reconciliation Adjustments (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseITCReconciliationAdjustments",
				InputMapping: map[string]string{
					"reconciliationID": "$.steps.1.result.reconciliation_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Delete Reconciliation Report (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "gst-ledger",
				HandlerMethod: "DeleteReconciliationReport",
				InputMapping: map[string]string{
					"reconciliationID": "$.steps.1.result.reconciliation_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 108: Revert ITC Finalization (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "gst-ledger",
				HandlerMethod: "RevertITCFinalization",
				InputMapping: map[string]string{
					"reconciliationID": "$.steps.1.result.reconciliation_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ITCReconciliationSaga) SagaType() string {
	return "SAGA-G02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ITCReconciliationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ITCReconciliationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ITCReconciliationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["reconciliation_period"] == nil {
		return errors.New("reconciliation_period is required (format: YYYY-MM)")
	}

	if inputMap["itc_type"] == nil {
		return errors.New("itc_type is required (PURCHASE_IMPORT, SERVICES, CAPITAL_GOODS, etc.)")
	}

	itcType, ok := inputMap["itc_type"].(string)
	if !ok {
		return errors.New("itc_type must be a string")
	}

	validTypes := map[string]bool{
		"PURCHASE_IMPORT":  true,
		"SERVICES":         true,
		"CAPITAL_GOODS":    true,
		"INELIGIBLE":       true,
		"ALL":              true,
	}

	if !validTypes[itcType] {
		return errors.New("itc_type must be PURCHASE_IMPORT, SERVICES, CAPITAL_GOODS, INELIGIBLE, or ALL")
	}

	if inputMap["gstin"] == nil {
		return errors.New("gstin is required")
	}

	return nil
}
