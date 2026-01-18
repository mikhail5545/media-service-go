package errors

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidArgument  = errors.New("invalid argument")    // ErrInvalidArgument invalid argument passed, invalid argument format error.
	ErrValidationFailed = errors.New("validation failed")   // ErrValidationFailed argument is well-formatted, but conflicts with the validation rules error.
	ErrNotFound         = errors.New("not found")           // ErrNotFound resource not found error.
	ErrConflict         = errors.New("state conflict")      // ErrConflict resource state conflict error.
	ErrAlreadyExists    = errors.New("already exists")      // ErrAlreadyExists resource already exists error.
	ErrPermissionDenied = errors.New("permission denied")   // ErrPermissionDenied caller is not allowed to use this error.
	ErrTooManyRequests  = errors.New("too many requests")   // ErrTooManyRequests request is rate limited error.
	ErrUnimplemented    = errors.New("unimplemented")       // ErrUnimplemented functionality is not implemented error.
	ErrCanceled         = errors.New("context canceled")    // ErrCanceled request context cancelled error.
	ErrUnavailable      = errors.New("service unavailable") // ErrUnavailable external service error.
)

var ErrorAliases = map[error]string{
	ErrInvalidArgument:  "INVALID_ARGUMENT",
	ErrValidationFailed: "VALIDATION_FAILED",
	ErrNotFound:         "NOT_FOUND",
	ErrConflict:         "CONFLICT",
	ErrAlreadyExists:    "ALREADY_EXISTS",
	ErrPermissionDenied: "PERMISSION_DENIED",
	ErrTooManyRequests:  "TOO_MANY_REQUESTS",
	ErrUnimplemented:    "UNIMPLEMENTED",
	ErrCanceled:         "CANCELED",
	ErrUnavailable:      "UNAVAILABLE",
}

func NewInvalidArgumentError(v any) error {
	return fmt.Errorf("%w: %v", ErrInvalidArgument, v)
}

func NewValidationFailedError(v any) error {
	return fmt.Errorf("%w: %v", ErrValidationFailed, v)
}

func NewNotFoundError(v any) error {
	return fmt.Errorf("%w: %v", ErrNotFound, v)
}

func NewConflictError(v any) error {
	return fmt.Errorf("%w: %v", ErrConflict, v)
}

func newConflictError(v any) error {
	return fmt.Errorf("%w: %v", ErrConflict, v)
}

func NewAlreadyExistsError(v any) error {
	return fmt.Errorf("%w: %v", ErrAlreadyExists, v)
}

func NewPermissionDeniedError(v any) error {
	return fmt.Errorf("%w: %v", ErrPermissionDenied, v)
}

func NewTooManyRequestsError(v any) error {
	return fmt.Errorf("%w: %v", ErrTooManyRequests, v)
}

func NewUnimplementedError(v any) error {
	return fmt.Errorf("%w: %v", ErrUnimplemented, v)
}

func NewCanceledError(v any) error {
	return fmt.Errorf("%w: %v", ErrCanceled, v)
}

func NewUnavailableError(v any) error {
	return fmt.Errorf("%w: %v", ErrUnavailable, v)
}
