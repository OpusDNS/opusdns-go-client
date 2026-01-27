package opusdns

import (
	"context"
	"net/url"
	"strconv"

	"github.com/opusdns/opusdns-go-client/models"
)

// UsersService provides methods for managing users.
type UsersService struct {
	client *Client
}

// GetCurrentUser retrieves the currently authenticated user.
func (s *UsersService) GetCurrentUser(ctx context.Context) (*models.User, error) {
	path := s.client.http.BuildPath("users", "me")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := s.client.http.DecodeResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// ListUsers retrieves all users with automatic pagination.
func (s *UsersService) ListUsers(ctx context.Context, opts *models.ListUsersOptions) ([]models.User, error) {
	var all []models.User
	page := 1

	for {
		pageOpts := opts
		if pageOpts == nil {
			pageOpts = &models.ListUsersOptions{}
		}
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := s.ListUsersPage(ctx, pageOpts)
		if err != nil {
			return nil, err
		}

		all = append(all, resp.Results...)

		if !resp.Pagination.HasNextPage {
			break
		}
		page++
	}

	return all, nil
}

// ListUsersPage retrieves a single page of users.
func (s *UsersService) ListUsersPage(ctx context.Context, opts *models.ListUsersOptions) (*models.UserListResponse, error) {
	path := s.client.http.BuildPath("organizations", "users")

	query := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.SortBy != "" {
			query.Set("sort_by", opts.SortBy)
		}
		if opts.SortOrder != "" {
			query.Set("sort_order", string(opts.SortOrder))
		}
		if opts.Search != "" {
			query.Set("search", opts.Search)
		}
		if opts.Email != "" {
			query.Set("email", opts.Email)
		}
		if opts.Username != "" {
			query.Set("username", opts.Username)
		}
		if opts.Status != "" {
			query.Set("status", string(opts.Status))
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.UserListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUser retrieves a specific user by ID.
func (s *UsersService) GetUser(ctx context.Context, userID models.UserID) (*models.User, error) {
	path := s.client.http.BuildPath("users", string(userID))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := s.client.http.DecodeResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser creates a new user.
func (s *UsersService) CreateUser(ctx context.Context, req *models.UserCreateRequest) (*models.User, error) {
	path := s.client.http.BuildPath("users")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := s.client.http.DecodeResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates a user.
func (s *UsersService) UpdateUser(ctx context.Context, userID models.UserID, req *models.UserUpdateRequest) (*models.User, error) {
	path := s.client.http.BuildPath("users", string(userID))

	resp, err := s.client.http.Patch(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := s.client.http.DecodeResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// DeleteUser deletes a user.
func (s *UsersService) DeleteUser(ctx context.Context, userID models.UserID) error {
	path := s.client.http.BuildPath("users", string(userID))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}
