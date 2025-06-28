# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2025-06-29

### Added
- **Initial Release**: Complete GitLab YAML tag updater tool
  - CLI tool for updating YAML files in GitLab repositories through automated merge requests
  - Support for GitLab API integration with authentication tokens
  - YAML parsing and tag updating functionality
  - Automated merge request creation and management
  - Comprehensive error handling and logging system
  - Docker containerization support with multi-stage builds

- **Professional CI/CD Pipeline**: Complete GitHub Actions workflow system
  - Automated testing across multiple Go versions (1.22, 1.23, 1.24)
  - Cross-platform testing matrix (Ubuntu, macOS, Windows)
  - Security scanning with CodeQL, Trivy, Nancy, Gosec, and Govulncheck
  - Automated releases with cross-platform binary building
  - Docker image building and publishing to GitHub Container Registry
  - Cosign keyless signing for all release artifacts
  - SBOM generation with Syft for supply chain security

- **Build System**: Professional Makefile with comprehensive targets
  - Cross-platform binary building (Linux, macOS, Windows for amd64/arm64)
  - Version management system using .release-version file
  - Integration testing with GitLab API functionality
  - Linting with golangci-lint and staticcheck
  - Coverage reporting and benchmark testing
  - Docker build and test targets

- **Documentation**: Complete project documentation
  - Comprehensive README with installation and usage instructions
  - Architecture documentation explaining system design
  - API integration guides for GitLab connectivity
  - Docker usage examples and best practices

### Fixed
- **Cross-Platform Path Security**: Enhanced file path validation for Windows compatibility
  - Fixed `validateAndCleanFilePath` function to properly detect absolute paths on Windows systems
  - Replaced Unix-specific path checks with cross-platform `filepath.IsAbs()` implementation
  - Added proper Windows drive letter detection and path normalization
  - Enhanced security by correctly blocking absolute paths outside safe directories on all platforms
  - Resolved test failures on Windows for path traversal security validation

- **Code Quality Improvements**: Eliminated magic numbers and strings according to project standards
  - Added comprehensive path security constants for cross-platform compatibility
  - Defined `PathTraversalPattern`, `WindowsDriveLetterSeparator`, and platform-specific path separators
  - Created safe directory prefix constants for Unix (`UnixTempDir`, `UnixVarTempDir`) and Windows (`WindowsCTempDir`, `WindowsCTmpDir`, etc.)
  - Fixed linter warnings about exported constant comments to follow Go documentation standards
  - All path validation logic now uses named constants instead of magic strings

- **Linter Configuration**: Resolved golangci-lint compatibility issues
  - Removed unsupported `allow-havelen-0` parameter from ginkgolinter configuration
  - Updated golangci-lint installation paths in Makefile to use correct v1.x format
  - Fixed GitHub Actions workflow to use compatible golangci-lint version
  - All linter checks now pass without warnings across all supported platforms

### Technical Details
- **Go Version**: 1.24 with modern Go modules
- **Architecture**: Clean architecture with separated concerns
  - `cmd/` - CLI entry points
  - `internal/` - Core business logic (config, git, gitlab, workflow, yaml)
  - `pkg/` - Reusable packages (errors)
- **Container**: Multi-stage Docker build with scratch final image
- **Security**: All binaries and containers signed with Cosign
- **Version Management**: Build-time version injection with commit, date, and builder information 