// Package topic implements per-topic ABAC for the realtime gateway.
//
// → REQ-FUNC-RT-002 (per-topic ABAC; ITAR topics require
//                     `is_us_person`).
// → REQ-CONST-011 (NO ad-hoc role check — every topic-auth
//                   decision routes through `authz/v1.Decide`
//                   against the chetana policy set).
//
// Topic naming convention:
//
//   <module>.<resource>.<facet>     e.g. telemetry.params.frame
//   pass.state.{passid}             e.g. pass.state.abc123
//   alert.<severity>                e.g. alert.critical
//   command.state.{cmdid}
//   notify.inapp.v1                 (special — every authenticated
//                                    user can subscribe; the
//                                    notify-svc producer scopes
//                                    by user_id in the message)
//
// Each topic maps deterministically to a permission identifier of
// the form `realtime.<topic-class>.subscribe`. The mapper's job
// is the topic-class extraction; the chetana policy set then
// decides via the standard RBAC + clearance + ITAR pipeline.

package topic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	authzv1 "p9e.in/chetana/packages/authz/v1"
)

// CloseReason enumerates the typed close reasons the WS layer
// emits when topic auth denies a subscription. Surfaced to the
// client as a WebSocket close frame with a chetana-defined code +
// human-readable phrase (REQ-FUNC-RT-002 acceptance #2: typed
// close code on ITAR denial).
type CloseReason struct {
	Code   int
	Reason string
}

// Canonical close codes (out of the 4000-4999 application range).
var (
	ClosePolicyDeny       = CloseReason{Code: 4001, Reason: "policy_deny"}
	CloseITARRequiresUSP  = CloseReason{Code: 4002, Reason: "itar_requires_us_person"}
	CloseClearance        = CloseReason{Code: 4003, Reason: "insufficient_clearance"}
	CloseUnknownTopic     = CloseReason{Code: 4004, Reason: "unknown_topic"}
)

// Authorizer is the topic-auth surface the WS layer calls.
type Authorizer interface {
	// Authorize returns nil when the principal is permitted to
	// subscribe to `topic`; otherwise a *DenyError.
	Authorize(ctx context.Context, principal *authzv1.Principal, topic string) error
}

// PolicyAuthorizer is the production Authorizer. It maps
// `topic` → permission via `Mapper` then delegates to authzv1.Decide
// against the chetana PolicySet.
type PolicyAuthorizer struct {
	policies authzv1.PolicySource
	mapper   Mapper
}

// PolicySource mirrors authzv1.PolicySource so we don't pull
// the entire interceptor surface into this package.
type PolicySource = authzv1.PolicySource

// NewPolicyAuthorizer wires the dependencies.
func NewPolicyAuthorizer(p PolicySource, m Mapper) (*PolicyAuthorizer, error) {
	if p == nil {
		return nil, errors.New("topic: nil policy source")
	}
	if m == nil {
		m = DefaultMapper
	}
	return &PolicyAuthorizer{policies: p, mapper: m}, nil
}

// Authorize implements Authorizer.
func (a *PolicyAuthorizer) Authorize(ctx context.Context, principal *authzv1.Principal, topic string) error {
	if principal == nil {
		return &DenyError{Topic: topic, Close: ClosePolicyDeny, Reason: "no principal"}
	}
	permission, err := a.mapper(topic)
	if err != nil {
		return &DenyError{Topic: topic, Close: CloseUnknownTopic, Reason: err.Error()}
	}
	decision, derr := authzv1.Decide(principal, authzv1.Request{
		Permission: permission,
		TenantID:   principal.TenantID,
	}, a.policies.Snapshot())
	if derr != nil {
		return &DenyError{Topic: topic, Close: ClosePolicyDeny, Reason: derr.Error()}
	}
	if decision.Effect == authzv1.EffectAllow {
		return nil
	}
	// Map the canonical reason string to the most informative
	// close code so a client knows whether to prompt the user
	// for re-auth (clearance) vs hard-stop (ITAR).
	switch decision.Reason {
	case authzv1.ReasonITAR:
		return &DenyError{Topic: topic, Close: CloseITARRequiresUSP, Reason: decision.Reason}
	case authzv1.ReasonClearance:
		return &DenyError{Topic: topic, Close: CloseClearance, Reason: decision.Reason}
	}
	return &DenyError{Topic: topic, Close: ClosePolicyDeny, Reason: decision.Reason}
}

// Mapper translates `topic` → `{module}.{resource}.{action}`
// permission identifier the policy set knows about.
type Mapper func(topic string) (string, error)

// DefaultMapper covers the chetana-shipped topic taxonomy.
var DefaultMapper Mapper = func(topic string) (string, error) {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return "", errors.New("topic: empty topic")
	}
	parts := strings.SplitN(topic, ".", 3)
	switch parts[0] {
	case "telemetry":
		// telemetry.params.frame → realtime.telemetry.subscribe
		return "realtime.telemetry.subscribe", nil
	case "pass":
		// pass.state.{passid} → realtime.pass.subscribe
		return "realtime.pass.subscribe", nil
	case "alert":
		// alert.<severity> → realtime.alert.subscribe
		return "realtime.alert.subscribe", nil
	case "command":
		// command.state.{cmdid} → realtime.command.subscribe
		return "realtime.command.subscribe", nil
	case "notify":
		// notify.inapp.v1 → realtime.notify.subscribe
		return "realtime.notify.subscribe", nil
	case "itar":
		// Any explicitly-itar-classified topic routes to a
		// dedicated permission so the ITAR deny gate fires.
		return "realtime.itar.subscribe", nil
	}
	return "", fmt.Errorf("topic: unknown class %q", parts[0])
}

// DenyError is the structured error every Authorize call returns
// on rejection. Carries the close-frame the WS layer should emit.
type DenyError struct {
	Topic  string
	Close  CloseReason
	Reason string
}

// Error implements error.
func (e *DenyError) Error() string {
	return fmt.Sprintf("topic: %s denied: %s (close=%d %s)",
		e.Topic, e.Reason, e.Close.Code, e.Close.Reason)
}

// IsDeny reports whether `err` is a DenyError. Convenience for
// callers that want the typed close-frame.
func IsDeny(err error) (*DenyError, bool) {
	var d *DenyError
	if errors.As(err, &d) {
		return d, true
	}
	return nil, false
}
