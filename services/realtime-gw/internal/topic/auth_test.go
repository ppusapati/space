package topic

import (
	"context"
	"errors"
	"testing"

	authzv1 "p9e.in/chetana/packages/authz/v1"
)

// fakePolicySource serves a static PolicySet.
type fakePolicySource struct{ set *authzv1.PolicySet }

func (f *fakePolicySource) Snapshot() *authzv1.PolicySet { return f.set }

func TestDefaultMapper_KnownClasses(t *testing.T) {
	cases := []struct {
		topic, want string
	}{
		{"telemetry.params.frame", "realtime.telemetry.subscribe"},
		{"pass.state.abc123", "realtime.pass.subscribe"},
		{"alert.critical", "realtime.alert.subscribe"},
		{"command.state.cmd1", "realtime.command.subscribe"},
		{"notify.inapp.v1", "realtime.notify.subscribe"},
		{"itar.classified.frame", "realtime.itar.subscribe"},
	}
	for _, c := range cases {
		got, err := DefaultMapper(c.topic)
		if err != nil {
			t.Errorf("DefaultMapper(%q): %v", c.topic, err)
			continue
		}
		if got != c.want {
			t.Errorf("DefaultMapper(%q): got %q want %q", c.topic, got, c.want)
		}
	}
}

func TestDefaultMapper_UnknownClass(t *testing.T) {
	if _, err := DefaultMapper("unknown.foo.bar"); err == nil {
		t.Error("expected error for unknown topic class")
	}
	if _, err := DefaultMapper(""); err == nil {
		t.Error("expected error for empty topic")
	}
}

func TestNewPolicyAuthorizer_RejectsNilPolicies(t *testing.T) {
	if _, err := NewPolicyAuthorizer(nil, nil); err == nil {
		t.Error("expected error for nil policy source")
	}
}

func TestPolicyAuthorizer_AllowsKnownTopic(t *testing.T) {
	set, _ := authzv1.NewPolicySet([]authzv1.Policy{{
		ID:         "telemetry-allow",
		Effect:     authzv1.EffectAllow,
		Permission: "realtime.telemetry.subscribe",
		Roles:      []string{"operator"},
	}})
	a, err := NewPolicyAuthorizer(&fakePolicySource{set: set}, nil)
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	p := &authzv1.Principal{UserID: "u", TenantID: "t", Roles: []string{"operator"}}
	if err := a.Authorize(context.Background(), p, "telemetry.params.frame"); err != nil {
		t.Errorf("expected allow: %v", err)
	}
}

// REQ-FUNC-RT-002: ITAR topics deny non-US-persons with a typed close code.
func TestPolicyAuthorizer_ITARDeniesNonUSPerson(t *testing.T) {
	set, _ := authzv1.NewPolicySet([]authzv1.Policy{
		{
			ID:         "itar-allow",
			Effect:     authzv1.EffectAllow,
			Priority:   100,
			Permission: "realtime.itar.subscribe",
			Roles:      []string{"operator"},
		},
		{
			ID:              "itar-deny",
			Effect:          authzv1.EffectDeny,
			Priority:        1000,
			Permission:      "realtime.itar.subscribe",
			RequireUSPerson: true,
		},
	})
	a, err := NewPolicyAuthorizer(&fakePolicySource{set: set}, nil)
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	non := &authzv1.Principal{UserID: "u", TenantID: "t", Roles: []string{"operator"}, IsUSPerson: false}
	err = a.Authorize(context.Background(), non, "itar.classified.frame")
	deny, ok := IsDeny(err)
	if !ok {
		t.Fatalf("expected DenyError, got %v", err)
	}
	if deny.Close.Code != CloseITARRequiresUSP.Code {
		t.Errorf("close code: got %d want %d", deny.Close.Code, CloseITARRequiresUSP.Code)
	}

	usp := &authzv1.Principal{UserID: "u", TenantID: "t", Roles: []string{"operator"}, IsUSPerson: true}
	if err := a.Authorize(context.Background(), usp, "itar.classified.frame"); err != nil {
		t.Errorf("US-person should be allowed: %v", err)
	}
}

func TestPolicyAuthorizer_UnknownTopicReturnsTypedClose(t *testing.T) {
	set, _ := authzv1.NewPolicySet(nil)
	a, _ := NewPolicyAuthorizer(&fakePolicySource{set: set}, nil)
	err := a.Authorize(context.Background(), &authzv1.Principal{UserID: "u"}, "weird.unknown.topic")
	deny, ok := IsDeny(err)
	if !ok {
		t.Fatalf("expected DenyError, got %v", err)
	}
	if deny.Close.Code != CloseUnknownTopic.Code {
		t.Errorf("close code: got %d want %d", deny.Close.Code, CloseUnknownTopic.Code)
	}
}

func TestPolicyAuthorizer_NoPrincipalDenies(t *testing.T) {
	set, _ := authzv1.NewPolicySet(nil)
	a, _ := NewPolicyAuthorizer(&fakePolicySource{set: set}, nil)
	err := a.Authorize(context.Background(), nil, "alert.critical")
	if _, ok := IsDeny(err); !ok {
		t.Errorf("expected DenyError, got %v", err)
	}
}

func TestIsDeny_DistinguishesErrorTypes(t *testing.T) {
	if _, ok := IsDeny(errors.New("plain")); ok {
		t.Error("plain error should not match")
	}
	d := &DenyError{Topic: "t", Close: ClosePolicyDeny, Reason: "x"}
	if _, ok := IsDeny(d); !ok {
		t.Error("DenyError should match")
	}
}
