# Start from the official Golang base image (build stage)
FROM golang:1.24.5-alpine AS builder

# Set environment variables
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Create working directory
WORKDIR /app

# Install git (needed for Go module fetching)
RUN apk add --no-cache git

# Copy go.mod and go.sum first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary
RUN go build -o server .

# Start a minimal base image (runtime stage)
FROM alpine:latest

# Set working directory
WORKDIR /root/

# Copy the built binary from builder
COPY --from=builder /app/server .

# Expose the port your Gin app listens on (default 8080)
EXPOSE 8080

# Command to run the executable
CMD ["./server"]
