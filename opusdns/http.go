package opusdns

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// HTTPClient is the low-level HTTP transport for the OpusDNS API.
// It handles authentication, retries, rate limiting, and request/response serialization.
type HTTPClient struct {
	config     *Config
	httpClient *http.Client
	baseURL    *url.URL

	// streamClient sends requests whose response body is copied to
	// Request.Output. When the client built its own transport, it is httpClient
	// without the overall timeout, so a long download is bounded by its context
	// instead of being cut off mid-transfer.
	streamClient *http.Client
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

	var httpClient, streamClient *http.Client
	if config.HTTPClient != nil {
		// A copy, so the redirect policy can be added without changing the
		// caller's client. Its own timeout also applies to downloads.
		custom := *config.HTTPClient
		custom.CheckRedirect = checkRedirect(custom.CheckRedirect)
		httpClient = &custom
		streamClient = &custom
	} else {
		httpClient = &http.Client{
			Timeout:       config.HTTPTimeout,
			Transport:     newTransport(config.HTTPTimeout),
			CheckRedirect: checkRedirect(nil),
		}
		stream := *httpClient
		stream.Timeout = 0
		streamClient = &stream
	}

	return &HTTPClient{
		config:       config,
		httpClient:   httpClient,
		streamClient: streamClient,
		baseURL:      baseURL,
	}, nil
}

// newTransport returns the transport of the default HTTP client: Go's default
// transport, which honours HTTP_PROXY, HTTPS_PROXY and NO_PROXY, with a larger
// connection pool. responseHeaderTimeout bounds the wait for a response to
// start, which is what limits a streamed download before its body arrives.
func newTransport(responseHeaderTimeout time.Duration) *http.Transport {
	var transport *http.Transport
	if defaultTransport, ok := http.DefaultTransport.(*http.Transport); ok {
		transport = defaultTransport.Clone()
	} else {
		transport = &http.Transport{Proxy: http.ProxyFromEnvironment}
	}
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 100
	transport.IdleConnTimeout = 90 * time.Second
	transport.ResponseHeaderTimeout = responseHeaderTimeout
	return transport
}

// credentialHeaders are the headers that authenticate a request. They are not
// sent on to another origin when the API redirects.
var credentialHeaders = []string{"X-Api-Key", "Authorization"}

// maxRedirects matches the limit of net/http's default redirect policy.
const maxRedirects = 10

// checkRedirect wraps a redirect policy so that the API key never reaches
// another origin, and a redirect that only drops a trailing slash is refused.
//
// net/http strips Authorization and Cookie when a redirect leaves the domain,
// but forwards every other header, X-Api-Key included. It also keeps them on a
// switch from https to http on the same host.
//
// The API answers a path it serves only without a trailing slash with such a
// redirect. The client produces that path when the last ID in it is empty
// (GetContact(ctx, "") requests /v1/contacts/), and following the redirect
// would decode the contact list as a contact, without an error.
func checkRedirect(next func(*http.Request, []*http.Request) error) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		prev := via[len(via)-1].URL
		if sameOrigin(prev, req.URL) && strings.HasSuffix(prev.Path, "/") && req.URL.Path == strings.TrimSuffix(prev.Path, "/") {
			return &ValidationError{
				Field:   "path",
				Message: "the API serves this path only without the trailing slash; an ID in the path is probably empty",
				Value:   prev.Path,
			}
		}

		if !sameOrigin(via[0].URL, req.URL) {
			for _, header := range credentialHeaders {
				req.Header.Del(header)
			}
		}

		if next != nil {
			return next(req, via)
		}
		if len(via) >= maxRedirects {
			return fmt.Errorf("stopped after %d redirects", maxRedirects)
		}
		return nil
	}
}

// sameOrigin reports whether both URLs have the same scheme, host and port.
func sameOrigin(a, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) && strings.EqualFold(a.Host, b.Host)
}

// Request represents an HTTP request to the OpusDNS API.
type Request struct {
	Method      string
	Path        string
	Query       url.Values
	Body        interface{}
	Headers     http.Header
	ContentType string

	// Output, if set, receives the body of a 2xx response as it arrives,
	// instead of Response.Body, which stays nil. The body of any other
	// response is still returned in Response.Body.
	Output io.Writer
}

// Response represents an HTTP response from the OpusDNS API.
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// Do executes an HTTP request with retry logic and returns the response.
//
// The request is checked before anything is sent: a path segment that is
// empty, "." or "..", a body that cannot be encoded, or a header value with a
// control character is returned as an error matching ErrInvalidInput.
//
// A 429 is retried for every method, since the API turned the request away
// without acting on it. So is a 409 or 503 that carries a Retry-After header:
// the API sends one only when it did nothing and the request will succeed
// later, for example a delete or renew of a domain whose registration is still
// being finalized. Server errors and network failures are retried only
// for methods that are safe to send twice (GET, HEAD, OPTIONS, PUT). POST,
// PATCH and DELETE may already have taken effect when they fail that way, and
// sending them again could renew, register or charge twice, or turn a delete
// that worked into a 404; they are retried only when the connection could not
// be opened at all.
//
// Between attempts the client waits with exponential backoff, or for as long
// as the response's Retry-After header asks. A Retry-After longer than
// RetryWaitMax is not waited out: that response is returned straight away.
//
// A response that is not retried, or the last one when retries run out, is
// returned as is, so DecodeResponse turns it into an APIError. If ctx ends
// while waiting to retry, the error matches both ctx's error and the failure
// of the previous attempt, e.g. ErrRateLimited.
//
// With req.Output set, a 2xx body is copied there as it arrives. A failure
// during that copy is returned as an error and not retried, since part of the
// body has been written already.
func (c *HTTPClient) Do(ctx context.Context, req *Request) (*Response, error) {
	prepared, err := c.prepare(req)
	if err != nil {
		return nil, err
	}
	resendable := isResendable(req.Method)

	var lastErr error
	for attempt := 1; ; attempt++ {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, &abortedError{ctxErr: ctxErr, last: lastErr}
		}

		resp, err := c.send(ctx, prepared, req.Output)
		retriesLeft := attempt <= c.config.MaxRetries

		var wait time.Duration
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, &abortedError{ctxErr: ctxErr, last: lastErr}
			}
			if !retriesLeft || !isRetryableRequestError(err, resendable) {
				if attempt > 1 {
					return nil, fmt.Errorf("opusdns: giving up after %d attempts: %w", attempt, err)
				}
				return nil, err
			}
			wait = c.calculateBackoff(attempt)
		} else {
			if !retriesLeft || !isRetryableStatus(resp.StatusCode, resp.Headers, resendable) {
				return resp, nil
			}
			var ok bool
			if wait, ok = c.retryDelay(attempt, resp.Headers); !ok {
				c.logf("Status %d asks to retry after more than %v; not retrying", resp.StatusCode, c.config.RetryWaitMax)
				return resp, nil
			}
			err = responseError(resp)
		}

		lastErr = err
		c.logf("Attempt %d failed (%v); retrying in %v", attempt, err, wait)
		if sleepErr := sleep(ctx, wait); sleepErr != nil {
			return nil, &abortedError{ctxErr: sleepErr, last: lastErr}
		}
	}
}

// preparedRequest is a Request checked and encoded once, and then sent as many
// times as the retry policy allows.
type preparedRequest struct {
	method string
	url    string
	header http.Header
	body   []byte
}

// prepare checks and encodes req. None of the errors it returns can be fixed
// by sending again, so they are never retried.
func (c *HTTPClient) prepare(req *Request) (*preparedRequest, error) {
	reqURL, err := c.resolveURL(req.Path)
	if err != nil {
		return nil, err
	}
	if req.Query != nil {
		reqURL.RawQuery = req.Query.Encode()
	}

	header := make(http.Header)
	header.Set("X-Api-Key", c.config.APIKey)
	header.Set("User-Agent", c.config.UserAgent)
	header.Set("Accept", "application/json")
	// Names the calling client so the API can attribute the request to an origin channel.
	// Analytics only on their side, and omitted entirely when the caller clears it.
	if c.config.ClientToken != "" {
		header.Set(clientHeader, c.config.ClientToken)
	}

	var body []byte
	if req.Body != nil {
		body, err = json.Marshal(req.Body)
		if err != nil {
			return nil, &RequestError{Op: "marshal", URL: reqURL.String(), Err: err}
		}
		contentType := req.ContentType
		if contentType == "" {
			contentType = "application/json"
		}
		header.Set("Content-Type", contentType)
	}

	// Copy custom headers
	for key, values := range req.Headers {
		for _, value := range values {
			header.Add(key, value)
		}
	}

	// net/http rejects these only when sending, with an error that looks like a
	// network failure and would be retried.
	for key, values := range header {
		if !validHeaderName(key) {
			return nil, &ValidationError{Field: "header", Message: "invalid header name", Value: key}
		}
		for _, value := range values {
			if !validHeaderValue(value) {
				// The value is left out of the error: it may be the API key.
				return nil, &ValidationError{Field: "header " + key, Message: "value contains a control character such as a newline"}
			}
		}
	}

	return &preparedRequest{
		method: req.Method,
		url:    reqURL.String(),
		header: header,
		body:   body,
	}, nil
}

// resolveURL joins path to the API endpoint. Unlike url.URL.JoinPath it does
// not resolve "." and ".." and keeps escaped characters such as %2F escaped,
// so an ID cannot move the request to another resource.
func (c *HTTPClient) resolveURL(path string) (*url.URL, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if err := checkPath(path); err != nil {
		return nil, err
	}
	decoded, err := url.PathUnescape(path)
	if err != nil {
		return nil, &ValidationError{Field: "path", Message: "invalid percent-encoding", Value: path}
	}

	reqURL := *c.baseURL
	reqURL.Path = strings.TrimSuffix(c.baseURL.Path, "/") + decoded
	reqURL.RawPath = strings.TrimSuffix(c.baseURL.EscapedPath(), "/") + path
	return &reqURL, nil
}

// checkPath rejects a path with a segment that would make the request address
// another resource than the one named: an empty segment, which an empty ID
// leaves behind, or "." and "..", which an unescaped ID can contain.
func checkPath(path string) error {
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	for i, segment := range segments {
		if segment == "" {
			// Some routes end in a slash (/v1/tlds/).
			if i == len(segments)-1 {
				continue
			}
			return &ValidationError{Field: "path", Message: "empty path segment; an ID in the path is probably empty", Value: path}
		}
		if decoded, err := url.PathUnescape(segment); err == nil && (decoded == "." || decoded == "..") {
			return &ValidationError{Field: "path", Message: fmt.Sprintf("path segment %q is not allowed", segment), Value: path}
		}
	}
	return nil
}

// validHeaderName reports whether name is an RFC 9110 token.
func validHeaderName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		b := name[i]
		switch {
		case 'a' <= b && b <= 'z', 'A' <= b && b <= 'Z', '0' <= b && b <= '9':
		case strings.IndexByte("!#$%&'*+-.^_`|~", b) >= 0:
		default:
			return false
		}
	}
	return true
}

// validHeaderValue reports whether value holds no control characters other
// than horizontal tab, as RFC 9110 requires.
func validHeaderValue(value string) bool {
	for i := 0; i < len(value); i++ {
		b := value[i]
		if (b < ' ' && b != '\t') || b == 0x7f {
			return false
		}
	}
	return true
}

// send performs one attempt. With output set, the body of a 2xx response is
// copied to it instead of being returned.
func (c *HTTPClient) send(ctx context.Context, req *preparedRequest, output io.Writer) (*Response, error) {
	var bodyReader io.Reader
	if req.body != nil {
		bodyReader = bytes.NewReader(req.body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.method, req.url, bodyReader)
	if err != nil {
		return nil, &RequestError{Op: "create", URL: req.url, Err: err}
	}
	httpReq.Header = req.header.Clone()

	c.logf("%s %s", req.method, redactURL(req.url))
	if req.body != nil {
		c.logf("Request body: %s", redactBody(req.body))
	}

	client := c.httpClient
	if output != nil {
		client = c.streamClient
	}

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, &RequestError{Op: "execute", URL: req.url, Err: err}
	}
	defer httpResp.Body.Close() //nolint:errcheck

	resp := &Response{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header,
	}

	if output != nil && httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
		written, err := io.Copy(output, httpResp.Body)
		if err != nil {
			return nil, &RequestError{Op: "stream", URL: req.url, Err: err}
		}
		c.logf("Response: %d (%d bytes written to the output)", httpResp.StatusCode, written)
		return resp, nil
	}

	// Read response body
	resp.Body, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, &RequestError{Op: "read", URL: req.url, Err: err}
	}

	c.logf("Response: %d %s", httpResp.StatusCode, redactBody(resp.Body))

	return resp, nil
}

// isResendable reports whether sending a request with this method twice is
// harmless. DELETE is idempotent in RFC 9110 terms, but a resent DELETE that
// had worked the first time fails with a 404.
func isResendable(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut:
		return true
	}
	return false
}

// isRetryableStatus reports whether a response with this status and headers
// is retried.
func isRetryableStatus(status int, headers http.Header, resendable bool) bool {
	if status == http.StatusTooManyRequests || askedToRetry(status, headers) {
		return true
	}
	return status >= 500 && resendable
}

// askedToRetry reports whether a 409 or 503 carries a Retry-After header, the
// API's sign that it did not act on the request and will accept it later.
func askedToRetry(status int, headers http.Header) bool {
	if status != http.StatusConflict && status != http.StatusServiceUnavailable {
		return false
	}
	_, found := parseRetryAfter(headers.Get("Retry-After"))
	return found
}

// isRetryableRequestError reports whether a request that failed without a
// response is sent again.
func isRetryableRequestError(err error, resendable bool) bool {
	// Refused by the redirect policy.
	if errors.Is(err, ErrInvalidInput) {
		return false
	}
	// A request that could not be built, or a body partly written to
	// Request.Output already.
	var requestErr *RequestError
	if errors.As(err, &requestErr) && (requestErr.Op == "create" || requestErr.Op == "stream") {
		return false
	}
	// An untrusted certificate stays untrusted.
	var certErr *tls.CertificateVerificationError
	if errors.As(err, &certErr) {
		return false
	}
	// The request may have reached the API; only resend it if that is harmless.
	return resendable || isDialError(err)
}

// isDialError reports whether err happened while opening the connection,
// before any part of the request was sent.
func isDialError(err error) bool {
	var opErr *net.OpError
	return errors.As(err, &opErr) && opErr.Op == "dial"
}

// abortedError is returned when the context ends before a request completes.
// It matches the context's error, ErrTimeout when the deadline passed, and the
// failure of the previous attempt, if there was one: a request that ran out of
// time while waiting out a 429 still matches ErrRateLimited.
type abortedError struct {
	ctxErr error
	last   error
}

func (e *abortedError) Error() string {
	if e.last == nil {
		return fmt.Sprintf("opusdns: request aborted: %v", e.ctxErr)
	}
	return fmt.Sprintf("opusdns: request aborted (%v) after a failed attempt: %v", e.ctxErr, e.last)
}

func (e *abortedError) Unwrap() []error {
	if e.last == nil {
		return []error{e.ctxErr}
	}
	return []error{e.ctxErr, e.last}
}

func (e *abortedError) Is(target error) bool {
	return target == ErrTimeout && errors.Is(e.ctxErr, context.DeadlineExceeded)
}

// sleep waits for d or until ctx ends, whichever comes first.
func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
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

const (
	// clientHeader carries the RFC 9110 product token identifying the calling client.
	clientHeader = "X-OpusDNS-Client"
)

// timestampRegex matches a JSON string holding an ISO 8601 timestamp without
// timezone info, with or without fractional seconds (e.g. "2026-01-23T08:26:55").
var timestampRegex = regexp.MustCompile(`"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?"`)

// fixTimestamps adds a UTC timezone suffix to JSON string values that are
// timestamps without timezone information. Object keys and quotes inside other
// strings are left alone.
func fixTimestamps(data []byte) []byte {
	matches := timestampRegex.FindAllIndex(data, -1)
	if matches == nil {
		return data
	}

	fixed := make([]byte, 0, len(data)+len(matches))
	last := 0
	for _, match := range matches {
		start, end := match[0], match[1]
		// An escaped quote: the match sits inside a longer string.
		if start > 0 && data[start-1] == '\\' {
			continue
		}
		if isObjectKey(data[end:]) {
			continue
		}
		fixed = append(fixed, data[last:end-1]...)
		fixed = append(fixed, 'Z', '"')
		last = end
	}
	return append(fixed, data[last:]...)
}

// isObjectKey reports whether the JSON string that ends just before rest is an
// object key, that is, followed by a colon.
func isObjectKey(rest []byte) bool {
	for _, b := range rest {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case ':':
			return true
		default:
			return false
		}
	}
	return false
}

// DecodeResponse decodes the response body into the given target.
// Returns an APIError if the response indicates an error (status >= 400).
func (c *HTTPClient) DecodeResponse(resp *Response, target interface{}) error {
	// Check for error responses
	if resp.StatusCode >= 400 {
		return responseError(resp)
	}

	// Handle empty responses (204 No Content, etc.)
	if len(resp.Body) == 0 || target == nil {
		return nil
	}

	err := json.Unmarshal(resp.Body, target)
	if err == nil {
		return nil
	}

	// Responses have been RFC 3339 since the 2026-09-01 cutover. Older
	// environments and cached payloads may still send timestamps without a
	// timezone, which time.Time rejects; those are UTC. The body is only
	// rewritten when it did not decode as is, so a string value that merely
	// looks like such a timestamp normally reaches the caller unchanged.
	if fixed := fixTimestamps(resp.Body); !bytes.Equal(fixed, resp.Body) {
		resetTarget(target)
		if json.Unmarshal(fixed, target) == nil {
			return nil
		}
	}

	return fmt.Errorf("opusdns: failed to decode response: %w", err)
}

// resetTarget zeroes what target points to, discarding a partial decode.
func resetTarget(target interface{}) {
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Pointer && !v.IsNil() {
		v.Elem().Set(reflect.Zero(v.Elem().Type()))
	}
}

// responseError turns an error response into an APIError.
func responseError(resp *Response) *APIError {
	return NewAPIError(&http.Response{StatusCode: resp.StatusCode, Header: resp.Headers}, resp.Body)
}

// retryDelay returns how long to wait before retry number retry (1 for the
// first) of a request whose response carried headers. ok is false when the
// response's Retry-After asks for longer than RetryWaitMax.
func (c *HTTPClient) retryDelay(retry int, headers http.Header) (wait time.Duration, ok bool) {
	if retryAfter, found := parseRetryAfter(headers.Get("Retry-After")); found {
		if retryAfter > c.config.RetryWaitMax {
			return 0, false
		}
		return retryAfter, true
	}
	return c.calculateBackoff(retry), true
}

// parseRetryAfter reads a Retry-After header value, either delay-seconds or an
// HTTP date. A date in the past yields zero.
func parseRetryAfter(value string) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		if seconds < 0 {
			return 0, false
		}
		if seconds > int64(math.MaxInt64/time.Second) {
			return time.Duration(math.MaxInt64), true
		}
		return time.Duration(seconds) * time.Second, true
	}
	if t, err := http.ParseTime(value); err == nil {
		if d := time.Until(t); d > 0 {
			return d, true
		}
		return 0, true
	}
	return 0, false
}

// calculateBackoff calculates the backoff duration for retry number retry
// (1 for the first): RetryWaitMin doubled per retry, with ±20% jitter, never
// more than RetryWaitMax.
func (c *HTTPClient) calculateBackoff(retry int) time.Duration {
	maxWait := float64(c.config.RetryWaitMax)

	backoff := float64(c.config.RetryWaitMin) * math.Pow(2, float64(retry-1))
	if backoff > maxWait {
		backoff = maxWait
	}

	backoff += backoff * c.calculateJitterFactor()
	if backoff > maxWait {
		backoff = maxWait
	}
	if backoff < 0 {
		backoff = 0
	}

	return time.Duration(backoff)
}

func (c *HTTPClient) calculateJitterFactor() float64 {
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return 0
	}
	return (float64(n.Int64())/999_999 - 0.5) * 0.4
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
		// Standard error, so the log does not end up in a program's output.
		fmt.Fprintf(os.Stderr, "[opusdns] %s\n", msg)
	}
}

// redacted replaces secrets in debug logs.
const redacted = "[REDACTED]"

// sensitiveSuffixes mark the JSON keys and query parameters whose values debug
// logging masks, compared case-insensitively without "_" and "-": api_key,
// client_secret, access_token, refresh_token, auth_code, new_password, ...
var sensitiveSuffixes = []string{"apikey", "secret", "secretkey", "privatekey", "token", "password", "authcode", "authinfo"}

// isSensitiveKey reports whether a JSON key or query parameter holds a secret.
func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.NewReplacer("_", "", "-", "").Replace(key))
	for _, suffix := range sensitiveSuffixes {
		if strings.HasSuffix(normalized, suffix) {
			return true
		}
	}
	return false
}

// maxLoggedBody caps how much of a body that is not JSON debug logging prints.
const maxLoggedBody = 1024

// redactBody returns body for a debug log, with the values of sensitive keys
// masked. A body that is not JSON (a report archive, an error page from a
// proxy) is shortened, or only described when it is binary.
func redactBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var value interface{}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		if !utf8.Valid(body) || bytes.IndexByte(body, 0) >= 0 {
			return fmt.Sprintf("<%d bytes of binary data>", len(body))
		}
		if len(body) > maxLoggedBody {
			return fmt.Sprintf("%s... (%d bytes)", body[:maxLoggedBody], len(body))
		}
		return string(body)
	}

	if !redactValue(value) {
		return string(body)
	}
	masked, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("<%d bytes>", len(body))
	}
	return string(masked)
}

// redactValue masks the values of sensitive keys in a decoded JSON value and
// reports whether it masked any.
func redactValue(value interface{}) bool {
	changed := false
	switch v := value.(type) {
	case map[string]interface{}:
		for key, child := range v {
			if isSensitiveKey(key) && child != nil {
				v[key] = redacted
				changed = true
				continue
			}
			if redactValue(child) {
				changed = true
			}
		}
	case []interface{}:
		for _, child := range v {
			if redactValue(child) {
				changed = true
			}
		}
	}
	return changed
}

// redactURL returns rawURL for a debug log, with sensitive query parameters
// masked.
func redactURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.RawQuery == "" {
		return rawURL
	}
	query := u.Query()
	changed := false
	for key := range query {
		if isSensitiveKey(key) {
			query.Set(key, redacted)
			changed = true
		}
	}
	if !changed {
		return rawURL
	}
	u.RawQuery = query.Encode()
	return u.String()
}

// BuildPath constructs an API path with the configured version prefix.
//
// The parts are joined as they are. Escape every part that holds an ID or a
// name with url.PathEscape: Do refuses a path with an empty, "." or ".."
// segment, but cannot tell an unescaped "/" inside an ID from a separator.
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
