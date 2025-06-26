package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const (
	// GitLab URLs
	DefaultGitLabURL = "https://gitlab.com"

	// API version and endpoints
	DefaultAPIVersion     = "v4"
	ProjectsEndpoint      = "/api/v4/projects"
	MergeRequestsEndpoint = "/api/v4/projects/%d/merge_requests"

	// Timeout and retry settings
	DefaultTimeout   = 30 * time.Second
	MaxRetryAttempts = 3
	RetryDelayBase   = 1 * time.Second

	// Rate limiting
	DefaultRateLimitRPS = 10
	MaxConcurrentReqs   = 5

	// Response size limits
	MaxResponseSize = 10 * 1024 * 1024 // 10MB
)

// Client wraps the GitLab API client with additional functionality
type Client struct {
	client     *gitlab.Client
	debug      bool
	baseURL    string
	token      string
	timeout    time.Duration
	retryCount int
}

// NewClient creates a new GitLab client instance
func NewClient(token, baseURL string) (*Client, error) {
	return NewClientWithConfig(token, baseURL, false, DefaultTimeout, MaxRetryAttempts)
}

// NewClientWithConfig creates a new GitLab client with custom configuration
func NewClientWithConfig(token, baseURL string, debug bool, timeout time.Duration, retryCount int) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("GitLab token cannot be empty")
	}

	if baseURL == "" {
		baseURL = DefaultGitLabURL
	}

	// Create GitLab client with custom HTTP client
	httpClient := &http.Client{
		Timeout: timeout,
	}

	gitlabClient, err := gitlab.NewClient(token, gitlab.WithBaseURL(baseURL), gitlab.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return &Client{
		client:     gitlabClient,
		debug:      debug,
		baseURL:    baseURL,
		token:      token,
		timeout:    timeout,
		retryCount: retryCount,
	}, nil
}

// GetProject retrieves project information by ID or path
func (c *Client) GetProject(projectID interface{}) (*gitlab.Project, error) {
	if c.client == nil {
		return nil, fmt.Errorf("GitLab client not initialized")
	}

	project, _, err := c.client.Projects.GetProject(projectID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %v: %w", projectID, err)
	}

	return project, nil
}

// ResolveProjectID converts a project path to numeric ID using ProjectManager
func (c *Client) ResolveProjectID(projectIdentifier string) (int, error) {
	if c.client == nil {
		return 0, fmt.Errorf("GitLab client not initialized")
	}

	projectManager := NewProjectManager(c.client)
	return projectManager.ResolveProjectIdentifier(context.Background(), projectIdentifier)
}

// IsHealthy checks if the GitLab instance is accessible
func (c *Client) IsHealthy() error {
	if c.client == nil {
		return fmt.Errorf("GitLab client not initialized")
	}

	// Try to get current user as a health check
	_, _, err := c.client.Users.CurrentUser()
	if err != nil {
		return fmt.Errorf("GitLab health check failed: %w", err)
	}

	return nil
}

// GetBaseURL returns the base URL of the GitLab instance
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// IsDebugEnabled returns true if debug mode is enabled
func (c *Client) IsDebugEnabled() bool {
	return c.debug
}

// GetTimeout returns the configured timeout
func (c *Client) GetTimeout() time.Duration {
	return c.timeout
}

// GetRetryCount returns the configured retry count
func (c *Client) GetRetryCount() int {
	return c.retryCount
}

// GetGitLabClient returns the underlying GitLab client
func (c *Client) GetGitLabClient() *gitlab.Client {
	return c.client
}
