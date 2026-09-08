# Stage 1: Build binary
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

# Copy module definition
COPY go.mod ./
RUN go mod download

# Copy source code
COPY main.go ./
COPY engine/ ./engine/

# Build static binary using Go's fast native cross-compilation with BuildKit cache
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-arm64} \
    go build -ldflags="-s -w" -o /bin/app main.go

# Stage 2: Final minimal image
FROM alpine:3.21

WORKDIR /app

# Install ca-certificates and tzdata for TLS and local time handling
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root system user
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/home/appuser" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid 10001 \
    appuser

# Copy static binary from builder
COPY --from=builder /bin/app /app/app

# Copy data directory
COPY data/ /app/data/

# Set ownership of the app directory to the non-root user
RUN chown -R appuser:appuser /app

# Switch to the non-root user
USER appuser

EXPOSE 8080

# P1: Container Health Check (Docker & Watchtower monitor container readiness)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

CMD ["/app/app"]
