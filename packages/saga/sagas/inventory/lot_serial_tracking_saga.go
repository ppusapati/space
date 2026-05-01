// Package inventory provides saga handlers for inventory module workflows
package inventory

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// LotSerialTrackingSaga implements SAGA-I04: Lot/Serial Tracking workflow
// Business Flow: Generate Lot → Assign Serials → Update Lot Master → Record Genealogy → Update Expiry → Activate Lot → Audit Log
type LotSerialTrackingSaga struct {
	steps []*saga.StepDefinition
}

// NewLotSerialTrackingSaga creates a new Lot/Serial Tracking saga handler
func NewLotSerialTrackingSaga() saga.SagaHandler {
	return &LotSerialTrackingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Generate Lot Number
			{
				StepNumber:    1,
				ServiceName:   "lot-serial",
				HandlerMethod: "GenerateLotNumber",
				InputMapping: map[string]string{
					"tenantID":    "$.tenantID",
					"companyID":   "$.companyID",
					"branchID":    "$.branchID",
					"productID":   "$.input.product_id",
					"manufacturingDate": "$.input.manufacturing_date",
				},
				TimeoutSeconds: 15,
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
			// Step 2: Assign Serial Numbers
			{
				StepNumber:    2,
				ServiceName:   "lot-serial",
				HandlerMethod: "AssignSerialNumbers",
				InputMapping: map[string]string{
					"lotID":         "$.steps.1.result.lot_id",
					"serialStart":   "$.input.serial_start",
					"serialEnd":     "$.input.serial_end",
					"quantity":      "$.input.quantity",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{102},
			},
			// Step 3: Update Lot Master
			{
				StepNumber:    3,
				ServiceName:   "lot-serial",
				HandlerMethod: "UpdateLotMaster",
				InputMapping: map[string]string{
					"lotID":              "$.steps.1.result.lot_id",
					"productID":          "$.input.product_id",
					"quantity":           "$.input.quantity",
					"manufacturingDate":  "$.input.manufacturing_date",
					"warehouseID":        "$.input.warehouse_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{103},
			},
			// Step 4: Record Genealogy
			{
				StepNumber:    4,
				ServiceName:   "lot-serial",
				HandlerMethod: "RecordGenealogy",
				InputMapping: map[string]string{
					"lotID":       "$.steps.1.result.lot_id",
					"parentLotID": "$.input.parent_lot_id",
					"rawMaterials": "$.input.raw_materials",
				},
				TimeoutSeconds:    15,
				IsCritical:        false, // Non-critical for basic tracking
				CompensationSteps: []int32{104},
			},
			// Step 5: Update Expiry Tracking
			{
				StepNumber:    5,
				ServiceName:   "lot-serial",
				HandlerMethod: "UpdateExpiryTracking",
				InputMapping: map[string]string{
					"lotID":        "$.steps.1.result.lot_id",
					"expiryDate":   "$.input.expiry_date",
					"shelfLife":    "$.input.shelf_life",
					"productID":    "$.input.product_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        false, // Non-critical
				CompensationSteps: []int32{105},
			},
			// Step 6: Activate Lot
			{
				StepNumber:    6,
				ServiceName:   "lot-serial",
				HandlerMethod: "ActivateLot",
				InputMapping: map[string]string{
					"lotID": "$.steps.1.result.lot_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{106},
			},
			// Step 7: Audit Log (non-critical)
			{
				StepNumber:    7,
				ServiceName:   "audit",
				HandlerMethod: "LogLotCreation",
				InputMapping: map[string]string{
					"lotID":     "$.steps.1.result.lot_id",
					"productID": "$.input.product_id",
					"userID":    "$.input.user_id",
					"action":    "LOT_CREATED",
				},
				TimeoutSeconds:    15,
				IsCritical:        false, // Non-critical audit
				CompensationSteps: []int32{},
			},
			// ===== COMPENSATION STEPS =====

			// Step 101: Delete Lot (compensates step 1)
			{
				StepNumber:    101,
				ServiceName:   "lot-serial",
				HandlerMethod: "DeleteLot",
				InputMapping: map[string]string{
					"lotID": "$.steps.1.result.lot_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 102: Release Serial Numbers (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "lot-serial",
				HandlerMethod: "ReleaseSerialNumbers",
				InputMapping: map[string]string{
					"lotID": "$.steps.1.result.lot_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 103: Delete Lot Master (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "lot-serial",
				HandlerMethod: "DeleteLotMaster",
				InputMapping: map[string]string{
					"lotID": "$.steps.1.result.lot_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 104: Delete Genealogy (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "lot-serial",
				HandlerMethod: "DeleteGenealogy",
				InputMapping: map[string]string{
					"lotID": "$.steps.1.result.lot_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 105: Clear Expiry Data (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "lot-serial",
				HandlerMethod: "ClearExpiryData",
				InputMapping: map[string]string{
					"lotID": "$.steps.1.result.lot_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 106: Deactivate Lot (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "lot-serial",
				HandlerMethod: "DeactivateLot",
				InputMapping: map[string]string{
					"lotID": "$.steps.1.result.lot_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *LotSerialTrackingSaga) SagaType() string {
	return "SAGA-I04"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *LotSerialTrackingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *LotSerialTrackingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *LotSerialTrackingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["product_id"] == nil {
		return errors.New("product_id is required")
	}

	if inputMap["quantity"] == nil {
		return errors.New("quantity is required")
	}

	if inputMap["manufacturing_date"] == nil {
		return errors.New("manufacturing_date is required")
	}

	if inputMap["serial_start"] == nil {
		return errors.New("serial_start is required")
	}

	if inputMap["serial_end"] == nil {
		return errors.New("serial_end is required")
	}

	return nil
}
