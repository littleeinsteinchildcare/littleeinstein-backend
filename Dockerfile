# Multi-stage build Dockerfile which creates a smaller, more secure final image

# ========================================
# Stage 1: Build Stage
# ========================================
FROM golang:1.24.2-alpine AS builder
# Using Alpine Linux with Go 1.24 preinstalled as the build environment
# The "AS builder" names this stage so we can reference it later

WORKDIR /app
# Sets the working directory inside the container to /app
# All subsequent commands will run from this directory

# Copy dependency files first to leverage Docker layer caching
# This step will only re-run if go.mod or go.sum change, not when source code changes
COPY go.mod go.sum ./
RUN go mod download
# Downloads all dependencies defined in go.mod
# This is separated from copying source code to improve build caching

# Now copy the rest of the source code
COPY . .
# Copies all files from your project into the container
# This includes all Go source files, configs, etc.

# Build the Go application with specific optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api_server ./cmd/api/main.go
# CGO_ENABLED=0 disables C Go features, creating a statically linked binary
#   - More portable across different Linux distributions
#   - No dynamic library dependencies
#   - Better security isolation
# GOOS=linux targets Linux regardless of build environment (works on Windows/Mac)
# GOARCH=amd64 builds for 64-bit x86 architecture
# -o api_server names the output binary
# ./cmd/api/main.go is the path to the application entry point

# ========================================
# Stage 2: Final Runtime Image
# ========================================
FROM alpine:latest
# Starting fresh with a minimal Alpine Linux image
# This creates a much smaller final image by discarding the build environment

WORKDIR /app
# Sets the working directory in the final container

# Copy just the compiled binary from the builder stage
# This is the key to multi-stage builds - we only take what we need
COPY --from=builder /app/api_server /app/api_server
# The --from=builder flag references the previous stage
# Only the compiled binary is copied, not source code or build tools

# Install minimal runtime dependencies
RUN apk add --no-cache ca-certificates tzdata
# ca-certificates: Required for HTTPS connections
# tzdata: Timezone data for proper time handling
# --no-cache: Doesn't store the index locally, keeping the image smaller

# Make the binary executable (ensure proper permissions)
RUN chmod +x /app/api_server

# Document which port the application uses
EXPOSE 8080
# This is documentation only - it doesn't actually publish the port
# You still need to use -p flag when running the container

# Define the command to run when the container starts
CMD ["/app/api_server"]
# Uses exec form (recommended) rather than shell form
# This runs the API server directly, making it PID 1 in the container
# The container will exit if this process exits