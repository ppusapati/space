// Package statutory provides saga handlers for statutory compliance workflows
package statutory

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// StatutorySagasModule provides all statutory compliance saga handlers
var StatutorySagasModule = fx.Module("statutory-sagas",
	fx.Provide(
		// SAGA-ST01: GSTR-1 Filing (Sales Tax Return) - Priority 1
		fx.Annotate(
			NewGSTR1FilingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-ST02: GSTR-2 ITC Claim (Input Tax Claim) - Priority 1
		fx.Annotate(
			NewGSTR2ITCSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-ST03: GSTR-9 Annual Return (Annual GST Reconciliation) - Priority 1
		fx.Annotate(
			NewGSTR9AnnualSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-ST04: TDS Return Filing (Tax Deduction Return) - Priority 1
		fx.Annotate(
			NewTDSReturnFilingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
	),
)

// StatutorySagasRegistrationModule registers all statutory saga handlers with the global registry
var StatutorySagasRegistrationModule = fx.Module("statutory-sagas-registration",
	fx.Invoke(RegisterStatutorySagaHandlers),
)

// RegisterStatutorySagaHandlers registers all statutory saga handlers with the global saga registry
func RegisterStatutorySagaHandlers(handlers []saga.SagaHandler) {
	for _, handler := range handlers {
		saga.GlobalSagaRegistry.Register(handler.SagaType(), handler)
	}
}

// ProvideStatutorySagaHandlers provides all statutory saga handlers as a slice
func ProvideStatutorySagaHandlers() []saga.SagaHandler {
	return []saga.SagaHandler{
		NewGSTR1FilingSaga(),
		NewGSTR2ITCSaga(),
		NewGSTR9AnnualSaga(),
		NewTDSReturnFilingSaga(),
	}
}
