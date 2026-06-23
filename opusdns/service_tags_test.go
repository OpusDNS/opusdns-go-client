package opusdns

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagsService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tags", r.URL.Path)
		assert.Equal(t, []string{"DOMAIN"}, r.URL.Query()["tag_types"])

		_ = json.NewEncoder(w).Encode(models.TagListResponse{
			Results: []models.Tag{
				{TagID: "tag_123", Label: "Production", Type: models.TagTypeDomain, Color: models.TagColor1},
			},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	tags, err := client.Tags.ListTags(context.Background(), &models.ListTagsOptions{
		TagTypes: []models.TagType{models.TagTypeDomain},
	})

	require.NoError(t, err)
	require.Len(t, tags, 1)
	assert.Equal(t, "Production", tags[0].Label)
}

func TestTagsService_ListTagsPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tags", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "50", r.URL.Query().Get("page_size"))

		_ = json.NewEncoder(w).Encode(models.TagListResponse{
			Results: []models.Tag{
				{TagID: "tag_123", Label: "Production", Type: models.TagTypeDomain, Color: models.TagColor1},
			},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Tags.ListTagsPage(context.Background(), &models.ListTagsOptions{
		Page:     2,
		PageSize: 50,
	})

	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "Production", resp.Results[0].Label)
	assert.False(t, resp.Pagination.HasNextPage)
}

func TestTagsService_GetTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tags/tag_123", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.Tag{
			TagID: "tag_123",
			Label: "Production",
			Type:  models.TagTypeDomain,
			Color: models.TagColor1,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	tag, err := client.Tags.GetTag(context.Background(), "tag_123")

	require.NoError(t, err)
	assert.Equal(t, models.TagID("tag_123"), tag.TagID)
	assert.Equal(t, "Production", tag.Label)
}

func TestTagsService_CreateTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/tags", r.URL.Path)

		var req models.TagCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "Production", req.Label)
		assert.Equal(t, models.TagTypeDomain, req.Type)

		_ = json.NewEncoder(w).Encode(models.Tag{
			TagID: "tag_123",
			Label: req.Label,
			Type:  req.Type,
			Color: models.TagColor1,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	color := models.TagColor1
	tag, err := client.Tags.CreateTag(context.Background(), &models.TagCreateRequest{
		Label: "Production",
		Type:  models.TagTypeDomain,
		Color: &color,
	})

	require.NoError(t, err)
	assert.Equal(t, models.TagID("tag_123"), tag.TagID)
	assert.Equal(t, "Production", tag.Label)
}

func TestTagsService_UpdateTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/tags/tag_123", r.URL.Path)

		var req models.TagUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.NotNil(t, req.Label)
		assert.Equal(t, "Staging", *req.Label)

		_ = json.NewEncoder(w).Encode(models.Tag{
			TagID: "tag_123",
			Label: *req.Label,
			Type:  models.TagTypeDomain,
			Color: models.TagColor2,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	label := "Staging"
	tag, err := client.Tags.UpdateTag(context.Background(), "tag_123", &models.TagUpdateRequest{
		Label: &label,
	})

	require.NoError(t, err)
	assert.Equal(t, "Staging", tag.Label)
}

func TestTagsService_DeleteTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/tags/tag_123", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Tags.DeleteTag(context.Background(), "tag_123")
	require.NoError(t, err)
}

func TestTagsService_UpdateTagObjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/tags/tag_123/objects", r.URL.Path)

		var req models.ObjectTagChanges
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, []string{"example.com"}, req.Add)

		_ = json.NewEncoder(w).Encode(models.ObjectTagChangesResponse{
			Added:   1,
			Removed: 0,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Tags.UpdateTagObjects(context.Background(), "tag_123", &models.ObjectTagChanges{
		Add: []string{"example.com"},
	})

	require.NoError(t, err)
	assert.Equal(t, 1, resp.Added)
}

func TestTagsService_BulkUpdateObjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/tags/objects", r.URL.Path)

		var req models.BulkObjectTagChanges
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, models.TagTypeDomain, req.Type)
		assert.Equal(t, []string{"example.com"}, req.Objects)

		_ = json.NewEncoder(w).Encode(models.ObjectTagChangesResponse{
			Added:   2,
			Removed: 1,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Tags.BulkUpdateObjects(context.Background(), &models.BulkObjectTagChanges{
		Type:    models.TagTypeDomain,
		Objects: []string{"example.com"},
		Add:     []models.TagID{"tag_123"},
		Replace: &[]models.TagID{},
	})

	require.NoError(t, err)
	assert.Equal(t, 2, resp.Added)
	assert.Equal(t, 1, resp.Removed)
}
