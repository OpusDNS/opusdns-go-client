package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/opusdns/opusdns-go-client/opusdns"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create client
	client, err := opusdns.NewClient(
		opusdns.WithAPIKey(os.Getenv("OPUSDNS_API_KEY")),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Example 1: List all domains
	fmt.Println("=== Listing Domains ===")
	listDomainsExample(ctx, client)

	// Example 2: Get domain summary
	fmt.Println("\n=== Domain Summary ===")
	domainSummaryExample(ctx, client)

	// Example 3: Check domain availability
	fmt.Println("\n=== Domain Availability ===")
	checkAvailabilityExample(ctx, client)

	// Example 4: Get TLD information
	fmt.Println("\n=== TLD Information ===")
	tldInfoExample(ctx, client)

	fmt.Println("\n=== Examples completed ===")
}

func listDomainsExample(ctx context.Context, client *opusdns.Client) {
	// List domains with filtering and sorting
	domains, err := client.Domains.ListDomainsPage(ctx, &models.ListDomainsOptions{
		Page:      1,
		PageSize:  10,
		SortBy:    models.DomainSortByExpiresOn,
		SortOrder: models.SortAsc,
	})
	if err != nil {
		log.Printf("Failed to list domains: %v", err)
		return
	}

	fmt.Printf("Found %d domains (page 1 of %d):\n",
		len(domains.Results),
		domains.Pagination.TotalPages,
	)

	for _, domain := range domains.Results {
		expiresOn := "N/A"
		if domain.ExpiresOn != nil {
			expiresOn = domain.ExpiresOn.Format("2006-01-02")
		}
		autoRenew := "off"
		if domain.AutoRenew {
			autoRenew = "on"
		}
		fmt.Printf("  - %s (expires: %s, auto-renew: %s)\n",
			domain.Name,
			expiresOn,
			autoRenew,
		)
	}
}

func domainSummaryExample(ctx context.Context, client *opusdns.Client) {
	summary, err := client.Domains.GetSummary(ctx)
	if err != nil {
		log.Printf("Failed to get domain summary: %v", err)
		return
	}

	fmt.Printf("Total domains: %d\n", summary.TotalDomains)
	fmt.Printf("Expiring within 30 days: %d\n", summary.ExpiringWithin30Days)
	fmt.Printf("Expiring within 90 days: %d\n", summary.ExpiringWithin90Days)

	if len(summary.DomainsByTLD) > 0 {
		fmt.Println("Domains by TLD:")
		for tld, count := range summary.DomainsByTLD {
			fmt.Printf("  .%s: %d\n", tld, count)
		}
	}
}

func checkAvailabilityExample(ctx context.Context, client *opusdns.Client) {
	domainsToCheck := []string{
		"example-test-domain-12345.com",
		"example-test-domain-12345.de",
		"example-test-domain-12345.io",
		"google.com",
	}

	result, err := client.Availability.CheckAvailability(ctx, domainsToCheck)
	if err != nil {
		log.Printf("Failed to check availability: %v", err)
		return
	}

	fmt.Printf("Checked %d domains in %dms:\n", result.Meta.Total, result.Meta.ProcessingTimeMs)
	for _, avail := range result.Results {
		statusIcon := "❌"
		if avail.Status.IsAvailable() {
			statusIcon = "✓"
		}
		fmt.Printf("  %s %s (%s)\n", statusIcon, avail.Domain, avail.Status)
	}
}

func tldInfoExample(ctx context.Context, client *opusdns.Client) {
	// Get TLD portfolio
	portfolio, err := client.TLDs.GetPortfolio(ctx)
	if err != nil {
		log.Printf("Failed to get TLD portfolio: %v", err)
		return
	}

	fmt.Printf("Available TLDs in portfolio: %d\n", portfolio.Total)

	// Get details for a specific TLD
	tldDetails, err := client.TLDs.GetTLD(ctx, "com")
	if err != nil {
		log.Printf("Failed to get TLD details: %v", err)
		return
	}

	fmt.Printf("\nDetails for .%s:\n", tldDetails.Name)
	fmt.Printf("  Type: %s\n", tldDetails.Type)
	fmt.Printf("  DNSSEC Supported: %v\n", tldDetails.DNSSECSupported)
	fmt.Printf("  IDN Supported: %v\n", tldDetails.IDNSupported)
	fmt.Printf("  Registration Period: %d-%d years\n",
		tldDetails.MinRegistrationPeriod,
		tldDetails.MaxRegistrationPeriod,
	)

	if tldDetails.Pricing != nil {
		fmt.Printf("  Registration Price: %s %s\n",
			tldDetails.Pricing.RegisterPrice,
			tldDetails.Pricing.Currency,
		)
	}
}
