# Build stage
FROM golang:1.22-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy all source code
COPY . .

# Explicitly downgrade dependencies and tidy
RUN go mod edit -go=1.22 && \
    go get golang.org/x/net@v0.17.0 && \
    go get github.com/cespare/xxhash/v2@v2.2.0 && \
    go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o analyzer cmd/analyzer/main.go

# Final stage
FROM alpine:latest

# Install CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary and necessary files from builder
COPY --from=builder /app/analyzer .
COPY --from=builder /app/config ./config
COPY --from=builder /app/ui ./ui

# Expose port
EXPOSE 8080

# Run the application
CMD ["./analyzer"]