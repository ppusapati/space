// Package agriculture provides saga handlers for agriculture module workflows
package agriculture

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// AgricultureSagasModule provides all agriculture saga handlers with dependency injection
var AgricultureSagasModule = fx.Module("agriculture-sagas",
	fx.Provide(
		// SAGA-A01: Crop Planning & Resource Allocation (Phase 5B)
		fx.Annotate(
			NewCropPlanningSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-A02: Farm Operations & Activity Tracking (Phase 5B)
		fx.Annotate(
			NewFarmOperationsSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-A03: Harvest & Post-Harvest Management (Phase 5B)
		fx.Annotate(
			NewHarvestManagementSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-A04: Agricultural Procurement & Supply Chain (Phase 5B)
		fx.Annotate(
			NewProcurementSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-A05: Farmer Payment & Advance Management (Phase 5B)
		fx.Annotate(
			NewFarmerPaymentSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-A06: Agricultural Produce Sales & Billing (Phase 5B)
		fx.Annotate(
			NewProduceSalesSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-A07: Agricultural Compliance & Certification (Phase 5B)
		fx.Annotate(
			NewComplianceCertificationSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
	),
)

// AgricultureSagasRegistrationModule registers all agriculture saga handlers with the global registry
var AgricultureSagasRegistrationModule = fx.Module("agriculture-sagas-registration",
	fx.Invoke(RegisterAgricultureSagaHandlers),
)

// RegisterAgricultureSagaHandlers registers all agriculture saga handlers with the global saga registry
func RegisterAgricultureSagaHandlers(handlers []saga.SagaHandler) {
	for _, handler := range handlers {
		saga.GlobalSagaRegistry.Register(handler.SagaType(), handler)
	}
}

// ProvideAgricultureSagaHandlers provides all agriculture saga handlers as a slice
// This is a convenience function for cases where manual aggregation is needed
func ProvideAgricultureSagaHandlers() []saga.SagaHandler {
	return []saga.SagaHandler{
		NewCropPlanningSaga(),
		NewFarmOperationsSaga(),
		NewHarvestManagementSaga(),
		NewProcurementSaga(),
		NewFarmerPaymentSaga(),
		NewProduceSalesSaga(),
		NewComplianceCertificationSaga(),
	}
}
