// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// OrganizationID is a TypeID for organizations.
type OrganizationID = TypeID

// OrganizationStatus represents the status of an organization.
type OrganizationStatus string

const (
	// OrganizationStatusActive indicates the organization is active.
	OrganizationStatusActive OrganizationStatus = "active"

	// OrganizationStatusInactive indicates the organization is inactive.
	OrganizationStatusInactive OrganizationStatus = "inactive"
)

// OrganizationSortField represents fields that can be used for sorting organizations.
type OrganizationSortField string

const (
	OrganizationSortByCreatedOn   OrganizationSortField = "created_on"
	OrganizationSortByName        OrganizationSortField = "name"
	OrganizationSortByCountryCode OrganizationSortField = "country_code"
)

// Organization represents an organization (account) in OpusDNS.
type Organization struct {
	// OrganizationID is the unique identifier for the organization.
	OrganizationID OrganizationID `json:"organization_id,omitempty"`

	// Name is the organization name.
	Name string `json:"name"`

	// ParentOrganizationID is the ID of the parent organization.
	ParentOrganizationID *OrganizationID `json:"parent_organization_id,omitempty"`

	// KeycloakOrganizationID is the Keycloak organization ID.
	KeycloakOrganizationID *string `json:"keycloak_organization_id,omitempty"`

	// Status is the status of the organization.
	Status OrganizationStatus `json:"status,omitempty"`

	// Address1 is the first line of the organization's address.
	Address1 *string `json:"address_1,omitempty"`

	// Address2 is the second line of the organization's address.
	Address2 *string `json:"address_2,omitempty"`

	// City is the city of the organization's address.
	City *string `json:"city,omitempty"`

	// State is the state or province of the organization's address.
	State *string `json:"state,omitempty"`

	// PostalCode is the postal code of the organization's address.
	PostalCode *string `json:"postal_code,omitempty"`

	// CountryCode is the ISO 3166-1 alpha-2 country code.
	CountryCode *string `json:"country_code,omitempty"`

	// BusinessNumber is the government issued business identifier.
	BusinessNumber *string `json:"business_number,omitempty"`

	// TaxID is the tax ID of the organization.
	TaxID *string `json:"tax_id,omitempty"`

	// TaxIDType is the type of tax ID.
	TaxIDType *string `json:"tax_id_type,omitempty"`

	// TaxRate is the tax rate for the organization.
	TaxRate *string `json:"tax_rate,omitempty"`

	// Currency is the currency used by the organization.
	Currency *Currency `json:"currency,omitempty"`

	// DefaultLocale is the default locale for the organization.
	DefaultLocale *string `json:"default_locale,omitempty"`

	// Attributes contains organization attributes.
	Attributes []OrganizationAttribute `json:"attributes,omitempty"`

	// Users contains the users belonging to this organization.
	Users []User `json:"users,omitempty"`

	// CreatedOn is when the organization was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// DeletedOn is when the organization was deleted.
	DeletedOn *time.Time `json:"deleted_on,omitempty"`
}

// OrganizationAttribute represents a custom attribute for an organization.
type OrganizationAttribute struct {
	// OrganizationAttributeID is the unique identifier for the attribute.
	OrganizationAttributeID int `json:"organization_attribute_id"`

	// Key is the attribute key.
	Key string `json:"key"`

	// Value is the attribute value.
	Value interface{} `json:"value,omitempty"`

	// Private indicates if the attribute is private and not visible to users.
	Private bool `json:"private,omitempty"`

	// Protected indicates if the attribute is protected and cannot be modified by users.
	Protected bool `json:"protected,omitempty"`

	// CreatedOn is when the attribute was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the attribute was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// OrganizationListResponse represents the paginated response when listing organizations.
type OrganizationListResponse struct {
	// Results contains the list of organizations for the current page.
	Results []Organization `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// ListOrganizationsOptions contains options for listing organizations.
type ListOrganizationsOptions struct {
	Page        int
	PageSize    int
	SortBy      OrganizationSortField
	SortOrder   SortOrder
	Search      string
	CountryCode string
}

// OrganizationCreateRequest represents a request to create an organization.
type OrganizationCreateRequest struct {
	// Name is the organization name.
	Name string `json:"name"`

	// ParentOrganizationID is the optional parent organization ID.
	// The public API derives this from the authenticated caller when omitted.
	ParentOrganizationID *OrganizationID `json:"parent_organization_id,omitempty"`

	// Address1 is the first line of the organization's address.
	Address1 *string `json:"address_1,omitempty"`

	// Address2 is the second line of the organization's address.
	Address2 *string `json:"address_2,omitempty"`

	// City is the city of the organization's address.
	City *string `json:"city,omitempty"`

	// State is the state or province of the organization's address.
	State *string `json:"state,omitempty"`

	// PostalCode is the postal code of the organization's address.
	PostalCode *string `json:"postal_code,omitempty"`

	// CountryCode is the ISO 3166-1 alpha-2 country code.
	CountryCode *string `json:"country_code,omitempty"`

	// BusinessNumber is the government issued business identifier.
	BusinessNumber *string `json:"business_number,omitempty"`

	// TaxID is the tax ID of the organization.
	TaxID *string `json:"tax_id,omitempty"`

	// TaxIDType is the type of tax ID.
	TaxIDType *string `json:"tax_id_type,omitempty"`

	// TaxRate is the tax rate for the organization.
	TaxRate *string `json:"tax_rate,omitempty"`

	// Currency is the currency used by the organization.
	Currency *Currency `json:"currency,omitempty"`

	// DefaultLocale is the default locale for the organization.
	DefaultLocale *string `json:"default_locale,omitempty"`

	// Users contains optional initial users to create with the organization.
	Users []UserCreateRequest `json:"users,omitempty"`

	// Attributes contains optional organization attributes.
	Attributes []OrganizationAttributeCreate `json:"attributes,omitempty"`
}

// OrganizationUpdateRequest represents a request to update an organization.
type OrganizationUpdateRequest struct {
	// Name is the organization name.
	Name *string `json:"name,omitempty"`

	// Address1 is the first line of the organization's address.
	Address1 *string `json:"address_1,omitempty"`

	// Address2 is the second line of the organization's address.
	Address2 *string `json:"address_2,omitempty"`

	// City is the city of the organization's address.
	City *string `json:"city,omitempty"`

	// State is the state or province of the organization's address.
	State *string `json:"state,omitempty"`

	// PostalCode is the postal code of the organization's address.
	PostalCode *string `json:"postal_code,omitempty"`

	// CountryCode is the ISO 3166-1 alpha-2 country code.
	CountryCode *string `json:"country_code,omitempty"`

	// BusinessNumber is the government issued business identifier.
	BusinessNumber *string `json:"business_number,omitempty"`

	// TaxID is the tax ID of the organization.
	TaxID *string `json:"tax_id,omitempty"`

	// TaxIDType is the type of tax ID.
	TaxIDType *string `json:"tax_id_type,omitempty"`

	// Currency is the currency used by the organization.
	Currency *Currency `json:"currency,omitempty"`

	// DefaultLocale is the default locale for the organization.
	DefaultLocale *string `json:"default_locale,omitempty"`
}

// OrganizationAttributeCreate represents a request to create an organization attribute.
type OrganizationAttributeCreate struct {
	// Key is the attribute key.
	Key string `json:"key"`

	// Value is the attribute value.
	Value interface{} `json:"value,omitempty"`

	// Private indicates if the attribute is private.
	Private bool `json:"private,omitempty"`

	// Protected indicates if the attribute is protected.
	Protected bool `json:"protected,omitempty"`
}

// OrganizationAttributeUpdateRequest represents a request to update organization attributes.
type OrganizationAttributeUpdateRequest struct {
	// Attributes contains the attributes to set.
	Attributes []OrganizationAttributeCreate `json:"attributes"`
}

// IPRestriction represents an IP restriction for API access.
type IPRestriction struct {
	// IPRestrictionID is the unique identifier for the restriction.
	IPRestrictionID int `json:"ip_restriction_id"`

	// OrganizationID is the organization this restriction belongs to.
	OrganizationID OrganizationID `json:"organization_id"`

	// IPNetwork is the IP address or CIDR network range.
	IPNetwork string `json:"ip_network"`

	// LastUsedOn is when the restriction was last used.
	LastUsedOn *time.Time `json:"last_used_on,omitempty"`

	// CreatedOn is when the restriction was created.
	CreatedOn time.Time `json:"created_on"`
}

// IPRestrictionListResponse represents the paginated response when listing IP restrictions.
type IPRestrictionListResponse struct {
	// Results contains the list of IP restrictions.
	Results []IPRestriction `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// IPRestrictionCreateRequest represents a request to create an IP restriction.
type IPRestrictionCreateRequest struct {
	// IPNetwork is the IP address or CIDR network range.
	IPNetwork string `json:"ip_network"`
}

// IPRestrictionUpdateRequest represents a request to update an IP restriction.
type IPRestrictionUpdateRequest struct {
	// IPNetwork is the IP address or CIDR network range.
	IPNetwork *string `json:"ip_network,omitempty"`

	// LastUsedOn is the timestamp of the last usage.
	LastUsedOn *time.Time `json:"last_used_on,omitempty"`
}

// Role and RoleListResponse moved to rbac.go and were replaced by RoleDefinition to
// match the new public RBAC role model (label/built_in/"resource:scope" permissions).

// OrganizationAttributesResponse represents the response when getting organization attributes.
type OrganizationAttributesResponse struct {
	// Attributes contains the list of attributes.
	Attributes []OrganizationAttribute `json:"attributes"`
}

// ProductPricing represents pricing for a specific product.
type ProductPricing struct {
	// ProductType is the type of product.
	ProductType string `json:"product_type"`

	// ProductReference is a reference for the product (e.g., TLD name).
	ProductReference *string `json:"product_reference,omitempty"`

	// Actions contains pricing by action type.
	Actions map[string]PriceInfo `json:"actions,omitempty"`
}

// PriceInfo represents pricing information.
type PriceInfo struct {
	// Price is the base price.
	Price string `json:"price"`

	// Currency is the currency code.
	Currency Currency `json:"currency"`

	// Period is the pricing period (e.g., 1 year, 2 months).
	Period *PricingPeriod `json:"period,omitempty"`

	// ProductAction is the product action this price applies to.
	ProductAction *string `json:"product_action,omitempty"`

	// ProductClass is the product class this price applies to.
	ProductClass *string `json:"product_class,omitempty"`
}

// PricingPeriod represents a pricing period.
type PricingPeriod struct {
	// Value is the amount of time in the unit.
	Value int `json:"value"`

	// Unit is the unit of the period.
	Unit PeriodUnit `json:"unit"`
}

// BillingTransactionID is a TypeID for billing transactions.
type BillingTransactionID = TypeID

// BillingTransactionProductType represents the product type in a transaction.
type BillingTransactionProductType string

const (
	BillingProductTypeDomain           BillingTransactionProductType = "domain"
	BillingProductTypeZones            BillingTransactionProductType = "zones"
	BillingProductTypeEmailForward     BillingTransactionProductType = "email_forward"
	BillingProductTypeDomainForward    BillingTransactionProductType = "domain_forward"
	BillingProductTypeAccountWallet    BillingTransactionProductType = "account_wallet"
	BillingProductTypeVanityNameserver BillingTransactionProductType = "vanity_nameserver"
)

// BillingTransactionAction represents the action in a transaction.
type BillingTransactionAction string

const (
	BillingActionCreate      BillingTransactionAction = "create"
	BillingActionTransfer    BillingTransactionAction = "transfer"
	BillingActionRenew       BillingTransactionAction = "renew"
	BillingActionRestore     BillingTransactionAction = "restore"
	BillingActionTrade       BillingTransactionAction = "trade"
	BillingActionApplication BillingTransactionAction = "application"
	BillingActionServiceFee  BillingTransactionAction = "service_fee"
	BillingActionWalletTopUp BillingTransactionAction = "wallet_top_up"
)

// BillingTransactionStatus represents the status of a transaction.
type BillingTransactionStatus string

const (
	BillingStatusPending   BillingTransactionStatus = "pending"
	BillingStatusSucceeded BillingTransactionStatus = "succeeded"
	BillingStatusFailed    BillingTransactionStatus = "failed"
	BillingStatusCanceled  BillingTransactionStatus = "canceled"
)

// BillingTransaction represents a billing transaction.
type BillingTransaction struct {
	// BillingTransactionID is the unique identifier for the transaction.
	BillingTransactionID BillingTransactionID `json:"billing_transaction_id"`

	// ProductType is the type of product.
	ProductType BillingTransactionProductType `json:"product_type"`

	// ProductReference is the reference for the product (e.g., domain name).
	ProductReference *string `json:"product_reference,omitempty"`

	// Action is the action performed.
	Action BillingTransactionAction `json:"action"`

	// Status is the transaction status.
	Status BillingTransactionStatus `json:"status"`

	// Volume is the quantity the transaction covers, expressed in Unit.
	Volume string `json:"volume"`

	// Unit is the unit for Volume (e.g. 'y' for years); null when not applicable.
	Unit *string `json:"unit,omitempty"`

	// Price is the base price.
	Price string `json:"price"`

	// TaxRate is the tax rate applied.
	TaxRate string `json:"tax_rate"`

	// TaxAmount is the tax amount.
	TaxAmount string `json:"tax_amount"`

	// Amount is the total amount including tax.
	Amount string `json:"amount"`

	// Currency is the currency code.
	Currency Currency `json:"currency"`

	// OriginalPrice is the original amount in the supplier's base currency (before exchange).
	OriginalPrice *string `json:"original_price,omitempty"`

	// OriginalCurrency is the currency the original price was in (before exchange).
	OriginalCurrency *Currency `json:"original_currency,omitempty"`

	// ExchangeRate is the exchange rate applied to convert from OriginalCurrency to Currency.
	ExchangeRate *string `json:"exchange_rate,omitempty"`

	// CreatedOn is when the transaction was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the transaction was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`

	// CompletedOn is when the transaction was completed.
	CompletedOn *time.Time `json:"completed_on,omitempty"`
}

// BillingTransactionListResponse represents the paginated response when listing transactions.
type BillingTransactionListResponse struct {
	// Results contains the list of transactions for the current page.
	Results []BillingTransaction `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// BillingTransactionSortField represents fields for sorting transactions.
type BillingTransactionSortField string

const (
	BillingTransactionSortByProductType BillingTransactionSortField = "product_type"
	BillingTransactionSortByAction      BillingTransactionSortField = "action"
	BillingTransactionSortByStatus      BillingTransactionSortField = "status"
	BillingTransactionSortByCreatedOn   BillingTransactionSortField = "created_on"
	BillingTransactionSortByCompletedOn BillingTransactionSortField = "completed_on"
)

// ListTransactionsOptions contains options for listing transactions.
type ListTransactionsOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of transactions per page.
	PageSize int

	// SortBy is the field to sort by.
	SortBy BillingTransactionSortField

	// SortOrder is the sort direction.
	SortOrder SortOrder

	// ProductType filters by product type.
	ProductType BillingTransactionProductType

	// Action filters by action type.
	Action BillingTransactionAction

	// Status filters by transaction status.
	Status BillingTransactionStatus

	// CreatedAfter filters transactions created after this time.
	CreatedAfter *time.Time

	// CreatedBefore filters transactions created before this time.
	CreatedBefore *time.Time
}

// InvoiceResponseStatus represents the status of an invoice.
type InvoiceResponseStatus string

const (
	InvoiceStatusDraft     InvoiceResponseStatus = "draft"
	InvoiceStatusFinalized InvoiceResponseStatus = "finalized"
	InvoiceStatusFailed    InvoiceResponseStatus = "failed"
	InvoiceStatusPending   InvoiceResponseStatus = "pending"
	InvoiceStatusVoided    InvoiceResponseStatus = "voided"
)

// InvoiceResponseType represents the type of an invoice.
type InvoiceResponseType string

const (
	InvoiceTypeSubscription       InvoiceResponseType = "subscription"
	InvoiceTypeAddOn              InvoiceResponseType = "add_on"
	InvoiceTypeCredit             InvoiceResponseType = "credit"
	InvoiceTypeOneOff             InvoiceResponseType = "one_off"
	InvoiceTypeAdvanceCharges     InvoiceResponseType = "advance_charges"
	InvoiceTypeProgressiveBilling InvoiceResponseType = "progressive_billing"
)

// InvoiceResponsePaymentStatus represents the payment status of an invoice.
type InvoiceResponsePaymentStatus string

const (
	InvoicePaymentStatusPending   InvoiceResponsePaymentStatus = "pending"
	InvoicePaymentStatusFailed    InvoiceResponsePaymentStatus = "failed"
	InvoicePaymentStatusSucceeded InvoiceResponsePaymentStatus = "succeeded"
)

// Invoice represents a billing invoice.
type Invoice struct {
	// ExternalID is the Lago (external) ID for this invoice.
	ExternalID string `json:"external_id"`

	// Number is the human-readable invoice number.
	Number string `json:"number"`

	// IssuingDate is when the invoice was issued.
	IssuingDate *time.Time `json:"issuing_date,omitempty"`

	// PaymentDueDate is when payment is due.
	PaymentDueDate *time.Time `json:"payment_due_date,omitempty"`

	// InvoiceType is the type of the invoice.
	InvoiceType InvoiceResponseType `json:"invoice_type"`

	// Status is the invoice status.
	Status InvoiceResponseStatus `json:"status"`

	// PaymentStatus is the payment status.
	PaymentStatus InvoiceResponsePaymentStatus `json:"payment_status"`

	// PaymentOverdue indicates whether payment is overdue.
	PaymentOverdue bool `json:"payment_overdue"`

	// Currency is the currency code.
	Currency Currency `json:"currency"`

	// Amount is the total invoice amount.
	Amount string `json:"amount"`

	// FeesAmount is the fees amount.
	FeesAmount string `json:"fees_amount"`

	// TaxesAmount is the taxes amount.
	TaxesAmount string `json:"taxes_amount"`

	// FileURL is the URL to download the invoice PDF.
	FileURL *string `json:"file_url,omitempty"`
}

// InvoiceListResponse represents the paginated response when listing invoices.
type InvoiceListResponse struct {
	// Results contains the list of invoices for the current page.
	Results []Invoice `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}
