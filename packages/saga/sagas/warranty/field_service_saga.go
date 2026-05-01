// Package warranty provides saga handlers for warranty and service module workflows
package warranty

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// FieldServiceSaga implements SAGA-W02: Field Service workflow
// Business Flow: ReceiveServiceRequest → AssignTechnician → ScheduleTravel → DiagnoseOnSite →
// DeterminePartsRequirement → ExecuteRepair → PerformQualityCheck → IssuePartsFromInventory →
// PostLaborCost → PostMaterialCost → CompleteService → GenerateInvoice → PostGL → NotifyCustomer
// Timeout: 240 seconds, Critical steps: 1,2,3,11,12,13
type FieldServiceSaga struct {
	steps []*saga.StepDefinition
}

// NewFieldServiceSaga creates a new Field Service saga handler (SAGA-W02)
func NewFieldServiceSaga() saga.SagaHandler {
	return &FieldServiceSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Receive Service Request
			{
				StepNumber:    1,
				ServiceName:   "field-service",
				HandlerMethod: "ReceiveServiceRequest",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"customerID":    "$.input.customer_id",
					"assetID":       "$.input.asset_id",
					"issueDescription": "$.input.issue_description",
					"requestDate":   "$.input.request_date",
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
			// Step 2: Assign Technician
			{
				StepNumber:    2,
				ServiceName:   "technician",
				HandlerMethod: "AssignTechnician",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"skillSet":      "$.input.skill_set",
					"serviceRequest": "$.steps.1.result.service_request",
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
			// Step 3: Schedule Travel
			{
				StepNumber:    3,
				ServiceName:   "scheduling",
				HandlerMethod: "ScheduleTravel",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"technicianID":  "$.steps.2.result.technician_id",
					"customerLocation": "$.input.customer_location",
					"scheduledDate": "$.input.scheduled_date",
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
			// Step 4: Diagnose On-Site
			{
				StepNumber:    4,
				ServiceName:   "field-service",
				HandlerMethod: "DiagnoseOnSite",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"assetID":       "$.input.asset_id",
					"travelSchedule": "$.steps.3.result.travel_schedule",
				},
				TimeoutSeconds: 90,
				IsCritical:     false,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Determine Parts Requirement
			{
				StepNumber:    5,
				ServiceName:   "parts-inventory",
				HandlerMethod: "DeterminePartsRequirement",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"assetID":       "$.input.asset_id",
					"diagnosis":     "$.steps.4.result.diagnosis",
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
			// Step 6: Execute Repair
			{
				StepNumber:    6,
				ServiceName:   "work-order",
				HandlerMethod: "ExecuteRepair",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"assetID":       "$.input.asset_id",
					"partsRequirement": "$.steps.5.result.parts_requirement",
					"diagnosis":     "$.steps.4.result.diagnosis",
				},
				TimeoutSeconds: 120,
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
			// Step 7: Perform Quality Check
			{
				StepNumber:    7,
				ServiceName:   "asset",
				HandlerMethod: "PerformQualityCheck",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"assetID":       "$.input.asset_id",
					"repairExecution": "$.steps.6.result.repair_execution",
				},
				TimeoutSeconds: 60,
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
			// Step 8: Issue Parts From Inventory
			{
				StepNumber:    8,
				ServiceName:   "parts-inventory",
				HandlerMethod: "IssuePartsFromInventory",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"partsRequirement": "$.steps.5.result.parts_requirement",
					"qualityCheck":  "$.steps.7.result.quality_check",
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
			// Step 9: Post Labor Cost
			{
				StepNumber:    9,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostLaborCost",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"technicianID":  "$.steps.2.result.technician_id",
					"repairExecution": "$.steps.6.result.repair_execution",
					"laborHours":    "$.input.labor_hours",
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
			// Step 10: Post Material Cost
			{
				StepNumber:    10,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostMaterialCost",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"partsIssuance": "$.steps.8.result.parts_issuance",
					"materialsUsed": "$.input.materials_used",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{110},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 11: Complete Service
			{
				StepNumber:    11,
				ServiceName:   "field-service",
				HandlerMethod: "CompleteService",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"qualityCheck":  "$.steps.7.result.quality_check",
					"laborCostPosting": "$.steps.9.result.journal_entry_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				CompensationSteps: []int32{111},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 12: Generate Invoice
			{
				StepNumber:    12,
				ServiceName:   "sales-invoice",
				HandlerMethod: "GenerateServiceInvoice",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"customerID":    "$.input.customer_id",
					"serviceCompletion": "$.steps.11.result.service_completion",
					"invoiceDate":   "$.input.request_date",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{112},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 13: Post GL (General Ledger)
			{
				StepNumber:    13,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostServiceJournal",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"invoiceID":     "$.steps.12.result.invoice_id",
					"laborCosts":    "$.steps.9.result.labor_cost_amount",
					"materialCosts": "$.steps.10.result.material_cost_amount",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{113},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 14: Notify Customer
			{
				StepNumber:    14,
				ServiceName:   "notification",
				HandlerMethod: "NotifyServiceCompletion",
				InputMapping: map[string]string{
					"tenantID":      "$.tenantID",
					"companyID":     "$.companyID",
					"branchID":      "$.branchID",
					"serviceRequestID": "$.input.service_request_id",
					"customerID":    "$.input.customer_id",
					"invoiceID":     "$.steps.12.result.invoice_id",
					"glPosting":     "$.steps.13.result.journal_entries",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
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

			// Step 102: UnassignTechnician (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "technician",
				HandlerMethod: "UnassignTechnician",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"technicianID": "$.steps.2.result.technician_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 103: CancelTravelSchedule (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "scheduling",
				HandlerMethod: "CancelTravelSchedule",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"travelSchedule": "$.steps.3.result.travel_schedule",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 104: ClearDiagnosis (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "field-service",
				HandlerMethod: "ClearDiagnosis",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"diagnosis": "$.steps.4.result.diagnosis",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 105: CancelPartsRequirement (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "parts-inventory",
				HandlerMethod: "CancelPartsRequirement",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"partsRequirement": "$.steps.5.result.parts_requirement",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 106: ReverseRepairExecution (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "work-order",
				HandlerMethod: "ReverseRepairExecution",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"repairExecution": "$.steps.6.result.repair_execution",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 107: ClearQualityCheck (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "asset",
				HandlerMethod: "ClearQualityCheck",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"qualityCheck": "$.steps.7.result.quality_check",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 108: ReturnPartsToInventory (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "parts-inventory",
				HandlerMethod: "ReturnPartsToInventory",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"partsIssuance": "$.steps.8.result.parts_issuance",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 109: ReverseLaborCostPosting (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseLaborCostJournal",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"journalEntryID": "$.steps.9.result.journal_entry_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 110: ReverseMaterialCostPosting (compensates step 10)
			{
				StepNumber:    110,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseMaterialCostJournal",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"journalEntryID": "$.steps.10.result.journal_entry_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 111: ReverseServiceCompletion (compensates step 11)
			{
				StepNumber:    111,
				ServiceName:   "field-service",
				HandlerMethod: "ReverseServiceCompletion",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"serviceCompletion": "$.steps.11.result.service_completion",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 112: ReverseServiceInvoice (compensates step 12)
			{
				StepNumber:    112,
				ServiceName:   "sales-invoice",
				HandlerMethod: "ReverseServiceInvoice",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"invoiceID": "$.steps.12.result.invoice_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 113: ReverseServiceJournal (compensates step 13)
			{
				StepNumber:    113,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseServiceJournal",
				InputMapping: map[string]string{
					"serviceRequestID": "$.input.service_request_id",
					"journalEntryID": "$.steps.13.result.journal_entry_id",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *FieldServiceSaga) SagaType() string {
	return "SAGA-W02"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *FieldServiceSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *FieldServiceSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *FieldServiceSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["service_request_id"] == nil {
		return errors.New("service_request_id is required")
	}

	serviceRequestID, ok := inputMap["service_request_id"].(string)
	if !ok || serviceRequestID == "" {
		return errors.New("service_request_id must be a non-empty string")
	}

	if inputMap["customer_id"] == nil {
		return errors.New("customer_id is required")
	}

	customerID, ok := inputMap["customer_id"].(string)
	if !ok || customerID == "" {
		return errors.New("customer_id must be a non-empty string")
	}

	if inputMap["asset_id"] == nil {
		return errors.New("asset_id is required")
	}

	assetID, ok := inputMap["asset_id"].(string)
	if !ok || assetID == "" {
		return errors.New("asset_id must be a non-empty string")
	}

	if inputMap["issue_description"] == nil {
		return errors.New("issue_description is required")
	}

	return nil
}
