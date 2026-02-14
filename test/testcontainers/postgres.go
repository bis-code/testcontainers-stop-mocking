package testcontainers

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	Container testcontainers.Container
	Pool      *pgxpool.Pool
	URI       string
}

func CreatePostgresContainer(ctx context.Context, schema string) (*PostgresContainer, error) {
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
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("get connection string: %w", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if schema != "" {
		if _, err := pool.Exec(ctx, schema); err != nil {
			pool.Close()
			_ = pgContainer.Terminate(ctx)
			return nil, fmt.Errorf("apply schema: %w", err)
		}
	}

	return &PostgresContainer{
		Container: pgContainer,
		Pool:      pool,
		URI:       connStr,
	}, nil
}

func (p *PostgresContainer) Terminate(ctx context.Context) error {
	p.Pool.Close()
	return p.Container.Terminate(ctx)
}
