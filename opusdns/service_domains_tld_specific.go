package opusdns

import (
	"context"
	"net/url"

	"github.com/opusdns/opusdns-go-client/models"
)

// TLD-specific domain operations. These are registry actions that exist only
// for certain TLDs and have no equivalent on the generic domain endpoints.
// Everything here lives on DomainsService, under /v1/domains/tld-specific/.

// RequestAuthCode asks the registry to issue a transfer auth code for a domain.
// Registries behind this endpoint never return the code over the API; they
// deliver it out of band, normally by email to the registrant. Use one of the
// AuthCodeTLD constants for tld.
//
// The TLD is a path segment here, so this one method serves the nine per-TLD
// routes the specification lists separately, and it opts out of the route check
// that pairs a method with a single path.
//
//speccheck:ignore
func (s *DomainsService) RequestAuthCode(ctx context.Context, tld models.AuthCodeTLD, domainRef string) (*models.RequestAuthCodeResponse, error) {
	path := s.client.http.BuildPath("domains", "tld-specific", url.PathEscape(string(tld)),
		url.PathEscape(domainRef), "auth_code", "request")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.RequestAuthCodeResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// WithdrawATDomain withdraws an .at domain, handing it back to the registry
// rather than deleting it.
func (s *DomainsService) WithdrawATDomain(ctx context.Context, domainRef string) (*models.DomainWithdrawResponse, error) {
	path := s.client.http.BuildPath("domains", "tld-specific", "at", url.PathEscape(domainRef), "withdraw")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.DomainWithdrawResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// TransitDEDomain puts a .de domain into transit, handing responsibility for it
// back to DENIC.
func (s *DomainsService) TransitDEDomain(ctx context.Context, domainRef string) (*models.DomainTransitResponse, error) {
	path := s.client.http.BuildPath("domains", "tld-specific", "de", url.PathEscape(domainRef), "transit")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.DomainTransitResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SubmitNorIDDeclaration signs the .no applicant declaration on the applicant's
// behalf. A .no registration stays pending until the declaration is signed.
func (s *DomainsService) SubmitNorIDDeclaration(ctx context.Context, domainRef string, req *models.NorIDDeclarationRequest) (*models.NorIDDeclarationResponse, error) {
	path := s.client.http.BuildPath("domains", "tld-specific", "no", url.PathEscape(domainRef), "applicant-declaration")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var result models.NorIDDeclarationResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ResendNorIDDeclarationEmail sends the .no applicant declaration email again,
// for a registration still waiting on the applicant's signature.
func (s *DomainsService) ResendNorIDDeclarationEmail(ctx context.Context, domainRef string) error {
	path := s.client.http.BuildPath("domains", "tld-specific", "no", url.PathEscape(domainRef), "resend-declaration-email")

	resp, err := s.client.http.Post(ctx, path, nil)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}
