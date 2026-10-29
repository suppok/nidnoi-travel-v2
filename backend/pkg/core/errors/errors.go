package cerrors

import (
	"errors"
	"fmt"
	"net/http"

	ozzo "github.com/go-ozzo/ozzo-validation/v4"

	"backend/pkg/constants/errorcode"
)

// AppError is the only error type use cases return. ApiResponseHandler maps it to
// { "success": false, "error": { code, message, fields? } } with Status.
type AppError struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string // only for VALIDATION
	Err     error             // internal cause, logged, never sent to the client
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Err }

// New builds an AppError from a business code.
func New(status int, code errorcode.Code) *AppError {
	return &AppError{Status: status, Code: code.Code, Message: code.Message}
}

// Validation converts ozzo validation errors into a VALIDATION error with per-field messages.
// Any other error becomes INTERNAL.
func Validation(err error) *AppError {
	var verrs ozzo.Errors
	if !errors.As(err, &verrs) {
		return Wrap(err, "validation")
	}
	fields := make(map[string]string, len(verrs))
	for name, ferr := range verrs {
		fields[name] = ferr.Error()
	}
	e := New(http.StatusBadRequest, errorcode.Validation)
	e.Fields = fields
	return e
}

// FieldError is a VALIDATION error for a single field, for rules checked outside Validate.
func FieldError(field, message string) *AppError {
	e := New(http.StatusBadRequest, errorcode.Validation)
	e.Fields = map[string]string{field: message}
	return e
}

func Unauthenticated() *AppError {
	return New(http.StatusUnauthorized, errorcode.Unauthenticated)
}

func PermissionDenied() *AppError {
	return New(http.StatusForbidden, errorcode.PermissionDenied)
}

func NotFound() *AppError {
	return New(http.StatusNotFound, errorcode.NotFound)
}

func Conflict(code errorcode.Code) *AppError {
	return New(http.StatusConflict, code)
}

func RateLimited() *AppError {
	return New(http.StatusTooManyRequests, errorcode.RateLimited)
}

// Wrap turns an unexpected error into INTERNAL, keeping the cause for logs.
// An error that already is an *AppError is returned unchanged.
func Wrap(err error, context string) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	e := New(http.StatusInternalServerError, errorcode.Internal)
	e.Err = fmt.Errorf("%s: %w", context, err)
	return e
}
