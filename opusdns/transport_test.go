package opusdns

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// problemUnauthorized is the 401 example of the spec's Problem schema.
const problemUnauthorized = `{"code":"ERROR_AUTHENTICATION","detail":"Additional error context.","status":401,"title":"Authentication Error","type":"authentication"}`

// problemInvalidPostalCode is the invalid_postal_code example of POST /v1/contacts (HTTPValidationError).
const problemInvalidPostalCode = `{
  "errors": [
    {
      "ctx": {"message": "'93053' is not a valid postal code for country 'NL', expected a format like '1234 AB'"},
      "input": "93053",
      "loc": ["body", "postal_code"],
      "msg": "'93053' is not a valid postal code for country 'NL', expected a format like '1234 AB'",
      "type": "invalid_postal_code"
    }
  ],
  "status": 422,
  "title": "Request validation error.",
  "type": "request-validation-failed"
}`

func newTestClient(t *testing.T, opts ...Option) *Client {
	t.Helper()
	client, err := NewClient(append([]Option{WithAPIKey("opk_test")}, opts...)...)
	require.NoError(t, err)
	return client
}

func TestRateLimit(t *testing.T) {
	t.Run("a 429 holds back only its own request", func(t *testing.T) {
		var dnsCalls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/v1/reports":
				// The report cooldown answers a repeat request within 5 minutes like this.
				w.Header().Set("Retry-After", "287")
				w.WriteHeader(http.StatusTooManyRequests)
			case "/v1/dns":
				dnsCalls.Add(1)
				_, _ = w.Write([]byte(`{"results":[],"pagination":{"has_next_page":false}}`))
			}
		}))
		defer server.Close()

		client := newTestClient(t, WithAPIEndpoint(server.URL), WithRetryWait(10*time.Millisecond, time.Second))

		start := time.Now()
		_, err := client.Reports.CreateReport(context.Background(), &models.CreateReportRequest{})
		assert.Less(t, time.Since(start), 500*time.Millisecond, "a Retry-After beyond RetryWaitMax is not waited out")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrRateLimited))
		apiErr, ok := IsAPIError(err)
		require.True(t, ok)
		assert.Equal(t, 287*time.Second, apiErr.RetryAfter)

		start = time.Now()
		_, err = client.DNS.ListZones(context.Background(), nil)
		require.NoError(t, err)
		assert.Less(t, time.Since(start), 500*time.Millisecond, "other requests are not held back")
		assert.EqualValues(t, 1, dnsCalls.Load())
	})

	t.Run("without Retry-After the wait starts at RetryWaitMin", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts == 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client := newTestClient(t, WithAPIEndpoint(server.URL), WithRetryWait(20*time.Millisecond, 10*time.Second))

		start := time.Now()
		_, err := client.http.Get(context.Background(), "/v1/dns", nil)
		require.NoError(t, err)
		assert.Equal(t, 2, attempts)
		assert.Less(t, time.Since(start), time.Second)
	})

	t.Run("MaxRetries 0 neither waits nor reports exhausted retries", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		client := newTestClient(t, WithAPIEndpoint(server.URL), WithMaxRetries(0))

		start := time.Now()
		resp, err := client.http.Get(context.Background(), "/v1/dns", nil)
		require.NoError(t, err)
		err = client.http.DecodeResponse(resp, nil)
		assert.Less(t, time.Since(start), 500*time.Millisecond)
		assert.Equal(t, 1, attempts)
		assert.True(t, errors.Is(err, ErrRateLimited))
		assert.NotContains(t, err.Error(), "retries")
	})

	t.Run("a context ending while waiting keeps the 429", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		client := newTestClient(t, WithAPIEndpoint(server.URL), WithRetryWait(5*time.Second, 10*time.Second))

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		_, err := client.http.Get(ctx, "/v1/dns", nil)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrRateLimited))
		assert.True(t, errors.Is(err, context.DeadlineExceeded))
		assert.True(t, errors.Is(err, ErrTimeout))
		_, ok := IsAPIError(err)
		assert.True(t, ok)
	})
}

func TestRetryDelay(t *testing.T) {
	client := newTestClient(t, WithRetryWait(time.Second, 30*time.Second))
	header := func(retryAfter string) http.Header {
		h := http.Header{}
		if retryAfter != "" {
			h.Set("Retry-After", retryAfter)
		}
		return h
	}

	wait, ok := client.http.retryDelay(1, header("2"))
	assert.True(t, ok)
	assert.Equal(t, 2*time.Second, wait, "Retry-After is waited as is, not added to the backoff")

	_, ok = client.http.retryDelay(1, header("31"))
	assert.False(t, ok, "a Retry-After beyond RetryWaitMax is not waited out")

	wait, ok = client.http.retryDelay(1, header(""))
	assert.True(t, ok)
	assert.InDelta(t, float64(time.Second), float64(wait), float64(200*time.Millisecond))
}

func TestParseRetryAfter(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  time.Duration
		ok    bool
	}{
		{"", 0, false},
		{"0", 0, true},
		{"120", 2 * time.Minute, true},
		{" 5 ", 5 * time.Second, true},
		{"-1", 0, false},
		{"soon", 0, false},
		{"99999999999999999", time.Duration(math.MaxInt64), true},
		{"Wed, 21 Oct 2015 07:28:00 GMT", 0, true},
	} {
		got, ok := parseRetryAfter(tc.value)
		assert.Equal(t, tc.ok, ok, tc.value)
		assert.Equal(t, tc.want, got, tc.value)
	}

	got, ok := parseRetryAfter(time.Now().Add(time.Minute).UTC().Format(http.TimeFormat))
	assert.True(t, ok)
	assert.InDelta(t, float64(time.Minute), float64(got), float64(2*time.Second))
}

func TestCalculateBackoff(t *testing.T) {
	client := newTestClient(t, WithRetryWait(time.Second, time.Second))
	for i := 0; i < 200; i++ {
		assert.LessOrEqual(t, client.http.calculateBackoff(5), time.Second, "jitter must not exceed RetryWaitMax")
	}

	client = newTestClient(t, WithRetryWait(100*time.Millisecond, time.Hour))
	for retry, want := range map[int]time.Duration{1: 100 * time.Millisecond, 2: 200 * time.Millisecond, 3: 400 * time.Millisecond} {
		got := client.http.calculateBackoff(retry)
		assert.InDelta(t, float64(want), float64(got), float64(want)*0.2+1, "retry %d", retry)
	}
	huge := newTestClient(t, WithRetryWait(time.Second, time.Hour)).http.calculateBackoff(1000)
	assert.LessOrEqual(t, huge, time.Hour, "no overflow")
	assert.GreaterOrEqual(t, huge, 48*time.Minute)
}

func TestNewAPIError(t *testing.T) {
	newErr := func(status int, header http.Header, body string) *APIError {
		if header == nil {
			header = http.Header{}
		}
		return NewAPIError(&http.Response{StatusCode: status, Header: header}, []byte(body))
	}

	t.Run("problem", func(t *testing.T) {
		header := http.Header{}
		header.Set("X-Server-Request-ID", "request-1234")
		err := newErr(http.StatusUnauthorized, header, problemUnauthorized)

		assert.Equal(t, "ERROR_AUTHENTICATION", err.ErrorCode)
		assert.Equal(t, "Additional error context.", err.Message)
		assert.Equal(t, "Authentication Error", err.Title)
		assert.Equal(t, "authentication", err.Type)
		assert.Equal(t, "request-1234", err.RequestID)
		assert.Equal(t, "opusdns: API error 401 [ERROR_AUTHENTICATION]: Additional error context. (request_id: request-1234)", err.Error())
		assert.True(t, errors.Is(err, ErrUnauthorized))
	})

	t.Run("validation error", func(t *testing.T) {
		err := newErr(http.StatusUnprocessableEntity, nil, problemInvalidPostalCode)

		require.Len(t, err.Errors, 1)
		assert.Equal(t, "body.postal_code", err.Errors[0].Field())
		assert.Equal(t, "invalid_postal_code", err.Errors[0].Type)
		assert.Equal(t, "Request validation error.", err.Title)
		assert.Equal(t, "request-validation-failed", err.Type)
		assert.Equal(t, "opusdns: API error 422: body.postal_code: '93053' is not a valid postal code for country 'NL', expected a format like '1234 AB'", err.Error())
		assert.True(t, errors.Is(err, ErrInvalidInput))
		assert.False(t, errors.Is(err, ErrBadRequest))
	})

	t.Run("FastAPI validation error", func(t *testing.T) {
		err := newErr(http.StatusUnprocessableEntity, nil, `{"detail":[{"loc":["query","type"],"msg":"Field required","type":"missing"}]}`)

		require.Len(t, err.Errors, 1)
		assert.Equal(t, "query.type: Field required", err.Message)
	})

	t.Run("older field names", func(t *testing.T) {
		header := http.Header{}
		header.Set("X-Request-ID", "legacy-1")
		err := newErr(http.StatusConflict, header, `{"error_code":"zone_exists","message":"zone exists"}`)

		assert.Equal(t, "zone_exists", err.ErrorCode)
		assert.Equal(t, "zone exists", err.Message)
		assert.Equal(t, "legacy-1", err.RequestID)
	})

	t.Run("a field of an unexpected type loses only itself", func(t *testing.T) {
		err := newErr(http.StatusBadRequest, nil, `{"code":7,"detail":"name is required","errors":"none","title":"Bad Request"}`)

		assert.Equal(t, "", err.ErrorCode)
		assert.Equal(t, "name is required", err.Message)
		assert.Equal(t, "Bad Request", err.Title)
		assert.Empty(t, err.Errors)
	})

	t.Run("title when there is no detail", func(t *testing.T) {
		err := newErr(http.StatusForbidden, nil, `{"title":"Forbidden","type":"forbidden","status":403}`)
		assert.Equal(t, "Forbidden", err.Message)
	})

	t.Run("Retry-After", func(t *testing.T) {
		header := http.Header{}
		header.Set("Retry-After", "60")
		err := newErr(http.StatusTooManyRequests, header, "")
		assert.Equal(t, time.Minute, err.RetryAfter)
	})
}

func TestPathSafety(t *testing.T) {
	var (
		mu       sync.Mutex
		received []string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		received = append(received, r.Method+" "+r.URL.EscapedPath())
		mu.Unlock()
		if r.URL.Path == "/v1/contacts/" {
			// FastAPI's redirect_slashes.
			http.Redirect(w, r, "/v1/contacts", http.StatusTemporaryRedirect)
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	reset := func() {
		mu.Lock()
		received = nil
		mu.Unlock()
	}
	requests := func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), received...)
	}

	client := newTestClient(t, WithAPIEndpoint(server.URL))
	ctx := context.Background()

	for name, path := range map[string]string{
		"dot-dot":         client.http.BuildPath("contacts", "..", "domains", "example.com"),
		"escaped dot-dot": client.http.BuildPath("contacts", "%2E%2E", "domains", "example.com"),
		"dot":             client.http.BuildPath("contacts", ".", "x"),
		"empty segment":   client.http.BuildPath("contacts", "", "verification"),
	} {
		t.Run("rejects "+name, func(t *testing.T) {
			reset()
			_, err := client.http.Delete(ctx, path)
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrInvalidInput), err.Error())
			assert.Empty(t, requests(), "nothing is sent")
		})
	}

	t.Run("an escaped slash stays escaped", func(t *testing.T) {
		reset()
		_, err := client.http.Get(ctx, client.http.BuildPath("contacts", "a%2Fb"), nil)
		require.NoError(t, err)
		assert.Equal(t, []string{"GET /v1/contacts/a%2Fb"}, requests())
	})

	t.Run("a trailing slash a route needs is kept", func(t *testing.T) {
		reset()
		_, err := client.http.Get(ctx, client.http.BuildPath("tlds", ""), nil)
		require.NoError(t, err)
		assert.Equal(t, []string{"GET /v1/tlds/"}, requests())
	})

	t.Run("an empty last ID is not redirected to the collection", func(t *testing.T) {
		reset()
		_, err := client.Contacts.GetContact(ctx, "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidInput), err.Error())
		assert.Equal(t, []string{"GET /v1/contacts/"}, requests(), "the redirect is not followed")
	})

	t.Run("endpoint with a base path", func(t *testing.T) {
		reset()
		based := newTestClient(t, WithAPIEndpoint(server.URL+"/api/"))
		_, err := based.http.Get(ctx, based.http.BuildPath("dns"), nil)
		require.NoError(t, err)
		assert.Equal(t, []string{"GET /api/v1/dns"}, requests())
	})
}

func TestRequestsThatCannotSucceedAreNotRetried(t *testing.T) {
	for name, tc := range map[string]struct {
		req      *Request
		transErr error
	}{
		"body that cannot be encoded": {req: &Request{Method: http.MethodPut, Path: "/v1/dns/example.com", Body: math.NaN()}},
		"header with a newline":       {req: &Request{Method: http.MethodGet, Path: "/v1/dns", Headers: http.Header{"X-Trace": {"a\nb"}}}},
		"invalid header name":         {req: &Request{Method: http.MethodGet, Path: "/v1/dns", Headers: http.Header{"X Trace": {"a"}}}},
		"untrusted certificate":       {req: &Request{Method: http.MethodGet, Path: "/v1/dns"}, transErr: &tls.CertificateVerificationError{Err: errors.New("x509: unknown authority")}},
	} {
		t.Run(name, func(t *testing.T) {
			attempts := 0
			transport := roundTripFunc(func(*http.Request) (*http.Response, error) {
				attempts++
				if tc.transErr != nil {
					return nil, tc.transErr
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
			})
			client := newTestClient(t,
				WithAPIEndpoint("https://api.example.test"),
				WithHTTPClient(&http.Client{Transport: transport}),
				WithRetryWait(time.Millisecond, 2*time.Millisecond),
			)

			_, err := client.http.Do(context.Background(), tc.req)
			require.Error(t, err)
			if tc.transErr != nil {
				assert.Equal(t, 1, attempts)
			} else {
				assert.Equal(t, 0, attempts)
			}
		})
	}

	t.Run("exhausted retries say how many attempts were made", func(t *testing.T) {
		transport := roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("connection reset")
		})
		client := newTestClient(t,
			WithAPIEndpoint("https://api.example.test"),
			WithHTTPClient(&http.Client{Transport: transport}),
			WithRetryWait(time.Millisecond, 2*time.Millisecond),
		)
		_, err := client.http.Get(context.Background(), "/v1/dns", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "giving up after 4 attempts")
	})
}

func TestConfigValidation(t *testing.T) {
	for name, tc := range map[string]struct {
		opts  []Option
		field string
	}{
		"endpoint without scheme":     {[]Option{WithAPIKey("opk_test"), WithAPIEndpoint("api.opusdns.com")}, "APIEndpoint"},
		"endpoint with other scheme":  {[]Option{WithAPIKey("opk_test"), WithAPIEndpoint("ftp://api.opusdns.com")}, "APIEndpoint"},
		"endpoint without host":       {[]Option{WithAPIKey("opk_test"), WithAPIEndpoint("https://")}, "APIEndpoint"},
		"API key with a newline":      {[]Option{WithAPIKey("opk_test\n")}, "APIKey"},
		"client token with a newline": {[]Option{WithAPIKey("opk_test"), WithClientToken("acme/1\r\n")}, "ClientToken"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewClient(tc.opts...)
			var configErr *ConfigError
			require.ErrorAs(t, err, &configErr)
			assert.Equal(t, tc.field, configErr.Field)
			assert.NotContains(t, err.Error(), "opk_test", "the key is not echoed")
		})
	}

	for _, endpoint := range []string{"https://api.opusdns.com", "http://127.0.0.1:8080", "https://sandbox.opusdns.com/"} {
		_, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(endpoint))
		assert.NoError(t, err, endpoint)
	}
}

func TestRedirects(t *testing.T) {
	var otherKey atomic.Value
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		otherKey.Store(r.Header.Get("X-Api-Key") + "|" + r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{}`))
	}))
	defer other.Close()

	var sameKey atomic.Value
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/elsewhere":
			http.Redirect(w, r, other.URL+"/landing", http.StatusFound)
		case "/v1/moved":
			http.Redirect(w, r, "/v1/here", http.StatusFound)
		case "/v1/here":
			sameKey.Store(r.Header.Get("X-Api-Key"))
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer origin.Close()

	headers := http.Header{"Authorization": {"Bearer token"}}

	t.Run("default client", func(t *testing.T) {
		client := newTestClient(t, WithAPIEndpoint(origin.URL))

		_, err := client.http.Do(context.Background(), &Request{Method: http.MethodGet, Path: "/v1/elsewhere", Headers: headers})
		require.NoError(t, err)
		assert.Equal(t, "|", otherKey.Load(), "no credentials reach another host")

		_, err = client.http.Get(context.Background(), "/v1/moved", nil)
		require.NoError(t, err)
		assert.Equal(t, "opk_test", sameKey.Load(), "credentials stay on a redirect within the API")
	})

	t.Run("custom client", func(t *testing.T) {
		custom := &http.Client{}
		client := newTestClient(t, WithAPIEndpoint(origin.URL), WithHTTPClient(custom))

		otherKey.Store("unset")
		_, err := client.http.Do(context.Background(), &Request{Method: http.MethodGet, Path: "/v1/elsewhere", Headers: headers})
		require.NoError(t, err)
		assert.Equal(t, "|", otherKey.Load())
		assert.Nil(t, custom.CheckRedirect, "the caller's client is not changed")
	})
}

func TestDefaultTransportHonoursProxyEnvironment(t *testing.T) {
	client := newTestClient(t)
	transport, ok := client.http.httpClient.Transport.(*http.Transport)
	require.True(t, ok)
	assert.NotNil(t, transport.Proxy)
}

func TestTimeouts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	t.Run("HTTP client timeout", func(t *testing.T) {
		client := newTestClient(t, WithAPIEndpoint(server.URL), WithHTTPTimeout(20*time.Millisecond), WithMaxRetries(0))
		_, err := client.http.Get(context.Background(), "/v1/dns", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrTimeout), err.Error())
	})

	t.Run("context deadline", func(t *testing.T) {
		client := newTestClient(t, WithAPIEndpoint(server.URL))
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		_, err := client.http.Get(ctx, "/v1/dns", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrTimeout), err.Error())
		assert.True(t, errors.Is(err, context.DeadlineExceeded))
	})

	t.Run("context canceled is no timeout", func(t *testing.T) {
		client := newTestClient(t, WithAPIEndpoint(server.URL))
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := client.http.Get(ctx, "/v1/dns", nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
		assert.False(t, errors.Is(err, ErrTimeout))
	})
}

func TestDecodeResponseTimestamps(t *testing.T) {
	type record struct {
		CreatedAt time.Time         `json:"created_at"`
		Label     string            `json:"label"`
		Labels    map[string]string `json:"labels"`
	}
	client := newTestClient(t)
	decode := func(body string) (record, error) {
		var r record
		err := client.http.DecodeResponse(&Response{StatusCode: http.StatusOK, Body: []byte(body)}, &r)
		return r, err
	}

	t.Run("RFC 3339 payload is decoded unchanged", func(t *testing.T) {
		r, err := decode(`{"created_at":"2026-01-23T08:26:55Z","label":"2026-01-23T08:26:55"}`)
		require.NoError(t, err)
		assert.Equal(t, "2026-01-23T08:26:55", r.Label)
	})

	t.Run("timestamp without zone, with fractional seconds", func(t *testing.T) {
		r, err := decode(`{"created_at":"2026-01-23T08:26:55.123456"}`)
		require.NoError(t, err)
		assert.Equal(t, time.Date(2026, 1, 23, 8, 26, 55, 123456000, time.UTC), r.CreatedAt)
	})

	t.Run("map keys and escaped quotes are left alone", func(t *testing.T) {
		r, err := decode(`{"created_at":"2026-01-23T08:26:55","labels":{"2026-01-23T08:26:55":"x \"2026-01-23T08:26:55\" y"}}`)
		require.NoError(t, err)
		assert.Equal(t, time.Date(2026, 1, 23, 8, 26, 55, 0, time.UTC), r.CreatedAt)
		assert.Equal(t, map[string]string{"2026-01-23T08:26:55": `x "2026-01-23T08:26:55" y`}, r.Labels)
	})

	t.Run("other decode errors are reported", func(t *testing.T) {
		_, err := decode(`{"created_at":"yesterday"}`)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode response")
	})
}

type capturingLogger struct {
	mu    sync.Mutex
	lines []string
}

func (l *capturingLogger) Printf(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, fmt.Sprintf(format, v...))
}

func (l *capturingLogger) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return strings.Join(l.lines, "\n")
}

func TestDebugLogging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"eyJhbGciOi.secret-token","token_type":"bearer","expires_in":3600}`))
	}))
	defer server.Close()

	t.Run("secrets are masked", func(t *testing.T) {
		logger := &capturingLogger{}
		client := newTestClient(t, WithAPIEndpoint(server.URL), WithDebug(true), WithLogger(logger), WithAPIKey("opk_live_key"))

		_, err := client.Auth.IssueToken(context.Background(), "organization_01h45ytscbebyvny4gc8cr8ma2", "client-s3cr3t")
		require.NoError(t, err)
		_, err = client.http.Get(context.Background(), "/v1/domains/example.com", map[string][]string{"auth_code": {"Xy7!abc"}})
		require.NoError(t, err)

		logged := logger.String()
		for _, secret := range []string{"client-s3cr3t", "secret-token", "opk_live_key", "Xy7!abc"} {
			assert.NotContains(t, logged, secret)
		}
		assert.Contains(t, logged, redacted)
		assert.Contains(t, logged, "bearer", "other fields are still logged")
	})

	t.Run("without a logger, to standard error", func(t *testing.T) {
		stdout, stderr := captureOutput(t, func() {
			client := newTestClient(t, WithAPIEndpoint(server.URL), WithDebug(true))
			_, err := client.http.Get(context.Background(), "/v1/dns", nil)
			require.NoError(t, err)
		})
		assert.Empty(t, stdout)
		assert.Contains(t, stderr, "[opusdns] GET ")
	})

	t.Run("binary bodies are described, not printed", func(t *testing.T) {
		assert.Equal(t, "<6 bytes of binary data>", redactBody([]byte("PK\x03\x04\x00\x01")))
	})
}

// captureOutput runs fn with os.Stdout and os.Stderr redirected and returns
// what was written to each.
func captureOutput(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()
	capture := func(target **os.File) (restore func() string) {
		r, w, err := os.Pipe()
		require.NoError(t, err)
		original := *target
		*target = w
		done := make(chan string)
		go func() {
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			done <- buf.String()
		}()
		return func() string {
			*target = original
			_ = w.Close()
			return <-done
		}
	}
	restoreStdout := capture(&os.Stdout)
	restoreStderr := capture(&os.Stderr)
	fn()
	return restoreStdout(), restoreStderr()
}

// TestDownloadReportStreams covers group A's change to the report download:
// the body reaches the writer while it is still arriving, a slow transfer is
// not cut off by the HTTP timeout, and errors keep their classification.
func TestDownloadReportStreams(t *testing.T) {
	t.Run("streams past the HTTP timeout", func(t *testing.T) {
		firstChunkSeen := make(chan struct{})
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/reports/report_01h45ytscbebyvny4gc8cr8ma2/download", r.URL.EscapedPath())
			w.Header().Set("Content-Type", "application/zip")
			_, _ = w.Write([]byte("PK\x03\x04"))
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			select {
			case <-firstChunkSeen:
			case <-time.After(2 * time.Second):
				t.Error("the first chunk did not reach the writer before the body ended")
			}
			time.Sleep(150 * time.Millisecond) // longer than the HTTP timeout
			_, _ = w.Write([]byte(" rest"))
		}))
		defer server.Close()

		client := newTestClient(t, WithAPIEndpoint(server.URL), WithHTTPTimeout(100*time.Millisecond))

		w := &signallingWriter{signal: firstChunkSeen}
		err := client.Reports.DownloadReportToWriter(context.Background(), "report_01h45ytscbebyvny4gc8cr8ma2", w)
		require.NoError(t, err)
		assert.Equal(t, "PK\x03\x04 rest", w.buf.String())
	})

	t.Run("not ready", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"code":"ERROR_REPORT_NOT_READY","detail":"Report is not ready for download (status: PENDING)","status":409,"title":"Conflict","type":"conflict"}`))
		}))
		defer server.Close()

		client := newTestClient(t, WithAPIEndpoint(server.URL))

		var buf bytes.Buffer
		err := client.Reports.DownloadReportToWriter(context.Background(), "report_01h45ytscbebyvny4gc8cr8ma2", &buf)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrConflict))
		assert.Contains(t, err.Error(), "not ready")
		assert.Zero(t, buf.Len(), "the error body is not written to the report")
		assert.Equal(t, 1, attempts)

		data, err := client.Reports.DownloadReport(context.Background(), "report_01h45ytscbebyvny4gc8cr8ma2")
		assert.Nil(t, data)
		assert.True(t, errors.Is(err, ErrConflict))
	})

	t.Run("report ID is escaped", func(t *testing.T) {
		client := newTestClient(t, WithAPIEndpoint("https://api.example.test"), WithHTTPClient(&http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				assert.Equal(t, "/v1/reports/a%2F..%2Fb/download", r.URL.EscapedPath())
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("zip"))}, nil
			}),
		}))
		data, err := client.Reports.DownloadReport(context.Background(), "a/../b")
		require.NoError(t, err)
		assert.Equal(t, []byte("zip"), data)
	})
}

// signallingWriter closes signal on its first write.
type signallingWriter struct {
	buf    bytes.Buffer
	signal chan struct{}
	once   sync.Once
}

func (w *signallingWriter) Write(p []byte) (int, error) {
	w.once.Do(func() { close(w.signal) })
	return w.buf.Write(p)
}
