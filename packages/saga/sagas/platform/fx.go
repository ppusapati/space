// Package platform provides saga handlers for platform module workflows
package platform

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// PlatformSagasModule provides all platform saga handlers (Phase 7)
var PlatformSagasModule = fx.Module("platform-sagas",
	fx.Provide(
		// SAGA-PLAT01: Data Archive & Retention Management (9 forward + 5 compensation steps)
		fx.Annotate(
			NewDataArchiveRetentionSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-PLAT02: Cross-Module Reconciliation & Validation (10 forward + 9 compensation steps)
		fx.Annotate(
			NewCrossModuleReconciliationSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-PLAT03: Master Data Synchronization (8 forward + 6 compensation steps)
		fx.Annotate(
			NewMasterDataSynchronizationSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
	),
)

// PlatformSagasRegistrationModule registers all platform saga handlers with the global registry
var PlatformSagasRegistrationModule = fx.Module("platform-sagas-registration",
	fx.Invoke(RegisterPlatformSagaHandlers),
)

// RegisterPlatformSagaHandlers registers all platform saga handlers with the global saga registry
func RegisterPlatformSagaHandlers(handlers []saga.SagaHandler) {
	for _, handler := range handlers {
		saga.GlobalSagaRegistry.Register(handler.SagaType(), handler)
	}
}

// ProvidePlatformSagaHandlers provides all platform saga handlers as a slice
func ProvidePlatformSagaHandlers() []saga.SagaHandler {
	return []saga.SagaHandler{
		NewDataArchiveRetentionSaga(),
		NewCrossModuleReconciliationSaga(),
		NewMasterDataSynchronizationSaga(),
	}
}
