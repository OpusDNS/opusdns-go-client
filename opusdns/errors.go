package opusdns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// Standard sentinel errors for common error conditions.
var (
	// ErrNotFound is returned when a resource is not found.
	ErrNotFound = errors.New("opusdns: resource not found")

	// ErrUnauthorized is returned when the API key is invalid or missing.
	ErrUnauthorized = errors.New("opusdns: unauthorized - invalid or missing API key")

	// ErrForbidden is returned when access to a resource is forbidden.
	ErrForbidden = errors.New("opusdns: forbidden - insufficient permissions")

	// ErrRateLimited is matched by an HTTP 429 the client did not retry: retries
	// ran out, the API asked to wait longer than RetryWaitMax (see
	// APIError.RetryAfter), or the context ended while waiting to retry.
	ErrRateLimited = errors.New("opusdns: rate limited - too many requests")

	// ErrBadRequest is returned when the request is malformed.
	ErrBadRequest = errors.New("opusdns: bad request - invalid input")

	// ErrConflict is returned when there is a resource conflict.
	ErrConflict = errors.New("opusdns: conflict - resource already exists or state conflict")

	// ErrTimeout is matched by a request that ran out of time: its context's
	// deadline passed, or the HTTP client's timeout fired. For the former,
	// errors.Is(err, context.DeadlineExceeded) holds as well.
	ErrTimeout = errors.New("opusdns: request timeout")

	// ErrZoneNotFound is returned when a zone cannot be found for a given FQDN.
	//
	// Deprecated: no method returns it. A zone that does not exist is reported
	// as ErrNotFound.
	ErrZoneNotFound = errors.New("opusdns: no matching zone found for FQDN")

	// ErrInvalidInput is matched by input the client or the API rejected: a
	// ValidationError or ConfigError from the client, or an HTTP 422 from the
	// API, whose APIError.Errors says which fields failed.
	ErrInvalidInput = errors.New("opusdns: invalid input")

	// ErrServerError is returned when the server returns an internal error.
	ErrServerError = errors.New("opusdns: server error")
)

// APIError represents an error response from the OpusDNS API.
type APIError struct {
	// StatusCode is the HTTP status code.
	StatusCode int `json:"status_code"`

	// ErrorCode is the API-specific error code (e.g., "ERROR_USER_EXISTS"),
	// from the problem's "code" field.
	ErrorCode string `json:"error_code,omitempty"`

	// Message is the human-readable error message: the problem's detail, or,
	// for a validation error, the fields that failed.
	Message string `json:"message,omitempty"`

	// Title is the problem's short summary (e.g., "Authentication Error").
	Title string `json:"title,omitempty"`

	// Type identifies the kind of problem (e.g., "request-validation-failed").
	Type string `json:"type,omitempty"`

	// Errors lists the fields that failed validation (HTTP 422).
	Errors []FieldError `json:"errors,omitempty"`

	// Details contains additional error details from the API.
	Details map[string]interface{} `json:"details,omitempty"`

	// RequestID identifies the request to OpusDNS support, from the
	// X-Server-Request-ID header.
	RequestID string `json:"request_id,omitempty"`

	// RetryAfter is how long the API asked the caller to wait before trying
	// again, from the Retry-After header. Zero when the response had none.
	RetryAfter time.Duration `json:"retry_after,omitempty"`

	// RawBody contains the raw response body (not serialized to JSON).
	RawBody string `json:"-"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	var msg string

	if e.ErrorCode != "" && e.Message != "" {
		msg = fmt.Sprintf("opusdns: API error %d [%s]: %s", e.StatusCode, e.ErrorCode, e.Message)
	} else if e.Message != "" {
		msg = fmt.Sprintf("opusdns: API error %d: %s", e.StatusCode, e.Message)
	} else if e.ErrorCode != "" {
		msg = fmt.Sprintf("opusdns: API error %d [%s]", e.StatusCode, e.ErrorCode)
	} else {
		msg = fmt.Sprintf("opusdns: API error %d", e.StatusCode)
	}

	if e.RequestID != "" {
		msg += fmt.Sprintf(" (request_id: %s)", e.RequestID)
	}

	return msg
}

// Is implements errors.Is for APIError, allowing comparison with sentinel errors.
func (e *APIError) Is(target error) bool {
	switch target {
	case ErrNotFound:
		return e.StatusCode == http.StatusNotFound
	case ErrUnauthorized:
		return e.StatusCode == http.StatusUnauthorized
	case ErrForbidden:
		return e.StatusCode == http.StatusForbidden
	case ErrRateLimited:
		return e.StatusCode == http.StatusTooManyRequests
	case ErrBadRequest:
		return e.StatusCode == http.StatusBadRequest
	case ErrConflict:
		return e.StatusCode == http.StatusConflict
	case ErrInvalidInput:
		return e.StatusCode == http.StatusUnprocessableEntity
	case ErrServerError:
		return e.StatusCode >= 500
	}
	return false
}

// Unwrap returns the underlying standard error based on status code.
func (e *APIError) Unwrap() error {
	switch e.StatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusBadRequest:
		return ErrBadRequest
	case http.StatusConflict:
		return ErrConflict
	case http.StatusUnprocessableEntity:
		return ErrInvalidInput
	default:
		if e.StatusCode >= 500 {
			return ErrServerError
		}
		return nil
	}
}

// IsRetryable returns true if the status code marks a transient failure.
// It does not consider the request method: resending a POST or PATCH after a
// server error may repeat its effect.
func (e *APIError) IsRetryable() bool {
	return e.StatusCode == http.StatusTooManyRequests ||
		e.StatusCode >= http.StatusInternalServerError
}

// IsClientError returns true if the error is a client error (4xx).
func (e *APIError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// IsServerError returns true if the error is a server error (5xx).
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500
}

// FieldError is one field that failed validation (HTTP 422).
type FieldError struct {
	// Loc locates the field, e.g. ["body", "postal_code"].
	Loc []interface{} `json:"loc"`

	// Msg says what is wrong with it.
	Msg string `json:"msg"`

	// Type classifies the failure (e.g., "missing", "invalid_postal_code").
	Type string `json:"type"`
}

// Field returns Loc joined with dots, e.g. "body.postal_code".
func (e FieldError) Field() string {
	parts := make([]string, len(e.Loc))
	for i, part := range e.Loc {
		parts[i] = fmt.Sprint(part)
	}
	return strings.Join(parts, ".")
}

// NewAPIError creates an APIError from an HTTP response and body.
//
// The API answers with an RFC 9457 problem: "code", "title", "detail" and
// "type", plus "errors" for a validation error. Each field is read on its own,
// so one of an unexpected type does not lose the others.
func NewAPIError(resp *http.Response, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		RawBody:    string(body),
		RequestID:  firstNonEmpty(resp.Header.Get("X-Server-Request-ID"), resp.Header.Get("X-Request-ID")),
	}
	apiErr.RetryAfter, _ = parseRetryAfter(resp.Header.Get("Retry-After"))

	var fields map[string]json.RawMessage
	if len(body) == 0 || json.Unmarshal(body, &fields) != nil {
		return apiErr
	}
	str := func(key string) string {
		var s string
		_ = json.Unmarshal(fields[key], &s)
		return s
	}

	apiErr.ErrorCode = firstNonEmpty(str("code"), str("error_code"))
	apiErr.Title = str("title")
	apiErr.Type = str("type")
	_ = json.Unmarshal(fields["details"], &apiErr.Details)
	if json.Unmarshal(fields["errors"], &apiErr.Errors) != nil || len(apiErr.Errors) == 0 {
		// FastAPI's own validation error lists the fields under "detail".
		apiErr.Errors = nil
		if json.Unmarshal(fields["detail"], &apiErr.Errors) != nil {
			apiErr.Errors = nil
		}
	}
	apiErr.Message = firstNonEmpty(str("message"), str("detail"), str("error"), describeFieldErrors(apiErr.Errors), apiErr.Title)

	return apiErr
}

// describeFieldErrors joins failed fields into one message, e.g.
// "body.postal_code: '93053' is not a valid postal code for country 'NL'".
func describeFieldErrors(fieldErrors []FieldError) string {
	descriptions := make([]string, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		if field := fieldErr.Field(); field != "" {
			descriptions = append(descriptions, field+": "+fieldErr.Msg)
		} else {
			descriptions = append(descriptions, fieldErr.Msg)
		}
	}
	return strings.Join(descriptions, "; ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// RequestError represents an error that occurred while making a request.
type RequestError struct {
	// Op is the operation that was attempted: "marshal", "create", "execute",
	// "read", or "stream" (copying a body to Request.Output).
	Op string

	// URL is the URL that was requested.
	URL string

	// Err is the underlying error.
	Err error
}

// Error implements the error interface.
func (e *RequestError) Error() string {
	return fmt.Sprintf("opusdns: request %s failed for %s: %v", e.Op, e.URL, e.Err)
}

// Unwrap returns the underlying error.
func (e *RequestError) Unwrap() error {
	return e.Err
}

// Is reports a request that timed out as ErrTimeout.
func (e *RequestError) Is(target error) bool {
	return target == ErrTimeout && isTimeout(e.Err)
}

// isTimeout reports whether err is a context deadline or a network timeout,
// such as the HTTP client's.
func isTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

// ValidationError represents a validation error for input data.
type ValidationError struct {
	// Field is the name of the field that failed validation.
	Field string

	// Message describes the validation failure.
	Message string

	// Value is the invalid value (optional).
	Value interface{}
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("opusdns: validation error: %s: %s (got: %v)", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("opusdns: validation error: %s: %s", e.Field, e.Message)
}

// Is implements errors.Is for ValidationError.
func (e *ValidationError) Is(target error) bool {
	return target == ErrInvalidInput
}

// Unwrap returns ErrInvalidInput.
func (e *ValidationError) Unwrap() error {
	return ErrInvalidInput
}

// ConfigError represents a configuration validation error.
type ConfigError struct {
	// Field is the configuration field that failed validation.
	Field string

	// Message describes the configuration error.
	Message string
}

// Error implements the error interface.
func (e *ConfigError) Error() string {
	return fmt.Sprintf("opusdns: config error: %s: %s", e.Field, e.Message)
}

// Is implements errors.Is for ConfigError.
func (e *ConfigError) Is(target error) bool {
	return target == ErrInvalidInput
}

// Unwrap returns ErrInvalidInput.
func (e *ConfigError) Unwrap() error {
	return ErrInvalidInput
}

// Helper functions for error checking

// IsAPIError returns true if err is an APIError and extracts it.
func IsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// IsNotFoundError returns true if the error indicates a resource was not found.
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsUnauthorizedError returns true if the error indicates an authentication failure.
func IsUnauthorizedError(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbiddenError returns true if the error indicates a permission failure.
func IsForbiddenError(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsRateLimitError returns true if the error indicates rate limiting.
func IsRateLimitError(err error) bool {
	return errors.Is(err, ErrRateLimited)
}

// IsConflictError returns true if the error indicates a resource conflict.
func IsConflictError(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsRetryableError returns true if err is an APIError whose status code marks
// a transient failure. See APIError.IsRetryable for the caveat on POST and PATCH.
func IsRetryableError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.IsRetryable()
	}
	return false
}

// IsValidationError returns true if the error is a validation error.
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}
