package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/spf13/cobra"
)

var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "Manage domains",
	Long:  `List, get, and manage domain registrations.`,
}

var domainsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all domains",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		search, _ := cmd.Flags().GetString("search")
		tld, _ := cmd.Flags().GetString("tld")

		opts := &models.ListDomainsOptions{}
		if search != "" {
			opts.Search = search
		}
		if tld != "" {
			opts.TLD = tld
		}

		domains, err := getClient().Domains.ListDomains(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to list domains: %w", err)
		}

		if len(domains) == 0 {
			fmt.Println("No domains found.")
			return nil
		}

		fmt.Printf("Found %d domain(s):\n\n", len(domains))
		for _, domain := range domains {
			expiresOn := "N/A"
			if domain.ExpiresOn != nil {
				expiresOn = domain.ExpiresOn.Format("2006-01-02")
			}
			renewMode := string(domain.RenewMode)
			if renewMode == "" {
				renewMode = "unknown"
			}
			fmt.Printf("  • %s (expires: %s, renewal: %s)\n", domain.Name, expiresOn, renewMode)
		}

		return nil
	},
}

var domainsGetCmd = &cobra.Command{
	Use:   "get <domain-name>",
	Short: "Get details of a domain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		domainName := args[0]

		domain, err := getClient().Domains.GetDomain(ctx, domainName)
		if err != nil {
			return fmt.Errorf("failed to get domain: %w", err)
		}

		data, err := json.MarshalIndent(domain, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format domain: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var domainsSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Get a summary of all domains",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		summary, err := getClient().Domains.GetSummary(ctx)
		if err != nil {
			return fmt.Errorf("failed to get domain summary: %w", err)
		}

		fmt.Println("Domain Summary")
		fmt.Println("==============")
		fmt.Printf("Total domains:           %d\n", summary.TotalDomains)
		fmt.Printf("Expiring within 30 days: %d\n", summary.ExpiringWithin30Days)
		fmt.Printf("Expiring within 90 days: %d\n", summary.ExpiringWithin90Days)

		if len(summary.DomainsByTLD) > 0 {
			fmt.Println("\nDomains by TLD:")
			for tld, count := range summary.DomainsByTLD {
				fmt.Printf("  .%s: %d\n", tld, count)
			}
		}

		if len(summary.DomainsByStatus) > 0 {
			fmt.Println("\nDomains by Status:")
			for status, count := range summary.DomainsByStatus {
				fmt.Printf("  %s: %d\n", status, count)
			}
		}

		return nil
	},
}

var domainsRenewCmd = &cobra.Command{
	Use:   "renew <domain-name>",
	Short: "Renew a domain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		domainName := args[0]
		period, _ := cmd.Flags().GetInt("period")

		if period <= 0 {
			period = 1
		}

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Are you sure you want to renew '%s' for %d year(s)?\n", domainName, period)
			fmt.Print("Type 'yes' to confirm: ")
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		domain, err := getClient().Domains.RenewDomain(ctx, domainName, &models.DomainRenewRequest{
			Period: period,
		})
		if err != nil {
			return fmt.Errorf("failed to renew domain: %w", err)
		}

		fmt.Printf("✓ Domain '%s' renewed successfully!\n\n", domain.Name)

		if domain.ExpiresOn != nil {
			fmt.Printf("New expiration date: %s\n", domain.ExpiresOn.Format("2006-01-02"))
		}

		return nil
	},
}

var domainsUpdateCmd = &cobra.Command{
	Use:   "update <domain-name>",
	Short: "Update domain settings",
	Long: `Update domain settings such as renewal mode and transfer lock.

Examples:
  opusdns domains update example.com --renewal-mode renew
  opusdns domains update example.com --transfer-lock=true`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		domainName := args[0]

		req := &models.DomainUpdateRequest{}
		hasChanges := false

		if cmd.Flags().Changed("renewal-mode") {
			renewalMode, _ := cmd.Flags().GetString("renewal-mode")
			mode := models.RenewMode(renewalMode)
			req.RenewMode = &mode
			hasChanges = true
		}

		if cmd.Flags().Changed("transfer-lock") {
			transferLock, _ := cmd.Flags().GetBool("transfer-lock")
			req.TransferLock = &transferLock
			hasChanges = true
		}

		if !hasChanges {
			return fmt.Errorf("no changes specified. Use --renewal-mode or --transfer-lock")
		}

		domain, err := getClient().Domains.UpdateDomain(ctx, domainName, req)
		if err != nil {
			return fmt.Errorf("failed to update domain: %w", err)
		}

		fmt.Printf("✓ Domain '%s' updated successfully!\n\n", domain.Name)

		data, err := json.MarshalIndent(domain, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format domain: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var domainsCheckCmd = &cobra.Command{
	Use:   "check <domain-name> [domain-name...]",
	Short: "Check domain availability",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		result, err := getClient().Availability.CheckAvailability(ctx, args)
		if err != nil {
			return fmt.Errorf("failed to check availability: %w", err)
		}

		fmt.Printf("Availability check (%dms):\n\n", result.Meta.ProcessingTimeMs)
		for _, avail := range result.Results {
			status := "❌ unavailable"
			if avail.Status.IsAvailable() {
				status = "✓ available"
			}
			fmt.Printf("  %s: %s\n", avail.Domain, status)

			if avail.Price != nil && avail.Price.RegisterPrice != nil {
				fmt.Printf("      Price: %s %s\n", *avail.Price.RegisterPrice, avail.Price.Currency)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(domainsCmd)

	// List subcommand
	domainsCmd.AddCommand(domainsListCmd)
	domainsListCmd.Flags().String("search", "", "Search domains by name")
	domainsListCmd.Flags().String("tld", "", "Filter by TLD")

	// Get subcommand
	domainsCmd.AddCommand(domainsGetCmd)

	// Summary subcommand
	domainsCmd.AddCommand(domainsSummaryCmd)

	// Renew subcommand
	domainsCmd.AddCommand(domainsRenewCmd)
	domainsRenewCmd.Flags().Int("period", 1, "Renewal period in years")
	domainsRenewCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	// Update subcommand
	domainsCmd.AddCommand(domainsUpdateCmd)
	domainsUpdateCmd.Flags().String("renewal-mode", "", "Renewal mode (renew or expire)")
	domainsUpdateCmd.Flags().Bool("transfer-lock", false, "Enable/disable transfer lock")

	// Check availability subcommand
	domainsCmd.AddCommand(domainsCheckCmd)
}
