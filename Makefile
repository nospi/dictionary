.PHONY: run build test clean docker-build docker-up docker-down install-deps

# Run the server locally
run:
	go run cmd/server/main.go

# Build the binary
build:
	go build -o bin/server cmd/server/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Install Go dependencies
install-deps:
	go mod download
	go mod tidy

# Build Docker image
docker-build:
	docker compose build

# Start services with Docker Compose
docker-up:
	docker compose up

# Stop Docker services
docker-down:
	docker compose down

# Development mode (with auto-reload if air is installed)
dev:
	@which air > /dev/null || (echo "air not installed. Run: go install github.com/air-verse/air@latest" && exit 1)
	air
