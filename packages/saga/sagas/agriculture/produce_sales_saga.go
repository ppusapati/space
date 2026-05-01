// Package agriculture provides saga handlers for agricultural workflows
package agriculture

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ProduceSalesSaga implements SAGA-A06: Agricultural Produce Sales & Billing workflow
// Business Flow: InitiateSalesOrder → ValidateProduceAvailability → PerformQualityCheck → CreateSalesInvoice → UpdateInventoryAllocation → ProcessDelivery → ProcessCustomerPayment → UpdateReceivableEntry → PostSalesJournal → GenerateSalesConfirmation → CompleteSalesOrder
// Steps: 10 forward + 9 compensation = 19 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,7,10
type ProduceSalesSaga struct {
	steps []*saga.StepDefinition
}

// NewProduceSalesSaga creates a new Agricultural Produce Sales saga handler
func NewProduceSalesSaga() saga.SagaHandler {
	return &ProduceSalesSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Sales Order
			{
				StepNumber:    1,
				ServiceName:   "agriculture",
				HandlerMethod: "InitiateSalesOrder",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"salesOrderID":   "$.input.sales_order_id",
					"farmID":         "$.input.farm_id",
					"produceType":    "$.input.produce_type",
					"quantity":       "$.input.quantity",
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
			// Step 2: Validate Produce Availability
			{
				StepNumber:    2,
				ServiceName:   "inventory",
				HandlerMethod: "ValidateProduceAvailability",
				InputMapping: map[string]string{
					"salesOrderID":  "$.steps.1.result.sales_order_id",
					"produceType":   "$.input.produce_type",
					"quantity":      "$.input.quantity",
					"validateStock": "true",
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
			// Step 3: Perform Quality Check
			{
				StepNumber:    3,
				ServiceName:   "quality-inspection",
				HandlerMethod: "PerformQualityCheckForSales",
				InputMapping: map[string]string{
					"salesOrderID": "$.steps.1.result.sales_order_id",
					"produceType":  "$.input.produce_type",
					"quantity":     "$.input.quantity",
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
			// Step 4: Create Sales Invoice
			{
				StepNumber:    4,
				ServiceName:   "sales",
				HandlerMethod: "CreateSalesInvoice",
				InputMapping: map[string]string{
					"salesOrderID":   "$.steps.1.result.sales_order_id",
					"produceType":    "$.input.produce_type",
					"quantity":       "$.input.quantity",
					"qualityResult":  "$.steps.3.result.quality_result",
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
			// Step 5: Update Inventory Allocation
			{
				StepNumber:    5,
				ServiceName:   "inventory",
				HandlerMethod: "UpdateInventoryAllocation",
				InputMapping: map[string]string{
					"salesOrderID": "$.steps.1.result.sales_order_id",
					"produceType":  "$.input.produce_type",
					"quantity":     "$.input.quantity",
					"invoiceData":  "$.steps.4.result.invoice_data",
				},
				TimeoutSeconds:    25,
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
			// Step 6: Process Delivery
			{
				StepNumber:    6,
				ServiceName:   "sales",
				HandlerMethod: "ProcessDelivery",
				InputMapping: map[string]string{
					"salesOrderID":    "$.steps.1.result.sales_order_id",
					"inventoryData":   "$.steps.5.result.inventory_data",
					"produceType":     "$.input.produce_type",
				},
				TimeoutSeconds:    35,
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
			// Step 7: Process Customer Payment
			{
				StepNumber:    7,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ProcessCustomerPayment",
				InputMapping: map[string]string{
					"salesOrderID":  "$.steps.1.result.sales_order_id",
					"invoiceData":   "$.steps.4.result.invoice_data",
					"deliveryData":  "$.steps.6.result.delivery_data",
				},
				TimeoutSeconds:    40,
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
			// Step 8: Update Receivable Entry
			{
				StepNumber:    8,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "UpdateReceivableEntry",
				InputMapping: map[string]string{
					"salesOrderID": "$.steps.1.result.sales_order_id",
					"paymentData":  "$.steps.7.result.payment_data",
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
			// Step 9: Post Sales Journal Entries
			{
				StepNumber:    9,
				ServiceName:   "general-ledger",
				HandlerMethod: "ApplyProduceSalesJournal",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"salesOrderID":   "$.steps.1.result.sales_order_id",
					"invoiceData":    "$.steps.4.result.invoice_data",
					"paymentData":    "$.steps.7.result.payment_data",
					"journalDate":    "$.input.sales_date",
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
			// Step 10: Complete Sales Order
			{
				StepNumber:    10,
				ServiceName:   "agriculture",
				HandlerMethod: "CompleteSalesOrder",
				InputMapping: map[string]string{
					"salesOrderID":      "$.steps.1.result.sales_order_id",
					"journalEntries":    "$.steps.9.result.journal_entries",
					"completionStatus":  "Completed",
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

			// Step 101: Revert Produce Availability Check (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "inventory",
				HandlerMethod: "RevertProduceAvailabilityCheck",
				InputMapping: map[string]string{
					"salesOrderID": "$.steps.1.result.sales_order_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Revert Quality Check (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "quality-inspection",
				HandlerMethod: "RevertQualityCheckForSales",
				InputMapping: map[string]string{
					"salesOrderID": "$.steps.1.result.sales_order_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Cancel Sales Invoice (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "sales",
				HandlerMethod: "CancelSalesInvoice",
				InputMapping: map[string]string{
					"salesOrderID": "$.steps.1.result.sales_order_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Revert Inventory Allocation (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "inventory",
				HandlerMethod: "RevertInventoryAllocation",
				InputMapping: map[string]string{
					"salesOrderID": "$.steps.1.result.sales_order_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Revert Delivery Process (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "sales",
				HandlerMethod: "RevertDeliveryProcess",
				InputMapping: map[string]string{
					"salesOrderID": "$.steps.1.result.sales_order_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 106: Revert Customer Payment (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "RevertCustomerPayment",
				InputMapping: map[string]string{
					"salesOrderID": "$.steps.1.result.sales_order_id",
				},
				TimeoutSeconds: 40,
				IsCritical:     false,
			},
			// Step 107: Revert Receivable Entry (compensates step 8)
			{
				StepNumber:    107,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "RevertReceivableEntry",
				InputMapping: map[string]string{
					"salesOrderID": "$.steps.1.result.sales_order_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 108: Reverse Sales Journal (compensates step 9)
			{
				StepNumber:    108,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseProduceSalesJournal",
				InputMapping: map[string]string{
					"salesOrderID": "$.steps.1.result.sales_order_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ProduceSalesSaga) SagaType() string {
	return "SAGA-A06"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ProduceSalesSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ProduceSalesSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ProduceSalesSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["sales_order_id"] == nil {
		return errors.New("sales_order_id is required")
	}

	if inputMap["farm_id"] == nil {
		return errors.New("farm_id is required")
	}

	if inputMap["produce_type"] == nil {
		return errors.New("produce_type is required")
	}

	if inputMap["quantity"] == nil {
		return errors.New("quantity is required")
	}

	return nil
}
