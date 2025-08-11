# Build stage for DynamoDB initialization tool
FROM golang:1.24.6-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files from parent directory
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download

# Copy only the local development tools
COPY backend/local/ ./local/

# Build the DynamoDB initialization tool
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dynamodb-init ./local/dynamodb-init.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/dynamodb-init .

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Run the DynamoDB initialization
CMD ["./dynamodb-init"]
