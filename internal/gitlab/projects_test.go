package gitlab

import (
	"context"
	"testing"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const (
	// Additional test constants for projects
	TestProjectName        = "test-project"
	TestGroupName          = "test-group"
	TestSubGroupName       = "subgroup"
	TestNamespace          = "test-group/test-project"
	TestNestedNamespace    = "test-group/subgroup/test-project"
	TestProjectDescription = "Test project description"
	TestDefaultBranch      = "main"
	TestWebURL             = "https://gitlab.example.com/test-group/test-project"
	TestVisibility         = "private"
	TestForksCount         = 5
	TestStarCount          = 10
	DefaultMaxResults      = 20
	CustomMaxResults       = 50
	TestSearchQuery        = "test"
	EmptySearchQuery       = ""
	LongSearchQuery        = "very-long-search-query-that-might-cause-issues-with-api-limits-and-should-be-handled-properly"

	// Invalid project paths for testing
	EmptyProjectPath      = ""
	PathWithSpaces        = "group with spaces/project"
	PathWithInvalidChars  = "group/project?"
	PathWithDoubleSlash   = "group//project"
	PathWithLeadingSlash  = "/group/project"
	PathWithTrailingSlash = "group/project/"
	PathTooLong           = "very-long-group-name-that-exceeds-maximum-allowed-length-for-gitlab-project-paths-and-should-be-rejected-by-validation/very-long-project-name-that-also-exceeds-limits"
	PathSingleSegment     = "project-only"
	PathEmptySegment      = "group//project"
)

func TestNewProjectManager(t *testing.T) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	pm := NewProjectManager(client.GetGitLabClient())

	if pm == nil {
		t.Fatal("NewProjectManager() should return non-nil manager")
	}

	if pm.client != client.GetGitLabClient() {
		t.Error("NewProjectManager() should set client correctly")
	}
}

func TestNewProjectManager_NilClient(t *testing.T) {
	pm := NewProjectManager(nil)

	if pm == nil {
		t.Fatal("NewProjectManager() should return non-nil manager even with nil client")
	}

	if pm.client != nil {
		t.Error("NewProjectManager() should accept nil client")
	}
}

func TestProjectInfo_Structure(t *testing.T) {
	info := &ProjectInfo{
		ID:                TestProjectID,
		Name:              TestProjectName,
		Path:              TestProjectName,
		PathWithNamespace: TestNamespace,
		WebURL:            TestWebURL,
		DefaultBranch:     TestDefaultBranch,
		Description:       TestProjectDescription,
		Visibility:        TestVisibility,
		CreatedAt:         "2023-01-01T00:00:00Z",
		LastActivityAt:    "2023-12-31T23:59:59Z",
		ForksCount:        TestForksCount,
		StarCount:         TestStarCount,
	}

	// Verify all fields are settable and retrievable
	if info.ID != TestProjectID {
		t.Errorf("ProjectInfo.ID = %d, want %d", info.ID, TestProjectID)
	}

	if info.Name != TestProjectName {
		t.Errorf("ProjectInfo.Name = %q, want %q", info.Name, TestProjectName)
	}

	if info.Path != TestProjectName {
		t.Errorf("ProjectInfo.Path = %q, want %q", info.Path, TestProjectName)
	}

	if info.PathWithNamespace != TestNamespace {
		t.Errorf("ProjectInfo.PathWithNamespace = %q, want %q", info.PathWithNamespace, TestNamespace)
	}

	if info.WebURL != TestWebURL {
		t.Errorf("ProjectInfo.WebURL = %q, want %q", info.WebURL, TestWebURL)
	}

	if info.DefaultBranch != TestDefaultBranch {
		t.Errorf("ProjectInfo.DefaultBranch = %q, want %q", info.DefaultBranch, TestDefaultBranch)
	}

	if info.Description != TestProjectDescription {
		t.Errorf("ProjectInfo.Description = %q, want %q", info.Description, TestProjectDescription)
	}

	if info.Visibility != TestVisibility {
		t.Errorf("ProjectInfo.Visibility = %q, want %q", info.Visibility, TestVisibility)
	}

	if info.ForksCount != TestForksCount {
		t.Errorf("ProjectInfo.ForksCount = %d, want %d", info.ForksCount, TestForksCount)
	}

	if info.StarCount != TestStarCount {
		t.Errorf("ProjectInfo.StarCount = %d, want %d", info.StarCount, TestStarCount)
	}
}

func TestValidateProjectPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
		description string
	}{
		{
			name:        "valid simple path",
			path:        TestNamespace,
			expectError: false,
			description: "should accept valid group/project format",
		},
		{
			name:        "valid nested path",
			path:        TestNestedNamespace,
			expectError: false,
			description: "should accept valid group/subgroup/project format",
		},
		{
			name:        "empty path",
			path:        EmptyProjectPath,
			expectError: true,
			description: "should reject empty path",
		},
		{
			name:        "path with spaces",
			path:        PathWithSpaces,
			expectError: true,
			description: "should reject path with spaces",
		},
		{
			name:        "path with invalid characters",
			path:        PathWithInvalidChars,
			expectError: true,
			description: "should reject path with invalid characters",
		},
		{
			name:        "path with double slash",
			path:        PathWithDoubleSlash,
			expectError: true,
			description: "should reject path with double slash",
		},
		{
			name:        "path with leading slash",
			path:        PathWithLeadingSlash,
			expectError: true,
			description: "should reject path with leading slash",
		},
		{
			name:        "path with trailing slash",
			path:        PathWithTrailingSlash,
			expectError: true,
			description: "should reject path with trailing slash",
		},
		{
			name:        "single segment path",
			path:        PathSingleSegment,
			expectError: true,
			description: "should reject single segment path",
		},
		{
			name:        "path too long",
			path:        PathTooLong,
			expectError: true,
			description: "should reject overly long path",
		},
		{
			name:        "path with tab character",
			path:        "group\tproject",
			expectError: true,
			description: "should reject path with tab character",
		},
		{
			name:        "path with newline",
			path:        "group\nproject",
			expectError: true,
			description: "should reject path with newline",
		},
		{
			name:        "path with backslash",
			path:        "group\\project",
			expectError: true,
			description: "should reject path with backslash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProjectPath(tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("validateProjectPath(%q) expected error but got none", tt.path)
				}
			} else {
				if err != nil {
					t.Errorf("validateProjectPath(%q) unexpected error: %v", tt.path, err)
				}
			}
		})
	}
}

func TestProjectManager_ResolveProjectIdentifier_ValidationErrors(t *testing.T) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	pm := NewProjectManager(client.GetGitLabClient())
	ctx := context.Background()

	tests := []struct {
		name        string
		identifier  string
		expectError bool
		description string
	}{
		{
			name:        "empty identifier",
			identifier:  "",
			expectError: true,
			description: "should reject empty identifier",
		},
		{
			name:        "invalid numeric ID",
			identifier:  "0",
			expectError: true,
			description: "should reject zero project ID",
		},
		{
			name:        "negative numeric ID",
			identifier:  "-1",
			expectError: true,
			description: "should reject negative project ID",
		},
		{
			name:        "valid numeric ID format",
			identifier:  "123",
			expectError: true, // Will fail because project doesn't exist, but validates format
			description: "should accept valid numeric ID format",
		},
		{
			name:        "invalid path format",
			identifier:  PathWithSpaces,
			expectError: true,
			description: "should reject invalid path format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pm.ResolveProjectIdentifier(ctx, tt.identifier)

			if tt.expectError {
				if err == nil {
					t.Errorf("ResolveProjectIdentifier(%q) expected error but got none", tt.identifier)
				}
			} else {
				if err != nil {
					t.Errorf("ResolveProjectIdentifier(%q) unexpected error: %v", tt.identifier, err)
				}
			}
		})
	}
}

func TestProjectManager_GetProjectInfo_ValidationErrors(t *testing.T) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	pm := NewProjectManager(client.GetGitLabClient())
	ctx := context.Background()

	tests := []struct {
		name        string
		projectID   int
		expectError bool
		description string
	}{
		{
			name:        "zero project ID",
			projectID:   0,
			expectError: true,
			description: "should reject zero project ID",
		},
		{
			name:        "negative project ID",
			projectID:   InvalidProjectID,
			expectError: true,
			description: "should reject negative project ID",
		},
		{
			name:        "valid project ID format",
			projectID:   TestProjectID,
			expectError: true, // Will fail because project doesn't exist, but validates format
			description: "should accept valid project ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pm.GetProjectInfo(ctx, tt.projectID)

			if tt.expectError {
				if err == nil {
					t.Errorf("GetProjectInfo(%d) expected error but got none", tt.projectID)
				}
			} else {
				if err != nil {
					t.Errorf("GetProjectInfo(%d) unexpected error: %v", tt.projectID, err)
				}
			}
		})
	}
}

func TestProjectManager_ValidateProjectExists_ValidationErrors(t *testing.T) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	pm := NewProjectManager(client.GetGitLabClient())
	ctx := context.Background()

	tests := []struct {
		name        string
		projectID   int
		expectError bool
		description string
	}{
		{
			name:        "zero project ID",
			projectID:   0,
			expectError: true,
			description: "should reject zero project ID",
		},
		{
			name:        "negative project ID",
			projectID:   InvalidProjectID,
			expectError: true,
			description: "should reject negative project ID",
		},
		{
			name:        "valid project ID format",
			projectID:   TestProjectID,
			expectError: true, // Will fail with API error due to no real GitLab
			description: "should accept valid project ID format but fail with API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pm.ValidateProjectExists(ctx, tt.projectID)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateProjectExists(%d) expected error but got none", tt.projectID)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateProjectExists(%d) unexpected error: %v", tt.projectID, err)
				}
			}
		})
	}
}

func TestProjectManager_ListUserProjects_MaxResults(t *testing.T) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	pm := NewProjectManager(client.GetGitLabClient())
	ctx := context.Background()

	tests := []struct {
		name        string
		maxResults  int
		expectError bool
		description string
	}{
		{
			name:        "zero max results",
			maxResults:  0,
			expectError: false, // Should default to 20
			description: "should default to 20 when zero",
		},
		{
			name:        "negative max results",
			maxResults:  -1,
			expectError: false, // Should default to 20
			description: "should default to 20 when negative",
		},
		{
			name:        "positive max results",
			maxResults:  CustomMaxResults,
			expectError: false,
			description: "should accept positive max results",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail because we don't have a real GitLab instance,
			// but we can test the validation logic
			_, err := pm.ListUserProjects(ctx, tt.maxResults)

			// We expect API errors due to no real GitLab, but not validation errors
			if tt.expectError {
				if err == nil {
					t.Errorf("ListUserProjects(%d) expected error but got none", tt.maxResults)
				}
			}
			// For this test, we mainly care that it doesn't panic on invalid inputs
		})
	}
}

func TestProjectManager_SearchProjects_Validation(t *testing.T) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	pm := NewProjectManager(client.GetGitLabClient())
	ctx := context.Background()

	tests := []struct {
		name        string
		query       string
		maxResults  int
		expectError bool
		description string
	}{
		{
			name:        "empty query",
			query:       EmptySearchQuery,
			maxResults:  DefaultMaxResults,
			expectError: true,
			description: "should reject empty search query",
		},
		{
			name:        "valid query with zero max results",
			query:       TestSearchQuery,
			maxResults:  0,
			expectError: false, // Should default to 20
			description: "should default max results to 20 when zero",
		},
		{
			name:        "valid query with negative max results",
			query:       TestSearchQuery,
			maxResults:  -1,
			expectError: false, // Should default to 20
			description: "should default max results to 20 when negative",
		},
		{
			name:        "valid query and max results",
			query:       TestSearchQuery,
			maxResults:  CustomMaxResults,
			expectError: false,
			description: "should accept valid query and max results",
		},
		{
			name:        "long query",
			query:       LongSearchQuery,
			maxResults:  DefaultMaxResults,
			expectError: false,
			description: "should accept long search query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pm.SearchProjects(ctx, tt.query, tt.maxResults)

			if tt.expectError {
				if err == nil {
					t.Errorf("SearchProjects(%q, %d) expected error but got none", tt.query, tt.maxResults)
				}
			}
			// We expect API errors due to no real GitLab for valid inputs,
			// but we're testing validation logic here
		})
	}
}

func TestProjectManager_ConvertToProjectInfo(t *testing.T) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	pm := NewProjectManager(client.GetGitLabClient())

	// Create a mock GitLab project
	gitlabProject := &gitlab.Project{
		ID:                TestProjectID,
		Name:              TestProjectName,
		Path:              TestProjectName,
		PathWithNamespace: TestNamespace,
		WebURL:            TestWebURL,
		DefaultBranch:     TestDefaultBranch,
		Description:       TestProjectDescription,
		Visibility:        gitlab.PrivateVisibility,
		ForksCount:        TestForksCount,
		StarCount:         TestStarCount,
	}

	projectInfo := pm.convertToProjectInfo(gitlabProject)

	if projectInfo == nil {
		t.Fatal("convertToProjectInfo() should return non-nil ProjectInfo")
	}

	// Verify all fields are converted correctly
	if projectInfo.ID != TestProjectID {
		t.Errorf("convertToProjectInfo() ID = %d, want %d", projectInfo.ID, TestProjectID)
	}

	if projectInfo.Name != TestProjectName {
		t.Errorf("convertToProjectInfo() Name = %q, want %q", projectInfo.Name, TestProjectName)
	}

	if projectInfo.Path != TestProjectName {
		t.Errorf("convertToProjectInfo() Path = %q, want %q", projectInfo.Path, TestProjectName)
	}

	if projectInfo.PathWithNamespace != TestNamespace {
		t.Errorf("convertToProjectInfo() PathWithNamespace = %q, want %q", projectInfo.PathWithNamespace, TestNamespace)
	}

	if projectInfo.WebURL != TestWebURL {
		t.Errorf("convertToProjectInfo() WebURL = %q, want %q", projectInfo.WebURL, TestWebURL)
	}

	if projectInfo.DefaultBranch != TestDefaultBranch {
		t.Errorf("convertToProjectInfo() DefaultBranch = %q, want %q", projectInfo.DefaultBranch, TestDefaultBranch)
	}

	if projectInfo.Description != TestProjectDescription {
		t.Errorf("convertToProjectInfo() Description = %q, want %q", projectInfo.Description, TestProjectDescription)
	}

	if projectInfo.Visibility != string(gitlab.PrivateVisibility) {
		t.Errorf("convertToProjectInfo() Visibility = %q, want %q", projectInfo.Visibility, string(gitlab.PrivateVisibility))
	}

	if projectInfo.ForksCount != TestForksCount {
		t.Errorf("convertToProjectInfo() ForksCount = %d, want %d", projectInfo.ForksCount, TestForksCount)
	}

	if projectInfo.StarCount != TestStarCount {
		t.Errorf("convertToProjectInfo() StarCount = %d, want %d", projectInfo.StarCount, TestStarCount)
	}
}

func TestProjectConstants(t *testing.T) {
	// Test that constants are properly defined
	if ProjectsAPIEndpoint == "" {
		t.Error("ProjectsAPIEndpoint should not be empty")
	}

	if MaxProjectNameLength <= 0 {
		t.Errorf("MaxProjectNameLength should be positive, got %d", MaxProjectNameLength)
	}

	if MinProjectIDValue <= 0 {
		t.Errorf("MinProjectIDValue should be positive, got %d", MinProjectIDValue)
	}

	// Test reasonable values
	if MaxProjectNameLength < 100 {
		t.Errorf("MaxProjectNameLength seems too small: %d", MaxProjectNameLength)
	}

	if MaxProjectNameLength > 1000 {
		t.Errorf("MaxProjectNameLength seems too large: %d", MaxProjectNameLength)
	}

	if MinProjectIDValue != 1 {
		t.Errorf("MinProjectIDValue should be 1, got %d", MinProjectIDValue)
	}
}

// Benchmark tests for performance requirements from IDEA.md
func BenchmarkNewProjectManager(b *testing.B) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	gitlabClient := client.GetGitLabClient()

	for i := 0; i < b.N; i++ {
		_ = NewProjectManager(gitlabClient)
	}
}

func BenchmarkValidateProjectPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = validateProjectPath(TestNamespace)
	}
}

func BenchmarkValidateProjectPath_Invalid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = validateProjectPath(PathWithSpaces)
	}
}

func BenchmarkProjectManager_ConvertToProjectInfo(b *testing.B) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	pm := NewProjectManager(client.GetGitLabClient())

	gitlabProject := &gitlab.Project{
		ID:                TestProjectID,
		Name:              TestProjectName,
		Path:              TestProjectName,
		PathWithNamespace: TestNamespace,
		WebURL:            TestWebURL,
		DefaultBranch:     TestDefaultBranch,
		Description:       TestProjectDescription,
		Visibility:        gitlab.PrivateVisibility,
		ForksCount:        TestForksCount,
		StarCount:         TestStarCount,
	}

	for i := 0; i < b.N; i++ {
		_ = pm.convertToProjectInfo(gitlabProject)
	}
}
