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

// GetOrganization retrieves an organization by ID.
func (s *OrganizationsService) GetOrganization(ctx context.Context, orgID models.OrganizationID) (*models.Organization, error) {
	path := s.client.http.BuildPath("organizations", string(orgID))

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

// UpdateOrganization updates an organization.
func (s *OrganizationsService) UpdateOrganization(ctx context.Context, orgID models.OrganizationID, req *models.OrganizationUpdateRequest) (*models.Organization, error) {
	path := s.client.http.BuildPath("organizations", string(orgID))

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

// ListIPRestrictions retrieves IP restrictions for the organization.
func (s *OrganizationsService) ListIPRestrictions(ctx context.Context) (*models.IPRestrictionListResponse, error) {
	path := s.client.http.BuildPath("organizations", "ip-restrictions")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.IPRestrictionListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetIPRestriction retrieves a specific IP restriction by ID.
func (s *OrganizationsService) GetIPRestriction(ctx context.Context, restrictionID models.TypeID) (*models.IPRestriction, error) {
	path := s.client.http.BuildPath("organizations", "ip-restrictions", string(restrictionID))

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
func (s *OrganizationsService) UpdateIPRestriction(ctx context.Context, restrictionID models.TypeID, req *models.IPRestrictionUpdateRequest) (*models.IPRestriction, error) {
	path := s.client.http.BuildPath("organizations", "ip-restrictions", string(restrictionID))

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
func (s *OrganizationsService) DeleteIPRestriction(ctx context.Context, restrictionID models.TypeID) error {
	path := s.client.http.BuildPath("organizations", "ip-restrictions", string(restrictionID))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// ListRoles retrieves available roles.
func (s *OrganizationsService) ListRoles(ctx context.Context) (*models.RoleListResponse, error) {
	path := s.client.http.BuildPath("organizations", "roles")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.RoleListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetAttributes retrieves organization attributes.
func (s *OrganizationsService) GetAttributes(ctx context.Context, orgID models.OrganizationID) (*models.OrganizationAttributesResponse, error) {
	path := s.client.http.BuildPath("organizations", "attributes", string(orgID))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.OrganizationAttributesResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateAttributes updates organization attributes.
func (s *OrganizationsService) UpdateAttributes(ctx context.Context, orgID models.OrganizationID, req *models.OrganizationAttributeUpdateRequest) (*models.OrganizationAttributesResponse, error) {
	path := s.client.http.BuildPath("organizations", "attributes", string(orgID))

	resp, err := s.client.http.Patch(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var result models.OrganizationAttributesResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ListTransactions retrieves billing transactions for an organization.
func (s *OrganizationsService) ListTransactions(ctx context.Context, orgID models.OrganizationID, opts *models.ListTransactionsOptions) (*models.BillingTransactionListResponse, error) {
	path := s.client.http.BuildPath("organizations", string(orgID), "transactions")

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
			query.Set("created_after", opts.CreatedAfter.Format(time.RFC3339))
		}
		if opts.CreatedBefore != nil {
			query.Set("created_before", opts.CreatedBefore.Format(time.RFC3339))
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
	path := s.client.http.BuildPath("organizations", string(orgID), "transactions", string(transactionID))

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

// ListInvoices retrieves invoices for an organization.
func (s *OrganizationsService) ListInvoices(ctx context.Context, orgID models.OrganizationID) (*models.InvoiceListResponse, error) {
	path := s.client.http.BuildPath("organizations", string(orgID), "billing", "invoices")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.InvoiceListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetPricing retrieves pricing for a specific product type.
func (s *OrganizationsService) GetPricing(ctx context.Context, orgID models.OrganizationID, productType string) (*models.ProductPricing, error) {
	path := s.client.http.BuildPath("organizations", string(orgID), "pricing", "product-type", url.PathEscape(productType))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var pricing models.ProductPricing
	if err := s.client.http.DecodeResponse(resp, &pricing); err != nil {
		return nil, err
	}

	return &pricing, nil
}
