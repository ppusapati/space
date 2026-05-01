// Package connector provides FX dependency injection for RPC connectivity
package connector

import (
	"go.uber.org/fx"
	"p9e.in/samavaya/packages/saga"
	"p9e.in/samavaya/packages/saga/sagas"
)

// ConnectorParams defines dependencies for RPC connector
type ConnectorParams struct {
	fx.In
}

// ConnectorResult provides RPC connector and service registry
type ConnectorResult struct {
	fx.Out

	RpcConnector saga.RpcConnector
}

// ConnectorModule provides the RPC connector with service registry
var ConnectorModule = fx.Module(
	"saga-connector",

	// Provide service registry with default endpoints
	fx.Provide(
		func() ServiceRegistry {
			return NewDefaultServiceRegistry(sagas.ServiceRegistry)
		},
	),

	// Provide RPC connector
	fx.Provide(
		func(params ConnectorParams, registry ServiceRegistry) ConnectorResult {
			connector := NewConnectRPCConnector(registry)
			return ConnectorResult{
				RpcConnector: connector,
			}
		},
	),
)
