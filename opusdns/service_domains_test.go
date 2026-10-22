package opusdns

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomainsService_DNSSEC(t *testing.T) {
	t.Run("get returns dnssec data array", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/v1/domains/example.com/dnssec", r.URL.Path)
			_ = json.NewEncoder(w).Encode([]models.DomainDNSSECDataResponse{
				{RecordType: models.DNSSECRecordTypeDSData, Algorithm: 13},
			})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		data, err := client.Domains.GetDNSSEC(context.Background(), "example.com")

		require.NoError(t, err)
		require.Len(t, data, 1)
		assert.Equal(t, models.DNSSECRecordTypeDSData, data[0].RecordType)
	})

	t.Run("put replaces dnssec data", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PUT", r.Method)
			assert.Equal(t, "/v1/domains/example.com/dnssec", r.URL.Path)

			var req []models.DomainDNSSECDataCreate
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			require.Len(t, req, 1)
			assert.Equal(t, models.DNSSECRecordTypeKeyData, req[0].RecordType)

			_ = json.NewEncoder(w).Encode([]models.DomainDNSSECDataResponse{
				{RecordType: req[0].RecordType, Algorithm: req[0].Algorithm},
			})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		data, err := client.Domains.PutDNSSEC(context.Background(), "example.com", []models.DomainDNSSECDataCreate{
			{RecordType: models.DNSSECRecordTypeKeyData, Algorithm: 13},
		})

		require.NoError(t, err)
		require.Len(t, data, 1)
	})

	t.Run("disable expects no content", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/domains/example.com/dnssec/disable", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		err = client.Domains.DisableDNSSEC(context.Background(), "example.com")

		require.NoError(t, err)
	})
}

func TestDomainsService_ListDomains(t *testing.T) {
	t.Run("returns domains", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/domains")

			_ = json.NewEncoder(w).Encode(models.DomainListResponse{
				Results: []models.Domain{
					{DomainID: "domain_123", Name: "example.com"},
				},
				Pagination: models.Pagination{HasNextPage: false},
			})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		domains, err := client.Domains.ListDomains(context.Background(), nil)

		require.NoError(t, err)
		assert.Len(t, domains, 1)
		assert.Equal(t, "example.com", domains[0].Name)
	})

	t.Run("sends documented filters", func(t *testing.T) {
		now := time.Date(2026, 5, 5, 8, 0, 0, 0, time.UTC)
		trueValue := true
		falseValue := false

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			assert.Equal(t, []string{"tag_123", "tag_456"}, query["tag_ids"])
			assert.Equal(t, "match_all", query.Get("tag_mode"))
			assert.Equal(t, "true", query.Get("is_premium"))
			assert.Equal(t, now.Format(time.RFC3339), query.Get("updated_after"))
			assert.Equal(t, "true", query.Get("expires_in_30_days"))
			assert.Equal(t, now.Format(time.RFC3339), query.Get("registered_after"))
			assert.Equal(t, []string{"ok", "clientTransferProhibited"}, query["registry_statuses"])
			assert.Equal(t, []string{"tags"}, query["include"])
			assert.Equal(t, []string{"EXTERNAL", "DNSSEC_PENDING"}, query["status_tags"])
			assert.Equal(t, "match_none", query.Get("status_tag_mode"))
			assert.Equal(t, "false", query.Get("read_only"))
			assert.Equal(t, now.Format(time.RFC3339), query.Get("transferred_after"))
			assert.Equal(t, now.Format(time.RFC3339), query.Get("transferred_before"))
			assert.Equal(t, "true", query.Get("has_tags"))
			assert.Equal(t, "ns1.example.net", query.Get("nameserver"))
			assert.Equal(t, []string{"ams", "ras"}, query["account_type"])
			assert.Equal(t, "match_none", query.Get("account_type_mode"))
			assert.Equal(t, []string{"INTERNETX"}, query["registrar"])
			assert.Equal(t, []string{"registrar_credential_1"}, query["registrar_credential_id"])

			_ = json.NewEncoder(w).Encode(models.DomainListResponse{
				Results:    []models.Domain{},
				Pagination: models.Pagination{HasNextPage: false},
			})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		_, err = client.Domains.ListDomains(context.Background(), &models.ListDomainsOptions{
			TagIDs:          []models.TagID{"tag_123", "tag_456"},
			TagMode:         models.TagFilterModeMatchAll,
			IsPremium:       &trueValue,
			UpdatedAfter:    &now,
			ExpiresIn30Days: &trueValue,
			RegisteredAfter: &now,
			RegistryStatuses: []string{
				"ok",
				"clientTransferProhibited",
			},
			Include: []models.DomainIncludeField{models.DomainIncludeTags},
			StatusTags: []models.StatusTagType{
				models.StatusTagTypeExternal,
				models.StatusTagTypeDNSSECPending,
			},
			StatusTagMode:     models.TagFilterModeMatchNone,
			ReadOnly:          &falseValue,
			TransferredAfter:  &now,
			TransferredBefore: &now,
			HasTags:           &trueValue,
			Nameserver:        "ns1.example.net",
			AccountTypes: []models.ConnectedAccountType{
				models.ConnectedAccountTypeAMS,
				models.ConnectedAccountTypeRAS,
			},
			AccountTypeMode:        models.AccountTypeFilterMatchNone,
			Registrars:             []models.Registrar{models.RegistrarInternetX},
			RegistrarCredentialIDs: []models.TypeID{"registrar_credential_1"},
		})

		require.NoError(t, err)
	})
}

func TestDomainsService_CancelTransfer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/domains/example.com/transfer", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Domains.CancelTransfer(context.Background(), "example.com")
	require.NoError(t, err)
}

func TestDomainsService_ListDomainsPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domains", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "50", r.URL.Query().Get("page_size"))

		_ = json.NewEncoder(w).Encode(models.DomainListResponse{
			Results: []models.Domain{
				{DomainID: "domain_456", Name: "second.com"},
			},
			Pagination: models.Pagination{HasNextPage: true},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Domains.ListDomainsPage(context.Background(), &models.ListDomainsOptions{
		Page:     2,
		PageSize: 50,
	})

	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "second.com", resp.Results[0].Name)
	assert.True(t, resp.Pagination.HasNextPage)
}

func TestDomainsService_GetDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domains/example.com", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.Domain{
			DomainID: "domain_123",
			Name:     "example.com",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	domain, err := client.Domains.GetDomain(context.Background(), "example.com")

	require.NoError(t, err)
	assert.Equal(t, "example.com", domain.Name)
	assert.Equal(t, models.DomainID("domain_123"), domain.DomainID)
}

func TestDomainsService_GetDomainWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domains/example.com", r.URL.Path)
		assert.Equal(t, []string{"tags", "connected_account", "registrant_change"}, r.URL.Query()["include"])

		_, _ = w.Write([]byte(`{
			"domain_id": "domain_123",
			"name": "example.com",
			"connected_account": {"type": "ras", "registrar_credential_id": "registrar_credential_1", "name": "My InterNetX", "registrar": "INTERNETX"},
			"registrant_change": {
				"registrant_change_id": "registrant_change_1",
				"domain_id": "domain_123",
				"status": "completed",
				"material": true,
				"previous_registrant": "contact_1",
				"new_registrant": "contact_2",
				"confirmation_method": "email",
				"transfer_lock_opt_out": false,
				"transfer_lock_until": "2026-12-01T00:00:00Z",
				"confirmation_deadline": null,
				"failure_reason": null,
				"created_on": "2026-10-01T00:00:00Z"
			}
		}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	domain, err := client.Domains.GetDomainWithOptions(context.Background(), "example.com", &models.GetDomainOptions{
		Include: []models.DomainIncludeField{
			models.DomainIncludeTags,
			models.DomainIncludeConnectedAccount,
			models.DomainIncludeRegistrantChange,
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "example.com", domain.Name)

	require.NotNil(t, domain.ConnectedAccount)
	assert.Equal(t, models.ConnectedAccountTypeRAS, domain.ConnectedAccount.Type)
	assert.Equal(t, models.RegistrarInternetX, domain.ConnectedAccount.Registrar)
	assert.Equal(t, "My InterNetX", domain.ConnectedAccount.Name)

	require.NotNil(t, domain.RegistrantChange)
	assert.Equal(t, models.RegistrantChangeStatusCompleted, domain.RegistrantChange.Status)
	assert.Equal(t, models.ContactID("contact_2"), domain.RegistrantChange.NewRegistrant)
	require.NotNil(t, domain.RegistrantChange.ConfirmationMethod)
	assert.Equal(t, models.RegistrantChangeConfirmationEmail, *domain.RegistrantChange.ConfirmationMethod)
	require.NotNil(t, domain.RegistrantChange.TransferLockUntil)
}

func TestDomainsService_CreateDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/domains", r.URL.Path)

		// The body must have the shape of the spec's DomainCreate.
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{
			"name": "example.it",
			"contacts": {"registrant": [{"contact_id": "contact_01h45ytscbebyvny4gc8cr8ma2"}]},
			"renewal_mode": "renew",
			"period": {"value": 1, "unit": "y"},
			"claims_notice_acceptance_hash": "a1b2c3",
			"expected_price": "1250.00",
			"attributes": {"nic_it_compliance_confirmation": "true", "auto_renew_period": "monthly"}
		}`, string(body))

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"domain_id":"domain_01h45ytscbebyvny4gc8cr8ma2","name":"example.it","sld":"example","tld":"it","roid":"D1-IT","renewal_mode":"renew","registry_statuses":["ok"]}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	domain, err := client.Domains.CreateDomain(context.Background(), &models.DomainCreateRequest{
		Name:        "example.it",
		RenewalMode: models.RenewalModeRenew,
		Period:      models.DomainPeriod{Value: 1, Unit: models.PeriodUnitYear},
		Contacts: map[models.DomainContactType][]models.ContactHandle{
			models.DomainContactTypeRegistrant: {{ContactID: "contact_01h45ytscbebyvny4gc8cr8ma2"}},
		},
		ClaimsNoticeAcceptanceHash: models.StringPtr("a1b2c3"),
		ExpectedPrice:              models.StringPtr("1250.00"),
		Attributes: map[models.DomainAttributeKey]string{
			models.DomainAttributeNicITComplianceConfirmation: "true",
			models.DomainAttributeAutoRenewPeriod:             models.AutoRenewPeriodMonthly,
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "example.it", domain.Name)
}

func TestDomainsService_UpdateDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/domains/example.dk", r.URL.Path)

		// The body must have the shape of the spec's DomainUpdate.
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{
			"statuses": ["clientUpdateProhibited"],
			"auth_code": "n3w-c0de",
			"attributes": {"punktum_dk_terms_acceptance": "true", "auto_renew_period": "yearly"}
		}`, string(body))

		_, _ = w.Write([]byte(`{"domain_id":"domain_01h45ytscbebyvny4gc8cr8ma2","name":"example.dk","sld":"example","tld":"dk","roid":"D1-DK"}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	domain, err := client.Domains.UpdateDomain(context.Background(), "example.dk", &models.DomainUpdateRequest{
		Statuses: []models.DomainClientStatus{models.DomainClientStatusUpdateProhibited},
		AuthCode: models.StringPtr("n3w-c0de"),
		Attributes: map[models.DomainAttributeKey]string{
			models.DomainAttributePunktumDKTermsAcceptance: "true",
			models.DomainAttributeAutoRenewPeriod:          models.AutoRenewPeriodYearly,
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "example.dk", domain.Name)
}

func TestDomainsService_DeleteDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/domains/example.com", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Domains.DeleteDomain(context.Background(), "example.com")
	require.NoError(t, err)
}

func TestDomainsService_TransferDomain(t *testing.T) {
	t.Run("full request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/domains/transfer", r.URL.Path)

			// The body must have the shape of the spec's DomainTransferIn:
			// period is a DomainPeriod object, not a number.
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			assert.JSONEq(t, `{
				"name": "example.it",
				"auth_code": "auth123",
				"renewal_mode": "renew",
				"period": {"value": 1, "unit": "y"},
				"create_zone": true,
				"expected_price": "12.50",
				"attributes": {"nic_it_compliance_confirmation": "true"}
			}`, string(body))

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"domain_id":"domain_01h45ytscbebyvny4gc8cr8ma2","name":"example.it","sld":"example","tld":"it","roid":"D1-IT"}`))
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		domain, err := client.Domains.TransferDomain(context.Background(), &models.DomainTransferRequest{
			Name:          "example.it",
			AuthCode:      "auth123",
			RenewalMode:   models.RenewalModeRenew,
			Period:        &models.DomainPeriod{Value: 1, Unit: models.PeriodUnitYear},
			CreateZone:    true,
			ExpectedPrice: models.StringPtr("12.50"),
			Attributes: map[models.DomainAttributeKey]string{
				models.DomainAttributeNicITComplianceConfirmation: "true",
			},
		})

		require.NoError(t, err)
		assert.Equal(t, "example.it", domain.Name)
	})

	t.Run("no auth code is not sent as an empty string", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The spec allows null for auth_code but requires at least 6
			// characters when it is a string.
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			assert.JSONEq(t, `{"name": "example.com", "renewal_mode": "renew"}`, string(body))

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"domain_id":"domain_01h45ytscbebyvny4gc8cr8ma2","name":"example.com"}`))
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		_, err = client.Domains.TransferDomain(context.Background(), &models.DomainTransferRequest{
			Name:        "example.com",
			RenewalMode: models.RenewalModeRenew,
		})
		require.NoError(t, err)
	})
}

func TestDomainsService_RenewDomain(t *testing.T) {
	expiresOn := time.Date(2027, 5, 1, 10, 15, 30, 123456000, time.UTC)

	t.Run("renewed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/domains/example.com/renew", r.URL.Path)

			// The body must have the shape of the spec's DomainRenewRequest:
			// period is a DomainPeriod object and current_expiry_date is
			// required.
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			assert.JSONEq(t, `{
				"period": {"value": 2, "unit": "y"},
				"current_expiry_date": "2027-05-01T10:15:30.123456Z",
				"expected_price": "99.00"
			}`, string(body))

			_, _ = w.Write([]byte(`{"name":"example.com","new_expiry_date":"2029-05-01T10:15:30Z","period_extended":{"value":2,"unit":"y"}}`))
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		result, err := client.Domains.RenewDomain(context.Background(), "example.com", &models.DomainRenewRequest{
			Period:            models.DomainPeriod{Value: 2, Unit: models.PeriodUnitYear},
			CurrentExpiryDate: &expiresOn,
			ExpectedPrice:     models.StringPtr("99.00"),
		})

		require.NoError(t, err)
		assert.Equal(t, "example.com", result.Name)
		assert.False(t, result.TransferPending())
		require.NotNil(t, result.NewExpiryDate)
		assert.Equal(t, time.Date(2029, 5, 1, 10, 15, 30, 0, time.UTC), result.NewExpiryDate.UTC())
		require.NotNil(t, result.PeriodExtended)
		assert.Equal(t, models.DomainPeriod{Value: 2, Unit: models.PeriodUnitYear}, *result.PeriodExtended)
	})

	t.Run("transfer pending", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"name":"example.com","request_id":"req_123","status":"transfer_pending"}`))
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		result, err := client.Domains.RenewDomain(context.Background(), "example.com", &models.DomainRenewRequest{
			Period:            models.DomainPeriod{Value: 1, Unit: models.PeriodUnitYear},
			CurrentExpiryDate: &expiresOn,
		})

		require.NoError(t, err)
		assert.True(t, result.TransferPending())
		assert.Equal(t, "req_123", result.RequestID)
		assert.Nil(t, result.NewExpiryDate)
	})

	t.Run("incomplete requests are rejected before sending", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		for name, req := range map[string]*models.DomainRenewRequest{
			"nil request":       nil,
			"no period":         {CurrentExpiryDate: &expiresOn},
			"no current expiry": {Period: models.DomainPeriod{Value: 1, Unit: models.PeriodUnitYear}},
			"zero period value": {Period: models.DomainPeriod{Unit: models.PeriodUnitYear}, CurrentExpiryDate: &expiresOn},
			"negative period":   {Period: models.DomainPeriod{Value: -1, Unit: models.PeriodUnitYear}, CurrentExpiryDate: &expiresOn},
		} {
			_, err := client.Domains.RenewDomain(context.Background(), "example.com", req)
			assert.ErrorIs(t, err, ErrInvalidInput, name)
		}
	})
}

func TestDomainsService_RequestDomainAuthCode(t *testing.T) {
	t.Run("in the response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/domains/example.com/auth-code", r.URL.Path)

			_, _ = w.Write([]byte(`{"name":"example.com","delivery":"response","auth_code":"s3cret","auth_code_expires_on":"2026-11-01T00:00:00Z"}`))
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		result, err := client.Domains.RequestDomainAuthCode(context.Background(), "example.com")

		require.NoError(t, err)
		assert.Equal(t, models.AuthCodeDeliveryResponse, result.Delivery)
		require.NotNil(t, result.AuthCode)
		assert.Equal(t, "s3cret", *result.AuthCode)
		require.NotNil(t, result.AuthCodeExpiresOn)
	})

	t.Run("emailed to the registrant", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/domains/example.se/auth-code", r.URL.Path)

			_, _ = w.Write([]byte(`{"name":"example.se","delivery":"registrant_email","auth_code":null,"recipients":["owner@example.se"]}`))
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		result, err := client.Domains.RequestDomainAuthCode(context.Background(), "example.se")

		require.NoError(t, err)
		assert.Equal(t, models.AuthCodeDeliveryRegistrantEmail, result.Delivery)
		assert.Nil(t, result.AuthCode)
		assert.Equal(t, []string{"owner@example.se"}, result.Recipients)
	})
}

func TestDomainsService_ListDomainVerifications(t *testing.T) {
	var pages []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domains/example.com/verifications", r.URL.Path)
		pages = append(pages, r.URL.Query().Get("page"))

		if r.URL.Query().Get("page") == "1" {
			_, _ = w.Write([]byte(`{"results":[{"verification_id":"domain_verification_1","source":"ICANN","trigger":"CREATE","status":"SATISFIED","claims":["email"],"deadlines":[],"contact_id":"contact_1","opened_on":"2026-01-01T00:00:00Z","closed_on":"2026-01-02T00:00:00Z","close_reason":"EVIDENCE_VERIFIED"}],"pagination":{"has_next_page":true}}`))
			return
		}
		_, _ = w.Write([]byte(`{"results":[{"verification_id":"domain_verification_2","source":"REGISTRY","trigger":"REGISTRY_REQUEST","status":"OPEN","claims":["legal_entity"],"deadlines":[{"type":"verification","date":"2026-12-01T00:00:00Z"}],"contact_id":"contact_1","opened_on":"2026-10-01T00:00:00Z","closed_on":null,"close_reason":null}],"pagination":{"has_next_page":false}}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	cases, err := client.Domains.ListDomainVerifications(context.Background(), "example.com", nil)

	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2"}, pages)
	require.Len(t, cases, 2)

	assert.Equal(t, models.DomainVerificationKindICANN, cases[0].Source)
	assert.Equal(t, models.DomainVerificationStatusSatisfied, cases[0].Status)
	require.NotNil(t, cases[0].CloseReason)
	assert.Equal(t, models.DomainVerificationCloseEvidenceVerified, *cases[0].CloseReason)

	assert.Equal(t, models.DomainVerificationTriggerRegistryRequest, cases[1].Trigger)
	assert.Equal(t, []models.VerificationClaimType{models.VerificationClaimTypeLegalEntity}, cases[1].Claims)
	require.Len(t, cases[1].Deadlines, 1)
	assert.Equal(t, models.VerificationDeadlineTypeVerification, cases[1].Deadlines[0].Type)
	assert.Nil(t, cases[1].ClosedOn)
	assert.Nil(t, cases[1].CloseReason)
}

func TestDomainsService_RestoreDomain(t *testing.T) {
	tests := []struct {
		name     string
		req      *models.DomainRestoreRequest
		wantBody string
	}{
		{
			name: "full request",
			req: &models.DomainRestoreRequest{
				Reason:         models.StringPtr("deleted by mistake"),
				AdditionalInfo: models.StringPtr("ticket 4711"),
				ExpectedPrice:  models.StringPtr("45.00"),
			},
			wantBody: `{"reason":"deleted by mistake","additional_info":"ticket 4711","expected_price":"45.00"}`,
		},
		{
			// The deprecated Period is not part of the spec's
			// DomainRestoreRequest and must not be sent.
			name:     "deprecated period is not sent",
			req:      &models.DomainRestoreRequest{Period: 1}, //nolint:staticcheck // checks that the deprecated field is not sent
			wantBody: `{}`,
		},
		{
			// The API requires a body, so nil is sent as an empty object.
			name:     "nil request",
			req:      nil,
			wantBody: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/domains/example.com/restore", r.URL.Path)

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tt.wantBody, string(body))

				_, _ = w.Write([]byte(`{"domain_id":"domain_123","name":"example.com","restored_at":"2026-10-22T09:30:00Z"}`))
			}))
			defer server.Close()

			client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
			require.NoError(t, err)

			result, err := client.Domains.RestoreDomain(context.Background(), "example.com", tt.req)

			require.NoError(t, err)
			assert.Equal(t, models.DomainID("domain_123"), result.DomainID)
			assert.Equal(t, "example.com", result.Name)
			require.NotNil(t, result.RestoredAt)
			assert.Equal(t, time.Date(2026, 10, 22, 9, 30, 0, 0, time.UTC), result.RestoredAt.UTC())
		})
	}
}

func TestDomainsService_GetSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domains/summary", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.DomainSummary{
			OrganizationID: "organization_123",
			Domains: models.DomainSummaryData{
				TotalCount:  42,
				ByStatus:    map[models.DomainStatus]int{models.DomainStatusOK: 40},
				ByStatusTag: map[models.StatusTagType]int{models.StatusTagTypeDNSSECPending: 2},
				ByTLD:       map[string]int{"com": 42},
				ByAccountType: map[models.ConnectedAccountType]int{
					models.ConnectedAccountTypeDefault: 40,
					models.ConnectedAccountTypeRAS:     2,
				},
				ExpiringSoon: &models.DomainsExpiringSoon{Next30Days: 3, Next60Days: 5, Next90Days: 9},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	summary, err := client.Domains.GetSummary(context.Background())

	require.NoError(t, err)
	assert.Equal(t, models.OrganizationID("organization_123"), summary.OrganizationID)
	assert.Equal(t, 42, summary.Domains.TotalCount)
	assert.Equal(t, 40, summary.Domains.ByStatus[models.DomainStatusOK])
	assert.Equal(t, 2, summary.Domains.ByStatusTag[models.StatusTagTypeDNSSECPending])
	assert.Equal(t, 2, summary.Domains.ByAccountType[models.ConnectedAccountTypeRAS])
	require.NotNil(t, summary.Domains.ExpiringSoon)
	assert.Equal(t, 3, summary.Domains.ExpiringSoon.Next30Days)
}

func TestDomainsService_CheckDomains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domains/check", r.URL.Path)
		assert.Equal(t, []string{"example.com", "example.net"}, r.URL.Query()["domains"])

		_ = json.NewEncoder(w).Encode(models.DomainCheckResponse{
			Results: []models.DomainAvailabilityResult{
				{Domain: "example.com", Available: false},
				{Domain: "example.net", Available: true},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	resp, err := client.Domains.CheckDomains(context.Background(), []string{"example.com", "example.net"})

	require.NoError(t, err)
	require.Len(t, resp.Results, 2)
	assert.True(t, resp.Results[1].Available)
}

func TestDomainsService_DeleteDNSSEC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/domains/example.com/dnssec", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	err = client.Domains.DeleteDNSSEC(context.Background(), "example.com")
	require.NoError(t, err)
}

func TestDomainsService_EnableDNSSEC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/domains/example.com/dnssec/enable", r.URL.Path)

		_ = json.NewEncoder(w).Encode([]models.DomainDNSSECDataResponse{
			{RecordType: models.DNSSECRecordTypeDSData, Algorithm: 13},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	data, err := client.Domains.EnableDNSSEC(context.Background(), "example.com")

	require.NoError(t, err)
	require.Len(t, data, 1)
	assert.Equal(t, models.DNSSECRecordTypeDSData, data[0].RecordType)
}

func TestDomainsService_ResolveOutboundTransfer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/domains/example.com/transfer/outbound", r.URL.Path)

		var body models.OutboundTransferRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, models.OutboundTransferApprove, body.Action)

		_ = json.NewEncoder(w).Encode(models.OutboundTransferResponse{
			DomainID:   "domain_123",
			DomainName: "example.com",
			Action:     models.OutboundTransferApprove,
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.Domains.ResolveOutboundTransfer(context.Background(), "example.com",
		&models.OutboundTransferRequest{Action: models.OutboundTransferApprove})
	require.NoError(t, err)
	assert.Equal(t, "example.com", result.DomainName)
	assert.Equal(t, models.OutboundTransferApprove, result.Action)
}

func TestDomainsService_GetClaimsNotices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/domains/claims-notices", r.URL.Path)

		var body models.ClaimsNoticesRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, []string{"claims-key-1"}, body.ClaimsKeys)

		_ = json.NewEncoder(w).Encode(models.ClaimsNoticesResponse{
			ClaimsNotices: []models.ClaimsNotice{{
				ClaimsKey:                  "claims-key-1",
				Label:                      "example",
				ClaimsNoticeAcceptanceHash: "hash-1",
				Claims: []models.TmClaim{{
					MarkName: "EXAMPLE",
					Holders: []models.TmHolder{{
						Entitlement: models.HolderEntitlementOwner,
						Addr: models.TmAddr{
							Street: []string{"Musterstrasse 1"},
							City:   "Berlin",
							CC:     "DE",
						},
					}},
					JurDesc: &models.TmJurDesc{JurCC: "DE", Description: "Germany"},
					ClassDescs: []models.TmClassDesc{
						{ClassNum: 9, Description: "Computer software"},
					},
				}},
			}},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	notices, err := client.Domains.GetClaimsNotices(context.Background(), "claims-key-1")
	require.NoError(t, err)
	require.Len(t, notices, 1)
	assert.Equal(t, "hash-1", notices[0].ClaimsNoticeAcceptanceHash)
	require.Len(t, notices[0].Claims, 1)
	assert.Equal(t, "EXAMPLE", notices[0].Claims[0].MarkName)
	assert.Equal(t, models.HolderEntitlementOwner, notices[0].Claims[0].Holders[0].Entitlement)
	assert.Equal(t, "DE", notices[0].Claims[0].Holders[0].Addr.CC)
}
