.PHONY: run build migrate gen gen-wire gen-mock gen-swagger test test-unit test-cover test-integration test-up test-down docker-up docker-down docker-logs tidy fmt lint help

BIN_DIR        := bin
API_BIN        := $(BIN_DIR)/api
CMD_API        := ./cmd/api
CMD_MIGRATE    := ./cmd/migrate

SWAGGER_ENTRY  := cmd/api/main.go
SWAGGER_OUT    := docs

TEST_COMPOSE   := docker-compose.test.yml
TEST_FLAGS     := -race -count=1
COVER_PROFILE  := coverage.out
COVER_PKGS     := ./internal/...,./pkg/...

run:
	go run $(CMD_API)

build:
	go build -o $(API_BIN) $(CMD_API)

migrate:
	go run $(CMD_MIGRATE)

gen: gen-wire gen-mock gen-swagger

gen-wire:
	bash scripts/gen_wire.sh

gen-mock:
	mockery

gen-swagger:
	bash -c 'set -o pipefail; swag init -g $(SWAGGER_ENTRY) -o $(SWAGGER_OUT) --parseDependency --parseInternal 2>&1 | grep -Ev "failed to get package name|reflect: call of reflect.Value"'

test-unit:
	go test $(shell go list ./... | grep -v /test/integration) $(TEST_FLAGS)

test: test-down
	docker compose -f $(TEST_COMPOSE) up -d --wait
	go test ./... $(TEST_FLAGS); \
	docker compose -f $(TEST_COMPOSE) down

test-cover:
	docker compose -f $(TEST_COMPOSE) up -d --wait
	go test ./... $(TEST_FLAGS) -coverprofile=$(COVER_PROFILE) -coverpkg=$(COVER_PKGS); \
	docker compose -f $(TEST_COMPOSE) down
	@echo "Coverage profile saved to $(COVER_PROFILE)"
	@echo "  GoLand : Run → Manage Coverage Reports → Add → select $(COVER_PROFILE)"
	@echo "  Browser: make cover-html"

test-integration:
	docker compose -f $(TEST_COMPOSE) up -d --wait
	go test ./test/integration/... -v $(TEST_FLAGS); \
	docker compose -f $(TEST_COMPOSE) down

test-up:
	docker compose -f $(TEST_COMPOSE) up -d --wait

test-down:
	docker compose -f $(TEST_COMPOSE) down

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
	@echo "  make run              - Run the API server"
	@echo "  make build            - Build binary to $(API_BIN)"
	@echo "  make migrate          - Run AutoMigrate"
	@echo "  make gen              - Run all generators (wire, mock, swagger)"
	@echo "  make gen-wire         - Generate Wire DI code"
	@echo "  make gen-mock         - Generate mocks with mockery"
	@echo "  make gen-swagger      - Generate Swagger docs"
	@echo "  make test-unit        - Run unit tests only (no Docker) with race detection"
	@echo "  make test             - Run all tests (unit + integration) with race detection"
	@echo "  make test-cover       - Run all tests + save coverage.out (GoLand compatible)"
	@echo "  make test-integration - Run integration tests only (verbose)"
	@echo "  make test-up          - Start test containers"
	@echo "  make test-down        - Stop test containers"
	@echo "  make docker-up        - Start Docker Compose"
	@echo "  make docker-down      - Stop Docker Compose"
	@echo "  make docker-logs      - Stream API logs"
	@echo "  make tidy             - Run go mod tidy"
	@echo "  make fmt              - Format code"
	@echo "  make lint             - Run linter"
