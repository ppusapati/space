// Package manufacturing provides saga handlers for manufacturing module workflows
package manufacturing

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// MRPLotSizingOptimizationSaga implements SAGA-M12: MRP & Lot Sizing Optimization workflow
// Business Flow: Explode demand into requirements → Calculate MRP net requirements →
// Run lot sizing algorithm → Calculate planned order quantities → Determine procurement timing →
// Create planned purchase orders → Create planned production orders → Validate plan against capacity →
// Update inventory forecasts → Archive MRP plan
//
// Compensation: If any critical step fails, automatically reverts planned orders and
// restores forecasts to maintain accurate demand planning
type MRPLotSizingOptimizationSaga struct {
	steps []*saga.StepDefinition
}

// NewMRPLotSizingOptimizationSaga creates a new MRP & Lot Sizing Optimization saga handler
func NewMRPLotSizingOptimizationSaga() saga.SagaHandler {
	return &MRPLotSizingOptimizationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Explode Demand into Requirements (production-planning service)
			// Explodes master demand schedule into material requirements
			{
				StepNumber:    1,
				ServiceName:   "production-planning",
				HandlerMethod: "ExplodeDemandIntoRequirements",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"planningHorizon":  "$.input.planning_horizon",
					"productID":        "$.input.product_id",
					"demandQuantity":   "$.input.demand_quantity",
					"demandDate":       "$.input.demand_date",
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
				CompensationSteps: []int32{},
			},
			// Step 2: Calculate MRP Net Requirements (production-planning service)
			// Calculates net requirements considering on-hand inventory and safety stock
			{
				StepNumber:    2,
				ServiceName:   "production-planning",
				HandlerMethod: "CalculateMRPNetRequirements",
				InputMapping: map[string]string{
					"explosionID":      "$.steps.1.result.explosion_id",
					"grossRequirements": "$.steps.1.result.gross_requirements",
					"onHandInventory":  "$.input.on_hand_inventory",
					"safetyStock":      "$.input.safety_stock",
					"scheduledReceipts": "$.input.scheduled_receipts",
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
			// Step 3: Run Lot Sizing Algorithm (cost-center service)
			// Runs lot sizing algorithm (EOQ, LUC, POQ, etc.)
			{
				StepNumber:    3,
				ServiceName:   "cost-center",
				HandlerMethod: "RunLotSizingAlgorithm",
				InputMapping: map[string]string{
					"netRequirements":   "$.steps.2.result.net_requirements",
					"lotSizingMethod":   "$.input.lot_sizing_method",
					"orderingCost":      "$.input.ordering_cost",
					"holdingCost":       "$.input.holding_cost",
					"productID":         "$.input.product_id",
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
				CompensationSteps: []int32{111},
			},
			// Step 4: Calculate Planned Order Quantities (production-planning service)
			// Calculates planned order quantities based on lot sizing results
			{
				StepNumber:    4,
				ServiceName:   "production-planning",
				HandlerMethod: "CalculatePlannedOrderQuantities",
				InputMapping: map[string]string{
					"lotSizingResults":  "$.steps.3.result.lot_sizing_results",
					"minimumOrderQty":   "$.input.minimum_order_qty",
					"maximumOrderQty":   "$.input.maximum_order_qty",
					"multipleOrderQty":  "$.input.multiple_order_qty",
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
			// Step 5: Determine Procurement Timing (production-planning service)
			// Determines procurement timing considering lead times
			{
				StepNumber:    5,
				ServiceName:   "production-planning",
				HandlerMethod: "DetermineProcurementTiming",
				InputMapping: map[string]string{
					"plannedOrders":     "$.steps.4.result.planned_orders",
					"procurementLeadTime": "$.input.procurement_lead_time",
					"productionLeadTime": "$.input.production_lead_time",
					"productID":         "$.input.product_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},
			// Step 6: Create Planned Purchase Orders (procurement service)
			// Creates planned purchase orders for procurement
			{
				StepNumber:    6,
				ServiceName:   "procurement",
				HandlerMethod: "CreatePlannedPurchaseOrders",
				InputMapping: map[string]string{
					"plannedOrders":     "$.steps.4.result.planned_orders",
					"procurementTiming": "$.steps.5.result.procurement_timing",
					"vendorID":          "$.input.vendor_id",
					"costCenterID":      "$.input.cost_center_id",
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
				CompensationSteps: []int32{113},
			},
			// Step 7: Create Planned Production Orders (production-planning service)
			// Creates planned production orders for internal manufacturing
			{
				StepNumber:    7,
				ServiceName:   "production-planning",
				HandlerMethod: "CreatePlannedProductionOrders",
				InputMapping: map[string]string{
					"plannedOrders":     "$.steps.4.result.planned_orders",
					"procurementTiming": "$.steps.5.result.procurement_timing",
					"productID":         "$.input.product_id",
					"costCenterID":      "$.input.cost_center_id",
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
				CompensationSteps: []int32{114},
			},
			// Step 8: Validate Plan Against Capacity (work-center service)
			// Validates MRP plan against work center capacity
			{
				StepNumber:    8,
				ServiceName:   "work-center",
				HandlerMethod: "ValidatePlanAgainstCapacity",
				InputMapping: map[string]string{
					"plannedProductionOrders": "$.steps.7.result.planned_production_orders",
					"planningHorizon":       "$.input.planning_horizon",
					"capacityBuffer":        "$.input.capacity_buffer",
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
				CompensationSteps: []int32{115},
			},
			// Step 9: Update Inventory Forecasts (inventory-core service)
			// Updates inventory forecasts based on MRP plan
			{
				StepNumber:    9,
				ServiceName:   "inventory-core",
				HandlerMethod: "UpdateInventoryForecasts",
				InputMapping: map[string]string{
					"plannedOrders":       "$.steps.4.result.planned_orders",
					"projectedInventory": "$.steps.8.result.projected_inventory",
					"planningHorizon":     "$.input.planning_horizon",
					"productID":           "$.input.product_id",
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
				CompensationSteps: []int32{116},
			},
			// Step 10: Archive MRP Plan (production-planning service)
			// Archives MRP plan for historical tracking and analysis
			{
				StepNumber:    10,
				ServiceName:   "production-planning",
				HandlerMethod: "ArchiveMRPPlan",
				InputMapping: map[string]string{
					"planID":               "$.steps.1.result.explosion_id",
					"planningHorizon":      "$.input.planning_horizon",
					"plannedPOs":           "$.steps.6.result.planned_po_ids",
					"plannedProdOrders":    "$.steps.7.result.planned_production_order_ids",
					"capacityValidation":   "$.steps.8.result.validation_result",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{},
			},

			// ===== COMPENSATION STEPS =====

			// Compensation Step 110: Restore Net Requirements (compensates step 2)
			// Restores net requirements calculation from backup
			{
				StepNumber:    110,
				ServiceName:   "production-planning",
				HandlerMethod: "RestoreNetRequirements",
				InputMapping: map[string]string{
					"explosionID":    "$.steps.1.result.explosion_id",
					"backupID":      "$.steps.2.result.backup_id",
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
			// Compensation Step 111: Restore Lot Sizing Results (compensates step 3)
			// Restores lot sizing calculation results
			{
				StepNumber:    111,
				ServiceName:   "cost-center",
				HandlerMethod: "RestoreLotSizingResults",
				InputMapping: map[string]string{
					"productID":          "$.input.product_id",
					"backupID":          "$.steps.3.result.backup_id",
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
			// Compensation Step 112: Cancel Planned Order Quantities (compensates step 4)
			// Cancels planned order quantities
			{
				StepNumber:    112,
				ServiceName:   "production-planning",
				HandlerMethod: "CancelPlannedOrderQuantities",
				InputMapping: map[string]string{
					"plannedOrders": "$.steps.4.result.planned_orders",
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
			// Compensation Step 113: Cancel Planned Purchase Orders (compensates step 6)
			// Cancels planned purchase orders
			{
				StepNumber:    113,
				ServiceName:   "procurement",
				HandlerMethod: "CancelPlannedPurchaseOrders",
				InputMapping: map[string]string{
					"plannedPOIDs": "$.steps.6.result.planned_po_ids",
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
			// Compensation Step 114: Cancel Planned Production Orders (compensates step 7)
			// Cancels planned production orders
			{
				StepNumber:    114,
				ServiceName:   "production-planning",
				HandlerMethod: "CancelPlannedProductionOrders",
				InputMapping: map[string]string{
					"plannedProdOrderIDs": "$.steps.7.result.planned_production_order_ids",
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
			// Compensation Step 115: Restore Capacity Validation (compensates step 8)
			// Restores capacity validation status
			{
				StepNumber:    115,
				ServiceName:   "work-center",
				HandlerMethod: "RestoreCapacityValidation",
				InputMapping: map[string]string{
					"planningHorizon": "$.input.planning_horizon",
					"validationBackupID": "$.steps.8.result.validation_backup_id",
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
			// Compensation Step 116: Restore Inventory Forecasts (compensates step 9)
			// Restores inventory forecasts to previous state
			{
				StepNumber:    116,
				ServiceName:   "inventory-core",
				HandlerMethod: "RestoreInventoryForecasts",
				InputMapping: map[string]string{
					"productID":          "$.input.product_id",
					"planningHorizon":    "$.input.planning_horizon",
					"forecastBackupID":  "$.steps.9.result.forecast_backup_id",
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
		},
	}
}

// SagaType returns the saga type identifier
func (s *MRPLotSizingOptimizationSaga) SagaType() string {
	return "SAGA-M12"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *MRPLotSizingOptimizationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *MRPLotSizingOptimizationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
// Required fields: planning_horizon, product_id, demand_quantity, demand_date
func (s *MRPLotSizingOptimizationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type: expected map[string]interface{}")
	}

	// Validate planning_horizon
	if inputMap["planning_horizon"] == nil || inputMap["planning_horizon"] == "" {
		return errors.New("planning_horizon is required for MRP & Lot Sizing saga")
	}

	// Validate product_id
	if inputMap["product_id"] == nil || inputMap["product_id"] == "" {
		return errors.New("product_id is required for MRP & Lot Sizing saga")
	}

	// Validate demand_quantity
	if inputMap["demand_quantity"] == nil {
		return errors.New("demand_quantity is required for MRP & Lot Sizing saga")
	}

	// Validate demand_date
	if inputMap["demand_date"] == nil || inputMap["demand_date"] == "" {
		return errors.New("demand_date is required for MRP & Lot Sizing saga")
	}

	// Validate lot_sizing_method
	if inputMap["lot_sizing_method"] == nil || inputMap["lot_sizing_method"] == "" {
		return errors.New("lot_sizing_method is required for MRP & Lot Sizing saga (EOQ, LUC, POQ, etc.)")
	}

	return nil
}
