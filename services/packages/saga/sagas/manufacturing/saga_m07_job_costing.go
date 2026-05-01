// Package manufacturing provides saga handlers for manufacturing module workflows
package manufacturing

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// JobCostingSaga implements SAGA-M07: Job Costing & Overhead Allocation workflow
// Business Flow: Identify job for costing → Extract material costs → Extract labor costs →
// Get overhead allocation rate → Calculate allocated overhead → Sum total job cost →
// Validate cost against budget → Post job costs to GL → Create cost variance analysis →
// Archive cost records → Update job status with finalized costs
//
// Compensation: If any critical step fails, automatically reverses GL postings and
// restores cost records to maintain financial integrity and data consistency
type JobCostingSaga struct {
	steps []*saga.StepDefinition
}

// NewJobCostingSaga creates a new Job Costing & Overhead Allocation saga handler
func NewJobCostingSaga() saga.SagaHandler {
	return &JobCostingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Identify Job for Costing (production-order service)
			// Identifies and retrieves the production order/job for costing
			{
				StepNumber:    1,
				ServiceName:   "production-order",
				HandlerMethod: "IdentifyJobForCosting",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"productionOrderID": "$.input.production_order_id",
					"jobID":            "$.input.job_id",
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
			// Step 2: Extract Material Costs (job-card service)
			// Extracts material consumption costs from job card records
			{
				StepNumber:    2,
				ServiceName:   "job-card",
				HandlerMethod: "ExtractMaterialCosts",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
					"jobID":            "$.steps.1.result.job_id",
					"costingPeriod":    "$.input.costing_period",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 3: Extract Labor Costs (job-card service)
			// Extracts labor costs from job card time tracking records
			{
				StepNumber:    3,
				ServiceName:   "job-card",
				HandlerMethod: "ExtractLaborCosts",
				InputMapping: map[string]string{
					"productionOrderID": "$.steps.1.result.production_order_id",
					"jobID":            "$.steps.1.result.job_id",
					"costingPeriod":    "$.input.costing_period",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 4: Get Overhead Allocation Rate (cost-center service)
			// Retrieves the overhead allocation rate for the cost center
			{
				StepNumber:    4,
				ServiceName:   "cost-center",
				HandlerMethod: "GetOverheadAllocationRate",
				InputMapping: map[string]string{
					"costCenterID":  "$.input.cost_center_id",
					"costingPeriod": "$.input.costing_period",
					"allocationBase": "$.input.allocation_base",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 5: Calculate Allocated Overhead (cost-center service)
			// Calculates overhead allocation based on material or labor basis
			{
				StepNumber:    5,
				ServiceName:   "cost-center",
				HandlerMethod: "CalculateAllocatedOverhead",
				InputMapping: map[string]string{
					"materialCost":          "$.steps.2.result.material_cost",
					"laborCost":             "$.steps.3.result.labor_cost",
					"overheadAllocationRate": "$.steps.4.result.allocation_rate",
					"allocationBase":        "$.input.allocation_base",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 6: Sum Total Job Cost (cost-center service)
			// Sums material, labor, and overhead costs to get total job cost
			{
				StepNumber:    6,
				ServiceName:   "cost-center",
				HandlerMethod: "CalculateTotalJobCost",
				InputMapping: map[string]string{
					"materialCost":      "$.steps.2.result.material_cost",
					"laborCost":         "$.steps.3.result.labor_cost",
					"overheadCost":      "$.steps.5.result.overhead_cost",
					"costCenterID":      "$.input.cost_center_id",
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
			// Step 7: Validate Cost Against Budget (work-center service)
			// Validates that total job cost is within approved budget
			{
				StepNumber:    7,
				ServiceName:   "work-center",
				HandlerMethod: "ValidateCostAgainstBudget",
				InputMapping: map[string]string{
					"jobID":         "$.steps.1.result.job_id",
					"totalJobCost": "$.steps.6.result.total_job_cost",
					"budgetAmount": "$.input.budget_amount",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 8: Post Job Costs to GL (general-ledger service)
			// Posts calculated job costs to work-in-progress GL account
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostJobCostToGL",
				InputMapping: map[string]string{
					"jobID":          "$.steps.1.result.job_id",
					"totalJobCost":  "$.steps.6.result.total_job_cost",
					"wipAccountCode": "$.input.wip_account_code",
					"costCenterID":   "$.input.cost_center_id",
					"postingDate":    "$.input.posting_date",
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
				CompensationSteps: []int32{110},
			},
			// Step 9: Create Cost Variance Analysis (cost-center service)
			// Creates analysis of cost variances between planned and actual
			{
				StepNumber:    9,
				ServiceName:   "cost-center",
				HandlerMethod: "CreateCostVarianceAnalysis",
				InputMapping: map[string]string{
					"jobID":              "$.steps.1.result.job_id",
					"plannedCost":       "$.input.planned_cost",
					"actualCost":        "$.steps.6.result.total_job_cost",
					"materialVariance":  "$.steps.2.result.variance",
					"laborVariance":     "$.steps.3.result.variance",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 10: Archive Cost Records (production-order service)
			// Archives the cost records for historical tracking and compliance
			{
				StepNumber:    10,
				ServiceName:   "production-order",
				HandlerMethod: "ArchiveCostRecords",
				InputMapping: map[string]string{
					"jobID":             "$.steps.1.result.job_id",
					"costingPeriod":     "$.input.costing_period",
					"totalJobCost":     "$.steps.6.result.total_job_cost",
					"varianceAnalysisID": "$.steps.9.result.variance_analysis_id",
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
			// Step 11: Update Job Status with Finalized Costs (production-order service)
			// Updates job status to "Cost Finalized" with cost details
			{
				StepNumber:    11,
				ServiceName:   "production-order",
				HandlerMethod: "UpdateJobStatusWithFinalizedCosts",
				InputMapping: map[string]string{
					"jobID":          "$.steps.1.result.job_id",
					"totalJobCost":  "$.steps.6.result.total_job_cost",
					"costStatus":    "FINALIZED",
					"varianceID":    "$.steps.9.result.variance_analysis_id",
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
				CompensationSteps: []int32{112},
			},

			// ===== COMPENSATION STEPS =====

			// Compensation Step 110: Reverse GL Postings (compensates step 8)
			// Reverses GL postings made to WIP account
			{
				StepNumber:    110,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseJobCostGLPosting",
				InputMapping: map[string]string{
					"jobID":          "$.steps.1.result.job_id",
					"totalJobCost":  "$.steps.6.result.total_job_cost",
					"wipAccountCode": "$.input.wip_account_code",
					"reversalDate":   "$.input.reversal_date",
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
			// Compensation Step 111: Restore Cost Records (compensates step 10)
			// Restores archived cost records from archive
			{
				StepNumber:    111,
				ServiceName:   "production-order",
				HandlerMethod: "RestoreCostRecordsFromArchive",
				InputMapping: map[string]string{
					"jobID":          "$.steps.1.result.job_id",
					"costingPeriod":  "$.input.costing_period",
					"archiveID":     "$.steps.10.result.archive_id",
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
			// Compensation Step 112: Revert Job Status (compensates step 11)
			// Reverts job status from "Cost Finalized" back to previous state
			{
				StepNumber:    112,
				ServiceName:   "production-order",
				HandlerMethod: "RevertJobStatus",
				InputMapping: map[string]string{
					"jobID":         "$.steps.1.result.job_id",
					"previousStatus": "$.input.previous_job_status",
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
func (s *JobCostingSaga) SagaType() string {
	return "SAGA-M07"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *JobCostingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *JobCostingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: production_order_id, job_id, costing_period, cost_center_id, allocation_base
func (s *JobCostingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type: expected map[string]interface{}")
	}

	// Validate production_order_id
	if inputMap["production_order_id"] == nil || inputMap["production_order_id"] == "" {
		return errors.New("production_order_id is required for Job Costing saga")
	}

	// Validate job_id
	if inputMap["job_id"] == nil || inputMap["job_id"] == "" {
		return errors.New("job_id is required for Job Costing saga")
	}

	// Validate costing_period
	if inputMap["costing_period"] == nil || inputMap["costing_period"] == "" {
		return errors.New("costing_period is required for Job Costing saga")
	}

	// Validate cost_center_id
	if inputMap["cost_center_id"] == nil || inputMap["cost_center_id"] == "" {
		return errors.New("cost_center_id is required for Job Costing saga")
	}

	// Validate allocation_base
	if inputMap["allocation_base"] == nil || inputMap["allocation_base"] == "" {
		return errors.New("allocation_base is required for Job Costing saga (MATERIAL or LABOR)")
	}

	// Validate wip_account_code
	if inputMap["wip_account_code"] == nil || inputMap["wip_account_code"] == "" {
		return errors.New("wip_account_code is required for Job Costing saga")
	}

	return nil
}
