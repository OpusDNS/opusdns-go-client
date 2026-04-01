# Copilot Instructions

## Build, test, and lint

Use the same commands the repository uses in CI:

```bash
go test ./...
go test ./opusdns -run TestDNSService_ListZones
go build ./...
golangci-lint run
go mod tidy && git diff --exit-code go.mod go.sum
```

CI and the README are standardized on **Go 1.25** (`.github/workflows/ci.yml`, `README.md`), even though `go.mod` still declares `go 1.21`. Prefer matching CI when validating changes.

## High-level architecture

This repository has three main layers:

1. `models/` contains the API data model: request/response structs, enums, pagination metadata, and pointer helpers such as `models.BoolPtr` and `models.StringPtr`.
2. `opusdns/` is the reusable client library. `Client` wires together service structs like `DNSService`, `DomainsService`, `ContactsService`, and `AvailabilityService` around one shared `HTTPClient`.
3. `cmd/opusdns/` is a Cobra CLI that is only a thin wrapper over the library. It builds one `opusdns.Client` in `PersistentPreRunE`, derives a timeout-backed context per command, and delegates to the same service methods the library exposes.

The shared `HTTPClient` in `opusdns/http.go` is where cross-cutting behavior lives: request construction, API version path prefixing, JSON encoding/decoding, timestamp normalization, retries with exponential backoff, and 429 rate-limit handling. Service files should stay thin and mostly translate typed options into query parameters plus request/response decoding.

For list endpoints, the public `List...` methods usually do **automatic pagination** by repeatedly calling the corresponding `List...Page` method until `Pagination.HasNextPage` is false. Keep both forms aligned when changing list behavior.

## Key conventions

- **Preserve the thin-service pattern.** New service methods should build paths with `client.http.BuildPath(...)`, call the appropriate shared HTTP helper (`Get`, `Post`, `Patch`, etc.), and decode into `models` types. Avoid duplicating transport, retry, or auth logic inside services.
- **Normalize resource references before building paths where existing code does so.** DNS methods trim trailing dots from zone names before `url.PathEscape`; domain and other resource references are usually passed through `url.PathEscape` directly.
- **Use typed `models` enums and helpers instead of raw strings/bools** when request types already provide them (`models.RRSetTypeA`, `models.SortDesc`, `models.BoolPtr`, etc.).
- **Keep automatic pagination behavior intact.** When editing `List...` methods, maintain the default `DefaultPageSize` fallback and accumulation loop, not just the single-page request.
- **Errors are part of the public API surface.** HTTP status mapping is centralized in `opusdns/errors.go`; callers are expected to use `errors.Is` against sentinels like `opusdns.ErrNotFound` or helpers like `opusdns.IsRetryableError`. Preserve that behavior rather than introducing ad hoc error handling.
- **Tests use `httptest` at the client boundary.** The existing suite in `opusdns/client_test.go` validates HTTP method, path, headers, payload shape, pagination, retries, and error classification using `testify/assert` and `require`. Follow that pattern for new client behavior instead of mocking internals.
- **The CLI is a consumer of the library, not a second implementation.** If behavior changes in `opusdns/`, keep the Cobra commands in `cmd/opusdns/cmd/` delegating to the library rather than re-encoding API rules in command handlers.
