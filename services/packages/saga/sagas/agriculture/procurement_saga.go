// Package agriculture provides saga handlers for agricultural workflows
package agriculture

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// ProcurementSaga implements SAGA-A04: Agricultural Procurement & Supply Chain workflow
// Business Flow: InitiateProcurement → ValidateProduceQuality → CreateProcurementOrder → ProcessWarehouseReceipt → UpdateInventoryStock → MatchReceiptWithOrder → ProcessPayableEntry → UpdateSupplyChain → PostProcurementJournal → CompleteProcurement
// Steps: 10 forward + 9 compensation = 19 total
// Timeout: 120 seconds, Critical steps: 1,2,3,4,7,10
type ProcurementSaga struct {
	steps []*saga.StepDefinition
}

// NewProcurementSaga creates a new Agricultural Procurement saga handler
func NewProcurementSaga() saga.SagaHandler {
	return &ProcurementSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initiate Procurement
			{
				StepNumber:    1,
				ServiceName:   "agriculture",
				HandlerMethod: "InitiateProcurement",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"procurementID":  "$.input.procurement_id",
					"farmID":         "$.input.farm_id",
					"produceType":    "$.input.produce_type",
					"quantity":       "$.input.quantity",
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
			// Step 2: Validate Produce Quality
			{
				StepNumber:    2,
				ServiceName:   "quality-inspection",
				HandlerMethod: "ValidateProduceQuality",
				InputMapping: map[string]string{
					"procurementID":  "$.steps.1.result.procurement_id",
					"produceType":    "$.input.produce_type",
					"quantity":       "$.input.quantity",
					"validateQuality": "true",
				},
				TimeoutSeconds:    30,
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
			// Step 3: Create Procurement Order
			{
				StepNumber:    3,
				ServiceName:   "procurement",
				HandlerMethod: "CreateProcurementOrder",
				InputMapping: map[string]string{
					"procurementID":  "$.steps.1.result.procurement_id",
					"produceType":    "$.input.produce_type",
					"quantity":       "$.input.quantity",
					"qualityResult":  "$.steps.2.result.quality_result",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{102},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Process Warehouse Receipt
			{
				StepNumber:    4,
				ServiceName:   "warehouse",
				HandlerMethod: "ProcessWarehouseReceipt",
				InputMapping: map[string]string{
					"procurementID":     "$.steps.1.result.procurement_id",
					"procurementOrder":  "$.steps.3.result.procurement_order",
					"quantity":          "$.input.quantity",
				},
				TimeoutSeconds:    30,
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
			// Step 5: Update Inventory Stock
			{
				StepNumber:    5,
				ServiceName:   "inventory",
				HandlerMethod: "UpdateInventoryStock",
				InputMapping: map[string]string{
					"procurementID": "$.steps.1.result.procurement_id",
					"produceType":   "$.input.produce_type",
					"quantity":      "$.input.quantity",
					"receiptData":   "$.steps.4.result.receipt_data",
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
			// Step 6: Match Receipt with Order
			{
				StepNumber:    6,
				ServiceName:   "procurement",
				HandlerMethod: "MatchReceiptWithOrder",
				InputMapping: map[string]string{
					"procurementID":     "$.steps.1.result.procurement_id",
					"procurementOrder":  "$.steps.3.result.procurement_order",
					"receiptData":       "$.steps.4.result.receipt_data",
				},
				TimeoutSeconds:    25,
				IsCritical:        false,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Process Payable Entry
			{
				StepNumber:    7,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ProcessPayableEntry",
				InputMapping: map[string]string{
					"procurementID": "$.steps.1.result.procurement_id",
					"matchResult":   "$.steps.6.result.match_result",
					"quantity":      "$.input.quantity",
				},
				TimeoutSeconds:    30,
				IsCritical:        true,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Update Supply Chain Records
			{
				StepNumber:    8,
				ServiceName:   "agriculture",
				HandlerMethod: "UpdateSupplyChainRecords",
				InputMapping: map[string]string{
					"procurementID": "$.steps.1.result.procurement_id",
					"farmID":        "$.input.farm_id",
					"produceType":   "$.input.produce_type",
				},
				TimeoutSeconds:    20,
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
			// Step 9: Post Procurement Journal Entries
			{
				StepNumber:    9,
				ServiceName:   "general-ledger",
				HandlerMethod: "ApplyProcurementJournal",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"procurementID": "$.steps.1.result.procurement_id",
					"payableEntry":  "$.steps.7.result.payable_entry",
					"journalDate":   "$.input.procurement_date",
				},
				TimeoutSeconds:    30,
				IsCritical:        false,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 10: Complete Procurement
			{
				StepNumber:    10,
				ServiceName:   "agriculture",
				HandlerMethod: "CompleteProcurement",
				InputMapping: map[string]string{
					"procurementID":     "$.steps.1.result.procurement_id",
					"journalEntries":    "$.steps.9.result.journal_entries",
					"completionStatus":  "Completed",
				},
				TimeoutSeconds:    15,
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

			// Step 101: Revert Quality Validation (compensates step 2)
			{
				StepNumber:    101,
				ServiceName:   "quality-inspection",
				HandlerMethod: "RevertProduceQualityValidation",
				InputMapping: map[string]string{
					"procurementID": "$.steps.1.result.procurement_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 102: Cancel Procurement Order (compensates step 3)
			{
				StepNumber:    102,
				ServiceName:   "procurement",
				HandlerMethod: "CancelProcurementOrder",
				InputMapping: map[string]string{
					"procurementID": "$.steps.1.result.procurement_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: Revert Warehouse Receipt (compensates step 4)
			{
				StepNumber:    103,
				ServiceName:   "warehouse",
				HandlerMethod: "RevertWarehouseReceipt",
				InputMapping: map[string]string{
					"procurementID": "$.steps.1.result.procurement_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: Revert Inventory Stock Update (compensates step 5)
			{
				StepNumber:    104,
				ServiceName:   "inventory",
				HandlerMethod: "RevertInventoryStockUpdate",
				InputMapping: map[string]string{
					"procurementID": "$.steps.1.result.procurement_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 105: Revert Receipt Match (compensates step 6)
			{
				StepNumber:    105,
				ServiceName:   "procurement",
				HandlerMethod: "RevertReceiptMatch",
				InputMapping: map[string]string{
					"procurementID": "$.steps.1.result.procurement_id",
				},
				TimeoutSeconds: 25,
				IsCritical:     false,
			},
			// Step 106: Revert Payable Entry (compensates step 7)
			{
				StepNumber:    106,
				ServiceName:   "accounts-payable",
				HandlerMethod: "RevertPayableEntry",
				InputMapping: map[string]string{
					"procurementID": "$.steps.1.result.procurement_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 107: Revert Supply Chain Update (compensates step 8)
			{
				StepNumber:    107,
				ServiceName:   "agriculture",
				HandlerMethod: "RevertSupplyChainUpdate",
				InputMapping: map[string]string{
					"procurementID": "$.steps.1.result.procurement_id",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 108: Reverse Procurement Journal (compensates step 9)
			{
				StepNumber:    108,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseProcurementJournal",
				InputMapping: map[string]string{
					"procurementID": "$.steps.1.result.procurement_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *ProcurementSaga) SagaType() string {
	return "SAGA-A04"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *ProcurementSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *ProcurementSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *ProcurementSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["procurement_id"] == nil {
		return errors.New("procurement_id is required")
	}

	if inputMap["farm_id"] == nil {
		return errors.New("farm_id is required")
	}

	if inputMap["produce_type"] == nil {
		return errors.New("produce_type is required")
	}

	if inputMap["quantity"] == nil {
		return errors.New("quantity is required")
	}

	return nil
}
