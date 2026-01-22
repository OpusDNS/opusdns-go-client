# OpusDNS Go Client Library

A production-ready Go client library for the OpusDNS API with comprehensive support for DNS zone and record management, including ACME DNS-01 challenge workflows.

[![Go Reference](https://pkg.go.dev/badge/github.com/opusdns/opusdns-go-client.svg)](https://pkg.go.dev/github.com/opusdns/opusdns-go-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/opusdns/opusdns-go-client)](https://goreportcard.com/report/github.com/opusdns/opusdns-go-client)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- ✅ **Complete API Coverage**: Zone listing, RRSet operations (upsert/remove)
- ✅ **Automatic Zone Detection**: Matches FQDN to the correct zone from your account
- ✅ **DNS Propagation Polling**: Verifies record propagation on public DNS servers
- ✅ **Smart Retry Logic**: Exponential backoff for rate limiting and transient failures
- ✅ **Zone Caching**: Minimizes API calls with intelligent caching (5-minute TTL)
- ✅ **Comprehensive Error Handling**: Clear error messages for 401, 404, 409, 429, 5xx
- ✅ **Production-Ready**: Well-tested, documented, and battle-tested for ACME workflows

## Installation

```bash
go get github.com/opusdns/opusdns-go-client
```

## Quick Start

```go
package main

import (
    "log"
    
    "github.com/opusdns/opusdns-go-client"
)

func main() {
    // Create client
    client := opusdns.NewClient(&opusdns.Config{
        APIKey:      "opk_your_api_key_here",
        APIEndpoint: "https://api.opusdns.com", // or https://sandbox.opusdns.com
    })

    // Add ACME challenge record
    fqdn := "_acme-challenge.example.com"
    value := "challenge-token-value"
    
    if err := client.UpsertTXTRecord(fqdn, value); err != nil {
        log.Fatalf("Failed to create TXT record: %v", err)
    }
    log.Println("TXT record created successfully")

    // Wait for DNS propagation
    if err := client.WaitForPropagation(fqdn, value); err != nil {
        log.Fatalf("DNS propagation failed: %v", err)
    }
    log.Println("DNS record propagated successfully")

    // Clean up (best-effort)
    if err := client.RemoveTXTRecord(fqdn, "TXT"); err != nil {
        log.Printf("Warning: Failed to remove TXT record: %v", err)
    } else {
        log.Println("TXT record removed successfully")
    }
}
```

## Configuration

### Basic Configuration

```go
config := &opusdns.Config{
    APIKey:      "opk_...",                      // Required: Your OpusDNS API key
    APIEndpoint: "https://api.opusdns.com",      // Optional: API endpoint (default: production)
}
```

### Advanced Configuration

```go
config := &opusdns.Config{
    APIKey:          "opk_...",
    APIEndpoint:     "https://api.opusdns.com",
    TTL:             60,                          // Record TTL in seconds (default: 60)
    HTTPTimeout:     30 * time.Second,            // HTTP request timeout (default: 30s)
    MaxRetries:      3,                           // Max retries for 429/5xx (default: 3)
    PollingInterval: 6 * time.Second,             // DNS check interval (default: 6s)
    PollingTimeout:  60 * time.Second,            // Max propagation wait (default: 60s)
    DNSResolvers:    []string{"8.8.8.8:53", "1.1.1.1:53"}, // DNS servers to check
}
```

## API Key Setup

### Production API Key

1. Log in to your OpusDNS dashboard at https://dashboard.opusdns.com
2. Navigate to **API Keys** section
3. Click **Create API Key**
4. Copy the generated key (format: `opk_{typeid}_{secret}{checksum}`, 67 characters)
5. Store securely (keys cannot be retrieved after creation)

### Sandbox API Key (for testing)

```bash
# Create sandbox API key via API
curl -X POST https://sandbox.opusdns.com/auth/client_credentials \
  -H "Authorization: Bearer <your_user_token>" \
  -H "Content-Type: application/json"
```

**Environment Variables:**

```bash
export OPUSDNS_API_KEY="opk_your_api_key_here"
export OPUSDNS_API_ENDPOINT="https://sandbox.opusdns.com"  # for testing
```

## Core API Methods

### List Zones

```go
zones, err := client.ListZones()
if err != nil {
    log.Fatal(err)
}

for _, zone := range zones {
    fmt.Printf("Zone: %s (DNSSEC: %s)\n", zone.Name, zone.DNSSECStatus)
}
```

### Find Zone for FQDN

Automatically detects the correct zone for a given FQDN:

```go
zone, err := client.FindZoneForFQDN("_acme-challenge.subdomain.example.com")
if err != nil {
    log.Fatal(err)
}
fmt.Println("Matched zone:", zone) // Output: example.com
```

**Zone Matching Algorithm:**
- Input: `_acme-challenge.sub.example.com`
- Candidates: `sub.example.com.`, `example.com.`
- Match: Longest matching zone (`example.com`)

### Upsert TXT Record

Creates or updates a TXT record (automatically quotes values):

```go
err := client.UpsertTXTRecord("_acme-challenge.example.com", "challenge-value")
if err != nil {
    log.Fatal(err)
}
```

### Remove TXT Record

```go
err := client.RemoveTXTRecord("_acme-challenge.example.com", "TXT")
if err != nil {
    log.Printf("Cleanup warning: %v", err)
}
```

### Wait for DNS Propagation

Polls public DNS servers to verify record propagation:

```go
err := client.WaitForPropagation("_acme-challenge.example.com", "challenge-value")
if err != nil {
    log.Fatal(err) // Timeout or verification failed
}
```

**Polling Strategy:**
- Queries: `8.8.8.8:53`, `1.1.1.1:53` (configurable)
- Interval: 6 seconds (configurable)
- Max attempts: 10 (60 seconds total)
- Returns immediately when record is found

## Error Handling

The library provides detailed error information:

```go
err := client.UpsertTXTRecord(fqdn, value)
if err != nil {
    if apiErr, ok := err.(*opusdns.APIError); ok {
        switch apiErr.StatusCode {
        case 401:
            log.Fatal("Invalid API key - check your credentials")
        case 404:
            log.Fatal("Zone not found - verify zone exists in your account")
        case 409:
            log.Fatal("Zone operation in progress - retry later")
        case 429:
            log.Fatal("Rate limited - client retries automatically")
        default:
            log.Fatalf("API error: %v", apiErr)
        }
    } else {
        log.Fatalf("Request error: %v", err)
    }
}
```

### Common Error Codes

| HTTP | Error Code | Description | Action |
|------|-----------|-------------|--------|
| 401 | - | Invalid/missing API key | Verify API key is correct |
| 404 | `ERROR_ZONE_NOT_FOUND` | Zone doesn't exist | Check zone in dashboard |
| 409 | `ERROR_ZONE_WORKFLOW_IN_PROGRESS` | Concurrent operation | Retry after delay |
| 429 | - | Rate limit exceeded | Automatic retry with backoff |
| 5xx | - | Server error | Automatic retry with backoff |

## Advanced Usage

### Custom RRSet Operations

```go
ops := []opusdns.RRSetOperation{
    {
        Op: "upsert",
        RRSet: opusdns.RRSet{
            Name: "_acme-challenge",
            Type: "TXT",
            TTL:  60,
            Records: []opusdns.Record{
                {RData: "\"challenge-value\""},
            },
        },
    },
}

err := client.PatchRRSets("example.com", ops)
if err != nil {
    log.Fatal(err)
}
```

### Batch Operations

```go
ops := []opusdns.RRSetOperation{
    {Op: "upsert", RRSet: opusdns.RRSet{Name: "record1", Type: "TXT", TTL: 60, Records: []opusdns.Record{{RData: "\"value1\""}}}},
    {Op: "upsert", RRSet: opusdns.RRSet{Name: "record2", Type: "TXT", TTL: 60, Records: []opusdns.Record{{RData: "\"value2\""}}}},
    {Op: "remove", RRSet: opusdns.RRSet{Name: "old-record", Type: "TXT"}},
}

err := client.PatchRRSets("example.com", ops)
```

## Testing

### Run Unit Tests

```bash
go test -v ./...
```

### Run Tests with Coverage

```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Docker Testing

```bash
# Build and test
docker-compose up opusdns-go-client

# Integration tests (requires API key)
export OPUSDNS_API_KEY="opk_..."
docker-compose up integration-test
```

### Integration Testing

Create a `.env` file:

```bash
OPUSDNS_API_KEY=opk_your_sandbox_key
OPUSDNS_API_ENDPOINT=https://sandbox.opusdns.com
```

Run integration tests:

```bash
docker-compose up integration-test
```

## ACME Integration Example

This library is designed for ACME DNS-01 challenge providers:

```go
func (p *Provider) Present(domain, token, keyAuth string) error {
    fqdn := fmt.Sprintf("_acme-challenge.%s", domain)
    value := computeKeyAuth(keyAuth)
    
    if err := p.client.UpsertTXTRecord(fqdn, value); err != nil {
        return fmt.Errorf("failed to create challenge record: %w", err)
    }
    
    return p.client.WaitForPropagation(fqdn, value)
}

func (p *Provider) CleanUp(domain, token, keyAuth string) error {
    fqdn := fmt.Sprintf("_acme-challenge.%s", domain)
    
    // Best-effort cleanup - don't fail ACME flow
    if err := p.client.RemoveTXTRecord(fqdn, "TXT"); err != nil {
        log.Printf("Warning: cleanup failed for %s: %v", fqdn, err)
    }
    
    return nil
}
```

## API Reference

See [GoDoc](https://pkg.go.dev/github.com/opusdns/opusdns-go-client) for complete API documentation.

### Key Types

- **`Config`**: Client configuration
- **`Client`**: API client with methods for DNS operations
- **`Zone`**: DNS zone metadata
- **`RRSet`**: Resource record set
- **`Record`**: Individual DNS record
- **`APIError`**: Detailed API error information

## Environment Support

| Environment | Base URL | Purpose |
|-------------|----------|---------|
| Production | `https://api.opusdns.com` | Live DNS operations |
| Sandbox | `https://sandbox.opusdns.com` | Testing (free) |

## Requirements

- Go 1.21 or later
- OpusDNS API key (production or sandbox)
- DNS zones created in OpusDNS account

## Dependencies

- `github.com/miekg/dns` - DNS propagation checking
- `github.com/stretchr/testify` - Testing utilities

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `go test -v ./...`
5. Submit a pull request

## Security

- **Never commit API keys** to version control
- **Rotate API keys regularly** via the OpusDNS dashboard
- **Use environment variables** for production deployments
- **Use sandbox** for development and testing

Report security vulnerabilities to: security@opusdns.com

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- **Documentation**: https://docs.opusdns.com
- **API Reference**: https://api.opusdns.com/docs
- **Issues**: https://github.com/opusdns/opusdns-go-client/issues
- **Email**: support@opusdns.com

## Related Projects

- [opusdns-lego-provider](https://github.com/opusdns/opusdns-lego-provider) - go-acme/lego DNS provider
- [libdns-opusdns](https://github.com/opusdns/libdns-opusdns) - libdns provider (for Caddy)
- [certbot-dns-opusdns](https://github.com/opusdns/certbot-dns-opusdns) - Certbot DNS plugin
- [opusdns-acme-sh-hook](https://github.com/opusdns/opusdns-acme-sh-hook) - acme.sh DNS API hook

---

**Made with ❤️ for the ACME community**
