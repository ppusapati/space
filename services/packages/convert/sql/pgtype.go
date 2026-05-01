package sql

import (
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// ==================== Numeric Conversions ====================

// Float64FromNumeric converts pgtype.Numeric to float64.
// Returns 0 if the Numeric is not valid.
func Float64FromNumeric(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	return f.Float64
}

// Float64PtrFromNumeric converts pgtype.Numeric to *float64.
// Returns nil if the Numeric is not valid.
func Float64PtrFromNumeric(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	f, _ := n.Float64Value()
	return &f.Float64
}

// NumericFromFloat64 converts float64 to pgtype.Numeric.
//
// Implementation detail: pgx/v5's pgtype.Numeric.Scan only accepts string
// inputs — passing a float64 silently returns ErrScanTargetTypeChanged and
// leaves Valid=false, which Postgres would then persist as NULL. Prior
// revisions of this helper called n.Scan(f) directly and hit exactly that
// bug; money fields across asset / depreciation / equipment mappers were
// landing in the DB as NULL without a visible error.
//
// We instead route through decimal.NewFromFloat which keeps full precision
// (including negative / fractional values and NaN / Inf which map to an
// invalid Numeric rather than a panic).
func NumericFromFloat64(f float64) pgtype.Numeric {
	// NaN and +/-Inf cannot be represented as SQL NUMERIC. Returning an
	// invalid Numeric is the correct DB-shape for "no value" and mirrors
	// how NULL would be stored; callers that want a specific fallback
	// should check upstream.
	if isUnrepresentable(f) {
		return pgtype.Numeric{Valid: false}
	}
	return NumericFromDecimal(decimal.NewFromFloat(f))
}

// NumericFromFloat64Ptr converts *float64 to pgtype.Numeric.
// Returns an invalid Numeric if the pointer is nil.
func NumericFromFloat64Ptr(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{Valid: false}
	}
	return NumericFromFloat64(*f)
}

// isUnrepresentable reports whether f is NaN or +/-Inf. Used to guard
// against turning those into a corrupt pgtype.Numeric.
func isUnrepresentable(f float64) bool {
	// NaN is never equal to itself; Inf compares strictly greater than
	// math.MaxFloat64. We only rely on IEEE-754 semantics here, no stdlib
	// import needed.
	return f != f || f > 1e308 || f < -1e308
}

// NumericFromInt64 converts int64 to pgtype.Numeric.
func NumericFromInt64(i int64) pgtype.Numeric {
	return pgtype.Numeric{
		Int:   big.NewInt(i),
		Valid: true,
	}
}

// Int64FromNumeric converts pgtype.Numeric to int64.
// Returns 0 if the Numeric is not valid.
func Int64FromNumeric(n pgtype.Numeric) int64 {
	if !n.Valid || n.Int == nil {
		return 0
	}
	return n.Int.Int64()
}

// ==================== Decimal Conversions ====================

// DecimalFromNumeric converts pgtype.Numeric to decimal.Decimal.
// Returns zero decimal if the Numeric is not valid.
func DecimalFromNumeric(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid || n.Int == nil {
		return decimal.Zero
	}
	return decimal.NewFromBigInt(n.Int, n.Exp)
}

// DecimalPtrFromNumeric converts pgtype.Numeric to *decimal.Decimal.
// Returns nil if the Numeric is not valid or NaN.
func DecimalPtrFromNumeric(n pgtype.Numeric) *decimal.Decimal {
	if !n.Valid || n.NaN || n.Int == nil {
		return nil
	}
	d := decimal.NewFromBigInt(n.Int, n.Exp)
	return &d
}

// NumericFromDecimal converts decimal.Decimal to pgtype.Numeric.
func NumericFromDecimal(d decimal.Decimal) pgtype.Numeric {
	coef := d.Coefficient()
	exp := d.Exponent()
	return pgtype.Numeric{
		Int:   coef,
		Exp:   exp,
		Valid: true,
	}
}

// NumericFromDecimalPtr converts *decimal.Decimal to pgtype.Numeric.
// Returns an invalid Numeric if the pointer is nil.
func NumericFromDecimalPtr(d *decimal.Decimal) pgtype.Numeric {
	if d == nil {
		return pgtype.Numeric{Valid: false}
	}
	coef := d.Coefficient()
	exp := d.Exponent()
	return pgtype.Numeric{
		Int:   coef,
		Exp:   exp,
		Valid: true,
	}
}

// ==================== Date Conversions ====================

// TimeFromDate converts pgtype.Date to time.Time.
// Returns zero time if the Date is not valid.
func TimeFromDate(d pgtype.Date) time.Time {
	if !d.Valid {
		return time.Time{}
	}
	return d.Time
}

// TimePtrFromDate converts pgtype.Date to *time.Time.
// Returns nil if the Date is not valid.
func TimePtrFromDate(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	return &d.Time
}

// DateFromTime converts time.Time to pgtype.Date.
func DateFromTime(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

// DateFromTimePtr converts *time.Time to pgtype.Date.
// Returns an invalid Date if the pointer is nil.
func DateFromTimePtr(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

// ==================== Timestamptz Conversions ====================

// TimeFromTimestamptz converts pgtype.Timestamptz to time.Time.
// Returns zero time if the Timestamptz is not valid.
func TimeFromTimestamptz(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}

// TimePtrFromTimestamptz converts pgtype.Timestamptz to *time.Time.
// Returns nil if the Timestamptz is not valid.
func TimePtrFromTimestamptz(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}

// TimestamptzFromTime converts time.Time to pgtype.Timestamptz.
func TimestamptzFromTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// TimestamptzFromTimePtr converts *time.Time to pgtype.Timestamptz.
// Returns an invalid Timestamptz if the pointer is nil.
func TimestamptzFromTimePtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// ==================== Timestamp (without timezone) Conversions ====================

// TimeFromTimestamp converts pgtype.Timestamp to time.Time.
// Returns zero time if the Timestamp is not valid.
func TimeFromTimestamp(ts pgtype.Timestamp) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}

// TimePtrFromTimestamp converts pgtype.Timestamp to *time.Time.
// Returns nil if the Timestamp is not valid.
func TimePtrFromTimestamp(ts pgtype.Timestamp) *time.Time {
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}

// TimestampFromTime converts time.Time to pgtype.Timestamp.
func TimestampFromTime(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t, Valid: true}
}

// TimestampFromTimePtr converts *time.Time to pgtype.Timestamp.
// Returns an invalid Timestamp if the pointer is nil.
func TimestampFromTimePtr(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: *t, Valid: true}
}

// ==================== Text Conversions ====================

// StringFromText converts pgtype.Text to string.
// Returns empty string if the Text is not valid.
func StringFromText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// StringPtrFromText converts pgtype.Text to *string.
// Returns nil if the Text is not valid.
func StringPtrFromText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// TextFromString converts string to pgtype.Text.
func TextFromString(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

// TextFromStringPtr converts *string to pgtype.Text.
// Returns an invalid Text if the pointer is nil.
func TextFromStringPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// ==================== Bool Conversions ====================

// BoolFromPgBool converts pgtype.Bool to bool.
// Returns false if the Bool is not valid.
func BoolFromPgBool(b pgtype.Bool) bool {
	if !b.Valid {
		return false
	}
	return b.Bool
}

// BoolPtrFromPgBool converts pgtype.Bool to *bool.
// Returns nil if the Bool is not valid.
func BoolPtrFromPgBool(b pgtype.Bool) *bool {
	if !b.Valid {
		return nil
	}
	return &b.Bool
}

// PgBoolFromBool converts bool to pgtype.Bool.
func PgBoolFromBool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}

// PgBoolFromBoolPtr converts *bool to pgtype.Bool.
// Returns an invalid Bool if the pointer is nil.
func PgBoolFromBoolPtr(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Valid: false}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

// ==================== Int4 (int32) Conversions ====================

// Int32FromInt4 converts pgtype.Int4 to int32.
// Returns 0 if the Int4 is not valid.
func Int32FromInt4(i pgtype.Int4) int32 {
	if !i.Valid {
		return 0
	}
	return i.Int32
}

// Int32PtrFromInt4 converts pgtype.Int4 to *int32.
// Returns nil if the Int4 is not valid.
func Int32PtrFromInt4(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}
	return &i.Int32
}

// Int4FromInt32 converts int32 to pgtype.Int4.
func Int4FromInt32(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}

// Int4FromInt32Ptr converts *int32 to pgtype.Int4.
// Returns an invalid Int4 if the pointer is nil.
func Int4FromInt32Ptr(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

// ==================== Int8 (int64) Conversions ====================

// Int64FromInt8 converts pgtype.Int8 to int64.
// Returns 0 if the Int8 is not valid.
func Int64FromInt8(i pgtype.Int8) int64 {
	if !i.Valid {
		return 0
	}
	return i.Int64
}

// Int64PtrFromInt8 converts pgtype.Int8 to *int64.
// Returns nil if the Int8 is not valid.
func Int64PtrFromInt8(i pgtype.Int8) *int64 {
	if !i.Valid {
		return nil
	}
	return &i.Int64
}

// Int8FromInt64 converts int64 to pgtype.Int8.
func Int8FromInt64(i int64) pgtype.Int8 {
	return pgtype.Int8{Int64: i, Valid: true}
}

// Int8FromInt64Ptr converts *int64 to pgtype.Int8.
// Returns an invalid Int8 if the pointer is nil.
func Int8FromInt64Ptr(i *int64) pgtype.Int8 {
	if i == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: *i, Valid: true}
}

// ==================== UUID Conversions ====================

// StringFromUUID converts pgtype.UUID to string.
// Returns empty string if the UUID is not valid.
func StringFromUUID(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	// Format UUID bytes as string
	b := u.Bytes
	return uuidBytesToString(b)
}

// UUIDFromString converts string to pgtype.UUID.
// Returns an invalid UUID if the string is not a valid UUID format.
func UUIDFromString(s string) pgtype.UUID {
	if s == "" {
		return pgtype.UUID{Valid: false}
	}
	var uuid pgtype.UUID
	if err := uuid.Scan(s); err != nil {
		return pgtype.UUID{Valid: false}
	}
	return uuid
}

// uuidBytesToString formats UUID bytes as a string.
func uuidBytesToString(b [16]byte) string {
	const hexDigits = "0123456789abcdef"
	buf := make([]byte, 36)
	hex := func(dst []byte, src byte) {
		dst[0] = hexDigits[src>>4]
		dst[1] = hexDigits[src&0x0f]
	}
	hex(buf[0:2], b[0])
	hex(buf[2:4], b[1])
	hex(buf[4:6], b[2])
	hex(buf[6:8], b[3])
	buf[8] = '-'
	hex(buf[9:11], b[4])
	hex(buf[11:13], b[5])
	buf[13] = '-'
	hex(buf[14:16], b[6])
	hex(buf[16:18], b[7])
	buf[18] = '-'
	hex(buf[19:21], b[8])
	hex(buf[21:23], b[9])
	buf[23] = '-'
	hex(buf[24:26], b[10])
	hex(buf[26:28], b[11])
	hex(buf[28:30], b[12])
	hex(buf[30:32], b[13])
	hex(buf[32:34], b[14])
	hex(buf[34:36], b[15])
	return string(buf)
}

// ==================== Interval Conversions ====================

// DurationFromInterval converts pgtype.Interval to time.Duration.
// Returns 0 if the Interval is not valid.
// Note: This only considers microseconds, not months/days which require calendar context.
func DurationFromInterval(i pgtype.Interval) time.Duration {
	if !i.Valid {
		return 0
	}
	return time.Duration(i.Microseconds) * time.Microsecond
}

// IntervalFromDuration converts time.Duration to pgtype.Interval.
func IntervalFromDuration(d time.Duration) pgtype.Interval {
	return pgtype.Interval{
		Microseconds: d.Microseconds(),
		Valid:        true,
	}
}

// ==================== Aliases for alternate naming conventions ====================

// TimeToPgTimestamptz is an alias for TimestamptzFromTime.
func TimeToPgTimestamptz(t time.Time) pgtype.Timestamptz {
	return TimestamptzFromTime(t)
}

// TimePtrToPgTimestamptz is an alias for TimestamptzFromTimePtr.
func TimePtrToPgTimestamptz(t *time.Time) pgtype.Timestamptz {
	return TimestamptzFromTimePtr(t)
}

// TimestamptzToPtr converts pgtype.Timestamptz to *time.Time. Alias for TimePtrFromTimestamptz.
func TimestamptzToPtr(ts pgtype.Timestamptz) *time.Time {
	return TimePtrFromTimestamptz(ts)
}

// PtrToTimestamptz is an alias for TimestamptzFromTimePtr.
func PtrToTimestamptz(t *time.Time) pgtype.Timestamptz {
	return TimestamptzFromTimePtr(t)
}

// TimeToTimestamptz is an alias for TimestamptzFromTime.
func TimeToTimestamptz(t time.Time) pgtype.Timestamptz {
	return TimestamptzFromTime(t)
}

// DateToTimePtr converts pgtype.Date to *time.Time. Alias for TimePtrFromDate.
func DateToTimePtr(d pgtype.Date) *time.Time {
	return TimePtrFromDate(d)
}

// TimestamptzToTime converts pgtype.Timestamptz to time.Time. Alias for TimeFromTimestamptz.
func TimestamptzToTime(ts pgtype.Timestamptz) time.Time {
	return TimeFromTimestamptz(ts)
}

// TimePtrToPgDate converts *time.Time to pgtype.Date. Alias for DateFromTimePtr.
func TimePtrToPgDate(t *time.Time) pgtype.Date {
	return DateFromTimePtr(t)
}

// NumericFromString converts a string to pgtype.Numeric.
func NumericFromString(s string) pgtype.Numeric {
	if s == "" {
		return pgtype.Numeric{Valid: false}
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return NumericFromDecimal(d)
}

// StringFromNumeric converts pgtype.Numeric to string.
func StringFromNumeric(n pgtype.Numeric) string {
	if !n.Valid {
		return ""
	}
	d := DecimalFromNumeric(n)
	return d.String()
}

