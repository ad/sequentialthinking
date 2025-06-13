FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy go.mod for dependency caching
COPY go.mod ./

# Copy source code
COPY main.go ./

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /app/sequentialthinking-server main.go

# Final stage
FROM scratch
# Use scratch as the base image for a minimal final image
# Copy the CA certificates from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /root/

# Copy compiled binary
COPY --from=builder /app/sequentialthinking-server .

# Set entrypoint
ENTRYPOINT ["./sequentialthinking-server"]
