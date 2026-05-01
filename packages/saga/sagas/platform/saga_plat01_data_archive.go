// Package platform provides saga handlers for platform module workflows
package platform

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// DataArchiveRetentionSaga implements SAGA-PLAT01: Data Archive & Retention Management workflow
// Business Flow: IdentifyArchivableData → ValidateComplianceRequirements → CompressEncryptData →
// CopyToArchiveStorage → VerifyArchiveIntegrity → CreateArchiveIndex → UpdateRetentionStatus →
// PurgeFromOperationalDB → LogArchiveAction
// Purpose: Archive historical data, manage retention policies, support compliance
type DataArchiveRetentionSaga struct {
	steps []*saga.StepDefinition
}

// NewDataArchiveRetentionSaga creates a new Data Archive & Retention saga handler
func NewDataArchiveRetentionSaga() saga.SagaHandler {
	return &DataArchiveRetentionSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Identify Archivable Data (transactions older than X months)
			{
				StepNumber:    1,
				ServiceName:   "archive",
				HandlerMethod: "IdentifyArchivableData",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"archiveOlderThan":  "$.input.archive_older_than_months",
					"dataTypes":         "$.input.data_types",
					"excludeCategories": "$.input.exclude_categories",
				},
				TimeoutSeconds: 90,
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
			// Step 2: Validate Compliance Requirements (legal holds, audit flags)
			{
				StepNumber:    2,
				ServiceName:   "compliance",
				HandlerMethod: "ValidateArchiveCompliance",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"dataRecords":     "$.steps.1.result.records",
					"recordCount":     "$.steps.1.result.record_count",
					"legalHoldCheck":  "$.input.check_legal_holds",
					"auditFlagCheck":  "$.input.check_audit_flags",
					"complianceLevel": "$.input.compliance_level",
				},
				TimeoutSeconds: 60,
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
			// Step 3: Compress and Encrypt Data
			{
				StepNumber:    3,
				ServiceName:   "archive",
				HandlerMethod: "CompressEncryptData",
				InputMapping: map[string]string{
					"dataRecords":       "$.steps.1.result.records",
					"recordCount":       "$.steps.1.result.record_count",
					"compressionMethod": "$.input.compression_method",
					"encryptionMethod":  "$.input.encryption_method",
					"encryptionKey":     "$.input.encryption_key_id",
					"archiveFormat":     "$.input.archive_format",
				},
				TimeoutSeconds: 120,
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
			// Step 4: Copy to Archive Storage (cloud or offline)
			{
				StepNumber:    4,
				ServiceName:   "archive",
				HandlerMethod: "CopyToArchiveStorage",
				InputMapping: map[string]string{
					"compressedData": "$.steps.3.result.compressed_data_path",
					"archiveID":      "$.steps.3.result.archive_id",
					"storageType":    "$.input.storage_type",
					"storageLocation": "$.input.storage_location",
					"redundancyLevel": "$.input.redundancy_level",
					"retentionYears":  "$.input.retention_years",
				},
				TimeoutSeconds: 600,
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
			// Step 5: Verify Archive Integrity (checksums)
			{
				StepNumber:    5,
				ServiceName:   "archive",
				HandlerMethod: "VerifyArchiveIntegrity",
				InputMapping: map[string]string{
					"archiveID":        "$.steps.3.result.archive_id",
					"storagePath":      "$.steps.4.result.storage_path",
					"sourceChecksum":   "$.steps.3.result.source_checksum",
					"verificationMode": "$.input.verification_mode",
					"checksumAlgorithm": "$.input.checksum_algorithm",
				},
				TimeoutSeconds: 120,
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
			// Step 6: Create Archive Index/Manifest
			{
				StepNumber:    6,
				ServiceName:   "archive",
				HandlerMethod: "CreateArchiveManifest",
				InputMapping: map[string]string{
					"archiveID":       "$.steps.3.result.archive_id",
					"dataRecords":     "$.steps.1.result.records",
					"recordCount":     "$.steps.1.result.record_count",
					"integrityStatus": "$.steps.5.result.integrity_verified",
					"archiveDate":     "$.input.archive_date",
				},
				TimeoutSeconds: 60,
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
			// Step 7: Update Retention Policy Status
			{
				StepNumber:    7,
				ServiceName:   "retention-policy",
				HandlerMethod: "UpdateRetentionPolicyStatus",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"archiveID":        "$.steps.3.result.archive_id",
					"recordCount":      "$.steps.1.result.record_count",
					"status":           "ARCHIVED",
					"retentionMonths":  "$.input.retention_months",
					"disposalSchedule": "$.input.disposal_schedule",
				},
				TimeoutSeconds: 45,
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
			// Step 8: Purge from Operational Database
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "PurgeArchivedRecords",
				InputMapping: map[string]string{
					"tenantID":       "$.tenantID",
					"companyID":      "$.companyID",
					"branchID":       "$.branchID",
					"dataRecords":    "$.steps.1.result.records",
					"archiveID":      "$.steps.3.result.archive_id",
					"purgeConfirmed": "$.input.purge_confirmed",
				},
				TimeoutSeconds: 180,
				IsCritical:     true,
				CompensationSteps: []int32{114},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Log Archive Action for Audit
			{
				StepNumber:    9,
				ServiceName:   "notification",
				HandlerMethod: "LogArchiveAuditTrail",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"archiveID":        "$.steps.3.result.archive_id",
					"recordCount":      "$.steps.1.result.record_count",
					"storagePath":      "$.steps.4.result.storage_path",
					"integrityStatus":  "$.steps.5.result.integrity_verified",
					"completionTime":   "$.input.completion_time",
					"initiatedBy":      "$.input.initiated_by",
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

			// Step 108: Reverse Compliance Validation (compensates step 2)
			{
				StepNumber:    108,
				ServiceName:   "compliance",
				HandlerMethod: "ReverseArchiveCompliance",
				InputMapping: map[string]string{
					"tenantID":  "$.tenantID",
					"companyID": "$.companyID",
					"branchID":  "$.branchID",
					"reason":    "Saga compensation - archive failed",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 109: Decrypt and Decompress Data (compensates step 3)
			{
				StepNumber:    109,
				ServiceName:   "archive",
				HandlerMethod: "DecryptDecompressData",
				InputMapping: map[string]string{
					"compressedData":    "$.steps.3.result.compressed_data_path",
					"archiveID":         "$.steps.3.result.archive_id",
					"encryptionKey":     "$.input.encryption_key_id",
					"compressionMethod": "$.input.compression_method",
				},
				TimeoutSeconds: 120,
				IsCritical:     false,
			},
			// Step 110: Remove from Archive Storage (compensates step 4)
			{
				StepNumber:    110,
				ServiceName:   "archive",
				HandlerMethod: "RemoveFromArchiveStorage",
				InputMapping: map[string]string{
					"archiveID":      "$.steps.3.result.archive_id",
					"storagePath":    "$.steps.4.result.storage_path",
					"storageType":    "$.input.storage_type",
					"reason":         "Saga compensation - archive integrity check failed",
				},
				TimeoutSeconds: 180,
				IsCritical:     false,
			},
			// Step 111: Revert Integrity Verification (compensates step 5)
			{
				StepNumber:    111,
				ServiceName:   "archive",
				HandlerMethod: "RevertIntegrityVerification",
				InputMapping: map[string]string{
					"archiveID": "$.steps.3.result.archive_id",
					"reason":    "Saga compensation - integrity verification failed",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 112: Delete Archive Manifest (compensates step 6)
			{
				StepNumber:    112,
				ServiceName:   "archive",
				HandlerMethod: "DeleteArchiveManifest",
				InputMapping: map[string]string{
					"archiveID": "$.steps.3.result.archive_id",
					"reason":    "Saga compensation - archive creation aborted",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 113: Revert Retention Policy Status (compensates step 7)
			{
				StepNumber:    113,
				ServiceName:   "retention-policy",
				HandlerMethod: "RevertRetentionPolicyStatus",
				InputMapping: map[string]string{
					"tenantID":  "$.tenantID",
					"companyID": "$.companyID",
					"branchID":  "$.branchID",
					"archiveID": "$.steps.3.result.archive_id",
					"reason":    "Saga compensation - archive aborted",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 114: Restore to Operational Database (compensates step 8)
			{
				StepNumber:    114,
				ServiceName:   "general-ledger",
				HandlerMethod: "RestoreArchivedRecords",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"archiveID":       "$.steps.3.result.archive_id",
					"decompressedData": "$.steps.109.result.decompressed_data",
					"reason":          "Saga compensation - archive failed",
				},
				TimeoutSeconds: 180,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *DataArchiveRetentionSaga) SagaType() string {
	return "SAGA-PLAT01"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *DataArchiveRetentionSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *DataArchiveRetentionSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *DataArchiveRetentionSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["archive_older_than_months"] == nil {
		return errors.New("archive_older_than_months is required")
	}

	months, ok := inputMap["archive_older_than_months"].(float64)
	if !ok || months <= 0 || months > 120 {
		return errors.New("archive_older_than_months must be a positive number between 1 and 120")
	}

	if inputMap["data_types"] == nil {
		return errors.New("data_types is required")
	}

	dataTypes, ok := inputMap["data_types"].([]interface{})
	if !ok || len(dataTypes) == 0 {
		return errors.New("data_types must be a non-empty list")
	}

	if inputMap["compression_method"] == nil {
		return errors.New("compression_method is required (e.g., GZIP, BZIP2, DEFLATE)")
	}

	if inputMap["encryption_method"] == nil {
		return errors.New("encryption_method is required (e.g., AES-256, RSA)")
	}

	if inputMap["encryption_key_id"] == nil {
		return errors.New("encryption_key_id is required")
	}

	if inputMap["storage_type"] == nil {
		return errors.New("storage_type is required (e.g., CLOUD_S3, CLOUD_AZURE, OFFLINE_TAPE)")
	}

	if inputMap["storage_location"] == nil {
		return errors.New("storage_location is required")
	}

	if inputMap["retention_years"] == nil {
		return errors.New("retention_years is required")
	}

	retentionYears, ok := inputMap["retention_years"].(float64)
	if !ok || retentionYears <= 0 || retentionYears > 100 {
		return errors.New("retention_years must be a positive number between 1 and 100")
	}

	if inputMap["retention_months"] == nil {
		return errors.New("retention_months is required")
	}

	retentionMonths, ok := inputMap["retention_months"].(float64)
	if !ok || retentionMonths <= 0 {
		return errors.New("retention_months must be a positive number")
	}

	if inputMap["purge_confirmed"] == nil {
		return errors.New("purge_confirmed is required (boolean)")
	}

	purgeConfirmed, ok := inputMap["purge_confirmed"].(bool)
	if !ok || !purgeConfirmed {
		return errors.New("purge_confirmed must be explicitly set to true to proceed with archive purge")
	}

	return nil
}
