// Package money provides conversion utilities between protobuf Money types and Go decimal types.
package money

import (
	"github.com/shopspring/decimal"

	pb "p9e.in/samavaya/packages/api/v1/money"
)

// DefaultDecimalPlaces is the number of decimal places for minor units (paise/cents).
const DefaultDecimalPlaces = 2

// =============================================================================
// Money ↔ Decimal Conversions
// =============================================================================

// DecimalFromMoney converts *pb.Money to decimal.Decimal.
// Returns zero decimal if Money is nil.
// Converts from minor units (paise/cents) to decimal (e.g., 10050 -> 100.50).
func DecimalFromMoney(m *pb.Money) decimal.Decimal {
	if m == nil {
		return decimal.Zero
	}
	// AmountMinor is in paise/cents, divide by 100 to get the decimal value
	return decimal.NewFromInt(m.AmountMinor).Shift(-DefaultDecimalPlaces)
}

// MoneyFromDecimal converts decimal.Decimal to *pb.Money.
// Converts from decimal to minor units (e.g., 100.50 -> 10050).
func MoneyFromDecimal(d decimal.Decimal, currency string) *pb.Money {
	// Shift by 2 decimal places to convert to minor units (paise/cents)
	minorUnits := d.Shift(DefaultDecimalPlaces).IntPart()
	return &pb.Money{
		Currency:    currency,
		AmountMinor: minorUnits,
	}
}

// =============================================================================
// Money ↔ *Decimal Pointer Conversions
// =============================================================================

// DecimalPtrFromMoney converts *pb.Money to *decimal.Decimal.
// Returns nil if Money is nil.
func DecimalPtrFromMoney(m *pb.Money) *decimal.Decimal {
	if m == nil {
		return nil
	}
	d := decimal.NewFromInt(m.AmountMinor).Shift(-DefaultDecimalPlaces)
	return &d
}

// MoneyFromDecimalPtr converts *decimal.Decimal to *pb.Money.
// Returns nil if decimal pointer is nil.
func MoneyFromDecimalPtr(d *decimal.Decimal, currency string) *pb.Money {
	if d == nil {
		return nil
	}
	minorUnits := d.Shift(DefaultDecimalPlaces).IntPart()
	return &pb.Money{
		Currency:    currency,
		AmountMinor: minorUnits,
	}
}

// =============================================================================
// Money ↔ Int64 (Minor Units) Conversions
// =============================================================================

// Int64FromMoney extracts the amount_minor from *pb.Money.
// Returns 0 if Money is nil.
func Int64FromMoney(m *pb.Money) int64 {
	if m == nil {
		return 0
	}
	return m.AmountMinor
}

// MoneyFromInt64 creates *pb.Money from minor units (paise/cents).
func MoneyFromInt64(amountMinor int64, currency string) *pb.Money {
	return &pb.Money{
		Currency:    currency,
		AmountMinor: amountMinor,
	}
}

// =============================================================================
// Currency Extraction
// =============================================================================

// CurrencyFromMoney extracts the currency code from *pb.Money.
// Returns empty string if Money is nil.
func CurrencyFromMoney(m *pb.Money) string {
	if m == nil {
		return ""
	}
	return m.Currency
}

// =============================================================================
// String ↔ Decimal Conversions
// =============================================================================

// DecimalPtrFromString converts a string to *decimal.Decimal.
// Returns nil if string is empty or invalid.
func DecimalPtrFromString(s string) *decimal.Decimal {
	if s == "" {
		return nil
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return nil
	}
	return &d
}

// StringFromDecimalPtr converts *decimal.Decimal to string.
// Returns empty string if decimal pointer is nil.
func StringFromDecimalPtr(d *decimal.Decimal) string {
	if d == nil {
		return ""
	}
	return d.String()
}

// DecimalFromString converts a string to decimal.Decimal.
// Returns zero decimal if string is empty or invalid.
func DecimalFromString(s string) decimal.Decimal {
	if s == "" {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	return d
}

// StringFromDecimal converts decimal.Decimal to string.
func StringFromDecimal(d decimal.Decimal) string {
	return d.String()
}

// =============================================================================
// Int64 ↔ Decimal Conversions
// =============================================================================

// DecimalFromInt64 converts int64 to decimal.Decimal.
func DecimalFromInt64(i int64) decimal.Decimal {
	return decimal.NewFromInt(i)
}

// DecimalPtrFromInt64 converts int64 to *decimal.Decimal.
func DecimalPtrFromInt64(i int64) *decimal.Decimal {
	d := decimal.NewFromInt(i)
	return &d
}

// Int64FromDecimal converts decimal.Decimal to int64 (truncates).
func Int64FromDecimal(d decimal.Decimal) int64 {
	return d.IntPart()
}

// Int64FromDecimalPtr converts *decimal.Decimal to int64.
// Returns 0 if pointer is nil.
func Int64FromDecimalPtr(d *decimal.Decimal) int64 {
	if d == nil {
		return 0
	}
	return d.IntPart()
}

// =============================================================================
// Int32 ↔ Decimal Conversions
// =============================================================================

// DecimalFromInt32 converts int32 to decimal.Decimal.
func DecimalFromInt32(i int32) decimal.Decimal {
	return decimal.NewFromInt32(i)
}

// DecimalPtrFromInt32 converts int32 to *decimal.Decimal.
func DecimalPtrFromInt32(i int32) *decimal.Decimal {
	d := decimal.NewFromInt32(i)
	return &d
}

// Int32FromDecimal converts decimal.Decimal to int32 (truncates).
func Int32FromDecimal(d decimal.Decimal) int32 {
	return int32(d.IntPart())
}

// Int32FromDecimalPtr converts *decimal.Decimal to int32.
// Returns 0 if pointer is nil.
func Int32FromDecimalPtr(d *decimal.Decimal) int32 {
	if d == nil {
		return 0
	}
	return int32(d.IntPart())
}

// =============================================================================
// Int32 ↔ String Conversions
// =============================================================================

// StringFromInt32Ptr converts *int32 to string.
// Returns empty string if pointer is nil.
func StringFromInt32Ptr(i *int32) string {
	if i == nil {
		return ""
	}
	return decimal.NewFromInt32(*i).String()
}

// Int32PtrFromString converts string to *int32.
// Returns nil if string is empty or invalid.
func Int32PtrFromString(s string) *int32 {
	if s == "" {
		return nil
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return nil
	}
	val := int32(d.IntPart())
	return &val
}

// =============================================================================
// Validation Helpers
// =============================================================================

// IsZero checks if Money represents zero amount.
func IsZero(m *pb.Money) bool {
	return m == nil || m.AmountMinor == 0
}

// IsPositive checks if Money represents a positive amount.
func IsPositive(m *pb.Money) bool {
	return m != nil && m.AmountMinor > 0
}

// IsNegative checks if Money represents a negative amount.
func IsNegative(m *pb.Money) bool {
	return m != nil && m.AmountMinor < 0
}
