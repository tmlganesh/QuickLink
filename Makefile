# Run the URL shortener server
run:
	go run main.go

# Run tests
test:
	go test -v

# Run demo client (requires server to be running)
demo:
	go run -tags demo demo.go

# Install dependencies
deps:
	go mod tidy

# Clean build artifacts
clean:
	go clean

# Build binary
build:
	go build -o url-shortener.exe main.go

# Run with race detection
run-race:
	go run -race main.go

# Test with race detection
test-race:
	go test -v -race

# Format code
fmt:
	go fmt ./...

# Check for common issues
vet:
	go vet ./...

# Full check (format, vet, test)
check: fmt vet test

.PHONY: run test demo deps clean build run-race test-race fmt vet check
