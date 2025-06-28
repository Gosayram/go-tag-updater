package gitlab

import (
	"testing"
	"time"
)

const (
	// Test constants as per IDEA.md requirements
	TestTimeout        = 5000 // milliseconds
	MaxTestRetries     = 3
	TestGitLabToken    = "test-token-12345"
	TestGitLabURL      = "https://gitlab.example.com"
	TestProjectID      = 123
	TestProjectPath    = "group/project"
	InvalidToken       = ""
	InvalidProjectID   = -1
	NonExistentID      = 999999
	LongProjectName    = "very-long-project-name-that-exceeds-normal-limits-and-should-be-validated-properly"
	TestUserAgent      = "go-tag-updater/test"
	DefaultTestTimeout = 30 * time.Second
	CustomTestTimeout  = 60 * time.Second
	TestRetryCount     = 5
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		baseURL     string
		expectError bool
		description string
	}{
		{
			name:        "valid token and URL",
			token:       TestGitLabToken,
			baseURL:     TestGitLabURL,
			expectError: false,
			description: "should create client with valid inputs",
		},
		{
			name:        "valid token with default URL",
			token:       TestGitLabToken,
			baseURL:     "",
			expectError: false,
			description: "should use default GitLab URL when empty",
		},
		{
			name:        "empty token",
			token:       InvalidToken,
			baseURL:     TestGitLabURL,
			expectError: true,
			description: "should reject empty token",
		},
		{
			name:        "valid token with gitlab.com",
			token:       TestGitLabToken,
			baseURL:     DefaultGitLabURL,
			expectError: false,
			description: "should work with default gitlab.com URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.token, tt.baseURL)

			if tt.expectError {
				if err == nil {
					t.Errorf("NewClient() expected error but got none")
				}
				if client != nil {
					t.Errorf("NewClient() expected nil client on error")
				}
			} else {
				if err != nil {
					t.Errorf("NewClient() unexpected error: %v", err)
				}
				if client == nil {
					t.Errorf("NewClient() expected non-nil client")
				} else {
					// Verify client properties
					if client.token != tt.token {
						t.Errorf("NewClient() token = %q, want %q", client.token, tt.token)
					}

					expectedURL := tt.baseURL
					if expectedURL == "" {
						expectedURL = DefaultGitLabURL
					}
					if client.baseURL != expectedURL {
						t.Errorf("NewClient() baseURL = %q, want %q", client.baseURL, expectedURL)
					}

					if client.timeout != DefaultTimeout {
						t.Errorf("NewClient() timeout = %v, want %v", client.timeout, DefaultTimeout)
					}

					if client.retryCount != MaxRetryAttempts {
						t.Errorf("NewClient() retryCount = %d, want %d", client.retryCount, MaxRetryAttempts)
					}
				}
			}
		})
	}
}

func TestNewClientWithConfig(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		baseURL     string
		debug       bool
		timeout     time.Duration
		retryCount  int
		expectError bool
		description string
	}{
		{
			name:        "valid config with custom settings",
			token:       TestGitLabToken,
			baseURL:     TestGitLabURL,
			debug:       true,
			timeout:     CustomTestTimeout,
			retryCount:  TestRetryCount,
			expectError: false,
			description: "should create client with custom configuration",
		},
		{
			name:        "minimal valid config",
			token:       TestGitLabToken,
			baseURL:     "",
			debug:       false,
			timeout:     DefaultTestTimeout,
			retryCount:  MaxRetryAttempts,
			expectError: false,
			description: "should work with minimal configuration",
		},
		{
			name:        "empty token with config",
			token:       InvalidToken,
			baseURL:     TestGitLabURL,
			debug:       false,
			timeout:     DefaultTestTimeout,
			retryCount:  MaxRetryAttempts,
			expectError: true,
			description: "should reject empty token even with config",
		},
		{
			name:        "zero timeout",
			token:       TestGitLabToken,
			baseURL:     TestGitLabURL,
			debug:       false,
			timeout:     0,
			retryCount:  MaxRetryAttempts,
			expectError: false,
			description: "should accept zero timeout",
		},
		{
			name:        "negative retry count",
			token:       TestGitLabToken,
			baseURL:     TestGitLabURL,
			debug:       false,
			timeout:     DefaultTestTimeout,
			retryCount:  -1,
			expectError: false,
			description: "should accept negative retry count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClientWithConfig(tt.token, tt.baseURL, tt.debug, tt.timeout, tt.retryCount)

			if tt.expectError {
				if err == nil {
					t.Errorf("NewClientWithConfig() expected error but got none")
				}
				if client != nil {
					t.Errorf("NewClientWithConfig() expected nil client on error")
				}
			} else {
				if err != nil {
					t.Errorf("NewClientWithConfig() unexpected error: %v", err)
				}
				if client == nil {
					t.Errorf("NewClientWithConfig() expected non-nil client")
				} else {
					// Verify all configuration properties
					if client.token != tt.token {
						t.Errorf("NewClientWithConfig() token = %q, want %q", client.token, tt.token)
					}

					expectedURL := tt.baseURL
					if expectedURL == "" {
						expectedURL = DefaultGitLabURL
					}
					if client.baseURL != expectedURL {
						t.Errorf("NewClientWithConfig() baseURL = %q, want %q", client.baseURL, expectedURL)
					}

					if client.debug != tt.debug {
						t.Errorf("NewClientWithConfig() debug = %v, want %v", client.debug, tt.debug)
					}

					if client.timeout != tt.timeout {
						t.Errorf("NewClientWithConfig() timeout = %v, want %v", client.timeout, tt.timeout)
					}

					if client.retryCount != tt.retryCount {
						t.Errorf("NewClientWithConfig() retryCount = %d, want %d", client.retryCount, tt.retryCount)
					}
				}
			}
		})
	}
}

func TestClient_GetBaseURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		expected    string
		description string
	}{
		{
			name:        "custom URL",
			baseURL:     TestGitLabURL,
			expected:    TestGitLabURL,
			description: "should return custom base URL",
		},
		{
			name:        "default URL",
			baseURL:     "",
			expected:    DefaultGitLabURL,
			description: "should return default URL when empty",
		},
		{
			name:        "gitlab.com URL",
			baseURL:     DefaultGitLabURL,
			expected:    DefaultGitLabURL,
			description: "should return gitlab.com URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(TestGitLabToken, tt.baseURL)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			result := client.GetBaseURL()
			if result != tt.expected {
				t.Errorf("GetBaseURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestClient_IsDebugEnabled(t *testing.T) {
	tests := []struct {
		name        string
		debug       bool
		description string
	}{
		{
			name:        "debug enabled",
			debug:       true,
			description: "should return true when debug is enabled",
		},
		{
			name:        "debug disabled",
			debug:       false,
			description: "should return false when debug is disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClientWithConfig(TestGitLabToken, TestGitLabURL, tt.debug, DefaultTestTimeout, MaxRetryAttempts)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			result := client.IsDebugEnabled()
			if result != tt.debug {
				t.Errorf("IsDebugEnabled() = %v, want %v", result, tt.debug)
			}
		})
	}
}

func TestClient_GetTimeout(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		description string
	}{
		{
			name:        "default timeout",
			timeout:     DefaultTimeout,
			description: "should return default timeout",
		},
		{
			name:        "custom timeout",
			timeout:     CustomTestTimeout,
			description: "should return custom timeout",
		},
		{
			name:        "zero timeout",
			timeout:     0,
			description: "should return zero timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClientWithConfig(TestGitLabToken, TestGitLabURL, false, tt.timeout, MaxRetryAttempts)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			result := client.GetTimeout()
			if result != tt.timeout {
				t.Errorf("GetTimeout() = %v, want %v", result, tt.timeout)
			}
		})
	}
}

func TestClient_GetRetryCount(t *testing.T) {
	tests := []struct {
		name        string
		retryCount  int
		description string
	}{
		{
			name:        "default retry count",
			retryCount:  MaxRetryAttempts,
			description: "should return default retry count",
		},
		{
			name:        "custom retry count",
			retryCount:  TestRetryCount,
			description: "should return custom retry count",
		},
		{
			name:        "zero retry count",
			retryCount:  0,
			description: "should return zero retry count",
		},
		{
			name:        "negative retry count",
			retryCount:  -1,
			description: "should return negative retry count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClientWithConfig(TestGitLabToken, TestGitLabURL, false, DefaultTestTimeout, tt.retryCount)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			result := client.GetRetryCount()
			if result != tt.retryCount {
				t.Errorf("GetRetryCount() = %d, want %d", result, tt.retryCount)
			}
		})
	}
}

func TestClient_GetGitLabClient(t *testing.T) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	gitlabClient := client.GetGitLabClient()
	if gitlabClient == nil {
		t.Error("GetGitLabClient() should return non-nil GitLab client")
	}

	// Verify it's the same instance
	if gitlabClient != client.client {
		t.Error("GetGitLabClient() should return the same client instance")
	}
}

func TestConstants(t *testing.T) {
	// Test that constants are properly defined and have reasonable values
	if DefaultGitLabURL == "" {
		t.Error("DefaultGitLabURL should not be empty")
	}

	if DefaultAPIVersion == "" {
		t.Error("DefaultAPIVersion should not be empty")
	}

	if ProjectsEndpoint == "" {
		t.Error("ProjectsEndpoint should not be empty")
	}

	if MergeRequestsEndpoint == "" {
		t.Error("MergeRequestsEndpoint should not be empty")
	}

	if DefaultTimeout <= 0 {
		t.Errorf("DefaultTimeout should be positive, got %v", DefaultTimeout)
	}

	if MaxRetryAttempts < 0 {
		t.Errorf("MaxRetryAttempts should be non-negative, got %d", MaxRetryAttempts)
	}

	if RetryDelayBase <= 0 {
		t.Errorf("RetryDelayBase should be positive, got %v", RetryDelayBase)
	}

	if DefaultRateLimitRPS <= 0 {
		t.Errorf("DefaultRateLimitRPS should be positive, got %d", DefaultRateLimitRPS)
	}

	if MaxConcurrentReqs <= 0 {
		t.Errorf("MaxConcurrentReqs should be positive, got %d", MaxConcurrentReqs)
	}

	if MaxResponseSize <= 0 {
		t.Errorf("MaxResponseSize should be positive, got %d", MaxResponseSize)
	}
}

func TestClient_GetProject_NilClient(t *testing.T) {
	// Test behavior when client is nil
	client := &Client{
		client: nil,
		token:  TestGitLabToken,
	}

	_, err := client.GetProject(TestProjectID)
	if err == nil {
		t.Error("GetProject() should return error when client is nil")
	}

	expectedMsg := "GitLab client not initialized"
	if err.Error() != expectedMsg {
		t.Errorf("GetProject() error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestClient_ResolveProjectID_NilClient(t *testing.T) {
	// Test behavior when client is nil
	client := &Client{
		client: nil,
		token:  TestGitLabToken,
	}

	_, err := client.ResolveProjectID(TestProjectPath)
	if err == nil {
		t.Error("ResolveProjectID() should return error when client is nil")
	}

	expectedMsg := "GitLab client not initialized"
	if err.Error() != expectedMsg {
		t.Errorf("ResolveProjectID() error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestClient_IsHealthy_NilClient(t *testing.T) {
	// Test behavior when client is nil
	client := &Client{
		client: nil,
		token:  TestGitLabToken,
	}

	err := client.IsHealthy()
	if err == nil {
		t.Error("IsHealthy() should return error when client is nil")
	}

	expectedMsg := "GitLab client not initialized"
	if err.Error() != expectedMsg {
		t.Errorf("IsHealthy() error = %q, want %q", err.Error(), expectedMsg)
	}
}

// Benchmark tests for performance requirements from IDEA.md
func BenchmarkNewClient(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewClient(TestGitLabToken, TestGitLabURL)
	}
}

func BenchmarkNewClientWithConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewClientWithConfig(TestGitLabToken, TestGitLabURL, false, DefaultTestTimeout, MaxRetryAttempts)
	}
}

func BenchmarkClient_GetBaseURL(b *testing.B) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	for i := 0; i < b.N; i++ {
		_ = client.GetBaseURL()
	}
}

func BenchmarkClient_IsDebugEnabled(b *testing.B) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	for i := 0; i < b.N; i++ {
		_ = client.IsDebugEnabled()
	}
}

func BenchmarkClient_GetTimeout(b *testing.B) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	for i := 0; i < b.N; i++ {
		_ = client.GetTimeout()
	}
}
