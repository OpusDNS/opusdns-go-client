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
| `spec/go.mod` | Not a real module. Its presence keeps this directory out of the published module zip, so the specification is not downloaded by everyone who runs `go get`. Deleting it silently puts 130 KB back into every download. |

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

The npm version in `spec/version.yaml` is the `@opusdns/api` release that ships
the vendored spec, also for an older `SPEC_REF`. `api-spec` commits a new spec
first and bumps `package.json` in a second commit, so the script takes the
first version bump at or after the last spec change, not `package.json` at the
ref itself. It fails when that spec has no npm release yet. These lookups go
to the GitHub API, which allows 60 unauthenticated requests an hour; set
`GITHUB_TOKEN` if you hit the limit.

## The checks

They live in `opusdns/spec_coverage_test.go` (with the shape comparison in
`opusdns/spec_shape_test.go`) and run as part of `go test ./...`, offline,
against the vendored copy.

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
| `UNCLAIMED` | The client implements a route no manifest entry points at, or a `//speccheck:ignore` method no implemented entry names. |
| `MISSING` | A manifest entry names a method that builds no route. |

A method whose path cannot be read off a single `BuildPath` call opts out with a
`//speccheck:ignore` directive in its doc comment, written with no space after
the slashes so it stays out of the rendered documentation. It still has to
appear in the manifest (`UNCLAIMED` otherwise), and the reflect check and
`TestSpecOperations` still apply; only the route comparison is skipped.
`Domains.RequestAuthCode` is the current example: the TLD is a path segment
there, so one method serves nine spec paths.

**`TestSpecOperations`** reads every method that an implemented manifest entry
names, with `go/types`, and compares what it does with its operation:

- the request body it passes to `Post`/`Put`/`Patch`/`Do`, against the
  operation's JSON request schema;
- the query parameters it sets (literal keys of `url.Values.Set`/`Add`, in the
  method and in the package functions it calls), against the operation's query
  parameters;
- the type it hands to `DecodeResponse`, against the JSON body of the
  operation's 2xx responses (a union when there are several, as for renew).

Types are compared recursively: nested structs, slices, maps and pointers,
following `$ref`, `allOf`, `anyOf` and `oneOf`. A union is checked as a union:
a response struct may carry the fields of every alternative; a request struct
has to match one alternative on its own, including its discriminator.

| Label | Meaning |
|---|---|
| `KIND` | The Go type encodes a different kind of JSON value than the schema allows: a number for an object, an object for an array, ... |
| `DROPPED` | A response field the schema does not declare. It always decodes to the zero value. |
| `UNKNOWN` | A request field the schema does not declare. The API ignores or rejects it. |
| `MISSING` | A required request property the struct does not carry, or a property a `models.strict` type lacks. |
| `OMITTED` | A required bool behind `omitempty`, so `false` is never sent. |
| `ZERO` | An optional field without `omitempty` whose zero value the schema rejects (`""` for `minLength: 6`), so leaving it unset fails. |
| `NOBODY` | The operation requires a request body and the method sends none. |
| `BODY` | The method sends a body the operation does not take. |
| `QUERY` | The method sends a query parameter the operation does not declare, or never sends a required one. |
| `RESPONSE` | The method decodes a body the operation does not return. |

**`TestSpecModels`** checks the types listed under `models:` in the manifest:

- `models.mappings` names the schema a response type mirrors. Most types are
  reached through `TestSpecOperations` anyway; a mapping makes the schema the
  type's canonical one. A type shared between operations may get only a subset
  of that schema from some of them (organization lists carry no balance); a
  field the subset lacks is then a warning, not `DROPPED`.
- `models.requests` does the same for request types, in the encoding
  direction.
- A property a response struct does not carry is a **warning**, because the
  client has never mirrored every field. Move a type into `models.strict` once
  its fields are complete and the warning becomes an error, against its
  canonical schema.
- A response field documented `Deprecated:` that the spec no longer declares is
  a warning: that is how a field is retired without breaking callers.

Types are looked up by name in package `models` with `go/types`, so a mapping
needs no registration in the test file.

`TestShapeCheckerCatchesDrift` keeps the comparison honest: it feeds it small
reproductions of the bugs these checks were written after.

Not checked: headers, enum values, status codes, path parameter types and query
keys that are not string literals. Read the OpenAPI changes report on the sync
PR for those.

## Two things can drift, and they are caught differently

**The client against the vendored spec.** `TestSpecCoverage`,
`TestSpecOperations` and `TestSpecModels` run in the ordinary test suite, so
every pull request and every push to `main` already catches this. Nothing extra
is needed.

**The vendored spec against what is published.** Nothing catches this on its
own. The vendored copy is a snapshot: it only moves when somebody runs
`make spec-sync`, and until then CI is happily checking the client against a
specification that may be months old. This is the one part of the loop that is
deliberately manual: it needs the network, so it does not belong in the test
suite, and it was judged not worth a scheduled job of its own.

So make it a habit rather than a memory. Run it:

- before cutting a release, which the checklist below repeats;
- when you are about to add or change a service method, so you write against
  what the API serves today;
- when a changelog entry or a colleague mentions a new endpoint.

```bash
make spec-outdated
```

It reads only and exits non-zero when upstream is ahead, so it is also safe to
put behind a shell alias or a local git hook if you would rather not think about
it.

If remembering ever becomes the weak link, a scheduled GitHub Actions job that
runs `make spec-update` and fails would cover it, read-only and with no
credentials beyond the built-in token. Its absence is a choice, not an
oversight.

## Acting on a drift report

`make spec-update` vendors the new specification and then prints the coverage
report and ready-to-paste manifest stubs. For the changes that add no operation,
diff the specification itself: `git diff spec/openapi.yaml`, or run
[`openapi-changes`](https://pb33f.io/openapi-changes/) over the old and new
copies for a readable summary.

1. Read the OpenAPI changes report first, for the changes that add no operation:
   new required fields, changed types, new enum values, new status codes.
2. Paste the stubs into `spec/coverage.yaml` and give each a real status:
   - `implemented` — write the service method and its httptest test, then name
     the method that calls `BuildPath` first under `method:`.
   - `deferred` — in scope, not done yet. Say why in `reason:`.
   - `excluded` — deliberately out of scope. Say why in `reason:`.
3. Fix every error line (`STALE`, `MISMATCH`, `ORPHAN`, `DROPPED`, `KIND`,
   `UNKNOWN`, `QUERY`, ...). Those are bugs: the client is calling something the
   API does not serve, reading a field it no longer sends, or sending something
   it does not take.
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
- [ ] Tag `vX.Y.Z`; the release workflow checks that `go.mod` is tidy, runs
      `go vet` and `go test`, and only then runs GoReleaser from the tag.
