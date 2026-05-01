// Package sales provides FX dependency injection for sales saga handlers
package sales

import (
	"go.uber.org/fx"
	"p9e.in/samavaya/packages/saga"
)

// SalesSagasModule provides all sales saga handler implementations
var SalesSagasModule = fx.Module(
	"saga-sales-sagas",

	// Provide all 7 saga handlers, tagged for group collection
	fx.Provide(
		fx.Annotate(
			NewOrderToCashSaga,
			fx.As(new(saga.SagaHandler)),
			fx.ResultTags(`group:"saga-handlers"`),
		),
		fx.Annotate(
			NewQuotationToOrderSaga,
			fx.As(new(saga.SagaHandler)),
			fx.ResultTags(`group:"saga-handlers"`),
		),
		fx.Annotate(
			NewOrderToFulfillmentSaga,
			fx.As(new(saga.SagaHandler)),
			fx.ResultTags(`group:"saga-handlers"`),
		),
		fx.Annotate(
			NewSalesReturnSaga,
			fx.As(new(saga.SagaHandler)),
			fx.ResultTags(`group:"saga-handlers"`),
		),
		fx.Annotate(
			NewCommissionCalculationSaga,
			fx.As(new(saga.SagaHandler)),
			fx.ResultTags(`group:"saga-handlers"`),
		),
		fx.Annotate(
			NewEInvoiceGenerationSaga,
			fx.As(new(saga.SagaHandler)),
			fx.ResultTags(`group:"saga-handlers"`),
		),
		fx.Annotate(
			NewDealerIncentiveSaga,
			fx.As(new(saga.SagaHandler)),
			fx.ResultTags(`group:"saga-handlers"`),
		),
	),
)

// SalesSagasRegistrationModule registers all sales saga handlers with the orchestrator
type RegistrationParams struct {
	fx.In
	Orchestrator saga.SagaOrchestrator
	Handlers     []saga.SagaHandler `group:"saga-handlers"`
}

func RegisterSalesSagaHandlers(params RegistrationParams) error {
	for _, handler := range params.Handlers {
		if err := params.Orchestrator.RegisterSagaHandler(handler.SagaType(), handler); err != nil {
			return err
		}
	}
	return nil
}

var SalesSagasRegistrationModule = fx.Module(
	"saga-sales-registration",
	fx.Invoke(RegisterSalesSagaHandlers),
)
