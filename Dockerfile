# Build stage for Go application
FROM golang:alpine AS go-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the server binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o brique-server ./cmd/brique-server

# Final runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# Copy binary from builder
COPY --from=go-builder /app/brique-server .

# Create data directory
RUN mkdir -p /var/lib/brique

# Expose port for HTTP server
EXPOSE 8080

# Set environment variables
ENV BRIQUE_DATA_DIR=/var/lib/brique
ENV BRIQUE_PORT=8080

# Run as non-root user
RUN addgroup -g 1000 brique && \
    adduser -D -u 1000 -G brique brique && \
    chown -R brique:brique /app /var/lib/brique

USER brique

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

VOLUME ["/var/lib/brique"]

CMD ["./brique-server"]
