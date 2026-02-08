# Justfile

# Default recipe
default: build

# Build the application
build:
    @echo "Building binary..."
    go build -o bin/comicbookcollector ./cmd/scraper

# Run the scraper
run:
    go run ./cmd/scraper/main.go

# Run tests
test:
    go test -v ./...

# Clean up build artifacts and locks
clean:
    rm -rf bin/
    rm -f chrome-data/SingletonLock
    @echo "Cleaned up."

# Install dependencies
deps:
    go mod download
    go mod tidy
