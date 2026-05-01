package sql

import (
	"math"
	"testing"
)

// ============================================================================
// Regression tests for NumericFromFloat64 / NumericFromFloat64Ptr.
//
// Why these tests exist: an earlier revision of NumericFromFloat64 called
// pgtype.Numeric.Scan(f) directly. In pgx/v5 that returns an error and
// leaves the Numeric with Valid=false — i.e. silently NULL. Money fields
// across the asset / depreciation / equipment mappers were landing in the
// DB as NULL. The helper was fixed to route through decimal.NewFromFloat.
// These tests lock down the contract so the bug can't be reintroduced.
// ============================================================================

func TestNumericFromFloat64_RoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   float64
	}{
		{"zero", 0},
		{"positive integer", 100},
		{"positive fractional", 10000.50},
		{"small fractional", 0.01},
		{"negative integer", -42},
		{"negative fractional", -1234.56},
		{"very small", 1e-10},
		{"large", 9999999.99},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n := NumericFromFloat64(tc.in)
			if !n.Valid {
				t.Fatalf("NumericFromFloat64(%v) returned Valid=false; round-trip impossible", tc.in)
			}
			got := Float64FromNumeric(n)
			// Allow a tiny delta: decimal.NewFromFloat → pgtype.Numeric →
			// Float64Value can drop the last ULP on values that aren't
			// exactly representable as base-2 floats.
			if math.Abs(got-tc.in) > 1e-9 {
				t.Errorf("round-trip for %v: got %v (delta %v)", tc.in, got, got-tc.in)
			}
		})
	}
}

// TestNumericFromFloat64_Unrepresentable covers the three IEEE-754 sentinel
// values that cannot map to a SQL NUMERIC. They must produce an invalid
// Numeric (which maps to NULL on the wire) rather than a corrupt value
// or a panic.
func TestNumericFromFloat64_Unrepresentable(t *testing.T) {
	cases := []struct {
		name string
		in   float64
	}{
		{"NaN", math.NaN()},
		{"+Inf", math.Inf(1)},
		{"-Inf", math.Inf(-1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n := NumericFromFloat64(tc.in)
			if n.Valid {
				t.Errorf("NumericFromFloat64(%v) should return Valid=false, got Valid=true", tc.in)
			}
		})
	}
}

// TestNumericFromFloat64Ptr_Nil locks the nil-input contract: nil pointer
// must return an invalid Numeric (maps to NULL) rather than zero-value or
// a panic.
func TestNumericFromFloat64Ptr_Nil(t *testing.T) {
	n := NumericFromFloat64Ptr(nil)
	if n.Valid {
		t.Errorf("NumericFromFloat64Ptr(nil) should return Valid=false, got Valid=true")
	}
}

// TestNumericFromFloat64Ptr_RoundTrip matches the non-ptr round-trip for
// a non-nil pointer; serves as a canary for drift between the two
// helpers.
func TestNumericFromFloat64Ptr_RoundTrip(t *testing.T) {
	in := 10000.50
	n := NumericFromFloat64Ptr(&in)
	if !n.Valid {
		t.Fatalf("NumericFromFloat64Ptr(&%v) returned Valid=false", in)
	}
	got := Float64FromNumeric(n)
	if math.Abs(got-in) > 1e-9 {
		t.Errorf("round-trip for %v: got %v (delta %v)", in, got, got-in)
	}
}

// TestFloat64FromNumeric_InvalidReturnsZero confirms the read-side
// behaviour the mappers rely on: an invalid Numeric is treated as zero,
// not as an error. Callers that need to distinguish NULL from zero use
// Float64PtrFromNumeric instead.
func TestFloat64FromNumeric_InvalidReturnsZero(t *testing.T) {
	var invalid = NumericFromFloat64(math.NaN()) // Valid=false by construction
	if got := Float64FromNumeric(invalid); got != 0 {
		t.Errorf("Float64FromNumeric(invalid) = %v, want 0", got)
	}
}
