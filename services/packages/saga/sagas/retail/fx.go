// Package retail provides saga handlers for retail module workflows
package retail

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// RetailSagasModule provides all retail saga handlers with dependency injection
var RetailSagasModule = fx.Module("retail-sagas",
	fx.Provide(
		// SAGA-R01: POS Transaction Processing & Settlement (Phase 6A)
		fx.Annotate(
			NewPOSTransactionSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-R02: Inventory Synchronization & Stock Management (Phase 6A)
		fx.Annotate(
			NewInventorySyncSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-R03: Return & Refund Processing (Phase 6A)
		fx.Annotate(
			NewReturnRefundSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-R04: Loyalty Program & Customer Rewards (Phase 6A)
		fx.Annotate(
			NewLoyaltyProgramSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-R05: Promotion & Discount Management (Phase 6A)
		fx.Annotate(
			NewPromotionSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-R06: Stock Transfer Between Stores (Phase 6A)
		fx.Annotate(
			NewStockTransferSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-R07: Customer Account & Profile Management (Phase 6A)
		fx.Annotate(
			NewCustomerAccountSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-R08: Analytics & Reporting (Phase 6A)
		fx.Annotate(
			NewAnalyticsReportingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-R09: Merchandise Planning & Assortment (Phase 6A)
		fx.Annotate(
			NewMerchandisePlanningallocation,
			fx.ResultTags(`group:"saga_handlers"`),
		),
	),
)

// RetailSagasRegistrationModule registers all retail saga handlers with the global registry
var RetailSagasRegistrationModule = fx.Module("retail-sagas-registration",
	fx.Invoke(RegisterRetailSagaHandlers),
)

// RegisterRetailSagaHandlers registers all retail saga handlers with the global saga registry
func RegisterRetailSagaHandlers(handlers []saga.SagaHandler) {
	for _, handler := range handlers {
		saga.GlobalSagaRegistry.Register(handler.SagaType(), handler)
	}
}

// ProvideRetailSagaHandlers provides all retail saga handlers as a slice
// This is a convenience function for cases where manual aggregation is needed
func ProvideRetailSagaHandlers() []saga.SagaHandler {
	return []saga.SagaHandler{
		NewPOSTransactionSaga(),
		NewInventorySyncSaga(),
		NewReturnRefundSaga(),
		NewLoyaltyProgramSaga(),
		NewPromotionSaga(),
		NewStockTransferSaga(),
		NewCustomerAccountSaga(),
		NewAnalyticsReportingSaga(),
		NewMerchandisePlanningallocation(),
	}
}
