.PHONY: build run test help

# Build production image
build:
	docker build -t danielapatin/sequentialthinking .

# Run production container
run: build
	docker run --rm -p 8080:8080 danielapatin/sequentialthinking

# Run tests in container
test:
	docker run --rm -v $(PWD):/app -w /app golang:1.24-alpine go test -v

# Build binary locally
build-local:
	go build -o sequentialthinking-server main.go

# Run locally
run-local:
	./sequentialthinking-server

# Run in stdio mode
run-stdio:
	./sequentialthinking-server --stdio

# Show help
help:
	@echo "Available commands:"
	@echo "  build       - Build production Docker image"
	@echo "  run         - Run production container"
	@echo "  test        - Run tests in container"
	@echo "  build-local - Build binary locally"
	@echo "  run-local   - Run locally"
	@echo "  run-stdio   - Run in stdio mode"
	@echo "  help        - Show this help"
