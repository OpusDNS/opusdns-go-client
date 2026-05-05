//go:build integration
// +build integration

package opusdns_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/opusdns/opusdns-go-client/opusdns"
)

const (
	envIntegrationZone    = "OPUSDNS_INTEGRATION_ZONE"
	envIntegrationTimeout = "OPUSDNS_INTEGRATION_TIMEOUT"

	integrationRecordName = "www"
	integrationRecordData = "192.0.2.123"
)

func TestRealAPIReadOnlySmoke(t *testing.T) {
	client := newIntegrationClient(t)
	ctx, cancel := newIntegrationContext(t)
	defer cancel()

	zones, err := client.DNS.ListZonesPage(ctx, &models.ListZonesOptions{
		Page:      1,
		PageSize:  1,
		SortBy:    models.ZoneSortByName,
		SortOrder: models.SortAsc,
	})
	if err != nil {
		fatalRealAPIError(t, "list zones page", err)
	}
	if zones == nil {
		t.Fatal("list zones page returned nil response")
	}
	t.Logf("GET /v1/dns?page=1&page_size=1&sort_by=name&sort_order=asc -> OK: results=%d current_page=%d total_pages=%d total_items=%d has_next_page=%t",
		len(zones.Results),
		zones.Pagination.CurrentPage,
		zones.Pagination.TotalPages,
		paginationTotalItems(zones.Pagination),
		zones.Pagination.HasNextPage,
	)

	summary, err := client.DNS.GetSummary(ctx)
	if err != nil {
		fatalRealAPIError(t, "get DNS summary", err)
	}
	t.Logf("GET /v1/dns/summary -> OK: total_zones=%d dnssec_status_counts=%v", summary.TotalZones, summary.ZonesByDNSSEC)

	availability, err := client.Availability.CheckSingleAvailability(ctx, "example.com")
	if err != nil {
		fatalRealAPIError(t, "check domain availability", err)
	}
	if availability == nil || availability.Domain == "" || availability.Status == "" {
		t.Fatalf("availability response missing expected fields: %#v", availability)
	}
	t.Logf("GET /v1/availability?domains=example.com -> OK: domain=%s status=%s", availability.Domain, availability.Status)
}

func TestRealAPIDNSZoneLifecycle(t *testing.T) {
	zoneName := strings.TrimSuffix(strings.TrimSpace(os.Getenv(envIntegrationZone)), ".")
	if zoneName == "" {
		t.Skipf("set %s to a unique disposable zone name to run the DNS write lifecycle", envIntegrationZone)
	}
	validateIntegrationZoneName(t, zoneName)

	client := newIntegrationClient(t)
	ctx, cancel := newIntegrationContext(t)
	defer cancel()

	if _, err := client.DNS.GetZone(ctx, zoneName); err == nil {
		t.Fatalf("refusing to run write lifecycle because zone %q already exists", zoneName)
	} else if !opusdns.IsNotFoundError(err) {
		fatalRealAPIError(t, "preflight get zone "+zoneName, err)
	} else {
		t.Logf("GET /v1/dns/%s -> OK: not_found, safe to create disposable zone", zoneName)
	}

	created := false
	defer func() {
		if !created {
			return
		}

		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), integrationTimeout(t))
		defer cleanupCancel()

		if err := client.DNS.DeleteZone(cleanupCtx, zoneName); err != nil && !opusdns.IsNotFoundError(err) {
			t.Errorf("cleanup delete zone %q: %v", zoneName, err)
		} else {
			t.Logf("DELETE /v1/dns/%s -> OK: cleanup completed", zoneName)
		}
	}()

	zone, err := client.DNS.CreateZone(ctx, &models.ZoneCreateRequest{Name: zoneName})
	if err != nil {
		fatalRealAPIError(t, "create zone "+zoneName, err)
	}
	created = true
	if zone.Name != "" && !sameDNSName(zone.Name, zoneName) {
		t.Fatalf("created zone name mismatch: got %q, want %q", zone.Name, zoneName)
	}
	t.Logf("POST /v1/dns -> OK: %s", zoneSummary(zone))

	zone, err = client.DNS.GetZone(ctx, zoneName+".")
	if err != nil {
		fatalRealAPIError(t, "get created zone "+zoneName+" with trailing dot", err)
	}
	t.Logf("GET /v1/dns/%s. -> OK: %s", zoneName, zoneSummary(zone))

	record := models.Record{
		Name:  integrationRecordName,
		Type:  models.RRSetTypeA,
		TTL:   300,
		RData: integrationRecordData,
	}
	if err := client.DNS.UpsertRecord(ctx, zoneName, record); err != nil {
		fatalRealAPIError(t, "upsert A record in zone "+zoneName, err)
	}
	t.Logf("PATCH /v1/dns/%s/records -> OK: upsert %s %s ttl=%d rdata=%s", zoneName, record.Name, record.Type, record.TTL, record.RData)

	zone, err = client.DNS.GetZone(ctx, zoneName)
	if err != nil {
		fatalRealAPIError(t, "get zone after A record upsert "+zoneName, err)
	}
	t.Logf("GET /v1/dns/%s -> OK after record upsert: %s", zoneName, zoneSummary(zone))

	if err := client.DNS.DeleteRecord(ctx, zoneName, record); err != nil {
		fatalRealAPIError(t, "delete A record from zone "+zoneName, err)
	}
	t.Logf("PATCH /v1/dns/%s/records -> OK: remove %s %s ttl=%d rdata=%s", zoneName, record.Name, record.Type, record.TTL, record.RData)

	zone, err = client.DNS.GetZone(ctx, zoneName)
	if err != nil {
		fatalRealAPIError(t, "get zone after A record delete "+zoneName, err)
	}
	t.Logf("GET /v1/dns/%s -> OK after record delete: %s", zoneName, zoneSummary(zone))

	if err := client.DNS.DeleteZone(ctx, zoneName); err != nil {
		fatalRealAPIError(t, "delete zone "+zoneName, err)
	}
	created = false
	t.Logf("DELETE /v1/dns/%s -> OK: zone deleted", zoneName)

	if _, err := client.DNS.GetZone(ctx, zoneName); err == nil {
		t.Fatalf("expected zone %q to be deleted", zoneName)
	} else if !opusdns.IsNotFoundError(err) {
		fatalRealAPIError(t, "verify deleted zone "+zoneName, err)
	} else {
		t.Logf("GET /v1/dns/%s -> OK: not_found after delete", zoneName)
	}
}

func fatalRealAPIError(t *testing.T, operation string, err error) {
	t.Helper()

	if opusdns.IsUnauthorizedError(err) {
		endpoint := os.Getenv(opusdns.EnvAPIEndpoint)
		if endpoint == "" {
			endpoint = opusdns.DefaultAPIEndpoint
		}
		t.Fatalf("%s: %v\nThe API endpoint %s rejected OPUSDNS_API_KEY. Use a raw API key value issued for that environment, without quotes or a Bearer prefix.", operation, err, endpoint)
	}

	t.Fatalf("%s: %v", operation, err)
}

func newIntegrationClient(t *testing.T) *opusdns.Client {
	t.Helper()

	if os.Getenv(opusdns.EnvAPIKey) == "" {
		t.Skipf("set %s to run real API integration tests", opusdns.EnvAPIKey)
	}

	client, err := opusdns.NewClient(opusdns.WithHTTPTimeout(integrationTimeout(t)))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	return client
}

func newIntegrationContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), integrationTimeout(t))
}

func integrationTimeout(t *testing.T) time.Duration {
	t.Helper()

	raw := os.Getenv(envIntegrationTimeout)
	if raw == "" {
		return 2 * time.Minute
	}

	timeout, err := time.ParseDuration(raw)
	if err != nil {
		t.Fatalf("parse %s: %v", envIntegrationTimeout, err)
	}
	if timeout <= 0 {
		t.Fatalf("%s must be greater than zero", envIntegrationTimeout)
	}
	return timeout
}

func validateIntegrationZoneName(t *testing.T, zoneName string) {
	t.Helper()

	if !strings.Contains(zoneName, ".") {
		t.Fatalf("%s must be a fully qualified disposable zone name, got %q", envIntegrationZone, zoneName)
	}
	if strings.ContainsAny(zoneName, " \t\r\n/") {
		t.Fatalf("%s contains invalid characters: %q", envIntegrationZone, zoneName)
	}
}

func paginationTotalItems(p models.Pagination) int {
	if p.TotalItems != 0 {
		return p.TotalItems
	}
	return p.TotalCount
}

func zoneSummary(zone *models.Zone) string {
	if zone == nil {
		return "zone=<nil>"
	}

	return fmt.Sprintf(
		"zone=%s zone_id=%s dnssec_status=%s rrsets=%d",
		normalizeDNSName(zone.Name),
		zone.ZoneID.String(),
		zone.DNSSECStatus,
		len(zone.RRSets),
	)
}

func sameDNSName(left, right string) bool {
	return normalizeDNSName(left) == normalizeDNSName(right)
}

func normalizeDNSName(name string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(name)), ".")
}
