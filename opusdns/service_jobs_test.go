package opusdns

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobsService_ListBatches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/jobs", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.JobBatchListResponse{
			Results: []models.JobBatchMetadataResponse{
				{BatchID: "batch_1", Status: models.BatchStatusPending, TotalJobs: 5},
			},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	batches, err := client.Jobs.ListBatches(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, batches, 1)
	assert.Equal(t, models.BatchID("batch_1"), batches[0].BatchID)
}

func TestJobsService_ListBatchesPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/jobs", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "pending", r.URL.Query().Get("status"))
		_ = json.NewEncoder(w).Encode(models.JobBatchListResponse{
			Results: []models.JobBatchMetadataResponse{
				{BatchID: "batch_2"},
			},
			Pagination: models.Pagination{HasNextPage: true},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Jobs.ListBatchesPage(context.Background(), &models.ListBatchesOptions{Page: 2, Status: models.BatchStatusPending})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, models.BatchID("batch_2"), resp.Results[0].BatchID)
	assert.True(t, resp.Pagination.HasNextPage)
}

func TestJobsService_CreateBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/jobs", r.URL.Path)

		var req models.JobBatchRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		require.Len(t, req.Commands, 1)
		assert.Equal(t, "domain_create", req.Commands[0].Command)

		_ = json.NewEncoder(w).Encode(models.CreateJobBatchResponse{
			BatchID:       "batch_1",
			JobsCreated:   1,
			TotalCommands: 1,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.Jobs.CreateBatch(context.Background(), &models.JobBatchRequest{
		Commands: []models.CommandPayload{
			{Command: "domain_create", Payload: map[string]string{"name": "example.com"}},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, models.BatchID("batch_1"), result.BatchID)
	assert.Equal(t, 1, result.JobsCreated)
}

func TestJobsService_GetBatchStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/jobs/batch_1", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.JobBatchStatusResponse{
			BatchID:            "batch_1",
			Total:              5,
			Succeeded:          5,
			ProgressPercentage: 100,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	status, err := client.Jobs.GetBatchStatus(context.Background(), "batch_1")
	require.NoError(t, err)
	assert.Equal(t, models.BatchID("batch_1"), status.BatchID)
	assert.Equal(t, 5, status.Total)
}

func TestJobsService_DeleteBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/jobs/batch_1", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Jobs.DeleteBatch(context.Background(), "batch_1")
	require.NoError(t, err)
}

func TestJobsService_PauseBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/jobs/batch_1/pause", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Jobs.PauseBatch(context.Background(), "batch_1")
	require.NoError(t, err)
}

func TestJobsService_ResumeBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/jobs/batch_1/resume", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Jobs.ResumeBatch(context.Background(), "batch_1")
	require.NoError(t, err)
}

func TestJobsService_RetryBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/jobs/batch_1/retry", r.URL.Path)
		assert.Equal(t, []string{"BillingInsufficientFundsError"}, r.URL.Query()["error_class"])
		_ = json.NewEncoder(w).Encode(models.JobBatchRetryResponse{BatchID: "batch_1", RetriedCount: 3})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.Jobs.RetryBatch(context.Background(), "batch_1", []string{"BillingInsufficientFundsError"})
	require.NoError(t, err)
	assert.Equal(t, 3, result.RetriedCount)
}

func TestJobsService_ListBatchJobs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/jobs/batch_1/jobs", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.JobListResponse{
			Results: []models.JobResponse{
				{JobID: "job_1", Status: models.JobStatusSucceeded},
			},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	jobs, err := client.Jobs.ListBatchJobs(context.Background(), "batch_1", nil)
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, models.JobID("job_1"), jobs[0].JobID)
}

func TestJobsService_ListBatchJobsPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/jobs/batch_1/jobs", r.URL.Path)
		assert.Equal(t, "3", r.URL.Query().Get("page"))
		assert.Equal(t, []string{"failed"}, r.URL.Query()["status"])
		_ = json.NewEncoder(w).Encode(models.JobListResponse{
			Results: []models.JobResponse{
				{JobID: "job_2", Status: models.JobStatusFailed},
			},
			Pagination: models.Pagination{HasNextPage: true},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Jobs.ListBatchJobsPage(context.Background(), "batch_1", &models.ListBatchJobsOptions{
		Page:   3,
		Status: []models.JobStatus{models.JobStatusFailed},
	})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, models.JobID("job_2"), resp.Results[0].JobID)
	assert.True(t, resp.Pagination.HasNextPage)
}

func TestJobsService_GetJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/job/job_1", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.JobResponse{JobID: "job_1", Status: models.JobStatusRunning})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	job, err := client.Jobs.GetJob(context.Background(), "job_1")
	require.NoError(t, err)
	assert.Equal(t, models.JobID("job_1"), job.JobID)
	assert.Equal(t, models.JobStatusRunning, job.Status)
}

func TestJobsService_PauseJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/job/job_1/pause", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Jobs.PauseJob(context.Background(), "job_1")
	require.NoError(t, err)
}

func TestJobsService_ResumeJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/job/job_1/resume", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.JobResponse{JobID: "job_1", Status: models.JobStatusQueued})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	job, err := client.Jobs.ResumeJob(context.Background(), "job_1")
	require.NoError(t, err)
	assert.Equal(t, models.JobID("job_1"), job.JobID)
	assert.Equal(t, models.JobStatusQueued, job.Status)
}

func TestJobsService_RetryJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/job/job_1/retry", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.JobResponse{JobID: "job_1"})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	job, err := client.Jobs.RetryJob(context.Background(), "job_1")
	require.NoError(t, err)
	assert.Equal(t, models.JobID("job_1"), job.JobID)
}

func TestJobsService_DeleteJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/job/job_1", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Jobs.DeleteJob(context.Background(), "job_1")
	require.NoError(t, err)
}
