// Package supplychain provides saga handlers for supply chain module workflows
package supplychain

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// SupplyChainVisibilityTrackingSaga implements SAGA-SC07: Supply Chain Visibility & Tracking
// Business Flow: InitializeTracking → CollectTrackingData → ProcessLocationUpdates → AggregateVisibilityData → UpdateShipmentStatus → GenerateVisibilityReport → NotifyStakeholders → ArchiveTrackingData
// Timeout: 180 seconds, Critical steps: 1,2,3,4,7,8
type SupplyChainVisibilityTrackingSaga struct {
	steps []*saga.StepDefinition
}

// NewSupplyChainVisibilityTrackingSaga creates a new Supply Chain Visibility & Tracking saga handler
func NewSupplyChainVisibilityTrackingSaga() saga.SagaHandler {
	return &SupplyChainVisibilityTrackingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initialize Tracking
			{
				StepNumber:    1,
				ServiceName:   "supply-chain-visibility",
				HandlerMethod: "InitializeTracking",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"companyID":   "$.companyID",
					"branchID":    "$.branchID",
					"trackingID":  "$.input.tracking_id",
					"shipmentID":  "$.input.shipment_id",
					"startDate":   "$.input.start_date",
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
			// Step 2: Collect Tracking Data
			{
				StepNumber:    2,
				ServiceName:   "logistics",
				HandlerMethod: "CollectTrackingData",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"shipmentID":    "$.input.shipment_id",
					"startDate":     "$.input.start_date",
					"trackingData":  "$.steps.1.result.tracking_data",
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
			// Step 3: Process Location Updates
			{
				StepNumber:    3,
				ServiceName:   "warehouse",
				HandlerMethod: "ProcessLocationUpdates",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"shipmentID":       "$.input.shipment_id",
					"collectedData":    "$.steps.2.result.collected_data",
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
			// Step 4: Aggregate Visibility Data
			{
				StepNumber:    4,
				ServiceName:   "supply-chain-visibility",
				HandlerMethod: "AggregateVisibilityData",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"trackingID":          "$.input.tracking_id",
					"processedLocations":  "$.steps.3.result.processed_locations",
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
			// Step 5: Update Shipment Status
			{
				StepNumber:    5,
				ServiceName:   "shipment",
				HandlerMethod: "UpdateShipmentStatus",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"shipmentID":         "$.input.shipment_id",
					"aggregatedData":     "$.steps.4.result.aggregated_data",
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
			// Step 6: Generate Visibility Report
			{
				StepNumber:    6,
				ServiceName:   "supply-chain-visibility",
				HandlerMethod: "GenerateVisibilityReport",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"trackingID":     "$.input.tracking_id",
					"aggregatedData": "$.steps.4.result.aggregated_data",
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
			// Step 7: Notify Stakeholders
			{
				StepNumber:    7,
				ServiceName:   "inventory",
				HandlerMethod: "NotifyStakeholders",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"shipmentID":     "$.input.shipment_id",
					"visibilityReport": "$.steps.6.result.visibility_report",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Archive Tracking Data
			{
				StepNumber:    8,
				ServiceName:   "supply-chain-visibility",
				HandlerMethod: "ArchiveTrackingData",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"trackingID":     "$.input.tracking_id",
					"shipmentID":     "$.input.shipment_id",
					"aggregatedData": "$.steps.4.result.aggregated_data",
				},
				TimeoutSeconds: 45,
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

			// Step 102: ReverseTrackingCollection (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "logistics",
				HandlerMethod: "ReverseTrackingCollection",
				InputMapping: map[string]string{
					"shipmentID":    "$.input.shipment_id",
					"collectedData": "$.steps.2.result.collected_data",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 103: ReverseLocationProcessing (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "warehouse",
				HandlerMethod: "ReverseLocationProcessing",
				InputMapping: map[string]string{
					"shipmentID":          "$.input.shipment_id",
					"processedLocations":  "$.steps.3.result.processed_locations",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 104: ReverseAggregation (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "supply-chain-visibility",
				HandlerMethod: "ReverseAggregation",
				InputMapping: map[string]string{
					"trackingID":     "$.input.tracking_id",
					"aggregatedData": "$.steps.4.result.aggregated_data",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 105: ReverseShipmentStatusUpdate (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "shipment",
				HandlerMethod: "ReverseShipmentStatusUpdate",
				InputMapping: map[string]string{
					"shipmentID": "$.input.shipment_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 106: ReverseVisibilityReport (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "supply-chain-visibility",
				HandlerMethod: "ReverseVisibilityReport",
				InputMapping: map[string]string{
					"trackingID":      "$.input.tracking_id",
					"visibilityReport": "$.steps.6.result.visibility_report",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 107: ReverseStakeholderNotification (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "inventory",
				HandlerMethod: "ReverseStakeholderNotification",
				InputMapping: map[string]string{
					"shipmentID":     "$.input.shipment_id",
					"visibilityReport": "$.steps.6.result.visibility_report",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 108: CancelTrackingArchival (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "supply-chain-visibility",
				HandlerMethod: "CancelTrackingArchival",
				InputMapping: map[string]string{
					"trackingID":  "$.input.tracking_id",
					"shipmentID":  "$.input.shipment_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *SupplyChainVisibilityTrackingSaga) SagaType() string {
	return "SAGA-SC07"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *SupplyChainVisibilityTrackingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *SupplyChainVisibilityTrackingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *SupplyChainVisibilityTrackingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["tracking_id"] == nil {
		return errors.New("tracking_id is required")
	}

	trackingID, ok := inputMap["tracking_id"].(string)
	if !ok || trackingID == "" {
		return errors.New("tracking_id must be a non-empty string")
	}

	if inputMap["shipment_id"] == nil {
		return errors.New("shipment_id is required")
	}

	shipmentID, ok := inputMap["shipment_id"].(string)
	if !ok || shipmentID == "" {
		return errors.New("shipment_id must be a non-empty string")
	}

	if inputMap["start_date"] == nil {
		return errors.New("start_date is required")
	}

	startDate, ok := inputMap["start_date"].(string)
	if !ok || startDate == "" {
		return errors.New("start_date must be a non-empty string")
	}

	return nil
}
