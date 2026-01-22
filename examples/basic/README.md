# Example: Basic Usage

This example demonstrates basic usage of the OpusDNS Go client library.

## Prerequisites

- OpusDNS API key (sandbox or production)
- At least one DNS zone in your OpusDNS account

## Setup

```bash
# Set environment variables
export OPUSDNS_API_KEY="opk_your_api_key_here"
export OPUSDNS_API_ENDPOINT="https://sandbox.opusdns.com"  # or https://api.opusdns.com
```

## Run

```bash
cd examples/basic
go run main.go
```

## What This Example Does

1. **Lists all DNS zones** in your OpusDNS account
2. **Performs ACME DNS-01 challenge workflow**:
   - Creates a `_acme-challenge` TXT record
   - Waits for DNS propagation on public DNS servers
   - Cleans up the TXT record
3. **Demonstrates zone detection** for various FQDNs
4. **Shows error handling** for non-existent zones

## Expected Output

```
=== Listing DNS Zones ===
Found 2 zones:
  - example.com (DNSSEC: disabled, Created: 2025-01-15T12:00:00Z)
  - test.com (DNSSEC: enabled, Created: 2025-01-10T08:30:00Z)

=== ACME DNS-01 Challenge Workflow ===
Domain: _acme-challenge.example.com
Challenge Value: example-challenge-1705324800

[1/3] Creating TXT record...
✓ TXT record created successfully

[2/3] Waiting for DNS propagation...
Checking DNS servers: 8.8.8.8, 1.1.1.1
✓ DNS record propagated successfully (took 12s)

[3/3] Cleaning up TXT record...
✓ TXT record removed successfully

=== Zone Detection Example ===
  _acme-challenge.example.com -> example.com
  _acme-challenge.subdomain.example.com -> example.com
  _acme-challenge.deep.subdomain.example.com -> example.com

=== Error Handling Example ===
Attempting to add record to non-existent zone: _acme-challenge.nonexistent-zone-12345.com
✓ Caught error (expected): no zone found for FQDN _acme-challenge.nonexistent-zone-12345.com. (available zones: 2)

=== Example completed successfully ===
```
