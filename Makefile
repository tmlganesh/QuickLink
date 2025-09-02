run:
	go run main.go

test:
	go test -v

demo:
	go run -tags demo demo.go

deps:
	go mod tidy

clean:
	go clean

build:
	go build -o url-shortener.exe main.go

run-race:
	go run -race main.go

test-race:
	go test -v -race

fmt:
	go fmt ./...

vet:
	go vet ./...

check: fmt vet test

.PHONY: run test demo deps clean build run-race test-race fmt vet check
