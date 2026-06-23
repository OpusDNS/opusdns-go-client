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

func TestDomainsService_CancelTransfer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/domains/example.com/transfer", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Domains.CancelTransfer(context.Background(), "example.com")
	require.NoError(t, err)
}

func TestDomainsService_ListDomainsPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domains", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "50", r.URL.Query().Get("page_size"))

		_ = json.NewEncoder(w).Encode(models.DomainListResponse{
			Results: []models.Domain{
				{DomainID: "domain_456", Name: "second.com"},
			},
			Pagination: models.Pagination{HasNextPage: true},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Domains.ListDomainsPage(context.Background(), &models.ListDomainsOptions{
		Page:     2,
		PageSize: 50,
	})

	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "second.com", resp.Results[0].Name)
	assert.True(t, resp.Pagination.HasNextPage)
}

func TestDomainsService_GetDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domains/example.com", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.Domain{
			DomainID: "domain_123",
			Name:     "example.com",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	domain, err := client.Domains.GetDomain(context.Background(), "example.com")

	require.NoError(t, err)
	assert.Equal(t, "example.com", domain.Name)
	assert.Equal(t, models.DomainID("domain_123"), domain.DomainID)
}

func TestDomainsService_GetDomainWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domains/example.com", r.URL.Path)
		assert.Equal(t, []string{"tags"}, r.URL.Query()["include"])

		_ = json.NewEncoder(w).Encode(models.Domain{
			DomainID: "domain_123",
			Name:     "example.com",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	domain, err := client.Domains.GetDomainWithOptions(context.Background(), "example.com", &models.GetDomainOptions{
		Include: []models.DomainIncludeField{models.DomainIncludeTags},
	})

	require.NoError(t, err)
	assert.Equal(t, "example.com", domain.Name)
}

func TestDomainsService_CreateDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/domains", r.URL.Path)

		var req models.DomainCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "example.com", req.Name)

		_ = json.NewEncoder(w).Encode(models.Domain{
			DomainID: "domain_123",
			Name:     req.Name,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	domain, err := client.Domains.CreateDomain(context.Background(), &models.DomainCreateRequest{
		Name:        "example.com",
		RenewalMode: models.RenewalModeRenew,
	})

	require.NoError(t, err)
	assert.Equal(t, "example.com", domain.Name)
}

func TestDomainsService_UpdateDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/domains/example.com", r.URL.Path)

		var req models.DomainUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, []models.DomainClientStatus{models.DomainClientStatusUpdateProhibited}, req.Statuses)

		_ = json.NewEncoder(w).Encode(models.Domain{
			DomainID: "domain_123",
			Name:     "example.com",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	domain, err := client.Domains.UpdateDomain(context.Background(), "example.com", &models.DomainUpdateRequest{
		Statuses: []models.DomainClientStatus{models.DomainClientStatusUpdateProhibited},
	})

	require.NoError(t, err)
	assert.Equal(t, "example.com", domain.Name)
}

func TestDomainsService_DeleteDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/domains/example.com", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Domains.DeleteDomain(context.Background(), "example.com")
	require.NoError(t, err)
}

func TestDomainsService_TransferDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/domains/transfer", r.URL.Path)

		var req models.DomainTransferRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "example.com", req.Name)
		assert.Equal(t, "auth123", req.AuthCode)

		_ = json.NewEncoder(w).Encode(models.Domain{
			DomainID: "domain_123",
			Name:     req.Name,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	domain, err := client.Domains.TransferDomain(context.Background(), &models.DomainTransferRequest{
		Name:        "example.com",
		AuthCode:    "auth123",
		RenewalMode: models.RenewalModeRenew,
	})

	require.NoError(t, err)
	assert.Equal(t, "example.com", domain.Name)
}

func TestDomainsService_RenewDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/domains/example.com/renew", r.URL.Path)

		var req models.DomainRenewRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, 1, req.Period)

		_ = json.NewEncoder(w).Encode(models.Domain{
			DomainID: "domain_123",
			Name:     "example.com",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	domain, err := client.Domains.RenewDomain(context.Background(), "example.com", &models.DomainRenewRequest{
		Period: 1,
	})

	require.NoError(t, err)
	assert.Equal(t, "example.com", domain.Name)
}

func TestDomainsService_RestoreDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/domains/example.com/restore", r.URL.Path)

		var req models.DomainRestoreRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, 1, req.Period)

		_ = json.NewEncoder(w).Encode(models.Domain{
			DomainID: "domain_123",
			Name:     "example.com",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	domain, err := client.Domains.RestoreDomain(context.Background(), "example.com", &models.DomainRestoreRequest{
		Period: 1,
	})

	require.NoError(t, err)
	assert.Equal(t, "example.com", domain.Name)
}

func TestDomainsService_GetSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domains/summary", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.DomainSummary{
			TotalDomains: 42,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	summary, err := client.Domains.GetSummary(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 42, summary.TotalDomains)
}

func TestDomainsService_CheckDomains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domains/check", r.URL.Path)
		assert.Equal(t, []string{"example.com", "example.net"}, r.URL.Query()["domains"])

		_ = json.NewEncoder(w).Encode(models.DomainCheckResponse{
			Results: []models.DomainAvailabilityResult{
				{Domain: "example.com", Available: false},
				{Domain: "example.net", Available: true},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Domains.CheckDomains(context.Background(), []string{"example.com", "example.net"})

	require.NoError(t, err)
	require.Len(t, resp.Results, 2)
	assert.True(t, resp.Results[1].Available)
}

func TestDomainsService_DeleteDNSSEC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/domains/example.com/dnssec", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Domains.DeleteDNSSEC(context.Background(), "example.com")
	require.NoError(t, err)
}

func TestDomainsService_EnableDNSSEC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/domains/example.com/dnssec/enable", r.URL.Path)

		_ = json.NewEncoder(w).Encode([]models.DomainDNSSECDataResponse{
			{RecordType: models.DNSSECRecordTypeDSData, Algorithm: 13},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	data, err := client.Domains.EnableDNSSEC(context.Background(), "example.com")

	require.NoError(t, err)
	require.Len(t, data, 1)
	assert.Equal(t, models.DNSSECRecordTypeDSData, data[0].RecordType)
}
