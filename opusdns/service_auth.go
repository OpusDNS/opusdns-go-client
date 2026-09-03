package opusdns

import (
	"context"

	"github.com/opusdns/opusdns-go-client/models"
)

// AuthService provides methods for authentication-related operations.
type AuthService struct {
	client *Client
}

// IntrospectAPIKey returns the stored record for the API key (or organization token)
// used to authenticate the request, including the role bound to it.
func (s *AuthService) IntrospectAPIKey(ctx context.Context) (*models.OrganizationCredential, error) {
	path := s.client.http.BuildPath("auth", "client_credentials", "introspect")

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var credential models.OrganizationCredential
	if err := s.client.http.DecodeResponse(resp, &credential); err != nil {
		return nil, err
	}

	return &credential, nil
}

// IssueToken exchanges organization client credentials for a short lived access
// token, for callers that would rather hand a bearer token than an API key to
// whatever consumes it. The exchange itself is authenticated like every other
// request, with the client's configured API key.
func (s *AuthService) IssueToken(ctx context.Context, clientID models.OrganizationID, clientSecret string) (*models.TokenResponse, error) {
	path := s.client.http.BuildPath("auth", "token")

	resp, err := s.client.http.Post(ctx, path, &models.TokenRequest{
		GrantType:    models.GrantTypeClientCredentials,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
	if err != nil {
		return nil, err
	}

	var result models.TokenResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
