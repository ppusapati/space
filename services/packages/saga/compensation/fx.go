// Package compensation provides FX dependency injection for compensation engine
package compensation

import (
	"go.uber.org/fx"

	"p9e.in/samavaya/packages/saga"
)

// CompensationEngineParams defines compensation engine dependencies
type CompensationEngineParams struct {
	fx.In

	StepExecutor          saga.SagaStepExecutor
	EventPublisher        saga.SagaEventPublisher
	Repository            saga.SagaRepository
	ExecutionLogRepository saga.SagaExecutionLogRepository
}

// CompensationEngineResult provides compensation engine and log repository
type CompensationEngineResult struct {
	fx.Out

	CompensationEngine saga.SagaCompensationEngine
	LogRepository      *CompensationLogRepositoryImpl
}

// SagaCompensationEngineModule provides compensation engine components
var SagaCompensationEngineModule = fx.Module(
	"saga_compensation",

	// Provide compensation log repository
	fx.Provide(
		func() *CompensationLogRepositoryImpl {
			return NewCompensationLogRepositoryImpl()
		},
	),

	// Provide compensation engine
	fx.Provide(
		func(params CompensationEngineParams) CompensationEngineResult {
			engine := NewCompensationEngineImpl(
				params.StepExecutor,
				params.EventPublisher,
				params.Repository,
				params.ExecutionLogRepository,
			)

			return CompensationEngineResult{
				CompensationEngine: engine,
				LogRepository:      NewCompensationLogRepositoryImpl(),
			}
		},
	),
)
