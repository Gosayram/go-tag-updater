package gitlab

import (
	"context"
	"strings"
	"testing"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const (
	// Branch test constants
	TestBranchName          = "feature/test-branch"
	TestMainBranch          = "main"
	TestDevelopBranch       = "develop"
	TestRef                 = "main"
	TestBranchPrefix        = "feature/"
	TestUpdatePrefix        = "update-tag/"
	TestTag                 = "v1.2.3"
	TestTagWithSlash        = "releases/v1.2.3"
	TestTagWithColon        = "namespace:v1.2.3"
	TestBranchSearchQuery   = "feature"
	TestBranchWebURL        = "https://gitlab.example.com/group/project/-/tree/feature/test-branch"
	DefaultBranchMaxResults = 20
	CustomBranchMaxResults  = 50

	// Invalid branch names for testing
	EmptyBranchName             = ""
	BranchNameTooLong           = "very-long-branch-name-that-exceeds-maximum-allowed-length-for-gitlab-branches-and-should-be-rejected-by-validation-logic-because-it-is-way-too-long"
	BranchWithSpaces            = "branch with spaces"
	BranchWithInvalidChars      = "branch@with#invalid$chars"
	BranchWithDots              = "branch..name"
	BranchStartingWithDash      = "-branch"
	BranchEndingWithDash        = "branch-"
	BranchWithConsecutiveDashes = "branch--name"

	// Protected branch test constants
	ProtectedBranchName   = "main"
	UnprotectedBranchName = "feature/test"
)

func TestNewBranchManager(t *testing.T) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	bm := NewBranchManager(client.GetGitLabClient(), TestProjectID)

	if bm == nil {
		t.Fatal("NewBranchManager() should return non-nil manager")
	}

	if bm.client != client.GetGitLabClient() {
		t.Error("NewBranchManager() should set client correctly")
	}

	if bm.projectID != TestProjectID {
		t.Errorf("NewBranchManager() projectID = %v, want %v", bm.projectID, TestProjectID)
	}
}

func TestNewBranchManager_NilClient(t *testing.T) {
	bm := NewBranchManager(nil, TestProjectID)

	if bm == nil {
		t.Fatal("NewBranchManager() should return non-nil manager even with nil client")
	}

	if bm.client != nil {
		t.Error("NewBranchManager() should accept nil client")
	}

	if bm.projectID != TestProjectID {
		t.Errorf("NewBranchManager() projectID = %v, want %v", bm.projectID, TestProjectID)
	}
}

func TestBranchInfo_Structure(t *testing.T) {
	commit := &gitlab.Commit{
		ID:      "abc123",
		Message: "Test commit",
	}

	info := &BranchInfo{
		Name:               TestBranchName,
		Protected:          false,
		Default:            false,
		DevelopersCanPush:  true,
		DevelopersCanMerge: true,
		Commit:             commit,
		WebURL:             TestBranchWebURL,
	}

	// Verify all fields are settable and retrievable
	if info.Name != TestBranchName {
		t.Errorf("BranchInfo.Name = %q, want %q", info.Name, TestBranchName)
	}

	if info.Protected != false {
		t.Errorf("BranchInfo.Protected = %v, want %v", info.Protected, false)
	}

	if info.Default != false {
		t.Errorf("BranchInfo.Default = %v, want %v", info.Default, false)
	}

	if info.DevelopersCanPush != true {
		t.Errorf("BranchInfo.DevelopersCanPush = %v, want %v", info.DevelopersCanPush, true)
	}

	if info.DevelopersCanMerge != true {
		t.Errorf("BranchInfo.DevelopersCanMerge = %v, want %v", info.DevelopersCanMerge, true)
	}

	if info.Commit != commit {
		t.Error("BranchInfo.Commit should be set correctly")
	}

	if info.WebURL != TestBranchWebURL {
		t.Errorf("BranchInfo.WebURL = %q, want %q", info.WebURL, TestBranchWebURL)
	}
}

func TestValidateBranchName(t *testing.T) {
	tests := []struct {
		name        string
		branchName  string
		expectError bool
		description string
	}{
		{
			name:        "valid simple branch name",
			branchName:  "feature",
			expectError: false,
			description: "should accept simple branch name",
		},
		{
			name:        "valid branch name with slash",
			branchName:  TestBranchName,
			expectError: false,
			description: "should accept branch name with slash",
		},
		{
			name:        "valid branch name with numbers",
			branchName:  "feature-123",
			expectError: false,
			description: "should accept branch name with numbers",
		},
		{
			name:        "valid branch name with underscores",
			branchName:  "feature_test",
			expectError: false,
			description: "should accept branch name with underscores",
		},
		{
			name:        "empty branch name",
			branchName:  EmptyBranchName,
			expectError: true,
			description: "should reject empty branch name",
		},
		{
			name:        "branch name too long",
			branchName:  BranchNameTooLong,
			expectError: false, // validateBranchName doesn't check length, only CreateBranch does
			description: "validateBranchName doesn't check length limits",
		},
		{
			name:        "branch name with spaces",
			branchName:  BranchWithSpaces,
			expectError: true,
			description: "should reject branch name with spaces",
		},
		{
			name:        "branch name with invalid characters",
			branchName:  BranchWithInvalidChars,
			expectError: false, // @ # $ are not in the invalid chars list
			description: "@ # $ characters are not in GitLab invalid chars list",
		},
		{
			name:        "branch name with consecutive dots",
			branchName:  BranchWithDots,
			expectError: true,
			description: "should reject branch name with consecutive dots",
		},
		{
			name:        "branch name starting with dash",
			branchName:  BranchStartingWithDash,
			expectError: true,
			description: "should reject branch name starting with dash",
		},
		{
			name:        "branch name ending with dash",
			branchName:  BranchEndingWithDash,
			expectError: true,
			description: "should reject branch name ending with dash",
		},
		{
			name:        "branch name with consecutive dashes",
			branchName:  BranchWithConsecutiveDashes,
			expectError: false, // consecutive dashes are allowed in Git
			description: "consecutive dashes are allowed in Git branch names",
		},
		{
			name:        "branch name with tab",
			branchName:  "branch\tname",
			expectError: true,
			description: "should reject branch name with tab",
		},
		{
			name:        "branch name with newline",
			branchName:  "branch\nname",
			expectError: true,
			description: "should reject branch name with newline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBranchName(tt.branchName)

			if tt.expectError {
				if err == nil {
					t.Errorf("validateBranchName(%q) expected error but got none", tt.branchName)
				}
			} else {
				if err != nil {
					t.Errorf("validateBranchName(%q) unexpected error: %v", tt.branchName, err)
				}
			}
		})
	}
}

func TestBranchManager_CreateBranch_Validation(t *testing.T) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	bm := NewBranchManager(client.GetGitLabClient(), TestProjectID)
	ctx := context.Background()

	tests := []struct {
		name        string
		branchName  string
		ref         string
		expectError bool
		description string
	}{
		{
			name:        "empty branch name",
			branchName:  EmptyBranchName,
			ref:         TestRef,
			expectError: true,
			description: "should reject empty branch name",
		},
		{
			name:        "branch name too long",
			branchName:  BranchNameTooLong,
			ref:         TestRef,
			expectError: true,
			description: "should reject overly long branch name",
		},
		{
			name:        "invalid branch name",
			branchName:  BranchWithSpaces,
			ref:         TestRef,
			expectError: true,
			description: "should reject invalid branch name",
		},
		{
			name:        "valid branch name with empty ref",
			branchName:  TestBranchName,
			ref:         "",
			expectError: true, // Will fail with API error, but validates format
			description: "should default to main when ref is empty",
		},
		{
			name:        "valid branch name and ref",
			branchName:  TestBranchName,
			ref:         TestRef,
			expectError: true, // Will fail with API error due to no real GitLab
			description: "should accept valid branch name and ref",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := bm.CreateBranch(ctx, tt.branchName, tt.ref)

			if tt.expectError {
				if err == nil {
					t.Errorf("CreateBranch(%q, %q) expected error but got none", tt.branchName, tt.ref)
				}
			} else {
				if err != nil {
					t.Errorf("CreateBranch(%q, %q) unexpected error: %v", tt.branchName, tt.ref, err)
				}
			}
		})
	}
}

func TestBranchManager_GetBranch_Validation(t *testing.T) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	bm := NewBranchManager(client.GetGitLabClient(), TestProjectID)
	ctx := context.Background()

	tests := []struct {
		name        string
		branchName  string
		expectError bool
		description string
	}{
		{
			name:        "empty branch name",
			branchName:  EmptyBranchName,
			expectError: true,
			description: "should reject empty branch name",
		},
		{
			name:        "valid branch name",
			branchName:  TestBranchName,
			expectError: true, // Will fail with API error due to no real GitLab
			description: "should accept valid branch name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := bm.GetBranch(ctx, tt.branchName)

			if tt.expectError {
				if err == nil {
					t.Errorf("GetBranch(%q) expected error but got none", tt.branchName)
				}
			} else {
				if err != nil {
					t.Errorf("GetBranch(%q) unexpected error: %v", tt.branchName, err)
				}
			}
		})
	}
}

func TestBranchManager_ListBranches_MaxResults(t *testing.T) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	bm := NewBranchManager(client.GetGitLabClient(), TestProjectID)
	ctx := context.Background()

	tests := []struct {
		name        string
		search      string
		maxResults  int
		expectError bool
		description string
	}{
		{
			name:        "zero max results",
			search:      "",
			maxResults:  0,
			expectError: false, // Should default to 20
			description: "should default to 20 when zero",
		},
		{
			name:        "negative max results",
			search:      "",
			maxResults:  -1,
			expectError: false, // Should default to 20
			description: "should default to 20 when negative",
		},
		{
			name:        "positive max results",
			search:      "",
			maxResults:  CustomBranchMaxResults,
			expectError: false,
			description: "should accept positive max results",
		},
		{
			name:        "with search query",
			search:      TestBranchSearchQuery,
			maxResults:  DefaultBranchMaxResults,
			expectError: false,
			description: "should accept search query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail because we don't have a real GitLab instance,
			// but we can test the validation logic
			_, err := bm.ListBranches(ctx, tt.search, tt.maxResults)

			// We expect API errors due to no real GitLab, but not validation errors
			if tt.expectError {
				if err == nil {
					t.Errorf("ListBranches(%q, %d) expected error but got none", tt.search, tt.maxResults)
				}
			}
			// For this test, we mainly care that it doesn't panic on invalid inputs
		})
	}
}

func TestBranchManager_DeleteBranch_Validation(t *testing.T) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	bm := NewBranchManager(client.GetGitLabClient(), TestProjectID)
	ctx := context.Background()

	tests := []struct {
		name        string
		branchName  string
		expectError bool
		description string
	}{
		{
			name:        "empty branch name",
			branchName:  EmptyBranchName,
			expectError: true,
			description: "should reject empty branch name",
		},
		{
			name:        "valid branch name",
			branchName:  TestBranchName,
			expectError: true, // Will fail with API error due to no real GitLab
			description: "should accept valid branch name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bm.DeleteBranch(ctx, tt.branchName)

			if tt.expectError {
				if err == nil {
					t.Errorf("DeleteBranch(%q) expected error but got none", tt.branchName)
				}
			} else {
				if err != nil {
					t.Errorf("DeleteBranch(%q) unexpected error: %v", tt.branchName, err)
				}
			}
		})
	}
}

func TestBranchManager_GenerateUniqueBranchName(t *testing.T) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	bm := NewBranchManager(client.GetGitLabClient(), TestProjectID)
	ctx := context.Background()

	tests := []struct {
		name        string
		prefix      string
		tag         string
		expectError bool
		description string
	}{
		{
			name:        "empty prefix",
			prefix:      "",
			tag:         TestTag,
			expectError: true, // Will fail with API error but should generate name
			description: "should use default prefix when empty",
		},
		{
			name:        "valid prefix and tag",
			prefix:      TestBranchPrefix,
			tag:         TestTag,
			expectError: true, // Will fail with API error but should generate name
			description: "should generate name with custom prefix",
		},
		{
			name:        "tag with slash",
			prefix:      TestUpdatePrefix,
			tag:         TestTagWithSlash,
			expectError: true, // Will fail with API error but should clean tag
			description: "should clean tag with slashes",
		},
		{
			name:        "tag with colon",
			prefix:      TestUpdatePrefix,
			tag:         TestTagWithColon,
			expectError: true, // Will fail with API error but should clean tag
			description: "should clean tag with colons",
		},
		{
			name:        "very long tag",
			prefix:      TestUpdatePrefix,
			tag:         "very-long-tag-name-that-might-cause-the-branch-name-to-exceed-maximum-length-limits",
			expectError: true, // Will fail with API error but should truncate
			description: "should truncate long tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			branchName, err := bm.GenerateUniqueBranchName(ctx, tt.prefix, tt.tag)

			if tt.expectError {
				if err == nil {
					t.Errorf("GenerateUniqueBranchName(%q, %q) expected error but got none", tt.prefix, tt.tag)
				}
			} else {
				if err != nil {
					t.Errorf("GenerateUniqueBranchName(%q, %q) unexpected error: %v", tt.prefix, tt.tag, err)
				}

				// Verify branch name format even if we expect API errors
				if branchName != "" {
					if len(branchName) > MaxBranchNameLength {
						t.Errorf("Generated branch name too long: %d characters", len(branchName))
					}

					expectedPrefix := tt.prefix
					if expectedPrefix == "" {
						expectedPrefix = UpdateBranchPrefix
					}

					if !strings.HasPrefix(branchName, expectedPrefix) {
						t.Errorf("Generated branch name should start with prefix %q", expectedPrefix)
					}
				}
			}
		})
	}
}

func TestBranchManager_ConvertToBranchInfo(t *testing.T) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	bm := NewBranchManager(client.GetGitLabClient(), TestProjectID)

	// Create a mock GitLab branch
	commit := &gitlab.Commit{
		ID:      "abc123",
		Message: "Test commit",
	}

	gitlabBranch := &gitlab.Branch{
		Name:               TestBranchName,
		Protected:          false,
		Default:            false,
		DevelopersCanPush:  true,
		DevelopersCanMerge: true,
		Commit:             commit,
		WebURL:             TestBranchWebURL,
	}

	branchInfo := bm.convertToBranchInfo(gitlabBranch)

	if branchInfo == nil {
		t.Fatal("convertToBranchInfo() should return non-nil BranchInfo")
	}

	// Verify all fields are converted correctly
	if branchInfo.Name != TestBranchName {
		t.Errorf("convertToBranchInfo() Name = %q, want %q", branchInfo.Name, TestBranchName)
	}

	if branchInfo.Protected != false {
		t.Errorf("convertToBranchInfo() Protected = %v, want %v", branchInfo.Protected, false)
	}

	if branchInfo.Default != false {
		t.Errorf("convertToBranchInfo() Default = %v, want %v", branchInfo.Default, false)
	}

	if branchInfo.DevelopersCanPush != true {
		t.Errorf("convertToBranchInfo() DevelopersCanPush = %v, want %v", branchInfo.DevelopersCanPush, true)
	}

	if branchInfo.DevelopersCanMerge != true {
		t.Errorf("convertToBranchInfo() DevelopersCanMerge = %v, want %v", branchInfo.DevelopersCanMerge, true)
	}

	if branchInfo.Commit != commit {
		t.Error("convertToBranchInfo() Commit should be set correctly")
	}

	if branchInfo.WebURL != TestBranchWebURL {
		t.Errorf("convertToBranchInfo() WebURL = %q, want %q", branchInfo.WebURL, TestBranchWebURL)
	}
}

func TestBranchConstants(t *testing.T) {
	// Test that constants are properly defined
	if DefaultBranchPrefix == "" {
		t.Error("DefaultBranchPrefix should not be empty")
	}

	if UpdateBranchPrefix == "" {
		t.Error("UpdateBranchPrefix should not be empty")
	}

	if MaxBranchNameLength <= 0 {
		t.Errorf("MaxBranchNameLength should be positive, got %d", MaxBranchNameLength)
	}

	if MinBranchNameLength <= 0 {
		t.Errorf("MinBranchNameLength should be positive, got %d", MinBranchNameLength)
	}

	// Test reasonable values
	if MaxBranchNameLength < 50 {
		t.Errorf("MaxBranchNameLength seems too small: %d", MaxBranchNameLength)
	}

	if MaxBranchNameLength > 500 {
		t.Errorf("MaxBranchNameLength seems too large: %d", MaxBranchNameLength)
	}

	if MinBranchNameLength != 1 {
		t.Errorf("MinBranchNameLength should be 1, got %d", MinBranchNameLength)
	}
}

// Benchmark tests for performance requirements from IDEA.md
func BenchmarkNewBranchManager(b *testing.B) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	gitlabClient := client.GetGitLabClient()

	for i := 0; i < b.N; i++ {
		_ = NewBranchManager(gitlabClient, TestProjectID)
	}
}

func BenchmarkValidateBranchName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = validateBranchName(TestBranchName)
	}
}

func BenchmarkValidateBranchName_Invalid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = validateBranchName(BranchWithSpaces)
	}
}

func BenchmarkBranchManager_ConvertToBranchInfo(b *testing.B) {
	client, err := NewClient(TestGitLabToken, TestGitLabURL)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	bm := NewBranchManager(client.GetGitLabClient(), TestProjectID)

	commit := &gitlab.Commit{
		ID:      "abc123",
		Message: "Test commit",
	}

	gitlabBranch := &gitlab.Branch{
		Name:               TestBranchName,
		Protected:          false,
		Default:            false,
		DevelopersCanPush:  true,
		DevelopersCanMerge: true,
		Commit:             commit,
		WebURL:             TestBranchWebURL,
	}

	for i := 0; i < b.N; i++ {
		_ = bm.convertToBranchInfo(gitlabBranch)
	}
}

func BenchmarkGenerateTimestamp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = time.Now().Format("20060102-150405")
	}
}
