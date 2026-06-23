// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// ContactAttributeSetID is a TypeID for contact attribute sets.
type ContactAttributeSetID = TypeID

// ContactAttributeLinkID is a TypeID for contact attribute set links.
type ContactAttributeLinkID = TypeID

// ContactAttributeSet represents a TLD-specific set of registry contact attributes.
type ContactAttributeSet struct {
	// ContactAttributeSetID is the unique identifier of the attribute set.
	ContactAttributeSetID ContactAttributeSetID `json:"contact_attribute_set_id"`

	// OrganizationID is the organization that owns the attribute set.
	OrganizationID OrganizationID `json:"organization_id"`

	// TLD is the TLD this attribute set applies to.
	TLD string `json:"tld"`

	// Label is a human-readable label explaining the purpose of the set.
	Label string `json:"label"`

	// Attributes is the key-value map of registry contact attributes.
	Attributes map[string]string `json:"attributes"`

	// LinkedContacts is the number of contacts linked to this attribute set.
	LinkedContacts int `json:"linked_contacts,omitempty"`

	// CreatedOn is when the attribute set was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the attribute set was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// ContactAttributeSetCreateRequest is the request body for creating an attribute set.
type ContactAttributeSetCreateRequest struct {
	// Label is a human-readable label explaining the purpose of the set.
	Label string `json:"label"`

	// Attributes is the key-value map of registry contact attributes.
	Attributes map[string]string `json:"attributes"`

	// TLD is the TLD this attribute set applies to (e.g. "de").
	TLD string `json:"tld"`
}

// ContactAttributeSetUpdateRequest is the request body for updating an attribute set.
type ContactAttributeSetUpdateRequest struct {
	// Label is the new label (optional).
	Label *string `json:"label,omitempty"`
}

// ContactAttributeSetListResponse is the paginated response when listing attribute sets.
type ContactAttributeSetListResponse struct {
	// Results contains the attribute sets for the current page.
	Results []ContactAttributeSet `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// ListContactAttributeSetsOptions contains options for listing contact attribute sets.
type ListContactAttributeSetsOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of items per page.
	PageSize int
}

// ContactAttributeLink represents a link between a contact and an attribute set.
type ContactAttributeLink struct {
	// ContactAttributeLinkID is the unique identifier of the link.
	ContactAttributeLinkID ContactAttributeLinkID `json:"contact_attribute_link_id"`

	// ContactID is the contact this link belongs to.
	ContactID ContactID `json:"contact_id"`

	// ContactAttributeSetID is the attribute set linked to the contact.
	ContactAttributeSetID ContactAttributeSetID `json:"contact_attribute_set_id"`

	// TLD is the TLD this link applies to.
	TLD string `json:"tld"`

	// CreatedOn is when the link was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`
}

// ContactVerificationClaim is a claim type that can be verified for a contact.
type ContactVerificationClaim string

const (
	ContactVerificationClaimName        ContactVerificationClaim = "NAME"
	ContactVerificationClaimAddress     ContactVerificationClaim = "ADDRESS"
	ContactVerificationClaimEmail       ContactVerificationClaim = "EMAIL"
	ContactVerificationClaimPhone       ContactVerificationClaim = "PHONE"
	ContactVerificationClaimLegalEntity ContactVerificationClaim = "LEGAL_ENTITY"
)

// ContactVerificationState is the verification state of a contact claim.
type ContactVerificationState string

const (
	ContactVerificationStateUnverified ContactVerificationState = "UNVERIFIED"
	ContactVerificationStateVerified   ContactVerificationState = "VERIFIED"
	ContactVerificationStateInProgress ContactVerificationState = "IN_PROGRESS"
	ContactVerificationStateExpired    ContactVerificationState = "EXPIRED"
)

// ContactVerificationMethod is the method used to verify a contact claim.
type ContactVerificationMethod string

const (
	ContactVerificationMethodAuth          ContactVerificationMethod = "AUTH"
	ContactVerificationMethodVDIG          ContactVerificationMethod = "VDIG"
	ContactVerificationMethodElectronicDoc ContactVerificationMethod = "ELECTRONIC_DOCUMENT"
	ContactVerificationMethodPhysicalDoc   ContactVerificationMethod = "PHYSICAL_DOCUMENT"
	ContactVerificationMethodBVR           ContactVerificationMethod = "BVR"
	ContactVerificationMethodPVR           ContactVerificationMethod = "PVR"
	ContactVerificationMethodData          ContactVerificationMethod = "DATA"
	ContactVerificationMethodReachability  ContactVerificationMethod = "REACHABILITY"
)

// ContactVerificationProof is a proof type backing a contact verification.
type ContactVerificationProof string

const (
	ContactVerificationProofIDCard                  ContactVerificationProof = "IDCARD"
	ContactVerificationProofPassport                ContactVerificationProof = "PASSPORT"
	ContactVerificationProofPopulationRegister      ContactVerificationProof = "POPULATION_REGISTER"
	ContactVerificationProofResidencePermit         ContactVerificationProof = "RESIDENCE_PERMIT"
	ContactVerificationProofProofOfArrival          ContactVerificationProof = "PROOF_OF_ARRIVAL"
	ContactVerificationProofDriversLicence          ContactVerificationProof = "DRIVERS_LICENCE"
	ContactVerificationProofCompanyRegister         ContactVerificationProof = "COMPANY_REGISTER"
	ContactVerificationProofCompanyStatement        ContactVerificationProof = "COMPANY_STATEMENT"
	ContactVerificationProofBankAccount             ContactVerificationProof = "BANK_ACCOUNT"
	ContactVerificationProofOnlinePaymentAccount    ContactVerificationProof = "ONLINE_PAYMENT_ACCOUNT"
	ContactVerificationProofUtilityAccount          ContactVerificationProof = "UTILITY_ACCOUNT"
	ContactVerificationProofBankStatement           ContactVerificationProof = "BANK_STATEMENT"
	ContactVerificationProofTaxStatement            ContactVerificationProof = "TAX_STATEMENT"
	ContactVerificationProofWrittenAttestation      ContactVerificationProof = "WRITTEN_ATTESTATION"
	ContactVerificationProofDigitalAttestation      ContactVerificationProof = "DIGITAL_ATTESTATION"
	ContactVerificationProofPostalVerTransactionLog ContactVerificationProof = "POSTAL_VER_TRANSACTION_LOG"
	ContactVerificationProofEmailVerTransactionLog  ContactVerificationProof = "EMAIL_VER_TRANSACTION_LOG"
	ContactVerificationProofAddressDatabase         ContactVerificationProof = "ADDRESS_DATABASE"
)

// ContactAttestVerificationRequest is a single contact-verification attestation.
type ContactAttestVerificationRequest struct {
	// Claim is the claim being attested.
	Claim ContactVerificationClaim `json:"claim"`

	// Method is the verification method used.
	Method ContactVerificationMethod `json:"method"`

	// Proof is the proof type backing the attestation.
	Proof ContactVerificationProof `json:"proof"`

	// AttestationReference is an optional reference for the attestation.
	AttestationReference string `json:"attestation_reference"`
}

// ContactAttestRequest is the request body for attesting contact verifications.
type ContactAttestRequest struct {
	// Attestations is the list of attestations to submit (max 50).
	Attestations []ContactAttestVerificationRequest `json:"attestations"`
}

// ContactVerificationStatus is the verification state of a single contact claim.
type ContactVerificationStatus struct {
	// Claim is the claim this status applies to.
	Claim ContactVerificationClaim `json:"claim"`

	// State is the current verification state.
	State ContactVerificationState `json:"state"`

	// Method is the verification method used, if any.
	Method *ContactVerificationMethod `json:"method,omitempty"`

	// Proof is the proof type backing the verification, if any.
	Proof *ContactVerificationProof `json:"proof,omitempty"`

	// AttestationReference is the reference for the attestation, if any.
	AttestationReference *string `json:"attestation_reference,omitempty"`

	// VerifiedOn is when the claim was verified, if applicable.
	VerifiedOn *time.Time `json:"verified_on,omitempty"`

	// ExpiresOn is when the verification expires, if applicable.
	ExpiresOn *time.Time `json:"expires_on,omitempty"`
}

// ContactAttestResponse is the response with the per-claim verification state.
type ContactAttestResponse struct {
	// Verifications is the list of per-claim verification states.
	Verifications []ContactVerificationStatus `json:"verifications"`
}
