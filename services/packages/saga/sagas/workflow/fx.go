// Package workflow provides saga handlers for workflow management
package workflow

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// WorkflowSagasModule provides all workflow saga handlers with dependency injection
var WorkflowSagasModule = fx.Module("workflow-sagas",
	fx.Provide(
		// SAGA-WF01: Multi-Level Approval Routing (Phase 7)
		fx.Annotate(
			NewMultiLevelApprovalRoutingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-WF02: Conditional Workflow Routing (Phase 7)
		fx.Annotate(
			NewConditionalWorkflowRoutingSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
		// SAGA-WF03: Parallel Consolidation & Multi-Branch Processing (Phase 7)
		fx.Annotate(
			NewParallelConsolidationSaga,
			fx.ResultTags(`group:"saga_handlers"`),
		),
	),
)

// WorkflowSagasRegistrationModule registers all workflow saga handlers with the global registry
var WorkflowSagasRegistrationModule = fx.Module("workflow-sagas-registration",
	fx.Invoke(RegisterWorkflowSagaHandlers),
)

// RegisterWorkflowSagaHandlers registers all workflow saga handlers with the global saga registry
func RegisterWorkflowSagaHandlers(handlers []saga.SagaHandler) {
	for _, handler := range handlers {
		saga.GlobalSagaRegistry.Register(handler.SagaType(), handler)
	}
}

// ProvideWorkflowSagaHandlers provides all workflow saga handlers as a slice
// This is a convenience function for cases where manual aggregation is needed
func ProvideWorkflowSagaHandlers() []saga.SagaHandler {
	return []saga.SagaHandler{
		NewMultiLevelApprovalRoutingSaga(),
		NewConditionalWorkflowRoutingSaga(),
		NewParallelConsolidationSaga(),
	}
}
