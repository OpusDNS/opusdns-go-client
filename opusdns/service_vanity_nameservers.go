package opusdns

import (
	"context"
	"net/url"
	"strconv"

	"github.com/opusdns/opusdns-go-client/models"
)

// VanityNameserversService provides methods for managing vanity nameserver sets.
type VanityNameserversService struct {
	client *Client
}

// ListSets retrieves all vanity nameserver sets with automatic pagination.
func (s *VanityNameserversService) ListSets(ctx context.Context, opts *models.ListVanityNameserverSetsOptions) ([]models.VanityNameserverSet, error) {
	var all []models.VanityNameserverSet
	page := 1

	for {
		pageOpts := models.ListVanityNameserverSetsOptions{PageSize: DefaultPageSize}
		if opts != nil {
			pageOpts = *opts
		}
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := s.ListSetsPage(ctx, &pageOpts)
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

// ListSetsPage retrieves a single page of vanity nameserver sets.
func (s *VanityNameserversService) ListSetsPage(ctx context.Context, opts *models.ListVanityNameserverSetsOptions) (*models.VanityNameserverSetListResponse, error) {
	path := s.client.http.BuildPath("vanity-nameserver-sets")

	query := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(opts.PageSize))
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.VanityNameserverSetListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSet retrieves a vanity nameserver set by ID.
func (s *VanityNameserversService) GetSet(ctx context.Context, setID models.VanityNameserverSetID) (*models.VanityNameserverSet, error) {
	path := s.client.http.BuildPath("vanity-nameserver-sets", string(setID))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var set models.VanityNameserverSet
	if err := s.client.http.DecodeResponse(resp, &set); err != nil {
		return nil, err
	}

	return &set, nil
}

// CreateSet creates a vanity nameserver set. Creation is asynchronous: the returned set
// starts with status "provisioning" until the provisioning chain finalizes it.
func (s *VanityNameserversService) CreateSet(ctx context.Context, req *models.VanityNameserverSetCreateRequest) (*models.VanityNameserverSet, error) {
	path := s.client.http.BuildPath("vanity-nameserver-sets")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var set models.VanityNameserverSet
	if err := s.client.http.DecodeResponse(resp, &set); err != nil {
		return nil, err
	}

	return &set, nil
}

// DeleteSet deletes a vanity nameserver set. Deletion is asynchronous.
func (s *VanityNameserversService) DeleteSet(ctx context.Context, setID models.VanityNameserverSetID) error {
	path := s.client.http.BuildPath("vanity-nameserver-sets", string(setID))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// CheckSet runs a read-only diagnostic on a vanity nameserver set.
func (s *VanityNameserversService) CheckSet(ctx context.Context, setID models.VanityNameserverSetID) (*models.VanityNsCheckResponse, error) {
	path := s.client.http.BuildPath("vanity-nameserver-sets", "check")

	resp, err := s.client.http.Post(ctx, path, &models.VanityNsCheckRequest{SetID: setID})
	if err != nil {
		return nil, err
	}

	var result models.VanityNsCheckResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SetDefault marks a vanity nameserver set as the organization's default.
func (s *VanityNameserversService) SetDefault(ctx context.Context, setID models.VanityNameserverSetID) (*models.VanityNameserverSetDefaultResponse, error) {
	path := s.client.http.BuildPath("vanity-nameserver-sets", string(setID), "default")

	resp, err := s.client.http.Patch(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.VanityNameserverSetDefaultResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ClearDefault unsets the organization's default vanity nameserver set.
func (s *VanityNameserversService) ClearDefault(ctx context.Context) (*models.ClearVanityNameserverSetDefaultResponse, error) {
	path := s.client.http.BuildPath("vanity-nameserver-sets", "default")

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return nil, err
	}

	var result models.ClearVanityNameserverSetDefaultResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// RestoreSet restores a suspended vanity nameserver set.
func (s *VanityNameserversService) RestoreSet(ctx context.Context, setID models.VanityNameserverSetID) (*models.VanityNameserverSet, error) {
	path := s.client.http.BuildPath("vanity-nameserver-sets", string(setID), "restore")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var set models.VanityNameserverSet
	if err := s.client.http.DecodeResponse(resp, &set); err != nil {
		return nil, err
	}

	return &set, nil
}

// ListZonesReferencingSet lists the DNS zones whose apex is branded by a vanity NS set.
func (s *VanityNameserversService) ListZonesReferencingSet(ctx context.Context, setID models.VanityNameserverSetID, opts *models.ListVanityNameserverSetsOptions) (*models.ZonesReferencingSetResponse, error) {
	path := s.client.http.BuildPath("vanity-nameserver-sets", string(setID), "zones")

	query := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(opts.PageSize))
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.ZonesReferencingSetResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
