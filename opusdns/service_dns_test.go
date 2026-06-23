package opusdns

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestDNSService_SetZoneVanitySet(t *testing.T) {
	t.Run("assigns set", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PATCH", r.Method)
			assert.Equal(t, "/v1/dns/example.com/vanity-set", r.URL.Path)

			var req models.ZoneVanitySetUpdateRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			require.NotNil(t, req.VanityNameserverSetID)
			assert.Equal(t, models.VanityNameserverSetID("vns_1"), *req.VanityNameserverSetID)

			_ = json.NewEncoder(w).Encode(map[string]models.Zone{"zone": {Name: "example.com"}})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		setID := models.VanityNameserverSetID("vns_1")
		zone, err := client.DNS.SetZoneVanitySet(context.Background(), "example.com.", &setID)
		require.NoError(t, err)
		assert.Equal(t, "example.com", zone.Name)
	})

	t.Run("clears set with null", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var raw map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&raw)
			val, present := raw["vanity_nameserver_set_id"]
			assert.True(t, present)
			assert.Nil(t, val)
			_ = json.NewEncoder(w).Encode(map[string]models.Zone{"zone": {Name: "example.com"}})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		_, err = client.DNS.SetZoneVanitySet(context.Background(), "example.com", nil)
		require.NoError(t, err)
	})
}

func TestDNSService_GetZoneWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/dns/example.com", r.URL.Path)
		assert.Equal(t, []string{"tags"}, r.URL.Query()["include"])

		_ = json.NewEncoder(w).Encode(models.Zone{
			Name:         "example.com",
			DNSSECStatus: models.DNSSECStatusDisabled,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	zone, err := client.DNS.GetZoneWithOptions(context.Background(), "example.com.", &models.GetZoneOptions{
		Include: []models.ZoneIncludeField{models.ZoneIncludeTags},
	})

	require.NoError(t, err)
	assert.Equal(t, "example.com", zone.Name)
}

func TestDNSService_GetSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/dns/summary", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.ZoneSummary{
			TotalZones: 7,
			ZonesByDNSSEC: map[models.DNSSECStatus]int{
				models.DNSSECStatusEnabled:  3,
				models.DNSSECStatusDisabled: 4,
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	summary, err := client.DNS.GetSummary(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 7, summary.TotalZones)
	assert.Equal(t, 3, summary.ZonesByDNSSEC[models.DNSSECStatusEnabled])
}

func TestDNSService_PatchRecords(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/dns/example.com/records", r.URL.Path)

		var req models.RecordPatchRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.Len(t, req.Ops, 2)
		assert.Equal(t, models.RecordOpUpsert, req.Ops[0].Op)
		assert.Equal(t, "www", req.Ops[0].Record.Name)
		assert.Equal(t, models.RecordOpRemove, req.Ops[1].Op)
		assert.Equal(t, "old", req.Ops[1].Record.Name)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.DNS.PatchRecords(context.Background(), "example.com.", []models.RecordOperation{
		{Op: models.RecordOpUpsert, Record: models.Record{Name: "www", Type: models.RRSetTypeA, TTL: 3600, RData: "1.2.3.4"}},
		{Op: models.RecordOpRemove, Record: models.Record{Name: "old", Type: models.RRSetTypeA, TTL: 3600, RData: "5.6.7.8"}},
	})

	require.NoError(t, err)
}

func TestDNSService_ListZonesPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/dns", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "50", r.URL.Query().Get("page_size"))

		_ = json.NewEncoder(w).Encode(models.ZoneListResponse{
			Results:    []models.Zone{{Name: "example.com"}},
			Pagination: models.Pagination{HasNextPage: true, CurrentPage: 2},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.DNS.ListZonesPage(context.Background(), &models.ListZonesOptions{
		Page:     2,
		PageSize: 50,
	})

	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "example.com", resp.Results[0].Name)
	assert.True(t, resp.Pagination.HasNextPage)
	assert.Equal(t, 2, resp.Pagination.CurrentPage)
}
