# Testcontainers: Stop Mocking Your Database

Your mock tests pass. Production breaks. Sound familiar?

Mock tests can't enforce database constraints like `UNIQUE`. Testcontainers gives you real databases in tests — no shared infrastructure, no flaky setups.

## The Problem

```go
// Mock happily accepts duplicate emails
u1, _ := mockRepo.Create(ctx, "alice@example.com", "Alice")
u2, _ := mockRepo.Create(ctx, "alice@example.com", "Alice Again")
// No error. Test passes. Production explodes.
```

## The Fix

```go
// Real PostgreSQL catches it immediately
u1, _ := realRepo.Create(ctx, "alice@example.com", "Alice")
u2, err := realRepo.Create(ctx, "alice@example.com", "Alice Again")
// err: duplicate key value violates unique constraint "users_email_key"
```

## Code Structure

```
cmd/main.go                          # HTTP server with user CRUD
internal/user/
  user.go                            # User model
  repository.go                      # PostgreSQL repository + schema
  repository_test.go                 # Integration tests (Testcontainers + testify/suite)
  repository_mock_test.go            # Mock tests (the false positive)
test/testcontainers/
  postgres.go                        # Reusable Postgres container helper
Makefile                             # Dev and test commands
docker-compose.yml                   # Local dev database
```

## Quick Start

**Prerequisites:** Go 1.21+ and Docker running.

```bash
# Run unit tests only (no Docker required)
make unit-test

# Run integration tests (spins up Postgres via Testcontainers)
make integration-test

# Run all tests
make test

# Start the dev server (starts infra + runs app)
make dev
```

### What You'll See

| Test | Result | Why |
|------|--------|-----|
| `TestCreateUser_Mock` | PASS | Mock doesn't enforce UNIQUE constraint |
| `TestIntegrationSuite/TestCreateUser` | PASS | Real PostgreSQL rejects duplicate email |

The mock test passes when it shouldn't — that's the whole point.

## Run the Server

```bash
# Start PostgreSQL
make infra-up-detached

# Run the server
go run ./cmd

# Create a user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "name": "Alice"}'

# Try duplicate — gets rejected
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "name": "Alice Again"}'

# List users
curl http://localhost:8080/users

# Stop infrastructure
make infra-down
```

> **Note:** The compose file maps to port **5433** to avoid conflicts with a local PostgreSQL on 5432.

## Testcontainers Setup

The container helper (`test/testcontainers/postgres.go`) is reusable across packages:

```go
pg, err := tc.CreatePostgresContainer(ctx, user.Schema)
defer pg.Terminate(ctx)

repo := user.NewPostgresRepository(pg.Pool)
```

Integration tests use `testify/suite` for lifecycle management — one container per suite, table truncation between tests.

Real databases. Real confidence. Zero infrastructure.
