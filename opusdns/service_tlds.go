package opusdns

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/opusdns/opusdns-go-client/models"
)

// TLDsService provides methods for accessing TLD information.
type TLDsService struct {
	client *Client
}

// ListTLDs retrieves the TLDs available to the organization.
// The API returns one TLD specification per entry, each covering one or more
// TLDs. This method flattens that structure and leaves out TLDs whose
// configuration is not enabled. Each TLD carries its entry's specification and
// the flat fields derived from it. The endpoint is not paginated; opts.TLDs
// limits the list to the named TLDs.
func (s *TLDsService) ListTLDs(ctx context.Context, opts *models.ListTLDsOptions) ([]models.TLD, error) {
	path := s.client.http.BuildPath("tlds", "")

	query := url.Values{}
	if opts != nil && len(opts.TLDs) > 0 {
		query.Set("tlds", strings.Join(opts.TLDs, ","))
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.TLDListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	// Extract TLDs from the nested structure
	// The API returns: { "tlds": [{ "enabled": true, "tlds": [{"name": "com", "type": "gTLD"}], ...spec }] }
	var tlds []models.TLD
	for i := range result.TLDConfigurations {
		config := &result.TLDConfigurations[i]
		if !config.Enabled {
			continue
		}
		spec := &config.Specification
		minYears, maxYears := registrationYears(spec.DomainLifecycle.RegistrationPeriods)
		for _, tldInfo := range config.TLDs {
			tld := models.TLD{
				Name:                  tldInfo.Name,
				Type:                  tldInfo.Type,
				Available:             config.Enabled,
				IDNSupported:          spec.IDN.IDNCapable,
				DNSSECSupported:       spec.DNSConfiguration.DNSSECAllowed,
				MinRegistrationPeriod: minYears,
				MaxRegistrationPeriod: maxYears,
				Specification:         spec,
			}
			tlds = append(tlds, tld)
		}
	}

	return tlds, nil
}

// registrationYears returns the shortest and longest of the periods that are
// whole years, in years. Both are 0 when no period is.
func registrationYears(periods []models.DomainPeriod) (minYears, maxYears int) {
	for _, p := range periods {
		var years int
		switch p.Unit {
		case models.PeriodUnitYear:
			years = p.Value
		case models.PeriodUnitMonth:
			if p.Value%12 != 0 {
				continue
			}
			years = p.Value / 12
		default:
			continue
		}
		if years <= 0 {
			continue
		}
		if minYears == 0 || years < minYears {
			minYears = years
		}
		if years > maxYears {
			maxYears = years
		}
	}
	return minYears, maxYears
}

// GetTLD retrieves the specification of a TLD, using the specification the
// caller's organization is pinned to.
func (s *TLDsService) GetTLD(ctx context.Context, tld string) (*models.TLDSpecification, error) {
	return s.GetTLDWithOptions(ctx, tld, nil)
}

// GetTLDWithOptions retrieves the specification of a TLD, optionally
// overriding which specification variant is resolved.
func (s *TLDsService) GetTLDWithOptions(ctx context.Context, tld string, opts *models.GetTLDOptions) (*models.TLDSpecification, error) {
	path := s.client.http.BuildPath("tlds", url.PathEscape(tld))

	query := url.Values{}
	if opts != nil {
		if opts.Backend != "" {
			query.Set("backend", opts.Backend)
		}
		if opts.CustomerSpecRef != "" {
			query.Set("customer_spec_ref", opts.CustomerSpecRef)
		}
		if opts.Version != "" {
			query.Set("version", opts.Version)
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var spec models.TLDSpecification
	if err := s.client.http.DecodeResponse(resp, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

// GetPortfolio retrieves the TLDs in the organization's portfolio.
func (s *TLDsService) GetPortfolio(ctx context.Context) ([]models.TLDPortfolioEntry, error) {
	path := s.client.http.BuildPath("tlds", "portfolio")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var portfolio []models.TLDPortfolioEntry
	if err := s.client.http.DecodeResponse(resp, &portfolio); err != nil {
		return nil, err
	}

	return portfolio, nil
}

// AvailabilityService provides methods for checking domain availability.
type AvailabilityService struct {
	client *Client
}

// MaxAvailabilityDomains is the number of domains the API checks in one
// availability request. CheckAvailability splits longer lists.
const MaxAvailabilityDomains = 50

// CheckAvailability checks the availability of multiple domains. Lists longer
// than MaxAvailabilityDomains are checked in several requests, one after the
// other; the results keep the order of domains, and Meta adds up the totals
// and processing times of all requests.
func (s *AvailabilityService) CheckAvailability(ctx context.Context, domains []string) (*models.AvailabilityResponse, error) {
	if len(domains) == 0 {
		return nil, &ValidationError{Field: "domains", Message: "at least one domain is required"}
	}

	path := s.client.http.BuildPath("availability")

	merged := &models.AvailabilityResponse{}
	for start := 0; start < len(domains); start += MaxAvailabilityDomains {
		end := start + MaxAvailabilityDomains
		if end > len(domains) {
			end = len(domains)
		}

		query := url.Values{}
		for _, domain := range domains[start:end] {
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

		merged.Results = append(merged.Results, result.Results...)
		merged.Meta.Total += result.Meta.Total
		merged.Meta.ProcessingTimeMs += result.Meta.ProcessingTimeMs
	}

	return merged, nil
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
		if opts.Premium != nil {
			urlQuery.Set("premium", strconv.FormatBool(*opts.Premium))
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
