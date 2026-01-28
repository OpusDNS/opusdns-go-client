package main

import (
	"fmt"
	"os"

	"github.com/opusdns/opusdns-go-client/cmd/opusdns/cmd"
	"github.com/opusdns/opusdns-go-client/opusdns"
)

// Build-time variables set by goreleaser
var (
	commit = "none"
	date   = "unknown"
)

func main() {
	cmd.SetVersion(opusdns.Version, commit, date)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
