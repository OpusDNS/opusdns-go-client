package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/opusdns/opusdns-go-client/client"
	"github.com/opusdns/opusdns-go-client/models"
)

func main() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create client using environment variables or explicit configuration
	// The API key can be set via OPUSDNS_API_KEY environment variable
	c, err := client.NewClient(
		client.WithAPIKey(os.Getenv("OPUSDNS_API_KEY")),
		client.WithDebug(os.Getenv("OPUSDNS_DEBUG") == "true"),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Example 1: List all DNS zones
	fmt.Println("=== Listing DNS Zones ===")
	listZonesExample(ctx, c)

	// Example 2: Create and manage a zone (if DEMO_ZONE is set)
	if demoZone := os.Getenv("DEMO_ZONE"); demoZone != "" {
		fmt.Println("\n=== Zone Management Demo ===")
		zoneManagementExample(ctx, c, demoZone)
	}

	// Example 3: Check domain availability
	fmt.Println("\n=== Domain Availability Check ===")
	availabilityExample(ctx, c)

	// Example 4: Error handling
	fmt.Println("\n=== Error Handling Example ===")
	errorHandlingExample(ctx, c)

	fmt.Println("\n=== Examples completed successfully ===")
}

func listZonesExample(ctx context.Context, c *client.Client) {
	// List all zones with automatic pagination
	zones, err := c.DNS.ListZones(ctx, nil)
	if err != nil {
		log.Printf("Failed to list zones: %v", err)
		return
	}

	fmt.Printf("Found %d zones:\n", len(zones))
	for _, zone := range zones {
		createdOn := "unknown"
		if zone.CreatedOn != nil {
			createdOn = zone.CreatedOn.Format(time.RFC3339)
		}
		fmt.Printf("  - %s (DNSSEC: %s, Created: %s)\n",
			zone.Name,
			zone.DNSSECStatus,
			createdOn,
		)
	}

	// Example of paginated listing
	fmt.Println("\nPaginated listing (first page only):")
	resp, err := c.DNS.ListZonesPage(ctx, &models.ListZonesOptions{
		Page:      1,
		PageSize:  5,
		SortBy:    models.ZoneSortByCreatedOn,
		SortOrder: models.SortDesc,
	})
	if err != nil {
		log.Printf("Failed to list zones page: %v", err)
		return
	}

	fmt.Printf("  Page %d of %d (total: %d zones)\n",
		resp.Pagination.CurrentPage,
		resp.Pagination.TotalPages,
		resp.Pagination.TotalCount,
	)
}

func zoneManagementExample(ctx context.Context, c *client.Client, zoneName string) {
	// Create a new zone
	fmt.Printf("Creating zone: %s\n", zoneName)
	zone, err := c.DNS.CreateZone(ctx, &models.ZoneCreateRequest{
		Name: zoneName,
		RRSets: []models.RRSetCreate{
			{
				Name:    "www",
				Type:    models.RRSetTypeA,
				TTL:     3600,
				Records: []string{"192.0.2.1"},
			},
		},
	})
	if err != nil {
		log.Printf("Failed to create zone: %v", err)
		return
	}
	fmt.Printf("✓ Zone created: %s\n", zone.Name)

	// Add a TXT record
	fmt.Println("Adding TXT record...")
	err = c.DNS.UpsertRecord(ctx, zoneName, models.Record{
		Name:  "_test",
		Type:  models.RRSetTypeTXT,
		TTL:   300,
		RData: "hello-from-opusdns-go-client",
	})
	if err != nil {
		log.Printf("Failed to add TXT record: %v", err)
	} else {
		fmt.Println("✓ TXT record added")
	}

	// Batch operations
	fmt.Println("Performing batch operations...")
	err = c.DNS.PatchRecords(ctx, zoneName, []models.RecordOperation{
		{
			Op: models.RecordOpUpsert,
			Record: models.Record{
				Name:  "mail",
				Type:  models.RRSetTypeA,
				TTL:   3600,
				RData: "192.0.2.10",
			},
		},
		{
			Op: models.RecordOpUpsert,
			Record: models.Record{
				Name:  "@",
				Type:  models.RRSetTypeMX,
				TTL:   3600,
				RData: "10 mail." + zoneName + ".",
			},
		},
	})
	if err != nil {
		log.Printf("Failed batch operations: %v", err)
	} else {
		fmt.Println("✓ Batch operations completed")
	}

	// Get zone details with records
	fmt.Println("Fetching zone details...")
	zone, err = c.DNS.GetZone(ctx, zoneName)
	if err != nil {
		log.Printf("Failed to get zone: %v", err)
	} else {
		fmt.Printf("✓ Zone %s has %d RRSets\n", zone.Name, len(zone.RRSets))
		for _, rrset := range zone.RRSets {
			fmt.Printf("    %s %s (TTL: %d)\n", rrset.Name, rrset.Type, rrset.TTL)
		}
	}

	// Clean up - delete the zone
	fmt.Println("Cleaning up (deleting zone)...")
	err = c.DNS.DeleteZone(ctx, zoneName)
	if err != nil {
		log.Printf("Failed to delete zone: %v", err)
	} else {
		fmt.Println("✓ Zone deleted")
	}
}

func availabilityExample(ctx context.Context, c *client.Client) {
	domains := []string{"example.com", "opusdns-test-12345.com", "google.com"}

	fmt.Printf("Checking availability for: %v\n", domains)

	result, err := c.Availability.CheckAvailability(ctx, domains)
	if err != nil {
		log.Printf("Failed to check availability: %v", err)
		return
	}

	fmt.Printf("Results (processed in %dms):\n", result.Meta.ProcessingTimeMs)
	for _, avail := range result.Results {
		status := "❌ unavailable"
		if avail.Status.IsAvailable() {
			status = "✓ available"
		}
		fmt.Printf("  %s: %s (%s)\n", avail.Domain, status, avail.Status)
	}
}

func errorHandlingExample(ctx context.Context, c *client.Client) {
	// Try to get a non-existent zone
	_, err := c.DNS.GetZone(ctx, "this-zone-definitely-does-not-exist-12345.com")
	if err != nil {
		if client.IsNotFoundError(err) {
			fmt.Println("✓ Correctly caught NotFound error")
		} else if apiErr, ok := client.IsAPIError(err); ok {
			fmt.Printf("✓ Caught API error: HTTP %d - %s\n", apiErr.StatusCode, apiErr.Message)
		} else {
			fmt.Printf("✓ Caught error: %v\n", err)
		}
	}

	// Demonstrate error type checking
	fmt.Println("\nError type checking examples:")
	fmt.Printf("  IsNotFoundError: %v\n", client.IsNotFoundError(err))
	fmt.Printf("  IsUnauthorizedError: %v\n", client.IsUnauthorizedError(err))
	fmt.Printf("  IsRetryableError: %v\n", client.IsRetryableError(err))
}
