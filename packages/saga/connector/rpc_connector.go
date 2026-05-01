// Package connector provides RPC connectivity for saga step execution
package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"p9e.in/samavaya/packages/saga"
)

// ConnectRPCConnector implements the saga.RpcConnector interface using HTTP clients
// It maintains a cache of HTTP clients per service to avoid recreating connections
type ConnectRPCConnector struct {
	registry     ServiceRegistry
	clientCache  map[string]*http.Client // Service name → HTTP client
	clientMutex  sync.RWMutex
}

// ServiceRegistry defines the contract for service discovery
type ServiceRegistry interface {
	GetServiceEndpoint(serviceName string) string
	RegisterService(serviceName string, endpoint string) error
}

// DefaultServiceRegistry implements ServiceRegistry using a static map
type DefaultServiceRegistry struct {
	endpoints map[string]string
	mu        sync.RWMutex
}

// NewDefaultServiceRegistry creates a new service registry with provided endpoints
func NewDefaultServiceRegistry(endpoints map[string]string) *DefaultServiceRegistry {
	return &DefaultServiceRegistry{
		endpoints: endpoints,
	}
}

// GetServiceEndpoint returns the endpoint for a service
func (r *DefaultServiceRegistry) GetServiceEndpoint(serviceName string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.endpoints[serviceName]
}

// RegisterService registers a service endpoint
func (r *DefaultServiceRegistry) RegisterService(serviceName string, endpoint string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.endpoints[serviceName] = endpoint
	return nil
}

// NewConnectRPCConnector creates a new ConnectRPC connector with service registry
func NewConnectRPCConnector(registry ServiceRegistry) saga.RpcConnector {
	return &ConnectRPCConnector{
		registry:    registry,
		clientCache: make(map[string]*http.Client),
	}
}

// InvokeHandler invokes a service handler via HTTP POST
// serviceName: e.g., "sales-order"
// method: e.g., "CreateOrder"
// input: request body (will be JSON-serialized)
// returns: response body (as json.RawMessage)
func (c *ConnectRPCConnector) InvokeHandler(
	ctx context.Context,
	serviceName string,
	method string,
	input interface{},
) (interface{}, error) {
	// 1. Get service endpoint
	endpoint := c.registry.GetServiceEndpoint(serviceName)
	if endpoint == "" {
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	// 2. Get or create HTTP client for service
	client := c.getOrCreateClient(serviceName)

	// 3. Build request URL (ConnectRPC convention: /package.Service/Method)
	// For sagas, we use simplified path: /service/method
	url := fmt.Sprintf("%s/%s/%s", endpoint, serviceName, method)

	// 4. Marshal input to JSON
	jsonInput, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	// 5. Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 6. Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connect-Protocol-Version", "1")

	// 7. Set request body — use the local NewReadCloser helper; the
	// stdlib doesn't expose a ReadCloser constructor at this path.
	req.Body = NewReadCloser(string(jsonInput))

	// 8. Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke handler: %w", err)
	}
	defer resp.Body.Close()

	// 9. Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("handler returned status %d", resp.StatusCode)
	}

	// 10. Parse response
	var result interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetServiceEndpoint returns the endpoint URL for a service
func (c *ConnectRPCConnector) GetServiceEndpoint(serviceName string) (string, error) {
	endpoint := c.registry.GetServiceEndpoint(serviceName)
	if endpoint == "" {
		return "", fmt.Errorf("service not found: %s", serviceName)
	}
	return endpoint, nil
}

// RegisterService registers a service endpoint
func (c *ConnectRPCConnector) RegisterService(serviceName string, endpoint string) error {
	return c.registry.RegisterService(serviceName, endpoint)
}

// getOrCreateClient returns a cached HTTP client for the service, creating one if needed
func (c *ConnectRPCConnector) getOrCreateClient(serviceName string) *http.Client {
	c.clientMutex.RLock()
	if client, ok := c.clientCache[serviceName]; ok {
		c.clientMutex.RUnlock()
		return client
	}
	c.clientMutex.RUnlock()

	// Create new client (with default timeout settings)
	client := &http.Client{
		Timeout: 0, // Timeouts managed by saga engine
	}

	c.clientMutex.Lock()
	c.clientCache[serviceName] = client
	c.clientMutex.Unlock()

	return client
}

// NewReadCloser creates a ReadCloser from a string
// This is a helper to work with http.Request.Body. ReadCloser lives in
// io — stdlib net/http re-exposes it but returning through io. keeps
// this helper usable outside http contexts too.
func NewReadCloser(s string) io.ReadCloser {
	return &readCloser{
		reader: NewStringReader(s),
	}
}

type readCloser struct {
	reader *StringReader
}

func (rc *readCloser) Read(p []byte) (int, error) {
	return rc.reader.Read(p)
}

func (rc *readCloser) Close() error {
	return nil
}

// StringReader implements io.Reader for strings
type StringReader struct {
	data   string
	offset int
}

// NewStringReader creates a new StringReader
func NewStringReader(s string) *StringReader {
	return &StringReader{data: s, offset: 0}
}

// Read reads from the string
func (sr *StringReader) Read(p []byte) (int, error) {
	if sr.offset >= len(sr.data) {
		return 0, fmt.Errorf("EOF")
	}

	n := copy(p, sr.data[sr.offset:])
	sr.offset += n
	return n, nil
}
