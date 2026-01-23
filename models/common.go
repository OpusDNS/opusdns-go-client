// Package models contains all the data types for the OpusDNS API.
package models

import (
	"encoding/json"
	"time"
)

// SortOrder represents the sort direction.
type SortOrder string

const (
	// SortAsc sorts in ascending order.
	SortAsc SortOrder = "asc"

	// SortDesc sorts in descending order.
	SortDesc SortOrder = "desc"
)

// Pagination represents pagination metadata in API responses.
type Pagination struct {
	// TotalPages is the total number of pages available.
	TotalPages int `json:"total_pages"`

	// CurrentPage is the current page number (1-indexed).
	CurrentPage int `json:"current_page"`

	// HasNextPage indicates whether there are more pages available.
	HasNextPage bool `json:"has_next_page"`

	// HasPreviousPage indicates whether there are previous pages available.
	HasPreviousPage bool `json:"has_previous_page"`

	// TotalCount is the total number of items across all pages.
	TotalCount int `json:"total_count,omitempty"`

	// PageSize is the number of items per page.
	PageSize int `json:"page_size,omitempty"`
}

// ListOptions contains common options for listing resources.
type ListOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of items per page.
	PageSize int

	// SortBy is the field to sort by.
	SortBy string

	// SortOrder is the sort direction (asc or desc).
	SortOrder SortOrder

	// Search is an optional search query to filter results.
	Search string
}

// TypeID is a type-safe identifier following the TypeID specification.
// Format: prefix_01h45ytscbebyvny4gc8cr8ma2
type TypeID string

// String returns the string representation of the TypeID.
func (t TypeID) String() string {
	return string(t)
}

// IsEmpty returns true if the TypeID is empty.
func (t TypeID) IsEmpty() bool {
	return t == ""
}

// Currency represents a currency code.
type Currency string

const (
	CurrencyEUR Currency = "EUR"
	CurrencyUSD Currency = "USD"
	CurrencyGBP Currency = "GBP"
	CurrencyCHF Currency = "CHF"
)

// Timestamp is a custom time type for JSON parsing.
type Timestamp struct {
	time.Time
}

// MarshalJSON implements the json.Marshaler interface.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Format(time.RFC3339))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	// Handle null values
	if string(data) == "null" {
		t.Time = time.Time{}
		return nil
	}

	// Parse the time string
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// Try multiple formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		parsed, err := time.Parse(format, s)
		if err == nil {
			t.Time = parsed
			return nil
		}
	}

	// Fall back to time.Parse with RFC3339
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// DomainNameParts represents the parts of a domain name.
type DomainNameParts struct {
	// SLD is the second-level domain (e.g., "example" in "example.com").
	SLD string `json:"sld"`

	// TLD is the top-level domain (e.g., "com" in "example.com").
	TLD string `json:"tld"`

	// Subdomain is the subdomain portion (if any).
	Subdomain string `json:"subdomain,omitempty"`
}

// PaginatedResponse is a generic wrapper for paginated API responses.
type PaginatedResponse[T any] struct {
	// Results contains the items for the current page.
	Results []T `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// HasMore returns true if there are more pages available.
func (p *PaginatedResponse[T]) HasMore() bool {
	return p.Pagination.HasNextPage
}

// NextPage returns the next page number, or 0 if there are no more pages.
func (p *PaginatedResponse[T]) NextPage() int {
	if p.Pagination.HasNextPage {
		return p.Pagination.CurrentPage + 1
	}
	return 0
}

// Meta contains metadata about an API response.
type Meta struct {
	// ProcessingTimeMs is the time taken to process the request in milliseconds.
	ProcessingTimeMs int `json:"processing_time_ms,omitempty"`

	// Total is the total number of items.
	Total int `json:"total,omitempty"`
}

// StringPtr returns a pointer to the given string.
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the given int.
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to the given bool.
func BoolPtr(b bool) *bool {
	return &b
}

// TimePtr returns a pointer to the given time.Time.
func TimePtr(t time.Time) *time.Time {
	return &t
}

// Deref safely dereferences a pointer, returning the zero value if nil.
func Deref[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}
