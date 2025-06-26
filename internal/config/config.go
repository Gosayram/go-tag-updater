// Package config provides configuration management for go-tag-updater.
package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

const (
	// DefaultConfigFile specifies the primary configuration file name
	DefaultConfigFile = "go-tag-updater.yaml"
	// AlternateConfigFile specifies the alternative configuration file name
	AlternateConfigFile = "go-tag-updater.yml"
	// EnvPrefix defines the prefix for environment variables
	EnvPrefix = "GO_TAG_UPDATER"
	// DefaultMergeTimeout specifies the default timeout for merge operations
	DefaultMergeTimeout = 300 * time.Second

	// DefaultBufferSize specifies the default buffer size for I/O operations
	DefaultBufferSize = 1024
	// DefaultTimeout specifies the default timeout for API requests
	DefaultTimeout = 30 * time.Second
	// MaxConfigFileSize defines the maximum allowed configuration file size
	MaxConfigFileSize = 1024 * 1024 // 1MB
	// ConfigFilePermission defines the file permissions for configuration files
	ConfigFilePermission = 0o644

	// DefaultRateLimitRPS specifies the default rate limit in requests per second
	DefaultRateLimitRPS = 10
	// DefaultMaxConcurrentReqs specifies the default maximum concurrent requests
	DefaultMaxConcurrentReqs = 5
	// DefaultRetryCount specifies the default number of retry attempts
	DefaultRetryCount = 3
)

// Config holds the application configuration
type Config struct {
	// GitLab settings
	GitLab GitLabConfig `mapstructure:"gitlab"`

	// Default behavior settings
	Defaults DefaultsConfig `mapstructure:"defaults"`

	// Performance settings
	Performance PerformanceConfig `mapstructure:"performance"`

	// Logging settings
	Logging LoggingConfig `mapstructure:"logging"`
}

// GitLabConfig contains GitLab-specific configuration
type GitLabConfig struct {
	BaseURL      string        `mapstructure:"base_url"`
	Token        string        `mapstructure:"token"`
	Timeout      time.Duration `mapstructure:"timeout"`
	RetryCount   int           `mapstructure:"retry_count"`
	RateLimitRPS int           `mapstructure:"rate_limit_rps"`
}

// DefaultsConfig contains default values for common operations
type DefaultsConfig struct {
	TargetBranch   string        `mapstructure:"target_branch"`
	BranchPrefix   string        `mapstructure:"branch_prefix"`
	MergeTimeout   time.Duration `mapstructure:"merge_timeout"`
	WaitPreviousMR bool          `mapstructure:"wait_previous_mr"`
	AutoMerge      bool          `mapstructure:"auto_merge"`
}

// PerformanceConfig contains performance-related settings
type PerformanceConfig struct {
	MaxConcurrentRequests int           `mapstructure:"max_concurrent_requests"`
	RequestTimeout        time.Duration `mapstructure:"request_timeout"`
	BufferSize            int           `mapstructure:"buffer_size"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	EnableFile bool   `mapstructure:"enable_file"`
	FilePath   string `mapstructure:"file_path"`
}

// CLIConfig represents configuration from command line arguments
type CLIConfig struct {
	// Required fields
	ProjectID string
	FilePath  string
	NewTag    string

	// GitLab configuration
	GitLabToken string
	GitLabURL   string

	// Branch configuration
	BranchName   string
	TargetBranch string

	// Behavior flags
	WaitForPreviousMR bool
	AutoMerge         bool
	DryRun            bool
	Debug             bool

	// Logging configuration
	LogLevel  string
	LogFormat string

	// Timeouts
	Timeout time.Duration
}

// NewFromViper creates a CLI configuration from viper values
func NewFromViper() (*CLIConfig, error) {
	return &CLIConfig{
		ProjectID:         viper.GetString("project-id"),
		FilePath:          viper.GetString("file"),
		NewTag:            viper.GetString("new-tag"),
		GitLabToken:       viper.GetString("token"),
		GitLabURL:         viper.GetString("gitlab-url"),
		BranchName:        viper.GetString("branch-name"),
		TargetBranch:      viper.GetString("target-branch"),
		WaitForPreviousMR: viper.GetBool("wait-previous-mr"),
		AutoMerge:         viper.GetBool("auto-merge"),
		DryRun:            viper.GetBool("dry-run"),
		Debug:             viper.GetBool("debug"),
		LogLevel:          viper.GetString("log-level"),
		LogFormat:         viper.GetString("log-format"),
		Timeout:           viper.GetDuration("timeout"),
	}, nil
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	return LoadFromFile(DefaultConfigFile)
}

// LoadFromFile loads configuration from a specific file
func LoadFromFile(configFile string) (*Config, error) {
	viper.SetConfigName("go-tag-updater")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath("/etc/go-tag-updater")

	// Set environment variable prefix
	viper.SetEnvPrefix(EnvPrefix)
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Try to read config file if it exists
	if configFile != "" && fileExists(configFile) {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// GitLab defaults
	viper.SetDefault("gitlab.base_url", "https://gitlab.com")
	viper.SetDefault("gitlab.timeout", DefaultTimeout)
	viper.SetDefault("gitlab.retry_count", DefaultRetryCount)
	viper.SetDefault("gitlab.rate_limit_rps", DefaultRateLimitRPS)

	// Default behavior
	viper.SetDefault("defaults.target_branch", "main")
	viper.SetDefault("defaults.branch_prefix", "update-tag")
	viper.SetDefault("defaults.merge_timeout", DefaultMergeTimeout)
	viper.SetDefault("defaults.wait_previous_mr", false)
	viper.SetDefault("defaults.auto_merge", false)

	// Performance defaults
	viper.SetDefault("performance.max_concurrent_requests", DefaultMaxConcurrentReqs)
	viper.SetDefault("performance.request_timeout", DefaultTimeout)
	viper.SetDefault("performance.buffer_size", DefaultBufferSize)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "text")
	viper.SetDefault("logging.enable_file", false)
	viper.SetDefault("logging.file_path", "go-tag-updater.log")
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
