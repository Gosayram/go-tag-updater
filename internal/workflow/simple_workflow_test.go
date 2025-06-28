package workflow

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gosayram/go-tag-updater/internal/config"
	"github.com/Gosayram/go-tag-updater/internal/logger"
)

const (
	// Test constants as per IDEA.md requirements
	TestTimeout       = 5000 // milliseconds
	MaxTestRetries    = 3
	TestProjectID     = "test/project"
	TestGitLabToken   = "test-token"
	TestGitLabURL     = "https://gitlab.example.com"
	TestFilePath      = "deployment.yaml"
	TestNewTag        = "v1.2.3"
	TestOldTag        = "v1.0.0"
	TestTargetBranch  = "main"
	TestBranchName    = "update-tag/v1.2.3"
	TestCommitMessage = "Update tag to v1.2.3"
	TestYAMLContent   = `
name: test-app
version: 1.0.0
image:
  tag: v1.0.0
  repository: test/repo
`
	TestYAMLContentUpdated = `
name: test-app
version: 1.0.0
image:
  tag: v1.2.3
  repository: test/repo
`
	TestInvalidYAMLContent = `
name: test-app
version: 1.0.0
image:
  tag: v1.0.0
  repository: test/repo
invalid yaml content: [
`
)

func TestNewSimpleTagUpdater(t *testing.T) {
	log := logger.New(false)

	tests := []struct {
		name        string
		config      *config.CLIConfig
		logger      *logger.Logger
		expectError bool
		description string
	}{
		{
			name: "valid config and logger",
			config: &config.CLIConfig{
				ProjectID:    TestProjectID,
				GitLabToken:  TestGitLabToken,
				FilePath:     TestFilePath,
				NewTag:       TestNewTag,
				TargetBranch: TestTargetBranch,
			},
			logger:      log,
			expectError: false,
			description: "should create updater with valid inputs",
		},
		{
			name:        "nil config",
			config:      nil,
			logger:      log,
			expectError: true,
			description: "should reject nil config",
		},
		{
			name: "nil logger",
			config: &config.CLIConfig{
				ProjectID: TestProjectID,
			},
			logger:      nil,
			expectError: true,
			description: "should reject nil logger",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater, err := NewSimpleTagUpdater(tt.config, tt.logger)

			if tt.expectError {
				if err == nil {
					t.Errorf("NewSimpleTagUpdater() expected error but got none")
				}
				if updater != nil {
					t.Errorf("NewSimpleTagUpdater() expected nil updater on error")
				}
			} else {
				if err != nil {
					t.Errorf("NewSimpleTagUpdater() unexpected error: %v", err)
				}
				if updater == nil {
					t.Errorf("NewSimpleTagUpdater() expected non-nil updater")
				} else {
					if updater.config != tt.config {
						t.Errorf("NewSimpleTagUpdater() config not set correctly")
					}
					if updater.logger != tt.logger {
						t.Errorf("NewSimpleTagUpdater() logger not set correctly")
					}
				}
			}
		})
	}
}

func TestSimpleUpdateResult_Structure(t *testing.T) {
	result := &SimpleUpdateResult{
		Success:     true,
		BranchName:  TestBranchName,
		FileUpdated: true,
		Message:     "Test message",
	}

	if !result.Success {
		t.Error("SimpleUpdateResult.Success should be settable")
	}

	if result.BranchName != TestBranchName {
		t.Errorf("SimpleUpdateResult.BranchName = %q, want %q", result.BranchName, TestBranchName)
	}

	if !result.FileUpdated {
		t.Error("SimpleUpdateResult.FileUpdated should be settable")
	}

	if result.Message != "Test message" {
		t.Errorf("SimpleUpdateResult.Message = %q, want %q", result.Message, "Test message")
	}
}

func TestUpdateYAMLContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		newTag      string
		dryRun      bool
		expectError bool
		description string
	}{
		{
			name:        "valid YAML with tag",
			content:     TestYAMLContent,
			newTag:      TestNewTag,
			dryRun:      false,
			expectError: false,
			description: "should update valid YAML content",
		},
		{
			name:        "empty content",
			content:     "",
			newTag:      TestNewTag,
			dryRun:      false,
			expectError: true,
			description: "should reject empty content",
		},
		{
			name:        "invalid YAML",
			content:     TestInvalidYAMLContent,
			newTag:      TestNewTag,
			dryRun:      false,
			expectError: true,
			description: "should reject invalid YAML",
		},
		{
			name:        "empty tag",
			content:     TestYAMLContent,
			newTag:      "",
			dryRun:      false,
			expectError: true,
			description: "should reject empty tag",
		},
		{
			name: "YAML without tags",
			content: `
name: test-app
version: 1.0.0
description: "No tags here"
`,
			newTag:      TestNewTag,
			dryRun:      false,
			expectError: false,
			description: "should handle YAML without tag fields gracefully",
		},
	}

	// Create logger for testing
	log := logger.New(false)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create updater with test-specific tag
			testCfg := &config.CLIConfig{
				ProjectID:    TestProjectID,
				GitLabToken:  TestGitLabToken,
				FilePath:     TestFilePath,
				NewTag:       tt.newTag,
				TargetBranch: TestTargetBranch,
			}

			testUpdater, err := NewSimpleTagUpdater(testCfg, log)
			if err != nil {
				t.Fatalf("Failed to create test updater: %v", err)
			}

			result, err := testUpdater.updateYAMLContent(tt.content)

			if tt.expectError {
				if err == nil {
					t.Errorf("updateYAMLContent() expected error but got none")
				}
				if result != "" {
					t.Errorf("updateYAMLContent() expected empty result on error")
				}
			} else {
				if err != nil {
					t.Errorf("updateYAMLContent() unexpected error: %v", err)
				}
				if result == "" {
					t.Errorf("updateYAMLContent() expected non-empty result")
				}
				if tt.newTag != "" && !strings.Contains(result, tt.newTag) {
					t.Errorf("updateYAMLContent() result should contain new tag %q", tt.newTag)
				}
			}
		})
	}
}

func TestCreateTempFileWithContent(t *testing.T) {
	log := logger.New(false)
	cfg := &config.CLIConfig{
		ProjectID:    TestProjectID,
		GitLabToken:  TestGitLabToken,
		FilePath:     TestFilePath,
		NewTag:       TestNewTag,
		TargetBranch: TestTargetBranch,
	}

	updater, err := NewSimpleTagUpdater(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create updater: %v", err)
	}

	testContent := TestYAMLContent

	tempFile2, tempErr := updater.createTempFileWithContent(testContent)
	if tempErr != nil {
		t.Fatalf("createTempFileWithContent() failed: %v", tempErr)
	}
	tempFile := tempFile2

	// Check that file was created
	if tempFile == "" {
		t.Error("createTempFileWithContent() should return non-empty path")
	}

	// Check that file exists
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Errorf("createTempFileWithContent() file should exist at %s", tempFile)
	}

	// Check content
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Errorf("createTempFileWithContent() failed to read created file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("createTempFileWithContent() content mismatch")
	}

	// Check permissions
	info, err := os.Stat(tempFile)
	if err != nil {
		t.Errorf("createTempFileWithContent() failed to stat file: %v", err)
	}

	expectedPerms := os.FileMode(TempFilePermissions)
	if info.Mode().Perm() != expectedPerms {
		t.Errorf("createTempFileWithContent() permissions = %o, want %o", info.Mode().Perm(), expectedPerms)
	}

	// Cleanup
	defer func() {
		if removeErr := os.Remove(tempFile); removeErr != nil {
			t.Logf("Failed to clean up temp file: %v", removeErr)
		}
	}()
}

func TestMinInt(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "a smaller than b",
			a:        5,
			b:        10,
			expected: 5,
		},
		{
			name:     "b smaller than a",
			a:        10,
			b:        5,
			expected: 5,
		},
		{
			name:     "equal values",
			a:        7,
			b:        7,
			expected: 7,
		},
		{
			name:     "negative values",
			a:        -5,
			b:        -10,
			expected: -10,
		},
		{
			name:     "zero and positive",
			a:        0,
			b:        5,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := minInt(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("minInt(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestSimpleTagUpdater_HandleDryRun(t *testing.T) {
	log := logger.New(false)
	cfg := &config.CLIConfig{
		ProjectID:    TestProjectID,
		GitLabToken:  TestGitLabToken,
		FilePath:     TestFilePath,
		NewTag:       TestNewTag,
		TargetBranch: TestTargetBranch,
		DryRun:       true,
	}

	updater, err := NewSimpleTagUpdater(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create updater: %v", err)
	}

	initialResult := &SimpleUpdateResult{
		BranchName: TestBranchName,
	}

	result := updater.handleDryRun(initialResult, TestYAMLContentUpdated)

	if !result.Success {
		t.Error("handleDryRun() should set Success to true")
	}

	if result.Message == "" {
		t.Error("handleDryRun() should set a message")
	}

	if result.BranchName != TestBranchName {
		t.Errorf("handleDryRun() should preserve BranchName, got %q", result.BranchName)
	}
}

func TestSimpleTagUpdater_PrepareBranchName(t *testing.T) {
	log := logger.New(false)

	// Test with custom branch name (doesn't require GitLab client)
	t.Run("custom branch name provided", func(t *testing.T) {
		cfg := &config.CLIConfig{
			ProjectID:    TestProjectID,
			GitLabToken:  TestGitLabToken,
			FilePath:     TestFilePath,
			NewTag:       TestNewTag,
			TargetBranch: TestTargetBranch,
			BranchName:   TestBranchName,
		}

		updater, err := NewSimpleTagUpdater(cfg, log)
		if err != nil {
			t.Fatalf("Failed to create updater: %v", err)
		}

		ctx := context.Background()
		branchName, err := updater.prepareBranchName(ctx)

		if err != nil {
			t.Errorf("prepareBranchName() unexpected error: %v", err)
		}

		if branchName != TestBranchName {
			t.Errorf("prepareBranchName() = %q, want %q", branchName, TestBranchName)
		}
	})

	// Note: Testing auto-generation requires GitLab client initialization
	// which is complex to mock, so we only test the direct branch name case
}

func TestConstants(t *testing.T) {
	// Test that constants are properly defined
	if PreviewContentMaxLength <= 0 {
		t.Errorf("PreviewContentMaxLength should be positive, got %d", PreviewContentMaxLength)
	}

	if TempFilePermissions == 0 {
		t.Error("TempFilePermissions should not be zero")
	}

	// Test reasonable values
	if PreviewContentMaxLength < 100 {
		t.Errorf("PreviewContentMaxLength seems too small: %d", PreviewContentMaxLength)
	}

	if PreviewContentMaxLength > 10000 {
		t.Errorf("PreviewContentMaxLength seems too large: %d", PreviewContentMaxLength)
	}

	// Test that temp file permissions are restrictive
	expectedPerms := os.FileMode(0o600)
	if TempFilePermissions != expectedPerms {
		t.Errorf("TempFilePermissions = %o, want %o for security", TempFilePermissions, expectedPerms)
	}
}

func TestSimpleTagUpdater_Cleanup(t *testing.T) {
	log := logger.New(false)
	cfg := &config.CLIConfig{
		ProjectID:    TestProjectID,
		GitLabToken:  TestGitLabToken,
		FilePath:     TestFilePath,
		NewTag:       TestNewTag,
		TargetBranch: TestTargetBranch,
	}

	updater, err := NewSimpleTagUpdater(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create updater: %v", err)
	}

	err = updater.Cleanup()
	if err != nil {
		t.Errorf("Cleanup() should not return error, got: %v", err)
	}
}

func TestTempFileCleanup(t *testing.T) {
	// Test that temp files can be cleaned up properly
	log := logger.New(false)
	cfg := &config.CLIConfig{
		ProjectID:    TestProjectID,
		GitLabToken:  TestGitLabToken,
		FilePath:     TestFilePath,
		NewTag:       TestNewTag,
		TargetBranch: TestTargetBranch,
	}

	updater, err := NewSimpleTagUpdater(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create updater: %v", err)
	}

	testContent := "test content"
	tempFile, tempErr := updater.createTempFileWithContent(testContent)
	if tempErr != nil {
		t.Fatalf("createTempFileWithContent() failed: %v", tempErr)
	}

	// Verify file exists
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("Temp file should exist after creation")
	}

	// Clean up
	err = os.Remove(tempFile)
	if err != nil {
		t.Errorf("Failed to remove temp file: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Error("Temp file should not exist after cleanup")
	}
}

func TestTempFileLocation(t *testing.T) {
	log := logger.New(false)
	cfg := &config.CLIConfig{
		ProjectID:    TestProjectID,
		GitLabToken:  TestGitLabToken,
		FilePath:     TestFilePath,
		NewTag:       TestNewTag,
		TargetBranch: TestTargetBranch,
	}

	updater, err := NewSimpleTagUpdater(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create updater: %v", err)
	}

	testContent := "test content"
	tempFile, err := updater.createTempFileWithContent(testContent)
	if err != nil {
		t.Fatalf("createTempFileWithContent() failed: %v", err)
	}
	defer func() {
		if removeErr := os.Remove(tempFile); removeErr != nil {
			t.Logf("Failed to clean up temp file: %v", removeErr)
		}
	}()

	// Check that temp file is in current directory (not system temp)
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	expectedDir := filepath.Clean(currentDir)

	// Get absolute path of temp file for comparison
	absTempFile, err := filepath.Abs(tempFile)
	if err != nil {
		t.Fatalf("Failed to get absolute path of temp file: %v", err)
	}
	actualDir := filepath.Clean(filepath.Dir(absTempFile))

	if actualDir != expectedDir {
		t.Errorf("Temp file created in %q, expected %q", actualDir, expectedDir)
	}

	// Check filename pattern - should start with "go-tag-updater-" and end with ".yaml"
	actualName := filepath.Base(tempFile)
	expectedPrefix := "go-tag-updater-"
	expectedSuffix := ".yaml"

	if !strings.HasPrefix(actualName, expectedPrefix) {
		t.Errorf("Temp file name = %q, should start with %q", actualName, expectedPrefix)
	}

	if !strings.HasSuffix(actualName, expectedSuffix) {
		t.Errorf("Temp file name = %q, should end with %q", actualName, expectedSuffix)
	}
}

// Benchmark tests for performance requirements from IDEA.md
func BenchmarkUpdateYAMLContent(b *testing.B) {
	log := logger.New(false)
	cfg := &config.CLIConfig{
		ProjectID:    TestProjectID,
		GitLabToken:  TestGitLabToken,
		FilePath:     TestFilePath,
		NewTag:       TestNewTag,
		TargetBranch: TestTargetBranch,
	}

	updater, err := NewSimpleTagUpdater(cfg, log)
	if err != nil {
		b.Fatalf("Failed to create updater: %v", err)
	}

	content := TestYAMLContent

	for i := 0; i < b.N; i++ {
		_, _ = updater.updateYAMLContent(content)
	}
}

func BenchmarkMinInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = minInt(i, i+1)
	}
}

func BenchmarkCreateTempFileWithContent(b *testing.B) {
	log := logger.New(false)
	cfg := &config.CLIConfig{
		ProjectID:    TestProjectID,
		GitLabToken:  TestGitLabToken,
		FilePath:     TestFilePath,
		NewTag:       TestNewTag,
		TargetBranch: TestTargetBranch,
	}

	updater, err := NewSimpleTagUpdater(cfg, log)
	if err != nil {
		b.Fatalf("Failed to create updater: %v", err)
	}

	content := TestYAMLContent

	for i := 0; i < b.N; i++ {
		tempFile, err := updater.createTempFileWithContent(content)
		if err != nil {
			continue
		}
		// Clean up immediately to avoid filling disk
		_ = os.Remove(tempFile)
	}
}

func BenchmarkNewSimpleTagUpdater(b *testing.B) {
	log := logger.New(false)
	cfg := &config.CLIConfig{
		ProjectID:    TestProjectID,
		GitLabToken:  TestGitLabToken,
		FilePath:     TestFilePath,
		NewTag:       TestNewTag,
		TargetBranch: TestTargetBranch,
	}

	for i := 0; i < b.N; i++ {
		_, _ = NewSimpleTagUpdater(cfg, log)
	}
}
