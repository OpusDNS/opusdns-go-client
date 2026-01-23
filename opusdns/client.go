// Package opusdns provides a Go client library for the OpusDNS API.
package opusdns

// Client is the high-level OpusDNS API client.
// It provides access to all API services through dedicated service objects.
type Client struct {
	// Config holds the client configuration.
	Config *Config

	// http is the underlying HTTP client.
	http *HTTPClient

	// DNS provides access to DNS zone and record management.
	DNS *DNSService

	// Domains provides access to domain registration and management.
	Domains *DomainsService

	// Contacts provides access to contact management.
	Contacts *ContactsService

	// EmailForwards provides access to email forwarding configuration.
	EmailForwards *EmailForwardsService

	// DomainForwards provides access to domain/URL forwarding configuration.
	DomainForwards *DomainForwardsService

	// TLDs provides access to TLD information and portfolio.
	TLDs *TLDsService

	// Availability provides access to domain availability checking.
	Availability *AvailabilityService

	// Organizations provides access to organization management.
	Organizations *OrganizationsService

	// Users provides access to user management.
	Users *UsersService

	// Events provides access to event and audit log data.
	Events *EventsService
}

// NewClient creates a new OpusDNS client with the given options.
//
// Example:
//
//	client, err := opusdns.NewClient(
//	    opusdns.WithAPIKey("opk_your_api_key"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// The API key can also be set via the OPUSDNS_API_KEY environment variable.
func NewClient(opts ...Option) (*Client, error) {
	config := NewConfig(opts...)

	httpClient, err := NewHTTPClient(config)
	if err != nil {
		return nil, err
	}

	client := &Client{
		Config: config,
		http:   httpClient,
	}

	// Initialize all services
	client.DNS = &DNSService{client: client}
	client.Domains = &DomainsService{client: client}
	client.Contacts = &ContactsService{client: client}
	client.EmailForwards = &EmailForwardsService{client: client}
	client.DomainForwards = &DomainForwardsService{client: client}
	client.TLDs = &TLDsService{client: client}
	client.Availability = &AvailabilityService{client: client}
	client.Organizations = &OrganizationsService{client: client}
	client.Users = &UsersService{client: client}
	client.Events = &EventsService{client: client}

	return client, nil
}

// NewClientWithConfig creates a new OpusDNS client with an existing configuration.
func NewClientWithConfig(config *Config) (*Client, error) {
	if config == nil {
		config = NewConfig()
	}

	httpClient, err := NewHTTPClient(config)
	if err != nil {
		return nil, err
	}

	client := &Client{
		Config: config,
		http:   httpClient,
	}

	// Initialize all services
	client.DNS = &DNSService{client: client}
	client.Domains = &DomainsService{client: client}
	client.Contacts = &ContactsService{client: client}
	client.EmailForwards = &EmailForwardsService{client: client}
	client.DomainForwards = &DomainForwardsService{client: client}
	client.TLDs = &TLDsService{client: client}
	client.Availability = &AvailabilityService{client: client}
	client.Organizations = &OrganizationsService{client: client}
	client.Users = &UsersService{client: client}
	client.Events = &EventsService{client: client}

	return client, nil
}

// DefaultTTL returns the configured default TTL for DNS records.
func (c *Client) DefaultTTL() int {
	return c.Config.TTL
}

// HTTPClient returns the underlying HTTP client for advanced use cases.
func (c *Client) HTTPClient() *HTTPClient {
	return c.http
}
