// Package finance provides saga handlers for finance module workflows
package finance

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// GSTCreditReversalSaga implements SAGA-F07: GST Input Credit Reversal workflow
// Business Flow: IdentifyIneligibleCredit → CalculateReversalAmount → PostITCReversal → UpdateGSTR2 → UpdateGSTR3B → RecordCompliance → CompleteReversal
// GST Compliance: Input Tax Credit (ITC) reversal as per CGST Rules 42 & 43
type GSTCreditReversalSaga struct {
	steps []*saga.StepDefinition
}

// NewGSTCreditReversalSaga creates a new GST Input Credit Reversal saga handler
func NewGSTCreditReversalSaga() saga.SagaHandler {
	return &GSTCreditReversalSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Identify Ineligible Credit
			{
				StepNumber:    1,
				ServiceName:   "gst",
				HandlerMethod: "IdentifyIneligibleCredit",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"invoiceID":       "$.input.invoice_id",
					"reversalReason":  "$.input.reversal_reason",
					"reversalType":    "$.input.reversal_type",
					"financialPeriod": "$.input.financial_period",
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
			// Step 2: Calculate Reversal Amount (Rule 42/43 computation)
			{
				StepNumber:    2,
				ServiceName:   "gst",
				HandlerMethod: "CalculateReversalAmount",
				InputMapping: map[string]string{
					"creditID":        "$.steps.1.result.credit_id",
					"invoiceID":       "$.input.invoice_id",
					"reversalReason":  "$.input.reversal_reason",
					"reversalType":    "$.input.reversal_type",
					"exemptSupplies":  "$.input.exempt_supplies",
					"totalTurnover":   "$.input.total_turnover",
					"cgstAmount":      "$.input.cgst_amount",
					"sgstAmount":      "$.input.sgst_amount",
					"igstAmount":      "$.input.igst_amount",
				},
				TimeoutSeconds:    25,
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
			// Step 3: Post ITC Reversal Journal Entry
			{
				StepNumber:    3,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostITCReversalJournal",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"creditID":        "$.steps.1.result.credit_id",
					"reversalAmount":  "$.steps.2.result.reversal_amount",
					"cgstReversal":    "$.steps.2.result.cgst_reversal",
					"sgstReversal":    "$.steps.2.result.sgst_reversal",
					"igstReversal":    "$.steps.2.result.igst_reversal",
					"journalDate":     "$.input.reversal_date",
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
			// Step 4: Update GSTR-2 Record
			{
				StepNumber:    4,
				ServiceName:   "gst",
				HandlerMethod: "UpdateGSTR2Record",
				InputMapping: map[string]string{
					"creditID":        "$.steps.1.result.credit_id",
					"invoiceID":       "$.input.invoice_id",
					"reversalAmount":  "$.steps.2.result.reversal_amount",
					"financialPeriod": "$.input.financial_period",
					"gstin":           "$.input.gstin",
				},
				TimeoutSeconds:    20,
				IsCritical:        false,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Update GSTR-3B Return
			{
				StepNumber:    5,
				ServiceName:   "gst",
				HandlerMethod: "UpdateGSTR3BReturn",
				InputMapping: map[string]string{
					"creditID":        "$.steps.1.result.credit_id",
					"reversalAmount":  "$.steps.2.result.reversal_amount",
					"cgstReversal":    "$.steps.2.result.cgst_reversal",
					"sgstReversal":    "$.steps.2.result.sgst_reversal",
					"igstReversal":    "$.steps.2.result.igst_reversal",
					"financialPeriod": "$.input.financial_period",
					"gstin":           "$.input.gstin",
					"returnSection":   "$.input.return_section",
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
			// Step 6: Record Compliance Documentation
			{
				StepNumber:    6,
				ServiceName:   "compliance-postings",
				HandlerMethod: "RecordGSTComplianceEvent",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"creditID":        "$.steps.1.result.credit_id",
					"eventType":       "ITC_REVERSAL",
					"reversalReason":  "$.input.reversal_reason",
					"reversalAmount":  "$.steps.2.result.reversal_amount",
					"financialPeriod": "$.input.financial_period",
					"complianceDate":  "$.input.reversal_date",
				},
				TimeoutSeconds:    20,
				IsCritical:        false,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Complete ITC Reversal
			{
				StepNumber:    7,
				ServiceName:   "gst",
				HandlerMethod: "CompleteITCReversal",
				InputMapping: map[string]string{
					"creditID":        "$.steps.1.result.credit_id",
					"reversalAmount":  "$.steps.2.result.reversal_amount",
					"completionDate":  "$.input.reversal_date",
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

			// Step 101: Cancel Credit Identification (compensates step 1)
			{
				StepNumber:    101,
				ServiceName:   "gst",
				HandlerMethod: "CancelCreditIdentification",
				InputMapping: map[string]string{
					"creditID": "$.steps.1.result.credit_id",
					"reason":   "Saga compensation - ITC reversal failed",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 102: Clear Reversal Calculation (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "gst",
				HandlerMethod: "ClearReversalCalculation",
				InputMapping: map[string]string{
					"creditID": "$.steps.1.result.credit_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: Reverse ITC Reversal Journal (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseITCReversalJournal",
				InputMapping: map[string]string{
					"creditID": "$.steps.1.result.credit_id",
					"journalDate": "$.input.reversal_date",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Revert GSTR-2 Update (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "gst",
				HandlerMethod: "RevertGSTR2Update",
				InputMapping: map[string]string{
					"creditID": "$.steps.1.result.credit_id",
					"invoiceID": "$.input.invoice_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 105: Revert GSTR-3B Update (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "gst",
				HandlerMethod: "RevertGSTR3BUpdate",
				InputMapping: map[string]string{
					"creditID": "$.steps.1.result.credit_id",
					"financialPeriod": "$.input.financial_period",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Delete Compliance Record (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "compliance-postings",
				HandlerMethod: "DeleteComplianceRecord",
				InputMapping: map[string]string{
					"creditID": "$.steps.1.result.credit_id",
					"eventType": "ITC_REVERSAL",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *GSTCreditReversalSaga) SagaType() string {
	return "SAGA-F07"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *GSTCreditReversalSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *GSTCreditReversalSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *GSTCreditReversalSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["invoice_id"] == nil {
		return errors.New("invoice_id is required")
	}

	if inputMap["reversal_reason"] == nil {
		return errors.New("reversal_reason is required")
	}

	if inputMap["reversal_type"] == nil {
		return errors.New("reversal_type is required (e.g., RULE_42, RULE_43)")
	}

	if inputMap["financial_period"] == nil {
		return errors.New("financial_period is required")
	}

	if inputMap["gstin"] == nil {
		return errors.New("gstin is required")
	}

	if inputMap["reversal_date"] == nil {
		return errors.New("reversal_date is required")
	}

	// Validate at least one GST component is present
	cgst := inputMap["cgst_amount"]
	sgst := inputMap["sgst_amount"]
	igst := inputMap["igst_amount"]

	if cgst == nil && sgst == nil && igst == nil {
		return errors.New("at least one GST component (cgst_amount, sgst_amount, or igst_amount) is required")
	}

	// For Rule 42/43 reversal, exempt supplies and total turnover are required
	reversalType, ok := inputMap["reversal_type"].(string)
	if ok && (reversalType == "RULE_42" || reversalType == "RULE_43") {
		if inputMap["exempt_supplies"] == nil {
			return errors.New("exempt_supplies is required for Rule 42/43 reversal")
		}
		if inputMap["total_turnover"] == nil {
			return errors.New("total_turnover is required for Rule 42/43 reversal")
		}
	}

	return nil
}
