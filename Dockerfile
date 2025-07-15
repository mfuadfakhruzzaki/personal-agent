FROM golang:1.21-alpine AS builder

# Install ca-certificates for HTTPS calls
RUN apk add --no-cache ca-certificates git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates and optionally tesseract for OCR
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy config files
COPY --from=builder /app/config ./config

# Create temp directory for file uploads
RUN mkdir -p /tmp/todo-agent

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

# Run the binary
CMD ["./main"]
