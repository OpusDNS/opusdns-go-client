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

func TestUsersService_GetCurrentUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/users/me", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.User{
			UserID: "user_123",
			Email:  "user@example.com",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	user, err := client.Users.GetCurrentUser(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "user@example.com", user.Email)
}

func TestUsersService_ListUsersOmitsUnsupportedFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		assert.Equal(t, "sam", query.Get("search"))
		assert.Empty(t, query.Get("email"))
		assert.Empty(t, query.Get("username"))
		assert.Empty(t, query.Get("status"))

		_ = json.NewEncoder(w).Encode(models.UserListResponse{
			Results:    []models.User{},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	_, err = client.Users.ListUsers(context.Background(), &models.ListUsersOptions{
		Search:   "sam",
		Email:    "sam@example.com",
		Username: "sam",
		Status:   models.UserStatusActive,
	})

	require.NoError(t, err)
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

func TestUsersService_GetUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/users/usr_123", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.User{
			UserID: "usr_123",
			Email:  "user@example.com",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	user, err := client.Users.GetUser(context.Background(), "usr_123")
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", user.Email)
}

func TestUsersService_GetUserWithAttributes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/users/usr_123", r.URL.Path)
		assert.Equal(t, []string{"phone", "locale"}, r.URL.Query()["attributes"])

		_ = json.NewEncoder(w).Encode(models.User{
			UserID: "usr_123",
			Email:  "user@example.com",
			UserAttributes: map[string]interface{}{
				"phone": "+1234567890",
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	user, err := client.Users.GetUserWithAttributes(context.Background(), "usr_123", []string{"phone", "locale"})
	require.NoError(t, err)
	assert.Equal(t, "+1234567890", user.UserAttributes["phone"])
}

func TestUsersService_GetUserPermissions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/users/usr_123/permissions", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.PermissionSet{
			Permissions: []models.Permission{"domains:read", "domains:write"},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	perms, err := client.Users.GetUserPermissions(context.Background(), "usr_123")
	require.NoError(t, err)
	require.Len(t, perms.Permissions, 2)
	assert.Equal(t, models.Permission("domains:read"), perms.Permissions[0])
}

func TestUsersService_CreateUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/users", r.URL.Path)

		var req models.UserCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "newuser", req.Username)
		assert.Equal(t, "new@example.com", req.Email)

		_ = json.NewEncoder(w).Encode(models.User{
			UserID:   "usr_456",
			Username: req.Username,
			Email:    req.Email,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	user, err := client.Users.CreateUser(context.Background(), &models.UserCreateRequest{
		Username: "newuser",
		Email:    "new@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, models.UserID("usr_456"), user.UserID)
	assert.Equal(t, "new@example.com", user.Email)
}

func TestUsersService_UpdateUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/users/usr_123", r.URL.Path)

		var req models.UserUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.NotNil(t, req.FirstName)
		assert.Equal(t, "Jane", *req.FirstName)

		_ = json.NewEncoder(w).Encode(models.User{
			UserID:    "usr_123",
			FirstName: "Jane",
			Email:     "user@example.com",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	user, err := client.Users.UpdateUser(context.Background(), "usr_123", &models.UserUpdateRequest{
		FirstName: models.StringPtr("Jane"),
	})
	require.NoError(t, err)
	assert.Equal(t, "Jane", user.FirstName)
}

func TestUsersService_DeleteUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/users/usr_123", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Users.DeleteUser(context.Background(), "usr_123")
	require.NoError(t, err)
}

func TestUsersService_ListUsersPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/users", r.URL.Path)

		query := r.URL.Query()
		assert.Equal(t, "2", query.Get("page"))
		assert.Equal(t, "50", query.Get("page_size"))
		assert.Equal(t, "email", query.Get("sort_by"))
		assert.Equal(t, "asc", query.Get("sort_order"))
		assert.Equal(t, "sam", query.Get("search"))

		_ = json.NewEncoder(w).Encode(models.UserListResponse{
			Results:    []models.User{{UserID: "usr_123", Email: "user@example.com"}},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Users.ListUsersPage(context.Background(), &models.ListUsersOptions{
		Page:      2,
		PageSize:  50,
		SortBy:    models.UserSortByEmail,
		SortOrder: models.SortAsc,
		Search:    "sam",
	})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "user@example.com", resp.Results[0].Email)
}
