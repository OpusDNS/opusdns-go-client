// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// UserID is a TypeID for users.
type UserID = TypeID

// UserAttributeBase represents a single user attribute key/value pair.
type UserAttributeBase struct {
	// Key is the attribute key.
	Key string `json:"key"`

	// Value is the attribute value.
	Value interface{} `json:"value,omitempty"`
}

// UserSortField represents fields that can be used for sorting users.
type UserSortField string

const (
	UserSortByCreatedOn UserSortField = "created_on"
	UserSortByUsername  UserSortField = "username"
	UserSortByEmail     UserSortField = "email"
)

// User represents a user in the OpusDNS system.
type User struct {
	// UserID is the unique identifier for the user.
	UserID UserID `json:"user_id,omitempty"`

	// Username is the user's unique username.
	Username string `json:"username"`

	// FirstName is the user's first name.
	FirstName string `json:"first_name"`

	// LastName is the user's last name.
	LastName string `json:"last_name"`

	// Email is the user's email address.
	Email string `json:"email"`

	// Phone is the user's phone number.
	Phone *string `json:"phone,omitempty"`

	// Locale is the user's locale.
	Locale string `json:"locale"`

	// OrganizationID is the ID of the user's organization.
	OrganizationID OrganizationID `json:"organization_id,omitempty"`

	// CreatedOn is when the user was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the user was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`

	// UserAttributes contains requested user attributes.
	UserAttributes map[string]interface{} `json:"user_attributes,omitempty"`
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Username
	}
	if u.FirstName == "" {
		return u.LastName
	}
	if u.LastName == "" {
		return u.FirstName
	}
	return u.FirstName + " " + u.LastName
}

// UserWithRole is a user as the organization's user list returns it: the
// user plus the role assigned to them (UserPublicWithRole).
type UserWithRole struct {
	User

	// Role is the user's role: a built-in role name (see AssignableRole, plus
	// "owner") or the label of a custom role. Nil when no role is assigned.
	Role *string `json:"role,omitempty"`
}

// UserProfile is the authenticated user as GetCurrentUser returns it: the
// user plus their role, resolved permissions and organization (UserProfile).
type UserProfile struct {
	User

	// Role is the user's role: a built-in role name (see AssignableRole, plus
	// "owner") or the label of a custom role. Nil when no role is assigned.
	Role *string `json:"role,omitempty"`

	// Permissions are the "resource:scope" permissions the role resolves to.
	Permissions []Permission `json:"permissions,omitempty"`

	// Organization carries details of the user's organization.
	Organization *UserOrganizationMeta `json:"organization,omitempty"`
}

// UserOrganizationMeta carries details of a user's organization.
type UserOrganizationMeta struct {
	// Currency is the organization's billing currency.
	Currency *Currency `json:"currency,omitempty"`

	// ParentOrganizationID is the organization's parent, if it has one.
	ParentOrganizationID *OrganizationID `json:"parent_organization_id,omitempty"`
}

// UserListResponse represents the paginated response when listing users.
type UserListResponse struct {
	// Results contains the list of users for the current page.
	Results []UserWithRole `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// UserCreateRequest represents a request to create a new user.
type UserCreateRequest struct {
	// Username is the user's unique username.
	Username string `json:"username"`

	// FirstName is the user's first name.
	FirstName string `json:"first_name"`

	// LastName is the user's last name.
	LastName string `json:"last_name"`

	// Email is the user's email address.
	Email string `json:"email"`

	// Phone is the user's phone number.
	Phone *string `json:"phone,omitempty"`

	// Locale is the user's locale.
	Locale string `json:"locale"`

	// Password is the user's plaintext password.
	// When omitted, the API can provision the user without an initial password.
	Password *string `json:"password,omitempty"`

	// UserAttributes contains optional custom user attributes.
	UserAttributes []UserAttributeBase `json:"user_attributes,omitempty"`
}

// UserUpdateRequest represents a request to update an existing user.
type UserUpdateRequest struct {
	// Username is the user's unique username.
	Username *string `json:"username,omitempty"`

	// FirstName is the user's first name.
	FirstName *string `json:"first_name,omitempty"`

	// LastName is the user's last name.
	LastName *string `json:"last_name,omitempty"`

	// Email is the user's email address.
	Email *string `json:"email,omitempty"`

	// Phone is the user's phone number.
	Phone *string `json:"phone,omitempty"`

	// Locale is the user's locale.
	Locale *string `json:"locale,omitempty"`

	// UserAttributes contains optional custom user attributes.
	UserAttributes []UserAttributeBase `json:"user_attributes,omitempty"`
}

// Permission is a user permission identifier, in "resource:scope" form
// (e.g. "domains:read").
type Permission string

// PermissionSet represents the resolved permissions assigned to a user.
type PermissionSet struct {
	Permissions []Permission `json:"permissions"`
}

// ListUsersOptions contains options for listing users.
type ListUsersOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of users per page.
	PageSize int

	// SortBy is the field to sort by.
	SortBy UserSortField

	// SortOrder is the sort direction.
	SortOrder SortOrder

	// Search is an optional search query to filter users.
	Search string
}
