# OpusDNS Go Client

A Go client library for the [OpusDNS](https://opusdns.com) DNS API.

📚 **API Documentation**: [developers.opusdns.com](https://developers.opusdns.com)

## Installation

```bash
go get github.com/opusdns/opusdns-go-client
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    opusdns "github.com/opusdns/opusdns-go-client"
)

func main() {
    client := opusdns.NewClient(&opusdns.Config{
        APIKey: "opk_...", // Your OpusDNS API key
    })

    // List all zones
    zones, err := client.ListZones()
    if err != nil {
        log.Fatal(err)
    }

    for _, zone := range zones {
        fmt.Printf("Zone: %s (DNSSEC: %s)\n", zone.Name, zone.DNSSECStatus)
    }
}
```

## Configuration

```go
client := opusdns.NewClient(&opusdns.Config{
    APIKey:      "opk_...",                       // Required
    APIEndpoint: "https://api.opusdns.com",       // Optional (default)
    TTL:         60,                              // Optional: default TTL for records
    HTTPTimeout: 30 * time.Second,                // Optional: request timeout
    MaxRetries:  3,                               // Optional: retry count for 429/5xx
})
```

## Zone Operations

```go
// List all zones
zones, err := client.ListZones()

// Get a specific zone
zone, err := client.GetZone("example.com")

// Create a zone (empty)
zone, err := client.CreateZone("example.com", nil)

// Create a zone with initial records
zone, err := client.CreateZone("example.com", []opusdns.Record{
    {Name: "www", Type: "A", TTL: 3600, RData: "192.0.2.1"},
    {Name: "www", Type: "AAAA", TTL: 3600, RData: "2001:db8::1"},
})

// Delete a zone
err := client.DeleteZone("example.com")
```

## Record Operations

```go
// Get all record sets
rrsets, err := client.GetRecords("example.com")

// Create or update a record
err := client.UpsertRecord("example.com", opusdns.Record{
    Name:  "www",
    Type:  "A",
    TTL:   3600,
    RData: "192.0.2.1",
})

// Delete a record
err := client.DeleteRecord("example.com", opusdns.Record{
    Name:  "www",
    Type:  "A",
    TTL:   3600,
    RData: "192.0.2.1",
})

// Batch operations
err := client.PatchRecords("example.com", []opusdns.RecordOperation{
    {Op: "upsert", Record: opusdns.Record{Name: "www", Type: "A", TTL: 3600, RData: "192.0.2.1"}},
    {Op: "remove", Record: opusdns.Record{Name: "old", Type: "CNAME", TTL: 3600, RData: "legacy.example.com."}},
})
```

## DNSSEC

```go
// Enable DNSSEC
changes, err := client.EnableDNSSEC("example.com")
fmt.Printf("DNSSEC enabled: %d changes\n", changes.NumChanges)

// Disable DNSSEC
changes, err := client.DisableDNSSEC("example.com")
```

## Error Handling

```go
zones, err := client.ListZones()
if err != nil {
    if apiErr, ok := err.(*opusdns.APIError); ok {
        fmt.Printf("API error %d: %s\n", apiErr.StatusCode, apiErr.Message)
    } else {
        fmt.Printf("Error: %v\n", err)
    }
}
```

| Status | Description | Behavior |
|--------|-------------|----------|
| 401 | Invalid API key | Returns error |
| 404 | Resource not found | Returns error |
| 429 | Rate limited | Auto-retry with backoff |
| 5xx | Server error | Auto-retry with backoff |

## Requirements

- Go 1.21+
- OpusDNS API key ([Get one here](https://opusdns.com))

## License

MIT License - see [LICENSE](LICENSE)
