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

func TestWhitelabelService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/whitelabel-branding", r.URL.Path)

		_ = json.NewEncoder(w).Encode(models.Whitelabel{
			WhitelabelBrandingID: "whitelabel_branding_123",
			OrganizationID:       "organization_123",
			Tier:                 models.WhitelabelTierPlus,
			OnboardingStatus:     models.WhitelabelOnboardingActive,
			Hostname:             "app.reseller.com",
			Subscription: &models.WhitelabelSubscription{
				RenewalMode: models.WhitelabelRenewalModeRenew,
				Period:      models.DomainPeriod{Value: 1, Unit: models.PeriodUnitMonth},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	config, err := client.Whitelabel.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, models.WhitelabelTierPlus, config.Tier)
	assert.Equal(t, models.WhitelabelOnboardingActive, config.OnboardingStatus)
	require.NotNil(t, config.Subscription)
	assert.Equal(t, models.WhitelabelRenewalModeRenew, config.Subscription.RenewalMode)
}

func TestWhitelabelService_CreateBase(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/whitelabel-branding", r.URL.Path)

		var body map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "base", body["tier"])
		assert.Equal(t, "reseller", body["label"])

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(models.ProductCreateResponse{
			SubscriptionID: "sub_1",
			SubscribableID: "whitelabel_branding_123",
			ProductName:    "whitelabel_branding",
		})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.Whitelabel.CreateBase(context.Background(), &models.WhitelabelBaseCreateRequest{
		Label:  "reseller",
		Period: models.DomainPeriod{Value: 1, Unit: models.PeriodUnitMonth},
	})
	require.NoError(t, err)
	assert.Equal(t, "sub_1", result.SubscriptionID)
}

func TestWhitelabelService_CreateRequiresRequest(t *testing.T) {
	client, err := NewClient(WithAPIKey("opk_test"))
	require.NoError(t, err)

	_, err = client.Whitelabel.CreateBase(context.Background(), nil)
	require.ErrorIs(t, err, ErrInvalidInput)

	_, err = client.Whitelabel.CreatePlus(context.Background(), nil)
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestWhitelabelService_CreatePlus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "plus", body["tier"])
		assert.Equal(t, "reseller.com", body["hostname"])
		assert.Equal(t, "auth", body["auth_subdomain"])
		assert.Equal(t, true, body["create_zone"])

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(models.ProductCreateResponse{SubscriptionID: "sub_2"})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	_, err = client.Whitelabel.CreatePlus(context.Background(), &models.WhitelabelPlusCreateRequest{
		Label:         "reseller",
		Period:        models.DomainPeriod{Value: 1, Unit: models.PeriodUnitYear},
		Hostname:      "reseller.com",
		AuthSubdomain: "auth",
		CreateZone:    models.BoolPtr(true),
	})
	require.NoError(t, err)
}

func TestWhitelabelService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/whitelabel-branding", r.URL.Path)

		var body models.WhitelabelUpdateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.NotNil(t, body.RenewalMode)
		assert.Equal(t, models.WhitelabelRenewalModeExpire, *body.RenewalMode)

		_ = json.NewEncoder(w).Encode(models.Whitelabel{WhitelabelBrandingID: "whitelabel_branding_123"})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	mode := models.WhitelabelRenewalModeExpire
	_, err = client.Whitelabel.Update(context.Background(), &models.WhitelabelUpdateRequest{RenewalMode: &mode})
	require.NoError(t, err)
}

func TestWhitelabelService_UpgradeToPlus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/whitelabel-branding/tier", r.URL.Path)

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(models.ProductCreateResponse{SubscriptionID: "sub_3"})
	}))
	defer server.Close()

	client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
	require.NoError(t, err)

	result, err := client.Whitelabel.UpgradeToPlus(context.Background(), &models.WhitelabelUpgradeRequest{
		Hostname:      "reseller.com",
		AuthSubdomain: "auth",
	})
	require.NoError(t, err)
	assert.Equal(t, "sub_3", result.SubscriptionID)
}

func TestWhitelabelService_RecheckAndRestore(t *testing.T) {
	t.Run("recheck", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/whitelabel-branding/recheck", r.URL.Path)

			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(models.Whitelabel{
				OnboardingStatus: models.WhitelabelOnboardingVerifying,
			})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		config, err := client.Whitelabel.Recheck(context.Background(), nil)
		require.NoError(t, err)
		assert.Equal(t, models.WhitelabelOnboardingVerifying, config.OnboardingStatus)
	})

	t.Run("restore", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/whitelabel-branding/restore", r.URL.Path)

			_ = json.NewEncoder(w).Encode(models.Whitelabel{Enabled: true})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		config, err := client.Whitelabel.Restore(context.Background())
		require.NoError(t, err)
		assert.True(t, config.Enabled)
	})
}

func TestWhitelabelService_EmailTemplates(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/v1/whitelabel-branding/email/templates", r.URL.Path)

			_ = json.NewEncoder(w).Encode(map[string]models.MailTemplate{
				"domain_expiring": {
					Label:   "Domain expiring",
					Locales: []string{"en", "de"},
					Subject: map[string]string{"en": "Your domain expires soon"},
					Blocks: map[string]models.MailTemplateBlock{
						"intro": {Label: "Intro", MaxLen: 200, Multiline: true},
					},
				},
			})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		templates, err := client.Whitelabel.ListEmailTemplates(context.Background())
		require.NoError(t, err)
		require.Contains(t, templates, "domain_expiring")
		assert.Equal(t, []string{"en", "de"}, templates["domain_expiring"].Locales)
		assert.Equal(t, 200, templates["domain_expiring"].Blocks["intro"].MaxLen)
	})

	t.Run("preview", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/whitelabel-branding/email/preview", r.URL.Path)

			var body models.PreviewMailRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "domain_expiring", body.TemplateName)
			assert.Equal(t, "de", body.LanguageCode)
			require.NotNil(t, body.BrandingDocument)
			require.NotNil(t, body.BrandingDocument.Theme)
			require.NotNil(t, body.BrandingDocument.Theme.Light)
			assert.Equal(t, "#0a7cff", *body.BrandingDocument.Theme.Light.Primary)

			_ = json.NewEncoder(w).Encode(models.PreviewMailResponse{
				Subject: "Deine Domain läuft bald ab",
				HTML:    "<p>…</p>",
				Text:    "…",
			})
		}))
		defer server.Close()

		client, err := NewClient(WithAPIKey("opk_test"), WithAPIEndpoint(server.URL))
		require.NoError(t, err)

		preview, err := client.Whitelabel.PreviewEmail(context.Background(), &models.PreviewMailRequest{
			TemplateName: "domain_expiring",
			LanguageCode: "de",
			BrandingDocument: &models.BrandingDocument{
				Theme: &models.Theme{Light: &models.Palette{Primary: models.StringPtr("#0a7cff")}},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "Deine Domain läuft bald ab", preview.Subject)
	})
}
