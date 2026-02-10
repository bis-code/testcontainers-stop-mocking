# Video Script: Testcontainers — Stop Mocking Your Database

## Format: Long-form (10-12 min)

---

### Act 1: The Problem (2-3 min)
"My tests were passing. All green. CI was clean. And then production broke — duplicate email addresses getting through to the database. The test literally covered this exact case. How?"

- Show `repository_mock_test.go` — the mock test for duplicate emails
- Run it: green. "See? Passes."
- Reveal: "The mock doesn't enforce database constraints. It's just an in-memory map."
- "Your mock tests are testing your mock. Not your database."

### Act 2: What Testcontainers Is (2 min)
"What if your tests used a real database? Not a shared staging database that's always dirty. A fresh, disposable database that only your test knows about."

- Explain Testcontainers: spins up a real Docker container per test
- Show the setup code: ~10 lines
- "PostgreSQL 16, fresh, isolated, torn down automatically when the test ends."

### Act 3: Live Coding Walkthrough (4-5 min)
Walk through the full implementation:
1. `user.go` — the model (simple)
2. `repository.go` — the PostgreSQL repository with CRUD + UNIQUE constraint
3. `repository_mock_test.go` — the mock approach (show why it's insufficient)
4. `repository_test.go` — the Testcontainers approach (show the real test)
5. Key moment: run both side-by-side, show the mock passes but the real test catches the bug

### Act 4: When to Use What (2 min)
"I'm not saying never use mocks. Mocks are great for unit tests — testing business logic, not database behavior."

- **Use mocks**: business logic, service layer, fast unit tests
- **Use Testcontainers**: anything touching the database — constraints, queries, migrations
- "Test pyramid: mocks at the bottom (fast, many), Testcontainers in the middle (slower, fewer but real)"

### Act 5: Demo (1-2 min)
- Run `go test -v ./...` — show both mock and integration tests
- Start the server, hit the API with curl
- Show duplicate rejection in the real app
- "The code is in the description. Clone it, run it, break it."

---

## Shorts Extraction Points

| Timestamp (approx) | Short Title | Duration |
|---|---|---|
| Act 1, 0:30-1:30 | "Your mock tests are lying to you" | 60s |
| Act 2, full | "Testcontainers in 60 seconds" | 60s |
| Act 3, side-by-side run | "Mock vs real database: watch this test" | 45s |
| Act 4, summary | "When to mock, when to use real databases" | 45s |
| Act 1, hook | "All tests green. Production broke. Here's why." | 30s |

## Tags
`#go #golang #testing #testcontainers #docker #tdd #softwaredevelopment`
