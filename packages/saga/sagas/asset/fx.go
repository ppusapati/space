// Package asset provides saga handlers for asset management workflows
package asset

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// AssetSagasModule provides all asset saga handlers
var AssetSagasModule = fx.Module("asset-sagas",
	fx.Provide(
		// SAGA-A01: Asset Acquisition (IAS 16 Capitalization)
		fx.Annotate(
			NewAssetAcquisitionSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-A02: Asset Depreciation (Monthly Accrual)
		fx.Annotate(
			NewAssetDepreciationSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-A03: Asset Disposal (Gain/Loss Calculation)
		fx.Annotate(
			NewAssetDisposalSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-A04: Asset Revaluation (IAS 16 Fair Value)
		fx.Annotate(
			NewAssetRevaluationSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
	),
)

// AssetSagasRegistrationModule registers all asset saga handlers with the global registry
var AssetSagasRegistrationModule = fx.Module("asset-sagas-registration",
	fx.Invoke(RegisterAssetSagaHandlers),
)

// RegisterAssetSagaHandlers registers all asset saga handlers with the global saga registry
func RegisterAssetSagaHandlers(handlers []saga.SagaHandler) {
	for _, handler := range handlers {
		saga.GlobalSagaRegistry.Register(handler.SagaType(), handler)
	}
}

// ProvideAssetSagaHandlers provides all asset saga handlers as a slice
func ProvideAssetSagaHandlers() []saga.SagaHandler {
	return []saga.SagaHandler{
		NewAssetAcquisitionSaga(),
		NewAssetDepreciationSaga(),
		NewAssetDisposalSaga(),
		NewAssetRevaluationSaga(),
	}
}
