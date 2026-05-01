// Package manufacturing provides saga handlers for manufacturing module workflows
package manufacturing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"p9e.in/samavaya/packages/saga"
)

// ===== SAGA-M07: Job Costing & Overhead Allocation Tests =====

func TestJobCostingSaga_SagaType(t *testing.T) {
	s := NewJobCostingSaga()
	assert.Equal(t, "SAGA-M07", s.SagaType())
}

func TestJobCostingSaga_StepCount(t *testing.T) {
	s := NewJobCostingSaga()
	steps := s.GetStepDefinitions()
	// 11 forward steps + 3 compensation steps (110, 111, 112)
	assert.Equal(t, 14, len(steps))
}

func TestJobCostingSaga_ForwardSteps(t *testing.T) {
	s := NewJobCostingSaga()
	steps := s.GetStepDefinitions()

	forwardSteps := 0
	for _, step := range steps {
		if step.StepNumber >= 1 && step.StepNumber <= 11 {
			forwardSteps++
		}
	}
	assert.Equal(t, 11, forwardSteps)
}

func TestJobCostingSaga_CompensationSteps(t *testing.T) {
	s := NewJobCostingSaga()
	steps := s.GetStepDefinitions()

	compensationSteps := 0
	for _, step := range steps {
		if step.StepNumber >= 100 {
			compensationSteps++
		}
	}
	assert.Equal(t, 3, compensationSteps)
}

func TestJobCostingSaga_GetStepDefinition(t *testing.T) {
	s := NewJobCostingSaga()

	// Test forward step
	step1 := s.GetStepDefinition(1)
	assert.NotNil(t, step1)
	assert.Equal(t, int32(1), step1.StepNumber)
	assert.Equal(t, "production-order", step1.ServiceName)
	assert.Equal(t, "IdentifyJobForCosting", step1.HandlerMethod)
	assert.True(t, step1.IsCritical)

	// Test middle step
	step6 := s.GetStepDefinition(6)
	assert.NotNil(t, step6)
	assert.Equal(t, int32(6), step6.StepNumber)
	assert.Equal(t, "cost-center", step6.ServiceName)
	assert.True(t, step6.IsCritical)

	// Test compensation step
	step110 := s.GetStepDefinition(110)
	assert.NotNil(t, step110)
	assert.Equal(t, int32(110), step110.StepNumber)
	assert.Equal(t, "general-ledger", step110.ServiceName)
	assert.False(t, step110.IsCritical)

	// Test non-existent step
	stepNone := s.GetStepDefinition(999)
	assert.Nil(t, stepNone)
}

func TestJobCostingSaga_CriticalSteps(t *testing.T) {
	s := NewJobCostingSaga()
	steps := s.GetStepDefinitions()

	criticalStepNumbers := []int{1, 6, 8, 10, 11}
	criticalCount := 0

	for _, step := range steps {
		if step.StepNumber >= 1 && step.StepNumber <= 11 && step.IsCritical {
			criticalCount++
		}
	}
	assert.Equal(t, len(criticalStepNumbers), criticalCount)
}

func TestJobCostingSaga_ValidateInput_Success(t *testing.T) {
	s := NewJobCostingSaga()

	input := map[string]interface{}{
		"production_order_id": "PO-001",
		"job_id":             "JOB-001",
		"costing_period":     "2026-02",
		"cost_center_id":     "CC-001",
		"allocation_base":    "MATERIAL",
		"wip_account_code":   "12100",
	}

	err := s.ValidateInput(input)
	assert.NoError(t, err)
}

func TestJobCostingSaga_ValidateInput_MissingProductionOrderID(t *testing.T) {
	s := NewJobCostingSaga()

	input := map[string]interface{}{
		"job_id":            "JOB-001",
		"costing_period":    "2026-02",
		"cost_center_id":    "CC-001",
		"allocation_base":   "MATERIAL",
		"wip_account_code":  "12100",
	}

	err := s.ValidateInput(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "production_order_id")
}

func TestJobCostingSaga_ValidateInput_MissingJobID(t *testing.T) {
	s := NewJobCostingSaga()

	input := map[string]interface{}{
		"production_order_id": "PO-001",
		"costing_period":      "2026-02",
		"cost_center_id":      "CC-001",
		"allocation_base":     "MATERIAL",
		"wip_account_code":    "12100",
	}

	err := s.ValidateInput(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job_id")
}

func TestJobCostingSaga_ValidateInput_InvalidType(t *testing.T) {
	s := NewJobCostingSaga()

	err := s.ValidateInput("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input type")
}

func TestJobCostingSaga_Step8PostingCritical(t *testing.T) {
	s := NewJobCostingSaga()
	step8 := s.GetStepDefinition(8)

	assert.NotNil(t, step8)
	assert.Equal(t, int32(8), step8.StepNumber)
	assert.True(t, step8.IsCritical)
	assert.Equal(t, "general-ledger", step8.ServiceName)
	assert.Equal(t, "PostJobCostToGL", step8.HandlerMethod)
	assert.Contains(t, step8.CompensationSteps, int32(110))
}

func TestJobCostingSaga_RetryConfig(t *testing.T) {
	s := NewJobCostingSaga()
	step1 := s.GetStepDefinition(1)

	assert.NotNil(t, step1.RetryConfig)
	assert.Equal(t, int32(3), step1.RetryConfig.MaxRetries)
	assert.Equal(t, int32(1000), step1.RetryConfig.InitialBackoffMs)
	assert.Equal(t, int32(30000), step1.RetryConfig.MaxBackoffMs)
	assert.Equal(t, 2.0, step1.RetryConfig.BackoffMultiplier)
	assert.Equal(t, 0.1, step1.RetryConfig.JitterFraction)
}

// ===== SAGA-M08: Cost Variance Analysis Tests =====

func TestCostVarianceAnalysisSaga_SagaType(t *testing.T) {
	s := NewCostVarianceAnalysisSaga()
	assert.Equal(t, "SAGA-M08", s.SagaType())
}

func TestCostVarianceAnalysisSaga_StepCount(t *testing.T) {
	s := NewCostVarianceAnalysisSaga()
	steps := s.GetStepDefinitions()
	// 10 forward steps + 5 compensation steps (110-114)
	assert.Equal(t, 15, len(steps))
}

func TestCostVarianceAnalysisSaga_CriticalSteps(t *testing.T) {
	s := NewCostVarianceAnalysisSaga()
	steps := s.GetStepDefinitions()

	criticalStepNumbers := []int32{1, 2, 3, 4, 6, 9}
	criticalCount := 0

	for _, step := range steps {
		if step.IsCritical && step.StepNumber <= 10 {
			criticalCount++
			assert.Contains(t, criticalStepNumbers, step.StepNumber)
		}
	}
	assert.Equal(t, len(criticalStepNumbers), criticalCount)
}

func TestCostVarianceAnalysisSaga_ValidateInput_Success(t *testing.T) {
	s := NewCostVarianceAnalysisSaga()

	input := map[string]interface{}{
		"product_id":      "PROD-001",
		"job_id":          "JOB-001",
		"cost_center_id":  "CC-001",
		"costing_period":  "2026-02",
		"bom_version":     "V1.0",
	}

	err := s.ValidateInput(input)
	assert.NoError(t, err)
}

func TestCostVarianceAnalysisSaga_ValidateInput_MissingFields(t *testing.T) {
	s := NewCostVarianceAnalysisSaga()

	input := map[string]interface{}{
		"job_id":         "JOB-001",
		"cost_center_id": "CC-001",
		"costing_period": "2026-02",
	}

	err := s.ValidateInput(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product_id")
}

func TestCostVarianceAnalysisSaga_VarianceCalcSequence(t *testing.T) {
	s := NewCostVarianceAnalysisSaga()

	// Steps 3, 4, 5 should calculate material, labor, overhead variance
	step3 := s.GetStepDefinition(3)
	assert.Equal(t, "CalculateMaterialVariance", step3.HandlerMethod)

	step4 := s.GetStepDefinition(4)
	assert.Equal(t, "CalculateLaborVariance", step4.HandlerMethod)

	step5 := s.GetStepDefinition(5)
	assert.Equal(t, "CalculateOverheadVariance", step5.HandlerMethod)
}

// ===== SAGA-M09: Scrap & Rework Management Tests =====

func TestScrapReworkManagementSaga_SagaType(t *testing.T) {
	s := NewScrapReworkManagementSaga()
	assert.Equal(t, "SAGA-M09", s.SagaType())
}

func TestScrapReworkManagementSaga_StepCount(t *testing.T) {
	s := NewScrapReworkManagementSaga()
	steps := s.GetStepDefinitions()
	// 9 forward steps + 4 compensation steps (109-112)
	assert.Equal(t, 13, len(steps))
}

func TestScrapReworkManagementSaga_ReworkDecisionLogic(t *testing.T) {
	s := NewScrapReworkManagementSaga()

	// Step 1: Identify scrap
	step1 := s.GetStepDefinition(1)
	assert.NotNil(t, step1)
	assert.Equal(t, "quality-production", step1.ServiceName)

	// Step 2: Determine if rework or scrap
	step2 := s.GetStepDefinition(2)
	assert.NotNil(t, step2)
	assert.Equal(t, "quality-production", step2.ServiceName)
	assert.Equal(t, "DetermineReworkOrScrap", step2.HandlerMethod)

	// Step 3: Route for rework (conditional)
	step3 := s.GetStepDefinition(3)
	assert.NotNil(t, step3)
	assert.Equal(t, "production-order", step3.ServiceName)

	// Step 4: Remove from inventory
	step4 := s.GetStepDefinition(4)
	assert.NotNil(t, step4)
	assert.Equal(t, "inventory-core", step4.ServiceName)
	assert.True(t, step4.IsCritical)
}

func TestScrapReworkManagementSaga_ValidateInput_Success(t *testing.T) {
	s := NewScrapReworkManagementSaga()

	input := map[string]interface{}{
		"production_order_id":  "PO-001",
		"scrap_quantity":       100,
		"scrap_reason_code":    "QRJ001",
		"quality_inspection_id": "QI-001",
		"unit_cost":            50.00,
	}

	err := s.ValidateInput(input)
	assert.NoError(t, err)
}

func TestScrapReworkManagementSaga_ValidateInput_MissingScrapReason(t *testing.T) {
	s := NewScrapReworkManagementSaga()

	input := map[string]interface{}{
		"production_order_id":   "PO-001",
		"scrap_quantity":        100,
		"quality_inspection_id": "QI-001",
		"unit_cost":             50.00,
	}

	err := s.ValidateInput(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scrap_reason_code")
}

func TestScrapReworkManagementSaga_CriticalSteps(t *testing.T) {
	s := NewScrapReworkManagementSaga()
	steps := s.GetStepDefinitions()

	criticalStepNumbers := []int32{1, 2, 4, 6, 8}
	criticalCount := 0

	for _, step := range steps {
		if step.StepNumber <= 9 && step.IsCritical {
			criticalCount++
			assert.Contains(t, criticalStepNumbers, step.StepNumber)
		}
	}
	assert.Equal(t, len(criticalStepNumbers), criticalCount)
}

// ===== SAGA-M10: Subcontracting Cost Tracking Tests =====

func TestSubcontractingCostTrackingSaga_SagaType(t *testing.T) {
	s := NewSubcontractingCostTrackingSaga()
	assert.Equal(t, "SAGA-M10", s.SagaType())
}

func TestSubcontractingCostTrackingSaga_StepCount(t *testing.T) {
	s := NewSubcontractingCostTrackingSaga()
	steps := s.GetStepDefinitions()
	// 10 forward steps + 5 compensation steps (110-114)
	assert.Equal(t, 15, len(steps))
}

func TestSubcontractingCostTrackingSaga_ValidateInput_Success(t *testing.T) {
	s := NewSubcontractingCostTrackingSaga()

	input := map[string]interface{}{
		"production_order_id": "PO-001",
		"vendor_id":           "V-001",
		"product_id":          "PROD-001",
		"po_id":               "SCPO-001",
		"work_order_quantity": 500,
	}

	err := s.ValidateInput(input)
	assert.NoError(t, err)
}

func TestSubcontractingCostTrackingSaga_ValidateInput_MissingVendor(t *testing.T) {
	s := NewSubcontractingCostTrackingSaga()

	input := map[string]interface{}{
		"production_order_id": "PO-001",
		"product_id":          "PROD-001",
		"po_id":               "SCPO-001",
		"work_order_quantity": 500,
	}

	err := s.ValidateInput(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "vendor_id")
}

func TestSubcontractingCostTrackingSaga_CostCalculationSequence(t *testing.T) {
	s := NewSubcontractingCostTrackingSaga()

	// Step 2: Extract unit cost
	step2 := s.GetStepDefinition(2)
	assert.Equal(t, "purchase-invoice", step2.ServiceName)
	assert.Equal(t, "ExtractUnitCostFromPO", step2.HandlerMethod)

	// Step 3: Calculate total cost
	step3 := s.GetStepDefinition(3)
	assert.Equal(t, "cost-center", step3.ServiceName)
	assert.Equal(t, "CalculateTotalSubcontractingCost", step3.HandlerMethod)
}

func TestSubcontractingCostTrackingSaga_InvoiceMatching(t *testing.T) {
	s := NewSubcontractingCostTrackingSaga()

	// Step 5: Match invoice
	step5 := s.GetStepDefinition(5)
	assert.NotNil(t, step5)
	assert.Equal(t, int32(5), step5.StepNumber)
	assert.Equal(t, "purchase-invoice", step5.ServiceName)
	assert.Equal(t, "MatchSubcontractInvoiceWithOrder", step5.HandlerMethod)
	assert.True(t, step5.IsCritical)
}

// ===== SAGA-M11: Batch/Lot Costing & Traceability Tests =====

func TestBatchCostingTraceabilitySaga_SagaType(t *testing.T) {
	s := NewBatchCostingTraceabilitySaga()
	assert.Equal(t, "SAGA-M11", s.SagaType())
}

func TestBatchCostingTraceabilitySaga_StepCount(t *testing.T) {
	s := NewBatchCostingTraceabilitySaga()
	steps := s.GetStepDefinitions()
	// 9 forward steps + 5 compensation steps (109-113)
	assert.Equal(t, 14, len(steps))
}

func TestBatchCostingTraceabilitySaga_ValidateInput_Success(t *testing.T) {
	s := NewBatchCostingTraceabilitySaga()

	input := map[string]interface{}{
		"production_order_id": "PO-001",
		"product_id":          "PROD-001",
		"creation_date":       "2026-02-16",
		"bom_version":         "V1.0",
		"cost_center_id":      "CC-001",
	}

	err := s.ValidateInput(input)
	assert.NoError(t, err)
}

func TestBatchCostingTraceabilitySaga_ValidateInput_MissingCreationDate(t *testing.T) {
	s := NewBatchCostingTraceabilitySaga()

	input := map[string]interface{}{
		"production_order_id": "PO-001",
		"product_id":          "PROD-001",
		"bom_version":         "V1.0",
		"cost_center_id":      "CC-001",
	}

	err := s.ValidateInput(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creation_date")
}

func TestBatchCostingTraceabilitySaga_BatchIDCreation(t *testing.T) {
	s := NewBatchCostingTraceabilitySaga()

	// Step 1: Create batch ID
	step1 := s.GetStepDefinition(1)
	assert.Equal(t, int32(1), step1.StepNumber)
	assert.Equal(t, "inventory-core", step1.ServiceName)
	assert.Equal(t, "CreateBatchID", step1.HandlerMethod)
	assert.True(t, step1.IsCritical)
}

func TestBatchCostingTraceabilitySaga_GenealogyTracking(t *testing.T) {
	s := NewBatchCostingTraceabilitySaga()

	// Step 5: Update genealogy
	step5 := s.GetStepDefinition(5)
	assert.Equal(t, int32(5), step5.StepNumber)
	assert.Equal(t, "inventory-core", step5.ServiceName)
	assert.Equal(t, "UpdateBatchGenealogy", step5.HandlerMethod)
	assert.Contains(t, step5.CompensationSteps, int32(110))
}

func TestBatchCostingTraceabilitySaga_ClosingSequence(t *testing.T) {
	s := NewBatchCostingTraceabilitySaga()

	// Step 8: Archive traceability
	step8 := s.GetStepDefinition(8)
	assert.Equal(t, "ArchiveBatchTraceabilityChain", step8.HandlerMethod)
	assert.True(t, step8.IsCritical)

	// Step 9: Close accounting
	step9 := s.GetStepDefinition(9)
	assert.Equal(t, "CloseBatchAccounting", step9.HandlerMethod)
	assert.True(t, step9.IsCritical)
}

// ===== SAGA-M12: MRP & Lot Sizing Optimization Tests =====

func TestMRPLotSizingOptimizationSaga_SagaType(t *testing.T) {
	s := NewMRPLotSizingOptimizationSaga()
	assert.Equal(t, "SAGA-M12", s.SagaType())
}

func TestMRPLotSizingOptimizationSaga_StepCount(t *testing.T) {
	s := NewMRPLotSizingOptimizationSaga()
	steps := s.GetStepDefinitions()
	// 10 forward steps + 7 compensation steps (110-116)
	assert.Equal(t, 17, len(steps))
}

func TestMRPLotSizingOptimizationSaga_ValidateInput_Success(t *testing.T) {
	s := NewMRPLotSizingOptimizationSaga()

	input := map[string]interface{}{
		"planning_horizon":     "2026-02-16T2026-03-31T",
		"product_id":           "PROD-001",
		"demand_quantity":      1000,
		"demand_date":          "2026-03-15",
		"lot_sizing_method":    "EOQ",
	}

	err := s.ValidateInput(input)
	assert.NoError(t, err)
}

func TestMRPLotSizingOptimizationSaga_ValidateInput_MissingLotSizingMethod(t *testing.T) {
	s := NewMRPLotSizingOptimizationSaga()

	input := map[string]interface{}{
		"planning_horizon": "2026-02-16T2026-03-31T",
		"product_id":       "PROD-001",
		"demand_quantity":  1000,
		"demand_date":      "2026-03-15",
	}

	err := s.ValidateInput(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lot_sizing_method")
}

func TestMRPLotSizingOptimizationSaga_MRPSequence(t *testing.T) {
	s := NewMRPLotSizingOptimizationSaga()

	// Step 1: Explode demand
	step1 := s.GetStepDefinition(1)
	assert.Equal(t, "production-planning", step1.ServiceName)
	assert.Equal(t, "ExplodeDemandIntoRequirements", step1.HandlerMethod)

	// Step 2: Calculate net requirements
	step2 := s.GetStepDefinition(2)
	assert.Equal(t, "production-planning", step2.ServiceName)
	assert.Equal(t, "CalculateMRPNetRequirements", step2.HandlerMethod)

	// Step 3: Run lot sizing
	step3 := s.GetStepDefinition(3)
	assert.Equal(t, "cost-center", step3.ServiceName)
	assert.Equal(t, "RunLotSizingAlgorithm", step3.HandlerMethod)
}

func TestMRPLotSizingOptimizationSaga_PlannedOrderCreation(t *testing.T) {
	s := NewMRPLotSizingOptimizationSaga()

	// Step 6: Create planned purchase orders
	step6 := s.GetStepDefinition(6)
	assert.Equal(t, int32(6), step6.StepNumber)
	assert.Equal(t, "procurement", step6.ServiceName)
	assert.Equal(t, "CreatePlannedPurchaseOrders", step6.HandlerMethod)
	assert.True(t, step6.IsCritical)

	// Step 7: Create planned production orders
	step7 := s.GetStepDefinition(7)
	assert.Equal(t, int32(7), step7.StepNumber)
	assert.Equal(t, "production-planning", step7.ServiceName)
	assert.Equal(t, "CreatePlannedProductionOrders", step7.HandlerMethod)
	assert.True(t, step7.IsCritical)
}

func TestMRPLotSizingOptimizationSaga_CapacityValidation(t *testing.T) {
	s := NewMRPLotSizingOptimizationSaga()

	// Step 8: Validate against capacity
	step8 := s.GetStepDefinition(8)
	assert.Equal(t, int32(8), step8.StepNumber)
	assert.Equal(t, "work-center", step8.ServiceName)
	assert.Equal(t, "ValidatePlanAgainstCapacity", step8.HandlerMethod)
	assert.True(t, step8.IsCritical)
}

func TestMRPLotSizingOptimizationSaga_ForecastUpdate(t *testing.T) {
	s := NewMRPLotSizingOptimizationSaga()

	// Step 9: Update forecasts
	step9 := s.GetStepDefinition(9)
	assert.Equal(t, int32(9), step9.StepNumber)
	assert.Equal(t, "inventory-core", step9.ServiceName)
	assert.Equal(t, "UpdateInventoryForecasts", step9.HandlerMethod)
	assert.True(t, step9.IsCritical)
}

// ===== Integration Tests =====

func TestAllExtendedSagasRegistrable(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewJobCostingSaga(),
		NewCostVarianceAnalysisSaga(),
		NewScrapReworkManagementSaga(),
		NewSubcontractingCostTrackingSaga(),
		NewBatchCostingTraceabilitySaga(),
		NewMRPLotSizingOptimizationSaga(),
	}

	expectedTypes := []string{
		"SAGA-M07",
		"SAGA-M08",
		"SAGA-M09",
		"SAGA-M10",
		"SAGA-M11",
		"SAGA-M12",
	}

	for i, sagaHandler := range sagas {
		assert.Equal(t, expectedTypes[i], sagaHandler.SagaType())
		assert.NotNil(t, sagaHandler.GetStepDefinitions())
		assert.Greater(t, len(sagaHandler.GetStepDefinitions()), 0)
	}
}

func TestJobCostingSaga_ServiceRegistry(t *testing.T) {
	s := NewJobCostingSaga()
	steps := s.GetStepDefinitions()

	services := make(map[string]bool)
	for _, step := range steps {
		if step.StepNumber >= 1 && step.StepNumber <= 12 {
			services[step.ServiceName] = true
		}
	}

	// Verify all required services are present
	requiredServices := []string{
		"production-order",
		"job-card",
		"cost-center",
		"general-ledger",
		"work-center",
	}

	for _, svc := range requiredServices {
		assert.True(t, services[svc], "Service %s should be used", svc)
	}
}

func TestCostVarianceAnalysisSaga_ServiceRegistry(t *testing.T) {
	s := NewCostVarianceAnalysisSaga()
	steps := s.GetStepDefinitions()

	services := make(map[string]bool)
	for _, step := range steps {
		if step.StepNumber >= 1 && step.StepNumber <= 10 {
			services[step.ServiceName] = true
		}
	}

	// Verify all required services are present
	requiredServices := []string{
		"production-order",
		"cost-center",
		"general-ledger",
	}

	for _, svc := range requiredServices {
		assert.True(t, services[svc], "Service %s should be used", svc)
	}
}

func TestScrapReworkManagementSaga_ServiceRegistry(t *testing.T) {
	s := NewScrapReworkManagementSaga()
	steps := s.GetStepDefinitions()

	services := make(map[string]bool)
	for _, step := range steps {
		if step.StepNumber >= 1 && step.StepNumber <= 9 {
			services[step.ServiceName] = true
		}
	}

	// Verify all required services are present
	requiredServices := []string{
		"quality-production",
		"production-order",
		"inventory-core",
		"cost-center",
		"general-ledger",
	}

	for _, svc := range requiredServices {
		assert.True(t, services[svc], "Service %s should be used", svc)
	}
}

func TestSubcontractingCostTrackingSaga_ServiceRegistry(t *testing.T) {
	s := NewSubcontractingCostTrackingSaga()
	steps := s.GetStepDefinitions()

	services := make(map[string]bool)
	for _, step := range steps {
		if step.StepNumber >= 1 && step.StepNumber <= 10 {
			services[step.ServiceName] = true
		}
	}

	// Verify all required services are present
	requiredServices := []string{
		"subcontracting",
		"purchase-invoice",
		"cost-center",
		"production-order",
		"general-ledger",
		"vendor",
	}

	for _, svc := range requiredServices {
		assert.True(t, services[svc], "Service %s should be used", svc)
	}
}

func TestBatchCostingTraceabilitySaga_ServiceRegistry(t *testing.T) {
	s := NewBatchCostingTraceabilitySaga()
	steps := s.GetStepDefinitions()

	services := make(map[string]bool)
	for _, step := range steps {
		if step.StepNumber >= 1 && step.StepNumber <= 9 {
			services[step.ServiceName] = true
		}
	}

	// Verify all required services are present
	requiredServices := []string{
		"inventory-core",
		"production-order",
		"job-card",
		"cost-center",
		"general-ledger",
	}

	for _, svc := range requiredServices {
		assert.True(t, services[svc], "Service %s should be used", svc)
	}
}

func TestMRPLotSizingOptimizationSaga_ServiceRegistry(t *testing.T) {
	s := NewMRPLotSizingOptimizationSaga()
	steps := s.GetStepDefinitions()

	services := make(map[string]bool)
	for _, step := range steps {
		if step.StepNumber >= 1 && step.StepNumber <= 10 {
			services[step.ServiceName] = true
		}
	}

	// Verify all required services are present
	requiredServices := []string{
		"production-planning",
		"cost-center",
		"procurement",
		"work-center",
		"inventory-core",
	}

	for _, svc := range requiredServices {
		assert.True(t, services[svc], "Service %s should be used", svc)
	}
}

func TestAllSagasHaveTimeouts(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewJobCostingSaga(),
		NewCostVarianceAnalysisSaga(),
		NewScrapReworkManagementSaga(),
		NewSubcontractingCostTrackingSaga(),
		NewBatchCostingTraceabilitySaga(),
		NewMRPLotSizingOptimizationSaga(),
	}

	for _, sagaHandler := range sagas {
		steps := sagaHandler.GetStepDefinitions()
		for _, step := range steps {
			assert.Greater(t, step.TimeoutSeconds, int32(0), "Step %d should have timeout", step.StepNumber)
		}
	}
}

func TestAllCriticalStepsHaveRetryConfig(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewJobCostingSaga(),
		NewCostVarianceAnalysisSaga(),
		NewScrapReworkManagementSaga(),
		NewSubcontractingCostTrackingSaga(),
		NewBatchCostingTraceabilitySaga(),
		NewMRPLotSizingOptimizationSaga(),
	}

	for _, sagaHandler := range sagas {
		steps := sagaHandler.GetStepDefinitions()
		for _, step := range steps {
			if step.IsCritical {
				assert.NotNil(t, step.RetryConfig, "Critical step %d should have retry config", step.StepNumber)
				assert.Greater(t, step.RetryConfig.MaxRetries, int32(0))
				assert.Greater(t, step.RetryConfig.InitialBackoffMs, int32(0))
				assert.Greater(t, step.RetryConfig.MaxBackoffMs, step.RetryConfig.InitialBackoffMs)
			}
		}
	}
}

func TestCompensationStepsPresent(t *testing.T) {
	sagas := []saga.SagaHandler{
		NewJobCostingSaga(),
		NewCostVarianceAnalysisSaga(),
		NewScrapReworkManagementSaga(),
		NewSubcontractingCostTrackingSaga(),
		NewBatchCostingTraceabilitySaga(),
		NewMRPLotSizingOptimizationSaga(),
	}

	for _, sagaHandler := range sagas {
		steps := sagaHandler.GetStepDefinitions()

		compensationSteps := 0
		for _, step := range steps {
			if step.StepNumber >= 100 {
				compensationSteps++
			}
		}
		assert.Greater(t, compensationSteps, 0, "Saga %s should have compensation steps", sagaHandler.SagaType())
	}
}
