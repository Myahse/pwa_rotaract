package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotaract-civ/backend/internal/domain"
)

const diarySelectColumns = `
	id, club_id, parent_id, entry_type, user_id,
	first_name, last_name, organization, photo_path,
	started_at, ended_at, notes, sort_order,
	created_by, created_at, updated_at
`

type DiaryRepository struct {
	pool *pgxpool.Pool
}

func NewDiaryRepository(pool *pgxpool.Pool) *DiaryRepository {
	return &DiaryRepository{pool: pool}
}

func (r *DiaryRepository) ListByClub(ctx context.Context, clubID uuid.UUID) ([]domain.ClubDiaryEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+diarySelectColumns+`
		FROM club_diary_entries
		WHERE club_id = $1
		ORDER BY sort_order ASC, started_at ASC NULLS LAST, created_at ASC
	`, clubID)
	if err != nil {
		return nil, fmt.Errorf("list diary entries: %w", err)
	}
	defer rows.Close()
	return r.scanEntries(rows)
}

func (r *DiaryRepository) GetByID(ctx context.Context, clubID, entryID uuid.UUID) (*domain.ClubDiaryEntry, error) {
	query := `
		SELECT ` + diarySelectColumns + `
		FROM club_diary_entries
		WHERE id = $1 AND club_id = $2
	`
	return r.scanEntry(r.pool.QueryRow(ctx, query, entryID, clubID))
}

func (r *DiaryRepository) Create(ctx context.Context, entry *domain.ClubDiaryEntry) error {
	query := `
		INSERT INTO club_diary_entries (
			club_id, parent_id, entry_type, user_id,
			first_name, last_name, organization, photo_path,
			started_at, ended_at, notes, sort_order, created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		entry.ClubID, entry.ParentID, entry.EntryType, entry.UserID,
		entry.FirstName, entry.LastName, entry.Organization, entry.PhotoPath,
		entry.StartedAt, entry.EndedAt, entry.Notes, entry.SortOrder, entry.CreatedBy,
	).Scan(&entry.ID, &entry.CreatedAt, &entry.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create diary entry: %w", err)
	}
	return nil
}

func (r *DiaryRepository) Update(ctx context.Context, entry *domain.ClubDiaryEntry) (*domain.ClubDiaryEntry, error) {
	err := r.pool.QueryRow(ctx, `
		UPDATE club_diary_entries
		SET parent_id = $3, entry_type = $4, user_id = $5,
		    first_name = $6, last_name = $7, organization = $8,
		    started_at = $9, ended_at = $10, notes = $11, sort_order = $12,
		    updated_at = NOW()
		WHERE id = $1 AND club_id = $2
		RETURNING `+diarySelectColumns+`
	`, entry.ID, entry.ClubID, entry.ParentID, entry.EntryType, entry.UserID,
		entry.FirstName, entry.LastName, entry.Organization,
		entry.StartedAt, entry.EndedAt, entry.Notes, entry.SortOrder,
	).Scan(
		&entry.ID, &entry.ClubID, &entry.ParentID, &entry.EntryType, &entry.UserID,
		&entry.FirstName, &entry.LastName, &entry.Organization, &entry.PhotoPath,
		&entry.StartedAt, &entry.EndedAt, &entry.Notes, &entry.SortOrder,
		&entry.CreatedBy, &entry.CreatedAt, &entry.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update diary entry: %w", err)
	}
	return entry, nil
}

func (r *DiaryRepository) UpdatePhotoPath(ctx context.Context, clubID, entryID uuid.UUID, path string) (*domain.ClubDiaryEntry, error) {
	var photoPath *string
	if path != "" {
		photoPath = &path
	}
	return r.scanEntry(r.pool.QueryRow(ctx, `
		UPDATE club_diary_entries SET photo_path = $3, updated_at = NOW()
		WHERE id = $1 AND club_id = $2
		RETURNING `+diarySelectColumns+`
	`, entryID, clubID, photoPath))
}

func (r *DiaryRepository) Delete(ctx context.Context, clubID, entryID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM club_diary_entries WHERE id = $1 AND club_id = $2
	`, entryID, clubID)
	if err != nil {
		return fmt.Errorf("delete diary entry: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *DiaryRepository) scanEntry(row pgx.Row) (*domain.ClubDiaryEntry, error) {
	var entry domain.ClubDiaryEntry
	err := row.Scan(
		&entry.ID, &entry.ClubID, &entry.ParentID, &entry.EntryType, &entry.UserID,
		&entry.FirstName, &entry.LastName, &entry.Organization, &entry.PhotoPath,
		&entry.StartedAt, &entry.EndedAt, &entry.Notes, &entry.SortOrder,
		&entry.CreatedBy, &entry.CreatedAt, &entry.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan diary entry: %w", err)
	}
	return &entry, nil
}

func (r *DiaryRepository) scanEntries(rows pgx.Rows) ([]domain.ClubDiaryEntry, error) {
	entries := make([]domain.ClubDiaryEntry, 0)
	for rows.Next() {
		var entry domain.ClubDiaryEntry
		if err := rows.Scan(
			&entry.ID, &entry.ClubID, &entry.ParentID, &entry.EntryType, &entry.UserID,
			&entry.FirstName, &entry.LastName, &entry.Organization, &entry.PhotoPath,
			&entry.StartedAt, &entry.EndedAt, &entry.Notes, &entry.SortOrder,
			&entry.CreatedBy, &entry.CreatedAt, &entry.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan diary entry: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
