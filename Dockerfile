# syntax=docker/dockerfile:1.7

# Build stage
FROM golang:1.23.4-alpine3.20 AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /workspace

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Allow for multi-arch builds via BuildKit TARGET* args
ARG TARGETOS=linux
ARG TARGETARCH=amd64

# Build the application with optimization flags
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" \
    -o /workspace/appointment-service \
    ./cmd/server

# Busybox stage used for lightweight healthcheck tooling
FROM busybox:1.36.1 AS healthcheck-tools

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot

# Add labels for metadata
LABEL org.opencontainers.image.title="LSV HBYS Service" \
      org.opencontainers.image.description="Bridge service to HBYS (MSSQL) with appointment APIs" \
      org.opencontainers.image.vendor="LSV" \
      org.opencontainers.image.source="https://github.com/berkecengiz/appointment-service-boilerplate"

WORKDIR /app

# Copy binary from builder
COPY --from=builder /workspace/appointment-service /app/appointment-service

# Copy timezone and CA data for runtime dependencies
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy busybox applet to drive HTTP health checks
COPY --from=healthcheck-tools /bin/busybox /bin/busybox
COPY --from=healthcheck-tools /bin/busybox /bin/wget

# Expose port
EXPOSE 8080

# Use nonroot user for security
USER nonroot:nonroot

# Health check against service endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/bin/wget", "-qO-", "http://127.0.0.1:8080/health"]

ENTRYPOINT ["/app/appointment-service"]
