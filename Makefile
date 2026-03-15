.PHONY: build run test lint tidy session

build:
	go build -o scraper ./cmd/harvester/main.go

run:
	go run ./cmd/harvester/main.go

session:
	go run ./cmd/tg-session/main.go

test:
	go test ./...

lint:
	golangci-lint run

tidy:
	go mod tidy

qa: tidy lint test