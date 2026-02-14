package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/bis-code/testcontainers-stop-mocking/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/userdb?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Ensure schema exists
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id         SERIAL PRIMARY KEY,
			email      VARCHAR(255) UNIQUE NOT NULL,
			name       VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatalf("failed to create schema: %v", err)
	}

	repo := user.NewPostgresRepository(pool)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", createUserHandler(repo))
	mux.HandleFunc("GET /users/{id}", getUserHandler(repo))
	mux.HandleFunc("GET /users", listUsersHandler(repo))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

type createUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func createUserHandler(repo *user.PostgresRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Email == "" || req.Name == "" {
			http.Error(w, "email and name are required", http.StatusBadRequest)
			return
		}

		u, err := repo.Create(r.Context(), req.Email, req.Name)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create user: %v", err), http.StatusConflict)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(u)
	}
}

func getUserHandler(repo *user.PostgresRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid user ID", http.StatusBadRequest)
			return
		}

		u, err := repo.Get(r.Context(), id)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(u)
	}
}

func listUsersHandler(repo *user.PostgresRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := repo.List(r.Context())
		if err != nil {
			http.Error(w, "failed to list users", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}
