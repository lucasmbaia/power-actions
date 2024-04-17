# Stage 1: Build the Go application
FROM golang:1.21-alpine AS builder

# Set necessary environment variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o powerpr .

# Stage 2: Build a small image
FROM alpine

# Install ca-certificates in case you need to interact with HTTPS endpoints
RUN apk add --no-cache ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /build/powerpr /app/powerpr

# Set the working directory in the container
WORKDIR /app

# Set the entrypoint to the application binary
ENTRYPOINT ["/app/powerpr"]
