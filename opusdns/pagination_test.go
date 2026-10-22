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

// TestPaginationDecodesTotalItems decodes the API's PaginationMetadata. The
// API sends total_items and no total_count; the deprecated TotalCount alias
// is filled from it.
func TestPaginationDecodesTotalItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"results": [],
			"pagination": {"total_pages": 3, "current_page": 1, "has_next_page": true,
			               "has_previous_page": false, "total_items": 52, "page_size": 20}
		}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Contacts.ListContactsPage(context.Background(), &models.ListContactsOptions{Page: 1, PageSize: 20})
	require.NoError(t, err)

	p := resp.Pagination
	assert.Equal(t, 52, p.TotalItems)
	assert.Equal(t, 52, p.TotalCount) //nolint:staticcheck // the deprecated alias is what is under test
	assert.Equal(t, 3, p.TotalPages)
	assert.Equal(t, 1, p.CurrentPage)
	assert.Equal(t, 20, p.PageSize)
	assert.True(t, p.HasNextPage)
	assert.False(t, p.HasPreviousPage)
}

// TestPaginationEncodesWithoutTotalCount checks that the alias does not leak
// back into JSON the client writes.
func TestPaginationEncodesWithoutTotalCount(t *testing.T) {
	data, err := json.Marshal(models.Pagination{TotalItems: 7, TotalCount: 7}) //nolint:staticcheck // deprecated alias set on purpose
	require.NoError(t, err)

	var fields map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &fields))
	assert.Equal(t, float64(7), fields["total_items"])
	assert.NotContains(t, fields, "total_count")
}
