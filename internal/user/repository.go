package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Schema is the SQL DDL for the users table.
// Exported so integration tests can apply it to a container database.
const Schema = `
CREATE TABLE IF NOT EXISTS users (
	id         SERIAL PRIMARY KEY,
	email      VARCHAR(255) UNIQUE NOT NULL,
	name       VARCHAR(255) NOT NULL,
	created_at TIMESTAMP DEFAULT NOW()
);
`

// Repository defines the operations for managing users.
type Repository interface {
	Create(ctx context.Context, email, name string) (*User, error)
	Get(ctx context.Context, id int) (*User, error)
	List(ctx context.Context) ([]User, error)
}

// PostgresRepository implements Repository using a real PostgreSQL database.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, email, name string) (*User, error) {
	// BUG: No application-level duplicate check.
	// The database UNIQUE constraint is our only safety net.
	// Mock tests won't catch this — real database tests will.
	var u User
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, name) VALUES ($1, $2)
		 RETURNING id, email, name, created_at`,
		email, name,
	).Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

func (r *PostgresRepository) Get(ctx context.Context, id int) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, name, created_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &u, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]User, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, email, name, created_at FROM users`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}
