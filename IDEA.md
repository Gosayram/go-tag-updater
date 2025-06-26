# go-tag-updater

## Project Overview

A Go CLI tool for safely updating YAML files in GitLab repositories through automated merge request workflows. The tool provides intelligent conflict detection, flexible project identification, comprehensive merge request lifecycle management, and built-in version tracking.

## Core Objectives

- Update YAML files in GitLab repositories with new tag values
- Create and manage merge requests automatically
- Implement safe merging with conflict detection and prevention
- Support both numeric project IDs and human-readable project paths
- Provide comprehensive debugging and logging capabilities
- Maintain secure path handling for file operations
- Display version information with build-time metadata

## Key Features

### 1. Version Information

Built-in version tracking with comprehensive build metadata:

- **Version flag**: `-v` or `--version` displays complete version information
- **Build-time variables**: Version, Git commit hash, build date, and builder info
- **Runtime details**: Go version and target platform information

Example output:
```
go-tag-updater 1.0.0 (7488953) built 2025-06-25 22:40:38 by Gosayram with go1.24.4 for darwin/arm64
```

### 2. Flexible Project Identification

The tool accepts project identification in two formats:

- **Numeric ID**: `--project-id=4323829`
- **Human-readable path**: `--project-id=openproject/infra/dev`

When a path containing forward slashes is provided, the tool automatically resolves it to a numeric ID using the GitLab API:
```
GET /api/v4/projects/$(urlencode(project-path))
```

### 3. Debug Mode

Comprehensive debugging capabilities through the `--debug` flag:

- HTTP request/response logging for all GitLab API calls
- Git command execution details
- File modification tracking
- JSON response parsing and validation
- Error context and stack traces

### 4. Merge Request Conflict Prevention

Before creating new merge requests, the tool performs safety checks:

- Scans for existing merge requests from the same branch
- Checks for pending merge requests targeting the same file
- Implements configurable waiting behavior for conflicting merge requests
- Provides clear conflict resolution guidance

### 5. Safe Automation with Security

- Atomic operations with rollback capabilities
- Validation of YAML syntax before committing changes
- Pre-merge conflict detection
- Comprehensive error handling with actionable messages
- **Path traversal protection**: Validates file paths to prevent `../` attacks
- **Secure file operations**: All file reads/writes are validated and sanitized

## CLI Interface

### Command Structure
```bash
go-tag-updater [flags]
```

### Version Information
```bash
go-tag-updater --version
go-tag-updater -v
```

### Required Flags

| Flag | Type | Description |
|------|------|-------------|
| `--project-id` | string | GitLab project ID or path (group/subgroup/project) |
| `--file` | string | Path to target YAML file within repository |
| `--new-tag` | string | New tag value to set in YAML file |
| `--token` | string | GitLab Personal Access Token |

### Optional Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--branch-name` | string | auto-generated | Name for the new feature branch |
| `--target-branch` | string | `main` | Target branch for merge request |
| `--wait-previous-mr` | bool | `false` | Wait for conflicting merge requests to complete |
| `--debug` | bool | `false` | Enable verbose debugging output |
| `--dry-run` | bool | `false` | Preview changes without execution |
| `--auto-merge` | bool | `false` | Automatically merge when pipeline passes |
| `-v, --version` | bool | `false` | Display version information and exit |

### Usage Examples

#### Version Information
```bash
go-tag-updater --version
```

#### Basic Usage
```bash
go-tag-updater \
  --project-id=openproject/infra/dev \
  --file=apps/dev/project1/app.yml \
  --new-tag=v1.2.3 \
  --token=$GITLAB_TOKEN
```

#### Advanced Usage with Debug
```bash
go-tag-updater \
  --project-id=4323829 \
  --file=config/deployment.yaml \
  --new-tag=abc123 \
  --branch-name=update-tag-abc123 \
  --target-branch=development \
  --wait-previous-mr=true \
  --debug \
  --token=$GITLAB_TOKEN
```

## Architecture Design

### Package Structure

```
go-tag-updater/
├── cmd/
│   └── go-tag-updater/
│       ├── main.go          # Application entry point
│       └── root.go          # CLI command definitions and flag parsing
├── internal/
│   ├── version/
│   │   └── version.go       # Version information and build metadata
│   ├── config/
│   │   └── config.go        # Configuration management and validation
│   ├── gitlab/
│   │   ├── client.go        # GitLab API client implementation
│   │   ├── projects.go      # Project resolution and management
│   │   ├── merge_requests_simple.go # MR lifecycle management
│   │   ├── conflicts.go     # Conflict detection logic
│   │   ├── branches.go      # Branch management operations
│   │   └── files.go         # File operations in GitLab
│   ├── yaml/
│   │   ├── parser.go        # YAML file parsing and validation
│   │   └── updater.go       # Tag update operations with security
│   ├── workflow/
│   │   └── simple_workflow.go # Main workflow orchestration
│   ├── logger/
│   │   └── debug.go         # Debugging and logging utilities
│   └── git/                 # Git operations (reserved for future use)
├── pkg/
│   └── errors/
│       └── types.go         # Custom error types
├── docs/
│   └── ARCHITECTURE.md      # Detailed architecture documentation
├── scripts/
│   └── check-commit-msg.sh  # Git commit message validation
├── bin/                     # Build output directory
├── Makefile                 # Build automation and quality checks
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums
├── .golangci.yml           # Linter configuration
├── .gosec.json             # Security scanner configuration
├── .errcheck_excludes.txt  # Error check exclusions
└── IDEA.md                 # This file
```

### Core Components

#### 1. Version Management (`internal/version/`)

**Constants:**
```go
const (
    DefaultVersion = "dev"
    DefaultCommit = "unknown"
    DefaultDate = "unknown"
    DefaultBuiltBy = "unknown"
)
```

**Responsibilities:**
- Version information display
- Build metadata management
- Runtime environment details
- CLI version flag handling

#### 2. GitLab Client (`internal/gitlab/`)

**Constants:**
```go
const (
    DefaultAPIVersion = "v4"
    ProjectsEndpoint = "/api/v4/projects"
    MergeRequestsEndpoint = "/api/v4/projects/%d/merge_requests"
    DefaultTimeout = 30 * time.Second
    MaxRetryAttempts = 3
)
```

**Responsibilities:**
- Project ID resolution from paths
- Merge request creation and management
- API authentication and rate limiting
- Conflict detection and prevention
- Branch and file operations

#### 3. YAML Processor (`internal/yaml/`)

**Constants:**
```go
const (
    DefaultIndentation = 2
    MaxFileSize = 10 * 1024 * 1024 // 10MB
    BackupExtension = ".backup"
    MaxPathLength = 4096
)
```

**Responsibilities:**
- YAML file parsing and validation
- Tag value updates with preservation of formatting
- Backup creation and restoration
- Syntax validation
- **Secure path validation** to prevent directory traversal attacks

#### 4. Workflow Orchestration (`internal/workflow/`)

**Constants:**
```go
const (
    DefaultWorkflowTimeout = 300 * time.Second
    MaxConcurrentOperations = 5
    RetryBackoffDuration = 2 * time.Second
)
```

**Responsibilities:**
- Main workflow execution and coordination
- Step-by-step process management
- Error handling and recovery
- Progress tracking and reporting

#### 5. Configuration Management (`internal/config/`)

**Constants:**
```go
const (
    DefaultConfigFile = ".go-tag-updater.yaml"
    EnvPrefix = "GO_TAG_UPDATER"
    DefaultMergeTimeout = 300 * time.Second
    MinTokenLength = 10
)
```

**Responsibilities:**
- CLI flag parsing and validation
- Environment variable support
- Configuration file loading
- Default value management
- Input sanitization and security checks

## Security Features

### Path Traversal Protection

The tool implements comprehensive security measures for file operations:

- **Path validation**: All file paths are validated to prevent `../` directory traversal attacks
- **Absolute path restrictions**: Limits absolute paths to safe system directories
- **Path normalization**: Cleans and normalizes file paths before operations
- **Safe file operations**: All read/write operations are protected with validation

### Security Implementation

```go
const (
    MaxPathLength = 4096
    ForbiddenPathPattern = `\.\.`
    SafeDirectoryPrefixes = []string{"/tmp", "/var/tmp"}
)

// File operations include security validation
func validateAndCleanFilePath(filePath string) (string, error) {
    // Security validation logic
    // Path normalization
    // Directory traversal prevention
}
```

## Implementation Status

### Completed Features ✅
- ✅ CLI framework with Cobra
- ✅ GitLab API client implementation  
- ✅ YAML processing with security
- ✅ Project resolution (ID and path)
- ✅ Merge request lifecycle management
- ✅ Conflict detection and prevention
- ✅ Debug mode and comprehensive logging
- ✅ Version information with build metadata
- ✅ Security hardening with path validation
- ✅ Complete build system with quality checks

### Architecture Phases

#### Phase 1: Core Infrastructure ✅
- CLI framework setup with cobra
- GitLab API client implementation
- Basic YAML processing capabilities
- Version management system

#### Phase 2: Project Resolution ✅  
- Numeric ID handling
- Human-readable path resolution
- API endpoint construction
- Error handling for invalid projects

#### Phase 3: Merge Request Management ✅
- MR creation and lifecycle management
- Conflict detection implementation
- Wait logic for pending merge requests
- Status monitoring and reporting

#### Phase 4: Safety and Validation ✅
- YAML syntax validation
- File backup and restoration
- Atomic operations implementation
- Comprehensive error handling
- **Security hardening with path validation**

#### Phase 5: Advanced Features ✅
- Debug mode implementation
- Dry-run capabilities
- Auto-merge functionality
- Version information display
- Performance optimization

## Error Handling Strategy

### Custom Error Types
```go
const (
    ErrCodeInvalidProject = 1001
    ErrCodeFileNotFound = 1002
    ErrCodeInvalidYAML = 1003
    ErrCodeMergeConflict = 1004
    ErrCodeAPITimeout = 1005
    ErrCodeSecurityViolation = 1006
    ErrCodePathTraversal = 1007
)
```

### Error Categories
- **Configuration Errors**: Invalid flags, missing tokens, malformed inputs
- **API Errors**: GitLab connectivity, authentication, rate limiting
- **File System Errors**: Missing files, permission issues, disk space
- **Security Errors**: Path traversal attempts, unauthorized file access
- **Validation Errors**: YAML syntax, tag format, project existence
- **Workflow Errors**: Process failures, timeout conditions

## Testing Strategy

### Unit Tests
- Individual component testing with mocked dependencies
- Error condition simulation
- Edge case validation
- Security vulnerability testing
- Performance benchmarking

### Integration Tests
- GitLab API interaction testing
- End-to-end workflow validation
- Security validation testing
- YAML processing accuracy

### Security Tests
- Path traversal attack prevention
- File access boundary validation
- Input sanitization verification

### Test Constants
```go
const (
    TestProjectID = 123456
    TestProjectPath = "test/group/project"
    TestTimeout = 5 * time.Second
    MaxTestRetries = 3
    TestSecureDirectory = "/tmp/go-tag-updater-test"
)
```

## Security Considerations

- Secure token handling with environment variable support
- Input validation and sanitization
- **Path traversal protection for all file operations**
- Temporary file cleanup
- Rate limiting compliance
- Audit logging for sensitive operations
- Build-time security scanning with gosec
- Dependency vulnerability checking

## Performance Requirements

- Repository operations under 30 seconds for typical repositories
- API response handling within 5 seconds per request
- Memory usage under 100MB for standard operations
- File operations with security validation under 100ms
- Concurrent operation support for batch processing

## Quality Assurance

The project maintains high quality standards through:

- **Linting**: golangci-lint with comprehensive rule set
- **Security Scanning**: gosec for vulnerability detection
- **Static Analysis**: staticcheck for code quality
- **Error Checking**: errcheck for unhandled errors
- **Vulnerability Scanning**: govulncheck for dependency issues
- **SBOM Generation**: Software Bill of Materials tracking

## Monitoring and Observability

- Structured logging with configurable levels
- Operation timing and performance metrics
- API call tracking and rate limit monitoring
- Error categorization and reporting
- Debug mode with comprehensive request/response logging
- Security event logging and monitoring

## Future Enhancements

- Batch processing for multiple files
- Configuration file templates
- Integration with CI/CD pipelines
- Rollback capabilities for failed operations
- Advanced conflict resolution strategies
- Support for additional file formats beyond YAML
- Enhanced security features and audit trails
- Performance optimizations for large repositories 