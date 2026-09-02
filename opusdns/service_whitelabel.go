package opusdns

import (
	"context"

	"github.com/opusdns/opusdns-go-client/models"
)

// WhitelabelService manages an organization's whitelabel branding: serving the
// dashboard and its transactional email under the reseller's own brand.
//
// An organization has at most one configuration, so every call here acts on
// that single configuration and none of them takes an identifier.
type WhitelabelService struct {
	client *Client
}

// Get retrieves the organization's whitelabel configuration.
func (s *WhitelabelService) Get(ctx context.Context) (*models.Whitelabel, error) {
	path := s.client.http.BuildPath("whitelabel-branding")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.Whitelabel
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Create provisions a whitelabel configuration. Pass a
// models.WhitelabelBaseCreateRequest to be served on a subdomain of an
// OpusDNS-owned zone, or a models.WhitelabelPlusCreateRequest to be served on
// your own domain. Onboarding continues asynchronously; poll Get to follow it.
func (s *WhitelabelService) Create(ctx context.Context, req interface{}) (*models.ProductCreateResponse, error) {
	path := s.client.http.BuildPath("whitelabel-branding")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var result models.ProductCreateResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Update changes the label, the enabled flag or the renewal intent of the
// configuration.
func (s *WhitelabelService) Update(ctx context.Context, req *models.WhitelabelUpdateRequest) (*models.Whitelabel, error) {
	path := s.client.http.BuildPath("whitelabel-branding")

	resp, err := s.client.http.Patch(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var result models.Whitelabel
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpgradeToPlus moves a base-tier configuration onto the customer's own domain.
// The upgrade is asynchronous; poll Get to follow onboarding.
func (s *WhitelabelService) UpgradeToPlus(ctx context.Context, req *models.WhitelabelUpgradeRequest) (*models.ProductCreateResponse, error) {
	path := s.client.http.BuildPath("whitelabel-branding", "tier")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var result models.ProductCreateResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Recheck re-runs verification for a configuration whose onboarding failed,
// optionally correcting the hostname it was pointed at. Pass nil to retry
// unchanged.
func (s *WhitelabelService) Recheck(ctx context.Context, req *models.WhitelabelRecheckRequest) (*models.Whitelabel, error) {
	path := s.client.http.BuildPath("whitelabel-branding", "recheck")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var result models.Whitelabel
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Restore brings back a configuration that was force-disabled.
func (s *WhitelabelService) Restore(ctx context.Context) (*models.Whitelabel, error) {
	path := s.client.http.BuildPath("whitelabel-branding", "restore")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.Whitelabel
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ListEmailTemplates retrieves the catalogue of editable transactional email
// templates, keyed by template name.
func (s *WhitelabelService) ListEmailTemplates(ctx context.Context) (map[string]models.MailTemplate, error) {
	path := s.client.http.BuildPath("whitelabel-branding", "email", "templates")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	result := map[string]models.MailTemplate{}
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// PreviewEmail renders one template so branding changes can be checked before
// they are saved.
func (s *WhitelabelService) PreviewEmail(ctx context.Context, req *models.PreviewMailRequest) (*models.PreviewMailResponse, error) {
	path := s.client.http.BuildPath("whitelabel-branding", "email", "preview")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var result models.PreviewMailResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
