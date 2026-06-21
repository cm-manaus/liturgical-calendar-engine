# Stage 1: Build binary
FROM golang:1.22-alpine AS builder

WORKDIR /src

# Copy module definition
COPY go.mod ./
RUN go mod download

# Copy source code
COPY main.go ./
COPY engine/ ./engine/

# Build static binary
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/app main.go

# Stage 2: Final minimal image
FROM alpine:3.19

WORKDIR /app

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

ENV DATA_DIR="/app/data"

CMD ["/app/app"]
