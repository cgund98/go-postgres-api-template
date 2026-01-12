FROM golang:1.25

# Install build dependencies and development tools
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    git \
    make \
    bash \
    curl \
    ca-certificates \
    unzip \
    less && \
    rm -rf /var/lib/apt/lists/*

# Install golangci-lint for linting
RUN curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.8.0

# Install AWS CLI v2
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
    unzip -q awscliv2.zip && \
    ./aws/install && \
    rm -rf awscliv2.zip aws

# Create a non-root user
RUN groupadd -g 1000 workspace && \
    useradd -u 1000 -g workspace -m -s /bin/bash workspace

# Set permissions for Go module cache and GOPATH directories
# This ensures workspace user can access golang-migrate and other Go tools
RUN mkdir -p /go/pkg/mod /go/bin && \
    chown -R workspace:workspace /go

# Ensure /go/bin is in PATH (Go image already includes it, but make it explicit)
ENV PATH="/go/bin:${PATH}"

# Disable AWS CLI pager for non-interactive use
ENV AWS_PAGER=""

# Set working directory and ensure workspace user owns it
WORKDIR /workspace
RUN chown -R workspace:workspace /workspace

# Switch to non-root user
USER workspace

# Install golang-migrate for database migrations
# Install as root, then ensure workspace user can access it
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Install air for live reload
RUN go install github.com/air-verse/air@v1.63.8

# Default command
CMD ["/bin/bash"]

