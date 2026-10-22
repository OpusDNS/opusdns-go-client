package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
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
	// main prints the returned error once. Without this, cobra prints it too.
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Flags and arguments were valid, so an error from here on comes from
		// the command itself, and the usage text would not help.
		cmd.SilenceUsage = true

		if !needsClient(cmd) {
			return nil
		}

		// Get API key from flag or environment
		if apiKey == "" {
			apiKey = os.Getenv("OPUSDNS_API_KEY")
		}
		if apiKey == "" {
			return fmt.Errorf("API key is required. Set OPUSDNS_API_KEY or use --api-key flag")
		}

		opts := []opusdns.Option{
			opusdns.WithAPIKey(apiKey),
			opusdns.WithHTTPTimeout(timeout),
			// Debug output goes to stderr so that it never mixes with the
			// JSON a command prints to stdout.
			opusdns.WithLogger(log.New(os.Stderr, "", 0)),
		}
		// Without --debug the client keeps its own default, which honours
		// OPUSDNS_DEBUG.
		if cmd.Flags().Changed("debug") {
			opts = append(opts, opusdns.WithDebug(debug))
		}

		// Create client
		var err error
		client, err = opusdns.NewClient(opts...)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		return nil
	},
}

// needsClient reports whether cmd talks to the API. Help, version and shell
// completion work without an API key.
func needsClient(cmd *cobra.Command) bool {
	switch cmd.Name() {
	case "help", "version", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
		return false
	}
	// "opusdns completion bash" runs the "bash" subcommand of "completion".
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "completion" && c.Parent() == cmd.Root() {
			return false
		}
	}
	return true
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
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Write debug output to stderr (default from OPUSDNS_DEBUG)")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", opusdns.DefaultTimeout, "Timeout for each HTTP request (0 means none)")

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

// getContext returns the context for one command. It ends when the user
// interrupts the CLI, not after a fixed time: --timeout limits each HTTP
// request instead, so a paginated listing or a confirmation prompt is not
// cut off.
func getContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

// getClient returns the initialized client
func getClient() *opusdns.Client {
	return client
}

// printJSON marshals v as indented JSON and prints it to stdout.
func printJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
