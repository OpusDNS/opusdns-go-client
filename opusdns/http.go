// Package opusdns provides a Go client library for the OpusDNS API.
package opusdns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// HTTPClient is the low-level HTTP transport for the OpusDNS API.
// It handles authentication, retries, rate limiting, and request/response serialization.
type HTTPClient struct {
	config     *Config
	httpClient *http.Client
	baseURL    *url.URL

	// Rate limiting
	mu          sync.Mutex
	rateLimited bool
	retryAfter  time.Time
}

// NewHTTPClient creates a new low-level HTTP client with the given configuration.
func NewHTTPClient(config *Config) (*HTTPClient, error) {
	if config == nil {
		config = NewConfig()
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Parse base URL
	baseURL, err := url.Parse(strings.TrimSuffix(config.APIEndpoint, "/"))
	if err != nil {
		return nil, &ConfigError{Field: "APIEndpoint", Message: fmt.Sprintf("invalid URL: %v", err)}
	}

	// Use provided HTTP client or create default
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: config.HTTPTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		}
	}

	return &HTTPClient{
		config:     config,
		httpClient: httpClient,
		baseURL:    baseURL,
	}, nil
}

// Request represents an HTTP request to the OpusDNS API.
type Request struct {
	Method      string
	Path        string
	Query       url.Values
	Body        interface{}
	Headers     http.Header
	ContentType string
}

// Response represents an HTTP response from the OpusDNS API.
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// Do executes an HTTP request with retry logic and returns the response.
func (c *HTTPClient) Do(ctx context.Context, req *Request) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		// Check if we should wait due to rate limiting
		if err := c.waitForRateLimit(ctx); err != nil {
			return nil, err
		}

		// Calculate backoff delay for retries
		if attempt > 0 {
			delay := c.calculateBackoff(attempt)
			c.logf("Retry attempt %d after %v", attempt, delay)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		// Execute the request
		resp, err := c.doRequest(ctx, req)
		if err != nil {
			lastErr = err

			// Don't retry on context errors
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			// Retry on network errors
			c.logf("Request failed (attempt %d): %v", attempt+1, err)
			continue
		}

		// Handle rate limiting
		if resp.StatusCode == http.StatusTooManyRequests {
			c.handleRateLimit(resp)
			lastErr = NewAPIError(&http.Response{StatusCode: resp.StatusCode, Header: resp.Headers}, resp.Body)
			continue
		}

		// Retry on server errors (5xx)
		if resp.StatusCode >= 500 {
			lastErr = NewAPIError(&http.Response{StatusCode: resp.StatusCode, Header: resp.Headers}, resp.Body)
			c.logf("Server error %d (attempt %d)", resp.StatusCode, attempt+1)
			continue
		}

		// Return response (success or client error)
		return resp, nil
	}

	return nil, fmt.Errorf("opusdns: max retries exceeded: %w", lastErr)
}

// doRequest performs a single HTTP request without retries.
func (c *HTTPClient) doRequest(ctx context.Context, req *Request) (*Response, error) {
	// Build URL
	reqURL := c.baseURL.JoinPath(req.Path)
	if req.Query != nil {
		reqURL.RawQuery = req.Query.Encode()
	}

	// Serialize body
	var bodyReader io.Reader
	if req.Body != nil {
		data, err := json.Marshal(req.Body)
		if err != nil {
			return nil, &RequestError{Op: "marshal", URL: reqURL.String(), Err: err}
		}
		bodyReader = bytes.NewReader(data)
		c.logf("Request body: %s", string(data))
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, reqURL.String(), bodyReader)
	if err != nil {
		return nil, &RequestError{Op: "create", URL: reqURL.String(), Err: err}
	}

	// Set headers
	httpReq.Header.Set("X-Api-Key", c.config.APIKey)
	httpReq.Header.Set("User-Agent", c.config.UserAgent)
	httpReq.Header.Set("Accept", "application/json")

	if req.Body != nil {
		contentType := req.ContentType
		if contentType == "" {
			contentType = "application/json"
		}
		httpReq.Header.Set("Content-Type", contentType)
	}

	// Copy custom headers
	for key, values := range req.Headers {
		for _, value := range values {
			httpReq.Header.Add(key, value)
		}
	}

	c.logf("%s %s", req.Method, reqURL.String())

	// Execute request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &RequestError{Op: "execute", URL: reqURL.String(), Err: err}
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, &RequestError{Op: "read", URL: reqURL.String(), Err: err}
	}

	c.logf("Response: %d %s", httpResp.StatusCode, string(body))

	return &Response{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header,
		Body:       body,
	}, nil
}

// Get performs a GET request.
func (c *HTTPClient) Get(ctx context.Context, path string, query url.Values) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodGet,
		Path:   path,
		Query:  query,
	})
}

// Post performs a POST request with a JSON body.
func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodPost,
		Path:   path,
		Body:   body,
	})
}

// Put performs a PUT request with a JSON body.
func (c *HTTPClient) Put(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodPut,
		Path:   path,
		Body:   body,
	})
}

// Patch performs a PATCH request with a JSON body.
func (c *HTTPClient) Patch(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodPatch,
		Path:   path,
		Body:   body,
	})
}

// Delete performs a DELETE request.
func (c *HTTPClient) Delete(ctx context.Context, path string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method: http.MethodDelete,
		Path:   path,
	})
}

// DecodeResponse decodes the response body into the given target.
// Returns an APIError if the response indicates an error (status >= 400).
func (c *HTTPClient) DecodeResponse(resp *Response, target interface{}) error {
	// Check for error responses
	if resp.StatusCode >= 400 {
		return NewAPIError(&http.Response{
			StatusCode: resp.StatusCode,
			Header:     resp.Headers,
		}, resp.Body)
	}

	// Handle empty responses (204 No Content, etc.)
	if len(resp.Body) == 0 || target == nil {
		return nil
	}

	// Decode JSON response
	if err := json.Unmarshal(resp.Body, target); err != nil {
		return fmt.Errorf("opusdns: failed to decode response: %w", err)
	}

	return nil
}

// calculateBackoff calculates the backoff duration for a retry attempt.
// Uses exponential backoff with jitter.
func (c *HTTPClient) calculateBackoff(attempt int) time.Duration {
	// Calculate exponential backoff: min * 2^attempt
	backoff := float64(c.config.RetryWaitMin) * math.Pow(2, float64(attempt-1))

	// Apply maximum cap
	if backoff > float64(c.config.RetryWaitMax) {
		backoff = float64(c.config.RetryWaitMax)
	}

	// Add jitter (±20%)
	jitter := backoff * 0.2 * (0.5 - float64(time.Now().UnixNano()%100)/100)
	backoff += jitter

	return time.Duration(backoff)
}

// handleRateLimit processes a 429 rate limit response.
func (c *HTTPClient) handleRateLimit(resp *Response) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.rateLimited = true

	// Parse Retry-After header
	retryAfter := c.config.RetryWaitMax
	if retryAfterStr := resp.Headers.Get("Retry-After"); retryAfterStr != "" {
		if seconds, err := strconv.Atoi(retryAfterStr); err == nil {
			retryAfter = time.Duration(seconds) * time.Second
		} else if t, err := http.ParseTime(retryAfterStr); err == nil {
			retryAfter = time.Until(t)
		}
	}

	c.retryAfter = time.Now().Add(retryAfter)
	c.logf("Rate limited, will retry after %v", retryAfter)
}

// waitForRateLimit blocks until the rate limit period has passed.
func (c *HTTPClient) waitForRateLimit(ctx context.Context) error {
	c.mu.Lock()
	if !c.rateLimited || time.Now().After(c.retryAfter) {
		c.rateLimited = false
		c.mu.Unlock()
		return nil
	}

	waitDuration := time.Until(c.retryAfter)
	c.mu.Unlock()

	c.logf("Waiting %v for rate limit", waitDuration)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitDuration):
		c.mu.Lock()
		c.rateLimited = false
		c.mu.Unlock()
		return nil
	}
}

// logf logs a debug message if debug logging is enabled.
func (c *HTTPClient) logf(format string, args ...interface{}) {
	if !c.config.Debug {
		return
	}

	msg := fmt.Sprintf(format, args...)
	if c.config.Logger != nil {
		c.config.Logger.Printf("[opusdns] %s", msg)
	} else {
		fmt.Printf("[opusdns] %s\n", msg)
	}
}

// BuildPath constructs an API path with the configured version prefix.
func (c *HTTPClient) BuildPath(parts ...string) string {
	allParts := make([]string, 0, len(parts)+1)
	allParts = append(allParts, c.config.APIVersion)
	allParts = append(allParts, parts...)
	return "/" + strings.Join(allParts, "/")
}

// BuildQuery creates a url.Values from a map, omitting empty values.
func BuildQuery(params map[string]string) url.Values {
	query := url.Values{}
	for key, value := range params {
		if value != "" {
			query.Set(key, value)
		}
	}
	return query
}

// PaginationParams represents pagination query parameters.
type PaginationParams struct {
	Page      int
	PageSize  int
	SortBy    string
	SortOrder string
}

// ToQuery converts pagination params to URL query values.
func (p *PaginationParams) ToQuery() url.Values {
	query := url.Values{}
	if p.Page > 0 {
		query.Set("page", strconv.Itoa(p.Page))
	}
	if p.PageSize > 0 {
		query.Set("page_size", strconv.Itoa(p.PageSize))
	}
	if p.SortBy != "" {
		query.Set("sort_by", p.SortBy)
	}
	if p.SortOrder != "" {
		query.Set("sort_order", p.SortOrder)
	}
	return query
}

// MergeQuery merges multiple url.Values into one.
func MergeQuery(queries ...url.Values) url.Values {
	result := url.Values{}
	for _, q := range queries {
		for key, values := range q {
			for _, value := range values {
				result.Add(key, value)
			}
		}
	}
	return result
}
