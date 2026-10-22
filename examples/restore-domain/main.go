// Command restore-domain restores a deleted domain while it is still in the
// redemption grace period. An optional reason is passed on with the restore.
//
//	export OPUSDNS_API_KEY="opk_your_api_key"
//	go run main.go example.com "deleted by mistake"
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
		log.Fatal("Usage: restore-domain <domain> [reason]\nExample: restore-domain example.com \"deleted by mistake\"")
	}
	domain := os.Args[1]

	req := &models.DomainRestoreRequest{}
	if len(os.Args) > 2 {
		req.Reason = models.StringPtr(os.Args[2])
	}

	client, err := opusdns.NewClient(opusdns.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("Restoring domain: %s\n", domain)
	restored, err := client.Domains.RestoreDomain(ctx, domain, req)
	if err != nil {
		log.Fatalf("Failed to restore domain: %v", err)
	}

	fmt.Println("Domain restored successfully!")
	data, _ := json.MarshalIndent(restored, "", "  ")
	fmt.Println(string(data))
}
