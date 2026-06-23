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
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("page_size"))
		_ = json.NewEncoder(w).Encode(models.ContactAttributeSetListResponse{
			Results:    []models.ContactAttributeSet{{ContactAttributeSetID: "cas_1", TLD: "de"}},
			Pagination: models.Pagination{CurrentPage: 2, HasNextPage: false},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Contacts.ListContactAttributeSetsPage(context.Background(), &models.ListContactAttributeSetsOptions{
		Page:     2,
		PageSize: 10,
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
		_ = json.NewEncoder(w).Encode(models.ContactVerification{
			ContactID: "contact_123",
			Status:    "pending",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	verification, err := client.Contacts.RequestVerification(context.Background(), "contact_123")

	require.NoError(t, err)
	assert.Equal(t, models.ContactID("contact_123"), verification.ContactID)
	assert.Equal(t, models.EmailVerificationPending, verification.Status)
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
