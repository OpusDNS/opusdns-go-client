# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-01-15

### Added
- Initial release of OpusDNS Go client library
- Complete OpusDNS API integration:
  - Zone listing with pagination support
  - RRSet operations (upsert/remove)
  - Automatic zone detection from FQDN
- DNS propagation polling:
  - Query multiple public DNS servers (8.8.8.8, 1.1.1.1)
  - Configurable polling interval and timeout
  - Return immediately when record is verified
- Robust error handling:
  - Retry logic with exponential backoff for 429/5xx errors
  - Clear error messages for all API error codes
  - APIError type with status codes and details
- Zone caching:
  - 5-minute cache TTL to minimize API calls
  - Thread-safe cache implementation
- Configuration options:
  - Customizable API endpoint (production/sandbox)
  - Configurable TTL, timeouts, retry attempts
  - Custom DNS resolvers for propagation checks
- Comprehensive test suite:
  - Unit tests with 95%+ coverage
  - Mocked HTTP responses for reliable testing
  - Table-driven test patterns
- Documentation:
  - Complete README with usage examples
  - Godoc for all exported types and functions
  - Example application demonstrating ACME workflow
- Docker support:
  - Dockerfile for testing in containerized environment
  - docker-compose.yml for local development
  - Integration test configuration
- CI/CD:
  - GitHub Actions workflow for testing and linting
  - Multi-version Go testing (1.21, 1.22)
  - Code coverage reporting
- Development tools:
  - .gitignore for Go projects
  - .dockerignore for optimized builds
  - Contributing guidelines
  - MIT License

### API Methods
- `NewClient(config *Config) *Client` - Create new client
- `ListZones() ([]Zone, error)` - List all DNS zones
- `FindZoneForFQDN(fqdn string) (string, error)` - Find zone for FQDN
- `PatchRRSets(zone string, ops []RRSetOperation) error` - Apply RRSet operations
- `UpsertTXTRecord(fqdn, value string) error` - Create/update TXT record
- `RemoveTXTRecord(fqdn, recordType string) error` - Delete TXT record
- `WaitForPropagation(fqdn, expectedValue string) error` - Verify DNS propagation

### Configuration
- Default API endpoint: `https://api.opusdns.com`
- Default TTL: 60 seconds
- Default HTTP timeout: 30 seconds
- Default max retries: 3
- Default polling interval: 6 seconds
- Default polling timeout: 60 seconds
- Default DNS resolvers: 8.8.8.8:53, 1.1.1.1:53

### Dependencies
- `github.com/miekg/dns` v1.1.58 - DNS operations
- `github.com/stretchr/testify` v1.8.4 - Testing utilities

## [0.1.0] - 2025-01-10

### Added
- Initial development version (not released)

[Unreleased]: https://github.com/opusdns/opusdns-go-client/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/opusdns/opusdns-go-client/releases/tag/v1.0.0
[0.1.0]: https://github.com/opusdns/opusdns-go-client/releases/tag/v0.1.0
