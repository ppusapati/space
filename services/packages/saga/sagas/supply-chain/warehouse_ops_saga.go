// Package supplychain provides saga handlers for supply chain module workflows
package supplychain

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// WarehouseOperationsManagementSaga implements SAGA-SC02: Warehouse Operations & Management
// Business Flow: InitializeWarehouseOperation → PlanWarehouseLayout → AllocateStorageLocations → MovePallets → ReconcileInventoryLevels → UpdateStockLocations → ScheduleLaborTasks → ProcessWarehouseOrders → PostWarehouseJournals → CompleteWarehouseOperation
// Timeout: 180 seconds, Critical steps: 1,2,3,4,7,10
type WarehouseOperationsManagementSaga struct {
	steps []*saga.StepDefinition
}

// NewWarehouseOperationsManagementSaga creates a new Warehouse Operations & Management saga handler
func NewWarehouseOperationsManagementSaga() saga.SagaHandler {
	return &WarehouseOperationsManagementSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initialize Warehouse Operation
			{
				StepNumber:    1,
				ServiceName:   "warehouse",
				HandlerMethod: "InitializeWarehouseOperation",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"warehouseID":    "$.input.warehouse_id",
					"operationDate":  "$.input.operation_date",
					"operationType":  "$.input.operation_type",
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
			// Step 2: Plan Warehouse Layout
			{
				StepNumber:    2,
				ServiceName:   "warehouse",
				HandlerMethod: "PlanWarehouseLayout",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"warehouseID":      "$.input.warehouse_id",
					"operationDate":    "$.input.operation_date",
					"operationType":    "$.input.operation_type",
					"operationData":    "$.steps.1.result.operation_data",
				},
				TimeoutSeconds: 45,
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
			// Step 3: Allocate Storage Locations
			{
				StepNumber:    3,
				ServiceName:   "inventory",
				HandlerMethod: "AllocateStorageLocations",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"warehouseID":   "$.input.warehouse_id",
					"operationData": "$.steps.1.result.operation_data",
					"layoutPlan":    "$.steps.2.result.layout_plan",
				},
				TimeoutSeconds: 45,
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
			// Step 4: Move Pallets
			{
				StepNumber:    4,
				ServiceName:   "warehouse",
				HandlerMethod: "MovePallets",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"warehouseID":         "$.input.warehouse_id",
					"operationType":       "$.input.operation_type",
					"layoutPlan":          "$.steps.2.result.layout_plan",
					"storageAllocations":  "$.steps.3.result.storage_allocations",
				},
				TimeoutSeconds: 60,
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
			// Step 5: Reconcile Inventory Levels
			{
				StepNumber:    5,
				ServiceName:   "inventory",
				HandlerMethod: "ReconcileInventoryLevels",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"warehouseID":        "$.input.warehouse_id",
					"storageAllocations": "$.steps.3.result.storage_allocations",
					"palletMovements":    "$.steps.4.result.pallet_movements",
				},
				TimeoutSeconds: 45,
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
			// Step 6: Update Stock Locations
			{
				StepNumber:    6,
				ServiceName:   "warehouse",
				HandlerMethod: "UpdateStockLocations",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"warehouseID":     "$.input.warehouse_id",
					"palletMovements": "$.steps.4.result.pallet_movements",
					"inventoryReconciliation": "$.steps.5.result.inventory_reconciliation",
				},
				TimeoutSeconds: 45,
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
			// Step 7: Schedule Labor Tasks
			{
				StepNumber:    7,
				ServiceName:   "labor-management",
				HandlerMethod: "ScheduleLaborTasks",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"warehouseID":     "$.input.warehouse_id",
					"operationDate":   "$.input.operation_date",
					"operationType":   "$.input.operation_type",
					"palletMovements": "$.steps.4.result.pallet_movements",
				},
				TimeoutSeconds: 45,
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
			// Step 8: Process Warehouse Orders
			{
				StepNumber:    8,
				ServiceName:   "warehouse",
				HandlerMethod: "ProcessWarehouseOrders",
				InputMapping: map[string]string{
					"tenantID":                  "$.tenantID",
					"companyID":                 "$.companyID",
					"branchID":                  "$.branchID",
					"warehouseID":               "$.input.warehouse_id",
					"operationType":             "$.input.operation_type",
					"stockLocationUpdates":      "$.steps.6.result.stock_location_updates",
					"laborSchedule":             "$.steps.7.result.labor_schedule",
				},
				TimeoutSeconds: 45,
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
			// Step 9: Post Warehouse Journals
			{
				StepNumber:    9,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostWarehouseJournals",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"warehouseID":     "$.input.warehouse_id",
					"operationDate":   "$.input.operation_date",
					"inventoryReconciliation": "$.steps.5.result.inventory_reconciliation",
					"warehouseOrders": "$.steps.8.result.warehouse_orders",
				},
				TimeoutSeconds: 45,
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
			// Step 10: Complete Warehouse Operation
			{
				StepNumber:    10,
				ServiceName:   "warehouse",
				HandlerMethod: "CompleteWarehouseOperation",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"warehouseID":    "$.input.warehouse_id",
					"operationDate":  "$.input.operation_date",
					"warehouseOrders": "$.steps.8.result.warehouse_orders",
					"journalEntries": "$.steps.9.result.journal_entries",
				},
				TimeoutSeconds: 30,
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

			// Step 102: UndoLayoutPlan (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "warehouse",
				HandlerMethod: "UndoLayoutPlan",
				InputMapping: map[string]string{
					"warehouseID": "$.input.warehouse_id",
					"layoutPlan":  "$.steps.2.result.layout_plan",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 103: DeallocateStorageLocations (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "inventory",
				HandlerMethod: "DeallocateStorageLocations",
				InputMapping: map[string]string{
					"warehouseID":         "$.input.warehouse_id",
					"storageAllocations":  "$.steps.3.result.storage_allocations",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 104: ReversePalletMovements (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "warehouse",
				HandlerMethod: "ReversePalletMovements",
				InputMapping: map[string]string{
					"warehouseID":      "$.input.warehouse_id",
					"palletMovements":  "$.steps.4.result.pallet_movements",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 105: ReverseInventoryReconciliation (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "inventory",
				HandlerMethod: "ReverseInventoryReconciliation",
				InputMapping: map[string]string{
					"warehouseID":              "$.input.warehouse_id",
					"inventoryReconciliation": "$.steps.5.result.inventory_reconciliation",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 106: ReverseStockLocationUpdate (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "warehouse",
				HandlerMethod: "ReverseStockLocationUpdate",
				InputMapping: map[string]string{
					"warehouseID":          "$.input.warehouse_id",
					"stockLocationUpdates": "$.steps.6.result.stock_location_updates",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 107: CancelLaborSchedule (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "labor-management",
				HandlerMethod: "CancelLaborSchedule",
				InputMapping: map[string]string{
					"warehouseID":    "$.input.warehouse_id",
					"laborSchedule":  "$.steps.7.result.labor_schedule",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 108: CancelWarehouseOrders (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "warehouse",
				HandlerMethod: "CancelWarehouseOrders",
				InputMapping: map[string]string{
					"warehouseID":     "$.input.warehouse_id",
					"warehouseOrders": "$.steps.8.result.warehouse_orders",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 109: ReverseWarehouseJournals (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseWarehouseJournals",
				InputMapping: map[string]string{
					"warehouseID":    "$.input.warehouse_id",
					"journalEntries": "$.steps.9.result.journal_entries",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 110: CancelWarehouseOperationCompletion (compensates step 10)
			{
				StepNumber:    110,
				ServiceName:   "warehouse",
				HandlerMethod: "CancelWarehouseOperationCompletion",
				InputMapping: map[string]string{
					"warehouseID":     "$.input.warehouse_id",
					"operationDate":   "$.input.operation_date",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *WarehouseOperationsManagementSaga) SagaType() string {
	return "SAGA-SC02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *WarehouseOperationsManagementSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *WarehouseOperationsManagementSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *WarehouseOperationsManagementSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["warehouse_id"] == nil {
		return errors.New("warehouse_id is required")
	}

	warehouseID, ok := inputMap["warehouse_id"].(string)
	if !ok || warehouseID == "" {
		return errors.New("warehouse_id must be a non-empty string")
	}

	if inputMap["operation_date"] == nil {
		return errors.New("operation_date is required")
	}

	operationDate, ok := inputMap["operation_date"].(string)
	if !ok || operationDate == "" {
		return errors.New("operation_date must be a non-empty string")
	}

	if inputMap["operation_type"] == nil {
		return errors.New("operation_type is required")
	}

	operationType, ok := inputMap["operation_type"].(string)
	if !ok || operationType == "" {
		return errors.New("operation_type must be a non-empty string")
	}

	return nil
}
