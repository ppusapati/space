package calchelpers

import (
	"context"
	"errors"
	"strings"
	"testing"

	p9errors "p9e.in/chetana/packages/errors"
)

// A realistic calculation shape — input + result types. Kept trivial
// so tests focus on the dispatcher, not the math.
type sampleInput struct {
	EntityID string
	Factor   float64
}

type sampleResult struct {
	Class string
	Value float64
}

// ───────────────────────────────────────────────────────────────────────────
// Dispatch — happy path
// ───────────────────────────────────────────────────────────────────────────

func TestClassDispatcher_RegisterAndDispatch(t *testing.T) {
	disp := NewClassDispatcher[sampleInput, sampleResult]("test_calc")
	disp.Register("class_a", func(_ context.Context, in sampleInput) (*sampleResult, error) {
		return &sampleResult{Class: "class_a", Value: in.Factor * 2}, nil
	})
	disp.Register("class_b", func(_ context.Context, in sampleInput) (*sampleResult, error) {
		return &sampleResult{Class: "class_b", Value: in.Factor * 3}, nil
	})

	out, err := disp.Dispatch(context.Background(), "class_a", sampleInput{Factor: 10})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if out.Class != "class_a" || out.Value != 20 {
		t.Errorf("class_a result: %+v", out)
	}

	out, err = disp.Dispatch(context.Background(), "class_b", sampleInput{Factor: 10})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if out.Class != "class_b" || out.Value != 30 {
		t.Errorf("class_b result: %+v", out)
	}
}

// ───────────────────────────────────────────────────────────────────────────
// Dispatch — unknown class yields typed BadRequest with list
// ───────────────────────────────────────────────────────────────────────────

func TestClassDispatcher_UnknownClass(t *testing.T) {
	disp := NewClassDispatcher[sampleInput, sampleResult]("test_calc")
	disp.Register("class_a", func(_ context.Context, in sampleInput) (*sampleResult, error) {
		return &sampleResult{}, nil
	})

	_, err := disp.Dispatch(context.Background(), "unknown_class", sampleInput{})
	if err == nil {
		t.Fatal("expected error for unknown class")
	}
	if p9errors.Reason(err) != "CLASS_UNSUPPORTED" {
		t.Errorf("reason: %q", p9errors.Reason(err))
	}
	if !strings.Contains(err.Error(), "test_calc") {
		t.Errorf("error should name the calculation, got: %v", err)
	}
	if !strings.Contains(err.Error(), "unknown_class") {
		t.Errorf("error should name the class, got: %v", err)
	}
	if !strings.Contains(err.Error(), "class_a") {
		t.Errorf("error should list supported classes (class_a), got: %v", err)
	}
}

// ───────────────────────────────────────────────────────────────────────────
// Dispatch — nil handler is refused
// ───────────────────────────────────────────────────────────────────────────

func TestClassDispatcher_NilHandlerRefused(t *testing.T) {
	disp := NewClassDispatcher[sampleInput, sampleResult]("test_calc")
	disp.Register("class_a", nil)

	_, err := disp.Dispatch(context.Background(), "class_a", sampleInput{})
	if err == nil {
		t.Fatal("expected error for nil handler")
	}
	if p9errors.Reason(err) != "CLASS_REGISTERED_BUT_NIL" {
		t.Errorf("reason: %q", p9errors.Reason(err))
	}
}

// ───────────────────────────────────────────────────────────────────────────
// Handler errors propagate unchanged
// ───────────────────────────────────────────────────────────────────────────

func TestClassDispatcher_HandlerErrorPropagates(t *testing.T) {
	want := errors.New("inner failure")
	disp := NewClassDispatcher[sampleInput, sampleResult]("test_calc")
	disp.Register("class_a", func(_ context.Context, _ sampleInput) (*sampleResult, error) {
		return nil, want
	})
	_, err := disp.Dispatch(context.Background(), "class_a", sampleInput{})
	if err != want {
		t.Errorf("got %v, want %v", err, want)
	}
}

// ───────────────────────────────────────────────────────────────────────────
// Re-register overwrites (used by tests that want to inject variants)
// ───────────────────────────────────────────────────────────────────────────

func TestClassDispatcher_RegisterOverwrite(t *testing.T) {
	disp := NewClassDispatcher[sampleInput, sampleResult]("test_calc")
	disp.Register("class_a", func(_ context.Context, _ sampleInput) (*sampleResult, error) {
		return &sampleResult{Value: 1}, nil
	})
	disp.Register("class_a", func(_ context.Context, _ sampleInput) (*sampleResult, error) {
		return &sampleResult{Value: 2}, nil
	})
	out, err := disp.Dispatch(context.Background(), "class_a", sampleInput{})
	if err != nil {
		t.Fatal(err)
	}
	if out.Value != 2 {
		t.Errorf("got %v, want 2 (overwritten)", out.Value)
	}
}

// ───────────────────────────────────────────────────────────────────────────
// SupportedClasses is sorted
// ───────────────────────────────────────────────────────────────────────────

func TestClassDispatcher_SupportedClassesSorted(t *testing.T) {
	disp := NewClassDispatcher[sampleInput, sampleResult]("test_calc")
	disp.Register("zebra", nil)
	disp.Register("alpha", nil)
	disp.Register("mango", nil)

	got := disp.SupportedClasses()
	want := []string{"alpha", "mango", "zebra"}
	if len(got) != 3 {
		t.Fatalf("got %v", got)
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("got[%d] = %q, want %q", i, got[i], v)
		}
	}
}

// ───────────────────────────────────────────────────────────────────────────
// CalculationName passes through
// ───────────────────────────────────────────────────────────────────────────

func TestClassDispatcher_CalculationName(t *testing.T) {
	disp := NewClassDispatcher[sampleInput, sampleResult]("my_calc")
	if got := disp.CalculationName(); got != "my_calc" {
		t.Errorf("got %q", got)
	}
}

// ───────────────────────────────────────────────────────────────────────────
// Concurrent Register + Dispatch is safe
// ───────────────────────────────────────────────────────────────────────────

func TestClassDispatcher_Concurrency(t *testing.T) {
	disp := NewClassDispatcher[sampleInput, sampleResult]("test_calc")
	disp.Register("class_a", func(_ context.Context, _ sampleInput) (*sampleResult, error) {
		return &sampleResult{}, nil
	})

	// Fire parallel dispatches + registrations; failure would surface
	// as a race detector trip in -race builds.
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			disp.Register("class_b", func(_ context.Context, _ sampleInput) (*sampleResult, error) {
				return &sampleResult{}, nil
			})
		}
		close(done)
	}()
	for i := 0; i < 100; i++ {
		_, _ = disp.Dispatch(context.Background(), "class_a", sampleInput{})
	}
	<-done
}
