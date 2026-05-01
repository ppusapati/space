// Package finance provides saga handlers for finance module workflows
package finance

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// IntercompanyTransactionSaga implements SAGA-F04: Intercompany Transaction workflow
// Business Flow: RecordIntercompanyPayable → RecordIntercompanyReceivable → PostIntercompanyEntries → ReviewEliminationItems → AllocateIntercompanyCosts → CreateEliminationEntries → PostEliminationEntries → FinalizeIntercompanyTransaction
// Timeout: 90 seconds, Critical steps: 1,2,3,6,8
type IntercompanyTransactionSaga struct {
	steps []*saga.StepDefinition
}

// NewIntercompanyTransactionSaga creates a new Intercompany Transaction saga handler
func NewIntercompanyTransactionSaga() saga.SagaHandler {
	return &IntercompanyTransactionSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Record Intercompany Payable
			{
				StepNumber:    1,
				ServiceName:   "accounts-payable",
				HandlerMethod: "RecordIntercompanyPayable",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"invoiceID":       "$.input.invoice_id",
					"fromCompanyID":   "$.input.from_company_id",
					"toCompanyID":     "$.input.to_company_id",
					"amount":          "$.input.amount",
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
			// Step 2: Record Intercompany Receivable
			{
				StepNumber:    2,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "RecordIntercompanyReceivable",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"invoiceID":       "$.input.invoice_id",
					"fromCompanyID":   "$.input.from_company_id",
					"toCompanyID":     "$.input.to_company_id",
					"amount":          "$.input.amount",
					"payableID":       "$.steps.1.result.payable_id",
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
				CompensationSteps: []int32{102},
			},
			// Step 3: Post Intercompany Entries to GL
			{
				StepNumber:    3,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostIntercompanyEntries",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"invoiceID":       "$.input.invoice_id",
					"fromCompanyID":   "$.input.from_company_id",
					"toCompanyID":     "$.input.to_company_id",
					"amount":          "$.input.amount",
					"payableID":       "$.steps.1.result.payable_id",
					"receivableID":    "$.steps.2.result.receivable_id",
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
				CompensationSteps: []int32{103},
			},
			// Step 4: Review Elimination Items
			{
				StepNumber:    4,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReviewEliminationItems",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"invoiceID":       "$.input.invoice_id",
					"fromCompanyID":   "$.input.from_company_id",
					"toCompanyID":     "$.input.to_company_id",
					"glEntries":       "$.steps.3.result.gl_entries",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{104},
			},
			// Step 5: Allocate Intercompany Costs
			{
				StepNumber:    5,
				ServiceName:   "cost-center",
				HandlerMethod: "AllocateIntercompanyCosts",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"invoiceID":       "$.input.invoice_id",
					"fromCompanyID":   "$.input.from_company_id",
					"toCompanyID":     "$.input.to_company_id",
					"amount":          "$.input.amount",
					"eliminationItems": "$.steps.4.result.elimination_items",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{105},
			},
			// Step 6: Create Elimination Entries
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "CreateEliminationEntries",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"invoiceID":       "$.input.invoice_id",
					"fromCompanyID":   "$.input.from_company_id",
					"toCompanyID":     "$.input.to_company_id",
					"amount":          "$.input.amount",
					"glEntries":       "$.steps.3.result.gl_entries",
					"eliminationItems": "$.steps.4.result.elimination_items",
					"costAllocation":  "$.steps.5.result.cost_allocation",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{106},
			},
			// Step 7: Post Elimination Entries to GL
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostEliminationEntries",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"invoiceID":       "$.input.invoice_id",
					"eliminationEntries": "$.steps.6.result.elimination_entries",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{107},
			},
			// Step 8: Finalize Intercompany Transaction
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "FinalizeIntercompanyTransaction",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"invoiceID":       "$.input.invoice_id",
					"fromCompanyID":   "$.input.from_company_id",
					"toCompanyID":     "$.input.to_company_id",
					"amount":          "$.input.amount",
					"payableID":       "$.steps.1.result.payable_id",
					"receivableID":    "$.steps.2.result.receivable_id",
					"intercompanyEntries": "$.steps.3.result.gl_entries",
					"eliminationPosting": "$.steps.7.result.posting_result",
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
			// ===== COMPENSATION STEPS =====

			// Step 102: ReverseIntercompanyReceivable (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ReverseIntercompanyReceivable",
				InputMapping: map[string]string{
					"invoiceID":   "$.input.invoice_id",
					"receivableID": "$.steps.2.result.receivable_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: ReverseIntercompanyGLEntries (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseIntercompanyGLEntries",
				InputMapping: map[string]string{
					"invoiceID": "$.input.invoice_id",
					"glEntries": "$.steps.3.result.gl_entries",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: ReverseEliminationReview (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseEliminationReview",
				InputMapping: map[string]string{
					"invoiceID":        "$.input.invoice_id",
					"eliminationItems": "$.steps.4.result.elimination_items",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: ReverseIntercompanyCostAllocation (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "cost-center",
				HandlerMethod: "ReverseIntercompanyCostAllocation",
				InputMapping: map[string]string{
					"invoiceID":       "$.input.invoice_id",
					"costAllocation":  "$.steps.5.result.cost_allocation",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: ReverseEliminationEntries (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseEliminationEntries",
				InputMapping: map[string]string{
					"invoiceID":       "$.input.invoice_id",
					"eliminationEntries": "$.steps.6.result.elimination_entries",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 107: ReverseEliminationPosting (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseEliminationPosting",
				InputMapping: map[string]string{
					"invoiceID":       "$.input.invoice_id",
					"postingResult":   "$.steps.7.result.posting_result",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *IntercompanyTransactionSaga) SagaType() string {
	return "SAGA-F04"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *IntercompanyTransactionSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *IntercompanyTransactionSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *IntercompanyTransactionSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["invoice_id"] == nil {
		return errors.New("invoice_id is required")
	}

	invoiceID, ok := inputMap["invoice_id"].(string)
	if !ok || invoiceID == "" {
		return errors.New("invoice_id must be a non-empty string")
	}

	if inputMap["from_company_id"] == nil {
		return errors.New("from_company_id is required")
	}

	fromCompanyID, ok := inputMap["from_company_id"].(string)
	if !ok || fromCompanyID == "" {
		return errors.New("from_company_id must be a non-empty string")
	}

	if inputMap["to_company_id"] == nil {
		return errors.New("to_company_id is required")
	}

	toCompanyID, ok := inputMap["to_company_id"].(string)
	if !ok || toCompanyID == "" {
		return errors.New("to_company_id must be a non-empty string")
	}

	if inputMap["amount"] == nil {
		return errors.New("amount is required")
	}

	amount, ok := inputMap["amount"].(float64)
	if !ok || amount <= 0 {
		return errors.New("amount must be a positive number")
	}

	if fromCompanyID == toCompanyID {
		return errors.New("from_company_id and to_company_id must be different")
	}

	return nil
}
