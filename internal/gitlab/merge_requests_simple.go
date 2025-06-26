// Package gitlab provides utilities for GitLab API operations
package gitlab

import (
	"context"
	"fmt"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/Gosayram/go-tag-updater/pkg/errors"
)

// SimpleMergeRequestManager handles basic GitLab merge request operations
type SimpleMergeRequestManager struct {
	client    *gitlab.Client
	projectID interface{}
}

// SimpleMergeRequestOptions contains basic options for creating merge requests
type SimpleMergeRequestOptions struct {
	Title        string
	Description  string
	SourceBranch string
	TargetBranch string
}

// NewSimpleMergeRequestManager creates a new simple merge request manager
func NewSimpleMergeRequestManager(client *gitlab.Client, projectID interface{}) *SimpleMergeRequestManager {
	return &SimpleMergeRequestManager{
		client:    client,
		projectID: projectID,
	}
}

// CreateMergeRequest creates a new merge request with basic options
func (smr *SimpleMergeRequestManager) CreateMergeRequest(ctx context.Context, opts *SimpleMergeRequestOptions) (*gitlab.MergeRequest, error) {
	if opts == nil {
		return nil, errors.NewValidationError("merge request options cannot be nil")
	}

	if opts.SourceBranch == "" {
		return nil, errors.NewValidationError("source branch cannot be empty")
	}

	if opts.TargetBranch == "" {
		return nil, errors.NewValidationError("target branch cannot be empty")
	}

	if opts.Title == "" {
		opts.Title = "Update tag via go-tag-updater"
	}

	createOpts := &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr(opts.Title),
		Description:  gitlab.Ptr(opts.Description),
		SourceBranch: gitlab.Ptr(opts.SourceBranch),
		TargetBranch: gitlab.Ptr(opts.TargetBranch),
	}

	mr, _, err := smr.client.MergeRequests.CreateMergeRequest(smr.projectID, createOpts)
	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to create merge request: %v", err))
	}

	return mr, nil
}

// GetMergeRequest retrieves merge request by IID
func (smr *SimpleMergeRequestManager) GetMergeRequest(ctx context.Context, mrIID int) (*gitlab.MergeRequest, error) {
	if mrIID <= 0 {
		return nil, errors.NewValidationError("merge request IID must be positive")
	}

	mr, _, err := smr.client.MergeRequests.GetMergeRequest(smr.projectID, mrIID, nil)
	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to get merge request %d: %v", mrIID, err))
	}

	return mr, nil
}

// ListMergeRequests lists merge requests with basic filtering
func (smr *SimpleMergeRequestManager) ListMergeRequests(ctx context.Context, state string) ([]*gitlab.BasicMergeRequest, error) {
	opts := &gitlab.ListProjectMergeRequestsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20,
			Page:    1,
		},
	}

	if state != "" {
		opts.State = gitlab.Ptr(state)
	}

	mrs, _, err := smr.client.MergeRequests.ListProjectMergeRequests(smr.projectID, opts)
	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to list merge requests: %v", err))
	}

	return mrs, nil
}
