// Package validation wraps protovalidate-go so handlers don't need to
// instantiate the validator at every call site. The default global
// validator is shared and thread-safe.
package validation

import (
	"errors"
	"fmt"
	"sync"

	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	once     sync.Once
	shared   protovalidate.Validator
	sharedEr error
)

// ValidationError is returned by [Validate] when one or more constraints fail.
type ValidationError struct {
	Cause error
}

func (e *ValidationError) Error() string { return fmt.Sprintf("validation: %v", e.Cause) }
func (e *ValidationError) Unwrap() error { return e.Cause }

// Default returns a process-wide validator, lazily constructed.
func Default() (protovalidate.Validator, error) {
	once.Do(func() {
		v, err := protovalidate.New()
		shared, sharedEr = v, err
	})
	return shared, sharedEr
}

// Validate runs the default validator against `m`. It returns nil if
// the message satisfies its protovalidate constraints; otherwise the
// returned error wraps a [*ValidationError].
//
// `m` must be a generated protobuf message implementing
// `protoreflect.ProtoMessage`.
func Validate(m protoreflect.ProtoMessage) error {
	v, err := Default()
	if err != nil {
		return fmt.Errorf("validation: build validator: %w", err)
	}
	if err := v.Validate(m); err != nil {
		return &ValidationError{Cause: err}
	}
	return nil
}

// Is reports whether err is a *ValidationError.
func Is(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}
