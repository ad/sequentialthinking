FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy go.mod for dependency caching
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code and templates
COPY main.go ./
COPY templates/ ./templates/

# Build application
RUN CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o /app/sequentialthinking-server main.go

# Final stage
FROM scratch

# Copy the CA certificates from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /root/

# Copy compiled binary
COPY --from=builder /app/sequentialthinking-server .

# Expose port for HTTP mode
EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["./sequentialthinking-server"]
