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

func TestEmailForwardsService_ListEmailForwards(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/email-forwards", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.EmailForwardListResponse{
			Results: []models.EmailForward{
				{EmailForwardID: "email_forward_123", Hostname: "example.com.", Enabled: true},
				{EmailForwardID: "email_forward_456", Hostname: "www.example.com.", Enabled: false},
			},
			Pagination: models.Pagination{HasNextPage: false, CurrentPage: 1},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	forwards, err := client.EmailForwards.ListEmailForwards(context.Background(), nil)

	require.NoError(t, err)
	require.Len(t, forwards, 2)
	assert.Equal(t, models.EmailForwardID("email_forward_123"), forwards[0].EmailForwardID)
	assert.Equal(t, "www.example.com.", forwards[1].Hostname)
}

func TestEmailForwardsService_ListEmailForwardsPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/email-forwards", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "25", r.URL.Query().Get("page_size"))
		assert.Equal(t, "hostname", r.URL.Query().Get("sort_by"))
		assert.Equal(t, "true", r.URL.Query().Get("enabled"))

		_ = json.NewEncoder(w).Encode(models.EmailForwardListResponse{
			Results: []models.EmailForward{
				{EmailForwardID: "email_forward_123", Hostname: "example.com.", Enabled: true},
			},
			Pagination: models.Pagination{HasNextPage: false, CurrentPage: 2},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	enabled := true
	resp, err := client.EmailForwards.ListEmailForwardsPage(context.Background(), &models.ListEmailForwardsOptions{
		Page:     2,
		PageSize: 25,
		SortBy:   models.EmailForwardSortByHostname,
		Enabled:  &enabled,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, models.EmailForwardID("email_forward_123"), resp.Results[0].EmailForwardID)
}

func TestEmailForwardsService_GetEmailForward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/email-forwards/email_forward_123", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.EmailForward{
			EmailForwardID: "email_forward_123",
			Hostname:       "example.com.",
			Enabled:        true,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	forward, err := client.EmailForwards.GetEmailForward(context.Background(), models.EmailForwardID("email_forward_123"))

	require.NoError(t, err)
	require.NotNil(t, forward)
	assert.Equal(t, models.EmailForwardID("email_forward_123"), forward.EmailForwardID)
	assert.Equal(t, "example.com.", forward.Hostname)
}

func TestEmailForwardsService_CreateEmailForward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/email-forwards", r.URL.Path)

		var req models.EmailForwardCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "example.com.", req.Hostname)
		require.Len(t, req.Aliases, 1)
		assert.Equal(t, "info", req.Aliases[0].Alias)

		_ = json.NewEncoder(w).Encode(models.EmailForward{
			EmailForwardID: "email_forward_123",
			Hostname:       "example.com.",
			Enabled:        true,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	forward, err := client.EmailForwards.CreateEmailForward(context.Background(), &models.EmailForwardCreateRequest{
		Hostname: "example.com.",
		Aliases: []models.EmailForwardAliasCreate{
			{Alias: "info", ForwardTo: []string{"dest@example.com"}},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, forward)
	assert.Equal(t, models.EmailForwardID("email_forward_123"), forward.EmailForwardID)
}

func TestEmailForwardsService_DeleteEmailForward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/email-forwards/email_forward_123", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.EmailForwards.DeleteEmailForward(context.Background(), models.EmailForwardID("email_forward_123"))

	require.NoError(t, err)
}

func TestEmailForwardsService_EnableEmailForward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/email-forwards/email_forward_123/enable", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.EmailForwards.EnableEmailForward(context.Background(), models.EmailForwardID("email_forward_123"))

	require.NoError(t, err)
}

func TestEmailForwardsService_DisableEmailForward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/email-forwards/email_forward_123/disable", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.EmailForwards.DisableEmailForward(context.Background(), models.EmailForwardID("email_forward_123"))

	require.NoError(t, err)
}

func TestEmailForwardsService_CreateAlias(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/email-forwards/email_forward_123/aliases", r.URL.Path)

		var req models.EmailForwardAliasCreate
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "info", req.Alias)
		assert.Equal(t, []string{"dest@example.com"}, req.ForwardTo)

		_ = json.NewEncoder(w).Encode(models.EmailForwardAlias{
			EmailForwardAliasID: models.EmailForwardAliasID("email_forward_alias_456"),
			Alias:               "info",
			ForwardTo:           []string{"dest@example.com"},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	alias, err := client.EmailForwards.CreateAlias(
		context.Background(),
		models.EmailForwardID("email_forward_123"),
		&models.EmailForwardAliasCreate{
			Alias:     "info",
			ForwardTo: []string{"dest@example.com"},
		},
	)

	require.NoError(t, err)
	require.NotNil(t, alias)
	assert.Equal(t, models.EmailForwardAliasID("email_forward_alias_456"), alias.EmailForwardAliasID)
	assert.Equal(t, "info", alias.Alias)
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

func TestEmailForwardsService_DeleteAlias(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/email-forwards/email_forward_123/aliases/email_forward_alias_456", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.EmailForwards.DeleteAlias(
		context.Background(),
		models.EmailForwardID("email_forward_123"),
		models.EmailForwardAliasID("email_forward_alias_456"),
	)

	require.NoError(t, err)
}

func TestEmailForwardsService_ListEmailForwardsByZone(t *testing.T) {
	t.Run("decodes zone wrapper response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/v1/dns/example.com/email-forwards", r.URL.Path)

			_ = json.NewEncoder(w).Encode(models.EmailForwardZone{
				ZoneID:   "zone_123",
				ZoneName: "example.com.",
				EmailForwards: []models.EmailForward{
					{EmailForwardID: "email_forward_123", Hostname: "example.com.", Enabled: true},
					{EmailForwardID: "email_forward_456", Hostname: "www.example.com.", Enabled: false},
				},
			})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		forwards, err := client.EmailForwards.ListEmailForwardsByZone(context.Background(), "example.com")

		require.NoError(t, err)
		require.Len(t, forwards, 2)
		assert.Equal(t, models.EmailForwardID("email_forward_123"), forwards[0].EmailForwardID)
		assert.Equal(t, "www.example.com.", forwards[1].Hostname)
	})

	t.Run("falls back to bare list response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/v1/dns/example.com/email-forwards", r.URL.Path)

			_ = json.NewEncoder(w).Encode([]models.EmailForward{
				{EmailForwardID: "email_forward_123", Hostname: "example.com.", Enabled: true},
			})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		forwards, err := client.EmailForwards.ListEmailForwardsByZone(context.Background(), "example.com")

		require.NoError(t, err)
		require.Len(t, forwards, 1)
		assert.Equal(t, "example.com.", forwards[0].Hostname)
	})
}

func TestEmailForwardsService_GetMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/email-forwards/email_forward_123/metrics", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.EmailForwardMetrics{
			TotalLogs: 42,
			ByStatus: map[models.EmailForwardLogStatus]int{
				models.EmailForwardLogStatusDelivered:  40,
				models.EmailForwardLogStatusHardBounce: 2,
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	metrics, err := client.EmailForwards.GetMetrics(context.Background(), models.EmailForwardID("email_forward_123"), nil)

	require.NoError(t, err)
	require.NotNil(t, metrics)
	assert.Equal(t, 42, metrics.TotalLogs)
	assert.Equal(t, 40, metrics.ByStatus[models.EmailForwardLogStatusDelivered])
}
