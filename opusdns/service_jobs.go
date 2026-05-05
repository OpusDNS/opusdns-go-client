package opusdns

import (
	"context"
	"net/url"
	"strconv"

	"github.com/opusdns/opusdns-go-client/models"
)

// JobsService provides methods for managing async job batches.
type JobsService struct {
	client *Client
}

// ListBatches retrieves all job batches with automatic pagination.
func (s *JobsService) ListBatches(ctx context.Context, opts *models.ListBatchesOptions) ([]models.JobBatchMetadataResponse, error) {
	var all []models.JobBatchMetadataResponse
	page := 1

	for {
		pageOpts := cloneOptions(opts)
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := s.ListBatchesPage(ctx, pageOpts)
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

// ListBatchesPage retrieves a single page of job batches.
func (s *JobsService) ListBatchesPage(ctx context.Context, opts *models.ListBatchesOptions) (*models.JobBatchListResponse, error) {
	path := s.client.http.BuildPath("jobs")

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
		if opts.Status != "" {
			query.Set("status", string(opts.Status))
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.JobBatchListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateBatch creates a new job batch with the given commands.
func (s *JobsService) CreateBatch(ctx context.Context, req *models.JobBatchRequest) (*models.CreateJobBatchResponse, error) {
	path := s.client.http.BuildPath("jobs")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var result models.CreateJobBatchResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetBatchStatus retrieves the detailed status of a job batch.
func (s *JobsService) GetBatchStatus(ctx context.Context, batchID models.BatchID) (*models.JobBatchStatusResponse, error) {
	path := s.client.http.BuildPath("jobs", string(batchID))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.JobBatchStatusResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteBatch cancels all jobs in a batch.
func (s *JobsService) DeleteBatch(ctx context.Context, batchID models.BatchID) error {
	path := s.client.http.BuildPath("jobs", string(batchID))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// PauseBatch pauses all jobs in a batch.
func (s *JobsService) PauseBatch(ctx context.Context, batchID models.BatchID) error {
	path := s.client.http.BuildPath("jobs", string(batchID), "pause")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// ResumeBatch resumes all paused jobs in a batch.
func (s *JobsService) ResumeBatch(ctx context.Context, batchID models.BatchID) error {
	path := s.client.http.BuildPath("jobs", string(batchID), "resume")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// ListBatchJobs retrieves all jobs within a batch with automatic pagination.
func (s *JobsService) ListBatchJobs(ctx context.Context, batchID models.BatchID, opts *models.ListBatchJobsOptions) ([]models.JobResponse, error) {
	var all []models.JobResponse
	page := 1

	for {
		pageOpts := cloneOptions(opts)
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := s.ListBatchJobsPage(ctx, batchID, pageOpts)
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

// ListBatchJobsPage retrieves a single page of jobs within a batch.
func (s *JobsService) ListBatchJobsPage(ctx context.Context, batchID models.BatchID, opts *models.ListBatchJobsOptions) (*models.JobListResponse, error) {
	path := s.client.http.BuildPath("jobs", string(batchID), "jobs")

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
		for _, status := range opts.Status {
			query.Add("status", string(status))
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.JobListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetJob retrieves the details of a specific job.
func (s *JobsService) GetJob(ctx context.Context, jobID models.JobID) (*models.JobResponse, error) {
	path := s.client.http.BuildPath("job", string(jobID))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.JobResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// PauseJob pauses an individual job.
func (s *JobsService) PauseJob(ctx context.Context, jobID models.JobID) error {
	path := s.client.http.BuildPath("job", string(jobID), "pause")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// ResumeJob resumes a paused individual job.
func (s *JobsService) ResumeJob(ctx context.Context, jobID models.JobID) (*models.JobResponse, error) {
	path := s.client.http.BuildPath("job", string(jobID), "resume")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.JobResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteJob cancels an individual job.
func (s *JobsService) DeleteJob(ctx context.Context, jobID models.JobID) error {
	path := s.client.http.BuildPath("job", string(jobID))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}
