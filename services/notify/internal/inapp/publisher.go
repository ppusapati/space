// Package inapp publishes in-app notifications onto the Kafka
// `notify.inapp.v1` topic, which the realtime gateway consumes
// and fans out to subscribed user-agent connections.
//
// → REQ-FUNC-PLT-NOTIFY-001 (in-app channel).
//
// Same abstract-interface pattern as email + sms: the cmd layer
// wires a real Kafka producer; tests pass CapturingPublisher.

package inapp

import (
	"context"
	"errors"
	"strings"
	"time"
)

// Topic is the canonical Kafka topic name. Consumed by
// services/realtime-gw/internal/fanout/kafka.go (TASK-P1-RT-001).
const Topic = "notify.inapp.v1"

// Message is the per-call payload.
type Message struct {
	UserID    string            // recipient user (the realtime-gw filters on this)
	TenantID  string
	Title     string
	Body      string
	Severity  string // "info" | "warn" | "critical"
	Action    string // optional CTA (URL or RPC procedure name)
	OccurredAt time.Time
	Metadata  map[string]string
}

// Publisher is the abstract surface.
type Publisher interface {
	Publish(ctx context.Context, msg Message) error
}

// CapturingPublisher records every Publish call.
type CapturingPublisher struct {
	Sent []Message
	Err  error
}

// Publish implements Publisher.
func (c *CapturingPublisher) Publish(_ context.Context, msg Message) error {
	if c.Err != nil {
		return c.Err
	}
	c.Sent = append(c.Sent, msg)
	return nil
}

// Validate runs the per-call sanity checks.
func Validate(m Message) error {
	if m.UserID == "" {
		return errors.New("inapp: UserID is required")
	}
	if strings.TrimSpace(m.Title) == "" {
		return errors.New("inapp: Title is required")
	}
	if strings.TrimSpace(m.Body) == "" {
		return errors.New("inapp: Body is required")
	}
	switch m.Severity {
	case "info", "warn", "critical":
		// ok
	case "":
		// caller-defaultable; the dispatcher fills "info"
	default:
		return errors.New("inapp: Severity must be info|warn|critical")
	}
	return nil
}
