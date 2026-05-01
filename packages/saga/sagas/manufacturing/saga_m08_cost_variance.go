// Package manufacturing provides saga handlers for manufacturing module workflows
package manufacturing

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// CostVarianceAnalysisSaga implements SAGA-M08: Cost Variance Analysis workflow
// Business Flow: Extract standard cost from BOM → Extract actual cost from records →
// Calculate material variance → Calculate labor variance → Calculate overhead variance →
// Sum total variance → Analyze variance root causes → Create variance report →
// Alert if variance exceeds tolerance → Archive variance details
//
// Compensation: If any critical step fails, automatically restores original cost records
// to maintain data consistency and accuracy of variance analysis
type CostVarianceAnalysisSaga struct {
	steps []*saga.StepDefinition
}

// NewCostVarianceAnalysisSaga creates a new Cost Variance Analysis saga handler
func NewCostVarianceAnalysisSaga() saga.SagaHandler {
	return &CostVarianceAnalysisSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Extract Standard Cost from BOM (production-order service)
			// Extracts standard cost from bill of materials for the product
			{
				StepNumber:    1,
				ServiceName:   "production-order",
				HandlerMethod: "ExtractStandardCostFromBOM",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"productID":        "$.input.product_id",
					"bomVersion":       "$.input.bom_version",
					"costingPeriod":    "$.input.costing_period",
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
				CompensationSteps: []int32{},
			},
			// Step 2: Extract Actual Cost from Material/Labor Records (cost-center service)
			// Extracts actual costs from material consumption and labor time records
			{
				StepNumber:    2,
				ServiceName:   "cost-center",
				HandlerMethod: "ExtractActualCostFromRecords",
				InputMapping: map[string]string{
					"productionOrderID": "$.input.production_order_id",
					"jobID":            "$.input.job_id",
					"costingPeriod":    "$.input.costing_period",
					"costCenterID":     "$.input.cost_center_id",
				},
				TimeoutSeconds: 60,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{110},
			},
			// Step 3: Calculate Material Variance (cost-center service)
			// Calculates quantity variance and price variance for materials
			{
				StepNumber:    3,
				ServiceName:   "cost-center",
				HandlerMethod: "CalculateMaterialVariance",
				InputMapping: map[string]string{
					"standardMaterialCost": "$.steps.1.result.standard_material_cost",
					"actualMaterialCost":   "$.steps.2.result.actual_material_cost",
					"standardQuantity":     "$.input.standard_quantity",
					"actualQuantity":       "$.steps.2.result.actual_quantity",
					"standardPrice":        "$.steps.1.result.standard_price",
					"actualPrice":          "$.steps.2.result.actual_price",
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
				CompensationSteps: []int32{111},
			},
			// Step 4: Calculate Labor Variance (cost-center service)
			// Calculates labor efficiency variance and rate variance
			{
				StepNumber:    4,
				ServiceName:   "cost-center",
				HandlerMethod: "CalculateLaborVariance",
				InputMapping: map[string]string{
					"standardLaborCost":    "$.steps.1.result.standard_labor_cost",
					"actualLaborCost":      "$.steps.2.result.actual_labor_cost",
					"standardHours":        "$.input.standard_hours",
					"actualHours":          "$.steps.2.result.actual_hours",
					"standardRate":         "$.steps.1.result.standard_rate",
					"actualRate":           "$.steps.2.result.actual_rate",
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
				CompensationSteps: []int32{112},
			},
			// Step 5: Calculate Overhead Variance (cost-center service)
			// Calculates spending variance and volume variance for overhead
			{
				StepNumber:    5,
				ServiceName:   "cost-center",
				HandlerMethod: "CalculateOverheadVariance",
				InputMapping: map[string]string{
					"standardOverheadCost": "$.steps.1.result.standard_overhead_cost",
					"actualOverheadCost":   "$.steps.2.result.actual_overhead_cost",
					"budgetedOverhead":     "$.input.budgeted_overhead",
					"actualActivityLevel":  "$.steps.2.result.actual_activity_level",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 6: Sum Total Variance (cost-center service)
			// Sums all variances to get total variance from standard
			{
				StepNumber:    6,
				ServiceName:   "cost-center",
				HandlerMethod: "CalculateTotalVariance",
				InputMapping: map[string]string{
					"materialVariance":  "$.steps.3.result.total_material_variance",
					"laborVariance":     "$.steps.4.result.total_labor_variance",
					"overheadVariance":  "$.steps.5.result.total_overhead_variance",
					"jobID":            "$.input.job_id",
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
				CompensationSteps: []int32{113},
			},
			// Step 7: Analyze Variance Root Causes (production-order service)
			// Analyzes root causes of variances and categorizes them
			{
				StepNumber:    7,
				ServiceName:   "production-order",
				HandlerMethod: "AnalyzeVarianceRootCauses",
				InputMapping: map[string]string{
					"jobID":              "$.input.job_id",
					"materialVariance":  "$.steps.3.result.total_material_variance",
					"laborVariance":     "$.steps.4.result.total_labor_variance",
					"overheadVariance":  "$.steps.5.result.total_overhead_variance",
					"totalVariance":     "$.steps.6.result.total_variance",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 8: Create Variance Report (general-ledger service)
			// Creates variance report for analysis and management review
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "CreateVarianceReport",
				InputMapping: map[string]string{
					"jobID":              "$.input.job_id",
					"costingPeriod":      "$.input.costing_period",
					"totalVariance":      "$.steps.6.result.total_variance",
					"materialVariance":   "$.steps.3.result.total_material_variance",
					"laborVariance":      "$.steps.4.result.total_labor_variance",
					"overheadVariance":   "$.steps.5.result.total_overhead_variance",
					"rootCauseAnalysis":  "$.steps.7.result.root_cause_analysis",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 9: Alert if Variance Exceeds Tolerance (production-order service)
			// Sends alert if variance exceeds configured tolerance threshold
			{
				StepNumber:    9,
				ServiceName:   "production-order",
				HandlerMethod: "AlertIfVarianceExceedsTolerance",
				InputMapping: map[string]string{
					"jobID":          "$.input.job_id",
					"totalVariance": "$.steps.6.result.total_variance",
					"varianceTolerance": "$.input.variance_tolerance_percent",
					"standardCost": "$.steps.1.result.total_standard_cost",
					"reportID": "$.steps.8.result.report_id",
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
				CompensationSteps: []int32{114},
			},
			// Step 10: Archive Variance Details (cost-center service)
			// Archives variance analysis records for historical tracking
			{
				StepNumber:    10,
				ServiceName:   "cost-center",
				HandlerMethod: "ArchiveVarianceDetails",
				InputMapping: map[string]string{
					"jobID":              "$.input.job_id",
					"costingPeriod":      "$.input.costing_period",
					"reportID":           "$.steps.8.result.report_id",
					"totalVariance":      "$.steps.6.result.total_variance",
					"varianceComponents": "$.steps.7.result.variance_components",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},

			// ===== COMPENSATION STEPS =====

			// Compensation Step 110: Restore Actual Cost Records (compensates step 2)
			// Restores actual cost records from backup if analysis fails
			{
				StepNumber:    110,
				ServiceName:   "cost-center",
				HandlerMethod: "RestoreActualCostRecords",
				InputMapping: map[string]string{
					"jobID":          "$.input.job_id",
					"costingPeriod":  "$.input.costing_period",
					"backupID":      "$.steps.2.result.backup_id",
				},
				TimeoutSeconds: 60,
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
			// Compensation Step 111: Restore Material Cost Baseline (compensates step 3)
			// Restores material cost records to previous state
			{
				StepNumber:    111,
				ServiceName:   "cost-center",
				HandlerMethod: "RestoreMaterialCostBaseline",
				InputMapping: map[string]string{
					"jobID":          "$.input.job_id",
					"costingPeriod":  "$.input.costing_period",
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
				CompensationSteps: []int32{},
			},
			// Compensation Step 112: Restore Labor Cost Baseline (compensates step 4)
			// Restores labor cost records to previous state
			{
				StepNumber:    112,
				ServiceName:   "cost-center",
				HandlerMethod: "RestoreLaborCostBaseline",
				InputMapping: map[string]string{
					"jobID":          "$.input.job_id",
					"costingPeriod":  "$.input.costing_period",
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
				CompensationSteps: []int32{},
			},
			// Compensation Step 113: Cancel Variance Summary (compensates step 6)
			// Cancels the variance summary record
			{
				StepNumber:    113,
				ServiceName:   "cost-center",
				HandlerMethod: "CancelVarianceSummary",
				InputMapping: map[string]string{
					"jobID":          "$.input.job_id",
					"varianceSummaryID": "$.steps.6.result.variance_summary_id",
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
			// Compensation Step 114: Cancel Variance Alert (compensates step 9)
			// Cancels any alerts sent for exceeding tolerance
			{
				StepNumber:    114,
				ServiceName:   "production-order",
				HandlerMethod: "CancelVarianceAlert",
				InputMapping: map[string]string{
					"jobID":    "$.input.job_id",
					"alertID": "$.steps.9.result.alert_id",
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
func (s *CostVarianceAnalysisSaga) SagaType() string {
	return "SAGA-M08"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *CostVarianceAnalysisSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *CostVarianceAnalysisSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: product_id, job_id, cost_center_id, costing_period
func (s *CostVarianceAnalysisSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type: expected map[string]interface{}")
	}

	// Validate product_id
	if inputMap["product_id"] == nil || inputMap["product_id"] == "" {
		return errors.New("product_id is required for Cost Variance Analysis saga")
	}

	// Validate job_id
	if inputMap["job_id"] == nil || inputMap["job_id"] == "" {
		return errors.New("job_id is required for Cost Variance Analysis saga")
	}

	// Validate cost_center_id
	if inputMap["cost_center_id"] == nil || inputMap["cost_center_id"] == "" {
		return errors.New("cost_center_id is required for Cost Variance Analysis saga")
	}

	// Validate costing_period
	if inputMap["costing_period"] == nil || inputMap["costing_period"] == "" {
		return errors.New("costing_period is required for Cost Variance Analysis saga")
	}

	// Validate bom_version
	if inputMap["bom_version"] == nil || inputMap["bom_version"] == "" {
		return errors.New("bom_version is required for Cost Variance Analysis saga")
	}

	return nil
}
