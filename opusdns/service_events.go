package opusdns

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
//
// With opts or opts.Acknowledged nil, the API returns only unacknowledged
// events. Set opts.Acknowledged to models.BoolPtr(true) to list the
// acknowledged ones.
func (s *EventsService) ListEvents(ctx context.Context, opts *models.ListEventsOptions) ([]models.Event, error) {
	var all []models.Event
	page := 1

	for {
		pageOpts := cloneOptions(opts)
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

// ListEventsPage retrieves a single page of events. Like ListEvents, it
// returns only unacknowledged events unless opts.Acknowledged is set.
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
		if opts.Acknowledged != nil {
			query.Set("acknowledged", strconv.FormatBool(*opts.Acknowledged))
		}
		if opts.ObjectType != "" {
			query.Set("object_type", string(opts.ObjectType))
		}
		if opts.ObjectID != "" {
			query.Set("object_id", opts.ObjectID)
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
	path := s.client.http.BuildPath("events", url.PathEscape(string(eventID)))

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

// AcknowledgeEvent acknowledges an event by ID.
func (s *EventsService) AcknowledgeEvent(ctx context.Context, eventID models.EventID) error {
	path := s.client.http.BuildPath("events", url.PathEscape(string(eventID)))

	resp, err := s.client.http.Patch(ctx, path, nil)
	if err != nil {
		return err
	}

	return s.client.http.DecodeResponse(resp, nil)
}

// ListObjectLogs retrieves a single page of object logs across all objects.
func (s *EventsService) ListObjectLogs(ctx context.Context, opts *models.ListObjectLogsOptions) (*models.ObjectLogListResponse, error) {
	path := s.client.http.BuildPath("archive", "object-logs")

	query := objectLogsQuery(opts)
	if opts != nil && opts.ObjectID != "" {
		query.Set("object_id", opts.ObjectID)
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

// ListObjectLogsByObject retrieves the logs of one object with automatic
// pagination. opts.ObjectID is ignored.
func (s *EventsService) ListObjectLogsByObject(ctx context.Context, objectID string, opts *models.ListObjectLogsOptions) ([]models.ObjectLog, error) {
	var all []models.ObjectLog
	page := 1

	for {
		pageOpts := cloneOptions(opts)
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := s.ListObjectLogsByObjectPage(ctx, objectID, pageOpts)
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

// ListObjectLogsByObjectPage retrieves a single page of the logs of one
// object. opts.ObjectID is ignored.
func (s *EventsService) ListObjectLogsByObjectPage(ctx context.Context, objectID string, opts *models.ListObjectLogsOptions) (*models.ObjectLogListResponse, error) {
	path := s.client.http.BuildPath("archive", "object-logs", url.PathEscape(objectID))

	resp, err := s.client.http.Get(ctx, path, objectLogsQuery(opts))
	if err != nil {
		return nil, err
	}

	var result models.ObjectLogListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetObjectLog retrieves the first page of logs for a specific object.
//
// Deprecated: GetObjectLog cannot read past the first page. Use
// ListObjectLogsByObject or ListObjectLogsByObjectPage.
func (s *EventsService) GetObjectLog(ctx context.Context, objectID string) (*models.ObjectLogListResponse, error) {
	return s.ListObjectLogsByObjectPage(ctx, objectID, nil)
}

// objectLogsQuery builds the query shared by both object log routes. The
// object_id filter only exists on the route without the ID in its path, so
// the caller adds it.
func objectLogsQuery(opts *models.ListObjectLogsOptions) url.Values {
	query := url.Values{}
	if opts == nil {
		return query
	}

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
		query.Set("object_type", opts.ObjectType)
	}
	if opts.ObjectLogID != "" {
		query.Set("object_log_id", opts.ObjectLogID)
	}
	if opts.Action != "" {
		query.Set("action", opts.Action)
	}
	if opts.ServerRequestID != "" {
		query.Set("server_request_id", opts.ServerRequestID)
	}
	if opts.PerformedByType != "" {
		query.Set("performed_by_type", string(opts.PerformedByType))
	}
	performedByID := opts.PerformedByID
	if performedByID == "" {
		performedByID = string(opts.UserID) //nolint:staticcheck // deprecated alias
	}
	if performedByID != "" {
		query.Set("performed_by_id", performedByID)
	}
	if opts.CreatedAfter != nil {
		query.Set("created_after", opts.CreatedAfter.Format(time.RFC3339Nano))
	}
	if opts.CreatedBefore != nil {
		query.Set("created_before", opts.CreatedBefore.Format(time.RFC3339Nano))
	}

	return query
}

// ListRequestHistory retrieves a single page of the API request history.
func (s *EventsService) ListRequestHistory(ctx context.Context, opts *models.ListRequestHistoryOptions) (*models.RequestHistoryListResponse, error) {
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
		if opts.Method != "" {
			query.Set("method", string(opts.Method))
		}
		if opts.Path != "" {
			query.Set("path", opts.Path)
		}
		if opts.StatusCode != nil {
			query.Set("status_code", strconv.Itoa(*opts.StatusCode))
		}
		if opts.MinStatusCode != nil {
			query.Set("min_status_code", strconv.Itoa(*opts.MinStatusCode))
		}
		if opts.MaxStatusCode != nil {
			query.Set("max_status_code", strconv.Itoa(*opts.MaxStatusCode))
		}
		if opts.MinDuration != nil {
			query.Set("min_duration", strconv.FormatFloat(*opts.MinDuration, 'f', -1, 64))
		}
		if opts.MaxDuration != nil {
			query.Set("max_duration", strconv.FormatFloat(*opts.MaxDuration, 'f', -1, 64))
		}
		if opts.ClientIP != "" {
			query.Set("client_ip", opts.ClientIP)
		}
		if opts.ServerRequestID != "" {
			query.Set("server_request_id", opts.ServerRequestID)
		}
		if opts.PerformedByType != "" {
			query.Set("performed_by_type", string(opts.PerformedByType))
		}
		if opts.PerformedByID != "" {
			query.Set("performed_by_id", opts.PerformedByID)
		}
		if opts.RequestStartedBefore != nil {
			query.Set("request_started_before", opts.RequestStartedBefore.Format(time.RFC3339Nano))
		}
		if opts.RequestStartedAfter != nil {
			query.Set("request_started_after", opts.RequestStartedAfter.Format(time.RFC3339Nano))
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

// ListEmailForwardLogs retrieves the logs of an email forward with automatic
// pagination.
func (s *EventsService) ListEmailForwardLogs(ctx context.Context, emailForwardID models.EmailForwardID, opts *models.ListEmailForwardLogsOptions) ([]models.EmailForwardLog, error) {
	return collectEmailForwardLogs(opts, func(pageOpts *models.ListEmailForwardLogsOptions) (*models.EmailForwardLogListResponse, error) {
		return s.ListEmailForwardLogsPage(ctx, emailForwardID, pageOpts)
	})
}

// ListEmailForwardLogsPage retrieves a single page of the logs of an email
// forward.
func (s *EventsService) ListEmailForwardLogsPage(ctx context.Context, emailForwardID models.EmailForwardID, opts *models.ListEmailForwardLogsOptions) (*models.EmailForwardLogListResponse, error) {
	path := s.client.http.BuildPath("archive", "email-forward-logs", url.PathEscape(string(emailForwardID)))

	resp, err := s.client.http.Get(ctx, path, emailForwardLogsQuery(opts))
	if err != nil {
		return nil, err
	}

	var result models.EmailForwardLogListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ListEmailForwardLogsByAlias retrieves the logs of one alias of an email
// forward with automatic pagination.
func (s *EventsService) ListEmailForwardLogsByAlias(ctx context.Context, emailForwardID models.EmailForwardID, aliasID models.EmailForwardAliasID, opts *models.ListEmailForwardLogsOptions) ([]models.EmailForwardLog, error) {
	return collectEmailForwardLogs(opts, func(pageOpts *models.ListEmailForwardLogsOptions) (*models.EmailForwardLogListResponse, error) {
		return s.ListEmailForwardLogsByAliasPage(ctx, emailForwardID, aliasID, pageOpts)
	})
}

// ListEmailForwardLogsByAliasPage retrieves a single page of the logs of one
// alias of an email forward. The API requires the email forward the alias
// belongs to.
func (s *EventsService) ListEmailForwardLogsByAliasPage(ctx context.Context, emailForwardID models.EmailForwardID, aliasID models.EmailForwardAliasID, opts *models.ListEmailForwardLogsOptions) (*models.EmailForwardLogListResponse, error) {
	path := s.client.http.BuildPath("archive", "email-forward-logs", "aliases", url.PathEscape(string(aliasID)))

	query := emailForwardLogsQuery(opts)
	query.Set("email_forward_id", string(emailForwardID))

	resp, err := s.client.http.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}

	var result models.EmailForwardLogListResponse
	if err := s.client.http.DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// collectEmailForwardLogs pages through fetch until the last page.
func collectEmailForwardLogs(opts *models.ListEmailForwardLogsOptions, fetch func(*models.ListEmailForwardLogsOptions) (*models.EmailForwardLogListResponse, error)) ([]models.EmailForwardLog, error) {
	var all []models.EmailForwardLog
	page := 1

	for {
		pageOpts := cloneOptions(opts)
		pageOpts.Page = page
		if pageOpts.PageSize == 0 {
			pageOpts.PageSize = DefaultPageSize
		}

		resp, err := fetch(pageOpts)
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

// emailForwardLogsQuery builds the query shared by both email forward log
// routes.
func emailForwardLogsQuery(opts *models.ListEmailForwardLogsOptions) url.Values {
	query := url.Values{}
	if opts == nil {
		return query
	}

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
	if opts.FinalStatus != "" {
		query.Set("final_status", string(opts.FinalStatus))
	}
	if opts.StartTime != nil {
		query.Set("start_time", opts.StartTime.Format(time.RFC3339Nano))
	}
	if opts.EndTime != nil {
		query.Set("end_time", opts.EndTime.Format(time.RFC3339Nano))
	}

	return query
}
