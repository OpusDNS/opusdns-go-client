// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// TLDInfo represents the basic TLD name and type information.
type TLDInfo struct {
	// Name is the TLD name without the leading dot (e.g., "com", "de", "io").
	Name string `json:"name"`

	// Type is the TLD type (e.g., "gTLD", "ccTLD", "newGTLD").
	Type TLDType `json:"type,omitempty"`

	// ThirdLevelStructure contains third level domain configurations.
	ThirdLevelStructure []interface{} `json:"third_level_structure,omitempty"`
}

// TLD represents a top-level domain with its configuration.
type TLD struct {
	// Name is the TLD name without the leading dot (e.g., "com", "de", "io").
	// This is populated from the nested tlds array for convenience.
	Name string `json:"name"`

	// Type is the TLD type (e.g., "gTLD", "ccTLD", "newGTLD").
	// This is populated from the nested tlds array for convenience.
	Type TLDType `json:"type,omitempty"`

	// Available indicates whether the TLD is available for registration.
	Available bool `json:"available"`

	// RegistrationEnabled indicates if new registrations are accepted.
	RegistrationEnabled bool `json:"registration_enabled,omitempty"`

	// TransferEnabled indicates if transfers are accepted.
	TransferEnabled bool `json:"transfer_enabled,omitempty"`

	// IDNSupported indicates if internationalized domain names are supported.
	IDNSupported bool `json:"idn_supported,omitempty"`

	// DNSSECSupported indicates if DNSSEC is supported.
	DNSSECSupported bool `json:"dnssec_supported,omitempty"`

	// MinRegistrationPeriod is the minimum registration period in years.
	MinRegistrationPeriod int `json:"min_registration_period,omitempty"`

	// MaxRegistrationPeriod is the maximum registration period in years.
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
}

// TLDType represents the type of TLD.
type TLDType string

const (
	// TLDTypeGTLD is a generic top-level domain (e.g., .com, .net, .org).
	TLDTypeGTLD TLDType = "gTLD"

	// TLDTypeCCTLD is a country-code top-level domain (e.g., .de, .uk, .fr).
	TLDTypeCCTLD TLDType = "ccTLD"

	// TLDTypeNewGTLD is a new generic top-level domain (e.g., .app, .dev, .xyz).
	TLDTypeNewGTLD TLDType = "newGTLD"

	// TLDTypeSponsoredGTLD is a sponsored generic TLD (e.g., .gov, .edu).
	TLDTypeSponsoredGTLD TLDType = "sponsoredGTLD"
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

// TLDConfiguration represents the full TLD configuration from the API.
type TLDConfiguration struct {
	Enabled        bool      `json:"enabled"`
	ParkingEnabled bool      `json:"parking_enabled,omitempty"`
	TLDs           []TLDInfo `json:"tlds"`
}

// TLDListResponse represents the response when listing TLDs.
type TLDListResponse struct {
	// TLDConfigurations contains the list of TLD configurations.
	TLDConfigurations []TLDConfiguration `json:"tlds"`
}

// TLDPortfolio represents a collection of TLDs available to an organization.
type TLDPortfolio struct {
	// TLDs contains the list of available TLDs.
	TLDs []TLD `json:"tlds"`

	// Total is the total number of TLDs in the portfolio.
	Total int `json:"total"`

	// UpdatedOn is when the portfolio was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// ListTLDsOptions contains options for listing TLDs.
type ListTLDsOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of TLDs per page.
	PageSize int

	// Search is an optional search query to filter TLDs by name.
	Search string

	// Type filters by TLD type.
	Type TLDType

	// Available filters by availability status.
	Available *bool

	// RegistrationEnabled filters by registration enabled status.
	RegistrationEnabled *bool

	// DNSSECSupported filters by DNSSEC support.
	DNSSECSupported *bool
}

// TLDDetails represents detailed information about a TLD.
type TLDDetails struct {
	TLD

	// Description provides a description of the TLD.
	Description *string `json:"description,omitempty"`

	// Registry is the name of the registry operator.
	Registry *string `json:"registry,omitempty"`

	// LaunchDate is when the TLD was launched.
	LaunchDate *time.Time `json:"launch_date,omitempty"`

	// WhoisServer is the WHOIS server for the TLD.
	WhoisServer *string `json:"whois_server,omitempty"`

	// RDAPServer is the RDAP server for the TLD.
	RDAPServer *string `json:"rdap_server,omitempty"`

	// SupportedIDNScripts is a list of supported IDN scripts.
	SupportedIDNScripts []string `json:"supported_idn_scripts,omitempty"`

	// ProhibitedCharacters is a list of characters not allowed in domain names.
	ProhibitedCharacters []string `json:"prohibited_characters,omitempty"`

	// MinDomainLength is the minimum domain name length (SLD).
	MinDomainLength int `json:"min_domain_length,omitempty"`

	// MaxDomainLength is the maximum domain name length (SLD).
	MaxDomainLength int `json:"max_domain_length,omitempty"`

	// Phases contains information about launch phases (if applicable).
	Phases []TLDPhase `json:"phases,omitempty"`
}

// TLDPhase represents a launch phase for a TLD.
type TLDPhase struct {
	// Name is the phase name (e.g., "sunrise", "landrush", "general_availability").
	Name string `json:"name"`

	// Status is the phase status (e.g., "active", "completed", "upcoming").
	Status string `json:"status"`

	// StartDate is when the phase starts.
	StartDate *time.Time `json:"start_date,omitempty"`

	// EndDate is when the phase ends.
	EndDate *time.Time `json:"end_date,omitempty"`

	// AllocationMethod is how domains are allocated during this phase.
	AllocationMethod AllocationMethodType `json:"allocation_method,omitempty"`

	// Requirements describes any special requirements for this phase.
	Requirements *string `json:"requirements,omitempty"`
}
