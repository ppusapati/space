// Package supplychain provides saga handlers for supply chain module workflows
package supplychain

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// DistributionCenterOperationsSaga implements SAGA-SC05: Distribution Center Operations
// Business Flow: InitializeDCOperation → ReceiveBatchInventory → SortAndCategorize → QualityCheck → AllocateToRegions → PrepareForShipment → UpdateDCInventory → PostDCJournals → CompleteDCOperation
// Timeout: 180 seconds, Critical steps: 1,2,3,4,6,9
type DistributionCenterOperationsSaga struct {
	steps []*saga.StepDefinition
}

// NewDistributionCenterOperationsSaga creates a new Distribution Center Operations saga handler
func NewDistributionCenterOperationsSaga() saga.SagaHandler {
	return &DistributionCenterOperationsSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initialize DC Operation
			{
				StepNumber:    1,
				ServiceName:   "distribution-center",
				HandlerMethod: "InitializeDCOperation",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"dcID":           "$.input.dc_id",
					"operationDate":  "$.input.operation_date",
					"batchID":        "$.input.batch_id",
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
			// Step 2: Receive Batch Inventory
			{
				StepNumber:    2,
				ServiceName:   "warehouse",
				HandlerMethod: "ReceiveBatchInventory",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"dcID":             "$.input.dc_id",
					"batchID":          "$.input.batch_id",
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
			// Step 3: Sort and Categorize
			{
				StepNumber:    3,
				ServiceName:   "distribution-center",
				HandlerMethod: "SortAndCategorize",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"batchID":         "$.input.batch_id",
					"batchInventory":  "$.steps.2.result.batch_inventory",
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
			// Step 4: Quality Check
			{
				StepNumber:    4,
				ServiceName:   "distribution-center",
				HandlerMethod: "QualityCheck",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"batchID":         "$.input.batch_id",
					"sortedInventory": "$.steps.3.result.sorted_inventory",
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
			// Step 5: Allocate to Regions
			{
				StepNumber:    5,
				ServiceName:   "warehouse",
				HandlerMethod: "AllocateToRegions",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"dcID":               "$.input.dc_id",
					"batchID":            "$.input.batch_id",
					"qualityCheckedInventory": "$.steps.4.result.quality_checked_inventory",
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
			// Step 6: Prepare for Shipment
			{
				StepNumber:    6,
				ServiceName:   "distribution-center",
				HandlerMethod: "PrepareForShipment",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"dcID":               "$.input.dc_id",
					"batchID":            "$.input.batch_id",
					"regionalAllocations": "$.steps.5.result.regional_allocations",
				},
				TimeoutSeconds: 45,
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
			// Step 7: Update DC Inventory
			{
				StepNumber:    7,
				ServiceName:   "inventory",
				HandlerMethod: "UpdateDCInventory",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"dcID":              "$.input.dc_id",
					"batchID":           "$.input.batch_id",
					"shipmentPreparation": "$.steps.6.result.shipment_preparation",
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
			// Step 8: Post DC Journals
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostDCJournals",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"dcID":              "$.input.dc_id",
					"operationDate":     "$.input.operation_date",
					"shipmentPreparation": "$.steps.6.result.shipment_preparation",
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
			// Step 9: Complete DC Operation
			{
				StepNumber:    9,
				ServiceName:   "distribution-center",
				HandlerMethod: "CompleteDCOperation",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"dcID":                "$.input.dc_id",
					"batchID":             "$.input.batch_id",
					"shipmentPreparation": "$.steps.6.result.shipment_preparation",
					"journalEntries":      "$.steps.8.result.journal_entries",
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

			// Step 102: RejectBatchInventory (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "warehouse",
				HandlerMethod: "RejectBatchInventory",
				InputMapping: map[string]string{
					"dcID":           "$.input.dc_id",
					"batchID":        "$.input.batch_id",
					"batchInventory": "$.steps.2.result.batch_inventory",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 103: UndoSortAndCategorize (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "distribution-center",
				HandlerMethod: "UndoSortAndCategorize",
				InputMapping: map[string]string{
					"batchID":         "$.input.batch_id",
					"sortedInventory": "$.steps.3.result.sorted_inventory",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 104: ReverseQualityCheck (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "distribution-center",
				HandlerMethod: "ReverseQualityCheck",
				InputMapping: map[string]string{
					"batchID":                "$.input.batch_id",
					"qualityCheckedInventory": "$.steps.4.result.quality_checked_inventory",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 105: ReverseRegionalAllocation (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "warehouse",
				HandlerMethod: "ReverseRegionalAllocation",
				InputMapping: map[string]string{
					"dcID":                "$.input.dc_id",
					"regionalAllocations": "$.steps.5.result.regional_allocations",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 106: UndoShipmentPreparation (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "distribution-center",
				HandlerMethod: "UndoShipmentPreparation",
				InputMapping: map[string]string{
					"dcID":                  "$.input.dc_id",
					"batchID":               "$.input.batch_id",
					"shipmentPreparation":   "$.steps.6.result.shipment_preparation",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 107: ReverseDCInventoryUpdate (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "inventory",
				HandlerMethod: "ReverseDCInventoryUpdate",
				InputMapping: map[string]string{
					"dcID":    "$.input.dc_id",
					"batchID": "$.input.batch_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 108: ReverseDCJournals (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseDCJournals",
				InputMapping: map[string]string{
					"dcID":           "$.input.dc_id",
					"journalEntries": "$.steps.8.result.journal_entries",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 109: CancelDCOperationCompletion (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "distribution-center",
				HandlerMethod: "CancelDCOperationCompletion",
				InputMapping: map[string]string{
					"dcID":    "$.input.dc_id",
					"batchID": "$.input.batch_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *DistributionCenterOperationsSaga) SagaType() string {
	return "SAGA-SC05"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *DistributionCenterOperationsSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *DistributionCenterOperationsSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *DistributionCenterOperationsSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["dc_id"] == nil {
		return errors.New("dc_id is required")
	}

	dcID, ok := inputMap["dc_id"].(string)
	if !ok || dcID == "" {
		return errors.New("dc_id must be a non-empty string")
	}

	if inputMap["operation_date"] == nil {
		return errors.New("operation_date is required")
	}

	operationDate, ok := inputMap["operation_date"].(string)
	if !ok || operationDate == "" {
		return errors.New("operation_date must be a non-empty string")
	}

	if inputMap["batch_id"] == nil {
		return errors.New("batch_id is required")
	}

	batchID, ok := inputMap["batch_id"].(string)
	if !ok || batchID == "" {
		return errors.New("batch_id must be a non-empty string")
	}

	return nil
}
