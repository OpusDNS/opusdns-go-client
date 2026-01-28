// Package client provides a Go client library for the OpusDNS API.
package opusdns

import (
	"net/http"
	"os"
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

// UserAgent returns the default user agent string.
func GetUserAgent() string {
	return "opusdns-go-client/" + Version
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
	// Default: opusdns-go-client/1.0.0
	UserAgent string

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
