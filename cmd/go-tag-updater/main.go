// Package main provides the CLI entry point for go-tag-updater tool.
package main

import (
	"fmt"
	"os"

	"github.com/Gosayram/go-tag-updater/pkg/errors"
)

const (
	// ExitCodeSuccess indicates successful program execution
	ExitCodeSuccess = 0
	// ExitCodeError indicates program execution failed
	ExitCodeError = 1
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			fmt.Fprintf(os.Stderr, "Error [%s]: %s\n", appErr.Category, appErr.Message)
			if appErr.Context != "" {
				fmt.Fprintf(os.Stderr, "Context: %s\n", appErr.Context)
			}
			os.Exit(appErr.Code)
		}
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(ExitCodeError)
	}
}
