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

// --- RBAC roles ---

func TestOrganizationsService_ListRoles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/roles", r.URL.Path)

		// The endpoint returns a bare JSON array, not a wrapped object.
		_ = json.NewEncoder(w).Encode([]models.RoleDefinition{
			{Label: "admin", Name: "Admin", BuiltIn: true, Permissions: []string{"domains:manage"}},
			{Label: "support_staff", Name: "Support Staff", BuiltIn: false, Permissions: []string{"domains:read"}},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	roles, err := client.Organizations.ListRoles(context.Background())
	require.NoError(t, err)
	require.Len(t, roles, 2)
	assert.Equal(t, "admin", roles[0].Label)
	assert.True(t, roles[0].BuiltIn)
	assert.Equal(t, "support_staff", roles[1].Label)
	assert.False(t, roles[1].BuiltIn)
}

func TestOrganizationsService_GetRole(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/roles/support_staff", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.RoleDefinition{Label: "support_staff", Name: "Support Staff"})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	role, err := client.Organizations.GetRole(context.Background(), "support_staff")
	require.NoError(t, err)
	assert.Equal(t, "support_staff", role.Label)
}

func TestOrganizationsService_CreateRole(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/organizations/roles", r.URL.Path)

		var req models.CustomRoleCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "Support Staff", req.Name)
		assert.Equal(t, []string{"domains:read", "dns:manage"}, req.Permissions)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(models.RoleDefinition{Label: "support_staff", Name: req.Name, Permissions: req.Permissions})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	role, err := client.Organizations.CreateRole(context.Background(), &models.CustomRoleCreateRequest{
		Name:        "Support Staff",
		Permissions: []string{"domains:read", "dns:manage"},
	})
	require.NoError(t, err)
	assert.Equal(t, "support_staff", role.Label)
}

func TestOrganizationsService_UpdateRole(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/organizations/roles/support_staff", r.URL.Path)

		var req models.CustomRoleUpdateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		require.NotNil(t, req.Permissions)
		assert.Equal(t, []string{"domains:read"}, *req.Permissions)

		_ = json.NewEncoder(w).Encode(models.RoleDefinition{Label: "support_staff", Name: "Support Staff"})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	perms := []string{"domains:read"}
	role, err := client.Organizations.UpdateRole(context.Background(), "support_staff", &models.CustomRoleUpdateRequest{
		Permissions: &perms,
	})
	require.NoError(t, err)
	assert.Equal(t, "support_staff", role.Label)
}

func TestOrganizationsService_DeleteRole(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/organizations/roles/support_staff", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Organizations.DeleteRole(context.Background(), "support_staff")
	require.NoError(t, err)
}

func TestOrganizationsService_ListRolePermissions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/role-permissions", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.PermissionCatalogResponse{Permissions: []string{"domains:read", "domains:manage"}})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	catalog, err := client.Organizations.ListRolePermissions(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"domains:read", "domains:manage"}, catalog.Permissions)
}

func TestUsersService_GetUserRole(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/users/usr_123/role", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.RoleAssignment{Role: models.StringPtr("admin")})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	assignment, err := client.Users.GetUserRole(context.Background(), "usr_123")
	require.NoError(t, err)
	require.NotNil(t, assignment.Role)
	assert.Equal(t, "admin", *assignment.Role)
}

func TestUsersService_SetUserRole(t *testing.T) {
	t.Run("sets role", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PUT", r.Method)
			assert.Equal(t, "/v1/users/usr_123/role", r.URL.Path)

			var req models.RoleAssignmentRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			require.NotNil(t, req.Role)
			assert.Equal(t, "domain_manager", *req.Role)

			_ = json.NewEncoder(w).Encode(models.RoleAssignment{Role: models.StringPtr("domain_manager")})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		assignment, err := client.Users.SetUserRole(context.Background(), "usr_123", models.StringPtr("domain_manager"))
		require.NoError(t, err)
		assert.Equal(t, "domain_manager", *assignment.Role)
	})

	t.Run("clears role with null", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var raw map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&raw)
			val, present := raw["role"]
			assert.True(t, present)
			assert.Nil(t, val)

			_ = json.NewEncoder(w).Encode(models.RoleAssignment{Role: nil})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		assignment, err := client.Users.SetUserRole(context.Background(), "usr_123", nil)
		require.NoError(t, err)
		assert.Nil(t, assignment.Role)
	})
}

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

// --- Vanity nameservers ---

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
		_ = json.NewEncoder(w).Encode(models.VanityNameserverSetDefaultResponse{
			VanityNameserverSet: models.VanityNameserverSet{SetID: "vns_1", IsDefault: true},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.VanityNameservers.SetDefault(context.Background(), "vns_1")
	require.NoError(t, err)
	assert.True(t, result.VanityNameserverSet.IsDefault)
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/vanity-nameserver-sets/vns_1/zones", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.ZonesReferencingSetResponse{
			Results:    []models.Zone{{Name: "example.com"}},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.VanityNameservers.ListZonesReferencingSet(context.Background(), "vns_1", nil)
	require.NoError(t, err)
	require.Len(t, result.Results, 1)
	assert.Equal(t, "example.com", result.Results[0].Name)
}

// --- Host objects ---

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
