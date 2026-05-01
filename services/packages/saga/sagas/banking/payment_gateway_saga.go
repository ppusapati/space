// Package banking provides saga handlers for banking module workflows
package banking

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// PaymentGatewaySaga implements SAGA-B05: Payment Gateway Integration & Settlement workflow
// Business Flow: ValidateGatewayTransaction → InitiateGatewayPayment → CheckFraudDetection → ProcessPayment → UpdateTransactionStatus → PostGatewayJournal → SettlePayment → NotifyCustomer → LogGatewayTransaction → ArchiveGatewayData
// Timeout: 120 seconds, Critical steps: 1,2,3,4,7,10
type PaymentGatewaySaga struct {
	steps []*saga.StepDefinition
}

// NewPaymentGatewaySaga creates a new Payment Gateway Integration & Settlement saga handler
func NewPaymentGatewaySaga() saga.SagaHandler {
	return &PaymentGatewaySaga{
		steps: []*saga.StepDefinition{
			// Step 1: Validate Gateway Transaction
			{
				StepNumber:    1,
				ServiceName:   "payment-gateway",
				HandlerMethod: "ValidateGatewayTransaction",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"transactionID":      "$.input.transaction_id",
					"transactionAmount":  "$.input.transaction_amount",
					"gatewayType":        "$.input.gateway_type",
					"customerID":         "$.input.customer_id",
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
			// Step 2: Initiate Gateway Payment
			{
				StepNumber:    2,
				ServiceName:   "payment-gateway",
				HandlerMethod: "InitiateGatewayPayment",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"transactionID":      "$.input.transaction_id",
					"transactionAmount":  "$.input.transaction_amount",
					"gatewayType":        "$.input.gateway_type",
					"customerID":         "$.input.customer_id",
					"validationResult":   "$.steps.1.result.validation_result",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				CompensationSteps: []int32{102},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Check Fraud Detection
			{
				StepNumber:    3,
				ServiceName:   "fraud-detection",
				HandlerMethod: "CheckFraudDetection",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"transactionID":      "$.input.transaction_id",
					"transactionAmount":  "$.input.transaction_amount",
					"gatewayType":        "$.input.gateway_type",
					"customerID":         "$.input.customer_id",
					"gatewayInitiation":  "$.steps.2.result.gateway_initiation",
				},
				TimeoutSeconds: 25,
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
			// Step 4: Process Payment
			{
				StepNumber:    4,
				ServiceName:   "banking",
				HandlerMethod: "ProcessPayment",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"transactionID":      "$.input.transaction_id",
					"transactionAmount":  "$.input.transaction_amount",
					"customerID":         "$.input.customer_id",
					"fraudCheckResult":   "$.steps.3.result.fraud_check_result",
					"gatewayInitiation":  "$.steps.2.result.gateway_initiation",
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
			// Step 5: Update Transaction Status
			{
				StepNumber:    5,
				ServiceName:   "payment-gateway",
				HandlerMethod: "UpdateTransactionStatus",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"transactionID":      "$.input.transaction_id",
					"transactionStatus":  "PROCESSED",
					"paymentProcessing":  "$.steps.4.result.payment_processing",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Post Gateway Journal
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostGatewayJournal",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"transactionID":      "$.input.transaction_id",
					"transactionAmount":  "$.input.transaction_amount",
					"customerID":         "$.input.customer_id",
					"paymentProcessing":  "$.steps.4.result.payment_processing",
					"journalDate":        "$.input.journal_date",
				},
				TimeoutSeconds: 30,
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
			// Step 7: Settle Payment
			{
				StepNumber:    7,
				ServiceName:   "settlement",
				HandlerMethod: "SettlePayment",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"transactionID":      "$.input.transaction_id",
					"transactionAmount":  "$.input.transaction_amount",
					"customerID":         "$.input.customer_id",
					"statusUpdate":       "$.steps.5.result.status_update",
					"journalEntryID":     "$.steps.6.result.journal_entry_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Notify Customer
			{
				StepNumber:    8,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "NotifyCustomer",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"transactionID":      "$.input.transaction_id",
					"transactionAmount":  "$.input.transaction_amount",
					"customerID":         "$.input.customer_id",
					"settlementDetails":  "$.steps.7.result.settlement_details",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Log Gateway Transaction
			{
				StepNumber:    9,
				ServiceName:   "payment-gateway",
				HandlerMethod: "LogGatewayTransaction",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"transactionID":      "$.input.transaction_id",
					"transactionAmount":  "$.input.transaction_amount",
					"customerID":         "$.input.customer_id",
					"transactionStatus":  "$.steps.5.result.transaction_status",
					"notificationDetails": "$.steps.8.result.notification_details",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Archive Gateway Data
			{
				StepNumber:    10,
				ServiceName:   "banking",
				HandlerMethod: "ArchiveGatewayData",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"transactionID":      "$.input.transaction_id",
					"transactionLog":     "$.steps.9.result.transaction_log",
					"archiveDate":        "$.input.archive_date",
					"settlementDetails":  "$.steps.7.result.settlement_details",
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
			// ===== COMPENSATION STEPS =====

			// Step 102: RevokeGatewayPayment (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "payment-gateway",
				HandlerMethod: "RevokeGatewayPayment",
				InputMapping: map[string]string{
					"transactionID":      "$.input.transaction_id",
					"gatewayInitiation":  "$.steps.2.result.gateway_initiation",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: ClearFraudFlag (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "fraud-detection",
				HandlerMethod: "ClearFraudFlag",
				InputMapping: map[string]string{
					"transactionID":      "$.input.transaction_id",
					"fraudCheckResult":   "$.steps.3.result.fraud_check_result",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 104: ReversePaymentProcessing (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "banking",
				HandlerMethod: "ReversePaymentProcessing",
				InputMapping: map[string]string{
					"transactionID":      "$.input.transaction_id",
					"transactionAmount":  "$.input.transaction_amount",
					"paymentProcessing":  "$.steps.4.result.payment_processing",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: RevertTransactionStatus (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "payment-gateway",
				HandlerMethod: "RevertTransactionStatus",
				InputMapping: map[string]string{
					"transactionID":      "$.input.transaction_id",
					"previousStatus":     "INITIATED",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: ReverseGatewayJournal (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseGatewayJournal",
				InputMapping: map[string]string{
					"transactionID":      "$.input.transaction_id",
					"journalEntryID":     "$.steps.6.result.journal_entry_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: ReverseSettlement (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "settlement",
				HandlerMethod: "ReverseSettlement",
				InputMapping: map[string]string{
					"transactionID":      "$.input.transaction_id",
					"settlementDetails":  "$.steps.7.result.settlement_details",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: CancelCustomerNotification (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "CancelCustomerNotification",
				InputMapping: map[string]string{
					"transactionID":      "$.input.transaction_id",
					"notificationDetails": "$.steps.8.result.notification_details",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 109: ClearTransactionLog (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "payment-gateway",
				HandlerMethod: "ClearTransactionLog",
				InputMapping: map[string]string{
					"transactionID":      "$.input.transaction_id",
					"transactionLog":     "$.steps.9.result.transaction_log",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *PaymentGatewaySaga) SagaType() string {
	return "SAGA-B05"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *PaymentGatewaySaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *PaymentGatewaySaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *PaymentGatewaySaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["transaction_id"] == nil {
		return errors.New("transaction_id is required")
	}

	transactionID, ok := inputMap["transaction_id"].(string)
	if !ok || transactionID == "" {
		return errors.New("transaction_id must be a non-empty string")
	}

	if inputMap["transaction_amount"] == nil {
		return errors.New("transaction_amount is required")
	}

	transactionAmount, ok := inputMap["transaction_amount"].(string)
	if !ok || transactionAmount == "" {
		return errors.New("transaction_amount must be a non-empty string")
	}

	if inputMap["gateway_type"] == nil {
		return errors.New("gateway_type is required")
	}

	gatewayType, ok := inputMap["gateway_type"].(string)
	if !ok || gatewayType == "" {
		return errors.New("gateway_type must be a non-empty string")
	}

	validGateways := map[string]bool{"CARD": true, "NETBANKING": true, "WALLET": true}
	if !validGateways[gatewayType] {
		return errors.New("gateway_type must be one of: CARD, NETBANKING, WALLET")
	}

	if inputMap["customer_id"] == nil {
		return errors.New("customer_id is required")
	}

	customerID, ok := inputMap["customer_id"].(string)
	if !ok || customerID == "" {
		return errors.New("customer_id must be a non-empty string")
	}

	return nil
}
