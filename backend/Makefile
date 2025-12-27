.PHONY: help build run test clean docker-build

# Default config path (relative to backend directory)
CONFIG_PATH ?= ../configs/config.yaml

APP_NAME=filelocker
BUILD_DIR=bin

help:
	@echo "Backend Makefile"
	@echo "  make run    - Run server locally using ../configs/config.yaml"
	@echo "  make build  - Build binary"
	@echo "  make test   - Run tests"

run:
	@echo "Starting Backend using config: $(CONFIG_PATH)"
	@CONFIG_PATH=$(CONFIG_PATH) go run cmd/server/main.go

build:
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) cmd/server/main.go

test:
	@go test ./... -v -race
