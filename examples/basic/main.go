package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/opusdns/opusdns-go-client"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPUSDNS_API_KEY")
	if apiKey == "" {
		log.Fatal("OPUSDNS_API_KEY environment variable is required")
	}

	// Get API endpoint (defaults to production)
	apiEndpoint := os.Getenv("OPUSDNS_API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "https://api.opusdns.com"
	}

	// Create client with custom configuration
	client := opusdns.NewClient(&opusdns.Config{
		APIKey:          apiKey,
		APIEndpoint:     apiEndpoint,
		TTL:             60,
		HTTPTimeout:     30 * time.Second,
		MaxRetries:      3,
		PollingInterval: 6 * time.Second,
		PollingTimeout:  60 * time.Second,
	})

	// Example 1: List all zones
	fmt.Println("=== Listing DNS Zones ===")
	zones, err := client.ListZones()
	if err != nil {
		log.Fatalf("Failed to list zones: %v", err)
	}

	fmt.Printf("Found %d zones:\n", len(zones))
	for _, zone := range zones {
		fmt.Printf("  - %s (DNSSEC: %s, Created: %s)\n",
			zone.Name,
			zone.DNSSECStatus,
			zone.CreatedOn.Format(time.RFC3339),
		)
	}

	// Example 2: Add ACME challenge record
	if len(zones) > 0 {
		// Use first zone for demonstration
		zoneName := zones[0].Name
		fqdn := fmt.Sprintf("_acme-challenge.%s", zoneName)
		challengeValue := fmt.Sprintf("example-challenge-%d", time.Now().Unix())

		fmt.Printf("\n=== ACME DNS-01 Challenge Workflow ===\n")
		fmt.Printf("Domain: %s\n", fqdn)
		fmt.Printf("Challenge Value: %s\n", challengeValue)

		// Step 1: Create TXT record
		fmt.Println("\n[1/3] Creating TXT record...")
		if err := client.UpsertTXTRecord(fqdn, challengeValue); err != nil {
			log.Fatalf("Failed to create TXT record: %v", err)
		}
		fmt.Println("✓ TXT record created successfully")

		// Step 2: Wait for DNS propagation
		fmt.Println("\n[2/3] Waiting for DNS propagation...")
		fmt.Println("Checking DNS servers: 8.8.8.8, 1.1.1.1")
		
		startTime := time.Now()
		if err := client.WaitForPropagation(fqdn, challengeValue); err != nil {
			log.Fatalf("DNS propagation failed: %v", err)
		}
		elapsed := time.Since(startTime)
		fmt.Printf("✓ DNS record propagated successfully (took %s)\n", elapsed.Round(time.Second))

		// Step 3: Clean up
		fmt.Println("\n[3/3] Cleaning up TXT record...")
		if err := client.RemoveTXTRecord(fqdn, "TXT"); err != nil {
			log.Printf("⚠ Warning: Failed to remove TXT record: %v", err)
		} else {
			fmt.Println("✓ TXT record removed successfully")
		}
	}

	// Example 3: Zone detection
	fmt.Println("\n=== Zone Detection Example ===")
	testFQDNs := []string{
		"_acme-challenge.example.com",
		"_acme-challenge.subdomain.example.com",
		"_acme-challenge.deep.subdomain.example.com",
	}

	for _, fqdn := range testFQDNs {
		zone, err := client.FindZoneForFQDN(fqdn)
		if err != nil {
			fmt.Printf("  %s -> ERROR: %v\n", fqdn, err)
		} else {
			fmt.Printf("  %s -> %s\n", fqdn, zone)
		}
	}

	// Example 4: Error handling
	fmt.Println("\n=== Error Handling Example ===")
	invalidFQDN := "_acme-challenge.nonexistent-zone-12345.com"
	fmt.Printf("Attempting to add record to non-existent zone: %s\n", invalidFQDN)
	
	if err := client.UpsertTXTRecord(invalidFQDN, "test-value"); err != nil {
		if apiErr, ok := err.(*opusdns.APIError); ok {
			fmt.Printf("✓ Caught API error (expected): HTTP %d - %s\n", apiErr.StatusCode, apiErr.Error())
		} else {
			fmt.Printf("✓ Caught error (expected): %v\n", err)
		}
	}

	fmt.Println("\n=== Example completed successfully ===")
}
