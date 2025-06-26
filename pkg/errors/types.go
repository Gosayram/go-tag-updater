// Package errors provides custom error types and utilities for go-tag-updater.
package errors

import (
	"fmt"
)

const (
	// ErrCodeInvalidProject indicates an invalid GitLab project identifier
	ErrCodeInvalidProject = 1001
	// ErrCodeFileNotFound indicates a file was not found in the repository
	ErrCodeFileNotFound = 1002
	// ErrCodeInvalidYAML indicates malformed YAML content
	ErrCodeInvalidYAML = 1003
	// ErrCodeGitOperation indicates a Git operation failure
	ErrCodeGitOperation = 1004
	// ErrCodeAPIError indicates a GitLab API error
	ErrCodeAPIError = 1005
	// ErrCodeMergeConflict indicates a merge conflict occurred
	ErrCodeMergeConflict = 1006
	// ErrCodeValidation indicates input validation failure
	ErrCodeValidation = 1007
	// ErrCodeConfiguration indicates configuration error
	ErrCodeConfiguration = 1008
	// ErrCodeNetworkError indicates network connectivity issues
	ErrCodeNetworkError = 1009
	// ErrCodeAuthError indicates authentication or authorization failure
	ErrCodeAuthError = 1010

	// MaxErrorMessageLength defines the maximum length for error messages
	MaxErrorMessageLength = 500
	// MaxErrorContextLength defines the maximum length for error context
	MaxErrorContextLength = 200
)

// Error categories
const (
	CategoryConfig     = "configuration"
	CategoryAPI        = "api"
	CategoryFile       = "file"
	CategoryFileSystem = "filesystem"
	CategoryGit        = "git"
	CategoryValidation = "validation"
	CategoryNetwork    = "network"
)

// AppError represents a structured application error
type AppError struct {
	Code     int    `json:"code"`
	Category string `json:"category"`
	Message  string `json:"message"`
	Context  string `json:"context,omitempty"`
	Cause    error  `json:"cause,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Category, e.Message, e.Context)
	}
	return fmt.Sprintf("[%s] %s", e.Category, e.Message)
}

// Unwrap returns the underlying cause error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError creates a new application error
func NewAppError(code int, category, message string) *AppError {
	return &AppError{
		Code:     code,
		Category: category,
		Message:  truncateString(message, MaxErrorMessageLength),
	}
}

// NewAppErrorWithContext creates a new application error with context
func NewAppErrorWithContext(code int, category, message, context string) *AppError {
	return &AppError{
		Code:     code,
		Category: category,
		Message:  truncateString(message, MaxErrorMessageLength),
		Context:  truncateString(context, MaxErrorContextLength),
	}
}

// NewAppErrorWithCause creates a new application error wrapping another error
func NewAppErrorWithCause(code int, category, message string, cause error) *AppError {
	return &AppError{
		Code:     code,
		Category: category,
		Message:  truncateString(message, MaxErrorMessageLength),
		Cause:    cause,
	}
}

// NewConfigError creates a new configuration error
func NewConfigError(message string) *AppError {
	return NewAppError(ErrCodeConfiguration, CategoryConfig, message)
}

// NewConfigErrorWithCause creates a new configuration error with an underlying cause
func NewConfigErrorWithCause(message string, cause error) *AppError {
	return NewAppErrorWithCause(ErrCodeConfiguration, CategoryConfig, message, cause)
}

// NewAPIError creates a new GitLab API error
func NewAPIError(message string) *AppError {
	return NewAppError(ErrCodeAPIError, CategoryAPI, message)
}

// NewAPIErrorWithContext creates a new API error with additional context
func NewAPIErrorWithContext(message, context string) *AppError {
	return NewAppErrorWithContext(ErrCodeAPIError, CategoryAPI, message, context)
}

// NewAuthError creates a new authentication/authorization error
func NewAuthError(message string) *AppError {
	return NewAppError(ErrCodeAuthError, CategoryAPI, message)
}

// NewFileNotFoundError creates a new file not found error
func NewFileNotFoundError(filePath string) *AppError {
	return NewAppErrorWithContext(ErrCodeFileNotFound, CategoryFile, "file not found", filePath)
}

// NewInvalidYAMLError creates a new invalid YAML format error
func NewInvalidYAMLError(message string) *AppError {
	return NewAppError(ErrCodeInvalidYAML, CategoryFile, message)
}

// NewGitError creates a new Git operation error
func NewGitError(message string) *AppError {
	return NewAppError(ErrCodeGitOperation, CategoryGit, message)
}

// NewGitOperationError creates a new Git operation error
func NewGitOperationError(message string) *AppError {
	return NewAppError(ErrCodeGitOperation, CategoryGit, message)
}

// NewGitErrorWithCause creates a new Git error with an underlying cause
func NewGitErrorWithCause(message string, cause error) *AppError {
	return NewAppErrorWithCause(ErrCodeGitOperation, CategoryGit, message, cause)
}

// NewFileSystemError creates a new filesystem operation error
func NewFileSystemError(message string) *AppError {
	return NewAppError(ErrCodeFileNotFound, CategoryFileSystem, message)
}

// NewMergeConflictError creates a new merge conflict error
func NewMergeConflictError(message string) *AppError {
	return NewAppError(ErrCodeMergeConflict, CategoryGit, message)
}

// NewValidationError creates a new input validation error
func NewValidationError(message string) *AppError {
	return NewAppError(ErrCodeValidation, CategoryValidation, message)
}

// NewValidationErrorWithContext creates a new validation error with additional context
func NewValidationErrorWithContext(message, context string) *AppError {
	return NewAppErrorWithContext(ErrCodeValidation, CategoryValidation, message, context)
}

// NewInvalidProjectError creates a new invalid project identifier error
func NewInvalidProjectError(projectID string) *AppError {
	return NewAppErrorWithContext(ErrCodeInvalidProject, CategoryValidation, "invalid project identifier", projectID)
}

// NewProjectNotFoundError creates a new project not found error
func NewProjectNotFoundError(message string) *AppError {
	return NewAppError(ErrCodeInvalidProject, CategoryValidation, message)
}

// NewNetworkError creates a new network connectivity error
func NewNetworkError(message string) *AppError {
	return NewAppError(ErrCodeNetworkError, CategoryNetwork, message)
}

// NewNetworkErrorWithCause creates a new network error with an underlying cause
func NewNetworkErrorWithCause(message string, cause error) *AppError {
	return NewAppErrorWithCause(ErrCodeNetworkError, CategoryNetwork, message, cause)
}

// Helper functions

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetErrorCode extracts error code from an error
func GetErrorCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return 0
}

// GetErrorCategory extracts error category from an error
func GetErrorCategory(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Category
	}
	return "unknown"
}

// truncateString truncates a string to the specified maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	const suffix = "..."
	if maxLen <= len(suffix) {
		return s[:maxLen]
	}
	return s[:maxLen-len(suffix)] + suffix
}
