// Package errs maps domain errors to ConnectRPC error codes so that
// every layer above the repository can return idiomatic Go errors and
// the handler layer translates them to wire status with one helper.
package errs

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"
)

// Domain is a sentinel-error category. Each value maps deterministically
// to a Connect error code via [Code].
type Domain int

const (
	// DomainUnknown maps to CodeInternal.
	DomainUnknown Domain = iota
	// DomainNotFound maps to CodeNotFound.
	DomainNotFound
	// DomainAlreadyExists maps to CodeAlreadyExists.
	DomainAlreadyExists
	// DomainInvalidArgument maps to CodeInvalidArgument.
	DomainInvalidArgument
	// DomainPreconditionFailed maps to CodeFailedPrecondition.
	DomainPreconditionFailed
	// DomainPermissionDenied maps to CodePermissionDenied.
	DomainPermissionDenied
	// DomainUnauthenticated maps to CodeUnauthenticated.
	DomainUnauthenticated
	// DomainResourceExhausted maps to CodeResourceExhausted.
	DomainResourceExhausted
	// DomainUnavailable maps to CodeUnavailable.
	DomainUnavailable
	// DomainCanceled maps to CodeCanceled.
	DomainCanceled
	// DomainDeadlineExceeded maps to CodeDeadlineExceeded.
	DomainDeadlineExceeded
)

// E is the canonical domain error. Lower layers wrap upstream errors
// with [Wrap] / [New]; the handler layer calls [ToConnect] to convert.
type E struct {
	Domain  Domain
	Message string
	Cause   error
}

// Error implements error.
func (e *E) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the inner error so errors.Is / errors.As work.
func (e *E) Unwrap() error { return e.Cause }

// New constructs an error of the given domain.
func New(d Domain, format string, args ...any) *E {
	return &E{Domain: d, Message: fmt.Sprintf(format, args...)}
}

// Wrap wraps an existing error in a domain. Returns nil if `cause` is nil.
func Wrap(d Domain, cause error, format string, args ...any) *E {
	if cause == nil {
		return nil
	}
	return &E{Domain: d, Message: fmt.Sprintf(format, args...), Cause: cause}
}

// Code returns the Connect code corresponding to the domain.
func Code(d Domain) connect.Code {
	switch d {
	case DomainNotFound:
		return connect.CodeNotFound
	case DomainAlreadyExists:
		return connect.CodeAlreadyExists
	case DomainInvalidArgument:
		return connect.CodeInvalidArgument
	case DomainPreconditionFailed:
		return connect.CodeFailedPrecondition
	case DomainPermissionDenied:
		return connect.CodePermissionDenied
	case DomainUnauthenticated:
		return connect.CodeUnauthenticated
	case DomainResourceExhausted:
		return connect.CodeResourceExhausted
	case DomainUnavailable:
		return connect.CodeUnavailable
	case DomainCanceled:
		return connect.CodeCanceled
	case DomainDeadlineExceeded:
		return connect.CodeDeadlineExceeded
	default:
		return connect.CodeInternal
	}
}

// ToConnect maps any error to a *connect.Error preserving the domain
// when the underlying error is an [E].
func ToConnect(err error) *connect.Error {
	if err == nil {
		return nil
	}
	var e *E
	if errors.As(err, &e) {
		return connect.NewError(Code(e.Domain), errors.New(e.Error()))
	}
	return connect.NewError(connect.CodeInternal, err)
}
