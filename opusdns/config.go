// Package client provides a Go client library for the OpusDNS API.
package opusdns

import (
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"sync"
	"time"
)

const (
	// DefaultAPIEndpoint is the production OpusDNS API endpoint.
	DefaultAPIEndpoint = "https://api.opusdns.com"

	// DefaultAPIVersion is the API version to use.
	DefaultAPIVersion = "v1"

	// DefaultTTL is the default TTL for DNS records (60 seconds).
	DefaultTTL = 60

	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default number of retries for transient failures.
	DefaultMaxRetries = 3

	// DefaultRetryWaitMin is the minimum wait time between retries.
	DefaultRetryWaitMin = 1 * time.Second

	// DefaultRetryWaitMax is the maximum wait time between retries.
	DefaultRetryWaitMax = 30 * time.Second

	// DefaultPageSize is the default page size for paginated requests.
	DefaultPageSize = 100

	// MaxPageSize is the maximum allowed page size.
	MaxPageSize = 1000
)

// Version information - can be set via ldflags at build time
var (
	// Version is the client library version.
	Version = "dev"
)

const (
	// ProductName is the client name the API knows this library by. It is the first half
	// of both the User-Agent and the X-OpusDNS-Client product token.
	ProductName = "opusdns-go-client"

	// unknownVersion is what both fall back to when no usable version can be determined.
	unknownVersion = "dev"

	// modulePath is this module, looked up in the build info of whatever binary embeds it.
	modulePath = "github.com/opusdns/opusdns-go-client"
)

// versionPattern is the charset the API accepts for the version half of the client product
// token. A value outside it is dropped on their side rather than stored, so anything that
// does not match is replaced with unknownVersion here instead of being sent as noise.
var versionPattern = regexp.MustCompile(`^[A-Za-z0-9._+-]{1,64}$`)

var (
	resolvedVersionOnce sync.Once
	resolvedVersion     string
)

// ResolvedVersion returns the version this build reports to the API.
//
// Releases set Version through ldflags, which covers the CLI. That does nothing for a program
// that imports this package, so when Version is still the default the module version recorded
// in the importing binary is used instead - otherwise every library consumer would report
// "dev" and the version dimension would be empty for exactly the callers it is most useful for.
func ResolvedVersion() string {
	resolvedVersionOnce.Do(func() {
		resolvedVersion = resolveVersion(Version, debug.ReadBuildInfo)
	})
	return resolvedVersion
}

// resolveVersion takes its inputs as arguments so it can be tested without a real build.
func resolveVersion(ldflagsVersion string, readBuildInfo func() (*debug.BuildInfo, bool)) string {
	if ldflagsVersion != unknownVersion && versionPattern.MatchString(ldflagsVersion) {
		return ldflagsVersion
	}

	info, ok := readBuildInfo()
	if !ok || info == nil {
		return unknownVersion
	}

	// In a consumer's binary this module is a dependency; in our own CLI built without
	// ldflags it is the main module.
	candidate := ""
	for _, dep := range info.Deps {
		if dep != nil && dep.Path == modulePath {
			candidate = dep.Version
			if dep.Replace != nil {
				candidate = dep.Replace.Version
			}
			break
		}
	}
	if candidate == "" && info.Main.Path == modulePath {
		candidate = info.Main.Version
	}

	// An unversioned build reports "(devel)", whose parentheses are outside the accepted
	// charset. Reporting "dev" says the same thing in a value the API will keep.
	if !versionPattern.MatchString(candidate) {
		return unknownVersion
	}
	return candidate
}

// GetUserAgent returns the default user agent string.
func GetUserAgent() string {
	return ProductName + "/" + ResolvedVersion()
}

// GetClientToken returns the default value for the X-OpusDNS-Client header: an RFC 9110
// product token naming this library and its version.
//
// The API uses it to attribute a request to an origin channel for its own analytics. It is
// informational, never an authentication or authorization signal, and carries nothing about
// the caller beyond which client library made the call.
func GetClientToken() string {
	return ProductName + "/" + ResolvedVersion()
}

// Environment variable names for configuration.
const (
	EnvAPIKey      = "OPUSDNS_API_KEY"
	EnvAPIEndpoint = "OPUSDNS_API_ENDPOINT"
	EnvAPIVersion  = "OPUSDNS_API_VERSION"
	EnvDebug       = "OPUSDNS_DEBUG"
)

// Config holds the configuration for the OpusDNS client.
type Config struct {
	// APIKey is the OpusDNS API key (format: opk_...).
	// This is required for authentication.
	// Can also be set via OPUSDNS_API_KEY environment variable.
	APIKey string

	// APIEndpoint is the base URL for the OpusDNS API.
	// Default: https://api.opusdns.com
	// Can also be set via OPUSDNS_API_ENDPOINT environment variable.
	APIEndpoint string

	// APIVersion is the API version to use.
	// Default: v1
	// Can also be set via OPUSDNS_API_VERSION environment variable.
	APIVersion string

	// TTL is the default TTL for DNS records in seconds.
	// Default: 60
	TTL int

	// HTTPTimeout is the timeout for HTTP requests.
	// Default: 30s
	HTTPTimeout time.Duration

	// MaxRetries is the maximum number of retries for transient failures (429, 5xx).
	// Set to 0 to disable retries.
	// Default: 3
	MaxRetries int

	// RetryWaitMin is the minimum wait time between retries.
	// Default: 1s
	RetryWaitMin time.Duration

	// RetryWaitMax is the maximum wait time between retries.
	// Default: 30s
	RetryWaitMax time.Duration

	// HTTPClient allows providing a custom HTTP client.
	// If nil, a default client with the configured timeout will be used.
	// Use this to configure custom transport settings, proxies, etc.
	HTTPClient *http.Client

	// UserAgent is the user agent string to use for API requests.
	// Default: opusdns-go-client/<version>
	UserAgent string

	// ClientToken is sent as the X-OpusDNS-Client header, which the API uses to attribute
	// a request to an origin channel for its own analytics. It is informational only.
	// Set it to the empty string to omit the header.
	// Default: opusdns-go-client/<version>
	ClientToken string

	// Debug enables debug logging of HTTP requests and responses.
	// Can also be enabled via OPUSDNS_DEBUG=true environment variable.
	Debug bool

	// Logger is the logger to use for debug output.
	// If nil, logs will be written to stdout.
	Logger Logger
}

// Logger is the interface for logging debug messages.
type Logger interface {
	Printf(format string, v ...interface{})
}

// Option is a functional option for configuring the client.
type Option func(*Config)

// WithAPIKey sets the API key for authentication.
func WithAPIKey(apiKey string) Option {
	return func(c *Config) {
		c.APIKey = apiKey
	}
}

// WithAPIEndpoint sets a custom API endpoint.
func WithAPIEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.APIEndpoint = endpoint
	}
}

// WithAPIVersion sets the API version.
func WithAPIVersion(version string) Option {
	return func(c *Config) {
		c.APIVersion = version
	}
}

// WithTTL sets the default TTL for DNS records.
func WithTTL(ttl int) Option {
	return func(c *Config) {
		c.TTL = ttl
	}
}

// WithHTTPTimeout sets the HTTP request timeout.
func WithHTTPTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.HTTPTimeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retries.
func WithMaxRetries(retries int) Option {
	return func(c *Config) {
		c.MaxRetries = retries
	}
}

// WithRetryWait sets the minimum and maximum retry wait times.
func WithRetryWait(min, max time.Duration) Option {
	return func(c *Config) {
		c.RetryWaitMin = min
		c.RetryWaitMax = max
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) {
		c.HTTPClient = client
	}
}

// WithUserAgent sets a custom user agent string.
func WithUserAgent(userAgent string) Option {
	return func(c *Config) {
		c.UserAgent = userAgent
	}
}

// WithClientToken sets the X-OpusDNS-Client product token identifying the calling client.
//
// The default names this library, which is what most callers want. Pass a token of your own
// ("acme-provisioner/2.1", say) when this library is embedded in a product that should be
// attributed in its own right, or "" to send no header at all.
func WithClientToken(token string) Option {
	return func(c *Config) {
		c.ClientToken = token
	}
}

// WithDebug enables debug logging.
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
	}
}

// WithLogger sets a custom logger for debug output.
func WithLogger(logger Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

// NewConfig creates a new Config with default values.
// Optionally applies the provided functional options.
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		APIEndpoint:  DefaultAPIEndpoint,
		APIVersion:   DefaultAPIVersion,
		TTL:          DefaultTTL,
		HTTPTimeout:  DefaultTimeout,
		MaxRetries:   DefaultMaxRetries,
		RetryWaitMin: DefaultRetryWaitMin,
		RetryWaitMax: DefaultRetryWaitMax,
		UserAgent:    GetUserAgent(),
		ClientToken:  GetClientToken(),
	}

	// Apply environment variables
	if apiKey := os.Getenv(EnvAPIKey); apiKey != "" {
		cfg.APIKey = apiKey
	}
	if endpoint := os.Getenv(EnvAPIEndpoint); endpoint != "" {
		cfg.APIEndpoint = endpoint
	}
	if version := os.Getenv(EnvAPIVersion); version != "" {
		cfg.APIVersion = version
	}
	if debug := os.Getenv(EnvDebug); debug == "true" || debug == "1" {
		cfg.Debug = true
	}

	// Apply functional options
	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// NewConfigFromEnv creates a new Config populated from environment variables.
// This is a convenience function that calls NewConfig() without additional options.
func NewConfigFromEnv() *Config {
	return NewConfig()
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return &ConfigError{Field: "APIKey", Message: "API key is required (set via config or OPUSDNS_API_KEY env var)"}
	}
	if c.APIEndpoint == "" {
		return &ConfigError{Field: "APIEndpoint", Message: "API endpoint is required"}
	}
	if c.APIVersion == "" {
		return &ConfigError{Field: "APIVersion", Message: "API version is required"}
	}
	if c.TTL < 0 {
		return &ConfigError{Field: "TTL", Message: "TTL must be non-negative"}
	}
	if c.HTTPTimeout < 0 {
		return &ConfigError{Field: "HTTPTimeout", Message: "HTTP timeout must be non-negative"}
	}
	if c.MaxRetries < 0 {
		return &ConfigError{Field: "MaxRetries", Message: "MaxRetries must be non-negative"}
	}
	if c.RetryWaitMin < 0 {
		return &ConfigError{Field: "RetryWaitMin", Message: "RetryWaitMin must be non-negative"}
	}
	if c.RetryWaitMax < 0 {
		return &ConfigError{Field: "RetryWaitMax", Message: "RetryWaitMax must be non-negative"}
	}
	if c.RetryWaitMin > c.RetryWaitMax {
		return &ConfigError{Field: "RetryWaitMin", Message: "RetryWaitMin must not exceed RetryWaitMax"}
	}
	return nil
}

// Clone creates a deep copy of the configuration.
func (c *Config) Clone() *Config {
	clone := *c
	return &clone
}

// WithOptions applies functional options to a copy of the configuration.
func (c *Config) WithOptions(opts ...Option) *Config {
	clone := c.Clone()
	for _, opt := range opts {
		opt(clone)
	}
	return clone
}
