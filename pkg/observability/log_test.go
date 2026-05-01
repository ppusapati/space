package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestLoggerEmitsServiceFields(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(LogConfig{Level: "info", Service: "iam", Environment: "dev", Writer: &buf})
	l.Info("hello", "k", "v")
	var rec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if rec["service.name"] != "iam" || rec["service.environment"] != "dev" {
		t.Fatalf("missing service fields: %v", rec)
	}
	if rec["msg"] != "hello" {
		t.Fatalf("missing msg: %v", rec)
	}
}

func TestLoggerLevelFilters(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(LogConfig{Level: "warn", Service: "x", Environment: "dev", Writer: &buf})
	l.Info("info-skipped")
	l.Warn("warn-kept")
	if strings.Contains(buf.String(), "info-skipped") {
		t.Fatal("info record leaked through warn-level filter")
	}
	if !strings.Contains(buf.String(), "warn-kept") {
		t.Fatal("warn record was filtered")
	}
}

func TestContextRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(LogConfig{Level: "info", Service: "x", Environment: "dev", Writer: &buf})
	ctx := WithLogger(context.Background(), l)
	got := LoggerFromContext(ctx)
	if got != l {
		t.Fatal("LoggerFromContext did not return the stored logger")
	}
	def := LoggerFromContext(context.Background())
	if def == nil {
		t.Fatal("LoggerFromContext should never return nil")
	}
}
