// Package manufacturing provides saga handlers for manufacturing module workflows
package manufacturing

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// BatchCostingTraceabilitySaga implements SAGA-M11: Batch/Lot Costing & Traceability workflow
// Business Flow: Create batch ID → Record batch components and quantities →
// Track material consumption per batch → Calculate batch cost per unit →
// Update batch genealogy → Post batch costs to GL → Record batch expiry/recall flags →
// Archive batch traceability chain → Close batch accounting
//
// Compensation: If any critical step fails, automatically reverts genealogy and
// reverses costs to maintain batch traceability and cost integrity
type BatchCostingTraceabilitySaga struct {
	steps []*saga.StepDefinition
}

// NewBatchCostingTraceabilitySaga creates a new Batch/Lot Costing & Traceability saga handler
func NewBatchCostingTraceabilitySaga() saga.SagaHandler {
	return &BatchCostingTraceabilitySaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Batch ID (inventory-core service)
			// Creates unique batch ID with format LOT-YYYYMM-XXXXX
			{
				StepNumber:    1,
				ServiceName:   "inventory-core",
				HandlerMethod: "CreateBatchID",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"productionOrderID": "$.input.production_order_id",
					"productID":        "$.input.product_id",
					"creationDate":     "$.input.creation_date",
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
			// Step 2: Record Batch Components and Quantities (production-order service)
			// Records all components and their quantities in the batch
			{
				StepNumber:    2,
				ServiceName:   "production-order",
				HandlerMethod: "RecordBatchComponentsAndQuantities",
				InputMapping: map[string]string{
					"batchID":              "$.steps.1.result.batch_id",
					"productID":            "$.input.product_id",
					"bomVersion":           "$.input.bom_version",
					"productionQuantity":   "$.input.production_quantity",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 3: Track Material Consumption per Batch (job-card service)
			// Tracks material consumption records for the batch
			{
				StepNumber:    3,
				ServiceName:   "job-card",
				HandlerMethod: "TrackMaterialConsumptionPerBatch",
				InputMapping: map[string]string{
					"batchID":              "$.steps.1.result.batch_id",
					"productionOrderID":    "$.input.production_order_id",
					"trackingPeriod":       "$.input.tracking_period",
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
				CompensationSteps: []int32{109},
			},
			// Step 4: Calculate Batch Cost Per Unit (cost-center service)
			// Calculates cost per unit for the batch
			{
				StepNumber:    4,
				ServiceName:   "cost-center",
				HandlerMethod: "CalculateBatchCostPerUnit",
				InputMapping: map[string]string{
					"batchID":               "$.steps.1.result.batch_id",
					"totalBatchCost":        "$.steps.3.result.total_batch_cost",
					"productionQuantity":    "$.input.production_quantity",
					"costCenterID":          "$.input.cost_center_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 5: Update Batch Genealogy (inventory-core service)
			// Updates batch genealogy tracking parent/child batch relationships
			{
				StepNumber:    5,
				ServiceName:   "inventory-core",
				HandlerMethod: "UpdateBatchGenealogy",
				InputMapping: map[string]string{
					"batchID":           "$.steps.1.result.batch_id",
					"parentBatchID":     "$.input.parent_batch_id",
					"childBatchIDs":     "$.input.child_batch_ids",
					"genealogyRelation": "$.input.genealogy_relation",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{110},
			},
			// Step 6: Post Batch Costs to GL (general-ledger service)
			// Posts batch costs to GL WIP account
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostBatchCostsToGL",
				InputMapping: map[string]string{
					"batchID":          "$.steps.1.result.batch_id",
					"totalBatchCost":   "$.steps.3.result.total_batch_cost",
					"wipAccount":       "$.input.wip_account",
					"costCenterID":     "$.input.cost_center_id",
					"postingDate":      "$.input.posting_date",
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
			// Step 7: Record Batch Expiry/Recall Flags (inventory-core service)
			// Records batch expiry date and potential recall flags
			{
				StepNumber:    7,
				ServiceName:   "inventory-core",
				HandlerMethod: "RecordBatchExpiryRecallFlags",
				InputMapping: map[string]string{
					"batchID":        "$.steps.1.result.batch_id",
					"expiryDate":     "$.input.expiry_date",
					"recallFlag":     "$.input.recall_flag",
					"recallReason":   "$.input.recall_reason",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 8: Archive Batch Traceability Chain (inventory-core service)
			// Archives complete traceability chain for batch
			{
				StepNumber:    8,
				ServiceName:   "inventory-core",
				HandlerMethod: "ArchiveBatchTraceabilityChain",
				InputMapping: map[string]string{
					"batchID":              "$.steps.1.result.batch_id",
					"archiveDate":          "$.input.archive_date",
					"genealogicalRecords": "$.steps.5.result.genealogical_records",
					"costRecords":          "$.steps.6.result.cost_records",
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
				CompensationSteps: []int32{112},
			},
			// Step 9: Close Batch Accounting (cost-center service)
			// Closes batch accounting and finalizes all batch records
			{
				StepNumber:    9,
				ServiceName:   "cost-center",
				HandlerMethod: "CloseBatchAccounting",
				InputMapping: map[string]string{
					"batchID":            "$.steps.1.result.batch_id",
					"costPerUnit":        "$.steps.4.result.cost_per_unit",
					"closingDate":        "$.input.closing_date",
					"batchStatus":        "CLOSED",
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

			// ===== COMPENSATION STEPS =====

			// Compensation Step 109: Undo Material Consumption Tracking (compensates step 3)
			// Undoes material consumption tracking for batch
			{
				StepNumber:    109,
				ServiceName:   "job-card",
				HandlerMethod: "UndoMaterialConsumptionTracking",
				InputMapping: map[string]string{
					"batchID":              "$.steps.1.result.batch_id",
					"productionOrderID":    "$.input.production_order_id",
					"trackingRecords":     "$.steps.3.result.tracking_records",
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
			// Compensation Step 110: Revert Batch Genealogy (compensates step 5)
			// Reverts batch genealogy to previous state
			{
				StepNumber:    110,
				ServiceName:   "inventory-core",
				HandlerMethod: "RevertBatchGenealogy",
				InputMapping: map[string]string{
					"batchID":              "$.steps.1.result.batch_id",
					"previousGenealogy":    "$.input.previous_genealogy",
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
			// Compensation Step 111: Reverse Batch Cost GL Posting (compensates step 6)
			// Reverses GL posting made for batch costs
			{
				StepNumber:    111,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseBatchCostGLPosting",
				InputMapping: map[string]string{
					"batchID":          "$.steps.1.result.batch_id",
					"totalBatchCost":   "$.steps.3.result.total_batch_cost",
					"wipAccount":       "$.input.wip_account",
					"reversalDate":     "$.input.reversal_date",
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
			// Compensation Step 112: Restore Batch Archive (compensates step 8)
			// Restores batch archive from backup
			{
				StepNumber:    112,
				ServiceName:   "inventory-core",
				HandlerMethod: "RestoreBatchArchive",
				InputMapping: map[string]string{
					"batchID":        "$.steps.1.result.batch_id",
					"archiveBackupID": "$.steps.8.result.archive_backup_id",
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
			// Compensation Step 113: Reopen Batch Accounting (compensates step 9)
			// Reopens batch accounting from closed state
			{
				StepNumber:    113,
				ServiceName:   "cost-center",
				HandlerMethod: "ReopenBatchAccounting",
				InputMapping: map[string]string{
					"batchID":      "$.steps.1.result.batch_id",
					"previousStatus": "OPEN",
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
func (s *BatchCostingTraceabilitySaga) SagaType() string {
	return "SAGA-M11"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *BatchCostingTraceabilitySaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *BatchCostingTraceabilitySaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: production_order_id, product_id, creation_date
func (s *BatchCostingTraceabilitySaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type: expected map[string]interface{}")
	}

	// Validate production_order_id
	if inputMap["production_order_id"] == nil || inputMap["production_order_id"] == "" {
		return errors.New("production_order_id is required for Batch Costing & Traceability saga")
	}

	// Validate product_id
	if inputMap["product_id"] == nil || inputMap["product_id"] == "" {
		return errors.New("product_id is required for Batch Costing & Traceability saga")
	}

	// Validate creation_date
	if inputMap["creation_date"] == nil || inputMap["creation_date"] == "" {
		return errors.New("creation_date is required for Batch Costing & Traceability saga")
	}

	// Validate bom_version
	if inputMap["bom_version"] == nil || inputMap["bom_version"] == "" {
		return errors.New("bom_version is required for Batch Costing & Traceability saga")
	}

	// Validate cost_center_id
	if inputMap["cost_center_id"] == nil || inputMap["cost_center_id"] == "" {
		return errors.New("cost_center_id is required for Batch Costing & Traceability saga")
	}

	return nil
}
