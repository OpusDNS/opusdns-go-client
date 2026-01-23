package opusdns

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("applies defaults", func(t *testing.T) {
		client := NewClient(&Config{APIKey: "opk_test"})

		assert.Equal(t, "opk_test", client.config.APIKey)
		assert.Equal(t, DefaultAPIEndpoint, client.config.APIEndpoint)
		assert.Equal(t, DefaultTTL, client.config.TTL)
		assert.Equal(t, DefaultTimeout, client.config.HTTPTimeout)
		assert.Equal(t, DefaultMaxRetries, client.config.MaxRetries)
	})

	t.Run("uses custom values", func(t *testing.T) {
		client := NewClient(&Config{
			APIKey:      "opk_custom",
			APIEndpoint: "https://custom.api",
			TTL:         300,
			HTTPTimeout: 60 * time.Second,
			MaxRetries:  5,
		})

		assert.Equal(t, "opk_custom", client.config.APIKey)
		assert.Equal(t, "https://custom.api", client.config.APIEndpoint)
		assert.Equal(t, 300, client.config.TTL)
		assert.Equal(t, 60*time.Second, client.config.HTTPTimeout)
		assert.Equal(t, 5, client.config.MaxRetries)
	})
}

func TestListZones(t *testing.T) {
	t.Run("returns zones", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "opk_test", r.Header.Get("X-Api-Key"))
			assert.Contains(t, r.URL.Path, "/v1/dns")

			_ = json.NewEncoder(w).Encode(zoneListResponse{
				Results: []Zone{
					{Name: "example.com", DNSSECStatus: "disabled"},
					{Name: "test.com", DNSSECStatus: "enabled"},
				},
				Pagination: Pagination{HasNextPage: false},
			})
		}))
		defer server.Close()

		client := NewClient(&Config{APIKey: "opk_test", APIEndpoint: server.URL})
		zones, err := client.ListZones()

		require.NoError(t, err)
		assert.Len(t, zones, 2)
		assert.Equal(t, "example.com", zones[0].Name)
	})

	t.Run("handles pagination", func(t *testing.T) {
		page := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			page++
			if page == 1 {
				_ = json.NewEncoder(w).Encode(zoneListResponse{
					Results:    []Zone{{Name: "zone1.com"}, {Name: "zone2.com"}},
					Pagination: Pagination{HasNextPage: true, CurrentPage: 1},
				})
			} else {
				_ = json.NewEncoder(w).Encode(zoneListResponse{
					Results:    []Zone{{Name: "zone3.com"}},
					Pagination: Pagination{HasNextPage: false, CurrentPage: 2},
				})
			}
		}))
		defer server.Close()

		client := NewClient(&Config{APIKey: "opk_test", APIEndpoint: server.URL})
		zones, err := client.ListZones()

		require.NoError(t, err)
		assert.Len(t, zones, 3)
	})

	t.Run("returns error on unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "invalid API key"})
		}))
		defer server.Close()

		client := NewClient(&Config{APIKey: "bad_key", APIEndpoint: server.URL})
		_, err := client.ListZones()

		require.Error(t, err)
		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, 401, apiErr.StatusCode)
	})
}

func TestGetZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/dns/example.com", r.URL.Path)
		_ = json.NewEncoder(w).Encode(Zone{
			Name:         "example.com",
			DNSSECStatus: "disabled",
		})
	}))
	defer server.Close()

	client := NewClient(&Config{APIKey: "opk_test", APIEndpoint: server.URL})
	zone, err := client.GetZone("example.com.")

	require.NoError(t, err)
	assert.Equal(t, "example.com", zone.Name)
}

func TestCreateZone(t *testing.T) {
	t.Run("creates empty zone", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/dns", r.URL.Path)

			var req zoneCreateRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "newzone.com", req.Name)
			assert.Empty(t, req.RRSets)

			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(Zone{Name: "newzone.com"})
		}))
		defer server.Close()

		client := NewClient(&Config{APIKey: "opk_test", APIEndpoint: server.URL})
		zone, err := client.CreateZone("newzone.com", nil)

		require.NoError(t, err)
		assert.Equal(t, "newzone.com", zone.Name)
	})

	t.Run("creates zone with records", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req zoneCreateRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.NotEmpty(t, req.RRSets)

			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(Zone{Name: "newzone.com"})
		}))
		defer server.Close()

		client := NewClient(&Config{APIKey: "opk_test", APIEndpoint: server.URL})
		_, err := client.CreateZone("newzone.com", []Record{
			{Name: "www", Type: "A", TTL: 3600, RData: "1.2.3.4"},
		})

		require.NoError(t, err)
	})
}

func TestDeleteZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/dns/example.com", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(&Config{APIKey: "opk_test", APIEndpoint: server.URL})
	err := client.DeleteZone("example.com")

	require.NoError(t, err)
}

func TestDNSSEC(t *testing.T) {
	t.Run("enable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/dns/example.com/dnssec/enable", r.URL.Path)
			_ = json.NewEncoder(w).Encode(DNSChanges{NumChanges: 5})
		}))
		defer server.Close()

		client := NewClient(&Config{APIKey: "opk_test", APIEndpoint: server.URL})
		changes, err := client.EnableDNSSEC("example.com")

		require.NoError(t, err)
		assert.Equal(t, 5, changes.NumChanges)
	})

	t.Run("disable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/dns/example.com/dnssec/disable", r.URL.Path)
			_ = json.NewEncoder(w).Encode(DNSChanges{NumChanges: 3})
		}))
		defer server.Close()

		client := NewClient(&Config{APIKey: "opk_test", APIEndpoint: server.URL})
		changes, err := client.DisableDNSSEC("example.com")

		require.NoError(t, err)
		assert.Equal(t, 3, changes.NumChanges)
	})
}

func TestGetRecords(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/dns/example.com/rrsets", r.URL.Path)
		_ = json.NewEncoder(w).Encode([]RRSet{
			{Name: "www", Type: "A", TTL: 3600},
			{Name: "mail", Type: "MX", TTL: 3600},
		})
	}))
	defer server.Close()

	client := NewClient(&Config{APIKey: "opk_test", APIEndpoint: server.URL})
	rrsets, err := client.GetRecords("example.com")

	require.NoError(t, err)
	assert.Len(t, rrsets, 2)
}

func TestUpsertRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/dns/example.com/records", r.URL.Path)

		var req recordPatchRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		require.Len(t, req.Ops, 1)
		assert.Equal(t, "upsert", req.Ops[0].Op)
		assert.Equal(t, "www", req.Ops[0].Record.Name)
		assert.Equal(t, "A", req.Ops[0].Record.Type)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(&Config{APIKey: "opk_test", APIEndpoint: server.URL})
	err := client.UpsertRecord("example.com", Record{
		Name:  "www",
		Type:  "A",
		TTL:   3600,
		RData: "1.2.3.4",
	})

	require.NoError(t, err)
}

func TestDeleteRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req recordPatchRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		require.Len(t, req.Ops, 1)
		assert.Equal(t, "remove", req.Ops[0].Op)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(&Config{APIKey: "opk_test", APIEndpoint: server.URL})
	err := client.DeleteRecord("example.com", Record{
		Name:  "www",
		Type:  "A",
		TTL:   3600,
		RData: "1.2.3.4",
	})

	require.NoError(t, err)
}

func TestRetryLogic(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_ = json.NewEncoder(w).Encode(zoneListResponse{
			Results:    []Zone{{Name: "example.com"}},
			Pagination: Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client := NewClient(&Config{
		APIKey:      "opk_test",
		APIEndpoint: server.URL,
		MaxRetries:  3,
	})

	zones, err := client.ListZones()

	require.NoError(t, err)
	assert.Len(t, zones, 1)
	assert.Equal(t, 3, attempts)
}

func TestAPIError(t *testing.T) {
	err := &APIError{StatusCode: 404, Message: "zone not found"}
	assert.Equal(t, "opusdns: API error 404: zone not found", err.Error())

	err2 := &APIError{StatusCode: 500}
	assert.Equal(t, "opusdns: API error 500", err2.Error())
}
