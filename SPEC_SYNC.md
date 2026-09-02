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
make spec-sync                  # refresh spec/ from api-spec main
make spec-sync SPEC_REF=<sha>   # ... or from one api-spec commit
make spec-check                 # verify the client against the vendored spec
make spec-stubs                 # print coverage.yaml entries for untriaged operations
```

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

## How a change reaches this repo

```
opusdns-api (push to main)
  └─ repository_dispatch → api-spec: regenerate types, bump npm version
       └─ repository_dispatch → this repo: .github/workflows/sync-openapi-spec.yaml
            └─ bot PR on branch bot/openapi-spec-sync
```

A weekly cron runs the same workflow, so a lost dispatch self-heals within seven
days. `workflow_dispatch` takes a `spec_ref` input for a manual run.

## Triaging a sync PR

The PR body carries the OpenAPI changes report, the coverage report, and
ready-to-paste manifest stubs. **CI on that PR is expected to be red** while new
operations are untriaged; that failure is the to-do list, and it is what stops an
unnoticed endpoint from sitting unimplemented for months.

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

Operations that are always `excluded`, so they need no fresh argument each time:

- server-sent event streams, which are not request/response operations;
- routes with no `security` block, which are unauthenticated token flows
  completed in a browser rather than by an API client;
- `ai-concierge` and `parking`, which are intentionally not part of this client.

Promoting a `deferred` entry later is just changing the status and adding
`method:` in the same PR that adds the code.

## Release checklist

- [ ] The sync PR is merged and `go test ./...` is green on `main`.
- [ ] The release notes name the spec the release was built against, from
      `spec/version.yaml` (npm version and `info.version`).
- [ ] Removed struct fields or renamed methods are called out as breaking.
- [ ] Tag `vX.Y.Z`; the release workflow runs GoReleaser from the tag.
