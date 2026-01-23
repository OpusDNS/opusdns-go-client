// Package client provides a Go client library for the OpusDNS API.
package opusdns

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Standard sentinel errors for common error conditions.
var (
	// ErrNotFound is returned when a resource is not found.
	ErrNotFound = errors.New("opusdns: resource not found")

	// ErrUnauthorized is returned when the API key is invalid or missing.
	ErrUnauthorized = errors.New("opusdns: unauthorized - invalid or missing API key")

	// ErrForbidden is returned when access to a resource is forbidden.
	ErrForbidden = errors.New("opusdns: forbidden - insufficient permissions")

	// ErrRateLimited is returned when rate limited (after retries exhausted).
	ErrRateLimited = errors.New("opusdns: rate limited - too many requests")

	// ErrBadRequest is returned when the request is malformed.
	ErrBadRequest = errors.New("opusdns: bad request - invalid input")

	// ErrConflict is returned when there is a resource conflict.
	ErrConflict = errors.New("opusdns: conflict - resource already exists or state conflict")

	// ErrTimeout is returned when a request times out.
	ErrTimeout = errors.New("opusdns: request timeout")

	// ErrZoneNotFound is returned when a zone cannot be found for a given FQDN.
	ErrZoneNotFound = errors.New("opusdns: no matching zone found for FQDN")

	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("opusdns: invalid input")

	// ErrServerError is returned when the server returns an internal error.
	ErrServerError = errors.New("opusdns: server error")
)

// APIError represents an error response from the OpusDNS API.
type APIError struct {
	// StatusCode is the HTTP status code.
	StatusCode int `json:"status_code"`

	// ErrorCode is the API-specific error code (e.g., "zone_not_found").
	ErrorCode string `json:"error_code,omitempty"`

	// Message is the human-readable error message.
	Message string `json:"message,omitempty"`

	// Details contains additional error details from the API.
	Details map[string]interface{} `json:"details,omitempty"`

	// RequestID is the unique identifier for the request (from X-Request-ID header).
	RequestID string `json:"request_id,omitempty"`

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
	default:
		if e.StatusCode >= 500 {
			return ErrServerError
		}
		return nil
	}
}

// IsRetryable returns true if the error is retryable.
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

// NewAPIError creates an APIError from an HTTP response and body.
func NewAPIError(resp *http.Response, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		RawBody:    string(body),
	}

	// Try to get request ID from headers
	if reqID := resp.Header.Get("X-Request-ID"); reqID != "" {
		apiErr.RequestID = reqID
	}
	if reqID := resp.Header.Get("X-Request-Id"); reqID != "" && apiErr.RequestID == "" {
		apiErr.RequestID = reqID
	}

	// Try to parse error details from body
	if len(body) > 0 {
		var parsed struct {
			ErrorCode string                 `json:"error_code"`
			Message   string                 `json:"message"`
			Error     string                 `json:"error"`
			Detail    string                 `json:"detail"`
			Details   map[string]interface{} `json:"details"`
		}
		if err := json.Unmarshal(body, &parsed); err == nil {
			apiErr.ErrorCode = parsed.ErrorCode
			apiErr.Details = parsed.Details

			// Try different message fields
			if parsed.Message != "" {
				apiErr.Message = parsed.Message
			} else if parsed.Error != "" {
				apiErr.Message = parsed.Error
			} else if parsed.Detail != "" {
				apiErr.Message = parsed.Detail
			}
		}
	}

	return apiErr
}

// RequestError represents an error that occurred while making a request.
type RequestError struct {
	// Op is the operation that was attempted (e.g., "marshal", "create", "execute", "read").
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

// IsRetryableError returns true if the error is retryable.
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
