package opusdns

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
		assert.NotNil(t, client.Tags)
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

	t.Run("does not mutate caller options", func(t *testing.T) {
		requestedPages := []string{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestedPages = append(requestedPages, r.URL.Query().Get("page"))

			_ = json.NewEncoder(w).Encode(models.ZoneListResponse{
				Results: []models.Zone{{Name: "example.com"}},
				Pagination: models.Pagination{
					HasNextPage: r.URL.Query().Get("page") == "1",
					CurrentPage: 1,
				},
			})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		opts := &models.ListZonesOptions{
			Page:     42,
			PageSize: 25,
			SortBy:   models.ZoneSortByName,
		}

		_, err = client.DNS.ListZones(context.Background(), opts)

		require.NoError(t, err)
		assert.Equal(t, 42, opts.Page)
		assert.Equal(t, 25, opts.PageSize)
		assert.Equal(t, []string{"1", "2"}, requestedPages)
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
				{Name: "www", Type: models.RRSetTypeA, TTL: 3600, Records: []models.RecordCreate{{RData: "1.2.3.4"}}},
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

func TestDNSService_RRSetWrites(t *testing.T) {
	t.Run("patch rrsets uses rrset payload shape", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PATCH", r.Method)
			assert.Equal(t, "/v1/dns/example.com/rrsets", r.URL.Path)

			var req models.RRSetPatchRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			require.Len(t, req.Ops, 1)
			assert.Equal(t, models.RecordOpUpsert, req.Ops[0].Op)
			assert.Equal(t, "www", req.Ops[0].RRSet.Name)
			assert.Equal(t, models.RRSetTypeHTTPS, req.Ops[0].RRSet.Type)
			assert.Equal(t, "1 . alpn=h2", req.Ops[0].RRSet.Records[0].RData)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		err = client.DNS.PatchRRSets(context.Background(), "example.com", []models.RRSetPatchOp{
			{
				Op: models.RecordOpUpsert,
				RRSet: models.RRSetPatch{
					Name:    "www",
					Type:    models.RRSetTypeHTTPS,
					TTL:     300,
					Records: []models.RecordCreate{{RData: "1 . alpn=h2"}},
				},
			},
		})

		require.NoError(t, err)
	})

	t.Run("put rrsets replaces all rrsets", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PUT", r.Method)
			assert.Equal(t, "/v1/dns/example.com/rrsets", r.URL.Path)

			var req models.RRSetUpdateRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			require.Len(t, req.RRSets, 1)
			assert.Equal(t, models.RRSetTypeA, req.RRSets[0].Type)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		err = client.DNS.PutRRSets(context.Background(), "example.com.", []models.RRSetCreate{
			{Name: "@", Type: models.RRSetTypeA, TTL: 300, Records: []models.RecordCreate{{RData: "192.0.2.1"}}},
		})

		require.NoError(t, err)
	})
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

func TestContactsService_VerifyContact(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/contacts/verify", r.URL.Path)
		assert.Equal(t, "verification-token", r.URL.Query().Get("token"))
		_ = json.NewEncoder(w).Encode(struct{}{})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Contacts.VerifyContact(context.Background(), &models.ContactVerificationRequest{Token: "verification-token"})

	require.NoError(t, err)
}

func TestDomainsService_DNSSEC(t *testing.T) {
	t.Run("get returns dnssec data array", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/v1/domains/example.com/dnssec", r.URL.Path)
			_ = json.NewEncoder(w).Encode([]models.DomainDNSSECDataResponse{
				{RecordType: models.DNSSECRecordTypeDSData, Algorithm: 13},
			})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		data, err := client.Domains.GetDNSSEC(context.Background(), "example.com")

		require.NoError(t, err)
		require.Len(t, data, 1)
		assert.Equal(t, models.DNSSECRecordTypeDSData, data[0].RecordType)
	})

	t.Run("put replaces dnssec data", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PUT", r.Method)
			assert.Equal(t, "/v1/domains/example.com/dnssec", r.URL.Path)

			var req []models.DomainDNSSECDataCreate
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			require.Len(t, req, 1)
			assert.Equal(t, models.DNSSECRecordTypeKeyData, req[0].RecordType)

			_ = json.NewEncoder(w).Encode([]models.DomainDNSSECDataResponse{
				{RecordType: req[0].RecordType, Algorithm: req[0].Algorithm},
			})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		data, err := client.Domains.PutDNSSEC(context.Background(), "example.com", []models.DomainDNSSECDataCreate{
			{RecordType: models.DNSSECRecordTypeKeyData, Algorithm: 13},
		})

		require.NoError(t, err)
		require.Len(t, data, 1)
	})

	t.Run("disable expects no content", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/domains/example.com/dnssec/disable", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		err = client.Domains.DisableDNSSEC(context.Background(), "example.com")

		require.NoError(t, err)
	})
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
	t.Run("returns domains", func(t *testing.T) {
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
	})

	t.Run("sends documented filters", func(t *testing.T) {
		now := time.Date(2026, 5, 5, 8, 0, 0, 0, time.UTC)
		trueValue := true

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			assert.Equal(t, []string{"tag_123", "tag_456"}, query["tag_ids"])
			assert.Equal(t, "match_all", query.Get("tag_mode"))
			assert.Equal(t, "true", query.Get("is_premium"))
			assert.Equal(t, now.Format(time.RFC3339), query.Get("updated_after"))
			assert.Equal(t, "true", query.Get("expires_in_30_days"))
			assert.Equal(t, now.Format(time.RFC3339), query.Get("registered_after"))
			assert.Equal(t, []string{"ok", "clientTransferProhibited"}, query["registry_statuses"])
			assert.Equal(t, []string{"tags"}, query["include"])

			_ = json.NewEncoder(w).Encode(models.DomainListResponse{
				Results:    []models.Domain{},
				Pagination: models.Pagination{HasNextPage: false},
			})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		_, err = client.Domains.ListDomains(context.Background(), &models.ListDomainsOptions{
			TagIDs:          []models.TagID{"tag_123", "tag_456"},
			TagMode:         models.TagFilterModeMatchAll,
			IsPremium:       &trueValue,
			UpdatedAfter:    &now,
			ExpiresIn30Days: &trueValue,
			RegisteredAfter: &now,
			RegistryStatuses: []string{
				"ok",
				"clientTransferProhibited",
			},
			Include: []models.DomainIncludeField{models.DomainIncludeTags},
		})

		require.NoError(t, err)
	})
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

		_ = json.NewEncoder(w).Encode(models.User{
			UserID: "user_123",
			Email:  "user@example.com",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	user, err := client.Users.GetCurrentUser(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "user@example.com", user.Email)
}

func TestUsersService_ListUsersOmitsUnsupportedFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		assert.Equal(t, "sam", query.Get("search"))
		assert.Empty(t, query.Get("email"))
		assert.Empty(t, query.Get("username"))
		assert.Empty(t, query.Get("status"))

		_ = json.NewEncoder(w).Encode(models.UserListResponse{
			Results:    []models.User{},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	_, err = client.Users.ListUsers(context.Background(), &models.ListUsersOptions{
		Search:   "sam",
		Email:    "sam@example.com",
		Username: "sam",
		Status:   models.UserStatusActive,
	})

	require.NoError(t, err)
}

func TestOrganizationsService_ListOrganizations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations", r.URL.Path)
		assert.Equal(t, "DE", r.URL.Query().Get("country_code"))

		_ = json.NewEncoder(w).Encode(models.OrganizationListResponse{
			Results:    []models.Organization{{OrganizationID: "organization_123", Name: "Example"}},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	orgs, err := client.Organizations.ListOrganizations(context.Background(), &models.ListOrganizationsOptions{
		CountryCode: "DE",
	})

	require.NoError(t, err)
	require.Len(t, orgs, 1)
	assert.Equal(t, "Example", orgs[0].Name)
}

func TestTagsService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tags", r.URL.Path)
		assert.Equal(t, []string{"DOMAIN"}, r.URL.Query()["tag_types"])

		_ = json.NewEncoder(w).Encode(models.TagListResponse{
			Results: []models.Tag{
				{TagID: "tag_123", Label: "Production", Type: models.TagTypeDomain, Color: models.TagColor1},
			},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	tags, err := client.Tags.ListTags(context.Background(), &models.ListTagsOptions{
		TagTypes: []models.TagType{models.TagTypeDomain},
	})

	require.NoError(t, err)
	require.Len(t, tags, 1)
	assert.Equal(t, "Production", tags[0].Label)
}

func TestDomainForwardsService_UpdateDomainForwardConfig(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++

		switch requests {
		case 1:
			assert.Equal(t, "PUT", r.Method)
			assert.Equal(t, "/v1/domain-forwards/adaenemark.de/http", r.URL.Path)

			var req models.DomainForwardProtocolSetRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			require.Len(t, req.Redirects, 1)
			assert.Equal(t, "/shop", req.Redirects[0].RequestPath)

			_ = json.NewEncoder(w).Encode(models.DomainForwardProtocolSet{
				Redirects: []models.HttpRedirect{
					{
						RequestProtocol: models.HttpProtocolHTTP,
						RequestHostname: "adaenemark.de",
						RequestPath:     "/shop",
						TargetProtocol:  models.HttpProtocolHTTPS,
						TargetHostname:  "www.adaenemark.de",
						TargetPath:      "/store",
						RedirectCode:    models.RedirectCodePermanent,
					},
				},
			})
		case 2:
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/v1/domain-forwards/adaenemark.de", r.URL.Path)

			_ = json.NewEncoder(w).Encode(models.DomainForward{
				Hostname: "adaenemark.de",
				Enabled:  true,
				HTTP: &models.DomainForwardProtocolSet{
					Redirects: []models.HttpRedirect{
						{
							RequestProtocol: models.HttpProtocolHTTP,
							RequestHostname: "adaenemark.de",
							RequestPath:     "/shop",
							TargetProtocol:  models.HttpProtocolHTTPS,
							TargetHostname:  "www.adaenemark.de",
							TargetPath:      "/store",
							RedirectCode:    models.RedirectCodePermanent,
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected request %d: %s %s", requests, r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	forward, err := client.DomainForwards.UpdateDomainForwardConfig(
		context.Background(),
		"adaenemark.de",
		models.HttpProtocolHTTP,
		&models.DomainForwardProtocolSetRequest{
			Redirects: []models.HttpRedirectRequest{
				{
					RequestPath:    "/shop",
					TargetProtocol: models.HttpProtocolHTTPS,
					TargetHostname: "www.adaenemark.de",
					TargetPath:     "/store",
					RedirectCode:   models.RedirectCodePermanent,
				},
			},
		},
	)

	require.NoError(t, err)
	require.NotNil(t, forward)
	assert.Equal(t, "adaenemark.de", forward.Hostname)
	require.NotNil(t, forward.HTTP)
	require.Len(t, forward.HTTP.Redirects, 1)
	assert.Equal(t, "/store", forward.HTTP.Redirects[0].TargetPath)
	assert.Equal(t, 2, requests)
}

func TestDomainForwardsService_ListDomainForwardsByZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/dns/example.com/domain-forwards", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.DomainForwardZone{
			ZoneID:   "zone_123",
			ZoneName: "example.com.",
			DomainForwards: []models.DomainForward{
				{
					Hostname: "example.com.",
					Enabled:  true,
				},
				{
					Hostname: "www.example.com.",
					Enabled:  true,
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	forwards, err := client.DomainForwards.ListDomainForwardsByZone(context.Background(), "example.com")

	require.NoError(t, err)
	require.Len(t, forwards, 2)
	assert.Equal(t, "example.com.", forwards[0].Hostname)
	assert.Equal(t, "www.example.com.", forwards[1].Hostname)
}

func TestEmailForwardsService_UpdateAlias(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/email-forwards/email_forward_123/aliases/email_forward_alias_456", r.URL.Path)

		var req models.EmailForwardAliasUpdate
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, []string{"new@adaenemark.de"}, req.ForwardTo)

		_ = json.NewEncoder(w).Encode(models.EmailForwardAlias{
			EmailForwardAliasID: models.EmailForwardAliasID("email_forward_alias_456"),
			Alias:               "info",
			ForwardTo:           []string{"new@adaenemark.de"},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	alias, err := client.EmailForwards.UpdateAlias(
		context.Background(),
		models.EmailForwardID("email_forward_123"),
		models.EmailForwardAliasID("email_forward_alias_456"),
		&models.EmailForwardAliasUpdate{
			ForwardTo: []string{"new@adaenemark.de"},
		},
	)

	require.NoError(t, err)
	require.NotNil(t, alias)
	assert.Equal(t, models.EmailForwardAliasID("email_forward_alias_456"), alias.EmailForwardAliasID)
	assert.Equal(t, "info", alias.Alias)
	assert.Equal(t, []string{"new@adaenemark.de"}, alias.ForwardTo)
}
