// Package logger provides structured logging functionality for go-tag-updater.
package logger

import (
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// LevelTrace represents trace log level
	LevelTrace = "trace"
	// LevelDebug represents debug log level
	LevelDebug = "debug"
	// LevelInfo represents info log level
	LevelInfo = "info"
	// LevelWarn represents warn log level
	LevelWarn = "warn"
	// LevelError represents error log level
	LevelError = "error"
	// LevelFatal represents fatal log level
	LevelFatal = "fatal"
	// LevelPanic represents panic log level
	LevelPanic = "panic"

	// FormatJSON represents JSON log format
	FormatJSON = "json"
	// FormatText represents text log format
	FormatText = "text"

	// DefaultLogLevel is the default logging level
	DefaultLogLevel = LevelInfo
	// DefaultLogFormat is the default log format
	DefaultLogFormat = FormatJSON
	// DefaultReportCaller determines if caller information should be reported
	DefaultReportCaller = false
	// DefaultTimestamp determines if timestamps should be included
	DefaultTimestamp = true

	// MaxLogFileSize is the maximum size of a log file before rotation
	MaxLogFileSize = 100 * 1024 * 1024 // 100MB
	// DefaultLogFilePerm is the default file permissions for log files
	DefaultLogFilePerm = 0o644
	// LogFileBufferSize is the buffer size for log file operations
	LogFileBufferSize = 4096
	// MaxLogRotationCount is the maximum number of rotated log files to keep
	MaxLogRotationCount = 5
	// LogRotationAge is the maximum age of log files before rotation
	LogRotationAge = 24 * time.Hour
	// DefaultLogDirectory is the default directory for log files
	DefaultLogDirectory = "logs"

	// MaxFieldsPerLog is the maximum number of fields allowed per log entry
	MaxFieldsPerLog = 50
	// MaxMessageLength is the maximum length of a log message
	MaxMessageLength = 1024
	// DefaultFlushTimeout is the default timeout for flushing log buffers
	DefaultFlushTimeout = 5 * time.Second

	// FieldComponent is the field name for component identification
	FieldComponent = "component"
	// FieldOperation is the field name for operation identification
	FieldOperation = "operation"
	// FieldDuration is the field name for duration measurements
	FieldDuration = "duration"
	// FieldError is the field name for error information
	FieldError = "error"
	// FieldProjectID is the field name for project ID
	FieldProjectID = "project_id"
	// FieldFilePath is the field name for file path information
	FieldFilePath = "file_path"
	// FieldBranch is the field name for branch information
	FieldBranch = "branch"
	// FieldMRID is the field name for merge request ID
	FieldMRID = "mr_id"
	// FieldUserAgent is the field name for user agent information
	FieldUserAgent = "user_agent"
	// FieldRequestID is the field name for request ID tracking
	FieldRequestID = "request_id"
)

// Logger provides structured logging with Logrus backend
type Logger struct {
	logrus       *logrus.Logger
	debug        bool
	level        string
	format       string
	component    string
	reportCaller bool
}

// Config holds logger configuration
type Config struct {
	Level        string
	Format       string
	Debug        bool
	ReportCaller bool
	Output       io.Writer
	Component    string
}

// New creates a new logger instance with debug mode setting
func New(debug bool) *Logger {
	config := &Config{
		Debug:        debug,
		Level:        getLogLevel(debug),
		Format:       DefaultLogFormat,
		ReportCaller: DefaultReportCaller,
		Output:       os.Stdout,
		Component:    "go-tag-updater",
	}
	return NewWithConfig(config)
}

// NewWithConfig creates a new logger with specific configuration
func NewWithConfig(config *Config) *Logger {
	if config == nil {
		config = &Config{
			Level:        DefaultLogLevel,
			Format:       DefaultLogFormat,
			Debug:        false,
			ReportCaller: DefaultReportCaller,
			Output:       os.Stdout,
			Component:    "go-tag-updater",
		}
	}

	logger := logrus.New()

	// Set output
	if config.Output != nil {
		logger.SetOutput(config.Output)
	}

	// Set log level
	level := parseLogLevel(config.Level)
	logger.SetLevel(level)

	// Set formatter
	switch config.Format {
	case FormatJSON:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	case FormatText:
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   DefaultTimestamp,
			TimestampFormat: time.RFC3339,
			DisableColors:   false,
			ForceColors:     false,
		})
	default:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	}

	// Set caller reporting
	logger.SetReportCaller(config.ReportCaller)

	return &Logger{
		logrus:       logger,
		debug:        config.Debug,
		level:        config.Level,
		format:       config.Format,
		component:    config.Component,
		reportCaller: config.ReportCaller,
	}
}

// WithComponent creates a new logger with a specific component name
func (l *Logger) WithComponent(component string) *Logger {
	newLogger := &Logger{
		logrus:       l.logrus,
		debug:        l.debug,
		level:        l.level,
		format:       l.format,
		component:    component,
		reportCaller: l.reportCaller,
	}
	return newLogger
}

// WithFields creates a new logger entry with structured fields
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	// Add component field if set
	if l.component != "" {
		fields[FieldComponent] = l.component
	}
	return l.logrus.WithFields(fields)
}

// WithField creates a new logger entry with a single field
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	fields := logrus.Fields{key: value}
	return l.WithFields(fields)
}

// WithError creates a new logger entry with error field
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.WithField(FieldError, err)
}

// WithOperation creates a new logger entry with operation field
func (l *Logger) WithOperation(operation string) *logrus.Entry {
	return l.WithField(FieldOperation, operation)
}

// WithDuration creates a new logger entry with duration field
func (l *Logger) WithDuration(duration time.Duration) *logrus.Entry {
	return l.WithField(FieldDuration, duration)
}

// WithProjectID creates a new logger entry with project ID field
func (l *Logger) WithProjectID(projectID interface{}) *logrus.Entry {
	return l.WithField(FieldProjectID, projectID)
}

// WithFilePath creates a new logger entry with file path field
func (l *Logger) WithFilePath(filePath string) *logrus.Entry {
	return l.WithField(FieldFilePath, filePath)
}

// WithBranch creates a new logger entry with branch field
func (l *Logger) WithBranch(branch string) *logrus.Entry {
	return l.WithField(FieldBranch, branch)
}

// WithMRID creates a new logger entry with merge request ID field
func (l *Logger) WithMRID(mrID int) *logrus.Entry {
	return l.WithField(FieldMRID, mrID)
}

// Trace logs trace messages (only if trace level is enabled)
func (l *Logger) Trace(message string) {
	l.getBaseEntry().Trace(message)
}

// Tracef logs formatted trace messages
func (l *Logger) Tracef(format string, args ...interface{}) {
	l.getBaseEntry().Tracef(format, args...)
}

// Debug logs debug messages (only if debug level is enabled)
func (l *Logger) Debug(message string) {
	l.getBaseEntry().Debug(message)
}

// Debugf logs formatted debug messages
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.getBaseEntry().Debugf(format, args...)
}

// Info logs informational messages
func (l *Logger) Info(message string) {
	l.getBaseEntry().Info(message)
}

// Infof logs formatted informational messages
func (l *Logger) Infof(format string, args ...interface{}) {
	l.getBaseEntry().Infof(format, args...)
}

// Warn logs warning messages
func (l *Logger) Warn(message string) {
	l.getBaseEntry().Warn(message)
}

// Warnf logs formatted warning messages
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.getBaseEntry().Warnf(format, args...)
}

// Error logs error messages
func (l *Logger) Error(message string) {
	l.getBaseEntry().Error(message)
}

// Errorf logs formatted error messages
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.getBaseEntry().Errorf(format, args...)
}

// Fatal logs fatal messages and exits the program
func (l *Logger) Fatal(message string) {
	l.getBaseEntry().Fatal(message)
}

// Fatalf logs formatted fatal messages and exits the program
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.getBaseEntry().Fatalf(format, args...)
}

// Panic logs panic messages and panics
func (l *Logger) Panic(message string) {
	l.getBaseEntry().Panic(message)
}

// Panicf logs formatted panic messages and panics
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.getBaseEntry().Panicf(format, args...)
}

// IsDebugEnabled returns true if debug logging is enabled
func (l *Logger) IsDebugEnabled() bool {
	return l.debug || l.logrus.IsLevelEnabled(logrus.DebugLevel)
}

// IsLevelEnabled returns true if the given level is enabled
func (l *Logger) IsLevelEnabled(level string) bool {
	logrusLevel := parseLogLevel(level)
	return l.logrus.IsLevelEnabled(logrusLevel)
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() string {
	return l.level
}

// GetFormat returns the current log format
func (l *Logger) GetFormat() string {
	return l.format
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level string) {
	l.level = level
	logrusLevel := parseLogLevel(level)
	l.logrus.SetLevel(logrusLevel)
}

// SetFormat sets the log format
func (l *Logger) SetFormat(format string) {
	l.format = format
	switch format {
	case FormatJSON:
		l.logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	case FormatText:
		l.logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   DefaultTimestamp,
			TimestampFormat: time.RFC3339,
		})
	}
}

// SetOutput sets the log output destination
func (l *Logger) SetOutput(output io.Writer) {
	l.logrus.SetOutput(output)
}

// GetLogrus returns the underlying logrus logger
func (l *Logger) GetLogrus() *logrus.Logger {
	return l.logrus
}

// getBaseEntry returns a base entry with component field if set
func (l *Logger) getBaseEntry() *logrus.Entry {
	if l.component != "" {
		return l.logrus.WithField(FieldComponent, l.component)
	}
	return l.logrus.WithFields(logrus.Fields{})
}

// getLogLevel returns the appropriate log level based on debug flag
func getLogLevel(debug bool) string {
	if debug {
		return LevelDebug
	}
	return DefaultLogLevel
}

// parseLogLevel converts string level to logrus level
func parseLogLevel(level string) logrus.Level {
	switch level {
	case LevelTrace:
		return logrus.TraceLevel
	case LevelDebug:
		return logrus.DebugLevel
	case LevelInfo:
		return logrus.InfoLevel
	case LevelWarn:
		return logrus.WarnLevel
	case LevelError:
		return logrus.ErrorLevel
	case LevelFatal:
		return logrus.FatalLevel
	case LevelPanic:
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}
