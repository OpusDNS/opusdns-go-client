// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// EventID is a TypeID for events.
type EventID = TypeID

// EventType represents the type of an event.
type EventType string

const (
	// Account events
	EventTypeAccountCreate EventType = "ACCOUNT_CREATE"
	EventTypeAccountUpdate EventType = "ACCOUNT_UPDATE"
	EventTypeAccountDelete EventType = "ACCOUNT_DELETE"

	// Domain events
	EventTypeDomainCreate       EventType = "DOMAIN_CREATE"
	EventTypeDomainUpdate       EventType = "DOMAIN_UPDATE"
	EventTypeDomainDelete       EventType = "DOMAIN_DELETE"
	EventTypeDomainTransfer     EventType = "DOMAIN_TRANSFER"
	EventTypeDomainRenew        EventType = "DOMAIN_RENEW"
	EventTypeDomainRestore      EventType = "DOMAIN_RESTORE"
	EventTypeDomainModification EventType = "DOMAIN_MODIFICATION"
	EventTypeDomainExpiration   EventType = "DOMAIN_EXPIRATION"

	// Contact events
	EventTypeContactCreate EventType = "CONTACT_CREATE"
	EventTypeContactUpdate EventType = "CONTACT_UPDATE"
	EventTypeContactDelete EventType = "CONTACT_DELETE"
	EventTypeContactVerify EventType = "CONTACT_VERIFY"

	// DNS events
	EventTypeZoneCreate    EventType = "ZONE_CREATE"
	EventTypeZoneUpdate    EventType = "ZONE_UPDATE"
	EventTypeZoneDelete    EventType = "ZONE_DELETE"
	EventTypeDNSSECEnable  EventType = "DNSSEC_ENABLE"
	EventTypeDNSSECDisable EventType = "DNSSEC_DISABLE"

	// Host events
	EventTypeHostCreate EventType = "HOST_CREATE"
	EventTypeHostUpdate EventType = "HOST_UPDATE"
	EventTypeHostDelete EventType = "HOST_DELETE"

	// Notification events
	EventTypeNotification EventType = "NOTIFICATION"
)

// EventSubtype represents a more specific event subtype.
type EventSubtype string

const (
	// Domain modification subtypes
	EventSubtypeNameserverChange EventSubtype = "NAMESERVER_CHANGE"
	EventSubtypeContactChange    EventSubtype = "CONTACT_CHANGE"
	EventSubtypeStatusChange     EventSubtype = "STATUS_CHANGE"
	EventSubtypeAuthCodeChange   EventSubtype = "AUTH_CODE_CHANGE"

	// Transfer subtypes
	EventSubtypeTransferRequest   EventSubtype = "TRANSFER_REQUEST"
	EventSubtypeTransferApproved  EventSubtype = "TRANSFER_APPROVED"
	EventSubtypeTransferRejected  EventSubtype = "TRANSFER_REJECTED"
	EventSubtypeTransferCanceled  EventSubtype = "TRANSFER_CANCELED"
	EventSubtypeTransferCompleted EventSubtype = "TRANSFER_COMPLETED"

	// Notification subtypes
	EventSubtypeExpirationWarning EventSubtype = "EXPIRATION_WARNING"
	EventSubtypeRegistryMessage   EventSubtype = "REGISTRY_MESSAGE"
	EventSubtypeSystemAlert       EventSubtype = "SYSTEM_ALERT"
)

// EventObjectType represents the type of object an event relates to.
type EventObjectType string

const (
	EventObjectTypeDomain       EventObjectType = "DOMAIN"
	EventObjectTypeContact      EventObjectType = "CONTACT"
	EventObjectTypeZone         EventObjectType = "ZONE"
	EventObjectTypeHost         EventObjectType = "HOST"
	EventObjectTypeUser         EventObjectType = "USER"
	EventObjectTypeOrganization EventObjectType = "ORGANIZATION"
	EventObjectTypeTransaction  EventObjectType = "TRANSACTION"
	EventObjectTypeRaw          EventObjectType = "RAW"
)

// Event represents an event in the OpusDNS system.
type Event struct {
	// EventID is the unique identifier for the event.
	EventID EventID `json:"event_id"`

	// Type is the event type.
	Type EventType `json:"type,omitempty"`

	// Subtype is the event subtype for more specific categorization.
	Subtype *EventSubtype `json:"subtype,omitempty"`

	// ObjectType is the type of object the event relates to.
	ObjectType EventObjectType `json:"object_type,omitempty"`

	// ObjectID is the ID of the related object.
	ObjectID *string `json:"object_id,omitempty"`

	// EventData contains the event-specific data.
	EventData map[string]interface{} `json:"event_data"`

	// OrganizationID is the ID of the organization the event belongs to.
	OrganizationID *OrganizationID `json:"organization_id,omitempty"`

	// UserID is the ID of the user who triggered the event (if applicable).
	UserID *UserID `json:"user_id,omitempty"`

	// IPAddress is the IP address from which the event was triggered.
	IPAddress *string `json:"ip_address,omitempty"`

	// UserAgent is the user agent string from the request.
	UserAgent *string `json:"user_agent,omitempty"`

	// Source indicates the source of the event (e.g., "api", "dashboard", "system").
	Source *string `json:"source,omitempty"`

	// CreatedOn is when the event occurred.
	CreatedOn *time.Time `json:"created_on,omitempty"`
}

// GetString retrieves a string value from the event data.
func (e *Event) GetString(key string) string {
	if e.EventData == nil {
		return ""
	}
	if val, ok := e.EventData[key].(string); ok {
		return val
	}
	return ""
}

// GetInt retrieves an int value from the event data.
func (e *Event) GetInt(key string) int {
	if e.EventData == nil {
		return 0
	}
	switch v := e.EventData[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}

// GetBool retrieves a bool value from the event data.
func (e *Event) GetBool(key string) bool {
	if e.EventData == nil {
		return false
	}
	if val, ok := e.EventData[key].(bool); ok {
		return val
	}
	return false
}

// EventListResponse represents the paginated response when listing events.
type EventListResponse struct {
	// Results contains the list of events for the current page.
	Results []Event `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// EventSortField represents fields that can be used for sorting events.
type EventSortField string

const (
	EventSortByCreatedOn  EventSortField = "created_on"
	EventSortByType       EventSortField = "type"
	EventSortByObjectType EventSortField = "object_type"
)

// ListEventsOptions contains options for listing events.
type ListEventsOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of events per page.
	PageSize int

	// SortBy is the field to sort by.
	SortBy EventSortField

	// SortOrder is the sort direction.
	SortOrder SortOrder

	// Type filters by event type.
	Type EventType

	// Subtype filters by event subtype.
	Subtype EventSubtype

	// ObjectType filters by object type.
	ObjectType EventObjectType

	// ObjectID filters by object ID.
	ObjectID string

	// CreatedAfter filters events created after this time.
	CreatedAfter *time.Time

	// CreatedBefore filters events created before this time.
	CreatedBefore *time.Time
}

// ObjectLog represents a log entry for an object.
type ObjectLog struct {
	// LogID is the unique identifier for the log entry.
	LogID TypeID `json:"log_id,omitempty"`

	// ObjectID is the ID of the object.
	ObjectID string `json:"object_id"`

	// ObjectType is the type of object.
	ObjectType EventObjectType `json:"object_type"`

	// Action is the action performed.
	Action string `json:"action"`

	// Changes contains the changes made (for update actions).
	Changes *ObjectChanges `json:"changes,omitempty"`

	// UserID is the ID of the user who performed the action.
	UserID *UserID `json:"user_id,omitempty"`

	// IPAddress is the IP address from which the action was performed.
	IPAddress *string `json:"ip_address,omitempty"`

	// CreatedOn is when the log entry was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`
}

// ObjectChanges represents the before and after state of an object.
type ObjectChanges struct {
	// Before contains the state before the change.
	Before map[string]interface{} `json:"before,omitempty"`

	// After contains the state after the change.
	After map[string]interface{} `json:"after,omitempty"`
}

// ObjectLogListResponse represents the paginated response when listing object logs.
type ObjectLogListResponse struct {
	// Results contains the list of log entries for the current page.
	Results []ObjectLog `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// ListObjectLogsOptions contains options for listing object logs.
type ListObjectLogsOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of logs per page.
	PageSize int

	// SortBy is the field to sort by.
	SortBy string

	// SortOrder is the sort direction.
	SortOrder SortOrder

	// ObjectType filters by object type.
	ObjectType EventObjectType

	// ObjectID filters by object ID.
	ObjectID string

	// Action filters by action.
	Action string

	// UserID filters by user ID.
	UserID UserID

	// CreatedAfter filters logs created after this time.
	CreatedAfter *time.Time

	// CreatedBefore filters logs created before this time.
	CreatedBefore *time.Time
}

// RequestHistoryEntry represents an entry in the API request history.
type RequestHistoryEntry struct {
	// RequestID is the unique identifier for the request.
	RequestID TypeID `json:"request_id,omitempty"`

	// Method is the HTTP method.
	Method string `json:"method"`

	// Path is the request path.
	Path string `json:"path"`

	// StatusCode is the HTTP response status code.
	StatusCode int `json:"status_code"`

	// Duration is the request duration in milliseconds.
	Duration int `json:"duration_ms,omitempty"`

	// UserID is the ID of the user who made the request.
	UserID *UserID `json:"user_id,omitempty"`

	// APIKeyID is the ID of the API key used (if applicable).
	APIKeyID *TypeID `json:"api_key_id,omitempty"`

	// IPAddress is the client IP address.
	IPAddress *string `json:"ip_address,omitempty"`

	// UserAgent is the client user agent.
	UserAgent *string `json:"user_agent,omitempty"`

	// CreatedOn is when the request was made.
	CreatedOn *time.Time `json:"created_on,omitempty"`
}

// RequestHistoryListResponse represents the paginated response when listing request history.
type RequestHistoryListResponse struct {
	// Results contains the list of request history entries.
	Results []RequestHistoryEntry `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}
