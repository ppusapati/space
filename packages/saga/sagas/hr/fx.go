// Package hr provides saga handlers for human resources module workflows
package hr

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// HRSagasModule provides all HR saga handlers with dependency injection
var HRSagasModule = fx.Module("hr-sagas",
	fx.Provide(
		// SAGA-H01: Payroll Processing Saga (Phase 4C - Critical)
		fx.Annotate(
			NewPayrollProcessingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-H02: Employee Onboarding Saga (Phase 4B)
		fx.Annotate(
			NewEmployeeOnboardingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-H03: Employee Exit & Full and Final Settlement Saga (Phase 4C - Critical)
		fx.Annotate(
			NewEmployeeExitSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-H04: Expense Reimbursement Saga (Phase 4B)
		fx.Annotate(
			NewExpenseReimbursementSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-H05: Leave Application Saga (Phase 4A)
		fx.Annotate(
			NewLeaveApplicationSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-H06: Appraisal & Salary Revision Saga (Phase 4B)
		fx.Annotate(
			NewAppraisalSalaryRevisionSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-SR01: Form 16 Generation (Annual Tax Certificate) - Priority 1 Statutory
		fx.Annotate(
			NewForm16GenerationSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-SR02: PF/ESI Monthly Remittance - Priority 1 Statutory
		fx.Annotate(
			NewPFESIRemittanceSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-SR03: TDS Payment & Return Filing - Priority 1 Statutory
		fx.Annotate(
			NewTDSPaymentReturnFilingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-SR04: Leave Encashment & Settlement (Full & Final) - Priority 1 Statutory
		fx.Annotate(
			NewLeaveEncashmentSettlementSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
	),
)

// HRSagasRegistrationModule registers all HR saga handlers with the global registry
var HRSagasRegistrationModule = fx.Module("hr-sagas-registration",
	fx.Invoke(RegisterHRSagaHandlers),
)

// RegisterHRSagaHandlers registers all HR saga handlers with the global saga registry
func RegisterHRSagaHandlers(handlers []saga.SagaHandler) {
	for _, handler := range handlers {
		saga.GlobalSagaRegistry.Register(handler.SagaType(), handler)
	}
}

// ProvideHRSagaHandlers provides all HR saga handlers as a slice
// This is a convenience function for cases where manual aggregation is needed
func ProvideHRSagaHandlers() []saga.SagaHandler {
	return []saga.SagaHandler{
		NewPayrollProcessingSaga(),
		NewEmployeeOnboardingSaga(),
		NewEmployeeExitSaga(),
		NewExpenseReimbursementSaga(),
		NewLeaveApplicationSaga(),
		NewAppraisalSalaryRevisionSaga(),
		NewForm16GenerationSaga(),
		NewPFESIRemittanceSaga(),
		NewTDSPaymentReturnFilingSaga(),
		NewLeaveEncashmentSettlementSaga(),
	}
}
