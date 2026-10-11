.PHONY: start test lint build

start:
	go run ./cmd/api

test:
	go test -race ./...

lint:
	golangci-lint run ./...

build:
	go build -o bin/api ./cmd/api
