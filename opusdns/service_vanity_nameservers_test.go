package opusdns

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVanityNameserversService_ListSets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/vanity-nameserver-sets", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.VanityNameserverSetListResponse{
			Results: []models.VanityNameserverSet{
				{SetID: "vns_1", Name: "Primary", Status: models.VanityNameserverSetStatusActive, IsDefault: true},
			},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	sets, err := client.VanityNameservers.ListSets(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, sets, 1)
	assert.Equal(t, models.VanityNameserverSetID("vns_1"), sets[0].SetID)
	assert.True(t, sets[0].IsDefault)
}

func TestVanityNameserversService_ListSetsPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/vanity-nameserver-sets", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("page_size"))
		_ = json.NewEncoder(w).Encode(models.VanityNameserverSetListResponse{
			Results: []models.VanityNameserverSet{
				{SetID: "vns_2", Name: "Secondary", Status: models.VanityNameserverSetStatusActive},
			},
			Pagination: models.Pagination{CurrentPage: 2, TotalPages: 2, HasNextPage: false, HasPreviousPage: true},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.VanityNameservers.ListSetsPage(context.Background(), &models.ListVanityNameserverSetsOptions{Page: 2, PageSize: 10})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, models.VanityNameserverSetID("vns_2"), resp.Results[0].SetID)
	assert.Equal(t, 2, resp.Pagination.CurrentPage)
	assert.False(t, resp.Pagination.HasNextPage)
}

func TestVanityNameserversService_GetSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/vanity-nameserver-sets/vns_1", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.VanityNameserverSet{SetID: "vns_1", Name: "Primary"})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	set, err := client.VanityNameservers.GetSet(context.Background(), "vns_1")
	require.NoError(t, err)
	assert.Equal(t, "Primary", set.Name)
}

func TestVanityNameserversService_CreateSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/vanity-nameserver-sets", r.URL.Path)

		var req models.VanityNameserverSetCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "Primary", req.Name)
		assert.Equal(t, "example.com", req.ParentDomainName)
		assert.Equal(t, []string{"ns1.example.com", "ns2.example.com"}, req.Hostnames)

		// Creation is asynchronous (202 Accepted) and returns the provisioning set.
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(models.VanityNameserverSet{SetID: "vns_1", Name: req.Name, Status: models.VanityNameserverSetStatusProvisioning})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	set, err := client.VanityNameservers.CreateSet(context.Background(), &models.VanityNameserverSetCreateRequest{
		Name:             "Primary",
		ParentDomainName: "example.com",
		SOARName:         "hostmaster.example.com",
		Hostnames:        []string{"ns1.example.com", "ns2.example.com"},
	})
	require.NoError(t, err)
	assert.Equal(t, models.VanityNameserverSetStatusProvisioning, set.Status)
}

func TestVanityNameserversService_CreateSet_RequiresHostnames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no request expected, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	// Without hostnames the request would carry "hostnames": null.
	_, err = client.VanityNameservers.CreateSet(context.Background(), &models.VanityNameserverSetCreateRequest{
		Name:             "Primary",
		ParentDomainName: "example.com",
		SOARName:         "hostmaster.example.com",
	})
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestVanityNameserversService_DeleteSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/vanity-nameserver-sets/vns_1", r.URL.Path)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.VanityNameservers.DeleteSet(context.Background(), "vns_1")
	require.NoError(t, err)
}

func TestVanityNameserversService_CheckSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/vanity-nameserver-sets/check", r.URL.Path)

		var req models.VanityNsCheckRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, models.VanityNameserverSetID("vns_1"), req.SetID)

		_ = json.NewEncoder(w).Encode(models.VanityNsCheckResponse{
			SetID:   "vns_1",
			Status:  models.VanityNameserverSetStatusActive,
			Summary: models.VanityNsCheckSummary{State: models.VanityNsCheckSummaryStateReady, Detail: "ok"},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.VanityNameservers.CheckSet(context.Background(), "vns_1")
	require.NoError(t, err)
	assert.Equal(t, models.VanityNsCheckSummaryStateReady, result.Summary.State)
}

func TestVanityNameserversService_SetDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/vanity-nameserver-sets/vns_1/default", r.URL.Path)
		// SetVanityNameserverSetDefaultRes wraps a VanityNameserverSetDTO,
		// which has no status or is_default.
		_, _ = w.Write([]byte(`{"vanity_nameserver_set": {
			"set_id": "vns_1", "name": "Primary", "parent_domain_name": "example.com",
			"soa_rname": "hostmaster.example.com",
			"nameservers": [{"hostname": "ns1.example.com", "position": 0},
			                {"hostname": "ns2.example.com", "position": 1}]
		}}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.VanityNameservers.SetDefault(context.Background(), "vns_1")
	require.NoError(t, err)
	set := result.VanityNameserverSet
	assert.Equal(t, models.VanityNameserverSetID("vns_1"), set.SetID)
	assert.Equal(t, "Primary", set.Name)
	assert.Equal(t, "hostmaster.example.com", set.SOARName)
	require.Len(t, set.Nameservers, 2)
	assert.Equal(t, "ns2.example.com", set.Nameservers[1].Hostname)
}

func TestVanityNameserversService_ClearDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/vanity-nameserver-sets/default", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.ClearVanityNameserverSetDefaultResponse{Cleared: true})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.VanityNameservers.ClearDefault(context.Background())
	require.NoError(t, err)
	assert.True(t, result.Cleared)
}

func TestVanityNameserversService_RestoreSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/vanity-nameserver-sets/vns_1/restore", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.VanityNameserverSet{SetID: "vns_1", Status: models.VanityNameserverSetStatusActive})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	set, err := client.VanityNameservers.RestoreSet(context.Background(), "vns_1")
	require.NoError(t, err)
	assert.Equal(t, models.VanityNameserverSetStatusActive, set.Status)
}

func TestVanityNameserversService_ListZonesReferencingSet(t *testing.T) {
	var pages []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/vanity-nameserver-sets/vns_1/zones", r.URL.Path)
		assert.Equal(t, strconv.Itoa(DefaultPageSize), r.URL.Query().Get("page_size"))
		page := r.URL.Query().Get("page")
		pages = append(pages, page)

		// ListZonesReferencingSetRes with DnsZoneDTO entries.
		hasNext := page == "1"
		name := "example.com"
		if !hasNext {
			name = "example.org"
		}
		_, _ = fmt.Fprintf(w, `{"results": [{"name": %q, "dnssec_status": "disabled", "rrsets": [],
			"vanity_nameserver_set_id": "vns_1", "created_on": null, "updated_on": null}],
			"pagination": {"total_pages": 2, "current_page": %s, "has_next_page": %t,
			               "has_previous_page": %t, "total_items": 2, "page_size": %d}}`,
			name, page, hasNext, !hasNext, DefaultPageSize)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	zones, err := client.VanityNameservers.ListZonesReferencingSet(context.Background(), "vns_1", nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2"}, pages)
	require.Len(t, zones, 2)
	assert.Equal(t, "example.com", zones[0].Name)
	assert.Equal(t, "example.org", zones[1].Name)
}

func TestVanityNameserversService_ListZonesReferencingSetPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/vanity-nameserver-sets/vns_1/zones", r.URL.Path)
		assert.Equal(t, "3", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("page_size"))
		_ = json.NewEncoder(w).Encode(models.ZonesReferencingSetResponse{
			Results:    []models.Zone{{Name: "example.com"}},
			Pagination: models.Pagination{CurrentPage: 3, HasNextPage: true},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.VanityNameservers.ListZonesReferencingSetPage(context.Background(), "vns_1",
		&models.ListVanityNameserverSetsOptions{Page: 3, PageSize: 10})
	require.NoError(t, err)
	require.Len(t, result.Results, 1)
	assert.True(t, result.Pagination.HasNextPage)
}

func TestVanityNameserversService_SetIDIsEscaped(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.VanityNameservers.DeleteSet(context.Background(), "../default")
	require.NoError(t, err)
	assert.Equal(t, "/v1/vanity-nameserver-sets/..%2Fdefault", gotPath)
}

func TestVanityNameserversService_SetRenewalMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/vanity-nameserver-sets/vns_123", r.URL.Path)

		var body models.SetRenewalModeRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, models.RenewalModeExpire, body.RenewalMode)

		_ = json.NewEncoder(w).Encode(models.VanityNameserverSet{
			SetID:       "vns_123",
			RenewalMode: models.RenewalModePtr(models.RenewalModeExpire),
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	set, err := client.VanityNameservers.SetRenewalMode(context.Background(), "vns_123", models.RenewalModeExpire)
	require.NoError(t, err)
	require.NotNil(t, set.RenewalMode)
	assert.Equal(t, models.RenewalModeExpire, *set.RenewalMode)
}

func TestVanityNameserversService_RetrySet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/vanity-nameserver-sets/vns_123/retry", r.URL.Path)

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(models.VanityNameserverSet{SetID: "vns_123"})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	set, err := client.VanityNameservers.RetrySet(context.Background(), "vns_123")
	require.NoError(t, err)
	assert.Equal(t, models.VanityNameserverSetID("vns_123"), set.SetID)
}
