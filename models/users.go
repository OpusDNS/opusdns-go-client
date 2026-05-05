// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// UserID is a TypeID for users.
type UserID = TypeID

// UserStatus represents the status of a user.
type UserStatus string

const (
	// UserStatusActive indicates the user is active.
	UserStatusActive UserStatus = "active"

	// UserStatusInactive indicates the user is inactive.
	UserStatusInactive UserStatus = "inactive"

	// UserStatusPending indicates the user is pending activation.
	UserStatusPending UserStatus = "pending"
)

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

	// Status is the user's status.
	Status UserStatus `json:"status"`

	// OrganizationID is the ID of the user's organization.
	OrganizationID OrganizationID `json:"organization_id,omitempty"`

	// KeycloakUserID is the Keycloak user id.
	KeycloakUserID *string `json:"keycloak_user_id,omitempty"`

	// CreatedOn is when the user was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the user was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`

	// DeletedOn is when the user was deleted.
	DeletedOn *time.Time `json:"deleted_on,omitempty"`

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

// UserListResponse represents the paginated response when listing users.
type UserListResponse struct {
	// Results contains the list of users for the current page.
	Results []User `json:"results"`

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
}

// UserUpdateRequest represents a request to update an existing user.
type UserUpdateRequest struct {
	// FirstName is the user's first name.
	FirstName *string `json:"first_name,omitempty"`

	// LastName is the user's last name.
	LastName *string `json:"last_name,omitempty"`

	// Phone is the user's phone number.
	Phone *string `json:"phone,omitempty"`

	// Locale is the user's locale.
	Locale *string `json:"locale,omitempty"`
}

// Permission is a user permission identifier.
type Permission string

// PermissionSet represents the permissions assigned to a user.
type PermissionSet struct {
	Permissions []Permission `json:"permissions"`
}

// Relation is a user relation/role identifier.
type Relation string

// RelationSet represents the roles assigned to a user.
type RelationSet struct {
	Relations []Relation `json:"relations"`
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

	// Email filters by email address.
	//
	// Deprecated: the current API only supports Search for organization user lists.
	Email string

	// Username filters by username.
	//
	// Deprecated: the current API only supports Search for organization user lists.
	Username string

	// Status filters by user status.
	//
	// Deprecated: the current API only supports Search for organization user lists.
	Status UserStatus
}
