package models

import "time"

// Types for the TLD-specific domain operations: registry actions that only
// exist for certain TLDs and have no equivalent on the generic domain
// endpoints.

// AuthCodeTLD is a TLD whose registry issues the transfer auth code on request
// rather than returning it with the domain.
type AuthCodeTLD string

const (
	AuthCodeTLDBE    AuthCodeTLD = "be"
	AuthCodeTLDCymru AuthCodeTLD = "cymru"
	AuthCodeTLDCZ    AuthCodeTLD = "cz"
	AuthCodeTLDDK    AuthCodeTLD = "dk"
	AuthCodeTLDEU    AuthCodeTLD = "eu"
	AuthCodeTLDLT    AuthCodeTLD = "lt"
	AuthCodeTLDNU    AuthCodeTLD = "nu"
	AuthCodeTLDSE    AuthCodeTLD = "se"
	AuthCodeTLDWales AuthCodeTLD = "wales"
)

// RequestAuthCodeResponse is the outcome of an auth code request. The registry
// delivers the code out of band, usually by email to the registrant.
type RequestAuthCodeResponse struct {
	// Name is the domain the auth code was requested for.
	Name string `json:"name"`

	// Success says whether the registry accepted the request.
	Success bool `json:"success"`

	// Detail explains the outcome when Success is false.
	Detail *string `json:"detail,omitempty"`
}

// DomainWithdrawResponse is the outcome of withdrawing an .at domain.
type DomainWithdrawResponse struct {
	// Name is the domain that was withdrawn.
	Name string `json:"name"`

	// Success says whether the withdraw succeeded.
	Success bool `json:"success"`

	// Detail explains the outcome when Success is false.
	Detail *string `json:"detail,omitempty"`
}

// DomainTransitResponse is the outcome of transiting a .de domain.
type DomainTransitResponse struct {
	// Name is the domain that was transited.
	Name string `json:"name"`

	// Success says whether the transit succeeded.
	Success bool `json:"success"`
}

// NorIDDeclarationRequest signs the .no applicant declaration on behalf of the
// applicant. A .no registration stays pending until the declaration is signed.
type NorIDDeclarationRequest struct {
	// AcceptName is the full name of the person signing the declaration. For a
	// private individual this is the subscriber; for an organization it must be
	// an authorized representative.
	AcceptName string `json:"accept_name"`

	// AcceptDate is when the declaration was signed. Set it only when the
	// declaration was signed out of band; it defaults to the submission time.
	AcceptDate *time.Time `json:"accept_date,omitempty"`
}

// NorIDDeclarationStatus is the state of a .no applicant declaration.
type NorIDDeclarationStatus string

const (
	// NorIDDeclarationStatusPending awaits the applicant's signature.
	NorIDDeclarationStatusPending NorIDDeclarationStatus = "pending"

	// NorIDDeclarationStatusConfirmed has been signed.
	NorIDDeclarationStatusConfirmed NorIDDeclarationStatus = "confirmed"

	// NorIDDeclarationStatusCompleted has been signed and the create finished.
	NorIDDeclarationStatusCompleted NorIDDeclarationStatus = "completed"

	// NorIDDeclarationStatusExpired lapsed before it was signed.
	NorIDDeclarationStatusExpired NorIDDeclarationStatus = "expired"

	// NorIDDeclarationStatusFailed could not be processed.
	NorIDDeclarationStatusFailed NorIDDeclarationStatus = "failed"
)

// NorIDDeclarationResponse is the .no applicant declaration, including the
// fixed Norwegian contract text that has to be presented to the applicant.
type NorIDDeclarationResponse struct {
	// DomainName is the domain the declaration applies to.
	DomainName string `json:"domain_name"`

	// Status is the state of the declaration.
	Status NorIDDeclarationStatus `json:"status,omitempty"`

	// DeclarationVersion is the version of the declaration text.
	DeclarationVersion string `json:"declaration_version,omitempty"`

	// DeclarationHeader is the fixed Norwegian declaration header.
	DeclarationHeader string `json:"declaration_header,omitempty"`

	// DeclarationIntroduction is the fixed Norwegian introduction.
	DeclarationIntroduction string `json:"declaration_introduction,omitempty"`

	// DeclarationContractText is the fixed Norwegian contract text.
	DeclarationContractText string `json:"declaration_contract_text,omitempty"`

	// ExpiresOn is when the unconfirmed create request expires.
	ExpiresOn *time.Time `json:"expires_on,omitempty"`
}
