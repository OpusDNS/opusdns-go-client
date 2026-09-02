# Development entry points. CI runs the same commands; see .github/workflows/.

# api-spec ref to vendor the OpenAPI specification from. Override to pin a
# review to one upstream commit: make spec-sync SPEC_REF=<sha>
SPEC_REF ?= main
# Alternative spec source, for debugging only. Skips commit pinning.
SPEC_URL ?=

SPEC_TESTS := 'TestSpecCoverage|TestSpecModels'

.PHONY: help build test lint tidy vuln spec-sync spec-check spec-stubs

help: ## Show this help
	@grep -hE '^[a-z-]+:.*?## ' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

build: ## Compile every package
	go build ./...

test: ## Run the unit tests (integration tests are behind the `integration` tag)
	go test ./...

lint: ## Run golangci-lint
	golangci-lint run

tidy: ## Tidy go.mod/go.sum
	go mod tidy

vuln: ## Run the vulnerability scanner
	govulncheck ./...

spec-sync: ## Refresh the vendored OpenAPI spec from api-spec
	go run scripts/specsync.go -ref "$(SPEC_REF)" $(if $(SPEC_URL),-spec-url "$(SPEC_URL)")

spec-check: ## Verify the client against the vendored spec (coverage + model drift)
	go test ./opusdns -run $(SPEC_TESTS) -v

spec-stubs: ## Print coverage.yaml stubs for operations that are not yet triaged
	@go test ./opusdns -run TestSpecCoverage -v 2>&1 | \
		sed -n '/^--- STUBS/,/^--- END STUBS/p'
