// Package opusdns provides a Go client library for the OpusDNS API.
//
// This library enables DNS zone and record management through the OpusDNS API,
// with specific support for ACME DNS-01 challenge workflows.
//
// Example usage:
//
//	config := &opusdns.Config{
//	    APIKey:      "opk_...",
//	    APIEndpoint: "https://api.opusdns.com",
//	    TTL:         60,
//	}
//
//	client := opusdns.NewClient(config)
//
//	// Add ACME challenge record
//	err := client.UpsertTXTRecord("_acme-challenge.example.com", "challenge-value")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Wait for DNS propagation
//	err = client.WaitForPropagation("_acme-challenge.example.com", "challenge-value")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Clean up
//	err = client.RemoveTXTRecord("_acme-challenge.example.com", "TXT")
//	if err != nil {
//	    log.Printf("cleanup error: %v", err)
//	}
package opusdns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultAPIEndpoint is the production OpusDNS API endpoint
	DefaultAPIEndpoint = "https://api.opusdns.com"

	// DefaultTTL is the default TTL for DNS records (60 seconds for ACME challenges)
	DefaultTTL = 60

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default number of retries for transient failures
	DefaultMaxRetries = 3

	// DefaultPollingInterval is the default interval between DNS propagation checks
	DefaultPollingInterval = 6 * time.Second

	// DefaultPollingTimeout is the default maximum time to wait for DNS propagation
	DefaultPollingTimeout = 60 * time.Second
)

// Config holds the configuration for the OpusDNS client.
type Config struct {
	// APIKey is the OpusDNS API key (format: opk_{typeid}_{secret}{checksum})
	APIKey string

	// APIEndpoint is the base URL for the OpusDNS API (default: https://api.opusdns.com)
	APIEndpoint string

	// TTL is the default TTL for DNS records in seconds (default: 60)
	TTL int

	// HTTPTimeout is the timeout for HTTP requests (default: 30s)
	HTTPTimeout time.Duration

	// MaxRetries is the maximum number of retries for transient failures (default: 3)
	MaxRetries int

	// PollingInterval is the interval between DNS propagation checks (default: 6s)
	PollingInterval time.Duration

	// PollingTimeout is the maximum time to wait for DNS propagation (default: 60s)
	PollingTimeout time.Duration

	// DNSResolvers is the list of DNS servers to query for propagation checks
	// (default: ["8.8.8.8:53", "1.1.1.1:53"])
	DNSResolvers []string
}

// Client is the OpusDNS API client.
type Client struct {
	config     *Config
	httpClient *http.Client
	zoneCacheMu sync.RWMutex
	zoneCache   []Zone
	zoneCacheTime time.Time
	zoneCacheTTL  time.Duration
}

// Zone represents a DNS zone in OpusDNS.
type Zone struct {
	Name         string    `json:"name"`
	DNSSECStatus string    `json:"dnssec_status"`
	CreatedOn    time.Time `json:"created_on"`
}

// RRSet represents a resource record set.
type RRSet struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	TTL     int      `json:"ttl"`
	Records []Record `json:"records"`
}

// Record represents a DNS record within an RRSet.
type Record struct {
	RData string `json:"rdata"`
}

// RRSetOperation represents an operation on an RRSet.
type RRSetOperation struct {
	Op    string `json:"op"` // "upsert" or "remove"
	RRSet RRSet  `json:"rrset"`
}

// RRSetPatchRequest represents a PATCH request to /v1/dns/{zone}/rrsets.
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

// APIError represents an error response from the OpusDNS API.
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
func NewClient(config *Config) *Client {
	if config.APIEndpoint == "" {
		config.APIEndpoint = DefaultAPIEndpoint
	}
	if config.TTL == 0 {
		config.TTL = DefaultTTL
	}
	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = DefaultTimeout
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = DefaultMaxRetries
	}
	if config.PollingInterval == 0 {
		config.PollingInterval = DefaultPollingInterval
	}
	if config.PollingTimeout == 0 {
		config.PollingTimeout = DefaultPollingTimeout
	}
	if len(config.DNSResolvers) == 0 {
		config.DNSResolvers = []string{"8.8.8.8:53", "1.1.1.1:53"}
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
		zoneCacheTTL: 5 * time.Minute,
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
	// Check cache
	c.zoneCacheMu.RLock()
	if time.Since(c.zoneCacheTime) < c.zoneCacheTTL && len(c.zoneCache) > 0 {
		zones := make([]Zone, len(c.zoneCache))
		copy(zones, c.zoneCache)
		c.zoneCacheMu.RUnlock()
		return zones, nil
	}
	c.zoneCacheMu.RUnlock()

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

	// Update cache
	c.zoneCacheMu.Lock()
	c.zoneCache = allZones
	c.zoneCacheTime = time.Now()
	c.zoneCacheMu.Unlock()

	return allZones, nil
}

// FindZoneForFQDN finds the appropriate zone for a given FQDN.
// It lists all zones and matches the longest matching zone name.
func (c *Client) FindZoneForFQDN(fqdn string) (string, error) {
	// Ensure FQDN ends with a dot
	if !strings.HasSuffix(fqdn, ".") {
		fqdn = fqdn + "."
	}

	zones, err := c.ListZones()
	if err != nil {
		return "", fmt.Errorf("failed to list zones: %w", err)
	}

	if len(zones) == 0 {
		return "", fmt.Errorf("no zones found in OpusDNS account")
	}

	// Find the longest matching zone
	var matchedZone string
	for _, zone := range zones {
		zoneName := zone.Name
		if !strings.HasSuffix(zoneName, ".") {
			zoneName = zoneName + "."
		}

		if strings.HasSuffix(fqdn, zoneName) {
			if len(zoneName) > len(matchedZone) {
				matchedZone = zoneName
			}
		}
	}

	if matchedZone == "" {
		return "", fmt.Errorf("no zone found for FQDN %s (available zones: %d)", fqdn, len(zones))
	}

	// Return without trailing dot for API usage
	return strings.TrimSuffix(matchedZone, "."), nil
}

// RRSetListResponse represents the response from GET /v1/dns/{zone}/rrsets.
type RRSetListResponse struct {
	RRSets []RRSet `json:"rrsets"`
}

// ListRRSets retrieves all RRSets for a given zone.
func (c *Client) ListRRSets(zone string) ([]RRSet, error) {
	zone = strings.TrimSuffix(zone, ".")
	path := fmt.Sprintf("/v1/dns/%s/rrsets", zone)

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
	path := fmt.Sprintf("/v1/dns/%s/rrsets", zone)

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
		RRSet: RRSet{
			Name: recordName,
			Type: "TXT",
			TTL:  c.config.TTL,
			Records: []Record{
				{RData: value},
			},
		},
	}

	return c.PatchRRSets(zone, []RRSetOperation{op})
}

// RemoveTXTRecord removes a TXT record for the given FQDN.
// It automatically detects the appropriate zone.
func (c *Client) RemoveTXTRecord(fqdn, recordType string) error {
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

	op := RRSetOperation{
		Op: "remove",
		RRSet: RRSet{
			Name: recordName,
			Type: recordType,
		},
	}

	return c.PatchRRSets(zone, []RRSetOperation{op})
}
