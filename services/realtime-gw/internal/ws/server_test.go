package ws

import (
	"context"
	"net/http/httptest"
	"testing"

	authzv1 "p9e.in/chetana/packages/authz/v1"

	"github.com/ppusapati/space/services/realtime-gw/internal/topic"
)

// fakeAuthorizer always allows + records calls. Implements
// topic.Authorizer.
type fakeAuthorizer struct {
	calls []string
	err   error
}

func (f *fakeAuthorizer) Authorize(_ context.Context, _ *authzv1.Principal, t string) error {
	f.calls = append(f.calls, t)
	return f.err
}

func TestNewServer_RejectsMissingDeps(t *testing.T) {
	if _, err := NewServer(Config{}); err == nil {
		t.Error("expected error for empty config")
	}
}

func TestNewServer_RejectsMissingAuthorizer(t *testing.T) {
	if _, err := NewServer(Config{Registry: NewRegistry()}); err == nil {
		t.Error("expected error for missing verifier")
	}
}

func TestRegistry_AddRemove_TracksCount(t *testing.T) {
	r := NewRegistry()
	if r.Count() != 0 {
		t.Errorf("initial count: %d", r.Count())
	}
	c := &Connection{
		principal:   &authzv1.Principal{UserID: "u1"},
		authorizer:  &fakeAuthorizer{},
		registry:    r,
		bufCapacity: 10,
	}
	r.Add(c)
	if r.Count() != 1 {
		t.Errorf("after add: %d", r.Count())
	}
	r.Remove(c)
	if r.Count() != 0 {
		t.Errorf("after remove: %d", r.Count())
	}
}

func TestConnection_SubscribeUnsubscribe(t *testing.T) {
	c := &Connection{
		principal:   &authzv1.Principal{UserID: "u1"},
		authorizer:  &fakeAuthorizer{},
		bufCapacity: 5,
	}
	if err := c.Subscribe(context.Background(), "alert.critical"); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	subs := c.Subscriptions()
	if len(subs) != 1 || subs[0] != "alert.critical" {
		t.Errorf("subs: %v", subs)
	}
	c.Unsubscribe("alert.critical")
	if len(c.Subscriptions()) != 0 {
		t.Error("unsubscribe failed")
	}
}

func TestConnection_PushOnUnsubscribedTopicNoOp(t *testing.T) {
	c := &Connection{
		principal:   &authzv1.Principal{UserID: "u1"},
		authorizer:  &fakeAuthorizer{},
		bufCapacity: 5,
	}
	// Push with no buffer = silent noop, returns true.
	if !c.Push("not.subscribed", "msg") {
		t.Error("push to unsubscribed should be a no-op (return true)")
	}
}

func TestRegistry_FanOut_DeliversAndCountsDrops(t *testing.T) {
	r := NewRegistry()
	c := &Connection{
		principal:   &authzv1.Principal{UserID: "u"},
		authorizer:  &fakeAuthorizer{},
		bufCapacity: 2,
	}
	r.Add(c)
	if err := c.Subscribe(context.Background(), "telemetry.params.frame"); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	// Three messages into a 2-slot buffer → 1 drop.
	for i := 0; i < 3; i++ {
		r.FanOut("telemetry.params.frame", i)
	}
	delivered, dropped := r.FanOut("telemetry.params.frame", "extra")
	if delivered != 1 {
		t.Errorf("delivered: %d", delivered)
	}
	if dropped != 1 {
		t.Errorf("dropped: %d", dropped)
	}
}

func TestRegistry_FanOut_SkipsUnsubscribed(t *testing.T) {
	r := NewRegistry()
	c := &Connection{
		principal:   &authzv1.Principal{UserID: "u"},
		authorizer:  &fakeAuthorizer{},
		bufCapacity: 5,
	}
	r.Add(c)
	delivered, _ := r.FanOut("alert.critical", "msg")
	if delivered != 0 {
		t.Errorf("delivered: %d", delivered)
	}
}

func TestBearerFrom_Authorization(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/rt", nil)
	req.Header.Set("Authorization", "Bearer abc.def.ghi")
	if got := bearerFrom(req); got != "abc.def.ghi" {
		t.Errorf("got %q", got)
	}
}

func TestBearerFrom_Subprotocol(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/rt", nil)
	req.Header.Set("Sec-WebSocket-Protocol", "chetana.v1, chetana.bearer.xyz.token")
	if got := bearerFrom(req); got != "xyz.token" {
		t.Errorf("got %q", got)
	}
}

func TestBearerFrom_NoToken(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/rt", nil)
	if got := bearerFrom(req); got != "" {
		t.Errorf("got %q", got)
	}
}

// Compile-time check: topic.PolicyAuthorizer + fakeAuthorizer
// both satisfy the topic.Authorizer interface the Server uses.
func TestAuthorizerInterface_Compiles(t *testing.T) {
	var _ topic.Authorizer = (*topic.PolicyAuthorizer)(nil)
	var _ topic.Authorizer = (*fakeAuthorizer)(nil)
}
