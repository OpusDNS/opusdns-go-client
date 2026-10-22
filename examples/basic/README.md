# Example: Basic Usage

This example demonstrates basic usage of the OpusDNS Go client library.

## Prerequisites

- OpusDNS API key (sandbox or production)
- Optional: a disposable zone name for the zone management demo

## Setup

```bash
# Set environment variables
export OPUSDNS_API_KEY="opk_your_api_key_here"
export OPUSDNS_API_ENDPOINT="https://sandbox.opusdns.com"  # or https://api.opusdns.com

# Optional: log every request and response
export OPUSDNS_DEBUG=true

# Optional: run the zone management demo against this zone.
# The demo creates the zone and deletes it again at the end,
# so use a name that does not exist in your account yet.
export DEMO_ZONE="my-disposable-zone.example"
```

## Run

```bash
cd examples/basic
go run main.go
```

## What This Example Does

1. **Lists DNS zones**: all zones with automatic pagination, then the first
   page of five zones sorted by creation date to show manual paging.
2. **Manages a zone** (only when `DEMO_ZONE` is set):
   - creates the zone with a `www` A record,
   - adds a `_test` TXT record with `UpsertRecord`,
   - upserts a `mail` A record and an MX record in one `PatchRecords` call,
   - fetches the zone and prints its RRsets,
   - deletes the zone again.
3. **Checks domain availability** for `example.com`, `opusdns-test-12345.com`
   and `google.com`.
4. **Shows error handling**: looks up a zone that does not exist and inspects
   the error with `IsNotFoundError`, `IsAPIError`, `IsUnauthorizedError` and
   `IsRetryableError`.

## Expected Output

The zones, availability results and timings depend on your account and on the
registries; the output below only shows the shape.

```
=== Listing DNS Zones ===
Found 2 zones:
  - example.com (DNSSEC: disabled, Created: 2025-01-15T12:00:00Z)
  - test.com (DNSSEC: enabled, Created: 2025-01-10T08:30:00Z)

Paginated listing (first page only):
  Page 1 of 1 (total: 2 zones)

=== Domain Availability Check ===
Checking availability for: [example.com opusdns-test-12345.com google.com]
Results (processed in 120ms):
  example.com: ❌ unavailable (unavailable)
  opusdns-test-12345.com: ✓ available (available)
  google.com: ❌ unavailable (unavailable)

=== Error Handling Example ===
✓ Correctly caught NotFound error

Error type checking examples:
  IsNotFoundError: true
  IsUnauthorizedError: false
  IsRetryableError: false

=== Examples completed successfully ===
```

With `DEMO_ZONE` set, a `=== Zone Management Demo ===` section between the zone
listing and the availability check reports each step with a `✓` line.
