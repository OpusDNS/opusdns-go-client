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
