package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

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
			renewMode := string(domain.RenewalMode)
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

		counts := summary.Domains

		fmt.Println("Domain Summary")
		fmt.Println("==============")
		fmt.Printf("Total domains:           %d\n", counts.TotalCount)
		if expiring := counts.ExpiringSoon; expiring != nil {
			fmt.Printf("Expiring within 30 days: %d\n", expiring.Next30Days)
			fmt.Printf("Expiring within 60 days: %d\n", expiring.Next60Days)
			fmt.Printf("Expiring within 90 days: %d\n", expiring.Next90Days)
		}

		printCounts("Domains by TLD", counts.ByTLD, ".")
		printCounts("Domains by Status", stringKeyed(counts.ByStatus), "")
		printCounts("Domains by Status Tag", stringKeyed(counts.ByStatusTag), "")
		printCounts("Domains by Organization", counts.ByOrganization, "")

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
	Long: `Update domain settings such as renewal mode.

Examples:
  opusdns domains update example.com --renewal-mode renew
  opusdns domains update example.com --renewal-mode expire`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		domainName := args[0]

		req := &models.DomainUpdateRequest{}
		hasChanges := false

		if cmd.Flags().Changed("renewal-mode") {
			renewalMode, _ := cmd.Flags().GetString("renewal-mode")
			mode := models.RenewalMode(renewalMode)
			req.RenewalMode = &mode
			hasChanges = true
		}

		if !hasChanges {
			return fmt.Errorf("no changes specified. Use --renewal-mode")
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

		}

		return nil
	},
}

var domainsCancelTransferCmd = &cobra.Command{
	Use:   "cancel-transfer <domain-name>",
	Short: "Cancel an in-progress domain transfer",
	Long:  `Cancel an in-progress domain transfer. This deletes the domain object.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		domain := args[0]

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Are you sure you want to cancel the transfer of '%s'? This deletes the domain object.\n", domain)
			fmt.Print("Type 'yes' to confirm: ")
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		if err := getClient().Domains.CancelTransfer(ctx, domain); err != nil {
			return fmt.Errorf("failed to cancel transfer: %w", err)
		}

		fmt.Printf("✓ Transfer of '%s' cancelled!\n", domain)
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

	// Check availability subcommand
	domainsCmd.AddCommand(domainsCheckCmd)

	// Cancel transfer subcommand
	domainsCmd.AddCommand(domainsCancelTransferCmd)
	domainsCancelTransferCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

// stringKeyed re-keys a count map so it can be printed by printCounts.
func stringKeyed[K ~string](counts map[K]int) map[string]int {
	out := make(map[string]int, len(counts))
	for key, count := range counts {
		out[string(key)] = count
	}
	return out
}

// printCounts writes a sorted count breakdown, or nothing when it is empty.
func printCounts(title string, counts map[string]int, prefix string) {
	if len(counts) == 0 {
		return
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fmt.Printf("\n%s:\n", title)
	for _, key := range keys {
		fmt.Printf("  %s%s: %d\n", prefix, key, counts[key])
	}
}
