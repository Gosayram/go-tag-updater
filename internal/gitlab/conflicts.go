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
	// Conflict detection constants
	MaxConflictCheckAttempts = 15
	ConflictCheckInterval    = 10 * time.Second
	DefaultWaitTimeout       = 15 * time.Minute

	// MR states for conflict checking
	StateOpened = "opened"
	StateMerged = "merged"
	StateClosed = "closed"
)

// ConflictDetector handles merge request conflict detection and prevention
type ConflictDetector struct {
	client    *gitlab.Client
	projectID interface{}
}

// ConflictInfo contains information about conflicting merge requests
type ConflictInfo struct {
	ConflictingMRs []ConflictingMR
	TotalConflicts int
	Recommendation string
}

// ConflictingMR represents a merge request that conflicts with our operation
type ConflictingMR struct {
	ID           int
	IID          int
	Title        string
	SourceBranch string
	TargetBranch string
	State        string
	WebURL       string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
	Author       string
	ConflictType string // "same_branch", "same_file", "same_target"
}

// NewConflictDetector creates a new conflict detector
func NewConflictDetector(client *gitlab.Client, projectID interface{}) *ConflictDetector {
	return &ConflictDetector{
		client:    client,
		projectID: projectID,
	}
}

// CheckForConflicts performs comprehensive conflict detection
func (cd *ConflictDetector) CheckForConflicts(ctx context.Context, sourceBranch, targetBranch, filePath string) (*ConflictInfo, error) {
	if sourceBranch == "" {
		return nil, errors.NewValidationError("source branch cannot be empty")
	}

	if targetBranch == "" {
		return nil, errors.NewValidationError("target branch cannot be empty")
	}

	conflictInfo := &ConflictInfo{
		ConflictingMRs: []ConflictingMR{},
		TotalConflicts: 0,
	}

	// Check for same source branch conflicts
	sameBranchConflicts, err := cd.checkSameBranchConflicts(ctx, sourceBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to check same branch conflicts: %w", err)
	}
	conflictInfo.ConflictingMRs = append(conflictInfo.ConflictingMRs, sameBranchConflicts...)

	// Check for same file conflicts
	if filePath != "" {
		sameFileConflicts, err := cd.checkSameFileConflicts(ctx, targetBranch, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to check same file conflicts: %w", err)
		}
		conflictInfo.ConflictingMRs = append(conflictInfo.ConflictingMRs, sameFileConflicts...)
	}

	// Check for same target branch conflicts with similar changes
	targetBranchConflicts, err := cd.checkTargetBranchConflicts(ctx, targetBranch, sourceBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to check target branch conflicts: %w", err)
	}
	conflictInfo.ConflictingMRs = append(conflictInfo.ConflictingMRs, targetBranchConflicts...)

	conflictInfo.TotalConflicts = len(conflictInfo.ConflictingMRs)
	conflictInfo.Recommendation = cd.generateRecommendation(conflictInfo)

	return conflictInfo, nil
}

// checkSameBranchConflicts checks for existing MRs with the same source branch
func (cd *ConflictDetector) checkSameBranchConflicts(ctx context.Context, sourceBranch string) ([]ConflictingMR, error) {
	opts := &gitlab.ListProjectMergeRequestsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 50,
			Page:    1,
		},
		SourceBranch: gitlab.Ptr(sourceBranch),
		State:        gitlab.Ptr(StateOpened),
	}

	mrs, _, err := cd.client.MergeRequests.ListProjectMergeRequests(cd.projectID, opts)
	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to list merge requests for branch %s: %v", sourceBranch, err))
	}

	var conflicts []ConflictingMR
	for _, mr := range mrs {
		conflicts = append(conflicts, ConflictingMR{
			ID:           mr.ID,
			IID:          mr.IID,
			Title:        mr.Title,
			SourceBranch: mr.SourceBranch,
			TargetBranch: mr.TargetBranch,
			State:        mr.State,
			WebURL:       mr.WebURL,
			CreatedAt:    mr.CreatedAt,
			UpdatedAt:    mr.UpdatedAt,
			Author:       getAuthorName(mr.Author),
			ConflictType: "same_branch",
		})
	}

	return conflicts, nil
}

// checkSameFileConflicts checks for MRs affecting the same file
func (cd *ConflictDetector) checkSameFileConflicts(ctx context.Context, targetBranch, filePath string) ([]ConflictingMR, error) {
	opts := &gitlab.ListProjectMergeRequestsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 50,
			Page:    1,
		},
		TargetBranch: gitlab.Ptr(targetBranch),
		State:        gitlab.Ptr(StateOpened),
	}

	mrs, _, err := cd.client.MergeRequests.ListProjectMergeRequests(cd.projectID, opts)
	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to list merge requests for target branch %s: %v", targetBranch, err))
	}

	var conflicts []ConflictingMR
	for _, mr := range mrs {
		// Check if this MR affects the same file
		affectsFile, err := cd.checkMRAffectsFile(ctx, mr.IID, filePath)
		if err != nil {
			// Log warning but continue checking other MRs
			continue
		}

		if affectsFile {
			conflicts = append(conflicts, ConflictingMR{
				ID:           mr.ID,
				IID:          mr.IID,
				Title:        mr.Title,
				SourceBranch: mr.SourceBranch,
				TargetBranch: mr.TargetBranch,
				State:        mr.State,
				WebURL:       mr.WebURL,
				CreatedAt:    mr.CreatedAt,
				UpdatedAt:    mr.UpdatedAt,
				Author:       getAuthorName(mr.Author),
				ConflictType: "same_file",
			})
		}
	}

	return conflicts, nil
}

// checkTargetBranchConflicts checks for recent MRs to the same target branch
func (cd *ConflictDetector) checkTargetBranchConflicts(ctx context.Context, targetBranch, excludeSourceBranch string) ([]ConflictingMR, error) {
	opts := &gitlab.ListProjectMergeRequestsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20, // Check recent ones only
			Page:    1,
		},
		TargetBranch: gitlab.Ptr(targetBranch),
		State:        gitlab.Ptr(StateOpened),
		OrderBy:      gitlab.Ptr("updated_at"),
		Sort:         gitlab.Ptr("desc"),
	}

	mrs, _, err := cd.client.MergeRequests.ListProjectMergeRequests(cd.projectID, opts)
	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to list merge requests for target branch %s: %v", targetBranch, err))
	}

	var conflicts []ConflictingMR
	for _, mr := range mrs {
		// Skip our own source branch
		if mr.SourceBranch == excludeSourceBranch {
			continue
		}

		// Check if it's a potentially conflicting change (e.g., tag updates)
		if cd.isPotentiallyConflicting(mr.Title, mr.SourceBranch) {
			conflicts = append(conflicts, ConflictingMR{
				ID:           mr.ID,
				IID:          mr.IID,
				Title:        mr.Title,
				SourceBranch: mr.SourceBranch,
				TargetBranch: mr.TargetBranch,
				State:        mr.State,
				WebURL:       mr.WebURL,
				CreatedAt:    mr.CreatedAt,
				UpdatedAt:    mr.UpdatedAt,
				Author:       getAuthorName(mr.Author),
				ConflictType: "same_target",
			})
		}
	}

	return conflicts, nil
}

// checkMRAffectsFile checks if a merge request affects a specific file
func (cd *ConflictDetector) checkMRAffectsFile(ctx context.Context, mrIID int, filePath string) (bool, error) {
	// Note: This is a simplified implementation
	// In a real implementation, you would use the MR changes API
	// For now, we'll just return false to avoid API calls that might not be available
	// TODO: Implement proper file change detection when the API supports it
	return false, nil
}

// isPotentiallyConflicting checks if an MR might conflict based on title/branch patterns
func (cd *ConflictDetector) isPotentiallyConflicting(title, branch string) bool {
	// Look for patterns that suggest tag updates or similar changes
	conflictPatterns := []string{
		"update",
		"tag",
		"version",
		"release",
		"deployment",
		"config",
	}

	titleLower := strings.ToLower(title)
	branchLower := strings.ToLower(branch)

	for _, pattern := range conflictPatterns {
		if strings.Contains(titleLower, pattern) || strings.Contains(branchLower, pattern) {
			return true
		}
	}

	return false
}

// WaitForConflictsToResolve waits for conflicting merge requests to be resolved
func (cd *ConflictDetector) WaitForConflictsToResolve(ctx context.Context, conflicts *ConflictInfo, maxWaitTime time.Duration) error {
	if conflicts.TotalConflicts == 0 {
		return nil // No conflicts to wait for
	}

	if maxWaitTime <= 0 {
		maxWaitTime = DefaultWaitTimeout
	}

	timeout := time.After(maxWaitTime)
	ticker := time.NewTicker(ConflictCheckInterval)
	defer ticker.Stop()

	attempt := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return errors.NewAPIError(fmt.Sprintf("timeout waiting for %d conflicts to resolve after %v", conflicts.TotalConflicts, maxWaitTime))
		case <-ticker.C:
			attempt++

			// Re-check conflicts
			stillConflicting := 0
			for _, conflict := range conflicts.ConflictingMRs {
				mr, _, err := cd.client.MergeRequests.GetMergeRequest(cd.projectID, conflict.IID, nil)
				if err != nil {
					continue // MR might have been deleted, which is good
				}

				if mr.State == StateOpened {
					stillConflicting++
				}
			}

			if stillConflicting == 0 {
				return nil // All conflicts resolved
			}

			if attempt >= MaxConflictCheckAttempts {
				return errors.NewAPIError(fmt.Sprintf("gave up waiting for conflicts to resolve after %d attempts", attempt))
			}
		}
	}
}

// generateRecommendation generates a recommendation based on conflict analysis
func (cd *ConflictDetector) generateRecommendation(conflictInfo *ConflictInfo) string {
	if conflictInfo.TotalConflicts == 0 {
		return "No conflicts detected. Safe to proceed."
	}

	recommendation := fmt.Sprintf("Found %d potential conflicts:\n", conflictInfo.TotalConflicts)

	sameBranchCount := 0
	sameFileCount := 0
	sameTargetCount := 0

	for _, conflict := range conflictInfo.ConflictingMRs {
		switch conflict.ConflictType {
		case "same_branch":
			sameBranchCount++
		case "same_file":
			sameFileCount++
		case "same_target":
			sameTargetCount++
		}
	}

	if sameBranchCount > 0 {
		recommendation += fmt.Sprintf("- %d MR(s) using the same source branch (high conflict risk)\n", sameBranchCount)
	}
	if sameFileCount > 0 {
		recommendation += fmt.Sprintf("- %d MR(s) affecting the same file (medium conflict risk)\n", sameFileCount)
	}
	if sameTargetCount > 0 {
		recommendation += fmt.Sprintf("- %d MR(s) with similar changes to same target (low conflict risk)\n", sameTargetCount)
	}

	if sameBranchCount > 0 {
		recommendation += "\nRecommendation: Wait for same-branch MRs to be merged or use a different branch name."
	} else if sameFileCount > 0 {
		recommendation += "\nRecommendation: Consider waiting for file-affecting MRs to be merged to avoid conflicts."
	} else {
		recommendation += "\nRecommendation: Conflicts are low-risk. You may proceed with caution."
	}

	return recommendation
}

// getAuthorName safely extracts author name from GitLab user
func getAuthorName(author *gitlab.BasicUser) string {
	if author == nil {
		return "unknown"
	}
	if author.Name != "" {
		return author.Name
	}
	if author.Username != "" {
		return author.Username
	}
	return "unknown"
}
