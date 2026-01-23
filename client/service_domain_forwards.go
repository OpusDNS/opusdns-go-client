package client

import (
	"context"
	"net/url"
	"strconv"

	"github.com/opusdns/opusdns-go-client/models"
)

// DomainForwardsService provides methods for managing domain/URL forwarding.
type DomainForwardsService struct {
	client *Client
}

// ListDomainForwards retrieves all domain forwards with automatic pagination.
func (s *DomainForwardsService) ListDomainForwards(ctx context.Context, opts *models.ListDomainForwardsOptions) ([]models.DomainForward, error) {
	var all []models.DomainForward
	page := 1

	for {
		pageOpts := opts
		if pageOpts == nil {
			pageOpts = &models.ListDomainForwardsOptions{}
		}
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := s.ListDomainForwardsPage(ctx, pageOpts)
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

// ListDomainForwardsPage retrieves a single page of domain forwards.
func (s *DomainForwardsService) ListDomainForwardsPage(ctx context.Context, opts *models.ListDomainForwardsOptions) (*models.DomainForwardListResponse, error) {
	path := s.client.http.BuildPath("domain-forwards")

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

	var result models.DomainForwardListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetDomainForward retrieves a specific domain forward by hostname.
func (s *DomainForwardsService) GetDomainForward(ctx context.Context, hostname string) (*models.DomainForward, error) {
	path := s.client.http.BuildPath("domain-forwards", url.PathEscape(hostname))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var domainForward models.DomainForward
	if err := s.client.http.DecodeResponse(resp, &domainForward); err != nil {
		return nil, err
	}

	return &domainForward, nil
}

// CreateDomainForward creates domain forwarding for a hostname.
func (s *DomainForwardsService) CreateDomainForward(ctx context.Context, req *models.DomainForwardCreateRequest) (*models.DomainForward, error) {
	path := s.client.http.BuildPath("domain-forwards")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var domainForward models.DomainForward
	if err := s.client.http.DecodeResponse(resp, &domainForward); err != nil {
		return nil, err
	}

	return &domainForward, nil
}

// UpdateDomainForwardConfig updates the configuration for a specific protocol.
func (s *DomainForwardsService) UpdateDomainForwardConfig(ctx context.Context, hostname string, protocol models.DomainForwardProtocol, req *models.DomainForwardConfigUpdate) (*models.DomainForward, error) {
	path := s.client.http.BuildPath("domain-forwards", url.PathEscape(hostname), string(protocol))

	resp, err := s.client.http.Patch(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var domainForward models.DomainForward
	if err := s.client.http.DecodeResponse(resp, &domainForward); err != nil {
		return nil, err
	}

	return &domainForward, nil
}

// DeleteDomainForward deletes domain forwarding for a hostname.
func (s *DomainForwardsService) DeleteDomainForward(ctx context.Context, hostname string) error {
	path := s.client.http.BuildPath("domain-forwards", url.PathEscape(hostname))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// DeleteDomainForwardConfig deletes a specific protocol configuration.
func (s *DomainForwardsService) DeleteDomainForwardConfig(ctx context.Context, hostname string, protocol models.DomainForwardProtocol) error {
	path := s.client.http.BuildPath("domain-forwards", url.PathEscape(hostname), string(protocol))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// EnableDomainForward enables a domain forward.
func (s *DomainForwardsService) EnableDomainForward(ctx context.Context, hostname string) (*models.DomainForward, error) {
	path := s.client.http.BuildPath("domain-forwards", url.PathEscape(hostname), "enable")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var domainForward models.DomainForward
	if err := s.client.http.DecodeResponse(resp, &domainForward); err != nil {
		return nil, err
	}

	return &domainForward, nil
}

// DisableDomainForward disables a domain forward.
func (s *DomainForwardsService) DisableDomainForward(ctx context.Context, hostname string) (*models.DomainForward, error) {
	path := s.client.http.BuildPath("domain-forwards", url.PathEscape(hostname), "disable")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var domainForward models.DomainForward
	if err := s.client.http.DecodeResponse(resp, &domainForward); err != nil {
		return nil, err
	}

	return &domainForward, nil
}

// ListDomainForwardsByZone retrieves domain forwards for a specific DNS zone.
func (s *DomainForwardsService) ListDomainForwardsByZone(ctx context.Context, zoneName string) ([]models.DomainForward, error) {
	path := s.client.http.BuildPath("dns", url.PathEscape(zoneName), "domain-forwards")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result []models.DomainForward
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}
