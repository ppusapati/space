// Package retail provides saga handlers for retail workflows
package retail

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// POSTransactionSaga implements SAGA-R01: Point-of-Sale (POS) Transaction workflow
// Business Flow: InitiatePOSTransaction → ValidatePOSTerminal → LoadCustomerProfile → CalculatePricing → ValidatePaymentMethod → ProcessPayment → UpdateInventoryAllocation → RecordSalesTransaction → ApplyLoyaltyPoints → GenerateSalesJournal → UpdateRevenueLedger → RecordPaymentJournal → ClosePOSTransaction
// Steps: 12 forward + 11 compensation = 23 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,7,8,12
type POSTransactionSaga struct {
	steps []*saga.StepDefinition
}

// NewPOSTransactionSaga creates a new POS Transaction saga handler
func NewPOSTransactionSaga() saga.SagaHandler {
	return &POSTransactionSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate POS Transaction
			{
				StepNumber:    1,
				ServiceName:   "pos",
				HandlerMethod: "InitiatePOSTransaction",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"transactionID": "$.input.transaction_id",
					"terminalID":    "$.input.terminal_id",
					"operatorID":    "$.input.operator_id",
					"transactionTime": "$.input.transaction_time",
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
			// Step 2: Validate POS Terminal
			{
				StepNumber:    2,
				ServiceName:   "pos",
				HandlerMethod: "ValidatePOSTerminal",
				InputMapping: map[string]string{
					"transactionID": "$.steps.1.result.transaction_id",
					"terminalID":    "$.input.terminal_id",
					"validateStatus": "active",
				},
				TimeoutSeconds:    20,
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
			// Step 3: Load Customer Profile
			{
				StepNumber:    3,
				ServiceName:   "customer",
				HandlerMethod: "LoadCustomerProfile",
				InputMapping: map[string]string{
					"transactionID": "$.steps.1.result.transaction_id",
					"customerID":    "$.input.customer_id",
				},
				TimeoutSeconds:    25,
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
			// Step 4: Calculate Pricing with Discounts
			{
				StepNumber:    4,
				ServiceName:   "pricing",
				HandlerMethod: "CalculatePricingWithDiscounts",
				InputMapping: map[string]string{
					"transactionID":  "$.steps.1.result.transaction_id",
					"customerProfile": "$.steps.3.result.customer_profile",
					"itemDetails":    "$.input.item_details",
					"totalAmount":    "$.input.total_amount",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{101},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Validate Payment Method
			{
				StepNumber:    5,
				ServiceName:   "payment-gateway",
				HandlerMethod: "ValidatePaymentMethod",
				InputMapping: map[string]string{
					"transactionID":   "$.steps.1.result.transaction_id",
					"paymentMethod":   "$.input.payment_method",
					"finalAmount":     "$.steps.4.result.final_amount",
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
			// Step 6: Process Payment
			{
				StepNumber:    6,
				ServiceName:   "payment-gateway",
				HandlerMethod: "ProcessPayment",
				InputMapping: map[string]string{
					"transactionID":   "$.steps.1.result.transaction_id",
					"paymentMethod":   "$.input.payment_method",
					"finalAmount":     "$.steps.4.result.final_amount",
					"paymentRef":      "$.input.payment_reference",
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
			// Step 7: Update Inventory Allocation
			{
				StepNumber:    7,
				ServiceName:   "inventory",
				HandlerMethod: "UpdateInventoryAllocation",
				InputMapping: map[string]string{
					"transactionID": "$.steps.1.result.transaction_id",
					"itemDetails":   "$.input.item_details",
					"warehouseID":   "$.input.warehouse_id",
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
			// Step 8: Record Sales Transaction
			{
				StepNumber:    8,
				ServiceName:   "sales",
				HandlerMethod: "RecordSalesTransaction",
				InputMapping: map[string]string{
					"transactionID":   "$.steps.1.result.transaction_id",
					"customerID":      "$.input.customer_id",
					"itemDetails":     "$.input.item_details",
					"finalAmount":     "$.steps.4.result.final_amount",
					"paymentStatus":   "$.steps.6.result.payment_status",
					"terminalID":      "$.input.terminal_id",
				},
				TimeoutSeconds:    35,
				IsCritical:        true,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Apply Loyalty Points
			{
				StepNumber:    9,
				ServiceName:   "loyalty",
				HandlerMethod: "ApplyLoyaltyPoints",
				InputMapping: map[string]string{
					"transactionID": "$.steps.1.result.transaction_id",
					"customerID":    "$.input.customer_id",
					"purchaseAmount": "$.steps.4.result.final_amount",
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
			// Step 10: Generate Sales Journal
			{
				StepNumber:    10,
				ServiceName:   "general-ledger",
				HandlerMethod: "GenerateSalesJournal",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"transactionID": "$.steps.1.result.transaction_id",
					"salesAmount":   "$.steps.4.result.final_amount",
					"journalDate":   "$.input.transaction_time",
				},
				TimeoutSeconds:    25,
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
			// Step 11: Update Revenue Ledger
			{
				StepNumber:    11,
				ServiceName:   "general-ledger",
				HandlerMethod: "UpdateRevenueLedger",
				InputMapping: map[string]string{
					"transactionID": "$.steps.1.result.transaction_id",
					"salesAmount":   "$.steps.4.result.final_amount",
					"terminalID":    "$.input.terminal_id",
				},
				TimeoutSeconds:    20,
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
			// Step 12: Record Payment Journal
			{
				StepNumber:    12,
				ServiceName:   "general-ledger",
				HandlerMethod: "RecordPaymentJournal",
				InputMapping: map[string]string{
					"transactionID": "$.steps.1.result.transaction_id",
					"paymentAmount": "$.steps.4.result.final_amount",
					"paymentMethod": "$.input.payment_method",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{110},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 13: Close POS Transaction
			{
				StepNumber:    13,
				ServiceName:   "pos",
				HandlerMethod: "ClosePOSTransaction",
				InputMapping: map[string]string{
					"transactionID":   "$.steps.1.result.transaction_id",
					"terminalID":      "$.input.terminal_id",
					"journalEntries":  "$.steps.10.result.journal_entries",
					"paymentJournal":  "$.steps.12.result.payment_journal",
					"transactionStatus": "Completed",
				},
				TimeoutSeconds:    20,
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

			// Step 101: Revert Pricing Calculation (compensates step 4)
			{
				StepNumber:    101,
				ServiceName:   "pricing",
				HandlerMethod: "RevertPricingCalculation",
				InputMapping: map[string]string{
					"transactionID": "$.steps.1.result.transaction_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 102: Revert Payment Method Validation (compensates step 5)
			{
				StepNumber:    102,
				ServiceName:   "payment-gateway",
				HandlerMethod: "RevertPaymentMethodValidation",
				InputMapping: map[string]string{
					"transactionID": "$.steps.1.result.transaction_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: Reverse Payment (compensates step 6)
			{
				StepNumber:    103,
				ServiceName:   "payment-gateway",
				HandlerMethod: "ReversePayment",
				InputMapping: map[string]string{
					"transactionID": "$.steps.1.result.transaction_id",
					"paymentAmount": "$.steps.4.result.final_amount",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Revert Inventory Allocation (compensates step 7)
			{
				StepNumber:    104,
				ServiceName:   "inventory",
				HandlerMethod: "RevertInventoryAllocation",
				InputMapping: map[string]string{
					"transactionID": "$.steps.1.result.transaction_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: Reverse Sales Transaction (compensates step 8)
			{
				StepNumber:    105,
				ServiceName:   "sales",
				HandlerMethod: "ReverseSalesTransaction",
				InputMapping: map[string]string{
					"transactionID": "$.steps.1.result.transaction_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 106: Revert Loyalty Points (compensates step 9)
			{
				StepNumber:    106,
				ServiceName:   "loyalty",
				HandlerMethod: "RevertLoyaltyPoints",
				InputMapping: map[string]string{
					"transactionID": "$.steps.1.result.transaction_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: Reverse Sales Journal (compensates step 10)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseSalesJournal",
				InputMapping: map[string]string{
					"transactionID": "$.steps.1.result.transaction_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 108: Revert Revenue Ledger Update (compensates step 11)
			{
				StepNumber:    108,
				ServiceName:   "general-ledger",
				HandlerMethod: "RevertRevenueLedgerUpdate",
				InputMapping: map[string]string{
					"transactionID": "$.steps.1.result.transaction_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 109: Revert Revenue Ledger Update (compensates step 11)
			{
				StepNumber:    109,
				ServiceName:   "general-ledger",
				HandlerMethod: "RevertRevenueLedgerUpdate",
				InputMapping: map[string]string{
					"transactionID": "$.steps.1.result.transaction_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 110: Reverse Payment Journal (compensates step 12)
			{
				StepNumber:    110,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReversePaymentJournal",
				InputMapping: map[string]string{
					"transactionID": "$.steps.1.result.transaction_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *POSTransactionSaga) SagaType() string {
	return "SAGA-R01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *POSTransactionSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *POSTransactionSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *POSTransactionSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["transaction_id"] == nil {
		return errors.New("transaction_id is required")
	}

	if inputMap["terminal_id"] == nil {
		return errors.New("terminal_id is required")
	}

	if inputMap["customer_id"] == nil {
		return errors.New("customer_id is required")
	}

	if inputMap["total_amount"] == nil {
		return errors.New("total_amount is required")
	}

	if inputMap["payment_method"] == nil {
		return errors.New("payment_method is required")
	}

	return nil
}
