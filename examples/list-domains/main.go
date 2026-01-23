package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/opusdns/opusdns-go-client/opusdns"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("OPUSDNS_API_KEY")
	if apiKey == "" {
		log.Fatal("OPUSDNS_API_KEY environment variable is required")
	}

	// Create client
	client, err := opusdns.NewClient(
		opusdns.WithAPIKey(apiKey),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch all domains
	domains, err := client.Domains.ListDomains(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list domains: %v", err)
	}

	// Print results
	fmt.Printf("Found %d domains:\n\n", len(domains))
	for _, domain := range domains {
		data, err := json.MarshalIndent(domain, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal domain %s: %v", domain.Name, err)
			continue
		}
		fmt.Printf("%s\n\n", string(data))
	}
}
