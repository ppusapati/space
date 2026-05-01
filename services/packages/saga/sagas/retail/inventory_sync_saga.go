// Package retail provides saga handlers for retail workflows
package retail

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// InventorySyncSaga implements SAGA-R02: Inventory Synchronization (Multi-Channel) workflow
// Business Flow: InitiateSync → ValidateSyncChannels → LoadInventorySnapshot → SyncToChannel1 → SyncToChannel2 → SyncToChannel3 → ReconcileInventory → UpdateStockLevels → PublishInventoryUpdate → ApplyInventoryJournal → CompleteSyncRun
// Steps: 11 forward + 10 compensation = 21 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,6,8,11
type InventorySyncSaga struct {
	steps []*saga.StepDefinition
}

// NewInventorySyncSaga creates a new Inventory Synchronization saga handler
func NewInventorySyncSaga() saga.SagaHandler {
	return &InventorySyncSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Sync Run
			{
				StepNumber:    1,
				ServiceName:   "inventory",
				HandlerMethod: "InitiateSyncRun",
				InputMapping: map[string]string{
					"tenantID":     "$.tenantID",
					"companyID":    "$.companyID",
					"branchID":     "$.branchID",
					"syncRunID":    "$.input.sync_run_id",
					"syncType":     "$.input.sync_type",
					"channelList":  "$.input.channel_list",
					"syncTime":     "$.input.sync_time",
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
			// Step 2: Validate Sync Channels
			{
				StepNumber:    2,
				ServiceName:   "pos",
				HandlerMethod: "ValidateSyncChannels",
				InputMapping: map[string]string{
					"syncRunID":   "$.steps.1.result.sync_run_id",
					"channelList": "$.input.channel_list",
				},
				TimeoutSeconds:    20,
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
			// Step 3: Load Inventory Snapshot
			{
				StepNumber:    3,
				ServiceName:   "inventory",
				HandlerMethod: "LoadInventorySnapshot",
				InputMapping: map[string]string{
					"syncRunID": "$.steps.1.result.sync_run_id",
					"syncType":  "$.input.sync_type",
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
			// Step 4: Sync to Channel 1
			{
				StepNumber:    4,
				ServiceName:   "pos",
				HandlerMethod: "SyncInventoryToChannel",
				InputMapping: map[string]string{
					"syncRunID":        "$.steps.1.result.sync_run_id",
					"channel":          "0",
					"inventoryData":    "$.steps.3.result.inventory_snapshot",
					"channelEndpoint":  "$.input.channel_list[0]",
				},
				TimeoutSeconds:    35,
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
			// Step 5: Sync to Channel 2
			{
				StepNumber:    5,
				ServiceName:   "ecommerce",
				HandlerMethod: "SyncInventoryToChannel",
				InputMapping: map[string]string{
					"syncRunID":        "$.steps.1.result.sync_run_id",
					"channel":          "1",
					"inventoryData":    "$.steps.3.result.inventory_snapshot",
					"channelEndpoint":  "$.input.channel_list[1]",
				},
				TimeoutSeconds:    35,
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
			// Step 6: Sync to Channel 3
			{
				StepNumber:    6,
				ServiceName:   "warehouse",
				HandlerMethod: "SyncInventoryToWarehouse",
				InputMapping: map[string]string{
					"syncRunID":        "$.steps.1.result.sync_run_id",
					"channel":          "2",
					"inventoryData":    "$.steps.3.result.inventory_snapshot",
				},
				TimeoutSeconds:    35,
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
			// Step 7: Reconcile Inventory
			{
				StepNumber:    7,
				ServiceName:   "inventory",
				HandlerMethod: "ReconcileInventory",
				InputMapping: map[string]string{
					"syncRunID":       "$.steps.1.result.sync_run_id",
					"channel1Sync":    "$.steps.4.result.sync_status",
					"channel2Sync":    "$.steps.5.result.sync_status",
					"channel3Sync":    "$.steps.6.result.sync_status",
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
			// Step 8: Update Stock Levels
			{
				StepNumber:    8,
				ServiceName:   "inventory",
				HandlerMethod: "UpdateStockLevels",
				InputMapping: map[string]string{
					"syncRunID":       "$.steps.1.result.sync_run_id",
					"reconciliationData": "$.steps.7.result.reconciliation_result",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Publish Inventory Update
			{
				StepNumber:    9,
				ServiceName:   "pos",
				HandlerMethod: "PublishInventoryUpdate",
				InputMapping: map[string]string{
					"syncRunID":   "$.steps.1.result.sync_run_id",
					"stockLevels": "$.steps.8.result.updated_stock_levels",
				},
				TimeoutSeconds:    20,
				IsCritical:        false,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Apply Inventory Journal
			{
				StepNumber:    10,
				ServiceName:   "general-ledger",
				HandlerMethod: "ApplyInventorySyncJournal",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"companyID":   "$.companyID",
					"branchID":    "$.branchID",
					"syncRunID":   "$.steps.1.result.sync_run_id",
					"syncType":    "$.input.sync_type",
					"journalDate": "$.input.sync_time",
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
			// Step 11: Complete Sync Run
			{
				StepNumber:    11,
				ServiceName:   "inventory",
				HandlerMethod: "CompleteSyncRun",
				InputMapping: map[string]string{
					"syncRunID":        "$.steps.1.result.sync_run_id",
					"journalEntries":   "$.steps.10.result.journal_entries",
					"syncStatus":       "Completed",
				},
				TimeoutSeconds:    20,
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

			// Step 101: Revert Channel 1 Sync (compensates step 4)
			{
				StepNumber:    101,
				ServiceName:   "pos",
				HandlerMethod: "RevertChannelSync",
				InputMapping: map[string]string{
					"syncRunID": "$.steps.1.result.sync_run_id",
					"channel":   "0",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 102: Revert Channel 2 Sync (compensates step 5)
			{
				StepNumber:    102,
				ServiceName:   "ecommerce",
				HandlerMethod: "RevertChannelSync",
				InputMapping: map[string]string{
					"syncRunID": "$.steps.1.result.sync_run_id",
					"channel":   "1",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 103: Revert Warehouse Sync (compensates step 6)
			{
				StepNumber:    103,
				ServiceName:   "warehouse",
				HandlerMethod: "RevertWarehouseSync",
				InputMapping: map[string]string{
					"syncRunID": "$.steps.1.result.sync_run_id",
				},
				TimeoutSeconds: 35,
				IsCritical:     false,
			},
			// Step 104: Revert Inventory Reconciliation (compensates step 7)
			{
				StepNumber:    104,
				ServiceName:   "inventory",
				HandlerMethod: "RevertInventoryReconciliation",
				InputMapping: map[string]string{
					"syncRunID": "$.steps.1.result.sync_run_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Revert Stock Level Updates (compensates step 8)
			{
				StepNumber:    105,
				ServiceName:   "inventory",
				HandlerMethod: "RevertStockLevelUpdates",
				InputMapping: map[string]string{
					"syncRunID": "$.steps.1.result.sync_run_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: Revert Inventory Update Publication (compensates step 9)
			{
				StepNumber:    106,
				ServiceName:   "pos",
				HandlerMethod: "RevertInventoryUpdatePublication",
				InputMapping: map[string]string{
					"syncRunID": "$.steps.1.result.sync_run_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: Reverse Inventory Sync Journal (compensates step 10)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseInventorySyncJournal",
				InputMapping: map[string]string{
					"syncRunID": "$.steps.1.result.sync_run_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 108: Revert Initiate Sync (compensates step 1)
			{
				StepNumber:    108,
				ServiceName:   "inventory",
				HandlerMethod: "RevertInitiateSyncRun",
				InputMapping: map[string]string{
					"syncRunID": "$.steps.1.result.sync_run_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 109: Revert Validate Sync Channels (compensates step 2)
			{
				StepNumber:    109,
				ServiceName:   "pos",
				HandlerMethod: "RevertValidateSyncChannels",
				InputMapping: map[string]string{
					"syncRunID": "$.steps.1.result.sync_run_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 110: Revert Load Inventory Snapshot (compensates step 3)
			{
				StepNumber:    110,
				ServiceName:   "inventory",
				HandlerMethod: "RevertLoadInventorySnapshot",
				InputMapping: map[string]string{
					"syncRunID": "$.steps.1.result.sync_run_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *InventorySyncSaga) SagaType() string {
	return "SAGA-R02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *InventorySyncSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *InventorySyncSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *InventorySyncSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["sync_run_id"] == nil {
		return errors.New("sync_run_id is required")
	}

	if inputMap["sync_type"] == nil {
		return errors.New("sync_type is required")
	}

	if inputMap["channel_list"] == nil {
		return errors.New("channel_list is required")
	}

	return nil
}
