package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	dryRun := flag.Bool("dry-run", false, "list deletions without applying")
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

	keep := map[string]struct{}{}
	for _, email := range splitEmails(os.Getenv("BOOTSTRAP_ADMIN_EMAIL")) {
		keep[email] = struct{}{}
	}
	for _, email := range tombolaEmails(ctx, pool) {
		keep[email] = struct{}{}
	}

	rows, err := pool.Query(ctx, `
		SELECT id::text, email, is_admin, first_name, last_name
		FROM users
		ORDER BY created_at
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var toDelete []string
	for rows.Next() {
		var id, email, firstName, lastName string
		var isAdmin bool
		if err := rows.Scan(&id, &email, &isAdmin, &firstName, &lastName); err != nil {
			log.Fatal(err)
		}
		normalized := strings.ToLower(strings.TrimSpace(email))
		if isAdmin {
			log.Printf("keep admin: %s <%s>", firstName+" "+lastName, email)
			continue
		}
		if _, ok := keep[normalized]; ok {
			log.Printf("keep imported: %s <%s>", firstName+" "+lastName, email)
			continue
		}
		if shouldDeleteMock(normalized) {
			log.Printf("delete mock: %s <%s>", firstName+" "+lastName, email)
			toDelete = append(toDelete, id)
			continue
		}
		log.Printf("keep other: %s <%s>", firstName+" "+lastName, email)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	if len(toDelete) == 0 {
		log.Println("No mock users to delete.")
		return
	}
	if *dryRun {
		log.Printf("Dry run: would delete %d user(s).", len(toDelete))
		cleanupOrphans(ctx, pool, true)
		return
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback(ctx)

	if err := detachUsers(ctx, tx, toDelete); err != nil {
		log.Fatal(err)
	}
	tag, err := tx.Exec(ctx, `DELETE FROM users WHERE id = ANY($1::uuid[])`, toDelete)
	if err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		log.Fatal(err)
	}
	log.Printf("Deleted %d mock user(s).", tag.RowsAffected())

	cleanupOrphans(ctx, pool, false)
}

func detachUsers(ctx context.Context, tx pgx.Tx, userIDs []string) error {
	stmts := []string{
		`UPDATE clubs SET created_by = NULL WHERE created_by = ANY($1::uuid[])`,
		`UPDATE commissions SET created_by = NULL WHERE created_by = ANY($1::uuid[])`,
		`UPDATE chat_groups SET created_by = NULL WHERE created_by = ANY($1::uuid[])`,
		`UPDATE club_member_role_assignments SET assigned_by = NULL WHERE assigned_by = ANY($1::uuid[])`,
		`UPDATE club_diary_entries SET created_by = NULL WHERE created_by = ANY($1::uuid[])`,
		`UPDATE club_mandates SET created_by = NULL WHERE created_by = ANY($1::uuid[])`,
		`UPDATE access_requests SET reviewed_by = NULL WHERE reviewed_by = ANY($1::uuid[])`,
		`UPDATE club_registration_requests SET reviewed_by = NULL WHERE reviewed_by = ANY($1::uuid[])`,
		`DELETE FROM email_invites WHERE invited_by = ANY($1::uuid[])`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(ctx, stmt, userIDs); err != nil {
			return err
		}
	}
	return nil
}

func tombolaEmails(ctx context.Context, pool *pgxpool.Pool) []string {
	sourceURL := strings.TrimSpace(os.Getenv("TOMBOLA_DATABASE_URL"))
	if sourceURL == "" {
		return nil
	}
	source, err := pgxpool.New(ctx, sourceURL)
	if err != nil {
		log.Printf("skip tombola email list: %v", err)
		return nil
	}
	defer source.Close()

	rows, err := source.Query(ctx, `SELECT LOWER(TRIM(email)) FROM members WHERE email IS NOT NULL AND TRIM(email) <> ''`)
	if err != nil {
		log.Printf("skip tombola email list: %v", err)
		return nil
	}
	defer rows.Close()

	emails := make([]string, 0)
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			log.Fatal(err)
		}
		emails = append(emails, email)
	}
	return emails
}

func shouldDeleteMock(email string) bool {
	switch {
	case strings.HasSuffix(email, "@smoke.test"):
		return true
	case strings.HasSuffix(email, "@rotaract-civ.local"):
		return true
	case strings.HasSuffix(email, "@example.com"):
		return true
	case strings.HasSuffix(email, "@test.com"):
		return true
	case strings.Contains(email, "smoke.test"):
		return true
	default:
		return false
	}
}

func splitEmails(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		email := strings.ToLower(strings.TrimSpace(part))
		if email != "" && strings.Contains(email, "@") {
			out = append(out, email)
		}
	}
	return out
}

func cleanupOrphans(ctx context.Context, pool *pgxpool.Pool, dryRun bool) {
	// Remove empty smoke clubs with no members.
	rows, err := pool.Query(ctx, `
		SELECT c.id::text, c.name
		FROM clubs c
		LEFT JOIN club_memberships m ON m.club_id = c.id
		WHERE m.id IS NULL
		ORDER BY c.created_at
	`)
	if err != nil {
		log.Printf("club cleanup skipped: %v", err)
		return
	}
	defer rows.Close()

	var clubIDs []string
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Fatal(err)
		}
		log.Printf("delete empty club: %s", name)
		clubIDs = append(clubIDs, id)
	}
	if len(clubIDs) == 0 {
		return
	}
	if dryRun {
		log.Printf("Dry run: would delete %d empty club(s).", len(clubIDs))
		return
	}
	tag, err := pool.Exec(ctx, `DELETE FROM clubs WHERE id = ANY($1::uuid[])`, clubIDs)
	if err != nil {
		log.Printf("club cleanup failed: %v", err)
		return
	}
	log.Printf("Deleted %d empty club(s).", tag.RowsAffected())
}
