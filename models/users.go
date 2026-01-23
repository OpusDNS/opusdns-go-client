// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// UserID is a TypeID for users.
type UserID = TypeID

// User represents a user in the OpusDNS system.
type User struct {
	// UserID is the unique identifier for the user.
	UserID UserID `json:"user_id"`

	// Email is the user's email address.
	Email string `json:"email"`

	// FirstName is the user's first name.
	FirstName string `json:"first_name,omitempty"`

	// LastName is the user's last name.
	LastName string `json:"last_name,omitempty"`

	// Phone is the user's phone number.
	Phone *string `json:"phone,omitempty"`

	// OrganizationID is the ID of the user's organization.
	OrganizationID OrganizationID `json:"organization_id,omitempty"`

	// Roles contains the user's roles.
	Roles []Role `json:"roles,omitempty"`

	// Permissions contains the user's direct permissions.
	Permissions []string `json:"permissions,omitempty"`

	// Active indicates if the user account is active.
	Active bool `json:"active,omitempty"`

	// Verified indicates if the user's email has been verified.
	Verified bool `json:"verified,omitempty"`

	// TwoFactorEnabled indicates if 2FA is enabled for the user.
	TwoFactorEnabled bool `json:"two_factor_enabled,omitempty"`

	// LastLoginAt is when the user last logged in.
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`

	// CreatedOn is when the user was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the user was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Email
	}
	if u.FirstName == "" {
		return u.LastName
	}
	if u.LastName == "" {
		return u.FirstName
	}
	return u.FirstName + " " + u.LastName
}

// UserListResponse represents the paginated response when listing users.
type UserListResponse struct {
	// Results contains the list of users for the current page.
	Results []User `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// UserCreateRequest represents a request to create a new user.
type UserCreateRequest struct {
	// Email is the user's email address.
	Email string `json:"email"`

	// FirstName is the user's first name.
	FirstName string `json:"first_name,omitempty"`

	// LastName is the user's last name.
	LastName string `json:"last_name,omitempty"`

	// Phone is the user's phone number.
	Phone *string `json:"phone,omitempty"`

	// Password is the user's initial password (optional, may trigger email invite).
	Password *string `json:"password,omitempty"`

	// RoleIDs is the list of role IDs to assign to the user.
	RoleIDs []TypeID `json:"role_ids,omitempty"`

	// SendInvite indicates whether to send an invitation email.
	SendInvite bool `json:"send_invite,omitempty"`
}

// UserUpdateRequest represents a request to update an existing user.
type UserUpdateRequest struct {
	// FirstName is the user's first name.
	FirstName *string `json:"first_name,omitempty"`

	// LastName is the user's last name.
	LastName *string `json:"last_name,omitempty"`

	// Phone is the user's phone number.
	Phone *string `json:"phone,omitempty"`

	// Active indicates if the user account should be active.
	Active *bool `json:"active,omitempty"`
}

// UserRolesUpdateRequest represents a request to update a user's roles.
type UserRolesUpdateRequest struct {
	// RoleIDs is the list of role IDs to assign to the user.
	RoleIDs []TypeID `json:"role_ids"`
}

// UserPermissionsResponse represents the response when getting user permissions.
type UserPermissionsResponse struct {
	// Permissions contains the list of permission strings.
	Permissions []string `json:"permissions"`

	// EffectivePermissions contains permissions including those from roles.
	EffectivePermissions []string `json:"effective_permissions,omitempty"`
}

// PasswordResetRequest represents a request to reset a user's password.
type PasswordResetRequest struct {
	// CurrentPassword is the user's current password (for authenticated resets).
	CurrentPassword *string `json:"current_password,omitempty"`

	// NewPassword is the new password to set.
	NewPassword string `json:"new_password"`

	// Token is the password reset token (for token-based resets).
	Token *string `json:"token,omitempty"`
}

// PasswordResetInitiateRequest represents a request to initiate a password reset.
type PasswordResetInitiateRequest struct {
	// Email is the user's email address.
	Email string `json:"email"`
}

// ListUsersOptions contains options for listing users.
type ListUsersOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of users per page.
	PageSize int

	// SortBy is the field to sort by.
	SortBy string

	// SortOrder is the sort direction.
	SortOrder SortOrder

	// Search is an optional search query to filter users.
	Search string

	// Email filters by email address.
	Email string

	// Active filters by active status.
	Active *bool

	// Verified filters by verified status.
	Verified *bool

	// RoleID filters by role ID.
	RoleID TypeID
}

// CurrentUser represents the currently authenticated user with additional context.
type CurrentUser struct {
	User

	// Organization contains the user's organization details.
	Organization *Organization `json:"organization,omitempty"`

	// EffectivePermissions contains all permissions the user has.
	EffectivePermissions []string `json:"effective_permissions,omitempty"`

	// APIKeyID is the ID of the API key used for authentication (if applicable).
	APIKeyID *TypeID `json:"api_key_id,omitempty"`
}

// AuthToken represents an authentication token.
type AuthToken struct {
	// AccessToken is the JWT access token.
	AccessToken string `json:"access_token"`

	// TokenType is the token type (usually "Bearer").
	TokenType string `json:"token_type"`

	// ExpiresIn is the token expiration time in seconds.
	ExpiresIn int `json:"expires_in,omitempty"`

	// RefreshToken is the refresh token (if applicable).
	RefreshToken *string `json:"refresh_token,omitempty"`

	// Scope is the token scope (if applicable).
	Scope *string `json:"scope,omitempty"`
}

// AuthTokenRequest represents a request to obtain an authentication token.
type AuthTokenRequest struct {
	// GrantType is the OAuth2 grant type.
	GrantType string `json:"grant_type"`

	// Username is the user's email (for password grant).
	Username *string `json:"username,omitempty"`

	// Password is the user's password (for password grant).
	Password *string `json:"password,omitempty"`

	// RefreshToken is the refresh token (for refresh_token grant).
	RefreshToken *string `json:"refresh_token,omitempty"`

	// Scope is the requested scope.
	Scope *string `json:"scope,omitempty"`
}
