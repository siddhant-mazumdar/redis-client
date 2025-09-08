SHELL := /bin/bash

APP := redis-lite-server
PKG := ./...

.PHONY: all tidy fmt vet lint test build run docker

all: tidy fmt vet lint test build

fmt:
	@echo "> go fmt"
	@go fmt $(PKG)

vet:
	@echo "> go vet"
	@go vet $(PKG)

tidy:
	@echo "> go mod tidy"
	@go mod tidy

lint:
	@echo "> golangci-lint"
	@golangci-lint run ./...

test:
	@echo "> go test"
	@CGO_ENABLED=1 go test $(PKG)

build:
	@echo "> go build"
	@CGO_ENABLED=1 go build -o bin/$(APP) ./cmd

run:
	@echo "> go run ./cmd"
	@GO_ENV=development CGO_ENABLED=1 go run ./cmd

docker:
	docker build -t $(APP):latest .

