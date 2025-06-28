// Package yaml provides YAML file parsing and manipulation utilities
package yaml

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Gosayram/go-tag-updater/pkg/errors"
)

const (
	// DefaultFilePermissions defines the default file permissions for created files
	DefaultFilePermissions = 0o644
	// TempFileSuffix defines the suffix for temporary files
	TempFileSuffix = ".tmp"
	// BackupTimestampFormat defines the timestamp format for backup files
	BackupTimestampFormat = "20060102_150405"
	// MaxBackupFiles defines the maximum number of backup files to keep
	MaxBackupFiles = 5
	// BackupDirPermissions defines the permissions for backup directories
	BackupDirPermissions = 0o750

	// PathTraversalPattern defines the pattern used to detect path traversal attacks
	PathTraversalPattern = ".."
	// WindowsDriveLetterSeparator defines the separator used in Windows drive letters
	WindowsDriveLetterSeparator = ":"
	// WindowsPathSeparator defines the Windows path separator
	WindowsPathSeparator = "\\"
	// UnixPathSeparator defines the Unix path separator
	UnixPathSeparator = "/"

	// UnixTempDir defines the Unix temporary directory prefix
	UnixTempDir = "/tmp/"
	// UnixVarTempDir defines the Unix variable temporary directory prefix
	UnixVarTempDir = "/var/tmp/"
	// WindowsCTempDir defines the Windows C: drive temporary directory prefix
	WindowsCTempDir = "C:/temp/"
	// WindowsCTmpDir defines the Windows C: drive tmp directory prefix
	WindowsCTmpDir = "C:/tmp/"
	// WindowsDTempDir defines the Windows D: drive temporary directory prefix
	WindowsDTempDir = "D:/temp/"
	// WindowsDTmpDir defines the Windows D: drive tmp directory prefix
	WindowsDTmpDir = "D:/tmp/"
)

// Updater handles YAML file updates with backup and rollback capabilities
type Updater struct {
	parser      *Parser
	backupDir   string
	keepBackups bool
	atomicWrite bool
}

// UpdateRequest contains all information needed for a tag update
type UpdateRequest struct {
	FilePath      string
	NewTagValue   string
	TagPath       []string
	CreateBackup  bool
	ValidateAfter bool
	DryRun        bool
}

// UpdateResult contains the result of an update operation
type UpdateResult struct {
	Success         bool
	UpdatedContent  string
	BackupPath      string
	OriginalContent string
	ValidationError error
	ChangesDetected bool
}

// NewUpdater creates a new YAML updater with default settings
func NewUpdater() *Updater {
	return &Updater{
		parser:      NewParser(),
		keepBackups: true,
		atomicWrite: true,
	}
}

// NewUpdaterWithOptions creates a new YAML updater with custom options
func NewUpdaterWithOptions(backupDir string, keepBackups, atomicWrite bool) *Updater {
	return &Updater{
		parser:      NewParser(),
		backupDir:   backupDir,
		keepBackups: keepBackups,
		atomicWrite: atomicWrite,
	}
}

// UpdateTagInFile updates a tag in a YAML file with comprehensive error handling
func (u *Updater) UpdateTagInFile(request *UpdateRequest) (*UpdateResult, error) {
	if request == nil {
		return nil, errors.NewValidationError("update request cannot be nil")
	}

	if request.FilePath == "" {
		return nil, errors.NewValidationError("file path cannot be empty")
	}

	if request.NewTagValue == "" {
		return nil, errors.NewValidationError("new tag value cannot be empty")
	}

	result := &UpdateResult{}

	// Read the original file
	originalContent, err := u.readFile(request.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", request.FilePath, err)
	}

	result.OriginalContent = originalContent

	// Parse the YAML content
	parseResult, err := u.parser.ParseContent(originalContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML file %s: %w", request.FilePath, err)
	}

	// Determine tag path if not provided
	tagPath := request.TagPath
	if len(tagPath) == 0 {
		tagPath, err = u.autoDetectTagPath(parseResult)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-detect tag path: %w", err)
		}
	}

	// Update the tag
	updateOptions := &UpdateOptions{
		TagPath:         tagPath,
		NewValue:        request.NewTagValue,
		CreateIfMissing: false, // For safety, don't create missing tags
	}

	updatedContent, err := u.parser.UpdateTag(parseResult, updateOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	result.UpdatedContent = updatedContent
	result.ChangesDetected = originalContent != updatedContent

	// Validate the updated content if requested
	if request.ValidateAfter {
		result.ValidationError = u.parser.ValidateYAML(updatedContent)
	}

	// Handle dry run
	if request.DryRun {
		result.Success = true
		return result, nil
	}

	// Create backup if requested
	if request.CreateBackup && u.keepBackups {
		var backupErr error
		result.BackupPath, backupErr = u.createBackup(request.FilePath, originalContent)
		if backupErr != nil {
			return nil, fmt.Errorf("failed to create backup: %w", backupErr)
		}
	}

	// Write the updated content
	err = u.writeFile(request.FilePath, updatedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to write updated file: %w", err)
	}

	result.Success = true
	return result, nil
}

// UpdateTagSimpleInFile provides a simple interface for common tag updates
func (u *Updater) UpdateTagSimpleInFile(filePath, newTagValue string, createBackup bool) (*UpdateResult, error) {
	request := &UpdateRequest{
		FilePath:      filePath,
		NewTagValue:   newTagValue,
		CreateBackup:  createBackup,
		ValidateAfter: true,
		DryRun:        false,
	}

	return u.UpdateTagInFile(request)
}

// PreviewUpdate shows what changes would be made without actually applying them
func (u *Updater) PreviewUpdate(request *UpdateRequest) (*UpdateResult, error) {
	// Create a copy of the request with dry run enabled
	previewRequest := *request
	previewRequest.DryRun = true
	previewRequest.CreateBackup = false

	return u.UpdateTagInFile(&previewRequest)
}

// RollbackFromBackup restores a file from its backup
func (u *Updater) RollbackFromBackup(filePath, backupPath string) error {
	if filePath == "" {
		return errors.NewValidationError("file path cannot be empty")
	}

	if backupPath == "" {
		return errors.NewValidationError("backup path cannot be empty")
	}

	// Verify backup file exists
	if !u.fileExists(backupPath) {
		return errors.NewFileNotFoundError(backupPath)
	}

	// Read backup content
	backupContent, err := u.readFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file %s: %w", backupPath, err)
	}

	// Validate backup content
	err = u.parser.ValidateYAML(backupContent)
	if err != nil {
		return fmt.Errorf("backup file contains invalid YAML: %w", err)
	}

	// Restore the file
	err = u.writeFile(filePath, backupContent)
	if err != nil {
		return fmt.Errorf("failed to restore file from backup: %w", err)
	}

	return nil
}

// CleanupOldBackups removes old backup files to save disk space
func (u *Updater) CleanupOldBackups(filePath string) error {
	if u.backupDir == "" {
		return nil // No backup directory configured
	}

	// List all backup files for this file
	baseFileName := filepath.Base(filePath)
	backupPattern := fmt.Sprintf("%s.*.backup", baseFileName)
	backupGlob := filepath.Join(u.backupDir, backupPattern)

	matches, err := filepath.Glob(backupGlob)
	if err != nil {
		return fmt.Errorf("failed to list backup files: %w", err)
	}

	// Sort by modification time and keep only recent ones
	if len(matches) > MaxBackupFiles {
		// This is a simplified cleanup - in a full implementation,
		// you would sort by file modification time and remove oldest files
		for i := MaxBackupFiles; i < len(matches); i++ {
			err := os.Remove(matches[i])
			if err != nil {
				// Log warning but continue
				continue
			}
		}
	}

	return nil
}

// autoDetectTagPath attempts to automatically detect the tag path in YAML
func (u *Updater) autoDetectTagPath(parseResult *ParseResult) ([]string, error) {
	if len(parseResult.TagLocations) == 0 {
		return nil, errors.NewValidationError("no tag fields found in YAML content")
	}

	// Prefer common tag field names
	preferredPaths := []string{"tag", "version", "image"}

	for _, preferred := range preferredPaths {
		for _, location := range parseResult.TagLocations {
			if len(location.Path) > 0 &&
				strings.ToLower(location.Path[len(location.Path)-1]) == preferred {
				return location.Path, nil
			}
		}
	}

	// If no preferred path found, return the first detected tag
	return parseResult.TagLocations[0].Path, nil
}

// createBackup creates a backup of the original file
func (u *Updater) createBackup(filePath, content string) (string, error) {
	timestamp := time.Now().Format(BackupTimestampFormat)
	baseFileName := filepath.Base(filePath)

	var backupPath string
	if u.backupDir != "" {
		// Ensure backup directory exists
		err := os.MkdirAll(u.backupDir, BackupDirPermissions)
		if err != nil {
			return "", fmt.Errorf("failed to create backup directory: %w", err)
		}
		backupPath = filepath.Join(u.backupDir, fmt.Sprintf("%s.%s.backup", baseFileName, timestamp))
	} else {
		// Create backup in the same directory as the original file
		backupPath = fmt.Sprintf("%s.%s.backup", filePath, timestamp)
	}

	err := u.writeFile(backupPath, content)
	if err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	return backupPath, nil
}

// validateAndCleanFilePath validates that the file path is safe and returns a cleaned version
func (u *Updater) validateAndCleanFilePath(filePath string) (string, error) {
	if filePath == "" {
		return "", errors.NewValidationError("file path cannot be empty")
	}

	// Clean the path to resolve any .. or . elements
	cleanPath := filepath.Clean(filePath)

	// Check for suspicious patterns that could indicate path traversal
	if strings.Contains(cleanPath, PathTraversalPattern) {
		return "", errors.NewValidationError("file path contains invalid traversal patterns")
	}

	// Additional security checks for absolute paths (cross-platform)
	if filepath.IsAbs(cleanPath) {
		// Only allow absolute paths in specific safe directories
		// For Unix-like systems: /tmp/, /var/tmp/
		// For Windows: temp directories or explicitly allowed paths
		allowedPrefixes := []string{
			UnixTempDir,
			UnixVarTempDir,
		}

		// Add Windows temp directory patterns
		if strings.Contains(cleanPath, WindowsDriveLetterSeparator) { // Windows drive letter
			// Convert to forward slashes for consistent checking
			normalizedPath := strings.ReplaceAll(cleanPath, WindowsPathSeparator, UnixPathSeparator)
			allowedPrefixes = append(allowedPrefixes,
				WindowsCTempDir, WindowsCTmpDir,
				WindowsDTempDir, WindowsDTmpDir,
			)
			cleanPath = normalizedPath
		}

		allowed := false
		for _, prefix := range allowedPrefixes {
			if strings.HasPrefix(strings.ToLower(cleanPath), strings.ToLower(prefix)) {
				allowed = true
				break
			}
		}

		if !allowed {
			return "", errors.NewValidationError("absolute file paths outside safe directories are not allowed")
		}
	}

	return cleanPath, nil
}

// readFile reads content from a file with error handling and path validation
func (u *Updater) readFile(filePath string) (string, error) {
	// Validate and clean file path for security
	cleanPath, err := u.validateAndCleanFilePath(filePath)
	if err != nil {
		return "", err
	}

	if !u.fileExists(cleanPath) {
		return "", errors.NewFileNotFoundError(cleanPath)
	}

	content, err := os.ReadFile(cleanPath) // #nosec G304 -- path validated and cleaned above
	if err != nil {
		return "", errors.NewFileSystemError(fmt.Sprintf("failed to read file %s: %v", cleanPath, err))
	}

	return string(content), nil
}

// writeFile writes content to a file with atomic operations if enabled
func (u *Updater) writeFile(filePath, content string) error {
	// Validate and clean file path for security
	cleanPath, err := u.validateAndCleanFilePath(filePath)
	if err != nil {
		return err
	}

	if u.atomicWrite {
		return u.writeFileAtomic(cleanPath, content)
	}

	// #nosec G304 -- path validated and cleaned above
	err = os.WriteFile(cleanPath, []byte(content), DefaultFilePermissions)
	if err != nil {
		return errors.NewFileSystemError(fmt.Sprintf("failed to write file %s: %v", cleanPath, err))
	}

	return nil
}

// writeFileAtomic writes to a temporary file first, then renames it
func (u *Updater) writeFileAtomic(filePath, content string) error {
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	tempPath := filepath.Join(dir, base+TempFileSuffix)

	// Write to temporary file
	// #nosec G304 -- tempPath constructed safely
	err := os.WriteFile(tempPath, []byte(content), DefaultFilePermissions)
	if err != nil {
		return errors.NewFileSystemError(fmt.Sprintf("failed to write temporary file: %v", err))
	}

	// Atomically rename temporary file to target
	err = os.Rename(tempPath, filePath)
	if err != nil {
		// Clean up temporary file on failure
		// Clean up temporary file on failure, ignoring cleanup errors
		_ = os.Remove(tempPath)
		return errors.NewFileSystemError(fmt.Sprintf("failed to rename temporary file: %v", err))
	}

	return nil
}

// fileExists checks if a file exists and is readable
func (u *Updater) fileExists(filePath string) bool {
	// Validate and clean file path for security - if invalid, consider file as non-existent
	cleanPath, err := u.validateAndCleanFilePath(filePath)
	if err != nil {
		return false
	}

	_, err = os.Stat(cleanPath)
	return err == nil
}

// ValidateFile validates that a YAML file is syntactically correct
func (u *Updater) ValidateFile(filePath string) error {
	content, err := u.readFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file for validation: %w", err)
	}

	return u.parser.ValidateYAML(content)
}

// GetFileTagLocations returns all tag locations found in a YAML file
func (u *Updater) GetFileTagLocations(filePath string) ([]TagLocation, error) {
	content, err := u.readFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	parseResult, err := u.parser.ParseContent(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return parseResult.TagLocations, nil
}
