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
  repository.go                      # PostgreSQL repository
  repository_test.go                 # Integration tests (Testcontainers)
  repository_mock_test.go            # Mock tests (the false positive)
docker-compose.yml                   # Local dev database
```

## Run the Tests

**Prerequisites:** Go 1.21+ and Docker running.

```bash
# Run all tests (mock + integration)
go test -v ./...

# Run only mock tests (fast, but misleading)
go test -v -short ./...

# Run only integration tests
go test -v -run Integration ./...
```

### What You'll See

| Test | Result | Why |
|------|--------|-----|
| `TestCreateUser_Mock` | PASS | Mock doesn't enforce UNIQUE constraint |
| `TestCreateUser_Integration` | PASS | Real PostgreSQL rejects duplicate email |

The mock test passes when it shouldn't — that's the whole point.

## Run the Server

```bash
# Start PostgreSQL
docker compose up -d

# Run the server
go run cmd/main.go

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
```

## Testcontainers Setup

The entire Testcontainers setup is ~10 lines:

```go
pgContainer, err := postgres.Run(ctx,
    "postgres:16-alpine",
    postgres.WithDatabase("testdb"),
    postgres.WithUsername("testuser"),
    postgres.WithPassword("testpass"),
    testcontainers.WithWaitStrategy(
        wait.ForLog("database system is ready to accept connections").
            WithOccurrence(2).
            WithStartupTimeout(30*time.Second),
    ),
)
```

Real databases. Real confidence. Zero infrastructure.
