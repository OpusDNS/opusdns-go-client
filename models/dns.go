// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// DNSSECStatus represents the DNSSEC status of a zone.
type DNSSECStatus string

const (
	// DNSSECStatusDisabled indicates DNSSEC is not enabled.
	DNSSECStatusDisabled DNSSECStatus = "disabled"

	// DNSSECStatusEnabled indicates DNSSEC is fully enabled.
	DNSSECStatusEnabled DNSSECStatus = "enabled"

	// DNSSECStatusPending indicates DNSSEC is being enabled or disabled.
	DNSSECStatusPending DNSSECStatus = "pending"
)

// RRSetType represents a DNS record type.
type RRSetType string

const (
	RRSetTypeA      RRSetType = "A"
	RRSetTypeAAAA   RRSetType = "AAAA"
	RRSetTypeALIAS  RRSetType = "ALIAS"
	RRSetTypeCAA    RRSetType = "CAA"
	RRSetTypeCNAME  RRSetType = "CNAME"
	RRSetTypeDNSKEY RRSetType = "DNSKEY"
	RRSetTypeDS     RRSetType = "DS"
	RRSetTypeMX     RRSetType = "MX"
	RRSetTypeNS     RRSetType = "NS"
	RRSetTypePTR    RRSetType = "PTR"
	RRSetTypeTXT    RRSetType = "TXT"
	RRSetTypeSOA    RRSetType = "SOA"
	RRSetTypeSRV    RRSetType = "SRV"
	RRSetTypeSMIMEA RRSetType = "SMIMEA"
	RRSetTypeTLSA   RRSetType = "TLSA"
	RRSetTypeURI    RRSetType = "URI"
)

// ZoneSortField represents fields that can be used for sorting zones.
type ZoneSortField string

const (
	ZoneSortByName      ZoneSortField = "name"
	ZoneSortByCreatedOn ZoneSortField = "created_on"
	ZoneSortByUpdatedOn ZoneSortField = "updated_on"
)

// Zone represents a DNS zone managed by OpusDNS.
type Zone struct {
	// Name is the domain name of the zone (e.g., "example.com").
	Name string `json:"name"`

	// DNSSECStatus indicates the DNSSEC status of the zone.
	DNSSECStatus DNSSECStatus `json:"dnssec_status,omitempty"`

	// DomainParts contains the parsed parts of the domain name.
	DomainParts *DomainNameParts `json:"domain_parts,omitempty"`

	// RRSets contains the resource record sets for this zone.
	// This field is populated when fetching a single zone with records.
	RRSets []RRSet `json:"rrsets,omitempty"`

	// CreatedOn is the timestamp when the zone was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is the timestamp when the zone was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// ZoneListResponse represents the paginated response when listing zones.
type ZoneListResponse struct {
	// Results contains the list of zones for the current page.
	Results []Zone `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// ZoneSummary represents a summary of DNS zones.
type ZoneSummary struct {
	// TotalZones is the total number of DNS zones.
	TotalZones int `json:"total_zones"`

	// ZonesByDNSSEC is a count of zones grouped by DNSSEC status.
	ZonesByDNSSEC map[DNSSECStatus]int `json:"zones_by_dnssec,omitempty"`
}

// ZoneCreateRequest represents the request body for creating a new zone.
type ZoneCreateRequest struct {
	// Name is the domain name for the new zone (e.g., "example.com").
	Name string `json:"name"`

	// RRSets is an optional list of initial resource record sets to create with the zone.
	RRSets []RRSetCreate `json:"rrsets,omitempty"`
}

// RRSet represents a resource record set (multiple records with same name/type).
type RRSet struct {
	// Name is the record name relative to the zone (e.g., "www" or "@" for apex).
	Name string `json:"name"`

	// Type is the DNS record type (A, AAAA, CNAME, etc.).
	Type RRSetType `json:"type"`

	// TTL is the time-to-live in seconds.
	TTL int `json:"ttl"`

	// Records contains the individual record data.
	Records []RecordData `json:"records,omitempty"`
}

// RecordData represents the data portion of a DNS record.
type RecordData struct {
	// RData is the record data (e.g., IP address, hostname, etc.).
	RData string `json:"rdata"`

	// Protected indicates if the record is protected from deletion.
	Protected bool `json:"protected,omitempty"`
}

// RRSetCreate represents a resource record set for creation.
type RRSetCreate struct {
	// Name is the record name relative to the zone.
	Name string `json:"name"`

	// Type is the DNS record type.
	Type RRSetType `json:"type"`

	// TTL is the time-to-live in seconds.
	TTL int `json:"ttl"`

	// Records contains the record data values.
	Records []string `json:"records"`
}

// Record represents a single DNS record (convenience type).
type Record struct {
	// Name is the record name relative to the zone.
	Name string `json:"name"`

	// Type is the DNS record type.
	Type RRSetType `json:"type"`

	// TTL is the time-to-live in seconds.
	TTL int `json:"ttl"`

	// RData is the record data.
	RData string `json:"rdata"`

	// Protected indicates if the record is protected from deletion.
	Protected bool `json:"protected,omitempty"`
}

// RecordPatchOp represents an operation for patching records.
type RecordPatchOp string

const (
	// RecordOpUpsert creates or updates a record.
	RecordOpUpsert RecordPatchOp = "upsert"

	// RecordOpRemove deletes a record.
	RecordOpRemove RecordPatchOp = "remove"
)

// RecordOperation represents an operation on a DNS record.
type RecordOperation struct {
	// Op is the operation type ("upsert" or "remove").
	Op RecordPatchOp `json:"op"`

	// Record is the record to operate on.
	Record Record `json:"record"`
}

// RecordPatchRequest represents a request to patch records in a zone.
type RecordPatchRequest struct {
	// Ops is the list of operations to perform.
	Ops []RecordOperation `json:"ops"`
}

// RRSetPatchOp represents an operation for patching RRSets.
type RRSetPatchOp struct {
	// Op is the operation type.
	Op RecordPatchOp `json:"op"`

	// Name is the RRSet name.
	Name string `json:"name"`

	// Type is the RRSet type.
	Type RRSetType `json:"type"`

	// TTL is the time-to-live in seconds.
	TTL int `json:"ttl,omitempty"`

	// RData is the record data (for single-record operations).
	RData string `json:"rdata,omitempty"`

	// Records is the list of record data (for multi-record operations).
	Records []string `json:"records,omitempty"`
}

// RRSetPatchRequest represents a request to patch RRSets in a zone.
type RRSetPatchRequest struct {
	// Ops is the list of operations to perform.
	Ops []RRSetPatchOp `json:"ops"`
}

// DNSChanges represents the response from operations that modify DNS records.
type DNSChanges struct {
	// ChangesetID is the unique identifier for this changeset.
	ChangesetID string `json:"changeset_id,omitempty"`

	// ZoneName is the name of the zone that was modified.
	ZoneName string `json:"zone_name,omitempty"`

	// NumChanges is the number of changes made.
	NumChanges int `json:"num_changes"`

	// Changes contains the individual changes made.
	Changes []DNSChange `json:"changes,omitempty"`
}

// DNSChange represents a single change in a DNS changeset.
type DNSChange struct {
	// Action is the action performed (e.g., "create", "update", "delete").
	Action string `json:"action"`

	// RRSetName is the name of the affected RRSet.
	RRSetName string `json:"rrset_name,omitempty"`

	// RRSetType is the type of the affected RRSet.
	RRSetType RRSetType `json:"rrset_type,omitempty"`

	// RecordData is the affected record data.
	RecordData string `json:"record_data,omitempty"`

	// TTL is the TTL of the affected record.
	TTL int `json:"ttl,omitempty"`
}

// DNSSECInfo represents DNSSEC information for a zone.
type DNSSECInfo struct {
	// Status is the current DNSSEC status.
	Status DNSSECStatus `json:"status"`

	// DSRecords contains the DS records for the zone.
	DSRecords []DSRecord `json:"ds_records,omitempty"`

	// DNSKEYRecords contains the DNSKEY records for the zone.
	DNSKEYRecords []DNSKEYRecord `json:"dnskey_records,omitempty"`
}

// DSRecord represents a DNSSEC Delegation Signer record.
type DSRecord struct {
	// KeyTag is the key tag.
	KeyTag int `json:"key_tag"`

	// Algorithm is the DNSSEC algorithm number.
	Algorithm int `json:"algorithm"`

	// DigestType is the digest algorithm type.
	DigestType int `json:"digest_type"`

	// Digest is the digest value.
	Digest string `json:"digest"`
}

// DNSKEYRecord represents a DNSSEC DNSKEY record.
type DNSKEYRecord struct {
	// Flags is the DNSKEY flags.
	Flags int `json:"flags"`

	// Protocol is the protocol value (always 3).
	Protocol int `json:"protocol"`

	// Algorithm is the DNSSEC algorithm number.
	Algorithm int `json:"algorithm"`

	// PublicKey is the base64-encoded public key.
	PublicKey string `json:"public_key"`
}

// ListZonesOptions contains options for listing zones.
type ListZonesOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of zones per page.
	PageSize int

	// SortBy is the field to sort by.
	SortBy ZoneSortField

	// SortOrder is the sort direction.
	SortOrder SortOrder

	// Search is an optional search query to filter zones by name.
	Search string

	// Name filters by exact zone name.
	Name string

	// Suffix filters by domain suffix (e.g., ".com").
	Suffix string

	// DNSSECStatus filters by DNSSEC status.
	DNSSECStatus DNSSECStatus

	// CreatedAfter filters zones created after this time.
	CreatedAfter *time.Time

	// CreatedBefore filters zones created before this time.
	CreatedBefore *time.Time
}
