# Go Tag Updater Architecture

This document describes the architecture and design of the go-tag-updater application, a professional CLI tool for automated YAML tag updates in GitLab repositories through merge request workflows.

## Overview

The go-tag-updater is a production-ready CLI tool designed to automate tag updates in YAML files within GitLab repositories. The application leverages the official [gitlab.com/gitlab-org/api/client-go](https://pkg.go.dev/gitlab.com/gitlab-org/api/client-go) library to interact with GitLab's REST API, providing enterprise-grade automation capabilities.

## Architecture Principles

### Clean Architecture
The application follows clean architecture principles with clear separation of concerns:
- **Presentation Layer**: CLI interface with Cobra framework
- **Application Layer**: Workflow orchestration and business logic
- **Domain Layer**: Core entities and business rules
- **Infrastructure Layer**: GitLab API integration and external services

### Security-First Design
- **Zero Tolerance for Magic Numbers**: All numeric literals are named constants
- **Path Traversal Protection**: Comprehensive file path validation with cross-platform support
- **Input Validation**: Rigorous validation of all user inputs and API responses
- **Secure Error Handling**: No sensitive information disclosure in error messages

### Quality Standards
- **Professional Documentation**: All comments and documentation in English
- **Comprehensive Testing**: Unit tests with high coverage for critical components
- **Static Analysis**: Multiple linters and security scanners
- **Cross-Platform Compatibility**: Full support for Unix, Linux, macOS, and Windows

## Architecture Components

### 1. Command Layer (`cmd/go-tag-updater/`)

**Files:**
- `main.go`: Application entry point with version injection and exit codes
- `root.go`: Cobra command definition and CLI argument parsing

**Responsibilities:**
- CLI framework integration with Cobra and Viper
- Configuration management with environment variable support
- Version information display with build metadata
- Professional error handling and exit codes

**Constants and Standards:**
```go
const (
    ExitCodeSuccess = 0
    ExitCodeError   = 1
    ExitCodeConfig  = 2
)
```

### 2. Configuration Layer (`internal/config/`)

**Files:**
- `config.go`: Configuration structures and validation

**Key Constants:**
```go
const (
    DefaultConfigFile = "go-tag-updater.yaml"
    EnvPrefix = "GO_TAG_UPDATER"
    DefaultTimeout = 30 * time.Second
    DefaultRetryCount = 3
    DefaultRateLimitRPS = 10
)
```

**Configuration Types:**
- **Config**: File-based configuration with YAML support
- **CLIConfig**: Command-line specific configuration
- **GitLabConfig**: GitLab API specific settings with timeouts and rate limiting
- **PerformanceConfig**: Performance tuning parameters

**Responsibilities:**
- CLI flag parsing and validation with comprehensive error handling
- Environment variable support with `GO_TAG_UPDATER_` prefix
- Configuration file loading with fallback mechanisms
- Input sanitization and security validation

### 3. GitLab API Layer (`internal/gitlab/`)

#### Core Client (`client.go`)
**Constants:**
```go
const (
    DefaultTimeout = 30 * time.Second
    MaxRetryAttempts = 3
    DefaultBaseURL = "https://gitlab.com"
    HealthCheckEndpoint = "/api/v4/user"
)
```

**Responsibilities:**
- Wrapper around official GitLab client with enhanced error handling
- Health checks and connection validation
- Project ID resolution supporting both numeric IDs and path-based identifiers
- Rate limiting and timeout configuration

#### File Operations (`files.go`)
**Constants:**
```go
const (
    DefaultCommitMessage = "Update file via go-tag-updater"
    MaxFileSize = 1024 * 1024 // 1MB
    DefaultBranch = "main"
)
```

**FileManager Components:**
- Repository file CRUD operations with Base64 encoding/decoding
- YAML content validation and processing
- File existence checks and history retrieval
- Atomic file operations with rollback capabilities

#### Branch Management (`branches.go`)
**Constants:**
```go
const (
    DefaultBranchPrefix = "feature/"
    UpdateBranchPrefix = "update-tag/"
    MaxBranchNameLength = 100
    MinBranchNameLength = 1
)
```

**BranchManager Components:**
- Git branch lifecycle management
- Unique branch name generation with timestamps
- Branch name validation according to Git standards
- Protected branch detection and handling

#### Project Management (`projects.go`)
**Constants:**
```go
const (
    ProjectsAPIEndpoint = "/api/v4/projects"
    MaxProjectNameLength = 255
    MinProjectIDValue = 1
)
```

**ProjectManager Components:**
- Project resolution from numeric IDs or human-readable paths
- Project validation and existence checks
- Project search and listing capabilities
- Comprehensive path validation with security checks

#### Conflict Detection (`conflicts.go`)
**Constants:**
```go
const (
    MaxConflictCheckAttempts = 15
    ConflictCheckInterval = 10 * time.Second
    DefaultWaitTimeout = 15 * time.Minute
)
```

**ConflictDetector Components:**
- Merge request conflict detection and prevention
- Intelligent conflict analysis based on patterns
- Waiting mechanisms for conflict resolution
- Comprehensive conflict reporting

#### Merge Requests (`merge_requests_simple.go`)
**SimpleMergeRequestManager Components:**
- Create, retrieve, and list merge requests
- Simplified options structure compatible with GitLab API
- Status monitoring and lifecycle management
- Returns native GitLab API types for compatibility

### 4. YAML Processing Layer (`internal/yaml/`)

#### Core Processing (`updater.go`)
**Security Constants:**
```go
const (
    // File system security
    DefaultFilePermissions = 0o644
    BackupDirPermissions = 0o750
    MaxBackupFiles = 5
    
    // Path security (24 security constants total)
    PathTraversalPattern = ".."
    WindowsDriveLetterSeparator = ":"
    UnixEtcPath = "/etc/"
    UnixUsrPath = "/usr/"
    UnixBinPath = "/bin/"
    UnixSbinPath = "/sbin/"
    WindowsSystemPath = "/windows/system32/"
    WindowsProgramFilesPath = "/program files/"
    
    // Safe directory prefixes
    UnixTempDir = "/tmp/"
    UnixVarTempDir = "/var/tmp/"
    WindowsCTempDir = "C:/temp/"
    WindowsCTmpDir = "C:/tmp/"
)
```

**Security Features:**
- **Cross-Platform Path Validation**: Comprehensive protection against path traversal attacks
- **System Directory Blocking**: Prevents access to Unix (`/etc/`, `/usr/`, `/bin/`, `/sbin/`) and Windows (`System32`, `Program Files`) system directories
- **Relative Path Security**: Blocks malicious patterns like `etc/passwd`, `usr/bin/bash`
- **Drive Letter Detection**: Windows-specific security for drive letter paths
- **Safe Directory Validation**: Only allows absolute paths in designated safe directories

**Updater Components:**
- **validateAndCleanFilePath**: Multi-layer security validation (refactored from complex function)
- **checkSuspiciousSystemPaths**: System directory access prevention
- **checkWindowsSystemPaths**: Windows-specific security checks
- **validateAbsolutePath**: Absolute path validation with safe directory enforcement

#### YAML Parser (`parser.go`)
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
- YAML file parsing and validation with proper error handling
- Tag value updates with formatting preservation
- Syntax validation before and after updates
- Backup creation and restoration mechanisms

### 5. Workflow Layer (`internal/workflow/`)

#### Simple Workflow (`simple_workflow.go`)
**Constants:**
```go
const (
    PreviewContentMaxLength = 500
    TempFilePermissions = 0o600
)
```

**SimpleTagUpdater Components:**
- Complete workflow orchestration with step-by-step execution
- File existence validation with security checks
- YAML content processing with temporary file handling
- Branch name generation (automatic or custom)
- Dry-run mode with content preview (500 character limit)
- Comprehensive error handling with cleanup and rollback

**Workflow Steps:**
1. Configuration validation and security checks
2. GitLab connectivity and project validation
3. File existence and accessibility verification
4. YAML content parsing and tag detection
5. Unique branch name generation with conflict avoidance
6. Branch creation with proper Git conventions
7. File content update with atomic operations
8. Merge request creation with detailed descriptions
9. Status monitoring and reporting
10. Cleanup and resource management

### 6. Version Management (`internal/version/`)

**Constants:**
```go
const (
    ShortCommitHashLength = 7
    UnknownValue = "unknown"
)
```

**Build Variables (Injected at Build Time):**
```go
var (
    Version = "dev"
    Commit = "unknown"
    Date = "unknown"
    BuiltBy = "unknown"
    BuildNumber = "0"
)
```

**Responsibilities:**
- Version information display with comprehensive build metadata
- Build-time variable management via linker flags
- Runtime environment details (Go version, OS, architecture)
- CLI version flag handling with formatted output

### 7. Logging (`internal/logger/`)

**Constants:**
```go
const (
    MaxLogFileSize = 100 * 1024 * 1024 // 100MB
    DefaultLogFilePerm = 0o644
    LogFileBufferSize = 4096
    MaxLogRotationCount = 10
    LogRotationAge = 7 * 24 * time.Hour // 7 days
)
```

**Responsibilities:**
- Structured logging with configurable debug mode and log levels
- Operation tracking with context fields (project ID, operation type)
- Error logging with stack traces and categorization
- GitLab API request/response logging in debug mode
- Log rotation and file management

### 8. Error Handling (`pkg/errors/`)

**Error Categories and Codes:**
```go
const (
    // Validation errors (1001-1003)
    ErrCodeInvalidProject = 1001
    ErrCodeFileNotFound = 1002
    ErrCodeInvalidYAML = 1003
    
    // Operation errors (1004-1007)
    ErrCodeGitOperation = 1004
    ErrCodeAPIError = 1005
    ErrCodeMergeConflict = 1006
    ErrCodeValidation = 1007
    
    // System errors (1008-1010)
    ErrCodeConfiguration = 1008
    ErrCodeNetworkError = 1009
    ErrCodeAuthError = 1010
    
    // Security constants
    MaxErrorMessageLength = 500
    MaxErrorContextLength = 200
)
```

**Error Categories:**
- **Configuration Errors**: Invalid flags, missing tokens, malformed inputs
- **API Errors**: GitLab connectivity, authentication, rate limiting
- **File System Errors**: Missing files, permission issues, disk space
- **Security Errors**: Path traversal attempts, unauthorized file access
- **Validation Errors**: YAML syntax, tag format, project existence
- **Workflow Errors**: Process failures, timeout conditions
- **Git Errors**: Branch operations, merge conflicts

## Security Architecture

### Path Security System
The application implements a comprehensive multi-layer path security system:

#### Layer 1: Path Traversal Prevention
- Detects and blocks `..` patterns in file paths
- Validates path structure before processing
- Cross-platform path normalization

#### Layer 2: System Directory Protection
- **Unix System Paths**: Blocks access to `/etc/`, `/usr/`, `/bin/`, `/sbin/`
- **Windows System Paths**: Blocks access to `System32`, `Program Files`
- **Relative Path Security**: Prevents `etc/passwd`, `usr/bin/bash` patterns
- **Drive Letter Detection**: Windows-specific security for `C:\`, `D:\` paths

#### Layer 3: Safe Directory Validation
- **Allowed Absolute Paths**: Only `/tmp/`, `/var/tmp/` on Unix
- **Windows Temp Directories**: `C:/temp/`, `C:/tmp/`, `D:/temp/`, `D:/tmp/`
- **Cross-Platform Compatibility**: Handles filesystem differences correctly

#### Layer 4: File System Security
- **Secure Permissions**: Files created with `0o644`, directories with `0o750`
- **Atomic Operations**: Temporary files with `0o600` permissions
- **Backup Management**: Automatic cleanup with configurable retention

### Authentication and Authorization
- **Token-Based Authentication**: GitLab personal access tokens
- **Configurable Instances**: Support for GitLab.com and self-hosted instances
- **Health Check Validation**: Connectivity and permission verification
- **Rate Limiting**: Configurable request rate limiting to prevent abuse

## Testing Strategy

### Current Test Coverage
- **internal/version**: 100% coverage - Complete version management testing
- **internal/logger**: 79.2% coverage - Comprehensive logging functionality
- **internal/yaml**: 49.4% coverage - Core YAML processing and security
- **internal/workflow**: 33.1% coverage - Workflow orchestration
- **internal/gitlab**: 31.2% coverage - GitLab API integration

### Test Types Implemented

#### Unit Tests
- **Security Testing**: 15+ malicious path patterns, 8+ allowed path patterns
- **Cross-Platform Testing**: Windows file permissions, Unix path validation
- **Error Condition Testing**: Comprehensive error scenario coverage
- **Edge Case Validation**: Boundary conditions and invalid inputs
- **Performance Benchmarking**: All critical operations benchmarked

#### Table-Driven Tests
Following Go best practices with zero magic numbers:
```go
tests := []struct {
    name     string
    input    string
    expected string
    wantErr  bool
}{
    {
        name:     "valid relative path",
        input:    "config/app.yaml",
        expected: "config/app.yaml",
        wantErr:  false,
    },
    // ... more test cases
}
```

#### Security Tests
- **Path Traversal**: `../../../etc/passwd`, `..\\..\\..\\windows\\system32\\config\\sam`
- **System Directory Access**: `/etc/passwd`, `/usr/bin/bash`, `C:\Windows\System32\config\SAM`
- **Relative Path Attacks**: `etc/passwd`, `usr/local/bin/test`
- **Cross-Platform Validation**: Windows and Unix path handling

### Quality Assurance Pipeline

#### Code Quality Tools
- **golangci-lint**: Comprehensive linting with 50+ linters including staticcheck
- **staticcheck**: Advanced static analysis for performance and correctness
- **errcheck**: Unchecked error detection with exclusion list
- **gosec**: Security vulnerability scanning (0 issues)
- **govulncheck**: Dependency vulnerability scanning (no vulnerabilities)

#### Security Scanning
- **SARIF Reports**: Structured security analysis results
- **SBOM Generation**: Software Bill of Materials with Syft
- **Dependency Analysis**: Automated vulnerability detection
- **Supply Chain Security**: Complete dependency tracking

#### Performance Monitoring
- **Benchmark Tests**: Performance validation for critical paths
- **Memory Profiling**: Memory usage optimization
- **CPU Profiling**: Performance bottleneck identification
- **Coverage Analysis**: HTML coverage reports with detailed metrics

## GitLab API Integration

### Official Client Library Integration
The application uses the official GitLab Go client library:
- **Library**: `gitlab.com/gitlab-org/api/client-go v0.130.1`
- **Features**: Complete GitLab REST API v4 support
- **Authentication**: Token-based with configurable instances
- **Error Handling**: Comprehensive API error categorization

### API Services Utilized

#### Projects Service
- Project information retrieval and validation
- Path-based project resolution (`group/subgroup/project`)
- Project search and listing capabilities
- Access permission validation

#### RepositoryFiles Service
- File CRUD operations with Base64 encoding/decoding
- Content validation and syntax checking
- Commit message customization and author attribution
- Branch-specific file operations

#### Branches Service
- Branch creation from any Git reference
- Branch listing with search and pagination
- Branch existence validation and conflict detection
- Protected branch handling and validation

#### MergeRequests Service
- Merge request creation with comprehensive metadata
- Status monitoring and lifecycle management
- Conflict detection and resolution workflows
- Integration with CI/CD pipeline status

#### Users Service
- Authentication health checks and token validation
- User information retrieval for attribution
- Permission and access level validation

### API Capabilities and Features

#### Advanced File Operations
- **Content Processing**: Automatic Base64 encoding/decoding
- **Atomic Updates**: Create or update files with conflict detection
- **History Tracking**: File change history and blame information
- **Branch Isolation**: File operations scoped to specific branches

#### Intelligent Branch Management
- **Conflict Avoidance**: Automatic unique branch name generation
- **Validation**: Git branch naming convention compliance
- **Cleanup**: Automatic branch cleanup on workflow completion
- **Protection**: Protected branch detection and handling

#### Sophisticated Merge Request Workflows
- **Metadata Management**: Comprehensive title and description generation
- **Status Tracking**: Real-time merge request status monitoring
- **Conflict Resolution**: Intelligent conflict detection and reporting
- **Integration**: CI/CD pipeline integration and status monitoring

## Performance Characteristics

### Operational Performance
- **Repository Operations**: Under 30 seconds for typical repositories
- **API Response Handling**: Within 5 seconds per request with retries
- **Memory Usage**: Under 100MB for standard operations
- **File Operations**: Security validation under 100ms per file
- **Concurrent Support**: Configurable concurrent operation limits

### Scalability Features
- **Rate Limiting**: Configurable requests per second (default: 10 RPS)
- **Connection Pooling**: HTTP client connection reuse
- **Pagination Support**: Large result set handling
- **Batch Operations**: Multiple file processing capabilities
- **Resource Management**: Automatic cleanup and resource limits

## Configuration Management

### CLI Configuration
**Required Parameters:**
- `--project-id`: GitLab project ID or path (`group/project`)
- `--file`: Target YAML file path in repository
- `--new-tag`: New tag value to set
- `--token`: GitLab access token

**Optional Parameters:**
- `--branch-name`: Custom branch name (auto-generated if not provided)
- `--target-branch`: Target branch for merge request (default: `main`)
- `--gitlab-url`: GitLab instance URL (default: `https://gitlab.com`)
- `--dry-run`: Preview mode without making changes
- `--debug`: Enable detailed logging and API tracing
- `--timeout`: Operation timeout (default: 30s)

### Environment Variables
All CLI flags support environment variables with `GO_TAG_UPDATER_` prefix:
```bash
export GO_TAG_UPDATER_PROJECT_ID="mygroup/myproject"
export GO_TAG_UPDATER_FILE="config/deployment.yaml"
export GO_TAG_UPDATER_NEW_TAG="v1.2.3"
export GO_TAG_UPDATER_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
```

### Configuration Files
**Supported Formats:**
- `go-tag-updater.yaml` - Primary configuration file
- `go-tag-updater.yml` - Alternative YAML format
- Environment-specific overrides

**Configuration Structure:**
```yaml
gitlab:
  base_url: "https://gitlab.example.com"
  token: "${GITLAB_TOKEN}"
  timeout: "30s"
  retry_count: 3

defaults:
  target_branch: "main"
  branch_prefix: "update-tag/"
  auto_merge: false
  wait_previous_mr: true

performance:
  max_concurrent_requests: 5
  request_timeout: "10s"
  buffer_size: 4096

logging:
  level: "info"
  format: "json"
  enable_file: false
```

## Build System and Quality Assurance

### Build Process
**Requirements:**
- Go 1.24+ (specified in `.go-version`)
- CGO disabled for static binaries
- Multi-architecture support (amd64, arm64)

**Build Metadata Injection:**
```bash
go build -ldflags="-X 'internal/version.Version=${VERSION}' \
                   -X 'internal/version.Commit=${COMMIT}' \
                   -X 'internal/version.Date=${DATE}' \
                   -X 'internal/version.BuiltBy=${BUILT_BY}'"
```

**Cross-Platform Targets:**
- Linux (amd64, arm64)
- macOS (amd64, arm64) 
- Windows (amd64, arm64)

### Quality Gates
**Code Standards:**
- **gofmt**: Code formatting compliance
- **goimports**: Import organization and cleanup
- **go vet**: Static analysis and common errors
- **golangci-lint**: 50+ linters including complexity analysis
- **staticcheck**: Advanced static analysis for performance

**Security Standards:**
- **gosec**: Security vulnerability scanning (0 issues)
- **govulncheck**: Known vulnerability detection (no vulnerabilities)
- **errcheck**: Unchecked error detection with exclusions
- **Path Security**: Comprehensive path traversal protection

**Documentation Standards:**
- **English-only**: All comments and documentation in English
- **Professional tone**: No emojis, casual language, or exclamation marks
- **Technical accuracy**: Precise technical terminology
- **Comprehensive coverage**: All public APIs documented

### Continuous Integration
**GitHub Actions Workflows:**
- **CI/CD Pipeline**: Automated testing across Go 1.22-1.24
- **Cross-Platform Testing**: Ubuntu, macOS, Windows matrix
- **Security Scanning**: CodeQL, Trivy, Nancy, Gosec, Govulncheck
- **Release Automation**: Cross-platform binaries with Cosign signing
- **Container Building**: Multi-stage Docker builds with SBOM generation

## Dependencies and Supply Chain Security

### Core Dependencies
```go
// Core GitLab integration
gitlab.com/gitlab-org/api/client-go v0.130.1

// CLI framework
github.com/spf13/cobra v1.9.1
github.com/spf13/viper v1.20.1

// Logging and utilities
github.com/sirupsen/logrus v1.9.3
gopkg.in/yaml.v3 v3.0.1
```

### Development Dependencies
- **Build Tools**: Comprehensive Makefile with 50+ targets
- **Quality Tools**: golangci-lint, staticcheck, gosec, govulncheck
- **Security Tools**: Syft for SBOM generation, Trivy for container scanning
- **Documentation**: godoc for API documentation generation

### Supply Chain Security
- **SBOM Generation**: Complete Software Bill of Materials
- **Dependency Scanning**: Automated vulnerability detection
- **License Compliance**: Open source license tracking
- **Signature Verification**: Cosign keyless signing for releases

## Future Enhancements

### Phase 2: Advanced YAML Processing
1. **Structured YAML Parsing**: Replace string replacement with proper YAML AST manipulation
2. **Multiple Tag Support**: Update multiple tags in single operation
3. **Conditional Updates**: Update based on current tag values or conditions
4. **Template Support**: YAML template processing with variable substitution

### Phase 3: Enterprise Integration
1. **Pipeline Integration**: Wait for CI/CD pipeline completion before merge
2. **Approval Workflows**: Integration with GitLab approval rules and policies
3. **Webhook Support**: Trigger updates via GitLab webhooks
4. **Audit Logging**: Comprehensive audit trail with compliance reporting

### Phase 4: Advanced Automation
1. **Batch Operations**: Multiple files and projects in single operation
2. **Rollback Capabilities**: Automatic rollback on failure or validation errors
3. **Conflict Resolution**: Intelligent merge conflict resolution
4. **High Availability**: Support for multiple GitLab instances and failover

### Phase 5: Monitoring and Observability
1. **Metrics Export**: Prometheus metrics for operation monitoring
2. **Distributed Tracing**: OpenTelemetry integration for request tracing
3. **Health Monitoring**: Comprehensive health checks and status reporting
4. **Performance Analytics**: Operation performance analysis and optimization

## Conclusion

The go-tag-updater architecture provides a robust, secure, and scalable foundation for automated YAML tag updates in GitLab environments. The design emphasizes:

**Security First**: Comprehensive path validation, system directory protection, and secure file operations across all platforms.

**Quality Assurance**: Extensive testing, static analysis, and security scanning ensure production readiness.

**Professional Standards**: English-only documentation, zero magic numbers, and comprehensive error handling follow enterprise development practices.

**Cross-Platform Compatibility**: Full support for Unix, Linux, macOS, and Windows with platform-specific optimizations.

**GitLab Integration**: Deep integration with GitLab's official API client provides access to the complete GitLab feature set.

The modular architecture enables easy extension and maintenance while maintaining backward compatibility and following Go best practices. The comprehensive security model, quality assurance pipeline, and professional documentation standards make this suitable for enterprise deployment and critical automation workflows. 