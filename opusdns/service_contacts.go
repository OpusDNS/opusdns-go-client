package opusdns

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/opusdns/opusdns-go-client/models"
)

// ContactsService provides methods for managing contacts.
type ContactsService struct {
	client *Client
}

// ListContacts retrieves all contacts with automatic pagination.
func (s *ContactsService) ListContacts(ctx context.Context, opts *models.ListContactsOptions) ([]models.Contact, error) {
	var allContacts []models.Contact
	page := 1

	for {
		pageOpts := cloneOptions(opts)
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := s.ListContactsPage(ctx, pageOpts)
		if err != nil {
			return nil, err
		}

		allContacts = append(allContacts, resp.Results...)

		if !resp.Pagination.HasNextPage {
			break
		}
		page++
	}

	return allContacts, nil
}

// ListContactsPage retrieves a single page of contacts.
func (s *ContactsService) ListContactsPage(ctx context.Context, opts *models.ListContactsOptions) (*models.ContactListResponse, error) {
	path := s.client.http.BuildPath("contacts")

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
		for _, tagID := range opts.TagIDs {
			query.Add("tag_ids", string(tagID))
		}
		if opts.TagMode != "" {
			query.Set("tag_mode", string(opts.TagMode))
		}
		if opts.Search != "" {
			query.Set("search", opts.Search)
		}
		if opts.FirstName != "" {
			query.Set("first_name", opts.FirstName)
		}
		if opts.LastName != "" {
			query.Set("last_name", opts.LastName)
		}
		if opts.Email != "" {
			query.Set("email", opts.Email)
		}
		if opts.Country != "" {
			query.Set("country", opts.Country)
		}
		if opts.Verified != nil {
			query.Set("verified", strconv.FormatBool(*opts.Verified))
		}
		if opts.CreatedAfter != nil {
			query.Set("created_after", opts.CreatedAfter.Format(time.RFC3339))
		}
		if opts.CreatedBefore != nil {
			query.Set("created_before", opts.CreatedBefore.Format(time.RFC3339))
		}
		for _, include := range opts.Include {
			query.Add("include", string(include))
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.ContactListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetContact retrieves a specific contact by ID.
func (s *ContactsService) GetContact(ctx context.Context, contactID models.ContactID) (*models.Contact, error) {
	path := s.client.http.BuildPath("contacts", string(contactID))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var contact models.Contact
	if err := s.client.http.DecodeResponse(resp, &contact); err != nil {
		return nil, err
	}

	return &contact, nil
}

// CreateContact creates a new contact.
func (s *ContactsService) CreateContact(ctx context.Context, req *models.ContactCreateRequest) (*models.Contact, error) {
	path := s.client.http.BuildPath("contacts")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var contact models.Contact
	if err := s.client.http.DecodeResponse(resp, &contact); err != nil {
		return nil, err
	}

	return &contact, nil
}

// DeleteContact deletes a contact.
func (s *ContactsService) DeleteContact(ctx context.Context, contactID models.ContactID) error {
	path := s.client.http.BuildPath("contacts", string(contactID))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// RequestVerification initiates email verification for a contact.
func (s *ContactsService) RequestVerification(ctx context.Context, contactID models.ContactID) (*models.ContactVerification, error) {
	path := s.client.http.BuildPath("contacts", string(contactID), "verification")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var verification models.ContactVerification
	if err := s.client.http.DecodeResponse(resp, &verification); err != nil {
		return nil, err
	}

	return &verification, nil
}

// GetVerificationStatus retrieves the verification status for a contact.
func (s *ContactsService) GetVerificationStatus(ctx context.Context, contactID models.ContactID) (*models.ContactVerification, error) {
	path := s.client.http.BuildPath("contacts", string(contactID), "verification")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var verification models.ContactVerification
	if err := s.client.http.DecodeResponse(resp, &verification); err != nil {
		return nil, err
	}

	return &verification, nil
}

// VerifyContact verifies a contact with the provided token.
func (s *ContactsService) VerifyContact(ctx context.Context, req *models.ContactVerificationRequest) error {
	if req == nil || req.Token == "" {
		return &ValidationError{Field: "token", Message: "verification token is required"}
	}

	path := s.client.http.BuildPath("contacts", "verify")

	query := url.Values{}
	query.Set("token", req.Token)

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// ListContactAttributeSets retrieves all contact attribute sets with automatic pagination.
func (s *ContactsService) ListContactAttributeSets(ctx context.Context, opts *models.ListContactAttributeSetsOptions) ([]models.ContactAttributeSet, error) {
	var all []models.ContactAttributeSet
	page := 1

	for {
		pageOpts := models.ListContactAttributeSetsOptions{PageSize: DefaultPageSize}
		if opts != nil {
			pageOpts = *opts
		}
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := s.ListContactAttributeSetsPage(ctx, &pageOpts)
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

// ListContactAttributeSetsPage retrieves a single page of contact attribute sets.
func (s *ContactsService) ListContactAttributeSetsPage(ctx context.Context, opts *models.ListContactAttributeSetsOptions) (*models.ContactAttributeSetListResponse, error) {
	path := s.client.http.BuildPath("contacts", "attribute-sets")

	query := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(opts.PageSize))
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.ContactAttributeSetListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetContactAttributeSet retrieves a contact attribute set by ID.
func (s *ContactsService) GetContactAttributeSet(ctx context.Context, setID models.ContactAttributeSetID) (*models.ContactAttributeSet, error) {
	path := s.client.http.BuildPath("contacts", "attribute-sets", string(setID))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var set models.ContactAttributeSet
	if err := s.client.http.DecodeResponse(resp, &set); err != nil {
		return nil, err
	}

	return &set, nil
}

// CreateContactAttributeSet creates a contact attribute set.
func (s *ContactsService) CreateContactAttributeSet(ctx context.Context, req *models.ContactAttributeSetCreateRequest) (*models.ContactAttributeSet, error) {
	path := s.client.http.BuildPath("contacts", "attribute-sets")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var set models.ContactAttributeSet
	if err := s.client.http.DecodeResponse(resp, &set); err != nil {
		return nil, err
	}

	return &set, nil
}

// UpdateContactAttributeSet updates a contact attribute set.
func (s *ContactsService) UpdateContactAttributeSet(ctx context.Context, setID models.ContactAttributeSetID, req *models.ContactAttributeSetUpdateRequest) (*models.ContactAttributeSet, error) {
	path := s.client.http.BuildPath("contacts", "attribute-sets", string(setID))

	resp, err := s.client.http.Patch(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var set models.ContactAttributeSet
	if err := s.client.http.DecodeResponse(resp, &set); err != nil {
		return nil, err
	}

	return &set, nil
}

// DeleteContactAttributeSet deletes a contact attribute set.
func (s *ContactsService) DeleteContactAttributeSet(ctx context.Context, setID models.ContactAttributeSetID) error {
	path := s.client.http.BuildPath("contacts", "attribute-sets", string(setID))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// LinkContactAttributeSet links a contact to a contact attribute set.
func (s *ContactsService) LinkContactAttributeSet(ctx context.Context, contactID models.ContactID, setID models.ContactAttributeSetID) (*models.ContactAttributeLink, error) {
	path := s.client.http.BuildPath("contacts", string(contactID), "link", string(setID))

	resp, err := s.client.http.Patch(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var link models.ContactAttributeLink
	if err := s.client.http.DecodeResponse(resp, &link); err != nil {
		return nil, err
	}

	return &link, nil
}

// AttestContactVerification submits one or more contact-verification attestations and
// returns the per-claim verification state.
func (s *ContactsService) AttestContactVerification(ctx context.Context, contactID models.ContactID, req *models.ContactAttestRequest) (*models.ContactAttestResponse, error) {
	path := s.client.http.BuildPath("contacts", string(contactID), "verifications", "attest")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var result models.ContactAttestResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetContactVerifications retrieves the current verification state for a contact.
func (s *ContactsService) GetContactVerifications(ctx context.Context, contactID models.ContactID) (*models.ContactAttestResponse, error) {
	path := s.client.http.BuildPath("contacts", string(contactID), "verifications")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.ContactAttestResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CancelContactVerification deletes a contact's verification.
func (s *ContactsService) CancelContactVerification(ctx context.Context, contactID models.ContactID) error {
	path := s.client.http.BuildPath("contacts", string(contactID), "verification")

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}
