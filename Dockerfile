# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webailyzer-api ./cmd/webailyzer-api

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S webailyzer && \
    adduser -u 1001 -S webailyzer -G webailyzer

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/webailyzer-api .
COPY --from=builder /app/config.yaml .

# Create directories for migrations and logs
RUN mkdir -p migrations logs && \
    chown -R webailyzer:webailyzer /app

# Switch to non-root user
USER webailyzer

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the application
CMD ["./webailyzer-api"]