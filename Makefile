.PHONY: run build migrate gen gen-wire gen-mock gen-swagger test test-unit test-cover test-integration docker-up docker-down docker-logs tidy fmt lint help

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

migrate:
	go run ./cmd/migrate

gen: gen-wire gen-mock gen-swagger

gen-wire:
	bash scripts/gen_wire.sh

gen-mock:
	mockery

gen-swagger:
	bash -c 'set -o pipefail; swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal 2>&1 | grep -Ev "failed to get package name|reflect: call of reflect.Value"'

test-unit:
	go test $(shell go list ./... | grep -v /test/integration) -race -count=1

test:
	docker compose -f docker-compose.test.yml up -d --wait
	go test ./... -race -count=1; \
	docker compose -f docker-compose.test.yml down

test-cover:
	docker compose -f docker-compose.test.yml up -d --wait
	go test ./... -race -count=1 -coverprofile=coverage.out; \
	docker compose -f docker-compose.test.yml down
	go tool cover -html=coverage.out

test-integration:
	docker compose -f docker-compose.test.yml up -d --wait
	go test ./test/integration/... -v -race -count=1; \
	docker compose -f docker-compose.test.yml down

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f api

tidy:
	go mod tidy

fmt:
	gofmt -w .

lint:
	golangci-lint run

help:
	@echo "Available commands:"
	@echo "  make run          - Run the API server"
	@echo "  make build        - Build binary to bin/api"
	@echo "  make migrate      - Run AutoMigrate"
	@echo "  make gen          - Run all generators (wire, mock, swagger)"
	@echo "  make gen-wire     - Generate Wire DI code"
	@echo "  make gen-mock     - Generate mocks with mockery"
	@echo "  make gen-swagger  - Generate Swagger docs"
	@echo "  make test-unit        - Run unit tests only (no Docker) with race detection"
	@echo "  make test             - Run all tests (unit + integration) with race detection"
	@echo "  make test-cover       - Run all tests with race detection + coverage report"
	@echo "  make test-integration - Run integration tests only (verbose)"
	@echo "  make docker-up    - Start Docker Compose"
	@echo "  make docker-down  - Stop Docker Compose"
	@echo "  make docker-logs  - Stream API logs"
	@echo "  make tidy         - Run go mod tidy"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Run linter"
