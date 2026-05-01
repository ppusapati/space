// Package warranty provides saga handlers for warranty module workflows
package warranty

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// WarrantySagasModule provides all warranty saga handlers with dependency injection
var WarrantySagasModule = fx.Module("warranty-sagas",
	fx.Provide(
		// SAGA-W01: Warranty Claim Registration & Assessment (Phase 6D)
		fx.Annotate(
			NewWarrantyClaimSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-W02: Field Service & On-Site Repair (Phase 6D)
		fx.Annotate(
			NewFieldServiceSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-W03: Spare Parts Management & Fulfillment (Phase 6D)
		fx.Annotate(
			NewSparePartsSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-W04: SLA Management & Compliance Tracking (Phase 6D)
		fx.Annotate(
			NewSLAManagementSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-W05: Customer Satisfaction & Feedback (Phase 6D)
		fx.Annotate(
			NewCustomerSatisfactionSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-W06: Extended Warranty & Coverage Plans (Phase 6D)
		fx.Annotate(
			NewExtendedWarrantySaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
	),
)

// WarrantySagasRegistrationModule registers all warranty saga handlers with the global registry
var WarrantySagasRegistrationModule = fx.Module("warranty-sagas-registration",
	fx.Invoke(RegisterWarrantySagaHandlers),
)

// RegisterWarrantySagaHandlers registers all warranty saga handlers with the global saga registry
func RegisterWarrantySagaHandlers(handlers []saga.SagaHandler) {
	for _, handler := range handlers {
		saga.GlobalSagaRegistry.Register(handler.SagaType(), handler)
	}
}

// ProvideWarrantySagaHandlers provides all warranty saga handlers as a slice
// This is a convenience function for cases where manual aggregation is needed
func ProvideWarrantySagaHandlers() []saga.SagaHandler {
	return []saga.SagaHandler{
		NewWarrantyClaimSaga(),
		NewFieldServiceSaga(),
		NewSparePartsSaga(),
		NewSLAManagementSaga(),
		NewCustomerSatisfactionSaga(),
		NewExtendedWarrantySaga(),
	}
}
