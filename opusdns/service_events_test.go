package opusdns

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestEventsService_ListObjectLogs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/archive/object-logs", r.URL.Path)
		query := r.URL.Query()
		assert.Equal(t, "DOMAIN", query.Get("object_type"))
		assert.Equal(t, "dom_1", query.Get("object_id"))

		_ = json.NewEncoder(w).Encode(models.ObjectLogListResponse{
			Results: []models.ObjectLog{
				{ObjectLogID: "olog_1", ObjectID: "dom_1", ObjectType: "domain", Action: models.ObjectEventTypeCreated},
			},
			Pagination: models.Pagination{HasNextPage: false, CurrentPage: 1},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	opts := &models.ListObjectLogsOptions{
		ObjectType: models.EventObjectTypeDomain,
		ObjectID:   "dom_1",
	}

	resp, err := client.Events.ListObjectLogs(context.Background(), opts)
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "dom_1", resp.Results[0].ObjectID)
	assert.Equal(t, models.ObjectEventTypeCreated, resp.Results[0].Action)
}

func TestEventsService_GetObjectLog(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/archive/object-logs/dom_1", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.ObjectLogListResponse{
			Results: []models.ObjectLog{
				{ObjectLogID: "olog_1", ObjectID: "dom_1", ObjectType: "domain", Action: models.ObjectEventTypeUpdated},
			},
			Pagination: models.Pagination{HasNextPage: false, CurrentPage: 1},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Events.GetObjectLog(context.Background(), "dom_1")
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "dom_1", resp.Results[0].ObjectID)
	assert.Equal(t, models.ObjectEventTypeUpdated, resp.Results[0].Action)
}

func TestEventsService_ListRequestHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/archive/request-history", r.URL.Path)
		query := r.URL.Query()
		assert.Equal(t, "GET", query.Get("method"))
		assert.Equal(t, "/v1/domains", query.Get("path"))

		_ = json.NewEncoder(w).Encode(models.RequestHistoryListResponse{
			Results: []models.RequestHistoryEntry{
				{
					ServerRequestID: "req_1",
					Method:          models.HTTPMethodGet,
					Path:            "/v1/domains",
					StatusCode:      200,
					Duration:        12.5,
					ClientIP:        "203.0.113.10",
				},
			},
			Pagination: models.Pagination{HasNextPage: false, CurrentPage: 1},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	opts := &models.ListOptions{
		Method: models.HTTPMethodGet,
		Path:   "/v1/domains",
	}

	resp, err := client.Events.ListRequestHistory(context.Background(), opts)
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, models.HTTPMethodGet, resp.Results[0].Method)
	assert.Equal(t, 200, resp.Results[0].StatusCode)
}

func TestEventsService_ListEmailForwardLogs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/archive/email-forward-logs/ef_1", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.EmailForwardLogListResponse{
			Results: []models.EmailForwardLog{
				{
					LogID:          "log_1",
					Domain:         "example.com",
					SenderEmail:    "sender@example.com",
					RecipientEmail: "alias@example.com",
					ForwardEmail:   "dest@example.com",
					FinalStatus:    models.EmailForwardLogStatusDelivered,
					CreatedOn:      time.Now(),
					SyncedOn:       time.Now(),
				},
			},
			Pagination: models.Pagination{HasNextPage: false, CurrentPage: 1},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Events.ListEmailForwardLogs(context.Background(), "ef_1")
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "log_1", resp.Results[0].LogID)
	assert.Equal(t, models.EmailForwardLogStatusDelivered, resp.Results[0].FinalStatus)
}

func TestEventsService_ListEmailForwardLogsByAlias(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/archive/email-forward-logs/aliases/alias_1", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.EmailForwardLogListResponse{
			Results: []models.EmailForwardLog{
				{
					LogID:          "log_2",
					Domain:         "example.com",
					SenderEmail:    "sender@example.com",
					RecipientEmail: "alias@example.com",
					ForwardEmail:   "dest@example.com",
					FinalStatus:    models.EmailForwardLogStatusQueued,
					CreatedOn:      time.Now(),
					SyncedOn:       time.Now(),
				},
			},
			Pagination: models.Pagination{HasNextPage: false, CurrentPage: 1},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Events.ListEmailForwardLogsByAlias(context.Background(), "alias_1")
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "log_2", resp.Results[0].LogID)
	assert.Equal(t, models.EmailForwardLogStatusQueued, resp.Results[0].FinalStatus)
}
