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
