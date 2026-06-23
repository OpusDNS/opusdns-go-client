// Command delete-domain deletes a domain, moving it into the redemption /
// pending-delete window. Use restore-domain to bring it back during redemption.
//
//	export OPUSDNS_API_KEY="opk_your_api_key"
//	go run main.go example.com
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/opusdns/opusdns-go-client/opusdns"
)

func main() {
	apiKey := os.Getenv("OPUSDNS_API_KEY")
	if apiKey == "" {
		log.Fatal("OPUSDNS_API_KEY environment variable is required")
	}
	if len(os.Args) < 2 {
		log.Fatal("Usage: delete-domain <domain>\nExample: delete-domain example.com")
	}
	domain := os.Args[1]

	client, err := opusdns.NewClient(opusdns.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("Deleting domain: %s\n", domain)
	if err := client.Domains.DeleteDomain(ctx, domain); err != nil {
		log.Fatalf("Failed to delete domain: %v", err)
	}

	fmt.Println("Domain deleted (entered redemption / pending-delete).")
}
