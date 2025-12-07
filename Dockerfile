# Build stage
FROM golang:1.24.0-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
# Using CGO_ENABLED=0 for a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o kpf-app ./cmd/bin/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/kpf-app .

# Copy necessary files (config, migrations, etc.)
COPY --from=builder /app/db ./db
COPY --from=builder /app/config.yaml ./config.yaml

# Create necessary directories with proper permissions
RUN mkdir -p /app/logs /app/storage /app/backups && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port (default Fiber port is 3000, adjust if needed)
EXPOSE 3333

# Run the application
CMD ["./kpf-app"]
