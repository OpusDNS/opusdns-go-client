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

func TestHostsService_CreateHost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/hosts", r.URL.Path)

		var req models.HostCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "ns1.example.com", req.Hostname)
		assert.Equal(t, []string{"192.0.2.53", "2001:db8::53"}, req.IPAddresses)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(models.Host{HostID: "host_1", Hostname: req.Hostname, IPAddresses: req.IPAddresses})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	host, err := client.Hosts.CreateHost(context.Background(), &models.HostCreateRequest{
		Hostname:    "ns1.example.com",
		IPAddresses: []string{"192.0.2.53", "2001:db8::53"},
	})
	require.NoError(t, err)
	assert.Equal(t, models.HostID("host_1"), host.HostID)
}

func TestHostsService_GetHost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		// Reference may be a hostname; ensure it is escaped into the path.
		assert.Equal(t, "/v1/hosts/ns1.example.com", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.Host{HostID: "host_1", Hostname: "ns1.example.com"})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	host, err := client.Hosts.GetHost(context.Background(), "ns1.example.com")
	require.NoError(t, err)
	assert.Equal(t, "ns1.example.com", host.Hostname)
}

func TestHostsService_UpdateHost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/hosts/host_1", r.URL.Path)

		var req models.HostUpdateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, []string{"198.51.100.53"}, req.IPAddresses)

		_ = json.NewEncoder(w).Encode(models.Host{HostID: "host_1", Hostname: "ns1.example.com", IPAddresses: req.IPAddresses})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	host, err := client.Hosts.UpdateHost(context.Background(), "host_1", &models.HostUpdateRequest{
		IPAddresses: []string{"198.51.100.53"},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"198.51.100.53"}, host.IPAddresses)
}

func TestHostsService_DeleteHost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/hosts/host_1", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Hosts.DeleteHost(context.Background(), "host_1")
	require.NoError(t, err)
}
