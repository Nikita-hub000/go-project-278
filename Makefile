.PHONY: start test lint build generate

SQLC_VERSION := v1.31.1

start:
	go run ./cmd/api

test:
	go test -race ./...

lint:
	golangci-lint run ./...

build:
	go build -o bin/api ./cmd/api

generate:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) generate
