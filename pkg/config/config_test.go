package config

import (
	"errors"
	"testing"
	"time"
)

func TestStringFallback(t *testing.T) {
	t.Setenv("FOO", "")
	if got := String("FOO", "bar"); got != "bar" {
		t.Fatalf("want bar, got %q", got)
	}
	t.Setenv("FOO", "baz")
	if got := String("FOO", "bar"); got != "baz" {
		t.Fatalf("want baz, got %q", got)
	}
}

func TestMustStringMissing(t *testing.T) {
	t.Setenv("BAR", "")
	if _, err := MustString("BAR"); !errors.Is(err, ErrMissing) {
		t.Fatalf("want ErrMissing, got %v", err)
	}
}

func TestIntParsesAndFallsBack(t *testing.T) {
	t.Setenv("N", "42")
	if got := Int("N", 7); got != 42 {
		t.Fatalf("want 42, got %d", got)
	}
	t.Setenv("N", "not-a-number")
	if got := Int("N", 7); got != 7 {
		t.Fatalf("want fallback 7, got %d", got)
	}
}

func TestBoolKnownValues(t *testing.T) {
	t.Setenv("B", "yes")
	if !Bool("B", false) {
		t.Fatal("expected true for yes")
	}
	t.Setenv("B", "off")
	if Bool("B", true) {
		t.Fatal("expected false for off")
	}
}

func TestDurationParses(t *testing.T) {
	t.Setenv("D", "750ms")
	if got := Duration("D", time.Second); got != 750*time.Millisecond {
		t.Fatalf("want 750ms, got %s", got)
	}
}

func TestLoadCommonRequiresServiceName(t *testing.T) {
	t.Setenv("SERVICE_NAME", "")
	if _, err := LoadCommon(); !errors.Is(err, ErrMissing) {
		t.Fatalf("want ErrMissing, got %v", err)
	}
	t.Setenv("SERVICE_NAME", "iam")
	c, err := LoadCommon()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ServiceName != "iam" || c.HTTPAddr != ":8080" || c.MetricsAddr != ":9090" {
		t.Fatalf("unexpected defaults: %+v", c)
	}
}
