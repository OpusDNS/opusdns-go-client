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

func TestDomainForwardsService_GetDomainForwardSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domain-forwards/example.com/https", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.DomainForwardSetResponse{
			Hostname: "example.com",
			Protocol: models.HttpProtocolHTTPS,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	set, err := client.DomainForwards.GetDomainForwardSet(context.Background(), "example.com", models.HttpProtocolHTTPS)
	require.NoError(t, err)
	assert.Equal(t, models.HttpProtocolHTTPS, set.Protocol)
}

func TestDomainForwardsService_CreateDomainForwardSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/domain-forwards/example.com", r.URL.Path)

		var req models.DomainForwardSetCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, models.HttpProtocolHTTPS, req.Protocol)
		require.Len(t, req.Redirects, 1)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(models.DomainForwardSetResponse{Hostname: "example.com", Protocol: req.Protocol})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	set, err := client.DomainForwards.CreateDomainForwardSet(context.Background(), "example.com", &models.DomainForwardSetCreateRequest{
		Protocol: models.HttpProtocolHTTPS,
		Redirects: []models.HttpRedirectRequest{
			{RequestPath: "/", TargetProtocol: models.HttpProtocolHTTPS, TargetHostname: "dest.com", TargetPath: "/", RedirectCode: models.RedirectCodePermanent},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "example.com", set.Hostname)
}

func TestDomainForwardsService_PatchRedirects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/domain-forwards", r.URL.Path)

		var req models.DomainForwardPatchOps
		_ = json.NewDecoder(r.Body).Decode(&req)
		require.Len(t, req.Ops, 1)
		assert.Equal(t, models.PatchOpRemove, req.Ops[0].Op)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.DomainForwards.PatchRedirects(context.Background(), &models.DomainForwardPatchOps{
		Ops: []models.DomainForwardPatchOp{
			{Op: models.PatchOpRemove, Redirect: models.HttpRedirectRemove{RequestProtocol: models.HttpProtocolHTTPS, RequestHostname: "example.com", RequestPath: "/"}},
		},
	})
	require.NoError(t, err)
}

func TestDomainForwardsService_ListDomainForwards(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domain-forwards", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.DomainForwardListResponse{
			Results: []models.DomainForward{
				{Hostname: "example.com", Enabled: true},
			},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	forwards, err := client.DomainForwards.ListDomainForwards(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, forwards, 1)
	assert.Equal(t, "example.com", forwards[0].Hostname)
}

func TestDomainForwardsService_ListDomainForwardsPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domain-forwards", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "hostname", r.URL.Query().Get("sort_by"))
		assert.Equal(t, "true", r.URL.Query().Get("enabled"))

		_ = json.NewEncoder(w).Encode(models.DomainForwardListResponse{
			Results: []models.DomainForward{
				{Hostname: "example.com", Enabled: true},
			},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	enabled := true
	resp, err := client.DomainForwards.ListDomainForwardsPage(context.Background(), &models.ListDomainForwardsOptions{
		Page:    2,
		SortBy:  models.DomainForwardSortByHostname,
		Enabled: &enabled,
	})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "example.com", resp.Results[0].Hostname)
}

func TestDomainForwardsService_GetDomainForward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domain-forwards/example.com", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.DomainForward{
			Hostname: "example.com",
			Enabled:  true,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	forward, err := client.DomainForwards.GetDomainForward(context.Background(), "example.com")
	require.NoError(t, err)
	require.NotNil(t, forward)
	assert.Equal(t, "example.com", forward.Hostname)
	assert.True(t, forward.Enabled)
}

func TestDomainForwardsService_CreateDomainForward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/domain-forwards", r.URL.Path)

		var req models.DomainForwardCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "example.com", req.Hostname)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(models.DomainForward{
			Hostname: req.Hostname,
			Enabled:  true,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	forward, err := client.DomainForwards.CreateDomainForward(context.Background(), &models.DomainForwardCreateRequest{
		Hostname: "example.com",
		Enabled:  true,
		HTTPS: &models.DomainForwardProtocolSetRequest{
			Redirects: []models.HttpRedirectRequest{
				{RequestPath: "/", TargetProtocol: models.HttpProtocolHTTPS, TargetHostname: "dest.com", TargetPath: "/", RedirectCode: models.RedirectCodePermanent},
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, forward)
	assert.Equal(t, "example.com", forward.Hostname)
}

func TestDomainForwardsService_DeleteDomainForward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/domain-forwards/example.com", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.DomainForwards.DeleteDomainForward(context.Background(), "example.com")
	require.NoError(t, err)
}

func TestDomainForwardsService_DeleteDomainForwardConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/domain-forwards/example.com/https", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.DomainForwards.DeleteDomainForwardConfig(context.Background(), "example.com", models.HttpProtocolHTTPS)
	require.NoError(t, err)
}

func TestDomainForwardsService_EnableDomainForward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/domain-forwards/example.com/enable", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.DomainForwards.EnableDomainForward(context.Background(), "example.com")
	require.NoError(t, err)
}

func TestDomainForwardsService_DisableDomainForward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/domain-forwards/example.com/disable", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.DomainForwards.DisableDomainForward(context.Background(), "example.com")
	require.NoError(t, err)
}

func TestDomainForwardsService_GetMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domain-forwards/metrics", r.URL.Path)
		assert.Equal(t, "example.com", r.URL.Query().Get("hostname"))
		assert.Equal(t, "7d", r.URL.Query().Get("time_range"))

		_ = json.NewEncoder(w).Encode(models.DomainForwardMetrics{
			InvokedForwards:    3,
			ConfiguredForwards: 5,
			TotalVisits:        100,
			UniqueVisits:       80,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	metrics, err := client.DomainForwards.GetMetrics(context.Background(), &models.DomainForwardMetricsOptions{
		Hostname:  "example.com",
		TimeRange: models.TimeRange7D,
	})
	require.NoError(t, err)
	require.NotNil(t, metrics)
	assert.Equal(t, 100, metrics.TotalVisits)
	assert.Equal(t, 5, metrics.ConfiguredForwards)
}
