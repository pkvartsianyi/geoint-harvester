.PHONY: build run test lint tidy

build:
	go build -o scraper main.go

run:
	go run main.go

test:
	go test ./...

lint:
	golangci-lint run

tidy:
	go mod tidy

qa: tidy lint test