package opusdns

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportsService_CreateReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/reports", r.URL.Path)

		var body models.CreateReportRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, models.ReportTypeDomainInventory, body.ReportType)

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(models.Report{
			ReportID:   "report_123",
			ReportType: models.ReportTypeDomainInventory,
			Status:     models.ReportStatusPending,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	report, err := client.Reports.CreateReport(context.Background(), &models.CreateReportRequest{
		ReportType: models.ReportTypeDomainInventory,
	})
	require.NoError(t, err)
	assert.Equal(t, models.ReportID("report_123"), report.ReportID)
	assert.Equal(t, models.ReportStatusPending, report.Status)
}

func TestReportsService_ListReports(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/reports", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.ReportListResponse{
			Results: []models.Report{
				{ReportID: "report_1", Status: models.ReportStatusCompleted},
				{ReportID: "report_2", Status: models.ReportStatusGenerating},
			},
			Pagination: models.Pagination{HasNextPage: false, CurrentPage: 1},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	reports, err := client.Reports.ListReports(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, reports, 2)
	assert.Equal(t, models.ReportID("report_1"), reports[0].ReportID)
	assert.Equal(t, models.ReportID("report_2"), reports[1].ReportID)
}

func TestReportsService_ListReportsPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/reports", r.URL.Path)

		query := r.URL.Query()
		assert.Equal(t, "2", query.Get("page"))
		assert.Equal(t, "10", query.Get("page_size"))
		assert.Equal(t, string(models.ReportTypeDNSZoneSummary), query.Get("report_type"))
		assert.Equal(t, string(models.ReportStatusCompleted), query.Get("status"))
		assert.Equal(t, string(models.ReportTriggerScheduled), query.Get("trigger_type"))

		_ = json.NewEncoder(w).Encode(models.ReportListResponse{
			Results: []models.Report{
				{ReportID: "report_1", Status: models.ReportStatusCompleted},
			},
			Pagination: models.Pagination{CurrentPage: 2, HasNextPage: true},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Reports.ListReportsPage(context.Background(), &models.ListReportsOptions{
		Page:        2,
		PageSize:    10,
		ReportType:  []models.ReportType{models.ReportTypeDNSZoneSummary},
		Status:      []models.ReportStatus{models.ReportStatusCompleted},
		TriggerType: models.ReportTriggerScheduled,
	})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, 2, resp.Pagination.CurrentPage)
	assert.True(t, resp.Pagination.HasNextPage)
}

func TestReportsService_GetReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/reports/report_123", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.Report{
			ReportID:   "report_123",
			ReportType: models.ReportTypeDomainInventory,
			Status:     models.ReportStatusCompleted,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	report, err := client.Reports.GetReport(context.Background(), models.ReportID("report_123"))
	require.NoError(t, err)
	assert.Equal(t, models.ReportID("report_123"), report.ReportID)
	assert.Equal(t, models.ReportStatusCompleted, report.Status)
}

func TestReportsService_DownloadReport(t *testing.T) {
	payload := []byte("PK\x03\x04 fake zip bytes")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/reports/report_123/download", r.URL.Path)

		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	data, err := client.Reports.DownloadReport(context.Background(), models.ReportID("report_123"))
	require.NoError(t, err)
	assert.Equal(t, payload, data)
}

func TestReportsService_DownloadReportToWriter(t *testing.T) {
	payload := []byte("PK\x03\x04 fake zip bytes")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/reports/report_123/download", r.URL.Path)

		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	var buf bytes.Buffer
	var w io.Writer = &buf
	err = client.Reports.DownloadReportToWriter(context.Background(), models.ReportID("report_123"), w)
	require.NoError(t, err)
	assert.Equal(t, payload, buf.Bytes())
}
