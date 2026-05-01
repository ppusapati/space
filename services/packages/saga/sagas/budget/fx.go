// Package budget provides saga handlers for budget management workflows
package budget

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// BudgetSagasModule provides all budget saga handlers
var BudgetSagasModule = fx.Module("budget-sagas",
	fx.Provide(
		// SAGA-BU01: Budget Approval & Control (9 forward + 8 compensation steps)
		fx.Annotate(
			NewBudgetApprovalControlSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-BU02: Variance Analysis & Budget Review (8 forward + 7 compensation steps)
		fx.Annotate(
			NewVarianceAnalysisSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-BU03: CapEx Proposal & Investment Approval (11 forward + 10 compensation steps)
		fx.Annotate(
			NewCapExInvestmentSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
	),
)

// BudgetSagasRegistrationModule registers all budget saga handlers with the global registry
var BudgetSagasRegistrationModule = fx.Module("budget-sagas-registration",
	fx.Invoke(RegisterBudgetSagaHandlers),
)

// RegisterBudgetSagaHandlers registers all budget saga handlers with the global saga registry
func RegisterBudgetSagaHandlers(handlers []saga.SagaHandler) {
	for _, handler := range handlers {
		saga.GlobalSagaRegistry.Register(handler.SagaType(), handler)
	}
}

// ProvideBudgetSagaHandlers provides all budget saga handlers as a slice
func ProvideBudgetSagaHandlers() []saga.SagaHandler {
	return []saga.SagaHandler{
		NewBudgetApprovalControlSaga(),
		NewVarianceAnalysisSaga(),
		NewCapExInvestmentSaga(),
	}
}
