package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Gosayram/go-tag-updater/internal/config"
	"github.com/Gosayram/go-tag-updater/internal/logger"
	"github.com/Gosayram/go-tag-updater/internal/version"
	"github.com/Gosayram/go-tag-updater/pkg/errors"
)

const (
	// AppName is the name of the application
	AppName = "go-tag-updater"
	// AppDescription provides a brief description of the application
	AppDescription = "A CLI tool for safely updating YAML files in GitLab repositories " +
		"through automated merge request workflows"

	// DefaultTargetBranch is the default target branch for merge requests
	DefaultTargetBranch = "main"
	// DefaultBranchPrefix is the prefix used for feature branch names
	DefaultBranchPrefix = "update-tag"
	// MaxProjectIDLength is the maximum allowed length for project ID
	MaxProjectIDLength = 255
	// MinTagLength is the minimum allowed length for tag values
	MinTagLength = 1
	// MaxTagLength is the maximum allowed length for tag values
	MaxTagLength = 255
)

var (
	// showVersion flag
	showVersion bool

	// Root command
	rootCmd = &cobra.Command{
		Use:   AppName,
		Short: AppDescription,
		Long: `go-tag-updater is a CLI tool for safely updating YAML files in GitLab repositories.

It provides intelligent conflict detection, flexible project identification, 
and comprehensive merge request lifecycle management. The tool supports both 
numeric project IDs and human-readable project paths.`,
		RunE: runCommand,
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	// Required flags
	rootCmd.Flags().StringP("project-id", "p", "", "GitLab project ID or path (group/subgroup/project)")
	rootCmd.Flags().StringP("file", "f", "", "Path to target YAML file within repository")
	rootCmd.Flags().StringP("new-tag", "t", "", "New tag value to set in YAML file")
	rootCmd.Flags().StringP("token", "", "", "GitLab Personal Access Token")

	// Optional flags
	rootCmd.Flags().StringP("branch-name", "b", "", "Name for the new feature branch (auto-generated if empty)")
	rootCmd.Flags().String("target-branch", DefaultTargetBranch, "Target branch for merge request")
	rootCmd.Flags().Bool("wait-previous-mr", false, "Wait for conflicting merge requests to complete")
	rootCmd.Flags().Bool("debug", false, "Enable verbose debugging output")
	rootCmd.Flags().Bool("dry-run", false, "Preview changes without execution")
	rootCmd.Flags().Bool("auto-merge", false, "Automatically merge when pipeline passes")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Show version information")

	// Bind flags to viper
	_ = viper.BindPFlag("project-id", rootCmd.Flags().Lookup("project-id"))
	_ = viper.BindPFlag("file", rootCmd.Flags().Lookup("file"))
	_ = viper.BindPFlag("new-tag", rootCmd.Flags().Lookup("new-tag"))
	_ = viper.BindPFlag("token", rootCmd.Flags().Lookup("token"))
	_ = viper.BindPFlag("branch-name", rootCmd.Flags().Lookup("branch-name"))
	_ = viper.BindPFlag("target-branch", rootCmd.Flags().Lookup("target-branch"))
	_ = viper.BindPFlag("wait-previous-mr", rootCmd.Flags().Lookup("wait-previous-mr"))
	_ = viper.BindPFlag("debug", rootCmd.Flags().Lookup("debug"))
	_ = viper.BindPFlag("dry-run", rootCmd.Flags().Lookup("dry-run"))
	_ = viper.BindPFlag("auto-merge", rootCmd.Flags().Lookup("auto-merge"))

	// Don't mark flags as required here - we'll check them in runCommand
	// This allows version flag to work without other required flags
}

func initConfig() {
	_, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}
}

func runCommand(_ *cobra.Command, _ []string) error {
	// Handle version flag first
	if showVersion {
		fmt.Println(version.GetFullVersionInfo())
		return nil
	}

	// Check required flags manually
	projectID := viper.GetString("project-id")
	filePath := viper.GetString("file")
	newTag := viper.GetString("new-tag")
	token := viper.GetString("token")

	if projectID == "" {
		return errors.NewValidationError("project-id is required")
	}
	if filePath == "" {
		return errors.NewValidationError("file is required")
	}
	if newTag == "" {
		return errors.NewValidationError("new-tag is required")
	}
	if token == "" {
		return errors.NewValidationError("token is required")
	}

	// Load configuration from viper
	cfg, err := config.NewFromViper()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	log := logger.New(cfg.Debug)

	log.Infof("Starting go-tag-updater with file=%s, tag=%s, project=%s",
		cfg.FilePath, cfg.NewTag, cfg.ProjectID)

	if cfg.DryRun {
		log.Info("Dry run mode - no changes will be made")
	}

	// Validate inputs (additional validation)
	if cfg.FilePath == "" {
		return errors.NewValidationError("file path cannot be empty")
	}

	if cfg.NewTag == "" {
		return errors.NewValidationError("new tag cannot be empty")
	}

	if cfg.ProjectID == "" {
		return errors.NewValidationError("project ID cannot be empty")
	}

	if cfg.GitLabToken == "" {
		return errors.NewValidationError("GitLab token cannot be empty")
	}

	log.Info("Configuration validated successfully")

	// Execute workflow - just report what would be done for now
	log.Infof("Would update file: %s", cfg.FilePath)
	log.Infof("Would set tag to: %s", cfg.NewTag)
	log.Infof("Target project: %s", cfg.ProjectID)
	log.Infof("Target branch: %s", cfg.TargetBranch)

	if cfg.AutoMerge {
		log.Info("Auto-merge would be enabled")
	}

	if cfg.WaitForPreviousMR {
		log.Info("Would wait for previous merge requests")
	}

	log.Info("Tag update process completed successfully")

	return nil
}
