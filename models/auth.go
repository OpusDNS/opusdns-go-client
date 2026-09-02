// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// OrganizationCredentialID is a TypeID for API keys (organization credentials).
type OrganizationCredentialID = TypeID

// OrganizationCredentialStatus represents the status of an API key.
type OrganizationCredentialStatus string

const (
	// OrganizationCredentialStatusActive indicates the API key is active.
	OrganizationCredentialStatusActive OrganizationCredentialStatus = "active"

	// OrganizationCredentialStatusExpired indicates the API key has expired.
	OrganizationCredentialStatusExpired OrganizationCredentialStatus = "expired"

	// OrganizationCredentialStatusRevoked indicates the API key has been revoked.
	OrganizationCredentialStatusRevoked OrganizationCredentialStatus = "revoked"
)

// OrganizationCredential represents an API key (organization credential), including
// the role it is bound to.
type OrganizationCredential struct {
	// APIKeyID is the unique identifier of the API key.
	APIKeyID OrganizationCredentialID `json:"api_key_id"`

	// OrganizationID is the organization the API key belongs to.
	OrganizationID OrganizationID `json:"organization_id"`

	// APIKeyName is the optional name of the API key.
	APIKeyName *string `json:"api_key_name,omitempty"`

	// APIKeyDescription is the optional description of the API key.
	APIKeyDescription *string `json:"api_key_description,omitempty"`

	// ExpiresAt is when the API key expires, if set.
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// Role is the role bound to the API key: a built-in role name, the label of a
	// custom role, or nil when no role is assigned.
	Role *string `json:"role,omitempty"`

	// Status is the current status of the API key (active, expired or revoked).
	Status OrganizationCredentialStatus `json:"status,omitempty"`

	// CreatedOn is when the API key was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// LastUsedOn is when the API key was last used.
	LastUsedOn *time.Time `json:"last_used_on,omitempty"`
}

// GrantTypeClientCredentials is the only grant type the token endpoint accepts.
const GrantTypeClientCredentials = "client_credentials"

// TokenRequest exchanges organization client credentials for an access token.
type TokenRequest struct {
	// GrantType is always GrantTypeClientCredentials.
	GrantType string `json:"grant_type"`

	// ClientID is the organization ID the credentials belong to.
	ClientID OrganizationID `json:"client_id"`

	// ClientSecret is the client secret.
	ClientSecret string `json:"client_secret"`
}

// TokenResponse is a freshly issued organization access token, as returned by
// the client-credentials grant. The user-credentials grant, which is the only
// one that also returns a refresh token, is not part of this client.
type TokenResponse struct {
	// AccessToken is the bearer token to authenticate with.
	AccessToken string `json:"access_token"`

	// TokenType is the token type, normally "Bearer".
	TokenType string `json:"token_type,omitempty"`

	// ExpiresIn is the access token lifetime in seconds.
	ExpiresIn int `json:"expires_in"`
}
