# Testcontainers: Stop Mocking Your Database

## Video Hook: Problem → Solution
"I ran into this problem: my tests were passing, but production kept breaking on duplicate emails. Here's how I thought through it."

The reveal: mock tests can't enforce database constraints. Testcontainers gives you real databases in tests.

## Narrative Flow
1. **Problem (0-5s):** "My tests said everything was fine. Production said otherwise."
2. **Show the gap (5-20s):** Mock test passes on duplicate email — the mock doesn't know about UNIQUE constraints
3. **The thinking (20-40s):** "What if the test used a real database? But I don't want shared test environments..."
4. **Solution (40-55s):** Testcontainers — real PostgreSQL in Docker, 10 lines of setup, catches the bug
5. **Payoff (55-60s):** "Real databases. Real confidence. Zero infrastructure."

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

## Inspired By
Real-world experience with Testcontainers for distributed SQL and integration testing at scale.
