# Simple Banking API

A production-ready RESTful banking API built with Go, Gin, GORM, and PostgreSQL. Supports account management, deposits, withdrawals, and transfers with full ACID compliance, Redis caching, JWT authentication, and comprehensive test coverage.

---

## Table of Contents

- [Running Locally](#running-locally)
- [Running with Docker Compose](#running-with-docker-compose)
- [Running Tests](#running-tests)
- [API Reference](#api-reference)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Database Design](#database-design)
- [Environment Variables](#environment-variables)
- [Make Commands](#make-commands)
- [Design Decisions](#design-decisions)

---

## Running Locally

### Prerequisites

| Tool | Version | Install |
|---|---|---|
| Go | 1.25+ | https://go.dev/dl |
| Docker | 24+ | https://docs.docker.com/get-docker |
| Docker Compose | v2 (bundled with Docker Desktop) | https://docs.docker.com/compose/install |
| `wire` | latest | `go install github.com/google/wire/cmd/wire@latest` |
| `swag` | latest | `go install github.com/swaggo/swag/cmd/swag@latest` |
| `mockery` | v2 | `go install github.com/vektra/mockery/v2@latest` |
| `golangci-lint` | latest | https://golangci-lint.run/usage/install |

> `wire`, `swag`, `mockery`, and `golangci-lint` are only needed if you plan to modify and regenerate code. You can skip them just to run and explore the API.

### Steps

```bash
# 1. Clone the repository
git clone <repo-url>
cd simple-banking-api

# 2. Copy environment config
cp .env.example .env

# 3. Start infrastructure (PostgreSQL + Redis)
make docker-up

# 4. Run database migrations
make migrate

# 5. (Optional) Seed demo users — password: Password@123
psql -h localhost -U postgres -d banking -f db/seeds/001_users.sql

# 6. Start the API server
make run
```

The API is available at **http://localhost:8080**
Swagger UI at **http://localhost:8080/swagger/index.html**

---

## Running with Docker Compose

To run the entire stack (infrastructure + API) with a single command, uncomment the `migrate` and `api` services in `docker-compose.yml`, then:

```bash
# Build and start everything
docker compose up --build -d

# Stream logs
make docker-logs

# Stop
make docker-down
```

The `migrate` service runs first and exits on completion. The `api` service waits for `migrate` to complete before accepting traffic.

**Services:**

| Service | Image | Port |
|---|---|---|
| `postgres` | postgres:15-alpine | 5432 |
| `redis` | redis:7-alpine | 6379 |
| `migrate` | Built from Dockerfile | — |
| `api` | Built from Dockerfile | 8080 |

The **Dockerfile** uses a multi-stage build:
- **Stage 1 (builder):** `golang:1.25-alpine` — downloads dependencies and compiles both `api` and `migrate` binaries
- **Stage 2 (runtime):** `alpine:3.19` — copies binaries only, resulting in a ~20 MB image

---

## Running Tests

### Unit Tests

Unit tests live alongside service code in `internal/service/**/*_test.go`. All external dependencies are replaced by Mockery-generated mocks — no Docker required.

```bash
make test-unit
```

### Integration Tests

Integration tests in `test/integration/` run against real PostgreSQL and Redis instances (started via `docker-compose.test.yml` on ports 5433 / 6380). Each test case runs in a clean state — all tables are truncated between tests.

```bash
# All tests: unit + integration
make test

# Integration tests only (verbose output)
make test-integration

# All tests + HTML coverage report
make test-cover
```

### Test Coverage Areas

| Suite | Scenarios |
|---|---|
| `AuthSuite` | Register, login, token refresh, duplicate email, wrong password |
| `AccountSuite` | Create, list, get balance, unauthorized access, forbidden (other user) |
| `TransactionSuite` | Deposit, withdraw, transfer — success, insufficient balance, forbidden, same-account, balance updates |
| `AccountServiceSuite` | Unit tests with mocks: create, balance (cache hit/miss), list, forbidden |
| `TransactionServiceSuite` | Unit tests: deposit/withdraw/transfer error paths |
| `AuthServiceSuite` | Unit tests: register, login, refresh token flows |

---

## API Reference

**Base URL:** `http://localhost:8080/api/v1`

**Swagger UI:** `http://localhost:8080/swagger/index.html`

Protected endpoints require:
```
Authorization: Bearer <access_token>
```

---

### Auth

#### `POST /auth/register`
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com", "password": "Password@123"}'
```
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Alice",
  "email": "alice@example.com",
  "createdAt": "2025-01-01T00:00:00Z"
}
```

#### `POST /auth/login`
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "password": "Password@123"}'
```
```json
{
  "accessToken": "<jwt>",
  "refreshToken": "<opaque-token>",
  "tokenType": "Bearer",
  "expiresIn": 900
}
```

#### `POST /auth/refresh`
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refreshToken": "<opaque-token>"}'
```
Returns a new token pair identical to the login response.

---

### Accounts _(requires auth)_

#### `POST /accounts` — Create account
```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"currency": "THB"}'
```
```json
{
  "userId": "550e8400-...",
  "accountNumber": "4831927560",
  "balance": "0",
  "currency": "THB"
}
```

#### `GET /accounts` — List my accounts
```bash
curl http://localhost:8080/api/v1/accounts \
  -H "Authorization: Bearer <token>"
```

#### `GET /accounts/:accountNumber` — Get balance
```bash
curl http://localhost:8080/api/v1/accounts/4831927560 \
  -H "Authorization: Bearer <token>"
```
```json
{
  "accountNumber": "4831927560",
  "balance": "1000"
}
```

#### `GET /accounts/:accountNumber/transactions` — List transactions
```bash
curl "http://localhost:8080/api/v1/accounts/4831927560/transactions?page=1&limit=20" \
  -H "Authorization: Bearer <token>"
```
```json
{
  "transactions": [
    {
      "id": "...",
      "fromAccountNumber": "4831927560",
      "amount": "200",
      "type": "withdraw",
      "status": "success",
      "createdAt": "2025-01-01T00:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20
}
```

---

### Transactions _(requires auth)_

#### `POST /transactions/deposit`
```bash
curl -X POST http://localhost:8080/api/v1/transactions/deposit \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"accountNumber": "4831927560", "amount": "500.00"}'
```

#### `POST /transactions/withdraw`
```bash
curl -X POST http://localhost:8080/api/v1/transactions/withdraw \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"accountNumber": "4831927560", "amount": "200.00"}'
```

#### `POST /transactions/transfer`
```bash
curl -X POST http://localhost:8080/api/v1/transactions/transfer \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "fromAccountNumber": "4831927560",
    "toAccountNumber":   "9102384756",
    "amount":            "300.00"
  }'
```

Transaction response shape:
```json
{
  "id": "...",
  "fromAccountNumber": "4831927560",
  "toAccountNumber": "9102384756",
  "amount": "300",
  "type": "transfer",
  "status": "success",
  "createdAt": "2025-01-01T00:00:00Z"
}
```

### Error Responses

All errors follow a consistent format with an appropriate HTTP status code:

```json
{
  "message": "insufficient balance",
  "detail": "balance 50 is less than requested amount 200"
}
```

| HTTP Status | Scenario |
|---|---|
| 400 | Invalid request body / same-account transfer |
| 401 | Missing or invalid JWT |
| 403 | Account belongs to another user |
| 404 | Account or user not found |
| 409 | Duplicate account (same user + currency) |
| 422 | Insufficient balance |
| 500 | Internal server error |

---

## Tech Stack

| Layer | Technology | Purpose |
|---|---|---|
| Language | Go 1.25 | |
| Web Framework | [Gin](https://github.com/gin-gonic/gin) v1.10 | HTTP routing, middleware, binding |
| ORM | [GORM](https://gorm.io) v1.25 + `gorm.io/driver/postgres` | Database access, AutoMigrate |
| Database | PostgreSQL 15 | Persistent storage |
| Cache | Redis 7 + [go-redis](https://github.com/redis/go-redis) v9 | Balance caching |
| Authentication | [golang-jwt/jwt](https://github.com/golang-jwt/jwt) v5 (HS256) + bcrypt | Auth & password hashing |
| Dependency Injection | [Google Wire](https://github.com/google/wire) v0.6 | Compile-time DI |
| Decimal | [shopspring/decimal](https://github.com/shopspring/decimal) v1.4 | Precision money arithmetic |
| Validation | [go-playground/validator](https://github.com/go-playground/validator) v10 | Request validation |
| Docs | [Swaggo](https://github.com/swaggo/swag) | Swagger UI |
| Testing | [Testify](https://github.com/stretchr/testify) + [Mockery](https://github.com/vektra/mockery) | Unit & integration tests |
| Logging | slog (Go stdlib) | Structured logging |

---

## Architecture

The project follows **Clean Architecture** with four layers that depend strictly inward:

```
HTTP Request
    │
    ▼
┌─────────────────────────────────┐
│  Handler Layer (Gin)            │  ← Bind & validate request, map to/from DTO
│  internal/handler/              │
└────────────────┬────────────────┘
                 │
                 ▼
┌─────────────────────────────────┐
│  Service Layer                  │  ← Business rules, authorization, transactions
│  internal/service/              │
└────────────────┬────────────────┘
                 │
                 ▼
┌─────────────────────────────────┐
│  Repository Layer               │  ← Data access abstractions (interfaces)
│  internal/repository/           │
│   ├── db/      (PostgreSQL)     │
│   └── cache/   (Redis)          │
└─────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────┐
│  Domain Layer                   │  ← Entities, constants, no dependencies
│  internal/domain/               │
└─────────────────────────────────┘
```

**Dependency Injection** is handled by Google Wire — all wiring is generated at compile time. No service locators or runtime reflection.

---

## Project Structure

```
simple-banking-api/
├── cmd/
│   ├── api/                        # Application entry point
│   │   ├── main.go                 # Server bootstrap & graceful shutdown
│   │   └── di/                     # Wire injector (wire.go + wire_gen.go)
│   └── migrate/                    # Standalone migration runner
│       └── di/
├── internal/
│   ├── adapter/
│   │   ├── postgres/               # PostgreSQL connection setup
│   │   └── redis/                  # Redis client
│   ├── config/                     # Configuration structs loaded via envconfig
│   ├── domain/
│   │   ├── entity/                 # Domain models: User, Account, Transaction, RefreshToken
│   │   └── constrant/              # Transaction types, statuses, cache key prefixes
│   ├── errs/                       # Domain-specific sentinel errors
│   ├── handler/
│   │   ├── handler/                # Auth, Account, Transaction HTTP handlers
│   │   ├── dto/
│   │   │   ├── request/            # Request DTOs with validation tags
│   │   │   └── response/           # Response DTOs & entity mappers
│   │   ├── middleware/             # JWTAuth, request Logger
│   │   └── router.go               # Route definitions
│   ├── repository/
│   │   ├── db/
│   │   │   ├── model/              # GORM models + ToEntity/FromEntity converters
│   │   │   ├── migrate/            # AutoMigrate setup
│   │   │   ├── mock/               # Mockery-generated mocks
│   │   │   ├── account.go          # AccountRepository implementation
│   │   │   ├── transaction.go      # TransactionRepository implementation
│   │   │   ├── user.go             # UserRepository + TokenRepository
│   │   │   ├── tx.go               # TxManager (database transaction context)
│   │   │   └── repo.go             # Repository & TxManager interfaces
│   │   └── cache/
│   │       ├── mock/               # Mockery-generated cache mock
│   │       └── cache.go            # Redis cache implementation
│   ├── service/
│   │   ├── auth/                   # Register, Login, RefreshToken
│   │   ├── account/                # CreateAccount, GetBalance, ListAccounts
│   │   └── transaction/            # Deposit, Withdraw, Transfer
│   ├── server/                     # HTTP server lifecycle
│   └── utils/                      # Account number generator, auth context helpers
├── pkg/
│   ├── errs/                       # Generic error types (AppError, Sentinel)
│   ├── jwt/                        # JWT manager (sign & parse HS256 tokens)
│   └── logger/                     # slog wrapper
├── test/
│   └── integration/                # End-to-end integration test suite
│       ├── main_test.go            # TestMain: start/stop Docker test containers
│       ├── suite_test.go           # BaseSuite: per-test DB truncate & helpers
│       ├── auth_test.go
│       ├── account_test.go
│       ├── transaction_test.go
│       ├── helpers_test.go
│       ├── testutil/               # DB & Redis setup for tests
│       └── di/                     # Test-specific Wire injector
├── db/
│   └── seeds/
│       └── 001_users.sql           # 5 seeded users (password: Password@123)
├── docs/                           # Swagger generated files (do not edit)
├── scripts/
│   └── gen_wire.sh                 # Wire provider set generator
├── Dockerfile                      # Multi-stage build: builder → alpine runtime
├── docker-compose.yml              # Infrastructure: PostgreSQL 15 + Redis 7
├── docker-compose.test.yml         # Test infrastructure (separate ports)
├── Makefile
├── .env.example                    # Environment variable template
└── .env.test                       # Test environment config
```

---

## Database Design

Migrations run automatically via GORM `AutoMigrate` (`cmd/migrate`).

### Entity-Relationship Overview

```
users ──< accounts ──< transactions (to_account_id)
                  └──< transactions (from_account_id)
users ──< refresh_tokens
```

### Tables

#### `users`
| Column | Type | Constraints |
|---|---|---|
| `id` | UUID | PK |
| `name` | varchar | not null |
| `email` | varchar | not null, unique |
| `password_hash` | varchar | not null (bcrypt) |
| `created_at` | timestamp | auto |
| `updated_at` | timestamp | auto |
| `deleted_at` | timestamp | null (soft delete) |

#### `accounts`
| Column | Type | Constraints |
|---|---|---|
| `id` | UUID | PK |
| `user_id` | UUID | FK → users, indexed |
| `account_number` | varchar | not null, unique |
| `balance` | decimal(20,2) | not null, default 0 |
| `currency` | varchar | not null, default 'THB' |
| `created_at` | timestamp | auto |
| `updated_at` | timestamp | auto |
| `deleted_at` | timestamp | null (soft delete) |

#### `transactions`
| Column | Type | Constraints |
|---|---|---|
| `id` | UUID | PK |
| `from_account_id` | UUID | nullable, FK → accounts, indexed |
| `to_account_id` | UUID | nullable, FK → accounts, indexed |
| `amount` | decimal(20,2) | not null |
| `type` | varchar(20) | `deposit` \| `withdraw` \| `transfer` |
| `status` | varchar(20) | `pending` \| `success` \| `failed` |
| `note` | text | nullable |
| `created_at` | timestamp | auto, indexed |

| Operation | `from_account_id` | `to_account_id` |
|---|---|---|
| deposit | null | destination account |
| withdraw | source account | null |
| transfer | source account | destination account |

#### `refresh_tokens`
| Column | Type | Constraints |
|---|---|---|
| `id` | UUID | PK |
| `user_id` | UUID | FK → users, indexed |
| `token` | varchar | not null, unique |
| `expires_at` | timestamp | not null |
| `created_at` | timestamp | auto |

---

## Environment Variables

Copy `.env.example` to `.env` and adjust as needed.

```env
# Application
APP_ENV=development       # development | production
APP_PORT=8080

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=banking
DB_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=           # leave empty to disable auth
REDIS_DB=0

# JWT
JWT_SECRET=change-me-to-a-random-secret   # use a long random string in production
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h      # 7 days

# Logging
LOG_LEVEL=info            # debug | info | warn | error
```

---

## Make Commands

```
make run              Run the API server locally
make build            Compile binary to bin/api
make migrate          Run AutoMigrate against the configured database

make gen              Run all code generators (wire + mock + swagger)
make gen-wire         Regenerate Wire DI code
make gen-mock         Regenerate Mockery mocks
make gen-swagger      Regenerate Swagger docs

make test-unit        Unit tests only (no Docker, race detector on)
make test             All tests: unit + integration (auto-starts test containers)
make test-integration Integration tests only (verbose)
make test-cover       All tests + HTML coverage report
make test-up          Start test containers only
make test-down        Stop test containers

make docker-up        Build and start Docker Compose stack
make docker-down      Stop Docker Compose stack
make docker-logs      Stream API container logs

make tidy             go mod tidy
make fmt              gofmt -w .
make lint             golangci-lint run
```

---

## Design Decisions

### Decimal arithmetic for money
All amounts use `github.com/shopspring/decimal` and are stored as `DECIMAL(20,2)` in PostgreSQL. This avoids floating-point precision errors inherent in `float64` (e.g. `0.1 + 0.2 ≠ 0.3` in IEEE 754).

### Pessimistic locking on transfers
`FindByAccountNumberForUpdate` issues `SELECT … FOR UPDATE` (via GORM's `Clauses(clause.Locking{Strength: "UPDATE"})`) inside a database transaction. This acquires a row-level lock on both account rows before modifying balances, preventing lost updates under concurrent transfers to the same accounts.

### Redis balance cache
`GET /accounts/:accountNumber` is a hot read path. The balance is cached in Redis with a 60-second TTL and immediately invalidated on any mutating operation (deposit, withdraw, transfer). Cache misses fall back to PostgreSQL transparently.

### Authorization at the service layer
Ownership checks (`account.UserID == callerID`) live in the service, not the handler. This keeps business rules centralized and ensures they cannot be bypassed regardless of how the service is invoked.

### Transaction context propagation
`TxManager.Transaction(ctx, fn)` stores the GORM `*DB` (with the active transaction) in the context. Repository methods that must run inside a transaction call `requireTx(ctx)` to enforce this at runtime — they fail immediately if called outside a transaction, making misuse detectable rather than silently corrupting data.

### Account number as the public identifier
All public-facing responses use the human-readable `accountNumber` (e.g. `"4831927560"`) instead of internal UUIDs. This keeps the API surface stable and avoids leaking internal implementation details.

### Structured error types
`AppError` carries both an HTTP status code and a machine-readable `message`. Services return sentinel errors (`errs.ErrAccountNotFound`, `errs.ErrInsufficientBalance`, etc.); the handler maps them to HTTP responses. `.New("…")` adds a human-readable `detail` without changing the sentinel identity, so `errors.Is` matching still works across the call stack.
