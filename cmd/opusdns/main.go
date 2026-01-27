package main

import (
	"fmt"
	"os"

	"github.com/opusdns/opusdns-go-client/cmd/opusdns/cmd"
)

// Build-time variables set by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersion(version, commit, date)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
