// Package banking provides saga handlers for banking module workflows
package banking

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// BankingSagasModule provides all banking saga handlers with dependency injection
var BankingSagasModule = fx.Module("banking-sagas",
	fx.Provide(
		// SAGA-B01: Wire Transfer & Payment Authorization (Phase 5A)
		fx.Annotate(
			NewWireTransferSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-B02: Bank Reconciliation - Multi-Bank (Phase 5A)
		fx.Annotate(
			NewBankReconciliationMultiSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-B03: Cash Positioning & Forecasting (Phase 5A)
		fx.Annotate(
			NewCashPositioningSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-B04: Cheque Management & Processing (Phase 5A)
		fx.Annotate(
			NewChequeManagementSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-B05: Payment Gateway Integration & Settlement (Phase 5A)
		fx.Annotate(
			NewPaymentGatewaySaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-B06: Banking Compliance & Transaction Monitoring (Phase 5A)
		fx.Annotate(
			NewComplianceMonitoringSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
	),
)

// BankingSagasRegistrationModule registers all banking saga handlers with the global registry
var BankingSagasRegistrationModule = fx.Module("banking-sagas-registration",
	fx.Invoke(RegisterBankingSagaHandlers),
)

// RegisterBankingSagaHandlers registers all banking saga handlers with the global saga registry
func RegisterBankingSagaHandlers(handlers []saga.SagaHandler) {
	for _, handler := range handlers {
		saga.GlobalSagaRegistry.Register(handler.SagaType(), handler)
	}
}

// ProvideBankingSagaHandlers provides all banking saga handlers as a slice
// This is a convenience function for cases where manual aggregation is needed
func ProvideBankingSagaHandlers() []saga.SagaHandler {
	return []saga.SagaHandler{
		NewWireTransferSaga(),
		NewBankReconciliationMultiSaga(),
		NewCashPositioningSaga(),
		NewChequeManagementSaga(),
		NewPaymentGatewaySaga(),
		NewComplianceMonitoringSaga(),
	}
}
