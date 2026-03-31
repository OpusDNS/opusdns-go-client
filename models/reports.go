// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// ReportID is a TypeID for reports (prefix: "report").
type ReportID = TypeID

// ReportType represents the type of report to generate.
type ReportType string

const (
	ReportTypeDomainInventory ReportType = "domain_inventory"
	ReportTypeDNSZoneSummary  ReportType = "dns_zone_summary"
	ReportTypeDNSZoneRecords  ReportType = "dns_zone_records"
)

// ReportStatus represents the status of a report.
type ReportStatus string

const (
	ReportStatusPending    ReportStatus = "pending"
	ReportStatusGenerating ReportStatus = "generating"
	ReportStatusCompleted  ReportStatus = "completed"
	ReportStatusFailed     ReportStatus = "failed"
)

// ReportTriggerType represents how the report was triggered.
type ReportTriggerType string

const (
	ReportTriggerOnDemand  ReportTriggerType = "on_demand"
	ReportTriggerScheduled ReportTriggerType = "scheduled"
)

// Report represents a generated report.
type Report struct {
	// ReportID is the unique identifier for the report.
	ReportID ReportID `json:"report_id"`

	// OrganizationID is the organization that owns the report.
	OrganizationID OrganizationID `json:"organization_id"`

	// ReportType is the type of report.
	ReportType ReportType `json:"report_type"`

	// Status is the current status of the report.
	Status ReportStatus `json:"status"`

	// TriggerType indicates how the report was triggered.
	TriggerType ReportTriggerType `json:"trigger_type"`

	// FileSizeBytes is the size of the generated report file in bytes.
	FileSizeBytes *int `json:"file_size_bytes,omitempty"`

	// RecordCount is the number of records in the report.
	RecordCount *int `json:"record_count,omitempty"`

	// GeneratedOn is when the report was generated.
	GeneratedOn *time.Time `json:"generated_on,omitempty"`

	// CreatedOn is when the report was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the report was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// CreateReportRequest represents a request to create a new report.
type CreateReportRequest struct {
	// ReportType is the type of report to generate (defaults to domain_inventory).
	ReportType ReportType `json:"report_type,omitempty"`
}

// ReportListResponse represents the paginated response when listing reports.
type ReportListResponse struct {
	// Results contains the list of reports for the current page.
	Results []Report `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// ListReportsOptions contains options for listing reports.
type ListReportsOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of reports per page.
	PageSize int

	// ReportType filters by report type (repeatable, matches any).
	ReportType []ReportType

	// Status filters by report status (repeatable, matches any).
	Status []ReportStatus

	// TriggerType filters by trigger type.
	TriggerType ReportTriggerType

	// CreatedAfter filters reports created after this time.
	CreatedAfter *time.Time

	// CreatedBefore filters reports created before this time.
	CreatedBefore *time.Time
}
