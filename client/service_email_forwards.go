package client

import (
	"context"
	"net/url"
	"strconv"

	"github.com/opusdns/opusdns-go-client/models"
)

// EmailForwardsService provides methods for managing email forwarding.
type EmailForwardsService struct {
	client *Client
}

// ListEmailForwards retrieves all email forwards with automatic pagination.
func (s *EmailForwardsService) ListEmailForwards(ctx context.Context, opts *models.ListEmailForwardsOptions) ([]models.EmailForward, error) {
	var all []models.EmailForward
	page := 1

	for {
		pageOpts := opts
		if pageOpts == nil {
			pageOpts = &models.ListEmailForwardsOptions{}
		}
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := s.ListEmailForwardsPage(ctx, pageOpts)
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

// ListEmailForwardsPage retrieves a single page of email forwards.
func (s *EmailForwardsService) ListEmailForwardsPage(ctx context.Context, opts *models.ListEmailForwardsOptions) (*models.EmailForwardListResponse, error) {
	path := s.client.http.BuildPath("email-forwards")

	query := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.SortBy != "" {
			query.Set("sort_by", string(opts.SortBy))
		}
		if opts.SortOrder != "" {
			query.Set("sort_order", string(opts.SortOrder))
		}
		if opts.Search != "" {
			query.Set("search", opts.Search)
		}
		if opts.Enabled != nil {
			query.Set("enabled", strconv.FormatBool(*opts.Enabled))
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.EmailForwardListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetEmailForward retrieves a specific email forward by ID.
func (s *EmailForwardsService) GetEmailForward(ctx context.Context, emailForwardID models.EmailForwardID) (*models.EmailForward, error) {
	path := s.client.http.BuildPath("email-forwards", string(emailForwardID))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var emailForward models.EmailForward
	if err := s.client.http.DecodeResponse(resp, &emailForward); err != nil {
		return nil, err
	}

	return &emailForward, nil
}

// CreateEmailForward creates email forwarding for a hostname.
func (s *EmailForwardsService) CreateEmailForward(ctx context.Context, req *models.EmailForwardCreateRequest) (*models.EmailForward, error) {
	path := s.client.http.BuildPath("email-forwards")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var emailForward models.EmailForward
	if err := s.client.http.DecodeResponse(resp, &emailForward); err != nil {
		return nil, err
	}

	return &emailForward, nil
}

// DeleteEmailForward deletes email forwarding for a hostname.
func (s *EmailForwardsService) DeleteEmailForward(ctx context.Context, emailForwardID models.EmailForwardID) error {
	path := s.client.http.BuildPath("email-forwards", string(emailForwardID))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// EnableEmailForward enables an email forward.
func (s *EmailForwardsService) EnableEmailForward(ctx context.Context, emailForwardID models.EmailForwardID) (*models.EmailForward, error) {
	path := s.client.http.BuildPath("email-forwards", string(emailForwardID), "enable")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var emailForward models.EmailForward
	if err := s.client.http.DecodeResponse(resp, &emailForward); err != nil {
		return nil, err
	}

	return &emailForward, nil
}

// DisableEmailForward disables an email forward.
func (s *EmailForwardsService) DisableEmailForward(ctx context.Context, emailForwardID models.EmailForwardID) (*models.EmailForward, error) {
	path := s.client.http.BuildPath("email-forwards", string(emailForwardID), "disable")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var emailForward models.EmailForward
	if err := s.client.http.DecodeResponse(resp, &emailForward); err != nil {
		return nil, err
	}

	return &emailForward, nil
}

// CreateAlias creates a new email alias.
func (s *EmailForwardsService) CreateAlias(ctx context.Context, emailForwardID models.EmailForwardID, req *models.EmailForwardAliasCreate) (*models.EmailForwardAlias, error) {
	path := s.client.http.BuildPath("email-forwards", string(emailForwardID), "aliases")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var alias models.EmailForwardAlias
	if err := s.client.http.DecodeResponse(resp, &alias); err != nil {
		return nil, err
	}

	return &alias, nil
}

// UpdateAlias updates an email alias.
func (s *EmailForwardsService) UpdateAlias(ctx context.Context, emailForwardID models.EmailForwardID, aliasID models.EmailForwardAliasID, req *models.EmailForwardAliasUpdate) (*models.EmailForwardAlias, error) {
	path := s.client.http.BuildPath("email-forwards", string(emailForwardID), "aliases", string(aliasID))

	resp, err := s.client.http.Patch(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var alias models.EmailForwardAlias
	if err := s.client.http.DecodeResponse(resp, &alias); err != nil {
		return nil, err
	}

	return &alias, nil
}

// DeleteAlias deletes an email alias.
func (s *EmailForwardsService) DeleteAlias(ctx context.Context, emailForwardID models.EmailForwardID, aliasID models.EmailForwardAliasID) error {
	path := s.client.http.BuildPath("email-forwards", string(emailForwardID), "aliases", string(aliasID))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// ListEmailForwardsByZone retrieves email forwards for a specific DNS zone.
func (s *EmailForwardsService) ListEmailForwardsByZone(ctx context.Context, zoneName string) ([]models.EmailForward, error) {
	path := s.client.http.BuildPath("dns", url.PathEscape(zoneName), "email-forwards")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result []models.EmailForward
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}
