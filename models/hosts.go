// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// HostID is a TypeID for host objects (prefix "host_").
type HostID = TypeID

// HostStatus represents the lifecycle status of a host object.
type HostStatus string

const (
	// HostStatusRequestedCreate indicates the host is part of a local request but not
	// yet created at the registry.
	HostStatusRequestedCreate HostStatus = "requested_create"

	// HostStatusPendingCreate indicates the host exists locally but is not yet created
	// by the registry.
	HostStatusPendingCreate HostStatus = "pending_create"

	// HostStatusActive indicates the host exists and is in use by the parent domain.
	HostStatusActive HostStatus = "active"

	// HostStatusInactive indicates the host exists but is not in use.
	HostStatusInactive HostStatus = "inactive"

	// HostStatusPendingDelete indicates the host is pending deletion.
	HostStatusPendingDelete HostStatus = "pending_delete"
)

// Host represents a host object.
type Host struct {
	// HostID is the unique identifier of the host object.
	HostID HostID `json:"host_id"`

	// Hostname is the hostname of the host object (e.g. "ns1.example.com").
	Hostname string `json:"hostname"`

	// IPAddresses is the list of IP addresses (IPv4 and/or IPv6) for the host object.
	IPAddresses []string `json:"ip_addresses"`

	// CreatedOn is when the host was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the host was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// HostCreateRequest is the request body for creating a host object.
type HostCreateRequest struct {
	// Hostname is the hostname of the host object (e.g. "ns1.example.com").
	Hostname string `json:"hostname"`

	// IPAddresses is the list of IP addresses for the host object (at least one).
	IPAddresses []string `json:"ip_addresses"`
}

// HostUpdateRequest is the request body for updating a host object's IP addresses.
type HostUpdateRequest struct {
	// IPAddresses is the updated list of IP addresses for the host object (at least one).
	IPAddresses []string `json:"ip_addresses"`
}

// HostAvailability represents the availability of a single hostname.
type HostAvailability struct {
	// Hostname is the hostname checked.
	Hostname string `json:"hostname"`

	// Available indicates whether the hostname is available.
	Available bool `json:"available"`

	// Reason is the reason the hostname is unavailable, if any.
	Reason *string `json:"reason,omitempty"`
}

// HostCheckResponse is the response from a host availability check.
type HostCheckResponse struct {
	// Results contains the availability result for each checked hostname.
	Results []HostAvailability `json:"results"`
}
