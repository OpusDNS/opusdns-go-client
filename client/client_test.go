package client

import (
	"context"
	"encoding/json"
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
	})
}

func TestDNSService_ListZones(t *testing.T) {
	t.Run("returns zones", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "opk_test", r.Header.Get("X-Api-Key"))
			assert.Contains(t, r.URL.Path, "/v1/dns")

			_ = json.NewEncoder(w).Encode(models.ZoneListResponse{
				Results: []models.Zone{
					{Name: "example.com", DNSSECStatus: models.DNSSECStatusDisabled},
					{Name: "test.com", DNSSECStatus: models.DNSSECStatusEnabled},
				},
				Pagination: models.Pagination{HasNextPage: false},
			})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		zones, err := client.DNS.ListZones(context.Background(), nil)

		require.NoError(t, err)
		assert.Len(t, zones, 2)
		assert.Equal(t, "example.com", zones[0].Name)
	})

	t.Run("handles pagination", func(t *testing.T) {
		page := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			page++
			if page == 1 {
				_ = json.NewEncoder(w).Encode(models.ZoneListResponse{
					Results:    []models.Zone{{Name: "zone1.com"}, {Name: "zone2.com"}},
					Pagination: models.Pagination{HasNextPage: true, CurrentPage: 1},
				})
			} else {
				_ = json.NewEncoder(w).Encode(models.ZoneListResponse{
					Results:    []models.Zone{{Name: "zone3.com"}},
					Pagination: models.Pagination{HasNextPage: false, CurrentPage: 2},
				})
			}
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		zones, err := client.DNS.ListZones(context.Background(), nil)

		require.NoError(t, err)
		assert.Len(t, zones, 3)
	})

	t.Run("returns error on unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "invalid API key"})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("bad_key"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		_, err = client.DNS.ListZones(context.Background(), nil)

		require.Error(t, err)
		assert.True(t, IsUnauthorizedError(err))
	})
}

func TestDNSService_GetZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/dns/example.com", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.Zone{
			Name:         "example.com",
			DNSSECStatus: models.DNSSECStatusDisabled,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	zone, err := client.DNS.GetZone(context.Background(), "example.com.")

	require.NoError(t, err)
	assert.Equal(t, "example.com", zone.Name)
}

func TestDNSService_CreateZone(t *testing.T) {
	t.Run("creates empty zone", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/dns", r.URL.Path)

			var req models.ZoneCreateRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "newzone.com", req.Name)
			assert.Empty(t, req.RRSets)

			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(models.Zone{Name: "newzone.com"})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		zone, err := client.DNS.CreateZone(context.Background(), &models.ZoneCreateRequest{
			Name: "newzone.com",
		})

		require.NoError(t, err)
		assert.Equal(t, "newzone.com", zone.Name)
	})

	t.Run("creates zone with records", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req models.ZoneCreateRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.NotEmpty(t, req.RRSets)

			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(models.Zone{Name: "newzone.com"})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		_, err = client.DNS.CreateZone(context.Background(), &models.ZoneCreateRequest{
			Name: "newzone.com",
			RRSets: []models.RRSetCreate{
				{Name: "www", Type: models.RRSetTypeA, TTL: 3600, Records: []string{"1.2.3.4"}},
			},
		})

		require.NoError(t, err)
	})
}

func TestDNSService_DeleteZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/dns/example.com", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.DNS.DeleteZone(context.Background(), "example.com")

	require.NoError(t, err)
}

func TestDNSService_DNSSEC(t *testing.T) {
	t.Run("enable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/dns/example.com/dnssec/enable", r.URL.Path)
			_ = json.NewEncoder(w).Encode(models.DNSChanges{NumChanges: 5})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		changes, err := client.DNS.EnableDNSSEC(context.Background(), "example.com")

		require.NoError(t, err)
		assert.Equal(t, 5, changes.NumChanges)
	})

	t.Run("disable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/dns/example.com/dnssec/disable", r.URL.Path)
			_ = json.NewEncoder(w).Encode(models.DNSChanges{NumChanges: 3})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		changes, err := client.DNS.DisableDNSSEC(context.Background(), "example.com")

		require.NoError(t, err)
		assert.Equal(t, 3, changes.NumChanges)
	})
}

func TestDNSService_GetRRSets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/dns/example.com/rrsets", r.URL.Path)
		_ = json.NewEncoder(w).Encode([]models.RRSet{
			{Name: "www", Type: models.RRSetTypeA, TTL: 3600},
			{Name: "mail", Type: models.RRSetTypeMX, TTL: 3600},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	rrsets, err := client.DNS.GetRRSets(context.Background(), "example.com")

	require.NoError(t, err)
	assert.Len(t, rrsets, 2)
}

func TestDNSService_UpsertRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/dns/example.com/records", r.URL.Path)

		var req models.RecordPatchRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		require.Len(t, req.Ops, 1)
		assert.Equal(t, models.RecordOpUpsert, req.Ops[0].Op)
		assert.Equal(t, "www", req.Ops[0].Record.Name)
		assert.Equal(t, models.RRSetTypeA, req.Ops[0].Record.Type)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.DNS.UpsertRecord(context.Background(), "example.com", models.Record{
		Name:  "www",
		Type:  models.RRSetTypeA,
		TTL:   3600,
		RData: "1.2.3.4",
	})

	require.NoError(t, err)
}

func TestDNSService_DeleteRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req models.RecordPatchRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		require.Len(t, req.Ops, 1)
		assert.Equal(t, models.RecordOpRemove, req.Ops[0].Op)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.DNS.DeleteRecord(context.Background(), "example.com", models.Record{
		Name:  "www",
		Type:  models.RRSetTypeA,
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

func TestAvailabilityService_CheckAvailability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/availability")

		domains := r.URL.Query()["domains"]
		assert.Contains(t, domains, "example.com")

		_ = json.NewEncoder(w).Encode(models.AvailabilityResponse{
			Results: []models.DomainAvailability{
				{Domain: "example.com", Status: models.AvailabilityStatusUnavailable},
			},
			Meta: models.AvailabilityMeta{Total: 1, ProcessingTimeMs: 50},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.Availability.CheckAvailability(context.Background(), []string{"example.com"})

	require.NoError(t, err)
	assert.Len(t, result.Results, 1)
	assert.Equal(t, "example.com", result.Results[0].Domain)
	assert.Equal(t, models.AvailabilityStatusUnavailable, result.Results[0].Status)
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

func TestDomainsService_ListDomains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/domains")

		_ = json.NewEncoder(w).Encode(models.DomainListResponse{
			Results: []models.Domain{
				{DomainID: "domain_123", Name: "example.com"},
			},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	domains, err := client.Domains.ListDomains(context.Background(), nil)

	require.NoError(t, err)
	assert.Len(t, domains, 1)
	assert.Equal(t, "example.com", domains[0].Name)
}

func TestContactsService_CreateContact(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/contacts", r.URL.Path)

		var req models.ContactCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "John", req.FirstName)
		assert.Equal(t, "Doe", req.LastName)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(models.Contact{
			ContactID: "contact_123",
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	contact, err := client.Contacts.CreateContact(context.Background(), &models.ContactCreateRequest{
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john@example.com",
		Phone:      "+1.5551234567",
		Street:     "123 Main St",
		City:       "New York",
		PostalCode: "10001",
		Country:    "US",
		Disclose:   false,
	})

	require.NoError(t, err)
	assert.Equal(t, "John", contact.FirstName)
	assert.Equal(t, "Doe", contact.LastName)
}

func TestUsersService_GetCurrentUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/users/me", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.CurrentUser{
			User: models.User{
				UserID: "user_123",
				Email:  "user@example.com",
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	user, err := client.Users.GetCurrentUser(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "user@example.com", user.Email)
}
