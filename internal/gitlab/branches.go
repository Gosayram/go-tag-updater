// Package gitlab provides utilities for GitLab API operations
package gitlab

import (
	"context"
	"fmt"
	"strings"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/Gosayram/go-tag-updater/pkg/errors"
)

const (
	// Default branch prefixes
	DefaultBranchPrefix = "feature/"
	UpdateBranchPrefix  = "update-tag/"

	// Branch name constraints
	MaxBranchNameLength = 100
	MinBranchNameLength = 1
)

// BranchManager handles GitLab branch operations
type BranchManager struct {
	client    *gitlab.Client
	projectID interface{}
}

// BranchInfo represents branch information
type BranchInfo struct {
	Name               string
	Protected          bool
	Default            bool
	DevelopersCanPush  bool
	DevelopersCanMerge bool
	Commit             *gitlab.Commit
	WebURL             string
}

// NewBranchManager creates a new branch manager
func NewBranchManager(client *gitlab.Client, projectID interface{}) *BranchManager {
	return &BranchManager{
		client:    client,
		projectID: projectID,
	}
}

// CreateBranch creates a new branch from a reference
func (bm *BranchManager) CreateBranch(ctx context.Context, branchName, ref string) (*BranchInfo, error) {
	if branchName == "" {
		return nil, errors.NewValidationError("branch name cannot be empty")
	}

	if ref == "" {
		ref = "main" // Default to main branch
	}

	if len(branchName) > MaxBranchNameLength {
		return nil, errors.NewValidationError(fmt.Sprintf("branch name too long: %d characters (max %d)", len(branchName), MaxBranchNameLength))
	}

	// Validate branch name format
	if err := validateBranchName(branchName); err != nil {
		return nil, err
	}

	opts := &gitlab.CreateBranchOptions{
		Branch: gitlab.Ptr(branchName),
		Ref:    gitlab.Ptr(ref),
	}

	branch, _, err := bm.client.Branches.CreateBranch(bm.projectID, opts)
	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to create branch %s: %v", branchName, err))
	}

	return bm.convertToBranchInfo(branch), nil
}

// GetBranch retrieves branch information
func (bm *BranchManager) GetBranch(ctx context.Context, branchName string) (*BranchInfo, error) {
	if branchName == "" {
		return nil, errors.NewValidationError("branch name cannot be empty")
	}

	branch, _, err := bm.client.Branches.GetBranch(bm.projectID, branchName)
	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to get branch %s: %v", branchName, err))
	}

	return bm.convertToBranchInfo(branch), nil
}

// ListBranches lists branches with optional search filter
func (bm *BranchManager) ListBranches(ctx context.Context, search string, maxResults int) ([]*BranchInfo, error) {
	if maxResults <= 0 {
		maxResults = 20
	}

	opts := &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: maxResults,
			Page:    1,
		},
	}

	if search != "" {
		opts.Search = gitlab.Ptr(search)
	}

	branches, _, err := bm.client.Branches.ListBranches(bm.projectID, opts)
	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to list branches: %v", err))
	}

	var result []*BranchInfo
	for _, branch := range branches {
		result = append(result, bm.convertToBranchInfo(branch))
	}

	return result, nil
}

// DeleteBranch deletes a branch
func (bm *BranchManager) DeleteBranch(ctx context.Context, branchName string) error {
	if branchName == "" {
		return errors.NewValidationError("branch name cannot be empty")
	}

	// Check if branch is protected
	branch, err := bm.GetBranch(ctx, branchName)
	if err != nil {
		return fmt.Errorf("failed to get branch info before deletion: %w", err)
	}

	if branch.Protected {
		return errors.NewValidationError(fmt.Sprintf("cannot delete protected branch: %s", branchName))
	}

	_, err = bm.client.Branches.DeleteBranch(bm.projectID, branchName)
	if err != nil {
		return errors.NewAPIError(fmt.Sprintf("failed to delete branch %s: %v", branchName, err))
	}

	return nil
}

// BranchExists checks if a branch exists
func (bm *BranchManager) BranchExists(ctx context.Context, branchName string) (bool, error) {
	_, err := bm.GetBranch(ctx, branchName)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetProtectedBranches lists protected branches
func (bm *BranchManager) GetProtectedBranches(ctx context.Context) ([]*gitlab.ProtectedBranch, error) {
	branches, _, err := bm.client.ProtectedBranches.ListProtectedBranches(bm.projectID, nil)
	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to list protected branches: %v", err))
	}
	return branches, nil
}

// GenerateUniqueBranchName generates a unique branch name with timestamp
func (bm *BranchManager) GenerateUniqueBranchName(ctx context.Context, prefix, tag string) (string, error) {
	if prefix == "" {
		prefix = UpdateBranchPrefix
	}

	// Clean the tag for use in branch name
	cleanTag := strings.ReplaceAll(tag, "/", "-")
	cleanTag = strings.ReplaceAll(cleanTag, ":", "-")

	timestamp := time.Now().Format("20060102-150405")
	baseName := fmt.Sprintf("%s%s-%s", prefix, cleanTag, timestamp)

	// Ensure the name doesn't exceed maximum length
	if len(baseName) > MaxBranchNameLength {
		// Truncate if too long
		maxTagLength := MaxBranchNameLength - len(prefix) - len(timestamp) - 2 // 2 for dashes
		if maxTagLength > 0 {
			cleanTag = cleanTag[:maxTagLength]
			baseName = fmt.Sprintf("%s%s-%s", prefix, cleanTag, timestamp)
		} else {
			baseName = fmt.Sprintf("%s%s", prefix, timestamp)
		}
	}

	// Check if branch already exists (unlikely with timestamp, but just in case)
	exists, err := bm.BranchExists(ctx, baseName)
	if err != nil {
		return "", fmt.Errorf("failed to check branch existence: %w", err)
	}

	if exists {
		// Add additional suffix if somehow it still exists
		baseName = fmt.Sprintf("%s-alt", baseName)
	}

	return baseName, nil
}

// FindBranchesByTag finds branches that might be related to a specific tag
func (bm *BranchManager) FindBranchesByTag(ctx context.Context, tag string) ([]*BranchInfo, error) {
	if tag == "" {
		return nil, errors.NewValidationError("tag cannot be empty")
	}

	// Search for branches containing the tag
	cleanTag := strings.ReplaceAll(tag, ".", "-")
	cleanTag = strings.ReplaceAll(cleanTag, ":", "-")

	return bm.ListBranches(ctx, cleanTag, 50)
}

// convertToBranchInfo converts GitLab Branch to our BranchInfo
func (bm *BranchManager) convertToBranchInfo(branch *gitlab.Branch) *BranchInfo {
	return &BranchInfo{
		Name:               branch.Name,
		Protected:          branch.Protected,
		Default:            branch.Default,
		DevelopersCanPush:  branch.DevelopersCanPush,
		DevelopersCanMerge: branch.DevelopersCanMerge,
		Commit:             branch.Commit,
		WebURL:             branch.WebURL,
	}
}

// validateBranchName validates branch name according to Git rules
func validateBranchName(name string) error {
	if name == "" {
		return errors.NewValidationError("branch name cannot be empty")
	}

	if len(name) < MinBranchNameLength {
		return errors.NewValidationError(fmt.Sprintf("branch name too short: %d characters (min %d)", len(name), MinBranchNameLength))
	}

	// Check for invalid characters
	invalidChars := []string{" ", "..", "~", "^", ":", "?", "*", "[", "\\", "\t", "\n", "\r"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return errors.NewValidationError(fmt.Sprintf("branch name contains invalid character: %s", char))
		}
	}

	// Check for invalid patterns
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return errors.NewValidationError("branch name cannot start or end with dash")
	}

	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return errors.NewValidationError("branch name cannot start or end with period")
	}

	if strings.HasSuffix(name, ".lock") {
		return errors.NewValidationError("branch name cannot end with .lock")
	}

	return nil
}
