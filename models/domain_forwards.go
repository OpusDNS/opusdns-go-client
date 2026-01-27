// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// DomainForwardID is a TypeID for domain forwards.
type DomainForwardID = TypeID

// HttpProtocol represents the HTTP protocol type.
type HttpProtocol string

const (
	// HttpProtocolHTTP represents HTTP protocol.
	HttpProtocolHTTP HttpProtocol = "http"

	// HttpProtocolHTTPS represents HTTPS protocol.
	HttpProtocolHTTPS HttpProtocol = "https"
)

// RedirectCode represents the HTTP redirect status code.
type RedirectCode int

const (
	// RedirectCodePermanent performs a 301 permanent redirect.
	RedirectCodePermanent RedirectCode = 301

	// RedirectCodeTemporary performs a 302 temporary redirect.
	RedirectCodeTemporary RedirectCode = 302

	// RedirectCodeTemporaryRedirect performs a 307 Temporary Redirect.
	RedirectCodeTemporaryRedirect RedirectCode = 307

	// RedirectCodePermanentRedirect performs a 308 Permanent Redirect.
	RedirectCodePermanentRedirect RedirectCode = 308
)

// DomainForwardSortField represents fields that can be used for sorting domain forwards.
type DomainForwardSortField string

const (
	DomainForwardSortByHostname DomainForwardSortField = "hostname"
)

// DomainForwardZoneSortField represents fields that can be used for sorting domain forward zones.
type DomainForwardZoneSortField string

const (
	DomainForwardZoneSortByHostname DomainForwardZoneSortField = "hostname"
)

// DomainForward represents a domain forwarding configuration.
type DomainForward struct {
	// Hostname is the source hostname (e.g., "www.example.com" or "example.com").
	Hostname string `json:"hostname"`

	// Enabled indicates whether the domain forward is active.
	Enabled bool `json:"enabled"`

	// HTTP contains the HTTP protocol forwarding configuration.
	HTTP *DomainForwardProtocolSet `json:"http,omitempty"`

	// HTTPS contains the HTTPS protocol forwarding configuration.
	HTTPS *DomainForwardProtocolSet `json:"https,omitempty"`

	// CreatedOn is when the domain forward was created.
	CreatedOn time.Time `json:"created_on"`

	// UpdatedOn is when the domain forward was last updated.
	UpdatedOn time.Time `json:"updated_on"`
}

// DomainForwardProtocolSet represents the forwarding configuration for a specific protocol.
type DomainForwardProtocolSet struct {
	// Redirects contains the list of redirect configurations.
	Redirects []HttpRedirect `json:"redirects"`

	// CreatedOn is when the protocol set was created.
	CreatedOn time.Time `json:"created_on"`

	// UpdatedOn is when the protocol set was last updated.
	UpdatedOn time.Time `json:"updated_on"`
}

// HttpRedirect represents a single HTTP redirect configuration.
type HttpRedirect struct {
	// RequestProtocol is the source protocol (http or https).
	RequestProtocol HttpProtocol `json:"request_protocol"`

	// RequestHostname is the source hostname.
	RequestHostname string `json:"request_hostname"`

	// RequestPath is the source path to match.
	RequestPath string `json:"request_path"`

	// RequestSubdomain is the optional subdomain to match.
	RequestSubdomain *string `json:"request_subdomain,omitempty"`

	// TargetProtocol is the destination protocol.
	TargetProtocol HttpProtocol `json:"target_protocol"`

	// TargetHostname is the destination hostname.
	TargetHostname string `json:"target_hostname"`

	// TargetPath is the destination path.
	TargetPath string `json:"target_path"`

	// RedirectCode is the HTTP redirect status code.
	RedirectCode RedirectCode `json:"redirect_code"`
}

// DomainForwardListResponse represents the paginated response when listing domain forwards.
type DomainForwardListResponse struct {
	// Results contains the list of domain forwards for the current page.
	Results []DomainForward `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// DomainForwardZone represents domain forwards associated with a DNS zone.
type DomainForwardZone struct {
	// Hostname is the zone hostname.
	Hostname string `json:"hostname"`

	// Forwards contains the domain forwards for this zone.
	Forwards []DomainForward `json:"forwards,omitempty"`

	// CreatedOn is when the zone was created.
	CreatedOn time.Time `json:"created_on"`

	// UpdatedOn is when the zone was last updated.
	UpdatedOn time.Time `json:"updated_on"`
}

// DomainForwardZoneListResponse represents the paginated response when listing domain forward zones.
type DomainForwardZoneListResponse struct {
	// Results contains the list of domain forward zones for the current page.
	Results []DomainForwardZone `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// DomainForwardCreateRequest represents a request to create a domain forward.
type DomainForwardCreateRequest struct {
	// Hostname is the source hostname to forward from.
	Hostname string `json:"hostname"`

	// Enabled indicates whether the domain forward should be active.
	Enabled bool `json:"enabled,omitempty"`

	// HTTP contains the HTTP protocol forwarding configuration.
	HTTP *DomainForwardProtocolSetRequest `json:"http,omitempty"`

	// HTTPS contains the HTTPS protocol forwarding configuration.
	HTTPS *DomainForwardProtocolSetRequest `json:"https,omitempty"`
}

// DomainForwardProtocolSetRequest represents the request for a protocol set configuration.
type DomainForwardProtocolSetRequest struct {
	// Redirects contains the list of redirect configurations.
	Redirects []HttpRedirectRequest `json:"redirects"`
}

// HttpRedirectRequest represents a request to create/update an HTTP redirect.
type HttpRedirectRequest struct {
	// RequestPath is the source path to match.
	RequestPath string `json:"request_path"`

	// TargetProtocol is the destination protocol.
	TargetProtocol HttpProtocol `json:"target_protocol"`

	// TargetHostname is the destination hostname.
	TargetHostname string `json:"target_hostname"`

	// TargetPath is the destination path.
	TargetPath string `json:"target_path"`

	// RedirectCode is the HTTP redirect status code.
	RedirectCode RedirectCode `json:"redirect_code"`
}

// WildcardHttpRedirectRequest represents a request for a wildcard HTTP redirect.
type WildcardHttpRedirectRequest struct {
	// RequestPath is the source path pattern to match (supports wildcards).
	RequestPath string `json:"request_path"`

	// RequestSubdomain is the subdomain pattern to match.
	RequestSubdomain string `json:"request_subdomain"`

	// TargetProtocol is the destination protocol.
	TargetProtocol HttpProtocol `json:"target_protocol"`

	// TargetHostname is the destination hostname.
	TargetHostname string `json:"target_hostname"`

	// TargetPath is the destination path.
	TargetPath string `json:"target_path"`

	// RedirectCode is the HTTP redirect status code.
	RedirectCode RedirectCode `json:"redirect_code"`
}

// DomainForwardSetCreateRequest represents a request to create a protocol-specific forward set.
type DomainForwardSetCreateRequest struct {
	// Protocol is the protocol for this forward set.
	Protocol HttpProtocol `json:"protocol"`

	// Redirects contains the list of redirect configurations.
	Redirects []HttpRedirectRequest `json:"redirects"`
}

// DomainForwardSetRequest represents a request to update a protocol-specific forward set.
type DomainForwardSetRequest struct {
	// Redirects contains the list of redirect configurations.
	Redirects []HttpRedirectRequest `json:"redirects"`
}

// PatchOp represents the operation type for patch requests.
type PatchOp string

const (
	// PatchOpUpsert creates or updates a resource.
	PatchOpUpsert PatchOp = "upsert"

	// PatchOpRemove deletes a resource.
	PatchOpRemove PatchOp = "remove"
)

// DomainForwardPatchOp represents a single patch operation for domain forwards.
type DomainForwardPatchOp struct {
	// Op is the operation type.
	Op PatchOp `json:"op"`

	// Redirect is the redirect configuration for the operation.
	Redirect interface{} `json:"redirect"`
}

// DomainForwardPatchOps represents a batch of patch operations for domain forwards.
type DomainForwardPatchOps struct {
	// Ops is the list of patch operations.
	Ops []DomainForwardPatchOp `json:"ops"`
}

// HttpRedirectRemove represents a request to remove an HTTP redirect.
type HttpRedirectRemove struct {
	// RequestProtocol is the source protocol.
	RequestProtocol HttpProtocol `json:"request_protocol"`

	// RequestHostname is the source hostname.
	RequestHostname string `json:"request_hostname"`

	// RequestPath is the source path.
	RequestPath string `json:"request_path"`

	// RequestSubdomain is the optional subdomain.
	RequestSubdomain *string `json:"request_subdomain,omitempty"`
}

// ListDomainForwardsOptions contains options for listing domain forwards.
type ListDomainForwardsOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of items per page.
	PageSize int

	// SortBy is the field to sort by.
	SortBy DomainForwardSortField

	// SortOrder is the sort direction.
	SortOrder SortOrder

	// Search is an optional search query to filter by hostname.
	Search string

	// Enabled filters by enabled status.
	Enabled *bool
}

// DomainForwardMetrics represents metrics for domain forwards.
type DomainForwardMetrics struct {
	// InvokedForwards is the number of forwards that have been invoked.
	InvokedForwards int `json:"invoked_forwards"`

	// ConfiguredForwards is the number of configured forwards.
	ConfiguredForwards int `json:"configured_forwards"`

	// TotalVisits is the total number of visits.
	TotalVisits int `json:"total_visits"`

	// UniqueVisits is the number of unique visits.
	UniqueVisits int `json:"unique_visits"`
}

// TimeSeriesBucket represents a time series data point.
type TimeSeriesBucket struct {
	// Timestamp is the timestamp for this bucket.
	Timestamp time.Time `json:"timestamp"`

	// Value is the value for this bucket.
	Value int `json:"value"`
}

// DomainForwardTimeSeriesResponse represents time series metrics for domain forwards.
type DomainForwardTimeSeriesResponse struct {
	// Results contains the time series data points.
	Results []TimeSeriesBucket `json:"results"`
}

// GeoStatsBucket represents geographic statistics.
type GeoStatsBucket struct {
	// Country is the country code.
	Country string `json:"country"`

	// Count is the number of visits from this country.
	Count int `json:"count"`
}

// DomainForwardGeoStatsResponse represents geographic statistics for domain forwards.
type DomainForwardGeoStatsResponse struct {
	// Results contains the geographic statistics.
	Results []GeoStatsBucket `json:"results"`
}

// BrowserStatsBucket represents browser statistics.
type BrowserStatsBucket struct {
	// Browser is the browser name.
	Browser string `json:"browser"`

	// Count is the number of visits from this browser.
	Count int `json:"count"`
}

// DomainForwardBrowserStatsResponse represents browser statistics for domain forwards.
type DomainForwardBrowserStatsResponse struct {
	// Results contains the browser statistics.
	Results []BrowserStatsBucket `json:"results"`
}

// PlatformStatsBucket represents platform/OS statistics.
type PlatformStatsBucket struct {
	// Platform is the platform/OS name.
	Platform string `json:"platform"`

	// Count is the number of visits from this platform.
	Count int `json:"count"`
}

// DomainForwardPlatformStatsResponse represents platform statistics for domain forwards.
type DomainForwardPlatformStatsResponse struct {
	// Results contains the platform statistics.
	Results []PlatformStatsBucket `json:"results"`
}

// ReferrerStatsBucket represents referrer statistics.
type ReferrerStatsBucket struct {
	// Referrer is the referrer URL or domain.
	Referrer string `json:"referrer"`

	// Count is the number of visits from this referrer.
	Count int `json:"count"`
}

// DomainForwardReferrerStatsResponse represents referrer statistics for domain forwards.
type DomainForwardReferrerStatsResponse struct {
	// Results contains the referrer statistics.
	Results []ReferrerStatsBucket `json:"results"`
}

// StatusCodeStatsBucket represents HTTP status code statistics.
type StatusCodeStatsBucket struct {
	// StatusCode is the HTTP status code.
	StatusCode int `json:"status_code"`

	// Count is the number of responses with this status code.
	Count int `json:"count"`
}

// DomainForwardStatusCodeStatsResponse represents status code statistics for domain forwards.
type DomainForwardStatusCodeStatsResponse struct {
	// Results contains the status code statistics.
	Results []StatusCodeStatsBucket `json:"results"`
}

// UserAgentStatsBucket represents user agent statistics.
type UserAgentStatsBucket struct {
	// UserAgent is the user agent string.
	UserAgent string `json:"user_agent"`

	// Count is the number of visits from this user agent.
	Count int `json:"count"`
}

// DomainForwardUserAgentStatsResponse represents user agent statistics for domain forwards.
type DomainForwardUserAgentStatsResponse struct {
	// Results contains the user agent statistics.
	Results []UserAgentStatsBucket `json:"results"`
}

// VisitsByKeyBucket represents visits grouped by a key.
type VisitsByKeyBucket struct {
	// Key is the grouping key.
	Key string `json:"key"`

	// Count is the number of visits for this key.
	Count int `json:"count"`
}

// DomainForwardVisitsByKeyResponse represents visits grouped by key for domain forwards.
type DomainForwardVisitsByKeyResponse struct {
	// Results contains the visits by key data.
	Results []VisitsByKeyBucket `json:"results"`
}
