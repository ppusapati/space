// Package manufacturing provides saga handlers for manufacturing module workflows
package manufacturing

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// ManufacturingSagasModule provides all manufacturing saga handlers with dependency injection
var ManufacturingSagasModule = fx.Module("manufacturing-sagas",
	fx.Provide(
		// SAGA-M01: Production Order Execution Saga (Phase 4C - Critical)
		fx.Annotate(
			NewProductionOrderSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-M02: Subcontracting Saga (Phase 4C - Critical)
		fx.Annotate(
			NewSubcontractingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-M03: BOM Explosion & MRP Saga (Phase 4B)
		fx.Annotate(
			NewBOMExplosionMRPSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-M04: Job Card Consumption Saga (Phase 4B)
		fx.Annotate(
			NewJobCardConsumptionSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-M05: Routing & Operation Sequencing Saga (Phase 4A)
		fx.Annotate(
			NewRoutingSequencingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-M06: Quality & Rework Saga (Phase 4B)
		fx.Annotate(
			NewQualityReworkSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-M07: Job Costing & Overhead Allocation Saga (Phase 7)
		fx.Annotate(
			NewJobCostingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-M08: Cost Variance Analysis Saga (Phase 7)
		fx.Annotate(
			NewCostVarianceAnalysisSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-M09: Scrap & Rework Management Saga (Phase 7)
		fx.Annotate(
			NewScrapReworkManagementSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-M10: Subcontracting Cost Tracking Saga (Phase 7)
		fx.Annotate(
			NewSubcontractingCostTrackingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-M11: Batch/Lot Costing & Traceability Saga (Phase 7)
		fx.Annotate(
			NewBatchCostingTraceabilitySaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-M12: MRP & Lot Sizing Optimization Saga (Phase 7)
		fx.Annotate(
			NewMRPLotSizingOptimizationSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
	),
)

// ManufacturingSagasRegistrationModule registers all manufacturing saga handlers with the global registry
var ManufacturingSagasRegistrationModule = fx.Module("manufacturing-sagas-registration",
	fx.Invoke(RegisterManufacturingSagaHandlers),
)

// RegisterManufacturingSagaHandlers registers all manufacturing saga handlers with the global saga registry
func RegisterManufacturingSagaHandlers(handlers []saga.SagaHandler) {
	for _, handler := range handlers {
		saga.GlobalSagaRegistry.Register(handler.SagaType(), handler)
	}
}

// ProvideManufacturingSagaHandlers provides all manufacturing saga handlers as a slice
// This is a convenience function for cases where manual aggregation is needed
func ProvideManufacturingSagaHandlers() []saga.SagaHandler {
	return []saga.SagaHandler{
		NewProductionOrderSaga(),
		NewSubcontractingSaga(),
		NewBOMExplosionMRPSaga(),
		NewJobCardConsumptionSaga(),
		NewRoutingSequencingSaga(),
		NewQualityReworkSaga(),
		NewJobCostingSaga(),
		NewCostVarianceAnalysisSaga(),
		NewScrapReworkManagementSaga(),
		NewSubcontractingCostTrackingSaga(),
		NewBatchCostingTraceabilitySaga(),
		NewMRPLotSizingOptimizationSaga(),
	}
}
