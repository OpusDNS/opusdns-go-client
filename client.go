// Package opusdns provides a Go client library for the OpusDNS API.
//
// This library enables DNS zone and record management through the OpusDNS API.
//
// Example usage:
//
//	client := opusdns.NewClient(&opusdns.Config{
//	    APIKey: "opk_...",
//	})
//
//	// List all zones
//	zones, err := client.ListZones()
//
//	// Get a specific zone
//	zone, err := client.GetZone("example.com")
//
//	// Create a zone
//	zone, err := client.CreateZone("example.com", nil)
//
//	// Manage records
//	err := client.UpsertRecord("example.com", opusdns.Record{...})
package opusdns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultAPIEndpoint is the production OpusDNS API endpoint.
	DefaultAPIEndpoint = "https://api.opusdns.com"

	// DefaultTTL is the default TTL for DNS records (60 seconds).
	DefaultTTL = 60

	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default number of retries for transient failures.
	DefaultMaxRetries = 3
)

// Config holds the configuration for the OpusDNS client.
type Config struct {
	// APIKey is the OpusDNS API key (format: opk_...).
	APIKey string

	// APIEndpoint is the base URL for the OpusDNS API.
	// Default: https://api.opusdns.com
	APIEndpoint string

	// TTL is the default TTL for DNS records in seconds.
	// Default: 60
	TTL int

	// HTTPTimeout is the timeout for HTTP requests.
	// Default: 30s
	HTTPTimeout time.Duration

	// MaxRetries is the maximum number of retries for transient failures.
	// Default: 3
	MaxRetries int
}

// Client is the OpusDNS API client.
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient creates a new OpusDNS API client.
//
// If config is nil, a default configuration is used (but an API key is required).
func NewClient(config *Config) *Client {
	cfg := &Config{}
	if config != nil {
		*cfg = *config
	}

	// Apply defaults
	if cfg.APIEndpoint == "" {
		cfg.APIEndpoint = DefaultAPIEndpoint
	}
	if cfg.TTL == 0 {
		cfg.TTL = DefaultTTL
	}
	if cfg.HTTPTimeout == 0 {
		cfg.HTTPTimeout = DefaultTimeout
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = DefaultMaxRetries
	}

	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
	}
}

// ----------------------------------------------------------------------------
// Types
// ----------------------------------------------------------------------------

// Zone represents a DNS zone.
type Zone struct {
	Name         string          `json:"name"`
	DNSSECStatus string          `json:"dnssec_status,omitempty"`
	CreatedOn    *time.Time      `json:"created_on,omitempty"`
	UpdatedOn    *time.Time      `json:"updated_on,omitempty"`
	RRSets       []RRSet         `json:"rrsets,omitempty"`
}

// Record represents a single DNS record within an RRSet.
type Record struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	TTL       int    `json:"ttl"`
	RData     string `json:"rdata"`
	Protected bool   `json:"protected,omitempty"`
}

// RRSet represents a resource record set (multiple records with same name/type).
type RRSet struct {
	Name    string       `json:"name"`
	Type    string       `json:"type"`
	TTL     int          `json:"ttl"`
	Records []RecordData `json:"records,omitempty"`
}

// RecordData represents the data portion of a DNS record.
type RecordData struct {
	RData     string `json:"rdata"`
	Protected bool   `json:"protected,omitempty"`
}

// RecordOperation represents an operation on a DNS record.
type RecordOperation struct {
	Op     string `json:"op"` // "upsert" or "remove"
	Record Record `json:"record"`
}

// Pagination represents pagination metadata in API responses.
type Pagination struct {
	TotalPages  int  `json:"total_pages"`
	CurrentPage int  `json:"current_page"`
	HasNextPage bool `json:"has_next_page"`
}

// DNSChanges represents the response from operations that modify DNS records.
type DNSChanges struct {
	ChangesetID string      `json:"changeset_id"`
	ZoneName    string      `json:"zone_name"`
	NumChanges  int         `json:"num_changes"`
	Changes     []DNSChange `json:"changes"`
}

// DNSChange represents a single change in a changeset.
type DNSChange struct {
	Action     string `json:"action"`
	RRSetName  string `json:"rrset_name,omitempty"`
	RRSetType  string `json:"rrset_type,omitempty"`
	RecordData string `json:"record_data,omitempty"`
	TTL        int    `json:"ttl,omitempty"`
}

// APIError represents an error response from the OpusDNS API.
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Body       string `json:"-"` // Raw response body (not serialized)
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("opusdns: API error %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("opusdns: API error %d", e.StatusCode)
}

// ----------------------------------------------------------------------------
// Internal types for API communication
// ----------------------------------------------------------------------------

type zoneListResponse struct {
	Results    []Zone     `json:"results"`
	Pagination Pagination `json:"pagination"`
}

type zoneCreateRequest struct {
	Name   string            `json:"name"`
	RRSets []rrsetCreateData `json:"rrsets,omitempty"`
}

type rrsetCreateData struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	TTL     int      `json:"ttl"`
	Records []string `json:"records"`
}

type recordPatchRequest struct {
	Ops []RecordOperation `json:"ops"`
}

// ----------------------------------------------------------------------------
// HTTP helpers
// ----------------------------------------------------------------------------

// doRequest executes an HTTP request with retry logic for transient failures.
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, ...
			time.Sleep(time.Duration(1<<uint(attempt-1)) * time.Second)
		}

		var reqBody io.Reader
		if body != nil {
			data, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("opusdns: failed to marshal request: %w", err)
			}
			reqBody = bytes.NewReader(data)
		}

		reqURL := strings.TrimSuffix(c.config.APIEndpoint, "/") + path
		req, err := http.NewRequest(method, reqURL, reqBody)
		if err != nil {
			return nil, fmt.Errorf("opusdns: failed to create request: %w", err)
		}

		req.Header.Set("X-Api-Key", c.config.APIKey)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("opusdns: request failed: %w", err)
			continue
		}

		// Success
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Read error response
		bodyBytes, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		// Retry on rate limiting (429) and server errors (5xx)
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = &APIError{StatusCode: resp.StatusCode, Body: string(bodyBytes)}
			continue
		}

		// Parse error message
		var errResp map[string]interface{}
		var errMsg string
		if json.Unmarshal(bodyBytes, &errResp) == nil {
			if msg, ok := errResp["message"].(string); ok {
				errMsg = msg
			} else if code, ok := errResp["error_code"].(string); ok {
				errMsg = code
			}
		}

		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    errMsg,
			Body:       string(bodyBytes),
		}
	}

	return nil, fmt.Errorf("opusdns: max retries exceeded: %w", lastErr)
}

// decodeResponse reads and decodes a JSON response body.
func decodeResponse[T any](resp *http.Response) (*T, error) {
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("opusdns: failed to read response: %w", err)
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("opusdns: failed to parse response: %w", err)
	}

	return &result, nil
}

// ----------------------------------------------------------------------------
// Zone operations
// ----------------------------------------------------------------------------

// ListZones retrieves all DNS zones with automatic pagination.
func (c *Client) ListZones() ([]Zone, error) {
	var zones []Zone
	page := 1

	for {
		path := fmt.Sprintf("/v1/dns?page=%d&page_size=100", page)
		resp, err := c.doRequest("GET", path, nil)
		if err != nil {
			return nil, err
		}

		result, err := decodeResponse[zoneListResponse](resp)
		if err != nil {
			return nil, err
		}

		zones = append(zones, result.Results...)

		if !result.Pagination.HasNextPage {
			break
		}
		page++
	}

	return zones, nil
}

// GetZone retrieves a specific zone by name.
func (c *Client) GetZone(name string) (*Zone, error) {
	name = strings.TrimSuffix(name, ".")
	path := "/v1/dns/" + url.PathEscape(name)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	return decodeResponse[Zone](resp)
}

// CreateZone creates a new DNS zone.
//
// The records parameter is optional; pass nil to create an empty zone.
func (c *Client) CreateZone(name string, records []Record) (*Zone, error) {
	name = strings.TrimSuffix(name, ".")

	req := zoneCreateRequest{Name: name}

	// Group records by name+type into RRSets
	if len(records) > 0 {
		rrsetMap := make(map[string]*rrsetCreateData)
		for _, r := range records {
			key := r.Name + "|" + r.Type
			if rrset, exists := rrsetMap[key]; exists {
				rrset.Records = append(rrset.Records, r.RData)
			} else {
				rrsetMap[key] = &rrsetCreateData{
					Name:    r.Name,
					Type:    r.Type,
					TTL:     r.TTL,
					Records: []string{r.RData},
				}
			}
		}
		for _, rrset := range rrsetMap {
			req.RRSets = append(req.RRSets, *rrset)
		}
	}

	resp, err := c.doRequest("POST", "/v1/dns", req)
	if err != nil {
		return nil, err
	}

	return decodeResponse[Zone](resp)
}

// DeleteZone deletes a DNS zone.
func (c *Client) DeleteZone(name string) error {
	name = strings.TrimSuffix(name, ".")
	path := "/v1/dns/" + url.PathEscape(name)

	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	return nil
}

// ----------------------------------------------------------------------------
// DNSSEC operations
// ----------------------------------------------------------------------------

// EnableDNSSEC enables DNSSEC for a zone.
func (c *Client) EnableDNSSEC(zoneName string) (*DNSChanges, error) {
	zoneName = strings.TrimSuffix(zoneName, ".")
	path := fmt.Sprintf("/v1/dns/%s/dnssec/enable", url.PathEscape(zoneName))

	resp, err := c.doRequest("POST", path, nil)
	if err != nil {
		return nil, err
	}

	return decodeResponse[DNSChanges](resp)
}

// DisableDNSSEC disables DNSSEC for a zone.
func (c *Client) DisableDNSSEC(zoneName string) (*DNSChanges, error) {
	zoneName = strings.TrimSuffix(zoneName, ".")
	path := fmt.Sprintf("/v1/dns/%s/dnssec/disable", url.PathEscape(zoneName))

	resp, err := c.doRequest("POST", path, nil)
	if err != nil {
		return nil, err
	}

	return decodeResponse[DNSChanges](resp)
}

// ----------------------------------------------------------------------------
// Record operations
// ----------------------------------------------------------------------------

// GetRecords retrieves all record sets for a zone.
func (c *Client) GetRecords(zoneName string) ([]RRSet, error) {
	zoneName = strings.TrimSuffix(zoneName, ".")
	path := fmt.Sprintf("/v1/dns/%s/rrsets", url.PathEscape(zoneName))

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	result, err := decodeResponse[[]RRSet](resp)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

// PatchRecords applies multiple record operations atomically.
func (c *Client) PatchRecords(zoneName string, ops []RecordOperation) error {
	zoneName = strings.TrimSuffix(zoneName, ".")
	path := fmt.Sprintf("/v1/dns/%s/records", url.PathEscape(zoneName))

	resp, err := c.doRequest("PATCH", path, recordPatchRequest{Ops: ops})
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	return nil
}

// UpsertRecord creates or updates a single DNS record.
func (c *Client) UpsertRecord(zoneName string, record Record) error {
	return c.PatchRecords(zoneName, []RecordOperation{
		{Op: "upsert", Record: record},
	})
}

// DeleteRecord removes a single DNS record.
func (c *Client) DeleteRecord(zoneName string, record Record) error {
	return c.PatchRecords(zoneName, []RecordOperation{
		{Op: "remove", Record: record},
	})
}

// ----------------------------------------------------------------------------
// Convenience methods
// ----------------------------------------------------------------------------

// DefaultTTL returns the configured default TTL.
func (c *Client) DefaultTTL() int {
	return c.config.TTL
}
