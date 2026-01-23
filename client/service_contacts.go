package client

import (
	"context"
	"net/url"
	"strconv"

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
		pageOpts := opts
		if pageOpts == nil {
			pageOpts = &models.ListContactsOptions{}
		}
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

// UpdateContact updates an existing contact.
func (s *ContactsService) UpdateContact(ctx context.Context, contactID models.ContactID, req *models.ContactUpdateRequest) (*models.Contact, error) {
	path := s.client.http.BuildPath("contacts", string(contactID))

	resp, err := s.client.http.Patch(ctx, path, req)
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
	path := s.client.http.BuildPath("contacts", "verify")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}
