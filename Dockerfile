FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod ./
COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webanalyzer ./cmd/api

# Use a minimal image for the final stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/webanalyzer .

# Copy static files and templates
COPY --from=builder /app/ui ./ui

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./webanalyzer"] 