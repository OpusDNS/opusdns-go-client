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
	// Get API key from environment variable
	apiKey := os.Getenv("OPUSDNS_API_KEY")
	if apiKey == "" {
		log.Fatal("OPUSDNS_API_KEY environment variable is required")
	}

	// Get zone name from command line argument
	if len(os.Args) < 2 {
		log.Fatal("Usage: create-zone <zone-name>\nExample: create-zone example.com")
	}
	zoneName := os.Args[1]

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

	// Create zone with no records
	fmt.Printf("Creating zone: %s\n", zoneName)
	changes, err := client.DNS.CreateZone(ctx, &models.ZoneCreateRequest{
		Name: zoneName,
	})
	if err != nil {
		log.Fatalf("Failed to create zone: %v", err)
	}
	fmt.Printf("Zone created successfully (%d changes)\n", changes.NumChanges)

	// The API answers with the changeset, so read the zone back to show it.
	zone, err := client.DNS.GetZone(ctx, zoneName)
	if err != nil {
		log.Fatalf("Failed to get zone: %v", err)
	}

	data, err := json.MarshalIndent(zone, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal zone: %v", err)
	}
	fmt.Println(string(data))
}
