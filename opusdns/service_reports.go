package opusdns

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/opusdns/opusdns-go-client/models"
)

// ReportsService provides methods for managing reports.
type ReportsService struct {
	client *Client
}

// CreateReport creates a new report. Returns 202 Accepted.
// Rate limited to 1 report per 5 minutes per organization per report type.
func (s *ReportsService) CreateReport(ctx context.Context, req *models.CreateReportRequest) (*models.Report, error) {
	path := s.client.http.BuildPath("reports")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var result models.Report
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ListReports retrieves all reports with automatic pagination.
func (s *ReportsService) ListReports(ctx context.Context, opts *models.ListReportsOptions) ([]models.Report, error) {
	var all []models.Report
	page := 1

	for {
		pageOpts := opts
		if pageOpts == nil {
			pageOpts = &models.ListReportsOptions{}
		}
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := s.ListReportsPage(ctx, pageOpts)
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

// ListReportsPage retrieves a single page of reports.
func (s *ReportsService) ListReportsPage(ctx context.Context, opts *models.ListReportsOptions) (*models.ReportListResponse, error) {
	path := s.client.http.BuildPath("reports")

	query := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		for _, rt := range opts.ReportType {
			query.Add("report_type", string(rt))
		}
		for _, st := range opts.Status {
			query.Add("status", string(st))
		}
		if opts.TriggerType != "" {
			query.Set("trigger_type", string(opts.TriggerType))
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

	var result models.ReportListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetReport retrieves a specific report by ID.
func (s *ReportsService) GetReport(ctx context.Context, reportID models.ReportID) (*models.Report, error) {
	path := s.client.http.BuildPath("reports", string(reportID))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.Report
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DownloadReport downloads a completed report as a ZIP file.
// Returns the raw bytes of the ZIP file. The report must be in COMPLETED status.
// Returns ErrConflict (409) if the report is not ready for download.
func (s *ReportsService) DownloadReport(ctx context.Context, reportID models.ReportID) ([]byte, error) {
	path := s.client.http.BuildPath("reports", string(reportID), "download")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		if decErr := s.client.http.DecodeResponse(resp, nil); decErr != nil {
			return nil, decErr
		}
		return nil, fmt.Errorf("opusdns: report download failed with status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// DownloadReportToWriter downloads a completed report and writes it to the given writer.
// This is useful for streaming the report directly to a file.
func (s *ReportsService) DownloadReportToWriter(ctx context.Context, reportID models.ReportID, w io.Writer) error {
	data, err := s.DownloadReport(ctx, reportID)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("opusdns: failed to write report data: %w", err)
	}

	return nil
}
