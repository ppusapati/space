// Package inventory provides saga handlers for inventory module workflows
package inventory

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// QualityRejectionSaga implements SAGA-I03: Quality Rejection workflow
// Business Flow: Create Inspection → Fail Inspection → Create Rejection → Adjust Stock → Update Lot Status → Post GL
type QualityRejectionSaga struct {
	steps []*saga.StepDefinition
}

// NewQualityRejectionSaga creates a new Quality Rejection saga handler
func NewQualityRejectionSaga() saga.SagaHandler {
	return &QualityRejectionSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Create Inspection
			{
				StepNumber:    1,
				ServiceName:   "qc",
				HandlerMethod: "CreateInspection",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"receiptID":     "$.input.receipt_id",
					"productID":     "$.input.product_id",
					"quantity":      "$.input.quantity",
					"inspectionType": "$.input.inspection_type",
				},
				TimeoutSeconds: 20,
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
			// Step 2: Fail Inspection
			{
				StepNumber:    2,
				ServiceName:   "qc",
				HandlerMethod: "FailInspection",
				InputMapping: map[string]string{
					"inspectionID": "$.steps.1.result.inspection_id",
					"failureReason": "$.input.failure_reason",
					"failureNotes":  "$.input.failure_notes",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{102},
			},
			// Step 3: Create Rejection Record
			{
				StepNumber:    3,
				ServiceName:   "qc",
				HandlerMethod: "CreateRejectionRecord",
				InputMapping: map[string]string{
					"inspectionID":  "$.steps.1.result.inspection_id",
					"rejectionType": "$.input.rejection_type",
					"quantity":      "$.input.quantity",
					"receiptID":     "$.input.receipt_id",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{103},
			},
			// Step 4: Adjust Rejected Stock
			{
				StepNumber:    4,
				ServiceName:   "inventory-core",
				HandlerMethod: "AdjustRejectedStock",
				InputMapping: map[string]string{
					"receiptID":      "$.input.receipt_id",
					"productID":      "$.input.product_id",
					"quantity":       "$.input.quantity",
					"rejectionID":    "$.steps.3.result.rejection_id",
					"warehouseID":    "$.input.warehouse_id",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{104},
			},
			// Step 5: Update Lot Status
			{
				StepNumber:    5,
				ServiceName:   "lot-serial",
				HandlerMethod: "UpdateLotStatus",
				InputMapping: map[string]string{
					"lotID":     "$.input.lot_id",
					"newStatus": "REJECTED",
					"reason":    "Quality inspection failed",
				},
				TimeoutSeconds:    15,
				IsCritical:        true,
				CompensationSteps: []int32{105},
			},
			// Step 6: Post Rejection Journal
			{
				StepNumber:    6,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostRejectionJournal",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"rejectionID":   "$.steps.3.result.rejection_id",
					"quantity":      "$.input.quantity",
					"unitCost":      "$.input.unit_cost",
					"receiptID":     "$.input.receipt_id",
					"journalDate":   "$.input.journal_date",
				},
				TimeoutSeconds:    20,
				IsCritical:        true,
				CompensationSteps: []int32{106},
			},
			// ===== COMPENSATION STEPS =====

			// Step 101: Cancel Inspection (compensates step 1)
			{
				StepNumber:    101,
				ServiceName:   "qc",
				HandlerMethod: "CancelInspection",
				InputMapping: map[string]string{
					"inspectionID": "$.steps.1.result.inspection_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 102: Revert Inspection Result (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "qc",
				HandlerMethod: "RevertInspectionResult",
				InputMapping: map[string]string{
					"inspectionID": "$.steps.1.result.inspection_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 103: Delete Rejection Record (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "qc",
				HandlerMethod: "DeleteRejectionRecord",
				InputMapping: map[string]string{
					"rejectionID": "$.steps.3.result.rejection_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 104: Reverse Rejection (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "inventory-core",
				HandlerMethod: "ReverseRejection",
				InputMapping: map[string]string{
					"rejectionID": "$.steps.3.result.rejection_id",
					"quantity":    "$.input.quantity",
					"warehouseID": "$.input.warehouse_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 105: Revert Lot Status (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "lot-serial",
				HandlerMethod: "RevertLotStatus",
				InputMapping: map[string]string{
					"lotID": "$.input.lot_id",
				},
				TimeoutSeconds: 15,
				IsCritical:     false,
			},
			// Step 106: Reverse Rejection Journal (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseRejectionJournal",
				InputMapping: map[string]string{
					"rejectionID": "$.steps.3.result.rejection_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *QualityRejectionSaga) SagaType() string {
	return "SAGA-I03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *QualityRejectionSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *QualityRejectionSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *QualityRejectionSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["receipt_id"] == nil {
		return errors.New("receipt_id is required")
	}

	if inputMap["product_id"] == nil {
		return errors.New("product_id is required")
	}

	if inputMap["quantity"] == nil {
		return errors.New("quantity is required")
	}

	if inputMap["failure_reason"] == nil {
		return errors.New("failure_reason is required")
	}

	if inputMap["lot_id"] == nil {
		return errors.New("lot_id is required")
	}

	return nil
}
