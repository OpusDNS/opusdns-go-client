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
	DomainSortByUpdatedOn    DomainSortField = "updated_on"
	DomainSortByExpiresOn    DomainSortField = "expires_on"
	DomainSortByRegisteredOn DomainSortField = "registered_on"
)

// DomainClientStatus represents an EPP client status that can be set on a domain.
type DomainClientStatus string

const (
	DomainClientStatusTransferProhibited DomainClientStatus = "clientTransferProhibited"
	DomainClientStatusUpdateProhibited   DomainClientStatus = "clientUpdateProhibited"
	DomainClientStatusDeleteProhibited   DomainClientStatus = "clientDeleteProhibited"
	DomainClientStatusRenewProhibited    DomainClientStatus = "clientRenewProhibited"
	DomainClientStatusHold               DomainClientStatus = "clientHold"
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

	// Tags contains tags assigned to this domain when requested via include=tags.
	Tags []TagEnriched `json:"tags,omitempty"`

	// IsPremium indicates whether the domain has premium pricing.
	IsPremium bool `json:"is_premium,omitempty"`

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

	// TransferredOn is when the domain was last transferred to OpusDNS.
	TransferredOn *time.Time `json:"transferred_on,omitempty"`

	// ReadOnly indicates the domain is listed in the portfolio but cannot be
	// managed, for example while it awaits a migration. Set only by OpusDNS.
	ReadOnly bool `json:"read_only,omitempty"`

	// RenewalPeriod is the renewal period as an ISO 8601 duration (e.g. "P1Y").
	// Nil when the domain is not set to renew.
	RenewalPeriod *string `json:"renewal_period,omitempty"`

	// RenewalPrice is the price this organization pays to renew the domain.
	// Only present when renewal_price is requested via Include.
	RenewalPrice *DomainRenewalPrice `json:"renewal_price,omitempty"`

	// StatusTags are the status tags assigned to this domain. Only present
	// when tags are requested via Include.
	StatusTags []StatusTagResponse `json:"status_tags,omitempty"`

	// VerificationRequired carries pending verification claims and deadlines.
	VerificationRequired *VerificationRequired `json:"verification_required,omitempty"`

	// ConnectedAccount is the registrar credential or registry accreditation
	// the domain is synced from or managed on. Only present when
	// connected_account is requested via Include, and nil even then for
	// natively registered domains.
	ConnectedAccount *DomainConnectedAccount `json:"connected_account,omitempty"`

	// RegistrantChange is the registrant change in progress, or the completed
	// one whose transfer lock is still active. Only present when
	// registrant_change is requested via Include.
	RegistrantChange *DomainRegistrantChange `json:"registrant_change,omitempty"`
}

// DomainRenewalPrice is the price to renew a domain for one period.
type DomainRenewalPrice struct {
	// Price is the renewal price for Period, taxes excluded, in the
	// organization's billing currency.
	Price string `json:"price"`

	// Currency is the ISO 4217 currency code.
	Currency string `json:"currency"`

	// Period is the period the price covers.
	Period DomainPeriod `json:"period"`

	// IsPremium is true when the price came from the registry's premium
	// classification rather than standard TLD pricing.
	IsPremium bool `json:"is_premium"`
}

// VerificationRequired carries the verification a domain is waiting on.
type VerificationRequired struct {
	// Sources are the parties (ICANN, the registry) that require the
	// verification.
	Sources []DomainVerificationKind `json:"sources,omitempty"`

	// Claims are the outstanding verification claims.
	Claims []VerificationClaimType `json:"claims"`

	// Deadlines are the deadlines attached to those claims.
	Deadlines []VerificationDeadline `json:"deadlines,omitempty"`
}

// ConnectedAccountType says how a domain is connected to OpusDNS.
type ConnectedAccountType string

const (
	// ConnectedAccountTypeDefault is a domain registered natively through OpusDNS.
	ConnectedAccountTypeDefault ConnectedAccountType = "default"

	// ConnectedAccountTypeAMS is a domain on the organization's own registry
	// accreditation.
	ConnectedAccountTypeAMS ConnectedAccountType = "ams"

	// ConnectedAccountTypeRAS is a domain synced from a registrar credential.
	ConnectedAccountTypeRAS ConnectedAccountType = "ras"
)

// AccountTypeFilterMode controls how an account type filter matches.
type AccountTypeFilterMode string

const (
	// AccountTypeFilterMatchAny matches domains with any of the account types.
	AccountTypeFilterMatchAny AccountTypeFilterMode = "match_any"

	// AccountTypeFilterMatchNone matches domains with none of the account types.
	AccountTypeFilterMatchNone AccountTypeFilterMode = "match_none"
)

// Registrar identifies a registrar a registrar credential connects to.
type Registrar string

const (
	RegistrarInternetX           Registrar = "INTERNETX"
	RegistrarMoniker             Registrar = "MONIKER"
	RegistrarDomainBestellsystem Registrar = "DOMAIN_BESTELLSYSTEM"
	RegistrarCentralNic          Registrar = "CENTRALNIC"
	RegistrarNicManager          Registrar = "NICMANAGER"
	RegistrarOpusDNS             Registrar = "OPUSDNS"
	RegistrarEnom                Registrar = "ENOM"
	RegistrarOpenSRS             Registrar = "OPENSRS"
	RegistrarNicDirect           Registrar = "NIC_DIRECT"
	RegistrarOpenprovider        Registrar = "OPENPROVIDER"
	RegistrarRealtimeRegister    Registrar = "REALTIMEREGISTER"
)

// DomainConnectedAccount is the account a domain is synced from or managed on.
// Type selects the variant: "ras" for a registrar credential, "ams" for the
// organization's own registry accreditation. Only the fields of that variant
// are set.
type DomainConnectedAccount struct {
	// Type is the discriminator, ConnectedAccountTypeRAS or ConnectedAccountTypeAMS.
	Type ConnectedAccountType `json:"type"`

	// RegistrarCredentialID is set for the "ras" variant.
	RegistrarCredentialID TypeID `json:"registrar_credential_id,omitempty"`

	// Name is the name of the registrar credential, set for the "ras" variant.
	Name string `json:"name,omitempty"`

	// Registrar is set for the "ras" variant.
	Registrar Registrar `json:"registrar,omitempty"`

	// RegistryAccountID is set for the "ams" variant.
	RegistryAccountID TypeID `json:"registry_account_id,omitempty"`

	// Registry is the registry backend (e.g. "verisign", "denic"), set for
	// the "ams" variant.
	Registry string `json:"registry,omitempty"`
}

// RegistrantChangeStatus is the state of a registrant change.
type RegistrantChangeStatus string

const (
	RegistrantChangeStatusPendingConfirmation RegistrantChangeStatus = "pending_confirmation"
	RegistrantChangeStatusConfirmed           RegistrantChangeStatus = "confirmed"
	RegistrantChangeStatusExecuting           RegistrantChangeStatus = "executing"
	RegistrantChangeStatusRegistryPending     RegistrantChangeStatus = "registry_pending"
	RegistrantChangeStatusCompleted           RegistrantChangeStatus = "completed"
	RegistrantChangeStatusNotMaterial         RegistrantChangeStatus = "not_material"
	RegistrantChangeStatusExpired             RegistrantChangeStatus = "expired"
	RegistrantChangeStatusCanceled            RegistrantChangeStatus = "canceled"
	RegistrantChangeStatusRejected            RegistrantChangeStatus = "rejected"
	RegistrantChangeStatusFailed              RegistrantChangeStatus = "failed"
)

// RegistrantChangeConfirmationMethod says how a registrant change was confirmed.
type RegistrantChangeConfirmationMethod string

const (
	RegistrantChangeConfirmationDesignatedAgent RegistrantChangeConfirmationMethod = "designated_agent"
	RegistrantChangeConfirmationEmail           RegistrantChangeConfirmationMethod = "email"
	RegistrantChangeConfirmationAPI             RegistrantChangeConfirmationMethod = "api"
)

// DomainRegistrantChange is a change of a domain's registrant.
type DomainRegistrantChange struct {
	// RegistrantChangeID is the ID of the registrant change.
	RegistrantChangeID TypeID `json:"registrant_change_id"`

	// DomainID is the domain whose registrant changes.
	DomainID DomainID `json:"domain_id"`

	// Status is the state of the change.
	Status RegistrantChangeStatus `json:"status"`

	// Material is true when the name, organization or email of the
	// registrant changes.
	Material bool `json:"material"`

	// PreviousRegistrant is the registrant contact before the change.
	PreviousRegistrant ContactID `json:"previous_registrant"`

	// NewRegistrant is the registrant contact after the change.
	NewRegistrant ContactID `json:"new_registrant"`

	// ConfirmationMethod is how the change was confirmed; nil when no
	// confirmation applies.
	ConfirmationMethod *RegistrantChangeConfirmationMethod `json:"confirmation_method,omitempty"`

	// ConfirmationDeadline is when the change expires if not confirmed.
	ConfirmationDeadline *time.Time `json:"confirmation_deadline,omitempty"`

	// TransferLockOptOut is true when the transfer lock after the change was
	// opted out of.
	TransferLockOptOut bool `json:"transfer_lock_opt_out"`

	// TransferLockUntil is until when the domain cannot be transferred away
	// because of the change; nil when no transfer lock applies.
	TransferLockUntil *time.Time `json:"transfer_lock_until,omitempty"`

	// FailureReason says why the change failed.
	FailureReason *string `json:"failure_reason,omitempty"`

	// CreatedOn is when the change was requested.
	CreatedOn *time.Time `json:"created_on,omitempty"`
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

// DomainSummary is the domain summary of an organization.
type DomainSummary struct {
	// OrganizationID is the organization the summary covers.
	OrganizationID OrganizationID `json:"organization_id"`

	// Domains holds the counts.
	Domains DomainSummaryData `json:"domains"`
}

// DomainSummaryData holds the domain counts of a DomainSummary.
type DomainSummaryData struct {
	// TotalCount is the total number of domains, including sub-organizations.
	TotalCount int `json:"total_count"`

	// ByStatus counts domains per status.
	ByStatus map[DomainStatus]int `json:"by_status"`

	// ByStatusTag counts domains per status tag, listing only tags with at
	// least one domain.
	ByStatusTag map[StatusTagType]int `json:"by_status_tag"`

	// ByTLD counts domains per TLD.
	ByTLD map[string]int `json:"by_tld"`

	// ByAccountType counts domains per connected account type.
	ByAccountType map[ConnectedAccountType]int `json:"by_account_type"`

	// ByOrganization counts domains per organization name.
	ByOrganization map[string]int `json:"by_organization,omitempty"`

	// ExpiringSoon counts domains by how soon they expire.
	ExpiringSoon *DomainsExpiringSoon `json:"expiring_soon,omitempty"`
}

// DomainsExpiringSoon counts domains expiring within the coming days.
type DomainsExpiringSoon struct {
	// Next30Days is the number of domains expiring in the next 30 days.
	Next30Days int `json:"next_30_days"`

	// Next60Days is the number of domains expiring in the next 60 days.
	Next60Days int `json:"next_60_days"`

	// Next90Days is the number of domains expiring in the next 90 days.
	Next90Days int `json:"next_90_days"`
}

// DomainAttributeKey names an attribute in the Attributes of a domain create,
// transfer or update request. Attribute values are strings; boolean keys take
// "true".
type DomainAttributeKey string

// The attribute keys a customer may set. The API rejects the keys it writes
// itself, such as verification_required.
const (
	// DomainAttributeAutoRenewPeriod selects the period of the next renewal:
	// AutoRenewPeriodMonthly or AutoRenewPeriodYearly. All TLDs, on create,
	// transfer and update.
	DomainAttributeAutoRenewPeriod DomainAttributeKey = "auto_renew_period"

	// DomainAttributeMusicRegistrantAttestation is required to register a
	// .music domain.
	DomainAttributeMusicRegistrantAttestation DomainAttributeKey = "music_registrant_attestation"

	// DomainAttributeNicITComplianceConfirmation is required to register or
	// transfer a .it domain.
	DomainAttributeNicITComplianceConfirmation DomainAttributeKey = "nic_it_compliance_confirmation"

	// DomainAttributeTravelIndustryAcknowledgement is required to register a
	// .travel domain.
	DomainAttributeTravelIndustryAcknowledgement DomainAttributeKey = "travel_industry_acknowledgement"

	// DomainAttributeInternetEERegistrantAgreement is required to register a
	// .ee domain.
	DomainAttributeInternetEERegistrantAgreement DomainAttributeKey = "internet_ee_registrant_agreement"

	// DomainAttributePunktumDKTermsAcceptance is required to register a .dk
	// domain and to change its registrant.
	DomainAttributePunktumDKTermsAcceptance DomainAttributeKey = "punktum_dk_terms_acceptance"

	// DomainAttributeDEGeneralRequestContact is the ID of a contact whose .de
	// attribute set has DE_CONTACT_TYPE REQUEST. .de only.
	DomainAttributeDEGeneralRequestContact DomainAttributeKey = "de_general_request_contact"

	// DomainAttributeDEAbuseContact is the ID of a contact whose .de attribute
	// set has DE_CONTACT_TYPE REQUEST. .de only.
	DomainAttributeDEAbuseContact DomainAttributeKey = "de_abuse_contact"
)

// Values of DomainAttributeAutoRenewPeriod.
const (
	AutoRenewPeriodMonthly = "monthly"
	AutoRenewPeriodYearly  = "yearly"
)

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

	// ClaimsNoticeAcceptanceHash confirms that the registrant accepted the
	// trademark claims notice of the domain. Required while the TLD is in its
	// claims period and the name matches a trademark; take it from
	// ClaimsNotice.ClaimsNoticeAcceptanceHash (DomainsService.GetClaimsNotices).
	ClaimsNoticeAcceptanceHash *string `json:"claims_notice_acceptance_hash,omitempty"`

	// ExpectedPrice confirms the price of a premium domain, as a decimal string
	// with two decimal places (e.g. "1250.00"). The API rejects the
	// registration when the price differs.
	ExpectedPrice *string `json:"expected_price,omitempty"`

	// Attributes are additional attributes of the domain. Some TLDs require
	// one at registration, for example DomainAttributeNicITComplianceConfirmation
	// for .it.
	Attributes map[DomainAttributeKey]string `json:"attributes,omitempty"`
}

// ContactHandle represents a contact reference with optional attributes.
type ContactHandle struct {
	// ContactID is the ID of the contact.
	ContactID ContactID `json:"contact_id"`

	// Attributes contains additional contact attributes for this domain.
	Attributes map[string]string `json:"attributes,omitempty"`
}

// StatusChanges represents add/remove semantics for domain client statuses.
// Cannot be used together with Statuses in DomainUpdateRequest.
type StatusChanges struct {
	// Add is a list of client statuses to add.
	Add []DomainClientStatus `json:"add,omitempty"`

	// Remove is a list of client statuses to remove.
	Remove []DomainClientStatus `json:"remove,omitempty"`
}

// DomainUpdateRequest represents a request to update a domain.
type DomainUpdateRequest struct {
	// Contacts maps contact types to contact handles to update.
	Contacts map[DomainContactType][]ContactHandle `json:"contacts,omitempty"`

	// Nameservers is the new list of nameservers.
	Nameservers []Nameserver `json:"nameservers,omitempty"`

	// RenewalMode updates the renewal mode (renew or expire).
	RenewalMode *RenewalMode `json:"renewal_mode,omitempty"`

	// Statuses is the complete list of client statuses to set on the domain.
	// This replaces the entire client status list.
	// Mutually exclusive with StatusChanges.
	Statuses []DomainClientStatus `json:"statuses,omitempty"`

	// StatusChanges provides add/remove semantics for client statuses.
	// Mutually exclusive with Statuses.
	StatusChanges *StatusChanges `json:"status_changes,omitempty"`

	// AuthCode sets a new auth code for the domain.
	AuthCode *string `json:"auth_code,omitempty"`

	// Attributes sets attributes of the domain, for example
	// DomainAttributeAutoRenewPeriod, or DomainAttributePunktumDKTermsAcceptance
	// to confirm a .dk registrant change. An empty value removes the stored
	// attribute. Attributes the registrant accepted at registration cannot be
	// changed.
	Attributes map[DomainAttributeKey]string `json:"attributes,omitempty"`
}

// DomainTransferRequest represents a request to transfer a domain.
type DomainTransferRequest struct {
	// Name is the domain name to transfer.
	Name string `json:"name"`

	// AuthCode is the authorization code for the transfer. Leave it empty for
	// registries that do not use one; it is then not sent.
	AuthCode string `json:"auth_code,omitempty"`

	// Contacts maps contact types to contact handles.
	Contacts map[DomainContactType][]ContactHandle `json:"contacts,omitempty"`

	// RenewalMode sets the renewal mode (renew or expire).
	RenewalMode RenewalMode `json:"renewal_mode"`

	// Nameservers is the list of nameservers to use after transfer.
	Nameservers []Nameserver `json:"nameservers,omitempty"`

	// Period is the registration period to add when the transfer completes.
	// Nil applies the registry's default.
	Period *DomainPeriod `json:"period,omitempty"`

	// CreateZone creates a zone for the domain on OpusDNS nameserver infrastructure.
	CreateZone bool `json:"create_zone,omitempty"`

	// ExpectedPrice confirms the price of a premium domain, as a decimal string
	// with two decimal places (e.g. "1250.00").
	ExpectedPrice *string `json:"expected_price,omitempty"`

	// Attributes are additional attributes of the domain, for example
	// DomainAttributeNicITComplianceConfirmation for a .it transfer.
	Attributes map[DomainAttributeKey]string `json:"attributes,omitempty"`
}

// DomainRenewRequest represents a request to renew a domain.
type DomainRenewRequest struct {
	// Period is how long to extend the registration.
	Period DomainPeriod `json:"period"`

	// CurrentExpiryDate is the domain's current expiration date, exactly as
	// the API reports it in Domain.ExpiresOn. Required: the API rejects the
	// renewal when it does not match, so a renewal is not applied twice.
	CurrentExpiryDate *time.Time `json:"current_expiry_date"`

	// ExpectedPrice confirms the price of a premium domain, as a decimal string
	// with two decimal places (e.g. "1250.00").
	ExpectedPrice *string `json:"expected_price,omitempty"`
}

// DomainRenewStatusTransferPending is the Status of a DomainRenewResult whose
// renewal became a transfer into OpusDNS.
const DomainRenewStatusTransferPending = "transfer_pending"

// DomainRenewResult is the outcome of a renewal. The API answers in one of two
// shapes, and only the fields of that shape are set:
//
//   - The domain was renewed: NewExpiryDate and PeriodExtended are set.
//   - The domain is held at another registrar, so the renewal transfers it
//     into OpusDNS: Status is DomainRenewStatusTransferPending and RequestID
//     identifies the request. The transfer's progress is reported as domain
//     events. Use TransferPending to tell the two apart.
type DomainRenewResult struct {
	// Name is the domain name.
	Name string `json:"name"`

	// NewExpiryDate is the expiration date after the renewal.
	NewExpiryDate *time.Time `json:"new_expiry_date,omitempty"`

	// PeriodExtended is the period the registration was extended by.
	PeriodExtended *DomainPeriod `json:"period_extended,omitempty"`

	// Status is DomainRenewStatusTransferPending when the renewal started a
	// transfer into OpusDNS, empty otherwise.
	Status string `json:"status,omitempty"`

	// RequestID identifies the transfer request when Status is set.
	RequestID string `json:"request_id,omitempty"`
}

// TransferPending reports whether the renewal started a transfer into OpusDNS
// instead of renewing the domain directly.
func (r *DomainRenewResult) TransferPending() bool {
	return r.Status == DomainRenewStatusTransferPending
}

// AuthCodeDelivery says how the auth code of a domain is delivered.
type AuthCodeDelivery string

const (
	// AuthCodeDeliveryResponse means the code is in the API response.
	AuthCodeDeliveryResponse AuthCodeDelivery = "response"

	// AuthCodeDeliveryRegistrantEmail means the registry emails the code to
	// the registrant; the API response does not carry it.
	AuthCodeDeliveryRegistrantEmail AuthCodeDelivery = "registrant_email"
)

// DomainAuthCodeResponse is the result of requesting a domain's auth code.
type DomainAuthCodeResponse struct {
	// Name is the domain name.
	Name string `json:"name"`

	// Delivery says where the auth code went.
	Delivery AuthCodeDelivery `json:"delivery"`

	// AuthCode is the auth code; set only when Delivery is
	// AuthCodeDeliveryResponse.
	AuthCode *string `json:"auth_code,omitempty"`

	// AuthCodeExpiresOn is when the auth code expires, if the registry
	// reports it.
	AuthCodeExpiresOn *time.Time `json:"auth_code_expires_on,omitempty"`

	// Recipients are the addresses the code was sent to when Delivery is
	// AuthCodeDeliveryRegistrantEmail.
	Recipients []string `json:"recipients,omitempty"`
}

// DomainRestoreRequest represents a request to restore a deleted domain. All
// fields are optional.
type DomainRestoreRequest struct {
	// Period is not part of the restore request.
	//
	// Deprecated: The API has no restore period and ignored this field. It is
	// no longer sent.
	Period int `json:"-"`

	// Reason is the reason for restoring the domain.
	Reason *string `json:"reason,omitempty"`

	// AdditionalInfo is additional information for the restore.
	AdditionalInfo *string `json:"additional_info,omitempty"`

	// ExpectedPrice confirms the price of a premium domain, as a decimal string
	// with two decimal places (e.g. "1250.00").
	ExpectedPrice *string `json:"expected_price,omitempty"`
}

// DomainRestoreResult is the outcome of restoring a domain from redemption.
type DomainRestoreResult struct {
	// DomainID is the ID of the restored domain.
	DomainID DomainID `json:"domain_id"`

	// Name is the domain name.
	Name string `json:"name"`

	// RestoredAt is when the domain was restored.
	RestoredAt *time.Time `json:"restored_at,omitempty"`
}

// DomainDNSSECRequest represents a legacy request to configure DNSSEC for a domain.
//
// Deprecated: Use DomainDNSSECDataCreate with DomainsService.PutDNSSEC.
type DomainDNSSECRequest struct {
	// DSRecords contains DS records to add.
	DSRecords []DSRecord `json:"ds_records,omitempty"`

	// DNSKEYRecords contains DNSKEY records to add.
	DNSKEYRecords []DNSKEYRecord `json:"dnskey_records,omitempty"`
}

// DNSSECRecordType represents a domain DNSSEC record type.
type DNSSECRecordType string

const (
	DNSSECRecordTypeDSData  DNSSECRecordType = "ds_data"
	DNSSECRecordTypeKeyData DNSSECRecordType = "key_data"
)

// DNSSECAlgorithm represents a DNSSEC algorithm number.
type DNSSECAlgorithm int

// DNSSECDigestType represents a DNSSEC digest type.
type DNSSECDigestType int

// DomainDNSSECDataCreate represents DNSSEC data for create/update requests.
type DomainDNSSECDataCreate struct {
	RecordType DNSSECRecordType  `json:"record_type"`
	Algorithm  DNSSECAlgorithm   `json:"algorithm"`
	Digest     *string           `json:"digest,omitempty"`
	DigestType *DNSSECDigestType `json:"digest_type,omitempty"`
	Flags      *int              `json:"flags,omitempty"`
	KeyTag     *int              `json:"key_tag,omitempty"`
	Protocol   *int              `json:"protocol,omitempty"`
	PublicKey  *string           `json:"public_key,omitempty"`
}

// DomainDNSSECDataResponse represents DNSSEC data returned by the API.
type DomainDNSSECDataResponse struct {
	DomainDNSSECDataID TypeID            `json:"domain_dnssec_data_id,omitempty"`
	DomainID           DomainID          `json:"domain_id,omitempty"`
	RecordType         DNSSECRecordType  `json:"record_type"`
	Algorithm          DNSSECAlgorithm   `json:"algorithm"`
	Digest             *string           `json:"digest,omitempty"`
	DigestType         *DNSSECDigestType `json:"digest_type,omitempty"`
	Flags              *int              `json:"flags,omitempty"`
	KeyTag             *int              `json:"key_tag,omitempty"`
	Protocol           *int              `json:"protocol,omitempty"`
	PublicKey          *string           `json:"public_key,omitempty"`
	CreatedOn          *time.Time        `json:"created_on,omitempty"`
	UpdatedOn          *time.Time        `json:"updated_on,omitempty"`
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

	// TagIDs filters by tag IDs. Multiple values are sent as repeated tag_ids params.
	TagIDs []TagID

	// TagMode controls whether any or all tag IDs must match.
	TagMode TagFilterMode

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

	// IsPremium filters by premium status.
	IsPremium *bool

	// CreatedAfter filters domains created after this date.
	CreatedAfter *time.Time

	// CreatedBefore filters domains created before this date.
	CreatedBefore *time.Time

	// UpdatedAfter filters domains updated after this date.
	UpdatedAfter *time.Time

	// UpdatedBefore filters domains updated before this date.
	UpdatedBefore *time.Time

	// ExpiresAfter filters domains expiring after this date.
	ExpiresAfter *time.Time

	// ExpiresBefore filters domains expiring before this date.
	ExpiresBefore *time.Time

	// ExpiresIn30Days filters domains expiring within 30 days.
	ExpiresIn30Days *bool

	// ExpiresIn60Days filters domains expiring within 60 days.
	ExpiresIn60Days *bool

	// ExpiresIn90Days filters domains expiring within 90 days.
	ExpiresIn90Days *bool

	// RegisteredAfter filters domains registered after this date.
	RegisteredAfter *time.Time

	// RegisteredBefore filters domains registered before this date.
	RegisteredBefore *time.Time

	// RegistryStatuses filters by registry statuses.
	RegistryStatuses []string

	// Include requests additional response data.
	Include []DomainIncludeField

	// StatusTags filters by status tag. Multiple values are sent as repeated
	// status_tags params.
	StatusTags []StatusTagType

	// StatusTagMode controls whether any, all or none of StatusTags must match.
	StatusTagMode TagFilterMode

	// ReadOnly filters by whether the domain is read-only in OpusDNS.
	ReadOnly *bool

	// TransferredAfter filters domains transferred in after this date.
	TransferredAfter *time.Time

	// TransferredBefore filters domains transferred in before this date.
	TransferredBefore *time.Time

	// HasTags filters by whether the domain has any tags.
	HasTags *bool

	// Nameserver filters by a nameserver hostname the domain delegates to.
	Nameserver string

	// AccountTypes filters by connected account type. Multiple values are
	// sent as repeated account_type params.
	AccountTypes []ConnectedAccountType

	// AccountTypeMode controls whether any or none of AccountTypes must match.
	AccountTypeMode AccountTypeFilterMode

	// Registrars filters by the registrar of the domain's registrar
	// credential. Multiple values are sent as repeated registrar params.
	Registrars []Registrar

	// RegistrarCredentialIDs filters by registrar credential. Multiple values
	// are sent as repeated registrar_credential_id params.
	RegistrarCredentialIDs []TypeID
}

// DomainIncludeField represents optional domain response expansions.
type DomainIncludeField string

const (
	// DomainIncludeTags includes tags assigned to the domain.
	DomainIncludeTags DomainIncludeField = "tags"

	// DomainIncludeRenewalPrice includes the renewal price. GetDomainWithOptions only.
	DomainIncludeRenewalPrice DomainIncludeField = "renewal_price"

	// DomainIncludeConnectedAccount includes the connected account.
	DomainIncludeConnectedAccount DomainIncludeField = "connected_account"

	// DomainIncludeRegistrantChange includes the registrant change in
	// progress. GetDomainWithOptions only.
	DomainIncludeRegistrantChange DomainIncludeField = "registrant_change"
)

// DomainVerificationKind is the party that requires a domain verification.
type DomainVerificationKind string

const (
	DomainVerificationKindICANN    DomainVerificationKind = "ICANN"
	DomainVerificationKindRegistry DomainVerificationKind = "REGISTRY"
)

// DomainVerificationStatus is the state of a domain verification case.
type DomainVerificationStatus string

const (
	DomainVerificationStatusOpen      DomainVerificationStatus = "OPEN"
	DomainVerificationStatusSubmitted DomainVerificationStatus = "SUBMITTED"
	DomainVerificationStatusOverdue   DomainVerificationStatus = "OVERDUE"
	DomainVerificationStatusSatisfied DomainVerificationStatus = "SATISFIED"
	DomainVerificationStatusFailed    DomainVerificationStatus = "FAILED"
	DomainVerificationStatusCanceled  DomainVerificationStatus = "CANCELED"
)

// DomainVerificationTrigger is what opened a domain verification case.
type DomainVerificationTrigger string

const (
	DomainVerificationTriggerCreate           DomainVerificationTrigger = "CREATE"
	DomainVerificationTriggerRegistrantChange DomainVerificationTrigger = "REGISTRANT_CHANGE"
	DomainVerificationTriggerTransfer         DomainVerificationTrigger = "TRANSFER"
	DomainVerificationTriggerRegistryRequest  DomainVerificationTrigger = "REGISTRY_REQUEST"
	DomainVerificationTriggerBackfill         DomainVerificationTrigger = "BACKFILL"
)

// DomainVerificationCloseReason is why a domain verification case was closed.
type DomainVerificationCloseReason string

const (
	DomainVerificationCloseEvidenceVerified DomainVerificationCloseReason = "EVIDENCE_VERIFIED"
	DomainVerificationCloseRegistryCleared  DomainVerificationCloseReason = "REGISTRY_CLEARED"
	DomainVerificationCloseRegistryRejected DomainVerificationCloseReason = "REGISTRY_REJECTED"
	DomainVerificationCloseSuperseded       DomainVerificationCloseReason = "SUPERSEDED"
	DomainVerificationCloseDomainDeleted    DomainVerificationCloseReason = "DOMAIN_DELETED"
	DomainVerificationCloseTransferredOut   DomainVerificationCloseReason = "TRANSFERRED_OUT"
	DomainVerificationCloseTransferAborted  DomainVerificationCloseReason = "TRANSFER_ABORTED"
)

// DomainVerification is one verification case of a domain.
type DomainVerification struct {
	// VerificationID is the ID of the verification case.
	VerificationID TypeID `json:"verification_id"`

	// Source is the party that requires the verification.
	Source DomainVerificationKind `json:"source"`

	// Trigger is what opened the case.
	Trigger DomainVerificationTrigger `json:"trigger"`

	// Status is the state of the case.
	Status DomainVerificationStatus `json:"status"`

	// Claims are the claims that must be verified.
	Claims []VerificationClaimType `json:"claims"`

	// Deadlines are the deadlines attached to the case.
	Deadlines []VerificationDeadline `json:"deadlines"`

	// ContactID is the contact being verified.
	ContactID ContactID `json:"contact_id"`

	// OpenedOn is when the case was opened.
	OpenedOn *time.Time `json:"opened_on,omitempty"`

	// ClosedOn is when the case was closed; nil while it is open.
	ClosedOn *time.Time `json:"closed_on,omitempty"`

	// CloseReason is why the case was closed; nil while it is open.
	CloseReason *DomainVerificationCloseReason `json:"close_reason,omitempty"`
}

// DomainVerificationListResponse is a page of domain verification cases.
type DomainVerificationListResponse struct {
	// Results contains the verification cases of the current page.
	Results []DomainVerification `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// ListDomainVerificationsOptions contains options for listing the
// verification cases of a domain.
type ListDomainVerificationsOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of cases per page.
	PageSize int
}

// GetDomainOptions contains options for retrieving a domain.
type GetDomainOptions struct {
	// Include requests additional response data.
	Include []DomainIncludeField
}

// OutboundTransferAction resolves a pending outbound transfer.
type OutboundTransferAction string

const (
	// OutboundTransferApprove releases the domain to the gaining registrar.
	OutboundTransferApprove OutboundTransferAction = "approve"

	// OutboundTransferReject keeps the domain and rejects the transfer.
	OutboundTransferReject OutboundTransferAction = "reject"
)

// OutboundTransferRequest resolves a pending outbound transfer.
type OutboundTransferRequest struct {
	// Action says whether to approve or reject the transfer.
	Action OutboundTransferAction `json:"action"`
}

// OutboundTransferResponse reports how an outbound transfer was resolved.
type OutboundTransferResponse struct {
	// DomainID is the domain whose transfer was resolved.
	DomainID DomainID `json:"domain_id"`

	// DomainName is the name of that domain.
	DomainName string `json:"domain_name"`

	// Action is the action that was applied.
	Action OutboundTransferAction `json:"action"`
}
