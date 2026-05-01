// Package gst provides saga handlers for GST compliance module workflows
package gst

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// GSTSagasModule provides all GST saga handlers with dependency injection
var GSTSagasModule = fx.Module("gst-sagas",
	fx.Provide(
		// SAGA-G01: GST Return Filing (Phase 5A)
		fx.Annotate(
			NewGSTReturnFilingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-G02: ITC Reconciliation (Phase 5A)
		fx.Annotate(
			NewITCReconciliationSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-G03: E-way Bill Generation (Phase 5A)
		fx.Annotate(
			NewEwayBillSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-G04: GST Audit & Compliance (Phase 5A)
		fx.Annotate(
			NewGSTAuditSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-G05: GST Amendment & Correction (Phase 5A)
		fx.Annotate(
			NewGSTAmendmentSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-G06: GST Payment & Settlement (Phase 5A)
		fx.Annotate(
			NewGSTPaymentSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-G07: Reverse Charge Mechanism (Priority 1)
		fx.Annotate(
			NewReverseChargeMechanismSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-G08: Credit Note & Debit Note Processing (Priority 1)
		fx.Annotate(
			NewCreditDebitNoteSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
	),
)

// GSTSagasRegistrationModule registers all GST saga handlers with the global registry
var GSTSagasRegistrationModule = fx.Module("gst-sagas-registration",
	fx.Invoke(RegisterGSTSagaHandlers),
)

// RegisterGSTSagaHandlers registers all GST saga handlers with the global saga registry
func RegisterGSTSagaHandlers(handlers []saga.SagaHandler) {
	for _, handler := range handlers {
		saga.GlobalSagaRegistry.Register(handler.SagaType(), handler)
	}
}

// ProvideGSTSagaHandlers provides all GST saga handlers as a slice
// This is a convenience function for cases where manual aggregation is needed
func ProvideGSTSagaHandlers() []saga.SagaHandler {
	return []saga.SagaHandler{
		NewGSTReturnFilingSaga(),
		NewITCReconciliationSaga(),
		NewEwayBillSaga(),
		NewGSTAuditSaga(),
		NewGSTAmendmentSaga(),
		NewGSTPaymentSaga(),
		NewReverseChargeMechanismSaga(),
		NewCreditDebitNoteSaga(),
	}
}
