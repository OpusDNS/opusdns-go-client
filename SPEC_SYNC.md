# Staying in sync with the OpusDNS API

This client is hand-written. Nothing here is generated, so a change to the API
does not reach the client on its own — somebody has to write the code. The point
of the machinery described below is that nobody has to *remember* to look: the
build tells you what changed and what is still missing.

## What is pinned

| File | What it holds |
|---|---|
| `spec/openapi.yaml` | The published public specification, byte-for-byte as `api-spec` serves it. |
| `spec/version.yaml` | Which revision that is: npm version, the spec's own `info.version`, and the `api-spec` commit. |
| `spec/coverage.yaml` | One entry per operation, saying whether the client implements it, has deferred it, or excludes it on purpose. |

The source of truth is
`https://raw.githubusercontent.com/OpusDNS/api-spec/main/src/openapi.yaml`.
Do not use the rendered document on `developers.opusdns.com`; it is a Scalar
build of the site and lags behind by days.

## Commands

```bash
make spec-outdated              # has api-spec moved on? changes nothing
make spec-update                # refresh spec/, then report what needs code
make spec-check                 # check the client against the vendored spec
make spec-stubs                 # print coverage.yaml entries for untriaged operations
make spec-sync SPEC_REF=<sha>   # vendor one specific api-spec commit
```

Everything here runs locally and needs no credentials. `make spec-outdated` only
reads, so it is safe on a dirty tree; it exits non-zero when upstream is ahead.
`make spec-update` is the usual entry point: it vendors the new spec and then
tells you what that costs in code.

`make spec-sync` is idempotent: run it twice against an unchanged upstream and
the working tree stays clean. The version stamp deliberately records no
timestamp of its own, only upstream identity.

## The two checks

Both live in `opusdns/spec_coverage_test.go` and run as part of `go test ./...`,
offline, against the vendored copy.

**`TestSpecCoverage`** compares three things: the operations in the spec, the
entries in `spec/coverage.yaml`, and the routes the code actually builds. The
routes are read out of the source with `go/ast`: each exported service method
calls `BuildPath` once and issues one request, which is enough to reconstruct
`METHOD /v1/...`. Failures read as:

| Label | Meaning |
|---|---|
| `UNMAPPED` | The spec has an operation nobody has triaged. Add an entry. |
| `STALE` | `coverage.yaml` names an operation the API no longer serves. |
| `MISMATCH` | The Go method builds a different route than the manifest claims. |
| `ORPHAN` | The client calls a route that is not in the spec at all. |
| `UNCLAIMED` | The client implements a route no manifest entry points at. |
| `MISSING` | A manifest entry names a method that builds no route. |

A method whose path cannot be read off a single `BuildPath` call opts out with a
`//speccheck:ignore` directive in its doc comment, written with no space after
the slashes so it stays out of the rendered documentation. It still has to
appear in the manifest, and the reflect check still applies; only the route
comparison is skipped. `Domains.RequestAuthCode` is the current example: the TLD
is a path segment there, so one method serves nine spec paths.

**`TestSpecModels`** compares the `json` tags of the hand-written response
structs with the properties of the schemas they mirror, for the mappings listed
under `models:` in the manifest. It is deliberately asymmetric:

- A struct field the spec no longer declares is an **error** (`DROPPED`). That
  field can only ever decode to the zero value now, which is the failure mode
  that silently produces wrong data.
- A property the struct does not carry is a **warning**, because the client has
  never mirrored every field. Move a type into `models.strict` once its fields
  are complete and the warning becomes an error too.

Every Go type named under `models.mappings` must also appear in `modelRegistry`
in the test file: Go cannot look a type up by name at run time.

Not checked: query parameters, headers, request bodies, enum values and status
codes. Read the OpenAPI changes report on the sync PR for those.

## Two things can drift, and they are caught differently

**The client against the vendored spec.** `TestSpecCoverage` and
`TestSpecModels` run in the ordinary test suite, so every pull request and every
push to `main` already catches this. Nothing extra is needed.

**The vendored spec against what is published.** No local test can see this: the
vendored copy is a snapshot, and it only moves when somebody runs
`make spec-sync`. `.github/workflows/check-openapi-spec.yaml` is what notices,
weekly. It fetches the published spec, re-runs both checks against it, writes
the whole report into the run summary, and fails.

That workflow is deliberately read-only. It uses no token beyond the built-in
one, opens no pull request and pushes nothing, so there is no secret to set up
or rotate. A failed scheduled run is the notification: it shows in the Actions
tab, and GitHub emails whoever last edited the cron in that file. Run it on
demand from the Actions tab with a `spec_ref` input, or reproduce it exactly
with `make spec-update`.

It also listens for a `repository_dispatch` of type `openapi_spec_update`, which
does nothing until somebody gives `api-spec` a token scoped to this repository.
Adding that would turn the weekly notice into a same-day one; the mechanism does
not depend on it.

## Acting on a drift report

The run summary carries the OpenAPI changes report, the coverage report and
ready-to-paste manifest stubs. Locally, `make spec-update` produces the same
thing.

1. Read the OpenAPI changes report first, for the changes that add no operation:
   new required fields, changed types, new enum values, new status codes.
2. Paste the stubs into `spec/coverage.yaml` and give each a real status:
   - `implemented` — write the service method and its httptest test, then name
     the method that calls `BuildPath` first under `method:`.
   - `deferred` — in scope, not done yet. Say why in `reason:`.
   - `excluded` — deliberately out of scope. Say why in `reason:`.
3. Fix every `STALE`, `MISMATCH`, `ORPHAN` and `DROPPED` line. Those are bugs:
   the client is calling something the API does not serve, or reading a field it
   no longer sends.
4. `make spec-check` until green, then `go test ./...`.
5. Commit `spec/` together with the code it needed. They belong in one change:
   the vendored spec is the evidence for what the code was written against.

Operations that are always `excluded`, so they need no fresh argument each time:

- server-sent event streams, which are not request/response operations;
- routes with no `security` block, which are unauthenticated token flows
  completed in a browser rather than by an API client;
- `ai-concierge` and `parking`, which are intentionally not part of this client.

Promoting a `deferred` entry later is just changing the status and adding
`method:` in the same PR that adds the code.

## Release checklist

- [ ] `make spec-outdated` is quiet and `go test ./...` is green on `main`.
- [ ] The release notes name the spec the release was built against, from
      `spec/version.yaml` (npm version and `info.version`).
- [ ] Removed struct fields or renamed methods are called out as breaking.
- [ ] Tag `vX.Y.Z`; the release workflow runs GoReleaser from the tag.
