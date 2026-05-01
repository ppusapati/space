// Package purchase provides FX module for purchase saga handlers
package purchase

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// PurchaseSagasModule provides all purchase saga handlers
var PurchaseSagasModule = fx.Module(
	"saga-purchase-sagas",

	// Provide each saga handler with group tagging for collection
	fx.Provide(
		fx.Annotate(
			NewBudgetCheckSaga,
			fx.As(new(saga.SagaHandler)),
			fx.ResultTags(`group:"saga-handlers"`),
		),
		fx.Annotate(
			NewPurchaseReturnSaga,
			fx.As(new(saga.SagaHandler)),
			fx.ResultTags(`group:"saga-handlers"`),
		),
		fx.Annotate(
			NewProcureToPaySaga,
			fx.As(new(saga.SagaHandler)),
			fx.ResultTags(`group:"saga-handlers"`),
		),
		fx.Annotate(
			NewVendorPaymentTDSSaga,
			fx.As(new(saga.SagaHandler)),
			fx.ResultTags(`group:"saga-handlers"`),
		),
	),
)

// PurchaseSagasRegistrationModule registers all purchase saga handlers with orchestrator
var PurchaseSagasRegistrationModule = fx.Module(
	"saga-purchase-registration",
	fx.Invoke(RegisterPurchaseSagaHandlers),
)

// PurchaseRegistrationParams contains dependencies for saga registration
type PurchaseRegistrationParams struct {
	fx.In
	Orchestrator saga.SagaOrchestrator
	Handlers     []saga.SagaHandler `group:"saga-handlers"`
}

// RegisterPurchaseSagaHandlers registers all purchase saga handlers with the orchestrator
func RegisterPurchaseSagaHandlers(params PurchaseRegistrationParams) error {
	for _, handler := range params.Handlers {
		sagaType := handler.SagaType()

		// Only register purchase sagas (P01-P04)
		if sagaType == "SAGA-P01" || sagaType == "SAGA-P02" ||
			sagaType == "SAGA-P03" || sagaType == "SAGA-P04" {
			if err := params.Orchestrator.RegisterSagaHandler(sagaType, handler); err != nil {
				return err
			}
		}
	}
	return nil
}
