// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// DomainForwardID is a TypeID for domain forwards.
type DomainForwardID = TypeID

// DomainForwardProtocol represents the protocol for domain forwarding.
type DomainForwardProtocol string

const (
	// DomainForwardProtocolHTTP forwards HTTP traffic.
	DomainForwardProtocolHTTP DomainForwardProtocol = "http"

	// DomainForwardProtocolHTTPS forwards HTTPS traffic.
	DomainForwardProtocolHTTPS DomainForwardProtocol = "https"
)

// DomainForwardType represents the type of domain forward.
type DomainForwardType string

const (
	// DomainForwardTypePermanent performs a 301 permanent redirect.
	DomainForwardTypePermanent DomainForwardType = "permanent"

	// DomainForwardTypeTemporary performs a 302 temporary redirect.
	DomainForwardTypeTemporary DomainForwardType = "temporary"

	// DomainForwardTypeFrame displays the destination in a frame (masked).
	DomainForwardTypeFrame DomainForwardType = "frame"
)

// DomainForwardSortField represents fields that can be used for sorting domain forwards.
type DomainForwardSortField string

const (
	DomainForwardSortByHostname  DomainForwardSortField = "hostname"
	DomainForwardSortByCreatedOn DomainForwardSortField = "created_on"
	DomainForwardSortByUpdatedOn DomainForwardSortField = "updated_on"
)

// DomainForward represents a domain forwarding configuration.
type DomainForward struct {
	// DomainForwardID is the unique identifier for the domain forward.
	DomainForwardID DomainForwardID `json:"domain_forward_id,omitempty"`

	// Hostname is the source hostname (e.g., "www.example.com" or "example.com").
	Hostname string `json:"hostname"`

	// Enabled indicates whether the domain forward is active.
	Enabled bool `json:"enabled"`

	// Configs contains the forwarding configurations for each protocol.
	Configs []DomainForwardConfig `json:"configs,omitempty"`

	// CreatedOn is when the domain forward was created.
	CreatedOn *time.Time `json:"created_on,omitempty"`

	// UpdatedOn is when the domain forward was last updated.
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// GetConfig returns the config for the specified protocol, or nil if not found.
func (d *DomainForward) GetConfig(protocol DomainForwardProtocol) *DomainForwardConfig {
	for i := range d.Configs {
		if d.Configs[i].Protocol == protocol {
			return &d.Configs[i]
		}
	}
	return nil
}

// DomainForwardConfig represents the forwarding configuration for a specific protocol.
type DomainForwardConfig struct {
	// Protocol is the source protocol (http or https).
	Protocol DomainForwardProtocol `json:"protocol"`

	// DestinationURL is the URL to forward to.
	DestinationURL string `json:"destination_url"`

	// ForwardType is the type of redirect.
	ForwardType DomainForwardType `json:"forward_type"`

	// IncludePath indicates whether to append the request path to the destination.
	IncludePath bool `json:"include_path"`

	// IncludeQuery indicates whether to append query parameters to the destination.
	IncludeQuery bool `json:"include_query"`

	// FrameTitle is the title for frame-type forwards (optional).
	FrameTitle *string `json:"frame_title,omitempty"`

	// FrameDescription is the meta description for frame-type forwards (optional).
	FrameDescription *string `json:"frame_description,omitempty"`

	// FrameKeywords are the meta keywords for frame-type forwards (optional).
	FrameKeywords *string `json:"frame_keywords,omitempty"`

	// FrameFavicon is the favicon URL for frame-type forwards (optional).
	FrameFavicon *string `json:"frame_favicon,omitempty"`
}

// DomainForwardListResponse represents the paginated response when listing domain forwards.
type DomainForwardListResponse struct {
	// Results contains the list of domain forwards for the current page.
	Results []DomainForward `json:"results"`

	// Pagination contains the pagination metadata.
	Pagination Pagination `json:"pagination"`
}

// DomainForwardCreateRequest represents a request to create a domain forward.
type DomainForwardCreateRequest struct {
	// Hostname is the source hostname to forward from.
	Hostname string `json:"hostname"`

	// Configs contains the forwarding configurations for each protocol.
	Configs []DomainForwardConfigCreate `json:"configs"`
}

// DomainForwardConfigCreate represents the configuration for creating a domain forward.
type DomainForwardConfigCreate struct {
	// Protocol is the source protocol (http or https).
	Protocol DomainForwardProtocol `json:"protocol"`

	// DestinationURL is the URL to forward to.
	DestinationURL string `json:"destination_url"`

	// ForwardType is the type of redirect (permanent, temporary, or frame).
	ForwardType DomainForwardType `json:"forward_type"`

	// IncludePath indicates whether to append the request path to the destination.
	IncludePath bool `json:"include_path,omitempty"`

	// IncludeQuery indicates whether to append query parameters to the destination.
	IncludeQuery bool `json:"include_query,omitempty"`

	// FrameTitle is the title for frame-type forwards (optional).
	FrameTitle *string `json:"frame_title,omitempty"`

	// FrameDescription is the meta description for frame-type forwards (optional).
	FrameDescription *string `json:"frame_description,omitempty"`

	// FrameKeywords are the meta keywords for frame-type forwards (optional).
	FrameKeywords *string `json:"frame_keywords,omitempty"`

	// FrameFavicon is the favicon URL for frame-type forwards (optional).
	FrameFavicon *string `json:"frame_favicon,omitempty"`
}

// DomainForwardConfigUpdate represents the configuration for updating a domain forward.
type DomainForwardConfigUpdate struct {
	// DestinationURL is the URL to forward to.
	DestinationURL *string `json:"destination_url,omitempty"`

	// ForwardType is the type of redirect.
	ForwardType *DomainForwardType `json:"forward_type,omitempty"`

	// IncludePath indicates whether to append the request path to the destination.
	IncludePath *bool `json:"include_path,omitempty"`

	// IncludeQuery indicates whether to append query parameters to the destination.
	IncludeQuery *bool `json:"include_query,omitempty"`

	// FrameTitle is the title for frame-type forwards.
	FrameTitle *string `json:"frame_title,omitempty"`

	// FrameDescription is the meta description for frame-type forwards.
	FrameDescription *string `json:"frame_description,omitempty"`

	// FrameKeywords are the meta keywords for frame-type forwards.
	FrameKeywords *string `json:"frame_keywords,omitempty"`

	// FrameFavicon is the favicon URL for frame-type forwards.
	FrameFavicon *string `json:"frame_favicon,omitempty"`
}

// ListDomainForwardsOptions contains options for listing domain forwards.
type ListDomainForwardsOptions struct {
	// Page is the page number to retrieve (1-indexed).
	Page int

	// PageSize is the number of items per page.
	PageSize int

	// SortBy is the field to sort by.
	SortBy DomainForwardSortField

	// SortOrder is the sort direction.
	SortOrder SortOrder

	// Search is an optional search query to filter by hostname.
	Search string

	// Enabled filters by enabled status.
	Enabled *bool
}
