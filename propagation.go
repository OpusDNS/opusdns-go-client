package opusdns

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// WaitForPropagation polls DNS servers to verify that a TXT record has propagated.
// It returns nil when the record is found with the expected value, or an error if
// the timeout is reached or an error occurs.
func (c *Client) WaitForPropagation(fqdn, expectedValue string) error {
	// Ensure FQDN ends with a dot for DNS queries
	if !strings.HasSuffix(fqdn, ".") {
		fqdn = fqdn + "."
	}

	// Remove quotes from expected value for comparison
	expectedValue = strings.Trim(expectedValue, "\"")

	ctx, cancel := context.WithTimeout(context.Background(), c.config.PollingTimeout)
	defer cancel()

	ticker := time.NewTicker(c.config.PollingInterval)
	defer ticker.Stop()

	attempt := 0
	maxAttempts := int(c.config.PollingTimeout / c.config.PollingInterval)

	for {
		attempt++

		// Try all configured DNS resolvers
		for _, resolver := range c.config.DNSResolvers {
			found, err := c.checkDNSRecord(fqdn, expectedValue, resolver)
			if err == nil && found {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("DNS propagation timeout after %d attempts (%v): record not found for %s",
				attempt, c.config.PollingTimeout, fqdn)
		case <-ticker.C:
			if attempt >= maxAttempts {
				return fmt.Errorf("DNS propagation timeout after %d attempts: record not found for %s",
					attempt, fqdn)
			}
			// Continue to next iteration
		}
	}
}

// checkDNSRecord queries a specific DNS resolver for a TXT record and checks if it matches the expected value.
func (c *Client) checkDNSRecord(fqdn, expectedValue, resolver string) (bool, error) {
	m := new(dns.Msg)
	m.SetQuestion(fqdn, dns.TypeTXT)
	m.RecursionDesired = true

	dnsClient := &dns.Client{
		Timeout: 5 * time.Second,
	}

	r, _, err := dnsClient.Exchange(m, resolver)
	if err != nil {
		return false, fmt.Errorf("DNS query failed for %s at %s: %w", fqdn, resolver, err)
	}

	if r.Rcode != dns.RcodeSuccess {
		return false, nil
	}

	// Check all TXT records in the answer section
	for _, ans := range r.Answer {
		if txt, ok := ans.(*dns.TXT); ok {
			for _, record := range txt.Txt {
				// Compare without quotes
				cleanRecord := strings.Trim(record, "\"")
				if cleanRecord == expectedValue {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
