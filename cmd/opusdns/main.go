package main

import (
	"fmt"
	"os"

	"github.com/opusdns/opusdns-go-client/cmd/opusdns/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
