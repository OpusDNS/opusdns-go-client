// Package opusdns provides a Go client library for the OpusDNS API.
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
//	c, err := opusdns.NewClient(
//	    opusdns.WithAPIKey("opk_your_api_key_here"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// List all DNS zones
//	ctx := context.Background()
//	zones, err := c.DNS.ListZones(ctx, nil)
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
//	c, err := opusdns.NewClient(
//	    opusdns.WithAPIKey("opk_..."),
//	    opusdns.WithAPIEndpoint("https://api.opusdns.com"),
//	    opusdns.WithHTTPTimeout(60 * time.Second),
//	    opusdns.WithMaxRetries(5),
//	    opusdns.WithDebug(true),
//	)
//
// Configuration can also be loaded from environment variables:
//   - OPUSDNS_API_KEY: Your API key
//   - OPUSDNS_API_ENDPOINT: Custom API endpoint (optional)
//   - OPUSDNS_DEBUG: Enable debug logging ("true" or "1")
//
// Debug logging goes to standard error unless a Logger is set, and masks
// secrets such as API keys, client secrets, tokens and auth codes.
//
// # Services
//
// The client provides access to various services:
//
//   - c.DNS: Zone and record management
//   - c.Domains: Domain registration and management
//   - c.Contacts: Contact management
//   - c.EmailForwards: Email forwarding configuration
//   - c.DomainForwards: Domain/URL forwarding configuration
//   - c.TLDs: TLD information and portfolio
//   - c.Availability: Domain availability checking
//   - c.Organizations: Organization management
//   - c.Users: User management
//   - c.Events: Event and audit log access
//
// # Error Handling
//
// The client returns detailed errors that can be inspected:
//
//	zone, err := c.DNS.GetZone(ctx, "example.com")
//	if err != nil {
//	    if errors.Is(err, opusdns.ErrNotFound) {
//	        // Zone doesn't exist
//	        log.Println("Zone not found")
//	    } else if errors.Is(err, opusdns.ErrUnauthorized) {
//	        // Invalid API key
//	        log.Println("Invalid API key")
//	    } else if apiErr, ok := opusdns.IsAPIError(err); ok {
//	        // Handle specific API error
//	        log.Printf("API error %d [%s]: %s", apiErr.StatusCode, apiErr.ErrorCode, apiErr.Message)
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
//	// Create a zone, then read it back
//	_, err = c.DNS.CreateZone(ctx, &models.ZoneCreateRequest{
//	    Name: "example.com",
//	})
//	zone, err := c.DNS.GetZone(ctx, "example.com")
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
//	        Name:        "example.com",
//	        Period:      models.DomainPeriod{Value: 1, Unit: models.PeriodUnitYear},
//	        RenewalMode: models.RenewalModeRenew,
//	        Contacts: map[models.DomainContactType][]models.ContactHandle{
//	            models.DomainContactTypeRegistrant: {{ContactID: contactID}},
//	        },
//	    })
//	}
//
// # Thread Safety
//
// The client is safe for concurrent use by multiple goroutines.
//
// # Rate Limiting and Retries
//
// The client retries a rate limited request (HTTP 429) after the wait the
// API asks for in Retry-After, or with exponential backoff when it names
// none. A 429 holds back only the request that received it. When the API asks
// to wait longer than RetryWaitMax, the client returns the error at once, with
// the requested wait in APIError.RetryAfter. Server errors and network failures
// are retried only for GET, PUT and other methods that are safe to send twice;
// POST, PATCH and DELETE are not resent once the API may have acted on them.
// You can configure the retry behavior:
//
//	c, err := opusdns.NewClient(
//	    opusdns.WithAPIKey("opk_..."),
//	    opusdns.WithMaxRetries(5),
//	    opusdns.WithRetryWait(1*time.Second, 30*time.Second),
//	)
//
// # API Documentation
//
// For complete API documentation, visit https://developers.opusdns.com
package opusdns
