package opusdns

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
		pageOpts := cloneOptions(opts)
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
		for _, tagID := range opts.TagIDs {
			query.Add("tag_ids", string(tagID))
		}
		if opts.TagMode != "" {
			query.Set("tag_mode", string(opts.TagMode))
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
		if opts.IsPremium != nil {
			query.Set("is_premium", strconv.FormatBool(*opts.IsPremium))
		}
		if opts.RenewalMode != nil {
			query.Set("renewal_mode", string(*opts.RenewalMode))
		}
		if opts.CreatedAfter != nil {
			query.Set("created_after", opts.CreatedAfter.Format(time.RFC3339))
		}
		if opts.CreatedBefore != nil {
			query.Set("created_before", opts.CreatedBefore.Format(time.RFC3339))
		}
		if opts.UpdatedAfter != nil {
			query.Set("updated_after", opts.UpdatedAfter.Format(time.RFC3339))
		}
		if opts.UpdatedBefore != nil {
			query.Set("updated_before", opts.UpdatedBefore.Format(time.RFC3339))
		}
		if opts.ExpiresAfter != nil {
			query.Set("expires_after", opts.ExpiresAfter.Format(time.RFC3339))
		}
		if opts.ExpiresBefore != nil {
			query.Set("expires_before", opts.ExpiresBefore.Format(time.RFC3339))
		}
		if opts.ExpiresIn30Days != nil {
			query.Set("expires_in_30_days", strconv.FormatBool(*opts.ExpiresIn30Days))
		}
		if opts.ExpiresIn60Days != nil {
			query.Set("expires_in_60_days", strconv.FormatBool(*opts.ExpiresIn60Days))
		}
		if opts.ExpiresIn90Days != nil {
			query.Set("expires_in_90_days", strconv.FormatBool(*opts.ExpiresIn90Days))
		}
		if opts.RegisteredAfter != nil {
			query.Set("registered_after", opts.RegisteredAfter.Format(time.RFC3339))
		}
		if opts.RegisteredBefore != nil {
			query.Set("registered_before", opts.RegisteredBefore.Format(time.RFC3339))
		}
		for _, status := range opts.RegistryStatuses {
			query.Add("registry_statuses", status)
		}
		for _, include := range opts.Include {
			query.Add("include", string(include))
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
	return s.GetDomainWithOptions(ctx, domainRef, nil)
}

// GetDomainWithOptions retrieves a specific domain by ID or name with optional response expansions.
func (s *DomainsService) GetDomainWithOptions(ctx context.Context, domainRef string, opts *models.GetDomainOptions) (*models.Domain, error) {
	path := s.client.http.BuildPath("domains", url.PathEscape(domainRef))

	query := url.Values{}
	if opts != nil {
		for _, include := range opts.Include {
			query.Add("include", string(include))
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
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
func (s *DomainsService) GetDNSSEC(ctx context.Context, domainRef string) ([]models.DomainDNSSECDataResponse, error) {
	path := s.client.http.BuildPath("domains", url.PathEscape(domainRef), "dnssec")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var dnssec []models.DomainDNSSECDataResponse
	if err := s.client.http.DecodeResponse(resp, &dnssec); err != nil {
		return nil, err
	}

	return dnssec, nil
}

// PutDNSSEC replaces all DNSSEC data for a domain.
func (s *DomainsService) PutDNSSEC(ctx context.Context, domainRef string, data []models.DomainDNSSECDataCreate) ([]models.DomainDNSSECDataResponse, error) {
	path := s.client.http.BuildPath("domains", url.PathEscape(domainRef), "dnssec")

	resp, err := s.client.http.Put(ctx, path, data)
	if err != nil {
		return nil, err
	}

	var result []models.DomainDNSSECDataResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteDNSSEC removes all DNSSEC data for a domain.
func (s *DomainsService) DeleteDNSSEC(ctx context.Context, domainRef string) error {
	path := s.client.http.BuildPath("domains", url.PathEscape(domainRef), "dnssec")

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// EnableDNSSEC enables DNSSEC for a domain at the registry.
func (s *DomainsService) EnableDNSSEC(ctx context.Context, domainRef string) ([]models.DomainDNSSECDataResponse, error) {
	path := s.client.http.BuildPath("domains", url.PathEscape(domainRef), "dnssec", "enable")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result []models.DomainDNSSECDataResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// DisableDNSSEC disables DNSSEC for a domain at the registry.
func (s *DomainsService) DisableDNSSEC(ctx context.Context, domainRef string) error {
	path := s.client.http.BuildPath("domains", url.PathEscape(domainRef), "dnssec", "disable")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
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
