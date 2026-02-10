package user_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bis-code/testcontainers-stop-mocking/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupPostgres spins up a real PostgreSQL container for testing.
// This is ~10 lines of meaningful setup — that's it.
func setupPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

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
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	// Create the schema — same as production would have.
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id         SERIAL PRIMARY KEY,
			email      VARCHAR(255) UNIQUE NOT NULL,
			name       VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return pool
}

// TestCreateUser_Integration uses a real PostgreSQL instance.
// It catches bugs that mocks silently ignore.
func TestCreateUser_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	pool := setupPostgres(t)
	repo := user.NewPostgresRepository(pool)
	ctx := context.Background()

	// First user — should succeed
	u1, err := repo.Create(ctx, "alice@example.com", "Alice")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u1.Email != "alice@example.com" {
		t.Fatalf("expected email alice@example.com, got %s", u1.Email)
	}
	if u1.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	// Duplicate email — the REAL database rejects this.
	// This is the bug that mock tests miss entirely.
	_, err = repo.Create(ctx, "alice@example.com", "Alice Duplicate")
	if err == nil {
		t.Fatal("expected error on duplicate email, got nil — this is the bug mocks hide!")
	}
	if !strings.Contains(err.Error(), "duplicate key") && !strings.Contains(err.Error(), "unique") {
		t.Logf("got error (constraint enforced): %v", err)
	}

	t.Log("Integration test caught the duplicate email — real database, real confidence!")
}

func TestGetUser_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	pool := setupPostgres(t)
	repo := user.NewPostgresRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, "bob@example.com", "Bob")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Email != "bob@example.com" {
		t.Fatalf("expected email bob@example.com, got %s", got.Email)
	}
	if got.Name != "Bob" {
		t.Fatalf("expected name Bob, got %s", got.Name)
	}
}

func TestListUsers_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	pool := setupPostgres(t)
	repo := user.NewPostgresRepository(pool)
	ctx := context.Background()

	// Empty table
	users, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("expected 0 users, got %d", len(users))
	}

	// Add two users
	_, err = repo.Create(ctx, "alice@example.com", "Alice")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_, err = repo.Create(ctx, "bob@example.com", "Bob")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	users, err = repo.List(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
}
