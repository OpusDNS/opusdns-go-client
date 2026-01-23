package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/spf13/cobra"
)

var zonesCmd = &cobra.Command{
	Use:   "zones",
	Short: "Manage DNS zones",
	Long:  `List, create, get, and delete DNS zones.`,
}

var zonesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all DNS zones",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		search, _ := cmd.Flags().GetString("search")
		opts := &models.ListZonesOptions{}
		if search != "" {
			opts.Search = search
		}

		zones, err := getClient().DNS.ListZones(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to list zones: %w", err)
		}

		if len(zones) == 0 {
			fmt.Println("No zones found.")
			return nil
		}

		fmt.Printf("Found %d zone(s):\n\n", len(zones))
		for _, zone := range zones {
			dnssec := string(zone.DNSSECStatus)
			if dnssec == "" {
				dnssec = "unknown"
			}
			fmt.Printf("  • %s (DNSSEC: %s)\n", zone.Name, dnssec)
		}

		return nil
	},
}

var zonesGetCmd = &cobra.Command{
	Use:   "get <zone-name>",
	Short: "Get details of a DNS zone",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		zoneName := args[0]

		zone, err := getClient().DNS.GetZone(ctx, zoneName)
		if err != nil {
			return fmt.Errorf("failed to get zone: %w", err)
		}

		data, err := json.MarshalIndent(zone, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format zone: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var zonesCreateCmd = &cobra.Command{
	Use:   "create <zone-name>",
	Short: "Create a new DNS zone",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		zoneName := args[0]

		zone, err := getClient().DNS.CreateZone(ctx, &models.ZoneCreateRequest{
			Name: zoneName,
		})
		if err != nil {
			return fmt.Errorf("failed to create zone: %w", err)
		}

		fmt.Printf("✓ Zone '%s' created successfully!\n\n", zone.Name)

		data, err := json.MarshalIndent(zone, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format zone: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var zonesDeleteCmd = &cobra.Command{
	Use:   "delete <zone-name>",
	Short: "Delete a DNS zone",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		zoneName := args[0]

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Are you sure you want to delete zone '%s'? This action cannot be undone.\n", zoneName)
			fmt.Print("Type 'yes' to confirm: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		err := getClient().DNS.DeleteZone(ctx, zoneName)
		if err != nil {
			return fmt.Errorf("failed to delete zone: %w", err)
		}

		fmt.Printf("✓ Zone '%s' deleted successfully!\n", zoneName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(zonesCmd)

	// List subcommand
	zonesCmd.AddCommand(zonesListCmd)
	zonesListCmd.Flags().String("search", "", "Search zones by name")

	// Get subcommand
	zonesCmd.AddCommand(zonesGetCmd)

	// Create subcommand
	zonesCmd.AddCommand(zonesCreateCmd)

	// Delete subcommand
	zonesCmd.AddCommand(zonesDeleteCmd)
	zonesDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}
