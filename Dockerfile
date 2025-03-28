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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webanalyzer ./cmd/analyzer

# Use a minimal image for the final stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/webanalyzer .

# Copy templates and CSS files
COPY --from=builder /app/ui/templates ./ui/templates
COPY --from=builder /app/ui/css ./ui/css

# Expose ports for web interface and metrics
EXPOSE 8080 9090

# Run the binary
CMD ["./webanalyzer"] 