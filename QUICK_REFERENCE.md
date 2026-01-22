# OpusDNS Go Client - Quick Reference

## Installation

```bash
go get github.com/opusdns/opusdns-go-client
```

## Basic Usage

```go
import "github.com/opusdns/opusdns-go-client"

// Create client
client := opusdns.NewClient(&opusdns.Config{
    APIKey: "opk_your_api_key_here",
})

// Add TXT record
client.UpsertTXTRecord("_acme-challenge.example.com", "value")

// Wait for propagation
client.WaitForPropagation("_acme-challenge.example.com", "value")

// Remove TXT record
client.RemoveTXTRecord("_acme-challenge.example.com", "TXT")
```

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `APIKey` | string | required | OpusDNS API key (opk_...) |
| `APIEndpoint` | string | `https://api.opusdns.com` | API endpoint URL |
| `TTL` | int | `60` | Record TTL in seconds |
| `HTTPTimeout` | duration | `30s` | HTTP request timeout |
| `MaxRetries` | int | `3` | Max retry attempts |
| `PollingInterval` | duration | `6s` | DNS check interval |
| `PollingTimeout` | duration | `60s` | Max propagation wait |
| `DNSResolvers` | []string | `["8.8.8.8:53", "1.1.1.1:53"]` | DNS servers |

## Core Methods

### Zone Management

```go
// List all zones
zones, err := client.ListZones()

// Find zone for FQDN
zone, err := client.FindZoneForFQDN("sub.example.com")
// Returns: "example.com"
```

### Record Operations

```go
// Upsert TXT record (auto-quotes value)
err := client.UpsertTXTRecord(fqdn, value)

// Remove record
err := client.RemoveTXTRecord(fqdn, "TXT")

// Advanced: Batch operations
ops := []opusdns.RRSetOperation{
    {Op: "upsert", RRSet: ...},
    {Op: "remove", RRSet: ...},
}
err := client.PatchRRSets("example.com", ops)
```

### DNS Propagation

```go
// Wait for record to propagate
err := client.WaitForPropagation(fqdn, expectedValue)
// Polls 8.8.8.8 and 1.1.1.1 every 6s for up to 60s
```

## Error Handling

```go
if err != nil {
    if apiErr, ok := err.(*opusdns.APIError); ok {
        switch apiErr.StatusCode {
        case 401: // Invalid API key
        case 404: // Zone not found
        case 409: // Operation in progress
        case 429: // Rate limited (auto-retried)
        }
    }
}
```

## Environment Variables

```bash
export OPUSDNS_API_KEY="opk_..."
export OPUSDNS_API_ENDPOINT="https://sandbox.opusdns.com"
```

## Testing

```bash
# Unit tests
go test -v ./...

# With coverage
go test -race -coverprofile=coverage.out ./...

# Docker
docker-compose up opusdns-go-client
```

## Common Patterns

### ACME DNS-01 Challenge

```go
func Present(domain, keyAuth string) error {
    fqdn := "_acme-challenge." + domain
    value := computeKeyAuth(keyAuth)
    
    if err := client.UpsertTXTRecord(fqdn, value); err != nil {
        return err
    }
    return client.WaitForPropagation(fqdn, value)
}

func CleanUp(domain string) error {
    fqdn := "_acme-challenge." + domain
    return client.RemoveTXTRecord(fqdn, "TXT")
}
```

### Multiple Records

```go
ops := []opusdns.RRSetOperation{
    {Op: "upsert", RRSet: opusdns.RRSet{
        Name: "_acme-challenge",
        Type: "TXT",
        TTL:  60,
        Records: []opusdns.Record{
            {RData: "\"value1\""},
            {RData: "\"value2\""},
        },
    }},
}
client.PatchRRSets("example.com", ops)
```

## API Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/v1/dns` | List zones |
| PATCH | `/v1/dns/{zone}/rrsets` | Update records |

## HTTP Headers

```
X-Api-Key: opk_...
Content-Type: application/json
```

## Zone Detection Algorithm

Input: `_acme-challenge.sub.example.com`

1. List all zones: `["example.com", "other.com"]`
2. Match FQDN suffix against zones
3. Return longest match: `example.com`

## Propagation Strategy

1. Query DNS servers: 8.8.8.8:53, 1.1.1.1:53
2. Check every 6 seconds (configurable)
3. Max 10 attempts = 60 seconds (configurable)
4. Return immediately when found

## Production Best Practices

1. **Use environment variables** for API keys
2. **Implement retry logic** for transient failures (built-in)
3. **Cache zones** to minimize API calls (built-in)
4. **Log cleanup errors** but don't fail
5. **Test in sandbox** before production

## Troubleshooting

### "no zone found for FQDN"
- Verify zone exists in OpusDNS account
- Check FQDN matches zone name

### "invalid API key" (401)
- Verify API key format (67 chars, starts with opk_)
- Check API key is active in dashboard

### "DNS propagation timeout"
- Increase `PollingTimeout`
- Verify DNS servers are reachable
- Check record was created successfully

### Rate limiting (429)
- Automatic retry with backoff
- Reduce API call frequency
- Use zone caching (automatic)

## Links

- **Docs**: https://docs.opusdns.com
- **API**: https://api.opusdns.com/docs
- **GitHub**: https://github.com/opusdns/opusdns-go-client
- **Issues**: https://github.com/opusdns/opusdns-go-client/issues
