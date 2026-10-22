// Package models contains all the data types for the OpusDNS API.
package models

// NOTE: The TLD struct in this file is a flattened convenience view that
// ListTLDs builds from the TLD list. It fills Name, Type, Available,
// IDNSupported, DNSSECSupported, Min/MaxRegistrationPeriod and Specification;
// the other fields have no source in the API. The full specification of a TLD
// is TLDSpecification (see tld_specification.go), which GetTLD returns and
// TLD.Specification carries.

import "encoding/json"

// TLDInfo represents the basic TLD name and type information.
type TLDInfo struct {
	// Name is the TLD name without the leading dot (e.g., "com", "de", "io").
	Name string `json:"name"`

	// Type is the TLD type (e.g., "gTLD", "ccTLD").
	Type TLDType `json:"type,omitempty"`

	// ThirdLevelStructure contains third level domain configurations (names).
	ThirdLevelStructure []string `json:"third_level_structure,omitempty"`
}

// TLD represents a top-level domain with its configuration.
type TLD struct {
	// Name is the TLD name without the leading dot (e.g., "com", "de", "io").
	// This is populated from the nested tlds array for convenience.
	Name string `json:"name"`

	// Type is the TLD type (e.g., "gTLD", "ccTLD").
	// This is populated from the nested tlds array for convenience.
	Type TLDType `json:"type,omitempty"`

	// Available indicates whether the TLD is available for registration.
	Available bool `json:"available"`

	// RegistrationEnabled indicates if new registrations are accepted.
	//
	// Deprecated: The API does not send this, so it is always false. Available
	// tells whether the TLD is enabled for the organization.
	RegistrationEnabled bool `json:"registration_enabled,omitempty"`

	// TransferEnabled indicates if transfers are accepted.
	//
	// Deprecated: The API does not send this, so it is always false.
	TransferEnabled bool `json:"transfer_enabled,omitempty"`

	// IDNSupported indicates if internationalized domain names are supported.
	// ListTLDs reads it from Specification.IDN.IDNCapable.
	IDNSupported bool `json:"idn_supported,omitempty"`

	// DNSSECSupported indicates if DNSSEC is supported. ListTLDs reads it from
	// Specification.DNSConfiguration.DNSSECAllowed.
	DNSSECSupported bool `json:"dnssec_supported,omitempty"`

	// MinRegistrationPeriod is the minimum registration period in years.
	// ListTLDs derives it from Specification.DomainLifecycle.RegistrationPeriods;
	// periods that are not whole years are left out, and it is 0 when none is.
	MinRegistrationPeriod int `json:"min_registration_period,omitempty"`

	// MaxRegistrationPeriod is the maximum registration period in years,
	// derived like MinRegistrationPeriod.
	MaxRegistrationPeriod int `json:"max_registration_period,omitempty"`

	// GracePeriodDays is the number of days in the grace period after expiration.
	GracePeriodDays int `json:"grace_period_days,omitempty"`

	// RedemptionPeriodDays is the number of days in the redemption period.
	RedemptionPeriodDays int `json:"redemption_period_days,omitempty"`

	// Pricing contains pricing information for the TLD.
	Pricing *TLDPricing `json:"pricing,omitempty"`

	// Restrictions contains any registration restrictions.
	Restrictions *TLDRestrictions `json:"restrictions,omitempty"`

	// ContactConfig contains contact requirements for the TLD.
	ContactConfig []ContactConfig `json:"contact_config,omitempty"`

	// NameserverConfig contains nameserver requirements.
	NameserverConfig *NameserverConfig `json:"nameserver_config,omitempty"`

	// AttributeDefinitions contains TLD-specific attribute definitions.
	AttributeDefinitions []ContactAttributeDefinition `json:"attribute_definitions,omitempty"`

	// RoleAttributeRequirements contains attribute requirements by contact role.
	RoleAttributeRequirements []ContactRoleAttributeRequirement `json:"role_attribute_requirements,omitempty"`

	// Specification is the full specification of the TLD as the list endpoint
	// sends it. TLDs that share a specification share the pointer.
	Specification *TLDSpecification `json:"-"`
}

// TLDType represents the type of TLD.
type TLDType string

const (
	// TLDTypeGTLD is a generic top-level domain (e.g., .com, .net, .org).
	TLDTypeGTLD TLDType = "gTLD"

	// TLDTypeCCTLD is a country-code top-level domain (e.g., .de, .uk, .fr).
	TLDTypeCCTLD TLDType = "ccTLD"
)

// TLDPricing contains pricing information for a TLD.
type TLDPricing struct {
	// RegisterPrice is the registration price.
	RegisterPrice string `json:"register_price,omitempty"`

	// RenewPrice is the renewal price.
	RenewPrice string `json:"renew_price,omitempty"`

	// TransferPrice is the transfer price.
	TransferPrice string `json:"transfer_price,omitempty"`

	// RestorePrice is the restore price (from redemption).
	RestorePrice string `json:"restore_price,omitempty"`

	// Currency is the currency code.
	Currency Currency `json:"currency,omitempty"`

	// PremiumPricing indicates if this TLD has premium pricing tiers.
	PremiumPricing bool `json:"premium_pricing,omitempty"`
}

// TLDRestrictions contains registration restrictions for a TLD.
type TLDRestrictions struct {
	// LocalPresenceRequired indicates if a local address is required.
	LocalPresenceRequired bool `json:"local_presence_required,omitempty"`

	// RestrictedCountries is a list of countries that cannot register.
	RestrictedCountries []string `json:"restricted_countries,omitempty"`

	// AllowedCountries is a list of countries that can register (if restricted).
	AllowedCountries []string `json:"allowed_countries,omitempty"`

	// RequiresVerification indicates if additional verification is required.
	RequiresVerification bool `json:"requires_verification,omitempty"`

	// TrademarkRequired indicates if a trademark is required.
	TrademarkRequired bool `json:"trademark_required,omitempty"`

	// RegistrantTypes is a list of allowed registrant types.
	RegistrantTypes []string `json:"registrant_types,omitempty"`

	// Notes provides additional information about restrictions.
	Notes *string `json:"notes,omitempty"`
}

// ContactConfig contains contact requirements for a TLD.
type ContactConfig struct {
	// Type is the contact type (registrant, admin, tech, billing).
	Type DomainContactType `json:"type"`

	// Min is the minimum number of contacts of this type.
	Min int `json:"min"`

	// Max is the maximum number of contacts of this type.
	Max int `json:"max"`

	// Required indicates if this contact type is required.
	Required bool `json:"required,omitempty"`
}

// NameserverConfig contains nameserver requirements for a TLD.
type NameserverConfig struct {
	// Min is the minimum number of nameservers required.
	Min int `json:"min"`

	// Max is the maximum number of nameservers allowed.
	Max int `json:"max"`

	// GlueRecordsRequired indicates if glue records are required.
	GlueRecordsRequired bool `json:"glue_records_required,omitempty"`
}

// AllocationMethodType represents how domain registrations are allocated.
type AllocationMethodType string

const (
	// AllocationMethodFCFS is first-come-first-served allocation.
	AllocationMethodFCFS AllocationMethodType = "fcfs"

	// AllocationMethodAuction allocates via auction.
	AllocationMethodAuction AllocationMethodType = "auction"

	// AllocationMethodLottery allocates via lottery.
	AllocationMethodLottery AllocationMethodType = "lottery"
)

// TLDConfiguration is one entry of the TLD list: a TLD specification plus
// whether it is enabled for the organization.
type TLDConfiguration struct {
	Enabled        bool      `json:"enabled"`
	ParkingEnabled bool      `json:"parking_enabled,omitempty"`
	TLDs           []TLDInfo `json:"tlds"`

	// Specification is the full TLD specification of this entry. The list
	// endpoint sends the same fields as GetTLD.
	Specification TLDSpecification `json:"-"`
}

// UnmarshalJSON decodes the entry and its full specification from the same
// object.
func (c *TLDConfiguration) UnmarshalJSON(data []byte) error {
	type plain TLDConfiguration
	if err := json.Unmarshal(data, (*plain)(c)); err != nil {
		return err
	}
	return json.Unmarshal(data, &c.Specification)
}

// TLDListResponse represents the response when listing TLDs.
type TLDListResponse struct {
	// TLDConfigurations contains the list of TLD configurations.
	TLDConfigurations []TLDConfiguration `json:"tlds"`
}

// TLDPortfolioEntry is one TLD of an organization's portfolio
// (TldResponseShort).
type TLDPortfolioEntry struct {
	// TLD is the TLD name without the leading dot (e.g. "com", "de").
	TLD string `json:"tld"`

	// Type is the TLD type (gTLD or ccTLD).
	Type TLDType `json:"type"`
}

// GetTLDOptions overrides which TLD specification variant the API resolves.
// All fields are optional; the API falls back to the organization's pinned
// specification.
type GetTLDOptions struct {
	// Backend selects the registry backend variant of the specification.
	Backend string

	// CustomerSpecRef selects a customer-specific specification.
	CustomerSpecRef string

	// Version selects a specific specification version.
	Version string
}

// ListTLDsOptions contains options for listing TLDs. The endpoint is not
// paginated and filters only by TLD name.
type ListTLDsOptions struct {
	// TLDs limits the list to these TLDs (names without the leading dot).
	TLDs []string

	// Page is the page number to retrieve (1-indexed).
	//
	// Deprecated: The endpoint is not paginated. The API ignored this field,
	// so the client no longer sends it.
	Page int

	// PageSize is the number of TLDs per page.
	//
	// Deprecated: The endpoint is not paginated. The API ignored this field,
	// so the client no longer sends it.
	PageSize int

	// Search is an optional search query to filter TLDs by name.
	//
	// Deprecated: The API ignored this filter, so the client no longer sends
	// it. Use TLDs to select TLDs by name.
	Search string

	// Type filters by TLD type.
	//
	// Deprecated: The API ignored this filter, so the client no longer sends
	// it. Filter the result on TLD.Type instead.
	Type TLDType

	// Available filters by availability status.
	//
	// Deprecated: The API ignored this filter, so the client no longer sends
	// it. ListTLDs already leaves out TLDs that are not enabled.
	Available *bool

	// RegistrationEnabled filters by registration enabled status.
	//
	// Deprecated: The API ignored this filter, so the client no longer sends
	// it.
	RegistrationEnabled *bool

	// DNSSECSupported filters by DNSSEC support.
	//
	// Deprecated: The API ignored this filter, so the client no longer sends
	// it. Read TLDSpecification.DNSConfiguration from GetTLD instead.
	DNSSECSupported *bool
}

// LaunchPhaseType represents the type of a TLD launch phase.
type LaunchPhaseType string

const (
	// LaunchPhaseSunrise is the sunrise launch phase.
	LaunchPhaseSunrise LaunchPhaseType = "sunrise"

	// LaunchPhaseLandrush is the landrush launch phase.
	LaunchPhaseLandrush LaunchPhaseType = "landrush"

	// LaunchPhaseEAP is the Early Access Program launch phase.
	LaunchPhaseEAP LaunchPhaseType = "eap"
)
