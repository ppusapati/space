package errs

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
)

func TestDomainToConnectCode(t *testing.T) {
	cases := []struct {
		d    Domain
		want connect.Code
	}{
		{DomainNotFound, connect.CodeNotFound},
		{DomainAlreadyExists, connect.CodeAlreadyExists},
		{DomainInvalidArgument, connect.CodeInvalidArgument},
		{DomainPreconditionFailed, connect.CodeFailedPrecondition},
		{DomainPermissionDenied, connect.CodePermissionDenied},
		{DomainUnauthenticated, connect.CodeUnauthenticated},
		{DomainResourceExhausted, connect.CodeResourceExhausted},
		{DomainUnavailable, connect.CodeUnavailable},
		{DomainCanceled, connect.CodeCanceled},
		{DomainDeadlineExceeded, connect.CodeDeadlineExceeded},
		{DomainUnknown, connect.CodeInternal},
	}
	for _, c := range cases {
		if got := Code(c.d); got != c.want {
			t.Fatalf("Code(%d) = %v, want %v", c.d, got, c.want)
		}
	}
}

func TestNewAndWrap(t *testing.T) {
	root := errors.New("boom")
	e := Wrap(DomainNotFound, root, "user %s", "alice")
	if e == nil {
		t.Fatal("Wrap returned nil for non-nil cause")
	}
	if !errors.Is(e, root) {
		t.Fatal("errors.Is failed to unwrap")
	}
	if Wrap(DomainNotFound, nil, "x") != nil {
		t.Fatal("Wrap should return nil for nil cause")
	}
	n := New(DomainAlreadyExists, "duplicate %d", 7)
	if n.Domain != DomainAlreadyExists || n.Error() != "duplicate 7" {
		t.Fatalf("unexpected New result: %+v", n)
	}
}

func TestToConnectPreservesDomain(t *testing.T) {
	e := New(DomainPermissionDenied, "no")
	c := ToConnect(e)
	if c == nil || c.Code() != connect.CodePermissionDenied {
		t.Fatalf("ToConnect dropped domain: %v", c)
	}
	c2 := ToConnect(errors.New("plain"))
	if c2 == nil || c2.Code() != connect.CodeInternal {
		t.Fatalf("plain error must map to Internal: %v", c2)
	}
	if ToConnect(nil) != nil {
		t.Fatal("nil in must produce nil out")
	}
}
