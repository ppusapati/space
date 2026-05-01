// Package finance provides saga handlers for finance module workflows
package finance

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// FinanceSagasModule provides all finance saga handlers
var FinanceSagasModule = fx.Module("finance-sagas",
	fx.Provide(
		// SAGA-F01: Month-End Financial Close (Phase 4C - Critical, No Compensation)
		fx.Annotate(
			NewMonthEndCloseSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-F02: Bank Reconciliation (Phase 4C - Critical)
		fx.Annotate(
			NewBankReconciliationSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-F03: Multi-Currency Revaluation (Phase 4C - Critical)
		fx.Annotate(
			NewMultiCurrencyRevaluationSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-F04: Intercompany Transaction (Phase 4C - Critical)
		fx.Annotate(
			NewIntercompanyTransactionSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-F05: Revenue Recognition (IndAS 115) (Phase 4A)
		fx.Annotate(
			NewRevenueRecognitionSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-F06: Asset Capitalization (Phase 4A)
		fx.Annotate(
			NewAssetCapitalizationSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-F07: GST Input Credit Reversal (Phase 4A)
		fx.Annotate(
			NewGSTCreditReversalSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-F08: Cost Center Allocation (Phase 4A)
		fx.Annotate(
			NewCostCenterAllocationSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
	),
)

// FinanceSagasRegistrationModule registers all finance saga handlers with the global registry
var FinanceSagasRegistrationModule = fx.Module("finance-sagas-registration",
	fx.Invoke(RegisterFinanceSagaHandlers),
)

// RegisterFinanceSagaHandlers registers all finance saga handlers with the global saga registry
func RegisterFinanceSagaHandlers(handlers []saga.SagaHandler) {
	for _, handler := range handlers {
		saga.GlobalSagaRegistry.Register(handler.SagaType(), handler)
	}
}

// ProvideFinanceSagaHandlers provides all finance saga handlers as a slice
func ProvideFinanceSagaHandlers() []saga.SagaHandler {
	return []saga.SagaHandler{
		NewMonthEndCloseSaga(),
		NewBankReconciliationSaga(),
		NewMultiCurrencyRevaluationSaga(),
		NewIntercompanyTransactionSaga(),
		NewRevenueRecognitionSaga(),
		NewAssetCapitalizationSaga(),
		NewGSTCreditReversalSaga(),
		NewCostCenterAllocationSaga(),
	}
}
