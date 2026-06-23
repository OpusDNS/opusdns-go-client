// Command transfer-domain starts an inbound transfer of a domain using its auth
// code, creating a registrant contact for it first.
//
//	export OPUSDNS_API_KEY="opk_your_api_key"
//	go run main.go example.com <auth-code>
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
	if len(os.Args) < 3 {
		log.Fatal("Usage: transfer-domain <domain> <auth-code>\nExample: transfer-domain example.com abc123xyz")
	}
	domain, authCode := os.Args[1], os.Args[2]

	client, err := opusdns.NewClient(opusdns.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

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

	fmt.Printf("Transferring domain: %s\n", domain)
	transferred, err := client.Domains.TransferDomain(ctx, &models.DomainTransferRequest{
		Name:        domain,
		AuthCode:    authCode,
		RenewalMode: models.RenewalModeRenew,
		Contacts: map[models.DomainContactType][]models.ContactHandle{
			models.DomainContactTypeRegistrant: {{ContactID: contact.ContactID}},
		},
	})
	if err != nil {
		log.Fatalf("Failed to transfer domain: %v", err)
	}

	fmt.Println("Transfer started! (cancel a pending transfer with client.Domains.CancelTransfer)")
	data, _ := json.MarshalIndent(transferred, "", "  ")
	fmt.Println(string(data))
}
