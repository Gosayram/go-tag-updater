# Go Tag Updater Architecture

This document describes the architecture and design of the go-tag-updater application, which uses the official GitLab Go API client library.

## Overview

The go-tag-updater is a CLI tool designed to automate tag updates in YAML files within GitLab repositories. The application leverages the powerful [gitlab.com/gitlab-org/api/client-go](https://pkg.go.dev/gitlab.com/gitlab-org/api/client-go) library to interact with GitLab's REST API.

## Architecture Components

### 1. Command Layer (`cmd/go-tag-updater/`)

- **main.go**: Entry point of the application
- **root.go**: Cobra command definition and CLI argument parsing
- Uses Viper for configuration management with environment variable and flag support

### 2. Configuration Layer (`internal/config/`)

- **config.go**: Configuration structure and Viper integration
- Supports all required parameters: file path, new tag, project ID, GitLab token, etc.
- Environment variable bindings with `GOTAG_` prefix

### 3. GitLab API Layer (`internal/gitlab/`)

#### Core Client (`client.go`)
- Wrapper around the official GitLab client
- Health checks and connection management
- Project ID resolution (supports both numeric IDs and path-based identifiers)

#### File Operations (`files.go`)
- **FileManager**: Handles repository file operations
- Uses GitLab's RepositoryFiles API for CRUD operations
- Base64 content encoding/decoding
- YAML tag update functionality with simple string replacement
- File existence checks and history retrieval

#### Branch Management (`branches.go`)
- **BranchManager**: Handles Git branch operations
- Branch creation, deletion, and listing
- Unique branch name generation with timestamps
- Branch name validation according to Git rules
- Protected branch detection

#### Merge Requests (`merge_requests_simple.go`)
- **SimpleMergeRequestManager**: Basic merge request operations
- Create, retrieve, and list merge requests
- Simplified options structure compatible with API
- Returns native GitLab API types

### 4. Workflow Layer (`internal/workflow/`)

#### Simple Workflow (`simple_workflow.go`)
- **SimpleTagUpdater**: Orchestrates the complete tag update process
- Step-by-step workflow execution:
  1. File existence validation
  2. Unique branch name generation
  3. Branch creation
  4. File content update
  5. Merge request creation
- Dry-run support for testing
- Error handling with cleanup on failure

### 5. Error Handling (`pkg/errors/`)

- **types.go**: Custom error types with specific codes
- Structured error handling for different failure scenarios:
  - Validation errors (1001-1003)
  - File system errors (1004-1005)
  - API errors (1006-1007)
  - Git operation errors (1008-1010)

### 6. Logging (`internal/logger/`)

- **debug.go**: Structured logging with configurable debug mode
- Log levels: Info, Warn, Error, Debug
- Simple interface compatible with the workflow

## GitLab API Integration

### Client Library Features Used

The application leverages several GitLab API services:

1. **Projects Service**: Project information and resolution
2. **RepositoryFiles Service**: File CRUD operations
3. **Branches Service**: Branch management
4. **MergeRequests Service**: Merge request operations
5. **Users Service**: Authentication health checks

### API Capabilities

#### File Operations
- Get file content with automatic base64 decoding
- Update existing files or create new ones
- Commit messages and author information
- Branch-specific operations

#### Branch Operations
- Create branches from any reference
- List and search branches
- Branch existence checks
- Branch name validation

#### Merge Request Operations
- Create merge requests with title and description
- Retrieve merge request details
- List merge requests with filtering

### Authentication

- Token-based authentication
- Configurable GitLab instance URL (defaults to gitlab.com)
- Health check validation

## Configuration

### CLI Flags

```bash
--file          Target YAML file path
--new-tag       New tag value to set
--project-id    GitLab project ID or path
--token         GitLab access token
--branch-name   Custom branch name (optional)
--target-branch Target branch (default: main)
--gitlab-url    GitLab instance URL (default: https://gitlab.com)
--dry-run       Simulate operations without making changes
--debug         Enable debug logging
--auto-merge    Enable auto-merge (placeholder)
--wait          Wait for previous MRs (placeholder)
```

### Environment Variables

All flags can be set via environment variables with `GOTAG_` prefix:
- `GOTAG_FILE`
- `GOTAG_NEW_TAG`
- `GOTAG_PROJECT_ID`
- `GOTAG_TOKEN`
- etc.

## Build and Quality Assurance

### Build Process
- Go 1.21+ required
- CGO disabled for static binaries
- Build metadata injection (version, commit, date, builder)
- Multi-architecture support

### Quality Gates
- **gofmt**: Code formatting
- **goimports**: Import organization
- **go vet**: Static analysis
- **golangci-lint**: Comprehensive linting with multiple linters
- **staticcheck**: Advanced static analysis
- **errcheck**: Unchecked error detection
- **gosec**: Security vulnerability scanning
- **govulncheck**: Known vulnerability detection
- **syft**: Software Bill of Materials (SBOM) generation

### Code Standards
- Zero tolerance for magic numbers (all numeric literals as named constants)
- English-only comments and documentation
- Professional documentation standards
- Error handling for all operations
- Structured logging

## Testing Strategy

### Current State
- Basic validation and configuration testing
- Dry-run mode for safe testing
- Health checks for connectivity validation

### Future Enhancements
- Unit tests for each manager component
- Integration tests with GitLab API
- Mock testing for offline development
- Table-driven tests following Go best practices

## Security Considerations

- Token-based authentication with secure storage
- HTTPS-only communication
- Input validation and sanitization
- No sensitive information in logs
- Secure error handling without information disclosure

## Performance Considerations

- HTTP client with configurable timeouts
- Connection reuse for multiple API calls
- Pagination support for large result sets
- Rate limiting awareness (configurable)
- Efficient base64 encoding/decoding

## Dependencies

### Core Dependencies
- `gitlab.com/gitlab-org/api/client-go`: Official GitLab API client
- `github.com/spf13/cobra`: CLI framework
- `github.com/spf13/viper`: Configuration management

### Development Dependencies
- `github.com/golangci/golangci-lint`: Linting
- Various security and analysis tools

## Future Enhancements

### Phase 2: Advanced Features
1. **Complex YAML Parsing**: Replace simple string replacement with proper YAML parsing
2. **Pipeline Integration**: Monitor and wait for CI/CD pipelines
3. **Auto-merge Logic**: Implement smart auto-merge with conditions
4. **Conflict Resolution**: Handle merge conflicts automatically
5. **Batch Operations**: Support multiple files and projects
6. **Rollback Capability**: Implement rollback mechanisms

### Phase 3: Enterprise Features
1. **Webhook Integration**: Trigger updates via webhooks
2. **Approval Workflows**: Integration with GitLab approval rules
3. **Audit Logging**: Comprehensive audit trail
4. **Metrics and Monitoring**: Prometheus metrics
5. **High Availability**: Support for multiple GitLab instances

## Conclusion

The current architecture provides a solid foundation for automated tag updates using GitLab's official API. The modular design allows for easy extension and maintenance while following Go best practices and maintaining high code quality standards.

The use of the official GitLab client library ensures compatibility and access to the full range of GitLab features, providing a reliable base for future enhancements. 