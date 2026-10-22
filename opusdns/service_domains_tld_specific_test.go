package opusdns

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomainsService_RequestAuthCode(t *testing.T) {
	for _, tld := range []models.AuthCodeTLD{
		models.AuthCodeTLDBE,
		models.AuthCodeTLDCymru,
		models.AuthCodeTLDCZ,
		models.AuthCodeTLDDK,
		models.AuthCodeTLDEU,
		models.AuthCodeTLDLT,
		models.AuthCodeTLDNU,
		models.AuthCodeTLDSE,
		models.AuthCodeTLDWales,
	} {
		t.Run(string(tld), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/domains/tld-specific/"+string(tld)+"/example."+string(tld)+"/auth_code/request", r.URL.Path)

				_ = json.NewEncoder(w).Encode(models.RequestAuthCodeResponse{
					Name:    "example." + string(tld),
					Success: true,
				})
			}))
			defer server.Close()

			client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
			require.NoError(t, err)

			result, err := client.Domains.RequestAuthCode(context.Background(), tld, "example."+string(tld))
			require.NoError(t, err)
			assert.True(t, result.Success)
		})
	}
}

func TestDomainsService_RequestAuthCodeFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(models.RequestAuthCodeResponse{
			Name:    "example.be",
			Success: false,
			Detail:  models.StringPtr("registrant email bounced"),
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.Domains.RequestAuthCode(context.Background(), models.AuthCodeTLDBE, "example.be")
	require.NoError(t, err)
	assert.False(t, result.Success)
	require.NotNil(t, result.Detail)
	assert.Equal(t, "registrant email bounced", *result.Detail)
}

func TestDomainsService_WithdrawATDomain(t *testing.T) {
	// zone_delete is a required boolean in the spec's DomainWithdrawRequest,
	// so false must be sent, not omitted.
	for _, zoneDelete := range []bool{true, false} {
		t.Run(fmt.Sprintf("zone_delete=%t", zoneDelete), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/domains/tld-specific/at/example.at/withdraw", r.URL.Path)

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				assert.JSONEq(t, fmt.Sprintf(`{"zone_delete":%t}`, zoneDelete), string(body))

				_, _ = w.Write([]byte(`{"name":"example.at","success":true,"detail":null}`))
			}))
			defer server.Close()

			client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
			require.NoError(t, err)

			result, err := client.Domains.WithdrawATDomain(context.Background(), "example.at",
				&models.DomainWithdrawRequest{ZoneDelete: zoneDelete})
			require.NoError(t, err)
			assert.True(t, result.Success)
		})
	}

	t.Run("nil request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		_, err = client.Domains.WithdrawATDomain(context.Background(), "example.at", nil)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})
}

func TestDomainsService_TransitDEDomain(t *testing.T) {
	// disconnect is a required boolean in the spec's DomainTransitRequest, so
	// false must be sent, not omitted.
	for _, disconnect := range []bool{true, false} {
		t.Run(fmt.Sprintf("disconnect=%t", disconnect), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/domains/tld-specific/de/example.de/transit", r.URL.Path)

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				assert.JSONEq(t, fmt.Sprintf(`{"disconnect":%t}`, disconnect), string(body))

				_, _ = w.Write([]byte(`{"name":"example.de","success":true}`))
			}))
			defer server.Close()

			client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
			require.NoError(t, err)

			result, err := client.Domains.TransitDEDomain(context.Background(), "example.de",
				&models.DomainTransitRequest{Disconnect: disconnect})
			require.NoError(t, err)
			assert.True(t, result.Success)
		})
	}

	t.Run("nil request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		_, err = client.Domains.TransitDEDomain(context.Background(), "example.de", nil)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})
}

func TestDomainsService_SubmitNorIDDeclaration(t *testing.T) {
	signedAt := time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/domains/tld-specific/no/example.no/applicant-declaration", r.URL.Path)

		var body models.NorIDDeclarationRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "Kari Nordmann", body.AcceptName)
		require.NotNil(t, body.AcceptDate)
		assert.True(t, signedAt.Equal(*body.AcceptDate))

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(models.NorIDDeclarationResponse{
			DomainName:         "example.no",
			Status:             models.NorIDDeclarationStatusConfirmed,
			DeclarationVersion: "1.0",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.Domains.SubmitNorIDDeclaration(context.Background(), "example.no",
		&models.NorIDDeclarationRequest{AcceptName: "Kari Nordmann", AcceptDate: &signedAt})
	require.NoError(t, err)
	assert.Equal(t, models.NorIDDeclarationStatusConfirmed, result.Status)
}

func TestDomainsService_ResendNorIDDeclarationEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/domains/tld-specific/no/example.no/resend-declaration-email", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	require.NoError(t, client.Domains.ResendNorIDDeclarationEmail(context.Background(), "example.no"))
}
