// Package banking provides saga handlers for banking module workflows
package banking

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ComplianceMonitoringSaga implements SAGA-B06: Banking Compliance & Transaction Monitoring workflow
// Business Flow: CollectTransactionData → IdentifyRiskTransactions → CheckAMLThreshold → CheckCTRThreshold → GenerateComplianceReport → NotifyRegulatoryBody → PostComplianceEntry → UpdateComplianceStatus → ArchiveComplianceData
// Timeout: 120 seconds, Critical steps: 1,2,3,4,6,9
type ComplianceMonitoringSaga struct {
	steps []*saga.StepDefinition
}

// NewComplianceMonitoringSaga creates a new Banking Compliance & Transaction Monitoring saga handler
func NewComplianceMonitoringSaga() saga.SagaHandler {
	return &ComplianceMonitoringSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Collect Transaction Data
			{
				StepNumber:    1,
				ServiceName:   "banking",
				HandlerMethod: "CollectTransactionData",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"monitoringPeriod":      "$.input.monitoring_period",
					"amlThreshold":          "$.input.aml_threshold",
					"ctrThreshold":          "$.input.ctr_threshold",
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
			// Step 2: Identify Risk Transactions
			{
				StepNumber:    2,
				ServiceName:   "compliance",
				HandlerMethod: "IdentifyRiskTransactions",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"monitoringPeriod":      "$.input.monitoring_period",
					"amlThreshold":          "$.input.aml_threshold",
					"ctrThreshold":          "$.input.ctr_threshold",
					"transactionData":       "$.steps.1.result.transaction_data",
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
			// Step 3: Check AML Threshold
			{
				StepNumber:    3,
				ServiceName:   "fraud-detection",
				HandlerMethod: "CheckAMLThreshold",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"monitoringPeriod":      "$.input.monitoring_period",
					"amlThreshold":          "$.input.aml_threshold",
					"riskTransactions":      "$.steps.2.result.risk_transactions",
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
			// Step 4: Check CTR Threshold
			{
				StepNumber:    4,
				ServiceName:   "regulatory-reporting",
				HandlerMethod: "CheckCTRThreshold",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"monitoringPeriod":      "$.input.monitoring_period",
					"ctrThreshold":          "$.input.ctr_threshold",
					"riskTransactions":      "$.steps.2.result.risk_transactions",
					"amlResult":             "$.steps.3.result.aml_result",
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
			// Step 5: Generate Compliance Report
			{
				StepNumber:    5,
				ServiceName:   "compliance",
				HandlerMethod: "GenerateComplianceReport",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"monitoringPeriod":      "$.input.monitoring_period",
					"amlResult":             "$.steps.3.result.aml_result",
					"ctrResult":             "$.steps.4.result.ctr_result",
				},
				TimeoutSeconds: 25,
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
			// Step 6: Notify Regulatory Body
			{
				StepNumber:    6,
				ServiceName:   "regulatory-reporting",
				HandlerMethod: "NotifyRegulatoryBody",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"monitoringPeriod":      "$.input.monitoring_period",
					"complianceReport":      "$.steps.5.result.compliance_report",
					"amlResult":             "$.steps.3.result.aml_result",
					"ctrResult":             "$.steps.4.result.ctr_result",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Post Compliance Entry
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostComplianceEntry",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"monitoringPeriod":      "$.input.monitoring_period",
					"complianceReport":      "$.steps.5.result.compliance_report",
					"notificationDetails":   "$.steps.6.result.notification_details",
					"journalDate":           "$.input.monitoring_period",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Update Compliance Status
			{
				StepNumber:    8,
				ServiceName:   "compliance",
				HandlerMethod: "UpdateComplianceStatus",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"monitoringPeriod":      "$.input.monitoring_period",
					"complianceStatus":      "PROCESSED",
					"journalEntryID":        "$.steps.7.result.journal_entry_id",
				},
				TimeoutSeconds: 25,
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
			// Step 9: Archive Compliance Data
			{
				StepNumber:    9,
				ServiceName:   "audit",
				HandlerMethod: "ArchiveComplianceData",
				InputMapping: map[string]string{
					"tenantID":              "$.tenantID",
					"companyID":             "$.companyID",
					"branchID":              "$.branchID",
					"monitoringPeriod":      "$.input.monitoring_period",
					"complianceReport":      "$.steps.5.result.compliance_report",
					"transactionData":       "$.steps.1.result.transaction_data",
					"statusUpdate":          "$.steps.8.result.status_update",
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

			// Step 102: ClearRiskIdentification (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "compliance",
				HandlerMethod: "ClearRiskIdentification",
				InputMapping: map[string]string{
					"monitoringPeriod":      "$.input.monitoring_period",
					"riskTransactions":      "$.steps.2.result.risk_transactions",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: ClearAMLCheck (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "fraud-detection",
				HandlerMethod: "ClearAMLCheck",
				InputMapping: map[string]string{
					"monitoringPeriod":      "$.input.monitoring_period",
					"amlThreshold":          "$.input.aml_threshold",
					"amlResult":             "$.steps.3.result.aml_result",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 104: ClearCTRCheck (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "regulatory-reporting",
				HandlerMethod: "ClearCTRCheck",
				InputMapping: map[string]string{
					"monitoringPeriod":      "$.input.monitoring_period",
					"ctrThreshold":          "$.input.ctr_threshold",
					"ctrResult":             "$.steps.4.result.ctr_result",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: DeleteComplianceReport (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "compliance",
				HandlerMethod: "DeleteComplianceReport",
				InputMapping: map[string]string{
					"monitoringPeriod":      "$.input.monitoring_period",
					"complianceReport":      "$.steps.5.result.compliance_report",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: ReverseRegulatoryNotification (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "regulatory-reporting",
				HandlerMethod: "ReverseRegulatoryNotification",
				InputMapping: map[string]string{
					"monitoringPeriod":      "$.input.monitoring_period",
					"notificationDetails":   "$.steps.6.result.notification_details",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: ReverseComplianceEntry (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseComplianceEntry",
				InputMapping: map[string]string{
					"monitoringPeriod":      "$.input.monitoring_period",
					"journalEntryID":        "$.steps.7.result.journal_entry_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: RevertComplianceStatus (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "compliance",
				HandlerMethod: "RevertComplianceStatus",
				InputMapping: map[string]string{
					"monitoringPeriod":      "$.input.monitoring_period",
					"previousStatus":        "PENDING",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ComplianceMonitoringSaga) SagaType() string {
	return "SAGA-B06"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ComplianceMonitoringSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ComplianceMonitoringSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ComplianceMonitoringSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["monitoring_period"] == nil {
		return errors.New("monitoring_period is required")
	}

	monitoringPeriod, ok := inputMap["monitoring_period"].(string)
	if !ok || monitoringPeriod == "" {
		return errors.New("monitoring_period must be a non-empty string")
	}

	if inputMap["aml_threshold"] == nil {
		return errors.New("aml_threshold is required")
	}

	amlThreshold, ok := inputMap["aml_threshold"].(string)
	if !ok || amlThreshold == "" {
		return errors.New("aml_threshold must be a non-empty string")
	}

	if inputMap["ctr_threshold"] == nil {
		return errors.New("ctr_threshold is required")
	}

	ctrThreshold, ok := inputMap["ctr_threshold"].(string)
	if !ok || ctrThreshold == "" {
		return errors.New("ctr_threshold must be a non-empty string")
	}

	return nil
}
