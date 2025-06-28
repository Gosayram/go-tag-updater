# go-tag-updater

[![Go Version](https://img.shields.io/github/go-mod/go-version/Gosayram/go-tag-updater)](https://golang.org/)
[![License](https://img.shields.io/github/license/Gosayram/go-tag-updater)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Gosayram/go-tag-updater)](https://goreportcard.com/report/github.com/Gosayram/go-tag-updater)
[![Security](https://github.com/Gosayram/go-tag-updater/actions/workflows/security.yml/badge.svg)](https://github.com/Gosayram/go-tag-updater/actions/workflows/security.yml)

A professional Go CLI tool for safely updating YAML files in GitLab repositories through automated merge request workflows. Built with the official GitLab Go API client library for reliable integration and enterprise-grade security.

## Features

- **🔄 Automated YAML Tag Updates**: Update tag values in YAML files with atomic operations
- **🔀 GitLab Integration**: Native GitLab API integration using official client library
- **🛡️ Security First**: Path traversal protection, input validation, and secure file operations
- **🌿 Branch Management**: Automatic branch creation with conflict detection
- **📋 Merge Request Automation**: Complete MR lifecycle management with customizable options
- **🔍 Project Resolution**: Support for both numeric project IDs and human-readable paths
- **🧪 Dry Run Mode**: Preview changes without making actual modifications
- **📊 Comprehensive Logging**: Structured logging with debug mode for troubleshooting
- **⚡ Performance Optimized**: Benchmarked operations with sub-second response times
- **🔒 Enterprise Security**: Security scanning, vulnerability checks, and SBOM generation

## Quick Start

### Installation

#### Download Binary

Download the latest release from the [releases page](https://github.com/Gosayram/go-tag-updater/releases).

#### Build from Source

```bash
git clone https://github.com/Gosayram/go-tag-updater.git
cd go-tag-updater
make build
```

#### Install via Go

```bash
go install github.com/Gosayram/go-tag-updater/cmd/go-tag-updater@latest
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

## Configuration

### Required Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `--project-id` | GitLab project ID or path | `123456` or `group/project` |
| `--file` | Path to YAML file in repository | `k8s/deployment.yaml` |
| `--new-tag` | New tag value to set | `v1.2.3` |
| `--token` | GitLab Personal Access Token | `glpat-xxxxxxxxxxxxxxxxxxxx` |

### Optional Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `--branch-name` | auto-generated | Custom branch name |
| `--target-branch` | `main` | Target branch for merge request |
| `--wait-previous-mr` | `false` | Wait for conflicting merge requests |
| `--debug` | `false` | Enable verbose debugging |
| `--dry-run` | `false` | Preview changes only |
| `--auto-merge` | `false` | Auto-merge when pipeline passes |

### Environment Variables

All CLI flags can be set via environment variables with the `GO_TAG_UPDATER_` prefix:

```bash
export GO_TAG_UPDATER_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
export GO_TAG_UPDATER_PROJECT_ID="mygroup/myproject"
export GO_TAG_UPDATER_DEBUG="true"
```

### Configuration File

Create a `go-tag-updater.yaml` file in your project root:

```yaml
gitlab:
  base_url: "https://gitlab.example.com"
  token: "${GITLAB_TOKEN}"
  timeout: 30s
  retry_count: 3

defaults:
  target_branch: "main"
  branch_prefix: "update-tag"
  auto_merge: false
  wait_previous_mr: false

performance:
  max_concurrent_requests: 5
  request_timeout: 30s
  buffer_size: 1024

logging:
  level: "info"
  format: "text"
  enable_file: false
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

### Batch Processing with Shell Script

```bash
#!/bin/bash
PROJECTS=("group/project1" "group/project2" "group/project3")
NEW_TAG="v2.1.0"

for project in "${PROJECTS[@]}"; do
  echo "Updating $project..."
  go-tag-updater \
    --project-id="$project" \
    --file="k8s/deployment.yaml" \
    --new-tag="$NEW_TAG" \
    --token="$GITLAB_TOKEN" \
    --wait-previous-mr=true
done
```

## GitLab Integration

### Required GitLab Permissions

The GitLab Personal Access Token must have the following scopes:

- `api` - Full API access
- `read_repository` - Read repository files
- `write_repository` - Create branches and update files

### Supported GitLab Versions

- GitLab.com (SaaS)
- GitLab CE/EE 13.0+
- GitLab API v4

### Project Identification

The tool supports flexible project identification:

```bash
# Numeric project ID
--project-id=123456

# Full project path
--project-id=group/subgroup/project

# URL-encoded paths are automatically handled
--project-id="group%2Fsubgroup%2Fproject"
```

## Security

### Security Features

- **Path Traversal Protection**: Validates all file paths to prevent `../` attacks
- **Input Sanitization**: Comprehensive validation of all user inputs
- **Secure Temporary Files**: Restricted permissions (0600) for temporary files
- **Token Security**: Secure handling of GitLab tokens with environment variable support
- **Audit Logging**: Comprehensive logging of all operations for security monitoring

### Security Scanning

The project includes comprehensive security scanning:

```bash
# Run security scan
make security-scan

# Check for vulnerabilities
make vuln-check

# Generate SBOM
make sbom-generate
```

### Security Reports

- **gosec**: Static security analysis
- **govulncheck**: Vulnerability scanning
- **SBOM**: Software Bill of Materials generation
- **SARIF**: Security reports in SARIF format

## Development

### Prerequisites

- Go 1.24.4 or later
- GitLab Personal Access Token
- Access to GitLab repository

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-cross

# Build with debug symbols
make build-debug
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run benchmarks
make benchmark

# Run integration tests
make test-integration
```

### Code Quality

```bash
# Run all quality checks
make check-all

# Individual checks
make lint
make staticcheck
make security-scan
make vuln-check
```

### Project Structure

```
go-tag-updater/
├── cmd/go-tag-updater/     # CLI application entry point
├── internal/               # Private application code
│   ├── config/            # Configuration management
│   ├── gitlab/            # GitLab API integration
│   ├── logger/            # Structured logging
│   ├── version/           # Version management
│   ├── workflow/          # Workflow orchestration
│   └── yaml/              # YAML processing
├── pkg/errors/            # Public error types
├── docs/                  # Documentation
├── scripts/               # Build and utility scripts
└── Makefile              # Build automation
```

## Performance

### Benchmarks

The tool is optimized for performance with comprehensive benchmarking:

- **File Operations**: Sub-100ms for typical YAML files
- **API Calls**: Under 5 seconds per GitLab API request
- **Memory Usage**: Under 100MB for standard operations
- **Concurrent Operations**: Support for batch processing

### Performance Testing

```bash
# Run performance benchmarks
make benchmark

# Generate benchmark report
make benchmark-report
```

## Troubleshooting

### Common Issues

#### Authentication Errors

```bash
# Verify token has correct permissions
curl -H "Authorization: Bearer $GITLAB_TOKEN" \
  "https://gitlab.com/api/v4/user"
```

#### Project Not Found

```bash
# Test project access
go-tag-updater \
  --project-id=your/project \
  --file=test.yaml \
  --new-tag=test \
  --dry-run \
  --debug \
  --token=$GITLAB_TOKEN
```

#### File Not Found

Ensure the file exists in the target branch:

```bash
# Check file exists in target branch
curl -H "Authorization: Bearer $GITLAB_TOKEN" \
  "https://gitlab.com/api/v4/projects/PROJECT_ID/repository/files/path%2Fto%2Ffile.yaml?ref=main"
```

### Debug Mode

Enable debug mode for detailed logging:

```bash
go-tag-updater \
  --debug \
  --project-id=your/project \
  --file=config.yaml \
  --new-tag=v1.0.0 \
  --token=$GITLAB_TOKEN
```

### Log Analysis

Debug logs include:

- GitLab API request/response details
- File operation tracking
- Branch creation and management
- Merge request lifecycle
- Error context and stack traces

## CI/CD Integration

### GitLab CI Example

```yaml
stages:
  - update-tags

update-deployment-tag:
  stage: update-tags
  image: golang:1.24.4
  before_script:
    - go install github.com/Gosayram/go-tag-updater/cmd/go-tag-updater@latest
  script:
    - go-tag-updater
        --project-id=$CI_PROJECT_ID
        --file=k8s/deployment.yaml
        --new-tag=$NEW_TAG
        --token=$GITLAB_TOKEN
        --auto-merge=true
  only:
    - tags
```

### GitHub Actions Example

```yaml
name: Update GitLab Tags
on:
  release:
    types: [published]

jobs:
  update-tags:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: 1.24.4
      - run: go install github.com/Gosayram/go-tag-updater/cmd/go-tag-updater@latest
      - run: |
          go-tag-updater \
            --project-id=${{ secrets.GITLAB_PROJECT_ID }} \
            --file=deployment.yaml \
            --new-tag=${{ github.event.release.tag_name }} \
            --token=${{ secrets.GITLAB_TOKEN }}
```

## API Reference

### Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |

### Error Types

| Code | Type | Description |
|------|------|-------------|
| 1001 | ValidationError | Invalid input parameters |
| 1002 | FileError | File operation failed |
| 1003 | YAMLError | YAML parsing/validation failed |
| 1004 | APIError | GitLab API error |
| 1005 | SecurityError | Security validation failed |

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone repository
git clone https://github.com/Gosayram/go-tag-updater.git
cd go-tag-updater

# Install dependencies
make deps

# Install development tools
make install-tools

# Run tests
make test

# Run quality checks
make check-all
```

### Code Standards

- All code comments and documentation in English
- Zero tolerance for magic numbers - use named constants
- Comprehensive test coverage with benchmarks
- Security-first approach with input validation
- Professional documentation without emojis

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/Gosayram/go-tag-updater/issues)
- **Security**: Report security issues privately to the maintainers

## Acknowledgments

- [GitLab Go API Client](https://gitlab.com/gitlab-org/api/client-go) - Official GitLab API client library
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management

---

**go-tag-updater** - Ideal for DevOps pipelines that rely on tag-based deployments and GitOps workflows. 