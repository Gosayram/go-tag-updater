// Package logger provides structured logging functionality for go-tag-updater.
package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// Test constants
	TestTimeout       = 5 * time.Second
	TestComponent     = "test-component"
	TestOperation     = "test-operation"
	TestProjectID     = 12345
	TestFilePath      = "/path/to/test/file.yaml"
	TestBranch        = "test-branch"
	TestMRID          = 42
	TestMessage       = "test message"
	TestFormatMessage = "test format %s"
	TestFormatArg     = "argument"
	TestErrorMessage  = "test error"
	TestDuration      = 100 * time.Millisecond
	TestUserAgent     = "go-tag-updater/1.0.0"
	TestRequestID     = "req-123456"

	// Buffer size for log output capture
	LogBufferSize = 1024

	// Expected field counts
	MinExpectedFields = 2
	MaxExpectedFields = 10
)

// TestNew tests the New function with debug mode
func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		debug         bool
		expectedLevel string
		description   string
	}{
		{
			name:          "debug enabled",
			debug:         true,
			expectedLevel: LevelDebug,
			description:   "should create logger with debug level when debug is true",
		},
		{
			name:          "debug disabled",
			debug:         false,
			expectedLevel: DefaultLogLevel,
			description:   "should create logger with default level when debug is false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.debug)

			if logger == nil {
				t.Errorf("New() returned nil logger")
				return
			}

			if logger.debug != tt.debug {
				t.Errorf("New() debug = %v, want %v", logger.debug, tt.debug)
			}

			if logger.level != tt.expectedLevel {
				t.Errorf("New() level = %v, want %v", logger.level, tt.expectedLevel)
			}

			if logger.format != DefaultLogFormat {
				t.Errorf("New() format = %v, want %v", logger.format, DefaultLogFormat)
			}

			if logger.logrus == nil {
				t.Errorf("New() logrus instance is nil")
			}
		})
	}
}

// TestNewWithConfig tests the NewWithConfig function
func TestNewWithConfig(t *testing.T) {
	var buf bytes.Buffer

	tests := []struct {
		name        string
		config      *Config
		expectError bool
		description string
	}{
		{
			name: "valid config",
			config: &Config{
				Level:        LevelInfo,
				Format:       FormatJSON,
				Debug:        false,
				ReportCaller: false,
				Output:       &buf,
				Component:    TestComponent,
			},
			expectError: false,
			description: "should create logger with valid config",
		},
		{
			name:        "nil config",
			config:      nil,
			expectError: false,
			description: "should create logger with default config when config is nil",
		},
		{
			name: "text format config",
			config: &Config{
				Level:        LevelWarn,
				Format:       FormatText,
				Debug:        true,
				ReportCaller: true,
				Output:       &buf,
				Component:    TestComponent,
			},
			expectError: false,
			description: "should create logger with text format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewWithConfig(tt.config)

			if logger == nil {
				t.Errorf("NewWithConfig() returned nil logger")
				return
			}

			if logger.logrus == nil {
				t.Errorf("NewWithConfig() logrus instance is nil")
			}

			// Test expected config values
			if tt.config != nil {
				if logger.level != tt.config.Level {
					t.Errorf("NewWithConfig() level = %v, want %v", logger.level, tt.config.Level)
				}
				if logger.format != tt.config.Format {
					t.Errorf("NewWithConfig() format = %v, want %v", logger.format, tt.config.Format)
				}
				if logger.component != tt.config.Component {
					t.Errorf("NewWithConfig() component = %v, want %v", logger.component, tt.config.Component)
				}
			}
		})
	}
}

// TestWithComponent tests the WithComponent method
func TestWithComponent(t *testing.T) {
	logger := New(false)
	componentLogger := logger.WithComponent(TestComponent)

	if componentLogger == nil {
		t.Errorf("WithComponent() returned nil logger")
		return
	}

	if componentLogger.component != TestComponent {
		t.Errorf("WithComponent() component = %v, want %v", componentLogger.component, TestComponent)
	}

	// Should be a new instance
	if componentLogger == logger {
		t.Errorf("WithComponent() returned same instance instead of new one")
	}
}

// TestStructuredLogging tests structured logging methods
func TestStructuredLogging(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:     LevelDebug,
		Format:    FormatJSON,
		Debug:     true,
		Output:    &buf,
		Component: TestComponent,
	}
	logger := NewWithConfig(config)

	tests := []struct {
		name           string
		logFunc        func()
		expectedFields map[string]interface{}
		description    string
	}{
		{
			name: "WithFields logging",
			logFunc: func() {
				logger.WithFields(logrus.Fields{
					"project_id": TestProjectID,
					"file_path":  TestFilePath,
				}).Info(TestMessage)
			},
			expectedFields: map[string]interface{}{
				"project_id": float64(TestProjectID), // JSON numbers are float64
				"file_path":  TestFilePath,
				"component":  TestComponent,
			},
			description: "should log with structured fields",
		},
		{
			name: "WithProjectID logging",
			logFunc: func() {
				logger.WithProjectID(TestProjectID).Info(TestMessage)
			},
			expectedFields: map[string]interface{}{
				"project_id": float64(TestProjectID),
				"component":  TestComponent,
			},
			description: "should log with project ID field",
		},
		{
			name: "WithOperation logging",
			logFunc: func() {
				logger.WithOperation(TestOperation).Info(TestMessage)
			},
			expectedFields: map[string]interface{}{
				"operation": TestOperation,
				"component": TestComponent,
			},
			description: "should log with operation field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()

			output := buf.String()
			if output == "" {
				t.Errorf("No log output captured")
				return
			}

			// Parse JSON log entry
			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
				t.Errorf("Failed to parse JSON log entry: %v", err)
				return
			}

			// Check expected fields
			for key, expectedValue := range tt.expectedFields {
				actualValue, exists := logEntry[key]
				if !exists {
					t.Errorf("Expected field %s not found in log entry", key)
					continue
				}
				if actualValue != expectedValue {
					t.Errorf("Field %s = %v, want %v", key, actualValue, expectedValue)
				}
			}

			// Check message
			if message, exists := logEntry["message"]; !exists || message != TestMessage {
				t.Errorf("Message = %v, want %v", message, TestMessage)
			}
		})
	}
}

// TestLogLevels tests all log levels
func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  LevelTrace,
		Format: FormatJSON,
		Debug:  true,
		Output: &buf,
	}
	logger := NewWithConfig(config)

	tests := []struct {
		name          string
		logFunc       func()
		expectedLevel string
		description   string
	}{
		{
			name: "trace level",
			logFunc: func() {
				logger.Trace(TestMessage)
			},
			expectedLevel: "trace",
			description:   "should log trace messages",
		},
		{
			name: "debug level",
			logFunc: func() {
				logger.Debug(TestMessage)
			},
			expectedLevel: "debug",
			description:   "should log debug messages",
		},
		{
			name: "info level",
			logFunc: func() {
				logger.Info(TestMessage)
			},
			expectedLevel: "info",
			description:   "should log info messages",
		},
		{
			name: "warn level",
			logFunc: func() {
				logger.Warn(TestMessage)
			},
			expectedLevel: "warning",
			description:   "should log warning messages",
		},
		{
			name: "error level",
			logFunc: func() {
				logger.Error(TestMessage)
			},
			expectedLevel: "error",
			description:   "should log error messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()

			output := buf.String()
			if output == "" {
				t.Errorf("No log output captured for level %s", tt.expectedLevel)
				return
			}

			// Parse JSON log entry
			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
				t.Errorf("Failed to parse JSON log entry: %v", err)
				return
			}

			// Check level
			if level, exists := logEntry["level"]; !exists || level != tt.expectedLevel {
				t.Errorf("Level = %v, want %v", level, tt.expectedLevel)
			}

			// Check message
			if message, exists := logEntry["message"]; !exists || message != TestMessage {
				t.Errorf("Message = %v, want %v", message, TestMessage)
			}
		})
	}
}

// TestFormattedLogging tests formatted logging methods
func TestFormattedLogging(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  LevelDebug,
		Format: FormatJSON,
		Debug:  true,
		Output: &buf,
	}
	logger := NewWithConfig(config)

	expectedMessage := "test format argument"

	tests := []struct {
		name        string
		logFunc     func()
		description string
	}{
		{
			name: "debugf",
			logFunc: func() {
				logger.Debugf(TestFormatMessage, TestFormatArg)
			},
			description: "should log formatted debug messages",
		},
		{
			name: "infof",
			logFunc: func() {
				logger.Infof(TestFormatMessage, TestFormatArg)
			},
			description: "should log formatted info messages",
		},
		{
			name: "warnf",
			logFunc: func() {
				logger.Warnf(TestFormatMessage, TestFormatArg)
			},
			description: "should log formatted warning messages",
		},
		{
			name: "errorf",
			logFunc: func() {
				logger.Errorf(TestFormatMessage, TestFormatArg)
			},
			description: "should log formatted error messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()

			output := buf.String()
			if output == "" {
				t.Errorf("No log output captured")
				return
			}

			// Parse JSON log entry
			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
				t.Errorf("Failed to parse JSON log entry: %v", err)
				return
			}

			// Check formatted message
			if message, exists := logEntry["message"]; !exists || message != expectedMessage {
				t.Errorf("Formatted message = %v, want %v", message, expectedMessage)
			}
		})
	}
}

// TestIsDebugEnabled tests debug detection
func TestIsDebugEnabled(t *testing.T) {
	tests := []struct {
		name        string
		debug       bool
		level       string
		expected    bool
		description string
	}{
		{
			name:        "debug flag enabled",
			debug:       true,
			level:       LevelInfo,
			expected:    true,
			description: "should return true when debug flag is enabled",
		},
		{
			name:        "debug level enabled",
			debug:       false,
			level:       LevelDebug,
			expected:    true,
			description: "should return true when debug level is set",
		},
		{
			name:        "debug disabled",
			debug:       false,
			level:       LevelInfo,
			expected:    false,
			description: "should return false when debug is disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Debug:  tt.debug,
				Level:  tt.level,
				Format: FormatJSON,
				Output: &bytes.Buffer{},
			}
			logger := NewWithConfig(config)

			result := logger.IsDebugEnabled()
			if result != tt.expected {
				t.Errorf("IsDebugEnabled() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsLevelEnabled tests level detection
func TestIsLevelEnabled(t *testing.T) {
	logger := NewWithConfig(&Config{
		Level:  LevelWarn,
		Format: FormatJSON,
		Output: &bytes.Buffer{},
	})

	tests := []struct {
		name        string
		testLevel   string
		expected    bool
		description string
	}{
		{
			name:        "trace level disabled",
			testLevel:   LevelTrace,
			expected:    false,
			description: "should return false for trace when warn level is set",
		},
		{
			name:        "debug level disabled",
			testLevel:   LevelDebug,
			expected:    false,
			description: "should return false for debug when warn level is set",
		},
		{
			name:        "info level disabled",
			testLevel:   LevelInfo,
			expected:    false,
			description: "should return false for info when warn level is set",
		},
		{
			name:        "warn level enabled",
			testLevel:   LevelWarn,
			expected:    true,
			description: "should return true for warn when warn level is set",
		},
		{
			name:        "error level enabled",
			testLevel:   LevelError,
			expected:    true,
			description: "should return true for error when warn level is set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logger.IsLevelEnabled(tt.testLevel)
			if result != tt.expected {
				t.Errorf("IsLevelEnabled(%s) = %v, want %v", tt.testLevel, result, tt.expected)
			}
		})
	}
}

// TestSetLevel tests level setting
func TestSetLevel(t *testing.T) {
	logger := New(false)

	tests := []struct {
		name        string
		level       string
		description string
	}{
		{
			name:        "set debug level",
			level:       LevelDebug,
			description: "should set debug level",
		},
		{
			name:        "set error level",
			level:       LevelError,
			description: "should set error level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.SetLevel(tt.level)

			if logger.GetLevel() != tt.level {
				t.Errorf("SetLevel() level = %v, want %v", logger.GetLevel(), tt.level)
			}

			// Test that the level is actually effective
			if !logger.IsLevelEnabled(tt.level) {
				t.Errorf("SetLevel() level %s is not enabled after setting", tt.level)
			}
		})
	}
}

// TestSetFormat tests format setting
func TestSetFormat(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: &buf,
	}
	logger := NewWithConfig(config)

	// Test JSON format
	logger.SetFormat(FormatJSON)
	buf.Reset()
	logger.Info(TestMessage)

	jsonOutput := buf.String()
	if jsonOutput == "" {
		t.Errorf("No JSON output captured")
	} else {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(jsonOutput), &logEntry); err != nil {
			t.Errorf("Failed to parse JSON output: %v", err)
		}
	}

	// Test text format
	logger.SetFormat(FormatText)
	buf.Reset()
	logger.Info(TestMessage)

	textOutput := buf.String()
	if textOutput == "" {
		t.Errorf("No text output captured")
	} else if !strings.Contains(textOutput, TestMessage) {
		t.Errorf("Text output does not contain expected message: %s", textOutput)
	}
}

// TestParseLogLevel tests log level parsing
func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		expected    logrus.Level
		description string
	}{
		{
			name:        "trace level",
			level:       LevelTrace,
			expected:    logrus.TraceLevel,
			description: "should parse trace level",
		},
		{
			name:        "debug level",
			level:       LevelDebug,
			expected:    logrus.DebugLevel,
			description: "should parse debug level",
		},
		{
			name:        "info level",
			level:       LevelInfo,
			expected:    logrus.InfoLevel,
			description: "should parse info level",
		},
		{
			name:        "warn level",
			level:       LevelWarn,
			expected:    logrus.WarnLevel,
			description: "should parse warn level",
		},
		{
			name:        "error level",
			level:       LevelError,
			expected:    logrus.ErrorLevel,
			description: "should parse error level",
		},
		{
			name:        "fatal level",
			level:       LevelFatal,
			expected:    logrus.FatalLevel,
			description: "should parse fatal level",
		},
		{
			name:        "panic level",
			level:       LevelPanic,
			expected:    logrus.PanicLevel,
			description: "should parse panic level",
		},
		{
			name:        "invalid level",
			level:       "invalid",
			expected:    logrus.InfoLevel,
			description: "should default to info level for invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLogLevel(tt.level)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%s) = %v, want %v", tt.level, result, tt.expected)
			}
		})
	}
}

// TestConstants validates all constants are properly defined
func TestConstants(t *testing.T) {
	tests := []struct {
		name          string
		value         interface{}
		expectNonZero bool
		description   string
	}{
		{
			name:          "MaxLogFileSize",
			value:         MaxLogFileSize,
			expectNonZero: true,
			description:   "should have non-zero max log file size",
		},
		{
			name:          "DefaultLogFilePerm",
			value:         DefaultLogFilePerm,
			expectNonZero: true,
			description:   "should have non-zero default file permissions",
		},
		{
			name:          "LogFileBufferSize",
			value:         LogFileBufferSize,
			expectNonZero: true,
			description:   "should have non-zero buffer size",
		},
		{
			name:          "MaxLogRotationCount",
			value:         MaxLogRotationCount,
			expectNonZero: true,
			description:   "should have non-zero rotation count",
		},
		{
			name:          "LogRotationAge",
			value:         LogRotationAge,
			expectNonZero: true,
			description:   "should have non-zero rotation age",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.value.(type) {
			case int:
				if tt.expectNonZero && v == 0 {
					t.Errorf("Constant %s = %d, expected non-zero", tt.name, v)
				}
			case time.Duration:
				if tt.expectNonZero && v == 0 {
					t.Errorf("Constant %s = %v, expected non-zero", tt.name, v)
				}
			default:
				t.Errorf("Unsupported constant type for %s", tt.name)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkNewLogger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		logger := New(false)
		if logger == nil {
			b.Fatal("New() returned nil")
		}
	}
}

func BenchmarkStructuredLogging(b *testing.B) {
	var buf bytes.Buffer
	config := &Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: &buf,
	}
	logger := NewWithConfig(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(logrus.Fields{
			"project_id": TestProjectID,
			"file_path":  TestFilePath,
			"operation":  TestOperation,
		}).Info(TestMessage)
	}
}

func BenchmarkSimpleLogging(b *testing.B) {
	var buf bytes.Buffer
	config := &Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: &buf,
	}
	logger := NewWithConfig(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(TestMessage)
	}
}

func BenchmarkFormattedLogging(b *testing.B) {
	var buf bytes.Buffer
	config := &Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: &buf,
	}
	logger := NewWithConfig(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Infof(TestFormatMessage, TestFormatArg)
	}
}

func BenchmarkLogLevelCheck(b *testing.B) {
	logger := New(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.IsLevelEnabled(LevelDebug)
	}
}
