// Package platform provides saga handlers for platform module workflows
package platform

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// MasterDataSynchronizationSaga implements SAGA-PLAT03: Master Data Synchronization workflow
// Business Flow: IdentifyMasterDataChanges → ValidateChangeCompliance → UpdateModule1 →
// UpdateModule2 → UpdateModule3 → CreateAuditTrail → SendNotifications → ArchivePreviousVersion
// Purpose: Sync master data across ERP modules, handle updates/changes
type MasterDataSynchronizationSaga struct {
	steps []*saga.StepDefinition
}

// NewMasterDataSynchronizationSaga creates a new Master Data Synchronization saga handler
func NewMasterDataSynchronizationSaga() saga.SagaHandler {
	return &MasterDataSynchronizationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Identify Master Data Changes (customer, vendor, item, GL code)
			{
				StepNumber:    1,
				ServiceName:   "master-data",
				HandlerMethod: "IdentifyMasterDataChanges",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"entityType":        "$.input.entity_type",
					"entityID":          "$.input.entity_id",
					"changeFields":      "$.input.change_fields",
					"changeReason":      "$.input.change_reason",
					"effectiveDate":     "$.input.effective_date",
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
			// Step 2: Validate Change Compliance (blocking periods, approval rules)
			{
				StepNumber:    2,
				ServiceName:   "compliance",
				HandlerMethod: "ValidateMasterDataChangeCompliance",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"entityType":      "$.input.entity_type",
					"entityID":        "$.input.entity_id",
					"changeFields":    "$.input.change_fields",
					"checkBlocking":   "$.input.check_blocking_periods",
					"requireApproval": "$.input.require_approval",
					"approverID":      "$.input.approver_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Update Entity in Module 1 (e.g., customer in AR)
			{
				StepNumber:    3,
				ServiceName:   "sales-invoice",
				HandlerMethod: "UpdateCustomerInAR",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"entityType":      "$.input.entity_type",
					"entityID":        "$.input.entity_id",
					"changeFields":    "$.input.change_fields",
					"previousVersion": "$.steps.1.result.previous_version",
					"effectiveDate":   "$.input.effective_date",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{109},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Update Entity in Module 2 (e.g., customer in CRM)
			{
				StepNumber:    4,
				ServiceName:   "replication",
				HandlerMethod: "UpdateCustomerInCRM",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"entityType":      "$.input.entity_type",
					"entityID":        "$.input.entity_id",
					"changeFields":    "$.input.change_fields",
					"previousVersion": "$.steps.1.result.previous_version",
					"effectiveDate":   "$.input.effective_date",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{110},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Update Entity in Module 3 (e.g., customer in sales-order)
			{
				StepNumber:    5,
				ServiceName:   "master-data",
				HandlerMethod: "UpdateCustomerInSalesOrder",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"entityType":      "$.input.entity_type",
					"entityID":        "$.input.entity_id",
					"changeFields":    "$.input.change_fields",
					"previousVersion": "$.steps.1.result.previous_version",
					"effectiveDate":   "$.input.effective_date",
				},
				TimeoutSeconds: 45,
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
			// Step 6: Create Audit Trail for All Changes
			{
				StepNumber:    6,
				ServiceName:   "audit",
				HandlerMethod: "CreateMasterDataAuditTrail",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"entityType":      "$.input.entity_type",
					"entityID":        "$.input.entity_id",
					"changeFields":    "$.input.change_fields",
					"previousVersion": "$.steps.1.result.previous_version",
					"newVersion":      "$.steps.3.result.updated_entity",
					"changedBy":       "$.input.changed_by",
					"changeDate":      "$.input.change_date",
					"changeReason":    "$.input.change_reason",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{112},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Send Notifications to Affected Users
			{
				StepNumber:    7,
				ServiceName:   "notification",
				HandlerMethod: "SendMasterDataChangeNotifications",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"entityType":      "$.input.entity_type",
					"entityID":        "$.input.entity_id",
					"changeFields":    "$.input.change_fields",
					"affectedUsers":   "$.input.affected_users",
					"notificationTemplate": "$.input.notification_template",
					"effectiveDate":   "$.input.effective_date",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
				CompensationSteps: []int32{113},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Archive Previous Master Data Version
			{
				StepNumber:    8,
				ServiceName:   "master-data",
				HandlerMethod: "ArchivePreviousMasterDataVersion",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"entityType":      "$.input.entity_type",
					"entityID":        "$.input.entity_id",
					"previousVersion": "$.steps.1.result.previous_version",
					"archiveDate":     "$.input.archive_date",
					"retentionYears":  "$.input.retention_years",
				},
				TimeoutSeconds: 60,
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

			// Step 108: Revert Compliance Validation (compensates step 2)
			{
				StepNumber:    108,
				ServiceName:   "compliance",
				HandlerMethod: "RevertMasterDataComplianceValidation",
				InputMapping: map[string]string{
					"tenantID":  "$.tenantID",
					"companyID": "$.companyID",
					"branchID":  "$.branchID",
					"entityID":  "$.input.entity_id",
					"reason":    "Saga compensation - master data sync failed",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 109: Revert Update in Module 1 (compensates step 3)
			{
				StepNumber:    109,
				ServiceName:   "sales-invoice",
				HandlerMethod: "RevertCustomerUpdateInAR",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"entityType":      "$.input.entity_type",
					"entityID":        "$.input.entity_id",
					"previousVersion": "$.steps.1.result.previous_version",
					"reason":          "Saga compensation - master data sync aborted",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 110: Revert Update in Module 2 (compensates step 4)
			{
				StepNumber:    110,
				ServiceName:   "replication",
				HandlerMethod: "RevertCustomerUpdateInCRM",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"entityType":      "$.input.entity_type",
					"entityID":        "$.input.entity_id",
					"previousVersion": "$.steps.1.result.previous_version",
					"reason":          "Saga compensation - master data sync aborted",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 111: Revert Update in Module 3 (compensates step 5)
			{
				StepNumber:    111,
				ServiceName:   "master-data",
				HandlerMethod: "RevertCustomerUpdateInSalesOrder",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"entityType":      "$.input.entity_type",
					"entityID":        "$.input.entity_id",
					"previousVersion": "$.steps.1.result.previous_version",
					"reason":          "Saga compensation - master data sync aborted",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 112: Delete Audit Trail (compensates step 6)
			{
				StepNumber:    112,
				ServiceName:   "audit",
				HandlerMethod: "DeleteMasterDataAuditTrail",
				InputMapping: map[string]string{
					"tenantID":  "$.tenantID",
					"companyID": "$.companyID",
					"branchID":  "$.branchID",
					"entityID":  "$.input.entity_id",
					"reason":    "Saga compensation - master data sync cancelled",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
			// Step 113: Retract Notifications (compensates step 7)
			{
				StepNumber:    113,
				ServiceName:   "notification",
				HandlerMethod: "RetractMasterDataChangeNotifications",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"entityID":        "$.input.entity_id",
					"affectedUsers":   "$.input.affected_users",
					"reason":          "Saga compensation - master data sync cancelled",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *MasterDataSynchronizationSaga) SagaType() string {
	return "SAGA-PLAT03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *MasterDataSynchronizationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *MasterDataSynchronizationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *MasterDataSynchronizationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["entity_type"] == nil {
		return errors.New("entity_type is required (e.g., CUSTOMER, VENDOR, ITEM, GL_CODE)")
	}

	entityType, ok := inputMap["entity_type"].(string)
	if !ok || entityType == "" {
		return errors.New("entity_type must be a non-empty string")
	}

	validEntityTypes := map[string]bool{
		"CUSTOMER": true,
		"VENDOR":   true,
		"ITEM":     true,
		"GL_CODE":  true,
	}
	if !validEntityTypes[entityType] {
		return errors.New("entity_type must be one of: CUSTOMER, VENDOR, ITEM, GL_CODE")
	}

	if inputMap["entity_id"] == nil {
		return errors.New("entity_id is required")
	}

	entityID, ok := inputMap["entity_id"].(string)
	if !ok || entityID == "" {
		return errors.New("entity_id must be a non-empty string")
	}

	if inputMap["change_fields"] == nil {
		return errors.New("change_fields is required")
	}

	changeFields, ok := inputMap["change_fields"].(map[string]interface{})
	if !ok || len(changeFields) == 0 {
		return errors.New("change_fields must be a non-empty map of field names to new values")
	}

	if inputMap["change_reason"] == nil {
		return errors.New("change_reason is required")
	}

	changeReason, ok := inputMap["change_reason"].(string)
	if !ok || changeReason == "" {
		return errors.New("change_reason must be a non-empty string")
	}

	if inputMap["effective_date"] == nil {
		return errors.New("effective_date is required (YYYY-MM-DD format)")
	}

	if inputMap["changed_by"] == nil {
		return errors.New("changed_by is required (user ID)")
	}

	if inputMap["change_date"] == nil {
		return errors.New("change_date is required (YYYY-MM-DD format)")
	}

	if inputMap["affected_users"] == nil {
		return errors.New("affected_users is required")
	}

	affectedUsers, ok := inputMap["affected_users"].([]interface{})
	if !ok || len(affectedUsers) == 0 {
		return errors.New("affected_users must be a non-empty list of user IDs")
	}

	if inputMap["check_blocking_periods"] == nil {
		return errors.New("check_blocking_periods is required (boolean)")
	}

	if inputMap["require_approval"] == nil {
		return errors.New("require_approval is required (boolean)")
	}

	requireApproval, ok := inputMap["require_approval"].(bool)
	if ok && requireApproval {
		if inputMap["approver_id"] == nil {
			return errors.New("approver_id is required when require_approval is true")
		}
	}

	if inputMap["retention_years"] == nil {
		return errors.New("retention_years is required")
	}

	retentionYears, ok := inputMap["retention_years"].(float64)
	if !ok || retentionYears <= 0 || retentionYears > 100 {
		return errors.New("retention_years must be a positive number between 1 and 100")
	}

	return nil
}
