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

func TestAuthService_IntrospectAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/auth/client_credentials/introspect", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.OrganizationCredential{
			APIKeyID: "apikey_1",
			Role:     models.StringPtr("admin"),
			Status:   models.OrganizationCredentialStatusActive,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	cred, err := client.Auth.IntrospectAPIKey(context.Background())
	require.NoError(t, err)
	require.NotNil(t, cred.Role)
	assert.Equal(t, "admin", *cred.Role)
	assert.Equal(t, models.OrganizationCredentialStatusActive, cred.Status)
}

func TestAuthService_IssueToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/auth/token", r.URL.Path)

		var body models.TokenRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, models.GrantTypeClientCredentials, body.GrantType)
		assert.Equal(t, models.OrganizationID("organization_123"), body.ClientID)
		assert.Equal(t, "secret", body.ClientSecret)

		_ = json.NewEncoder(w).Encode(models.TokenResponse{
			AccessToken:      "access",
			TokenType:        "Bearer",
			ExpiresIn:        3600,
			RefreshToken:     "refresh",
			RefreshExpiresIn: 86400,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	token, err := client.Auth.IssueToken(context.Background(), "organization_123", "secret")
	require.NoError(t, err)
	assert.Equal(t, "access", token.AccessToken)
	assert.Equal(t, 3600, token.ExpiresIn)
}
