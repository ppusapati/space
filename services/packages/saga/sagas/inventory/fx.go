// Package inventory provides FX module for inventory saga handlers
package inventory

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// InventorySagasModule provides all inventory saga handlers
var InventorySagasModule = fx.Module(
	"saga-inventory-sagas",

	// Provide each saga handler with group tagging for collection
	fx.Provide(
		fx.Annotate(
			NewInterWarehouseTransferSaga,
			fx.As(new(saga.SagaHandler)),
			fx.ResultTags(`group:"saga-handlers"`),
		),
		fx.Annotate(
			NewCycleCountSaga,
			fx.As(new(saga.SagaHandler)),
			fx.ResultTags(`group:"saga-handlers"`),
		),
		fx.Annotate(
			NewQualityRejectionSaga,
			fx.As(new(saga.SagaHandler)),
			fx.ResultTags(`group:"saga-handlers"`),
		),
		fx.Annotate(
			NewLotSerialTrackingSaga,
			fx.As(new(saga.SagaHandler)),
			fx.ResultTags(`group:"saga-handlers"`),
		),
		fx.Annotate(
			NewDemandPlanningSaga,
			fx.As(new(saga.SagaHandler)),
			fx.ResultTags(`group:"saga-handlers"`),
		),
	),
)

// InventorySagasRegistrationModule registers all inventory saga handlers with orchestrator
var InventorySagasRegistrationModule = fx.Module(
	"saga-inventory-registration",
	fx.Invoke(RegisterInventorySagaHandlers),
)

// InventoryRegistrationParams contains dependencies for saga registration
type InventoryRegistrationParams struct {
	fx.In
	Orchestrator saga.SagaOrchestrator
	Handlers     []saga.SagaHandler `group:"saga-handlers"`
}

// RegisterInventorySagaHandlers registers all inventory saga handlers with the orchestrator
func RegisterInventorySagaHandlers(params InventoryRegistrationParams) error {
	for _, handler := range params.Handlers {
		sagaType := handler.SagaType()

		// Only register inventory sagas (I01-I05)
		if sagaType == "SAGA-I01" || sagaType == "SAGA-I02" ||
			sagaType == "SAGA-I03" || sagaType == "SAGA-I04" ||
			sagaType == "SAGA-I05" {
			if err := params.Orchestrator.RegisterSagaHandler(sagaType, handler); err != nil {
				return err
			}
		}
	}
	return nil
}
