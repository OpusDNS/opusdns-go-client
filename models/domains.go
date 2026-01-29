// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// DomainID is a TypeID for domains.
type DomainID = TypeID

// DomainStatus represents the status of a domain.
type DomainStatus string

const (
	// Standard EPP statuses
	DomainStatusOK                       DomainStatus = "ok"
	DomainStatusServerTransferProhibited DomainStatus = "serverTransferProhibited"
	DomainStatusServerUpdateProhibited   DomainStatus = "serverUpdateProhibited"
	DomainStatusServerDeleteProhibited   DomainStatus = "serverDeleteProhibited"
	DomainStatusServerRenewProhibited    DomainStatus = "serverRenewProhibited"
	DomainStatusServerRestoreProhibited  DomainStatus = "serverRestoreProhibited"
	DomainStatusServerHold               DomainStatus = "serverHold"
	DomainStatusTransferPeriod           DomainStatus = "transferPeriod"
	DomainStatusRenewPeriod              DomainStatus = "renewPeriod"
	DomainStatusRedemptionPeriod         DomainStatus = "redemptionPeriod"
	DomainStatusPendingUpdate            DomainStatus = "pendingUpdate"
	DomainStatusPendingTransfer          DomainStatus = "pendingTransfer"
	DomainStatusPendingRestore           DomainStatus = "pendingRestore"
	DomainStatusPendingRenew             DomainStatus = "pendingRenew"
	DomainStatusPendingDelete            DomainStatus = "pendingDelete"
	DomainStatusPendingCreate            DomainStatus = "pendingCreate"
	DomainStatusInactive                 DomainStatus = "inactive"
	DomainStatusAutoRenewPeriod          DomainStatus = "autoRenewPeriod"
	DomainStatusAddPeriod                DomainStatus = "addPeriod"
	DomainStatusDeleted                  DomainStatus = "deleted"
	DomainStatusClientTransferProhibited DomainStatus = "clientTransferProhibited"
	DomainStatusClientUpdateProhibited   DomainStatus = "clientUpdateProhibited"
	DomainStatusClientDeleteProhibited   DomainStatus = "clientDeleteProhibited"
	DomainStatusClientRenewProhibited    DomainStatus = "clientRenewProhibited"
	DomainStatusClientHold               DomainStatus = "clientHold"
	DomainStatusFree                     DomainStatus = "free"
	DomainStatusConnect                  DomainStatus = "connect"
	DomainStatusFailed                   DomainStatus = "failed"
	DomainStatusInvalid                  DomainStatus = "invalid"
)

// PeriodUnit represents the unit of time for a domain period.
type PeriodUnit string

const (
	// PeriodUnitYear represents years.
	PeriodUnitYear PeriodUnit = "y"

	// PeriodUnitMonth represents months.
	PeriodUnitMonth PeriodUnit = "m"

	// PeriodUnitDay represents days.
	PeriodUnitDay PeriodUnit = "d"
)

// DomainPeriod represents a registration/renewal period.
type DomainPeriod struct {
	// Value is the amount of time in the specified unit.
	Value int `json:"value"`

	// Unit is the unit of time (y, m, d).
	Unit PeriodUnit `json:"unit"`
}

// RenewalMode represents the renewal mode for a domain.
type RenewalMode string

const (
	// RenewalModeRenew indicates the domain will auto-renew.
	RenewalModeRenew RenewalMode = "renew"

	// RenewalModeExpire indicates the domain will expire without renewal.
	RenewalModeExpire RenewalMode = "expire"
)

// IsAutoRenew returns true if the domain is set to auto-renew.
func (r RenewalMode) IsAutoRenew() bool {
	return r == RenewalModeRenew
}

// RenewalModePtr returns a pointer to the given RenewalMode.
func RenewalModePtr(r RenewalMode) *RenewalMode {
	return &r
}

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

	// SLD is the second-level domain (e.g., "example" in "example.com").
	SLD string `json:"sld"`

	// TLD is the top-level domain (e.g., "com" in "example.com").
	TLD string `json:"tld"`

	// ROID is the registry object identifier for the domain.
	ROID string `json:"roid"`

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

	// RegistryStatuses contains all the domain statuses from the registry.
	RegistryStatuses []string `json:"registry_statuses,omitempty"`

	// AuthCode is the authorization code for transfers.
	AuthCode *string `json:"auth_code,omitempty"`

	// AuthCodeExpiresOn is when the auth code expires.
	AuthCodeExpiresOn *time.Time `json:"auth_code_expires_on,omitempty"`

	// RenewalMode indicates the renewal mode (renew or expire).
	RenewalMode RenewalMode `json:"renewal_mode,omitempty"`

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

	// IPAddresses contains the IP addresses of the nameserver (both IPv4 and IPv6).
	IPAddresses []string `json:"ip_addresses,omitempty"`
}

// DomainContact represents a contact associated with a domain.
type DomainContact struct {
	// ContactID is the ID of the contact.
	ContactID ContactID `json:"contact_id"`

	// ContactType is the contact type (registrant, admin, tech, billing).
	ContactType DomainContactType `json:"contact_type"`
}

// DomainHost represents a subordinate host (glue record) for a domain.
type DomainHost struct {
	// HostID is the unique identifier for the host.
	HostID TypeID `json:"host_id"`
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

	// Contacts maps contact types to contact handles.
	Contacts map[DomainContactType][]ContactHandle `json:"contacts"`

	// RenewalMode sets the renewal mode (renew or expire).
	RenewalMode RenewalMode `json:"renewal_mode"`

	// Period is the registration period.
	Period DomainPeriod `json:"period"`

	// Nameservers is the list of nameservers.
	Nameservers []Nameserver `json:"nameservers,omitempty"`

	// AuthCode is the auth code used for the domain (optional).
	AuthCode *string `json:"auth_code,omitempty"`

	// CreateZone creates a zone for the domain on OpusDNS nameserver infrastructure.
	CreateZone bool `json:"create_zone,omitempty"`
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
	Contacts map[DomainContactType][]ContactHandle `json:"contacts,omitempty"`

	// Nameservers is the new list of nameservers.
	Nameservers []Nameserver `json:"nameservers,omitempty"`

	// RenewalMode updates the renewal mode (renew or expire).
	RenewalMode *RenewalMode `json:"renewal_mode,omitempty"`

	// ClientStatuses is the complete list of client statuses to set on the domain.
	// This replaces the entire client status list.
	Statuses []string `json:"statuses"`
}

// DomainTransferRequest represents a request to transfer a domain.
type DomainTransferRequest struct {
	// Name is the domain name to transfer.
	Name string `json:"name"`

	// AuthCode is the authorization code for the transfer.
	AuthCode string `json:"auth_code"`

	// Contacts maps contact types to contact handles.
	Contacts map[DomainContactType][]ContactHandle `json:"contacts,omitempty"`

	// RenewalMode sets the renewal mode (renew or expire).
	RenewalMode RenewalMode `json:"renewal_mode"`

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

	// RenewalMode filters by renewal mode.
	RenewalMode *RenewalMode

	// ExpiresAfter filters domains expiring after this date.
	ExpiresAfter *time.Time

	// ExpiresBefore filters domains expiring before this date.
	ExpiresBefore *time.Time

	// Status filters by domain status.
	Status DomainStatus
}
