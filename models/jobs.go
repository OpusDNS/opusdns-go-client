// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// BatchID is a TypeID for job batches (prefix: "batch").
type BatchID = TypeID

// JobID is a TypeID for individual jobs (prefix: "job").
type JobID = TypeID

// JobStatus represents the status of an individual job.
type JobStatus string

const (
	JobStatusBlocked    JobStatus = "blocked"
	JobStatusQueued     JobStatus = "queued"
	JobStatusPaused     JobStatus = "paused"
	JobStatusRunning    JobStatus = "running"
	JobStatusSucceeded  JobStatus = "succeeded"
	JobStatusFailed     JobStatus = "failed"
	JobStatusCanceled   JobStatus = "canceled"
	JobStatusDeadLetter JobStatus = "dead_letter"
)

// BatchStatus represents the status of a job batch.
type BatchStatus string

const (
	BatchStatusPending  BatchStatus = "pending"
	BatchStatusComplete BatchStatus = "complete"
)

// BatchSortField represents fields that can be used for sorting batches.
type BatchSortField string

const (
	BatchSortByCreatedOn  BatchSortField = "created_on"
	BatchSortByStartedAt  BatchSortField = "started_at"
	BatchSortByFinishedAt BatchSortField = "finished_at"
)

// CommandPayload represents a single command in a batch job request.
// The "command" field acts as the discriminator for the payload type.
type CommandPayload struct {
	// Command is the command type identifier (e.g., "domain_create", "dns_zone_update").
	Command string `json:"command"`

	// Version is the command version (defaults to "v1").
	Version string `json:"version,omitempty"`

	// Payload is the command-specific payload data.
	Payload interface{} `json:"payload"`

	// IdempotencyKey is an optional key to prevent duplicate command execution.
	IdempotencyKey *string `json:"idempotency_key,omitempty"`
}

// JobBatchRequest represents a request to create a batch of async commands.
type JobBatchRequest struct {
	// Commands is the list of commands to execute (max 50,000).
	Commands []CommandPayload `json:"commands"`

	// Label is an optional human-readable label for the batch.
	Label *string `json:"label,omitempty"`

	// NotBefore is the earliest time jobs can execute (UTC).
	// If not provided, jobs run immediately.
	NotBefore *time.Time `json:"not_before,omitempty"`

	// Paused indicates whether jobs should be created in a paused state.
	Paused bool `json:"paused,omitempty"`
}

// CommandError represents an error for a specific command in a batch creation.
type CommandError struct {
	// Index is the zero-based index of the failed command in the request.
	Index int `json:"index"`

	// Error is the error message.
	Error string `json:"error"`
}

// CreateJobBatchResponse represents the response when creating a new job batch.
type CreateJobBatchResponse struct {
	// BatchID is the unique identifier for the created batch.
	BatchID BatchID `json:"batch_id"`

	// JobsCreated is the number of jobs successfully created.
	JobsCreated int `json:"jobs_created"`

	// JobsFailed is the number of jobs that failed to create.
	JobsFailed int `json:"jobs_failed"`

	// TotalCommands is the total number of commands in the batch.
	TotalCommands int `json:"total_commands"`

	// StatusURL is the URL to check batch status.
	StatusURL string `json:"status_url"`

	// Errors contains details of any failed commands.
	Errors []CommandError `json:"errors,omitempty"`
}

// JobBatchStatusResponse represents the detailed status of a job batch.
type JobBatchStatusResponse struct {
	// BatchID is the unique identifier for the batch.
	BatchID BatchID `json:"batch_id"`

	// Total is the total number of jobs in the batch.
	Total int `json:"total"`

	// Blocked is the number of jobs waiting for eligibility.
	Blocked int `json:"blocked"`

	// Queued is the number of jobs awaiting processing.
	Queued int `json:"queued"`

	// Paused is the number of jobs in a paused state.
	Paused int `json:"paused"`

	// Running is the number of jobs currently being executed.
	Running int `json:"running"`

	// Succeeded is the number of jobs completed successfully.
	Succeeded int `json:"succeeded"`

	// Failed is the number of jobs that failed execution.
	Failed int `json:"failed"`

	// Canceled is the number of jobs that were canceled.
	Canceled int `json:"canceled"`

	// DeadLetter is the number of jobs permanently failed after retries.
	DeadLetter int `json:"dead_letter"`

	// ProgressPercentage is the completion percentage (0-100).
	ProgressPercentage float64 `json:"progress_percentage"`
}

// JobCountsByStatus contains job counts grouped by status.
type JobCountsByStatus struct {
	Blocked    int `json:"blocked"`
	Queued     int `json:"queued"`
	Paused     int `json:"paused"`
	Running    int `json:"running"`
	Succeeded  int `json:"succeeded"`
	Failed     int `json:"failed"`
	Canceled   int `json:"canceled"`
	DeadLetter int `json:"dead_letter"`
}

// JobBatchMetadataResponse represents batch metadata in list responses.
type JobBatchMetadataResponse struct {
	// BatchID is the unique identifier for the batch.
	BatchID BatchID `json:"batch_id"`

	// Label is the optional human-readable label for the batch.
	Label *string `json:"label,omitempty"`

	// Status is the batch status (pending or complete).
	Status BatchStatus `json:"status"`

	// JobCounts contains the number of jobs in each status.
	JobCounts JobCountsByStatus `json:"job_counts"`

	// TotalJobs is the total number of jobs in the batch.
	TotalJobs int `json:"total_jobs"`

	// CreatedOn is when the batch was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// StartedAt is when the first job in the batch started.
	StartedAt *time.Time `json:"started_at,omitempty"`

	// FinishedAt is when the last job in the batch finished.
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// JobResponse represents an individual job within a batch.
type JobResponse struct {
	// JobID is the unique identifier for the job.
	JobID JobID `json:"job_id"`

	// Status is the current job status.
	Status JobStatus `json:"status"`

	// Command is the command name (e.g., "domain_create", "dns_zone_update").
	Command *string `json:"command,omitempty"`

	// Operation is the operation type (e.g., "create", "update", "transfer").
	Operation *string `json:"operation,omitempty"`

	// ResourceKey is the resource identifier for this job.
	ResourceKey *string `json:"resource_key,omitempty"`

	// Display is a human-readable description of the job.
	Display *string `json:"display,omitempty"`

	// Payload is the original request payload.
	Payload interface{} `json:"payload,omitempty"`

	// ErrorClass is the error type if the job failed.
	ErrorClass *string `json:"error_class,omitempty"`

	// ErrorMessage is the detailed error message if the job failed.
	ErrorMessage *string `json:"error_message,omitempty"`

	// CreatedOn is when the job was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// StartedAt is when job execution began.
	StartedAt *time.Time `json:"started_at,omitempty"`

	// FinishedAt is when job execution completed.
	FinishedAt *time.Time `json:"finished_at,omitempty"`

	// PausedAt is when the job was paused.
	PausedAt *time.Time `json:"paused_at,omitempty"`

	// Attempts is the number of execution attempts.
	Attempts int `json:"attempts"`
}

// JobBatchListResponse represents the paginated response when listing job batches.
type JobBatchListResponse struct {
	// Results contains the list of batch metadata for the current page.
	Results []JobBatchMetadataResponse `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// JobListResponse represents the paginated response when listing jobs within a batch.
type JobListResponse struct {
	// Results contains the list of jobs for the current page.
	Results []JobResponse `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// ListBatchesOptions contains options for listing job batches.
type ListBatchesOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of batches per page.
	PageSize int

	// SortBy is the field to sort by.
	SortBy BatchSortField

	// SortOrder is the sort direction.
	SortOrder SortOrder

	// Status filters by batch status (pending or complete).
	Status BatchStatus
}

// ListBatchJobsOptions contains options for listing jobs within a batch.
type ListBatchJobsOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of jobs per page.
	PageSize int

	// SortBy is the field to sort by.
	SortBy BatchSortField

	// SortOrder is the sort direction.
	SortOrder SortOrder

	// Status filters by job status (repeatable, matches any).
	Status []JobStatus
}
