// Package logger provides structured logging functionality for go-tag-updater.
package logger

import (
	"fmt"
	"log"
	"os"
)

const (
	// LevelDebug represents debug log level
	LevelDebug = "debug"
	// LevelInfo represents info log level
	LevelInfo = "info"
	// LevelWarn represents warn log level
	LevelWarn = "warn"
	// LevelError represents error log level
	LevelError = "error"

	// FormatText represents plain text log format
	FormatText = "text"
	// FormatJSON represents JSON log format
	FormatJSON = "json"

	// DefaultLogLevel is the default logging level
	DefaultLogLevel = LevelInfo
	// DefaultLogFormat is the default logging format
	DefaultLogFormat = FormatJSON
	// MaxLogFileSize defines the maximum size for log files
	MaxLogFileSize = 100 * 1024 * 1024 // 100MB
)

// Logger provides structured logging with debug capabilities
type Logger struct {
	debug      bool
	level      string
	format     string
	infoLogger *log.Logger
	errLogger  *log.Logger
}

// New creates a new logger instance
func New(debug bool) *Logger {
	return &Logger{
		debug:      debug,
		level:      getLogLevel(debug),
		format:     DefaultLogFormat,
		infoLogger: log.New(os.Stdout, "", log.LstdFlags),
		errLogger:  log.New(os.Stderr, "", log.LstdFlags),
	}
}

// NewWithConfig creates a new logger with specific configuration
func NewWithConfig(debug bool, level, format string) *Logger {
	return &Logger{
		debug:      debug,
		level:      level,
		format:     format,
		infoLogger: log.New(os.Stdout, "", log.LstdFlags),
		errLogger:  log.New(os.Stderr, "", log.LstdFlags),
	}
}

// Debug logs debug messages (only if debug mode is enabled)
func (l *Logger) Debug(message string) {
	if l.debug {
		l.log(LevelDebug, message)
	}
}

// Debugf logs formatted debug messages (only if debug mode is enabled)
func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.debug {
		l.log(LevelDebug, fmt.Sprintf(format, args...))
	}
}

// Info logs informational messages
func (l *Logger) Info(message string) {
	l.log(LevelInfo, message)
}

// Infof logs formatted informational messages
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(LevelInfo, fmt.Sprintf(format, args...))
}

// Warn logs warning messages
func (l *Logger) Warn(message string) {
	l.log(LevelWarn, message)
}

// Warnf logs formatted warning messages
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(LevelWarn, fmt.Sprintf(format, args...))
}

// Error logs error messages
func (l *Logger) Error(message string) {
	l.log(LevelError, message)
}

// Errorf logs formatted error messages
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(LevelError, fmt.Sprintf(format, args...))
}

// log performs the actual logging
func (l *Logger) log(level, message string) {
	prefix := getLogPrefix(level)
	logMessage := fmt.Sprintf("%s %s", prefix, message)

	switch level {
	case LevelError:
		l.errLogger.Print(logMessage)
	default:
		l.infoLogger.Print(logMessage)
	}
}

// IsDebugEnabled returns true if debug logging is enabled
func (l *Logger) IsDebugEnabled() bool {
	return l.debug
}

// getLogLevel returns the appropriate log level based on debug flag
func getLogLevel(debug bool) string {
	if debug {
		return LevelDebug
	}
	return DefaultLogLevel
}

// getLogPrefix returns the prefix for each log level
func getLogPrefix(level string) string {
	switch level {
	case LevelDebug:
		return "[DEBUG]"
	case LevelInfo:
		return "[INFO]"
	case LevelWarn:
		return "[WARN]"
	case LevelError:
		return "[ERROR]"
	default:
		return "[INFO]"
	}
}
