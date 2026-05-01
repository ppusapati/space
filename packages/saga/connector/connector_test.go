// Package connector provides tests for RPC connectivity
package connector

import (
	"context"
	"testing"
)

// TestGetServiceEndpoint tests service endpoint retrieval
func TestGetServiceEndpoint(t *testing.T) {
	registry := NewDefaultServiceRegistry(map[string]string{
		"sales-order": "http://localhost:8119",
		"inventory":   "http://localhost:8179",
	})

	tests := []struct {
		name        string
		serviceName string
		expected    string
		shouldExist bool
	}{
		{"existing service", "sales-order", "http://localhost:8119", true},
		{"another existing", "inventory", "http://localhost:8179", true},
		{"non-existent", "unknown-service", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := registry.GetServiceEndpoint(tt.serviceName)
			if tt.shouldExist {
				if endpoint != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, endpoint)
				}
			} else {
				if endpoint != "" {
					t.Errorf("expected empty string, got %s", endpoint)
				}
			}
		})
	}
}

// TestRegisterService tests service registration
func TestRegisterService(t *testing.T) {
	registry := NewDefaultServiceRegistry(map[string]string{})

	err := registry.RegisterService("sales-order", "http://localhost:8119")
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	endpoint := registry.GetServiceEndpoint("sales-order")
	if endpoint != "http://localhost:8119" {
		t.Errorf("expected http://localhost:8119, got %s", endpoint)
	}
}

// TestConnectorInvokeHandler tests RPC handler invocation
func TestConnectorInvokeHandler(t *testing.T) {
	registry := NewDefaultServiceRegistry(map[string]string{
		"test-service": "http://localhost:9999",
	})

	connector := NewConnectRPCConnector(registry)

	ctx := context.Background()

	// This test will fail with connection error since we don't have a real service
	// but it validates the basic invocation flow
	_, err := connector.InvokeHandler(
		ctx,
		"test-service",
		"TestMethod",
		map[string]interface{}{"test": "data"},
	)

	// We expect an error since there's no service running
	if err == nil {
		t.Error("expected error from failed connection")
	}
}

// TestConnectorGetServiceEndpoint tests endpoint retrieval via connector
func TestConnectorGetServiceEndpoint(t *testing.T) {
	registry := NewDefaultServiceRegistry(map[string]string{
		"sales-order": "http://localhost:8119",
	})

	connector := NewConnectRPCConnector(registry)

	endpoint, err := connector.GetServiceEndpoint("sales-order")
	if err != nil {
		t.Fatalf("failed to get endpoint: %v", err)
	}

	if endpoint != "http://localhost:8119" {
		t.Errorf("expected http://localhost:8119, got %s", endpoint)
	}
}

// TestConnectorGetMissingService tests endpoint retrieval for missing service
func TestConnectorGetMissingService(t *testing.T) {
	registry := NewDefaultServiceRegistry(map[string]string{})

	connector := NewConnectRPCConnector(registry)

	_, err := connector.GetServiceEndpoint("unknown-service")
	if err == nil {
		t.Error("expected error for missing service")
	}
}

// TestConnectorRegisterService tests service registration via connector
func TestConnectorRegisterService(t *testing.T) {
	registry := NewDefaultServiceRegistry(map[string]string{})

	connector := NewConnectRPCConnector(registry)

	err := connector.RegisterService("sales-order", "http://localhost:8119")
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	endpoint, err := connector.GetServiceEndpoint("sales-order")
	if err != nil {
		t.Fatalf("failed to get endpoint: %v", err)
	}

	if endpoint != "http://localhost:8119" {
		t.Errorf("expected http://localhost:8119, got %s", endpoint)
	}
}

// TestStringReader tests string reading functionality
func TestStringReader(t *testing.T) {
	reader := NewStringReader("hello world")

	buf := make([]byte, 5)
	n, err := reader.Read(buf)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if n != 5 {
		t.Errorf("expected 5 bytes, got %d", n)
	}

	if string(buf) != "hello" {
		t.Errorf("expected 'hello', got %s", string(buf))
	}
}
