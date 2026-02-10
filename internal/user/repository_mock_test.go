package user_test

import (
	"context"
	"sync"
	"testing"

	"github.com/bis-code/testcontainers-stop-mocking/internal/user"
)

// MockRepository simulates the database — but can't enforce constraints.
// This is the trap: your tests pass, but production breaks.
type MockRepository struct {
	mu    sync.Mutex
	users []user.User
	nextID int
}

func NewMockRepository() *MockRepository {
	return &MockRepository{nextID: 1}
}

func (m *MockRepository) Create(_ context.Context, email, name string) (*user.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// No UNIQUE constraint check — the mock doesn't know about schema rules.
	u := user.User{
		ID:    m.nextID,
		Email: email,
		Name:  name,
	}
	m.nextID++
	m.users = append(m.users, u)
	return &u, nil
}

func (m *MockRepository) Get(_ context.Context, id int) (*user.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, u := range m.users {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, nil
}

func (m *MockRepository) List(_ context.Context) ([]user.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]user.User, len(m.users))
	copy(result, m.users)
	return result, nil
}

// TestCreateUser_Mock demonstrates the FALSE POSITIVE problem.
// This test passes — but the behavior it validates is WRONG.
// In production, inserting two users with the same email would violate
// the UNIQUE constraint. The mock has no idea.
func TestCreateUser_Mock(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	// First user — works fine
	u1, err := repo.Create(ctx, "alice@example.com", "Alice")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u1.Email != "alice@example.com" {
		t.Fatalf("expected email alice@example.com, got %s", u1.Email)
	}

	// Duplicate email — the mock happily accepts it.
	// In a real database, this would fail with a UNIQUE violation.
	u2, err := repo.Create(ctx, "alice@example.com", "Alice Duplicate")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u2.ID == u1.ID {
		t.Fatal("expected different IDs")
	}

	// The mock says we have 2 users. The real DB would have rejected the second insert.
	users, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	// This test PASSES. That's the problem.
	// It gave us confidence that doesn't reflect reality.
	t.Log("Mock test passed — but this hides a real bug!")
}

func TestGetUser_Mock(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	created, err := repo.Create(ctx, "bob@example.com", "Bob")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got == nil {
		t.Fatal("expected user, got nil")
	}
	if got.Email != "bob@example.com" {
		t.Fatalf("expected email bob@example.com, got %s", got.Email)
	}
}
