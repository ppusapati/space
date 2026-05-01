// Package platform provides tests for platform saga handlers
package platform

import (
	"testing"

	"p9e.in/samavaya/packages/saga"
)

// ============================================================================
// SAGA-PLAT01: Data Archive & Retention Management Tests
// ============================================================================

// TestDataArchiveRetentionSagaType verifies saga type identification
func TestDataArchiveRetentionSagaType(t *testing.T) {
	s := NewDataArchiveRetentionSaga()
	expected := "SAGA-PLAT01"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestDataArchiveRetentionSagaStepCount verifies forward and compensation steps
func TestDataArchiveRetentionSagaStepCount(t *testing.T) {
	s := NewDataArchiveRetentionSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 15 // 9 forward + 5 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestDataArchiveRetentionSagaForwardSteps verifies forward step sequence
func TestDataArchiveRetentionSagaForwardSteps(t *testing.T) {
	s := NewDataArchiveRetentionSaga()
	expectedSteps := []int32{1, 2, 3, 4, 5, 6, 7, 8, 9}
	steps := s.GetStepDefinitions()

	forwardSteps := []int32{}
	for _, step := range steps {
		if step.StepNumber <= 9 {
			forwardSteps = append(forwardSteps, step.StepNumber)
		}
	}

	if len(forwardSteps) != len(expectedSteps) {
		t.Errorf("expected %d forward steps, got %d", len(expectedSteps), len(forwardSteps))
	}

	for i, expected := range expectedSteps {
		if i < len(forwardSteps) && forwardSteps[i] != expected {
			t.Errorf("step %d: expected step %d, got %d", i, expected, forwardSteps[i])
		}
	}
}

// TestDataArchiveRetentionSagaValidation verifies input validation
func TestDataArchiveRetentionSagaValidation(t *testing.T) {
	s := NewDataArchiveRetentionSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"archive_older_than_months": 12.0,
				"data_types":                []interface{}{"TRANSACTIONS", "LOGS"},
				"compression_method":        "GZIP",
				"encryption_method":         "AES-256",
				"encryption_key_id":         "KEY001",
				"storage_type":              "CLOUD_S3",
				"storage_location":          "s3://archive-bucket",
				"retention_years":           7.0,
				"retention_months":          84.0,
				"purge_confirmed":           true,
			},
			hasErr: false,
		},
		{
			name: "missing archive_older_than_months",
			input: map[string]interface{}{
				"data_types":         []interface{}{"TRANSACTIONS"},
				"compression_method": "GZIP",
			},
			hasErr: true,
			errMsg: "archive_older_than_months is required",
		},
		{
			name: "invalid archive_older_than_months (zero)",
			input: map[string]interface{}{
				"archive_older_than_months": 0.0,
				"data_types":                []interface{}{"TRANSACTIONS"},
			},
			hasErr: true,
			errMsg: "archive_older_than_months must be a positive number between 1 and 120",
		},
		{
			name: "invalid archive_older_than_months (exceeds max)",
			input: map[string]interface{}{
				"archive_older_than_months": 121.0,
				"data_types":                []interface{}{"TRANSACTIONS"},
			},
			hasErr: true,
			errMsg: "archive_older_than_months must be a positive number between 1 and 120",
		},
		{
			name: "missing data_types",
			input: map[string]interface{}{
				"archive_older_than_months": 12.0,
				"compression_method":        "GZIP",
			},
			hasErr: true,
			errMsg: "data_types is required",
		},
		{
			name: "empty data_types",
			input: map[string]interface{}{
				"archive_older_than_months": 12.0,
				"data_types":                []interface{}{},
				"compression_method":        "GZIP",
			},
			hasErr: true,
			errMsg: "data_types must be a non-empty list",
		},
		{
			name: "missing compression_method",
			input: map[string]interface{}{
				"archive_older_than_months": 12.0,
				"data_types":                []interface{}{"TRANSACTIONS"},
			},
			hasErr: true,
			errMsg: "compression_method is required",
		},
		{
			name: "missing encryption_method",
			input: map[string]interface{}{
				"archive_older_than_months": 12.0,
				"data_types":                []interface{}{"TRANSACTIONS"},
				"compression_method":        "GZIP",
			},
			hasErr: true,
			errMsg: "encryption_method is required",
		},
		{
			name: "missing storage_type",
			input: map[string]interface{}{
				"archive_older_than_months": 12.0,
				"data_types":                []interface{}{"TRANSACTIONS"},
				"compression_method":        "GZIP",
				"encryption_method":         "AES-256",
				"encryption_key_id":         "KEY001",
			},
			hasErr: true,
			errMsg: "storage_type is required",
		},
		{
			name: "invalid retention_years (negative)",
			input: map[string]interface{}{
				"archive_older_than_months": 12.0,
				"data_types":                []interface{}{"TRANSACTIONS"},
				"compression_method":        "GZIP",
				"encryption_method":         "AES-256",
				"encryption_key_id":         "KEY001",
				"storage_type":              "CLOUD_S3",
				"storage_location":          "s3://archive",
				"retention_years":           -1.0,
			},
			hasErr: true,
			errMsg: "retention_years must be a positive number between 1 and 100",
		},
		{
			name: "purge_confirmed not true",
			input: map[string]interface{}{
				"archive_older_than_months": 12.0,
				"data_types":                []interface{}{"TRANSACTIONS"},
				"compression_method":        "GZIP",
				"encryption_method":         "AES-256",
				"encryption_key_id":         "KEY001",
				"storage_type":              "CLOUD_S3",
				"storage_location":          "s3://archive",
				"retention_years":           7.0,
				"retention_months":          84.0,
				"purge_confirmed":           false,
			},
			hasErr: true,
			errMsg: "purge_confirmed must be explicitly set to true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if (err != nil) != tt.hasErr {
				t.Errorf("expected error: %v, got: %v", tt.hasErr, err)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("expected error message: %q, got: %q", tt.errMsg, err.Error())
			}
		})
	}
}

// TestDataArchiveRetentionSagaCriticalSteps verifies critical steps
func TestDataArchiveRetentionSagaCriticalSteps(t *testing.T) {
	s := NewDataArchiveRetentionSaga()
	criticalStepNums := []int32{1, 2, 3, 4, 5, 8}

	for _, criticalNum := range criticalStepNums {
		step := s.GetStepDefinition(int(criticalNum))
		if step == nil {
			t.Errorf("step %d not found", criticalNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be critical", criticalNum)
		}
	}
}

// TestDataArchiveRetentionSagaCompensationChain verifies compensation step linkage
func TestDataArchiveRetentionSagaCompensationChain(t *testing.T) {
	s := NewDataArchiveRetentionSaga()

	tests := []struct {
		stepNum         int32
		expectedCompSteps []int32
	}{
		{stepNum: 2, expectedCompSteps: []int32{108}},
		{stepNum: 3, expectedCompSteps: []int32{109}},
		{stepNum: 4, expectedCompSteps: []int32{110}},
		{stepNum: 5, expectedCompSteps: []int32{111}},
		{stepNum: 6, expectedCompSteps: []int32{112}},
		{stepNum: 7, expectedCompSteps: []int32{113}},
		{stepNum: 8, expectedCompSteps: []int32{114}},
	}

	for _, tt := range tests {
		step := s.GetStepDefinition(int(tt.stepNum))
		if step == nil {
			t.Errorf("step %d not found", tt.stepNum)
			continue
		}
		if len(step.CompensationSteps) != len(tt.expectedCompSteps) {
			t.Errorf("step %d: expected %d compensation steps, got %d",
				tt.stepNum, len(tt.expectedCompSteps), len(step.CompensationSteps))
		}
		for i, expected := range tt.expectedCompSteps {
			if i < len(step.CompensationSteps) && step.CompensationSteps[i] != expected {
				t.Errorf("step %d: expected compensation step %d, got %d",
					tt.stepNum, expected, step.CompensationSteps[i])
			}
		}
	}
}

// TestDataArchiveRetentionSagaTimeouts verifies timeout configurations
func TestDataArchiveRetentionSagaTimeouts(t *testing.T) {
	s := NewDataArchiveRetentionSaga()

	tests := []struct {
		stepNum         int32
		expectedTimeout int32
	}{
		{stepNum: 1, expectedTimeout: 90},
		{stepNum: 3, expectedTimeout: 120},
		{stepNum: 4, expectedTimeout: 600}, // Large data operation
		{stepNum: 8, expectedTimeout: 180}, // Purge operation
	}

	for _, tt := range tests {
		step := s.GetStepDefinition(int(tt.stepNum))
		if step == nil {
			t.Errorf("step %d not found", tt.stepNum)
			continue
		}
		if step.TimeoutSeconds != tt.expectedTimeout {
			t.Errorf("step %d: expected timeout %ds, got %ds",
				tt.stepNum, tt.expectedTimeout, step.TimeoutSeconds)
		}
	}
}

// ============================================================================
// SAGA-PLAT02: Cross-Module Reconciliation & Validation Tests
// ============================================================================

// TestCrossModuleReconciliationSagaType verifies saga type identification
func TestCrossModuleReconciliationSagaType(t *testing.T) {
	s := NewCrossModuleReconciliationSaga()
	expected := "SAGA-PLAT02"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestCrossModuleReconciliationSagaStepCount verifies forward and compensation steps
func TestCrossModuleReconciliationSagaStepCount(t *testing.T) {
	s := NewCrossModuleReconciliationSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 19 // 10 forward + 9 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestCrossModuleReconciliationSagaForwardSteps verifies forward step sequence
func TestCrossModuleReconciliationSagaForwardSteps(t *testing.T) {
	s := NewCrossModuleReconciliationSaga()
	expectedSteps := []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	steps := s.GetStepDefinitions()

	forwardSteps := []int32{}
	for _, step := range steps {
		if step.StepNumber <= 10 {
			forwardSteps = append(forwardSteps, step.StepNumber)
		}
	}

	if len(forwardSteps) != len(expectedSteps) {
		t.Errorf("expected %d forward steps, got %d", len(expectedSteps), len(forwardSteps))
	}
}

// TestCrossModuleReconciliationSagaValidation verifies input validation
func TestCrossModuleReconciliationSagaValidation(t *testing.T) {
	s := NewCrossModuleReconciliationSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"period_start":          "2024-01-01",
				"period_end":            "2024-01-31",
				"account_types":         []interface{}{"ASSET", "LIABILITY"},
				"ar_account_codes":      []interface{}{"1100", "1110"},
				"ap_account_codes":      []interface{}{"2100", "2110"},
				"inventory_account_codes": []interface{}{"1500"},
				"tolerance_amount":      1000.0,
				"tolerance_percent":     0.5,
				"journal_date":          "2024-02-01",
				"post_reconciliation_entries": true,
				"reconciliation_method": "THREE_WAY",
			},
			hasErr: false,
		},
		{
			name: "missing period_start",
			input: map[string]interface{}{
				"period_end": "2024-01-31",
			},
			hasErr: true,
			errMsg: "period_start is required",
		},
		{
			name: "missing period_end",
			input: map[string]interface{}{
				"period_start": "2024-01-01",
			},
			hasErr: true,
			errMsg: "period_end is required",
		},
		{
			name: "missing account_types",
			input: map[string]interface{}{
				"period_start": "2024-01-01",
				"period_end":   "2024-01-31",
			},
			hasErr: true,
			errMsg: "account_types is required",
		},
		{
			name: "empty ar_account_codes",
			input: map[string]interface{}{
				"period_start":    "2024-01-01",
				"period_end":      "2024-01-31",
				"account_types":   []interface{}{"ASSET"},
				"ar_account_codes": []interface{}{},
			},
			hasErr: true,
			errMsg: "ar_account_codes must be a non-empty list",
		},
		{
			name: "invalid tolerance_percent (exceeds 100)",
			input: map[string]interface{}{
				"period_start":          "2024-01-01",
				"period_end":            "2024-01-31",
				"account_types":         []interface{}{"ASSET"},
				"ar_account_codes":      []interface{}{"1100"},
				"ap_account_codes":      []interface{}{"2100"},
				"inventory_account_codes": []interface{}{"1500"},
				"tolerance_amount":      1000.0,
				"tolerance_percent":     150.0,
			},
			hasErr: true,
			errMsg: "tolerance_percent must be a number between 0 and 100",
		},
		{
			name: "missing tolerance_amount",
			input: map[string]interface{}{
				"period_start":          "2024-01-01",
				"period_end":            "2024-01-31",
				"account_types":         []interface{}{"ASSET"},
				"ar_account_codes":      []interface{}{"1100"},
				"ap_account_codes":      []interface{}{"2100"},
				"inventory_account_codes": []interface{}{"1500"},
			},
			hasErr: true,
			errMsg: "tolerance_amount is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if (err != nil) != tt.hasErr {
				t.Errorf("expected error: %v, got: %v", tt.hasErr, err)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("expected error message: %q, got: %q", tt.errMsg, err.Error())
			}
		})
	}
}

// TestCrossModuleReconciliationSagaCriticalSteps verifies critical steps
func TestCrossModuleReconciliationSagaCriticalSteps(t *testing.T) {
	s := NewCrossModuleReconciliationSaga()
	criticalStepNums := []int32{1, 2, 3, 4, 5, 6, 7, 10}

	for _, criticalNum := range criticalStepNums {
		step := s.GetStepDefinition(int(criticalNum))
		if step == nil {
			t.Errorf("step %d not found", criticalNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be critical", criticalNum)
		}
	}
}

// TestCrossModuleReconciliationSagaTimeouts verifies timeout configurations
func TestCrossModuleReconciliationSagaTimeouts(t *testing.T) {
	s := NewCrossModuleReconciliationSaga()

	tests := []struct {
		stepNum         int32
		expectedTimeout int32
	}{
		{stepNum: 1, expectedTimeout: 90},
		{stepNum: 5, expectedTimeout: 60}, // GL vs AR comparison
		{stepNum: 6, expectedTimeout: 60}, // GL vs AP comparison
		{stepNum: 7, expectedTimeout: 75}, // GL vs Inventory comparison
		{stepNum: 10, expectedTimeout: 45}, // Post reconciliation JE
	}

	for _, tt := range tests {
		step := s.GetStepDefinition(int(tt.stepNum))
		if step == nil {
			t.Errorf("step %d not found", tt.stepNum)
			continue
		}
		if step.TimeoutSeconds != tt.expectedTimeout {
			t.Errorf("step %d: expected timeout %ds, got %ds",
				tt.stepNum, tt.expectedTimeout, step.TimeoutSeconds)
		}
	}
}

// TestCrossModuleReconciliationSagaVarianceScenarios verifies variance handling
func TestCrossModuleReconciliationSagaVarianceScenarios(t *testing.T) {
	s := NewCrossModuleReconciliationSaga()

	scenarios := []struct {
		name              string
		toleranceAmount   float64
		tolerancePercent  float64
		shouldBeValid     bool
	}{
		{
			name:             "Zero tolerance",
			toleranceAmount:  0.0,
			tolerancePercent: 0.0,
			shouldBeValid:    true,
		},
		{
			name:             "Moderate tolerance",
			toleranceAmount:  1000.0,
			tolerancePercent: 0.5,
			shouldBeValid:    true,
		},
		{
			name:             "High tolerance",
			toleranceAmount:  50000.0,
			tolerancePercent: 5.0,
			shouldBeValid:    true,
		},
		{
			name:             "Invalid negative tolerance",
			toleranceAmount:  -1000.0,
			tolerancePercent: 0.5,
			shouldBeValid:    false,
		},
	}

	for _, scenario := range scenarios {
		input := map[string]interface{}{
			"period_start":            "2024-01-01",
			"period_end":              "2024-01-31",
			"account_types":           []interface{}{"ASSET"},
			"ar_account_codes":        []interface{}{"1100"},
			"ap_account_codes":        []interface{}{"2100"},
			"inventory_account_codes": []interface{}{"1500"},
			"tolerance_amount":        scenario.toleranceAmount,
			"tolerance_percent":       scenario.tolerancePercent,
			"journal_date":            "2024-02-01",
			"post_reconciliation_entries": true,
			"reconciliation_method":   "THREE_WAY",
		}

		err := s.ValidateInput(input)
		if scenario.shouldBeValid && err != nil {
			t.Errorf("%s: expected no error but got: %v", scenario.name, err)
		}
		if !scenario.shouldBeValid && err == nil {
			t.Errorf("%s: expected error but got none", scenario.name)
		}
	}
}

// ============================================================================
// SAGA-PLAT03: Master Data Synchronization Tests
// ============================================================================

// TestMasterDataSynchronizationSagaType verifies saga type identification
func TestMasterDataSynchronizationSagaType(t *testing.T) {
	s := NewMasterDataSynchronizationSaga()
	expected := "SAGA-PLAT03"
	if s.SagaType() != expected {
		t.Errorf("expected %s, got %s", expected, s.SagaType())
	}
}

// TestMasterDataSynchronizationSagaStepCount verifies forward and compensation steps
func TestMasterDataSynchronizationSagaStepCount(t *testing.T) {
	s := NewMasterDataSynchronizationSaga()
	steps := s.GetStepDefinitions()
	expectedCount := 15 // 8 forward + 6 compensation
	if len(steps) != expectedCount {
		t.Errorf("expected %d steps, got %d", expectedCount, len(steps))
	}
}

// TestMasterDataSynchronizationSagaForwardSteps verifies forward step sequence
func TestMasterDataSynchronizationSagaForwardSteps(t *testing.T) {
	s := NewMasterDataSynchronizationSaga()
	expectedSteps := []int32{1, 2, 3, 4, 5, 6, 7, 8}
	steps := s.GetStepDefinitions()

	forwardSteps := []int32{}
	for _, step := range steps {
		if step.StepNumber <= 8 {
			forwardSteps = append(forwardSteps, step.StepNumber)
		}
	}

	if len(forwardSteps) != len(expectedSteps) {
		t.Errorf("expected %d forward steps, got %d", len(expectedSteps), len(forwardSteps))
	}
}

// TestMasterDataSynchronizationSagaValidation verifies input validation
func TestMasterDataSynchronizationSagaValidation(t *testing.T) {
	s := NewMasterDataSynchronizationSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
		errMsg string
	}{
		{
			name: "valid customer update",
			input: map[string]interface{}{
				"entity_type":            "CUSTOMER",
				"entity_id":              "CUST001",
				"change_fields":          map[string]interface{}{"name": "New Name"},
				"change_reason":          "Customer name correction",
				"effective_date":         "2024-02-01",
				"changed_by":             "USER001",
				"change_date":            "2024-01-31",
				"affected_users":         []interface{}{"USER002", "USER003"},
				"check_blocking_periods": false,
				"require_approval":       false,
				"retention_years":        7.0,
			},
			hasErr: false,
		},
		{
			name: "valid vendor update with approval",
			input: map[string]interface{}{
				"entity_type":            "VENDOR",
				"entity_id":              "VEND001",
				"change_fields":          map[string]interface{}{"payment_terms": "Net 60"},
				"change_reason":          "Payment terms update",
				"effective_date":         "2024-02-01",
				"changed_by":             "USER001",
				"change_date":            "2024-01-31",
				"affected_users":         []interface{}{"USER002"},
				"check_blocking_periods": true,
				"require_approval":       true,
				"approver_id":            "APPROVER001",
				"retention_years":        7.0,
			},
			hasErr: false,
		},
		{
			name: "missing entity_type",
			input: map[string]interface{}{
				"entity_id": "CUST001",
			},
			hasErr: true,
			errMsg: "entity_type is required",
		},
		{
			name: "invalid entity_type",
			input: map[string]interface{}{
				"entity_type": "INVALID",
				"entity_id":   "CUST001",
			},
			hasErr: true,
			errMsg: "entity_type must be one of: CUSTOMER, VENDOR, ITEM, GL_CODE",
		},
		{
			name: "missing entity_id",
			input: map[string]interface{}{
				"entity_type": "CUSTOMER",
			},
			hasErr: true,
			errMsg: "entity_id is required",
		},
		{
			name: "empty change_fields",
			input: map[string]interface{}{
				"entity_type":   "CUSTOMER",
				"entity_id":     "CUST001",
				"change_fields": map[string]interface{}{},
			},
			hasErr: true,
			errMsg: "change_fields must be a non-empty map",
		},
		{
			name: "missing change_reason",
			input: map[string]interface{}{
				"entity_type":   "CUSTOMER",
				"entity_id":     "CUST001",
				"change_fields": map[string]interface{}{"name": "New Name"},
			},
			hasErr: true,
			errMsg: "change_reason is required",
		},
		{
			name: "empty affected_users",
			input: map[string]interface{}{
				"entity_type":     "CUSTOMER",
				"entity_id":       "CUST001",
				"change_fields":   map[string]interface{}{"name": "New Name"},
				"change_reason":   "Update",
				"affected_users":  []interface{}{},
			},
			hasErr: true,
			errMsg: "affected_users must be a non-empty list",
		},
		{
			name: "require_approval but no approver_id",
			input: map[string]interface{}{
				"entity_type":            "CUSTOMER",
				"entity_id":              "CUST001",
				"change_fields":          map[string]interface{}{"name": "New Name"},
				"change_reason":          "Update",
				"effective_date":         "2024-02-01",
				"changed_by":             "USER001",
				"change_date":            "2024-01-31",
				"affected_users":         []interface{}{"USER002"},
				"check_blocking_periods": false,
				"require_approval":       true,
				"retention_years":        7.0,
			},
			hasErr: true,
			errMsg: "approver_id is required when require_approval is true",
		},
		{
			name: "invalid retention_years (negative)",
			input: map[string]interface{}{
				"entity_type":            "CUSTOMER",
				"entity_id":              "CUST001",
				"change_fields":          map[string]interface{}{"name": "New Name"},
				"change_reason":          "Update",
				"effective_date":         "2024-02-01",
				"changed_by":             "USER001",
				"change_date":            "2024-01-31",
				"affected_users":         []interface{}{"USER002"},
				"check_blocking_periods": false,
				"require_approval":       false,
				"retention_years":        -1.0,
			},
			hasErr: true,
			errMsg: "retention_years must be a positive number between 1 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if (err != nil) != tt.hasErr {
				t.Errorf("expected error: %v, got: %v", tt.hasErr, err)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("expected error message: %q, got: %q", tt.errMsg, err.Error())
			}
		})
	}
}

// TestMasterDataSynchronizationSagaCriticalSteps verifies critical steps
func TestMasterDataSynchronizationSagaCriticalSteps(t *testing.T) {
	s := NewMasterDataSynchronizationSaga()
	criticalStepNums := []int32{1, 2, 3, 4, 5}

	for _, criticalNum := range criticalStepNums {
		step := s.GetStepDefinition(int(criticalNum))
		if step == nil {
			t.Errorf("step %d not found", criticalNum)
			continue
		}
		if !step.IsCritical {
			t.Errorf("step %d should be critical", criticalNum)
		}
	}
}

// TestMasterDataSynchronizationSagaEntityTypes verifies entity type support
func TestMasterDataSynchronizationSagaEntityTypes(t *testing.T) {
	s := NewMasterDataSynchronizationSaga()

	entityTypes := []string{"CUSTOMER", "VENDOR", "ITEM", "GL_CODE"}

	for _, entityType := range entityTypes {
		input := map[string]interface{}{
			"entity_type":            entityType,
			"entity_id":              "ID001",
			"change_fields":          map[string]interface{}{"field": "value"},
			"change_reason":          "Test",
			"effective_date":         "2024-02-01",
			"changed_by":             "USER001",
			"change_date":            "2024-01-31",
			"affected_users":         []interface{}{"USER002"},
			"check_blocking_periods": false,
			"require_approval":       false,
			"retention_years":        7.0,
		}

		err := s.ValidateInput(input)
		if err != nil {
			t.Errorf("entity_type %s should be valid but got error: %v", entityType, err)
		}
	}
}

// TestMasterDataSynchronizationSagaCompensationChain verifies compensation step linkage
func TestMasterDataSynchronizationSagaCompensationChain(t *testing.T) {
	s := NewMasterDataSynchronizationSaga()

	tests := []struct {
		stepNum           int32
		expectedCompSteps []int32
	}{
		{stepNum: 2, expectedCompSteps: []int32{108}},
		{stepNum: 3, expectedCompSteps: []int32{109}},
		{stepNum: 4, expectedCompSteps: []int32{110}},
		{stepNum: 5, expectedCompSteps: []int32{111}},
		{stepNum: 6, expectedCompSteps: []int32{112}},
		{stepNum: 7, expectedCompSteps: []int32{113}},
	}

	for _, tt := range tests {
		step := s.GetStepDefinition(int(tt.stepNum))
		if step == nil {
			t.Errorf("step %d not found", tt.stepNum)
			continue
		}
		if len(step.CompensationSteps) != len(tt.expectedCompSteps) {
			t.Errorf("step %d: expected %d compensation steps, got %d",
				tt.stepNum, len(tt.expectedCompSteps), len(step.CompensationSteps))
		}
	}
}

// TestMasterDataSynchronizationSagaRetryConfig verifies retry configurations
func TestMasterDataSynchronizationSagaRetryConfig(t *testing.T) {
	s := NewMasterDataSynchronizationSaga()
	steps := s.GetStepDefinitions()

	for _, step := range steps {
		if step.RetryConfig == nil {
			t.Errorf("step %d has no retry config", step.StepNumber)
			continue
		}

		if step.RetryConfig.MaxRetries != 3 {
			t.Errorf("step %d: expected MaxRetries=3, got %d", step.StepNumber, step.RetryConfig.MaxRetries)
		}

		if step.RetryConfig.InitialBackoffMs != 1000 {
			t.Errorf("step %d: expected InitialBackoffMs=1000, got %d", step.StepNumber, step.RetryConfig.InitialBackoffMs)
		}

		if step.RetryConfig.MaxBackoffMs != 30000 {
			t.Errorf("step %d: expected MaxBackoffMs=30000, got %d", step.StepNumber, step.RetryConfig.MaxBackoffMs)
		}

		if step.RetryConfig.BackoffMultiplier != 2.0 {
			t.Errorf("step %d: expected BackoffMultiplier=2.0, got %v", step.StepNumber, step.RetryConfig.BackoffMultiplier)
		}

		if step.RetryConfig.JitterFraction != 0.1 {
			t.Errorf("step %d: expected JitterFraction=0.1, got %v", step.StepNumber, step.RetryConfig.JitterFraction)
		}
	}
}

// ============================================================================
// Cross-Saga Tests
// ============================================================================

// TestPlatformSagasRegistration verifies all platform sagas can be instantiated
func TestPlatformSagasRegistration(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewDataArchiveRetentionSaga(),
		NewCrossModuleReconciliationSaga(),
		NewMasterDataSynchronizationSaga(),
	}

	expectedTypes := map[string]bool{
		"SAGA-PLAT01": true,
		"SAGA-PLAT02": true,
		"SAGA-PLAT03": true,
	}

	for _, s := range sagas {
		if !expectedTypes[s.SagaType()] {
			t.Errorf("unexpected saga type: %s", s.SagaType())
		}
	}
}

// TestPlatformSagasStepDefinitionIntegrity verifies step definition structure
func TestPlatformSagasStepDefinitionIntegrity(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewDataArchiveRetentionSaga(),
		NewCrossModuleReconciliationSaga(),
		NewMasterDataSynchronizationSaga(),
	}

	for _, s := range sagas {
		steps := s.GetStepDefinitions()
		if len(steps) == 0 {
			t.Errorf("saga %s has no steps", s.SagaType())
			continue
		}

		for _, step := range steps {
			if step.StepNumber <= 0 {
				t.Errorf("saga %s step has invalid step number: %d", s.SagaType(), step.StepNumber)
			}

			if step.ServiceName == "" {
				t.Errorf("saga %s step %d has no service name", s.SagaType(), step.StepNumber)
			}

			if step.HandlerMethod == "" {
				t.Errorf("saga %s step %d has no handler method", s.SagaType(), step.StepNumber)
			}

			if len(step.InputMapping) == 0 {
				t.Errorf("saga %s step %d has no input mapping", s.SagaType(), step.StepNumber)
			}

			if step.TimeoutSeconds <= 0 {
				t.Errorf("saga %s step %d has invalid timeout", s.SagaType(), step.StepNumber)
			}

			if step.RetryConfig == nil {
				t.Errorf("saga %s step %d has no retry config", s.SagaType(), step.StepNumber)
			}
		}
	}
}

// TestPlatformSagasGetStepDefinitionLookup verifies step lookup mechanism
func TestPlatformSagasGetStepDefinitionLookup(t *testing.T) {
	sagas := map[string]saga.SagaHandler{
		"SAGA-PLAT01": NewDataArchiveRetentionSaga(),
		"SAGA-PLAT02": NewCrossModuleReconciliationSaga(),
		"SAGA-PLAT03": NewMasterDataSynchronizationSaga(),
	}

	for sagaType, s := range sagas {
		steps := s.GetStepDefinitions()

		for _, expectedStep := range steps {
			lookedUpStep := s.GetStepDefinition(int(expectedStep.StepNumber))
			if lookedUpStep == nil {
				t.Errorf("saga %s: step %d lookup returned nil", sagaType, expectedStep.StepNumber)
			} else if lookedUpStep.StepNumber != expectedStep.StepNumber {
				t.Errorf("saga %s: step lookup mismatch for step %d", sagaType, expectedStep.StepNumber)
			}
		}

		// Test lookup of non-existent step
		nonExistentStep := s.GetStepDefinition(9999)
		if nonExistentStep != nil {
			t.Errorf("saga %s: expected nil for non-existent step, got step %d", sagaType, nonExistentStep.StepNumber)
		}
	}
}

// TestDataArchiveInputValidationEdgeCases verifies archive saga input validation edge cases
func TestDataArchiveInputValidationEdgeCases(t *testing.T) {
	s := NewDataArchiveRetentionSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name:   "invalid input type (string)",
			input:  "not a map",
			hasErr: true,
		},
		{
			name:   "invalid input type (array)",
			input:  []interface{}{1, 2, 3},
			hasErr: true,
		},
		{
			name:   "nil input",
			input:  nil,
			hasErr: true,
		},
		{
			name: "minimum valid months",
			input: map[string]interface{}{
				"archive_older_than_months": 1.0,
				"data_types":                []interface{}{"TRANSACTIONS"},
				"compression_method":        "GZIP",
				"encryption_method":         "AES-256",
				"encryption_key_id":         "KEY001",
				"storage_type":              "CLOUD_S3",
				"storage_location":          "s3://archive",
				"retention_years":           1.0,
				"retention_months":          12.0,
				"purge_confirmed":           true,
			},
			hasErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if (err != nil) != tt.hasErr {
				t.Errorf("expected error: %v, got: %v", tt.hasErr, err)
			}
		})
	}
}

// TestReconciliationInputValidationEdgeCases verifies reconciliation saga input validation edge cases
func TestReconciliationInputValidationEdgeCases(t *testing.T) {
	s := NewCrossModuleReconciliationSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name:   "invalid input type (string)",
			input:  "not a map",
			hasErr: true,
		},
		{
			name:   "nil input",
			input:  nil,
			hasErr: true,
		},
		{
			name: "zero tolerance values",
			input: map[string]interface{}{
				"period_start":            "2024-01-01",
				"period_end":              "2024-01-31",
				"account_types":           []interface{}{"ASSET"},
				"ar_account_codes":        []interface{}{"1100"},
				"ap_account_codes":        []interface{}{"2100"},
				"inventory_account_codes": []interface{}{"1500"},
				"tolerance_amount":        0.0,
				"tolerance_percent":       0.0,
				"journal_date":            "2024-02-01",
				"post_reconciliation_entries": true,
				"reconciliation_method":   "THREE_WAY",
			},
			hasErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if (err != nil) != tt.hasErr {
				t.Errorf("expected error: %v, got: %v", tt.hasErr, err)
			}
		})
	}
}

// TestMasterDataInputValidationEdgeCases verifies master data saga input validation edge cases
func TestMasterDataInputValidationEdgeCases(t *testing.T) {
	s := NewMasterDataSynchronizationSaga()

	tests := []struct {
		name   string
		input  interface{}
		hasErr bool
	}{
		{
			name:   "invalid input type (string)",
			input:  "not a map",
			hasErr: true,
		},
		{
			name:   "nil input",
			input:  nil,
			hasErr: true,
		},
		{
			name: "all valid entity types",
			input: map[string]interface{}{
				"entity_type":            "GL_CODE",
				"entity_id":              "1000",
				"change_fields":          map[string]interface{}{"description": "New GL Code"},
				"change_reason":          "GL update",
				"effective_date":         "2024-02-01",
				"changed_by":             "USER001",
				"change_date":            "2024-01-31",
				"affected_users":         []interface{}{"USER002"},
				"check_blocking_periods": false,
				"require_approval":       false,
				"retention_years":        7.0,
			},
			hasErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateInput(tt.input)
			if (err != nil) != tt.hasErr {
				t.Errorf("expected error: %v, got: %v", tt.hasErr, err)
			}
		})
	}
}
