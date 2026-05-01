package closer

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunInLIFOOrder(t *testing.T) {
	var c Closer
	var order []string
	c.Add("a", func(ctx context.Context) error { order = append(order, "a"); return nil })
	c.Add("b", func(ctx context.Context) error { order = append(order, "b"); return nil })
	c.Add("c", func(ctx context.Context) error { order = append(order, "c"); return nil })
	if err := c.Run(context.Background(), time.Second); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 3 || order[0] != "c" || order[1] != "b" || order[2] != "a" {
		t.Fatalf("LIFO violated: %v", order)
	}
}

func TestRunJoinsErrors(t *testing.T) {
	var c Closer
	c.Add("a", func(ctx context.Context) error { return errors.New("ea") })
	c.Add("b", func(ctx context.Context) error { return errors.New("eb") })
	err := c.Run(context.Background(), time.Second)
	if err == nil {
		t.Fatal("expected joined error")
	}
	if msg := err.Error(); !contains(msg, "ea") || !contains(msg, "eb") {
		t.Fatalf("expected both ea and eb in %q", msg)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && (indexOf(s, sub) >= 0)))
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestRunEmpty(t *testing.T) {
	var c Closer
	if err := c.Run(context.Background(), time.Millisecond); err != nil {
		t.Fatalf("empty closer must not error: %v", err)
	}
}
