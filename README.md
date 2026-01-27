# OpusDNS Go Client

[![Go Reference](https://pkg.go.dev/badge/github.com/opusdns/opusdns-go-client.svg)](https://pkg.go.dev/github.com/opusdns/opusdns-go-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/opusdns/opusdns-go-client)](https://goreportcard.com/report/github.com/opusdns/opusdns-go-client)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

The official Go client library for the [OpusDNS](https://opusdns.com) API - a comprehensive DNS and domain management platform.

## Features

- **Complete API Coverage**: Full support for DNS zones, domains, contacts, email forwarding, domain forwarding, and more
- **Type-Safe**: Strongly typed models with Go idioms
- **Automatic Pagination**: Easily iterate through all resources
- **Retry Logic**: Built-in exponential backoff for transient failures
- **Rate Limit Handling**: Automatic handling of rate limits (HTTP 429)
- **Context Support**: All methods accept `context.Context` for cancellation and timeouts
- **Configurable**: Flexible configuration via functional options or environment variables
- **Debug Mode**: Optional request/response logging for troubleshooting

## Installation

```bash
go get github.com/opusdns/opusdns-go-client
```

Requires Go 1.25 or later.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/opusdns/opusdns-go-client/models"
    "github.com/opusdns/opusdns-go-client/opusdns"
)

func main() {
    // Create a client with your API key
    client, err := opusdns.NewClient(
        opusdns.WithAPIKey("opk_your_api_key_here"),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // List all DNS zones
    zones, err := client.DNS.ListZones(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }

    for _, zone := range zones {
        fmt.Printf("Zone: %s (DNSSEC: %s)\n", zone.Name, zone.DNSSECStatus)
    }
}
```

## Configuration

### Using Functional Options

```go
client, err := opusdns.NewClient(
    opusdns.WithAPIKey("opk_..."),
    opusdns.WithAPIEndpoint("https://api.opusdns.com"),
    opusdns.WithHTTPTimeout(60 * time.Second),
    opusdns.WithMaxRetries(5),
    opusdns.WithRetryWait(1*time.Second, 30*time.Second),
    opusdns.WithDebug(true),
)
```

### Using Environment Variables

The client automatically reads from environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `OPUSDNS_API_KEY` | Your API key (required) | - |
| `OPUSDNS_API_ENDPOINT` | Custom API endpoint | `https://api.opusdns.com` |
| `OPUSDNS_API_VERSION` | API version | `v1` |
| `OPUSDNS_DEBUG` | Enable debug logging (`true`/`1`) | `false` |

```go
// API key is read from OPUSDNS_API_KEY environment variable
client, err := opusdns.NewClient()
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithAPIKey(key)` | Set the API key | - |
| `WithAPIEndpoint(url)` | Set custom API endpoint | `https://api.opusdns.com` |
| `WithAPIVersion(version)` | Set API version | `v1` |
| `WithHTTPTimeout(duration)` | HTTP request timeout | `30s` |
| `WithMaxRetries(n)` | Max retries for transient failures | `3` |
| `WithRetryWait(min, max)` | Retry backoff bounds | `1s`, `30s` |
| `WithHTTPClient(client)` | Use custom HTTP client | - |
| `WithUserAgent(ua)` | Custom User-Agent string | `opusdns-go-client/1.0.0` |
| `WithDebug(enabled)` | Enable debug logging | `false` |
| `WithLogger(logger)` | Custom logger for debug output | stdout |
| `WithTTL(ttl)` | Default TTL for DNS records | `60` |

## Services

The client provides access to the following services:

| Service | Description |
|---------|-------------|
| `client.DNS` | DNS zone and record management |
| `client.Domains` | Domain registration, transfer, and renewal |
| `client.Contacts` | Contact (registrant/admin/tech) management |
| `client.EmailForwards` | Email forwarding configuration |
| `client.DomainForwards` | Domain/URL forwarding (redirects) |
| `client.TLDs` | TLD information and pricing |
| `client.Availability` | Domain availability checking |
| `client.Organizations` | Organization and billing management |
| `client.Users` | User management |
| `client.Events` | Event and audit log access |

## DNS Management

### List Zones

```go
// List all zones (automatic pagination)
zones, err := client.DNS.ListZones(ctx, nil)

// List with filtering and sorting
zones, err := client.DNS.ListZones(ctx, &models.ListZonesOptions{
    Search:       "example",
    DNSSECStatus: models.DNSSECStatusEnabled,
    SortBy:       models.ZoneSortByCreatedOn,
    SortOrder:    models.SortDesc,
})

// Paginated access
resp, err := client.DNS.ListZonesPage(ctx, &models.ListZonesOptions{
    Page:     1,
    PageSize: 50,
})
fmt.Printf("Page %d of %d\n", resp.Pagination.CurrentPage, resp.Pagination.TotalPages)
```

### Create a Zone

```go
zone, err := client.DNS.CreateZone(ctx, &models.ZoneCreateRequest{
    Name: "example.com",
    RRSets: []models.RRSetCreate{
        {
            Name:    "www",
            Type:    models.RRSetTypeA,
            TTL:     3600,
            Records: []models.RecordCreate{{RData: "192.0.2.1"}, {RData: "192.0.2.2"}},
        },
        {
            Name:    "@",
            Type:    models.RRSetTypeMX,
            TTL:     3600,
            Records: []models.RecordCreate{{RData: "10 mail.example.com."}},
        },
    },
})
```

### Manage Records

```go
// Add or update a single record
err := client.DNS.UpsertRecord(ctx, "example.com", models.Record{
    Name:  "www",
    Type:  models.RRSetTypeA,
    TTL:   3600,
    RData: "192.0.2.1",
})

// Delete a record
err := client.DNS.DeleteRecord(ctx, "example.com", models.Record{
    Name:  "www",
    Type:  models.RRSetTypeA,
    TTL:   3600,
    RData: "192.0.2.1",
})

// Batch operations (atomic)
err := client.DNS.PatchRecords(ctx, "example.com", []models.RecordOperation{
    {
        Op: models.RecordOpUpsert,
        Record: models.Record{
            Name:  "api",
            Type:  models.RRSetTypeA,
            TTL:   300,
            RData: "192.0.2.10",
        },
    },
    {
        Op: models.RecordOpRemove,
        Record: models.Record{
            Name:  "old-api",
            Type:  models.RRSetTypeCNAME,
            TTL:   3600,
            RData: "legacy.example.com.",
        },
    },
})
```

### DNSSEC

```go
// Enable DNSSEC
changes, err := client.DNS.EnableDNSSEC(ctx, "example.com")
fmt.Printf("DNSSEC enabled with %d changes\n", changes.NumChanges)

// Disable DNSSEC
changes, err := client.DNS.DisableDNSSEC(ctx, "example.com")
```

## Domain Registration

### Check Availability

```go
// Check multiple domains
result, err := client.Availability.CheckAvailability(ctx, []string{
    "example.com",
    "example.de",
    "example.io",
})

for _, avail := range result.Results {
    if avail.Status.IsAvailable() {
        fmt.Printf("%s is available!\n", avail.Domain)
    }
}

// Check single domain
avail, err := client.Availability.CheckSingleAvailability(ctx, "example.com")
```

### Register a Domain

```go
// First, create a contact
contact, err := client.Contacts.CreateContact(ctx, &models.ContactCreateRequest{
    FirstName:  "John",
    LastName:   "Doe",
    Email:      "john@example.com",
    Phone:      "+1.5551234567",
    Street:     "123 Main Street",
    City:       "New York",
    PostalCode: "10001",
    Country:    "US",
    Disclose:   false,
})

// Then register the domain
domain, err := client.Domains.CreateDomain(ctx, &models.DomainCreateRequest{
    Name:   "example.com",
    Period: 1, // 1 year
    Contacts: map[models.DomainContactType]models.ContactHandle{
        models.DomainContactTypeRegistrant: {ContactID: contact.ContactID},
        models.DomainContactTypeAdmin:      {ContactID: contact.ContactID},
        models.DomainContactTypeTech:       {ContactID: contact.ContactID},
    },
    Nameservers: []models.Nameserver{
        {Hostname: "ns1.opusdns.com"},
        {Hostname: "ns2.opusdns.com"},
    },
    TransferLock: models.BoolPtr(true),
    RenewMode:    models.RenewModePtr(models.RenewModeRenew),
})
```

### Transfer a Domain

```go
domain, err := client.Domains.TransferDomain(ctx, &models.DomainTransferRequest{
    Name:     "example.com",
    AuthCode: "abc123xyz",
    Contacts: map[models.DomainContactType]models.ContactHandle{
        models.DomainContactTypeRegistrant: {ContactID: contactID},
    },
})
```

### Renew a Domain

```go
domain, err := client.Domains.RenewDomain(ctx, "example.com", &models.DomainRenewRequest{
    Period: 2, // Renew for 2 years
})
```

## Email Forwarding

```go
// Create email forwarding for a domain
emailFwd, err := client.EmailForwards.CreateEmailForward(ctx, &models.EmailForwardCreateRequest{
    Hostname: "example.com",
    Aliases: []models.EmailForwardAliasCreate{
        {
            LocalPart:    "info",
            Destinations: []string{"john@gmail.com"},
        },
        {
            LocalPart:    "*", // Catch-all
            Destinations: []string{"catchall@company.com"},
        },
    },
})

// Add another alias
alias, err := client.EmailForwards.CreateAlias(ctx, emailFwd.EmailForwardID, &models.EmailForwardAliasCreate{
    LocalPart:    "support",
    Destinations: []string{"support@company.com", "backup@company.com"},
})
```

## Domain Forwarding (URL Redirects)

```go
forward, err := client.DomainForwards.CreateDomainForward(ctx, &models.DomainForwardCreateRequest{
    Hostname: "old-domain.com",
    Configs: []models.DomainForwardConfigCreate{
        {
            Protocol:       models.DomainForwardProtocolHTTP,
            DestinationURL: "https://new-domain.com",
            ForwardType:    models.DomainForwardTypePermanent, // 301 redirect
            IncludePath:    true,
            IncludeQuery:   true,
        },
        {
            Protocol:       models.DomainForwardProtocolHTTPS,
            DestinationURL: "https://new-domain.com",
            ForwardType:    models.DomainForwardTypePermanent,
            IncludePath:    true,
            IncludeQuery:   true,
        },
    },
})
```

## Error Handling

The client provides detailed error types for different failure scenarios:

```go
zone, err := client.DNS.GetZone(ctx, "example.com")
if err != nil {
    // Check for specific error types
    if errors.Is(err, opusdns.ErrNotFound) {
        fmt.Println("Zone not found")
    } else if errors.Is(err, opusdns.ErrUnauthorized) {
        fmt.Println("Invalid API key")
    } else if errors.Is(err, opusdns.ErrRateLimited) {
        fmt.Println("Rate limited - try again later")
    } else if errors.Is(err, opusdns.ErrForbidden) {
        fmt.Println("Insufficient permissions")
    }

    // Get detailed API error information
    if apiErr, ok := opusdns.IsAPIError(err); ok {
        fmt.Printf("Status: %d\n", apiErr.StatusCode)
        fmt.Printf("Error Code: %s\n", apiErr.ErrorCode)
        fmt.Printf("Message: %s\n", apiErr.Message)
        fmt.Printf("Request ID: %s\n", apiErr.RequestID)
    }

    // Check if error is retryable
    if opusdns.IsRetryableError(err) {
        fmt.Println("This error is retryable")
    }
}
```

### Error Types

| Error | Description |
|-------|-------------|
| `ErrNotFound` | Resource not found (HTTP 404) |
| `ErrUnauthorized` | Invalid or missing API key (HTTP 401) |
| `ErrForbidden` | Insufficient permissions (HTTP 403) |
| `ErrBadRequest` | Invalid request (HTTP 400) |
| `ErrConflict` | Resource conflict (HTTP 409) |
| `ErrRateLimited` | Rate limit exceeded (HTTP 429) |
| `ErrServerError` | Server error (HTTP 5xx) |
| `ErrTimeout` | Request timeout |
| `ErrZoneNotFound` | No matching zone for FQDN |
| `ErrInvalidInput` | Input validation failed |

### Helper Functions

```go
opusdns.IsNotFoundError(err)      // Check for 404
opusdns.IsUnauthorizedError(err)  // Check for 401
opusdns.IsForbiddenError(err)     // Check for 403
opusdns.IsRateLimitError(err)     // Check for 429
opusdns.IsConflictError(err)      // Check for 409
opusdns.IsRetryableError(err)     // Check if retryable (429, 5xx)
opusdns.IsAPIError(err)           // Extract APIError details
```

## Thread Safety

The client is safe for concurrent use by multiple goroutines. All service methods are thread-safe.

```go
var wg sync.WaitGroup
for _, zoneName := range zoneNames {
    wg.Add(1)
    go func(name string) {
        defer wg.Done()
        zone, err := client.DNS.GetZone(ctx, name)
        // ...
    }(zoneName)
}
wg.Wait()
```

## Examples

See the [examples](examples/) directory for complete working examples:

- [Basic Usage](examples/basic/) - DNS zone management, availability checking
- [Domain Registration](examples/domains/) - Domain registration workflow

Run examples:

```bash
export OPUSDNS_API_KEY="opk_your_api_key"
cd examples/basic
go run main.go
```

## Requirements

- Go 1.25 or later
- OpusDNS API key ([Get one here](https://opusdns.com))

## API Documentation

For complete API documentation, visit [developers.opusdns.com](https://developers.opusdns.com).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [developers.opusdns.com](https://developers.opusdns.com)
- **Issues**: [GitHub Issues](https://github.com/opusdns/opusdns-go-client/issues)
- **Email**: support@opusdns.com