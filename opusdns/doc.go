// Package client provides a Go client library for the OpusDNS API.
//
// OpusDNS is a DNS and domain management platform. This client library provides
// a convenient way to interact with the OpusDNS API from Go applications.
//
// # Installation
//
//	go get github.com/opusdns/opusdns-go-client
//
// # Quick Start
//
// Create a client using your API key:
//
//	client, err := client.NewClient(
//	    client.WithAPIKey("opk_your_api_key_here"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// List all DNS zones
//	ctx := context.Background()
//	zones, err := client.DNS.ListZones(ctx, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, zone := range zones {
//	    fmt.Println(zone.Name)
//	}
//
// # Configuration
//
// The client can be configured using functional options:
//
//	c, err := client.NewClient(
//	    client.WithAPIKey("opk_..."),
//	    client.WithAPIEndpoint("https://api.opusdns.com"),
//	    client.WithHTTPTimeout(60 * time.Second),
//	    client.WithMaxRetries(5),
//	    client.WithDebug(true),
//	)
//
// Configuration can also be loaded from environment variables:
//   - OPUSDNS_API_KEY: Your API key
//   - OPUSDNS_API_ENDPOINT: Custom API endpoint (optional)
//   - OPUSDNS_DEBUG: Enable debug logging ("true" or "1")
//
// # Services
//
// The client provides access to various services:
//
//   - client.DNS: Zone and record management
//   - client.Domains: Domain registration and management
//   - client.Contacts: Contact management
//   - client.EmailForwards: Email forwarding configuration
//   - client.DomainForwards: Domain/URL forwarding configuration
//   - client.TLDs: TLD information and portfolio
//   - client.Availability: Domain availability checking
//   - client.Organizations: Organization management
//   - client.Users: User management
//   - client.Events: Event and audit log access
//
// # Error Handling
//
// The client returns detailed errors that can be inspected:
//
//	zone, err := c.DNS.GetZone(ctx, "example.com")
//	if err != nil {
//	    if errors.Is(err, client.ErrNotFound) {
//	        // Zone doesn't exist
//	        log.Println("Zone not found")
//	    } else if errors.Is(err, client.ErrUnauthorized) {
//	        // Invalid API key
//	        log.Println("Invalid API key")
//	    } else if apiErr, ok := client.IsAPIError(err); ok {
//	        // Handle specific API error
//	        log.Printf("API error %d: %s", apiErr.StatusCode, apiErr.Message)
//	    } else {
//	        // Other error (network, etc.)
//	        log.Printf("Error: %v", err)
//	    }
//	}
//
// # Pagination
//
// List methods support both automatic pagination and page-by-page access:
//
//	// Automatic pagination (fetches all pages)
//	allZones, err := c.DNS.ListZones(ctx, nil)
//
//	// Page-by-page access
//	opts := &models.ListZonesOptions{
//	    Page:     1,
//	    PageSize: 50,
//	}
//	resp, err := c.DNS.ListZonesPage(ctx, opts)
//	fmt.Printf("Page %d of %d\n", resp.Pagination.CurrentPage, resp.Pagination.TotalPages)
//
// # DNS Record Management
//
// Create, update, and delete DNS records:
//
//	// Create a zone
//	zone, err := c.DNS.CreateZone(ctx, &models.ZoneCreateRequest{
//	    Name: "example.com",
//	})
//
//	// Add a record
//	err = c.DNS.UpsertRecord(ctx, "example.com", models.Record{
//	    Name:  "www",
//	    Type:  models.RRSetTypeA,
//	    TTL:   3600,
//	    RData: "192.0.2.1",
//	})
//
//	// Batch operations
//	err = c.DNS.PatchRecords(ctx, "example.com", []models.RecordOperation{
//	    {Op: models.RecordOpUpsert, Record: models.Record{Name: "www", Type: models.RRSetTypeAAAA, TTL: 3600, RData: "2001:db8::1"}},
//	    {Op: models.RecordOpRemove, Record: models.Record{Name: "old", Type: models.RRSetTypeCNAME, TTL: 3600, RData: "legacy.example.com."}},
//	})
//
// # Domain Registration
//
// Register and manage domains:
//
//	// Check availability
//	avail, err := c.Availability.CheckSingleAvailability(ctx, "example.com")
//	if avail.Status.IsAvailable() {
//	    // Register the domain
//	    domain, err := c.Domains.CreateDomain(ctx, &models.DomainCreateRequest{
//	        Name:   "example.com",
//	        Period: 1,
//	        Contacts: map[models.DomainContactType]models.ContactHandle{
//	            models.DomainContactTypeRegistrant: {ContactID: contactID},
//	        },
//	    })
//	}
//
// # Thread Safety
//
// The client is safe for concurrent use by multiple goroutines.
//
// # Rate Limiting
//
// The client automatically handles rate limiting (HTTP 429) by retrying with
// exponential backoff. You can configure the retry behavior:
//
//	c, err := client.NewClient(
//	    client.WithAPIKey("opk_..."),
//	    client.WithMaxRetries(5),
//	    client.WithRetryWait(1*time.Second, 30*time.Second),
//	)
//
// # API Documentation
//
// For complete API documentation, visit https://developers.opusdns.com
package opusdns
