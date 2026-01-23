package client

import (
	"context"
	"net/url"
	"strconv"

	"github.com/opusdns/opusdns-go-client/models"
)

// TLDsService provides methods for accessing TLD information.
type TLDsService struct {
	client *Client
}

// ListTLDs retrieves all available TLDs.
func (s *TLDsService) ListTLDs(ctx context.Context, opts *models.ListTLDsOptions) ([]models.TLD, error) {
	path := s.client.http.BuildPath("tlds", "")

	query := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.Search != "" {
			query.Set("search", opts.Search)
		}
		if opts.Type != "" {
			query.Set("type", string(opts.Type))
		}
		if opts.Available != nil {
			query.Set("available", strconv.FormatBool(*opts.Available))
		}
		if opts.RegistrationEnabled != nil {
			query.Set("registration_enabled", strconv.FormatBool(*opts.RegistrationEnabled))
		}
		if opts.DNSSECSupported != nil {
			query.Set("dnssec_supported", strconv.FormatBool(*opts.DNSSECSupported))
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.TLDListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

// GetTLD retrieves details for a specific TLD.
func (s *TLDsService) GetTLD(ctx context.Context, tld string) (*models.TLDDetails, error) {
	path := s.client.http.BuildPath("tlds", url.PathEscape(tld))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var details models.TLDDetails
	if err := s.client.http.DecodeResponse(resp, &details); err != nil {
		return nil, err
	}

	return &details, nil
}

// GetPortfolio retrieves the TLD portfolio for the organization.
func (s *TLDsService) GetPortfolio(ctx context.Context) (*models.TLDPortfolio, error) {
	path := s.client.http.BuildPath("tlds", "portfolio")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var portfolio models.TLDPortfolio
	if err := s.client.http.DecodeResponse(resp, &portfolio); err != nil {
		return nil, err
	}

	return &portfolio, nil
}

// AvailabilityService provides methods for checking domain availability.
type AvailabilityService struct {
	client *Client
}

// CheckAvailability checks the availability of multiple domains.
func (s *AvailabilityService) CheckAvailability(ctx context.Context, domains []string) (*models.AvailabilityResponse, error) {
	path := s.client.http.BuildPath("availability")

	query := url.Values{}
	for _, domain := range domains {
		query.Add("domains", domain)
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.AvailabilityResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CheckSingleAvailability is a convenience method for checking a single domain.
func (s *AvailabilityService) CheckSingleAvailability(ctx context.Context, domain string) (*models.DomainAvailability, error) {
	result, err := s.CheckAvailability(ctx, []string{domain})
	if err != nil {
		return nil, err
	}

	if len(result.Results) == 0 {
		return nil, ErrNotFound
	}

	return &result.Results[0], nil
}

// GetSuggestions retrieves domain name suggestions based on a query.
func (s *AvailabilityService) GetSuggestions(ctx context.Context, query string, opts *models.DomainSuggestRequest) (*models.DomainSuggestResponse, error) {
	path := s.client.http.BuildPath("domain-search", "suggest")

	urlQuery := url.Values{}
	urlQuery.Set("query", query)

	if opts != nil {
		if len(opts.TLDs) > 0 {
			for _, tld := range opts.TLDs {
				urlQuery.Add("tlds", tld)
			}
		}
		if opts.Limit > 0 {
			urlQuery.Set("limit", strconv.Itoa(opts.Limit))
		}
		if opts.IncludeUnavailable {
			urlQuery.Set("include_unavailable", "true")
		}
	}

	resp, err := s.client.http.Get(ctx, path, urlQuery)
	if err != nil {
		return nil, err
	}

	var result models.DomainSuggestResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
