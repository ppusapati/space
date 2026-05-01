// Package construction provides saga handlers for construction project module workflows
package construction

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// ConstructionSagasModule provides all construction saga handlers with dependency injection
var ConstructionSagasModule = fx.Module("construction-sagas",
	fx.Provide(
		// SAGA-C01: Construction Project Initiation (Phase 5B)
		fx.Annotate(
			NewConstructionProjectInitiationSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-C02: Progress Billing (Construction) (Phase 5B)
		fx.Annotate(
			NewProgressBillingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-C03: Material Procurement & Site Delivery (Phase 5B)
		fx.Annotate(
			NewMaterialProcurementSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-C04: Subcontractor Management & Payment (Phase 5B)
		fx.Annotate(
			NewSubcontractorManagementSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-C05: Quality Assurance & Inspection Management (Phase 5B)
		fx.Annotate(
			NewQualityAssuranceSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-C06: Construction Site Expenses & Cost Control (Phase 5B)
		fx.Annotate(
			NewSiteExpensesSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-C07: Project Closure & Final Settlement (Phase 5B)
		fx.Annotate(
			NewProjectClosureSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
	),
)

// ConstructionSagasRegistrationModule registers all construction saga handlers with the global registry
var ConstructionSagasRegistrationModule = fx.Module("construction-sagas-registration",
	fx.Invoke(RegisterConstructionSagaHandlers),
)

// RegisterConstructionSagaHandlers registers all construction saga handlers with the global saga registry
func RegisterConstructionSagaHandlers(handlers []saga.SagaHandler) {
	for _, handler := range handlers {
		saga.GlobalSagaRegistry.Register(handler.SagaType(), handler)
	}
}

// ProvideConstructionSagaHandlers provides all construction saga handlers as a slice
// This is a convenience function for cases where manual aggregation is needed
func ProvideConstructionSagaHandlers() []saga.SagaHandler {
	return []saga.SagaHandler{
		NewConstructionProjectInitiationSaga(),
		NewProgressBillingSaga(),
		NewMaterialProcurementSaga(),
		NewSubcontractorManagementSaga(),
		NewQualityAssuranceSaga(),
		NewSiteExpensesSaga(),
		NewProjectClosureSaga(),
	}
}
