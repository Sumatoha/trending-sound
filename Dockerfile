# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies (gcc and musl-dev for CGO/sqlite3)
RUN apk add --no-cache \
    git \
    gcc \
    musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled for sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o bot ./cmd/bot

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata

# Set timezone
ENV TZ=UTC

# Create app directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bot /app/bot

# Copy migrations
COPY --from=builder /app/migrations /app/migrations

# Create data directory
RUN mkdir -p /app/data

# Set environment variables
ENV DATA_DIR=/app/data

# Volume for persistent data
VOLUME /app/data

# Expose health check (optional)
HEALTHCHECK --interval=60s --timeout=5s --start-period=10s --retries=3 \
  CMD pgrep bot || exit 1

# Run the bot
CMD ["/app/bot"]
