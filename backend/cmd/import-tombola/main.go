package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rotaract-civ/backend/internal/auth"
	"github.com/rotaract-civ/backend/internal/database"
)

type tombolaMember struct {
	ID       string
	Name     string
	Email    string
	Phone    *string
	ClubName *string
}

func main() {
	_ = godotenv.Load()
	var sourceURL string
	var dryRun bool
	flag.StringVar(&sourceURL, "source-url", os.Getenv("TOMBOLA_DATABASE_URL"), "Tombola PostgreSQL connection string")
	flag.BoolVar(&dryRun, "dry-run", false, "show the import without writing users or memberships")
	flag.Parse()
	if strings.TrimSpace(sourceURL) == "" {
		log.Fatal("TOMBOLA_DATABASE_URL or -source-url is required")
	}
	targetURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(targetURL) == "" {
		log.Fatal("DATABASE_URL is required for the club-management database")
	}

	ctx := context.Background()
	source, err := pgxpool.New(ctx, sourceURL)
	if err != nil {
		log.Fatal(err)
	}
	defer source.Close()
	target, err := database.Connect(ctx, targetURL)
	if err != nil {
		log.Fatal(err)
	}
	defer target.Close()

	rows, err := source.Query(ctx, `SELECT id::text, name, email, phone, club_name FROM members ORDER BY created_at`)
	if err != nil {
		log.Fatal(fmt.Errorf("read Tombola members: %w", err))
	}
	defer rows.Close()

	imported, existing, unmatched := 0, 0, 0
	for rows.Next() {
		var member tombolaMember
		if err := rows.Scan(&member.ID, &member.Name, &member.Email, &member.Phone, &member.ClubName); err != nil {
			log.Fatal(err)
		}
		email := strings.ToLower(strings.TrimSpace(member.Email))
		if email == "" {
			log.Printf("skip member %s: missing email", member.ID)
			continue
		}
		firstName, lastName := splitName(member.Name)
		var userID string
		err := target.Pool.QueryRow(ctx, `SELECT id::text FROM users WHERE email = $1`, email).Scan(&userID)
		if err == nil {
			existing++
		} else if err == pgx.ErrNoRows {
			if dryRun {
				imported++
				userID = "dry-run"
			} else {
				temporaryHash, hashErr := temporaryPasswordHash()
				if hashErr != nil {
					log.Fatal(hashErr)
				}
				err = target.Pool.QueryRow(ctx, `
					INSERT INTO users (email, password_hash, first_name, last_name, phone, is_active)
					VALUES ($1, $2, $3, $4, $5, TRUE) RETURNING id::text
				`, email, temporaryHash, firstName, lastName, member.Phone).Scan(&userID)
				if err != nil {
					log.Printf("skip %s: create user: %v", email, err)
					continue
				}
				imported++
			}
		} else {
			log.Fatal(err)
		}

		if member.ClubName == nil || strings.TrimSpace(*member.ClubName) == "" {
			continue
		}
		var clubID string
		err = target.Pool.QueryRow(ctx, `SELECT id::text FROM clubs WHERE LOWER(name) = LOWER($1) AND is_active = TRUE LIMIT 1`, strings.TrimSpace(*member.ClubName)).Scan(&clubID)
		if err == pgx.ErrNoRows {
			unmatched++
			log.Printf("unmatched club for %s: %s", email, *member.ClubName)
			continue
		}
		if err != nil {
			log.Fatal(err)
		}
		if dryRun || userID == "dry-run" {
			continue
		}
		_, err = target.Pool.Exec(ctx, `INSERT INTO club_memberships (club_id, user_id, member_role) VALUES ($1, $2, 'member') ON CONFLICT (club_id, user_id) DO NOTHING`, clubID, userID)
		if err != nil {
			log.Printf("membership failed for %s: %v", email, err)
		}
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	log.Printf("Tombola import complete: imported=%d existing=%d unmatched_clubs=%d dry_run=%t", imported, existing, unmatched, dryRun)
}

func splitName(fullName string) (string, string) {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return "Membre", "Rotaract"
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func temporaryPasswordHash() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return auth.HashPassword(fmt.Sprintf("migration-%x", raw))
}
