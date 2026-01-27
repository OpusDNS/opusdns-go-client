// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// EmailForwardID is a TypeID for email forwards.
type EmailForwardID = TypeID

// EmailForwardAliasID is a TypeID for email forward aliases.
type EmailForwardAliasID = TypeID

// EmailForwardSortField represents fields that can be used for sorting email forwards.
type EmailForwardSortField string

const (
	EmailForwardSortByHostname  EmailForwardSortField = "hostname"
	EmailForwardSortByEnabled   EmailForwardSortField = "enabled"
	EmailForwardSortByCreatedOn EmailForwardSortField = "created_on"
	EmailForwardSortByUpdatedOn EmailForwardSortField = "updated_on"
)

// EmailForwardLogSortField represents fields that can be used for sorting email forward logs.
type EmailForwardLogSortField string

const (
	EmailForwardLogSortByLogID          EmailForwardLogSortField = "log_id"
	EmailForwardLogSortBySenderEmail    EmailForwardLogSortField = "sender_email"
	EmailForwardLogSortByRecipientEmail EmailForwardLogSortField = "recipient_email"
	EmailForwardLogSortByForwardEmail   EmailForwardLogSortField = "forward_email"
	EmailForwardLogSortByFinalStatus    EmailForwardLogSortField = "final_status"
	EmailForwardLogSortByCreatedOn      EmailForwardLogSortField = "created_on"
	EmailForwardLogSortBySyncedOn       EmailForwardLogSortField = "synced_on"
)

// EmailForwardLogStatus represents the status of an email forward log.
type EmailForwardLogStatus string

const (
	EmailForwardLogStatusQueued     EmailForwardLogStatus = "QUEUED"
	EmailForwardLogStatusDelivered  EmailForwardLogStatus = "DELIVERED"
	EmailForwardLogStatusRefused    EmailForwardLogStatus = "REFUSED"
	EmailForwardLogStatusSoftBounce EmailForwardLogStatus = "SOFT-BOUNCE"
	EmailForwardLogStatusHardBounce EmailForwardLogStatus = "HARD-BOUNCE"
)

// EmailForward represents an email forwarding configuration for a hostname.
type EmailForward struct {
	// EmailForwardID is the unique identifier for the email forward.
	EmailForwardID EmailForwardID `json:"email_forward_id"`

	// Hostname is the domain name for email forwarding (e.g., "example.com").
	Hostname string `json:"hostname"`

	// Enabled indicates whether email forwarding is active.
	Enabled bool `json:"enabled"`

	// Aliases contains the list of email aliases for this hostname.
	Aliases []EmailForwardAlias `json:"aliases"`

	// CreatedOn is when the email forward was created.
	CreatedOn time.Time `json:"created_on"`

	// UpdatedOn is when the email forward was last updated.
	UpdatedOn time.Time `json:"updated_on"`
}

// EmailForwardAlias represents a single email alias that forwards to a destination.
type EmailForwardAlias struct {
	// EmailForwardAliasID is the unique identifier for the alias.
	EmailForwardAliasID EmailForwardAliasID `json:"email_forward_alias_id"`

	// Alias is the email alias (e.g., "info" for "info@example.com").
	// Use "*" for catch-all.
	Alias string `json:"alias"`

	// ForwardTo contains the email addresses to forward to.
	ForwardTo []string `json:"forward_to"`
}

// FullAddress returns the full email address for this alias.
func (a *EmailForwardAlias) FullAddress(hostname string) string {
	if a.Alias == "*" {
		return "*@" + hostname
	}
	return a.Alias + "@" + hostname
}

// IsCatchAll returns true if this is a catch-all alias.
func (a *EmailForwardAlias) IsCatchAll() bool {
	return a.Alias == "*"
}

// EmailForwardListResponse represents the paginated response when listing email forwards.
type EmailForwardListResponse struct {
	// Results contains the list of email forwards for the current page.
	Results []EmailForward `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// EmailForwardZone represents email forwards associated with a DNS zone.
type EmailForwardZone struct {
	// ZoneID is the ID of the associated zone.
	ZoneID TypeID `json:"zone_id"`

	// ZoneName is the name of the zone.
	ZoneName string `json:"zone_name"`

	// EmailForwards contains the email forwards for this zone.
	EmailForwards []EmailForward `json:"email_forwards"`
}

// EmailForwardZoneSortField represents fields that can be used for sorting email forward zones.
type EmailForwardZoneSortField string

const (
	EmailForwardZoneSortByZoneName EmailForwardZoneSortField = "zone_name"
)

// EmailForwardCreateRequest represents a request to create email forwarding for a hostname.
type EmailForwardCreateRequest struct {
	// Hostname is the domain name to enable email forwarding for.
	Hostname string `json:"hostname"`

	// Aliases is an optional list of initial aliases to create.
	Aliases []EmailForwardAliasCreate `json:"aliases,omitempty"`
}

// EmailForwardAliasCreate represents a request to create an email alias.
type EmailForwardAliasCreate struct {
	// Alias is the part before @ (use "*" for catch-all).
	Alias string `json:"alias"`

	// ForwardTo contains the email addresses to forward to.
	ForwardTo []string `json:"forward_to"`
}

// EmailForwardAliasUpdate represents a request to update an email alias.
type EmailForwardAliasUpdate struct {
	// ForwardTo contains the new list of destination email addresses.
	ForwardTo []string `json:"forward_to"`
}

// ListEmailForwardsOptions contains options for listing email forwards.
type ListEmailForwardsOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of items per page.
	PageSize int

	// SortBy is the field to sort by.
	SortBy EmailForwardSortField

	// SortOrder is the sort direction.
	SortOrder SortOrder

	// Search is an optional search query to filter by hostname.
	Search string

	// Enabled filters by enabled status.
	Enabled *bool
}

// EmailForwardLog represents a log entry for email forwarding activity.
type EmailForwardLog struct {
	// LogID is the unique identifier for the log entry from ImprovMX.
	LogID string `json:"log_id"`

	// Domain is the domain name.
	Domain string `json:"domain"`

	// SenderEmail is the sender email address.
	SenderEmail string `json:"sender_email"`

	// SenderName is the sender name (optional).
	SenderName *string `json:"sender_name,omitempty"`

	// RecipientEmail is the recipient email address (the alias).
	RecipientEmail string `json:"recipient_email"`

	// RecipientName is the recipient name (optional).
	RecipientName *string `json:"recipient_name,omitempty"`

	// ForwardEmail is the forward destination email address.
	ForwardEmail string `json:"forward_email"`

	// ForwardName is the forward destination name (optional).
	ForwardName *string `json:"forward_name,omitempty"`

	// Subject is the email subject.
	Subject string `json:"subject"`

	// Hostname is the hostname that received the email.
	Hostname string `json:"hostname"`

	// MessageID is the email message ID.
	MessageID string `json:"message_id"`

	// Transport is the transport method (mx or smtp).
	Transport string `json:"transport"`

	// FinalStatus is the final status of the email.
	FinalStatus EmailForwardLogStatus `json:"final_status"`

	// Events contains the list of processing events.
	Events []EmailForwardLogEvent `json:"events,omitempty"`

	// CreatedOn is when the email was received by ImprovMX.
	CreatedOn time.Time `json:"created_on"`

	// SyncedOn is when the record was synced to ClickHouse.
	SyncedOn time.Time `json:"synced_on"`
}

// EmailForwardLogEvent represents a processing event for an email forward log.
type EmailForwardLogEvent struct {
	// ID is the event ID.
	ID string `json:"id"`

	// Code is the event status code.
	Code int `json:"code"`

	// Status is the event status (QUEUED, DELIVERED, REFUSED, SOFT-BOUNCE, HARD-BOUNCE).
	Status string `json:"status"`

	// Message is the event message.
	Message string `json:"message"`

	// Server is the server that processed the event.
	Server string `json:"server"`

	// Local is the ImprovMX server that processed the event.
	Local string `json:"local"`

	// Created is when the event occurred.
	Created time.Time `json:"created"`
}

// EmailForwardLogListResponse represents the paginated response when listing email forward logs.
type EmailForwardLogListResponse struct {
	// Results contains the list of log entries for the current page.
	Results []EmailForwardLog `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// EmailForwardMetrics represents metrics for email forwards.
type EmailForwardMetrics struct {
	// TotalLogs is the total number of email forward logs.
	TotalLogs int `json:"total_logs"`

	// AliasCount is the number of aliases (optional).
	AliasCount *int `json:"alias_count,omitempty"`

	// ByStatus contains log counts grouped by status.
	ByStatus map[EmailForwardLogStatus]int `json:"by_status"`

	// ByAlias contains metrics breakdown per alias (optional).
	ByAlias []EmailForwardAliasMetrics `json:"by_alias,omitempty"`

	// Rates contains rate percentages for each status.
	Rates *EmailForwardMetricsRates `json:"rates,omitempty"`

	// Filters contains the applied filters.
	Filters *EmailForwardMetricsFilters `json:"filters,omitempty"`
}

// EmailForwardAliasMetrics represents metrics for a specific email alias.
type EmailForwardAliasMetrics struct {
	// Alias is the email alias address.
	Alias string `json:"alias"`

	// TotalLogs is the total number of logs for this alias.
	TotalLogs int `json:"total_logs"`

	// ByStatus contains log counts grouped by status.
	ByStatus map[EmailForwardLogStatus]int `json:"by_status"`
}

// EmailForwardMetricsRates contains rate percentages for email forward statuses.
type EmailForwardMetricsRates struct {
	// DeliveryRate is the percentage of delivered emails.
	DeliveryRate float64 `json:"delivery_rate,omitempty"`

	// BounceRate is the percentage of bounced emails.
	BounceRate float64 `json:"bounce_rate,omitempty"`

	// RefusedRate is the percentage of refused emails.
	RefusedRate float64 `json:"refused_rate,omitempty"`
}

// EmailForwardMetricsFilters contains the filters applied to email forward metrics.
type EmailForwardMetricsFilters struct {
	// StartDate is the start date filter.
	StartDate *time.Time `json:"start_date,omitempty"`

	// EndDate is the end date filter.
	EndDate *time.Time `json:"end_date,omitempty"`

	// Alias is the alias filter.
	Alias *string `json:"alias,omitempty"`
}
