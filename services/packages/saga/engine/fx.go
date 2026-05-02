// Package engine is the FX composition for the saga orchestration engine.
//
// It was split out of the root `saga` package on 2026-04-19 (roadmap B.8)
// to break a root-↔-subpackage import cycle. Root `saga` now holds only
// contract types (interfaces, errors, config); this package imports those
// contracts and wires every subpackage (compensation / orchestrator /
// executor / events / connector / timeout / sagas/<domain>) into an fx
// Module the app composition root consumes.
//
// External callers switched from `saga.SagaEngineModule` to
// `engine.SagaEngineModule`; at the time of the split there are zero
// external importers.
package engine

import (
	"time"

	"go.uber.org/fx"

	"p9e.in/chetana/packages/saga"
	"p9e.in/chetana/packages/saga/compensation"
	"p9e.in/chetana/packages/saga/connector"
	"p9e.in/chetana/packages/saga/events"
	"p9e.in/chetana/packages/saga/executor"
	"p9e.in/chetana/packages/saga/models"
	"p9e.in/chetana/packages/saga/orchestrator"
	"p9e.in/chetana/packages/saga/sagas/agriculture"
	"p9e.in/chetana/packages/saga/sagas/banking"
	"p9e.in/chetana/packages/saga/sagas/construction"
	"p9e.in/chetana/packages/saga/sagas/finance"
	"p9e.in/chetana/packages/saga/sagas/gst"
	"p9e.in/chetana/packages/saga/sagas/healthcare"
	"p9e.in/chetana/packages/saga/sagas/hr"
	"p9e.in/chetana/packages/saga/sagas/inventory"
	"p9e.in/chetana/packages/saga/sagas/manufacturing"
	"p9e.in/chetana/packages/saga/sagas/projects"
	"p9e.in/chetana/packages/saga/sagas/purchase"
	"p9e.in/chetana/packages/saga/sagas/retail"
	"p9e.in/chetana/packages/saga/sagas/sales"
	supplychain "p9e.in/chetana/packages/saga/sagas/supply-chain"
	"p9e.in/chetana/packages/saga/sagas/warranty"
	"p9e.in/chetana/packages/saga/timeout"
)

// SagaEngineParams is the fx.In dependency bundle the engine Module needs.
// Callers (typically the app composition root) wire these types from their
// own modules so the engine stays decoupled from infrastructure choices —
// e.g. swap SagaRepository with a Postgres impl or an in-memory test fake
// without touching the engine wiring.
type SagaEngineParams struct {
	fx.In

	Config                 *saga.DefaultConfig
	StepExecutor           saga.SagaStepExecutor
	TimeoutHandler         saga.SagaTimeoutHandler
	EventPublisher         saga.SagaEventPublisher
	Repository             saga.SagaRepository
	ExecutionLogRepository saga.SagaExecutionLogRepository
}

// SagaEngineResult is the fx.Out set the engine Module provides back to
// the rest of the graph.
type SagaEngineResult struct {
	fx.Out

	Orchestrator   saga.SagaOrchestrator
	StepExecutor   saga.SagaStepExecutor
	TimeoutHandler saga.SagaTimeoutHandler
	EventPublisher saga.SagaEventPublisher
	Registry       *orchestrator.SagaRegistry
	CircuitBreaker saga.CircuitBreaker
}

// SagaEngineModule provides all saga engine components.
var SagaEngineModule = fx.Module(
	"saga_engine",

	// Step Executor Components
	fx.Provide(
		func() *executor.IdempotencyImpl {
			// TTL: 1 hour, Max cache size: 10,000 entries.
			return executor.NewIdempotencyImpl(1*time.Hour, 10000)
		},
	),

	fx.Provide(
		func() saga.RpcConnector {
			return executor.NewRpcConnectorImpl()
		},
	),

	fx.Provide(
		func(
			rpcConnector saga.RpcConnector,
			idempotency *executor.IdempotencyImpl,
		) saga.SagaStepExecutor {
			return executor.NewStepExecutorImpl(rpcConnector, idempotency)
		},
	),

	// Timeout Handler Components
	fx.Provide(
		func(config *saga.DefaultConfig) saga.SagaTimeoutHandler {
			defaultRetryConfig := &models.RetryConfiguration{
				MaxRetries:        config.DefaultMaxRetries,
				InitialBackoffMs:  int32(config.DefaultInitialBackoff.Milliseconds()),
				MaxBackoffMs:      int32(config.DefaultMaxBackoff.Milliseconds()),
				BackoffMultiplier: config.BackoffMultiplier,
				JitterFraction:    config.JitterFraction,
			}

			retryStrategies := make(map[string]*models.RetryConfiguration)

			return timeout.NewTimeoutHandlerImpl(defaultRetryConfig, retryStrategies)
		},
	),

	// Event Publisher Components
	fx.Provide(
		func() events.KafkaProducer {
			// Placeholder — the real Kafka producer binding is wired at the
			// app composition root. This mock keeps the module bootable in
			// isolation for tests.
			return &events.MockKafkaProducer{}
		},
	),

	fx.Provide(
		func(kafkaProducer events.KafkaProducer, config *saga.DefaultConfig) saga.SagaEventPublisher {
			return events.NewEventPublisherImpl(config.KafkaTopic, kafkaProducer)
		},
	),

	// Orchestrator Components
	orchestrator.SagaOrchestratorModule,

	// Compensation Components
	compensation.SagaCompensationEngineModule,

	// RPC Connector Components
	connector.ConnectorModule,

	// Domain saga handlers, registered in phase order.
	sales.SalesSagasModule,
	sales.SalesSagasRegistrationModule,

	purchase.PurchaseSagasModule,
	purchase.PurchaseSagasRegistrationModule,

	inventory.InventorySagasModule,
	inventory.InventorySagasRegistrationModule,

	manufacturing.ManufacturingSagasModule,
	manufacturing.ManufacturingSagasRegistrationModule,

	finance.FinanceSagasModule,
	finance.FinanceSagasRegistrationModule,

	hr.HRSagasModule,
	hr.HRSagasRegistrationModule,

	projects.ProjectsSagasModule,
	projects.ProjectsSagasRegistrationModule,

	gst.GSTSagasModule,
	gst.GSTSagasRegistrationModule,

	banking.BankingSagasModule,
	banking.BankingSagasRegistrationModule,

	construction.ConstructionSagasModule,
	construction.ConstructionSagasRegistrationModule,

	agriculture.AgricultureSagasModule,
	agriculture.AgricultureSagasRegistrationModule,

	retail.RetailSagasModule,
	retail.RetailSagasRegistrationModule,

	supplychain.SupplyChainSagasModule,
	supplychain.SupplyChainSagasRegistrationModule,

	healthcare.HealthcareSagasModule,
	healthcare.HealthcareSagasRegistrationModule,

	warranty.WarrantySagasModule,
	warranty.WarrantySagasRegistrationModule,

	// Orchestrator composer — builds the concrete SagaOrchestrator from
	// the registered subsystems.
	fx.Provide(
		func(
			registry *orchestrator.SagaRegistry,
			stepExecutor saga.SagaStepExecutor,
			timeoutHandler saga.SagaTimeoutHandler,
			eventPublisher saga.SagaEventPublisher,
			repository saga.SagaRepository,
			execLogRepository saga.SagaExecutionLogRepository,
			config *saga.DefaultConfig,
		) SagaOrchestratorResult {
			orch := orchestrator.NewSagaOrchestratorImpl(
				registry,
				stepExecutor,
				timeoutHandler,
				eventPublisher,
				repository,
				execLogRepository,
				config,
			)

			return SagaOrchestratorResult{
				Orchestrator: orch,
				Registry:     registry,
			}
		},
	),
)

// SagaOrchestratorResult provides orchestrator and registry.
type SagaOrchestratorResult struct {
	fx.Out

	Orchestrator saga.SagaOrchestrator
	Registry     *orchestrator.SagaRegistry
}

// MinimalSagaEngineModule is the engine Module plus a DefaultConfig provider.
// Useful for tests that want a bootable engine without a real composition
// root supplying every dep.
var MinimalSagaEngineModule = fx.Module(
	"saga_engine_minimal",

	fx.Provide(
		func() *saga.DefaultConfig {
			return &saga.DefaultConfig{
				DefaultTimeoutSeconds:   60,
				DefaultMaxRetries:       3,
				DefaultInitialBackoff:   time.Second,
				DefaultMaxBackoff:       30 * time.Second,
				BackoffMultiplier:       2.0,
				JitterFraction:          0.1,
				CircuitBreakerThreshold: 5,
				CircuitBreakerResetMs:   60000,
				KafkaTopic:              "saga-events",
				KafkaPartitions:         5,
			}
		},
	),

	SagaEngineModule,
)
