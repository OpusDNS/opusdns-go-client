package opusdns

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTLDsService_ListTLDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tlds/", r.URL.Path)
		// The endpoint only knows `fields` and `tlds`; deprecated filters
		// must not be sent.
		assert.Equal(t, "com,net,xyz", r.URL.Query().Get("tlds"))
		assert.Empty(t, r.URL.Query().Get("search"))
		assert.Empty(t, r.URL.Query().Get("type"))

		// Each entry is a full TLD specification plus `enabled`.
		_, _ = w.Write([]byte(`{"tlds": [
			{"enabled": true, "parking_enabled": false,
			 "tlds": [{"name": "com", "type": "gTLD", "third_level_structure": []},
			          {"name": "net", "type": "gTLD", "third_level_structure": []}],
			 "characters": {"min": 1, "max": 63}},
			{"enabled": false,
			 "tlds": [{"name": "xyz", "type": "gTLD", "third_level_structure": []}]}
		]}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	tlds, err := client.TLDs.ListTLDs(context.Background(), &models.ListTLDsOptions{
		TLDs: []string{"com", "net", "xyz"},
		// Deprecated filters, set on purpose: they must not be sent.
		Search: "ignored",           //nolint:staticcheck
		Type:   models.TLDTypeCCTLD, //nolint:staticcheck
	})

	require.NoError(t, err)
	assert.Len(t, tlds, 2)
	assert.Equal(t, "com", tlds[0].Name)
	assert.Equal(t, models.TLDTypeGTLD, tlds[0].Type)
	assert.True(t, tlds[0].Available)
	assert.Equal(t, "net", tlds[1].Name)
}

func TestTLDsService_ListTLDs_NoOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery)
		_, _ = w.Write([]byte(`{"tlds": []}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	tlds, err := client.TLDs.ListTLDs(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, tlds)
}

// tldSpecificationBody is a TldSpecificationResponse for .de, trimmed to one
// or two entries per list.
const tldSpecificationBody = `{
	"tlds": [{"name": "de", "type": "ccTLD", "third_level_structure": []}],
	"characters": {"min": 1, "max": 63},
	"idn": {"idn_capable": true, "idn_tables": ["de"]},
	"dns_configuration": {
		"allowed_number_of_nameserver": {"min": 2, "max": 5},
		"czds_available": null,
		"dnssec_allowed": true,
		"dnssec_mandatory": false,
		"dnssec_mode": "DNSKEY",
		"dnssec_submit_digest_types": null,
		"host_objects": false,
		"host_parent_check_tlds": null,
		"registry_nameserver_check": true,
		"registry_root_nameserver_update": 0
	},
	"domain_lifecycle": {
		"add_grace_period": null,
		"grace_period": "P0D",
		"redemption_period": "P30D",
		"registration_periods": [{"value": 1, "unit": "y"}],
		"renewal_periods": [{"value": 1, "unit": "y"}],
		"transfer_renewal_periods": null,
		"default_transfer_renewal_period": null,
		"restore_triggers_renewal": true,
		"rgp_operations": {"report": false, "request": true},
		"explicit_renew": false
	},
	"domain_statuses": {
		"default": ["ok"],
		"supported_statuses": ["ok", "clientHold"],
		"status_mapping": null
	},
	"transfer_policies": {
		"authinfo_required": true,
		"authinfo_set_by_registrar": true,
		"authinfo_min_length": 8,
		"authinfo_max_length": 16,
		"authinfo_validity_period": "P30D",
		"contacts_in_transfer_command": false,
		"transfer_lock_policy": {"supported_by_registrar": false, "supported_by_registry": false},
		"transfer_time": "PT0S"
	},
	"contacts": {
		"authinfo_supported": false,
		"is_thick": true,
		"possible_attributes": [{
			"key": "DE_CONTACT_TYPE", "type": "enum", "values": ["PERSON", "ORG"],
			"required": false, "contact_roles": ["registrant"],
			"conditions": [{"field": "DE_CONTACT_TYPE", "operator": "in", "value": ["ORG"]}]
		}],
		"supported_roles": [{"type": "registrant", "min": 1, "max": 1}]
	},
	"launch_phases": {
		"general_availability": {"start_date": "2009-01-01T00:00:00Z"},
		"phases": [{"type": "sunrise", "supported": false, "allocation": null,
		            "start_date": null, "end_date": null, "smd_required": null}],
		"trademark_claims": null
	},
	"legal_requirements": [{"key": "denic_terms", "type": "notice", "operations": ["registration"],
	                        "url": "https://www.denic.de/en/terms"}],
	"local_presence": {"required": false, "eligible_countries": null, "exemptions": null,
	                   "requirement": null, "type": null},
	"parking_enabled": true,
	"premium_domains": {"supported": false, "affects": null, "source": null},
	"rdap": {"rdap_server": "https://rdap.denic.de/"},
	"registry_lock": {"supported": true, "prevents": ["update", "delete"],
	                  "removal_process": null, "requires_manual_request": true},
	"reserved_domains": {"supported": false, "source": null, "url": null},
	"verification_policies": {
		"email": {"enabled": true, "contact_roles": ["registrant"], "trigger": ["create"],
		          "suspension_on_failure": true, "suspension_delay": "P15D",
		          "validity_period": {"value": 1, "unit": "y"},
		          "verification_method": "link",
		          "communication": {"type": "email", "timedelta": ["P1D", "P7D"]}}
	},
	"whois": {"whois_server": "whois.denic.de"}
}`

func TestTLDsService_GetTLD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tlds/de", r.URL.Path)
		_, _ = w.Write([]byte(tldSpecificationBody))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	spec, err := client.TLDs.GetTLD(context.Background(), "de")
	require.NoError(t, err)

	require.Len(t, spec.TLDs, 1)
	assert.Equal(t, "de", spec.TLDs[0].Name)
	assert.Equal(t, models.TLDTypeCCTLD, spec.TLDs[0].Type)
	assert.Equal(t, models.TLDRange{Min: 1, Max: 63}, spec.Characters)
	assert.True(t, spec.IDN.IDNCapable)
	assert.Equal(t, []string{"de"}, spec.IDN.IDNTables)

	dns := spec.DNSConfiguration
	assert.Equal(t, models.TLDRange{Min: 2, Max: 5}, dns.AllowedNumberOfNameservers)
	assert.True(t, dns.DNSSECAllowed)
	require.NotNil(t, dns.DNSSECMode)
	assert.Equal(t, "DNSKEY", *dns.DNSSECMode)

	lifecycle := spec.DomainLifecycle
	require.NotNil(t, lifecycle.RedemptionPeriod)
	assert.Equal(t, "P30D", *lifecycle.RedemptionPeriod)
	assert.Equal(t, []models.DomainPeriod{{Value: 1, Unit: models.PeriodUnitYear}}, lifecycle.RegistrationPeriods)
	require.NotNil(t, lifecycle.RGPOperations)
	assert.True(t, lifecycle.RGPOperations.Request)

	assert.Equal(t, []string{"ok", "clientHold"}, spec.DomainStatuses.SupportedStatuses)

	transfer := spec.TransferPolicies
	assert.True(t, transfer.AuthinfoRequired)
	require.NotNil(t, transfer.AuthinfoMinLength)
	assert.Equal(t, 8, *transfer.AuthinfoMinLength)

	require.NotNil(t, spec.Contacts)
	require.Len(t, spec.Contacts.PossibleAttributes, 1)
	attribute := spec.Contacts.PossibleAttributes[0]
	assert.Equal(t, models.RegistryAttrDEContactType, attribute.Key)
	assert.Equal(t, []models.DomainContactType{models.DomainContactTypeRegistrant}, attribute.ContactRoles)
	require.Len(t, attribute.Conditions, 1)
	assert.Equal(t, models.ConditionOperatorIn, attribute.Conditions[0].Operator)
	assert.Equal(t, []interface{}{"ORG"}, attribute.Conditions[0].Value)
	require.Len(t, spec.Contacts.SupportedRoles, 1)
	assert.Equal(t, 1, spec.Contacts.SupportedRoles[0].Min)

	require.NotNil(t, spec.LaunchPhases)
	require.NotNil(t, spec.LaunchPhases.GeneralAvailability.StartDate)
	assert.True(t, spec.LaunchPhases.GeneralAvailability.StartDate.Equal(time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC)))
	require.Len(t, spec.LaunchPhases.Phases, 1)
	assert.Equal(t, models.LaunchPhaseSunrise, spec.LaunchPhases.Phases[0].Type)

	require.Len(t, spec.LegalRequirements, 1)
	assert.Equal(t, "notice", spec.LegalRequirements[0].Type)
	assert.True(t, spec.ParkingEnabled)
	require.NotNil(t, spec.RDAP)
	assert.Equal(t, "https://rdap.denic.de/", *spec.RDAP.RDAPServer)
	require.NotNil(t, spec.RegistryLock)
	assert.Equal(t, []string{"update", "delete"}, spec.RegistryLock.Prevents)
	require.NotNil(t, spec.Whois)
	assert.Equal(t, "whois.denic.de", *spec.Whois.WhoisServer)

	policy, ok := spec.VerificationPolicies["email"]
	require.True(t, ok)
	assert.True(t, policy.Enabled)
	assert.Equal(t, "P15D", policy.SuspensionDelay)
	require.NotNil(t, policy.Communication)
	assert.Equal(t, []string{"P1D", "P7D"}, policy.Communication.Timedelta)
}

func TestTLDsService_GetPortfolio(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tlds/portfolio", r.URL.Path)
		// A bare array of TldResponseShort.
		_, _ = w.Write([]byte(`[{"tld": "com", "type": "gTLD"}, {"tld": "de", "type": "ccTLD"}]`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	portfolio, err := client.TLDs.GetPortfolio(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []models.TLDPortfolioEntry{
		{TLD: "com", Type: models.TLDTypeGTLD},
		{TLD: "de", Type: models.TLDTypeCCTLD},
	}, portfolio)
}

func TestAvailabilityService_CheckAvailability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/availability", r.URL.Path)
		assert.Equal(t, []string{"example.com", "example.invalid"}, r.URL.Query()["domains"])

		_, _ = w.Write([]byte(`{
			"results": [
				{"domain": "example.com", "status": "unavailable", "error": null},
				{"domain": "example.invalid", "status": "error",
				 "error": {"type": "unsupported_tld", "message": "TLD not supported"}}
			],
			"meta": {"total": 2, "processing_time_ms": 50}
		}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.Availability.CheckAvailability(context.Background(), []string{"example.com", "example.invalid"})

	require.NoError(t, err)
	require.Len(t, result.Results, 2)
	assert.Equal(t, "example.com", result.Results[0].Domain)
	assert.Equal(t, models.AvailabilityStatusUnavailable, result.Results[0].Status)
	assert.Nil(t, result.Results[0].Error)
	assert.Equal(t, models.AvailabilityStatusError, result.Results[1].Status)
	require.NotNil(t, result.Results[1].Error)
	assert.Equal(t, "TLD not supported", result.Results[1].Error.Message)
	assert.Equal(t, 2, result.Meta.Total)
}

func TestAvailabilityService_CheckAvailability_SplitsLongLists(t *testing.T) {
	domains := make([]string, 120)
	for i := range domains {
		domains[i] = fmt.Sprintf("example%d.com", i)
	}

	var batchSizes []int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		batch := r.URL.Query()["domains"]
		batchSizes = append(batchSizes, len(batch))

		resp := models.AvailabilityResponse{Meta: models.AvailabilityMeta{Total: len(batch), ProcessingTimeMs: 10}}
		for _, domain := range batch {
			resp.Results = append(resp.Results, models.DomainAvailability{Domain: domain, Status: models.AvailabilityStatusAvailable})
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.Availability.CheckAvailability(context.Background(), domains)
	require.NoError(t, err)

	// The API takes at most 50 domains per request.
	assert.Equal(t, []int{50, 50, 20}, batchSizes)
	require.Len(t, result.Results, 120)
	for i, availability := range result.Results {
		assert.Equal(t, domains[i], availability.Domain)
	}
	assert.Equal(t, 120, result.Meta.Total)
	assert.Equal(t, 30, result.Meta.ProcessingTimeMs)
}

func TestAvailabilityService_CheckAvailability_RequiresDomains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no request expected, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	_, err = client.Availability.CheckAvailability(context.Background(), nil)
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestAvailabilityService_CheckSingleAvailability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/availability", r.URL.Path)
		assert.Equal(t, []string{"example.com"}, r.URL.Query()["domains"])

		_, _ = w.Write([]byte(`{"results": [{"domain": "example.com", "status": "available"}],
			"meta": {"total": 1, "processing_time_ms": 25}}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.Availability.CheckSingleAvailability(context.Background(), "example.com")

	require.NoError(t, err)
	assert.Equal(t, "example.com", result.Domain)
	assert.Equal(t, models.AvailabilityStatusAvailable, result.Status)
}

func TestAvailabilityService_GetSuggestions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/domain-search/suggest", r.URL.Path)
		assert.Equal(t, "example", r.URL.Query().Get("query"))
		assert.Equal(t, []string{"com"}, r.URL.Query()["tlds"])
		assert.Equal(t, "5", r.URL.Query().Get("limit"))

		// DomainSearchResponse: the suggestions arrive under "results".
		_, _ = w.Write([]byte(`{
			"results": [{
				"domain": "example.com", "available": true, "premium": false,
				"price": {"amount": "12.50", "currency": "USD", "period": {"value": 1, "unit": "y"}},
				"renewal_price": null,
				"marketplace": {"name": "fruits"},
				"connected_account": null
			}],
			"meta": {"total": 1, "processing_time_ms": 30}
		}`))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.Availability.GetSuggestions(context.Background(), "example", &models.DomainSuggestRequest{
		TLDs:  []string{"com"},
		Limit: 5,
	})

	require.NoError(t, err)
	require.Len(t, result.Suggestions, 1)
	suggestion := result.Suggestions[0]
	assert.Equal(t, "example.com", suggestion.Domain)
	assert.True(t, suggestion.Available)
	require.NotNil(t, suggestion.Price.Amount)
	assert.Equal(t, "12.50", *suggestion.Price.Amount)
	assert.Equal(t, models.DomainPeriod{Value: 1, Unit: models.PeriodUnitYear}, suggestion.Price.Period)
	require.NotNil(t, suggestion.Marketplace)
	assert.Equal(t, "fruits", suggestion.Marketplace.Name)
	assert.Equal(t, 1, result.Meta.Total)
}

func TestTLDsService_GetTLDWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tlds/com", r.URL.Path)

		query := r.URL.Query()
		assert.Equal(t, "nominet_dragon", query.Get("backend"))
		assert.Equal(t, "acme", query.Get("customer_spec_ref"))
		assert.Equal(t, "3", query.Get("version"))

		_, _ = w.Write([]byte(tldSpecificationBody))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	spec, err := client.TLDs.GetTLDWithOptions(context.Background(), "com", &models.GetTLDOptions{
		Backend:         "nominet_dragon",
		CustomerSpecRef: "acme",
		Version:         "3",
	})
	require.NoError(t, err)
	require.Len(t, spec.TLDs, 1)
}
