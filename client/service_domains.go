package client

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/opusdns/opusdns-go-client/models"
)

// DomainsService provides methods for managing domain registrations.
type DomainsService struct {
	client *Client
}

// ListDomains retrieves all domains with automatic pagination.
func (s *DomainsService) ListDomains(ctx context.Context, opts *models.ListDomainsOptions) ([]models.Domain, error) {
	var allDomains []models.Domain
	page := 1

	for {
		pageOpts := opts
		if pageOpts == nil {
			pageOpts = &models.ListDomainsOptions{}
		}
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := s.ListDomainsPage(ctx, pageOpts)
		if err != nil {
			return nil, err
		}

		allDomains = append(allDomains, resp.Results...)

		if !resp.Pagination.HasNextPage {
			break
		}
		page++
	}

	return allDomains, nil
}

// ListDomainsPage retrieves a single page of domains.
func (s *DomainsService) ListDomainsPage(ctx context.Context, opts *models.ListDomainsOptions) (*models.DomainListResponse, error) {
	path := s.client.http.BuildPath("domains")

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
		if opts.Name != "" {
			query.Set("name", opts.Name)
		}
		if opts.TLD != "" {
			query.Set("tld", opts.TLD)
		}
		if opts.SLD != "" {
			query.Set("sld", opts.SLD)
		}
		if opts.TransferLock != nil {
			query.Set("transfer_lock", strconv.FormatBool(*opts.TransferLock))
		}
		if opts.AutoRenew != nil {
			query.Set("auto_renew", strconv.FormatBool(*opts.AutoRenew))
		}
		if opts.ExpiresAfter != nil {
			query.Set("expires_after", opts.ExpiresAfter.Format(time.RFC3339))
		}
		if opts.ExpiresBefore != nil {
			query.Set("expires_before", opts.ExpiresBefore.Format(time.RFC3339))
		}
		if opts.Status != "" {
			query.Set("status", string(opts.Status))
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.DomainListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetDomain retrieves a specific domain by ID or name.
func (s *DomainsService) GetDomain(ctx context.Context, domainRef string) (*models.Domain, error) {
	path := s.client.http.BuildPath("domains", url.PathEscape(domainRef))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var domain models.Domain
	if err := s.client.http.DecodeResponse(resp, &domain); err != nil {
		return nil, err
	}

	return &domain, nil
}

// CreateDomain registers a new domain.
func (s *DomainsService) CreateDomain(ctx context.Context, req *models.DomainCreateRequest) (*models.Domain, error) {
	path := s.client.http.BuildPath("domains")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var domain models.Domain
	if err := s.client.http.DecodeResponse(resp, &domain); err != nil {
		return nil, err
	}

	return &domain, nil
}

// UpdateDomain updates a domain's configuration.
func (s *DomainsService) UpdateDomain(ctx context.Context, domainRef string, req *models.DomainUpdateRequest) (*models.Domain, error) {
	path := s.client.http.BuildPath("domains", url.PathEscape(domainRef))

	resp, err := s.client.http.Patch(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var domain models.Domain
	if err := s.client.http.DecodeResponse(resp, &domain); err != nil {
		return nil, err
	}

	return &domain, nil
}

// DeleteDomain deletes/cancels a domain registration.
func (s *DomainsService) DeleteDomain(ctx context.Context, domainRef string) error {
	path := s.client.http.BuildPath("domains", url.PathEscape(domainRef))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// TransferDomain initiates a domain transfer.
func (s *DomainsService) TransferDomain(ctx context.Context, req *models.DomainTransferRequest) (*models.Domain, error) {
	path := s.client.http.BuildPath("domains", "transfer")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var domain models.Domain
	if err := s.client.http.DecodeResponse(resp, &domain); err != nil {
		return nil, err
	}

	return &domain, nil
}

// RenewDomain renews a domain registration.
func (s *DomainsService) RenewDomain(ctx context.Context, domainRef string, req *models.DomainRenewRequest) (*models.Domain, error) {
	path := s.client.http.BuildPath("domains", url.PathEscape(domainRef), "renew")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var domain models.Domain
	if err := s.client.http.DecodeResponse(resp, &domain); err != nil {
		return nil, err
	}

	return &domain, nil
}

// RestoreDomain restores a deleted domain from redemption.
func (s *DomainsService) RestoreDomain(ctx context.Context, domainRef string, req *models.DomainRestoreRequest) (*models.Domain, error) {
	path := s.client.http.BuildPath("domains", url.PathEscape(domainRef), "restore")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var domain models.Domain
	if err := s.client.http.DecodeResponse(resp, &domain); err != nil {
		return nil, err
	}

	return &domain, nil
}

// GetSummary retrieves a summary of domains.
func (s *DomainsService) GetSummary(ctx context.Context) (*models.DomainSummary, error) {
	path := s.client.http.BuildPath("domains", "summary")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var summary models.DomainSummary
	if err := s.client.http.DecodeResponse(resp, &summary); err != nil {
		return nil, err
	}

	return &summary, nil
}

// GetDNSSEC retrieves DNSSEC information for a domain.
func (s *DomainsService) GetDNSSEC(ctx context.Context, domainRef string) (*models.DomainDNSSECRequest, error) {
	path := s.client.http.BuildPath("domains", url.PathEscape(domainRef), "dnssec")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var dnssec models.DomainDNSSECRequest
	if err := s.client.http.DecodeResponse(resp, &dnssec); err != nil {
		return nil, err
	}

	return &dnssec, nil
}

// EnableDNSSEC enables DNSSEC for a domain at the registry.
func (s *DomainsService) EnableDNSSEC(ctx context.Context, domainRef string, req *models.DomainDNSSECRequest) (*models.Domain, error) {
	path := s.client.http.BuildPath("domains", url.PathEscape(domainRef), "dnssec", "enable")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var domain models.Domain
	if err := s.client.http.DecodeResponse(resp, &domain); err != nil {
		return nil, err
	}

	return &domain, nil
}

// DisableDNSSEC disables DNSSEC for a domain at the registry.
func (s *DomainsService) DisableDNSSEC(ctx context.Context, domainRef string) (*models.Domain, error) {
	path := s.client.http.BuildPath("domains", url.PathEscape(domainRef), "dnssec", "disable")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var domain models.Domain
	if err := s.client.http.DecodeResponse(resp, &domain); err != nil {
		return nil, err
	}

	return &domain, nil
}

// GetTransferStatus retrieves the transfer status for a domain.
func (s *DomainsService) GetTransferStatus(ctx context.Context, domainRef string) (*models.Domain, error) {
	path := s.client.http.BuildPath("domains", url.PathEscape(domainRef), "transfer")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var domain models.Domain
	if err := s.client.http.DecodeResponse(resp, &domain); err != nil {
		return nil, err
	}

	return &domain, nil
}

// CheckDomains checks if domains are available for registration (simple check).
func (s *DomainsService) CheckDomains(ctx context.Context, domains []string) (*models.DomainCheckResponse, error) {
	path := s.client.http.BuildPath("domains", "check")

	query := url.Values{}
	for _, domain := range domains {
		query.Add("domains", domain)
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.DomainCheckResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
