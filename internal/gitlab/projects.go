// Package gitlab provides utilities for GitLab API operations
package gitlab

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/Gosayram/go-tag-updater/pkg/errors"
)

const (
	// Project resolution constants
	ProjectsAPIEndpoint  = "/api/v4/projects"
	MaxProjectNameLength = 255
	MinProjectIDValue    = 1
)

// ProjectManager handles GitLab project operations and resolution
type ProjectManager struct {
	client *gitlab.Client
}

// ProjectInfo contains detailed project information
type ProjectInfo struct {
	ID                int
	Name              string
	Path              string
	PathWithNamespace string
	WebURL            string
	DefaultBranch     string
	Description       string
	Visibility        string
	CreatedAt         string
	LastActivityAt    string
	ForksCount        int
	StarCount         int
}

// NewProjectManager creates a new project manager
func NewProjectManager(client *gitlab.Client) *ProjectManager {
	return &ProjectManager{
		client: client,
	}
}

// ResolveProjectIdentifier resolves project identifier to numeric ID
// Supports both numeric IDs and human-readable paths (group/subgroup/project)
func (pm *ProjectManager) ResolveProjectIdentifier(ctx context.Context, identifier string) (int, error) {
	if identifier == "" {
		return 0, errors.NewValidationError("project identifier cannot be empty")
	}

	// Check if it's already a numeric ID
	if projectID, err := strconv.Atoi(identifier); err == nil {
		if projectID < MinProjectIDValue {
			return 0, errors.NewValidationError(fmt.Sprintf("project ID must be >= %d", MinProjectIDValue))
		}

		// Validate that this project ID actually exists
		exists, err := pm.ValidateProjectExists(ctx, projectID)
		if err != nil {
			return 0, fmt.Errorf("failed to validate project ID %d: %w", projectID, err)
		}

		if !exists {
			return 0, errors.NewProjectNotFoundError(fmt.Sprintf("project with ID %d does not exist", projectID))
		}

		return projectID, nil
	}

	// It's a path-based identifier, resolve it
	return pm.ResolveProjectPath(ctx, identifier)
}

// ResolveProjectPath resolves a human-readable project path to numeric ID
// Handles paths like "group/subgroup/project" or "user/project"
func (pm *ProjectManager) ResolveProjectPath(ctx context.Context, projectPath string) (int, error) {
	if projectPath == "" {
		return 0, errors.NewValidationError("project path cannot be empty")
	}

	// Validate path format
	if err := validateProjectPath(projectPath); err != nil {
		return 0, err
	}

	// URL encode the path for API call
	encodedPath := url.PathEscape(projectPath)

	project, _, err := pm.client.Projects.GetProject(encodedPath, nil)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return 0, errors.NewProjectNotFoundError(fmt.Sprintf("project '%s' not found", projectPath))
		}
		return 0, errors.NewAPIError(fmt.Sprintf("failed to resolve project path '%s': %v", projectPath, err))
	}

	return project.ID, nil
}

// GetProjectInfo retrieves detailed project information
func (pm *ProjectManager) GetProjectInfo(ctx context.Context, projectID int) (*ProjectInfo, error) {
	if projectID < MinProjectIDValue {
		return nil, errors.NewValidationError(fmt.Sprintf("project ID must be >= %d", MinProjectIDValue))
	}

	project, _, err := pm.client.Projects.GetProject(projectID, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return nil, errors.NewProjectNotFoundError(fmt.Sprintf("project with ID %d not found", projectID))
		}
		return nil, errors.NewAPIError(fmt.Sprintf("failed to get project info for ID %d: %v", projectID, err))
	}

	return pm.convertToProjectInfo(project), nil
}

// ValidateProjectExists checks if a project exists and is accessible
func (pm *ProjectManager) ValidateProjectExists(ctx context.Context, projectID int) (bool, error) {
	if projectID < MinProjectIDValue {
		return false, errors.NewValidationError(fmt.Sprintf("project ID must be >= %d", MinProjectIDValue))
	}

	_, _, err := pm.client.Projects.GetProject(projectID, nil)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, errors.NewAPIError(fmt.Sprintf("failed to validate project %d: %v", projectID, err))
	}

	return true, nil
}

// ListUserProjects lists projects accessible to the authenticated user
func (pm *ProjectManager) ListUserProjects(ctx context.Context, maxResults int) ([]*ProjectInfo, error) {
	if maxResults <= 0 {
		maxResults = 20
	}

	opts := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: maxResults,
			Page:    1,
		},
		Membership: gitlab.Ptr(true),  // Only projects where user is a member
		Simple:     gitlab.Ptr(false), // Get full project info
	}

	projects, _, err := pm.client.Projects.ListProjects(opts)
	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to list user projects: %v", err))
	}

	var result []*ProjectInfo
	for _, project := range projects {
		result = append(result, pm.convertToProjectInfo(project))
	}

	return result, nil
}

// SearchProjects searches for projects by name or path
func (pm *ProjectManager) SearchProjects(ctx context.Context, query string, maxResults int) ([]*ProjectInfo, error) {
	if query == "" {
		return nil, errors.NewValidationError("search query cannot be empty")
	}

	if maxResults <= 0 {
		maxResults = 20
	}

	opts := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: maxResults,
			Page:    1,
		},
		Search: gitlab.Ptr(query),
		Simple: gitlab.Ptr(false),
	}

	projects, _, err := pm.client.Projects.ListProjects(opts)
	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to search projects with query '%s': %v", query, err))
	}

	var result []*ProjectInfo
	for _, project := range projects {
		result = append(result, pm.convertToProjectInfo(project))
	}

	return result, nil
}

// GetProjectDefaultBranch returns the default branch for a project
func (pm *ProjectManager) GetProjectDefaultBranch(ctx context.Context, projectID int) (string, error) {
	projectInfo, err := pm.GetProjectInfo(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to get project info: %w", err)
	}

	if projectInfo.DefaultBranch == "" {
		return "main", nil // Fallback to 'main' if not specified
	}

	return projectInfo.DefaultBranch, nil
}

// convertToProjectInfo converts GitLab Project to our ProjectInfo
func (pm *ProjectManager) convertToProjectInfo(project *gitlab.Project) *ProjectInfo {
	info := &ProjectInfo{
		ID:                project.ID,
		Name:              project.Name,
		Path:              project.Path,
		PathWithNamespace: project.PathWithNamespace,
		WebURL:            project.WebURL,
		DefaultBranch:     project.DefaultBranch,
		Description:       project.Description,
		Visibility:        string(project.Visibility),
		ForksCount:        project.ForksCount,
		StarCount:         project.StarCount,
	}

	// Handle optional time fields safely
	if project.CreatedAt != nil {
		info.CreatedAt = project.CreatedAt.String()
	}
	if project.LastActivityAt != nil {
		info.LastActivityAt = project.LastActivityAt.String()
	}

	return info
}

// validateProjectPath validates the format of a project path
func validateProjectPath(path string) error {
	if path == "" {
		return errors.NewValidationError("project path cannot be empty")
	}

	if len(path) > MaxProjectNameLength {
		return errors.NewValidationError(fmt.Sprintf("project path too long: %d characters (max %d)", len(path), MaxProjectNameLength))
	}

	// Check for basic path format (must contain at least one slash)
	if !strings.Contains(path, "/") {
		return errors.NewValidationError("project path must be in format 'group/project' or 'group/subgroup/project'")
	}

	// Check for invalid characters
	invalidChars := []string{" ", "\t", "\n", "\r", "\\", "?", "*", "<", ">", "|", "\""}
	for _, char := range invalidChars {
		if strings.Contains(path, char) {
			return errors.NewValidationError(fmt.Sprintf("project path contains invalid character: %s", char))
		}
	}

	// Check for double slashes or trailing/leading slashes
	if strings.Contains(path, "//") {
		return errors.NewValidationError("project path cannot contain double slashes")
	}

	if strings.HasPrefix(path, "/") || strings.HasSuffix(path, "/") {
		return errors.NewValidationError("project path cannot start or end with slash")
	}

	// Validate each path segment
	segments := strings.Split(path, "/")
	if len(segments) < 2 {
		return errors.NewValidationError("project path must have at least 2 segments (group/project)")
	}

	for i, segment := range segments {
		if segment == "" {
			return errors.NewValidationError(fmt.Sprintf("empty segment at position %d in project path", i))
		}

		if len(segment) > 100 { // Reasonable limit for individual segments
			return errors.NewValidationError(fmt.Sprintf("path segment too long: %s", segment))
		}
	}

	return nil
}
