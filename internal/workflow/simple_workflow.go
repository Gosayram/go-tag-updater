// Package workflow provides simple execution logic for tag updates
package workflow

import (
	"context"
	"fmt"
	"os"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/Gosayram/go-tag-updater/internal/config"
	gitlabapi "github.com/Gosayram/go-tag-updater/internal/gitlab"
	"github.com/Gosayram/go-tag-updater/internal/logger"
	"github.com/Gosayram/go-tag-updater/internal/yaml"
	"github.com/Gosayram/go-tag-updater/pkg/errors"
)

const (
	// PreviewContentMaxLength defines maximum length for content preview in dry run
	PreviewContentMaxLength = 500
	// TempFilePermissions defines permissions for temporary files
	TempFilePermissions = 0o600
)

// SimpleTagUpdater handles basic tag update workflow
type SimpleTagUpdater struct {
	config       *config.CLIConfig
	logger       *logger.Logger
	gitlabClient *gitlabapi.Client
	fileManager  *gitlabapi.FileManager
	branchMgr    *gitlabapi.BranchManager
	mrManager    *gitlabapi.SimpleMergeRequestManager
	projectID    int
}

// SimpleUpdateResult contains the results of the update operation
type SimpleUpdateResult struct {
	Success      bool
	BranchName   string
	MergeRequest *gitlab.MergeRequest
	FileUpdated  bool
	Message      string
}

// NewSimpleTagUpdater creates a new simple tag updater
func NewSimpleTagUpdater(cfg *config.CLIConfig, log *logger.Logger) (*SimpleTagUpdater, error) {
	if cfg == nil {
		return nil, errors.NewValidationError("config cannot be nil")
	}
	if log == nil {
		return nil, errors.NewValidationError("logger cannot be nil")
	}

	return &SimpleTagUpdater{
		config: cfg,
		logger: log,
	}, nil
}

// Initialize sets up the GitLab client and managers
func (stu *SimpleTagUpdater) Initialize(_ context.Context) error {
	// Create GitLab client
	client, err := gitlabapi.NewClient(stu.config.GitLabToken, stu.config.GitLabURL)
	if err != nil {
		return fmt.Errorf("failed to create GitLab client: %w", err)
	}

	stu.gitlabClient = client

	// Resolve project ID
	stu.projectID, err = stu.gitlabClient.ResolveProjectID(stu.config.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to resolve project ID %s: %w", stu.config.ProjectID, err)
	}

	stu.logger.WithProjectID(stu.projectID).WithField("project_path", stu.config.ProjectID).
		Info("Project ID resolved successfully")

	// Initialize managers
	stu.fileManager = gitlabapi.NewFileManager(client.GetGitLabClient(), stu.projectID)
	stu.branchMgr = gitlabapi.NewBranchManager(client.GetGitLabClient(), stu.projectID)
	stu.mrManager = gitlabapi.NewSimpleMergeRequestManager(client.GetGitLabClient(), stu.projectID)

	// Health check
	if err := stu.gitlabClient.IsHealthy(); err != nil {
		return fmt.Errorf("GitLab health check failed: %w", err)
	}

	stu.logger.WithOperation("health_check").Info("GitLab client initialized successfully")
	return nil
}

// Execute runs the basic tag update workflow
func (stu *SimpleTagUpdater) Execute(ctx context.Context) (*SimpleUpdateResult, error) {
	result := &SimpleUpdateResult{
		Success: false,
	}

	stu.logger.WithFields(map[string]interface{}{
		"file_path":  stu.config.FilePath,
		"new_tag":    stu.config.NewTag,
		"project_id": stu.config.ProjectID,
		"operation":  "tag_update_start",
	}).Info("Starting tag update workflow")

	// Step 1: Validate file and get content
	newContent, err := stu.validateAndUpdateContent(ctx)
	if err != nil {
		return result, err
	}

	// Step 2: Generate unique branch name
	branchName, err := stu.prepareBranchName(ctx)
	if err != nil {
		return result, err
	}
	result.BranchName = branchName

	// Step 3: Handle dry run
	if stu.config.DryRun {
		return stu.handleDryRun(result, newContent), nil
	}

	// Step 4: Execute actual update
	return stu.executeUpdate(ctx, result, newContent, branchName)
}

// validateAndUpdateContent validates the file exists and updates its content
func (stu *SimpleTagUpdater) validateAndUpdateContent(ctx context.Context) (string, error) {
	// Check if file exists
	exists, err := stu.fileManager.FileExists(ctx, stu.config.FilePath, stu.config.TargetBranch)
	if err != nil {
		stu.logger.WithError(err).WithFields(map[string]interface{}{
			"file_path": stu.config.FilePath,
			"branch":    stu.config.TargetBranch,
		}).Error("Failed to check file existence")
		return "", fmt.Errorf("failed to check if file exists: %w", err)
	}

	if !exists {
		stu.logger.WithFields(map[string]interface{}{
			"file_path": stu.config.FilePath,
			"branch":    stu.config.TargetBranch,
		}).Error("File does not exist in target branch")
		return "", fmt.Errorf("file %s does not exist in branch %s", stu.config.FilePath, stu.config.TargetBranch)
	}

	stu.logger.WithFields(map[string]interface{}{
		"file_path": stu.config.FilePath,
		"branch":    stu.config.TargetBranch,
	}).Info("File exists in target branch")

	// Get current file content
	content, err := stu.fileManager.GetFileContent(ctx, stu.config.FilePath, stu.config.TargetBranch)
	if err != nil {
		stu.logger.WithError(err).WithField("file_path", stu.config.FilePath).
			Error("Failed to get file content")
		return "", fmt.Errorf("failed to get file content: %w", err)
	}

	// Update YAML content
	newContent, err := stu.updateYAMLContent(content)
	if err != nil {
		stu.logger.WithError(err).WithField("file_path", stu.config.FilePath).
			Error("Failed to update YAML content")
		return "", fmt.Errorf("failed to update YAML content: %w", err)
	}

	stu.logger.WithFields(map[string]interface{}{
		"file_path": stu.config.FilePath,
		"new_tag":   stu.config.NewTag,
	}).Info("YAML content updated successfully")

	return newContent, nil
}

// updateYAMLContent updates YAML content using the proper parser
func (stu *SimpleTagUpdater) updateYAMLContent(content string) (string, error) {
	yamlUpdater := yaml.NewUpdater()

	// Create temporary file with content for validation
	tempFile, err := stu.createTempFileWithContent(content)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		if removeErr := os.Remove(tempFile); removeErr != nil {
			stu.logger.WithError(removeErr).WithField("temp_file", tempFile).
				Warn("Failed to remove temporary file")
		}
	}()

	// Validate the existing YAML
	if validationErr := yamlUpdater.ValidateFile(tempFile); validationErr != nil {
		return "", fmt.Errorf("invalid YAML in source file: %w", validationErr)
	}

	// Update the content
	request := &yaml.UpdateRequest{
		FilePath:      tempFile,
		NewTagValue:   stu.config.NewTag,
		CreateBackup:  false,
		ValidateAfter: true,
		DryRun:        true, // We only want the updated content, not to write it
	}

	result, err := yamlUpdater.UpdateTagInFile(request)
	if err != nil {
		return "", fmt.Errorf("failed to update YAML tag: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("YAML update was not successful")
	}

	newContent := result.UpdatedContent

	return newContent, nil
}

// createTempFileWithContent creates a temporary file with given content
func (stu *SimpleTagUpdater) createTempFileWithContent(content string) (string, error) {
	// Create temp file in current directory to avoid YAML validator path restrictions
	tempFile, err := os.CreateTemp(".", "go-tag-updater-*.yaml")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	if _, err := tempFile.WriteString(content); err != nil {
		_ = os.Remove(tempFile.Name()) // Ignore cleanup error, return original write error
		return "", err
	}

	return tempFile.Name(), nil
}

// prepareBranchName generates a unique branch name
func (stu *SimpleTagUpdater) prepareBranchName(ctx context.Context) (string, error) {
	var branchName string
	if stu.config.BranchName != "" {
		branchName = stu.config.BranchName
	} else {
		var err error
		branchName, err = stu.branchMgr.GenerateUniqueBranchName(ctx, "update-tag", stu.config.NewTag)
		if err != nil {
			return "", fmt.Errorf("failed to generate branch name: %w", err)
		}
	}

	stu.logger.WithFields(map[string]interface{}{
		"branch_name":    branchName,
		"auto_generated": stu.config.BranchName == "",
	}).Info("Branch name prepared")

	return branchName, nil
}

// handleDryRun handles dry run mode
func (stu *SimpleTagUpdater) handleDryRun(result *SimpleUpdateResult, newContent string) *SimpleUpdateResult {
	maxLen := minInt(PreviewContentMaxLength, len(newContent))

	stu.logger.WithFields(map[string]interface{}{
		"operation":      "dry_run",
		"branch_name":    result.BranchName,
		"content_length": len(newContent),
		"preview_length": maxLen,
	}).Info("Dry run mode: would create branch and update file")

	stu.logger.WithField("content_preview", newContent[:maxLen]).Debug("Content preview")

	result.Success = true
	result.Message = "Dry run completed successfully"
	return result
}

// executeUpdate performs the actual update operations
func (stu *SimpleTagUpdater) executeUpdate(
	ctx context.Context,
	result *SimpleUpdateResult,
	newContent, branchName string,
) (*SimpleUpdateResult, error) {
	// Create branch
	_, err := stu.branchMgr.CreateBranch(ctx, branchName, stu.config.TargetBranch)
	if err != nil {
		stu.logger.WithError(err).WithFields(map[string]interface{}{
			"branch_name":   branchName,
			"source_branch": stu.config.TargetBranch,
		}).Error("Failed to create branch")
		return result, fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}

	stu.logger.WithFields(map[string]interface{}{
		"branch_name":   branchName,
		"source_branch": stu.config.TargetBranch,
	}).Info("Branch created successfully")

	// Update file with new content
	updateOpts := &gitlabapi.FileUpdateOptions{
		Branch:        branchName,
		CommitMessage: fmt.Sprintf("Update tag to %s in %s", stu.config.NewTag, stu.config.FilePath),
		Content:       newContent,
	}

	_, err = stu.fileManager.UpdateFileContent(ctx, stu.config.FilePath, updateOpts)
	if err != nil {
		// Try to cleanup branch on failure
		_ = stu.branchMgr.DeleteBranch(ctx, branchName)
		stu.logger.WithError(err).WithFields(map[string]interface{}{
			"file_path":   stu.config.FilePath,
			"branch_name": branchName,
		}).Error("Failed to update file, branch cleaned up")
		return result, fmt.Errorf("failed to update file: %w", err)
	}

	result.FileUpdated = true
	stu.logger.WithFields(map[string]interface{}{
		"file_path":   stu.config.FilePath,
		"branch_name": branchName,
		"new_tag":     stu.config.NewTag,
	}).Info("File updated successfully")

	// Create merge request
	mrDescription := fmt.Sprintf("Automated tag update to %s\n\nFile: %s\nBranch: %s",
		stu.config.NewTag, stu.config.FilePath, branchName)
	mrOpts := &gitlabapi.SimpleMergeRequestOptions{
		Title:        fmt.Sprintf("Update tag to %s in %s", stu.config.NewTag, stu.config.FilePath),
		Description:  mrDescription,
		SourceBranch: branchName,
		TargetBranch: stu.config.TargetBranch,
	}

	mr, err := stu.mrManager.CreateMergeRequest(ctx, mrOpts)
	if err != nil {
		stu.logger.WithError(err).WithFields(map[string]interface{}{
			"branch_name":   branchName,
			"target_branch": stu.config.TargetBranch,
		}).Error("Failed to create merge request")
		return result, fmt.Errorf("failed to create merge request: %w", err)
	}

	result.MergeRequest = mr
	stu.logger.WithFields(map[string]interface{}{
		"mr_id":       mr.IID,
		"mr_url":      mr.WebURL,
		"branch_name": branchName,
	}).Info("Merge request created successfully")

	result.Success = true
	result.Message = fmt.Sprintf("Tag update completed successfully. MR: !%d", mr.IID)
	return result, nil
}

// Cleanup performs cleanup operations
func (stu *SimpleTagUpdater) Cleanup() error {
	// Currently no cleanup needed
	return nil
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
