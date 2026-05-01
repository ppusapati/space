// Package slice provides slice conversion and transformation utilities.
package slice

// Map transforms a slice of T to a slice of U using the given function.
func Map[T, U any](s []T, fn func(T) U) []U {
	if s == nil {
		return nil
	}
	result := make([]U, len(s))
	for i, v := range s {
		result[i] = fn(v)
	}
	return result
}

// MapErr transforms a slice of T to a slice of U, returning an error if any transformation fails.
func MapErr[T, U any](s []T, fn func(T) (U, error)) ([]U, error) {
	if s == nil {
		return nil, nil
	}
	result := make([]U, len(s))
	for i, v := range s {
		u, err := fn(v)
		if err != nil {
			return nil, err
		}
		result[i] = u
	}
	return result, nil
}

// MapIndex transforms a slice of T to a slice of U, passing both index and value to the function.
func MapIndex[T, U any](s []T, fn func(int, T) U) []U {
	if s == nil {
		return nil
	}
	result := make([]U, len(s))
	for i, v := range s {
		result[i] = fn(i, v)
	}
	return result
}

// Filter returns a new slice containing only elements for which fn returns true.
func Filter[T any](s []T, fn func(T) bool) []T {
	if s == nil {
		return nil
	}
	result := make([]T, 0, len(s))
	for _, v := range s {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// FilterMap combines Filter and Map: applies fn to each element, keeping only non-nil results.
func FilterMap[T, U any](s []T, fn func(T) *U) []U {
	if s == nil {
		return nil
	}
	result := make([]U, 0, len(s))
	for _, v := range s {
		if u := fn(v); u != nil {
			result = append(result, *u)
		}
	}
	return result
}

// Reduce reduces a slice to a single value using the given function.
func Reduce[T, U any](s []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, v := range s {
		result = fn(result, v)
	}
	return result
}

// Find returns the first element for which fn returns true, or nil if not found.
func Find[T any](s []T, fn func(T) bool) *T {
	for i := range s {
		if fn(s[i]) {
			return &s[i]
		}
	}
	return nil
}

// FindIndex returns the index of the first element for which fn returns true, or -1 if not found.
func FindIndex[T any](s []T, fn func(T) bool) int {
	for i, v := range s {
		if fn(v) {
			return i
		}
	}
	return -1
}

// Contains returns true if the slice contains the given value.
func Contains[T comparable](s []T, v T) bool {
	for _, e := range s {
		if e == v {
			return true
		}
	}
	return false
}

// Unique returns a new slice with duplicate elements removed.
// Preserves the order of first occurrence.
func Unique[T comparable](s []T) []T {
	if s == nil {
		return nil
	}
	seen := make(map[T]struct{}, len(s))
	result := make([]T, 0, len(s))
	for _, v := range s {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// Flatten flattens a slice of slices into a single slice.
func Flatten[T any](s [][]T) []T {
	if s == nil {
		return nil
	}
	var total int
	for _, inner := range s {
		total += len(inner)
	}
	result := make([]T, 0, total)
	for _, inner := range s {
		result = append(result, inner...)
	}
	return result
}

// Chunk splits a slice into chunks of the given size.
func Chunk[T any](s []T, size int) [][]T {
	if s == nil || size <= 0 {
		return nil
	}
	result := make([][]T, 0, (len(s)+size-1)/size)
	for i := 0; i < len(s); i += size {
		end := i + size
		if end > len(s) {
			end = len(s)
		}
		result = append(result, s[i:end])
	}
	return result
}

// Reverse returns a new slice with elements in reverse order.
func Reverse[T any](s []T) []T {
	if s == nil {
		return nil
	}
	result := make([]T, len(s))
	for i, v := range s {
		result[len(s)-1-i] = v
	}
	return result
}

// ==================== String Slice Helpers ====================

// StringsOrDefault returns the slice or an empty slice if nil.
func StringsOrDefault(s *[]string) []string {
	if s == nil {
		return []string{}
	}
	return *s
}

// StringsOrNil returns nil if the slice is empty, otherwise the slice.
func StringsOrNil(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	return s
}

// JoinStrings concatenates string slices.
func JoinStrings(slices ...[]string) []string {
	var total int
	for _, s := range slices {
		total += len(s)
	}
	result := make([]string, 0, total)
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

// ==================== Bytes Slice Helpers ====================

// BytesOrNil returns nil if the byte slice is nil or empty.
func BytesOrNil(b *[]byte) []byte {
	if b == nil {
		return nil
	}
	return *b
}

// BytesOrDefault returns an empty slice if nil.
func BytesOrDefault(b *[]byte) []byte {
	if b == nil {
		return []byte{}
	}
	return *b
}

// ==================== Conversion Helpers ====================

// ToMap converts a slice to a map using the key function.
func ToMap[T any, K comparable](s []T, keyFn func(T) K) map[K]T {
	if s == nil {
		return nil
	}
	result := make(map[K]T, len(s))
	for _, v := range s {
		result[keyFn(v)] = v
	}
	return result
}

// ToMapValues converts a slice to a map using key and value functions.
func ToMapValues[T any, K comparable, V any](s []T, keyFn func(T) K, valueFn func(T) V) map[K]V {
	if s == nil {
		return nil
	}
	result := make(map[K]V, len(s))
	for _, v := range s {
		result[keyFn(v)] = valueFn(v)
	}
	return result
}

// GroupBy groups slice elements by a key function.
func GroupBy[T any, K comparable](s []T, keyFn func(T) K) map[K][]T {
	if s == nil {
		return nil
	}
	result := make(map[K][]T)
	for _, v := range s {
		k := keyFn(v)
		result[k] = append(result[k], v)
	}
	return result
}

// Keys returns all keys from a map as a slice.
func Keys[K comparable, V any](m map[K]V) []K {
	if m == nil {
		return nil
	}
	result := make([]K, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// Values returns all values from a map as a slice.
func Values[K comparable, V any](m map[K]V) []V {
	if m == nil {
		return nil
	}
	result := make([]V, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}
