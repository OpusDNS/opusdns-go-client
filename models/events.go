// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// EventID is a TypeID for events.
type EventID = TypeID

// EventType represents the type of an event.
type EventType string

const (
	EventTypeRegistration        EventType = "REGISTRATION"
	EventTypeRenewal             EventType = "RENEWAL"
	EventTypeModification        EventType = "MODIFICATION"
	EventTypeDeletion            EventType = "DELETION"
	EventTypeInboundTransfer     EventType = "INBOUND_TRANSFER"
	EventTypeOutboundTransfer    EventType = "OUTBOUND_TRANSFER"
	EventTypeTransit             EventType = "TRANSIT"
	EventTypeWithdraw            EventType = "WITHDRAW"
	EventTypeVerification        EventType = "VERIFICATION"
	EventTypeBalance             EventType = "BALANCE"
	EventTypeVanityNSProvision   EventType = "VANITY_NS_PROVISION"
	EventTypeVanityNSSuspension  EventType = "VANITY_NS_SUSPENSION"
	EventTypeVanityNSRestoration EventType = "VANITY_NS_RESTORATION"
	EventTypeVanityNSTermination EventType = "VANITY_NS_TERMINATION"
)

// EventSubtype represents a more specific event subtype.
type EventSubtype string

const (
	EventSubtypeNotification EventSubtype = "NOTIFICATION"
	EventSubtypeSuccess      EventSubtype = "SUCCESS"
	EventSubtypeFailure      EventSubtype = "FAILURE"
	EventSubtypeCanceled     EventSubtype = "CANCELED"
)

// EventObjectType represents the type of object an event relates to.
type EventObjectType string

const (
	EventObjectTypeDomain      EventObjectType = "DOMAIN"
	EventObjectTypeContact     EventObjectType = "CONTACT"
	EventObjectTypeHost        EventObjectType = "HOST"
	EventObjectTypeAccount     EventObjectType = "ACCOUNT"
	EventObjectTypeVanityNSSet EventObjectType = "VANITY_NS_SET"
	EventObjectTypeRaw         EventObjectType = "RAW"
	EventObjectTypeUnknown     EventObjectType = "UNKNOWN"
)

// EventVersion represents the schema version of event data.
type EventVersion string

const (
	// EventVersionV1 is the "1.0" event data version.
	EventVersionV1 EventVersion = "1.0"
)

// EventError represents an error embedded in event data.
type EventError struct {
	// Code is the error code.
	Code string `json:"code"`

	// Detail is the human-readable error detail.
	Detail string `json:"detail"`
}

// VerificationClaimType represents a type of verification claim.
type VerificationClaimType string

const (
	VerificationClaimTypeName    VerificationClaimType = "name"
	VerificationClaimTypeAddress VerificationClaimType = "address"
	VerificationClaimTypeEmail   VerificationClaimType = "email"
	VerificationClaimTypePhone   VerificationClaimType = "phone"
)

// VerificationDeadlineType represents the type of a verification deadline.
type VerificationDeadlineType string

const (
	VerificationDeadlineTypeDedelegation VerificationDeadlineType = "dedelegation"
	VerificationDeadlineTypeDeletion     VerificationDeadlineType = "deletion"
)

// VerificationDeadline represents a deadline for domain verification.
type VerificationDeadline struct {
	// Type is the type of deadline.
	Type VerificationDeadlineType `json:"type"`

	// Date is when the deadline occurs.
	Date *time.Time `json:"date"`
}

// VerificationRegistrantDetails represents a registrant in a domain verification.
type VerificationRegistrantDetails struct {
	// ContactID is the ID of the registrant contact.
	ContactID string `json:"contact_id"`

	// Name is the registrant name.
	Name string `json:"name"`

	// Email is the registrant email.
	Email string `json:"email"`
}

// EventDetails contains the discriminated union of event detail variants.
// DetailType selects the variant ("domain_renewal" or "domain_verification");
// variant-specific fields are populated accordingly.
type EventDetails struct {
	// DetailType is the discriminator selecting the detail variant.
	DetailType string `json:"detail_type,omitempty"`

	// ExpiresOn is set for the "domain_renewal" variant.
	ExpiresOn *time.Time `json:"expires_on,omitempty"`

	// DomainID is set for the "domain_verification" variant.
	DomainID string `json:"domain_id,omitempty"`

	// VerificationDeadlines is set for the "domain_verification" variant.
	VerificationDeadlines []VerificationDeadline `json:"verification_deadlines,omitempty"`

	// VerificationClaims is set for the "domain_verification" variant.
	VerificationClaims []VerificationClaimType `json:"verification_claims,omitempty"`

	// Registrants is set for the "domain_verification" variant.
	Registrants []VerificationRegistrantDetails `json:"registrants,omitempty"`
}

// EventData contains the event-specific data.
type EventData struct {
	// Version is the schema version of the event data.
	Version EventVersion `json:"version,omitempty"`

	// Message is the event message.
	Message string `json:"message"`

	// Error contains error details, if any.
	Error *EventError `json:"error,omitempty"`

	// Details contains the discriminated detail union, if any.
	Details *EventDetails `json:"details,omitempty"`
}

// Event represents an event in the OpusDNS system.
type Event struct {
	// EventID is the unique identifier for the event.
	EventID EventID `json:"event_id"`

	// EventData contains the event-specific data.
	EventData EventData `json:"event_data"`

	// CreatedOn is when the event was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// ObjectID is the ID of the related object.
	ObjectID *string `json:"object_id,omitempty"`

	// ObjectType is the type of object the event relates to.
	ObjectType EventObjectType `json:"object_type,omitempty"`

	// Type is the event type.
	Type *EventType `json:"type,omitempty"`

	// Subtype is the event subtype for more specific categorization.
	Subtype *EventSubtype `json:"subtype,omitempty"`

	// AcknowledgedOn is when the event was acknowledged.
	AcknowledgedOn *time.Time `json:"acknowledged_on,omitempty"`
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
	EventSortByObjectID  EventSortField = "object_id"
	EventSortByCreatedOn EventSortField = "created_on"
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

	// Acknowledged filters events by acknowledgement status.
	Acknowledged *bool

	// ObjectType filters by object type.
	ObjectType EventObjectType

	// ObjectID filters by object ID.
	ObjectID string
}

// ObjectEventType represents the action recorded in an object log entry.
type ObjectEventType string

const (
	ObjectEventTypeCreated                     ObjectEventType = "CREATED"
	ObjectEventTypeUpdated                     ObjectEventType = "UPDATED"
	ObjectEventTypeDeleted                     ObjectEventType = "DELETED"
	ObjectEventTypeImported                    ObjectEventType = "IMPORTED"
	ObjectEventTypeTransferStarted             ObjectEventType = "TRANSFER_STARTED"
	ObjectEventTypeTransferCompleted           ObjectEventType = "TRANSFER_COMPLETED"
	ObjectEventTypeTransferOutStarted          ObjectEventType = "TRANSFER_OUT_STARTED"
	ObjectEventTypeTransferOutCompleted        ObjectEventType = "TRANSFER_OUT_COMPLETED"
	ObjectEventTypeRenewed                     ObjectEventType = "RENEWED"
	ObjectEventTypeRestored                    ObjectEventType = "RESTORED"
	ObjectEventTypeBillingTransactionReserved  ObjectEventType = "BILLING_TRANSACTION_RESERVED"
	ObjectEventTypeBillingTransactionSucceeded ObjectEventType = "BILLING_TRANSACTION_SUCCEEDED"
	ObjectEventTypeBillingTransactionFailed    ObjectEventType = "BILLING_TRANSACTION_FAILED"
	ObjectEventTypeBillingTransactionCancelled ObjectEventType = "BILLING_TRANSACTION_CANCELLED"
)

// ObjectLog represents a log entry for an object.
type ObjectLog struct {
	// ObjectLogID is the unique identifier for the log entry.
	ObjectLogID string `json:"object_log_id"`

	// ObjectID is the ID of the object.
	ObjectID string `json:"object_id"`

	// ObjectType is the type of object (free-form string, e.g. "domain", "billing_transaction").
	ObjectType string `json:"object_type"`

	// Action is the action performed.
	Action ObjectEventType `json:"action"`

	// Details contains the changes made to the object.
	Details *map[string]interface{} `json:"details,omitempty"`

	// PerformedByID is the ID of the actor who performed the action.
	PerformedByID *string `json:"performed_by_id,omitempty"`

	// PerformedByType is the type of the actor who performed the action.
	PerformedByType *ExecutingEntity `json:"performed_by_type,omitempty"`

	// ServerRequestID is the server request ID associated with the action.
	ServerRequestID *string `json:"server_request_id,omitempty"`

	// CreatedOn is when the log entry was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`
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
	// ServerRequestID is the unique identifier for the request.
	ServerRequestID string `json:"server_request_id"`

	// Method is the HTTP method.
	Method HTTPMethod `json:"method"`

	// Path is the request path.
	Path string `json:"path"`

	// StatusCode is the HTTP response status code.
	StatusCode int `json:"status_code"`

	// Duration is the request duration in milliseconds.
	Duration float64 `json:"duration"`

	// ClientIP is the client IP address.
	ClientIP string `json:"client_ip"`

	// PerformedByID is the ID of the actor who performed the request.
	PerformedByID *string `json:"performed_by_id,omitempty"`

	// PerformedByType is the type of the actor who performed the request.
	PerformedByType *ExecutingEntity `json:"performed_by_type,omitempty"`

	// RequestBody is the request body.
	RequestBody interface{} `json:"request_body,omitempty"`

	// ResponseBody is the response body.
	ResponseBody interface{} `json:"response_body,omitempty"`

	// RequestStartedAt is when the request started.
	RequestStartedAt *time.Time `json:"request_started_at,omitempty"`

	// RequestCompletedAt is when the request completed.
	RequestCompletedAt *time.Time `json:"request_completed_at,omitempty"`
}

// RequestHistoryListResponse represents the paginated response when listing request history.
type RequestHistoryListResponse struct {
	// Results contains the list of request history entries.
	Results []RequestHistoryEntry `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}
