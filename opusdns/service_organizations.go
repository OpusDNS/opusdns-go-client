package opusdns

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/opusdns/opusdns-go-client/models"
)

// OrganizationsService provides methods for managing organizations.
type OrganizationsService struct {
	client *Client
}

// ListOrganizations retrieves all child organizations with automatic pagination.
func (s *OrganizationsService) ListOrganizations(ctx context.Context, opts *models.ListOrganizationsOptions) ([]models.Organization, error) {
	var all []models.Organization
	page := 1

	for {
		pageOpts := cloneOptions(opts)
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := s.ListOrganizationsPage(ctx, pageOpts)
		if err != nil {
			return nil, err
		}

		all = append(all, resp.Results...)

		if !resp.Pagination.HasNextPage {
			break
		}
		page++
	}

	return all, nil
}

// ListOrganizationsPage retrieves a single page of child organizations.
func (s *OrganizationsService) ListOrganizationsPage(ctx context.Context, opts *models.ListOrganizationsOptions) (*models.OrganizationListResponse, error) {
	path := s.client.http.BuildPath("organizations")

	query := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.SortBy != "" {
			query.Set("sort_by", string(opts.SortBy))
		}
		if opts.SortOrder != "" {
			query.Set("sort_order", string(opts.SortOrder))
		}
		if opts.Search != "" {
			query.Set("search", opts.Search)
		}
		if opts.CountryCode != "" {
			query.Set("country_code", opts.CountryCode)
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.OrganizationListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetOrganization retrieves an organization by ID.
func (s *OrganizationsService) GetOrganization(ctx context.Context, orgID models.OrganizationID) (*models.Organization, error) {
	path := s.client.http.BuildPath("organizations", url.PathEscape(string(orgID)))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var org models.Organization
	if err := s.client.http.DecodeResponse(resp, &org); err != nil {
		return nil, err
	}

	return &org, nil
}

// CreateOrganization creates a child organization under the authenticated organization.
func (s *OrganizationsService) CreateOrganization(ctx context.Context, req *models.OrganizationCreateRequest) (*models.Organization, error) {
	path := s.client.http.BuildPath("organizations")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var org models.Organization
	if err := s.client.http.DecodeResponse(resp, &org); err != nil {
		return nil, err
	}

	return &org, nil
}

// UpdateOrganization updates an organization.
func (s *OrganizationsService) UpdateOrganization(ctx context.Context, orgID models.OrganizationID, req *models.OrganizationUpdateRequest) (*models.Organization, error) {
	path := s.client.http.BuildPath("organizations", url.PathEscape(string(orgID)))

	resp, err := s.client.http.Patch(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var org models.Organization
	if err := s.client.http.DecodeResponse(resp, &org); err != nil {
		return nil, err
	}

	return &org, nil
}

// DeleteOrganization deletes an organization.
func (s *OrganizationsService) DeleteOrganization(ctx context.Context, orgID models.OrganizationID) error {
	path := s.client.http.BuildPath("organizations", url.PathEscape(string(orgID)))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// ListIPRestrictions retrieves the IP restrictions of the organization. The
// endpoint is not paginated: every restriction comes back in Results.
func (s *OrganizationsService) ListIPRestrictions(ctx context.Context) (*models.IPRestrictionListResponse, error) {
	path := s.client.http.BuildPath("organizations", "ip-restrictions")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var restrictions []models.IPRestriction
	if err := s.client.http.DecodeResponse(resp, &restrictions); err != nil {
		return nil, err
	}

	return &models.IPRestrictionListResponse{Results: restrictions}, nil
}

// GetIPRestriction retrieves a specific IP restriction by its ID, which is an
// integer (IPRestriction.IPRestrictionID), not a TypeID.
func (s *OrganizationsService) GetIPRestriction(ctx context.Context, restrictionID int) (*models.IPRestriction, error) {
	path := s.client.http.BuildPath("organizations", "ip-restrictions", strconv.Itoa(restrictionID))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var restriction models.IPRestriction
	if err := s.client.http.DecodeResponse(resp, &restriction); err != nil {
		return nil, err
	}

	return &restriction, nil
}

// CreateIPRestriction creates a new IP restriction.
func (s *OrganizationsService) CreateIPRestriction(ctx context.Context, req *models.IPRestrictionCreateRequest) (*models.IPRestriction, error) {
	path := s.client.http.BuildPath("organizations", "ip-restrictions")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var restriction models.IPRestriction
	if err := s.client.http.DecodeResponse(resp, &restriction); err != nil {
		return nil, err
	}

	return &restriction, nil
}

// UpdateIPRestriction updates an IP restriction.
func (s *OrganizationsService) UpdateIPRestriction(ctx context.Context, restrictionID int, req *models.IPRestrictionUpdateRequest) (*models.IPRestriction, error) {
	path := s.client.http.BuildPath("organizations", "ip-restrictions", strconv.Itoa(restrictionID))

	resp, err := s.client.http.Patch(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var restriction models.IPRestriction
	if err := s.client.http.DecodeResponse(resp, &restriction); err != nil {
		return nil, err
	}

	return &restriction, nil
}

// DeleteIPRestriction deletes an IP restriction.
func (s *OrganizationsService) DeleteIPRestriction(ctx context.Context, restrictionID int) error {
	path := s.client.http.BuildPath("organizations", "ip-restrictions", strconv.Itoa(restrictionID))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// ListRoles retrieves all roles assignable in the organization: the built-in roles
// plus the organization's custom roles.
func (s *OrganizationsService) ListRoles(ctx context.Context) ([]models.RoleDefinition, error) {
	path := s.client.http.BuildPath("organizations", "roles")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var roles []models.RoleDefinition
	if err := s.client.http.DecodeResponse(resp, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

// GetRole retrieves a single role (built-in or custom) by its label.
func (s *OrganizationsService) GetRole(ctx context.Context, label string) (*models.RoleDefinition, error) {
	path := s.client.http.BuildPath("organizations", "roles", url.PathEscape(label))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var role models.RoleDefinition
	if err := s.client.http.DecodeResponse(resp, &role); err != nil {
		return nil, err
	}

	return &role, nil
}

// CreateRole creates an organization-owned custom role. A custom role must
// grant at least one permission.
func (s *OrganizationsService) CreateRole(ctx context.Context, req *models.CustomRoleCreateRequest) (*models.RoleDefinition, error) {
	if req == nil || len(req.Permissions) == 0 {
		return nil, &ValidationError{Field: "permissions", Message: "a custom role must grant at least one permission"}
	}

	path := s.client.http.BuildPath("organizations", "roles")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var role models.RoleDefinition
	if err := s.client.http.DecodeResponse(resp, &role); err != nil {
		return nil, err
	}

	return &role, nil
}

// UpdateRole updates a custom role's name, description and/or permission set.
// Built-in roles are immutable.
func (s *OrganizationsService) UpdateRole(ctx context.Context, label string, req *models.CustomRoleUpdateRequest) (*models.RoleDefinition, error) {
	path := s.client.http.BuildPath("organizations", "roles", url.PathEscape(label))

	resp, err := s.client.http.Patch(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var role models.RoleDefinition
	if err := s.client.http.DecodeResponse(resp, &role); err != nil {
		return nil, err
	}

	return &role, nil
}

// DeleteRole deletes a custom role. The request is refused while the role is still
// assigned to any subject, and for built-in roles.
func (s *OrganizationsService) DeleteRole(ctx context.Context, label string) error {
	path := s.client.http.BuildPath("organizations", "roles", url.PathEscape(label))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// ListRolePermissions retrieves the catalog of "resource:scope" permissions a custom
// role may grant.
func (s *OrganizationsService) ListRolePermissions(ctx context.Context) (*models.PermissionCatalogResponse, error) {
	path := s.client.http.BuildPath("organizations", "role-permissions")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.PermissionCatalogResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetCurrentAttributes retrieves the authenticated organization's attributes.
func (s *OrganizationsService) GetCurrentAttributes(ctx context.Context) (*models.OrganizationAttributesResponse, error) {
	path := s.client.http.BuildPath("organizations", "attributes")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	return s.decodeAttributes(resp)
}

// UpdateCurrentAttributes updates the authenticated organization's attributes.
func (s *OrganizationsService) UpdateCurrentAttributes(ctx context.Context, req *models.OrganizationAttributeUpdateRequest) (*models.OrganizationAttributesResponse, error) {
	body, err := attributeUpdateBody(req)
	if err != nil {
		return nil, err
	}

	path := s.client.http.BuildPath("organizations", "attributes")

	resp, err := s.client.http.Patch(ctx, path, body)
	if err != nil {
		return nil, err
	}

	return s.decodeAttributes(resp)
}

// GetAttributes retrieves organization attributes.
func (s *OrganizationsService) GetAttributes(ctx context.Context, orgID models.OrganizationID) (*models.OrganizationAttributesResponse, error) {
	path := s.client.http.BuildPath("organizations", url.PathEscape(string(orgID)), "attributes")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	return s.decodeAttributes(resp)
}

// UpdateAttributes updates organization attributes.
func (s *OrganizationsService) UpdateAttributes(ctx context.Context, orgID models.OrganizationID, req *models.OrganizationAttributeUpdateRequest) (*models.OrganizationAttributesResponse, error) {
	body, err := attributeUpdateBody(req)
	if err != nil {
		return nil, err
	}

	path := s.client.http.BuildPath("organizations", url.PathEscape(string(orgID)), "attributes")

	resp, err := s.client.http.Patch(ctx, path, body)
	if err != nil {
		return nil, err
	}

	return s.decodeAttributes(resp)
}

// attributeUpdateBody returns the attributes of req as the bare array the API
// takes. A nil slice is sent as an empty array rather than null.
func attributeUpdateBody(req *models.OrganizationAttributeUpdateRequest) ([]models.OrganizationAttributeCreate, error) {
	if req == nil {
		return nil, &ValidationError{Field: "request", Message: "an attribute update request is required"}
	}
	if req.Attributes == nil {
		return []models.OrganizationAttributeCreate{}, nil
	}
	return req.Attributes, nil
}

// decodeAttributes reads the bare attribute array the API answers with.
func (s *OrganizationsService) decodeAttributes(resp *Response) (*models.OrganizationAttributesResponse, error) {
	var attributes []models.OrganizationAttribute
	if err := s.client.http.DecodeResponse(resp, &attributes); err != nil {
		return nil, err
	}

	return &models.OrganizationAttributesResponse{Attributes: attributes}, nil
}

// ListTransactions retrieves one page of billing transactions for an organization.
func (s *OrganizationsService) ListTransactions(ctx context.Context, orgID models.OrganizationID, opts *models.ListTransactionsOptions) (*models.BillingTransactionListResponse, error) {
	path := s.client.http.BuildPath("organizations", url.PathEscape(string(orgID)), "transactions")

	query := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.SortBy != "" {
			query.Set("sort_by", string(opts.SortBy))
		}
		if opts.SortOrder != "" {
			query.Set("sort_order", string(opts.SortOrder))
		}
		if opts.ProductReference != "" {
			query.Set("product_reference", opts.ProductReference)
		}
		if opts.ProductType != "" {
			query.Set("product_type", string(opts.ProductType))
		}
		if opts.Action != "" {
			query.Set("action", string(opts.Action))
		}
		if opts.Status != "" {
			query.Set("status", string(opts.Status))
		}
		if opts.CreatedAfter != nil {
			query.Set("created_after", opts.CreatedAfter.Format(time.RFC3339Nano))
		}
		if opts.CreatedBefore != nil {
			query.Set("created_before", opts.CreatedBefore.Format(time.RFC3339Nano))
		}
		if opts.CompletedAfter != nil {
			query.Set("completed_after", opts.CompletedAfter.Format(time.RFC3339Nano))
		}
		if opts.CompletedBefore != nil {
			query.Set("completed_before", opts.CompletedBefore.Format(time.RFC3339Nano))
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.BillingTransactionListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTransaction retrieves a specific transaction by ID.
func (s *OrganizationsService) GetTransaction(ctx context.Context, orgID models.OrganizationID, transactionID models.BillingTransactionID) (*models.BillingTransaction, error) {
	path := s.client.http.BuildPath("organizations", url.PathEscape(string(orgID)), "transactions", url.PathEscape(string(transactionID)))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var transaction models.BillingTransaction
	if err := s.client.http.DecodeResponse(resp, &transaction); err != nil {
		return nil, err
	}

	return &transaction, nil
}

// invoiceQuery turns billing-document pagination options into query
// parameters. The endpoints default to ten documents per page, so a caller
// that wants more has to ask.
func invoiceQuery(opts *models.ListInvoicesOptions) url.Values {
	query := url.Values{}
	if opts == nil {
		return query
	}
	if opts.Page > 0 {
		query.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.PageSize > 0 {
		query.Set("page_size", strconv.Itoa(opts.PageSize))
	}
	return query
}

// ListInvoicesPage retrieves one page of invoices for an organization.
func (s *OrganizationsService) ListInvoicesPage(ctx context.Context, orgID models.OrganizationID, opts *models.ListInvoicesOptions) (*models.InvoiceListResponse, error) {
	path := s.client.http.BuildPath("organizations", url.PathEscape(string(orgID)), "billing", "invoices")

	resp, err := s.client.http.Get(ctx, path, invoiceQuery(opts))
	if err != nil {
		return nil, err
	}

	var result models.InvoiceListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ListInvoices retrieves every invoice of an organization, following pagination.
func (s *OrganizationsService) ListInvoices(ctx context.Context, orgID models.OrganizationID, opts *models.ListInvoicesOptions) ([]models.Invoice, error) {
	return s.listBillingDocuments(ctx, opts, func(ctx context.Context, pageOpts *models.ListInvoicesOptions) (*models.InvoiceListResponse, error) {
		return s.ListInvoicesPage(ctx, orgID, pageOpts)
	})
}

// GetPricing retrieves the organization's prices for a product type (for
// example "domain"). opts may narrow the result to one action or product
// class and may be nil.
func (s *OrganizationsService) GetPricing(ctx context.Context, orgID models.OrganizationID, productType string, opts *models.PricingOptions) (*models.ProductPricing, error) {
	path := s.client.http.BuildPath("organizations", url.PathEscape(string(orgID)), "pricing", "product-type", url.PathEscape(productType))

	query := url.Values{}
	if opts != nil {
		if opts.ProductAction != "" {
			query.Set("product_action", string(opts.ProductAction))
		}
		if opts.ProductClass != "" {
			query.Set("product_class", opts.ProductClass)
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var pricing models.ProductPricing
	if err := s.client.http.DecodeResponse(resp, &pricing); err != nil {
		return nil, err
	}

	return &pricing, nil
}

// ListReceiptsPage retrieves one page of the payment receipts of an
// organization. Receipts and invoices share a shape and are told apart by
// their document type.
func (s *OrganizationsService) ListReceiptsPage(ctx context.Context, orgID models.OrganizationID, opts *models.ListInvoicesOptions) (*models.InvoiceListResponse, error) {
	path := s.client.http.BuildPath("organizations", url.PathEscape(string(orgID)), "billing", "receipts")

	resp, err := s.client.http.Get(ctx, path, invoiceQuery(opts))
	if err != nil {
		return nil, err
	}

	var result models.InvoiceListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ListReceipts retrieves every payment receipt of an organization, following
// pagination.
func (s *OrganizationsService) ListReceipts(ctx context.Context, orgID models.OrganizationID, opts *models.ListInvoicesOptions) ([]models.Invoice, error) {
	return s.listBillingDocuments(ctx, opts, func(ctx context.Context, pageOpts *models.ListInvoicesOptions) (*models.InvoiceListResponse, error) {
		return s.ListReceiptsPage(ctx, orgID, pageOpts)
	})
}

// listBillingDocuments walks the pages of an invoice-shaped listing.
func (s *OrganizationsService) listBillingDocuments(
	ctx context.Context,
	opts *models.ListInvoicesOptions,
	fetch func(context.Context, *models.ListInvoicesOptions) (*models.InvoiceListResponse, error),
) ([]models.Invoice, error) {
	var all []models.Invoice
	page := 1

	for {
		pageOpts := models.ListInvoicesOptions{}
		if opts != nil {
			pageOpts = *opts
		}
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := fetch(ctx, &pageOpts)
		if err != nil {
			return nil, err
		}

		all = append(all, resp.Results...)

		if !resp.Pagination.HasNextPage {
			break
		}
		page++
	}

	return all, nil
}

// usageQuery turns usage filters into query parameters. Dates are sent as
// YYYY-MM-DD, which is what the endpoint expects. Both dates are required,
// so a missing one is reported before any request is sent.
func usageQuery(opts *models.UsageOptions, withGranularity bool) (url.Values, error) {
	if opts == nil || opts.StartDate == nil {
		return nil, &ValidationError{Field: "start_date", Message: "a start date is required"}
	}
	if opts.EndDate == nil {
		return nil, &ValidationError{Field: "end_date", Message: "an end date is required"}
	}

	query := url.Values{}
	query.Set("start_date", opts.StartDate.Format("2006-01-02"))
	query.Set("end_date", opts.EndDate.Format("2006-01-02"))
	if withGranularity && opts.Granularity != "" {
		query.Set("granularity", string(opts.Granularity))
	}
	return query, nil
}

// GetUsageSeries retrieves usage of a metered product over time, bucketed by
// the requested granularity. opts must set StartDate and EndDate.
func (s *OrganizationsService) GetUsageSeries(ctx context.Context, orgID models.OrganizationID, product models.UsageProduct, opts *models.UsageOptions) (*models.AIInferenceUsageSeries, error) {
	query, err := usageQuery(opts, true)
	if err != nil {
		return nil, err
	}

	path := s.client.http.BuildPath("organizations", url.PathEscape(string(orgID)), "usage", url.PathEscape(string(product)))

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.AIInferenceUsageSeries
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUsageSummary retrieves usage of a metered product totalled over a date
// range. opts must set StartDate and EndDate.
func (s *OrganizationsService) GetUsageSummary(ctx context.Context, orgID models.OrganizationID, product models.UsageProduct, opts *models.UsageOptions) (*models.AIInferenceUsageSummary, error) {
	query, err := usageQuery(opts, false)
	if err != nil {
		return nil, err
	}

	path := s.client.http.BuildPath("organizations", url.PathEscape(string(orgID)), "usage", url.PathEscape(string(product)), "summary")

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.AIInferenceUsageSummary
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
