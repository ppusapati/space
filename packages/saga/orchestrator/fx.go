// Package orchestrator provides FX dependency injection module
package orchestrator

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// OrchestratorParams defines orchestrator component dependencies
type OrchestratorParams struct {
	fx.In

	StepExecutor          saga.SagaStepExecutor
	TimeoutHandler        saga.SagaTimeoutHandler
	EventPublisher        saga.SagaEventPublisher
	Repository            saga.SagaRepository
	ExecutionLogRepository saga.SagaExecutionLogRepository
	Config                *saga.DefaultConfig
}

// RegistryParams defines registry component dependencies
type RegistryParams struct {
	fx.In
}

// SagaOrchestratorModule provides saga orchestrator and registry
var SagaOrchestratorModule = fx.Module(
	"saga_orchestrator",
	fx.Provide(NewSagaRegistry),
	fx.Provide(
		func(params OrchestratorParams, registry *SagaRegistry) saga.SagaOrchestrator {
			return NewSagaOrchestratorImpl(
				registry,
				params.StepExecutor,
				params.TimeoutHandler,
				params.EventPublisher,
				params.Repository,
				params.ExecutionLogRepository,
				params.Config,
			)
		},
	),
)
