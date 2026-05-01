// Package manufacturing provides saga handlers for manufacturing module workflows
package manufacturing

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ScrapReworkManagementSaga implements SAGA-M09: Scrap & Rework Management workflow
// Business Flow: Identify scrap quantity and reason → Determine if rework or scrap →
// If rework: Route for rework process → If scrap: Remove from WIP inventory →
// Calculate scrap loss value → Post scrap loss to GL → Update cost records →
// Adjust production order status → Send scrap notification
//
// Compensation: If any critical step fails, automatically restores inventory and
// reverses GL postings to maintain accurate inventory and financial records
type ScrapReworkManagementSaga struct {
	steps []*saga.StepDefinition
}

// NewScrapReworkManagementSaga creates a new Scrap & Rework Management saga handler
func NewScrapReworkManagementSaga() saga.SagaHandler {
	return &ScrapReworkManagementSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Identify Scrap Quantity and Reason (quality-production service)
			// Identifies scrap quantity and captures reason code
			{
				StepNumber:    1,
				ServiceName:   "quality-production",
				HandlerMethod: "IdentifyScrapQuantityAndReason",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"productionOrderID": "$.input.production_order_id",
					"scrapQuantity":    "$.input.scrap_quantity",
					"scrapReasonCode":  "$.input.scrap_reason_code",
					"qualityInspectionID": "$.input.quality_inspection_id",
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
			// Step 2: Determine if Rework or Scrap (quality-production service)
			// Determines whether scrap should be sent for rework or disposed
			{
				StepNumber:    2,
				ServiceName:   "quality-production",
				HandlerMethod: "DetermineReworkOrScrap",
				InputMapping: map[string]string{
					"scrapQuantity":    "$.steps.1.result.scrap_quantity",
					"scrapReasonCode":  "$.steps.1.result.scrap_reason_code",
					"reworkCostLimit":  "$.input.rework_cost_limit",
					"productionOrderID": "$.input.production_order_id",
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
			// Step 3: Route for Rework Process (production-order service)
			// Routes scrap items for rework if applicable
			{
				StepNumber:    3,
				ServiceName:   "production-order",
				HandlerMethod: "RouteForReworkProcess",
				InputMapping: map[string]string{
					"productionOrderID": "$.input.production_order_id",
					"scrapQuantity":     "$.steps.1.result.scrap_quantity",
					"reworkFlag":        "$.steps.2.result.is_rework",
					"reworkRoutingCode": "$.input.rework_routing_code",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{109},
			},
			// Step 4: Remove Scrap from WIP Inventory (inventory-core service)
			// Removes scrap quantity from work-in-progress inventory
			{
				StepNumber:    4,
				ServiceName:   "inventory-core",
				HandlerMethod: "RemoveScrapFromWIPInventory",
				InputMapping: map[string]string{
					"productionOrderID": "$.input.production_order_id",
					"scrapQuantity":     "$.steps.1.result.scrap_quantity",
					"wipLocation":       "$.input.wip_location",
					"scrapFlag":         "$.steps.2.result.is_scrap",
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
			// Step 5: Calculate Scrap Loss Value (cost-center service)
			// Calculates the financial loss value of scrap
			{
				StepNumber:    5,
				ServiceName:   "cost-center",
				HandlerMethod: "CalculateScrapLossValue",
				InputMapping: map[string]string{
					"scrapQuantity":    "$.steps.1.result.scrap_quantity",
					"unitCost":         "$.input.unit_cost",
					"salvageValue":     "$.input.salvage_value",
					"scrapReasonCode":  "$.steps.1.result.scrap_reason_code",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 6: Post Scrap Loss to GL (general-ledger service)
			// Posts scrap loss amount to GL expense account
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostScrapLossToGL",
				InputMapping: map[string]string{
					"scrapLossValue":   "$.steps.5.result.scrap_loss_value",
					"scrapExpenseAccount": "$.input.scrap_expense_account",
					"wipAccount":       "$.input.wip_account",
					"postingDate":      "$.input.posting_date",
					"productionOrderID": "$.input.production_order_id",
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
			// Step 7: Update Cost Records with Scrap Impact (cost-center service)
			// Updates job/product cost records to reflect scrap loss impact
			{
				StepNumber:    7,
				ServiceName:   "cost-center",
				HandlerMethod: "UpdateCostRecordsWithScrapImpact",
				InputMapping: map[string]string{
					"jobID":             "$.input.job_id",
					"scrapLossValue":    "$.steps.5.result.scrap_loss_value",
					"originalCost":      "$.input.original_cost",
					"updatedCostStatus": "SCRAP_ADJUSTED",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 8: Adjust Production Order Status (production-order service)
			// Adjusts production order status to reflect scrap/rework handling
			{
				StepNumber:    8,
				ServiceName:   "production-order",
				HandlerMethod: "AdjustProductionOrderStatus",
				InputMapping: map[string]string{
					"productionOrderID": "$.input.production_order_id",
					"scrapQuantity":     "$.steps.1.result.scrap_quantity",
					"reworkFlag":        "$.steps.2.result.is_rework",
					"statusAdjustment":  "$.steps.2.result.next_status",
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
			// Step 9: Send Scrap Notification for Reporting (production-order service)
			// Sends notification for scrap event reporting and tracking
			{
				StepNumber:    9,
				ServiceName:   "production-order",
				HandlerMethod: "SendScrapNotification",
				InputMapping: map[string]string{
					"productionOrderID": "$.input.production_order_id",
					"scrapQuantity":     "$.steps.1.result.scrap_quantity",
					"scrapLossValue":    "$.steps.5.result.scrap_loss_value",
					"scrapReasonCode":   "$.steps.1.result.scrap_reason_code",
					"notificationRecipients": "$.input.notification_recipients",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},

			// ===== COMPENSATION STEPS =====

			// Compensation Step 109: Undo Rework Routing (compensates step 3)
			// Undoes the rework routing if operation is cancelled
			{
				StepNumber:    109,
				ServiceName:   "production-order",
				HandlerMethod: "UndoReworkRouting",
				InputMapping: map[string]string{
					"productionOrderID": "$.input.production_order_id",
					"reworkRoutingID":  "$.steps.3.result.rework_routing_id",
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
			// Compensation Step 110: Restore Scrap to WIP Inventory (compensates step 4)
			// Restores scrap quantity back to WIP inventory
			{
				StepNumber:    110,
				ServiceName:   "inventory-core",
				HandlerMethod: "RestoreScrapToWIPInventory",
				InputMapping: map[string]string{
					"productionOrderID": "$.input.production_order_id",
					"scrapQuantity":     "$.steps.1.result.scrap_quantity",
					"wipLocation":       "$.input.wip_location",
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
			// Compensation Step 111: Reverse Scrap Loss GL Posting (compensates step 6)
			// Reverses GL posting made for scrap loss
			{
				StepNumber:    111,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseScrapLossGLPosting",
				InputMapping: map[string]string{
					"scrapLossValue":   "$.steps.5.result.scrap_loss_value",
					"scrapExpenseAccount": "$.input.scrap_expense_account",
					"reversalDate":     "$.input.reversal_date",
					"productionOrderID": "$.input.production_order_id",
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
			// Compensation Step 112: Revert Production Order Status (compensates step 8)
			// Reverts production order status to previous state
			{
				StepNumber:    112,
				ServiceName:   "production-order",
				HandlerMethod: "RevertProductionOrderStatus",
				InputMapping: map[string]string{
					"productionOrderID": "$.input.production_order_id",
					"previousStatus":    "$.input.previous_po_status",
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
func (s *ScrapReworkManagementSaga) SagaType() string {
	return "SAGA-M09"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ScrapReworkManagementSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ScrapReworkManagementSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: production_order_id, scrap_quantity, scrap_reason_code
func (s *ScrapReworkManagementSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type: expected map[string]interface{}")
	}

	// Validate production_order_id
	if inputMap["production_order_id"] == nil || inputMap["production_order_id"] == "" {
		return errors.New("production_order_id is required for Scrap & Rework saga")
	}

	// Validate scrap_quantity
	if inputMap["scrap_quantity"] == nil {
		return errors.New("scrap_quantity is required for Scrap & Rework saga")
	}

	// Validate scrap_reason_code
	if inputMap["scrap_reason_code"] == nil || inputMap["scrap_reason_code"] == "" {
		return errors.New("scrap_reason_code is required for Scrap & Rework saga")
	}

	// Validate quality_inspection_id
	if inputMap["quality_inspection_id"] == nil || inputMap["quality_inspection_id"] == "" {
		return errors.New("quality_inspection_id is required for Scrap & Rework saga")
	}

	// Validate unit_cost
	if inputMap["unit_cost"] == nil {
		return errors.New("unit_cost is required for Scrap & Rework saga")
	}

	return nil
}
