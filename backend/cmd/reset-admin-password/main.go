package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rotaract-civ/backend/internal/auth"
)

func main() {
	_ = godotenv.Load()
	email := flag.String("email", "", "admin email to reset (required)")
	password := flag.String("password", "", "new password (default: BOOTSTRAP_ADMIN_PASSWORD)")
	grantAdmin := flag.Bool("grant-admin", true, "ensure is_admin and is_active are true")
	listOnly := flag.Bool("list", false, "list admin users and exit")
	flag.Parse()

	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if *listOnly {
		rows, err := pool.Query(ctx, `
			SELECT email, is_admin, is_active,
			       CASE WHEN password_hash LIKE 'oauth:%' THEN 'oauth' ELSE 'password' END
			FROM users
			WHERE is_admin = TRUE
			ORDER BY created_at
		`)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var email string
			var isAdmin, isActive bool
			var authType string
			if err := rows.Scan(&email, &isAdmin, &isActive, &authType); err != nil {
				log.Fatal(err)
			}
			log.Printf("admin: %s active=%v auth=%s", email, isActive, authType)
		}
		return
	}

	normalized := strings.ToLower(strings.TrimSpace(*email))
	if normalized == "" {
		log.Fatal("-email is required (or use -list)")
	}

	newPassword := strings.TrimSpace(*password)
	if newPassword == "" {
		newPassword = strings.TrimSpace(os.Getenv("BOOTSTRAP_ADMIN_PASSWORD"))
	}
	if newPassword == "" {
		log.Fatal("provide -password or set BOOTSTRAP_ADMIN_PASSWORD")
	}

	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		log.Fatal(err)
	}

	var id string
	err = pool.QueryRow(ctx, `SELECT id::text FROM users WHERE email = $1`, normalized).Scan(&id)
	if err != nil {
		log.Fatalf("user not found: %s", normalized)
	}

	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2::uuid`
	args := []any{hash, id}
	if *grantAdmin {
		query = `UPDATE users SET password_hash = $1, is_admin = TRUE, is_active = TRUE, updated_at = NOW() WHERE id = $2::uuid`
	}

	tag, err := pool.Exec(ctx, query, args...)
	if err != nil {
		log.Fatal(err)
	}
	if tag.RowsAffected() == 0 {
		log.Fatal("no rows updated")
	}

	log.Printf("Password reset for %s (admin=%v).", normalized, *grantAdmin)
}
