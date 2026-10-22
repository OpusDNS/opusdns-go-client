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

func TestJobsService_CreateBatch_DuplicatesAndInstanceErrors(t *testing.T) {
	// A resubmitted command with a known idempotency_key creates no job; the
	// pointer to the original job is only in duplicates[]. A failing bulk
	// instance is identified by instance_index and resource_key.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/jobs", r.URL.Path)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"batch_id": "batch_01h45ytscbebyvny4gc8cr8ma2",
			"jobs_created": 0,
			"jobs_failed": 1,
			"jobs_duplicated": 1,
			"total_commands": 2,
			"status_url": "/v1/jobs/batch_01h45ytscbebyvny4gc8cr8ma2",
			"errors": [{
				"index": 1,
				"error": "Zone not found",
				"instance_index": 3,
				"resource_key": "example.com",
				"code": "ERROR_ZONE_NOT_FOUND",
				"type": "zone-not-found"
			}],
			"duplicates": [{
				"index": 0,
				"existing_job_id": "job_01h45ytscbebyvny4gc8cr8ma2",
				"existing_batch_id": "batch_01h45ytscbebyvny4gc8cr8ma3",
				"instance_index": null,
				"resource_key": "example.org"
			}]
		}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.Jobs.CreateBatch(context.Background(), &models.JobBatchRequest{
		Commands: []models.CommandPayload{
			{Command: "domain_create", IdempotencyKey: models.StringPtr("k1"), Payload: map[string]string{"name": "example.org"}},
			{Command: "zone_records_bulk", Payload: map[string]string{}},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, 0, result.JobsCreated)
	assert.Equal(t, 1, result.JobsDuplicated)

	require.Len(t, result.Duplicates, 1)
	dup := result.Duplicates[0]
	assert.Equal(t, 0, dup.Index)
	assert.Equal(t, models.JobID("job_01h45ytscbebyvny4gc8cr8ma2"), dup.ExistingJobID)
	require.NotNil(t, dup.ExistingBatchID)
	assert.Equal(t, models.BatchID("batch_01h45ytscbebyvny4gc8cr8ma3"), *dup.ExistingBatchID)
	assert.Nil(t, dup.InstanceIndex)
	require.NotNil(t, dup.ResourceKey)
	assert.Equal(t, "example.org", *dup.ResourceKey)

	require.Len(t, result.Errors, 1)
	cmdErr := result.Errors[0]
	require.NotNil(t, cmdErr.InstanceIndex)
	assert.Equal(t, 3, *cmdErr.InstanceIndex)
	require.NotNil(t, cmdErr.ResourceKey)
	assert.Equal(t, "example.com", *cmdErr.ResourceKey)
	require.NotNil(t, cmdErr.Code)
	assert.Equal(t, "ERROR_ZONE_NOT_FOUND", *cmdErr.Code)
	require.NotNil(t, cmdErr.Type)
	assert.Equal(t, "zone-not-found", *cmdErr.Type)
}

func TestJobsService_EscapesIDsInPath(t *testing.T) {
	var uris []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uris = append(uris, r.RequestURI)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	// Unescaped, "../jobs/batch_X" would resolve to DELETE /v1/jobs/batch_X
	// and cancel the whole batch.
	require.NoError(t, client.Jobs.DeleteJob(context.Background(), "../jobs/batch_X"))
	require.NoError(t, client.Jobs.DeleteBatch(context.Background(), "batch_1/../x"))

	assert.Equal(t, []string{
		"/v1/job/..%2Fjobs%2Fbatch_X",
		"/v1/jobs/batch_1%2F..%2Fx",
	}, uris)
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
