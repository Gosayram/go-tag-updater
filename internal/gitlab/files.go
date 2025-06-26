// Package gitlab provides utilities for GitLab API operations
package gitlab

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/Gosayram/go-tag-updater/pkg/errors"
)

const (
	// DefaultCommitMessage for file updates
	DefaultCommitMessage = "Update file via go-tag-updater"
	// MaxFileSize defines maximum file size for processing
	MaxFileSize = 1024 * 1024 // 1MB
	// DefaultBranch is the default branch name
	DefaultBranch = "main"
)

// FileManager handles repository file operations
type FileManager struct {
	client    *gitlab.Client
	projectID interface{}
}

// FileInfo represents file information
type FileInfo struct {
	FilePath   string
	Content    string
	SHA        string
	Size       int64
	Encoding   string
	Branch     string
	LastCommit *gitlab.Commit
}

// FileUpdateOptions contains options for file updates
type FileUpdateOptions struct {
	Branch        string
	Content       string
	CommitMessage string
	AuthorEmail   string
	AuthorName    string
	StartBranch   string
}

// NewFileManager creates a new file manager
func NewFileManager(client *gitlab.Client, projectID interface{}) *FileManager {
	return &FileManager{
		client:    client,
		projectID: projectID,
	}
}

// GetFile retrieves file content from repository
func (fm *FileManager) GetFile(ctx context.Context, filePath, branch string) (*FileInfo, error) {
	if filePath == "" {
		return nil, errors.NewValidationError("file path cannot be empty")
	}

	if branch == "" {
		branch = DefaultBranch
	}

	opts := &gitlab.GetFileOptions{
		Ref: gitlab.Ptr(branch),
	}

	file, _, err := fm.client.RepositoryFiles.GetFile(fm.projectID, filePath, opts)
	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to get file %s: %v", filePath, err))
	}

	// Decode base64 content
	content, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return nil, errors.NewFileSystemError(fmt.Sprintf("failed to decode file content: %v", err))
	}

	return &FileInfo{
		FilePath:   filePath,
		Content:    string(content),
		SHA:        file.BlobID,
		Size:       int64(file.Size),
		Encoding:   file.Encoding,
		Branch:     branch,
		LastCommit: nil, // LastCommit not available in File struct
	}, nil
}

// UpdateFile updates file content in repository
func (fm *FileManager) UpdateFile(ctx context.Context, filePath string, opts *FileUpdateOptions) (*gitlab.FileInfo, error) {
	if filePath == "" {
		return nil, errors.NewValidationError("file path cannot be empty")
	}

	if opts == nil {
		return nil, errors.NewValidationError("update options cannot be nil")
	}

	if opts.Content == "" {
		return nil, errors.NewValidationError("file content cannot be empty")
	}

	if opts.Branch == "" {
		opts.Branch = DefaultBranch
	}

	if opts.CommitMessage == "" {
		opts.CommitMessage = fmt.Sprintf("Update %s", filePath)
	}

	// Check if file exists
	_, err := fm.GetFile(ctx, filePath, opts.Branch)
	fileExists := err == nil

	updateOpts := &gitlab.UpdateFileOptions{
		Branch:        gitlab.Ptr(opts.Branch),
		Content:       gitlab.Ptr(opts.Content),
		CommitMessage: gitlab.Ptr(opts.CommitMessage),
	}

	if opts.AuthorEmail != "" {
		updateOpts.AuthorEmail = gitlab.Ptr(opts.AuthorEmail)
	}
	if opts.AuthorName != "" {
		updateOpts.AuthorName = gitlab.Ptr(opts.AuthorName)
	}
	if opts.StartBranch != "" {
		updateOpts.StartBranch = gitlab.Ptr(opts.StartBranch)
	}

	var fileInfo *gitlab.FileInfo
	var response *gitlab.Response

	if fileExists {
		// Update existing file
		fileInfo, response, err = fm.client.RepositoryFiles.UpdateFile(fm.projectID, filePath, updateOpts)
	} else {
		// Create new file
		createOpts := &gitlab.CreateFileOptions{
			Branch:        updateOpts.Branch,
			Content:       updateOpts.Content,
			CommitMessage: updateOpts.CommitMessage,
			AuthorEmail:   updateOpts.AuthorEmail,
			AuthorName:    updateOpts.AuthorName,
			StartBranch:   updateOpts.StartBranch,
		}
		fileInfo, response, err = fm.client.RepositoryFiles.CreateFile(fm.projectID, filePath, createOpts)
	}

	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to update file %s: %v", filePath, err))
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, errors.NewAPIError(fmt.Sprintf("unexpected status code %d when updating file", response.StatusCode))
	}

	return fileInfo, nil
}

// DeleteFile deletes a file from repository
func (fm *FileManager) DeleteFile(ctx context.Context, filePath, branch, commitMessage string) error {
	if filePath == "" {
		return errors.NewValidationError("file path cannot be empty")
	}

	if branch == "" {
		branch = DefaultBranch
	}

	if commitMessage == "" {
		commitMessage = fmt.Sprintf("Delete %s", filePath)
	}

	opts := &gitlab.DeleteFileOptions{
		Branch:        gitlab.Ptr(branch),
		CommitMessage: gitlab.Ptr(commitMessage),
	}

	_, err := fm.client.RepositoryFiles.DeleteFile(fm.projectID, filePath, opts)
	if err != nil {
		return errors.NewAPIError(fmt.Sprintf("failed to delete file %s: %v", filePath, err))
	}

	return nil
}

// FileExists checks if a file exists in the repository
func (fm *FileManager) FileExists(ctx context.Context, filePath, branch string) (bool, error) {
	_, err := fm.GetFile(ctx, filePath, branch)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetFileContent retrieves just the content of a file as a string
func (fm *FileManager) GetFileContent(ctx context.Context, filePath, branch string) (string, error) {
	fileInfo, err := fm.GetFile(ctx, filePath, branch)
	if err != nil {
		return "", err
	}
	return fileInfo.Content, nil
}

// UpdateFileContent updates file content using the existing UpdateFile method
func (fm *FileManager) UpdateFileContent(ctx context.Context, filePath string, opts *FileUpdateOptions) (*gitlab.FileInfo, error) {
	return fm.UpdateFile(ctx, filePath, opts)
}

// UpdateYAMLTag updates a tag value in a YAML file
func (fm *FileManager) UpdateYAMLTag(ctx context.Context, filePath, newTag, branch string, opts *FileUpdateOptions) (*gitlab.FileInfo, error) {
	if filePath == "" {
		return nil, errors.NewValidationError("file path cannot be empty")
	}

	if newTag == "" {
		return nil, errors.NewValidationError("new tag cannot be empty")
	}

	// Get current file content
	fileInfo, err := fm.GetFile(ctx, filePath, branch)
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %w", err)
	}

	// Update tag in YAML content (simple regex-based approach for now)
	updatedContent := updateTagInYAML(fileInfo.Content, newTag)

	if updatedContent == fileInfo.Content {
		return nil, errors.NewValidationError("no tag field found to update in YAML file")
	}

	// Prepare update options
	if opts == nil {
		opts = &FileUpdateOptions{}
	}
	opts.Content = updatedContent
	opts.Branch = branch
	if opts.CommitMessage == "" {
		opts.CommitMessage = fmt.Sprintf("Update tag to %s in %s", newTag, filePath)
	}

	// Update file
	return fm.UpdateFile(ctx, filePath, opts)
}

// updateTagInYAML updates tag value in YAML content using simple string replacement
// This is a simple implementation - in production, you'd want to use a proper YAML parser
func updateTagInYAML(content, newTag string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// Look for lines that contain "tag:" followed by a value
		if strings.Contains(line, "tag:") && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			// Simple regex replacement
			parts := strings.SplitN(line, "tag:", 2)
			if len(parts) == 2 {
				indentation := strings.Split(parts[0], "tag:")[0]
				lines[i] = fmt.Sprintf("%stag: %s", indentation, newTag)
				break
			}
		}
	}
	return strings.Join(lines, "\n")
}

// GetFileHistory retrieves the commit history for a specific file
func (fm *FileManager) GetFileHistory(ctx context.Context, filePath, branch string, maxResults int) ([]*gitlab.Commit, error) {
	if filePath == "" {
		return nil, errors.NewValidationError("file path cannot be empty")
	}

	if branch == "" {
		branch = DefaultBranch
	}

	if maxResults <= 0 {
		maxResults = 10
	}

	opts := &gitlab.ListCommitsOptions{
		RefName: gitlab.Ptr(branch),
		Path:    gitlab.Ptr(filePath),
		ListOptions: gitlab.ListOptions{
			PerPage: maxResults,
			Page:    1,
		},
	}

	commits, _, err := fm.client.Commits.ListCommits(fm.projectID, opts)
	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to get file history for %s: %v", filePath, err))
	}

	return commits, nil
}
