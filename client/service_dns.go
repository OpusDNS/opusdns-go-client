package client

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/opusdns/opusdns-go-client/models"
)

// DNSService provides methods for managing DNS zones and records.
type DNSService struct {
	client *Client
}

// ListZones retrieves all DNS zones with automatic pagination.
func (s *DNSService) ListZones(ctx context.Context, opts *models.ListZonesOptions) ([]models.Zone, error) {
	var allZones []models.Zone
	page := 1

	for {
		pageOpts := opts
		if pageOpts == nil {
			pageOpts = &models.ListZonesOptions{}
		}
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := s.ListZonesPage(ctx, pageOpts)
		if err != nil {
			return nil, err
		}

		allZones = append(allZones, resp.Results...)

		if !resp.Pagination.HasNextPage {
			break
		}
		page++
	}

	return allZones, nil
}

// ListZonesPage retrieves a single page of DNS zones.
func (s *DNSService) ListZonesPage(ctx context.Context, opts *models.ListZonesOptions) (*models.ZoneListResponse, error) {
	path := s.client.http.BuildPath("dns")

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
		if opts.Suffix != "" {
			query.Set("suffix", opts.Suffix)
		}
		if opts.DNSSECStatus != "" {
			query.Set("dnssec_status", string(opts.DNSSECStatus))
		}
		if opts.CreatedAfter != nil {
			query.Set("created_after", opts.CreatedAfter.Format(time.RFC3339))
		}
		if opts.CreatedBefore != nil {
			query.Set("created_before", opts.CreatedBefore.Format(time.RFC3339))
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.ZoneListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetZone retrieves a specific zone by name.
func (s *DNSService) GetZone(ctx context.Context, name string) (*models.Zone, error) {
	name = strings.TrimSuffix(name, ".")
	path := s.client.http.BuildPath("dns", url.PathEscape(name))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var zone models.Zone
	if err := s.client.http.DecodeResponse(resp, &zone); err != nil {
		return nil, err
	}

	return &zone, nil
}

// CreateZone creates a new DNS zone.
func (s *DNSService) CreateZone(ctx context.Context, req *models.ZoneCreateRequest) (*models.Zone, error) {
	path := s.client.http.BuildPath("dns")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var zone models.Zone
	if err := s.client.http.DecodeResponse(resp, &zone); err != nil {
		return nil, err
	}

	return &zone, nil
}

// DeleteZone deletes a DNS zone.
func (s *DNSService) DeleteZone(ctx context.Context, name string) error {
	name = strings.TrimSuffix(name, ".")
	path := s.client.http.BuildPath("dns", url.PathEscape(name))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// GetSummary retrieves a summary of DNS zones.
func (s *DNSService) GetSummary(ctx context.Context) (*models.ZoneSummary, error) {
	path := s.client.http.BuildPath("dns", "summary")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var summary models.ZoneSummary
	if err := s.client.http.DecodeResponse(resp, &summary); err != nil {
		return nil, err
	}

	return &summary, nil
}

// GetRRSets retrieves all resource record sets for a zone.
func (s *DNSService) GetRRSets(ctx context.Context, zoneName string) ([]models.RRSet, error) {
	zoneName = strings.TrimSuffix(zoneName, ".")
	path := s.client.http.BuildPath("dns", url.PathEscape(zoneName), "rrsets")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var rrsets []models.RRSet
	if err := s.client.http.DecodeResponse(resp, &rrsets); err != nil {
		return nil, err
	}

	return rrsets, nil
}

// PatchRecords applies multiple record operations atomically.
func (s *DNSService) PatchRecords(ctx context.Context, zoneName string, ops []models.RecordOperation) error {
	zoneName = strings.TrimSuffix(zoneName, ".")
	path := s.client.http.BuildPath("dns", url.PathEscape(zoneName), "records")

	req := models.RecordPatchRequest{Ops: ops}

	resp, err := s.client.http.Patch(ctx, path, req)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// UpsertRecord creates or updates a single DNS record.
func (s *DNSService) UpsertRecord(ctx context.Context, zoneName string, record models.Record) error {
	return s.PatchRecords(ctx, zoneName, []models.RecordOperation{
		{Op: models.RecordOpUpsert, Record: record},
	})
}

// DeleteRecord removes a single DNS record.
func (s *DNSService) DeleteRecord(ctx context.Context, zoneName string, record models.Record) error {
	return s.PatchRecords(ctx, zoneName, []models.RecordOperation{
		{Op: models.RecordOpRemove, Record: record},
	})
}

// EnableDNSSEC enables DNSSEC for a zone.
func (s *DNSService) EnableDNSSEC(ctx context.Context, zoneName string) (*models.DNSChanges, error) {
	zoneName = strings.TrimSuffix(zoneName, ".")
	path := s.client.http.BuildPath("dns", url.PathEscape(zoneName), "dnssec", "enable")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var changes models.DNSChanges
	if err := s.client.http.DecodeResponse(resp, &changes); err != nil {
		return nil, err
	}

	return &changes, nil
}

// DisableDNSSEC disables DNSSEC for a zone.
func (s *DNSService) DisableDNSSEC(ctx context.Context, zoneName string) (*models.DNSChanges, error) {
	zoneName = strings.TrimSuffix(zoneName, ".")
	path := s.client.http.BuildPath("dns", url.PathEscape(zoneName), "dnssec", "disable")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var changes models.DNSChanges
	if err := s.client.http.DecodeResponse(resp, &changes); err != nil {
		return nil, err
	}

	return &changes, nil
}
