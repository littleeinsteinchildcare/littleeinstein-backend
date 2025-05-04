# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go binary for Linux
#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/api/main.go
RUN go build -o api_server ./cmd/api/main.go

# Final stage: Alpine-based image
FROM alpine

WORKDIR /app

# Copy the binary from the builder
COPY --from=builder /app/api_server /api_server



# Optional: Add CA certs if your app needs HTTPS requests
#RUN apk add --no-cache ca-certificates

# Make sure it's executable
#RUN chmod +x /server

# Run the app
ENTRYPOINT ["/api_server"]
