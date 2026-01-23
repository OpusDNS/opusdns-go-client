// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// ContactID is a TypeID for contacts.
type ContactID = TypeID

// ContactSortField represents fields that can be used for sorting contacts.
type ContactSortField string

const (
	ContactSortByFirstName ContactSortField = "first_name"
	ContactSortByLastName  ContactSortField = "last_name"
	ContactSortByEmail     ContactSortField = "email"
	ContactSortByCreatedOn ContactSortField = "created_on"
	ContactSortByUpdatedOn ContactSortField = "updated_on"
)

// RegistryHandleAttributeType represents TLD-specific contact attribute types.
type RegistryHandleAttributeType string

const (
	// Common registry-specific attributes
	RegistryAttrVATID              RegistryHandleAttributeType = "vat_id"
	RegistryAttrCompanyNumber      RegistryHandleAttributeType = "company_number"
	RegistryAttrDateOfBirth        RegistryHandleAttributeType = "date_of_birth"
	RegistryAttrPlaceOfBirth       RegistryHandleAttributeType = "place_of_birth"
	RegistryAttrNationality        RegistryHandleAttributeType = "nationality"
	RegistryAttrIDCardNumber       RegistryHandleAttributeType = "id_card_number"
	RegistryAttrPassportNumber     RegistryHandleAttributeType = "passport_number"
	RegistryAttrLanguage           RegistryHandleAttributeType = "language"
	RegistryAttrEntityType         RegistryHandleAttributeType = "entity_type"
	RegistryAttrRegistrantType     RegistryHandleAttributeType = "registrant_type"
	RegistryAttrIntendedUse        RegistryHandleAttributeType = "intended_use"
	RegistryAttrNexusCategory      RegistryHandleAttributeType = "nexus_category"
	RegistryAttrNexusCountry       RegistryHandleAttributeType = "nexus_country"
	RegistryAttrAppPurpose         RegistryHandleAttributeType = "app_purpose"
	RegistryAttrSIREN              RegistryHandleAttributeType = "siren"
	RegistryAttrSIRET              RegistryHandleAttributeType = "siret"
	RegistryAttrTrademarkNumber    RegistryHandleAttributeType = "trademark_number"
	RegistryAttrTrademarkCountry   RegistryHandleAttributeType = "trademark_country"
	RegistryAttrAuthID             RegistryHandleAttributeType = "auth_id"
	RegistryAttrIdentificationForm RegistryHandleAttributeType = "identification_form"
)

// Contact represents a contact in the system.
type Contact struct {
	// ContactID is the unique identifier for the contact.
	ContactID ContactID `json:"contact_id"`

	// FirstName is the contact's first name.
	FirstName string `json:"first_name"`

	// LastName is the contact's last name.
	LastName string `json:"last_name"`

	// Org is the contact's organization (optional).
	Org *string `json:"org,omitempty"`

	// Title is the contact's title (optional, e.g., "Mr.", "Dr.").
	Title *string `json:"title,omitempty"`

	// Email is the contact's email address.
	Email string `json:"email"`

	// Phone is the contact's phone number in E.164 format.
	Phone string `json:"phone"`

	// Fax is the contact's fax number (optional).
	Fax *string `json:"fax,omitempty"`

	// Street is the street address.
	Street string `json:"street"`

	// City is the city.
	City string `json:"city"`

	// State is the state or province (optional).
	State *string `json:"state,omitempty"`

	// PostalCode is the postal or ZIP code.
	PostalCode string `json:"postal_code"`

	// Country is the two-letter country code (ISO 3166-1 alpha-2).
	Country string `json:"country"`

	// Disclose indicates whether contact information should be publicly disclosed.
	Disclose bool `json:"disclose"`

	// Verified indicates whether the contact has been verified.
	Verified bool `json:"verified,omitempty"`

	// VerifiedOn is when the contact was verified.
	VerifiedOn *time.Time `json:"verified_on,omitempty"`

	// CreatedOn is when the contact was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the contact was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// FullName returns the contact's full name.
func (c *Contact) FullName() string {
	if c.Title != nil && *c.Title != "" {
		return *c.Title + " " + c.FirstName + " " + c.LastName
	}
	return c.FirstName + " " + c.LastName
}

// ContactListResponse represents the paginated response when listing contacts.
type ContactListResponse struct {
	// Results contains the list of contacts for the current page.
	Results []Contact `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// ContactCreateRequest represents a request to create a new contact.
type ContactCreateRequest struct {
	// FirstName is the contact's first name.
	FirstName string `json:"first_name"`

	// LastName is the contact's last name.
	LastName string `json:"last_name"`

	// Org is the contact's organization (optional).
	Org *string `json:"org,omitempty"`

	// Title is the contact's title (optional).
	Title *string `json:"title,omitempty"`

	// Email is the contact's email address.
	Email string `json:"email"`

	// Phone is the contact's phone number in E.164 format (e.g., "+1.2125551234").
	Phone string `json:"phone"`

	// Fax is the contact's fax number (optional).
	Fax *string `json:"fax,omitempty"`

	// Street is the street address.
	Street string `json:"street"`

	// City is the city.
	City string `json:"city"`

	// State is the state or province (optional).
	State *string `json:"state,omitempty"`

	// PostalCode is the postal or ZIP code.
	PostalCode string `json:"postal_code"`

	// Country is the two-letter country code (ISO 3166-1 alpha-2).
	Country string `json:"country"`

	// Disclose indicates whether contact information should be publicly disclosed.
	Disclose bool `json:"disclose"`
}

// ContactUpdateRequest represents a request to update an existing contact.
type ContactUpdateRequest struct {
	// FirstName is the contact's first name.
	FirstName *string `json:"first_name,omitempty"`

	// LastName is the contact's last name.
	LastName *string `json:"last_name,omitempty"`

	// Org is the contact's organization.
	Org *string `json:"org,omitempty"`

	// Title is the contact's title.
	Title *string `json:"title,omitempty"`

	// Email is the contact's email address.
	Email *string `json:"email,omitempty"`

	// Phone is the contact's phone number.
	Phone *string `json:"phone,omitempty"`

	// Fax is the contact's fax number.
	Fax *string `json:"fax,omitempty"`

	// Street is the street address.
	Street *string `json:"street,omitempty"`

	// City is the city.
	City *string `json:"city,omitempty"`

	// State is the state or province.
	State *string `json:"state,omitempty"`

	// PostalCode is the postal or ZIP code.
	PostalCode *string `json:"postal_code,omitempty"`

	// Country is the two-letter country code.
	Country *string `json:"country,omitempty"`

	// Disclose indicates whether contact information should be publicly disclosed.
	Disclose *bool `json:"disclose,omitempty"`
}

// ContactVerification represents a contact verification request/response.
type ContactVerification struct {
	// ContactID is the ID of the contact being verified.
	ContactID ContactID `json:"contact_id"`

	// Status is the verification status.
	Status string `json:"status"`

	// VerificationURL is the URL for the contact to complete verification.
	VerificationURL *string `json:"verification_url,omitempty"`

	// ExpiresOn is when the verification request expires.
	ExpiresOn *time.Time `json:"expires_on,omitempty"`

	// CreatedOn is when the verification was requested.
	CreatedOn *time.Time `json:"created_on,omitempty"`
}

// ContactVerificationRequest represents a request to verify a contact.
type ContactVerificationRequest struct {
	// Token is the verification token (when verifying).
	Token string `json:"token,omitempty"`
}

// ListContactsOptions contains options for listing contacts.
type ListContactsOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of contacts per page.
	PageSize int

	// SortBy is the field to sort by.
	SortBy ContactSortField

	// SortOrder is the sort direction.
	SortOrder SortOrder

	// Search is an optional search query to filter contacts.
	Search string

	// FirstName filters by first name.
	FirstName string

	// LastName filters by last name.
	LastName string

	// Email filters by email address.
	Email string

	// Country filters by country code.
	Country string

	// Verified filters by verification status.
	Verified *bool
}

// ContactAttributeDefinition defines a TLD-specific contact attribute.
type ContactAttributeDefinition struct {
	// Key is the unique identifier for the attribute.
	Key RegistryHandleAttributeType `json:"key"`

	// Type is the data type of the attribute.
	Type string `json:"type"`

	// Values contains allowed values for enum types.
	Values []string `json:"values,omitempty"`

	// Required indicates if this attribute is required.
	Required bool `json:"required,omitempty"`

	// Description provides a human-readable description.
	Description string `json:"description,omitempty"`
}

// ContactRoleAttributeRequirement defines attribute requirements for a contact role.
type ContactRoleAttributeRequirement struct {
	// Role is the contact role this requirement applies to.
	Role DomainContactType `json:"role"`

	// Attributes is the list of required attribute keys.
	Attributes []RegistryHandleAttributeType `json:"attributes"`
}
