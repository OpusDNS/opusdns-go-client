# OpusDNS Go Client Library

A Go client library for the [OpusDNS](https://opusdns.com) DNS API with support for zone management, DNSSEC, and DNS record operations.

[![Go Reference](https://pkg.go.dev/badge/github.com/opusdns/opusdns-go-client.svg)](https://pkg.go.dev/github.com/opusdns/opusdns-go-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/opusdns/opusdns-go-client)](https://goreportcard.com/report/github.com/opusdns/opusdns-go-client)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- ✅ **Complete API Coverage**: Zones, RRSets, DNSSEC operations
- ✅ **Automatic Zone Detection**: Finds the correct zone for any FQDN
- ✅ **Smart Retry Logic**: Exponential backoff for rate limiting and transient failures
- ✅ **ACME Support**: Convenience methods for DNS-01 challenges
- ✅ **Production-Ready**: Well-tested and documented

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
    // Create client with your API key
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
    // Required: Your OpusDNS API key (format: opk_...)
    APIKey: "opk_...",

    // Optional: API endpoint (default: https://api.opusdns.com)
    APIEndpoint: "https://api.opusdns.com",

    // Optional: Default TTL for records in seconds (default: 60)
    TTL: 60,

    // Optional: HTTP request timeout (default: 30s)
    HTTPTimeout: 30 * time.Second,

    // Optional: Max retries for transient failures (default: 3)
    MaxRetries: 3,
})
```

## Zone Operations

### List Zones

```go
zones, err := client.ListZones()
```

Returns all DNS zones with automatic pagination.

### Get Zone

```go
zone, err := client.GetZone("example.com")
fmt.Printf("Zone: %s, DNSSEC: %s\n", zone.Name, zone.DNSSECStatus)
```

### Create Zone

```go
// Create an empty zone
zone, err := client.CreateZone("example.com", nil)

// Create a zone with initial records
zone, err := client.CreateZone("example.com", []opusdns.RRSetCreateRequest{
    {
        Name:    "www",
        Type:    "A",
        TTL:     3600,
        Records: []string{"192.0.2.1"},
    },
})
```

### Delete Zone

```go
err := client.DeleteZone("example.com")
```

## DNSSEC Operations

### Enable DNSSEC

```go
changes, err := client.EnableDNSSEC("example.com")
fmt.Printf("DNSSEC enabled, %d changes made\n", changes.NumChanges)
```

### Disable DNSSEC

```go
changes, err := client.DisableDNSSEC("example.com")
```

## Record Operations

### Get RRSets

```go
rrsets, err := client.GetRRSets("example.com")
for _, rrset := range rrsets {
    fmt.Printf("%s %s %d\n", rrset.Name, rrset.Type, rrset.TTL)
    for _, record := range rrset.Records {
        fmt.Printf("  %s\n", record.RData)
    }
}
```

### Upsert Record

```go
err := client.UpsertRecord("example.com", opusdns.RRSet{
    Name:  "www",
    Type:  "A",
    TTL:   3600,
    RData: "192.0.2.1",
})
```

### Remove Record

```go
err := client.RemoveRecord("example.com", opusdns.RRSet{
    Name:  "www",
    Type:  "A",
    TTL:   3600,
    RData: "192.0.2.1",
})
```

### Batch Operations

```go
err := client.PatchRRSets("example.com", []opusdns.RRSetOperation{
    {
        Op: "upsert",
        Record: opusdns.RRSet{
            Name:  "www",
            Type:  "A",
            TTL:   3600,
            RData: "192.0.2.1",
        },
    },
    {
        Op: "remove",
        Record: opusdns.RRSet{
            Name:  "old",
            Type:  "CNAME",
            TTL:   3600,
            RData: "example.com.",
        },
    },
})
```

## ACME DNS-01 Challenge Support

### Add Challenge Record

```go
// Automatically finds the zone and creates the TXT record
err := client.UpsertTXTRecord("_acme-challenge.www.example.com", "token")
```

### Remove Challenge Record

```go
err := client.RemoveTXTRecord("_acme-challenge.www.example.com", "token")
```

### Find Zone for FQDN

```go
zone, err := client.FindZoneForFQDN("_acme-challenge.sub.example.com")
// Returns "example.com" if that zone exists
```

The zone detection iterates through domain parts and checks the API:
- `_acme-challenge.sub.example.com` → tries `sub.example.com`, then `example.com`

## Error Handling

```go
zones, err := client.ListZones()
if err != nil {
    if apiErr, ok := err.(*opusdns.APIError); ok {
        fmt.Printf("API Error %d: %s\n", apiErr.StatusCode, apiErr.Message)
    } else {
        fmt.Printf("Error: %v\n", err)
    }
}
```

### Common Errors

| HTTP | Description | Action |
|------|-------------|--------|
| 401 | Invalid API key | Verify credentials |
| 404 | Zone not found | Check zone exists |
| 429 | Rate limited | Auto-retried |
| 5xx | Server error | Auto-retried |

## Retry Logic

The client automatically retries on:
- HTTP 429 (Rate Limited)
- HTTP 5xx (Server Errors)

Uses exponential backoff: 1s, 2s, 4s (up to `MaxRetries` attempts).

## Testing

```bash
go test -v ./...
```

## Environment Variables

```bash
export OPUSDNS_API_KEY="opk_..."
export OPUSDNS_API_ENDPOINT="https://api.opusdns.com"  # optional
```

## Requirements

- Go 1.21+
- OpusDNS API key

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related Projects

- [certbot-dns-opusdns](https://github.com/opusdns/certbot-dns-opusdns) - Certbot plugin
- [libdns-opusdns](https://github.com/opusdns/libdns-opusdns) - libdns provider for Caddy
- [acme.sh](https://github.com/acmesh-official/acme.sh) - acme.sh integration (dns_opusdns.sh)
