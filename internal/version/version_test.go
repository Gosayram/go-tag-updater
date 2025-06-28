package version

import (
	"runtime"
	"strings"
	"testing"
)

const (
	// Test constants as per IDEA.md requirements
	TestVersion     = "1.0.0"
	TestCommit      = "abc123"
	TestDate        = "2025-01-01_12:00:00"
	TestBuiltBy     = "test-user"
	TestBuildNumber = "42"
)

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		buildNumber string
		expected    string
	}{
		{
			name:        "returns version with build number",
			version:     TestVersion,
			buildNumber: TestBuildNumber,
			expected:    TestVersion + " (build " + TestBuildNumber + ")",
		},
		{
			name:        "returns version without build number when zero",
			version:     TestVersion,
			buildNumber: "0",
			expected:    TestVersion,
		},
		{
			name:        "returns version without build number when empty",
			version:     TestVersion,
			buildNumber: "",
			expected:    TestVersion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save originals
			originalVersion := Version
			originalBuildNumber := BuildNumber
			defer func() {
				Version = originalVersion
				BuildNumber = originalBuildNumber
			}()

			// Set test values
			Version = tt.version
			BuildNumber = tt.buildNumber

			result := GetVersion()
			if result != tt.expected {
				t.Errorf("GetVersion() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetFullVersionInfo(t *testing.T) {
	// Save originals
	originalVersion := Version
	originalCommit := Commit
	originalDate := Date
	originalBuiltBy := BuiltBy
	defer func() {
		Version = originalVersion
		Commit = originalCommit
		Date = originalDate
		BuiltBy = originalBuiltBy
	}()

	// Set test values
	Version = TestVersion
	Commit = TestCommit
	Date = TestDate
	BuiltBy = TestBuiltBy

	result := GetFullVersionInfo()

	// Should contain all components (note: date gets formatted with spaces instead of underscores)
	expectedComponents := []string{
		TestVersion,
		TestCommit,
		strings.ReplaceAll(TestDate, "_", " "), // Date gets formatted
		TestBuiltBy,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	}

	for _, component := range expectedComponents {
		if !strings.Contains(result, component) {
			t.Errorf("GetFullVersionInfo() = %q, expected to contain %q", result, component)
		}
	}

	// Should be multiline
	lines := strings.Split(result, "\n")
	minExpectedLines := 2
	if len(lines) < minExpectedLines {
		t.Errorf("GetFullVersionInfo() returned %d lines, expected at least %d", len(lines), minExpectedLines)
	}

	// Should start with app name
	if !strings.HasPrefix(result, "go-tag-updater") {
		t.Errorf("GetFullVersionInfo() = %q, expected to start with 'go-tag-updater'", result)
	}
}

func TestGet(t *testing.T) {
	// Save originals
	originalVersion := Version
	originalCommit := Commit
	originalDate := Date
	originalBuiltBy := BuiltBy
	defer func() {
		Version = originalVersion
		Commit = originalCommit
		Date = originalDate
		BuiltBy = originalBuiltBy
	}()

	// Set test values
	Version = TestVersion
	Commit = TestCommit
	Date = TestDate
	BuiltBy = TestBuiltBy

	result := Get()

	// Check all fields
	if result.Version != TestVersion {
		t.Errorf("Get().Version = %q, want %q", result.Version, TestVersion)
	}

	if result.Commit != TestCommit {
		t.Errorf("Get().Commit = %q, want %q", result.Commit, TestCommit)
	}

	if result.Date != TestDate {
		t.Errorf("Get().Date = %q, want %q", result.Date, TestDate)
	}

	if result.BuiltBy != TestBuiltBy {
		t.Errorf("Get().BuiltBy = %q, want %q", result.BuiltBy, TestBuiltBy)
	}

	if result.GoVersion != runtime.Version() {
		t.Errorf("Get().GoVersion = %q, want %q", result.GoVersion, runtime.Version())
	}

	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if result.Platform != expectedPlatform {
		t.Errorf("Get().Platform = %q, want %q", result.Platform, expectedPlatform)
	}
}

func TestBuildInfo_String(t *testing.T) {
	bi := &BuildInfo{
		Version:   TestVersion,
		Commit:    TestCommit,
		Date:      TestDate,
		BuiltBy:   TestBuiltBy,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}

	result := bi.String()

	// Should contain all components
	expectedComponents := []string{
		TestVersion,
		TestCommit,
		TestDate,
		TestBuiltBy,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	}

	for _, component := range expectedComponents {
		if !strings.Contains(result, component) {
			t.Errorf("BuildInfo.String() = %q, expected to contain %q", result, component)
		}
	}
}

func TestBuildInfo_Short(t *testing.T) {
	bi := &BuildInfo{
		Version: TestVersion,
	}

	result := bi.Short()
	expected := "go-tag-updater " + TestVersion

	if result != expected {
		t.Errorf("BuildInfo.Short() = %q, want %q", result, expected)
	}
}

func TestConstants(t *testing.T) {
	// Test that constants are properly defined
	if ShortCommitHashLength <= 0 {
		t.Errorf("ShortCommitHashLength should be positive, got %d", ShortCommitHashLength)
	}

	if UnknownValue == "" {
		t.Error("UnknownValue should not be empty")
	}
}

func TestCommitTruncation(t *testing.T) {
	// Save originals
	originalVersion := Version
	originalCommit := Commit
	defer func() {
		Version = originalVersion
		Commit = originalCommit
	}()

	// Set test values with long commit
	Version = TestVersion
	longCommit := "abcdef1234567890abcdef1234567890"
	Commit = longCommit

	result := GetFullVersionInfo()

	// Should contain truncated commit
	expectedShortCommit := longCommit[:ShortCommitHashLength]
	if !strings.Contains(result, expectedShortCommit) {
		t.Errorf("GetFullVersionInfo() = %q, expected to contain truncated commit %q", result, expectedShortCommit)
	}

	// Should not contain full long commit
	if strings.Contains(result, longCommit) {
		t.Errorf("GetFullVersionInfo() = %q, should not contain full commit %q", result, longCommit)
	}
}

func TestUnknownValues(t *testing.T) {
	// Save originals
	originalCommit := Commit
	originalDate := Date
	originalBuiltBy := BuiltBy
	defer func() {
		Commit = originalCommit
		Date = originalDate
		BuiltBy = originalBuiltBy
	}()

	// Set unknown values
	Commit = UnknownValue
	Date = UnknownValue
	BuiltBy = UnknownValue

	result := GetFullVersionInfo()

	// Should still work with unknown values
	if !strings.Contains(result, "go-tag-updater") {
		t.Errorf("GetFullVersionInfo() = %q, expected to contain app name", result)
	}

	// Should contain Go version and platform even with unknown values
	if !strings.Contains(result, runtime.Version()) {
		t.Errorf("GetFullVersionInfo() = %q, expected to contain Go version", result)
	}
}

// Benchmark tests for performance requirements from IDEA.md
func BenchmarkGetVersion(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetVersion()
	}
}

func BenchmarkGetFullVersionInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetFullVersionInfo()
	}
}

func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Get()
	}
}
