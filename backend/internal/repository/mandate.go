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

const mandateSelectColumns = `
	id, club_id, name, started_at, ended_at, is_current, notes, sort_order, created_by, created_at, updated_at
`

const assignmentSelectColumns = `
	id, mandate_id, role, commission_id, commission_name, user_id,
	first_name, last_name, photo_path, notes, sort_order, created_at, updated_at
`

type MandateRepository struct {
	pool *pgxpool.Pool
}

func NewMandateRepository(pool *pgxpool.Pool) *MandateRepository {
	return &MandateRepository{pool: pool}
}

func (r *MandateRepository) ListByClub(ctx context.Context, clubID uuid.UUID) ([]domain.ClubMandate, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+mandateSelectColumns+`
		FROM club_mandates
		WHERE club_id = $1
		ORDER BY sort_order ASC, started_at DESC
	`, clubID)
	if err != nil {
		return nil, fmt.Errorf("list mandates: %w", err)
	}
	defer rows.Close()

	mandates := make([]domain.ClubMandate, 0)
	for rows.Next() {
		var m domain.ClubMandate
		if err := rows.Scan(
			&m.ID, &m.ClubID, &m.Name, &m.StartedAt, &m.EndedAt, &m.IsCurrent,
			&m.Notes, &m.SortOrder, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan mandate: %w", err)
		}
		mandates = append(mandates, m)
	}
	return mandates, rows.Err()
}

func (r *MandateRepository) GetByID(ctx context.Context, clubID, mandateID uuid.UUID) (*domain.ClubMandate, error) {
	return r.scanMandate(r.pool.QueryRow(ctx, `
		SELECT `+mandateSelectColumns+`
		FROM club_mandates WHERE id = $1 AND club_id = $2
	`, mandateID, clubID))
}

func (r *MandateRepository) Create(ctx context.Context, mandate *domain.ClubMandate) error {
	if mandate.IsCurrent {
		if _, err := r.pool.Exec(ctx, `
			UPDATE club_mandates SET is_current = FALSE, updated_at = NOW() WHERE club_id = $1
		`, mandate.ClubID); err != nil {
			return fmt.Errorf("clear current mandate: %w", err)
		}
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO club_mandates (club_id, name, started_at, ended_at, is_current, notes, sort_order, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`, mandate.ClubID, mandate.Name, mandate.StartedAt, mandate.EndedAt, mandate.IsCurrent,
		mandate.Notes, mandate.SortOrder, mandate.CreatedBy,
	).Scan(&mandate.ID, &mandate.CreatedAt, &mandate.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create mandate: %w", err)
	}
	return nil
}

func (r *MandateRepository) Update(ctx context.Context, mandate *domain.ClubMandate) (*domain.ClubMandate, error) {
	if mandate.IsCurrent {
		if _, err := r.pool.Exec(ctx, `
			UPDATE club_mandates SET is_current = FALSE, updated_at = NOW()
			WHERE club_id = $1 AND id <> $2
		`, mandate.ClubID, mandate.ID); err != nil {
			return nil, fmt.Errorf("clear current mandate: %w", err)
		}
	}
	return r.scanMandate(r.pool.QueryRow(ctx, `
		UPDATE club_mandates
		SET name = $3, started_at = $4, ended_at = $5, is_current = $6, notes = $7, sort_order = $8, updated_at = NOW()
		WHERE id = $1 AND club_id = $2
		RETURNING `+mandateSelectColumns+`
	`, mandate.ID, mandate.ClubID, mandate.Name, mandate.StartedAt, mandate.EndedAt,
		mandate.IsCurrent, mandate.Notes, mandate.SortOrder))
}

func (r *MandateRepository) Delete(ctx context.Context, clubID, mandateID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM club_mandates WHERE id = $1 AND club_id = $2`, mandateID, clubID)
	if err != nil {
		return fmt.Errorf("delete mandate: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MandateRepository) ListAssignments(ctx context.Context, mandateID uuid.UUID) ([]domain.ClubMandateAssignment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+assignmentSelectColumns+`
		FROM club_mandate_assignments
		WHERE mandate_id = $1
		ORDER BY sort_order ASC, role ASC, last_name ASC
	`, mandateID)
	if err != nil {
		return nil, fmt.Errorf("list mandate assignments: %w", err)
	}
	defer rows.Close()
	return r.scanAssignments(rows)
}

func (r *MandateRepository) GetAssignment(ctx context.Context, mandateID, assignmentID uuid.UUID) (*domain.ClubMandateAssignment, error) {
	return r.scanAssignment(r.pool.QueryRow(ctx, `
		SELECT `+assignmentSelectColumns+`
		FROM club_mandate_assignments WHERE id = $1 AND mandate_id = $2
	`, assignmentID, mandateID))
}

func (r *MandateRepository) CreateAssignment(ctx context.Context, a *domain.ClubMandateAssignment) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO club_mandate_assignments (
			mandate_id, role, commission_id, commission_name, user_id,
			first_name, last_name, photo_path, notes, sort_order
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`, a.MandateID, a.Role, a.CommissionID, a.CommissionName, a.UserID,
		a.FirstName, a.LastName, a.PhotoPath, a.Notes, a.SortOrder,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create mandate assignment: %w", err)
	}
	return nil
}

func (r *MandateRepository) UpdateAssignment(ctx context.Context, a *domain.ClubMandateAssignment) (*domain.ClubMandateAssignment, error) {
	return r.scanAssignment(r.pool.QueryRow(ctx, `
		UPDATE club_mandate_assignments
		SET role = $3, commission_id = $4, commission_name = $5, user_id = $6,
		    first_name = $7, last_name = $8, notes = $9, sort_order = $10, updated_at = NOW()
		WHERE id = $1 AND mandate_id = $2
		RETURNING `+assignmentSelectColumns+`
	`, a.ID, a.MandateID, a.Role, a.CommissionID, a.CommissionName, a.UserID,
		a.FirstName, a.LastName, a.Notes, a.SortOrder))
}

func (r *MandateRepository) UpdateAssignmentPhoto(ctx context.Context, mandateID, assignmentID uuid.UUID, path string) (*domain.ClubMandateAssignment, error) {
	var photoPath *string
	if path != "" {
		photoPath = &path
	}
	return r.scanAssignment(r.pool.QueryRow(ctx, `
		UPDATE club_mandate_assignments SET photo_path = $3, updated_at = NOW()
		WHERE id = $1 AND mandate_id = $2
		RETURNING `+assignmentSelectColumns+`
	`, assignmentID, mandateID, photoPath))
}

func (r *MandateRepository) CountAssignmentsByRole(ctx context.Context, mandateID uuid.UUID, role domain.ClubMandateRole, excludeID *uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) FROM club_mandate_assignments
		WHERE mandate_id = $1 AND role = $2
	`
	args := []any{mandateID, role}
	if excludeID != nil {
		query += ` AND id <> $3`
		args = append(args, *excludeID)
	}
	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *MandateRepository) DeleteAssignment(ctx context.Context, mandateID, assignmentID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM club_mandate_assignments WHERE id = $1 AND mandate_id = $2
	`, assignmentID, mandateID)
	if err != nil {
		return fmt.Errorf("delete mandate assignment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MandateRepository) GetCurrentByClub(ctx context.Context, clubID uuid.UUID) (*domain.ClubMandate, error) {
	return r.scanMandate(r.pool.QueryRow(ctx, `
		SELECT `+mandateSelectColumns+`
		FROM club_mandates WHERE club_id = $1 AND is_current = TRUE
		LIMIT 1
	`, clubID))
}

func (r *MandateRepository) scanMandate(row pgx.Row) (*domain.ClubMandate, error) {
	var m domain.ClubMandate
	err := row.Scan(
		&m.ID, &m.ClubID, &m.Name, &m.StartedAt, &m.EndedAt, &m.IsCurrent,
		&m.Notes, &m.SortOrder, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan mandate: %w", err)
	}
	return &m, nil
}

func (r *MandateRepository) scanAssignment(row pgx.Row) (*domain.ClubMandateAssignment, error) {
	var a domain.ClubMandateAssignment
	err := row.Scan(
		&a.ID, &a.MandateID, &a.Role, &a.CommissionID, &a.CommissionName, &a.UserID,
		&a.FirstName, &a.LastName, &a.PhotoPath, &a.Notes, &a.SortOrder, &a.CreatedAt, &a.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan mandate assignment: %w", err)
	}
	return &a, nil
}

func (r *MandateRepository) scanAssignments(rows pgx.Rows) ([]domain.ClubMandateAssignment, error) {
	list := make([]domain.ClubMandateAssignment, 0)
	for rows.Next() {
		var a domain.ClubMandateAssignment
		if err := rows.Scan(
			&a.ID, &a.MandateID, &a.Role, &a.CommissionID, &a.CommissionName, &a.UserID,
			&a.FirstName, &a.LastName, &a.PhotoPath, &a.Notes, &a.SortOrder, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan mandate assignment: %w", err)
		}
		list = append(list, a)
	}
	return list, rows.Err()
}
