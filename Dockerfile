# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api_server ./cmd/api/main.go
# CGO_ENABLED=0 = Disables C Go features for a statically linked binary (more portable)
# GOOS=linux = Targets Linux operating system (regardless of build environment)

# Final stage - minimal runtime image
FROM alpine:latest
WORKDIR /app

# Copy the compiled binary from builder stage
COPY --from=builder /app/api_server /app/api_server

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Make the binary executable
RUN chmod +x /app/api_server

# Expose API port
EXPOSE 8080

# Set the API server as the main container process
CMD ["/app/api_server"]