# Contributing to OpusDNS Go Client

Thank you for your interest in contributing to the OpusDNS Go Client library! This document provides guidelines and instructions for contributing.

## Code of Conduct

Be respectful, inclusive, and professional in all interactions.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/opusdns-go-client.git
   cd opusdns-go-client
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/opusdns/opusdns-go-client.git
   ```

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

Use descriptive branch names:
- `feature/add-dnssec-support`
- `bugfix/fix-zone-caching`
- `docs/improve-readme`

### 2. Make Your Changes

- Write clean, idiomatic Go code
- Follow existing code style and conventions
- Add comments for complex logic
- Keep functions focused and testable

### 3. Write Tests

All new code must include tests:

```bash
# Run tests
go test -v ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Testing requirements:**
- Unit tests for all new functions
- Table-driven tests for multiple scenarios
- Mock HTTP responses using `httptest`
- Aim for >80% code coverage

### 4. Run Linters

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

### 5. Update Documentation

- Update README.md if adding new features
- Add godoc comments for exported types/functions
- Update examples if API changes
- Add entries to CHANGELOG.md

### 6. Commit Your Changes

Use clear, descriptive commit messages:

```bash
git add .
git commit -m "Add support for DNSSEC operations

- Implement GetDNSSECStatus method
- Add DNSSEC validation tests
- Update README with DNSSEC examples"
```

**Commit message format:**
- First line: Brief summary (50 chars or less)
- Blank line
- Detailed description (optional, wrap at 72 chars)

### 7. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub with:
- Clear title describing the change
- Description of what changed and why
- Link to any related issues
- Screenshots/examples if applicable

## Code Style Guidelines

### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `go vet` to catch common mistakes
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create TXT record for %s: %w", fqdn, err)
}

// Avoid: Returning bare errors
if err != nil {
    return err
}
```

### Testing

```go
// Good: Table-driven tests
func TestUpsertTXTRecord(t *testing.T) {
    tests := []struct {
        name    string
        fqdn    string
        value   string
        wantErr bool
    }{
        {name: "valid record", fqdn: "test.example.com", value: "value", wantErr: false},
        {name: "empty fqdn", fqdn: "", value: "value", wantErr: true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

### Documentation

```go
// Good: Clear godoc with examples
// UpsertTXTRecord creates or updates a TXT record for the given FQDN.
// It automatically detects the appropriate zone and quotes the value.
//
// Example:
//   err := client.UpsertTXTRecord("_acme-challenge.example.com", "challenge-value")
func (c *Client) UpsertTXTRecord(fqdn, value string) error {
    // Implementation
}
```

## Testing with Docker

Test your changes in a clean environment:

```bash
# Build and run tests
docker-compose up opusdns-go-client

# Run integration tests (requires API key)
export OPUSDNS_API_KEY="opk_your_sandbox_key"
docker-compose up integration-test
```

## Integration Testing

For testing against the OpusDNS API:

1. **Use the sandbox environment**: `https://sandbox.opusdns.com`
2. **Create a test API key** (don't use production keys)
3. **Set environment variables**:
   ```bash
   export OPUSDNS_API_KEY="opk_sandbox_key"
   export OPUSDNS_API_ENDPOINT="https://sandbox.opusdns.com"
   ```
4. **Run example**:
   ```bash
   cd examples/basic
   go run main.go
   ```

## Pull Request Checklist

Before submitting your pull request, ensure:

- [ ] Code follows Go style guidelines
- [ ] All tests pass: `go test -v ./...`
- [ ] Linter passes: `golangci-lint run`
- [ ] Code coverage is maintained or improved
- [ ] Documentation is updated (README, godoc)
- [ ] CHANGELOG.md is updated (for notable changes)
- [ ] Commit messages are clear and descriptive
- [ ] Examples run successfully (if applicable)
- [ ] No sensitive data (API keys, credentials) in code

## Review Process

1. **Automated checks** run on all PRs (tests, linters, build)
2. **Code review** by maintainers
3. **Feedback** may be provided for improvements
4. **Approval** and merge by maintainers

Please be patient - reviews may take a few days.

## Reporting Bugs

Found a bug? Please open an issue with:

- Clear, descriptive title
- Steps to reproduce
- Expected vs actual behavior
- Environment details (Go version, OS)
- Code samples (if applicable)
- Error messages or logs

**Template:**
```markdown
### Description
Brief description of the bug

### Steps to Reproduce
1. Create client with...
2. Call method...
3. Observe error...

### Expected Behavior
What should happen

### Actual Behavior
What actually happens

### Environment
- Go version: 1.21
- OS: macOS 14.0
- Library version: v1.0.0

### Additional Context
Error logs, stack traces, etc.
```

## Feature Requests

Have an idea? Open an issue with:

- Clear description of the feature
- Use case / motivation
- Proposed implementation (if any)
- Alternatives considered

## Security Issues

**Do not** open public issues for security vulnerabilities.

Email security reports to: security@opusdns.com

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

- **Documentation**: https://docs.opusdns.com
- **Issues**: https://github.com/opusdns/opusdns-go-client/issues
- **Email**: support@opusdns.com

---

Thank you for contributing! 🎉
