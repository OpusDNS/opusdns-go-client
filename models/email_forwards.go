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

// EmailForward represents an email forwarding configuration for a hostname.
type EmailForward struct {
	// EmailForwardID is the unique identifier for the email forward.
	EmailForwardID EmailForwardID `json:"email_forward_id"`

	// Hostname is the domain name for email forwarding (e.g., "example.com").
	Hostname string `json:"hostname"`

	// Enabled indicates whether email forwarding is active.
	Enabled bool `json:"enabled"`

	// Aliases contains the list of email aliases for this hostname.
	Aliases []EmailForwardAlias `json:"aliases,omitempty"`

	// CreatedOn is when the email forward was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the email forward was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// EmailForwardAlias represents a single email alias that forwards to a destination.
type EmailForwardAlias struct {
	// AliasID is the unique identifier for the alias.
	AliasID EmailForwardAliasID `json:"alias_id,omitempty"`

	// LocalPart is the part before @ (e.g., "info" for "info@example.com").
	// Use "*" for catch-all.
	LocalPart string `json:"local_part"`

	// Destinations contains the email addresses to forward to.
	Destinations []string `json:"destinations"`

	// Enabled indicates whether this specific alias is active.
	Enabled bool `json:"enabled"`

	// CreatedOn is when the alias was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the alias was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// FullAddress returns the full email address for this alias.
func (a *EmailForwardAlias) FullAddress(hostname string) string {
	if a.LocalPart == "*" {
		return "*@" + hostname
	}
	return a.LocalPart + "@" + hostname
}

// IsCatchAll returns true if this is a catch-all alias.
func (a *EmailForwardAlias) IsCatchAll() bool {
	return a.LocalPart == "*"
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

	// EmailForwards contains the email forwards for this zone.
	EmailForwards []EmailForward `json:"email_forwards"`
}

// EmailForwardCreateRequest represents a request to create email forwarding for a hostname.
type EmailForwardCreateRequest struct {
	// Hostname is the domain name to enable email forwarding for.
	Hostname string `json:"hostname"`

	// Aliases is an optional list of initial aliases to create.
	Aliases []EmailForwardAliasCreate `json:"aliases,omitempty"`
}

// EmailForwardAliasCreate represents a request to create an email alias.
type EmailForwardAliasCreate struct {
	// LocalPart is the part before @ (use "*" for catch-all).
	LocalPart string `json:"local_part"`

	// Destinations contains the email addresses to forward to.
	Destinations []string `json:"destinations"`
}

// EmailForwardAliasUpdate represents a request to update an email alias.
type EmailForwardAliasUpdate struct {
	// LocalPart is the part before @ (optional, cannot change to/from catch-all).
	LocalPart *string `json:"local_part,omitempty"`

	// Destinations contains the new list of destination email addresses.
	Destinations []string `json:"destinations,omitempty"`

	// Enabled updates the enabled status.
	Enabled *bool `json:"enabled,omitempty"`
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
	// LogID is the unique identifier for the log entry.
	LogID TypeID `json:"log_id,omitempty"`

	// EmailForwardID is the ID of the associated email forward.
	EmailForwardID EmailForwardID `json:"email_forward_id"`

	// AliasID is the ID of the associated alias (if applicable).
	AliasID *EmailForwardAliasID `json:"alias_id,omitempty"`

	// FromAddress is the sender email address.
	FromAddress string `json:"from_address"`

	// ToAddress is the original recipient address.
	ToAddress string `json:"to_address"`

	// ForwardedTo contains the addresses the email was forwarded to.
	ForwardedTo []string `json:"forwarded_to,omitempty"`

	// Status is the delivery status.
	Status string `json:"status"`

	// StatusMessage provides additional status information.
	StatusMessage *string `json:"status_message,omitempty"`

	// MessageID is the email message ID.
	MessageID *string `json:"message_id,omitempty"`

	// Subject is the email subject (may be truncated).
	Subject *string `json:"subject,omitempty"`

	// Size is the message size in bytes.
	Size int `json:"size,omitempty"`

	// Timestamp is when the email was processed.
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

// EmailForwardLogListResponse represents the paginated response when listing email forward logs.
type EmailForwardLogListResponse struct {
	// Results contains the list of log entries for the current page.
	Results []EmailForwardLog `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}
