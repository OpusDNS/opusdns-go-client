package cmd

import (
	"fmt"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/spf13/cobra"
)

var vanityNSCmd = &cobra.Command{
	Use:   "vanity-ns",
	Short: "Manage vanity nameserver sets",
	Long:  `List, get, create, delete, check, and set the default for vanity nameserver sets.`,
}

var vanityNSListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all vanity nameserver sets",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		sets, err := getClient().VanityNameservers.ListSets(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to list vanity nameserver sets: %w", err)
		}

		if len(sets) == 0 {
			fmt.Println("No vanity nameserver sets found.")
			return nil
		}

		fmt.Printf("Found %d vanity nameserver set(s):\n\n", len(sets))
		for _, set := range sets {
			def := ""
			if set.IsDefault {
				def = " [default]"
			}
			fmt.Printf("  • %s — %s (%s)%s\n", set.SetID, set.Name, set.Status, def)
		}

		return nil
	},
}

var vanityNSGetCmd = &cobra.Command{
	Use:   "get <set-id>",
	Short: "Get details of a vanity nameserver set",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		set, err := getClient().VanityNameservers.GetSet(ctx, models.VanityNameserverSetID(args[0]))
		if err != nil {
			return fmt.Errorf("failed to get vanity nameserver set: %w", err)
		}

		return printJSON(set)
	},
}

var vanityNSCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a vanity nameserver set",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		parentDomain, _ := cmd.Flags().GetString("parent-domain")
		soaRName, _ := cmd.Flags().GetString("soa-rname")
		hostnames, _ := cmd.Flags().GetStringArray("hostname")

		set, err := getClient().VanityNameservers.CreateSet(ctx, &models.VanityNameserverSetCreateRequest{
			Name:             args[0],
			ParentDomainName: parentDomain,
			SOARName:         soaRName,
			Hostnames:        hostnames,
		})
		if err != nil {
			return fmt.Errorf("failed to create vanity nameserver set: %w", err)
		}

		fmt.Printf("✓ Vanity nameserver set '%s' created (status: %s)!\n\n", set.SetID, set.Status)
		return printJSON(set)
	},
}

var vanityNSDeleteCmd = &cobra.Command{
	Use:   "delete <set-id>",
	Short: "Delete a vanity nameserver set",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		setID := args[0]

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Are you sure you want to delete vanity nameserver set '%s'?\n", setID)
			fmt.Print("Type 'yes' to confirm: ")
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		if err := getClient().VanityNameservers.DeleteSet(ctx, models.VanityNameserverSetID(setID)); err != nil {
			return fmt.Errorf("failed to delete vanity nameserver set: %w", err)
		}

		fmt.Printf("✓ Vanity nameserver set '%s' deletion requested!\n", setID)
		return nil
	},
}

var vanityNSCheckCmd = &cobra.Command{
	Use:   "check <set-id>",
	Short: "Run a read-only diagnostic on a vanity nameserver set",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		result, err := getClient().VanityNameservers.CheckSet(ctx, models.VanityNameserverSetID(args[0]))
		if err != nil {
			return fmt.Errorf("failed to check vanity nameserver set: %w", err)
		}

		return printJSON(result)
	},
}

var vanityNSSetDefaultCmd = &cobra.Command{
	Use:   "set-default <set-id>",
	Short: "Set a vanity nameserver set as the organization default",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		result, err := getClient().VanityNameservers.SetDefault(ctx, models.VanityNameserverSetID(args[0]))
		if err != nil {
			return fmt.Errorf("failed to set default vanity nameserver set: %w", err)
		}

		fmt.Println("✓ Default vanity nameserver set updated!")
		return printJSON(result)
	},
}

var vanityNSClearDefaultCmd = &cobra.Command{
	Use:   "clear-default",
	Short: "Unset the organization's default vanity nameserver set",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		result, err := getClient().VanityNameservers.ClearDefault(ctx)
		if err != nil {
			return fmt.Errorf("failed to clear default vanity nameserver set: %w", err)
		}

		if result.Cleared {
			fmt.Println("✓ Default vanity nameserver set cleared!")
		} else {
			fmt.Println("No default vanity nameserver set was set.")
		}
		return nil
	},
}

var vanityNSRestoreCmd = &cobra.Command{
	Use:   "restore <set-id>",
	Short: "Restore a suspended vanity nameserver set",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		set, err := getClient().VanityNameservers.RestoreSet(ctx, models.VanityNameserverSetID(args[0]))
		if err != nil {
			return fmt.Errorf("failed to restore vanity nameserver set: %w", err)
		}

		fmt.Printf("✓ Vanity nameserver set '%s' restored (status: %s)!\n\n", set.SetID, set.Status)
		return printJSON(set)
	},
}

var vanityNSZonesCmd = &cobra.Command{
	Use:   "zones <set-id>",
	Short: "List DNS zones referencing a vanity nameserver set",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		result, err := getClient().VanityNameservers.ListZonesReferencingSet(ctx, models.VanityNameserverSetID(args[0]), nil)
		if err != nil {
			return fmt.Errorf("failed to list zones referencing set: %w", err)
		}

		if len(result.Results) == 0 {
			fmt.Println("No zones reference this set.")
			return nil
		}

		fmt.Printf("Found %d zone(s):\n\n", len(result.Results))
		for _, zone := range result.Results {
			fmt.Printf("  • %s\n", zone.Name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(vanityNSCmd)

	vanityNSCmd.AddCommand(vanityNSListCmd)
	vanityNSCmd.AddCommand(vanityNSGetCmd)

	vanityNSCmd.AddCommand(vanityNSCreateCmd)
	vanityNSCreateCmd.Flags().String("parent-domain", "", "Apex domain of the vanity NS zone")
	vanityNSCreateCmd.Flags().String("soa-rname", "", "SOA RNAME stamped into the vanity NS zone")
	vanityNSCreateCmd.Flags().StringArray("hostname", nil, "Vanity NS hostname, ordered by position (repeatable)")

	vanityNSCmd.AddCommand(vanityNSDeleteCmd)
	vanityNSDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	vanityNSCmd.AddCommand(vanityNSCheckCmd)
	vanityNSCmd.AddCommand(vanityNSSetDefaultCmd)
	vanityNSCmd.AddCommand(vanityNSClearDefaultCmd)
	vanityNSCmd.AddCommand(vanityNSRestoreCmd)
	vanityNSCmd.AddCommand(vanityNSZonesCmd)
}
