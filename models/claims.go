package models

// Trademark claims notices, as returned for a claims key from a domain
// availability check. The structures follow RFC 9361 (the Trademark Claims
// Notice Information Specification); field names mirror the API, which in turn
// mirrors the RFC's element names.

// ClaimsNoticesRequest asks for the claims notices behind a claims key. The
// field is a list because the API models it as one, but the API accepts
// exactly one key per request.
type ClaimsNoticesRequest struct {
	// ClaimsKeys are the claims keys returned by an availability check.
	// Exactly one key is allowed.
	ClaimsKeys []string `json:"claims_keys"`
}

// ClaimsNoticesResponse carries one notice per requested claims key.
type ClaimsNoticesResponse struct {
	// ClaimsNotices are the notices, one per resolved claims key.
	ClaimsNotices []ClaimsNotice `json:"claims_notices"`
}

// ClaimsNotice is the trademark claims notice for a single domain label. It
// must be shown to the registrant, who acknowledges it by passing
// ClaimsNoticeAcceptanceHash when registering the domain.
type ClaimsNotice struct {
	// ClaimsKey is the claims key this notice was retrieved with.
	ClaimsKey string `json:"claims_key"`

	// Label is the domain name label the notice covers.
	Label string `json:"label"`

	// ClaimsNoticeAcceptanceHash acknowledges the notice at registration time.
	ClaimsNoticeAcceptanceHash string `json:"claims_notice_acceptance_hash"`

	// Claims are the trademark claims behind the notice.
	Claims []TmClaim `json:"claims,omitempty"`

	// RenderedHTML is a ready-to-display rendering of the whole notice.
	RenderedHTML string `json:"rendered_html,omitempty"`

	// NoticeTitle is the notice title.
	NoticeTitle string `json:"notice_title,omitempty"`

	// NoticeIntro is the introductory text.
	NoticeIntro string `json:"notice_intro,omitempty"`

	// NoticeNotExactMatchIntro introduces the non-exact-match section.
	NoticeNotExactMatchIntro string `json:"notice_not_exact_match_intro,omitempty"`

	// NoticeFooter is the footer text.
	NoticeFooter string `json:"notice_footer,omitempty"`

	// NoticeFooterURL is the footer URL.
	NoticeFooterURL string `json:"notice_footer_url,omitempty"`
}

// TmClaim is a single trademark claim within a notice.
type TmClaim struct {
	// MarkName is the mark text string.
	MarkName string `json:"mark_name"`

	// Holders are the holders of the mark, at least one.
	Holders []TmHolder `json:"holders"`

	// Contacts are the contacts or representatives for the mark.
	Contacts []TmContact `json:"contacts,omitempty"`

	// JurDesc is the jurisdiction the mark is protected in.
	JurDesc *TmJurDesc `json:"jur_desc,omitempty"`

	// ClassDescs are the Nice Classification descriptions.
	ClassDescs []TmClassDesc `json:"class_descs,omitempty"`

	// GoodsAndServices is the full description of goods and services.
	GoodsAndServices string `json:"goods_and_services,omitempty"`

	// NotExactMatch is set when the claim was added by a non-exact match rule.
	NotExactMatch *TmNotExactMatch `json:"not_exact_match,omitempty"`
}

// HolderEntitlement is how a holder is entitled to a mark.
type HolderEntitlement string

const (
	// HolderEntitlementOwner marks the owner of the trademark.
	HolderEntitlementOwner HolderEntitlement = "owner"

	// HolderEntitlementAssignee marks an assignee of the trademark.
	HolderEntitlementAssignee HolderEntitlement = "assignee"

	// HolderEntitlementLicensee marks a licensee of the trademark.
	HolderEntitlementLicensee HolderEntitlement = "licensee"
)

// TmHolder is a holder of a mark. Either Name or Org is set.
type TmHolder struct {
	// Entitlement is how the holder is entitled to the mark.
	Entitlement HolderEntitlement `json:"entitlement"`

	// Addr is the holder's address.
	Addr TmAddr `json:"addr"`

	// Name is the holder's personal name.
	Name *string `json:"name,omitempty"`

	// Org is the holder's organization.
	Org *string `json:"org,omitempty"`

	// Voice is the holder's phone number.
	Voice *string `json:"voice,omitempty"`

	// Fax is the holder's fax number.
	Fax *string `json:"fax,omitempty"`

	// Email is the holder's email address.
	Email *string `json:"email,omitempty"`
}

// TrademarkContactType is the role of a trademark contact.
type TrademarkContactType string

const (
	// TrademarkContactTypeOwner marks the mark owner.
	TrademarkContactTypeOwner TrademarkContactType = "owner"

	// TrademarkContactTypeAgent marks an agent of the owner.
	TrademarkContactTypeAgent TrademarkContactType = "agent"

	// TrademarkContactTypeThirdParty marks a third party.
	TrademarkContactTypeThirdParty TrademarkContactType = "third party"
)

// TmContact is a contact or representative of a mark.
type TmContact struct {
	// Type is the contact's role.
	Type TrademarkContactType `json:"type"`

	// Name is the contact's name.
	Name string `json:"name"`

	// Addr is the contact's address.
	Addr TmAddr `json:"addr"`

	// Voice is the contact's phone number.
	Voice string `json:"voice"`

	// Email is the contact's email address.
	Email string `json:"email"`

	// Org is the contact's organization.
	Org *string `json:"org,omitempty"`

	// Fax is the contact's fax number.
	Fax *string `json:"fax,omitempty"`
}

// TmAddr is an address on a trademark claim.
type TmAddr struct {
	// Street holds one to three street lines.
	Street []string `json:"street"`

	// City is the city.
	City string `json:"city"`

	// CC is the ISO 3166-2 two-character country code.
	CC string `json:"cc"`

	// SP is the state or province.
	SP *string `json:"sp,omitempty"`

	// PC is the postal code.
	PC *string `json:"pc,omitempty"`
}

// TmJurDesc is the jurisdiction a mark is protected in.
type TmJurDesc struct {
	// JurCC is the WIPO ST.3 two-character jurisdiction code.
	JurCC string `json:"jur_cc"`

	// Description is the name of the jurisdiction in English.
	Description string `json:"description"`
}

// TmClassDesc is a Nice Classification description.
type TmClassDesc struct {
	// ClassNum is the Nice Classification class number.
	ClassNum int `json:"class_num"`

	// Description describes the class in English.
	Description string `json:"description"`
}

// TmNotExactMatch records that a claim was added by a non-exact match rule,
// citing the decisions behind it.
type TmNotExactMatch struct {
	// Intro is the introductory text for the section.
	Intro string `json:"intro,omitempty"`

	// UDRP are the UDRP cases behind the claim.
	UDRP []TmUdrp `json:"udrp,omitempty"`

	// Court are the court resolutions behind the claim.
	Court []TmCourt `json:"court,omitempty"`
}

// TmUdrp is a UDRP case reference.
type TmUdrp struct {
	// CaseNo is the UDRP case number.
	CaseNo string `json:"case_no"`

	// UDRPProvider is the name of the UDRP provider.
	UDRPProvider string `json:"udrp_provider"`
}

// TmCourt is a court resolution reference.
type TmCourt struct {
	// RefNum is the reference number of the court resolution.
	RefNum string `json:"ref_num"`

	// CC is the ISO 3166-2 jurisdiction country code.
	CC string `json:"cc"`

	// CourtName is the name of the court.
	CourtName string `json:"court_name"`

	// Region names the regions within the jurisdiction.
	Region []string `json:"region,omitempty"`
}
