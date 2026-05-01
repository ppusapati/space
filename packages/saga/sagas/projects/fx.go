// Package projects provides saga handlers for projects module workflows
package projects

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// ProjectsSagasModule provides all projects saga handlers with dependency injection
var ProjectsSagasModule = fx.Module("projects-sagas",
	fx.Provide(
		// SAGA-PR01: Project Billing Saga (Phase 4B)
		fx.Annotate(
			NewProjectBillingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-PR02: Progress Billing Saga (Phase 4B)
		fx.Annotate(
			NewProgressBillingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-PR03: Subcontractor Payment Saga (Phase 4B)
		fx.Annotate(
			NewSubcontractorPaymentSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-PR04: Project Close Saga (Phase 4B)
		fx.Annotate(
			NewProjectCloseSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
	),
)

// ProjectsSagasRegistrationModule registers all projects saga handlers with the global registry
var ProjectsSagasRegistrationModule = fx.Module("projects-sagas-registration",
	fx.Invoke(RegisterProjectsSagaHandlers),
)

// RegisterProjectsSagaHandlers registers all projects saga handlers with the global saga registry
func RegisterProjectsSagaHandlers(handlers []saga.SagaHandler) {
	for _, handler := range handlers {
		saga.GlobalSagaRegistry.Register(handler.SagaType(), handler)
	}
}

// ProvideProjectsSagaHandlers provides all projects saga handlers as a slice
// This is a convenience function for cases where manual aggregation is needed
func ProvideProjectsSagaHandlers() []saga.SagaHandler {
	return []saga.SagaHandler{
		NewProjectBillingSaga(),
		NewProgressBillingSaga(),
		NewSubcontractorPaymentSaga(),
		NewProjectCloseSaga(),
	}
}
