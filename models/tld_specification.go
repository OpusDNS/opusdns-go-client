package models

import "time"

// TLDSpecification is the specification of a TLD (TldSpecificationResponse):
// the rules the registry and OpusDNS apply to its domains. Durations such as
// grace periods are ISO 8601 strings (for example "P5D" or "P1Y").
type TLDSpecification struct {
	// TLDs are the TLDs the specification covers. One specification can serve
	// several TLDs, for example a ccTLD and its second-level zones.
	TLDs []TLDInfo `json:"tlds"`

	// Characters is the allowed length of the second-level label.
	Characters TLDRange `json:"characters"`

	// IDN describes support for internationalized domain names.
	IDN TLDIDN `json:"idn"`

	// DNSConfiguration describes nameserver and DNSSEC rules.
	DNSConfiguration TLDDNSConfiguration `json:"dns_configuration"`

	// DomainLifecycle describes registration periods and grace periods.
	DomainLifecycle TLDDomainLifecycle `json:"domain_lifecycle"`

	// DomainStatuses lists the domain statuses the registry supports.
	DomainStatuses TLDDomainStatuses `json:"domain_statuses"`

	// TransferPolicies describes auth codes and transfer behaviour.
	TransferPolicies TLDTransferPolicies `json:"transfer_policies"`

	// Contacts describes contact roles and attributes, or nil.
	Contacts *TLDContacts `json:"contacts,omitempty"`

	// LaunchPhases describes sunrise, landrush and similar phases, or nil.
	LaunchPhases *TLDLaunchPhases `json:"launch_phases,omitempty"`

	// LegalRequirements lists notices and confirmations operations need.
	LegalRequirements []TLDLegalRequirement `json:"legal_requirements,omitempty"`

	// LocalPresence describes local presence requirements, or nil.
	LocalPresence *TLDLocalPresence `json:"local_presence,omitempty"`

	// ParkingEnabled says whether domains of the TLD can be parked.
	ParkingEnabled bool `json:"parking_enabled"`

	// PremiumDomains describes premium domain support, or nil.
	PremiumDomains *TLDPremiumDomains `json:"premium_domains,omitempty"`

	// RDAP describes the registry's RDAP service, or nil.
	RDAP *TLDRDAP `json:"rdap,omitempty"`

	// RegistryLock describes registry lock support, or nil.
	RegistryLock *TLDRegistryLock `json:"registry_lock,omitempty"`

	// ReservedDomains describes the registry's reserved names, or nil.
	ReservedDomains *TLDReservedDomains `json:"reserved_domains,omitempty"`

	// VerificationPolicies maps a policy type to its contact verification
	// policy.
	VerificationPolicies map[string]TLDVerificationPolicy `json:"verification_policies,omitempty"`

	// Whois describes the registry's WHOIS service, or nil.
	Whois *TLDWhois `json:"whois,omitempty"`
}

// TLDRange is an inclusive range, such as a label length or a number of
// nameservers.
type TLDRange struct {
	// Min is the lower bound.
	Min int `json:"min"`

	// Max is the upper bound.
	Max int `json:"max"`
}

// TLDIDN describes support for internationalized domain names.
type TLDIDN struct {
	// IDNCapable says whether the TLD accepts IDNs.
	IDNCapable bool `json:"idn_capable"`

	// IDNTables lists the IDN tables (scripts or languages) the TLD accepts.
	IDNTables []string `json:"idn_tables,omitempty"`
}

// TLDDNSConfiguration describes nameserver and DNSSEC rules of a TLD.
type TLDDNSConfiguration struct {
	// AllowedNumberOfNameservers is how many nameservers a domain may have.
	AllowedNumberOfNameservers TLDRange `json:"allowed_number_of_nameserver"`

	// CZDSAvailable says whether the zone file is available through ICANN's
	// Centralized Zone Data Service.
	CZDSAvailable *bool `json:"czds_available,omitempty"`

	// DNSSECAllowed says whether DNSSEC can be enabled.
	DNSSECAllowed bool `json:"dnssec_allowed"`

	// DNSSECMandatory says whether DNSSEC is required.
	DNSSECMandatory *bool `json:"dnssec_mandatory,omitempty"`

	// DNSSECMode is how DNSSEC data is submitted to the registry: "DS" or
	// "DNSKEY".
	DNSSECMode *string `json:"dnssec_mode,omitempty"`

	// DNSSECSubmitDigestTypes are the DS digest types the registry accepts.
	DNSSECSubmitDigestTypes []int `json:"dnssec_submit_digest_types,omitempty"`

	// HostObjects says whether the registry uses host objects.
	HostObjects bool `json:"host_objects"`

	// HostParentCheckTLDs lists TLDs whose hosts the registry checks for a
	// parent domain.
	HostParentCheckTLDs []string `json:"host_parent_check_tlds,omitempty"`

	// RegistryNameserverCheck says whether the registry checks nameservers
	// before accepting them.
	RegistryNameserverCheck bool `json:"registry_nameserver_check"`

	// RegistryRootNameserverUpdate is the registry's root nameserver update
	// interval.
	RegistryRootNameserverUpdate int `json:"registry_root_nameserver_update"`
}

// TLDDomainLifecycle describes registration periods and grace periods of a
// TLD. Durations are ISO 8601 strings.
type TLDDomainLifecycle struct {
	// AddGracePeriod is the add grace period.
	AddGracePeriod *string `json:"add_grace_period,omitempty"`

	// AutoRenewBeforeExpiration is how long before expiry a domain is renewed.
	AutoRenewBeforeExpiration *string `json:"auto_renew_before_expiration,omitempty"`

	// DefaultTransferRenewalPeriod is the renewal period a transfer applies
	// by default.
	DefaultTransferRenewalPeriod *DomainPeriod `json:"default_transfer_renewal_period,omitempty"`

	// DeleteBeforeExpiration is how long before expiry a domain that is not
	// renewed is deleted.
	DeleteBeforeExpiration *string `json:"delete_before_expiration,omitempty"`

	// DeletePolicy lists the delete policies of the TLD.
	DeletePolicy []string `json:"delete_policy,omitempty"`

	// ExplicitRenew says whether the registry needs an explicit renew command.
	ExplicitRenew *bool `json:"explicit_renew,omitempty"`

	// GracePeriod is the grace period after expiry.
	GracePeriod *string `json:"grace_period,omitempty"`

	// PendingDelete is how long a domain stays pending delete.
	PendingDelete *string `json:"pending_delete,omitempty"`

	// RedemptionPeriod is the redemption period.
	RedemptionPeriod *string `json:"redemption_period,omitempty"`

	// RegistrationPeriods are the periods a domain can be registered for.
	RegistrationPeriods []DomainPeriod `json:"registration_periods,omitempty"`

	// RegistryAutoRenew says whether the registry renews domains itself.
	RegistryAutoRenew *bool `json:"registry_auto_renew,omitempty"`

	// RenewalPeriods are the periods a domain can be renewed for.
	RenewalPeriods []DomainPeriod `json:"renewal_periods,omitempty"`

	// RestoreTriggersRenewal says whether a restore also renews the domain.
	RestoreTriggersRenewal bool `json:"restore_triggers_renewal"`

	// RGPOperations describes the registry's redemption grace period commands.
	RGPOperations *TLDRGPOperations `json:"rgp_operations,omitempty"`

	// SyncAfterOperations lists operations after which the domain is synced
	// from the registry.
	SyncAfterOperations []string `json:"sync_after_operations,omitempty"`

	// TransferGracePeriod is the transfer grace period.
	TransferGracePeriod *string `json:"transfer_grace_period,omitempty"`

	// TransferRenewalPeriods are the renewal periods a transfer can apply.
	TransferRenewalPeriods []DomainPeriod `json:"transfer_renewal_periods,omitempty"`
}

// TLDRGPOperations describes the registry's redemption grace period commands.
type TLDRGPOperations struct {
	// Report says whether the registry needs a restore report.
	Report bool `json:"report"`

	// Request says whether the registry needs a restore request.
	Request bool `json:"request"`
}

// TLDDomainStatuses lists the domain statuses a registry supports.
type TLDDomainStatuses struct {
	// Default are the statuses a new domain gets.
	Default []string `json:"default"`

	// StatusMapping maps a registry status to the statuses it stands for.
	StatusMapping map[string][]DomainStatus `json:"status_mapping,omitempty"`

	// SupportedStatuses are the statuses the registry supports.
	SupportedStatuses []string `json:"supported_statuses"`
}

// TLDTransferPolicies describes auth codes and transfer behaviour of a TLD.
// Durations are ISO 8601 strings.
type TLDTransferPolicies struct {
	// AuthinfoMaxLength is the maximum auth code length.
	AuthinfoMaxLength *int `json:"authinfo_max_length,omitempty"`

	// AuthinfoMinLength is the minimum auth code length.
	AuthinfoMinLength *int `json:"authinfo_min_length,omitempty"`

	// AuthinfoPattern is a regular expression auth codes must match.
	AuthinfoPattern *string `json:"authinfo_pattern,omitempty"`

	// AuthinfoRequired says whether a transfer needs an auth code.
	AuthinfoRequired bool `json:"authinfo_required"`

	// AuthinfoSetByRegistrar says whether the registrar sets the auth code.
	AuthinfoSetByRegistrar bool `json:"authinfo_set_by_registrar"`

	// AuthinfoTimeLimited says whether auth codes expire.
	AuthinfoTimeLimited *bool `json:"authinfo_time_limited,omitempty"`

	// AuthinfoValidityPeriod is how long an auth code stays valid.
	AuthinfoValidityPeriod *string `json:"authinfo_validity_period,omitempty"`

	// ContactsInTransferCommand says whether contacts go into the transfer
	// command.
	ContactsInTransferCommand bool `json:"contacts_in_transfer_command"`

	// InfoContactAuthinfo says whether contact info returns the auth code.
	InfoContactAuthinfo *bool `json:"info_contact_authinfo,omitempty"`

	// InfoDomainAuthinfo says whether domain info returns the auth code.
	InfoDomainAuthinfo *bool `json:"info_domain_authinfo,omitempty"`

	// PostTransferRequirements lists what has to happen after a transfer.
	PostTransferRequirements []string `json:"post_transfer_requirements,omitempty"`

	// TransferAck is how a transfer is acknowledged.
	TransferAck *string `json:"transfer_ack,omitempty"`

	// TransferEmailRequired says whether a transfer needs email confirmation.
	TransferEmailRequired *bool `json:"transfer_email_required,omitempty"`

	// TransferLockPolicy describes transfer lock support.
	TransferLockPolicy TLDTransferLockPolicy `json:"transfer_lock_policy"`

	// TransferNack is how a transfer is rejected.
	TransferNack *string `json:"transfer_nack,omitempty"`

	// TransferRenewalPeriod is the renewal a transfer applies.
	TransferRenewalPeriod *string `json:"transfer_renewal_period,omitempty"`

	// TransferRenewsDomain says whether a transfer renews the domain.
	TransferRenewsDomain *bool `json:"transfer_renews_domain,omitempty"`

	// TransferTime is how long a transfer takes.
	TransferTime *string `json:"transfer_time,omitempty"`
}

// TLDTransferLockPolicy describes transfer lock support of a TLD.
type TLDTransferLockPolicy struct {
	// SupportedByRegistrar says whether OpusDNS can lock transfers.
	SupportedByRegistrar bool `json:"supported_by_registrar"`

	// SupportedByRegistry says whether the registry can lock transfers.
	SupportedByRegistry bool `json:"supported_by_registry"`
}

// TLDContacts describes contact roles and attributes of a TLD.
type TLDContacts struct {
	// AuthinfoRequired says whether contacts need an auth code.
	AuthinfoRequired *bool `json:"authinfo_required,omitempty"`

	// AuthinfoSupported says whether contacts can have an auth code.
	AuthinfoSupported bool `json:"authinfo_supported"`

	// IsThick says whether the registry keeps contact data (thick registry).
	IsThick *bool `json:"is_thick,omitempty"`

	// PossibleAttributes are the TLD-specific contact attributes.
	PossibleAttributes []ContactAttributeDefinition `json:"possible_attributes,omitempty"`

	// PrivacyProxy says whether a privacy proxy is allowed.
	PrivacyProxy *bool `json:"privacy_proxy,omitempty"`

	// RegistrantChange is how a registrant change is carried out.
	RegistrantChange *string `json:"registrant_change,omitempty"`

	// RegistrantChangePolicy describes the rules of a registrant change.
	RegistrantChangePolicy *TLDRegistrantChangePolicy `json:"registrant_change_policy,omitempty"`

	// SupportCheck says whether the registry supports contact checks.
	SupportCheck *bool `json:"support_check,omitempty"`

	// SupportClientContactID says whether the registrar chooses contact IDs.
	SupportClientContactID *bool `json:"support_client_contact_id,omitempty"`

	// SupportTransfer says whether contacts can be transferred.
	SupportTransfer *bool `json:"support_transfer,omitempty"`

	// SupportedPostalTypes are the postal address types the registry takes.
	SupportedPostalTypes []string `json:"supported_postal_types,omitempty"`

	// SupportedRoles are the contact roles a registration takes.
	SupportedRoles []TLDContactRole `json:"supported_roles,omitempty"`

	// TransferSupportedRoles are the contact roles a transfer takes.
	TransferSupportedRoles []TLDContactRole `json:"transfer_supported_roles,omitempty"`

	// UpdateSupportedRoles are the contact roles an update can change.
	UpdateSupportedRoles []TLDContactRole `json:"update_supported_roles,omitempty"`
}

// TLDContactRole is how many contacts of one role an operation takes
// (ContactConfigBase). A role with Min > 0 is required.
type TLDContactRole struct {
	// Type is the contact role.
	Type DomainContactType `json:"type"`

	// Min is the minimum number of contacts of this role.
	Min int `json:"min"`

	// Max is the maximum number of contacts of this role.
	Max int `json:"max"`
}

// TLDRegistrantChangePolicy describes the rules of a registrant change.
type TLDRegistrantChangePolicy struct {
	// Confirmation is who has to confirm the change.
	Confirmation string `json:"confirmation"`

	// ConfirmationWindow is how long the confirmation may take.
	ConfirmationWindow *string `json:"confirmation_window,omitempty"`

	// DesignatedAgentAllowed says whether a designated agent may confirm.
	DesignatedAgentAllowed bool `json:"designated_agent_allowed"`

	// Fee says whether the change costs a fee.
	Fee string `json:"fee"`

	// PriorRegistrantNotice says whether the prior registrant is notified.
	PriorRegistrantNotice bool `json:"prior_registrant_notice"`

	// RenewsDomain says whether the change renews the domain.
	RenewsDomain bool `json:"renews_domain"`

	// TransferLockAfterChange is how long transfers are locked afterwards.
	TransferLockAfterChange string `json:"transfer_lock_after_change"`
}

// TLDLaunchPhases describes the launch phases of a TLD.
type TLDLaunchPhases struct {
	// GeneralAvailability is when general availability started.
	GeneralAvailability TLDGeneralAvailability `json:"general_availability"`

	// Phases are the launch phases before general availability.
	Phases []TLDLaunchPhase `json:"phases"`

	// TrademarkClaims describes the trademark claims period, or nil.
	TrademarkClaims *TLDTrademarkClaims `json:"trademark_claims,omitempty"`
}

// TLDGeneralAvailability is when general availability of a TLD started.
type TLDGeneralAvailability struct {
	// StartDate is when general availability started.
	StartDate *time.Time `json:"start_date,omitempty"`
}

// TLDLaunchPhase is one launch phase of a TLD.
type TLDLaunchPhase struct {
	// Type is the phase (sunrise, landrush or eap).
	Type LaunchPhaseType `json:"type"`

	// Supported says whether OpusDNS supports the phase.
	Supported bool `json:"supported"`

	// Allocation is how names are allocated during the phase.
	Allocation *AllocationMethodType `json:"allocation,omitempty"`

	// SMDRequired says whether a signed mark is required.
	SMDRequired *bool `json:"smd_required,omitempty"`

	// StartDate is when the phase starts.
	StartDate *time.Time `json:"start_date,omitempty"`

	// EndDate is when the phase ends.
	EndDate *time.Time `json:"end_date,omitempty"`
}

// TLDTrademarkClaims describes the trademark claims period of a TLD.
type TLDTrademarkClaims struct {
	// Supported says whether OpusDNS supports the claims period.
	Supported bool `json:"supported"`

	// TMCHRequired says whether the Trademark Clearinghouse is involved.
	TMCHRequired *bool `json:"tmch_required,omitempty"`

	// StartDate is when the claims period starts.
	StartDate *time.Time `json:"start_date,omitempty"`

	// EndDate is when the claims period ends.
	EndDate *time.Time `json:"end_date,omitempty"`
}

// TLDLegalRequirement is a notice or confirmation an operation needs.
type TLDLegalRequirement struct {
	// Key identifies the requirement.
	Key string `json:"key"`

	// Type is "notice" or "confirmation".
	Type string `json:"type"`

	// Operations are the operations the requirement applies to
	// (registration, transfer, renewal, registrant_change).
	Operations []string `json:"operations"`

	// URL points to the legal text, if there is one.
	URL *string `json:"url,omitempty"`
}

// TLDLocalPresence describes local presence requirements of a TLD.
type TLDLocalPresence struct {
	// Required says whether a local presence is required.
	Required bool `json:"required"`

	// EligibleCountries are the countries that satisfy the requirement.
	EligibleCountries []string `json:"eligible_countries,omitempty"`

	// Exemptions are conditions under which the requirement does not apply.
	Exemptions []AttributeCondition `json:"exemptions,omitempty"`

	// Requirement lists what satisfies the requirement.
	Requirement []string `json:"requirement,omitempty"`

	// Type lists the contact roles the requirement applies to.
	Type []DomainContactType `json:"type,omitempty"`
}

// TLDPremiumDomains describes premium domain support of a TLD.
type TLDPremiumDomains struct {
	// Supported says whether the TLD has premium domains.
	Supported bool `json:"supported"`

	// Affects lists the operations premium pricing applies to.
	Affects []string `json:"affects,omitempty"`

	// Source is where premium prices come from.
	Source *string `json:"source,omitempty"`
}

// TLDRDAP describes the RDAP service of a TLD's registry.
type TLDRDAP struct {
	// RDAPServer is the RDAP base URL.
	RDAPServer *string `json:"rdap_server,omitempty"`
}

// TLDRegistryLock describes registry lock support of a TLD.
type TLDRegistryLock struct {
	// Supported says whether registry lock is available.
	Supported bool `json:"supported"`

	// Prevents lists the operations a registry lock blocks.
	Prevents []string `json:"prevents,omitempty"`

	// RemovalProcess describes how a lock is removed.
	RemovalProcess *string `json:"removal_process,omitempty"`

	// RequiresManualRequest says whether a lock needs a manual request.
	RequiresManualRequest *bool `json:"requires_manual_request,omitempty"`
}

// TLDReservedDomains describes the reserved names of a TLD.
type TLDReservedDomains struct {
	// Supported says whether the registry reserves names.
	Supported bool `json:"supported"`

	// Source is where the reserved list comes from.
	Source *string `json:"source,omitempty"`

	// URL points to the reserved list.
	URL *string `json:"url,omitempty"`
}

// TLDWhois describes the WHOIS service of a TLD's registry.
type TLDWhois struct {
	// WhoisServer is the WHOIS server host name.
	WhoisServer *string `json:"whois_server,omitempty"`
}

// TLDVerificationPolicy is a contact verification policy of a TLD. It covers
// both the email and the identity verification policy; fields that belong to
// the other kind stay empty.
type TLDVerificationPolicy struct {
	// Enabled says whether the policy applies.
	Enabled bool `json:"enabled"`

	// ContactRoles are the contact roles that must be verified.
	ContactRoles []DomainContactType `json:"contact_roles"`

	// Trigger lists the events that start a verification.
	Trigger []string `json:"trigger"`

	// SuspensionOnFailure says whether a failed verification suspends the
	// domain.
	SuspensionOnFailure bool `json:"suspension_on_failure"`

	// SuspensionDelay is how long after a failure the domain is suspended,
	// as an ISO 8601 duration.
	SuspensionDelay string `json:"suspension_delay"`

	// ValidityPeriod is how long a verification stays valid.
	ValidityPeriod *DomainPeriod `json:"validity_period,omitempty"`

	// VerificationMethod is how an email verification is carried out (email
	// policy only).
	VerificationMethod string `json:"verification_method,omitempty"`

	// Communication describes the reminders an email verification sends
	// (email policy only).
	Communication *TLDVerificationCommunication `json:"communication,omitempty"`

	// RequiredClaims are the claims an identity verification must prove
	// (identity policy only).
	RequiredClaims []TLDRequiredClaim `json:"required_claims,omitempty"`
}

// TLDRequiredClaim is a claim an identity verification must prove, and the
// proofs accepted for it (RequiredClaim).
type TLDRequiredClaim struct {
	// Claim is what must be proven (e.g. NAME, ADDRESS).
	Claim ContactVerificationClaim `json:"claim"`

	// AcceptedProofs are the proofs that count. Empty means any proof.
	AcceptedProofs []ContactVerificationProof `json:"accepted_proofs,omitempty"`
}

// TLDVerificationCommunication describes the reminders a verification sends.
type TLDVerificationCommunication struct {
	// Type is the kind of communication.
	Type string `json:"type"`

	// Timedelta lists when reminders are sent, as ISO 8601 durations.
	Timedelta []string `json:"timedelta"`
}
