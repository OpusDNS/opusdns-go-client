# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go test ./...                                   # unit tests
go test ./opusdns -run TestDNSService_ListZones # single test
go build ./...
golangci-lint run                               # lint (config in .golangci.yml)
go mod tidy && git diff --exit-code go.mod go.sum  # CI fails if go.mod is not tidy
govulncheck ./...                               # security scan (CI also runs this)
```

Integration tests hit the real API and are behind the `integration` build tag, so `go test ./...` skips them. Run via:

```bash
OPUSDNS_API_KEY="opk_..." ./scripts/integration-test.sh
# set OPUSDNS_INTEGRATION_ZONE=<disposable-zone> to also exercise the DNS write lifecycle
```

The module targets **Go 1.21+**. CI runs the test and build matrices on Go 1.21, 1.23, and 1.26; lint and security jobs use `stable`. Keep changes compatible with 1.21 (no newer-stdlib-only APIs).

## Architecture

Three layers:

1. **`models/`** — API data model: request/response structs, enums, pagination metadata, and pointer helpers (`models.BoolPtr`, `models.StringPtr`, etc.).
2. **`opusdns/`** — the reusable client library. `Client` (`client.go`) wires together per-domain service structs (`DNS`, `Domains`, `Contacts`, `EmailForwards`, `DomainForwards`, `TLDs`, `Availability`, `Organizations`, `Users`, `Auth`, `VanityNameservers`, `Hosts`, `Events`, `Jobs`, `Reports`, `Tags`) around one shared `HTTPClient`.
3. **`cmd/opusdns/`** — a Cobra CLI that is a thin wrapper over the library. It builds one `opusdns.Client` in `PersistentPreRunE`, derives a timeout context per command, and delegates to the same service methods. It is a *consumer* of the library, not a second implementation — don't re-encode API rules in command handlers.

`opusdns/http.go` is where all cross-cutting transport lives: request construction, API version path prefixing, JSON encode/decode, timestamp normalization, retries with exponential backoff, and 429 rate-limit handling. Service files stay thin: translate typed options into query params, build paths, decode responses.

## Conventions

- **Thin-service pattern.** New service methods build paths with `client.http.BuildPath(...)`, call the shared HTTP helper (`Get`, `Post`, `Patch`, ...), and decode into `models` types. Never duplicate transport/retry/auth logic in a service.
- **Automatic pagination.** Public `List...` methods loop over `List...Page` until `Pagination.HasNextPage` is false, with a `DefaultPageSize` fallback. Keep both forms aligned and preserve the accumulation loop when editing list behavior.
- **Path normalization.** Mirror existing code: DNS methods trim trailing dots from zone names before `url.PathEscape`; other resource refs pass through `url.PathEscape` directly.
- **Typed enums/helpers over raw strings/bools** when the request type provides them (`models.RRSetTypeA`, `models.SortDesc`, `models.BoolPtr`).
- **Errors are public API.** Status mapping is centralized in `opusdns/errors.go`. Callers use `errors.Is` against sentinels (`opusdns.ErrNotFound`) and helpers (`opusdns.IsRetryableError`). Don't introduce ad hoc error handling.
- **Tests use `httptest` at the client boundary** (`opusdns/client_test.go` and `service_*_test.go`), validating method, path, headers, payload, pagination, retries, and error classification with `testify`. Follow this pattern rather than mocking internals.
