package models

import "time"

// Whitelabel branding serves the OpusDNS dashboard and its transactional email
// under a reseller's own brand. An organization has at most one configuration,
// on one of two tiers: base serves it on a subdomain of an OpusDNS-owned zone,
// plus serves it on the reseller's own domain. A configuration moves between
// the tiers in place, so the tier is a mutable attribute rather than a choice
// between two kinds of record.

// WhitelabelTier is the tier a whitelabel configuration is on.
type WhitelabelTier string

const (
	// WhitelabelTierBase serves the customer on a subdomain of an
	// OpusDNS-owned zone.
	WhitelabelTierBase WhitelabelTier = "base"

	// WhitelabelTierPlus serves the customer on their own domain.
	WhitelabelTierPlus WhitelabelTier = "plus"
)

// WhitelabelOnboardingStatus is how far onboarding has progressed.
type WhitelabelOnboardingStatus string

const (
	// WhitelabelOnboardingPendingDomainVerification waits for the customer to
	// delegate their domain.
	WhitelabelOnboardingPendingDomainVerification WhitelabelOnboardingStatus = "pending_domain_verification"

	// WhitelabelOnboardingVerifying is checking that delegation.
	WhitelabelOnboardingVerifying WhitelabelOnboardingStatus = "verifying"

	// WhitelabelOnboardingProvisioning is setting up DNS, auth and branding.
	WhitelabelOnboardingProvisioning WhitelabelOnboardingStatus = "provisioning"

	// WhitelabelOnboardingActive is serving.
	WhitelabelOnboardingActive WhitelabelOnboardingStatus = "active"

	// WhitelabelOnboardingFailed stopped on a terminal error. See the failure
	// code and type for why.
	WhitelabelOnboardingFailed WhitelabelOnboardingStatus = "failed"

	// WhitelabelOnboardingTerminated has been torn down.
	WhitelabelOnboardingTerminated WhitelabelOnboardingStatus = "terminated"
)

// WhitelabelFailureType is which subsystem a terminal onboarding failure came
// from.
type WhitelabelFailureType string

const (
	WhitelabelFailureTypeDNS      WhitelabelFailureType = "dns"
	WhitelabelFailureTypeAuth     WhitelabelFailureType = "auth"
	WhitelabelFailureTypeBranding WhitelabelFailureType = "branding"
	WhitelabelFailureTypeSystem   WhitelabelFailureType = "system"
)

// WhitelabelFailureCode is the stable reason a terminal onboarding failure
// happened. The free-text detail beside it is for support, not for display to
// the customer.
type WhitelabelFailureCode string

const (
	WhitelabelFailureZoneNotOwned       WhitelabelFailureCode = "zone_not_owned"
	WhitelabelFailureDelegationMissing  WhitelabelFailureCode = "delegation_missing"
	WhitelabelFailureDelegationMismatch WhitelabelFailureCode = "delegation_mismatch"
	WhitelabelFailureZoneCreateFailed   WhitelabelFailureCode = "zone_create_failed"
	WhitelabelFailureDNSRecordRejected  WhitelabelFailureCode = "dns_record_rejected"
	WhitelabelFailureAuthClientRejected WhitelabelFailureCode = "auth_client_rejected"
	WhitelabelFailureBrandingRejected   WhitelabelFailureCode = "branding_rejected"
	WhitelabelFailureRetriesExhausted   WhitelabelFailureCode = "retries_exhausted"
)

// WhitelabelRenewalMode is whether the whitelabel subscription renews or lapses
// at the end of its period. The wire values match those of domains and vanity
// nameserver sets.
type WhitelabelRenewalMode string

const (
	// WhitelabelRenewalModeRenew auto-renews the subscription.
	WhitelabelRenewalModeRenew WhitelabelRenewalMode = "renew"

	// WhitelabelRenewalModeExpire lets it lapse at period end.
	WhitelabelRenewalModeExpire WhitelabelRenewalMode = "expire"
)

// Whitelabel is an organization's whitelabel branding configuration.
type Whitelabel struct {
	// WhitelabelBrandingID is the identifier of the configuration.
	WhitelabelBrandingID TypeID `json:"whitelabel_branding_id"`

	// OrganizationID is the organization the configuration belongs to.
	OrganizationID OrganizationID `json:"organization_id"`

	// Tier is the tier the configuration is on.
	Tier WhitelabelTier `json:"tier"`

	// OnboardingStatus is how far onboarding has progressed.
	OnboardingStatus WhitelabelOnboardingStatus `json:"onboarding_status"`

	// Enabled says whether the configuration is currently serving.
	Enabled bool `json:"enabled,omitempty"`

	// Hostname is the host the dashboard is served on.
	Hostname string `json:"hostname"`

	// AuthHostname is the host authentication is served on.
	AuthHostname string `json:"auth_hostname"`

	// BaseLabel is the label identifying the configuration.
	BaseLabel string `json:"base_label"`

	// VerificationDomain is the domain the customer must delegate.
	VerificationDomain string `json:"verification_domain"`

	// KeycloakClientID is the auth client provisioned for the configuration.
	KeycloakClientID string `json:"keycloak_client_id"`

	// Subscription carries the billing period and renewal intent.
	Subscription *WhitelabelSubscription `json:"subscription,omitempty"`

	// FailureType is which subsystem a terminal failure came from.
	FailureType *WhitelabelFailureType `json:"failure_type,omitempty"`

	// FailureCode is the stable reason for a terminal failure.
	FailureCode *WhitelabelFailureCode `json:"failure_code,omitempty"`

	// FailureDetail is free-text failure context, for support rather than for
	// display to the customer.
	FailureDetail *string `json:"failure_detail,omitempty"`

	// CreatedOn is when the configuration was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the configuration was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// WhitelabelSubscription is the billing state of a whitelabel configuration.
type WhitelabelSubscription struct {
	// RenewalMode says whether the subscription renews or lapses.
	RenewalMode WhitelabelRenewalMode `json:"renewal_mode"`

	// Period is the billing period it renews on.
	Period DomainPeriod `json:"period"`

	// ExpiresOn is the end of the current paid term.
	ExpiresOn *time.Time `json:"expires_on,omitempty"`

	// RenewScheduledAt is when the next renewal runs, on a renewing config.
	RenewScheduledAt *time.Time `json:"renew_scheduled_at,omitempty"`

	// GracePeriodEndsAt is the end of the grace period, when in one.
	GracePeriodEndsAt *time.Time `json:"grace_period_ends_at,omitempty"`
}

// WhitelabelBaseCreateRequest provisions a base-tier whitelabel, served on a
// subdomain of an OpusDNS-owned zone.
type WhitelabelBaseCreateRequest struct {
	// Tier must be WhitelabelTierBase.
	Tier WhitelabelTier `json:"tier"`

	// Label is the label to serve under.
	Label string `json:"label"`

	// Period is the billing period.
	Period DomainPeriod `json:"period"`
}

// WhitelabelPlusCreateRequest provisions a plus-tier whitelabel, served on the
// customer's own domain.
type WhitelabelPlusCreateRequest struct {
	// Tier must be WhitelabelTierPlus.
	Tier WhitelabelTier `json:"tier"`

	// Label is the label to serve under.
	Label string `json:"label"`

	// Period is the billing period.
	Period DomainPeriod `json:"period"`

	// Hostname is the customer's own domain.
	Hostname string `json:"hostname"`

	// AuthSubdomain is the subdomain to serve authentication on.
	AuthSubdomain string `json:"auth_subdomain"`

	// DashboardSubdomain is the subdomain to serve the dashboard on.
	DashboardSubdomain *string `json:"dashboard_subdomain,omitempty"`

	// CreateZone asks OpusDNS to create the customer's zone during onboarding.
	CreateZone *bool `json:"create_zone,omitempty"`
}

// WhitelabelUpdateRequest changes the mutable attributes of a configuration.
type WhitelabelUpdateRequest struct {
	// Label is a new label to serve under.
	Label *string `json:"label,omitempty"`

	// Enabled turns serving on or off.
	Enabled *bool `json:"enabled,omitempty"`

	// RenewalMode changes the renewal intent of the subscription.
	RenewalMode *WhitelabelRenewalMode `json:"renewal_mode,omitempty"`
}

// WhitelabelUpgradeRequest moves a base-tier configuration to the plus tier,
// onto the customer's own domain.
type WhitelabelUpgradeRequest struct {
	// Hostname is the customer's own domain.
	Hostname string `json:"hostname"`

	// AuthSubdomain is the subdomain to serve authentication on.
	AuthSubdomain string `json:"auth_subdomain"`

	// DashboardSubdomain is the subdomain to serve the dashboard on.
	DashboardSubdomain *string `json:"dashboard_subdomain,omitempty"`

	// CreateZone asks OpusDNS to create the customer's zone.
	CreateZone *bool `json:"create_zone,omitempty"`
}

// WhitelabelRecheckRequest re-runs verification for a configuration whose
// onboarding failed, optionally correcting the hostname it was pointed at.
type WhitelabelRecheckRequest struct {
	// Hostname replaces the domain to verify.
	Hostname *string `json:"hostname,omitempty"`

	// AuthSubdomain replaces the auth subdomain.
	AuthSubdomain *string `json:"auth_subdomain,omitempty"`

	// DashboardSubdomain replaces the dashboard subdomain.
	DashboardSubdomain *string `json:"dashboard_subdomain,omitempty"`

	// CreateZone asks OpusDNS to create the customer's zone on this attempt.
	CreateZone *bool `json:"create_zone,omitempty"`
}

// ProductCreateResponse identifies the subscription a product purchase created.
type ProductCreateResponse struct {
	// SubscriptionID is the subscription that was created.
	SubscriptionID string `json:"subscription_id"`

	// SubscribableID is the resource the subscription bills for.
	SubscribableID string `json:"subscribable_id"`

	// ProductName is the product that was subscribed to.
	ProductName string `json:"product_name"`
}

// BrandingDocument is the reseller's visual and textual branding, applied to
// the dashboard and to transactional email.
type BrandingDocument struct {
	// Version is the branding document revision.
	Version *string `json:"version,omitempty"`

	// Brand carries the reseller's name and logos.
	Brand *Brand `json:"brand,omitempty"`

	// Theme carries the colour palettes and typography.
	Theme *Theme `json:"theme,omitempty"`

	// Content carries the email copy, footer and legal links.
	Content *BrandingContent `json:"content,omitempty"`

	// Auth carries the authentication client settings.
	Auth *BrandingAuth `json:"auth,omitempty"`
}

// Brand is the reseller's name and logos.
type Brand struct {
	// Name is the brand name.
	Name *string `json:"name,omitempty"`

	// Logo holds the logo variants.
	Logo *Logo `json:"logo,omitempty"`
}

// Logo holds the logo variants, each a URL.
type Logo struct {
	// Light is the logo for light backgrounds.
	Light *string `json:"light,omitempty"`

	// Dark is the logo for dark backgrounds.
	Dark *string `json:"dark,omitempty"`

	// Icon is the square icon.
	Icon *string `json:"icon,omitempty"`

	// Favicon is the browser tab icon.
	Favicon *string `json:"favicon,omitempty"`
}

// Theme is the typography and the light and dark colour palettes.
type Theme struct {
	// Light is the light-mode palette.
	Light *Palette `json:"light,omitempty"`

	// Dark is the dark-mode palette.
	Dark *Palette `json:"dark,omitempty"`

	// FontFamily is the CSS font stack.
	FontFamily *string `json:"font_family,omitempty"`

	// FontURL is a webfont stylesheet to load.
	FontURL *string `json:"font_url,omitempty"`

	// Radius is the CSS corner radius.
	Radius *string `json:"radius,omitempty"`
}

// Palette is one colour scheme. Every colour is a CSS hex string such as
// "#0a7cff" or "#0af".
type Palette struct {
	Background          *string `json:"background,omitempty"`
	Foreground          *string `json:"foreground,omitempty"`
	Card                *string `json:"card,omitempty"`
	CardForeground      *string `json:"card_foreground,omitempty"`
	Primary             *string `json:"primary,omitempty"`
	PrimaryForeground   *string `json:"primary_foreground,omitempty"`
	Secondary           *string `json:"secondary,omitempty"`
	SecondaryForeground *string `json:"secondary_foreground,omitempty"`
	Accent              *string `json:"accent,omitempty"`
	AccentForeground    *string `json:"accent_foreground,omitempty"`
	Muted               *string `json:"muted,omitempty"`
	MutedForeground     *string `json:"muted_foreground,omitempty"`
	Sidebar             *string `json:"sidebar,omitempty"`
	SidebarForeground   *string `json:"sidebar_foreground,omitempty"`
	Border              *string `json:"border,omitempty"`
	Input               *string `json:"input,omitempty"`
	Success             *string `json:"success,omitempty"`
	Warning             *string `json:"warning,omitempty"`
	Danger              *string `json:"danger,omitempty"`
	Info                *string `json:"info,omitempty"`

	// ChartColors is the categorical palette for charts.
	ChartColors []string `json:"chart_colors,omitempty"`

	// TagColors is the palette tags are drawn from.
	TagColors []string `json:"tag_colors,omitempty"`
}

// BrandingContent is the copy shown around the product.
type BrandingContent struct {
	// EmailBlocks overrides template copy, keyed by template, then block, then
	// locale.
	EmailBlocks map[string]map[string]map[string]string `json:"email_blocks,omitempty"`

	// Signature is the email signature.
	Signature *string `json:"signature,omitempty"`

	// Footer is the email footer.
	Footer *BrandingFooter `json:"footer,omitempty"`

	// Legal carries the terms and privacy links.
	Legal *BrandingLegal `json:"legal,omitempty"`

	// Support carries the support contact details.
	Support *BrandingSupport `json:"support,omitempty"`
}

// BrandingFooter is the legal footer of transactional email.
type BrandingFooter struct {
	// LegalEntity is the entity named in the footer.
	LegalEntity *string `json:"legal_entity,omitempty"`

	// Lines are the free-text footer lines.
	Lines []string `json:"lines,omitempty"`

	// VATID is the VAT identifier shown in the footer.
	VATID *string `json:"vat_id,omitempty"`
}

// BrandingLegal carries the legal document links.
type BrandingLegal struct {
	// TermsURL links to the terms of service.
	TermsURL *string `json:"terms_url,omitempty"`

	// PrivacyURL links to the privacy policy.
	PrivacyURL *string `json:"privacy_url,omitempty"`
}

// BrandingSupport carries the support contact details.
type BrandingSupport struct {
	// Email is the support email address.
	Email *string `json:"email,omitempty"`

	// URL is the support site.
	URL *string `json:"url,omitempty"`
}

// BrandingAuth carries the authentication client settings.
type BrandingAuth struct {
	// Authority is the OIDC issuer.
	Authority *string `json:"authority,omitempty"`

	// ClientID is the OIDC client.
	ClientID *string `json:"client_id,omitempty"`
}

// MailTemplate describes one editable transactional email template.
type MailTemplate struct {
	// Label is the human-readable template name.
	Label string `json:"label,omitempty"`

	// Version is the template revision, as major.minor.
	Version string `json:"version,omitempty"`

	// Locales are the supported locales; the first is the fallback.
	Locales []string `json:"locales,omitempty"`

	// Subject is the subject line per locale.
	Subject map[string]string `json:"subject,omitempty"`

	// Blocks are the content blocks an organization may override.
	Blocks map[string]MailTemplateBlock `json:"blocks,omitempty"`

	// Variables are the values substituted at send time.
	Variables map[string]MailTemplateVariable `json:"variables,omitempty"`
}

// MailTemplateBlock is one overridable block of a mail template.
type MailTemplateBlock struct {
	// Label is the human-readable field name for the editor.
	Label string `json:"label,omitempty"`

	// Default is the default copy per locale.
	Default map[string]string `json:"default,omitempty"`

	// MaxLen is the maximum length of an override.
	MaxLen int `json:"max_len,omitempty"`

	// Multiline renders the editor field as a text area.
	Multiline bool `json:"multiline,omitempty"`
}

// MailTemplateVariable is one value substituted into a template at send time.
type MailTemplateVariable struct {
	// Type is the value type, which drives render-time escaping.
	Type string `json:"type,omitempty"`

	// Sample is the value shown in previews.
	Sample string `json:"sample,omitempty"`

	// Required says whether the variable must be supplied.
	Required bool `json:"required,omitempty"`
}

// PreviewMailRequest renders one template with a branding document, to preview
// changes before saving them.
type PreviewMailRequest struct {
	// TemplateName is the template to render.
	TemplateName string `json:"template_name"`

	// LanguageCode is the locale to render in.
	LanguageCode string `json:"language_code"`

	// BrandingDocument is the branding to render with. When omitted the saved
	// branding is used.
	BrandingDocument *BrandingDocument `json:"branding_document,omitempty"`
}

// PreviewMailResponse is a rendered mail preview.
type PreviewMailResponse struct {
	// Subject is the rendered subject line.
	Subject string `json:"subject"`

	// HTML is the rendered HTML body.
	HTML string `json:"html"`

	// Text is the rendered plain-text body.
	Text string `json:"text"`
}
