// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// AssignableRole is a built-in role that can be assigned to a user or API key.
// The report-only "owner" role is intentionally absent — it can never be assigned.
type AssignableRole string

const (
	// RoleAdmin grants full access to the organization.
	RoleAdmin AssignableRole = "admin"

	// RoleViewer grants read-only access.
	RoleViewer AssignableRole = "viewer"

	// RoleDomainManager grants management of domains and related resources.
	RoleDomainManager AssignableRole = "domain_manager"

	// RoleDNSManager grants management of DNS zones and forwarding.
	RoleDNSManager AssignableRole = "dns_manager"

	// RoleBillingManager grants management of billing.
	RoleBillingManager AssignableRole = "billing_manager"
)

// PermissionResource is a resource area a permission applies to.
// Permission strings have the form "<resource>:<scope>" (e.g. "domains:read").
type PermissionResource string

const (
	PermissionResourceOrganization        PermissionResource = "organization"
	PermissionResourceDomains             PermissionResource = "domains"
	PermissionResourceContacts            PermissionResource = "contacts"
	PermissionResourceDNS                 PermissionResource = "dns"
	PermissionResourceHosts               PermissionResource = "hosts"
	PermissionResourceEmailForwards       PermissionResource = "email_forwards"
	PermissionResourceDomainForwards      PermissionResource = "domain_forwards"
	PermissionResourceParking             PermissionResource = "parking"
	PermissionResourceEvents              PermissionResource = "events"
	PermissionResourceJobs                PermissionResource = "jobs"
	PermissionResourceBilling             PermissionResource = "billing"
	PermissionResourceUsers               PermissionResource = "users"
	PermissionResourceAPIKeys             PermissionResource = "api_keys"
	PermissionResourceRegistrarCredential PermissionResource = "registrar_credentials"
	PermissionResourceTags                PermissionResource = "tags"
	PermissionResourceAuditLogs           PermissionResource = "audit_logs"
	PermissionResourceVanityNS            PermissionResource = "vanity_ns"
	PermissionResourceAIConcierge         PermissionResource = "ai_concierge"
)

// PermissionScope is the scope (level of access) a permission grants.
type PermissionScope string

const (
	// PermissionScopeRead grants read-only access.
	PermissionScopeRead PermissionScope = "read"

	// PermissionScopeManage grants create/update access.
	PermissionScopeManage PermissionScope = "manage"

	// PermissionScopeDelete grants delete access.
	PermissionScopeDelete PermissionScope = "delete"
)

// RoleDefinition is a role as listed or read through the public API — built-in or custom.
type RoleDefinition struct {
	// Label is the per-organization unique, URL-safe identifier (snake_case, e.g.
	// "support_staff"). It is used as the path parameter when getting, updating or
	// deleting a role.
	Label string `json:"label"`

	// Name is the human-readable display name (e.g. "Support Staff").
	Name string `json:"name"`

	// Description is an optional description of the role.
	Description *string `json:"description,omitempty"`

	// BuiltIn indicates whether this is a built-in role. Built-in roles are
	// immutable; custom roles are organization-owned.
	BuiltIn bool `json:"built_in"`

	// Permissions is the list of "resource:scope" permission strings the role grants
	// (e.g. "domains:read", "dns:manage").
	Permissions []string `json:"permissions"`

	// CreatedOn is when the role was created (custom roles only).
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the role was last updated (custom roles only).
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// CustomRoleCreateRequest is the request body for creating a custom role.
type CustomRoleCreateRequest struct {
	// Name is the display name of the custom role (1-64 characters).
	Name string `json:"name"`

	// Description is an optional description of the role.
	Description *string `json:"description,omitempty"`

	// Permissions is the list of "resource:scope" permission strings the role grants.
	// The escalation-bearing admin/owner permissions cannot be granted.
	Permissions []string `json:"permissions"`
}

// CustomRoleUpdateRequest is the request body for updating a custom role.
// Omitted fields are left unchanged; Permissions is a full replacement set when provided.
type CustomRoleUpdateRequest struct {
	// Name is the new display name (optional).
	Name *string `json:"name,omitempty"`

	// Description is the new description (optional).
	Description *string `json:"description,omitempty"`

	// Permissions is the full replacement set of "resource:scope" permissions (optional).
	Permissions *[]string `json:"permissions,omitempty"`
}

// PermissionCatalogResponse represents the catalog of "resource:scope" permissions
// a custom role may grant.
type PermissionCatalogResponse struct {
	// Permissions is the list of grantable "resource:scope" permission strings.
	Permissions []string `json:"permissions"`
}

// RoleAssignment represents the role assigned to a user.
type RoleAssignment struct {
	// Role is the assigned role: a built-in role name (which may be the report-only
	// "owner"), the label of a custom role, or nil when no role is assigned.
	Role *string `json:"role"`
}

// RoleAssignmentRequest is the request body for setting a user's role.
type RoleAssignmentRequest struct {
	// Role is the role to assign: a built-in assignable role name or the label of a
	// custom role owned by the user's organization. A nil value clears the role.
	Role *string `json:"role"`
}
