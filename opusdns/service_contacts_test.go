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

func TestContactsService_VerifyContact(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/contacts/verify", r.URL.Path)
		assert.Equal(t, "verification-token", r.URL.Query().Get("token"))
		_ = json.NewEncoder(w).Encode(struct{}{})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Contacts.VerifyContact(context.Background(), &models.ContactVerificationRequest{Token: "verification-token"})

	require.NoError(t, err)
}

func TestContactsService_CreateContact(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/contacts", r.URL.Path)

		var req models.ContactCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "John", req.FirstName)
		assert.Equal(t, "Doe", req.LastName)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(models.Contact{
			ContactID: "contact_123",
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	contact, err := client.Contacts.CreateContact(context.Background(), &models.ContactCreateRequest{
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john@example.com",
		Phone:      "+1.5551234567",
		Street:     "123 Main St",
		City:       "New York",
		PostalCode: "10001",
		Country:    "US",
		Disclose:   false,
	})

	require.NoError(t, err)
	assert.Equal(t, "John", contact.FirstName)
	assert.Equal(t, "Doe", contact.LastName)
}

func TestContactsService_ContactAttributeSets(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/contacts/attribute-sets", r.URL.Path)
			_ = json.NewEncoder(w).Encode(models.ContactAttributeSetListResponse{
				Results:    []models.ContactAttributeSet{{ContactAttributeSetID: "cas_1", TLD: "de", Label: "DENIC"}},
				Pagination: models.Pagination{HasNextPage: false},
			})
		}))
		defer server.Close()

		client, _ := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		sets, err := client.Contacts.ListContactAttributeSets(context.Background(), nil)
		require.NoError(t, err)
		require.Len(t, sets, 1)
		assert.Equal(t, "de", sets[0].TLD)
	})

	t.Run("create", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/contacts/attribute-sets", r.URL.Path)

			var req models.ContactAttributeSetCreateRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "de", req.TLD)
			assert.Equal(t, "individual", req.Attributes["denic_type"])

			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(models.ContactAttributeSet{ContactAttributeSetID: "cas_1", TLD: req.TLD, Label: req.Label})
		}))
		defer server.Close()

		client, _ := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		set, err := client.Contacts.CreateContactAttributeSet(context.Background(), &models.ContactAttributeSetCreateRequest{
			Label:      "DENIC individual",
			TLD:        "de",
			Attributes: map[string]string{"denic_type": "individual"},
		})
		require.NoError(t, err)
		assert.Equal(t, models.ContactAttributeSetID("cas_1"), set.ContactAttributeSetID)
	})

	t.Run("get", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/contacts/attribute-sets/cas_1", r.URL.Path)
			_ = json.NewEncoder(w).Encode(models.ContactAttributeSet{ContactAttributeSetID: "cas_1"})
		}))
		defer server.Close()

		client, _ := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		_, err := client.Contacts.GetContactAttributeSet(context.Background(), "cas_1")
		require.NoError(t, err)
	})

	t.Run("update", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PATCH", r.Method)
			assert.Equal(t, "/v1/contacts/attribute-sets/cas_1", r.URL.Path)
			_ = json.NewEncoder(w).Encode(models.ContactAttributeSet{ContactAttributeSetID: "cas_1"})
		}))
		defer server.Close()

		client, _ := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		_, err := client.Contacts.UpdateContactAttributeSet(context.Background(), "cas_1", &models.ContactAttributeSetUpdateRequest{Label: models.StringPtr("x")})
		require.NoError(t, err)
	})

	t.Run("delete", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "DELETE", r.Method)
			assert.Equal(t, "/v1/contacts/attribute-sets/cas_1", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, _ := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		err := client.Contacts.DeleteContactAttributeSet(context.Background(), "cas_1")
		require.NoError(t, err)
	})

	t.Run("link", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PATCH", r.Method)
			assert.Equal(t, "/v1/contacts/ct_1/link/cas_1", r.URL.Path)
			_ = json.NewEncoder(w).Encode(models.ContactAttributeLink{ContactID: "ct_1", ContactAttributeSetID: "cas_1"})
		}))
		defer server.Close()

		client, _ := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		link, err := client.Contacts.LinkContactAttributeSet(context.Background(), "ct_1", "cas_1")
		require.NoError(t, err)
		assert.Equal(t, models.ContactID("ct_1"), link.ContactID)
	})
}

func TestContactsService_Attestation(t *testing.T) {
	t.Run("attest", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/contacts/ct_1/verifications/attest", r.URL.Path)

			var req models.ContactAttestRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			require.Len(t, req.Attestations, 1)
			assert.Equal(t, models.ContactVerificationClaimName, req.Attestations[0].Claim)

			_ = json.NewEncoder(w).Encode(models.ContactAttestResponse{
				Verifications: []models.ContactVerificationStatus{{Claim: models.ContactVerificationClaimName, State: models.ContactVerificationStateVerified}},
			})
		}))
		defer server.Close()

		client, _ := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		res, err := client.Contacts.AttestContactVerification(context.Background(), "ct_1", &models.ContactAttestRequest{
			Attestations: []models.ContactAttestVerificationRequest{
				{Claim: models.ContactVerificationClaimName, Method: models.ContactVerificationMethodAuth, Proof: models.ContactVerificationProofIDCard, AttestationReference: "ref"},
			},
		})
		require.NoError(t, err)
		require.Len(t, res.Verifications, 1)
		assert.Equal(t, models.ContactVerificationStateVerified, res.Verifications[0].State)
	})

	t.Run("get verifications", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/v1/contacts/ct_1/verifications", r.URL.Path)
			_ = json.NewEncoder(w).Encode(models.ContactAttestResponse{})
		}))
		defer server.Close()

		client, _ := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		_, err := client.Contacts.GetContactVerifications(context.Background(), "ct_1")
		require.NoError(t, err)
	})

	t.Run("cancel verification", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "DELETE", r.Method)
			assert.Equal(t, "/v1/contacts/ct_1/verification", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, _ := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		err := client.Contacts.CancelContactVerification(context.Background(), "ct_1")
		require.NoError(t, err)
	})
}

func TestContactsService_ListContactVerificationCases(t *testing.T) {
	var pages []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/contacts/contact_1/verification-cases", r.URL.Path)
		assert.Equal(t, "100", r.URL.Query().Get("page_size"))
		pages = append(pages, r.URL.Query().Get("page"))

		if r.URL.Query().Get("page") == "1" {
			_, _ = w.Write([]byte(`{"results":[{"verification_id":"domain_verification_2","source":"ICANN","trigger":"TRANSFER","status":"OPEN","claims":["email"],"deadlines":[{"type":"verification","date":"2026-11-12T00:00:00Z"}],"contact_id":"contact_1","opened_on":"2026-10-28T00:00:00Z","closed_on":null,"close_reason":null,"reminders_sent":0,"last_reminder_sent_on":null,"domain_id":"domain_2","domain_name":"example.org"}],"pagination":{"has_next_page":true}}`))
			return
		}
		_, _ = w.Write([]byte(`{"results":[{"verification_id":"domain_verification_1","source":"ICANN","trigger":"CREATE","status":"CANCELED","claims":["email"],"deadlines":[],"contact_id":"contact_1","opened_on":"2026-01-01T00:00:00Z","closed_on":"2026-02-01T00:00:00Z","close_reason":"DOMAIN_DELETED","reminders_sent":3,"last_reminder_sent_on":"2026-01-20T00:00:00Z","domain_id":"domain_1","domain_name":"example.com"}],"pagination":{"has_next_page":false}}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	cases, err := client.Contacts.ListContactVerificationCases(context.Background(), "contact_1", nil)

	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2"}, pages)
	require.Len(t, cases, 2)

	assert.Equal(t, models.DomainID("domain_2"), cases[0].DomainID)
	assert.Equal(t, "example.org", cases[0].DomainName)
	assert.Equal(t, models.DomainVerificationID("domain_verification_2"), cases[0].VerificationID)
	assert.Equal(t, models.DomainVerificationTriggerTransfer, cases[0].Trigger)
	assert.Nil(t, cases[0].CloseReason)

	assert.Equal(t, "example.com", cases[1].DomainName)
	assert.Equal(t, models.DomainVerificationStatusCanceled, cases[1].Status)
	require.NotNil(t, cases[1].CloseReason)
	assert.Equal(t, models.DomainVerificationCloseDomainDeleted, *cases[1].CloseReason)
	assert.Equal(t, 3, cases[1].RemindersSent)
}

func TestContactsService_ListContactVerificationCasesPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/contacts/contact_1/verification-cases", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("page_size"))
		_, _ = w.Write([]byte(`{"results":[],"pagination":{"total_pages":2,"current_page":2,"page_size":10,"has_next_page":false,"has_previous_page":true,"total_items":11}}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Contacts.ListContactVerificationCasesPage(context.Background(), "contact_1", &models.ListContactVerificationCasesOptions{
		Page:     2,
		PageSize: 10,
	})

	require.NoError(t, err)
	assert.Empty(t, resp.Results)
	assert.Equal(t, 2, resp.Pagination.CurrentPage)
	assert.Equal(t, 11, resp.Pagination.TotalItems)
}

func TestContactsService_ListContacts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/contacts", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.ContactListResponse{
			Results: []models.Contact{
				{ContactID: "contact_1", FirstName: "John", LastName: "Doe"},
			},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	contacts, err := client.Contacts.ListContacts(context.Background(), nil)

	require.NoError(t, err)
	require.Len(t, contacts, 1)
	assert.Equal(t, models.ContactID("contact_1"), contacts[0].ContactID)
}

func TestContactsService_ListContactsPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/contacts", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("page_size"))
		assert.Equal(t, "john@example.com", r.URL.Query().Get("email"))
		_ = json.NewEncoder(w).Encode(models.ContactListResponse{
			Results:    []models.Contact{{ContactID: "contact_1"}},
			Pagination: models.Pagination{CurrentPage: 2, HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Contacts.ListContactsPage(context.Background(), &models.ListContactsOptions{
		Page:     2,
		PageSize: 10,
		Email:    "john@example.com",
	})

	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, 2, resp.Pagination.CurrentPage)
}

func TestContactsService_ListContactAttributeSetsPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/contacts/attribute-sets", r.URL.Path)
		query := r.URL.Query()
		assert.Equal(t, "2", query.Get("page"))
		assert.Equal(t, "10", query.Get("page_size"))
		assert.Equal(t, "label", query.Get("sort_by"))
		assert.Equal(t, "asc", query.Get("sort_order"))
		assert.Equal(t, "de", query.Get("tld"))
		assert.Equal(t, "DENIC", query.Get("label"))
		_ = json.NewEncoder(w).Encode(models.ContactAttributeSetListResponse{
			Results:    []models.ContactAttributeSet{{ContactAttributeSetID: "cas_1", TLD: "de"}},
			Pagination: models.Pagination{CurrentPage: 2, HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Contacts.ListContactAttributeSetsPage(context.Background(), &models.ListContactAttributeSetsOptions{
		Page:      2,
		PageSize:  10,
		SortBy:    models.ContactAttributeSetSortByLabel,
		SortOrder: models.SortAsc,
		TLD:       "de",
		Label:     "DENIC",
	})

	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, 2, resp.Pagination.CurrentPage)
}

func TestContactsService_GetContact(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/contacts/contact_123", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.Contact{
			ContactID: "contact_123",
			FirstName: "Jane",
			LastName:  "Smith",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	contact, err := client.Contacts.GetContact(context.Background(), "contact_123")

	require.NoError(t, err)
	assert.Equal(t, models.ContactID("contact_123"), contact.ContactID)
	assert.Equal(t, "Jane", contact.FirstName)
}

func TestContactsService_GetContactWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/contacts/contact_123", r.URL.Path)
		assert.Equal(t, []string{"tags"}, r.URL.Query()["include"])
		_ = json.NewEncoder(w).Encode(models.Contact{ContactID: "contact_123"})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	contact, err := client.Contacts.GetContactWithOptions(context.Background(), "contact_123", &models.GetContactOptions{
		Include: []models.ContactIncludeField{models.ContactIncludeTags},
	})
	require.NoError(t, err)
	assert.Equal(t, models.ContactID("contact_123"), contact.ContactID)
}

func TestContactsService_ContactIDIsEscaped(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	// The ID stays one path segment, so it cannot reach another resource.
	err = client.Contacts.DeleteContact(context.Background(), "../domains/example.com")
	require.NoError(t, err)
	assert.Equal(t, "/v1/contacts/..%2Fdomains%2Fexample.com", gotPath)
}

func TestContactsService_DeleteContact(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/contacts/contact_123", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Contacts.DeleteContact(context.Background(), "contact_123")

	require.NoError(t, err)
}

func TestContactsService_RequestVerification(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/contacts/contact_123/verification", r.URL.Path)
		// type is a required query parameter.
		assert.Equal(t, "api", r.URL.Query().Get("type"))

		w.WriteHeader(http.StatusCreated)
		// ContactVerificationApiResponse
		_, _ = w.Write([]byte(`{
			"contact_verification_id": "contact_verification_01h45ytscbebyvny4gc8cr8ma2",
			"contact_id": "contact_123",
			"type": "api",
			"status": "pending",
			"token": "tok_123",
			"verified_on": null,
			"canceled_on": null,
			"created_on": "2026-01-23T08:26:55Z",
			"updated_on": "2026-01-23T08:26:55Z"
		}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	verification, err := client.Contacts.RequestVerification(context.Background(), "contact_123", models.VerificationTypeAPI)

	require.NoError(t, err)
	assert.Equal(t, models.ContactID("contact_123"), verification.ContactID)
	assert.Equal(t, models.VerificationTypeAPI, verification.Type)
	assert.Equal(t, models.EmailVerificationPending, verification.Status)
	require.NotNil(t, verification.Token)
	assert.Equal(t, "tok_123", *verification.Token)
}

func TestContactsService_RequestVerification_RequiresType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no request expected, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	_, err = client.Contacts.RequestVerification(context.Background(), "contact_123", "")
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestContactsService_CreateContactAttributeSet_RequiresAttributes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no request expected, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	// A nil map would go out as "attributes": null.
	_, err = client.Contacts.CreateContactAttributeSet(context.Background(), &models.ContactAttributeSetCreateRequest{
		Label: "DENIC", TLD: "de",
	})
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestContactsService_GetVerificationStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/contacts/contact_123/verification", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.ContactVerification{
			ContactID: "contact_123",
			Status:    "verified",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	verification, err := client.Contacts.GetVerificationStatus(context.Background(), "contact_123")

	require.NoError(t, err)
	assert.Equal(t, models.ContactID("contact_123"), verification.ContactID)
	assert.Equal(t, models.EmailVerificationVerified, verification.Status)
}

func TestContactsService_ListContactsFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		assert.Equal(t, "true", query.Get("in_use"))
		assert.Equal(t, []string{"VERIFICATION_REQUIRED"}, query["status_tags"])
		assert.Equal(t, "match_any", query.Get("status_tag_mode"))
		assert.Equal(t, "false", query.Get("has_tags"))
		// Sub-second precision survives, so paging by a returned timestamp
		// neither skips nor repeats contacts created within one second.
		assert.Equal(t, "2026-01-23T08:26:55.123456Z", query.Get("created_after"))
		assert.Equal(t, "2026-01-24T00:00:00Z", query.Get("created_before"))

		_ = json.NewEncoder(w).Encode(models.ContactListResponse{
			Results:    []models.Contact{},
			Pagination: models.Pagination{HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	inUse := true
	hasTags := false
	createdAfter := time.Date(2026, 1, 23, 8, 26, 55, 123456000, time.UTC)
	createdBefore := time.Date(2026, 1, 24, 0, 0, 0, 0, time.UTC)
	_, err = client.Contacts.ListContacts(context.Background(), &models.ListContactsOptions{
		InUse:         &inUse,
		StatusTags:    []models.StatusTagType{models.StatusTagTypeVerificationRequired},
		StatusTagMode: models.TagFilterModeMatchAny,
		HasTags:       &hasTags,
		CreatedAfter:  &createdAfter,
		CreatedBefore: &createdBefore,
	})
	require.NoError(t, err)
}
