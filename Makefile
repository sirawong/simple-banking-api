.PHONY: run build migrate gen gen-wire gen-mock gen-swagger test test-cover docker-up docker-down docker-logs tidy fmt lint help

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

migrate:
	go run ./cmd/api -migrate

gen: gen-wire gen-mock gen-swagger

gen-wire:
	bash scripts/gen_wire.sh

gen-mock:
	mockery --all --dir=internal/repository --output=internal/mocks --outpkg=mocks

gen-swagger:
	swag init -g cmd/api/main.go -o docs

test:
	go test ./... -v -race

test-cover:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out

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
	@echo "  make test         - Run all tests"
	@echo "  make test-cover   - Run tests with coverage report"
	@echo "  make docker-up    - Start Docker Compose"
	@echo "  make docker-down  - Stop Docker Compose"
	@echo "  make docker-logs  - Stream API logs"
	@echo "  make tidy         - Run go mod tidy"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Run linter"
