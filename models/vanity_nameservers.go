// Package models contains all the data types for the OpusDNS API.
package models

// VanityNameserverSetID is a TypeID for vanity nameserver sets (prefix "vns_").
type VanityNameserverSetID = TypeID

// VanityNameserverSetStatus represents the lifecycle status of a vanity nameserver set.
type VanityNameserverSetStatus string

const (
	VanityNameserverSetStatusProvisioning VanityNameserverSetStatus = "provisioning"
	VanityNameserverSetStatusActive       VanityNameserverSetStatus = "active"
	VanityNameserverSetStatusSuspended    VanityNameserverSetStatus = "suspended"
	VanityNameserverSetStatusFailed       VanityNameserverSetStatus = "failed"
	VanityNameserverSetStatusDeleting     VanityNameserverSetStatus = "deleting"
)

// VanityNameserver is a single nameserver within a vanity nameserver set.
type VanityNameserver struct {
	// Hostname is the fully-qualified hostname of the vanity nameserver.
	Hostname string `json:"hostname"`

	// Position is the ordering within the set; the lowest position becomes the SOA MNAME.
	Position int `json:"position"`
}

// VanityNameserverSet represents a vanity nameserver set.
type VanityNameserverSet struct {
	// SetID is the stable identifier for the vanity NS set.
	SetID VanityNameserverSetID `json:"set_id"`

	// OrganizationID is the organization that owns the set.
	OrganizationID OrganizationID `json:"organization_id"`

	// Name is the human-readable name for the set.
	Name string `json:"name"`

	// ParentDomainName is the parent domain used as the apex of the vanity NS zone.
	ParentDomainName string `json:"parent_domain_name"`

	// SOARName is the SOA RNAME used verbatim when creating vanity-branded zones.
	SOARName string `json:"soa_rname"`

	// Status is the lifecycle status of the set.
	Status VanityNameserverSetStatus `json:"status"`

	// IsDefault indicates whether this is the organization's default vanity NS set.
	IsDefault bool `json:"is_default"`

	// Nameservers are the nameservers in the set, ordered by position.
	Nameservers []VanityNameserver `json:"nameservers,omitempty"`
}

// VanityNameserverSetCreateRequest is the request body for creating a vanity NS set.
type VanityNameserverSetCreateRequest struct {
	// Name is the human-readable name for the set (1-255 characters).
	Name string `json:"name"`

	// ParentDomainName is the apex domain of the vanity NS zone; all Hostnames must be
	// subdomains of it.
	ParentDomainName string `json:"parent_domain_name"`

	// SOARName is the SOA RNAME stamped verbatim into the vanity NS zone.
	SOARName string `json:"soa_rname"`

	// Hostnames are the fully-qualified vanity NS hostnames, ordered by intended position.
	Hostnames []string `json:"hostnames"`
}

// VanityNameserverSetListResponse is the paginated response when listing vanity NS sets.
type VanityNameserverSetListResponse struct {
	// Results contains the vanity NS sets for the current page (includes non-active rows).
	Results []VanityNameserverSet `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// VanityNameserverSetDefaultResponse is the response when setting a set as the org default.
type VanityNameserverSetDefaultResponse struct {
	// VanityNameserverSet is the set that is now the organization's default.
	VanityNameserverSet VanityNameserverSet `json:"vanity_nameserver_set"`
}

// ClearVanityNameserverSetDefaultResponse is the response when clearing the org default.
type ClearVanityNameserverSetDefaultResponse struct {
	// Cleared is true if an active default was unset; false on an idempotent no-op.
	Cleared bool `json:"cleared"`
}

// ZonesReferencingSetResponse is the paginated response when listing zones referencing a set.
type ZonesReferencingSetResponse struct {
	// Results contains the zones whose apex is branded by the set.
	Results []Zone `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// VanityNsCheckRequest is the request body for the vanity NS diagnostic check.
type VanityNsCheckRequest struct {
	// SetID is the vanity NS set to diagnose.
	SetID VanityNameserverSetID `json:"set_id"`
}

// VanityNsCheckSummaryState is the top-level verdict synthesized from individual checks.
type VanityNsCheckSummaryState string

const (
	VanityNsCheckSummaryStateReady          VanityNsCheckSummaryState = "ready"
	VanityNsCheckSummaryStatePropagating    VanityNsCheckSummaryState = "propagating"
	VanityNsCheckSummaryStateActionRequired VanityNsCheckSummaryState = "action_required"
	VanityNsCheckSummaryStateDegraded       VanityNsCheckSummaryState = "degraded"
)

// VanityNsCheckStatus is the per-check verdict.
type VanityNsCheckStatus string

const (
	VanityNsCheckStatusPass VanityNsCheckStatus = "pass"
	VanityNsCheckStatusFail VanityNsCheckStatus = "fail"
	VanityNsCheckStatusWarn VanityNsCheckStatus = "warn"
	VanityNsCheckStatusInfo VanityNsCheckStatus = "info"
)

// VanityNsCheckSeverity is how much a check matters to the overall verdict.
type VanityNsCheckSeverity string

const (
	VanityNsCheckSeverityRequired    VanityNsCheckSeverity = "required"
	VanityNsCheckSeverityRecommended VanityNsCheckSeverity = "recommended"
	VanityNsCheckSeverityOptional    VanityNsCheckSeverity = "optional"
)

// VanityNsCheckSource is where a check observation came from.
type VanityNsCheckSource string

const (
	VanityNsCheckSourcePublicDNS        VanityNsCheckSource = "public_dns"
	VanityNsCheckSourceAuthoritativeDNS VanityNsCheckSource = "authoritative_dns"
	VanityNsCheckSourceRegistryEPP      VanityNsCheckSource = "registry_epp"
)

// VanityNsCheckConfidence is how authoritative a check observation is.
type VanityNsCheckConfidence string

const (
	VanityNsCheckConfidenceAuthoritative VanityNsCheckConfidence = "authoritative"
	VanityNsCheckConfidenceBestEffort    VanityNsCheckConfidence = "best_effort"
)

// VanityNsCheckSummary is the synthesized overall verdict of a vanity NS check.
type VanityNsCheckSummary struct {
	// State is the overall verdict synthesized from the checks.
	State VanityNsCheckSummaryState `json:"state"`

	// Detail is a customer-facing summary of the overall verdict.
	Detail string `json:"detail"`
}

// VanityNsCheckResult is an individual diagnostic check result.
type VanityNsCheckResult struct {
	// ID is the stable identifier for the individual check.
	ID string `json:"id"`

	// Label is the human-readable check name.
	Label string `json:"label"`

	// Status is the per-check verdict.
	Status VanityNsCheckStatus `json:"status"`

	// Severity is how much this check matters to the overall verdict.
	Severity VanityNsCheckSeverity `json:"severity"`

	// Source is where the observation came from.
	Source VanityNsCheckSource `json:"source"`

	// Confidence is how authoritative the observation is.
	Confidence VanityNsCheckConfidence `json:"confidence"`

	// Detail is a customer-facing explanation of the result.
	Detail string `json:"detail"`

	// Observed is a structured observation (e.g. addresses seen, mismatches).
	Observed map[string]interface{} `json:"observed,omitempty"`

	// Remediation is a suggested next step when the check did not pass.
	Remediation *string `json:"remediation,omitempty"`
}

// VanityNsCheckResponse is the response from the vanity NS diagnostic check.
type VanityNsCheckResponse struct {
	// SetID is the diagnosed set.
	SetID VanityNameserverSetID `json:"set_id"`

	// ParentDomainName is the parent domain of the set's vanity NS hostnames.
	ParentDomainName string `json:"parent_domain_name"`

	// Status is the lifecycle status of the set at check time.
	Status VanityNameserverSetStatus `json:"status"`

	// Summary is the synthesized overall verdict.
	Summary VanityNsCheckSummary `json:"summary"`

	// Checks are the individual diagnostic checks.
	Checks []VanityNsCheckResult `json:"checks,omitempty"`
}

// ListVanityNameserverSetsOptions contains options for listing vanity NS sets.
type ListVanityNameserverSetsOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of items per page.
	PageSize int
}
