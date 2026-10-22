package opusdns

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Migrated from client_test.go ---

func TestOrganizationsService_ListOrganizations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations", r.URL.Path)
		assert.Equal(t, "DE", r.URL.Query().Get("country_code"))

		_ = json.NewEncoder(w).Encode(models.OrganizationListResponse{
			Results:    []models.Organization{{OrganizationID: "organization_123", Name: "Example"}},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	orgs, err := client.Organizations.ListOrganizations(context.Background(), &models.ListOrganizationsOptions{
		CountryCode: "DE",
	})

	require.NoError(t, err)
	require.Len(t, orgs, 1)
	assert.Equal(t, "Example", orgs[0].Name)
}

func TestOrganizationsService_CreateOrganization(t *testing.T) {
	password := "secret-password"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/organizations", r.URL.Path)

		var req models.OrganizationCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "Example Child", req.Name)
		require.Len(t, req.Users, 1)
		assert.Equal(t, "owner@example.com", req.Users[0].Username)
		require.NotNil(t, req.Users[0].Password)
		assert.Equal(t, password, *req.Users[0].Password)
		require.Len(t, req.Attributes, 1)
		assert.Equal(t, "plan", req.Attributes[0].Key)
		assert.Equal(t, "reseller", req.Attributes[0].Value)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(models.Organization{
			OrganizationID: "organization_123",
			Name:           "Example Child",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	org, err := client.Organizations.CreateOrganization(context.Background(), &models.OrganizationCreateRequest{
		Name: "Example Child",
		Users: []models.UserCreateRequest{
			{
				Username:  "owner@example.com",
				FirstName: "Owner",
				LastName:  "User",
				Email:     "owner@example.com",
				Locale:    "en-US",
				Password:  &password,
				UserAttributes: []models.UserAttributeBase{
					{Key: "department", Value: "sales"},
				},
			},
		},
		Attributes: []models.OrganizationAttributeCreate{
			{Key: "plan", Value: "reseller"},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, org)
	assert.Equal(t, models.OrganizationID("organization_123"), org.OrganizationID)
	assert.Equal(t, "Example Child", org.Name)
}

func TestOrganizationsService_DeleteOrganization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/organizations/organization_123", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Organizations.DeleteOrganization(context.Background(), models.OrganizationID("organization_123"))

	require.NoError(t, err)
}

// --- Migrated from gaps_test.go ---

// orgAttributesBody is a list of OrganizationAttributeResponse as the API
// sends it: a bare array, not an object.
const orgAttributesBody = `[
	{"organization_attribute_id": 1, "key": "plan", "value": "reseller", "protected": true,
	 "created_on": "2026-01-23T08:26:55Z", "updated_on": "2026-01-23T08:26:55Z"}
]`

func TestOrganizationsService_GetCurrentAttributes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/attributes", r.URL.Path)
		_, _ = w.Write([]byte(orgAttributesBody))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Organizations.GetCurrentAttributes(context.Background())
	require.NoError(t, err)
	require.Len(t, resp.Attributes, 1)
	assert.Equal(t, 1, resp.Attributes[0].OrganizationAttributeID)
	assert.Equal(t, "plan", resp.Attributes[0].Key)
	assert.Equal(t, "reseller", resp.Attributes[0].Value)
	assert.True(t, resp.Attributes[0].Protected)
}

func TestOrganizationsService_UpdateCurrentAttributes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/organizations/attributes", r.URL.Path)

		// The body is a bare array of OrganizationAttributeUpdate.
		var body []map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Len(t, body, 1)
		assert.Equal(t, "plan", body[0]["key"])
		assert.Equal(t, "reseller", body[0]["value"])

		_, _ = w.Write([]byte(orgAttributesBody))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Organizations.UpdateCurrentAttributes(context.Background(), &models.OrganizationAttributeUpdateRequest{
		Attributes: []models.OrganizationAttributeCreate{{Key: "plan", Value: "reseller"}},
	})
	require.NoError(t, err)
	require.Len(t, resp.Attributes, 1)
	assert.Equal(t, "plan", resp.Attributes[0].Key)
}

func TestOrganizationsService_UpdateCurrentAttributes_EmptySendsArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `[]`, string(body))
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Organizations.UpdateCurrentAttributes(context.Background(), &models.OrganizationAttributeUpdateRequest{})
	require.NoError(t, err)
	assert.Empty(t, resp.Attributes)

	_, err = client.Organizations.UpdateCurrentAttributes(context.Background(), nil)
	require.ErrorIs(t, err, ErrInvalidInput)
}

// --- Migrated from features_test.go (RBAC roles) ---

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

func TestOrganizationsService_CreateRole_RequiresPermissions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no request expected, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	// Without permissions the request would carry "permissions": null.
	_, err = client.Organizations.CreateRole(context.Background(), &models.CustomRoleCreateRequest{Name: "Empty"})
	require.ErrorIs(t, err, ErrInvalidInput)

	_, err = client.Organizations.CreateRole(context.Background(), nil)
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestOrganizationsService_RoleLabelIsEscaped(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	// A label must stay one path segment; it must not reach another resource.
	err = client.Organizations.DeleteRole(context.Background(), "../ip-restrictions/5")
	require.NoError(t, err)
	assert.Equal(t, "/v1/organizations/roles/..%2Fip-restrictions%2F5", gotPath)
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

// --- New coverage ---

func TestOrganizationsService_GetOrganization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/organization_123", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.Organization{
			OrganizationID: "organization_123",
			Name:           "Example",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	org, err := client.Organizations.GetOrganization(context.Background(), models.OrganizationID("organization_123"))
	require.NoError(t, err)
	require.NotNil(t, org)
	assert.Equal(t, models.OrganizationID("organization_123"), org.OrganizationID)
	assert.Equal(t, "Example", org.Name)
}

func TestOrganizationsService_UpdateOrganization(t *testing.T) {
	newName := "Renamed Org"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/organizations/organization_123", r.URL.Path)

		// Only OrganizationUpdate properties go out: tax_rate does, the
		// deprecated currency does not.
		var body map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, map[string]interface{}{"name": newName, "tax_rate": "0.19"}, body)

		_ = json.NewEncoder(w).Encode(models.Organization{
			OrganizationID: "organization_123",
			Name:           newName,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	eur := models.CurrencyEUR
	org, err := client.Organizations.UpdateOrganization(context.Background(), models.OrganizationID("organization_123"), &models.OrganizationUpdateRequest{
		Name:     &newName,
		TaxRate:  models.StringPtr("0.19"),
		Currency: &eur, //nolint:staticcheck // deprecated on purpose: proves it is not sent
	})
	require.NoError(t, err)
	require.NotNil(t, org)
	assert.Equal(t, newName, org.Name)
}

func TestOrganizationsService_ListIPRestrictions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/ip-restrictions", r.URL.Path)
		// A bare array of IpRestrictionResponse, with no pagination.
		_, _ = w.Write([]byte(`[
			{"ip_restriction_id": 1, "organization_id": "organization_01h45ytscbebyvny4gc8cr8ma2",
			 "ip_network": "203.0.113.0/24", "last_used_on": null, "created_on": "2026-01-23T08:26:55Z"},
			{"ip_restriction_id": 2, "organization_id": "organization_01h45ytscbebyvny4gc8cr8ma2",
			 "ip_network": "2001:db8::/32", "last_used_on": "2026-02-01T10:00:00Z", "created_on": "2026-01-24T08:26:55Z"}
		]`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Organizations.ListIPRestrictions(context.Background())
	require.NoError(t, err)
	require.Len(t, resp.Results, 2)
	assert.Equal(t, 1, resp.Results[0].IPRestrictionID)
	assert.Equal(t, "203.0.113.0/24", resp.Results[0].IPNetwork)
	assert.Nil(t, resp.Results[0].LastUsedOn)
	require.NotNil(t, resp.Results[1].LastUsedOn)
}

func TestOrganizationsService_GetIPRestriction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/ip-restrictions/123", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.IPRestriction{
			IPRestrictionID: 1,
			OrganizationID:  "organization_123",
			IPNetwork:       "203.0.113.0/24",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	restriction, err := client.Organizations.GetIPRestriction(context.Background(), 123)
	require.NoError(t, err)
	require.NotNil(t, restriction)
	assert.Equal(t, "203.0.113.0/24", restriction.IPNetwork)
}

func TestOrganizationsService_CreateIPRestriction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/organizations/ip-restrictions", r.URL.Path)

		var req models.IPRestrictionCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "203.0.113.0/24", req.IPNetwork)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(models.IPRestriction{
			IPRestrictionID: 1,
			OrganizationID:  "organization_123",
			IPNetwork:       req.IPNetwork,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	restriction, err := client.Organizations.CreateIPRestriction(context.Background(), &models.IPRestrictionCreateRequest{
		IPNetwork: "203.0.113.0/24",
	})
	require.NoError(t, err)
	require.NotNil(t, restriction)
	assert.Equal(t, "203.0.113.0/24", restriction.IPNetwork)
}

func TestOrganizationsService_UpdateIPRestriction(t *testing.T) {
	updated := "198.51.100.0/24"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/organizations/ip-restrictions/123", r.URL.Path)

		var req models.IPRestrictionUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.NotNil(t, req.IPNetwork)
		assert.Equal(t, updated, *req.IPNetwork)

		_ = json.NewEncoder(w).Encode(models.IPRestriction{
			IPRestrictionID: 1,
			OrganizationID:  "organization_123",
			IPNetwork:       updated,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	restriction, err := client.Organizations.UpdateIPRestriction(context.Background(), 123, &models.IPRestrictionUpdateRequest{
		IPNetwork: &updated,
	})
	require.NoError(t, err)
	require.NotNil(t, restriction)
	assert.Equal(t, updated, restriction.IPNetwork)
}

func TestOrganizationsService_DeleteIPRestriction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/organizations/ip-restrictions/123", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Organizations.DeleteIPRestriction(context.Background(), 123)
	require.NoError(t, err)
}

func TestOrganizationsService_GetAttributes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/organization_123/attributes", r.URL.Path)
		_, _ = w.Write([]byte(orgAttributesBody))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Organizations.GetAttributes(context.Background(), models.OrganizationID("organization_123"))
	require.NoError(t, err)
	require.Len(t, resp.Attributes, 1)
	assert.Equal(t, "plan", resp.Attributes[0].Key)
}

func TestOrganizationsService_UpdateAttributes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/organizations/organization_123/attributes", r.URL.Path)

		// The body is a bare array of OrganizationAttributeUpdate.
		var body []models.OrganizationAttributeCreate
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Len(t, body, 1)
		assert.Equal(t, "plan", body[0].Key)

		_, _ = w.Write([]byte(orgAttributesBody))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Organizations.UpdateAttributes(context.Background(), models.OrganizationID("organization_123"), &models.OrganizationAttributeUpdateRequest{
		Attributes: []models.OrganizationAttributeCreate{
			{Key: "plan", Value: "reseller"},
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Attributes, 1)
	assert.Equal(t, "plan", resp.Attributes[0].Key)
}

func TestOrganizationsService_ListTransactions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/organization_123/transactions", r.URL.Path)
		query := r.URL.Query()
		assert.Equal(t, "domain", query.Get("product_type"))
		assert.Equal(t, "example.com", query.Get("product_reference"))
		assert.Equal(t, "2026-01-02T03:04:05.123456789Z", query.Get("created_after"))
		assert.Equal(t, "2026-02-01T00:00:00Z", query.Get("created_before"))
		assert.Equal(t, "2026-01-05T00:00:00.5Z", query.Get("completed_after"))
		assert.Equal(t, "2026-01-31T00:00:00Z", query.Get("completed_before"))

		_ = json.NewEncoder(w).Encode(models.BillingTransactionListResponse{
			Results: []models.BillingTransaction{
				{
					BillingTransactionID: "txn_1",
					ProductType:          models.BillingProductTypeDomain,
					Action:               models.BillingActionCreate,
					Status:               models.BillingStatusSucceeded,
					Amount:               "10.00",
					Currency:             models.CurrencyUSD,
				},
			},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	createdAfter := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	createdBefore := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	completedAfter := time.Date(2026, 1, 5, 0, 0, 0, 500000000, time.UTC)
	completedBefore := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	resp, err := client.Organizations.ListTransactions(context.Background(), models.OrganizationID("organization_123"), &models.ListTransactionsOptions{
		ProductType:      models.BillingProductTypeDomain,
		ProductReference: "example.com",
		CreatedAfter:     &createdAfter,
		CreatedBefore:    &createdBefore,
		CompletedAfter:   &completedAfter,
		CompletedBefore:  &completedBefore,
	})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, models.BillingTransactionID("txn_1"), resp.Results[0].BillingTransactionID)
}

func TestOrganizationsService_GetTransaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/organization_123/transactions/txn_1", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.BillingTransaction{
			BillingTransactionID: "txn_1",
			ProductType:          models.BillingProductTypeDomain,
			Action:               models.BillingActionCreate,
			Status:               models.BillingStatusSucceeded,
			Amount:               "10.00",
			Currency:             models.CurrencyUSD,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	transaction, err := client.Organizations.GetTransaction(context.Background(), models.OrganizationID("organization_123"), models.BillingTransactionID("txn_1"))
	require.NoError(t, err)
	require.NotNil(t, transaction)
	assert.Equal(t, models.BillingTransactionID("txn_1"), transaction.BillingTransactionID)
}

func TestOrganizationsService_ListInvoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/organization_123/billing/invoices", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, strconv.Itoa(DefaultPageSize), r.URL.Query().Get("page_size"))
		_ = json.NewEncoder(w).Encode(models.InvoiceListResponse{
			Results: []models.Invoice{
				{Number: "INV-001", Status: models.InvoiceStatusFinalized, Amount: "10.00", Currency: models.CurrencyUSD},
			},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	invoices, err := client.Organizations.ListInvoices(context.Background(), models.OrganizationID("organization_123"), nil)
	require.NoError(t, err)
	require.Len(t, invoices, 1)
	assert.Equal(t, "INV-001", invoices[0].Number)
}

func TestOrganizationsService_GetPricing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/organization_123/pricing/product-type/domain", r.URL.Path)
		assert.Equal(t, "create", r.URL.Query().Get("product_action"))
		assert.Equal(t, "com", r.URL.Query().Get("product_class"))
		// GetPricesResponse
		_, _ = w.Write([]byte(`{"prices": [
			{"product_type": "domain", "product_action": "create", "product_class": "com",
			 "price": "10.00", "currency": "USD", "period": {"value": 1, "unit": "y"}},
			{"product_type": "domain", "product_action": "create", "product_class": "com",
			 "price": "20.00", "currency": "USD", "period": {"value": 2, "unit": "y"}}
		]}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	pricing, err := client.Organizations.GetPricing(context.Background(), models.OrganizationID("organization_123"), "domain", &models.PricingOptions{
		ProductAction: models.BillingActionCreate,
		ProductClass:  "com",
	})
	require.NoError(t, err)
	require.Len(t, pricing.Prices, 2)
	price := pricing.Prices[0]
	assert.Equal(t, "domain", price.ProductType)
	assert.Equal(t, "10.00", price.Price)
	assert.Equal(t, models.CurrencyUSD, price.Currency)
	require.NotNil(t, price.ProductAction)
	assert.Equal(t, "create", *price.ProductAction)
	require.NotNil(t, price.Period)
	assert.Equal(t, models.PricingPeriod{Value: 1, Unit: models.PeriodUnitYear}, *price.Period)
}

func TestOrganizationsService_GetPricing_NoOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery)
		_, _ = w.Write([]byte(`{"prices": []}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	pricing, err := client.Organizations.GetPricing(context.Background(), models.OrganizationID("organization_123"), "domain", nil)
	require.NoError(t, err)
	assert.Empty(t, pricing.Prices)
}

func TestOrganizationsService_ListOrganizationsPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "50", r.URL.Query().Get("page_size"))

		_ = json.NewEncoder(w).Encode(models.OrganizationListResponse{
			Results:    []models.Organization{{OrganizationID: "organization_123", Name: "Example"}},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Organizations.ListOrganizationsPage(context.Background(), &models.ListOrganizationsOptions{
		Page:     2,
		PageSize: 50,
	})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "Example", resp.Results[0].Name)
}

func TestOrganizationsService_ListReceipts(t *testing.T) {
	// The endpoint defaults to ten receipts per page, so the wrapper has to
	// ask for a page size and keep following pages.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/organization_123/billing/receipts", r.URL.Path)
		assert.Equal(t, strconv.Itoa(DefaultPageSize), r.URL.Query().Get("page_size"))

		page := r.URL.Query().Get("page")
		_ = json.NewEncoder(w).Encode(models.InvoiceListResponse{
			Results: []models.Invoice{{
				ExternalID:   "inv_" + page,
				Number:       "R-2026-00" + page,
				DocumentType: models.InvoiceDocumentTypeReceipt,
			}},
			Pagination: models.Pagination{HasNextPage: page == "1"},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	receipts, err := client.Organizations.ListReceipts(context.Background(), "organization_123", nil)
	require.NoError(t, err)
	require.Len(t, receipts, 2)
	assert.Equal(t, models.InvoiceDocumentTypeReceipt, receipts[0].DocumentType)
	assert.Equal(t, "R-2026-001", receipts[0].Number)
	assert.Equal(t, "R-2026-002", receipts[1].Number)
}

func TestOrganizationsService_ListReceiptsPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/organizations/organization_123/billing/receipts", r.URL.Path)
		assert.Equal(t, "3", r.URL.Query().Get("page"))
		assert.Equal(t, "100", r.URL.Query().Get("page_size"))

		_ = json.NewEncoder(w).Encode(models.InvoiceListResponse{
			Results:    []models.Invoice{{Number: "R-2026-021"}},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Organizations.ListReceiptsPage(context.Background(), "organization_123", &models.ListInvoicesOptions{
		Page:     3,
		PageSize: 100,
	})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "R-2026-021", resp.Results[0].Number)
}

func TestOrganizationsService_GetUsageSeries(t *testing.T) {
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/organization_123/usage/ai_inference", r.URL.Path)

		query := r.URL.Query()
		assert.Equal(t, "2026-08-01", query.Get("start_date"))
		assert.Equal(t, "2026-08-31", query.Get("end_date"))
		assert.Equal(t, "day", query.Get("granularity"))

		_ = json.NewEncoder(w).Encode(models.AIInferenceUsageSeries{
			Product:     models.UsageProductAIInference,
			Granularity: models.UsageGranularityDay,
			StartDate:   "2026-08-01",
			EndDate:     "2026-08-31",
			Buckets: []models.AIInferenceUsageBucket{{
				PeriodStart: "2026-08-01",
				Groups: []models.AIInferenceUsageGroup{{
					Model: "claude", InputTokens: 10, OutputTokens: 5, RequestCount: 2,
				}},
			}},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	series, err := client.Organizations.GetUsageSeries(context.Background(), "organization_123",
		models.UsageProductAIInference, &models.UsageOptions{
			StartDate:   &start,
			EndDate:     &end,
			Granularity: models.UsageGranularityDay,
		})
	require.NoError(t, err)
	require.Len(t, series.Buckets, 1)
	assert.Equal(t, int64(10), series.Buckets[0].Groups[0].InputTokens)
}

func TestOrganizationsService_GetUsageSummary(t *testing.T) {
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/organizations/organization_123/usage/ai_inference/summary", r.URL.Path)

		query := r.URL.Query()
		assert.Equal(t, "2026-08-01", query.Get("start_date"))
		assert.Equal(t, "2026-08-31", query.Get("end_date"))
		// The summary endpoint does not take a granularity.
		assert.Empty(t, query.Get("granularity"))

		_ = json.NewEncoder(w).Encode(models.AIInferenceUsageSummary{
			Product:   models.UsageProductAIInference,
			StartDate: "2026-08-01",
			EndDate:   "2026-08-31",
			Groups: []models.AIInferenceUsageGroup{{
				Model: "claude", InputTokens: 100, RequestCount: 4,
			}},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	summary, err := client.Organizations.GetUsageSummary(context.Background(), "organization_123",
		models.UsageProductAIInference, &models.UsageOptions{
			StartDate:   &start,
			EndDate:     &end,
			Granularity: models.UsageGranularityDay,
		})
	require.NoError(t, err)
	require.Len(t, summary.Groups, 1)
	assert.Equal(t, int64(100), summary.Groups[0].InputTokens)
}

func TestOrganizationsService_Usage_RequiresDateRange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no request expected, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	ctx := context.Background()

	// start_date and end_date are required query parameters.
	_, err = client.Organizations.GetUsageSeries(ctx, "organization_123", models.UsageProductAIInference, nil)
	require.ErrorIs(t, err, ErrInvalidInput)

	_, err = client.Organizations.GetUsageSummary(ctx, "organization_123", models.UsageProductAIInference, nil)
	require.ErrorIs(t, err, ErrInvalidInput)

	_, err = client.Organizations.GetUsageSummary(ctx, "organization_123", models.UsageProductAIInference,
		&models.UsageOptions{StartDate: &start})
	require.ErrorIs(t, err, ErrInvalidInput)
}
