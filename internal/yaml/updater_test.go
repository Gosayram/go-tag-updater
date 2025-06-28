package yaml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	// Test constants as per IDEA.md requirements
	TestTimeout         = 5000 // milliseconds
	MaxTestRetries      = 3
	TestSecureDirectory = "/tmp/go-tag-updater-test"
	TestYAMLContent     = `
name: test-app
version: 1.0.0
image:
  tag: old-tag
  repository: test/repo
`
	TestMaliciousPath1 = "../../../etc/passwd"
	TestMaliciousPath2 = "/etc/passwd"
	TestMaliciousPath3 = "../../sensitive.txt"
	TestValidPath      = "test-file.yaml"
	TestNewTag         = "new-tag"
	TestOldTag         = "old-tag"
)

func TestUpdater_validateAndCleanFilePath(t *testing.T) {
	updater := NewUpdater()

	tests := []struct {
		name        string
		filePath    string
		expectError bool
		description string
	}{
		{
			name:        "valid relative path",
			filePath:    TestValidPath,
			expectError: false,
			description: "should accept normal relative paths",
		},
		{
			name:        "path traversal attack 1",
			filePath:    TestMaliciousPath1,
			expectError: true,
			description: "should reject path traversal with ../",
		},
		{
			name:        "path traversal attack 2",
			filePath:    TestMaliciousPath3,
			expectError: true,
			description: "should reject multiple path traversal attempts",
		},
		{
			name:        "absolute path to system file",
			filePath:    TestMaliciousPath2,
			expectError: true,
			description: "should reject absolute paths to system files",
		},
		{
			name:        "empty path",
			filePath:    "",
			expectError: true,
			description: "should reject empty paths",
		},
		{
			name:        "valid nested path",
			filePath:    "configs/app/deployment.yaml",
			expectError: false,
			description: "should accept valid nested relative paths",
		},
		{
			name:        "valid tmp path",
			filePath:    "/tmp/test.yaml",
			expectError: false,
			description: "should accept /tmp paths",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := updater.validateAndCleanFilePath(tt.filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("validateAndCleanFilePath(%q) expected error but got none", tt.filePath)
				}
				if result != "" {
					t.Errorf("validateAndCleanFilePath(%q) expected empty result on error, got %q", tt.filePath, result)
				}
			} else {
				if err != nil {
					t.Errorf("validateAndCleanFilePath(%q) unexpected error: %v", tt.filePath, err)
				}
				if result == "" {
					t.Errorf("validateAndCleanFilePath(%q) expected non-empty result", tt.filePath)
				}
			}
		})
	}
}

func TestUpdater_NewUpdater(t *testing.T) {
	updater := NewUpdater()

	if updater == nil {
		t.Error("NewUpdater() should return non-nil updater")
		return
	}

	if updater.parser == nil {
		t.Error("NewUpdater() should initialize parser")
	}
}

func TestUpdater_NewUpdaterWithOptions(t *testing.T) {
	backupDir := "/tmp/backups"
	keepBackups := true
	atomicWrite := false

	updater := NewUpdaterWithOptions(backupDir, keepBackups, atomicWrite)

	if updater == nil {
		t.Error("NewUpdaterWithOptions() should return non-nil updater")
		return
	}

	if updater.backupDir != backupDir {
		t.Errorf("NewUpdaterWithOptions() backupDir = %q, want %q", updater.backupDir, backupDir)
	}

	if updater.keepBackups != keepBackups {
		t.Errorf("NewUpdaterWithOptions() keepBackups = %v, want %v", updater.keepBackups, keepBackups)
	}

	if updater.atomicWrite != atomicWrite {
		t.Errorf("NewUpdaterWithOptions() atomicWrite = %v, want %v", updater.atomicWrite, atomicWrite)
	}
}

func TestUpdateRequest_Validation(t *testing.T) {
	updater := NewUpdater()

	tests := []struct {
		name        string
		request     *UpdateRequest
		expectError bool
		description string
	}{
		{
			name:        "nil request",
			request:     nil,
			expectError: true,
			description: "should reject nil requests",
		},
		{
			name: "empty file path",
			request: &UpdateRequest{
				FilePath:    "",
				NewTagValue: TestNewTag,
			},
			expectError: true,
			description: "should reject empty file paths",
		},
		{
			name: "empty tag value",
			request: &UpdateRequest{
				FilePath:    TestValidPath,
				NewTagValue: "",
			},
			expectError: true,
			description: "should reject empty tag values",
		},
		{
			name: "valid request but nonexistent file",
			request: &UpdateRequest{
				FilePath:      TestValidPath,
				NewTagValue:   TestNewTag,
				CreateBackup:  false,
				ValidateAfter: true,
				DryRun:        true, // Use dry run to avoid file operations
			},
			expectError: true, // Will fail because file doesn't exist, but validation passes
			description: "should validate request structure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := updater.UpdateTagInFile(tt.request)

			if tt.expectError {
				if err == nil {
					t.Errorf("UpdateTagInFile(%v) expected error but got none", tt.request)
				}
			} else {
				if err != nil {
					t.Errorf("UpdateTagInFile(%v) unexpected error: %v", tt.request, err)
				}
			}
		})
	}
}

func TestUpdater_PreviewUpdateDryRun(t *testing.T) {
	// Test with working directory relative path
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Create test file in working directory
	testFile := filepath.Join(workingDir, TestValidPath)
	err = os.WriteFile(testFile, []byte(TestYAMLContent), DefaultFilePermissions)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer func() {
		if removeErr := os.Remove(testFile); removeErr != nil {
			t.Logf("Failed to clean up test file: %v", removeErr)
		}
	}()

	updater := NewUpdater()

	request := &UpdateRequest{
		FilePath:      TestValidPath, // Use relative path
		NewTagValue:   TestNewTag,
		CreateBackup:  false,
		ValidateAfter: true,
		DryRun:        false, // PreviewUpdate will set this to true
	}

	result, err := updater.PreviewUpdate(request)
	if err != nil {
		t.Fatalf("PreviewUpdate failed: %v", err)
	}

	if !result.Success {
		t.Error("PreviewUpdate should succeed")
	}

	if !result.ChangesDetected {
		t.Error("PreviewUpdate should detect changes")
	}

	// Verify original file was not modified
	originalContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read original file: %v", err)
	}

	if !strings.Contains(string(originalContent), TestOldTag) {
		t.Error("Original file should still contain old tag after preview")
	}

	if strings.Contains(string(originalContent), TestNewTag) {
		t.Error("Original file should not contain new tag after preview")
	}
}

func TestSecurityConstants(t *testing.T) {
	// Test that security constants are properly defined
	if DefaultFilePermissions == 0 {
		t.Error("DefaultFilePermissions should not be zero")
	}

	if TempFileSuffix == "" {
		t.Error("TempFileSuffix should not be empty")
	}

	if BackupTimestampFormat == "" {
		t.Error("BackupTimestampFormat should not be empty")
	}

	if MaxBackupFiles <= 0 {
		t.Errorf("MaxBackupFiles should be positive, got %d", MaxBackupFiles)
	}

	if BackupDirPermissions == 0 {
		t.Error("BackupDirPermissions should not be zero")
	}
}

func TestPathTraversalSecurity(t *testing.T) {
	updater := NewUpdater()

	// Test various malicious path patterns
	maliciousPaths := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"/etc/passwd",
		"/etc/shadow",
		"./../../etc/passwd",
		"config/../../../etc/passwd",
		"test/../../../../../../etc/passwd",
		"/usr/bin/bash",
		"/bin/sh",
		"/sbin/init",
		"\\windows\\system32\\drivers\\etc\\hosts",
		"C:\\Windows\\System32\\config\\SAM",
		"/Program Files/malicious.exe",
		"etc/passwd",         // relative path to etc
		"usr/local/bin/test", // relative path to usr
	}

	for _, maliciousPath := range maliciousPaths {
		t.Run("malicious_path_"+maliciousPath, func(t *testing.T) {
			// All these should be rejected by security validation
			_, err := updater.validateAndCleanFilePath(maliciousPath)
			if err == nil {
				t.Errorf("validateAndCleanFilePath(%q) should reject malicious path", maliciousPath)
			}

			// fileExists should return false for malicious paths
			exists := updater.fileExists(maliciousPath)
			if exists {
				t.Errorf("fileExists(%q) should return false for malicious path", maliciousPath)
			}
		})
	}
}

func TestPathSecurityAllowedPaths(t *testing.T) {
	updater := NewUpdater()

	// Test paths that should be allowed
	allowedPaths := []string{
		"config.yaml",
		"./config.yaml",
		"configs/app.yaml",
		"test/data/sample.yaml",
		"/tmp/test.yaml",
		"/var/tmp/backup.yaml",
		"C:/temp/config.yaml",
		"D:/tmp/test.yaml",
	}

	for _, allowedPath := range allowedPaths {
		t.Run("allowed_path_"+allowedPath, func(t *testing.T) {
			// These should pass validation (though file may not exist)
			_, err := updater.validateAndCleanFilePath(allowedPath)
			if err != nil {
				t.Errorf("validateAndCleanFilePath(%q) should allow safe path, got error: %v", allowedPath, err)
			}
		})
	}
}

// Benchmark tests for performance requirements from IDEA.md
func BenchmarkUpdater_validateAndCleanFilePath(b *testing.B) {
	updater := NewUpdater()
	for i := 0; i < b.N; i++ {
		_, _ = updater.validateAndCleanFilePath(TestValidPath)
	}
}

func BenchmarkUpdater_NewUpdater(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewUpdater()
	}
}

func BenchmarkUpdater_fileExists(b *testing.B) {
	updater := NewUpdater()
	for i := 0; i < b.N; i++ {
		updater.fileExists(TestValidPath)
	}
}
