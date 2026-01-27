package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/opusdns/opusdns-go-client/opusdns"
	"github.com/spf13/cobra"
)

var (
	apiKey  string
	debug   bool
	timeout time.Duration
	client  *opusdns.Client

	// Version information (set by main.go)
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "opusdns",
	Short: "OpusDNS CLI - Manage your DNS zones and domains",
	Long: `OpusDNS CLI is an interactive command-line tool for managing
your DNS zones, domains, contacts, and more through the OpusDNS API.

Set your API key via the OPUSDNS_API_KEY environment variable or use the --api-key flag.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip client initialization for help and version commands
		if cmd.Name() == "help" || cmd.Name() == "completion" || cmd.Name() == "version" {
			return nil
		}

		// Get API key from flag or environment
		if apiKey == "" {
			apiKey = os.Getenv("OPUSDNS_API_KEY")
		}
		if apiKey == "" {
			return fmt.Errorf("API key is required. Set OPUSDNS_API_KEY or use --api-key flag")
		}

		// Create client
		var err error
		client, err = opusdns.NewClient(
			opusdns.WithAPIKey(apiKey),
			opusdns.WithDebug(debug),
		)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

// SetVersion sets the version information from main.go
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "OpusDNS API key (or set OPUSDNS_API_KEY)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug output")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Second, "Request timeout")

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("opusdns %s\n", version)
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built:  %s\n", date)
		},
	})
}

// getContext returns a context with the configured timeout
func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// getClient returns the initialized client
func getClient() *opusdns.Client {
	return client
}
