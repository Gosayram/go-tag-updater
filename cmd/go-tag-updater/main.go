// Package main provides the CLI entry point for go-tag-updater tool.
package main

import (
	"fmt"
	"os"
)

const (
	// ExitCodeSuccess indicates successful program execution
	ExitCodeSuccess = 0
	// ExitCodeError indicates program execution failed
	ExitCodeError = 1
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(ExitCodeError)
	}
}
