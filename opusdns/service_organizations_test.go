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

func TestOrganizationsService_GetCurrentAttributes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/attributes", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.OrganizationAttributesResponse{})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	_, err = client.Organizations.GetCurrentAttributes(context.Background())
	require.NoError(t, err)
}

func TestOrganizationsService_UpdateCurrentAttributes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/organizations/attributes", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.OrganizationAttributesResponse{})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	_, err = client.Organizations.UpdateCurrentAttributes(context.Background(), &models.OrganizationAttributeUpdateRequest{})
	require.NoError(t, err)
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

		var req models.OrganizationUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.NotNil(t, req.Name)
		assert.Equal(t, newName, *req.Name)

		_ = json.NewEncoder(w).Encode(models.Organization{
			OrganizationID: "organization_123",
			Name:           newName,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	org, err := client.Organizations.UpdateOrganization(context.Background(), models.OrganizationID("organization_123"), &models.OrganizationUpdateRequest{
		Name: &newName,
	})
	require.NoError(t, err)
	require.NotNil(t, org)
	assert.Equal(t, newName, org.Name)
}

func TestOrganizationsService_ListIPRestrictions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/ip-restrictions", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.IPRestrictionListResponse{
			Results: []models.IPRestriction{
				{IPRestrictionID: 1, OrganizationID: "organization_123", IPNetwork: "203.0.113.0/24"},
			},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Organizations.ListIPRestrictions(context.Background())
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "203.0.113.0/24", resp.Results[0].IPNetwork)
}

func TestOrganizationsService_GetIPRestriction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/ip-restrictions/ipr_123", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.IPRestriction{
			IPRestrictionID: 1,
			OrganizationID:  "organization_123",
			IPNetwork:       "203.0.113.0/24",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	restriction, err := client.Organizations.GetIPRestriction(context.Background(), models.TypeID("ipr_123"))
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
		assert.Equal(t, "/v1/organizations/ip-restrictions/ipr_123", r.URL.Path)

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

	restriction, err := client.Organizations.UpdateIPRestriction(context.Background(), models.TypeID("ipr_123"), &models.IPRestrictionUpdateRequest{
		IPNetwork: &updated,
	})
	require.NoError(t, err)
	require.NotNil(t, restriction)
	assert.Equal(t, updated, restriction.IPNetwork)
}

func TestOrganizationsService_DeleteIPRestriction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/organizations/ip-restrictions/ipr_123", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Organizations.DeleteIPRestriction(context.Background(), models.TypeID("ipr_123"))
	require.NoError(t, err)
}

func TestOrganizationsService_GetAttributes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/attributes/organization_123", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.OrganizationAttributesResponse{
			Attributes: []models.OrganizationAttribute{
				{OrganizationAttributeID: 1, Key: "plan", Value: "reseller"},
			},
		})
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
		assert.Equal(t, "/v1/organizations/attributes/organization_123", r.URL.Path)

		var req models.OrganizationAttributeUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.Len(t, req.Attributes, 1)
		assert.Equal(t, "plan", req.Attributes[0].Key)

		_ = json.NewEncoder(w).Encode(models.OrganizationAttributesResponse{
			Attributes: []models.OrganizationAttribute{
				{OrganizationAttributeID: 1, Key: "plan", Value: "reseller"},
			},
		})
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
		assert.Equal(t, "domain", r.URL.Query().Get("product_type"))

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

	resp, err := client.Organizations.ListTransactions(context.Background(), models.OrganizationID("organization_123"), &models.ListTransactionsOptions{
		ProductType: models.BillingProductTypeDomain,
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

	resp, err := client.Organizations.ListInvoices(context.Background(), models.OrganizationID("organization_123"))
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "INV-001", resp.Results[0].Number)
}

func TestOrganizationsService_GetPricing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/organization_123/pricing/product-type/domain", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.ProductPricing{
			ProductType: "domain",
			Actions: map[string]models.PriceInfo{
				"create": {Price: "10.00", Currency: models.CurrencyUSD},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	pricing, err := client.Organizations.GetPricing(context.Background(), models.OrganizationID("organization_123"), "domain")
	require.NoError(t, err)
	require.NotNil(t, pricing)
	assert.Equal(t, "domain", pricing.ProductType)
	assert.Contains(t, pricing.Actions, "create")
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
