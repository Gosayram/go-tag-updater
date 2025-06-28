# go-tag-updater

## Project Overview

A Go CLI tool for safely updating YAML files in GitLab repositories through automated merge request workflows. The tool provides intelligent conflict detection, flexible project identification, comprehensive merge request lifecycle management, and built-in version tracking using the official GitLab Go API client library.

## Core Objectives

- Update YAML files in GitLab repositories with new tag values
- Create and manage merge requests automatically through GitLab API
- Implement safe merging with conflict detection and prevention
- Support both numeric project IDs and human-readable project paths
- Provide comprehensive debugging and logging capabilities
- Maintain secure path handling for file operations
- Display version information with build-time metadata
- Leverage official GitLab Go client for reliable API integration

## Key Features

### 1. Version Information

Built-in version tracking with comprehensive build metadata:

- **Version flag**: `-v` or `--version` displays complete version information
- **Build-time variables**: Version, Git commit hash, build date, and builder info
- **Runtime details**: Go version and target platform information

Example output:
```
go-tag-updater v1.0.0 (build 1) (7488953)
built 2025-01-25 12:30:45 by Gosayram with go1.24.4 for darwin/arm64
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

### 6. Dry Run Mode

Preview functionality without making actual changes:

- Validates file existence and accessibility
- Shows YAML content changes
- Generates branch names and commit messages
- Previews merge request creation
- No actual GitLab operations performed

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

#### Dry Run Preview
```bash
go-tag-updater \
  --project-id=mygroup/myproject \
  --file=k8s/deployment.yaml \
  --new-tag=v2.0.0 \
  --dry-run \
  --token=$GITLAB_TOKEN
```

## Architecture Design

### Package Structure

```
go-tag-updater/
├── cmd/
│   └── go-tag-updater/
│       ├── main.go          # Application entry point with exit codes
│       └── root.go          # CLI command definitions and flag parsing
├── internal/
│   ├── version/
│   │   ├── version.go       # Version information and build metadata
│   │   └── version_test.go  # ✅ Version management tests
│   ├── config/
│   │   └── config.go        # Configuration management and validation
│   ├── gitlab/
│   │   ├── client.go        # GitLab API client implementation
│   │   ├── client_test.go   # ✅ Client tests with benchmarks
│   │   ├── projects.go      # Project resolution and management
│   │   ├── projects_test.go # ✅ Project validation tests
│   │   ├── merge_requests_simple.go # MR lifecycle management
│   │   ├── conflicts.go     # Conflict detection logic
│   │   ├── branches.go      # Branch management operations
│   │   ├── branches_test.go # ✅ Branch management tests
│   │   └── files.go         # File operations in GitLab
│   ├── yaml/
│   │   ├── parser.go        # YAML file parsing and validation
│   │   ├── updater.go       # Tag update operations with security
│   │   └── updater_test.go  # ✅ YAML processing tests
│   ├── workflow/
│   │   ├── simple_workflow.go     # Main workflow orchestration
│   │   └── simple_workflow_test.go # ✅ Workflow orchestration tests
│   ├── logger/
│   │   ├── debug.go         # Debugging and logging utilities
│   │   └── debug_test.go    # ✅ Logger tests
│   └── git/                 # Git operations (reserved for future use)
├── pkg/
│   └── errors/
│       └── types.go         # Custom error types with specific codes
├── docs/
│   └── ARCHITECTURE.md      # Detailed architecture documentation
├── scripts/
│   └── check-commit-msg.sh  # Git commit message validation
├── bin/                     # Build output directory
├── Makefile                 # Build automation and quality checks
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums
├── .golangci.yml           # Linter configuration with staticcheck
├── .gosec.json             # Security scanner configuration
├── .errcheck_excludes.txt  # Error check exclusions
└── IDEA.md                 # This file
```

### Core Components

#### 1. Version Management (`internal/version/`)

**Constants:**
```go
const (
    ShortCommitHashLength = 7
    UnknownValue = "unknown"
)
```

**Build Variables:**
```go
var (
    Version     = "dev"
    Commit      = "unknown"
    Date        = "unknown"
    BuiltBy     = "unknown"
    BuildNumber = "0"
)
```

**Responsibilities:**
- Version information display with build metadata
- Build-time variable management via linker flags
- Runtime environment details (Go version, platform)
- CLI version flag handling with formatted output

#### 2. GitLab Client (`internal/gitlab/`)

**Constants:**
```go
const (
    DefaultBranchPrefix = "feature/"
    UpdateBranchPrefix  = "update-tag/"
    MaxBranchNameLength = 100
    MinBranchNameLength = 1
    DefaultBranch = "main"
    DefaultTimeout = 30 * time.Second
    MaxRetryAttempts = 3
)
```

**Components:**
- **Client**: Wrapper around official GitLab Go client
- **FileManager**: Repository file CRUD operations
- **BranchManager**: Git branch lifecycle management
- **SimpleMergeRequestManager**: Basic MR operations
- **Project resolution**: Supports both numeric IDs and paths

**Responsibilities:**
- Project ID resolution from paths using GitLab API
- Merge request creation and management
- API authentication and health checks
- File operations with Base64 encoding/decoding
- Branch creation, validation, and cleanup
- Conflict detection and prevention

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
- Syntax validation before and after updates
- **Secure path validation** to prevent directory traversal attacks

#### 4. Workflow Orchestration (`internal/workflow/`)

**Constants:**
```go
const (
    PreviewContentMaxLength = 500
    TempFilePermissions = 0o600
)
```

**SimpleTagUpdater Components:**
- File existence validation
- YAML content processing with temporary files
- Branch name generation (auto or custom)
- Dry-run mode with content preview
- Complete workflow execution with cleanup

**Responsibilities:**
- Main workflow execution and coordination
- Step-by-step process management
- Error handling and recovery with rollback
- Progress tracking and reporting
- Dry-run functionality for testing

#### 5. Configuration Management (`internal/config/`)

**Constants:**
```go
const (
    DefaultConfigFile = "go-tag-updater.yaml"
    AlternateConfigFile = "go-tag-updater.yml"
    EnvPrefix = "GO_TAG_UPDATER"
    DefaultMergeTimeout = 300 * time.Second
    DefaultTimeout = 30 * time.Second
    DefaultRateLimitRPS = 10
    DefaultMaxConcurrentReqs = 5
    DefaultRetryCount = 3
)
```

**Configuration Types:**
- **Config**: File-based configuration with YAML support
- **CLIConfig**: Command-line specific configuration
- **GitLabConfig**: GitLab API specific settings
- **PerformanceConfig**: Rate limiting and timeouts

**Responsibilities:**
- CLI flag parsing and validation with Cobra/Viper
- Environment variable support with prefix
- Configuration file loading (YAML)
- Default value management
- Input sanitization and security checks

#### 6. Logging (`internal/logger/`)

**Responsibilities:**
- Structured logging with configurable debug mode
- Operation tracking with context fields
- Error logging with stack traces
- GitLab API request/response logging in debug mode

## GitLab API Integration

### Official Client Library

The application uses the official GitLab Go client library:
- **Library**: `gitlab.com/gitlab-org/api/client-go v0.130.1`: Official GitLab API client
- **Version**: v0.130.1
- **Features**: Full GitLab REST API v4 support

### API Services Used

1. **Projects Service**: Project information and path resolution
2. **RepositoryFiles Service**: File CRUD operations with Base64 handling
3. **Branches Service**: Branch management and validation
4. **MergeRequests Service**: MR lifecycle operations
5. **Users Service**: Authentication health checks

### Security Features

- Token-based authentication
- Configurable GitLab instance URL
- Health check validation
- Rate limiting awareness
- Secure error handling without information disclosure

## Implementation Status

### Completed Features ✅
- ✅ CLI framework with Cobra and Viper
- ✅ GitLab API client implementation with official library
- ✅ YAML processing with security validation
- ✅ Project resolution (ID and path) via GitLab API
- ✅ Merge request lifecycle management
- ✅ Branch creation and management with validation
- ✅ File operations with Base64 encoding/decoding
- ✅ Conflict detection and prevention
- ✅ Debug mode and comprehensive logging
- ✅ Version information with build metadata
- ✅ Security hardening with path validation
- ✅ Dry-run mode for safe testing
- ✅ Complete build system with quality checks
- ✅ **Comprehensive test coverage** for critical components
- ✅ **Unit tests** for version, workflow, GitLab packages
- ✅ **Performance benchmarking** for all critical operations
- ✅ **Security scanning** with gosec and govulncheck
- ✅ **SBOM generation** for supply chain security

### Architecture Phases

#### Phase 1: Core Infrastructure ✅
- CLI framework setup with Cobra/Viper
- GitLab API client implementation with official library
- Basic YAML processing capabilities
- Version management system with build metadata

#### Phase 2: Project Resolution ✅  
- Numeric ID handling via GitLab API
- Human-readable path resolution
- API endpoint construction and validation
- Error handling for invalid projects

#### Phase 3: Merge Request Management ✅
- MR creation and lifecycle management
- Branch operations with validation
- File operations with Base64 handling
- Status monitoring and reporting

#### Phase 4: Safety and Validation ✅
- YAML syntax validation with proper parsing
- Temporary file handling with secure permissions
- Atomic operations implementation
- Comprehensive error handling with cleanup
- **Security hardening with path validation**

#### Phase 5: Advanced Features ✅
- Debug mode implementation with API logging
- Dry-run capabilities with content preview
- Auto-merge functionality (configuration)
- Version information display with build details
- Performance optimization and benchmarking

#### Phase 6: Comprehensive Testing ✅
- Unit test implementation for all core packages
- Security vulnerability testing with gosec
- Performance benchmarking with time constraints
- Error condition validation
- Table-driven test patterns with zero magic numbers

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
    ErrCodeGitOperation = 1008
    ErrCodeBranchOperation = 1009
    ErrCodeMergeRequestOperation = 1010
)
```

### Error Categories
- **Configuration Errors**: Invalid flags, missing tokens, malformed inputs
- **API Errors**: GitLab connectivity, authentication, rate limiting
- **File System Errors**: Missing files, permission issues, disk space
- **Security Errors**: Path traversal attempts, unauthorized file access
- **Validation Errors**: YAML syntax, tag format, project existence
- **Workflow Errors**: Process failures, timeout conditions
- **Git Errors**: Branch operations, merge conflicts

## Testing Strategy

### Unit Tests ✅
- ✅ Individual component testing with mocked dependencies
- ✅ Error condition simulation and validation
- ✅ Edge case validation with comprehensive test cases
- ✅ Security vulnerability testing (path traversal, validation)
- ✅ Performance benchmarking for all critical operations

### Test Coverage Status
- ✅ **GitLab Package**: Comprehensive test coverage
  - ✅ `client.go`: Client creation, configuration, timeout handling
  - ✅ `projects.go`: Project validation, path resolution, security checks
  - ✅ `branches.go`: Branch validation, name generation, operations
- ✅ **Version Package**: Complete version management testing
- ✅ **YAML Package**: YAML processing and security validation
- ✅ **Workflow Package**: Workflow orchestration and error handling
- ✅ **Logger Package**: Debug logging and structured output

### Integration Tests
- GitLab API interaction testing (requires real GitLab instance)
- End-to-end workflow validation
- Security validation testing
- YAML processing accuracy

### Security Tests ✅
- ✅ Path traversal attack prevention testing
- ✅ File access boundary validation
- ✅ Input sanitization verification
- ✅ Invalid character and format validation

### Test Implementation Standards ✅
All tests follow strict quality standards:
- ✅ **Zero magic numbers**: All numeric literals as named constants
- ✅ **Table-driven tests**: Comprehensive test case coverage
- ✅ **Benchmark tests**: Performance validation for critical operations
- ✅ **English-only documentation**: Professional, clear comments
- ✅ **Error handling**: Comprehensive validation of error conditions

### Test Constants
```go
const (
    TestProjectID = "test/project"
    TestGitLabToken = "test-token"
    TestGitLabURL = "https://gitlab.example.com"
    TestFilePath = "deployment.yaml"
    TestNewTag = "v1.2.3"
    TestOldTag = "v1.0.0"
    TestTargetBranch = "main"
    TestBranchName = "update-tag/v1.2.3"
    TestTimeout = 5000 // milliseconds
    MaxTestRetries = 3
)
```

## Security Considerations

- Secure token handling with environment variable support
- Input validation and sanitization
- **Path traversal protection for all file operations**
- Temporary file cleanup with secure permissions
- Rate limiting compliance
- Audit logging for sensitive operations
- Build-time security scanning with gosec
- Dependency vulnerability checking with govulncheck
- SBOM generation for supply chain security

## Performance Requirements

- Repository operations under 30 seconds for typical repositories
- API response handling within 5 seconds per request
- Memory usage under 100MB for standard operations
- File operations with security validation under 100ms
- Concurrent operation support for batch processing

## Quality Assurance

The project maintains high quality standards through:

- **Linting**: golangci-lint with comprehensive rule set including staticcheck
- **Security Scanning**: gosec for vulnerability detection
- **Static Analysis**: staticcheck for code quality and performance
- **Error Checking**: errcheck for unhandled errors
- **Vulnerability Scanning**: govulncheck for dependency issues
- **SBOM Generation**: Software Bill of Materials tracking
- **Coverage Analysis**: Test coverage reporting with HTML output
- **Benchmarking**: Performance validation for critical paths

## Monitoring and Observability

- Structured logging with configurable levels
- Operation timing and performance metrics
- API call tracking and rate limit monitoring
- Error categorization and reporting
- Debug mode with comprehensive request/response logging
- Security event logging and monitoring

## Dependencies

### Core Dependencies
- `gitlab.com/gitlab-org/api/client-go v0.130.1`: Official GitLab API client
- `github.com/spf13/cobra v1.9.1`: CLI framework
- `github.com/spf13/viper v1.20.1`: Configuration management
- `github.com/sirupsen/logrus v1.9.3`: Structured logging
- `gopkg.in/yaml.v3 v3.0.1`: YAML processing

### Development Dependencies
- Build and quality tools via Makefile
- Security scanners (gosec, govulncheck)
- Linters (golangci-lint with staticcheck)
- SBOM generation (syft)

## Build System

### Makefile Targets
- **Building**: `build`, `build-cross`, `build-debug`
- **Testing**: `test`, `test-coverage`, `benchmark`
- **Quality**: `lint`, `staticcheck`, `security-scan`, `vuln-check`
- **Packaging**: `package-rpm`, `package-deb`, `package-tarball`
- **Documentation**: `docs`, `sbom-generate`

### Version Management
- Build-time version injection via linker flags
- Git commit hash and build date embedding
- Semantic versioning support with bump commands

## Future Enhancements

### Next Priority: Enhanced YAML Processing

**Remaining improvements:**

#### High Priority (Core Features)
1. **Advanced YAML Parsing**: Replace simple string replacement with proper YAML tree manipulation
2. **Multiple Tag Support**: Update multiple tags in single operation
3. **Complex YAML Structures**: Handle nested tags and arrays
4. **YAML Validation**: Enhanced syntax and structure validation

#### Medium Priority (Workflow Enhancements)
1. **Pipeline Integration**: Monitor and wait for CI/CD pipelines
2. **Auto-merge Logic**: Implement smart auto-merge with conditions
3. **Conflict Resolution**: Handle merge conflicts automatically
4. **Batch Operations**: Support multiple files and projects simultaneously

#### Lower Priority (Advanced Features)
1. **Webhook Integration**: Trigger updates via webhooks
2. **Approval Workflows**: Integration with GitLab approval rules
3. **Audit Logging**: Comprehensive audit trail
4. **Metrics and Monitoring**: Prometheus metrics
5. **High Availability**: Support for multiple GitLab instances

### Implementation Guidelines

All new features must follow established patterns:
- **Zero magic numbers**: All numeric literals as named constants
- **Comprehensive testing**: Unit tests, benchmarks, security validation
- **Official GitLab API**: Use official client library features
- **Security focus**: Path validation, input sanitization
- **English-only documentation**: Professional, clear comments
- **Error handling**: Structured error types with specific codes

### Expected Enhancement Goals
- **YAML Processing**: 100% compatibility with YAML 1.2 specification
- **Performance**: Sub-second response for typical operations
- **Reliability**: 99.9% success rate for valid operations
- **Security**: Zero security vulnerabilities in static analysis
- **Maintainability**: 90%+ test coverage for all new features

The architecture provides a solid foundation for these enhancements while maintaining backward compatibility and following Go best practices. 