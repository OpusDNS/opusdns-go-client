package opusdns

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventsService_AcknowledgeEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/events/evt_1", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Events.AcknowledgeEvent(context.Background(), "evt_1")
	require.NoError(t, err)
}

func TestEventsService_ListEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/events", r.URL.Path)
		registration := models.EventTypeRegistration
		renewal := models.EventTypeRenewal
		_ = json.NewEncoder(w).Encode(models.EventListResponse{
			Results: []models.Event{
				{
					EventID:   "evt_1",
					Type:      &registration,
					EventData: models.EventData{Version: models.EventVersionV1, Message: "domain registered"},
				},
				{
					EventID:   "evt_2",
					Type:      &renewal,
					EventData: models.EventData{Version: models.EventVersionV1, Message: "domain renewed"},
				},
			},
			Pagination: models.Pagination{HasNextPage: false, CurrentPage: 1},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	events, err := client.Events.ListEvents(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, models.EventID("evt_1"), events[0].EventID)
}

func TestEventsService_ListEventsPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/events", r.URL.Path)
		query := r.URL.Query()
		assert.Equal(t, "2", query.Get("page"))
		assert.Equal(t, "50", query.Get("page_size"))
		assert.Equal(t, "REGISTRATION", query.Get("type"))
		assert.Equal(t, "DOMAIN", query.Get("object_type"))

		registration := models.EventTypeRegistration
		_ = json.NewEncoder(w).Encode(models.EventListResponse{
			Results: []models.Event{{
				EventID:   "evt_1",
				Type:      &registration,
				EventData: models.EventData{Version: models.EventVersionV1, Message: "domain registered"},
			}},
			Pagination: models.Pagination{HasNextPage: false, CurrentPage: 2},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	opts := &models.ListEventsOptions{
		Page:       2,
		PageSize:   50,
		Type:       models.EventTypeRegistration,
		ObjectType: models.EventObjectTypeDomain,
	}

	resp, err := client.Events.ListEventsPage(context.Background(), opts)
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, models.EventID("evt_1"), resp.Results[0].EventID)
}

func TestEventsService_GetEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/events/evt_1", r.URL.Path)
		registration := models.EventTypeRegistration
		_ = json.NewEncoder(w).Encode(models.Event{
			EventID:   "evt_1",
			Type:      &registration,
			EventData: models.EventData{Version: models.EventVersionV1, Message: "domain registered"},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	event, err := client.Events.GetEvent(context.Background(), "evt_1")
	require.NoError(t, err)
	assert.Equal(t, models.EventID("evt_1"), event.EventID)
	require.NotNil(t, event.Type)
	assert.Equal(t, models.EventTypeRegistration, *event.Type)
}

func TestEventsService_EscapesIDsInPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// An ID must stay one path segment: "/" and ".." must not reach
		// another route.
		assert.Equal(t, "/v1/events/evt_1%2F..%2F..%2Fjobs", r.RequestURI)
		_, _ = w.Write([]byte(`{"event_id":"evt_1","event_data":{"message":"m"}}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	_, err = client.Events.GetEvent(context.Background(), "evt_1/../../jobs")
	require.NoError(t, err)
}

// objectLogPage is a PageResponse[ObjectLog] body as the spec describes it.
func objectLogPage(logID string, hasNext bool, page int) string {
	return `{
		"results": [{
			"object_log_id": "` + logID + `",
			"object_id": "domain_01h45ytscbebyvny4gc8cr8ma2",
			"object_type": "DomainResponse",
			"action": "UPDATED",
			"created_on": "2026-01-23T08:26:55.123456Z",
			"performed_by_type": "user",
			"performed_by_id": "user_01h45ytscbebyvny4gc8cr8ma2",
			"server_request_id": "req_1",
			"changes": [{"field": "auto_renew", "old_value": false, "new_value": true}]
		}],
		"pagination": {
			"total_pages": 2, "total_items": 2, "current_page": ` + strconv.Itoa(page) + `,
			"page_size": 1, "has_next_page": ` + strconv.FormatBool(hasNext) + `, "has_previous_page": false
		}
	}`
}

func TestEventsService_ListObjectLogs(t *testing.T) {
	createdAfter := time.Date(2026, 1, 23, 8, 26, 55, 123456000, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/archive/object-logs", r.URL.Path)
		query := r.URL.Query()
		assert.Equal(t, "DomainResponse", query.Get("object_type"))
		assert.Equal(t, "domain_01h45ytscbebyvny4gc8cr8ma2", query.Get("object_id"))
		assert.Equal(t, "olog_1", query.Get("object_log_id"))
		assert.Equal(t, "UPDATED", query.Get("action"))
		assert.Equal(t, "req_1", query.Get("server_request_id"))
		assert.Equal(t, "user", query.Get("performed_by_type"))
		assert.Equal(t, "user_01h45ytscbebyvny4gc8cr8ma2", query.Get("performed_by_id"))
		assert.Equal(t, "2026-01-23T08:26:55.123456Z", query.Get("created_after"))
		assert.False(t, query.Has("user_id"), "the API has no user_id filter")

		_, _ = w.Write([]byte(objectLogPage("olog_1", false, 1)))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	opts := &models.ListObjectLogsOptions{
		ObjectType:      "DomainResponse",
		ObjectID:        "domain_01h45ytscbebyvny4gc8cr8ma2",
		ObjectLogID:     "olog_1",
		Action:          string(models.ObjectEventTypeUpdated),
		ServerRequestID: "req_1",
		PerformedByType: models.ExecutingEntityUser,
		PerformedByID:   "user_01h45ytscbebyvny4gc8cr8ma2",
		CreatedAfter:    &createdAfter,
	}

	resp, err := client.Events.ListObjectLogs(context.Background(), opts)
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "DomainResponse", resp.Results[0].ObjectType)
	assert.Equal(t, models.ObjectEventTypeUpdated, resp.Results[0].Action)
	require.Len(t, resp.Results[0].Changes, 1)
	assert.Equal(t, "auto_renew", resp.Results[0].Changes[0].Field)
}

func TestEventsService_ListObjectLogs_DeprecatedUserID(t *testing.T) {
	var got []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.False(t, r.URL.Query().Has("user_id"))
		got = append(got, r.URL.Query().Get("performed_by_id"))
		_, _ = w.Write([]byte(objectLogPage("olog_1", false, 1)))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	_, err = client.Events.ListObjectLogs(context.Background(), &models.ListObjectLogsOptions{UserID: "user_a"}) //nolint:staticcheck // the deprecated field must keep working
	require.NoError(t, err)
	_, err = client.Events.ListObjectLogs(context.Background(), &models.ListObjectLogsOptions{UserID: "user_a", PerformedByID: "user_b"}) //nolint:staticcheck // the deprecated field must keep working
	require.NoError(t, err)

	assert.Equal(t, []string{"user_a", "user_b"}, got)
}

func TestEventsService_ListObjectLogsByObject(t *testing.T) {
	var pages []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/archive/object-logs/domain_01h45ytscbebyvny4gc8cr8ma2", r.URL.Path)
		query := r.URL.Query()
		assert.Equal(t, "100", query.Get("page_size"))
		assert.Equal(t, "UPDATED", query.Get("action"))
		assert.Equal(t, "created_on", query.Get("sort_by"))
		assert.Equal(t, "desc", query.Get("sort_order"))
		assert.False(t, query.Has("object_id"), "the object ID is in the path")

		page := query.Get("page")
		pages = append(pages, page)
		if page == "1" {
			_, _ = w.Write([]byte(objectLogPage("olog_1", true, 1)))
			return
		}
		_, _ = w.Write([]byte(objectLogPage("olog_2", false, 2)))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	logs, err := client.Events.ListObjectLogsByObject(context.Background(), "domain_01h45ytscbebyvny4gc8cr8ma2", &models.ListObjectLogsOptions{
		ObjectID:  "ignored",
		Action:    string(models.ObjectEventTypeUpdated),
		SortBy:    "created_on",
		SortOrder: models.SortDesc,
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2"}, pages)
	require.Len(t, logs, 2)
	assert.Equal(t, "olog_1", logs[0].ObjectLogID)
	assert.Equal(t, "olog_2", logs[1].ObjectLogID)
}

func TestEventsService_GetObjectLog(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/archive/object-logs/domain_01h45ytscbebyvny4gc8cr8ma2", r.URL.Path)
		assert.Empty(t, r.URL.RawQuery)
		_, _ = w.Write([]byte(objectLogPage("olog_1", true, 1)))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Events.GetObjectLog(context.Background(), "domain_01h45ytscbebyvny4gc8cr8ma2") //nolint:staticcheck // the deprecated method must keep working
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.True(t, resp.Pagination.HasNextPage)
}

func TestEventsService_ListRequestHistory(t *testing.T) {
	startedAfter := time.Date(2026, 1, 23, 8, 26, 55, 500000000, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/archive/request-history", r.URL.Path)
		query := r.URL.Query()
		assert.Equal(t, "GET", query.Get("method"))
		assert.Equal(t, "/v1/domains", query.Get("path"))
		assert.Equal(t, "400", query.Get("min_status_code"))
		assert.Equal(t, "2026-01-23T08:26:55.5Z", query.Get("request_started_after"))

		_, _ = w.Write([]byte(`{
			"results": [{
				"server_request_id": "req_1",
				"method": "GET",
				"path": "/v1/domains",
				"status_code": 404,
				"duration": 12.5,
				"client_ip": "203.0.113.10",
				"performed_by_type": "user",
				"performed_by_id": "user_01h45ytscbebyvny4gc8cr8ma2",
				"request_started_at": "2026-01-23T08:26:55.512Z",
				"request_completed_at": "2026-01-23T08:26:55.524Z"
			}],
			"pagination": {"total_pages": 1, "total_items": 1, "current_page": 1, "page_size": 50, "has_next_page": false, "has_previous_page": false}
		}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	minStatus := 400
	opts := &models.ListRequestHistoryOptions{
		Method:              models.HTTPMethodGet,
		Path:                "/v1/domains",
		MinStatusCode:       &minStatus,
		RequestStartedAfter: &startedAfter,
	}

	resp, err := client.Events.ListRequestHistory(context.Background(), opts)
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, models.HTTPMethodGet, resp.Results[0].Method)
	assert.Equal(t, 404, resp.Results[0].StatusCode)
}

// emailForwardLogPage is a PageResponse[EmailForwardLog] body as the spec
// describes it.
func emailForwardLogPage(logID string, hasNext bool, page int) string {
	return `{
		"results": [{
			"log_id": "` + logID + `",
			"domain": "example.com",
			"sender_email": "sender@example.net",
			"recipient_email": "info@example.com",
			"forward_email": "dest@example.org",
			"subject": "Hello",
			"hostname": "mx1.example.com",
			"message_id": "<abc@example.net>",
			"transport": "mx",
			"final_status": "DELIVERED",
			"created_on": "2026-01-23T08:26:55Z",
			"synced_on": "2026-01-23T08:27:00Z",
			"events": []
		}],
		"pagination": {
			"total_pages": 2, "total_items": 2, "current_page": ` + strconv.Itoa(page) + `,
			"page_size": 1, "has_next_page": ` + strconv.FormatBool(hasNext) + `, "has_previous_page": false
		}
	}`
}

func TestEventsService_ListEmailForwardLogs(t *testing.T) {
	start := time.Date(2026, 1, 23, 0, 0, 0, 0, time.UTC)

	var pages []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/archive/email-forward-logs/email_forward_01h45ytscbebyvny4gc8cr8ma2", r.URL.Path)
		query := r.URL.Query()
		assert.Equal(t, "100", query.Get("page_size"))
		assert.Equal(t, "DELIVERED", query.Get("final_status"))
		assert.Equal(t, "created_on", query.Get("sort_by"))
		assert.Equal(t, "2026-01-23T00:00:00Z", query.Get("start_time"))
		assert.False(t, query.Has("email_forward_id"), "the email forward ID is in the path")

		page := query.Get("page")
		pages = append(pages, page)
		if page == "1" {
			_, _ = w.Write([]byte(emailForwardLogPage("log_1", true, 1)))
			return
		}
		_, _ = w.Write([]byte(emailForwardLogPage("log_2", false, 2)))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	logs, err := client.Events.ListEmailForwardLogs(context.Background(), "email_forward_01h45ytscbebyvny4gc8cr8ma2", &models.ListEmailForwardLogsOptions{
		FinalStatus: models.EmailForwardLogStatusDelivered,
		SortBy:      models.EmailForwardLogSortByCreatedOn,
		StartTime:   &start,
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2"}, pages)
	require.Len(t, logs, 2)
	assert.Equal(t, "log_1", logs[0].LogID)
	assert.Equal(t, "log_2", logs[1].LogID)
	assert.Equal(t, models.EmailForwardLogStatusDelivered, logs[0].FinalStatus)
}

func TestEventsService_ListEmailForwardLogsByAliasPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/archive/email-forward-logs/aliases/email_forward_alias_01h45ytscbebyvny4gc8cr8ma2", r.URL.Path)
		query := r.URL.Query()
		// email_forward_id is a required query parameter of this route.
		assert.Equal(t, "email_forward_01h45ytscbebyvny4gc8cr8ma2", query.Get("email_forward_id"))
		assert.Equal(t, "3", query.Get("page"))
		assert.Equal(t, "25", query.Get("page_size"))
		_, _ = w.Write([]byte(emailForwardLogPage("log_3", false, 3)))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Events.ListEmailForwardLogsByAliasPage(context.Background(),
		"email_forward_01h45ytscbebyvny4gc8cr8ma2", "email_forward_alias_01h45ytscbebyvny4gc8cr8ma2",
		&models.ListEmailForwardLogsOptions{Page: 3, PageSize: 25})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "log_3", resp.Results[0].LogID)
}

func TestEventsService_ListEmailForwardLogsByAlias(t *testing.T) {
	var pages []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/archive/email-forward-logs/aliases/email_forward_alias_01h45ytscbebyvny4gc8cr8ma2", r.URL.Path)
		query := r.URL.Query()
		assert.Equal(t, "email_forward_01h45ytscbebyvny4gc8cr8ma2", query.Get("email_forward_id"))

		page := query.Get("page")
		pages = append(pages, page)
		if page == "1" {
			_, _ = w.Write([]byte(emailForwardLogPage("log_1", true, 1)))
			return
		}
		_, _ = w.Write([]byte(emailForwardLogPage("log_2", false, 2)))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	logs, err := client.Events.ListEmailForwardLogsByAlias(context.Background(),
		"email_forward_01h45ytscbebyvny4gc8cr8ma2", "email_forward_alias_01h45ytscbebyvny4gc8cr8ma2", nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2"}, pages)
	assert.Len(t, logs, 2)
}
