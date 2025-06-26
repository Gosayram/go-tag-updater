// Package workflow provides simple execution logic for tag updates
package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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

	stu.logger.Infof("Resolved project ID: %s -> %d", stu.config.ProjectID, stu.projectID)

	// Initialize managers
	stu.fileManager = gitlabapi.NewFileManager(client.GetGitLabClient(), stu.projectID)
	stu.branchMgr = gitlabapi.NewBranchManager(client.GetGitLabClient(), stu.projectID)
	stu.mrManager = gitlabapi.NewSimpleMergeRequestManager(client.GetGitLabClient(), stu.projectID)

	// Health check
	if err := stu.gitlabClient.IsHealthy(); err != nil {
		return fmt.Errorf("GitLab health check failed: %w", err)
	}

	stu.logger.Info("GitLab client initialized successfully")
	return nil
}

// Execute runs the basic tag update workflow
func (stu *SimpleTagUpdater) Execute(ctx context.Context) (*SimpleUpdateResult, error) {
	result := &SimpleUpdateResult{
		Success: false,
	}

	stu.logger.Infof("Starting tag update: %s -> %s in %s",
		stu.config.FilePath, stu.config.NewTag, stu.config.ProjectID)

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

// updateYAMLContent updates YAML content using the proper parser
func updateYAMLContent(content, newTag string, _ bool) (string, error) {
	parser := yaml.NewParser()

	// Try simple update first (most common case)
	updatedContent, err := parser.UpdateTagSimple(content, newTag)
	if err == nil {
		return updatedContent, nil
	}

	// If simple update failed, try to find tag locations manually
	parseResult, err := parser.ParseContent(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse YAML: %w", err)
	}

	tagLocations := parser.ListAllTags(parseResult)
	if len(tagLocations) == 0 {
		return "", errors.NewValidationError("no tag fields found in YAML content")
	}

	// Use the first detected tag location
	updateOptions := &yaml.UpdateOptions{
		TagPath:         tagLocations[0].Path,
		NewValue:        newTag,
		CreateIfMissing: false,
	}

	return parser.UpdateTag(parseResult, updateOptions)
}

// createTempFileWithContent creates a temporary file with given content for validation
func createTempFileWithContent(content string) string {
	tempFile := filepath.Join(os.TempDir(), "go-tag-updater-temp.yaml")
	_ = os.WriteFile(tempFile, []byte(content), TempFilePermissions)
	return tempFile
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// validateAndUpdateContent validates file and updates YAML content
func (stu *SimpleTagUpdater) validateAndUpdateContent(ctx context.Context) (string, error) {
	// Check if file exists
	exists, err := stu.fileManager.FileExists(ctx, stu.config.FilePath, stu.config.TargetBranch)
	if err != nil {
		return "", fmt.Errorf("failed to check file existence: %w", err)
	}

	if !exists {
		return "", errors.NewFileNotFoundError(stu.config.FilePath)
	}

	stu.logger.Infof("File %s exists in branch %s", stu.config.FilePath, stu.config.TargetBranch)

	// Get current file content
	fileContent, err := stu.fileManager.GetFileContent(ctx, stu.config.FilePath, stu.config.TargetBranch)
	if err != nil {
		return "", fmt.Errorf("failed to get file content: %w", err)
	}

	// Validate and update YAML content
	yamlUpdater := yaml.NewUpdater()

	// Validate the existing YAML
	if validationErr := yamlUpdater.ValidateFile(createTempFileWithContent(fileContent)); validationErr != nil {
		return "", fmt.Errorf("invalid YAML in source file: %w", validationErr)
	}

	// Update the content
	newContent, err := updateYAMLContent(fileContent, stu.config.NewTag, stu.config.DryRun)
	if err != nil {
		return "", fmt.Errorf("failed to update YAML content: %w", err)
	}

	stu.logger.Infof("YAML content updated successfully")
	return newContent, nil
}

// prepareBranchName generates a unique branch name
func (stu *SimpleTagUpdater) prepareBranchName(ctx context.Context) (string, error) {
	branchName := stu.config.BranchName
	if branchName == "" {
		var err error
		branchName, err = stu.branchMgr.GenerateUniqueBranchName(ctx, "update-tag/", stu.config.NewTag)
		if err != nil {
			return "", fmt.Errorf("failed to generate branch name: %w", err)
		}
	}

	stu.logger.Infof("Using branch name: %s", branchName)
	return branchName, nil
}

// handleDryRun handles dry run mode execution
func (stu *SimpleTagUpdater) handleDryRun(result *SimpleUpdateResult, newContent string) *SimpleUpdateResult {
	stu.logger.Info("Dry run mode: would create branch and update file")
	maxLen := minInt(PreviewContentMaxLength, len(newContent))
	stu.logger.Infof("Content preview:\n%s", newContent[:maxLen])
	result.Success = true
	result.Message = "Dry run completed successfully"
	return result
}

// executeUpdate performs the actual update operations
func (stu *SimpleTagUpdater) executeUpdate(ctx context.Context, result *SimpleUpdateResult,
	newContent, branchName string) (*SimpleUpdateResult, error) {
	// Create branch
	_, err := stu.branchMgr.CreateBranch(ctx, branchName, stu.config.TargetBranch)
	if err != nil {
		return result, fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}

	stu.logger.Infof("Created branch: %s", branchName)

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
		return result, fmt.Errorf("failed to update file: %w", err)
	}

	result.FileUpdated = true
	stu.logger.Infof("Updated file %s with tag %s", stu.config.FilePath, stu.config.NewTag)

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
		return result, fmt.Errorf("failed to create merge request: %w", err)
	}

	result.MergeRequest = mr
	stu.logger.Infof("Created merge request: !%d - %s", mr.IID, mr.WebURL)

	result.Success = true
	result.Message = fmt.Sprintf("Tag update completed successfully. MR: !%d", mr.IID)
	return result, nil
}

// Cleanup performs cleanup operations
func (stu *SimpleTagUpdater) Cleanup() error {
	// Currently no cleanup needed
	return nil
}
