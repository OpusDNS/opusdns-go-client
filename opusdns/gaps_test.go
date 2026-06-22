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

// --- Domain forward sets ---

func TestDomainForwardsService_GetDomainForwardSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domain-forwards/example.com/https", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.DomainForwardSetResponse{
			Hostname: "example.com",
			Protocol: models.HttpProtocolHTTPS,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	set, err := client.DomainForwards.GetDomainForwardSet(context.Background(), "example.com", models.HttpProtocolHTTPS)
	require.NoError(t, err)
	assert.Equal(t, models.HttpProtocolHTTPS, set.Protocol)
}

func TestDomainForwardsService_CreateDomainForwardSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/domain-forwards/example.com", r.URL.Path)

		var req models.DomainForwardSetCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, models.HttpProtocolHTTPS, req.Protocol)
		require.Len(t, req.Redirects, 1)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(models.DomainForwardSetResponse{Hostname: "example.com", Protocol: req.Protocol})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	set, err := client.DomainForwards.CreateDomainForwardSet(context.Background(), "example.com", &models.DomainForwardSetCreateRequest{
		Protocol: models.HttpProtocolHTTPS,
		Redirects: []models.HttpRedirectRequest{
			{RequestPath: "/", TargetProtocol: models.HttpProtocolHTTPS, TargetHostname: "dest.com", TargetPath: "/", RedirectCode: models.RedirectCodePermanent},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "example.com", set.Hostname)
}

func TestDomainForwardsService_PatchRedirects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/domain-forwards", r.URL.Path)

		var req models.DomainForwardPatchOps
		_ = json.NewDecoder(r.Body).Decode(&req)
		require.Len(t, req.Ops, 1)
		assert.Equal(t, models.PatchOpRemove, req.Ops[0].Op)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.DomainForwards.PatchRedirects(context.Background(), &models.DomainForwardPatchOps{
		Ops: []models.DomainForwardPatchOp{
			{Op: models.PatchOpRemove, Redirect: models.HttpRedirectRemove{RequestProtocol: models.HttpProtocolHTTPS, RequestHostname: "example.com", RequestPath: "/"}},
		},
	})
	require.NoError(t, err)
}

// --- Zone vanity set assignment ---

func TestDNSService_SetZoneVanitySet(t *testing.T) {
	t.Run("assigns set", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PATCH", r.Method)
			assert.Equal(t, "/v1/dns/example.com/vanity-set", r.URL.Path)

			var req models.ZoneVanitySetUpdateRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			require.NotNil(t, req.VanityNameserverSetID)
			assert.Equal(t, models.VanityNameserverSetID("vns_1"), *req.VanityNameserverSetID)

			_ = json.NewEncoder(w).Encode(map[string]models.Zone{"zone": {Name: "example.com"}})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		setID := models.VanityNameserverSetID("vns_1")
		zone, err := client.DNS.SetZoneVanitySet(context.Background(), "example.com.", &setID)
		require.NoError(t, err)
		assert.Equal(t, "example.com", zone.Name)
	})

	t.Run("clears set with null", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var raw map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&raw)
			val, present := raw["vanity_nameserver_set_id"]
			assert.True(t, present)
			assert.Nil(t, val)
			_ = json.NewEncoder(w).Encode(map[string]models.Zone{"zone": {Name: "example.com"}})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		_, err = client.DNS.SetZoneVanitySet(context.Background(), "example.com", nil)
		require.NoError(t, err)
	})
}

// --- Jobs retry ---

func TestJobsService_RetryBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/jobs/batch_1/retry", r.URL.Path)
		assert.Equal(t, []string{"BillingInsufficientFundsError"}, r.URL.Query()["error_class"])
		_ = json.NewEncoder(w).Encode(models.JobBatchRetryResponse{BatchID: "batch_1", RetriedCount: 3})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.Jobs.RetryBatch(context.Background(), "batch_1", []string{"BillingInsufficientFundsError"})
	require.NoError(t, err)
	assert.Equal(t, 3, result.RetriedCount)
}

func TestJobsService_RetryJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/job/job_1/retry", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.JobResponse{JobID: "job_1"})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	job, err := client.Jobs.RetryJob(context.Background(), "job_1")
	require.NoError(t, err)
	assert.Equal(t, models.JobID("job_1"), job.JobID)
}

// --- Current-org attributes ---

func TestOrganizationsService_GetCurrentAttributes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/attributes", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.OrganizationAttributesResponse{})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	_, err = client.Organizations.GetCurrentAttributes(context.Background())
	require.NoError(t, err)
}

func TestOrganizationsService_UpdateCurrentAttributes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/organizations/attributes", r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.OrganizationAttributesResponse{})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	_, err = client.Organizations.UpdateCurrentAttributes(context.Background(), &models.OrganizationAttributeUpdateRequest{})
	require.NoError(t, err)
}

// --- Contact attribute sets & attestation ---

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
