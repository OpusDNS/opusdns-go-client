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

// tldListBody is a GET /v1/tlds/ response as production sends it for .de and
// .com (idn_tables trimmed), plus a disabled entry. Every entry is a full TLD
// specification plus `enabled`.
const tldListBody = `{"tlds": [
	{"characters": {"max": 63, "min": 1},
	 "contacts": {"authinfo_required": false, "authinfo_supported": true, "is_thick": true,
		"possible_attributes": [
			{"conditions": null, "contact_roles": ["registrant"], "key": "DE_CONTACT_TYPE", "required": false,
			 "type": "enum", "values": ["ORG", "PERSON", "REQUEST"]},
			{"conditions": [{"field": "DE_CONTACT_TYPE", "operator": "equals", "value": "REQUEST"}],
			 "contact_roles": null, "key": "DE_ABUSE_URI_TEMPLATE", "required": false, "type": "uri_template", "values": null}],
		"privacy_proxy": true, "registrant_change": "update",
		"registrant_change_policy": {"confirmation": "none", "confirmation_window": null, "designated_agent_allowed": false,
			"fee": "none", "prior_registrant_notice": false, "renews_domain": false, "transfer_lock_after_change": "P0D"},
		"support_check": true, "support_client_contact_id": true, "support_transfer": false,
		"supported_postal_types": ["int"], "supported_roles": [{"max": 5, "min": 1, "type": "registrant"}],
		"transfer_supported_roles": null, "update_supported_roles": null},
	 "dispute_resolutions": [{"name": "Domain Dispute", "reference": "https://www.denic.de/en/service/dispute", "supported": true}],
	 "dns_configuration": {"allowed_number_of_nameserver": {"max": 13, "min": 0}, "czds_available": false,
		"dnssec_allowed": true, "dnssec_mandatory": false, "dnssec_mode": "DNSKEY", "dnssec_submit_digest_types": null,
		"host_objects": false, "host_parent_check_tlds": null, "registry_nameserver_check": true,
		"registry_root_nameserver_update": 0},
	 "domain_lifecycle": {"add_grace_period": "P0D", "auto_renew_before_expiration": "P0D",
		"default_transfer_renewal_period": {"unit": "y", "value": 1}, "delete_before_expiration": "P0D",
		"delete_policy": ["immediate", "expiration"], "explicit_renew": false, "grace_period": "P0D",
		"pending_delete": "P0D", "redemption_period": "P30D",
		"registration_periods": [{"unit": "y", "value": 1}, {"unit": "m", "value": 1}],
		"registry_auto_renew": true, "renewal_periods": [{"unit": "y", "value": 1}, {"unit": "m", "value": 1}],
		"restore_triggers_renewal": true, "rgp_operations": {"report": true, "request": true},
		"sync_after_operations": ["registration", "renewal", "transfer"], "transfer_grace_period": null,
		"transfer_renewal_periods": [{"unit": "y", "value": 1}, {"unit": "m", "value": 1}]},
	 "domain_statuses": {"default": ["connect"], "status_mapping": {"connect": ["ok"], "failed": ["inactive"]},
		"supported_statuses": ["free", "connect", "failed", "invalid", "serverHold", "pendingCreate", "redemptionPeriod"]},
	 "enabled": true,
	 "idn": {"idn_capable": true, "idn_tables": ["idn_tables/denic_idn_de.yaml"]},
	 "launch_phases": {"general_availability": {"start_date": "2025-01-01T00:00:00Z"},
		"phases": [{"allocation": "fcfs", "end_date": null, "smd_required": true, "start_date": null, "supported": false, "type": "sunrise"}],
		"trademark_claims": {"end_date": null, "start_date": null, "supported": false, "tmch_required": false}},
	 "legal_requirements": [],
	 "local_presence": {"eligible_countries": null, "exemptions": null, "required": false,
		"requirement": ["physical_address", "business_entity"], "type": ["registrant", "admin", "tech", "billing"]},
	 "max_domains_per_check": 1, "parking_enabled": true, "partner_management": {"partners_file": null},
	 "premium_domains": {"affects": ["registration", "renewal"], "source": "EPP", "supported": false},
	 "rdap": {"rdap_server": "https://rdap.denic.de/"},
	 "registry_lock": {"prevents": ["update", "transfer", "delete"], "removal_process": "Manual with authentication",
		"requires_manual_request": true, "supported": true},
	 "required_communications": null,
	 "reserved_domains": {"source": null, "supported": false, "url": null},
	 "tlds": [{"name": "de", "third_level_structure": [], "type": "ccTLD"}],
	 "transfer_policies": {"authinfo_max_length": 16, "authinfo_min_length": 8, "authinfo_pattern": null,
		"authinfo_required": true, "authinfo_set_by_registrar": true, "authinfo_time_limited": true,
		"authinfo_validity_period": "P30D", "contacts_in_transfer_command": false, "info_contact_authinfo": true,
		"info_domain_authinfo": true, "post_transfer_requirements": null, "transfer_ack": "none",
		"transfer_email_required": false,
		"transfer_lock_policy": {"supported_by_registrar": false, "supported_by_registry": false},
		"transfer_nack": "none", "transfer_renewal_period": "P1Y", "transfer_renews_domain": false, "transfer_time": "P0D"},
	 "verification_policies": null,
	 "whois": {"whois_server": "whois.denic.de"}},
	{"characters": {"max": 63, "min": 1},
	 "contacts": {"authinfo_required": false, "authinfo_supported": true, "is_thick": false, "possible_attributes": [],
		"privacy_proxy": true, "registrant_change": "update",
		"registrant_change_policy": {"confirmation": "registrar", "confirmation_window": "P14D", "designated_agent_allowed": true,
			"fee": "none", "prior_registrant_notice": true, "renews_domain": false, "transfer_lock_after_change": "P60D"},
		"support_check": false, "support_client_contact_id": false, "support_transfer": false,
		"supported_postal_types": ["loc", "int"],
		"supported_roles": [{"max": 1, "min": 1, "type": "registrant"}, {"max": 0, "min": 0, "type": "billing"}],
		"transfer_supported_roles": null, "update_supported_roles": null},
	 "dispute_resolutions": [{"name": "UDRP", "reference": "https://www.icann.org/resources/pages/help/dndr/udrp-en", "supported": true}],
	 "dns_configuration": {"allowed_number_of_nameserver": {"max": 13, "min": 0}, "czds_available": false,
		"dnssec_allowed": true, "dnssec_mandatory": false, "dnssec_mode": "DS", "dnssec_submit_digest_types": null,
		"host_objects": true, "host_parent_check_tlds": ["com", "net"], "registry_nameserver_check": false,
		"registry_root_nameserver_update": 0},
	 "domain_lifecycle": {"add_grace_period": "P5D", "auto_renew_before_expiration": "-P7D",
		"default_transfer_renewal_period": {"unit": "y", "value": 1}, "delete_before_expiration": "-P2D",
		"delete_policy": ["immediate", "expiration"], "explicit_renew": true, "grace_period": "P45D",
		"pending_delete": "P5D", "redemption_period": "P30D",
		"registration_periods": [{"unit": "y", "value": 1}, {"unit": "y", "value": 2}, {"unit": "y", "value": 10}],
		"registry_auto_renew": true, "renewal_periods": [{"unit": "y", "value": 1}, {"unit": "y", "value": 10}],
		"restore_triggers_renewal": true, "rgp_operations": {"report": true, "request": true},
		"sync_after_operations": ["registration", "renewal", "transfer"], "transfer_grace_period": "P5D",
		"transfer_renewal_periods": [{"unit": "y", "value": 1}]},
	 "domain_statuses": {"default": ["ok"], "status_mapping": null,
		"supported_statuses": ["ok", "clientTransferProhibited", "serverHold", "pendingTransfer", "pendingDelete"]},
	 "enabled": true,
	 "idn": {"idn_capable": true, "idn_tables": ["idn_tables/com_de_1.2.yaml", "idn_tables/com_latn_2.6.yaml"]},
	 "launch_phases": {"general_availability": {"start_date": "2025-01-01T00:00:00Z"},
		"phases": [{"allocation": "fcfs", "end_date": null, "smd_required": false, "start_date": null, "supported": false, "type": "eap"}],
		"trademark_claims": {"end_date": null, "start_date": null, "supported": false, "tmch_required": false}},
	 "legal_requirements": [],
	 "local_presence": {"eligible_countries": null, "exemptions": null, "required": false,
		"requirement": ["physical_address", "business_entity"], "type": ["registrant", "admin", "tech", "billing"]},
	 "max_domains_per_check": 20, "parking_enabled": true, "partner_management": {"partners_file": null},
	 "premium_domains": {"affects": ["registration", "renewal"], "source": "EPP", "supported": false},
	 "rdap": {"rdap_server": "https://rdap.verisign.com/com/v1/"},
	 "registry_lock": {"prevents": ["update", "transfer", "delete"], "removal_process": "Manual with authentication",
		"requires_manual_request": true, "supported": true},
	 "required_communications": [
		{"required": true, "timedelta": ["-P120D"], "trigger": ["registration_anniversary"], "type": "registration_data_reminder"},
		{"required": true, "timedelta": ["-P30D", "-P7D", "P0D", "P7D"], "trigger": ["expiration"], "type": "renewal_reminder"}],
	 "reserved_domains": {"source": "manual", "supported": true,
		"url": "https://www.icann.org/en/registry-agreements/com/com-registry-agreement-appendix-6-1-12-2012-en"},
	 "tlds": [{"name": "com", "third_level_structure": [], "type": "gTLD"}],
	 "transfer_policies": {"authinfo_max_length": 32, "authinfo_min_length": 8, "authinfo_pattern": null,
		"authinfo_required": true, "authinfo_set_by_registrar": true, "authinfo_time_limited": false,
		"authinfo_validity_period": null, "contacts_in_transfer_command": false, "info_contact_authinfo": true,
		"info_domain_authinfo": true, "post_transfer_requirements": null, "transfer_ack": "both",
		"transfer_email_required": false,
		"transfer_lock_policy": {"supported_by_registrar": true, "supported_by_registry": true},
		"transfer_nack": "both", "transfer_renewal_period": "P1Y", "transfer_renews_domain": true, "transfer_time": "P5D"},
	 "verification_policies": {"email_verification": {"communication": {"timedelta": ["+0D", "+4D", "+12D"], "type": "email"},
		"contact_roles": ["registrant"], "enabled": true, "suspension_delay": "15D", "suspension_on_failure": true,
		"trigger": ["domain_registration", "domain_inbound_transfer"], "validity_period": null, "verification_method": "email"}},
	 "whois": {"whois_server": "whois.verisign-grs.com"}},
	{"enabled": false, "idn": {"idn_capable": false},
	 "tlds": [{"name": "xyz", "third_level_structure": [], "type": "gTLD"}]}
]}`

func TestTLDsService_ListTLDs_Specification(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(tldListBody))
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	tlds, err := client.TLDs.ListTLDs(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, tlds, 2)

	de := tlds[0]
	assert.Equal(t, "de", de.Name)
	assert.Equal(t, models.TLDTypeCCTLD, de.Type)
	assert.True(t, de.Available)
	assert.True(t, de.IDNSupported)
	assert.True(t, de.DNSSECSupported)
	// The one-month period is not a whole year, so it is left out.
	assert.Equal(t, 1, de.MinRegistrationPeriod)
	assert.Equal(t, 1, de.MaxRegistrationPeriod)
	require.NotNil(t, de.Specification)
	require.NotNil(t, de.Specification.Whois)
	assert.Equal(t, models.StringPtr("whois.denic.de"), de.Specification.Whois.WhoisServer)
	assert.Len(t, de.Specification.DomainLifecycle.RegistrationPeriods, 2)

	com := tlds[1]
	assert.Equal(t, "com", com.Name)
	assert.Equal(t, models.TLDTypeGTLD, com.Type)
	assert.Equal(t, 1, com.MinRegistrationPeriod)
	assert.Equal(t, 10, com.MaxRegistrationPeriod)
	require.NotNil(t, com.Specification)
	assert.Equal(t, []string{"com", "net"}, com.Specification.DNSConfiguration.HostParentCheckTLDs)
}

func TestRegistrationYears(t *testing.T) {
	tests := []struct {
		name     string
		periods  []models.DomainPeriod
		min, max int
	}{
		{"none", nil, 0, 0},
		{"years", []models.DomainPeriod{{Value: 2, Unit: "y"}, {Value: 1, Unit: "y"}, {Value: 5, Unit: "y"}}, 1, 5},
		{"whole months", []models.DomainPeriod{{Value: 24, Unit: "m"}, {Value: 12, Unit: "m"}}, 1, 2},
		{"only partial months", []models.DomainPeriod{{Value: 1, Unit: "m"}, {Value: 30, Unit: "d"}}, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minYears, maxYears := registrationYears(tt.periods)
			assert.Equal(t, tt.min, minYears)
			assert.Equal(t, tt.max, maxYears)
		})
	}
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
		"email_verification": {"enabled": true, "contact_roles": ["registrant"], "trigger": ["create"],
		          "suspension_on_failure": true, "suspension_delay": "P15D",
		          "validity_period": {"value": 1, "unit": "y"},
		          "verification_method": "link",
		          "communication": {"type": "email", "timedelta": ["P1D", "P7D"]}},
		"identity_verification": {"enabled": true, "contact_roles": ["registrant"], "trigger": ["create"],
		          "suspension_on_failure": false, "suspension_delay": "P30D",
		          "validity_period": null,
		          "required_claims": [
		              {"claim": "NAME", "accepted_proofs": ["IDCARD", "PASSPORT"]},
		              {"claim": "ADDRESS", "accepted_proofs": null}
		          ]}
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
	assert.Equal(t, []models.TLDContactRole{{Type: models.DomainContactTypeRegistrant, Min: 1, Max: 1}},
		spec.Contacts.SupportedRoles)

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

	policy, ok := spec.VerificationPolicies["email_verification"]
	require.True(t, ok)
	assert.True(t, policy.Enabled)
	assert.Equal(t, "P15D", policy.SuspensionDelay)
	require.NotNil(t, policy.Communication)
	assert.Equal(t, []string{"P1D", "P7D"}, policy.Communication.Timedelta)

	// required_claims is an array of RequiredClaim objects, not of strings.
	identity, ok := spec.VerificationPolicies["identity_verification"]
	require.True(t, ok)
	assert.Nil(t, identity.ValidityPeriod)
	assert.Equal(t, []models.TLDRequiredClaim{
		{
			Claim:          models.ContactVerificationClaimName,
			AcceptedProofs: []models.ContactVerificationProof{models.ContactVerificationProofIDCard, models.ContactVerificationProofPassport},
		},
		{Claim: models.ContactVerificationClaimAddress},
	}, identity.RequiredClaims)
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
			}, {
				"domain": "example.net", "available": true, "premium": false,
				"price": {"amount": null, "currency": "USD", "period": {"value": 1, "unit": "y"}},
				"renewal_price": null,
				"marketplace": {"name": "fruits"},
				"connected_account": {"type": "ams", "registry": "verisign",
				                      "registry_account_id": "registry_account_01h45ytscbebyvny4gc8cr8ma2"}
			}],
			"meta": {"total": 2, "processing_time_ms": 30}
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
	require.Len(t, result.Suggestions, 2)
	suggestion := result.Suggestions[0]
	assert.Equal(t, "example.com", suggestion.Domain)
	assert.True(t, suggestion.Available)
	require.NotNil(t, suggestion.Price.Amount)
	assert.Equal(t, "12.50", *suggestion.Price.Amount)
	assert.Equal(t, models.DomainPeriod{Value: 1, Unit: models.PeriodUnitYear}, suggestion.Price.Period)
	require.NotNil(t, suggestion.Marketplace)
	assert.Equal(t, "fruits", suggestion.Marketplace.Name)
	assert.Nil(t, suggestion.ConnectedAccount)

	// connected_account is a DomainAmsAccountResponse.
	assert.Equal(t, &models.DomainAMSAccount{
		Type:              "ams",
		Registry:          "verisign",
		RegistryAccountID: "registry_account_01h45ytscbebyvny4gc8cr8ma2",
	}, result.Suggestions[1].ConnectedAccount)
	assert.Equal(t, 2, result.Meta.Total)
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
