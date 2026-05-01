package db

import (
	"context"
	"strings"
	"testing"
)

func TestOpenRequiresDSN(t *testing.T) {
	if _, err := Open(context.Background(), Config{}); err == nil {
		t.Fatal("expected error when DSN is missing")
	}
}

func TestOpenInvalidDSN(t *testing.T) {
	_, err := Open(context.Background(), Config{DSN: "::not a url::"})
	if err == nil {
		t.Fatal("expected error from malformed DSN")
	}
	if !strings.Contains(err.Error(), "db:") {
		t.Fatalf("expected db: prefix, got %q", err)
	}
}

func TestInTxNilPoolReturnsError(t *testing.T) {
	if err := InTx(context.Background(), nil, "", nil); err == nil {
		t.Fatal("expected error for nil pool")
	}
}
