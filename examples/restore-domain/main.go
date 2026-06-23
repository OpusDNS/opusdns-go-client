// Command restore-domain restores a deleted domain while it is still in the
// redemption grace period, renewing it for a number of years (default 1).
//
//	export OPUSDNS_API_KEY="opk_your_api_key"
//	go run main.go example.com 1
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
		log.Fatal("Usage: restore-domain <domain> [years]\nExample: restore-domain example.com 1")
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

	fmt.Printf("Restoring domain: %s\n", domain)
	restored, err := client.Domains.RestoreDomain(ctx, domain, &models.DomainRestoreRequest{
		Period: years,
	})
	if err != nil {
		log.Fatalf("Failed to restore domain: %v", err)
	}

	fmt.Println("Domain restored successfully!")
	data, _ := json.MarshalIndent(restored, "", "  ")
	fmt.Println(string(data))
}
