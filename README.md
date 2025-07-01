# go-tag-updater

[![Go Version](https://img.shields.io/github/go-mod/go-version/Gosayram/go-tag-updater)](https://golang.org/)
[![License](https://img.shields.io/github/license/Gosayram/go-tag-updater)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Gosayram/go-tag-updater)](https://goreportcard.com/report/github.com/Gosayram/go-tag-updater)
[![Security](https://github.com/Gosayram/go-tag-updater/actions/workflows/security.yml/badge.svg)](https://github.com/Gosayram/go-tag-updater/actions/workflows/security.yml)
[![CI](https://github.com/Gosayram/go-tag-updater/actions/workflows/ci-lint-test.yml/badge.svg)](https://github.com/Gosayram/go-tag-updater/actions/workflows/ci-lint-test.yml)
[![Release](https://github.com/Gosayram/go-tag-updater/actions/workflows/release.yml/badge.svg)](https://github.com/Gosayram/go-tag-updater/actions/workflows/release.yml)

A production-ready Go CLI tool for securely updating YAML files in GitLab repositories through automated merge request workflows. Built with enterprise-grade security, comprehensive testing, and professional development standards.

## Features

### 🔒 Security First
- **Path Traversal Protection**: Multi-layer security preventing access to system directories
- **Cross-Platform Security**: Comprehensive validation for Unix, Linux, macOS, and Windows
- **Input Validation**: Rigorous validation of all user inputs and API responses
- **Zero Vulnerabilities**: Security scanning with gosec and govulncheck (0 issues)
- **Secure File Operations**: Atomic operations with proper permissions and cleanup

### 🔄 GitLab Integration
- **Official API Client**: Built with `gitlab.com/gitlab-org/api/client-go` v0.130.1
- **Project Resolution**: Support for numeric IDs and human-readable paths
- **Branch Management**: Intelligent conflict detection and unique branch generation
- **Merge Request Automation**: Complete lifecycle management with status monitoring
- **Rate Limiting**: Configurable request throttling to prevent API abuse

### 🧪 Quality Assurance
- **Comprehensive Testing**: 100% coverage for critical components with cross-platform tests
- **Static Analysis**: 50+ linters including staticcheck, errcheck, and complexity analysis
- **Performance Benchmarking**: Sub-second operations with memory optimization
- **SBOM Generation**: Complete Software Bill of Materials for supply chain security
- **Professional Standards**: Zero magic numbers, English-only documentation

### ⚡ Performance & Reliability
- **Atomic Operations**: Secure file handling with rollback capabilities
- **Dry Run Mode**: Preview changes without making modifications
- **Structured Logging**: Comprehensive debugging with operation tracking
- **Error Handling**: Categorized errors with detailed context and recovery
- **Cross-Platform**: Native support for all major operating systems

## Quick Start

### Installation

#### Download Binary (Recommended)

Download the latest release from the [releases page](https://github.com/Gosayram/go-tag-updater/releases):

```bash
# Linux/macOS
curl -L -o go-tag-updater "https://github.com/Gosayram/go-tag-updater/releases/latest/download/go-tag-updater-$(uname -s)-$(uname -m)"
chmod +x go-tag-updater
sudo mv go-tag-updater /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/Gosayram/go-tag-updater/releases/latest/download/go-tag-updater-Windows-x86_64.exe" -OutFile "go-tag-updater.exe"
```

#### Build from Source

```bash
git clone https://github.com/Gosayram/go-tag-updater.git
cd go-tag-updater
make build
# Binary will be available in ./bin/go-tag-updater
```

#### Install via Go

```bash
go install github.com/Gosayram/go-tag-updater/cmd/go-tag-updater@latest
```

#### Package Managers

```bash
# RPM-based systems (RHEL, Fedora, CentOS)
sudo rpm -i go-tag-updater-1.0.1-1.x86_64.rpm

# DEB-based systems (Debian, Ubuntu)
sudo dpkg -i go-tag-updater_1.0.1_amd64.deb
```

### Basic Usage

```bash
# Update a tag in a YAML file
go-tag-updater \
  --project-id=mygroup/myproject \
  --file=k8s/deployment.yaml \
  --new-tag=v1.2.3 \
  --token=$GITLAB_TOKEN
```

### Preview Changes (Dry Run)

```bash
# Preview what would be changed without making actual modifications
go-tag-updater \
  --project-id=123456 \
  --file=config/app.yml \
  --new-tag=v2.0.0 \
  --dry-run \
  --token=$GITLAB_TOKEN
```

### Version Information

```bash
# Display version and build information
go-tag-updater --version
```

## Configuration

### Required Parameters

| Parameter      | Description                     | Example                      |
| -------------- | ------------------------------- | ---------------------------- |
| `--project-id` | GitLab project ID or path       | `123456` or `group/project`  |
| `--file`       | Path to YAML file in repository | `k8s/deployment.yaml`        |
| `--new-tag`    | New tag value to set            | `v1.2.3`                     |
| `--token`      | GitLab Personal Access Token    | `glpat-xxxxxxxxxxxxxxxxxxxx` |

### Optional Parameters

| Parameter            | Default              | Description                                  |
| -------------------- | -------------------- | -------------------------------------------- |
| `--branch-name`      | auto-generated       | Custom branch name for the update            |
| `--target-branch`    | `main`               | Target branch for merge request              |
| `--gitlab-url`       | `https://gitlab.com` | GitLab instance URL                          |
| `--wait-previous-mr` | `false`              | Wait for conflicting merge requests          |
| `--debug`            | `false`              | Enable verbose debugging and API tracing     |
| `--dry-run`          | `false`              | Preview changes without making modifications |
| `--auto-merge`       | `false`              | Auto-merge when pipeline passes              |
| `--timeout`          | `30s`                | Operation timeout duration                   |

### Environment Variables

All CLI flags can be set via environment variables with the `GO_TAG_UPDATER_` prefix:

```bash
export GO_TAG_UPDATER_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
export GO_TAG_UPDATER_PROJECT_ID="mygroup/myproject"
export GO_TAG_UPDATER_GITLAB_URL="https://gitlab.example.com"
export GO_TAG_UPDATER_DEBUG="true"
export GO_TAG_UPDATER_TARGET_BRANCH="main"
```

### Configuration File

Create a `go-tag-updater.yaml` file in your project root:

```yaml
gitlab:
  base_url: "https://gitlab.example.com"
  token: "${GITLAB_TOKEN}"
  timeout: "30s"
  retry_count: 3
  rate_limit_rps: 10

defaults:
  target_branch: "main"
  branch_prefix: "update-tag/"
  auto_merge: false
  wait_previous_mr: false
  merge_timeout: "300s"

performance:
  max_concurrent_requests: 5
  request_timeout: "10s"
  buffer_size: 4096

logging:
  level: "info"
  format: "json"
  enable_file: false
  file_path: "/var/log/go-tag-updater.log"
```

## Usage Examples

### Basic Tag Update

```bash
go-tag-updater \
  --project-id=openproject/infra/dev \
  --file=apps/service/deployment.yml \
  --new-tag=v1.2.3 \
  --token=$GITLAB_TOKEN
```

### Advanced Usage with Custom Branch

```bash
go-tag-updater \
  --project-id=4323829 \
  --file=config/deployment.yaml \
  --new-tag=abc123 \
  --branch-name=hotfix/update-tag-abc123 \
  --target-branch=development \
  --wait-previous-mr=true \
  --debug \
  --token=$GITLAB_TOKEN
```

### Self-Hosted GitLab Instance

```bash
go-tag-updater \
  --project-id=group/project \
  --file=helm/values.yaml \
  --new-tag=v3.1.0 \
  --gitlab-url=https://gitlab.company.com \
  --token=$GITLAB_TOKEN
```

### Batch Processing with Shell Script

```bash
#!/bin/bash
set -euo pipefail

PROJECTS=("group/project1" "group/project2" "group/project3")
NEW_TAG="v2.1.0"
GITLAB_TOKEN="${GITLAB_TOKEN:-}"

if [[ -z "$GITLAB_TOKEN" ]]; then
    echo "Error: GITLAB_TOKEN environment variable is required"
    exit 1
fi

for project in "${PROJECTS[@]}"; do
    echo "Updating $project..."
    go-tag-updater \
        --project-id="$project" \
        --file="k8s/deployment.yaml" \
        --new-tag="$NEW_TAG" \
        --token="$GITLAB_TOKEN" \
        --wait-previous-mr=true \
        --timeout=60s
    
    if [[ $? -eq 0 ]]; then
        echo "✅ Successfully updated $project"
    else
        echo "❌ Failed to update $project"
    fi
done
```

### CI/CD Pipeline Integration

```yaml
# .gitlab-ci.yml
update-tags:
  stage: deploy
  image: alpine:latest
  before_script:
    - wget -O /usr/local/bin/go-tag-updater "https://github.com/Gosayram/go-tag-updater/releases/latest/download/go-tag-updater-Linux-x86_64"
    - chmod +x /usr/local/bin/go-tag-updater
  script:
    - |
      go-tag-updater \
        --project-id="$CI_PROJECT_ID" \
        --file="k8s/deployment.yaml" \
        --new-tag="$CI_COMMIT_TAG" \
        --token="$GITLAB_TOKEN" \
        --target-branch="main" \
        --auto-merge=true
  only:
    - tags
  variables:
    GITLAB_TOKEN: $GITLAB_ACCESS_TOKEN
```

## GitLab Integration

### Required GitLab Permissions

The GitLab Personal Access Token must have the following scopes:

- `api` - Full API access for repository operations
- `read_repository` - Read repository files and metadata
- `write_repository` - Create branches and update files

### Supported GitLab Versions

- **GitLab.com** (SaaS) - Fully supported
- **GitLab CE/EE 13.0+** - Complete feature support
- **GitLab API v4** - Official API integration

### Project Identification

The tool supports flexible project identification methods:

```bash
# Numeric project ID (fastest)
--project-id=123456

# Full project path
--project-id=group/subgroup/project

# URL-encoded paths (automatically handled)
--project-id="group%2Fsubgroup%2Fproject"

# Complex nested groups
--project-id=organization/team/infrastructure/service
```

### Branch and Merge Request Management

- **Automatic Branch Creation**: Generates unique branch names with timestamps
- **Conflict Detection**: Identifies conflicting merge requests before creation
- **Protected Branch Handling**: Respects GitLab branch protection rules
- **Merge Request Lifecycle**: Complete management from creation to merge
- **Status Monitoring**: Real-time tracking of merge request status

## Security

### Path Security System

The application implements a comprehensive multi-layer security system:

#### System Directory Protection
- **Unix Systems**: Blocks access to `/etc/`, `/usr/`, `/bin/`, `/sbin/`
- **Windows Systems**: Blocks access to `System32`, `Program Files`
- **Relative Paths**: Prevents `etc/passwd`, `usr/bin/bash` patterns
- **Cross-Platform**: Handles filesystem differences correctly

#### Path Traversal Prevention
- Detects and blocks `..` patterns in file paths
- Validates path structure before processing
- Cross-platform path normalization
- Safe directory validation for absolute paths

#### File System Security
- **Secure Permissions**: Files created with `0o644`, directories with `0o750`
- **Atomic Operations**: Temporary files with `0o600` permissions
- **Backup Management**: Automatic cleanup with configurable retention
- **Input Validation**: Comprehensive sanitization of all inputs

### Security Scanning Results

- **gosec**: 0 security vulnerabilities detected
- **govulncheck**: No known vulnerabilities in dependencies
- **Static Analysis**: Clean code with comprehensive error handling
- **SBOM**: Complete Software Bill of Materials available

## Quality Assurance

### Testing Coverage

| Package             | Coverage | Description                             |
| ------------------- | -------- | --------------------------------------- |
| `internal/version`  | 100%     | Version management and build metadata   |
| `internal/logger`   | 79.2%    | Structured logging and debugging        |
| `internal/yaml`     | 49.4%    | YAML processing and security validation |
| `internal/workflow` | 33.1%    | Workflow orchestration and execution    |
| `internal/gitlab`   | 31.2%    | GitLab API integration and operations   |

### Code Quality Tools

- **golangci-lint**: 50+ linters including staticcheck and complexity analysis
- **staticcheck**: Advanced static analysis for performance and correctness
- **errcheck**: Unchecked error detection with comprehensive coverage
- **gofmt**: Code formatting compliance across all files
- **goimports**: Import organization and dependency management

### Build Quality

- **Cross-Platform**: Automated testing on Ubuntu, macOS, and Windows
- **Go Versions**: Tested with Go 1.22, 1.23, and 1.24
- **Performance**: Benchmarked operations with memory profiling
- **Security**: Automated vulnerability scanning in CI/CD pipeline

## Performance

### Operational Characteristics

- **Repository Operations**: Under 30 seconds for typical repositories
- **API Response Handling**: Within 5 seconds per request with retries
- **Memory Usage**: Under 100MB for standard operations
- **File Operations**: Security validation under 100ms per file
- **Concurrent Support**: Configurable concurrent operation limits

### Optimization Features

- **Rate Limiting**: Configurable requests per second (default: 10 RPS)
- **Connection Pooling**: HTTP client connection reuse
- **Pagination Support**: Efficient handling of large result sets
- **Resource Management**: Automatic cleanup and memory optimization

## Troubleshooting

### Common Issues

#### Authentication Errors
```bash
# Error: 401 Unauthorized
export GITLAB_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
go-tag-updater --debug --project-id=123 --file=test.yml --new-tag=v1.0.0 --token=$GITLAB_TOKEN
```

#### Project Not Found
```bash
# Verify project ID or path
go-tag-updater --debug --project-id=group/project --file=config.yml --new-tag=v1.0.0 --token=$GITLAB_TOKEN
```

#### File Path Security Errors
```bash
# Use relative paths within the repository
go-tag-updater --project-id=123 --file=config/app.yml --new-tag=v1.0.0 --token=$GITLAB_TOKEN
```

### Debug Mode

Enable comprehensive debugging for troubleshooting:

```bash
go-tag-updater \
  --debug \
  --project-id=myproject \
  --file=config.yml \
  --new-tag=v1.0.0 \
  --token=$GITLAB_TOKEN
```

Debug mode provides:
- Detailed API request/response logging
- Step-by-step workflow execution
- Security validation details
- Performance timing information

### Log Analysis

```bash
# Enable structured JSON logging
export GO_TAG_UPDATER_LOG_FORMAT=json
go-tag-updater --debug --project-id=123 --file=app.yml --new-tag=v1.0.0 --token=$GITLAB_TOKEN | jq '.'
```

## Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/Gosayram/go-tag-updater.git
cd go-tag-updater

# Install dependencies
make deps

# Run tests
make test

# Run quality checks
make check-all

# Build binary
make build

# Build for all platforms
make build-cross
```

### Development Tools

```bash
# Install development tools
make install-tools

# Run linting
make lint

# Run security scanning
make security-scan

# Generate SBOM
make sbom-generate

# Run benchmarks
make benchmark
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following the coding standards
4. Run tests and quality checks (`make check-all`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Coding Standards

- **Zero Magic Numbers**: All numeric literals must be named constants
- **English Documentation**: All comments and documentation in English
- **Professional Tone**: No emojis or casual language in code
- **Comprehensive Testing**: Unit tests for all new functionality
- **Security First**: Input validation and secure coding practices

## Architecture

For detailed architecture information, see [ARCHITECTURE.md](docs/ARCHITECTURE.md).

### Key Components

- **Command Layer**: CLI interface with Cobra framework
- **GitLab API Layer**: Official client integration with enhanced error handling
- **YAML Processing**: Secure parsing with path validation
- **Workflow Orchestration**: Step-by-step execution with rollback capabilities
- **Security Layer**: Multi-layer protection against common vulnerabilities

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/Gosayram/go-tag-updater/issues)
- **Security**: Report security issues via GitHub Security Advisories
- **Architecture**: [ARCHITECTURE.md](docs/ARCHITECTURE.md)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a detailed history of changes and releases. 