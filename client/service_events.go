package client

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/opusdns/opusdns-go-client/models"
)

// EventsService provides methods for accessing events and audit logs.
type EventsService struct {
	client *Client
}

// ListEvents retrieves events with automatic pagination.
func (s *EventsService) ListEvents(ctx context.Context, opts *models.ListEventsOptions) ([]models.Event, error) {
	var all []models.Event
	page := 1

	for {
		pageOpts := opts
		if pageOpts == nil {
			pageOpts = &models.ListEventsOptions{}
		}
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := s.ListEventsPage(ctx, pageOpts)
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

// ListEventsPage retrieves a single page of events.
func (s *EventsService) ListEventsPage(ctx context.Context, opts *models.ListEventsOptions) (*models.EventListResponse, error) {
	path := s.client.http.BuildPath("events")

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
		if opts.Type != "" {
			query.Set("type", string(opts.Type))
		}
		if opts.Subtype != "" {
			query.Set("subtype", string(opts.Subtype))
		}
		if opts.ObjectType != "" {
			query.Set("object_type", string(opts.ObjectType))
		}
		if opts.ObjectID != "" {
			query.Set("object_id", opts.ObjectID)
		}
		if opts.CreatedAfter != nil {
			query.Set("created_after", opts.CreatedAfter.Format(time.RFC3339))
		}
		if opts.CreatedBefore != nil {
			query.Set("created_before", opts.CreatedBefore.Format(time.RFC3339))
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.EventListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetEvent retrieves a specific event by ID.
func (s *EventsService) GetEvent(ctx context.Context, eventID models.EventID) (*models.Event, error) {
	path := s.client.http.BuildPath("events", string(eventID))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var event models.Event
	if err := s.client.http.DecodeResponse(resp, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

// ListObjectLogs retrieves object logs.
func (s *EventsService) ListObjectLogs(ctx context.Context, opts *models.ListObjectLogsOptions) (*models.ObjectLogListResponse, error) {
	path := s.client.http.BuildPath("archive", "object-logs")

	query := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.SortBy != "" {
			query.Set("sort_by", opts.SortBy)
		}
		if opts.SortOrder != "" {
			query.Set("sort_order", string(opts.SortOrder))
		}
		if opts.ObjectType != "" {
			query.Set("object_type", string(opts.ObjectType))
		}
		if opts.ObjectID != "" {
			query.Set("object_id", opts.ObjectID)
		}
		if opts.Action != "" {
			query.Set("action", opts.Action)
		}
		if opts.UserID != "" {
			query.Set("user_id", string(opts.UserID))
		}
		if opts.CreatedAfter != nil {
			query.Set("created_after", opts.CreatedAfter.Format(time.RFC3339))
		}
		if opts.CreatedBefore != nil {
			query.Set("created_before", opts.CreatedBefore.Format(time.RFC3339))
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.ObjectLogListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetObjectLog retrieves logs for a specific object.
func (s *EventsService) GetObjectLog(ctx context.Context, objectID string) (*models.ObjectLogListResponse, error) {
	path := s.client.http.BuildPath("archive", "object-logs", url.PathEscape(objectID))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.ObjectLogListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ListRequestHistory retrieves API request history.
func (s *EventsService) ListRequestHistory(ctx context.Context, opts *models.ListOptions) (*models.RequestHistoryListResponse, error) {
	path := s.client.http.BuildPath("archive", "request-history")

	query := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.SortBy != "" {
			query.Set("sort_by", opts.SortBy)
		}
		if opts.SortOrder != "" {
			query.Set("sort_order", string(opts.SortOrder))
		}
	}

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.RequestHistoryListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ListEmailForwardLogs retrieves email forward logs for a specific email forward.
func (s *EventsService) ListEmailForwardLogs(ctx context.Context, emailForwardID models.EmailForwardID) (*models.EmailForwardLogListResponse, error) {
	path := s.client.http.BuildPath("archive", "email-forward-logs", string(emailForwardID))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.EmailForwardLogListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ListEmailForwardLogsByAlias retrieves email forward logs for a specific alias.
func (s *EventsService) ListEmailForwardLogsByAlias(ctx context.Context, aliasID models.EmailForwardAliasID) (*models.EmailForwardLogListResponse, error) {
	path := s.client.http.BuildPath("archive", "email-forward-logs", "aliases", string(aliasID))

	resp, err := s.client.http.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result models.EmailForwardLogListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
