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
//	// Create a zone
//	err := client.CreateZone("example.com", nil)
//
//	// Add a TXT record
//	err := client.UpsertTXTRecord("_acme-challenge.example.com", "challenge-value")
//
//	// Remove a TXT record
//	err := client.RemoveTXTRecord("_acme-challenge.example.com", "challenge-value")
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
	// DefaultAPIEndpoint is the production OpusDNS API endpoint
	DefaultAPIEndpoint = "https://api.opusdns.com"

	// DefaultTTL is the default TTL for DNS records (60 seconds)
	DefaultTTL = 60

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default number of retries for transient failures
	DefaultMaxRetries = 3
)

// Config holds the configuration for the OpusDNS client.
type Config struct {
	// APIKey is the OpusDNS API key (format: opk_...)
	APIKey string

	// APIEndpoint is the base URL for the OpusDNS API (default: https://api.opusdns.com)
	APIEndpoint string

	// TTL is the default TTL for DNS records in seconds (default: 60)
	TTL int

	// HTTPTimeout is the timeout for HTTP requests (default: 30s)
	HTTPTimeout time.Duration

	// MaxRetries is the maximum number of retries for transient failures (default: 3)
	MaxRetries int
}

// Client is the OpusDNS API client.
type Client struct {
	config     *Config
	httpClient *http.Client
}

// Zone represents a DNS zone in OpusDNS.
type Zone struct {
	Name         string     `json:"name"`
	DNSSECStatus string     `json:"dnssec_status,omitempty"`
	CreatedOn    *time.Time `json:"created_on,omitempty"`
	UpdatedOn    *time.Time `json:"updated_on,omitempty"`
	RRSets       []RRSetResponse `json:"rrsets,omitempty"`
}

// ZoneCreateRequest represents a request to create a new zone.
type ZoneCreateRequest struct {
	Name         string           `json:"name"`
	DNSSECStatus string           `json:"dnssec_status,omitempty"`
	RRSets       []RRSetCreateRequest `json:"rrsets,omitempty"`
}

// RRSetCreateRequest represents a record set for zone creation.
type RRSetCreateRequest struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	TTL     int      `json:"ttl"`
	Records []string `json:"records"`
}

// RRSetResponse represents a resource record set from the API response.
type RRSetResponse struct {
	Name    string          `json:"name"`
	Type    string          `json:"type"`
	TTL     int             `json:"ttl"`
	Records []RecordResponse `json:"records,omitempty"`
}

// RecordResponse represents a single DNS record in an RRSet.
type RecordResponse struct {
	RData     string `json:"rdata"`
	Protected bool   `json:"protected,omitempty"`
}

// RRSet represents a resource record for patch operations.
type RRSet struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	TTL   int    `json:"ttl"`
	RData string `json:"rdata"`
}

// RRSetOperation represents an operation on a Record.
type RRSetOperation struct {
	Op     string `json:"op"` // "upsert" or "remove"
	Record RRSet  `json:"record"`
}

// RRSetPatchRequest represents a PATCH request to /v1/dns/{zone}/records.
type RRSetPatchRequest struct {
	Ops []RRSetOperation `json:"ops"`
}

// ZoneListResponse represents the response from GET /v1/dns.
type ZoneListResponse struct {
	Results    []Zone     `json:"results"`
	Pagination Pagination `json:"pagination"`
}

// Pagination represents pagination metadata.
type Pagination struct {
	TotalPages  int  `json:"total_pages"`
	CurrentPage int  `json:"current_page"`
	HasNextPage bool `json:"has_next_page"`
}

// DNSChangesResponse represents the response from DNSSEC enable/disable.
type DNSChangesResponse struct {
	ChangesetID string       `json:"changeset_id"`
	ZoneName    string       `json:"zone_name"`
	NumChanges  int          `json:"num_changes"`
	Changes     []DNSChange  `json:"changes"`
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
//
// WARNING: The Body field may contain sensitive data from the API response,
// including API keys or tokens if they were echoed back. Avoid logging or
// exposing this field in production environments without sanitization.
type APIError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("OpusDNS API error (HTTP %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("OpusDNS API error (HTTP %d): %s", e.StatusCode, e.Body)
}

// NewClient creates a new OpusDNS API client with the given configuration.
// If config is nil, a default configuration is used.
// The provided config is copied internally to prevent mutation of the caller's struct.
func NewClient(config *Config) *Client {
	// Create a copy to avoid mutating the caller's config
	cfg := &Config{}
	if config != nil {
		*cfg = *config
	}

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

// doRequest executes an HTTP request with retry logic for transient failures.
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
		}

		var reqBody io.Reader
		if body != nil {
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewReader(jsonData)
		}

		url := strings.TrimSuffix(c.config.APIEndpoint, "/") + path
		req, err := http.NewRequest(method, url, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("X-Api-Key", c.config.APIKey)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %w", err)
			continue
		}

		// Success cases
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Read response body for error details
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Retry on rate limiting and server errors
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = &APIError{
				StatusCode: resp.StatusCode,
				Body:       string(bodyBytes),
			}
			continue
		}

		// Don't retry on client errors
		var errMsg string
		var errResp map[string]interface{}
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

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// ListZones retrieves all DNS zones from the OpusDNS API with pagination.
func (c *Client) ListZones() ([]Zone, error) {
	var allZones []Zone
	page := 1
	pageSize := 100

	for {
		path := fmt.Sprintf("/v1/dns?page=%d&page_size=%d", page, pageSize)
		resp, err := c.doRequest("GET", path, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list zones (page %d): %w", page, err)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var zoneResp ZoneListResponse
		if err := json.Unmarshal(bodyBytes, &zoneResp); err != nil {
			return nil, fmt.Errorf("failed to parse zone list response: %w", err)
		}

		allZones = append(allZones, zoneResp.Results...)

		if !zoneResp.Pagination.HasNextPage {
			break
		}
		page++
	}

	return allZones, nil
}

// FindZoneForFQDN finds the appropriate zone for a given FQDN.
// It iterates through domain parts and checks each against the API.
func (c *Client) FindZoneForFQDN(fqdn string) (string, error) {
	// Normalize FQDN (remove trailing dot)
	fqdn = strings.TrimSuffix(fqdn, ".")
	parts := strings.Split(fqdn, ".")

	// Start from second part (skip first like _acme-challenge)
	for i := 1; i < len(parts); i++ {
		candidate := strings.Join(parts[i:], ".")

		// Check if this zone exists via API
		resp, err := c.doRequest("GET", "/v1/dns/"+candidate, nil)
		if err != nil {
			// API error (not found, etc.) - try next candidate
			continue
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		// Valid zone response contains dnssec_status
		var zoneResp map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &zoneResp); err != nil {
			continue
		}

		if _, ok := zoneResp["dnssec_status"]; ok {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("no zone found for FQDN %s", fqdn)
}

// RRSetListResponse represents the response from GET /v1/dns/{zone}/records.
type RRSetListResponse struct {
	RRSets []RRSet `json:"rrsets"`
}

// ListRRSets retrieves all RRSets for a given zone.
func (c *Client) ListRRSets(zone string) ([]RRSet, error) {
	zone = strings.TrimSuffix(zone, ".")
	path := fmt.Sprintf("/v1/dns/%s/records", zone)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list RRSets for zone %s: %w", zone, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var rrsetResp RRSetListResponse
	if err := json.Unmarshal(bodyBytes, &rrsetResp); err != nil {
		return nil, fmt.Errorf("failed to parse RRSet list response: %w", err)
	}

	return rrsetResp.RRSets, nil
}

// PatchRRSets applies a patch operation to RRSets in a zone.
func (c *Client) PatchRRSets(zone string, ops []RRSetOperation) error {
	zone = strings.TrimSuffix(zone, ".")
	path := fmt.Sprintf("/v1/dns/%s/records", zone)

	req := RRSetPatchRequest{Ops: ops}

	resp, err := c.doRequest("PATCH", path, req)
	if err != nil {
		return fmt.Errorf("failed to patch RRSets in zone %s: %w", zone, err)
	}
	defer resp.Body.Close()

	// 204 No Content is expected for successful PATCH
	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// UpsertTXTRecord creates or updates a TXT record for the given FQDN.
// It automatically detects the appropriate zone.
func (c *Client) UpsertTXTRecord(fqdn, value string) error {
	zone, err := c.FindZoneForFQDN(fqdn)
	if err != nil {
		return err
	}

	// Extract record name (remove zone suffix)
	recordName := strings.TrimSuffix(fqdn, ".")
	zoneFQDN := zone + "."
	if strings.HasSuffix(recordName+".", zoneFQDN) {
		recordName = strings.TrimSuffix(recordName, "."+zone)
	}

	// Ensure value is quoted for TXT records
	if !strings.HasPrefix(value, "\"") {
		value = "\"" + value + "\""
	}

	op := RRSetOperation{
		Op: "upsert",
		Record: RRSet{
			Name:  recordName,
			Type:  "TXT",
			TTL:   c.config.TTL,
			RData: value,
		},
	}

	return c.PatchRRSets(zone, []RRSetOperation{op})
}

// RemoveTXTRecord removes a TXT record for the given FQDN.
// It automatically detects the appropriate zone.
// The value parameter is required to specify the complete record for removal.
func (c *Client) RemoveTXTRecord(fqdn, value string) error {
	zone, err := c.FindZoneForFQDN(fqdn)
	if err != nil {
		return err
	}

	// Extract record name (remove zone suffix)
	recordName := strings.TrimSuffix(fqdn, ".")
	zoneFQDN := zone + "."
	if strings.HasSuffix(recordName+".", zoneFQDN) {
		recordName = strings.TrimSuffix(recordName, "."+zone)
	}

	// Ensure value is quoted for TXT records
	if !strings.HasPrefix(value, "\"") {
		value = "\"" + value + "\""
	}

	op := RRSetOperation{
		Op: "remove",
		Record: RRSet{
			Name:  recordName,
			Type:  "TXT",
			TTL:   c.config.TTL,
			RData: value,
		},
	}

	return c.PatchRRSets(zone, []RRSetOperation{op})
}

// GetZone retrieves details for a specific zone.
func (c *Client) GetZone(zoneName string) (*Zone, error) {
	zoneName = strings.TrimSuffix(zoneName, ".")
	path := "/v1/dns/" + url.PathEscape(zoneName)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone %s: %w", zoneName, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var zone Zone
	if err := json.Unmarshal(bodyBytes, &zone); err != nil {
		return nil, fmt.Errorf("failed to parse zone response: %w", err)
	}

	return &zone, nil
}

// CreateZone creates a new DNS zone.
// The rrsets parameter is optional and can be nil to create an empty zone.
func (c *Client) CreateZone(zoneName string, rrsets []RRSetCreateRequest) (*Zone, error) {
	zoneName = strings.TrimSuffix(zoneName, ".")

	req := ZoneCreateRequest{
		Name:   zoneName,
		RRSets: rrsets,
	}

	resp, err := c.doRequest("POST", "/v1/dns", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create zone %s: %w", zoneName, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var zone Zone
	if err := json.Unmarshal(bodyBytes, &zone); err != nil {
		return nil, fmt.Errorf("failed to parse zone response: %w", err)
	}

	return &zone, nil
}

// DeleteZone deletes a DNS zone.
func (c *Client) DeleteZone(zoneName string) error {
	zoneName = strings.TrimSuffix(zoneName, ".")
	path := "/v1/dns/" + url.PathEscape(zoneName)

	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete zone %s: %w", zoneName, err)
	}
	defer resp.Body.Close()

	return nil
}

// EnableDNSSEC enables DNSSEC for a zone.
func (c *Client) EnableDNSSEC(zoneName string) (*DNSChangesResponse, error) {
	zoneName = strings.TrimSuffix(zoneName, ".")
	path := fmt.Sprintf("/v1/dns/%s/dnssec/enable", url.PathEscape(zoneName))

	resp, err := c.doRequest("POST", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to enable DNSSEC for zone %s: %w", zoneName, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var changes DNSChangesResponse
	if err := json.Unmarshal(bodyBytes, &changes); err != nil {
		return nil, fmt.Errorf("failed to parse DNSSEC response: %w", err)
	}

	return &changes, nil
}

// DisableDNSSEC disables DNSSEC for a zone.
func (c *Client) DisableDNSSEC(zoneName string) (*DNSChangesResponse, error) {
	zoneName = strings.TrimSuffix(zoneName, ".")
	path := fmt.Sprintf("/v1/dns/%s/dnssec/disable", url.PathEscape(zoneName))

	resp, err := c.doRequest("POST", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to disable DNSSEC for zone %s: %w", zoneName, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var changes DNSChangesResponse
	if err := json.Unmarshal(bodyBytes, &changes); err != nil {
		return nil, fmt.Errorf("failed to parse DNSSEC response: %w", err)
	}

	return &changes, nil
}

// GetRRSets retrieves all resource record sets for a zone.
func (c *Client) GetRRSets(zoneName string) ([]RRSetResponse, error) {
	zoneName = strings.TrimSuffix(zoneName, ".")
	path := fmt.Sprintf("/v1/dns/%s/rrsets", url.PathEscape(zoneName))

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get RRSets for zone %s: %w", zoneName, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var rrsets []RRSetResponse
	if err := json.Unmarshal(bodyBytes, &rrsets); err != nil {
		return nil, fmt.Errorf("failed to parse RRSets response: %w", err)
	}

	return rrsets, nil
}

// UpsertRecord creates or updates a DNS record of any type.
func (c *Client) UpsertRecord(zoneName string, record RRSet) error {
	zoneName = strings.TrimSuffix(zoneName, ".")

	op := RRSetOperation{
		Op:     "upsert",
		Record: record,
	}

	return c.PatchRRSets(zoneName, []RRSetOperation{op})
}

// RemoveRecord removes a DNS record.
func (c *Client) RemoveRecord(zoneName string, record RRSet) error {
	zoneName = strings.TrimSuffix(zoneName, ".")

	op := RRSetOperation{
		Op:     "remove",
		Record: record,
	}

	return c.PatchRRSets(zoneName, []RRSetOperation{op})
}
