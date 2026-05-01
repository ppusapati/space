// Package executor provides RPC communication for saga steps
package executor

import (
	"context"
	"fmt"
	"sync"
)

// RpcConnectorImpl implements RPC connector for service communication
type RpcConnectorImpl struct {
	mu              sync.RWMutex
	serviceRegistry map[string]string    // serviceName -> endpoint
	clientCache     map[string]interface{} // endpoint -> cached client
}

// NewRpcConnectorImpl creates a new RPC connector instance
func NewRpcConnectorImpl() *RpcConnectorImpl {
	return &RpcConnectorImpl{
		serviceRegistry: make(map[string]string),
		clientCache:     make(map[string]interface{}),
	}
}

// InvokeHandler invokes a handler on a remote service via RPC
func (r *RpcConnectorImpl) InvokeHandler(
	ctx context.Context,
	endpoint string,
	handlerMethod string,
	request interface{},
) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. Validate inputs
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}
	if handlerMethod == "" {
		return nil, fmt.Errorf("handler method cannot be empty")
	}

	// 2. Get or create client for endpoint
	client, err := r.getOrCreateClient(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for endpoint %s: %w", endpoint, err)
	}

	// 3. Invoke RPC method via ConnectRPC
	// In real implementation, this would use the generated client
	response, err := invokeRPCMethod(ctx, client, handlerMethod, request)
	if err != nil {
		return nil, fmt.Errorf("RPC invocation failed: %w", err)
	}

	return response, nil
}

// GetServiceEndpoint resolves a service endpoint from registry
func (r *RpcConnectorImpl) GetServiceEndpoint(serviceName string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. Look up service in registry
	endpoint, exists := r.serviceRegistry[serviceName]
	if !exists {
		return "", fmt.Errorf("service %s not registered in endpoint registry", serviceName)
	}

	// 2. Return endpoint
	return endpoint, nil
}

// RegisterService registers a service endpoint
func (r *RpcConnectorImpl) RegisterService(serviceName string, endpoint string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 1. Validate inputs
	if serviceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	if endpoint == "" {
		return fmt.Errorf("endpoint cannot be empty")
	}

	// 2. Check if already registered
	if existing, exists := r.serviceRegistry[serviceName]; exists {
		if existing != endpoint {
			return fmt.Errorf("service %s already registered with different endpoint: %s", serviceName, existing)
		}
		return nil // Already registered with same endpoint
	}

	// 3. Register service
	r.serviceRegistry[serviceName] = endpoint

	return nil
}

// getOrCreateClient gets or creates a client for an endpoint
func (r *RpcConnectorImpl) getOrCreateClient(endpoint string) (interface{}, error) {
	// Check if client already cached
	if cachedClient, exists := r.clientCache[endpoint]; exists {
		return cachedClient, nil
	}

	// Create new client (simplified - in real implementation this would create
	// a proper ConnectRPC client to the endpoint)
	client := &rpcClient{
		endpoint: endpoint,
	}

	// Cache the client
	r.clientCache[endpoint] = client

	return client, nil
}

// rpcClient is a simple wrapper for RPC client communication
type rpcClient struct {
	endpoint string
}

// invokeRPCMethod invokes an RPC method on a client
// This is a simplified version - real implementation would use ConnectRPC
func invokeRPCMethod(
	ctx context.Context,
	client interface{},
	methodName string,
	request interface{},
) (interface{}, error) {
	// In real implementation, this would:
	// 1. Use reflection or code generation to find the method
	// 2. Marshal the request
	// 3. Make the HTTP/2 call
	// 4. Unmarshal the response
	// 5. Return the response or error

	// For now, return a placeholder response
	return map[string]interface{}{
		"status":  "success",
		"method":  methodName,
		"request": request,
	}, nil
}

// GetRegisteredServices returns all registered services (for debugging/testing)
func (r *RpcConnectorImpl) GetRegisteredServices() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range r.serviceRegistry {
		result[k] = v
	}

	return result
}

// ClearCache clears the client cache (for testing)
func (r *RpcConnectorImpl) ClearCache() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clientCache = make(map[string]interface{})
}
