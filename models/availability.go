// Package models contains all the data types for the OpusDNS API.
package models

// DomainAvailabilityStatus represents the availability status of a domain.
type DomainAvailabilityStatus string

const (
	// AvailabilityStatusAvailable indicates the domain is available for registration.
	AvailabilityStatusAvailable DomainAvailabilityStatus = "available"

	// AvailabilityStatusUnavailable indicates the domain is not available.
	AvailabilityStatusUnavailable DomainAvailabilityStatus = "unavailable"

	// AvailabilityStatusMarketAvailable indicates the domain is available on the aftermarket.
	AvailabilityStatusMarketAvailable DomainAvailabilityStatus = "market_available"

	// AvailabilityStatusTMCHClaim indicates the domain has a trademark claim.
	AvailabilityStatusTMCHClaim DomainAvailabilityStatus = "tmch_claim"

	// AvailabilityStatusError indicates an error occurred during the check.
	AvailabilityStatusError DomainAvailabilityStatus = "error"
)

// IsAvailable returns true if the domain can be registered.
func (s DomainAvailabilityStatus) IsAvailable() bool {
	return s == AvailabilityStatusAvailable
}

// DomainAvailability represents the availability of a single domain.
type DomainAvailability struct {
	// Domain is the domain name that was checked.
	Domain string `json:"domain"`

	// Status is the availability status.
	Status DomainAvailabilityStatus `json:"status"`

	// Reason provides additional context for the status (optional).
	Reason *string `json:"reason,omitempty"`

	// Price contains pricing information if available.
	Price *DomainPrice `json:"price,omitempty"`

	// Premium indicates if this is a premium domain.
	Premium bool `json:"premium,omitempty"`

	// TMCHClaimKey is the TMCH claim key if applicable.
	TMCHClaimKey *string `json:"tmch_claim_key,omitempty"`
}

// DomainPrice represents pricing information for a domain.
type DomainPrice struct {
	// RegisterPrice is the registration price.
	RegisterPrice *string `json:"register_price,omitempty"`

	// RenewPrice is the renewal price.
	RenewPrice *string `json:"renew_price,omitempty"`

	// TransferPrice is the transfer price.
	TransferPrice *string `json:"transfer_price,omitempty"`

	// Currency is the currency code (e.g., "EUR", "USD").
	Currency Currency `json:"currency,omitempty"`

	// Period is the registration period in years.
	Period int `json:"period,omitempty"`
}

// AvailabilityCheckRequest represents a request to check domain availability.
type AvailabilityCheckRequest struct {
	// Domains is the list of domains to check.
	Domains []string `json:"domains"`
}

// AvailabilityResponse represents the response from a bulk availability check.
type AvailabilityResponse struct {
	// Results contains the availability results for each domain.
	Results []DomainAvailability `json:"results"`

	// Meta contains metadata about the request.
	Meta AvailabilityMeta `json:"meta"`
}

// AvailabilityMeta contains metadata about an availability check.
type AvailabilityMeta struct {
	// Total is the total number of domains checked.
	Total int `json:"total"`

	// ProcessingTimeMs is the time taken to process the request in milliseconds.
	ProcessingTimeMs int `json:"processing_time_ms"`
}

// DomainCheckResponse represents the response from a domain availability check.
type DomainCheckResponse struct {
	// Results contains the availability results.
	Results []DomainAvailabilityResult `json:"results"`
}

// DomainAvailabilityResult represents a single domain availability result.
type DomainAvailabilityResult struct {
	// Domain is the domain name.
	Domain string `json:"domain"`

	// Available indicates if the domain is available.
	Available bool `json:"available"`

	// Reason provides additional context (if not available).
	Reason *string `json:"reason,omitempty"`
}

// DomainSuggestion represents a suggested domain name.
type DomainSuggestion struct {
	// Domain is the suggested domain name.
	Domain string `json:"domain"`

	// Status is the availability status.
	Status DomainAvailabilityStatus `json:"status"`

	// Score is the relevance score (higher is more relevant).
	Score float64 `json:"score,omitempty"`

	// Price contains pricing information if available.
	Price *DomainPrice `json:"price,omitempty"`
}

// DomainSuggestRequest represents a request for domain suggestions.
type DomainSuggestRequest struct {
	// Query is the search query or seed domain.
	Query string `json:"query"`

	// TLDs is an optional list of TLDs to include.
	TLDs []string `json:"tlds,omitempty"`

	// Limit is the maximum number of suggestions to return.
	Limit int `json:"limit,omitempty"`

	// IncludeUnavailable includes unavailable suggestions.
	IncludeUnavailable bool `json:"include_unavailable,omitempty"`
}

// DomainSuggestResponse represents the response from a domain suggestion request.
type DomainSuggestResponse struct {
	// Suggestions contains the suggested domains.
	Suggestions []DomainSuggestion `json:"suggestions"`

	// Meta contains metadata about the request.
	Meta AvailabilityMeta `json:"meta,omitempty"`
}
