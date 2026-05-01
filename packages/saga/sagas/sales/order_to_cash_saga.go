// Package sales provides saga handlers for sales module workflows
package sales

import (
	"errors"
	"p9e.in/samavaya/packages/saga"
)

// OrderToCashSaga implements the Order-to-Cash workflow
// Business flow: Create Order → Reserve Stock → Confirm Order → Create Invoice →
// Generate E-Invoice → Create AR Entry → Post GL → Send Notification
//
// Compensation: If any critical step fails, automatically reverses previous steps
// in reverse order (Reverse GL, Reverse AR, Cancel Invoice, Revert Confirmation, Release Stock, Cancel Order)
type OrderToCashSaga struct {
	steps []*saga.StepDefinition
}

// NewOrderToCashSaga creates a new Order-to-Cash saga handler
func NewOrderToCashSaga() saga.SagaHandler {
	return &OrderToCashSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Order (sales-order service)
			// Creates a new sales order in DRAFT status
			{
				StepNumber:    1,
				ServiceName:   "sales-order",
				HandlerMethod: "CreateOrder",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"companyID":   "$.companyID",
					"branchID":    "$.branchID",
					"customerID":  "$.input.customer_id",
					"items":       "$.input.items",
					"totalAmount": "$.input.total_amount",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:              3,
					InitialBackoffMs:        1000,
					MaxBackoffMs:            30000,
					BackoffMultiplier:       2.0,
					JitterFraction:          0.1,
					CircuitBreakerThreshold: 5,
					CircuitBreakerResetMs:   60000,
				},
				CompensationSteps: []int32{}, // Compensated last
			},

			// Step 2: Reserve Stock (inventory-core service)
			// Checks availability and reserves stock for order items
			{
				StepNumber:    2,
				ServiceName:   "inventory-core",
				HandlerMethod: "ReserveStock",
				InputMapping: map[string]string{
					"orderID": "$.steps.1.result.order_id",
					"items":   "$.input.items",
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
				CompensationSteps: []int32{102}, // ReleaseReservation
			},

			// Step 3: Confirm Order (sales-order service)
			// Confirms order and locks it from editing (status: CONFIRMED)
			{
				StepNumber:    3,
				ServiceName:   "sales-order",
				HandlerMethod: "ConfirmOrder",
				InputMapping: map[string]string{
					"orderID": "$.steps.1.result.order_id",
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
				CompensationSteps: []int32{103}, // RevertConfirmation
			},

			// Step 4: Create Invoice (sales-invoice service)
			// Generates invoice from confirmed order
			{
				StepNumber:    4,
				ServiceName:   "sales-invoice",
				HandlerMethod: "CreateInvoice",
				InputMapping: map[string]string{
					"orderID":  "$.steps.1.result.order_id",
					"tenantID": "$.tenantID",
					"companyID": "$.companyID",
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
				CompensationSteps: []int32{104}, // CancelInvoice
			},

			// Step 5: Generate E-Invoice IRN (e-invoice service)
			// Calls GSTN API to generate IRN and QR code
			// This is an external API call, so more retries allowed
			{
				StepNumber:    5,
				ServiceName:   "e-invoice",
				HandlerMethod: "GenerateIRN",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.4.result.invoice_id",
					"tenantID":  "$.tenantID",
				},
				TimeoutSeconds: 60,
				IsCritical:     false, // Non-critical: invoice can proceed without IRN initially
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        5, // More retries for external API
					InitialBackoffMs:  2000,
					MaxBackoffMs:      120000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{105}, // MarkPendingIRN (partial compensation)
			},

			// Step 6: Create AR Entry (accounts-receivable service)
			// Creates account receivable entry for customer invoice
			{
				StepNumber:    6,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "CreateAREntry",
				InputMapping: map[string]string{
					"invoiceID":   "$.steps.4.result.invoice_id",
					"customerID":  "$.input.customer_id",
					"amount":      "$.input.total_amount",
					"dueDate":     "$.input.due_date",
					"tenantID":    "$.tenantID",
					"companyID":   "$.companyID",
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
				CompensationSteps: []int32{106}, // ReverseAREntry
			},

			// Step 7: Post GL Journal (general-ledger service)
			// Posts accounting entries to general ledger
			// Dr: Accounts Receivable, Cr: Sales Revenue, Cr: GST Output
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostJournal",
				InputMapping: map[string]string{
					"invoiceID":  "$.steps.4.result.invoice_id",
					"amount":     "$.input.total_amount",
					"tenantID":   "$.tenantID",
					"companyID":  "$.companyID",
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
				CompensationSteps: []int32{107}, // ReverseJournal
			},

			// Step 8: Send Notification (notification service)
			// Sends invoice to customer via email/SMS
			{
				StepNumber:    8,
				ServiceName:   "notification",
				HandlerMethod: "SendInvoice",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.4.result.invoice_id",
					"customerID": "$.input.customer_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false, // Non-critical: order complete even if notification fails
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{}, // No compensation for notification
			},

			// Compensation Step 102: Release Stock Reservation
			// Compensates Step 2: Releases reserved stock back to available inventory
			{
				StepNumber:    102,
				ServiceName:   "inventory-core",
				HandlerMethod: "ReleaseReservation",
				InputMapping: map[string]string{
					"reservationID": "$.steps.2.result.reservation_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false, // Non-critical compensation
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},

			// Compensation Step 103: Revert Order Confirmation
			// Compensates Step 3: Reverts order from CONFIRMED back to DRAFT
			{
				StepNumber:    103,
				ServiceName:   "sales-order",
				HandlerMethod: "RevertConfirmation",
				InputMapping: map[string]string{
					"orderID": "$.steps.1.result.order_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},

			// Compensation Step 104: Cancel Invoice
			// Compensates Step 4: Cancels the generated invoice
			{
				StepNumber:    104,
				ServiceName:   "sales-invoice",
				HandlerMethod: "CancelInvoice",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.4.result.invoice_id",
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
				CompensationSteps: []int32{},
			},

			// Compensation Step 105: Mark E-Invoice as Pending
			// Compensates Step 5: Marks e-invoice as pending (partial compensation)
			// Note: Cannot delete IRN from GSTN, so we mark as PENDING for manual follow-up
			{
				StepNumber:    105,
				ServiceName:   "e-invoice",
				HandlerMethod: "MarkPendingIRN",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.4.result.invoice_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},

			// Compensation Step 106: Reverse AR Entry
			// Compensates Step 6: Reverses the AR entry by creating negative transaction
			{
				StepNumber:    106,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ReverseAREntry",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.4.result.invoice_id",
					"amount":    "$.input.total_amount",
					"tenantID":  "$.tenantID",
					"companyID": "$.companyID",
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
				CompensationSteps: []int32{},
			},

			// Compensation Step 107: Reverse GL Journal
			// Compensates Step 7: Creates reverse journal entry
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseJournal",
				InputMapping: map[string]string{
					"invoiceID": "$.steps.4.result.invoice_id",
					"amount":    "$.input.total_amount",
					"tenantID":  "$.tenantID",
					"companyID": "$.companyID",
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
				CompensationSteps: []int32{},
			},

			// Compensation Step 101: Cancel Order
			// Compensates Step 1: Cancels the order
			// This is the final compensation step, executed last
			{
				StepNumber:    101,
				ServiceName:   "sales-order",
				HandlerMethod: "CancelOrder",
				InputMapping: map[string]string{
					"orderID": "$.steps.1.result.order_id",
					"reason":  "Saga compensation - subsequent steps failed",
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
				CompensationSteps: []int32{},
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *OrderToCashSaga) SagaType() string {
	return "SAGA-S01"
}

// GetStepDefinitions returns all steps (forward + compensation)
func (s *OrderToCashSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns definition for a specific step number
func (s *OrderToCashSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the input for saga execution
// Required fields: customer_id, items, total_amount
func (s *OrderToCashSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type: expected map[string]interface{}")
	}

	// Validate customer_id
	if inputMap["customer_id"] == nil || inputMap["customer_id"] == "" {
		return errors.New("customer_id is required for Order-to-Cash saga")
	}

	// Validate items
	if inputMap["items"] == nil {
		return errors.New("items are required for Order-to-Cash saga")
	}

	itemsList, ok := inputMap["items"].([]interface{})
	if !ok || len(itemsList) == 0 {
		return errors.New("items must be a non-empty list")
	}

	// Validate total_amount
	if inputMap["total_amount"] == nil {
		return errors.New("total_amount is required for Order-to-Cash saga")
	}

	return nil
}
