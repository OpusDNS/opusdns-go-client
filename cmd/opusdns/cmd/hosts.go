package cmd

import (
	"fmt"

	"github.com/opusdns/opusdns-go-client/models"
	"github.com/spf13/cobra"
)

var hostsCmd = &cobra.Command{
	Use:   "hosts",
	Short: "Manage host objects",
	Long:  `Create, get, update, and delete host objects. A host is referenced by either its ID or its hostname.`,
}

var hostsCreateCmd = &cobra.Command{
	Use:   "create <hostname>",
	Short: "Create a host object",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		ips, _ := cmd.Flags().GetStringArray("ip")

		host, err := getClient().Hosts.CreateHost(ctx, &models.HostCreateRequest{
			Hostname:    args[0],
			IPAddresses: ips,
		})
		if err != nil {
			return fmt.Errorf("failed to create host: %w", err)
		}

		fmt.Printf("✓ Host '%s' created successfully!\n\n", host.Hostname)
		return printJSON(host)
	},
}

var hostsGetCmd = &cobra.Command{
	Use:   "get <host-id-or-hostname>",
	Short: "Get details of a host object",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		host, err := getClient().Hosts.GetHost(ctx, args[0])
		if err != nil {
			return fmt.Errorf("failed to get host: %w", err)
		}

		return printJSON(host)
	},
}

var hostsUpdateCmd = &cobra.Command{
	Use:   "update <host-id-or-hostname>",
	Short: "Update a host object's IP addresses",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		ips, _ := cmd.Flags().GetStringArray("ip")

		host, err := getClient().Hosts.UpdateHost(ctx, args[0], &models.HostUpdateRequest{
			IPAddresses: ips,
		})
		if err != nil {
			return fmt.Errorf("failed to update host: %w", err)
		}

		fmt.Printf("✓ Host '%s' updated successfully!\n\n", host.Hostname)
		return printJSON(host)
	},
}

var hostsDeleteCmd = &cobra.Command{
	Use:   "delete <host-id-or-hostname>",
	Short: "Delete a host object",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := getContext()
		defer cancel()

		reference := args[0]

		if err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete host '%s'?", reference)); err != nil {
			return err
		}

		if err := getClient().Hosts.DeleteHost(ctx, reference); err != nil {
			return fmt.Errorf("failed to delete host: %w", err)
		}

		fmt.Printf("✓ Host '%s' deleted successfully!\n", reference)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(hostsCmd)

	hostsCmd.AddCommand(hostsCreateCmd)
	hostsCreateCmd.Flags().StringArray("ip", nil, "IP address for the host (repeatable, at least one)")
	_ = hostsCreateCmd.MarkFlagRequired("ip")

	hostsCmd.AddCommand(hostsGetCmd)

	hostsCmd.AddCommand(hostsUpdateCmd)
	hostsUpdateCmd.Flags().StringArray("ip", nil, "Full replacement set of IP addresses (repeatable, at least one)")
	_ = hostsUpdateCmd.MarkFlagRequired("ip")

	hostsCmd.AddCommand(hostsDeleteCmd)
	hostsDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}
