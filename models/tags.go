// Package models contains all the data types for the OpusDNS API.
package models

import "time"

// TagColor represents one of the API-supported tag colors.
type TagColor string

const (
	TagColor1  TagColor = "color-1"
	TagColor2  TagColor = "color-2"
	TagColor3  TagColor = "color-3"
	TagColor4  TagColor = "color-4"
	TagColor5  TagColor = "color-5"
	TagColor6  TagColor = "color-6"
	TagColor7  TagColor = "color-7"
	TagColor8  TagColor = "color-8"
	TagColor9  TagColor = "color-9"
	TagColor10 TagColor = "color-10"
)

// TagType is the resource category a tag applies to.
type TagType string

const (
	TagTypeDomain  TagType = "DOMAIN"
	TagTypeContact TagType = "CONTACT"
	TagTypeZone    TagType = "ZONE"
)

// TagSortField represents fields that can be used for sorting tags.
type TagSortField string

const (
	TagSortByLabel     TagSortField = "label"
	TagSortByCreatedOn TagSortField = "created_on"
	TagSortByUpdatedOn TagSortField = "updated_on"
)

// Tag represents a tag.
type Tag struct {
	TagID       TagID     `json:"tag_id"`
	Label       string    `json:"label"`
	Type        TagType   `json:"type"`
	Color       TagColor  `json:"color"`
	Description *string   `json:"description,omitempty"`
	ObjectCount int       `json:"object_count,omitempty"`
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"updated_on"`
}

// TagEnriched represents tag data embedded in resource responses.
type TagEnriched struct {
	TagID TagID    `json:"tag_id"`
	Label string   `json:"label"`
	Color TagColor `json:"color"`
}

// TagListResponse represents the paginated response when listing tags.
type TagListResponse struct {
	Results    []Tag      `json:"results"`
	Pagination Pagination `json:"pagination"`
}

// TagCreateRequest represents a request to create a tag.
type TagCreateRequest struct {
	Label       string    `json:"label"`
	Type        TagType   `json:"type"`
	Color       *TagColor `json:"color,omitempty"`
	Description *string   `json:"description,omitempty"`
}

// TagUpdateRequest represents a request to update a tag.
type TagUpdateRequest struct {
	Label       *string   `json:"label,omitempty"`
	Color       *TagColor `json:"color,omitempty"`
	Description *string   `json:"description,omitempty"`
}

// ListTagsOptions contains options for listing tags.
type ListTagsOptions struct {
	Page      int
	PageSize  int
	SortBy    TagSortField
	SortOrder SortOrder
	TagTypes  []TagType
	Search    string
}

// ObjectTagChanges describes object changes for a single tag.
type ObjectTagChanges struct {
	Add    []string `json:"add,omitempty"`
	Remove []string `json:"remove,omitempty"`
}

// BulkObjectTagChanges describes tag changes for multiple objects.
type BulkObjectTagChanges struct {
	Type    TagType  `json:"type"`
	Objects []string `json:"objects"`
	Add     []TagID  `json:"add,omitempty"`
	Remove  []TagID  `json:"remove,omitempty"`
	Replace []TagID  `json:"replace"`
}

// ObjectTagChangesResponse summarizes tag object update results.
type ObjectTagChangesResponse struct {
	Added      int      `json:"added"`
	Removed    int      `json:"removed"`
	Unresolved []string `json:"unresolved,omitempty"`
}
