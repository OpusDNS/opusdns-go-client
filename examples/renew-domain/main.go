// Command renew-domain renews a domain for a number of years (default 1).
//
//	export OPUSDNS_API_KEY="opk_your_api_key"
//	go run main.go example.com 2
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
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
		log.Fatal("Usage: renew-domain <domain> [years]\nExample: renew-domain example.com 2")
	}
	domain := os.Args[1]

	years := 1
	if len(os.Args) > 2 {
		n, err := strconv.Atoi(os.Args[2])
		if err != nil || n < 1 {
			log.Fatalf("Invalid years %q: must be a positive integer", os.Args[2])
		}
		years = n
	}

	client, err := opusdns.NewClient(opusdns.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// The API renews only when it is told the domain's current expiry date,
	// so read it first.
	current, err := client.Domains.GetDomain(ctx, domain)
	if err != nil {
		log.Fatalf("Failed to get domain: %v", err)
	}
	if current.ExpiresOn == nil {
		log.Fatalf("Domain %s has no expiry date to renew from", domain)
	}

	fmt.Printf("Renewing %s (expires %s) for %d year(s)\n", domain, current.ExpiresOn.Format("2006-01-02"), years)
	renewed, err := client.Domains.RenewDomain(ctx, domain, &models.DomainRenewRequest{
		Period:            models.DomainPeriod{Value: years, Unit: models.PeriodUnitYear},
		CurrentExpiryDate: current.ExpiresOn,
	})
	if err != nil {
		log.Fatalf("Failed to renew domain: %v", err)
	}

	if renewed.TransferPending() {
		fmt.Println("Domain is held at another registrar; a transfer into OpusDNS was requested.")
	} else {
		fmt.Println("Domain renewed successfully!")
	}
	data, _ := json.MarshalIndent(renewed, "", "  ")
	fmt.Println(string(data))
}
