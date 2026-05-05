package opusdns

import (
	"context"
	"net/url"
	"strconv"

	"github.com/opusdns/opusdns-go-client/models"
)

// TagsService provides methods for managing tags.
type TagsService struct {
	client *Client
}

// ListTags retrieves all tags with automatic pagination.
func (s *TagsService) ListTags(ctx context.Context, opts *models.ListTagsOptions) ([]models.Tag, error) {
	var all []models.Tag
	page := 1

	for {
		pageOpts := cloneOptions(opts)
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := s.ListTagsPage(ctx, pageOpts)
		if err != nil {
			return nil, err
		}

		all = append(all, resp.Results...)

		if !resp.Pagination.HasNextPage {
			break
		}
		page++
	}

	return all, nil
}

// ListTagsPage retrieves a single page of tags.
func (s *TagsService) ListTagsPage(ctx context.Context, opts *models.ListTagsOptions) (*models.TagListResponse, error) {
	path := s.client.http.BuildPath("tags")

	query := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.SortBy != "" {
			query.Set("sort_by", string(opts.SortBy))
		}
		if opts.SortOrder != "" {
			query.Set("sort_order", string(opts.SortOrder))
		}
		for _, tagType := range opts.TagTypes {
			query.Add("tag_types", string(tagType))
		}
		if opts.Search != "" {
			query.Set("search", opts.Search)
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.TagListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTag retrieves a tag by ID.
func (s *TagsService) GetTag(ctx context.Context, tagID models.TagID) (*models.Tag, error) {
	path := s.client.http.BuildPath("tags", string(tagID))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.Tag
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateTag creates a new tag.
func (s *TagsService) CreateTag(ctx context.Context, req *models.TagCreateRequest) (*models.Tag, error) {
	path := s.client.http.BuildPath("tags")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var result models.Tag
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateTag updates a tag.
func (s *TagsService) UpdateTag(ctx context.Context, tagID models.TagID, req *models.TagUpdateRequest) (*models.Tag, error) {
	path := s.client.http.BuildPath("tags", string(tagID))

	resp, err := s.client.http.Patch(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var result models.Tag
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteTag deletes a tag.
func (s *TagsService) DeleteTag(ctx context.Context, tagID models.TagID) error {
	path := s.client.http.BuildPath("tags", string(tagID))

	resp, err := s.client.http.Delete(ctx, path)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// UpdateTagObjects adds or removes objects from a tag.
func (s *TagsService) UpdateTagObjects(ctx context.Context, tagID models.TagID, req *models.ObjectTagChanges) (*models.ObjectTagChangesResponse, error) {
	path := s.client.http.BuildPath("tags", string(tagID), "objects")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var result models.ObjectTagChangesResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BulkUpdateObjects adds, removes, or replaces tags on multiple objects.
func (s *TagsService) BulkUpdateObjects(ctx context.Context, req *models.BulkObjectTagChanges) (*models.ObjectTagChangesResponse, error) {
	path := s.client.http.BuildPath("tags", "objects")

	resp, err := s.client.http.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var result models.ObjectTagChangesResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
