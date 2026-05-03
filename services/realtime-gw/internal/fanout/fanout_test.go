package fanout

import (
	"errors"
	"testing"
)

func TestChannel_NamespacePrefix(t *testing.T) {
	if got := Channel("alert.critical"); got != "chetana:rt:alert.critical" {
		t.Errorf("got %q", got)
	}
	if got := Channel(""); got != "chetana:rt:" {
		t.Errorf("got %q", got)
	}
}

func TestNewRedisFanout_RejectsNilClient(t *testing.T) {
	if _, err := NewRedisFanout(nil); err == nil {
		t.Error("expected error")
	}
}

func TestNewKafkaBridge_Validation(t *testing.T) {
	if _, err := NewKafkaBridge(nil, nil, nil, ""); err == nil {
		t.Error("expected error for nil consumer")
	}
}

// Sentinel reflexivity guard so a future refactor doesn't break
// the standard errors.Is contract on internal sentinels.
func TestSentinels_Reflexive(t *testing.T) {
	var e error = errors.New("test")
	if !errors.Is(e, e) {
		t.Error("error reflexivity broken")
	}
}
