// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// DomainID is a TypeID for domains.
type DomainID = TypeID

// DomainStatus represents the status of a domain.
type DomainStatus string

const (
	DomainStatusActive           DomainStatus = "active"
	DomainStatusPendingCreate    DomainStatus = "pending_create"
	DomainStatusPendingTransfer  DomainStatus = "pending_transfer"
	DomainStatusPendingRenew     DomainStatus = "pending_renew"
	DomainStatusPendingDelete    DomainStatus = "pending_delete"
	DomainStatusPendingRestore   DomainStatus = "pending_restore"
	DomainStatusExpired          DomainStatus = "expired"
	DomainStatusRedemptionPeriod DomainStatus = "redemption_period"
	DomainStatusDeleted          DomainStatus = "deleted"
)

// DomainClientStatus represents client-side domain statuses.
type DomainClientStatus string

const (
	DomainClientStatusTransferProhibited DomainClientStatus = "clientTransferProhibited"
	DomainClientStatusUpdateProhibited   DomainClientStatus = "clientUpdateProhibited"
	DomainClientStatusDeleteProhibited   DomainClientStatus = "clientDeleteProhibited"
	DomainClientStatusRenewProhibited    DomainClientStatus = "clientRenewProhibited"
	DomainClientStatusHold               DomainClientStatus = "clientHold"
)

// DomainServerStatus represents server-side domain statuses.
type DomainServerStatus string

const (
	DomainServerStatusOK                 DomainServerStatus = "ok"
	DomainServerStatusTransferProhibited DomainServerStatus = "serverTransferProhibited"
	DomainServerStatusUpdateProhibited   DomainServerStatus = "serverUpdateProhibited"
	DomainServerStatusDeleteProhibited   DomainServerStatus = "serverDeleteProhibited"
	DomainServerStatusRenewProhibited    DomainServerStatus = "serverRenewProhibited"
	DomainServerStatusHold               DomainServerStatus = "serverHold"
	DomainServerStatusPendingDelete      DomainServerStatus = "pendingDelete"
	DomainServerStatusPendingTransfer    DomainServerStatus = "pendingTransfer"
	DomainServerStatusRedemptionPeriod   DomainServerStatus = "redemptionPeriod"
)

// DomainContactType represents the type of contact for a domain.
type DomainContactType string

const (
	DomainContactTypeRegistrant DomainContactType = "registrant"
	DomainContactTypeAdmin      DomainContactType = "admin"
	DomainContactTypeTech       DomainContactType = "tech"
	DomainContactTypeBilling    DomainContactType = "billing"
)

// DomainSortField represents fields that can be used for sorting domains.
type DomainSortField string

const (
	DomainSortByName         DomainSortField = "name"
	DomainSortByCreatedOn    DomainSortField = "created_on"
	DomainSortByExpiresOn    DomainSortField = "expires_on"
	DomainSortByRegisteredOn DomainSortField = "registered_on"
)

// Domain represents a registered domain.
type Domain struct {
	// DomainID is the unique identifier for the domain.
	DomainID DomainID `json:"domain_id"`

	// Name is the domain name (e.g., "example.com").
	Name string `json:"name"`

	// OwnerID is the organization ID that owns the domain.
	OwnerID TypeID `json:"owner_id,omitempty"`

	// RegistryAccountID is the registry account ID.
	RegistryAccountID TypeID `json:"registry_account_id,omitempty"`

	// Nameservers is the list of nameservers for the domain.
	Nameservers []Nameserver `json:"nameservers,omitempty"`

	// Contacts is the list of contacts for the domain.
	Contacts []DomainContact `json:"contacts,omitempty"`

	// Hosts contains subordinate hosts (glue records) for the domain.
	Hosts []DomainHost `json:"hosts,omitempty"`

	// RegistryStatuses contains the server-side statuses.
	RegistryStatuses []DomainServerStatus `json:"registry_statuses,omitempty"`

	// ClientStatuses contains the client-side statuses.
	ClientStatuses []DomainClientStatus `json:"client_statuses,omitempty"`

	// AuthCode is the authorization code for transfers.
	AuthCode *string `json:"auth_code,omitempty"`

	// AuthCodeExpiresOn is when the auth code expires.
	AuthCodeExpiresOn *time.Time `json:"auth_code_expires_on,omitempty"`

	// AutoRenew indicates if the domain will auto-renew.
	AutoRenew bool `json:"auto_renew,omitempty"`

	// TransferLock indicates if transfers are prohibited.
	TransferLock bool `json:"transfer_lock,omitempty"`

	// RegisteredOn is when the domain was registered.
	RegisteredOn *time.Time `json:"registered_on,omitempty"`

	// ExpiresOn is when the domain expires.
	ExpiresOn *time.Time `json:"expires_on,omitempty"`

	// DeletedOn is when the domain will be deleted.
	DeletedOn *time.Time `json:"deleted_on,omitempty"`

	// CanceledOn is when the domain was canceled.
	CanceledOn *time.Time `json:"canceled_on,omitempty"`

	// CreatedOn is when the domain record was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the domain record was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// Nameserver represents a nameserver for a domain.
type Nameserver struct {
	// Hostname is the nameserver hostname.
	Hostname string `json:"hostname"`

	// IPv4 is the optional IPv4 glue record.
	IPv4 *string `json:"ipv4,omitempty"`

	// IPv6 is the optional IPv6 glue record.
	IPv6 *string `json:"ipv6,omitempty"`
}

// DomainContact represents a contact associated with a domain.
type DomainContact struct {
	// Type is the contact type (registrant, admin, tech, billing).
	Type DomainContactType `json:"type"`

	// ContactID is the ID of the contact.
	ContactID ContactID `json:"contact_id"`

	// Attributes contains additional contact attributes.
	Attributes map[string]string `json:"attributes,omitempty"`
}

// DomainHost represents a subordinate host (glue record) for a domain.
type DomainHost struct {
	// Hostname is the full hostname.
	Hostname string `json:"hostname"`

	// IPv4Addresses contains the IPv4 addresses.
	IPv4Addresses []string `json:"ipv4_addresses,omitempty"`

	// IPv6Addresses contains the IPv6 addresses.
	IPv6Addresses []string `json:"ipv6_addresses,omitempty"`
}

// DomainListResponse represents the paginated response when listing domains.
type DomainListResponse struct {
	// Results contains the list of domains for the current page.
	Results []Domain `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// DomainSummary represents a summary of domains.
type DomainSummary struct {
	// TotalDomains is the total number of domains.
	TotalDomains int `json:"total_domains"`

	// DomainsByTLD is a count of domains grouped by TLD.
	DomainsByTLD map[string]int `json:"domains_by_tld,omitempty"`

	// DomainsByStatus is a count of domains grouped by status.
	DomainsByStatus map[string]int `json:"domains_by_status,omitempty"`

	// ExpiringWithin30Days is the count of domains expiring within 30 days.
	ExpiringWithin30Days int `json:"expiring_within_30_days,omitempty"`

	// ExpiringWithin90Days is the count of domains expiring within 90 days.
	ExpiringWithin90Days int `json:"expiring_within_90_days,omitempty"`
}

// DomainCreateRequest represents a request to register a new domain.
type DomainCreateRequest struct {
	// Name is the domain name to register.
	Name string `json:"name"`

	// Period is the registration period in years.
	Period int `json:"period,omitempty"`

	// Contacts maps contact types to contact IDs.
	Contacts map[DomainContactType]ContactHandle `json:"contacts"`

	// Nameservers is the list of nameservers.
	Nameservers []Nameserver `json:"nameservers,omitempty"`

	// TransferLock enables transfer lock after registration.
	TransferLock *bool `json:"transfer_lock,omitempty"`

	// AutoRenew enables auto-renewal.
	AutoRenew *bool `json:"auto_renew,omitempty"`
}

// ContactHandle represents a contact reference with optional attributes.
type ContactHandle struct {
	// ContactID is the ID of the contact.
	ContactID ContactID `json:"contact_id"`

	// Attributes contains additional contact attributes for this domain.
	Attributes map[string]string `json:"attributes,omitempty"`
}

// DomainUpdateRequest represents a request to update a domain.
type DomainUpdateRequest struct {
	// Contacts maps contact types to contact handles to update.
	Contacts map[DomainContactType]ContactHandle `json:"contacts,omitempty"`

	// Nameservers is the new list of nameservers.
	Nameservers []Nameserver `json:"nameservers,omitempty"`

	// TransferLock updates the transfer lock status.
	TransferLock *bool `json:"transfer_lock,omitempty"`

	// AutoRenew updates the auto-renewal status.
	AutoRenew *bool `json:"auto_renew,omitempty"`

	// AddStatuses are client statuses to add.
	AddStatuses []DomainClientStatus `json:"add_statuses,omitempty"`

	// RemoveStatuses are client statuses to remove.
	RemoveStatuses []DomainClientStatus `json:"remove_statuses,omitempty"`
}

// DomainTransferRequest represents a request to transfer a domain.
type DomainTransferRequest struct {
	// Name is the domain name to transfer.
	Name string `json:"name"`

	// AuthCode is the authorization code for the transfer.
	AuthCode string `json:"auth_code"`

	// Contacts maps contact types to contact handles.
	Contacts map[DomainContactType]ContactHandle `json:"contacts,omitempty"`

	// Nameservers is the list of nameservers to use after transfer.
	Nameservers []Nameserver `json:"nameservers,omitempty"`

	// Period is the renewal period in years (if applicable).
	Period int `json:"period,omitempty"`
}

// DomainRenewRequest represents a request to renew a domain.
type DomainRenewRequest struct {
	// Period is the renewal period in years.
	Period int `json:"period"`

	// CurrentExpiryDate is the current expiration date (for verification).
	CurrentExpiryDate *time.Time `json:"current_expiry_date,omitempty"`
}

// DomainRestoreRequest represents a request to restore a deleted domain.
type DomainRestoreRequest struct {
	// Period is the renewal period in years after restoration.
	Period int `json:"period,omitempty"`
}

// DomainDNSSECRequest represents a request to configure DNSSEC for a domain.
type DomainDNSSECRequest struct {
	// DSRecords contains DS records to add.
	DSRecords []DSRecord `json:"ds_records,omitempty"`

	// DNSKEYRecords contains DNSKEY records to add.
	DNSKEYRecords []DNSKEYRecord `json:"dnskey_records,omitempty"`
}

// ListDomainsOptions contains options for listing domains.
type ListDomainsOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of domains per page.
	PageSize int

	// SortBy is the field to sort by.
	SortBy DomainSortField

	// SortOrder is the sort direction.
	SortOrder SortOrder

	// Search is an optional search query to filter domains.
	Search string

	// Name filters by exact domain name.
	Name string

	// TLD filters by top-level domain.
	TLD string

	// SLD filters by second-level domain.
	SLD string

	// TransferLock filters by transfer lock status.
	TransferLock *bool

	// AutoRenew filters by auto-renew status.
	AutoRenew *bool

	// ExpiresAfter filters domains expiring after this date.
	ExpiresAfter *time.Time

	// ExpiresBefore filters domains expiring before this date.
	ExpiresBefore *time.Time

	// Status filters by domain status.
	Status DomainStatus
}
