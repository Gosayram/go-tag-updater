# Build stage v1.24.4-alpine3.22
FROM golang:1.24.4-alpine3.22@sha256:68932fa6d4d4059845c8f40ad7e654e626f3ebd3706eef7846f319293ab5cb7a AS builder

# Build arguments
ARG COMMIT=""

# Install git and ca-certificates (needed for private repos and HTTPS)
RUN apk --no-cache add git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with proper version info
RUN set -e; \
    # Read version from .release-version and add -container suffix
    VERSION=$(cat .release-version 2>/dev/null || echo "1.0.0"); \
    VERSION_CONTAINER="${VERSION}-container"; \
    \
    # Use build arg COMMIT if provided, otherwise get from git
    if [ -n "$COMMIT" ]; then \
        COMMIT_HASH="$COMMIT"; \
    else \
        COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
    fi; \
    \
    # Get current date
    DATE=$(date -u +%Y-%m-%d_%H:%M:%S); \
    \
    # Build with ldflags
    CGO_ENABLED=0 GOOS=linux go build \
        -a -installsuffix cgo \
        -ldflags "-extldflags \"-static\" -s -w \
            -X github.com/Gosayram/go-tag-updater/internal/version.Version=${VERSION_CONTAINER} \
            -X github.com/Gosayram/go-tag-updater/internal/version.Commit=${COMMIT_HASH} \
            -X github.com/Gosayram/go-tag-updater/internal/version.Date=${DATE} \
            -X github.com/Gosayram/go-tag-updater/internal/version.BuiltBy=docker" \
        -o go-tag-updater \
        ./cmd/go-tag-updater

# Final stage
FROM scratch

# Copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/go-tag-updater /go-tag-updater

# Set the binary as entrypoint
ENTRYPOINT ["/go-tag-updater"]

# Default command
CMD ["--version"]

# Labels
LABEL org.opencontainers.image.title="go-tag-updater"
LABEL org.opencontainers.image.description="GitLab YAML tag updater tool"
LABEL org.opencontainers.image.source="https://github.com/Gosayram/go-tag-updater"
LABEL org.opencontainers.image.licenses="MIT"