//go:build integration

package user_test

import (
	"context"
	"strings"
	"testing"

	"github.com/bis-code/testcontainers-stop-mocking/internal/user"
	tc "github.com/bis-code/testcontainers-stop-mocking/test/testcontainers"
	"github.com/stretchr/testify/suite"
)

type IntegrationSuite struct {
	suite.Suite
	pg   *tc.PostgresContainer
	repo *user.PostgresRepository
	ctx  context.Context
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}

func (s *IntegrationSuite) SetupSuite() {
	s.ctx = context.Background()

	pg, err := tc.CreatePostgresContainer(s.ctx, user.Schema)
	s.Require().NoError(err, "failed to start postgres container")

	s.pg = pg
	s.repo = user.NewPostgresRepository(pg.Pool)
}

func (s *IntegrationSuite) TearDownSuite() {
	if s.pg != nil {
		_ = s.pg.Terminate(s.ctx)
	}
}

func (s *IntegrationSuite) TearDownTest() {
	_, err := s.pg.Pool.Exec(s.ctx, "TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	s.Require().NoError(err, "failed to truncate users table")
}

// TestCreateUser uses a real PostgreSQL instance.
// It catches bugs that mocks silently ignore.
func (s *IntegrationSuite) TestCreateUser() {
	// First user — should succeed
	u1, err := s.repo.Create(s.ctx, "alice@example.com", "Alice")
	s.Require().NoError(err)
	s.Equal("alice@example.com", u1.Email)
	s.NotZero(u1.ID)

	// Duplicate email — the REAL database rejects this.
	// This is the bug that mock tests miss entirely.
	_, err = s.repo.Create(s.ctx, "alice@example.com", "Alice Duplicate")
	s.Require().Error(err, "expected error on duplicate email — this is the bug mocks hide!")
	s.True(
		strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique"),
		"expected unique constraint error, got: %s", err.Error(),
	)
}

func (s *IntegrationSuite) TestGetUser() {
	created, err := s.repo.Create(s.ctx, "bob@example.com", "Bob")
	s.Require().NoError(err)

	got, err := s.repo.Get(s.ctx, created.ID)
	s.Require().NoError(err)
	s.Equal("bob@example.com", got.Email)
	s.Equal("Bob", got.Name)
}

func (s *IntegrationSuite) TestListUsers() {
	// Empty table
	users, err := s.repo.List(s.ctx)
	s.Require().NoError(err)
	s.Empty(users)

	// Add two users
	_, err = s.repo.Create(s.ctx, "alice@example.com", "Alice")
	s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, "bob@example.com", "Bob")
	s.Require().NoError(err)

	users, err = s.repo.List(s.ctx)
	s.Require().NoError(err)
	s.Len(users, 2)
}
