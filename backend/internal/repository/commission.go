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

type CommissionRepository struct {
	pool *pgxpool.Pool
}

func NewCommissionRepository(pool *pgxpool.Pool) *CommissionRepository {
	return &CommissionRepository{pool: pool}
}

func (r *CommissionRepository) Create(ctx context.Context, commission *domain.Commission) error {
	query := `
		INSERT INTO commissions (club_id, code, name, description, is_system, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query,
		commission.ClubID, commission.Code, commission.Name, commission.Description,
		commission.IsSystem, commission.CreatedBy,
	).Scan(&commission.ID, &commission.CreatedAt)
	if err != nil {
		return fmt.Errorf("create commission: %w", err)
	}
	return nil
}

func (r *CommissionRepository) ListByClub(ctx context.Context, clubID uuid.UUID) ([]domain.Commission, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, club_id, code, name, description, is_system, created_by, created_at
		FROM commissions WHERE club_id = $1 ORDER BY is_system DESC, created_at ASC
	`, clubID)
	if err != nil {
		return nil, fmt.Errorf("list commissions: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Commission, 0)
	for rows.Next() {
		var item domain.Commission
		if err := rows.Scan(
			&item.ID, &item.ClubID, &item.Code, &item.Name, &item.Description,
			&item.IsSystem, &item.CreatedBy, &item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan commission: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *CommissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Commission, error) {
	var item domain.Commission
	err := r.pool.QueryRow(ctx, `
		SELECT id, club_id, code, name, description, is_system, created_by, created_at
		FROM commissions WHERE id = $1
	`, id).Scan(
		&item.ID, &item.ClubID, &item.Code, &item.Name, &item.Description,
		&item.IsSystem, &item.CreatedBy, &item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get commission: %w", err)
	}
	return &item, nil
}

func (r *CommissionRepository) Update(ctx context.Context, commission *domain.Commission) (*domain.Commission, error) {
	err := r.pool.QueryRow(ctx, `
		UPDATE commissions SET name = $2, description = $3
		WHERE id = $1
		RETURNING id, club_id, code, name, description, is_system, created_by, created_at
	`, commission.ID, commission.Name, commission.Description).Scan(
		&commission.ID, &commission.ClubID, &commission.Code, &commission.Name, &commission.Description,
		&commission.IsSystem, &commission.CreatedBy, &commission.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update commission: %w", err)
	}
	return commission, nil
}

func (r *CommissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM commissions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete commission: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CommissionRepository) RemoveMember(ctx context.Context, commissionID, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM commission_memberships WHERE commission_id = $1 AND user_id = $2
	`, commissionID, userID)
	if err != nil {
		return fmt.Errorf("remove commission member: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CommissionRepository) AddMember(ctx context.Context, membership *domain.CommissionMembership) error {
	query := `
		INSERT INTO commission_memberships (commission_id, user_id, member_role)
		VALUES ($1, $2, $3)
		RETURNING id, joined_at
	`
	err := r.pool.QueryRow(ctx, query, membership.CommissionID, membership.UserID, membership.MemberRole).
		Scan(&membership.ID, &membership.JoinedAt)
	if err != nil {
		return fmt.Errorf("add commission member: %w", err)
	}
	return nil
}

func (r *CommissionRepository) ListMembers(ctx context.Context, commissionID uuid.UUID) ([]domain.CommissionMembership, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT cm.id, cm.commission_id, cm.user_id, cm.member_role, cm.joined_at,
		       `+userSelectColumnsAliased+`
		FROM commission_memberships cm
		JOIN users u ON u.id = cm.user_id
		WHERE cm.commission_id = $1
		ORDER BY cm.joined_at ASC
	`, commissionID)
	if err != nil {
		return nil, fmt.Errorf("list commission members: %w", err)
	}
	defer rows.Close()

	members := make([]domain.CommissionMembership, 0)
	for rows.Next() {
		var membership domain.CommissionMembership
		var user domain.User
		if err := rows.Scan(
			&membership.ID, &membership.CommissionID, &membership.UserID, &membership.MemberRole, &membership.JoinedAt,
			&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Phone,
			&user.BirthDate, &user.Profession, &user.MemberSince, &user.AvatarPath,
			&user.IsAdmin, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan commission member: %w", err)
		}
		membership.User = &user
		members = append(members, membership)
	}
	return members, rows.Err()
}

func (r *CommissionRepository) IsPresident(ctx context.Context, commissionID, userID uuid.UUID) (bool, error) {
	var role domain.CommissionMemberRole
	err := r.pool.QueryRow(ctx, `
		SELECT member_role FROM commission_memberships
		WHERE commission_id = $1 AND user_id = $2
	`, commissionID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get commission role: %w", err)
	}
	return role == domain.CommissionMemberRolePresident, nil
}

func (r *CommissionRepository) CountPresidents(ctx context.Context, commissionID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM commission_memberships
		WHERE commission_id = $1 AND member_role = 'president'
	`, commissionID).Scan(&count)
	return count, err
}

func (r *CommissionRepository) ListByUserInClub(ctx context.Context, clubID, userID uuid.UUID) ([]domain.UserCommissionAccess, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.name, cm.member_role
		FROM commission_memberships cm
		JOIN commissions c ON c.id = cm.commission_id
		WHERE c.club_id = $1 AND cm.user_id = $2
		ORDER BY c.name ASC
	`, clubID, userID)
	if err != nil {
		return nil, fmt.Errorf("list user commissions: %w", err)
	}
	defer rows.Close()

	items := make([]domain.UserCommissionAccess, 0)
	for rows.Next() {
		var item domain.UserCommissionAccess
		if err := rows.Scan(&item.CommissionID, &item.CommissionName, &item.MemberRole); err != nil {
			return nil, fmt.Errorf("scan user commission: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *CommissionRepository) CountSecretaries(ctx context.Context, commissionID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM commission_memberships
		WHERE commission_id = $1 AND member_role = 'secretary'
	`, commissionID).Scan(&count)
	return count, err
}
