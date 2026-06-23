// Command register-domain registers a new domain, creating a registrant contact
// for it first.
//
//	export OPUSDNS_API_KEY="opk_your_api_key"
//	go run main.go example.com
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/opusdns/opusdns-go-client/opusdns"
)

func main() {
	apiKey := os.Getenv("OPUSDNS_API_KEY")
	if apiKey == "" {
		log.Fatal("OPUSDNS_API_KEY environment variable is required")
	}
	if len(os.Args) < 2 {
		log.Fatal("Usage: register-domain <domain>\nExample: register-domain example.com")
	}
	domain := os.Args[1]

	client, err := opusdns.NewClient(opusdns.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// A domain needs at least a registrant contact.
	contact, err := client.Contacts.CreateContact(ctx, &models.ContactCreateRequest{
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john@example.com",
		Phone:      "+1.5551234567",
		Street:     "123 Main Street",
		City:       "New York",
		PostalCode: "10001",
		Country:    "US",
		Disclose:   false,
	})
	if err != nil {
		log.Fatalf("Failed to create contact: %v", err)
	}

	handles := []models.ContactHandle{{ContactID: contact.ContactID}}
	fmt.Printf("Registering domain: %s\n", domain)
	registered, err := client.Domains.CreateDomain(ctx, &models.DomainCreateRequest{
		Name:        domain,
		Period:      models.DomainPeriod{Value: 1, Unit: models.PeriodUnitYear},
		RenewalMode: models.RenewalModeRenew,
		Contacts: map[models.DomainContactType][]models.ContactHandle{
			models.DomainContactTypeRegistrant: handles,
			models.DomainContactTypeAdmin:      handles,
			models.DomainContactTypeTech:       handles,
		},
		Nameservers: []models.Nameserver{
			{Hostname: "ns1.opusdns.com"},
			{Hostname: "ns2.opusdns.com"},
		},
		CreateZone: true, // also provision a DNS zone on OpusDNS nameservers
	})
	if err != nil {
		log.Fatalf("Failed to register domain: %v", err)
	}

	fmt.Println("Domain registered successfully!")
	data, _ := json.MarshalIndent(registered, "", "  ")
	fmt.Println(string(data))
}
