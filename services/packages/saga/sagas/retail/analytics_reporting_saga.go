// Package retail provides saga handlers for retail workflows
package retail

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// AnalyticsReportingSaga implements SAGA-R08: Retail Sales Analytics & Reporting workflow
// Business Flow: InitiateReportGeneration → ValidateReportParameters → AggregateTransactionData → CalculateSalesMetrics → ComputeInventoryMetrics → GenerateAnalyticsData → ApplyReportingJournal → PublishReport → CompleteReporting
// Steps: 9 forward + 8 compensation = 17 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,6,9
type AnalyticsReportingSaga struct {
	steps []*saga.StepDefinition
}

// NewAnalyticsReportingSaga creates a new Retail Sales Analytics saga handler
func NewAnalyticsReportingSaga() saga.SagaHandler {
	return &AnalyticsReportingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Report Generation
			{
				StepNumber:    1,
				ServiceName:   "analytics",
				HandlerMethod: "InitiateReportGeneration",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"reportID":      "$.input.report_id",
					"reportType":    "$.input.report_type",
					"reportPeriod":  "$.input.report_period",
					"generationTime": "$.input.generation_time",
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
			// Step 2: Validate Report Parameters
			{
				StepNumber:    2,
				ServiceName:   "analytics",
				HandlerMethod: "ValidateReportParameters",
				InputMapping: map[string]string{
					"reportID":     "$.steps.1.result.report_id",
					"reportType":   "$.input.report_type",
					"reportPeriod": "$.input.report_period",
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
			// Step 3: Aggregate Transaction Data
			{
				StepNumber:    3,
				ServiceName:   "pos",
				HandlerMethod: "AggregateTransactionData",
				InputMapping: map[string]string{
					"reportID":     "$.steps.1.result.report_id",
					"reportPeriod": "$.input.report_period",
					"transactionFilter": "$.input.transaction_filter",
				},
				TimeoutSeconds:    30,
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
			// Step 4: Calculate Sales Metrics
			{
				StepNumber:    4,
				ServiceName:   "sales",
				HandlerMethod: "CalculateSalesMetrics",
				InputMapping: map[string]string{
					"reportID":           "$.steps.1.result.report_id",
					"transactionData":    "$.steps.3.result.transaction_data",
					"reportType":         "$.input.report_type",
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
			// Step 5: Compute Inventory Metrics
			{
				StepNumber:    5,
				ServiceName:   "inventory",
				HandlerMethod: "ComputeInventoryMetrics",
				InputMapping: map[string]string{
					"reportID":     "$.steps.1.result.report_id",
					"reportPeriod": "$.input.report_period",
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
			// Step 6: Generate Analytics Data
			{
				StepNumber:    6,
				ServiceName:   "analytics",
				HandlerMethod: "GenerateAnalyticsData",
				InputMapping: map[string]string{
					"reportID":          "$.steps.1.result.report_id",
					"salesMetrics":      "$.steps.4.result.sales_metrics",
					"inventoryMetrics":  "$.steps.5.result.inventory_metrics",
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
			// Step 7: Apply Reporting Journal
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "ApplyReportingJournal",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"reportID":       "$.steps.1.result.report_id",
					"analyticsData":  "$.steps.6.result.analytics_data",
					"journalDate":    "$.input.generation_time",
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
			// Step 8: Publish Report
			{
				StepNumber:    8,
				ServiceName:   "analytics",
				HandlerMethod: "PublishReport",
				InputMapping: map[string]string{
					"reportID":      "$.steps.1.result.report_id",
					"analyticsData": "$.steps.6.result.analytics_data",
					"reportFormat":  "$.input.report_format",
				},
				TimeoutSeconds:    20,
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
			// Step 9: Complete Reporting
			{
				StepNumber:    9,
				ServiceName:   "analytics",
				HandlerMethod: "CompleteReporting",
				InputMapping: map[string]string{
					"reportID":        "$.steps.1.result.report_id",
					"journalEntries":  "$.steps.7.result.journal_entries",
					"reportingStatus": "Completed",
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

			// Step 101: Revert Sales Metrics Calculation (compensates step 4)
			{
				StepNumber:    101,
				ServiceName:   "sales",
				HandlerMethod: "RevertSalesMetricsCalculation",
				InputMapping: map[string]string{
					"reportID": "$.steps.1.result.report_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Revert Inventory Metrics Computation (compensates step 5)
			{
				StepNumber:    102,
				ServiceName:   "inventory",
				HandlerMethod: "RevertInventoryMetricsComputation",
				InputMapping: map[string]string{
					"reportID": "$.steps.1.result.report_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 103: Revert Analytics Data Generation (compensates step 6)
			{
				StepNumber:    103,
				ServiceName:   "analytics",
				HandlerMethod: "RevertAnalyticsDataGeneration",
				InputMapping: map[string]string{
					"reportID": "$.steps.1.result.report_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Revert Reporting Journal (compensates step 7)
			{
				StepNumber:    104,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseReportingJournal",
				InputMapping: map[string]string{
					"reportID": "$.steps.1.result.report_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Revert Report Publication (compensates step 8)
			{
				StepNumber:    105,
				ServiceName:   "analytics",
				HandlerMethod: "RevertReportPublication",
				InputMapping: map[string]string{
					"reportID": "$.steps.1.result.report_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: Revert Initiate Report Generation (compensates step 1)
			{
				StepNumber:    106,
				ServiceName:   "analytics",
				HandlerMethod: "RevertInitiateReportGeneration",
				InputMapping: map[string]string{
					"reportID": "$.steps.1.result.report_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 107: Revert Validate Report Parameters (compensates step 2)
			{
				StepNumber:    107,
				ServiceName:   "analytics",
				HandlerMethod: "RevertValidateReportParameters",
				InputMapping: map[string]string{
					"reportID": "$.steps.1.result.report_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 108: Revert Complete Reporting (compensates step 9)
			{
				StepNumber:    108,
				ServiceName:   "analytics",
				HandlerMethod: "RevertCompleteReporting",
				InputMapping: map[string]string{
					"reportID": "$.steps.1.result.report_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *AnalyticsReportingSaga) SagaType() string {
	return "SAGA-R08"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *AnalyticsReportingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *AnalyticsReportingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *AnalyticsReportingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["report_id"] == nil {
		return errors.New("report_id is required")
	}

	if inputMap["report_period"] == nil {
		return errors.New("report_period is required")
	}

	if inputMap["report_type"] == nil {
		return errors.New("report_type is required")
	}

	return nil
}
