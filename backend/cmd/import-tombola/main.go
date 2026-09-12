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
	"github.com/rotaract-civ/backend/internal/clubname"
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

	clubLookup, err := loadClubLookup(ctx, target.Pool)
	if err != nil {
		log.Fatal(err)
	}

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
		clubID, ok := clubLookup.resolve(*member.ClubName)
		if !ok {
			unmatched++
			log.Printf("unmatched club for %s: %s", email, *member.ClubName)
			continue
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

type clubLookupMap struct {
	byExact      map[string]string
	byNormalized map[string]string
}

func loadClubLookup(ctx context.Context, pool *pgxpool.Pool) (*clubLookupMap, error) {
	rows, err := pool.Query(ctx, `SELECT id::text, name FROM clubs WHERE is_active = TRUE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lookup := &clubLookupMap{
		byExact:      make(map[string]string),
		byNormalized: make(map[string]string),
	}
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		lookup.byExact[strings.ToLower(strings.TrimSpace(name))] = id
		if key := clubname.NormalizeKey(name); key != "" {
			lookup.byNormalized[key] = id
		}
	}
	return lookup, rows.Err()
}

func (m *clubLookupMap) resolve(clubName string) (string, bool) {
	trimmed := strings.TrimSpace(clubName)
	if id, ok := m.byExact[strings.ToLower(trimmed)]; ok {
		return id, true
	}
	if id, ok := m.byNormalized[clubname.NormalizeKey(trimmed)]; ok {
		return id, true
	}
	return "", false
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
	// bcrypt accepts at most 72 bytes; "migration-" + 32-byte hex would exceed that.
	raw := make([]byte, 30)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return auth.HashPassword(fmt.Sprintf("migration-%x", raw))
}
