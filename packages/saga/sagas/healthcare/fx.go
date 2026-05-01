// Package healthcare provides saga handlers for healthcare module workflows
package healthcare

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// HealthcareSagasModule provides all healthcare saga handlers with dependency injection
var HealthcareSagasModule = fx.Module("healthcare-sagas",
	fx.Provide(
		// SAGA-HC01: Patient Registration & Onboarding (Phase 6C)
		fx.Annotate(
			NewPatientRegistrationSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-HC02: Service Delivery & Treatment Workflows (Phase 6C)
		fx.Annotate(
			NewServiceDeliverySaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-HC03: Claims Processing & Insurance Billing (Phase 6C)
		fx.Annotate(
			NewClaimsProcessingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-HC04: Medical Supply & Inventory Management (Phase 6C)
		fx.Annotate(
			NewMedicalSupplySaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-HC05: Regulatory Compliance & Auditing (Phase 6C)
		fx.Annotate(
			NewComplianceSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-HC06: Provider Network & Referral Management (Phase 6C)
		fx.Annotate(
			NewProviderNetworkSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-HC07: Quality Assurance & Patient Safety (Phase 6C)
		fx.Annotate(
			NewQualityAssuranceSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
	),
)

// HealthcareSagasRegistrationModule registers all healthcare saga handlers with the global registry
var HealthcareSagasRegistrationModule = fx.Module("healthcare-sagas-registration",
	fx.Invoke(RegisterHealthcareSagaHandlers),
)

// RegisterHealthcareSagaHandlers registers all healthcare saga handlers with the global saga registry
func RegisterHealthcareSagaHandlers(handlers []saga.SagaHandler) {
	for _, handler := range handlers {
		saga.GlobalSagaRegistry.Register(handler.SagaType(), handler)
	}
}

// ProvideHealthcareSagaHandlers provides all healthcare saga handlers as a slice
// This is a convenience function for cases where manual aggregation is needed
func ProvideHealthcareSagaHandlers() []saga.SagaHandler {
	return []saga.SagaHandler{
		NewPatientRegistrationSaga(),
		NewServiceDeliverySaga(),
		NewClaimsProcessingSaga(),
		NewMedicalSupplySaga(),
		NewComplianceSaga(),
		NewProviderNetworkSaga(),
		NewQualityAssuranceSaga(),
	}
}
