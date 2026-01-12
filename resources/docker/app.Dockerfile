# Build stage
FROM golang:1.24.3-alpine3.20 AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build API binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s' \
    -o /build/bin/api \
    ./cmd/api

# Build worker binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s' \
    -o /build/bin/worker \
    ./cmd/worker

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binaries
COPY --from=builder /build/bin/api /app/api
COPY --from=builder /build/bin/worker /app/worker

# Set working directory
WORKDIR /app

# Default to running API
# To run worker instead: docker run <image> /app/worker
CMD ["/app/api"]

