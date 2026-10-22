package opusdns

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("creates client with API key", func(t *testing.T) {
		client, err := NewClient(WithAPIKey("opk_test"))

		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "opk_test", client.Config.APIKey)
		assert.Equal(t, DefaultAPIEndpoint, client.Config.APIEndpoint)
		assert.Equal(t, DefaultTTL, client.Config.TTL)
		assert.Equal(t, DefaultTimeout, client.Config.HTTPTimeout)
		assert.Equal(t, DefaultMaxRetries, client.Config.MaxRetries)
	})

	t.Run("applies custom configuration", func(t *testing.T) {
		client, err := NewClient(
			WithAPIKey("opk_custom"),
			WithAPIEndpoint("https://custom.api"),
			WithTTL(300),
			WithHTTPTimeout(60*time.Second),
			WithMaxRetries(5),
		)

		require.NoError(t, err)
		assert.Equal(t, "opk_custom", client.Config.APIKey)
		assert.Equal(t, "https://custom.api", client.Config.APIEndpoint)
		assert.Equal(t, 300, client.Config.TTL)
		assert.Equal(t, 60*time.Second, client.Config.HTTPTimeout)
		assert.Equal(t, 5, client.Config.MaxRetries)
	})

	t.Run("returns error without API key", func(t *testing.T) {
		_, err := NewClient()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "API key is required")
	})

	t.Run("initializes all services", func(t *testing.T) {
		client, err := NewClient(WithAPIKey("opk_test"))

		require.NoError(t, err)
		assert.NotNil(t, client.DNS)
		assert.NotNil(t, client.Domains)
		assert.NotNil(t, client.Contacts)
		assert.NotNil(t, client.EmailForwards)
		assert.NotNil(t, client.DomainForwards)
		assert.NotNil(t, client.TLDs)
		assert.NotNil(t, client.Availability)
		assert.NotNil(t, client.Organizations)
		assert.NotNil(t, client.Users)
		assert.NotNil(t, client.Events)
		assert.NotNil(t, client.Tags)
	})
}

func TestRetryLogic(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_ = json.NewEncoder(w).Encode(models.ZoneListResponse{
			Results:    []models.Zone{{Name: "example.com"}},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(
		WithAPIKey("opk_test"),
		WithAPIEndpoint(server.URL),
		WithMaxRetries(3),
		WithRetryWait(10*time.Millisecond, 50*time.Millisecond),
	)
	require.NoError(t, err)

	zones, err := client.DNS.ListZones(context.Background(), nil)

	require.NoError(t, err)
	assert.Len(t, zones, 1)
	assert.Equal(t, 3, attempts)
}

// roundTripFunc lets a test stand in for the network.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRetryPolicy(t *testing.T) {
	newClient := func(t *testing.T, opts ...Option) *Client {
		t.Helper()
		opts = append([]Option{
			WithAPIKey("opk_test"),
			WithMaxRetries(3),
			WithRetryWait(time.Millisecond, 2*time.Millisecond),
		}, opts...)
		client, err := NewClient(opts...)
		require.NoError(t, err)
		return client
	}

	t.Run("server errors", func(t *testing.T) {
		for _, tc := range []struct {
			method   string
			attempts int
		}{
			{http.MethodGet, 4},
			{http.MethodPut, 4},
			{http.MethodDelete, 1},
			{http.MethodPost, 1},
			{http.MethodPatch, 1},
		} {
			t.Run(tc.method, func(t *testing.T) {
				attempts := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					attempts++
					w.WriteHeader(http.StatusServiceUnavailable)
				}))
				defer server.Close()

				client := newClient(t, WithAPIEndpoint(server.URL))
				resp, err := client.http.Do(context.Background(), &Request{Method: tc.method, Path: "/v1/domains/example.com/renew"})
				if err == nil {
					err = client.http.DecodeResponse(resp, nil)
				}

				assert.Equal(t, tc.attempts, attempts)
				var apiErr *APIError
				require.ErrorAs(t, err, &apiErr)
				assert.Equal(t, http.StatusServiceUnavailable, apiErr.StatusCode)
			})
		}
	})

	t.Run("rate limit retries POST", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts == 1 {
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client := newClient(t, WithAPIEndpoint(server.URL))
		resp, err := client.http.Post(context.Background(), "/v1/domains", nil)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 2, attempts)
	})

	t.Run("network errors", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			method   string
			op       string
			attempts int
		}{
			{"GET after the request was sent", http.MethodGet, "read", 4},
			{"POST after the request was sent", http.MethodPost, "read", 1},
			{"DELETE after the request was sent", http.MethodDelete, "read", 1},
			{"POST before the connection opened", http.MethodPost, "dial", 4},
			{"DELETE before the connection opened", http.MethodDelete, "dial", 4},
		} {
			t.Run(tc.name, func(t *testing.T) {
				attempts := 0
				transport := roundTripFunc(func(*http.Request) (*http.Response, error) {
					attempts++
					return nil, &net.OpError{Op: tc.op, Net: "tcp", Err: errors.New("connection reset")}
				})

				client := newClient(t,
					WithAPIEndpoint("https://api.example.test"),
					WithHTTPClient(&http.Client{Transport: transport}),
				)
				_, err := client.http.Do(context.Background(), &Request{Method: tc.method, Path: "/v1/domains"})

				require.Error(t, err)
				assert.Equal(t, tc.attempts, attempts)
			})
		}
	})
}

func TestAPIError(t *testing.T) {
	t.Run("error message formatting", func(t *testing.T) {
		err := &APIError{StatusCode: 404, Message: "zone not found"}
		assert.Equal(t, "opusdns: API error 404: zone not found", err.Error())

		err2 := &APIError{StatusCode: 500}
		assert.Equal(t, "opusdns: API error 500", err2.Error())

		err3 := &APIError{StatusCode: 400, ErrorCode: "invalid_input", Message: "name is required"}
		assert.Equal(t, "opusdns: API error 400 [invalid_input]: name is required", err3.Error())
	})

	t.Run("error type checking", func(t *testing.T) {
		err := &APIError{StatusCode: 404}
		assert.True(t, err.Is(ErrNotFound))
		assert.False(t, err.Is(ErrUnauthorized))

		err = &APIError{StatusCode: 401}
		assert.True(t, err.Is(ErrUnauthorized))

		err = &APIError{StatusCode: 429}
		assert.True(t, err.Is(ErrRateLimited))
		assert.True(t, err.IsRetryable())

		err = &APIError{StatusCode: 500}
		assert.True(t, err.IsServerError())
		assert.True(t, err.IsRetryable())
	})
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(models.ZoneListResponse{
			Results:    []models.Zone{{Name: "example.com"}},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = client.DNS.ListZones(ctx, nil)

	require.Error(t, err)
}
