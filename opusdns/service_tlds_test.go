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

func TestTLDsService_ListTLDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/tlds")
		assert.Equal(t, "com", r.URL.Query().Get("search"))

		_ = json.NewEncoder(w).Encode(models.TLDListResponse{
			TLDConfigurations: []models.TLDConfiguration{
				{
					Enabled: true,
					TLDs: []models.TLDInfo{
						{Name: "com", Type: models.TLDTypeGTLD},
						{Name: "net", Type: models.TLDTypeGTLD},
					},
				},
				{
					Enabled: false,
					TLDs: []models.TLDInfo{
						{Name: "xyz", Type: models.TLDTypeGTLD},
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	tlds, err := client.TLDs.ListTLDs(context.Background(), &models.ListTLDsOptions{Search: "com"})

	require.NoError(t, err)
	assert.Len(t, tlds, 2)
	assert.Equal(t, "com", tlds[0].Name)
	assert.Equal(t, models.TLDTypeGTLD, tlds[0].Type)
	assert.True(t, tlds[0].Available)
	assert.Equal(t, "net", tlds[1].Name)
}

func TestTLDsService_GetTLD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/tlds/com")

		_ = json.NewEncoder(w).Encode(models.TLDDetails{
			TLD: models.TLD{
				Name:      "com",
				Type:      models.TLDTypeGTLD,
				Available: true,
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	details, err := client.TLDs.GetTLD(context.Background(), "com")

	require.NoError(t, err)
	assert.Equal(t, "com", details.Name)
	assert.Equal(t, models.TLDTypeGTLD, details.Type)
}

func TestTLDsService_GetPortfolio(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/tlds/portfolio")

		_ = json.NewEncoder(w).Encode(models.TLDPortfolio{
			TLDs: []models.TLD{
				{Name: "com", Type: models.TLDTypeGTLD, Available: true},
			},
			Total: 1,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	portfolio, err := client.TLDs.GetPortfolio(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 1, portfolio.Total)
	assert.Len(t, portfolio.TLDs, 1)
	assert.Equal(t, "com", portfolio.TLDs[0].Name)
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

func TestAvailabilityService_CheckSingleAvailability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/availability")

		domains := r.URL.Query()["domains"]
		assert.Contains(t, domains, "example.com")

		_ = json.NewEncoder(w).Encode(models.AvailabilityResponse{
			Results: []models.DomainAvailability{
				{Domain: "example.com", Status: models.AvailabilityStatusAvailable},
			},
			Meta: models.AvailabilityMeta{Total: 1, ProcessingTimeMs: 25},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.Availability.CheckSingleAvailability(context.Background(), "example.com")

	require.NoError(t, err)
	assert.Equal(t, "example.com", result.Domain)
	assert.Equal(t, models.AvailabilityStatusAvailable, result.Status)
}

func TestAvailabilityService_GetSuggestions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/domain-search/suggest")
		assert.Equal(t, "example", r.URL.Query().Get("query"))
		assert.Contains(t, r.URL.Query()["tlds"], "com")

		_ = json.NewEncoder(w).Encode(models.DomainSuggestResponse{
			Suggestions: []models.DomainSuggestion{
				{Domain: "example.com", Available: true, Premium: false},
			},
			Meta: models.AvailabilityMeta{Total: 1, ProcessingTimeMs: 30},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.Availability.GetSuggestions(context.Background(), "example", &models.DomainSuggestRequest{
		TLDs:  []string{"com"},
		Limit: 5,
	})

	require.NoError(t, err)
	assert.Len(t, result.Suggestions, 1)
	assert.Equal(t, "example.com", result.Suggestions[0].Domain)
	assert.True(t, result.Suggestions[0].Available)
}
