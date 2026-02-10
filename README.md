# Testcontainers: Stop Mocking Your Database

## Video Concept
Short-form (60s) showing why mocking databases in tests gives false confidence, and how Testcontainers spins up real databases in Docker for integration tests.

## Key Points
1. **Hook (0-3s):** "Your database tests are lying to you"
2. **Problem (3-15s):** Show a mock-based test that passes but misses a real SQL constraint violation
3. **Solution (15-45s):** Same test with Testcontainers — real MySQL/PostgreSQL in Docker, catches the bug
4. **Payoff (45-60s):** "Real databases. Real confidence. Zero infrastructure."

## Tech Stack
- Go (using `testcontainers-go`)
- PostgreSQL container
- Simple user service with a unique constraint

## Code Structure
```
cmd/main.go              # Simple user service
internal/user/
  repository.go          # Database operations
  repository_test.go     # Testcontainers-based integration tests
  repository_mock_test.go # Mock-based test (the "wrong" way)
docker-compose.yml       # For local dev
go.mod
```

## The Demo
1. Mock test: `TestCreateUser_Mock` — passes even with duplicate emails
2. Real test: `TestCreateUser_Integration` — catches the unique constraint violation
3. Show the Testcontainers setup is ~10 lines of code

## Derived From
- scopito-core: Testcontainers setup for TiDB
- scopito-image-db-service: Integration tests with containerized databases
- Pattern: Real database testing without shared test environments
