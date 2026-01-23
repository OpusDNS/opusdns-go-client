// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// OrganizationID is a TypeID for organizations.
type OrganizationID = TypeID

// Organization represents an organization (account) in OpusDNS.
type Organization struct {
	// OrganizationID is the unique identifier for the organization.
	OrganizationID OrganizationID `json:"organization_id"`

	// Name is the organization name.
	Name string `json:"name"`

	// Email is the primary contact email for the organization.
	Email string `json:"email,omitempty"`

	// Phone is the primary contact phone number.
	Phone *string `json:"phone,omitempty"`

	// Address contains the organization's address.
	Address *OrganizationAddress `json:"address,omitempty"`

	// BillingPlan contains the current billing plan.
	BillingPlan *BillingPlan `json:"billing_plan,omitempty"`

	// BillingMetadata contains additional billing information.
	BillingMetadata *BillingMetadata `json:"billing_metadata,omitempty"`

	// Settings contains organization-specific settings.
	Settings *OrganizationSettings `json:"settings,omitempty"`

	// Verified indicates if the organization has been verified.
	Verified bool `json:"verified,omitempty"`

	// Active indicates if the organization is active.
	Active bool `json:"active,omitempty"`

	// CreatedOn is when the organization was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the organization was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// OrganizationAddress represents an organization's address.
type OrganizationAddress struct {
	// Street is the street address.
	Street string `json:"street,omitempty"`

	// City is the city.
	City string `json:"city,omitempty"`

	// State is the state or province.
	State *string `json:"state,omitempty"`

	// PostalCode is the postal or ZIP code.
	PostalCode string `json:"postal_code,omitempty"`

	// Country is the two-letter country code (ISO 3166-1 alpha-2).
	Country string `json:"country,omitempty"`
}

// OrganizationSettings contains organization-specific settings.
type OrganizationSettings struct {
	// DefaultTTL is the default TTL for DNS records.
	DefaultTTL int `json:"default_ttl,omitempty"`

	// DefaultNameservers is the default list of nameservers.
	DefaultNameservers []string `json:"default_nameservers,omitempty"`

	// AutoRenewEnabled is the default auto-renew setting for domains.
	AutoRenewEnabled bool `json:"auto_renew_enabled,omitempty"`

	// TransferLockEnabled is the default transfer lock setting for domains.
	TransferLockEnabled bool `json:"transfer_lock_enabled,omitempty"`

	// TwoFactorRequired indicates if 2FA is required for users.
	TwoFactorRequired bool `json:"two_factor_required,omitempty"`

	// IPRestrictionsEnabled indicates if IP restrictions are enabled.
	IPRestrictionsEnabled bool `json:"ip_restrictions_enabled,omitempty"`
}

// BillingPlan represents a billing plan.
type BillingPlan struct {
	// PlanID is the unique identifier for the plan.
	PlanID *string `json:"plan_id,omitempty"`

	// Name is the plan name.
	Name *string `json:"name,omitempty"`

	// PlanLevel is the plan level (e.g., "basic", "premium", "enterprise").
	PlanLevel *string `json:"plan_level,omitempty"`

	// Type is the plan type or billing interval.
	Type *string `json:"type,omitempty"`

	// Amount is the plan price.
	Amount string `json:"amount,omitempty"`

	// Currency is the plan currency.
	Currency Currency `json:"currency,omitempty"`
}

// BillingMetadata contains additional billing information.
type BillingMetadata struct {
	// CustomerNumber is the customer account number.
	CustomerNumber *int `json:"customer_number,omitempty"`

	// BillingModel is the payment terms.
	BillingModel *string `json:"billing_model,omitempty"`

	// CreditLimit is the credit limit for the organization.
	CreditLimit *int `json:"credit_limit,omitempty"`

	// WalletBalance is the current wallet balance.
	WalletBalance *string `json:"wallet_balance,omitempty"`

	// Currency is the billing currency.
	Currency Currency `json:"currency,omitempty"`
}

// OrganizationListResponse represents the paginated response when listing organizations.
type OrganizationListResponse struct {
	// Results contains the list of organizations for the current page.
	Results []Organization `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// OrganizationUpdateRequest represents a request to update an organization.
type OrganizationUpdateRequest struct {
	// Name is the organization name.
	Name *string `json:"name,omitempty"`

	// Email is the primary contact email.
	Email *string `json:"email,omitempty"`

	// Phone is the primary contact phone number.
	Phone *string `json:"phone,omitempty"`

	// Address is the organization address.
	Address *OrganizationAddress `json:"address,omitempty"`

	// Settings is the organization settings.
	Settings *OrganizationSettings `json:"settings,omitempty"`
}

// IPRestriction represents an IP restriction for API access.
type IPRestriction struct {
	// IPRestrictionID is the unique identifier for the restriction.
	IPRestrictionID TypeID `json:"ip_restriction_id"`

	// CIDR is the IP address or CIDR range (e.g., "192.0.2.0/24").
	CIDR string `json:"cidr"`

	// Description is an optional description of the restriction.
	Description *string `json:"description,omitempty"`

	// Enabled indicates if the restriction is active.
	Enabled bool `json:"enabled"`

	// CreatedOn is when the restriction was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the restriction was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// IPRestrictionListResponse represents the paginated response when listing IP restrictions.
type IPRestrictionListResponse struct {
	// Results contains the list of IP restrictions for the current page.
	Results []IPRestriction `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// IPRestrictionCreateRequest represents a request to create an IP restriction.
type IPRestrictionCreateRequest struct {
	// CIDR is the IP address or CIDR range.
	CIDR string `json:"cidr"`

	// Description is an optional description.
	Description *string `json:"description,omitempty"`

	// Enabled indicates if the restriction should be active.
	Enabled bool `json:"enabled"`
}

// IPRestrictionUpdateRequest represents a request to update an IP restriction.
type IPRestrictionUpdateRequest struct {
	// CIDR is the IP address or CIDR range.
	CIDR *string `json:"cidr,omitempty"`

	// Description is the description.
	Description *string `json:"description,omitempty"`

	// Enabled indicates if the restriction should be active.
	Enabled *bool `json:"enabled,omitempty"`
}

// Role represents a user role within an organization.
type Role struct {
	// RoleID is the unique identifier for the role.
	RoleID TypeID `json:"role_id"`

	// Name is the role name.
	Name string `json:"name"`

	// Description is the role description.
	Description *string `json:"description,omitempty"`

	// Permissions is the list of permissions for the role.
	Permissions []string `json:"permissions,omitempty"`

	// IsSystem indicates if this is a system-defined role.
	IsSystem bool `json:"is_system,omitempty"`

	// CreatedOn is when the role was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the role was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// RoleListResponse represents the response when listing roles.
type RoleListResponse struct {
	// Results contains the list of roles.
	Results []Role `json:"results"`
}

// OrganizationAttribute represents a custom attribute for an organization.
type OrganizationAttribute struct {
	// Key is the attribute key.
	Key string `json:"key"`

	// Value is the attribute value.
	Value string `json:"value"`

	// UpdatedOn is when the attribute was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// OrganizationAttributesResponse represents the response when getting organization attributes.
type OrganizationAttributesResponse struct {
	// Attributes contains the list of attributes.
	Attributes []OrganizationAttribute `json:"attributes"`
}

// OrganizationAttributeUpdateRequest represents a request to update organization attributes.
type OrganizationAttributeUpdateRequest struct {
	// Attributes contains the attributes to set.
	Attributes map[string]string `json:"attributes"`
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

	// TaxRate is the applicable tax rate.
	TaxRate *string `json:"tax_rate,omitempty"`

	// TotalPrice is the price including tax.
	TotalPrice *string `json:"total_price,omitempty"`
}

// BillingTransactionID is a TypeID for billing transactions.
type BillingTransactionID = TypeID

// BillingTransactionProductType represents the product type in a transaction.
type BillingTransactionProductType string

const (
	BillingProductTypeDomain        BillingTransactionProductType = "domain"
	BillingProductTypeZones         BillingTransactionProductType = "zones"
	BillingProductTypeEmailForward  BillingTransactionProductType = "email_forward"
	BillingProductTypeDomainForward BillingTransactionProductType = "domain_forward"
	BillingProductTypeAccountWallet BillingTransactionProductType = "account_wallet"
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

// Invoice represents a billing invoice.
type Invoice struct {
	// InvoiceID is the unique identifier for the invoice.
	InvoiceID TypeID `json:"invoice_id"`

	// InvoiceNumber is the human-readable invoice number.
	InvoiceNumber string `json:"invoice_number"`

	// Status is the invoice status.
	Status string `json:"status"`

	// Amount is the total invoice amount.
	Amount string `json:"amount"`

	// Currency is the currency code.
	Currency Currency `json:"currency"`

	// DueDate is when the invoice is due.
	DueDate *time.Time `json:"due_date,omitempty"`

	// PaidOn is when the invoice was paid.
	PaidOn *time.Time `json:"paid_on,omitempty"`

	// CreatedOn is when the invoice was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// DownloadURL is the URL to download the invoice PDF.
	DownloadURL *string `json:"download_url,omitempty"`
}

// InvoiceListResponse represents the paginated response when listing invoices.
type InvoiceListResponse struct {
	// Results contains the list of invoices for the current page.
	Results []Invoice `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}
